import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  IvyApp,
  configureLegacyAppDependencies,
  resetLegacyAppDependencies,
} from './legacyAppController.js';
import { FakeAPI, FakeControls, FakeGraph } from './test/fakes.js';

function makeApp() {
  resetLegacyAppDependencies();
  configureLegacyAppDependencies({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
  return new IvyApp();
}

afterEach(() => {
  resetLegacyAppDependencies();
  delete window.__ivyVueBridge;
  document.body.innerHTML = '';
  vi.useRealTimers();
});

describe('legacyAppController compatibility behavior', () => {
  it('routes toast notifications through the Vue bridge when available', () => {
    const app = makeApp();
    window.__ivyVueBridge = {
      showToast: vi.fn(() => 7),
    };

    const result = app._showToast('Connection lost', 'error', { persistent: true });

    expect(result).toBe(7);
    expect(window.__ivyVueBridge.showToast).toHaveBeenCalledWith('Connection lost', 'error', { persistent: true });
    expect(document.querySelector('.ivy-toast')).toBeNull();
  });

  it('keeps a direct DOM toast fallback for non-Vue compatibility harnesses', () => {
    const app = makeApp();

    app._showToast('Connection lost', 'error', { persistent: true, className: 'custom-toast' });

    const toast = document.querySelector('.ivy-toast');
    expect(toast.textContent).toBe('Connection lost');
    expect(toast.classList.contains('ivy-toast-floating')).toBe(true);
    expect(toast.classList.contains('ivy-toast-error')).toBe(true);
    expect(toast.classList.contains('custom-toast')).toBe(true);
  });
});
