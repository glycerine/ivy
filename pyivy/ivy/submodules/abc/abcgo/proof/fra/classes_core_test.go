package fra

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestClassPartitionGraphSelection(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	ab := g.AND(a, b)
	abc := g.AND(ab, c)
	g.AddOutput(abc)
	allSame := func(int) uint64 { return 0 }
	notConst := func(int) bool { return false }
	equal := func(int, int) bool { return true }
	p := newClassPartition(len(g.Nodes()))
	p.PrepareGraph(g, false, 1, allSame, notConst, equal)
	if !reflect.DeepEqual(p.Classes, [][]int{{a.NodeID(), b.NodeID(), c.NodeID(), ab.NodeID()}}) {
		t.Fatalf("level-filtered classes %v", p.Classes)
	}
	p.PrepareGraph(g, true, 0, allSame, notConst, equal)
	if !reflect.DeepEqual(p.Classes, [][]int{{a.NodeID(), b.NodeID(), c.NodeID()}}) {
		t.Fatalf("latch correspondence classes %v", p.Classes)
	}
}

func TestClassPartitionHashCollisionAndRefinement(t *testing.T) {
	signature := []int{0, 1, 2, 1, 3, 2, 0}
	p := newClassPartition(len(signature))
	hash := func(int) uint64 { return 0 } // Deliberately collide every signature.
	constant := func(id int) bool { return signature[id] == 0 }
	equal := func(a, b int) bool { return signature[a] == signature[b] }
	p.Prepare([]int{1, 2, 3, 4, 5, 6}, hash, constant, equal)
	if !reflect.DeepEqual(p.Classes, [][]int{{1, 3}, {2, 5}}) || !reflect.DeepEqual(p.Constants, []int{6}) {
		t.Fatalf("classes %v constants %v", p.Classes, p.Constants)
	}
	if p.LiteralCount() != 3 || p.PairCount() != 2 || p.Reprs[3] != 1 || p.Reprs[5] != 2 || p.Reprs[6] != 0 {
		t.Fatalf("counts or representatives incorrect: %+v", p)
	}
	signature[3] = 9
	if p.Refine(equal) != 1 || !reflect.DeepEqual(p.Classes, [][]int{{2, 5}}) || p.Reprs[3] != -1 {
		t.Fatalf("split did not remove singleton: %+v", p)
	}
}

func TestClassPartitionRejectsInvalidCandidatesBeforeMutation(t *testing.T) {
	p := newClassPartition(4)
	p.Prepare([]int{1, 2}, func(int) uint64 { return 0 }, func(int) bool { return false }, func(int, int) bool { return true })
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid order accepted")
			}
		}()
		p.Prepare([]int{2, 1}, func(int) uint64 { return 0 }, func(int) bool { return false }, func(int, int) bool { return true })
	}()
	if !reflect.DeepEqual(p.Classes, [][]int{{1, 2}}) || p.Reprs[2] != 1 {
		t.Fatal("invalid candidates mutated partition")
	}
}

func TestClassPartitionConstantRefinement(t *testing.T) {
	signature := []int{0, 0, 0, 0, 0}
	p := newClassPartition(len(signature))
	constant := func(id int) bool { return signature[id] == 0 }
	equal := func(a, b int) bool { return signature[a] == signature[b] }
	p.Prepare([]int{1, 2, 3, 4}, func(int) uint64 { return 1 }, constant, equal)
	signature[1], signature[2], signature[3] = 1, 1, 2
	if p.RefineConstants(constant, equal, true) != 2 {
		t.Fatal("expected removal from constant class and one split")
	}
	if !reflect.DeepEqual(p.Constants, []int{4}) || !reflect.DeepEqual(p.Classes, [][]int{{1, 2}}) || p.Reprs[2] != 1 || p.Reprs[3] != -1 {
		t.Fatalf("unexpected refined partition: %+v", p)
	}
}

