package goivy

import (
	"github.com/glycerine/ivy/goivy/xtracer"
)

// ParseFullV16 parses a complete Ivy file using the Ivy <=1.6 full-file grammar.
func ParseFullV16(input string, version Version, opts ...ParseOption) (*ParseResult, error) {
	xtracer.Trace("parser.Parse ENTER")
	cfg := newParseConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.setLanguageVersion(version)
	lex := newParser16LexAdapter(input, version)
	cfg.applyToParser16(lex)
	parser16Parse(lex)
	if lex.err != nil {
		return nil, lex.err
	}
	if lex.accum == nil {
		xtracer.Trace("parser.Parse EXIT decls=0")
		return &ParseResult{}, nil
	}
	if !lex.nested {
		expandAutoInstances(lex.accum)
	}
	if len(lex.accum.errors) > 0 {
		return nil, lex.accum.errors[0]
	}
	result := lex.accum.toResult()
	xtracer.Trace("parser.Parse EXIT decls=%d", len(result.Decls))
	return result, nil
}

type parser16LexAdapter struct {
	lex           *Lexer
	accum         *ivyAccum
	err           *ParseError
	importer      ImporterFunc
	included      map[string]bool
	nested        bool
	lastTok       Token
	prevTok       Token
	filename      string
	parentObjName string
	astCfg        *AstConfig
}

func newParser16LexAdapter(input string, version Version) *parser16LexAdapter {
	return &parser16LexAdapter{
		lex:      NewLexer(input, version),
		included: make(map[string]bool),
	}
}

