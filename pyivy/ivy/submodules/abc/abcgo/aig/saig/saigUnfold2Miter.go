package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_ManCreateIndMiter2 constructs the three induction obligations for
// each candidate: p & !Xp, p & Xp, and p & Xp & !XXp.
func Saig_ManCreateIndMiter2(g *aig.Graph, candidates *vec.Vec_Vec_t) *aig.Graph {
	if g == nil || candidates == nil {
		return nil
	}
	q := aig.NewGraph()
	q.Name, q.Spec = g.Name, g.Spec
	var mappings [3]map[aig.Edge]aig.Edge
	for f := range mappings {
		mappings[f] = map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
		for i := int32(0); i < Saig_ManPiNum(g); i++ {
			mappings[f][aig.Aig_ManCi(g, i)] = q.NewInput()
		}
	}
	for i := int32(0); i < Saig_ManRegNum(g); i++ {
		mappings[0][Saig_ManLo(g, i)] = q.NewInput()
	}
	for f := 0; f < 3; f++ {
		for _, e := range g.LiveEdges() {
			n := g.Node(e)
			if n.Kind == aig.And || n.Kind == aig.ExclusiveOr || n.Kind == aig.Buffer {
				temporCopy(q, g, mappings[f], e)
			}
		}
		if f < 2 {
			for i := int32(0); i < Saig_ManRegNum(g); i++ {
				li, lo := Saig_ManLi(g, i), Saig_ManLo(g, i)
				mappings[f+1][lo] = temporCopy(q, g, mappings[f], aig.Aig_ObjChild0(g, li))
			}
		}
	}
	for level := 0; level < vec.Vec_VecSize(candidates); level++ {
		nodes := vec.Vec_VecEntry(candidates, level)
		for _, item := range vec.Vec_PtrArray(nodes) {
			e, ok := item.(aig.Edge)
			if !ok || !g.ValidEdge(e) {
				panic("saig: invalid induction candidate")
			}
			v0, v1, v2 := temporCopy(q, g, mappings[0], e), temporCopy(q, g, mappings[1], e), temporCopy(q, g, mappings[2], e)
			q.AddOutput(q.AND(v0, v1.Not()))
			q.AddOutput(q.AND(v0, v1))
			q.AddOutput(q.AND(q.AND(v0, v1), v2.Not()))
		}
	}
	aig.Aig_ManCleanup(q)
	return q
}
