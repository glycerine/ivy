import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia } from 'pinia';
import App from './App.vue';
import { installAppServices, resetAppServicesForTests } from './services/appServices.js';

describe('Vue app shell', () => {
  let startServices;

  beforeEach(() => {
    resetAppServicesForTests();
    startServices = vi.fn(async () => undefined);
    installAppServices({ start: startServices });
  });

  it('renders the four user-facing column headers and starts app services', async () => {
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
    expect(document.getElementById('sheet-1').classList.contains('active')).toBe(true);
    expect(startServices).toHaveBeenCalledTimes(1);

    wrapper.unmount();
  });
});
