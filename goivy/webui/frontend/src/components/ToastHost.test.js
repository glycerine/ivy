import { beforeEach, describe, expect, it, vi } from 'vitest';
import ToastHost from './ToastHost.vue';
import { useToastStore } from '../stores/toastStore.js';
import { mountWithPinia } from '../test/mount.js';

describe('ToastHost', () => {
  beforeEach(() => {
    vi.useRealTimers();
  });

  it('renders toast notifications from Pinia and removes them on click', async () => {
    const wrapper = mountWithPinia(ToastHost);
    const toastStore = useToastStore();

    const id = toastStore.show('Connection lost', 'error', {
      persistent: true,
      className: 'ivy-toast-floating',
    });
    await wrapper.vm.$nextTick();

    const toast = wrapper.find('.ivy-toast');
    expect(toast.text()).toBe('Connection lost');
    expect(toast.classes()).toContain('ivy-toast-error');
    expect(toast.classes()).toContain('ivy-toast-floating');

    await toast.trigger('click');

    expect(toastStore.items.find((item) => item.id === id)).toBeUndefined();
  });
});
