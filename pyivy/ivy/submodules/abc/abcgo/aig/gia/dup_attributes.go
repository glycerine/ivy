package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManDupWithAttributes(src *Gia_Man_t) *Gia_Man_t {
	out := Gia_ManDup(src)
	if src.VMapping != nil {
		out.VMapping = vec.Vec_IntDup(src.VMapping)
	}
	if src.VPacking != nil {
		out.VPacking = vec.Vec_IntDup(src.VPacking)
	}
	if src.PManTime != nil {
		manager, ok := src.PManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
		out.PManTime = tim.Tim_ManDup(manager, 0)
	}
	if src.PAigExtra != nil {
		out.PAigExtra = Gia_ManDup(src.PAigExtra)
	}
	out.nAnd2Delay = src.nAnd2Delay
	if src.vRegClasses != nil {
		out.vRegClasses = vec.Vec_IntDup(src.vRegClasses)
	}
	if src.vRegInits != nil {
		out.vRegInits = vec.Vec_IntDup(src.vRegInits)
	}
	if src.VConfigs != nil {
		out.VConfigs = vec.Vec_IntDup(src.VConfigs)
	}
	out.PCellStr = src.PCellStr
	return out
}
