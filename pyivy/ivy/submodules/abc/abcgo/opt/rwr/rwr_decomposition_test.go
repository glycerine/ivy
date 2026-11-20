package rwr

import "github.com/glycerine/abcgo/bool/dec"

// rwrDecomposition implements Rwr_NodePreprocess and Rwr_TravCollect_rec.
func rwrDecomposition(forest []rwrForestNode, root int) *dec.Dec_Graph_t {
	if root < 0 || root >= len(forest) {
		panic("rwr: invalid root")
	}
	if forest[root].truth == 0 {
		return dec.Dec_GraphCreateConst0()
	}
	if forest[root].truth == 0x00ff {
		return dec.Dec_GraphCreateLeaf(3, 4, 1)
	}
	graph := dec.Dec_GraphCreate(4)
	memo := make(map[int]dec.Dec_Edge_t)
	var collect func(int) dec.Dec_Edge_t
	collect = func(id int) dec.Dec_Edge_t {
		if id < 5 {
			if id == 0 {
				panic("rwr: constant leaf in nonconstant decomposition")
			}
			return dec.Dec_EdgeCreate(int32(id-1), 0)
		}
		if edge, ok := memo[id]; ok {
			return edge
		}
		n := forest[id]
		a, b := collect(n.a.node), collect(n.b.node)
		if n.a.compl {
			a = dec.Dec_IntToEdge(dec.Dec_EdgeToInt(a) ^ 1)
		}
		if n.b.compl {
			b = dec.Dec_IntToEdge(dec.Dec_EdgeToInt(b) ^ 1)
		}
		var edge dec.Dec_Edge_t
		if n.xor {
			edge = dec.Dec_GraphAddNodeXor(graph, a, b, 0)
		} else {
			edge = dec.Dec_GraphAddNodeAnd(graph, a, b)
		}
		memo[id] = edge
		return edge
	}
	dec.Dec_GraphSetRoot(graph, collect(root))
	return graph
}
