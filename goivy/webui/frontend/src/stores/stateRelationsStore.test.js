import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useStateRelationsStore } from './stateRelationsStore.js';

describe('stateRelationsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('stores rows and dispatches toggle changes', () => {
    const stateRelations = useStateRelationsStore();
    const onToggle = vi.fn();

    stateRelations.setRows([
      { name: 'link(X,Y)', checked: { edge_unknown: true } },
    ], onToggle);
    stateRelations.toggle('link(X,Y)', 'all_to_all', true);

    expect(stateRelations.rows[0].checked.edge_unknown).toBe(true);
    expect(stateRelations.rows[0].checked.all_to_all).toBe(true);
    expect(onToggle).toHaveBeenCalledWith('link(X,Y)', 'all_to_all', true);
  });
});
