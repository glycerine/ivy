export async function runCheck(app) {
  const mode = app.getMode();
  app.controls.showLoading(`Running ${mode} check...`);
  app.controls.setStatus('Recompiling editor content...');
  try {
    const editorContent = app.cmEditor ? app.cmEditor.getValue() : app._persistedFileContent;
    if (editorContent) {
      await app.api.reloadContent(editorContent, app._persistedFileName || 'model.ivy');
    }
    app.controls.setStatus(`Running ${mode} check...`);
    const result = await app.api.runCheck(mode);

    const argData = await app.api.getARG();
    if (argData && argData.elements) {
      if (app.acceptArgSnapshot) app.acceptArgSnapshot(app.activeSheetId || 'sheet-1', argData);
      app.argGraph.update(argData.elements, argData.positions);
    }

    const conceptData = await app.api.getConceptGraph();
    if (conceptData && conceptData.elements) {
      app._lastConceptData = conceptData;
      if (app.acceptConceptSnapshot) app.acceptConceptSnapshot(app.activeSheetId || 'sheet-1', conceptData);
      app.conceptGraph.update(conceptData.elements, conceptData.positions);
    }

    if (result && result.used_relations) {
      await app._autoCheckUsedRelations(result.used_relations);
    }
    app.showCheckResult(result);
  } catch (err) {
    app.controls.setStatus(`Check failed: ${err.message}`, 'error');
    console.error('Check error:', err);
  } finally {
    app.controls.hideLoading();
  }
}

export async function autoCheckUsedRelations(app, relationNames, doc = globalThis.document) {
  if (!relationNames || relationNames.length === 0) return;
  const usedSet = {};
  for (const relationName of relationNames) {
    usedSet[relationName] = true;
  }
  const tbody = doc.getElementById('state-checkbox-body');
  if (!tbody) return;
  const rows = tbody.querySelectorAll('tr');
  for (const row of rows) {
    const nameCell = row.querySelector('.name-col a');
    if (!nameCell) continue;
    const name = nameCell.textContent.trim();
    const baseName = name.split('(')[0];
    if (usedSet[name] || usedSet[baseName]) {
      const inputs = row.querySelectorAll('input[type="checkbox"]');
      if (inputs.length > 0 && !inputs[0].checked) {
        inputs[0].checked = true;
        await app.onEdgeToggle(name, 'all_to_all', true);
      }
    }
  }
}

export function showCheckResult(app, result) {
  if (!result) return;
  const verdict = result.result || result.status;
  const z3note = result.z3_contacted ? ' [Z3: yes]' : ' [Z3: no]';
  const mode = result.mode ? ` (${result.mode})` : '';

  if (verdict === 'pass') {
    app.controls.setStatus(`Check PASSED${mode}${z3note}`, 'success');
    app.controls.showInfo('Verification Result', `PASSED${z3note}: ${result.message || 'All properties hold.'}`);
  } else if (verdict === 'fail') {
    let failDetails = result.message || 'Counterexample found.';
    if (result.failed_conjecture) {
      failDetails += `\n\n${result.failed_conjecture}`;
    }
    if (result.counterexample_details) {
      failDetails += `\n\n${result.counterexample_details}`;
    } else if (result.counterexample_trace) {
      failDetails += `\n\nCounterexample trace:\n${result.counterexample_trace}`;
    }
    app.controls.setStatus(`Check FAILED${mode}${z3note} - counterexample found`, 'error');
    app.controls.showInfo('Verification Result', `FAILED${z3note}: ${failDetails}`);
    app.addCheckResultViewActions(result);
    if (result.arg) {
      if (app.acceptArgSnapshot) app.acceptArgSnapshot(app.activeSheetId || 'sheet-1', result.arg);
      app.argGraph.update(result.arg.elements, result.arg.positions);
    }
  } else if (verdict === 'error') {
    app.controls.setStatus(`Check ERROR${mode}${z3note}`, 'error');
    app.controls.showInfo('Verification Error', result.message || 'Unknown error');
  } else {
    app.controls.setStatus(`Check result: ${verdict}${z3note}`);
    app.controls.showInfo('Verification Result', result.message || JSON.stringify(result));
  }
}

