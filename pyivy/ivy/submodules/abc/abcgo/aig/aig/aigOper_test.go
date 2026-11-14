package aig

import "testing"

func TestOriginalAigOperations(t *testing.T) {
	p := NewGraph()
	a, b := Aig_IthVar(p, 0), Aig_IthVar(p, 1)
	if _, ok := Aig_TableLookupInt(p, a, b); ok {
		t.Fatal("unexpected hash entry")
	}
	n := Aig_And(p, a, b)
	if got, ok := Aig_TableLookupInt(p, b, a); !ok || got != n {
		t.Fatal("lookup failed")
	}
	p.AddOutput(Aig_CreateAnd(p, 3))
	p.AddOutput(Aig_CreateOr(p, 3))
	p.AddOutput(Aig_CreateExor(p, 3))
	p.AddOutput(Aig_MiterTwo(p, []Edge{a, b}, []Edge{a, b}))
	p.AddOutput(Aig_Miter(p, []Edge{a, a, b, b.Not()}))
	for mask := 0; mask < 8; mask++ {
		x, y, z := mask&1 != 0, mask&2 != 0, mask&4 != 0
		out, err := p.Evaluate([]bool{x, y, z})
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{x && y && z, x || y || z, (x != y) != z, false, true}
		for i, v := range want {
			if out[i] != v {
				t.Fatalf("assignment %d output %d", mask, i)
			}
		}
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestExorTypeRequiresComplementedOperands(t *testing.T) {
	g := NewGraph()
	x := g.NewInput()
	y := g.NewInput()
	a := g.AND(x, y)
	b := g.AND(x.Not(), y.Not())
	for _, leftComplement := range []bool{false, true} {
		for _, rightComplement := range []bool{false, true} {
			left, right := a, b
			if leftComplement {
				left = left.Not()
			}
			if rightComplement {
				right = right.Not()
			}
			var fan0, fan1 Edge
			got := Aig_ObjIsExorType(g, left, right, &fan0, &fan1)
			want := leftComplement && rightComplement
			if got != want {
				t.Fatalf("left complemented=%v right complemented=%v: got %v", leftComplement, rightComplement, got)
			}
		}
	}
}

func TestMiterCompactsCOperandsInPlace(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	b := g.NewInput()
	c := g.NewInput()
	pairs := []Edge{a, b, b, c}
	result := Aig_Miter(g, pairs)
	want0 := g.XOR(a, b).Not()
	want1 := g.XOR(b, c).Not()
	if pairs[0] != want0 || pairs[1] != want1 {
		t.Fatalf("compacted operands %v, want %v %v", pairs, want0, want1)
	}
	if result != g.AND(want0, want1).Not() {
		t.Fatal("wrong balanced miter")
	}
}
func TestMiterTwoRewritesFirstVector(t *testing.T) {
	g := NewGraph()
	a := g.NewInput()
	b := g.NewInput()
	left := []Edge{a, b}
	right := []Edge{b, a}
	result := Aig_MiterTwo(g, left, right)
	want := g.XOR(a, b).Not()
	if left[0] != want || left[1] != want || result != g.AND(want, want).Not() {
		t.Fatalf("left=%v result=%v", left, result)
	}
}
