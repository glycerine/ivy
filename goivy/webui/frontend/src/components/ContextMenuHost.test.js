import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import ContextMenuHost from './ContextMenuHost.vue';
import { useContextMenuStore } from '../stores/contextMenuStore.js';

describe('ContextMenuHost', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.restoreAllMocks();
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

  it('keeps a Vue-rendered context menu inside the viewport', async () => {
    vi.spyOn(window, 'innerWidth', 'get').mockReturnValue(320);
    vi.spyOn(window, 'innerHeight', 'get').mockReturnValue(240);
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      x: 280,
      y: 220,
      left: 280,
      top: 220,
      right: 400,
      bottom: 300,
      width: 120,
      height: 80,
      toJSON: () => {},
    });
    const store = useContextMenuStore();

    store.show(280, 220, [{ name: 'Split', id: 'split', callback: vi.fn() }]);
    const wrapper = mount(ContextMenuHost);
    await wrapper.vm.$nextTick();
    await wrapper.vm.$nextTick();

    expect(store.x).toBe(200);
    expect(store.y).toBe(160);
    expect(wrapper.find('#context-menu').attributes('style')).toContain('left: 200px');
    expect(wrapper.find('#context-menu').attributes('style')).toContain('top: 160px');
  });
});
