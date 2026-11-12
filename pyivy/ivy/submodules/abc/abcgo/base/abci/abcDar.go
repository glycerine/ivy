package abci

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [abcDar.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [DAG-aware rewriting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcDar.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "base/main/main.h"
//#include "aig/gia/giaAig.h"
//#include "opt/dar/dar.h"
//#include "sat/cnf/cnf.h"
//#include "proof/fra/fra.h"
//#include "proof/fraig/fraig.h"
//#include "proof/int/int.h"
//#include "proof/dch/dch.h"
//#include "proof/ssw/ssw.h"
//#include "opt/cgt/cgt.h"
//#include "bdd/bbr/bbr.h"
//#include "aig/gia/gia.h"
//#include "proof/cec/cec.h"
//#include "opt/csw/csw.h"
//#include "proof/pdr/pdr.h"
//#include "sat/bmc/bmc.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjCompareById(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
    return Abc_ObjId(Abc_ObjRegular(*pp1)) - Abc_ObjId(Abc_ObjRegular(*pp2));
}

/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CollectTopOr_rec(pObj *Abc_Obj_t, vSuper *Vec_Ptr_t) {
    if Abc_ObjIsComplement(pObj) || !Abc_ObjIsNode(pObj) {
        Vec_PtrPush( vSuper, pObj );
        return;
    }
    // go through the branches
    Abc_CollectTopOr_rec( Abc_ObjChild0(pObj), vSuper );
    Abc_CollectTopOr_rec( Abc_ObjChild1(pObj), vSuper );
}
func Abc_CollectTopOr(pObj *Abc_Obj_t, vSuper *Vec_Ptr_t) {
    Vec_PtrClear( vSuper );
    if Abc_ObjIsComplement(pObj) {
        Abc_CollectTopOr_rec( Abc_ObjNot(pObj), vSuper );
        Vec_PtrUniqify( vSuper, (int (*)())Abc_ObjCompareById );
    
    } else {
        Vec_PtrPush( vSuper, Abc_ObjNot(pObj) ); }
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description [The returned map maps new PO IDs into old ones.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToDarBmc(pNtk *Abc_Ntk_t, pvMap **Vec_Int_t) *Aig_Man_t {
    var pMan *Aig_Man_t;
    var pObj, pTemp *Abc_Obj_t;
    var vDrivers *Vec_Ptr_t;
    var vSuper *Vec_Ptr_t;
    var i, k, nDontCares int32;

    // print warning about initial values
    nDontCares = 0;
    Abc_NtkForEachLatch( pNtk, pObj, i )
        if Abc_LatchIsInitDc(pObj) {
            Abc_LatchSetInit0(pObj);
            nDontCares++;
        }
    if nDontCares != 0 {
        Abc_Print( 1, "Warning: %d registers in this network have don't-care init values.\n", nDontCares );
        Abc_Print( 1, "The don't-care are assumed to be 0. The result may not verify.\n" );
        Abc_Print( 1, "Use command \"print_latch\" to see the init values of registers.\n" );
        Abc_Print( 1, "Use command \"zero\" to convert or \"init\" to change the values.\n" );
    }

    // collect the drivers
    vSuper   = Vec_PtrAlloc( 100 );
    vDrivers = Vec_PtrAlloc( 100 );
    if pvMap { 
    *pvMap   = Vec_IntAlloc( 100 ); }
    Abc_NtkForEachPo( pNtk, pObj, i )
    {
        if pNtk.nConstrs && i >= pNtk.nConstrs {
            Vec_PtrPush( vDrivers, Abc_ObjNot(Abc_ObjChild0(pObj)) );
            if pvMap {
            Vec_IntPush( *pvMap, i ); }
            continue;
        }
        Abc_CollectTopOr( Abc_ObjChild0(pObj), vSuper );
        Vec_PtrForEachEntry( Abc_Obj_t *, vSuper, pTemp, k )
        {
            Vec_PtrPush( vDrivers, pTemp );
            if pvMap {
            Vec_IntPush( *pvMap, i ); }
        }       
    }
    Vec_PtrFree( vSuper );

    // create network
    pMan = Aig_ManStart( Abc_NtkNodeNum(pNtk) + 100 );
    pMan.nConstrs = pNtk.nConstrs;
    pMan.nBarBufs = pNtk.nBarBufs;
    pMan.pName = Extra_UtilStrsav( pNtk.pName );
    pMan.pSpec = Extra_UtilStrsav( pNtk.pSpec );
    // transfer the pointers to the basic nodes
    Abc_AigConst1(pNtk).pCopy = (Abc_Obj_t *)Aig_ManConst1(pMan);
    Abc_NtkForEachCi( pNtk, pObj, i )
        pObj.pCopy = (Abc_Obj_t *)Aig_ObjCreateCi(pMan);
    // create flops
    Abc_NtkForEachLatch( pNtk, pObj, i )
        Abc_ObjFanout0(pObj).pCopy = Abc_ObjNotCond( Abc_ObjFanout0(pObj).pCopy, Abc_LatchIsInit1(pObj) );
    // copy internal nodes
    Abc_NtkForEachNode( pNtk, pObj, i )
        pObj.pCopy = (Abc_Obj_t *)Aig_And( pMan, (Aig_Obj_t *)Abc_ObjChild0Copy(pObj), (Aig_Obj_t *)Abc_ObjChild1Copy(pObj) );
    // create the POs
    Vec_PtrForEachEntry( Abc_Obj_t *, vDrivers, pTemp, k )
        Aig_ObjCreateCo( pMan, (Aig_Obj_t *)Abc_ObjNotCond(Abc_ObjRegular(pTemp).pCopy, !Abc_ObjIsComplement(pTemp)) );
    Vec_PtrFree( vDrivers );
    // create flops
    Abc_NtkForEachLatchInput( pNtk, pObj, i )
        Aig_ObjCreateCo( pMan, (Aig_Obj_t *)Abc_ObjNotCond(Abc_ObjChild0Copy(pObj), Abc_LatchIsInit1(Abc_ObjFanout0(pObj))) );

    // remove dangling nodes
    Aig_ManSetRegNum( pMan, Abc_NtkLatchNum(pNtk) );
    Aig_ManCleanup( pMan );
    if !Aig_ManCheck( pMan ) {
        Abc_Print( 1, "Abc_NtkToDarBmc: AIG check has failed.\n" );
        Aig_ManStop( pMan );
        return nil;
    }
    return pMan;
}


/**Function*************************************************************

  Synopsis    [Collects information about what flops have unknown values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFindDcLatches(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vUnknown *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i int32;
    vUnknown = Vec_IntStart( Abc_NtkLatchNum(pNtk) );
    Abc_NtkForEachLatch( pNtk, pObj, i )
        if Abc_LatchIsInitDc(pObj) {
            Vec_IntWriteEntry( vUnknown, i, 1 );
            Abc_LatchSetInit0(pObj);
        }
    return vUnknown;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description [Assumes that registers are ordered after PIs/POs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToDar(pNtk *Abc_Ntk_t, fExors int32, fRegisters int32) *Aig_Man_t {
    var vNodes *Vec_Ptr_t;
    var pMan *Aig_Man_t;
    var pObjNew *Aig_Obj_t;
    var pObj *Abc_Obj_t;
    var i, nNodes, nDontCares int32;
    // make sure the latches follow PIs/POs
    if fRegisters != 0 { 
        assert(Abc_NtkBoxNum(pNtk) == Abc_NtkLatchNum(pNtk), "Abc_NtkBoxNum(pNtk) == Abc_NtkLatchNum(pNtk)");
        Abc_NtkForEachCi( pNtk, pObj, i )
            if i < Abc_NtkPiNum(pNtk) {
                assert(Abc_ObjIsPi(pObj), "Abc_ObjIsPi(pObj)");
                if !Abc_ObjIsPi(pObj) {
                    Abc_Print( 1, "Abc_NtkToDar(): Temporary bug: The PI ordering is wrong!\n" ); }
            
            } else {
                assert(Abc_ObjIsBo(pObj), "Abc_ObjIsBo(pObj)"); }
        Abc_NtkForEachCo( pNtk, pObj, i )
            if i < Abc_NtkPoNum(pNtk) {
                assert(Abc_ObjIsPo(pObj), "Abc_ObjIsPo(pObj)");
                if !Abc_ObjIsPo(pObj) {
                    Abc_Print( 1, "Abc_NtkToDar(): Temporary bug: The PO ordering is wrong!\n" ); }
            
            } else {
                assert(Abc_ObjIsBi(pObj), "Abc_ObjIsBi(pObj)"); }
        // print warning about initial values
        nDontCares = 0;
        Abc_NtkForEachLatch( pNtk, pObj, i )
            if Abc_LatchIsInitDc(pObj) {
                Abc_LatchSetInit0(pObj);
                nDontCares++;
            }
        if nDontCares != 0 {
            Abc_Print( 1, "Warning: %d registers in this network have don't-care init values.\n", nDontCares );
            Abc_Print( 1, "The don't-care are assumed to be 0. The result may not verify.\n" );
            Abc_Print( 1, "Use command \"print_latch\" to see the init values of registers.\n" );
            Abc_Print( 1, "Use command \"zero\" to convert or \"init\" to change the values.\n" );
        }
    }
    // create the manager
    pMan = Aig_ManStart( Abc_NtkNodeNum(pNtk) + 100 );
    pMan.fCatchExor = fExors;
    pMan.nConstrs = pNtk.nConstrs;
    pMan.nBarBufs = pNtk.nBarBufs;
    pMan.pName = Extra_UtilStrsav( pNtk.pName );
    pMan.pSpec = Extra_UtilStrsav( pNtk.pSpec );
    // transfer the pointers to the basic nodes
    Abc_AigConst1(pNtk).pCopy = (Abc_Obj_t *)Aig_ManConst1(pMan);
    Abc_NtkForEachCi( pNtk, pObj, i )
    {
        pObj.pCopy = (Abc_Obj_t *)Aig_ObjCreateCi(pMan);
        // initialize logic level of the CIs
        ((Aig_Obj_t *)pObj.pCopy).Level = pObj.Level;
    }

    // complement the 1-values registers
    if fRegisters != 0 {
        Abc_NtkForEachLatch( pNtk, pObj, i )
            if ( Abc_LatchIsInit1(pObj) )
                Abc_ObjFanout0(pObj).pCopy = Abc_ObjNot(Abc_ObjFanout0(pObj).pCopy);
    }
    // perform the conversion of the internal nodes (assumes DFS ordering)
//    pMan->fAddStrash = 1;
    vNodes = Abc_NtkDfs( pNtk, 0 );
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
//    Abc_NtkForEachNode( pNtk, pObj, i )
    {
        pObj.pCopy = (Abc_Obj_t *)Aig_And( pMan, (Aig_Obj_t *)Abc_ObjChild0Copy(pObj), (Aig_Obj_t *)Abc_ObjChild1Copy(pObj) );
//        Abc_Print( 1, "%d->%d ", pObj->Id, ((Aig_Obj_t *)pObj->pCopy)->Id );
    }
    Vec_PtrFree( vNodes );
    pMan.fAddStrash = 0;
    // create the POs
    Abc_NtkForEachCo( pNtk, pObj, i )
        Aig_ObjCreateCo( pMan, (Aig_Obj_t *)Abc_ObjChild0Copy(pObj) );
    // complement the 1-valued registers
    Aig_ManSetRegNum( pMan, Abc_NtkLatchNum(pNtk) );
    if fRegisters != 0 {
        Aig_ManForEachLiSeq( pMan, pObjNew, i )
            if ( Abc_LatchIsInit1(Abc_ObjFanout0(Abc_NtkCo(pNtk,i))) )
                pObjNew.pFanin0 = Aig_Not(pObjNew.pFanin0); }
    // remove dangling nodes
    if Abc_NtkGetChoiceNum(pNtk) == 0 {
        nNodes = Aig_ManCleanup( pMan );
    } else {
        nNodes = 0;
    }
    if !(fExors != 0) && nNodes {
        Abc_Print( 1, "Abc_NtkToDar(): Unexpected %d dangling nodes when converting to AIG!\n", nNodes ); }
//Aig_ManDumpVerilog( pMan, "test.v" );
    // save the number of registers
    if fRegisters != 0 {
        Aig_ManSetRegNum( pMan, Abc_NtkLatchNum(pNtk) );
        pMan.vFlopNums = Vec_IntStartNatural( pMan.nRegs );
//        pMan->vFlopNums = NULL;
//        pMan->vOnehots = Abc_NtkConverLatchNamesIntoNumbers( pNtk );
        if pNtk.vOnehots {
            pMan.vOnehots = (Vec_Ptr_t *)Vec_VecDupInt( (Vec_Vec_t *)pNtk.vOnehots ); }
    }
    if !Aig_ManCheck( pMan ) {
        Abc_Print( 1, "Abc_NtkToDar: AIG check has failed.\n" );
        Aig_ManStop( pMan );
        return nil;
    }
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description [Assumes that registers are ordered after PIs/POs.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkToDarChoices(pNtk *Abc_Ntk_t) *Aig_Man_t {
    var pMan *Aig_Man_t;
    var pObj, pPrev, pFanin *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i int32;
    vNodes = Abc_AigDfs( pNtk, 0, 0 );
    // create the manager
    pMan = Aig_ManStart( Abc_NtkNodeNum(pNtk) + 100 );
    pMan.nConstrs = pNtk.nConstrs;
    pMan.nBarBufs = pNtk.nBarBufs;
    pMan.pName = Extra_UtilStrsav( pNtk.pName );
    pMan.pSpec = Extra_UtilStrsav( pNtk.pSpec );
    if Abc_NtkGetChoiceNum(pNtk) {
        pMan.pEquivs = ABC_ALLOC( Aig_Obj_t *, Abc_NtkObjNum(pNtk) );
        memset( pMan.pEquivs, 0, sizeof(Aig_Obj_t *) * Abc_NtkObjNum(pNtk) );
    }
    // transfer the pointers to the basic nodes
    Abc_AigConst1(pNtk).pCopy = (Abc_Obj_t *)Aig_ManConst1(pMan);
    Abc_NtkForEachCi( pNtk, pObj, i )
        pObj.pCopy = (Abc_Obj_t *)Aig_ObjCreateCi(pMan);
    // perform the conversion of the internal nodes (assumes DFS ordering)
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
    {
        pObj.pCopy = (Abc_Obj_t *)Aig_And( pMan, (Aig_Obj_t *)Abc_ObjChild0Copy(pObj), (Aig_Obj_t *)Abc_ObjChild1Copy(pObj) );
//        Abc_Print( 1, "%d->%d ", pObj->Id, ((Aig_Obj_t *)pObj->pCopy)->Id );
        if Abc_AigNodeIsChoice( pObj ) {
            for  pPrev = pObj, pFanin = (Abc_Obj_t *)pObj.pData; (pFanin != nil); pPrev = pFanin, pFanin = (Abc_Obj_t *)pFanin.pData  {
                Aig_ObjSetEquiv( pMan, (Aig_Obj_t *)pPrev.pCopy, (Aig_Obj_t *)pFanin.pCopy );
            }
//            Aig_ManCreateChoice( pIfMan, (Aig_Obj_t *)pNode->pCopy );
        }
    }
    Vec_PtrFree( vNodes );
    // create the POs
    Abc_NtkForEachCo( pNtk, pObj, i )
        Aig_ObjCreateCo( pMan, (Aig_Obj_t *)Abc_ObjChild0Copy(pObj) );
    // complement the 1-valued registers
    Aig_ManSetRegNum( pMan, 0 );
    if !Aig_ManCheck( pMan ) {
        Abc_Print( 1, "Abc_NtkToDar: AIG check has failed.\n" );
        Aig_ManStop( pMan );
        return nil;
    }
    return pMan;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromDar(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t;
    var pNtkNew *Abc_Ntk_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(pMan.nAsserts == 0, "pMan.nAsserts == 0");
//    assert( Aig_ManRegNum(pMan) == Abc_NtkLatchNum(pNtkOld) );
    // perform strashing
    pNtkNew = Abc_NtkStartFrom( pNtkOld, ABC_NTK_STRASH, ABC_FUNC_AIG );
    pNtkNew.nConstrs = pMan.nConstrs;
    pNtkNew.nBarBufs = pNtkOld.nBarBufs;
    // transfer the pointers to the basic nodes
    Aig_ManConst1(pMan).pData = Abc_AigConst1(pNtkNew);
    Aig_ManForEachCi( pMan, pObj, i )
    {
        pObj.pData = Abc_NtkCi(pNtkNew, i);
        // initialize logic level of the CIs
        ((Abc_Obj_t *)pObj.pData).Level = pObj.Level;
    }
    // rebuild the AIG
    vNodes = Aig_ManDfs( pMan, 1 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        if ( Aig_ObjIsBuf(pObj) )
            pObj.pData = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj);
        else
            pObj.pData = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, (Abc_Obj_t *)Aig_ObjChild0Copy(pObj), (Abc_Obj_t *)Aig_ObjChild1Copy(pObj) );
    Vec_PtrFree( vNodes );
    // connect the PO nodes
    Aig_ManForEachCo( pMan, pObj, i )
    {
        if pMan.nAsserts && i == Aig_ManCoNum(pMan) - pMan.nAsserts {
            break; }
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), (Abc_Obj_t *)Aig_ObjChild0Copy(pObj) );
    }
    // if there are assertions, add them
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkFromDar(): Network check has failed.\n" ); }
//Abc_NtkPrintCiLevels( pNtkNew );
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description [This procedure should be called after seq sweeping, 
  which changes the number of registers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromDarSeqSweep(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t; 
    var pNtkNew *Abc_Ntk_t;
    var pObjNew, pLatch *Abc_Obj_t;
    var pObj, pObjLo, pObjLi *Aig_Obj_t;
    var i, iNodeId, nDigits int32; 
    assert(pMan.nAsserts == 0, "pMan.nAsserts == 0");
    assert(pNtkOld.nBarBufs == 0, "pNtkOld.nBarBufs == 0");
//    assert( Aig_ManRegNum(pMan) != Abc_NtkLatchNum(pNtkOld) );
    // perform strashing
    pNtkNew = Abc_NtkStartFromNoLatches( pNtkOld, ABC_NTK_STRASH, ABC_FUNC_AIG );
    pNtkNew.nConstrs = pMan.nConstrs;
    pNtkNew.nBarBufs = pMan.nBarBufs;
    // consider the case of target enlargement
    if Abc_NtkCiNum(pNtkNew) < Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan) {
        for  i = Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan) - Abc_NtkCiNum(pNtkNew); i > 0; i--  {
            pObjNew = Abc_NtkCreatePi( pNtkNew );
            Abc_ObjAssignName( pObjNew, Abc_ObjName(pObjNew), nil );
        }
        Abc_NtkOrderCisCos( pNtkNew );
    }
    assert(Abc_NtkCiNum(pNtkNew) == Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan), "Abc_NtkCiNum(pNtkNew) == Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan)");
    assert(Abc_NtkCoNum(pNtkNew) == Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan), "Abc_NtkCoNum(pNtkNew) == Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan)");
    // transfer the pointers to the basic nodes
    Aig_ManConst1(pMan).pData = Abc_AigConst1(pNtkNew);
    Aig_ManForEachPiSeq( pMan, pObj, i )
        pObj.pData = Abc_NtkCi(pNtkNew, i);
    // create as many latches as there are registers in the manager
    Aig_ManForEachLiLoSeq( pMan, pObjLi, pObjLo, i )
    {
        pObjNew = Abc_NtkCreateLatch( pNtkNew );
        pObjLi.pData = Abc_NtkCreateBi( pNtkNew );
        pObjLo.pData = Abc_NtkCreateBo( pNtkNew );
        Abc_ObjAddFanin( pObjNew, (Abc_Obj_t *)pObjLi.pData );
        Abc_ObjAddFanin( (Abc_Obj_t *)pObjLo.pData, pObjNew );
        Abc_LatchSetInit0( pObjNew );
    }
    // rebuild the AIG
    vNodes = Aig_ManDfs( pMan, 1 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        if ( Aig_ObjIsBuf(pObj) )
            pObj.pData = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj);
        else
            pObj.pData = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, (Abc_Obj_t *)Aig_ObjChild0Copy(pObj), (Abc_Obj_t *)Aig_ObjChild1Copy(pObj) );
    Vec_PtrFree( vNodes );
    // connect the PO nodes
    Aig_ManForEachCo( pMan, pObj, i )
    {
//        if ( pMan->nAsserts && i == Aig_ManCoNum(pMan) - pMan->nAsserts )
//            break;
        iNodeId = Nm_ManFindIdByNameTwoTypes( pNtkNew.pManName, Abc_ObjName(Abc_NtkCo(pNtkNew, i)), ABC_OBJ_PI, ABC_OBJ_BO );
        if iNodeId >= 0 {
            pObjNew = Abc_NtkObj( pNtkNew, iNodeId );
        } else {
            pObjNew = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj); }
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), pObjNew );
    }
    if pMan.vFlopNums == nil {
        Abc_NtkAddDummyBoxNames( pNtkNew );
    } else {
/*
        {
            int i, k, iFlop, Counter = 0;
            FILE * pFile;
            pFile = fopen( "out.txt", "w" );
            fAbc_Print( 1, pFile, "The total of %d registers were removed (out of %d):\n", 
                Abc_NtkLatchNum(pNtkOld)-Vec_IntSize(pMan->vFlopNums), Abc_NtkLatchNum(pNtkOld) );
            for ( i = 0; i < Abc_NtkLatchNum(pNtkOld); i++ )
            {
                Vec_IntForEachEntry( pMan->vFlopNums, iFlop, k )
                {
                    if ( i == iFlop )
                        break;
                }
                if ( k == Vec_IntSize(pMan->vFlopNums) )
                    fAbc_Print( 1, pFile, "%6d (%6d)  :  %s\n", ++Counter, i, Abc_ObjName( Abc_ObjFanout0(Abc_NtkBox(pNtkOld, i)) ) );
            }
            fclose( pFile );
            //Abc_Print( 1, "\n" );
        }
*/
        assert(Abc_NtkBoxNum(pNtkOld) == Abc_NtkLatchNum(pNtkOld), "Abc_NtkBoxNum(pNtkOld) == Abc_NtkLatchNum(pNtkOld)");
        nDigits = Abc_Base10Log( Abc_NtkLatchNum(pNtkNew) );
        Abc_NtkForEachLatch( pNtkNew, pObjNew, i )
        {
            pLatch = Abc_NtkBox( pNtkOld, Vec_IntEntry( pMan.vFlopNums, i ) );
            iNodeId = Nm_ManFindIdByName( pNtkNew.pManName, Abc_ObjName(Abc_ObjFanout0(pLatch)), ABC_OBJ_PO );
            if iNodeId >= 0 {
                Abc_ObjAssignName( pObjNew, Abc_ObjNameDummy("l", i, nDigits), nil );
                Abc_ObjAssignName( Abc_ObjFanin0(pObjNew), Abc_ObjNameDummy("li", i, nDigits), nil );
                Abc_ObjAssignName( Abc_ObjFanout0(pObjNew), Abc_ObjNameDummy("lo", i, nDigits), nil );
//Abc_Print( 1, "happening   %s -> %s\n", Abc_ObjName(Abc_ObjFanin0(pObjNew)), Abc_ObjName(Abc_ObjFanout0(pObjNew)) );
                continue;
            }
            Abc_ObjAssignName( pObjNew, Abc_ObjName(pLatch), nil );
            Abc_ObjAssignName( Abc_ObjFanin0(pObjNew),  Abc_ObjName(Abc_ObjFanin0(pLatch)), nil );
            Abc_ObjAssignName( Abc_ObjFanout0(pObjNew), Abc_ObjName(Abc_ObjFanout0(pLatch)), nil );
        }
    }
    // if there are assertions, add them
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkFromDar(): Network check has failed.\n" ); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description [This procedure should be called after seq sweeping, 
  which changes the number of registers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromAigPhase(pMan *Aig_Man_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t; 
    var pNtkNew *Abc_Ntk_t;
    var pObjNew *Abc_Obj_t;
    var pObj, pObjLo, pObjLi *Aig_Obj_t;
    var i int32; 
    assert(pMan.nAsserts == 0, "pMan.nAsserts == 0");
    // perform strashing
    pNtkNew = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtkNew.nConstrs = pMan.nConstrs;
    pNtkNew.nBarBufs = pMan.nBarBufs;
    // duplicate the name and the spec
//    pNtkNew->pName = Extra_UtilStrsav(pMan->pName);
//    pNtkNew->pSpec = Extra_UtilStrsav(pMan->pSpec);
    Aig_ManConst1(pMan).pData = Abc_AigConst1(pNtkNew);
    // create PIs
    Aig_ManForEachPiSeq( pMan, pObj, i )
    {
        pObjNew = Abc_NtkCreatePi( pNtkNew );
//        Abc_ObjAssignName( pObjNew, Abc_ObjName(pObjNew), NULL );
        pObj.pData = pObjNew;
    }
    // create POs
    Aig_ManForEachPoSeq( pMan, pObj, i )
    {
        pObjNew = Abc_NtkCreatePo( pNtkNew );
//        Abc_ObjAssignName( pObjNew, Abc_ObjName(pObjNew), NULL );
        pObj.pData = pObjNew;
    }
    assert(Abc_NtkCiNum(pNtkNew) == Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan), "Abc_NtkCiNum(pNtkNew) == Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan)");
    assert(Abc_NtkCoNum(pNtkNew) == Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan), "Abc_NtkCoNum(pNtkNew) == Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan)");
    // create as many latches as there are registers in the manager
    Aig_ManForEachLiLoSeq( pMan, pObjLi, pObjLo, i )
    {
        pObjNew = Abc_NtkCreateLatch( pNtkNew );
        pObjLi.pData = Abc_NtkCreateBi( pNtkNew );
        pObjLo.pData = Abc_NtkCreateBo( pNtkNew );
        Abc_ObjAddFanin( pObjNew, (Abc_Obj_t *)pObjLi.pData );
        Abc_ObjAddFanin( (Abc_Obj_t *)pObjLo.pData, pObjNew );
        Abc_LatchSetInit0( pObjNew );
//        Abc_ObjAssignName( (Abc_Obj_t *)pObjLi->pData, Abc_ObjName((Abc_Obj_t *)pObjLi->pData), NULL );
//        Abc_ObjAssignName( (Abc_Obj_t *)pObjLo->pData, Abc_ObjName((Abc_Obj_t *)pObjLo->pData), NULL );
    }
    // rebuild the AIG
    vNodes = Aig_ManDfs( pMan, 1 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        if ( Aig_ObjIsBuf(pObj) )
            pObj.pData = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj);
        else
            pObj.pData = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, (Abc_Obj_t *)Aig_ObjChild0Copy(pObj), (Abc_Obj_t *)Aig_ObjChild1Copy(pObj) );
    Vec_PtrFree( vNodes );
    // connect the PO nodes
    Aig_ManForEachCo( pMan, pObj, i )
    {
        pObjNew = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj);
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), pObjNew );
    }

    Abc_NtkAddDummyPiNames( pNtkNew );
    Abc_NtkAddDummyPoNames( pNtkNew );
    Abc_NtkAddDummyBoxNames( pNtkNew );

    // check the resulting AIG
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkFromAigPhase(): Network check has failed.\n" ); }
    return pNtkNew;
}



