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

  it('exports and applies toggle snapshots without scraping the table DOM', () => {
    const stateRelations = useStateRelationsStore();

    stateRelations.setRows([
      { name: 'link(X,Y)', checked: { all_to_all: true, edge_unknown: false, none_to_none: true } },
    ]);

    expect(stateRelations.toggleSnapshot).toEqual({
      'link(X,Y)|all_to_all': true,
      'link(X,Y)|edge_unknown': false,
      'link(X,Y)|none_to_none': true,
      'link(X,Y)|transitive': false,
    });
    expect(stateRelations.visibilitySnapshot).toEqual({
      edges: {
        'link(X,Y)': {
          all_to_all: true,
          edge_unknown: false,
          none_to_none: true,
          transitive: false,
        },
      },
      labels: {
        'link(X,Y)': {
          node_necessarily: true,
          node_maybe: false,
          node_necessarily_not: true,
        },
      },
    });

    stateRelations.applyToggleSnapshot({
      'link(X,Y)|all_to_all': false,
      'link(X,Y)|edge_unknown': true,
      'link(X,Y)|transitive': true,
    });

    expect(stateRelations.rows[0].checked).toEqual({
      all_to_all: false,
      edge_unknown: true,
      none_to_none: true,
      transitive: true,
    });
  });
});
