package startingpoint

import (
	"strings"
	"testing"
)

func TestEmptyForBodies(t *testing.T) {
	for _, tc := range []struct {
		source, want string
		macro        bool
	}{
		{"for (i=0;i<n;i++)", "for (i=0;i<n;i++) {}", true},
		{"for (;;) ;", "for (;;) {}", false},
		{"for (;;);", "for (;;) {}", true},
		{"for (;;) /* keep */ ;", "for (;;) /* keep */ {}", false},
		{"for (;;) // trailing", "for (;;) {} // trailing", true},
		{"for (;;) { work(); }", "for (;;) { work(); }", true},
		{"for (;;) work();", "for (;;) work();", true},
		{`log("for (;;);"); // for (;;)`, `log("for (;;);"); // for (;;)`, true},
	} {
		got, err := emptyForBodies(tc.source, tc.macro)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("got %q; want %q", got, tc.want)
		}
		again, err := emptyForBodies(got, tc.macro)
		if err != nil || again != got {
			t.Fatal("not idempotent", again, err)
		}
	}
}

func TestNullLoopInFunctionDraft(t *testing.T) {
	r, err := Analyze("p/x.h", "static inline void spin() { for (;;) ; }")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "for {}") {
		t.Fatal(r.Draft)
	}
}
