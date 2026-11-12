package startingpoint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
	"testing"
)

func TestDecimalAndScientificDefines(t *testing.T) {
	source := `#define REO_REORDER_LIMIT      1.15 // keep quality note
#define SMALL 1e-3
#define LARGE 2.5E+12
#define WHOLE 1.
#define HALF .5
#define SIGNED - 3.25e-2
#define POSITIVE +.25E2
#define SINGLE 1.25f
#define SINGLE_EXP 2E3F
#define COMMENT /* note */ .125 // tail
#define UNSIGNED 42u
`
	r := checkedDraft(t, source, Options{Package: "p"})
	for _, want := range []string{
		"const REO_REORDER_LIMIT =      1.15 // keep quality note",
		"const SMALL = 1e-3", "const LARGE = 2.5E+12", "const WHOLE = 1.",
		"const HALF = .5", "const SIGNED = - 3.25e-2", "const POSITIVE = +.25E2",
		"const SINGLE float32 = 1.25", "const SINGLE_EXP float32 = 2E3",
		"const COMMENT = /* note */ .125 // tail", "const UNSIGNED uint32 = 42",
	} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, "constants.go", r.Draft, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	cfg := types.Config{}
	pkg, err := cfg.Check("p", files, []*ast.File{file}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if value := pkg.Scope().Lookup("REO_REORDER_LIMIT").(*types.Const).Val().ExactString(); value != "23/20" {
		t.Fatal(value)
	}
	if value := pkg.Scope().Lookup("SMALL").(*types.Const).Val().ExactString(); value != "1/1000" {
		t.Fatal(value)
	}
	if typ := pkg.Scope().Lookup("SINGLE").Type().String(); typ != "float32" {
		t.Fatal(typ)
	}
}

func TestNonLiteralFloatDefinesRemain(t *testing.T) {
	source := `#define EXPR (1.15 * limit)
#define FUNCTION(x) 1.5 * (x)
#define BAD_EXP 1e+
#define BAD_SUFFIX 1.0u
#define BAD_FLOAT 1f
#define LONG_DOUBLE 1.15L
// #define COMMENTED 1.15
`
	got, err := integerDefines(source)
	if err != nil || got != source {
		t.Fatal(got, err)
	}
}

func TestDecimalDefinesRespectOpaqueRegionsAndMetadata(t *testing.T) {
	r, err := Analyze("x.c", `#if UNKNOWN
#define OPAQUE 1.15
#endif
double f(void) {
#define LOCAL 1.25e2
 return LOCAL;
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "#define OPAQUE 1.15") || !strings.Contains(r.Draft, "const LOCAL = 1.25e2") || !strings.Contains(r.Functions[0].Draft, "const LOCAL = 1.25e2") {
		t.Fatal(r.Draft, r.Functions)
	}
	source, err := os.ReadFile("../src/bdd/reo/reo.h")
	if err != nil {
		t.Fatal(err)
	}
	r, err = Analyze("reo.h", string(source))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "const REO_REORDER_LIMIT =      1.15  // determines the quality/runtime trade-off") {
		t.Fatal("REO limit was not converted")
	}
}
