import { describe, expect, it, vi } from 'vitest';
import { executeConceptNodeAction, materializeEdge, removeConcept } from './conceptActionService.js';

describe('conceptActionService', () => {
  it('dispatches named concept node actions to high-level workflows', async () => {
    const app = {
      removeConcept: vi.fn(),
    };

    await executeConceptNodeAction(app, { obj: 'client=0' }, { id: 'remove' });

    expect(app.removeConcept).toHaveBeenCalledWith('client=0');
  });

  it('materializes edge objects through the API and refreshes', async () => {
    const app = {
      api: {
        materializeEdge: vi.fn(),
      },
      refreshConceptGraph: vi.fn(),
      controls: { setStatus: vi.fn() },
    };

    await materializeEdge(app, {
      obj: 'link',
      source_obj: 'client=0',
      target_obj: 'server=0',
    }, true);

    expect(app.api.materializeEdge).toHaveBeenCalledWith('link', 'client=0', 'server=0', true);
    expect(app.refreshConceptGraph).toHaveBeenCalled();
  });

  it('removes concepts through the API', async () => {
    const app = {
      api: { removeConcept: vi.fn() },
      refreshConceptGraph: vi.fn(),
      controls: { setStatus: vi.fn() },
    };

    await removeConcept(app, 'server=0');

    expect(app.api.removeConcept).toHaveBeenCalledWith('server=0');
    expect(app.controls.setStatus).toHaveBeenLastCalledWith('Concept removed', 'success');
  });
});
