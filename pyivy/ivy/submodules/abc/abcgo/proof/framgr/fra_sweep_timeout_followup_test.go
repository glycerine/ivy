package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestSweepRecordsDisproofsAfterTimeout(t *testing.T) {
	g := aig.NewGraph()
	v := make([]aig.Edge, 20)
	for i := range v {
		v[i] = g.NewInput()
	}
	left := v[0]
	for _, x := range v[1:8] {
		left = g.XOR(left, x)
	}
	var parity func([]aig.Edge) aig.Edge
	parity = func(xs []aig.Edge) aig.Edge {
		if len(xs) == 1 {
			return xs[0]
		}
		m := len(xs) / 2
		return g.XOR(parity(xs[:m]), parity(xs[m:]))
	}
	right := parity(v[:8])
	rare := v[0]
	for i, x := range v[1:] {
		if i%2 == 0 {
			x = x.Not()
		}
		rare = g.AND(rare, x)
	}
	g.AddOutput(left)
	g.AddOutput(right)
	g.AddOutput(rare)
	var pars CombOptions
	Fra_ParamsDefault(&pars)
	pars.NBTLimitNode = 1
	pars.FConeBias = 0
	pars.NSimWords = 1
	var ids []int
	var failures, sat int32
	out, _ := SweepMappedCallback(g, pars, func(p *Manager) {
		failures, sat = p.NSatFailsReal, p.NSatCallsSat
		if p.VTimeouts != nil {
			for _, item := range vec.Vec_PtrArray(p.VTimeouts) {
				ids = append(ids, item.(aig.Edge).NodeID())
			}
		}
	})
	want := []int{47, 59, 63, 68, 69, 72, 73, 76, 77}
	if failures != 2 || sat != 7 || !reflect.DeepEqual(ids, want) || rare.NodeID() != 77 || aig.Aig_ManNodeNum(out) != 57 {
		t.Fatalf("source mismatch: failures=%d SAT=%d ids=%v rare=%d nodes=%d", failures, sat, ids, rare.NodeID(), aig.Aig_ManNodeNum(out))
	}
	for _, id := range ids {
		if repr := aig.Aig_ObjRepr(g, aig.Edge(id<<1)); repr != aig.InvalidEdge {
			t.Fatalf("recorded candidate %d retained representative %v", id, repr)
		}
	}
}
