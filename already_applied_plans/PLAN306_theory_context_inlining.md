# Plan: Move TheoryContext inside `if method != nil` block to match Python

Created: 2026-04-14 ~23:15 UTC

## Context

**Previous fix** (applied, partially correct): Moved `TheoryContext()` after `WithSorts.Enter()` to fix ordering. That fixed the `WithSorts.Enter` appearing in the right place (line 845767 now matches). But the next line 845768 diverges because Go still calls `TheoryContext()` unconditionally, while Python only calls it when `method is not None`.

**Current divergence**: `make golden` → TestOrdLive → line 845768:
```
Go : XTRACE: module/clauses.go:262 defToConstraint lhsSort=Boolean resultType=Iff
Py : XTRACE: check.CheckIsolate ENTER
```

Lines 845765-845767 all match (module.Copy ENTER/EXIT, WithSorts.Enter). The divergence is that Go emits `defToConstraint` traces (from `TheoryContext()`) before `CheckIsolate`, while Python goes directly to `check.CheckIsolate ENTER`.

## Root cause

**Python** (`ivy_check.py:820-846`) — `check_subgoals`:
```python
with mod:
    vocab = ivy_proof.goal_vocab(goal)
    with lg.WithSymbols(vocab.symbols):
        with lg.WithSorts(vocab.sorts):
            if method is not None:
                with im.module.theory_context():   # ← ONLY when method is not None
                    foo = method()
            else:
                check_isolate()                    # ← NO theory_context here
```

Python's `theory_context()` is **guarded by `if method is not None:`**. When method IS None, Python calls `check_isolate()` directly — which handles its own `theory_context()` internally (at `ivy_check.py:540`).

**Go** (`isolate_check.go:786,865`) — both branches of `CheckSubgoals`:
```go
cleanup := fakeMod.TheoryContext()   // ← UNCONDITIONAL — WRONG
if method != nil {
    // ... method branch ...
} else {
    CheckIsolate(fakeMod, nil)       // CheckIsolate also calls TheoryContext at line 108
}
```

Go calls `TheoryContext()` unconditionally before the `if method != nil` check. This produces `defToConstraint` traces where Python doesn't. Additionally, when method is nil, Go calls `TheoryContext()` TWICE (once here, once inside `CheckIsolate` at line 108).

## Changes

### 1. `check/isolate_check.go` — Temporal branch (lines 786-843)

**Move `TheoryContext()` inside the `if method != nil` block only.**

Change from (current code):
```go
cleanup := fakeMod.TheoryContext()           // line 786 — unconditional
if method != nil {                           // line 787
    if mod.Cfg.OnlyCheckUnprovable {
        ...
        cleanup()
        wsorts.Exit()
        ws.Exit()
        continue
    }
    err := method()
    if err != nil {
        ...
        cleanup()
        wsorts.Exit()
        ws.Exit()
        ...
    }
    fmt.Println("PASS")
} else {
    ...
    err := CheckIsolate(fakeMod, nil)
    if err != nil {
        cleanup()
        wsorts.Exit()
        ws.Exit()
        return err
    }
}
cleanup()
wsorts.Exit()
ws.Exit()
```

To:
```go
if method != nil {
    cleanup := fakeMod.TheoryContext()       // Only when method != nil
    if mod.Cfg.OnlyCheckUnprovable {
        ...
        cleanup()
        wsorts.Exit()
        ws.Exit()
        continue
    }
    err := method()
    if err != nil {
        ...
        cleanup()
        wsorts.Exit()
        ws.Exit()
        ...
    }
    fmt.Println("PASS")
    cleanup()
    wsorts.Exit()
    ws.Exit()
} else {
    ...
    err := CheckIsolate(fakeMod, nil)       // CheckIsolate handles its own TheoryContext
    if err != nil {
        wsorts.Exit()
        ws.Exit()
        return err
    }
    wsorts.Exit()
    ws.Exit()
}
```

Key changes:
- `cleanup := fakeMod.TheoryContext()` moves inside `if method != nil` block
- In the `else` (method==nil) branch: remove all `cleanup()` calls — no TheoryContext to clean up
- Move `wsorts.Exit(); ws.Exit()` into both branches (they're the outermost contexts, always cleaned up)

### 2. `check/isolate_check.go` — Non-temporal branch (lines 865-916)

**Same pattern** — move `TheoryContext()` inside `if method != nil` block only.

### Update comments

Update the Python-reference comment at lines 775-780 and 858-859 to accurately reflect the conditional:
```go
// Python: with lg.WithSymbols → with lg.WithSorts →
//             if method is not None: with im.module.theory_context(): method()
//             else: check_isolate()
```

### NOT changed

- `isolate_check.go:108` (`mod.TheoryContext()` inside `CheckIsolate`) — This is correct; Python's `check_isolate()` also calls `theory_context()` internally.
- `isolate_check.go:1131` and `1157` — Different code paths, unrelated.

## Critical files

- `check/isolate_check.go:786-843` — temporal branch
- `check/isolate_check.go:865-916` — non-temporal branch
- `ivy_check.py:820-846` — Python source of truth

## Verification

1. `go build ./...` compiles clean
2. `go test ./...` passes
3. `make golden` — Go should now emit `check.CheckIsolate ENTER` at line 845768 instead of `defToConstraint`, matching Python
