# Fix: `undefined isolate: cf_live` — Only First Compilation Pass Running

**Created: 2026-03-22**

## Context

`goivy_check isolate=cf_live ord_live.ivy` fails with `undefined isolate: cf_live`. The isolate is defined with a body (`isolate cf_live = { ... }`), which produces `IsolateObjectDecl`. The ARGSetup pass (3rd compilation pass) registers isolates. But `SourceFile` only runs the 1st pass (DomainSetup).

## Root Cause

Python's `ivy_compile()` runs ALL THREE passes:
1. `IvyDomainSetup` — types, relations, constants, axioms, definitions
2. `IvyConjectureSetup` — conjectures
3. `IvyARGSetup` — **isolates**, exports, delegates, actions, initializers

Go's `SourceFile()` only runs DomainSetup:
```go
comp := compiler.New(sig, mod)
di := compiler.NewDomainSetup(comp)
di.ProcessDecls(result.Decls)  // Only pass 1!
```

Go's `compiler.IvyCompile()` runs all three passes, but it's NOT called from `SourceFile`.

## Fix Applied

Changed `SourceFile` to call `compiler.IvyCompile(result.Decls, mod)` which runs all three passes.

## Secondary Issue Exposed

The full three-pass compilation exposes a pre-existing bug: `action1.ivy` (which has `export bar.a`) fails with `undefined action: bar.a` in the ARGSetup pass. This means the ARGSetup's export handler (`check_is_action`) can't find actions that were declared inside isolate bodies and prefixed by the parser.

This is a separate bug in how ARGSetup processes action declarations from isolate bodies — it needs to be fixed independently. The bug was masked because `SourceFile` previously skipped ARGSetup.

## Status

- The `IvyCompile` call is correct and matches Python
- The `IsolateObjectDecl` → `registerIsolateDecl` fix is correct
- The ARGSetup action registration bug needs to be fixed next
