package abc

// Go-owned replacements for the timing/printing drivers in giaDfs.c.

import (
	"slices"
	"testing"
)

func TestGiaDFSDiagnosticDriversSuperseded(t *testing.T) {
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	c, _ := graph.AppendCI()
	d, _ := graph.AppendCI()
	first, err := graph.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	second, err := graph.AppendAnd(first, c)
	if err != nil {
		t.Fatal(err)
	}
	third, err := graph.AppendAnd(a.Not(), d)
	if err != nil {
		t.Fatal(err)
	}
	co0, err := graph.AppendCO(second)
	if err != nil {
		t.Fatal(err)
	}
	co1, err := graph.AppendCO(third)
	if err != nil {
		t.Fatal(err)
	}

	// Gia_ManCollectTest performs one shared traversal across all outputs and
	// discards its per-output vectors after reporting elapsed time.
	allOutputGates, err := graph.CollectConeAnds([]GiaID{co0.ID(), co1.ID()}, nil, 0)
	if err != nil || !slices.Equal(allOutputGates, []GiaID{first.ID(), second.ID(), third.ID()}) {
		t.Fatal(allOutputGates, err)
	}

	// Gia_ManSuppSizeTest counts gates whose structural CI support is at most
	// sixteen. A smaller threshold makes the same computation observable here.
	gates, err := graph.AndObjects(0)
	if err != nil {
		t.Fatal(err)
	}
	small := 0
	for _, gate := range gates {
		support, supportErr := graph.SupportSize([]GiaID{gate}, 0)
		if supportErr != nil {
			t.Fatal(supportErr)
		}
		if support <= 2 {
			small++
		}
	}
	if small != 2 {
		t.Fatal(small)
	}

	// Gia_ManCollectSeqTest invokes the owned sequential closure separately
	// for each primary output and only prints progress/timing.
	want := [][]GiaID{
		{0, c.ID(), a.ID(), b.ID(), first.ID(), second.ID(), co0.ID()},
		{0, a.ID(), d.ID(), third.ID(), co1.ID()},
	}
	for output := 0; output < graph.PrimaryOutputCount(); output++ {
		got, collectErr := graph.CollectSequential([]int{output}, 0)
		if collectErr != nil || !slices.Equal(got, want[output]) {
			t.Fatal(output, got, want[output], collectErr)
		}
	}
}
