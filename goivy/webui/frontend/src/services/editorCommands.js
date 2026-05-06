import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

export const EDITOR_COMMANDS = [
  'setEditorContent',
  'scrollEditorToLine',
  'getEditorKeymap',
  'setEditorKeymap',
];

export const EDITOR_COMMAND_METHODS = requiredMethods(EDITOR_COMMANDS);

export function registerEditorCommands(target) {
  return registerMethodCommands(target, EDITOR_COMMANDS);
}
