package gia

import "github.com/glycerine/abcgo/misc/tim"

// BoxTiming returns borrowed timing and functional-model metadata.
func (p *Gia_Man_t) BoxTiming() (*tim.Tim_Man_t, *Gia_Man_t) {
	if p.PManTime == nil {
		return nil, p.PAigExtra
	}
	timing, ok := p.PManTime.(*tim.Tim_Man_t)
	if !ok {
		panic("gia: invalid timing manager")
	}
	return timing, p.PAigExtra
}

// SetBoxTiming installs timing and functional-model metadata on the manager.
func (p *Gia_Man_t) SetBoxTiming(timing *tim.Tim_Man_t, extra *Gia_Man_t) {
	if timing == nil {
		p.PManTime = nil
	} else {
		p.PManTime = timing
	}
	p.PAigExtra = extra
}

// Gia_ManClearLevels releases the cached level vector.
func Gia_ManClearLevels(p *Gia_Man_t) { p.vLevels = nil }

// Gia_ManHasLevels reports whether a manager owns a cached level vector.
func Gia_ManHasLevels(p *Gia_Man_t) bool { return p.vLevels != nil }
