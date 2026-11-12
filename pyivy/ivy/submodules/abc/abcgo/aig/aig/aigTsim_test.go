package aig

import (
	"bytes"
	"strings"
	"testing"
)

func TestTernaryOperators(t *testing.T) {
	for a := AIG_XVS0; a <= AIG_XVSX; a++ {
		if Aig_XsimInv(Aig_XsimInv(a)) != a {
			t.Fatal("inversion")
		}
		for b := AIG_XVS0; b <= AIG_XVSX; b++ {
			want := AIG_XVS1
			if a == AIG_XVS0 || b == AIG_XVS0 {
				want = AIG_XVS0
			} else if a == AIG_XVSX || b == AIG_XVSX {
				want = AIG_XVSX
			}
			if got := Aig_XsimAnd(a, b); got != want {
				t.Fatalf("%d & %d = %d", a, b, got)
			}
		}
	}
	for i := 0; i < 100; i++ {
		if x := Aig_XsimRand2(); x < 1 || x > 2 {
			t.Fatal(x)
		}
		if x := Aig_XsimRand3(); x < 1 || x > 3 {
			t.Fatal(x)
		}
	}
}
func TestTernaryStatePackingAndOwnership(t *testing.T) {
	p := NewGraph()
	for i := 0; i < 17; i++ {
		e := p.NewInput()
		p.AddOutput(e)
	}
	p.nRegs = 17
	manager := Aig_TsiStart(p)
	defer Aig_TsiStop(manager)
	state := Aig_TsiStateNew(manager)
	if len(state) != 2 {
		t.Fatal("packed word count")
	}
	for i := 0; i < 17; i++ {
		state[i/16] |= uint32(1+i%3) << uint(2*(i%16))
	}
	if Aig_TsiStateCount(manager, state) != 12 {
		t.Fatal("constant count")
	}
	Aig_TsiStateInsert(manager, state, 2)
	copyState := append([]uint32(nil), state...)
	if Aig_TsiStateLookup(manager, copyState, 2) != 1 {
		t.Fatal("state content lookup")
	}
	copyState[0] ^= 3
	if Aig_TsiStateLookup(manager, copyState, 2) != 0 {
		t.Fatal("state identity confused with contents")
	}
	var printed bytes.Buffer
	manager.WriteState(&printed, state)
	if !strings.HasPrefix(printed.String(), "01x01x") || !strings.Contains(printed.String(), "x=    5") {
		t.Fatal(printed.String())
	}
	if Aig_TsiStateHash(state, 2, 503) < 0 {
		t.Fatal("hash sign")
	}
}
func TestTernarySimulationConstantAndUnknownRegisters(t *testing.T) {
	p := NewGraph()
	input := p.NewInput()
	zero, toggle, unknown := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(p.OR(zero, toggle))
	p.AddOutput(zero)
	p.AddOutput(toggle.Not())
	p.AddOutput(input)
	p.nRegs = 3
	mapping := Aig_ManTernarySimulate(p, 0, 0)
	if len(mapping) != 4 || mapping[0] != input || mapping[1] != p.Const0() || mapping[2] != toggle || mapping[3] != unknown {
		t.Fatal(mapping)
	}
}
func TestTernarySimulationSimultaneousTransfer(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(b)
	p.AddOutput(a.Not())
	p.nRegs = 2
	if got := Aig_ManTernarySimulate(p, 0, 0); got != nil {
		t.Fatal("four-state cycle misclassified as constant", got)
	}
}
func TestTernarySimulationNoRegistersAndConstantOneTransition(t *testing.T) {
	p := NewGraph()
	if Aig_ManTernarySimulate(p, 0, 0) != nil {
		t.Fatal("combinational graph has latch constants")
	}
	lo := p.NewInput()
	p.AddOutput(p.Const1())
	p.nRegs = 1
	if result := Aig_ManTernarySimulate(p, 0, 0); result != nil {
		t.Fatal("initial zero followed by one is not constant", lo, result)
	}
}
func TestTernarySimulationRetirementAndLimit(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(a.Not())
	p.AddOutput(b)
	p.nRegs = 2
	mapping := ternarySimulate(p, 5, 0, 0, 0)
	if len(mapping) != 2 || mapping[0] != a || mapping[1] != p.Const0() {
		t.Fatal(mapping)
	}
	if result := ternarySimulate(p, 1, 0, 0, 0); result != nil {
		t.Fatal("bounded unfinished run returned constants")
	}
}
func TestTernaryFrameSupportsExplicitOutputsXorAndBuffer(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	Aig_ObjSetXsim(p, a, AIG_XVS0)
	Aig_ObjSetXsim(p, b, AIG_XVS1)
	x := p.NewXor(a, b)
	buffer := Edge(len(p.nodes)) << 1
	p.nodes = append(p.nodes, Node{Kind: Buffer, Fanins: [2]Edge{x.Not(), InvalidEdge}})
	p.AddOutput(buffer)
	ternarySimulateFrame(p)
	if ternaryEdgeValue(p, p.outputs[0]) != AIG_XVS0 {
		t.Fatal("explicit output/XOR/buffer simulation")
	}
}

