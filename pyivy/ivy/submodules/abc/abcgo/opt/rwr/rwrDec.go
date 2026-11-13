package rwr

// Library preprocessing and decomposition traversal from rwrDec.c.
import "github.com/glycerine/abcgo/bool/dec"

func Rwr_ManPreprocess(p *Rwr_Man_t) {
	p.pMapInv = make([]uint16, 222)
	p.vClasses = make([][]*Rwr_Node_t, 222)
	for i := int32(0); i < p.nFuncs; i++ {
		head := p.pTable[i]
		for node := head; node != nil; node = node.pNext {
			assert(node.uTruth == head.uTruth, "class truth matches table")
			class := int(p.pMap[node.uTruth])
			assert(class < 222, "class index")
			p.vClasses[class] = append(p.vClasses[class], node)
			p.pMapInv[class] = p.puCanons[node.uTruth]
		}
	}
	for _, class := range p.vClasses {
		for _, node := range class {
			graph := Rwr_NodePreprocess(p, node)
			// C reuses pNext for a graph pointer after the linked lists are consumed.
			node.graph = graph
			assert(uint16(dec.Dec_GraphDeriveTruth(graph)) == node.uTruth, "decomposition truth")
		}
	}
}
func Rwr_NodePreprocess(p *Rwr_Man_t, node *Rwr_Node_t) *dec.Dec_Graph_t {
	if node.uTruth == 0 {
		return dec.Dec_GraphCreateConst0()
	}
	if node.uTruth == 0x00ff {
		return dec.Dec_GraphCreateLeaf(3, 4, 1)
	}
	graph := dec.Dec_GraphCreate(4)
	Rwr_ManIncTravId(p)
	dec.Dec_GraphSetRoot(graph, Rwr_TravCollect_rec(p, node, graph))
	return graph
}
func Rwr_TravCollect_rec(p *Rwr_Man_t, node *Rwr_Node_t, graph *dec.Dec_Graph_t) dec.Dec_Edge_t {
	if node.fUsed {
		return dec.Dec_EdgeCreate(node.Id-1, 0)
	}
	if node.TravId == p.nTravIds {
		return dec.Dec_IntToEdge(node.Volume)
	}
	node.TravId = p.nTravIds
	a := Rwr_TravCollect_rec(p, node.p0.node, graph)
	if node.p0.complemented {
		a = dec.Dec_IntToEdge(dec.Dec_EdgeToInt(a) ^ 1)
	}
	b := Rwr_TravCollect_rec(p, node.p1.node, graph)
	if node.p1.complemented {
		b = dec.Dec_IntToEdge(dec.Dec_EdgeToInt(b) ^ 1)
	}
	var out dec.Dec_Edge_t
	if node.fExor {
		out = dec.Dec_GraphAddNodeXor(graph, a, b, 0)
	} else {
		out = dec.Dec_GraphAddNodeAnd(graph, a, b)
	}
	node.Volume = dec.Dec_EdgeToInt(out) & 0xff
	return out
}
