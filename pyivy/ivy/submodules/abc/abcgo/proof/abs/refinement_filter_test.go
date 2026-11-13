package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRefinementFiltersSharedCutLeaves(t *testing.T) {
	for _, propagate := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, c)
		out := gia.Gia_ManAppendAnd(g, x, y)
		gia.Gia_ManAppendCo(g, out)
		size := int(gia.Gia_ManObjNum(g))
		cuts := vec.Vec_IntStart(size + 6)
		vec.Vec_IntWriteEntry(cuts, int(x>>1), int32(size))
		vec.Vec_IntWriteEntry(cuts, size, 2)
		vec.Vec_IntWriteEntry(cuts, size+1, a>>1)
		vec.Vec_IntWriteEntry(cuts, size+2, b>>1)
		vec.Vec_IntWriteEntry(cuts, int(y>>1), int32(size+3))
		vec.Vec_IntWriteEntry(cuts, size+3, 2)
		vec.Vec_IntWriteEntry(cuts, size+4, a>>1)
		vec.Vec_IntWriteEntry(cuts, size+5, c>>1)
		gia.Gia_ManSetMapping(g, cuts)
		p := Rnm_ManStart(g)
		mapping := vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(mapping, x>>1)
		vec.Vec_IntPush(mapping, y>>1)
		cex := util.Abc_CexAlloc(0, 2, 1)
		util.Abc_InfoSetBit(cex.PData, 0)
		util.Abc_InfoSetBit(cex.PData, 1)
		for repeat := 0; repeat < 2; repeat++ {
			got := vec.Vec_IntArray(Rnm_ManRefine(p, cex, mapping, propagate, 0, 0))
			if want := []int32{x >> 1, y >> 1}; !reflect.DeepEqual(got, want) {
				t.Fatalf("fanout=%d result=%v want=%v", propagate, got, want)
			}
			if Rnm_ObjCount(p, gia.Gia_ManObj(g, a>>1)) != 0 {
				t.Fatal("cut counter was not cleared")
			}
		}
		Rnm_ManStop(p, 0)
	}
}

func TestRefinementFilterDropsIndependentCut(t *testing.T) {
	g := gia.Gia_ManStart(20)
	a, b, c, d, e := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	x, y, z := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a, c), gia.Gia_ManAppendAnd(g, d, e)
	xy := gia.Gia_ManAppendAnd(g, x, y)
	root := gia.Gia_ManAppendAnd(g, xy, z)
	gia.Gia_ManAppendCo(g, root)
	size := int(gia.Gia_ManObjNum(g))
	cuts := vec.Vec_IntStart(size + 9)
	for i, cut := range [][3]int32{{x >> 1, a >> 1, b >> 1}, {y >> 1, a >> 1, c >> 1}, {z >> 1, d >> 1, e >> 1}} {
		off := size + 3*i
		vec.Vec_IntWriteEntry(cuts, int(cut[0]), int32(off))
		vec.Vec_IntWriteEntry(cuts, off, 2)
		vec.Vec_IntWriteEntry(cuts, off+1, cut[1])
		vec.Vec_IntWriteEntry(cuts, off+2, cut[2])
	}
	gia.Gia_ManSetMapping(g, cuts)
	p := Rnm_ManStart(g)
	defer Rnm_ManStop(p, 0)
	mapping := vec.Vec_IntAlloc(3)
	for _, id := range []int32{x >> 1, y >> 1, z >> 1} {
		vec.Vec_IntPush(mapping, id)
	}
	cex := util.Abc_CexAlloc(0, 3, 1)
	for i := int32(0); i < 3; i++ {
		util.Abc_InfoSetBit(cex.PData, i)
	}
	got := vec.Vec_IntArray(Rnm_ManRefine(p, cex, mapping, 0, 0, 0))
	if want := []int32{x >> 1, y >> 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("result=%v want=%v", got, want)
	}
}
