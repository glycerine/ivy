import { applyArgSnapshot, applyConceptSnapshot } from './uiDataRenderService.ts';
import { selectSheet, selectStateCheckboxRows } from '../models/uiDataSelectors.ts';
import { beginRunContext, reportRunContextError, runWithContext } from './runContextService.ts';

function activeCheckLabel(app) {
  const mode = app && typeof app.getMode === 'function' ? app.getMode() : 'verification';
  return `${mode} check`;
}

let fallbackTraceSheetCounter = 0;

function traceSheetIdForResult(app, result) {
  if (result && result.trace_sheet_id) return result.trace_sheet_id;
  if (app && typeof app.nextLocalSheetId === 'function') return app.nextLocalSheetId('trace');
  fallbackTraceSheetCounter += 1;
  return `trace-${fallbackTraceSheetCounter}`;
}

function traceSheetOptionsForResult(result) {
  return {
    reachabilityOnly: true,
    visualOnly: !(result && result.trace_sheet_id),
  };
}

function traceInfoElement(app, doc) {
  if (app && app.controls && typeof app.controls.activeInfoElement === 'function') {
    return app.controls.activeInfoElement();
  }
  return doc.getElementById('info-content');
}

function relationsToMinimizeOptions({
  doc = globalThis.document,
} = {}) {
  const input = doc && doc.getElementById('cti-relations-to-minimize') as HTMLInputElement | null;
  if (!input) return {};
  return { relations_to_minimize: input.value || '' };
}

function readSelectValue(doc, id) {
  const select = doc && doc.getElementById(id) as HTMLSelectElement | null;
  return select ? select.value : '';
}

function readInputValue(doc, id) {
  const input = doc && doc.getElementById(id) as HTMLInputElement | null;
  return input ? input.value : '';
}

function readBmcBound(doc) {
  const raw = readSelectValue(doc, 'analysis-bmc-bound');
  if (raw === '') return undefined;
  const value = Number.parseInt(raw, 10);
  return Number.isInteger(value) && value >= 0 ? value : undefined;
}

export function analysisControllerOptions({
  doc = globalThis.document,
  includeAbstractor = true,
  includeBound = true,
  includeRelations = true,
  includeTransitionLog = true,
} = {}) {
  const options: Record<string, any> = {};
  if (includeAbstractor) {
    const abstractor = readSelectValue(doc, 'analysis-abstractor-select');
    if (abstractor) options.abstractor = abstractor;
  }
  if (includeBound) {
    const bound = readBmcBound(doc);
    if (bound !== undefined) options.bound = bound;
  }
  if (includeRelations) {
    Object.assign(options, relationsToMinimizeOptions({ doc }));
  }
  if (includeTransitionLog) {
    const transitionLogFile = readInputValue(doc, 'transition-log-file').trim();
    if (transitionLogFile) options.transition_log_file = transitionLogFile;
  }
  return options;
}

export function openTraceArgFromResult(app, result, {
  label = 'Error trace',
} = {}) {
  if (!app || !result || !result.trace_arg) return false;
  if (typeof app.setUIMode === 'function') app.setUIMode('reachability');
  const sheetId = traceSheetIdForResult(app, result);
  // The "View error trace" button persists in the details pane, so it can be
  // clicked more than once for the same result. The backend hands back a stable
  // trace_sheet_id, so a second click would ask addSheet() to recreate a sheet
  // that already exists and it throws "duplicate sheet id". When the trace sheet
  // is already open, refresh it and switch to it instead (mirrors showReachableStates).
  if (sheetId && typeof app.sheetExists === 'function' && app.sheetExists(sheetId)) {
    if (typeof app.setSheetTabBaseLabel === 'function') app.setSheetTabBaseLabel(sheetId, label);
    if (typeof app.applyArgSnapshot === 'function') app.applyArgSnapshot(sheetId, result.trace_arg || {});
    if (typeof app.switchSheet === 'function') app.switchSheet(sheetId);
    return true;
  }
  app.openARGSheet(label, result.trace_arg, sheetId, traceSheetOptionsForResult(result));
  return true;
}

