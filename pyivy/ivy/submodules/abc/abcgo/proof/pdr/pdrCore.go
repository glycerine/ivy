package pdr

import (
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/hash"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [pdrCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrCore.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"
//#include "base/main/main.h"
//#include "misc/hash/hash.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Gia_ManToBridgeResult( FILE * pFile, int Result, Abc_Cex_t * pCex, int iPoProved );

// extern int Gia_ManToBridgeAbort( FILE * pFile, int Size, unsigned char * pBuffer );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the state could be blocked.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManSetDefaultParams(pPars *Pdr_Par_t) {
	*pPars = Pdr_Par_t{
		NRecycle: 300, NFrameMax: 10000, NRandomSeed: 91648253,
		FUseSupp: 1, FUsePropOut: 1, FSkipDown: 1, IFrame: -1,
	}
}

/*
*Function*************************************************************

	Synopsis    [Reduces clause using analyzeFinal.]

	Description [Assumes that the SAT solver just terminated an UNSAT call.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManReduceClause(p *Pdr_Man_t, k int32, pCube *Pdr_Set_t) *Pdr_Set_t {
	core := Pdr_ManSolver(p, k).UnsatCore()
	literals := Pdr_ManLitsToCube(p, k, core, int32(len(core)))
	if vec.Vec_IntSize(literals) == int(pCube.nLits) {
		return nil
	}
	if vec.Vec_IntSize(literals) > int(pCube.nLits) {
		panic("pdr: unsat core exceeds cube")
	}
	hasPositive := false
	for _, lit := range vec.Vec_IntArray(literals) {
		if lit&1 == 0 {
			hasPositive = true
			break
		}
	}
	if !hasPositive {
		for _, lit := range pCube.Lits[:pCube.nLits] {
			if lit&1 == 0 {
				vec.Vec_IntPush(literals, lit)
				hasPositive = true
				break
			}
		}
		if !hasPositive {
			panic("pdr: cube includes initial state")
		}
	}
	reduced := Pdr_SetCreateSubset(pCube, vec.Vec_IntArray(literals), int32(vec.Vec_IntSize(literals)))
	if Pdr_SetIsInit(reduced, -1) != 0 {
		panic("pdr: reduced cube includes initial state")
	}
	return reduced
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the state could be blocked.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManPushClauses(p *Pdr_Man_t) int32 {
	if p.iUseFrame <= 0 {
		panic("pdr: invalid starting frame")
	}
	start := util.Abc_Clock()
	last := len(p.vSolvers) - 1
	first := 1
	if p.pPars.FShiftStart != 0 {
		first = int(p.iUseFrame)
	}
	sortCubes := func(frame *vec.Vec_Ptr_t) {
		vec.Vec_PtrSort(frame, func(a, b *any) int {
			left, right := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
			return int(Pdr_SetCompare(&left, &right))
		})
	}
	removeCovered := func(frame *vec.Vec_Ptr_t, cube *Pdr_Set_t, first int) {
		for i := first; i < vec.Vec_PtrSize(frame); {
			candidate := vec.Vec_PtrEntry(frame, i).(*Pdr_Set_t)
			if Pdr_SetContains(candidate, cube) == 0 {
				i++
				continue
			}
			Pdr_SetDeref(candidate)
			vec.Vec_PtrWriteEntry(frame, i, vec.Vec_PtrEntryLast(frame))
			vec.Vec_PtrPop(frame)
		}
	}
	var result int32
	for k := first; k < last; k++ {
		current := vec.Vec_VecEntry(p.vClauses, k)
		next := vec.Vec_VecEntry(p.vClauses, k+1)
		sortCubes(current)
		for j := 0; j < vec.Vec_PtrSize(current); {
			cube := vec.Vec_PtrEntry(current, j).(*Pdr_Set_t)
			removeCovered(current, cube, j+1)
			status := Pdr_ManCheckCube(p, int32(k), cube, nil, 0, 0, 1)
			if status == -1 {
				return -1
			}
			if status == 0 {
				j++
				continue
			}
			if reduced := Pdr_ManReduceClause(p, int32(k), cube); reduced != nil {
				Pdr_SetDeref(cube)
				cube = reduced
			}
			Pdr_ManSolverAddClause(p, int32(k+1), cube)
			removeCovered(next, cube, 0)
			vec.Vec_PtrPush(next, cube)
			vec.Vec_PtrWriteEntry(current, j, vec.Vec_PtrEntryLast(current))
			vec.Vec_PtrPop(current)
		}
		if vec.Vec_PtrSize(current) == 0 {
			result = 1
		}
	}
	final := vec.Vec_VecEntry(p.vClauses, last)
	sortCubes(final)
	for j := 0; j < vec.Vec_PtrSize(final); j++ {
		removeCovered(final, vec.Vec_PtrEntry(final, j).(*Pdr_Set_t), j+1)
	}
	p.tPush += util.Abc_Clock() - start
	return result
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the clause is contained in higher clauses.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManCheckContainment(p *Pdr_Man_t, k int32, pSet *Pdr_Set_t) int32 {
	for frame := int(k); frame < len(p.vSolvers); frame++ {
		for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
			if Pdr_SetContains(pSet, item.(*Pdr_Set_t)) != 0 {
				return 1
			}
		}
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Sorts literals by priority.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManSortByPriority(p *Pdr_Man_t, cube *Pdr_Set_t) []int32 {
	n := int(cube.nLits)
	if cap(p.pOrder) < n {
		p.pOrder = make([]int32, n)
	} else {
		p.pOrder = p.pOrder[:n]
	}
	priorities := vec.Vec_IntArray(p.vPrio)
	for i := range p.pOrder {
		p.pOrder[i] = int32(i)
	}
	for i := 0; i+1 < n; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if priorities[cube.Lits[p.pOrder[j]]>>1] < priorities[cube.Lits[p.pOrder[best]]>>1] {
				best = j
			}
		}
		p.pOrder[i], p.pOrder[best] = p.pOrder[best], p.pOrder[i]
	}
	return p.pOrder
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func ZPdr_ManSimpleMic(p *Pdr_Man_t, k int32, cube **Pdr_Set_t) int32 {
	if p.pPars.FSkipGeneral != 0 {
		return 0
	}
	order := Pdr_ManSortByPriority(p, *cube)
	for j := 0; j < int((*cube).nLits); j++ {
		i := order[j]
		if (*cube).Lits[i] < 0 {
			panic("pdr: removed literal in starting cube")
		}
		if Pdr_SetIsInit(*cube, i) != 0 {
			continue
		}
		lit := (*cube).Lits[i]
		(*cube).Lits[i] = -1
		status := Pdr_ManCheckCube(p, k, *cube, nil, p.pPars.NConfLimit, 0, 1)
		if status == -1 {
			return -1
		}
		(*cube).Lits[i] = lit
		if status == 0 {
			continue
		}
		old := *cube
		*cube = Pdr_SetCreateFrom(old, i)
		Pdr_SetDeref(old)
		if (*cube).nLits == 0 {
			panic("pdr: empty generalized cube")
		}
		order = Pdr_ManSortByPriority(p, *cube)
		j--
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func ZPdr_ManDown(p *Pdr_Man_t, k int32, cube **Pdr_Set_t, pred *Pdr_Set_t, keep *hash.Hash_Int_t, inductive *Pdr_Set_t, added *int32) int32 {
	last := int32(len(p.vSolvers) - 1)
	// This helper consumes the predecessor reference on all return paths.
	defer func() {
		if pred != nil {
			Pdr_SetDeref(pred)
		}
	}()
	for {
		for attempts := 0; p.pPars.FCtgs != 0 && k > 1 && attempts < 3; attempts++ {
			ctg := Pdr_SetDup(pred)
			if Pdr_SetIsInit(ctg, -1) != 0 {
				Pdr_SetDeref(ctg)
				break
			}
			if *added == 0 {
				for frame := int32(1); frame <= k; frame++ {
					Pdr_ManSolverAddClause(p, frame, inductive)
				}
				*added = 1
			}
			status := Pdr_ManCheckCube(p, k-1, ctg, nil, p.pPars.NConfLimit, 0, 1)
			if status != 1 {
				Pdr_SetDeref(ctg)
				break
			}
			minimum := Pdr_ManReduceClause(p, k-1, ctg)
			if minimum == nil {
				minimum = Pdr_SetDup(ctg)
			}
			frame := k
			for ; frame < last; frame++ {
				status = Pdr_ManCheckCube(p, frame, minimum, nil, 0, 0, 1)
				if status == 0 {
					break
				}
			}
			if ZPdr_ManSimpleMic(p, frame-1, &minimum) == -1 {
				panic("pdr: CTG MIC returned unknown")
			}
			if p.pPars.FVeryVerbose != 0 {
				fmt.Fprint(os.Stdout, "Adding cube ")
				Pdr_SetPrint(os.Stdout, minimum, aig.Aig_ManRegNum(p.pAig), nil)
				fmt.Fprintf(os.Stdout, " to frame %d.\n", frame)
			}
			for _, lit := range minimum.Lits[:minimum.nLits] {
				reg := lit >> 1
				if lit < 0 || reg >= aig.Aig_ManRegNum(p.pAig) {
					panic("pdr: invalid priority literal")
				}
				if vec.Vec_IntEntry(p.vPrio, int(reg))>>p.nPrioShift == 0 {
					p.nAbsFlops++
				}
				vec.Vec_IntAddToEntry(p.vPrio, int(reg), int32(1)<<p.nPrioShift)
			}
			vec.Vec_VecPush(p.vClauses, int(frame), minimum)
			p.nCubes++
			for i := int32(1); i <= frame; i++ {
				Pdr_ManSolverAddClause(p, i, minimum)
			}
			Pdr_SetDeref(pred)
			pred = nil
			status = Pdr_ManCheckCube(p, k, *cube, &pred, p.pPars.NConfLimit, 0, 1)
			if status == -1 {
				panic("pdr: post-CTG cube query returned unknown")
			}
			Pdr_SetDeref(ctg)
			if status == 1 {
				return 1
			}
		}
		if p.pPars.FVeryVerbose != 0 {
			fmt.Println("Cube:")
			ZPdr_SetPrint(*cube)
			fmt.Println("Pred:")
			ZPdr_SetPrint(pred)
		}
		old := *cube
		*cube = ZPdr_SetIntersection(old, pred, keep)
		Pdr_SetDeref(old)
		Pdr_SetDeref(pred)
		pred = nil
		if *cube == nil {
			return 0
		}
		if p.pPars.FVeryVerbose != 0 {
			fmt.Println("Intersection:")
			ZPdr_SetPrint(*cube)
		}
		if Pdr_SetIsInit(*cube, -1) != 0 {
			if p.pPars.FVeryVerbose != 0 {
				fmt.Println("Failed initiation")
			}
			return 0
		}
		pred = nil
		status := Pdr_ManCheckCube(p, k, *cube, &pred, p.pPars.NConfLimit, 0, 1)
		if status == -1 {
			return -1
		}
		if status == 1 {
			return 1
		}
		if (*cube).nLits == 1 {
			return 0
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [Specialized sorting of flops based on priority.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vec_IntSelectSortPrioReverseLit(array []int32, nSize int32, priorities *Vec_Int_t) int32 {
	if nSize < 0 || int(nSize) > len(array) {
		panic("pdr: invalid sort length")
	}
	prio := vec.Vec_IntArray(priorities)
	for i := 0; i+1 < int(nSize); i++ {
		best := i
		for j := i + 1; j < int(nSize); j++ {
			if prio[array[j]>>1] > prio[array[best]>>1] {
				best = j
			}
		}
		array[i], array[best] = array[best], array[i]
	}
	return 1
}

/**Function*************************************************************

  Synopsis    [Performs generalization using a different idea.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
// Pdr_ManGeneralize2 is inactive in ABC: its entire implementation is under
// #if 0 in pdrCore.c and the compiled C function returns zero.
func Pdr_ManGeneralize2(_ *Pdr_Man_t, _ int32, _ *Pdr_Set_t, _ **Pdr_Set_t) int32 { return 0 }

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the state could be blocked.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManGeneralize(p *Pdr_Man_t, k int32, cube *Pdr_Set_t, predecessor **Pdr_Set_t, minimum **Pdr_Set_t) int32 {
	started := util.Abc_Clock()
	*minimum = nil
	if p.pPars.FFlopOrder != 0 {
		Vec_IntSelectSortPrioReverseLit(cube.Lits, cube.nLits, p.vPrio)
	}
	status := Pdr_ManCheckCube(p, k, cube, predecessor, p.pPars.NConfLimit, 0, 1)
	if p.pPars.FFlopOrder != 0 {
		slices.Sort(cube.Lits[:cube.nLits])
	}
	if status == -1 {
		return -1
	}
	if status == 0 {
		p.tGeneral += util.Abc_Clock() - started
		return 0
	}
	current := Pdr_ManReduceClause(p, k, cube)
	if current == nil {
		current = Pdr_SetDup(cube)
	}
	if current.nLits == 0 {
		panic("pdr: empty generalized cube")
	}
	if p.pPars.FSimpleGeneral != 0 {
		// ABC's alternative body is compiled out. Its caller still discards
		// the reduced cube when it contains more than one literal.
		if current.nLits > 1 {
			Pdr_ManGeneralize2(p, k, current, minimum)
			Pdr_SetDeref(current)
			current = *minimum
		}
		*minimum = current
		p.tGeneral += util.Abc_Clock() - started
		return 1
	}
	var keep *hash.Hash_Int_t
	if p.pPars.FSkipDown == 0 {
		keep = hash.Hash_IntAlloc(1)
		defer hash.Hash_IntFree(keep)
	}
	added := int32(0)
	if p.pPars.FSkipGeneral == 0 {
		order := Pdr_ManSortByPriority(p, current)
		for j := 0; j < int(current.nLits); j++ {
			i := order[j]
			if current.Lits[i] < 0 {
				panic("pdr: removed literal in starting cube")
			}
			if keep != nil && hash.Hash_IntExists(keep, current.Lits[i]) != 0 {
				continue
			}
			if Pdr_SetIsInit(current, i) != 0 {
				continue
			}
			lit := current.Lits[i]
			current.Lits[i] = -1
			var pred *Pdr_Set_t
			var predOut **Pdr_Set_t
			if p.pPars.FSkipDown == 0 {
				predOut = &pred
			}
			status = Pdr_ManCheckCube(p, k, current, predOut, p.pPars.NConfLimit, 1, 1)
			current.Lits[i] = lit
			if status == -1 {
				Pdr_SetDeref(current)
				return -1
			}
			if status == 0 {
				if p.pPars.FSkipDown != 0 {
					continue
				}
				candidate := Pdr_SetCreateFrom(current, i)
				status = ZPdr_ManDown(p, k, &candidate, pred, keep, current, &added)
				if p.pPars.FCtgs != 0 {
					order = Pdr_ManSortByPriority(p, current)
				}
				if status == -1 {
					Pdr_SetDeref(current)
					if candidate != nil {
						Pdr_SetDeref(candidate)
					}
					return -1
				}
				if status == 0 {
					if keep != nil {
						hash.Hash_IntWriteEntry(keep, lit, 0)
					}
					if candidate != nil {
						Pdr_SetDeref(candidate)
					}
					continue
				}
				added = 0
				Pdr_SetDeref(current)
				current = candidate
				if current == nil || current.nLits == 0 {
					panic("pdr: empty generalized cube")
				}
				order = Pdr_ManSortByPriority(p, current)
				j = -1
				continue
			}
			added = 0
			old := current
			current = Pdr_SetCreateFrom(old, i)
			Pdr_SetDeref(old)
			if current.nLits == 0 {
				panic("pdr: empty generalized cube")
			}
			order = Pdr_ManSortByPriority(p, current)
			j--
		}
		if p.pPars.FTwoRounds != 0 {
			if ZPdr_ManSimpleMic(p, k, &current) == -1 {
				Pdr_SetDeref(current)
				return -1
			}
		}
	}
	if p.pPars.FVeryVerbose != 0 {
		fmt.Println("Cube:")
		ZPdr_SetPrint(current)
	}
	*minimum = current
	p.tGeneral += util.Abc_Clock() - started
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if the state could be blocked.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManBlockCube(p *Pdr_Man_t, cube *Pdr_Set_t) int32 {
	priority := int32(util.ABC_INFINITY)
	lastFrame := int32(len(p.vSolvers)) - 1
	p.nBlocks++
	obligation := Pdr_OblStart(lastFrame, priority, cube, nil)
	priority--
	Pdr_QueuePush(p, obligation)
	for Pdr_QueueIsEmpty(p) == 0 {
		obligation = Pdr_QueueHead(p)
		if obligation.iFrame == 0 || (p.pPars.FUseAbs != 0 && Pdr_SetIsInit(obligation.pState, -1) != 0) {
			return 0
		}
		if obligation.iFrame > lastFrame {
			return 1
		}
		if p.nQueLim != 0 && p.nQueCur >= p.nQueLim {
			p.nQueLim = p.nQueLim * 3 / 2
			Pdr_QueueStop(p)
			return 1
		}
		obligation = Pdr_QueuePop(p)
		if obligation.iFrame <= 0 {
			panic("pdr: invalid proof obligation frame")
		}
		if Pdr_SetIsInit(obligation.pState, -1) != 0 {
			panic("pdr: proof obligation includes initial state")
		}
		if obligation.iFrame < p.iUseFrame {
			p.iUseFrame = obligation.iFrame
		}
		start := util.Abc_Clock()
		contained := Pdr_ManCheckContainment(p, obligation.iFrame, obligation.pState)
		p.tContain += util.Abc_Clock() - start
		if contained != 0 {
			Pdr_OblDeref(obligation)
			continue
		}
		status := Pdr_ManCheckCubeCs(p, obligation.iFrame, obligation.pState)
		if status == -1 {
			Pdr_OblDeref(obligation)
			return -1
		}
		if status != 0 {
			Pdr_OblDeref(obligation)
			continue
		}
		var predecessor, minimum *Pdr_Set_t
		status = Pdr_ManGeneralize(p, obligation.iFrame-1, obligation.pState, &predecessor, &minimum)
		if status == -1 {
			Pdr_OblDeref(obligation)
			return -1
		}
		if status != 0 {
			if minimum == nil || predecessor != nil {
				panic("pdr: invalid inductive generalization result")
			}
			k := obligation.iFrame
			for ; k < lastFrame; k++ {
				status = Pdr_ManCheckCube(p, k, minimum, nil, 0, 0, 1)
				if status == -1 {
					Pdr_SetDeref(minimum)
					Pdr_OblDeref(obligation)
					return -1
				}
				if status == 0 {
					break
				}
			}
			if p.pPars.FVeryVerbose != 0 {
				fmt.Print("Adding cube ")
				Pdr_SetPrint(os.Stdout, minimum, aig.Aig_ManRegNum(p.pAig), nil)
				fmt.Printf(" to frame %d.\n", k)
			}
			for _, lit := range minimum.Lits[:minimum.nLits] {
				index := lit / 2
				if (vec.Vec_IntEntry(p.vPrio, int(index)) >> p.nPrioShift) == 0 {
					p.nAbsFlops++
				}
				vec.Vec_IntAddToEntry(p.vPrio, int(index), 1<<p.nPrioShift)
			}
			vec.Vec_VecPush(p.vClauses, int(k), minimum)
			p.nCubes++
			for i := int32(1); i <= k; i++ {
				Pdr_ManSolverAddClause(p, i, minimum)
			}
			if (k < lastFrame || p.pPars.FReuseProofOblig != 0) && p.pPars.FShortest == 0 {
				obligation.iFrame = k + 1
				obligation.prio = priority
				priority--
				Pdr_QueuePush(p, obligation)
			} else {
				Pdr_OblDeref(obligation)
			}
		} else {
			if minimum != nil || predecessor == nil {
				panic("pdr: invalid predecessor generalization result")
			}
			obligation.prio = priority
			priority--
			Pdr_QueuePush(p, obligation)
			next := Pdr_OblStart(obligation.iFrame-1, priority, predecessor, Pdr_OblRef(obligation))
			priority--
			Pdr_QueuePush(p, next)
		}
		if p.pPars.pFuncStop != nil && p.pPars.pFuncStop(p.pPars.RunId) != 0 {
			return -1
		}
		now := pdrNow(p)
		if p.timeToStop != 0 && now > p.timeToStop || p.timeToStopOne != 0 && now > p.timeToStopOne ||
			p.pPars.NTimeOutGap != 0 && p.pPars.timeLastSolved != 0 && now > p.pPars.timeLastSolved+int64(p.pPars.NTimeOutGap)*int64(time.Second) {
			return -1
		}
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Pdr_ManSolveInt(p *Pdr_Man_t) int32 {
	defer pdrStartTimedClock(p)()
	pars := p.pPars
	graph := p.pAig
	outputs := graph.TruePoCount()
	inputs := graph.TruePiCount()
	nOutDigits := int(util.Abc_Base10Log(uint32(outputs)))
	clkStart := util.Abc_Clock()
	var clkOne abctime
	var pCube *Pdr_Set_t
	var RetValue int32 = -1
	gapExpired := func() bool {
		return pars.NTimeOutGap != 0 && pars.timeLastSolved != 0 && pdrNow(p) > pars.timeLastSolved+int64(pars.NTimeOutGap)*int64(time.Second)
	}
	progress := func(fClose int32) { Pdr_ManPrintProgress(p, fClose, util.Abc_Clock()-clkStart) }
	if pars.NTimeOut != 0 {
		p.timeToStop = pdrNow(p) + int64(pars.NTimeOut)*int64(time.Second)
	} else {
		p.timeToStop = 0
	}
	if len(p.vSolvers) != 0 {
		panic("pdr: solver already initialized")
	}
	outputEdges := graph.Outputs()[:outputs]
	// in the multi-output mode, mark trivial POs (those fed by const0) as solved
	if pars.FSolveAll != 0 {
		for i, output := range outputEdges {
			if aig.Aig_ObjChild0(graph, output) == graph.Const0() {
				vec.Vec_IntWriteEntry(pars.VOutMap, i, 1) // unsat
				pars.NProveOuts++
				if pars.FUseBridge != 0 {
					util.Gia_ManToBridgeResult(os.Stdout, 1, nil, int32(i))
				}
			}
		}
	}
	// create the first timeframe
	pars.timeLastSolved = pdrNow(p)
	iFrame := int32(0)
	Pdr_ManCreateSolver(p, iFrame)
	for {
		fRefined := false
		if pars.FUseAbs != 0 && p.vAbsFlops == nil && iFrame == 1 {
			regs := int(aig.Aig_ManRegNum(graph))
			p.vAbsFlops = vec.Vec_IntStart(regs)
			p.vMapFf2Ppi = vec.Vec_IntStartFull(regs)
			p.vMapPpi2Ff = vec.Vec_IntAlloc(100)
		}
		p.nFrames = iFrame
		if int(iFrame) != len(p.vSolvers)-1 {
			panic("pdr: frame and solver count disagree")
		}
		p.iUseFrame = max(iFrame, 1)
		for index, output := range outputEdges {
			p.iOutCur = int32(index)
			// skip disproved outputs
			if p.vCexes != nil && vec.Vec_PtrEntry(p.vCexes, index) != nil {
				continue
			}
			// skip output whose time has run out
			if p.pTime4Outs != nil && p.pTime4Outs[index] == 0 {
				continue
			}
			// check if the output is trivially solved
			driver := aig.Aig_ObjChild0(graph, output)
			if driver == graph.Const0() {
				continue
			}
			// check if the output is trivially solved
			if driver == graph.Const1() {
				if pars.FSolveAll == 0 {
					pCexNew := util.Abc_CexMakeTriv(aig.Aig_ManRegNum(graph), inputs, outputs, iFrame*outputs+int32(index))
					if err := graph.SetSequentialCounterexample(pCexNew); err != nil {
						panic(err)
					}
					return 0 // SAT
				}
				var pCexNew *util.Abc_Cex_t
				if pars.FUseBridge != 0 || pars.FStoreCex != 0 {
					pCexNew = util.Abc_CexMakeTriv(aig.Aig_ManRegNum(graph), inputs, outputs, iFrame*outputs+int32(index))
				}
				pars.NFailOuts++
				if pars.VOutMap != nil {
					vec.Vec_IntWriteEntry(pars.VOutMap, index, 0)
				}
				if pars.FNotVerbose == 0 {
					util.Abc_Print(1, "Output %*d was trivially asserted in frame %2d (solved %*d out of %*d outputs).\n",
						nOutDigits, index, iFrame, nOutDigits, pars.NFailOuts, nOutDigits, outputs)
				}
				if pars.FUseBridge != 0 {
					util.Gia_ManToBridgeResult(os.Stdout, 0, pCexNew, pCexNew.IPo)
				}
				pdrStoreCex(p, index, pCexNew)
				if pdrCallOnFail(p, index, pCexNew) {
					if pars.FVerbose != 0 {
						progress(1)
					}
					if pars.FSilent == 0 {
						util.Abc_Print(1, "Quitting due to callback on fail in frame %d.\n", iFrame)
					}
					pars.IFrame = iFrame
					return -1
				}
				if pars.NFailOuts+pars.NDropOuts == outputs {
					if pars.NFailOuts != 0 {
						return 0 // SAT
					}
					return -1 // UNDEC
				}
				pars.timeLastSolved = pdrNow(p)
				continue
			}
			// try to solve this output
			if p.pTime4Outs != nil {
				if p.pTime4Outs[index] <= 0 {
					panic("pdr: output time budget exhausted")
				}
				clkOne = pdrNow(p)
				p.timeToStopOne = p.pTime4Outs[index] + clkOne
			}
			for {
				if gapExpired() {
					if pars.FVerbose != 0 {
						progress(1)
					}
					if pars.FSilent == 0 {
						util.Abc_Print(1, "Reached gap timeout (%d seconds) in frame %d.\n", pars.NTimeOutGap, iFrame)
					}
					pars.IFrame = iFrame
					return -1
				}
				RetValue = Pdr_ManCheckCube(p, iFrame, nil, &pCube, pars.NConfLimit, 0, 1)
				if RetValue == 1 {
					break
				}
				if RetValue == -1 {
					if pars.FVerbose != 0 {
						progress(1)
					}
					if pdrReportStop(p, iFrame) {
						pCube = nil
						break // keep solving
					}
					pars.IFrame = iFrame
					return -1
				}
				if RetValue == 0 {
					RetValue = Pdr_ManBlockCube(p, pCube)
					if RetValue == -1 {
						if pars.FVerbose != 0 {
							progress(1)
						}
						if pdrReportStop(p, iFrame) {
							pCube = nil
							break // keep solving
						}
						pars.IFrame = iFrame
						return -1
					}
					if RetValue == 0 {
						if pars.FVerbose != 0 && pars.FUseAbs == 0 {
							progress(pdrBool(pars.FSolveAll == 0))
						}
						pars.IFrame = iFrame
						if pars.FSolveAll == 0 {
							clk := util.Abc_Clock()
							pCex := Pdr_ManDeriveCexAbs(p)
							p.tAbs += util.Abc_Clock() - clk
							if pCex == nil {
								if pars.FUseAbs == 0 {
									panic("pdr: no counterexample without abstraction")
								}
								Pdr_QueueClean(p)
								pCube = nil
								fRefined = true
								break // keep solving
							}
							if err := graph.SetSequentialCounterexample(pCex); err != nil {
								panic(err)
							}
							return 0 // SAT
						}
						pars.NFailOuts++
						var pCexNew *util.Abc_Cex_t
						if pars.FUseBridge != 0 || pars.FStoreCex != 0 {
							pCexNew = Pdr_ManDeriveCex(p)
						}
						if pars.VOutMap != nil {
							vec.Vec_IntWriteEntry(pars.VOutMap, index, 0)
						}
						if pars.FUseBridge != 0 {
							util.Gia_ManToBridgeResult(os.Stdout, 0, pCexNew, pCexNew.IPo)
						}
						pdrStoreCex(p, index, pCexNew)
						if pdrCallOnFail(p, index, pCexNew) {
							if pars.FVerbose != 0 {
								progress(1)
							}
							if pars.FSilent == 0 {
								util.Abc_Print(1, "Quitting due to callback on fail in frame %d.\n", iFrame)
							}
							pars.IFrame = iFrame
							return -1
						}
						if pars.FNotVerbose == 0 {
							util.Abc_Print(1, "Output %*d was asserted in frame %2d (%2d) (solved %*d out of %*d outputs).\n",
								nOutDigits, index, iFrame, iFrame, nOutDigits, pars.NFailOuts, nOutDigits, outputs)
						}
						if pars.NFailOuts == outputs {
							return 0 // all SAT
						}
						Pdr_QueueClean(p)
						pCube = nil
						break // keep solving
					}
					if pars.FVerbose != 0 {
						progress(0)
					}
				}
			}
			if fRefined {
				break
			}
			if p.pTime4Outs != nil {
				timeSince := pdrNow(p) - clkOne
				if p.pTime4Outs[index] > timeSince {
					p.pTime4Outs[index] -= timeSince
				} else {
					p.pTime4Outs[index] = 0
				}
				if p.pTime4Outs[index] == 0 && vec.Vec_PtrEntry(p.vCexes, index) == nil { // undecided
					pars.NDropOuts++
					if pars.VOutMap != nil {
						vec.Vec_IntWriteEntry(pars.VOutMap, index, -1)
					}
					if pars.FNotVerbose == 0 {
						util.Abc_Print(1, "Timing out on output %*d in frame %d.\n", nOutDigits, index, iFrame)
					}
				}
				p.timeToStopOne = 0
			}
		}
		if pars.FUseAbs != 0 && p.vAbsFlops != nil && !fRefined {
			for i, used := range vec.Vec_IntArray(p.vAbsFlops) {
				if used != 0 && vec.Vec_IntEntry(p.vPrio, i)>>p.nPrioShift == 0 {
					vec.Vec_IntWriteEntry(p.vAbsFlops, i, 0)
				}
			}
		}
		if pars.FVerbose != 0 {
			progress(pdrBool(!fRefined))
		}
		if fRefined {
			continue
		}
		// open a new timeframe
		p.nQueLim = pars.NRestLimit
		if pCube != nil {
			panic("pdr: pending cube when opening a frame")
		}
		Pdr_ManSetPropertyOutput(p, iFrame)
		iFrame++
		Pdr_ManCreateSolver(p, iFrame)
		// push clauses into this timeframe
		RetValue = Pdr_ManPushClauses(p)
		if RetValue == -1 {
			if pars.FVerbose != 0 {
				progress(1)
			}
			if pars.FSilent == 0 {
				if p.timeToStop != 0 && pdrNow(p) > p.timeToStop {
					util.Abc_Print(1, "Reached timeout (%d seconds) in frame %d.\n", pars.NTimeOut, iFrame)
				} else {
					util.Abc_Print(1, "Reached conflict limit (%d) in frame %d.\n", pars.NConfLimit, iFrame)
				}
			}
			pars.IFrame = iFrame
			return -1
		}
		if RetValue != 0 {
			if pars.FVerbose != 0 {
				progress(1)
			}
			if pars.FSilent == 0 {
				Pdr_ManReportInvariant(p)
			}
			if pars.FSilent == 0 {
				Pdr_ManVerifyInvariant(p)
			}
			pars.IFrame = iFrame
			// count the number of UNSAT outputs
			pars.NProveOuts = outputs - pars.NFailOuts - pars.NDropOuts
			// convert previously 'unknown' into 'unsat'
			if pars.VOutMap != nil {
				for i := int32(0); i < outputs; i++ {
					if vec.Vec_IntEntry(pars.VOutMap, int(i)) == -2 { // unknown
						vec.Vec_IntWriteEntry(pars.VOutMap, int(i), 1) // unsat
						if pars.FUseBridge != 0 {
							util.Gia_ManToBridgeResult(os.Stdout, 1, nil, i)
						}
					}
				}
			}
			if pars.NProveOuts == outputs {
				return 1 // UNSAT
			}
			if pars.NFailOuts > 0 {
				return 0 // SAT
			}
			return -1
		}
		if pars.FVerbose != 0 {
			progress(0)
		}
		// check termination
		if pars.pFuncStop != nil && pars.pFuncStop(pars.RunId) != 0 {
			pars.IFrame = iFrame
			return -1
		}
		if p.timeToStop != 0 && pdrNow(p) > p.timeToStop {
			if pars.FVerbose != 0 {
				progress(1)
			}
			if pars.FSilent == 0 {
				util.Abc_Print(1, "Reached timeout (%d seconds) in frame %d.\n", pars.NTimeOut, iFrame)
			}
			pars.IFrame = iFrame
			return -1
		}
		if gapExpired() {
			if pars.FVerbose != 0 {
				progress(1)
			}
			if pars.FSilent == 0 {
				util.Abc_Print(1, "Reached gap timeout (%d seconds) in frame %d.\n", pars.NTimeOutGap, iFrame)
			}
			pars.IFrame = iFrame
			return -1
		}
		if pars.NFrameMax != 0 && iFrame >= pars.NFrameMax {
			if pars.FVerbose != 0 {
				progress(1)
			}
			if pars.FSilent == 0 {
				util.Abc_Print(1, "Reached limit on the number of timeframes (%d).\n", pars.NFrameMax)
			}
			pars.IFrame = iFrame
			return -1
		}
	}
}

// pdrReportStop prints why Pdr_ManCheckCube or Pdr_ManBlockCube gave up, in
// the order of Pdr_ManSolveInt; it returns true (after cleaning the queue)
// when only the current output's time ran out, so solving continues.
func pdrReportStop(p *Pdr_Man_t, iFrame int32) bool {
	pars := p.pPars
	now := pdrNow(p)
	switch {
	case p.timeToStop != 0 && now > p.timeToStop:
		util.Abc_Print(1, "Reached timeout (%d seconds) in frame %d.\n", pars.NTimeOut, iFrame)
	case pars.NTimeOutGap != 0 && pars.timeLastSolved != 0 && now > pars.timeLastSolved+int64(pars.NTimeOutGap)*int64(time.Second):
		util.Abc_Print(1, "Reached gap timeout (%d seconds) in frame %d.\n", pars.NTimeOutGap, iFrame)
	case p.timeToStopOne != 0 && now > p.timeToStopOne:
		Pdr_QueueClean(p)
		return true
	case pars.NConfLimit != 0:
		util.Abc_Print(1, "Reached conflict limit (%d) in frame %d.\n", pars.NConfLimit, iFrame)
	case pars.FVerbose != 0:
		util.Abc_Print(1, "Computation cancelled by the callback in frame %d.\n", iFrame)
	}
	return false
}

// pdrStoreCex records a failed output; a nil witness (not requested) is
// stored as pdrNoCounterexample, C's (Abc_Cex_t *)1.
func pdrStoreCex(p *Pdr_Man_t, index int, witness *util.Abc_Cex_t) {
	if vec.Vec_PtrEntry(p.vCexes, index) != nil {
		panic("pdr: output already has a counterexample")
	}
	var stored any = witness
	if witness == nil {
		stored = pdrNoCounterexample{}
	}
	vec.Vec_PtrWriteEntry(p.vCexes, index, stored)
}

// pdrCallOnFail runs pFuncOnFail, passing the witness only with fStoreCex.
func pdrCallOnFail(p *Pdr_Man_t, index int, witness *util.Abc_Cex_t) bool {
	if p.pPars.pFuncOnFail == nil {
		return false
	}
	if p.pPars.FStoreCex == 0 {
		witness = nil
	}
	return p.pPars.pFuncOnFail(int32(index), witness) != 0
}

// Distinguishes a failed output whose witness was not requested from an unsolved output.
type pdrNoCounterexample struct{}

func Pdr_ManSolve(graph *aig.Graph, pars *Pdr_Par_t) int32 {
	// ABC's updated justification assumes dynamic-CNF reverse mappings.
	if pars.FNewXSim != 0 && pars.FMonoCnf != 0 {
		return -1
	}
	started := util.Abc_Clock()
	outputs := graph.TruePoCount()
	if pars.NTimeOutOne != 0 && pars.FSolveAll == 0 {
		pars.NTimeOutOne = 0
	}
	if pars.NTimeOutOne != 0 && pars.NTimeOut == 0 {
		pars.NTimeOut = (pars.NTimeOutOne*outputs + 999) / 1000
	}
	if pars.FVerbose != 0 {
		fmt.Printf("VarMax = %d. FrameMax = %d. QueMax = %d. TimeMax = %d. MonoCNF = %s. SkipGen = %s. SolveAll = %s.\n", pars.NRecycle, pars.NFrameMax, pars.NRestLimit, pars.NTimeOut, pdrYesNo(pars.FMonoCnf), pdrYesNo(pars.FSkipGeneral), pdrYesNo(pars.FSolveAll))
	}
	graph.ClearSequentialCounterexample()
	p := Pdr_ManStart(graph, pars, nil)
	result := Pdr_ManSolveInt(p)
	if p.vCexes != nil {
		if graph.SequentialCounterexamples() != nil {
			panic("pdr: graph already has per-output counterexamples")
		}
		graph.SetSequentialCounterexamples(p.vCexes)
		p.vCexes = nil
	}
	if pars.FDumpInv != 0 || result == 1 {
		var reduce int32
		if result != 1 {
			reduce = 1
		}
		abcmain.Abc_FrameSetInv(vec.Vec_IntArray(Pdr_ManDeriveInfinityClauses(p, reduce)))
		if pars.FDumpInv != 0 {
			filename := pars.PInvFileName
			if filename == "" {
				filename = aig.Aig_FileNameGenericAppend(graph.Name, "_inv.pla")
			}
			var proved int32
			if result == 1 {
				proved = 1
			}
			Pdr_ManDumpClauses(p, filename, proved)
			fmt.Printf("Dumped inductive invariant in file \"%s\".\n", filename)
		}
	}
	p.tTotal += util.Abc_Clock() - started
	Pdr_ManStop(p)
	pars.IFrame--
	if pars.VOutMap != nil {
		for i := int32(0); i < outputs; i++ {
			if vec.Vec_IntEntry(pars.VOutMap, int(i)) == -2 {
				vec.Vec_IntWriteEntry(pars.VOutMap, int(i), -1)
			}
		}
	}
	if pars.FUseBridge != 0 {
		util.Gia_ManToBridgeAbort(os.Stdout, 7, []byte("timeout"))
	}
	return result
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

func pdrYesNo(flag int32) string {
	if flag != 0 {
		return "yes"
	}
	return "no"
}