/**Function*************************************************************

  Synopsis    [Creates local function of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjHopFromGia_rec(pHopMan *Hop_Man_t, p *Gia_Man_t, Id int32, vCopies *Vec_Ptr_t) *Hop_Obj_t {
    var pObj *Gia_Obj_t;
    var gFunc, gFunc0, gFunc1 *Hop_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, Id) {
        return (Hop_Obj_t *)Vec_PtrEntry( vCopies, Id ); }
    Gia_ObjSetTravIdCurrentId(p, Id);
    pObj = Gia_ManObj(p, Id);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    // compute the functions of the children
    gFunc0 = Abc_ObjHopFromGia_rec( pHopMan, p, Gia_ObjFaninId0(pObj, Id), vCopies );
    gFunc1 = Abc_ObjHopFromGia_rec( pHopMan, p, Gia_ObjFaninId1(pObj, Id), vCopies );
    // get the function of the cut
    gFunc  = Hop_And( pHopMan, Hop_NotCond(gFunc0, Gia_ObjFaninC0(pObj)), Hop_NotCond(gFunc1, Gia_ObjFaninC1(pObj)) );  
    Vec_PtrWriteEntry( vCopies, Id, gFunc );
    return gFunc;
}
func Abc_ObjHopFromGia(pHopMan *Hop_Man_t, p *Gia_Man_t, GiaId int32, vCopies *Vec_Ptr_t) *Hop_Obj_t {
    var k, iFan int32;
    assert(Gia_ObjIsLut(p, GiaId), "Gia_ObjIsLut(p, GiaId)");
    assert(Gia_ObjLutSize(p, GiaId) > 0, "Gia_ObjLutSize(p, GiaId) > 0");
    Gia_ManIncrementTravId( p );
    Gia_LutForEachFanin( p, GiaId, iFan, k )
    {
        Gia_ObjSetTravIdCurrentId(p, iFan);
        Vec_PtrWriteEntry( vCopies, iFan, Hop_IthVar(pHopMan, k) );
    }
    return Abc_ObjHopFromGia_rec( pHopMan, p, GiaId, vCopies );
}

/**Function*************************************************************

  Synopsis    [Converts the network from the mapped GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromMappedGia_rec(pNtkNew *Abc_Ntk_t, p *Gia_Man_t, iObj int32, fAddInv int32) *Abc_Obj_t {
    var pObjNew *Abc_Obj_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
    if Gia_ObjValue(pObj) >= 0 {
        pObjNew = Abc_NtkObj( pNtkNew, Gia_ObjValue(pObj) );
    } else {
        Abc_NtkFromMappedGia_rec( pNtkNew, p, Gia_ObjFaninId0(pObj, iObj), 0 );
        Abc_NtkFromMappedGia_rec( pNtkNew, p, Gia_ObjFaninId1(pObj, iObj), 0 );
        pObjNew = Abc_NtkCreateNode( pNtkNew );
        Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin0(pObj))) );
        Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin1(pObj))) );
        pObjNew.pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtkNew.pManFunc, 2, nil );
        if Gia_ObjFaninC0(pObj) {  Abc_SopComplementVar( (char *)pObjNew.pData, 0 ); }
        if Gia_ObjFaninC1(pObj) {  Abc_SopComplementVar( (char *)pObjNew.pData, 1 ); }
        pObj.Value = Abc_ObjId( pObjNew );
    }
    if fAddInv != 0 {
        pObjNew = Abc_NtkCreateNodeInv(pNtkNew, pObjNew); }
    return pObjNew;
}
func Abc_NtkFromMappedGia(p *Gia_Man_t, fFindEnables int32, fUseBuffs int32) *Abc_Ntk_t {
    var fVerbose int32;
    var fDuplicate int32;
    var pNtkNew *Abc_Ntk_t;
    var pObjNew *Abc_Obj_t;
    var pObjNewLi *Abc_Obj_t;
    var pObjNewLo *Abc_Obj_t;
    var pConst0 *Abc_Obj_t = nil;
    var pObj, pObjLi, pObjLo *Gia_Obj_t;
    var vReflect *Vec_Ptr_t;
    var i, k, iFan, nDupGates, nCountMux int32; 
    assert(Gia_ManHasMapping(p) || p.pMuxes || (fFindEnables != 0), "Gia_ManHasMapping(p) || p.pMuxes || fFindEnables");
    assert(!(fFindEnables != 0) || !p.pMuxes, "!fFindEnables || !p.pMuxes");
    pNtkNew = Abc_NtkAlloc( ABC_NTK_LOGIC, Gia_ManHasMapping(p) ? ABC_FUNC_AIG : ABC_FUNC_SOP, 1 );
    // duplicate the name and the spec
    pNtkNew.pName = Extra_UtilStrsav(p.pName);
    pNtkNew.pSpec = Extra_UtilStrsav(p.pSpec);
    Gia_ManFillValue( p );
    // create constant
    pConst0 = Abc_NtkCreateNodeConst0( pNtkNew );
    Gia_ManConst0(p).Value = Abc_ObjId(pConst0);
    // create PIs
    Gia_ManForEachPi( p, pObj, i )
        pObj.Value = Abc_ObjId( Abc_NtkCreatePi( pNtkNew ) );
    // create POs
    Gia_ManForEachPo( p, pObj, i )
        pObj.Value = Abc_ObjId( Abc_NtkCreatePo( pNtkNew ) );
    // create as many latches as there are registers in the manager
    Gia_ManForEachRiRo( p, pObjLi, pObjLo, i )
    {
        pObjNew = Abc_NtkCreateLatch( pNtkNew );
        pObjNewLi = Abc_NtkCreateBi( pNtkNew );
        pObjNewLo = Abc_NtkCreateBo( pNtkNew );
        Abc_ObjAddFanin( pObjNew, pObjNewLi );
        Abc_ObjAddFanin( pObjNewLo, pObjNew );
        pObjLi.Value = Abc_ObjId( pObjNewLi );
        pObjLo.Value = Abc_ObjId( pObjNewLo );
        Abc_LatchSetInit0( pObjNew );
    }
    // rebuild the AIG
    if fFindEnables != 0 {
        Gia_ManForEachCo( p, pObj, i )
        {
            pObjNew = nil;
            if Gia_ObjIsRi(p, pObj) && Gia_ObjIsMuxType(Gia_ObjFanin0(pObj)) {
                var iObjRo int32  = Gia_ObjRiToRoId( p, Gia_ObjId(p, pObj) );
                var iLitE int32;
                var iLitT int32;
                var iCtrl int32 = Gia_ObjRecognizeMuxLits( p, Gia_ObjFanin0(pObj), &iLitT, &iLitE );
                iLitE = Abc_LitNotCond( iLitE, Gia_ObjFaninC0(pObj) );
                iLitT = Abc_LitNotCond( iLitT, Gia_ObjFaninC0(pObj) );
                if Abc_Lit2Var(iLitT) == iObjRo {
                    var iTemp int32  = iLitE;
                    iLitE = iLitT;
                    iLitT = iTemp;
                    iCtrl = Abc_LitNot( iCtrl );
                }
                if Abc_Lit2Var(iLitE) == iObjRo {
                    var pObjCtrl *Abc_Obj_t   = Abc_NtkFromMappedGia_rec( pNtkNew, p, Abc_Lit2Var(iCtrl), Abc_LitIsCompl(iCtrl) );
                    var pObjNodeT *Abc_Obj_t  = Abc_NtkFromMappedGia_rec( pNtkNew, p, Abc_Lit2Var(iLitT), Abc_LitIsCompl(iLitT) );
                    var pObjNodeE *Abc_Obj_t  = Abc_NtkFromMappedGia_rec( pNtkNew, p, Abc_Lit2Var(iLitE), Abc_LitIsCompl(iLitE) );
                    pObjNew = Abc_NtkCreateNode( pNtkNew );
                    Abc_ObjAddFanin( pObjNew, pObjCtrl );
                    Abc_ObjAddFanin( pObjNew, pObjNodeT );
                    Abc_ObjAddFanin( pObjNew, pObjNodeE );
                    pObjNew.pData = Abc_SopCreateMux( (Mem_Flex_t *)pNtkNew.pManFunc );
                    nCountMux++;
                }
            }
            if pObjNew == nil {
                pObjNew = Abc_NtkFromMappedGia_rec( pNtkNew, p, Gia_ObjFaninId0p(p, pObj), Gia_ObjFaninC0(pObj) ); }
            Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), pObjNew );
        }
    
    } else if p.pMuxes {
        Gia_ManForEachAnd( p, pObj, i )
        {
            pObjNew = Abc_NtkCreateNode( pNtkNew );
            if Gia_ObjIsMuxId(p, i) {
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin2(p, pObj))) );
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin1(pObj))) );
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin0(pObj))) );
                pObjNew.pData = Abc_SopCreateMux( (Mem_Flex_t *)pNtkNew.pManFunc );
                if Gia_ObjFaninC2(p, pObj) {  Abc_SopComplementVar( (char *)pObjNew.pData, 0 ); }
                if Gia_ObjFaninC1(pObj) {     Abc_SopComplementVar( (char *)pObjNew.pData, 1 ); }
                if Gia_ObjFaninC0(pObj) {     Abc_SopComplementVar( (char *)pObjNew.pData, 2 ); }
            
            } else if Gia_ObjIsXor(pObj) {
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin0(pObj))) );
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin1(pObj))) );
                pObjNew.pData = Abc_SopCreateXor( (Mem_Flex_t *)pNtkNew.pManFunc, 2 );
                if Gia_ObjFaninC0(pObj) {  Abc_SopComplementVar( (char *)pObjNew.pData, 0 ); }
                if Gia_ObjFaninC1(pObj) {  Abc_SopComplementVar( (char *)pObjNew.pData, 1 ); }
            
            } else {
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin0(pObj))) );
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ObjFanin1(pObj))) );
                pObjNew.pData = Abc_SopCreateAnd( (Mem_Flex_t *)pNtkNew.pManFunc, 2, nil );
                if Gia_ObjFaninC0(pObj) {  Abc_SopComplementVar( (char *)pObjNew.pData, 0 ); }
                if Gia_ObjFaninC1(pObj) {  Abc_SopComplementVar( (char *)pObjNew.pData, 1 ); }
            }
            pObj.Value = Abc_ObjId( pObjNew );
        }
    
    } else {
        vReflect = Vec_PtrStart( Gia_ManObjNum(p) );
        Gia_ManForEachLut( p, i )
        {
            pObj = Gia_ManObj(p, i);
            assert(pObj.Value == ~0, "pObj.Value == ~0");
            if Gia_ObjLutSize(p, i) == 0 {
                pObj.Value = Abc_ObjId(pConst0);
                continue;
            }
            pObjNew = Abc_NtkCreateNode( pNtkNew );
            Gia_LutForEachFanin( p, i, iFan, k )
                Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtkNew, Gia_ObjValue(Gia_ManObj(p, iFan))) );
            pObjNew.pData = Abc_ObjHopFromGia( (Hop_Man_t *)pNtkNew.pManFunc, p, i, vReflect );
            pObj.Value = Abc_ObjId( pObjNew );
        }
        Vec_PtrFree( vReflect );
    }
    //if ( fFindEnables )
    //    printf( "Extracted %d flop enable signals.\n", nCountMux );
    // connect the PO nodes
    if ( !fFindEnables )
    Gia_ManForEachCo( p, pObj, i )
    {
        pObjNew = Abc_NtkObj( pNtkNew, Gia_ObjValue(Gia_ObjFanin0(pObj)) );
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), Abc_ObjNotCond( pObjNew, Gia_ObjFaninC0(pObj) ) );
    }
    // create names
    Abc_NtkAddDummyPiNames( pNtkNew );
    Abc_NtkAddDummyPoNames( pNtkNew );
    Abc_NtkAddDummyBoxNames( pNtkNew );

    // decouple the PO driver nodes to reduce the number of levels
    nDupGates = Abc_NtkLogicMakeSimpleCos( pNtkNew, !fUseBuffs );
    if (fVerbose != 0) && (nDupGates != 0) && !Abc_FrameReadFlag("silentmode") {
        if !(fDuplicate != 0) {
            fmt.Printf("Added %d buffers/inverters to decouple the CO drivers.\n", nDupGates);
        } else {
            fmt.Printf("Duplicated %d gates to decouple the CO drivers.\n", nDupGates); }
    }
    // remove const node if it is not used
    if !Abc_ObjIsNone(pConst0) && Abc_ObjFanoutNum(pConst0) == 0 {
        Abc_NtkDeleteObj( pConst0 ); }

    assert(Gia_ManPiNum(p) == Abc_NtkPiNum(pNtkNew), "Gia_ManPiNum(p) == Abc_NtkPiNum(pNtkNew)");
    assert(Gia_ManPoNum(p) == Abc_NtkPoNum(pNtkNew), "Gia_ManPoNum(p) == Abc_NtkPoNum(pNtkNew)");
    assert(Gia_ManRegNum(p) == Abc_NtkLatchNum(pNtkNew), "Gia_ManRegNum(p) == Abc_NtkLatchNum(pNtkNew)");

    // check the resulting AIG
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkFromMappedGia(): Network check has failed.\n" ); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the mapped GIA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromCellWrite(vCopyLits *Vec_Int_t, i int32, c int32, Id int32) {
    Vec_IntWriteEntry( vCopyLits, Abc_Var2Lit(i, c), Id );
}
func Abc_NtkFromCellRead(p *Abc_Ntk_t, vCopyLits *Vec_Int_t, i int32, c int32) *Abc_Obj_t {
    var pObjNew *Abc_Obj_t;
    var iObjNew int32  = Vec_IntEntry( vCopyLits, Abc_Var2Lit(i, c) );
    if iObjNew >= 0 {
        return Abc_NtkObj(p, iObjNew); }
    // opposite phase should be already constructed
    assert(false, "0");
    if i == 0 {
        if c != 0 {
            pObjNew = Abc_NtkCreateNodeConst1(p);
        } else {
            pObjNew = Abc_NtkCreateNodeConst0(p);
        }
    } else {
        iObjNew = Vec_IntEntry( vCopyLits, Abc_Var2Lit(i, !c) );   assert(iObjNew >= 0, "iObjNew >= 0");
        pObjNew = Abc_NtkCreateNodeInv( p, Abc_NtkObj(p, iObjNew) );
    }
    Abc_NtkFromCellWrite( vCopyLits, i, c, Abc_ObjId(pObjNew) );
    return pObjNew;
}
func Abc_NtkFromCellMappedGia(p *Gia_Man_t, fUseBuffs int32) *Abc_Ntk_t {
    var fFixDrivers int32  = 1;
    var fVerbose int32;
    var pNtkNew *Abc_Ntk_t;
    var vCopyLits *Vec_Int_t;
    var pObjNew, pObjNewLi, pObjNewLo *Abc_Obj_t;
    var pObj, pObjLi, pObjLo *Gia_Obj_t;
    var i int32;
    var k int32;
    var iLit int32;
    var iFanLit int32;
    var nCells int32;
    var fNeedConst [2]int32 ; 
    var pCells *Mio_Cell2_t  = Mio_CollectRootsNewDefault2( 6, &nCells, 0 );
    assert(Gia_ManHasCellMapping(p), "Gia_ManHasCellMapping(p)");
    // start network
    pNtkNew = Abc_NtkAlloc( ABC_NTK_LOGIC, ABC_FUNC_MAP, 1 );
    pNtkNew.pName = Extra_UtilStrsav(p.pName);
    pNtkNew.pSpec = Extra_UtilStrsav(p.pSpec);
    assert(pNtkNew.pManFunc == Abc_FrameReadLibGen(), "pNtkNew.pManFunc == Abc_FrameReadLibGen()");
    vCopyLits = Vec_IntStartFull( 2*Gia_ManObjNum(p) );
    // create PIs
    Gia_ManForEachPi( p, pObj, i )
        Abc_NtkFromCellWrite( vCopyLits, Gia_ObjId(p, pObj), 0, Abc_ObjId( Abc_NtkCreatePi( pNtkNew ) ) );
    // create POs
    Gia_ManForEachPo( p, pObj, i )
        Abc_NtkFromCellWrite( vCopyLits, Gia_ObjId(p, pObj), 0, Abc_ObjId( Abc_NtkCreatePo( pNtkNew ) ) );
    // create as many latches as there are registers in the manager
    Gia_ManForEachRiRo( p, pObjLi, pObjLo, i )
    {
        pObjNew = Abc_NtkCreateLatch( pNtkNew );
        pObjNewLi = Abc_NtkCreateBi( pNtkNew );
        pObjNewLo = Abc_NtkCreateBo( pNtkNew );
        Abc_ObjAddFanin( pObjNew, pObjNewLi );
        Abc_ObjAddFanin( pObjNewLo, pObjNew );
//        pObjLi->Value = Abc_ObjId( pObjNewLi );
//        pObjLo->Value = Abc_ObjId( pObjNewLo );
        Abc_NtkFromCellWrite( vCopyLits, Gia_ObjId(p, pObjLi), 0, Abc_ObjId( pObjNewLi ) );
        Abc_NtkFromCellWrite( vCopyLits, Gia_ObjId(p, pObjLo), 0, Abc_ObjId( pObjNewLo ) );
        Abc_LatchSetInit0( pObjNew );
    }

    // create constants
    Gia_ManForEachCo( p, pObj, i )
        if ( Gia_ObjFaninId0p(p, pObj) == 0 )
            fNeedConst[Gia_ObjFaninC0(pObj)] = 1;
    Gia_ManForEachBuf( p, pObj, i )
        if ( Gia_ObjFaninId0p(p, pObj) == 0 )
            fNeedConst[Gia_ObjFaninC0(pObj)] = 1;
    if fNeedConst[0] {
        Abc_NtkFromCellWrite( vCopyLits, 0, 0, Abc_ObjId(Abc_NtkCreateNodeConst0(pNtkNew)) ); }
    if fNeedConst[1] {
        Abc_NtkFromCellWrite( vCopyLits, 0, 1, Abc_ObjId(Abc_NtkCreateNodeConst1(pNtkNew)) ); }

    // rebuild the AIG
    Gia_ManForEachCell( p, iLit )
    {
        var fSkip int32;
        if Gia_ObjIsCellBuf(p, iLit) {
            assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
            // build buffer
            pObjNew = Abc_NtkCreateNode( pNtkNew );
            iFanLit = Gia_ObjFaninLit0p( p, Gia_ManObj(p, Abc_Lit2Var(iLit)) );
            Abc_ObjAddFanin( pObjNew, Abc_NtkFromCellRead(pNtkNew, vCopyLits, Abc_Lit2Var(iFanLit), Abc_LitIsCompl(iFanLit)) );
            pObjNew.pData = nil; // barrier buffer
            assert(Abc_ObjIsBarBuf(pObjNew), "Abc_ObjIsBarBuf(pObjNew)");
            pNtkNew.nBarBufs2++;
        
        } else if Gia_ObjIsCellInv(p, iLit) {
            var iLitNot int32  = Abc_LitNot(iLit);
            if !Abc_LitIsCompl(iLit) { // positive phase
            
                // build negative phase
                assert(Vec_IntEntry(vCopyLits, iLitNot) == -1, "Vec_IntEntry(vCopyLits, iLitNot) == -1");
                assert(Gia_ObjCellId(p, iLitNot) > 0, "Gia_ObjCellId(p, iLitNot) > 0");
                pObjNew = Abc_NtkCreateNode( pNtkNew );
                Gia_CellForEachFanin( p, iLitNot, iFanLit, k )
                    Abc_ObjAddFanin( pObjNew, Abc_NtkFromCellRead(pNtkNew, vCopyLits, Abc_Lit2Var(iFanLit), Abc_LitIsCompl(iFanLit)) );
                pObjNew.pData = Mio_LibraryReadGateByName( (Mio_Library_t *)pNtkNew.pManFunc, pCells[Gia_ObjCellId(p, iLitNot)].pName, nil );
                Abc_NtkFromCellWrite( vCopyLits, Abc_Lit2Var(iLitNot), Abc_LitIsCompl(iLitNot), Abc_ObjId(pObjNew) );
                fSkip = 1;
            
            } else { // negative phase
            
                // positive phase is available
                assert(Vec_IntEntry(vCopyLits, iLitNot) != -1, "Vec_IntEntry(vCopyLits, iLitNot) != -1");
            }
            // build inverter
            pObjNew = Abc_NtkCreateNode( pNtkNew );
            Abc_ObjAddFanin( pObjNew, Abc_NtkFromCellRead(pNtkNew, vCopyLits, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLitNot)) );
            pObjNew.pData = Mio_LibraryReadGateByName( (Mio_Library_t *)pNtkNew.pManFunc, pCells[3].pName, nil );
        
        } else {
            assert(Gia_ObjCellId(p, iLit) >= 0, "Gia_ObjCellId(p, iLit) >= 0");
            pObjNew = Abc_NtkCreateNode( pNtkNew );
            Gia_CellForEachFanin( p, iLit, iFanLit, k )
                Abc_ObjAddFanin( pObjNew, Abc_NtkFromCellRead(pNtkNew, vCopyLits, Abc_Lit2Var(iFanLit), Abc_LitIsCompl(iFanLit)) );
            pObjNew.pData = Mio_LibraryReadGateByName( (Mio_Library_t *)pNtkNew.pManFunc, pCells[Gia_ObjCellId(p, iLit)].pName, nil );
        }
        assert(Vec_IntEntry(vCopyLits, iLit) == -1, "Vec_IntEntry(vCopyLits, iLit) == -1");
        Abc_NtkFromCellWrite( vCopyLits, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLit), Abc_ObjId(pObjNew) );
        // skip next
        iLit += fSkip;
    }

    // connect the PO nodes
    Gia_ManForEachCo( p, pObj, i )
    {
        pObjNew = Abc_NtkFromCellRead( pNtkNew, vCopyLits, Gia_ObjFaninId0p(p, pObj), Gia_ObjFaninC0(pObj) );
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), pObjNew );
    }
    // create names
    Abc_NtkAddDummyPiNames( pNtkNew );
    Abc_NtkAddDummyPoNames( pNtkNew );
    Abc_NtkAddDummyBoxNames( pNtkNew );

    // decouple the PO driver nodes to reduce the number of levels
    if fFixDrivers != 0 {
        var nDupGates int32  = Abc_NtkLogicMakeSimpleCos( pNtkNew, !fUseBuffs );
        if (fVerbose != 0) && (nDupGates != 0) && !Abc_FrameReadFlag("silentmode") {
            if fUseBuffs != 0 {
                fmt.Printf("Added %d buffers/inverters to decouple the CO drivers.\n", nDupGates);
            } else {
                fmt.Printf("Duplicated %d gates to decouple the CO drivers.\n", nDupGates); }
        }
    }

    assert(Gia_ManPiNum(p) == Abc_NtkPiNum(pNtkNew), "Gia_ManPiNum(p) == Abc_NtkPiNum(pNtkNew)");
    assert(Gia_ManPoNum(p) == Abc_NtkPoNum(pNtkNew), "Gia_ManPoNum(p) == Abc_NtkPoNum(pNtkNew)");
    assert(Gia_ManRegNum(p) == Abc_NtkLatchNum(pNtkNew), "Gia_ManRegNum(p) == Abc_NtkLatchNum(pNtkNew)");
    Vec_IntFree( vCopyLits );
    ABC_FREE( pCells );

    // check the resulting AIG
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkFromMappedGia(): Network check has failed.\n" ); }
    return pNtkNew;
}


/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description [This procedure should be called after seq sweeping, 
  which changes the number of registers.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAfterTrim(pMan *Aig_Man_t, pNtkOld *Abc_Ntk_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t; 
    var pNtkNew *Abc_Ntk_t;
    var pObjNew, pObjOld *Abc_Obj_t;
    var pObj, pObjLo, pObjLi *Aig_Obj_t;
    var i int32; 
    assert(pMan.nAsserts == 0, "pMan.nAsserts == 0");
    assert(pNtkOld.nBarBufs == 0, "pNtkOld.nBarBufs == 0");
    assert(Aig_ManRegNum(pMan) <= Abc_NtkLatchNum(pNtkOld), "Aig_ManRegNum(pMan) <= Abc_NtkLatchNum(pNtkOld)");
    assert(Saig_ManPiNum(pMan) <= Abc_NtkCiNum(pNtkOld), "Saig_ManPiNum(pMan) <= Abc_NtkCiNum(pNtkOld)");
    assert(Saig_ManPoNum(pMan) == Abc_NtkPoNum(pNtkOld), "Saig_ManPoNum(pMan) == Abc_NtkPoNum(pNtkOld)");
    assert(pMan.vCiNumsOrig != nil, "pMan.vCiNumsOrig != nil");
    // perform strashing
    pNtkNew = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtkNew.nConstrs = pMan.nConstrs;
    pNtkNew.nBarBufs = pMan.nBarBufs;
    // duplicate the name and the spec
//    pNtkNew->pName = Extra_UtilStrsav(pMan->pName);
//    pNtkNew->pSpec = Extra_UtilStrsav(pMan->pSpec);
    Aig_ManConst1(pMan).pData = Abc_AigConst1(pNtkNew);
    // create PIs
    Aig_ManForEachPiSeq( pMan, pObj, i )
    {
        pObjNew = Abc_NtkCreatePi( pNtkNew );
        pObj.pData = pObjNew;
        // find the name
        pObjOld = Abc_NtkCi( pNtkOld, Vec_IntEntry(pMan.vCiNumsOrig, i) );
        Abc_ObjAssignName( pObjNew, Abc_ObjName(pObjOld), nil );
    }
    // create POs
    Aig_ManForEachPoSeq( pMan, pObj, i )
    {
        pObjNew = Abc_NtkCreatePo( pNtkNew );
        pObj.pData = pObjNew;
        // find the name
        pObjOld = Abc_NtkCo( pNtkOld, i );
        Abc_ObjAssignName( pObjNew, Abc_ObjName(pObjOld), nil );
    }
    assert(Abc_NtkCiNum(pNtkNew) == Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan), "Abc_NtkCiNum(pNtkNew) == Aig_ManCiNum(pMan) - Aig_ManRegNum(pMan)");
    assert(Abc_NtkCoNum(pNtkNew) == Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan), "Abc_NtkCoNum(pNtkNew) == Aig_ManCoNum(pMan) - Aig_ManRegNum(pMan)");
    // create as many latches as there are registers in the manager
    Aig_ManForEachLiLoSeq( pMan, pObjLi, pObjLo, i )
    {
        pObjNew = Abc_NtkCreateLatch( pNtkNew );
        pObjLi.pData = Abc_NtkCreateBi( pNtkNew );
        pObjLo.pData = Abc_NtkCreateBo( pNtkNew );
        Abc_ObjAddFanin( pObjNew, (Abc_Obj_t *)pObjLi.pData );
        Abc_ObjAddFanin( (Abc_Obj_t *)pObjLo.pData, pObjNew );
        Abc_LatchSetInit0( pObjNew );
        // find the name
        pObjOld = Abc_NtkCi( pNtkOld, Vec_IntEntry(pMan.vCiNumsOrig, Saig_ManPiNum(pMan)+i) );
        Abc_ObjAssignName( (Abc_Obj_t *)pObjLo.pData, Abc_ObjName(pObjOld), nil );
        // find the name
        pObjOld = Abc_NtkCo( pNtkOld, Saig_ManPoNum(pMan)+i );
        Abc_ObjAssignName( (Abc_Obj_t *)pObjLi.pData, Abc_ObjName(pObjOld), nil );
    }
    // rebuild the AIG
    vNodes = Aig_ManDfs( pMan, 1 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
        if ( Aig_ObjIsBuf(pObj) )
            pObj.pData = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj);
        else
            pObj.pData = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, (Abc_Obj_t *)Aig_ObjChild0Copy(pObj), (Abc_Obj_t *)Aig_ObjChild1Copy(pObj) );
    Vec_PtrFree( vNodes );
    // connect the PO nodes
    Aig_ManForEachCo( pMan, pObj, i )
    {
        pObjNew = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj);
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), pObjNew );
    }
    // check the resulting AIG
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkAfterTrim(): Network check has failed.\n" ); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromDarChoices(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj, pTemp *Aig_Obj_t;
    var i int32;
    assert(pMan.pEquivs != nil, "pMan.pEquivs != nil");
    assert(Aig_ManBufNum(pMan) == 0, "Aig_ManBufNum(pMan) == 0");
    // perform strashing
    pNtkNew = Abc_NtkStartFrom( pNtkOld, ABC_NTK_STRASH, ABC_FUNC_AIG );
    pNtkNew.nConstrs = pMan.nConstrs;
    pNtkNew.nBarBufs = pNtkOld.nBarBufs;
    // transfer the pointers to the basic nodes
    Aig_ManCleanData( pMan );
    Aig_ManConst1(pMan).pData = Abc_AigConst1(pNtkNew);
    Aig_ManForEachCi( pMan, pObj, i )
        pObj.pData = Abc_NtkCi(pNtkNew, i);
    // rebuild the AIG
    Aig_ManForEachNode( pMan, pObj, i )
    {
        pObj.pData = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, (Abc_Obj_t *)Aig_ObjChild0Copy(pObj), (Abc_Obj_t *)Aig_ObjChild1Copy(pObj) );
        if pTemp = Aig_ObjEquiv(pMan, pObj) {
            assert(pTemp.pData != nil, "pTemp.pData != nil");
            ((Abc_Obj_t *)pObj.pData).pData = ((Abc_Obj_t *)pTemp.pData);
        }
    }
    // connect the PO nodes
    Aig_ManForEachCo( pMan, pObj, i )
        Abc_ObjAddFanin( Abc_NtkCo(pNtkNew, i), (Abc_Obj_t *)Aig_ObjChild0Copy(pObj) );
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkFromDar(): Network check has failed.\n" ); }

    // verify topological order
    if false {
        var pNode *Abc_Obj_t;
        Abc_NtkForEachNode( pNtkNew, pNode, i )
            if Abc_AigNodeIsChoice( pNode ) {
                var Counter int32;
                for  pNode = Abc_ObjEquiv(pNode); (pNode != nil); pNode = Abc_ObjEquiv(pNode)  {
                    Counter++;
                }
                fmt.Printf("%d ", Counter);
            }
        fmt.Printf("\n");
    }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Converts the network from the AIG manager into ABC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkFromDarSeq(pNtkOld *Abc_Ntk_t, pMan *Aig_Man_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t;
    var pNtkNew *Abc_Ntk_t;
    var pObjNew, pFaninNew, pFaninNew0, pFaninNew1 *Abc_Obj_t;
    var pObj *Aig_Obj_t;
    var i int32;
//    assert( Aig_ManLatchNum(pMan) > 0 );
    assert(pNtkOld.nBarBufs == 0, "pNtkOld.nBarBufs == 0");
    // perform strashing
    pNtkNew = Abc_NtkStartFromNoLatches( pNtkOld, ABC_NTK_STRASH, ABC_FUNC_AIG );
    pNtkNew.nConstrs = pMan.nConstrs;
    pNtkNew.nBarBufs = pMan.nBarBufs;
    // transfer the pointers to the basic nodes
    Aig_ManConst1(pMan).pData = Abc_AigConst1(pNtkNew);
    Aig_ManForEachCi( pMan, pObj, i )
        pObj.pData = Abc_NtkPi(pNtkNew, i);
    // create latches of the new network
    Aig_ManForEachObj( pMan, pObj, i )
    {
        pObjNew = Abc_NtkCreateLatch( pNtkNew );
        pFaninNew0 = Abc_NtkCreateBi( pNtkNew );
        pFaninNew1 = Abc_NtkCreateBo( pNtkNew );
        Abc_ObjAddFanin( pObjNew, pFaninNew0 );
        Abc_ObjAddFanin( pFaninNew1, pObjNew );
        Abc_LatchSetInit0( pObjNew );
        pObj.pData = Abc_ObjFanout0( pObjNew );
    }
    Abc_NtkAddDummyBoxNames( pNtkNew );
    // rebuild the AIG
    vNodes = Aig_ManDfs( pMan, 1 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        // add the first fanin
        pObj.pData = pFaninNew0 = (Abc_Obj_t *)Aig_ObjChild0Copy(pObj);
        if Aig_ObjIsBuf(pObj) {
            continue; }
        // add the second fanin
        pFaninNew1 = (Abc_Obj_t *)Aig_ObjChild1Copy(pObj);
        // create the new node
        if Aig_ObjIsExor(pObj) {
            pObj.pData = pObjNew = Abc_AigXor( (Abc_Aig_t *)pNtkNew.pManFunc, pFaninNew0, pFaninNew1 );
        } else {
            pObj.pData = pObjNew = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, pFaninNew0, pFaninNew1 ); }
    }
    Vec_PtrFree( vNodes );
    // connect the PO nodes
    Aig_ManForEachCo( pMan, pObj, i )
    {
        pFaninNew = (Abc_Obj_t *)Aig_ObjChild0Copy( pObj );
        Abc_ObjAddFanin( Abc_NtkPo(pNtkNew, i), pFaninNew );
    }
    // connect the latches
    Aig_ManForEachObj( pMan, pObj, i )
    {
        pFaninNew = (Abc_Obj_t *)Aig_ObjChild0Copy( pObj );
        Abc_ObjAddFanin( Abc_ObjFanin0(Abc_ObjFanin0((Abc_Obj_t *)pObj.pData)), pFaninNew );
    }
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkFromIvySeq(): Network check has failed.\n" ); }
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Collects CI of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectCiNames(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var pObj *Abc_Obj_t;
    var i int32;
    var vNames *Vec_Ptr_t;
    vNames = Vec_PtrAlloc( 100 );
    Abc_NtkForEachCi( pNtk, pObj, i )
        Vec_PtrPush( vNames, Extra_UtilStrsav(Abc_ObjName(pObj)) );
    return vNames;
}

/**Function*************************************************************

  Synopsis    [Collects CO of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectCoNames(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var pObj *Abc_Obj_t;
    var i int32;
    var vNames *Vec_Ptr_t;
    vNames = Vec_PtrAlloc( 100 );
    Abc_NtkForEachCo( pNtk, pObj, i )
        Vec_PtrPush( vNames, Extra_UtilStrsav(Abc_ObjName(pObj)) );
    return vNames;
}

/**Function*************************************************************

  Synopsis    [Collect latch values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkGetLatchValues(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vInits *Vec_Int_t;
    var pLatch *Abc_Obj_t;
    var i int32;
    vInits = Vec_IntAlloc( Abc_NtkLatchNum(pNtk) );
    Abc_NtkForEachLatch( pNtk, pLatch, i )
    {
        if Abc_LatchIsInit0(pLatch) {
            Vec_IntPush( vInits, 0 );
        } else if Abc_LatchIsInit1(pLatch) {
            Vec_IntPush( vInits, 1 );
        } else if Abc_LatchIsInitDc(pLatch) {
            Vec_IntPush( vInits, 2 );
        } else {
            assert(false, "0"); }
    }
    return vInits;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDar(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t  = nil;
    var pMan *Aig_Man_t;
    extern void Fra_ManPartitionTest( Aig_Man_t * p, int nComLim );

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // convert to the AIG manager
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }

    // perform computation
//    Fra_ManPartitionTest( pMan, 4 );
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );

    // make sure everything is okay
    if pNtkAig && !Abc_NtkCheck( pNtkAig ) {
        Abc_Print( 1, "Abc_NtkDar: The network check has failed.\n" );
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
    return pNtkAig;
}


/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarFraig(pNtk *Abc_Ntk_t, nConfLimit int32, fDoSparse int32, fProve int32, fTransfer int32, fSpeculate int32, fChoicing int32, fVerbose int32) *Abc_Ntk_t {
    var Pars Fra_Par_t;
    var pPars *Fra_Par_t = &Pars; 
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    Fra_ParamsDefault( pPars );
    pPars.nBTLimitNode = nConfLimit;
    pPars.fChoicing    = fChoicing;
    pPars.fDoSparse    = fDoSparse;
    pPars.fSpeculate   = fSpeculate;
    pPars.fProve       = fProve;
    pPars.fVerbose     = fVerbose;
    pMan = Fra_FraigPerform( pTemp = pMan, pPars );
    if fChoicing != 0 {
        pNtkAig = Abc_NtkFromDarChoices( pNtk, pMan );
    } else {
        pNtkAig = Abc_NtkFromDar( pNtk, pMan ); }
    Aig_ManStop( pTemp );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarFraigPart(pNtk *Abc_Ntk_t, nPartSize int32, nConfLimit int32, nLevelMax int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    pMan = Aig_ManFraigPartitioned( pTemp = pMan, nPartSize, nConfLimit, nLevelMax, fVerbose );
    Aig_ManStop( pTemp );
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCSweep(pNtk *Abc_Ntk_t, nCutsMax int32, nLeafMax int32, fVerbose int32) *Abc_Ntk_t {
//    extern Aig_Man_t * Csw_Sweep( Aig_Man_t * pAig, int nCutsMax, int nLeafMax, int fVerbose );
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    pMan = Csw_Sweep( pTemp = pMan, nCutsMax, nLeafMax, fVerbose );
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pTemp );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDRewrite(pNtk *Abc_Ntk_t, pPars *Dar_RwrPar_t) *Abc_Ntk_t {
    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t;
    var clk abctime;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
//    Aig_ManPrintStats( pMan );
/*
//    Aig_ManSupports( pMan );
    {
        Vec_Vec_t * vParts;
        vParts = Aig_ManPartitionSmart( pMan, 50, 1, NULL );
        Vec_VecFree( vParts );
    }
*/
    Dar_ManRewrite( pMan, pPars );
