package ast

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

func lvsCanons(stmts []Node) string {
	parts := make([]string, len(stmts))
	for i, s := range stmts {
		parts[i] = string(s.Canon())
	}
	return strings.Join(parts, " ")
}

// LowerVarStatements transforms var declarations into nested local scopes.
// Matches Python's lower_var_stmts (ivy_parser.py:2699-2726).
//
// Each VarAction introduces a new scope: the variable is renamed with "loc:" prefix
// and all subsequent statements are wrapped in a LocalAction.
// ThunkAction is also handled: the thunk name gets "loc:" prefix and
// a continuation Sequence is appended.
func LowerVarStatements(stmts []Node) []Node {
	xtracer.Trace("parser.lower_var_stmts ENTER in=%d canons=[%s]", len(stmts), lvsCanons(stmts))
	for idx, stmt := range stmts {
		// VarAction case: matches Python isinstance(stmt, VarAction)
		if v, ok := stmt.(*AstVarAction); ok {
			if len(v.Elems) < 1 {
				continue
			}
			lhs := v.Elems[0]
			var rhs Node
			if len(v.Elems) > 1 {
				rhs = v.Elems[1]
			}

			// Python: lsym = lhs.prefix('loc:')
			lsym := PrefixNode(lhs, "loc:")

			// Python: subst = {lhs.rep: lsym.rep}
			lhsRep := NodeRep(lhs)
			lsymRep := NodeRep(lsym)
			subst := map[string]string{lhsRep: lsymRep}

			// Python: lines = lower_var_stmts(stmts[idx+1:])
			lines := LowerVarStatements(stmts[idx+1:])

			// Python: lines = [subst_prefix_atoms_ast(s, subst, None, None) for s in lines]
			for i, line := range lines {
				lines[i] = SubstPrefixAtomsAst(line, subst, nil, nil, nil)
			}

			// Python: asgn = AssignAction(lsym, rhs) if rhs is not None else lsym
			cfg := v.Cfg
			var asgn Node
			if rhs != nil {
				asgn = cfg.NewAssignAction(lsym, rhs)
				asgn.SetLineno(stmt.GetLineno())
			} else {
				asgn = lsym
			}

			// Python: body = Sequence(*lines)
			body := cfg.NewSequence(lines...)
			body.SetLineno(stmt.GetLineno())

			// Python: res = LocalAction(*[asgn, body])
			res := cfg.NewLocalAction("parser.lower_var", asgn, body)
			res.SetLineno(body.GetLineno())

			result := append(stmts[:idx], res)
			xtracer.Trace("parser.lower_var_stmts RETURN out=%d canons=[%s]", len(result), lvsCanons(result))
			return result
		}

		// ThunkAction case: matches Python isinstance(stmt, ThunkAction)
		if t, ok := stmt.(*AstThunkAction); ok {
			// Python: name = stmt.args[1].rep
			name := NodeRep(t.Action)
			lname := "loc:" + name
			subst := map[string]string{name: lname}

			lines := LowerVarStatements(stmts[idx+1:])
			for i, line := range lines {
				lines[i] = SubstPrefixAtomsAst(line, subst, nil, nil, nil)
			}

			// Python: return stmts[:idx] + [stmt.clone(stmt.args + [Sequence(*lines)])]
			seq := &AstSequence{Stmts: lines}
			seq.Cfg = t.Cfg
			newArgs := append(t.Args(), seq)
			result := append(stmts[:idx:idx], t.Clone(newArgs))
			xtracer.Trace("parser.lower_var_stmts RETURN out=%d canons=[%s]", len(result), lvsCanons(result))
			return result
		}
	}
	xtracer.Trace("parser.lower_var_stmts RETURN out=%d canons=[%s]", len(stmts), lvsCanons(stmts))
	return stmts
}

// lvsCanon is unused but keeps fmt imported
var _ = fmt.Sprint

// PrefixNode clones a node and prepends s to its rep string.
// Matches Python Atom.prefix() / App.prefix() (ivy_ast.py:287-292, 358-363).
func PrefixNode(n Node, s string) Node {
	switch a := n.(type) {
	case *Atom:
		res := &Atom{Rep: s + a.Rep, Terms: a.Terms}
		res.Base = a.Base
		res.ASort = a.ASort
		return res
	case *App:
		newSym := &Symbol{Rep: s + a.Relname()}
		newSym.Cfg = a.Cfg
		res := &App{Rep: newSym, Terms: a.Terms}
		res.Base = a.Base
		res.ASort = a.ASort
		return res
	case *AstVariable:
		res := &AstVariable{Rep: s + a.Rep, VSort: a.VSort}
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
