import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  connectSessionEvents,
  createLegacyApi,
  createSession,
  updateSessionDisplay,
} from './sessionService.js';

afterEach(() => {
  vi.restoreAllMocks();
});

describe('sessionService', () => {
  it('creates the legacy API through the Vue bridge when available', () => {
    const api = { sessionId: 's1' };
    const bridge = {
      createLegacyApi: vi.fn(() => api),
    };

    expect(createLegacyApi({ bridge })).toBe(api);
    expect(bridge.createLegacyApi).toHaveBeenCalledTimes(1);
  });

  it('falls back to the provided legacy API factory', () => {
    vi.spyOn(console, 'warn').mockImplementation(() => {});
    const api = { sessionId: 'fallback' };
    const bridge = {
      createLegacyApi: vi.fn(() => {
        throw new Error('offline');
      }),
    };

    expect(createLegacyApi({ bridge, fallbackApiFactory: () => api })).toBe(api);
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

  it('updates the session display through the bridge or DOM fallback', () => {
    const bridge = {
      setSessionId: vi.fn(),
    };
    expect(updateSessionDisplay('abc', { bridge })).toBe(true);
    expect(bridge.setSessionId).toHaveBeenCalledWith('abc');

    const el = document.createElement('div');
    el.id = 'session-id';
    document.body.appendChild(el);
    expect(updateSessionDisplay('xyz', { bridge: null, doc: document })).toBe(true);
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
