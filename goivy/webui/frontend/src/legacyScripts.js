import { installLegacyAppRuntime } from './legacyAppRuntime.js';
import { installLegacyRuntimeGlobals } from './legacyRuntimeGlobals.js';

const DEFAULT_LEGACY_SCRIPTS = [];

export async function loadLegacyIvyRuntime({
  win = globalThis.window,
  installAppRuntime = installLegacyAppRuntime,
} = {}) {
  if (!win) return;
  win.__IVY_VUE_OWNS_BOOT__ = true;
  installLegacyRuntimeGlobals(win);
  if (typeof win.startIvyApp !== 'function') {
    installAppRuntime({ win });
  }
}

export { DEFAULT_LEGACY_SCRIPTS };
