import { mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import SessionOverlayHost from './SessionOverlayHost.vue';
import { useSessionStore } from '../stores/sessionStore.js';

describe('SessionOverlayHost', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('renders save notices and loading messages from session state', async () => {
    const session = useSessionStore();
    const wrapper = mount(SessionOverlayHost);

    session.setSaveAsNoticeVisible(true);
    session.showLoading('Working...');
    await wrapper.vm.$nextTick();

    expect(wrapper.find('#save-as-explain-notice').isVisible()).toBe(true);
    expect(wrapper.find('#loading-message').text()).toBe('Working...');
  });
});
