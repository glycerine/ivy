import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import StatusBar from './StatusBar.vue';
import { useSessionStore } from '../stores/sessionStore.js';

describe('StatusBar', () => {
  it('renders status text and level from Pinia', async () => {
    const pinia = createPinia();
    setActivePinia(pinia);
    const session = useSessionStore();
    const wrapper = mount(StatusBar, {
      global: {
        plugins: [pinia],
      },
    });

    session.setStatus('Check FAILED - counterexample found', 'error');
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toBe('Check FAILED - counterexample found');
    expect(wrapper.classes()).toContain('error');
  });
});
