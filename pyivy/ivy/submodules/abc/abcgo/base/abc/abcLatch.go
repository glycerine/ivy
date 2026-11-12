package abc

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcLatch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures working with latches.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcLatch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Checks if latches form self-loop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLatchIsSelfFeed_rec(pLatch *Abc_Obj_t, pLatchRoot *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    assert(Abc_ObjIsLatch(pLatch), "Abc_ObjIsLatch(pLatch)");
    if pLatch == pLatchRoot {
        return 1; }
    pFanin = Abc_ObjFanin0(Abc_ObjFanin0(pLatch));
    if !Abc_ObjIsBo(pFanin) || !Abc_ObjIsLatch(Abc_ObjFanin0(pFanin)) {
        return 0; }
    return Abc_NtkLatchIsSelfFeed_rec( Abc_ObjFanin0(pFanin), pLatch );
}

/**Function*************************************************************

  Synopsis    [Checks if latches form self-loop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkLatchIsSelfFeed(pLatch *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    assert(Abc_ObjIsLatch(pLatch), "Abc_ObjIsLatch(pLatch)");
    pFanin = Abc_ObjFanin0(Abc_ObjFanin0(pLatch));
    if !Abc_ObjIsBo(pFanin) || !Abc_ObjIsLatch(Abc_ObjFanin0(pFanin)) {
        return 0; }
    return Abc_NtkLatchIsSelfFeed_rec( Abc_ObjFanin0(pFanin), pLatch );
}

/**Function*************************************************************

  Synopsis    [Checks if latches form self-loop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCountSelfFeedLatches(pNtk *Abc_Ntk_t) int32 {
    var pLatch *Abc_Obj_t;
    var i, Counter int32;
    Counter = 0;
    Abc_NtkForEachLatch( pNtk, pLatch, i )
    {
//        if ( Abc_NtkLatchIsSelfFeed(pLatch) && Abc_ObjFanoutNum(pLatch) > 1 )
//            printf( "Fanouts = %d.\n", Abc_ObjFanoutNum(pLatch) );
        Counter += Abc_NtkLatchIsSelfFeed( pLatch );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Replaces self-feeding latches by latches with constant inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRemoveSelfFeedLatches(pNtk *Abc_Ntk_t) int32 {
    var pLatch, pConst1 *Abc_Obj_t;
    var i, Counter int32;
    Counter = 0;
    Abc_NtkForEachLatch( pNtk, pLatch, i )
    {
        if Abc_NtkLatchIsSelfFeed( pLatch ) != 0 {
            if Abc_NtkIsStrash(pNtk) {
                pConst1 = Abc_AigConst1(pNtk);
            } else {
                pConst1 = Abc_NtkCreateNodeConst1(pNtk); }
            Abc_ObjPatchFanin( Abc_ObjFanin0(pLatch), Abc_ObjFanin0(Abc_ObjFanin0(pLatch)), pConst1 );
            Counter++;
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Pipelines the network with latches.]

  Description []
               
  SideEffects [Does not check the names of the added latches!!!]

  SeeAlso     []

***********************************************************************/
func Abc_NtkLatchPipe(pNtk *Abc_Ntk_t, nLatches int32) {
    var vNodes *Vec_Ptr_t;
    var pObj, pLatch, pFanin, pFanout *Abc_Obj_t;
    var i, k, nTotal, nDigits int32;
    if nLatches < 1 {
        return; }
    nTotal = nLatches * Abc_NtkPiNum(pNtk);
    nDigits = Abc_Base10Log( nTotal );
    vNodes = Vec_PtrAlloc( 100 );
    Abc_NtkForEachPi( pNtk, pObj, i )
    {
        // remember current fanins of the PI
        Abc_NodeCollectFanouts( pObj, vNodes );
        // create the latches
        for  pFanin, k = pObj, 0; k < nLatches; k, pFanin = k+1, pLatch  {
            pLatch = Abc_NtkCreateLatch( pNtk );
            Abc_ObjAddFanin( pLatch, pFanin );
            Abc_LatchSetInitDc( pLatch );
            // create the name of the new latch
            Abc_ObjAssignName( pLatch, Abc_ObjNameDummy("LL", i*nLatches + k, nDigits), nil );
        }
        // patch the PI fanouts
        Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pFanout, k )
            Abc_ObjPatchFanin( pFanout, pObj, pFanin );
    }
    Vec_PtrFree( vNodes );
    Abc_NtkLogicMakeSimpleCos( pNtk, 0 );
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectLatchValues(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vValues *Vec_Int_t;
    var pLatch *Abc_Obj_t;
    var i int32;
    vValues = Vec_IntAlloc( Abc_NtkLatchNum(pNtk) );
    Abc_NtkForEachLatch( pNtk, pLatch, i )
        Vec_IntPush( vValues, Abc_LatchIsInit1(pLatch) );
    return vValues;
}

