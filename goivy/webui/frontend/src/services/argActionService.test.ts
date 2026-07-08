import { describe, expect, it, vi } from 'vitest';
import { executeArgEdgeAction, executeArgNodeAction, prepareArgNodeActionArgs } from './argActionService.ts';

describe('argActionService', () => {
  it('prompts for missing conjecture choices', async () => {
    const app = {
      api: {
        argNodeAction: vi.fn(async () => ({ choices: ['c1'] })),
      },
      listboxDialog: vi.fn(async () => 'c1'),
    };
    const args = await prepareArgNodeActionArgs(app, { id: 'n0' }, 'try_conjecture', {}, 'sheet-1');

    expect(args.conjecture).toBe('c1');
  });

  it('prompts for missing remembered-goal choices', async () => {
    const app = {
      api: {
        argNodeAction: vi.fn(async () => ({ choices: [{ label: 'goal-b', value: 'goal-b' }] })),
      },
      listboxDialog: vi.fn(async () => 'goal-b'),
    };

    const args = await prepareArgNodeActionArgs(app, { id: 'state_0' }, 'try_remembered', { sheet_id: 'sheet-1' }, 'sheet-1');

    expect(args).toEqual({ sheet_id: 'sheet-1', goal: 'goal-b' });
    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'try_remembered_choices', { sheet_id: 'sheet-1' });
  });

  it('prompts for ARG BMC bound and error condition', async () => {
    const app = {
      currentBound: 3,
      integerDialog: vi.fn(async () => 5),
      entryDialog: vi.fn(async () => 'bad(X)'),
    };

    const args = await prepareArgNodeActionArgs(app, { id: 'state_0' }, 'bmc', { sheet_id: 'sheet-1' }, 'sheet-1');

    expect(app.integerDialog).toHaveBeenCalledWith('Bounded check', 'Enter bound:', 3, {
      min: 0,
      okLabel: 'Check',
    });
    expect(app.entryDialog).toHaveBeenCalledWith(
      'Bounded check',
      'Enter error condition:',
      'true',
      { okLabel: 'Check' },
    );
    expect(app.currentBound).toBe(5);
    expect(args).toEqual({ sheet_id: 'sheet-1', bound: 5, err_cond: 'bad(X)' });
  });

  it('executes ARG node actions and refreshes returned graphs', async () => {
    const app = {
      activeSheetId: 'sheet-1',
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: (node, action, args) => args,
      api: {
        argNodeAction: vi.fn(async () => ({
          arg: { elements: ['arg'], positions: null },
          concept: { elements: ['concept'], positions: null },
        })),
      },
      sheets: {},
      uiDataStore: {
        applyArgSnapshot: vi.fn(),
        applyConceptSnapshot: vi.fn(),
      },
      controls: { setStatus: vi.fn() },
    };

    await executeArgNodeAction(app, { id: 'n0' }, { id: 'extend' }, 'sheet-1');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('n0', 'extend', { sheet_id: 'sheet-1' });
    expect(app.uiDataStore.applyArgSnapshot).toHaveBeenCalledWith('sheet-1', { elements: ['arg'], positions: null });
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-1', { elements: ['concept'], positions: null });
  });

  it('loads source and concept returned by ARG try-conjecture actions', async () => {
    const app = {
      activeSheetId: 'sheet-1',
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: (node, action, args) => args,
      api: {
        argNodeAction: vi.fn(async () => ({
          view: 'concept',
          message: 'Conjecture goal opened.',
          source: 'ivy source',
          lineno: 7,
          file: 'm.ivy',
          concept: { elements: ['concept'], positions: null },
        })),
      },
      sheets: {},
      uiDataStore: {
        applyConceptSnapshot: vi.fn(),
      },
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
      openSourceBrowser: vi.fn(),
      setEditorContent: vi.fn(),
      scrollEditorToLine: vi.fn(),
    };

    const result = await executeArgNodeAction(
      app,
      { id: 'state_0' },
      { id: 'try_conjecture', args: { conjecture: 'c' } },
      'sheet-1',
    );

    expect(result?.view).toBe('concept');
    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'try_conjecture', {
      sheet_id: 'sheet-1',
      conjecture: 'c',
    });
    expect(app.uiDataStore.applyConceptSnapshot).toHaveBeenCalledWith('sheet-1', { elements: ['concept'], positions: null });
    expect(app.openSourceBrowser).toHaveBeenCalledWith(expect.objectContaining({ source: 'ivy source', lineno: 7, file: 'm.ivy' }));
    expect(app.setEditorContent).not.toHaveBeenCalled();
    expect(app.scrollEditorToLine).not.toHaveBeenCalled();
    expect(app.controls.showInfo).toHaveBeenCalledWith('Source: m.ivy line 7', '');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Conjecture goal opened.', 'success');
  });

  it('shows try-conjecture message views in an ivyweb dialog', async () => {
    const app = {
      activeSheetId: 'sheet-1',
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: (node, action, args) => args,
      api: {
        argNodeAction: vi.fn(async () => ({
          view: 'message',
          reachable: false,
          message: 'The condition is unreachable along the given path.',
          source: 'ivy source',
          lineno: 3,
          file: 'm.ivy',
        })),
      },
      okDialog: vi.fn(),
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
      openSourceBrowser: vi.fn(),
      setEditorContent: vi.fn(),
      scrollEditorToLine: vi.fn(),
    };

    await executeArgNodeAction(
      app,
      { id: 'state_0' },
      { id: 'try_conjecture', args: { conjecture: 'c' } },
      'sheet-1',
    );

    expect(app.openSourceBrowser).toHaveBeenCalledWith(expect.objectContaining({ source: 'ivy source', lineno: 3, file: 'm.ivy' }));
    expect(app.setEditorContent).not.toHaveBeenCalled();
    expect(app.scrollEditorToLine).not.toHaveBeenCalled();
    expect(app.okDialog).toHaveBeenCalledWith('ivyweb', 'The condition is unreachable along the given path.');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('The condition is unreachable along the given path.', 'success');
  });

  it('opens try-conjecture trace views through the Python-style View dialog', async () => {
    const traceArg = { elements: [{ group: 'nodes', data: { id: 'state_0' } }] };
    const app = {
      activeSheetId: 'sheet-1',
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: (node, action, args) => args,
      api: {
        argNodeAction: vi.fn(async () => ({
          view: 'trace',
          reachable: true,
          message: 'The condition is reachable along the given path.',
          source: 'ivy source',
          lineno: 4,
          file: 'm.ivy',
          trace_arg: traceArg,
          trace_sheet_id: 'sheet-4',
          trace_label: 'Sheet 4',
        })),
      },
      textDialog: vi.fn(async () => ''),
      openARGSheet: vi.fn(),
      setUIMode: vi.fn(),
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
      openSourceBrowser: vi.fn(),
      setEditorContent: vi.fn(),
      scrollEditorToLine: vi.fn(),
    };

    await executeArgNodeAction(
      app,
      { id: 'state_0' },
      { id: 'try_conjecture', args: { conjecture: 'c' } },
      'sheet-1',
    );

    expect(app.openSourceBrowser).toHaveBeenCalledWith(expect.objectContaining({ source: 'ivy source', lineno: 4, file: 'm.ivy' }));
    expect(app.setEditorContent).not.toHaveBeenCalled();
    expect(app.scrollEditorToLine).not.toHaveBeenCalled();
    expect(app.textDialog).toHaveBeenCalledWith(
      'ivyweb',
      'The condition is reachable along the given path.',
      '',
      expect.objectContaining({ okLabel: 'View', cancel: true, primaryFirst: true }),
    );
    expect(app.setUIMode).toHaveBeenCalledWith('reachability');
    expect(app.openARGSheet).toHaveBeenCalledWith('Sheet 4', traceArg, 'sheet-4', {
      reachabilityOnly: true,
      visualOnly: false,
    });
  });

  it('shows the Python-style closed-node dialog for Extend', async () => {
    const app = {
      activeSheetId: 'sheet-1',
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: (node, action, args) => args,
      api: {
        argNodeAction: vi.fn(async () => ({
          closed: true,
          result: 'closed',
          message: 'State 0 is closed.',
          arg: { elements: ['arg'], positions: null },
        })),
      },
      sheets: {},
      uiDataStore: {
        applyArgSnapshot: vi.fn(),
      },
      okDialog: vi.fn(),
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
    };

    const result = await executeArgNodeAction(app, { id: 'state_0' }, { id: 'find_extension' }, 'sheet-1');

    expect(result?.closed).toBe(true);
    expect(app.okDialog).toHaveBeenCalledWith('ivyweb', 'State 0 is closed.');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('State 0 is closed.', 'warning');
  });

  it('passes the current mode and offers trace viewing for failed node safety', async () => {
    document.body.innerHTML = '<div id="info-content"></div>';
    const traceArg = { elements: [{ group: 'nodes', data: { id: 'state_0' } }] };
    const app = {
      activeSheetId: 'sheet-1',
      getMode: vi.fn(() => 'bounded'),
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: (node, action, args, sheetId) => prepareArgNodeActionArgs(app, node, action, args, sheetId),
      api: {
        argNodeAction: vi.fn(async () => ({
          result: 'fail',
          safe: false,
          message: 'The node is unsafe: View error trace?',
          arg: { elements: ['arg'], positions: null },
          trace_arg: traceArg,
          trace_sheet_id: 'sheet-2',
        })),
      },
      sheets: {},
      uiDataStore: {
        applyArgSnapshot: vi.fn(),
      },
      openARGSheet: vi.fn(),
      setUIMode: vi.fn(),
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
    };

    await executeArgNodeAction(app, { id: 'state_0' }, { id: 'check_safety' }, 'sheet-1');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'check_safety', {
      sheet_id: 'sheet-1',
      mode: 'bounded',
    });
    expect(app.controls.showInfo).toHaveBeenCalledWith('Safety Check', 'The node is unsafe: View error trace?');
    document.querySelector('[data-check-view-trace]')?.click();
    expect(app.setUIMode).toHaveBeenCalledWith('reachability');
    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', traceArg, 'sheet-2', {
      reachabilityOnly: true,
      visualOnly: false,
    });
  });

  it('opens ARG BMC counterexample traces through the Python-style View dialog', async () => {
    const traceArg = { elements: [{ group: 'nodes', data: { id: 'state_0' } }] };
    const app = {
      activeSheetId: 'sheet-1',
      isVisualOnlySheet: vi.fn(() => false),
      prepareArgNodeActionArgs: (node, action, args, sheetId) => prepareArgNodeActionArgs(app, node, action, args, sheetId),
      api: {
        argNodeAction: vi.fn(async () => ({
          result: 'fail',
          reachable: true,
          message: 'BMC with bound 5 found a counterexample to:\nbad(X)',
          trace_arg: traceArg,
          trace_sheet_id: 'sheet-3',
          trace_label: 'Sheet 3',
        })),
      },
      currentBound: 5,
      textDialog: vi.fn(async () => ''),
      openARGSheet: vi.fn(),
      setUIMode: vi.fn(),
      controls: { setStatus: vi.fn(), showLoading: vi.fn(), hideLoading: vi.fn() },
    };

    await executeArgNodeAction(app, { id: 'state_0' }, { id: 'bmc', args: { bound: 5, err_cond: 'bad(X)' } }, 'sheet-1');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'bmc', {
      sheet_id: 'sheet-1',
      bound: 5,
      err_cond: 'bad(X)',
    });
    expect(app.textDialog).toHaveBeenCalledWith(
      'ivyweb',
      'BMC with bound 5 found a counterexample to:',
      'bad(X)',
      expect.objectContaining({ okLabel: 'View', cancel: true, primaryFirst: true }),
    );
    expect(app.setUIMode).toHaveBeenCalledWith('reachability');
    expect(app.openARGSheet).toHaveBeenCalledWith('Sheet 3', traceArg, 'sheet-3', {
      reachabilityOnly: true,
      visualOnly: false,
    });
  });

  it('highlights existing editor source returned by ARG edge view-source actions', async () => {
    const sourceResult = { source: 'action a', lineno: 7, file: 'm.ivy' };
    const app = {
      activeSheetId: 'sheet-1',
      api: {
        argNodeAction: vi.fn(async () => sourceResult),
      },
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
      openSourceBrowser: vi.fn(),
      setEditorContent: vi.fn(),
      scrollEditorToLine: vi.fn(),
    };

    await executeArgEdgeAction(app, { source_obj: 's0', target_obj: 's1' }, 'view_source', 'sheet-1');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('s0', 'view_source', { target: 's1', sheet_id: 'sheet-1' });
    expect(app.openSourceBrowser).not.toHaveBeenCalled();
    expect(app.setEditorContent).not.toHaveBeenCalled();
    expect(app.scrollEditorToLine).toHaveBeenCalledWith(7);
    expect(app.controls.showInfo).not.toHaveBeenCalled();
  });

  it('does not create a source browser for edge view-source actions', async () => {
    document.body.innerHTML = [
      '<div class="sheet-content active">',
      '  <textarea id="model-editor">editable model</textarea>',
      '</div>',
    ].join('');
    const app = {
      activeSheetId: 'sheet-1',
      api: {
        argNodeAction: vi.fn(async (_node, _action, args) => (
          args.target === 's1'
            ? { source: 'first edge\nkeep', lineno: 1, file: 'first.ivy' }
            : { source: 'second edge\nupdated', lineno: 2, file: 'second.ivy' }
        )),
      },
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
      openSourceBrowser: vi.fn(),
      setEditorContent: vi.fn(),
      scrollEditorToLine: vi.fn(),
    };

    await executeArgEdgeAction(app, { source_obj: 's0', target_obj: 's1' }, 'view_source', 'sheet-1');
    await executeArgEdgeAction(app, { source_obj: 's0', target_obj: 's2' }, 'view_source', 'sheet-1');

    expect(document.querySelectorAll('[data-source-browser]')).toHaveLength(0);
    expect((document.getElementById('model-editor') as HTMLTextAreaElement).value).toBe('editable model');
    expect(app.openSourceBrowser).not.toHaveBeenCalled();
    expect(app.setEditorContent).not.toHaveBeenCalled();
    expect(app.scrollEditorToLine).toHaveBeenCalledWith(1);
    expect(app.scrollEditorToLine).toHaveBeenCalledWith(2);
    expect(app.controls.showInfo).not.toHaveBeenCalled();
  });
});
