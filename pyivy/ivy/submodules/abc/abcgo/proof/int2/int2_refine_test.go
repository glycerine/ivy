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

func TestRefineCubeConstantTrueNeedsNoInputs(t *testing.T) {
	g := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 1)
	assignment := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	priority := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	subset := Int2_ManRefineCube(g, assignment, priority)
	if vec.Vec_IntSize(subset) != 0 {
		t.Fatalf("constant true output needs no inputs: %v", vec.Vec_IntArray(subset))
	}
}

func TestComputePreimageConstantTrueSuffix(t *testing.T) {
	suffix := gia.Gia_ManStart(8)
	gia.Gia_ManAppendCi(suffix)
	gia.Gia_ManAppendCo(suffix, 1)
	prefixSAT, suffixSAT := sat_solver_new(), sat_solver_new()
	sat_solver_setnvars(prefixSAT, 1)
	sat_solver_setnvars(suffixSAT, 1)
	ciMap := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	coMap := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	priority := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	image, status := Int2_ManComputePreimage(suffix, prefixSAT, suffixSAT, ciMap, coMap, priority)
	if status != INT2_FALSE_NEG || image != nil {
		t.Fatalf("reachable constant suffix: image %v status %d", image, status)
	}
}

func TestComputePreimageBlocksUnreachableSuffixState(t *testing.T) {
	suffix := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(suffix)
	gia.Gia_ManAppendCo(suffix, x)
	prefixSAT, suffixSAT := sat_solver_new(), sat_solver_new()
	sat_solver_setnvars(prefixSAT, 1)
	sat_solver_setnvars(suffixSAT, 1)
	if sat_solver_addclause(prefixSAT, []int32{1}) == 0 || sat_solver_addclause(suffixSAT, []int32{0}) == 0 {
		t.Fatal("could not construct test solvers")
	}
	ciMap := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	coMap := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	prio := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	image, status := Int2_ManComputePreimage(suffix, prefixSAT, suffixSAT, ciMap, coMap, prio)
	if status != INT2_COMPUTED || image == nil {
		t.Fatalf("preimage status %d image %v", status, image)
	}
	if got := vec.Vec_IntArray(image); len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 0 {
		t.Fatalf("unreachable state cube %v, want [0 1 0]", got)
	}
	// Feed the produced cover into the next image construction: an image
	// containing state=1, composed with nextState=!state, selects state=0.
	machine := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(machine)
	gia.Gia_ManAppendCo(machine, 0)
	gia.Gia_ManAppendCo(machine, state^1)
	gia.Gia_ManSetRegNum(machine, 1)
	composed := Int2_ManProbToGia(machine, image)
	defer gia.Gia_ManStop(composed)
	for _, value := range []bool{false, true} {
		got := int2EvalGraph(composed, []bool{value}, nil, make([]bool, gia.Gia_ManObjNum(composed)), nil)
		if got != !value {
			t.Fatalf("composed image at state=%t: got %t", value, got)
		}
	}
}

func TestComputePreimageTranslatesSATVariableMappings(t *testing.T) {
	suffix := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(suffix)
	gia.Gia_ManAppendCo(suffix, x)
	prefixSAT, suffixSAT := sat_solver_new(), sat_solver_new()
	sat_solver_setnvars(prefixSAT, 1)
	sat_solver_setnvars(suffixSAT, 2)
	if sat_solver_addclause(prefixSAT, []int32{1}) == 0 || sat_solver_addclause(suffixSAT, []int32{2}) == 0 {
		t.Fatal("could not construct test solvers")
	}
	ciMap := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	coMap := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	prio := vec.Vec_IntAllocArrayCopy([]int32{1}, 1)
	image, status := Int2_ManComputePreimage(suffix, prefixSAT, suffixSAT, ciMap, coMap, prio)
	if status != INT2_COMPUTED || image == nil {
		t.Fatalf("status %d image %v", status, image)
	}
	if got := vec.Vec_IntArray(image); len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 0 {
		t.Fatalf("unreachable state cube %v, want [0 1 0]", got)
	}
}

func TestRefineCubeNativeGateSoundness(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for phase := int32(0); phase < 16; phase++ {
			g := gia.Gia_ManStart(16)
			a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			var gate int32
			switch kind {
			case "xor":
				gate = gia.Gia_ManAppendXorReal(g, a^(phase&1), b^((phase>>1)&1))
			case "mux":
				g.EnableMuxes()
				gate = gia.Gia_ManAppendMuxReal(g, a^(phase&1), b^((phase>>1)&1), c^((phase>>2)&1))
			case "buffer":
				gate = gia.Gia_ManAppendBuf(g, a^(phase&1))
			}
			gia.Gia_ManAppendCo(g, gate^((phase>>3)&1))
			for _, lit := range []int32{a, b, c} {
				gia.Gia_ManAppendCo(g, lit)
			}
			gia.Gia_ManSetRegNum(g, 3)
			for mask := int32(0); mask < 8; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
				outputs, err := g.Evaluate(inputs)
				if err != nil {
					t.Fatal(err)
				}
				if !outputs[0] {
					continue
				}
				assignment := vec.Vec_IntAllocArrayCopy([]int32{mask & 1, (mask >> 1) & 1, (mask >> 2) & 1}, 3)
				priority := vec.Vec_IntAllocArrayCopy([]int32{1, 2, 3}, 3)
				selected := Int2_ManRefineCube(g, assignment, priority)
				for candidate := int32(0); candidate < 8; candidate++ {
					matches := true
					for _, index := range vec.Vec_IntArray(selected) {
						if (candidate>>index)&1 != (mask>>index)&1 {
							matches = false
						}
					}
					if !matches {
						continue
					}
					out, err := g.Evaluate([]bool{candidate&1 != 0, candidate&2 != 0, candidate&4 != 0})
					if err != nil {
						t.Fatal(err)
					}
					if !out[0] {
						t.Fatalf("kind=%s phase=%d mask=%d selected=%v candidate=%d", kind, phase, mask, vec.Vec_IntArray(selected), candidate)
					}
				}
			}
		}
	}
}
