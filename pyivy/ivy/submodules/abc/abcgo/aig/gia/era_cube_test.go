package gia

import "testing"

func TestReachabilityCubeRelations(t *testing.T) {
	a, b := &Gia_StaAre_t{pData: make([]uint32, 2)}, &Gia_StaAre_t{pData: make([]uint32, 2)}
	Gia_StaSetValue0(a, 0)
	Gia_StaSetValue0(b, 0)
	Gia_StaSetValue1(b, 17)
	if Gia_StaAreContain(a, b, 2) != 1 || Gia_StaAreDashNum(a, b, 2) != 1 || Gia_StaAreSharpVar(a, b, 2) != 17 || Gia_StaAreDisjoint(a, b, 2) != 0 {
		t.Fatal("cube containment or sharp incorrect")
	}
	a.pData[0] = 0
	Gia_StaSetValue1(a, 0)
	if Gia_StaAreDisjoint(a, b, 2) != 1 || Gia_StaAreDisjointVar(a, b, 2) != 0 {
		t.Fatal("cube disjointness incorrect")
	}
}
func TestReachabilityCubeMinterms(t *testing.T) {
	c := &Gia_StaAre_t{pData: make([]uint32, 1)}
	Gia_StaSetValue1(c, 1)
	Gia_StaSetValue0(c, 2)
	store := make([]uint32, 1)
	Gia_ManCountMintermsInCube(c, 3, store)
	if store[0] != (1<<2)|(1<<3) {
		t.Fatalf("minterms %08b", store[0])
	}
	ptr := Gia_Int2Ptr(0x92345678)
	if Gia_Ptr2Int(ptr) != 0x12345678 || ptr.fMark != 1 {
		t.Fatalf("pointer conversion %#v", ptr)
	}
}

func TestReachabilityInputFanoutCones(t *testing.T) {
	p := Gia_ManStart(12)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, c)
	co := Gia_ManAppendCo(p, y)
	cones := Gia_ManDeriveCiTfo(p)
	expected := [][]int32{{a >> 1, x >> 1, y >> 1, co >> 1}, {b >> 1, x >> 1, y >> 1, co >> 1}, {c >> 1, y >> 1, co >> 1}}
	for i, want := range expected {
		got := Vec_IntArray(Vec_PtrEntry(cones, i).(*Vec_Int_t))
		if len(got) != len(want) {
			t.Fatalf("cone %d = %v", i, got)
		}
		for j := range want {
			if got[j] != want[j] {
				t.Fatalf("cone %d = %v", i, got)
			}
		}
	}
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 {
			t.Fatal("fanout marks retained")
		}
	}
}

func TestCubeReachabilityManager(t *testing.T) {
	g := Gia_ManStart(12)
	a, r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, r0)
	Gia_ManAppendCo(g, r1)
	Gia_ManSetRegNum(g, 2)
	p := Gia_ManAreCreate(g)
	initial := Gia_ManAreCreateStaInit(p)
	if p.nStas != 2 || Gia_StaHasValue0(initial, 0) != 1 || Gia_StaHasValue0(initial, 1) != 1 || Gia_ManCountMinterms(p) != 1 {
		t.Fatal("incorrect initial state")
	}
	second := Gia_ManAreCreateSta(p)
	second.iPrev = Gia_Int2Ptr(1)
	Gia_StaSetValue1(second, 0)
	Gia_StaSetValue0(second, 1)
	if Gia_ManAreDepth(p, 2) != 2 || Gia_ManCountMinterms(p) != 2 {
		t.Fatal("state storage or counting incorrect")
	}
	Gia_ManAreRycycleSta(p, second)
	if p.nStas != 2 || Gia_ManCountMinterms(p) != 1 {
		t.Fatal("last state not recycled")
	}
	obj := Gia_ManAreCreateObj(p)
	obj.iVar = 9
	if Gia_ManAreObjInt(p, 1) != obj || Gia_ManAreObjLast(p).iVar != 9 {
		t.Fatal("object handle incorrect")
	}
	Gia_ManAreFree(p)
}

func TestCubeReachabilityListContainment(t *testing.T) {
	g := Gia_ManStart(5)
	r := Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, r)
	Gia_ManSetRegNum(g, 1)
	p := Gia_ManAreCreate(g)
	first := Gia_ManAreCreateSta(p)
	Gia_StaSetValue0(first, 0)
	Gia_ManAreCubeAddToList(p, &p.Root, first)
	duplicate := Gia_ManAreCreateSta(p)
	Gia_StaSetValue0(duplicate, 0)
	if Gia_ManAreCubeCheckList(p, &p.Root, duplicate) != 0 || p.nStas != 2 {
		t.Fatal("contained cube was not discarded")
	}
	broader := Gia_ManAreCreateSta(p)
	if Gia_ManAreCubeCheckList(p, &p.Root, broader) != 1 || Gia_StaIsUnused(first) == 0 {
		t.Fatal("contained existing cube not marked unused")
	}
	Gia_ManAreCubeAddToList(p, &p.Root, broader)
	if Gia_ManAreListCountUsed(p) != 1 {
		t.Fatal("incorrect used count")
	}
	Gia_ManAreCompress(p, &p.Root)
	if Gia_Ptr2Int(Gia_ManAreSta(p, p.Root).iNext) != 0 {
		t.Fatal("compression retained unused cube")
	}
	Gia_ManAreFree(p)
}

