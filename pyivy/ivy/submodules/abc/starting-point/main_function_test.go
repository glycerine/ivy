package startingpoint

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestMainArgumentsAndReturns(t *testing.T) {
	r, err := Analyze("x.c", `int other(int argc) { return argc; }
int main(int count, char *args[]) {
 // args count return 0;
 if (count != 2) return 1;
 printf("%s", args[1]);
 return 0;
}`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"func main() {", `import "os"`, "if len(os.Args) != 2", "os.Args[1]", "os.Exit(int( 1))", "// args count return 0;", "return argc;"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	fs := token.NewFileSet()
	_, parseErr := parser.ParseFile(fs, "main.go", r.Draft, parser.ParseComments)
	if parseErr != nil {
		t.Fatal(parseErr)
	}

	tokens, _ := Lex(r.Functions[1].Draft)
	if strings.Contains(Mask(tokens), "return 0;") {
		t.Fatal(r.Draft)
	}
}

func TestMainWithoutArgumentsNeedsNoOSImport(t *testing.T) {
	checkedDraft(t, "int main(void) { return 0; }", Options{Package: "main"})
	r, err := Analyze("x.c", "int main(void) { return 0; }")
	if err != nil || strings.Contains(r.Draft, `import "os"`) {
		t.Fatal(r.Draft, err)
	}
}

func TestOSImportDeduplication(t *testing.T) {
	for _, source := range []string{"package p\nimport \"os\"\n", "package p\nimport (\n\"fmt\"\n\"os\"\n)\n"} {
		got, err := ensureOSImport(source)
		if err != nil || got != source {
			t.Fatal(got, err)
		}
	}
}
