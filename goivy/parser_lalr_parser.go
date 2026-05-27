package goivy

import (
	"github.com/glycerine/ivy/goivy/xtracer"
)

// Parse parses a complete Ivy file and returns the declarations.
// It dispatches to the appropriate version-specific LALR parser.
// The optional importer callback resolves `include` directives.
func Parse(input string, version Version, opts ...ParseOption) (*ParseResult, error) {
	parser := fullFileParserForVersion(version)
	return parser(input, version, opts...)
}

type fullFileParser func(input string, version Version, opts ...ParseOption) (*ParseResult, error)

func fullFileParserForVersion(version Version) fullFileParser {
	// Status quo: the hardened v1.7+ full-file grammar is still the only
	// complete parser. A future thin v1.6 grammar should route from here.
	return ParseV17
}

// ParseV17 parses a complete Ivy file using the v1.7+ LALR grammar.
func ParseV17(input string, version Version, opts ...ParseOption) (*ParseResult, error) {
	xtracer.Trace("parser.Parse ENTER")
	cfg := newParseConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	lex := newParser17LexAdapter(input, version)
	cfg.applyToParser17(lex)
	parser17Parse(lex)
	if lex.err != nil {
		return nil, lex.err
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
	if len(lex.accum.errors) > 0 {
		return nil, lex.accum.errors[0]
	}
	result := lex.accum.toResult()
	xtracer.Trace("parser.Parse EXIT decls=%d", len(result.Decls))
	return result, nil
}

// --- parser17LexAdapter: adapter from lexer.Lexer to goyacc's parser17Lexer interface ---

// ImporterFunc is the callback for resolving `include` directives.
// It takes a module name and the current parser accumulator, so nested
// imports can see the same include stack as Python's ivy_parser.stack.
// Matches Python's ivy_parser.importer function.
type ImporterFunc func(name string, parent *ivyAccum) (*ParseResult, error)

type parser17LexAdapter struct {
	lex              *Lexer
	accum            *ivyAccum
	result           Node
	err              *ParseError
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

func newParser17LexAdapter(input string, version Version) *parser17LexAdapter {
	return &parser17LexAdapter{
		lex:      NewLexer(input, version),
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
	case EOF:
		return 0
	case SYMBOL:
		return PARSER_TOK_PRESYMBOL
	case VARIABLE:
		return PARSER_TOK_VARIABLE
	case NATIVEQUOTE:
		return PARSER_TOK_NATIVEQUOTE

	// Punctuation
	case LPAREN:
		return PARSER_TOK_LPAREN
	case RPAREN:
		return PARSER_TOK_RPAREN
	case LB:
		return PARSER_TOK_LB
	case RB:
		return PARSER_TOK_RB
	case LCB:
		return PARSER_TOK_LCB
	case RCB:
		return PARSER_TOK_RCB
	case COMMA:
		return PARSER_TOK_COMMA
	case SEMI:
		return PARSER_TOK_SEMI
	case COLON:
		return PARSER_TOK_COLON
	case DOT:
		return PARSER_TOK_DOT
	case DOTS:
		return PARSER_TOK_DOTS
	case DOTDOTDOT:
		return PARSER_TOK_DOTDOTDOT

	// Operators
	case PLUS:
		return PARSER_TOK_PLUS
	case MINUS:
		return PARSER_TOK_MINUS
	case TIMES:
		return PARSER_TOK_TIMES
	case DIV:
		return PARSER_TOK_DIV
	case EQ:
		return PARSER_TOK_EQ
	case TILDAEQ:
		return PARSER_TOK_TILDAEQ
	case TILDA:
		return PARSER_TOK_TILDA
	case LE:
		return PARSER_TOK_LE
	case LT:
		return PARSER_TOK_LT
	case GE:
		return PARSER_TOK_GE
	case GT:
		return PARSER_TOK_GT
	case AND:
		return PARSER_TOK_AND
	case OR:
		return PARSER_TOK_OR
	case ARROW:
		return PARSER_TOK_ARROW
	case IFF:
		return PARSER_TOK_IFF
	case PTO:
		return PARSER_TOK_PTO
	case DOLLAR:
		return PARSER_TOK_DOLLAR
	case CARET:
		return PARSER_TOK_CARET
	case ASSIGN:
		return PARSER_TOK_ASSIGN

	// Logic keywords
	case FORALL:
		return PARSER_TOK_FORALL
	case EXISTS:
		return PARSER_TOK_EXISTS
	case TRUE:
		return PARSER_TOK_TRUE
	case FALSE:
		return PARSER_TOK_FALSE
	case OLD:
		return PARSER_TOK_OLD
	case THIS:
		return PARSER_TOK_THIS
	case ISA:
		return PARSER_TOK_ISA
	case IF:
		return PARSER_TOK_IF
	case ELSE:
		return PARSER_TOK_ELSE
	case GLOBALLY:
		return PARSER_TOK_GLOBALLY
	case EVENTUALLY:
		return PARSER_TOK_EVENTUALLY
	case WHENNEXT:
		return PARSER_TOK_WHENNEXT
	case WHENPREV:
		return PARSER_TOK_WHENPREV
	case WHENFIRST:
		return PARSER_TOK_WHENFIRST
	case WHENLAST:
		return PARSER_TOK_WHENLAST

	// Action keywords
	case ASSUME:
		return PARSER_TOK_ASSUME
	case ASSERT:
		return PARSER_TOK_ASSERT
	case REQUIRE:
		return PARSER_TOK_REQUIRE
	case ENSURE:
		return PARSER_TOK_ENSURE
	case VAR:
		return PARSER_TOK_VAR
	case LOCAL:
		return PARSER_TOK_LOCAL
	case LET:
		return PARSER_TOK_LET
	case CALL:
		return PARSER_TOK_CALL
	case WHILE:
		return PARSER_TOK_WHILE
	case FOR:
		return PARSER_TOK_FOR
	case IN:
		return PARSER_TOK_IN
	case INVARIANT:
		return PARSER_TOK_INVARIANT
	case DECREASES:
		return PARSER_TOK_DECREASES
	case RETURNS:
		return PARSER_TOK_RETURNS
	case SOME:
		return PARSER_TOK_SOME
	case MINIMIZING:
		return PARSER_TOK_MINIMIZING
	case MAXIMIZING:
		return PARSER_TOK_MAXIMIZING
	case DEBUG:
		return PARSER_TOK_DEBUG
	case THUNK:
		return PARSER_TOK_THUNK
	case UNPROVABLE:
		return PARSER_TOK_UNPROVABLE
	case PROOF:
		return PARSER_TOK_PROOF
	case INSTANTIATE:
		return PARSER_TOK_INSTANTIATE

	// Declaration keywords
	case RELATION:
		return PARSER_TOK_RELATION
	case INDIV:
		return PARSER_TOK_INDIV
	case FUNCTION:
		return PARSER_TOK_FUNCTION
	case DERIVED:
		return PARSER_TOK_DERIVED
	case AXIOM:
		return PARSER_TOK_AXIOM
	case CONJECTURE:
		return PARSER_TOK_CONJECTURE
	case SCHEMA:
		return PARSER_TOK_SCHEMA
	case THEOREM:
		return PARSER_TOK_THEOREM
	case PROPERTY:
		return PARSER_TOK_PROPERTY
	case DEFINITION:
		return PARSER_TOK_DEFINITION
	case TYPE:
		return PARSER_TOK_TYPE
	case STRUCT:
		return PARSER_TOK_STRUCT
	case MODULE:
		return PARSER_TOK_MODULE
	case OBJECT:
		return PARSER_TOK_OBJECT
	case CLASS:
		return PARSER_TOK_CLASS
	case SUBCLASS:
		return PARSER_TOK_SUBCLASS
	case ACTION:
		return PARSER_TOK_ACTION
	case METHOD:
		return PARSER_TOK_METHOD
	case BEFORE:
		return PARSER_TOK_BEFORE
	case AFTER:
		return PARSER_TOK_AFTER
	case AROUND:
		return PARSER_TOK_AROUND
	case MIXIN:
		return PARSER_TOK_MIXIN
	case IMPLEMENT:
		return PARSER_TOK_IMPLEMENT
	case ISOLATE:
		return PARSER_TOK_ISOLATE
	case EXTRACT:
		return PARSER_TOK_EXTRACT
	case TRUSTED:
		return PARSER_TOK_TRUSTED
	case EXPORT:
		return PARSER_TOK_EXPORT
	case IMPORT:
		return PARSER_TOK_IMPORT
	case DELEGATE:
		return PARSER_TOK_DELEGATE
	case USING:
		return PARSER_TOK_USING
	case INCLUDE:
		return PARSER_TOK_INCLUDE
	case INTERPRET:
		return PARSER_TOK_INTERPRET
	case MACRO:
		return PARSER_TOK_MACRO
	case ALIAS:
		return PARSER_TOK_ALIAS
	case ATTRIBUTE:
		return PARSER_TOK_ATTRIBUTE
	case VARIANT:
		return PARSER_TOK_VARIANT
	case OF:
		return PARSER_TOK_OF
	case SCENARIO:
		return PARSER_TOK_SCENARIO
	case PROGRESS:
		return PARSER_TOK_PROGRESS
	case RELY:
		return PARSER_TOK_RELY
	case MIXORD:
		return PARSER_TOK_MIXORD
	case CONCEPT:
		return PARSER_TOK_CONCEPT
	case STATE:
		return PARSER_TOK_STATE
	case UPDATE:
		return PARSER_TOK_UPDATE
	case FROM:
		return PARSER_TOK_FROM
	case PARAMS:
		return PARSER_TOK_PARAMS
	case MODIFIES:
		return PARSER_TOK_MODIFIES
	case ENSURES:
		return PARSER_TOK_ENSURES
	case REQUIRES:
		return PARSER_TOK_REQUIRES
	case INIT:
		return PARSER_TOK_INIT
	case ENTRY:
		return PARSER_TOK_ENTRY
	case SET:
		return PARSER_TOK_SET
	case NULL:
		return PARSER_TOK_NULL
	case MATCH:
		return PARSER_TOK_MATCH
	case FRESH:
		return PARSER_TOK_FRESH
	case NAMED:
		return PARSER_TOK_NAMED
	case TEMPORAL:
		return PARSER_TOK_TEMPORAL
	case EXPLICIT:
		return PARSER_TOK_EXPLICIT
	case SPECIFICATION:
		return PARSER_TOK_SPECIFICATION
	case IMPLEMENTATION:
		return PARSER_TOK_IMPLEMENTATION
	case PRIVATE:
		return PARSER_TOK_PRIVATE
	case GLOBAL:
		return PARSER_TOK_GLOBAL
	case COMMON:
		return PARSER_TOK_COMMON
	case GHOST:
		return PARSER_TOK_GHOST
	case FINITE:
		return PARSER_TOK_FINITE
	case PARAMETER:
		return PARSER_TOK_PARAMETER
	case DESTRUCTOR:
		return PARSER_TOK_DESTRUCTOR
	case CONSTRUCTOR:
		return PARSER_TOK_CONSTRUCTOR
	case FIELD:
		return PARSER_TOK_FIELD
	case AUTOINSTANCE:
		return PARSER_TOK_AUTOINSTANCE
	case WITH:
		return PARSER_TOK_WITH

	// Proof/tactic tokens
	case TACTIC:
		return PARSER_TOK_TACTIC
	case TRIGGER:
		return PARSER_TOK_TRIGGER
	case SHOWGOALS:
		return PARSER_TOK_SHOWGOALS
	case DEFERGOAL:
		return PARSER_TOK_DEFERGOAL
	case SPOIL:
		return PARSER_TOK_SPOIL
	case UNFOLD:
		return PARSER_TOK_UNFOLD
	case FORGET:
		return PARSER_TOK_FORGET
	case APPLY:
		return PARSER_TOK_APPLY

	default:
		lval.str = tok.Value
		return PARSER_TOK_PRESYMBOL
	}
}

// Error is called by goyacc when a parse error occurs.
// Mirrors Python p_error (ivy_parser.py:3598-3606): if there is a current
// token report its lineno+value; otherwise (EOF) report "unexpected end of input".
func (l *parser17LexAdapter) Error(s string) {
	pe := &ParseError{Filename: l.filename, Message: s}
	if l.lastTok.Type == EOF {
		pe.Message = "unexpected end of input"
	} else {
		pe.Lineno = l.lastTok.Line
		pe.Token = l.lastTok.Value
	}
	l.err = pe
}
