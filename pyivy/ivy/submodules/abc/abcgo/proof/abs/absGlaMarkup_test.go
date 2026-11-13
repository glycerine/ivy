package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaMarkupRecordsTruthAndRegisterCut(t *testing.T) {
	for _, simple := range []int32{0, 1} {
		t.Run(map[int32]string{0: "cuts", 1: "simple"}[simple], func(t *testing.T) {
			p := gia.Gia_ManStart(10)
			x := gia.Gia_ManAppendCi(p)
			q := gia.Gia_ManAppendCi(p)
			and := gia.Gia_ManAppendAnd(p, x, q^1)
			gia.Gia_ManAppendCo(p, and)
			gia.Gia_ManAppendCo(p, x^1)
			gia.Gia_ManSetRegNum(p, 1)
			if count := Ga2_ManMarkup(p, 3, simple); count != 2 {
				t.Fatalf("marked %d", count)
			}
			m := gia.Gia_ManMapping(p)
			ro := int(vec.Vec_IntEntry(m, int(q>>1)))
			if vec.Vec_IntEntry(m, ro) != 1 || vec.Vec_IntEntry(m, ro+1) != x>>1 || uint32(vec.Vec_IntEntry(m, ro+2)) != 0x55555555 {
				t.Fatal("bad register cut")
			}
			off := int(vec.Vec_IntEntry(m, int(and>>1)))
			if vec.Vec_IntEntry(m, off) != 2 || uint32(vec.Vec_IntEntry(m, off+3)) != 0x22222222 {
				t.Fatalf("bad AND truth: %08x", uint32(vec.Vec_IntEntry(m, off+3)))
			}
		})
	}
}

func TestGlaMarkupSplitsLongCone(t *testing.T) {
	p := gia.Gia_ManStart(20)
	a := gia.Gia_ManAppendCi(p)
	b := gia.Gia_ManAppendCi(p)
	c := gia.Gia_ManAppendCi(p)
	d := gia.Gia_ManAppendCi(p)
	ab := gia.Gia_ManAppendAnd(p, a, b)
	abc := gia.Gia_ManAppendAnd(p, ab, c)
	root := gia.Gia_ManAppendAnd(p, abc, d)
	gia.Gia_ManAppendCo(p, root)
	if count := Ga2_ManMarkup(p, 3, 0); count != 2 {
		t.Fatalf("expected two marked cuts, got %d", count)
	}
	mapping := gia.Gia_ManMapping(p)
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(obj) == 0 || gia.Gia_ObjPhase(obj) == 0 {
			continue
		}
		off := int(vec.Vec_IntEntry(mapping, int(id)))
		size := vec.Vec_IntEntry(mapping, off)
		if size < 1 || size > 3 {
			t.Fatalf("node %d has %d leaves", id, size)
		}
	}
}
