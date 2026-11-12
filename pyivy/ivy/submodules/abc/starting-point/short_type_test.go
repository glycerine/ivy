package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestShortTypeAcrossDeclarations(t *testing.T) {
	r := checkedDraft(t, `short global;
struct sample { short value; short *pointer; };
short f(short x, short *p) { short local; short array[2]; local=x; array[0]=local; *p=array[0]; return (short)x; }
`, Options{Package: "p"})
	for _, want := range []string{"var global int16;", "value int16;", "pointer *int16;", "func f(x int16, p *int16) int16", "var local int16;", "var array [2]int16;", "return int16(x);"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestReoShortFields(t *testing.T) {
	source, err := os.ReadFile("../src/bdd/reo/reo.h")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("reo.h", string(source))
	if err != nil {
		t.Fatal(err)
	}
	tokens, err := Lex(r.Draft)
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range tokens {
		if tok.Kind == "word" && tok.Text == "short" {
			t.Fatal("unconverted short", tok.Start)
		}
	}
	for _, name := range []string{"lev", "TopRef", "TopRefNew", "n"} {
		found := false
		for i, tok := range tokens {
			if tok.Text != name {
				continue
			}
			j := i + 1
			for j < len(tokens) && (tokens[j].Kind == "space" || tokens[j].Kind == "comment") {
				j++
			}
			if j < len(tokens) && tokens[j].Text == "int16" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("short field not converted", name)
		}
	}
}
