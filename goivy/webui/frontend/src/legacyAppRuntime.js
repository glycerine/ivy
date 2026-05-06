import { IvyApp, startIvyApp } from './legacyAppController.js';

export function installLegacyAppRuntime({
  win = globalThis.window,
} = {}) {
  if (!win || typeof win.startIvyApp === 'function') {
    return win && win.startIvyApp;
  }
  win.__IVY_VUE_OWNS_BOOT__ = true;
  win.IvyApp = IvyApp;
  win.startIvyApp = startIvyApp;
  return win.startIvyApp;
}
