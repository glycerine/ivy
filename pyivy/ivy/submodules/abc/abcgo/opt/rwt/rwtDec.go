package rwt

import "github.com/glycerine/abcgo/bool/dec"

func Rwt_ManPreprocess(p *Rwt_Man_t) {
	if p == nil {
		return
	}
	p.pMapInv = make([]uint16, 222)
	p.vClasses = make([][]*Rwt_Node_t, 222)
	p.nodeGraphs = make(map[*Rwt_Node_t]*dec.Dec_Graph_t)
	for truth, head := range p.pTable {
		for node := head; node != nil; node = node.pNext {
			class := int(p.pMap[node.uTruth])
			if class >= len(p.vClasses) {
				panic("RWT NPN class is out of range")
			}
			p.vClasses[class] = append(p.vClasses[class], node)
			p.pMapInv[class] = p.puCanons[truth]
		}
	}
	for _, class := range p.vClasses {
		for _, node := range class {
			graph := Rwt_NodePreprocess(p, node)
			p.nodeGraphs[node] = graph
		}
	}
}

func Rwt_NodePreprocess(p *Rwt_Man_t, node *Rwt_Node_t) *dec.Dec_Graph_t {
	node = Rwt_Regular(node)
	if node == nil {
		return nil
	}
	if node.uTruth == 0 {
		return dec.Dec_GraphCreateConst0()
	}
	if node.uTruth == 0x00ff {
		return dec.Dec_GraphCreateLeaf(3, 4, 1)
	}
	graph := dec.Dec_GraphCreate(4)
	Rwt_ManIncTravId(p)
	root := Rwt_TravCollect_rec(p, node, graph)
	dec.Dec_GraphSetRoot(graph, root)
	return graph
}

func Rwt_TravCollect_rec(p *Rwt_Man_t, node *Rwt_Node_t, graph *dec.Dec_Graph_t) dec.Dec_Edge_t {
	node = Rwt_Regular(node)
	if node == nil {
		panic("RWT decomposition encountered a nil node")
	}
	if node.fUsed {
		return dec.Dec_EdgeCreate(node.Id-1, 0)
	}
	if node.TravId == p.nTravIds {
		return dec.Dec_IntToEdge(node.Volume)
	}
	node.TravId = p.nTravIds
	e0 := Rwt_TravCollect_rec(p, node.p0, graph)
	if Rwt_IsComplement(node.p0) != 0 {
		e0 = dec.Dec_IntToEdge(dec.Dec_EdgeToInt(e0) ^ 1)
	}
	e1 := Rwt_TravCollect_rec(p, node.p1, graph)
	if Rwt_IsComplement(node.p1) != 0 {
		e1 = dec.Dec_IntToEdge(dec.Dec_EdgeToInt(e1) ^ 1)
	}
	var edge dec.Dec_Edge_t
	if node.fExor {
		edge = dec.Dec_GraphAddNodeXor(graph, e0, e1, 0)
	} else {
		edge = dec.Dec_GraphAddNodeAnd(graph, e0, e1)
	}
	node.Volume = dec.Dec_EdgeToInt(edge)
	return edge
}
