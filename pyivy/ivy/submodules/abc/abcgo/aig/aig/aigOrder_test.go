package aig

import (
	"math/rand/v2"
	"slices"
	"testing"
)

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

func TestOrderEditsAgainstSliceModel(t *testing.T) {
	g := NewGraph()
	inputs := make([]Edge, 41)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	var pool []Edge
	for i := 0; i < 40; i++ {
		pool = append(pool, g.AND(inputs[i], inputs[i+1]))
	}
	Aig_ManOrderStart(g)
	model := slices.Clone(pool)
	rng := rand.New(rand.NewPCG(19, 31))
	for step := 0; step < 1000; step++ {
		cut := rng.IntN(len(model) + 1)
		prev, next := int32(0), int32(0)
		if cut > 0 {
			prev = int32(model[cut-1].NodeID())
		}
		if cut < len(model) {
			next = int32(model[cut].NodeID())
		}
		g.iPrev, g.iNext, g.nAndPrev = prev, next, int32(cut)
		processed := int32(cut)
		switch rng.IntN(3) {
		case 0:
			var inactive []Edge
			for _, edge := range pool {
				if !slices.Contains(model, edge) {
					inactive = append(inactive, edge)
				}
			}
			if len(inactive) > 0 {
				edge := inactive[rng.IntN(len(inactive))]
				Aig_ObjOrderInsert(g, int32(edge.NodeID()))
				model = slices.Insert(model, cut, edge)
			}
		case 1:
			if len(model) > 0 {
				at := rng.IntN(len(model))
				id := int32(model[at].NodeID())
				Aig_ObjOrderRemove(g, id)
				if id == prev {
					prev = 0
					if at > 0 {
						prev = int32(model[at-1].NodeID())
					}
					processed--
				}
				if id == next {
					next = 0
					if at+1 < len(model) {
						next = int32(model[at+1].NodeID())
					}
				}
				model = slices.Delete(model, at, at+1)
			}
		case 2:
			if next != 0 {
				Aig_ObjOrderAdvance(g)
				prev = next
				processed++
			}
		}
		if got := g.OrderedEdges(); !slices.Equal(got, model) || g.nAndTotal != int32(len(model)) {
			t.Fatalf("step %d: order %v, want %v", step, got, model)
		}
		if g.iPrev != prev || g.iNext != next || g.nAndPrev != processed {
			t.Fatalf("step %d: cursor %d/%d/%d, want %d/%d/%d", step, g.iPrev, g.iNext, g.nAndPrev, prev, next, processed)
		}
		previous := int32(0)
		for _, edge := range model {
			id := int32(edge.NodeID())
			if g.order[id].prev != previous || g.order[previous].next != id {
				t.Fatalf("step %d: broken reciprocal link at %d", step, id)
			}
			previous = id
		}
		if g.order[0].prev != previous || g.order[previous].next != 0 {
			t.Fatalf("step %d: broken sentinel links", step)
		}
	}
}

func TestOrderGrowthForNewHighObjectIds(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	first := g.AND(a, b)
	Aig_ManOrderStart(g)
	for len(g.nodes) <= 4096 {
		g.NewInput()
	}
	last := g.AND(a.Not(), b)
	Aig_ObjOrderInsert(g, int32(last.NodeID()))
	if got := g.OrderedEdges(); !slices.Equal(got, []Edge{first, last}) {
		t.Fatal("growth lost an order entry", got)
	}
	Aig_ObjOrderRemove(g, int32(last.NodeID()))
	Aig_ObjOrderRemove(g, int32(first.NodeID()))
	if len(g.OrderedEdges()) != 0 || g.order[0] != (orderLinks{0, 0}) {
		t.Fatal("removal did not restore the empty sentinel")
	}
}
