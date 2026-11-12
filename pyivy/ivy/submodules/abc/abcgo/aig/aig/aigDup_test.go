package aig

import (
	"math/rand"
	"testing"
)

func dupFixture() *Graph {
	g := NewGraph()
	g.Name = "dup"
	g.Spec = "fixture"
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	g.AddOutput(g.Mux(c, a.Not(), b))
	g.AND(a, c)
	return g
}
func equalDup(t *testing.T, a, b *Graph) {
	t.Helper()
	if a.InputCount() != b.InputCount() || a.OutputCount() != b.OutputCount() {
		t.Fatal("terminal counts")
	}
	for bits := 0; bits < 1<<a.InputCount(); bits++ {
		inputs := make([]bool, a.InputCount())
		for i := range inputs {
			inputs[i] = bits&(1<<i) != 0
		}
		want, err := a.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := b.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		for i := range want {
			if want[i] != got[i] {
				t.Fatalf("bits %d output %d", bits, i)
			}
		}
	}
	if err := b.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestDuplicateVariants(t *testing.T) {
	for name, fn := range map[string]func(*Graph) *Graph{"simple": Aig_ManDupSimple, "simple dfs": Aig_ManDupSimpleDfs, "ordered": Aig_ManDupOrdered, "dfs": Aig_ManDupDfs, "levelized": Aig_ManDupLevelized, "exor": Aig_ManDupExor, "repr": Aig_ManDupRepres, "repr dfs": Aig_ManDupRepresDfs} {
		t.Run(name, func(t *testing.T) { g := dupFixture(); equalDup(t, g, fn(g)) })
	}
}
func TestDuplicateCofactor(t *testing.T) {
	g := dupFixture()
	for _, v := range []int32{0, 1} {
		q := Aig_ManDupCof(g, 1, v)
		if q.InputCount() != 2 {
			t.Fatal(q.InputCount())
		}
		for bits := 0; bits < 4; bits++ {
			input := []bool{bits&1 != 0, v != 0, bits&2 != 0}
			want, _ := g.Evaluate(input)
			got, err := q.Evaluate([]bool{input[0], input[2]})
			if err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if want[i] != got[i] {
					t.Fatal(v, bits, i)
				}
			}
		}
	}
}
func TestDuplicateHintsAndTrim(t *testing.T) {
	g := dupFixture()
	hints := make([]int32, len(g.nodes))
	for i := range hints {
		hints[i] = -1
	}
	hints[g.inputs[0].NodeID()] = 1
	q := Aig_ManDupSimpleWithHints(g, hints)
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, _ := g.Evaluate([]bool{false, in[1], in[2]})
		got, err := q.Evaluate(in)
		if err != nil {
			t.Fatal(err)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatal(bits, i)
			}
		}
	}
	g = NewGraph()
	a := g.NewInput()
	g.NewInput()
	g.AddOutput(a.Not())
	q = Aig_ManDupTrim(g)
	if q.InputCount() != 1 {
		t.Fatal(q.InputCount())
	}
	out, _ := q.Evaluate([]bool{false})
	if !out[0] {
		t.Fatal(out)
	}
}
func TestDuplicateExplicitOutputsAndPIOs(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	Aig_ObjCreateCo(g, g.AND(a, b))
	Aig_ObjCreateCo(g, a.Not())
	for _, fn := range []func(*Graph) *Graph{Aig_ManDupSimple, Aig_ManDupOrdered, Aig_ManDupSimpleDfs, Aig_ManDupDfs} {
		equalDup(t, g, fn(g))
	}
	q := Aig_ManDupDfsGuided(g, Aig_ManOrderPios(g, g))
	equalDup(t, g, q)
}
func TestDuplicateExorCapturesXor(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	q := Aig_ManDupExor(g)
	equalDup(t, g, q)
	found := false
	for id, n := range q.nodes {
		if !q.deleted[id] && n.Kind == ExclusiveOr {
			found = true
		}
	}
	if !found {
		t.Fatal("no explicit XOR captured")
	}
}
func TestDuplicateMiterAndArray(t *testing.T) {
	a := NewGraph()
	x, y := a.NewInput(), a.NewInput()
	a.AddOutput(a.AND(x, y))
	b := NewGraph()
	x, y = b.NewInput(), b.NewInput()
	b.AddOutput(b.OR(x, y))
	for op := int32(0); op < 4; op++ {
		m := Aig_ManCreateMiter(a, b, op)
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			av, _ := a.Evaluate(in)
			bv, _ := b.Evaluate(in)
			want := false
			switch op {
			case 0:
				want = av[0] != bv[0]
			case 1:
				want = av[0] && !bv[0]
			case 2:
				want = av[0] || bv[0]
			case 3:
				want = av[0] && bv[0]
			}
			got, _ := m.Evaluate(in)
			if got[0] != want {
				t.Fatal(op, bits)
			}
		}
	}
	q := Aig_ManDupArray([]*Graph{a, b})
	if q.OutputCount() != 2 {
		t.Fatal(q.OutputCount())
	}
}
func TestDuplicateSequentialOutputSelections(t *testing.T) {
	g := NewGraph()
	a, latch := g.NewInput(), g.NewInput()
	g.nRegs = 1
	g.AddOutput(a)
	g.AddOutput(g.Const0())
	g.AddOutput(g.AND(a, latch))
	q := Aig_ManDupOneOutput(g, 0, 1)
	if q.nRegs != 1 || q.OutputCount() != 2 {
		t.Fatal(q.nRegs, q.OutputCount())
	}
	q = Aig_ManDupOrpos(g, 0)
	if q.nRegs != 0 || q.OutputCount() != 1 {
		t.Fatal(q.nRegs, q.OutputCount())
	}
	q = Aig_ManDupUnsolvedOutputs(g, 1)
	if q.OutputCount() != 2 {
		t.Fatal(q.OutputCount())
	}
	q = Aig_ManDupFlopsOnly(g)
	if q.OutputCount() != 1 || q.nRegs != 1 {
		t.Fatal(q.OutputCount(), q.nRegs)
	}
}
func TestDuplicatePartialAndNodes(t *testing.T) {
	g := dupFixture()
	q := Aig_ManDupSimpleDfsPart(g, g.inputs[:2], g.outputs[:1])
	if q.InputCount() != 2 || q.OutputCount() != 1 {
		t.Fatal(q.InputCount(), q.OutputCount())
	}
	q = Aig_ManDupNodes(g, g.outputs[:1])
	if q.InputCount() != 2 || q.OutputCount() != 1 {
		t.Fatal(q.InputCount(), q.OutputCount())
	}
	no := Aig_ManDupWithoutPos(g)
	if no.OutputCount() != 0 {
		t.Fatal(no.OutputCount())
	}
}

