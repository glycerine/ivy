package startingpoint

import (
	"go/format"
	"strings"
	"testing"
)

func TestForHeaderParentheses(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"for (i = 0; i < size; i++) {}", "for i = 0; i < size; i++ {}"},
		{"for(i=0; i<(n+1); i++){}", "for i=0; i<(n+1); i++{}"},
		{"for (ready()) {}", "for ready() {}"},
		{"for ((a || b) && c) {}", "for (a || b) && c {}"},
		{"for (a || b) && c {}", "for (a || b) && c {}"},
		{"for /*keep*/ (i=0; i<n; i++) {}", "for /*keep*/ i=0; i<n; i++ {}"},
		{`log("for (x)"); // for (x)`, `log("for (x)"); // for (x)`},
	} {
		got, err := unwrapForHeaders(tc.source)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("got %q; want %q", got, tc.want)
		}
	}
}

func TestGeneratedFindCanBeFormatted(t *testing.T) {
	source := `static inline unsigned find(vec_t *p, unsigned entry) {
    unsigned i;
    for (i = 0; i < p->size; i++)
        if (p->data[i] == entry) return 1;
    return 0;
}`
	r, err := Analyze("p/file.h", source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "for i = 0; i < p.size; i++ {") {
		t.Fatal(r.Draft)
	}
	if _, err := format.Source([]byte(r.Draft)); err != nil {
		t.Fatalf("gofmt: %v\n%s", err, r.Draft)
	}
}
