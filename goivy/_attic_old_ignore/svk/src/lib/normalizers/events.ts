import type { TraceEvent, TraceEventSheet } from '$lib/types';
import { asRecord, asString, stableId } from './primitives';

export type TraceSheetNormalizeOptions = {
	id: string;
	projectId: string;
	sessionId: string;
	label?: string;
	revision?: number;
	patterns?: string[];
};

export function normalizeTraceEventSheet(
	events: unknown,
	options: TraceSheetNormalizeOptions
): TraceEventSheet {
	const rawEvents = Array.isArray(events) ? events : [];
	return {
		id: options.id,
		projectId: options.projectId,
		sessionId: options.sessionId,
		label: options.label ?? options.id,
		events: rawEvents.map((event, index) => normalizeTraceEvent(event, String(index))),
		patterns: options.patterns ?? [],
		expandedAddresses: [],
		revision: options.revision ?? 0
	};
}

function normalizeTraceEvent(value: unknown, fallbackAddress: string): TraceEvent {
	const data = asRecord(value);
	const address = asString(data.address, fallbackAddress);
	const children = Array.isArray(data.subs)
		? data.subs.map((child, index) => normalizeTraceEvent(child, `${address}/${index}`))
		: undefined;
	return {
		id: stableId('event', address),
		text: asString(data.text),
		address,
		children: children && children.length > 0 ? children : undefined
	};
}
