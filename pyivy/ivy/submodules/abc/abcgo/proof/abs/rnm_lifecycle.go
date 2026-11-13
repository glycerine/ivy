package abs

import (
	"fmt"
	"unsafe"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// Rnm_ManStart creates a refinement manager for a one-output GIA.
func Rnm_ManStart(graph *gia.Gia_Man_t) *Rnm_Man_t {
	if graph == nil || gia.Gia_ManPoNum(graph) != 1 {
		panic("refinement requires one primary output")
	}
	p := &Rnm_Man_t{
		pGia:       graph,
		vObjs:      vec.Vec_IntAlloc(100),
		vCounts:    vec.Vec_StrStart(int(gia.Gia_ManObjNum(graph))),
		vFanins:    vec.Vec_IntAlloc(1000),
		nObjsAlloc: 10000,
		pObjs:      make([]Rnm_Obj_t, 10000),
	}
	if !gia.Gia_ManHasStaticFanout(graph) {
		gia.Gia_ManStaticFanoutStart(graph)
	}
	gia.Gia_ManCleanValue(graph)
	gia.Gia_ManCleanMark0(graph)
	gia.Gia_ManCleanMark1(graph)
	return p
}

func Rnm_ManStop(p *Rnm_Man_t, profile int32) {
	if p == nil {
		return
	}
	if profile != 0 && p.nCalls != 0 {
		fmt.Printf("Abstraction refinement: calls = %d, average refine = %.1f, GIA memory = %.3f MB, other memory = %.3f MB, time = %v\n", p.nCalls, float64(p.nRefines)/float64(p.nCalls), gia.Gia_ManMemory(p.pGia)/(1<<20), Rnm_ManMemoryUsage(p)/(1<<20), p.timeTotal)
	}
	gia.Gia_ManCleanMark0(p.pGia)
	gia.Gia_ManCleanMark1(p.pGia)
	gia.Gia_ManStaticFanoutStop(p.pGia)
	p.vCounts, p.vFanins, p.vObjs, p.pObjs = nil, nil, nil, nil
}

func Rnm_ManMemoryUsage(p *Rnm_Man_t) float64 {
	if p == nil {
		return 0
	}
	return float64(unsafe.Sizeof(*p) + uintptr(cap(p.pObjs))*unsafe.Sizeof(Rnm_Obj_t{}) + uintptr(vec.Vec_IntCap(p.vObjs))*unsafe.Sizeof(int32(0)))
}
