package aig

import "testing"

// The native build is compiled by GCC, which evaluates function arguments
// right to left: in aigOper.c's Aig_Mux2 (Aig_Or( p, Aig_And(p, pC, p1),
// Aig_And(p, Aig_Not(pC), p0) )) the AND for the 0-branch is created first,
// in Aig_Exor's plain path Aig_And(!p0, p1) before Aig_And(p0, !p1), and in
// Aig_Maj Aig_And(pB, pC), then Aig_And(pA, pC), then Aig_And(pA, pB).
// Node IDs, and through them CNF numbering and written files, follow this.
func TestAigOperBuildOrderLikeGCC(t *testing.T) {
	and := func(g *Graph, a, b Edge) int32 {
		r := g.AND(a, b) // existing node: strashed, no new ID
		return Aig_ObjId(r)
	}
	t.Run("mux", func(t *testing.T) {
		g := NewGraph()
		c, y, n := Aig_ObjCreateCi(g), Aig_ObjCreateCi(g), Aig_ObjCreateCi(g)
		Aig_Mux(g, c, y, n)
		if no, yes := and(g, c.Not(), n), and(g, c, y); no > yes {
			t.Fatalf("Aig_Mux: AND(!c, n) id %d created after AND(c, y) id %d", no, yes)
		}
		g2 := NewGraph()
		c, y, n = Aig_ObjCreateCi(g2), Aig_ObjCreateCi(g2), Aig_ObjCreateCi(g2)
		Aig_Mux2(g2, c, y, n)
		if no, yes := and(g2, c.Not(), n), and(g2, c, y); no > yes {
			t.Fatalf("Aig_Mux2: AND(!c, n) id %d created after AND(c, y) id %d", no, yes)
		}
	})
	t.Run("xor", func(t *testing.T) {
		g := NewGraph()
		a, b := Aig_ObjCreateCi(g), Aig_ObjCreateCi(g)
		Aig_Exor(g, a, b)
		if first, second := and(g, a.Not(), b), and(g, a, b.Not()); first > second {
			t.Fatalf("Aig_Exor: AND(!a, b) id %d created after AND(a, !b) id %d", first, second)
		}
	})
	t.Run("maj", func(t *testing.T) {
		g := NewGraph()
		a, b, c := Aig_ObjCreateCi(g), Aig_ObjCreateCi(g), Aig_ObjCreateCi(g)
		Aig_Maj(g, a, b, c)
		bc, ac, ab := and(g, b, c), and(g, a, c), and(g, a, b)
		if !(bc < ac && ac < ab) {
			t.Fatalf("Aig_Maj: AND ids bc=%d ac=%d ab=%d, want bc < ac < ab", bc, ac, ab)
		}
	})
}
