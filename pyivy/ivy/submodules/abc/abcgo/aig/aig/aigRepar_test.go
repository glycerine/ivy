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
