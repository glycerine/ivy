package startingpoint

import (
	"strings"
	"testing"
)

func TestFilePointerTypes(t *testing.T) {
	r, err := Analyze("x.c", `FILE *global;
struct files_ { FILE *stream; FILE **streams; };
FILE *use(FILE *input) {
 FILE *local = input;
 return local;
}`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`import "os"`, "var global *os.File;", "stream *os.File;", "streams **os.File;", "func use(input *os.File) *os.File {", "var local *os.File"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatalf("missing %q:\n%s", want, r.Draft)
		}
	}
	if strings.Count(r.Draft, `import "os"`) != 1 {
		t.Fatal(r.Draft)
	}
}

func TestFilePointerImportProtection(t *testing.T) {
	r, err := Analyze("x.c", "// *os.File\n#if OPTION\nFILE *file;\n#endif\n")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Draft, `import "os"`) || !strings.Contains(r.Draft, "FILE *file;") {
		t.Fatal(r.Draft)
	}
	r, err = Analyze("x.c", "FILE *global;\nint main(int argc, char *argv[]) { return argc; }")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(r.Draft, `import "os"`) != 1 {
		t.Fatal(r.Draft)
	}
}
