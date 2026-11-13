package int2

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRefineCubeSelectsRelevantRegisterInput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	assignment := vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2)
	priority := vec.Vec_IntAllocArrayCopy([]int32{util.ABC_INFINITY, 1}, 2)
	subset := Int2_ManRefineCube(g, assignment, priority)
	if vec.Vec_IntSize(subset) != 1 || vec.Vec_IntEntry(subset, 0) != 1 {
		t.Fatalf("selected CI indices %v, want [1]", vec.Vec_IntArray(subset))
	}
}

func TestRefineCubeJustifiesBothInputsOfTrueAnd(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	gia.Gia_ManSetRegNum(g, 2)
	assignment := vec.Vec_IntAllocArrayCopy([]int32{1, 1}, 2)
	priority := vec.Vec_IntAllocArrayCopy([]int32{2, 3}, 2)
	subset := Int2_ManRefineCube(g, assignment, priority)
	if vec.Vec_IntSize(subset) != 2 {
		t.Fatalf("selected CI indices %v, want both registers", vec.Vec_IntArray(subset))
	}
}

func TestRefineCubeChoosesLowestPriorityControllingInput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate^1)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	gia.Gia_ManSetRegNum(g, 2)
	assignment := vec.Vec_IntAllocArrayCopy([]int32{0, 0}, 2)
	priority := vec.Vec_IntAllocArrayCopy([]int32{9, 2}, 2)
	subset := Int2_ManRefineCube(g, assignment, priority)
	if vec.Vec_IntSize(subset) != 1 || vec.Vec_IntEntry(subset, 0) != 1 {
		t.Fatalf("selected CI indices %v, want [1]", vec.Vec_IntArray(subset))
	}
}

func TestRefineCubeSoundForAllAndPhases(t *testing.T) {
	for phase := int32(0); phase < 8; phase++ {
		g := gia.Gia_ManStart(8)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gate := gia.Gia_ManAppendAnd(g, a^(phase&1), b^((phase>>1)&1))
		gia.Gia_ManAppendCo(g, gate^((phase>>2)&1))
		gia.Gia_ManAppendCo(g, a)
		gia.Gia_ManAppendCo(g, b)
		gia.Gia_ManSetRegNum(g, 2)
		for mask := 0; mask < 4; mask++ {
			state := []bool{mask&1 != 0, mask&2 != 0}
			if !int2EvalGraph(g, nil, state, make([]bool, gia.Gia_ManObjNum(g)), make([]bool, 2)) {
				continue
			}
			assignment := vec.Vec_IntAllocArrayCopy([]int32{int32(mask & 1), int32(mask >> 1 & 1)}, 2)
			priority := vec.Vec_IntAllocArrayCopy([]int32{2, 3}, 2)
			subset := Int2_ManRefineCube(g, assignment, priority)
			for candidate := 0; candidate < 4; candidate++ {
				matches := true
				for _, index := range vec.Vec_IntArray(subset) {
					if (mask>>index)&1 != (candidate>>index)&1 {
						matches = false
					}
				}
				if !matches {
					continue
				}
				other := []bool{candidate&1 != 0, candidate&2 != 0}
				if !int2EvalGraph(g, nil, other, make([]bool, gia.Gia_ManObjNum(g)), make([]bool, 2)) {
					t.Fatalf("phase %d assignment %d subset %v fails at %d", phase, mask, vec.Vec_IntArray(subset), candidate)
				}
			}
		}
	}
}

func TestRefineCubeSupportsCombinationalSuffix(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b))
	assignment := vec.Vec_IntAllocArrayCopy([]int32{1, 1}, 2)
	priority := vec.Vec_IntAllocArrayCopy([]int32{2, 3}, 2)
	subset := Int2_ManRefineCube(g, assignment, priority)
	if vec.Vec_IntSize(subset) != 2 {
		t.Fatalf("selected CI indices %v, want both inputs", vec.Vec_IntArray(subset))
	}
}
