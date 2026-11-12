package startingpoint

import (
	"strings"
	"testing"
)

func TestAllocationDeclarationCombined(t *testing.T) {
	r, err := Analyze("x.c", `void f(void) {
 Xyz_ManTry_t *p;
 p = ((Xyz_ManTry_t *) calloc((size_t)(1), sizeof(Xyz_ManTry_t)));
 use(p);
}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "p := &Xyz_ManTry_t{};") || strings.Contains(r.Draft, "var p ") {
		t.Fatal(r.Draft)
	}
}

func TestAllocationDeclarationFirstUseAndScope(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		combine    bool
	}{
		{"adjacent", "var p *T; p = &T{}; use(p);", true},
		{"earlier", "var p *T; var n int32; n=3; touch(n); p = &T{}; use(p);", true},
		{"other allocation", "var p *T; q = &U{}; p = &T{};", true},
		{"new", "var p *T; p = new(T);", true},
		{"comment", "var /* declaration */ p *T; // keep\n// inlining of macro ABC_CALLOC\np = &T{};", true},
		{"read", "var p *T; use(p); p = &T{};", false},
		{"address", "var p *T; save(&p); p = &T{};", false},
		{"earlier assignment", "var p *T; p = other; p = &T{};", false},
		{"different type", "var p *T; p = &U{};", false},
		{"initialized", "var p *T = other; p = &T{};", false},
		{"grouped", "var p, q *T; p = &T{};", false},
		{"if body", "var p *T; if ok { p = &T{}; } use(p);", false},
		{"if initializer", "var p *T; if p = &T{}; ok {} use(p);", false},
		{"loop header", "var p *T; for ; ok; p = &T{} {}", false},
		{"block", "var p *T; { touch(); } p = &T{};", false},
		{"shadow", "var p *T; { var p *T; use(p); } p = &T{};", false},
		{"label", "var p *T; finish: p = &T{};", false},
		{"goto", "var p *T; goto finish; p = &T{}; finish: use(p);", false},
		{"defer", "var p *T; defer use(p); p = &T{};", false},
		{"directive", "var p *T;\n#if UNKNOWN\n#endif\np = &T{};", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "func f() { " + tc.body + " }"
			got, err := combineAllocationDeclarations(source)
			if err != nil {
				t.Fatal(err)
			}
			if tc.combine {
				if strings.Contains(got, "var p ") || !strings.Contains(got, "p := ") {
					t.Fatal(got)
				}
				if tc.name == "comment" {
					for _, s := range []string{"/* declaration */", "// keep", "// inlining of macro ABC_CALLOC"} {
						if !strings.Contains(got, s) {
							t.Fatal(got)
						}
					}
				}
			} else if got != source {
				t.Fatalf("changed unsafe case:\n%s", got)
			}
		})
	}
	source := "var p *T;\nfunc f() { p = &T{}; }"
	got, err := combineAllocationDeclarations(source)
	if err != nil || got != source {
		t.Fatal("changed package scope", err, got)
	}
}

func TestAllocationDeclarationMultipleLocals(t *testing.T) {
	source := "func f() { var p *T; var q *U; p = &T{}; q = &U{}; use(p,q); }"
	got, err := combineAllocationDeclarations(source)
	if err != nil || !strings.Contains(got, "p := &T{};") || !strings.Contains(got, "q := &U{};") {
		t.Fatal(err, got)
	}
	again, err := combineAllocationDeclarations(got)
	if err != nil || again != got {
		t.Fatal("not idempotent", err, again)
	}
}

func TestEarlierAllocationDeclarationRuntime(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define ABC_CALLOC(type,num) ((type *)calloc((size_t)(num),sizeof(type)))
typedef struct T T;
struct T { int value; };
int main(void) {
 T *p;
 int n=0;
 n=7;
 p=ABC_CALLOC(T,1);
 p->value=n;
 printf("%d",p->value);
 return 0;
}`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Draft, "var p ") || !strings.Contains(r.Draft, "p:=&T{};") || !strings.Contains(r.Draft, "// inlining of macro ABC_CALLOC") {
		t.Fatal(r.Draft)
	}
	if !strings.Contains(r.Functions[0].Draft, "p:=&T{};") {
		t.Fatal("missing function metadata")
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "7" {
		t.Fatal(output, r.Draft)
	}
}
