import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.ts';

export const SESSION_COMMANDS = [
  'createApi',
  'updateSessionDisplay',
];

export const SESSION_COMMAND_METHODS = requiredMethods(SESSION_COMMANDS);

export function registerSessionCommands(target) {
  return registerMethodCommands(target, SESSION_COMMANDS);
}
