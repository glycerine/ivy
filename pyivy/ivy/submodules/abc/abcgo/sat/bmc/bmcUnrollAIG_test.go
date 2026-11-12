package bmc

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestUnrollAIGFramesPreservesTransitionOrder(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)

	unrolled := UnrollAIGFrames(graph, 2)
	if unrolled == nil {
		t.Fatal("expected unrolled graph")
	}
	if unrolled.InputCount() != 2 || unrolled.OutputCount() != 3 {
		t.Fatalf("terminals=(%d inputs,%d outputs), want (2,3)", unrolled.InputCount(), unrolled.OutputCount())
	}
	got, err := unrolled.Evaluate([]bool{true, false})
	if err != nil {
		t.Fatal(err)
	}
	if want := []bool{false, true, false}; !reflect.DeepEqual(got, want) {
		t.Fatalf("frame outputs=%v, want %v", got, want)
	}
}

func TestUnrollAIGFramesRejectsInvalidDepth(t *testing.T) {
	graph := aig.NewGraph()
	if UnrollAIGFrames(graph, 0) != nil {
		t.Fatal("expected zero frame request to fail")
	}
}
