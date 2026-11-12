package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec_ManVerifyTwoAigs compares ordered input/output interfaces and stores a
// validated mismatch in the first graph, following the original ownership rule.
func Cec_ManVerifyTwoAigs(left, right *Aig_Man_t, verbose int32) int32 {
	left.ClearCombinationalCounterexample()
	first, second := gia.Gia_ManFromAig(left), gia.Gia_ManFromAig(right)
	defer gia.Gia_ManStop(first)
	defer gia.Gia_ManStop(second)
	status := Cec_ManVerifyTwo(first, second, verbose)
	if witness := first.CombinationalCounterexample(); witness != nil {
		inputs := make([]bool, left.InputCount())
		// Combinational verification rejects registers, so the input frame starts
		// at bit zero; the attached utility model owns its own bit storage.
		for i := range inputs {
			inputs[i] = witness.PData[i>>5]&(uint32(1)<<uint(i&31)) != 0
		}
		if err := left.SetCombinationalCounterexample(inputs, int(witness.IPo)); err != nil {
			return -1
		}
	}
	return status
}
