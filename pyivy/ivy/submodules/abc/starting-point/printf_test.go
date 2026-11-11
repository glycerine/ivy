package startingpoint

import (
	"strings"
	"testing"
)

func TestFormattedCalls(t *testing.T) {
	r, err := Analyze("x.c", `void f(void) {
 // printf("unchanged");
 printf("value %d", value());
 fprintf(out, "value %d", value());
 sprintf(buffer, "%d", value());
 count = sprintf(buffer, "%d", value());
 obj.printf("method");
}`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`import "fmt"`, `// printf("unchanged");`, `fmt.Printf("value %d", value());`, `fmt.Fprintf(out, "value %d", value());`, `buffer = fmt.Sprintf("%d", value());`, `count = sprintf(buffer, "%d", value());`, `obj.printf("method");`} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if strings.Count(r.Draft, `import "fmt"`) != 1 {
		t.Fatal(r.Draft)
	}
	if !strings.Contains(r.Functions[0].Draft, "fmt.Printf") {
		t.Fatal("function metadata not updated")
	}
}

func TestExistingFmtImports(t *testing.T) {
	for _, imp := range []string{`import "fmt"`, "import (\n\"fmt\"\n)", `import f "fmt"`, `import . "fmt"`, `import _ "fmt"`} {
		source := "package p\n" + imp + "\nfunc f() { printf(\"hi\"); }"
		got, err := rewriteLibraryCalls(source, true)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(got, `"fmt"`) != 1 || strings.Contains(got, `printf(`) {
			t.Fatal(got)
		}
		again, err := rewriteLibraryCalls(got, true)
		if err != nil || again != got {
			t.Fatal("not idempotent", again, err)
		}
	}
}

func TestOpaquePrintfUntouched(t *testing.T) {
	source := "#if X\nvoid f(void) { printf(\"hi\"); }\n#endif\n"
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, source) || strings.Contains(r.Draft, `import "fmt"`) {
		t.Fatal(r.Draft)
	}
}
