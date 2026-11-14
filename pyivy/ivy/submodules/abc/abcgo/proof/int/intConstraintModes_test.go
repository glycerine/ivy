package int

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestConstrainedInterpolationModesAgainstReachability(t *testing.T) {
	type mode struct {
		name                              string
		backward, two, loop, mini, pudlak int32
	}
	modes := []mode{
		{"forward-two", 0, 1, 0, 0, 0}, {"forward-loop", 0, 0, 1, 0, 0},
		{"backward-loop", 1, 0, 1, 0, 0},
		{"mini-two", 0, 1, 0, 1, 0}, {"mini-loop", 0, 0, 1, 1, 0},
		{"pudlak-two", 0, 1, 0, 1, 1}, {"pudlak-loop", 0, 0, 1, 1, 1},
	}
	rng := rand.New(rand.NewSource(20261113))
	decided := make(map[string][2]int)
	for trial := 0; trial < 40; trial++ {
		bad := uint8(rng.Intn(256)) &^ 3
		forbidden, next0, next1 := uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256))
		for _, m := range modes {
			constraint := forbidden
			if m.backward != 0 {
				constraint = 0
			}
			unsafe := constrainedReachableBad(bad, constraint, next0, next1)
			g := aig.NewGraph()
			inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
			g.AddOutput(miniSatTruth(g, inputs, bad))
			if m.backward == 0 {
				g.AddOutput(miniSatTruth(g, inputs, constraint))
			}
			g.AddOutput(miniSatTruth(g, inputs, next0))
			g.AddOutput(miniSatTruth(g, inputs, next1))
			aig.Aig_ManSetRegNum(g, 2)
			if m.backward == 0 {
				aig.Aig_ManSetConstrNum(g, 1)
			}
			var p Inter_ManParams_t
			Inter_ManSetDefaultParams(&p)
			p.NFramesMax = 10
			p.FUseBackward, p.FUseTwoFrames, p.FTransLoop, p.FUseMiniSat, p.FUsePudlak = m.backward, m.two, m.loop, m.mini, m.pudlak
			var frame int32
			got := Inter_ManPerformInterpolation(g, &p, &frame)
			if got < 0 {
				continue
			}
			want := int32(1)
			if unsafe {
				want = 0
			}
			if got != want {
				t.Fatalf("trial=%d mode=%s bad=%02x forbidden=%02x next=%02x,%02x got=%d want=%d", trial, m.name, bad, forbidden, next0, next1, got, want)
			}
			counts := decided[m.name]
			counts[got]++
			decided[m.name] = counts
			if got != 0 {
				continue
			}
			cex := g.SequentialCounterexample()
			if cex == nil {
				t.Fatalf("trial=%d mode=%s missing witness", trial, m.name)
			}
			if m.backward == 0 {
				// Source INT sends constraint POs to scalable BMC as ordinary
				// targets and does not enforce them as assumptions.
				continue
			}
			state := []bool{false, false}
			for f := int32(0); f <= frame; f++ {
				x := util.Abc_InfoHasBit(cex.PData, 2+f)
				out, err := g.Evaluate([]bool{x, state[0], state[1]})
				if err != nil {
					t.Fatal(err)
				}
				if m.backward == 0 && out[1] || f == frame && !out[0] {
					t.Fatalf("trial=%d mode=%s invalid witness at frame %d", trial, m.name, f)
				}
				if m.backward == 0 {
					state = []bool{out[2], out[3]}
				} else {
					state = []bool{out[1], out[2]}
				}
			}
		}
	}
	for _, m := range modes {
		if decided[m.name][0] == 0 || decided[m.name][1] == 0 {
			t.Fatalf("mode %s insufficient decided coverage: %v", m.name, decided[m.name])
		}
	}
}

// The bad state is legal, but all of its successors violate the constraint.
// A two-frame property must not require the bad state to have a legal successor.
func TestConstrainedBadStateWithoutLegalSuccessor(t *testing.T) {
	for _, flags := range []struct{ two, loop, mini, pudlak int32 }{
		{1, 0, 0, 0}, {0, 1, 0, 0}, {1, 0, 1, 0}, {1, 0, 1, 1}, {0, 1, 1, 0}, {0, 1, 1, 1},
	} {
		g := aig.NewGraph()
		x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(a)
		g.AddOutput(b)
		g.AddOutput(x)
		g.AddOutput(a)
		aig.Aig_ManSetRegNum(g, 2)
		aig.Aig_ManSetConstrNum(g, 1)
		var p Inter_ManParams_t
		Inter_ManSetDefaultParams(&p)
		p.NFramesMax = 8
		p.FUseTwoFrames, p.FTransLoop, p.FUseMiniSat, p.FUsePudlak = flags.two, flags.loop, flags.mini, flags.pudlak
		var frame int32
		got := Inter_ManPerformInterpolation(g, &p, &frame)
		if got != 0 || frame != 1 {
			t.Fatalf("flags %+v: got %d frame %d", flags, got, frame)
		}
		cex := g.SequentialCounterexample()
		if cex == nil {
			t.Fatalf("flags %+v: missing witness", flags)
		}
	}
}