func TestTernaryReadabilityAndBugReviewRegressions(t *testing.T) {
	t.Run("padding bits stay outside register count", func(t *testing.T) {
		p := NewGraph()
		lo := p.NewInput()
		p.AddOutput(lo)
		p.nRegs = 1
		mapping := Aig_ManTernarySimulate(p, 0, 0)
		if len(mapping) != 1 || mapping[0] != p.Const0() {
			t.Fatal(mapping)
		}
	})
	t.Run("state map owns its key", func(t *testing.T) {
		p := NewGraph()
		lo := p.NewInput()
		p.AddOutput(lo)
		p.nRegs = 1
		manager := Aig_TsiStart(p)
		state := Aig_TsiStateNew(manager)
		state[0] = 1
		Aig_TsiStateInsert(manager, state, 1)
		state[0] = 2
		if Aig_TsiStateLookup(manager, []uint32{1}, 1) != 1 || Aig_TsiStateLookup(manager, state, 1) != 0 {
			t.Fatal("mutable state changed stored key")
		}
	})
	t.Run("invalid value rejected", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("invalid ternary value accepted")
			}
		}()
		Aig_XsimAnd(0, AIG_XVS0)
	})
	t.Run("combinational cycle rejected", func(t *testing.T) {
		p := NewGraph()
		a, b := p.NewInput(), p.NewInput()
		gate := p.AND(a, b)
		p.node(gate).Fanins[0] = gate
		Aig_ObjSetXsim(p, a, AIG_XVS0)
		Aig_ObjSetXsim(p, b, AIG_XVS0)
		defer func() {
			if recover() == nil {
				t.Fatal("combinational cycle accepted")
			}
		}()
		ternarySimulateFrame(p)
	})
}

func TestTernaryConstantReductionPreservesSequentialBehavior(t *testing.T) {
	p := NewGraph()
	pi, zero, toggle := p.NewInput(), p.NewInput(), p.NewInput()
	p.Name = "sequential"
	p.FlopNumbers = []int32{40, 41}
	p.AddOutput(p.OR(zero, p.AND(pi, toggle)))
	p.AddOutput(zero)
	p.AddOutput(toggle.Not())
	p.nRegs = 2
	reduced := Aig_ManConstReduce(p, 0, 0, 0, 0, 0)
	if reduced.nRegs != 1 || len(reduced.inputs) != 2 || len(reduced.outputs) != 2 || reduced.Name != "sequential" || len(reduced.FlopNumbers) != 1 || reduced.FlopNumbers[0] != 41 {
		t.Fatal("incorrect reduced sequential interface", reduced.nRegs, reduced.inputs, reduced.outputs, reduced.FlopNumbers)
	}
	for mask := 0; mask < 4; mask++ {
		input, state := mask&1 != 0, mask&2 != 0
		got, err := reduced.Evaluate([]bool{input, state})
		if err != nil || got[0] != (input && state) || got[1] != !state {
			t.Fatal(got, err)
		}
	}
	if err := reduced.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestTernaryReductionSequentialReachability(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	p.AddOutput(a)
	p.AddOutput(b)
	p.AddOutput(b)
	p.AddOutput(c)
	p.nRegs = 3
	mapping := []Edge{a, b, p.Const0()}
	reduced := ternaryRemapAndCleanup(p, mapping)
	if reduced.nRegs != 2 {
		t.Fatal("dependency through next-state function was lost", reduced.nRegs)
	}
	if err := reduced.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestTernaryReductionInjectedMultivalueEngine(t *testing.T) {
	p := NewGraph()
	lo := p.NewInput()
	p.AddOutput(lo)
	p.AddOutput(lo)
	p.nRegs = 1
	calls := 0
	engine := func(g *Graph, symbolic, satur, verbose, veryVerbose int32) []Edge {
		calls++
		if symbolic != 5 || satur != 7 {
			t.Fatal("engine limits lost")
		}
		return []Edge{g.Const0()}
	}
	result := Aig_ManConstReduce(p, 1, 5, 7, 0, 0, engine)
	if result.nRegs != 0 || calls != 1 {
		t.Fatal("engine not used")
	}
	got, err := result.Evaluate(nil)
	if err != nil || len(got) != 1 || got[0] {
		t.Fatal(got, err)
	}
}

func TestTernaryReductionEngineValidation(t *testing.T) {
	t.Run("nil engine", func(t *testing.T) {
		p := NewGraph()
		lo := p.NewInput()
		p.AddOutput(lo)
		p.nRegs = 1
		defer func() {
			if recover() == nil {
				t.Fatal("nil symbolic engine accepted")
			}
		}()
		Aig_ManConstReduce(p, 1, 0, 0, 0, 0, nil)
	})
	t.Run("no progress", func(t *testing.T) {
		p := NewGraph()
		lo := p.NewInput()
		p.AddOutput(lo)
		p.AddOutput(lo)
		p.nRegs = 1
		engine := func(g *Graph, _, _, _, _ int32) []Edge { return append([]Edge(nil), g.inputs...) }
		defer func() {
			if recover() == nil {
				t.Fatal("identity mapping made infinite progress")
			}
		}()
		Aig_ManConstReduce(p, 1, 0, 0, 0, 0, engine)
	})
}

func TestTernaryReductionRejectsMalformedFlopMetadata(t *testing.T) {
	p := NewGraph()
	lo := p.NewInput()
	p.AddOutput(lo)
	p.nRegs = 1
	p.FlopNumbers = []int32{10, 20}
	defer func() {
		if recover() == nil {
			t.Fatal("malformed flop metadata silently discarded")
		}
	}()
	ternaryRemapAndCleanup(p, []Edge{p.Const0()})
}
