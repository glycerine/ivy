package gia

import "github.com/glycerine/abcgo/misc/tim"

// Gia_ManBoxData exposes the timing manager and extra combinational graph
// attached to a GIA manager, keeping their storage owned by package gia.
func Gia_ManBoxData(p *Gia_Man_t) (timing any, extra *Gia_Man_t) {
	if p == nil {
		return nil, nil
	}
	return p.PManTime, p.PAigExtra
}

func Gia_ManSetTimingData(p *Gia_Man_t, timing any, extra *Gia_Man_t) {
	if p != nil {
		p.PManTime, p.PAigExtra = timing, extra
	}
}

func Gia_ManHasReprClasses(p *Gia_Man_t) bool {
	return p != nil && p.PReprs != nil && len(p.PReprs) == int(p.nObjs) && len(p.PNexts) == int(p.nObjs)
}

// Gia_ManStopTimingData releases the optional timing manager and auxiliary
// graph owned by a GIA timing hierarchy.
func Gia_ManStopTimingData(p *Gia_Man_t) {
	if p == nil {
		return
	}
	if timing, ok := p.PManTime.(*tim.Tim_Man_t); ok {
		tim.Tim_ManStop(timing)
	}
	Gia_ManStop(p.PAigExtra)
	p.PManTime, p.PAigExtra = nil, nil
}
