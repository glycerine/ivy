package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestProjectAIGPropertiesPreservesRegisterInterface(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state.Not())
	graph.AddOutput(state)
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)

	projected, err := projectAIGProperties(graph, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	if projected.InputCount() != 2 || projected.OutputCount() != 2 || projected.RegisterCount() != 1 {
		t.Fatalf("terminals=(inputs %d, outputs %d, registers %d), want (2,2,1)", projected.InputCount(), projected.OutputCount(), projected.RegisterCount())
	}
	property := projected.Node(projected.Outputs()[0])
	if !property.Fanins[0].Complemented() {
		t.Fatalf("projected property driver=%v, want complemented state", property.Fanins[0])
	}
	registerInput := projected.Node(projected.Outputs()[1])
	if registerInput.Fanins[0] != projected.Inputs()[0] {
		t.Fatalf("projected register driver=%v, want primary input %v", registerInput.Fanins[0], projected.Inputs()[0])
	}
}

func TestProjectAIGPropertiesRejectsInvalidIndex(t *testing.T) {
	graph := aig.NewGraph()
	graph.NewInput()
	graph.AddOutput(graph.Inputs()[0])
	if _, err := projectAIGProperties(graph, []int{1}); err == nil {
		t.Fatal("expected invalid output index to fail")
	}
}
