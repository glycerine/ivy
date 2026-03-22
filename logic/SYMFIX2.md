# Plan: Fix Structural Equality for Symbol Substitution/Renaming Maps

## Context

Python's `rename_ast` and `substitute_constants_ast` look up symbols via `subs.get(ast.rep, ...)` where `Symbol.rep = property(lambda self: self)` — the lookup key is the **Symbol object** with structural equality (name + sort via recstruct). ALL callers in Python build their dicts with Symbol object keys. There are zero string-keyed callers.

Go's `RenameAST` and `SubstituteConstantsAST` use `map[string]` keyed by `t.Name` — this is wrong for every call site. Fix: change to `map[lg.NodeKey]` keyed by `lg.Key(sym)`.

**Variable substitution is a different world.** Python's `Variable.rep = self.name` (a string). Functions like `SubstituteAstByName` correctly use string keys in both Python and Go. These are NOT changed.

---

## Phase 1: Core Function Signature Changes

### `clauseops/astutil.go`

**RenameAST (line 173):**
```go
// Before: func RenameAST(node lg.Expr, subs map[string]*lg.Symbol) lg.Expr
// After:  func RenameAST(node lg.Expr, subs map[lg.NodeKey]*lg.Symbol) lg.Expr
```
Internal lookup (line 183): `subs[t.Name]` → `subs[lg.Key(t)]`

**SubstituteConstantsAST (line 107):**
```go
// Before: func SubstituteConstantsAST(node lg.Expr, subs map[string]lg.Expr) lg.Expr
// After:  func SubstituteConstantsAST(node lg.Expr, subs map[lg.NodeKey]lg.Expr) lg.Expr
```
Internal lookup (line 117): `subs[t.Name]` → `subs[lg.Key(t)]`

### `clauseops/ops.go`

**RenameClauses (line 480):**
```go
// Before: func RenameClauses(clauses *Clauses, subs map[string]*lg.Symbol) *Clauses
// After:  func RenameClauses(clauses *Clauses, subs map[lg.NodeKey]*lg.Symbol) *Clauses
```

**SubstituteConstantsClauses (line 499):**
```go
// Before: func SubstituteConstantsClauses(clauses *Clauses, subs map[string]lg.Expr) *Clauses
// After:  func SubstituteConstantsClauses(clauses *Clauses, subs map[lg.NodeKey]lg.Expr) *Clauses
```

**RenameClausesByName (line 490):** This converts `map[string]string` (name→name) to Symbol map. The names-only case means we need the sort from somewhere. The caller's context determines the sort. Two options:
- (a) Change the caller to pass Symbols instead of strings
- (b) Use TopSort for the key, matching the TopSort already used for the value — BUT this won't match AST symbols that have real sorts.

Best approach: change RenameClausesByName to require a sig parameter or change callers to use RenameClauses directly with properly-keyed maps.

---

## Phase 2: Call Site Updates

Every call site must build `map[lg.NodeKey]*lg.Symbol` or `map[lg.NodeKey]lg.Expr` using `lg.Key(sym)` as the key.

### Pattern: Caller has full Symbol
```go
// Before: renaming[sym.Name] = replacement
// After:  renaming[lg.Key(sym)] = replacement
```

### Pattern: Caller only has string name + needs to construct Symbol
```go
// Before: renaming[name] = lg.NewSymbol(newName, lg.TopS)
// After:  sort := sig.Sorts[name]; sym := lg.NewSymbol(name, sort); renaming[lg.Key(sym)] = lg.NewSymbol(newName, sort)
```

### Call sites for RenameAST / RenameClauses:

| # | File | Line(s) | Current Key | Fix |
|---|------|---------|-------------|-----|
| 1 | `check/check.go` | 845 | `sym.Name` | `lg.Key(sym)` — sym available |
| 2 | `check/check.go` | 853 | `oldName` string | `lg.Key(lg.NewSymbol(oldName, s.CSort))` — s available |
| 3 | `check/helpers.go` | 191 | `renamedSym.Name` | `lg.Key(renamedSym)` |
| 4 | `mc/mine.go` | 150 | `sym.Name` | `lg.Key(sym)` |
| 5 | `mc/toaiger.go` | 150 | `newName` string | Construct Symbol with sort from sym |
| 6 | `mc/toaiger.go` | 347 | `sv` string | Construct Symbol with sort from sig |
| 7 | `mc/toaiger.go` | 368 | `tr.New(v)` string | Construct Symbol with sort from sig |
| 8 | `transrel/transrel.go` | 929 | `s.Name` | `lg.Key(s)` — s is `*lg.Symbol` |
| 9 | `transrel/transrel.go` | 933 | `name` string from sig.Sorts | Construct Symbol: `lg.Key(lg.NewSymbol(name, sig.Sorts[name]))` |
| 10 | `transrel/transrel.go` | 949 | `s.Name` | `lg.Key(s)` |
| 11 | `transrel/transrel.go` | 953 | `name` string from sig.Sorts | Same as #9 |
| 12 | `transrel/transrel.go` | 1004 | `New(s.Name)` | `lg.Key(lg.NewSymbol(New(s.Name), s.CSort))` |
| 13 | `transrel/transrel.go` | 1355 | `sym.Name` | `lg.Key(sym)` |
| 14 | `transrel/transrel.go` | 346-350 | `old` from name→name map | Need sig to get sort |
| 15 | `vmt/vmt.go` | 783-787 | `old` from name→name map | Need sig to get sort |
| 16 | `clauseops/ops.go` | 490-494 | `old` from name→name map | RenameClausesByName — redesign |
| 17 | `clauseops/subsume.go` | 337 | `s.Name` | `lg.Key(s)` |
| 18 | `clauseops/subsume.go` | 425+ | various | Update to lg.Key |
| 19 | `module/theory.go` | 258, 262 | via `subst` param | Update Rename caller |

