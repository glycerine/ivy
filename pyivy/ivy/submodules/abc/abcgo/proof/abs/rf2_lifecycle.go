package abs

import (
	"fmt"
	"unsafe"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Rf2_ManStart(g *gia.Gia_Man_t) *Rf2_Man_t {
	if g == nil || gia.Gia_ManPoNum(g) != 1 {
		panic("refinement requires one primary output")
	}
	p := &Rf2_Man_t{pGia: g, vObjs: vec.Vec_IntAlloc(1000), vFanins: vec.Vec_IntAlloc(1000), pvVecs: make([]vec.Vec_Int_t, gia.Gia_ManObjNum(g)), vGrp2Ppi: vec.Vec_VecAlloc(100)}
	vec.Vec_VecExpandInt(p.vGrp2Ppi, 99)
	gia.Gia_ManCleanMark0(g)
	gia.Gia_ManCleanMark1(g)
	return p
}

func Rf2_ManStop(p *Rf2_Man_t, profile int32) {
	if p == nil {
		return
	}
	if profile != 0 && p.nCalls != 0 {
		fmt.Printf("Abstraction refinement: calls = %d, average refine = %.1f, GIA memory = %.3f MB, other memory = %.3f MB, time = %v\n", p.nCalls, float64(p.nRefines)/float64(p.nCalls), gia.Gia_ManMemory(p.pGia)/(1<<20), Rf2_ManMemoryUsage(p)/(1<<20), p.timeTotal)
	}
	p.vObjs, p.vFanins, p.vGrp2Ppi, p.pvVecs, p.pObjs = nil, nil, nil, nil, nil
}

func Rf2_ManMemoryUsage(p *Rf2_Man_t) float64 {
	return float64(unsafe.Sizeof(*p) + uintptr(gia.Gia_ManObjNum(p.pGia))*unsafe.Sizeof(vec.Vec_Int_t{}))
}
