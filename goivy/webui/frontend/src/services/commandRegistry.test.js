import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  configureCommandRegistry,
  hasCommand,
  registeredCommands,
  registerCommand,
  registerControllerCommands,
  resetCommandRegistry,
  runCommand,
  unregisterCommand,
} from './commandRegistry.js';

afterEach(() => {
  delete window.ivyApp;
  resetCommandRegistry();
});

describe('commandRegistry', () => {
  it('registers, runs, lists, and unregisters commands', () => {
    const handler = vi.fn((name) => `hello ${name}`);
    const remove = registerCommand('demo.greet', handler);

    expect(hasCommand('demo.greet')).toBe(true);
    expect(registeredCommands()).toEqual(['demo.greet']);
    expect(runCommand('demo.greet', 'Ivy')).toBe('hello Ivy');
    expect(handler).toHaveBeenCalledWith('Ivy');

    remove();
    expect(hasCommand('demo.greet')).toBe(false);
    expect(runCommand('demo.greet')).toBeUndefined();
  });

  it('delegates unknown commands to the temporary legacy fallback target', () => {
    const legacy = {
      save: vi.fn(() => 'saved'),
    };
    configureCommandRegistry({ fallbackTarget: () => legacy });

    expect(hasCommand('save')).toBe(true);
    expect(runCommand('save')).toBe('saved');
    expect(legacy.save).toHaveBeenCalledTimes(1);
  });

  it('does not use window.ivyApp as an implicit production fallback', () => {
    window.ivyApp = {
      save: vi.fn(() => 'saved'),
    };

    expect(hasCommand('save')).toBe(false);
    expect(runCommand('save')).toBeUndefined();
    expect(window.ivyApp.save).not.toHaveBeenCalled();
  });

  it('registers controller methods and status updates as commands', () => {
    class Controller {
      constructor() {
        this.controls = {
          setStatus: vi.fn(() => true),
        };
      }

      saveAs(name) {
        return `saved ${name}`;
      }
    }
    const controller = new Controller();

    const remove = registerControllerCommands(controller);

    expect(runCommand('saveAs', 'demo.ivy')).toBe('saved demo.ivy');
    expect(runCommand('app.setStatus', 'Ready', 'success')).toBe(true);
    expect(controller.controls.setStatus).toHaveBeenCalledWith('Ready', 'success');

    remove();
    expect(hasCommand('saveAs')).toBe(false);
  });

  it('can refuse to unregister a command when the handler does not match', () => {
    const handler = vi.fn();
    registerCommand('demo.once', handler);

    expect(unregisterCommand('demo.once', () => {})).toBe(false);
    expect(hasCommand('demo.once')).toBe(true);
    expect(unregisterCommand('demo.once', handler)).toBe(true);
    expect(hasCommand('demo.once')).toBe(false);
  });
});
