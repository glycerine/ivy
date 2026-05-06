import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  IvyRuntime,
  configureIvyRuntimeDependencies,
  resetIvyRuntimeDependencies,
} from './ivyRuntime.js';
import { FakeAPI, FakeControls, FakeGraph } from '../test/fakes.js';

function makeRuntime() {
  resetIvyRuntimeDependencies();
  configureIvyRuntimeDependencies({
    IvyAPI: FakeAPI,
    IvyControls: FakeControls,
    IvyGraph: FakeGraph,
  });
  return new IvyRuntime();
}

afterEach(() => {
  resetIvyRuntimeDependencies();
  delete window.__ivyVueBridge;
  document.body.innerHTML = '';
  vi.useRealTimers();
});

describe('ivyRuntime compatibility behavior', () => {
  it('routes toast notifications through the Vue bridge when available', () => {
    const runtime = makeRuntime();
    window.__ivyVueBridge = {
      showToast: vi.fn(() => 7),
    };

    const result = runtime._showToast('Connection lost', 'error', { persistent: true });

    expect(result).toBe(7);
    expect(window.__ivyVueBridge.showToast).toHaveBeenCalledWith('Connection lost', 'error', { persistent: true });
    expect(document.querySelector('.ivy-toast')).toBeNull();
  });

  it('keeps a direct DOM toast fallback for non-Vue compatibility harnesses', () => {
    const runtime = makeRuntime();

    runtime._showToast('Connection lost', 'error', { persistent: true, className: 'custom-toast' });

    const toast = document.querySelector('.ivy-toast');
    expect(toast.textContent).toBe('Connection lost');
    expect(toast.classList.contains('ivy-toast-floating')).toBe(true);
    expect(toast.classList.contains('ivy-toast-error')).toBe(true);
    expect(toast.classList.contains('custom-toast')).toBe(true);
  });
});