export function addTraceResultViewAction(app, result, {
  doc = globalThis.document,
  label = 'Error trace',
} = {}) {
  if (!result || !result.trace_arg) return false;
  if (doc.querySelector('[data-check-view-trace]')) return false;
  const info = traceInfoElement(app, doc);
  if (!info) return false;
  const button = doc.createElement('button');
  button.type = 'button';
  button.className = 'btn small';
  button.setAttribute('data-check-view-trace', 'true');
  button.textContent = 'View error trace';
  button.addEventListener('click', () => openTraceArgFromResult(app, result, { label }));
  info.appendChild(doc.createElement('br'));
  info.appendChild(button);
  return true;
}

function setCheckControlsRunning(running) {
  const doc = globalThis.document;
  if (!doc) return;
  const checkButton = doc.getElementById('btn-check') as HTMLButtonElement | null;
  const overlayCancelButton = doc.getElementById('btn-cancel-loading') as HTMLButtonElement | null;
  if (checkButton) checkButton.disabled = running;
  if (overlayCancelButton) {
    overlayCancelButton.hidden = !running;
    overlayCancelButton.disabled = !running;
  }
}

function padNumber(value, width) {
  return String(value).padStart(width, '0');
}

function formatJobTimestamp(value) {
  const date = value instanceof Date ? value : new Date(value || Date.now());
  const offsetMinutes = -date.getTimezoneOffset();
  const absOffset = Math.abs(offsetMinutes);
  const offset = offsetMinutes === 0
    ? 'Z'
    : `${offsetMinutes >= 0 ? '+' : '-'}${padNumber(Math.floor(absOffset / 60), 2)}:${padNumber(absOffset % 60, 2)}`;
  const nanos = padNumber(date.getMilliseconds() * 1000000, 9);
  return `${date.getFullYear()}-${padNumber(date.getMonth() + 1, 2)}-${padNumber(date.getDate(), 2)} ` +
    `${padNumber(date.getHours(), 2)}:${padNumber(date.getMinutes(), 2)}:${padNumber(date.getSeconds(), 2)}.${nanos}${offset}`;
}

function nextJobNumber(app) {
  app._jobControlSeq = Number.isInteger(app._jobControlSeq) ? app._jobControlSeq + 1 : 1;
  return app._jobControlSeq;
}

function jobValue(value) {
  const text = String(value == null || value === '' ? '(none)' : value);
  if (text === '(none)') return text;
  if (/^[A-Za-z0-9_.:/@+-]+$/.test(text)) return text;
  return JSON.stringify(text);
}

function jobDetails(job) {
  const parts = [];
  if (job.filename) parts.push(`file=${jobValue(job.filename)}`);
  parts.push(`isolate=${jobValue(job.isolate)}`);
  if (job.mode) parts.push(`mode=${jobValue(job.mode)}`);
  if (job.bound != null) parts.push(`bound=${jobValue(job.bound)}`);
  if (job.z3Contacted != null) parts.push(`z3=${job.z3Contacted ? 'yes' : 'no'}`);
  if (job.failedLabel) parts.push(`failed_label=${jobValue(job.failedLabel)}`);
  if (job.message) parts.push(`message=${jobValue(job.message)}`);
  return parts.join(' ');
}

function renderJobLine(job) {
  const number = padNumber(job.number || 0, 3);
  const timestamp = formatJobTimestamp(job.startedAt);
  const details = jobDetails(job);
  const base = `${number} ${timestamp} ${job.label || 'job'} - ${job.status || 'running'} (${job.backend || 'backend'})`;
  return details ? `${base} ${details}` : base;
}

