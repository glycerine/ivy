package int

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

// A table indexes input, first register, second register, in that bit order.
func miniSatTruth(g *aig.Graph, inputs []aig.Edge, table uint8) aig.Edge {
	result := g.Const0()
	for assignment := 0; assignment < 8; assignment++ {
		if table&(1<<assignment) == 0 {
			continue
		}
		term := g.Const1()
		for i, edge := range inputs {
			if assignment&(1<<i) == 0 {
				edge = edge.Not()
			}
			term = g.AND(term, edge)
		}
		result = g.OR(result, term)
	}
	return result
}

func miniSatReachableBad(bad, next0, next1 uint8) bool {
	seen := [4]bool{true}
	for changed := true; changed; {
		changed = false
		for state := 0; state < 4; state++ {
			if !seen[state] {
				continue
			}
			for input := 0; input < 2; input++ {
				index := input | state<<1
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

func TestMiniSatTwoRegisterReachabilityAndWitness(t *testing.T) {
	rng := rand.New(rand.NewSource(11422006))
	decidedSafe, decidedUnsafe, boundedUnknown := 0, 0, 0
	for trial := 0; trial < 40; trial++ {
		bad := uint8(rng.Intn(256)) &^ 3 // exclude failures in the initial state
		next0, next1 := uint8(rng.Intn(256)), uint8(rng.Intn(256))
		unsafe := miniSatReachableBad(bad, next0, next1)
		for _, pudlak := range []int32{0, 1} {
			g := aig.NewGraph()
			inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
			g.AddOutput(miniSatTruth(g, inputs, bad))
			g.AddOutput(miniSatTruth(g, inputs, next0))
			g.AddOutput(miniSatTruth(g, inputs, next1))
			aig.Aig_ManSetRegNum(g, 2)
			var opts Inter_ManParams_t
			Inter_ManSetDefaultParams(&opts)
			opts.FUseMiniSat, opts.FUsePudlak, opts.NFramesMax = 1, pudlak, 30
			var frame int32
			got := Inter_ManPerformInterpolation(g, &opts, &frame)
			if got == -1 {
				// Bounded interpolation is not guaranteed to converge at this
				// frame limit. The established BSAT path must agree that this
				// particular case remains undecided.
				opts.FUseMiniSat = 0
				var baselineFrame int32
				if baseline := Inter_ManPerformInterpolation(g, &opts, &baselineFrame); baseline != -1 {
					t.Fatalf("trial %d bad=%02x next=%02x,%02x pudlak=%d: MiniSat unknown, BSAT=%d", trial, bad, next0, next1, pudlak, baseline)
				}
				boundedUnknown++
				continue
			}
			want := int32(1)
			if unsafe {
				want = 0
			}
			if got != want {
				t.Fatalf("trial %d bad=%02x next=%02x,%02x pudlak=%d: got %d want %d frame=%d", trial, bad, next0, next1, pudlak, got, want, frame)
			}
			if !unsafe {
				decidedSafe++
				continue
			}
			decidedUnsafe++
			cex := g.SequentialCounterexample()
			if cex == nil || cex.IFrame != frame {
				t.Fatalf("trial %d missing witness for frame %d", trial, frame)
			}
			if util.Abc_InfoHasBit(cex.PData, 0) || util.Abc_InfoHasBit(cex.PData, 1) {
				t.Fatalf("trial %d nonzero initial state", trial)
			}
			state := []bool{false, false}
			for f := int32(0); f <= frame; f++ {
				x := util.Abc_InfoHasBit(cex.PData, 2+f)
				out, err := g.Evaluate([]bool{x, state[0], state[1]})
				if err != nil {
					t.Fatal(err)
				}
				if f == frame && !out[0] {
					t.Fatalf("trial %d witness does not reach bad output", trial)
				}
				state = []bool{out[1], out[2]}
			}
		}
	}
	if decidedSafe == 0 || decidedUnsafe == 0 {
		t.Fatalf("insufficient decided coverage: safe=%d unsafe=%d unknown=%d", decidedSafe, decidedUnsafe, boundedUnknown)
	}
	t.Logf("two-register results: safe=%d unsafe=%d bounded-unknown=%d", decidedSafe, decidedUnsafe, boundedUnknown)
}
