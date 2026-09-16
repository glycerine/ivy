package goivy

import (
	"fmt"
	"unicode/utf8"
)

// allReserved is the complete set of reserved keywords mapping to token types.
var allReserved = map[string]TokenType{
	"relation":       RELATION,
	"individual":     INDIV,
	"function":       FUNCTION,
	"axiom":          AXIOM,
	"conjecture":     CONJECTURE,
	"schema":         SCHEMA,
	"instantiate":    INSTANTIATE,
	"instance":       INSTANTIATE,
	"derived":        DERIVED,
	"concept":        CONCEPT,
	"init":           INIT,
	"action":         ACTION,
	"method":         METHOD,
	"field":          FIELD,
	"state":          STATE,
	"assume":         ASSUME,
	"assert":         ASSERT,
	"set":            SET,
	"null":           NULL,
	"old":            OLD,
	"from":           FROM,
	"update":         UPDATE,
	"params":         PARAMS,
	"in":             IN,
	"match":          MATCH,
	"ensures":        ENSURES,
	"requires":       REQUIRES,
	"modifies":       MODIFIES,
	"true":           TRUE,
	"false":          FALSE,
	"fresh":          FRESH,
	"module":         MODULE,
	"template":       MODULE,
	"object":         OBJECT,
	"class":          CLASS,
	"type":           TYPE,
	"if":             IF,
	"else":           ELSE,
	"local":          LOCAL,
	"let":            LET,
	"call":           CALL,
	"entry":          ENTRY,
	"macro":          MACRO,
	"interpret":      INTERPRET,
	"forall":         FORALL,
	"exists":         EXISTS,
	"returns":        RETURNS,
	"mixin":          MIXIN,
	"execute":        MIXIN,
	"before":         BEFORE,
	"after":          AFTER,
	"isolate":        ISOLATE,
	"with":           WITH,
	"export":         EXPORT,
	"delegate":       DELEGATE,
	"import":         IMPORT,
	"using":          USING,
	"include":        INCLUDE,
	"progress":       PROGRESS,
	"rely":           RELY,
	"mixord":         MIXORD,
	"extract":        EXTRACT,
	"process":        EXTRACT,
	"destructor":     DESTRUCTOR,
	"some":           SOME,
	"maximizing":     MAXIMIZING,
	"minimizing":     MINIMIZING,
	"private":        PRIVATE,
	"implement":      IMPLEMENT,
	"property":       PROPERTY,
	"while":          WHILE,
	"invariant":      INVARIANT,
	"struct":         STRUCT,
	"definition":     DEFINITION,
	"ghost":          GHOST,
	"alias":          ALIAS,
	"trusted":        TRUSTED,
	"this":           THIS,
	"var":            VAR,
	"attribute":      ATTRIBUTE,
	"variant":        VARIANT,
	"of":             OF,
	"scenario":       SCENARIO,
	"proof":          PROOF,
	"named":          NAMED,
	"temporal":       TEMPORAL,
	"globally":       GLOBALLY,
	"eventually":     EVENTUALLY,
	"decreases":      DECREASES,
	"specification":  SPECIFICATION,
	"implementation": IMPLEMENTATION,
	"global":         GLOBAL,
	"common":         COMMON,
	"ensure":         ENSURE,
	"require":        REQUIRE,
	"around":         AROUND,
	"parameter":      PARAMETER,
	"apply":          APPLY,
	"theorem":        THEOREM,
	"showgoals":      SHOWGOALS,
	"defergoal":      DEFERGOAL,
	"spoil":          SPOIL,
	"explicit":       EXPLICIT,
	"thunk":          THUNK,
	"isa":            ISA,
	"autoinstance":   AUTOINSTANCE,
	"constructor":    CONSTRUCTOR,
	"finite":         FINITE,
	"tactic":         TACTIC,
	"unfold":         UNFOLD,
	"forget":         FORGET,
	"debug":          DEBUG,
	"for":            FOR,
	"subclass":       SUBCLASS,
	"whenfirst":      WHENFIRST,
	"whenlast":       WHENLAST,
	"whenprev":       WHENPREV,
	"whennext":       WHENNEXT,
	"unprovable":     UNPROVABLE,
	"trigger":        TRIGGER,
}

// Version represents an Ivy language version like [1, 7].
type Version [2]int

