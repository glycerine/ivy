package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"math/rand"
	"testing"
)

// exhaustiveSmallMiter independently explores all reachable states of a
// one-input, two-register miter and reports whether either PO can fail.
func exhaustiveSmallMiter(t *testing.T, g *aig.Graph) bool {
	t.Helper()
	seen := map[int]bool{0: true}
	queue := []int{0}
	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]
		for input := 0; input < 2; input++ {
			out, err := g.Evaluate([]bool{input != 0, state&1 != 0, state&2 != 0})
			if err != nil {
				t.Fatal(err)
			}
			if out[0] || out[1] {
				return true
			}
			next := 0
			if out[2] {
				next |= 1
			}
			if out[3] {
				next |= 2
			}
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

func TestSecSeparateInterpolationAgainstReachability(t *testing.T) {
	random := rand.New(rand.NewSource(7411))
	decided := 0
	for trial := 0; trial < 32; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 7; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		for i := 0; i < 4; i++ {
			edge := edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				edge = edge.Not()
			}
			g.AddOutput(edge)
		}
		aig.Aig_ManSetRegNum(g, 2)
		fails := exhaustiveSmallMiter(t, g)
		var pars Fra_Sec_t
		Fra_SecSetDefaultParams(&pars)
		pars.fSilent = 1
		pars.fInterSeparate = 1
		pars.fRetimeFirst = 0
		pars.fRetimeRegs = 0
		pars.fInduction = 0
		pars.fFraiging = 0
		pars.fReachability = 0
		pars.fUsePdr = 0
		status := Fra_FraigSec(g, &pars, nil)
		switch status {
		case 1:
			decided++
			if fails {
				t.Fatalf("trial %d proved a failing miter", trial)
			}
		case 0:
			decided++
			if !fails {
				t.Fatalf("trial %d reported failure for a safe miter", trial)
			}
			if cex := g.SequentialCounterexample(); cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
				t.Fatalf("trial %d invalid witness %+v", trial, cex)
			}
		case -1:
		default:
			t.Fatalf("trial %d invalid status %d", trial, status)
		}
	}
	if decided < 24 {
		t.Fatalf("only %d of 32 cases were decided", decided)
	}
}
