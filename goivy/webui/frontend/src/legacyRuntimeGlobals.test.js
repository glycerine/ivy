import { afterEach, describe, expect, it, vi } from 'vitest';
import { IvyAPIShim, IvyControlsShim, installLegacyRuntimeGlobals } from './legacyRuntimeGlobals.js';

afterEach(() => {
  delete window.IvyGraph;
  delete window.CONCEPT_STYLE;
  delete window.ARG_STYLE;
  delete window.PROOF_STYLE;
  delete window.IvyAPI;
  delete window.IvyControls;
  delete window.IvyPersist;
  delete window.__ivyVueBridge;
});

describe('legacyRuntimeGlobals', () => {
  it('installs frontend-bundled legacy compatibility globals', () => {
    installLegacyRuntimeGlobals(window);

    expect(typeof window.IvyGraph).toBe('function');
    expect(window.ARG_STYLE.length).toBeGreaterThan(0);
    expect(window.CONCEPT_STYLE.length).toBeGreaterThan(0);
    expect(window.IvyAPI).toBe(IvyAPIShim);
    expect(window.IvyControls).toBe(IvyControlsShim);
    expect(window.IvyPersist).toBeTruthy();
    expect(new window.IvyAPI()).toBeInstanceOf(IvyAPIShim);
    expect(new window.IvyControls({})).toBeInstanceOf(IvyControlsShim);
    expect(typeof window.IvyPersist.save).toBe('function');
  });

  it('routes old IvyControls calls into the Vue bridge', () => {
    window.__ivyVueBridge = {
      showContextMenu: vi.fn(),
      hideContextMenu: vi.fn(),
      updateStatus: vi.fn(),
      updateDetails: vi.fn(),
      clearDetails: vi.fn(),
      showLoading: vi.fn(),
      hideLoading: vi.fn(),
    };
    const controls = new IvyControlsShim({});

    controls.showContextMenu(1, 2, [{ name: 'Split' }]);
    controls.setStatus('Ready', 'success');
    controls.showInfo('short', ['long']);
    controls.clearInfo();
    controls.showLoading('Working...');
    controls.hideLoading();
    controls.hideContextMenu();

    expect(window.__ivyVueBridge.showContextMenu).toHaveBeenCalledWith(1, 2, [{ name: 'Split' }]);
    expect(window.__ivyVueBridge.updateStatus).toHaveBeenCalledWith('Ready', 'success');
    expect(window.__ivyVueBridge.updateDetails).toHaveBeenCalledWith({ shortInfo: 'short', longInfo: ['long'] });
    expect(window.__ivyVueBridge.clearDetails).toHaveBeenCalled();
    expect(window.__ivyVueBridge.showLoading).toHaveBeenCalledWith('Working...');
    expect(window.__ivyVueBridge.hideLoading).toHaveBeenCalled();
    expect(window.__ivyVueBridge.hideContextMenu).toHaveBeenCalled();
    expect(controls.isContextMenuVisible()).toBe(false);
  });
});