func TestClassPartitionRandomRefinementNeverMerges(t *testing.T) {
	rng := rand.New(rand.NewSource(68125))
	for trial := 0; trial < 500; trial++ {
		const n = 24
		signature := make([]int, n+1)
		nodes := make([]int, n)
		for i := range nodes {
			nodes[i] = i + 1
			signature[i+1] = 1 + rng.Intn(4)
		}
		p := newClassPartition(n + 1)
		equal := func(a, b int) bool { return signature[a] == signature[b] }
		p.Prepare(nodes, func(int) uint64 { return 0 }, func(int) bool { return false }, equal)
		before := append([]int(nil), signature...)
		for i := 1; i <= n; i++ {
			signature[i] = signature[i]*4 + rng.Intn(4)
		}
		p.Refine(equal)
		membership := make([]int, n+1)
		for i := range membership {
			membership[i] = -1
		}
		for i, class := range p.Classes {
			for _, id := range class {
				membership[id] = i
			}
		}
		for _, class := range p.Classes {
			for _, id := range class[1:] {
				if before[id] != before[class[0]] || signature[id] != signature[class[0]] || p.Reprs[id] != class[0] {
					t.Fatalf("trial %d invalid class %v", trial, class)
				}
			}
		}
		for a := 1; a <= n; a++ {
			for b := a + 1; b <= n; b++ {
				want := before[a] == before[b] && signature[a] == signature[b]
				got := membership[a] >= 0 && membership[a] == membership[b]
				if got != want {
					t.Fatalf("trial %d pair %d,%d: got %v want %v", trial, a, b, got, want)
				}
			}
		}
	}
}

func TestClassPartitionLatchCorrespondenceAndPrune(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	lo0, lo1 := g.NewInput(), g.NewInput()
	g.AddOutput(lo0)
	g.AddOutput(lo1)
	aig.Aig_ManSetRegNum(g, 2)
	p := newClassPartition(len(g.Nodes()))
	p.LatchCorrespondence(g)
	if !reflect.DeepEqual(p.Constants, []int{lo0.NodeID(), lo1.NodeID()}) || p.Reprs[lo0.NodeID()] != 0 {
		t.Fatalf("incorrect latch candidates: %+v", p)
	}
	weights := make([]int, len(p.Reprs))
	weights[lo0.NodeID()], weights[lo1.NodeID()] = 2, 8
	p.PruneByWeight(weights)
	if !reflect.DeepEqual(p.Constants, []int{lo1.NodeID()}) || p.Reprs[lo0.NodeID()] != -1 {
		t.Fatalf("incorrect pruning: %+v", p)
	}
}

func TestClassPartitionSelectRepresentatives(t *testing.T) {
	p := newClassPartition(6)
	p.Classes = [][]int{{1, 2, 3, 4}}
	for _, id := range p.Classes[0][1:] {
		p.Reprs[id] = 1
	}
	support := []int{0, 3, 2, 2, 2}
	level := []int{0, 1, 4, 3, 3}
	p.SelectRepresentatives(func(id int) int { return support[id] }, func(id int) int { return level[id] })
	if !reflect.DeepEqual(p.Classes, [][]int{{3, 2, 1, 4}}) || p.Reprs[3] != -1 || p.Reprs[1] != 3 || p.Reprs[2] != 3 || p.Reprs[4] != 3 {
		t.Fatalf("incorrect representative selection: %+v", p)
	}
}

func TestClassPartitionCopiesRepresentatives(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	p := newClassPartition(len(g.Nodes()))
	p.Reprs[a.NodeID()] = 0
	p.Reprs[b.NodeID()] = a.NodeID()
	p.Reprs[c.NodeID()] = a.NodeID()
	p.CopyRepresentatives(g, []int{b.NodeID()})
	if aig.Aig_ObjRepr(g, a) != g.Const1() || aig.Aig_ObjRepr(g, b) != aig.InvalidEdge || aig.Aig_ObjRepr(g, c) != a {
		t.Fatal("wrong copied representatives")
	}
	p.Reprs[c.NodeID()] = -1
	p.CopyRepresentatives(g, nil)
	if aig.Aig_ObjRepr(g, c) != aig.InvalidEdge {
		t.Fatal("stale representative retained")
	}
}

