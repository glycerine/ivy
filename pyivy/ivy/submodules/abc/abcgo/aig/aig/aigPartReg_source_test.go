package aig

import "testing"

// The upstream linear entry point is an unfinished traversal diagnostic:
// nPartSize is unused and the function explicitly returns NULL.
func TestRegisterPartitionLinearSourceReturnsNoPartitions(t *testing.T) {
	for _, size := range []int32{0, 1, 2, 100} {
		p := registerPartitionFixture()
		old := p.travID
		if got := Aig_ManRegPartitionLinear(p, size); got != nil {
			t.Errorf("size %d: source returns NULL, got %v", size, got)
		}
		if p.travID == old {
			t.Error("source traversal was skipped")
		}
	}
}

func TestRegisterPartitionCreateMarksFullCone(t *testing.T) {
	p := registerPartitionFixture()
	root := p.outputDriver(p.outputs[1])
	var back []int32
	part := Aig_ManRegCreatePart(p, []int32{0}, nil, nil, &back)
	for _, e := range []Edge{p.Const1(), p.inputs[0], p.inputs[1], p.inputs[2], Aig_Regular(root)} {
		if p.node(e).TravID != p.travID {
			t.Errorf("partition cone object %d not marked", e.NodeID())
		}
		if got := back[p.copies[e.NodeID()].NodeID()]; got != int32(e.NodeID()) {
			t.Errorf("object %d back map = %d", e.NodeID(), got)
		}
	}
	if p.node(p.inputs[3]).TravID == p.travID {
		t.Error("unrelated register marked")
	}
	for _, out := range part.outputs {
		if back[out.NodeID()] != -1 {
			t.Error("fake or real CO should not have a source back map")
		}
	}
}
