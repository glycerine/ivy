import { describe, expect, it, vi } from 'vitest';
import {
  addCheckResultViewActions,
  autoCheckUsedRelations,
  boundedCheck,
  ctiBoundedCheck,
  ctiConceptAction,
  cancelActiveCheck,
  runCheck,
  showCheckResult,
  weakenInvariant,
} from './checkService.ts';
import { UIDataModel } from '../models/uiDataModel.ts';
import { createUIDataModelStore } from '../models/uiDataModelStore.ts';

describe('checkService', () => {
  it('renders numbered job control rows with timestamp, isolate, and run details', async () => {
    vi.useFakeTimers();
    try {
      document.body.innerHTML = '<div id="job-control-list"></div>';
      vi.setSystemTime(new Date(2026, 4, 18, 12, 34, 56, 789));
      const app = {
        getMode: vi.fn(() => 'induction'),
        cmEditor: { getValue: vi.fn(() => 'ivy source') },
        _persistedFileName: 'client_server_example.ivy',
        activeIsolate: 'no_isolates_found',
        activeSheetId: 'sheet-1',
        jobSubmissionMode: 'remote',
        api: {
          reloadContent: vi.fn(async () => ({ status: 'ok' })),
          runCheck: vi.fn(async () => ({
            status: 'ok',
            result: 'pass',
            mode: 'induction',
            z3_contacted: true,
            message: 'All conjectures are inductive.',
          })),
          getARG: vi.fn(async () => null),
          getConceptGraph: vi.fn(async () => null),
        },
        controls: {
          showLoading: vi.fn(),
          hideLoading: vi.fn(),
          setStatus: vi.fn(),
        },
        showCheckResult: vi.fn(),
        _autoCheckUsedRelations: vi.fn(),
      };

      await runCheck(app);

      const row = document.querySelector('.job-control-job');
      expect(row?.textContent).toContain('001 2026-05-18 12:34:56.789000000');
      expect(row?.textContent).toContain(' induction check - pass (remote) ');
      expect(row?.textContent).toContain('file=client_server_example.ivy');
      expect(row?.textContent).toContain('isolate=no_isolates_found');
      expect(row?.textContent).toContain('mode=induction');
      expect(row?.textContent).toContain('z3=yes');
      expect(row?.textContent).toContain('message="All conjectures are inductive."');
    } finally {
      vi.useRealTimers();
    }
  });

  it('keeps job numbers stable and shows newest jobs first', async () => {
    vi.useFakeTimers();
    try {
      document.body.innerHTML = '<div id="job-control-list"></div>';
      const app = {
        getMode: vi.fn(() => 'induction'),
        cmEditor: { getValue: vi.fn(() => 'ivy source') },
        _persistedFileName: 'model.ivy',
        activeIsolate: 'cf_live',
        activeSheetId: 'sheet-1',
        jobSubmissionMode: 'browser',
        api: {
          reloadContent: vi.fn(async () => ({ status: 'ok' })),
          runCheck: vi.fn(async () => ({
            status: 'ok',
            result: 'fail',
            mode: 'induction',
            z3_contacted: false,
            message: 'Counterexample found',
          })),
          getARG: vi.fn(async () => null),
          getConceptGraph: vi.fn(async () => null),
        },
        controls: {
          showLoading: vi.fn(),
          hideLoading: vi.fn(),
          setStatus: vi.fn(),
        },
        showCheckResult: vi.fn(),
        _autoCheckUsedRelations: vi.fn(),
      };

      vi.setSystemTime(new Date(2026, 4, 18, 1, 2, 3, 4));
      await runCheck(app);
      vi.setSystemTime(new Date(2026, 4, 18, 1, 2, 4, 5));
      await runCheck(app);

      const rows = Array.from(document.querySelectorAll('.job-control-job')).map((row) => row.textContent || '');
      expect(rows).toHaveLength(2);
      expect(rows[0]).toMatch(/^002 /);
      expect(rows[1]).toMatch(/^001 /);
      expect(app._jobControlJobs.map((job) => job.number)).toEqual([1, 2]);
    } finally {
      vi.useRealTimers();
    }
  });

  it('passes an abort signal to hosted checks and reports cancellation', async () => {
    const app = {
      getMode: vi.fn(() => 'induction'),
      cmEditor: { getValue: vi.fn(() => 'ivy source') },
      _persistedFileName: 'model.ivy',
      activeIsolate: 'rnf2',
      api: {
        reloadContent: vi.fn(async () => ({ status: 'ok' })),
        runCheck: vi.fn((_mode, _options, requestOptions) => new Promise((_resolve, reject) => {
          if (requestOptions.signal.aborted) {
            reject(Object.assign(new Error('aborted'), { name: 'AbortError' }));
            return;
          }
          requestOptions.signal.addEventListener('abort', () => {
            reject(Object.assign(new Error('aborted'), { name: 'AbortError' }));
          });
        })),
      },
      controls: {
        showLoading: vi.fn(),
        hideLoading: vi.fn(),
        setStatus: vi.fn(),
      },
    };

    const running = runCheck(app);
    expect(app.api.reloadContent).toHaveBeenCalledWith('ivy source', 'model.ivy', {
      isolate: 'rnf2',
    }, expect.objectContaining({ signal: expect.any(AbortSignal) }));

    cancelActiveCheck(app);
    await running;

    expect(app.api.runCheck).toHaveBeenCalledWith('induction', {}, expect.objectContaining({
      signal: expect.any(AbortSignal),
    }));
    expect(app.controls.setStatus).toHaveBeenCalledWith('Cancelling induction check...', 'warning');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('induction check cancelled', 'warning');
    expect(app.controls.hideLoading).toHaveBeenCalled();
    expect(app._activeCheck).toBeNull();
  });

  it('passes the relations-to-minimize field to induction checks', async () => {
    document.body.innerHTML = '<input id="cti-relations-to-minimize" value="p q">';
    const app = {
      getMode: vi.fn(() => 'induction'),
      _persistedFileContent: '',
      _persistedFileName: 'model.ivy',
      activeIsolate: '',
      api: {
        runCheck: vi.fn(async () => ({ result: 'pass', mode: 'induction' })),
        getARG: vi.fn(async () => null),
        getConceptGraph: vi.fn(async () => null),
      },
      controls: {
        showLoading: vi.fn(),
        hideLoading: vi.fn(),
        setStatus: vi.fn(),
      },
      showCheckResult: vi.fn(),
      _autoCheckUsedRelations: vi.fn(),
    };

    await runCheck(app);

    expect(app.api.runCheck).toHaveBeenCalledWith('induction', {
      relations_to_minimize: 'p q',
    }, expect.objectContaining({ signal: expect.any(AbortSignal) }));
  });

  it('adds CTI details and trace actions for failed checks', () => {
    const app = {
      addCheckResultViewActions: vi.fn(),
      activeSheetId: 'sheet-1',
      uiDataStore: { applyArgSnapshot: vi.fn() },
      controls: {
        setStatus: vi.fn(),
        showInfo: vi.fn(),
      },
    };

    showCheckResult(app, {
      result: 'fail',
      mode: 'induction',
      z3_contacted: true,
      message: 'Counterexample found.',
      failed_conjecture: 'bad conjecture',
      counterexample_details: 'frame details',
      arg: { elements: [1], positions: {} },
    });

    expect(app.controls.setStatus).toHaveBeenCalledWith(
      'Check FAILED (induction) [Z3: yes] - counterexample found',
      'error',
    );
    expect(app.controls.showInfo.mock.calls[0][1]).toContain('frame details');
    expect(app.addCheckResultViewActions).toHaveBeenCalled();
    expect(app.uiDataStore.applyArgSnapshot).toHaveBeenCalledWith('sheet-1', {
      elements: [1],
      positions: {},
    });
  });

  it('adds a direct DOM view-trace action', () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const app = {
      openARGSheet: vi.fn(),
    };
    const result = {
      trace_arg: { elements: [] },
      trace_sheet_id: 'trace-1',
    };

    addCheckResultViewActions(app, result, { doc: document });
    document.querySelector('[data-check-view-trace]').click();

    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', result.trace_arg, 'trace-1', {
      reachabilityOnly: true,
      visualOnly: false,
    });
  });

  it('reuses an open trace sheet instead of recreating it when clicked again', () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const openSheetIds = new Set();
    const app = {
      sheetExists: vi.fn((id) => openSheetIds.has(id)),
      switchSheet: vi.fn(),
      applyArgSnapshot: vi.fn(),
      setSheetTabBaseLabel: vi.fn(),
      setUIMode: vi.fn(),
      openARGSheet: vi.fn((_label, _arg, id) => openSheetIds.add(id)),
    };
    const result = {
      trace_arg: { elements: [] },
      trace_sheet_id: 'sheet-2',
    };

    addCheckResultViewActions(app, result, { doc: document });
    const button = document.querySelector('[data-check-view-trace]');

    // First click opens the trace sheet.
    button.click();
    expect(app.openARGSheet).toHaveBeenCalledTimes(1);
    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', result.trace_arg, 'sheet-2', {
      reachabilityOnly: true,
      visualOnly: false,
    });

    // Second click must not throw "duplicate sheet id"; it reuses the open sheet.
    expect(() => button.click()).not.toThrow();
    expect(app.openARGSheet).toHaveBeenCalledTimes(1);
    expect(app.applyArgSnapshot).toHaveBeenCalledWith('sheet-2', result.trace_arg);
    expect(app.switchSheet).toHaveBeenCalledWith('sheet-2');
    expect(app.setSheetTabBaseLabel).toHaveBeenCalledWith('sheet-2', 'Error trace');
  });

  it('uses frontend-local ids for check traces without backend sheet ids', () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const app = {
      nextLocalSheetId: vi.fn(() => 'trace-7'),
      openARGSheet: vi.fn(),
    };
    const result = {
      trace_arg: { elements: [] },
    };

    addCheckResultViewActions(app, result, { doc: document });
    document.querySelector('[data-check-view-trace]').click();

    expect(app.nextLocalSheetId).toHaveBeenCalledWith('trace');
    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', result.trace_arg, 'trace-7', {
      reachabilityOnly: true,
      visualOnly: true,
    });
  });

  it('prompts for a bounded-check bound and sends it to the backend', async () => {
    const app = {
      currentBound: 3,
      integerDialog: vi.fn(async () => 7),
      api: {
        runCheck: vi.fn(async () => ({ result: 'pass' })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await boundedCheck(app);

    expect(app.integerDialog).toHaveBeenCalledWith('Bounded check', 'Enter bound:', 3, {
      min: 1,
      okLabel: 'OK',
    });
    expect(app.api.runCheck).toHaveBeenCalledWith('bounded', { bound: 7 });
    expect(app.currentBound).toBe(7);
  });

  it('runs CTI concept bounded check against the active sheet', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      currentBound: -1,
      integerDialog: vi.fn(async () => 0),
      uiDataStore: {
        applyConceptSnapshot: vi.fn(),
      },
      textDialog: vi.fn(async () => 'not p(X)'),
      api: {
        executeAction: vi.fn(async () => ({
          result: 'pass',
          message: 'BMC with bound 0 did not find a counter-example to:\nnot p(X)',
          concept: { elements: [] },
        })),
      },
      controls: {
        setStatus: vi.fn(),
        showInfo: vi.fn(),
      },
    };

    await ctiBoundedCheck(app);

    expect(app.integerDialog).toHaveBeenCalledWith('Bounded check', 'Number of steps to check:', 10, {
      min: 0,
      okLabel: 'OK',
    });
    expect(app.api.executeAction).toHaveBeenCalledWith('cti_bounded_check', {
      sheet_id: 'sheet-2',
      bound: 0,
    });
    expect(app.currentBound).toBe(0);
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-2', { elements: [] });
    expect(app.controls.setStatus).toHaveBeenLastCalledWith(
      'BMC with bound 0 did not find a counter-example to:\nnot p(X)',
      'success',
    );
    expect(app.textDialog).toHaveBeenCalledWith(
      'ivyweb',
      'BMC with bound 0 did not find a counter-example to:',
      'not p(X)',
      { okLabel: 'OK', cancel: false },
    );
  });

  it('opens CTI BMC counterexample traces through the Python-style View button', async () => {
    const traceArg = { elements: [{ group: 'nodes', data: { id: 'state_0' } }] };
    const app = {
      activeSheetId: 'sheet-2',
      currentBound: 10,
      integerDialog: vi.fn(async () => 0),
      textDialog: vi.fn(async () => '(~true)'),
      setUIMode: vi.fn(),
      openARGSheet: vi.fn(),
      api: {
        executeAction: vi.fn(async () => ({
          found: true,
          result: 'fail',
          message: 'BMC with bound 0 found a counter-example to:\n(~true)',
          conjecture: '(~true)',
          trace_arg: traceArg,
          trace_sheet_id: 'sheet-3',
          trace_label: 'Sheet 3',
        })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await ctiBoundedCheck(app);

    expect(app.textDialog).toHaveBeenCalledWith(
      'ivyweb',
      'BMC with bound 0 found a counter-example to:',
      '(~true)',
      { okLabel: 'View', cancel: true, primaryFirst: true },
    );
    expect(app.setUIMode).toHaveBeenCalledWith('reachability');
    expect(app.openARGSheet).toHaveBeenCalledWith('Sheet 3', traceArg, 'sheet-3', {
      reachabilityOnly: true,
    });
  });

  it('does not open CTI BMC counterexample traces when the dialog is cancelled', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      currentBound: 10,
      integerDialog: vi.fn(async () => 0),
      textDialog: vi.fn(async () => null),
      setUIMode: vi.fn(),
      openARGSheet: vi.fn(),
      api: {
        executeAction: vi.fn(async () => ({
          found: true,
          result: 'fail',
          message: 'BMC with bound 0 found a counter-example to:\n(~true)',
          trace_arg: { elements: [] },
          trace_sheet_id: 'sheet-3',
        })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await ctiBoundedCheck(app);

    expect(app.textDialog).toHaveBeenCalledWith(
      'ivyweb',
      'BMC with bound 0 found a counter-example to:',
      '(~true)',
      { okLabel: 'View', cancel: true, primaryFirst: true },
    );
    expect(app.setUIMode).not.toHaveBeenCalled();
    expect(app.openARGSheet).not.toHaveBeenCalled();
  });

  it('prompts for conjectures before weakening', async () => {
    const app = {
      refreshConceptGraph: vi.fn(),
      listboxDialog: vi.fn(async () => [0]),
      api: {
        executeAction: vi.fn(async (action) => {
          if (action === 'get_conjectures') {
            return {
              conjectures: [
                { label: 'c0', formula: 'p(X)' },
                { label: 'c1', formula: 'q(X)' },
              ],
            };
          }
          return { removed: ['p(X)'], removed_count: 1 };
        }),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await weakenInvariant(app);

    expect(app.api.executeAction).toHaveBeenNthCalledWith(1, 'get_conjectures', {});
    expect(app.listboxDialog).toHaveBeenCalledWith(
      'Weaken',
      'Choose conjectures to remove:',
      [
        { label: '[c0] p(X)', value: 0 },
        { label: '[c1] q(X)', value: 1 },
      ],
      { multiple: true, okLabel: 'Weaken' },
    );
    expect(app.api.executeAction).toHaveBeenNthCalledWith(2, 'weaken', { indices: [0] });
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Invariant weakened', 'success');
  });

  it('dispatches CTI concept graph actions with the active sheet id', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      uiDataStore: {
        applyConceptSnapshot: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      api: {
        executeAction: vi.fn(async () => ({ ok: true, message: 'strengthened', concept: { elements: [] } })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await ctiConceptAction(app, 'cti_gather');

    expect(app.api.executeAction).toHaveBeenCalledWith('cti_gather', { sheet_id: 'sheet-2' });
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-2', { elements: [] });
    expect(app.refreshConceptGraph).not.toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('strengthened', 'success');
  });

  it('cancels CTI strengthen after previewing the exact conjecture', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      textDialog: vi.fn(async () => null),
      uiDataStore: {
        applyConceptSnapshot: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      api: {
        executeAction: vi.fn(async (action) => {
          if (action === 'cti_strengthen_preview') {
            return { conjecture: 'forall X. p(X)' };
          }
          return { ok: true, message: 'strengthened', concept: { elements: [] } };
        }),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    const result = await ctiConceptAction(app, 'cti_strengthen');

    expect(result).toBeNull();
    expect(app.api.executeAction).toHaveBeenCalledTimes(1);
    expect(app.api.executeAction).toHaveBeenCalledWith('cti_strengthen_preview', { sheet_id: 'sheet-2' });
    expect(app.textDialog).toHaveBeenCalledWith(
      'Strengthen',
      'Add this conjecture as an invariant?',
      'forall X. p(X)',
      { readOnly: true, okLabel: 'Strengthen', cancel: true },
    );
    expect(app.uiDataStore.applyConceptSnapshot).not.toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Strengthen cancelled');
  });

  it('accepts CTI strengthen confirmation and appends once', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      textDialog: vi.fn(async () => 'forall X. p(X)'),
      uiDataStore: {
        applyConceptSnapshot: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      api: {
        executeAction: vi.fn(async (action) => {
          if (action === 'cti_strengthen_preview') {
            return { conjecture: 'forall X. p(X)' };
          }
          return { ok: true, message: 'Invariant strengthened', concept: { sheet_id: 'sheet-2', elements: [] } };
        }),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    const result = await ctiConceptAction(app, 'cti_strengthen');

    expect(result && result.message).toBe('Invariant strengthened');
    expect(app.api.executeAction).toHaveBeenCalledTimes(2);
    expect(app.api.executeAction).toHaveBeenNthCalledWith(1, 'cti_strengthen_preview', { sheet_id: 'sheet-2' });
    expect(app.api.executeAction).toHaveBeenNthCalledWith(2, 'cti_strengthen', { sheet_id: 'sheet-2' });
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-2', { sheet_id: 'sheet-2', elements: [] });
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Invariant strengthened', 'success');
  });

  it('passes relations-to-minimize to CTI minimize', async () => {
    document.body.innerHTML = '<input id="cti-relations-to-minimize" value="q">';
    const app = {
      activeSheetId: 'sheet-2',
      uiDataStore: {
        applyConceptSnapshot: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      api: {
        executeAction: vi.fn(async () => ({ ok: true, message: 'minimized', concept: { elements: [] } })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await ctiConceptAction(app, 'cti_minimize');

    expect(app.api.executeAction).toHaveBeenCalledWith('cti_minimize', {
      sheet_id: 'sheet-2',
      relations_to_minimize: 'q',
    });
  });

  it('shows CTI minimize bound and core details', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      textDialog: vi.fn(async () => true),
      uiDataStore: {
        applyConceptSnapshot: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      api: {
        executeAction: vi.fn(async () => ({
          ok: true,
          message: 'Conjecture minimized using BMC bound 2; kept 1 of 2 selected facts.',
          bound: 2,
          input_facts: ['false', 'true'],
          core_facts: ['false'],
          removed_facts: ['true'],
          conjecture: '~false',
          concept: { elements: [] },
        })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await ctiConceptAction(app, 'cti_minimize');

    expect(app.textDialog).toHaveBeenCalledWith(
      'CTI minimize',
      'Conjecture minimized using BMC bound 2; kept 1 of 2 selected facts.',
      [
        'BMC bound: 2',
        '',
        'Selected facts:',
        '- false',
        '- true',
        '',
        'Core facts kept:',
        '- false',
        '',
        'Removed facts:',
        '- true',
        '',
        'Resulting conjecture:',
        '~false',
      ].join('\n'),
      { readOnly: true, okLabel: 'OK', cancel: false },
    );
  });

  it('shows CTI sufficient and relative induction result dialogs', async () => {
    const cases = [
      {
        action: 'cti_check_sufficient',
        result: 'sufficient',
        ok: true,
        message: '(1) implies (2) at the next time.',
        title: 'CTI check sufficient',
        status: 'sufficient',
      },
      {
        action: 'cti_check_sufficient',
        result: 'insufficient',
        ok: false,
        message: '(1) does not imply (2) at the next time.',
        title: 'CTI check sufficient',
        status: 'insufficient',
      },
      {
        action: 'cti_check_inductive',
        result: 'inductive',
        ok: true,
        message: '(1) is relatively inductive.',
        title: 'CTI relative induction',
        status: 'inductive',
      },
      {
        action: 'cti_check_inductive',
        result: 'non_inductive',
        ok: false,
        message: '(1) is not relatively inductive.',
        title: 'CTI relative induction',
        status: 'non_inductive',
      },
    ];

    for (const scenario of cases) {
      const app = {
        activeSheetId: 'sheet-2',
        textDialog: vi.fn(async () => true),
        uiDataStore: {
          applyConceptSnapshot: vi.fn(),
        },
        refreshConceptGraph: vi.fn(),
        api: {
          executeAction: vi.fn(async () => ({
            ok: scenario.ok,
            check: scenario.action === 'cti_check_sufficient' ? 'sufficient' : 'relative_induction',
            result: scenario.result,
            message: scenario.message,
            selected_conjecture: 'forall X. p(X)',
            target_conjecture: 'forall X. q(X)',
            concept: { elements: [] },
          })),
        },
        controls: {
          setStatus: vi.fn(),
        },
      };

      await ctiConceptAction(app, scenario.action);

      expect(app.textDialog).toHaveBeenCalledWith(
        scenario.title,
        scenario.message,
        [
          `Result: ${scenario.status}`,
          '',
          'Selected conjecture:',
          'forall X. p(X)',
          '',
          'Target conjecture:',
          'forall X. q(X)',
        ].join('\n'),
        { readOnly: true, okLabel: 'OK', cancel: false },
      );
    }
  });

  it('auto-checks used relation rows', async () => {
    const uiDataModel = new UIDataModel();
    createUIDataModelStore(uiDataModel).applyConceptSnapshot('sheet-1', {
      relations: ['link(X,Y)'],
      toggles: { edges: { 'link(X,Y)': { all_to_all: false } } },
    });
    const app = {
      uiDataModel,
      activeSheetId: 'sheet-1',
      onEdgeToggle: vi.fn(),
    };

    await autoCheckUsedRelations(app, ['link']);

    expect(app.onEdgeToggle).toHaveBeenCalledWith('link(X,Y)', 'all_to_all', true);
  });
});
