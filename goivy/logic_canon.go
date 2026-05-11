package goivy

import "io"

// Canon() implementations for all logic types.
// These delegate to the existing Sexp() methods via type conversion.

// --- Sort types ---

func (s *UninterpretedSort) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, s) })
}
func (s *BooleanSort) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, s) })
}
func (s *LogicFunctionSort) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, s) })
}
func (s *LogicEnumeratedSort) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, s) })
}
func (s *RangeSort) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, s) })
}
func (s *TopSort) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, s) })
}

// --- Term types ---

func (v *LogicVariable) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, v) })
}
func (c *Const) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, c) })
}
func (a *Apply) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, a) })
}

// --- Formula types ---

func (e *Eq) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, e) })
}
func (n *LogicNot) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, n) })
}
func (a *LogicAnd) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, a) })
}
func (o *LogicOr) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, o) })
}
func (i *LogicImplies) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, i) })
}
func (i *LogicIff) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, i) })
}
func (t *LogicIte) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, t) })
}
func (g *LogicGlobally) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, g) })
}
func (e *LogicEventually) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, e) })
}
func (w *LogicWhenOperator) Canon() Canonical {
	return canonString(func(ww io.Writer) { writeExprCanon(ww, w) })
}
func (c *Cond) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, c) })
}
func (f *ForAll) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, f) })
}
func (e *LogicExists) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, e) })
}
func (l *Lambda) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, l) })
}
func (nb *LogicNamedBinder) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, nb) })
}

// --- Definition types ---

func (d *LogicDefinition) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, d) })
}
func (ds *LogicDefinitionSchema) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, ds) })
}

// --- NativeExpr ---

func (n *LogicNativeExpr) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, n) })
}
