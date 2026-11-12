package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestResubMffcReferencesRestored(t *testing.T) {
	g := Gia_ManStart(10)
	g.EnableMuxes()
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	out := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, out)
	Gia_ManCreateRefs(g)
	original := append([]int32(nil), g.Refs...)
	nodes, leaves, inners := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	if Gia_ObjCheckMffc(g, Gia_ManObj(g, out>>1), 20, nodes, leaves, inners) != 1 || !reflect.DeepEqual(vec.Vec_IntArray(leaves), []int32{a >> 1, b >> 1, c >> 1}) || !reflect.DeepEqual(vec.Vec_IntArray(inners), []int32{x >> 1, out >> 1}) {
		t.Fatalf("leaves %v inners %v", vec.Vec_IntArray(leaves), vec.Vec_IntArray(inners))
	}
	if !reflect.DeepEqual(original, g.Refs) {
		t.Fatal("reference counts changed")
	}
	if Gia_ObjCheckMffc(g, Gia_ManObj(g, out>>1), 0, nodes, leaves, inners) != 0 || !reflect.DeepEqual(original, g.Refs) {
		t.Fatal("failed-limit references not restored")
	}
}
func TestResubMffcCollection(t *testing.T) {
	g := Gia_ManStart(10)
	g.EnableMuxes()
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	out := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, out)
	m := Gia_ManComputeMffcs(g, 1, 20, 5, 100)
	if vec.Vec_WecSize(m) != 2 {
		t.Fatalf("MFFC count %d", vec.Vec_WecSize(m))
	}
	Gia_ManAddDivisors(g, m)
}
