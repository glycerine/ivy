package ast

import "github.com/glycerine/goivy/xtracer"

// LowerVarStatements transforms var declarations into nested local scopes.
// Matches Python's lower_var_stmts (ivy_parser.py:2699-2726).
//
// Each VarAction introduces a new scope: the variable is renamed with "loc:" prefix
// and all subsequent statements are wrapped in a LocalAction.
// ThunkAction is also handled: the thunk name gets "loc:" prefix and
// a continuation Sequence is appended.
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
			newArgs := append(t.Args(), NewSequence(lines...))
			return append(stmts[:idx], t.Clone(newArgs))
		}
	}
	return stmts
}

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
