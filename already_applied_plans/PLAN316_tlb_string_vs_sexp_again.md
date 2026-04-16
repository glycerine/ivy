# PLAN: Fix `_old_l2s_g count=30 vs 34` divergence at xtrace 930871 (NamedBinder dedup uses String, must use Sexp)

Created: 2026-04-16 (evening, seventh pass)

## Context

After the previous fix (compileAppNode NamedBinder-rep branch trace order at xtrace 775993), the user re-ran `make tlb` and the trace now matches all the way up to xtrace **930870** (`l2s.SharedStep11 ENTER nInvars=360 nAsms=304 nBindings=28 nPrems=0` — both sides agree). The new divergence is at xtrace **930871** (log.tlb:73388-73392):

```
930870  go : XTRACE: l2s.SharedStep11 ENTER nInvars=360 nAsms=304 nBindings=28 nPrems=0
        py : XTRACE: l2s.SharedStep11 ENTER nInvars=360 nAsms=304 nBindings=28 nPrems=0

930871  go : XTRACE: l2s.SharedStep11 namedBinders key=_old_l2s_g count=30
        py : XTRACE: l2s.SharedStep11 namedBinders key=_old_l2s_g count=34
```

Both sides agree on inputs (360 invars, 304 asms, 28 bindings, 0 prems). Go is missing **4** `_old_l2s_g` NamedBinders. `_old_l2s_g` sorts first alphabetically, so this is the first key emitted; subsequent keys (likely `l2s_g`, `l2s_init`, `l2s_w`, `l2s_when*`, etc.) almost certainly diverge by the same root cause but the comparator stops at the first mismatch.

## Root cause

Go dedups NamedBinders using `b.String()`. Python dedups using `set(v)` (struct equality).

- Go's `NamedBinder.String()` = `PrettyFmla(nb)` (logic/formula.go:523, logic/pretty.go:14-17) = `dropAnnotations(nb, false, {})` then `ugly(0)`. **Drops sort annotations from variables.**
- Python's `set()` uses `recstruct.__hash__` / `__eq__` (utils/recstruct_object.py:62-86), which compares `_tup = (name, variables_tuple, environ, body)`. `variables_tuple` element comparison uses `Var._tup = (name, sort)` (logic.py:117). **Preserves sort information.**

So two NamedBinders with identical `(name, environ, body shape, variable names)` but **different variable sorts** are:
- Python: distinct in `set()` — kept separate
- Go: collapse to one in String-keyed map — one is lost

### Why this manifests in tlb.ivy

`NormalizeNamedBinders` (logicutil/logic_utils.go:857; ivy_logic_utils.py:253-285) renames bound variables to `V0`, `V1`, `V2`, ... while **preserving** their original sorts. So multiple temporal-operator-derived `l2s_g` binders with the same body shape but different variable sorts (e.g., one over `tlb.thread`, one over `tlb.iface`) end up structurally distinct (different sorts) but pretty-print-identical (sorts dropped).

Each surviving `l2s_g` gprop produces one `_old_l2s_g` binder (3 calls in `prop_events`/`propEventsFunc` at ivy_l2s.py:1121,1125,1127 and l2s_shared.go:470,483,490 — all with identical name/vars/env/body, so they self-dedup to 1 per gprop). The `_old_l2s_g` count therefore equals the number of distinct gprops processed in `propEventsFunc` calls.

Python: 34 distinct gprops → 34 distinct `_old_l2s_g` binders.
Go: 4 sort-collisions get collapsed → 30 distinct.

## Sites to fix (all 7, coordinated)

The same `String()`-vs-`Sexp()` bug pattern appears at four cooperating sites in the L2S/substitution pipeline. All seven must change together so that the count, the substitution map, and the AST replacement all agree on the same structural identity.

| # | File:Line | Currently | After fix |
|---|---|---|---|
| 1 | check/l2s_shared.go:663 | `eventProps[prop.String()] = prop` | `eventProps[string(prop.Sexp())] = prop` |
| 2 | check/l2s_shared.go:666 | `eventWhens[when.String()] = when` | `eventWhens[string(when.Sexp())] = when` |
| 3 | check/l2s_shared.go:669 | `eventWaits[wait.String()] = wait` | `eventWaits[string(wait.Sexp())] = wait` |
| 4 | check/l2s.go:1008 | `key := b.String()` (collectAllNamedBinders dedup) | `key := string(b.Sexp())` |
| 5 | check/l2s_shared.go:782 | `subs[b.String()] = lg.NewConst(...)` | `subs[string(b.Sexp())] = lg.NewConst(...)` |
| 6 | logicutil/logic_utils.go:1134 | `key := nb.String()` (ReplaceNamedBindersAst, plain NamedBinder case) | `key := string(nb.Sexp())` |
| 7 | logicutil/logic_utils.go:1152 | `key := nb.String()` (ReplaceNamedBindersAst, Apply.Func case) | `key := string(nb.Sexp())` |

