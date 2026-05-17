import { beforeEach, describe, expect, it } from 'vitest';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import App from './App.vue';
import { authClientKey, passkeyBrowserKey } from './app/dependencies.js';
import { useSessionStore } from './stores/sessionStore.js';

function mountApp(options = {}) {
  const provided = options.provide || {};
  const authClient = {
    async currentSession() {
      return { authenticated: false };
    },
    async requestEmailLogin() {
      return { ok: true };
    },
    async consumeEmailLogin() {
      return { ok: true };
    },
    async listUnverifiedEmails() {
      return { emails: [] };
    },
    async beginPasskeyRegistration() {
      return { publicKey: { challenge: 'abc' } };
    },
    async finishPasskeyRegistration() {
      return { authenticated: true, passkey: { registered: true } };
    },
    async beginPasskeyLogin() {
      return { publicKey: { challenge: 'abc' } };
    },
    async finishPasskeyLogin() {
      return { authenticated: false };
    },
    ...(provided[authClientKey] || {}),
  };
  const passkeyBrowser = {
    isSupported() {
      return false;
    },
    async create() {
      return { id: 'credential-1' };
    },
    async authenticate() {
      return { id: 'credential-1' };
    },
    ...(provided[passkeyBrowserKey] || {}),
  };
  return mount(App, {
    global: {
      plugins: [createPinia()],
      provide: {
        ...provided,
        [authClientKey]: authClient,
        [passkeyBrowserKey]: passkeyBrowser,
      },
    },
  });
}

async function settleMountedAsync(wrapper) {
  for (let i = 0; i < 8; i += 1) {
    await wrapper.vm.$nextTick();
    await Promise.resolve();
  }
}

