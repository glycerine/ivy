import { describe, expect, it, vi } from 'vitest';
import {
  closeAllDropdowns,
  dispatchMenuDescriptorAction,
  flashAndClose,
  positionDropdownContent,
} from './menuService.ts';

describe('menuService', () => {
  it('closes open dropdowns in the DOM', () => {
    document.body.innerHTML = '<div class="dropdown open"></div>';
    closeAllDropdowns({ doc: document });
    expect(document.querySelector('.dropdown').classList.contains('open')).toBe(false);
  });

  it('clears floating dropdown positioning when menus close', () => {
    document.body.innerHTML = `
      <div class="dropdown open">
        <span class="panel-menu">Menu</span>
        <div class="dropdown-content dropdown-floating" style="left: 11px; top: 12px; min-width: 200px; max-height: 90px;"></div>
      </div>
    `;

    closeAllDropdowns({ doc: document });

    const content = document.querySelector('.dropdown-content');
    expect(content.classList.contains('dropdown-floating')).toBe(false);
    expect(content.getAttribute('style')).toBe('');
  });

  it('positions dropdown content in viewport space', () => {
    document.body.innerHTML = `
      <div class="dropdown open">
        <span class="panel-menu">Menu</span>
        <div class="dropdown-content"></div>
      </div>
    `;
    const dropdown = document.querySelector('.dropdown');
    const trigger = document.querySelector('.panel-menu');
    const content = document.querySelector('.dropdown-content');
    trigger.getBoundingClientRect = () => ({
      x: 740,
      y: 100,
      left: 740,
      top: 100,
      right: 800,
      bottom: 124,
      width: 60,
      height: 24,
      toJSON: () => ({}),
    });
    content.getBoundingClientRect = () => ({
      x: 0,
      y: 0,
      left: 0,
      top: 0,
      right: 200,
      bottom: 120,
      width: 200,
      height: 120,
      toJSON: () => ({}),
    });

    positionDropdownContent(dropdown, { win: { innerWidth: 800, innerHeight: 600 } });

    expect(content.classList.contains('dropdown-floating')).toBe(true);
    expect(content.style.left).toBe('592px');
    expect(content.style.top).toBe('125px');
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

  it('maps CTI concept BMC menu descriptors to the bounded-check dialog', async () => {
    const app = {
      ctiBoundedCheck: vi.fn(async () => ({ ok: true })),
      boundedCheck: vi.fn(),
      runAction: vi.fn(),
    };

    await dispatchMenuDescriptorAction(app, 'concept', { action: 'bmc_conjecture', dispatch: 'action' });

    expect(app.ctiBoundedCheck).toHaveBeenCalled();
    expect(app.boundedCheck).not.toHaveBeenCalled();
    expect(app.runAction).not.toHaveBeenCalled();
  });
});
