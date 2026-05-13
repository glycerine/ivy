import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  connectSessionEvents,
  createIvyApi,
  createSession,
  updateSessionDisplay,
} from './sessionService.js';

afterEach(() => {
  vi.restoreAllMocks();
  delete window.__IVY_ENGINE__;
});

describe('sessionService', () => {
  it('uses the injected browser engine when available', () => {
    const api = { sessionId: 's1' };
    window.__IVY_ENGINE__ = api;

    expect(createIvyApi()).toBe(api);
    delete window.__IVY_ENGINE__;
  });

  it('falls back to the provided Ivy API factory', () => {
    const api = { sessionId: 'fallback' };

    expect(createIvyApi({ fallbackApiFactory: () => api })).toBe(api);
  });

  it('creates sessions and reports failures through controls', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    const okApi = {
      sessionId: 's1',
      createSession: vi.fn(),
    };
    await expect(createSession(okApi)).resolves.toBe('s1');

    const err = new Error('boom');
    const badApi = {
      createSession: vi.fn(() => Promise.reject(err)),
    };
    const controls = {
      setStatus: vi.fn(),
    };
    await expect(createSession(badApi, { controls })).rejects.toThrow('boom');
    expect(controls.setStatus).toHaveBeenCalledWith('Failed to create session: boom', 'error');
  });

  it('updates the session display in the DOM', () => {
    const el = document.createElement('div');
    el.id = 'session-id';
    document.body.appendChild(el);
    expect(updateSessionDisplay('xyz', { doc: document })).toBe(true);
    expect(el.textContent).toBe('Session: xyz');
    el.remove();
  });

  it('connects SSE events and connection-lost callbacks', () => {
    const onEvent = vi.fn();
    const onLost = vi.fn();
    const api = {
      sessionId: 's1',
      connectEvents: vi.fn(),
    };

    expect(connectSessionEvents(api, onEvent, onLost)).toBe(true);
    expect(api.connectEvents).toHaveBeenCalledWith(onEvent);
    expect(api.onConnectionLost).toBe(onLost);
  });
});