// Lexer tokenizes Ivy source code.
type Lexer struct {
	input    string
	pos      int // current byte position
	line     int
	col      int
	reserved map[string]TokenType
	peeked   *Token
}

// New creates a new Lexer for the given input and language version.
func NewLexer(input string, version Version) *Lexer {
	l := &Lexer{
		input:    input,
		pos:      0,
		line:     1,
		col:      1,
		reserved: buildReserved(version),
	}
	return l
}

// buildReserved creates the version-gated reserved word map.
func buildReserved(v Version) map[string]TokenType {
	res := make(map[string]TokenType, len(allReserved))
	for k, v := range allReserved {
		res[k] = v
	}

	vle := func(major, minor int) bool {
		return v[0] < major || (v[0] == major && v[1] <= minor)
	}

	if vle(1, 0) {
		delete(res, "state")
		delete(res, "local")
	}
	if vle(1, 1) {
		for _, s := range []string{"returns", "mixin", "before", "after", "isolate",
			"with", "export", "delegate", "import", "include"} {
			delete(res, s)
		}
	} else {
		for _, s := range []string{"state", "set", "null", "match"} {
			delete(res, s)
		}
	}
	if vle(1, 4) {
		for _, s := range []string{"function", "class", "object", "method", "execute",
			"destructor", "some", "maximizing", "minimizing", "private", "implement",
			"using", "property", "while", "invariant", "struct", "definition", "ghost",
			"alias", "trusted", "this", "var", "attribute", "scenario", "proof", "named", "fresh"} {
			delete(res, s)
		}
	}
	if vle(1, 5) {
		for _, s := range []string{"variant", "of", "globally", "eventually", "temporal"} {
			delete(res, s)
		}
	}
	if vle(1, 6) {
		for _, s := range []string{"decreases", "specification", "implementation", "require",
			"ensure", "around", "parameter", "apply", "theorem", "showgoals", "spoil",
			"explicit", "thunk", "isa", "autoinstance", "constructor", "tactic", "finite",
			"unfold", "forget"} {
			delete(res, s)
		}
	}
	if vle(1, 7) {
		for _, s := range []string{"global", "common", "debug", "field", "for", "process",
			"subclass", "template", "whenfirst", "whenlast", "whennext", "whenprev",
			"unprovable", "trigger"} {
			delete(res, s)
		}
	} else {
		for _, s := range []string{"requires", "ensures"} {
			delete(res, s)
		}
	}

	return res
}

// peek returns the next rune without consuming it.
func (l *Lexer) peekRune() (rune, int) {
	if l.pos >= len(l.input) {
		return 0, 0
	}
	return utf8.DecodeRuneInString(l.input[l.pos:])
}

// advance consumes one rune and updates line/col.
//
// One cosmetic-only note: The lexer's col tracker increments
// by 1 for tabs (should arguably be more for display purposes),
// but this only affects error message column
// numbers — not tokenization or parsing correctness. Python's
// PLY lexer doesn't track columns at all, so there's no
// deviation from Python behavior.
func (l *Lexer) advance() rune {
	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

// match checks if the next bytes match s, and if so advances past them.
func (l *Lexer) match(s string) bool {
	if l.pos+len(s) <= len(l.input) && l.input[l.pos:l.pos+len(s)] == s {
		for range s {
			l.advance()
		}
		return true
	}
	return false
}

// skipWhitespaceAndComments skips spaces, tabs, \r, \n, and # comments.
func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.input) {
		r, _ := l.peekRune()
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			l.advance()
		} else if r == '#' {
			// Skip to end of line
			for l.pos < len(l.input) {
				r2, _ := l.peekRune()
				if r2 == '\n' {
					break
				}
				l.advance()
			}
		} else {
			break
		}
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	if l.peeked != nil {
		tok := *l.peeked
		l.peeked = nil
		return tok
	}
	return l.scan()
}

// Peek returns the next token without consuming it.
func (l *Lexer) Peek() Token {
	if l.peeked != nil {
		return *l.peeked
	}
	tok := l.scan()
	l.peeked = &tok
	return tok
}

