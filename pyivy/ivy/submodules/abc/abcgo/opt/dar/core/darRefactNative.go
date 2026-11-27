package core

// A line-by-line port of ABC's Dar_ManRefactor and its helpers
// (Dar_ManRefactorTryCuts, Dar_RefactTryGraph, Dar_RefactBuildGraph in
// darRefact.c), so that cut selection, factoring and gain/level tie-breaks
// follow ABC's order exactly.

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/misc/util"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
)

const darInfinity = 1<<30 - 1 // ABC_INFINITY

// darRefactorer holds Ref_Man_t's per-pass state.
type darRefactorer struct {
	g        *aig.Graph
	p        *RefactorParams
	cutNodes []aig.Edge
	memory   []uint32

	leavesBest []aig.Edge
	graphBest  *kit.Kit_Graph_t
	gainBest   int32
	levelBest  int32 // Kit_GraphRootLevel(graphBest)
}

// darRefactTryGraph is Dar_RefactTryGraph. It returns the number of nodes
// added (or -1) and the level of the graph's root.
func (r *darRefactorer) darRefactTryGraph(root aig.Edge, cut []aig.Edge, graph *kit.Kit_Graph_t, nodeMax, levelMax int32) (int32, int32) {
	g := r.g
	if kit.Kit_GraphIsConst(graph) != 0 || kit.Kit_GraphIsVar(graph) != 0 {
		return 0, 0
	}
	count := int(kit.Kit_GraphLeaveNum(graph) + kit.Kit_GraphNodeNum(graph))
	funcs := make([]aig.Edge, count)
	levels := make([]int32, count)
	for i := range funcs {
		funcs[i] = aig.InvalidEdge
	}
	for i := range kit.Kit_GraphLeaveNum(graph) {
		funcs[i] = cut[i]
		levels[i] = aig.Aig_ObjLevel(g, cut[i])
	}
	var counter int32
	for i, node := range kit.Kit_GraphForEachNode(graph) {
		e0, e1 := kit.Kit_NodeEdges(node)
		and0, and1 := funcs[e0.Node], funcs[e1.Node]
		and := aig.InvalidEdge
		if and0 != aig.InvalidEdge && and1 != aig.InvalidEdge {
			if kit.Kit_EdgeToInt(e0)&1 != 0 {
				and0 = and0.Not()
			}
			if kit.Kit_EdgeToInt(e1)&1 != 0 {
				and1 = and1.Not()
			}
			and = aig.Aig_TableLookupTwo(g, and0, and1)
			if and != aig.InvalidEdge && and.NodeID() == root.NodeID() {
				return -1, 0
			}
		}
		if and == aig.InvalidEdge || aig.Aig_ObjIsTravIdCurrent(g, aig.Aig_Regular(and)) {
			counter++
			if counter > nodeMax {
				return -1, 0
			}
		}
		levelNew := 1 + max(levels[e0.Node], levels[e1.Node])
		if and != aig.InvalidEdge {
			switch and.NodeID() {
			case 0:
				levelNew = 0
			case and0.NodeID():
				levelNew = aig.Aig_ObjLevel(g, and0)
			case and1.NodeID():
				levelNew = aig.Aig_ObjLevel(g, and1)
			}
		}
		if levelNew > levelMax {
			return -1, 0
		}
		funcs[i], levels[i] = and, levelNew
	}
	return counter, levels[kit.Kit_GraphRootEdge(graph).Node]
}

