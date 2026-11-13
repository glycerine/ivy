package gia

import "github.com/glycerine/abcgo/misc/tim"

// BoxTiming returns borrowed timing and functional-model metadata.
func (p *Gia_Man_t) BoxTiming() (*tim.Tim_Man_t, *Gia_Man_t) {
	if p.pManTime == nil {
		return nil, p.pAigExtra
	}
	timing, ok := p.pManTime.(*tim.Tim_Man_t)
	if !ok {
		panic("gia: invalid timing manager")
	}
	return timing, p.pAigExtra
}

// SetBoxTiming installs timing and functional-model metadata on the manager.
func (p *Gia_Man_t) SetBoxTiming(timing *tim.Tim_Man_t, extra *Gia_Man_t) {
	if timing == nil {
		p.pManTime = nil
	} else {
		p.pManTime = timing
	}
	p.pAigExtra = extra
}

// Gia_ManClearLevels releases the cached level vector.
func Gia_ManClearLevels(p *Gia_Man_t) { p.vLevels = nil }
