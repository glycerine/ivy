package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func assertIDs(t *testing.T, got *Vec_Int_t, want ...int32) {
	t.Helper()
	values := Vec_IntArray(got)
	if len(values) != len(want) {
		t.Fatalf("IDs %v != %v", values, want)
	}
	for i := range want {
		if values[i] != want[i] {
			t.Fatalf("IDs %v != %v", values, want)
		}
	}
}
func TestDfsSourceOrdersAndSharedEpoch(t *testing.T) {
	p := Gia_ManStart(8)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, c, x)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, x)
	roots := []int32{Gia_ObjId(p, Gia_ManCo(p, 0)), Gia_ObjId(p, Gia_ManCo(p, 1))}
	support := Vec_IntAlloc(3)
	Gia_ManCollectCis(p, roots, 2, support)
	assertIDs(t, support, 3, 1, 2)
	if Gia_ObjIsTravIdCurrentId(p, roots[0]) != 0 {
		t.Fatal("support collection must not mark CO wrappers")
	}
	nodes := Gia_ManCollectNodesCis(p, roots, 2)
	assertIDs(t, nodes, 3, 1, 2, 4, 5)
	if Gia_ManConeSize(p, roots, 2) != 2 {
		t.Fatal("shared cone counted twice")
	}
	Gia_ManIncrementTravId(p)
	work := Vec_IntAlloc(3)
	Gia_ManCollectAnds(p, roots, 1, work, nil)
	assertIDs(t, work, 4, 5)
	Gia_ManCollectAnds(p, roots[1:], 1, work, nil)
	assertIDs(t, work)
	Gia_ManLevelNum(p)
	order := Gia_ManOrderReverse(p)
	assertIDs(t, order, 6, 5, 7, 4, 0, 1, 2, 3)
	levelized := Gia_ManLevelize(p)
	if vec.Vec_PtrEntry(vec.Vec_VecEntry(levelized, 1), 0) != Gia_ManObj(p, 4) {
		t.Fatal("level bucket")
	}
	Gia_ManStaticFanoutStart(p)
	input := Vec_IntAlloc(1)
	Vec_IntPush(input, 4)
	Gia_ManCollectTfo(p, input, work)
	assertIDs(t, work, 5, 4)
	Gia_ManCollectTfi(p, input, work)
	assertIDs(t, work, 4)
}
func TestDfsRealMuxControlAndSequentialClosure(t *testing.T) {
	p := Gia_ManStart(8)
	pi, ro := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	other := Gia_ManAppendCi(p)
	p.EnableMuxes()
	mux := Gia_ManAppendMuxReal(p, pi, ro, other)
	Gia_ManAppendCo(p, mux)
	Gia_ManAppendCo(p, other)
	Gia_ManSetRegNum(p, 1) // 'other' is RO, with self-loop RI.
	support := Vec_IntAlloc(3)
	root := Gia_ObjId(p, Gia_ManCo(p, 0))
	Gia_ManCollectCis(p, []int32{root}, 1, support)
	if Vec_IntSize(support) != 3 {
		t.Fatal("real MUX control omitted", Vec_IntArray(support))
	}
	sequence := Gia_ManCollectSeq(p, []int32{0}, 1)
	seen := map[int32]bool{}
	for _, id := range Vec_IntArray(sequence) {
		if seen[id] {
			t.Fatal("duplicate sequential object", id)
		}
		seen[id] = true
	}
	if !seen[Gia_ObjId(p, Gia_ManRi(p, 0))] || !seen[Gia_ObjId(p, Gia_ManRo(p, 0))] || !seen[0] {
		t.Fatal("sequential closure missing register", seen)
	}
}
func TestDfsDeepConeAndConstants(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := a
	for i := 0; i < 20000; i++ {
		root = Gia_ManAppendAnd(p, root, b)
	}
	if Gia_ManConeSize(p, []int32{root >> 1, 0}, 2) != 20000 {
		t.Fatal("deep cone")
	}
}
