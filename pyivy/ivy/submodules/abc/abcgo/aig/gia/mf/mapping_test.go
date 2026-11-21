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
		if mapping.truthResults != nil || mapping.truthStore != nil || len(mapping.Truths) != len(mapping.ClauseSizes) {
			t.Fatal("finished mapping retained its hash index or lost a function")
		}
		for id, truth := range mapping.Truths {
			words := 1 << max(0, width-6)
			if len(truth) != words || cap(truth) != words {
				t.Fatalf("function %d at LUT width %d has length %d, capacity %d", id, width, len(truth), cap(truth))
			}
		}
		cut := mapping.Cuts[5][0]
		if len(cut.Leaves) != 3 || cut.Leaves[0] != 1 || cut.Leaves[1] != 2 || cut.Leaves[2] != 3 {
			t.Fatal("root cut support", cut.Leaves)
		}
		if len(mapping.Cuts[5]) != 2 {
			t.Fatalf("expected separate cuts for the XOR cone and its control node at width %d", width)
		}
		for assignment := 0; assignment < 8; assignment++ {
			av, bv, cv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
			want := !av
			if av != bv {
				want = cv
			}
			values := [...]bool{false, av, bv, cv, av != bv}
			for _, candidate := range mapping.Cuts[5] {
				index := 0
				for bit, leaf := range candidate.Leaves {
					if values[leaf] {
						index |= 1 << bit
					}
				}
				truth := mapping.Truths[candidate.Function>>1]
				got := (truth[0]>>uint(index)&1)^uint64(candidate.Function&1) != 0
				if got != want {
					t.Fatal("root cut truth", width, assignment, candidate.Leaves)
				}
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

func TestExactCutTrialsReuseScratchAndRestoreReferences(t *testing.T) {
	// The trial visits 62 fanins in this five-level binary tree.
	const root = 62
	m := &Mapping{nodes: make([]node, root+1), Cuts: make([][]Cut, root+1), ClauseSizes: []int{2}}
	for id := 32; id <= root; id++ {
		left := int32(2 * (id - 32))
		m.Cuts[id] = []Cut{{Leaves: []int32{left, left + 1}}}
		m.nodes[id].flowRefs = 1
	}
	if scratch := m.bestCut(root, false, nil); scratch != nil {
		t.Fatal("area-flow selection reserved exact-area scratch")
	}
	var scratch []int32
	if allocs := testing.AllocsPerRun(20, func() { scratch = m.bestCut(root, true, scratch) }); allocs != 0 {
		t.Fatalf("repeated exact cut trial allocated %.0f times", allocs)
	}
	if len(scratch) != 0 || cap(scratch) < 62 {
		t.Fatalf("returned scratch has length %d and capacity %d", len(scratch), cap(scratch))
	}
	for id, n := range m.nodes {
		if n.refs != 0 {
			t.Fatalf("trial left %d references at node %d", n.refs, id)
		}
	}
}

func TestChoiceCutOwnsItsSavedLeaves(t *testing.T) {
	n := &siblingTestNetwork{
		testNetwork: testNetwork{
			gates:  []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {a: 2, b: 4, kind: 2}, {a: 2, b: 4, kind: 2}},
			inputs: []int32{2, 4}, outputs: []int32{8},
		},
		sibs: map[int32]int32{4: 3},
	}
	m, err := GenerateMapping(n, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	first, second := m.Cuts[3][0], m.Cuts[4][0]
	if len(first.Leaves) != 2 || len(second.Leaves) != 2 || first.Function != second.Function {
		t.Fatal("expected identical two-input choice cuts")
	}
	first.Leaves[0] = 0
	if second.Leaves[0] != 1 {
		t.Fatal("editing the sibling's saved leaves changed the choice cut")
	}
}

func TestMinimizedCutPrunesLaterSupersets(t *testing.T) {
	// a AND (a OR b) reduces to a. Its signature must lose b before it
	// is compared with the later candidate whose leaves are a and gate 3.
	n := &testNetwork{
		gates:  []testGate{{kind: 0}, {kind: 1}, {kind: 1}, {a: 3, b: 5, kind: 2}, {a: 2, b: 7, kind: 2}},
		inputs: []int32{2, 4}, outputs: []int32{8},
	}
	m, err := GenerateMapping(n, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	cuts := m.Cuts[4]
	if len(cuts) != 1 || len(cuts[0].Leaves) != 1 || cuts[0].Leaves[0] != 1 {
		t.Fatal("expected only the minimal one-leaf cut", cuts)
	}
}
