# Plan: Fix the 8 Data Model Limitations for Solid GUI Porting Foundation

## Context

`DATA_MODEL.md:235` lists 8 known gaps between the Go/TypeScript data model and the
Python original.  Before porting GUI behaviors, these must be resolved so service and
UI code can rely on clean, fully-typed, never-null payloads.

Exploration confirmed:
- L1, L7: trivial one-line Go fixes (CIS domain serializes to `{}`; concept_domain
  can never actually be null because `SimpleSess` is initialized in `NewSession()`).
- L2: `ARGNode.Info` is hardcoded `"State %d"` — one-line fix populates it from
  `Clauses.String()` so the formula string reaches the browser.
- L5: `FactSelection` Go struct is simple (`Index int`, `Text string`, `Selected bool`)
  — needs a matching TypeScript class and typed `facts` field on `ConceptSnapshot`.
- L3, L4, L6, L8: by-design architectural choices — close with comments/docs, no code.

---

## Confirmed Facts from Exploration

### L1 — CDConceptDomain has private fields, always serializes to `{}`
`webui_concept_domain.go:716` — `CDConceptDomain.Concepts` is `*CDConceptDict` whose
fields (`entries []CDConceptEntry`, `index map[string]int`) are unexported.  No
`MarshalJSON`, no JSON tags.  Result on wire: `"domain":{}`.  Removing the key is safe
and correct.

### L2 — ArtToGraphState hardcodes Info
`webui_ui_main.go:211` — `Info: fmt.Sprintf("State %d", st.ID)`.  The `Clauses` type
has a `String()` method (`module_clauses.go:149`) that produces
`Clauses{defs=[…], fmlas=[…]}`.  Populating `Info` with this string makes the formula
reachable in the browser via `ARGNode.info` (already on the TypeScript wire model).

### L5 — FactSelection struct (webui_graph_widget.go:38)
```go
type FactSelection struct {
    Index    int    `json:"index"`
    Text     string `json:"text"`
    Selected bool   `json:"selected"`
}
```
Populated by `GraphWidget.ConstraintFacts()` (line 596) from
`g.InteractiveSess.SupposeConstraints` — these are the suppose-constraint formula
strings with checkbox state.

### L7 — SimpleSess is always non-nil
`webui_session.go:58` — `NewSession()` always calls `SimpleSess: NewConceptSession()`.
The `"concept_domain": nil` default in `GetConcept` (line 253) is dead-code defensive.

### L3, L4, L6, L8 — accepted as by-design
- L3: `AnalysisGraph`/`AnalysisTransition` not on wire — render intermediates are
  sufficient for all listed GUI behaviors.  Full graph has no JSON tags + circular
  refs (Pred, JoinOf, Unders fields on State).
- L4: two `abstract_value` fields by design — `ConceptSession.abstractValue` is
  node_label only (for the checkbox panel), CIS.abstractValue is the full TagValue map.
- L6: `ConceptGraphModel` is zero-valued before file load — valid runtime state; TS
  zero-value handling already covers it.
- L8: `PVars []Expr` / `StateGraphs []interface{}` are interface types with no JSON
  tags, not used in webui, not required for any GUI behavior in PLAN378.

---

## Files to Modify

| File | Change |
|------|--------|
| `~/ivy/goivy/webui/webui_session.go` | L1: remove `"domain"` from CIS payload (both nil and live branches) |
| `~/ivy/goivy/webui/webui_ui_main.go` | L2: populate ARGNode.Info from Clauses.String() |
| `~/ivy/goivy/webui/webui_backend_go.go` | L7: eliminate `nil` concept_domain default |
| `~/ivy/goivy/webui/webui_payload_test.go` | Go tests for L1, L2, L7 |
| `~/ivy/goivy/webui/frontend/src/models/uiDataModel.ts` | L5: add FactSelection class; retype facts; L1/L3 comments |
| `~/ivy/goivy/webui/frontend/src/models/uiDataModel.test.ts` | TS tests for L5 |
| `~/ivy/goivy/webui/DATA_MODEL.md` | Replace limitations list with fixed vs accepted entries |

---

## Change 1 — Go: Remove `domain` from CIS payload (`webui_session.go`)

`conceptInteractiveSessionPayload()` has two branches: nil-guard return and live return.
Remove the `"domain"` key from **both**.

Nil-guard (lines ~585-591): currently returns 5 keys — remove none (domain not present here, already clean).
Live return (lines ~593-604): remove `"domain": s.Domain`.

Result: CIS payload on wire: `abstract_value`, `axioms`, `cache`, `goal_constraints`,
`info`, `redo_depth`, `state`, `suppose_constraints`, `undo_depth` — 9 keys, no domain.

TypeScript change: Add a comment to `ConceptInteractiveSession` constructor — the
`domain` field (inherited from `ConceptSession`) will always be empty `ConceptDomain`
because Go no longer sends it.  No code change to parsing — `super(raw)` will silently
produce an empty domain which is correct.  **Use `snapshot.domain` from the top-level
`concept_domain` key for all domain reads.**

---

## Change 2 — Go: Populate ARGNode.Info with formula (`webui_ui_main.go`)

In `ArtToGraphState()` at line ~211, change:
```go
Info: fmt.Sprintf("State %d", st.ID),
```
to:
```go
Info: argNodeInfo(st),
```

Add helper function:
```go
func argNodeInfo(st *goivy.State) string {
    if st.Clauses != nil {
        return st.Clauses.String()
    }
    return fmt.Sprintf("State %d", st.ID)
}
```

This makes the abstract formula for each ARG node reachable in the browser via
`ARGNode.info` (string, already on the TypeScript wire model).  TypeScript needs no
changes — `ARGNode.info` is already parsed correctly.

