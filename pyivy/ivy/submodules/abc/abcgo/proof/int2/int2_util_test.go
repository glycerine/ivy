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
