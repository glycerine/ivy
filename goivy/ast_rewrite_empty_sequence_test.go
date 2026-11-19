package goivy

import "testing"

// Python ast_rewrite clones anything with an args attribute, and an empty
// Sequence has args == (); clone gives it lineno_add_ref(lineno), so an
// instantiated bodiless action carries the instantiation reference
// (test/test_array.ivy: "...in action arr.set" then "instantiated here").
func TestAstRewriteEmptySequenceAddsReference(t *testing.T) {
	cfg := NewAstConfig()
	seq := cfg.NewSequence()
	seq.SetLineno(Location{Filename: "lib.ivy", Line: 29})
	cfg.SetReferenceLineno(Location{Filename: "main.ivy", Line: 8})
	res := AstRewrite(seq, NewAstRewriteSubstPrefix(map[string]string{}, nil))
	cfg.SetReferenceLineno(Location{})
	loc := res.GetLineno()
	if loc.Line != 8 || loc.Reference == nil || loc.Reference.Line != 29 {
		t.Fatalf("loc %+v ref %+v", loc, loc.Reference)
	}
}