//    pMan = Dar_ManBalance( pTemp = pMan, pPars->fUpdateLevel );
//    Aig_ManStop( pTemp );

clk = Abc_Clock();
    pMan = Aig_ManDupDfs( pTemp = pMan ); 
    Aig_ManStop( pTemp );
//ABC_PRT( "time", Abc_Clock() - clk );

//    Aig_ManPrintStats( pMan );
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDRefactor(pNtk *Abc_Ntk_t, pPars *Dar_RefPar_t) *Abc_Ntk_t {
    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t;
    var clk abctime;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
//    Aig_ManPrintStats( pMan );

    Dar_ManRefactor( pMan, pPars );
//    pMan = Dar_ManBalance( pTemp = pMan, pPars->fUpdateLevel );
//    Aig_ManStop( pTemp );

clk = Abc_Clock();
    pMan = Aig_ManDupDfs( pTemp = pMan ); 
    Aig_ManStop( pTemp );
//ABC_PRT( "time", Abc_Clock() - clk );

//    Aig_ManPrintStats( pMan );
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDC2(pNtk *Abc_Ntk_t, fBalance int32, fUpdateLevel int32, fFanout int32, fPower int32, fVerbose int32) *Abc_Ntk_t {
    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t;
    var clk abctime;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
//    Aig_ManPrintStats( pMan );

clk = Abc_Clock();
    pMan = Dar_ManCompress2( pTemp = pMan, fBalance, fUpdateLevel, fFanout, fPower, fVerbose ); 
    Aig_ManStop( pTemp );
//ABC_PRT( "time", Abc_Clock() - clk );

//    Aig_ManPrintStats( pMan );
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDChoice(pNtk *Abc_Ntk_t, fBalance int32, fUpdateLevel int32, fConstruct int32, nConfMax int32, nLevelMax int32, fVerbose int32) *Abc_Ntk_t {
    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    pMan = Dar_ManChoice( pTemp = pMan, fBalance, fUpdateLevel, fConstruct, nConfMax, nLevelMax, fVerbose );
    Aig_ManStop( pTemp );
    pNtkAig = Abc_NtkFromDarChoices( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDch(pNtk *Abc_Ntk_t, pPars *Dch_Pars_t) *Abc_Ntk_t {
    extern Gia_Man_t * Dar_NewChoiceSynthesis( Aig_Man_t * pAig, int fBalance, int fUpdateLevel, int fPower, int fLightSynth, int fVerbose );
    extern Aig_Man_t * Cec_ComputeChoices( Gia_Man_t * pGia, Dch_Pars_t * pPars );

    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t;
    var pGia *Gia_Man_t;
    var clk abctime;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
clk = Abc_Clock();
    if pPars.fSynthesis {
        pGia = Dar_NewChoiceSynthesis( pMan, 1, 1, pPars.fPower, pPars.fLightSynth, pPars.fVerbose );
    } else {
        pGia = Gia_ManFromAig( pMan );
        Aig_ManStop( pMan );
    }
pPars.timeSynth = Abc_Clock() - clk;
    if pPars.fUseGia {
        pMan = Cec_ComputeChoices( pGia, pPars );
    } else {
        pMan = Gia_ManToAigSkip( pGia, 3 );
        Gia_ManStop( pGia );
        pMan = Dch_ComputeChoices( pTemp = pMan, pPars );
        Aig_ManStop( pTemp );
    }
    pNtkAig = Abc_NtkFromDarChoices( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDrwsat(pNtk *Abc_Ntk_t, fBalance int32, fVerbose int32) *Abc_Ntk_t {
    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t;
    var clk abctime;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
//    Aig_ManPrintStats( pMan );

clk = Abc_Clock();
    pMan = Dar_ManRwsat( pTemp = pMan, fBalance, fVerbose ); 
    Aig_ManStop( pTemp );
//ABC_PRT( "time", Abc_Clock() - clk );

//    Aig_ManPrintStats( pMan );
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConstructFromCnf(pNtk *Abc_Ntk_t, p *Cnf_Man_t, vMapped *Vec_Ptr_t) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pNode, pNodeNew *Abc_Obj_t;
    var pObj, pLeaf *Aig_Obj_t;
    var pCut *Cnf_Cut_t;
    var vCover *Vec_Int_t;
    var uTruth uint32;
    var i, k, nDupGates int32;
    // create the new network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_SOP );
    // make the mapper point to the new network
    Aig_ManConst1(p.pManAig).pData = Abc_NtkCreateNodeConst1(pNtkNew);
    Abc_NtkForEachCi( pNtk, pNode, i )
        Aig_ManCi(p.pManAig, i).pData = pNode.pCopy;
    // process the nodes in topological order
    vCover = Vec_IntAlloc( 1 << 16 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vMapped, pObj, i )
    {
        // create new node
        pNodeNew = Abc_NtkCreateNode( pNtkNew );
        // add fanins according to the cut
        pCut = (Cnf_Cut_t *)pObj.pData;
        Cnf_CutForEachLeaf( p.pManAig, pCut, pLeaf, k )
            Abc_ObjAddFanin( pNodeNew, (Abc_Obj_t *)pLeaf.pData );
        // add logic function
        if pCut.nFanins < 5 {
            uTruth = 0xFFFF & *Cnf_CutTruth(pCut);
            Cnf_SopConvertToVector( p.pSops[uTruth], p.pSopSizes[uTruth], vCover );
            pNodeNew.pData = Abc_SopCreateFromIsop( (Mem_Flex_t *)pNtkNew.pManFunc, pCut.nFanins, vCover );
        
        } else {
            pNodeNew.pData = Abc_SopCreateFromIsop( (Mem_Flex_t *)pNtkNew.pManFunc, pCut.nFanins, pCut.vIsop[1] ); }
        // save the node
        pObj.pData = pNodeNew;
    }
    Vec_IntFree( vCover );
    // add the CO drivers
    Abc_NtkForEachCo( pNtk, pNode, i )
    {
        pObj = Aig_ManCo(p.pManAig, i);
        pNodeNew = Abc_ObjNotCond( (Abc_Obj_t *)Aig_ObjFanin0(pObj).pData, Aig_ObjFaninC0(pObj) );
        Abc_ObjAddFanin( pNode.pCopy, pNodeNew );
    }

    // remove the constant node if not used
    pNodeNew = (Abc_Obj_t *)Aig_ManConst1(p.pManAig).pData;
    if Abc_ObjFanoutNum(pNodeNew) == 0 {
        Abc_NtkDeleteObj( pNodeNew ); }
    // minimize the node
//    Abc_NtkSweep( pNtkNew, 0 );
    // decouple the PO driver nodes to reduce the number of levels
    nDupGates = Abc_NtkLogicMakeSimpleCos( pNtkNew, 1 );
//    if ( nDupGates && If_ManReadVerbose(pIfMan) )
//        Abc_Print( 1, "Duplicated %d gates to decouple the CO drivers.\n", nDupGates );
    if !Abc_NtkCheck( pNtkNew ) {
        Abc_Print( 1, "Abc_NtkConstructFromCnf(): Network check has failed.\n" ); }
    return pNtkNew;
}
 
/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarToCnf(pNtk *Abc_Ntk_t, pFileName string, fFastAlgo int32, fChangePol int32, fVerbose int32) *Abc_Ntk_t {
//    Vec_Ptr_t * vMapped = NULL;
    var pMan *Aig_Man_t;
//    Cnf_Man_t * pManCnf = NULL;
    var pCnf *Cnf_Dat_t;
    var pNtkNew *Abc_Ntk_t  = nil;
    var clk abctime  = Abc_Clock();
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    // convert to the AIG manager
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    if pMan == nil {
        return nil; }
    if !Aig_ManCheck( pMan ) {
        Abc_Print( 1, "Abc_NtkDarToCnf: AIG check has failed.\n" );
        Aig_ManStop( pMan );
        return nil;
    }
    // perform balance
    if fVerbose != 0 {
    Aig_ManPrintStats( pMan ); }

    // derive CNF
    if fFastAlgo != 0 {
        pCnf = Cnf_DeriveFast( pMan, 0 );
    } else {
        pCnf = Cnf_Derive( pMan, 0 ); }

    // adjust polarity
    if fChangePol != 0 {
        Cnf_DataTranformPolarity( pCnf, 0 ); }

    // print stats
//    if ( fVerbose )
    {
        Abc_Print( 1, "CNF stats: Vars = %6d. Clauses = %7d. Literals = %8d.   ", pCnf.nVars, pCnf.nClauses, pCnf.nLiterals );
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }

/*
    // write the network for verification
    pManCnf = Cnf_ManRead();
    vMapped = Cnf_ManScanMapping( pManCnf, 1, 0 );
    pNtkNew = Abc_NtkConstructFromCnf( pNtk, pManCnf, vMapped );
    Vec_PtrFree( vMapped );
*/
    // write CNF into a file
    Cnf_DataWriteIntoFile( pCnf, pFileName, 0, nil, nil );
    Cnf_DataFree( pCnf );
    Cnf_ManFree();
    Aig_ManStop( pMan );
    return pNtkNew;
}


/**Function*************************************************************

  Synopsis    [Solves combinational miter using a SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDSat(pNtk *Abc_Ntk_t, nConfLimit ABC_INT64_T, nInsLimit ABC_INT64_T, nLearnedStart int32, nLearnedDelta int32, nLearnedPerce int32, fAlignPol int32, fAndOuts int32, fNewSolver int32, fVerbose int32) int32 {
    var pMan *Aig_Man_t;
    var RetValue int32;//, clk = Abc_Clock();
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0");
//    assert( Abc_NtkPoNum(pNtk) == 1 );
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    RetValue = Fra_FraigSat( pMan, nConfLimit, nInsLimit, nLearnedStart, nLearnedDelta, nLearnedPerce, fAlignPol, fAndOuts, fNewSolver, fVerbose ); 
    pNtk.pModel = (int *)pMan.pData, pMan.pData = nil;
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Solves combinational miter using a SAT solver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPartitionedSat(pNtk *Abc_Ntk_t, nAlgo int32, nPartSize int32, nConfPart int32, nConfTotal int32, fAlignPol int32, fSynthesize int32, fVerbose int32) int32 {
    extern int Aig_ManPartitionedSat( Aig_Man_t * pNtk, int nAlgo, int nPartSize, int nConfPart, int nConfTotal, int fAlignPol, int fSynthesize, int fVerbose );
    var pMan *Aig_Man_t;
    var RetValue int32;//, clk = Abc_Clock();
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0");
    pMan = Abc_NtkToDar( pNtk, 0, 0 );
    RetValue = Aig_ManPartitionedSat( pMan, nAlgo, nPartSize, nConfPart, nConfTotal, fAlignPol, fSynthesize, fVerbose ); 
    pNtk.pModel = (int *)pMan.pData, pMan.pData = nil;
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarCec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nConfLimit int32, fPartition int32, fVerbose int32) int32 {
    var pMan, pMan1, pMan2 *Aig_Man_t;
    var pMiter *Abc_Ntk_t;
    var RetValue int32;
    var clkTotal abctime  = Abc_Clock();
/*
    {
    extern void Cec_ManVerifyTwoAigs( Aig_Man_t * pAig0, Aig_Man_t * pAig1, int fVerbose );
    Aig_Man_t * pAig0 = Abc_NtkToDar( pNtk1, 0, 0 );
    Aig_Man_t * pAig1 = Abc_NtkToDar( pNtk2, 0, 0 );
    Cec_ManVerifyTwoAigs( pAig0, pAig1, 1 );
    Aig_ManStop( pAig0 );
    Aig_ManStop( pAig1 );
    return 1;
    }
*/
    // cannot partition if it is already a miter
    if pNtk2 == nil && fPartition == 1 {
        Abc_Print( 1, "Abc_NtkDarCec(): Switching to non-partitioned CEC for the miter.\n" );
        fPartition = 0;
    }

    // if partitioning is selected, call partitioned CEC
    if fPartition != 0 {
        pMan1 = Abc_NtkToDar( pNtk1, 0, 0 );
        pMan2 = Abc_NtkToDar( pNtk2, 0, 0 );
        RetValue = Fra_FraigCecPartitioned( pMan1, pMan2, nConfLimit, 100, 1, fVerbose );
        Aig_ManStop( pMan1 );
        Aig_ManStop( pMan2 );
        goto finish;
    }

    if pNtk2 != nil {
        // get the miter of the two networks
        pMiter = Abc_NtkMiter( pNtk1, pNtk2, 0, 0, 0, 0 );
        if pMiter == nil {
            Abc_Print( 1, "Miter computation has failed.\n" );
            return 0;
        }
    
    } else {
        pMiter = Abc_NtkDup( pNtk1 );
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
//        extern void Abc_NtkVerifyReportErrorSeq( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, int nFrames );
        Abc_Print( 1, "Networks are NOT EQUIVALENT after structural hashing.\n" );
        // report the error
        if pNtk2 == nil {
            pNtk1.pModel = Abc_NtkVerifyGetCleanModel( pNtk1, 1 ); }
//        pMiter->pModel = Abc_NtkVerifyGetCleanModel( pMiter, nFrames );
//        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pMiter->pModel, nFrames );
//        ABC_FREE( pMiter->pModel );
        Abc_NtkDelete( pMiter );
        return 0;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pMiter );
        Abc_Print( 1, "Networks are equivalent after structural hashing.\n" );
        return 1;
    }

    // derive the AIG manager
    pMan = Abc_NtkToDar( pMiter, 0, 0 );
    Abc_NtkDelete( pMiter );
    if pMan == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return -1;
    }
    // perform verification
    RetValue = Fra_FraigCec( &pMan, 100000, fVerbose );
    // transfer model if given
    if pNtk2 == nil {
        pNtk1.pModel = (int *)pMan.pData, pMan.pData = nil; }
    Aig_ManStop( pMan );

finish:
    // report the miter
    if RetValue == 1 {
        Abc_Print( 1, "Networks are equivalent.  " );
ABC_PRT( "Time", Abc_Clock() - clkTotal );
    
    } else if RetValue == 0 {
        Abc_Print( 1, "Networks are NOT EQUIVALENT.  " );
ABC_PRT( "Time", Abc_Clock() - clkTotal );
    
    } else {
        Abc_Print( 1, "Networks are UNDECIDED.  " );
ABC_PRT( "Time", Abc_Clock() - clkTotal );
    }
    fflush( os.Stdout );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSeqSweep(pNtk *Abc_Ntk_t, pPars *Fra_Ssw_t) *Abc_Ntk_t {
    var Params Fraig_Params_t;
    var pNtkAig *Abc_Ntk_t = nil;
    var pNtkFraig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    var clk abctime  = Abc_Clock();

    // preprocess the miter by fraiging it
    // (note that for each functional class, fraiging leaves one representative;
    // so fraiging does not reduce the number of functions represented by nodes
    Fraig_ParamsSetDefault( &Params );
    Params.nBTLimit = 100000;
    if pPars.fFraiging && pPars.nPartSize == 0 {
        pNtkFraig = Abc_NtkFraig( pNtk, &Params, 0, 0 );
if pPars.fVerbose {
ABC_PRT( "Initial fraiging time", Abc_Clock() - clk );
}
    
    } else {
        pNtkFraig = Abc_NtkDup( pNtk ); }

    pMan = Abc_NtkToDar( pNtkFraig, 0, 1 );
    Abc_NtkDelete( pNtkFraig );
    if pMan == nil {
        return nil; }

//    pPars->TimeLimit = 5.0;
    pMan = Fra_FraigInduction( pTemp = pMan, pPars );
    Aig_ManStop( pTemp );
    if pMan != nil {
        if Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
            pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
        } else {
            var pObj *Abc_Obj_t;
            var i int32;
            pNtkAig = Abc_NtkFromDar( pNtk, pMan );
            Abc_NtkForEachLatch( pNtkAig, pObj, i )
                Abc_LatchSetInit0( pObj );
        }
        Aig_ManStop( pMan );
    }
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Print Latch Equivalence Classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintLatchEquivClasses(pNtk *Abc_Ntk_t, pAig *Aig_Man_t) {
    var header_dumped int32;
    var num_orig_latches int32  = Abc_NtkLatchNum(pNtk);
    var pNames *string  = ABC_ALLOC( char *, num_orig_latches );
    var p_irrelevant *int32  = ABC_ALLOC( int, num_orig_latches );
    var pFlopName, pReprName string;
    var pFlop, pRepr *Aig_Obj_t;
    var pNtkFlop *Abc_Obj_t; 
    var repr_idx int32;
    var i int32;

    Abc_NtkForEachLatch( pNtk, pNtkFlop, i )
    {
        var temp_name string  = Abc_ObjName( Abc_ObjFanout0(pNtkFlop) );
        pNames[i] = ABC_ALLOC( char , len(temp_name)+1);
        strcpy(pNames[i], temp_name);
    }
    i = 0;
    
    Aig_ManSetCioIds( pAig );
    Saig_ManForEachLo( pAig, pFlop, i )
    {
        p_irrelevant[i] = false;
        
        pFlopName = pNames[i];

        pRepr = Aig_ObjRepr(pAig, pFlop);

        if pRepr == nil {
            // Abc_Print( 1, "Nothing equivalent to flop %s\n", pFlopName);
//            p_irrelevant[i] = true;
            continue;
        }

        if !(header_dumped != 0) {
            Abc_Print( 1, "Here are the flop equivalences:\n");
            header_dumped = true;
        }

        // pRepr is representative of the equivalence class, to which pFlop belongs
        if Aig_ObjIsConst1(pRepr) {
            Abc_Print( 1, "Original flop %s is proved equivalent to constant.\n", pFlopName );
            // Abc_Print( 1, "Original flop # %d is proved equivalent to constant.\n", i );
            continue;
        }

        assert(Saig_ObjIsLo( pAig, pRepr ), "Saig_ObjIsLo( pAig, pRepr )");
        repr_idx = Aig_ObjCioId(pRepr) - Saig_ManPiNum(pAig);
        pReprName = pNames[repr_idx];
        Abc_Print( 1, "Original flop %s is proved equivalent to flop %s.\n",  pFlopName, pReprName );
        // Abc_Print( 1, "Original flop # %d is proved equivalent to flop # %d.\n",  i, repr_idx );
    }

    header_dumped = false;
    for i=0; i<num_orig_latches; ++i {
        if p_irrelevant[i] {
            if !(header_dumped != 0) {
                Abc_Print( 1, "The following flops have been deemed irrelevant:\n");
                header_dumped = true;
            }
            Abc_Print( 1, "%s ", pNames[i]);
        }
        
        ABC_FREE(pNames[i]);
    }
    if header_dumped != 0 {
        Abc_Print( 1, "\n"); }
    
    ABC_FREE(pNames);
    ABC_FREE(p_irrelevant);
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSeqSweep2(pNtk *Abc_Ntk_t, pPars *Ssw_Pars_t) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;

    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }

    pMan = Ssw_SignalCorrespondence( pTemp = pMan, pPars );

    if pPars.fFlopVerbose {
        Abc_NtkPrintLatchEquivClasses(pNtk, pTemp); }

    Aig_ManStop( pTemp );
    if pMan == nil {
        return nil; }

    if Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
        pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    } else {
        var pObj *Abc_Obj_t;
        var i int32;
        pNtkAig = Abc_NtkFromDar( pNtk, pMan );
        Abc_NtkForEachLatch( pNtkAig, pObj, i )
            Abc_LatchSetInit0( pObj );
    }
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Computes latch correspondence.]

  Description [] 
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Abc_NtkDarLcorr(pNtk *Abc_Ntk_t, nFramesP int32, nConfMax int32, fVerbose int32) *Abc_Ntk_t {
    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t  = nil;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pMan = Fra_FraigLatchCorrespondence( pTemp = pMan, nFramesP, nConfMax, 0, fVerbose, nil, 0.0 );
    Aig_ManStop( pTemp );
    if pMan != nil {
        if Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
            pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
        } else {
            var pObj *Abc_Obj_t;
            var i int32;
            pNtkAig = Abc_NtkFromDar( pNtk, pMan );
            Abc_NtkForEachLatch( pNtkAig, pObj, i )
                Abc_LatchSetInit0( pObj );
        }
        Aig_ManStop( pMan );
    }
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Computes latch correspondence.]

  Description [] 
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Abc_NtkDarLcorrNew(pNtk *Abc_Ntk_t, nVarsMax int32, nConfMax int32, fVerbose int32) *Abc_Ntk_t {
    var Pars Ssw_Pars_t;
    var pPars *Ssw_Pars_t = &Pars;
    var pMan, pTemp *Aig_Man_t;
    var pNtkAig *Abc_Ntk_t  = nil;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    Ssw_ManSetDefaultParams( pPars );
    pPars.fLatchCorrOpt = 1;
    pPars.nBTLimit      = nConfMax;
    pPars.nSatVarMax    = nVarsMax;
    pPars.fVerbose      = fVerbose;
    pMan = Ssw_SignalCorrespondence( pTemp = pMan, pPars );
    Aig_ManStop( pTemp );
    if pMan != nil {
        if Aig_ManRegNum(pMan) < Abc_NtkLatchNum(pNtk) {
            pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
        } else {
            var pObj *Abc_Obj_t;
            var i int32;
            pNtkAig = Abc_NtkFromDar( pNtk, pMan );
            Abc_NtkForEachLatch( pNtkAig, pObj, i )
                Abc_LatchSetInit0( pObj );
        }
        Aig_ManStop( pMan );
    }
    return pNtkAig;
}

/*
#include <signal.h>
#include "misc/util/utilMem.h"
static void sigfunc( int signo ) 
{
    if (signo == SIGINT) {
        Abc_Print( 1, "SIGINT received!\n");
        s_fInterrupt = 1;
    }
}
*/

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarBmc(pNtk *Abc_Ntk_t, nStart int32, nFrames int32, nSizeMax int32, nNodeDelta int32, nTimeOut int32, nBTLimit int32, nBTLimitAll int32, fRewrite int32, fNewAlgo int32, fOrDecomp int32, nCofFanLit int32, fVerbose int32, piFrames *int32, fUseSatoko int32) int32 {
    var pMan *Aig_Man_t;
    var vMap *Vec_Int_t  = nil;
    var status int32;
    var RetValue int32 = -1;
    var clk abctime  = Abc_Clock();
    var nTimeLimit abctime;
    if nTimeOut {
        nTimeLimit = nTimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeLimit = 0;
    }
    // derive the AIG manager
    if fOrDecomp != 0 {
        pMan = Abc_NtkToDarBmc( pNtk, &vMap );
    } else {
        pMan = Abc_NtkToDar( pNtk, 0, 1 ); }
    if pMan == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return RetValue;
    }
    assert(pMan.nRegs > 0, "pMan.nRegs > 0");
    assert(vMap == nil || Vec_IntSize(vMap) == Saig_ManPoNum(pMan), "vMap == nil || Vec_IntSize(vMap) == Saig_ManPoNum(pMan)");
    if (fVerbose != 0) && (vMap != nil) && Abc_NtkPoNum(pNtk) != Saig_ManPoNum(pMan) { 
        Abc_Print( 1, "Expanded %d outputs into %d outputs using OR decomposition.\n", Abc_NtkPoNum(pNtk), Saig_ManPoNum(pMan) ); }

    // perform verification
    if fNewAlgo != 0 { // command 'bmc'
    
        var iFrame int32;
        RetValue = Saig_ManBmcSimple( pMan, nFrames, nSizeMax, nBTLimit, fRewrite, fVerbose, &iFrame, nCofFanLit, fUseSatoko );
        if piFrames != nil {
            *piFrames = iFrame; }
        ABC_FREE( pNtk.pModel );
        ABC_FREE( pNtk.pSeqModel );
        pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
        if RetValue == 1 {
            Abc_Print( 1, "Incorrect return value.  " );
        } else if RetValue == -1 {
            Abc_Print( 1, "No output asserted in %d frames. Resource limit reached ", Abc_MaxInt(iFrame+1,0) );
            if nTimeLimit && Abc_Clock() > nTimeLimit {
                Abc_Print( 1, "(timeout %d sec). ", nTimeLimit );
            } else {
                Abc_Print( 1, "(conf limit %d). ", nBTLimit ); }
        
        } else { // if ( RetValue == 0 )
        
            var pCex *Abc_Cex_t  = pNtk.pSeqModel;
            Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d. ", pCex.iPo, pNtk.pName, pCex.iFrame );
        }
ABC_PRT( "Time", Abc_Clock() - clk );
    
    } else { 
        RetValue = Saig_BmcPerform( pMan, nStart, nFrames, nNodeDelta, nTimeOut, nBTLimit, nBTLimitAll, fVerbose, 0, piFrames, 0, fUseSatoko );
        ABC_FREE( pNtk.pModel );
        ABC_FREE( pNtk.pSeqModel );
        pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
    }
    // verify counter-example
    if pNtk.pSeqModel {
        status = Saig_ManVerifyCex( pMan, pNtk.pSeqModel );
        if status == 0 {
            Abc_Print( 1, "Abc_NtkDarBmc(): Counter-example verification has FAILED.\n" ); }
    }
    Aig_ManStop( pMan );
    // update the counter-example
    if pNtk.pSeqModel && (vMap != nil) {
        pNtk.pSeqModel.iPo = Vec_IntEntry( vMap, pNtk.pSeqModel.iPo ); }
    Vec_IntFreeP( &vMap );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarBmc3(pNtk *Abc_Ntk_t, pPars *Saig_ParBmc_t, fOrDecomp int32) int32 {
    var pMan *Aig_Man_t;
    var vMap *Vec_Int_t  = nil;
    var status int32;
    var RetValue int32 = -1;
    var clk abctime  = Abc_Clock();
    var nTimeOut abctime;
    if pPars.nTimeOut {
        nTimeOut = pPars.nTimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeOut = 0;
    }
    if (fOrDecomp != 0) && !pPars.fSolveAll {
        pMan = Abc_NtkToDarBmc( pNtk, &vMap );
    } else {
        pMan = Abc_NtkToDar( pNtk, 0, 1 ); }
    if pMan == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return RetValue;
    }
    assert(pMan.nRegs > 0, "pMan.nRegs > 0");
    if pPars.fVerbose && (vMap != nil) && Abc_NtkPoNum(pNtk) != Saig_ManPoNum(pMan) { 
        Abc_Print( 1, "Expanded %d outputs into %d outputs using OR decomposition.\n", Abc_NtkPoNum(pNtk), Saig_ManPoNum(pMan) ); }

    RetValue = Saig_ManBmcScalable( pMan, pPars );
    ABC_FREE( pNtk.pModel );
    ABC_FREE( pNtk.pSeqModel );
    pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
    if !pPars.fSilent {
        if RetValue == 1 {
            Abc_Print( 1, "Explored all reachable states after completing %d frames.  ", 1<<Aig_ManRegNum(pMan) );
        
        } else if RetValue == -1 {
            if pPars.nFailOuts == 0 {
                Abc_Print( 1, "No output asserted in %d frames. Resource limit reached ", Abc_MaxInt(pPars.iFrame+1,0) );
                if nTimeOut && Abc_Clock() > nTimeOut {
                    Abc_Print( 1, "(timeout %d sec). ", pPars.nTimeOut );
                } else {
                    Abc_Print( 1, "(conf limit %d). ", pPars.nConfLimit ); }
            
            } else {
                Abc_Print( 1, "The total of %d outputs asserted in %d frames. Resource limit reached ", pPars.nFailOuts, pPars.iFrame );
                if Abc_Clock() > nTimeOut {
                    Abc_Print( 1, "(timeout %d sec). ", pPars.nTimeOut );
                } else {
                    Abc_Print( 1, "(conf limit %d). ", pPars.nConfLimit ); }
            }
        
        } else { // if ( RetValue == 0 )
        
            if !pPars.fSolveAll {
                var pCex *Abc_Cex_t  = pNtk.pSeqModel;
                Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d. ", pCex.iPo, pNtk.pName, pCex.iFrame );
            
            } else {
                var nOutputs int32  = Saig_ManPoNum(pMan) - Saig_ManConstrNum(pMan);
                if pMan.vSeqModelVec == nil || Vec_PtrCountZero(pMan.vSeqModelVec) == nOutputs {
                    Abc_Print( 1, "None of the %d outputs is found to be SAT", nOutputs );
                } else if Vec_PtrCountZero(pMan.vSeqModelVec) == 0 {
                    Abc_Print( 1, "All %d outputs are found to be SAT", nOutputs );
                } else {
                    Abc_Print( 1, "Some outputs are SAT (%d out of %d)", nOutputs - Vec_PtrCountZero(pMan.vSeqModelVec), nOutputs );
                    if pPars.nDropOuts {
                        Abc_Print( 1, " while others timed out (%d out of %d)", pPars.nDropOuts, nOutputs ); }
                }
                Abc_Print( 1, " after %d frames", pPars.iFrame+2 );
                Abc_Print( 1, ".   " );
            }
        }
        ABC_PRT( "Time", Abc_Clock() - clk );
    }
    if RetValue == 0 && pPars.fSolveAll {
        if pNtk.vSeqModelVec {
            Vec_PtrFreeFree( pNtk.vSeqModelVec ); }
        pNtk.vSeqModelVec = pMan.vSeqModelVec;  pMan.vSeqModelVec = nil;
    }
    if pNtk.pSeqModel {
        status = Saig_ManVerifyCex( pMan, pNtk.pSeqModel );
        if status == 0 {
            Abc_Print( 1, "Abc_NtkDarBmc3(): Counter-example verification has FAILED.\n" ); }
    }
    Aig_ManStop( pMan );
    // update the counter-example
    if pNtk.pSeqModel && (vMap != nil) {
        pNtk.pSeqModel.iPo = Vec_IntEntry( vMap, pNtk.pSeqModel.iPo ); }
    Vec_IntFreeP( &vMap );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarBmcInter_int(pMan *Aig_Man_t, pPars *Inter_ManParams_t, ppNtkRes **Aig_Man_t) int32 {
    var RetValue int32 = -1;
    var iFrame int32;
    var clk abctime  = Abc_Clock();
    var nTotalProvedSat int32;
    assert(pMan.nRegs > 0, "pMan.nRegs > 0");
    if ppNtkRes {
        *ppNtkRes = nil; }
    if pPars.fUseSeparate {
        var pTemp, pAux *Aig_Man_t;
        var pObjPo *Aig_Obj_t;
        var i, Counter int32;
        Saig_ManForEachPo( pMan, pObjPo, i )
        {
            if Aig_ObjFanin0(pObjPo) == Aig_ManConst1(pMan) {
                continue; }
            if pPars.fVerbose {
                Abc_Print( 1, "Solving output %2d (out of %2d):\n", i, Saig_ManPoNum(pMan) ); }
            pTemp = Aig_ManDupOneOutput( pMan, i, 1 );
            pTemp = Aig_ManScl( pAux = pTemp, 1, 1, 0, -1, -1, 0, 0 );
            Aig_ManStop( pAux );
            if Aig_ManRegNum(pTemp) == 0 {
                pTemp.pSeqModel = nil;
                RetValue = Fra_FraigSat( pTemp, pPars.nBTLimit, 0, 0, 0, 0, 0, 0, 0, 0 ); 
                if pTemp.pData {
                    pTemp.pSeqModel = Abc_CexCreate( Aig_ManRegNum(pMan), Saig_ManPiNum(pMan), (int *)pTemp.pData, 0, i, 1 ); }
//                pNtk->pModel = pTemp->pData, pTemp->pData = NULL;
            
            } else {
                RetValue = Inter_ManPerformInterpolation( pTemp, pPars, &iFrame ); }
            if pTemp.pSeqModel {
                if pPars.fDropSatOuts {
                    Abc_Print( 1, "Output %d proved SAT in frame %d (replacing by const 0 and continuing...)\n", i, pTemp.pSeqModel.iFrame );
                    Aig_ObjPatchFanin0( pMan, pObjPo, Aig_ManConst0(pMan) );
                    Aig_ManStop( pTemp );
                    nTotalProvedSat++;
                    continue;
                
                } else {
                    var pCex *Abc_Cex_t;
                    pCex = pMan.pSeqModel = pTemp.pSeqModel; pTemp.pSeqModel = nil;
                    pCex.iPo = i;
                    Aig_ManStop( pTemp );
                    break;
                }
            }
            // if solved, remove the output
            if RetValue == 1 {
                Aig_ObjPatchFanin0( pMan, pObjPo, Aig_ManConst0(pMan) );
//                    Abc_Print( 1, "Output %3d : Solved ", i );
            
            } else {
                Counter++;
//                    Abc_Print( 1, "Output %3d : Undec  ", i );
            }
//                Aig_ManPrintStats( pTemp );
            Aig_ManStop( pTemp );
            Abc_Print( 1, "Solving output %3d (out of %3d) using interpolation.\r", i, Saig_ManPoNum(pMan) );
        }
        Aig_ManCleanup( pMan );
        if pMan.pSeqModel == nil {
            Abc_Print( 1, "Interpolation left %d (out of %d) outputs unsolved              \n", Counter, Saig_ManPoNum(pMan) );
            if Counter != 0 {
                RetValue = -1; }
        }
        if ppNtkRes {
            pTemp = Aig_ManDupUnsolvedOutputs( pMan, 1 );
            *ppNtkRes = Aig_ManScl( pTemp, 1, 1, 0, -1, -1, 0, 0 );
            Aig_ManStop( pTemp );
        }
    
    } else {    
        RetValue = Inter_ManPerformInterpolation( pMan, pPars, &iFrame );
    }
    if nTotalProvedSat != 0 {
        Abc_Print( 1, "The total of %d outputs proved SAT and replaced by const 0 in this run.\n", nTotalProvedSat ); }
    if RetValue == 1 {
        Abc_Print( 1, "Property proved.  " );
    } else if RetValue == 0 {
        Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", pMan.pSeqModel ? pMan.pSeqModel.iPo : -1, pMan.pName, iFrame );
    } else if RetValue == -1 {
        Abc_Print( 1, "Property UNDECIDED.  " );
    } else {
        assert(false, "0"); }
ABC_PRT( "Time", Abc_Clock() - clk );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarBmcInter(pNtk *Abc_Ntk_t, pPars *Inter_ManParams_t, ppNtkRes **Abc_Ntk_t) int32 {
    var pMan *Aig_Man_t;
    var RetValue int32;
    if ppNtkRes {
        *ppNtkRes = nil; }
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return -1;
    }
    if pPars.fUseSeparate && ppNtkRes {
        var pManNew *Aig_Man_t;
        RetValue = Abc_NtkDarBmcInter_int( pMan, pPars, &pManNew );
        *ppNtkRes = Abc_NtkFromAigPhase( pManNew );
        Aig_ManStop( pManNew );
    
    } else {
        RetValue = Abc_NtkDarBmcInter_int( pMan, pPars, nil );
    }
    ABC_FREE( pNtk.pModel );
    ABC_FREE( pNtk.pSeqModel );
    pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarDemiter(pNtk *Abc_Ntk_t) int32 { 
    var pFileNameGeneric string;
    var pFileName0 string // was [1000]byte in C

    var pFileName1 string // was [1000]byte in C

    var pMan, pPart0, pPart1 *Aig_Man_t;//, * pMiter;
    // derive the AIG manager
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        Abc_Print( 1, "Converting network into AIG has failed.\n" );
        return 0;
    }
//    if ( !Saig_ManDemiterSimple( pMan, &pPart0, &pPart1 ) )
    if !Saig_ManDemiterSimpleDiff( pMan, &pPart0, &pPart1 ) {
        Aig_ManStop( pMan );
        Abc_Print( 1, "Demitering has failed.\n" );
        return 0;
    }
    // create file names
    pFileNameGeneric = Extra_FileNameGeneric( pNtk.pSpec ? pNtk.pSpec : pNtk.pName );
//    sprintf( pFileName0,  "%s%s",  pFileNameGeneric, "_part0.aig" ); 
//    sprintf( pFileName1,  "%s%s",  pFileNameGeneric, "_part1.aig" ); 
    pFileName0 = fmt.Sprintf("%s", "part0.aig"); 
    pFileName1 = fmt.Sprintf("%s", "part1.aig"); 
    ABC_FREE( pFileNameGeneric );
    // dump files
    Ioa_WriteAiger( pPart0, pFileName0, 0, 0 );
    Ioa_WriteAiger( pPart1, pFileName1, 0, 0 );
    Abc_Print( 1, "Demitering produced two files \"%s\" and \"%s\".\n", pFileName0, pFileName1 );
    // create two-level miter
//    pMiter = Saig_ManCreateMiterTwo( pPart0, pPart1, 2 );
//    Aig_ManDumpBlif( pMiter, "miter01.blif", NULL, NULL );
//    Aig_ManStop( pMiter );
//    Abc_Print( 1, "The new miter is written into file \"%s\".\n", "miter01.blif" );
    Aig_ManStop( pPart0 );
    Aig_ManStop( pPart1 );
    Aig_ManStop( pMan );
    return 1;
} 

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarDemiterNew(pNtk *Abc_Ntk_t) int32 { 
    var pFileNameGeneric string;
    var pFileName0 string // was [1000]byte in C

    var pFileName1 string // was [1000]byte in C

    var pMan, pPart0, pPart1 *Aig_Man_t;//, * pMiter;
    // derive the AIG manager
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        Abc_Print( 1, "Converting network into AIG has failed.\n" );
        return 0;
    }

    Saig_ManDemiterNew( pMan );
    Aig_ManStop( pMan );
    return 1;

//    if ( !Saig_ManDemiterSimple( pMan, &pPart0, &pPart1 ) )
    if !Saig_ManDemiterSimpleDiff( pMan, &pPart0, &pPart1 ) {
        Abc_Print( 1, "Demitering has failed.\n" );
        return 0;
    }
    // create file names
    pFileNameGeneric = Extra_FileNameGeneric( pNtk.pSpec );
    pFileName0 = fmt.Sprintf("%s%s", pFileNameGeneric, "_part0.aig"); 
    pFileName1 = fmt.Sprintf("%s%s", pFileNameGeneric, "_part1.aig"); 
    ABC_FREE( pFileNameGeneric );
    // dump files
    Ioa_WriteAiger( pPart0, pFileName0, 0, 0 );
    Ioa_WriteAiger( pPart1, pFileName1, 0, 0 );
    Abc_Print( 1, "Demitering produced two files \"%s\" and \"%s\".\n", pFileName0, pFileName1 );
    // create two-level miter
//    pMiter = Saig_ManCreateMiterTwo( pPart0, pPart1, 2 );
//    Aig_ManDumpBlif( pMiter, "miter01.blif", NULL, NULL );
//    Aig_ManStop( pMiter );
//    Abc_Print( 1, "The new miter is written into file \"%s\".\n", "miter01.blif" );
    Aig_ManStop( pPart0 );
    Aig_ManStop( pPart1 );
    Aig_ManStop( pMan );
    return 1;
} 

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarDemiterDual(pNtk *Abc_Ntk_t, fVerbose int32) int32 { 
    var pFileNameGeneric string;
    var pFileName0 string // was [1000]byte in C

    var pFileName1 string // was [1000]byte in C

    var pMan, pPart0, pPart1 *Aig_Man_t;//, * pMiter;
    if Abc_NtkPoNum(pNtk) & 1 {
        Abc_Print( 1, "The number of POs should be even.\n" );
        return 0;
    }
    // derive the AIG manager
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        Abc_Print( 1, "Converting network into AIG has failed.\n" );
        return 0;
    }
