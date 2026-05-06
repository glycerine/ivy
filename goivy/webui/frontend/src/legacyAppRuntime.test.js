import { afterEach, describe, expect, it } from 'vitest';
import { installLegacyAppRuntime } from './legacyAppRuntime.js';

afterEach(() => {
  delete window.IvyApp;
  delete window.startIvyApp;
  delete window.__IVY_VUE_OWNS_BOOT__;
});

describe('legacyAppRuntime', () => {
  it('evaluates the bundled controller with Vue as the boot owner', () => {
    const starter = installLegacyAppRuntime({
      win: window,
      source: `
        window.IvyApp = function IvyApp() {};
        window.startIvyApp = function startIvyApp() { return 'started'; };
        if (!window.__IVY_VUE_OWNS_BOOT__) {
          document.addEventListener('DOMContentLoaded', window.startIvyApp);
        }
      `,
    });

    expect(window.__IVY_VUE_OWNS_BOOT__).toBe(true);
    expect(typeof window.IvyApp).toBe('function');
    expect(starter()).toBe('started');
  });

  it('leaves an existing starter alone', () => {
    window.startIvyApp = () => 'existing';

    const starter = installLegacyAppRuntime({
      win: window,
      source: 'throw new Error("should not evaluate");',
    });

    expect(starter()).toBe('existing');
  });
});

