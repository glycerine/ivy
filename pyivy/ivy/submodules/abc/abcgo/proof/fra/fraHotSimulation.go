package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/framgr"
)

// One-hot simulation uses the canonical shared FRA manager.
func Fra_RegToLit(n, complement int32) int32 { return framgr.Fra_RegToLit(n, complement) }
func Fra_LitReg(n int32) int32               { return framgr.Fra_LitReg(n) }
func Fra_LitSign(n int32) int32              { return framgr.Fra_LitSign(n) }
func Fra_OneHotNodeIsConst(s *Fra_Sml_t, node aig.Edge) int32 {
	return framgr.Fra_OneHotNodeIsConst(s, node)
}
func Fra_OneHotNodesAreEqual(s *Fra_Sml_t, a, b aig.Edge) int32 {
	return framgr.Fra_OneHotNodesAreEqual(s, a, b)
}
func Fra_OneHotNodesAreClause(s *Fra_Sml_t, a, b aig.Edge, negA, negB int32) int32 {
	return framgr.Fra_OneHotNodesAreClause(s, a, b, negA, negB)
}
func Fra_OneHotRefineUsingCex(p *Fra_Man_t, clauses *vec.Vec_Int_t) int32 {
	return framgr.Fra_OneHotRefineUsingCex(p, clauses)
}
func Fra_OneHotCount(p *Fra_Man_t, clauses *vec.Vec_Int_t) int32 {
	return framgr.Fra_OneHotCount(p, clauses)
}
func Fra_OneHotEstimateCoverage(p *Fra_Man_t, clauses *vec.Vec_Int_t) {
	framgr.Fra_OneHotEstimateCoverage(p, clauses)
}
