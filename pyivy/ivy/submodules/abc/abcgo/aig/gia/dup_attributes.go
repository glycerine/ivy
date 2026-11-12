package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManDupWithAttributes(src *Gia_Man_t) *Gia_Man_t {
	out := Gia_ManDup(src)
	if src.vMapping != nil {
		out.vMapping = vec.Vec_IntDup(src.vMapping)
	}
	if src.vPacking != nil {
		out.vPacking = vec.Vec_IntDup(src.vPacking)
	}
	if src.pManTime != nil {
		manager, ok := src.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
		out.pManTime = tim.Tim_ManDup(manager, 0)
	}
	if src.pAigExtra != nil {
		out.pAigExtra = Gia_ManDup(src.pAigExtra)
	}
	out.nAnd2Delay = src.nAnd2Delay
	if src.vRegClasses != nil {
		out.vRegClasses = vec.Vec_IntDup(src.vRegClasses)
	}
	if src.vRegInits != nil {
		out.vRegInits = vec.Vec_IntDup(src.vRegInits)
	}
	if src.vConfigs != nil {
		out.vConfigs = vec.Vec_IntDup(src.vConfigs)
	}
	out.pCellStr = src.pCellStr
	return out
}
