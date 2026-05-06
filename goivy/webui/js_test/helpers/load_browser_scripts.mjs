import {
  IvyApp,
  configureLegacyAppDependencies,
  resetLegacyAppDependencies,
} from '../../frontend/src/legacyAppController.js';
import { IvyControlsShim } from '../../frontend/src/legacyRuntimeGlobals.js';
import { createIvyPersist } from '../../frontend/src/legacyPersist.js';

export function loadIvyPersist() {
  return createIvyPersist(window);
}

export function loadIvyControls() {
  return IvyControlsShim;
}

export function loadIvyApp(deps = {}) {
  resetLegacyAppDependencies();
  configureLegacyAppDependencies(deps);
  return IvyApp;
}

