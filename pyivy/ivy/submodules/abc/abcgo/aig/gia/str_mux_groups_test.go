package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestStructuredMuxTreeGrouping(t *testing.T) {
	p := Gia_ManStart(16)
	var in [5]int32
	for i := range in {
		in[i] = Gia_ManAppendCi(p)
	}
	p.EnableMuxes()
	child := Gia_ManAppendMuxReal(p, in[0], in[1], in[2])
	root := Gia_ManAppendMuxReal(p, in[3], child, in[4])
	Gia_ManAppendCo(p, root)
	groups := Str_ManDeriveTrees(p)
	if vec.Vec_WecSize(groups) != 2 || vec.Vec_IntSize(vec.Vec_WecEntry(groups, 1)) != 1 || vec.Vec_IntEntry(vec.Vec_WecEntry(groups, 1), 0) != root>>1 {
		t.Fatal("mux tree was not grouped")
	}
	nodes := vec.Vec_IntAlloc(8)
	Str_MuxStructCollect(p, Gia_ManObj(p, root>>1), nodes)
	if vec.Vec_IntSize(nodes) != 2 || vec.Vec_IntEntry(nodes, 0) != child>>1 || vec.Vec_IntEntry(nodes, 1) != root>>1 {
		t.Fatalf("wrong mux order %v", vec.Vec_IntArray(nodes))
	}
	inputs := vec.Vec_IntAlloc(8)
	Str_MuxInputsCollect(p, Gia_ManObj(p, root>>1), inputs)
	if vec.Vec_IntSize(inputs) != 5 {
		t.Fatalf("wrong inputs %v", vec.Vec_IntArray(inputs))
	}
	roots := Str_ManCreateRoots(groups, p.nObjs)
	if vec.Vec_IntEntry(roots, int(root>>1)) != 1 {
		t.Fatal("wrong root group")
	}
	Str_ManCheckOverlap(p, groups)
}
