package bmc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestExa32TwoInputAnd(t *testing.T) {
	p := Exa3_ManAllocV2(&Bmc_EsPar_t{nVars: 2, nNodes: 1, nLutSize: 2}, []uint64{0x8888888888888888})
	if Exa3_ManAddCnfStartV2(p, 1) == 0 {
		t.Fatal("startup unsatisfiable")
	}
	for mint := int32(1); mint >= 0; {
		if Exa3_ManAddCnfV2(p, mint) == 0 || p.pSat.Solve() != bsat.Satisfiable {
			t.Fatal("AND should be realizable")
		}
		mint = Exa3_ManEvalV2(p)
	}
}
func TestExa32OneLutCannotImplementThreeInputAnd(t *testing.T) {
	p := Exa3_ManAllocV2(&Bmc_EsPar_t{nVars: 3, nNodes: 1, nLutSize: 2}, []uint64{0x8080808080808080})
	if Exa3_ManAddCnfStartV2(p, 1) != 0 && p.pSat.Solve() != bsat.Unsatisfiable {
		t.Fatal("unused input should make this topology unsatisfiable")
	}
}
