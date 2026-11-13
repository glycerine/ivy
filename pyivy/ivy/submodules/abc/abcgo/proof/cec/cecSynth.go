package cec

import "fmt"

/**CFile****************************************************************

  FileName    [cecSynth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Partitioned sequential synthesis.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSynth.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Populate sequential synthesis parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SeqSynthesisSetDefaultParams(p *Cec_ParSeq_t) {
	*p = Cec_ParSeq_t{nBTLimit: 1000, nFrames: 1, nLevelMax: -1, fConsts: 1, fEquivs: 1, nMinDomSize: 100}
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SeqReadMinDomSize(p *Cec_ParSeq_t) int32 {
    return p.nMinDomSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SeqReadVerbose(p *Cec_ParSeq_t) int32 {
    return p.fVerbose;
}

/**Function*************************************************************

  Synopsis    [Computes partitioning of registers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Transfers the classes.]

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


/**Function*************************************************************

  Synopsis    [Normalizes equivalences.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Partitioned sequential synthesis.]

  Description [Returns AIG annotated with equivalence classes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SequentialSynthesisPart(p *Gia_Man_t, pPars *Cec_ParSeq_t) int32 {
    var fPrintParts int32;
    var Buffer string // was [100]byte in C

    var pTemp *Gia_Man_t;
    var vParts *Vec_Ptr_t  = (Vec_Ptr_t *)p.vClockDoms;
    var vPart *Vec_Int_t;
    var pMapBack, pReprs *int32;
    var i, nCountPis, nCountRegs int32;
    var nClasses int32;
    var clk abctime  = Abc_Clock();

    // save parameters
    if fPrintParts != 0 {
        // print partitions
        Abc_Print( 1, "The following clock domains are used:\n" );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
            pTemp = Gia_ManRegCreatePart( p, vPart, &nCountPis, &nCountRegs, nil );
            Buffer = fmt.Sprintf("part%03d.aig", i);
            Gia_AigerWrite( pTemp, Buffer, 0, 0, 0 );
            Abc_Print( 1, "part%03d.aig : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d.\n", 
                i, Vec_IntSize(vPart), Gia_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Gia_ManAndNum(pTemp) );
            Gia_ManStop( pTemp );
        }
    }

    // perform sequential synthesis for clock domains
    // inlining of macro ABC_FALLOC
    pReprs = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p))));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vParts)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vParts, i)), 1); i++  {
        pTemp = Gia_ManRegCreatePart( p, vPart, &nCountPis, &nCountRegs, &pMapBack );
        if nCountPis > 0 {
            if pPars.fUseScorr {
                var CorPars Cec_ParCor_t;
                var pCorPars *Cec_ParCor_t = &CorPars;
                Cec_ManCorSetDefaultParams( pCorPars );
                pCorPars.nBTLimit   = pPars.nBTLimit;
                pCorPars.nLevelMax  = pPars.nLevelMax;
                pCorPars.fVerbose   = pPars.fVeryVerbose;
                pCorPars.fUseCSat   = 1;
                Cec_ManLSCorrespondenceClasses( pTemp, pCorPars );
            
            } else if pPars.fUseLcorr {
                var CorPars Cec_ParCor_t;
                var pCorPars *Cec_ParCor_t = &CorPars;
                Cec_ManCorSetDefaultParams( pCorPars );
                pCorPars.fLatchCorr = 1;
                pCorPars.nBTLimit   = pPars.nBTLimit;
                pCorPars.fVerbose   = pPars.fVeryVerbose;
                pCorPars.fUseCSat   = 1;
                Cec_ManLSCorrespondenceClasses( pTemp, pCorPars );
            
            } else {
//                pNew = Gia_ManSeqStructSweep( pTemp, pPars->fConsts, pPars->fEquivs, pPars->fVerbose );
//                Gia_ManStop( pNew );
                Gia_ManSeqCleanupClasses( pTemp, pPars.fConsts, pPars.fEquivs, pPars.fVerbose );
            }
//Abc_Print( 1, "Part equivalences = %d.\n", Gia_ManEquivCountLitsAll(pTemp) );
            nClasses = Gia_TransferMappedClasses( pTemp, pMapBack, pReprs );
            if pPars.fVerbose {
                Abc_Print( 1, "%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. Cl = %5d.\n", 
                    i, Vec_IntSize(vPart), Gia_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Gia_ManAndNum(pTemp), nClasses );
            }
        }
        Gia_ManStop( pTemp );
        pMapBack = nil;
    }

    // generate resulting equivalences
    Gia_ManNormalizeEquivalences( p, pReprs );
//Abc_Print( 1, "Total equivalences = %d.\n", Gia_ManEquivCountLitsAll(p) );
    pReprs = nil;
    if pPars.fVerbose {
        Abc_PrintTime( 1, "Total time", Abc_Clock() - clk );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTestOnePair_rec(pSat *sat_solver, p *Gia_Man_t, iObj int32, vSatVar *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t; 
    var iVar, iVar0, iVar1 int32;
    if Vec_IntEntry(vSatVar, iObj) >= 0 {
        return Vec_IntEntry(vSatVar, iObj); }
    iVar = sat_solver_addvar(pSat);
    Vec_IntWriteEntry( vSatVar, iObj, iVar );
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjIsAnd(pObj) {
        iVar0 = Gia_ManTestOnePair_rec( pSat, p, Gia_ObjFaninId0(pObj, iObj), vSatVar );
        iVar1 = Gia_ManTestOnePair_rec( pSat, p, Gia_ObjFaninId1(pObj, iObj), vSatVar );
        sat_solver_add_and( pSat, iVar, iVar0, iVar1, Gia_ObjFaninC0(pObj), Gia_ObjFaninC1(pObj), 0 );
    }
    return iVar;
}
func Gia_ManTestOnePair(p *Gia_Man_t, iObj1 int32, iObj2 int32, fPhase int32) int32 {
    var Lits [2]int32 = [2]int32{1};
    var status int32;
    var pSat *sat_solver  = sat_solver_new();
    var vSatVar *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    Vec_IntWriteEntry( vSatVar, 0, sat_solver_addvar(pSat) );
    sat_solver_addclause( pSat, Lits, Lits + 1 );
    Gia_ManTestOnePair_rec( pSat, p, iObj1, vSatVar );
    Gia_ManTestOnePair_rec( pSat, p, iObj2, vSatVar );
    Lits[0] = Abc_Var2Lit( Vec_IntEntry(vSatVar, iObj1), 1 );
    Lits[1] = Abc_Var2Lit( Vec_IntEntry(vSatVar, iObj2), fPhase );
    status = sat_solver_solve( pSat, Lits, Lits + 2, 0, 0, 0, 0 );
    if status == l_False {
        Lits[0] = Abc_LitNot( Lits[0] );
        Lits[1] = Abc_LitNot( Lits[1] );
        status = sat_solver_solve( pSat, Lits, Lits + 2, 0, 0, 0, 0 );
    }
    Vec_IntFree( vSatVar );
    sat_solver_delete( pSat );
    return status;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

