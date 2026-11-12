package startingpoint

import (
	"go/format"
	"os"
	"strings"
	"testing"
)

func TestForHeaderParentheses(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"for (i = 0; i < size; i++) {}", "for i = 0; i < size; i++ {}"},
		{"for(i=0; i<(n+1); i++){}", "for i=0; i<(n+1); i++{}"},
		{"for (ready()) {}", "for ready() {}"},
		{"for ((a || b) && c) {}", "for (a || b) && c {}"},
		{"for (a || b) && c {}", "for (a || b) && c {}"},
		{"for /*keep*/ (i=0; i<n; i++) {}", "for /*keep*/ i=0; i<n; i++ {}"},
		{`log("for (x)"); // for (x)`, `log("for (x)"); // for (x)`},
	} {
		got, err := unwrapControlHeaders(tc.source)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("got %q; want %q", got, tc.want)
		}
	}
}

func TestGeneratedFindCanBeFormatted(t *testing.T) {
	source := `static inline unsigned find(vec_t *p, unsigned entry) {
    unsigned i;
    for (i = 0; i < p->size; i++)
        if (p->data[i] == entry) return 1;
    return 0;
}`
	r, err := Analyze("p/file.h", source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "for i = 0; i < p.size; i++ {") {
		t.Fatal(r.Draft)
	}
	if _, err := format.Source([]byte(r.Draft)); err != nil {
		t.Fatalf("gofmt: %v\n%s", err, r.Draft)
	}
}

func TestIfHeaderParentheses(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"if ( ready(p, k) ) { break; }", "if ready(p, k) { break; }"},
		{"if(x) {} else if ((a || b) && c) {}", "if x {} else if (a || b) && c {}"},
		{"if (a || b) && c {}", "if (a || b) && c {}"},
		{"if (\n ready()\n) {}", "if ready() {}"},
		{"if /* keep */ (ready()) {}", "if /* keep */ ready() {}"},
		{`log("if (x)"); // if (x)`, `log("if (x)"); // if (x)`},
	} {
		got, err := unwrapControlHeaders(tc.source)
		if err != nil || got != tc.want {
			t.Fatalf("got %q, want %q: %v", got, tc.want, err)
		}
	}
}

func TestAigPackIfHeader(t *testing.T) {
	source, err := os.ReadFile("../src/aig/aig/aigPack.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("aigPack.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "if Aig_ManPackAddPatternTry( p, k, vLits ) != 0 {") {
		t.Fatal("AIG condition not unwrapped")
	}
}

func TestConditionOnlyFor(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"for  ; q < p.nSupp-1;  {}", "for q < p.nSupp-1 {}"},
		{"for ; ; {}", "for {}"},
		{"for ; next(a,b); {}", "for next(a,b) {}"},
		{"for i=0; i<n; {}", "for i=0; i<n; {}"},
		{"for ; i<n; i++ {}", "for ; i<n; i++ {}"},
		{"for /*init*/ ; ready; /*post*/ {}", "for /*init*/  ready /*post*/ {}"},
	} {
		got, err := simplifyConditionOnlyFor(tc.source)
		if err != nil || got != tc.want {
			t.Fatalf("%q -> %q, want %q (%v)", tc.source, got, tc.want, err)
		}
	}
}

func TestConditionOnlyForRuntime(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `int main(void) { int i=0; for (; i<3;) i++; printf("%d\n",i); return 0; }`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "3\n" || !strings.Contains(r.Draft, "for i<3 {") {
		t.Fatal(output, r.Draft)
	}
	source, err := os.ReadFile("../src/bdd/reo/reoSift.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err = Analyze("reoSift.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "for q < p.nSupp-1 {") {
		t.Fatal("reoSift condition-only loop not simplified")
	}
}
