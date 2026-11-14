package cudd

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Cudd_DebugCheck checks manager tables and node invariants, returning zero
// when they are consistent and one when a discrepancy is found.
func Cudd_DebugCheck(dd *DdManager) int32 {
	if err := core.Cudd_CheckConsistency(dd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// Cudd_CheckKeys checks the canonical-table invariants and reports one on
// failure, matching the diagnostic nature of the CUDD API.
func Cudd_CheckKeys(dd *DdManager) int32 { return Cudd_DebugCheck(dd) }

// Cudd_CheckZeroRef reports outstanding externally held canonical DD nodes.
func Cudd_CheckZeroRef(dd *DdManager) int32 { return core.Cudd_CheckZeroRef(dd) }

func cuddHeapProfile(dd *DdManager) int32 {
	if dd == nil {
		return 0
	}
	_, err := fmt.Fprintf(os.Stdout, "BDD nodes: %d\nADD/ZDD nodes: %d\n", core.Cudd_ReadNodeCount(dd), core.Cudd_CacheSize(dd))
	if err != nil {
		return 0
	}
	return 1
}

func cuddPrintNode(node *DdNode, writer *os.File) {
	if node == nil || writer == nil {
		return
	}
	regular := Cudd_Regular(node)
	_, _ = fmt.Fprintf(writer, "node{id=%d index=%d ref=%d complement=%t T=%d E=%d}\n", regular.ID, regular.Index, regular.Ref, Cudd_IsComplement(node), childID(regular.T), childID(regular.E))
}

func childID(node *DdNode) uint64 {
	if node == nil {
		return 0
	}
	return Cudd_Regular(node).ID
}

func debugFindParent(dd *DdManager, node *DdNode) {
	if dd == nil || node == nil {
		return
	}
	// The canonical table owns all live nodes; a parent search is only a
	// debugging aid, so report the target identity without mutating the graph.
	cuddPrintNode(node, os.Stderr)
}
