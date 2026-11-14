package aig

import (
	"fmt"
	"strings"
	"testing"
)

func TestMultivalueFrameAllocatesSymbolsInSourceObjectOrder(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	latePI, lo := g.NewInput(), g.NewInput()
	g.AddOutput(gate)
	g.AddOutput(latePI)
	g.nRegs = 1
	p := Aig_MvManStart(g, 2)
	p.values[lo.NodeID()] = 1
	Aig_MvSimulateFrame(p, 1, 0)
	// Original C visits a,b,gate,latePI in object-ID order. The gate must
	// consume its symbolic literal before the later PI receives one.
	if p.values[gate.NodeID()] != 6 || p.values[latePI.NodeID()] != 8 || p.values[lo.NodeID()] != 8 {
		t.Fatalf("gate/late PI/LO = %d/%d/%d, want 6/8/8", p.values[gate.NodeID()], p.values[latePI.NodeID()], p.values[lo.NodeID()])
	}
}

func TestMultivalueTraceAndMappingAgainstCompiledSource(t *testing.T) {
	out := runAigSourceFixture(t, "multivalue_trace_compare.c", "aigMem.c")
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	late := g.NewInput()
	los := make([]Edge, 5)
	for i := range los {
		los[i] = g.NewInput()
	}
	g.AddOutput(gate)
	g.AddOutput(los[0])
	g.AddOutput(los[1].Not())
	g.AddOutput(late)
	g.AddOutput(late)
	g.AddOutput(gate)
	g.nRegs = 5
	p := Aig_MvManStart(g, 3)
	for _, e := range los {
		p.values[e.NodeID()] = 1
	}
	var got strings.Builder
	fmt.Fprintf(&got, "initial %d\n", Aig_MvSaveState(p))
	for f := 0; f < 8; f++ {
		first := int32(0)
		if f < 2 {
			first = 1
		}
		Aig_MvSimulateFrame(p, first, 0)
		state := Aig_MvSaveState(p)
		fmt.Fprintf(&got, "frame %d state %d objects %d:", f, state, p.nextLiteral/2)
		for _, v := range p.values {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
		if state >= 0 {
			break
		}
	}
	printInts := func(name string, values []int32) {
		fmt.Fprintf(&got, "%s:", name)
		for _, v := range values {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
	}
	var binary []int32
	constants := Aig_MvManFindConstBinaryFlops(p, &binary)
	printInts("undefined", Aig_MvManFindXFlops(p))
	printInts("constants", constants)
	printInts("binary", binary)
	printInts("oscillators", Aig_MvManFindOscilators(p, nil))
	printInts("skip", Aig_MvManCreateNextSkip(p))
	printMap := func(name string, values []Edge) {
		fmt.Fprintf(&got, "%s:", name)
		for _, e := range values {
			fmt.Fprintf(&got, " %d", e)
		}
		got.WriteByte('\n')
	}
	printMap("mapping", Aig_MvManDeriveMap(p, 0))
	printMap("driver", Aig_MvManSimulate(g, 2, 3, 0, 0))
	if got.String() != out {
		t.Fatalf("Go trace:\n%sC trace:\n%s", got.String(), out)
	}
}
