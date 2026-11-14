package aig

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/ioa"
)

// Interpret AIGER's false-at-zero literals independently of Graph.Evaluate.
func auditAigerValues(t *testing.T, m *ioa.Model, inputs []bool) []bool {
	t.Helper()
	values := make([]bool, m.MaxVar+1)
	for i, lit := range m.Inputs {
		values[lit>>1] = inputs[i]
	}
	for i, latch := range m.Latches {
		values[latch.Literal>>1] = inputs[len(m.Inputs)+i]
	}
	value := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
	for _, gate := range m.ANDs {
		if gate.Left >= gate.Literal || gate.Right >= gate.Literal {
			t.Fatal("snapshot gates are not topological")
		}
		values[gate.Literal>>1] = value(gate.Left) && value(gate.Right)
	}
	var outputs []bool
	for _, lit := range m.Outputs {
		outputs = append(outputs, value(lit))
	}
	for _, lit := range m.Bad {
		outputs = append(outputs, value(lit))
	}
	for _, lit := range m.Constraints {
		outputs = append(outputs, !value(lit))
	}
	for _, latch := range m.Latches {
		outputs = append(outputs, value(latch.Next))
	}
	return outputs
}

func TestAigerSnapshotMixedGraphs(t *testing.T) {
	rng := rand.New(rand.NewPCG(141, 592))
	for trial := 0; trial < 128; trial++ {
		g := NewGraph()
		g.Name = "snapshot"
		inputs := []Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		early := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{g.Const1(), InvalidEdge}})
		pool := append([]Edge{g.Const1(), g.Const0()}, inputs...)
		for step := 0; step < 16; step++ {
			a := dupPhase(pool[rng.IntN(len(pool))], rng.IntN(2) != 0)
			b := dupPhase(pool[rng.IntN(len(pool))], rng.IntN(2) != 0)
			var e Edge
			switch step % 3 {
			case 0:
				e = g.AND(a, b)
			case 1:
				e = g.NewXor(a, b)
			case 2:
				e = Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
			}
			pool = append(pool, e)
		}
		Aig_ObjConnect(g, early, pool[len(pool)-1].Not(), InvalidEdge)
		first := g.AddOutput(early)
		// A physical CO may itself be a fanin after rewriting.
		g.AddOutput(g.NewXor(first, inputs[0]))
		for i := 0; i < 3; i++ {
			g.AddOutput(pool[rng.IntN(len(pool))])
		}
		regs := int32(trial % 3)
		Aig_ManSetRegNum(g, regs)
		inOrder := append([]Edge(nil), g.inputs...)
		rng.Shuffle(len(inOrder), func(i, j int) { inOrder[i], inOrder[j] = inOrder[j], inOrder[i] })
		outOrder := append([]Edge(nil), g.outputs...)
		rng.Shuffle(len(outOrder), func(i, j int) { outOrder[i], outOrder[j] = outOrder[j], outOrder[i] })
		if err := Aig_ManReorderCisCos(g, inOrder, outOrder, regs); err != nil {
			t.Fatal(err)
		}
		g.nConstrs = int32(trial % (len(g.outputs) - int(regs) + 1))
		Aig_ManCleanup(g)
		Aig_ManFanoutStart(g)
		g.copies = map[int]Edge{0: 0, inputs[0].NodeID(): inputs[1].Not()}
		for id := range g.nodes {
			g.nodes[id].TravID = 17
			g.nodes[id].MarkA = id%2 == 0
		}
		nodes := append([]Node(nil), g.nodes...)
		copies := make(map[int]Edge)
		for k, v := range g.copies {
			copies[k] = v
		}
		model, err := g.AigerModel()
		if err != nil {
			t.Fatalf("trial %d: %v", trial, err)
		}
		if !reflect.DeepEqual(nodes, g.nodes) || !reflect.DeepEqual(copies, g.copies) || !reflect.DeepEqual(inOrder, g.inputs) || !reflect.DeepEqual(outOrder, g.outputs) {
			t.Fatalf("trial %d: snapshot mutated source graph", trial)
		}
		if model.Name != g.Name || len(model.Latches) != int(regs) || len(model.Constraints) != int(g.nConstrs) {
			t.Fatal("snapshot metadata changed")
		}
		encoded, err := ioa.Marshal(model, ioa.WriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := ioa.Read(encoded)
		if err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 16; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			want, err := g.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range []*ioa.Model{model, decoded} {
				got := auditAigerValues(t, m, in)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("trial %d assignment %d: got %v want %v", trial, mask, got, want)
				}
			}
		}
	}
}
