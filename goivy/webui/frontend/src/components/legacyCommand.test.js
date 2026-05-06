import { afterEach, describe, expect, it, vi } from 'vitest';
import { callApp, hasAppMethod, setAppStatus } from './legacyCommand.js';

afterEach(() => {
  window.ivyApp = undefined;
});

describe('legacyCommand helpers', () => {
  it('routes method calls through the current legacy app', () => {
    window.ivyApp = {
      runCheck: vi.fn(() => 'ok'),
    };

    expect(hasAppMethod('runCheck')).toBe(true);
    expect(hasAppMethod('missing')).toBe(false);
    expect(callApp('runCheck', 'pdr')).toBe('ok');
    expect(window.ivyApp.runCheck).toHaveBeenCalledWith('pdr');
    expect(callApp('missing')).toBeUndefined();
  });

  it('reports status through the legacy controls when present', () => {
    window.ivyApp = {
      controls: {
        setStatus: vi.fn(),
      },
    };

    expect(setAppStatus('Saved', 'ok')).toBe(true);
    expect(window.ivyApp.controls.setStatus).toHaveBeenCalledWith('Saved', 'ok');

    window.ivyApp = {};
    expect(setAppStatus('Ignored', 'error')).toBe(false);
  });
});