//    if ( !Saig_ManDemiterSimple( pMan, &pPart0, &pPart1 ) )
    if !Saig_ManDemiterDual( pMan, &pPart0, &pPart1 ) {
        Abc_Print( 1, "Demitering has failed.\n" );
        return 0;
    }
    // create new AIG
    ABC_FREE( pPart0.pName );
    pPart0.pName = Abc_UtilStrsav( "part0" );
    // create new AIGs
    ABC_FREE( pPart1.pName );
    pPart1.pName = Abc_UtilStrsav( "part1" );
    // create file names
    pFileNameGeneric = Extra_FileNameGeneric( pNtk.pSpec );
//    sprintf( pFileName0,  "%s%s",  pFileNameGeneric, "_part0.aig" ); 
//    sprintf( pFileName1,  "%s%s",  pFileNameGeneric, "_part1.aig" ); 
    pFileName0 = fmt.Sprintf("%s", "part0.aig"); 
    pFileName1 = fmt.Sprintf("%s", "part1.aig"); 
    ABC_FREE( pFileNameGeneric );
    Ioa_WriteAiger( pPart0, pFileName0, 0, 0 );
    Ioa_WriteAiger( pPart1, pFileName1, 0, 0 );
    Abc_Print( 1, "Demitering produced two files \"%s\" and \"%s\".\n", pFileName0, pFileName1 );
    // dump files
    if fVerbose != 0 {
//        Abc_Print( 1, "Init:  " );
        Aig_ManPrintStats( pMan );
//        Abc_Print( 1, "Part1: " );
        Aig_ManPrintStats( pPart0 );
//        Abc_Print( 1, "Part2: " );
        Aig_ManPrintStats( pPart1 );
    }
    // create two-level miter
