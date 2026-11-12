package deco

import "testing"

func TestEdgeBitfieldNormalization(t *testing.T) {
	for _, node := range []int32{0, 1, 12345, 1<<30 - 1, 1 << 30, -1, -1 << 31} {
		for _, phase := range []int32{-3, -1, 0, 1, 2, 7} {
			edge := Dec_EdgeCreate(node, phase)
			want := (uint32(node)&0x3fffffff)<<1 | uint32(phase)&1
			if Dec_EdgeToInt(edge) != want || Dec_EdgeToInt_(edge) != want {
				t.Fatal(node, phase, edge, want)
			}
		}
	}
	for _, value := range []uint32{0, 1, 2, 0x7fffffff, 0x80000000, 0xffffffff} {
		if got := Dec_EdgeToInt(Dec_IntToEdge(value)); got != value&0x7fffffff {
			t.Fatal(value, got)
		}
		if got := Dec_EdgeToInt_(Dec_IntToEdge_(value)); got != value&0x7fffffff {
			t.Fatal(value, got)
		}
	}
}
func TestBooleanGraphConstructions(t *testing.T) {
	for kind := int32(0); kind < 2; kind++ {
		for phase := int32(0); phase < 8; phase++ {
			for op := 0; op < 4; op++ {
				g := Dec_GraphCreate(3)
				a, b, c := Dec_EdgeCreate(0, phase&1), Dec_EdgeCreate(1, phase>>1&1), Dec_EdgeCreate(2, phase>>2&1)
				var root Dec_Edge_t
				switch op {
				case 0:
					root = Dec_GraphAddNodeAnd(g, a, b)
				case 1:
					root = Dec_GraphAddNodeOr(g, a, b)
				case 2:
					root = Dec_GraphAddNodeXor(g, a, b, kind)
				case 3:
					root = Dec_GraphAddNodeMux(g, c, a, b, kind)
				}
				Dec_GraphSetRoot(g, root)
				truth := g.Native().Truth32()
				for assignment := 0; assignment < 32; assignment++ {
					av, bv, cv := (assignment&1 != 0) != (phase&1 != 0), (assignment&2 != 0) != (phase&2 != 0), (assignment&4 != 0) != (phase&4 != 0)
					want := false
					switch op {
					case 0:
						want = av && bv
					case 1:
						want = av || bv
					case 2:
						want = av != bv
					case 3:
						want = cv && av || !cv && bv
					}
					if (truth>>uint(assignment)&1 != 0) != want {
						t.Fatal(kind, phase, op, assignment)
					}
				}
				Dec_GraphComplement(g)
				if got := g.Native().Truth32(); got != ^truth {
					t.Fatal("root complement")
				}
			}
		}
	}
}
func TestGraphConstantsLeavesIteratorsAndGrowth(t *testing.T) {
	if Dec_GraphIsConst0(Dec_GraphCreateConst0()) != 1 || Dec_GraphIsConst1(Dec_GraphCreateConst1()) != 1 {
		t.Fatal("constants")
	}
	g := Dec_GraphCreateLeaf(1, 3, 1)
	if Dec_GraphIsVar(g) != 1 || Dec_GraphVarInt(g) != 1 || Dec_GraphIsComplement(g) != 1 {
		t.Fatal("leaf")
	}
	leaf := Dec_GraphVar(g)
	if Dec_GraphNodeIsVar(g, leaf) != 1 || Dec_GraphNodeInt(g, leaf) != 1 {
		t.Fatal("node identity")
	}
	root := Dec_EdgeCreate(0, 0)
	for i := 0; i < 200; i++ {
		root = Dec_GraphAddNodeAnd(g, root, Dec_EdgeCreate(1, 0))
	}
	if Dec_GraphNode(g, 1) != leaf || Dec_GraphNodeNum(g) != 200 {
		t.Fatal("stable allocation")
	}
	if Dec_GraphNodeLast(g) != Dec_GraphNode(g, 202) {
		t.Fatal("last node")
	}
	n := 0
	for i, node := range Dec_GraphForEachLeaf(g) {
		if node != Dec_GraphNode(g, i) {
			t.Fatal(i)
		}
		n++
	}
	if n != 3 {
		t.Fatal(n)
	}
	n = 0
	for range Dec_GraphForEachNode(g) {
		n++
		if n == 4 {
			break
		}
	}
	if n != 4 {
		t.Fatal("early break", n)
	}
	node := Dec_GraphAppendNode(g)
	if Dec_GraphNodeInt(g, node) != 203 {
		t.Fatal("append")
	}
	Dec_NodeSetLevel(node, 0xffffffff)
	if Dec_NodeLevel(node) != 0x3fff {
		t.Fatal("level truncation")
	}
	Dec_GraphFree(g)
}
func TestGraphVarRejectsInternalRoot(t *testing.T) {
	g := Dec_GraphCreate(2)
	Dec_GraphSetRoot(g, Dec_GraphAddNodeAnd(g, Dec_EdgeCreate(0, 0), Dec_EdgeCreate(1, 0)))
	defer func() {
		if recover() == nil {
			t.Fatal("internal root accepted as variable")
		}
	}()
	Dec_GraphVar(g)
}

func TestEdgeBoundaryNormalization(t *testing.T) {
	g := Dec_GraphCreate(2)
	edge := Dec_Edge_t{Node: 0x40000000}
	root := Dec_GraphAddNodeAnd(g, edge, Dec_EdgeCreate(1, 0))
	Dec_GraphSetRoot(g, root)
	if got := g.Native().Truth32(); got != 0x88888888 {
		t.Fatal(got)
	}
	Dec_GraphSetRoot(g, Dec_Edge_t{Node: 0x40000001})
	if Dec_GraphVarInt(g) != 1 {
		t.Fatal("root was not normalized")
	}
}
func TestGraphVarIntRejectsConstant(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("constant accepted as variable")
		}
	}()
	Dec_GraphVarInt(Dec_GraphCreateConst0())
}
func TestManagerHeaderMetadata(t *testing.T) {
	m := Dec_Man_t{FactoringMemory: new(int), Cubes: []int32{1}, Literals: []int32{2}, CanonicalTruths: []uint16{3}, Phases: []byte{4}, Permutations: []byte{5}, Classes: []byte{6}}
	if m.FactoringMemory == nil || m.Cubes[0] != 1 || m.Literals[0] != 2 || m.CanonicalTruths[0] != 3 || m.Phases[0] != 4 || m.Permutations[0] != 5 || m.Classes[0] != 6 {
		t.Fatal("manager metadata")
	}
}

func TestNodeMetadataBitfieldWidths(t *testing.T) {
	g := Dec_GraphCreate(1)
	n := Dec_GraphNode(g, 0)
	n.SetLatches([3]uint32{32, 63, 255})
	if got := n.Latches(); got != [3]uint32{0, 31, 31} {
		t.Fatal(got)
	}
	n.SetPrintingFlags([3]uint32{2, 3, 255})
	if got := n.PrintingFlags(); got != [3]uint32{0, 1, 1} {
		t.Fatal(got)
	}
	n.SetLevel(1<<14 + 3)
	if n.Level != 3 {
		t.Fatal(n.Level)
	}
}
