# Fix: Golden test divergence at line 157372 — missing `LocalAction.__init__` trace

**Created**: 2026-04-01 22:30
**Previous fix**: Consolidated `SubstituteConstantsAST` (completed, moved divergence from 157358→157372)

## Context

After the SubstituteConstantsAST consolidation, the golden test now diverges at XTRACE line 157372:

```
157357  go : XTRACE: actions.IfAction.int_update ENTER
        py : XTRACE: actions.IfAction.int_update ENTER

157358-157371: 14 matching substitute_constants_action traces (And→Or→Eq→Apply tree)

157372  go : XTRACE: actions.IfAction.int_update EXIT
        py : XTRACE: LocalAction.__init__ uniqueID=1184 caller=actions.IfAction.action_update
```

Go exits `IfAction.int_update` while Python creates a `LocalAction`. The substitute traces prove both are on the `SomeCondition` path (traces come from `subactionsSome` → `SubstituteConstantsExpr(fmla, subst)` at `actions/action.go:531`). After the substitution, Go should call `NewLocalActionOn` at line 596, which traces `LocalAction.__init__`. But the trace never appears — instead, the deferred EXIT trace fires.

## Analysis

### Confirmed facts

1. **Both Go and Python are on the Some path**: substitute traces at 157358-157371 come from `subactionsSome`'s `SubstituteConstantsExpr(fmla, subst)` call (line 531). The simple boolean path would produce `IntUpdate ENTER type=...` as its first nested trace, not `substitute_constants_action`.

2. **The 14 substitute traces match perfectly**: the tree traversal And(3)→Or(4)→6×Eq(2)/Apply(1) is depth-first, all children processed. `SubstituteConstantsExpr` returns successfully.

3. **No early return between lines 531-596**: After `SubstituteConstantsExpr(fmla, subst)`, the code goes through SomeMinMax handling (lines 534-586, may be skipped if Kind is not min/max) and then simple struct construction (lines 588-595) before `NewLocalActionOn` at line 596. None of these have `return` statements.

4. **`NewLocalActionOn` traces BEFORE doing anything else**: Line 869 of `actions/action.go` traces `LocalAction.__init__` after the nil checks (lines 861-865) but before any other work.

5. **The `defer EXIT` fires when `IntUpdate` returns**: The defer on line 1516 fires when the function returns, which is AFTER all nested work completes. If `subactionsSome` completes normally, the trace order should be: substitute traces → LocalAction.__init__ → [nested IntUpdate traces] → EXIT.

### Most likely cause: PANIC between substitute and LocalAction

Since Go shows EXIT right after the substitute traces (no LocalAction trace, no nested IntUpdate traces), something prevents `subactionsSome` from reaching `NewLocalActionOn`. The `defer EXIT` fires during panic stack unwinding, outputting the EXIT trace on stdout. The panic message goes to stderr (not captured by the golden test).

Possible panic locations:
- **SomeMinMax handling (lines 534-586)**: If `some.Kind` is "some_min" or "some_max", operations like `lg.NewApply`, `il.NewEqualsNode`, or `lg.NewAnd` could panic on nil inputs
- **`NewLocalActionOn` nil checks (lines 861-865)**: If `actCfg` is nil or `actCfg.IuCfg` is nil, it panics BEFORE the trace. `actCfg` comes from `ctx.ActCfg` which is set from `domain.Cfg.ActCfg` at `update.go:2175`. Verified that `NewActionsConfig()` sets IuCfg and the compiler initializes ActCfg properly — but worth confirming at runtime.

### Alternative cause: Build not reflecting latest source

Unlikely since `make golden` rebuilds `goivy_check_xtrace`, but worth verifying with a clean build.

## Changes

### Step 1: Diagnose — check for panic on stderr

Run the Go binary directly and capture stderr:

```bash
cd ~/ivy/goivy && go build -o /tmp/goivy_check ./cmd/goivy_check/ && \
/tmp/goivy_check isolate=cf_live ~/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy \
  >/tmp/goivy_stdout.log 2>/tmp/goivy_stderr.log; \
echo "exit code: $?"
cat /tmp/goivy_stderr.log
```

If there's a panic, the stack trace in `stderr.log` will point to the exact line. Look for:
- `goroutine 1 [running]:` — panic stack trace
- The function and line number where the panic occurs

If no panic, add diagnostic traces to stderr in `subactionsSome`:

```go
// Add right after line 531 (SubstituteConstantsExpr call):
fmt.Fprintf(os.Stderr, "DEBUG: subactionsSome after substitute, kind=%q\n", some.Kind)

// Add right before line 596 (NewLocalActionOn call):
fmt.Fprintf(os.Stderr, "DEBUG: subactionsSome before NewLocalActionOn, actCfg=%v\n", actCfg != nil)
```

### Step 2: Fix based on diagnosis

**If panic in SomeMinMax handling**: Fix the nil-producing operations. Likely cause is `lg.NewApply` returning `(nil, error)` where the error is ignored. The nil then propagates to `lg.NewAnd(nil, ...)` which panics when later code calls methods on the nil term.

**If panic in NewLocalActionOn (nil actCfg or IuCfg)**: Trace where ActCfg becomes nil. Check if `domain.Cfg.ActCfg` is properly initialized for this particular module context.

**If no panic**: Deep investigation needed — add more traces to understand the exact execution path.

### Step 3: Remove diagnostic traces

After fixing, remove any debug `fmt.Fprintf(os.Stderr, ...)` calls added in Step 1.

## Key files

- `actions/action.go:519-605` — `subactionsSome` (where divergence occurs)
- `actions/action.go:860-882` — `NewLocalActionOn` (expected trace)
- `actions/update.go:1514-1527` — `IfAction.IntUpdate` (ENTER/defer EXIT)
- `actions/update.go:1558-1589` — `intUpdateWithSubactions`
- `clauseops/astutil.go:126-155` — new `SubstituteConstantsAST/Expr`

## Verification

```bash
cd ~/ivy/goivy && go build ./... && go test ./... && make golden
```

Check that:
1. `go build` and `go test` pass
2. No panic on stderr when running goivy_check directly
3. The divergence in `log.red` advances past line 157372
