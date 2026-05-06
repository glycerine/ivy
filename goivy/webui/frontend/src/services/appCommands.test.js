import { afterEach, describe, expect, it, vi } from 'vitest';
import { registeredCommands, resetCommandRegistry, runCommand } from './commandRegistry.js';
import { LEGACY_CONTROLLER_COMMAND_METHODS, registerLegacyControllerCommands } from './appCommands.js';

function fakeController() {
  const controller = {
    controls: {
      setStatus: vi.fn(() => 'status-set'),
    },
  };
  for (const method of LEGACY_CONTROLLER_COMMAND_METHODS) {
    controller[method] = vi.fn((...args) => ({ method, args }));
  }
  return controller;
}

afterEach(() => {
  resetCommandRegistry();
});

describe('app command registration', () => {
  it('registers the production command surface explicitly without constructing IvyApp', () => {
    const controller = fakeController();
    const unregister = registerLegacyControllerCommands(controller);

    expect(registeredCommands()).toContain('file.save');
    expect(registeredCommands()).toContain('checkInduction');
    expect(registeredCommands()).toContain('executeConceptNodeAction');
    expect(registeredCommands()).toContain('saveAnalysisState');
    expect(registeredCommands()).toContain('app.setStatus');

    expect(runCommand('file.save')).toEqual({ method: 'save', args: [] });
    expect(runCommand('executeArgNodeAction', { obj: 'state_0' }, 'join')).toEqual({
      method: 'executeArgNodeAction',
      args: [{ obj: 'state_0' }, 'join'],
    });
    expect(runCommand('app.setStatus', 'Ready', 'success')).toBe('status-set');
    expect(controller.controls.setStatus).toHaveBeenCalledWith('Ready', 'success');

    unregister();
    expect(runCommand('file.save')).toBeUndefined();
  });
});
