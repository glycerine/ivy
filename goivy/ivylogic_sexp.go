package goivy

import (
	"fmt"
	"strings"
)

// Sexp() methods for ivylogic types that implement logic.Expr.

func (s *LogicSome) Sexp() NodeKey {
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
	return NodeKey("(Some params:[" + strings.Join(params, " ") + "] fmla:" + string(s.Fmla.Sexp()) + " ifVal:" + ifVal + " elseVal:" + elseVal + ")")
}

func (l *LogicLet) Sexp() NodeKey {
	defs := make([]string, len(l.Defs))
	for i, d := range l.Defs {
		defs[i] = string(d.Sexp())
	}
	return NodeKey("(Let defs:[" + strings.Join(defs, " ") + "] body:" + string(l.Body.Sexp()) + ")")
}

func (lit *LogicLiteral) Sexp() NodeKey {
	return NodeKey(fmt.Sprintf("(Literal polarity:%d atom:%s)", lit.Polarity, lit.Atom.Sexp()))
}
