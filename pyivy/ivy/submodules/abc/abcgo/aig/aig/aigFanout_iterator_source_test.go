package aig

import "testing"

func TestFanoutAccessorsFollowMaintainedCircularList(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	first, second := g.AND(a, b), g.AND(a, c)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	// Reinsertion moves the first-created object behind the other consumer.
	Aig_ObjRemoveFanoutPin(g, a, first, 0)
	Aig_ObjAddFanoutPin(g, a, first, 0)
	head := Aig_FanoutCreate(int32(second.NodeID()), 0)
	tail := Aig_FanoutCreate(int32(first.NodeID()), 0)
	if got := Aig_ObjFanout0Int(g, int32(a.NodeID())); got != head {
		t.Errorf("first handle = %d, want maintained head %d", got, head)
	}
	if got := Aig_ObjFanoutNext(g, head); got != tail {
		t.Errorf("head successor = %d, want %d", got, tail)
	}
	if got := Aig_ObjFanoutNext(g, tail); got != head {
		t.Errorf("tail successor = %d, want circular head %d", got, head)
	}
	if got := Aig_ObjFanout0Int(g, int32(first.NodeID())); got != 0 {
		t.Errorf("unreferenced node head = %d, want 0", got)
	}
}

func TestFanoutAccessorsRequireAllocatedIndex(t *testing.T) {
	for _, next := range []bool{false, true} {
		func() {
			g := NewGraph()
			a, b := g.NewInput(), g.NewInput()
			consumer := g.AND(a, b)
			defer func() {
				if recover() == nil {
					t.Errorf("next=%v: accessor accepted missing fanout allocation", next)
				}
			}()
			if next {
				Aig_ObjFanoutNext(g, Aig_FanoutCreate(int32(consumer.NodeID()), 0))
			} else {
				Aig_ObjFanout0Int(g, int32(a.NodeID()))
			}
		}()
	}
}

func TestFanoutAccessorsIncludeUnaryAndSecondPinUses(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b)
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{b.Not(), InvalidEdge}})
	out := g.AddOutput(b)
	Aig_ManFanoutStart(g)
	defer Aig_ManFanoutStop(g)
	want := []int32{
		Aig_FanoutCreate(int32(gate.NodeID()), 1),
		Aig_FanoutCreate(int32(buffer.NodeID()), 0),
		Aig_FanoutCreate(int32(out.NodeID()), 0),
	}
	if Aig_ObjRefs(g, b) != int32(len(want)) {
		t.Fatal("fixture reference count differs from pin count")
	}
	head := Aig_ObjFanout0Int(g, int32(b.NodeID()))
	fan := head
	for i, expected := range want {
		if fan != expected {
			t.Fatalf("pin %d: handle = %d, want %d", i, fan, expected)
		}
		fan = Aig_ObjFanoutNext(g, fan)
	}
	if fan != head {
		t.Fatal("fanout list did not return to its head")
	}
	single := Aig_ObjFanout0Int(g, int32(a.NodeID()))
	if Aig_ObjFanoutNext(g, single) != single {
		t.Fatal("single-consumer ring must point to itself")
	}
}
