# Fix: `too many arguments to rfn2.abs.begun` — Parameterized Instance Expansion Broken

**Created: 2026-03-22**

## Context

`goivy_check` fails at line 3178: `too many arguments to rfn2.abs.begun`. The expression `rfn2.abs.begun(P,X)` passes 2 args but the compiler thinks `rfn2.abs.begun` only takes 1.

## Root Cause

There are TWO module instantiation code paths in the parser. The NEW one (`expandInstantiation` in parser.go) was fixed to use the unified rewriter, but the OLD one (`parseInstantiateDeclMulti` in decl.go) is still broken with two bugs:

### Bug 1: Prefix atom loses parameters (decl.go lines 1340-1356)

```go
prefix = a.Rep                    // extracts just the string "abs"
pref = ast.NewAtom(prefix)        // creates Atom("abs") with NO TERMS
```

Should use `inst.Name` directly which carries `Atom("abs", [Variable("P:proc")])` — with the parameter. This causes `ComposeAtoms(pref, atom)` to NOT prepend the `P:proc` parameter to inner declarations.

### Bug 2: All names marked as static (decl.go lines 1361-1365)

```go
static := make(map[string]bool)
for name := range defined {
    static[name] = true  // Marks EVERYTHING as static!
}
```

The `static` set should only contain types and destructors (matching `collectStaticNames()`). When a name is in `static`, `RewriteAtom` strips the pref's args: `thePref = NewAtom(thePref.Rep)` — creating a pref with NO terms. This prevents parameter propagation even when the pref has terms.

### Combined Effect

For `instance abs(P:proc) : mymod(nat)` expanding `var begun(X:n) : bool`:
1. `pref = Atom("abs")` — missing `P:proc` (Bug 1)
2. `static["begun"] = true` — incorrectly marked as static (Bug 2)
3. `ComposeAtoms(Atom("abs"), Atom("begun", [X:n]))` → `Atom("abs.begun", [X:n])` — missing P, n not substituted
4. Result: `abs.begun(X:n)` instead of `abs.begun(P:proc, X:nat)`

## Fix

**File: `parser/decl.go`**, in `parseInstantiateDeclMulti`, around lines 1340-1365:

### Fix Bug 1: Use inst.Name directly as pref

```go
// Before (broken):
var pref *ast.Atom
if prefix != "" {
    pref = ast.NewAtom(prefix)
}

// After (fixed, matching expandInstantiation in parser.go):
var pref *ast.Atom
if instNode.Name != nil {
    if a, ok := instNode.Name.(*ast.Atom); ok {
        pref = a
    }
}
```

### Fix Bug 2: Use collectStaticNames instead of marking everything

```go
// Before (broken):
static := make(map[string]bool)
for name := range defined {
    static[name] = true
}

// After (fixed, matching expandInstantiation in parser.go):
static := collectStaticNames(modDef.BodyDecls)
```

## Verification

1. `go build ./...` compiles
2. `go test ./...` passes
3. `instance abs(P:proc) : mymod(nat)` produces `abs.begun(P:proc, X:nat)` with 2 terms
4. `./goivy_check isolate=cf_live ord_live.ivy` gets past "too many arguments"
