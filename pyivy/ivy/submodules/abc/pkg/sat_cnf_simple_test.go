package abc

import (
	"slices"
	"testing"
)

func TestAigCNFExhaustiveProjection(t *testing.T) {
	for mask := 0; mask < 8; mask++ {
		g := NewAigManager()
		a, b, c := aigTestCI(t, g), aigTestCI(t, g), aigTestCI(t, g)
		a = a.NotIf(mask&1 != 0)
		b = b.NotIf(mask&2 != 0)
		c = c.NotIf(mask&4 != 0)
		mux, err := g.Mux(a, b, c)
		if err != nil {
			t.Fatal(err)
		}
		aigTestCO(t, g, mux)
		relation, err := EncodeAigCNF(g)
		if err != nil {
			t.Fatal(err)
		}
		s := NewSatSolver()
		ok, err := relation.LoadInto(s, 0)
		if err != nil || !ok {
			t.Fatal(ok, err)
		}
		clauses := relation.Clauses()
		ins, outs := relation.InputLiterals(), relation.OutputLiterals()
		for assignment := 0; assignment < 8; assignment++ {
			av, bv, cv := (assignment&1 != 0) != (mask&1 != 0), (assignment&2 != 0) != (mask&2 != 0), (assignment&4 != 0) != (mask&4 != 0)
			expected := cv
			if av {
				expected = bv
			}
			for _, outputValue := range []bool{false, true} {
				assumptions := make([]SatLit, 0, 4)
				for i, l := range ins {
					if assignment&(1<<i) == 0 {
						l = l.Not()
					}
					assumptions = append(assumptions, l)
				}
				output := outs[0]
				if !outputValue {
					output = output.Not()
				}
				assumptions = append(assumptions, output)
				projected, _ := satBruteForce(relation.VariableCount(), clauses, assumptions)
				if projected != (expected == outputValue) {
					t.Fatal("CNF existential projection", mask, assignment, outputValue)
				}
				result, err := s.Solve(assumptions, SatLimits{})
				if err != nil {
					t.Fatal(err)
				}
				satCheckAnswer(t, s, clauses, assumptions, result, false)
			}
		}
	}
}
func TestAigCNFConstantsBuffersAndOffsets(t *testing.T) {
	g := NewAigManager()
	a, b := aigTestCI(t, g), aigTestCI(t, g)
	ab := aigTestAnd(t, g, a, b)
	dead := aigTestAnd(t, g, a.Not(), b)
	buffer, err := g.CreateBuffer(ab.Not())
	if err != nil {
		t.Fatal(err)
	}
	aigTestCO(t, g, buffer)
	aigTestCO(t, g, AigConst0)
	aigTestCO(t, g, AigConst1)
	g.Cleanup()
	c, err := EncodeAigCNF(g)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ObjectLiteral(dead.ID()); err == nil {
		t.Fatal("mapped tombstone")
	}
	if _, err := c.ObjectLiteral(InvalidAigID); err == nil {
		t.Fatal("mapped invalid ID")
	}
	source, err := c.ObjectLiteral(ab.ID())
	if err != nil {
		t.Fatal(err)
	}
	buff, err := c.ObjectLiteral(buffer.ID())
	if err != nil || buff != source.Not() {
		t.Fatal("buffer mapping", buff, source, err)
	}
	s := NewSatSolver()
	c.LoadInto(s, 0)
	base := SatVar(c.VariableCount())
	c.LoadInto(s, base)
	inputs := c.InputLiterals()
	outputs := c.OutputLiterals()
	offset := SatLit(base) << 1
	assumptions := []SatLit{inputs[0], inputs[1], outputs[0].Not(), (inputs[0] + offset).Not(), inputs[1] + offset, outputs[0] + offset, outputs[1].Not(), outputs[2]}
	result, err := s.Solve(assumptions, SatLimits{})
	if err != nil || result.Status != SatSatisfiable {
		t.Fatal(result, err)
	}
	returned := c.Clauses()
	returned[0][0] = InvalidSatLit
	if c.Clauses()[0][0] != 0 {
		t.Fatal("aliased CNF clauses")
	}
	returnedInputs := c.InputLiterals()
	returnedInputs[0] = InvalidSatLit
	if slices.Equal(returnedInputs, c.InputLiterals()) {
		t.Fatal("aliased input map")
	}
	if _, err := c.LoadInto(s, SatVar(MaxLiteralID)); err == nil {
		t.Fatal("offset overflow")
	}
	if _, err := EncodeAigCNF(nil); err == nil {
		t.Fatal("nil graph")
	}
}

func FuzzAigCNFProjection(f *testing.F) {
	for _, seed := range [][]byte{{0, 2, 3, 0}, {2, 2, 3, 0, 1, 5, 4, 0}, {25, 0, 4, 0}, {}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 8 {
			return
		}
		g := NewAigManager()
		roots := []AigLit{AigConst0, AigConst1}
		truths := []uint8{0, 255}
		for _, table := range []uint8{0xaa, 0xcc, 0xf0} {
			roots = append(roots, aigTestCI(t, g))
			truths = append(truths, table)
		}
		for i := 0; i+3 < len(data); i += 4 {
			ai, bi := int(data[i+1])%len(roots), int(data[i+2])%len(roots)
			a, b, av, bv := roots[ai], roots[bi], truths[ai], truths[bi]
			if data[i]&8 != 0 {
				a = a.Not()
				av = ^av
			}
			if data[i]&16 != 0 {
				b = b.Not()
				bv = ^bv
			}
			var out AigLit
			var table uint8
			var err error
			switch data[i] % 3 {
			case 0:
				out, err = g.And(a, b)
				table = av & bv
			case 1:
				out, err = g.Or(a, b)
				table = av | bv
			case 2:
				out, err = g.Xor(a, b)
				table = av ^ bv
			}
			if err != nil {
				t.Fatal(err)
			}
			roots = append(roots, out)
			truths = append(truths, table)
		}
		aigTestCO(t, g, roots[len(roots)-1])
		relation, err := EncodeAigCNF(g)
		if err != nil {
			t.Fatal(err)
		}
		clauses := relation.Clauses()
		inputs, outputs := relation.InputLiterals(), relation.OutputLiterals()
		solver := NewSatSolver()
		relation.LoadInto(solver, 0)
		for assignment := 0; assignment < 8; assignment++ {
			for desired := 0; desired < 2; desired++ {
				assumptions := make([]SatLit, 0, 4)
				for i, l := range inputs {
					if assignment&(1<<i) == 0 {
						l = l.Not()
					}
					assumptions = append(assumptions, l)
				}
				output := outputs[0]
				if desired == 0 {
					output = output.Not()
				}
				assumptions = append(assumptions, output)
				expected := int((truths[len(truths)-1]>>assignment)&1) == desired
				projected, _ := satBruteForce(relation.VariableCount(), clauses, assumptions)
				if projected != expected {
					t.Fatal("CNF projection changed Boolean relation")
				}
				result, err := solver.Solve(assumptions, SatLimits{})
				if err != nil {
					t.Fatal(err)
				}
				if (result.Status == SatSatisfiable) != expected {
					t.Fatal("SAT/CNF disagreement")
				}
			}
		}
	})
}