func (l *parser16LexAdapter) Lex(lval *parser16SymType) int {
	tok := l.lex.NextToken()
	l.prevTok = l.lastTok
	l.lastTok = tok

	lval.tok = TokenInfo{Val: tok.Value, Line: tok.Line}

	switch tok.Type {
	case EOF:
		return 0
	case SYMBOL:
		return PARSER16_TOK_PRESYMBOL
	case VARIABLE:
		return PARSER16_TOK_VARIABLE
	case LPAREN:
		return PARSER16_TOK_LPAREN
	case RPAREN:
		return PARSER16_TOK_RPAREN
	case LB:
		return PARSER16_TOK_LB
	case RB:
		return PARSER16_TOK_RB
	case LCB:
		return PARSER16_TOK_LCB
	case RCB:
		return PARSER16_TOK_RCB
	case COMMA:
		return PARSER16_TOK_COMMA
	case SEMI:
		return PARSER16_TOK_SEMI
	case COLON:
		return PARSER16_TOK_COLON
	case DOT:
		return PARSER16_TOK_DOT
	case DOTS:
		return PARSER16_TOK_DOTS
	case DOTDOTDOT:
		return PARSER16_TOK_DOTDOTDOT
	case PLUS:
		return PARSER16_TOK_PLUS
	case MINUS:
		return PARSER16_TOK_MINUS
	case TIMES:
		return PARSER16_TOK_TIMES
	case DIV:
		return PARSER16_TOK_DIV
	case NATIVEQUOTE:
		return PARSER16_TOK_NATIVEQUOTE
	case EQ:
		return PARSER16_TOK_EQ
	case TILDAEQ:
		return PARSER16_TOK_TILDAEQ
	case TILDA:
		return PARSER16_TOK_TILDA
	case LE:
		return PARSER16_TOK_LE
	case LT:
		return PARSER16_TOK_LT
	case GE:
		return PARSER16_TOK_GE
	case GT:
		return PARSER16_TOK_GT
	case AND:
		return PARSER16_TOK_AND
	case OR:
		return PARSER16_TOK_OR
	case ARROW:
		return PARSER16_TOK_ARROW
	case IFF:
		return PARSER16_TOK_IFF
	case PTO:
		return PARSER16_TOK_PTO
	case DOLLAR:
		return PARSER16_TOK_DOLLAR
	case ASSIGN:
		return PARSER16_TOK_ASSIGN
	case FORALL:
		return PARSER16_TOK_FORALL
	case EXISTS:
		return PARSER16_TOK_EXISTS
	case TRUE:
		return PARSER16_TOK_TRUE
	case FALSE:
		return PARSER16_TOK_FALSE
	case OLD:
		return PARSER16_TOK_OLD
	case THIS:
		return PARSER16_TOK_THIS
	case IF:
		return PARSER16_TOK_IF
	case ELSE:
		return PARSER16_TOK_ELSE
	case WHILE:
		return PARSER16_TOK_WHILE
	case INVARIANT:
		return PARSER16_TOK_INVARIANT
	case GLOBALLY:
		return PARSER16_TOK_GLOBALLY
	case EVENTUALLY:
		return PARSER16_TOK_EVENTUALLY
	case TEMPORAL:
		return PARSER16_TOK_TEMPORAL
	case EXPLICIT:
		return PARSER16_TOK_EXPLICIT
	case SET:
		return PARSER16_TOK_SET
	case INSTANTIATE:
		return PARSER16_TOK_INSTANTIATE
	case LOCAL:
		return PARSER16_TOK_LOCAL
	case LET:
		return PARSER16_TOK_LET
	case IN:
		return PARSER16_TOK_IN
	case SOME:
		return PARSER16_TOK_SOME
	case MINIMIZING:
		return PARSER16_TOK_MINIMIZING
	case MAXIMIZING:
		return PARSER16_TOK_MAXIMIZING
	case DECREASES:
		return PARSER16_TOK_DECREASES
	case ASSUME:
		return PARSER16_TOK_ASSUME
	case RETURNS:
		return PARSER16_TOK_RETURNS
	case MODULE:
		return PARSER16_TOK_MODULE
	case CLASS:
		return PARSER16_TOK_CLASS
	case TYPE:
		return PARSER16_TOK_TYPE
	case OBJECT:
		return PARSER16_TOK_OBJECT
	case INDIV:
		return PARSER16_TOK_INDIV
	case VAR:
		return PARSER16_TOK_VAR
	case FUNCTION:
		return PARSER16_TOK_FUNCTION
	case RELATION:
		return PARSER16_TOK_RELATION
	case DERIVED:
		return PARSER16_TOK_DERIVED
	case STRUCT:
		return PARSER16_TOK_STRUCT
	case GHOST:
		return PARSER16_TOK_GHOST
	case FINITE:
		return PARSER16_TOK_FINITE
	case FRESH:
		return PARSER16_TOK_FRESH
	case DESTRUCTOR:
		return PARSER16_TOK_DESTRUCTOR
	case AXIOM:
		return PARSER16_TOK_AXIOM
	case PROPERTY:
		return PARSER16_TOK_PROPERTY
	case CONJECTURE:
		return PARSER16_TOK_CONJECTURE
	case SCHEMA:
		return PARSER16_TOK_SCHEMA
	case ASSERT:
		return PARSER16_TOK_ASSERT
	case DEFINITION:
		return PARSER16_TOK_DEFINITION
	case PROOF:
		return PARSER16_TOK_PROOF
	case WITH:
		return PARSER16_TOK_WITH
	case INIT:
		return PARSER16_TOK_INIT
	case INCLUDE:
		return PARSER16_TOK_INCLUDE
	case USING:
		return PARSER16_TOK_USING
	case ACTION:
		return PARSER16_TOK_ACTION
	case METHOD:
		return PARSER16_TOK_METHOD
	case CALL:
		return PARSER16_TOK_CALL
	case ENSURES:
		return PARSER16_TOK_ENSURES
	case MIXIN:
		return PARSER16_TOK_MIXIN
	case BEFORE:
		return PARSER16_TOK_BEFORE
	case AFTER:
		return PARSER16_TOK_AFTER
	case IMPLEMENT:
		return PARSER16_TOK_IMPLEMENT
	case TRUSTED:
		return PARSER16_TOK_TRUSTED
	case ISOLATE:
		return PARSER16_TOK_ISOLATE
	case EXTRACT:
		return PARSER16_TOK_EXTRACT
	case DELEGATE:
		return PARSER16_TOK_DELEGATE
	case INTERPRET:
		return PARSER16_TOK_INTERPRET
	case CONCEPT:
		return PARSER16_TOK_CONCEPT
	case STATE:
		return PARSER16_TOK_STATE
	case UPDATE:
		return PARSER16_TOK_UPDATE
	case FROM:
		return PARSER16_TOK_FROM
	case PARAMS:
		return PARSER16_TOK_PARAMS
	case ATTRIBUTE:
		return PARSER16_TOK_ATTRIBUTE
	case VARIANT:
		return PARSER16_TOK_VARIANT
	case OF:
		return PARSER16_TOK_OF
	case SCENARIO:
		return PARSER16_TOK_SCENARIO
	case REQUIRES:
		return PARSER16_TOK_REQUIRES
	case MODIFIES:
		return PARSER16_TOK_MODIFIES
	case ENTRY:
		return PARSER16_TOK_ENTRY
	case IMPORT:
		return PARSER16_TOK_IMPORT
	case EXPORT:
		return PARSER16_TOK_EXPORT
	case PRIVATE:
		return PARSER16_TOK_PRIVATE
	case MACRO:
		return PARSER16_TOK_MACRO
	case ALIAS:
		return PARSER16_TOK_ALIAS
	case PROGRESS:
		return PARSER16_TOK_PROGRESS
	case RELY:
		return PARSER16_TOK_RELY
	case MIXORD:
		return PARSER16_TOK_MIXORD
	case NAMED:
		return PARSER16_TOK_NAMED
	default:
		lval.tok = TokenInfo{Val: tok.Value, Line: tok.Line}
		return PARSER16_TOK_PRESYMBOL
	}
}

func (l *parser16LexAdapter) Error(s string) {
	pe := &ParseError{Filename: l.filename, Message: s}
	if l.lastTok.Type == EOF {
		pe.Message = "unexpected end of input"
	} else {
		pe.Lineno = l.lastTok.Line
		pe.Token = l.lastTok.Value
	}
	l.err = pe
}
