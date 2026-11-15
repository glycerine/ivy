package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
	"unsafe"
)

// Cofactors are ordinary hashed AIGs even when the source has real gates.
func giaAreCopyGate(p *Gia_ManAre_t, obj *Gia_Obj_t) int32 {
	return giaHashMappedGate(p.pNew, p.pAig, obj, func(lit int32) int32 {
		return Gia_ObjLitCopy(p.pAig, lit)
	})
}

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
			return 1
		}
		next.iPrev = state
		if p.fMiter != 0 && giaAreOutputAsserted(p) {
			p.fStopped = 1
			if p.pTarget != nil {
				panic("gia: duplicate reachability target")
			}
			p.pTarget = next
			return 1
		}
		if p.fStopped != 0 || !giaAreRefineSuccessor(p, next) {
			Gia_ManAreRycycleSta(p, next)
			return 1
		}
		Gia_ManAreCubeProcess(p, next)
		p.timeCube += time.Since(start).Microseconds()
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
			Gia_ObjSetValue(obj, giaAreCopyGate(p, obj))
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
			Gia_ObjSetValue(obj, giaAreCopyGate(p, obj))
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
		p.cofTruth = nil
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
			Gia_ObjSetValue(obj, giaAreCopyGate(p, obj))
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
	p.timeAig += time.Since(start).Microseconds()
	return result
}

func Gia_ManArePrintReport(p *Gia_ManAre_t, start time.Time, final int32) {
	depth := int32(0)
	if p.iStaCur > 0 && p.iStaCur < p.nStas {
		depth = Gia_ManAreDepth(p, p.iStaCur)
	}
	// Account for state/tree records and their storage, as the source report
	// does, using Go allocation sizes rather than the C fixed-size pages.
	pointerBytes := float64(unsafe.Sizeof((*Gia_StaAre_t)(nil)))
	memory := float64(unsafe.Sizeof(*p)) +
		float64(len(p.states))*(float64(unsafe.Sizeof(Gia_StaAre_t{}))+4*float64(p.nWords)) +
		float64(len(p.objects))*float64(unsafe.Sizeof(Gia_ObjAre_t{})) +
		float64(cap(p.states)+cap(p.objects))*pointerBytes
	fmt.Printf("States =%10d. Reached =%10d. R = %5.3f. Depth =%6d. Mem =%9.2f MB.  Time =%9.2f sec", p.iStaCur, p.nStas, float64(p.iStaCur)/float64(p.nStas), depth, memory/(1<<20), time.Since(start).Seconds())
	if final != 0 {
		fmt.Println()
	} else {
		fmt.Print("\r")
	}
}
func giaArePrintTiming(p *Gia_ManAre_t, start time.Time) {
	total := time.Since(start).Seconds()
	aigSeconds, cubeSeconds := float64(p.timeAig)/1e6, float64(p.timeCube)/1e6
	for _, part := range []struct {
		name    string
		seconds float64
	}{
		{"Cofactoring", aigSeconds - cubeSeconds},
		{"Containment", cubeSeconds},
		{"Other      ", total - aigSeconds},
		{"TOTAL      ", total},
	} {
		percent := float64(0)
		if total > 0 {
			percent = 100 * part.seconds / total
		}
		fmt.Printf("%s =%9.2f sec (%6.2f %%)\n", part.name, part.seconds, percent)
	}
}

func Gia_ManArePerform(g *Gia_Man_t, maxStates, miter, verbose int32) int32 {
	if g.nRegs > MAX_VARS_NUM {
		fmt.Printf("Currently can only handle circuit with up to %d registers.\n", MAX_VARS_NUM)
		return -1
	}
	g.pCexSeq = nil
	if g.nRegs < 0 || maxStates < 1 {
		return -1
	}
	start := time.Now()
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
			if g.pCexSeq == nil {
				result = -1
			}
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
	status := "Completed"
	if result != 1 {
		status = "Stopped"
	}
	fmt.Printf("%s after finding %d state cubes (%d not contained) with depth %d.  Time =%9.2f sec\n", status, p.nStas, Gia_ManAreListCountUsed(p), depth, time.Since(start).Seconds())
	if cex := g.pCexSeq; cex != nil {
		fmt.Printf("Output %d of miter %q was asserted in frame %d.\n", cex.IPo, g.pName, cex.IFrame)
	}
	if verbose != 0 {
		giaArePrintTiming(p, start)
	}
	if g.nRegs <= 30 {
		countStart := time.Now()
		count := Gia_ManCountMinterms(p)
		fmt.Printf("The number of unique state minterms in computed state cubes is %d.   Time =%9.2f sec\n", count, time.Since(countStart).Seconds())
	}
	Gia_ManAreFree(p)
	if g.pCexSeq != nil {
		if Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 {
			panic("gia: generated reachability counterexample is invalid")
		}
		fmt.Println("Generated counter-example verified correctly.               ")
	}
	return result
}
