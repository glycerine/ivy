package goivy

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ParseIvyV16 parses a complete Ivy file using the v1.6 LALR grammar.
func ParseIvyV16(input string, version Version, opts ...ParseOption) (*ParseResult, error) {
	xtracer.Trace("parser.Parse ENTER")
	lex := newParser16LexAdapter(input, version)
	// Create a default AstConfig; WithAstConfig option will override it
	// if the caller provides a shared one (for nested parses).
	lex.astCfg = NewAstConfig()
	for _, opt := range opts {
		opt(lex)
	}
	parser16Parse(lex)
	if lex.err != "" {
		return nil, fmt.Errorf("LALR parse error: %s", lex.err)
	}
	if lex.accum == nil {
		xtracer.Trace("parser.Parse EXIT decls=0")
		return &ParseResult{}, nil
	}
	// Post-parse: expand autoinstances (matches Python's expand_autoinstances)
	// Only for top-level (non-nested) parses — Python: if not nested: expand_autoinstances(res)
	// Python calls expand_autoinstances(res) before reading res.decls,
	// so we expand before converting to ParseResult.
	if !lex.nested {
		expandAutoInstances(lex.accum)
	}
	result := lex.accum.toResult()
	xtracer.Trace("parser.Parse EXIT decls=%d", len(result.Decls))
	return result, nil
}

// --- parser16LexAdapter: adapter from lexer.Lexer to goyacc's parser16Lexer interface ---

type parser16LexAdapter struct {
	lex              *Lexer
	accum            *ivyAccum
	result           Node
	err              string
	importer         ImporterFunc
	included         map[string]bool
	nested           bool       // true for nested (include) parses — skip expand_auto
	lastTok          Token      // most recently returned token (lookahead), for line tracking
	prevTok          Token      // token before lastTok — the last token actually consumed
	filename         string     // source filename, matching Python's iu.filename
	specialAttribute string     // Python: global special_attribute — for "spec", "impl", "private"
	globalAttribute  string     // Python: global global_attribute — for "global"
	commonAttribute  string     // Python: global common_attribute — for "common"
	parentObjName    string     // Python: global parent_object — passed to newIvyAccum
	astCfg           *AstConfig // session-wide config (shared across nested parses)
}

func (l *parser16LexAdapter) setImporter(fn ImporterFunc)     { l.importer = fn }
func (l *parser16LexAdapter) setIncluded(inc map[string]bool) { l.included = inc }
func (l *parser16LexAdapter) setNested(nested bool)           { l.nested = nested }
func (l *parser16LexAdapter) setFilename(name string)         { l.filename = name }
func (l *parser16LexAdapter) setAstConfig(cfg *AstConfig)     { l.astCfg = cfg }

func newParser16LexAdapter(input string, version Version) *parser16LexAdapter {
	return &parser16LexAdapter{
		lex:      NewLexer(input, version),
		included: make(map[string]bool),
	}
}

