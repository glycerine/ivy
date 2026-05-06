const commands = new Map();

let fallbackTargetGetter = null;

function assertCommandName(name) {
  if (typeof name !== 'string' || name.trim() === '') {
    throw new TypeError('command name must be a non-empty string');
  }
  return name;
}

function assertCommandFunction(fn) {
  if (typeof fn !== 'function') {
    throw new TypeError('command handler must be a function');
  }
  return fn;
}

export function configureCommandRegistry({ fallbackTarget } = {}) {
  if (typeof fallbackTarget === 'function') {
    fallbackTargetGetter = fallbackTarget;
  } else if (fallbackTarget === null) {
    fallbackTargetGetter = null;
  }
}

export function resetCommandRegistry() {
  commands.clear();
  fallbackTargetGetter = null;
}

export function registerCommand(name, fn) {
  const commandName = assertCommandName(name);
  commands.set(commandName, assertCommandFunction(fn));
  return () => unregisterCommand(commandName, fn);
}

export function unregisterCommand(name, fn) {
  const commandName = assertCommandName(name);
  if (fn && commands.get(commandName) !== fn) {
    return false;
  }
  return commands.delete(commandName);
}

export function registeredCommands() {
  return Array.from(commands.keys()).sort();
}

export function commandHandler(name) {
  const commandName = assertCommandName(name);
  if (commands.has(commandName)) {
    return commands.get(commandName);
  }
  const target = fallbackTargetGetter && fallbackTargetGetter();
  if (target && typeof target[commandName] === 'function') {
    return target[commandName].bind(target);
  }
  return undefined;
}

export function hasCommand(name) {
  return !!commandHandler(name);
}

export function runCommand(name, ...args) {
  const handler = commandHandler(name);
  if (!handler) {
    return undefined;
  }
  return handler(...args);
}

export function registerControllerCommands(controller) {
  if (!controller) {
    return () => {};
  }

  const names = new Set();
  let proto = Object.getPrototypeOf(controller);
  while (proto && proto !== Object.prototype) {
    for (const name of Object.getOwnPropertyNames(proto)) {
      if (name === 'constructor') continue;
      if (typeof controller[name] === 'function') {
        names.add(name);
      }
    }
    proto = Object.getPrototypeOf(proto);
  }

  for (const name of Object.keys(controller)) {
    if (typeof controller[name] === 'function') {
      names.add(name);
    }
  }

  const unregister = [];
  for (const name of names) {
    unregister.push(registerCommand(name, controller[name].bind(controller)));
  }
  if (controller.controls && typeof controller.controls.setStatus === 'function') {
    unregister.push(registerCommand('app.setStatus', controller.controls.setStatus.bind(controller.controls)));
  }

  return () => {
    for (const remove of unregister) remove();
  };
}
