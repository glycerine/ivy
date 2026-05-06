import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

export const CHECK_COMMANDS = [
  'runCheck',
  'checkInduction',
  'boundedCheck',
  'weakenInvariant',
  'showCheckResult',
  'addCheckResultViewActions',
];

export const CHECK_COMMAND_METHODS = requiredMethods(CHECK_COMMANDS);

export function registerCheckCommands(target) {
  return registerMethodCommands(target, CHECK_COMMANDS);
}
