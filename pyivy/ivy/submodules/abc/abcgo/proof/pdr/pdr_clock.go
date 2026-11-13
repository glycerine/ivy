package pdr

import "github.com/glycerine/abcgo/misc/util"

// Direct helper calls retain their existing wall-clock domain. Timed solve
// drivers install a CPU clock while the owning goroutine is pinned.
func pdrNow(p *Pdr_Man_t) int64 {
	if p.runtimeClock != nil {
		return p.runtimeClock()
	}
	return util.Abc_Clock()
}
func pdrStartTimedClock(p *Pdr_Man_t) func() {
	if p.pPars.nTimeOut == 0 && p.pPars.nTimeOutOne == 0 && p.pPars.nTimeOutGap == 0 {
		return func() {}
	}
	unpin := pdrPinClock()
	p.runtimeClock = pdrCPUClock
	return unpin
}
