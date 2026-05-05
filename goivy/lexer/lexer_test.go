package lexer

import (
	"testing"
)

var v17 = Version{1, 7}

func tok(input string, version Version) []Token {
	return Tokenize(input, version)
}

func assertToken(t *testing.T, tokens []Token, idx int, expected TokenType, value string) {
	t.Helper()
	if idx >= len(tokens) {
		t.Fatalf("token index %d out of range (have %d tokens)", idx, len(tokens))
	}
	if tokens[idx].Type != expected {
		t.Errorf("token[%d]: type = %s, want %s (value %q)", idx, tokens[idx].Type, expected, tokens[idx].Value)
	}
	if value != "" && tokens[idx].Value != value {
		t.Errorf("token[%d]: value = %q, want %q", idx, tokens[idx].Value, value)
	}
}

// --- Basic tokens ---

func TestEmpty(t *testing.T) {
	tokens := tok("", v17)
	assertToken(t, tokens, 0, EOF, "")
}

func TestPunctuation(t *testing.T) {
	tokens := tok("( ) { } [ ] , ; : .", v17)
	assertToken(t, tokens, 0, LPAREN, "(")
	assertToken(t, tokens, 1, RPAREN, ")")
	assertToken(t, tokens, 2, LCB, "{")
	assertToken(t, tokens, 3, RCB, "}")
	assertToken(t, tokens, 4, LB, "[")
	assertToken(t, tokens, 5, RB, "]")
	assertToken(t, tokens, 6, COMMA, ",")
	assertToken(t, tokens, 7, SEMI, ";")
	assertToken(t, tokens, 8, COLON, ":")
	assertToken(t, tokens, 9, DOT, ".")
}

func TestOperators(t *testing.T) {
	tokens := tok("+ - * / = ~ & | $ ^", v17)
	assertToken(t, tokens, 0, PLUS, "+")
	assertToken(t, tokens, 1, MINUS, "-")
	assertToken(t, tokens, 2, TIMES, "*")
	assertToken(t, tokens, 3, DIV, "/")
	assertToken(t, tokens, 4, EQ, "=")
	assertToken(t, tokens, 5, TILDA, "~")
	assertToken(t, tokens, 6, AND, "&")
	assertToken(t, tokens, 7, OR, "|")
	assertToken(t, tokens, 8, DOLLAR, "$")
	assertToken(t, tokens, 9, CARET, "^")
}

func TestMultiCharOperators(t *testing.T) {
	tokens := tok("-> <-> <= >= ~= := *> .. ...", v17)
	assertToken(t, tokens, 0, ARROW, "->")
	assertToken(t, tokens, 1, IFF, "<->")
	assertToken(t, tokens, 2, LE, "<=")
	assertToken(t, tokens, 3, GE, ">=")
	assertToken(t, tokens, 4, TILDAEQ, "~=")
	assertToken(t, tokens, 5, ASSIGN, ":=")
	assertToken(t, tokens, 6, PTO, "*>")
	assertToken(t, tokens, 7, DOTS, "..")
	assertToken(t, tokens, 8, DOTDOTDOT, "...")
}

func TestLTNotIFF(t *testing.T) {
	// < followed by something that isn't -> should be LT
	tokens := tok("< x", v17)
	assertToken(t, tokens, 0, LT, "<")
	assertToken(t, tokens, 1, SYMBOL, "x")
}

func TestGTAlone(t *testing.T) {
	tokens := tok("> x", v17)
	assertToken(t, tokens, 0, GT, ">")
}

// --- Identifiers ---

func TestSymbol(t *testing.T) {
	tokens := tok("foo bar_baz x123", v17)
	assertToken(t, tokens, 0, SYMBOL, "foo")
	assertToken(t, tokens, 1, SYMBOL, "bar_baz")
	assertToken(t, tokens, 2, SYMBOL, "x123")
}

func TestVariable(t *testing.T) {
	tokens := tok("X Y Z MyVar", v17)
	assertToken(t, tokens, 0, VARIABLE, "X")
	assertToken(t, tokens, 1, VARIABLE, "Y")
	assertToken(t, tokens, 2, VARIABLE, "Z")
	assertToken(t, tokens, 3, VARIABLE, "MyVar")
}

