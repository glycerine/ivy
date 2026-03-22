package ast

// LowerVarStatements transforms var declarations into nested local scopes.
// Matches Python's lower_var_stmts (ivy_parser.py:2324-2350).
//
// Input:  [Atom("var", p), Atom("var", m), Assign(wr, false)]
// Output: [Atom("local", loc:p, Sequence(Atom("local", loc:m, Sequence(Assign(wr, false)))))]
//
// Each var introduces a new scope: the variable is renamed with "loc:" prefix
// and all subsequent statements are wrapped in a LocalAction.
func LowerVarStatements(stmts []Node) []Node {
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
		subst := map[string]Node{lhsName: NewSymbol(locName, nil)}

		// Python: lines = lower_var_stmts(stmts[idx+1:])
		lines := LowerVarStatements(stmts[idx+1:])

		// Python: lines = [subst_prefix_atoms_ast(s, subst, None, None) for s in lines]
		for i, line := range lines {
			lines[i] = SubstituteConstantsAst(line, subst)
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
