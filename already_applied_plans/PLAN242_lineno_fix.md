# Plan: Carry Location forward through node creation (fixes "(internal)" placeholder for definitions)

Created: 2026-04-09 11:45

## Context

In `~/ivy/goivy/log.red` the non-xtracer console output for `TestOrdLive` shows Go printing `(internal)` for the definitions list while Python prints the proper `<file>: line N: ` prefix:

```
~go[after i=236093]:     The following definitions are used:
~go[after i=236093]:         (internal) ref.def80
~go[after i=236093]:         (internal) ref.def81
...

~py[after i=236093]:     The following definitions are used:
~py[after i=236093]:         <IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line 168: ref.def80
~py[after i=236093]:         <IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line 177: ref.def81
...
```

The "(internal)" sentinel is emitted by `check.PrettyLineno` (`check/helpers.go:39-48`) when an `ast.LabeledFormula`'s `Base.Loc` has both an empty `Filename` and a zero `Line`. The Python source-of-truth `pretty_lineno` (`pyivy/ivy/ivy/ivy_check.py:259-260`) prints `str(ast.lineno)`, where `ast.lineno` is a `LocationTuple([filename, line])` set during parsing — it produces the full `<file>: line N: ` form.

### Why Go currently loses the location for definitions

Definitions reach `mod.Definitions` from a `LabeledFormula` built in the v17 grammar's DEFINITION rule. In `parser/grammar_v17.y:980-983` the rule does:

```go
lf := acfg(v17lex).NewLabeledFormula($4, gdefn)
lf.Lineno = tokLineno(v17lex.(*v17LexAdapter), $3).Line   // ← only the int field
lf = addLabel(acfg(v17lex), lf, "def")
dd := acfg(v17lex).NewDefinitionDecl(lf)
```

`tokLineno(...)` returns a fully-populated `ast.Location{Filename: <normalized>, Line: N}`, but the rule throws away the filename by storing only `.Line` into the redundant `LabeledFormula.Lineno` (int) field. The actual `Base.Loc` (which is what `GetLineno()`/`PrettyLineno` read) is never set, so `HasLoc=false` and `Loc.Filename==""`. The `addLabel` helper at `parser/grammar_v17.y:73-84` checks `HasLocSet()` before propagating Loc, so the new labelled LF inherits the unset Loc. From there `mod.Definitions` carries an LF with no Location all the way to the `pretty_lineno` call.

The same shape — `clf.Lineno = src.GetLineno().Line` (or similar) without a matching `SetLineno(src.GetLineno())` — recurs in several compiler sites: `compiler/decl.go:1310,1341,1507`, `compiler/phase6.go:2279`, `compiler/ivy_compile.go:1884`, `check/isolate_check.go:78`. Every such site is a place where the source-file Location is silently dropped, even though the surrounding code already has the full `Location` in hand.

### Why this generalizes

The root cause is that `LabeledFormula` carries two parallel storages for the same datum:

1. `LabeledFormula.Lineno` (int) — a direct field on the struct.
2. `LabeledFormula.Base.Loc` (`ast.Location`) — the fully-typed Location with filename + line + reference chain.

Python only has `lf.lineno` (a `LocationTuple`); the int field in Go is a redundant duplicate that drifts out of sync because nothing forces a single setter to update both. The general fix is: **anywhere we know a source Location, always store it as a `Location` via `SetLineno`, never as a bare int via `lf.Lineno = ...`**. If we eliminate the int-only write path, `Base.Loc` becomes the single source of truth and the "(internal)" placeholder can no longer appear when the parser/compiler had the location available.

CLAUDE.md rule A: Python is the source of truth. CLAUDE.md rule 2: Python field names map to Go field names. Python's `lf.lineno` is a `LocationTuple`, so the Go equivalent should be a single `Location`-typed slot, not a `Location`+`int` pair.

## Plan

The fix has two parts: a surgical fix at the parser site that produces the visible "(internal)" output, plus a small audit to eliminate every other call site that writes `*.Lineno = <int>` without also storing the full `Location`. This generalizes the fix so future edits cannot regress the same way.

### Step 1 — Parser DEFINITION rule: store the full Location

