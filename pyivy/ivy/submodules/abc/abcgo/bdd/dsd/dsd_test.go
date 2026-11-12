package dsd

import (
	"fmt"
	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
	"math/rand"
	"strings"
	"testing"
)

func tableBDD(dd *DdManager, table uint64, n int) *DdNode {
	var rec func(int, uint64) *DdNode
	rec = func(i int, a uint64) *DdNode {
		if i == n {
			if table>>a&1 != 0 {
				return dd.One
			}
			return dd.LogicZero
		}
		return bdd.Cudd_bddIte(dd, dd.Vars[i], rec(i+1, a|1<<uint(i)), rec(i+1, a))
	}
	return rec(0, 0)
}
func evalBDD(dd *DdManager, f *DdNode, assignment uint64) bool {
	for !bdd.Cudd_IsConstant(f) {
		r := bdd.Cudd_Regular(f)
		child := r.E
		if assignment>>r.Index&1 != 0 {
			child = r.T
		}
		f = bdd.Cudd_NotCond(child, bdd.Cudd_IsComplement(f))
	}
	return f == dd.One
}
func reconstruct(dd *DdManager, p *Dsd_Node_t, memo map[*Dsd_Node_t]*DdNode) *DdNode {
	r := Dsd_Regular(p)
	f := memo[r]
	if f == nil {
		switch r.Type {
		case DSD_NODE_CONST1:
			f = dd.One
		case DSD_NODE_BUF:
			f = r.G
		case DSD_NODE_OR, DSD_NODE_EXOR:
			f = dd.LogicZero
			for _, c := range r.pDecs {
				g := reconstruct(dd, c, memo)
				if r.Type == DSD_NODE_OR {
					f = bdd.Cudd_bddOr(dd, f, g)
				} else {
					f = bdd.Cudd_bddXor(dd, f, g)
				}
			}
		case DSD_NODE_PRIME:
			f = Dsd_TreeGetPrimeFunction(dd, r) // simultaneous substitution through recursive ITE
			subst := map[uint32]*DdNode{}
			for i, c := range r.pDecs {
				subst[uint32(dd.InvPerm[i])] = reconstruct(dd, c, memo)
			}
			cache := map[*DdNode]*DdNode{}
			var compose func(*DdNode) *DdNode
			compose = func(n *DdNode) *DdNode {
				if bdd.Cudd_IsConstant(n) {
					return n
				}
				if v := cache[n]; v != nil {
					return v
				}
				rr := bdd.Cudd_Regular(n)
				v := bdd.Cudd_bddIte(dd, subst[rr.Index], compose(rr.T), compose(rr.E))
				v = bdd.Cudd_NotCond(v, bdd.Cudd_IsComplement(n))
				cache[n] = v
				return v
			}
			f = compose(f)
		}
		memo[r] = f
	}
	return bdd.Cudd_NotCond(f, Dsd_IsComplement(p))
}
func verifyTree(t *testing.T, m *Dsd_Manager_t, p *Dsd_Node_t, f *DdNode) {
	t.Helper()
	if got := reconstruct(m.dd, p, map[*Dsd_Node_t]*DdNode{}); got != f {
		t.Fatalf("reconstructed function differs: input=%d result=%d", f.ID, got.ID)
	}
	var visit func(*Dsd_Node_t)
	seen := map[*Dsd_Node_t]bool{}
	visit = func(p *Dsd_Node_t) {
		r := Dsd_Regular(p)
		if seen[r] {
			return
		}
		seen[r] = true
		if r.S != bdd.Cudd_Support(m.dd, r.G) {
			t.Fatal("incorrect support")
		}
		if r.Type == DSD_NODE_BUF || r.Type == DSD_NODE_CONST1 {
			return
		}
		union := m.dd.One
		for _, c := range r.pDecs {
			cr := Dsd_Regular(c)
			if suppOverlap(m.dd, union, cr.S) {
				t.Fatal("overlapping decomposition support")
			}
			union = bdd.Cudd_bddAnd(m.dd, union, cr.S)
			if r.Type != DSD_NODE_OR && Dsd_IsComplement(c) {
				t.Fatal("complemented non-OR child")
			}
			visit(c)
		}
		if union != r.S {
			t.Fatal("incomplete support partition")
		}
	}
	visit(p)
}
func TestAllThreeVariableFunctions(t *testing.T) {
	dd := bdd.NewBDDManager(3)
	m := Dsd_ManagerStart(dd, 3, 0)
	defer Dsd_ManagerStop(m)
	for mask := uint64(0); mask < 256; mask++ {
		f := tableBDD(dd, mask, 3)
		p := Dsd_DecomposeOne(m, f)
		verifyTree(t, m, p, f)
		if Dsd_DecomposeOne(m, bdd.Cudd_Not(f)) != Dsd_Not(p) {
			t.Fatal("complement cache")
		}
	}
}
func TestRandomFunctionsAndReordering(t *testing.T) {
	rng := rand.New(rand.NewSource(43))
	for _, order := range [][]int{{0, 1, 2, 3, 4, 5}, {5, 2, 4, 1, 3, 0}} {
		dd := bdd.NewBDDManager(6)
		if bdd.Cudd_ShuffleHeap(dd, order) == 0 {
			t.Fatal("reorder")
		}
		m := Dsd_ManagerStart(dd, 6, 0)
		for i := 0; i < 180; i++ {
			table := rng.Uint64()
			f := tableBDD(dd, table, 6)
			p := Dsd_DecomposeOne(m, f)
			verifyTree(t, m, p, f)
		}
		Dsd_ManagerStop(m)
	}
}
func TestStructuredDecompositionAndTreeAPI(t *testing.T) {
	dd := bdd.NewBDDManager(7)
	m := Dsd_ManagerStart(dd, 7, 0)
	defer Dsd_ManagerStop(m)
	v := dd.Vars
	or := bdd.Cudd_bddOr(dd, v[0], bdd.Cudd_bddOr(dd, v[1], v[2]))
	xor := bdd.Cudd_bddXor(dd, v[3], v[4])
	mux := bdd.Cudd_bddIte(dd, v[5], or, xor)
	f := bdd.Cudd_bddAnd(dd, mux, v[6])
	Dsd_Decompose(m, []*DdNode{f, mux, or, xor, dd.One, bdd.Cudd_Not(or)}, 6)
	for i, g := range []*DdNode{f, mux, or, xor, dd.One, bdd.Cudd_Not(or)} {
		verifyTree(t, m, m.roots[i], g)
	}
	if Dsd_NodeReadType(m.roots[2]) != DSD_NODE_OR || Dsd_NodeReadDecsNum(m.roots[2]) != 3 {
		t.Fatal("OR flattening")
	}
	if Dsd_NodeReadType(m.roots[3]) != DSD_NODE_EXOR {
		t.Fatal("XOR")
	}
	if Dsd_TreeCountPrimeNodesOne(m.roots[0]) != 1 || Dsd_TreeGetAigCost(m.roots[0]) != 9 {
		t.Fatal("tree cost")
	}
	n := 0
	nodes := Dsd_TreeCollectNodesDfs(m, &n)
	if n != Dsd_TreeCountNonTerminalNodes(m) {
		t.Fatal("DFS count")
	}
	position := map[*Dsd_Node_t]int{}
	for i, p := range nodes {
		position[p] = i
		Dsd_NodeSetMark(p, 20+i)
		for _, c := range p.pDecs {
			r := Dsd_Regular(c)
			if len(r.pDecs) > 1 && position[r] >= i {
				t.Fatal("DFS order")
			}
		}
	}
	Dsd_TreeUnmark(m)
	for _, p := range nodes {
		if p.Mark != 20+position[p] {
			t.Fatal("user mark changed")
		}
	}
	depth, gate := 0, 0
	Dsd_TreeNodeGetInfo(m, &depth, &gate)
	if depth < 2 || gate != 3 {
		t.Fatal("tree information")
	}
	var out strings.Builder
	Dsd_TreePrint(&out, m, nil, nil, 1, -1)
	Dsd_TreePrint2(&out, m, nil, nil, -1)
	Dsd_NodePrint(&out, m.roots[0])
	if !strings.Contains(out.String(), "PRIME(") || !strings.Contains(out.String(), "EXOR(") {
		t.Fatal(out.String())
	}
}

