const ENGINE_COMMAND_ALIASES: Record<string, string> = {
	checkInduction: 'check.induction',
	boundedCheck: 'check.bounded',
	runCheck: 'check.induction',
	showReachableStates: 'concept.action',
	concreteStep: 'check.concrete',
	pdrStep: 'check.pdr',
	doUndo: 'analysis.undo',
	doRedo: 'analysis.redo',
	resetDomain: 'analysis.reset-domain',
	diagramDomain: 'analysis.diagram-domain',
	weakenInvariant: 'analysis.weaken-invariant',
	gatherFacts: 'concept.gather',
	ctiConceptAction: 'concept.cti-gather',
	minimize: 'concept.minimize',
	strengthen: 'concept.strengthen',
	reverseStep: 'concept.reverse',
	pathReach: 'concept.path-reach',
	reachStep: 'concept.reach',
	makeConjecture: 'concept.conjecture',
	backtrack: 'concept.backtrack',
	recalculateGraph: 'concept.recalculate',
	rememberGraph: 'concept.remember',
	exportConjecture: 'concept.export',
	addRelationFromString: 'concept.add-relation'
};

export const COMMAND_SURFACE_IMPLEMENTATION = {
	file: 'goivy/svk/src/lib/workbench/services/fileLifecycle.ts',
	analysisAction: 'goivy/svk/src/lib/workbench/commands/commandRouting.ts',
	arg: 'goivy/svk/src/lib/workbench/commands/commandRouting.ts',
	concept: 'goivy/svk/src/lib/workbench/commands/commandRouting.ts',
	eventTrace: 'goivy/svk/src/lib/workbench/components/panes/EventTraceSheet.svelte'
} as const;

export function routeWorkbenchCommand(commandId: string) {
	return ENGINE_COMMAND_ALIASES[commandId] ?? commandId;
}
