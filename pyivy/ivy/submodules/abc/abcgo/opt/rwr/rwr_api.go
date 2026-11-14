package rwr

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/opt/cut"
)

// Rwr_NodeRewrite runs the source algorithm over the caller's cuts in order.
// The network adapter keeps this package independent of ABC's concrete graph
// types. fPlaceEnable is accepted for source API parity; placement scoring is
// disabled in the reference implementation's active evaluation path.
func Rwr_RewriteCuts(p *Rwr_Man_t, net Rwr_Network, root any, cuts []Rwr_Cut, required int32, fUseZeros, fPlaceEnable int32) int32 {
	_ = fPlaceEnable
	if p == nil {
		panic("rwr: nil rewrite manager")
	}
	if p.rewriteLibrary == nil {
		p.rewriteLibrary = rwrStartLibrary()
	}
	lib := p.rewriteLibrary
	choice := lib.rwrRewriteCuts(net, root, cuts, required, fUseZeros != 0)
	p.nNodesConsidered = int32(lib.nNodesConsidered)
	p.nNodesRewritten = int32(lib.nNodesRewritten)
	p.nNodesGained = int32(lib.nNodesGained)
	p.nCutsGood = int32(lib.nCutsGood)
	p.nCutsBad = int32(lib.nCutsBad)
	p.nSubgraphs = int32(lib.nSubgraphs)
	for i, score := range lib.classScores {
		p.nScores[i] = int32(score)
	}
	if choice.graph == nil {
		return -1
	}
	p.pGraph = choice.graph
	p.fCompl = 0
	if choice.complement {
		p.fCompl = 1
	}
	p.vFanins = append(p.vFanins[:0], choice.fanins...)
	return choice.gain
}

// Rwr_NodeRewrite preserves the source entry point while adapting ABC objects
// and cut lists to the network independent ordered rewrite implementation.
func Rwr_NodeRewrite(p *Rwr_Man_t, pManCut *cut.Cut_Man_t, pNode *abc.Object, fUpdateLevel, fUseZeros, fPlaceEnable int32) int32 {
	if pManCut == nil || pNode == nil {
		return -1
	}
	network := &abcRewriteNetwork{network: pNode.Network()}
	var cuts []Rwr_Cut
	for c := cut.Cut_NodeReadCutsNew(pManCut, int32(pNode.ID)); c != nil; c = c.Next {
		leaves := cut.Cut_CutReadLeaves(c)
		truth := cut.Cut_CutReadTruth(c)
		if len(truth) == 0 {
			continue
		}
		cuts = append(cuts, Rwr_Cut{Truth: uint16(truth[0]), Leaves: append([]int32(nil), leaves...)})
	}
	required := int32(1<<30 - 1)
	if fUpdateLevel != 0 {
		required = abc.Abc_ObjRequiredLevel(pNode)
	}
	return Rwr_RewriteCuts(p, network, pNode.Signal(), cuts, required, fUseZeros, fPlaceEnable)
}

type abcRewriteNetwork struct {
	network *abc.Network
	leaves  []*abc.Object
}

func (n *abcRewriteNetwork) Lookup(id int32) (any, bool) {
	o := n.network.Object(int(id))
	if o == nil {
		return nil, false
	}
	return o.Signal(), true
}
func (n *abcRewriteNetwork) Complement(value any) any { return value.(abc.Edge).Not() }
func (n *abcRewriteNetwork) FanoutCount(value any) int {
	return len(value.(abc.Edge).Object.Fanouts())
}
func (n *abcRewriteNetwork) MarkBoundary(values []any, delta int) {
	if delta > 0 {
		n.leaves = make([]*abc.Object, len(values))
		for i, value := range values {
			n.leaves[i] = value.(abc.Edge).Object
		}
	} else {
		n.leaves = nil
	}
}
func (n *abcRewriteNetwork) LabelMFFC(root any) int32 {
	return abc.Abc_NodeMffcLabelWithLeaves(root.(abc.Edge).Object, n.leaves, nil)
}
func (n *abcRewriteNetwork) CountGraph(root any, graph *dec.Dec_Graph_t, saved, levelMax int32) int32 {
	return dec.Dec_GraphToNetworkCount(abc.DecNetwork{Network: n.network}, root, graph, saved, levelMax)
}

// Rwr_ScoresClean clears per-library rewrite scoring counters.
func Rwr_ScoresClean(p *Rwr_Man_t) {
	for i := range p.nScores {
		p.nScores[i] = 0
	}
	if p.rewriteLibrary != nil {
		for i := range p.rewriteLibrary.classScores {
			p.rewriteLibrary.classScores[i] = 0
		}
		for i := range p.rewriteLibrary.scores {
			p.rewriteLibrary.scores[i] = 0
			p.rewriteLibrary.gains[i] = 0
			p.rewriteLibrary.added[i] = 0
		}
	}
}

// Rwr_ScoresReport reports NPN classes selected by the rewrite pass.
func Rwr_ScoresReport(p *Rwr_Man_t) {
	for class, score := range p.nScores {
		if score != 0 {
			fmt.Fprintf(os.Stdout, "CLASS %3d: Rewrites = %6d\n", class, score)
		}
	}
}
