package int2

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestInt2ManComputeCoPres(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		got := Int2_ManComputeCoPres(nil, 3)
		defer vec.Vec_IntFree(got)
		if want := []int32{0}; !reflect.DeepEqual(vec.Vec_IntArray(got), want) {
			t.Fatalf("COs = %v, want %v", vec.Vec_IntArray(got), want)
		}
	})
	t.Run("collects distinct register outputs in order", func(t *testing.T) {
		// SOP format: reserved word, then cube sizes followed by literals.
		sop := vec.Vec_IntAllocArrayCopy([]int32{0, 3, 5, 2, 5}, 5)
		defer vec.Vec_IntFree(sop)
		got := Int2_ManComputeCoPres(sop, 3)
		defer vec.Vec_IntFree(got)
		if want := []int32{2, 3}; !reflect.DeepEqual(vec.Vec_IntArray(got), want) {
			t.Fatalf("COs = %v, want %v", vec.Vec_IntArray(got), want)
		}
	})
}

func TestProbToGiaInitialImagePreservesProperty(t *testing.T) {
	g := gia.Gia_ManStart(8)
	input, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, input)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	suffix := Int2_ManProbToGia(g, nil)
	defer gia.Gia_ManStop(suffix)
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		got := int2EvalGraph(suffix, inputs, nil, make([]bool, gia.Gia_ManObjNum(suffix)), nil)
		if got != inputs[0] {
			t.Fatalf("inputs %v got %v, want property %v", inputs, got, inputs[0])
		}
	}
}

func TestProbToGiaSubstitutesNextStateIntoCubes(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, b)
	gia.Gia_ManAppendCo(g, a^1)
	gia.Gia_ManSetRegNum(g, 2)
	sop := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 0, 3}, 4) // next[0] && !next[1]
	suffix := Int2_ManProbToGia(g, sop)
	defer gia.Gia_ManStop(suffix)
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		got := int2EvalGraph(suffix, inputs, nil, make([]bool, gia.Gia_ManObjNum(suffix)), nil)
		if want := inputs[0] && inputs[1]; got != want {
			t.Fatalf("inputs %v got %v want %v", inputs, got, want)
		}
	}
}

func TestInt2ManComputeCoPresRejectsMalformedSOP(t *testing.T) {
	sop := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 1}, 3)
	defer vec.Vec_IntFree(sop)
	defer func() {
		if recover() == nil {
			t.Fatal("expected malformed SOP panic")
		}
	}()
	Int2_ManComputeCoPres(sop, 2)
}

// Compare suffix construction against direct simulation of the source graph,
// including support cones with shared, complemented and native gates.
func TestProbToGiaNativeGateCones(t *testing.T) {
	for _, kind := range []string{"xor", "mux", "buffer"} {
		for phase := int32(0); phase < 8; phase++ {
			g := gia.Gia_ManStart(32)
			a, b, s := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
			var gate int32
			switch kind {
			case "xor":
				gate = gia.Gia_ManAppendXorReal(g, a^(phase&1), s^((phase>>1)&1))
			case "mux":
				g.EnableMuxes()
				gate = gia.Gia_ManAppendMuxReal(g, a^(phase&1), b^((phase>>1)&1), s^((phase>>2)&1))
			case "buffer":
				gate = gia.Gia_ManAppendBuf(g, s^(phase&1))
			}
			gia.Gia_ManAppendCo(g, gate^1)
			gia.Gia_ManAppendCo(g, gate)
			gia.Gia_ManSetRegNum(g, 1)
			// Compare both the initial property and a negated next-state cube.
			for which := 0; which < 2; which++ {
				var sop *vec.Vec_Int_t
				if which == 1 {
					sop = vec.Vec_IntAllocArrayCopy([]int32{0, 1, 1}, 3)
				}
				out := Int2_ManProbToGia(g, sop)
				for mask := 0; mask < 8; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
					orig, err := g.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					got, err := out.Evaluate(inputs)
					if err != nil {
						t.Fatal(err)
					}
					want := orig[0]
					if which == 1 {
						want = !orig[1]
					}
					if got[0] != want {
						t.Fatalf("%s phase=%d suffix=%d input=%v: got %v want %v", kind, phase, which, inputs, got[0], want)
					}
				}
				gia.Gia_ManStop(out)
			}
		}
	}
}
