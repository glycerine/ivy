package reo

import (
	"os"
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

// truthToBdd is Kit_TruthToBdd( dd, pTruth, nVars, 0 ).
func truthToBdd(dd *DdManager, truth []uint32, n int32) *DdNode {
	var build func(int32, int32) *DdNode
	build = func(bit, vars int32) *DdNode {
		if vars <= 5 {
			mask := ^uint32(0) >> (32 - (1 << vars))
			t := (truth[bit>>5] >> (bit & 31)) & mask
			if t == 0 {
				return dd.LogicZero
			}
			if t == mask {
				return dd.One
			}
		}
		lo := build(bit, vars-1)
		cudd.Cudd_Ref(lo)
		hi := build(bit+(1<<uint(vars-1)), vars-1)
		cudd.Cudd_Ref(hi)
		f := cudd.Cudd_bddIte(dd, dd.Vars[vars-1], hi, lo)
		cudd.Cudd_Ref(f)
		cudd.Cudd_RecursiveDeref(dd, lo)
		cudd.Cudd_RecursiveDeref(dd, hi)
		cudd.Cudd_Deref(f)
		return f
	}
	return build(0, n)
}

// TestReoMatchesNativeEngine runs the REO engine on 300 random 7-input
// functions per mode and compares the DAG sizes and the returned variable
// orders with the native engine. The expected values were printed by a C
// program linked with libabc.a that performs the same steps:
// Kit_TruthToBdd, (Cudd_bddPermute, Cudd_BddToAdd,) Extra_ReorderArray.
// Modes: 0 nodes, 1 APL, 2 width on ADDs, 3 nodes with remapping and two
// iterations on functions of variables 3..9, 4 nodes on three functions.
func TestReoMatchesNativeEngine(t *testing.T) {
	want := []struct{ tot, osum int64 }{
		{10264, -3456353577016438086},
		{10610, 294684903684988934},
		{12203, -3430453439699815634},
		{10159, -212771436770884930},
		{64248, 6721219767245419694},
	}
	// width mode prints the initial profile of every call
	stdout := os.Stdout
	if null, err := os.Open(os.DevNull); err == nil {
		os.Stdout = null
		defer func() { os.Stdout = stdout; null.Close() }()
	}
	const N, nv = 300, 7
	for mode := 0; mode < 5; mode++ {
		var s uint32 = 777
		rnd := func() uint32 { s = s*1103515245 + 12345; return s }
		dd := cudd.NewBDDManager(nv + 3)
		p := Extra_ReorderInit(nv+3, 100)
		switch mode {
		case 1:
			Extra_ReorderSetMinimizationType(p, REO_MINIMIZE_APL)
		case 2:
			Extra_ReorderSetMinimizationType(p, REO_MINIMIZE_WIDTH)
		case 3:
			Extra_ReorderSetRemapping(p, 1)
			Extra_ReorderSetIterations(p, 2)
		}
		var tot, osum int64
		truth := make([]uint32, 4)
		order := make([]int, 16)
		for i := 0; i < N; i++ {
			nf := 1
			if mode == 4 {
				nf = 3
			}
			f, g := make([]*DdNode, nf), make([]*DdNode, nf)
			for j := 0; j < nf; j++ {
				for k := range truth {
					truth[k] = rnd() & (rnd() | rnd())
				}
				f[j] = truthToBdd(dd, truth, nv)
				cudd.Cudd_Ref(f[j])
				if mode == 3 {
					h := cudd.Cudd_bddPermute(dd, f[j], []int{3, 4, 5, 6, 7, 8, 9, 0, 1, 2})
					cudd.Cudd_Ref(h)
					cudd.Cudd_RecursiveDeref(dd, f[j])
					f[j] = h
				}
				if mode == 2 {
					h := cudd.Cudd_BddToAdd(dd, f[j])
					cudd.Cudd_Ref(h)
					cudd.Cudd_RecursiveDeref(dd, f[j])
					f[j] = h
				}
			}
			for k := range order {
				order[k] = -1
			}
			Extra_ReorderArray(p, dd, f, g, nf, order)
			for j := 0; j < nf; j++ {
				tot += int64(cudd.Cudd_DagSize(g[j]) * (j + 1))
				// the reordered function permuted by order is the original one
				var back *DdNode
				perm := identity(nv + 3)
				for v := range perm {
					if order[v] >= 0 {
						perm[v] = order[v]
					}
				}
				if mode == 2 {
					back = cudd.Cudd_addPermute(dd, g[j], perm)
				} else {
					back = cudd.Cudd_bddPermute(dd, g[j], perm)
				}
				if back != f[j] {
					t.Fatalf("mode %d function %d: permuting the result by the order does not give the original", mode, i)
				}
			}
			for k := 0; k < nv+3; k++ {
				osum = osum*31 + int64(order[k]) + 7
			}
			for j := 0; j < nf; j++ {
				cudd.Cudd_RecursiveDeref(dd, g[j])
				cudd.Cudd_RecursiveDeref(dd, f[j])
			}
		}
		Extra_ReorderQuit(p)
		if tot != want[mode].tot || osum != want[mode].osum {
			t.Errorf("mode %d: tot=%d osum=%d, native tot=%d osum=%d", mode, tot, osum, want[mode].tot, want[mode].osum)
		}
	}
}

// TestReoVerificationAndShuffle runs the engine with verification on and
// the shuffle test, which print a message on failure.
func TestReoVerificationAndShuffle(t *testing.T) {
	dd := cudd.NewBDDManager(8)
	var s uint32 = 99
	rnd := func() uint32 { s = s*1103515245 + 12345; return s }
	p := Extra_ReorderInit(8, 100)
	Extra_ReorderSetVerification(p, 1)
	truth := make([]uint32, 8)
	for i := 0; i < 50; i++ {
		for k := range truth {
			truth[k] = rnd() & rnd()
		}
		f := truthToBdd(dd, truth, 8)
		cudd.Cudd_Ref(f)
		out := captureStdout(t, func() {
			g := Extra_Reorder(p, dd, f, nil)
			cudd.Cudd_Ref(g)
			cudd.Cudd_RecursiveDeref(dd, g)
			Extra_ShuffleTest(p, dd, f)
		})
		if out != "REO: Internal verification is okay!\n" {
			t.Fatalf("function %d: %q", i, out)
		}
		cudd.Cudd_RecursiveDeref(dd, f)
	}
	Extra_ReorderQuit(p)
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	f()
	os.Stdout = stdout
	w.Close()
	buf := make([]byte, 1<<16)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}
