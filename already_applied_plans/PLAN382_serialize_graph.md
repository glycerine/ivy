# Plan: Full CTI Support — FullARGNode Wire Format + CTI Endpoint

## Context

The Go ARG endpoint currently sends only lightweight render intermediates
(`ARGNode`: id/label/is_bottom/info).  CTI (counterexample-to-induction) inspection
requires the browser to have the abstract-state formula, the action that produced each
state, and the concrete model universe from BMC results.  The `AnalysisGraph` and
`AnalysisTransition` TypeScript classes exist but are disconnected from the wire.

This plan fixes the gap by (1) defining a `FullARGNode` wire struct that is a superset
of `ARGNode`, (2) extending the existing ARG endpoint with `?full=true`, (3) adding a
dedicated `GET /api/session/{id}/arg/cti` endpoint for the CTI analysis graph, and (4)
adding matching TypeScript classes and tests.

---

## Confirmed Facts from Exploration

**`State.Universe` concrete type:** `map[string][]goivy.Expr` (sort → universe elements).
Assigned from `bmcRes.Universes` in `art.go:926`, `interp_phase4.go:257`, `trace.go:760`.
Also `map[string]interface{}` in `mc_phase7.go:275` (initial state).
Serialized as `map[string][]string` using `fmt.Sprint(expr)` on each element.
`structureUniverseConsts()` at `webui_session.go:812` already handles both types — we
reuse that pattern.

**`Clauses.String()`** produces debug format `Clauses{defs=[…], fmlas=[…]}`.
Use `clauses.ToOpenFormula()` instead — returns an `Expr` whose `.String()` produces
`p(X) & q(Y)` style output.  Fall back to `Clauses.String()` if `ToOpenFormula()`
returns nil.

**`State.Safe`:** No `safe bool` field on Go's `State` struct (Python has it, Go does
not yet).  `IsBottom()` (`art.go:145`) = Clauses.IsFalse() is the correct unsafe-state
signal and is already on the wire.  Skip `safe` until Go adds per-state safety tracking.

**`State.ActionName string`:** Set at `art.go:1809-1810` when a state is derived from
an action.  Always available; empty string for initial state.

**Pred/JoinOf/Unders:** NOT needed for CTI display per Python GUI analysis.  Circular
pointer refs; serialized as integer IDs only when needed.

**CTI analysis graph:** Lives in `sess.CTIUI.AnalysisGraphUI.AG` (separate from main
`sess.AGUI`).  `sess.CTIUI.HaveCTI bool` and `sess.CTIUI.CurrentConjecture *goivy.Clauses`
track CTI state.  No dedicated HTTP endpoint exists yet.

**Router pattern:** `webui_server.go` switch on `rest` string; adding `case "arg/cti":`
suffices.  Query param `?full=true` matches existing `?node=X` pattern in concept endpoint.

**Backend interface:** `GoBackend.GetARG()` signature must stay compatible with any
`Backend` interface — check `webui_backend.go` or similar for interface definition and
update it.

---

## Files to Modify

| File | Change |
|------|--------|
| `~/ivy/goivy/art_cyrender.go` | Add `FullARGNode`, `FullAnalysisGraphState` structs (goivy pkg) |
| `~/ivy/goivy/webui/webui_ui_main.go` | Add `ArtToFullGraphState()`, `argNodeClauses()`, `argNodeUniverse()` |
| `~/ivy/goivy/webui/webui_cyrender.go` | Add `FullARGPayload()`, `FullAnalysisUIARGPayload()` |
| `~/ivy/goivy/webui/webui_backend_go.go` | Update `GetARG(full bool)`, add `GetCTIARG()` |
| `~/ivy/goivy/webui/webui_backend.go` (or interface file) | Update `Backend` interface to match |
| `~/ivy/goivy/webui/webui_handlers.go` | Update `apiARG(?full=true)`, add `apiCTIARG()` |
| `~/ivy/goivy/webui/webui_server.go` | Add `case "arg/cti":` route |
| `~/ivy/goivy/webui/webui_payload_test.go` | New Go tests |
| `~/ivy/goivy/webui/frontend/src/models/uiDataModel.ts` | Add `FullARGNode`, `FullAnalysisGraphState`, `CTISnapshot`; update `ARGSnapshot` |
| `~/ivy/goivy/webui/frontend/src/models/uiDataModel.test.ts` | New TS tests |

