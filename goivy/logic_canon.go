package goivy

// Canon() implementations for all logic types.
// These delegate to the existing Sexp() methods via type conversion.

// --- Sort types ---

func (s *UninterpretedSort) Canon() Canonical { return Canonical(s.Sexp()) }
func (s *BooleanSort) Canon() Canonical       { return Canonical(s.Sexp()) }
func (s *FunctionSort) Canon() Canonical      { return Canonical(s.Sexp()) }
func (s *EnumeratedSort) Canon() Canonical    { return Canonical(s.Sexp()) }
func (s *RangeSort) Canon() Canonical         { return Canonical(s.Sexp()) }
func (s *TopSort) Canon() Canonical           { return Canonical(s.Sexp()) }

// --- Term types ---

func (v *Variable) Canon() Canonical { return Canonical(v.Sexp()) }
func (c *Const) Canon() Canonical    { return Canonical(c.Sexp()) }
func (a *Apply) Canon() Canonical    { return Canonical(a.Sexp()) }

// --- Formula types ---

func (e *Eq) Canon() Canonical           { return Canonical(e.Sexp()) }
func (n *Not) Canon() Canonical          { return Canonical(n.Sexp()) }
func (a *And) Canon() Canonical          { return Canonical(a.Sexp()) }
func (o *Or) Canon() Canonical           { return Canonical(o.Sexp()) }
func (i *Implies) Canon() Canonical      { return Canonical(i.Sexp()) }
func (i *Iff) Canon() Canonical          { return Canonical(i.Sexp()) }
func (t *Ite) Canon() Canonical          { return Canonical(t.Sexp()) }
func (g *Globally) Canon() Canonical     { return Canonical(g.Sexp()) }
func (e *Eventually) Canon() Canonical   { return Canonical(e.Sexp()) }
func (w *WhenOperator) Canon() Canonical { return Canonical(w.Sexp()) }
func (c *Cond) Canon() Canonical         { return Canonical(c.Sexp()) }
func (f *ForAll) Canon() Canonical       { return Canonical(f.Sexp()) }
func (e *Exists) Canon() Canonical       { return Canonical(e.Sexp()) }
func (l *Lambda) Canon() Canonical       { return Canonical(l.Sexp()) }
func (nb *NamedBinder) Canon() Canonical { return Canonical(nb.Sexp()) }

// --- Definition types ---

func (d *Definition) Canon() Canonical        { return Canonical(d.Sexp()) }
func (ds *DefinitionSchema) Canon() Canonical { return Canonical(ds.Sexp()) }

// --- NativeExpr ---

func (n *NativeExpr) Canon() Canonical { return Canonical(n.Sexp()) }
