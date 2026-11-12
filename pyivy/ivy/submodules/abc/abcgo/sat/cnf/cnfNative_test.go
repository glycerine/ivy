package cnf

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func cnfSatisfiableWithInputs(cnf *CNF, inputs map[int]bool) bool {
	for assignment := 0; assignment < 1<<cnf.Variables; assignment++ {
		matches := true
		for variable, value := range inputs {
			if ((assignment>>variable)&1 != 0) != value {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		valid := true
		for _, clause := range cnf.Clauses {
			satisfied := false
			for _, lit := range clause {
				value := (assignment>>(lit>>1))&1 != 0
				if lit&1 != 0 {
					value = !value
				}
				if value {
					satisfied = true
					break
				}
			}
			if !satisfied {
				valid = false
				break
			}
		}
		if valid {
			return true
		}
	}
	return false
}

func TestEncodeAIGAndAssertsOutput(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	and := g.AND(a, b)
	g.AddOutput(and)
	cnf, err := EncodeAIG(g, -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := cnf.Validate(); err != nil {
		t.Fatal(err)
	}
	if cnf.Variables != 4 {
		t.Fatalf("variables=%d, want 4", cnf.Variables)
	}
	if cnf.VarNums[0] != -1 || cnf.VarNums[a.NodeID()] != 0 || cnf.VarNums[b.NodeID()] != 1 {
		t.Fatalf("unexpected object map: %v", cnf.VarNums)
	}
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			want := x == 1 && y == 1
			if got := cnfSatisfiableWithInputs(cnf, map[int]bool{0: x == 1, 1: y == 1}); got != want {
				t.Errorf("inputs %d,%d satisfiable=%t, want %t", x, y, got, want)
			}
		}
	}
}

func TestEncodeAIGXORAndComplement(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	xor := g.XOR(a.Not(), b)
	g.AddOutput(xor)
	cnf, err := EncodeAIG(g, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := cnf.Validate(); err != nil {
		t.Fatal(err)
	}
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			want := (x == 0) != (y == 1)
			if got := cnfSatisfiableWithInputs(cnf, map[int]bool{0: x == 1, 1: y == 1}); got != want {
				t.Errorf("inputs %d,%d satisfiable=%t, want %t", x, y, got, want)
			}
		}
	}
}

func TestEncodeAIGConstantFalseOutput(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const0())
	cnf, err := EncodeAIG(g, 1)
	if err != nil {
		t.Fatal(err)
	}
	if cnfSatisfiableWithInputs(cnf, nil) {
		t.Fatal("constant-false output should make the formula unsatisfiable")
	}
}
