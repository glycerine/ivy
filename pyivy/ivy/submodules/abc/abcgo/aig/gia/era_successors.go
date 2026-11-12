package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Gia_ManAreDeriveNexts_rec(p *Gia_ManAre_t, state Gia_PtrAre_t) int32 {
	p.nRecCalls++
	if p.nRecCalls >= MAX_CALL_NUM {
		return 0
	}
	pivot := Gia_ManAreMostUsedPi(p)
	if pivot == nil {
		start := time.Now()
		next := Gia_ManAreCreateStaNew(p)
		if next == nil {
			p.fStopped = 1
			return 0
		}
		next.iPrev = state
		if p.fMiter != 0 && Gia_ManCheckPOstatus(p)&1 != 0 {
			p.fStopped = 1
			if p.pTarget != nil {
				panic("gia: duplicate reachability target")
			}
			p.pTarget = next
			return 1
		}
		Gia_ManAreCubeProcess(p, next)
		p.timeCube += int32(time.Since(start).Microseconds())
		return p.fStopped
	}
	index := int(Gia_ObjCioId(pivot))
	fanout := vec.Vec_VecEntryInt(p.vCiTfos, index)
	saved := vec.Vec_VecEntryInt(p.vCiLits, index)
	if vec.Vec_IntSize(fanout) != vec.Vec_IntSize(saved) {
		panic("gia: invalid CI fanout storage")
	}
	ids := vec.Vec_IntArray(fanout)
	for i, id := range ids {
		obj := Gia_ManObj(p.pAig, id)
		vec.Vec_IntWriteEntry(saved, i, Gia_ObjValue(obj))
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManHashAnd(p.pNew, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		case Gia_ObjIsCo(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ObjFanin0Copy(obj))
		default:
			if Gia_ObjIsCi(obj) == 0 {
				panic("gia: invalid cofactor object")
			}
			Gia_ObjSetValue(obj, 0)
		}
	}
	if Gia_ManAreDeriveNexts_rec(p, state) != 0 {
		return 1
	}
	for _, id := range ids {
		obj := Gia_ManObj(p.pAig, id)
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ManHashAnd(p.pNew, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		case Gia_ObjIsCo(obj) != 0:
			Gia_ObjSetValue(obj, Gia_ObjFanin0Copy(obj))
		default:
			Gia_ObjSetValue(obj, 1)
		}
	}
	if Gia_ManAreDeriveNexts_rec(p, state) != 0 {
		return 1
	}
	for i, id := range ids {
		Gia_ObjSetValue(Gia_ManObj(p.pAig, id), vec.Vec_IntEntry(saved, i))
	}
	return 0
}
func Gia_ManAreDeriveNexts(p *Gia_ManAre_t, state Gia_PtrAre_t) int32 {
	start := time.Now()
	cube := Gia_ManAreSta(p, state)
	if Gia_StaIsUnused(cube) != 0 {
		return 0
	}
	if p.pNew != nil && p.pNew.nObjs > 1000000 {
		Gia_ManStop(p.pNew)
		p.pNew = nil
	}
	if p.pNew == nil {
		p.pNew = Gia_ManStart(10 * p.pAig.nObjs)
		Gia_ManIncrementTravId(p.pNew)
		Gia_ManHashAlloc(p.pNew)
		Gia_ObjSetValue(Gia_ManConst0(p.pAig), 0)
		for i := int32(0); i < Gia_ManCiNum(p.pAig); i++ {
			Gia_ObjSetValue(Gia_ManCi(p.pAig, i), Gia_ManAppendCi(p.pNew))
		}
	}
	for i := int32(0); i < p.pAig.nRegs; i++ {
		ro := Gia_ManRo(p.pAig, i)
		switch {
		case Gia_StaHasValue0(cube, i) != 0:
			Gia_ObjSetValue(ro, 0)
		case Gia_StaHasValue1(cube, i) != 0:
			Gia_ObjSetValue(ro, 1)
		default:
			Gia_ObjSetValue(ro, Gia_ManCi(p.pNew, Gia_ObjCioId(ro)).id<<1)
		}
	}
	for _, obj := range p.pAig.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ObjSetValue(obj, Gia_ManHashAnd(p.pNew, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p.pAig); i++ {
		obj := Gia_ManCo(p.pAig, i)
		Gia_ObjSetValue(obj, Gia_ObjFanin0Copy(obj))
	}
	p.nRecCalls = 0
	result := Gia_ManAreDeriveNexts_rec(p, state)
	if p.nRecCalls >= MAX_CALL_NUM {
		fmt.Printf("Exceeded the limit on the number of transitions from a state cube (%d).\n", MAX_CALL_NUM)
		p.fStopped = 1
	}
	p.timeAig += int32(time.Since(start).Microseconds())
	return result
}

func Gia_ManArePrintReport(p *Gia_ManAre_t, start time.Time, final int32) {
	depth := int32(0)
	if p.iStaCur > 0 && p.iStaCur < p.nStas {
		depth = Gia_ManAreDepth(p, p.iStaCur)
	}
	fmt.Printf("States =%10d. Reached =%10d. R = %5.3f. Depth =%6d. Time = %.2f sec", p.iStaCur, p.nStas, float64(p.iStaCur)/float64(p.nStas), depth, time.Since(start).Seconds())
	if final != 0 {
		fmt.Println()
	} else {
		fmt.Print("\r")
	}
}
func Gia_ManArePerform(g *Gia_Man_t, maxStates, miter, verbose int32) int32 {
	if g.nRegs > MAX_VARS_NUM {
		fmt.Printf("Currently can only handle circuit with up to %d registers.\n", MAX_VARS_NUM)
		return -1
	}
	if g.nRegs <= 0 || maxStates < 1 {
		return -1
	}
	start := time.Now()
	g.pCexSeq = nil
	p := Gia_ManAreCreate(Gia_ManDup(g))
	p.fMiter = miter
	initial := Gia_ManAreCreateStaInit(p)
	if initial == nil {
		Gia_ManAreFree(p)
		return -1
	}
	Gia_ManAreCubeProcess(p, initial)
	result := int32(1)
	for p.iStaCur = 1; p.iStaCur < p.nStas; p.iStaCur++ {
		Gia_ManAreDeriveNexts(p, Gia_Int2Ptr(uint32(p.iStaCur)))
		if p.pTarget != nil {
			g.pCexSeq = Gia_ManAreDeriveCex(p, p.pTarget)
			result = 0
			break
		}
		if p.fStopped != 0 || p.nStas > maxStates {
			result = -1
			break
		}
		if verbose != 0 {
			Gia_ManArePrintReport(p, start, 0)
		}
	}
	Gia_ManArePrintReport(p, start, 1)
	depth := int32(0)
	if p.iStaCur > 0 && p.iStaCur-1 < p.nStas {
		depth = Gia_ManAreDepth(p, p.iStaCur-1)
	}
	fmt.Printf("Reached %d state cubes (%d not contained) with depth %d. Time = %.2f sec\n", p.nStas, Gia_ManAreListCountUsed(p), depth, time.Since(start).Seconds())
	Gia_ManAreFree(p)
	if g.pCexSeq != nil && Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 {
		panic("gia: generated reachability counterexample is invalid")
	}
	return result
}
