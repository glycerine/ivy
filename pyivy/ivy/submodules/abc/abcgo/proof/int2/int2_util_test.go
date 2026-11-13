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

func TestSOPSourceStrideBoundsRejectTruncatedCube(t *testing.T) {
	// The checked-in C loop reads the second literal at pos+2 because it
	// increments its outer index inside the cube loop. This truncated cube
	// passes the simple count bound but does not contain that source-indexed
	// element, so the Go helper must reject it before indexing the vector.
	sop := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 0, 1}, 4)
	defer vec.Vec_IntFree(sop)
	defer func() {
		if got := recover(); got != "int2: malformed SOP cube" {
			t.Fatalf("panic = %v, want malformed-cube panic", got)
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

func TestSourceCoPresDeduplicatesAndOrdersSingleLiteralCubes(t *testing.T) {
	// Single-literal cubes avoid the C loop's i+k indexing defect. The
	// source marks registers then visits them in increasing register order.
	sop := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 5, 1, 0, 1, 4, 1, 3}, 9)
	got := Int2_ManComputeCoPres(sop, 3)
	defer vec.Vec_IntFree(got)
	want := []int32{1, 2, 3}
	if entries := vec.Vec_IntArray(got); len(entries) != len(want) {
		t.Fatalf("CO indices %v, want %v", entries, want)
	} else {
		for i := range want {
			if entries[i] != want[i] {
				t.Fatalf("CO indices %v, want %v", entries, want)
			}
		}
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

// Each cube has one literal, so the source's increment-and-i+k indexing
// defect does not alter this case. The oracle evaluates the original
// transition functions rather than the constructed suffix graph.
func TestProbToGiaSingleLiteralCubesMatchSourceTransitions(t *testing.T) {
	g := gia.Gia_ManStart(24)
	in, x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n0 := gia.Gia_ManAppendAnd(g, in, x^1)
	n1 := gia.Gia_ManAppendAnd(g, in^1, y)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManAppendCo(g, n0)
	gia.Gia_ManAppendCo(g, n1^1)
	gia.Gia_ManSetRegNum(g, 2)
	// next[0] OR NOT next[1].
	sop := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 0, 1, 3}, 5)
	out := Int2_ManProbToGia(g, sop)
	defer gia.Gia_ManStop(out)
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		original, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		result, err := out.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		want := original[1] || !original[2]
		if result[0] != want {
			t.Fatalf("inputs %v: got %v want %v", inputs, result[0], want)
		}
	}
}

func TestSourceCollectInternalVisitsSharedConesOnce(t *testing.T) {
	g := gia.Gia_ManStart(24)
	a, b, c, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	shared := gia.Gia_ManAppendAnd(g, a, b)
	first := gia.Gia_ManAppendAnd(g, shared, c)
	second := gia.Gia_ManAppendAnd(g, shared, c^1)
	gia.Gia_ManAppendCo(g, first)
	gia.Gia_ManAppendCo(g, second)
	gia.Gia_ManSetRegNum(g, 1)
	_ = state
	cos := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 0}, 3)
	nodes := Int2_ManCollectInternal(g, cos)
	defer vec.Vec_IntFree(nodes)
	got := vec.Vec_IntArray(nodes)
	want := []int32{shared >> 1, first >> 1, second >> 1}
	if len(got) != len(want) {
		t.Fatalf("collected %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("collected %v, want %v", got, want)
		}
	}
}