//    pMiter = Saig_ManCreateMiterTwo( pPart0, pPart1, 2 );
//    Aig_ManDumpBlif( pMiter, "miter01.blif", NULL, NULL );
//    Aig_ManStop( pMiter );
//    Abc_Print( 1, "The new miter is written into file \"%s\".\n", "miter01.blif" );
    Aig_ManStop( pPart0 );
    Aig_ManStop( pPart1 );
    Aig_ManStop( pMan );
    return 1;
} 
 
/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarProve(pNtk *Abc_Ntk_t, pSecPar *Fra_Sec_t, nBmcFramesMax int32, nBmcConfMax int32) int32 {
    var pMan *Aig_Man_t;
    var iFrame int32 = -1;
    var RetValue int32 = -1;
    var clkTotal abctime  = Abc_Clock();
    if pSecPar.fTryComb || Abc_NtkLatchNum(pNtk) == 0 {
        var Params Prove_Params_t;
        var pParams *Prove_Params_t = &Params;
        var pNtkComb *Abc_Ntk_t;
        var RetValue int32;
        if Abc_NtkLatchNum(pNtk) == 0 {
            Abc_Print( 1, "The network has no latches. Running CEC.\n" ); }
        // create combinational network
        pNtkComb = Abc_NtkDup( pNtk );
        Abc_NtkMakeComb( pNtkComb, 1 );
        // solve it using combinational equivalence checking
        Prove_ParamsSetDefault( pParams );
        pParams.fVerbose = 1;
        RetValue = Abc_NtkIvyProve( &pNtkComb, pParams );
        // transfer model if given
//        pNtk->pModel = pNtkComb->pModel; pNtkComb->pModel = NULL;
        if RetValue == 0  && (Abc_NtkLatchNum(pNtk) == 0) {
            pNtk.pModel = pNtkComb.pModel; pNtkComb.pModel = nil;
            if pSecPar.fReportSolution {
                Abc_Print( 1, "SOLUTION: FAIL       " );
            } else {
                Abc_Print( 1, "SATISFIABLE    " ); }
            ABC_PRT( "Time", Abc_Clock() - clkTotal );
            return RetValue;
        }
        Abc_NtkDelete( pNtkComb );
        // return the result, if solved
        if RetValue == 1 {
            if pSecPar.fReportSolution {
                Abc_Print( 1, "SOLUTION: PASS       " );
            } else {
                Abc_Print( 1, "UNSATISFIABLE  " ); }
            ABC_PRT( "Time", Abc_Clock() - clkTotal );
            return RetValue;
        }
        // return undecided, if the miter is combinational
        if Abc_NtkLatchNum(pNtk) == 0 {
            Abc_Print( 1, "UNDECIDED      " );
            ABC_PRT( "Time", Abc_Clock() - clkTotal );
            return RetValue;
        }
    }
    // derive the AIG manager
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return -1;
    }
    assert(pMan.nRegs > 0, "pMan.nRegs > 0");

    if pSecPar.fTryBmc {
        RetValue = Saig_BmcPerform( pMan, 0, nBmcFramesMax, 2000, 0, nBmcConfMax, 0, pSecPar.fVerbose, 0, &iFrame, 0, 0 );
        if RetValue == 0 {
            Abc_Print( 1, "Networks are not equivalent.\n" );
            if pSecPar.fReportSolution {
                Abc_Print( 1, "SOLUTION: FAIL       " );
                ABC_PRT( "Time", Abc_Clock() - clkTotal );
            }
            // return the counter-example generated
            ABC_FREE( pNtk.pModel );
            ABC_FREE( pNtk.pSeqModel );
            pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
            Aig_ManStop( pMan );
            return RetValue;
        }
    } 
    // perform verification
    if pSecPar.fUseNewProver {
        RetValue = Ssw_SecGeneralMiter( pMan, nil );
    
    } else {
        RetValue = Fra_FraigSec( pMan, pSecPar, nil );
        ABC_FREE( pNtk.pModel );
        ABC_FREE( pNtk.pSeqModel );
        pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
        if pNtk.pSeqModel {
            var pCex *Abc_Cex_t  = pNtk.pSeqModel;
            Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.\n", pCex.iPo, pNtk.pName, pCex.iFrame );
            if !Saig_ManVerifyCex( pMan, pNtk.pSeqModel ) {
                Abc_Print( 1, "Abc_NtkDarProve(): Counter-example verification has FAILED.\n" ); }
        }
    }
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pSecPar *Fra_Sec_t) int32 {
//    Fraig_Params_t Params;
    var pMan *Aig_Man_t;
    var pMiter *Abc_Ntk_t;//, * pTemp;
    var RetValue int32;
 
    // get the miter of the two networks
    pMiter = Abc_NtkMiter( pNtk1, pNtk2, 0, 0, 0, 0 );
    if pMiter == nil {
        Abc_Print( 1, "Miter computation has failed.\n" );
        return -1;
    }
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if RetValue == 0 {
        extern void Abc_NtkVerifyReportErrorSeq( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, int nFrames );
        Abc_Print( 1, "Networks are NOT EQUIVALENT after structural hashing.\n" );
        // report the error
        pMiter.pModel = Abc_NtkVerifyGetCleanModel( pMiter, pSecPar.nFramesMax );
//        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pMiter->pModel, pSecPar->nFramesMax );
        ABC_FREE( pMiter.pModel );
        Abc_NtkDelete( pMiter );
        return 0;
    }
    if RetValue == 1 {
        Abc_NtkDelete( pMiter );
        Abc_Print( 1, "Networks are equivalent after structural hashing.\n" );
        return 1;
    }

    // commented out because sometimes the problem became non-inductive
/*
    // preprocess the miter by fraiging it
    // (note that for each functional class, fraiging leaves one representative;
    // so fraiging does not reduce the number of functions represented by nodes
    Fraig_ParamsSetDefault( &Params );
    Params.nBTLimit = 100000;
    pMiter = Abc_NtkFraig( pTemp = pMiter, &Params, 0, 0 );
    Abc_NtkDelete( pTemp );
    RetValue = Abc_NtkMiterIsConstant( pMiter );
    if ( RetValue == 0 )
    {
        extern void Abc_NtkVerifyReportErrorSeq( Abc_Ntk_t * pNtk1, Abc_Ntk_t * pNtk2, int * pModel, int nFrames );
        Abc_Print( 1, "Networks are NOT EQUIVALENT after structural hashing.\n" );
        // report the error
        pMiter->pModel = Abc_NtkVerifyGetCleanModel( pMiter, nFrames );
        Abc_NtkVerifyReportErrorSeq( pNtk1, pNtk2, pMiter->pModel, nFrames );
        ABC_FREE( pMiter->pModel );
        Abc_NtkDelete( pMiter );
        return 0;
    }
    if ( RetValue == 1 )
    {
        Abc_NtkDelete( pMiter );
        Abc_Print( 1, "Networks are equivalent after structural hashing.\n" );
        return 1;
    }
*/
    // derive the AIG manager
    pMan = Abc_NtkToDar( pMiter, 0, 1 );
    Abc_NtkDelete( pMiter );
    if pMan == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return -1;
    }
    assert(pMan.nRegs > 0, "pMan.nRegs > 0");

    // perform verification
    RetValue = Fra_FraigSec( pMan, pSecPar, nil );
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarPdr(pNtk *Abc_Ntk_t, pPars *Pdr_Par_t) int32 {
    var RetValue int32  = -1;
    var clk abctime  = Abc_Clock();
    var pMan *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        Abc_Print( 1, "Converting network into AIG has failed.\n" );
        return -1;
    }
    RetValue = Pdr_ManSolve( pMan, pPars );
    pPars.nDropOuts = Saig_ManPoNum(pMan) - pPars.nProveOuts - pPars.nFailOuts;
    if !pPars.fSilent {
        if pPars.fSolveAll {
            Abc_Print( 1, "Properties:  All = %d. Proved = %d. Disproved = %d. Undecided = %d.   ", 
                Saig_ManPoNum(pMan), pPars.nProveOuts, pPars.nFailOuts, pPars.nDropOuts );
        } else if RetValue == 1 {
            Abc_Print( 1, "Property proved.  " );
        } else {
            if RetValue == 0 {
                if pMan.pSeqModel == nil {
                    Abc_Print( 1, "Abc_NtkDarPdr(): Counter-example is not available.\n" );
                } else {
                    Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", pMan.pSeqModel.iPo, pNtk.pName, pMan.pSeqModel.iFrame );
                    if !Saig_ManVerifyCex( pMan, pMan.pSeqModel ) {
                        Abc_Print( 1, "Abc_NtkDarPdr(): Counter-example verification has FAILED.\n" ); }
                }
            
            } else if RetValue == -1 {
                Abc_Print( 1, "Property UNDECIDED.  " );
            } else {
                assert(false, "0"); }
        }
        ABC_PRT( "Time", Abc_Clock() - clk );
/*
        Abc_Print( 1, "Status: " );
        if ( pPars->pOutMap )
        {
            int i;
            for ( i = 0; i < Saig_ManPoNum(pMan); i++ )
                if ( pPars->pOutMap[i] == 1 )
                    Abc_Print( 1, "%d=%s ", i, "unsat" );
                else if ( pPars->pOutMap[i] == 0 )
                    Abc_Print( 1, "%d=%s ", i, "sat" );
                else if ( pPars->pOutMap[i] < 0 )
                    Abc_Print( 1, "%d=%s ", i, "undec" );
                else assert( 0 );
        }
        Abc_Print( 1, "\n" );
*/
    }
    ABC_FREE( pNtk.pSeqModel );
    pNtk.pSeqModel = pMan.pSeqModel;
    pMan.pSeqModel = nil;
    if pNtk.vSeqModelVec {
        Vec_PtrFreeFree( pNtk.vSeqModelVec ); }
    pNtk.vSeqModelVec = pMan.vSeqModelVec;
    pMan.vSeqModelVec = nil;
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarAbSec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nFrames int32, fVerbose int32) int32 {
    var pMan1 *Aig_Man_t;
    var pMan2 *Aig_Man_t = nil;
    var RetValue int32;
    // derive AIG manager
    pMan1 = Abc_NtkToDar( pNtk1, 0, 1 );
    if pMan1 == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return -1;
    }
    assert(Aig_ManRegNum(pMan1) > 0, "Aig_ManRegNum(pMan1) > 0");
    // derive AIG manager
    if pNtk2 != nil {
        pMan2 = Abc_NtkToDar( pNtk2, 0, 1 );
        if pMan2 == nil {
            Aig_ManStop( pMan1 );
            Abc_Print( 1, "Converting miter into AIG has failed.\n" );
            return -1;
        }
        assert(Aig_ManRegNum(pMan2) > 0, "Aig_ManRegNum(pMan2) > 0");
        if Saig_ManPiNum(pMan1) != Saig_ManPiNum(pMan2) {
            Aig_ManStop( pMan1 );
            Aig_ManStop( pMan2 );
            Abc_Print( 1, "The networks have different number of PIs.\n" );
            return -1;
        }
        if Saig_ManPoNum(pMan1) != Saig_ManPoNum(pMan2) {
            Aig_ManStop( pMan1 );
            Aig_ManStop( pMan2 );
            Abc_Print( 1, "The networks have different number of POs.\n" );
            return -1;
        }
        if Aig_ManRegNum(pMan1) != Aig_ManRegNum(pMan2) {
            Aig_ManStop( pMan1 );
            Aig_ManStop( pMan2 );
            Abc_Print( 1, "The networks have different number of flops.\n" );
            return -1;
        }
    }
    // perform verification
    RetValue = Ssw_SecSpecialMiter( pMan1, pMan2, nFrames, fVerbose );
    Aig_ManStop( pMan1 );
    if pMan2 != nil {
        Aig_ManStop( pMan2 ); }
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSimSec(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, pPars *Ssw_Pars_t) int32 {
    var pMan1 *Aig_Man_t;
    var pMan2 *Aig_Man_t = nil;
    var RetValue int32;
    // derive AIG manager
    pMan1 = Abc_NtkToDar( pNtk1, 0, 1 );
    if pMan1 == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return -1;
    }
    assert(Aig_ManRegNum(pMan1) > 0, "Aig_ManRegNum(pMan1) > 0");
    // derive AIG manager
    if pNtk2 != nil {
        pMan2 = Abc_NtkToDar( pNtk2, 0, 1 );
        if pMan2 == nil {
            Abc_Print( 1, "Converting miter into AIG has failed.\n" );
            return -1;
        }
        assert(Aig_ManRegNum(pMan2) > 0, "Aig_ManRegNum(pMan2) > 0");
    }

    // perform verification
    RetValue = Ssw_SecWithSimilarity( pMan1, pMan2, pPars );
    Aig_ManStop( pMan1 );
    if pMan2 != nil {
        Aig_ManStop( pMan2 ); }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarMatch(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nDist int32, fVerbose int32) *Abc_Ntk_t {
    extern Vec_Int_t * Saig_StrSimPerformMatching( Aig_Man_t * p0, Aig_Man_t * p1, int nDist, int fVerbose, Aig_Man_t ** ppMiter );
    var pNtkAig *Abc_Ntk_t;
    var pMan1 *Aig_Man_t;
    var pMan2 *Aig_Man_t = nil;
    var pManRes *Aig_Man_t;
    var vPairs *Vec_Int_t;
    assert(Abc_NtkIsStrash(pNtk1), "Abc_NtkIsStrash(pNtk1)");
    // derive AIG manager
    pMan1 = Abc_NtkToDar( pNtk1, 0, 1 );
    if pMan1 == nil {
        Abc_Print( 1, "Converting miter into AIG has failed.\n" );
        return nil;
    }
    assert(Aig_ManRegNum(pMan1) > 0, "Aig_ManRegNum(pMan1) > 0");
    // derive AIG manager
    if pNtk2 != nil {
        pMan2 = Abc_NtkToDar( pNtk2, 0, 1 );
        if pMan2 == nil {
            Abc_Print( 1, "Converting miter into AIG has failed.\n" );
            return nil;
        }
        assert(Aig_ManRegNum(pMan2) > 0, "Aig_ManRegNum(pMan2) > 0");
    }

    // perform verification
    vPairs = Saig_StrSimPerformMatching( pMan1, pMan2, nDist, 1, &pManRes );
    pNtkAig = Abc_NtkFromAigPhase( pManRes );
    if vPairs != nil {
        Vec_IntFree( vPairs ); }
    if pManRes != nil {
        Aig_ManStop( pManRes ); }
    Aig_ManStop( pMan1 );
    if pMan2 != nil {
        Aig_ManStop( pMan2 ); }
    return pNtkAig;
}


