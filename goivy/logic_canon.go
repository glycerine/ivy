package goivy

// Canon() implementations for all logic types.
// These delegate to the existing Sexp() methods via type conversion.

// --- Sort types ---

func (s *UninterpretedSort) Canon() Canonical   { return Canonical(s.Sexp()) }
func (s *BooleanSort) Canon() Canonical         { return Canonical(s.Sexp()) }
func (s *LogicFunctionSort) Canon() Canonical   { return Canonical(s.Sexp()) }
func (s *LogicEnumeratedSort) Canon() Canonical { return Canonical(s.Sexp()) }
func (s *RangeSort) Canon() Canonical           { return Canonical(s.Sexp()) }
func (s *TopSort) Canon() Canonical             { return Canonical(s.Sexp()) }

// --- Term types ---

func (v *LogicVariable) Canon() Canonical { return Canonical(v.Sexp()) }
func (c *Const) Canon() Canonical         { return Canonical(c.Sexp()) }
func (a *Apply) Canon() Canonical         { return Canonical(a.Sexp()) }

// --- Formula types ---

func (e *Eq) Canon() Canonical                { return Canonical(e.Sexp()) }
func (n *LogicNot) Canon() Canonical          { return Canonical(n.Sexp()) }
func (a *LogicAnd) Canon() Canonical          { return Canonical(a.Sexp()) }
func (o *LogicOr) Canon() Canonical           { return Canonical(o.Sexp()) }
func (i *LogicImplies) Canon() Canonical      { return Canonical(i.Sexp()) }
func (i *LogicIff) Canon() Canonical          { return Canonical(i.Sexp()) }
func (t *LogicIte) Canon() Canonical          { return Canonical(t.Sexp()) }
func (g *LogicGlobally) Canon() Canonical     { return Canonical(g.Sexp()) }
func (e *LogicEventually) Canon() Canonical   { return Canonical(e.Sexp()) }
func (w *LogicWhenOperator) Canon() Canonical { return Canonical(w.Sexp()) }
func (c *Cond) Canon() Canonical              { return Canonical(c.Sexp()) }
func (f *ForAll) Canon() Canonical            { return Canonical(f.Sexp()) }
func (e *LogicExists) Canon() Canonical       { return Canonical(e.Sexp()) }
func (l *Lambda) Canon() Canonical            { return Canonical(l.Sexp()) }
func (nb *LogicNamedBinder) Canon() Canonical { return Canonical(nb.Sexp()) }

// --- Definition types ---

func (d *LogicDefinition) Canon() Canonical        { return Canonical(d.Sexp()) }
func (ds *LogicDefinitionSchema) Canon() Canonical { return Canonical(ds.Sexp()) }

// --- NativeExpr ---

func (n *LogicNativeExpr) Canon() Canonical { return Canonical(n.Sexp()) }
