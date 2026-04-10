# Plan: Share Z3 Translation Cache Across Solver Instances Within a Module Session

Created: 2026-04-10 20:30 UTC

## Context

`make golden` diverges at `log.red:247689`:

```
247688  go : XTRACE: ivy_solver.py:263 uninterpretedsort() ENTER name=tar_cf_clock
        py : XTRACE: ivy_solver.py:263 uninterpretedsort() ENTER name=tar_cf_clock

247689  go : XTRACE: ivy_solver.py:263 uninterpretedsort top HASH canon= sorts=[]
        py : XTRACE: ivy_solver.py:263 uninterpretedsort top HASH canon= sorts=[addr_type, index, lclock, loc_type, mem_loc_type, mem_type, op_ack_type, op_type, proc, tar_cf_clock, tar_clock]
```

Python has 11 sorts cached in `z3_sorts` from prior solver calls within the current Module context. Go's cache is empty because every `z3bridge.NewSolver()` constructs a fresh `Translator` with a brand-new Z3Context and empty maps. The two systems are diverging because of an architectural mismatch:

- **Python (`pyivy/ivy/ivy/ivy_solver.py:252-260`)**: `z3_sorts`, `z3_predicates`, `z3_constants`, `z3_functions` are module-level globals. They are reset by `clear()` (line 249-255) when `Module.__enter__()` runs (`pyivy/ivy/ivy/ivy_module.py:102`). All `z3.Solver()` calls within one Module's `with` block share the same caches AND the same Z3 main context.

- **Go (`goivy/z3bridge/translate.go:40-61`)**: `sorts`, `sortsInv`, `consts`, `z3_functions`, `z3_predicates`, AND `Ctx` (a `*Z3Context`) are all per-`Translator`. Every `Solver.NewTranslator()` (`solver.go:52`) creates a fresh context AND fresh maps. Translators NEVER share state.

This causes two correctness problems beyond the trace divergence:

1. The same Ivy uninterpreted sort gets a NEW Z3 sort object on each solver call. Z3 distinguishes sorts by identity, not name. Across solver calls, formulas that "should" reference the same sort actually reference different sort objects. This can leak into model extraction (`SortFromZ3` reverse lookup) and into how Z3 interprets repeated declarations.

2. Each `NewSolver()` re-runs `z3.DeclareSort` for every sort it sees, paying full re-declaration cost on every check. Python pays it once per Module context.

The stack trace at `goivy/stack_trace_247689.txt` confirms the divergence point:
```
z3bridge.(*Translator).TranslateSort         translate.go:274
z3bridge.(*Translator).translateVarOrConst   translate.go:1034
z3bridge.(*Solver).ClausesToZ3                solver.go:268
z3bridge.(*Solver).GetSmallModelWithCond     solver_model.go:140
actions.SmallModelClauses                      phase4.go:304   ← module IS in scope here
check.checkFcsNormalPath.func1                check.go:636
check.CheckSafetyInStateWithAG                isolate_check.go:1344
check.CheckIsolate                             isolate_check.go:485
```

`actions.SmallModelClauses(cls, fc, diag, mod *module.Module)` already receives the module — it just unpacks `mod.Sig` and `mod.Cfg.SolverOpts` and discards the rest. We need to thread the `*module.Module` deeper so the Solver can find a module-scoped cache.

## Approach (high level)

Mirror Python's "one cache per Module context" by introducing a `Z3SessionCache` that lives on `module.Module` and is shared by every Translator created during that module's lifetime. **Replace** `NewSolver(sig, opts)` with `NewSolver(mod *module.Module, opts *module.SolverOptions)` — single API, no parallel constructors. Sig is derived from `mod.Sig` (or `il.NewSig()` if mod is nil for tests). All 24 production call sites and 150+ test call sites are updated in one pass.

### Why a per-Module field, not a package global

CLAUDE.md section C forbids package-level mutable variables: "We run thread pools of ivy models on multi-core machines. Package-level vars are shared across goroutines and break multi-tenancy." A per-`module.Module` cache satisfies this — concurrent checks of different modules use isolated caches.

### Why the Z3Context must move with the cache

A cached `Z3 sort` is bound to a specific `Z3Context`. Sharing `t.sorts` across Translators is meaningless unless they also share the `Z3Context`. So `Z3SessionCache` owns BOTH the maps AND the `*Z3Context`.

