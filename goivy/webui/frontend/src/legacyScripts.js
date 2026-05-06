import { installLegacyAppRuntime } from './legacyAppRuntime.js';
import { installLegacyRuntimeGlobals } from './legacyRuntimeGlobals.js';

const DEFAULT_LEGACY_SCRIPTS = [];

function scriptAlreadyLoaded(doc, src) {
  return !!(doc && doc.querySelector(`script[data-ivy-legacy-script="${src}"]`));
}

function loadScript(doc, src) {
  if (scriptAlreadyLoaded(doc, src)) {
    return Promise.resolve();
  }
  return new Promise((resolve, reject) => {
    const script = doc.createElement('script');
    script.src = src;
    script.async = false;
    script.dataset.ivyLegacyScript = src;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error(`Failed to load legacy Ivy script: ${src}`));
    doc.body.appendChild(script);
  });
}

export async function loadLegacyIvyRuntime({
  doc = globalThis.document,
  win = globalThis.window,
  scripts = DEFAULT_LEGACY_SCRIPTS,
  installAppRuntime = installLegacyAppRuntime,
} = {}) {
  if (!doc || !win) return;
  win.__IVY_VUE_OWNS_BOOT__ = true;
  installLegacyRuntimeGlobals(win);
  if (typeof win.startIvyApp !== 'function') {
    installAppRuntime({ win });
  }
  if (typeof win.startIvyApp === 'function') return;
  for (const script of scripts) {
    await loadScript(doc, script);
  }
}

export { DEFAULT_LEGACY_SCRIPTS };
