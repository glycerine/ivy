import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

export const ANALYSIS_STATE_COMMANDS = [
  'analysisStateLimits',
  'buildAnalysisState',
  'saveAnalysisState',
  'loadAnalysisStateFile',
  'loadAnalysisStateObject',
  'validateAnalysisStateObject',
  'validateAnalysisStateSheet',
  'validateAnalysisStateGraphPayload',
  'validateAnalysisStateEvents',
  'removeAnalysisStateExtraSheets',
];

export const ANALYSIS_STATE_COMMAND_METHODS = requiredMethods(ANALYSIS_STATE_COMMANDS);

export function registerAnalysisStateCommands(target) {
  return registerMethodCommands(target, ANALYSIS_STATE_COMMANDS);
}
