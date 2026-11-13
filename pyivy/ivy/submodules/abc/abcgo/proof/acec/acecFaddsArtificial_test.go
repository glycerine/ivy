package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func carryFixture(phase bool) (*gia.Gia_Man_t, int32, int32, int32) {
	g := gia.Gia_ManStart(32)
	a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x := gia.Gia_ManAppendAnd(g, a, b)
	chain := x
	if phase {
		chain ^= 1
	}
	y := gia.Gia_ManAppendAnd(g, chain, c)
	z := gia.Gia_ManAppendAnd(g, y, d)
	gia.Gia_ManAppendCo(g, z)
	gia.Gia_ManAppendCo(g, y)
	return g, x >> 1, y >> 1, z >> 1
}
func markCarry(g *gia.Gia_Man_t, node, fanin int32) {
	obj := gia.Gia_ManObj(g, node)
	if gia.Gia_ObjFaninId0(obj, node) == fanin {
		gia.Gia_ObjSetMark0(obj, 1)
	} else if gia.Gia_ObjFaninId1(obj, node) == fanin {
		gia.Gia_ObjSetMark1(obj, 1)
	} else {
		panic("missing fixture fanin")
	}
}
func TestArtificialBoxesCollapsePreservesFunction(t *testing.T) {
	for _, invert := range []bool{false, true} {
		for _, useFanout := range []int32{0, 1} {
			for _, xorTrick := range []int32{0, 1} {
				g, x, y, z := carryFixture(invert)
				markCarry(g, y, x)
				markCarry(g, z, y)
				boxed := Gia_ManDupWithArtificalFaddBoxes(g, useFanout, xorTrick)
				timing, extra := boxed.BoxTiming()
				if timing == nil || extra == nil || tim.Tim_ManBoxNum(timing) != 2 {
					t.Fatalf("missing boxes invert=%v fanout=%d xor=%d", invert, useFanout, xorTrick)
				}
				collapsed := gia.Gia_ManDupCollapse(boxed, extra, nil, 0)
				for mask := 0; mask < 16; mask++ {
					inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
					want, e1 := g.Evaluate(inputs)
					got, e2 := collapsed.Evaluate(inputs)
					if e1 != nil || e2 != nil || len(got) != len(want) {
						t.Fatalf("bad graphs: %v %v", e1, e2)
					}
					for i := range want {
						if got[i] != want[i] {
							t.Fatalf("invert=%v fanout=%d xor=%d mask=%d got=%v want=%v", invert, useFanout, xorTrick, mask, got, want)
						}
					}
				}
			}
		}
	}
}
func TestArtificialCarryPathAndDelay(t *testing.T) {
	g, x, y, z := carryFixture(false)
	var boxes int32
	if got := Gia_ManFindAnnotatedDelay(g, 10, &boxes, 1); got != 300 || boxes != 0 {
		t.Fatalf("initial delay=%d boxes=%d", got, boxes)
	}
	path := vec.Vec_IntAlloc(8)
	for i := int32(0); i < gia.Gia_ManCiNum(g); i++ {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManCi(g, i), 1)
	}
	if n := Gia_ManFindPath(g, 10, 1, 8, path); n != 2 {
		t.Fatalf("path %v", vec.Vec_IntArray(path))
	}
	if got := Gia_ManFindAnnotatedDelay(g, 10, &boxes, 1); got != 120 || boxes != 2 {
		t.Fatalf("annotated delay=%d boxes=%d", got, boxes)
	}
	if gia.Gia_ObjPhase(gia.Gia_ManObj(g, x)) == 0 || gia.Gia_ObjPhase(gia.Gia_ManObj(g, y)) == 0 || gia.Gia_ObjPhase(gia.Gia_ManObj(g, z)) != 0 {
		t.Fatal("incorrect carry annotation")
	}
}

func TestAutomaticArtificialBoxes(t *testing.T) {
	for _, limit := range []int32{0, 1} {
		g, _, _, _ := carryFixture(false)
		boxed := Gia_ManDupWithArtificialBoxes(g, 10, 1, 8, limit, 0, 0, 1, 0)
		timing, extra := boxed.BoxTiming()
		wantBoxes := 0
		if limit > 0 {
			wantBoxes = 2
		}
		if timing == nil || tim.Tim_ManBoxNum(timing) != wantBoxes {
			t.Fatalf("limit=%d boxes=%d want %d", limit, tim.Tim_ManBoxNum(timing), wantBoxes)
		}
		for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
			obj := gia.Gia_ManObj(g, id)
			if gia.Gia_ObjMark0(obj) != 0 || gia.Gia_ObjMark1(obj) != 0 || gia.Gia_ObjPhase(obj) != 0 {
				t.Fatal("temporary annotation not cleaned")
			}
		}
		collapsed := gia.Gia_ManDupCollapse(boxed, extra, nil, 0)
		for mask := 0; mask < 16; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			a, err1 := g.Evaluate(in)
			b, err2 := collapsed.Evaluate(in)
			if err1 != nil || err2 != nil || len(a) != len(b) {
				t.Fatal("bad collapsed graph")
			}
			for i := range a {
				if a[i] != b[i] {
					t.Fatalf("limit=%d mask=%d got=%v want=%v", limit, mask, b, a)
				}
			}
		}
	}
}
