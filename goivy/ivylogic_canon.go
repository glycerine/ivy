package goivy

// Canon() methods for ivylogic types, wrapping Sexp() — same pattern as logic/canon.go.

func (s *Some) Canon() Canonical      { return Canonical(s.Sexp()) }
func (l *Let) Canon() Canonical       { return Canonical(l.Sexp()) }
func (lit *Literal) Canon() Canonical { return Canonical(lit.Sexp()) }
