package ivylogic

import (
	"fmt"
	"strings"
)

// Sexp() methods for ivylogic types that implement logic.Expr.

func (s *Some) Sexp() string {
	params := make([]string, len(s.Params))
	for i, p := range s.Params {
		params[i] = p.Sexp()
	}
	ifVal := "nil"
	if s.IfVal != nil {
		ifVal = s.IfVal.Sexp()
	}
	elseVal := "nil"
	if s.ElseVal != nil {
		elseVal = s.ElseVal.Sexp()
	}
	return "(Some params:[" + strings.Join(params, " ") + "] fmla:" + s.Fmla.Sexp() + " ifVal:" + ifVal + " elseVal:" + elseVal + ")"
}

func (l *Let) Sexp() string {
	defs := make([]string, len(l.Defs))
	for i, d := range l.Defs {
		defs[i] = d.Sexp()
	}
	return "(Let defs:[" + strings.Join(defs, " ") + "] body:" + l.Body.Sexp() + ")"
}

func (lit *Literal) Sexp() string {
	return fmt.Sprintf("(Literal polarity:%d atom:%s)", lit.Polarity, lit.Atom.Sexp())
}
