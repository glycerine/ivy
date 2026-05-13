import { hasCommand, runCommand as runRegisteredCommand } from './commandRegistry.ts';

export function runCommand(event, callback) {
  event.preventDefault();
  event.stopPropagation();
  if (typeof event.stopImmediatePropagation === 'function') {
    event.stopImmediatePropagation();
  }
  if (typeof callback === 'function') callback();
}

export function runUiCommand(method, ...args) {
  return runRegisteredCommand(method, ...args);
}

export function hasUiCommand(method) {
  return hasCommand(method);
}

export function setAppStatus(message, level = '') {
  return runRegisteredCommand('app.setStatus', message, level) !== undefined;
}

export function runMenuCommand(event, callback) {
  runCommand(event, () => {
    if (hasCommand('flashAndClose')) {
      runRegisteredCommand('flashAndClose', event.currentTarget, callback);
    } else if (typeof callback === 'function') {
      callback();
    }
  });
}

export function toggleDropdownCommand(event, dropdownStore, id, onOpen) {
  runCommand(event, () => {
    const opened = dropdownStore.toggle(id);
    if (opened && typeof onOpen === 'function') onOpen();
  });
}