describe('App', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    window.history.pushState({}, '', '/');
  });

  it('starts at email sign-in when unauthenticated', async () => {
    const wrapper = mountApp();
    await settleMountedAsync(wrapper);

    expect(wrapper.find('[aria-label="Sign in"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Ivy workspace"]').exists()).toBe(false);
  });

  it('tries a passkey during startup before rendering the email sign-in fallback', async () => {
    const calls = [];
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async currentSession() {
            calls.push('currentSession');
            return { authenticated: false };
          },
          async beginPasskeyLogin() {
            calls.push('beginPasskeyLogin');
            return { publicKey: { challenge: 'challenge' } };
          },
          async finishPasskeyLogin() {
            calls.push('finishPasskeyLogin');
            return {
              authenticated: true,
              user: { id: 'user-1', email: 'alice@example.test' },
              projects: [{ id: 'project-1', displayName: 'Client/server' }],
              roles: { 'project-1': 'admin' },
              passkey: { registered: true },
            };
          },
        },
        [passkeyBrowserKey]: {
          isSupported() {
            return true;
          },
          async authenticate() {
            calls.push('browserPasskey');
            return { id: 'credential-1' };
          },
        },
      },
    });
    expect(wrapper.find('[aria-label="Loading session"]').text()).toContain('Loading your account');

    await settleMountedAsync(wrapper);

    expect(calls).toEqual(['currentSession', 'beginPasskeyLogin', 'browserPasskey', 'finishPasskeyLogin']);
    expect(wrapper.find('[aria-label="Sign in"]').exists()).toBe(false);
    expect(wrapper.find('[aria-label="Project picker"]').exists()).toBe(true);
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
    await settleMountedAsync(wrapper);

    await wrapper.find('input[type="email"]').setValue('alice@example.test');
    await wrapper.find('form').trigger('submit');
    await wrapper.vm.$nextTick();

    expect(requests).toEqual(['alice@example.test']);
    expect(wrapper.find('[role="status"]').text()).toContain('Check your email');
    expect(wrapper.find('form').exists()).toBe(false);
  });

  it('shows an expired-link retry message on the sign-in page without a redirect hop or query flag', async () => {
    const consumedTokens = [];
    window.history.pushState({}, '', '/auth/email/continue#token=expired-token');
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async consumeEmailLogin(token) {
            consumedTokens.push(token);
            throw new Error('expired');
          },
        },
      },
    });
    await settleMountedAsync(wrapper);

    expect(consumedTokens).toEqual(['expired-token']);
    expect(window.location.pathname).toBe('/');
    expect(window.location.search).toBe('');
    expect(window.location.hash).toBe('');
    expect(wrapper.find('[aria-label="Sign in"]').exists()).toBe(true);
    expect(wrapper.find('[role="alert"]').text()).toContain('sign-in link is invalid or expired');
    expect(wrapper.find('form').exists()).toBe(true);
  });

  it('shows only authorized projects before entering the workspace', async () => {
    const wrapper = mountApp();
    await settleMountedAsync(wrapper);
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
    await settleMountedAsync(wrapper);
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

  it('shows the admin unverified email dashboard with sign-in links', async () => {
    window.history.pushState({}, '', '/admin');
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async listUnverifiedEmails() {
            return {
              emails: [
                {
                  email: 'alice@example.test',
                  loginUrl: 'http://127.0.0.1:18080/auth/email/continue#token=abc',
                  expiresAt: '2026-05-07T04:10:00Z',
                  expired: false,
                  usedAt: null,
                },
              ],
            };
          },
        },
      },
    });

    await wrapper.vm.$nextTick();
    await wrapper.vm.$nextTick();

    expect(wrapper.find('[aria-label="Admin dashboard"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Unverified emails"]').text()).toContain('alice@example.test');
    expect(wrapper.find('a').attributes('href')).toBe('http://127.0.0.1:18080/auth/email/continue#token=abc');
  });

  it('shows a verified landing page after the magic link session cookie is accepted', async () => {
    window.history.pushState({}, '', '/verified');
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async currentSession() {
            return {
              authenticated: true,
              user: { id: 'user-1', email: 'alice@example.test' },
              projects: [{ id: 'project-1', displayName: 'Client/server' }],
              roles: { 'project-1': 'admin' },
            };
          },
        },
      },
    });

    await settleMountedAsync(wrapper);

    expect(wrapper.find('[aria-label="Email verified"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Email verified"]').text()).toContain('alice@example.test');
    expect(wrapper.find('[aria-label="Sign in"]').exists()).toBe(false);
  });

  it('asks a newly verified user if they want to create a passkey', async () => {
    window.history.pushState({}, '', '/verified');
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async currentSession() {
            return {
              authenticated: true,
              user: { id: 'user-1', email: 'alice@example.test' },
              projects: [{ id: 'project-1', displayName: 'Client/server' }],
              roles: { 'project-1': 'admin' },
              passkey: { registered: false },
            };
          },
        },
        [passkeyBrowserKey]: {
          isSupported() {
            return true;
          },
        },
      },
    });

    await settleMountedAsync(wrapper);

    expect(wrapper.find('[aria-label="Email verified"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Passkey setup"]').text()).toContain('Use a passkey next time');
  });

  it('creates a passkey from the post-login prompt and updates the session view', async () => {
    const calls = [];
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async currentSession() {
            return {
              authenticated: true,
              user: { id: 'user-1', email: 'alice@example.test' },
              projects: [{ id: 'project-1', displayName: 'Client/server' }],
              roles: { 'project-1': 'admin' },
              passkey: { registered: false },
            };
          },
          async beginPasskeyRegistration() {
            calls.push('beginPasskeyRegistration');
            return { publicKey: { challenge: 'challenge' } };
          },
          async finishPasskeyRegistration(credential) {
            calls.push(['finishPasskeyRegistration', credential.id]);
            return {
              authenticated: true,
              user: { id: 'user-1', email: 'alice@example.test' },
              projects: [{ id: 'project-1', displayName: 'Client/server' }],
              roles: { 'project-1': 'admin' },
              passkey: { registered: true },
            };
          },
        },
        [passkeyBrowserKey]: {
          isSupported() {
            return true;
          },
          async create() {
            calls.push('browserCreate');
            return { id: 'credential-1' };
          },
        },
      },
    });
    await settleMountedAsync(wrapper);

    await wrapper.find('[aria-label="Passkey setup"] button').trigger('click');
    await settleMountedAsync(wrapper);

    expect(calls).toEqual(['beginPasskeyRegistration', 'browserCreate', ['finishPasskeyRegistration', 'credential-1']]);
    expect(useSessionStore().passkey.registered).toBe(true);
    expect(wrapper.find('[aria-label="Passkey setup"]').exists()).toBe(false);
  });

  it('consumes a valid magic link in place and shows the verified landing page', async () => {
    const consumedTokens = [];
    window.history.pushState({}, '', '/auth/email/continue#token=valid-token');
    const wrapper = mountApp({
      provide: {
        [authClientKey]: {
          async consumeEmailLogin(token) {
            consumedTokens.push(token);
            return { ok: true };
          },
          async currentSession() {
            return {
              authenticated: true,
              user: { id: 'user-1', email: 'alice@example.test' },
              projects: [{ id: 'project-1', displayName: 'Client/server' }],
              roles: { 'project-1': 'admin' },
            };
          },
        },
      },
    });

    await settleMountedAsync(wrapper);

    expect(consumedTokens).toEqual(['valid-token']);
    expect(window.location.pathname).toBe('/verified');
    expect(window.location.search).toBe('');
    expect(window.location.hash).toBe('');
    expect(wrapper.find('[aria-label="Email verified"]').exists()).toBe(true);
    expect(wrapper.find('[aria-label="Email verified"]').text()).toContain('alice@example.test');
  });
});
