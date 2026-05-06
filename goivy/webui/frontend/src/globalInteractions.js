import { runUiCommand } from './services/uiCommandService.js';

export function installGlobalInteractions({
  doc = globalThis.document,
  contextMenuStore,
  dropdownStore,
  menuDescriptorStore,
} = {}) {
  if (!doc) return () => {};

  const closeMenus = () => {
    if (contextMenuStore) contextMenuStore.hide();
    if (dropdownStore) dropdownStore.closeAll();
    if (menuDescriptorStore) menuDescriptorStore.closeAll();
  };

  const onClick = (event) => {
    if (contextMenuStore) contextMenuStore.hide();
    if (!event.target || typeof event.target.closest !== 'function' || !event.target.closest('.dropdown')) {
      if (dropdownStore) dropdownStore.closeAll();
      if (menuDescriptorStore) menuDescriptorStore.closeAll();
    }
  };

  const onKeydown = (event) => {
    const key = String(event.key || '').toLowerCase();
    if (event.key === 'Escape') {
      closeMenus();
      return;
    }
    if ((event.ctrlKey || event.metaKey) && key === 's') {
      event.preventDefault();
      runUiCommand('file.save');
      return;
    }
    if ((event.ctrlKey || event.metaKey) && key === 'z') {
      event.preventDefault();
      runUiCommand('doUndo');
    }
  };

  doc.addEventListener('click', onClick);
  doc.addEventListener('keydown', onKeydown);

  return () => {
    doc.removeEventListener('click', onClick);
    doc.removeEventListener('keydown', onKeydown);
  };
}