func TestVariableWithSubscript(t *testing.T) {
	tokens := tok("X[0] Y[abc]", v17)
	assertToken(t, tokens, 0, VARIABLE, "X[0]")
	assertToken(t, tokens, 1, VARIABLE, "Y[abc]")
}

func TestQuotedString(t *testing.T) {
	tokens := tok(`"hello" "world 123"`, v17)
	assertToken(t, tokens, 0, SYMBOL, `"hello"`)
	assertToken(t, tokens, 1, SYMBOL, `"world 123"`)
}

func TestDigitStart(t *testing.T) {
	tokens := tok("42 0x10", v17)
	assertToken(t, tokens, 0, SYMBOL, "42")
	assertToken(t, tokens, 1, SYMBOL, "0x10")
}

func TestUnderscoreStart(t *testing.T) {
	tokens := tok("_foo _bar", v17)
	assertToken(t, tokens, 0, SYMBOL, "_foo")
	assertToken(t, tokens, 1, SYMBOL, "_bar")
}

// --- Keywords ---

func TestKeywordsV17(t *testing.T) {
	tests := map[string]TokenType{
		"relation":   RELATION,
		"individual": INDIV,
		"axiom":      AXIOM,
		"conjecture": CONJECTURE,
		"action":     ACTION,
		"type":       TYPE,
		"if":         IF,
		"else":       ELSE,
		"forall":     FORALL,
		"exists":     EXISTS,
		"module":     MODULE,
		"object":     OBJECT,
		"property":   PROPERTY,
		"invariant":  INVARIANT,
		"struct":     STRUCT,
		"variant":    VARIANT,
		"of":         OF,
		"true":       TRUE,
		"false":      FALSE,
		"this":       THIS,
		"old":        OLD,
		"with":       WITH,
		"export":     EXPORT,
		"import":     IMPORT,
		"before":     BEFORE,
		"after":      AFTER,
		"isolate":    ISOLATE,
		"assume":     ASSUME,
		"assert":     ASSERT,
		"while":      WHILE,
		"let":        LET,
		"some":       SOME,
		"function":   FUNCTION,
		"definition": DEFINITION,
		"proof":      PROOF,
		"apply":      APPLY,
		"theorem":    THEOREM,
		"tactic":     TACTIC,
	}
	for word, expected := range tests {
		tokens := tok(word, v17)
		if tokens[0].Type != expected {
			t.Errorf("%q: got %s, want %s", word, tokens[0].Type, expected)
		}
	}
}

func TestVersionGatingV10(t *testing.T) {
	v10 := Version{1, 0}
	// "state" and "local" should NOT be keywords in v1.0
	tokens := tok("state local", v10)
	assertToken(t, tokens, 0, SYMBOL, "state")
	assertToken(t, tokens, 1, SYMBOL, "local")
}

func TestVersionGatingV14(t *testing.T) {
	v14 := Version{1, 4}
	// "struct" should NOT be a keyword in v1.4
	tokens := tok("struct", v14)
	assertToken(t, tokens, 0, SYMBOL, "struct")
}

func TestVersionGatingV15(t *testing.T) {
	v15 := Version{1, 5}
	// "variant" and "of" should NOT be keywords in v1.5
	tokens := tok("variant of", v15)
	assertToken(t, tokens, 0, SYMBOL, "variant")
	assertToken(t, tokens, 1, SYMBOL, "of")
}

func TestVersionGatingV16(t *testing.T) {
	v16 := Version{1, 6}
	// "tactic" should NOT be a keyword in v1.6
	tokens := tok("tactic", v16)
	assertToken(t, tokens, 0, SYMBOL, "tactic")
}

func TestVersionV2RequiresEnsuresRemoved(t *testing.T) {
	v2 := Version{2, 0}
	// "requires" and "ensures" are removed in versions > 1.7
	tokens := tok("requires ensures", v2)
	assertToken(t, tokens, 0, SYMBOL, "requires")
	assertToken(t, tokens, 1, SYMBOL, "ensures")
}

func TestAliasKeywords(t *testing.T) {
	// "instance" → INSTANTIATE, "template" → MODULE, "execute" → MIXIN, "process" → EXTRACT
	tokens := tok("instance", v17)
	assertToken(t, tokens, 0, INSTANTIATE, "instance")

	tokens = tok("execute", v17)
	assertToken(t, tokens, 0, MIXIN, "execute")
}

// --- Comments ---

