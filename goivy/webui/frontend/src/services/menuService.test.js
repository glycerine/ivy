import { describe, expect, it, vi } from 'vitest';
import { closeAllDropdowns, dispatchMenuDescriptorAction, flashAndClose } from './menuService.js';

describe('menuService', () => {
  it('closes open dropdowns in the DOM', () => {
    document.body.innerHTML = '<div class="dropdown open"></div>';
    closeAllDropdowns({ doc: document });
    expect(document.querySelector('.dropdown').classList.contains('open')).toBe(false);
  });

  it('flashes menu items before invoking callbacks', () => {
    vi.useFakeTimers();
    const app = { closeAllDropdowns: vi.fn() };
    const el = document.createElement('a');
    el.id = 'file-load';
    const callback = vi.fn();

    flashAndClose(app, el, callback, { win: window });
    vi.advanceTimersByTime(50);

    expect(el.classList.contains('menu-flash')).toBe(false);
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
