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
      connectEvents: vi.fn(),
    };
    useEngineStore().setEngine('fake', fakeEngine);

    const session = useSessionStore();
    await expect(session.createSession()).resolves.toBe('sid-1');

    expect(fakeEngine.createSession).toHaveBeenCalled();
    expect(session.sessionId).toBe('sid-1');
    expect(session.sessionDisplay).toBe('Session: sid-1');
    expect(session.status).toBe('Ready');
  });

  it('tracks loading overlay state', () => {
    const session = useSessionStore();

    session.setMode('abstract');
    expect(session.mode).toBe('abstract');

    session.setMode('not-a-mode');
    expect(session.mode).toBe('pdr');

    session.setLoadedFile('client.ivy', '/models/client.ivy');
    expect(session.loadedFileDisplay).toBe('/models/client.ivy');
    expect(session.loadedFileTitle).toBe('/models/client.ivy');

    session.setSessionId('session-7');
    expect(session.sessionId).toBe('session-7');
    expect(session.sessionDisplay).toBe('Session: session-7');

    session.showLoading('Checking...');
    expect(session.loading).toBe(true);
    expect(session.loadingMessage).toBe('Checking...');

    session.hideLoading();
    expect(session.loading).toBe(false);

    session.setSaveAsNoticeVisible(true);
    expect(session.saveAsNoticeVisible).toBe(true);
  });
});
