package goivy

import "testing"

func TestCompileGenericEmptySequencePreservesActionLikePython(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	c := NewFromModule(mod)

	compiled, err := c.Thing(cfg.NewSequence())
	if err != nil {
		t.Fatalf("compile empty sequence: %v", err)
	}
	seq, ok := compiled.(*LogicSequence)
	if !ok {
		t.Fatalf("compiled empty sequence is %T, want *LogicSequence", compiled)
	}
	if len(seq.Elems) != 0 {
		t.Fatalf("compiled empty sequence has %d elems, want 0", len(seq.Elems))
	}
}

func TestCompileGenericNestedEmptySequencesStayActionsLikePython(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	c := NewFromModule(mod)

	compiled, err := c.Thing(cfg.NewSequence(
		cfg.NewSequence(),
		cfg.NewSequence(),
		cfg.NewAssumeAction(cfg.NewAnd()),
	))
	if err != nil {
		t.Fatalf("compile sequence: %v", err)
	}
	seq, ok := compiled.(*LogicSequence)
	if !ok {
		t.Fatalf("compiled sequence is %T, want *LogicSequence", compiled)
	}
	if len(seq.Elems) != 3 {
		t.Fatalf("compiled sequence has %d elems, want 3", len(seq.Elems))
	}
	for i := 0; i < 2; i++ {
		if _, ok := seq.Elems[i].(*LogicSequence); !ok {
			t.Fatalf("compiled child %d is %T, want *LogicSequence", i, seq.Elems[i])
		}
	}
	if unwrapToAction(seq.Elems[2]) == nil {
		t.Fatalf("compiled child 2 is %T, want action", seq.Elems[2])
	}
}
