package bmc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestExa3AND(t *testing.T) {
	for _, incremental := range []int32{0, 1} {
		p := Exa3_ManAlloc(&Bmc_EsPar_t{nVars: 2, nNodes: 1, nLutSize: 2, fUseIncr: incremental}, []uint64{8})
		if Exa3_ManAddCnfStart(p, 0) == 0 {
			t.Fatal("start")
		}
		if incremental != 0 {
			if p.pSat.Solve() != bsat.Satisfiable {
				t.Fatal("start solve")
			}
		}
		mint := int32(1)
		for step := 0; step < 20 && mint >= 0; step++ {
			var added int32
			if incremental != 0 {
				added = Exa3_ManAddCnf2(p, mint)
			} else {
				added = Exa3_ManAddCnf(p, mint)
			}
			if added == 0 || p.pSat.Solve() != bsat.Satisfiable {
				t.Fatalf("mode %d no solution", incremental)
			}
			mint = Exa3_ManEval(p)
		}
		if mint != -1 {
			t.Fatalf("mode %d mismatch at %d", incremental, mint)
		}
	}
}

func TestExa3ThreeInputAnd(t *testing.T) {
	for _, incremental := range []int32{0, 1} {
		p := Exa3_ManAlloc(&Bmc_EsPar_t{nVars: 3, nNodes: 2, nLutSize: 2, fUseIncr: incremental}, []uint64{0x80})
		if Exa3_ManAddCnfStart(p, 0) == 0 {
			t.Fatal("start")
		}
		if incremental != 0 && p.pSat.Solve() != bsat.Satisfiable {
			t.Fatal("initial solve")
		}
		mint := int32(1)
		for step := 0; step < 100 && mint >= 0; step++ {
			var added int32
			if incremental != 0 {
				added = Exa3_ManAddCnf2(p, mint)
			} else {
				added = Exa3_ManAddCnf(p, mint)
			}
			if added == 0 || p.pSat.Solve() != bsat.Satisfiable {
				t.Fatalf("mode %d step %d no solution", incremental, step)
			}
			mint = Exa3_ManEval(p)
		}
		if mint != -1 {
			t.Fatalf("mode %d did not converge, minterm %d", incremental, mint)
		}
	}
}