In `parser/grammar_v17.y` around line 980, replace:

```go
lf := acfg(v17lex).NewLabeledFormula($4, gdefn)
lf.Lineno = tokLineno(v17lex.(*v17LexAdapter), $3).Line
lf = addLabel(acfg(v17lex), lf, "def")
```

with:

```go
loc := tokLineno(v17lex.(*v17LexAdapter), $3)
lf := acfg(v17lex).NewLabeledFormula($4, gdefn)
lf.SetLineno(loc)        // sets Base.Loc + HasLoc
lf.Lineno = loc.Line     // keep redundant int in sync until Step 4 removes it
lf = addLabel(acfg(v17lex), lf, "def")
```

`addLabel` already propagates `Base.Loc` when `HasLocSet()` is true, so the freshly-labelled LF will inherit the location through `addLabel` → `cfg.NewLabeledFormula` → `res.SetLineno(lf.GetLineno())`.

After editing the `.y` file, regenerate the goyacc output:

```
cd /Users/jaten/go/src/github.com/glycerine/ivy/goivy/parser && go generate ./...
```

This rewrites `parser/grammar_v17.go` from `grammar_v17.y`. Do **not** hand-edit the generated `.go` file.

### Step 2 — Compiler/check sites: SetLineno everywhere we know the source Location

Convert each of the following bare `*.Lineno = <int>` writes into `*.SetLineno(<source>.GetLineno())`. Each call site already holds the source node, so the change is a one-line replacement; the redundant int store can be left alongside until Step 4 removes the field.

| File | Line | Current | New |
|---|---|---|---|
| `compiler/decl.go` | 1310 | `clf.Lineno = schema.GetLineno().Line` | `clf.SetLineno(schema.GetLineno())` |
| `compiler/decl.go` | 1341 | `clf.Lineno = lf.GetLineno().Line` | `clf.SetLineno(lf.GetLineno())` |
| `compiler/decl.go` | 1507 | `mlf.Lineno = lf.GetLineno().Line` | `mlf.SetLineno(lf.GetLineno())` |
| `compiler/phase6.go` | 2279 | `newProp.Lineno = prop.Lineno` | `newProp.SetLineno(prop.GetLineno())` |
| `compiler/ivy_compile.go` | 1884 | `result.Lineno = prop.Lineno` | `result.SetLineno(prop.GetLineno())` |
| `check/isolate_check.go` | 78 | `subgoal.Lineno = pfNode.GetLineno().Line` | `subgoal.SetLineno(pfNode.GetLineno())` |

After the change at `compiler/phase6.go:2184` — `if sg.Lineno > 0 {` — switch to reading from the canonical Loc as well: `if sg.GetLineno().Line > 0 {`. This guard then gates correctly on whether a real Location is present, not on the stale int.

### Step 3 — Other parser sites that build a `LabeledFormula` without `SetLineno`

In addition to the DEFINITION rule (Step 1), three other parser productions construct an `LabeledFormula` and never call `SetLineno`. For each, set the Loc from the most natural source already in scope; this is the same pattern Python uses (`lf.lineno = get_lineno(p, …)` or `lf.lineno = p[k].lineno`):

1. `parser/grammar_v17.y:1031-1034` — `top : top TOK_PROOF labelname proofstep`
   - Add: `lf.SetLineno(tokLineno(v17lex.(*v17LexAdapter), $3))` (the labelname token).
2. `parser/grammar_v17.y:90` — `mkLF` helper used by many call sites
   - Change to: `lf := cfg.NewLabeledFormula(nil, x); if x != nil { lf.SetLineno(x.GetLineno()) }`. This propagates Loc from the wrapped formula whenever it has one (matches Python `mk_lf` which copies `lineno` from `x`).
3. `parser/grammar_v17.y:2693` — `schdecl : schdefnrhs`
   - Add: `lf.SetLineno(nodeLineno($1))` after the `NewLabeledFormula` call.

After editing the `.y` file, re-run `go generate` (Step 1).

### Step 4 — General cleanup: remove the redundant `LabeledFormula.Lineno` int field

