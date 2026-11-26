package core

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
)

func btoi(v bool) int {
	if v {
		return 1
	}
	return 0
}

// refactorTryKit mirrors Dar_RefactTryGraph without mutating the AIG.
func refactorTryKit(g *aig.Graph, root aig.Edge, leaves []aig.Edge, candidate *kit.Kit_Graph_t, nodeMax, levelMax int) (int, int, bool) {
	if kit.Kit_GraphIsConst(candidate) != 0 {
		return 0, 0, true
	}
	if kit.Kit_GraphIsVar(candidate) != 0 {
		return 0, 0, true
	}
	count := int(kit.Kit_GraphLeaveNum(candidate) + kit.Kit_GraphNodeNum(candidate))
	images := make([]aig.Edge, count)
	levels := make([]int, count)
	for i := range images {
		images[i] = aig.InvalidEdge
	}
	for i, e := range leaves {
		images[i] = e
		levels[i] = g.Node(e).Level
	}
	added := 0
	for i, node := range kit.Kit_GraphForEachNode(candidate) {
		a, b := kit.Kit_NodeEdges(node)
		ai, bi := int(a.Node), int(b.Node)
		level := 1 + max(levels[ai], levels[bi])
		found := aig.InvalidEdge
		if images[ai] != aig.InvalidEdge && images[bi] != aig.InvalidEdge {
			x, y := images[ai], images[bi]
			if kit.Kit_EdgeToInt(a)&1 != 0 {
				x = x.Not()
			}
			if kit.Kit_EdgeToInt(b)&1 != 0 {
				y = y.Not()
			}
			found = aig.Aig_TableLookupTwo(g, x, y)
			if found != aig.InvalidEdge && found.NodeID() == root.NodeID() {
				return 0, 0, false
			}
			if found != aig.InvalidEdge {
				if found.NodeID() == 0 {
					level = 0
				} else if found.NodeID() == x.NodeID() {
					level = g.Node(x).Level
				} else if found.NodeID() == y.NodeID() {
					level = g.Node(y).Level
				}
			}
		}
		if found == aig.InvalidEdge || aig.Aig_ObjIsTravIdCurrent(g, found) {
			added++
			if added > nodeMax {
				return 0, 0, false
			}
		}
		if level > levelMax {
			return 0, 0, false
		}
		images[i] = found
		levels[i] = level
	}
	rootEdge := kit.Kit_GraphRootEdge(candidate)
	return added, levels[rootEdge.Node], true
}
func refactorBuildKit(g *aig.Graph, leaves []aig.Edge, candidate *kit.Kit_Graph_t) aig.Edge {
	if kit.Kit_GraphIsConst(candidate) != 0 {
		if kit.Kit_GraphIsConst0(candidate) != 0 {
			return g.Const0()
		}
		return g.Const1()
	}
	images := make([]aig.Edge, int(kit.Kit_GraphLeaveNum(candidate)+kit.Kit_GraphNodeNum(candidate)))
	copy(images, leaves)
	for i, node := range kit.Kit_GraphForEachNode(candidate) {
		a, b := kit.Kit_NodeEdges(node)
		x, y := images[a.Node], images[b.Node]
		if kit.Kit_EdgeToInt(a)&1 != 0 {
			x = x.Not()
		}
		if kit.Kit_EdgeToInt(b)&1 != 0 {
			y = y.Not()
		}
		images[i] = g.AND(x, y)
	}
	r := kit.Kit_GraphRootEdge(candidate)
	result := images[r.Node]
	if kit.Kit_EdgeToInt(r)&1 != 0 {
		result = result.Not()
	}
	return result
}

func refactorTruth(g *aig.Graph, root aig.Edge, leaves []aig.Edge) []bool {
	truth := make([]bool, 1<<len(leaves))
	for assignment := range truth {
		values := map[int]bool{0: true}
		for i, e := range leaves {
			values[e.NodeID()] = assignment&(1<<i) != 0
		}
		var evaluate func(aig.Edge) bool
		evaluate = func(e aig.Edge) bool {
			v, ok := values[e.NodeID()]
			if !ok {
				n := g.Node(e)
				a := evaluate(n.Fanins[0])
				switch n.Kind {
				case aig.And:
					v = a && evaluate(n.Fanins[1])
				case aig.ExclusiveOr:
					v = a != evaluate(n.Fanins[1])
				case aig.Buffer, aig.Output:
					v = a
				default:
					panic("dar: incomplete refactor boundary")
				}
				values[e.NodeID()] = v
			}
			return v != e.Complemented()
		}
		truth[assignment] = evaluate(root)
	}
	return truth
}
