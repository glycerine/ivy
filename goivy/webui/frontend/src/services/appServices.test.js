import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  createAppServices,
  currentAppServices,
  installAppServices,
  resetAppServicesForTests,
} from './appServices.js';

afterEach(() => {
  resetAppServicesForTests();
});

describe('appServices', () => {
  it('starts the current runtime once the static DOM shell is ready', async () => {
    const runtimeApp = { _refreshGraphsAndEditorLayout: vi.fn() };
    const startRuntime = vi.fn(() => runtimeApp);
    const stopRuntime = vi.fn();
    const unregisterCommands = vi.fn();
    const registerCommands = vi.fn(() => unregisterCommands);
    const services = createAppServices({ startRuntime, stopRuntime, registerCommands });

    await expect(services.start()).resolves.toBe(runtimeApp);
    await expect(services.start()).resolves.toBe(runtimeApp);

    expect(startRuntime).toHaveBeenCalledTimes(1);
    expect(registerCommands).toHaveBeenCalledWith(runtimeApp);
    services.refreshLayout();
    expect(runtimeApp._refreshGraphsAndEditorLayout).toHaveBeenCalledTimes(1);

    services.stop();
    expect(unregisterCommands).toHaveBeenCalledTimes(1);
    expect(stopRuntime).toHaveBeenCalledWith(runtimeApp);
  });

  it('provides an installable singleton for the app shell', () => {
    const services = createAppServices({ startRuntime: () => undefined, registerCommands: null });

    expect(installAppServices(services)).toBe(services);
    expect(currentAppServices()).toBe(services);
  });
});