Once all call sites consistently use `SetLineno`/`GetLineno`, the `Lineno int` field on `LabeledFormula` (`ast/decl_ast.go:25`) is dead duplication. Remove it and replace every read with `lf.GetLineno().Line`.

Concrete edits:

- `ast/decl_ast.go`
  - Delete the `Lineno int` field at line 25.
  - In `NewLabeledFormulaFrom` (line 60-71), drop `lf.Lineno = src.Lineno`; the `if src.HasLocSet() { lf.SetLineno(src.GetLineno()) }` already covers it. Tighten by always copying Loc unconditionally so an empty Location is still propagated (a no-op when source is also empty).
  - In `cloneInternal` (line 97-113), drop `Lineno: lf.Lineno`. The `Base: Base{Cfg: cfg, Loc: lf.Base.Loc, HasLoc: lf.Base.HasLoc}` line already preserves Loc.
  - Add a tiny convenience method for read sites that need just the int:
    ```go
    // LinenoLine returns the line number from the node's Location.
    // Equivalent to lf.GetLineno().Line. Provided for callers that
    // formerly read the standalone Lineno int field.
    func (lf *LabeledFormula) LinenoLine() int { return lf.Base.Loc.Line }
    ```
- All read sites (after grep `\.Lineno[^a-zA-Z=]`):
  - `compiler/phase6.go:2184` — already changed in Step 2.
  - `fragment/fragment.go:317,966,970,977,994,996,1003,1004,1010` — replace `ldf.Lineno` with `ldf.LinenoLine()` (or `ldf.GetLineno().Line`).
  - `compiler/ivy_compile.go:1444` — `prev.Lineno` in the error format string → `prev.LinenoLine()`.
  - `vmt/vmt.go:33,556` — `lf.Lineno` → `lf.LinenoLine()`.
