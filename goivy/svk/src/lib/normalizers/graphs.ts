import type {
	CytoscapeElement,
	CytoscapePayload,
	GraphEdge,
	GraphKind,
	GraphNode,
	GraphSnapshot
} from '$lib/types';
import { asNumber, asRecord, asString, nowIso, stableId } from './primitives';

export type GraphNormalizeOptions = {
	id: string;
	sheetId: string;
	kind: GraphKind;
	sourceRevision?: number;
	styleRevision?: number;
	now?: Date;
};

export function normalizeGraphPayload(
	payload: CytoscapePayload | null | undefined,
	options: GraphNormalizeOptions
): GraphSnapshot {
	const nodes: Record<string, GraphNode> = {};
	const edges: Record<string, GraphEdge> = {};
	const nodeOrder: string[] = [];
	const edgeOrder: string[] = [];
	const layout: Record<string, { x: number; y: number }> = {};
	const cyIdToNodeId: Record<string, string> = {};
	const elements = Array.isArray(payload?.elements) ? payload.elements : [];

	for (const element of elements) {
		if (element.group !== 'nodes') {
			continue;
		}
		const node = normalizeNode(element);
		nodes[node.id] = node;
		nodeOrder.push(node.id);
		const cyId = asString(asRecord(element.data).id);
		if (cyId) {
			cyIdToNodeId[cyId] = node.id;
		}
		if (element.position) {
			layout[node.id] = {
				x: asNumber(element.position.x),
				y: asNumber(element.position.y)
			};
		}
	}

	for (const element of elements) {
		if (element.group !== 'edges') {
			continue;
		}
		const edge = normalizeEdge(element, cyIdToNodeId);
		edges[edge.id] = edge;
		edgeOrder.push(edge.id);
	}

	return {
		id: options.id,
		sheetId: options.sheetId,
		kind: options.kind,
		sourceRevision: options.sourceRevision ?? 0,
		nodes,
		edges,
		nodeOrder,
		edgeOrder,
		layout: Object.keys(layout).length > 0 ? layout : undefined,
		styleRevision: options.styleRevision ?? 0,
		createdAt: nowIso(options.now)
	};
}

function normalizeNode(element: CytoscapeElement): GraphNode {
	const data = asRecord(element.data);
	const obj = asString(data.obj, asString(data.id));
	const id = stableId('node', obj || data.id);
	const actions = Array.isArray(data.actions)
		? data.actions.map(asRecord).map((action) => ({
				label: asString(action.label),
				action: asString(action.action),
				args: asRecord(action.args)
			}))
		: undefined;
	return {
		id,
		obj,
		label: asString(data.label, obj),
		classes: splitClasses(element.classes),
		shape: asString(data.shape) || undefined,
		shortInfo: asString(data.short_info) || undefined,
		longInfo: data.long_info,
		actions: actions && actions.length > 0 ? actions : undefined,
		width: data.width === undefined ? undefined : asNumber(data.width),
		height: data.height === undefined ? undefined : asNumber(data.height),
		color: asString(data.border_color) || undefined
	};
}

function normalizeEdge(element: CytoscapeElement, cyIdToNodeId: Record<string, string>): GraphEdge {
	const data = asRecord(element.data);
	const obj = asString(data.obj, asString(data.id));
	const sourceObj = asString(data.source_obj, asString(data.source));
	const targetObj = asString(data.target_obj, asString(data.target));
	return {
		id: stableId('edge', obj, sourceObj, targetObj),
		obj,
		source: cyIdToNodeId[asString(data.source)] ?? stableId('node', sourceObj),
		target: cyIdToNodeId[asString(data.target)] ?? stableId('node', targetObj),
		label: asString(data.label) || undefined,
		classes: splitClasses(element.classes),
		shortInfo: asString(data.short_info) || undefined,
		longInfo: data.long_info
	};
}

function splitClasses(classes: string | undefined): string[] {
	return classes ? classes.split(/\s+/).filter(Boolean) : [];
}
