package startingpoint

import (
	"os"
	"strings"
	"testing"
)

func TestCharPointerTypesAndStrlen(t *testing.T) {
	source := `char *saved;
struct text_ { char *name; char **names; };
char *copy(const char *s) {
 char*local = s;
 int n = strlen(s);
 int m = strlen(call(s));
 // strlen(s) and char * remain in comments
 obj.strlen(s);
 return local;
}`
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"var saved string;", "name string;", "names *string;", "func copy(s string) string {", "var local string", "len(s)", "len(call(s))", "// strlen(s) and char * remain in comments", "obj.strlen(s);"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if strings.Contains(r.Draft, `import "fmt"`) {
		t.Fatal("strlen introduced fmt import")
	}
	if !strings.Contains(r.Functions[0].Draft, "len(s)") {
		t.Fatal("function metadata not updated")
	}
}

func TestStarterStringSignature(t *testing.T) {
	source, err := os.ReadFile("../src/starter.c")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("starter.c", string(source))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "func Abc_UtilStrsav(s string) string {") || !strings.Contains(r.Draft, "malloc(len(s)+1)") {
		t.Fatal("starter string conversion missing")
	}
}

func TestOpaqueStringCodePreserved(t *testing.T) {
	source := "#if X\nchar *copy(char *s) { return malloc(strlen(s)+1); }\n#endif\n"
	r, err := Analyze("x.c", source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, source) {
		t.Fatal(r.Draft)
	}
}
