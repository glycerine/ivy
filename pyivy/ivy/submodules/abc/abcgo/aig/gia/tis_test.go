package gia

import "testing"

func TestTisMuxAndCone(t *testing.T) {
	p := Gia_ManStart(20)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendMux(p, a, b, c)
	Gia_ManAppendCo(p, x)
	q := Gia_ManTisDupMuxes(p)
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		old, e := p.Evaluate(in)
		out, f := q.Evaluate(in)
		if e != nil || f != nil || old[0] != out[0] {
			t.Fatalf("input=%v old=%v new=%v", in, old, out)
		}
	}
	Gia_ManCreateRefs(q)
	root := Gia_ObjFaninId0p(q, Gia_ManCo(q, 0))
	nodes, leaves := Vec_IntAlloc(4), Vec_IntAlloc(4)
	Gia_ManTisCollectMffc(q, root, nodes, leaves)
	if Vec_IntSize(nodes) != 1 || Vec_IntSize(leaves) != 3 {
		t.Fatalf("nodes=%v leaves=%v", Vec_IntArray(nodes), Vec_IntArray(leaves))
	}
}
