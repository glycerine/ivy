# Plan: Fix TypeScript Data Model + Comprehensive Tests

## Context

A code review of `goivy/webui/frontend/src/models/uiDataModel.ts` (the TypeScript
port of the Python/Go three-layer data model) found 10 concrete issues: two bugs that
silently discard data, one structural mismatch (TagValue array vs bool map), three dead
fields that are never populated by the wire format, and several missing/incomplete
fields.  The Go backend produces two payloads — the ARG payload (from
`WebUIARGPayload`) and the Concept payload (from `GetConcept`) — whose exact JSON
shapes are now fully documented.  This plan fixes every issue, then backs each fix
with targeted unit tests, and adds Go-side payload conformance tests.

---

## Confirmed Wire Formats

### ARG payload (`GET /api/session/{id}/arg`)
```json
{
  "analysis_graph_state": {
    "states":      [{"id":int,"label":str,"is_bottom":bool,"info":str}],
    "transitions": [{"source_id":int,"target_id":int,"label":str,"is_join":bool}],
    "covering":    [{"covered_id":int,"covering_id":int}]
  },
  "elements":  [CyElement],
  "positions": null
}
```
Key fact: **no `analysis_graph` key exists**.

### Concept payload (`GET /api/session/{id}/concept`)
Top-level keys (all flat):
- `abstract_value` — `Record<string,bool>` (node_label-prefixed keys only)
- `concept_domain` — `ConceptDomain` wire struct (concepts/combiners/nodes/edges/node_labels)
- `concept_session` — `{domain, abstract_value}` only — **no undo depth**
- `concept_interactive_session` — `{abstract_value: TagValue[], domain: CDConceptDomain,
  goal_constraints: string[], redo_depth, state: string, suppose_constraints: string[],
  undo_depth, info}` — **`abstract_value` is an array of TagValue objects, not a map**
- `display_checkboxes` — `{edges:{…},labels:{…}}` (same as `toggles`)
- `toggles` — duplicate of `display_checkboxes`
- `graph` — `conceptGraphPayload` result (concept_session, display_checkboxes, graph_stack, sorts, …)
- `graph_stack` — `{can_undo,can_redo,undo_depth,redo_depth}` (4 fields only)
- `elements` — `CyElement[]`
- `facts`, `nodes`, `edges`, `node_labels`, `relations`, `edge_sorts`, `label_sorts`, …

Key facts: `conceptInteractiveSessionPayload()` does **not** include `axioms` or `cache`.
`ConceptSession` wire is `{domain, abstract_value}` — no undo depth fields.
`GraphStack` wire is 4 scalar fields — no `current`/`undoStack`/`redoStack`.
`CyElements.NodeID`/`EdgeID` are tagged `json:"-"` — never on the wire.

---

## Files to Modify

1. `goivy/webui/frontend/src/models/uiDataModel.ts` — 8 targeted fixes
2. `goivy/webui/frontend/src/models/uiDataModel.test.ts` — ~18 new test cases
3. `goivy/webui/webui_session.go` — add `axioms` and `cache` to CIS payload (lines ~583-603)
4. New `goivy/webui/webui_payload_test.go` — Go conformance tests for wire format

---

## Change A — TypeScript `uiDataModel.ts`

### A1. Remove `CyElements.nodeId` and `CyElements.edgeId`
`art_cyrender.go:27-28` tags both `json:"-"`.  They are never on the wire.
Remove both fields from `CyElements` class and constructor.
Callers that need element lookups should use `element.data['id']` and
`element.data['obj']` directly from the elements array.

### A2. Add `State.expr` and `State.universe`
Python `State` has `expr` (how the state was reached) and `universe` (BMC model
universe).  Both are needed for CTI visualization.  Add:
```ts
readonly expr: unknown;      // pick(raw, 'expr', 'Expr')
readonly universe: unknown;  // pick(raw, 'universe', 'Universe')
```
Type `unknown` is correct — these are formula objects on the Go side.

### A3. Remove `ARGSnapshot.analysisGraph`
No `analysis_graph` key exists in the Go ARG payload.  This field is always an empty
`AnalysisGraph`.  Remove the field from `ARGSnapshot`.  The `AnalysisGraph`,
`AnalysisTransition`, `State` classes remain in the file for analysis-state
save/restore and future use.

