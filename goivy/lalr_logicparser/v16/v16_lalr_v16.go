package v16

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
)

// ParseV16 parses a formula string using the v1.3–v1.6 LALR grammar.
func ParseV16(input string, version lexer.Version, cfg ...*ast.AstConfig) (ast.Node, error) {
	var c *ast.AstConfig
	if len(cfg) > 0 {
		c = cfg[0]
	}
	lex := newLalr16LexAdapter(input, version, c)
	lalr16Parse(lex)
	if lex.err != "" {
		return nil, fmt.Errorf("LALR v1.6 parse error: %s", lex.err)
	}
	return lex.result, nil
}

type lalr16LexAdapter struct {
	lex    *lexer.Lexer
	cfg    *ast.AstConfig
	result ast.Node
	err    string
}

func newLalr16LexAdapter(input string, version lexer.Version, cfg *ast.AstConfig) *lalr16LexAdapter {
	if cfg == nil {
		cfg = ast.NewAstConfig()
	}
	return &lalr16LexAdapter{
		lex: lexer.New(input, version),
		cfg: cfg,
	}
}

func (l *lalr16LexAdapter) Lex(lval *lalr16SymType) int {
	tok := l.lex.NextToken()
	switch tok.Type {
	case lexer.EOF:
		return 0
	case lexer.SYMBOL:
		lval.str = tok.Value
		return LALR16_TOK_PRESYMBOL
	case lexer.VARIABLE:
		lval.str = tok.Value
		return LALR16_TOK_VARIABLE
	case lexer.LPAREN:
		return LALR16_TOK_LPAREN
	case lexer.RPAREN:
		return LALR16_TOK_RPAREN
	case lexer.LB:
		return LALR16_TOK_LB
	case lexer.RB:
		return LALR16_TOK_RB
	case lexer.LCB:
		return LALR16_TOK_LCB
	case lexer.RCB:
		return LALR16_TOK_RCB
	case lexer.COMMA:
		return LALR16_TOK_COMMA
	case lexer.SEMI:
		return LALR16_TOK_SEMI
	case lexer.COLON:
		return LALR16_TOK_COLON
	case lexer.DOT:
		return LALR16_TOK_DOT
	case lexer.PLUS:
		return LALR16_TOK_PLUS
	case lexer.MINUS:
		return LALR16_TOK_MINUS
	case lexer.TIMES:
		return LALR16_TOK_TIMES
	case lexer.DIV:
		return LALR16_TOK_DIV
	case lexer.EQ:
		return LALR16_TOK_EQ
	case lexer.TILDAEQ:
		return LALR16_TOK_TILDAEQ
	case lexer.TILDA:
		return LALR16_TOK_TILDA
	case lexer.LE:
		return LALR16_TOK_LE
	case lexer.LT:
		return LALR16_TOK_LT
	case lexer.GE:
		return LALR16_TOK_GE
	case lexer.GT:
		return LALR16_TOK_GT
	case lexer.AND:
		return LALR16_TOK_AND
	case lexer.OR:
		return LALR16_TOK_OR
	case lexer.ARROW:
		return LALR16_TOK_ARROW
	case lexer.IFF:
		return LALR16_TOK_IFF
	case lexer.PTO:
		return LALR16_TOK_PTO
	case lexer.DOLLAR:
		return LALR16_TOK_DOLLAR
	case lexer.FORALL:
		return LALR16_TOK_FORALL
	case lexer.EXISTS:
		return LALR16_TOK_EXISTS
	case lexer.TRUE:
		return LALR16_TOK_TRUE
	case lexer.FALSE:
		return LALR16_TOK_FALSE
	case lexer.OLD:
		return LALR16_TOK_OLD
	case lexer.THIS:
		return LALR16_TOK_THIS
	case lexer.IF:
		return LALR16_TOK_IF
	case lexer.ELSE:
		return LALR16_TOK_ELSE
	case lexer.GLOBALLY:
		return LALR16_TOK_GLOBALLY
	case lexer.EVENTUALLY:
		return LALR16_TOK_EVENTUALLY
	case lexer.WHENNEXT:
		return LALR16_TOK_WHENNEXT
	case lexer.WHENPREV:
		return LALR16_TOK_WHENPREV
	case lexer.WHENFIRST:
		return LALR16_TOK_WHENFIRST
	case lexer.WHENLAST:
		return LALR16_TOK_WHENLAST
	default:
		lval.str = tok.Value
		return LALR16_TOK_PRESYMBOL
	}
}

func (l *lalr16LexAdapter) Error(s string) { l.err = s }
