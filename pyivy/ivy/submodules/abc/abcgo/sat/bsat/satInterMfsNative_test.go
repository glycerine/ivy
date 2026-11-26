package bsat

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

// loadDumpedStore reads a clause store written by Sto_ManDumpClauses
// ("p nVars nClauses nRoots nClausesA" followed by DIMACS clauses) and
// re-creates the A / root marks the way the solver set them: the first
// nClausesA clauses are A clauses and the first nRoots clauses are roots.
func loadDumpedStore(t *testing.T, name string) *Sto_Man_t {
	t.Helper()
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	var nVars, nClauses, nRoots, nA int32
	if _, err := fmt.Fscanf(r, "p %d %d %d %d\n", &nVars, &nClauses, &nRoots, &nA); err != nil {
		t.Fatal(err)
	}
	p := Sto_ManAlloc()
	var lits []lit
	added := int32(0)
	for {
		var x int32
		if _, err := fmt.Fscan(r, &x); err != nil {
			break
		}
		if x != 0 {
			if x > 0 {
				lits = append(lits, toLit(x-1))
			} else {
				lits = append(lits, lit_neg(toLit(-x-1)))
			}
			continue
		}
		if Sto_ManAddClause(p, lits) == 0 {
			t.Fatal("bad clause")
		}
		lits = lits[:0]
		added++
		if added == nA {
			Sto_ManMarkClausesA(p)
		}
		if added == nRoots {
			Sto_ManMarkRoots(p)
		}
	}
	if added != nClauses || p.nVars != nVars {
		t.Fatalf("read %d clauses / %d vars, header says %d / %d", added, p.nVars, nClauses, nVars)
	}
	return p
}

// The store below was dumped by native ABC's mfs (sort16, if -K 6; mfs) for
// node 72; native Int_ManInterpolate returns truth f3710000 for it (checked
// with a driver linked against libabc.a).
func TestIntManInterpolateMatchesNativeOnMfsStore(t *testing.T) {
	p := loadDumpedStore(t, "test_vectors/mfs_sort16_interp.cnf")
	m := Int_ManAlloc()
	glo := Int_ManSetGlobalVars(m, 5)
	copy(glo, []int32{76, 77, 78, 80, 81})
	var truth []uint32
	n := Int_ManInterpolate(m, p, 0, &truth)
	if n != 5 {
		t.Fatalf("n = %d, want 5", n)
	}
	if truth[0] != 0xf3710000 {
		t.Fatalf("truth = %08x, want f3710000", truth[0])
	}
}
