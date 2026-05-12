import type { CommandRegistration } from './commandRegistry';

export const FILE_COMMANDS: Array<string | CommandRegistration> = [
	{ command: 'file.load', method: 'chooseAndLoadModelFile' },
	{ command: 'file.save', method: 'save' },
	{ command: 'file.saveAs', method: 'saveAs' },
	{ command: 'file.download', method: 'downloadModel' },
	{ command: 'file.close', method: 'closeCurrentFile' },
	{ command: 'file.new', method: 'newModel' },
	{ command: 'file.reopenLast', method: 'reopenLastFile' },
	'chooseAndLoadEventTraceFile',
	'chooseAndLoadAnalysisStateFile',
	'saveAnalysisState',
	'saveInvariant',
	'saveAbstraction'
];

export const CHECK_COMMANDS = [
	'runCheck',
	'checkInduction',
	'boundedCheck',
	'weakenInvariant',
	'showCheckResult',
	'addCheckResultViewActions'
];

export const ANALYSIS_ACTION_COMMANDS = [
	'doUndo',
	'doRedo',
	'resetDomain',
	'diagramDomain',
	'pdrStep',
	'showReachableStates',
	'concreteStep',
	'gatherFacts',
	'ctiConceptAction',
	'minimize',
	'strengthen',
	'reverseStep',
	'pathReach',
	'reachStep',
	'makeConjecture',
	'backtrack',
	'recalculateGraph',
	'rememberGraph',
	'exportConjecture'
];

export const ARG_COMMANDS = [
	'onArgNodeClick',
	'onArgNodeRightClick',
	'onArgEdgeRightClick',
	'executeArgNodeAction',
	'prepareArgNodeActionArgs',
	'executeArgEdgeAction'
];

export const CONCEPT_COMMANDS = [
	'addProjection',
	'addRelationFromString',
	'executeConceptEdgeAction',
	'executeConceptNodeAction',
	'materializeEdge',
	'materializeEdgeFromSelected',
	'materializeNode',
	'onConceptEdgeRightClick',
	'onConceptNodeRightClick',
	'removeConcept',
	'selectConceptNode',
	'splatterNode',
	'splitConcept',
	'supposeEmpty'
];

export const EVENT_TRACE_COMMANDS = [
	'addEventPattern',
	'clearEventPatterns',
	'filterEventTrace',
	'findEventTrace',
	'loadEventPatterns',
	'loadEventTraceFile',
	'openEventTraceSheet',
	'removeSelectedEventPattern',
	'saveEventPatterns',
	'selectEventTraceRow',
	'toggleEventTraceNode'
];

export const DIALOG_COMMANDS = [
	'okDialog',
	'okCancelDialog',
	'textDialog',
	'entryDialog',
	'integerDialog',
	'listboxDialog',
	'buttonListDialog'
];

export const APP_COMMANDS = [
	...FILE_COMMANDS,
	...CHECK_COMMANDS,
	...ANALYSIS_ACTION_COMMANDS,
	...ARG_COMMANDS,
	...CONCEPT_COMMANDS,
	...EVENT_TRACE_COMMANDS,
	...DIALOG_COMMANDS
] as const;

export function commandName(entry: string | CommandRegistration) {
	return typeof entry === 'string' ? entry : entry.command;
}
