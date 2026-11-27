package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManMemory estimates the manager storage covered by the original report.
func Gia_ManMemory(p *Gia_Man_t) float64 {
	if p == nil {
		return 0
	}
	// giaMan.c, with ABC's x86-64 sizes: sizeof(Gia_Man_t) = 1016, sizeof(Gia_Obj_t) = 12
	size := 1016 + 12*float64(p.nObjs)
	size += 4 * float64(Gia_ManCiNum(p)+Gia_ManCoNum(p)+int32(vec.Vec_IntSize(&p.vHTable)))
	if p.Refs != nil {
		size += 4 * float64(p.nObjs)
	}
	return size + vec.Vec_IntMemory(p.vLevels) + vec.Vec_IntMemory(p.VCellMapping) + vec.Vec_IntMemory(&p.vCopies) + vec.Vec_FltMemory(p.VInArrs) + vec.Vec_FltMemory(p.VOutReqs) + vec.Vec_PtrMemory(p.VNamesIn) + vec.Vec_PtrMemory(p.VNamesOut)
}
