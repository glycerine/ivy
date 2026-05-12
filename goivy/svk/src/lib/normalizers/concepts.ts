import type { Concept, ConceptState, ConceptToggles } from '$lib/types';
import { asBoolean, asRecord, asString, asStringArray, stableId } from './primitives';

export type ConceptNormalizeOptions = {
	sessionId: string;
	sheetId: string;
	revision?: number;
};

export function normalizeConceptPayload(payload: unknown, options: ConceptNormalizeOptions): ConceptState {
	const data = asRecord(payload);
	const concepts = normalizeConceptMap(data.concepts);
	return {
		id: stableId('concept', options.sessionId, options.sheetId, options.revision ?? 0),
		sessionId: options.sessionId,
		sheetId: options.sheetId,
		concepts,
		sortNodes: asStringArray(data.nodes),
		relationEdges: asStringArray(data.edges),
		nodeLabels: asStringArray(data.node_labels),
		abstractValue: normalizeBooleanRecord(data.abstract_value),
		toggles: normalizeToggles(data.toggles ?? { edges: data.edge_toggles, labels: data.label_toggles }),
		selectedConcept: asString(data.selected_concept) || undefined,
		revision: options.revision ?? 0
	};
}

function normalizeConceptMap(value: unknown): Record<string, Concept> {
	const out: Record<string, Concept> = {};
	for (const [name, raw] of Object.entries(asRecord(value))) {
		const concept = asRecord(raw);
		out[name] = {
			name: asString(concept.name, name),
			variables: asStringArray(concept.variables),
			formula: asString(concept.formula),
			sorts: asStringArray(concept.sorts),
			arity: typeof concept.arity === 'number' ? concept.arity : asStringArray(concept.variables).length
		};
	}
	return out;
}

function normalizeBooleanRecord(value: unknown): Record<string, boolean> {
	const out: Record<string, boolean> = {};
	for (const [key, raw] of Object.entries(asRecord(value))) {
		out[key] = asBoolean(raw);
	}
	return out;
}

function normalizeToggles(value: unknown): ConceptToggles {
	const data = asRecord(value);
	return {
		edges: normalizeNestedBooleans(data.edges),
		labels: normalizeNestedBooleans(data.labels)
	};
}

function normalizeNestedBooleans(value: unknown): Record<string, Record<string, boolean>> {
	const out: Record<string, Record<string, boolean>> = {};
	for (const [name, rawBoxes] of Object.entries(asRecord(value))) {
		out[name] = normalizeBooleanRecord(rawBoxes);
	}
	return out;
}
