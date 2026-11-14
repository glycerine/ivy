package gia

import "github.com/glycerine/abcgo/misc/tim"

// Gia_ManBoxData exposes the timing manager and extra combinational graph
// attached to a GIA manager, keeping their storage owned by package gia.
func Gia_ManBoxData(p *Gia_Man_t) (timing any, extra *Gia_Man_t) {
	if p == nil {
		return nil, nil
	}
	return p.pManTime, p.pAigExtra
}

func Gia_ManSetTimingData(p *Gia_Man_t, timing any, extra *Gia_Man_t) {
	if p != nil {
		p.pManTime, p.pAigExtra = timing, extra
	}
}

func Gia_ManHasReprClasses(p *Gia_Man_t) bool {
	return p != nil && p.pReprs != nil && len(p.pReprs) == int(p.nObjs) && len(p.pNexts) == int(p.nObjs)
}

// Gia_ManStopTimingData releases the optional timing manager and auxiliary
// graph owned by a GIA timing hierarchy.
func Gia_ManStopTimingData(p *Gia_Man_t) {
	if p == nil {
		return
	}
	if timing, ok := p.pManTime.(*tim.Tim_Man_t); ok {
		tim.Tim_ManStop(timing)
	}
	Gia_ManStop(p.pAigExtra)
	p.pManTime, p.pAigExtra = nil, nil
}
