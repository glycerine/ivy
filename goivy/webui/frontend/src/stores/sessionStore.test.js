import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useEngineStore } from './engineStore.js';
import { useSessionStore } from './sessionStore.js';

describe('sessionStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('creates sessions through the selected engine', async () => {
    const fakeEngine = {
      createSession: vi.fn(async () => 'sid-1'),
      subscribeEvents: vi.fn(),
    };
    useEngineStore().setEngine('fake', fakeEngine);

    const session = useSessionStore();
    await expect(session.createSession()).resolves.toBe('sid-1');

    expect(fakeEngine.createSession).toHaveBeenCalled();
    expect(session.sessionId).toBe('sid-1');
    expect(session.status).toBe('Ready');
  });
});
