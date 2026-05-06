import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useGraphStore } from './graphStore.js';

describe('graphStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-01T00:00:00Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('accepts graph snapshots by sheet and tracks selected ARG nodes', () => {
    const graphs = useGraphStore();
    const elements = [{ group: 'nodes', data: { id: 'state_0', label: '0' } }];

    graphs.setActiveSheet('sheet-2');
    graphs.applyGraphSnapshot('sheet-2', 'arg', { elements });
    graphs.selectArgNode('state_0', 'sheet-2');

    expect(graphs.activeSheetId).toBe('sheet-2');
    expect(graphs.argBySheet['sheet-2']).toMatchObject({
      elements,
      positions: null,
      updatedAt: Date.parse('2026-01-01T00:00:00Z'),
    });
    expect(graphs.selectedArgNode).toBe('state_0');
    expect(graphs.selectedArgNodeBySheet['sheet-2']).toBe('state_0');
  });
});
