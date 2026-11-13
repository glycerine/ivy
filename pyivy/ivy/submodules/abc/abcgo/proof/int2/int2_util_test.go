package int2

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

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

func TestSourceSOPIndexSkipsSecondContiguousLiteral(t *testing.T) {
	sop := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 0, 3, 0}, 5)
	got := Int2_ManComputeCoPres(sop, 2)
	if vec.Vec_IntSize(got) != 1 || vec.Vec_IntEntry(got, 0) != 1 {
		t.Fatalf("source indexing yields only register 0, got %v", vec.Vec_IntArray(got))
	}
}

func TestSourceNilSOPPanics(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("source nil SOP is dereferenced")
		}
	}()
	Int2_ManProbToGia(g, nil)
}

func TestSourceEmptySOPProducesFalse(t *testing.T) {
	g := gia.Gia_ManStart(8)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	sop := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
	out := Int2_ManProbToGia(g, sop)
	for _, b := range []bool{false, true} {
		if int2EvalGraph(out, []bool{b}, nil, make([]bool, gia.Gia_ManObjNum(out)), nil) {
			t.Fatal("empty SOP must be false")
		}
	}
}