/**Function*************************************************************

  Synopsis    [Derives latch init string.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectLatchValuesStr(pNtk *Abc_Ntk_t) string {
    var pInits string;
    var pLatch *Abc_Obj_t;
    var i int32;
    pInits = ABC_ALLOC( char, Abc_NtkLatchNum(pNtk) + 1 );
    Abc_NtkForEachLatch( pNtk, pLatch, i )
    {
        if Abc_LatchIsInit0(pLatch) {
            pInits[i] = '0';
        } else if Abc_LatchIsInit1(pLatch) {
            pInits[i] = '1';
        } else if Abc_LatchIsInitDc(pLatch) {
            pInits[i] = 'x';
        } else {
            assert(false, "0"); }
    }
    pInits[i] = 0;
    return pInits;
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node using its SOP.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInsertLatchValues(pNtk *Abc_Ntk_t, vValues *Vec_Int_t) {
    var pLatch *Abc_Obj_t;
    var i int32;
    Abc_NtkForEachLatch( pNtk, pLatch, i )
        pLatch.pData = (void *)(ABC_PTRINT_T)(vValues? (Vec_IntEntry(vValues,i)? ABC_INIT_ONE : ABC_INIT_ZERO) : ABC_INIT_DC);
}

/**Function*************************************************************

  Synopsis    [Creates latch with the given initial value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAddLatch(pNtk *Abc_Ntk_t, pDriver *Abc_Obj_t, Init Abc_InitType_t) *Abc_Obj_t {
    var pLatchOut, pLatch, pLatchIn *Abc_Obj_t;
    pLatchOut = Abc_NtkCreateBo(pNtk);
    pLatch    = Abc_NtkCreateLatch(pNtk);
    pLatchIn  = Abc_NtkCreateBi(pNtk);
    Abc_ObjAssignName( pLatchOut, Abc_ObjName(pLatch), "_lo" );
    Abc_ObjAssignName( pLatchIn,  Abc_ObjName(pLatch), "_li" );
    Abc_ObjAddFanin( pLatchOut, pLatch );
    Abc_ObjAddFanin( pLatch, pLatchIn );
    if pDriver != nil {
    Abc_ObjAddFanin( pLatchIn, pDriver ); }
    pLatch.pData = (void *)Init;
    return pLatchOut;
}

/**Function*************************************************************

  Synopsis    [Creates MUX.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkNodeConvertToMux(pNtk *Abc_Ntk_t, pNodeC *Abc_Obj_t, pNode1 *Abc_Obj_t, pNode0 *Abc_Obj_t, pMux *Abc_Obj_t) {
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    Abc_ObjAddFanin( pMux, pNodeC );
    Abc_ObjAddFanin( pMux, pNode1 );
    Abc_ObjAddFanin( pMux, pNode0 );
    if Abc_NtkHasSop(pNtk) {
        pMux.pData = Abc_SopRegister( (Mem_Flex_t *)pNtk.pManFunc, "11- 1\n0-1 1\n" );
#ifdef ABC_USE_CUDD
    else if ( Abc_NtkHasBdd(pNtk) )
        pMux->pData = Cudd_bddIte((DdManager *)pNtk->pManFunc,Cudd_bddIthVar((DdManager *)pNtk->pManFunc,0),Cudd_bddIthVar((DdManager *)pNtk->pManFunc,1),Cudd_bddIthVar((DdManager *)pNtk->pManFunc,2)), Cudd_Ref( (DdNode *)pMux->pData );
#endif
    } else if Abc_NtkHasAig(pNtk) {
        pMux.pData = Hop_Mux((Hop_Man_t *)pNtk.pManFunc,Hop_IthVar((Hop_Man_t *)pNtk.pManFunc,0),Hop_IthVar((Hop_Man_t *)pNtk.pManFunc,1),Hop_IthVar((Hop_Man_t *)pNtk.pManFunc,2));
    } else {
        assert(false, "0"); }
}

/**Function*************************************************************

  Synopsis    [Converts registers with DC values into additional PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConvertDcLatches(pNtk *Abc_Ntk_t) {
    var pCtrl, pLatch, pMux, pPi *Abc_Obj_t;
    var Init Abc_InitType_t  = ABC_INIT_ZERO;
    var i, fFound, Counter int32;
    // check if there are latches with DC values
    Abc_NtkForEachLatch( pNtk, pLatch, i )
        if Abc_LatchIsInitDc(pLatch) {
            fFound = 1;
            break;
        }
    if !fFound {
        return; }
    // add control latch
    pCtrl = Abc_NtkAddLatch( pNtk, Abc_NtkCreateNodeConst1(pNtk), Init );
    // add fanouts for each latch with DC values
    Counter = 0;
    Abc_NtkForEachLatch( pNtk, pLatch, i )
    {
        if !Abc_LatchIsInitDc(pLatch) {
            continue; }
        // change latch value
        pLatch.pData = (void *)Init;
        // if the latch output has the same name as a PO, rename it
        if Abc_NodeFindCoFanout( Abc_ObjFanout0(pLatch) ) {
            Nm_ManDeleteIdName( pLatch.pNtk.pManName, Abc_ObjFanout0(pLatch).Id );
            Abc_ObjAssignName( Abc_ObjFanout0(pLatch), Abc_ObjName(pLatch), "_lo" );
        }
        // create new PIs
        pPi = Abc_NtkCreatePi( pNtk );
        Abc_ObjAssignName( pPi, Abc_ObjName(pLatch), "_pi" );
        // create a new node and transfer fanout from latch output to the new node
        pMux = Abc_NtkCreateNode( pNtk );
        Abc_ObjTransferFanout( Abc_ObjFanout0(pLatch), pMux );
        // convert the node into a mux
        Abc_NtkNodeConvertToMux( pNtk, pCtrl, Abc_ObjFanout0(pLatch), pPi, pMux );
        Counter++;
    }
    fmt.Printf("The number of converted latches with DC values = %d.\n", Counter);
}

/**Function*************************************************************

  Synopsis    [Transfors the array of latch names into that of latch numbers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConverLatchNamesIntoNumbers(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vResult, vNames *Vec_Ptr_t;
    var vNumbers *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var pName string;
    var i, k, Num int32;
    if pNtk.vOnehots == nil {
        return nil; }
    // set register numbers
    Abc_NtkForEachLatch( pNtk, pObj, i )
        pObj.pNext = (Abc_Obj_t *)(ABC_PTRINT_T)i;
    // add the numbers
    vResult = Vec_PtrAlloc( Vec_PtrSize(pNtk.vOnehots) );
    Vec_PtrForEachEntry( Vec_Ptr_t *, pNtk.vOnehots, vNames, i )
    {
        vNumbers = Vec_IntAlloc( Vec_PtrSize(vNames) );
        Vec_PtrForEachEntry( char *, vNames, pName, k )
        {
            Num = Nm_ManFindIdByName( pNtk.pManName, pName, ABC_OBJ_BO );
            if Num < 0 {
                continue; }
            pObj = Abc_NtkObj( pNtk, Num );
            if Abc_ObjFaninNum(pObj) != 1 || !Abc_ObjIsLatch(Abc_ObjFanin0(pObj)) {
                continue; }
            Vec_IntPush( vNumbers, int32((ABC_PTRINT_T))pObj.pNext );
        }
        if Vec_IntSize( vNumbers ) > 1 {
            Vec_PtrPush( vResult, vNumbers );
fmt.Printf("Converted %d one-hot registers.\n", Vec_IntSize(vNumbers));
        
        } else {
            Vec_IntFree( vNumbers ); }
    }
    // clean the numbers
    Abc_NtkForEachLatch( pNtk, pObj, i )
        pObj.pNext = nil;
    return vResult;
}


/**Function*************************************************************

  Synopsis    [Converts registers with DC values into additional PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConvertOnehot(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t;
    var pNtkNew *Abc_Ntk_t;
    var pObj, pFanin, pObjNew, pObjLiNew, pObjLoNew *Abc_Obj_t;
    var i int32;
    var k int32;
    var nFlops int32;
    var nStates int32;
    var iState int32;
    var pfCompl [32]int32;
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    nFlops = Abc_NtkLatchNum(pNtk);
    if nFlops == 0 {
        return Abc_NtkDup( pNtk ); }
    if nFlops > 16 {
        fmt.Printf("Cannot reencode %d flops because it will lead to 2^%d states.\n", nFlops, nFlops);
        return nil;
    }
    // check if there are latches with DC values
    iState = 0;
    Abc_NtkForEachLatch( pNtk, pObj, i )
    {
        if Abc_LatchIsInitDc(pObj) {
            fmt.Printf("Cannot process logic network with don't-care init values. Run \"zero\".\n");
            return nil;
        }
        if Abc_LatchIsInit1(pObj) {
            iState |= (1 << i); }
    }
    // transfer logic to SOPs
    Abc_NtkToSop( pNtk, -1, ABC_INFINITY );
    // create new network
    pNtkNew = Abc_NtkStartFromNoLatches( pNtk, pNtk.ntkType, pNtk.ntkFunc );
    nStates = (1 << nFlops);
    for  i = 0; i < nStates; i++  {
        pObjNew   = Abc_NtkCreateLatch( pNtkNew );
        pObjLiNew = Abc_NtkCreateBi( pNtkNew );
        pObjLoNew = Abc_NtkCreateBo( pNtkNew );
        Abc_ObjAddFanin( pObjNew, pObjLiNew );
        Abc_ObjAddFanin( pObjLoNew, pObjNew );
        if i == iState {
            Abc_LatchSetInit1( pObjNew );
        } else {
            Abc_LatchSetInit0( pObjNew ); }
    }
    Abc_NtkAddDummyBoxNames( pNtkNew );
    assert(Abc_NtkLatchNum(pNtkNew) == nStates, "Abc_NtkLatchNum(pNtkNew) == nStates");
    assert(Abc_NtkPiNum(pNtkNew) == Abc_NtkPiNum(pNtk), "Abc_NtkPiNum(pNtkNew) == Abc_NtkPiNum(pNtk)");
    assert(Abc_NtkPoNum(pNtkNew) == Abc_NtkPoNum(pNtk), "Abc_NtkPoNum(pNtkNew) == Abc_NtkPoNum(pNtk)");
    assert(Abc_NtkCiNum(pNtkNew) == Abc_NtkPiNum(pNtkNew) + nStates, "Abc_NtkCiNum(pNtkNew) == Abc_NtkPiNum(pNtkNew) + nStates");
    assert(Abc_NtkCoNum(pNtkNew) == Abc_NtkPoNum(pNtkNew) + nStates, "Abc_NtkCoNum(pNtkNew) == Abc_NtkPoNum(pNtkNew) + nStates");
    assert(Abc_NtkCiNum(pNtk) == Abc_NtkPiNum(pNtk) + nFlops, "Abc_NtkCiNum(pNtk) == Abc_NtkPiNum(pNtk) + nFlops");
    assert(Abc_NtkCoNum(pNtk) == Abc_NtkPoNum(pNtk) + nFlops, "Abc_NtkCoNum(pNtk) == Abc_NtkPoNum(pNtk) + nFlops");
    // create hot-to-log transformers
    for  i = 0; i < nFlops; i++  {
        pObjNew = Abc_NtkCreateNode( pNtkNew );
        for  k = 0; k < nStates; k++  {
            if (k >> i) & 1 {
                Abc_ObjAddFanin( pObjNew, Abc_NtkCi(pNtkNew, Abc_NtkPiNum(pNtkNew)+k) ); }
        }
        assert(Abc_ObjFaninNum(pObjNew) == nStates/2, "Abc_ObjFaninNum(pObjNew) == nStates/2");
        pObjNew.pData = Abc_SopCreateOr( (Mem_Flex_t *)pNtkNew.pManFunc, nStates/2, nil );
        // save the new flop
        pObj = Abc_NtkCi( pNtk, Abc_NtkPiNum(pNtk) + i );
        pObj.pCopy = pObjNew;
    }
    // duplicate the nodes
    vNodes = Abc_NtkDfs( pNtk, 0 );
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
    {
        pObj.pCopy = Abc_NtkDupObj( pNtkNew, pObj, 1 );
        Abc_ObjForEachFanin( pObj, pFanin, k )
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy );
    }
    Vec_PtrFree( vNodes );
    // connect the POs
    Abc_NtkForEachPo( pNtk, pObj, i )
        Abc_ObjAddFanin( pObj.pCopy, Abc_ObjNotCond(Abc_ObjFanin0(pObj).pCopy, Abc_ObjFaninC0(pObj)) );
    // write entries into the nodes
    Abc_NtkForEachCo( pNtk, pObj, i )
        pObj.pCopy = Abc_ObjNotCond(Abc_ObjFanin0(pObj).pCopy, Abc_ObjFaninC0(pObj));
    // create log-to-hot transformers
    for  k = 0; k < nStates; k++  {
        pObjNew = Abc_NtkCreateNode( pNtkNew );
        for  i = 0; i < nFlops; i++  {
            pObj = Abc_NtkCo( pNtk, Abc_NtkPoNum(pNtk) + i );
            Abc_ObjAddFanin( pObjNew, Abc_ObjRegular(pObj.pCopy) );
            pfCompl[i] = Abc_ObjIsComplement(pObj.pCopy) ^ !((k >> i) & 1);
        }
        pObjNew.pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtkNew.pManFunc, nFlops, pfCompl );
        // connect it to the flop input
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, Abc_NtkPoNum(pNtkNew)+k), pObjNew );
    }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkConvertOnehot(): Network check has failed.\n"); }
    return pNtkNew;
}

//ABC_NAMESPACE_IMPL_END

//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    [Performs retiming with classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeWithClassesAig(pMan *Aig_Man_t, vClasses *Vec_Int_t, pvClasses **Vec_Int_t, fVerbose int32) *Aig_Man_t {
    var pManNew *Aig_Man_t;
    var pGia, pGiaNew *Gia_Man_t;
    pGia = Gia_ManFromAigSimple( pMan );
    assert(Gia_ManRegNum(pGia) == Vec_IntSize(vClasses), "Gia_ManRegNum(pGia) == Vec_IntSize(vClasses)");
    pGia.vFlopClasses = vClasses;
    pGiaNew = Gia_ManRetimeForward( pGia, 10, fVerbose );
    *pvClasses = pGiaNew.vFlopClasses; 
    pGiaNew.vFlopClasses = nil;
    pManNew = Gia_ManToAig( pGiaNew, 0 );
    Gia_ManStop( pGiaNew );
    Gia_ManStop( pGia );
    return pManNew;
}

/**Function*************************************************************

  Synopsis    [Performs retiming with classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeWithClassesNtk(pNtk *Abc_Ntk_t, vClasses *Vec_Int_t, pvClasses **Vec_Int_t, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    extern Abc_Ntk_t * Abc_NtkFromDarSeqSweep( Abc_Ntk_t * pNtkOld, Aig_Man_t * pMan );
    var pNtkAig, pNtkAigRet, pNtkRes *Abc_Ntk_t;
    var pMan, pManNew *Aig_Man_t;
    pNtkAig    = Abc_NtkStrash( pNtk, 0, 1, 0 );
    pMan       = Abc_NtkToDar( pNtkAig, 0, 1 );
    pManNew    = Abc_NtkRetimeWithClassesAig( pMan, vClasses, pvClasses, fVerbose );
    pNtkAigRet = Abc_NtkFromDarSeqSweep( pNtkAig, pManNew );
    pNtkRes    = Abc_NtkToLogic( pNtkAigRet );
    Abc_NtkDelete( pNtkAigRet );
    Abc_NtkDelete( pNtkAig );
    Aig_ManStop( pManNew );
    Aig_ManStop( pMan );
    return pNtkRes;
}

/**Function*************************************************************

  Synopsis    [Returns self-loops back into the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTransformBack(pNtkOld *Abc_Ntk_t, pNtkNew *Abc_Ntk_t, vControls *Vec_Ptr_t, vClasses *Vec_Int_t) {
    var pObj, pNodeNew, pCtrl, pDriver *Abc_Obj_t;
    var i, Class int32;
    assert(Abc_NtkPoNum(pNtkOld) == Abc_NtkPoNum(pNtkNew), "Abc_NtkPoNum(pNtkOld) == Abc_NtkPoNum(pNtkNew)");
    // match the POs of the old into new
    Abc_NtkForEachPo( pNtkOld, pObj, i )
        pObj.pCopy = Abc_NtkPo( pNtkNew, i );
    // remap the flops
    Vec_PtrForEachEntry( Abc_Obj_t *, vControls, pObj, i )
    {
        assert(Abc_ObjIsPo(pObj) && pObj.pNtk == pNtkOld, "Abc_ObjIsPo(pObj) && pObj.pNtk == pNtkOld");
        Vec_PtrWriteEntry( vControls, i, pObj.pCopy );
    }
    // create self-loops
    assert(Abc_NtkLatchNum(pNtkNew) == Vec_IntSize(vClasses), "Abc_NtkLatchNum(pNtkNew) == Vec_IntSize(vClasses)");
    Abc_NtkForEachLatch( pNtkNew, pObj, i )
    {
        Class = Vec_IntEntry( vClasses, i );
        if Class == -1 {
            continue; }
        pDriver = Abc_ObjFanin0(Abc_ObjFanin0(pObj));
        pCtrl = (Abc_Obj_t *)Vec_PtrEntry( vControls, Class );
        pCtrl = Abc_ObjFanin0( pCtrl );
        pNodeNew = Abc_NtkCreateNode( pNtkNew );
        Abc_ObjAddFanin( pNodeNew, pCtrl );
        Abc_ObjAddFanin( pNodeNew, pDriver );
        Abc_ObjAddFanin( pNodeNew, Abc_ObjFanout0(pObj) );
        Abc_ObjSetData( pNodeNew, Abc_SopRegister((Mem_Flex_t *)pNtkNew.pManFunc, "0-1 1\n11- 1\n") );
        Abc_ObjPatchFanin( Abc_ObjFanin0(pObj), pDriver, pNodeNew );
    }
    // remove the useless POs
    Vec_PtrForEachEntry( Abc_Obj_t *, vControls, pObj, i )
        Abc_NtkDeleteObj( pObj );
}

/**Function*************************************************************

  Synopsis    [Classify flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCRetime(pNtk *Abc_Ntk_t, fVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var vControls *Vec_Ptr_t;
    var vFlopClasses, vFlopClassesNew *Vec_Int_t;
    var pObj, pDriver, pFlopOut, pObjPo *Abc_Obj_t;
    var i, iFlop, CountN, Count2, Count1, Count0 int32;

    // duplicate the AIG
    pNtk = Abc_NtkDup( pNtk );

    // update registers
    vControls    = Vec_PtrAlloc( 100 );
    vFlopClasses = Vec_IntAlloc( 100 );
    Abc_NtkForEachLatch( pNtk, pObj, i )
    {
        pFlopOut = Abc_ObjFanout0(pObj);
        pDriver = Abc_ObjFanin0( Abc_ObjFanin0(pObj) );
        if Abc_ObjFaninNum(pDriver) != 3 {
            Vec_IntPush( vFlopClasses, -1 );
            CountN++;
            continue;
        }
        if Abc_ObjFanin(pDriver, 1) != pFlopOut && Abc_ObjFanin(pDriver, 2) != pFlopOut {
            Vec_IntPush( vFlopClasses, -1 );
            Count2++;
            continue;
        }
        if Abc_ObjFanin(pDriver, 1) == pFlopOut {
            Vec_IntPush( vFlopClasses, -1 );
            Count1++;
            continue;
        }
        assert(Abc_ObjFanin(pDriver, 2) == pFlopOut, "Abc_ObjFanin(pDriver, 2) == pFlopOut");
        Count0++;
        Vec_PtrPushUnique( vControls, Abc_ObjFanin0(pDriver) );
        // set the flop class
        iFlop = Vec_PtrFind( vControls, Abc_ObjFanin0(pDriver) );
        Vec_IntPush( vFlopClasses, iFlop );
        // update
        Abc_ObjPatchFanin( Abc_ObjFanin0(pObj), pDriver, Abc_ObjFanin(pDriver, 1) );
    }
    if Count1 {
        fmt.Printf("Opposite phase enable is present in %d flops (out of %d).\n", Count1, Abc_NtkLatchNum(pNtk)); }
    if fVerbose != 0 {
    fmt.Printf("CountN = %4d. Count2 = %4d. Count1 = %4d. Count0 = %4d. Ctrls = %d.\n", CountN, Count2, Count1, Count0, Vec_PtrSize(vControls)); }

    // add the controls to the list of POs
    Vec_PtrForEachEntry( Abc_Obj_t *, vControls, pObj, i )
    {
        pObjPo = Abc_NtkCreatePo( pNtk );
        Abc_ObjAddFanin( pObjPo, pObj );
        Abc_ObjAssignName( pObjPo, Abc_ObjName(pObjPo), nil );
        Vec_PtrWriteEntry( vControls, i, pObjPo );
    }
    Abc_NtkOrderCisCos( pNtk );
    Abc_NtkCleanup( pNtk, fVerbose );

    // performs retiming with classes
    pNtkNew = Abc_NtkRetimeWithClassesNtk( pNtk, vFlopClasses, &vFlopClassesNew, fVerbose );
    Abc_NtkTransformBack( pNtk, pNtkNew, vControls, vFlopClassesNew );
//    assert( Abc_NtkPoNum(pNtkNew) == Abc_NtkPoNum(pNtk) );
    Abc_NtkDelete( pNtk );

    Vec_PtrFree( vControls );
//    Vec_IntFree( vFlopClasses );
    Vec_IntFree( vFlopClassesNew );
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Resimulates CEX and return the ID of the PO that failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVerifyCex(pNtk *Abc_Ntk_t, p *Abc_Cex_t) int32 {
    var pObj *Abc_Obj_t;
    var RetValue, i, k, iBit int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(p.nPis  == Abc_NtkPiNum(pNtk), "p.nPis  == Abc_NtkPiNum(pNtk)");
//    assert( p->nRegs == Abc_NtkLatchNum(pNtk) );
    Abc_NtkCleanMarkC( pNtk );
    Abc_AigConst1(pNtk).fMarkC = 1;
    // initialize flops
    Abc_NtkForEachLatch( pNtk, pObj, i )
        Abc_ObjFanout0(pObj).fMarkC = Abc_InfoHasBit(p.pData, iBit++);
    // simulate timeframes
    iBit = p.nRegs;
    for  i = 0; i <= p.iFrame; i++  {
        Abc_NtkForEachPi( pNtk, pObj, k )
            pObj.fMarkC = Abc_InfoHasBit(p.pData, iBit++);
        Abc_NtkForEachNode( pNtk, pObj, k )
            pObj.fMarkC = (Abc_ObjFanin0(pObj).fMarkC ^ Abc_ObjFaninC0(pObj)) & 
                           (Abc_ObjFanin1(pObj).fMarkC ^ Abc_ObjFaninC1(pObj));
        Abc_NtkForEachCo( pNtk, pObj, k )
            pObj.fMarkC = Abc_ObjFanin0(pObj).fMarkC ^ Abc_ObjFaninC0(pObj);
        Abc_NtkForEachLatch( pNtk, pObj, k )
            Abc_ObjFanout0(pObj).fMarkC = Abc_ObjFanin0(pObj).fMarkC;
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    // figure out the number of failed output
    RetValue = -1;
    Abc_NtkForEachPo( pNtk, pObj, i )
    {
        if pObj.fMarkC {
            RetValue = i;
            break;
        }
    }
    Abc_NtkCleanMarkC( pNtk );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

