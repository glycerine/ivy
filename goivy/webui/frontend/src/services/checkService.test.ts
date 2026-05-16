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

    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', result.trace_arg, 'trace-1');
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

    await ctiConceptAction(app, 'cti_strengthen');

    expect(app.api.executeAction).toHaveBeenCalledWith('cti_strengthen', { sheet_id: 'sheet-2' });
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-2', { elements: [] });
    expect(app.refreshConceptGraph).not.toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('strengthened', 'success');
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