func TestCubeReachabilityTreeRebalance(t *testing.T) {
	g := Gia_ManStart(20)
	regs := make([]int32, 6)
	for i := range regs {
		regs[i] = Gia_ManAppendCi(g)
	}
	for _, r := range regs {
		Gia_ManAppendCo(g, r)
	}
	Gia_ManSetRegNum(g, 6)
	p := Gia_ManAreCreate(g)
	for assignment := 0; assignment < 64; assignment++ {
		s := Gia_ManAreCreateSta(p)
		for i := 0; i < 6; i++ {
			if assignment&(1<<i) == 0 {
				Gia_StaSetValue0(s, int32(i))
			} else {
				Gia_StaSetValue1(s, int32(i))
			}
		}
		if Gia_ManAreCubeProcess(p, s) != 1 {
			t.Fatalf("state %d unexpectedly rejected", assignment)
		}
	}
	if p.fTree == 0 || Gia_ManAreListCountUsed(p) != 64 || Gia_ManCountMinterms(p) != 64 {
		t.Fatalf("tree=%d cubes=%d minterms=%d", p.fTree, Gia_ManAreListCountUsed(p), Gia_ManCountMinterms(p))
	}
	duplicate := Gia_ManAreCreateSta(p)
	for i := int32(0); i < 6; i++ {
		Gia_StaSetValue0(duplicate, i)
	}
	if Gia_ManAreCubeProcess(p, duplicate) != 0 {
		t.Fatal("duplicate accepted")
	}
	Gia_ManAreFree(p)
}

func TestCubeReachabilitySupport(t *testing.T) {
	g := Gia_ManStart(10)
	a, r0, r1 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, r0)
	Gia_ManAppendCo(g, r1)
	Gia_ManSetRegNum(g, 2)
	p := Gia_ManAreCreate(g)
	p.pNew = Gia_ManStart(8)
	x, y := Gia_ManAppendCi(p.pNew), Gia_ManAppendCi(p.pNew)
	root := Gia_ManAppendAnd(p.pNew, x, y)
	Gia_ObjSetValue(Gia_ManPo(g, 0), root)
	Gia_ObjSetValue(Gia_ManRi(g, 0), root)
	Gia_ObjSetValue(Gia_ManRi(g, 1), root)
	if Gia_ManCheckPOs(p) != 2 || Gia_ManCheckPOstatus(p) != 1 || Gia_ManAreMostUsedPi(p) != Gia_ManCi(p.pNew, 0) {
		t.Fatal("support analysis incorrect")
	}
	Gia_ObjSetValue(Gia_ManPo(g, 0), 0)
	if Gia_ManCheckPOs(p) != 0 || Gia_ManCheckPOstatus(p) != 0 {
		t.Fatal("constant output status incorrect")
	}
	Gia_ManAreFree(p)
}

func TestCubeReachabilitySuccessorSplitting(t *testing.T) {
	g := Gia_ManStart(12)
	a, r0 := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, r0)
	Gia_ManAppendCo(g, a)
	Gia_ManAppendCo(g, a)
	Gia_ManSetRegNum(g, 2)
	p := Gia_ManAreCreate(g)
	p.iStaCur = 1
	if Gia_ManAreCubeProcess(p, Gia_ManAreCreateStaInit(p)) != 1 {
		t.Fatal("initial cube rejected")
	}
	if Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(1)) != 0 || p.nStas != 3 {
		t.Fatalf("successors: states=%d", p.nStas)
	}
	next := Gia_ManAreStaInt(p, 2)
	if Gia_StaHasValue1(next, 0) == 0 || Gia_StaHasValue1(next, 1) == 0 || Gia_Ptr2Int(next.iPrev) != 1 {
		t.Fatal("correlated inputs did not yield 11 successor")
	}
	Gia_ManAreFree(p)
}

func TestCubeReachabilityDriverAndCounterexample(t *testing.T) {
	g := Gia_ManStart(8)
	input, reg := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	Gia_ManAppendCo(g, reg)
	Gia_ManAppendCo(g, input)
	Gia_ManSetRegNum(g, 1)
	if got := Gia_ManArePerform(g, 100, 1, 0); got != 0 {
		t.Fatalf("expected reachable failure, got %d", got)
	}
	if g.pCexSeq == nil || Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 || g.pCexSeq.IFrame != 1 {
		t.Fatalf("invalid counterexample: %+v", g.pCexSeq)
	}
	if got := Gia_ManArePerform(g, 100, 0, 0); got != 1 {
		t.Fatalf("expected exhaustive reachability, got %d", got)
	}
	if got := Gia_ManArePerform(g, 1, 0, 0); got != -1 {
		t.Fatalf("expected state limit, got %d", got)
	}
}
