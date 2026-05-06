import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia } from 'pinia';
import App from './App.vue';

describe('Vue app shell', () => {
  beforeEach(() => {
    window.startIvyApp = vi.fn();
  });

  it('renders the four user-facing column headers and starts the legacy adapter', async () => {
    const wrapper = mount(App, {
      global: {
        plugins: [createPinia()],
      },
      attachTo: document.body,
    });
    await wrapper.vm.$nextTick();
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain('ARG (Abstract Reachability Graph)');
    expect(wrapper.text()).toContain('Concept graph');
    expect(wrapper.text()).toContain('State/relations');
    expect(wrapper.text()).toContain('Editing:');
    expect(document.getElementById('model-editor')).not.toBeNull();
    expect(window.startIvyApp).toHaveBeenCalledTimes(1);

    wrapper.unmount();
  });
});
