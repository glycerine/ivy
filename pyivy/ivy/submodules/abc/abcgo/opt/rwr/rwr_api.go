package rwr

import (
	"fmt"
	"os"
	"sort"

	"github.com/glycerine/abcgo/aig/ivy"

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
	if p.vClasses == nil {
		// The Go API also accepts a zero-value or precompute-only manager.
		normal := Rwr_ManStart(0)
		p.vClasses, p.pMapInv = normal.vClasses, normal.pMapInv
		p.pMap, p.pPerms, p.pPhases, p.pPerms4 = normal.pMap, normal.pPerms, normal.pPhases, normal.pPerms4
	}
	choice := p.rwrRewriteCuts(net, root, cuts, required, fUseZeros != 0)
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
	for c := rewriteNodeCuts(pManCut, pNode); c != nil; c = c.Next {
		if cut.Cut_CutReadLeaveNum(c) != 4 {
			continue
		}
		leaves := cut.Cut_CutReadLeaves(c)
		truth := cut.Cut_CutReadTruth(c)
		if len(truth) == 0 {
			continue
		}
		cuts = append(cuts, Rwr_Cut{Truth: uint16(truth[0]), Leaves: leaves})
	}
	required := int32(1<<30 - 1)
	if fUpdateLevel != 0 {
		required = abc.Abc_ObjRequiredLevel(pNode)
	}
	return Rwr_RewriteCuts(p, network, pNode.Signal(), cuts, required, fUseZeros, fPlaceEnable)
}

// Read live fanins because earlier replacements in the pass can rewire
// existing nodes or insert new ones beyond the original object count.
func rewriteNodeCuts(p *cut.Cut_Man_t, node *abc.Object) *cut.Cut_Cut_t {
	id := int32(node.ID)
	if cuts := cut.Cut_NodeReadCutsNew(p, id); cuts != nil {
		return cuts
	}
	a, b := abc.Abc_ObjChild0(node), abc.Abc_ObjChild1(node)
	rewriteNodeCuts(p, a.Object)
	rewriteNodeCuts(p, b.Object)
	if abc.Abc_ObjFanoutNum(node) > 1 && abc.Abc_NodeIsMuxControlType(node) == 0 {
		cut.Cut_ManIncrementDagNodes(p)
	}
	return cut.Cut_NodeComputeCuts(p, id, int32(a.Object.ID), int32(b.Object.ID), abc.Abc_ObjFaninC0(node), abc.Abc_ObjFaninC1(node), 1, 0)
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
	for _, class := range p.vClasses {
		for _, node := range class {
			node.nScore, node.nGain, node.nAdded = 0, 0, 0
		}
	}
}

// Rwr_ScoresReport orders classes by total gain and prints the scored graphs.
func Rwr_ScoresReport(p *Rwr_Man_t) {
	type classGain struct{ class, gain int }
	var ranked []classGain
	for class, nodes := range p.vClasses {
		var gain int
		for _, node := range nodes {
			gain += int(node.nGain)
		}
		ranked = append(ranked, classGain{class, gain})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].gain > ranked[j].gain })
	for _, entry := range ranked {
		if entry.gain == 0 {
			break
		}
		class := p.vClasses[entry.class]
		truth := uint32(p.pMapInv[entry.class])
		fmt.Fprintf(os.Stdout, "CLASS %3d: Subgr = %3d. Total gain = %6d.  %016b  ", entry.class, len(class), entry.gain, truth)
		ivy.Ivy_TruthDsdComputePrint(truth | truth<<16)
		for i, node := range class {
			if node.nScore != 0 {
				fmt.Fprintf(os.Stdout, "    %2d: S=%5d. A=%5d. G=%6d. ", i, node.nScore, node.nAdded, node.nGain)
				dec.Dec_GraphPrint(os.Stdout, node.graph, nil, "")
			}
		}
	}
}
