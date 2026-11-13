package saig

import "github.com/glycerine/abcgo/aig/aig"

// Saig_ManFramesNonInitial unrolls the graph with unconstrained initial
// registers. Each original object/frame maps to an object in the returned graph.
func Saig_ManFramesNonInitial(g *aig.Graph, nFrames int32) *aig.Graph {
	if g == nil || nFrames <= 0 || Saig_ManRegNum(g) <= 0 {
		return nil
	}
	Saig_ManStartMap1(g, nFrames)
	q := aig.NewGraph()
	q.Name = g.Name
	state := make([]aig.Edge, Saig_ManRegNum(g))
	for i := range state {
		state[i] = q.NewInput()
	}
	for frame := int32(0); frame < nFrames; frame++ {
		mapping := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
		for i := int32(0); i < Saig_ManPiNum(g); i++ {
			mapping[aig.Aig_ManCi(g, i)] = q.NewInput()
		}
		for i := range state {
			mapping[Saig_ManLo(g, int32(i))] = state[i]
		}
		for _, e := range g.LiveEdges() {
			n := g.Node(e)
			if n.Kind == aig.And || n.Kind == aig.ExclusiveOr || n.Kind == aig.Buffer {
				temporCopy(q, g, mapping, e)
			}
		}
		for i := int32(0); i < Saig_ManPoNum(g); i++ {
			po := aig.Aig_ManCo(g, i)
			mapping[po] = q.AddOutput(temporCopy(q, g, mapping, aig.Aig_ObjChild0(g, po)))
		}
		next := make([]aig.Edge, len(state))
		for i := range state {
			li := Saig_ManLi(g, int32(i))
			next[i] = temporCopy(q, g, mapping, aig.Aig_ObjChild0(g, li))
			mapping[li] = next[i]
		}
		for _, e := range g.LiveEdges() {
			if copy, ok := mapping[e]; ok {
				Saig_ManSetMap1(g, e, frame, aig.Aig_Regular(copy))
			}
		}
		state = next
	}
	for _, e := range state {
		q.AddOutput(e)
	}
	aig.Aig_ManCleanup(q)
	return q
}
