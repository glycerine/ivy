import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

export const DIALOG_COMMANDS = [
  'okDialog',
  'okCancelDialog',
  'textDialog',
  'showTextDialog',
  'entryDialog',
  'integerDialog',
  'listboxDialog',
  'buttonListDialog',
];

export const DIALOG_COMMAND_METHODS = requiredMethods(DIALOG_COMMANDS);

export function registerDialogCommands(target) {
  return registerMethodCommands(target, DIALOG_COMMANDS);
}
