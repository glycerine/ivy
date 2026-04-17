# PLAN: Strongly deterministic Sexp/canon-keyed sort across goivy/pyivy L2S — fix `l2s_g_21` vs `l2s_g_20` divergence at xtrace 931827, plus all latent tie-breaking landmines

Created: 2026-04-16 (evening, ninth pass)

## Context

After the previous fix (NamedBinder dedup uses Sexp not String + struct-equality audit) the user re-ran `make tlb`. The `_old_l2s_g count=30 vs 34` divergence at xtrace 930871 is resolved. Trace now matches through SharedStep11 ENTER, all `namedBinders key=... count=N` lines, and all `SharedStep11 sub freshName=... binderKey=...` lines. The new divergence is at xtrace **931827** (~/ivy/goivy/log.tlb:73345-73346):

```
931824  ENTER type=Apply HASH canon=(Apply func:(NamedBinder name:l2s_g environ: vars:[(Variable name:V0 sort:(UninterpretedSort name:tlb.processor))] body:(Not body:(Apply func:(NamedBinder name:l2s_g ...   — both sides identical
...
931827  go : EXIT type=Apply app HASH canon=(Apply func:(Symbol name:l2s_g_21 ...) terms:[(Variable name:P sort:(UninterpretedSort name:tlb.processor))])
        py : EXIT type=Apply app HASH canon=(Apply func:(Symbol name:l2s_g_20 ...) terms:[(Variable name:P sort:(UninterpretedSort name:tlb.processor))])
```

Same input NamedBinder, same Sexp, but different fresh constant assigned: Go=`l2s_g_21`, Python=`l2s_g_20`. Off by one in the sorted `l2s_g` binder list.

## Root cause

The current sort in `collectAllNamedBinders` (Go) and the equivalent in `ivy_l2s.py:1422` use `String()`/`str` (PrettyFmla) as the sort key. `PrettyFmla` calls `dropAnnotations(False, set())` then `ugly(0)` which **strips variable sort annotations**. So two NamedBinders that are structurally distinct (different Variable sorts after `NormalizeNamedBinders` renames) but pretty-print identically produce **ties** in the sort key.

For ties:
- **Python `sorted(set(v), key=str)`** — `set(v)` iterates in PYTHONHASHSEED-dependent hash order; `sorted` is stable (Timsort), so tie order = hash order.
- **Go `sort.Slice(deduped, by .String())`** — introsort, **not stable**; tie order depends on quicksort partitioning of input.

Both are deterministic per-run but disagree across runtimes.

The earlier `SharedStep11 sub freshName=%s binderKey=%s` trace lines deceptively matched because `binderKey=str(b)` collapses sort-distinct binders to identical pretty form — the trace masks the underlying subs-map disagreement.

