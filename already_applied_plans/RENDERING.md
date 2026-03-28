# Plan: Faithful render_rg — Use Open Formula Strings Instead of Clause Counts

**Created:** 2026-03-28 20:15

## Context

Python's `render_rg` (ivy_art.py:459-522) produces CyElements where each state node's `long_info` is a **list of formula strings** from `s.clauses.to_open_formula()`. The Go `AsCyElements` instead shows `"3 (5 clauses)"` — just a count. This loses the actual formula content that users need to see when clicking a state in the web UI.

The JS `showInfo` function (ivyweb_controls.js:291-308) already handles both `Array.isArray(longInfo)` and plain strings, so the frontend is ready for array `long_info`.

## What Python Does

```python
# ivy_art.py:466-479
g.add_node(
    obj=s,
    label=str(s.id),
    classes=['bottom_state'] if s.is_bottom() else ['state'],
    short_info=str(s.id),
    long_info=[str(x) for x in s.clauses.to_open_formula()],  # <-- list of formula strings
    locked=True,
)
```

`clauses.to_open_formula()` returns an `And` object. Iterating it yields each conjunct. `str(x)` on each conjunct produces a human-readable formula string.

## What Go Does Today (Broken)

```go
// art.go AsCyElements, line ~1286
info = fmt.Sprintf("%d (%d clauses)", s.ID, len(s.Clauses.Fmlas))
```

Just shows count. The Go `clauseops.Clauses` already has `ToOpenFormula()` at `clauseops/clauses.go:91` which returns `*lg.And` with `.Terms []lg.Expr`. Each term has a `.String()` method.

---

## Implementation

### Step 1: Change CyElements.AddNode to accept `interface{}` for longInfo

**File:** `~/goivy/art/cyrender.go`

Currently `longInfo` is `string`. Python passes a `[]string`. The JSON serialization and JS frontend both handle arrays. Change the parameter type:

```go
// Before:
func (g *CyElements) AddNode(obj, label string, classes []string, shortInfo, longInfo string, ...)

// After:
func (g *CyElements) AddNode(obj, label string, classes []string, shortInfo string, longInfo interface{}, ...)
```

The `data["long_info"] = longInfo` assignment already stores whatever type is passed, so the body doesn't change beyond the signature.

Do the same for `AddNodeWithColor` which has the same parameter.

### Step 2: Fix AsCyElements to pass formula strings

**File:** `~/goivy/art/art.go` — method `AsCyElements`

Replace the clause-count logic with actual formula string extraction:

```go
// Current (broken):
info := fmt.Sprintf("%d", s.ID)
if s.Clauses != nil {
    info = fmt.Sprintf("%d (%d clauses)", s.ID, len(s.Clauses.Fmlas))
}

// New (faithful):
shortInfo := fmt.Sprintf("%d", s.ID)
var longInfo interface{} = shortInfo
if s.Clauses != nil {
    openFmla := s.Clauses.ToOpenFormula()
    if and, ok := openFmla.(*lg.And); ok && len(and.Terms) > 0 {
        fmlaStrings := make([]string, len(and.Terms))
        for i, term := range and.Terms {
            fmlaStrings[i] = term.String()
        }
        longInfo = fmlaStrings
    }
}
```

Then pass `shortInfo` and `longInfo` separately to `AddNode` (today the same `info` string is used for both).

### Step 3: Fix AsCyElements to produce CyElements directly (not via intermediate AnalysisGraphState)

The current Go code builds an intermediate `AnalysisGraphState` struct, then calls `RenderARG` to convert it to CyElements. Python's `render_rg` builds CyElements directly. The intermediate step loses information (ARGNode.Info is a single string, not []string).

Rewrite `AsCyElements` to build CyElements directly, matching Python's `render_rg`:

```go
func (ag *AnalysisGraph) AsCyElements(dotLayout func(*CyElements) *CyElements) *CyElements {
    g := NewCyElements()

    // Add nodes for states — matches Python ivy_art.py:467-479
    for _, s := range ag.States {
        var classes []string
        if s.IsBottom() {
            classes = []string{"bottom_state"}
        } else {
            classes = []string{"state"}
        }

        shortInfo := fmt.Sprintf("%d", s.ID)
        var longInfo interface{} = shortInfo
        if s.Clauses != nil {
            openFmla := s.Clauses.ToOpenFormula()
            if and, ok := openFmla.(*lg.And); ok && len(and.Terms) > 0 {
                fmlaStrings := make([]string, len(and.Terms))
                for i, term := range and.Terms {
                    fmlaStrings[i] = term.String()
                }
                longInfo = fmlaStrings
            }
        }

        g.AddNode(
            fmt.Sprintf("state_%d", s.ID),
            fmt.Sprintf("%d", s.ID),
            classes,
            shortInfo,
            longInfo,
            nil,      // actions
            "ellipse",
        )
    }

    // Add edges for transitions — matches Python ivy_art.py:482-506
    for _, t := range ag.Transitions {
        sourceKey := fmt.Sprintf("state_%d", -1)
        targetKey := fmt.Sprintf("state_%d", -1)
        if t.Pre != nil {
            sourceKey = fmt.Sprintf("state_%d", t.Pre.ID)
        }
        if t.Post != nil {
            targetKey = fmt.Sprintf("state_%d", t.Post.ID)
        }

        var label, info string
        var classes []string

        if t.Label == "join" {
            classes = []string{"transition_join"}
            label = "join"
            info = "join"
        } else {
            classes = []string{"transition_action"}
            label = t.Label
            if label == "" {
                label = "(unlabeled)"
            }
            // Python: label.replace('}',']-').replace('{','-[')
            label = strings.ReplaceAll(label, "}", "]-")
            label = strings.ReplaceAll(label, "{", "-[")
            // Python: label.replace('\n','\\l')+'\\l'
            label = strings.ReplaceAll(label, "\n", "\\l") + "\\l"
            if t.Op != nil {
                info = t.Op.String()
            } else {
                info = label
            }
        }

        edgeKey := fmt.Sprintf("tr_%s_%s", sourceKey, targetKey)
        g.AddEdge(edgeKey, sourceKey, targetKey, label, classes, info, info)
    }

    // Add edges for covering — matches Python ivy_art.py:509-520
    for _, c := range ag.Covering {
        coveredKey := fmt.Sprintf("state_%d", -1)
        coveringKey := fmt.Sprintf("state_%d", -1)
        if c.Covered != nil {
            coveredKey = fmt.Sprintf("state_%d", c.Covered.ID)
        }
        if c.Covering != nil {
            coveringKey = fmt.Sprintf("state_%d", c.Covering.ID)
        }
        edgeKey := fmt.Sprintf("cover_%s_%s", coveredKey, coveringKey)
        g.AddEdge(edgeKey, coveredKey, coveringKey, "", []string{"cover"}, "", "")
    }

    if dotLayout != nil {
        g = dotLayout(g)
    }
    return g
}
```

### Step 4: Fix RenderRg in phase7.go

`RenderRg` is a text-based debug function. It should also show formula strings instead of clause counts:

```go
func RenderRg(rg *AnalysisGraph) string {
    // ... states section:
    for _, s := range rg.States {
        label := fmt.Sprintf("  [%d]", s.ID)
        if s.IsBottom() {
            label += " (bottom)"
        } else if s.Clauses != nil {
            openFmla := s.Clauses.ToOpenFormula()
            if and, ok := openFmla.(*lg.And); ok {
                for _, term := range and.Terms {
                    label += fmt.Sprintf("\n    %s", term.String())
                }
            }
        }
        sb.WriteString(label + "\n")
    }
    // transitions and covering sections stay the same
}
```

### Step 5: Update AddEdge longInfo parameter for consistency

For edges, Python passes the same `info` string for both `short_info` and `long_info`. The Go `AddEdge` already takes two string params for this. No change needed for edges.

### Step 6: Update callers of AddNode

Search for all callers of `AddNode` and `AddNodeWithColor` — their `longInfo` argument type changes from `string` to `interface{}`. Most callers pass a string, which is still valid for `interface{}`.

**Known callers:**
- `art/cyrender.go` — `RenderARG` function
- `webui/concept.go` or similar — concept graph rendering

---

## Files Modified

| File | Changes |
|------|---------|
| `~/goivy/art/cyrender.go` | Change `AddNode` and `AddNodeWithColor` longInfo param from `string` to `interface{}` |
| `~/goivy/art/art.go` | Rewrite `AsCyElements` to build CyElements directly with formula strings |
| `~/goivy/art/phase7.go` | Update `RenderRg` to show formula strings instead of clause counts |
| External callers of `AddNode`/`AddNodeWithColor` | May need minor type adjustments (string → interface{} is backward-compatible) |

## Existing Facilities to Reuse

| Function | Location | Purpose |
|----------|----------|---------|
| `clauseops.Clauses.ToOpenFormula()` | `clauseops/clauses.go:91` | Returns `*lg.And` with conjuncts from defs + fmlas |
| `lg.Expr.String()` | `logic/formula.go` | Human-readable formula strings |
| `art.CyElements.AddNode` | `art/cyrender.go:49` | Add node to Cytoscape graph |
| `art.CyElements.AddEdge` | `art/cyrender.go:135` | Add edge to Cytoscape graph |

## Verification

1. `cd ~/goivy && go build ./...` — compilation
2. `cd ~/goivy && make test` — full test suite
3. Add a test that creates an AnalysisGraph with states that have non-trivial clauses, calls `AsCyElements`, and verifies `long_info` contains formula strings (not clause counts)
4. Add a test for `RenderRg` verifying formula strings appear in the text output