/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarLatchSweep(pNtk *Abc_Ntk_t, fLatchConst int32, fLatchEqual int32, fSaveNames int32, fUseMvSweep int32, nFramesSymb int32, nFramesSatur int32, fVerbose int32, fVeryVerbose int32) *Abc_Ntk_t {
    extern void Aig_ManPrintControlFanouts( Aig_Man_t * p );
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    if fSaveNames != 0 {
        Aig_ManSeqCleanup( pMan );
        if (fLatchConst != 0) && pMan.nRegs {
            pMan = Aig_ManConstReduce( pMan, fUseMvSweep, nFramesSymb, nFramesSatur, fVerbose, fVeryVerbose ); }
        if (fLatchEqual != 0) && pMan.nRegs {
            pMan = Aig_ManReduceLaches( pMan, fVerbose ); }
    
    } else {
        if pMan.vFlopNums {
            Vec_IntFree( pMan.vFlopNums ); }
        pMan.vFlopNums = nil;
        pMan = Aig_ManScl( pTemp = pMan, fLatchConst, fLatchEqual, fUseMvSweep, nFramesSymb, nFramesSatur, fVerbose, fVeryVerbose );
        Aig_ManStop( pTemp );
    }

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
//Aig_ManPrintControlFanouts( pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarRetime(pNtk *Abc_Ntk_t, nStepsMax int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
//    Aig_ManReduceLachesCount( pMan );
    if pMan.vFlopNums {
        Vec_IntFree( pMan.vFlopNums ); } 
    pMan.vFlopNums = nil;

    pMan = Rtm_ManRetime( pTemp = pMan, 1, nStepsMax, fVerbose );
    Aig_ManStop( pTemp );

//    pMan = Aig_ManReduceLaches( pMan, 1 );
//    pMan = Aig_ManConstReduce( pMan, 1, 0 );

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarRetimeF(pNtk *Abc_Ntk_t, nStepsMax int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
//    Aig_ManReduceLachesCount( pMan );
    if pMan.vFlopNums {
        Vec_IntFree( pMan.vFlopNums ); } 
    pMan.vFlopNums = nil;

    pMan = Aig_ManRetimeFrontier( pTemp = pMan, nStepsMax );
    Aig_ManStop( pTemp );

//    pMan = Aig_ManReduceLaches( pMan, 1 );
//    pMan = Aig_ManConstReduce( pMan, 1, 0 );

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarRetimeMostFwd(pNtk *Abc_Ntk_t, nMaxIters int32, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Saig_ManRetimeForward( Aig_Man_t * p, int nIters, int fVerbose );

    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
//    Aig_ManReduceLachesCount( pMan );
    if pMan.vFlopNums {
        Vec_IntFree( pMan.vFlopNums ); } 
    pMan.vFlopNums = nil;

    pMan = Saig_ManRetimeForward( pTemp = pMan, nMaxIters, fVerbose );
    Aig_ManStop( pTemp );

//    pMan = Aig_ManReduceLaches( pMan, 1 );
//    pMan = Aig_ManConstReduce( pMan, 1, 0 );

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarRetimeMinArea(pNtk *Abc_Ntk_t, nMaxIters int32, fForwardOnly int32, fBackwardOnly int32, fInitial int32, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Saig_ManRetimeMinArea( Aig_Man_t * p, int nMaxIters, int fForwardOnly, int fBackwardOnly, int fInitial, int fVerbose );
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    if pMan.vFlopNums {
        Vec_IntFree( pMan.vFlopNums ); } 
    pMan.vFlopNums = nil;

    pMan = Saig_ManRetimeMinArea( pTemp = pMan, nMaxIters, fForwardOnly, fBackwardOnly, fInitial, fVerbose );
    Aig_ManStop( pTemp );

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarRetimeStep(pNtk *Abc_Ntk_t, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan *Aig_Man_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    if pMan.vFlopNums {
        Vec_IntFree( pMan.vFlopNums ); } 
    pMan.vFlopNums = nil;

    Aig_ManPrintStats(pMan);
    Saig_ManRetimeSteps( pMan, 1000, 1, 0 );
    Aig_ManPrintStats(pMan);

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
Abc_Ntk_t * Abc_NtkDarHaigRecord( Abc_Ntk_t * pNtk, int nIters, int nSteps, int fRetimingOnly, int fAddBugs, int fUseCnf, int fVerbose )
{
    Abc_Ntk_t * pNtkAig;
    Aig_Man_t * pMan, * pTemp;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if ( pMan == NULL )
        return NULL;
    if ( pMan->vFlopNums )
        Vec_IntFree( pMan->vFlopNums ); 
    pMan->vFlopNums = NULL;

    pMan = Saig_ManHaigRecord( pTemp = pMan, nIters, nSteps, fRetimingOnly, fAddBugs, fUseCnf, fVerbose );
    Aig_ManStop( pTemp );

    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}
*/

/**Function*************************************************************

  Synopsis    [Performs random simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSeqSim(pNtk *Abc_Ntk_t, nFrames int32, nWords int32, TimeOut int32, fNew int32, fMiter int32, fVerbose int32, pFileSim string) int32 {
    var pMan *Aig_Man_t;
    var pCex *Abc_Cex_t;
    var status int32;
    var RetValue int32 = -1;
    var clk abctime  = Abc_Clock();
    if Abc_NtkGetChoiceNum(pNtk) {
        Abc_Print( 1, "Removing %d choices from the AIG.\n", Abc_NtkGetChoiceNum(pNtk) );
        Abc_AigCleanup((Abc_Aig_t *)pNtk.pManFunc);
    }
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if fNew != 0 {
        var pGia *Gia_Man_t;
        var Pars Gia_ParSim_t;
        var pPars *Gia_ParSim_t = &Pars;
        Gia_ManSimSetDefaultParams( pPars );
        pPars.nWords = nWords;
        pPars.nIters = nFrames;
        pPars.TimeLimit = TimeOut;
        pPars.fCheckMiter = fMiter;
        pPars.fVerbose = fVerbose;
        pGia = Gia_ManFromAig( pMan );
        if Gia_ManSimSimulate( pGia, pPars ) { 
            if pGia.pCexSeq {
                Abc_Print( 1, "Simulation of %d frames with %d words asserted output %d in frame %d. ", 
                    nFrames, nWords, pGia.pCexSeq.iPo, pGia.pCexSeq.iFrame );
                status = Saig_ManVerifyCex( pMan, pGia.pCexSeq );
                if status == 0 {
                    Abc_Print( 1, "Abc_NtkDarSeqSim(): Counter-example verification has FAILED.\n" ); }
            }
            ABC_FREE( pNtk.pModel );
            ABC_FREE( pNtk.pSeqModel );
            pNtk.pSeqModel = pGia.pCexSeq; pGia.pCexSeq = nil;
            RetValue = 0;
        
        } else {
            Abc_Print( 1, "Simulation of %d frames with %d words did not assert the outputs.    ", 
                nFrames, nWords );
        }
        Gia_ManStop( pGia );
    
    } else { // comb/seq simulator
    
        var pSml *Fra_Sml_t;
        if pFileSim != nil {
            assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0");
            pSml = Fra_SmlSimulateCombGiven( pMan, pFileSim, fMiter, fVerbose );
        
        } else if Abc_NtkLatchNum(pNtk) == 0 {
            pSml = Fra_SmlSimulateComb( pMan, nWords, fMiter );
        } else {
            pSml = Fra_SmlSimulateSeq( pMan, 0, nFrames, nWords, fMiter ); }
        if pSml.fNonConstOut {
            pCex = Fra_SmlGetCounterExample( pSml );
            if pCex != nil {
                Abc_Print( 1, "Simulation of %d frame%s with %d word%s asserted output %d in frame %d. ", 
                    pSml.nFrames, pSml.nFrames == 1 ? "": "s", 
                    pSml.nWordsFrame, pSml.nWordsFrame == 1 ? "": "s", 
                    pCex.iPo, pCex.iFrame );
                status = Saig_ManVerifyCex( pMan, pCex );
                if status == 0 {
                    Abc_Print( 1, "Abc_NtkDarSeqSim(): Counter-example verification has FAILED.\n" ); }
            }
            ABC_FREE( pNtk.pModel );
            ABC_FREE( pNtk.pSeqModel );
            pNtk.pSeqModel = pCex;
            RetValue = 0;
        
        } else {
            Abc_Print( 1, "Simulation of %d frames with %d words did not assert the outputs.    ", 
                nFrames, nWords );
        }
        Fra_SmlStop( pSml );
    }
    ABC_PRT( "Time", Abc_Clock() - clk );
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Performs random simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSeqSim3(pNtk *Abc_Ntk_t, pPars *Ssw_RarPars_t) int32 {
    var pMan *Aig_Man_t;
    var status int32;
    var RetValue int32 = -1;
//    abctime clk = Abc_Clock();
    if Abc_NtkGetChoiceNum(pNtk) {
        Abc_Print( 1, "Removing %d choices from the AIG.\n", Abc_NtkGetChoiceNum(pNtk) );
        Abc_AigCleanup((Abc_Aig_t *)pNtk.pManFunc);
    }
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if Ssw_RarSimulate( pMan, pPars ) == 0 { 
        if pMan.pSeqModel {
//            Abc_Print( 1, "Simulation of %d frames with %d words asserted output %d in frame %d. ", 
//                nFrames, nWords, pMan->pSeqModel->iPo, pMan->pSeqModel->iFrame );
            status = Saig_ManVerifyCex( pMan, pMan.pSeqModel );
            if status == 0 {
                Abc_Print( 1, "Abc_NtkDarSeqSim(): Counter-example verification has FAILED.\n" ); }
        }
        ABC_FREE( pNtk.pModel );
        ABC_FREE( pNtk.pSeqModel );
        pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
        RetValue = 0;
    
    } else {
//        Abc_Print( 1, "Simulation of %d frames with %d words did not assert the outputs.    ", 
//            nFrames, nWords );
    }
    if pNtk.vSeqModelVec {
        Vec_PtrFreeFree( pNtk.vSeqModelVec ); }
    pNtk.vSeqModelVec = pMan.vSeqModelVec;  pMan.vSeqModelVec = nil;
//    ABC_PRT( "Time", Abc_Clock() - clk );
    pNtk.pData = pMan.pData; pMan.pData = nil;
    Aig_ManStop( pMan );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarClau(pNtk *Abc_Ntk_t, nFrames int32, nPref int32, nClauses int32, nLutSize int32, nLevels int32, nCutsMax int32, nBatches int32, fStepUp int32, fBmc int32, fRefs int32, fTarget int32, fVerbose int32, fVeryVerbose int32) int32 {
    extern int Fra_Clau( Aig_Man_t * pMan, int nIters, int fVerbose, int fVeryVerbose );
    extern int Fra_Claus( Aig_Man_t * pAig, int nFrames, int nPref, int nClauses, int nLutSize, int nLevels, int nCutsMax, int nBatches, int fStepUp, int fBmc, int fRefs, int fTarget, int fVerbose, int fVeryVerbose );
    var pMan *Aig_Man_t;
    if (fTarget != 0) && Abc_NtkPoNum(pNtk) != 1 {
        Abc_Print( 1, "The number of outputs should be 1.\n" );
        return 1;
    }
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return 1; }
//    Aig_ManReduceLachesCount( pMan );
    if pMan.vFlopNums {
        Vec_IntFree( pMan.vFlopNums ); } 
    pMan.vFlopNums = nil;

//    Fra_Clau( pMan, nStepsMax, fVerbose, fVeryVerbose );
    Fra_Claus( pMan, nFrames, nPref, nClauses, nLutSize, nLevels, nCutsMax, nBatches, fStepUp, fBmc, fRefs, fTarget, fVerbose, fVeryVerbose );
    Aig_ManStop( pMan );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Performs targe enlargement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarEnlarge(pNtk *Abc_Ntk_t, nFrames int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pMan = Aig_ManFrames( pTemp = pMan, nFrames, 0, 1, 1, 1, nil );
    Aig_ManStop( pTemp );
    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs targe enlargement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarTempor(pNtk *Abc_Ntk_t, nFrames int32, TimeOut int32, nConfLimit int32, fUseBmc int32, fUseTransSigs int32, fVerbose int32, fVeryVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Saig_ManTempor( Aig_Man_t * pAig, int nFrames, int TimeOut, int nConfLimit, int fUseBmc, int fUseTransSigs, int fVerbose, int fVeryVerbose );
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pTemp = Saig_ManTempor( pMan, nFrames, TimeOut, nConfLimit, fUseBmc, fUseTransSigs, fVerbose, fVeryVerbose );
    Aig_ManStop( pMan );
    if pTemp == nil {
        return Abc_NtkDup( pNtk ); }
    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pTemp );
    Aig_ManStop( pTemp );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs induction for property only.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarInduction(pNtk *Abc_Ntk_t, nTimeOut int32, nFramesMax int32, nConfMax int32, fUnique int32, fUniqueAll int32, fGetCex int32, fVerbose int32, fVeryVerbose int32) int32 { 
    var pMan *Aig_Man_t;
    var clkTotal abctime  = Abc_Clock();
    var RetValue int32;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return -1; }
    RetValue = Saig_ManInduction( pMan, nTimeOut, nFramesMax, nConfMax, fUnique, fUniqueAll, fGetCex, fVerbose, fVeryVerbose );
    if RetValue == 1 {
        Abc_Print( 1, "Networks are equivalent.  " );
ABC_PRT( "Time", Abc_Clock() - clkTotal );
    
    } else if RetValue == 0 {
        Abc_Print( 1, "Networks are NOT EQUIVALENT.  " );
ABC_PRT( "Time", Abc_Clock() - clkTotal );
    
    } else {
        Abc_Print( 1, "Networks are UNDECIDED.  " );
ABC_PRT( "Time", Abc_Clock() - clkTotal );
    }
    if fGetCex != 0 {
        ABC_FREE( pNtk.pModel );
        ABC_FREE( pNtk.pSeqModel );
        pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
    }
    Aig_ManStop( pMan );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Interplates two networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInterOne(pNtkOn *Abc_Ntk_t, pNtkOff *Abc_Ntk_t, fRelation int32, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Aig_ManInter( Aig_Man_t * pManOn, Aig_Man_t * pManOff, int fRelation, int fVerbose );
    var pNtkAig *Abc_Ntk_t;
    var pManOn, pManOff, pManAig *Aig_Man_t;
    if Abc_NtkCoNum(pNtkOn) != 1 || Abc_NtkCoNum(pNtkOff) != 1 {
        Abc_Print( 1, "Currently works only for single-output networks.\n" );
        return nil;
    }
    if Abc_NtkCiNum(pNtkOn) != Abc_NtkCiNum(pNtkOff) {
        Abc_Print( 1, "The number of PIs should be the same.\n" );
        return nil;
    }
    // create internal AIGs
    pManOn = Abc_NtkToDar( pNtkOn, 0, 0 );
    if pManOn == nil {
        return nil; }
    pManOff = Abc_NtkToDar( pNtkOff, 0, 0 );
    if pManOff == nil {
        return nil; }
    // derive the interpolant
    pManAig = Aig_ManInter( pManOn, pManOff, fRelation, fVerbose );
    if pManAig == nil {
        Abc_Print( 1, "Interpolant computation failed.\n" );
        return nil;
    }
    Aig_ManStop( pManOn );
    Aig_ManStop( pManOff );
    // for the relation, add an extra input
    if fRelation != 0 {
        var pObj *Abc_Obj_t;
        pObj = Abc_NtkCreatePi( pNtkOff );
        Abc_ObjAssignName( pObj, "New", nil );
    }
    // create logic network
    pNtkAig = Abc_NtkFromDar( pNtkOff, pManAig );
    Aig_ManStop( pManAig );
    return pNtkAig;
}
func Gia_ManInterOne(pNtkOn *Gia_Man_t, pNtkOff *Gia_Man_t, fVerbose int32) *Gia_Man_t {
    extern Aig_Man_t * Aig_ManInter( Aig_Man_t * pManOn, Aig_Man_t * pManOff, int fRelation, int fVerbose );
    var pNtkAig *Gia_Man_t;
    var pManOn, pManOff, pManAig *Aig_Man_t;
    assert(Gia_ManCiNum(pNtkOn)  == Gia_ManCiNum(pNtkOff), "Gia_ManCiNum(pNtkOn)  == Gia_ManCiNum(pNtkOff)");
    assert(Gia_ManCoNum(pNtkOn)  == 1, "Gia_ManCoNum(pNtkOn)  == 1");
    assert(Gia_ManCoNum(pNtkOff) == 1, "Gia_ManCoNum(pNtkOff) == 1");
    // create internal AIGs
    pManOn = Gia_ManToAigSimple( pNtkOn );
    if pManOn == nil {
        return nil; }
    pManOff = Gia_ManToAigSimple( pNtkOff );
    if pManOff == nil {
        return nil; }
    // derive the interpolant
    pManAig = Aig_ManInter( pManOn, pManOff, 0, fVerbose );
    if pManAig == nil {
        Abc_Print( 1, "Interpolant computation failed.\n" );
        return nil;
    }
    Aig_ManStop( pManOn );
    Aig_ManStop( pManOff );
    // create logic network
    pNtkAig = Gia_ManFromAigSimple( pManAig );
    Aig_ManStop( pManAig );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Fast interpolation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInterFast(pNtkOn *Abc_Ntk_t, pNtkOff *Abc_Ntk_t, fVerbose int32) {
    extern void Aig_ManInterFast( Aig_Man_t * pManOn, Aig_Man_t * pManOff, int fVerbose );
    var pManOn, pManOff *Aig_Man_t;
    // create internal AIGs
    pManOn = Abc_NtkToDar( pNtkOn, 0, 0 );
    if pManOn == nil {
        return; }
    pManOff = Abc_NtkToDar( pNtkOff, 0, 0 );
    if pManOff == nil {
        return; }
    Aig_ManInterFast( pManOn, pManOff, fVerbose );
    Aig_ManStop( pManOn );
    Aig_ManStop( pManOff );
}

var timeCnf abctime;
var timeSat abctime;
var timeInt abctime;

/**Function*************************************************************

  Synopsis    [Interplates two networks.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInter(pNtkOn *Abc_Ntk_t, pNtkOff *Abc_Ntk_t, fRelation int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkOn1, pNtkOff1, pNtkInter1, pNtkInter *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i int32; //, clk = Abc_Clock();
    if Abc_NtkCoNum(pNtkOn) != Abc_NtkCoNum(pNtkOff) {
        Abc_Print( 1, "Currently works only for networks with equal number of POs.\n" );
        return nil;
    }
    // compute the fast interpolation time
//    Abc_NtkInterFast( pNtkOn, pNtkOff, fVerbose );
    // consider the case of one output
    if Abc_NtkCoNum(pNtkOn) == 1 {
        return Abc_NtkInterOne( pNtkOn, pNtkOff, fRelation, fVerbose ); }
    // start the new network
    pNtkInter = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtkInter.pName = Extra_UtilStrsav(pNtkOn.pName);
    Abc_NtkForEachPi( pNtkOn, pObj, i )
        Abc_NtkDupObj( pNtkInter, pObj, 1 );
    // process each POs separately
timeCnf = 0;
timeSat = 0;
timeInt = 0;
    Abc_NtkForEachCo( pNtkOn, pObj, i )
    {
        pNtkOn1 = Abc_NtkCreateCone( pNtkOn, Abc_ObjFanin0(pObj), Abc_ObjName(pObj), 1 );
        if Abc_ObjFaninC0(pObj) {
            Abc_ObjXorFaninC( Abc_NtkPo(pNtkOn1, 0), 0 ); }

        pObj   = Abc_NtkCo(pNtkOff, i);
        pNtkOff1 = Abc_NtkCreateCone( pNtkOff, Abc_ObjFanin0(pObj), Abc_ObjName(pObj), 1 );
        if Abc_ObjFaninC0(pObj) {
            Abc_ObjXorFaninC( Abc_NtkPo(pNtkOff1, 0), 0 ); }

        if Abc_NtkNodeNum(pNtkOn1) == 0 {
            pNtkInter1 = Abc_NtkDup( pNtkOn1 );
        } else if Abc_NtkNodeNum(pNtkOff1) == 0 {
            pNtkInter1 = Abc_NtkDup( pNtkOff1 );
            Abc_ObjXorFaninC( Abc_NtkPo(pNtkInter1, 0), 0 );
        
        } else {
            pNtkInter1 = Abc_NtkInterOne( pNtkOn1, pNtkOff1, 0, fVerbose ); }
        if pNtkInter1 {
            Abc_NtkAppend( pNtkInter, pNtkInter1, 1 );
            Abc_NtkDelete( pNtkInter1 );
        }

        Abc_NtkDelete( pNtkOn1 );
        Abc_NtkDelete( pNtkOff1 );
    }
//    ABC_PRT( "CNF", timeCnf );
//    ABC_PRT( "SAT", timeSat );
//    ABC_PRT( "Int", timeInt );
//    ABC_PRT( "Slow interpolation time", Abc_Clock() - clk );

    // return the network
    if !Abc_NtkCheck( pNtkInter ) {
        Abc_Print( 1, "Abc_NtkAttachBottom(): Network check has failed.\n" ); }
    return pNtkInter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintSccs(pNtk *Abc_Ntk_t, fVerbose int32) {
//    extern Vec_Ptr_t * Aig_ManRegPartitionLinear( Aig_Man_t * pAig, int nPartSize );
    var pMan *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return; }
    Aig_ManComputeSccs( pMan );
//    Aig_ManRegPartitionLinear( pMan, 1000 );
    Aig_ManStop( pMan );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarPrintCone(pNtk *Abc_Ntk_t) int32 {
    extern void Saig_ManPrintCones( Aig_Man_t * pAig );
    var pMan *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return 0; }
    assert(Aig_ManRegNum(pMan) > 0, "Aig_ManRegNum(pMan) > 0");
    Saig_ManPrintCones( pMan );
    Aig_ManStop( pMan );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Abc_NtkBalanceExor(pNtk *Abc_Ntk_t, fUpdateLevel int32, fVerbose int32) *Abc_Ntk_t {
    extern void Dar_BalancePrintStats( Aig_Man_t * p );
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;//, * pTemp2;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // derive AIG with EXORs
    pMan = Abc_NtkToDar( pNtk, 1, 0 );
    if pMan == nil {
        return nil; }
//    Aig_ManPrintStats( pMan );
    if fVerbose != 0 {
        Dar_BalancePrintStats( pMan ); }
    // perform balancing
    pTemp = Dar_ManBalance( pMan, fUpdateLevel );
//    Aig_ManPrintStats( pTemp );
    // create logic network
    pNtkAig = Abc_NtkFromDar( pNtk, pTemp );
    Aig_ManStop( pTemp );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPhaseAbstract(pNtk *Abc_Ntk_t, nFrames int32, nPref int32, fIgnore int32, fPrint int32, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Saig_ManPhaseAbstract( Aig_Man_t * p, Vec_Int_t * vInits, int nFrames, int nPref, int fIgnore, int fPrint, int fVerbose );
    var vInits *Vec_Int_t;
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    vInits = Abc_NtkGetLatchValues(pNtk);
    pMan = Saig_ManPhaseAbstract( pTemp = pMan, vInits, nFrames, nPref, fIgnore, fPrint, fVerbose );
    Vec_IntFree( vInits );
    Aig_ManStop( pTemp );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromAigPhase( pMan );
//    pNtkAig->pName = Extra_UtilStrsav(pNtk->pName);
//    pNtkAig->pSpec = Extra_UtilStrsav(pNtk->pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPhaseFrameNum(pNtk *Abc_Ntk_t) int32 {
    extern int Saig_ManPhaseFrameNum( Aig_Man_t * p, Vec_Int_t * vInits );
    var vInits *Vec_Int_t;
    var pMan *Aig_Man_t;
    var nFrames int32;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return 1; }
    vInits = Abc_NtkGetLatchValues(pNtk);
    nFrames = Saig_ManPhaseFrameNum( pMan, vInits );
    Vec_IntFree( vInits );
    Aig_ManStop( pMan );
    return nFrames;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSynchOne(pNtk *Abc_Ntk_t, nWords int32, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Saig_SynchSequenceApply( Aig_Man_t * pAig, int nWords, int fVerbose );
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pMan = Saig_SynchSequenceApply( pTemp = pMan, nWords, fVerbose );
    Aig_ManStop( pTemp );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarSynch(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, nWords int32, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Saig_Synchronize( Aig_Man_t * pAig1, Aig_Man_t * pAig2, int nWords, int fVerbose );
    var pNtkAig *Abc_Ntk_t;
    var pMan1, pMan2, pMan *Aig_Man_t;
    pMan1 = Abc_NtkToDar( pNtk1, 0, 1 );
    if pMan1 == nil {
        return nil; }
    pMan2 = Abc_NtkToDar( pNtk2, 0, 1 );
    if pMan2 == nil {
        Aig_ManStop( pMan1 );
        return nil;
    }
    pMan = Saig_Synchronize( pMan1, pMan2, nWords, fVerbose );
    Aig_ManStop( pMan1 );
    Aig_ManStop( pMan2 );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromAigPhase( pMan );
//    pNtkAig->pName = Extra_UtilStrsav("miter");
//    pNtkAig->pSpec = NULL;
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarClockGate(pNtk *Abc_Ntk_t, pCare *Abc_Ntk_t, pPars *Cgt_Par_t) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan1 *Aig_Man_t;
    var pMan2 *Aig_Man_t = nil;
    var pMan *Aig_Man_t;
    pMan1 = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan1 == nil {
        return nil; }
    if pCare != nil {
        pMan2 = Abc_NtkToDar( pCare, 0, 0 );
        if pMan2 == nil {
            Aig_ManStop( pMan1 );
            return nil;
        }
    }
    pMan = Cgt_ClockGating( pMan1, pMan2, pPars );
    Aig_ManStop( pMan1 );
    if pMan2 != nil {
        Aig_ManStop( pMan2 ); }
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarExtWin(pNtk *Abc_Ntk_t, nObjId int32, nDist int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan1, pMan *Aig_Man_t;
    var pObj *Aig_Obj_t;
    pMan1 = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan1 == nil {
        return nil; }
    if nObjId == -1 {
        pObj = Saig_ManFindPivot( pMan1 );
        Abc_Print( 1, "Selected object %d as a window pivot.\n", pObj.Id );
    
    } else {
        if nObjId >= Aig_ManObjNumMax(pMan1) {
            Aig_ManStop( pMan1 );
            Abc_Print( 1, "The ID is too large.\n" );
            return nil;
        }
        pObj = Aig_ManObj( pMan1, nObjId );
        if pObj == nil {
            Aig_ManStop( pMan1 );
            Abc_Print( 1, "Object with ID %d does not exist.\n", nObjId );
            return nil;
        }
        if !Saig_ObjIsLo(pMan1, pObj) && !Aig_ObjIsNode(pObj) {
            Aig_ManStop( pMan1 );
            Abc_Print( 1, "Object with ID %d is not a node or reg output.\n", nObjId );
            return nil;
        }
    }
    pMan = Saig_ManWindowExtract( pMan1, pObj, nDist );
    Aig_ManStop( pMan1 );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pNtk.pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarInsWin(pNtk *Abc_Ntk_t, pCare *Abc_Ntk_t, nObjId int32, nDist int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan1 *Aig_Man_t;
    var pMan2 *Aig_Man_t = nil;
    var pMan *Aig_Man_t;
    var pObj *Aig_Obj_t;
    pMan1 = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan1 == nil {
        return nil; }
    if nObjId == -1 {
        pObj = Saig_ManFindPivot( pMan1 );
        Abc_Print( 1, "Selected object %d as a window pivot.\n", pObj.Id );
    
    } else {
        if nObjId >= Aig_ManObjNumMax(pMan1) {
            Aig_ManStop( pMan1 );
            Abc_Print( 1, "The ID is too large.\n" );
            return nil;
        }
        pObj = Aig_ManObj( pMan1, nObjId );
        if pObj == nil {
            Aig_ManStop( pMan1 );
            Abc_Print( 1, "Object with ID %d does not exist.\n", nObjId );
            return nil;
        }
        if !Saig_ObjIsLo(pMan1, pObj) && !Aig_ObjIsNode(pObj) {
            Aig_ManStop( pMan1 );
            Abc_Print( 1, "Object with ID %d is not a node or reg output.\n", nObjId );
            return nil;
        }
    }
    if pCare != nil {
        pMan2 = Abc_NtkToDar( pCare, 0, 0 );
        if pMan2 == nil {
            Aig_ManStop( pMan1 );
            return nil;
        }
    }
    pMan = Saig_ManWindowInsert( pMan1, pObj, nDist, pMan2 );
    Aig_ManStop( pMan1 );
    if pMan2 != nil {
        Aig_ManStop( pMan2 ); }
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromDarSeqSweep( pNtk, pMan );
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarFrames(pNtk *Abc_Ntk_t, nPrefix int32, nFrames int32, fInit int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pMan = Saig_ManTimeframeSimplify( pTemp = pMan, nPrefix, nFrames, fInit, fVerbose );
    Aig_ManStop( pTemp );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pNtk.pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs phase abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarCleanupAig(pNtk *Abc_Ntk_t, fCleanupPis int32, fCleanupPos int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan *Aig_Man_t;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    if fCleanupPis != 0 {
        var Temp int32  = Aig_ManCiCleanup( pMan );
        if fVerbose != 0 {
            Abc_Print( 1, "Cleanup removed %d primary inputs without fanout.\n", Temp ); }                                                                     
    }
    if fCleanupPos != 0 {
        var Temp int32  = Aig_ManCoCleanup( pMan );
        if fVerbose != 0 {
            Abc_Print( 1, "Cleanup removed %d primary outputs driven by const-0.\n", Temp ); }                                                                     
    }
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pNtk.pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarReach(pNtk *Abc_Ntk_t, pPars *Saig_ParBbr_t) int32 {
    var pMan *Aig_Man_t;
    var RetValue int32;
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return -1; }
    RetValue = Aig_ManVerifyUsingBdds( pMan, pPars );
    ABC_FREE( pNtk.pModel );
    ABC_FREE( pNtk.pSeqModel );
    pNtk.pSeqModel = pMan.pSeqModel; pMan.pSeqModel = nil;
    Aig_ManStop( pMan );
    return RetValue;
}

//ABC_NAMESPACE_IMPL_END

//#include "map/amap/amap.h"
//#include "map/mio/mio.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Amap_ManProduceNetwork(pNtk *Abc_Ntk_t, vMapping *Vec_Ptr_t) *Abc_Ntk_t {
//    extern void * Abc_FrameReadLibGen();
    var pLib *Mio_Library_t  = (Mio_Library_t *)Abc_FrameReadLibGen();
    var pRes *Amap_Out_t;
    var vNodesNew *Vec_Ptr_t;
    var pNtkNew *Abc_Ntk_t;
    var pNodeNew, pFaninNew *Abc_Obj_t;
    var i, k, iPis, iPos, nDupGates int32;
    // make sure gates exist in the current library
    Vec_PtrForEachEntry( Amap_Out_t *, vMapping, pRes, i )
        if pRes.pName && Mio_LibraryReadGateByName( pLib, pRes.pName, nil ) == nil {
            Abc_Print( 1, "Current library does not contain gate \"%s\".\n", pRes.pName );
            return nil;
        }
    // create the network
    pNtkNew = Abc_NtkStartFrom( pNtk, ABC_NTK_LOGIC, ABC_FUNC_MAP );
    pNtkNew.pManFunc = pLib;
    iPis = iPos = 0;
    vNodesNew = Vec_PtrAlloc( Vec_PtrSize(vMapping) );
    Vec_PtrForEachEntry( Amap_Out_t *, vMapping, pRes, i )
    {
        if pRes.Type == -1 {
            pNodeNew = Abc_NtkCi( pNtkNew, iPis++ );
        } else if pRes.Type == 1 {
            pNodeNew = Abc_NtkCo( pNtkNew, iPos++ );
        } else {
            pNodeNew = Abc_NtkCreateNode( pNtkNew );
            pNodeNew.pData = Mio_LibraryReadGateByName( pLib, pRes.pName, nil );
        }
        for  k = 0; k < pRes.nFans; k++  {
            pFaninNew = (Abc_Obj_t *)Vec_PtrEntry( vNodesNew, pRes.pFans[k] );
            Abc_ObjAddFanin( pNodeNew, pFaninNew );
        }
        Vec_PtrPush( vNodesNew, pNodeNew );
    }
    Vec_PtrFree( vNodesNew );
    assert(iPis == Abc_NtkCiNum(pNtkNew), "iPis == Abc_NtkCiNum(pNtkNew)");
    assert(iPos == Abc_NtkCoNum(pNtkNew), "iPos == Abc_NtkCoNum(pNtkNew)");
    // decouple the PO driver nodes to reduce the number of levels
    nDupGates = Abc_NtkLogicMakeSimpleCos( pNtkNew, 0 );
//    if ( nDupGates && Map_ManReadVerbose(pMan) )
//        Abc_Print( 1, "Duplicated %d gates to decouple the CO drivers.\n", nDupGates );
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Gives the current ABC network to AIG manager for processing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarAmap(pNtk *Abc_Ntk_t, pPars *Amap_Par_t) *Abc_Ntk_t {
    extern Vec_Ptr_t * Amap_ManTest( Aig_Man_t * pAig, Amap_Par_t * pPars );
    var vMapping *Vec_Ptr_t;
    var pNtkAig *Abc_Ntk_t  = nil;
    var pMan *Aig_Man_t;
    var pMem *Aig_MmFlex_t;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // convert to the AIG manager
    pMan = Abc_NtkToDarChoices( pNtk );
    if pMan == nil {
        return nil; }

    // perform computation
    vMapping = Amap_ManTest( pMan, pPars );
    Aig_ManStop( pMan );
    if vMapping == nil {
        return nil; }
    pMem = (Aig_MmFlex_t *)Vec_PtrPop( vMapping );
    pNtkAig = Amap_ManProduceNetwork( pNtk, vMapping );
    Aig_MmFlexStop( pMem, 0 );
    Vec_PtrFree( vMapping );

    // make sure everything is okay
    if pNtkAig && !Abc_NtkCheck( pNtkAig ) {
        Abc_Print( 1, "Abc_NtkDar: The network check has failed.\n" );
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarConstr(pNtk *Abc_Ntk_t, nFrames int32, nConfs int32, nProps int32, fStruct int32, fOldAlgo int32, fVerbose int32) {
    var pMan *Aig_Man_t;//, * pMan2;//, * pTemp;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return; }
    if fStruct != 0 {
        Saig_ManDetectConstrTest( pMan );
    } else {
        Saig_ManDetectConstrFuncTest( pMan, nFrames, nConfs, nProps, fOldAlgo, fVerbose ); }
    Aig_ManStop( pMan );
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarOutdec(pNtk *Abc_Ntk_t, nLits int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pMan = Saig_ManDecPropertyOutput( pTemp = pMan, nLits, fVerbose );
    Aig_ManStop( pTemp );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pMan.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pMan.pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarUnfold(pNtk *Abc_Ntk_t, nFrames int32, nConfs int32, nProps int32, fStruct int32, fOldAlgo int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    if fStruct != 0 {
        pMan = Saig_ManDupUnfoldConstrs( pTemp = pMan );
    } else {
        pMan = Saig_ManDupUnfoldConstrsFunc( pTemp = pMan, nFrames, nConfs, nProps, fOldAlgo, fVerbose ); }
    Aig_ManStop( pTemp );
    if pMan == nil {
        return nil; }
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pMan.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pMan.pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarFold(pNtk *Abc_Ntk_t, fCompl int32, fVerbose int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pMan, pTemp *Aig_Man_t;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }
    pMan = Saig_ManDupFoldConstrsFunc( pTemp = pMan, fCompl, fVerbose );
    Aig_ManStop( pTemp );
    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig.pName = Extra_UtilStrsav(pMan.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pMan.pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarConstrProfile(pNtk *Abc_Ntk_t, fVerbose int32) {
    extern int Ssw_ManProfileConstraints( Aig_Man_t * p, int nWords, int nFrames, int fVerbose );
    extern Vec_Int_t * Saig_ManComputeSwitchProbs( Aig_Man_t * p, int nFrames, int nPref, int fProbOne );
    var pMan *Aig_Man_t;
//    Vec_Int_t * vProbOne;
//    Aig_Obj_t * pObj;
//    int i, Entry;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(Abc_NtkConstrNum(pNtk), "Abc_NtkConstrNum(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return; }
    // value in the init state
//    Abc_AigSetNodePhases( pNtk );
/*
    // derive probabilities
    vProbOne = Saig_ManComputeSwitchProbs( pMan, 48, 16, 1 );
    // iterate over the constraint outputs
    Saig_ManForEachPo( pMan, pObj, i )
    {
        Entry = Vec_IntEntry( vProbOne, Aig_ObjId(pObj) );
        if ( i < Saig_ManPoNum(pMan) - Saig_ManConstrNum(pMan) )
            Abc_Print( 1, "Primary output :  ", i );
        else
            Abc_Print( 1, "Constraint %3d :  ", i-(Saig_ManPoNum(pMan) - Saig_ManConstrNum(pMan)) );
        Abc_Print( 1, "ProbOne = %f  ", Abc_Int2Float(Entry) );
        Abc_Print( 1, "AllZeroValue = %d ", Aig_ObjPhase(pObj) );
        Abc_Print( 1, "\n" );
    }
*/
    // double-check
    Ssw_ManProfileConstraints( pMan, 16, 64, 1 );
    Abc_Print( 1, "TwoFrameSatValue = %d.\n", Ssw_ManSetConstrPhases(pMan, 2, nil) );
    // clean up
//    Vec_IntFree( vProbOne );
    Aig_ManStop( pMan );
}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarTest(pNtk *Abc_Ntk_t, Num int32) {
//    extern void Saig_ManDetectConstr( Aig_Man_t * p );
//    extern void Saig_ManDetectConstrFuncTest( Aig_Man_t * p );
//    extern void Saig_ManFoldConstrTest( Aig_Man_t * pAig );
    extern void Llb_ManComputeDomsTest( Aig_Man_t * pAig, int Num );



//    extern void Fsim_ManTest( Aig_Man_t * pAig );
    extern Vec_Int_t * Saig_StrSimPerformMatching( Aig_Man_t * p0, Aig_Man_t * p1, int nDist, int fVerbose, Aig_Man_t ** ppMiter );
//    Vec_Int_t * vPairs;
    var pMan *Aig_Man_t;//, * pMan2;//, * pTemp;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return; }
/*
Aig_ManSetRegNum( pMan, pMan->nRegs );
Aig_ManPrintStats( pMan );
Saig_ManDumpBlif( pMan, "_temp_.blif" );
Aig_ManStop( pMan );
pMan = Saig_ManReadBlif( "_temp_.blif" );
Aig_ManPrintStats( pMan );
*/
/*
    Aig_ManSetRegNum( pMan, pMan->nRegs );
    pTemp = Ssw_SignalCorrespondeceTestPairs( pMan );
    Aig_ManStop( pTemp );
*/

/*
//    Ssw_SecSpecialMiter( pMan, NULL, 2, 1 );
    pMan2 = Aig_ManDupSimple(pMan);
    vPairs = Saig_StrSimPerformMatching( pMan, pMan2, 0, 1, NULL );
    Vec_IntFree( vPairs );
    Aig_ManStop( pMan );
    Aig_ManStop( pMan2 );
*/
//    Ioa_WriteAigerBufferTest( pMan, "test.aig", 0, 0 );
//    Saig_ManFoldConstrTest( pMan );
    {
    extern void Saig_ManBmcSectionsTest( Aig_Man_t * p );
    extern void Saig_ManBmcTerSimTest( Aig_Man_t * p );
    extern void Saig_ManBmcSupergateTest( Aig_Man_t * p );
    extern void Saig_ManBmcMappingTest( Aig_Man_t * p );
//    Saig_ManBmcSectionsTest( pMan );
//    Saig_ManBmcTerSimTest( pMan );
//    Saig_ManBmcSupergateTest( pMan );
//    Saig_ManBmcMappingTest( pMan );
    }

    {
//        void Pdr_ManEquivClasses( Aig_Man_t * pMan );
//        Pdr_ManEquivClasses( pMan );
    }

//    Llb_ManComputeDomsTest( pMan, Num );
    {
        extern void Llb_ManMinCutTest( Aig_Man_t * pMan, int Num );
        extern void Llb_BddStructAnalysis( Aig_Man_t * pMan );
        extern void Llb_NonlinExperiment( Aig_Man_t * pAig, int Num );
//        Llb_BddStructAnalysis( pMan );
//        Llb_ManMinCutTest( pMan, Num );
//        Llb_NonlinExperiment( pMan, Num );
    }

//    Saig_MvManSimulate( pMan, 1 );
//    Saig_ManDetectConstr( pMan );
//    Saig_ManDetectConstrFuncTest( pMan );

//    Fsim_ManTest( pMan );
    Aig_ManStop( pMan );

}

/**Function*************************************************************

  Synopsis    [Performs BDD-based reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDarTestNtk(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
//    extern Aig_Man_t * Saig_ManDualRail( Aig_Man_t * p, int fMiter );

/*
    extern Aig_Man_t * Ssw_SignalCorrespondeceTestPairs( Aig_Man_t * pAig );

    Abc_Ntk_t * pNtkAig;
    Aig_Man_t * pMan, * pTemp;
    assert( Abc_NtkIsStrash(pNtk) );
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if ( pMan == NULL )
        return NULL;

    Aig_ManSetRegNum( pMan, pMan->nRegs );
    pMan = Ssw_SignalCorrespondeceTestPairs( pTemp = pMan );
    Aig_ManStop( pTemp );
    if ( pMan == NULL )
        return NULL;

    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig->pName = Extra_UtilStrsav(pNtk->pName);
    pNtkAig->pSpec = Extra_UtilStrsav(pNtk->pSpec);
    Aig_ManStop( pMan );
    return pNtkAig;
*/
    var pNtkAig *Abc_Ntk_t;
    var pMan *Aig_Man_t;//, * pTemp;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    pMan = Abc_NtkToDar( pNtk, 0, 1 );
    if pMan == nil {
        return nil; }

/*
    Aig_ManSetRegNum( pMan, pMan->nRegs );
    pMan = Saig_ManDualRail( pTemp = pMan, 1 );
    Aig_ManStop( pTemp );
    if ( pMan == NULL )
        return NULL;

    pNtkAig = Abc_NtkFromAigPhase( pMan );
    pNtkAig->pName = Extra_UtilStrsav(pNtk->pName);
    pNtkAig->pSpec = Extra_UtilStrsav(pNtk->pSpec);
    Aig_ManStop( pMan );
*/

    pNtkAig = Abc_NtkFromDar( pNtk, pMan );
    Aig_ManStop( pMan );

    return pNtkAig;

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//#include "abcDarUnfold2.c"
//ABC_NAMESPACE_IMPL_END