func TestTreeQueriesPreserveUserMarks(t *testing.T) {
	dd := bdd.NewBDDManager(3)
	m := Dsd_ManagerStart(dd, 3, 0)
	defer Dsd_ManagerStop(m)
	p := Dsd_DecomposeOne(m, tableBDD(dd, 0xe8, 3))
	m.roots = []*Dsd_Node_t{p}
	Dsd_NodeSetMark(p, 42)
	Dsd_TreeUnmark(m)
	Dsd_TreeCountPrimeNodes(m)
	n := 0
	Dsd_TreeCollectNodesDfs(m, &n)
	if Dsd_NodeReadMark(p) != 42 {
		t.Fatal("tree queries overwrote the user mark")
	}
	Dsd_TreeNodeGetInfoOne(p, nil, nil)
}

func TestOverlapRefinementRegression(t *testing.T) {
	dd := bdd.NewBDDManager(4)
	m := Dsd_ManagerStart(dd, 4, 0)
	defer Dsd_ManagerStop(m)
	f := tableBDD(dd, 0xf5ce, 4)
	p := Dsd_DecomposeOne(m, f)
	verifyTree(t, m, p, f)
}

func structuredTable(rng *rand.Rand, n int) uint64 {
	outer, g, h := rng.Uint64(), rng.Uint64(), rng.Uint64()
	var table uint64
	for a := 0; a < 1<<uint(n); a++ {
		left := a&1 | (a>>1)&2
		right := (a >> 1) & 1
		if n >= 5 {
			left |= (a >> 2) & 4
		}
		if n >= 6 {
			right |= (a >> 3) & 2
		}
		b := (g>>uint(left))&1 | ((h>>uint(right))&1)<<1 | uint64((a>>3)&1)<<2
		if outer>>b&1 != 0 {
			table |= 1 << uint(a)
		}
	}
	return table
}
func TestStructuredCofactorCases(t *testing.T) {
	rng := rand.New(rand.NewSource(612))
	for _, order := range [][]int{{0, 1, 2, 3, 4, 5}, {5, 4, 3, 2, 1, 0}} {
		dd := bdd.NewBDDManager(6)
		bdd.Cudd_ShuffleHeap(dd, order)
		m := Dsd_ManagerStart(dd, 6, 0)
		for i := 0; i < 600; i++ {
			f := tableBDD(dd, structuredTable(rng, 6), 6)
			p := Dsd_DecomposeOne(m, f)
			verifyTree(t, m, p, f)
		}
		Dsd_ManagerStop(m)
	}
}
func TestLargeDisjointFunctions(t *testing.T) {
	dd := bdd.NewBDDManager(80)
	m := Dsd_ManagerStart(dd, 80, 0)
	defer Dsd_ManagerStop(m)
	f := dd.LogicZero
	for i := 0; i < 80; i += 5 {
		a := bdd.Cudd_bddOr(dd, dd.Vars[i], dd.Vars[i+1])
		b := bdd.Cudd_bddXor(dd, dd.Vars[i+2], dd.Vars[i+3])
		mux := bdd.Cudd_bddIte(dd, dd.Vars[i+4], a, b)
		f = bdd.Cudd_bddXor(dd, f, mux)
	}
	p := Dsd_DecomposeOne(m, f)
	verifyTree(t, m, p, f)
	if Dsd_NodeReadType(p) != DSD_NODE_EXOR || Dsd_NodeReadDecsNum(p) != 16 || Dsd_TreeCountPrimeNodesOne(p) != 16 {
		t.Fatal("large structured decomposition")
	}
}

