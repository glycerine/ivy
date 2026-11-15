package gia

import (
	"testing"
	"unsafe"
)

func TestCubeReachabilityHandleEncoding(t *testing.T) {
	var handle Gia_PtrAre_t
	t.Logf("handle=%d cube=%d tree=%d bytes", unsafe.Sizeof(handle), unsafe.Sizeof(Gia_StaAre_t{}), unsafe.Sizeof(Gia_ObjAre_t{}))
	check := func(n uint32) {
		ptr := Gia_Int2Ptr(n)
		if got := Gia_Ptr2Int(ptr); got != n&0x7fffffff {
			t.Fatalf("handle%x index%x", n, got)
		}
		s := Gia_StaAre_t{iPrev: ptr, iNext: Gia_Int2Ptr(n ^ 0x12345678)}
		if got := Gia_StaIsUnused(&s); got != int32(n>>31) {
			t.Fatalf("handle%x unused%d", n, got)
		}
		if got := Gia_StaIsUsed(&s); got != int32(1-(n>>31)) {
			t.Fatalf("handle%x used%d", n, got)
		}
		next := s.iNext
		Gia_StaSetUnused(&s)
		if Gia_StaIsUnused(&s) != 1 || Gia_StaIsUsed(&s) != 0 || Gia_Ptr2Int(s.iPrev) != n&0x7fffffff || s.iNext != next {
			t.Fatalf("mark changed handle index/next for%x", n)
		}
	}
	for _, n := range []uint32{0, 1, 0xfffff, 0x100000, 0x7fffffff, 0x80000000, 0x800fffff, 0xffffffff} {
		check(n)
	}
	n := uint32(0x12345678)
	for i := 0; i < 65536; i++ {
		n = n*1664525 + 1013904223
		check(n)
	}
}

func TestCubeReachabilityNestedTreeMembership(t *testing.T) {
	g := Gia_ManStart(32)
	r := make([]int32, 8)
	for i := range r {
		r[i] = Gia_ManAppendCi(g)
	}
	for _, lit := range r {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, 8)
	p := Gia_ManAreCreate(g)
	defer Gia_ManAreFree(p)
	makeState := func(mask int) *Gia_StaAre_t {
		s := Gia_ManAreCreateSta(p)
		for i := int32(0); i < 8; i++ {
			if mask&(1<<i) != 0 {
				Gia_StaSetValue1(s, i)
			} else {
				Gia_StaSetValue0(s, i)
			}
		}
		return s
	}
	// Full minterms are mutually disjoint: all must survive every rebalance.
	// 256 rows force nested branches through the 63-entry split threshold.
	for mask := 0; mask < 256; mask++ {
		if Gia_ManAreCubeProcess(p, makeState(mask)) != 1 {
			t.Fatalf("minterm%d discarded", mask)
		}
		if got := Gia_ManCountMinterms(p); got != int32(mask+1) {
			t.Fatalf("after%d union size%d", mask, got)
		}
	}
	if p.fTree != 1 || p.nObjs < 4 || Gia_ManAreListCountUsed(p) != 256 {
		t.Fatalf("tree%d nodes%d count%d", p.fTree, p.nObjs, Gia_ManAreListCountUsed(p))
	}
	count := p.nStas
	for mask := 255; mask >= 0; mask-- {
		if Gia_ManAreCubeProcess(p, makeState(mask)) != 0 || p.nStas != count {
			t.Fatalf("duplicate%d lost in nested tree", mask)
		}
	}
	if Gia_ManCountMinterms(p) != 256 {
		t.Fatal("duplicates changed covered state set")
	}
}

func TestCubeReachabilityHighestTreeVariable(t *testing.T) {
	g := Gia_ManStart(2*MAX_VARS_NUM + 1)
	regs := make([]int32, MAX_VARS_NUM)
	for i := range regs {
		regs[i] = Gia_ManAppendCi(g)
	}
	for _, lit := range regs {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, MAX_VARS_NUM)
	// Only list/tree storage is needed: avoid constructing unused all-CI TFO
	// tables while testing the supported highest variable in a real interface.
	p := &Gia_ManAre_t{pAig: g, nWords: (2*MAX_VARS_NUM + 31) / 32}
	defer Gia_ManAreFree(p)
	a := Gia_ManAreCreateSta(p)
	Gia_StaSetValue0(a, MAX_VARS_NUM-1)
	Gia_ManAreCubeAddToList(p, &p.Root, a)
	b := Gia_ManAreCreateSta(p)
	Gia_StaSetValue1(b, MAX_VARS_NUM-1)
	Gia_ManAreCubeAddToList(p, &p.Root, b)
	Gia_ManAreRebalance(p, &p.Root)
	root := Gia_ManAreObj(p, p.Root)
	if int32(root.iVar) != MAX_VARS_NUM-1 || root.nStas0 != 1 || root.nStas1 != 1 {
		t.Fatalf("highest-variable partition: %+v", root)
	}
	query := Gia_ManAreCreateSta(p)
	Gia_StaSetValue1(query, MAX_VARS_NUM-1)
	if Gia_ManAreCubeCheckTree_rec(p, root, query) != 0 {
		t.Fatal("highest variable lookup lost matching cube")
	}
}

func TestCubeReachabilityPageCountRange(t *testing.T) {
	for _, count := range []int32{0, 1, MAX_ITEM_NUM - 1, MAX_ITEM_NUM, MAX_ITEM_NUM + 1, 2 * MAX_ITEM_NUM, 2047 * MAX_ITEM_NUM, 2047*MAX_ITEM_NUM + 1, 1<<31 - 2, 1<<31 - 1} {
		want := (int64(count) + MAX_ITEM_NUM - 1) / MAX_ITEM_NUM
		if got := giaArePages(count); int64(got) != want {
			t.Errorf("count%d pages%d want%d", count, got, want)
		}
	}
}

func TestCubeReachabilityStorageExhaustionDoesNotAllocate(t *testing.T) {
	// Simulate the capacity counters without allocating two billion records.
	// Exhaustion must be checked before any array access/allocation or overflow.
	p := &Gia_ManAre_t{nObjs: 1<<31 - 1, nStas: 1<<31 - 1}
	hashReviewStdout(t, func() {
		if Gia_ManAreCreateObj(p) != nil || Gia_ManAreCreateSta(p) != nil || Gia_ManAreCreateStaNew(p) != nil {
			t.Fatal("capacity exhaustion returned a record")
		}
	})
	if p.nObjs != 1<<31-1 || p.nStas != 1<<31-1 || len(p.objects) != 0 || len(p.states) != 0 {
		t.Fatal("capacity stop allocated or changed counters")
	}
}
