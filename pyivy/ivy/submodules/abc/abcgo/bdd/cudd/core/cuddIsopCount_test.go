package core

import "testing"

func TestBoundedISOPCountsMatchCovers(t *testing.T) {
	for table := 0; table < 256; table++ {
		dd := NewBDDManager(3)
		f := bddFromTruth(dd, uint8(table), 3)
		beforeZ := dd.ZddSize
		chosen, count := CountISOPBounded(dd, f, f, 100, 1000000)
		if chosen != f || count < 0 || dd.ZddSize != beforeZ {
			t.Fatal("bounded count", table, count)
		}
		var cover *DdNode
		Cudd_zddIsop(dd, f, f, &cover)
		if Cudd_zddCount(dd, cover) != count {
			t.Fatal("cover count", table, count)
		}
		if count > 0 {
			before, next := len(dd.unique), dd.nextID
			if limited, n := CountISOPBounded(dd, f, f, count-1, 1000000); limited != nil || n != -1 || len(dd.unique) != before || dd.nextID != next {
				t.Fatal("cube limit/rollback", table, n)
			}
		}
	}
}
func TestBoundedISOPNodeBudgetAndStateRestoration(t *testing.T) {
	found := false
	for table := 1; table < 255; table++ {
		dd := NewBDDManager(3)
		f := bddFromTruth(dd, uint8(table), 3)
		before, next := len(dd.unique), dd.nextID
		result, count := CountISOPBounded(dd, f, f, 100, 0)
		if len(dd.unique) != before || dd.allocationLimited {
			t.Fatal("node budget state")
		}
		if result == nil {
			found = true
			if count != -1 || dd.nextID != next {
				t.Fatal("node limit rollback")
			}
			if r, n := CountISOPBounded(dd, f, f, 100, 10000); r != f || n < 0 {
				t.Fatal("retry after limit")
			}
			break
		}
	}
	if !found {
		t.Fatal("fixture never required new intermediate nodes")
	}
	dd := NewBDDManager(1)
	if r, n := CountISOPBounded(dd, dd.One, dd.One, 0, 10); r != nil || n != -1 {
		t.Fatal("constant cube limit")
	}
	if r, n := CountISOPBounded(dd, dd.One, dd.LogicZero, 10, 10); r != nil || n != -1 {
		t.Fatal("invalid interval")
	}
}
