import { describe, expect, it, vi } from 'vitest';
import {
  executeAndRefresh,
  exportConjecture,
  refreshConceptGraph,
  rememberGraph,
  runAction,
} from './analysisActionService.ts';

describe('analysisActionService', () => {
  it('runs backend actions with loading and structured results', async () => {
    const app = {
      api: {
        executeAction: vi.fn(async () => ({ ok: true })),
      },
      controls: {
        showLoading: vi.fn(),
        hideLoading: vi.fn(),
        setStatus: vi.fn(),
      },
    };

    await expect(runAction(app, 'undo', {}, {})).resolves.toEqual({ ok: true, result: { ok: true } });
    expect(app.controls.showLoading).toHaveBeenCalledWith('Running: undo...');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Done: undo', 'success');
    expect(app.controls.hideLoading).toHaveBeenCalled();
  });

  it('refreshes concept graph and state relations from the backend', async () => {
    const concept = { elements: [{ data: { id: 'n' } }], positions: { n: { x: 1, y: 2 } } };
    const app = {
      selectedArgNode: '0',
      api: {
        getConceptGraph: vi.fn(async () => concept),
      },
      conceptGraph: {
        update: vi.fn(),
      },
      populateStateCheckboxes: vi.fn(),
    };

    await refreshConceptGraph(app);

    expect(app.api.getConceptGraph).toHaveBeenCalledWith('0');
    expect(app.conceptGraph.update).toHaveBeenCalledWith(concept.elements, concept.positions);
    expect(app.populateStateCheckboxes).toHaveBeenCalledWith(concept);
  });

  it('executes simple refresh actions with expected status messages', async () => {
    const app = {
      api: {
        executeAction: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      controls: {
        setStatus: vi.fn(),
      },
    };

    await executeAndRefresh(app, {
      action: 'redo',
      running: 'Redo...',
      success: 'Redo complete',
      failure: 'Redo failed',
    });

    expect(app.api.executeAction).toHaveBeenCalledWith('redo', {});
    expect(app.refreshConceptGraph).toHaveBeenCalled();
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Redo complete', 'success');
  });

  it('prompts for a graph name before remembering', async () => {
    const app = {
      activeSheetId: 'sheet-1',
      entryDialog: vi.fn(async () => 'goal-a'),
      api: {
        executeAction: vi.fn(async () => ({ status: 'ok' })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await rememberGraph(app);

    expect(app.entryDialog).toHaveBeenCalledWith('Remember graph', 'Enter a name for this goal:', '', { okLabel: 'Remember' });
    expect(app.api.executeAction).toHaveBeenCalledWith('remember', { name: 'goal-a', sheet_id: 'sheet-1' });
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Graph remembered', 'success');
  });

  it('requests DOT export for the active sheet', async () => {
    const app = {
      activeSheetId: 'sheet-2',
      downloadTextFile: vi.fn(),
      api: {
        executeAction: vi.fn(async () => ({
          content: 'digraph concept_graph {\n}\n',
          filename: 'concept_graph.dot',
        })),
      },
      controls: {
        setStatus: vi.fn(),
      },
    };

    await exportConjecture(app, { win: {}, doc: document });

    expect(app.api.executeAction).toHaveBeenCalledWith('export', { sheet_id: 'sheet-2' });
    expect(app.downloadTextFile).toHaveBeenCalledWith('concept_graph.dot', 'digraph concept_graph {\n}\n', 'text/vnd.graphviz');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Graph exported', 'success');
  });
});
