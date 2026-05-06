import { beforeEach, describe, expect, it, vi } from 'vitest';
import StateRelationsPane from './StateRelationsPane.vue';
import { useStateRelationsStore } from '../../stores/stateRelationsStore.js';
import { createFreshPinia } from '../../test/pinia.js';
import { mountWithPinia } from '../../test/mount.js';

describe('StateRelationsPane', () => {
  beforeEach(() => {
    createFreshPinia();
  });

  it('renders relation rows from Pinia and dispatches checkbox changes', async () => {
    const stateRelations = useStateRelationsStore();
    const onToggle = vi.fn();
    stateRelations.setRows([
      { name: 'link(X,Y)', checked: { all_to_all: false, edge_unknown: true, none_to_none: false, transitive: false } },
    ], onToggle);

    const wrapper = mountWithPinia(StateRelationsPane);

    expect(wrapper.find('#state-label').text()).toBe('State: 0');
    expect(wrapper.find('.name-col a').text()).toBe('link(X,Y)');
    const checkboxes = wrapper.findAll('input[type="checkbox"]');
    expect(checkboxes.map((box) => box.element.checked)).toEqual([false, true, false, false]);

    await checkboxes[0].setValue(true);

    expect(stateRelations.rows[0].checked.all_to_all).toBe(true);
    expect(onToggle).toHaveBeenCalledWith('link(X,Y)', 'all_to_all', true);
  });

  it('shows a placeholder when no relations are loaded', () => {
    const stateRelations = useStateRelationsStore();
    stateRelations.setRows([]);

    const wrapper = mountWithPinia(StateRelationsPane);

    expect(wrapper.find('.state-relations-placeholder').text()).toBe('No relations loaded');
  });
});
