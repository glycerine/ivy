package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestDirectiveMultilineComment(t *testing.T) {
	for _, newline := range []string{"\n", "\r\n"} {
		source := strings.ReplaceAll("#define LIMIT 5 /* ratio (\n ) { unmatched } */\nstatic inline int value() { return LIMIT; }\n", "\n", newline)
		ts, err := Lex(source)
		if err != nil {
			t.Fatal(err)
		}
		var joined strings.Builder
		for _, tok := range ts {
			joined.WriteString(tok.Text)
		}
		if joined.String() != source {
			t.Fatal("source not preserved")
		}
		if ts[0].Kind != "directive" || !strings.Contains(ts[0].Text, "*/") {
			t.Fatal(ts[0])
		}
		if _, err := MatchDelimiters(ts); err != nil {
			t.Fatal(err)
		}
		r, err := Analyze("p/file.h", source)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.Draft, "const LIMIT = 5 /* ratio (") {
			t.Fatal(r.Draft)
		}
		if len(r.Functions) != 1 {
			t.Fatal("directive swallowed following code")
		}
	}
}

func TestDirectiveCommentMarkersInLiterals(t *testing.T) {
	for _, directive := range []string{`#define TEXT "/* not a comment"`, `#define TEXT "escaped \" /*"`, `#define SLASH '/'`, `#define X 1 // /* not a block comment`} {
		source := directive + "\nint next;\n"
		ts, err := Lex(source)
		if err != nil {
			t.Fatal(err)
		}
		if ts[0].Text != directive {
			t.Fatal(ts[0])
		}
	}
	if _, err := Lex("#define X 1 /* unfinished\n"); err == nil {
		t.Fatal("accepted unterminated block comment")
	}
}

func TestCuddIntNoFallback(t *testing.T) {
	path := "../src/bdd/cudd/cuddInt.h"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(path, string(source))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "const DD_MAX_LOOSE_FRACTION =") {
		t.Fatal("missing converted constant")
	}
	if !strings.Contains(r.Draft, "unique table in fast growth mode) */") {
		t.Fatal("comment lost")
	}
}
