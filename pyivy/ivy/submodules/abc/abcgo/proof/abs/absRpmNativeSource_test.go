package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// absRpm.c asserts that the support root is an AND before clearing vSupp,
// even when a primary input is marked and the recursive helper accepts it.
func TestSupport1RejectsMarkedInputBeforeChangingSupport(t *testing.T) {
	p := gia.Gia_ManStart(4)
	x := gia.Gia_ManAppendCi(p)
	obj := gia.Gia_ManObj(p, x>>1)
	gia.Gia_ObjSetMark1(obj, 1)
	support := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(support, 42)
	defer func() {
		if got := recover(); got != "abs: expected AND for support" {
			t.Fatalf("panic=%v", got)
		}
		if got := vec.Vec_IntArray(support); len(got) != 1 || got[0] != 42 {
			t.Fatalf("support changed before assertion: %v", got)
		}
	}()
	Abs_ManSupport1(p, obj, support)
}

// absRpmOld.c duplicates every Gia_ManForEachAnd node with AppendAnd on the
// first two physical fanins. Native XOR and MUX functions are not preserved.
func TestDupIn2FfNativeGatesFollowSourceAndCopy(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	xor := gia.Gia_ManAppendXorReal(p, a, b)
	p.EnableMuxes()
	mux := gia.Gia_ManAppendMuxReal(p, a, b, c)
	gia.Gia_ManAppendCo(p, xor)
	gia.Gia_ManAppendCo(p, mux)
	q := Gia_ManDupIn2Ff(p)
	for bits := 0; bits < 8; bits++ {
		av, bv, cv := bits&1 != 0, bits&2 != 0, bits&4 != 0
		// The three new PIs are followed by the original PIs as ROs.
		got, err := q.Evaluate([]bool{false, false, false, av, bv, cv})
		if err != nil || got[0] != (av && bv) || got[1] != (bv && cv) {
			t.Fatalf("inputs=%03b outputs=%v error=%v", bits, got, err)
		}
	}
}

// absRpm.c uses the same two-fanin AND reconstruction after treating marked
// source inputs as cuts; the MUX control is not visited on this path.
func TestDupRpmNativeGatesFollowSourceAndCopy(t *testing.T) {
	p := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	xor := gia.Gia_ManAppendXorReal(p, a, b)
	p.EnableMuxes()
	mux := gia.Gia_ManAppendMuxReal(p, a, b, c)
	gia.Gia_ManAppendCo(p, xor)
	gia.Gia_ManAppendCo(p, mux)
	for _, lit := range []int32{a, b, c} {
		gia.Gia_ObjSetMark1(gia.Gia_ManObj(p, lit>>1), 1)
	}
	q := Gia_ManDupRpm(p)
	for bits := 0; bits < 8; bits++ {
		values := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		inputs := make([]bool, gia.Gia_ManPiNum(q))
		for i, lit := range []int32{a, b, c} {
			index := int(gia.Gia_ManObj(p, lit>>1).Value>>1) - 1
			if index < 0 || index >= len(inputs) {
				t.Fatalf("original input %d has invalid copy index %d", i, index)
			}
			inputs[index] = values[i]
		}
		got, err := q.Evaluate(inputs)
		if err != nil || got[0] != (values[0] && values[1]) || got[1] != (values[1] && values[2]) {
			t.Fatalf("inputs=%03b outputs=%v error=%v", bits, got, err)
		}
	}
}
