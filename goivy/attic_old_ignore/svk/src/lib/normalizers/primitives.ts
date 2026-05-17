export function asRecord(value: unknown): Record<string, unknown> {
	return value && typeof value === 'object' && !Array.isArray(value)
		? (value as Record<string, unknown>)
		: {};
}

export function asString(value: unknown, fallback = ''): string {
	return typeof value === 'string' ? value : fallback;
}

export function asNumber(value: unknown, fallback = 0): number {
	return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

export function asBoolean(value: unknown, fallback = false): boolean {
	return typeof value === 'boolean' ? value : fallback;
}

export function asStringArray(value: unknown): string[] {
	return Array.isArray(value) ? value.filter((item): item is string => typeof item === 'string') : [];
}

export function asStringRecord(value: unknown): Record<string, unknown> {
	return asRecord(value);
}

export function stableId(prefix: string, ...parts: unknown[]): string {
	const body = parts
		.map((part) => String(part ?? ''))
		.join('|')
		.replace(/[^A-Za-z0-9_.:-]+/g, '_')
		.replace(/^_+|_+$/g, '');
	return body ? `${prefix}_${body}` : prefix;
}

export function nowIso(now: Date = new Date()): string {
	return now.toISOString();
}
