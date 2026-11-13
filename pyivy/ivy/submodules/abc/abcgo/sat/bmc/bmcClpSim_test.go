package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestComputeSimDiff(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b))
	p := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(p, 1)
	vec.Vec_IntPush(p, 0)
	diff := vec.Vec_IntAlloc(2)
	if Bmc_ComputeSimDiff(g, p, diff) != 0 || vec.Vec_IntEntry(diff, 0) != 0 || vec.Vec_IntEntry(diff, 1) != 1 {
		t.Fatalf("diff %v", vec.Vec_IntArray(diff))
	}
}
