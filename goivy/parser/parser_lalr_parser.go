package parser

import (
	"fmt"

	"github.com/glycerine/ivy/goivy/ast"
	"github.com/glycerine/ivy/goivy/lexer"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// Parse parses a complete Ivy file and returns the declarations.
// It dispatches to the appropriate version-specific LALR parser.
// The optional importer callback resolves `include` directives.
func Parse(input string, version lexer.Version, opts ...ParseOption) (*ParseResult, error) {
	// For now, only v1.7+ is implemented
	return ParseV17(input, version, opts...)
}

// ParseOption configures optional behavior for the LALR parser.
type ParseOption func(*parser17LexAdapter)

// WithImporter sets the include-resolution callback.
func WithImporter(fn ImporterFunc) ParseOption {
	return func(lex *parser17LexAdapter) {
		lex.importer = fn
	}
}

// WithIncluded sets the already-included module set (for nested parses).
func WithIncluded(inc map[string]bool) ParseOption {
	return func(lex *parser17LexAdapter) {
		lex.included = inc
	}
}

// WithNested marks this as a nested (include) parse.
// Nested parses skip expand_autoinstances, matching Python behavior.
func WithNested() ParseOption {
	return func(lex *parser17LexAdapter) {
		lex.nested = true
	}
}

// WithFilename sets the source filename, matching Python's iu.filename.
// Used by getLineno to produce Location with filename for canon matching.
func WithFilename(name string) ParseOption {
	return func(lex *parser17LexAdapter) {
		lex.filename = name
	}
}

// WithAstConfig shares an existing AstConfig with this parse.
// Python uses module-level globals (lf_counter, always_clone_with_fresh_id,
// label_counter, check_unprovable) shared across all parses. This option
// ensures nested/imported parses share the parent's AstConfig so counters
// and flags stay in sync.
func WithAstConfig(cfg *ast.AstConfig) ParseOption {
	return func(lex *parser17LexAdapter) {
		lex.astCfg = cfg
	}
}

// ParseV17 parses a complete Ivy file using the v1.7+ LALR grammar.
func ParseV17(input string, version lexer.Version, opts ...ParseOption) (*ParseResult, error) {
	xtracer.Trace("parser.Parse ENTER")
	lex := newParser17LexAdapter(input, version)
	// Create a default AstConfig; WithAstConfig option will override it
	// if the caller provides a shared one (for nested parses).
	lex.astCfg = ast.NewAstConfig()
	for _, opt := range opts {
		opt(lex)
	}
	parser17Parse(lex)
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

// --- parser17LexAdapter: adapter from lexer.Lexer to goyacc's parser17Lexer interface ---

// ImporterFunc is the callback for resolving `include` directives.
// It takes a module name and returns the parsed declarations.
// Matches Python's ivy_parser.importer function.
type ImporterFunc func(name string) (*ParseResult, error)

type parser17LexAdapter struct {
	lex              *lexer.Lexer
	accum            *ivyAccum
	result           ast.Node
	err              string
	importer         ImporterFunc
	included         map[string]bool
	nested           bool           // true for nested (include) parses — skip expand_auto
	lastTok          lexer.Token    // most recently returned token (lookahead), for line tracking
	prevTok          lexer.Token    // token before lastTok — the last token actually consumed
	filename         string         // source filename, matching Python's iu.filename
	specialAttribute string         // Python: global special_attribute — for "spec", "impl", "private"
	globalAttribute  string         // Python: global global_attribute — for "global"
	commonAttribute  string         // Python: global common_attribute — for "common"
	parentObjName    string         // Python: global parent_object — passed to newIvyAccum
	astCfg           *ast.AstConfig // session-wide config (shared across nested parses)
}

func newParser17LexAdapter(input string, version lexer.Version) *parser17LexAdapter {
	return &parser17LexAdapter{
		lex:      lexer.NewLexer(input, version),
		included: make(map[string]bool),
	}
}

// Lex returns the next token for goyacc.
func (l *parser17LexAdapter) Lex(lval *parser17SymType) int {
	tok := l.lex.NextToken()
	l.prevTok = l.lastTok
	l.lastTok = tok

	// Set tok on every token so line numbers are always available via $N.Line
	lval.tok = TokenInfo{Val: tok.Value, Line: tok.Line}

	switch tok.Type {
	case lexer.EOF:
		return 0
	case lexer.SYMBOL:
		return PARSER_TOK_PRESYMBOL
	case lexer.VARIABLE:
		return PARSER_TOK_VARIABLE
	case lexer.NATIVEQUOTE:
		return PARSER_TOK_NATIVEQUOTE

	// Punctuation
	case lexer.LPAREN:
		return PARSER_TOK_LPAREN
	case lexer.RPAREN:
		return PARSER_TOK_RPAREN
	case lexer.LB:
		return PARSER_TOK_LB
	case lexer.RB:
		return PARSER_TOK_RB
	case lexer.LCB:
		return PARSER_TOK_LCB
	case lexer.RCB:
		return PARSER_TOK_RCB
	case lexer.COMMA:
		return PARSER_TOK_COMMA
	case lexer.SEMI:
		return PARSER_TOK_SEMI
	case lexer.COLON:
		return PARSER_TOK_COLON
	case lexer.DOT:
		return PARSER_TOK_DOT
	case lexer.DOTS:
		return PARSER_TOK_DOTS
	case lexer.DOTDOTDOT:
		return PARSER_TOK_DOTDOTDOT

	// Operators
	case lexer.PLUS:
		return PARSER_TOK_PLUS
	case lexer.MINUS:
		return PARSER_TOK_MINUS
	case lexer.TIMES:
		return PARSER_TOK_TIMES
	case lexer.DIV:
		return PARSER_TOK_DIV
	case lexer.EQ:
		return PARSER_TOK_EQ
	case lexer.TILDAEQ:
		return PARSER_TOK_TILDAEQ
	case lexer.TILDA:
		return PARSER_TOK_TILDA
	case lexer.LE:
		return PARSER_TOK_LE
	case lexer.LT:
		return PARSER_TOK_LT
	case lexer.GE:
		return PARSER_TOK_GE
	case lexer.GT:
		return PARSER_TOK_GT
	case lexer.AND:
		return PARSER_TOK_AND
	case lexer.OR:
		return PARSER_TOK_OR
	case lexer.ARROW:
		return PARSER_TOK_ARROW
	case lexer.IFF:
		return PARSER_TOK_IFF
	case lexer.PTO:
		return PARSER_TOK_PTO
	case lexer.DOLLAR:
		return PARSER_TOK_DOLLAR
	case lexer.CARET:
		return PARSER_TOK_CARET
	case lexer.ASSIGN:
		return PARSER_TOK_ASSIGN

	// Logic keywords
	case lexer.FORALL:
		return PARSER_TOK_FORALL
	case lexer.EXISTS:
		return PARSER_TOK_EXISTS
	case lexer.TRUE:
		return PARSER_TOK_TRUE
	case lexer.FALSE:
		return PARSER_TOK_FALSE
	case lexer.OLD:
		return PARSER_TOK_OLD
	case lexer.THIS:
		return PARSER_TOK_THIS
	case lexer.ISA:
		return PARSER_TOK_ISA
	case lexer.IF:
		return PARSER_TOK_IF
	case lexer.ELSE:
		return PARSER_TOK_ELSE
	case lexer.GLOBALLY:
		return PARSER_TOK_GLOBALLY
	case lexer.EVENTUALLY:
		return PARSER_TOK_EVENTUALLY
	case lexer.WHENNEXT:
		return PARSER_TOK_WHENNEXT
	case lexer.WHENPREV:
		return PARSER_TOK_WHENPREV
	case lexer.WHENFIRST:
		return PARSER_TOK_WHENFIRST
	case lexer.WHENLAST:
		return PARSER_TOK_WHENLAST

	// Action keywords
	case lexer.ASSUME:
		return PARSER_TOK_ASSUME
	case lexer.ASSERT:
		return PARSER_TOK_ASSERT
	case lexer.REQUIRE:
		return PARSER_TOK_REQUIRE
	case lexer.ENSURE:
		return PARSER_TOK_ENSURE
	case lexer.VAR:
		return PARSER_TOK_VAR
	case lexer.LOCAL:
		return PARSER_TOK_LOCAL
	case lexer.LET:
		return PARSER_TOK_LET
	case lexer.CALL:
		return PARSER_TOK_CALL
	case lexer.WHILE:
		return PARSER_TOK_WHILE
	case lexer.FOR:
		return PARSER_TOK_FOR
	case lexer.IN:
		return PARSER_TOK_IN
	case lexer.INVARIANT:
		return PARSER_TOK_INVARIANT
	case lexer.DECREASES:
		return PARSER_TOK_DECREASES
	case lexer.RETURNS:
		return PARSER_TOK_RETURNS
	case lexer.SOME:
		return PARSER_TOK_SOME
	case lexer.MINIMIZING:
		return PARSER_TOK_MINIMIZING
	case lexer.MAXIMIZING:
		return PARSER_TOK_MAXIMIZING
	case lexer.DEBUG:
		return PARSER_TOK_DEBUG
	case lexer.THUNK:
		return PARSER_TOK_THUNK
	case lexer.UNPROVABLE:
		return PARSER_TOK_UNPROVABLE
	case lexer.PROOF:
		return PARSER_TOK_PROOF
	case lexer.INSTANTIATE:
		return PARSER_TOK_INSTANTIATE

	// Declaration keywords
	case lexer.RELATION:
		return PARSER_TOK_RELATION
	case lexer.INDIV:
		return PARSER_TOK_INDIV
	case lexer.FUNCTION:
		return PARSER_TOK_FUNCTION
	case lexer.DERIVED:
		return PARSER_TOK_DERIVED
	case lexer.AXIOM:
		return PARSER_TOK_AXIOM
	case lexer.CONJECTURE:
		return PARSER_TOK_CONJECTURE
	case lexer.SCHEMA:
		return PARSER_TOK_SCHEMA
	case lexer.THEOREM:
		return PARSER_TOK_THEOREM
	case lexer.PROPERTY:
		return PARSER_TOK_PROPERTY
	case lexer.DEFINITION:
		return PARSER_TOK_DEFINITION
	case lexer.TYPE:
		return PARSER_TOK_TYPE
	case lexer.STRUCT:
		return PARSER_TOK_STRUCT
	case lexer.MODULE:
		return PARSER_TOK_MODULE
	case lexer.OBJECT:
		return PARSER_TOK_OBJECT
	case lexer.CLASS:
		return PARSER_TOK_CLASS
	case lexer.SUBCLASS:
		return PARSER_TOK_SUBCLASS
	case lexer.ACTION:
		return PARSER_TOK_ACTION
	case lexer.METHOD:
		return PARSER_TOK_METHOD
	case lexer.BEFORE:
		return PARSER_TOK_BEFORE
	case lexer.AFTER:
		return PARSER_TOK_AFTER
	case lexer.AROUND:
		return PARSER_TOK_AROUND
	case lexer.MIXIN:
		return PARSER_TOK_MIXIN
	case lexer.IMPLEMENT:
		return PARSER_TOK_IMPLEMENT
	case lexer.ISOLATE:
		return PARSER_TOK_ISOLATE
	case lexer.EXTRACT:
		return PARSER_TOK_EXTRACT
	case lexer.TRUSTED:
		return PARSER_TOK_TRUSTED
	case lexer.EXPORT:
		return PARSER_TOK_EXPORT
	case lexer.IMPORT:
		return PARSER_TOK_IMPORT
	case lexer.DELEGATE:
		return PARSER_TOK_DELEGATE
	case lexer.USING:
		return PARSER_TOK_USING
	case lexer.INCLUDE:
		return PARSER_TOK_INCLUDE
	case lexer.INTERPRET:
		return PARSER_TOK_INTERPRET
	case lexer.MACRO:
		return PARSER_TOK_MACRO
	case lexer.ALIAS:
		return PARSER_TOK_ALIAS
	case lexer.ATTRIBUTE:
		return PARSER_TOK_ATTRIBUTE
	case lexer.VARIANT:
		return PARSER_TOK_VARIANT
	case lexer.OF:
		return PARSER_TOK_OF
	case lexer.SCENARIO:
		return PARSER_TOK_SCENARIO
	case lexer.PROGRESS:
		return PARSER_TOK_PROGRESS
	case lexer.RELY:
		return PARSER_TOK_RELY
	case lexer.MIXORD:
		return PARSER_TOK_MIXORD
	case lexer.CONCEPT:
		return PARSER_TOK_CONCEPT
	case lexer.STATE:
		return PARSER_TOK_STATE
	case lexer.UPDATE:
		return PARSER_TOK_UPDATE
	case lexer.FROM:
		return PARSER_TOK_FROM
	case lexer.PARAMS:
		return PARSER_TOK_PARAMS
	case lexer.MODIFIES:
		return PARSER_TOK_MODIFIES
	case lexer.ENSURES:
		return PARSER_TOK_ENSURES
	case lexer.REQUIRES:
		return PARSER_TOK_REQUIRES
	case lexer.INIT:
		return PARSER_TOK_INIT
	case lexer.ENTRY:
		return PARSER_TOK_ENTRY
	case lexer.SET:
		return PARSER_TOK_SET
	case lexer.NULL:
		return PARSER_TOK_NULL
	case lexer.MATCH:
		return PARSER_TOK_MATCH
	case lexer.FRESH:
		return PARSER_TOK_FRESH
	case lexer.NAMED:
		return PARSER_TOK_NAMED
	case lexer.TEMPORAL:
		return PARSER_TOK_TEMPORAL
	case lexer.EXPLICIT:
		return PARSER_TOK_EXPLICIT
	case lexer.SPECIFICATION:
		return PARSER_TOK_SPECIFICATION
	case lexer.IMPLEMENTATION:
		return PARSER_TOK_IMPLEMENTATION
	case lexer.PRIVATE:
		return PARSER_TOK_PRIVATE
	case lexer.GLOBAL:
		return PARSER_TOK_GLOBAL
	case lexer.COMMON:
		return PARSER_TOK_COMMON
	case lexer.GHOST:
		return PARSER_TOK_GHOST
	case lexer.FINITE:
		return PARSER_TOK_FINITE
	case lexer.PARAMETER:
		return PARSER_TOK_PARAMETER
	case lexer.DESTRUCTOR:
		return PARSER_TOK_DESTRUCTOR
	case lexer.CONSTRUCTOR:
		return PARSER_TOK_CONSTRUCTOR
	case lexer.FIELD:
		return PARSER_TOK_FIELD
	case lexer.AUTOINSTANCE:
		return PARSER_TOK_AUTOINSTANCE
	case lexer.WITH:
		return PARSER_TOK_WITH

	// Proof/tactic tokens
	case lexer.TACTIC:
		return PARSER_TOK_TACTIC
	case lexer.TRIGGER:
		return PARSER_TOK_TRIGGER
	case lexer.SHOWGOALS:
		return PARSER_TOK_SHOWGOALS
	case lexer.DEFERGOAL:
		return PARSER_TOK_DEFERGOAL
	case lexer.SPOIL:
		return PARSER_TOK_SPOIL
	case lexer.UNFOLD:
		return PARSER_TOK_UNFOLD
	case lexer.FORGET:
		return PARSER_TOK_FORGET
	case lexer.APPLY:
		return PARSER_TOK_APPLY

	default:
		lval.str = tok.Value
		return PARSER_TOK_PRESYMBOL
	}
}

// Error is called by goyacc when a parse error occurs.
func (l *parser17LexAdapter) Error(s string) {
	l.err = s
}