func TestComment(t *testing.T) {
	tokens := tok("x # this is a comment\ny", v17)
	assertToken(t, tokens, 0, SYMBOL, "x")
	assertToken(t, tokens, 1, SYMBOL, "y")
}

func TestCommentOnly(t *testing.T) {
	tokens := tok("# just a comment", v17)
	assertToken(t, tokens, 0, EOF, "")
}

// --- Whitespace ---

func TestNewlines(t *testing.T) {
	tokens := tok("x\n\ny", v17)
	assertToken(t, tokens, 0, SYMBOL, "x")
	assertToken(t, tokens, 1, SYMBOL, "y")
	if tokens[1].Line != 3 {
		t.Errorf("y should be on line 3, got %d", tokens[1].Line)
	}
}

func TestMixedWhitespace(t *testing.T) {
	tokens := tok("  x\t\ry  ", v17)
	assertToken(t, tokens, 0, SYMBOL, "x")
	assertToken(t, tokens, 1, SYMBOL, "y")
}

// --- Native quotes ---

func TestNativeQuote(t *testing.T) {
	tokens := tok("<<<hello world>>>", v17)
	assertToken(t, tokens, 0, NATIVEQUOTE, "hello world")
}

func TestNativeQuoteMultiline(t *testing.T) {
	tokens := tok("<<<\nline1\nline2\n>>>", v17)
	assertToken(t, tokens, 0, NATIVEQUOTE, "\nline1\nline2\n")
}

func TestNativeQuoteUnterminated(t *testing.T) {
	tokens := tok("<<<not closed", v17)
	assertToken(t, tokens, 0, ERROR, "")
}

// --- Edge cases ---

func TestUnterminatedString(t *testing.T) {
	tokens := tok(`"hello`, v17)
	assertToken(t, tokens, 0, ERROR, "")
}

func TestIllegalCharacter(t *testing.T) {
	tokens := tok("@", v17)
	assertToken(t, tokens, 0, ERROR, "")
}

// --- Line/column tracking ---

func TestLineColumn(t *testing.T) {
	tokens := tok("x\n  y", v17)
	if tokens[0].Line != 1 || tokens[0].Column != 1 {
		t.Errorf("x at %d:%d, want 1:1", tokens[0].Line, tokens[0].Column)
	}
	if tokens[1].Line != 2 || tokens[1].Column != 3 {
		t.Errorf("y at %d:%d, want 2:3", tokens[1].Line, tokens[1].Column)
	}
}

// --- Peek ---

func TestPeek(t *testing.T) {
	l := NewLexer("x y", v17)
	p := l.Peek()
	assertToken(t, []Token{p}, 0, SYMBOL, "x")
	// Peek again should return same token
	p2 := l.Peek()
	assertToken(t, []Token{p2}, 0, SYMBOL, "x")
	// NextToken should return the peeked token
	tok := l.NextToken()
	assertToken(t, []Token{tok}, 0, SYMBOL, "x")
	// Now should get y
	tok2 := l.NextToken()
	assertToken(t, []Token{tok2}, 0, SYMBOL, "y")
}

// --- Realistic input ---

func TestRealisticIvy(t *testing.T) {
	input := `type node
relation link(X:node, Y:node)

axiom [transitivity] forall X:node, Y:node, Z:node.
    link(X,Y) & link(Y,Z) -> link(X,Z)

action connect(x:node, y:node) = {
    link(x,y) := true
}`
	tokens := tok(input, v17)

	// Spot check key tokens
	assertToken(t, tokens, 0, TYPE, "type")
	assertToken(t, tokens, 1, SYMBOL, "node")
	assertToken(t, tokens, 2, RELATION, "relation")
	assertToken(t, tokens, 3, SYMBOL, "link")
	assertToken(t, tokens, 4, LPAREN, "(")
	assertToken(t, tokens, 5, VARIABLE, "X")
	assertToken(t, tokens, 6, COLON, ":")

	// Count tokens, should end with EOF
	lastTok := tokens[len(tokens)-1]
	if lastTok.Type != EOF {
		t.Errorf("last token should be EOF, got %s", lastTok.Type)
	}
}

