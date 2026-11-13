package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// fraSat.c forwards local and global limits to BSAT. A real parity proof
// distinguishes conflict exhaustion from a valid equivalence proof.
func TestEquivalenceConflictLimitReturnsUnknown(t *testing.T) {
	for _, tc := range []struct {
		name                string
		local               int32
		global, inspections int64
		want                int32
	}{
		{"local conflicts", 1, 0, 0, -1}, {"global conflicts", 0, 1, 0, -1}, {"global inspections", 0, 0, 1, -1}, {"unlimited", 0, 0, 0, 1},
	} {
		g := aig.NewGraph()
		inputs := make([]aig.Edge, 8)
		for i := range inputs {
			inputs[i] = g.NewInput()
		}
		left := inputs[0]
		for _, x := range inputs[1:] {
			left = g.XOR(left, x)
		}
		var parity func([]aig.Edge) aig.Edge
		parity = func(v []aig.Edge) aig.Edge {
			if len(v) == 1 {
				return v[0]
			}
			m := len(v) / 2
			return g.XOR(parity(v[:m]), parity(v[m:]))
		}
		right := parity(inputs)
		p := satTestManager(g)
		p.PPars.NBTLimitNode = tc.local
		p.NBTLimitGlobal, p.NInsLimitGlobal = tc.global, tc.inspections
		result := Fra_NodesAreEquiv(p, aig.Aig_Regular(left), aig.Aig_Regular(right))
		if tc.want == -1 {
			if result != -1 || p.NSatFailsReal != 1 || p.NSatProof != 0 {
				t.Fatalf("case=%s result=%d failures=%d proofs=%d", tc.name, result, p.NSatFailsReal, p.NSatProof)
			}
		} else if result != 1 || p.NSatProof != 1 {
			t.Fatalf("case=%s result=%d proofs=%d", tc.name, result, p.NSatProof)
		}
	}
}
