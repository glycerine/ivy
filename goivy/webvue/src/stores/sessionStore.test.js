import { beforeEach, describe, expect, it } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useSessionStore } from './sessionStore.js';

describe('sessionStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('starts unauthenticated with no selected project', () => {
    const store = useSessionStore();

    expect(store.authenticated).toBe(false);
    expect(store.selectedProject).toBe(null);
  });

  it('applies a multi-tenant auth view and selects only visible projects', () => {
    const store = useSessionStore();
    store.applyAuthView({
      authenticated: true,
      user: { id: 'user-1', email: 'alice@example.test' },
      accounts: [{ id: 'acct-1', slug: 'acme', displayName: 'Acme' }],
      teams: [{ id: 'team-1', accountId: 'acct-1', slug: 'core', displayName: 'Core' }],
      projects: [{ id: 'project-1', accountId: 'acct-1', slug: 'client-server', displayName: 'Client/server' }],
      roles: { 'project-1': 'admin' },
      passkey: { registered: true },
    });

    store.selectProject('project-1');

    expect(store.authenticated).toBe(true);
    expect(store.accounts).toHaveLength(1);
    expect(store.teams).toHaveLength(1);
    expect(store.selectedProject.displayName).toBe('Client/server');
    expect(store.projectRole('project-1')).toBe('admin');
    expect(store.passkey.registered).toBe(true);
  });

  it('rejects selecting a project that is not in the authorized project list', () => {
    const store = useSessionStore();
    store.applyAuthView({
      authenticated: true,
      user: { id: 'user-1', email: 'alice@example.test' },
      projects: [{ id: 'project-1', accountId: 'acct-1', slug: 'client-server', displayName: 'Client/server' }],
      roles: { 'project-1': 'write' },
    });

    expect(() => store.selectProject('project-2')).toThrow('project project-2 is not visible');
    expect(store.selectedProject).toBe(null);
  });

  it('clears the selected project when the session view no longer grants it', () => {
    const store = useSessionStore();
    store.applyAuthView({
      authenticated: true,
      projects: [{ id: 'project-1', accountId: 'acct-1', slug: 'client-server', displayName: 'Client/server' }],
      roles: { 'project-1': 'admin' },
    });
    store.selectProject('project-1');

    store.applyAuthView({
      authenticated: true,
      projects: [{ id: 'project-2', accountId: 'acct-2', slug: 'other', displayName: 'Other' }],
      roles: { 'project-2': 'read' },
    });

    expect(store.selectedProject).toBe(null);
  });
});
