package startingpoint

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestConditionalIslandPreserved(t *testing.T) {
	island := `#if FIRST
const int x = 1;
#elif SECOND
#define VALUE 3u
#else
const int x = 2;
#endif`
	source := "static inline int before() { return 1; }\n" + island + "\nstatic inline int after() { return 2; }\n"
	r, err := Analyze("p/file.c", source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, island) || !strings.Contains(r.Draft, "func before() int32") || !strings.Contains(r.Draft, "func after() int32") {
		t.Fatal(r.Draft)
	}
	if strings.Contains(r.Draft, "SP_OPAQUE_") {
		t.Fatal("projection leaked")
	}
	for _, f := range r.Functions {
		if source[f.Start:f.End] != f.Original || !strings.HasPrefix(f.Original, "static inline") {
			t.Fatal("bad original mapping", f)
		}
	}
	found := false
	for _, issue := range r.Issues {
		if issue.Category == "conditional-review" {
			found = true
			if source[issue.Start:issue.End] != island {
				t.Fatal("wrong island span")
			}
		}
	}
	if !found {
		t.Fatal("missing review issue")
	}
}

func TestConditionalSharedDelimiters(t *testing.T) {
	for _, body := range []string{
		`call(
#if WIDE
1);
#else
2);
#endif
`,
		`#if WIDE
if (wide &&
#else
if (
#endif
ready) { call(); }
`,
		`#if WIDE
if (wide) {
#else
if (small) {
#endif
call();
#if OTHER
} else if (a) {
#else
} else if (b) {
#endif
call(); }
`,
		`#if OUTER
#if INNER
call(1);
#else
call(2);
#endif
#else
call(3);
#endif
`,
	} {
		source := "static inline int f() {\n" + body + "\nreturn 7;\n}\nstatic inline int after() { return 9; }"
		r, err := Analyze("p/file.c", source)
		if err != nil {
			t.Fatalf("%v\n%s", err, source)
		}
		if !strings.Contains(r.Draft, "func after() int32") {
			t.Fatal("lost following function", r.Draft)
		}
		if strings.Contains(r.Draft, "SP_OPAQUE_") {
			t.Fatal("projection leaked", r.Draft)
		}
		_, regions, err := projectConditionals(source)
		if err != nil {
			t.Fatal(err)
		}
		for _, region := range regions {
			if !strings.Contains(r.Draft, region.original) {
				t.Fatal("island changed", r.Draft)
			}
		}
		for _, f := range r.Functions {
			if strings.Contains(f.Draft, "SP_OPAQUE_") {
				t.Fatal("metadata leaked", f)
			}
		}
	}
}

func TestConditionalRecoveryAndOriginalDiagnostic(t *testing.T) {
	source := "static inline int before() { return 1; }\n#if X\n{\n#else\n(\n#endif\nremaining;"
	r, err := Analyze("p/file.c", source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "func before() int32") || !strings.Contains(r.Draft, source[strings.Index(source, "#if"):]) {
		t.Fatal(r.Draft)
	}
	bad := "#if X\nint a;\n#else\nint b;\n#endif\n}\n"
	_, err = Analyze("p/file.c", bad)
	var diagnostic *SourceError
	if !errors.As(err, &diagnostic) || diagnostic.Location.Line != 6 || diagnostic.Location.Column != 1 || diagnostic.Location.ByteOffset != strings.Index(bad, "}") {
		t.Fatal(err)
	}
}

func TestFormerPreprocessorFallbacks(t *testing.T) {
	for _, path := range []string{"../src/bdd/cudd/cuddExport.c", "../src/bdd/cudd/cuddZddMisc.c", "../src/bdd/mtr/mtrBasic.c", "../src/misc/zlib/inflate.c", "../src/misc/zlib/trees.c"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		r, err := Analyze(path, string(source))
		if err != nil {
			t.Fatal(path, err)
		}
		if strings.Contains(r.Draft, "SP_OPAQUE_") {
			t.Fatal("projection leaked", path)
		}
		_, regions, err := projectConditionals(string(source))
		if err != nil {
			t.Fatal(err)
		}
		for _, region := range regions {
			if !strings.Contains(r.Draft, region.original) {
				t.Fatal("region not preserved", path)
			}
		}
		t.Logf("%s: %d isolated regions", path, len(regions))
	}
}