### A4. Remove `ConceptSession.undoDepth` and `ConceptSession.redoDepth`
Go's `ConceptSession` serializes as `{domain, abstract_value}` only (the `undoStack`
field is unexported/lowercase).  Move `undoDepth`/`redoDepth` exclusively to
`ConceptInteractiveSession` which does receive `undo_depth`/`redo_depth` from
`conceptInteractiveSessionPayload()`.

### A5. Remove `GraphStack.current`, `.undoStack`, `.redoStack`
`conceptGraphStackPayload()` returns only `{can_undo, can_redo, undo_depth, redo_depth}`.
Remove `current`, `undoStack`, `redoStack` from `GraphStack`.  The class retains the 4
scalar fields which are populated correctly.

### A6. Fix `ConceptInteractiveSession.abstractValue` — TagValue array→map conversion
Go sends:
```json
"abstract_value": [{"Tag": ["node_info","at_least_one","client"], "Value": true}, …]
```
The current `boolMap()` call returns `{}` for arrays (silent empty map).
Add a private helper `tagValueArrayToMap()` and override in the CIS constructor:
```ts
function tagValueArrayToMap(arr: unknown[]): Record<string, boolean> {
  const out: Record<string, boolean> = {};
  for (const item of arr) {
    if (!isRecord(item)) continue;
    const tag = item['Tag'] ?? item['tag'];
    const val = item['Value'] ?? item['value'];
    if (Array.isArray(tag) && typeof val === 'boolean')
      out[(tag as string[]).join('|')] = val;
  }
  return out;
}
```
In `ConceptInteractiveSession` constructor after `super(raw)`, re-set `abstractValue`:
```ts
const avRaw = pick(raw, 'abstract_value', 'abstractValue', 'AbstractValue');
if (Array.isArray(avRaw)) {
  (this as any).abstractValue = tagValueArrayToMap(avRaw);
}
```
Also inherit `undoDepth`/`redoDepth` (moved from A4) here with explicit parsing.

### A7. Document `ConceptInteractiveSession.domain` mismatch
Go's CIS payload sends `CDConceptDomain` (PascalCase keys, CDConceptDict entries) for
`domain`.  TypeScript's `new ConceptDomain(pick(raw, 'domain'))` silently returns empty.
This is acceptable because browsers read the concept domain from the top-level
`snapshot.domain` field (parsed from the `concept_domain` key), not from
`interactiveSession.domain`.  Add a comment to the `ConceptInteractiveSession`
constructor noting this and that `domain` inherited from `ConceptSession` will be empty.

### A8. Add missing `axioms` and `cache` fields for `ConceptInteractiveSession`
These are valid Python/Go fields that are not currently emitted by
`conceptInteractiveSessionPayload()`.  Track them in TypeScript now:
- `readonly axioms: string` — parses `pick(raw,'axioms','Axioms')`, fallback `''`
- `readonly cache: Record<string,boolean>` — parses `boolMap(pick(raw,'cache','Cache'))`
Once A8-Go (below) adds them to the Go payload they will auto-populate.

---

## Change B — TypeScript `uiDataModel.test.ts` (~18 new tests)

Group new tests as a second `describe('UIDataModel — wire format coverage', …)` block.

### ARG payload tests
1. **Flat ARG payload → all fields parsed** — full realistic payload, check
   `states[0].id`, `transitions[0].isJoin`, `covering[0].coveringId`, first element
2. **Missing `analysis_graph_state` → empty arrays** — payload with only `elements`
3. **`ARGSnapshot` has no `analysisGraph` property** — `expect('analysisGraph' in snapshot).toBe(false)`
4. **`CyElement.position` is null when absent** — element without `position` key
5. **`CyElement.position` parsed when present** — `{x:10,y:20}` → `position.x === 10`
6. **`State.expr` and `State.universe` preserved** — raw with `expr:"foo"` and `universe:{}`

### CyElements tests
7. **`CyElements` has no `nodeId` or `edgeId` fields** — structural regression

### `ConceptInteractiveSession` tests
8. **`abstractValue` from TagValue array** — `[{Tag:["a","b","c"],Value:true}]` →
   `abstractValue["a|b|c"] === true`
9. **`abstractValue` from TagValue with tag containing five parts** — full realistic key like
   `"edge_info|all_to_all|link|client|server"` → parses correctly
10. **`abstractValue` falls back gracefully when `abstract_value` is missing**
11. **`undoDepth`/`redoDepth` on CIS, not on base `ConceptSession`**
12. **`goalConstraints` and `supposeConstraints` as string arrays**
13. **`axioms` parses as string, `cache` parses as bool map**

