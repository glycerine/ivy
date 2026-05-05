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
	lex := newLalr12LexAdapter(input, version, c)
	lalr12Parse(lex)
	if lex.err != "" {
		return nil, fmt.Errorf("LALR v1.2 parse error: %s", lex.err)
	}
	return lex.result, nil
}

type lalr12LexAdapter struct {
	lex    *lexer.Lexer
	cfg    *ast.AstConfig
	result ast.Node
	err    string
}

func newLalr12LexAdapter(input string, version lexer.Version, cfg *ast.AstConfig) *lalr12LexAdapter {
	if cfg == nil {
		cfg = ast.NewAstConfig()
	}
	return &lalr12LexAdapter{
		lex: lexer.New(input, version),
		cfg: cfg,
	}
}

func (l *lalr12LexAdapter) Lex(lval *lalr12SymType) int {
	tok := l.lex.NextToken()
	switch tok.Type {
	case lexer.EOF:
		return 0
	case lexer.SYMBOL:
		lval.str = tok.Value
		return LALR12_TOK_PRESYMBOL
	case lexer.VARIABLE:
		lval.str = tok.Value
		return LALR12_TOK_VARIABLE
	case lexer.LPAREN:
		return LALR12_TOK_LPAREN
	case lexer.RPAREN:
		return LALR12_TOK_RPAREN
	case lexer.LB:
		return LALR12_TOK_LB
	case lexer.RB:
		return LALR12_TOK_RB
	case lexer.LCB:
		return LALR12_TOK_LCB
	case lexer.RCB:
		return LALR12_TOK_RCB
	case lexer.COMMA:
		return LALR12_TOK_COMMA
	case lexer.SEMI:
		return LALR12_TOK_SEMI
	case lexer.COLON:
		return LALR12_TOK_COLON
	case lexer.DOT:
		return LALR12_TOK_DOT
	case lexer.PLUS:
		return LALR12_TOK_PLUS
	case lexer.MINUS:
		return LALR12_TOK_MINUS
	case lexer.TIMES:
		return LALR12_TOK_TIMES
	case lexer.DIV:
		return LALR12_TOK_DIV
	case lexer.EQ:
		return LALR12_TOK_EQ
	case lexer.TILDAEQ:
		return LALR12_TOK_TILDAEQ
	case lexer.TILDA:
		return LALR12_TOK_TILDA
	case lexer.LE:
		return LALR12_TOK_LE
	case lexer.LT:
		return LALR12_TOK_LT
	case lexer.GE:
		return LALR12_TOK_GE
	case lexer.GT:
		return LALR12_TOK_GT
	case lexer.AND:
		return LALR12_TOK_AND
	case lexer.OR:
		return LALR12_TOK_OR
	case lexer.ARROW:
		return LALR12_TOK_ARROW
	case lexer.IFF:
		return LALR12_TOK_IFF
	case lexer.PTO:
		return LALR12_TOK_PTO
	case lexer.DOLLAR:
		return LALR12_TOK_DOLLAR
	case lexer.FORALL:
		return LALR12_TOK_FORALL
	case lexer.EXISTS:
		return LALR12_TOK_EXISTS
	case lexer.TRUE:
		return LALR12_TOK_TRUE
	case lexer.FALSE:
		return LALR12_TOK_FALSE
	case lexer.OLD:
		return LALR12_TOK_OLD
	case lexer.THIS:
		return LALR12_TOK_THIS
	case lexer.IF:
		return LALR12_TOK_IF
	case lexer.ELSE:
		return LALR12_TOK_ELSE
	case lexer.GLOBALLY:
		return LALR12_TOK_GLOBALLY
	case lexer.EVENTUALLY:
		return LALR12_TOK_EVENTUALLY
	case lexer.WHENNEXT:
		return LALR12_TOK_WHENNEXT
	case lexer.WHENPREV:
		return LALR12_TOK_WHENPREV
	case lexer.WHENFIRST:
		return LALR12_TOK_WHENFIRST
	case lexer.WHENLAST:
		return LALR12_TOK_WHENLAST
	default:
		lval.str = tok.Value
		return LALR12_TOK_PRESYMBOL
	}
}

func (l *lalr12LexAdapter) Error(s string) { l.err = s }
