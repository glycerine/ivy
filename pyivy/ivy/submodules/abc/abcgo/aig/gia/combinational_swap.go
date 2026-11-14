package gia

// SwapCombinationalCounterexamples transfers witness ownership between managers.
// Callers must ensure that each witness is meaningful for its new manager.
func SwapCombinationalCounterexamples(a, b *Gia_Man_t) {
	a.PCexComb, b.PCexComb = b.PCexComb, a.PCexComb
}
