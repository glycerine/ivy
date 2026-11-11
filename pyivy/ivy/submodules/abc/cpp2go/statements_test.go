package cpp2go

import (
	"strings"
	"testing"
)

func TestBraceIfStatements(t *testing.T) {
	cases := []struct{ in, want string }{
		{"    if (cap > 0 && cap < 16) cap = 16;", "    if (cap > 0 && cap < 16) { cap = 16; }"},
		{"if (a) if (b) x(); else y();", "if (a) { if (b) { x(); } else { y(); } }"},
		{"if (a) x(); else if (b) y(); else z();", "if (a) { x(); } else if (b) { y(); } else { z(); }"},
		{"if (a) /* keep */\n x(); // trailing", "if (a) { /* keep */\n x(); } // trailing"},
		{"if (a) for (i=0; i<n; i++) x();", "if (a) { for (i=0; i<n; i++) x(); }"},
		{"if (a) do x(); while (b);", "if (a) { do x(); while (b); }"},
		{"if (a) { x(); }", "if (a) { x(); }"},
		{`if (a) log("; if (fake) {");`, `if (a) { log("; if (fake) {"); }`},
	}
	for _, tc := range cases {
		got, err := braceIfStatements(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("input: %s\ngot: %s\nwant: %s", tc.in, got, tc.want)
		}
		again, err := braceIfStatements(got)
		if err != nil || again != got {
			t.Fatal("not idempotent", again, err)
		}
	}
}

func TestIfBracesInDraft(t *testing.T) {
	r := checkedDraft(t, `static inline int clamp(int cap) {
    if (cap > 0 && cap < 16) cap = 16;
    if (cap < 0) return cap > -10 ? -10 : cap;
    return cap;
}`, Options{Package: "p"})
	if !strings.Contains(r.Draft, "if (cap > 0 && cap < 16) { cap = 16; }") {
		t.Fatal(r.Draft)
	}
	if strings.Contains(r.Draft, "?") {
		t.Fatal("ternary in if body not lowered")
	}
}

func TestElseOnFollowingLine(t *testing.T) {
	checkedDraft(t, `static inline int f(int x) {
 if (x > 0)
   x = 1; // branch comment
 else
   x = 2;
 if (x == 2) { x = 3; }
 // existing block comment
 else { x = 4; }
 return x;
}`, Options{Package: "p"})
}