---

## Change 1 — Go: `FullARGNode` + `FullAnalysisGraphState` structs (`art_cyrender.go`)

Add after the existing `ARGNode`/`ARGTransition`/`ARGCover`/`AnalysisGraphState` block:

```go
// FullARGNode extends ARGNode with formula and universe data for CTI inspection.
// Universe is omitted from JSON when nil (no BMC has run).
type FullARGNode struct {
    ID         int                 `json:"id"`
    Label      string              `json:"label"`
    IsBottom   bool                `json:"is_bottom"`
    Info       string              `json:"info"`
    Clauses    string              `json:"clauses"`
    ActionName string              `json:"action_name"`
    Universe   map[string][]string `json:"universe,omitempty"`
}

type FullAnalysisGraphState struct {
    States      []FullARGNode   `json:"states"`
    Transitions []ARGTransition `json:"transitions"`
    Covering    []ARGCover      `json:"covering"`
}

func NewFullAnalysisGraphState() *FullAnalysisGraphState {
    return &FullAnalysisGraphState{
        States:      []FullARGNode{},
        Transitions: []ARGTransition{},
        Covering:    []ARGCover{},
    }
}
```

The `WebUI*` type aliases in `webui_cyrender.go` do NOT need to be extended —
`FullARGNode` is a separate, parallel type used only when `?full=true` is requested.

---

## Change 2 — Go: Conversion helpers (`webui_ui_main.go`)

Add after the existing `ArtToGraphState()` and `argNodeInfo()` functions:

```go
// ArtToFullGraphState converts a goivy.AnalysisGraph to FullAnalysisGraphState,
// including formula strings and universe data for CTI inspection.
func ArtToFullGraphState(ag *goivy.AnalysisGraph) *FullAnalysisGraphState {
    gs := NewFullAnalysisGraphState()
    for _, st := range ag.States {
        label := st.Label
        if label == "" {
            label = fmt.Sprintf("%d", st.ID)
        }
        gs.States = append(gs.States, FullARGNode{
            ID:         st.ID,
            Label:      label,
            IsBottom:   st.IsBottom(),
            Info:       argNodeInfo(st),
            Clauses:    argNodeClauses(st),
            ActionName: st.ActionName,
            Universe:   argNodeUniverse(st),
        })
    }
    for _, t := range ag.Transitions {
        gs.Transitions = append(gs.Transitions, WebUIARGTransition{
            SourceID: t.Pre.ID,
            TargetID: t.Post.ID,
            Label:    t.Label,
        })
    }
    for _, c := range ag.Covering {
        gs.Covering = append(gs.Covering, WebUIARGCover{
            CoveredID:  c.Covered.ID,
            CoveringID: c.Covering.ID,
        })
    }
    return gs
}

func argNodeClauses(st *goivy.State) string {
    if st.Clauses == nil {
        return ""
    }
    if f := st.Clauses.ToOpenFormula(); f != nil {
        return fmt.Sprint(f)
    }
    return st.Clauses.String()
}

func argNodeUniverse(st *goivy.State) map[string][]string {
    if st.Universe == nil {
        return nil
    }
    switch u := st.Universe.(type) {
    case map[string][]goivy.Expr:
        out := make(map[string][]string, len(u))
        for sort, exprs := range u {
            strs := make([]string, len(exprs))
            for i, e := range exprs {
                strs[i] = fmt.Sprint(e)
            }
            out[sort] = strs
        }
        return out
    default:
        return nil
    }
}
```

---

## Change 3 — Go: `FullARGPayload` + `FullAnalysisUIARGPayload` (`webui_cyrender.go`)

Add after the existing `WebUIARGPayload` and `AnalysisUIARGPayload`:

```go
// FullARGPayload is the full version of WebUIARGPayload — same shape but uses
// FullAnalysisGraphState (which includes clauses, action_name, universe per node).
func FullARGPayload(state *FullAnalysisGraphState, cy *WebUICyElements) map[string]interface{} {
    if state == nil {
        state = NewFullAnalysisGraphState()
    }
    if cy == nil {
        cy = &WebUICyElements{}
    }
    if cy.Elements == nil {
        cy.Elements = []WebUICyElement{}
    }
    payload := map[string]interface{}{
        "elements": cy.Elements,
    }
    if len(state.States) > 0 || len(state.Transitions) > 0 || len(state.Covering) > 0 {
        payload["analysis_graph_state"] = state
    }
    return payload
}

func FullAnalysisUIARGPayload(ui *AnalysisGraphUI) map[string]interface{} {
    if ui == nil || ui.AG == nil {
        return FullARGPayload(NewFullAnalysisGraphState(), nil)
    }
    state := ArtToFullGraphState(ui.AG)
    cy := RenderAnalysisUIARG(ui)
    return FullARGPayload(state, cy)
}
```

---

## Change 4 — Go: Update `GetARG`, add `GetCTIARG` (`webui_backend_go.go`)

Update `GetARG` signature to accept a `full bool` parameter:

```go
func (gbe *GoBackend) GetARG(sessionID string, full bool) (by []byte, err error) {
    gbe.do(func(b *GoBackend) error {
        var sess *Session
        sess, err = b.getSession(sessionID)
        if err != nil {
            return nil
        }
        if full {
            by, err = canonicalJSON(FullAnalysisUIARGPayload(sess.AGUI))
        } else {
            by, err = canonicalJSON(AnalysisUIARGPayload(sess.AGUI))
        }
        return nil
    })
    return
}
```

Add `GetCTIARG`:

```go
func (gbe *GoBackend) GetCTIARG(sessionID string) (by []byte, err error) {
    gbe.do(func(b *GoBackend) error {
        var sess *Session
        sess, err = b.getSession(sessionID)
        if err != nil {
            return nil
        }
        haveCTI := sess.CTIUI != nil && sess.CTIUI.HaveCTI
        conjectureStr := ""
        if sess.CTIUI != nil && sess.CTIUI.CurrentConjecture != nil {
            if f := sess.CTIUI.CurrentConjecture.ToOpenFormula(); f != nil {
                conjectureStr = fmt.Sprint(f)
            } else {
                conjectureStr = sess.CTIUI.CurrentConjecture.String()
            }
        }
        var argPayload map[string]interface{}
        if sess.CTIUI != nil && sess.CTIUI.AnalysisGraphUI != nil {
            argPayload = FullAnalysisUIARGPayload(sess.CTIUI.AnalysisGraphUI)
        } else {
            argPayload = FullARGPayload(NewFullAnalysisGraphState(), nil)
        }
        argPayload["have_cti"] = haveCTI
        argPayload["current_conjecture"] = conjectureStr
        by, err = canonicalJSON(argPayload)
        return nil
    })
    return
}
```

**Also:** Search for the `Backend` interface definition (likely in `webui_backend.go` or
`webui_server.go`) and update `GetARG` signature there too.

---

## Change 5 — Go: Update handler + add CTI route (`webui_handlers.go`, `webui_server.go`)

In `webui_handlers.go`, update `apiARG`:
```go
func (s *Server) apiARG(w http.ResponseWriter, r *http.Request, sessionID string) {
    if r.Method != http.MethodGet {
        writeErr(w, http.StatusMethodNotAllowed, "GET required")
        return
    }
    full := r.URL.Query().Get("full") == "true"
    data, err := s.backend.GetARG(sessionID, full)
    if err != nil {
        writeBackendErr(w, err)
        return
    }
    writeBackend(w, data)
}
```