### Call sites for SubstituteConstantsAST / SubstituteConstantsClauses:

| # | File | Line(s) | Current Key | Fix |
|---|------|---------|-------------|-----|
| 1 | `actions/action.go` | 106 | `fmt.Sprint(formal)` | `lg.Key(formal)` — formal is `*lg.Symbol` |
| 2 | `actions/action.go` | 448 | `p.Name` | `lg.Key(p)` — p is `*lg.Symbol` |
| 3 | `actions/action.go` | 537 | `p.Name` | `lg.Key(p)` |
| 4 | `actions/update.go` | 1794 | `oldSym.Name` | `lg.Key(oldSym)` |
| 5 | `actions/update.go` | 2130 | via subs param | Update caller |
| 6 | `compiler/phase6.go` | 588+ | symbol name | `lg.Key(sym)` |
| 7 | `module/theory.go` | 345 | `c.Name` via getLhsArgs | `lg.Key(c)` if c is Symbol |
| 8 | `isolate/strip.go` | 537 | `p.Name` | `lg.Key(p)` |
| 9 | `isolate/create.go` | 652 | `p.Name` | `lg.Key(p)` |
| 10 | `solver/clauses.go` | 232 | via subs param | Update caller |
| 11 | `solver/model.go` | 471 | via subs param | Update caller |
| 12 | `l2s/l2s_auto.go` | 201 | via subs param | Update caller |
| 13 | `ranking/tactic.go` | 192 | via subs param | Update caller |
| 14 | `clauseops/skolem.go` | 282 | `c.Name` | `lg.Key(c)` — c is from used_constants |

### NOT changed (Variable substitution — World 2, correctly string-keyed):

| File | Function | Why string keys are correct |
|------|----------|----------------------------|
| `clauseops/batch16.go` | DistinctVariableRenaming | Variable.rep = self.name (string) |
| `clauseops/skolem.go:55,82,113` | Skolemize* | Variable names |
| `actions/update.go:174,331,880` | dualFormula, skolemize, updateNondetAssign | Variable names |
| `proof/match.go:441` | betaReduce | Lambda variable names |
| `proof/phase5_goals.go:385` | nodeMapToStringMap | Variable names |
| `check/check.go:249` | variable subst in check | Variable names |

---

## Phase 3: RenameClausesByName redesign

`RenameClausesByName` takes `map[string]string` (name→name). It currently constructs Symbol keys with TopSort. Options:

(a) **Change signature** to `RenameClausesByName(clauses *Clauses, subs map[string]string, sig *il.Sig)` — look up sort from sig for each name.

(b) **Deprecate it** — have callers build `map[lg.NodeKey]*lg.Symbol` directly.

Option (a) is less disruptive. The function would become:
```go
func RenameClausesByName(clauses *Clauses, subs map[string]string, sig *il.Sig) *Clauses {
    constSubs := make(map[lg.NodeKey]*lg.Symbol, len(subs))
    for old, new_ := range subs {
        sort := lg.TopS
        if sig != nil {
            if entry, ok := sig.Symbols[old]; ok {
                sort = entry.Sort
            }
        }
        oldSym := lg.NewSymbol(old, sort)
        constSubs[lg.Key(oldSym)] = lg.NewSymbol(new_, sort)
    }
    return RenameClauses(clauses, constSubs)
}
```

Similarly, `transrel/transrel.go:renameFormula` and `vmt/vmt.go:renameNode` which do the same name→name→Symbol conversion need the same treatment.

---

## Files Modified

| File | Changes |
|------|---------|
| `clauseops/astutil.go` | RenameAST, SubstituteConstantsAST signatures + lookups |
| `clauseops/ops.go` | RenameClauses, SubstituteConstantsClauses, RenameClausesByName |
| `clauseops/subsume.go` | 2 call sites |
| `clauseops/skolem.go` | 1 call site (line 282, SubstituteConstantsClauses) |
| `check/check.go` | renaming map construction |
| `check/helpers.go` | RenameAST call |
| `transrel/transrel.go` | ~10 renaming map constructions + renameFormula |
| `transrel/phase4.go` | 1 call site |
| `mc/toaiger.go` | 3 renaming map constructions |
| `mc/mine.go` | 1 renaming map |
| `vmt/vmt.go` | 1 renaming map + renameNode |
| `actions/action.go` | 3 SubstituteConstantsAST calls |
| `actions/update.go` | SubstConstantsAction + related |
| `module/theory.go` | 3 call sites |
| `compiler/phase6.go` | SubstituteConstantsAST calls |
| `isolate/strip.go` | 1 call site |
| `isolate/create.go` | 1 call site |
| `solver/clauses.go` | 1 call site |
| `solver/model.go` | 1 call site |
| `l2s/l2s_auto.go` | 1 call site |
| `ranking/tactic.go` | 1 call site |

---

## Implementation Order

1. Change core functions in `clauseops/astutil.go` and `clauseops/ops.go`
2. Update `clauseops/` internal callers (subsume.go, skolem.go)
3. Update each downstream package, compiling after each:
   - module/ → actions/ → compiler/ → check/ → transrel/ → mc/ → vmt/ → isolate/ → solver/ → l2s/ → ranking/
4. `go build ./...` after each package
5. `go test ./...` at the end

## Verification

1. `go build ./...` — clean compile
2. `go test ./...` — all tests pass
3. `grep -rn 'map\[string\]\*lg\.Symbol' --include='*.go'` — verify no remaining string-keyed symbol substitution maps (except variable-only ones)
4. `grep -rn 'map\[string\]lg\.Expr' --include='*.go'` — verify remaining ones are variable-only
