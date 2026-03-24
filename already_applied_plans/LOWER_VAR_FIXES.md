# Fix: `LowerVarStatements` uses wrong types — checks `*Atom` instead of `*VarAction`

**Created:** 2026-03-24

## Context

`make golden` diverges at line 9034. Go emits `parser.get_lineno ENTER` where Python emits `parser.lower_var_stmts ENTER`. The root cause: Go's `LowerVarStatements` (`ast/lower_var.go:16`) checks for `stmt.(*Atom)` with `Rep == "var"`, but the lalr_full parser creates `*VarAction` nodes. So the type check always fails, the function never finds var declarations, never recurses, and returns early.

## Python reference: `lower_var_stmts` (`ivy_parser.py:2699-2726`)

```python
def lower_var_stmts(stmts):
    for idx, stmt in enumerate(stmts):
        if isinstance(stmt, VarAction):           # ← checks VarAction, not Atom
            lhs = stmt.args[0]
            rhs = stmt.args[1] if len(stmt.args) > 1 else None
            lsym = lhs.prefix('loc:')             # ← clone with 'loc:' prefix, preserving args
            subst = {lhs.rep: lsym.rep}
            lines = lower_var_stmts(stmts[idx+1:])
            lines = [subst_prefix_atoms_ast(s, subst, None, None) for s in lines]
            asgn = AssignAction(lsym, rhs) if rhs is not None else lsym  # ← AssignAction
            asgn.lineno = stmt.lineno
            body = Sequence(*lines)               # ← Sequence, not And
            body.lineno = stmt.lineno
            res = LocalAction(*[asgn, body])       # ← LocalAction, not Atom("local")
            res.lineno = body.lineno
            return stmts[:idx] + [res]
        if isinstance(stmt, ThunkAction):          # ← also handles ThunkAction
            name = stmt.args[1].rep
            lname = 'loc:' + name
            subst = {name: lname}
            lines = lower_var_stmts(stmts[idx+1:])
            lines = [subst_prefix_atoms_ast(s, subst, None, None) for s in lines]
            return stmts[:idx] + [stmt.clone(stmt.args + [Sequence(*lines)])]
    return stmts
```

## Bugs in current Go implementation (`ast/lower_var.go:13-82`)

