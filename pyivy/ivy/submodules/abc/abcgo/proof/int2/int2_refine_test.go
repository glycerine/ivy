package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRefineCubeSourceRejectsRegisterCI(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	assignment := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	priority := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("C source rejects the register CI as not an AND")
		}
	}()
	Int2_ManRefineCube(g, assignment, priority)
}

func TestPreimageReportsUnsupportedForPrimaryInput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, s := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, x, s))
	prefix, suffix := sat_solver_new(), sat_solver_new()
	ci := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	co := vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	prio := vec.Vec_IntAllocArrayCopy([]int32{1, 1}, 2)
	image, status := Int2_ManComputePreimage(g, prefix, suffix, ci, co, prio)
	if image != nil || status != INT2_TIME_OUT {
		t.Fatalf("image %v status %d", image, status)
	}
}
