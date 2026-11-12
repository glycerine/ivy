package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcChain.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Implementation of chain BMC.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcChain.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "aig/gia/giaAig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Find the first failure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_ChainFailOneOutput(p *Gia_Man_t, nFrameMax int32, nConfMax int32, fVerbose int32, fVeryVerbose int32) *Abc_Cex_t {
    var RetValue int32;
    var pCex *Abc_Cex_t  = nil;
    var pAig *Aig_Man_t  = Gia_ManToAigSimple( p );
    var Pars Saig_ParBmc_t;
    var pPars *Saig_ParBmc_t = &Pars;
    Saig_ParBmcSetDefaultParams( pPars );
    pPars.nFramesMax = nFrameMax;
    pPars.nConfLimit = nConfMax;
    pPars.fVerbose   = fVeryVerbose;
    RetValue = Saig_ManBmcScalable( pAig, pPars );
    if RetValue == 0 { // SAT
    
        pCex = pAig.pSeqModel;
pAig.pSeqModel = nil;
        if fVeryVerbose != 0 {
            Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.\n", pCex.iPo, p.pName, pCex.iFrame ); }
    
    } else if fVeryVerbose != 0 {
        Abc_Print( 1, "No output asserted in %d frames. Resource limit reached.\n", pPars.iFrame+2 ); }
    Aig_ManStop( pAig );
    return pCex;
}

