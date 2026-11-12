package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestIsoMappingRecoversInputPermutation(t *testing.T) {
	a := aig.NewGraph()
	ap, aq := a.NewInput(), a.NewInput()
	a.AddOutput(a.AND(ap, aq.Not()))

	b := aig.NewGraph()
	bq, bp := b.NewInput(), b.NewInput()
	b.AddOutput(b.AND(bp, bq.Not()))

	got := Iso_ManFindMapping(a, b, nil, nil, 0)
	if len(got) != 2 || got[0] != 1 || got[1] != 0 {
		t.Fatalf("input mapping = %v, want [1 0]", got)
	}
	if Iso_ManCheckMapping(a, b, []int32{1, 0}, 0) == 0 {
		t.Fatal("explicit isomorphic input mapping was rejected")
	}
	if Iso_ManCheckMapping(a, b, []int32{0, 1}, 0) != 0 {
		t.Fatal("incorrect input mapping was accepted")
	}
}

func TestIsoCanonicalDuplicatePreservesSequentialFunctions(t *testing.T) {
	p := aig.NewGraph()
	pi0, pi1 := p.NewInput(), p.NewInput()
	lo0, lo1 := p.NewInput(), p.NewInput()
	po0 := p.AND(pi0.Not(), lo1)
	po1 := p.OR(pi1, lo0)
	li0 := p.XOR(lo0, pi0)
	li1 := p.AND(lo1, pi1.Not())
	for _, output := range []aig.Edge{po0, po1, li0, li1} {
		p.AddOutput(output)
	}
	aig.Aig_ManSetRegNum(p, 2)

	inputOrder := Saig_ManFindIsoPerm(p, 0)
	outputOrder := Saig_ManFindIsoPermCos(p, inputOrder)
	copyGraph := Saig_ManDupIsoCanonical(p, 0)
	if copyGraph.InputCount() != p.InputCount() || copyGraph.OutputCount() != p.OutputCount() || aig.Aig_ManRegNum(copyGraph) != aig.Aig_ManRegNum(p) {
		t.Fatal("canonical duplicate changed terminal counts")
	}
	for mask := 0; mask < 16; mask++ {
		oldInputs := make([]bool, p.InputCount())
		for i := range oldInputs {
			oldInputs[i] = mask&(1<<i) != 0
		}
		newInputs := make([]bool, len(inputOrder))
		for i, oldIndex := range inputOrder {
			newInputs[i] = oldInputs[oldIndex]
		}
		oldOutputs, err := p.Evaluate(oldInputs)
		if err != nil {
			t.Fatal(err)
		}
		newOutputs, err := copyGraph.Evaluate(newInputs)
		if err != nil {
			t.Fatal(err)
		}
		for i, oldIndex := range outputOrder {
			if newOutputs[i] != oldOutputs[oldIndex] {
				t.Fatalf("mask %04b output %d differs after canonical copy", mask, i)
			}
		}
	}
}

func TestIsoFilterGroupsIdenticalPrimaryOutputsAndKeepsLatchInputs(t *testing.T) {
	p := aig.NewGraph()
	pi := p.NewInput()
	lo := p.NewInput()
	root := p.AND(pi, lo.Not())
	p.AddOutput(root)
	p.AddOutput(root)
	p.AddOutput(root.Not())
	p.AddOutput(p.XOR(pi, lo))
	aig.Aig_ManSetRegNum(p, 1)

	reduced, classes := Iso_ManFilterPos(p, 0)
	duplicateClass := false
	for _, group := range classes {
		duplicateClass = duplicateClass || len(group) == 2
	}
	if len(classes) != 2 || !duplicateClass {
		t.Fatalf("candidate output classes = %v, want two classes including one duplicate pair", classes)
	}
	if reduced.OutputCount() != 3 || aig.Aig_ManRegNum(reduced) != 1 {
		t.Fatalf("reduced output count/registers = %d/%d, want 3/1", reduced.OutputCount(), aig.Aig_ManRegNum(reduced))
	}
}

func TestIsoNegativeEdgeCountIncludesNodesAndOutputDrivers(t *testing.T) {
	p := aig.NewGraph()
	a, b := p.NewInput(), p.NewInput()
	root := p.AND(a.Not(), b)
	p.AddOutput(root.Not())
	if got := Iso_ManNegEdgeNum(p); got != 2 {
		t.Fatalf("complemented edge count = %d, want 2", got)
	}
}

func TestIsoObjectCountCountsUniqueConeNodesAndComplementedEdges(t *testing.T) {
	p := aig.NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	shared := p.AND(a.Not(), b)
	root := p.AND(shared, c.Not())

	nodes, complemented := Iso_ManObjCount(p, root)
	if nodes != 2 || complemented != 2 {
		t.Fatalf("cone counts = (%d nodes, %d complemented edges), want (2, 2)", nodes, complemented)
	}
}