export function addCheckResultViewActions(app, result, {
  doc = globalThis.document,
} = {}) {
  if (!result || !result.trace_arg) return;
  const openTrace = () => {
    app.openARGSheet('Error trace', result.trace_arg, result.trace_sheet_id);
  };

  if (doc.querySelector('[data-check-view-trace]')) return;
  const info = doc.getElementById('info-content');
  if (!info) return;
  const button = doc.createElement('button');
  button.type = 'button';
  button.className = 'btn small';
  button.setAttribute('data-check-view-trace', 'true');
  button.textContent = 'View error trace';
  button.addEventListener('click', openTrace);
  info.appendChild(doc.createElement('br'));
  info.appendChild(button);
}

export async function checkInduction(app) {
  app.controls.setStatus('Checking induction...');
  app.controls.showLoading('Checking inductiveness...');
  try {
    const result = await app.api.runCheck('induction');
    if (result.result === 'fail' && result.failed_conjecture) {
      app.showTextDialog(
        'ivyweb',
        result.message || 'The following conjecture is not relatively inductive:',
        result.failed_conjecture,
      );
      app.controls.setStatus('Induction check: not inductive');
    } else if (result.result === 'pass') {
      app.showTextDialog(
        'ivyweb',
        'Inductive invariant found:',
        result.message.replace('Inductive invariant found:\n', ''),
      );
      app.controls.setStatus('Induction check: PASSED', 'success');
    } else {
      app.controls.setStatus(`Induction check: ${result.message || result.result}`);
    }
  } catch (err) {
    app.controls.setStatus(`Induction check failed: ${err.message}`, 'error');
  } finally {
    app.controls.hideLoading();
  }
}

export async function boundedCheck(app) {
  try {
    const bound = await app.integerDialog('Bounded check', 'Enter bound:', app.currentBound, {
      min: 1,
      okLabel: 'OK',
    });
    if (bound === null) {
      app.controls.setStatus('Bounded check cancelled');
      return;
    }
    app.currentBound = bound;
    app.controls.setStatus('Running bounded check...');
    const result = await app.api.runCheck('bounded', { bound });
    app.controls.setStatus(`Bounded check: ${result.result || 'done'}`, 'success');
  } catch (err) {
    app.controls.setStatus(`Bounded check failed: ${err.message}`, 'error');
  }
}

export async function weakenInvariant(app) {
  try {
    app.controls.setStatus('Choosing conjectures...');
    const choicesResult = await app.api.executeAction('get_conjectures', {});
    const conjectures = (choicesResult && choicesResult.conjectures) || [];
    const choices = conjectures.map((conj, index) => {
      const label = conj.label ? `[${conj.label}] ${conj.formula}` : conj.formula;
      return { label: label || String(index), value: index };
    });
    const indices = await app.listboxDialog('Weaken', 'Choose conjectures to remove:', choices, {
      multiple: true,
      okLabel: 'Weaken',
    });
    if (!indices || indices.length === 0) {
      app.controls.setStatus('Weaken cancelled');
      return undefined;
    }
    app.controls.setStatus('Weakening invariant...');
    const result = await app.api.executeAction('weaken', { indices });
    await app.refreshConceptGraph();
    app.controls.setStatus('Invariant weakened', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`Weaken failed: ${err.message}`, 'error');
    return null;
  }
}

export async function ctiConceptAction(app, actionName) {
  app.controls.setStatus('Running CTI action...');
  try {
    const result = await app.api.executeAction(actionName, { sheet_id: app.activeSheetId || 'sheet-1' });
    if (result && result.concept) {
      app.updateConceptGraph(result.concept);
    } else {
      await app.refreshConceptGraph();
    }
    app.controls.setStatus((result && result.message) || 'CTI action complete', 'success');
    return result;
  } catch (err) {
    app.controls.setStatus(`CTI action failed: ${err.message}`, 'error');
    return null;
  }
}
