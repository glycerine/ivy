package gia

import (
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/vec"
)

func (p *nfMatcher) outputRequired(index int, arrival, delay int32) int32 {
	required := delay
	if p.cuts.pars.FDoAverage != 0 {
		required = arrival * (100 + p.cuts.pars.NRelaxRatio) / 100
	}
	if scl.Scl_ConIsRunning() && scl.Scl_ConHasOutReqs() {
		external := scl.Scl_ConGetOutReq(index)
		if external > 0 && required <= external {
			required = external
		}
	} else if p.cuts.graph.VOutReqs != nil {
		external := scl.Scl_Flt2Int(vec.Vec_FltEntry(p.cuts.graph.VOutReqs, index))
		if external > 0 && required <= external {
			required = min(2*required, external)
		}
	}
	return required
}

func (p *nfMatcher) delayTarget(delay int32) int32 {
	pars := p.cuts.pars
	if pars.MapDelayTarget == 0 && pars.NRelaxRatio != 0 {
		pars.MapDelayTarget = float32(delay * (100 + pars.NRelaxRatio) / 100)
	}
	if pars.MapDelayTarget > 0 && float32(delay) < pars.MapDelayTarget {
		delay = int32(pars.MapDelayTarget)
	}
	return delay
}
