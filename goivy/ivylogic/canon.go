package ivylogic

import iu "github.com/glycerine/ivy/goivy/ivyutils"

// Canon() methods for ivylogic types, wrapping Sexp() — same pattern as logic/canon.go.

func (s *Some) Canon() iu.Canonical      { return iu.Canonical(s.Sexp()) }
func (l *Let) Canon() iu.Canonical       { return iu.Canonical(l.Sexp()) }
func (lit *Literal) Canon() iu.Canonical { return iu.Canonical(lit.Sexp()) }
