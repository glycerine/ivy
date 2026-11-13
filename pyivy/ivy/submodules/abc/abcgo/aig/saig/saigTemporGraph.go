package saig

import "github.com/glycerine/abcgo/aig/aig"

// temporCopy duplicates a combinational cone, using the supplied input map.
func temporCopy(dst, src *aig.Graph, mapping map[aig.Edge]aig.Edge, root aig.Edge) aig.Edge {
	reg := aig.Aig_Regular(root)
	if e, ok := mapping[reg]; ok {
		return aig.Aig_NotCond(e, int32(root&1))
	}
	n := src.Node(reg)
	var e aig.Edge
	switch n.Kind {
	case aig.And:
		e = dst.AND(temporCopy(dst, src, mapping, n.Fanins[0]), temporCopy(dst, src, mapping, n.Fanins[1]))
	case aig.ExclusiveOr:
		e = dst.NewXor(temporCopy(dst, src, mapping, n.Fanins[0]), temporCopy(dst, src, mapping, n.Fanins[1]))
	case aig.Buffer, aig.Output:
		e = temporCopy(dst, src, mapping, n.Fanins[0])
	default:
		panic("saig: unmapped temporal input")
	}
	mapping[reg] = e
	return aig.Aig_NotCond(e, int32(root&1))
}

// Saig_ManTemporFrames unrolls the graph from its all-zero initial register
// state and exposes the register values after the requested number of frames.
func Saig_ManTemporFrames(g *aig.Graph, nFrames int32) *aig.Graph {
	if g == nil || nFrames <= 0 {
		return nil
	}
	q := aig.NewGraph()
	q.Name = g.Name
	state := make([]aig.Edge, Saig_ManRegNum(g))
	for i := range state {
		state[i] = q.Const0()
	}
	for frame := int32(0); frame < nFrames; frame++ {
		mapping := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
		for i := int32(0); i < Saig_ManPiNum(g); i++ {
			mapping[aig.Aig_ManCi(g, i)] = q.NewInput()
		}
		for i := int32(0); i < Saig_ManRegNum(g); i++ {
			mapping[Saig_ManLo(g, i)] = state[i]
		}
		next := make([]aig.Edge, len(state))
		for i := range state {
			next[i] = temporCopy(q, g, mapping, aig.Aig_ObjChild0(g, Saig_ManLi(g, int32(i))))
		}
		state = next
	}
	for _, e := range state {
		q.AddOutput(e)
	}
	aig.Aig_ManCleanup(q)
	return q
}

// Saig_ManTemporDecompose inserts a reset register and an initialized prefix
// so the first cycle selects the state reached by the unrolled prefix.
func Saig_ManTemporDecompose(g *aig.Graph, nFrames int32) *aig.Graph {
	if g == nil || nFrames <= 0 || aig.Aig_ManConstrNum(g) != 0 {
		return nil
	}
	frames := Saig_ManTemporFrames(g, nFrames)
	q := aig.NewGraph()
	q.Name = g.Name
	mapping := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
	for i := int32(0); i < Saig_ManPiNum(g); i++ {
		mapping[aig.Aig_ManCi(g, i)] = q.NewInput()
	}
	prefix := map[aig.Edge]aig.Edge{frames.Const1(): q.Const1()}
	for _, e := range frames.Inputs() {
		prefix[e] = q.NewInput()
	}
	initial := make([]aig.Edge, Saig_ManRegNum(g))
	for i := range initial {
		initial[i] = temporCopy(q, frames, prefix, aig.Aig_ObjChild0(frames, aig.Aig_ManCo(frames, int32(i))))
	}
	reset := q.NewInput()
	for i := range initial {
		lo := Saig_ManLo(g, int32(i))
		mapping[lo] = q.Mux(reset, q.NewInput(), initial[i])
	}
	for i := int32(0); i < Saig_ManPoNum(g); i++ {
		q.AddOutput(temporCopy(q, g, mapping, aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, i))))
	}
	q.AddOutput(q.Const1())
	for i := int32(0); i < Saig_ManRegNum(g); i++ {
		q.AddOutput(temporCopy(q, g, mapping, aig.Aig_ObjChild0(g, Saig_ManLi(g, i))))
	}
	aig.Aig_ManSetRegNum(q, Saig_ManRegNum(g)+1)
	aig.Aig_ManCleanup(q)
	return q
}
