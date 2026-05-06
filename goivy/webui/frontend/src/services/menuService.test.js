import { describe, expect, it, vi } from 'vitest';
import { closeAllDropdowns, dispatchMenuDescriptorAction, flashAndClose } from './menuService.js';

describe('menuService', () => {
  it('closes dropdowns through Vue or DOM fallback', () => {
    const bridge = { closeDropdownMenus: vi.fn() };
    closeAllDropdowns({ bridge });
    expect(bridge.closeDropdownMenus).toHaveBeenCalled();

    document.body.innerHTML = '<div class="dropdown open"></div>';
    closeAllDropdowns({ bridge: null, doc: document });
    expect(document.querySelector('.dropdown').classList.contains('open')).toBe(false);
  });

  it('flashes Vue-owned menu items before invoking callbacks', () => {
    vi.useFakeTimers();
    const app = { closeAllDropdowns: vi.fn() };
    const el = document.createElement('a');
    el.id = 'file-load';
    const bridge = { flashMenuItem: vi.fn() };
    const callback = vi.fn();

    flashAndClose(app, el, callback, { bridge, win: window });
    vi.advanceTimersByTime(50);

    expect(bridge.flashMenuItem).toHaveBeenCalledWith('file-load', 50);
    expect(app.closeAllDropdowns).toHaveBeenCalled();
    expect(callback).toHaveBeenCalled();
    vi.useRealTimers();
  });

  it('dispatches menu descriptors through the shared action runner', async () => {
    const app = {
      runAction: vi.fn(async () => ({ ok: true })),
    };

    await dispatchMenuDescriptorAction(app, 'concept', { action: 'undo', dispatch: 'action' });

    expect(app.runAction).toHaveBeenCalledWith('undo', {}, {
      runningMessage: 'Running: undo...',
      successMessage: 'Done: undo',
    });
  });
});