func (l *Lexer) scan() Token {
	l.skipWhitespaceAndComments()

	if l.pos >= len(l.input) {
		return Token{Type: EOF, Line: l.line, Column: l.col}
	}

	line, col := l.line, l.col
	r, _ := l.peekRune()

	// Native quote: <<<...>>>
	if r == '<' && l.pos+3 <= len(l.input) && l.input[l.pos:l.pos+3] == "<<<" {
		return l.scanNativeQuote(line, col)
	}

	// Multi-character operators (order matters for longest match)
	switch r {
	case '.':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '.' {
			l.advance()
			if l.pos < len(l.input) && l.input[l.pos] == '.' {
				l.advance()
				return Token{Type: DOTDOTDOT, Value: "...", Line: line, Column: col}
			}
			return Token{Type: DOTS, Value: "..", Line: line, Column: col}
		}
		return Token{Type: DOT, Value: ".", Line: line, Column: col}

	case '-':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '>' {
			l.advance()
			return Token{Type: ARROW, Value: "->", Line: line, Column: col}
		}
		return Token{Type: MINUS, Value: "-", Line: line, Column: col}

	case '<':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '-' {
			l.advance()
			if l.pos < len(l.input) && l.input[l.pos] == '>' {
				l.advance()
				return Token{Type: IFF, Value: "<->", Line: line, Column: col}
			}
			// Back up: we consumed '<' and '-', put '-' back
			l.pos--
			l.col--
			return Token{Type: LT, Value: "<", Line: line, Column: col}
		}
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return Token{Type: LE, Value: "<=", Line: line, Column: col}
		}
		return Token{Type: LT, Value: "<", Line: line, Column: col}

	case '>':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return Token{Type: GE, Value: ">=", Line: line, Column: col}
		}
		return Token{Type: GT, Value: ">", Line: line, Column: col}

	case '~':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return Token{Type: TILDAEQ, Value: "~=", Line: line, Column: col}
		}
		return Token{Type: TILDA, Value: "~", Line: line, Column: col}

	case '*':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '>' {
			l.advance()
			return Token{Type: PTO, Value: "*>", Line: line, Column: col}
		}
		return Token{Type: TIMES, Value: "*", Line: line, Column: col}

	case ':':
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.advance()
			return Token{Type: ASSIGN, Value: ":=", Line: line, Column: col}
		}
		return Token{Type: COLON, Value: ":", Line: line, Column: col}

	// Single-character tokens
	case ',':
		l.advance()
		return Token{Type: COMMA, Value: ",", Line: line, Column: col}
	case '(':
		l.advance()
		return Token{Type: LPAREN, Value: "(", Line: line, Column: col}
	case ')':
		l.advance()
		return Token{Type: RPAREN, Value: ")", Line: line, Column: col}
	case '{':
		l.advance()
		return Token{Type: LCB, Value: "{", Line: line, Column: col}
	case '}':
		l.advance()
		return Token{Type: RCB, Value: "}", Line: line, Column: col}
	case '[':
		l.advance()
		return Token{Type: LB, Value: "[", Line: line, Column: col}
	case ']':
		l.advance()
		return Token{Type: RB, Value: "]", Line: line, Column: col}
	case '+':
		l.advance()
		return Token{Type: PLUS, Value: "+", Line: line, Column: col}
	case '/':
		l.advance()
		return Token{Type: DIV, Value: "/", Line: line, Column: col}
	case '&':
		l.advance()
		return Token{Type: AND, Value: "&", Line: line, Column: col}
	case '|':
		l.advance()
		return Token{Type: OR, Value: "|", Line: line, Column: col}
	case '=':
		l.advance()
		return Token{Type: EQ, Value: "=", Line: line, Column: col}
	case ';':
		l.advance()
		return Token{Type: SEMI, Value: ";", Line: line, Column: col}
	case '$':
		l.advance()
		return Token{Type: DOLLAR, Value: "$", Line: line, Column: col}
	case '^':
		l.advance()
		return Token{Type: CARET, Value: "^", Line: line, Column: col}
	}

	// Unicode temporal operators
	if r == '\u25A1' { // □ globally
		l.advance()
		return Token{Type: GLOBALLY, Value: "□", Line: line, Column: col}
	}
	if r == '\u25C7' { // ◇ eventually
		l.advance()
		return Token{Type: EVENTUALLY, Value: "◇", Line: line, Column: col}
	}

	// Quoted string → SYMBOL
	if r == '"' {
		return l.scanQuotedString(line, col)
	}

	// Uppercase identifier → VARIABLE (or reserved word)
	if isASCIIUpperRune(r) {
		return l.scanVariable(line, col)
	}

	// Lowercase identifier or digit → SYMBOL or keyword
	if r == '_' || isASCIILowerRune(r) || isASCIIDigitRune(r) {
		return l.scanSymbol(line, col)
	}

	// Unknown character
	l.advance()
	return Token{Type: ERROR, Value: fmt.Sprintf("illegal character '%c'", r), Line: line, Column: col}
}

