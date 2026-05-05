package v12

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
)

// ParseV12 parses a formula string using the v1.2 (and earlier) LALR grammar.
func ParseV12(input string, version lexer.Version, cfg ...*ast.AstConfig) (ast.Node, error) {
	var c *ast.AstConfig
	if len(cfg) > 0 {
		c = cfg[0]
	}
	lex := newV12LexAdapter(input, version, c)
	v12Parse(lex)
	if lex.err != "" {
		return nil, fmt.Errorf("LALR v1.2 parse error: %s", lex.err)
	}
	return lex.result, nil
}

type v12LexAdapter struct {
	lex    *lexer.Lexer
	cfg    *ast.AstConfig
	result ast.Node
	err    string
}

func newV12LexAdapter(input string, version lexer.Version, cfg *ast.AstConfig) *v12LexAdapter {
	if cfg == nil {
		cfg = ast.NewAstConfig()
	}
	return &v12LexAdapter{
		lex: lexer.New(input, version),
		cfg: cfg,
	}
}

func (l *v12LexAdapter) Lex(lval *v12SymType) int {
	tok := l.lex.NextToken()
	switch tok.Type {
	case lexer.EOF:
		return 0
	case lexer.SYMBOL:
		lval.str = tok.Value
		return TOK_PRESYMBOL
	case lexer.VARIABLE:
		lval.str = tok.Value
		return TOK_VARIABLE
	case lexer.LPAREN:
		return TOK_LPAREN
	case lexer.RPAREN:
		return TOK_RPAREN
	case lexer.LB:
		return TOK_LB
	case lexer.RB:
		return TOK_RB
	case lexer.LCB:
		return TOK_LCB
	case lexer.RCB:
		return TOK_RCB
	case lexer.COMMA:
		return TOK_COMMA
	case lexer.SEMI:
		return TOK_SEMI
	case lexer.COLON:
		return TOK_COLON
	case lexer.DOT:
		return TOK_DOT
	case lexer.PLUS:
		return TOK_PLUS
	case lexer.MINUS:
		return TOK_MINUS
	case lexer.TIMES:
		return TOK_TIMES
	case lexer.DIV:
		return TOK_DIV
	case lexer.EQ:
		return TOK_EQ
	case lexer.TILDAEQ:
		return TOK_TILDAEQ
	case lexer.TILDA:
		return TOK_TILDA
	case lexer.LE:
		return TOK_LE
	case lexer.LT:
		return TOK_LT
	case lexer.GE:
		return TOK_GE
	case lexer.GT:
		return TOK_GT
	case lexer.AND:
		return TOK_AND
	case lexer.OR:
		return TOK_OR
	case lexer.ARROW:
		return TOK_ARROW
	case lexer.IFF:
		return TOK_IFF
	case lexer.PTO:
		return TOK_PTO
	case lexer.DOLLAR:
		return TOK_DOLLAR
	case lexer.FORALL:
		return TOK_FORALL
	case lexer.EXISTS:
		return TOK_EXISTS
	case lexer.TRUE:
		return TOK_TRUE
	case lexer.FALSE:
		return TOK_FALSE
	case lexer.OLD:
		return TOK_OLD
	case lexer.THIS:
		return TOK_THIS
	case lexer.IF:
		return TOK_IF
	case lexer.ELSE:
		return TOK_ELSE
	case lexer.GLOBALLY:
		return TOK_GLOBALLY
	case lexer.EVENTUALLY:
		return TOK_EVENTUALLY
	case lexer.WHENNEXT:
		return TOK_WHENNEXT
	case lexer.WHENPREV:
		return TOK_WHENPREV
	case lexer.WHENFIRST:
		return TOK_WHENFIRST
	case lexer.WHENLAST:
		return TOK_WHENLAST
	default:
		lval.str = tok.Value
		return TOK_PRESYMBOL
	}
}

func (l *v12LexAdapter) Error(s string) { l.err = s }
