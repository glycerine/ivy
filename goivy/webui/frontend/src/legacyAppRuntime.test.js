import { afterEach, describe, expect, it } from 'vitest';
import { installLegacyAppRuntime } from './legacyAppRuntime.js';

afterEach(() => {
  delete window.IvyApp;
  delete window.startIvyApp;
  delete window.ivyApp;
  delete window.__IVY_VUE_OWNS_BOOT__;
});

describe('legacyAppRuntime', () => {
  it('installs the bundled controller module with Vue as the boot owner', () => {
    const starter = installLegacyAppRuntime({ win: window });

    expect(window.__IVY_VUE_OWNS_BOOT__).toBe(true);
    expect(typeof window.IvyApp).toBe('function');
    expect(window.startIvyApp).toBe(starter);
  });

  it('leaves an existing starter alone', () => {
    window.startIvyApp = () => 'existing';

    const starter = installLegacyAppRuntime({ win: window });

    expect(starter()).toBe('existing');
  });
});