func TestDecomposeIgnoresUserMarkValues(t *testing.T) {
	for mark := 1; mark <= 30; mark++ {
		dd := bdd.NewBDDManager(6)
		m := Dsd_ManagerStart(dd, 6, 0)
		rng := rand.New(rand.NewSource(170))
		for i := 0; i < 50; i++ {
			for _, p := range m.table {
				Dsd_Regular(p).Mark = mark
			}
			f := tableBDD(dd, structuredTable(rng, 6), 6)
			verifyTree(t, m, Dsd_DecomposeOne(m, f), f)
		}
		Dsd_ManagerStop(m)
	}
}

func TestConditionalRootIdentityAndCacheLifecycle(t *testing.T) {
	dd := bdd.NewBDDManager(5)
	g := bdd.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
	h := bdd.Cudd_bddAnd(dd, dd.Vars[2], dd.Vars[3])
	z := dd.Vars[4]
	constant := func(x uint) *DdNode {
		if x&1 != 0 {
			return dd.One
		}
		return dd.LogicZero
	}
	outer := func(mask uint, input *DdNode) *DdNode {
		return bdd.Cudd_bddIte(dd, z, bdd.Cudd_bddIte(dd, input, constant(mask>>3), constant(mask>>2)), bdd.Cudd_bddIte(dd, input, constant(mask>>1), constant(mask)))
	}
	Dsd_CheckCacheAllocate(32)
	defer Dsd_CheckCacheDeallocate()
	for a := uint(0); a < 16; a++ {
		for b := uint(0); b < 16; b++ {
			for phase := uint(0); phase < 4; phase++ {
				x, y := uint(1)^(phase&1), uint(1)^((phase>>1)&1)
				expected := (a>>x&1) == (b>>y&1) && (a>>(2+x)&1) == (b>>(2+y)&1)
				got := Dsd_CheckRootFunctionIdentity(dd, outer(a, g), outer(b, h), bdd.Cudd_NotCond(g, phase&1 != 0), bdd.Cudd_NotCond(h, phase&2 != 0))
				if (got != 0) != expected {
					t.Fatalf("identity a=%x b=%x phase=%d", a, b, phase)
				}
			}
		}
	}
	Dsd_CheckCacheClear()
	if len(checkCache.entries) != 0 {
		t.Fatal("cache not cleared")
	}
	Dsd_CheckCacheDeallocate()
	if checkCache.entries != nil || checkCache.capacity != 0 {
		t.Fatal("cache not released")
	}
	if Dsd_CheckRootFunctionIdentity(dd, g, g, dd.One, dd.One) != 1 {
		t.Fatal("uncached identity")
	}
}
func TestLargePrimePrintingAndLegacyRemap(t *testing.T) {
	dd := bdd.NewBDDManager(7)
	order := []int{6, 5, 4, 3, 2, 1, 0}
	bdd.Cudd_ShuffleHeap(dd, order)
	type key struct{ i, k int }
	memo := map[key]*DdNode{}
	var threshold func(int, int) *DdNode
	threshold = func(i, k int) *DdNode {
		if k == 0 {
			return dd.One
		}
		if 7-i < k {
			return dd.LogicZero
		}
		key := key{i, k}
		if r := memo[key]; r != nil {
			return r
		}
		r := bdd.Cudd_bddIte(dd, dd.Vars[i], threshold(i+1, k-1), threshold(i+1, k))
		memo[key] = r
		return r
	}
	f := threshold(0, 4)
	m := Dsd_ManagerStart(dd, 7, 0)
	defer Dsd_ManagerStop(m)
	Dsd_Decompose(m, []*DdNode{f, bdd.Cudd_Not(f)}, 2)
	if Dsd_NodeReadType(m.roots[0]) != DSD_NODE_PRIME || Dsd_NodeReadDecsNum(m.roots[0]) != 7 {
		t.Fatal("expected seven-input prime")
	}
	verifyTree(t, m, m.roots[0], f)
	r := Dsd_Regular(m.roots[0])
	local := Dsd_TreeGetPrimeFunction(dd, r)
	if Dsd_TreeGetPrimeFunctionOld(dd, r, 0) != r.G {
		t.Fatal("legacy nonremapped function")
	}
	remapped := Dsd_TreeGetPrimeFunctionOld(dd, r, 1)
	for a := uint64(0); a < 128; a++ {
		reverse := uint64(0)
		for i := 0; i < 7; i++ {
			reverse |= (a >> uint(i) & 1) << uint(6-i)
		}
		if evalBDD(dd, local, reverse) != evalBDD(dd, remapped, a) {
			t.Fatal("legacy index remapping")
		}
	}
	if Dsd_TreeCollectDecomposableVars(m, nil) != 0 {
		t.Fatal("large prime marked decomposable")
	}
	var out strings.Builder
	Dsd_TreePrint2(&out, m, nil, nil, -1)
	if !strings.Contains(out.String(), "|7|{") || !strings.Contains(out.String(), "!|7|{") {
		t.Fatal(out.String())
	}
}

func TestDeclaredSupportBelowUnrelatedVariables(t *testing.T) {
	dd := bdd.NewBDDManager(12)
	order := make([]int, 12)
	for i := range order {
		order[i] = 11 - i
	}
	bdd.Cudd_ShuffleHeap(dd, order)
	m := Dsd_ManagerStart(dd, 3, 0)
	defer Dsd_ManagerStop(m)
	for table := uint64(0); table < 256; table++ {
		f := tableBDD(dd, table, 3)
		verifyTree(t, m, Dsd_DecomposeOne(m, f), f)
	}
}
func TestIndependentManagersConcurrent(t *testing.T) {
	for worker := 0; worker < 4; worker++ {
		t.Run(fmt.Sprintf("manager%d", worker), func(t *testing.T) {
			t.Parallel()
			dd := bdd.NewBDDManager(6)
			m := Dsd_ManagerStart(dd, 6, 0)
			defer Dsd_ManagerStop(m)
			rng := rand.New(rand.NewSource(int64(worker + 100)))
			for i := 0; i < 40; i++ {
				f := tableBDD(dd, structuredTable(rng, 6), 6)
				verifyTree(t, m, Dsd_DecomposeOne(m, f), f)
			}
		})
	}
}
