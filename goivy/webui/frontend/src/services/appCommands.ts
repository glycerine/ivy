import { registerCommand } from './commandRegistry.ts';
import { unregisterAll } from './commandRegistrationHelpers.ts';
import { ANALYSIS_ACTION_COMMAND_METHODS, registerAnalysisActionCommands } from './analysisActionCommands.ts';
import { ANALYSIS_STATE_COMMAND_METHODS, registerAnalysisStateCommands } from './analysisStateCommands.ts';
import { ARG_COMMAND_METHODS, registerArgCommands } from './argCommands.ts';
import { CHECK_COMMAND_METHODS, registerCheckCommands } from './checkCommands.ts';
import { CONCEPT_COMMAND_METHODS, registerConceptCommands } from './conceptCommands.ts';
import { DIALOG_COMMAND_METHODS, registerDialogCommands } from './dialogCommands.ts';
import { EDITOR_COMMAND_METHODS, registerEditorCommands } from './editorCommands.ts';
import { EVENT_TRACE_COMMAND_METHODS, registerEventTraceCommands } from './eventTraceCommands.ts';
import { FILE_COMMAND_METHODS, registerFileCommands } from './fileCommands.ts';
import { GRAPH_COMMAND_METHODS, registerGraphCommands } from './graphCommands.ts';
import { MENU_COMMAND_METHODS, registerMenuCommands } from './menuCommands.ts';
import { SESSION_COMMAND_METHODS, registerSessionCommands } from './sessionCommands.ts';
import { SHEET_COMMAND_METHODS, registerSheetCommands } from './sheetCommands.ts';

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
