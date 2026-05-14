import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.ts';

export const ANALYSIS_ACTION_COMMANDS = [
  'runAction',
  'doUndo',
  'doRedo',
  'resetDomain',
  'diagramDomain',
  'pdrStep',
  'showReachableStates',
  'concreteStep',
  'gatherFacts',
  'ctiConceptAction',
  'reverseStep',
  'pathReach',
  'reachStep',
  'makeConjecture',
  'backtrack',
  'recalculateGraph',
  'relayoutConceptGraph',
  'rememberGraph',
  'exportConjecture',
  'refreshAfterLoad',
  'getMode',
  'setMode',
];

export const ANALYSIS_ACTION_COMMAND_METHODS = requiredMethods(ANALYSIS_ACTION_COMMANDS);

export function registerAnalysisActionCommands(target) {
  return registerMethodCommands(target, ANALYSIS_ACTION_COMMANDS);
}
