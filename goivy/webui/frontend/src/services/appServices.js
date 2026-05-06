import { nextTick } from 'vue';
import { startIvyRuntime, stopIvyRuntime } from './ivyRuntime.js';
import { registerAppCommands } from './appCommands.js';

let installedServices = null;

export function createAppServices({
  startRuntime = startIvyRuntime,
  stopRuntime = stopIvyRuntime,
  registerCommands = registerAppCommands,
} = {}) {
  let started = false;
  let runtimeApp = null;
  let unregisterCommands = null;

  return {
    async start() {
      await nextTick();
      if (started) return runtimeApp;
      started = true;
      if (typeof startRuntime === 'function') {
        runtimeApp = startRuntime();
      }
      if (runtimeApp && typeof registerCommands === 'function') {
        unregisterCommands = registerCommands(runtimeApp);
      }
      return runtimeApp;
    },

    stop() {
      if (typeof unregisterCommands === 'function') {
        unregisterCommands();
      }
      unregisterCommands = null;
      if (typeof stopRuntime === 'function') {
        stopRuntime(runtimeApp);
      }
      started = false;
      runtimeApp = null;
    },

    runtime() {
      return runtimeApp;
    },

    refreshLayout() {
      if (runtimeApp && typeof runtimeApp._refreshGraphsAndEditorLayout === 'function') {
        runtimeApp._refreshGraphsAndEditorLayout();
      }
    },
  };
}

export function installAppServices(services = createAppServices()) {
  installedServices = services;
  return installedServices;
}

export function currentAppServices() {
  if (!installedServices) {
    installedServices = createAppServices();
  }
  return installedServices;
}

export function resetAppServicesForTests() {
  installedServices = null;
}
