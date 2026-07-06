import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.ts';

export const CONCEPT_COMMANDS = [
  'addProjection',
  'addRelationFromString',
  'executeConceptEdgeAction',
  'executeConceptNodeAction',
  'loadConceptDomain',
  'materializeEdge',
  'materializeEdgeFromSelected',
  'materializeNode',
  'onConceptEdgeRightClick',
  'onConceptNodeRightClick',
  'removeConcept',
  'replaceConceptDomain',
  'saveConceptDomain',
  'selectConceptNode',
  'splatterNode',
  'splitConcept',
  'supposeEmpty',
];

export const CONCEPT_COMMAND_METHODS = requiredMethods(CONCEPT_COMMANDS);

export function registerConceptCommands(target) {
  return registerMethodCommands(target, CONCEPT_COMMANDS);
}
