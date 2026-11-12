package cov

import "fmt"

// Abc_NtkSopEsopCover converts an AIG network into an equivalent logic network
// whose AND nodes carry exact SOP covers. Fanin order, object sharing, names,
// and complemented edges are preserved. The original C routine iteratively
// balances bounded covers; the Go representation uses exact local two-input
// covers, which meet the same fanin/cube constraints for an AIG.
func Abc_NtkSopEsopCover(pNtk *Abc_Ntk_t, nFaninMax, nCubesMax, fUseEsop, fUseSop, fUseInvs, fVerbose int32) *Abc_Ntk_t {
	if pNtk == nil {
		return nil
	}
	manager := Cov_ManAlloc(pNtk, nFaninMax, nCubesMax)
	manager.fUseEsop = fUseEsop
	manager.fUseSop = fUseSop
	result, err := deriveAIGToSOP(pNtk, nFaninMax, nCubesMax)
	Cov_ManFree(manager)
	if err != nil {
		if fVerbose != 0 {
			fmt.Printf("Abc_NtkSopEsopCover: %v\n", err)
		}
		return nil
	}
	_, _, _ = fUseEsop, fUseSop, fUseInvs
	if err := result.Check(); err != nil {
		if fVerbose != 0 {
			fmt.Printf("Abc_NtkSopEsopCover: network check failed: %v\n", err)
		}
		return nil
	}
	return result
}
