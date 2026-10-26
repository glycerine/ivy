package abc

import (
	"errors"
	"slices"
	"testing"
)

func simpleNativeFixture(t testing.TB) *AigManager {
	t.Helper()
	g := NewAigManager()
	a, b := aigTestCI(t, g), aigTestCI(t, g)
	ab := aigTestAnd(t, g, a, b)
	aigTestCO(t, g, ab)
	aigTestCO(t, g, a.Not())
	return g
}

func TestAigSimpleCNFNativeOrder(t *testing.T) {
	g := simpleNativeFixture(t)
	c, err := DeriveAigSimpleCNF(g, 1, AigSimpleCNFOptions{})
	if err != nil || c.VariableCount() != 7 || c.ExposedOutputCount() != 1 || c.RetimingInitialization() {
		t.Fatal(c, err)
	}
	if want := []int{6, 4, 5, 3, 1, 2}; !slices.Equal(c.ObjectVariables(), want) {
		t.Fatal(c.ObjectVariables(), want)
	}
	wantClauses := [][]SatLit{{6, 9, 11}, {7, 8}, {7, 10}, {12}, {6}, {4, 8}, {5, 9}}
	if !equalSatClauses(c.Clauses(), wantClauses) || !slices.Equal(c.InputLiterals(), []SatLit{8, 10}) || !slices.Equal(c.OutputLiterals(), []SatLit{6, 4}) {
		t.Fatal(c.Clauses(), c.InputLiterals(), c.OutputLiterals())
	}
	asserted, err := DeriveAigSimpleCNF(g, 0, AigSimpleCNFOptions{})
	if err != nil || asserted.VariableCount() != 5 || !equalSatClauses(asserted.Clauses(), [][]SatLit{{2, 5, 7}, {3, 4}, {3, 6}, {8}, {2}, {5}}) {
		t.Fatal(asserted, err)
	}
	retiming, err := DeriveAigSimpleRetimingCNF(g, AigSimpleCNFOptions{})
	if err != nil || !retiming.RetimingInitialization() || !equalSatClauses(retiming.Clauses(), [][]SatLit{{6, 9, 11}, {7, 8}, {7, 10}, {12}, {2, 7}, {3, 6}, {3}, {4, 8}, {5, 9}, {5}}) {
		t.Fatal(retiming, err)
	}
}

func checkAigSimpleCNFSemantics(t testing.TB, g *AigManager, c *AigSimpleCNF) {
	t.Helper()
	solver := NewSatSolver()
	_, err := c.LoadInto(solver, 0)
	if err != nil {
		t.Fatal(err)
	}
	inputs := c.InputLiterals()
	for assignment := 0; assignment < 1<<uint(len(inputs)); assignment++ {
		values := make([]bool, len(inputs))
		assumptions := make([]SatLit, len(inputs))
		for i, literal := range inputs {
			values[i] = assignment>>uint(i)&1 != 0
			assumptions[i] = literal
			if !values[i] {
				assumptions[i] = literal.Not()
			}
		}
		outputs, err := g.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		want := true
		if c.retiming {
			for _, value := range outputs {
				want = want && !value
			}
		} else {
			for _, value := range outputs[:len(outputs)-c.exposedOutputs] {
				want = want && value
			}
		}
		result, err := solver.Solve(assumptions, SatLimits{})
		if err != nil || (result.Status == SatSatisfiable) != want {
			t.Fatal(assignment, outputs, want, result, err)
		}
		if want && !c.retiming {
			for i := len(outputs) - c.exposedOutputs; i < len(outputs); i++ {
				literal := c.outputs[i]
				if !outputs[i] {
					literal = literal.Not()
				}
				answer, err := solver.Solve(append(slices.Clone(assumptions), literal), SatLimits{})
				if err != nil || answer.Status != SatSatisfiable {
					t.Fatal("exposed", assignment, i, answer, err)
				}
				answer, err = solver.Solve(append(slices.Clone(assumptions), literal.Not()), SatLimits{})
				if err != nil || answer.Status != SatUnsatisfiable {
					t.Fatal("opposite", assignment, i, answer, err)
				}
			}
		}
	}
}

func TestAigSimpleCNFLimitsOwnershipAndSemantics(t *testing.T) {
	g := simpleNativeFixture(t)
	for exposed := 0; exposed <= g.OutputCount(); exposed++ {
		c, err := DeriveAigSimpleCNF(g, exposed, AigSimpleCNFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkAigSimpleCNFSemantics(t, g, c)
	}
	retiming, err := DeriveAigSimpleRetimingCNF(g, AigSimpleCNFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	checkAigSimpleCNFSemantics(t, g, retiming)
	for _, options := range []AigSimpleCNFOptions{{MaxObjects: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}} {
		if c, err := DeriveAigSimpleCNF(g, 1, options); c != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(c, err)
		}
	}
	if c, err := DeriveAigSimpleCNF(g, -1, AigSimpleCNFOptions{}); c != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(c, err)
	}
	buffered := NewAigManager()
	a := aigTestCI(t, buffered)
	b, _ := buffered.CreateBuffer(a)
	aigTestCO(t, buffered, b)
	if c, err := DeriveAigSimpleCNF(buffered, 0, AigSimpleCNFOptions{}); c != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(c, err)
	}
	clauses := retiming.Clauses()
	clauses[0][0] ^= 1
	if equalSatClauses(clauses, retiming.Clauses()) {
		t.Fatal("clause alias")
	}
}

func FuzzAigSimpleCNF(f *testing.F) {
	f.Add(uint16(0x1234), uint8(1))
	f.Add(uint16(0xabcd), uint8(2))
	f.Fuzz(func(t *testing.T, seed uint16, rawExposed uint8) {
		g := NewAigManager()
		nodes := []AigLit{aigTestCI(t, g), aigTestCI(t, g), aigTestCI(t, g)}
		for i := 0; i < 4; i++ {
			first := nodes[int(seed>>uint(2*i))%len(nodes)]
			second := nodes[int(seed>>uint(2*i+5))%len(nodes)]
			if seed>>uint(i+9)&1 != 0 {
				first = first.Not()
			}
			if seed>>uint(i+12)&1 != 0 {
				second = second.Not()
			}
			node, err := g.And(first, second)
			if err != nil {
				t.Fatal(err)
			}
			nodes = append(nodes, node)
		}
		for i := 0; i < 3; i++ {
			literal := nodes[(int(seed)+3*i)%len(nodes)]
			if seed>>uint(i)&1 != 0 {
				literal = literal.Not()
			}
			aigTestCO(t, g, literal)
		}
		exposed := int(rawExposed) % (g.OutputCount() + 1)
		c, err := DeriveAigSimpleCNF(g, exposed, AigSimpleCNFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkAigSimpleCNFSemantics(t, g, c)
		retiming, err := DeriveAigSimpleRetimingCNF(g, AigSimpleCNFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkAigSimpleCNFSemantics(t, g, retiming)
	})
}
