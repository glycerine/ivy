package startingpoint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func checkedDraft(t *testing.T, source string, opts Options) Result {
	t.Helper()
	r, err := AnalyzeWithOptions("fixture.h", source, opts)
	if err != nil {
		t.Fatal(err)
	}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "draft.go", r.Draft, parser.ParseComments)
	if err != nil {
		t.Fatalf("%v\n%s", err, r.Draft)
	}
	if _, err := new(types.Config).Check("fixture", fs, []*ast.File{f}, nil); err != nil {
		t.Fatalf("%v\n%s", err, r.Draft)
	}
	return r
}

func TestPackageAndMacroComments(t *testing.T) {
	source := "// keep this comment\n#define twice(x) ((x) + (x))\n#define choose(x) ((x)>0 ? /* yes */ (x) : /* no */ -(x))\n"
	r, err := AnalyzeWithOptions("fixture.h", source, Options{Package: "example"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"package example\n", "// keep this comment", "// #define twice(x)", "// #define choose(x)"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q in %s", want, r.Draft)
		}
	}
	if strings.Contains(r.Draft, "func ") {
		t.Fatal(r.Draft)
	}
	for _, name := range []string{"for", "_", "bad-name"} {
		if _, err := AnalyzeWithOptions("x", "", Options{Package: name}); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
}

func TestNestedTernariesAndLazyBranches(t *testing.T) {
	source := `static inline int select_value(int x) {
 int value /* initializer comment */ = x > 0 ? yes() : (x < 0 ? no() : zero());
 value = x > 1 ? value + 1 : value;
 return x == 99 ? (x > 0 ? value : no()) : value;
}`
	r, err := AnalyzeWithOptions("fixture.h", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "/* initializer comment */") {
		t.Fatal("lost initializer comment")
	}
	if strings.Contains(r.Draft, "?") {
		t.Fatal(r.Draft)
	}
	// Execute generated Go, verifying exactly one branch is evaluated, including
	// nested conditionals. This is stronger than checking output spelling.
	extra := `
var yesCalls, noCalls, zeroCalls int
func yes() int32 { yesCalls++; return 10 }
func no() int32 { noCalls++; return -10 }
func zero() int32 { zeroCalls++; return 0 }
func main() {
 if select_value(2)!=11 || yesCalls!=1 || noCalls!=0 || zeroCalls!=0 {panic("positive branch")}
 if select_value(-1)!=-10 || yesCalls!=1 || noCalls!=1 || zeroCalls!=0 {panic("negative branch")}
 if select_value(0)!=0 || yesCalls!=1 || noCalls!=1 || zeroCalls!=1 {panic("zero branch")}
 if select_value(99)!=11 || yesCalls!=2 || noCalls!=1 {panic("nested branch")}
}
`
	path := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(path, []byte(r.Draft+extra), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("go", "run", path).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s\n%s", err, out, r.Draft)
	}
}

func TestEmbeddedTernaryRetained(t *testing.T) {
	s := "static inline int f(int x) { return call(x > 0 ? 1 : 2); }"
	r, err := Analyze("x", s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "call(x > 0 ? 1 : 2)") {
		t.Fatal("unsafe embedded rewrite")
	}
}

func TestIteratorMacroHeaderExpansion(t *testing.T) {
	body := "for (i = 0; (i < vec_uint_size(vec)) && (((entry) = vec_uint_at(vec, i)), 1); i++)"
	source := "#define vec_uint_foreach(vec, entry, i) " + body + "\nvoid walk(vec_uint_t *vec) { int i, entry; vec_uint_foreach(vec, entry, i) visit(entry); }\n"
	r, err := Analyze("vec/vec_uint.h", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"for i = 0;", "i++ { visit(entry); }"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(r.Draft)
		}
	}
	if strings.Contains(r.Draft, "func vec_uint_foreach(") || strings.Contains(r.Draft, "i++ {}") {
		t.Fatal(r.Draft)
	}
}

func TestMacroDefinitionCommentsPreserved(t *testing.T) {
	for _, body := range []string{"do { /* keep */ x++; } while (0)", "x ## suffix", "#x", "call(x)", "", "x ? yes() : no()"} {
		source := "#define M(x) " + body
		r, err := Analyze("macros/file.h", source+"\n")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(r.Draft, "// "+source) || strings.Contains(r.Draft, "func M(") || strings.Contains(r.Draft, "argument types were guessed") {
			t.Fatal(r.Draft)
		}
	}
}
