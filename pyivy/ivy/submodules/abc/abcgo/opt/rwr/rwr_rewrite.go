package rwr

import (
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/opt/cut"
)

// rwrNetwork is the portion of the ABC network used by Rwr_NodeRewrite.
// MarkBoundary must increment or decrement each regular leaf's fanout count;
// LabelMFFC increments the network traversal ID and labels the root's cone.
// Rwr_Network is the concrete-network boundary used by the ordered rewrite
// pass. Implementations adapt node lookup, fanouts, MFFC labeling, and graph
// counting without introducing a dependency from opt/rwr to a network package.
type Rwr_Network interface {
	Lookup(id int32) (any, bool)
	Complement(any) any
	FanoutCount(any) int
	MarkBoundary([]any, int)
	LabelMFFC(root any) int32
	CountGraph(root any, graph *dec.Dec_Graph_t, saved, levelMax int32) int32
}

// Rwr_Cut describes one four-input cut. Leaves are network object IDs.
type Rwr_Cut struct {
	Truth  uint16
	Leaves []int32
}

// The concrete network owns a linked cut list. The public adapter supplies
// the same traversal order in a slice. Both are borrowed for this call.
type rewriteCutIterator struct {
	ordered []Rwr_Cut
	list    *cut.Cut_Cut_t
}

func (cuts *rewriteCutIterator) next() (Rwr_Cut, bool) {
	if len(cuts.ordered) != 0 {
		c := cuts.ordered[0]
		cuts.ordered = cuts.ordered[1:]
		return c, true
	}
	for cuts.list != nil {
		c := cuts.list
		cuts.list = c.Next
		if cut.Cut_CutReadLeaveNum(c) == 4 {
			if truth := cut.Cut_CutReadTruth(c); len(truth) != 0 {
				return Rwr_Cut{uint16(truth[0]), cut.Cut_CutReadLeaves(c)}, true
			}
		}
	}
	return Rwr_Cut{}, false
}

func (p *Rwr_Man_t) rewriteCuts(net Rwr_Network, root any, cuts rewriteCutIterator, required int32, useZeros bool, fanins *[4]any) int32 {
	if p == nil {
		panic("rwr: nil rewrite manager")
	}
	if p.vClasses == nil {
		// The Go API also accepts a zero-value or precompute-only manager.
		normal := Rwr_ManStart(0)
		p.vClasses, p.pMapInv = normal.vClasses, normal.pMapInv
		p.pMap, p.pPerms, p.pPhases, p.pPerms4 = normal.pMap, normal.pPerms, normal.pPhases, normal.pPerms4
	}
	p.nNodesConsidered++
	var bestGraph *dec.Dec_Graph_t
	bestGain := int32(-1)
	var bestFanins [4]any
	var bestTruth uint16
	var regular [4]any
	for cut, ok := cuts.next(); ok; cut, ok = cuts.next() {
		if len(cut.Leaves) < 4 {
			continue
		}
		if len(cut.Leaves) != 4 {
			panic("rwr: expected four-leaf cut")
		}
		phase := p.pPhases[cut.Truth]
		perm := p.pPerms4[p.pPerms[cut.Truth]]
		valid := true
		for i := range fanins {
			fanin, ok := net.Lookup(cut.Leaves[perm[i]])
			if !ok {
				valid = false
				break
			}
			regular[i] = fanin
			if phase&(1<<i) != 0 {
				fanin = net.Complement(fanin)
			}
			fanins[i] = fanin
		}
		if !valid {
			p.nCutsBad++
			continue
		}
		p.nCutsGood++
		single := 0
		for _, fanin := range regular {
			if net.FanoutCount(fanin) == 1 {
				single++
			}
		}
		if single > 2 {
			continue
		}
		net.MarkBoundary(fanins[:], 1)
		saved := net.LabelMFFC(root)
		net.MarkBoundary(fanins[:], -1)
		graph, gain := p.rwrEvaluateCut(cut.Truth, fanins[:], saved, func(graph *dec.Dec_Graph_t) int32 {
			return net.CountGraph(root, graph, saved, required)
		})
		if graph != nil && gain > bestGain {
			bestGraph, bestGain, bestFanins = graph, gain, *fanins
			bestTruth = cut.Truth
		}
	}
	if bestGraph == nil {
		return -1
	}
	for i, fanin := range bestFanins {
		dec.Dec_GraphSetLeafFunc(bestGraph, int32(i), fanin)
	}
	p.vFanins = append(p.vFanins[:0], bestFanins[:]...)
	p.pGraph = bestGraph
	p.fCompl = int32(p.pPhases[bestTruth] >> 4 & 1)
	p.nScores[p.pMap[bestTruth]]++
	p.nNodesGained += bestGain
	if useZeros || bestGain > 0 {
		p.nNodesRewritten++
	}
	return bestGain
}
