import { nextTick } from 'vue';
import { startIvyApp } from '../legacyAppController.js';

let installedServices = null;

export function createAppServices({
  startRuntime = startIvyApp,
} = {}) {
  let started = false;
  let runtimeApp = null;

  return {
    async start() {
      await nextTick();
      if (started) return runtimeApp;
      started = true;
      if (typeof startRuntime === 'function') {
        runtimeApp = startRuntime();
      }
      return runtimeApp;
    },

    stop() {
      if (runtimeApp && typeof runtimeApp._unregisterCommands === 'function') {
        runtimeApp._unregisterCommands();
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
