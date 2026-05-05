package goivy

import (
	"fmt"
)

// ParseV12 parses a formula string using the v1.2 (and earlier) LALR grammar.
func ParseV12(input string, version Version, cfg ...*AstConfig) (Node, error) {
	var c *AstConfig
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
	lex    *Lexer
	cfg    *AstConfig
	result Node
	err    string
}

func newLalr12LexAdapter(input string, version Version, cfg *AstConfig) *lalr12LexAdapter {
	if cfg == nil {
		cfg = NewAstConfig()
	}
	return &lalr12LexAdapter{
		lex: NewLexer(input, version),
		cfg: cfg,
	}
}

func (l *lalr12LexAdapter) Lex(lval *lalr12SymType) int {
	tok := l.lex.NextToken()
	switch tok.Type {
	case EOF:
		return 0
	case SYMBOL:
		lval.str = tok.Value
		return LALR12_TOK_PRESYMBOL
	case VARIABLE:
		lval.str = tok.Value
		return LALR12_TOK_VARIABLE
	case LPAREN:
		return LALR12_TOK_LPAREN
	case RPAREN:
		return LALR12_TOK_RPAREN
	case LB:
		return LALR12_TOK_LB
	case RB:
		return LALR12_TOK_RB
	case LCB:
		return LALR12_TOK_LCB
	case RCB:
		return LALR12_TOK_RCB
	case COMMA:
		return LALR12_TOK_COMMA
	case SEMI:
		return LALR12_TOK_SEMI
	case COLON:
		return LALR12_TOK_COLON
	case DOT:
		return LALR12_TOK_DOT
	case PLUS:
		return LALR12_TOK_PLUS
	case MINUS:
		return LALR12_TOK_MINUS
	case TIMES:
		return LALR12_TOK_TIMES
	case DIV:
		return LALR12_TOK_DIV
	case EQ:
		return LALR12_TOK_EQ
	case TILDAEQ:
		return LALR12_TOK_TILDAEQ
	case TILDA:
		return LALR12_TOK_TILDA
	case LE:
		return LALR12_TOK_LE
	case LT:
		return LALR12_TOK_LT
	case GE:
		return LALR12_TOK_GE
	case GT:
		return LALR12_TOK_GT
	case AND:
		return LALR12_TOK_AND
	case OR:
		return LALR12_TOK_OR
	case ARROW:
		return LALR12_TOK_ARROW
	case IFF:
		return LALR12_TOK_IFF
	case PTO:
		return LALR12_TOK_PTO
	case DOLLAR:
		return LALR12_TOK_DOLLAR
	case FORALL:
		return LALR12_TOK_FORALL
	case EXISTS:
		return LALR12_TOK_EXISTS
	case TRUE:
		return LALR12_TOK_TRUE
	case FALSE:
		return LALR12_TOK_FALSE
	case OLD:
		return LALR12_TOK_OLD
	case THIS:
		return LALR12_TOK_THIS
	case IF:
		return LALR12_TOK_IF
	case ELSE:
		return LALR12_TOK_ELSE
	case GLOBALLY:
		return LALR12_TOK_GLOBALLY
	case EVENTUALLY:
		return LALR12_TOK_EVENTUALLY
	case WHENNEXT:
		return LALR12_TOK_WHENNEXT
	case WHENPREV:
		return LALR12_TOK_WHENPREV
	case WHENFIRST:
		return LALR12_TOK_WHENFIRST
	case WHENLAST:
		return LALR12_TOK_WHENLAST
	default:
		lval.str = tok.Value
		return LALR12_TOK_PRESYMBOL
	}
}

func (l *lalr12LexAdapter) Error(s string) { l.err = s }
