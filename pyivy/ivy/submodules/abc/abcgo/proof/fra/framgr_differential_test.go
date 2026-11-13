package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/framgr"
	"math/rand"
	"testing"
)

// The extracted manager must preserve the original manager's combinational
// behavior under the exact configuration used by the legacy CEC fallback.
func TestSharedCECSweepMatchesOriginalManager(t *testing.T) {
	random := rand.New(rand.NewSource(94137))
	for trial := 0; trial < 50; trial++ {
		source := aig.NewGraph()
		nodes := []aig.Edge{source.NewInput(), source.NewInput(), source.NewInput(), source.NewInput()}
		for i := 0; i < 24; i++ {
			a, b := nodes[random.Intn(len(nodes))], nodes[random.Intn(len(nodes))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			if random.Intn(2) == 0 {
				nodes = append(nodes, source.AND(a, b))
			} else {
				nodes = append(nodes, source.XOR(a, b))
			}
		}
		for i := 0; i < 3; i++ {
			edge := nodes[4+random.Intn(24)]
			if random.Intn(2) == 0 {
				edge = edge.Not()
			}
			source.AddOutput(edge)
		}
		original, shared := aig.Aig_ManDupSimple(source), aig.Aig_ManDupSimple(source)
		var pars Fra_Par_t
		Fra_ParamsDefault(&pars)
		pars.fDontShowBar, pars.fProve = 1, 1
		pars.nBTLimitNode, pars.nBTLimitMiter = 2, 300
		old := Fra_FraigPerform(original, &pars)
		got := framgr.SweepForCEC(shared, 2, 300)
		if old == nil || got == nil {
			t.Fatalf("trial %d missing reduced result", trial)
		}
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			want, err := old.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := got.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			reference, err := source.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if want[i] != actual[i] || actual[i] != reference[i] {
					t.Fatalf("trial %d assignment %d: old=%v shared=%v source=%v", trial, mask, want, actual, reference)
				}
			}
		}
	}
}

func TestSharedCECSweepSparseSatAndUnsat(t *testing.T) {
	for _, sat := range []bool{false, true} {
		makeGraph := func() *aig.Graph {
			g := aig.NewGraph()
			g.NewInput()
			a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
			left := g.AND(a, b)
			right := g.AND(a, b.Not())
			if sat {
				right = g.AND(a, c)
			}
			g.AddOutput(g.AND(left, right))
			return g
		}
		original, shared := makeGraph(), makeGraph()
		var pars Fra_Par_t
		Fra_ParamsDefault(&pars)
		pars.fDontShowBar, pars.fProve = 1, 1
		pars.nBTLimitNode, pars.nBTLimitMiter = 2, 300
		old := Fra_FraigPerform(original, &pars)
		got := framgr.SweepForCEC(shared, 2, 300)
		if old == nil || got == nil || old.InputCount() != 4 || got.InputCount() != 4 {
			t.Fatalf("sat=%t incorrect input retention", sat)
		}
		oldSat, newSat := false, false
		for mask := 0; mask < 16; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			a, e := old.Evaluate(inputs)
			if e != nil {
				t.Fatal(e)
			}
			b, e := got.Evaluate(inputs)
			if e != nil {
				t.Fatal(e)
			}
			if a[0] != b[0] {
				t.Fatalf("sat=%t input %v outputs %v %v", sat, inputs, a, b)
			}
			oldSat = oldSat || a[0]
			newSat = newSat || b[0]
		}
		if oldSat != sat || newSat != sat {
			t.Fatalf("sat=%t old=%t new=%t", sat, oldSat, newSat)
		}
	}
}
