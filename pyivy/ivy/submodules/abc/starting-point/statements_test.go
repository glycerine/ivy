package startingpoint

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
		{"if (a) for (i=0; i<n; i++) x();", "if (a) { for (i=0; i<n; i++) { x(); } }"},
		{"if (a) do x(); while (b);", "if (a) { do x(); while (b); }"},
		{"if (a) { x(); }", "if (a) { x(); }"},
		{`if (a) log("; if (fake) {");`, `if (a) { log("; if (fake) {"); }`},
	}
	for _, tc := range cases {
		got, err := braceStatements(tc.in)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("input: %s\ngot: %s\nwant: %s", tc.in, got, tc.want)
		}
		again, err := braceStatements(got)
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
	if !strings.Contains(r.Draft, "if cap > 0 && cap < 16 { cap = 16; }") {
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

func TestForBodyWithNestedIf(t *testing.T) {
	source := `{
    for (i = 0; i < p.size; i++)
        if (p.data[i] == entry)
            return 1;
    return 0;
}`
	want := `{
    for (i = 0; i < p.size; i++) {
        if (p.data[i] == entry) {
            return 1; }
    }
    return 0;
}`
	got, err := braceStatements(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	again, err := braceStatements(got)
	if err != nil || again != got {
		t.Fatal("not idempotent", again, err)
	}
}

func TestForNestedElseAndComments(t *testing.T) {
	source := `for (i=0; i<n; i++) /* outer */ if (a) x(); else y(); after();`
	want := `for (i=0; i<n; i++) { /* outer */ if (a) { x(); } else { y(); } } after();`
	got, err := braceStatements(source)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatal(got)
	}
}

func TestWhileBecomesFor(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"while (ready()) { tick(); }", "for (ready()) { tick(); }"},
		{"while (ready()) tick();", "for (ready()) { tick(); }"},
		{"while (ready()) if (done()) break;", "for (ready()) { if (done()) { break; } }"},
		{"while /* keep */ (ready()) tick();", "for /* keep */ (ready()) { tick(); }"},
		{`log("while (x)"); // while (x)`, `log("while (x)"); // while (x)`},
		// The trailing while of do/while cannot be renamed independently.
		{"do { while (a) tick(); } while (b);", "do { for (a) { tick(); } } while (b);"},
	} {
		got, err := braceStatements(tc.source)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("got %q; want %q", got, tc.want)
		}
		again, err := braceStatements(got)
		if err != nil || again != got {
			t.Fatal("not idempotent", again, err)
		}
	}
	checkedDraft(t, `static inline int countdown(int n) {
 while (n > 0) n--;
 return n;
}`, Options{Package: "p"})
}

func TestQualifiedSignatureKeepsBodyBraces(t *testing.T) {
	source := `static inline int compare(const void *p1, const void *p2) {
    if (*p1 > *p2)
        return 1;
    if (*p1 < *p2)
        return -1;
    return 0;
}`
	r, err := Analyze("p/x.h", source)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "func compare(p1 any, p2 any) int32 {") {
		t.Fatal("qualified signature was not translated")
	}
	for _, want := range []string{"if *p1 > *p2 {", "if *p1 < *p2 {"} {
		if !strings.Contains(r.Draft, want) {
			t.Fatal(r.Draft)
		}
	}
}

func TestMacroIfBraces(t *testing.T) {
	r, err := Analyze("p/x.h", "#define CHECK(x) if (x) act(); else other();\nvoid check(int x) { CHECK(x) }\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "if x != 0 { act(); } else { other(); }") {
		t.Fatal(r.Draft)
	}
}
