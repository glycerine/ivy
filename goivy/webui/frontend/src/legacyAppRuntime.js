import ivyAppSource from '../../static/js/ivyweb_app.js?raw';

export function installLegacyAppRuntime({
  win = globalThis.window,
  source = ivyAppSource,
} = {}) {
  if (!win || typeof win.startIvyApp === 'function') {
    return win && win.startIvyApp;
  }
  win.__IVY_VUE_OWNS_BOOT__ = true;
  const evaluate = win.eval || globalThis.eval;
  if (typeof evaluate !== 'function') {
    throw new Error('Unable to install IvyApp runtime: global eval is unavailable');
  }
  evaluate.call(win, `${source}\n//# sourceURL=ivyweb_app.compat.js`);
  return win.startIvyApp;
}

