import { describe, expect, it, vi } from 'vitest';
import { executeArgEdgeAction, executeArgNodeAction, prepareArgNodeActionArgs } from './argActionService.js';

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
      argGraph: { update: vi.fn() },
      conceptGraph: { update: vi.fn() },
      controls: { setStatus: vi.fn() },
    };

    await executeArgNodeAction(app, { id: 'n0' }, { id: 'extend' }, 'sheet-1');

    expect(app.api.argNodeAction).toHaveBeenCalledWith('n0', 'extend', { sheet_id: 'sheet-1' });
    expect(app.argGraph.update).toHaveBeenCalledWith(['arg'], null);
    expect(app.conceptGraph.update).toHaveBeenCalledWith(['concept'], null);
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

    expect(app.setEditorContent).toHaveBeenCalledWith('action a');
    expect(app.scrollEditorToLine).toHaveBeenCalledWith(7);
    expect(app.controls.showInfo).toHaveBeenCalledWith('Source: m.ivy line 7', '');
  });
});
