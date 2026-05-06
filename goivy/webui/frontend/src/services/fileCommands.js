import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

export const FILE_COMMANDS = [
  { command: 'file.load', method: 'chooseAndLoadModelFile' },
  { command: 'file.save', method: 'save' },
  { command: 'file.saveAs', method: 'saveAs' },
  { command: 'file.download', method: 'downloadModel' },
  { command: 'file.close', method: 'closeCurrentFile' },
  { command: 'file.new', method: 'newModel' },
  { command: 'file.reopenLast', method: 'reopenLastFile' },
  'chooseAndLoadModelFile',
  'chooseAndLoadEventTraceFile',
  'chooseAndLoadAnalysisStateFile',
  'loadFile',
  'readFileText',
  'saveInvariant',
  'saveAbstraction',
  'downloadModel',
  'downloadTextFile',
  'save',
  'downloadModelForUnsupportedSave',
  'saveAs',
  'saveSession',
  'closeCurrentFile',
  'newModel',
  'populateRecentFiles',
  'loadRecentSession',
  'reopenLastFile',
  'showDirtyCloseDialog',
  'showExternalChangeDialog',
  'showSaveAsExplanationNotice',
  'hideSaveAsExplanationNotice',
];

export const FILE_COMMAND_METHODS = requiredMethods(FILE_COMMANDS);

export function registerFileCommands(target) {
  return registerMethodCommands(target, FILE_COMMANDS);
}
