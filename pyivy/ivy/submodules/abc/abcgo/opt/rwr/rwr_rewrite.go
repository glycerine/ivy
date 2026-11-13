package rwr

import (
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra/npn"
)

// rwrNetwork is the portion of the ABC network used by Rwr_NodeRewrite.
// MarkBoundary must increment or decrement each regular leaf's fanout count;
// LabelMFFC increments the network traversal ID and labels the root's cone.
type rwrNetwork interface {
	Lookup(id int32) (any, bool)
	Complement(any) any
	FanoutCount(any) int
	MarkBoundary([]any, int)
	LabelMFFC(root any) int32
	CountGraph(root any, graph *dec.Dec_Graph_t, saved, levelMax int32) int32
}

type rwrCut struct {
	truth  uint16
	leaves []int32
}

type rwrChoice struct {
	graph      *dec.Dec_Graph_t
	fanins     []any
	complement bool
	gain       int32
}

// rwrRewriteCuts ports the cut loop of Rwr_NodeRewrite. The caller supplies
// cuts following the trivial self cut, matching C's pCut = pCut->pNext.
func (p *rwrLibrary) rwrRewriteCuts(net rwrNetwork, root any, cuts []rwrCut, required int32, useZeros bool) rwrChoice {
	p.nNodesConsidered++
	result := rwrChoice{gain: -1}
	var bestTruth uint16
	perms := npn.Permutations()
	for _, cut := range cuts {
		if len(cut.leaves) < 4 {
			continue
		}
		if len(cut.leaves) != 4 {
			panic("rwr: expected four-leaf cut")
		}
		phase := p.phases[cut.truth]
		perm := perms[p.permutations[cut.truth]]
		fanins := make([]any, 4)
		valid, single := true, 0
		for i := range fanins {
			fanin, ok := net.Lookup(cut.leaves[perm[i]])
			if !ok {
				valid = false
				break
			}
			if net.FanoutCount(fanin) == 1 {
				single++
			}
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
		if single > 2 {
			continue
		}
		net.MarkBoundary(fanins, 1)
		saved := net.LabelMFFC(root)
		net.MarkBoundary(fanins, -1)
		graph, gain := p.rwrEvaluateCut(cut.truth, fanins, saved, func(graph *dec.Dec_Graph_t) int32 {
			return net.CountGraph(root, graph, saved, required)
		})
		if graph != nil && gain > result.gain {
			result = rwrChoice{graph, append([]any(nil), fanins...), phase&16 != 0, gain}
			bestTruth = cut.truth
		}
	}
	if result.graph != nil {
		for i, fanin := range result.fanins {
			dec.Dec_GraphSetLeafFunc(result.graph, int32(i), fanin)
		}
		p.classScores[p.classMap[bestTruth]]++
		p.nNodesGained += int(result.gain)
		if useZeros || result.gain > 0 {
			p.nNodesRewritten++
		}
	}
	return result
}
