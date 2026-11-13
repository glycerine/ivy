package saig

import "github.com/glycerine/abcgo/aig/aig"

// windowMiter extracts the unmatched cones and compares their boundary
// signals. Corresponding matched signals become shared primary inputs.
func (m *strSimMatcher) windowMiter() *aig.Graph {
	g := [2]*aig.Graph{m.sim[0].graph, m.sim[1].graph}
	if Saig_ManPoNum(g[0]) != Saig_ManPoNum(g[1]) {
		panic("saig: incompatible window outputs")
	}
	// Output and latch-input correspondences are added only for extraction.
	pair := [2]map[aig.Edge]aig.Edge{make(map[aig.Edge]aig.Edge), make(map[aig.Edge]aig.Edge)}
	for side := 0; side < 2; side++ {
		for a, b := range m.pair[side] {
			pair[side][a] = b
		}
	}
	link := func(a, b aig.Edge) { pair[0][a], pair[1][b] = b, a }
	for i := int32(0); i < Saig_ManPoNum(g[0]); i++ {
		link(aig.Aig_ManCo(g[0], i), aig.Aig_ManCo(g[1], i))
	}
	for i := int32(0); i < Saig_ManRegNum(g[0]); i++ {
		lo := Saig_ManLo(g[0], i)
		if other, ok := pair[0][lo]; ok {
			link(Saig_ObjLoToLi(g[0], lo), Saig_ObjLoToLi(g[1], other))
		}
	}
	w := aig.NewGraph()
	w.Name = "wnd"
	copies := [2]map[aig.Edge]aig.Edge{{g[0].Const1(): w.Const1()}, {g[1].Const1(): w.Const1()}}
	// Make shared cut inputs before register outputs.
	for side := 0; side < 2; side++ {
		for _, e := range g[side].LiveEdges() {
			if _, ok := pair[side][e]; ok {
				continue
			}
			n := g[side].Node(e)
			var fanins []aig.Edge
			switch {
			case Saig_ObjIsLo(g[side], e):
				fanins = []aig.Edge{aig.Aig_ObjFanin0(g[side], Saig_ObjLoToLi(g[side], e))}
			case n.Kind == aig.And || n.Kind == aig.ExclusiveOr:
				fanins = []aig.Edge{aig.Aig_Regular(n.Fanins[0]), aig.Aig_Regular(n.Fanins[1])}
			}
			for _, f := range fanins {
				other, ok := pair[side][f]
				if !ok {
					continue
				}
				if _, exists := copies[side][f]; exists {
					continue
				}
				input := w.NewInput()
				copies[side][f], copies[1-side][other] = input, input
			}
		}
	}
	var regs [2][]aig.Edge
	for side := 0; side < 2; side++ {
		for i := int32(0); i < Saig_ManRegNum(g[side]); i++ {
			lo := Saig_ManLo(g[side], i)
			if _, ok := pair[side][lo]; ok {
				continue
			}
			copies[side][lo] = w.NewInput()
			regs[side] = append(regs[side], lo)
		}
	}
	var copyEdge func(int, aig.Edge) aig.Edge
	copyEdge = func(side int, e aig.Edge) aig.Edge {
		reg := aig.Aig_Regular(e)
		if c, ok := copies[side][reg]; ok {
			return aig.Aig_NotCond(c, int32(e&1))
		}
		n := g[side].Node(reg)
		var c aig.Edge
		switch n.Kind {
		case aig.And:
			c = w.AND(copyEdge(side, n.Fanins[0]), copyEdge(side, n.Fanins[1]))
		case aig.ExclusiveOr:
			c = w.NewXor(copyEdge(side, n.Fanins[0]), copyEdge(side, n.Fanins[1]))
		case aig.Buffer, aig.Output:
			c = copyEdge(side, n.Fanins[0])
		default:
			panic("saig: unmapped window input")
		}
		copies[side][reg] = c
		return aig.Aig_NotCond(c, int32(e&1))
	}
	for _, a := range g[0].LiveEdges() {
		b, ok := pair[0][a]
		if !ok {
			continue
		}
		na, nb := g[0].Node(a), g[1].Node(b)
		count := 0
		switch na.Kind {
		case aig.Output:
			count = 1
		case aig.And, aig.ExclusiveOr:
			count = 2
		}
		for i := 0; i < count; i++ {
			fa, fb := aig.Aig_Regular(na.Fanins[i]), aig.Aig_Regular(nb.Fanins[i])
			if _, matched := pair[0][fa]; !matched {
				w.AddOutput(w.XOR(copyEdge(0, fa), copyEdge(1, fb)))
			}
		}
	}
	for side := 0; side < 2; side++ {
		for _, lo := range regs[side] {
			li := Saig_ObjLoToLi(g[side], lo)
			w.AddOutput(copyEdge(side, aig.Aig_ObjChild0(g[side], li)))
		}
	}
	aig.Aig_ManSetRegNum(w, int32(len(regs[0])+len(regs[1])))
	aig.Aig_ManCleanup(w)
	return w
}
