import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.ts';

export const GRAPH_COMMANDS = [
  'installConceptGraphVisibilityHook',
  'attachGraphEventHandlers',
  'graphElementsSnapshot',
  'refreshConceptGraph',
  'populateStateCheckboxes',
  'populateConstraintFacts',
  'onEdgeToggle',
  'onEdgeToggleChange',
  'onLabelToggleChange',
  'updateStateLabel',
];

export const GRAPH_COMMAND_METHODS = requiredMethods(GRAPH_COMMANDS);

export function registerGraphCommands(target) {
  return registerMethodCommands(target, GRAPH_COMMANDS);
}
