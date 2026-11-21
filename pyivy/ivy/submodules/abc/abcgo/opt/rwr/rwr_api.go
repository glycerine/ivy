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

// Rwr_RewriteCuts runs the source algorithm over the caller's cuts in order.
// The network adapter keeps this package independent of ABC's concrete graph
// types. fPlaceEnable is accepted for source API parity; placement scoring is
// disabled in the reference implementation's active evaluation path.
func Rwr_RewriteCuts(p *Rwr_Man_t, net Rwr_Network, root any, cuts []Rwr_Cut, required int32, fUseZeros, fPlaceEnable int32) int32 {
	return p.rewriteCuts(net, root, rewriteCutIterator{ordered: cuts}, required, fUseZeros != 0, new([4]any))
}

// Rwr_NodeRewrite preserves the source entry point while adapting ABC objects
// and cut lists to the network independent ordered rewrite implementation.
func Rwr_NodeRewrite(p *Rwr_Man_t, pManCut *cut.Cut_Man_t, pNode *abc.Object, fUpdateLevel, fUseZeros, fPlaceEnable int32) int32 {
	if pManCut == nil || pNode == nil {
		return -1
	}
	cuts := rewriteNodeCuts(pManCut, pNode)
	required := int32(1<<30 - 1)
	if fUpdateLevel != 0 {
		required = abc.Abc_ObjRequiredLevel(pNode)
	}
	// Concrete counting only calls ABC routines. Its boundary buffer belongs to
	// this manager, as in native Rwr_NodeRewrite. Public adapters get a fresh
	// buffer because a caller can retain it or reenter through a callback.
	if p == nil {
		panic("rwr: nil rewrite manager")
	}
	if p.concreteNet == nil {
		p.concreteNet = new(cachedRewriteNetwork)
	}
	network := p.concreteNet
	network.beginNode(pNode.Network())
	defer func() {
		network.Network = nil
		clear(p.concreteFanins[:])
	}()
	return p.rewriteCuts(network, pNode.Signal().DecValue(), rewriteCutIterator{list: cuts}, required, fUseZeros != 0, &p.concreteFanins)
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
	abc.DecNetwork
}

func (n abcRewriteNetwork) Lookup(id int32) (any, bool) {
	o := n.Object(int(id))
	if o == nil {
		return nil, false
	}
	return o.Signal().DecValue(), true
}
func (n abcRewriteNetwork) FanoutCount(value any) int {
	return int(abc.Abc_ObjFanoutNum(value.(abc.Edge).Object))
}
func (n abcRewriteNetwork) MarkBoundary(values []any, delta int) {
	for _, value := range values {
		abc.Abc_ObjAdjustFanoutCount(value.(abc.Edge).Object, int32(delta))
	}
}
func (n abcRewriteNetwork) LabelMFFC(root any) int32 {
	abc.Abc_NtkIncrementTravId(n.Network)
	return abc.Abc_NodeMffcLabelAig(root.(abc.Edge).Object)
}
func (n abcRewriteNetwork) CountGraph(root any, graph *dec.Dec_Graph_t, saved, levelMax int32) int32 {
	return dec.Dec_GraphToNetworkCount(n.DecNetwork, root, graph, saved, levelMax)
}

// Candidate graphs for the same root share many two-input ANDs. The topology
// and levels are stable throughout one concrete rewrite. MFFC membership is
// relabeled for each cut and is always queried on the underlying network.
type cachedRewriteNetwork struct {
	abcRewriteNetwork
	generation uint32
	entries    [256]rewriteAndEntry
}

// Store object IDs so that keeping a rewrite manager cannot pin old networks.
type rewriteAndEntry struct {
	left, right     uint
	objectID, level int
	generation      uint32
	complement      bool
}

func (n *cachedRewriteNetwork) beginNode(network *abc.Network) {
	n.Network = network
	n.generation++
	if n.generation == 0 {
		clear(n.entries[:])
		n.generation = 1
	}
}

func rewriteLiteral(edge abc.Edge, invert bool) uint {
	literal := uint(edge.Object.ID) << 1
	if edge.Complement != invert {
		literal++
	}
	return literal
}

func (n *cachedRewriteNetwork) LookupAndLevel(a, b any, invertA, invertB bool) (any, int) {
	left := rewriteLiteral(a.(abc.Edge), invertA)
	right := rewriteLiteral(b.(abc.Edge), invertB)
	if left > right {
		left, right = right, left
	}
	hash := left*12582917 ^ right*4256249 ^ left>>4 ^ right>>5
	entry := &n.entries[hash&uint(len(n.entries)-1)]
	if entry.generation == n.generation && entry.left == left && entry.right == right {
		if entry.objectID < 0 {
			return nil, -1
		}
		edge := abc.Edge{Object: n.Object(entry.objectID), Complement: entry.complement}
		return edge.DecValue(), entry.level
	}
	signal, level := n.DecNetwork.LookupAndLevel(a, b, invertA, invertB)
	*entry = rewriteAndEntry{left: left, right: right, objectID: -1, level: level, generation: n.generation}
	if signal != nil {
		edge := signal.(abc.Edge)
		entry.objectID, entry.complement = edge.Object.ID, edge.Complement
	}
	return signal, level
}

func (n *cachedRewriteNetwork) CountGraph(root any, graph *dec.Dec_Graph_t, saved, levelMax int32) int32 {
	return dec.Dec_GraphToNetworkCount(n, root, graph, saved, levelMax)
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