function renderJobControl(app) {
  const doc = globalThis.document;
  const list = doc && doc.getElementById('job-control-list');
  if (!list) return;
  const jobs = (app && app._jobControlJobs) || [];
  list.innerHTML = '';
  if (jobs.length === 0) {
    const empty = doc.createElement('div');
    empty.className = 'job-control-empty';
    empty.textContent = 'No running jobs';
    list.appendChild(empty);
    return;
  }
  for (const job of jobs.slice().reverse()) {
    const row = doc.createElement('div');
    row.className = 'job-control-job';
    row.setAttribute('data-job-status', job.status || '');
    row.textContent = renderJobLine(job);
    list.appendChild(row);
  }
}

function upsertJob(app, job) {
  if (!app) return;
  app._jobControlJobs = app._jobControlJobs || [];
  const existing = app._jobControlJobs.find((candidate) => candidate.id === job.id);
  if (existing) Object.assign(existing, job);
  else app._jobControlJobs.push({
    number: nextJobNumber(app),
    startedAt: new Date(),
    ...job,
  });
  renderJobControl(app);
}

function makeCheckAbortController() {
  if (typeof AbortController === 'undefined') return null;
  return new AbortController();
}

function isAbortError(err) {
  return !!err && (
    err.name === 'AbortError' ||
    String(err.message || '').toLowerCase().includes('aborted') ||
    String(err.message || '').toLowerCase().includes('cancelled')
  );
}

export function cancelActiveCheck(app) {
  const active = app && app._activeCheck;
  if (!active) {
    if (app && app.controls) app.controls.setStatus('No verification check is running');
    return false;
  }
  active.cancelled = true;
  if (active.controller) active.controller.abort();
  if (app && app.controls) app.controls.setStatus(`Cancelling ${active.label || activeCheckLabel(app)}...`, 'warning');
  // Recover the UI immediately rather than waiting for the aborted request to
  // settle, so a wedged backend connection can never leave the loading overlay
  // stuck on screen.
  if (typeof active.forceRecoverUI === 'function') active.forceRecoverUI();
  if (app._activeCheck === active) app._activeCheck = null;
  return true;
}

export async function runCheck(app) {
  if (app._activeCheck) {
    app.controls.setStatus(`${app._activeCheck.label || activeCheckLabel(app)} is already running`, 'warning');
    return null;
  }
  const mode = app.getMode();
  const controller = makeCheckAbortController();
  const active = {
    controller,
    cancelled: false,
    label: `${mode} check`,
    jobId: `check-${Date.now()}`,
    forceRecoverUI: null as null | (() => void),
  };
  app._activeCheck = active;
  upsertJob(app, {
    id: active.jobId,
    label: active.label,
    status: 'running',
    backend: app.jobSubmissionMode || (app.api && app.api.kind) || 'backend',
    filename: app._persistedFileName || 'model.ivy',
    isolate: app.activeIsolate || '',
    mode,
  });
  setCheckControlsRunning(true);
  const finishRunContext = beginRunContext(app, { busyMessage: `Running ${mode} check...` });
  let runContextFinished = false;
  const finishRunContextOnce = () => {
    if (runContextFinished) return;
    runContextFinished = true;
    finishRunContext();
  };
  // Let cancelActiveCheck tear down the loading overlay and check controls
  // directly. Aborting a request does not always make the awaited promise
  // settle promptly (a wedged remote connection can leave it pending), so cancel
  // must be able to give the user back the UI without waiting for this run to
  // unwind. Sharing finishRunContextOnce keeps the run-context depth balanced.
  active.forceRecoverUI = () => {
    setCheckControlsRunning(false);
    finishRunContextOnce();
  };
  app.controls.setStatus('Recompiling editor content...');
  try {
    const requestOptions = controller ? { signal: controller.signal } : {};
    const editorContent = app.cmEditor ? app.cmEditor.getValue() : app._persistedFileContent;
    if (editorContent) {
      await app.api.reloadContent(editorContent, app._persistedFileName || 'model.ivy', {
        isolate: app.activeIsolate || '',
      }, requestOptions);
    }
    app.controls.setStatus(`Running ${mode} check...`);
    const result = await app.api.runCheck(mode, analysisControllerOptions(), requestOptions);

    if (active.cancelled || result?.result === 'cancelled') {
      app.controls.setStatus(`${mode} check cancelled`, 'warning');
      upsertJob(app, { id: active.jobId, status: 'cancelled' });
      return result;
    }

    const argData = await app.api.getARG({}, requestOptions);
    if (argData && argData.elements) {
      applyArgSnapshot(app, app.activeSheetId || 'sheet-1', argData);
    }

    const conceptData = await app.api.getConceptGraph(undefined, undefined, requestOptions);
    if (conceptData && conceptData.elements) {
      applyConceptSnapshot(app, app.activeSheetId || 'sheet-1', conceptData);
    }

    if (result && result.used_relations) {
      await app._autoCheckUsedRelations(result.used_relations);
    }
    app.showCheckResult(result);
    upsertJob(app, {
      id: active.jobId,
      status: (result && (result.result || result.status)) || 'complete',
      mode: (result && result.mode) || mode,
      z3Contacted: result && result.z3_contacted,
      failedLabel: result && result.failed_label,
      message: result && result.message,
    });
    return result;
  } catch (err) {
    if (active.cancelled || isAbortError(err)) {
      app.controls.setStatus(`${mode} check cancelled`, 'warning');
      upsertJob(app, { id: active.jobId, status: 'cancelled' });
      finishRunContextOnce();
      return null;
    }
    finishRunContextOnce();
    await reportRunContextError(app, err, { prefix: 'Check failed' });
    upsertJob(app, { id: active.jobId, status: 'error', message: err.message });
    console.error('Check error:', err);
    return null;
  } finally {
    if (app._activeCheck === active) {
      app._activeCheck = null;
    }
    setCheckControlsRunning(false);
    finishRunContextOnce();
  }
}

