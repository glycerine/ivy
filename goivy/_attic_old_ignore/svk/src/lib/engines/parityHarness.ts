import type { CheckResult, GraphSnapshot, IvyEngine, ModelDocument, VerificationJob } from '$lib/types';

export type EngineParityObservation = {
	engine: string;
	loadJob: Pick<VerificationJob, 'kind' | 'status' | 'modelRevision'>;
	graphShape?: Pick<GraphSnapshot, 'kind' | 'nodeOrder' | 'edgeOrder'>;
	checkResult?: Pick<CheckResult, 'result' | 'failedConjecture' | 'failedLabel'>;
};

export type EngineParityDifference = {
	field: string;
	left: unknown;
	right: unknown;
	note?: string;
};

export async function observeEngine(engine: IvyEngine, projectId: string, model: ModelDocument): Promise<EngineParityObservation> {
	const session = await engine.newSession(projectId);
	const loadJob = await engine.loadModel(session.id, model);
	const snapshot = await engine.getSnapshot(session.id, {});
	const graph = snapshot.graphs?.[0];
	return {
		engine: engine.kind,
		loadJob: {
			kind: loadJob.kind,
			status: loadJob.status,
			modelRevision: loadJob.modelRevision
		},
		graphShape: graph
			? {
					kind: graph.kind,
					nodeOrder: graph.nodeOrder,
					edgeOrder: graph.edgeOrder
				}
			: undefined
	};
}

export function compareObservations(
	left: EngineParityObservation,
	right: EngineParityObservation,
	knownDifferences: Record<string, string> = {}
): EngineParityDifference[] {
	const differences: EngineParityDifference[] = [];
	compareField(differences, 'loadJob.kind', left.loadJob.kind, right.loadJob.kind, knownDifferences);
	compareField(differences, 'loadJob.modelRevision', left.loadJob.modelRevision, right.loadJob.modelRevision, knownDifferences);
	compareField(differences, 'graphShape.kind', left.graphShape?.kind, right.graphShape?.kind, knownDifferences);
	compareField(differences, 'graphShape.nodeCount', left.graphShape?.nodeOrder.length, right.graphShape?.nodeOrder.length, knownDifferences);
	compareField(differences, 'graphShape.edgeCount', left.graphShape?.edgeOrder.length, right.graphShape?.edgeOrder.length, knownDifferences);
	return differences;
}

function compareField(
	differences: EngineParityDifference[],
	field: string,
	left: unknown,
	right: unknown,
	knownDifferences: Record<string, string>
) {
	if (Object.is(left, right)) {
		return;
	}
	differences.push({ field, left, right, note: knownDifferences[field] });
}
