package startingpoint

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlineMacrosBasic(t *testing.T) {
	source := "#define ADD(a,b) ((a)+(b))\nint sum(void) { return ADD(2,3); }\n"
	r, err := AnalyzeWithOptions("x.c", source, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "return ((2)+(3))") || strings.Contains(r.Draft, "func ADD(") {
		t.Fatal(r.Draft)
	}
	ordinary, err := Analyze("x.c", source)
	if err != nil || ordinary.Draft != r.Draft {
		t.Fatal(err, ordinary.Draft)
	}
}

func inlineSource(t *testing.T, source string) (string, []Issue) {
	t.Helper()
	p := &macroInliner{opts: Options{}, macros: map[string]inlineMacro{}, visiting: map[string]bool{}, once: map[string]bool{}}
	out, _, err := p.process("x.c", source, true)
	if err != nil {
		t.Fatal(err)
	}
	return out, p.issues
}

func TestInlineMacroSubstitution(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{"nested", "#define ADD(a,b) ((a)+(b))\n#define TWICE(x) ADD(x,x)\nTWICE(ADD(1,2))", "((((1)+(2)))+(((1)+(2))))"},
		{"same macro argument", "#define ID(x) (x)\nID(ID(3))", "((3))"},
		{"zero args", "#define ZERO() 0\nZERO()", "\n0"},
		{"empty arg", "#define EMPTY(x) x\nEMPTY()", "\n"},
		{"empty body", "#define NOTHING()\nNOTHING();", "\n;"},
		{"balanced args", "#define PAIR(a,b) a + b\nPAIR(f(1,2), a[g(3,4)])", "f(1,2) + a[g(3,4)]"},
		{"braced args", "#define ID(x) x\nID((T){1,2})", "\n(T){1,2}"},
		{"token boundaries", "#define F(x) x + xx + \"x\" /* x */\nF(foo)", "foo + xx + \"x\" /* x */"},
		{"precedence", "#define RAW(x) x + 1\n2 * RAW(3)", "2 * 3 + 1"},
		{"multiline", "#define SUM(a,b) ((a) + \\\n(b))\nSUM(2,3)", "((2) + (3))"},
		{"crlf", "#define SUM(a,b) ((a) + \\\r\n(b))\r\nSUM(2,3)", "((2) + (3))"},
		{"unused comment", "#define DROP(x) 4\nDROP(/* retain */ 2)", "/* retain */\n4"},
		{"between comment", "#define ID(x) x\nID /* retain */ (2)", "/* retain */\n2"},
		{"line comment", "#define ID(x) x // explanation\nID(2);", "2 // explanation\n;"},
		{"string commas", "#define ID(x) x\nID(\"a,b)\")", "\n\"a,b)\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, issues := inlineSource(t, tc.source)
			if !strings.HasSuffix(out, tc.want) || len(issues) != 0 {
				t.Fatalf("got %q, want suffix %q; issues %+v", out, tc.want, issues)
			}
		})
	}
}

func TestInlineMacroVisibilityAndProtection(t *testing.T) {
	source := `F(1);
#define F(x) ((x)+1)
F(2);
"F(3)"; /* F(4) */ // F(5)
#undef F
F(6);
#define F(x) ((x)+7)
F(8);
#define F 9
F(10);
`
	out, issues := inlineSource(t, source)
	for _, want := range []string{"F(1);", "((2)+1);", `"F(3)"; /* F(4) */ // F(5)`, "F(6);", "((8)+7);", "F(10);"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q:\n%s", want, out)
		}
	}
	if len(issues) != 0 {
		t.Fatal(issues)
	}
}

func TestInlineMacroRetainedReview(t *testing.T) {
	for _, tc := range []struct{ name, source, reason string }{
		{"variadic", "#define F(...) printf(__VA_ARGS__)\nF(1,2)", "Variadic"},
		{"stringify", "#define F(x) #x\nF(1)", "Stringification"},
		{"paste", "#define F(x) x ## x\nF(1)", "token pasting"},
		{"arity", "#define F(x,y) x+y\nF(1)", "argument count"},
		{"recursive", "#define F(x) F(x)\nF(1)", "Recursive"},
		{"mutual", "#define F(x) G(x)\n#define G(x) F(x)\nF(1)", "Recursive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, issues := inlineSource(t, tc.source)
			if !strings.HasSuffix(out, "F(1)") && tc.name != "variadic" {
				t.Fatal(out)
			}
			if len(issues) == 0 || !strings.Contains(issues[0].Message, tc.reason) {
				t.Fatal(issues)
			}
		})
	}
}

