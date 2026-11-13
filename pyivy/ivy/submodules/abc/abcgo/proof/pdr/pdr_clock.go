package pdr

import (
	"time"

	"github.com/glycerine/abcgo/misc/util"
)

// Direct helper calls retain their existing wall-clock domain. Timed solve
// drivers install a CPU clock while the owning goroutine is pinned.
func pdrNow(p *Pdr_Man_t) int64 {
	if p.runtimeClock != nil {
		return p.runtimeClock()
	}
	return util.Abc_Clock()
}

// The C drivers handle global and gap timeouts before dropping one output.
func pdrOnlyOutputTimedOut(p *Pdr_Man_t, now int64) bool {
	if p.timeToStop != 0 && now > p.timeToStop {
		return false
	}
	if p.pPars.nTimeOutGap != 0 && p.pPars.timeLastSolved != 0 && now > p.pPars.timeLastSolved+int64(p.pPars.nTimeOutGap)*int64(time.Second) {
		return false
	}
	return p.timeToStopOne != 0 && now > p.timeToStopOne
}
func pdrStartTimedClock(p *Pdr_Man_t) func() {
	if p.pPars.nTimeOut == 0 && p.pPars.nTimeOutOne == 0 && p.pPars.nTimeOutGap == 0 {
		return func() {}
	}
	unpin := pdrPinClock()
	if p.runtimeClock == nil {
		p.runtimeClock = pdrCPUClock
	}
	return unpin
}