func (l *Lexer) scanQuotedString(line, col int) Token {
	l.advance() // consume opening "
	start := l.pos
	for l.pos < len(l.input) {
		r, _ := l.peekRune()
		if r == '"' {
			val := l.input[start:l.pos]
			l.advance() // consume closing "
			return Token{Type: SYMBOL, Value: `"` + val + `"`, Line: line, Column: col}
		}
		l.advance()
	}
	// Unterminated string
	return Token{Type: ERROR, Value: "unterminated string", Line: line, Column: col}
}

func (l *Lexer) scanSymbol(line, col int) Token {
	start := l.pos
	for l.pos < len(l.input) {
		r, _ := l.peekRune()
		if r == '_' || isASCIILetterRune(r) || isASCIIDigitRune(r) {
			l.advance()
		} else {
			break
		}
	}
	val := l.input[start:l.pos]
	if tt, ok := l.reserved[val]; ok {
		return Token{Type: tt, Value: val, Line: line, Column: col}
	}
	return Token{Type: SYMBOL, Value: val, Line: line, Column: col}
}

func (l *Lexer) scanVariable(line, col int) Token {
	start := l.pos
	// First char is uppercase letter
	l.advance()
	for l.pos < len(l.input) {
		r, _ := l.peekRune()
		if r == '_' || isASCIILetterRune(r) || isASCIIDigitRune(r) {
			l.advance()
		} else if r == '[' {
			end, ok := l.validVariableSubscriptEnd(l.pos)
			if !ok {
				break
			}
			for l.pos < end {
				l.advance()
			}
		} else {
			break
		}
	}
	val := l.input[start:l.pos]
	if tt, ok := l.reserved[val]; ok {
		return Token{Type: tt, Value: val, Line: line, Column: col}
	}
	return Token{Type: VARIABLE, Value: val, Line: line, Column: col}
}

func (l *Lexer) validVariableSubscriptEnd(pos int) (int, bool) {
	if pos >= len(l.input) || l.input[pos] != '[' {
		return pos, false
	}
	i := pos + 1
	for i < len(l.input) && isPythonVariableSubscriptByte(l.input[i]) {
		i++
	}
	if i < len(l.input) && l.input[i] == ']' {
		return i + 1, true
	}
	return pos, false
}

func isPythonVariableSubscriptByte(b byte) bool {
	return b == '_' ||
		('a' <= b && b <= 'z') ||
		('A' <= b && b <= 'Z') ||
		('0' <= b && b <= '9')
}

func isASCIILowerRune(r rune) bool {
	return 'a' <= r && r <= 'z'
}

func isASCIIUpperRune(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func isASCIIDigitRune(r rune) bool {
	return '0' <= r && r <= '9'
}

func isASCIILetterRune(r rune) bool {
	return isASCIILowerRune(r) || isASCIIUpperRune(r)
}

func (l *Lexer) scanNativeQuote(line, col int) Token {
	// Consume <<<
	for i := 0; i < 3; i++ {
		l.advance()
	}
	start := l.pos
	for l.pos < len(l.input) {
		if l.pos+3 <= len(l.input) && l.input[l.pos:l.pos+3] == ">>>" {
			val := l.input[start:l.pos]
			for i := 0; i < 3; i++ {
				l.advance()
			}
			return Token{Type: NATIVEQUOTE, Value: val, Line: line, Column: col}
		}
		l.advance()
	}
	return Token{Type: ERROR, Value: "unterminated native quote", Line: line, Column: col}
}

// Tokenize returns all tokens from the input.
func Tokenize(input string, version Version) []Token {
	l := NewLexer(input, version)
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF || tok.Type == ERROR {
			break
		}
	}
	return tokens
}
