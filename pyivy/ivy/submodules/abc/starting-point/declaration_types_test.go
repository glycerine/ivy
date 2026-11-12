package startingpoint

import (
	"strings"
	"testing"
)

func TestNamedLocalDeclarations(t *testing.T) {
	source := `typedef struct solver solver;
solver *create(void) { solver *p; solver *copy = p; return p; }
void use(widget *arg) { widget *local; int x; x * arg; return; }
void other(void) { solver /* keep */ *p; }
`
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var p *solver;", "var copy *solver", "var local *widget;", "x * arg;", "return p;", "solver /* keep */ *p;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestCommaDeclarationsAndZeroInitializers(t *testing.T) {
	r, err := Analyze("x.c", `void f(void) {
 int i, Lit, RetValue = 0;
 int zero = 0;
 int first = next(1, 2), second = first + 1;
 int *pointer, value;
 int kept /* comment */ = 0;
 result = 0;
 }`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var i, Lit, RetValue int32;", "var zero int32;", "var first int32 = next(1, 2);", "var second int32 = first + 1;", "var pointer *int32;", "var value int32;", "/* comment */ = 0;", "result = 0;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestHeaderNamedTypeDeclarations(t *testing.T) {
	r, err := Analyze("x.c", `void f(void) {
 ProgressBar *pProgress;
 MyWidget *first, *second;
 lower * operand;
 ALL_CAPS * operand;
 }`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var pProgress *ProgressBar;", "var first, second *MyWidget;", "lower * operand;", "ALL_CAPS * operand;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}
