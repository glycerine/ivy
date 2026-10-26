package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestCNFTransformNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_transform.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Seed, IncludeOutputs, NativeVariables int
		Ops                                   [][2]uint32
		Outputs                               []uint32
		ObjectVariables                       []int
		Clauses                               [][]SatLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 48 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		g := darBuildGraph(t, row.Ops, row.Outputs)
		c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
		if e != nil {
			t.Fatal(e)
		}
		transformed, e := c.TransformPolarity(g, row.IncludeOutputs != 0, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		shifted, e := transformed.Shifted(3, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if shifted.variables != row.NativeVariables+3 || !slices.Equal(shifted.objectVariables, row.ObjectVariables) || !equalSatClauses(shifted.clauses, row.Clauses) {
			t.Fatal(row.Seed, row.IncludeOutputs, shifted, row)
		}
		checkDerivedCNFSemantics(t, g, shifted)
		twice, e := transformed.TransformPolarity(g, row.IncludeOutputs != 0, CNFTransformOptions{})
		if e != nil || !equalSatClauses(twice.clauses, c.clauses) || !slices.Equal(twice.outputs, c.outputs) {
			t.Fatal(twice, e)
		}
	}
}
func TestCNFTransformOwnershipLimitsAndLoad(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
	if e != nil {
		t.Fatal(e)
	}
	clone := c.Clone()
	clone.clauses[0][0] ^= 1
	clone.inputs[0] ^= 1
	clone.outputs[0] ^= 1
	clone.objectVariables[0] = 99
	if equalSatClauses(clone.clauses, c.clauses) || clone.inputs[0] == c.inputs[0] || clone.outputs[0] == c.outputs[0] || c.objectVariables[0] == 99 {
		t.Fatal("alias")
	}
	for _, o := range []CNFTransformOptions{{MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {Stop: func() bool { return true }}} {
		r, e := c.Shifted(0, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	if r, e := c.Shifted(SatVar(MaxLiteralID), CNFTransformOptions{}); r != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(r, e)
	}
	s := NewSatSolver()
	var assumptions []SatLit
	for copy := 0; copy < 2; copy++ {
		base := SatVar(copy * c.variables)
		ok, e := c.LoadInto(s, base)
		if e != nil || !ok {
			t.Fatal(ok, e)
		}
		offset := SatLit(base) << 1
		for _, input := range c.inputs {
			l := input + offset
			if copy == 1 {
				l = l.Not()
			}
			assumptions = append(assumptions, l)
		}
		for i, output := range c.outputs {
			l := output + offset
			want := (i == 0) == (copy == 0)
			if !want {
				l = l.Not()
			}
			assumptions = append(assumptions, l)
		}
	}
	answer, e := s.Solve(assumptions, SatLimits{})
	if e != nil || answer.Status != SatSatisfiable {
		t.Fatal(answer, e)
	}
	assumptions[len(c.inputs)] ^= 1
	answer, e = s.Solve(assumptions, SatLimits{})
	if e != nil || answer.Status != SatUnsatisfiable {
		t.Fatal(answer, e)
	}
	constant, e := WriteAigObjectCNF4(NewAigManager(), nil, MappedCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	solver := NewSatSolver()
	ok, e := constant.LoadInto(solver, 0)
	if e != nil || !ok {
		t.Fatal(ok, e)
	}
	answer, e = solver.Solve([]SatLit{1}, SatLimits{})
	if e != nil || answer.Status != SatUnsatisfiable {
		t.Fatal(answer, e)
	}
}
func FuzzCNFTransform(f *testing.F) {
	f.Add([]byte{2, 4, 6, 8, 14, 16}, uint8(3), true)
	f.Fuzz(func(t *testing.T, data []byte, base uint8, include bool) {
		if len(data) > 64 {
			data = data[:64]
		}
		var ops [][2]uint32
		for i := 0; i+1 < len(data); i += 2 {
			ops = append(ops, [2]uint32{uint32(data[i]), uint32(data[i+1])})
		}
		g := darBuildGraph(t, ops, []uint32{uint32(2 * (6 + len(ops))), 3})
		c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
		if e != nil {
			t.Fatal(e)
		}
		phase, e := c.TransformPolarity(g, include, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		shifted, e := phase.Shifted(SatVar(base%32), CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkDerivedCNFSemantics(t, g, shifted)
	})
}
