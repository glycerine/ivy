package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec_ManVerifyTwoAigs compares ordered input/output interfaces and stores a
// validated mismatch in the first graph, following the original ownership rule.
func Cec_ManVerifyTwoAigs(left, right *Aig_Man_t, verbose int32) int32 {
	if left == nil || right == nil {
		return -1
	}
	left.ClearCombinationalCounterexample()
	first, second := gia.Gia_ManFromAig(left), gia.Gia_ManFromAig(right)
	defer gia.Gia_ManStop(first)
	defer gia.Gia_ManStop(second)
	status := Cec_ManVerifyTwo(first, second, verbose)
	if witness := first.CombinationalCounterexample(); witness != nil {
		inputs := make([]bool, left.InputCount())
		// The witness stores register outputs first, followed by primary inputs;
		// the graph interface lists primary inputs before register outputs.
		regs := int(witness.NRegs)
		pis := len(inputs) - regs
		for i := range inputs {
			bit := regs + i
			if i >= pis {
				bit = i - pis
			}
			inputs[i] = witness.PData[bit>>5]&(uint32(1)<<uint(bit&31)) != 0
		}
		if err := left.SetCombinationalCounterexample(inputs, int(witness.IPo)); err != nil {
			return -1
		}
	}
	return status
}
