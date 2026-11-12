package dec

import "testing"

func TestGraphTruth(t *testing.T) {
	for _, kind := range []string{"and", "or", "xor", "mux"} {
		g := NewGraph(3)
		a, b, c := Edge{Node: 0}, Edge{Node: 1}, Edge{Node: 2}
		var root Edge
		switch kind {
		case "and":
			root = g.And(a, b)
		case "or":
			root = g.Or(a, b)
		case "xor":
			root = g.Xor(a, b)
		case "mux":
			root = g.Mux(a, b, c)
		}
		g.SetRoot(root)
		for bits := 0; bits < 8; bits++ {
			x, y, z := bits&1 != 0, bits&2 != 0, bits&4 != 0
			want := false
			switch kind {
			case "and":
				want = x && y
			case "or":
				want = x || y
			case "xor":
				want = x != y
			case "mux":
				if x {
					want = y
				} else {
					want = z
				}
			}
			if got := g.Evaluate([]bool{x, y, z}); got != want {
				t.Fatalf("%s %d: %v", kind, bits, got)
			}
		}
	}
	g := NewGraph(2)
	g.SetRoot(g.And(Edge{Node: 0}, Edge{Node: 1}))
	if got := g.Truth32(); got != 0x88888888 {
		t.Fatalf("%08x", got)
	}
	g.Complement()
	if got := g.Truth32(); got != 0x77777777 {
		t.Fatalf("%08x", got)
	}
	for _, v := range []bool{false, true} {
		g := NewConstant(v)
		if g.Evaluate(nil) != v {
			t.Fatal(v)
		}
	}
}
func TestEdgeLiteral(t *testing.T) {
	for _, v := range []uint32{0, 1, 2, 3, 123456, 0x7fffffff} {
		if EdgeFromLiteral(v).Literal() != v {
			t.Fatal(v)
		}
	}
}

func TestLegacyGraphAPI(t *testing.T) {
	for _, kind := range []int32{0, 1} {
		g := Dec_GraphCreate(3)
		a, b, c := Dec_EdgeCreate(0, 0), Dec_EdgeCreate(1, 0), Dec_EdgeCreate(2, 0)
		e := Dec_GraphAddNodeMux(g, a, b, c, kind)
		Dec_GraphSetRoot(g, e)
		if got := Dec_GraphDeriveTruth(g); got != 0xd8d8d8d8 {
			t.Fatalf("mux %d: %08x", kind, got)
		}
		g = Dec_GraphCreate(2)
		Dec_GraphSetRoot(g, Dec_GraphAddNodeXor(g, a, b, kind))
		if got := Dec_GraphDeriveTruth(g); got != 0x66666666 {
			t.Fatalf("xor %d: %08x", kind, got)
		}
	}
	g := Dec_GraphCreate(1)
	leaf := Dec_GraphNode(g, 0)
	for i := 0; i < 200; i++ {
		Dec_GraphAddNodeAnd(g, Dec_EdgeCreate(0, 0), Dec_EdgeCreate(0, 0))
	}
	if leaf != Dec_GraphNode(g, 0) {
		t.Fatal("unstable node")
	}
	if Dec_GraphDeriveTruth(Dec_GraphCreateConst0()) != 0 || Dec_GraphDeriveTruth(Dec_GraphCreateConst1()) != ^uint32(0) {
		t.Fatal("constant truth")
	}
}

func TestGraphVariableAccessRejectsInternalRoot(t *testing.T) {
	for _, indexOnly := range []bool{false, true} {
		func() {
			g := Dec_GraphCreate(2)
			g.eRoot = Dec_GraphAddNodeAnd(g, Dec_EdgeCreate(0, 0), Dec_EdgeCreate(1, 0))
			defer func() {
				if recover() == nil {
					t.Fatal("internal root accepted as variable", indexOnly)
				}
			}()
			if indexOnly {
				Dec_GraphVarInt(g)
			} else {
				Dec_GraphVar(g)
			}
		}()
	}
}

func TestNodeMetadataBounds(t *testing.T) {
	g := Dec_GraphCreate(1)
	n := Dec_GraphNode(g, 0)
	n.SetLatches([3]uint32{0, 32, 255})
	if got := n.Latches(); got != [3]uint32{0, 0, 31} {
		t.Fatal(got)
	}
	n.SetPrintingFlags([3]uint32{2, 3, 255})
	if got := n.PrintingFlags(); got != [3]uint32{0, 1, 1} {
		t.Fatal(got)
	}
	n.SetLevel(^uint32(0))
	if n.Level != 16383 {
		t.Fatal(n.Level)
	}
	if Dec_GraphNodeInt(g, n) != 0 {
		t.Fatal("metadata changed identity")
	}
}
