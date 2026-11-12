package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestMacroWholeExpressionParentheses(t *testing.T) {
	for _, c := range [][2]string{
		{"(((x + y)))", "x + y"},
		{"(x) + (y)", "(x) + (y)"},
		{"((reo_unit *)((ABC_PTRUINT_T)(u) & ~01))", "(reo_unit *)((ABC_PTRUINT_T)(u) & ~01)"},
		{"/* before */ ((x)) /* after */", "/* before */ x /* after */"},
	} {
		if got := unwrapMacroExpression(c[0]); got != c[1] {
			t.Fatalf("%q -> %q; want %q", c[0], got, c[1])
		}
	}
}

func TestRealReoMacroReturn(t *testing.T) {
	source, err := os.ReadFile("../src/bdd/reo/reo.h")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("reo.h", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"// #define Unit_Regular(u)     ((reo_unit *)((ABC_PTRUINT_T)(u) & ~01))",
		"func Unit_Regular(u int) int {",
		"return (reo_unit *)((ABC_PTRUINT_T)(u) & ~01)",
	} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal("missing", want)
		}
	}
}

func TestMacroValuesRuntime(t *testing.T) {
	r := checkedDraft(t, `#define ADD(x,y) (((x)+(y)))
#define PICK(x,y) ((x = x + 1), (y))
#define ZERO() (0)
#define NOTE(x) /* before
comment */ ((x)) // after
#define LOOP(x) for (; x < 2;) x++;
`, Options{Package: "main"})
	output, _ := runLibraryProgram(t, r.Draft+`func main() { println(ADD(2,3), PICK(1,4), ZERO(), NOTE(7), LOOP(0)) }
`)
	if output != "5 4 0 7 0\n" {
		t.Fatal(output)
	}
}
