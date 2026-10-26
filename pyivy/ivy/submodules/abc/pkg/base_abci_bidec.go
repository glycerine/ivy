package abc

// Native HOP/BDC resynthesis for src/base/abci/abcBidec.c.

import (
	"math"
	"slices"
)

type ABCBidecOptions struct {
	MaxRoots, MaxTruthWords, MaxBDCNodes, MaxNewNodes int
	Stop                                              func() bool
}

type ABCBidecNode struct {
	Root        HopEdge
	Variables   int
	Care        *TruthTable
	Probability float64
}

type ABCBidecNodeResult struct {
	Root                          HopEdge
	HopNodesBefore, HopNodesAfter int
	Gain, BDCGates                int
	CompletedDontCares            bool
}

type ABCBidecResult struct {
	Roots                       []HopEdge
	Nodes                       []ABCBidecNodeResult
	Gain, BDCGates, SkippedWide int
}

func normalizeABCBidecOptions(options ABCBidecOptions) (ABCBidecOptions, error) {
	if options.MaxRoots < 0 || options.MaxTruthWords < 0 || options.MaxBDCNodes < 0 || options.MaxNewNodes < 0 {
		return options, ErrDimension
	}
	if options.MaxRoots == 0 {
		options.MaxRoots = 1 << 20
	}
	if options.MaxTruthWords == 0 {
		options.MaxTruthWords = 1 << 16
	}
	if options.MaxBDCNodes == 0 {
		options.MaxBDCNodes = 2048
	}
	if options.MaxNewNodes == 0 {
		options.MaxNewNodes = 1000
	}
	return options, nil
}

func abcBidecPreparedTruth(manager *HopManager, node ABCBidecNode, options ABCBidecOptions) (*TruthTable, *TruthTable, bool, error) {
	if manager == nil || node.Variables < 0 || node.Variables > 15 || math.IsNaN(node.Probability) || math.IsInf(node.Probability, 0) || node.Probability > 1 {
		return nil, nil, false, ErrDimension
	}
	if err := manager.check(); err != nil {
		return nil, nil, false, err
	}
	if !manager.valid(node.Root) || node.Variables > len(manager.pis) {
		return nil, nil, false, ErrDimension
	}
	words, err := manager.Truth(node.Root, node.Variables, false)
	if err != nil {
		return nil, nil, false, err
	}
	function, err := TruthTableFromWords(node.Variables, words, options.MaxTruthWords)
	if err != nil {
		return nil, nil, false, err
	}
	var care *TruthTable
	if node.Care != nil {
		if err = node.Care.Check(); err != nil || node.Care.Variables() != node.Variables {
			return nil, nil, false, ErrDimension
		}
		care = node.Care.Clone()
	}
	if node.Probability < 0 || 2*node.Probability*(1-node.Probability) < 0.4 {
		return function, care, false, nil
	}
	if care != nil {
		if node.Probability > 0.5 {
			function, err = function.Or(care.Not())
		} else {
			function, err = function.And(care, false, false)
		}
		if err != nil {
			return nil, nil, false, err
		}
	}
	return function, nil, true, nil
}

func abcBidecCopyToHop(manager *HopManager, decomposition *BdcManager) (HopEdge, error) {
	if manager == nil || decomposition == nil {
		return 0, ErrDimension
	}
	copies := make([]HopEdge, decomposition.NodeCount())
	if len(copies) == 0 {
		return 0, ErrInvariant
	}
	copies[0] = manager.Const1()
	for variable := 0; variable < decomposition.Variables(); variable++ {
		if variable >= len(manager.pis) {
			return 0, ErrDimension
		}
		copies[variable+1] = manager.pis[variable]
	}
	copyLiteral := func(literal BdcLit) (HopEdge, error) {
		id := int(literal.ID())
		if !literal.Valid() || id < 0 || id >= len(copies) {
			return 0, ErrInvariant
		}
		return copies[id].NotIf(literal.Complemented()), nil
	}
	for id := decomposition.Variables() + 1; id < decomposition.NodeCount(); id++ {
		node, err := decomposition.Node(id)
		if err != nil || node.Type != BdcNodeAND {
			return 0, ErrInvariant
		}
		first, err := copyLiteral(node.Fanin0)
		if err != nil {
			return 0, err
		}
		second, err := copyLiteral(node.Fanin1)
		if err != nil {
			return 0, err
		}
		if copies[id], err = manager.And(first, second); err != nil {
			return 0, err
		}
	}
	root, err := decomposition.Root()
	if err != nil {
		return 0, err
	}
	return copyLiteral(root)
}

