package aig

import (
	"math/rand"
	"slices"
	"testing"
)

func retFExample(stages int, firstCompl, secondCompl, xor bool) *Graph {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	registers := make([]Edge, 2*stages)
	for i := range registers {
		registers[i] = p.NewInput()
	}
	left, right := registers[0], registers[1]
	if firstCompl {
		left = left.Not()
	}
	if secondCompl {
		right = right.Not()
	}
	var root Edge
	if xor {
		root = p.NewXor(left, right)
	} else {
		root = p.AND(left, right)
	}
	p.AddOutput(root)
	for i := range registers {
		driver := a
		if i+2 < len(registers) {
			driver = registers[i+2]
		} else if i%2 != 0 {
			driver = b
		}
		p.AddOutput(driver)
	}
	p.nRegs = int32(len(registers))
	return p
}
func retFTrace(t *testing.T, p *Graph, sequence [][]bool) [][]bool {
	t.Helper()
	registers, pis, pos := sclLayout(p)
	state := make([]bool, registers)
	trace := make([][]bool, 0, len(sequence))
	for _, input := range sequence {
		if len(input) != pis {
			t.Fatal("wrong trace input count")
		}
		values := append(append([]bool(nil), input...), state...)
		out, err := p.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		trace = append(trace, append([]bool(nil), out[:pos]...))
		copy(state, out[pos:])
	}
	return trace
}
func checkRetFTraces(t *testing.T, before, after *Graph) {
	t.Helper()
	rng := rand.New(rand.NewSource(303))
	pis := len(before.inputs) - int(before.nRegs)
	for trial := 0; trial < 60; trial++ {
		sequence := make([][]bool, 12)
		for i := range sequence {
			sequence[i] = make([]bool, pis)
			for j := range sequence[i] {
				sequence[i][j] = rng.Intn(2) != 0
			}
		}
		a, b := retFTrace(t, before, sequence), retFTrace(t, after, sequence)
		for i := range a {
			if !slices.Equal(a[i], b[i]) {
				t.Fatal("retiming changed zero-initialized trace", trial, i, a[i], b[i])
			}
		}
	}
}
func TestRetimeFrontierPolarityAndXor(t *testing.T) {
	for _, xor := range []bool{false, true} {
		for _, a := range []bool{false, true} {
			for _, b := range []bool{false, true} {
				p := retFExample(1, a, b, xor)
				before := Aig_ManDupOrdered(p)
				q := Aig_ManRetimeFrontier(p, 10)
				if q.nRegs != 1 {
					t.Fatal("two registers should combine", q.nRegs, xor, a, b)
				}
				if !q.RetimingChanged {
					t.Fatal("missing retiming marker")
				}
				if err := q.Check(); err != nil {
					t.Fatal(err)
				}
				checkRetFTraces(t, before, q)
				for id, n := range p.nodes {
					if !p.deleted[id] && (n.MarkA || n.MarkB) {
						t.Fatal("scratch marks leaked", id)
					}
				}
			}
		}
	}
}
func TestRetimeFrontierStepLimitAndForwardIDs(t *testing.T) {
	for steps := int32(0); steps <= 2; steps++ {
		p := retFExample(2, false, false, false)
		before := Aig_ManDupOrdered(p)
		q := Aig_ManRetimeFrontier(p, steps)
		if q.nRegs != 4-max(steps, 1) {
			t.Fatal("retiming step limit", steps, q.nRegs)
		}
		if err := q.Check(); err != nil {
			t.Fatal(err)
		}
		checkRetFTraces(t, before, q)
	}
}
func TestRetimeFrontierReachability(t *testing.T) {
	p := NewGraph()
	a := p.NewInput()
	lo0, lo1 := p.NewInput(), p.NewInput()
	p.AddOutput(p.AND(lo0, lo1))
	p.AddOutput(p.Const1())
	p.AddOutput(p.Const0())
	p.nRegs = 2
	Aig_ManRetimeMark(p)
	if !p.node(a).MarkB || p.node(lo0).MarkB || p.node(lo1).MarkB {
		t.Fatal("autonomous registers were marked")
	}
	for i := range p.nodes {
		p.nodes[i].MarkB = false
	}
	q := Aig_ManRetimeFrontier(p, 10)
	if q.nRegs != 2 || q.RetimingChanged {
		t.Fatal("autonomous cone retimed")
	}
}
func TestRetimeFrontierMetadataAndDomains(t *testing.T) {
	for _, move := range []bool{false, true} {
		p := retFExample(1, false, false, false)
		p.Name, p.Spec = "retime", "source"
		p.FlopNumbers = []int32{40, 41}
		p.FlopReprs = []int32{40, 41}
		if move {
			p.ClockDomains = [][]int32{{0, 1}}
		} else {
			p.ClockDomains = [][]int32{{0}, {1}}
		}
		q := Aig_ManRetimeFrontier(p, 10)
		if q.Name != "retime" || q.Spec != "source" {
			t.Fatal("lost metadata")
		}
		if move {
			if q.FlopNumbers != nil || q.FlopReprs != nil || !q.RetimingChanged || len(q.ClockDomains) != 1 || !slices.Equal(q.ClockDomains[0], []int32{0}) {
				t.Fatal("invalid retimed metadata", q.ClockDomains, q.FlopNumbers)
			}
		} else if q.RetimingChanged || !slices.Equal(q.FlopNumbers, []int32{40, 41}) || !slices.Equal(q.FlopReprs, []int32{40, 41}) || q.nRegs != 2 {
			t.Fatal("no-op metadata changed")
		}
	}
}

