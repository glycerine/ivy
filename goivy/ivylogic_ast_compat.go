// ast_compat.go provides Args() and Clone() methods so that ivylogic types
// (Some, Let, Literal) also satisfy ast.Node. Since lg.Expr embeds ast.Node,
// Expr values can be used directly as ast.Node — no conversion helpers needed.
package goivy

// --- Some ---

func (s *Some) Args() []Node {
	r := make([]Node, 0, len(s.Params)+3)
	for _, p := range s.Params {
		r = append(r, p)
	}
	r = append(r, s.Fmla)
	if s.IfVal != nil {
		r = append(r, s.IfVal)
	}
	if s.ElseVal != nil {
		r = append(r, s.ElseVal)
	}
	return r
}

func (s *Some) Clone(args []Node) Node {
	nParams := len(s.Params)
	params := make([]Expr, nParams)
	for i := 0; i < nParams; i++ {
		params[i] = args[i].(Expr)
	}
	result := &Some{Params: params, Fmla: args[nParams].(Expr)}
	idx := nParams + 1
	if s.IfVal != nil && idx < len(args) {
		result.IfVal = args[idx].(Expr)
		idx++
	}
	if s.ElseVal != nil && idx < len(args) {
		result.ElseVal = args[idx].(Expr)
	}
	return result
}

// --- Let ---

func (l *Let) Args() []Node {
	r := make([]Node, 0, len(l.Defs)+1)
	for _, d := range l.Defs {
		r = append(r, d)
	}
	r = append(r, l.Body)
	return r
}

func (l *Let) Clone(args []Node) Node {
	nDefs := len(l.Defs)
	defs := make([]Expr, nDefs)
	for i := 0; i < nDefs; i++ {
		defs[i] = args[i].(Expr)
	}
	return &Let{Defs: defs, Body: args[nDefs].(Expr)}
}

// --- Literal ---

func (l *Literal) Args() []Node { return []Node{l.Atom} }

func (l *Literal) Clone(args []Node) Node {
	return &Literal{Polarity: l.Polarity, Atom: args[0].(Expr)}
}
