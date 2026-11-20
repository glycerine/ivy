package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// The two zero-initialized shift registers receive the same unconstrained
// input. PDR must prove their last stages can never disagree.
func BenchmarkPairedShiftRegisters(b *testing.B) {
	const stages = 12
	b.ReportAllocs()
	var frame int32
	for b.Loop() {
		g := aig.NewGraph()
		input := g.NewInput()
		var left, right [stages]aig.Edge
		for i := range stages {
			left[i], right[i] = g.NewInput(), g.NewInput()
		}
		g.AddOutput(g.XOR(left[stages-1], right[stages-1]))
		g.AddOutput(input)
		g.AddOutput(input)
		for i := 1; i < stages; i++ {
			g.AddOutput(left[i-1])
			g.AddOutput(right[i-1])
		}
		aig.Aig_ManSetRegNum(g, 2*stages)
		var p Pdr_Par_t
		Pdr_ManSetDefaultParams(&p)
		p.FSilent = 1
		if Pdr_ManSolve(g, &p) != 1 {
			b.Fatal("paired shift registers can disagree")
		}
		frame = p.IFrame
		aig.Aig_ManStop(g)
	}
	b.ReportMetric(float64(frame), "last-frame/op")
}
