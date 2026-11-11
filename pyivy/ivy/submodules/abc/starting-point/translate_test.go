package startingpoint

import (
	"strings"
	"testing"
)

func TestLosslessAndProtectedText(t *testing.T) {
	s := "// comment -> NULL { \\\ncontinued\n/* λ */\n#define X(a) \\\n ((a)+1)\nstatic inline int f(int x) { /* keep -> NULL */ const char *s = R\"tag({ \" -> NULL })tag\"; return x; }\n"
	ts, err := Lex(s)
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, tok := range ts {
		if s[tok.Start:tok.End] != tok.Text {
			t.Fatal("bad offsets")
		}
		b.WriteString(tok.Text)
	}
	if b.String() != s {
		t.Fatal("round trip lost source")
	}
	if _, err := MatchDelimiters(ts); err != nil {
		t.Fatal(err)
	}
	r, err := Analyze("fixture.h", s)
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range ts {
		if tok.Kind == "comment" || tok.Kind == "string" {
			if !strings.Contains(r.Draft, tok.Text) {
				t.Fatalf("lost protected text %q", tok.Text)
			}
		}
	}
	if len(r.Functions) != 1 {
		t.Fatalf("functions: %d", len(r.Functions))
	}
}
func TestSignatureCommentsAndMalformedInput(t *testing.T) {
	s := "static inline int f(/* argument */ int x) { return x; }"
	r, err := Analyze("x.h", s)
	if err != nil {
		t.Fatal(err)
	}
	if r.Draft != "package starting_point\n\n"+s {
		t.Fatal("signature comment changed")
	}
	for _, s := range []string{"/*", "\"unfinished", `R"tag(unfinished`, "{]"} {
		ts, err := Lex(s)
		if err == nil {
			_, err = MatchDelimiters(ts)
		}
		if err == nil {
			t.Fatalf("accepted malformed %q", s)
		}
	}
}
func TestDraftAndEdits(t *testing.T) {
	s := "static inline unsigned get(vec_t *p) { unsigned x = p->size; /* preserve */ return x; }"
	a, err := Analyze("x.h", s)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Analyze("x.h", s)
	if err != nil {
		t.Fatal(err)
	}
	if a.Draft != b.Draft || a.SHA256 != b.SHA256 {
		t.Fatal("not deterministic")
	}
	if !strings.Contains(a.Draft, "func get(p *vec_t) uint32") || !strings.Contains(a.Draft, "p.size") {
		t.Fatal(a.Draft)
	}
	if _, err := ApplyEdits("abcdef", []Edit{{1, 4, "x"}, {3, 5, "y"}}); err == nil {
		t.Fatal("accepted overlapping edits")
	}
}

func TestPointerDeclarationAndCRLFComment(t *testing.T) {
	s := "static inline vec_t *f() { vec_t *p = alloc(); // keep \\\r\n p->hidden NULL\r\n return p; }"
	r, err := Analyze("x.h", s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Draft, "var p *vec_t") {
		t.Fatal(r.Draft)
	}
	if !strings.Contains(r.Draft, "p->hidden NULL") {
		t.Fatal("rewrote continued comment")
	}
}

func TestPackageFromSourceDirectory(t *testing.T) {
	for _, path := range []string{"src/sat/satoko/utils/heap.h", "/tmp/project/utils/file.c"} {
		r, err := Analyze(path, "")
		if err != nil {
			t.Fatal(err)
		}
		if r.Draft != "package utils\n\n" {
			t.Fatalf("%s: %q", path, r.Draft)
		}
	}
	r, err := AnalyzeWithOptions("src/vec/file.h", "", Options{Package: "satoko"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Draft != "package satoko\n\n" {
		t.Fatal(r.Draft)
	}
	for path, want := range map[string]string{"src/bad-name/file.c": "bad_name", "src/for/file.c": "for_"} {
		r, err := Analyze(path, "")
		if err != nil {
			t.Fatal(err)
		}
		if r.Draft != "package "+want+"\n\n" {
			t.Fatal(r.Draft)
		}
	}
}
