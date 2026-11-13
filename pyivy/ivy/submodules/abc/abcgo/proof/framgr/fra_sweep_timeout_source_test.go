package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCompleteSweepPropagatesSATExhaustion(t *testing.T) {
	g := aig.NewGraph()
	v := make([]aig.Edge, 8)
	for i := range v {
		v[i] = g.NewInput()
	}
	left := v[0]
	for _, x := range v[1:] {
		left = g.XOR(left, x)
	}
	var parity func([]aig.Edge) aig.Edge
	parity = func(xs []aig.Edge) aig.Edge {
		if len(xs) == 1 {
			return xs[0]
		}
		mid := len(xs) / 2
		return g.XOR(parity(xs[:mid]), parity(xs[mid:]))
	}
	right := parity(v)
	g.AddOutput(left)
	g.AddOutput(right)
	var pars CombOptions
	Fra_ParamsDefault(&pars)
	pars.NBTLimitNode = 1
	pars.FConeBias = 0
	pars.NSimWords = 1
	var failures, logged int32
	out, mapping := SweepMappedCallback(g, pars, func(p *Manager) {
		failures = p.NSatFailsReal
		if p.VTimeouts != nil {
			logged = int32(vec.Vec_PtrSize(p.VTimeouts))
		}
	})
	if out == nil || len(mapping) == 0 {
		t.Fatal("missing sweep result")
	}
	// ABC on the same construction and parameters reports two exhausted
	// SAT calls, two timeout entries, and 39 nodes in the result.
	if failures != 2 || logged != 2 || aig.Aig_ManNodeNum(out) != 39 {
		t.Fatalf("sweep differs from compiled C: failures=%d timeout entries=%d nodes=%d", failures, logged, aig.Aig_ManNodeNum(out))
	}
	for mask := 0; mask < 256; mask++ {
		values := make([]bool, 8)
		for i := range values {
			values[i] = mask&(1<<i) != 0
		}
		got, err := out.Evaluate(values)
		if err != nil || got[0] != got[1] {
			t.Fatalf("input %v outputs %v err %v", values, got, err)
		}
	}
}
