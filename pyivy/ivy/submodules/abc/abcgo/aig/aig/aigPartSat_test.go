package aig

import (
	"errors"
	"math/rand"
	"testing"
)

func exhaustivePartSAT(g *Graph) (bool, []bool) {
	for assignment := 0; assignment < 1<<len(g.inputs); assignment++ {
		in := make([]bool, len(g.inputs))
		for i := range in {
			in[i] = assignment&(1<<i) != 0
		}
		out, err := g.Evaluate(in)
		if err != nil {
			panic(err)
		}
		yes := true
		for _, v := range out {
			yes = yes && v
		}
		if yes {
			return true, in
		}
	}
	return false, nil
}
func checkPartitionedSAT(t *testing.T, g *Graph) {
	t.Helper()
	want, _ := exhaustivePartSAT(g)
	for algo := int32(0); algo < 4; algo++ {
		for _, aligned := range []bool{false, true} {
			status, err := PartitionedSAT(g, PartSATOptions{Algorithm: algo, PartitionSize: 1, AlignPolarity: aligned})
			if err != nil {
				t.Fatal(err)
			}
			if (status == 0) != want || status < 0 {
				t.Fatalf("algorithm=%d aligned=%v got %d want SAT=%v", algo, aligned, status, want)
			}
			if status == 0 {
				model, ok := g.Data.([]int32)
				if !ok || len(model) != len(g.inputs) {
					t.Fatal("missing model", g.Data)
				}
				in := make([]bool, len(model))
				for i, v := range model {
					in[i] = v != 0
				}
				out, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				for i, v := range out {
					if !v {
						t.Fatalf("invalid model output %d", i)
					}
				}
			}
		}
	}
}
func TestPartitionedSATDegenerate(t *testing.T) {
	for _, kind := range []string{"empty", "true", "false", "pi", "negative pi", "conflicting pis", "isolated pi with gates"} {
		t.Run(kind, func(t *testing.T) {
			g := NewGraph()
			switch kind {
			case "true":
				g.AddOutput(g.Const1())
			case "false":
				g.AddOutput(g.Const0())
			case "pi":
				g.AddOutput(g.NewInput())
			case "negative pi":
				g.AddOutput(g.NewInput().Not())
			case "conflicting pis":
				a := g.NewInput()
				g.AddOutput(a)
				g.AddOutput(a.Not())
			case "isolated pi with gates":
				a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
				g.AddOutput(a)
				g.AddOutput(g.XOR(b, c))
			}
			checkPartitionedSAT(t, g)
		})
	}
}
func TestPartitionedSATRandomGraphs(t *testing.T) {
	r := rand.New(rand.NewSource(861))
	for sample := 0; sample < 24; sample++ {
		g := NewGraph()
		edges := []Edge{g.Const0(), g.Const1()}
		for i := 0; i < 4; i++ {
			edges = append(edges, g.NewInput())
		}
		for i := 0; i < 15; i++ {
			a, b := edges[r.Intn(len(edges))], edges[r.Intn(len(edges))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			var e Edge
			switch r.Intn(3) {
			case 0:
				e = g.AND(a, b)
			case 1:
				e = g.OR(a, b)
			case 2:
				e = g.NewXor(a, b)
			}
			edges = append(edges, e)
		}
		for i := 0; i < 3; i++ {
			g.AddOutput(edges[r.Intn(len(edges))])
		}
		checkPartitionedSAT(t, g)
	}
}
func TestPartitionSplitSparseLabelsAndMappings(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	y := g.NewXor(x, c)
	g.AddOutput(y.Not())
	labels := make([]int32, len(g.nodes))
	labels[x.NodeID()] = 7
	labels[y.NodeID()] = 91
	parts, pio, outputs := Aig_ManPartSplit(g, labels)
	if len(parts) != 2 || len(pio) != 2 || len(outputs) != 2 || len(outputs[1]) != 1 {
		t.Fatal(len(parts), pio, outputs)
	}
	for i, part := range parts {
		if len(pio[i]) != part.InputCount()+part.OutputCount() {
			t.Fatal("PIO count")
		}
		if err := part.Check(); err != nil {
			t.Fatal(err)
		}
	}
}

type testPartSynth struct {
	calls int
	fail  bool
}

func (s *testPartSynth) SynthesizePart(g *Graph) (*Graph, error) {
	s.calls++
	if s.fail {
		return nil, errors.New("test synthesis failure")
	}
	return Aig_ManDupSimple(g), nil
}
func TestPartitionSATInjectedSynthesis(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	if status, err := PartitionedSAT(g, PartSATOptions{Synthesize: true}); status != -1 || err == nil {
		t.Fatal(status, err)
	}
	engine := new(testPartSynth)
	status, err := PartitionedSAT(g, PartSATOptions{Synthesize: true, Synthesizer: engine})
	if err != nil || status != 0 || engine.calls == 0 {
		t.Fatal(status, err, engine.calls)
	}
	engine.fail = true
	if status, err = PartitionedSAT(g, PartSATOptions{Synthesize: true, Synthesizer: engine}); status != -1 || err == nil {
		t.Fatal(status, err)
	}
}
func TestPartitionSATInvalidOptions(t *testing.T) {
	for _, options := range []PartSATOptions{{Algorithm: 4}, {Algorithm: 1}, {PartConflictLimit: -1}, {TotalConflictLimit: -1}} {
		if status, err := PartitionedSAT(NewGraph(), options); status != -1 || err == nil {
			t.Fatal(options, status, err)
		}
	}
}

func TestPartitionSATPreservesClientMarks(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.AddOutput(root)
	g.node(a).MarkA = true
	g.node(b).MarkB = true
	g.node(root).MarkA = true
	checkPartitionedSAT(t, g)
	if !g.node(a).MarkA || !g.node(b).MarkB || !g.node(root).MarkA {
		t.Fatal("client marks changed")
	}
}
func TestPartitionSATOriginalConstantBoundary(t *testing.T) {
	for _, negative := range []bool{false, true} {
		g := NewGraph()
		a := g.NewInput()
		node := Aig_ObjCreate(g, &Node{Kind: ExclusiveOr, Fanins: [2]Edge{g.Const1(), a}})
		if negative {
			node = node.Not()
		}
		g.AddOutput(node)
		checkPartitionedSAT(t, g)
	}
}
func TestPartitionSATInputPhaseAndCounterexample(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.node(a).Phase = true
	g.node(b).Phase = true
	g.AddOutput(g.AND(a.Not(), b))
	g.recomputeLevels()
	checkPartitionedSAT(t, g)
}
func TestPartitionSATBoundedAnswersSound(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	for _, e := range []Edge{g.OR(a, b), g.OR(a, b.Not()), g.OR(a.Not(), b), g.OR(a.Not(), b.Not())} {
		g.AddOutput(e)
	}
	for algorithm := int32(0); algorithm < 4; algorithm++ {
		status, err := PartitionedSAT(g, PartSATOptions{Algorithm: algorithm, PartitionSize: 1, PartConflictLimit: 1, TotalConflictLimit: 2})
		if err != nil || status == 0 {
			t.Fatal("bounded UNSAT graph reported SAT", algorithm, status, err)
		}
		status, err = PartitionedSAT(g, PartSATOptions{Algorithm: algorithm, PartitionSize: 1})
		if err != nil || status != 1 {
			t.Fatal("unlimited solve failed", algorithm, status, err)
		}
	}
}

func TestPartitionSATTransparentOutputFanins(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(g.AND(g.outputs[0], b.Not()))
	g.AddOutput(g.outputs[1])
	checkPartitionedSAT(t, g)
}
func TestPartitionSATInputContradictionThroughCO(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(g.AND(g.outputs[0].Not(), b))
	checkPartitionedSAT(t, g)
}

type mutatingPartSynth struct{}

func (mutatingPartSynth) SynthesizePart(g *Graph) (*Graph, error) { g.NewInput(); return g, nil }
func TestPartitionSATRejectsMutatingInterface(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	if status, err := PartitionedSAT(g, PartSATOptions{Synthesize: true, Synthesizer: mutatingPartSynth{}}); status != -1 || err == nil {
		t.Fatal(status, err)
	}
}
func TestPartitionSplitRejectsDuplicates(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	g.AddOutput(root)
	defer func() {
		if recover() == nil {
			t.Fatal("accepted duplicate partition node")
		}
	}()
	Aig_ManPartSplitOne(g, []Edge{root, root})
}
