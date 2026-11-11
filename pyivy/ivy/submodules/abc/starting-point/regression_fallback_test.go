package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestReportedDraftDelimiterFailures(t *testing.T) {
	for _, name := range []string{"bdd/cas/casCore.c", "bdd/cas/casDec.c", "opt/fret/fretFlow.c"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("../src/" + name)
			if err != nil {
				t.Fatal(err)
			}
			projected, regions, err := projectConditionals(string(source))
			if err != nil {
				t.Fatal(err)
			}
			r, err := analyzeCore(name, projected, Options{opaque: regions})
			if err != nil {
				t.Fatal(err)
			}
			ts, err := Lex(r.Draft)
			if err != nil {
				t.Fatal(err)
			}
			_, err = MatchDelimiters(ts)
			if err != nil {
				e := err.(*syntaxError)
				a, b := e.Offset-170, e.Offset+170
				if a < 0 {
					a = 0
				}
				if b > len(r.Draft) {
					b = len(r.Draft)
				}
				t.Fatalf("%v, draft near failure:\n%s", err, r.Draft[a:b])
			}
			if _, err = Analyze(name, string(source)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMacroHashOperatorsAreNotDirectives(t *testing.T) {
	source := "  /* note */ #define REAL(x) #x\nfunc f(x int) {\n printf(#x);\n (p->x##_field);\n}\n"
	tokens, err := Lex(source)
	if err != nil {
		t.Fatal(err)
	}
	directives := 0
	for _, tok := range tokens {
		if tok.Kind == "directive" {
			directives++
		}
	}
	if directives != 1 {
		t.Fatalf("got %d directives", directives)
	}
	if _, err := MatchDelimiters(tokens); err != nil {
		t.Fatal(err)
	}
	if got := Mask(tokens); !strings.Contains(got, "#x") || !strings.Contains(got, "x##_field") {
		t.Fatal(got)
	}
}

func TestTransformedDiagnosticCoordinates(t *testing.T) {
	err := transformedDiagnostic("x.c", "formatted-call rewrite", "{\n  ]", &syntaxError{"mismatched delimiter", 4})
	for _, want := range []string{"not original-source coordinates", "line 2, column 3", "byte offset 4"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatal(err)
		}
	}
}
