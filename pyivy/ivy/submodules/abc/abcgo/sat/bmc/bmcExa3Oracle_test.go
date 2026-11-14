package bmc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestExa3IndependentTruthTables(t *testing.T) {
	// All normal two-input functions that depend on both inputs.
	for _, truth := range []uint64{0x2, 0x4, 0x6, 0x8, 0xe} {
		for _, incremental := range []int32{0, 1} {
			p := Exa3_ManAlloc(&Bmc_EsPar_t{NVars: 2, NNodes: 1, NLutSize: 2, FUseIncr: incremental}, []uint64{truth})
			if Exa3_ManAddCnfStart(p, 0) == 0 {
				t.Fatalf("truth=%x mode=%d initial CNF", truth, incremental)
			}
			if incremental != 0 && p.pSat.Solve() != bsat.Satisfiable {
				t.Fatal("initial model")
			}
			mint := int32(1)
			for step := 0; step < 20 && mint >= 0; step++ {
				added := int32(0)
				if incremental != 0 {
					added = Exa3_ManAddCnf2(p, mint)
				} else {
					added = Exa3_ManAddCnf(p, mint)
				}
				if added == 0 || p.pSat.Solve() != bsat.Satisfiable {
					t.Fatalf("truth=%x mode=%d unexpectedly unsat", truth, incremental)
				}
				mint = Exa3_ManEval(p)
			}
			if mint != -1 {
				t.Fatalf("truth=%x mode=%d did not converge", truth, incremental)
			}
			// Decode selected fanins and LUT bits directly from the SAT model.
			fans := [2]int{-1, -1}
			for k := 0; k < 2; k++ {
				for id, v := range p.marks[2][k] {
					if v != 0 {
						if value, _ := p.pSat.Value(v); value {
							if fans[k] >= 0 {
								t.Fatal("multiple fanins")
							}
							fans[k] = id
						}
					}
				}
			}
			for m := 0; m < 4; m++ {
				lookup := 0
				for k, id := range fans {
					if id < 0 || id > 1 {
						t.Fatalf("invalid fanins %v", fans)
					}
					if m&(1<<id) != 0 {
						lookup |= 1 << k
					}
				}
				got := false
				if lookup > 0 {
					got, _ = p.pSat.Value(int32(lookup))
				}
				if got != (truth&(1<<m) != 0) {
					t.Fatalf("truth=%x mode=%d assignment=%d fans=%v got=%v", truth, incremental, m, fans, got)
				}
			}
		}
	}
}
