import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.ts';

export const SHEET_COMMANDS = [
  'currentSheet',
  'registerSheet',
  'isVisualOnlySheet',
  'setVisualOnlySheet',
  'visualOnlyMessage',
  'isValidSheetId',
  'assertValidSheetId',
  'sheetTab',
  'sheetExists',
  'switchSheet',
  'addSheet',
  'openARGSheet',
  'removeSheet',
  'tabLabelForSheet',
];

export const SHEET_COMMAND_METHODS = requiredMethods(SHEET_COMMANDS);

export function registerSheetCommands(target) {
  return registerMethodCommands(target, SHEET_COMMANDS);
}
