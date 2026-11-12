package gia

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math"
	"testing"
)

func TestSwitchEntryPoints(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	p.pSwitching = make([]byte, p.nObjs)
	p.pSwitching[a>>1] = 255
	p.pSwitching[b>>1] = 128
	if got := Gia_ManEvaluateSwitching(p); math.Abs(float64(got)-(1+128.0/255)) > 1e-6 {
		t.Fatalf("weighted switching %v", got)
	}
	probs := Gia_ManPrintOutputProb(p)
	if Vec_FltSize(probs) != int(Gia_ManObjNum(p)) || math.Abs(float64(Vec_FltEntry(probs, int(x>>1)))-.25) > .15 {
		t.Fatal("invalid probability vector")
	}
	g := aig.NewGraph()
	ga, gb := g.NewInput(), g.NewInput()
	gate := g.AND(ga, gb)
	g.AddOutput(gate)
	v := Saig_ManComputeSwitchProbs(g, 48, 16, 1)
	got := math.Float32frombits(uint32(Vec_IntEntry(v, gate.NodeID())))
	if math.Abs(float64(got)-.25) > .05 {
		t.Fatalf("AIG switching probability %v", got)
	}
}
