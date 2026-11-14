package if_

/**CFile****************************************************************

  FileName    [ifCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [The central part of the mapper.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: ifCore.c,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "if.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var s_MappingTime abctime

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManSetDefaultPars(pPars *If_Par_t) {
	*pPars = If_Par_t_{
		NLutSize:    -1,
		NCutsMax:    8,
		NFlowIters:  1,
		NAreaIters:  2,
		DelayTarget: -1,
		Epsilon:     0.005,
		FPreprocess: 1,
		FExpRed:     1,
		FEdge:       1,
	}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManPerformMapping(p *If_Man_t) int32 {
	p.pPars.fAreaOnly = p.pPars.FArea // temporary
	// create the CI cutsets
	If_ManSetupCiCutSets(p)
	// allocate memory for other cutsets
	If_ManSetupSetAll(p, If_ManCrossCut(p))
	// derive reverse top order
	p.vObjsRev = If_ManReverseOrder(p)
	return If_ManPerformMappingComb(p)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func If_ManPerformMappingComb(p *If_Man_t) int32 {
	clkTotal := Abc_Clock()

	// set arrival times and fanout estimates
	// inlining of macro If_ManForEachCi
	// inlining of macro Vec_PtrForEachEntry
	for i, entry := range Vec_PtrArray(p.vCis) {
		pObj := entry.(*If_Obj_t)
		arrival := float32(0)
		if i < len(p.pPars.pTimesArr) {
			arrival = p.pPars.pTimesArr[i]
		}
		If_ObjSetArrTime(pObj, arrival)
		pObj.EstRefs = 1
	}

	// delay oriented mapping
	if p.pPars.FPreprocess != 0 && p.pPars.FArea == 0 {
		// map for delay
		If_ManPerformMappingRound(p, p.pPars.NCutsMax, 0, 1, 1, "Delay")
		// map for delay second option
		p.pPars.FFancy = 1
		If_ManResetOriginalRefs(p)
		If_ManPerformMappingRound(p, p.pPars.NCutsMax, 0, 1, 0, "Delay-2")
		p.pPars.FFancy = 0
		// map for area
		p.pPars.FArea = 1
		If_ManResetOriginalRefs(p)
		If_ManPerformMappingRound(p, p.pPars.NCutsMax, 0, 1, 0, "Area")
		p.pPars.FArea = 0

	} else {
		If_ManPerformMappingRound(p, p.pPars.NCutsMax, 0, 0, 1, "Delay")
	}

	// try to improve area by expanding and reducing the cuts
	if p.pPars.FExpRed != 0 {
		If_ManImproveMapping(p)
	}

	// area flow oriented mapping
	for i := int32(0); i < p.pPars.NFlowIters; i++ {
		If_ManPerformMappingRound(p, p.pPars.NCutsMax, 1, 0, 0, "Flow")
		if p.pPars.FExpRed != 0 {
			If_ManImproveMapping(p)
		}
	}

	// area oriented mapping
	for i := int32(0); i < p.pPars.NAreaIters; i++ {
		If_ManPerformMappingRound(p, p.pPars.NCutsMax, 2, 0, 0, "Area")
		if p.pPars.FExpRed != 0 {
			If_ManImproveMapping(p)
		}
	}

	if p.pPars.FVerbose != 0 {
		//        Abc_Print( 1, "Total memory = %7.2f MB. Peak cut memory = %7.2f MB.  ",
		//            1.0 * (p->nObjBytes + 2*sizeof(void *)) * If_ManObjNum(p) / (1<<20),
		//            1.0 * p->nSetBytes * Mem_FixedReadMaxEntriesUsed(p->pMemSet) / (1<<20) );
		Abc_PrintTime(1, "Total time", Abc_Clock()-clkTotal)
	}
	//    Abc_Print( 1, "Cross cut memory = %d.\n", Mem_FixedReadMaxEntriesUsed(p->pMemSet) );
	s_MappingTime = Abc_Clock() - clkTotal
	//    Abc_Print( 1, "Special POs = %d.\n", If_ManCountSpecialPos(p) );

	/*
	   {
	       static char * pLastName = NULL;
	       FILE * pTable = fopen( "fpga/ucsb/stats.txt", "a+" );
	       if ( pLastName == NULL || strcmp(pLastName, p->pName) )
	       {
	           fprintf( pTable, "\n" );
	           fprintf( pTable, "%s ", p->pName );

	           fprintf( pTable, "%d ", If_ManCiNum(p) );
	           fprintf( pTable, "%d ", If_ManCoNum(p) );
	           fprintf( pTable, "%d ", If_ManAndNum(p) );

	           ABC_FREE( pLastName );
	           pLastName = Abc_UtilStrsav( p->pName );
	       }

	       fprintf( pTable, "%d ", (int)p->AreaGlo );
	       fprintf( pTable, "%d ", (int)p->RequiredGlo );
	       fclose( pTable );
	   }
	*/
	p.pPars.FinalDelay = p.RequiredGlo
	p.pPars.FinalArea = p.AreaGlo
	return 1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
