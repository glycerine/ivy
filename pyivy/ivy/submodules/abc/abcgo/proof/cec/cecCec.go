package cec

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [cecCec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Integrated combinatinal equivalence checker.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecCec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "proof/fra/fra.h"
//#include "aig/gia/giaAig.h"
//#include "misc/extra/extra.h"
//#include "sat/cnf/cnf.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Saves the input pattern with the given number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManTransformPattern(p *Gia_Man_t, iOut int32, pValues *int32) {
    var i int32;
    assert(p.pCexComb == nil, "p.pCexComb == nil");
    p.pCexComb = Abc_CexAlloc( 0, Gia_ManCiNum(p), 1 );
    p.pCexComb.iPo = iOut;
    for  i = 0; i < Gia_ManCiNum(p); i++  {
        if (pValues != nil) && pValues[i] {
            Abc_InfoSetBit( p.pCexComb.pData, i ); }
    }
}

/**Function*************************************************************

  Synopsis    [Interface to the old CEC engine]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManVerifyOld(pMiter *Gia_Man_t, fVerbose int32, piOutFail *int32, clkTotal abctime, fSilent int32) int32 {
//    extern int Fra_FraigCec( Aig_Man_t ** ppAig, int nConfLimit, int fVerbose );
    extern int Ssw_SecCexResimulate( Aig_Man_t * p, int * pModel, int * pnOutputs );
    var pTemp *Gia_Man_t  = Gia_ManTransformMiter( pMiter );
    var pMiterCec *Aig_Man_t  = Gia_ManToAig( pTemp, 0 );
    var RetValue, iOut, nOuts int32;
    if piOutFail != nil {
        *piOutFail = -1; }
    Gia_ManStop( pTemp );
    // run CEC on this miter
    RetValue = Fra_FraigCec( &pMiterCec, 10000000, fVerbose );
    // report the miter
    if RetValue == 1 {
        if !(fSilent != 0) {
            Abc_Print( 1, "Networks are equivalent.  " );
            Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
        }
    
    } else if RetValue == 0 {
        if !(fSilent != 0) {
            Abc_Print( 1, "Networks are NOT EQUIVALENT.  " );
            Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
        }
        if pMiterCec.pData == nil {
            Abc_Print( 1, "Counter-example is not available.\n" );
        } else {
            iOut = Ssw_SecCexResimulate( pMiterCec, (int *)pMiterCec.pData, &nOuts );
            if iOut == -1 {
                Abc_Print( 1, "Counter-example verification has failed.\n" );
            } else {
                if !(fSilent != 0) {
                    Abc_Print( 1, "Primary output %d has failed", iOut );
                    if nOuts-1 >= 0 {
                        Abc_Print( 1, ", along with other %d incorrect outputs", nOuts-1 ); }
                    Abc_Print( 1, ".\n" );
                }
                if piOutFail != nil {
                    *piOutFail = iOut; }
            }
            Cec_ManTransformPattern( pMiter, iOut, (int *)pMiterCec.pData );
        }
    
    } else if !(fSilent != 0) {
        Abc_Print( 1, "Networks are UNDECIDED.  " );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clkTotal );
    }
    fflush( os.Stdout );
    Aig_ManStop( pMiterCec );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManHandleSpecialCases(p *Gia_Man_t, pPars *Cec_ParCec_t) int32 {
    var pObj1, pObj2 *Gia_Obj_t;
    var pDri1, pDri2 *Gia_Obj_t;
    var i int32;
    var clk abctime  = Abc_Clock();
    Gia_ManSetPhase( p );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj1) = Gia_ManCo(p, i)); i++  {
        pObj2 = Gia_ManPo( p, ++i );
        // check if they different on all-0 pattern
        // (for example, when they have the same driver but complemented)
        if Gia_ObjPhase(pObj1) != Gia_ObjPhase(pObj2) {
            if !pPars.fSilent {
            Abc_Print( 1, "Networks are NOT EQUIVALENT. Output %d trivially differs (different phase).  ", i/2 );
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            pPars.iOutFail = i/2;
            Cec_ManTransformPattern( p, i/2, nil );
            return 0;
        }
        // get the drivers
        pDri1 = Gia_ObjFanin0(pObj1);
        pDri2 = Gia_ObjFanin0(pObj2);
        // drivers are different PIs
        if Gia_ObjIsPi(p, pDri1) && Gia_ObjIsPi(p, pDri2) && pDri1 != pDri2 {
            if !pPars.fSilent {
            Abc_Print( 1, "Networks are NOT EQUIVALENT. Output %d trivially differs (different PIs).  ", i/2 );
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            pPars.iOutFail = i/2;
            Cec_ManTransformPattern( p, i/2, nil );
            // if their compl attributes are the same - one should be complemented
            assert(Gia_ObjFaninC0(pObj1) == Gia_ObjFaninC0(pObj2), "Gia_ObjFaninC0(pObj1) == Gia_ObjFaninC0(pObj2)");
            Abc_InfoSetBit( p.pCexComb.pData, Gia_ObjCioId(pDri1) );
            return 0;
        }
        // one of the drivers is a PI; another is a constant 0
        if (Gia_ObjIsPi(p, pDri1) && Gia_ObjIsConst0(pDri2)) || 
             (Gia_ObjIsPi(p, pDri2) && Gia_ObjIsConst0(pDri1)) {
            if !pPars.fSilent {
            Abc_Print( 1, "Networks are NOT EQUIVALENT. Output %d trivially differs (PI vs. constant).  ", i/2 );
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            pPars.iOutFail = i/2;
            Cec_ManTransformPattern( p, i/2, nil );
            // the compl attributes are the same - the PI should be complemented
            assert(Gia_ObjFaninC0(pObj1) == Gia_ObjFaninC0(pObj2), "Gia_ObjFaninC0(pObj1) == Gia_ObjFaninC0(pObj2)");
            if Gia_ObjIsPi(p, pDri1) {
                Abc_InfoSetBit( p.pCexComb.pData, Gia_ObjCioId(pDri1) );
            } else {
                Abc_InfoSetBit( p.pCexComb.pData, Gia_ObjCioId(pDri2) ); }
            return 0;
        }
    }
    if Gia_ManAndNum(p) == 0 {
        if !pPars.fSilent {
        Abc_Print( 1, "Networks are equivalent.  " );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        return 1;
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Performs naive checking.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManVerifyNaive(p *Gia_Man_t, pPars *Cec_ParCec_t) int32 {
    var pCnf *Cnf_Dat_t  = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pSat *sat_solver  = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    var pObj0, pObj1 *Gia_Obj_t;
    var clkStart abctime  = Abc_Clock();
    var nPairs int32  = Gia_ManPoNum(p)/2;
    var nUnsats, nSats, nUndecs, nTrivs int32;
    var i int32;
    var iVar0 int32;
    var iVar1 int32;
    var pLits [2]int32;
    var status int32;
    var RetValue int32;
    var pProgress *ProgressBar  = Extra_ProgressBarStart( os.Stdout, nPairs );
    assert(Gia_ManPoNum(p) % 2 == 0, "Gia_ManPoNum(p) % 2 == 0");
    for  i = 0; i < nPairs; i++  {
        if (i & 0xFF) == 0 {
            Extra_ProgressBarUpdate( pProgress, i, nil ); }
        pObj0 = Gia_ManPo(p, 2*i);
        pObj1 = Gia_ManPo(p, 2*i+1);
        if Gia_ObjChild0(pObj0) == Gia_ObjChild0(pObj1) {
            nUnsats++;
            nTrivs++;
            continue;
        }
        if pPars.TimeLimit && (Abc_Clock() - clkStart)/int64(time.Second) >= pPars.TimeLimit {
            fmt.Printf("Timeout (%d sec) is reached.\n", pPars.TimeLimit);
            nUndecs = nPairs - nUnsats - nSats;
            break;
        }
        iVar0 = pCnf.pVarNums[ Gia_ObjId(p, pObj0) ];
        iVar1 = pCnf.pVarNums[ Gia_ObjId(p, pObj1) ];
        assert(iVar0 >= 0 && iVar1 >= 0, "iVar0 >= 0 && iVar1 >= 0");
        pLits[0] = Abc_Var2Lit( iVar0, 0 );
        pLits[1] = Abc_Var2Lit( iVar1, 0 );
        // check direct
        pLits[0] = lit_neg(pLits[0]);
        status = sat_solver_solve( pSat, pLits, pLits + 2, pPars.nBTLimit, 0, 0, 0 );
        if status == l_False {
            pLits[0] = lit_neg( pLits[0] );
            pLits[1] = lit_neg( pLits[1] );
            RetValue = sat_solver_addclause( pSat, pLits, pLits + 2 );
            assert((RetValue != 0), "RetValue");
        
        } else if status == l_True {
            fmt.Printf("Output %d is SAT.\n", i);
            nSats++;
            continue;
        
        } else {
            nUndecs++;
            continue;
        }
        // check inverse
        status = sat_solver_solve( pSat, pLits, pLits + 2, pPars.nBTLimit, 0, 0, 0 );
        if status == l_False {
            pLits[0] = lit_neg( pLits[0] );
            pLits[1] = lit_neg( pLits[1] );
            RetValue = sat_solver_addclause( pSat, pLits, pLits + 2 );
            assert((RetValue != 0), "RetValue");
        
        } else if status == l_True {
            fmt.Printf("Output %d is SAT.\n", i);
            nSats++;
            continue;
        
        } else {
            nUndecs++;
            continue;
        }
        nUnsats++;
    }
    Extra_ProgressBarStop( pProgress );
    fmt.Printf("UNSAT = %6d.  SAT = %6d.   UNDEC = %6d.  Trivial = %6d.  ", nUnsats, nSats, nUndecs, nTrivs);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    Cnf_DataFree( pCnf );
    sat_solver_delete( pSat );
    if nSats {
        return 0; }
    if nUndecs {
        return -1; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [New CEC engine.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManVerify(pInit *Gia_Man_t, pPars *Cec_ParCec_t) int32 {
    var fDumpUndecided int32;
    var ParsFra Cec_ParFra_t;
    var pParsFra *Cec_ParFra_t = &ParsFra;
    var p, pNew *Gia_Man_t;
    var RetValue int32;
    var clk abctime  = Abc_Clock();
    var clkTotal abctime  = Abc_Clock();
    // consider special cases:
    // 1) (SAT) a pair of POs have different value under all-0 pattern
    // 2) (SAT) a pair of POs has different PI/Const drivers
    // 3) (UNSAT) 1-2 do not hold and there is no nodes
    RetValue = Cec_ManHandleSpecialCases( pInit, pPars );
    if RetValue == 0 || RetValue == 1 {
        return RetValue; }
    // preprocess 
    p = Gia_ManDup( pInit );
    Gia_ManEquivFixOutputPairs( p );
    p = Gia_ManCleanup( pNew = p );
    Gia_ManStop( pNew );
    if pPars.fNaive {
        RetValue = Cec_ManVerifyNaive( p, pPars );
        Gia_ManStop( p );
        return RetValue;
    }
    if pInit.vSimsPi {
        p.vSimsPi = Vec_WrdDup(pInit.vSimsPi); 
        p.nSimWords = pInit.nSimWords;
    }
    // sweep for equivalences
    Cec_ManFraSetDefaultParams( pParsFra );
    pParsFra.nItersMax    = 1000;
    pParsFra.nBTLimit     = pPars.nBTLimit;
    pParsFra.TimeLimit    = pPars.TimeLimit;
    pParsFra.fVerbose     = pPars.fVerbose;
    pParsFra.fVeryVerbose = pPars.fVeryVerbose;
    pParsFra.fCheckMiter  = 1;
    pParsFra.fDualOut     = 1;
    pNew = Cec_ManSatSweeping( p, pParsFra, pPars.fSilent );
    pPars.iOutFail = pParsFra.iOutFail;
    // update
    pInit.pCexComb = p.pCexComb; p.pCexComb = nil;
    Gia_ManStop( p );
    p = pInit;
    // continue
    if pNew == nil {
        if p.pCexComb != nil {
            if p.pCexComb && !Gia_ManVerifyCex( p, p.pCexComb, 1 ) {
                Abc_Print( 1, "Counter-example simulation has failed.\n" ); }
            if !pPars.fSilent {
            Abc_Print( 1, "Networks are NOT EQUIVALENT.  " );
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            return 0;
        }
        p = Gia_ManDup( pInit );
        Gia_ManEquivFixOutputPairs( p );
        p = Gia_ManCleanup( pNew = p );
        Gia_ManStop( pNew );
        pNew = p;
    }
    if pPars.fVerbose {
        Abc_Print( 1, "Networks are UNDECIDED after the new CEC engine.  " );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    if fDumpUndecided != 0 {
        pNew.pReprs = nil;
        pNew.pNexts = nil;
        Gia_AigerWrite( pNew, "gia_cec_undecided.aig", 0, 0, 0 );
        Abc_Print( 1, "The result is written into file \"%s\".\n", "gia_cec_undecided.aig" );
    }
    if pPars.TimeLimit && (Abc_Clock() - clkTotal)/int64(time.Second) >= pPars.TimeLimit {
        Gia_ManStop( pNew );
        return -1;
    }
    // call other solver
    if pPars.fVerbose {
        Abc_Print( 1, "Calling the old CEC engine.\n" ); }
    fflush( os.Stdout );
    RetValue = Cec_ManVerifyOld( pNew, pPars.fVerbose, &pPars.iOutFail, clkTotal, pPars.fSilent );
    p.pCexComb = pNew.pCexComb; pNew.pCexComb = nil;
    if p.pCexComb && !Gia_ManVerifyCex( p, p.pCexComb, 1 ) {
        Abc_Print( 1, "Counter-example simulation has failed.\n" ); }
    Gia_ManStop( pNew );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Simple SAT run to check equivalence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManVerifySimple(p *Gia_Man_t) int32 {
    var ParsCec Cec_ParCec_t;
    var pPars *Cec_ParCec_t = &ParsCec;
    Cec_ManCecSetDefaultParams( pPars );
    pPars.fSilent = 1;
    assert(Gia_ManCoNum(p) == 2, "Gia_ManCoNum(p) == 2");
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    return Cec_ManVerify( p, pPars );
}

/**Function*************************************************************

  Synopsis    [New CEC engine applied to two circuits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManVerifyTwo(p0 *Gia_Man_t, p1 *Gia_Man_t, fVerbose int32) int32 {
    var ParsCec Cec_ParCec_t;
    var pPars *Cec_ParCec_t = &ParsCec;
    var pMiter *Gia_Man_t;
    var RetValue int32;
    Cec_ManCecSetDefaultParams( pPars );
    pPars.fVerbose = fVerbose;
    pMiter = Gia_ManMiter( p0, p1, 0, 1, 0, 0, pPars.fVerbose );
    if pMiter == nil {
        return -1; }
    RetValue = Cec_ManVerify( pMiter, pPars );
    p0.pCexComb = pMiter.pCexComb; pMiter.pCexComb = nil;
    Gia_ManStop( pMiter );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [New CEC engine applied to two circuits.]

  Description [Returns 1 if equivalent, 0 if counter-example, -1 if undecided.
  Counter-example is returned in the first manager as pAig0->pSeqModel.
  The format is given in Abc_Cex_t (file "abc\src\aig\gia\gia.h").]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManVerifyTwoAigs(pAig0 *Aig_Man_t, pAig1 *Aig_Man_t, fVerbose int32) int32 {
    var p0, p1, pTemp *Gia_Man_t;
    var RetValue int32;

    p0 = Gia_ManFromAig( pAig0 );
    p0 = Gia_ManCleanup( pTemp = p0 );
    Gia_ManStop( pTemp );

    p1 = Gia_ManFromAig( pAig1 );
    p1 = Gia_ManCleanup( pTemp = p1 );
    Gia_ManStop( pTemp );

    RetValue = Cec_ManVerifyTwo( p0, p1, fVerbose );
    pAig0.pSeqModel = p0.pCexComb; p0.pCexComb = nil;
    Gia_ManStop( p0 );
    Gia_ManStop( p1 );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Implementation of new signal correspodence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_LatchCorrespondence(pAig *Aig_Man_t, nConfs int32, fUseCSat int32) *Aig_Man_t {
    var pGia *Gia_Man_t;
    var CorPars Cec_ParCor_t;
    var pCorPars *Cec_ParCor_t = &CorPars;
    Cec_ManCorSetDefaultParams( pCorPars );
    pCorPars.fLatchCorr = 1;
    pCorPars.fUseCSat   = fUseCSat;
    pCorPars.nBTLimit   = nConfs;
    pGia = Gia_ManFromAigSimple( pAig );
    Cec_ManLSCorrespondenceClasses( pGia, pCorPars );
    Gia_ManReprToAigRepr( pAig, pGia );
    Gia_ManStop( pGia );
    return Aig_ManDupSimple( pAig );
}

/**Function*************************************************************

  Synopsis    [Implementation of new signal correspodence.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_SignalCorrespondence(pAig *Aig_Man_t, nConfs int32, fUseCSat int32) *Aig_Man_t {
    var pGia *Gia_Man_t;
    var CorPars Cec_ParCor_t;
    var pCorPars *Cec_ParCor_t = &CorPars;
    Cec_ManCorSetDefaultParams( pCorPars );
    pCorPars.fUseCSat  = fUseCSat;
    pCorPars.nBTLimit  = nConfs;
    pGia = Gia_ManFromAigSimple( pAig );
    Cec_ManLSCorrespondenceClasses( pGia, pCorPars );
    Gia_ManReprToAigRepr( pAig, pGia );
    Gia_ManStop( pGia );
    return Aig_ManDupSimple( pAig );
}

/**Function*************************************************************

  Synopsis    [Implementation of fraiging.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_FraigCombinational(pAig *Aig_Man_t, nConfs int32, fVerbose int32) *Aig_Man_t {
    var pGia *Gia_Man_t;
    var FraPars Cec_ParFra_t;
    var pFraPars *Cec_ParFra_t = &FraPars;
    Cec_ManFraSetDefaultParams( pFraPars );
    pFraPars.fSatSweeping = 1;
    pFraPars.nBTLimit  = nConfs;
    pFraPars.nItersMax = 20;
    pFraPars.fVerbose  = fVerbose;
    pGia = Gia_ManFromAigSimple( pAig );
    Cec_ManSatSweeping( pGia, pFraPars, 0 );
    Gia_ManReprToAigRepr( pAig, pGia );
    Gia_ManStop( pGia );
    return Aig_ManDupSimple( pAig );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

