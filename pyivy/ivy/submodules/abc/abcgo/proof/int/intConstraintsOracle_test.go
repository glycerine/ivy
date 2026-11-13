package int

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func constrainedReachableBad(bad, forbidden, next0, next1 uint8) bool {
	seen := [4]bool{true}
	for changed := true; changed; {
		changed = false
		for state := 0; state < 4; state++ {
			if !seen[state] {
				continue
			}
			for input := 0; input < 2; input++ {
				index := input | state<<1
				if forbidden&(1<<index) != 0 {
					continue
				}
				if bad&(1<<index) != 0 {
					return true
				}
				next := int(next0>>index)&1 | (int(next1>>index)&1)<<1
				if !seen[next] {
					seen[next], changed = true, true
				}
			}
		}
	}
	return false
}

func TestConstrainedInterpolationAgainstReachability(t *testing.T) {
	rng := rand.New(rand.NewSource(11422007))
	var decidedSafe, decidedUnsafe [3]int
	for trial := 0; trial < 30; trial++ {
		bad := uint8(rng.Intn(256)) &^ 3
		forbidden, next0, next1 := uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256))
		unsafe := constrainedReachableBad(bad, forbidden, next0, next1)
		for mode := 0; mode < 3; mode++ {
			g := aig.NewGraph()
			inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
			g.AddOutput(miniSatTruth(g, inputs, bad))
			g.AddOutput(miniSatTruth(g, inputs, forbidden))
			g.AddOutput(miniSatTruth(g, inputs, next0))
			g.AddOutput(miniSatTruth(g, inputs, next1))
			aig.Aig_ManSetRegNum(g, 2)
			aig.Aig_ManSetConstrNum(g, 1)
			var opts Inter_ManParams_t
			Inter_ManSetDefaultParams(&opts)
			opts.NFramesMax = 12
			if mode > 0 {
				opts.FUseMiniSat = 1
				opts.FUsePudlak = int32(mode - 1)
			}
			var frame int32
			got := Inter_ManPerformInterpolation(g, &opts, &frame)
			if got == -1 {
				continue
			}
			want := int32(1)
			if unsafe {
				want = 0
			}
			if got != want {
				t.Fatalf("trial=%d mode=%d bad=%02x forbidden=%02x next=%02x,%02x: got=%d want=%d", trial, mode, bad, forbidden, next0, next1, got, want)
			}
			if !unsafe {
				decidedSafe[mode]++
				continue
			}
			decidedUnsafe[mode]++
			cex := g.SequentialCounterexample()
			if cex == nil || cex.IFrame != frame {
				t.Fatalf("trial=%d mode=%d missing witness", trial, mode)
			}
			state := []bool{false, false}
			for f := int32(0); f <= frame; f++ {
				x := util.Abc_InfoHasBit(cex.PData, 2+f)
				out, err := g.Evaluate([]bool{x, state[0], state[1]})
				if err != nil {
					t.Fatal(err)
				}
				if out[1] {
					t.Fatalf("trial=%d mode=%d constraint violated in frame %d bad=%02x forbidden=%02x next=%02x,%02x", trial, mode, f, bad, forbidden, next0, next1)
				}
				if f == frame && !out[0] {
					t.Fatalf("trial=%d mode=%d witness is not bad", trial, mode)
				}
				state = []bool{out[2], out[3]}
			}
		}
	}
	for mode := 0; mode < 3; mode++ {
		if decidedSafe[mode] == 0 || decidedUnsafe[mode] == 0 {
			t.Fatalf("mode %d insufficient decided coverage: safe=%d unsafe=%d", mode, decidedSafe[mode], decidedUnsafe[mode])
		}
	}
}
