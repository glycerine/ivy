package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// Compare the engine with explicit reachability over all eight states. The
// truth tables are deterministic, with independent tables for each output.
func TestPdrThreeRegisterReachability(t *testing.T) {
	seed := uint32(0xd12f89a7)
	next := func() uint16 { seed ^= seed << 13; seed ^= seed >> 17; seed ^= seed << 5; return uint16(seed) }
	for trial := 0; trial < 160; trial++ {
		tables := [4]uint16{next(), next(), next(), next()}
		g := aig.NewGraph()
		vars := [4]aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		truth := func(table uint16) aig.Edge {
			result := g.Const0()
			for row := 0; row < 16; row++ {
				if table&(1<<row) == 0 {
					continue
				}
				term := g.Const1()
				for i, variable := range vars {
					if row&(1<<i) == 0 {
						variable = variable.Not()
					}
					term = g.AND(term, variable)
				}
				result = g.AND(result.Not(), term.Not()).Not()
			}
			return result
		}
		g.AddOutput(truth(tables[3]))
		for _, table := range tables[:3] {
			g.AddOutput(truth(table))
		}
		aig.Aig_ManSetRegNum(g, 3)
		reached := [8]bool{true}
		bad := false
		for changed := true; changed; {
			changed = false
			for state, seen := range reached {
				if !seen {
					continue
				}
				for input := 0; input < 2; input++ {
					row := input + 2*state
					bad = bad || tables[3]&(1<<row) != 0
					nextState := 0
					for i := 0; i < 3; i++ {
						nextState |= int((tables[i]>>row)&1) << i
					}
					if !reached[nextState] {
						reached[nextState], changed = true, true
					}
				}
			}
		}
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSilent = 1
		pars.fMonoCnf = int32(trial & 1)
		// ABC’s updated simulator can produce invalid witnesses; see PORTING.md.
		pars.fNewXSim = 0
		pars.fUseAbs = int32((trial >> 2) & 1)
		pars.fFlopPrio = int32((trial >> 3) & 1)
		pars.fShiftStart = int32((trial >> 4) & 1)
		pars.fReuseProofOblig = int32((trial >> 5) & 1)
		var result int32
		if trial&64 != 0 {
			pars.nFrameMax = 1
			result = IPdr_ManSolve(g, &pars)
		} else {
			result = Pdr_ManSolve(g, &pars)
		}
		want := int32(1)
		if bad {
			want = 0
		}
		if result != want {
			t.Fatalf("trial %d tables %v: got %d, want %d", trial, tables, result, want)
		}
		if result == 0 && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
			t.Fatalf("trial %d tables %v: invalid counterexample", trial, tables)
		}
	}
}
