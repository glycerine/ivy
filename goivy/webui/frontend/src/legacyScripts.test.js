import { describe, expect, it, vi } from 'vitest';
import { DEFAULT_LEGACY_SCRIPTS, loadLegacyIvyRuntime } from './legacyScripts.js';

describe('legacyScripts', () => {
  it('installs the bundled compatibility runtime and marks Vue as boot owner', async () => {
    const win = {};
    const installAppRuntime = vi.fn(({ win: target }) => {
      target.IvyApp = function IvyApp() {};
      target.startIvyApp = () => {};
    });

    await loadLegacyIvyRuntime({ win, installAppRuntime });

    expect(win.__IVY_VUE_OWNS_BOOT__).toBe(true);
    expect(typeof win.IvyAPI).toBe('function');
    expect(typeof win.IvyControls).toBe('function');
    expect(typeof win.IvyPersist.save).toBe('function');
    expect(typeof win.IvyGraph).toBe('function');
    expect(typeof win.startIvyApp).toBe('function');
    expect(installAppRuntime).toHaveBeenCalledWith({ win });
    expect(DEFAULT_LEGACY_SCRIPTS).toEqual([]);
  });

  it('skips loading when the legacy starter is already present', async () => {
    const win = { startIvyApp: () => {} };

    await loadLegacyIvyRuntime({ win });

    expect(win.__IVY_VUE_OWNS_BOOT__).toBe(true);
  });
});
