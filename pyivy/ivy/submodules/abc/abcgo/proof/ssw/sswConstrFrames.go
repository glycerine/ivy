package ssw

import "github.com/glycerine/abcgo/aig/aig"

// Ssw_FramesWithConstraints unrolls from zero state and emits inverted constraint outputs.
func Ssw_FramesWithConstraints(g *aig.Graph, nFrames int32) *aig.Graph {
	inputs, outputs := g.Inputs(), g.Outputs()
	regs := int(g.RegisterCount())
	if regs <= 0 || regs >= len(inputs) || nFrames < 0 {
		panic("ssw: invalid constrained unrolling")
	}
	target := aig.NewGraph()
	nodes := aig.Aig_ManDfsAll(g)
	mapped := make([]aig.Edge, len(g.Nodes()))
	for i := range mapped {
		mapped[i] = aig.InvalidEdge
	}
	for _, lo := range inputs[len(inputs)-regs:] {
		mapped[lo.NodeID()] = target.Const0()
	}
	edgeCopy := func(e aig.Edge) aig.Edge {
		v := mapped[e.NodeID()]
		if v == aig.InvalidEdge {
			panic("ssw: unmapped frame fanin")
		}
		if e.Complemented() {
			v = v.Not()
		}
		return v
	}
	nPO := len(outputs) - regs
	nConstr := int(aig.Aig_ManConstrNum(g))
	for f := int32(0); f < nFrames; f++ {
		mapped[0] = target.Const1()
		for _, pi := range inputs[:len(inputs)-regs] {
			mapped[pi.NodeID()] = target.NewInput()
		}
		for _, obj := range nodes {
			n := g.Node(obj)
			switch n.Kind {
			case aig.And:
				mapped[obj.NodeID()] = target.AND(edgeCopy(n.Fanins[0]), edgeCopy(n.Fanins[1]))
			case aig.ExclusiveOr:
				mapped[obj.NodeID()] = target.NewXor(edgeCopy(n.Fanins[0]), edgeCopy(n.Fanins[1]))
			case aig.Buffer:
				mapped[obj.NodeID()] = edgeCopy(n.Fanins[0])
			}
		}
		for _, out := range outputs {
			mapped[out.NodeID()] = edgeCopy(g.Node(out).Fanins[0])
		}
		for _, out := range outputs[nPO-nConstr : nPO] {
			target.AddOutput(mapped[out.NodeID()].Not())
		}
		for i, lo := range inputs[len(inputs)-regs:] {
			mapped[lo.NodeID()] = mapped[outputs[nPO+i].NodeID()]
		}
	}
	aig.Aig_ManCleanup(target)
	return target
}
