package abc

import (
	"errors"
	"slices"
	"testing"
)

func pdrCheckCubeAig(t testing.TB, c bmcTruthCase, code uint8) {
	t.Helper()
	m := bmcTruthModel(t, c)
	m.graph.Name = "machine"
	m.graph.Spec = "model.aig"
	before := slices.Clone(m.graph.objects)
	cubes := make([]*PdrCube, 0, 3)
	for _, value := range []uint8{code, code ^ 15, 0} {
		state := []PdrLit{}
		for bit := 0; bit < 2; bit++ {
			digit := (value >> (2 * bit)) & 3
			if digit == 1 {
				state = append(state, PdrLit(2*bit))
			}
			if digit == 2 {
				state = append(state, PdrLit(2*bit+1))
			}
		}
		cube, err := NewPdrCube(state, nil)
		if err != nil {
			t.Fatal(err)
		}
		cubes = append(cubes, cube)
	}
	inv, err := NewPdrInvariant(cubes)
	if err != nil {
		t.Fatal(err)
	}
	g, err := inv.ToAig(m.graph, PdrAigOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != m.graph.Name || g.Spec != m.graph.Spec || g.RegisterCount() != 2 || g.PrimaryInputCount() != 1 || g.PrimaryOutputCount() != 3 || g.ConstraintCount() != 0 {
		t.Fatal("graph metadata")
	}
	for assignment := 0; assignment < 8; assignment++ {
		bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		values, err := g.Evaluate(bits)
		if err != nil {
			t.Fatal(err)
		}
		for i, cube := range cubes {
			expected, err := cube.Matches(bits[1:])
			if err != nil || values[i] != expected {
				t.Fatal("cube polarity", assignment, i, err)
			}
		}
		if values[3] != (c.next0&(1<<assignment) != 0) || values[4] != (c.next1&(1<<assignment) != 0) {
			t.Fatal("changed transition", assignment)
		}
	}
	if !slices.Equal(before, m.graph.objects) {
		t.Fatal("source mutated")
	}
}
func TestPdrCubeAigTruthTables(t *testing.T) {
	for i := 0; i < 256; i++ {
		pdrCheckCubeAig(t, bmcTruthCase{bad: uint8(i), constraint: uint8(i * 13), next0: uint8(i * 37), next1: uint8(i * 73)}, uint8(i))
	}
}
func TestPdrCubeAigLimitsAndEmpty(t *testing.T) {
	m := bmcTruthModel(t, bmcTruthCase{constraint: 255, next0: 170, next1: 204})
	inv, _ := NewPdrInvariant(nil)
	g, err := inv.ToAig(m.graph, PdrAigOptions{})
	if err != nil || g.PrimaryOutputCount() != 0 || g.RegisterCount() != 2 {
		t.Fatal(g, err)
	}
	for _, options := range []PdrAigOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		if _, err := inv.ToAig(m.graph, options); !errors.Is(err, ErrLimit) {
			t.Fatal(err)
		}
	}
	for _, options := range []PdrAigOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		if _, err := inv.ToAig(m.graph, options); err == nil {
			t.Fatal("invalid limits")
		}
	}
	for _, threshold := range []int{1, 4, 10, 30} {
		calls := 0
		_, err := inv.ToAig(m.graph, PdrAigOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) {
			t.Fatal("cancellation", threshold, err)
		}
	}
	cube, _ := NewPdrCube([]PdrLit{4}, nil)
	invalid, _ := NewPdrInvariant([]*PdrCube{cube})
	if _, err := invalid.ToAig(m.graph, PdrAigOptions{}); err == nil {
		t.Fatal("invalid register")
	}
}
func FuzzPdrCubeAig(f *testing.F) {
	f.Add(uint8(192), uint8(170), uint8(204), uint8(255), uint8(6))
	f.Fuzz(func(t *testing.T, bad, n0, n1, env, code uint8) {
		pdrCheckCubeAig(t, bmcTruthCase{bad: bad, next0: n0, next1: n1, constraint: env}, code)
	})
}

func TestPdrCubeAigBuffersAndTombstones(t *testing.T) {
	source := NewAigManager()
	input := aigTestCI(t, source)
	reg := aigTestCI(t, source)
	buffer, err := source.CreateBuffer(input.Not())
	if err != nil {
		t.Fatal(err)
	}
	// Remove a dangling node, leaving a permanent hole in source IDs.
	aigTestAnd(t, source, input, reg)
	source.Cleanup()
	aigTestCO(t, source, reg)
	// Cleanup may remove the unused buffer; append a live one afterward.
	buffer, err = source.CreateBuffer(input.Not())
	if err != nil {
		t.Fatal(err)
	}
	aigTestCO(t, source, buffer)
	if err := source.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	cube, _ := NewPdrCube([]PdrLit{1}, nil)
	inv, _ := NewPdrInvariant([]*PdrCube{cube})
	graph, err := inv.ToAig(source, PdrAigOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		values, err := graph.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
		if err != nil || values[0] != (assignment&2 == 0) || values[1] != (assignment&1 == 0) {
			t.Fatal(assignment, values, err)
		}
	}
}
