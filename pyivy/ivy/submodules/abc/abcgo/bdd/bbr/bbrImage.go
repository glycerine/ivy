// The BBR and extra image schedulers share the original support-driven
// quantification algorithm. BBR adds manager-node guards at scheduled merges.
package bbr

import "github.com/glycerine/abcgo/bdd/extrab"

type Bbr_ImageTree_t = extrab.Extra_ImageTree_t
type Bbr_ImageTree2_t = extrab.Extra_ImageTree2_t

func Bbr_bddImageStart(dd *DdManager, care *DdNode, nParts int, parts []*DdNode, nVars int, vars []*DdNode, nBddMax int, verbose bool) *Bbr_ImageTree_t {
	if nBddMax < 0 {
		return nil
	}
	return extrab.Extra_bddImageStartLimited(dd, care, nParts, parts, nVars, vars, nBddMax, verbose)
}
func Bbr_bddImageCompute(tree *Bbr_ImageTree_t, care *DdNode) *DdNode {
	return extrab.Extra_bddImageCompute(tree, care)
}
func Bbr_bddImageRead(tree *Bbr_ImageTree_t) *DdNode { return extrab.Extra_bddImageRead(tree) }
func Bbr_bddImageTreeDelete(tree *Bbr_ImageTree_t)   { extrab.Extra_bddImageTreeDelete(tree) }
func Bbr_bddImageStart2(dd *DdManager, care *DdNode, nParts int, parts []*DdNode, nVars int, vars []*DdNode, verbose bool) *Bbr_ImageTree2_t {
	return extrab.Extra_bddImageStart2(dd, care, nParts, parts, nVars, vars, verbose)
}
func Bbr_bddImageCompute2(tree *Bbr_ImageTree2_t, care *DdNode) *DdNode {
	return extrab.Extra_bddImageCompute2(tree, care)
}
func Bbr_bddImageRead2(tree *Bbr_ImageTree2_t) *DdNode { return extrab.Extra_bddImageRead2(tree) }
func Bbr_bddImageTreeDelete2(tree *Bbr_ImageTree2_t)   { extrab.Extra_bddImageTreeDelete2(tree) }
func Bbr_bddComputeCube(dd *DdManager, vars []*DdNode, n int) *DdNode {
	return extrab.Extra_bddComputeCube(dd, vars, n)
}
func Bbr_bddPrint(dd *DdManager, f *DdNode) { extrab.Extra_bddPrint(dd, f) }
