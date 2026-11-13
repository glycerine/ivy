package live

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSourceBuildersCopyUnusedNodesInInputOrder(t *testing.T) {
	builders := []struct {
		name  string
		build func(*aig.Graph, aig.Edge) *aig.Graph
	}{
		{"liveness", func(p *aig.Graph, used aig.Edge) *aig.Graph {
			return LivenessToSafetyTransformation(FULL_BIERE_MODE, &Network{Graph: p}, p, []aig.Edge{used}, nil, nil, nil)
		}},
		{"absorber", func(p *aig.Graph, _ aig.Edge) *aig.Graph {
			zero, index := int32(0), int32(-1)
			return introduceAbsorberLogic(p, &zero, &index, 0)
		}},
		{"constrained", func(p *aig.Graph, used aig.Edge) *aig.Graph {
			index := int32(-1)
			return createNewAigWith0LivePo(p, []aig.Edge{used}, &index)
		}},
		{"arena", func(p *aig.Graph, used aig.Edge) *aig.Graph {
			index := int32(-1)
			return createNewAigWith0LivePoWithDSC(p, []aig.Edge{used}, &index, 0, 0, 0, nil)
		}},
		{"monotone", func(p *aig.Graph, _ aig.Edge) *aig.Graph {
			return createMonotoneTester(p, &aigPoIndices{0, 0, 0, -1}, &monotoneVectorsStruct{attrHintMonotone: []int32{0}}, nil)
		}},
		{"disjunctive", func(p *aig.Graph, _ aig.Edge) *aig.Graph {
			return createDisjunctiveMonotoneTester(p, &aigPoIndices{0, 0, 0, -1}, &antecedentConsequentVectorsStruct{attrConsequentCandidates: []int32{0}}, nil)
		}},
		{"combination", func(p *aig.Graph, _ aig.Edge) *aig.Graph { return generateGeneralDisjunctiveTester(nil, p, 1) }},
	}
	for _, tc := range builders {
		t.Run(tc.name, func(t *testing.T) {
			p := aig.Aig_ManStart(20)
			a, b := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
			unused := aig.Aig_And(p, a, b)
			used := aig.Aig_And(p, a, b.Not())
			unusedXor := p.NewXor(a, b)
			aig.Aig_ObjCreateCo(p, used)
			aig.Aig_ObjCreateCo(p, b)
			aig.Aig_ManSetRegNum(p, 1)
			dst := tc.build(p, used)
			u, v := aig.Aig_ObjCopy(p, unused), aig.Aig_ObjCopy(p, used)
			x := aig.Aig_ObjCopy(p, unusedXor)
			if x != u {
				t.Fatalf("unused XOR should map to same AND as source: got %v want %v", x, u)
			}
			if u == aig.InvalidEdge || v == aig.InvalidEdge {
				t.Fatalf("missing input-node scratch copies: unused=%v used=%v", u, v)
			}
			if u.NodeID() >= v.NodeID() {
				t.Fatalf("source input node order changed: unused=%d used=%d", u.NodeID(), v.NodeID())
			}
			if dst.Node(u).Kind != aig.And || dst.Node(v).Kind != aig.And {
				t.Fatalf("copies have unexpected kinds")
			}
		})
	}
}

func TestNativeLivenessKeepsLazyCopies(t *testing.T) {
	p := aig.Aig_ManStart(20)
	a, b := aig.Aig_ObjCreateCi(p), aig.Aig_ObjCreateCi(p)
	unused := aig.Aig_And(p, a, b)
	used := aig.Aig_And(p, a, b.Not())
	aig.Aig_ObjCreateCo(p, used)
	aig.Aig_ObjCreateCo(p, b)
	aig.Aig_ManSetRegNum(p, 1)
	_, err := LivenessToSafety(&Network{Graph: p}, TransformOptions{Mode: IGNORE_LIVENESS_KEEP_SAFETY_MODE, SafetyAssertions: []aig.Edge{used}})
	if err != nil {
		t.Fatal(err)
	}
	if got := aig.Aig_ObjCopy(p, unused); got != aig.InvalidEdge {
		t.Fatalf("native convenience copied unused node: %v", got)
	}
}
