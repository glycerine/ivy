// ast_compat.go provides Args() and Clone() methods so that all logic.Expr
// concrete types also satisfy ast.Node. Since logic.Expr embeds ast.Node,
// every Expr value IS an ast.Node — no adapters or conversion helpers needed.
package logic

import "github.com/glycerine/goivy/ast"

// --- Sort types ---

func (s *UninterpretedSort) Args() []ast.Node          { return nil }
func (s *UninterpretedSort) Clone([]ast.Node) ast.Node  { return s }

func (s *BooleanSort) Args() []ast.Node          { return nil }
func (s *BooleanSort) Clone([]ast.Node) ast.Node  { return s }

func (s *FunctionSort) Args() []ast.Node {
	r := make([]ast.Node, len(s.Sorts))
	for i, sub := range s.Sorts {
		r[i] = sub
	}
	return r
}
func (s *FunctionSort) Clone(args []ast.Node) ast.Node {
	sorts := make([]Sort, len(args))
	for i, a := range args {
		sorts[i] = a.(Sort)
	}
	return &FunctionSort{Sorts: sorts}
}

func (s *EnumeratedSort) Args() []ast.Node          { return nil }
func (s *EnumeratedSort) Clone([]ast.Node) ast.Node  { return s }

func (s *TopSort) Args() []ast.Node          { return nil }
func (s *TopSort) Clone([]ast.Node) ast.Node  { return s }

func (s *RangeSort) Args() []ast.Node          { return nil }
func (s *RangeSort) Clone([]ast.Node) ast.Node  { return s }

// --- Term types ---

func (v *Variable) Args() []ast.Node          { return nil }
func (v *Variable) Clone([]ast.Node) ast.Node  { return v }

func (c *Symbol) Args() []ast.Node          { return nil }
func (c *Symbol) Clone([]ast.Node) ast.Node  { return c }

func (a *Apply) Args() []ast.Node {
	r := make([]ast.Node, len(a.Terms))
	for i, t := range a.Terms {
		r[i] = t
	}
	return r
}
func (a *Apply) Clone(args []ast.Node) ast.Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &Apply{Func: a.Func, Terms: terms, aSort: a.aSort}
}

// --- Formula types ---

func (e *Eq) Args() []ast.Node { return []ast.Node{e.T1, e.T2} }
func (e *Eq) Clone(args []ast.Node) ast.Node {
	return &Eq{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (t *Ite) Args() []ast.Node { return []ast.Node{t.Cond, t.Then, t.Else} }
func (t *Ite) Clone(args []ast.Node) ast.Node {
	return &Ite{ISort: t.ISort, Cond: args[0].(Expr), Then: args[1].(Expr), Else: args[2].(Expr)}
}

func (n *Not) Args() []ast.Node { return []ast.Node{n.Body} }
func (n *Not) Clone(args []ast.Node) ast.Node {
	return &Not{Body: args[0].(Expr)}
}

func (g *Globally) Args() []ast.Node { return []ast.Node{g.Body} }
func (g *Globally) Clone(args []ast.Node) ast.Node {
	return &Globally{Environ: g.Environ, Body: args[0].(Expr)}
}

func (e *Eventually) Args() []ast.Node { return []ast.Node{e.Body} }
func (e *Eventually) Clone(args []ast.Node) ast.Node {
	return &Eventually{Environ: e.Environ, Body: args[0].(Expr)}
}

func (w *WhenOperator) Args() []ast.Node { return []ast.Node{w.T1, w.T2} }
func (w *WhenOperator) Clone(args []ast.Node) ast.Node {
	return &WhenOperator{WSort: w.WSort, Name: w.Name, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (c *Cond) Args() []ast.Node { return []ast.Node{c.T1, c.T2} }
func (c *Cond) Clone(args []ast.Node) ast.Node {
	return &Cond{CSort: c.CSort, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (a *And) Args() []ast.Node {
	r := make([]ast.Node, len(a.Terms))
	for i, t := range a.Terms {
		r[i] = t
	}
	return r
}
func (a *And) Clone(args []ast.Node) ast.Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &And{Terms: terms}
}

func (o *Or) Args() []ast.Node {
	r := make([]ast.Node, len(o.Terms))
	for i, t := range o.Terms {
		r[i] = t
	}
	return r
}
func (o *Or) Clone(args []ast.Node) ast.Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &Or{Terms: terms}
}

func (im *Implies) Args() []ast.Node { return []ast.Node{im.T1, im.T2} }
func (im *Implies) Clone(args []ast.Node) ast.Node {
	return &Implies{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (i *Iff) Args() []ast.Node { return []ast.Node{i.T1, i.T2} }
func (i *Iff) Clone(args []ast.Node) ast.Node {
	return &Iff{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (f *ForAll) Args() []ast.Node { return []ast.Node{f.Body} }
func (f *ForAll) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(f.Variables))
	copy(cp, f.Variables)
	return &ForAll{Variables: cp, Body: args[0].(Expr)}
}

func (e *Exists) Args() []ast.Node { return []ast.Node{e.Body} }
func (e *Exists) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(e.Variables))
	copy(cp, e.Variables)
	return &Exists{Variables: cp, Body: args[0].(Expr)}
}

func (l *Lambda) Args() []ast.Node { return []ast.Node{l.Body} }
func (l *Lambda) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(l.Variables))
	copy(cp, l.Variables)
	return &Lambda{Variables: cp, Body: args[0].(Expr)}
}

func (nb *NamedBinder) Args() []ast.Node { return []ast.Node{nb.Body} }
func (nb *NamedBinder) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(nb.Variables))
	copy(cp, nb.Variables)
	return &NamedBinder{Name: nb.Name, Variables: cp, Environ: nb.Environ, Body: args[0].(Expr)}
}

// --- Definition ---

func (d *Definition) Args() []ast.Node { return []ast.Node{d.Lhs, d.Rhs} }
func (d *Definition) Clone(args []ast.Node) ast.Node {
	return &Definition{Lhs: args[0].(Expr), Rhs: args[1].(Expr)}
}
