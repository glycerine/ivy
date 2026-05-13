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

  it('accepts graph snapshots by explicit sheet id', () => {
    const graphs = useGraphStore();
    const elements = [{ group: 'nodes', data: { id: 'state_0', label: '0' } }];

    graphs.applyGraphSnapshot('sheet-2', 'arg', { elements });

    expect(graphs.argBySheet['sheet-2']).toMatchObject({
      elements,
      positions: null,
      updatedAt: Date.parse('2026-01-01T00:00:00Z'),
    });
  });
});