### Things explicitly NOT changed

- The `xtracer.Trace(...)` lines that emit `b.String()` for the trace text (e.g., l2s_shared.go:784 `binderKey=%s ... b.String()`, and the dedup-result sort comparator at l2s.go:1015 `deduped[i].String() < deduped[j].String()`). Python's traces use `str(b)` = pretty form; trace text format must continue to match. The bug is only in the **dedup/lookup keys**, not in the **trace/sort keys**.
- `cfg.Subs` (the inverse `map[string]string` at l2s_shared.go:775,783) is for trace-display only (renaming hook) — leave keyed by pretty form so trace output continues to match Python.

## Why all 7 fixes must change together

- **Sites 1-3** fix SharedStep7's per-statement event-set dedup. Without these, sort-distinct gprops in any single `instr_stmt`'s `eventProps` collapse → fewer iterations of `propEventsFunc` → fewer `_old_l2s_g` binders inserted into the AST.
- **Site 4** fixes SharedStep11's cross-statement collection dedup. Without this, even if all 34 binders are present in the AST after Sites 1-3, `collectAllNamedBinders` collapses 4 of them by String during dedup, giving a count of 30.

  Either bug alone is sufficient to explain count=30. Both need to be fixed because we cannot know without execution which one (or both) is actually triggering on tlb.ivy. If Sites 1-3 are the cause: Site 4 alone won't help (binders never made it into AST). If Site 4 is the cause: Sites 1-3 alone won't help (collected binders still collapse during dedup). Fixing all four guarantees the count goes to 34 regardless of which site is the trigger.

- **Sites 5-7** fix the substitution pipeline. After Sites 1-4 land, `namedBinders["_old_l2s_g"]` will hold 34 distinct binders. The loop at l2s_shared.go:780-786 generates fresh constants `_old_l2s_g_0` ... `_old_l2s_g_33` and writes them to `subs[b.String()]`. If `subs` is still String-keyed, the 4 sort-collisions cause the later-written entries to overwrite the earlier ones, leaving only ~30 entries. Then `ReplaceNamedBindersAst` (also String-keyed lookup) replaces all 4 distinct binders with the same fresh constant. This is silently wrong and surfaces only as downstream proof divergences (different solver queries, different counterexamples). Fix together for correctness.

  Python ivy_l2s.py:1432-1436 keys `subs` by the NamedBinder OBJECT itself, and `replace_named_binders_ast` (ivy_logic_utils.py:1142+) looks up via `if ast in subs` — both struct-eq based. Sexp() is the Go equivalent of struct eq for these purposes.

## Why Sexp vs String

```
Go:    NamedBinder.String() = PrettyFmla(nb) = dropAnnotations(nb, false, {}).ugly(0)
       → drops variable sort annotations: "($l2s_g V0. body)" (no sort on V0)

Go:    NamedBinder.Sexp() = "(NamedBinder name:N environ:E vars:[V] body:B)"
       where VarsSexp(vars) emits each Variable as "(Variable name:X sort:(...))"
       → preserves all sort information
```

For dedup that mirrors Python's `set()` struct equality on NamedBinder, **Sexp** is the correct discriminator. **String** is correct only for trace text display.

## Audit (other String-keyed NamedBinder dedup sites — NOT in this plan)

| Site | What it dedups | Risk / Why deferred |
|---|---|---|
| check/l2s_shared.go:119, 123 (`cfg.L2sWhensSet[res.String()]`) | l2s_when binders in SharedStep1 | Could cause future divergence in `l2s_when*` count keys at SharedStep11. Defer because (a) the trace doesn't show those keys yet — they'd be reported AFTER `_old_l2s_g` if its count is fixed, and (b) `_old_l2s_g` count alone doesn't depend on this. Will surface as the next divergence if it's also buggy. |
| check/l2s.go:181 (`dedupeVarBodyPairs` keys by `fmt.Sprintf("%v:%v", Vars, Body)`) | NamedBindersConjs entries in SharedStep3 | Could cause divergence in `nEntries` traces at SharedStep3, but those traces match in the current log. Likely benign because SharedStep3 collects from invars (which use canonical sorts) before any normalize step that would create sort collisions. |
| check/ranking.go:319 (`seen[wo.String()]`) | WhenOperator dedup in ranking tactic | Tlb.ivy uses L2S not ranking — irrelevant to this trace. |

