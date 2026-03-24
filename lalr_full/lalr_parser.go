package lalr_full

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/xtracer"
)

// Parse parses a complete Ivy file and returns the declarations.
// It dispatches to the appropriate version-specific LALR parser.
// The optional importer callback resolves `include` directives.
func Parse(input string, version lexer.Version, opts ...ParseOption) (*ParseResult, error) {
	// For now, only v1.7+ is implemented
	return ParseV17(input, version, opts...)
}

// ParseOption configures optional behavior for the LALR parser.
type ParseOption func(*v17LexAdapter)

// WithImporter sets the include-resolution callback.
func WithImporter(fn ImporterFunc) ParseOption {
	return func(lex *v17LexAdapter) {
		lex.importer = fn
	}
}

// WithIncluded sets the already-included module set (for nested parses).
func WithIncluded(inc map[string]bool) ParseOption {
	return func(lex *v17LexAdapter) {
		lex.included = inc
	}
}

// WithNested marks this as a nested (include) parse.
// Nested parses skip expand_autoinstances, matching Python behavior.
func WithNested() ParseOption {
	return func(lex *v17LexAdapter) {
		lex.nested = true
	}
}

// WithFilename sets the source filename, matching Python's iu.filename.
// Used by getLineno to produce Location with filename for canon matching.
func WithFilename(name string) ParseOption {
	return func(lex *v17LexAdapter) {
		lex.filename = name
	}
}

// ParseV17 parses a complete Ivy file using the v1.7+ LALR grammar.
func ParseV17(input string, version lexer.Version, opts ...ParseOption) (*ParseResult, error) {
	xtracer.Trace("parser.Parse ENTER")
	lalrLabelCounter = 0
	lex := newV17LexAdapter(input, version)
	for _, opt := range opts {
		opt(lex)
	}
	v17Parse(lex)
	if lex.err != "" {
		return nil, fmt.Errorf("LALR parse error: %s", lex.err)
	}
	if lex.accum == nil {
		xtracer.Trace("parser.Parse EXIT decls=0")
		return &ParseResult{}, nil
	}
	result := lex.accum.toResult()
	// Post-parse: expand autoinstances (matches Python's expand_autoinstances)
	// Only for top-level (non-nested) parses — Python: if not nested: expand_autoinstances(res)
	if !lex.nested {
		result.Decls = expandAutoInstances(lex.accum, result.Decls)
	}
	xtracer.Trace("parser.Parse EXIT decls=%d", len(result.Decls))
	return result, nil
}

// --- v17LexAdapter: adapter from lexer.Lexer to goyacc's v17Lexer interface ---

// ImporterFunc is the callback for resolving `include` directives.
// It takes a module name and returns the parsed declarations.
// Matches Python's ivy_parser.importer function.
type ImporterFunc func(name string) (*ParseResult, error)

type v17LexAdapter struct {
	lex      *lexer.Lexer
	accum    *ivyAccum
	result   ast.Node
	err      string
	importer ImporterFunc
	included map[string]bool
	nested   bool        // true for nested (include) parses — skip expand_auto
	lastTok  lexer.Token // most recently returned token, for line tracking
	filename string      // source filename, matching Python's iu.filename
}

func newV17LexAdapter(input string, version lexer.Version) *v17LexAdapter {
	return &v17LexAdapter{
		lex:      lexer.New(input, version),
		included: make(map[string]bool),
	}
}

