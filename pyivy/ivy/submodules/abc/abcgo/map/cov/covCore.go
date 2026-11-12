package cov

import "fmt"

// Abc_NtkSopEsopCover converts an AIG network into an equivalent bounded-cover
// logic network. It grows cut boundaries iteratively when a local support or
// cover exceeds the configured limits, as covCore.c does.
func Abc_NtkSopEsopCover(pNtk *Abc_Ntk_t, nFaninMax, nCubesMax, fUseEsop, fUseSop, fUseInvs, fVerbose int32) *Abc_Ntk_t {
	if pNtk == nil || nFaninMax < 0 || nCubesMax < 0 || fUseEsop == 0 && fUseSop == 0 {
		return nil
	}
	manager := Cov_ManAlloc(pNtk, nFaninMax, nCubesMax)
	manager.fUseEsop = fUseEsop
	manager.fUseSop = fUseSop
	result, err := deriveAIGToSOP(pNtk, nFaninMax, nCubesMax, fUseEsop, fUseSop)
	Cov_ManFree(manager)
	if err != nil {
		if fVerbose != 0 {
			fmt.Printf("Abc_NtkSopEsopCover: %v\n", err)
		}
		return nil
	}
	_ = fUseInvs // C's regular derivation path also leaves inversion cleanup disabled.
	if err := result.Check(); err != nil {
		if fVerbose != 0 {
			fmt.Printf("Abc_NtkSopEsopCover: network check failed: %v\n", err)
		}
		return nil
	}
	return result
}
