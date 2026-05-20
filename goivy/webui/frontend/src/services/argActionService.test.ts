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
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
    };

    await executeArgNodeAction(app, { id: 'state_0' }, { id: 'check_safety' }, 'sheet-1');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('state_0', 'check_safety', {
      sheet_id: 'sheet-1',
      mode: 'bounded',
    });
    expect(app.controls.showInfo).toHaveBeenCalledWith('Safety Check', 'The node is unsafe: View error trace?');
    document.querySelector('[data-check-view-trace]')?.click();
    expect(app.openARGSheet).toHaveBeenCalledWith('Error trace', traceArg, 'sheet-2', {
      reachabilityOnly: true,
      visualOnly: false,
    });
  });

  it('loads source returned by ARG edge view-source actions', async () => {
    const app = {
      activeSheetId: 'sheet-1',
      api: {
        argNodeAction: vi.fn(async () => ({ source: 'action a', lineno: 7, file: 'm.ivy' })),
      },
      controls: { setStatus: vi.fn(), showInfo: vi.fn() },
      setEditorContent: vi.fn(),
      scrollEditorToLine: vi.fn(),
    };

    await executeArgEdgeAction(app, { source_obj: 's0', target_obj: 's1' }, 'view_source', 'sheet-1');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('s0', 'view_source', { target: 's1', sheet_id: 'sheet-1' });
    expect(app.setEditorContent).toHaveBeenCalledWith('action a');
    expect(app.scrollEditorToLine).toHaveBeenCalledWith(7);
    expect(app.controls.showInfo).toHaveBeenCalledWith('Source: m.ivy line 7', '');
  });
});
