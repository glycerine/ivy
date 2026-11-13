package fra

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/framgr"
)

// Implication simulation uses the canonical shared FRA manager.
func Sml_NodeCheckImp(s *Fra_Sml_t, left, right int32) int32 {
	return framgr.Sml_NodeCheckImp(s, left, right)
}
func Fra_ImpRefineUsingCex(p *Fra_Man_t, imps *vec.Vec_Int_t) int32 {
	return framgr.Fra_ImpRefineUsingCex(p, imps)
}
func Fra_ImpCompactArray(imps *vec.Vec_Int_t) { framgr.Fra_ImpCompactArray(imps) }
func Fra_ImpComputeStateSpaceRatio(p *Fra_Man_t) float64 {
	return framgr.Fra_ImpComputeStateSpaceRatio(p)
}