// Lex returns the next token for goyacc.
func (l *v17LexAdapter) Lex(lval *v17SymType) int {
	tok := l.lex.NextToken()
	l.lastTok = tok

	switch tok.Type {
	case lexer.EOF:
		return 0
	case lexer.SYMBOL:
		lval.str = tok.Value
		return TOK_PRESYMBOL
	case lexer.VARIABLE:
		lval.str = tok.Value
		return TOK_VARIABLE
	case lexer.NATIVEQUOTE:
		lval.str = tok.Value
		return TOK_NATIVEQUOTE

	// Punctuation
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
	case lexer.DOTS:
		return TOK_DOTS
	case lexer.DOTDOTDOT:
		return TOK_DOTDOTDOT

	// Operators
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
	case lexer.CARET:
		return TOK_CARET
	case lexer.ASSIGN:
		return TOK_ASSIGN

	// Logic keywords
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
	case lexer.ISA:
		return TOK_ISA
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

	// Action keywords
	case lexer.ASSUME:
		return TOK_ASSUME
	case lexer.ASSERT:
		return TOK_ASSERT
	case lexer.REQUIRE:
		return TOK_REQUIRE
	case lexer.ENSURE:
		return TOK_ENSURE
	case lexer.VAR:
		return TOK_VAR
	case lexer.LOCAL:
		return TOK_LOCAL
	case lexer.LET:
		return TOK_LET
	case lexer.CALL:
		return TOK_CALL
	case lexer.WHILE:
		return TOK_WHILE
	case lexer.FOR:
		return TOK_FOR
	case lexer.IN:
		return TOK_IN
	case lexer.INVARIANT:
		return TOK_INVARIANT
	case lexer.DECREASES:
		return TOK_DECREASES
	case lexer.RETURNS:
		return TOK_RETURNS
	case lexer.SOME:
		return TOK_SOME
	case lexer.MINIMIZING:
		return TOK_MINIMIZING
	case lexer.MAXIMIZING:
		return TOK_MAXIMIZING
	case lexer.DEBUG:
		return TOK_DEBUG
	case lexer.THUNK:
		return TOK_THUNK
	case lexer.UNPROVABLE:
		return TOK_UNPROVABLE
	case lexer.PROOF:
		return TOK_PROOF
	case lexer.INSTANTIATE:
		return TOK_INSTANTIATE

	// Declaration keywords
	case lexer.RELATION:
		return TOK_RELATION
	case lexer.INDIV:
		return TOK_INDIV
	case lexer.FUNCTION:
		return TOK_FUNCTION
	case lexer.DERIVED:
		return TOK_DERIVED
	case lexer.AXIOM:
		return TOK_AXIOM
	case lexer.CONJECTURE:
		return TOK_CONJECTURE
	case lexer.SCHEMA:
		return TOK_SCHEMA
	case lexer.THEOREM:
		return TOK_THEOREM
	case lexer.PROPERTY:
		return TOK_PROPERTY
	case lexer.DEFINITION:
		return TOK_DEFINITION
	case lexer.TYPE:
		return TOK_TYPE
	case lexer.STRUCT:
		return TOK_STRUCT
	case lexer.MODULE:
		return TOK_MODULE
	case lexer.OBJECT:
		return TOK_OBJECT
	case lexer.CLASS:
		return TOK_CLASS
	case lexer.SUBCLASS:
		return TOK_SUBCLASS
	case lexer.ACTION:
		return TOK_ACTION
	case lexer.METHOD:
		return TOK_METHOD
	case lexer.BEFORE:
		return TOK_BEFORE
	case lexer.AFTER:
		return TOK_AFTER
	case lexer.AROUND:
		return TOK_AROUND
	case lexer.MIXIN:
		return TOK_MIXIN
	case lexer.IMPLEMENT:
		return TOK_IMPLEMENT
	case lexer.ISOLATE:
		return TOK_ISOLATE
	case lexer.EXTRACT:
		return TOK_EXTRACT
	case lexer.TRUSTED:
		return TOK_TRUSTED
	case lexer.EXPORT:
		return TOK_EXPORT
	case lexer.IMPORT:
		return TOK_IMPORT
	case lexer.DELEGATE:
		return TOK_DELEGATE
	case lexer.USING:
		return TOK_USING
	case lexer.INCLUDE:
		return TOK_INCLUDE
	case lexer.INTERPRET:
		return TOK_INTERPRET
	case lexer.MACRO:
		return TOK_MACRO
	case lexer.ALIAS:
		return TOK_ALIAS
	case lexer.ATTRIBUTE:
		return TOK_ATTRIBUTE
	case lexer.VARIANT:
		return TOK_VARIANT
	case lexer.OF:
		return TOK_OF
	case lexer.SCENARIO:
		return TOK_SCENARIO
	case lexer.PROGRESS:
		return TOK_PROGRESS
	case lexer.RELY:
		return TOK_RELY
	case lexer.MIXORD:
		return TOK_MIXORD
	case lexer.CONCEPT:
		return TOK_CONCEPT
	case lexer.STATE:
		return TOK_STATE
	case lexer.UPDATE:
		return TOK_UPDATE
	case lexer.FROM:
		return TOK_FROM
	case lexer.PARAMS:
		return TOK_PARAMS
	case lexer.MODIFIES:
		return TOK_MODIFIES
	case lexer.ENSURES:
		return TOK_ENSURES
	case lexer.REQUIRES:
		return TOK_REQUIRES
	case lexer.INIT:
		return TOK_INIT
	case lexer.ENTRY:
		return TOK_ENTRY
	case lexer.SET:
		return TOK_SET
	case lexer.NULL:
		return TOK_NULL
	case lexer.MATCH:
		return TOK_MATCH
	case lexer.FRESH:
		return TOK_FRESH
	case lexer.NAMED:
		return TOK_NAMED
	case lexer.TEMPORAL:
		return TOK_TEMPORAL
	case lexer.EXPLICIT:
		return TOK_EXPLICIT
	case lexer.SPECIFICATION:
		return TOK_SPECIFICATION
	case lexer.IMPLEMENTATION:
		return TOK_IMPLEMENTATION
	case lexer.PRIVATE:
		return TOK_PRIVATE
	case lexer.GLOBAL:
		return TOK_GLOBAL
	case lexer.COMMON:
		return TOK_COMMON
	case lexer.GHOST:
		return TOK_GHOST
	case lexer.FINITE:
		return TOK_FINITE
	case lexer.PARAMETER:
		return TOK_PARAMETER
	case lexer.DESTRUCTOR:
		return TOK_DESTRUCTOR
	case lexer.CONSTRUCTOR:
		return TOK_CONSTRUCTOR
	case lexer.FIELD:
		return TOK_FIELD
	case lexer.AUTOINSTANCE:
		return TOK_AUTOINSTANCE
	case lexer.WITH:
		return TOK_WITH

	// Proof/tactic tokens
	case lexer.TACTIC:
		return TOK_TACTIC
	case lexer.TRIGGER:
		return TOK_TRIGGER
	case lexer.SHOWGOALS:
		return TOK_SHOWGOALS
	case lexer.DEFERGOAL:
		return TOK_DEFERGOAL
	case lexer.SPOIL:
		return TOK_SPOIL
	case lexer.UNFOLD:
		return TOK_UNFOLD
	case lexer.FORGET:
		return TOK_FORGET
	case lexer.APPLY:
		return TOK_APPLY

	default:
		lval.str = tok.Value
		return TOK_PRESYMBOL
	}
}

// Error is called by goyacc when a parse error occurs.
func (l *v17LexAdapter) Error(s string) {
	l.err = s
}
