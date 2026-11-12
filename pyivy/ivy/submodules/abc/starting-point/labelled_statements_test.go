package startingpoint

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestBraceIfAfterLabel(t *testing.T) {
	got, err := braceStatements("{ finish:\nif (x) f(); }")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "if (x) { f(); }") {
		t.Fatal(got)
	}
}

func TestBraceStackedLabels(t *testing.T) {
	got, err := braceStatements("{ first: /* keep */ finish: if (x) for (;;) f(); else g(); }")
	if err != nil {
		t.Fatal(err)
	}
	want := "{ first: /* keep */ finish: if (x) { for (;;) { f(); } } else { g(); } }"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReoShuffleLabelledIf(t *testing.T) {
	source, err := os.ReadFile("../src/bdd/reo/reoShuffle.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("reoShuffle.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	want := regexp.MustCompile(`finish\s*:\s*if\s+bFuncRes(?:\s*!=\s*nil)?\s*\{\s*Cudd_Deref\(\s*bFuncRes\s*\);?\s*\}`)
	if !want.MatchString(r.Draft) {
		t.Fatalf("missing braced cleanup after finish label:\n%s", r.Draft)
	}
}

func TestTransformationsAfterLabels(t *testing.T) {
	for _, tc := range []struct{ name, statement string }{
		{"declaration", "int n = 0;"},
		{"initialized declaration", "int n = 3;"},
		{"array", "int a[2];"},
		{"ternary assignment", "x = x ? 1 : 2;"},
		{"ternary return", "return x ? 1 : 2;"},
		{"comma", "x = 1, x = 2;"},
		{"parenthesized comma", "(x = 1, x = 2);"},
		{"multiple declarations", "int a, b = 0;"},
		{"pointer declaration", "FILE * p;"},
		{"ternary initializer", "int n = x ? 1 : 2;"},
		{"static", "static int n = 0; n++;"},
		{"aggregate", "int a[2] = {1, 2};"},
		{"do while", "do { x--; } while (x);"},
		{"if else", "if (x) x--; else x++;"},
		{"switch", "switch (x) { case 1: x++; break; default: x--; }"},
		{"while", "while (x) x--;"},
		{"for", "for (;x;) x--;"},
		{"assert", "assert(x);"},
		{"printf", "printf(\"%d\", x);"},
		{"cast", "x = (double)x;"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plain, err := Analyze("x.c", "int f(int x) { "+tc.statement+" }")
			if err != nil {
				t.Fatal(err)
			}
			labelled, err := Analyze("x.c", "int f(int x) { first: /* label */ finish: "+tc.statement+" }")
			if err != nil {
				t.Fatal(err)
			}
			clean := strings.ReplaceAll(labelled.Draft, "first: /* label */ finish:", "")
			if strings.Join(strings.Fields(clean), " ") != strings.Join(strings.Fields(plain.Draft), " ") {
				t.Fatalf("labels changed transformation:\nplain: %s\nlabelled: %s", plain.Draft, labelled.Draft)
			}
		})
	}
}

func TestStatementLabelBoundaries(t *testing.T) {
	for _, tc := range []struct {
		source string
		count  int
	}{
		{"{ first: /* comment */ second: f(); }", 2},
		{"{ if (x) finish: f(); }", 1},
		{"{ x = flag ? yes : no; }", 0},
		{"{ x = flag ? yes : flag ? next : last; }", 0},
		{"struct S { unsigned field : 3; };", 0},
		{"{ ns::f(); }", 0},
	} {
		tokens, err := Lex(tc.source)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(statementLabelEnds(tokens)); got != tc.count {
			t.Fatalf("%s: got %d labels, want %d", tc.source, got, tc.count)
		}
	}
}
