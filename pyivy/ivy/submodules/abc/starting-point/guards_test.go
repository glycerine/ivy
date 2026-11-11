package startingpoint

import (
	"strings"
	"testing"
)

func TestHeaderGuardRemoved(t *testing.T) {
	source := `// license stays
#ifndef satoko__utils__vec__vec_uint_h
#define satoko__utils__vec__vec_uint_h

#if FEATURE
#define VALUE 1
#else
#define VALUE 2
#endif

#endif /* satoko__utils__vec__vec_uint_h */
// trailing comment stays
`
	r, err := Analyze("vec/file.h", source)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(r.Draft, "satoko__utils__vec__vec_uint_h") {
		t.Fatal(r.Draft)
	}
	for _, want := range []string{"// license stays", "#if FEATURE", "#else", "#endif", "// trailing comment stays"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(r.Draft)
		}
	}
	if strings.Count(r.Draft, "#endif") != 1 {
		t.Fatal(r.Draft)
	}
}

func TestConditionalNotMistakenForGuard(t *testing.T) {
	for _, source := range []string{
		"#ifndef A\n#define B\n#endif",
		"#ifndef A\n#define A 1\n#endif",
		"#ifndef A\n#define A\n#else\n#endif",
		"#ifndef A\n#define A\n#endif\nint outside;",
		"int before;\n#ifndef A\n#define A\n#endif",
		"#ifndef A\n#define A\n#if B\n#endif",
	} {
		r, err := Analyze("p/file.h", source)
		if err != nil {
			t.Fatal(err)
		}
		want := strings.NewReplacer("int outside;", "var outside int32;", "int before;", "var before int32;").Replace(source)
		if r.Draft != "package p\n\n"+want {
			t.Fatalf("removed unrelated conditional: %s", r.Draft)
		}
	}
}
