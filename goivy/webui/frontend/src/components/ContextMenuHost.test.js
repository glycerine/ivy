import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import ContextMenuHost from './ContextMenuHost.vue';
import { useContextMenuStore } from '../stores/contextMenuStore.js';

describe('ContextMenuHost', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('renders context menu items from the store and dispatches callbacks', async () => {
    const callback = vi.fn();
    const store = useContextMenuStore();
    store.show(12, 34, [
      { header: 'Actions' },
      { name: 'Split', id: 'split', callback },
    ]);

    const wrapper = mount(ContextMenuHost);

    expect(wrapper.find('#context-menu').attributes('style')).toContain('display: block');
    expect(wrapper.text()).toContain('Actions');
    await wrapper.find('[data-action-id="split"]').trigger('click');

    expect(callback).toHaveBeenCalled();
    expect(store.visible).toBe(false);
  });
});
