package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationAbstractionCoverAndRefine(t *testing.T) {
	sims := vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(sims, 0xc)
	vec.Vec_WrdPush(sims, 0xa)
	selected := vec.Vec_IntAlloc(2)
	p := Gia_SimAbsAlloc(nil, []uint64{3}, []uint64{12}, sims, 1, selected, 0)
	defer Gia_SimAbsFree(p)
	vec.Vec_IntPushTwo(p.vPatPairs, 0, 2)
	vec.Vec_IntPushTwo(p.vPatPairs, 1, 3)
	Gia_SimAbsSolve(p)
	if vec.Vec_IntSize(selected) != 1 || vec.Vec_IntEntry(selected, 0) != 0 || Gia_SimAbsRefine(p) != 0 {
		t.Fatalf("wrong support %v", vec.Vec_IntArray(selected))
	}
	vec.Vec_IntClear(selected)
	vec.Vec_IntPush(selected, 1)
	if Gia_SimAbsRefine(p) != 1 || vec.Vec_IntSize(p.vPatPairs) <= 4 {
		t.Fatal("failed to refine ambiguous support")
	}
}
