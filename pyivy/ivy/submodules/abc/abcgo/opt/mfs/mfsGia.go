package mfs

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [mfsGia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Experimental code based on the new AIG package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 29, 2009.]

  Revision    [$Id: mfsGia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "mfsInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Gia_ObjChild0Copy(pObj *Aig_Obj_t) int32 { return Gia_LitNotCond( Aig_ObjFanin0(pObj).iData, Aig_ObjFaninC0(pObj) ); }
func Gia_ObjChild1Copy(pObj *Aig_Obj_t) int32 { return Gia_LitNotCond( Aig_ObjFanin1(pObj).iData, Aig_ObjFaninC1(pObj) ); }

// r i10_if6.blif;  ps; mfs -v
// r pj1_if6.blif;  ps; mfs -v
// r x/01_if6.blif; ps; mfs -v

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives the resubstitution miter as an GIA.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Gia_ManCreateResubMiter(p *Aig_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;//, * pTemp;
    var pObj *Aig_Obj_t;
    var i int32;
    var pOuts0 *int32;
    var pOuts1 *int32;
    Aig_ManSetPioNumbers( p );
    // create the new manager
    pNew = Gia_ManStart( Aig_ManObjNum(p) );
    pNew.pName = Gia_UtilStrsav( p.pName );
    pNew.pSpec = Gia_UtilStrsav( p.pSpec );
    Gia_ManHashAlloc( pNew );
    // create the objects
    // inlining of macro ABC_ALLOC
    pOuts0 = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManPoNum(p))));
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsAnd(pObj) {
            pObj.iData = Gia_ManHashAnd( pNew, Gia_ObjChild0Copy(pObj), Gia_ObjChild1Copy(pObj) );
        } else if Aig_ObjIsPi(pObj) {
            pObj.iData = Gia_ManAppendCi( pNew );
        } else if Aig_ObjIsPo(pObj) {
            pOuts0[ Aig_ObjPioNum(pObj) ] = Gia_ObjChild0Copy(pObj);
        } else if Aig_ObjIsConst1(pObj) {
            pObj.iData = 1;
        } else {
            assert(false, "0"); }
    }
    }
    // create the objects
    // inlining of macro ABC_ALLOC
    pOuts1 = ((int *) malloc(sizeof(int) * (size_t)(Aig_ManPoNum(p))));
    // inlining of macro Aig_ManForEachObj
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil {} else {
        if Aig_ObjIsAnd(pObj) {
            pObj.iData = Gia_ManHashAnd( pNew, Gia_ObjChild0Copy(pObj), Gia_ObjChild1Copy(pObj) );
        } else if Aig_ObjIsPi(pObj) {
            pObj.iData = Gia_ManAppendCi( pNew );
        } else if Aig_ObjIsPo(pObj) {
            pOuts1[ Aig_ObjPioNum(pObj) ] = Gia_ObjChild0Copy(pObj);
        } else if Aig_ObjIsConst1(pObj) {
            pObj.iData = 1;
        } else {
            assert(false, "0"); }
    }
    }
    // add the outputs
    Gia_ManAppendCo( pNew, pOuts0[0] );
    Gia_ManAppendCo( pNew, pOuts1[0] );
    Gia_ManAppendCo( pNew, pOuts0[1] );
    Gia_ManAppendCo( pNew, Gia_LitNot(pOuts1[1]) );
    for  i = 2; i < Aig_ManPoNum(p); i++  {
        Gia_ManAppendCo( pNew, Gia_LitNot( Gia_ManHashXor(pNew, pOuts0[i], pOuts1[i]) ) );
    }
    Gia_ManHashStop( pNew );
    // inlining of macro ABC_FREE
    ((pOuts0) ? (free((char *) (pOuts0)), (pOuts0) = 0) : 0);
    // inlining of macro ABC_FREE
    ((pOuts1) ? (free((char *) (pOuts1)), (pOuts1) = 0) : 0);
//    pNew = Gia_ManCleanup( pTemp = pNew );
//    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsConstructGia(p *Mfs_Man_t) {
    var nBTLimit int32  = 500;
    // prepare AIG
    assert(p.pGia == nil, "p.pGia == nil");
    p.pGia = Gia_ManCreateResubMiter( p.pAigWin );
    // prepare AIG
    Gia_ManCreateRefs( p.pGia );
    Gia_ManCleanMark0( p.pGia );
    Gia_ManCleanMark1( p.pGia );
    Gia_ManFillValue ( p.pGia ); // maps nodes into trail ids
    Gia_ManCleanPhase( p.pGia ); 
    // prepare solver
    p.pTas = Tas_ManAlloc( p.pGia, nBTLimit );
    p.vCex = Tas_ReadModel( p.pTas );
    p.vGiaLits = Vec_PtrAlloc( 100 );
}


