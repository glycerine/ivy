package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManCollectedDiffNodes marks and returns objects without a paired object
// in the other graph. The representative links must be reciprocal.
func Saig_ManCollectedDiffNodes(a, b *aig.Graph) *vec.Vec_Ptr_t {
	if a == nil || b == nil {
		return nil
	}
	aig.Aig_ManIncrementTravId(a)
	nodes := vec.Vec_PtrAlloc(0)
	for _, e := range a.LiveEdges() {
		other := aig.Aig_ObjRepr(a, e)
		if other != aig.InvalidEdge {
			if !b.ValidEdge(other) || aig.Aig_ObjRepr(b, other) != e {
				panic("saig: nonreciprocal window correspondence")
			}
			continue
		}
		aig.Aig_ObjSetTravIdCurrent(a, e)
		n := a.Node(e)
		if n.Kind == aig.Input || n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
			vec.Vec_PtrPush(nodes, e)
		}
	}
	for i := int32(0); i < Saig_ManRegNum(a); i++ {
		if aig.Aig_ObjIsTravIdCurrent(a, Saig_ManLi(a, i)) != aig.Aig_ObjIsTravIdCurrent(a, Saig_ManLo(a, i)) {
			panic("saig: inconsistent latch correspondence")
		}
	}
	return nodes
}

// Saig_ManWindowExtractMiter compares unmatched portions of two graphs using
// reciprocal cross-graph representative links established by the caller.
func Saig_ManWindowExtractMiter(a, b *aig.Graph) *aig.Graph {
	if a == nil || b == nil {
		return nil
	}
	m := &strSimMatcher{sim: [2]*strSimWords{newStrSimWords(a), newStrSimWords(b)}, pair: [2]map[aig.Edge]aig.Edge{make(map[aig.Edge]aig.Edge), make(map[aig.Edge]aig.Edge)}}
	for _, e := range a.LiveEdges() {
		other := aig.Aig_ObjRepr(a, e)
		if other == aig.InvalidEdge {
			continue
		}
		if !b.ValidEdge(other) || aig.Aig_ObjRepr(b, other) != e {
			panic("saig: nonreciprocal window correspondence")
		}
		m.link(e, other)
	}
	return m.windowMiter()
}