The same partial-key-sort pattern also exists at three other "landmine" sites that have not yet surfaced as a divergence on tlb.ivy but **will** when an example exercises them with sort-distinct l2s_g triples that share a body:
- Go `check/l2s_shared.go:94-96` (`sortL2sGTriples`) sorts by **Body.Canon() only** — leaves Vars and Environ as untracked tie components.
- Go `check/l2s_shared.go:337-339` (`SharedStep6_BuildTableau`'s inline sort of `toG`) — same partial key.
- Python `ivy_l2s.py:1051` (`to_g.sort(key=lambda x: x[1].canon())`) and `ivy_l2s.py:1202` (`sorted(l2s_gs, key=lambda x: x[1].canon())`) — partial key.

The user has directed: fix all of these now, eliminate every tie window. Use Sexp/canon as the **primary** (fully discriminating) sort key on both sides. No "secondary" tie-breaker — the canon is the sort key.

## Fix: switch every L2S sort over AST nodes/triples to a fully-discriminating canonical key

The canonical s-expression (`Sexp`/`canon`) is fully discriminating for both `NamedBinder` (Go's `b.Sexp()` and Python's `b.canon()` produce byte-identical output by construction — this is the foundation of the cross-language conformance test). For `(Vars, Body, Environ)` triples, we introduce/reuse a triple-canon helper that emits the SAME format on both sides.

### Triple canon format

Go already has a fully-discriminating triple key at `check/l2s.go:162-168`:

```go
func (t l2sGTriple) key() lg.NodeKey {
    env := "nil"
    if t.Environ != nil {
        env = *t.Environ
    }
    return lg.NodeKey("(l2sGTriple environ:" + env + " vars:" + lg.VarsSexp(t.Vars) + " body:" + string(t.Body.Sexp()) + ")")
}
```

Format: `(l2sGTriple environ:<env-or-nil> vars:[v1.sexp v2.sexp ...] body:<body.sexp>)`

We will mirror this in Python with a helper. The Python `_vars_sexp` (logic_sexp.py:8-13) already produces the matching `[v1.sexp v2.sexp ...]` form; `body.canon()` matches `Body.Sexp()`. So the helper is a one-liner:

```python
def _l2s_g_triple_canon(vs, t, env):
    env_str = env if env is not None else 'nil'
    return '(l2sGTriple environ:%s vars:%s body:%s)' % (env_str, _vars_sexp(vs), t.canon())
```

Place this in `ivy_l2s.py` as a module-level helper near the top (after imports). Import `_vars_sexp` from `logic_sexp` (or duplicate the 1-line vars sort+join inline if the import is awkward — both produce identical strings).

### Sites to change

| # | File:Line | Currently | After fix |
|---|---|---|---|
| 1 | check/l2s.go:1028-1030 | `sort.Slice(deduped, …) deduped[i].String() < deduped[j].String()` | sort by `string(deduped[i].Sexp())` |
| 2 | check/l2s_shared.go:94-96 (`sortL2sGTriples`) | sort by `Body.Canon()` only | sort by `triple.key()` (full triple canon) |
| 3 | check/l2s_shared.go:337-339 (`SharedStep6_BuildTableau` inline) | sort by `Body.Canon()` only | sort by `triple.key()` (full triple canon) |
| 4 | ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1422 | `sorted(set(v), key=str)` | `sorted(set(v), key=lambda b: b.canon())` |
| 5 | ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1051 | `to_g.sort(key=lambda x: x[1].canon())` | `to_g.sort(key=lambda x: _l2s_g_triple_canon(*x))` |
| 6 | ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1202 | `sorted(l2s_gs, key=lambda x: x[1].canon())` | `sorted(l2s_gs, key=lambda x: _l2s_g_triple_canon(*x))` |
| 7 | ~/ivy/pyivy/ivy/ivy/ivy_l2s.py (top of file, near imports) | (no helper) | add `_l2s_g_triple_canon(vs, t, env)` returning the matching format string |

All seven changes must land together. The Go and Python helpers must produce **byte-identical** strings for the same logical triple — this is what makes cross-language sort orders agree.

### Sites verified safe (no change needed)

Audit of every sort site touching AST/expr nodes in `check/` and `ivy_l2s.py`:

**Go check/ — already fully discriminating or not tie-prone:**
- `check/l2s.go:982-984` — sorts `*lg.Const` by `.Name`; names unique.
- `check/l2s_shared.go:81-84` (`sortNamedBinderMap`) — sorts by `Canon()` (full Sexp); fully discriminating.
- `check/ranking.go:254-256` — sorts `UninterpretedSort` by `String()` = name; names unique.
- `check/isolate_check.go:230, 242, 254` — sort by string fields (`Mixer`, `Name`); strings unique.
- `check/l2s_auto.go:153`, `check/l2s_shared.go:188, 451, 459, 467, 651, 656, 767, 957` — `sort.Strings`; strings unique.

**Python ivy_l2s.py — already fully discriminating or not tie-prone:**
- `:187` — `sorted(defn_deps.keys(), key=str)`; keys are symbols, strings unique in this dict.
- `:295`, `:1306`, `:1307` — sort over distinct strings; no AST.
- `:884`, `:1423` — sort dict keys (strings).
- `:1066`, `:1116`, `:1134`, `:1163`, `:1209` — sort by `lambda x: x.canon()` where x is a NamedBinder; full canon, fully discriminating.

## Why this is the right scope

The user explicitly directed: "Use Sexp as the primary sort key on both sides... this leaves a time-bomb [at l2s_shared.go:337]... Fix it now... Make it strongly deterministic on both sides."

This plan does exactly that:
1. The xtrace-931827 root cause (sites 1, 4) — **the immediate divergence**.
2. The l2s_shared.go:337 landmine + its sortL2sGTriples sibling (sites 2, 3) — **deferred no longer**.
3. The Python l2s_g triple sorts at lines 1051 and 1202 (sites 5, 6) — **kept in lockstep with the Go landmine fixes**.
4. The Python triple-canon helper (site 7) — **the byte-identity contract** between Python and Go for triple sorts.

After this lands, **every** L2S sort over AST nodes or triples uses a canon key that is byte-identical on both sides. No tie windows remain. Future ivy programs that exercise sort-distinct triples (currently latent, but real for any program with multiple gprops over differently-sorted variable lists with the same body shape) will not regress.

## Verification

**Do NOT run `make tlb` from this agent — the user runs it on another machine.**

Local checks this agent will run after the user approves:

1. `go build ./check/... ./logicutil/...` — confirm Go changes compile.
2. `go test ./check/... ./logicutil/...` — confirm existing tests still pass. If a test relies on the old String-based sort tie-order, it was relying on undefined behavior; investigate and update.
3. `go vet ./check/... ./logicutil/...` — confirm no vet warnings.
4. `python3 -c "from ivy.ivy import ivy_l2s"` — confirm Python module loads (catches syntax/import errors in the helper).

After the user re-runs `make tlb`, expected outcomes:

A. **Best case** — xtrace 931827 now matches. Trace continues into solver-input generation / theory context / SAT call. New divergence appears further downstream (or the run completes successfully).

B. **Possible case** — count or sub-trace lines now diverge at a different specific binder. This would mean the new canonical sort order placed the divergence at a different binder; the cause is **not** the sort tie-break (which is now deterministic and matching), but some other earlier issue we hadn't yet seen. Investigate as a separate divergence.

C. **Unexpected case** — divergence persists at 931827 with similar fresh-name index off-by-one. This would mean either:
   - Python virtualenv didn't reload the changed `ivy_l2s.py` (unlikely if installed in editable mode via `pip install -e`; verify with `python3 -c "import ivy.ivy.ivy_l2s; print(ivy.ivy.ivy_l2s.__file__)"`).
   - Python's `_l2s_g_triple_canon` and Go's `l2sGTriple.key()` produce different strings for some triple shape (would need to dump both and diff the first divergence).
   - A sort site we missed. Re-grep both codebases for `sorted(`, `sort.Slice`, `key=`, etc.

## Critical files

- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s.go (Site 1)
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go (Sites 2, 3)
- /Users/jaten/ivy/pyivy/ivy/ivy/ivy_l2s.py (Sites 4, 5, 6, 7)

## Reference

Python:
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1414-1424 — accumulate, dedup, sort named_binders
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1432-1436 — subs construction `subs[b] = lg.Const('{}_{}'.format(k,i), b.sort)`
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1048-1051 — to_g build, dedup (`dict.fromkeys`), sort
- ~/ivy/pyivy/ivy/ivy/ivy_l2s.py:1196-1209 — l2s_gs sort and iteration
- ~/ivy/pyivy/ivy/ivy/logic_sexp.py:8-13 — `_vars_sexp`
- ~/ivy/pyivy/ivy/ivy/logic_sexp.py:105-108 — `_named_binder_sexp` (lg.NamedBinder.canon)
- ~/ivy/pyivy/ivy/ivy/logic_sexp.py:265 — registration

Go:
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s.go:156-168 — `l2sGTriple.key()` (the existing fully-discriminating triple canon)
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s.go:988-1033 — `collectAllNamedBinders`
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go:73-98 — `sortNamedBinderMap`, `sortL2sGTriples`
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/check/l2s_shared.go:332-339 — `SharedStep6_BuildTableau` inline sort
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/sexp.go:141-148 — `VarsSexp`
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/sexp.go:162-168 — `NamedBinder.Sexp`
- /Users/jaten/go/src/github.com/glycerine/ivy/goivy/logic/canon.go:41 — `NamedBinder.Canon` returns Sexp

Trace evidence (from ~/ivy/goivy/log.tlb):
- :73334-73337 — xtrace 931824, both sides identical canon for the Apply at the divergence point
- :73345-73346 — xtrace 931827, divergence: `l2s_g_21` (Go) vs `l2s_g_20` (Py)
- :72463-72464, :72550-72551, :72615-72616, :72689-72690 — earlier matching `l2s_s_N`, `l2s_w_N`; only `l2s_g` exhibits the tie issue on tlb.ivy (consistent with the sort-distinct-collision hypothesis since l2s_g is the only key with multiple sort-collision-prone binders in this proof).
