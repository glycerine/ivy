package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
)

/**CFile****************************************************************

  FileName    [giaSupMin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Support minimization for AIGs with don't-cares.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSupMin.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// decomposition manager

type Gia_ManSup_t struct {
	nVarsMax    int32      // the max number of variables
	nWordsMax   int32      // the max number of words
	vTruthVars  *Vec_Ptr_t // elementary truth tables
	vTruthNodes *Vec_Ptr_t // internal truth tables
	// current problem
	pGia      *Gia_Man_t
	iData     int32
	iCare     int32
	vConeCare *Vec_Int_t
	vConeData *Vec_Int_t
	pTruthIn  []uint32
	pTruthOut []uint32
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Starts Decmetry manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ManSupStart(nVarsMax int32) *Gia_ManSup_t {
	var p *Gia_ManSup_t
	if nVarsMax < 0 || nVarsMax > 20 {
		panic("gia: support variable count out of range")
	}
	p = new(Gia_ManSup_t)
	p.nVarsMax = nVarsMax
	p.nWordsMax = kit.Kit_TruthWordNum(p.nVarsMax)
	p.vTruthVars = Vec_PtrAllocTruthTables(int(p.nVarsMax))
	p.vTruthNodes = Vec_PtrAllocSimInfo(512, int(p.nWordsMax))
	p.vConeCare = Vec_IntAlloc(512)
	p.vConeData = Vec_IntAlloc(512)
	p.pTruthIn = make([]uint32, p.nWordsMax)
	p.pTruthOut = make([]uint32, p.nWordsMax)
	return p
}

/*
*Function*************************************************************

	Synopsis    [Stops Decmetry manager.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ManSupStop(p *Gia_ManSup_t) {
	p.pTruthIn = nil
	p.pTruthOut = nil
	Vec_IntFreeP(&p.vConeCare)
	Vec_IntFreeP(&p.vConeData)
	Vec_PtrFreeP(&p.vTruthVars)
	Vec_PtrFreeP(&p.vTruthNodes)
	p = nil
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ManSupExperimentOne(p *Gia_ManSup_t, pData *Gia_Obj_t, pCare *Gia_Obj_t) {
	var iData int32 = Gia_ObjId(p.pGia, Gia_Regular(pData))
	var iCare int32 = Gia_ObjId(p.pGia, Gia_Regular(pCare))
	if Gia_ObjIsAnd(Gia_Regular(pCare)) == 0 {
		fmt.Printf("Enable is not an AND.\n")
		return
	}
	fmt.Printf("DataSupp = %6d. DataCone = %6d.   CareSupp = %6d. CareCone = %6d.",
		Gia_ManSuppSize(p.pGia, []int32{iData}, 1),
		Gia_ManConeSize(p.pGia, []int32{iData}, 1),
		Gia_ManSuppSize(p.pGia, []int32{iCare}, 1),
		Gia_ManConeSize(p.pGia, []int32{iCare}, 1))
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Gia_ManSupExperiment(pGia *Gia_Man_t, vPairs *Vec_Int_t) {
	var p *Gia_ManSup_t
	var pData, pCare *Gia_Obj_t
	var i int
	p = Gia_ManSupStart(16)
	p.pGia = pGia
	if Vec_IntSize(vPairs)%2 != 0 {
		panic("gia: expected pairs of output indices")
	}
	for i = 0; i < Vec_IntSize(vPairs)/2; i++ {
		fmt.Printf("%6d : ", i)
		pData = Gia_ManPo(pGia, Vec_IntEntry(vPairs, 2*i+0))
		pCare = Gia_ManPo(pGia, Vec_IntEntry(vPairs, 2*i+1))
		Gia_ManSupExperimentOne(p, Gia_ObjChild0(pData), Gia_ObjChild0(pCare))
	}
	Gia_ManSupStop(p)
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
