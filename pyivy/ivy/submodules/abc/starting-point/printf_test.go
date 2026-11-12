package startingpoint

import (
	"os"
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

func TestSprintfIntoCharArray(t *testing.T) {
	r, err := Analyze("x.c", `void f(int *pCounter) {
 char Suffix[2000] = {0};
 if (*pCounter)
  sprintf(Suffix, "_%s_%d", Abc_NtkName(pNtk), *pCounter);
 }`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "var Suffix string // was [2000]byte in C") {
		t.Fatalf("destination declaration not converted:\n%s", r.Draft)
	}
	want := `Suffix = fmt.Sprintf("_%s_%d", Abc_NtkName(pNtk), *pCounter);`
	if !strings.Contains(r.Draft, want) {
		t.Fatalf("missing %q:\n%s", want, r.Draft)
	}
	if !strings.Contains(r.Draft, `"fmt"`) || strings.Contains(r.Draft, "func()") {
		t.Fatal(r.Draft)
	}
}

func TestAbcHieSprintf(t *testing.T) {
	source, err := os.ReadFile("../src/base/abc/abcHie.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("abcHie.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "var Suffix string // was [2000]byte in C") {
		t.Fatalf("destination declaration not converted:\n%s", r.Draft)
	}
	want := `Suffix = fmt.Sprintf("_%s_%d", Abc_NtkName(pNtk), *pCounter);`
	if !strings.Contains(r.Draft, want) {
		t.Fatalf("missing %q", want)
	}
	if strings.Count(r.Draft, `"fmt"`) != 1 {
		t.Fatal("expected exactly one fmt import")
	}
	found := false
	for _, f := range r.Functions {
		if strings.Contains(f.Draft, want) {
			found = true
		}
	}
	if !found {
		t.Fatal("function metadata missing converted sprintf")
	}
}

func TestSprintfDestinationScope(t *testing.T) {
	r, err := Analyze("x.c", `
void converted(void) { char buf[20]; sprintf(buf, "%d", 1); sprintf(buf, "%d", 2); }
void unrelated(void) { char buf[30]; use(buf); }
void nested(void) { char buf[40]; { char buf[50]; sprintf(buf, "%d", 3); } use(buf); }
void stringAlready(void) { char *buf; sprintf(buf, "%d", 4); }
void retained(void) { char buf[60]; count = sprintf(buf, "%d", 5); }
void unsafeFormat(void) { char buf[70]; sprintf(buf, "%n", count); }
`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"var buf string // was [20]byte in C", "var buf string // was [50]byte in C",
		"var buf [30]byte", "var buf [40]byte", "var buf [60]byte", "var buf [70]byte",
	} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if strings.Count(r.Draft, "// was") != 2 {
		t.Fatal(r.Draft)
	}
	// A trailing comment must not swallow a same-line following statement.
	if !strings.Contains(r.Draft, "in C\n") {
		t.Fatal(r.Draft)
	}
	for _, f := range r.Functions {
		if f.Name == "converted" && !strings.Contains(f.Draft, "var buf string // was [20]byte in C") {
			t.Fatal("declaration absent from function metadata", f.Draft)
		}
	}
}

func TestSprintfDestinationImportAlias(t *testing.T) {
	for _, imp := range []string{`import f "fmt"`, `import . "fmt"`} {
		got, _, err := RewriteStandardLibrary("package p\n" + imp + "\nfunc demo() { var buf [10]byte; sprintf(buf, \"hello\"); }")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, "var buf string // was [10]byte in C\n") {
			t.Fatal(got)
		}
	}
}

func TestSprintfDestinationRuntime(t *testing.T) {
	output, _ := runLibraryProgram(t, `package main
func main() { var buf [20]byte; sprintf(buf, "%s_%d", "hello", 7); if buf != "hello_7" { panic(buf) }; printf("%d", len(buf)); }
`)
	if output != "7" {
		t.Fatal(output)
	}
}
