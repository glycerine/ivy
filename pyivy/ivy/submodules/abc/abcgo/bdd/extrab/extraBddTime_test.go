package extrab

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"testing"
	"time"
)

func TestTimedOperatorsUnlimitedAndExpired(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	dst := cudd.NewBDDManager(3)
	cube := dd.Vars[1]
	past := time.Now().Add(-time.Second)
	future := time.Now().Add(time.Hour)
	for a := 0; a < 256; a++ {
		f := autoTruth(dd, uint64(a), 3)
		g := autoTruth(dd, uint64(a^173), 3)
		want := cudd.Cudd_bddAnd(dd, f, g)
		for _, deadline := range []time.Time{{}, future} {
			if got := Extra_bddAndTime(dd, f, g, deadline); got != want {
				t.Fatal("and", a)
			}
			got := Extra_bddAndAbstractTime(dd, f, g, cube, deadline)
			if got != cudd.Cudd_bddExistAbstract(dd, want, cube) {
				t.Fatal("abstract", a)
			}
			if got := Extra_TransferPermuteTime(dd, dst, f, []int{2, 0, 1}, deadline); got != dd.Transfer(f, dst, []int{2, 0, 1}) {
				t.Fatal("transfer", a)
			}
		}
		if Extra_bddAndTime(dd, f, g, past) != nil || Extra_bddAndAbstractTime(dd, f, g, cube, past) != nil || Extra_TransferPermuteTime(dd, dst, f, nil, past) != nil {
			t.Fatal("expired", a)
		}
	}
	if Extra_bddAndAbstractTime(dd, dd.One, dd.One, cudd.Cudd_Not(cube), time.Time{}) != nil {
		t.Fatal("negative cube")
	}
	if Extra_TransferPermuteTime(dd, dst, dd.Vars[2], []int{0, 1, -1}, time.Time{}) != nil {
		t.Fatal("negative target")
	}
}
func TestOriginalCTimedUnlimitedFixture(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	dst := cudd.NewBDDManager(3)
	functions := make([]*DdNode, 256)
	for i := range functions {
		functions[i] = autoTruth(dd, uint64(i), 3)
	}
	hash := uint64(14695981039346656037)
	put := func(f *DdNode, manager *DdManager) { hash ^= autoTruthBits(t, manager, f, 3); hash *= 1099511628211 }
	for _, f := range functions {
		put(Extra_TransferPermuteTime(dd, dst, f, []int{2, 0, 1}, time.Time{}), dst)
		for _, g := range functions {
			put(Extra_bddAndTime(dd, f, g, time.Time{}), dd)
			put(Extra_bddAndAbstractTime(dd, f, g, dd.Vars[1], time.Time{}), dd)
		}
	}
	if hash != 0xf6ee7ba969cdfbe5 {
		t.Fatalf("original C unlimited hash %016x", hash)
	}
}
