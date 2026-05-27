package goivy

import "testing"

func TestV16CalledActionWithUnlabeledAssumeIsInlinable(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "assume", body: "assume p(a)"},
		{name: "assert", body: "assert p(a)"},
		{name: "ensures", body: "ensures p(a)"},
		{name: "while invariant", body: "while p(a) invariant p(a) {\n        assume p(a)\n    }"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := `#lang ivy1.6
type t
individual x:t
relation p(X:t)
action callee(a:t) = {
    ` + tc.body + `
}
action caller = {
    call callee(x)
}
`
			mod := New()
			if err := SourceString("v16_call_unlabeled_"+tc.name+".ivy", src, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
				t.Fatalf("SourceString: %v", err)
			}
			act, ok := mod.Actions.Get2("caller")
			if !ok {
				t.Fatalf("compiled module is missing caller action")
			}
			ctx := &UpdateContext{
				Domain:       mod,
				ActCfg:       mod.Cfg.ActCfg,
				Instantiator: mod.Instantiator,
			}
			if upd := GetUpdate(act.(ActionsAction), ctx); upd == nil {
				t.Fatal("GetUpdate returned nil")
			}
		})
	}
}

func TestV16InstantiatedAxiomKeepsSourceLine(t *testing.T) {
	src := `#lang ivy1.6
module reflexive(t) = {
    axiom X:t = X
}
type node
instantiate reflexive(node)
`
	mod := New()
	if err := SourceString("v16_instantiated_axiom_lines.ivy", src, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatalf("SourceString: %v", err)
	}
	if len(mod.LabeledAxioms) == 0 {
		t.Fatal("compiled module has no labeled axioms")
	}
	loc := mod.LabeledAxioms[0].GetLineno()
	if loc.Filename == "" || loc.Line != 3 {
		t.Fatalf("instantiated axiom location = %#v, want filename and line 3", loc)
	}
	if got := PrettyLF(mod.LabeledAxioms[0], 8); got == "        (internal) (no name)" {
		t.Fatalf("instantiated axiom pretty-printed as internal: %q", got)
	}
}

func TestV16IffAxiomKeepsSourceLine(t *testing.T) {
	src := `#lang ivy1.6
type t
individual x:t
relation p(X:t)
axiom p(x) <-> x = x
`
	mod := New()
	if err := SourceString("v16_iff_axiom_lines.ivy", src, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatalf("SourceString: %v", err)
	}
	if len(mod.LabeledAxioms) != 1 {
		t.Fatalf("compiled module axioms = %d, want 1", len(mod.LabeledAxioms))
	}
	loc := mod.LabeledAxioms[0].GetLineno()
	if loc.Filename == "" || loc.Line != 5 {
		t.Fatalf("iff axiom location = %#v, want filename and line 5", loc)
	}
}
