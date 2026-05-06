import { afterEach, describe, expect, it, vi } from 'vitest';
import { loadIvyApp } from './helpers/load_browser_scripts.mjs';
import { FakeAPI, FakeControls, FakeGraph } from './helpers/fakes.mjs';

function loadApp() {
  return loadIvyApp({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
}

afterEach(() => {
  delete window.__ivyVueBridge;
  document.body.innerHTML = '';
});

describe('IvyApp toast rendering', () => {
  it('routes toast notifications through the Vue bridge when available', () => {
    const IvyApp = loadApp();
    const app = new IvyApp();
    window.__ivyVueBridge = {
      showToast: vi.fn(() => 7),
    };

    const result = app._showToast('Connection lost', 'error', { persistent: true });

    expect(result).toBe(7);
    expect(window.__ivyVueBridge.showToast).toHaveBeenCalledWith('Connection lost', 'error', { persistent: true });
    expect(document.querySelector('.ivy-toast')).toBeNull();
  });

  it('keeps a direct DOM fallback for no-bridge harnesses', () => {
    const IvyApp = loadApp();
    const app = new IvyApp();

    app._showToast('Connection lost', 'error', { persistent: true });

    const toast = document.querySelector('.ivy-toast');
    expect(toast.textContent).toBe('Connection lost');
    expect(toast.classList.contains('ivy-toast-floating')).toBe(true);
    expect(toast.classList.contains('ivy-toast-error')).toBe(true);
  });
});
