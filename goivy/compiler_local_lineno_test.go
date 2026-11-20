package goivy

import "testing"

// Python compile_local sets res.lineno = self.lineno: a compiled local
// action reports the line of its `local` keyword, not its body's
// (test/learning_switch1.ivy: "line 72: implementation of new_packet").
func TestCompileLocalUsesOwnLineno(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()
	natSort := &UninterpretedSort{Name: "nat"}
	c.Sig.Sorts.Set("nat", natSort)
	c.Sig.AddSymbol("y", natSort)

	decl := cfg.NewAtom("loc:x")
	decl.ASort = cfg.NewAtom("nat")
	body := cfg.NewAssumeAction(cfg.NewAtom("true"))
	body.SetLineno(Location{Filename: "f.ivy", Line: 74})
	local := cfg.NewLocalAction("test", decl, body)
	local.SetLineno(Location{Filename: "f.ivy", Line: 72})

	res, err := c.CompileNode(local)
	if err != nil {
		t.Fatalf("CompileNode: %v", err)
	}
	la, ok := res.(*LogicLocalAction)
	if !ok {
		t.Fatalf("want *LogicLocalAction, got %T", res)
	}
	if got := la.GetLineno().Line; got != 72 {
		t.Errorf("compiled local lineno = %d, want 72 (the local's own line)", got)
	}
}
