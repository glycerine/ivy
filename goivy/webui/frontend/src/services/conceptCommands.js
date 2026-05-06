import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

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
  'supposeEmpty',
];

export const CONCEPT_COMMAND_METHODS = requiredMethods(CONCEPT_COMMANDS);

export function registerConceptCommands(target) {
  return registerMethodCommands(target, CONCEPT_COMMANDS);
}
