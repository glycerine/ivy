package gia

import "fmt"

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Slv_ManForEachObj( p, iObj )                                  \
//     for ( iObj = 1; iObj < Slv_ManObjNum(p); iObj++ )

// Original macro:
// #define Slv_ObjForEachFanout( p, iObj, iFanLit )                      \
//     for ( iFanLit = Slv_ObjFanout0(p, iObj); iFanLit; iFanLit = Slv_ObjNextFanout(p, iFanLit) )

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ManCofPisVars(p *Gia_Man_t, nVars int32) *Gia_Man_t {
	if nVars < 0 || nVars > Gia_ManPiNum(p) || nVars >= 31 {
		panic("gia: invalid cofactor variable count")
	}
	out := Gia_ManStart(1000)
	out.pName = p.pName
	out.PSpec = p.PSpec
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ManAppendCi(out)
	}
	Gia_ManHashStart(out)
	for assignment := int32(0); assignment < 1<<uint(nVars); assignment++ {
		for _, obj := range p.pObjs {
			obj.Value = ^uint32(0)
		}
		p.pObjs[0].Value = 0
		for i := int32(0); i < Gia_ManPiNum(p); i++ {
			obj := Gia_ManCi(p, i)
			if i < nVars {
				obj.Value = uint32(assignment>>uint(i)) & 1
			} else {
				obj.Value = uint32(Gia_ObjId(out, Gia_ManCi(out, i)) << 1)
			}
		}
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				obj.Value = uint32(Gia_ManHashAnd(out, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
			}
		}
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			Gia_ManAppendCo(out, Gia_ObjFanin0Copy(Gia_ManCo(p, i)))
		}
	}
	Gia_ManHashStop(out)
	cleaned := Gia_ManCleanup(out)
	Gia_ManStop(out)
	return cleaned
}

// Gia_ManStructExperiment groups output cones using the CEC equivalence engine.
// The higher-level CEC dependency remains to be integrated without a package cycle.
func Gia_ManStructExperiment(p *Gia_Man_t) {
	graphs := Vec_PtrAlloc(100)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		output := Gia_ManCo(p, i)
		fanin := Gia_ObjFaninId0(output, output.id)
		candidate := Gia_ManDupAndCones(p, []int32{fanin}, 1, 1)
		match := 0
		for ; match < Vec_PtrSize(graphs); match++ {
			existing := Vec_PtrEntry(graphs, match).(*Gia_Man_t)
			if Gia_ManCiNum(candidate) == Gia_ManCiNum(existing) && Cec_ManVerifyTwo(candidate, existing, 0) == 1 {
				values := Vec_PtrArray(graphs)
				values[0], values[match] = values[match], values[0]
				break
			}
			candidate.PCexComb = nil
		}
		fmt.Printf("\nOut %6d : ", i)
		if match == Vec_PtrSize(graphs) {
			fmt.Print("Equiv to none    ")
		} else {
			fmt.Printf("Equiv to %6d  ", match)
		}
		Gia_ManPrintStats(candidate, nil)
		if match == Vec_PtrSize(graphs) {
			Vec_PtrPush(graphs, candidate)
		} else {
			Gia_ManStop(candidate)
		}
	}
	fmt.Printf("\nComputed %d classes.\n\n", Vec_PtrSize(graphs))
	for _, value := range Vec_PtrArray(graphs) {
		Gia_ManStop(value.(*Gia_Man_t))
	}
	Vec_PtrFree(graphs)
}
