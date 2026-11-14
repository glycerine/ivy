package gia

import "github.com/glycerine/abcgo/proof/cec/core"

// SourceCNFGraph supplies the source CEC encoder with GIA metadata without
// making the encoder import GIA. The graph must remain alive and immutable
// in topology for the lifetime of its SAT manager.
func (g *Gia_Man_t) SourceCNFGraph() core.SourceCNFGraph {
	return core.SourceCNFGraph{
		Nodes: g.ObjectCount(),
		Node: func(id int32) core.SourceCNFNode {
			o := Gia_ManObj(g, id)
			n := core.SourceCNFNode{And: Gia_ObjIsAnd(o) != 0, Input: Gia_ObjIsCi(o) != 0, Mux: Gia_ObjIsMuxType(o) != 0, Phase: Gia_ObjPhase(o), Value: Gia_ObjValue(o)}
			if n.And {
				n.Fanins = [2]int32{Gia_ObjFaninLit0(o, id), Gia_ObjFaninLit1(o, id)}
			}
			return n
		},
		Mux: func(id int32) core.SourceCNFMux {
			o := Gia_ManObj(g, id)
			a, b := Gia_ObjFanin0(o), Gia_ObjFanin1(o)
			var m core.SourceCNFMux
			m.Raw = [4]int32{Gia_ObjId(g, Gia_ObjFanin0(a)), Gia_ObjId(g, Gia_ObjFanin0(b)), Gia_ObjId(g, Gia_ObjFanin1(a)), Gia_ObjId(g, Gia_ObjFanin1(b))}
			m.Control = Gia_ObjRecognizeMuxLits(g, o, &m.Then, &m.Else)
			return m
		},
	}
}
