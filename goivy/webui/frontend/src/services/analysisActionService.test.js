import { describe, expect, it, vi } from 'vitest';
import { executeAndRefresh, refreshConceptGraph, runAction } from './analysisActionService.js';

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
});
