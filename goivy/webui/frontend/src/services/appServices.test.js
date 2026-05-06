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
  it('starts the current runtime after Vue has produced its DOM', async () => {
    const runtimeApp = { _unregisterCommands: vi.fn(), _refreshGraphsAndEditorLayout: vi.fn() };
    const startRuntime = vi.fn(() => runtimeApp);
    const services = createAppServices({ startRuntime });

    await expect(services.start()).resolves.toBe(runtimeApp);
    await expect(services.start()).resolves.toBe(runtimeApp);

    expect(startRuntime).toHaveBeenCalledTimes(1);
    services.refreshLayout();
    expect(runtimeApp._refreshGraphsAndEditorLayout).toHaveBeenCalledTimes(1);

    services.stop();
    expect(runtimeApp._unregisterCommands).toHaveBeenCalledTimes(1);
  });

  it('provides an installable singleton for the Vue shell', () => {
    const services = createAppServices({ startRuntime: () => undefined });

    expect(installAppServices(services)).toBe(services);
    expect(currentAppServices()).toBe(services);
  });
});