- All write sites in tests that set the int directly (`module/module_test.go:242`, `check/regression_test.go:99,383,385,451,453,474,476`, `check/check_test.go:108,494,506,508`, `check/isolate_check.go:78` — already done in Step 2):
  - Replace `lf.Lineno = N` with `lf.SetLineno(ast.Location{Line: N})`. (Tests don't usually need a filename — an empty filename + nonzero line still satisfies the `loc.Filename != "" || loc.Line > 0` guard in `PrettyLineno`.)
- `parser/grammar_v17.y:79` — `res.Lineno = lf.Lineno` line in `addLabel`: delete (the subsequent `if lf.HasLocSet() { res.SetLineno(lf.GetLineno()) }` becomes unconditional after the audit, since every produced LF now has Loc set).
- After deleting the field, any forgotten read or write becomes a compile error — the type checker becomes the regression test for this fix, which is the value of doing the cleanup.

This step is what makes the solution truly "general": the impossible state ("`Lineno` set but `Base.Loc` not set") becomes unrepresentable.

### Step 5 — Verify

Run the failing test and confirm the "(internal)" lines are replaced with `<IVY_EXAMPLES>/...` paths matching Python:

```
cd /Users/jaten/ivy/goivy && go test -tags xtracer -run TestOrdLive ./parser/ 2>&1 | tee log.red
grep -A1 'The following definitions are used' log.red | head -40
```

Expected: `~go` lines for definitions match the `~py` lines exactly (modulo XTRACE indices). The "(internal)" placeholder should not appear for any line preceded by a `definition` keyword in the source `.ivy` file.

Run the broader test groups to confirm no regression:

```
cd /Users/jaten/ivy/goivy && go test ./ast/... ./parser/... ./compiler/... ./check/... ./fragment/... ./vmt/...
cd /Users/jaten/ivy/goivy && go test -tags xtracer -run 'TestOrd' ./parser/ 2>&1 | tail -50
cd /Users/jaten/ivy/goivy && go test -tags xtracer -run 'TestApple' ./parser/ 2>&1 | tail -50
```

If after Step 4 a previously-passing test now fails because it manually set `lf.Lineno = 0` to reset the line, the test was depending on the int/Loc divergence as a reset signal. Update such tests to use a fresh `LabeledFormula` instead.

## Critical files

- `/Users/jaten/ivy/goivy/log.red` — current divergence log (read-only reference)
- `/Users/jaten/ivy/goivy/parser/grammar_v17.y` — DEFINITION rule (line 980), `addLabel` (line 73), `mkLF` (line 86), proof rule (line 1027), schdecl rule (line 2690). Source-of-truth for the generated parser.
- `/Users/jaten/ivy/goivy/parser/grammar_v17.go` — goyacc-generated; regenerate via `go generate ./parser/`. Do not hand-edit.
- `/Users/jaten/ivy/goivy/parser/generate.go` — `//go:generate goyacc -o grammar_v17.go -p v17 grammar_v17.y`
- `/Users/jaten/ivy/goivy/ast/decl_ast.go` — `LabeledFormula` struct (line 20), constructor (line 45), `NewLabeledFormulaFrom` (line 60), `cloneInternal` (line 97), `Lineno int` field to remove (line 25)
- `/Users/jaten/ivy/goivy/ast/ast.go` — `Base` (line 191), `Location` (line 18), `GetLineno`/`SetLineno`/`HasLocSet` (lines 287-289)
- `/Users/jaten/ivy/goivy/check/helpers.go` — `PrettyLineno` (line 39), `PrettyLF` (line 51) — the consumers; no edits needed.
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_check.py` — Python source of truth: `pretty_label`/`pretty_lineno`/`pretty_lf` (lines 256-263), definitions print (lines 551-554)
- `/Users/jaten/ivy/pyivy/ivy/ivy/ivy_parser.py` — Python parser: `p_top_definition_optlabel_gdefn_optproof` (line 1681), `get_lineno` (line 127)
- `/Users/jaten/ivy/goivy/compiler/decl.go` — three `clf.Lineno = ...` sites (lines 1310, 1341, 1507)
- `/Users/jaten/ivy/goivy/compiler/phase6.go` — `newProp.Lineno = prop.Lineno` (line 2279), `if sg.Lineno > 0` (line 2184)
- `/Users/jaten/ivy/goivy/compiler/ivy_compile.go` — `result.Lineno = prop.Lineno` (line 1884), `prev.Lineno` in error message (line 1444)
- `/Users/jaten/ivy/goivy/check/isolate_check.go` — `subgoal.Lineno = ...` (line 78)
- `/Users/jaten/ivy/goivy/fragment/fragment.go` — read sites at lines 317, 966, 970, 977, 994, 996, 1003, 1004, 1010
- `/Users/jaten/ivy/goivy/vmt/vmt.go` — read sites at lines 33, 556
- `/Users/jaten/ivy/goivy/check/regression_test.go`, `check/check_test.go`, `module/module_test.go` — test sites that write `lf.Lineno = N`

## Verification

1. Build the parser package after regenerating goyacc output:
   ```
   cd /Users/jaten/ivy/goivy && go generate ./parser/ && go build ./...
   ```
2. Run the focused conformance test and inspect the definitions block:
   ```
   cd /Users/jaten/ivy/goivy && go test -tags xtracer -run TestOrdLive ./parser/ 2>&1 | tee log.red
   ```
   Expected: the `~go[after i=...]:    The following definitions are used:` block now contains `<IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy: line N: ref.def…` lines that match the corresponding `~py[…]` lines exactly. No `(internal)` should remain in the definitions block.
3. Run the broader test groups to confirm no regression:
   ```
   cd /Users/jaten/ivy/goivy && go test ./ast/... ./parser/... ./compiler/... ./check/... ./fragment/... ./vmt/... ./module/...
   cd /Users/jaten/ivy/goivy && go test -tags xtracer -run 'TestOrd' ./parser/ 2>&1 | tail -50
   cd /Users/jaten/ivy/goivy && go test -tags xtracer -run 'TestApple' ./parser/ 2>&1 | tail -50
   ```
4. Sanity-check the field removal (only after Step 4): a `grep -n '\.Lineno\b' ast compiler check fragment vmt parser` should yield zero results other than `LinenoLine()` calls and the `Lineno int` field on `DefineInfo` (a different struct, `ast/decl_ast.go:2102`, that stores parameter line numbers — leave it alone unless it shows up in a divergence).
