package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// test_vectors/native_xor_copy_oracle.c links the original ABC constructor;
// it reports origxor=1 copiedand=1 copiedxor=0 complemented=0.
func TestSourceBuildersCopyNativeXorAsAnd(t *testing.T) {
	p := aig.Aig_ManStart(20)
	a, b := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	x := p.NewXor(a, b)
	if p.Node(x).Kind != aig.ExclusiveOr {
		t.Fatal("fixture requires native XOR")
	}
	aig.Aig_ObjCreateCo(p, x)
	aig.Aig_ObjCreateCo(p, a)
	aig.Aig_ManSetRegNum(p, 1)
	index := int32(-1)
	source := createNewAigWith0LivePo(p, []aig.Edge{x}, &index)
	copied := aig.Aig_Regular(outputDriver(source, 0))
	if source.Node(copied).Kind != aig.And {
		t.Fatalf("C source copied XOR as AND; got %v", source.Node(copied).Kind)
	}
	// The native convenience model retains XOR semantics.
	network := &Network{Graph: p}
	native, err := LivenessToSafety(network, TransformOptions{Mode: IGNORE_LIVENESS_KEEP_SAFETY_MODE, SafetyAssertions: []aig.Edge{x}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := native.Graph.Evaluate([]bool{true, false})
	if err != nil {
		t.Fatal(err)
	}
	if out[0] {
		t.Fatalf("native convenience changed XOR truth table: %v", out)
	}
}
