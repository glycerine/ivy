package startingpoint

import (
	"strings"
	"testing"
)

func TestConstParameterVoidResultSignature(t *testing.T) {
	r := checkedDraft(t, `static inline void vec_uint_duplicate(vec_uint_t *dest, const vec_uint_t *src)
{
    dest.size = src.size;
}
type vec_uint_t struct { size uint32 }
`, Options{Package: "p"})
	if !strings.Contains(r.Draft, "func vec_uint_duplicate(dest *vec_uint_t, src *vec_uint_t) {\n") {
		t.Fatal(r.Draft)
	}
	for _, word := range []string{"static inline", "void", "const"} {
		if strings.Contains(r.Draft, word) {
			t.Fatal(r.Draft)
		}
	}
}

func TestVoidPointerSignature(t *testing.T) {
	r := checkedDraft(t, "static inline void take(const void *p)\n{\n}\n", Options{Package: "p"})
	if !strings.Contains(r.Draft, "func take(p any) {\n") {
		t.Fatal(r.Draft)
	}
	found := false
	for _, issue := range r.Issues {
		if issue.Category == "void-pointer-review" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing void pointer representation review")
	}
}