func TestInlineMacroConfiguredAndOpaque(t *testing.T) {
	source := `#define F(x) ((x)+1)
#if FLAG
#undef F
#define F(x) ((x)+2)
#else
#define F(x) ((x)+3)
#endif
int f(void) { return F(4); }
`
	for _, tc := range []struct {
		defines map[string]int64
		want    string
	}{{nil, "return F(4);"}, {map[string]int64{"FLAG": 1}, "return ((4)+2);"}, {map[string]int64{"FLAG": 0}, "return ((4)+3);"}} {
		r, err := AnalyzeWithOptions("x.c", source, Options{Defines: tc.defines})
		if err != nil || !strings.Contains(r.Draft, tc.want) {
			t.Fatal(err, r.Draft)
		}
		if tc.defines == nil {
			if !strings.Contains(r.Draft, "#if FLAG\n#undef F") {
				t.Fatal(r.Draft)
			}
			found := false
			for _, issue := range r.Issues {
				if issue.Category == "macro-inline-review" {
					found = true
				}
			}
			if !found {
				t.Fatal("missing unknown-branch review")
			}
		}
	}
}

func TestInlineMacroPipelineAndCoordinates(t *testing.T) {
	source := `#define SET(x,v) if (x) x = v
#define LOOP(i,n) for (i = 0; i < n; i++)
#define SAY(x) printf("%d",x)
#define PICK(x) ((x) ? 3 : 4)
int f(int x) {
 int i = 0;
 finish: SET(x, PICK(x));
 LOOP(i, 2) SAY(i);
 return x;
}
#if UNKNOWN
opaque bytes
#endif
`
	r, err := AnalyzeWithOptions("x.c", source, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"func f(", "var i int32", "finish: if", "fmt.Printf", "for i = 0; i < 2; i++", "x = 3;", "#if UNKNOWN"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if len(r.Functions) != 1 || r.Functions[0].Original != source[r.Functions[0].Start:r.Functions[0].End] || !strings.Contains(r.Functions[0].Original, "LOOP(i, 2)") {
		t.Fatal(r.Functions)
	}
	for _, c := range r.Conditionals {
		if source[c.Location.ByteOffset:c.Location.ByteOffset+3] != "#if" {
			t.Fatal(c)
		}
	}
	if r.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(source))) {
		t.Fatal("hash changed")
	}
}

