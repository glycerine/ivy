package startingpoint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"
	"testing"
)

func TestABCCommentsPreserved(t *testing.T) {
	for _, name := range []string{"heap.h", "vec/vec_uint.h", "vec/vec_int.h", "vec/vec_sdbl.h"} {
		path := "../src/sat/satoko/utils/" + name
		s, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		r, err := Analyze(path, string(s))
		if err != nil {
			t.Fatal(err)
		}
		ts, err := Lex(string(s))
		if err != nil {
			t.Fatal(err)
		}
		for _, tok := range ts {
			if tok.Kind == "comment" && !strings.Contains(r.Draft, tok.Text) {
				t.Fatalf("lost comment in %s: %q", name, tok.Text)
			}
		}
	}
}

// A fresh related vector checks transfer of simple rules only. This is not a
// behavioral validation of the entire vector or an independent porting baseline.
func TestFreshCharVectorHelpers(t *testing.T) {
	path := "../src/sat/satoko/utils/vec/vec_char.h"
	s, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Analyze(path, string(s))
	if err != nil {
		t.Fatal(err)
	}
	output := "package p\ntype vec_char_t struct { size uint32 }\n"
	count := 0
	for _, f := range r.Functions {
		if f.Name == "vec_char_size" || f.Name == "vec_char_clear" {
			output += f.Draft + "\n"
			count++
		}
	}
	if count != 2 {
		t.Fatal("missing helpers")
	}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "generated.go", output, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := new(types.Config).Check("p", fs, []*ast.File{f}, nil); err != nil {
		t.Fatal(err)
	}
}

// Audit the entire assembled draft, not just functions whose signatures could
// be translated. Comments, original-macro comments and strings are excluded.
func TestABCAllIfBodiesBraced(t *testing.T) {
	count := 0
	for _, name := range []string{"heap.h", "vec/vec_uint.h", "vec/vec_int.h", "vec/vec_sdbl.h"} {
		path := "../src/sat/satoko/utils/" + name
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		r, err := Analyze(path, string(source))
		if err != nil {
			t.Fatal(err)
		}
		all, err := Lex(r.Draft)
		if err != nil {
			t.Fatal(err)
		}
		var tokens []Token
		for _, tok := range all {
			if tok.Kind != "space" && tok.Kind != "comment" {
				tokens = append(tokens, tok)
			}
		}
		pairs, err := MatchDelimiters(tokens)
		if err != nil {
			t.Fatal(err)
		}
		for i, tok := range tokens {
			if tok.Kind != "word" || tok.Text != "if" {
				continue
			}
			count++
			if ifBodyIndex(tokens, pairs, i) < 0 {
				t.Fatalf("%s: unbraced if at byte %d", name, tok.Start)
			}
		}
	}
	if count == 0 {
		t.Fatal("audit found no if statements")
	}
	t.Logf("audited %d if statements across four assembled drafts", count)
}

func TestABCFunctionSignaturesTranslated(t *testing.T) {
	count := 0
	for _, name := range []string{"heap.h", "vec/vec_uint.h", "vec/vec_int.h", "vec/vec_sdbl.h"} {
		path := "../src/sat/satoko/utils/" + name
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		r, err := Analyze(path, string(source))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range r.Functions {
			count++
			if !strings.HasPrefix(f.Draft, "func "+f.Name+"(") {
				t.Fatalf("untranslated signature: %s", f.Draft)
			}
			brace := strings.Index(f.Draft, "{")
			if brace < 0 || strings.Contains(f.Draft[:brace], "\n") {
				t.Fatalf("brace not on signature line: %s", f.Draft)
			}
		}
	}
	if count != 95 {
		t.Fatalf("expected 95 inventoried functions, got %d", count)
	}
}

func ifBodyIndex(tokens []Token, pairs map[int]int, start int) int {
	for j := start + 1; j < len(tokens); j++ {
		switch tokens[j].Text {
		case "{":
			return j
		case ";", "}", "if", "return", "break", "continue":
			return -1
		}
		if end, ok := pairs[j]; ok && end > j {
			j = end
		}
	}
	return -1
}