// ABCBidecResynthesizeNode converts one local HOP function to truth, performs
// bi-decomposition, and translates the result back into the same HOP manager.
// The source root and care table remain owned by the caller. Failed calls are
// atomic with respect to newly constructed HOP nodes.
func ABCBidecResynthesizeNode(manager *HopManager, node ABCBidecNode, options ABCBidecOptions) (ABCBidecNodeResult, error) {
	options, err := normalizeABCBidecOptions(options)
	if err != nil {
		return ABCBidecNodeResult{}, err
	}
	if options.Stop != nil && options.Stop() {
		return ABCBidecNodeResult{}, ErrCancelled
	}
	function, care, completed, err := abcBidecPreparedTruth(manager, node, options)
	if err != nil {
		return ABCBidecNodeResult{}, err
	}
	before, err := manager.DagSize(node.Root)
	if err != nil {
		return ABCBidecNodeResult{}, err
	}
	decomposition, err := NewBdcManager(BdcParameters{MaxVariables: max(2, node.Variables), MaxWords: options.MaxTruthWords, MaxNodes: options.MaxBDCNodes})
	if err != nil {
		return ABCBidecNodeResult{}, err
	}
	if _, err = decomposition.Decompose(function, care, nil, BdcDecomposeOptions{MaxNewNodes: options.MaxNewNodes, Stop: options.Stop}); err != nil {
		return ABCBidecNodeResult{}, err
	}
	start := len(manager.objects)
	root, err := abcBidecCopyToHop(manager, decomposition)
	if err != nil {
		manager.rollback(start)
		return ABCBidecNodeResult{}, err
	}
	after, err := manager.DagSize(root)
	if err != nil {
		manager.rollback(start)
		return ABCBidecNodeResult{}, err
	}
	return ABCBidecNodeResult{Root: root, HopNodesBefore: before, HopNodesAfter: after, Gain: before - after, BDCGates: decomposition.GateCount(), CompletedDontCares: completed}, nil
}

// ABCBidecResynthesize applies the network-level loop in Abc_NtkBidecResyn to
// an owned sequence of local HOP roots. Nodes wider than 15 inputs are retained
// unchanged, matching the source. The whole batch is atomic on error.
func ABCBidecResynthesize(manager *HopManager, nodes []ABCBidecNode, options ABCBidecOptions) (ABCBidecResult, error) {
	options, err := normalizeABCBidecOptions(options)
	if err != nil || manager == nil {
		if err == nil {
			err = ErrDimension
		}
		return ABCBidecResult{}, err
	}
	if len(nodes) > options.MaxRoots {
		return ABCBidecResult{}, ErrLimit
	}
	if err = manager.check(); err != nil {
		return ABCBidecResult{}, err
	}
	start := len(manager.objects)
	result := ABCBidecResult{Roots: make([]HopEdge, len(nodes)), Nodes: make([]ABCBidecNodeResult, len(nodes))}
	for index, node := range nodes {
		if options.Stop != nil && options.Stop() {
			manager.rollback(start)
			return ABCBidecResult{}, ErrCancelled
		}
		if node.Variables > 15 {
			if node.Variables > len(manager.pis) || !manager.valid(node.Root) {
				manager.rollback(start)
				return ABCBidecResult{}, ErrDimension
			}
			result.Roots[index] = node.Root
			result.Nodes[index] = ABCBidecNodeResult{Root: node.Root}
			result.SkippedWide++
			continue
		}
		one, oneErr := ABCBidecResynthesizeNode(manager, node, options)
		if oneErr != nil {
			manager.rollback(start)
			return ABCBidecResult{}, oneErr
		}
		result.Roots[index], result.Nodes[index] = one.Root, one
		result.Gain += one.Gain
		result.BDCGates += one.BDCGates
	}
	result.Roots = slices.Clone(result.Roots)
	return result, nil
}