func TestDuplicateGuidedInterleaving(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	Aig_ObjCreateCo(g, root)
	c := g.NewInput()
	other := g.AND(a, c)
	Aig_ObjCreateCo(g, other)
	g.equivs = map[int]Edge{root.NodeID(): other}
	order := Aig_ManOrderPios(g, g)
	if len(order) != 5 || order[2] != (PIO{false, 0}) || order[3] != (PIO{true, 2}) {
		t.Fatal(order)
	}
	q := Aig_ManDupDfsGuided(g, order)
	equalDup(t, g, q)
}
func TestDuplicateChoicesAndForwardRepresentatives(t *testing.T) {
	for _, reps := range []bool{false, true} {
		g := NewGraph()
		a, b := g.NewInput(), g.NewInput()
		root := g.AND(a, b)
		other := g.NewXor(a, g.AND(a, b.Not()))
		g.AddOutput(root)
		if reps {
			g.reprs = map[int]Edge{root.NodeID(): other}
			equalDup(t, g, Aig_ManDupRepresDfs(g))
			equalDup(t, g, Aig_ManDupRepres(g))
		} else {
			g.equivs = map[int]Edge{root.NodeID(): other}
			q := Aig_ManDupDfs(g)
			equalDup(t, g, q)
			if len(q.equivs) == 0 {
				t.Fatal("lost choices")
			}
		}
	}
}

type dupTiming struct{ value int }

func (t *dupTiming) CloneTiming() any { return &dupTiming{t.value} }