// Lex returns the next token for goyacc.
func (l *parser16LexAdapter) Lex(lval *parser16SymType) int {
	tok := l.lex.NextToken()
	l.prevTok = l.lastTok
	l.lastTok = tok

	// Set tok on every token so line numbers are always available via $N.Line
	lval.tok = TokenInfo{Val: tok.Value, Line: tok.Line}

	switch tok.Type {
	case EOF:
		return 0
	case SYMBOL:
		return PARSER16_TOK_PRESYMBOL
	case VARIABLE:
		return PARSER16_TOK_VARIABLE
	case NATIVEQUOTE:
		return PARSER16_TOK_NATIVEQUOTE

	// Punctuation
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

	// Operators
	case PLUS:
		return PARSER16_TOK_PLUS
	case MINUS:
		return PARSER16_TOK_MINUS
	case TIMES:
		return PARSER16_TOK_TIMES
	case DIV:
		return PARSER16_TOK_DIV
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
	case CARET:
		return PARSER16_TOK_CARET
	case ASSIGN:
		return PARSER16_TOK_ASSIGN

	// Logic keywords
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
	case ISA:
		return PARSER16_TOK_ISA
	case IF:
		return PARSER16_TOK_IF
	case ELSE:
		return PARSER16_TOK_ELSE
	case GLOBALLY:
		return PARSER16_TOK_GLOBALLY
	case EVENTUALLY:
		return PARSER16_TOK_EVENTUALLY
	case WHENNEXT:
		return PARSER16_TOK_WHENNEXT
	case WHENPREV:
		return PARSER16_TOK_WHENPREV
	case WHENFIRST:
		return PARSER16_TOK_WHENFIRST
	case WHENLAST:
		return PARSER16_TOK_WHENLAST

	// Action keywords
	case ASSUME:
		return PARSER16_TOK_ASSUME
	case ASSERT:
		return PARSER16_TOK_ASSERT
	case REQUIRE:
		return PARSER16_TOK_REQUIRE
	case ENSURE:
		return PARSER16_TOK_ENSURE
	case VAR:
		return PARSER16_TOK_VAR
	case LOCAL:
		return PARSER16_TOK_LOCAL
	case LET:
		return PARSER16_TOK_LET
	case CALL:
		return PARSER16_TOK_CALL
	case WHILE:
		return PARSER16_TOK_WHILE
	case FOR:
		return PARSER16_TOK_FOR
	case IN:
		return PARSER16_TOK_IN
	case INVARIANT:
		return PARSER16_TOK_INVARIANT
	case DECREASES:
		return PARSER16_TOK_DECREASES
	case RETURNS:
		return PARSER16_TOK_RETURNS
	case SOME:
		return PARSER16_TOK_SOME
	case MINIMIZING:
		return PARSER16_TOK_MINIMIZING
	case MAXIMIZING:
		return PARSER16_TOK_MAXIMIZING
	case DEBUG:
		return PARSER16_TOK_DEBUG
	case THUNK:
		return PARSER16_TOK_THUNK
	case UNPROVABLE:
		return PARSER16_TOK_UNPROVABLE
	case PROOF:
		return PARSER16_TOK_PROOF
	case INSTANTIATE:
		return PARSER16_TOK_INSTANTIATE

	// Declaration keywords
	case RELATION:
		return PARSER16_TOK_RELATION
	case INDIV:
		return PARSER16_TOK_INDIV
	case FUNCTION:
		return PARSER16_TOK_FUNCTION
	case DERIVED:
		return PARSER16_TOK_DERIVED
	case AXIOM:
		return PARSER16_TOK_AXIOM
	case CONJECTURE:
		return PARSER16_TOK_CONJECTURE
	case SCHEMA:
		return PARSER16_TOK_SCHEMA
	case THEOREM:
		return PARSER16_TOK_THEOREM
	case PROPERTY:
		return PARSER16_TOK_PROPERTY
	case DEFINITION:
		return PARSER16_TOK_DEFINITION
	case TYPE:
		return PARSER16_TOK_TYPE
	case STRUCT:
		return PARSER16_TOK_STRUCT
	case MODULE:
		return PARSER16_TOK_MODULE
	case OBJECT:
		return PARSER16_TOK_OBJECT
	case CLASS:
		return PARSER16_TOK_CLASS
	case SUBCLASS:
		return PARSER16_TOK_SUBCLASS
	case ACTION:
		return PARSER16_TOK_ACTION
	case METHOD:
		return PARSER16_TOK_METHOD
	case BEFORE:
		return PARSER16_TOK_BEFORE
	case AFTER:
		return PARSER16_TOK_AFTER
	case AROUND:
		return PARSER16_TOK_AROUND
	case MIXIN:
		return PARSER16_TOK_MIXIN
	case IMPLEMENT:
		return PARSER16_TOK_IMPLEMENT
	case ISOLATE:
		return PARSER16_TOK_ISOLATE
	case EXTRACT:
		return PARSER16_TOK_EXTRACT
	case TRUSTED:
		return PARSER16_TOK_TRUSTED
	case EXPORT:
		return PARSER16_TOK_EXPORT
	case IMPORT:
		return PARSER16_TOK_IMPORT
	case DELEGATE:
		return PARSER16_TOK_DELEGATE
	case USING:
		return PARSER16_TOK_USING
	case INCLUDE:
		return PARSER16_TOK_INCLUDE
	case INTERPRET:
		return PARSER16_TOK_INTERPRET
	case MACRO:
		return PARSER16_TOK_MACRO
	case ALIAS:
		return PARSER16_TOK_ALIAS
	case ATTRIBUTE:
		return PARSER16_TOK_ATTRIBUTE
	case VARIANT:
		return PARSER16_TOK_VARIANT
	case OF:
		return PARSER16_TOK_OF
	case SCENARIO:
		return PARSER16_TOK_SCENARIO
	case PROGRESS:
		return PARSER16_TOK_PROGRESS
	case RELY:
		return PARSER16_TOK_RELY
	case MIXORD:
		return PARSER16_TOK_MIXORD
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
	case MODIFIES:
		return PARSER16_TOK_MODIFIES
	case ENSURES:
		return PARSER16_TOK_ENSURES
	case REQUIRES:
		return PARSER16_TOK_REQUIRES
	case INIT:
		return PARSER16_TOK_INIT
	case ENTRY:
		return PARSER16_TOK_ENTRY
	case SET:
		return PARSER16_TOK_SET
	case NULL:
		return PARSER16_TOK_NULL
	case MATCH:
		return PARSER16_TOK_MATCH
	case FRESH:
		return PARSER16_TOK_FRESH
	case NAMED:
		return PARSER16_TOK_NAMED
	case TEMPORAL:
		return PARSER16_TOK_TEMPORAL
	case EXPLICIT:
		return PARSER16_TOK_EXPLICIT
	case SPECIFICATION:
		return PARSER16_TOK_SPECIFICATION
	case IMPLEMENTATION:
		return PARSER16_TOK_IMPLEMENTATION
	case PRIVATE:
		return PARSER16_TOK_PRIVATE
	case GLOBAL:
		return PARSER16_TOK_GLOBAL
	case COMMON:
		return PARSER16_TOK_COMMON
	case GHOST:
		return PARSER16_TOK_GHOST
	case FINITE:
		return PARSER16_TOK_FINITE
	case PARAMETER:
		return PARSER16_TOK_PARAMETER
	case DESTRUCTOR:
		return PARSER16_TOK_DESTRUCTOR
	case CONSTRUCTOR:
		return PARSER16_TOK_CONSTRUCTOR
	case FIELD:
		return PARSER16_TOK_FIELD
	case AUTOINSTANCE:
		return PARSER16_TOK_AUTOINSTANCE
	case WITH:
		return PARSER16_TOK_WITH

	// Proof/tactic tokens
	case TACTIC:
		return PARSER16_TOK_TACTIC
	case TRIGGER:
		return PARSER16_TOK_TRIGGER
	case SHOWGOALS:
		return PARSER16_TOK_SHOWGOALS
	case DEFERGOAL:
		return PARSER16_TOK_DEFERGOAL
	case SPOIL:
		return PARSER16_TOK_SPOIL
	case UNFOLD:
		return PARSER16_TOK_UNFOLD
	case FORGET:
		return PARSER16_TOK_FORGET
	case APPLY:
		return PARSER16_TOK_APPLY

	default:
		lval.str = tok.Value
		return PARSER16_TOK_PRESYMBOL
	}
}

// Error is called by goyacc when a parse error occurs.
func (l *parser16LexAdapter) Error(s string) {
	l.err = s
}
