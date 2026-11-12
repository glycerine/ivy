package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestPostfixIndexAssignment(t *testing.T) {
	r, err := Analyze("x.c", `void f(void) { pD->pNodes[c] = pD0->pNodes[i++]; }`)
	if err != nil {
		t.Fatal(err)
	}
	want := "pD.pNodes[c] = pD0.pNodes[i]; i++;"
	if !strings.Contains(r.Draft, want) {
		t.Fatalf("missing %q:\n%s", want, r.Draft)
	}
}

func TestPostfixIndexBoundaries(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"{ dst = src[i++]; }", "{ dst = src[i]; i++; }"},
		{"{ dst = src[i--]; }", "{ dst = src[i]; i--; }"},
		{"{ finish: dst = src[i++]; }", "{ finish: dst = src[i]; i++; }"},
		{"{ dst[0] = src.rows[2][i++]; }", "{ dst[0] = src.rows[2][i]; i++; }"},
		{"{ dst = src[i /* index */ ++]; // tail\n}", "{ dst = src[i /* index */ ]; i++; // tail\n}"},
		{"{ dst = src[i++] + i; }", ""},
		{"{ dst[i] = src[i++]; }", ""},
		{"{ i = src[i++]; }", ""},
		{"{ dst[f()] = src[i++]; }", ""},
		{"{ dst = get()[i++]; }", ""},
		{"{ dst = src[++i]; }", ""},
		{"{ dst = src[i++] + src[j++]; }", ""},
		{"{ use(src[i++]); }", ""},
		{"{ return src[i++]; }", ""},
		{"{ for (; dst = src[i++];) {} }", ""},
		{"{ for ; dst = src[i++]; {} }", ""},
		{"{ if dst = src[i++]; dst != 0 {} }", ""},
		{"{ // dst = src[i++];\n use(\"dst = src[i++];\"); }", ""},
	} {
		got, err := lowerPostfixIndexAssignments(tc.source)
		if err != nil {
			t.Fatal(err)
		}
		want := tc.want
		if want == "" {
			want = tc.source
		}
		if got != want {
			t.Fatalf("%s: got %q, want %q", tc.source, got, want)
		}
	}
}

func TestPostfixIndexRuntime(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define COPY(dst,src,i) dst = src[i++]
int main(void) {
 int src[3] = {10,20,30};
 int dst[3] = {0};
 int i=0;
 if (1) COPY(dst[0],src,i);
 dst[1] = src[i++];
 dst[2] = src[i--];
 printf("%d %d %d %d",dst[0],dst[1],dst[2],i);
 return 0;
}`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "10 20 30 1" {
		t.Fatal(output, r.Draft)
	}
}

func TestAigDomsPostfixIndices(t *testing.T) {
	path := "../src/aig/aig/aigDoms.c"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := AnalyzeWithOptions(path, string(source), Options{IncludePaths: []string{"../src"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"pD.pNodes[c] = pD0.pNodes[i]; i++;", "pD.pNodes[c] = pD1.pNodes[k]; k++;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if strings.Contains(r.Draft, "pNodes[i++]") || strings.Contains(r.Draft, "pNodes[k++]") {
		t.Fatal("postfix index retained")
	}
	found := false
	for _, f := range r.Functions {
		if strings.Contains(f.Draft, "pD0.pNodes[i]; i++;") {
			found = true
		}
	}
	if !found {
		t.Fatal("function metadata missing fix")
	}
}