type dupHistory struct{ value int }

func (h *dupHistory) CloneHistory() any { return &dupHistory{h.value} }
func TestDuplicateMetadata(t *testing.T) {
	g := dupFixture()
	g.Assertions = 2
	g.BarrierBuffers = 3
	g.FlopNumbers = []int32{9, 8}
	g.Timing = &dupTiming{7}
	g.History = &dupHistory{6}
	q := Aig_ManDupSimple(g)
	if q.Name != g.Name || q.Spec != g.Spec || q.Assertions != 2 || q.BarrierBuffers != 3 || q.Timing == g.Timing || q.History == g.History {
		t.Fatal("metadata")
	}
	q.FlopNumbers[0] = 0
	if g.FlopNumbers[0] != 9 {
		t.Fatal("shared metadata slice")
	}
}

func TestDuplicateRandomGraphs(t *testing.T) {
	random := rand.New(rand.NewSource(273))
	for sample := 0; sample < 20; sample++ {
		g := NewGraph()
		edges := []Edge{g.Const0(), g.Const1()}
		for i := 0; i < 4; i++ {
			edges = append(edges, g.NewInput())
		}
		for i := 0; i < 25; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			switch random.Intn(3) {
			case 0:
				edges = append(edges, g.AND(a, b))
			case 1:
				edges = append(edges, g.OR(a, b))
			case 2:
				edges = append(edges, g.NewXor(a, b))
			}
		}
		for i := 0; i < 4; i++ {
			e := edges[random.Intn(len(edges))]
			if i%2 == 0 {
				g.AddOutput(e)
			} else {
				Aig_ObjCreateCo(g, e)
			}
		}
		for _, fn := range []func(*Graph) *Graph{Aig_ManDupSimple, Aig_ManDupSimpleDfs, Aig_ManDupOrdered, Aig_ManDupDfs, Aig_ManDupLevelized, Aig_ManDupExor, Aig_ManDupRepresDfs} {
			equalDup(t, g, fn(g))
		}
	}
}
func TestDuplicateGuidedOrderRejectsDuplicates(t *testing.T) {
	g := dupFixture()
	defer func() {
		if recover() == nil {
			t.Fatal("accepted duplicate CI ordering")
		}
	}()
	Aig_ManDupDfsGuided(g, []PIO{{true, 0}, {true, 0}, {true, 1}, {false, 0}, {false, 1}})
}

func TestDuplicateClockAndFlopRepresentativeMetadata(t *testing.T) {
	g := dupFixture()
	g.ClockDomains = [][]int32{{0, 1}, {2}}
	g.FlopReprs = []int32{3, -1}
	q := Aig_ManDupSimple(g)
	q.ClockDomains[0][0] = 99
	q.FlopReprs[0] = 99
	if g.ClockDomains[0][0] != 0 || g.FlopReprs[0] != 3 {
		t.Fatal("aliased sequential metadata")
	}
}

func TestDuplicateEmptySequentialMetadata(t *testing.T) {
	for _, nonNil := range []bool{false, true} {
		g := dupFixture()
		if nonNil {
			g.FlopNumbers = []int32{}
			g.FlopReprs = []int32{}
			g.ClockDomains = [][]int32{{}, nil}
		}
		q := Aig_ManDupSimple(g)
		if (q.FlopNumbers != nil) != nonNil || (q.FlopReprs != nil) != nonNil || (q.ClockDomains != nil) != nonNil {
			t.Fatal("metadata presence changed", nonNil)
		}
		if nonNil && (q.ClockDomains[0] == nil || q.ClockDomains[1] != nil) {
			t.Fatal("nested metadata presence changed")
		}
	}
}

func TestDuplicatePartitionMaps(t *testing.T) {
	g := dupFixture()
	g.PartitionInputs = []int32{5, 3}
	g.PartitionOutputs = []int32{7}
	q := Aig_ManDupSimple(g)
	q.PartitionInputs[0] = 0
	q.PartitionOutputs[0] = 0
	if g.PartitionInputs[0] != 5 || g.PartitionOutputs[0] != 7 {
		t.Fatal("aliased partition maps")
	}
}