// darRefactBuildGraph is Dar_RefactBuildGraph.
func darRefactBuildGraph(g *aig.Graph, cut []aig.Edge, graph *kit.Kit_Graph_t) aig.Edge {
	if kit.Kit_GraphIsConst(graph) != 0 {
		e := g.Const1()
		if kit.Kit_GraphIsComplement(graph) != 0 {
			e = e.Not()
		}
		return e
	}
	funcs := make([]aig.Edge, int(kit.Kit_GraphLeaveNum(graph)+kit.Kit_GraphNodeNum(graph)))
	copy(funcs, cut)
	root := kit.Kit_GraphRootEdge(graph)
	if kit.Kit_GraphIsVar(graph) == 0 {
		for i, node := range kit.Kit_GraphForEachNode(graph) {
			e0, e1 := kit.Kit_NodeEdges(node)
			a, b := funcs[e0.Node], funcs[e1.Node]
			if kit.Kit_EdgeToInt(e0)&1 != 0 {
				a = a.Not()
			}
			if kit.Kit_EdgeToInt(e1)&1 != 0 {
				b = b.Not()
			}
			funcs[i] = g.AND(a, b)
		}
	}
	e := funcs[root.Node]
	if kit.Kit_GraphIsComplement(graph) != 0 {
		e = e.Not()
	}
	return e
}

// darManRefactorTryCuts is Dar_ManRefactorTryCuts.
func (r *darRefactorer) darManRefactorTryCuts(obj aig.Edge, cuts [][]aig.Edge, saved, required int32) {
	g := r.g
	r.gainBest, r.graphBest = -1, nil
	nodeMax := saved
	if !r.p.UseZeros {
		nodeMax--
	}
	for _, cut := range cuts {
		if len(cut) == 0 {
			continue
		}
		n := int32(len(cut))
		aig.Aig_ObjCollectCut(g, obj, cut, &r.cutNodes)
		truth, err := aig.Aig_ManCutTruth(g, obj, cut)
		if err != nil {
			panic(err)
		}
		if kit.Kit_TruthIsConst0(truth, n) != 0 || kit.Kit_TruthIsConst1(truth, n) != 0 {
			r.gainBest = aig.Aig_NodeMffcSupp(g, obj, 0, nil)
			if kit.Kit_TruthIsConst0(truth, n) != 0 {
				r.graphBest = kit.Kit_GraphCreateConst0()
			} else {
				r.graphBest = kit.Kit_GraphCreateConst1()
			}
			r.leavesBest = append(r.leavesBest[:0], cut...)
			return
		}
		for phase := int32(0); phase < 2; phase++ {
			if phase == 1 {
				kit.Kit_TruthNot(truth, truth, n)
			}
			if kit.Kit_TruthIsop(truth, n, &r.memory, 0) <= -1 {
				continue
			}
			cur := kit.Kit_SopFactor(r.memory, phase, n, &r.memory)
			added, level := r.darRefactTryGraph(obj, cut, cur, nodeMax, required)
			if added <= -1 {
				continue
			}
			gain := saved - added
			if r.gainBest < gain || r.gainBest == gain && (kit.Kit_GraphIsConst(cur) != 0 || level < r.levelBest) {
				r.gainBest, r.graphBest, r.levelBest = gain, cur, level
				r.leavesBest = append(r.leavesBest[:0], cut...)
			}
		}
	}
}

// darObjCutLevelAchieved is Dar_ObjCutLevelAchieved.
func darObjCutLevelAchieved(g *aig.Graph, cut []aig.Edge, levelMin int) bool {
	for _, e := range cut {
		if n := g.Node(e); n.Kind != aig.Input && n.Level <= levelMin {
			return true
		}
	}
	return false
}

