import { registerCommand } from './commandRegistry.js';
import { unregisterAll } from './commandRegistrationHelpers.js';
import { ANALYSIS_ACTION_COMMAND_METHODS, registerAnalysisActionCommands } from './analysisActionCommands.js';
import { ANALYSIS_STATE_COMMAND_METHODS, registerAnalysisStateCommands } from './analysisStateCommands.js';
import { ARG_COMMAND_METHODS, registerArgCommands } from './argCommands.js';
import { CHECK_COMMAND_METHODS, registerCheckCommands } from './checkCommands.js';
import { CONCEPT_COMMAND_METHODS, registerConceptCommands } from './conceptCommands.js';
import { DIALOG_COMMAND_METHODS, registerDialogCommands } from './dialogCommands.js';
import { EDITOR_COMMAND_METHODS, registerEditorCommands } from './editorCommands.js';
import { EVENT_TRACE_COMMAND_METHODS, registerEventTraceCommands } from './eventTraceCommands.js';
import { FILE_COMMAND_METHODS, registerFileCommands } from './fileCommands.js';
import { GRAPH_COMMAND_METHODS, registerGraphCommands } from './graphCommands.js';
import { MENU_COMMAND_METHODS, registerMenuCommands } from './menuCommands.js';
import { SESSION_COMMAND_METHODS, registerSessionCommands } from './sessionCommands.js';
import { SHEET_COMMAND_METHODS, registerSheetCommands } from './sheetCommands.js';

const registrars = [
  registerFileCommands,
  registerSessionCommands,
  registerEditorCommands,
  registerSheetCommands,
  registerGraphCommands,
  registerEventTraceCommands,
  registerArgCommands,
  registerConceptCommands,
  registerAnalysisActionCommands,
  registerCheckCommands,
  registerAnalysisStateCommands,
  registerDialogCommands,
  registerMenuCommands,
];

export const APP_COMMAND_METHODS = Array.from(new Set([
  ...FILE_COMMAND_METHODS,
  ...SESSION_COMMAND_METHODS,
  ...EDITOR_COMMAND_METHODS,
  ...SHEET_COMMAND_METHODS,
  ...GRAPH_COMMAND_METHODS,
  ...EVENT_TRACE_COMMAND_METHODS,
  ...ARG_COMMAND_METHODS,
  ...CONCEPT_COMMAND_METHODS,
  ...ANALYSIS_ACTION_COMMAND_METHODS,
  ...CHECK_COMMAND_METHODS,
  ...ANALYSIS_STATE_COMMAND_METHODS,
  ...DIALOG_COMMAND_METHODS,
  ...MENU_COMMAND_METHODS,
])).sort();

export function registerAppCommands(commandTarget) {
  const unregisters = registrars.flatMap((register) => register(commandTarget));
  if (commandTarget && commandTarget.controls && typeof commandTarget.controls.setStatus === 'function') {
    unregisters.push(registerCommand('app.setStatus', commandTarget.controls.setStatus.bind(commandTarget.controls)));
  }
  return () => unregisterAll(unregisters);
}
