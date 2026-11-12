package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestRealReoMacroDefinitionComment(t *testing.T) {
	source, err := os.ReadFile("../src/bdd/reo/reo.h")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("../src/bdd/reo/reo.h", string(source))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "// #define Unit_Regular(u)     ((reo_unit *)((ABC_PTRUINT_T)(u) & ~01))") || strings.Contains(r.Draft, "func Unit_Regular(") {
		t.Fatal(r.Draft)
	}
}

func TestMacroValuesRuntime(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define ADD(x,y) (((x)+(y)))
#define ZERO() (0)
#define NOTE(x) /* before
comment */ ((x)) // after
#define LOOP(x) for (; x < 2;) x++;
int main(void) { int n=0; int noted; noted = NOTE(7);
 LOOP(n) printf("%d %d %d %d", ADD(2,3), ZERO(), noted, n); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "5 0 7 2" {
		t.Fatal(output, r.Draft)
	}
}
