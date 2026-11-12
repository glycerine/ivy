package if_

import "github.com/glycerine/abcgo/aig/gia"

// Gia_ManBuildFromMini balances a mapped cut and expands the resulting compact
// AIG into the destination GIA. leaves are destination literals in cut order.
func Gia_ManBuildFromMini(dst *Gia_Man_t, mapping *If_Man_t, cut *If_Cut_t, leaves, mini *Vec_Int_t, hash, useDSD int32) int32 {
	if useDSD != 0 {
		if If_CutDsdBalanceEval(mapping, cut, mini) < 0 {
			return -1
		}
	} else if If_CutSopBalanceEval(mapping, cut, mini) < 0 {
		return -1
	}
	return gia.Gia_ManBuildFromMiniInt(dst, leaves, mini, hash)
}
