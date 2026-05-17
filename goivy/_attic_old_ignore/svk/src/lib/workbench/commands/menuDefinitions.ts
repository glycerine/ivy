export type WorkbenchMenuItem = {
	id: string;
	label: string;
	commandId: string;
	enabled?: boolean;
	separatorBefore?: boolean;
};

export const FILE_MENU: WorkbenchMenuItem[] = [
	{ id: 'file-load', label: 'Load', commandId: 'file.load' },
	{ id: 'file-open-event-trace', label: 'Open Event Trace', commandId: 'chooseAndLoadEventTraceFile' },
	{ id: 'file-save-as', label: 'Save as', commandId: 'file.saveAs' },
	{ id: 'file-download-current-model', label: 'Download current model', commandId: 'file.download' },
	{ id: 'file-save-analysis-state', label: 'Save Analysis State', commandId: 'saveAnalysisState', separatorBefore: true },
	{ id: 'file-load-analysis-state', label: 'Load Analysis State', commandId: 'chooseAndLoadAnalysisStateFile' },
	{ id: 'file-save-invariant', label: 'Save Invariant', commandId: 'saveInvariant', separatorBefore: true },
	{ id: 'file-new-model', label: 'New Model', commandId: 'file.new', separatorBefore: true }
];

export const ARG_INVARIANT_MENU: WorkbenchMenuItem[] = [
	{ id: 'arg-check-induction', label: 'Check induction', commandId: 'checkInduction' },
	{ id: 'arg-bounded-check', label: 'Bounded check', commandId: 'boundedCheck' },
	{ id: 'arg-diagram', label: 'Diagram', commandId: 'diagramDomain' },
	{ id: 'arg-weaken', label: 'Weaken', commandId: 'weakenInvariant' },
	{ id: 'arg-save-invariant', label: 'Save Invariant...', commandId: 'saveInvariant', separatorBefore: true },
	{ id: 'arg-save-abstraction', label: 'Save Abstraction...', commandId: 'saveAbstraction' }
];

export const CONCEPT_CONJECTURE_MENU: WorkbenchMenuItem[] = [
	{ id: 'concept-undo', label: 'Undo', commandId: 'doUndo' },
	{ id: 'concept-redo', label: 'Redo', commandId: 'doRedo' },
	{ id: 'concept-pdr-step', label: 'PDR step', commandId: 'pdrStep', separatorBefore: true },
	{ id: 'concept-concrete', label: 'Concrete', commandId: 'concreteStep' },
	{ id: 'concept-gather', label: 'Gather', commandId: 'gatherFacts' },
	{ id: 'concept-cti-gather', label: 'CTI Gather', commandId: 'ctiConceptAction' },
	{ id: 'concept-minimize', label: 'Minimize', commandId: 'minimize' },
	{ id: 'concept-check-sufficient', label: 'Check sufficient', commandId: 'showCheckResult' },
	{ id: 'concept-check-relative-induction', label: 'Check relative induction', commandId: 'checkInduction' },
	{ id: 'concept-strengthen', label: 'Strengthen', commandId: 'strengthen' },
	{ id: 'concept-reverse', label: 'Reverse', commandId: 'reverseStep' },
	{ id: 'concept-path-reach', label: 'Path reach', commandId: 'pathReach' },
	{ id: 'concept-reach', label: 'Reach', commandId: 'reachStep' },
	{ id: 'concept-conjecture', label: 'Conjecture', commandId: 'makeConjecture' },
	{ id: 'concept-backtrack', label: 'Backtrack', commandId: 'backtrack' },
	{ id: 'concept-recalculate', label: 'Recalculate', commandId: 'recalculateGraph' },
	{ id: 'concept-diagram', label: 'Diagram', commandId: 'diagramDomain' },
	{ id: 'concept-remember', label: 'Remember', commandId: 'rememberGraph', separatorBefore: true },
	{ id: 'concept-export', label: 'Export', commandId: 'exportConjecture' }
];

export const CONCEPT_VIEW_MENU: WorkbenchMenuItem[] = [
	{ id: 'concept-add-relation', label: 'Add relation', commandId: 'addRelationFromString' }
];

export function allStaticMenuItems() {
	return [...FILE_MENU, ...ARG_INVARIANT_MENU, ...CONCEPT_CONJECTURE_MENU, ...CONCEPT_VIEW_MENU];
}
