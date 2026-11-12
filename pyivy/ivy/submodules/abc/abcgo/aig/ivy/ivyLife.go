// Basic manager lifecycle ported from ABC ivyMan.c (Alan Mishchenko).
package ivy

import (
	"fmt"
	"time"
)

func Ivy_ManStart() *Ivy_Man_t {
	p := &Ivy_Man_t{nTravIds: 1, fCatchExor: 1}
	p.Ghost = Ivy_Obj_t{Id: -1, owner: p}
	p.vPis = make([]*Ivy_Obj_t, 0, 100)
	p.vPos = make([]*Ivy_Obj_t, 0, 100)
	p.vBufs = make([]*Ivy_Obj_t, 0, 100)
	p.vObjs = make([]*Ivy_Obj_t, 0, 100)
	Ivy_ManStartMemory(p)
	p.pConst1 = Ivy_ManFetchMemory(p)
	p.pConst1.fPhase = 1
	p.vObjs = append(p.vObjs, p.pConst1)
	p.nCreated = 1
	p.pTable = make(map[ivyHashKey]*Ivy_Obj_t)
	return p
}
func Ivy_ManStop(p *Ivy_Man_t) {
	if p == nil {
		return
	}
	if !p.memoryStopped && p.time1 != 0 {
		fmt.Printf("Update lev   %.3f sec\n", float64(p.time1)/float64(time.Second))
	}
	if !p.memoryStopped && p.time2 != 0 {
		fmt.Printf("Update levR  %.3f sec\n", float64(p.time2)/float64(time.Second))
	}
	Ivy_ManStopMemory(p)
	*p = Ivy_Man_t{memoryStopped: true}
}

// Ivy_ManStartFrom creates the constant/input boundary of a fresh manager and
// maps the source inputs through pEquiv. It intentionally does not copy logic.
func Ivy_ManStartFrom(source *Ivy_Man_t) *Ivy_Man_t {
	if source == nil || source.memoryStopped || source.pConst1 == nil {
		panic("source Ivy manager is stopped")
	}
	result := Ivy_ManStart()
	source.pConst1.pEquiv = result.pConst1
	for _, old := range source.vPis {
		input := Ivy_ObjCreatePi(result)
		old.pEquiv = input
	}
	return result
}
func Ivy_ManCleanup(p *Ivy_Man_t) int32 {
	old := Ivy_ManNodeNum(p)
	for _, node := range p.vObjs {
		if node != nil && (Ivy_ObjIsNode(node) != 0 || node.Type == IVY_LATCH || node.Type == IVY_BUF) && node.nRefs == 0 {
			Ivy_ObjDelete_rec(p, node, 1)
		}
	}
	return old - Ivy_ManNodeNum(p)
}
