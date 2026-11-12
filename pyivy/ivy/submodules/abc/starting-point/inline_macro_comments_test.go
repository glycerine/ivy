package startingpoint

import (
	"strings"
	"testing"
)

func TestInlineMacroComments(t *testing.T) {
	source := `#define ADD(a,b) ((a)+(b))
#define DOUBLE(x) ADD(x,x)
#define EACH(i,n) for (i=0;i<n;i++)
int sum(int x) {
    return DOUBLE(x);
}
void walk() {
    int i=0;
    EACH(i,3) visit(i);
}
`
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"    // inlining of macro DOUBLE\n    // inlining of macro ADD\n    return",
		"    // inlining of macro EACH\n    for i=0;i<3;i++ { visit(i); }",
	} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if strings.Contains(r.Draft, "__starting_point_inline_macro") || strings.Contains(r.Draft, "func EACH(") {
		t.Fatal(r.Draft)
	}
	if !strings.Contains(r.Functions[0].Draft, "// inlining of macro DOUBLE") {
		t.Fatal(r.Functions)
	}
	for _, f := range r.Functions {
		if strings.Contains(f.Original, "// inlining of macro") {
			t.Fatal("original source changed")
		}
	}
}

func TestInlineMacroCommentSameLineAndArguments(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define ID(x) (x)
#define INC(x) ((x)+1)
int main(void) { printf("%d", INC(ID(4))); return 0; }
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"INC", "ID"} {
		if !strings.Contains(r.Draft, "// inlining of macro "+name+"\n") {
			t.Fatal(r.Draft)
		}
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "5" {
		t.Fatal(output, r.Draft)
	}
}

func TestDefaultLoopHeaderMacroOwnsBody(t *testing.T) {
	r, err := AnalyzeWithOptions("x.c", `#define EACH(i,n) for (i=0;i<n;i++)
int main(void) {
    int i=0,total=0;
    EACH(i,5) {
        if (i==3) break;
        total += i;
    }
    printf("%d",total);
    return 0;
}
`, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "// inlining of macro EACH\n") || strings.Contains(r.Draft, "func EACH(") {
		t.Fatal(r.Draft)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "3" {
		t.Fatal(output, r.Draft)
	}
}

func TestRetainedMacrosHaveNoInliningComment(t *testing.T) {
	r, err := Analyze("x.c", "#define S(x) #x\nvoid f() { use(S(x)); }\n")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Draft, "// inlining of macro S") || !strings.Contains(r.Draft, "S(x)") {
		t.Fatal(r.Draft)
	}
}
