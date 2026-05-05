package goivy

import (
	"fmt"
)

// ParseV16 parses a formula string using the v1.3–v1.6 LALR grammar.
func ParseV16(input string, version Version, cfg ...*AstConfig) (Node, error) {
	var c *AstConfig
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
	lex    *Lexer
	cfg    *AstConfig
	result Node
	err    string
}

func newLalr16LexAdapter(input string, version Version, cfg *AstConfig) *lalr16LexAdapter {
	if cfg == nil {
		cfg = NewAstConfig()
	}
	return &lalr16LexAdapter{
		lex: NewLexer(input, version),
		cfg: cfg,
	}
}

func (l *lalr16LexAdapter) Lex(lval *lalr16SymType) int {
	tok := l.lex.NextToken()
	switch tok.Type {
	case EOF:
		return 0
	case SYMBOL:
		lval.str = tok.Value
		return LALR16_TOK_PRESYMBOL
	case VARIABLE:
		lval.str = tok.Value
		return LALR16_TOK_VARIABLE
	case LPAREN:
		return LALR16_TOK_LPAREN
	case RPAREN:
		return LALR16_TOK_RPAREN
	case LB:
		return LALR16_TOK_LB
	case RB:
		return LALR16_TOK_RB
	case LCB:
		return LALR16_TOK_LCB
	case RCB:
		return LALR16_TOK_RCB
	case COMMA:
		return LALR16_TOK_COMMA
	case SEMI:
		return LALR16_TOK_SEMI
	case COLON:
		return LALR16_TOK_COLON
	case DOT:
		return LALR16_TOK_DOT
	case PLUS:
		return LALR16_TOK_PLUS
	case MINUS:
		return LALR16_TOK_MINUS
	case TIMES:
		return LALR16_TOK_TIMES
	case DIV:
		return LALR16_TOK_DIV
	case EQ:
		return LALR16_TOK_EQ
	case TILDAEQ:
		return LALR16_TOK_TILDAEQ
	case TILDA:
		return LALR16_TOK_TILDA
	case LE:
		return LALR16_TOK_LE
	case LT:
		return LALR16_TOK_LT
	case GE:
		return LALR16_TOK_GE
	case GT:
		return LALR16_TOK_GT
	case AND:
		return LALR16_TOK_AND
	case OR:
		return LALR16_TOK_OR
	case ARROW:
		return LALR16_TOK_ARROW
	case IFF:
		return LALR16_TOK_IFF
	case PTO:
		return LALR16_TOK_PTO
	case DOLLAR:
		return LALR16_TOK_DOLLAR
	case FORALL:
		return LALR16_TOK_FORALL
	case EXISTS:
		return LALR16_TOK_EXISTS
	case TRUE:
		return LALR16_TOK_TRUE
	case FALSE:
		return LALR16_TOK_FALSE
	case OLD:
		return LALR16_TOK_OLD
	case THIS:
		return LALR16_TOK_THIS
	case IF:
		return LALR16_TOK_IF
	case ELSE:
		return LALR16_TOK_ELSE
	case GLOBALLY:
		return LALR16_TOK_GLOBALLY
	case EVENTUALLY:
		return LALR16_TOK_EVENTUALLY
	case WHENNEXT:
		return LALR16_TOK_WHENNEXT
	case WHENPREV:
		return LALR16_TOK_WHENPREV
	case WHENFIRST:
		return LALR16_TOK_WHENFIRST
	case WHENLAST:
		return LALR16_TOK_WHENLAST
	default:
		lval.str = tok.Value
		return LALR16_TOK_PRESYMBOL
	}
}

func (l *lalr16LexAdapter) Error(s string) { l.err = s }
