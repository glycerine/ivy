const commands = new Map<string, Function>();

let fallbackTargetGetter: null | (() => any) = null;

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

export function configureCommandRegistry({
  fallbackTarget = undefined,
}: { fallbackTarget?: null | (() => any) } = {}) {
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