---

## Change 3 — Go: Eliminate null concept_domain (`webui_backend_go.go`)

In `GetConcept()` around lines 251-273, replace the two-step
`"concept_domain": nil` + conditional override pattern with a single assignment:

```go
conceptDomain := NewConceptDomain()      // always non-nil fallback
if sess.SimpleSess != nil {              // defensive guard (SimpleSess is always set)
    conceptDomain = sess.SimpleSess.Domain
}
```

Then include `"concept_domain": conceptDomain` in the initial response map (not the
override block).  Remove the separate `if sess.SimpleSess != nil` override for
`concept_domain`.

TypeScript change: none needed — `ConceptSnapshot.domain` already handles the empty
`ConceptDomain{}` case; the `?? {}` fallback stays correct.

---

## Change 4 — TypeScript: Add `FactSelection` class (`uiDataModel.ts`)

Add after `CyElements` (before `Option`):

```typescript
export class FactSelection extends RawBackedModel<unknown> {
  readonly index: number;
  readonly text: string;
  readonly selected: boolean;

  constructor(raw: unknown = {}) {
    super(raw);
    this.index = intValue(pick(raw, 'index', 'Index'));
    this.text = stringValue(pick(raw, 'text', 'Text'));
    this.selected = boolValue(pick(raw, 'selected', 'Selected'));
  }
}
```

In `ConceptSnapshot`, change:
```typescript
readonly facts: unknown[];
```
to:
```typescript
readonly facts: FactSelection[];
```
and update the constructor:
```typescript
this.facts = typedArray(pick(raw, 'facts'), FactSelection);
```

---

## Change 5 — Go tests: Lock wire contracts (`webui_payload_test.go`)

Add three new test functions:

**`TestCISPayloadNoDomainKey`**
- Call `conceptInteractiveSessionPayload(nil)` and live CIS
- Assert JSON does NOT contain key `"domain"`
- Assert keys present: `abstract_value`, `goal_constraints`, `suppose_constraints`,
  `undo_depth`, `redo_depth`, `axioms`, `cache`, `state`, `info`

**`TestARGNodeInfoIsFormula`**
- Build a `goivy.State` with non-nil `Clauses` (use `goivy.TrueClauses(nil)`)
- Call `ArtToGraphState()` on an AnalysisGraph containing that state
- Assert that `ARGNode.Info` is NOT `"State 0"` (i.e., it contains the Clauses string)
- Assert it is a non-empty string

**`TestConceptDomainNeverNull`**
- Marshal the `GetConcept` response map built with `conceptDomain := NewConceptDomain()`
- Assert `"concept_domain"` key is present and is NOT JSON `null`
- Assert it is an object (`{}` at minimum)

---

## Change 6 — TypeScript tests: FactSelection (`uiDataModel.test.ts`)

Add to `describe('UIDataModel — wire format coverage')`:

**`FactSelection: three-field parse`**
```typescript
const f = new FactSelection({ index: 2, text: 'forall X. p(X)', selected: true });
expect(f.index).toBe(2);
expect(f.text).toBe('forall X. p(X)');
expect(f.selected).toBe(true);
```

**`FactSelection: zero-value defaults`**
```typescript
const f = new FactSelection({});
expect(f.index).toBe(0);
expect(f.text).toBe('');
expect(f.selected).toBe(false);
```

**`ConceptSnapshot.facts parsed as FactSelection[]`**
```typescript
const snap = new ConceptSnapshot({
  facts: [
    { index: 0, text: '~r(X)', selected: true },
    { index: 1, text: 'p(a)',  selected: false },
  ],
});
expect(snap.facts).toHaveLength(2);
expect(snap.facts[0]).toBeInstanceOf(FactSelection);
expect(snap.facts[0].text).toBe('~r(X)');
expect(snap.facts[1].selected).toBe(false);
```

---

## Change 7 — Docs: Update DATA_MODEL.md limitations section

Replace the current "8 known limitations" list with two subsections:

**Fixed limitations** (L1, L2, L5, L7) — one-line each, pointing to the Go/TS change.

**Accepted as by-design** (L3, L4, L6, L8) — one-line rationale each:
- L3: AnalysisGraph not on wire — render intermediates sufficient; State has circular refs
- L4: abstract_value split — ConceptSession = node_label only (checkbox panel), CIS = full
- L6: ConceptGraphModel zero-values before file load — valid; TS zero-value handling covers it
- L8: pvars/state_graphs — interface types, not needed for any GUI behavior

---

## Wire Format After All Fixes

### CIS payload (no `domain` key):
```json
{
  "abstract_value": [{"Tag":["node_info","at_least_one","client"],"Value":true}],
  "axioms": "",
  "cache": {},
  "goal_constraints": [],
  "info": "",
  "redo_depth": 0,
  "state": "",
  "suppose_constraints": [],
  "undo_depth": 0
}
```

### ARGNode (Info now carries formula):
```json
{"id": 3, "label": "3", "is_bottom": false, "info": "Clauses{defs=[], fmlas=[pre(x)]}"}
```

### concept_domain (never null):
```json
{"concepts": {}, "combiners": [], "nodes": [], "edges": [], "node_labels": []}
```

### facts (now typed on both sides):
```json
[{"index": 0, "text": "~r(X)", "selected": true}]
```

---

## Verification

```
cd ~/ivy/goivy && make test
cd ~/ivy/goivy/webui && npm --prefix frontend run test
npm --prefix ~/ivy/goivy/webui/frontend run typecheck
```

All three must pass with no new failures.  The new Go tests lock the three fixed wire
contracts; the new TS tests lock `FactSelection` parsing.
