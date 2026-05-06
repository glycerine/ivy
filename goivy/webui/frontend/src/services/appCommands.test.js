import { afterEach, describe, expect, it, vi } from 'vitest';
import { registeredCommands, resetCommandRegistry, runCommand } from './commandRegistry.js';
import { APP_COMMAND_METHODS, registerAppCommands } from './appCommands.js';

function fakeCommandTarget() {
  const commandTarget = {
    controls: {
      setStatus: vi.fn(() => 'status-set'),
    },
  };
  for (const method of APP_COMMAND_METHODS) {
    commandTarget[method] = vi.fn((...args) => ({ method, args }));
  }
  return commandTarget;
}

afterEach(() => {
  resetCommandRegistry();
});

describe('app command registration', () => {
  it('registers the production command surface explicitly without constructing the runtime', () => {
    const commandTarget = fakeCommandTarget();
    const unregister = registerAppCommands(commandTarget);

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
    expect(commandTarget.controls.setStatus).toHaveBeenCalledWith('Ready', 'success');

    unregister();
    expect(runCommand('file.save')).toBeUndefined();
  });
});
