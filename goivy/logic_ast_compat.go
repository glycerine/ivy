// ast_compat.go provides Args() and Clone() methods so that all logic.Expr
// concrete types also satisfy ast.Node. Since logic.Expr embeds ast.Node,
// every Expr value IS an ast.Node — no adapters or conversion helpers needed.
package goivy

// --- Sort types ---

func (s *UninterpretedSort) Args() []Node      { return nil }
func (s *UninterpretedSort) Clone([]Node) Node { return s }

func (s *BooleanSort) Args() []Node      { return nil }
func (s *BooleanSort) Clone([]Node) Node { return s }

func (s *LogicFunctionSort) Args() []Node {
	r := make([]Node, len(s.Sorts))
	for i, sub := range s.Sorts {
		r[i] = sub
	}
	return r
}
func (s *LogicFunctionSort) Clone(args []Node) Node {
	sorts := make([]Sort, len(args))
	for i, a := range args {
		sorts[i] = a.(Sort)
	}
	return &LogicFunctionSort{Sorts: sorts}
}

func (s *LogicEnumeratedSort) Args() []Node      { return nil }
func (s *LogicEnumeratedSort) Clone([]Node) Node { return s }

func (s *TopSort) Args() []Node      { return nil }
func (s *TopSort) Clone([]Node) Node { return s }

func (s *RangeSort) Args() []Node      { return nil }
func (s *RangeSort) Clone([]Node) Node { return s }

// --- Term types ---

func (v *LogicVariable) Args() []Node      { return nil }
func (v *LogicVariable) Clone([]Node) Node { return v }

func (c *Const) Args() []Node      { return nil }
func (c *Const) Clone([]Node) Node { return c }

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

func (t *LogicIte) Args() []Node { return []Node{t.Cond, t.Then, t.Else} }
func (t *LogicIte) Clone(args []Node) Node {
	return &LogicIte{ISort: t.ISort, Cond: args[0].(Expr), Then: args[1].(Expr), Else: args[2].(Expr)}
}

func (n *LogicNot) Args() []Node { return []Node{n.Body} }
func (n *LogicNot) Clone(args []Node) Node {
	return &LogicNot{Body: args[0].(Expr)}
}

func (g *LogicGlobally) Args() []Node { return []Node{g.Body} }
func (g *LogicGlobally) Clone(args []Node) Node {
	return &LogicGlobally{Environ: g.Environ, Body: args[0].(Expr)}
}

func (e *LogicEventually) Args() []Node { return []Node{e.Body} }
func (e *LogicEventually) Clone(args []Node) Node {
	return &LogicEventually{Environ: e.Environ, Body: args[0].(Expr)}
}

func (w *LogicWhenOperator) Args() []Node { return []Node{w.T1, w.T2} }
func (w *LogicWhenOperator) Clone(args []Node) Node {
	return &LogicWhenOperator{WSort: w.WSort, Name: w.Name, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (c *Cond) Args() []Node { return []Node{c.T1, c.T2} }
func (c *Cond) Clone(args []Node) Node {
	return &Cond{CSort: c.CSort, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (a *LogicAnd) Args() []Node {
	r := make([]Node, len(a.Terms))
	for i, t := range a.Terms {
		r[i] = t
	}
	return r
}
func (a *LogicAnd) Clone(args []Node) Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &LogicAnd{Terms: terms}
}

func (o *LogicOr) Args() []Node {
	r := make([]Node, len(o.Terms))
	for i, t := range o.Terms {
		r[i] = t
	}
	return r
}
func (o *LogicOr) Clone(args []Node) Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &LogicOr{Terms: terms}
}

func (im *LogicImplies) Args() []Node { return []Node{im.T1, im.T2} }
func (im *LogicImplies) Clone(args []Node) Node {
	return &LogicImplies{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (i *LogicIff) Args() []Node { return []Node{i.T1, i.T2} }
func (i *LogicIff) Clone(args []Node) Node {
	return &LogicIff{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (f *ForAll) Args() []Node { return []Node{f.Body} }
func (f *ForAll) Clone(args []Node) Node {
	cp := make([]*LogicVariable, len(f.Variables))
	copy(cp, f.Variables)
	return &ForAll{Variables: cp, Body: args[0].(Expr)}
}

func (e *LogicExists) Args() []Node { return []Node{e.Body} }
func (e *LogicExists) Clone(args []Node) Node {
	cp := make([]*LogicVariable, len(e.Variables))
	copy(cp, e.Variables)
	return &LogicExists{Variables: cp, Body: args[0].(Expr)}
}

func (l *Lambda) Args() []Node { return []Node{l.Body} }
func (l *Lambda) Clone(args []Node) Node {
	cp := make([]*LogicVariable, len(l.Variables))
	copy(cp, l.Variables)
	return &Lambda{Variables: cp, Body: args[0].(Expr)}
}

func (nb *LogicNamedBinder) Args() []Node { return []Node{nb.Body} }
func (nb *LogicNamedBinder) Clone(args []Node) Node {
	cp := make([]*LogicVariable, len(nb.Variables))
	copy(cp, nb.Variables)
	return &LogicNamedBinder{Name: nb.Name, Variables: cp, Environ: nb.Environ, Body: args[0].(Expr)}
}

// --- Definition ---

func (d *LogicDefinition) Args() []Node { return []Node{d.Lhs, d.Rhs} }
func (d *LogicDefinition) Clone(args []Node) Node {
	return &LogicDefinition{Lhs: args[0].(Expr), Rhs: args[1].(Expr)}
}
