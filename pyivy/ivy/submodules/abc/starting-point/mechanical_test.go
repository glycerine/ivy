package startingpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMechanicalControlFlowTogether(t *testing.T) {
	source := `enum { FIRST, SECOND, THIRD };
int f(int n) {
 int result = 0;
 switch(n) {
 case FIRST:
 case SECOND: result += 2;
 case THIRD: result += 3; break;
 default: result = 9;
 }
 do { result--; } while(result > 5);
 for(1) { break; }
 if (++result == 4) result += 10;
 return (int)result;
}
int main(void) { printf("%d %d %d\n",f(0),f(2),f(3)); return 0; }
`
	r, err := AnalyzeWithOptions("x.c", source, Options{Package: "main"})
	if err != nil {
		t.Fatal(err)
	}
	output, _ := runLibraryProgram(t, r.Draft)
	if output != "5 3 6\n" {
		t.Fatal(output, r.Draft)
	}
	if strings.Contains(r.Draft, "do {") || strings.Contains(r.Draft, "switch(") || strings.Contains(r.Draft, "++result") {
		t.Fatal(r.Draft)
	}
}

func TestDoContinueRetainedAndComments(t *testing.T) {
	source := "do { if (x) continue; x++; } while (x < 4);"
	got, err := lowerDoWhile(source)
	if err != nil || got != source {
		t.Fatal(got, err)
	}
	source = `// p->field and for 1
func macro() { p->field = 1; }
var s = "p->field"`
	got, err = mechanicalDraft(source)
	if err != nil || !strings.Contains(got, "p.field = 1") || !strings.Contains(got, `"p->field"`) || !strings.Contains(got, "// p->field") {
		t.Fatal(got, err)
	}
}

func TestSharedTypedefIndex(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	out := filepath.Join(root, "out")
	putSource(t, src, "types.h", "typedef long abctime;\ntypedef struct { int x; } widget;\n")
	putSource(t, src, "consumer.c", "void f(void) { abctime clk; widget *p; x * y; }")
	_, err := TranslateTree(src, out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	draft, err := os.ReadFile(filepath.Join(out, "consumer.go"))
	for _, want := range []string{"var clk abctime;", "var p *widget;", "x * y;"} {
		if err != nil || !strings.Contains(string(draft), want) {
			t.Fatal(string(draft), err)
		}
	}
}
