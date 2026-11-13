package gia

// ClearSequentialCounterexamples releases the output-indexed witness list.
func (p *Gia_Man_t) ClearSequentialCounterexamples() {
	if p != nil {
		p.vSeqModelVec = nil
	}
}
