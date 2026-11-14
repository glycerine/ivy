package aig

import "testing"

func clauseSatisfied(clause []int32, assignment []bool) bool {
	for _, literal := range clause {
		if assignment[literal>>1] != (literal&1 != 0) {
			return true
		}
	}
	return false
}
func TestInterpolationClausePrimitives(t *testing.T) {
	for phase := int32(0); phase < 2; phase++ {
		buffer := new(Aig_InterpolationClauses)
		if err := Aig_ManInterAddBuffer(buffer, 0, 1, phase, 1); err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 4; mask++ {
			values := []bool{mask&1 != 0, mask&2 != 0}
			satisfied := true
			for _, clause := range buffer.Clauses {
				satisfied = satisfied && clauseSatisfied(clause, values)
			}
			if satisfied != (values[0] == (values[1] != (phase != 0))) {
				t.Fatal(phase, mask, buffer.Clauses)
			}
		}
		xor := new(Aig_InterpolationClauses)
		if err := Aig_ManInterAddXor(xor, 0, 1, 2, phase, 0); err != nil {
			t.Fatal(err)
		}
		for mask := 0; mask < 8; mask++ {
			values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			satisfied := true
			for _, clause := range xor.Clauses {
				satisfied = satisfied && clauseSatisfied(clause, values)
			}
			if satisfied != ((values[0] != values[1]) == (values[2] != (phase != 0))) {
				t.Fatal(phase, mask, xor.Clauses)
			}
		}
		for _, marked := range buffer.PartA {
			if !marked {
				t.Fatal("partition mark lost")
			}
		}
		for _, marked := range xor.PartA {
			if marked {
				t.Fatal("B clause mislabeled")
			}
		}
	}
}
func outputImage(t *testing.T, p *Graph) map[uint64]bool {
	t.Helper()
	image := make(map[uint64]bool)
	for mask := 0; mask < 1<<uint(p.InputCount()); mask++ {
		inputs := make([]bool, p.InputCount())
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		outputs, err := p.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		var code uint64
		for i, value := range outputs {
			if value {
				code |= 1 << uint(i)
			}
		}
		image[code] = true
	}
	return image
}
func TestExactReparameterizationPreservesImage(t *testing.T) {
	for shape := 0; shape < 4; shape++ {
		p := NewGraph()
		a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
		switch shape {
		case 0:
			p.AddOutput(p.AND(a, b))
			p.AddOutput(p.OR(a, b))
		case 1:
			p.AddOutput(a)
			p.AddOutput(a.Not())
			p.AddOutput(p.Const1())
		case 2:
			p.AddOutput(p.NewXor(a, b))
			p.AddOutput(p.AND(b, c))
			p.AddOutput(p.OR(a, c))
		case 3:
			p.AddOutput(p.Const0())
			p.AddOutput(p.Const1())
		}
		q := Aig_ManInterRepar(p, 0)
		if q.InputCount() != 2*p.OutputCount() || q.OutputCount() != p.OutputCount() {
			t.Fatal("incorrect reparameterization interface")
		}
		want, got := outputImage(t, p), outputImage(t, q)
		if len(want) != len(got) {
			t.Fatal(shape, want, got)
		}
		for code := range want {
			if !got[code] {
				t.Fatal(shape, want, got)
			}
		}
		for k, output := range q.outputs {
			for _, input := range Aig_Support(q, output) {
				if Aig_ObjCioId(q, input) > int32(k) {
					t.Fatal("non-triangular output", shape, k, input)
				}
			}
		}
		if err := q.Check(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestInterpolationAppend(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(p.NewXor(a, b))
	base := NewGraph()
	base.NewInput()
	base.NewInput()
	Aig_ManAppend(base, p)
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		want, _ := p.Evaluate(inputs)
		got, err := base.Evaluate(inputs)
		if err != nil || want[0] != got[0] {
			t.Fatal(want, got, err)
		}
	}
}