1. **Line 16**: `stmt.(*Atom)` — should be `stmt.(*VarAction)`
2. **Line 20**: `a.Terms` — should be `v.Elems` (VarAction uses Elems, not Terms)
3. **Line 42**: `lsym := NewAtom(locName)` — should clone lhs with prefix, preserving args (Python's `lhs.prefix('loc:')`)
4. **Line 43-45**: Special-cases Variable sort — Python's `prefix()` just clones and prepends, preserving all args regardless of type
5. **Line 61**: `NewAtom(":=", lsym, rhs)` — should be `NewAssignAction(lsym, rhs)` with lineno
6. **Lines 68-74**: `NewAnd(lines...)` — should be `NewSequence(lines...)` with lineno
7. **Line 77**: `NewAtom("local", asgn, body)` — should be `NewLocalAction(asgn, body)` with lineno
8. **Missing**: No `ThunkAction` handling at all
9. **Missing**: No lineno propagation on asgn, body, res

## Fix: Rewrite `ast/lower_var.go`

```go
func LowerVarStatements(stmts []Node) []Node {
	xtracer.Trace("parser.lower_var_stmts ENTER")
	for idx, stmt := range stmts {
		// VarAction case: matches Python isinstance(stmt, VarAction)
		if v, ok := stmt.(*VarAction); ok {
			if len(v.Elems) < 1 {
				continue
			}
			lhs := v.Elems[0]
			var rhs Node
			if len(v.Elems) > 1 {
				rhs = v.Elems[1]
			}

			// Python: lsym = lhs.prefix('loc:')
			// prefix() clones with 'loc:' + rep, preserving args and lineno
			lsym := prefixNode(lhs, "loc:")

			// Python: subst = {lhs.rep: lsym.rep}
			lhsRep := nodeRep(lhs)
			lsymRep := nodeRep(lsym)
			subst := map[string]string{lhsRep: lsymRep}

			// Python: lines = lower_var_stmts(stmts[idx+1:])
			lines := LowerVarStatements(stmts[idx+1:])

			// Python: lines = [subst_prefix_atoms_ast(s, subst, None, None) for s in lines]
			for i, line := range lines {
				lines[i] = SubstPrefixAtomsAst(line, subst, nil, nil, nil)
			}

			// Python: asgn = AssignAction(lsym, rhs) if rhs is not None else lsym
			var asgn Node
			if rhs != nil {
				asgn = NewAssignAction(lsym, rhs)
				asgn.SetLineno(stmt.GetLineno())
			} else {
				asgn = lsym
			}

			// Python: body = Sequence(*lines)
			body := NewSequence(lines...)
			body.SetLineno(stmt.GetLineno())

			// Python: res = LocalAction(*[asgn, body])
			res := NewLocalAction(asgn, body)
			res.SetLineno(body.GetLineno())

			return append(stmts[:idx], res)
		}

		// ThunkAction case: matches Python isinstance(stmt, ThunkAction)
		if t, ok := stmt.(*ThunkAction); ok {
			// Python: name = stmt.args[1].rep
			name := nodeRep(t.Action)
			lname := "loc:" + name
			subst := map[string]string{name: lname}

			lines := LowerVarStatements(stmts[idx+1:])
			for i, line := range lines {
				lines[i] = SubstPrefixAtomsAst(line, subst, nil, nil, nil)
			}

			// Python: return stmts[:idx] + [stmt.clone(stmt.args + [Sequence(*lines)])]
			// Python ThunkAction stores args as flat list and clone handles 5 args.
			// Go's Clone only handles 4. We need to add a Continuation field to ThunkAction
			// OR construct manually. For now, add Continuation field.
			cloned := &ThunkAction{
				Base:         t.Base,
				Label:        t.Label,
				Action:       t.Action,
				Sort:         t.Sort,
				Body:         t.Body,
				Continuation: NewSequence(lines...),
			}
			return append(stmts[:idx], cloned)
		}
	}
	return stmts
}
```

### Helper needed: `prefixNode` (matches Python's `Atom.prefix` / `App.prefix`)

Python's `prefix('loc:')` clones the node and prepends `s` to its `rep`, preserving args and lineno. Need a Go helper:

```go
// prefixNode clones a node and prepends s to its rep string.
// Matches Python Atom.prefix() / App.prefix() (ivy_ast.py:287-292, 358-363).
func prefixNode(n Node, s string) Node {
	switch a := n.(type) {
	case *Atom:
		res := NewAtom(s+a.Rep, a.Terms...)
		res.Base = a.Base
		res.ASort = a.ASort
		return res
	case *App:
		newSym := NewSymbol(s+a.Relname(), nil)
		res := NewApp(newSym, a.Terms...)
		res.Base = a.Base
		res.ASort = a.ASort
		return res
	case *Variable:
		res := &Variable{Rep: s + a.Rep, VSort: a.VSort}
		res.Base = a.Base
		return res
	case *Symbol:
		res := &Symbol{Rep: s + a.Rep}
		res.Base = a.Base
		return res
	default:
		return n
	}
}
```

### ThunkAction.Args() consideration

Python's `ThunkAction` stores args as a flat list: `[label, action, sort, body]`. When `lower_var_stmts` does `stmt.clone(stmt.args + [Sequence(*lines)])`, it appends a Sequence as a 5th arg.

Go's `ThunkAction` has named fields `{Label, Action, Sort, Body}` and `Args()` returns `[Label, Action, Sort, Body]`. `Clone(args)` sets these from `args[0..3]`. Appending a 5th element would need Clone to handle it. Let me verify Go's Clone:

Check `ThunkAction.Clone()` at `ast/ast.go` — it takes exactly 4 args. Python appends a 5th. This needs a different approach: create a new ThunkAction manually with the Sequence appended to the body, or extend Clone. Need to read more carefully how Python uses the cloned result downstream.

Actually, re-reading Python: `stmt.clone(stmt.args + [Sequence(*lines)])` — ThunkAction's args are `[label, action_atom, type, body]` (4 args normally). Adding `Sequence(*lines)` makes it 5. But ThunkAction.clone creates `type(self)(*args)` which passes all as positional to `__init__`. We'd need to check what ThunkAction.__init__ does with 5 args. This is a subtlety to handle during implementation.

## Files to modify

1. **`ast/lower_var.go`** — rewrite `LowerVarStatements` to use `*VarAction`, `*ThunkAction`, `AssignAction`, `Sequence`, `LocalAction`, and proper lineno propagation. Add `prefixNode` helper.
2. **`ast/ast.go`** — add `Continuation Node` field to `ThunkAction` struct (Python supports 5-arg ThunkAction via flat args list; Go needs an explicit field). Update `Args()`, `Clone()`, `String()`, and `Canon()` to handle the optional continuation.

## Verification

```bash
cd ~/go/src/github.com/glycerine/goivy/ast && go test -v -run "LowerVar"
cd ~/go/src/github.com/glycerine/goivy && make golden
```

Confirm divergence moves past line 9034.
