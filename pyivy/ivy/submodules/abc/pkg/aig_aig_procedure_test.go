package abc

// Tests for the owned aigTest.c diagnostic constructor.

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestAigProcedureNativeAndSemantics(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_procedure.json")
	if err != nil {
		t.Fatal(err)
	}
	var row struct {
		Check, Registers int
		Cis, Cos         []AigID
		Objects          []aigOperationObject
	}
	if err = json.Unmarshal(data, &row); err != nil {
		t.Fatal(err)
	}
	graph, err := BuildAigProcedureTest()
	if err != nil || (graph.Check() == nil) != (row.Check != 0) || graph.RegisterCount() != row.Registers || !slices.Equal(graph.Inputs(), row.Cis) || !slices.Equal(graph.Outputs(), row.Cos) || !slices.Equal(aigOperationObjects(graph), row.Objects) {
		t.Fatal(err, graph.Check(), graph.RegisterCount(), row.Registers, graph.Inputs(), row.Cis, graph.Outputs(), row.Cos, aigOperationObjects(graph), row.Objects)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, err := graph.Evaluate(inputs)
		mux := inputs[2]
		if inputs[0] {
			mux = inputs[1]
		}
		want := []bool{mux, inputs[0] && inputs[1]}
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, err)
		}
	}
	again, err := BuildAigProcedureTest()
	if err != nil {
		t.Fatal(err)
	}
	graph.Name = "mutated"
	if again.Name != "" || graph == again || &graph.objects[0] == &again.objects[0] {
		t.Fatal("returned managers are not independently owned")
	}
}
