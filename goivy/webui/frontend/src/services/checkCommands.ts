import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.ts';

export const CHECK_COMMANDS = [
  'runCheck',
  'checkInduction',
  'boundedCheck',
  'ctiBoundedCheck',
  'weakenInvariant',
  'showCheckResult',
  'addCheckResultViewActions',
];

export const CHECK_COMMAND_METHODS = requiredMethods(CHECK_COMMANDS);

export function registerCheckCommands(target) {
  return registerMethodCommands(target, CHECK_COMMANDS);
}