export async function autoCheckUsedRelations(app, relationNames) {
  if (!relationNames || relationNames.length === 0) return;
  const usedSet = {};
  for (const relationName of relationNames) {
    usedSet[relationName] = true;
  }
  const sheet = selectSheet(app && app.uiDataModel, app && app.activeSheetId);
  const rows = selectStateCheckboxRows(sheet);
  for (const row of rows) {
    const name = row.name;
    const baseName = name.split('(')[0];
    if ((usedSet[name] || usedSet[baseName]) && !(row.checked && row.checked.all_to_all)) {
      await app.onEdgeToggle(name, 'all_to_all', true);
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
      applyArgSnapshot(app, app.activeSheetId || 'sheet-1', result.arg);
    }
  } else if (verdict === 'error') {
    app.controls.setStatus(`Check ERROR${mode}${z3note}`, 'error');
    app.controls.showInfo('Verification Error', result.message || 'Unknown error');
  } else if (verdict === 'cancelled') {
    app.controls.setStatus(`Check cancelled${mode}`, 'warning');
  } else {
    app.controls.setStatus(`Check result: ${verdict}${z3note}`);
    app.controls.showInfo('Verification Result', result.message || JSON.stringify(result));
  }
}

export function addCheckResultViewActions(app, result, {
  doc = globalThis.document,
} = {}) {
  addTraceResultViewAction(app, result, { doc, label: 'Error trace' });
}

function splitResultMessage(message, result) {
  const textParts = String(message || '').split('\n');
  const dialogMessage = textParts.shift() || 'Bounded check';
  const dialogText = textParts.length > 0 ? textParts.join('\n') : (result && result.conjecture) || '';
  return { dialogMessage, dialogText };
}

function traceLabelForResult(result) {
  if (result && result.trace_label) return result.trace_label;
  const sheetId = result && result.trace_sheet_id;
  const match = /^sheet-(\d+)$/.exec(String(sheetId || ''));
  if (match) return `Sheet ${match[1]}`;
  return 'BMC counterexample';
}

