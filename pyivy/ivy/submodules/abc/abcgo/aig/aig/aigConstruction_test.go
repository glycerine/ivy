package aig

import "testing"

func TestRegisterCountDuringConstruction(t *testing.T) {
	p := NewGraph()
	p.NewInput()
	lo := p.NewInput()
	Aig_ManSetRegNumForConstruction(p, 1)
	if p.RegisterCount() != 1 {
		t.Fatal("register count unavailable during construction")
	}
	p.AddOutput(lo)
	Aig_ManSetRegNum(p, 1)
	if Aig_ManLo(p, 0) != lo {
		t.Fatal("wrong latch output")
	}
}

func TestRegisterCountConstructionRejectsInvalidInputCount(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected invalid register count to panic")
		}
	}()
	Aig_ManSetRegNumForConstruction(NewGraph(), 1)
}

func TestRegisterCountCompletionChecksOutputCount(t *testing.T) {
	p := NewGraph()
	p.NewInput()
	Aig_ManSetRegNumForConstruction(p, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("expected incomplete interface to panic")
		}
	}()
	Aig_ManSetRegNum(p, 1)
}
