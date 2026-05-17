import { describe, expect, it, vi } from 'vitest';
import { CommandRegistry } from './commandRegistry';
import { APP_COMMANDS, commandName } from './commandGroups';
import { allStaticMenuItems } from './menuDefinitions';

describe('CommandRegistry', () => {
	it('registers, runs, lists, and unregisters commands', () => {
		expect.hasAssertions();
		const registry = new CommandRegistry();
		const handler = vi.fn((value: unknown) => `ok:${value}`);
		const unregister = registry.register('demo.run', handler);

		expect(registry.has('demo.run')).toBe(true);
		expect(registry.run('demo.run', 7)).toBe('ok:7');
		expect(handler).toHaveBeenCalledWith(7);
		expect(registry.list()).toEqual(['demo.run']);

		unregister();
		expect(registry.has('demo.run')).toBe(false);
	});

	it('registers method command groups against a target object', () => {
		expect.hasAssertions();
		const registry = new CommandRegistry();
		const target = {
			save: vi.fn(() => 'saved'),
			saveAs: vi.fn(() => 'saved-as')
		};

		registry.registerMethods(target, [
			{ command: 'file.save', method: 'save' },
			{ command: 'file.saveAs', method: 'saveAs' }
		]);

		expect(registry.run('file.save')).toBe('saved');
		expect(registry.run('file.saveAs')).toBe('saved-as');
	});

	it('covers static menu command ids in the app command surface', () => {
		expect.hasAssertions();
		const commandIds = new Set(APP_COMMANDS.map(commandName));
		const uncovered = allStaticMenuItems()
			.map((item) => item.commandId)
			.filter((commandId) => !commandIds.has(commandId));

		expect(uncovered).toEqual([]);
	});
});
