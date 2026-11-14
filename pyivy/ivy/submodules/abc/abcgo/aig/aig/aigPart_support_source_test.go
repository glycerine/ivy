package aig

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"
)

func TestWholeSupportsDoNotRunConeTraversals(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	shared := g.AND(a, b)
	g.AddOutput(shared)
	g.AddOutput(g.AND(shared, c))
	g.AddOutput(g.Const0())
	for id := range g.nodes {
		g.nodes[id].TravID = int32(31 + id)
	}
	before := g.Nodes()
	rows := Aig_ManSupports(g)
	if !reflect.DeepEqual(rows, [][]int32{{0, 1, 2, 1}, {0, 1, 0}, {2}}) {
		t.Fatalf("support rows %v", rows)
	}
	if g.travID != 1 {
		t.Errorf("support propagation started %d extra traversals", g.travID-1)
	}
	for id := range g.nodes {
		if g.nodes[id].TravID != before[id].TravID || g.nodes[id].Refs != before[id].Refs {
			t.Errorf("support propagation changed graph scratch/reference state at %d", id)
		}
	}
}

func TestWholeSupportsForwardUnaryAndTypedXor(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{g.Const1(), InvalidEdge}})
	xor := g.NewXor(a, b)
	Aig_ObjConnect(g, buffer, xor.Not(), InvalidEdge)
	co := g.AddOutput(buffer)
	g.AddOutput(g.AND(co, c).Not())
	g.AddOutput(g.Const0())
	g.objectData = map[int]any{a.NodeID(): "a", buffer.NodeID(): "saved"}
	g.copies = map[int]Edge{a.NodeID(): c}
	got := Aig_ManSupports(g)
	want := [][]int32{{0, 1, 2, 1}, {0, 1, 0}, {2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("typed/forward supports %v, want %v", got, want)
	}
	if g.objectData[a.NodeID()] != "a" || g.objectData[buffer.NodeID()] != "saved" || g.copies[a.NodeID()] != c {
		t.Fatal("local support records overwrote caller payloads/copies")
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestWholeSupportsAgainstCompiledPropagation(t *testing.T) {
	rng := rand.New(rand.NewPCG(43, 59))
	var input, want strings.Builder
	for trial := 0; trial < 256; trial++ {
		g := NewGraph()
		available := []Edge{g.Const1()}
		for i := 0; i < 5; i++ {
			available = append(available, g.NewInput())
		}
		for i := 0; i < 30; i++ {
			a, b := available[rng.IntN(len(available))], available[rng.IntN(len(available))]
			if rng.IntN(2) == 0 {
				a = a.Not()
			}
			root := g.AND(a, b)
			if i%3 == 0 {
				root = g.NewXor(a, b)
			}
			available = append(available, root)
		}
		for i := 0; i < 10; i++ {
			g.AddOutput(available[rng.IntN(len(available))])
		}
		regs := trial % 6
		fmt.Fprintf(&input, "%d %d\n", len(g.nodes), regs)
		for _, node := range g.nodes[1:] {
			a, b := int64(-1), int64(-1)
			pins := nodeFanins(node)
			if len(pins) > 0 {
				a = int64(pins[0])
			}
			if len(pins) > 1 {
				b = int64(pins[1])
			}
			fmt.Fprintf(&input, "%d %d %d\n", node.Kind, a, b)
		}
		rng.Shuffle(len(g.inputs), func(i, j int) { g.inputs[i], g.inputs[j] = g.inputs[j], g.inputs[i] })
		rng.Shuffle(len(g.outputs), func(i, j int) { g.outputs[i], g.outputs[j] = g.outputs[j], g.outputs[i] })
		for _, ci := range g.inputs {
			fmt.Fprintf(&input, "%d ", ci.NodeID())
		}
		input.WriteByte('\n')
		for _, co := range g.outputs {
			fmt.Fprintf(&input, "%d ", co.NodeID())
		}
		input.WriteByte('\n')
		Aig_ManSetRegNum(g, int32(regs))
		for id := range g.nodes {
			g.nodes[id].TravID = int32(71 + id)
		}
		rows := Aig_ManSupports(g)
		byOutput := make([][]int32, len(g.outputs))
		for i, row := range rows {
			if i > 0 && len(rows[i-1]) < len(row) {
				t.Fatal("support rows are not sorted by decreasing size")
			}
			byOutput[row[len(row)-1]] = row[:len(row)-1]
		}
		for i, row := range byOutput {
			fmt.Fprintf(&want, "row %d:", i)
			for _, ci := range row {
				fmt.Fprintf(&want, " %d", ci)
			}
			want.WriteByte('\n')
		}
		for i, row := range Aig_ManSupportsRegisters(g) {
			fmt.Fprintf(&want, "reg %d:", i)
			for _, input := range row {
				fmt.Fprintf(&want, " %d", input)
			}
			want.WriteByte('\n')
		}
		want.WriteString(captureAigOutput(t, func() { Aig_ManComputeSccs(g) }))
		fmt.Fprintf(&want, "state %d:", g.travID)
		for _, node := range g.nodes {
			fmt.Fprintf(&want, " %d/%d", node.TravID, node.Refs)
		}
		want.WriteString(" | ci")
		for _, ci := range g.inputs {
			fmt.Fprintf(&want, " %d", g.node(ci).CioID)
		}
		want.WriteString(" | co")
		for _, co := range g.outputs {
			fmt.Fprintf(&want, " %d", g.node(co).CioID)
		}
		want.WriteByte('\n')
		if err := g.Check(); err != nil {
			t.Fatal(err)
		}
	}
	got := runAigSourceFixtureInput(t, "support_propagation_compare.c", input.String(), "aigMan.c", "aigObj.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "aigPart.c", "aigScl.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	if got != want.String() {
		c, goLines := strings.Split(got, "\n"), strings.Split(want.String(), "\n")
		for i := 0; i < len(c) && i < len(goLines); i++ {
			if c[i] != goLines[i] {
				t.Fatalf("support trace %d: C %q, Go %q", i, c[i], goLines[i])
			}
		}
		t.Fatalf("support trace lengths C %d, Go %d", len(c), len(goLines))
	}
}

func TestWholeSupportsUseInputPositionsForRawOutputs(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	g.outputs = []Edge{gate.Not(), c, g.Const0()}
	g.recountReferences()
	got := Aig_ManSupports(g)
	want := [][]int32{{0, 1, 0}, {2, 1}, {2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("raw output supports = %v, want %v", got, want)
	}
}

func TestWholeSupportsWideSharedConeOwnsRows(t *testing.T) {
	g := NewGraph()
	inputs := make([]Edge, 128)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	root := Aig_Multi(g, inputs, int32(len(inputs)), 5)
	for i := 0; i < 100; i++ {
		g.AddOutput(root)
	}
	rows := Aig_ManSupports(g)
	for output, row := range rows {
		if len(row) != 129 || row[128] != int32(output) {
			t.Fatalf("output %d: support dimensions/index %v", output, row)
		}
		for input := 0; input < 128; input++ {
			if row[input] != int32(input) {
				t.Fatalf("output %d: support position %d is %d", output, input, row[input])
			}
		}
	}
	rows[0][0] = -1
	if rows[1][0] != 0 || Aig_ManSupports(g)[0][0] != 0 {
		t.Fatal("returned rows alias pooled records, one another or a later result")
	}
}

func TestRegisterComponentsUsesSourceSupportScratch(t *testing.T) {
	g := NewGraph()
	pi, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	g.AddOutput(g.AND(pi, a))
	Aig_ManSetRegNum(g, 2)
	before := g.travID
	RegisterComponents(g)
	if g.travID != before {
		t.Errorf("register grouping changed traversal generation from %d to %d", before, g.travID)
	}
	for _, e := range append(g.Inputs(), g.Outputs()...) {
		if g.node(e).CioID != 0 {
			t.Errorf("support propagation did not clear terminal %d's union", e)
		}
	}
}

func TestRegisterSupportsAfterWholeSupportCall(t *testing.T) {
	g := NewGraph()
	pi, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	g.AddOutput(g.AND(pi, a))
	Aig_ManSetRegNum(g, 2)
	Aig_ManSupports(g) // Clears the terminal union, as any prior caller may do.
	before := g.travID
	got := Aig_ManSupportsRegisters(g)
	want := [][]int32{{1}, {0}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("register supports %v, want %v", got, want)
	}
	if g.travID != before {
		t.Error("register supports changed traversal generation")
	}
}
