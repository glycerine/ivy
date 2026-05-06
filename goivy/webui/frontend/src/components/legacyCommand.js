export function runCommand(event, callback) {
  event.preventDefault();
  event.stopPropagation();
  if (typeof event.stopImmediatePropagation === 'function') {
    event.stopImmediatePropagation();
  }
  if (typeof callback === 'function') callback();
}

export function ivyApp() {
  return globalThis.window && globalThis.window.ivyApp;
}

export function callApp(method, ...args) {
  const app = ivyApp();
  if (app && typeof app[method] === 'function') {
    return app[method](...args);
  }
  return undefined;
}

export function hasAppMethod(method) {
  const app = ivyApp();
  return !!(app && typeof app[method] === 'function');
}

export function setAppStatus(message, level = '') {
  const app = ivyApp();
  if (app && app.controls && typeof app.controls.setStatus === 'function') {
    app.controls.setStatus(message, level);
    return true;
  }
  return false;
}

export function runMenuCommand(event, callback) {
  runCommand(event, () => {
    const app = ivyApp();
    if (app && typeof app.flashAndClose === 'function') {
      app.flashAndClose(event.currentTarget, callback);
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
