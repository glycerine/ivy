// ast_compat.go provides Args() and Clone() methods so that all logic.Expr
// concrete types also satisfy ast.Node. This bridges the two interface
// hierarchies without adapter wrappers.
package logic

import "github.com/glycerine/goivy/ast"

// exprToAstNode converts an Expr to ast.Node via type assertion.
// Safe because all concrete Expr types satisfy ast.Node after this file.
func exprToAstNode(e Expr) ast.Node {
	if isNil(e) {
		return nil
	}
	return e.(ast.Node)
}

// exprsToAstNodes converts a slice of Expr to []ast.Node.
func exprsToAstNodes(es []Expr) []ast.Node {
	r := make([]ast.Node, len(es))
	for i, e := range es {
		r[i] = exprToAstNode(e)
	}
	return r
}

// sortsToAstNodes converts a slice of Sort to []ast.Node.
func sortsToAstNodes(ss []Sort) []ast.Node {
	r := make([]ast.Node, len(ss))
	for i, s := range ss {
		r[i] = s.(ast.Node)
	}
	return r
}

// --- Sort types ---

func (s *UninterpretedSort) Args() []ast.Node          { return nil }
func (s *UninterpretedSort) Clone([]ast.Node) ast.Node  { return s }

func (s *BooleanSort) Args() []ast.Node          { return nil }
func (s *BooleanSort) Clone([]ast.Node) ast.Node  { return s }

func (s *FunctionSort) Args() []ast.Node           { return sortsToAstNodes(s.Sorts) }
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

func (a *Apply) Args() []ast.Node              { return exprsToAstNodes(a.Terms) }
func (a *Apply) Clone(args []ast.Node) ast.Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &Apply{Func: a.Func, Terms: terms, aSort: a.aSort}
}

// --- Formula types ---

func (e *Eq) Args() []ast.Node { return []ast.Node{exprToAstNode(e.T1), exprToAstNode(e.T2)} }
func (e *Eq) Clone(args []ast.Node) ast.Node {
	return &Eq{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (t *Ite) Args() []ast.Node {
	return []ast.Node{exprToAstNode(t.Cond), exprToAstNode(t.Then), exprToAstNode(t.Else)}
}
func (t *Ite) Clone(args []ast.Node) ast.Node {
	return &Ite{ISort: t.ISort, Cond: args[0].(Expr), Then: args[1].(Expr), Else: args[2].(Expr)}
}

func (n *Not) Args() []ast.Node { return []ast.Node{exprToAstNode(n.Body)} }
func (n *Not) Clone(args []ast.Node) ast.Node {
	return &Not{Body: args[0].(Expr)}
}

func (g *Globally) Args() []ast.Node { return []ast.Node{exprToAstNode(g.Body)} }
func (g *Globally) Clone(args []ast.Node) ast.Node {
	return &Globally{Environ: g.Environ, Body: args[0].(Expr)}
}

func (e *Eventually) Args() []ast.Node { return []ast.Node{exprToAstNode(e.Body)} }
func (e *Eventually) Clone(args []ast.Node) ast.Node {
	return &Eventually{Environ: e.Environ, Body: args[0].(Expr)}
}

func (w *WhenOperator) Args() []ast.Node {
	return []ast.Node{exprToAstNode(w.T1), exprToAstNode(w.T2)}
}
func (w *WhenOperator) Clone(args []ast.Node) ast.Node {
	return &WhenOperator{WSort: w.WSort, Name: w.Name, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (c *Cond) Args() []ast.Node { return []ast.Node{exprToAstNode(c.T1), exprToAstNode(c.T2)} }
func (c *Cond) Clone(args []ast.Node) ast.Node {
	return &Cond{CSort: c.CSort, T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (a *And) Args() []ast.Node              { return exprsToAstNodes(a.Terms) }
func (a *And) Clone(args []ast.Node) ast.Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &And{Terms: terms}
}

func (o *Or) Args() []ast.Node              { return exprsToAstNodes(o.Terms) }
func (o *Or) Clone(args []ast.Node) ast.Node {
	terms := make([]Expr, len(args))
	for i, arg := range args {
		terms[i] = arg.(Expr)
	}
	return &Or{Terms: terms}
}

func (im *Implies) Args() []ast.Node {
	return []ast.Node{exprToAstNode(im.T1), exprToAstNode(im.T2)}
}
func (im *Implies) Clone(args []ast.Node) ast.Node {
	return &Implies{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (i *Iff) Args() []ast.Node { return []ast.Node{exprToAstNode(i.T1), exprToAstNode(i.T2)} }
func (i *Iff) Clone(args []ast.Node) ast.Node {
	return &Iff{T1: args[0].(Expr), T2: args[1].(Expr)}
}

func (f *ForAll) Args() []ast.Node { return []ast.Node{exprToAstNode(f.Body)} }
func (f *ForAll) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(f.Variables))
	copy(cp, f.Variables)
	return &ForAll{Variables: cp, Body: args[0].(Expr)}
}

func (e *Exists) Args() []ast.Node { return []ast.Node{exprToAstNode(e.Body)} }
func (e *Exists) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(e.Variables))
	copy(cp, e.Variables)
	return &Exists{Variables: cp, Body: args[0].(Expr)}
}

func (l *Lambda) Args() []ast.Node { return []ast.Node{exprToAstNode(l.Body)} }
func (l *Lambda) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(l.Variables))
	copy(cp, l.Variables)
	return &Lambda{Variables: cp, Body: args[0].(Expr)}
}

func (nb *NamedBinder) Args() []ast.Node { return []ast.Node{exprToAstNode(nb.Body)} }
func (nb *NamedBinder) Clone(args []ast.Node) ast.Node {
	cp := make([]*Variable, len(nb.Variables))
	copy(cp, nb.Variables)
	return &NamedBinder{Name: nb.Name, Variables: cp, Environ: nb.Environ, Body: args[0].(Expr)}
}

// --- Definition ---

func (d *Definition) Args() []ast.Node {
	return []ast.Node{exprToAstNode(d.Lhs), exprToAstNode(d.Rhs)}
}
func (d *Definition) Clone(args []ast.Node) ast.Node {
	return &Definition{Lhs: args[0].(Expr), Rhs: args[1].(Expr)}
}

// --- Interface satisfaction compile-time checks ---

var (
	_ ast.Node = (*UninterpretedSort)(nil)
	_ ast.Node = (*BooleanSort)(nil)
	_ ast.Node = (*FunctionSort)(nil)
	_ ast.Node = (*EnumeratedSort)(nil)
	_ ast.Node = (*TopSort)(nil)
	_ ast.Node = (*RangeSort)(nil)
	_ ast.Node = (*Variable)(nil)
	_ ast.Node = (*Symbol)(nil)
	_ ast.Node = (*Apply)(nil)
	_ ast.Node = (*Eq)(nil)
	_ ast.Node = (*Ite)(nil)
	_ ast.Node = (*Not)(nil)
	_ ast.Node = (*Globally)(nil)
	_ ast.Node = (*Eventually)(nil)
	_ ast.Node = (*WhenOperator)(nil)
	_ ast.Node = (*Cond)(nil)
	_ ast.Node = (*And)(nil)
	_ ast.Node = (*Or)(nil)
	_ ast.Node = (*Implies)(nil)
	_ ast.Node = (*Iff)(nil)
	_ ast.Node = (*ForAll)(nil)
	_ ast.Node = (*Exists)(nil)
	_ ast.Node = (*Lambda)(nil)
	_ ast.Node = (*NamedBinder)(nil)
	_ ast.Node = (*Definition)(nil)
)
