package gia

import "testing"

func TestEmbeddingSimpleGraph(t *testing.T) {
	g := Gia_ManStart(10)
	a := Gia_ManAppendCi(g)
	b := Gia_ManAppendCi(g)
	node := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, node)
	p := Emb_ManStartSimple(g)
	if p.nObjs != 5 || Emb_ManCiNum(p) != 2 || Emb_ManNodeNum(p) != 2 {
		t.Fatalf("unexpected graph sizes: %+v", p)
	}
	n := Emb_ManObj(p, Gia_ObjValue(Gia_ManObj(g, node>>1)))
	if Emb_ObjFaninNum(n) != 2 || Emb_ObjFanin(n, 0) != Emb_ManCi(p, 0) || Emb_ObjFanin(n, 1) != Emb_ManCi(p, 1) {
		t.Fatal("fanin links incorrect")
	}
	if Emb_ObjFanout(Emb_ManCi(p, 0), 0) != n || Emb_ObjFanout(n, 0) != Emb_ManCo(p, 0) {
		t.Fatal("fanout links incorrect")
	}
	Emb_ManSetValue(p)
	Emb_ManIncrementTravId(p)
	Emb_ObjSetTravIdCurrent(p, n)
	if Emb_ObjIsTravIdCurrent(p, n) == 0 {
		t.Fatal("traversal mark missing")
	}
	Emb_ManResetTravId(p)
	if Emb_ObjIsTravIdCurrent(p, n) != 0 {
		t.Fatal("traversal mark not reset")
	}
}

func TestEmbeddingSimpleRegisterLinks(t *testing.T) {
	g := Gia_ManStart(10)
	input, reg := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	node := Gia_ManAppendAnd(g, input, reg)
	Gia_ManAppendCo(g, node)
	Gia_ManAppendCo(g, input)
	Gia_ManSetRegNum(g, 1)
	p := Emb_ManStartSimple(g)
	ro, ri := Emb_ManCi(p, 1), Emb_ManCo(p, 1)
	if Emb_ManPiNum(p) != 1 || Emb_ManPoNum(p) != 1 || Emb_ObjFaninNum(ro) != 1 || Emb_ObjFanin(ro, 0) != ri {
		t.Fatal("missing register feedback")
	}
	if Emb_ObjFanoutNum(ri) != 2 || Emb_ObjFanout(ri, 0) != ro || Emb_ObjFanout(ri, 1) != ri {
		t.Fatal("register output fanouts incorrect")
	}
}

func TestEmbeddingCollapsesSupernodes(t *testing.T) {
	g := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	y := Gia_ManAppendAnd(g, x, c)
	Gia_ManAppendCo(g, y)
	p := Emb_ManStart(g)
	if p.nObjs != 6 {
		t.Fatalf("object count = %d", p.nObjs)
	}
	if Gia_ObjValue(Gia_ManObj(g, x>>1)) != -1 {
		t.Fatal("internal supernode unexpectedly retained")
	}
	root := Emb_ManObj(p, Gia_ObjValue(Gia_ManObj(g, y>>1)))
	if Emb_ObjFaninNum(root) != 3 {
		t.Fatalf("supernode fanins = %d", Emb_ObjFaninNum(root))
	}
	seen := make(map[*Emb_Obj_t]bool)
	for i := int32(0); i < 3; i++ {
		seen[Emb_ObjFanin(root, i)] = true
	}
	for i := int32(0); i < 3; i++ {
		if !seen[Emb_ManCi(p, i)] {
			t.Fatalf("input %d absent", i)
		}
	}
	if g.Refs != nil {
		t.Fatal("temporary references retained")
	}
}

func TestEmbeddingDistances(t *testing.T) {
	g := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	Gia_ManAppendCo(g, root)
	p := Emb_ManStart(g)
	Emb_ManSetValue(p)
	starts := Vec_IntAlloc(1)
	Vec_IntPush(starts, Emb_ManCo(p, 0).hHandle)
	d := make([]Emb_Dat_t, p.nObjs)
	furthest := Emb_ManFindDistances(p, starts, d)
	if p.nReached != 4 || p.nDistMax != 3 || d[furthest.Value] != 2 {
		t.Fatalf("BFS: reached=%d levels=%d far=%v", p.nReached, p.nDistMax, furthest)
	}
	if d[Emb_ManCi(p, 0).Value] != 2 || d[Emb_ManCi(p, 1).Value] != 2 {
		t.Fatalf("input distances: %v", d)
	}
	if components := Emb_ManConnectedComponents(p); Vec_IntSize(components) != 1 {
		t.Fatalf("components: %v", Vec_IntArray(components))
	}
}

func TestEmbeddingGnuplotRuns(t *testing.T) {
	p := &Emb_Man_t{objects: []*Emb_Obj_t{{Value: 0}, {Value: 1}, {Value: 2}}, nObjs: 3, pPlacement: []uint16{0, 0, 0, 0, 65535, 65535}}
	lines := Vec_IntArray(Emb_ManDumpGnuplotPrepare(p))
	if len(lines) != 8 || lines[0] != 0 || lines[1] != 0 || lines[2] != 0 || lines[3] != 0 || lines[4] != 699 || lines[5] != 499 {
		t.Fatalf("plot runs = %v", lines)
	}
}
