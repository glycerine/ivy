package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ga2_GlaAbsCount(p *Ga2_Man_t, flops, ands int32) int32 {
	if flops == 0 && ands == 0 {
		panic("abs: missing abstraction count kind")
	}
	var count int32
	for i := 0; i < vec.Vec_IntSize(p.vAbs); i++ {
		obj := gia.Gia_ManObj(p.pGia, vec.Vec_IntEntry(p.vAbs, i))
		if flops != 0 {
			count += gia.Gia_ObjIsRo(p.pGia, obj)
		} else {
			count += gia.Gia_ObjIsAnd(obj)
		}
	}
	return count
}