export async function showCtiBoundedCheckResult(app, result, message) {
  const { dialogMessage, dialogText } = splitResultMessage(message, result);
  const foundTrace = !!(result && result.found && result.trace_arg);
  if (foundTrace && typeof app.textDialog === 'function') {
    const action = await app.textDialog('ivyweb', dialogMessage, dialogText, {
      okLabel: 'View',
      cancel: true,
      primaryFirst: true,
    });
    if (action !== null) {
      if (typeof app.setUIMode === 'function') app.setUIMode('reachability');
      app.openARGSheet(traceLabelForResult(result), result.trace_arg, result.trace_sheet_id, {
        reachabilityOnly: true,
      });
    }
    return;
  }
  if (typeof app.textDialog === 'function') {
    await app.textDialog('ivyweb', dialogMessage, dialogText, {
      okLabel: 'OK',
      cancel: false,
    });
  } else if (app.controls.showInfo) {
    app.controls.showInfo('Bounded check', message);
  }
}

function ctiResultList(title, items) {
  const values = Array.isArray(items) ? items : [];
  if (values.length === 0) return [title, '- (none)'];
  return [title, ...values.map((item) => `- ${item}`)];
}

function ctiMinimizeDetailsText(result) {
  const bound = result && result.bound !== undefined ? result.bound : 'unknown';
  const coreFacts = result && (result.core_facts || result.minimized_facts);
  const lines = [
    `BMC bound: ${bound}`,
    '',
    ...ctiResultList('Selected facts:', result && result.input_facts),
    '',
    ...ctiResultList('Core facts kept:', coreFacts),
    '',
    ...ctiResultList('Removed facts:', result && result.removed_facts),
  ];
  if (result && result.conjecture) {
    lines.push('', 'Resulting conjecture:', String(result.conjecture));
  }
  return lines.join('\n');
}

async function showCtiMinimizeDetails(app, result) {
  if (!result || result.bound === undefined || typeof app.textDialog !== 'function') return;
  await app.textDialog(
    'CTI minimize',
    result.message || 'Conjecture minimized',
    ctiMinimizeDetailsText(result),
    { readOnly: true, okLabel: 'OK', cancel: false },
  );
}

function isCtiCheckAction(actionName) {
  return actionName === 'cti_check_sufficient' || actionName === 'cti_check_inductive';
}

function ctiCheckTitle(actionName, result) {
  if (result && result.dialog_title) return String(result.dialog_title);
  if (actionName === 'cti_check_inductive') return 'CTI relative induction';
  return 'CTI check sufficient';
}

function ctiCheckDetailsText(result) {
  if (result && result.dialog_text) return String(result.dialog_text);
  const resultKind = result && result.result ? String(result.result) : (result && result.ok ? 'pass' : 'fail');
  const selected = result && result.selected_conjecture ? String(result.selected_conjecture) : '';
  const target = result && result.target_conjecture ? String(result.target_conjecture) : '';
  return [
    `Result: ${resultKind}`,
    '',
    'Selected conjecture:',
    selected,
    '',
    'Target conjecture:',
    target,
  ].join('\n');
}

async function showCtiCheckDetails(app, actionName, result) {
  if (!isCtiCheckAction(actionName) || !result || typeof app.textDialog !== 'function') return;
  await app.textDialog(
    ctiCheckTitle(actionName, result),
    result.dialog_message || result.message || 'CTI check complete',
    ctiCheckDetailsText(result),
    { readOnly: true, okLabel: 'OK', cancel: false },
  );
}

export async function checkInduction(app) {
  app.controls.setStatus('Checking induction...');
  const result = await runWithContext(app, {
    busyMessage: 'Checking inductiveness...',
    failurePrefix: 'Induction check failed',
  }, () => app.api.runCheck('induction'));
  if (!result) return null;
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
  return result;
}

