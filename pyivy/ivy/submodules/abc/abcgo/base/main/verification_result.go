package abcmain

// SetVerificationResult atomically replaces the verification result and trace.
func (f *Frame) SetVerificationResult(status, frames int32, cex *Counterexample) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Status, f.NFrames, f.Cex = status, frames, cex
}

// VerificationResult returns a snapshot of the verification result.
func (f *Frame) VerificationResult() (int32, int32, *Counterexample) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.Cex == nil {
		return f.Status, f.NFrames, nil
	}
	copy := *f.Cex
	copy.Bits = append([]uint64(nil), f.Cex.Bits...)
	return f.Status, f.NFrames, &copy
}
