package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// createClasses hashes simulation signatures and splits hash collisions, as
// Cec3_ManCreateClasses does. Outputs are excluded from equivalence classes.
func (s *cec3Simulation) createClasses(graph *gia.Gia_Man_t) {
	n := graph.ObjectCount()
	graph.StartRepresentatives()
	for id := int32(0); int(id) < n; id++ {
		gia.Gia_ObjSetNext(graph, id, -1)
	}
	tableSize := util.Abc_PrimeCudd(uint32(n))
	table := make([]int32, tableSize)
	for i := range table {
		table[i] = -1
	}
	for id := int32(0); int(id) < n; id++ {
		_, _, _, kind := graph.Gate(id)
		if kind == 6 {
			continue
		}
		key := Cec3_ManSimHashKey(s.object(id), int32(s.words), tableSize)
		if table[key] == -1 {
			table[key] = id
		} else {
			gia.Gia_ObjSetRepr(graph, id, table[key])
		}
	}
	for id := int32(n - 1); id >= 0; id-- {
		repr := gia.Gia_ObjRepr(graph, id)
		if repr != gia.GIA_VOID {
			gia.Gia_ObjSetNext(graph, id, gia.Gia_ObjNext(graph, repr))
			gia.Gia_ObjSetNext(graph, repr, id)
		}
	}
	for id := int32(0); int(id) < n; id++ {
		if gia.Gia_ObjIsHead(graph, id) != 0 {
			s.refineClass(graph, id)
		}
	}
}

// simulateAnd uses GIA's fanin literals to perform Cec3_ObjSimAnd.
func (s *cec3Simulation) simulateAnd(graph *gia.Gia_Man_t, id int32) {
	a, b, _, kind := graph.Gate(id)
	if kind != 2 {
		panic("cec: Cec3 AND simulation requires an AND node")
	}
	s.andFanins(id, a>>1, b>>1, a&1 != 0, b&1 != 0)
}

// simulateOutput uses the output's complemented driver as in Cec3_ObjSimCo.
func (s *cec3Simulation) simulateOutput(graph *gia.Gia_Man_t, id int32) {
	driver, _, _, kind := graph.Gate(id)
	if kind != 6 {
		panic("cec: Cec3 output simulation requires an output node")
	}
	s.copyDriver(id, driver>>1, driver&1 != 0)
}

// refineClass implements Cec3_ManSimClassRefineOne: the first signature that
// differs from the head becomes the second class head, preserving list order.
func (s *cec3Simulation) refineClass(graph *gia.Gia_Man_t, head int32) {
	if gia.Gia_ObjIsHead(graph, head) == 0 {
		panic("cec: expected simulation class head")
	}
	previous := head
	second := gia.Gia_ObjNext(graph, head)
	for second > 0 && s.equivalent(head, second) {
		previous = second
		second = gia.Gia_ObjNext(graph, second)
	}
	if second <= 0 {
		return
	}
	gia.Gia_ObjSetRepr(graph, second, gia.GIA_VOID)
	previousSecond := second
	for id := gia.Gia_ObjNext(graph, second); id > 0; {
		next := gia.Gia_ObjNext(graph, id)
		if s.equivalent(head, id) {
			gia.Gia_ObjSetNext(graph, previous, id)
			previous = id
		} else {
			gia.Gia_ObjSetRepr(graph, id, second)
			gia.Gia_ObjSetNext(graph, previousSecond, id)
			previousSecond = id
		}
		id = next
	}
	gia.Gia_ObjSetNext(graph, previous, -1)
	gia.Gia_ObjSetNext(graph, previousSecond, -1)
}
