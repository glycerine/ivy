package ivy

import "testing"

func TestIvyComplementEdgesAndIntegerEncoding(t *testing.T) {
	p := &Ivy_Man_t{}
	Ivy_ManStartMemory(p)
	defer Ivy_ManStopMemory(p)
	a := Ivy_ManFetchMemory(p)
	a.Id = 1
	a.Type = IVY_PI
	p.vObjs = []*Ivy_Obj_t{nil, a}
	negative := Ivy_Not(a)
	if Ivy_Not(negative) != a || Ivy_Not(a) != negative || Ivy_Regular(negative) != a || Ivy_IsComplement(negative) != 1 || Ivy_IsComplement(a) != 0 {
		t.Fatal("complement invariant")
	}
	if Ivy_ObjIsPi(negative) != 1 {
		t.Fatal("proxy type not regularized")
	}
	for _, edge := range []*Ivy_Obj_t{a, negative} {
		if Ivy_EdgeToNode(p, Ivy_EdgeFromNode(edge)) != edge {
			t.Fatal("integer edge round trip")
		}
	}
	large := Ivy_EdgeCreate(1<<30, 1)
	if Ivy_EdgeId(large) != 1<<30 || Ivy_EdgeIsComplement(large) != 1 {
		t.Fatal("signed edge shift overflow")
	}
	leaf := Ivy_LeafCreate(0xffffff, 255)
	if Ivy_LeafId(leaf) != 0xffffff || Ivy_LeafLat(leaf) != 255 {
		t.Fatal("sequential leaf packing")
	}
}
func TestIvyMemoryPagesAndOwnership(t *testing.T) {
	p := &Ivy_Man_t{}
	Ivy_ManStartMemory(p)
	nodes := make([]*Ivy_Obj_t, (1<<IVY_PAGE_SIZE)+1)
	seen := map[*Ivy_Obj_t]bool{}
	for i := range nodes {
		nodes[i] = Ivy_ManFetchMemory(p)
		if seen[nodes[i]] {
			t.Fatal("same live memory twice")
		}
		seen[nodes[i]] = true
		nodes[i].Id = int32(i)
	}
	if len(p.pages) != 2 {
		t.Fatal("page expansion", len(p.pages))
	}
	old := nodes[0]
	neg := Ivy_Not(old)
	Ivy_ManRecycleMemory(p, old)
	next := Ivy_ManFetchMemory(p)
	if next != old || Ivy_Not(next) != neg || next.owner != p || next.Type != IVY_NONE || next.Id != 0 {
		t.Fatal("memory reuse did not reset node/retain edge identity")
	}
	Ivy_ManStopMemory(p)
	Ivy_ManStopMemory(p)
	if len(p.pages) != 0 || len(p.freeNodes) != 0 || old.owner != nil {
		t.Fatal("memory ownership retained after stop")
	}
}
func TestIvyObjectCopyKeepsIdentity(t *testing.T) {
	p := &Ivy_Man_t{}
	Ivy_ManStartMemory(p)
	defer Ivy_ManStopMemory(p)
	base, data := Ivy_ManFetchMemory(p), Ivy_ManFetchMemory(p)
	base.Id = 7
	negative := Ivy_Not(base)
	data.Id = 19
	data.Type = IVY_AND
	data.Level = 9
	data.fMarkA = 1
	Ivy_ObjOverwrite(base, data)
	if base.Id != 7 || base.Level != 9 || base.Type != IVY_AND || Ivy_Not(base) != negative || base.owner != p {
		t.Fatal("overwrite changed identity")
	}
	Ivy_ObjClean(base)
	if base.Id != 7 || base.Type != IVY_NONE || base.Level != 0 || Ivy_Not(base) != negative || base.owner != p {
		t.Fatal("clean changed identity")
	}
}
func TestIvyFanoutPolarityUsesFanoutEdge(t *testing.T) {
	a := &Ivy_Obj_t{Id: 1, Type: IVY_PI}
	b := &Ivy_Obj_t{Id: 2, Type: IVY_PI}
	node := &Ivy_Obj_t{Id: 3, Type: IVY_AND, pFanin0: Ivy_Not(a), pFanin1: b}
	if Ivy_ObjFanoutC(a, node) != 1 || Ivy_ObjFanoutC(b, node) != 0 {
		t.Fatal("polarity read from driving node instead of fanout")
	}
	if Ivy_ObjWhatFanin(node, a) != 0 || Ivy_ObjWhatFanin(node, b) != 1 {
		t.Fatal("fanin selection")
	}
}
func TestIvyInitializationAndBits(t *testing.T) {
	values := []Ivy_Init_t{IVY_INIT_0, IVY_INIT_1, IVY_INIT_DC}
	for _, a := range values {
		if Ivy_InitNotCond(Ivy_InitNotCond(a, 1), 1) != a {
			t.Fatal("initialization negation")
		}
		for _, b := range values {
			wantAnd := IVY_INIT_1
			if a == IVY_INIT_0 || b == IVY_INIT_0 {
				wantAnd = IVY_INIT_0
			} else if a == IVY_INIT_DC || b == IVY_INIT_DC {
				wantAnd = IVY_INIT_DC
			}
			if Ivy_InitAnd(a, b) != wantAnd {
				t.Fatal(a, b)
			}
			wantXor := IVY_INIT_0
			if a == IVY_INIT_DC || b == IVY_INIT_DC {
				wantXor = IVY_INIT_DC
			} else if a != b {
				wantXor = IVY_INIT_1
			}
			if Ivy_InitExor(a, b) != wantXor {
				t.Fatal(a, b)
			}
		}
	}
	data := make([]uint32, Ivy_BitWordNum(65))
	for _, i := range []int32{0, 31, 32, 64} {
		Ivy_InfoSetBit(data, i)
		if Ivy_InfoHasBit(data, i) != 1 {
			t.Fatal(i)
		}
		Ivy_InfoXorBit(data, i)
		if Ivy_InfoHasBit(data, i) != 0 {
			t.Fatal(i)
		}
	}
}
func TestIvyMemoryRejectsInvalidRecycle(t *testing.T) {
	p, q := &Ivy_Man_t{}, &Ivy_Man_t{}
	Ivy_ManStartMemory(p)
	Ivy_ManStartMemory(q)
	defer Ivy_ManStopMemory(p)
	defer Ivy_ManStopMemory(q)
	node := Ivy_ManFetchMemory(p)
	mustPanic := func(f func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Error("invalid operation accepted")
			}
		}()
		f()
	}
	mustPanic(func() { Ivy_ManRecycleMemory(q, node) })
	mustPanic(func() { Ivy_ManRecycleMemory(p, Ivy_Not(node)) })
	node.nRefs = 1
	mustPanic(func() { Ivy_ManRecycleMemory(p, node) })
	node.nRefs = 0
	Ivy_ManRecycleMemory(p, node)
	Ivy_ObjClean(node)
	mustPanic(func() { Ivy_ManRecycleMemory(p, node) })
	Ivy_ManStopMemory(p)
	mustPanic(func() { Ivy_ManFetchMemory(p) })
}
func TestIvyManagerLifecycle(t *testing.T) {
	p := Ivy_ManStart()
	constant := Ivy_ManConst1(p)
	if constant.Id != 0 || constant.fPhase != 1 || constant.owner != p || Ivy_ManObjNum(p) != 1 || Ivy_ManNodeNum(p) != 0 || Ivy_ManPiNum(p) != 0 || len(p.vObjs) != 1 || p.Ghost.Id != -1 || len(p.pTable) != 0 {
		t.Fatal("manager initialization")
	}
	if Ivy_EdgeFromNode(constant) != 0 || Ivy_EdgeFromNode(Ivy_ManConst0(p)) != 1 {
		t.Fatal("constant polarity")
	}
	q := Ivy_ManStart()
	if q.pConst1 == p.pConst1 || q.pages[0][0].owner != q {
		t.Fatal("manager ownership mixed")
	}
	Ivy_ManStop(q)
	Ivy_ManStop(p)
	Ivy_ManStop(p)
	if p.pConst1 != nil || p.vObjs != nil || p.pTable != nil || len(p.pages) != 0 || constant.owner != nil {
		t.Fatal("stop retained owned state")
	}
}
func TestIvyReservedNodesCannotEnterFreeList(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	for _, node := range []*Ivy_Obj_t{p.pConst1, &p.Ghost} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("reserved node recycled")
				}
			}()
			Ivy_ManRecycleMemory(p, node)
		}()
	}
	if Ivy_BitWordNum(0x7fffffff) != 67108864 {
		t.Fatal("bit word count overflow")
	}
}
func TestIvyManagerStartFromInputBoundary(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a := Ivy_ManFetchMemory(p)
	b := Ivy_ManFetchMemory(p)
	a.Id, a.Type = 1, IVY_PI
	b.Id, b.Type = 2, IVY_PI
	p.vPis = []*Ivy_Obj_t{a, b}
	p.vObjs = append(p.vObjs, a, b)
	p.nObjs[IVY_PI] = 2
	p.nCreated += 2
	q := Ivy_ManStartFrom(p)
	defer Ivy_ManStop(q)
	if Ivy_ManPiNum(q) != 2 || Ivy_ManObjNum(q) != 3 || Ivy_ManPoNum(q) != 0 || Ivy_ManNodeNum(q) != 0 {
		t.Fatal("input boundary copied extra/missing objects")
	}
	if p.pConst1.pEquiv != q.pConst1 || a.pEquiv != q.vPis[0] || b.pEquiv != q.vPis[1] || Ivy_ObjEquiv(Ivy_Not(a)) != Ivy_Not(q.vPis[0]) {
		t.Fatal("source mapping lost phase")
	}
	if q.vPis[0].owner != q || q.vPis[0] == a {
		t.Fatal("input ownership shared")
	}
}
func TestIvyGhostRejectsForeignAndMalformedInputs(t *testing.T) {
	p, q := Ivy_ManStart(), Ivy_ManStart()
	defer Ivy_ManStop(p)
	defer Ivy_ManStop(q)
	a := Ivy_ManFetchMemory(p)
	a.Id = 1
	a.Type = IVY_PI
	b := Ivy_ManFetchMemory(q)
	b.Id = 2
	b.Type = IVY_PI
	for _, makeGhost := range []func(){func() { Ivy_ObjCreateGhost(p, a, b, IVY_AND, IVY_INIT_NONE) }, func() { Ivy_ObjCreateGhost(p, a, nil, IVY_AND, IVY_INIT_NONE) }, func() { Ivy_ObjCreateGhost(p, a, nil, IVY_LATCH, IVY_INIT_NONE) }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid ghost accepted")
				}
			}()
			makeGhost()
		}()
	}
}
func TestIvyManagerStopAfterMemoryStopDropsReferences(t *testing.T) {
	p := Ivy_ManStart()
	Ivy_ManStopMemory(p)
	Ivy_ManStop(p)
	if p.pConst1 != nil || p.vObjs != nil || p.pTable != nil {
		t.Fatal("manager references retained after memory had stopped")
	}
}
