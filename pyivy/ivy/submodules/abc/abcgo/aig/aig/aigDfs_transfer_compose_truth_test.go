package aig

import "testing"

func twoInputSop(g *Graph, a, b Edge, table uint8) Edge {
	root := g.Const0()
	for mask := 0; mask < 4; mask++ {
		if table&(1<<mask) == 0 {
			continue
		}
		x, y := a, b
		if mask&1 == 0 {
			x = x.Not()
		}
		if mask&2 == 0 {
			y = y.Not()
		}
		root = g.OR(root, g.AND(x, y))
	}
	return root
}

func TestComposeExhaustiveTwoInputTruthTables(t *testing.T) {
	for original := uint8(0); original < 16; original++ {
		for replacement := uint8(0); replacement < 16; replacement++ {
			for variable := int32(0); variable < 2; variable++ {
				g := NewGraph()
				a, b := g.NewInput(), g.NewInput()
				root := twoInputSop(g, a, b, original)
				function := twoInputSop(g, a, b, replacement)
				g.nodes = g.nodes[:len(g.nodes):len(g.nodes)]
				result := Aig_Compose(g, root, function, variable)
				if result == InvalidEdge {
					t.Fatalf("composition failed: %x %x variable %d", original, replacement, variable)
				}
				g.AddOutput(result)
				for mask := 0; mask < 4; mask++ {
					got, err := g.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
					substituted := mask & ^(1 << variable)
					if replacement&(1<<mask) != 0 {
						substituted |= 1 << variable
					}
					want := original&(1<<substituted) != 0
					if err != nil || got[0] != want {
						t.Fatalf("%x[%d <- %x], mask %d: got %v, error %v, want %v", original, variable, replacement, mask, got, err, want)
					}
				}
				for _, e := range g.LiveEdges() {
					if g.node(e).MarkA {
						t.Fatalf("composition left object %d marked", e)
					}
				}
				if err := g.Check(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestTransferExhaustiveTwoInputTruthTables(t *testing.T) {
	for table := uint8(0); table < 16; table++ {
		src, dst := NewGraph(), NewGraph()
		a, b := src.NewInput(), src.NewInput()
		root := twoInputSop(src, a, b, table)
		for range 3 {
			dst.NewInput()
		}
		result := Aig_Transfer(src, dst, root, 3)
		dst.AddOutput(result)
		for mask := 0; mask < 8; mask++ {
			got, err := dst.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
			want := table&(1<<(mask&3)) != 0
			if err != nil || got[0] != want {
				t.Fatalf("transfer %x, mask %d: got %v, error %v, want %v", table, mask, got, err, want)
			}
		}
		for _, e := range src.LiveEdges() {
			if src.node(e).MarkA {
				t.Fatalf("transfer left object %d marked", e)
			}
		}
		if err := dst.Check(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTransferRetainsExistingExplicitXorExtension(t *testing.T) {
	src, dst := NewGraph(), NewGraph()
	a, b := src.NewInput(), src.NewInput()
	root := src.NewXor(a, b)
	dst.AddOutput(Aig_Transfer(src, dst, root.Not(), 2))
	for mask := 0; mask < 4; mask++ {
		got, err := dst.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || got[0] != (mask == 0 || mask == 3) {
			t.Fatalf("XNOR mask %d: output %v, error %v", mask, got, err)
		}
	}
}
