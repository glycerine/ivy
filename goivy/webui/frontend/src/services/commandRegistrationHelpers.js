import { registerCommand } from './commandRegistry.js';

export function commandMethod(spec) {
  return typeof spec === 'string' ? spec : spec.method;
}

export function commandName(spec) {
  return typeof spec === 'string' ? spec : spec.command;
}

export function requiredMethods(specs) {
  return Array.from(new Set(specs.map(commandMethod))).sort();
}

export function registerMethodCommands(target, specs) {
  if (!target) {
    throw new Error('cannot register commands without a target');
  }
  return specs.map((spec) => {
    const method = commandMethod(spec);
    const name = commandName(spec);
    if (typeof target[method] !== 'function') {
      throw new Error(`cannot register command "${name}": missing method "${method}"`);
    }
    return registerCommand(name, target[method].bind(target));
  });
}

export function unregisterAll(unregisters) {
  for (let i = unregisters.length - 1; i >= 0; i -= 1) {
    unregisters[i]();
  }
}
