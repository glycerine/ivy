package pdr

import (
	"fmt"
	"slices"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

/**CFile****************************************************************

  FileName    [pdrIncr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [PDR with incremental solving.]

  Author      [Yen-Sheng Ho, Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Feb. 17, 2017.]

  Revision    [$Id: pdrIncr.c$]

***********************************************************************/

//#include "pdrInt.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );

// extern int Pdr_ManBlockCube( Pdr_Man_t * p, Pdr_Set_t * pCube );

// extern int Pdr_ManPushClauses( Pdr_Man_t * p );

// extern Pdr_Set_t * Pdr_ManReduceClause( Pdr_Man_t * p, int k, Pdr_Set_t * pCube );

// extern int Gia_ManToBridgeAbort( FILE * pFile, int Size, unsigned char * pBuffer );

// extern int Gia_ManToBridgeResult( FILE * pFile, int Result, Abc_Cex_t * pCex, int iPoProved );

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func IPdr_ManPushClausesK(p *Pdr_Man_t, k int32) *Vec_Ptr_t {
	if vec.Vec_VecSize(p.vClauses) != int(k)+1 {
		panic("pdr: unexpected final frame")
	}
	current := vec.Vec_VecEntry(p.vClauses, int(k))
	vec.Vec_PtrSort(current, func(a, b *any) int {
		left, right := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
		return int(Pdr_SetCompare(&left, &right))
	})
	next := vec.Vec_PtrAlloc(100)
	remove := func(frame *Vec_Ptr_t, index int) {
		Pdr_SetDeref(vec.Vec_PtrEntry(frame, index).(*Pdr_Set_t))
		vec.Vec_PtrWriteEntry(frame, index, vec.Vec_PtrEntryLast(frame))
		vec.Vec_PtrPop(frame)
	}
	for j := 0; j < vec.Vec_PtrSize(current); {
		cube := vec.Vec_PtrEntry(current, j).(*Pdr_Set_t)
		for m := j + 1; m < vec.Vec_PtrSize(current); {
			if Pdr_SetContains(vec.Vec_PtrEntry(current, m).(*Pdr_Set_t), cube) == 0 {
				m++
				continue
			}
			remove(current, m)
		}
		result := Pdr_ManCheckCube(p, k, cube, nil, 0, 0, 1)
		if result == -1 {
			panic("pdr: incremental clause push returned unknown")
		}
		if result == 0 {
			j++
			continue
		}
		if reduced := Pdr_ManReduceClause(p, k, cube); reduced != nil {
			Pdr_SetDeref(cube)
			cube = reduced
		}
		for i := 0; i < vec.Vec_PtrSize(next); {
			if Pdr_SetContains(vec.Vec_PtrEntry(next, i).(*Pdr_Set_t), cube) == 0 {
				i++
				continue
			}
			remove(next, i)
		}
		vec.Vec_PtrPush(next, cube)
		vec.Vec_PtrWriteEntry(current, j, vec.Vec_PtrEntryLast(current))
		vec.Vec_PtrPop(current)
	}
	return next
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func IPdr_ManPrintClauses(clauses *Vec_Vec_t, start, nRegs int32) {
	var count int
	for frame := int(start); frame < vec.Vec_VecSize(clauses); frame++ {
		cubes := vec.Vec_VecEntry(clauses, frame)
		vec.Vec_PtrSort(cubes, func(a, b *any) int {
			left, right := (*a).(*Pdr_Set_t), (*b).(*Pdr_Set_t)
			return int(Pdr_SetCompare(&left, &right))
		})
		for _, item := range vec.Vec_PtrArray(cubes) {
			fmt.Printf("Frame[%4d]Cube[%4d] = ", frame, count)
			ZPdr_SetPrint(item.(*Pdr_Set_t))
			fmt.Println()
			count++
		}
	}
}

/*
*Function*************************************************************

	Synopsis    [ Check if each cube c_k in frame k satisfies the query
	              R_{k-1} && T && !c_k && c_k' (must be UNSAT).
	              Return True if all cubes pass the check. ]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func IPdr_ManCheckClauses(p *Pdr_Man_t) int32 {
	for frame := 1; frame < len(p.vSolvers); frame++ {
		for i, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
			result := Pdr_ManCheckCube(p, int32(frame-1), item.(*Pdr_Set_t), nil, 0, 0, 1)
			if result == 0 {
				fmt.Printf("Cube[%d][%d] not inductive!\n", frame, i)
			}
			if result == -1 {
				return -1
			}
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
func IPdr_ManSaveClauses(p *Pdr_Man_t, dropLast int32) *Vec_Vec_t {
	n := vec.Vec_VecSize(p.vClauses)
	if n == 1 || (n == 2 && dropLast != 0) {
		return nil
	}
	if dropLast != 0 {
		n--
	}
	saved := vec.Vec_VecStart(n)
	for frame := 0; frame < n; frame++ {
		for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
			vec.Vec_VecPush(saved, frame, Pdr_SetDup(item.(*Pdr_Set_t)))
		}
	}
	return saved
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func IPdr_ManSetSolver(p *Pdr_Man_t, k, property int32) *sat_solver {
	if len(p.vSolvers) != int(k) || vec.Vec_IntSize(p.vActVars) != int(k) {
		panic("pdr: inconsistent solver frame")
	}
	initial := int32(0)
	if k == 0 {
		initial = 1
	}
	solver := Pdr_ManNewSolver(bsat.NewSolverSeed(float64(p.pPars.nRandomSeed)), p, k, initial)
	p.vSolvers = append(p.vSolvers, solver)
	vec.Vec_IntPush(p.vActVars, 0)
	if property != 0 {
		Pdr_ManSetPropertyOutput(p, k)
	}
	if k != 0 {
		for frame := int(k); frame < vec.Vec_VecSize(p.vClauses); frame++ {
			for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
				Pdr_ManSolverAddClause(p, k, item.(*Pdr_Set_t))
			}
		}
	}
	return solver
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func IPdr_ManRebuildClauses(p *Pdr_Man_t, clauses *Vec_Vec_t) int32 {
	if clauses == nil {
		return -1
	}
	if vec.Vec_VecSize(clauses) < 2 || vec.Vec_VecSize(p.vClauses) != 0 {
		panic("pdr: invalid saved frames")
	}
	vec.Vec_VecExpand(p.vClauses, 1)
	IPdr_ManSetSolver(p, 0, 1)
	total := 0
	for frame := 1; frame < vec.Vec_VecSize(clauses); frame++ {
		for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(clauses, frame)) {
			cube := item.(*Pdr_Set_t)
			total++
			result := Pdr_ManCheckCube(p, 0, cube, nil, 0, 0, 1)
			vec.Vec_IntWriteEntry(p.vActVars, 0, 0)
			if result == -1 {
				panic("pdr: clause rebuilding returned unknown")
			}
			if result == 0 {
				fmt.Printf("Cube[%d] cannot be pushed from R0 to R1.\n", frame)
				Pdr_SetDeref(cube)
				continue
			}
			vec.Vec_VecPush(p.vClauses, 1, cube)
		}
	}
	fmt.Printf("RebuildClauses: %d out of %d cubes reused in R1.\n", vec.Vec_PtrSize(vec.Vec_VecEntry(p.vClauses, 1)), total)
	IPdr_ManSetSolver(p, 1, 0)
	vec.Vec_VecFree(clauses)
	return 0
}

func IPdr_ManRestoreAbsFlops(p *Pdr_Man_t) int32 {
	for frame := 0; frame < vec.Vec_VecSize(p.vClauses); frame++ {
		for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
			cube := item.(*Pdr_Set_t)
			for _, lit := range cube.Lits[:cube.nLits] {
				vec.Vec_IntWriteEntry(p.vAbsFlops, int(lit>>1), 1)
			}
		}
	}
	return 0
}

func IPdr_ManRestoreClauses(p *Pdr_Man_t, clauses *Vec_Vec_t, mapping *Vec_Int_t) int32 {
	if clauses == nil {
		panic("pdr: nil saved clauses")
	}
	vec.Vec_VecFree(p.vClauses)
	p.vClauses = clauses
	if mapping != nil {
		for frame := 0; frame < vec.Vec_VecSize(clauses); frame++ {
			for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(clauses, frame)) {
				cube := item.(*Pdr_Set_t)
				for i := int32(0); i < cube.nLits; i++ {
					lit := cube.Lits[i]
					cube.Lits[i] = 2*vec.Vec_IntEntry(mapping, int(lit>>1)) + (lit & 1)
				}
				slices.Sort(cube.Lits[:cube.nLits])
				cube.Sign = 0
				for _, lit := range cube.Lits[:cube.nLits] {
					cube.Sign |= uint64(1) << uint(lit%63)
				}
			}
		}
	}
	for frame := 0; frame < vec.Vec_VecSize(clauses); frame++ {
		property := int32(0)
		if frame < vec.Vec_VecSize(clauses)-1 {
			property = 1
		}
		IPdr_ManSetSolver(p, int32(frame), property)
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
func IPdr_ManCheckCombUnsat(p *Pdr_Man_t) int32 {
	Pdr_ManCreateSolver(p, 0)
	Pdr_ManCreateSolver(p, 1)
	p.nFrames, p.iUseFrame = 1, 1
	return Pdr_ManCheckCube(p, 1, nil, nil, p.pPars.nConfLimit, 0, 1)
}

func IPdr_ManReduceClauses(p *Pdr_Man_t, clauses *Vec_Vec_t) int32 {
	Pdr_ManCreateSolver(p, 0)
	Pdr_ManCreateSolver(p, 1)
	p.nFrames, p.iUseFrame = 1, 1
	last := vec.Vec_VecEntry(clauses, vec.Vec_VecSize(clauses)-1)
	return IPdr_ManCheckCubeReduce(p, last, nil, p.pPars.nConfLimit)
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