Add `apiCTIARG`:
```go
func (s *Server) apiCTIARG(w http.ResponseWriter, r *http.Request, sessionID string) {
    if r.Method != http.MethodGet {
        writeErr(w, http.StatusMethodNotAllowed, "GET required")
        return
    }
    data, err := s.backend.GetCTIARG(sessionID)
    if err != nil {
        writeBackendErr(w, err)
        return
    }
    writeBackend(w, data)
}
```

In `webui_server.go`, add route to the switch:
```go
case "arg/cti":
    s.apiCTIARG(w, r, sid)
```

---

## Change 6 — TypeScript: `FullARGNode`, `FullAnalysisGraphState`, `CTISnapshot` (`uiDataModel.ts`)

Add `stringArrayMap` helper (after existing helpers):
```typescript
function stringArrayMap(value: unknown): Record<string, string[]> {
  const obj = rawRecord(value);
  const out: Record<string, string[]> = {};
  for (const key of Object.keys(obj)) {
    out[key] = stringArray(obj[key]);
  }
  return out;
}
```

Add `FullARGNode` class (after existing `ARGNode`):
```typescript
// FullARGNode is a superset of ARGNode — sent when ?full=true is requested.
// clauses: clean formula string (ToOpenFormula().String()).
// actionName: the action that produced this state ("" for initial state).
// universe: sort → concrete element strings from BMC model (null if no BMC run).
export class FullARGNode extends RawBackedModel<unknown> {
  readonly id: number;
  readonly label: string;
  readonly isBottom: boolean;
  readonly info: string;
  readonly clauses: string;
  readonly actionName: string;
  readonly universe: Record<string, string[]> | null;

  constructor(raw: unknown = {}) {
    super(raw);
    this.id = intValue(pick(raw, 'id', 'ID'), -1);
    this.label = stringValue(pick(raw, 'label', 'Label'));
    this.isBottom = boolValue(pick(raw, 'is_bottom', 'isBottom', 'IsBottom'));
    this.info = stringValue(pick(raw, 'info', 'Info'));
    this.clauses = stringValue(pick(raw, 'clauses', 'Clauses'));
    this.actionName = stringValue(pick(raw, 'action_name', 'actionName', 'ActionName'));
    const u = pick(raw, 'universe', 'Universe');
    this.universe = u ? stringArrayMap(u) : null;
  }
}
```

Add `FullAnalysisGraphState` (after `AnalysisGraphState`):
```typescript
export class FullAnalysisGraphState extends RawBackedModel<unknown> {
  readonly states: FullARGNode[];
  readonly transitions: ARGTransition[];
  readonly covering: ARGCover[];

  constructor(raw: unknown = {}) {
    super(raw);
    this.states = typedArray(pick(raw, 'states', 'States'), FullARGNode);
    this.transitions = typedArray(pick(raw, 'transitions', 'Transitions'), ARGTransition);
    this.covering = typedArray(pick(raw, 'covering', 'Covering'), ARGCover);
  }
}
```

Update `ARGSnapshot` to carry both lightweight and full state:
```typescript
export class ARGSnapshot extends RawBackedModel<unknown> {
  readonly render: CyElements;
  readonly analysisGraphState: AnalysisGraphState;       // lightweight (always parsed)
  readonly fullAnalysisGraphState: FullAnalysisGraphState; // superset (clauses empty when ?full not used)

  constructor(raw: unknown = {}) {
    super(raw);
    this.render = new CyElements(raw);
    const ags = pick(raw, 'analysis_graph_state', 'analysisGraphState') ?? {};
    this.analysisGraphState = new AnalysisGraphState(ags);
    this.fullAnalysisGraphState = new FullAnalysisGraphState(ags);
    // Note: no 'analysis_graph' key exists in the Go wire format.
  }
}
```

Add `CTISnapshot` (after `ARGSnapshot`):
```typescript
export class CTISnapshot extends ARGSnapshot {
  readonly haveCti: boolean;
  readonly currentConjecture: string;

  constructor(raw: unknown = {}) {
    super(raw);
    this.haveCti = boolValue(pick(raw, 'have_cti', 'haveCti'));
    this.currentConjecture = stringValue(pick(raw, 'current_conjecture', 'currentConjecture'));
  }
}
```

