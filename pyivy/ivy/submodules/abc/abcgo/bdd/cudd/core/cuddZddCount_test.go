package core

import (
	"math"
	"math/big"
	"testing"
)

func TestZDDCountsExactAndOverflow(t *testing.T) {
	dd := Cudd_Init(0, 3, 256, 256, 0)
	for table := 0; table < 256; table++ {
		f := zddFromTable(dd, uint16(table))
		expected := 0
		for set := 0; set < 8; set++ {
			if table&(1<<set) != 0 {
				expected++
			}
		}
		n, err := dd.CountZDD(f)
		if err != nil || n.Int64() != int64(expected) || Cudd_zddCount(dd, f) != expected || Cudd_zddCountDouble(dd, f) != float64(expected) {
			t.Fatal(table, n, expected, err)
		}
	}
	huge := Cudd_Init(0, 100, 256, 256, 0)
	n, err := huge.CountZDD(huge.Univ[0])
	want := new(big.Int).Lsh(big.NewInt(1), 100)
	if err != nil || n.Cmp(want) != 0 || Cudd_zddCount(huge, huge.Univ[0]) != CUDD_OUT_OF_MEM || Cudd_zddCountDouble(huge, huge.Univ[0]) != math.Ldexp(1, 100) {
		t.Fatal("large count", n, err)
	}
	n.SetInt64(0)
	again, _ := huge.CountZDD(huge.Univ[0])
	if again.Cmp(want) != 0 {
		t.Fatal("mutable count escaped")
	}
	if _, err := dd.CountZDD(dd.LogicZero); err == nil {
		t.Fatal("BDD complement counted as ZDD")
	}
}
