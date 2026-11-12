package int2

import (
	"reflect"
	"testing"

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
