import { describe, expect, it, vi } from 'vitest';
import { loadLegacyIvyRuntime } from './legacyScripts.js';

function makeFakeDocument() {
  const appended = [];
  const doc = {
    body: {
      appendChild(script) {
        appended.push(script);
        setTimeout(() => script.onload(), 0);
      },
    },
    createElement(tag) {
      expect(tag).toBe('script');
      return {
        async: true,
        dataset: {},
        onload: null,
        onerror: null,
      };
    },
    querySelector: vi.fn(() => null),
  };
  return { doc, appended };
}

describe('legacyScripts', () => {
  it('installs the bundled compatibility runtime and marks Vue as boot owner', async () => {
    const { doc, appended } = makeFakeDocument();
    const win = {};
    const installAppRuntime = vi.fn(({ win: target }) => {
      target.IvyApp = function IvyApp() {};
      target.startIvyApp = () => {};
    });

    await loadLegacyIvyRuntime({ doc, win, installAppRuntime });

    expect(win.__IVY_VUE_OWNS_BOOT__).toBe(true);
    expect(typeof win.IvyAPI).toBe('function');
    expect(typeof win.IvyControls).toBe('function');
    expect(typeof win.IvyPersist.save).toBe('function');
    expect(typeof win.IvyGraph).toBe('function');
    expect(typeof win.startIvyApp).toBe('function');
    expect(installAppRuntime).toHaveBeenCalledWith({ win });
    expect(appended).toHaveLength(0);
  });

  it('skips loading when the legacy starter is already present', async () => {
    const { doc, appended } = makeFakeDocument();
    const win = { startIvyApp: () => {} };

    await loadLegacyIvyRuntime({ doc, win });

    expect(appended).toHaveLength(0);
    expect(win.__IVY_VUE_OWNS_BOOT__).toBe(true);
  });

  it('can still load explicit compatibility scripts after a custom runtime installer declines', async () => {
    const { doc, appended } = makeFakeDocument();
    const win = {};

    await loadLegacyIvyRuntime({
      doc,
      win,
      scripts: ['/static/js/extra-compat.js'],
      installAppRuntime: vi.fn(),
    });

    expect(appended.map((script) => script.src)).toEqual(['/static/js/extra-compat.js']);
    expect(appended.every((script) => script.async === false)).toBe(true);
  });
});
