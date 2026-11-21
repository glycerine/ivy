package bbr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// At each enabled clock the register bank increments. Its first all-ones
// state is reachable exactly 63 frames after the six-bit zero initial state.
func counterExample(partitioned int32) (int32, int32, bool) {
	const width = 6
	g := aig.NewGraph()
	carry := g.NewInput()
	var state, next [width]aig.Edge
	bad := g.Const1()
	for i := range state {
		state[i] = g.NewInput()
		bad = g.AND(bad, state[i])
	}
	for i, bit := range state {
		next[i] = g.XOR(bit, carry)
		carry = g.AND(carry, bit)
	}
	g.AddOutput(bad)
	for _, edge := range next {
		g.AddOutput(edge)
	}
	aig.Aig_ManSetRegNum(g, width)
	var params Saig_ParBbr_t
	Bbr_ManSetDefaultParams(&params)
	params.FSilent, params.FReorder, params.FReorderImage, params.FPartition = 1, 0, 0, partitioned
	status := Aig_ManVerifyUsingBdds(g, &params)
	model := g.SequentialCounterexample()
	if model == nil {
		return status, -1, false
	}
	return status, model.IFrame, saig.Saig_ManVerifyCex(g, model) != 0
}

func TestSixBitCounterReachability(t *testing.T) {
	for _, partitioned := range []int32{0, 1} {
		status, frame, replays := counterExample(partitioned)
		if status != 0 || frame != 63 || !replays {
			t.Errorf("partitioned %d: status %d frame %d replays %t", partitioned, status, frame, replays)
		}
	}
}

func BenchmarkSixBitCounterReachability(b *testing.B) {
	for _, tc := range []struct {
		name        string
		partitioned int32
	}{{"Partitioned", 1}, {"Monolithic", 0}} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if status, frame, replays := counterExample(tc.partitioned); status != 0 || frame != 63 || !replays {
					b.Fatalf("status %d frame %d replays %t", status, frame, replays)
				}
			}
		})
	}
}

func TestInitialStateRegistersPresentNextVariables(t *testing.T) {
	g := aig.NewGraph()
	for range 3 {
		g.NewInput()
	}
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const1())
	aig.Aig_ManSetRegNum(g, 2)
	dd := core.NewBDDManager(5)
	defer core.Cudd_Quit(dd)
	if Aig_ManInitStateVarMap(dd, g, false) == nil || core.Cudd_bddVarMap(dd, dd.Vars[0]) != dd.Vars[0] ||
		core.Cudd_bddVarMap(dd, dd.Vars[1]) != dd.Vars[3] || core.Cudd_bddVarMap(dd, dd.Vars[4]) != dd.Vars[2] {
		t.Fatal("initial cube did not register the latch variable swap")
	}
}
