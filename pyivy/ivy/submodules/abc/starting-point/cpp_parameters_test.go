package startingpoint

import (
	"strings"
	"testing"
)

func TestQualifiedReferenceAndCallbackParameters(t *testing.T) {
	r, err := Analyze("x.cpp", `void stop(Gluco2::SimpSolver* S) { use(S); }
int callback(Gluco2::SimpSolver &s, int(*fn)(void*, int, int*), void *data) { int i=0; return fn(data, 1, &i); }
void plain(void (*fn)(int), int x) { fn(x); }
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"func stop(S *Gluco2.SimpSolver)", "func callback(s *Gluco2.SimpSolver, fn func(any, int32, *int32) int32, data any) int32", "func plain(fn func(int32), x int32)"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
}

func TestCallbackParameterRuntime(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `int inc(int x) { return x+1; }
int apply(int(*fn)(int), int x) { return fn(x); }
int main(void) { printf("%d", apply(inc, 4)); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "5" {
		t.Fatal(output, r.Draft)
	}
}
