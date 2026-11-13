package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRefinementCollectConeAndClearValues(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	and := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, and)
	p := &Rnm_Man_t{pGia: g, vMap: vec.Vec_IntAlloc(2), vObjs: vec.Vec_IntAlloc(2)}
	vec.Vec_IntPush(p.vMap, a>>1)
	vec.Vec_IntPush(p.vMap, b>>1)
	for repeat := 0; repeat < 2; repeat++ {
		Rnm_ManCollect(p)
		got := append([]int32(nil), vec.Vec_IntArray(p.vObjs)...)
		want := []int32{and >> 1, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0))}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("objects=%v want=%v", got, want)
		}
		for i, id := range append(append([]int32(nil), vec.Vec_IntArray(p.vMap)...), got...) {
			if value := gia.Gia_ManObj(g, id).Value; value != uint32(i+1) {
				t.Fatalf("object %d value=%d", id, value)
			}
		}
		Rnm_ManCleanValues(p)
		for _, id := range append(append([]int32(nil), vec.Vec_IntArray(p.vMap)...), got...) {
			if gia.Gia_ManObj(g, id).Value != 0 {
				t.Fatalf("object %d was not cleared", id)
			}
		}
	}
}

func TestRefinementSensitization(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	and := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, and)
	for _, mapping := range [][]int32{{a >> 1, b >> 1}, {a >> 1, and >> 1}} {
		for mask := 0; mask < 4; mask++ {
			p := &Rnm_Man_t{pGia: g, vMap: vec.Vec_IntAlloc(2), vObjs: vec.Vec_IntAlloc(2)}
			for _, id := range mapping {
				vec.Vec_IntPush(p.vMap, id)
			}
			p.pCex = util.Abc_CexAlloc(0, 2, 1)
			for i := int32(0); i < 2; i++ {
				if mask&(1<<i) != 0 {
					util.Abc_InfoSetBit(p.pCex.PData, i)
				}
			}
			Rnm_ManCollect(p)
			p.nObjsFrame = int32(1 + vec.Vec_IntSize(p.vMap) + vec.Vec_IntSize(p.vObjs))
			p.pObjs = make([]Rnm_Obj_t, p.nObjsFrame)
			priority := Rnm_ManSensitize(p)
			out := Rnm_ManObj(p, gia.Gia_ManPo(g, 0), 0)
			want := mask == 3
			if mapping[1] == and>>1 {
				want = mask&2 != 0
			}
			if (out.Value != 0) != want {
				t.Fatalf("map=%v bits=%d got=%d", mapping, mask, out.Value)
			}
			if mapping[1] == and>>1 && priority != 2 {
				t.Fatalf("PPI priority %d", priority)
			}
			Rnm_ManCleanValues(p)
		}
	}
}

func TestRefinementJustifiesPseudoInputs(t *testing.T) {
	for _, propagate := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		left, right := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, c, d)
		top := gia.Gia_ManAppendAnd(g, left, right)
		gia.Gia_ManAppendCo(g, top)
		gia.Gia_ManStaticFanoutStart(g)
		p := &Rnm_Man_t{pGia: g, vMap: vec.Vec_IntAlloc(2), vObjs: vec.Vec_IntAlloc(3), fPropFanout: propagate}
		vec.Vec_IntPush(p.vMap, left>>1)
		vec.Vec_IntPush(p.vMap, right>>1)
		p.pCex = util.Abc_CexAlloc(0, 2, 1)
		util.Abc_InfoSetBit(p.pCex.PData, 0)
		util.Abc_InfoSetBit(p.pCex.PData, 1)
		Rnm_ManCollect(p)
		p.nObjsFrame = int32(1 + vec.Vec_IntSize(p.vMap) + vec.Vec_IntSize(p.vObjs))
		p.pObjs = make([]Rnm_Obj_t, p.nObjsFrame)
		Rnm_ManSensitize(p)
		selected := vec.Vec_IntAlloc(2)
		Rnm_ManJustify_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManPo(g, 0)), 0, selected)
		vec.Vec_IntUniqify(selected)
		if got, want := vec.Vec_IntArray(selected), []int32{left >> 1, right >> 1}; !reflect.DeepEqual(got, want) {
			t.Fatalf("fanout=%d selected=%v want=%v", propagate, got, want)
		}
	}
}

func TestRefinementManagerLifecycle(t *testing.T) {
	for _, preexisting := range []bool{false, true} {
		g := gia.Gia_ManStart(4)
		a := gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, a)
		if preexisting {
			gia.Gia_ManStaticFanoutStart(g)
		}
		p := Rnm_ManStart(g)
		if !gia.Gia_ManHasStaticFanout(g) || Rnm_ManMemoryUsage(p) <= 0 {
			t.Fatal("manager was not initialized")
		}
		Rnm_ManStop(p, 0)
		if gia.Gia_ManHasStaticFanout(g) {
			t.Fatal("manager stop did not release static fanouts")
		}
	}
}

func TestRefinementTrueCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	both := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, both)
	p := Rnm_ManStart(g)
	defer Rnm_ManStop(p, 0)
	mapping := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(mapping, a>>1)
	vec.Vec_IntPush(mapping, b>>1)
	cex := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	util.Abc_InfoSetBit(cex.PData, 1)
	for i := 0; i < 2; i++ {
		got := Rnm_ManRefine(p, cex, mapping, 0, 0, 0)
		if vec.Vec_IntSize(got) != 0 {
			t.Fatalf("true counterexample selected %v", vec.Vec_IntArray(got))
		}
		if gia.Gia_ObjTerSimGet1(gia.Gia_ManPo(g, 0)) == 0 {
			t.Fatal("ternary simulation did not prove output")
		}
	}
}
