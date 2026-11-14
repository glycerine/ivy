package core

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
	"math"
)

// Dar_ManRefactor follows DAR's bounded MFFC cuts, ISOP factoring in both
// output phases, and area/required-level candidate selection.
func Dar_ManRefactor(g *aig.Graph, params *RefactorParams) error {
	p := DefaultRefactorParams()
	if params != nil {
		p = *params
	}
	if g == nil {
		return fmt.Errorf("dar: nil graph")
	}
	if p.LeafMax < 2 || p.LeafMax > 15 || p.CutsMax < 1 || p.MffcMin < 0 {
		return fmt.Errorf("dar: invalid refactor limits")
	}
	if err := g.Check(); err != nil {
		return err
	}
	aig.Aig_ManCleanup(g)
	aig.Aig_ManFanoutStart(g)
	defer aig.Aig_ManFanoutStop(g)
	if p.UpdateLevel {
		aig.Aig_ManStartReverseLevels(g, 0)
		defer aig.Aig_ManStopReverseLevels(g)
	}
	for _, root := range g.LiveEdges() {
		if p.Context != nil {
			if err := p.Context.Err(); err != nil {
				return err
			}
		}
		if !g.ValidEdge(root) || g.Node(root).Kind != aig.And {
			continue
		}
		levelMin := max(0, g.Node(root).Level-10)
		var cut []aig.Edge
		saved := int(aig.Aig_NodeMffcSupp(g, root, int32(levelMin), &cut))
		if saved < p.MffcMin {
			continue
		}
		cuts := [][]aig.Edge{cut}
		if len(cut) > p.LeafMax {
			var visited []aig.Edge
			aig.Aig_ManFindCut(g, root, &cut, &visited, int32(p.LeafMax), 50)
			saved = int(aig.Aig_NodeMffcLabelCut(g, root, cut))
			cuts[0] = cut
		} else if len(cut) < p.LeafMax-2 && p.Extend {
			reached := false
			for _, leaf := range cut {
				if g.Node(leaf).Kind != aig.Input && g.Node(leaf).Level <= levelMin {
					reached = true
					break
				}
			}
			if !reached {
				var extended []aig.Edge
				if aig.Aig_NodeMffcExtendCut(g, root, cut, &extended) && len(extended) <= p.LeafMax && len(extended) > 0 {
					aig.Aig_NodeMffcLabelCut(g, root, cut)
					cuts = append(cuts, extended)
				}
			}
		}
		required := math.MaxInt32
		if p.UpdateLevel {
			required = int(aig.Aig_ObjRequiredLevel(g, root))
		}
		gainBest, bestLevel := -1, math.MaxInt32
		var best *kit.Kit_Graph_t
		var bestCut []aig.Edge
		for _, leaves := range cuts {
			if len(leaves) == 0 {
				continue
			}
			truth := refactorTruth(g, root, leaves)
			constant := true
			for _, bit := range truth {
				if bit != truth[0] {
					constant = false
					break
				}
			}
			if constant {
				gainBest = int(aig.Aig_NodeMffcSupp(g, root, 0, nil))
				if truth[0] {
					best = kit.Kit_GraphCreateConst1()
				} else {
					best = kit.Kit_GraphCreateConst0()
				}
				bestCut = append([]aig.Edge(nil), leaves...)
				break
			}
			words := make([]uint32, kit.Kit_TruthWordNum(int32(len(leaves))))
			for i := 0; i < len(words)*32; i++ {
				if truth[i%len(truth)] {
					words[i/32] |= 1 << uint(i%32)
				}
			}
			for phase := 0; phase < 2; phase++ {
				if phase == 1 {
					for i := range words {
						words[i] = ^words[i]
					}
				}
				var cover []uint32
				if kit.Kit_TruthIsop(words, int32(len(leaves)), &cover, 0) < 0 {
					continue
				}
				candidate := kit.Kit_SopFactor(cover, int32(phase), int32(len(leaves)), &cover)
				cost, level, ok := refactorTryKit(g, root, leaves, candidate, saved-btoi(!p.UseZeros), required)
				if ok && (saved-cost > gainBest || saved-cost == gainBest && level < bestLevel) {
					gainBest, bestLevel, best, bestCut = saved-cost, level, candidate, append([]aig.Edge(nil), leaves...)
				}
			}
		}
		if best == nil || gainBest < 0 || gainBest == 0 && !p.UseZeros {
			continue
		}
		replacement := refactorBuildKit(g, bestCut, best)
		if graphDependsOn(g, replacement, root) {
			aig.Aig_ManCleanup(g)
			continue
		}
		aig.Aig_ObjReplace(g, root, replacement, 0)
		aig.Aig_ManCleanup(g)
		if p.UpdateLevel {
			aig.Aig_ManUpdateReverseLevel(g, g.Const1())
		}
	}
	return g.Check()
}
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