func TestInlineMacroRuntime(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define ADD(a,b) ((a)+(b))
#define REPEAT(i,n) for (i=0;i<n;i++)
#define INC(x) do { x++; } while (0)
int main(void) { int i=0, n=0; REPEAT(i,3) INC(n); printf("%d",ADD(n,4)); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "7" {
		t.Fatal(output, r.Draft)
	}
}

func TestInlineMacroIncludes(t *testing.T) {
	root := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("include/helper.h", "#ifndef HELPER_H\n#define HELPER_H\n#define PLUS(x) ((x)+1)\n#include \"cycle.h\"\n#endif\n")
	write("include/cycle.h", "#pragma once\n#include \"helper.h\"\n#define WRAP(x) PLUS(x)\n")
	source := "#include \"include/helper.h\"\nint f(void) { return WRAP(8); }\n"
	write("sub/input.c", source)
	r, err := AnalyzeWithOptions(filepath.Join(root, "sub/input.c"), source, Options{IncludePaths: []string{root}})
	if err != nil || !strings.Contains(r.Draft, "return ((8)+1);") {
		t.Fatal(err, r.Draft)
	}
	out := filepath.Join(root, "out")
	report, err := TranslateTree(root, out, Options{})
	if err != nil || len(report.Files) != 3 {
		t.Fatal(err, report)
	}
	data, err := os.ReadFile(filepath.Join(out, "sub/input.go"))
	if err != nil || !strings.Contains(string(data), "return ((8)+1);") {
		t.Fatal(err, string(data))
	}
}

func TestInlineMacroRetainedIncludeAndCoordinates(t *testing.T) {
	source := "#include \"missing-inline-header.h\"\n#define BIG(x) ((x)+(x)+(x))\nint f() { return BIG(3); }\n#define BAD(x,y) x+y\nint g() { return BAD(1); }\n"
	r, err := AnalyzeWithOptions("x.c", source, Options{})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, issue := range r.Issues {
		if issue.Category == "macro-inline-review" && issue.Symbol == "BAD" {
			found = true
			if source[issue.Start:issue.End] != "BAD(1)" {
				t.Fatal(issue)
			}
		}
	}
	if !found {
		t.Fatal("missing arity diagnostic")
	}
	bad := strings.Replace(source, "return BAD(1);", "return BAD(1); ]", 1)
	_, err = AnalyzeWithOptions("x.c", bad, Options{})
	var diagnostic *SourceError
	if !errors.As(err, &diagnostic) || diagnostic.Location.ByteOffset != strings.Index(bad, "]") {
		t.Fatal(err)
	}
}

func TestInlineMacroExpansionLimits(t *testing.T) {
	var source strings.Builder
	for i := 0; i < 70; i++ {
		fmt.Fprintf(&source, "#define M%d(x) M%d(x)\n", i, i+1)
	}
	source.WriteString("M0(1)")
	_, issues := inlineSource(t, source.String())
	if len(issues) == 0 || !strings.Contains(issues[0].Message, "limit") {
		t.Fatal(issues)
	}
	p := &macroInliner{macros: map[string]inlineMacro{"BIG": {params: []string{"x"}, body: strings.Repeat("x", macroExpansionLimit+1)}}}
	out, _, notes := p.expand("BIG(1)", map[string]bool{}, 0)
	if out != "BIG(1)" || len(notes) != 1 {
		t.Fatal(len(out), notes)
	}
}

func TestInlineMacroUnknownConditionalCallsUntouched(t *testing.T) {
	source := "#define ID(x) x\n#if UNKNOWN\nint a() { return ID(1); }\n#endif\nint b() { return ID(2); }\n"
	r, err := AnalyzeWithOptions("x.c", source, Options{})
	if err != nil || !strings.Contains(r.Draft, "return ID(1);") || !strings.Contains(r.Draft, "return 2;") {
		t.Fatal(err, r.Draft)
	}
}

func TestInlineMacroHeaderGuardOnce(t *testing.T) {
	root := t.TempDir()
	header := filepath.Join(root, "once.h")
	if err := os.WriteFile(header, []byte("#ifndef ONCE_H\n#define ONCE_H\n#define F(x) ((x)+1)\n#endif\n"), 0644); err != nil {
		t.Fatal(err)
	}
	source := "#include \"once.h\"\n#undef F\n#include \"once.h\"\nint f() { return F(1); }\n"
	r, err := AnalyzeWithOptions(filepath.Join(root, "x.c"), source, Options{})
	if err != nil || !strings.Contains(r.Draft, "return F(1);") {
		t.Fatal(err, r.Draft)
	}
}

func TestInlineMacroSideEffectsAndComments(t *testing.T) {
	out, issues := inlineSource(t, "#define TWICE(x) ((x)+(x))\nTWICE(i++ /* argument */)")
	if !strings.HasSuffix(out, "((i++ /* argument */)+(i++ /* argument */))") || len(issues) != 0 {
		t.Fatal(out, issues)
	}
	// Nested unsupported calls must remain visible and produce an invocation-site
	// diagnostic, even when the surrounding supported macro expands.
	source := "#define BAD(...) call(__VA_ARGS__)\n#define WRAP(x) (x)\nWRAP(BAD(1,2))"
	out, issues = inlineSource(t, source)
	if !strings.HasSuffix(out, "(BAD(1,2))") || len(issues) == 0 {
		t.Fatal(out, issues)
	}
	for _, issue := range issues {
		if source[issue.Start:issue.End] != "WRAP(BAD(1,2))" {
			t.Fatal(issue)
		}
	}
}

func TestInlineMacroRealReoConsumer(t *testing.T) {
	path := "../src/bdd/reo/reoTransfer.c"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := AnalyzeWithOptions(path, string(source), Options{IncludePaths: []string{"../src"}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Draft, "Unit_Regular(pUnit)") || !strings.Contains(r.Draft, "ABC_PTRUINT_T") {
		t.Fatal("header macro not expanded", r.Draft)
	}
	if !r.InlineMacros {
		t.Fatal("inline policy missing from metadata")
	}
}
