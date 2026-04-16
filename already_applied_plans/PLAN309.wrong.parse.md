# Fix grammar rule: NamedBinder application creates wrong AST node type
**Created: 2026-04-15 (during investigation of log.tlb divergence at line 092064)**

## Context

The Go grammar rule for `($name vars. fmla)(terms)` creates an `Atom(rep="", terms=[NamedBinder, ...])`, but Python creates `App(rep=NamedBinder, terms=[...])`. This structural mismatch causes:

1. **Wrong vocab traversal**: Go's `IterSymbolsASTNode` skips the empty-rep Atom but recurses into the NamedBinder body as a child. Python yields the NamedBinder (as App.rep) and does NOT recurse into it.
2. **Wrong trace output**: Go emits symbols from the NamedBinder's body where Python emits the NamedBinder itself.
3. **The `default: panicf(...)` in `VocabNamesUpdate`** would crash if a NamedBinder were ever yielded (it currently isn't, due to the Atom workaround).

The divergence appears in tlb.ivy's l2s proof at invariants like:
```
invariant (pc_b1(sk0) & ($l2s_w P. scheduled(P))(sk0)) | l2s_d(userpmap(sk0))
```

The Go compiler at `compiler.go:823-838` already handles App with NamedBinder rep correctly — this code path was just never reached from parsing.

## Plan

### 1. Fix Go grammar rule — `grammar_v17.y:480-481`

**Bug**: Go creates `NewAtom("", [binder, terms...])` — wrong node type, wrong structure.
**Fix**: Create `NewApp(binder, terms...)` to match Python `App(NamedBinder, terms)` at `ivy_logic_parser.py:761`.

```go
binder := acfg(v17lex).NewNamedBinder($3, $4, $6)
binder.SetLineno(getLineno(v17lex))
$$ = acfg(v17lex).NewApp(binder, $9...)
$$.(*ast.App).SetLineno(getLineno(v17lex))
```

### 2. Regenerate goyacc parser

Run `goyacc` to regenerate `grammar_v17.go` from the updated `.y` file.

### 3. Add `*NamedBinder` case to `VocabNamesUpdate` — `ast/tactic.go:463-483`

Add a case after `*This`:
```go
case *NamedBinder:
    // Python adds the NamedBinder object to the set, but it never matches
    // any string lookup (it's inert). Match that: trace but don't add.
    if xtracer.Enabled {
        xtracer.Trace("vocab.add src=App name=%s val_type=NamedBinder", v.Name)
    }
```

Remove or update the `default: panicf(...)` to be safe.

### 4. Fix Python trace to be portable — `ivy_ast.py:1892`

Change the trace in `symbols_ivy_ast` to use `val.name` when val is a NamedBinder (instead of `str(val)` which gives a non-reproducible memory address):

```python
if isinstance(val, NamedBinder):
    xtracer.trace("vocab.add src=%s name=%s val_type=%s" % (src, val.name, val_type))
else:
    xtracer.trace("vocab.add src=%s name=%s val_type=%s" % (src, val, val_type))
```

## Files to modify

| File | Change |
|------|--------|
| `lalr_logicparser/grammar_v17.y:480-481` | `NewAtom("",...)` → `NewApp(binder, $9...)` |
| `lalr_logicparser/grammar_v17.go` | Regenerated from .y |
| `ast/tactic.go:463-483` | Add `*NamedBinder` case in `VocabNamesUpdate` |
| `~/ivy/pyivy/ivy/ivy/ivy_ast.py:1892` | Portable trace for NamedBinder |

## Verification

Run `make tlb` and verify line 092064 now matches between Go and Python (both should emit `vocab.add src=App name=l2s_w val_type=NamedBinder`), and the test progresses further.
