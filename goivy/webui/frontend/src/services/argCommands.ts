import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.ts';

export const ARG_COMMANDS = [
  'onArgNodeClick',
  'onArgNodeRightClick',
  'onArgEdgeRightClick',
  'executeArgNodeAction',
  'prepareArgNodeActionArgs',
  'executeArgEdgeAction',
];

export const ARG_COMMAND_METHODS = requiredMethods(ARG_COMMANDS);

export function registerArgCommands(target) {
  return registerMethodCommands(target, ARG_COMMANDS);
}