/**Function*************************************************************

  Synopsis    [Move GIA into the failing state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupWithInit(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
            pObj.Value = Abc_LitNotCond( pObj.Value, pObj.fMark0 );
        
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
            pObj.Value = Abc_LitNotCond( pObj.Value, pObj.fMark0 );
            pObj.Value = Gia_ManAppendCo( pNew, pObj.Value );
        }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}
func Gia_ManVerifyCexAndMove(pGia *Gia_Man_t, p *Abc_Cex_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var RetValue, i, k, iBit int32;
    Gia_ManCleanMark0(pGia);
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(pGia)) && ((pObj) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+i)); i++  {
        pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit++);
    }
    for  i = 0; i <= p.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(pGia)) && ((pObj) = Gia_ManCi(pGia, k)); k++  {
            pObj.fMark0 = Abc_InfoHasBit(p.pData, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, k)); k++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        if i == p.iFrame {
            break; }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(pGia)) && ((pObjRi) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+k)) && ((pObjRo) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+k)); k++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    RetValue = Gia_ManPo(pGia, p.iPo).fMark0;
    assert((RetValue != 0), "RetValue");
    // set PI/PO values to zero and transfer RO values to RI
    // inlining of macro Gia_ManForEachPi
    for  k = 0; (k < Gia_ManPiNum(pGia)) && ((pObj) = Gia_ManCi(pGia, k)); k++  {
        pObj.fMark0 = 0;
    }
    // inlining of macro Gia_ManForEachPo
    for  k = 0; (k < Gia_ManPoNum(pGia)) && ((pObj) = Gia_ManCo(pGia, k)); k++  {
        pObj.fMark0 = 0;
    }
    // inlining of macro Gia_ManForEachRiRo
    for  k = 0; (k < Gia_ManRegNum(pGia)) && ((pObjRi) = Gia_ManCo(pGia, Gia_ManPoNum(pGia)+k)) && ((pObjRo) = Gia_ManCi(pGia, Gia_ManPiNum(pGia)+k)); k++  {
        pObjRi.fMark0 = pObjRo.fMark0;
    }
    // duplicate assuming CI/CO marks are set correctly
    pNew = Gia_ManDupWithInit( pGia );
    Gia_ManCleanMark0(pGia);
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Find what outputs fail in this state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupPosAndPropagateInit(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;  var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsPi(p, pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = 0;
        } else if Gia_ObjIsPo(p, pObj) {
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    assert(Gia_ManPiNum(p) == Gia_ManPiNum(pNew), "Gia_ManPiNum(p) == Gia_ManPiNum(pNew)");
    assert(Gia_ManPoNum(p) == Gia_ManPoNum(pNew), "Gia_ManPoNum(p) == Gia_ManPoNum(pNew)");
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManDeriveSatSolver(p *Gia_Man_t, vSatIds *Vec_Int_t) *sat_solver {
    var pSat *sat_solver;
    var pAig *Aig_Man_t  = Gia_ManToAigSimple( p );
//    Cnf_Dat_t * pCnf = (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
    var pCnf *Cnf_Dat_t  = Cnf_Derive( pAig, Aig_ManCoNum(pAig) );
    // save SAT vars for the primary inputs
    if vSatIds != nil {
        var pObj *Aig_Obj_t; var i int32;
        Vec_IntClear( vSatIds );
        // inlining of macro Aig_ManForEachCi
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            Vec_IntPush( vSatIds, pCnf.pVarNums[ Aig_ObjId(pObj) ] );
        }
        assert(Vec_IntSize(vSatIds) == Gia_ManPiNum(p), "Vec_IntSize(vSatIds) == Gia_ManPiNum(p)");
    }
    Aig_ManStop( pAig );
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    Cnf_DataFree( pCnf );
    assert(p.nRegs == 0, "p.nRegs == 0");
    return pSat;
}
func Bmc_ChainFindFailedOutputs(p *Gia_Man_t, vCexes *Vec_Ptr_t) *Vec_Int_t {
    var vOutputs *Vec_Int_t;
    var vSatIds *Vec_Int_t;
    var pInit *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var pSat *sat_solver;
    var i, j, Lit, status int32;
    // derive output logic cones
    pInit = Gia_ManDupPosAndPropagateInit( p );
    // derive SAT solver and test
    vSatIds = Vec_IntAlloc( Gia_ManPiNum(p) );
    pSat = Gia_ManDeriveSatSolver( pInit, vSatIds );
    vOutputs = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(pInit)) && ((pObj) = Gia_ManCo(pInit, i)); i++  {
        if Gia_ObjFaninLit0p(pInit, pObj) == 0 {
            continue; }
        Lit = Abc_Var2Lit( i+1, 0 );
        status = sat_solver_solve( pSat, &Lit, &Lit + 1, 0, 0, 0, 0 );
        if status == l_True {
            // save the index of solved output
            Vec_IntPush( vOutputs, i );
            // create CEX for this output
            if vCexes != nil {
                var pCex *Abc_Cex_t  = Abc_CexAlloc( Gia_ManRegNum(p), Gia_ManPiNum(p), 1 );
                pCex.iFrame = 0;
                pCex.iPo = i;
                for  j = 0; j < Gia_ManPiNum(p); j++  {
                    if sat_solver_var_value( pSat, Vec_IntEntry(vSatIds, j) ) {
                        Abc_InfoSetBit( pCex.pData, Gia_ManRegNum(p) + j ); }
                }
                Vec_PtrPush( vCexes, pCex );
            }
        }
    }
    Gia_ManStop( pInit );
    sat_solver_delete( pSat );
    Vec_IntFree( vSatIds );
    return vOutputs;
}

/**Function*************************************************************

  Synopsis    [Cleanup AIG by removing COs replacing failed out by const0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjMakePoConst0(p *Gia_Man_t, pObj *Gia_Obj_t) { 
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    pObj.iDiff0 = Gia_ObjId( p, pObj );
    pObj.fCompl0 = 0;
}
func Gia_ManCountNonConst0(p *Gia_Man_t) int32 { 
    var pObj *Gia_Obj_t;
    var i, Count int32;
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Count += (Gia_ObjFaninLit0p(p, pObj) != 0);
    }
    return Count;
}
func Bmc_ChainCleanup(p *Gia_Man_t, vOutputs *Vec_Int_t) *Gia_Man_t {
    var i, iOut int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOutputs)) && (((iOut) = Vec_IntEntry(vOutputs, i)), 1); i++  {
        var pObj *Gia_Obj_t  = Gia_ManPo( p, iOut );
        assert(Gia_ObjFaninLit0p(p, pObj) != 0, "Gia_ObjFaninLit0p(p, pObj) != 0");
        Gia_ObjMakePoConst0( p, pObj );
        assert(Gia_ObjFaninLit0p(p, pObj) == 0, "Gia_ObjFaninLit0p(p, pObj) == 0");
    }
    return Gia_ManCleanup( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_ChainTest(p *Gia_Man_t, nFrameMax int32, nConfMax int32, fVerbose int32, fVeryVerbose int32, pvCexes **Vec_Ptr_t) int32 {
    var Iter int32;
    var IterMax int32 = 10000;
    var pTemp *Gia_Man_t;
    var pNew *Gia_Man_t = Gia_ManDup(p);
    var pCex *Abc_Cex_t  = nil;
    var vOutputs *Vec_Int_t;
    var clk2 abctime;
    var clk abctime = Abc_Clock(); 
    var clkBmc abctime;
    var clkMov abctime;
    var clkSat abctime;
    var clkCln abctime;
    var clkSwp abctime;
    var DepthTotal int32;
    if pvCexes {
        *pvCexes = Vec_PtrAlloc( 1000 ); }
    for  Iter = 0; Iter < IterMax; Iter++  {
        if Gia_ManPoNum(pNew) == 0 {
            if fVerbose != 0 {
                fmt.Printf("Finished all POs.\n"); }
            break;
        }
        // run BMC till the first failure
        clk2 = Abc_Clock(); 
        pCex = Bmc_ChainFailOneOutput( pNew, nFrameMax, nConfMax, fVerbose, fVeryVerbose );
        clkBmc += Abc_Clock() - clk2;
        if pCex == nil {
            if fVerbose != 0 {
                fmt.Printf("BMC could not detect a failed output in %d frames with %d conflicts.\n", nFrameMax, nConfMax); }
            break;
        }
        assert(!(Iter != 0) || pCex.iFrame > 0, "!Iter || pCex.iFrame > 0");
        // move the AIG to the failure state
        clk2 = Abc_Clock(); 
        pNew = Gia_ManVerifyCexAndMove( pTemp = pNew, pCex );
        Gia_ManStop( pTemp );
        DepthTotal += pCex.iFrame;
        clkMov += Abc_Clock() - clk2;
        // find outputs that fail in this state
        clk2 = Abc_Clock(); 
        vOutputs = Bmc_ChainFindFailedOutputs( pNew, pvCexes ? *pvCexes : nil );
        assert(Vec_IntFind(vOutputs, pCex.iPo) >= 0, "Vec_IntFind(vOutputs, pCex.iPo) >= 0");
        // save the counter-example
        if pvCexes {
            Vec_PtrPush( *pvCexes, pCex );
        } else {
            Abc_CexFree( pCex ); }
        clkSat += Abc_Clock() - clk2;
        // remove them from the AIG 
        clk2 = Abc_Clock(); 
        pNew = Bmc_ChainCleanup( pTemp = pNew, vOutputs );
        Gia_ManStop( pTemp );
        clkCln += Abc_Clock() - clk2;
        // perform sequential cleanup
        clk2 = Abc_Clock(); 
//        pNew = Gia_ManSeqStructSweep( pTemp = pNew, 0, 1, 0 );
//        Gia_ManStop( pTemp );
        clkSwp += Abc_Clock() - clk2;
        // printout
        if fVerbose != 0 {
            var nNonConst int32  = Gia_ManCountNonConst0(pNew);
            fmt.Printf("Iter %4d : ", Iter+1);
            fmt.Printf("Depth =%5d  ", DepthTotal);
            fmt.Printf("FailPo =%5d  ", Vec_IntSize(vOutputs));
            fmt.Printf("UndecPo =%5d ", nNonConst);
            fmt.Printf("(%6.2f %%)  ", 100.0 * nNonConst / Gia_ManPoNum(pNew));
            fmt.Printf("AIG =%8d  ", Gia_ManAndNum(pNew));
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        Vec_IntFree( vOutputs );
    }
    fmt.Printf("Completed a CEX chain with %d segments, %d frames, and %d failed POs (out of %d). ", Iter, DepthTotal, Gia_ManPoNum(pNew) - Gia_ManCountNonConst0(pNew), Gia_ManPoNum(pNew));
    if fVerbose != 0 {
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Abc_PrintTimeP( 1, "BMC  ", clkBmc, Abc_Clock() - clk );
    Abc_PrintTimeP( 1, "Init ", clkMov, Abc_Clock() - clk );
    Abc_PrintTimeP( 1, "SAT  ", clkSat, Abc_Clock() - clk );
    Abc_PrintTimeP( 1, "Clean", clkCln, Abc_Clock() - clk );
//    Abc_PrintTimeP( 1, "Sweep", clkSwp, Abc_Clock() - clk );
    }
    Gia_ManStop( pNew );
    if pvCexes {
        fmt.Printf("Total number of CEXes collected = %d.\n", Vec_PtrSize(*pvCexes)); }
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

