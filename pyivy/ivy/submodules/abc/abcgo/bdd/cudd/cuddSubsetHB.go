package cudd

import (
	"math"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

const (
	heavyBranchDefaultPageSize  = 2048
	heavyBranchNodeDataPageSize = 1024
	heavyBranchInitialPages     = 128
)

// NodeData stores the metrics used by the heavy branch heuristic.
type NodeData struct {
	mintermPointer         *float64
	nodesPointer           *int32
	lightChildNodesPointer *int32
}

type NodeData_t = NodeData

var memOut int32
var hbZero, hbOne *DdNode
var max float64
var pageSize int32 = heavyBranchDefaultPageSize
var nodeDataPageSize int32 = heavyBranchNodeDataPageSize

// Cudd_SubsetHeavyBranch extracts a dense subset by repeatedly retaining the
// child that contributes more minterms until the requested node bound is met.
func Cudd_SubsetHeavyBranch(dd *DdManager, f *DdNode, numVars, threshold int32) *DdNode {
	memOut = 0
	return cuddSubsetHeavyBranch(dd, f, numVars, threshold)
}

// Cudd_SupersetHeavyBranch computes a superset by complementing the input,
// taking a heavy branch subset, and complementing the result.
func Cudd_SupersetHeavyBranch(dd *DdManager, f *DdNode, numVars, threshold int32) *DdNode {
	memOut = 0
	g := core.Cudd_Not(f)
	var subset *DdNode
	subset = cuddSubsetHeavyBranch(dd, g, numVars, threshold)
	return core.Cudd_NotCond(subset, subset != nil)
}

// cuddSubsetHeavyBranch builds the metrics for each node and performs the
// recursive heavy branch selection. Metrics are keyed by edge so complemented
// edges retain their distinct minterm counts.
func cuddSubsetHeavyBranch(dd *DdManager, f *DdNode, numVars, threshold int32) *DdNode {
	if f == nil {
		return nil
	}
	hbOne = Cudd_ReadOne(dd)
	hbZero = core.Cudd_Not(hbOne)
	if core.Cudd_IsConstant(f) {
		return f
	}
	if numVars == 0 {
		numVars = 1023
	}
	max = math.Pow(2, float64(numVars))
	minterms := make(map[*DdNode]float64)
	countMinterm(f, max, minterms)
	nodes := make(map[*DdNode]int32)
	count := countNodes(f, nodes)
	if count < 0 {
		return nil
	}
	return buildHeavySubset(dd, f, &count, threshold, minterms, nodes, make(map[*DdNode]*DdNode), make(map[*DdNode]bool))
}

func countMinterm(node *DdNode, total float64, cache map[*DdNode]float64) float64 {
	if core.Cudd_IsConstant(node) {
		if node == hbZero {
			return 0
		}
		return total
	}
	if v, ok := cache[node]; ok {
		return v
	}
	t := countMinterm(heavyThen(node), total, cache) / 2
	e := countMinterm(heavyElse(node), total, cache) / 2
	cache[node] = t + e
	return t + e
}

func countNodes(node *DdNode, cache map[*DdNode]int32) int32 {
	if node == nil || core.Cudd_IsConstant(node) {
		return 0
	}
	if v, ok := cache[node]; ok {
		return v
	}
	// Install a provisional value before recursion; a BDD is acyclic.
	cache[node] = 0
	t, e := countNodes(heavyThen(node), cache), countNodes(heavyElse(node), cache)
	v := t + e + 1
	cache[node] = v
	return v
}

func buildHeavySubset(dd *DdManager, node *DdNode, size *int32, threshold int32, mins map[*DdNode]float64, nodes map[*DdNode]int32, memo map[*DdNode]*DdNode, stored map[*DdNode]bool) *DdNode {
	if *size <= threshold || core.Cudd_IsConstant(node) {
		storeHeavyNodes(node, stored)
		return node
	}
	if result, ok := memo[node]; ok {
		return result
	}
	t, e := heavyThen(node), heavyElse(node)
	tm, em := branchMinterm(t, mins), branchMinterm(e, mins)
	light := e
	heavy := t
	if em > tm {
		light, heavy = t, e
	}
	*size -= nodes[light]
	kept := buildHeavySubset(dd, heavy, size, threshold, mins, nodes, memo, stored)
	if kept == nil {
		return nil
	}
	lightApprox := hbZero
	if stored[core.Cudd_Regular(light)] {
		lightApprox = light
	} else if approx, ok := memo[light]; ok {
		lightApprox = approx
	}
	var thenBranch, elseBranch *DdNode
	if heavy == t {
		thenBranch, elseBranch = kept, lightApprox
	} else {
		thenBranch, elseBranch = lightApprox, kept
	}
	result := core.Cudd_bddIte(dd, Cudd_ReadVars(dd, int32(Cudd_NodeReadIndex(core.Cudd_Regular(node)))), thenBranch, elseBranch)
	if result != nil {
		memo[node] = result
	}
	return result
}

func branchMinterm(node *DdNode, mins map[*DdNode]float64) float64 {
	if core.Cudd_IsConstant(node) {
		if node == hbZero {
			return 0
		}
		return max
	}
	return mins[node]
}

func storeHeavyNodes(node *DdNode, seen map[*DdNode]bool) {
	if node == nil || core.Cudd_IsConstant(node) {
		return
	}
	node = core.Cudd_Regular(node)
	if seen[node] {
		return
	}
	seen[node] = true
	storeHeavyNodes(heavyThen(node), seen)
	storeHeavyNodes(heavyElse(node), seen)
}

func heavyThen(node *DdNode) *DdNode {
	return core.Cudd_NotCond(core.Cudd_T(node), core.Cudd_IsComplement(node))
}

func heavyElse(node *DdNode) *DdNode {
	return core.Cudd_NotCond(core.Cudd_E(node), core.Cudd_IsComplement(node))
}
