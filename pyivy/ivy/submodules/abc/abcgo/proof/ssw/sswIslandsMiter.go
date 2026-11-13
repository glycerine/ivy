package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// sswIslandMiter duplicates the two circuits while retaining each node's
// miter copy, which the original Saig_ManCreateMiter stores in pData.
func sswIslandMiter(p0, p1 *aig.Graph) *aig.Graph {
	if saig.Saig_ManPiNum(p0) != saig.Saig_ManPiNum(p1) || saig.Saig_ManPoNum(p0) != saig.Saig_ManPoNum(p1) {
		panic("ssw: incompatible miter interfaces")
	}
	dst := aig.NewGraph()
	dst.Name = "miter"
	pi := int(saig.Saig_ManPiNum(p0))
	in0, in1 := p0.Inputs(), p1.Inputs()
	for _, p := range []*aig.Graph{p0, p1} {
		p.SetObjectData(p.Const1(), dst.Const1())
	}
	for i := 0; i < pi; i++ {
		e := dst.NewInput()
		p0.SetObjectData(in0[i], e)
		p1.SetObjectData(in1[i], e)
	}
	for _, e := range in0[pi:] {
		p0.SetObjectData(e, dst.NewInput())
	}
	for _, e := range in1[pi:] {
		p1.SetObjectData(e, dst.NewInput())
	}
	copyGraph := func(src *aig.Graph) {
		copyLit := func(e aig.Edge) aig.Edge {
			v, ok := sswIslandMatch(src, e)
			if !ok {
				panic("ssw: missing miter fanin")
			}
			return v ^ (e & 1)
		}
		for _, e := range aig.Aig_ManDfsAll(src) {
			n := src.Node(e)
			var v aig.Edge
			switch n.Kind {
			case aig.And:
				v = dst.AND(copyLit(n.Fanins[0]), copyLit(n.Fanins[1]))
			case aig.ExclusiveOr:
				v = dst.XOR(copyLit(n.Fanins[0]), copyLit(n.Fanins[1]))
			case aig.Buffer:
				v = copyLit(n.Fanins[0])
			default:
				continue
			}
			src.SetObjectData(e, v)
		}
	}
	copyGraph(p0)
	copyGraph(p1)
	copyOutput := func(src *aig.Graph, e aig.Edge) aig.Edge {
		lit := src.Node(e).Fanins[0]
		v, ok := sswIslandMatch(src, lit)
		if !ok {
			panic("ssw: missing miter output")
		}
		return v ^ (lit & 1)
	}
	outs0, outs1 := p0.Outputs(), p1.Outputs()
	po := int(saig.Saig_ManPoNum(p0))
	for i := 0; i < po; i++ {
		dst.AddOutput(dst.XOR(copyOutput(p0, outs0[i]), copyOutput(p1, outs1[i])))
	}
	for _, e := range outs0[po:] {
		dst.AddOutput(copyOutput(p0, e))
	}
	for _, e := range outs1[po:] {
		dst.AddOutput(copyOutput(p1, e))
	}
	aig.Aig_ManSetRegNum(dst, aig.Aig_ManRegNum(p0)+aig.Aig_ManRegNum(p1))
	return dst
}
