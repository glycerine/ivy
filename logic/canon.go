package logic

import (
	iu "github.com/glycerine/goivy/ivyutils"
)

// Canon() implementations for all logic types.
// These delegate to the existing Sexp() methods via type conversion.

// --- Sort types ---

func (s *UninterpretedSort) Canon() iu.Canonical { return iu.Canonical(s.Sexp()) }
func (s *BooleanSort) Canon() iu.Canonical       { return iu.Canonical(s.Sexp()) }
func (s *FunctionSort) Canon() iu.Canonical       { return iu.Canonical(s.Sexp()) }
func (s *EnumeratedSort) Canon() iu.Canonical     { return iu.Canonical(s.Sexp()) }
func (s *RangeSort) Canon() iu.Canonical          { return iu.Canonical(s.Sexp()) }
func (s *TopSort) Canon() iu.Canonical            { return iu.Canonical(s.Sexp()) }

// --- Term types ---

func (v *Variable) Canon() iu.Canonical { return iu.Canonical(v.Sexp()) }
func (c *Symbol) Canon() iu.Canonical   { return iu.Canonical(c.Sexp()) }
func (a *Apply) Canon() iu.Canonical    { return iu.Canonical(a.Sexp()) }

// --- Formula types ---

func (e *Eq) Canon() iu.Canonical           { return iu.Canonical(e.Sexp()) }
func (n *Not) Canon() iu.Canonical          { return iu.Canonical(n.Sexp()) }
func (a *And) Canon() iu.Canonical          { return iu.Canonical(a.Sexp()) }
func (o *Or) Canon() iu.Canonical           { return iu.Canonical(o.Sexp()) }
func (i *Implies) Canon() iu.Canonical      { return iu.Canonical(i.Sexp()) }
func (i *Iff) Canon() iu.Canonical          { return iu.Canonical(i.Sexp()) }
func (t *Ite) Canon() iu.Canonical          { return iu.Canonical(t.Sexp()) }
func (g *Globally) Canon() iu.Canonical     { return iu.Canonical(g.Sexp()) }
func (e *Eventually) Canon() iu.Canonical   { return iu.Canonical(e.Sexp()) }
func (w *WhenOperator) Canon() iu.Canonical { return iu.Canonical(w.Sexp()) }
func (c *Cond) Canon() iu.Canonical         { return iu.Canonical(c.Sexp()) }
func (f *ForAll) Canon() iu.Canonical       { return iu.Canonical(f.Sexp()) }
func (e *Exists) Canon() iu.Canonical       { return iu.Canonical(e.Sexp()) }
func (l *Lambda) Canon() iu.Canonical       { return iu.Canonical(l.Sexp()) }
func (nb *NamedBinder) Canon() iu.Canonical { return iu.Canonical(nb.Sexp()) }

// --- Definition types ---

func (d *Definition) Canon() iu.Canonical       { return iu.Canonical(d.Sexp()) }
func (ds *DefinitionSchema) Canon() iu.Canonical { return iu.Canonical(ds.Sexp()) }

// --- NativeExpr ---

func (n *NativeExpr) Canon() iu.Canonical { return iu.Canonical(n.Sexp()) }
