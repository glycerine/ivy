# Fix: Reorder check_ranking.go to match Python execution order (neg_prop_init before desugar)
Created: 2026-05-21 UTC

## Status of previously applied changes

These changes from the prior session are already applied:
- `ivy_ranking.py:129` — list comprehension → for loop with xtracer.trace()
- `ivy_l2s.py:214-219` — added compiled=nil trace
- `check_ranking_tactic.go` — added rankingTrigGlob, strInvarMap build, prem-update, T1 usage throughout

## Current divergence: trace line 229278

- Go emits:  `ast.LF.clone PRESERVE origid=848 counter=1693`
- Python emits: `ast.LF.__init__ id=1693 counter=1694`

After Python creates LF id=1692 (the `l2s_consts_d` invariant from `rankingInvariants()`),
Python creates NEW LFs (`l2s_init_glob_*`, `neg_prop_init`, then `l2s_when_*`) BEFORE
running the desugar loop. Go starts the desugar loop immediately after `rankingInvariants()`
returns, skipping these creation steps.

## Root cause

**Python execution order** (`ivy_ranking.py` lines 431–515):
1. `l2s_consts_d` (line 431, inside `rankingInvariants()`)
2. `convert_to_init` + `iinvs` + `neg_prop_init` (lines 433–457) — **MISSING in Go**
3. `winvs` / `l2s_when_*` (lines 461–471) — exists in Go but AFTER desugar (wrong)
4. Print (lines 474–482)
5. Desugar (line 515) — exists in Go but BEFORE winvs (wrong)
6. Add to model (line 521+)

**Current Go execution order** (`check_ranking.go` lines 299–410):
1. `rankingInvariants()` call (includes `l2s_consts_d`)
2. Desugar loop (lines 306–328) ← wrong position
3. `winvs` / `l2s_when_*` (lines 330–384) ← wrong position, and neg_prop_init missing before it
4. Print (lines 386–399)
5. Add to model (line 402)

## Fix: `check_ranking.go`

Between line 304 (end of `rankingInvariants()` block) and line 306 (start of desugar block),
insert the `neg_prop_init` + `l2s_init_glob_*` step. Then reorder the existing blocks so
`winvs` comes before desugar.

### New block to insert (Python lines 433–457)

```go
// Python ivy_ranking.py:433-457: convert_to_init + iinvs + neg_prop_init
{
    knownInits := make(map[string]bool)
    var iinvs []Expr
    var localCTI func(f Expr) Expr
    localCTI = func(f Expr) Expr {
        switch n := f.(type) {
        case *LogicAnd:
            terms := make([]Expr, len(n.Terms))
            for i, t := range n.Terms { terms[i] = localCTI(t) }
            return &LogicAnd{Terms: terms}
        case *LogicOr:
            terms := make([]Expr, len(n.Terms))
            for i, t := range n.Terms { terms[i] = localCTI(t) }
            return &LogicOr{Terms: terms}
        case *LogicNot:
            return &LogicNot{Body: localCTI(n.Body)}
        case *LogicImplies:
            return &LogicImplies{T1: localCTI(n.T1), T2: localCTI(n.T2)}
        case *LogicIff:
            return &LogicIff{T1: localCTI(n.T1), T2: localCTI(n.T2)}
        case *ForAll:
            return &ForAll{Variables: n.Variables, Body: localCTI(n.Body)}
        case *LogicExists:
            return &LogicExists{Variables: n.Variables, Body: localCTI(n.Body)}
        default:
            vs := collectVarsSlice(f)
            ini := applyNB(l2sInit(vs, f, proofLabel), checkVarsToNodes(vs)...)
            key := string(f.Sexp())
            if _, ok := f.(*LogicGlobally); ok && !knownInits[key] {
                iinvs = append(iinvs, &LogicImplies{T1: ini, T2: f})
                knownInits[key] = true
            }
            if _, ok := f.(*LogicEventually); ok && !knownInits[key] {
                iinvs = append(iinvs, &LogicImplies{T1: f, T2: ini})
                knownInits[key] = true
            }
            return ini
        }
    }
    negPropInit := &LogicNot{Body: localCTI(fmla)}
    acfg := cfg.Mod.Cfg.AstCfg
    for i, iinv := range iinvs {
        invars = appendLF(acfg, invars, fmt.Sprintf("l2s_init_glob_%d", i), iinv, lineno)
    }
    invars = appendLF(acfg, invars, "neg_prop_init", negPropInit, lineno)
}
```

### Reorder: move winvs block BEFORE desugar

The existing `winvs` block (Python lines 461–471) currently lives at Go lines 330–384, AFTER
the desugar loop. Move it to immediately after the new `neg_prop_init` block above.

### Move desugar block AFTER winvs

The desugar block (Python line 515) currently lives at Go lines 306–328, BEFORE winvs.
Move it to after the winvs block (and after the print block).

Also remove the now-redundant `l2sSaved := L2SSaved()` / `_ = l2sSaved` lines
(the `Desugar()` function creates its own `l2sSaved` internally at check_l2s.go:1082).

### Resulting Go order (matches Python):

```
rankingInvariants() call                    ← unchanged
neg_prop_init + l2s_init_glob_* block       ← NEW (Python lines 433-457)
winvs / l2s_when_* block                    ← MOVED from after desugar
print block                                 ← unchanged position
desugar loop                                ← MOVED from before winvs
model.Invars = append(...)                  ← unchanged
```

## Critical files

- `~/ivy/go/src/github.com/glycerine/ivy/goivy/check_ranking.go` lines 299–410
  - Functions used (all in same package):
    - `appendLF` (check_l2s_auto.go:1026) — creates labeled formula + appends
    - `collectVarsSlice` (check_l2s_auto.go:1097) — collects free vars
    - `checkVarsToNodes` (check_l2s.go:107) — converts []*LogicVariable → []Expr
    - `applyNB` (check_l2s.go:93) — applies named binder to args
    - `l2sInit` (check_l2s.go:77) — creates l2s_init named binder
    - `Desugar` (check_l2s.go:1081) — desugars $was/$happened

## Verification

```
cd ~/ivy/goivy && make golden-all
```

Expect the divergence at trace 229278 to disappear. Test should either pass fully or
diverge at a higher trace number.
