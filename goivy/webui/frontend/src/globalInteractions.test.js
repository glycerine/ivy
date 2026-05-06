import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { installGlobalInteractions } from './globalInteractions.js';
import { useContextMenuStore, useDropdownStore, useMenuDescriptorStore } from './stores/index.js';
import { registerCommand, resetCommandRegistry } from './services/commandRegistry.js';

describe('globalInteractions', () => {
  let saveCommand;

  beforeEach(() => {
    setActivePinia(createPinia());
    window.ivyApp = {
      doUndo: vi.fn(),
    };
    saveCommand = vi.fn();
    registerCommand('file.save', saveCommand);
  });

  afterEach(() => {
    document.body.innerHTML = '';
    window.ivyApp = undefined;
    resetCommandRegistry();
  });

  it('closes Vue-owned menus on document click and Escape', () => {
    document.body.innerHTML = '<button id="outside"></button>';
    const contextMenuStore = useContextMenuStore();
    const dropdownStore = useDropdownStore();
    const menuDescriptorStore = useMenuDescriptorStore();
    contextMenuStore.show(1, 2, [{ name: 'Inspect' }]);
    dropdownStore.toggle('file-menu');
    menuDescriptorStore.setRegion('concept', [{ label: 'Action', items: [] }]);
    menuDescriptorStore.toggle('concept', 0);

    const cleanup = installGlobalInteractions({ contextMenuStore, dropdownStore, menuDescriptorStore });

    document.getElementById('outside').dispatchEvent(new MouseEvent('click', { bubbles: true }));

    expect(contextMenuStore.visible).toBe(false);
    expect(dropdownStore.isOpen('file-menu')).toBe(false);
    expect(menuDescriptorStore.isOpen('concept', 0)).toBe(false);

    contextMenuStore.show(1, 2, [{ name: 'Inspect' }]);
    dropdownStore.toggle('file-menu');
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));

    expect(contextMenuStore.visible).toBe(false);
    expect(dropdownStore.isOpen('file-menu')).toBe(false);
    cleanup();
  });

  it('routes save and undo shortcuts to the legacy service layer', () => {
    const cleanup = installGlobalInteractions({});
    const saveEvent = new KeyboardEvent('keydown', {
      key: 's',
      ctrlKey: true,
      bubbles: true,
      cancelable: true,
    });
    const undoEvent = new KeyboardEvent('keydown', {
      key: 'z',
      metaKey: true,
      bubbles: true,
      cancelable: true,
    });

    document.dispatchEvent(saveEvent);
    document.dispatchEvent(undoEvent);

    expect(saveEvent.defaultPrevented).toBe(true);
    expect(undoEvent.defaultPrevented).toBe(true);
    expect(saveCommand).toHaveBeenCalled();
    expect(window.ivyApp.doUndo).toHaveBeenCalled();
    cleanup();
  });
});
