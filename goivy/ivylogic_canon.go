package goivy

// Canon() methods for ivylogic types, wrapping Sexp() — same pattern as logic/canon.go.

func (s *LogicSome) Canon() Canonical      { return Canonical(s.Sexp()) }
func (l *LogicLet) Canon() Canonical       { return Canonical(l.Sexp()) }
func (lit *LogicLiteral) Canon() Canonical { return Canonical(lit.Sexp()) }