### Why we cannot put a `*z3bridge.Z3SessionCache` field directly on `module.Module`

`z3bridge` already imports `module` (`z3bridge/solver.go:20`). The reverse import would be a cycle. Solution: store as `any` on `module.Module`, OR define a tiny `Z3CacheIface` interface in `module/` and have `z3bridge.Z3SessionCache` implement it.

We choose `any` because the cache only needs to be opaque to the `module` package — only z3bridge ever reads its content.

## Architecture

### New file: `goivy/z3bridge/session_cache.go`

```go
package z3bridge

import (
    "sync"
    lg "github.com/glycerine/ivy/goivy/logic"
)

// Z3SessionCache is the Go equivalent of Python's z3_sorts / z3_predicates /
// z3_constants / z3_functions module-level globals (ivy_solver.py:252-260).
// One Z3SessionCache lives per module.Module — shared by every Translator
// created via NewTranslatorWithCache. Cleared by Clear() when entering a
// new module context (Python: ivy_solver.clear() in Module.__enter__).
//
// Multi-tenancy: each module.Module has its own cache, so concurrent checks
// of different modules are isolated. Within one module session, all solver
// calls share Ctx and the maps below — matching Python's main_ctx singleton
// + module-level dicts.
type Z3SessionCache struct {
    mu sync.Mutex

    // Ctx is the shared Z3 context. All cached Z3 sorts/exprs are tied
    // to this context. Solvers that share this cache share this context.
    Ctx *Z3Context

    // Translation caches — same names/semantics as the old Translator
    // fields they replace.
    sorts         map[lg.NodeKey]Sort
    sortsInv      map[uint]lg.Sort
    consts        map[lg.NodeKey]Expr
    z3_functions  map[lg.NodeKey]FuncDecl
    z3_predicates map[lg.NodeKey]func(args ...Expr) Expr
}

// NewZ3SessionCache builds a fresh cache with its own Z3 context and empty
// maps. Mirrors Python's clear() at startup.
func NewZ3SessionCache() *Z3SessionCache {
    c := &Z3SessionCache{
        Ctx: NewZ3Context(),
    }
    c.resetMaps()
    return c
}

// Clear resets the cache to its initial state (empty maps, but the same
// Z3Context). Mirrors Python ivy_solver.clear() called from
// Module.__enter__ (ivy_module.py:102).
func (c *Z3SessionCache) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.resetMaps()
}

func (c *Z3SessionCache) resetMaps() {
    c.sorts         = make(map[lg.NodeKey]Sort)
    c.sortsInv      = make(map[uint]lg.Sort)
    c.consts        = make(map[lg.NodeKey]Expr)
    c.z3_functions  = make(map[lg.NodeKey]FuncDecl)
    c.z3_predicates = make(map[lg.NodeKey]func(args ...Expr) Expr)
}
```

### Modify: `goivy/z3bridge/translate.go`

Replace the per-`Translator` cache fields with a `*Z3SessionCache` reference. Translator becomes a thin wrapper that delegates all cache reads/writes to `t.cache.<field>`. The `Ctx` field becomes an alias for `t.cache.Ctx` (kept as a struct field for code-search compatibility).

```go
type Translator struct {
    s     *Solver
    cache *Z3SessionCache  // shared (when via Module) or owned (legacy nil-mod path)
    Ctx   *Z3Context       // == cache.Ctx, kept for callers reading t.Ctx directly

    TranslateMerkle iu.MerkleState
    translateDepth  int
}
```

Replace `s.NewTranslator()` with two constructors:

```go
// NewTranslator creates a translator that owns its own per-Solver cache.
// Used by tests and ad-hoc Solver instances that don't have a module.
// Behavior is identical to today: fresh Z3Context, empty maps.
func (s *Solver) NewTranslator() *Translator {
    cache := NewZ3SessionCache()
    return s.newTranslatorFromCache(cache)
}

// NewTranslatorWithCache creates a translator that shares the given
// session cache (and therefore the cache's Z3Context). Multiple
// Translators built this way will see each other's cached sorts /
// predicates / constants / functions.
func (s *Solver) NewTranslatorWithCache(cache *Z3SessionCache) *Translator {
    return s.newTranslatorFromCache(cache)
}

func (s *Solver) newTranslatorFromCache(cache *Z3SessionCache) *Translator {
    t := &Translator{
        s:     s,
        cache: cache,
        Ctx:   cache.Ctx,
    }
    t.initEqPred()
    return t
}
```

