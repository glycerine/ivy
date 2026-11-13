package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAlternateRefinementCollect(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	node := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, node)
	mapping := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(mapping, a>>1)
	vec.Vec_IntPush(mapping, b>>1)
	p := &Rf2_Man_t{pGia: g, vMap: mapping, vObjs: vec.Vec_IntAlloc(2)}
	for i := 0; i < 2; i++ {
		Rf2_ManCollect(p)
		want := []int32{node >> 1, gia.Gia_ObjId(g, gia.Gia_ManPo(g, 0))}
		if got := vec.Vec_IntArray(p.vObjs); !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestAlternateRefinementLifecycle(t *testing.T) {
	g := gia.Gia_ManStart(4)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	p := Rf2_ManStart(g)
	if len(p.pvVecs) != int(gia.Gia_ManObjNum(g)) || Rf2_ManMemoryUsage(p) <= 0 {
		t.Fatal("alternate manager not initialized")
	}
	Rf2_ManStop(p, 0)
	if p.pvVecs != nil || p.vObjs != nil {
		t.Fatal("alternate manager not released")
	}
}
