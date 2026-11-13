package int2

import (
	"reflect"
	"testing"
)

// The checked-in constructor references an undeclared pFrames; no complete
// source contract exists for allocating its structural hash table.
func TestSourceManagerConstructorUnsupported(t *testing.T) {
	if got := Int2_ManCreate(nil, nil); got != nil {
		t.Fatal("incomplete source constructor unexpectedly returned a manager")
	}
	graph := testInt2InputToBadMachine()
	objects := graph.ObjectCount()
	outputs := append([]int32(nil), graph.OutputLiterals()...)
	if got := Int2_ManCreate(graph, &Int2_ManPars_t{}); got != nil {
		t.Fatal("incomplete source constructor reconstructed a manager")
	}
	if graph.ObjectCount() != objects || !reflect.DeepEqual(graph.OutputLiterals(), outputs) {
		t.Fatal("unsupported constructor changed its input graph")
	}
}
