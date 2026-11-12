// Manager lifecycle, ported from ABC hopMan.c (Alan Mishchenko).
package hop

import "fmt"

func Hop_ManStart() *Hop_Man_t {
	p := &Hop_Man_t{nTravIds: 1, fRefCount: 1, nTableSize: 10007}
	p.vPis = Vec_PtrAlloc(100)
	p.vPos = Vec_PtrAlloc(100)
	Hop_ManStartMemory(p)
	p.pConst1 = Hop_ManFetchMemory(p)
	p.pConst1.Type = AIG_CONST1
	p.pConst1.fPhase = 1
	p.pTable = make([]*Hop_Obj_t, p.nTableSize)
	return p
}
func Hop_ManStop(p *Hop_Man_t) {
	if p == nil {
		return
	}
	Vec_PtrFree(p.vPis)
	Vec_PtrFree(p.vPos)
	Vec_PtrFree(p.vObjs)
	Hop_ManStopMemory(p)
	p.pTable = nil
	p.pConst1 = nil
}
func Hop_ManCleanup(p *Hop_Man_t) int32 {
	assert(p.fRefCount != 0, "cleanup requires reference counts, not levels")
	old := Hop_ManNodeNum(p)
	var dangling []*Hop_Obj_t
	for _, head := range p.pTable {
		for o := head; o != nil; o = o.pNext {
			if o.nRefs == 0 {
				dangling = append(dangling, o)
			}
		}
	}
	for _, o := range dangling {
		if o.Type != AIG_NONE && o.nRefs == 0 {
			Hop_ObjDelete_rec(p, o)
		}
	}
	return old - Hop_ManNodeNum(p)
}
func Hop_ManPrintStats(p *Hop_Man_t) {
	fmt.Printf("PI/PO = %d/%d. A = %d. X = %d. Cre = %d. Del = %d. Lev = %d.\n", Hop_ManPiNum(p), Hop_ManPoNum(p), Hop_ManAndNum(p), Hop_ManExorNum(p), p.nCreated, p.nDeleted, Hop_ManCountLevels(p))
}
