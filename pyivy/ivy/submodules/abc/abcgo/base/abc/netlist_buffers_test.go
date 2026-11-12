package abc

import "testing"

func TestAddPoBuffersPreservesOutputPolarity(t *testing.T) {
	n := NewNetwork("buffers")
	input := n.NewObject(ObjectPI)
	for _, inverted := range []bool{false, true} {
		output := n.NewObject(ObjectPO)
		if err := n.AddFanin(output, Edge{Object: input, Complement: inverted}); err != nil {
			t.Fatal(err)
		}
	}
	Abc_NtkAddPoBuffers(n)
	for i, output := range n.ObjectsOfKind(ObjectPO) {
		edge := output.fanins[0]
		if edge.Complement || edge.Object.Kind != ObjectNode || len(edge.Object.fanins) != 1 {
			t.Fatalf("output %d has invalid buffer: %+v", i, edge)
		}
		if got := edge.Object.fanins[0]; got.Object != input || got.Complement != (i == 1) {
			t.Fatalf("output %d has wrong buffer input: %+v", i, got)
		}
	}
}
