import { afterEach, describe, expect, it, vi } from 'vitest';
import { callApp, hasAppMethod, setAppStatus } from './legacyCommand.js';
import { registerCommand, resetCommandRegistry } from '../services/commandRegistry.js';

afterEach(() => {
  window.ivyApp = undefined;
  resetCommandRegistry();
});

describe('legacyCommand helpers', () => {
  it('routes method calls through the command registry first', () => {
    const runCheck = vi.fn(() => 'ok');
    registerCommand('runCheck', runCheck);

    expect(hasAppMethod('runCheck')).toBe(true);
    expect(hasAppMethod('missing')).toBe(false);
    expect(callApp('runCheck', 'pdr')).toBe('ok');
    expect(runCheck).toHaveBeenCalledWith('pdr');
    expect(callApp('missing')).toBeUndefined();
  });

  it('keeps the temporary fallback to the current legacy app', () => {
    window.ivyApp = {
      runCheck: vi.fn(() => 'ok'),
    };

    expect(hasAppMethod('runCheck')).toBe(true);
    expect(hasAppMethod('missing')).toBe(false);
    expect(callApp('runCheck', 'pdr')).toBe('ok');
    expect(window.ivyApp.runCheck).toHaveBeenCalledWith('pdr');
    expect(callApp('missing')).toBeUndefined();
  });

  it('reports status through the command registry', () => {
    const setStatus = vi.fn(() => true);
    registerCommand('app.setStatus', setStatus);

    expect(setAppStatus('Saved', 'ok')).toBe(true);
    expect(setStatus).toHaveBeenCalledWith('Saved', 'ok');

    resetCommandRegistry();
    expect(setAppStatus('Ignored', 'error')).toBe(false);
  });
});
