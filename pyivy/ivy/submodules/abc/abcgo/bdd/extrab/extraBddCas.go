/**CFile****************************************************************

  FileName    [extraBddCas.c]

  PackageName [extra]

  Synopsis    [Procedures related to LUT cascade synthesis.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - September 1, 2003.]

  Revision    [$Id: extraBddCas.c,v 1.0 2003/05/21 18:03:50 alanmi Exp $]

***********************************************************************/

package extrab

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab/cascade"
)

func Extra_bddEncodingBinary(dd *DdManager, functions []*DdNode, nFunctions int, vars []*DdNode, nVars int) *DdNode {
	if dd == nil || nFunctions < 0 || nFunctions > len(functions) || nVars < 0 || nVars > len(vars) {
		return nil
	}
	return cascade.Extra_bddEncodingBinary(dd, functions[:nFunctions], vars[:nVars])
}
func Extra_bddEncodingNonStrict(dd *DdManager, columns []*DdNode, nColumns int, varsCol *DdNode, codes []*DdNode, nCodes int, simple *int) *DdNode {
	if simple == nil {
		return nil
	}
	*simple = 0
	if dd == nil || nColumns < 0 || nColumns > len(columns) || nCodes < 0 || nCodes > len(codes) {
		return nil
	}
	result, nSimple := cascade.Extra_bddEncodingNonStrict(dd, columns[:nColumns], varsCol, codes[:nCodes])
	*simple = nSimple
	return result
}

// Extra_bddNodePathsUnderCut maps each Boolean cofactor to its path sum. Both
// keys and values are referenced, matching ownership of the original st table.
func Extra_bddNodePathsUnderCut(dd *DdManager, f *DdNode, cut int) map[*DdNode]*DdNode {
	if !autoBDD(dd, f) || cut < 0 || cut > len(dd.Vars) {
		return nil
	}
	paths := cascade.Extra_bddNodePathsUnderCutArray(dd, []*DdNode{f}, []*DdNode{dd.One}, cut)
	result := make(map[*DdNode]*DdNode, len(paths.Nodes))
	for i, n := range paths.Nodes {
		result[n] = paths.Columns[i]
		cudd.Cudd_Ref(n)
		cudd.Cudd_Ref(paths.Columns[i])
	}
	return result
}

// Extra_bddNodePathsUnderCutArray fills caller buffers with referenced cofactors
// and paths. An invalid domain/buffer returns -1 before writing either buffer.
func Extra_bddNodePathsUnderCutArray(dd *DdManager, nodes, cubes []*DdNode, n int, outNodes, outPaths []*DdNode, cut int) int {
	if dd == nil || n < 1 || n > len(nodes) || n > len(cubes) || cut < 0 || cut > len(dd.Vars) {
		return -1
	}
	for i := 0; i < n; i++ {
		if !autoBDD(dd, cubes[i]) || !autoBDD(dd, nodes[i]) && cudd.Cudd_addFindMin(dd, nodes[i]) == nil {
			return -1
		}
	}
	paths := cascade.Extra_bddNodePathsUnderCutArray(dd, nodes[:n], cubes[:n], cut)
	if len(outNodes) < len(paths.Nodes) || len(outPaths) < len(paths.Nodes) {
		return -1
	}
	for i, f := range paths.Nodes {
		outNodes[i] = f
		outPaths[i] = paths.Columns[i]
		cudd.Cudd_Ref(f)
		cudd.Cudd_Ref(paths.Columns[i])
	}
	return len(paths.Nodes)
}
func Extra_ProfileWidth(dd *DdManager, f *DdNode, profile []int, cut int) int {
	if dd == nil || len(profile) < max(len(dd.Vars), dd.ZddSize)+1 || cut < -1 || cut > len(profile) {
		return -1
	}
	values := cascade.Extra_ProfileWidth(dd, f)
	if values == nil {
		return -1
	}
	clear(profile)
	copy(profile, values)
	limit := len(profile)
	if cut > 0 {
		limit = cut
	}
	width := 0
	for _, value := range profile[:limit] {
		width = max(width, value)
	}
	return width
}

// Extra_CollectNodes replaces the original pointer-key st table with a Go set.
func Extra_CollectNodes(f *DdNode) map[*DdNode]struct{} {
	if f == nil {
		return nil
	}
	result := make(map[*DdNode]struct{})
	extraCollectNodes(f, result)
	return result
}
func extraCollectNodes(f *DdNode, nodes map[*DdNode]struct{}) {
	r := cudd.Cudd_Regular(f)
	if _, seen := nodes[r]; seen {
		return
	}
	nodes[r] = struct{}{}
	if !cudd.Cudd_IsConstant(r) {
		extraCollectNodes(r.E, nodes)
		extraCollectNodes(r.T, nodes)
	}
}
