package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSimulationAbstractionDriver(t *testing.T) {
	sims := vec.Vec_WrdAlloc(2)
	vec.Vec_WrdPush(sims, 0xaaaaaaaaaaaaaaaa)
	vec.Vec_WrdPush(sims, 0xcccccccccccccccc)
	support := Gia_SimAbsPerformOne(nil, []uint64{0x5555555555555555}, []uint64{0xaaaaaaaaaaaaaaaa}, sims, 1, 0)
	if vec.Vec_IntSize(support) != 1 || vec.Vec_IntEntry(support, 0) != 0 {
		t.Fatalf("wrong support %v", vec.Vec_IntArray(support))
	}
}