func TestClassPartitionSequentialDerivation(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(x) // latch input becomes x in the next frame
	aig.Aig_ManSetRegNum(g, 1)
	p := newClassPartition(len(g.Nodes()))
	p.Reprs[lo.NodeID()] = 0
	unrolled := p.DeriveSequentialAIG(g, 1)
	if unrolled.InputCount() != 3 || unrolled.OutputCount() != 2 || unrolled.Assertions != 1 {
		t.Fatalf("incorrect unrolling sizes: inputs=%d outputs=%d assertions=%d", unrolled.InputCount(), unrolled.OutputCount(), unrolled.Assertions)
	}
	for mask := 0; mask < 8; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		outputs, err := unrolled.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if outputs[0] != !values[0] || outputs[1] != !values[1] {
			t.Fatalf("inputs=%v outputs=%v", values, outputs)
		}
	}
}

func TestClassPartitionSequentialInternalMiter(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	gate := g.AND(x, lo)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	p := newClassPartition(len(g.Nodes()))
	p.Reprs[gate.NodeID()] = x.NodeID()
	unrolled := p.DeriveSequentialAIG(g, 1)
	if unrolled.InputCount() != 3 || unrolled.OutputCount() != 2 || unrolled.Assertions != 1 {
		t.Fatal("incorrect internal miter layout")
	}
	for mask := 0; mask < 8; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		outputs, err := unrolled.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		if outputs[0] != (!values[1] || values[0]) || outputs[1] != (!values[2] || values[1]) {
			t.Fatalf("inputs=%v outputs=%v", values, outputs)
		}
	}
}

func TestSimulationDrivesCandidateRefinement(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	g.AddOutput(gate)
	s := newFraSimulation(g, 0, 1, 1)
	s.FrameWords(x.NodeID(), 0)[0] = 0xaaaaaaaa
	s.FrameWords(y.NodeID(), 0)[0] = 0xaaaaaaaa
	s.SimulateOne()
	p := newClassPartition(len(g.Nodes()))
	p.PrepareGraph(g, false, 0, s.Hash, s.IsConstant, s.Equal)
	if !reflect.DeepEqual(p.Classes, [][]int{{x.NodeID(), y.NodeID(), gate.NodeID()}}) {
		t.Fatalf("unexpected initial simulation classes %v", p.Classes)
	}
	s.FrameWords(y.NodeID(), 0)[0] = 0xcccccccc
	s.SimulateOne()
	p.Refine(s.Equal)
	if len(p.Classes) != 0 || p.Reprs[y.NodeID()] != -1 || p.Reprs[gate.NodeID()] != -1 {
		t.Fatalf("simulation failed to split candidates: %+v", p)
	}
}

func TestClassPartitionSequentialBufferFanin(t *testing.T) {
	g := aig.NewGraph()
	input, latch := g.NewInput(), g.NewInput()
	buffer := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{input.Not(), aig.InvalidEdge}, Phase: true})
	gate := g.AND(buffer, latch)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 1)
	p := newClassPartition(len(g.Nodes()))
	p.Reprs[gate.NodeID()] = latch.NodeID()
	unrolled := p.DeriveSequentialAIG(g, 1)
	if unrolled.OutputCount() != 2 {
		t.Fatalf("outputs = %d", unrolled.OutputCount())
	}
	for mask := 0; mask < 8; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := unrolled.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		if out[0] != (!in[0] || !in[1]) || out[1] != (!in[1] || !in[2]) {
			t.Fatalf("inputs %v outputs %v", in, out)
		}
	}
}