Update all 5+ accesses to `t.sorts`, `t.sortsInv`, `t.consts`, `t.z3_functions`, `t.z3_predicates` in `translate.go` (and other z3bridge files — see "Files affected" below) to use `t.cache.sorts` etc. The cache mutex must be held around any read-modify-write. (Use `c.mu.Lock(); defer c.mu.Unlock()` inside `TranslateSort`'s "miss" branch — see the patch sketch in the verification section.)

Update `Translator.Clear()` and `Translator.dumpSortsCanon()` to delegate to the cache.

Note: `t.initEqPred()` writes to `t.cache.z3_predicates`. With the cache shared, `initEqPred()` must be idempotent — it already is, since writing the same key with the same closure is safe; but to be defensive, check for the key first or move the eq-pred init into `NewZ3SessionCache()`. (Recommended: move it to the cache constructor and have Translator skip it.)

### Modify: `goivy/z3bridge/solver.go`

**Replace** the existing `NewSolver(sig, opts)` signature with `NewSolver(mod, opts)`:

```go
type Solver struct {
    mu               sync.Mutex
    tr               *Translator
    z3u              *Z3Utils
    opts             *module.SolverOptions
    sig              *il.Sig
    mod              *module.Module  // NEW: module this solver belongs to (nil = ad-hoc/test)
    HandleRangeSorts bool
}

// NewSolver creates a Solver bound to the given module. The module's
// Z3SessionCache is reused (and lazily created) so that all NewSolver
// calls against the same module share Z3 sorts, constants, functions,
// and predicates — matching Python's "z3.Solver() instances within a
// Module context share z3_sorts/etc." semantics.
//
// Pass nil for mod to get an ad-hoc per-Solver cache (tests, low-level
// usage that has no module). Pass nil for opts to get default options.
func NewSolver(mod *module.Module, opts *module.SolverOptions) *Solver {
    if opts == nil {
        opts = module.DefaultSolverOptions()
    }
    var sig *il.Sig
    if mod != nil {
        sig = mod.Sig
    }
    if sig == nil {
        sig = il.NewSig()
    }
    s := &Solver{
        z3u:              NewZ3Utils(),
        opts:             opts,
        sig:              sig,
        mod:              mod,
        HandleRangeSorts: true,
    }
    cache := getOrCreateModuleCache(mod)
    s.tr = s.NewTranslatorWithCache(cache)
    return s
}

// getOrCreateModuleCache returns mod's Z3SessionCache, lazily creating it on
// first access. For mod==nil, returns a fresh per-Solver cache (no sharing —
// matches today's behavior for tests). Stored as `any` on module.Module
// to avoid the module → z3bridge import cycle.
func getOrCreateModuleCache(mod *module.Module) *Z3SessionCache {
    if mod == nil {
        return NewZ3SessionCache()
    }
    if existing := mod.GetZ3SessionCache(); existing != nil {
        return existing.(*Z3SessionCache)
    }
    cache := NewZ3SessionCache()
    mod.SetZ3SessionCache(cache)
    return cache
}
```

**Note**: changing the first parameter type from `*il.Sig` to `*module.Module` is intentionally a compile-time-breaking change. Every caller MUST be updated. There is no parallel `NewSolverForModule` API — one entry point only.

### Modify: `goivy/module/module.go`

Add a hidden field on `Module` and accessor methods. The field is typed `any` to avoid the cycle.

```go
type Module struct {
    // ... existing fields ...

    // z3SessionCache holds the Go equivalent of Python's per-module-context
    // z3_sorts/z3_predicates/z3_constants/z3_functions globals. Stored as
    // `any` because the concrete type lives in z3bridge (which imports
    // module). Accessed only via GetZ3SessionCache / SetZ3SessionCache.
    // Cleared by Module.Enter() (mirroring Python's clear() in __enter__).
    z3SessionCache any
}

// GetZ3SessionCache returns the opaque z3bridge cache attached to this module,
// or nil. Type-assert to *z3bridge.Z3SessionCache in z3bridge code.
func (m *Module) GetZ3SessionCache() any {
    return m.z3SessionCache
}

// SetZ3SessionCache attaches a z3bridge cache to this module. Called by
// z3bridge.NewSolverForModule on first access.
func (m *Module) SetZ3SessionCache(c any) {
    m.z3SessionCache = c
}
```

### Modify: `goivy/module/context.go`

`Module.Enter()` already calls `cfg.SolverClearFn()`. We replace that hook with a direct call to the cache's Clear method via an interface, so the existing wiring stays minimal:

```go
// Z3CacheClearer is implemented by *z3bridge.Z3SessionCache.
type Z3CacheClearer interface {
    Clear()
}

func (m *Module) Enter() {
    cfg := m.Cfg
    if cfg == nil {
        panic("module.Enter: Cfg is nil")
    }
    m.prevModule = cfg.CurrentModule
    if m.prevModule != nil {
        m.oldSig = m.prevModule.Sig
    }
    cfg.CurrentModule = m
    // Python: ivy_solver.clear() — clear cached Z3 values when entering.
    if c, ok := m.z3SessionCache.(Z3CacheClearer); ok {
        c.Clear()
    }
    // Keep SolverClearFn for backward compat (no current setters, but harmless).
    if cfg.SolverClearFn != nil {
        cfg.SolverClearFn()
    }
}
```

Note: production never calls `Module.Enter()` today (only tests do). So this hook is mostly defensive — the cache is created lazily, and a single check session never re-enters the module. If we later wire `Enter()` into the production check entry path, the hook starts firing.

## Migration of call sites

The signature change is `NewSolver(sig *il.Sig, opts) → NewSolver(mod *module.Module, opts)`. Every caller must be updated. The compiler will reject any miss.

### Category A — call sites with `mod *module.Module` already in scope (direct swap):

| File:line | New code |
|---|---|
| `actions/phase4.go:303` (`SmallModelClauses`) | `z3bridge.NewSolver(m, sopts)` (m already in scope) |
| `interp/phase4.go:89` | `z3bridge.NewSolver(state.Domain, nil)` |
| `interp/phase4.go:144` | `z3bridge.NewSolver(state.Domain, nil)` |
| `interp/helpers.go:290` | `z3bridge.NewSolver(state.Domain, nil)` |
| `tactics/tactics.go:149,213,280` | `z3bridge.NewSolver(tc.mod(), tc.modSolverOpts())` (add `mod()` accessor — currently has `modSig()`) |
| `vmt/vmt.go:624` | `z3bridge.NewSolver(m, m.Cfg.SolverOpts)` |
| `alpha/alpha.go:238,767` | `z3bridge.NewSolver(m, nil)` |

### Category B — call sites that need a new `mod` parameter threaded in:

| File:line | Current | What to add |
|---|---|---|
| `actions/phase4.go:110` (`ClausesImplyFormulaCex`) | `NewSolver(nil, nil)` | Add `mod *module.Module` parameter; update all callers |
| `actions/phase4.go:146` (`Implies`) | `NewSolver(nil, nil)` | Same |
| `actions/phase4.go:242` (`ExtractPrePostModel`) | `NewSolver(nil, nil)` | Same |
| `actions/transrel.go:1279,1521,1555,1576,2242,2266` | `NewSolver(nil, nil)` (×6) | Each function (`ReverseInterpolantCase`, `InterpolantCase`, etc.) gets a `mod *module.Module` parameter. Chase callers up the stack until they hit a frame with mod in scope. The check loop has it. |
| `art/art.go:537` | `NewSolver(nil, nil)` | `*ART` likely has a `Mod *Module` field — use it. If not, add one. |
| `interp/helpers.go:149,207,231,459` | `NewSolver(nil, nil)` | These helpers receive interp state or conjs; thread `state.Domain` (or add a `mod` parameter). |
| `iupdr/iupdr.go:135,517` | `NewSolver(nil, nil)` | iupdr's `u` struct should hold `Mod *Module`; thread from constructor. |
| `webui/widget_analysis_session.go:1289,1442` | `NewSolver(nil, nil)` | Session has mod context — thread through. |
| `webui/concept_isession.go:348` | `NewSolver(nil, nil)` | Same. |
| `webui/concept_alpha.go:31` | `NewSolver(nil, nil)` | Same. |
| `trace/trace.go:607` | `NewSolver(nil, nil)` | trace package — `tb.LastAction` may give a path to mod; if not, accept `nil` mod (degraded). |
| `end2end/verify_test.go:144` | `NewSolver(nil, nil)` | Test — keep `nil` for mod. |

### Category C — test sites (≈150+ files):

Mechanical edit: replace `z3bridge.NewSolver(nil, nil)` with `z3bridge.NewSolver(nil, nil)` (still passes nil; only the type of the first arg differs, not the value). Most tests already pass `nil` for sig, so the source text stays the same. Tests that pass an explicit `*il.Sig` (e.g., `z3bridge.NewSolver(mySig, nil)`) need to be rewritten:

- If they need a sig but no module: construct a throwaway module via `m := module.NewWithSig(mySig)` and pass `m`.
- If they need a real module session: same.

Run `grep -rn "z3bridge.NewSolver(" --include='*.go' goivy/` after the signature change and fix each compile error. The compiler will catch them all.

### Migration ordering (one big PR, but staged within it):

1. **Land the Z3SessionCache infrastructure** — `session_cache.go`, Translator delegation, module field/accessors, `Module.Enter()` clear hook. **Do not change `NewSolver` yet.** All tests must pass. This is a no-op refactor: the per-Translator behavior is preserved by giving each Translator a freshly-allocated `Z3SessionCache` (1 cache per Translator, equivalent to today).

2. **Change the `NewSolver` signature** in one commit, fix every call site (Category A, B, C). All tests compile and pass — the Category A/B sites now share caches per-module, the Category C tests still get per-Solver caches via `nil` mod.

3. **Run `make golden`**. Expected outcome: the divergence at line 247689 either disappears or shifts substantially (the cache is now populated by all the prior solver calls in this module session). Iterate on remaining downstream divergences as before.

4. **Run the regression sweep** on every package.

## Files affected

### New files
- `goivy/z3bridge/session_cache.go` — `Z3SessionCache` struct, `NewZ3SessionCache`, `Clear`.

### Modified files
- `goivy/z3bridge/translate.go` — Translator struct field swap, `NewTranslator`, `NewTranslatorWithCache`, `newTranslatorFromCache`, `Clear`, `initEqPred`, `dumpSortsCanon`, ALL accesses to the 5 cache maps inside `TranslateSort`, `translateVarOrConst`, `atomToZ3`, `applyZ3Func`, `Formula_to_z3_int` (anywhere `t.sorts` / `t.consts` / etc. appear).
- `goivy/z3bridge/solver.go` — Replace `NewSolver(sig, opts)` with `NewSolver(mod, opts)`. Add `getOrCreateModuleCache`. Add `mod *module.Module` field on the `Solver` struct. Update the comment block at lines 79-100 (which currently asserts "Go doesn't need clear" — that comment is the bug we're fixing).
- `goivy/z3bridge/solver_z3convert.go` — `NewTranslatorWithInterpolation` (line ~237) needs to take a cache parameter for the same reasons.
- `goivy/module/module.go` — Add `z3SessionCache any` field and `GetZ3SessionCache` / `SetZ3SessionCache`.
- `goivy/module/context.go` — Add `Z3CacheClearer` interface; call `c.Clear()` from `Module.Enter()`.
- All 24 production call sites listed in Category A/B (above) — update the first argument from `*il.Sig` to `*module.Module`. Add `mod` parameters to wrapper functions where needed.
- All ~150 test call sites — most need zero source change (already pass `nil` for the first arg). The few that pass an explicit `*il.Sig` need to construct a throwaway `module.NewWithSig(sig)`.

### Critical files to read before editing
- `goivy/z3bridge/translate.go:40-343` — Translator struct and TranslateSort. The cache delegation patch goes here.
- `goivy/z3bridge/translate.go:1020-1050` — `translateVarOrConst` (uses `t.consts`).
- `goivy/z3bridge/translate.go:600-700` — `atomToZ3` and `applyZ3Func` (use `t.z3_predicates`, `t.z3_functions`).
- `goivy/z3bridge/solver.go:28-107` — Solver struct and Clear comment.
- `goivy/z3bridge/solver_z3convert.go:200-250` — `NewTranslatorWithInterpolation` variant.
- `goivy/actions/phase4.go:280-310` — SmallModelClauses (the divergence's call site).
- `goivy/module/module.go:18-148` — Module struct.
- `goivy/module/context.go:25-40` — Module.Enter.

## Verification

### Unit tests

After landing the infrastructure (step 1, no signature change yet), run:
```
cd ~/ivy/goivy && go test ./z3bridge/... ./module/... ./actions/... ./interp/... ./check/...
```
All should pass — step 1 is a no-op refactor where each Translator still gets its own freshly-allocated `Z3SessionCache` (1:1 with today's per-Translator behavior).

After step 2 (signature change + call-site migration), re-run the same package set. The Category A/B sites now share caches per-module; Category C tests still pass `nil` mod and get per-Solver caches. Compile errors will guide migration.

### Targeted divergence test

After migrating ALL 24 production sites in one pass:
```
cd ~/ivy/goivy && make golden
```

Expected outcomes (in order of preference):
1. **Best case**: divergence at line 247689 disappears entirely; new divergence either appears much later or `make golden` passes.
2. **Acceptable**: line 247689 changes shape — Go now reports `sorts=[addr_type, index, ...]` matching Python.
3. **Bad case**: divergence still says `sorts=[]`. Means a call site was missed (most likely in test code that ran a real solver call, or in `trace/trace.go` where mod is hard to reach). Inspect the `log.red` lines just before 247689 for `uninterpretedsort EXIT` events to see which earlier solver calls populated Python's cache, then verify those Go call sites passed a non-nil mod.

### Cross-language sanity check

Add a one-shot test that:
1. Creates a `module.Module`.
2. Calls `NewSolver(mod, nil)` twice.
3. Calls `TranslateSort(myUninterpSort)` on the first translator, then `TranslateSort(myUninterpSort)` on the second.
4. Asserts both return the SAME Z3 sort (`zs1.GetId() == zs2.GetId()`).

Place under `z3bridge/session_cache_test.go`. This guards against future regressions.

### Multi-tenancy sanity check

Add a test that creates two `module.Module` instances and asserts they get DIFFERENT `Z3SessionCache` instances (so concurrent goroutines do not share state):
```go
m1, m2 := module.New(), module.New()
s1 := z3bridge.NewSolver(m1, nil)
s2 := z3bridge.NewSolver(m2, nil)
if s1.tr.cache == s2.tr.cache {
    t.Fatal("modules must not share cache")
}
```

### Concurrency test (optional but recommended)

Run two goroutines, each checking a different module via `NewSolver(mod, ...)`, with `-race`. Should not deadlock or race.

## Risks and rollback

- **Largest risk**: missing a `t.sorts` access that needs to become `t.cache.sorts`. The Go compiler catches all of these because the struct field is removed. Check that every `t.sorts`, `t.sortsInv`, `t.consts`, `t.z3_functions`, `t.z3_predicates`, `t.Clear` reference compiles after the swap.
- **Signature change blast radius**: changing `NewSolver(sig, opts)` to `NewSolver(mod, opts)` will produce ~24 production + ~150 test compile errors. This is intentional — fix each one. Most test sites change zero source bytes (they already pass `nil`).
- **Z3 context lifetime**: when a `Z3SessionCache` is garbage collected, its `Z3Context` is freed. As long as the cache is reachable via `module.Module.z3SessionCache`, it stays alive. After the module is gone, all derived Solver expressions become invalid — same as today.
- **`Translator.Close()` currently calls `t.Ctx.Close()`**. With shared context, closing one Translator must NOT close the cache's context (other Translators may still need it). Change `Translator.Close()` to be a no-op for cache-shared translators; only close the context if the Translator owns it. (Track ownership via a `ownsCache bool` field on Translator, or simply: never close — let GC handle disposal of the Z3Context when the cache is unreachable.)
- **Rollback**: if the migration introduces an unfixable regression, the cleanest revert is `git revert` of the signature-change commit. The Z3SessionCache infrastructure can stay (it's a no-op refactor in step 1) or be removed if desired.

## Out of scope

- Fixing the trace name. The user added the trace as `ivy_solver.py:263 uninterpretedsort top HASH canon= sorts=...` even though there's no actual hash here. We keep the trace text exactly as-is so the goivy/pyivy strings match.
- Adding `Module.Enter()` calls to production code paths. The cache is created lazily on first solver use; we don't need explicit Enter() yet.
- Forcing test sites to share caches across modules. Tests pass `nil` for mod and get per-Solver caches — that's correct for unit-test isolation.