func TestRetimeFrontierInvalidDomainLeavesScratchClean(t *testing.T) {
	p := retFExample(1, false, false, false)
	p.ClockDomains = [][]int32{{0, 0}}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("duplicate clock domain accepted")
			}
		}()
		Aig_ManRetimeFrontier(p, 1)
	}()
	for id, n := range p.nodes {
		if n.MarkA || n.MarkB {
			t.Fatal("validation failure leaked scratch marks", id)
		}
	}
}
func TestRetimeFrontierBuffersAndFeedback(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	lo0, lo1 := p.NewInput(), p.NewInput()
	buf0 := Aig_ObjCreate(p, &Node{Kind: Buffer, Fanins: [2]Edge{lo0.Not(), InvalidEdge}, Phase: true})
	buf1 := Aig_ObjCreate(p, &Node{Kind: Buffer, Fanins: [2]Edge{lo1.Not(), InvalidEdge}, Phase: true})
	gate := p.AND(buf0, buf1)
	p.AddOutput(gate)
	p.AddOutput(a)
	p.AddOutput(b)
	p.nRegs = 2
	before := Aig_ManDupOrdered(p)
	q := Aig_ManRetimeFrontier(p, 1)
	if q.nRegs != 1 {
		t.Fatal("buffered registers not combined")
	}
	checkRetFTraces(t, before, q)
	feedback := NewGraph()
	input := feedback.NewInput()
	x, y := feedback.NewInput(), feedback.NewInput()
	root := feedback.AND(x, y)
	feedback.AddOutput(root)
	feedback.AddOutput(root)
	feedback.AddOutput(input)
	feedback.nRegs = 2
	original := Aig_ManDupOrdered(feedback)
	retimed := Aig_ManRetimeFrontier(feedback, 3)
	if err := retimed.Check(); err != nil {
		t.Fatal(err)
	}
	checkRetFTraces(t, original, retimed)
}
func TestRetimeFrontierNoEligibleMovePreservesMetadata(t *testing.T) {
	p := retFExample(1, false, false, false)
	p.Name, p.Spec = "no-op", "origin"
	p.FlopNumbers = []int32{7, 9}
	p.FlopReprs = []int32{7, 9}
	p.ClockDomains = [][]int32{{0}, {1}}
	q := Aig_ManRetimeFrontier(p, 0)
	if q.RetimingChanged || p.RetimingChanged || !slices.Equal(q.FlopNumbers, p.FlopNumbers) || !slices.Equal(q.FlopReprs, p.FlopReprs) || q.Name != p.Name || q.Spec != p.Spec || !slices.Equal(q.ClockDomains[0], p.ClockDomains[0]) {
		t.Fatal("no-move metadata changed")
	}
	q.FlopNumbers[0] = 42
	if p.FlopNumbers[0] != 7 {
		t.Fatal("no-op shares mutable metadata")
	}
}

func TestRetimeFrontierNonpositiveLimitMatchesSource(t *testing.T) {
	for _, steps := range []int32{-1, 0} {
		p := retFExample(1, true, true, false)
		before := Aig_ManDupOrdered(p)
		q := Aig_ManRetimeFrontier(p, steps)
		if q.nRegs != 1 || !q.RetimingChanged {
			t.Fatalf("limit %d: source permits one move, got %d registers, changed=%t", steps, q.nRegs, q.RetimingChanged)
		}
		checkRetFTraces(t, before, q)
	}
}