export async function boundedCheck(app) {
  try {
    const options = analysisControllerOptions();
    let bound = options.bound;
    if (bound === undefined) {
      bound = await app.integerDialog('Bounded check', 'Enter bound:', app.currentBound, {
        min: 1,
        okLabel: 'OK',
      });
    }
    if (bound === null) {
      app.controls.setStatus('Bounded check cancelled');
      return;
    }
    app.currentBound = bound;
    app.controls.setStatus('Running bounded check...');
    const result = await app.api.runCheck('bounded', { ...options, bound });
    app.controls.setStatus(`Bounded check: ${result.result || 'done'}`, 'success');
  } catch (err) {
    app.controls.setStatus(`Bounded check failed: ${err.message}`, 'error');
  }
}

export async function ctiBoundedCheck(app) {
  try {
    const initialBound = Number.isInteger(app.currentBound) && app.currentBound >= 0 ? app.currentBound : 10;
    const bound = await app.integerDialog('Bounded check', 'Number of steps to check:', initialBound, {
      min: 0,
      okLabel: 'OK',
    });
    if (bound === null) {
      app.controls.setStatus('Bounded check cancelled');
      return null;
    }
    app.currentBound = bound;
    app.controls.setStatus('Running bounded check...');
    const result = await app.api.executeAction('cti_bounded_check', {
      sheet_id: app.activeSheetId || 'sheet-1',
      bound,
    });
    if (result && result.concept) {
      applyConceptSnapshot(app, result.concept.sheet_id || app.activeSheetId || 'sheet-1', result.concept);
    }
    const message = (result && result.message) || 'Bounded check complete';
    const resultKind = result && (result.result || result.status);
    const statusKind = resultKind === 'fail' ? 'error' : 'success';
    app.controls.setStatus(message, statusKind);
    await showCtiBoundedCheckResult(app, result, message);
    return result;
  } catch (err) {
    app.controls.setStatus(`Bounded check failed: ${err.message}`, 'error');
    return null;
  }
}

async function confirmCtiStrengthen(app, args) {
  const preview = await runWithContext(app, {
    busyMessage: 'Preparing strengthen confirmation...',
    failurePrefix: 'CTI strengthen failed',
  }, () => app.api.executeAction('cti_strengthen_preview', args));
  if (!preview) {
    return false;
  }
  const conjecture = String((preview && preview.conjecture) || '');
  let accepted = true;
  if (typeof app.textDialog === 'function') {
    const answer = await app.textDialog('Strengthen', 'Add this conjecture as an invariant?', conjecture, {
      readOnly: true,
      okLabel: 'Strengthen',
      cancel: true,
    });
    accepted = answer !== null;
  } else if (typeof app.okCancelDialog === 'function') {
    accepted = await app.okCancelDialog('Strengthen', `Add this conjecture as an invariant?\n\n${conjecture}`);
  }
  if (!accepted) {
    app.controls.setStatus('Strengthen cancelled');
    return false;
  }
  return true;
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
  const args = {
    sheet_id: app.activeSheetId || 'sheet-1',
    ...analysisControllerOptions(),
  };
  if (actionName === 'cti_strengthen') {
    const accepted = await confirmCtiStrengthen(app, args);
    if (!accepted) {
      return null;
    }
  }
  const result = await runWithContext(app, {
    busyMessage: 'Running CTI action...',
    failurePrefix: 'CTI action failed',
  }, () => app.api.executeAction(actionName, args));
  if (!result) {
    return null;
  }
  if (result && result.concept) {
    applyConceptSnapshot(app, result.concept.sheet_id || app.activeSheetId || 'sheet-1', result.concept);
  } else {
    await app.refreshConceptGraph();
  }
  const statusKind = isCtiCheckAction(actionName) && result && result.ok === false ? 'error' : 'success';
  app.controls.setStatus((result && result.message) || 'CTI action complete', statusKind);
  if (actionName === 'cti_minimize') {
    await showCtiMinimizeDetails(app, result);
  } else if (isCtiCheckAction(actionName)) {
    await showCtiCheckDetails(app, actionName, result);
  }
  return result;
}
