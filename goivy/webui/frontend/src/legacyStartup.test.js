import { describe, expect, it, vi } from 'vitest';
import { startLegacyAppWhenReady } from './legacyStartup.js';

describe('legacyStartup', () => {
  it('starts the legacy Ivy app after Vue has produced its DOM', async () => {
    const win = { startIvyApp: vi.fn(() => 'started') };

    await expect(startLegacyAppWhenReady(win)).resolves.toBe('started');

    expect(win.startIvyApp).toHaveBeenCalledTimes(1);
  });

  it('does nothing when the legacy app script is unavailable', async () => {
    await expect(startLegacyAppWhenReady({})).resolves.toBeUndefined();
  });
});