### `ConceptSession` tests
14. **`ConceptSession` base class has no `undoDepth`/`redoDepth`** — structural regression

### `GraphStack` tests
15. **`GraphStack` has only four scalar fields** — no `current`/`undoStack`/`redoStack`
16. **`GraphStack` with all four fields from realistic payload**

### `ConceptSnapshot` tests
17. **`ConceptSnapshot` reads `domain` from `concept_domain` key**
18. **`ConceptSnapshot.displayCheckboxes` reads from `toggles` key when `display_checkboxes` absent**
19. **`ConceptSnapshot.displayCheckboxes` prefers `display_checkboxes` over `toggles`**
20. **Full concept payload round-trip** — realistic 15-key concept payload; check domain
    concepts, CIS abstract_value map conversion, graphStack.canUndo, edgeVisible()

---

## Change C — Go `webui_session.go` (conceptInteractiveSessionPayload)

At `conceptInteractiveSessionPayload()` (line ~593), add two fields to the returned map:
```go
"axioms": fmt.Sprint(s.Axioms),
"cache":  s.Cache,   // map[string]bool — marshals directly
```
This closes the gap where TypeScript has `axioms`/`cache` fields but Go never sent them.

---

## Change D — New `goivy/webui/webui_payload_test.go`

Go conformance tests that document and lock the wire format TypeScript depends on.
These run as part of `make test`.

```
TestARGPayloadStructure
  - marshal WebUIARGPayload with a small AnalysisGraphState
  - assert JSON has top-level keys: "analysis_graph_state", "elements", "positions"
  - assert NO "analysis_graph" key
  - assert "analysis_graph_state" contains "states", "transitions", "covering"
  - assert "elements" is an array

TestConceptPayloadStructure
  - build a concept response map via GetConcept path
  - assert top-level keys include: "concept_domain", "concept_session",
    "concept_interactive_session", "display_checkboxes", "graph_stack", "graph",
    "elements", "toggles"
  - assert "graph_stack" has keys: "can_undo", "can_redo", "undo_depth", "redo_depth"
  - assert "graph_stack" does NOT have "current"/"undo_stack"/"redo_stack"

TestConceptInteractiveSessionPayloadTagValues
  - build a ConceptInteractiveSession with a known AbstractValue TagValue slice
  - call conceptInteractiveSessionPayload()
  - marshal to JSON; unmarshal "abstract_value" as []interface{}
  - assert each entry has "Tag" (array) and "Value" (bool) fields
  - assert "undo_depth" and "redo_depth" are present as ints
  - assert "axioms" is a string (after Change C)
  - assert "cache" is a bool map (after Change C)

TestCyElementsNodeIDNotSerialized
  - build a CyElements, add a node (NodeID map is populated)
  - marshal to JSON; assert "node_id" and "edge_id" keys are absent
  - assert "elements" key is present

TestConceptGraphStackPayload
  - call conceptGraphStackPayload(nil) → assert four keys with zero values
  - call with a live GraphStack that has 2 undo entries → assert undo_depth==2, can_undo==true

TestConceptGraphPayload
  - call conceptGraphPayload(nil, nil) → assert required keys present with zero values
  - call with a Graph → assert "concept_session", "display_checkboxes", "graph_stack", "sorts"
```

---

## Verification

1. `cd ~/ivy/goivy && make test` — all Go tests must pass including new payload tests
2. `cd ~/ivy/goivy/webui && npm --prefix frontend run test` — all Vitest tests must pass
3. `npm --prefix frontend run typecheck` — TypeScript must compile with no errors after removals

Run in that order; fix any compilation errors from the removed fields (e.g., callers
of `snapshot.analysisGraph`, callers that read `nodeId`/`edgeId`) before running tests.

---

## Audit: callers to update after field removals

Before implementing, grep for uses of removed fields:
- `snapshot\.analysisGraph` — likely in `ivyRuntime.ts` or service files
- `\.nodeId\b` / `\.edgeId\b` on CyElements — check graphService, analysisStateService
- `\.undoDepth\b` / `\.redoDepth\b` on ConceptSession (not CIS) — check sheetService/
  conceptActionService
- `graphStack\.current\b` / `\.undoStack\b` / `\.redoStack\b` — check undo UI code

Each use must be updated to the replacement (either removed, or moved to the correct
class) before the code compiles.
