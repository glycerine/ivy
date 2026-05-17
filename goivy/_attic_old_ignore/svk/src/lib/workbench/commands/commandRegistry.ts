export type CommandHandler = (...args: unknown[]) => unknown;

export type CommandRegistration = {
	command: string;
	method: string;
};

export class CommandRegistry {
	private readonly commands = new Map<string, CommandHandler>();
	private fallbackTarget: unknown = null;

	configure({ fallbackTarget }: { fallbackTarget?: unknown } = {}) {
		this.fallbackTarget = fallbackTarget ?? null;
	}

	reset() {
		this.commands.clear();
		this.fallbackTarget = null;
	}

	register(name: string, handler: CommandHandler) {
		const commandName = assertCommandName(name);
		if (typeof handler !== 'function') {
			throw new TypeError('command handler must be a function');
		}
		this.commands.set(commandName, handler);
		return () => this.unregister(commandName, handler);
	}

	registerMethods(target: Record<string, CommandHandler>, methods: Array<string | CommandRegistration>) {
		return methods.map((entry) => {
			const command = typeof entry === 'string' ? entry : entry.command;
			const method = typeof entry === 'string' ? entry : entry.method;
			const handler = target[method];
			if (typeof handler !== 'function') {
				throw new TypeError(`missing command method: ${method}`);
			}
			return this.register(command, handler.bind(target));
		});
	}

	unregister(name: string, handler?: CommandHandler) {
		const commandName = assertCommandName(name);
		if (handler && this.commands.get(commandName) !== handler) return false;
		return this.commands.delete(commandName);
	}

	list() {
		return [...this.commands.keys()].sort();
	}

	has(name: string) {
		return Boolean(this.handler(name));
	}

	handler(name: string): CommandHandler | undefined {
		const commandName = assertCommandName(name);
		if (this.commands.has(commandName)) return this.commands.get(commandName);
		if (this.fallbackTarget && typeof this.fallbackTarget === 'object') {
			const candidate = (this.fallbackTarget as Record<string, unknown>)[commandName];
			if (typeof candidate === 'function') return candidate.bind(this.fallbackTarget) as CommandHandler;
		}
		return undefined;
	}

	run(name: string, ...args: unknown[]) {
		return this.handler(name)?.(...args);
	}
}

export const commandRegistry = new CommandRegistry();

export function assertCommandName(name: string) {
	if (typeof name !== 'string' || name.trim() === '') {
		throw new TypeError('command name must be a non-empty string');
	}
	return name;
}
