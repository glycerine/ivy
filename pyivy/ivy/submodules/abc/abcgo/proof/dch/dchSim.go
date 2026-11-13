package dch

import aig "github.com/glycerine/abcgo/aig/aig"

func simulateWords(g *aig.Graph, order []aig.Edge, words int, random *aig.AigRandomSource) [][]uint32 {
	count := int(aig.Aig_ManObjNumMax(g))
	sims := make([][]uint32, count)
	storage := make([]uint32, count*words)
	for _, input := range g.Inputs() {
		row := storage[input.NodeID()*words : (input.NodeID()+1)*words]
		sims[input.NodeID()] = row
		for k := range row {
			row[k] = random.Uint32()
		}
		row[0] <<= 1
	}
	for _, edge := range order {
		node := g.Node(edge)
		row := storage[edge.NodeID()*words : (edge.NodeID()+1)*words]
		sims[edge.NodeID()] = row
		switch node.Kind {
		case aig.Constant:
			for k := range row {
				row[k] = ^uint32(0)
			}
		case aig.Input:
			// CI patterns were drawn in interface order above.
		default:
			first := node.Fanins[0]
			for k := range row {
				a := sims[first.NodeID()][k]
				if first.Complemented() {
					a = ^a
				}
				switch node.Kind {
				case aig.Buffer, aig.Output:
					row[k] = a
				case aig.And, aig.ExclusiveOr:
					second := node.Fanins[1]
					b := sims[second.NodeID()][k]
					if second.Complemented() {
						b = ^b
					}
					if node.Kind == aig.And {
						row[k] = a & b
					} else {
						row[k] = a ^ b
					}
				}
			}
		}
	}
	return sims
}
func simulateAssignment(g *aig.Graph, order []aig.Edge, inputs []bool) []bool {
	values := make([]bool, int(aig.Aig_ManObjNumMax(g)))
	for i, edge := range g.Inputs() {
		values[edge.NodeID()] = inputs[i]
	}
	value := func(edge aig.Edge) bool { return values[edge.NodeID()] != edge.Complemented() }
	for _, edge := range order {
		n := g.Node(edge)
		switch n.Kind {
		case aig.Constant:
			values[edge.NodeID()] = true
		case aig.Input:
		case aig.Buffer, aig.Output:
			values[edge.NodeID()] = value(n.Fanins[0])
		case aig.And:
			values[edge.NodeID()] = value(n.Fanins[0]) && value(n.Fanins[1])
		case aig.ExclusiveOr:
			values[edge.NodeID()] = value(n.Fanins[0]) != value(n.Fanins[1])
		}
	}
	return values
}
