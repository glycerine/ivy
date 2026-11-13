package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec_ManVerifyTwoAigs compares ordered input/output interfaces and stores a
// the miter witness in the first graph, following the original ownership rule.
func Cec_ManVerifyTwoAigs(left, right *Aig_Man_t, verbose int32) int32 {
	if left == nil || right == nil {
		return -1
	}
	first0, second0 := gia.Gia_ManFromAig(left), gia.Gia_ManFromAig(right)
	first, second := gia.Gia_ManCleanup(first0), gia.Gia_ManCleanup(second0)
	gia.Gia_ManStop(first0)
	gia.Gia_ManStop(second0)
	defer gia.Gia_ManStop(first)
	defer gia.Gia_ManStop(second)
	status := Cec_ManVerifyTwo(first, second, verbose)
	left.SetCecCounterexample(first.TakeCombinationalCounterexample())
	return status
}
