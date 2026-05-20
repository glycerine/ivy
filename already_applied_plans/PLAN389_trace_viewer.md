# Safety Error Trace Viewer For ARG Node Safety

## Summary
Port Python’s working bounded-safety/BMC error-trace behavior to the Go/web UI. Python’s real path is: node "Check safety" in bounded/induction mode runs BMC, receives a counterexample 'AnalysisGraph', asks "View error trace?", then opens that ARG in a new tab. Python’s local "concrete trace" hook is stubbed, so v1 will not invent new local-safety trace behavior.

## Key Changes
- Add a Go/web node-safety result path that preserves the existing 'CheckSafetyNode()' API but adds an internal result form carrying 'safe', 'message', updated ARG, and optional counterexample trace ARG.
- For 'ArgNodeAction("check_safety")', read 'args.mode'; if it is 'bounded' or 'induction', run bounded safety/BMC, matching Python. The frontend will pass 'mode: app.getMode()' for node safety actions.
- When bounded node safety fails and 'goivy.SafetyResult.Art' is present, register it as a new analysis sheet and return:
  - 'result: "fail"'
  - 'safe: false'
  - 'message: "The node is unsafe: View error trace?"'
  - 'trace_arg'
  - 'trace_sheet_id'
  - 'trace_label'
  - 'arg' for the refreshed original ARG
- Mark the final node in the trace sheet with the existing 'marked_state' styling so the failing endpoint is visually obvious without adding a new graph style contract.
- Refactor the existing top-level check trace button helper into reusable frontend trace-view logic. Use it from node safety so the info modal shows a visible 'View error trace' button that opens 'trace_arg' via 'openARGSheet'.
- Preserve conformance mode safety by stripping Go-only trace fields from 'GoBackend.ArgAction' when 'WebUIConformCheck' is enabled, mirroring the existing top-level 'Check' behavior.

## Public Interface Additions
- Node action responses may now include optional trace fields: 'trace_arg', 'trace_sheet_id', 'trace_label', and 'result'.
- Frontend ARG node action requests for 'check_safety' will include 'mode'.
- No Python changes. No local concrete-trace behavior in v1.

## Test Plan
- Go backend test: create a real unsafe bounded-safety node by adding a false assertion to the ARG, run 'ArgNodeAction("state_0", "check_safety", {"mode":"bounded"})', assert 'safe=false', 'trace_arg' exists, 'trace_sheet_id' exists, and the trace ARG has at least one state with the final state marked.
- Go compatibility test: existing safe-node safety tests still pass and still return the updated original 'arg'.
- Frontend unit test: 'executeArgNodeAction' sends 'mode', shows the failed-safety message, creates the 'View error trace' button, and clicking it calls 'openARGSheet' with the returned trace payload.
- Playwright test: right-click an ARG node, choose 'Check safety', click 'View error trace', and verify a new ARG sheet opens with the counterexample nodes/edge visible.
- Full verification: run 'cd /Users/jaten/ivy/goivy && source /Users/jaten/ivy/pyivy/goivy-venv/bin/activate && make test-web'.

## Assumptions
- V1 scope is bounded/induction node safety only, per the selected option.
- Local safety failures may still show unsafe-state messaging, but they will not show a trace button unless future work ports behavior beyond Python’s stub.
- The trace sheet is an ARG sheet, not an event-trace sheet, because that is what Python’s 'view_ag(res)' opens for bounded safety.

