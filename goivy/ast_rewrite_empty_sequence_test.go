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

// Python lower_var_stmts clones a thunk with stmt.args + [Sequence(...)],
// so a thunk lowered twice has a sixth arg; canon shows it
// (test/disk_token_ring.ivy).
func TestThunkActionKeepsExtraArgs(t *testing.T) {
	cfg := NewAstConfig()
	th := &ThunkAction{Label: cfg.NewAtom("l"), Action: cfg.NewAtom("f"), Sort: cfg.NewAtom("t"), Body: cfg.NewSequence()}
	once := th.Clone(append(th.Args(), cfg.NewSequence(cfg.NewAtom("a")))).(*ThunkAction)
	twice := once.Clone(append(once.Args(), cfg.NewSequence())).(*ThunkAction)
	if n := len(twice.Args()); n != 6 {
		t.Fatalf("args %d, want 6: %v", n, twice.Canon())
	}
}
