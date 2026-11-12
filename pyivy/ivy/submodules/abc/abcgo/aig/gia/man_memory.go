package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"unsafe"
)

// Gia_ManMemory estimates the manager storage covered by the original report.
func Gia_ManMemory(p *Gia_Man_t) float64 {
	if p == nil {
		return 0
	}
	size := float64(unsafe.Sizeof(*p)) + float64(unsafe.Sizeof(Gia_Obj_t{}))*float64(p.nObjs)
	size += 4 * float64(Gia_ManCiNum(p)+Gia_ManCoNum(p)+int32(vec.Vec_IntSize(&p.vHTable)))
	if p.Refs != nil {
		size += 4 * float64(p.nObjs)
	}
	return size + vec.Vec_IntMemory(p.vLevels) + vec.Vec_IntMemory(p.vCellMapping) + vec.Vec_IntMemory(&p.vCopies) + vec.Vec_FltMemory(p.vInArrs) + vec.Vec_FltMemory(p.vOutReqs) + vec.Vec_PtrMemory(p.vNamesIn) + vec.Vec_PtrMemory(p.vNamesOut)
}
