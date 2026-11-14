package core

import (
	"fmt"
	"io"
	"math"
)

// Cudd_zddDagSize counts nonterminal nodes reachable from a ZDD edge.
func Cudd_zddDagSize(root *DdNode) int32 {
	seen := map[*DdNode]bool{}
	var visit func(*DdNode)
	visit = func(n *DdNode) {
		if n == nil || Cudd_IsConstant(n) {
			return
		}
		n = Cudd_Regular(n)
		if seen[n] {
			return
		}
		seen[n] = true
		visit(n.T)
		visit(n.E)
	}
	visit(root)
	return int32(len(seen))
}

// Cudd_zddCountMinterm counts members after accounting for path variables
// omitted from the represented support, as in the original ZDD utility.
func Cudd_zddCountMinterm(dd *DdManager, root *DdNode, path int32) float64 {
	if dd == nil || dd.closed {
		return CUDD_OUT_OF_MEM
	}
	return Cudd_zddCountDouble(dd, root) / math.Pow(2, float64(dd.ZddSize-int(path)))
}

// Cudd_zddPrintSubtable prints the current ZDD unique table by level. The
// pure-Go manager has no hash-chain slot layout, so nodes are ordered by ID.
func Cudd_zddPrintSubtable(dd *DdManager) {
	if dd == nil || dd.closed {
		return
	}
	for level := dd.ZddSize - 1; level >= 0; level-- {
		index := dd.InvPermZ[level]
		fmt.Fprintf(dd.stdout, "subtable[%d]:\n", level)
		var nodes []*DdNode
		for _, node := range dd.zddUnique {
			if int(node.Index) == index {
				nodes = append(nodes, node)
			}
		}
		for i := 1; i < len(nodes); i++ {
			for j := i; j > 0 && nodes[j].ID < nodes[j-1].ID; j-- {
				nodes[j], nodes[j-1] = nodes[j-1], nodes[j]
			}
		}
		for _, node := range nodes {
			fmt.Fprintf(dd.stdout, "ID = %d\tindex = %d\tr = %d\t", node.ID, node.Index, node.Ref)
			printZddChild(dd.stdout, "T", node.T, dd.One)
			printZddChild(dd.stdout, "E", node.E, dd.One)
		}
	}
	fmt.Fprintln(dd.stdout)
}

func printZddChild(w io.Writer, label string, child, base *DdNode) {
	if Cudd_IsConstant(child) {
		value := 0
		if child == base {
			value = 1
		}
		fmt.Fprintf(w, "%s = %d\t", label, value)
		return
	}
	fmt.Fprintf(w, "%s = %d\t", label, child.ID)
}
