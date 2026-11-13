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

// Saig_ManFramesInitialMapped unrolls initialized or arbitrary-state frames,
// applying any previously discovered equivalence map to each internal node.
func Saig_ManFramesInitialMapped(g *aig.Graph, nFrames, nFramesMax, init int32) *aig.Graph {
	if g == nil || nFrames <= 0 || nFrames > nFramesMax || Saig_ManRegNum(g) <= 0 {
		return nil
	}
	Saig_ManStartMap1(g, nFramesMax)
	defer Saig_ManStopMap1(g)
	q := aig.NewGraph()
	q.Name = g.Name
	var pis [][]aig.Edge
	if init == 0 {
		pis = make([][]aig.Edge, nFramesMax)
		for f := range pis {
			pis[f] = make([]aig.Edge, Saig_ManPiNum(g))
			for i := range pis[f] {
				pis[f][i] = q.NewInput()
			}
		}
	}
	state := make([]aig.Edge, Saig_ManRegNum(g))
	for i := range state {
		state[i] = q.Const0()
		if init == 0 {
			state[i] = q.NewInput()
		}
	}
	for f := int32(0); f < nFramesMax; f++ {
		mapping := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
		for i := int32(0); i < Saig_ManPiNum(g); i++ {
			e := aig.Aig_ManCi(g, i)
			if init != 0 {
				mapping[e] = q.NewInput()
			} else {
				mapping[e] = pis[f][i]
			}
		}
		for i := range state {
			mapping[Saig_ManLo(g, int32(i))] = state[i]
		}
		for e, copy := range mapping {
			Saig_ManSetMap1(g, e, f, aig.Aig_Regular(copy))
		}
		active := make(map[aig.Edge]bool)
		var copyEdge func(aig.Edge) aig.Edge
		copyEdge = func(e aig.Edge) aig.Edge {
			regular := aig.Aig_Regular(e)
			if c, ok := mapping[regular]; ok {
				return aig.Aig_NotCond(c, int32(e&1))
			}
			if active[regular] {
				panic("saig: cyclic frame graph")
			}
			active[regular] = true
			n := g.Node(regular)
			var c aig.Edge
			switch n.Kind {
			case aig.And:
				c = q.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.ExclusiveOr:
				c = q.NewXor(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.Buffer, aig.Output:
				c = copyEdge(n.Fanins[0])
			default:
				panic("saig: unmapped frame input")
			}
			Saig_ManSetMap1(g, regular, f, aig.Aig_Regular(c))
			if Saig_ManHasMap2(g) != 0 && (n.Kind == aig.And || n.Kind == aig.ExclusiveOr) {
				look := f
				if look >= nFrames {
					look = nFrames - 1
				}
				var reprFrame int32
				id := Saig_ManGetMap2(g, regular, look, &reprFrame)
				if f >= nFrames {
					reprFrame += f - (nFrames - 1)
				}
				if id < 0 || reprFrame < 0 || reprFrame > f {
					panic("saig: invalid frame representative")
				}
				mapped := Saig_ManGetMap1(g, aig.Edge(id)<<1, reprFrame)
				if mapped < 0 {
					panic("saig: representative not yet mapped")
				}
				repr := aig.Edge(mapped) << 1
				if aig.Aig_ObjPhaseReal(q, c) != q.Node(repr).Phase {
					repr = repr.Not()
				}
				c = repr
			}
			mapping[regular] = c
			delete(active, regular)
			return aig.Aig_NotCond(c, int32(e&1))
		}
		for _, e := range g.LiveEdges() {
			n := g.Node(e)
			if n.Kind == aig.And || n.Kind == aig.ExclusiveOr || n.Kind == aig.Buffer {
				copyEdge(e)
			}
		}
		for i := int32(0); i < Saig_ManPoNum(g); i++ {
			po := aig.Aig_ManCo(g, i)
			out := q.AddOutput(copyEdge(aig.Aig_ObjChild0(g, po)))
			Saig_ManSetMap1(g, po, f, out)
		}
		next := make([]aig.Edge, len(state))
		for i := range state {
			li := Saig_ManLi(g, int32(i))
			next[i] = copyEdge(aig.Aig_ObjChild0(g, li))
			Saig_ManSetMap1(g, li, f, aig.Aig_Regular(next[i]))
		}
		state = next
	}
	if init == 0 {
		for _, e := range state {
			q.AddOutput(e)
		}
		aig.Aig_ManSetRegNum(q, Saig_ManRegNum(g))
	}
	aig.Aig_ManCleanup(q)
	return q
}
