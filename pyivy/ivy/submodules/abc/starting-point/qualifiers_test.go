package startingpoint

import (
	"strings"
	"testing"
)

func TestStripConstQualifiers(t *testing.T) {
	source := "const unsigned *p; int const *q; const\nint x; const/*keep*/ int y;\n// const stays\nlog(\"const\"); constant = 1;"
	want := "unsigned *p; int *q; \nint x; /*keep*/ int y;\n// const stays\nlog(\"const\"); constant = 1;"
	got, err := stripConstQualifiers(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q; want %q", got, want)
	}
}

func TestConstRemovedFromWholeDraft(t *testing.T) {
	source := `#define PTR(x) ((const unsigned *) x)
static inline int compare(const void *p1, const void *p2) {
    const unsigned *pp1 = (const unsigned *) p1;
    const unsigned *pp2 = (const unsigned *) p2;
    if (*pp1 < *pp2) return -1;
    return 0;
}`
	r, err := Analyze("p/file.h", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{r.Draft, r.Functions[0].Draft} {
		ts, err := Lex(code)
		if err != nil {
			t.Fatal(err)
		}
		for _, tok := range ts {
			if tok.Kind == "word" && tok.Text == "const" {
				t.Fatal("const qualifier remains", code)
			}
		}
	}
	if !strings.Contains(r.Draft, "// #define PTR(x) ((const unsigned *) x)") {
		t.Fatal("original macro comment changed")
	}
}
