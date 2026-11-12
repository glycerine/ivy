package startingpoint

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestEnumValues(t *testing.T) {
	for _, source := range []string{
		"enum { A, B, C };", "typedef enum { A, B, C } Kind;", "enum Kind { A=3, B, C=8, D };",
		"typedef enum Tag { A=-1, B, C=4, D } Kind;", "enum { A /* first */, B=4, C /*last*/ };",
	} {
		r, err := Analyze("x.c", source)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parser.ParseFile(token.NewFileSet(), "x.go", r.Draft, parser.AllErrors); err != nil {
			t.Fatal(r.Draft, err)
		}
		if strings.Contains(r.Draft, "enum") || !strings.Contains(r.Draft, "const") {
			t.Fatal(r.Draft)
		}
		if strings.Contains(source, "/* first */") && (!strings.Contains(r.Draft, "/* first */") || !strings.Contains(r.Draft, "/*last*/")) {
			t.Fatal(r.Draft)
		}
	}
	draft, err := lowerEnums(`package main
import "fmt"
typedef enum Tag { A=-1, B, C=4, D } Kind;
enum { X, Y, Z };
func main() { fmt.Println(A,B,C,D,X,Y,Z) }
`)
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, draft)
	if output != "-1 0 4 5 0 1 2\n" {
		t.Fatal(output)
	}
}

func TestEnumProtection(t *testing.T) {
	source := `enum Tag { A,B } variable;
const text = "enum { X,Y };";
// enum { C,D };
`
	got, err := lowerEnums(source)
	if err != nil || got != source {
		t.Fatal(got, err)
	}
	r, err := Analyze("x.c", "#if UNKNOWN\ntypedef enum { A, B } Kind;\n#endif\n")
	if err != nil || !strings.Contains(r.Draft, "typedef enum { A, B } Kind;") {
		t.Fatal(r.Draft, err)
	}
}

func TestAigEnum(t *testing.T) {
	source, err := os.ReadFile("../src/aig/aig/aig.h")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("aig.h", string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"type Aig_Type_t int32", "AIG_OBJ_NONE Aig_Type_t = iota", "// 0: non-existent object"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal("missing", want)
		}
	}
}