These sites have the same bug pattern but are NOT directly causing xtrace 930871's divergence. Per the user's iterative-debugging pattern, fix one divergence at a time and let subsequent ones surface.

## Verification

**Do NOT run `make tlb` from this agent — the user runs it on another machine.**

Local checks this agent will run after the user approves:

1. `go build ./check/... ./logicutil/...` — confirm changes compile.
2. `go test ./check/... ./logicutil/...` — confirm existing tests still pass. If a test was relying on the buggy String-keyed dedup, that test was wrong; investigate before changing it.
3. `go vet ./check/... ./logicutil/...` — confirm no new vet warnings.

After the user re-runs `make tlb`, expected outcomes:

A. **Best case**: xtrace 930871 now matches `_old_l2s_g count=34` on both sides. Trace continues into the next sorted key (likely `l2s_g`), then sub trace lines (`l2s.SharedStep11 sub freshName=... binderKey=...`), then `SharedStep11 EXIT`. New divergence appears further along the proof pipeline.

B. **Possible case**: count matches at 930871 but a subsequent count key (e.g., `l2s_when_first`, `l2s_whennext`) diverges with the same pattern. That would mean the L2sWhensSet bug (audit row 1) is also triggering — fix it the same way (use Sexp instead of String).

C. **Possible case**: counts all match but the sub trace ordering differs between Python and Go for ties. Sort tie-breaking is non-deterministic on both sides (Python `sorted(set(v), key=str)` ties depend on set's hash-based iteration order; Go `sort.Slice` ties depend on the random map iteration order before sort). If this becomes an issue, add a secondary sort key by Sexp to break ties deterministically — but only AFTER seeing it as a real divergence.

D. **Unexpected case**: count still off after the fix. Re-investigate; the sort-distinct hypothesis would be wrong. Likely candidates: (a) `NamedBindersAst` traversal misses something specific to one side, (b) the AST mutation itself differs (e.g., `prefix_action` vs Go's equivalent skips a binder somewhere).

## Critical files

- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go (Sites 1-3, 5)
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s.go (Site 4)
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logicutil/logic_utils.go (Sites 6, 7)

## Reference (source of truth)

Python:
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1115-1131 — `prop_events` (creates `_old_l2s_g` binders, 3 per gprop)
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1286-1318 — `instr_stmt` builds `event_props` as a `set()` (struct eq)
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1414-1424 — `named_binders_asts` collection + `set(v)` dedup
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1432-1436 — `subs` dict keyed by NamedBinder OBJECT (struct eq)
- ~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:1142+ — `replace_named_binders_ast` lookup via `if ast in subs`
- ~/ivy/pyivy/ivy/ivy/ivy_logic_utils.py:253-285 — `normalize_named_binders` renames vars to `V0`, `V1`, ... preserving sorts
- ~/ivy/pyivy/ivy/ivy/utils/recstruct_object.py:62-86 — recstruct `__eq__`/`__hash__` use tuple equality on `_tup`
- ~/ivy/pyivy/ivy/ivy/logic.py:117 — `Var = recstruct('Var', ['name', 'sort'], [])`
- ~/ivy/pyivy/ivy/ivy/logic.py:419 — `NamedBinder = recstruct('NamedBinder', ['name','variables','environ'], ['body'])`
- ~/ivy/pyivy/ivy/ivy/ivy_logic.py:1330-1331 — `lg.Var.ugly` drops sort if TopSort/SortVar (and `drop_annotations` makes it TopSort)
- ~/ivy/pyivy/ivy/ivy/ivy_logic.py:1455-1461 — `pretty_fmla` (str of NamedBinder) = `drop_annotations(False, set()).ugly(0)`

Go:
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/sexp.go:63-65 — `Variable.Sexp()` preserves sort
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/sexp.go:141-148 — `VarsSexp` preserves sorts
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/sexp.go:162-168 — `NamedBinder.Sexp()` includes VarsSexp
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/formula.go:523 — `NamedBinder.String()` = PrettyFmla
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/pretty.go:14-17 — PrettyFmla calls `dropAnnotations(false, {})` then `ugly(0)`
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logicutil/logic_utils.go:857 — `NormalizeNamedBinders` (mirrors Python)
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go:625-682 — `instrStmt` event_props loop
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s.go:982-1020 — `collectAllNamedBinders`
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go:741-797 — `SharedStep11_ReplaceNamedBinders`
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logicutil/logic_utils.go:1131-1160 — `ReplaceNamedBindersAst`
