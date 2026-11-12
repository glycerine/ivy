package bmc

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [bmcBmcG.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [New BMC package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 20, 2017.]

  Revision    [$Id: bmcBmcG.c,v 1.00 2017/07/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/glucose/AbcGlucose.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const PAR_THR_MAX = 100

   
type Bmcg_Man_t struct {
    pPars     *Bmc_AndPar_t;               // parameters
    pGia        *Gia_Man_t;                // user's AIG
    pFrames        *Gia_Man_t;             // unfolded AIG (pFrames->vCopies point to pClean)
    pClean        *Gia_Man_t;              // incremental AIG (pClean->Value point to pFrames)
    vGia2Fr         Vec_Ptr_t;             // copies of GIA in each timeframe
    vFr2Sat         Vec_Int_t;             // mapping of objects in pFrames into SAT variables
    vCiMap         Vec_Int_t;              // maps CIs of pFrames into CIs/frames of GIA
    bmcg_sat_solver * pSats[PAR_THR_MAX];  // concurrent SAT solvers
    nSatVars               int32;            // number of SAT variables used
    nOldFrPis               int32;           // number of primary inputs
    nOldFrPos               int32;           // number of primary output
    fStopNow               int32;            // signal when it is time to stop
    timeUnf           abctime;             // runtime of unfolding
    timeCnf           abctime;             // runtime of CNF generation
    timeSmp           abctime;             // runtime of CNF simplification
    timeSat           abctime;             // runtime of the solvers
    timeOth           abctime;             // other runtime
}

func Bmcg_ManCopies(p *Bmcg_Man_t, f int32) *int32 { return (int*)Vec_PtrEntry(&p.vGia2Fr, f); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmcg_ManStart(pGia *Gia_Man_t, pPars *Bmc_AndPar_t) *Bmcg_Man_t {
    // inlining of macro ABC_CALLOC
    var p *Bmcg_Man_t  = &Bmcg_Man_t{}; 
    var i int32;
    var Lit int32 = Abc_Var2Lit( 0, 1 );
//    opts.conf_limit = pPars->nConfLimit;
    assert(Gia_ManRegNum(pGia) > 0, "Gia_ManRegNum(pGia) > 0");
    p.pPars   = pPars;
    p.pGia    = pGia;
    p.pFrames = Gia_ManStart( 3*Gia_ManObjNum(pGia) );  Gia_ManHashStart(p.pFrames);
    p.pClean  = nil;
//    Vec_PtrFill( &p->vGia2Fr, Gia_ManCountTents(pGia)+1, NULL );
//    for ( i = 0; i < Vec_PtrSize(&p->vGia2Fr); i++ )
//        Vec_PtrWriteEntry( &p->vGia2Fr, i, ABC_FALLOC(int, Gia_ManObjNum(pGia)) );
    Vec_PtrGrow( &p.vGia2Fr, 1000 );  
    Vec_IntGrow( &p.vFr2Sat, 3*Gia_ManCiNum(pGia) );  
    Vec_IntPush( &p.vFr2Sat, 0 );
    Vec_IntGrow( &p.vCiMap, 3*Gia_ManCiNum(pGia) );
    for  i = 0; i < p.pPars.nProcs; i++  {
        p.pSats[i] = bmcg_sat_solver_start();  
//        p->pSats[i]->SolverType = i;
        bmcg_sat_solver_addvar( p.pSats[i] );
        bmcg_sat_solver_addclause( p.pSats[i], &Lit, 1 );  
        bmcg_sat_solver_set_stop( p.pSats[i], &p.fStopNow );
    }
    p.nSatVars = 1;
    return p;
}
func Bmcg_ManStop(p *Bmcg_Man_t) {
    var i int32;
    Gia_ManStopP( &p.pFrames );
    Gia_ManStopP( &p.pClean );
    Vec_PtrFreeData( &p.vGia2Fr );
    Vec_PtrErase( &p.vGia2Fr );
    Vec_IntErase( &p.vFr2Sat );
    Vec_IntErase( &p.vCiMap );
    for  i = 0; i < p.pPars.nProcs; i++  {
        if p.pSats[i] { 
            bmcg_sat_solver_stop( p.pSats[i] ); }
    }
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Incremental unfolding.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmcg_ManUnfold_rec(p *Bmcg_Man_t, iObj int32, f int32) int32 {
    var pObj *Gia_Obj_t; 
    var iLit int32;
    var pCopies *int32 = Bmcg_ManCopies( p, f );
    if pCopies[iObj] >= 0 {
        return pCopies[iObj]; }
    pObj = Gia_ManObj( p.pGia, iObj );
    if Gia_ObjIsCi(pObj) {
        if Gia_ObjIsPi(p.pGia, pObj) {
            Vec_IntPushTwo( &p.vCiMap, Gia_ObjCioId(pObj), f );
            iLit = Gia_ManAppendCi( p.pFrames );
        
        } else if f > 0 {
            pObj = Gia_ObjRoToRi( p.pGia, pObj );
            iLit = Bmcg_ManUnfold_rec( p, Gia_ObjFaninId0p(p.pGia, pObj), f-1 );
            iLit = Abc_LitNotCond( iLit, Gia_ObjFaninC0(pObj) );
        }
    
    } else if Gia_ObjIsAnd(pObj) {
        iLit = Bmcg_ManUnfold_rec( p, Gia_ObjFaninId0(pObj, iObj), f );
        iLit = Abc_LitNotCond( iLit, Gia_ObjFaninC0(pObj) );
        if iLit > 0 {
            var iNew int32;
            iNew = Bmcg_ManUnfold_rec( p, Gia_ObjFaninId1(pObj, iObj), f );
            iNew = Abc_LitNotCond( iNew, Gia_ObjFaninC1(pObj) );
            iLit = Gia_ManHashAnd( p.pFrames, iLit, iNew );
        }
    
    } else { assert(false, "0"); }
    return (pCopies[iObj] = iLit);
}
func Bmcg_ManCollect_rec(p *Bmcg_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t; 
    var iSatVar int32;
    var iLitClean int32 = Gia_ObjCopyArray( p.pFrames, iObj );
    if iLitClean >= 0 {
        return iLitClean; }
    pObj = Gia_ManObj( p.pFrames, iObj );
    iSatVar = Vec_IntEntry( &p.vFr2Sat, iObj );
    if iSatVar > 0 || Gia_ObjIsCi(pObj) {
        iLitClean = Gia_ManAppendCi( p.pClean );
    } else if Gia_ObjIsAnd(pObj) {
        var iLit0 int32  = Bmcg_ManCollect_rec( p, Gia_ObjFaninId0(pObj, iObj) );
        var iLit1 int32  = Bmcg_ManCollect_rec( p, Gia_ObjFaninId1(pObj, iObj) );
        iLit0 = Abc_LitNotCond( iLit0, Gia_ObjFaninC0(pObj) );
        iLit1 = Abc_LitNotCond( iLit1, Gia_ObjFaninC1(pObj) );
        iLitClean = Gia_ManAppendAnd( p.pClean, iLit0, iLit1 );
    
    } else { assert(false, "0"); }
    assert(!Abc_LitIsCompl(iLitClean), "!Abc_LitIsCompl(iLitClean)");
    Gia_ManObj( p.pClean, Abc_Lit2Var(iLitClean) ).Value = iObj;
    Gia_ObjSetCopyArray( p.pFrames, iObj, iLitClean );
    return iLitClean;
}
func Bmcg_ManUnfold(p *Bmcg_Man_t, f int32, nFramesAdd int32) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil; var pObj *Gia_Obj_t;
    var i int32;
    var k int32;
    var iLitFrame int32;
    var iLitClean int32;
    var fTrivial int32 = 1;
    var pCopies *int32;
    var nFrameObjs int32 = Gia_ManObjNum(p.pFrames);
    assert(Gia_ManPoNum(p.pFrames) == f * Gia_ManPoNum(p.pGia), "Gia_ManPoNum(p.pFrames) == f * Gia_ManPoNum(p.pGia)");
    for  k = 0; k < nFramesAdd; k++  {
        // unfold this timeframe
        // inlining of macro ABC_FALLOC
        Vec_PtrPush( &p.vGia2Fr, ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(p.pGia))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(p.pGia)))) );
        assert(Vec_PtrSize(&p.vGia2Fr) == f+k+1, "Vec_PtrSize(&p.vGia2Fr) == f+k+1");
        pCopies = Bmcg_ManCopies( p, f+k );
        //memset( pCopies, 0xFF, sizeof(int)*Gia_ManObjNum(p->pGia) );  
        pCopies[0] = 0;
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p.pGia)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
            iLitFrame = Bmcg_ManUnfold_rec( p, Gia_ObjFaninId0p(p.pGia, pObj), f+k );
            iLitFrame = Abc_LitNotCond( iLitFrame, Gia_ObjFaninC0(pObj) );
            pCopies[Gia_ObjId(p.pGia, pObj)] = Gia_ManAppendCo( p.pFrames, iLitFrame );
            fTrivial &= (iLitFrame == 0);
        }
    }
    if fTrivial != 0 {
        return nil; }
    // create a clean copy of the new nodes of this timeframe
    Vec_IntFillExtra( &p.vFr2Sat, Gia_ManObjNum(p.pFrames), -1 );
    Vec_IntFillExtra( &p.pFrames.vCopies, Gia_ManObjNum(p.pFrames), -1 );
    //assert( Vec_IntCountEntry(&p->pFrames->vCopies, -1) == Vec_IntSize(&p->pFrames->vCopies) );
    Gia_ManStopP( &p.pClean );
    p.pClean = Gia_ManStart( Gia_ManObjNum(p.pFrames) - nFrameObjs + 1000 );
    Gia_ObjSetCopyArray( p.pFrames, 0, 0 );
    for  k = 0; k < nFramesAdd; k++  {
    for  i = 0; i < Gia_ManPoNum(p.pGia); i++  {
        pObj = Gia_ManCo( p.pFrames, (f+k) * Gia_ManPoNum(p.pGia) + i );
        iLitClean = Bmcg_ManCollect_rec( p, Gia_ObjFaninId0p(p.pFrames, pObj) );
        iLitClean = Abc_LitNotCond( iLitClean, Gia_ObjFaninC0(pObj) );
        iLitClean = Gia_ManAppendCo( p.pClean, iLitClean );
        Gia_ManObj( p.pClean, Abc_Lit2Var(iLitClean) ).Value = Gia_ObjId(p.pFrames, pObj);
        Gia_ObjSetCopyArray( p.pFrames, Gia_ObjId(p.pFrames, pObj), iLitClean );
    }
    }
    pNew = p.pClean; p.pClean = nil;
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pNew.nObjs) && ((pObj) = Gia_ManObj(pNew, i)); i++  {
        Gia_ObjSetCopyArray( p.pFrames, pObj.Value, -1 );
    }
    return pNew;
}
func Bmcg_ManAddNewCnf(p *Bmcg_Man_t, f int32, nFramesAdd int32) *Cnf_Dat_t {
    var clk abctime  = Abc_Clock();
    var pNew *Gia_Man_t  = Bmcg_ManUnfold( p, f, nFramesAdd );
    var pCnf *Cnf_Dat_t;
    var pObj *Gia_Obj_t; 
    var i int32;
    var iVar int32;
    var pMap *int32;
    p.timeUnf += Abc_Clock() - clk;
    if pNew == nil {
        return nil; }
    clk = Abc_Clock();
    pCnf = (Cnf_Dat_t *) Mf_ManGenerateCnf( pNew, 8, 1, 0, 0, 0 );
    // inlining of macro ABC_FALLOC
    pMap = ((int *) memset(malloc(sizeof(int) * (size_t)(Gia_ManObjNum(pNew))), 0xff, sizeof(int) * (size_t)(Gia_ManObjNum(pNew))));
    pMap[0] = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < pNew.nObjs) && ((pObj) = Gia_ManObj(pNew, i)); i++  {
        if pCnf.pObj2Count[i] <= 0 && !Gia_ObjIsCi(pObj) {
            continue; }
        iVar = Vec_IntEntry( &p.vFr2Sat, pObj.Value );
        if iVar == -1 {
            Vec_IntWriteEntry( &p.vFr2Sat, pObj.Value, (iVar = p.nSatVars++) ); }
        pMap[i] = iVar;
    }
    Gia_ManStop( pNew );
    for  i = 0; i < pCnf.nLiterals; i++  {
        pCnf.pClauses[0][i] = Abc_Lit2LitV( pMap, pCnf.pClauses[0][i] );
    }
    pMap = nil;
    p.timeCnf += Abc_Clock() - clk;
    return pCnf;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmcg_ManPrintFrame(p *Bmcg_Man_t, f int32, nClauses int32, Solver int32, clkStart abctime) {
    var fUnfinished int32;
    if !p.pPars.fVerbose {
        return; }
    Abc_Print( 1, "%4d %s : ", f,   fUnfinished ? "-" : "+" );
//    Abc_Print( 1, "Var =%8.0f.  ",  (double)p->nSatVars ); 
//    Abc_Print( 1, "Cla =%9.0f.  ",  (double)nClauses );  
    Abc_Print( 1, "Var =%8.0f.  ",  float64((bmcg_sat_solver_varnum(p.pSats[0])-bmcg_sat_solver_elim_varnum(p.pSats[0]))) ); 
    Abc_Print( 1, "Cla =%9.0f.  ",  float64(bmcg_sat_solver_clausenum(p.pSats[0])) );  
    Abc_Print( 1, "Learn =%9.0f.  ",float64(bmcg_sat_solver_learntnum(p.pSats[0])) );  
    Abc_Print( 1, "Conf =%9.0f.  ", float64(bmcg_sat_solver_conflictnum(p.pSats[0])) );  
    if p.pPars.nProcs > 1 {
        Abc_Print( 1, "S = %3d. ",  Solver ); }
    Abc_Print( 1, "%4.0f MB",       1.0*(int32(Gia_ManMemory(p.pFrames)) + Vec_IntMemory(&p.vFr2Sat))/(1<<20) );
    Abc_Print( 1, "%9.2f sec  ",    float32((Abc_Clock() - clkStart))/float32((int64(time.Second))) );
    fmt.Printf("\n");
    fflush( os.Stdout );
}
func Bmcg_ManPrintTime(p *Bmcg_Man_t) {
    var clkTotal abctime  = p.timeUnf + p.timeCnf + p.timeSmp + p.timeSat + p.timeOth;
    if !p.pPars.fVerbose {
        return; }
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Unfolding     "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeUnf)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((p.timeUnf)))/(float64((clkTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("CNF generation"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCnf)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((p.timeCnf)))/(float64((clkTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("CNF simplify  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSmp)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((p.timeSmp)))/(float64((clkTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("SAT solving   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((p.timeSat)))/(float64((clkTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other         "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOth)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((p.timeOth)))/(float64((clkTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("TOTAL         "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((clkTotal)))/(float64(int64(time.Second))), (float64((clkTotal)))? 100.0*(float64((clkTotal)))/(float64((clkTotal))) : 0.0);;
}
func Bmcg_ManGenerateCex(p *Bmcg_Man_t, i int32, f int32, s int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t  = Abc_CexMakeTriv( Gia_ManRegNum(p.pGia), Gia_ManPiNum(p.pGia), Gia_ManPoNum(p.pGia), f*Gia_ManPoNum(p.pGia)+i );
    var pObj *Gia_Obj_t;  var k int32;
    // inlining of macro Gia_ManForEachPi
    for  k = 0; (k < Gia_ManPiNum(p.pFrames)) && ((pObj) = Gia_ManCi(p.pFrames, k)); k++  {
        var iSatVar int32  = Vec_IntEntry( &p.vFr2Sat, Gia_ObjId(p.pFrames, pObj) );
        if iSatVar > 0 && bmcg_sat_solver_read_cex_varvalue(p.pSats[s], iSatVar) { // 1 bit
        
            var iCiId int32    = Vec_IntEntry( &p.vCiMap, 2*k+0 );
            var iFrame int32   = Vec_IntEntry( &p.vCiMap, 2*k+1 );
            Abc_InfoSetBit( pCex.pData, Gia_ManRegNum(p.pGia) + iFrame * Gia_ManPiNum(p.pGia) + iCiId );
        }
    }
    return pCex;
}
func Bmcg_ManAddCnf(p *Bmcg_Man_t, pSat *bmcg_sat_solver, pCnf *Cnf_Dat_t) {
    var i, iSatVar int32;
    var clk abctime  = Abc_Clock();
    bmcg_sat_solver_set_nvars( pSat, p.nSatVars );
    if p.pPars.fUseEliminate {
        for  i = p.nOldFrPis; i < Gia_ManPiNum(p.pFrames); i++  {
            var pObj *Gia_Obj_t  = Gia_ManPi( p.pFrames, i );
            var iSatVar int32  = Vec_IntEntry( &p.vFr2Sat, Gia_ObjId(p.pFrames, pObj) );
            if iSatVar > 0 {
                bmcg_sat_solver_var_set_frozen( pSat, iSatVar, 1 ); }
        }
        for  i = p.nOldFrPos; i < Gia_ManPoNum(p.pFrames); i++  {
            var pObj *Gia_Obj_t  = Gia_ManPo( p.pFrames, i );
            var iSatVar int32  = Vec_IntEntry( &p.vFr2Sat, Gia_ObjId(p.pFrames, pObj) );
            if iSatVar > 0 {
                bmcg_sat_solver_var_set_frozen( pSat, iSatVar, 1 ); }
        }
        p.nOldFrPis = Gia_ManPiNum(p.pFrames);
        p.nOldFrPos = Gia_ManPoNum(p.pFrames);
    }
    for  i = 0; i < pCnf.nClauses; i++  {
        if !bmcg_sat_solver_addclause( pSat, pCnf.pClauses[i], pCnf.pClauses[i+1]-pCnf.pClauses[i] ) {
            assert(false, "0"); }
    }
    if !p.pPars.fUseEliminate {
        return; }
    bmcg_sat_solver_eliminate( pSat, 0 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vFr2Sat)) && (((iSatVar) = Vec_IntEntry(&p.vFr2Sat, i)), 1); i++  {
        if iSatVar > 0 && bmcg_sat_solver_var_is_elim(pSat, iSatVar) {
            Vec_IntWriteEntry( &p.vFr2Sat, i, -1 ); }
    }
    p.timeSmp += Abc_Clock() - clk;
}
func Bmcg_ManPerformOne(pGia *Gia_Man_t, pPars *Bmc_AndPar_t) int32 {
    var clkStart abctime  = Abc_Clock();
    var p *Bmcg_Man_t  = Bmcg_ManStart( pGia, pPars );
    var f int32;
    var k int32;
    var i int32 = Gia_ManPoNum(pGia);
    var status int32;
    var RetValue int32 = -1;
    var nClauses int32;
    Abc_CexFreeP( &pGia.pCexSeq );
    for  f = 0; !pPars.nFramesMax || f < pPars.nFramesMax; f += pPars.nFramesAdd  {
        var pCnf *Cnf_Dat_t  = Bmcg_ManAddNewCnf( p, f, pPars.nFramesAdd );
        if pCnf == nil {
            Bmcg_ManPrintFrame( p, f, nClauses, -1, clkStart );
            if pPars.pFuncOnFrameDone {
                for  k = 0; k < pPars.nFramesAdd; k++  {
                for  i = 0; i < Gia_ManPoNum(pGia); i++  {
                    pPars.pFuncOnFrameDone(f+k, i, 0);
                }
                } }
            continue;
        }
        nClauses += pCnf.nClauses;
        Bmcg_ManAddCnf( p, p.pSats[0], pCnf );
        Cnf_DataFree( pCnf );
        assert(Gia_ManPoNum(p.pFrames) == (f + pPars.nFramesAdd) * Gia_ManPoNum(pGia), "Gia_ManPoNum(p.pFrames) == (f + pPars.nFramesAdd) * Gia_ManPoNum(pGia)");
        for  k = 0; k < pPars.nFramesAdd; k++  {
            for  i = 0; i < Gia_ManPoNum(pGia); i++  {
                var clk abctime  = Abc_Clock();
                var iObj int32  = Gia_ObjId( p.pFrames, Gia_ManCo(p.pFrames, (f+k) * Gia_ManPoNum(pGia) + i) );
                var iLit int32  = Abc_Var2Lit( Vec_IntEntry(&p.vFr2Sat, iObj), 0 );
                if pPars.nTimeOut && (Abc_Clock() - clkStart)/int64(time.Second) >= pPars.nTimeOut {
                    break; }
                status = bmcg_sat_solver_solve( p.pSats[0], &iLit, 1 );
                p.timeSat += Abc_Clock() - clk;
                if status == -1 { // unsat
                
                    if i == Gia_ManPoNum(pGia)-1 {
                        Bmcg_ManPrintFrame( p, f+k, nClauses, -1, clkStart ); }
                    if pPars.pFuncOnFrameDone {
                        pPars.pFuncOnFrameDone(f+k, i, 0); }
                    continue;
                }
                if status == 1 { // sat
                
                    RetValue = 0;
                    pPars.iFrame = f+k;
                    pGia.pCexSeq = Bmcg_ManGenerateCex( p, i, f+k, 0 );
                    pPars.nFailOuts++;
                    Bmcg_ManPrintFrame( p, f+k, nClauses, -1, clkStart );
                    if !pPars.fNotVerbose {
                        var nOutDigits int32  = Abc_Base10Log( Gia_ManPoNum(pGia) );
                        Abc_Print( 1, "Output %*d was asserted in frame %2d (solved %*d out of %*d outputs).  ",  
                            nOutDigits, i, f+k, nOutDigits, pPars.nFailOuts, nOutDigits, Gia_ManPoNum(pGia) );
                        fflush( os.Stdout );
                    }
                    if pPars.pFuncOnFrameDone {
                        pPars.pFuncOnFrameDone(f+k, i, 1); }
                }
                break;
            }
            if i < Gia_ManPoNum(pGia) || f+k == pPars.nFramesMax-1 {
                break; }
        }
        if k < pPars.nFramesAdd {
            break; }
    }
    p.timeOth = Abc_Clock() - clkStart - p.timeUnf - p.timeCnf - p.timeSmp - p.timeSat;
    if RetValue == -1 && !pPars.fNotVerbose {
        fmt.Printf("No output failed in %d frames.  ", f + (k < pPars.nFramesAdd ? k+1 : 0)); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
    Bmcg_ManPrintTime( p );
    Bmcg_ManStop( p );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmcg_ManPerform(pGia *Gia_Man_t, pPars *Bmc_AndPar_t) int32 { 
    pPars.nProcs = 1;
    return Bmcg_ManPerformOne( pGia, pPars );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

