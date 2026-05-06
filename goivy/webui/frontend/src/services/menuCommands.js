import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

export const MENU_COMMANDS = [
  'bindMenuAction',
  'closeAllDropdowns',
  'dispatchMenuDescriptorAction',
  'flashAndClose',
  'handleEvent',
  'loadMenuDescriptors',
  'renderMenuDescriptor',
  'renderMenuRegion',
  'setupDetailsResizer',
  'setupDropdownMenus',
  'setupEventHandlers',
  'setupKeyboardShortcuts',
  'setupResizer',
  'setupResizer2',
  'setupResizer3',
  'setupResizerH',
  'setupTabs',
  'setupTutorialUrlBar',
  'toggleTutorial',
];

export const MENU_COMMAND_METHODS = requiredMethods(MENU_COMMANDS);

export function registerMenuCommands(target) {
  return registerMethodCommands(target, MENU_COMMANDS);
}