// Dar_ManRefactor is ABC's Dar_ManRefactor.
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
	r := &darRefactorer{g: g, p: &p}
	aig.Aig_ManCleanup(g)
	aig.Aig_ManFanoutStart(g)
	if p.UpdateLevel {
		aig.Aig_ManStartReverseLevels(g, 0)
	}
	var cut, cut2 []aig.Edge
	var st refactorStats // Ref_Man_t's counters for Dar_ManRefPrintStats
	clkStart := time.Now()
	st.nodesInit = int(aig.Aig_ManNodeNum(g))
	nNodesOld := aig.Aig_ManObjNumMax(g)
	for i := int32(0); i < aig.Aig_ManObjNumMax(g); i++ {
		obj := aig.Edge(i) << 1
		if !g.ValidEdge(obj) || g.Kind(obj) != aig.And {
			continue
		}
		if i > nNodesOld {
			break
		}
		if p.Context != nil && i&256 == 0 {
			if err := p.Context.Err(); err != nil {
				aig.Aig_ManFanoutStop(g)
				if p.UpdateLevel {
					aig.Aig_ManStopReverseLevels(g)
				}
				return err
			}
		}
		cut, cut2 = cut[:0], cut2[:0]
		clk := time.Now()
		levelMin := max(0, g.Node(obj).Level-10)
		saved := aig.Aig_NodeMffcSupp(g, obj, int32(levelMin), &cut)
		if int(saved) < p.MffcMin {
			st.timeCuts += time.Since(clk)
			continue
		}
		st.nodesTried++
		if len(cut) > p.LeafMax {
			var visited []aig.Edge
			aig.Aig_ManFindCut(g, obj, &cut, &visited, int32(p.LeafMax), 50)
			saved = aig.Aig_NodeMffcLabelCut(g, obj, cut)
		} else if len(cut) < p.LeafMax-2 && p.Extend {
			if !darObjCutLevelAchieved(g, cut, levelMin) {
				if aig.Aig_NodeMffcExtendCut(g, obj, cut, &cut2) {
					aig.Aig_NodeMffcLabelCut(g, obj, cut)
				}
				if len(cut2) > p.LeafMax {
					cut2 = cut2[:0]
				}
				if len(cut2) > 0 {
					st.nodesExten++
				}
			} else {
				st.nodesBelow++
			}
		}
		st.timeCuts += time.Since(clk)
		clk = time.Now()
		required := int32(darInfinity)
		if p.UpdateLevel {
			required = aig.Aig_ObjRequiredLevel(g, obj)
		}
		r.darManRefactorTryCuts(obj, [][]aig.Edge{cut, cut2}, saved, required)
		st.timeEval += time.Since(clk)
		if !(r.gainBest > 0 || r.gainBest == 0 && p.UseZeros) {
			continue
		}
		repl := darRefactBuildGraph(g, r.leavesBest, r.graphBest)
		aig.Aig_ObjReplace(g, obj, repl, btoi32(p.UpdateLevel))
		st.cutsUsed++
	}
	st.timeTotal = time.Since(clkStart)
	aig.Aig_ManFanoutStop(g)
	if p.UpdateLevel {
		aig.Aig_ManStopReverseLevels(g)
	}
	if p.Verbose { // Dar_ManRefStop
		st.print(g)
	}
	aig.Aig_ManCleanup(g)
	return g.Check()
}

type refactorStats struct {
	nodesInit, nodesTried, nodesBelow, nodesExten, cutsUsed int
	timeCuts, timeEval, timeTotal                           time.Duration
}

// print is Dar_ManRefPrintStats.
func (st *refactorStats) print(g *aig.Graph) {
	nodes := int(aig.Aig_ManNodeNum(g))
	gain := st.nodesInit - nodes
	fmt.Printf("NodesBeg = %8d. NodesEnd = %8d. Gain = %6d. (%6.2f %%).\n", st.nodesInit, nodes, gain, 100.0*float64(gain)/float64(st.nodesInit))
	fmt.Printf("Tried = %6d. Below = %5d. Extended = %5d.  Used = %5d.  Levels = %4d.\n", st.nodesTried, st.nodesBelow, st.nodesExten, st.cutsUsed, aig.Aig_ManLevels(g))
	util.Abc_PrintTime(1, "Cuts  ", st.timeCuts.Nanoseconds())
	util.Abc_PrintTime(1, "Eval  ", st.timeEval.Nanoseconds())
	util.Abc_PrintTime(1, "Other ", (st.timeTotal - st.timeCuts - st.timeEval).Nanoseconds())
	util.Abc_PrintTime(1, "TOTAL ", st.timeTotal.Nanoseconds())
}
