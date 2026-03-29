// ast_compat.go provides Args() and Clone() methods so that all logic.Expr
// concrete types also satisfy ast.Node. Since logic.Expr embeds ast.Node,
// every Expr value IS an ast.Node — no adapters or conversion helpers needed.
package ast

// --- Sort types ---

func (s *UninterpretedSort) Args() []Node      { return nil }
func (s *UninterpretedSort) Clone([]Node) Node { return s }

func (s *BooleanSort) Args() []Node      { return nil }
func (s *BooleanSort) Clone([]Node) Node { return s }

func (s *FunctionSort) Args() []Node {
	r := make([]Node, len(s.Sorts))
	for i, sub := range s.Sorts {
		r[i] = sub
	}
	return r
}
func (s *FunctionSort) Clone(args []Node) Node {
	sorts := make([]Sort, len(args))
	for i, a := range args {
		sorts[i] = a.(Sort)
	}
	return &FunctionSort{Sorts: sorts}
}

func (s *EnumeratedSort) Args() []Node      { return nil }
func (s *EnumeratedSort) Clone([]Node) Node { return s }

func (s *TopSort) Args() []Node      { return nil }
func (s *TopSort) Clone([]Node) Node { return s }

func (s *RangeSort) Args() []Node      { return nil }
func (s *RangeSort) Clone([]Node) Node { return s }

// --- Term types ---

func (v *Variable) Args() []Node      { return nil }
func (v *Variable) Clone([]Node) Node { return v }

func (c *Symbol) Args() []Node      { return nil }
func (c *Symbol) Clone([]Node) Node { return c }

func (a *Apply) Args() []Node {
	r := make([]Node, len(a.Terms))
	for i, t := range a.Terms {
		r[i] = t
	}
	return r
}
func (a *Apply) Clone(args []Node) Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &Apply{Func: a.Func, Terms: terms, aSort: a.aSort}
}

// --- Formula types ---

func (e *Eq) Args() []Node { return []Node{e.T1, e.T2} }
func (e *Eq) Clone(args []Node) Node {
	return &Eq{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (t *Ite) Args() []Node { return []Node{t.Cond, t.Then, t.Else} }
func (t *Ite) Clone(args []Node) Node {
	return &Ite{ISort: t.ISort, Cond: args[0].(Expr), Then: args[1].(Expr), Else: args[2].(Expr)}
}

func (n *Not) Args() []Node { return []Node{n.Body} }
func (n *Not) Clone(args []Node) Node {
	return &Not{Body: args[0].(Expr)}
}

func (g *Globally) Args() []Node { return []Node{g.Body} }
func (g *Globally) Clone(args []Node) Node {
	return &Globally{Environ: g.Environ, Body: args[0].(Expr)}
}

func (e *Eventually) Args() []Node { return []Node{e.Body} }
func (e *Eventually) Clone(args []Node) Node {
	return &Eventually{Environ: e.Environ, Body: args[0].(Expr)}
}

func (w *WhenOperator) Args() []Node { return []Node{w.T1, w.T2} }
func (w *WhenOperator) Clone(args []Node) Node {
	return &WhenOperator{WSort: w.WSort, Name: w.Name, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (c *Cond) Args() []Node { return []Node{c.T1, c.T2} }
func (c *Cond) Clone(args []Node) Node {
	return &Cond{CSort: c.CSort, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (a *And) Args() []Node {
	r := make([]Node, len(a.Terms))
	for i, t := range a.Terms {
		r[i] = t
	}
	return r
}
func (a *And) Clone(args []Node) Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &And{Terms: terms}
}

func (o *Or) Args() []Node {
	r := make([]Node, len(o.Terms))
	for i, t := range o.Terms {
		r[i] = t
	}
	return r
}
func (o *Or) Clone(args []Node) Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &Or{Terms: terms}
}

func (im *Implies) Args() []Node { return []Node{im.T1, im.T2} }
func (im *Implies) Clone(args []Node) Node {
	return &Implies{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (i *Iff) Args() []Node { return []Node{i.T1, i.T2} }
func (i *Iff) Clone(args []Node) Node {
	return &Iff{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (f *ForAll) Args() []Node { return []Node{f.Body} }
func (f *ForAll) Clone(args []Node) Node {
	cp := make([]*Variable, len(f.Variables))
	copy(cp, f.Variables)
	return &ForAll{Variables: cp, Body: args[0].(Expr)}
}

func (e *Exists) Args() []Node { return []Node{e.Body} }
func (e *Exists) Clone(args []Node) Node {
	cp := make([]*Variable, len(e.Variables))
	copy(cp, e.Variables)
	return &Exists{Variables: cp, Body: args[0].(Expr)}
}

func (l *Lambda) Args() []Node { return []Node{l.Body} }
func (l *Lambda) Clone(args []Node) Node {
	cp := make([]*Variable, len(l.Variables))
	copy(cp, l.Variables)
	return &Lambda{Variables: cp, Body: args[0].(Expr)}
}

func (nb *NamedBinder) Args() []Node { return []Node{nb.Body} }
func (nb *NamedBinder) Clone(args []Node) Node {
	cp := make([]*Variable, len(nb.Variables))
	copy(cp, nb.Variables)
	return &NamedBinder{Name: nb.Name, Variables: cp, Environ: nb.Environ, Body: args[0].(Expr)}
}

// --- LogicDefinition ---

func (d *LogicDefinition) Args() []Node { return []Node{d.Lhs, d.Rhs} }
func (d *LogicDefinition) Clone(args []Node) Node {
	return &LogicDefinition{Lhs: args[0].(Expr), Rhs: args[1].(Expr)}
}