In `SheetModel`, add `cti: CTISnapshot | null` field (initialized to `null`).

In `UIDataModel`, add:
```typescript
acceptCtiSnapshot(sheetId: string, payload: unknown = {}): CTISnapshot {
  const sheet = this.registerSheet(sheetId || this.activeSheetId);
  sheet.cti = new CTISnapshot(payload);
  return sheet.cti;
}
```

---

## Change 7 — Go tests (`webui_payload_test.go`)

**`TestFullARGNodeUniverse`**
- Build a `goivy.State` with `Universe = map[string][]goivy.Expr{"node": {someExpr}}`
- Call `argNodeUniverse(st)`; assert result is non-nil map with key `"node"`

**`TestFullARGPayloadClauses`**
- Build a `goivy.State` with non-nil Clauses
- Call `ArtToFullGraphState()` on an AnalysisGraph with that state
- Assert `FullARGNode.Clauses` is non-empty string (not the debug `Clauses{…}` format
  if `ToOpenFormula` returns something useful, though accepting either is fine)

**`TestFullARGPayloadStructure`**
- Call `FullAnalysisUIARGPayload` with a non-empty AnalysisGraphUI
- Assert JSON has `"elements"` and `"analysis_graph_state"`
- Assert `analysis_graph_state.states[0]` has keys `"clauses"` and `"action_name"`
- Assert NO `"universe"` key when universe is nil (`omitempty`)

**`TestCTIARGPayloadShape`**
- Call `GetCTIARG` on a session
- Assert JSON has `"have_cti"` (bool), `"current_conjecture"` (string), `"elements"`

---

## Change 8 — TypeScript tests (`uiDataModel.test.ts`)

Add to `describe('UIDataModel — wire format coverage')`:

```typescript
it('FullARGNode: clauses, action_name, universe parsed', () => { … })
it('FullARGNode: universe is null when absent', () => { … })
it('FullARGNode: universe Record<string,string[]> when present', () => { … })
it('FullAnalysisGraphState: reuses ARGTransition and ARGCover', () => { … })
it('ARGSnapshot: fullAnalysisGraphState populated alongside analysisGraphState', () => { … })
it('CTISnapshot: haveCti and currentConjecture parsed', () => { … })
it('CTISnapshot: extends ARGSnapshot — render elements present', () => { … })
```

---

## Wire Formats After This Plan

### `GET /api/session/{id}/arg?full=true`
```json
{
  "elements": […],
  "analysis_graph_state": {
    "states": [
      { "id": 3, "label": "3", "is_bottom": false, "info": "Clauses{…}",
        "clauses": "pre(x) & inv(y)", "action_name": "send",
        "universe": { "node": ["n0", "n1"], "data": ["d0"] } }
    ],
    "transitions": [{ "source_id": 0, "target_id": 3, "label": "send", "is_join": false }],
    "covering": []
  }
}
```

### `GET /api/session/{id}/arg/cti`
```json
{
  "elements": […],
  "analysis_graph_state": {
    "states": [
      { "id": 0, "label": "pre", "is_bottom": false, "clauses": "inv(x)",
        "universe": { "node": ["n0"] } },
      { "id": 1, "label": "post", "is_bottom": true, "clauses": "false",
        "universe": { "node": ["n0"] } }
    ],
    "transitions": [{ "source_id": 0, "target_id": 1, "label": "step", "is_join": false }],
    "covering": []
  },
  "have_cti": true,
  "current_conjecture": "forall X. inv(X)"
}
```

---

## Verification

```
cd ~/ivy/goivy && make test
cd ~/ivy/goivy/webui && npm --prefix frontend run test
npm --prefix ~/ivy/goivy/webui/frontend run typecheck
```

Also manually test:
- `GET /api/session/{id}/arg` — unchanged lightweight response
- `GET /api/session/{id}/arg?full=true` — nodes have `clauses` and `action_name`
- `GET /api/session/{id}/arg/cti` — `have_cti`, `current_conjecture`, full nodes
