package mf

import "testing"

type testGate struct {
	a, b, c int32
	kind    int
}
type testNetwork struct {
	gates           []testGate
	inputs, outputs []int32
}

func (n *testNetwork) ObjectCount() int        { return len(n.gates) }
func (n *testNetwork) InputLiterals() []int32  { return n.inputs }
func (n *testNetwork) OutputLiterals() []int32 { return n.outputs }
func (n *testNetwork) Gate(id int32) (int32, int32, int32, int) {
	g := n.gates[id]
	return g.a, g.b, g.c, g.kind
}
func TestMappedCutsTruthAndWidths(t *testing.T) {
	// ((a XOR b) ? c : !a), represented by real gates after coarsening.
	n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {kind: 1}, {a: 2, b: 4, kind: 3}, {a: 8, b: 6, c: 3, kind: 4}, {a: 10, kind: 6}}, inputs: []int32{2, 4, 6}, outputs: []int32{10}}
	for width := 3; width <= 8; width++ {
		options := DefaultOptions()
		options.LUTSize = width
		mapping, err := GenerateMapping(n, options)
		if err != nil {
			t.Fatal(err)
		}
		cut := mapping.Cuts[5][0]
		if len(cut.Leaves) != 3 || cut.Leaves[0] != 1 || cut.Leaves[1] != 2 || cut.Leaves[2] != 3 {
			t.Fatal("root cut support", cut.Leaves)
		}
		for assignment := 0; assignment < 8; assignment++ {
			av, bv, cv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
			want := !av
			if av != bv {
				want = cv
			}
			truth := mapping.Truths[cut.Function>>1]
			got := (truth[0]>>uint(assignment)&1)^uint64(cut.Function&1) != 0
			if got != want {
				t.Fatal("root cut truth", width, assignment)
			}
		}
		if mapping.Refs[5] != 1 || mapping.Refs[4] != 0 {
			t.Fatal("mapping should absorb XOR cone", mapping.Refs)
		}
	}
}
func TestMappingRejectsMalformedGraph(t *testing.T) {
	n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {a: 8, b: 2, kind: 2}}, inputs: []int32{2}, outputs: []int32{4}}
	if _, err := GenerateMapping(n, DefaultOptions()); err == nil {
		t.Fatal("forward fanin accepted")
	}
}

func TestMappingRejectsMalformedInterfaces(t *testing.T) {
	for _, n := range []*testNetwork{
		{gates: []testGate{{kind: 0}, {kind: 1}}, inputs: []int32{2}, outputs: []int32{100}},
		{gates: []testGate{{kind: 0}, {kind: 1}}, inputs: []int32{2, 2}, outputs: []int32{2}},
		{gates: []testGate{{kind: 0}, {kind: 1}}, outputs: []int32{2}},
	} {
		if _, err := GenerateMapping(n, DefaultOptions()); err == nil {
			t.Fatal("invalid interface accepted")
		}
	}
}

func TestMappedNetworkOwnershipAndPhase(t *testing.T) {
	n := &testNetwork{gates: []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {a: 3, b: 5, kind: 2}, {a: 6, kind: 6}}, inputs: []int32{2, 4}, outputs: []int32{6}}
	mapping, err := GenerateMapping(n, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	leaves, truth, ok := mapping.LUT(3)
	if !ok || len(leaves) != 2 || truth[0]&15 != 1 {
		t.Fatal("actual complemented truth", leaves, truth, ok)
	}
	leaves[0] = -1
	truth[0] = 0
	again, table, ok := mapping.LUT(3)
	if !ok || again[0] != 1 || table[0]&15 != 1 {
		t.Fatal("LUT views alias canonical storage")
	}
	inputs := mapping.InputLiterals()
	inputs[0] = 0
	if n.inputs[0] != 2 {
		t.Fatal("forwarded interface aliases source")
	}
	if _, _, ok := mapping.LUT(1); ok {
		t.Fatal("CI exposed as selected LUT")
	}
}
