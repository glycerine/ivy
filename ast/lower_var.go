package ast

import "github.com/glycerine/goivy/xtracer"

// LowerVarStatements transforms var declarations into nested local scopes.
// Matches Python's lower_var_stmts (ivy_parser.py:2324-2350).
//
// Input:  [Atom("var", p), Atom("var", m), Assign(wr, false)]
// Output: [Atom("local", loc:p, Sequence(Atom("local", loc:m, Sequence(Assign(wr, false)))))]
//
// Each var introduces a new scope: the variable is renamed with "loc:" prefix
// and all subsequent statements are wrapped in a LocalAction.
func LowerVarStatements(stmts []Node) []Node {
	xtracer.Trace("parser.lower_var_stmts ENTER")
	for idx, stmt := range stmts {
		a, ok := stmt.(*Atom)
		if !ok || a.Rep != "var" {
			continue
		}
		if len(a.Terms) < 1 {
			continue
		}

		// Python: lhs = stmt.args[0]; rhs = stmt.args[1] if len > 1 else None
		lhs := a.Terms[0]
		var rhs Node
		if len(a.Terms) > 1 {
			rhs = a.Terms[1]
		}

		// Python: lsym = lhs.prefix('loc:')
		var lhsName string
		switch v := lhs.(type) {
		case *Variable:
			lhsName = v.Rep
		case *Atom:
			lhsName = v.Rep
		case *Symbol:
			lhsName = v.Rep
		}
		locName := "loc:" + lhsName
		lsym := NewAtom(locName)
		if v, ok := lhs.(*Variable); ok && v.VSort != nil {
			lsym = NewAtom(locName, v.VSort)
		}

		// Python: subst = {lhs.rep: lsym.rep}
		subst := map[string]string{lhsName: locName}

		// Python: lines = lower_var_stmts(stmts[idx+1:])
		lines := LowerVarStatements(stmts[idx+1:])

		// Python: lines = [subst_prefix_atoms_ast(s, subst, None, None) for s in lines]
		for i, line := range lines {
			lines[i] = SubstPrefixAtomsAst(line, subst, nil, nil, nil)
		}

		// Python: asgn = AssignAction(lsym, rhs) if rhs else lsym
		var asgn Node
		if rhs != nil {
			asgn = NewAtom(":=", lsym, rhs)
		} else {
			asgn = lsym
		}

		// Python: body = Sequence(*lines)
		var body Node
		if len(lines) == 0 {
			body = NewAnd() // empty sequence
		} else if len(lines) == 1 {
			body = lines[0]
		} else {
			body = NewAnd(lines...)
		}

		// Python: res = LocalAction(*[asgn, body])
		local := NewAtom("local", asgn, body)

		return append(stmts[:idx], local)
	}
	return stmts
}

/*
To answer your question: yes, lower_var.go:52 was also a bug. It was using
  SubstituteConstantsAst but the Python uses subst_prefix_atoms_ast(s, subst, None,
  None) — a completely different mechanism that works through SubstSubscripts /
  ParseName().Subst() for name rewriting. The faithful fix is to call
  SubstPrefixAtomsAst(line, subst, nil, nil, nil) with a map[string]string subst
  (matching Python's string→string subst dict), not any flavor of
  SubstituteConstantsAst.

❯ are we sure this is a map[string] and not a map[NodeKey] situation? have we
verified that the python is not using structural equivalence?

Good question. Let me verify exactly what Python uses.

- Line 2330: subst = {lhs.rep: lsym.rep} — both .rep are strings. And line 2346:
  subst = {name: lname} — also plain strings.

  Let me confirm .rep is always a string by checking Atom.rep and prefix():

- It's str → str all the way through. Atom.rep is a string, prefix('loc:') prepends
  to self.rep (a string), and the subst dict is {str: str}. The lookup in
  Symbol.subst is subst.get(self.rep, self.rep) — plain string key lookup.

  No structural equivalence, no NodeKey. Our map[string]string is the faithful port.
*/
