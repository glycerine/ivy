import { beforeEach, describe, expect, it } from 'vitest';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import App from './App.vue';
import { authClientKey } from './app/dependencies.js';
import { useSessionStore } from './stores/sessionStore.js';

function mountApp(options = {}) {
  return mount(App, {
    global: {
      plugins: [createPinia()],
      provide: {
        ...(options.provide || {}),
      },
    },
  });
}

describe('App', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('starts at email sign-in when unauthenticated', () => {
    const wrapper = mountApp();

    expect(wrapper.find('[aria-label="Sign in"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Ivy workspace"]').exists()).toBe(false);
  });

  it('submits the email sign-up form and shows the neutral sent state', async () => {
    const requests = [];
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async requestEmailLogin(email) {
            requests.push(email);
            return { ok: true };
          },
        },
      },
    });

    await wrapper.find('input[type="email"]').setValue('alice@example.test');
    await wrapper.find('form').trigger('submit');
    await wrapper.vm.$nextTick();

    expect(requests).toEqual(['alice@example.test']);
    expect(wrapper.find('[role="status"]').text()).toContain('Check your email');
    expect(wrapper.find('form').exists()).toBe(false);
  });

  it('shows only authorized projects before entering the workspace', async () => {
    const wrapper = mountApp();
    const session = useSessionStore();
    session.applyAuthView({
      authenticated: true,
      user: { id: 'user-1', email: 'alice@example.test' },
      accounts: [{ id: 'acct-1', slug: 'acme', displayName: 'Acme' }],
      teams: [{ id: 'team-1', accountId: 'acct-1', slug: 'core', displayName: 'Core' }],
      projects: [{ id: 'project-1', accountId: 'acct-1', slug: 'client-server', displayName: 'Client/server' }],
      roles: { 'project-1': 'admin' },
    });
    await wrapper.vm.$nextTick();

    expect(wrapper.find('[aria-label="Project picker"]').text()).toContain('Client/server');
    expect(wrapper.find('[aria-label="Project picker"]').text()).not.toContain('Hidden project');
    expect(wrapper.find('[aria-label="Ivy workspace"]').exists()).toBe(false);
  });

  it('renders the project-scoped workspace only after project selection', async () => {
    const wrapper = mountApp();
    const session = useSessionStore();
    session.applyAuthView({
      authenticated: true,
      projects: [{ id: 'project-1', accountId: 'acct-1', slug: 'client-server', displayName: 'Client/server' }],
      roles: { 'project-1': 'write' },
    });
    session.selectProject('project-1');
    await wrapper.vm.$nextTick();

    expect(wrapper.find('[aria-label="Ivy workspace"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="ARG (Abstract Reachability Graph)"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Concept graph"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="State/relations"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Editing"]').exists()).toBe(true);
  });
});