func TestDotVsDots(t *testing.T) {
	// Ensure "." vs ".." vs "..." are distinguished
	tokens := tok("a.b c..d e...f", v17)
	assertToken(t, tokens, 0, SYMBOL, "a")
	assertToken(t, tokens, 1, DOT, ".")
	assertToken(t, tokens, 2, SYMBOL, "b")
	assertToken(t, tokens, 3, SYMBOL, "c")
	assertToken(t, tokens, 4, DOTS, "..")
	assertToken(t, tokens, 5, SYMBOL, "d")
	assertToken(t, tokens, 6, SYMBOL, "e")
	assertToken(t, tokens, 7, DOTDOTDOT, "...")
	assertToken(t, tokens, 8, SYMBOL, "f")
}

func TestArrowVsMinus(t *testing.T) {
	tokens := tok("x - y x -> y", v17)
	assertToken(t, tokens, 0, SYMBOL, "x")
	assertToken(t, tokens, 1, MINUS, "-")
	assertToken(t, tokens, 2, SYMBOL, "y")
	assertToken(t, tokens, 3, SYMBOL, "x")
	assertToken(t, tokens, 4, ARROW, "->")
	assertToken(t, tokens, 5, SYMBOL, "y")
}

func TestIffVsLt(t *testing.T) {
	tokens := tok("x <-> y x < y", v17)
	assertToken(t, tokens, 0, SYMBOL, "x")
	assertToken(t, tokens, 1, IFF, "<->")
	assertToken(t, tokens, 2, SYMBOL, "y")
	assertToken(t, tokens, 3, SYMBOL, "x")
	assertToken(t, tokens, 4, LT, "<")
	assertToken(t, tokens, 5, SYMBOL, "y")
}

func TestTokenize(t *testing.T) {
	tokens := Tokenize("a + b", v17)
	if len(tokens) != 4 { // a, +, b, EOF
		t.Errorf("got %d tokens, want 4", len(tokens))
	}
}

// --- Fuzz tests ---

func FuzzLexer(f *testing.F) {
	// Seed corpus
	f.Add([]byte(""))
	f.Add([]byte("type node"))
	f.Add([]byte("relation link(X:node, Y:node)"))
	f.Add([]byte("forall X. exists Y. X -> Y"))
	f.Add([]byte("<<<hello>>>"))
	f.Add([]byte("<<<unterminated"))
	f.Add([]byte(`"hello"`))
	f.Add([]byte(`"unterminated`))
	f.Add([]byte("# comment\n"))
	f.Add([]byte("a.b c..d e...f"))
	f.Add([]byte("<-> <= >= ~= := *> -> ..."))
	f.Add([]byte("X[0] Y[abc]"))
	f.Add([]byte("\x00\xff\xfe"))

	f.Fuzz(func(t *testing.T, data []byte) {
		input := string(data)
		l := NewLexer(input, Version{1, 7})
		for i := 0; i < 10000; i++ {
			tok := l.NextToken()
			if tok.Type == EOF || tok.Type == ERROR {
				break
			}
		}
	})
}

func FuzzLexerAllVersions(f *testing.F) {
	// Seed corpus with edge cases and version components
	f.Add([]byte(""), 1, 7)
	f.Add([]byte("   \t\n\r  "), 1, 0)
	f.Add([]byte("héllo wörld"), 1, 4)
	f.Add([]byte(`"unterminated string`), 1, 5)
	f.Add([]byte("<<<unterminated native quote"), 1, 6)
	f.Add([]byte("<<<properly closed>>>"), 2, 0)
	f.Add([]byte("type struct variant of tactic"), 1, 7)
	f.Add([]byte("requires ensures"), 2, 0)
	f.Add([]byte("state local"), 1, 0)
	f.Add([]byte("@@@!!!???"), 1, 7)
	f.Add([]byte("X[\x00]"), 1, 7)
	f.Add([]byte("\u25A1\u25C7"), 1, 7) // □◇ temporal operators
	f.Add([]byte("a + b * c / d - e & f | g"), 0, 0)

	f.Fuzz(func(t *testing.T, data []byte, major int, minor int) {
		// Clamp version to reasonable range to avoid
		// spending all fuzz time on degenerate versions.
		if major < 0 {
			major = 0
		}
		if major > 10 {
			major = 10
		}
		if minor < 0 {
			minor = 0
		}
		if minor > 20 {
			minor = 20
		}

		input := string(data)
		l := NewLexer(input, Version{major, minor})
		for i := 0; i < 10000; i++ {
			tok := l.NextToken()
			if tok.Type == EOF || tok.Type == ERROR {
				break
			}
		}
	})
}
