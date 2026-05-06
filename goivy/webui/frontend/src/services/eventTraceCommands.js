import { registerMethodCommands, requiredMethods } from './commandRegistrationHelpers.js';

export const EVENT_TRACE_COMMANDS = [
  'activeEventSheet',
  'addEventPattern',
  'applyEventPatternResult',
  'attachEventTraceHandlers',
  'clearEventPatterns',
  'eventTraceRow',
  'filterEventTrace',
  'findEventTrace',
  'loadEventPatterns',
  'loadEventTraceFile',
  'lookupEventTrace',
  'openEventTraceSheet',
  'removeSelectedEventPattern',
  'renderEventPatternList',
  'renderEventTraceSheet',
  'renderEventTree',
  'renderEventTreeNode',
  'saveEventPatterns',
  'selectEventTraceRow',
  'selectedEventPattern',
  'toggleEventTraceNode',
  'uncoverEventTraceAddress',
];

export const EVENT_TRACE_COMMAND_METHODS = requiredMethods(EVENT_TRACE_COMMANDS);

export function registerEventTraceCommands(target) {
  return registerMethodCommands(target, EVENT_TRACE_COMMANDS);
}
