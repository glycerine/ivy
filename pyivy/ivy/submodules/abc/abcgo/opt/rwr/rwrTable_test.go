package rwr

import "testing"

func TestTruthTablePageBoundariesAndMissingReads(t *testing.T) {
	var table rwrTruthTable
	for key := int32(0); key < 65536; key++ {
		if table.head(key) != nil {
			t.Fatalf("empty table returned a chain at %04x", key)
		}
	}
	for _, page := range table {
		if page != nil {
			t.Fatal("reading a missing chain allocated a page")
		}
	}
	for _, key := range []uint16{0, rwrPageSize - 1, rwrPageSize, rwrPageSize + 1, 65535} {
		first, second := &Rwr_Node_t{Id: int32(key)}, &Rwr_Node_t{Id: int32(key) + 1}
		slot := table.slot(key)
		if *slot != nil {
			t.Fatalf("new key %04x aliases another chain", key)
		}
		rwrAddTail(slot, first)
		rwrAddTail(table.slot(key), second)
		if table.head(int32(key)) != first || first.pNext != second || second.pNext != nil {
			t.Fatalf("chain at %04x lost insertion order", key)
		}
	}
}

func TestBushLookupRejectsOutOfRangeTruths(t *testing.T) {
	p := &Rwr_Man_t{pTable: new(rwrTruthTable)}
	for _, key := range []int32{-1, -64, 65536} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("accepted invalid truth index %d", key)
				}
			}()
			Rwr_GetBushSumOfVolumes(p, key)
		}()
	}
}

func TestStoppedManagerInvalidatesBushTable(t *testing.T) {
	p := Rwr_ManStart(0)
	Rwr_ManStop(p)
	defer func() {
		if recover() == nil {
			t.Fatal("read an uninitialized table from a stopped manager")
		}
	}()
	Rwr_GetBushSumOfVolumes(p, 0)
}
