package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

// The permanent rows are immutable test inputs. Independent truth-table and
// counterexample checks below make their correctness stand alone.
func TestFoundationDatasets(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/foundations.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Counterexamples []struct {
			Name                                string
			Registers, Inputs, Frames, Property int
			Bits                                []int
		}
		AigOperations []struct{ Op, A, B, C, AC, BC, CC, Truth int } `json:"aig_operations"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Counterexamples) != 6 || len(fixture.AigOperations) != 128 {
		t.Fatal("incomplete foundation dataset")
	}
	base, err := CounterexampleFromValues(1, 4, 2, 3, []int{1, 1, 0, 0, 1, 0, 1, 1, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range fixture.Counterexamples {
		var c *Counterexample
		switch record.Name {
		case "base":
			c = base
		case "dup":
			c, err = base.DupInputs(3)
		case "permute":
			c, err = base.PermuteInputs([]int{2, 0, 3, 1})
		case "phase":
			c, err = base.TransformPhase(2, 2, 3)
		case "temporal":
			c, err = base.TransformTemporal(2, 2)
		case "unknown-init":
			c, err = base.TransformUnknownInitial("1xX0")
		default:
			t.Fatalf("unknown fixture %q", record.Name)
		}
		if err != nil {
			t.Fatal(record.Name, err)
		}
		if c.registers != record.Registers || c.inputs != record.Inputs || c.frames != record.Frames || c.Property != record.Property {
			t.Fatal("counterexample dimension mismatch", record.Name)
		}
		bits := cexBits(t, c)
		if len(bits) != len(record.Bits) {
			t.Fatal(record.Name, "bit count")
		}
		for i, bit := range bits {
			if bit != (record.Bits[i] != 0) {
				t.Fatal(record.Name, "bit mismatch", i)
			}
		}
	}
	m := NewAigManager()
	roots := []AigLit{AigConst0, AigConst1}
	for i := 0; i < 4; i++ {
		roots = append(roots, aigTestCI(t, m))
	}
	for index, op := range fixture.AigOperations {
		a, b, c := roots[op.A].NotIf(op.AC != 0), roots[op.B].NotIf(op.BC != 0), roots[op.C].NotIf(op.CC != 0)
		var result AigLit
		switch op.Op {
		case 0:
			result, err = m.And(a, b)
		case 1:
			result, err = m.Or(a, b)
		case 2:
			result, err = m.Xor(a, b)
		case 3:
			result, err = m.Mux(a, b, c)
		case 4:
			result, err = m.Majority(a, b, c)
		default:
			t.Fatal("bad fixture operation")
		}
		if err != nil {
			t.Fatal(err)
		}
		roots = append(roots, result)
		aigTestCO(t, m, result)
		for assignment := 0; assignment < 16; assignment++ {
			input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
			got, err := m.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			if got[index] != (op.Truth&(1<<assignment) != 0) {
				t.Fatalf("dataset truth mismatch operation %d assignment %d", index, assignment)
			}
		}
	}
	aigTestCheck(t, m)
	clone := m.Clone()
	m.Cleanup()
	for assignment := 0; assignment < 16; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		a, _ := m.Evaluate(inputs)
		b, _ := clone.Evaluate(inputs)
		if !slices.Equal(a, b) {
			t.Fatal("cleanup changed foundation operation")
		}
	}
}
