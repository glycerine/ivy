package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestResimulationDetectsConstantMiterWithoutInputs(t *testing.T) {
	g := gia.Gia_ManStart(4)
	gia.Gia_ManAppendCo(g, 1)
	g.StartRepresentatives()
	info := vec.Vec_PtrAllocSimInfo(0, 1)
	if got := Cec_ManSeqResimulateInfo(g, info, nil, 1); got != 1 {
		t.Fatalf("constant miter status %d", got)
	}
}