/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsDeconstructGia(p *Mfs_Man_t) {
    assert(p.pGia != nil, "p.pGia != nil");
    Gia_ManStop( p.pGia ); p.pGia = nil;
    Tas_ManStop( p.pTas ); p.pTas = nil;
    Vec_PtrFree( p.vGiaLits );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsResimulate(p *Gia_Man_t, vCex *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i, Entry int32;
//    Gia_ManCleanMark1( p );
    Gia_ManConst0(p).fMark1 = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark1 = 0;
    }
//        pObj->fMark1 = Gia_ManRandom(0);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCex)) && (((Entry) = Vec_IntEntry(vCex, i)), 1); i++  {
        pObj = Gia_ManCi( p, Gia_Lit2Var(Entry) );
        pObj.fMark1 = !Gia_LitIsCompl(Entry);
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.fMark1 = (Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj)) & 
                       (Gia_ObjFanin1(pObj).fMark1 ^ Gia_ObjFaninC1(pObj)); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.fMark1 =  Gia_ObjFanin0(pObj).fMark1 ^ Gia_ObjFaninC0(pObj);
    }
}




/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsTryResubOnceGia(p *Mfs_Man_t, pCands *int32, nCands int32) int32 {
    var fVeryVerbose int32;
    var fUseGia int32  = 1;
    var pData *uint32;
    var i, iVar, status, iOut int32;
    var clk int64  = time.Now().UnixNano();
    p.nSatCalls++;
//    return -1;
    assert(p.pGia != nil, "p.pGia != nil");
    assert(p.pTas != nil, "p.pTas != nil");
    // convert to literals
    Vec_PtrClear( p.vGiaLits );
    // create the first four literals
    Vec_PtrPush( p.vGiaLits, Gia_ObjChild0(Gia_ManPo(p.pGia, 0)) );
    Vec_PtrPush( p.vGiaLits, Gia_ObjChild0(Gia_ManPo(p.pGia, 1)) );
    Vec_PtrPush( p.vGiaLits, Gia_ObjChild0(Gia_ManPo(p.pGia, 2)) );
    Vec_PtrPush( p.vGiaLits, Gia_ObjChild0(Gia_ManPo(p.pGia, 3)) );
    for  i = 0; i < nCands; i++  {
        // get the output number
        iOut = Gia_Lit2Var(pCands[i]) - 2 * p.pCnf.nVars;
        // write the literal
        Vec_PtrPush( p.vGiaLits, Gia_ObjChild0(Gia_ManPo(p.pGia, 4 + iOut)) );
    }
    // perform SAT solving
    status = Tas_ManSolveArray( p.pTas, p.vGiaLits );
    if status == -1 {
        p.nTimeOuts++;
        if fVeryVerbose != 0 {
        fmt.Printf("t"); }
//        p->nSatUndec++;
//        p->nConfUndec += p->Pars.nBTThis;
//        Cec_ManSatAddToStore( vCexStore, NULL, i ); // timeout
//        p->timeSatUndec += clock() - clk;
    
    } else if status == 1 {
        if fVeryVerbose != 0 {
        fmt.Printf("u"); }
//        p->nSatUnsat++;
//        p->nConfUnsat += p->Pars.nBTThis;
//        p->timeSatUnsat += clock() - clk;
    
    } else {
        p.nSatCexes++;
        if fVeryVerbose != 0 {
        fmt.Printf("s"); }
//        p->nSatSat++;
//        p->nConfSat += p->Pars.nBTThis;
//        Gia_SatVerifyPattern( pAig, pRoot, vCex, vVisit );
//        Cec_ManSatAddToStore( vCexStore, vCex, i );
//        p->timeSatSat += clock() - clk;

        // resimulate the counter-example
        Abc_NtkMfsResimulate( p.pGia, Tas_ReadModel(p.pTas) );
   
        if fUseGia != 0 {
/*
            int Val0 = Gia_ManPo(p->pGia, 0)->fMark1;
            int Val1 = Gia_ManPo(p->pGia, 1)->fMark1;
            int Val2 = Gia_ManPo(p->pGia, 2)->fMark1;
            int Val3 = Gia_ManPo(p->pGia, 3)->fMark1;
            assert( Val0 == 1 );
            assert( Val1 == 1 );
            assert( Val2 == 1 );
            assert( Val3 == 1 );
*/
            // store the counter-example
            // inlining of macro Vec_IntForEachEntry
            for  i = 0; (i < Vec_IntSize(p.vProjVarsSat)) && (((iVar) = Vec_IntEntry(p.vProjVarsSat, i)), 1); i++  {
                pData = (unsigned *)Vec_PtrEntry( p.vDivCexes, i );
                iOut = iVar - 2 * p.pCnf.nVars;
//                if ( !Gia_ManPo( p->pGia, 4 + iOut )->fMark1 ) // remove 0s!!!
                if Gia_ManPo( p.pGia, 4 + iOut ).fMark1 { // remove 0s!!!  - rememeber complemented attribute
                
                    assert(Aig_InfoHasBit(pData, p.nCexes), "Aig_InfoHasBit(pData, p.nCexes)");
                    Aig_InfoXorBit( pData, p.nCexes );
                }
            }
            p.nCexes++;
        }

    }
    return status;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

