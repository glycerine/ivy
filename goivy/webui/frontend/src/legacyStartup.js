import { nextTick } from 'vue';

export async function startLegacyAppWhenReady(win = globalThis.window) {
  await nextTick();
  if (win && typeof win.startIvyApp === 'function') {
    return win.startIvyApp();
  }
  return undefined;
}
