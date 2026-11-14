package bmc

import (
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestMajoritySynthesis(t *testing.T) {
	if Maj_ManExactSynthesis(3, 1, 0, 0, 0) != 1 {
		t.Fatal("3-input majority requires one gate")
	}
	if Maj_ManExactSynthesis(5, 1, 0, 0, 0) != 0 {
		t.Fatal("5-input majority cannot use one gate")
	}
}

func TestExactTwoInputSynthesis(t *testing.T) {
	for _, want := range []uint64{0x8888888888888888, 0x6666666666666666, 0xeeeeeeeeeeeeeeee} {
		pars := &Bmc_EsPar_t{NVars: 2, NNodes: 1}
		p := Exa_ManAlloc(pars, []uint64{want})
		if Exa_ManAddCnfStart(p, 0) == 0 {
			t.Fatal("structural constraints")
		}
		for mint := int32(1); ; {
			if Exa_ManAddCnf(p, mint) == 0 {
				t.Fatalf("cannot synthesize %x", want)
			}
			if p.pSat.Solve() != bsat.Satisfiable {
				t.Fatalf("unsat for %x", want)
			}
			mint = Exa_ManEval(p)
			if mint < 0 {
				break
			}
		}
	}
}
