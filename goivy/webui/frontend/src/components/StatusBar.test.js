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
    session.setSessionId('s1');
    await wrapper.vm.$nextTick();

    expect(wrapper.find('.status-message').text()).toBe('Check FAILED - counterexample found');
    expect(wrapper.find('#session-id').text()).toBe('Session: s1');
    expect(wrapper.classes()).toContain('error');
  });
});
