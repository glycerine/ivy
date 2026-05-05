package ivylogic

import (
	"fmt"
	"strings"

	lg "github.com/glycerine/ivy/goivy/logic"
)

// Sexp() methods for ivylogic types that implement logic.Expr.

func (s *Some) Sexp() lg.NodeKey {
	params := make([]string, len(s.Params))
	for i, p := range s.Params {
		params[i] = string(p.Sexp())
	}
	ifVal := "nil"
	if s.IfVal != nil {
		ifVal = string(s.IfVal.Sexp())
	}
	elseVal := "nil"
	if s.ElseVal != nil {
		elseVal = string(s.ElseVal.Sexp())
	}
	return lg.NodeKey("(Some params:[" + strings.Join(params, " ") + "] fmla:" + string(s.Fmla.Sexp()) + " ifVal:" + ifVal + " elseVal:" + elseVal + ")")
}

func (l *Let) Sexp() lg.NodeKey {
	defs := make([]string, len(l.Defs))
	for i, d := range l.Defs {
		defs[i] = string(d.Sexp())
	}
	return lg.NodeKey("(Let defs:[" + strings.Join(defs, " ") + "] body:" + string(l.Body.Sexp()) + ")")
}

func (lit *Literal) Sexp() lg.NodeKey {
	return lg.NodeKey(fmt.Sprintf("(Literal polarity:%d atom:%s)", lit.Polarity, lit.Atom.Sexp()))
}
