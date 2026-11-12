package aig

import (
	"reflect"
	"testing"
)

func TestMultivalueAndReferenceRules(t *testing.T) {
	g := NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.nRegs = 1
	p := Aig_MvManStart(g, 2)
	a, b := p.newLiteral(), p.newLiteral()
	cases := []struct {
		a, b  int32
		first bool
		want  int32
	}{{a, a, true, a}, {a, mvNot(a), true, 1}, {0, a, true, a}, {1, a, true, 1}, {SAIG_UNDEF_VALUE, 1, true, 1}, {SAIG_UNDEF_VALUE, 0, true, SAIG_UNDEF_VALUE}, {SAIG_UNDEF_VALUE, SAIG_UNDEF_VALUE, true, SAIG_UNDEF_VALUE}, {a, b, false, SAIG_UNDEF_VALUE}}
	for _, c := range cases {
		if got := p.and(c.a, c.b, c.first); got != c.want {
			t.Fatalf("and(%d,%d,%v)=%d want%d", c.a, c.b, c.first, got, c.want)
		}
	}
	first := p.and(a, b, true)
	if first != p.and(b, a, true) || first < 2 {
		t.Fatal("symbolic structural hash")
	}
	if mvNot(SAIG_UNDEF_VALUE) != SAIG_UNDEF_VALUE {
		t.Fatal("unknown complemented into known symbol")
	}
}
func TestMultivalueFindsConstantsAndSymbolicEquivalentRegisters(t *testing.T) {
	g := NewGraph()
	input, zero, a, b := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.OR(a, b))
	g.AddOutput(zero)
	g.AddOutput(input)
	g.AddOutput(input)
	g.nRegs = 3
	mapping := Aig_MvManSimulate(g, 1, 2, 0, 0)
	if len(mapping) != 4 || mapping[0] != input || mapping[1] != g.Const0() {
		t.Fatal(mapping)
	} // Both driven PIs later become undef, so they are not certified equivalent.
	if mapping[2] != a || mapping[3] != b {
		t.Fatal("undefined histories merged", mapping)
	}
	h := NewGraph()
	x, y := h.NewInput(), h.NewInput()
	h.AddOutput(h.OR(x, y))
	h.AddOutput(x.Not())
	h.AddOutput(x.Not())
	h.nRegs = 2
	mapping = Aig_MvManSimulate(h, 1, 3, 0, 0)
	if len(mapping) != 2 || mapping[0] != x || mapping[1] != x {
		t.Fatal("equivalent binary latches not merged", mapping)
	}
}
func TestMultivalueOscillatorAndSkipReference(t *testing.T) {
	g := NewGraph()
	for i := 0; i < 3; i++ {
		lo := g.NewInput()
		g.AddOutput(lo)
	}
	g.nRegs = 3
	p := Aig_MvManStart(g, 3)
	p.states = [][]int32{{1, 1, 1}, {0, 1, 2}, {1, 1, SAIG_UNDEF_VALUE}, {0, 1, SAIG_UNDEF_VALUE}, {1, 1, SAIG_UNDEF_VALUE}}
	if period := Aig_MvManCheckOscilator(p, 0); period != 2 {
		t.Fatal(period)
	}
	var binary []int32
	if constants := Aig_MvManFindConstBinaryFlops(p, &binary); !reflect.DeepEqual(constants, []int32{1}) || !reflect.DeepEqual(binary, []int32{0}) {
		t.Fatal(constants, binary)
	}
	if skip := Aig_MvManCreateNextSkip(p); !reflect.DeepEqual(skip, []int32{0, 0, 1}) {
		t.Fatal(skip)
	}
}
func TestMultivalueSymbolicFrameTraceAndRetirement(t *testing.T) {
	g := NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.nRegs = 1
	p := Aig_MvManStart(g, 2)
	p.values[lo.NodeID()] = 1
	Aig_MvSaveState(p)
	Aig_MvSimulateFrame(p, 1, 0)
	first := p.currentState()[0]
	if first < 2 || first == SAIG_UNDEF_VALUE {
		t.Fatal("first PI not symbolic", first)
	}
	Aig_MvSaveState(p)
	Aig_MvSimulateFrame(p, 1, 0)
	if next := p.currentState()[0]; next == first {
		t.Fatal("PI reused across symbolic frames")
	}
	Aig_MvSimulateFrame(p, 0, 0)
	if p.currentState()[0] != SAIG_UNDEF_VALUE {
		t.Fatal("non-symbolic PI remained known")
	}
	Aig_MvSaveState(p)
	if !reflect.DeepEqual(Aig_MvManFindXFlops(p), []int32{1}) {
		t.Fatal("unknown history missing")
	}
	if got := Aig_MvManSimulate(g, 1, 2, 0, 0); got != nil {
		t.Fatal("unknown register mapped", got)
	}
}
func TestMultivalueStateIndicesAndSnapshotOwnership(t *testing.T) {
	g := NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.nRegs = 1
	p := Aig_MvManStart(g, 1)
	p.values[lo.NodeID()] = 1
	if Aig_MvSaveState(p) != -1 || Aig_MvSaveState(p) != 1 {
		t.Fatal("initial state index mismatch")
	}
	p.values[lo.NodeID()] = 0
	if Aig_MvSaveState(p) != -1 || Aig_MvSaveState(p) != 2 || p.states[0][0] != 1 {
		t.Fatal("state contents aliased current values")
	}
}
func TestMultivalueConstantReductionDefaultEngine(t *testing.T) {
	g := NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.OR(x, y))
	g.AddOutput(x.Not())
	g.AddOutput(x.Not())
	g.nRegs = 2
	mapping := Aig_MvManSimulate(g, 1, 3, 0, 0)
	if mapping == nil {
		t.Fatal("equivalence absent")
	}
	reduced := ternaryRemapAndCleanup(g, mapping)
	if reduced.nRegs != 1 {
		t.Fatal("symbolic equivalence failed to reduce", reduced.nRegs)
	}
}

func TestMultivalueSaturationLimitRejectsOverflow(t *testing.T) {
	g := NewGraph()
	lo := g.NewInput()
	g.AddOutput(lo)
	g.nRegs = 1
	defer func() {
		if recover() == nil {
			t.Fatal("overflowing saturation limit accepted")
		}
	}()
	Aig_MvManSimulate(g, 1, 1<<30, 0, 0)
}
