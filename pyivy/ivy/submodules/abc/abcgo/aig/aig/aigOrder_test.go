package aig

import "testing"

func TestOrderInsertionRemovalAndCursor(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	first := p.AND(a, b)
	last := p.AND(first, c)
	Aig_ManOrderStart(p)
	order := p.OrderedEdges()
	if len(order) != 2 || order[0] != first || order[1] != last {
		t.Fatal(order)
	}
	p.iNext = int32(first.NodeID())
	Aig_ObjOrderAdvance(p)
	p.iNext = p.order[p.iPrev].next
	inserted := p.AND(a, c)
	Aig_ObjOrderInsert(p, int32(inserted.NodeID()))
	order = p.OrderedEdges()
	if len(order) != 3 || order[1] != inserted {
		t.Fatal(order)
	}
	Aig_ObjOrderRemove(p, int32(first.NodeID()))
	if p.iPrev != 0 {
		t.Fatal("previous cursor not repaired")
	}
	Aig_ObjOrderRemove(p, int32(last.NodeID()))
	if p.iNext != 0 {
		t.Fatal("next cursor not repaired")
	}
	order = p.OrderedEdges()
	if len(order) != 1 || order[0] != inserted {
		t.Fatal(order)
	}
	Aig_ManOrderStop(p)
	if p.OrderedEdges() != nil {
		t.Fatal("order storage retained")
	}
}
