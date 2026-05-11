package goivy

import "io"

// Canon() methods for ivylogic types, wrapping Sexp() — same pattern as logic/canon.go.

func (s *LogicSome) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, s) })
}
func (l *LogicLet) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, l) })
}
func (lit *LogicLiteral) Canon() Canonical {
	return canonString(func(w io.Writer) { writeExprCanon(w, lit) })
}
