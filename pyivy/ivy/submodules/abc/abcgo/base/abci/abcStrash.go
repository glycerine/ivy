package abci

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcStrash.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Strashing of the current network.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcStrash.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "bool/dec/dec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Abc_NtkStrashPerform( Abc_Ntk_t * pNtk, Abc_Ntk_t * pNtkNew, int fAllNodes, int fRecord );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reapplies structural hashing to the AIG.]

  Description [Because of the structural hashing, this procedure should not 
  change the number of nodes. It is useful to detect the bugs in the original AIG.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRestrash(pNtk *Abc_Ntk_t, fCleanup int32) *Abc_Ntk_t {
//    extern int timeRetime;
    var vNodes *Vec_Ptr_t;
    var pNtkAig *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, nNodes int32;//, RetValue;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
//timeRetime = Abc_Clock();
    // print warning about choice nodes
    if Abc_NtkGetChoiceNum( pNtk ) {
        fmt.Printf("Warning: The choice nodes in the original AIG are removed by strashing.\n"); }
    // start the new network (constants and CIs of the old network will point to the their counterparts in the new network)
    pNtkAig = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );
    // restrash the nodes (assuming a topological order of the old network)
    vNodes = Abc_NtkDfs( pNtk, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkAig.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
    }
    Vec_PtrFree( vNodes );
    // finalize the network
    Abc_NtkFinalize( pNtk, pNtkAig );
    // print warning about self-feed latches
//    if ( Abc_NtkCountSelfFeedLatches(pNtkAig) )
//        printf( "Warning: The network has %d self-feeding latches.\n", Abc_NtkCountSelfFeedLatches(pNtkAig) );
    // perform cleanup if requested
    if fCleanup && (nNodes = Abc_AigCleanup((Abc_Aig_t *)pNtkAig.pManFunc)) {
//        printf( "Abc_NtkRestrash(): AIG cleanup removed %d nodes (this is a bug).\n", nNodes );
    }
    // duplicate EXDC 
    if pNtk.pExdc {
        pNtkAig.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkStrash: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
//timeRetime = Abc_Clock() - timeRetime;
//    if ( RetValue = Abc_NtkRemoveSelfFeedLatches(pNtkAig) )
//        printf( "Modified %d self-feeding latches. The result may not verify.\n", RetValue );
    return pNtkAig;

}

/**Function*************************************************************

  Synopsis    [Performs structural hashing by generating random number.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRestrashRandom_rec(pNtk *Abc_Ntk_t, pObj *Abc_Obj_t) {
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    if !Abc_ObjIsNode(pObj) {
        return; }
    if rand() & 1 {
        Abc_NtkRestrashRandom_rec( pNtk, Abc_ObjFanin0(pObj) );
        Abc_NtkRestrashRandom_rec( pNtk, Abc_ObjFanin1(pObj) );
    
    } else {
        Abc_NtkRestrashRandom_rec( pNtk, Abc_ObjFanin1(pObj) );
        Abc_NtkRestrashRandom_rec( pNtk, Abc_ObjFanin0(pObj) );
    }
    pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtk.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Reapplies structural hashing to the AIG.]

  Description [Because of the structural hashing, this procedure should not 
  change the number of nodes. It is useful to detect the bugs in the original AIG.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRestrashRandom(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // print warning about choice nodes
    if Abc_NtkGetChoiceNum( pNtk ) {
        fmt.Printf("Warning: The choice nodes in the original AIG are removed by strashing.\n"); }
    // start the new network (constants and CIs of the old network will point to the their counterparts in the new network)
    pNtkAig = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );
    // restrash the nodes (assuming a topological order of the old network)
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NtkRestrashRandom_rec( pNtkAig, Abc_ObjFanin0(pObj) );
    }
    // finalize the network
    Abc_NtkFinalize( pNtk, pNtkAig );
    // duplicate EXDC 
    if pNtk.pExdc {
        pNtkAig.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkStrash: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
    return pNtkAig;

}

/**Function*************************************************************

  Synopsis    [Reapplies structural hashing to the AIG.]

  Description [Because of the structural hashing, this procedure should not 
  change the number of nodes. It is useful to detect the bugs in the original AIG.]
               
  SideEffects []
 
  SeeAlso     []

***********************************************************************/
func Abc_NtkRestrashZero(pNtk *Abc_Ntk_t, fCleanup int32) *Abc_Ntk_t {
//    extern int timeRetime;
    var pNtkAig *Abc_Ntk_t;
    var pObj *Abc_Obj_t;
    var i, nNodes int32;//, RetValue;
    var Counter int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
//timeRetime = Abc_Clock();
    // print warning about choice nodes
    if Abc_NtkGetChoiceNum( pNtk ) {
        fmt.Printf("Warning: The choice nodes in the original AIG are removed by strashing.\n"); }
    // start the new network (constants and CIs of the old network will point to the their counterparts in the new network)
    pNtkAig = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );
    // complement the 1-values registers
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pObj) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        if Abc_LatchIsInitDc(pObj) {
            Counter++;
        } else if Abc_LatchIsInit1(pObj) {
            Abc_ObjFanout0(pObj).pCopy = Abc_ObjNot(Abc_ObjFanout0(pObj).pCopy); }
    }
    }
    if Counter != 0 {
    fmt.Printf("Converting %d flops from don't-care to zero initial value.\n", Counter); }
    // restrash the nodes (assuming a topological order of the old network)
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkAig.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) ); }
    }
    // finalize the network
    Abc_NtkFinalize( pNtk, pNtkAig );
    // complement the 1-valued registers
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtkAig).vBoxes)) && (((pObj) = Abc_NtkBox(pNtkAig, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else
        if Abc_LatchIsInit1(pObj) {
            Abc_ObjXorFaninC( Abc_ObjFanin0(pObj), 0 );
            // if latch has PO as one of its fanouts change latch name
            if Abc_NodeFindCoFanout( Abc_ObjFanout0(pObj) ) {
                Nm_ManDeleteIdName( pObj.pNtk.pManName, Abc_ObjFanout0(pObj).Id );
                Abc_ObjAssignName( Abc_ObjFanout0(pObj), Abc_ObjName(Abc_ObjFanout0(pObj)), "_inv" );
            }
        }
    }
    // set all constant-0 values
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtkAig).vBoxes)) && (((pObj) = Abc_NtkBox(pNtkAig, i)), 1); i++  {           if !Abc_ObjIsLatch(pObj) {} else {
        Abc_LatchSetInit0( pObj ); }
    }

    // print warning about self-feed latches
//    if ( Abc_NtkCountSelfFeedLatches(pNtkAig) )
//        printf( "Warning: The network has %d self-feeding latches.\n", Abc_NtkCountSelfFeedLatches(pNtkAig) );
    // perform cleanup if requested
    if fCleanup && (nNodes = Abc_AigCleanup((Abc_Aig_t *)pNtkAig.pManFunc)) { 
        fmt.Printf("Abc_NtkRestrash(): AIG cleanup removed %d nodes (this is a bug).\n", nNodes); }
    // duplicate EXDC 
    if pNtk.pExdc {
        pNtkAig.pExdc = Abc_NtkDup( pNtk.pExdc ); }
    // transfer name IDs
    if pNtk.vNameIds {
        Abc_NtkTransferNameIds( pNtk, pNtkAig ); }
    if pNtk.vNameIds {
        Abc_NtkUpdateNameIds( pNtkAig ); }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkStrash: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
//timeRetime = Abc_Clock() - timeRetime;
//    if ( RetValue = Abc_NtkRemoveSelfFeedLatches(pNtkAig) )
//        printf( "Modified %d self-feeding latches. The result may not verify.\n", RetValue );
    return pNtkAig;

}

/**Function*************************************************************

  Synopsis    [Transforms logic network into structurally hashed AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStrash(pNtk *Abc_Ntk_t, fAllNodes int32, fCleanup int32, fRecord int32) *Abc_Ntk_t {
    var pNtkAig *Abc_Ntk_t;
    var nNodes int32;
    assert(Abc_NtkIsLogic(pNtk) || Abc_NtkIsStrash(pNtk), "Abc_NtkIsLogic(pNtk) || Abc_NtkIsStrash(pNtk)");
    // consider the special case when the network is already structurally hashed
    if Abc_NtkIsStrash(pNtk) {
        return Abc_NtkRestrash( pNtk, fCleanup ); }
    // convert the node representation in the logic network to the AIG form
    if !Abc_NtkToAig(pNtk) {
        fmt.Printf("Converting to AIGs has failed.\n");
        return nil;
    }
    // perform strashing
//    Abc_NtkCleanCopy( pNtk );
    pNtkAig = Abc_NtkStartFrom( pNtk, ABC_NTK_STRASH, ABC_FUNC_AIG );
    Abc_NtkStrashPerform( pNtk, pNtkAig, fAllNodes, fRecord );
    Abc_NtkFinalize( pNtk, pNtkAig );
    // transfer name IDs
    if pNtk.vNameIds {
        Abc_NtkTransferNameIds( pNtk, pNtkAig ); }
    // print warning about self-feed latches
//    if ( Abc_NtkCountSelfFeedLatches(pNtkAig) )
//        printf( "Warning: The network has %d self-feeding latches.\n", Abc_NtkCountSelfFeedLatches(pNtkAig) );
    // perform cleanup if requested
    if fCleanup != 0 {
        nNodes = Abc_AigCleanup((Abc_Aig_t *)pNtkAig.pManFunc);
    } else {
        nNodes = 0;
    }
//    if ( nNodes )
//        printf( "Warning: AIG cleanup removed %d nodes (this is not a bug).\n", nNodes );
    // duplicate EXDC 
    if pNtk.pExdc {
        pNtkAig.pExdc = Abc_NtkStrash( pNtk.pExdc, fAllNodes, fCleanup, fRecord ); }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkStrash: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Appends the second network to the first.]

  Description [Modifies the first network by adding the logic of the second. 
  Performs structural hashing while appending the networks. Does not change 
  the second network. Returns 0 if the appending failed, 1 otherise.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAppend(pNtk1 *Abc_Ntk_t, pNtk2 *Abc_Ntk_t, fAddPos int32) int32 {
    var pObj *Abc_Obj_t;
    var pName string;
    var i, nNewCis int32;
    // the first network should be an AIG
    assert(Abc_NtkIsStrash(pNtk1), "Abc_NtkIsStrash(pNtk1)");
    assert(Abc_NtkIsLogic(pNtk2) || Abc_NtkIsStrash(pNtk2), "Abc_NtkIsLogic(pNtk2) || Abc_NtkIsStrash(pNtk2)"); 
    if Abc_NtkIsLogic(pNtk2) && !Abc_NtkToAig(pNtk2) {
        fmt.Printf("Converting to AIGs has failed.\n");
        return 0;
    }
    // check that the networks have the same PIs
    // reorder PIs of pNtk2 according to pNtk1
    if !Abc_NtkCompareSignals( pNtk1, pNtk2, 1, 1 ) {
        fmt.Printf("Abc_NtkAppend(): The union of the network PIs is computed (warning).\n"); }
    // perform strashing
    nNewCis = 0;
    Abc_NtkCleanCopy( pNtk2 );
    if Abc_NtkIsStrash(pNtk2) {
        Abc_AigConst1(pNtk2).pCopy = Abc_AigConst1(pNtk1); }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk2)) && (((pObj) = Abc_NtkCi(pNtk2, i)), 1); i++  {
        pName = Abc_ObjName(pObj);
        pObj.pCopy = Abc_NtkFindCi(pNtk1, Abc_ObjName(pObj));
        if pObj.pCopy == nil {
            pObj.pCopy = Abc_NtkDupObj(pNtk1, pObj, 1);
            nNewCis++;
        }
    }
    if nNewCis != 0 {
        fmt.Printf("Warning: Procedure Abc_NtkAppend() added %d new CIs.\n", nNewCis); }
    // add pNtk2 to pNtk1 while strashing
    if Abc_NtkIsLogic(pNtk2) {
        Abc_NtkStrashPerform( pNtk2, pNtk1, 1, 0 );
    } else {
        // inlining of macro Abc_NtkForEachNode
        for  i = 0; (i < Vec_PtrSize((pNtk2).vObjs)) && (((pObj) = Abc_NtkObj(pNtk2, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
            pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtk1.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) ); }
        } }
    // add the COs of the second network
    if fAddPos != 0 {
        // inlining of macro Abc_NtkForEachPo
        for  i = 0; (i < Abc_NtkPoNum(pNtk2)) && (((pObj) = Abc_NtkPo(pNtk2, i)), 1); i++  {
            Abc_NtkDupObj( pNtk1, pObj, 0 );
            Abc_ObjAddFanin( pObj.pCopy, Abc_ObjChild0Copy(pObj) );
            Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(pObj), nil );
        }
    
    } else {
        var pObjOld, pDriverOld, pDriverNew *Abc_Obj_t;
        var fCompl, iNodeId int32;
        // OR the choices
        // inlining of macro Abc_NtkForEachCo
        for  i = 0; (i < Abc_NtkCoNum(pNtk2)) && (((pObj) = Abc_NtkCo(pNtk2, i)), 1); i++  {
            iNodeId = Nm_ManFindIdByNameTwoTypes( pNtk1.pManName, Abc_ObjName(pObj), ABC_OBJ_PO, ABC_OBJ_BI );
//            if ( iNodeId < 0 )
//                continue;
            assert(iNodeId >= 0, "iNodeId >= 0");
            pObjOld = Abc_NtkObj( pNtk1, iNodeId );
            // derive the new driver
            pDriverOld = Abc_ObjChild0( pObjOld );
            pDriverNew = Abc_ObjChild0Copy( pObj );
            pDriverNew = Abc_AigOr( (Abc_Aig_t *)pNtk1.pManFunc, pDriverOld, pDriverNew );
            if Abc_ObjRegular(pDriverOld) == Abc_ObjRegular(pDriverNew) {
                continue; }
            // replace the old driver by the new driver
            fCompl = Abc_ObjRegular(pDriverOld).fPhase ^ Abc_ObjRegular(pDriverNew).fPhase;
            Abc_ObjPatchFanin( pObjOld, Abc_ObjRegular(pDriverOld), Abc_ObjNotCond(Abc_ObjRegular(pDriverNew), fCompl) );
        }
    }
    // make sure that everything is okay
    if !Abc_NtkCheck( pNtk1 ) {
        fmt.Printf("Abc_NtkAppend: The network check has failed.\n");
        return 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prepares the network for strashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkStrashPerform(pNtkOld *Abc_Ntk_t, pNtkNew *Abc_Ntk_t, fAllNodes int32, fRecord int32) {
    var vNodes *Vec_Ptr_t;
    var pNodeOld *Abc_Obj_t;
    var i int32; //, clk = Abc_Clock();
    assert(Abc_NtkIsLogic(pNtkOld), "Abc_NtkIsLogic(pNtkOld)");
    assert(Abc_NtkIsStrash(pNtkNew), "Abc_NtkIsStrash(pNtkNew)");
//    vNodes = Abc_NtkDfs( pNtkOld, fAllNodes );
    vNodes = Abc_NtkDfsIter( pNtkOld, fAllNodes );
//printf( "Nodes = %d. ", Vec_PtrSize(vNodes) );
//ABC_PRT( "Time", Abc_Clock() - clk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNodeOld) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        if Abc_ObjIsBarBuf(pNodeOld) {
            pNodeOld.pCopy = Abc_ObjChild0Copy(pNodeOld);
        } else {
            pNodeOld.pCopy = Abc_NodeStrash( pNtkNew, pNodeOld, fRecord ); }
    }
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Transfers the AIG from one manager into another.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeStrash_rec(pMan *Abc_Aig_t, pObj *Hop_Obj_t) {
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return; }
    Abc_NodeStrash_rec( pMan, Hop_ObjFanin0(pObj) ); 
    Abc_NodeStrash_rec( pMan, Hop_ObjFanin1(pObj) );
    pObj.pData = Abc_AigAnd( pMan, (Abc_Obj_t *)Hop_ObjChild0Copy(pObj), (Abc_Obj_t *)Hop_ObjChild1Copy(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
}

/**Function*************************************************************

  Synopsis    [Strashes one logic node.]

  Description [Assume the network is in the AIG form]
               
  SideEffects []
 
  SeeAlso     []

***********************************************************************/
func Abc_NodeStrash(pNtkNew *Abc_Ntk_t, pNodeOld *Abc_Obj_t, fRecord int32) *Abc_Obj_t {
    var pMan *Hop_Man_t;
    var pRoot *Hop_Obj_t;
    var pFanin *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjIsNode(pNodeOld), "Abc_ObjIsNode(pNodeOld)");
    assert(Abc_NtkHasAig(pNodeOld.pNtk) && !Abc_NtkIsStrash(pNodeOld.pNtk), "Abc_NtkHasAig(pNodeOld.pNtk) && !Abc_NtkIsStrash(pNodeOld.pNtk)");
    // get the local AIG manager and the local root node
    pMan = (Hop_Man_t *)pNodeOld.pNtk.pManFunc;
    pRoot = (Hop_Obj_t *)pNodeOld.pData;
    // check the constant case
    if Abc_NodeIsConst(pNodeOld) || Hop_Regular(pRoot) == Hop_ManConst1(pMan) {
        return Abc_ObjNotCond( Abc_AigConst1(pNtkNew), Hop_IsComplement(pRoot) ); }
    // perform special case-strashing using the record of AIG subgraphs
/*
    if ( fRecord && Abc_NtkRecIsRunning() && Abc_ObjFaninNum(pNodeOld) > 2 && Abc_ObjFaninNum(pNodeOld) <= Abc_NtkRecVarNum() )
    {
        extern Vec_Int_t * Abc_NtkRecMemory();
        extern int Abc_NtkRecStrashNode( Abc_Ntk_t * pNtkNew, Abc_Obj_t * pObj, unsigned * pTruth, int nVars );
        int nVars = Abc_NtkRecVarNum();
        Vec_Int_t * vMemory = Abc_NtkRecMemory();
        unsigned * pTruth = Hop_ManConvertAigToTruth( pMan, Hop_Regular(pRoot), nVars, vMemory, 0 );
        assert( Extra_TruthSupportSize(pTruth, nVars) == Abc_ObjFaninNum(pNodeOld) ); // should be swept
        if ( Hop_IsComplement(pRoot) )
            Extra_TruthNot( pTruth, pTruth, nVars );
        if ( Abc_NtkRecStrashNode( pNtkNew, pNodeOld, pTruth, nVars ) )
            return pNodeOld->pCopy;
    }
*/
    // set elementary variables
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNodeOld)) && (((pFanin) = Abc_ObjFanin(pNodeOld, i)), 1); i++  {
        Hop_IthVar(pMan, i).pData = pFanin.pCopy;
    }
    // strash the AIG of this node
    Abc_NodeStrash_rec( (Abc_Aig_t *)pNtkNew.pManFunc, Hop_Regular(pRoot) );
    Hop_ConeUnmark_rec( Hop_Regular(pRoot) );
    // return the final node
    return Abc_ObjNotCond( (Abc_Obj_t *)Hop_Regular(pRoot).pData, Hop_IsComplement(pRoot) );
}







/**Function*************************************************************

  Synopsis    [Copies the topmost levels of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTopmost_rec(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t, LevelCut int32) *Abc_Obj_t {
    assert(!Abc_ObjIsComplement(pNode), "!Abc_ObjIsComplement(pNode)");
    if pNode.pCopy {
        return pNode.pCopy; }
    if pNode.Level <= uint32(LevelCut) {
        return pNode.pCopy = Abc_NtkCreatePi( pNtkNew ); }
    Abc_NtkTopmost_rec( pNtkNew, Abc_ObjFanin0(pNode), LevelCut );
    Abc_NtkTopmost_rec( pNtkNew, Abc_ObjFanin1(pNode), LevelCut );
    return pNode.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode) );
}

/**Function*************************************************************

  Synopsis    [Copies the topmost levels of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTopmost(pNtk *Abc_Ntk_t, nLevels int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObjNew, pObj *Abc_Obj_t;
    var LevelCut, i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // get the cutoff level
    LevelCut = Abc_MaxInt( 0, Abc_AigLevel(pNtk) - nLevels );
    // start the network
    pNtkNew = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtkNew.pName = Extra_UtilStrsav(pNtk.pName);
    // create PIs below the cut and nodes above the cut
    Abc_NtkCleanCopy( pNtk );
    Abc_AigConst1(pNtk).pCopy = Abc_AigConst1(pNtkNew);
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObjNew = Abc_NtkTopmost_rec( pNtkNew, Abc_ObjFanin0(pObj), LevelCut );
        pObjNew = Abc_ObjNotCond( pObjNew, Abc_ObjFaninC0(pObj) );
        Abc_ObjAddFanin( (pObj.pCopy = Abc_NtkCreatePo(pNtkNew)), pObjNew );
    }
    // add the PO node and name
    Abc_NtkAddDummyPiNames( pNtkNew );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_ObjAssignName( pObj.pCopy, Abc_ObjName(pObj), nil );
    }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkTopmost: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}


/**Function*************************************************************

  Synopsis    [Copies the bottommost levels of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkBottommost_rec(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t, LevelCut int32) *Abc_Obj_t {
    assert(!Abc_ObjIsComplement(pNode), "!Abc_ObjIsComplement(pNode)");
    if pNode.pCopy {
        return pNode.pCopy; }
    Abc_NtkBottommost_rec( pNtkNew, Abc_ObjFanin0(pNode), LevelCut );
    Abc_NtkBottommost_rec( pNtkNew, Abc_ObjFanin1(pNode), LevelCut );
    if pNode.Level > uint32(LevelCut) {
        return nil; }
    return pNode.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkNew.pManFunc, Abc_ObjChild0Copy(pNode), Abc_ObjChild1Copy(pNode) );
}

/**Function*************************************************************

  Synopsis    [Copies the topmost levels of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkBottommost(pNtk *Abc_Ntk_t, nLevels int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var pObj, pObjNew *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(nLevels >= 0, "nLevels >= 0");
    // start the network
    pNtkNew = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtkNew.pName = Extra_UtilStrsav(pNtk.pName);
    // create PIs below the cut and nodes above the cut
    Abc_NtkCleanCopy( pNtk );
    Abc_AigConst1(pNtk).pCopy = Abc_AigConst1(pNtkNew);
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.pCopy = Abc_NtkCreatePi( pNtkNew );
    }
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        Abc_NtkBottommost_rec( pNtkNew, Abc_ObjFanin0(pObj), nLevels );
    }
    // add POs to nodes without fanout
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtkNew).vObjs)) && (((pObjNew) = Abc_NtkObj(pNtkNew, i)), 1); i++  {           if (pObjNew) == nil || !Abc_ObjIsNode(pObjNew) {} else
        if Abc_ObjFanoutNum(pObjNew) == 0 {
            Abc_ObjAddFanin( Abc_NtkCreatePo(pNtkNew), pObjNew ); }
    }
    Abc_NtkAddDummyPiNames( pNtkNew );
    Abc_NtkAddDummyPoNames( pNtkNew );
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Printf("Abc_NtkBottommost: The network check has failed.\n");
        Abc_NtkDelete( pNtkNew );
        return nil;
    }
    return pNtkNew;
}



/**Function*************************************************************

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_CompareNodeIds(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
    if Abc_ObjRegular(*pp1).Id < Abc_ObjRegular(*pp2).Id {
        return -1; }
    if Abc_ObjRegular(*pp1).Id > Abc_ObjRegular(*pp2).Id { //
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Collects the large supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeGetSuper(pNode *Abc_Obj_t) *Vec_Ptr_t {
    var vSuper, vFront *Vec_Ptr_t;
    var pAnd, pFanin *Abc_Obj_t;
    var i int32;
    assert(Abc_ObjIsNode(pNode) && !Abc_ObjIsComplement(pNode), "Abc_ObjIsNode(pNode) && !Abc_ObjIsComplement(pNode)");
    vSuper = Vec_PtrAlloc( 100 ); 
    // explore the frontier
    vFront = Vec_PtrAlloc( 100 );
    Vec_PtrPush( vFront, pNode );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pAnd) = (Abc_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        pFanin = Abc_ObjChild0(pAnd);
        if Abc_ObjIsNode(pFanin) && !Abc_ObjIsComplement(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 {
            Vec_PtrPush( vFront, pFanin );
        } else {
            Vec_PtrPush( vSuper, pFanin ); }

        pFanin = Abc_ObjChild1(pAnd);
        if Abc_ObjIsNode(pFanin) && !Abc_ObjIsComplement(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 {
            Vec_PtrPush( vFront, pFanin );
        } else {
            Vec_PtrPush( vSuper, pFanin ); }
    }
    Vec_PtrFree( vFront );
    // reverse the array of pointers to start with lower IDs
    vFront = Vec_PtrAlloc( Vec_PtrSize(vSuper) );
    // inlining of macro Vec_PtrForEachEntryReverse
    for  i = Vec_PtrSize(vSuper) - 1; (i >= 0) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i--  {
        Vec_PtrPush( vFront, pNode );
    }
    Vec_PtrFree( vSuper );
    vSuper = vFront;
    // uniquify and return the frontier
    Vec_PtrUniqify( vSuper, (int (*)())Vec_CompareNodeIds );
    return vSuper;
}

/**Function*************************************************************

  Synopsis    [Copies the topmost levels of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTopAnd(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
    var vNodes, vOrder *Vec_Ptr_t;
    var pNtkAig *Abc_Ntk_t;
    var pObj, pDriver, pObjPo *Abc_Obj_t;
    var i, nNodes int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // get the first PO
    pObjPo = Abc_NtkPo(pNtk, 0);
    vNodes = Abc_NodeGetSuper( Abc_ObjChild0(pObjPo) );
    assert(Vec_PtrSize(vNodes) >= 2, "Vec_PtrSize(vNodes) >= 2");
    // start the new network (constants and CIs of the old network will point to the their counterparts in the new network)
    Abc_NtkCleanCopy( pNtk );
    pNtkAig = Abc_NtkAlloc( ABC_NTK_STRASH, ABC_FUNC_AIG, 1 );
    pNtkAig.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkAig.pSpec = Extra_UtilStrsav(pNtk.pSpec);
    Abc_AigConst1(pNtk).pCopy = Abc_AigConst1(pNtkAig);
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NtkDupObj( pNtkAig, pObj, 1 );
    }
    // restrash the nodes reachable from the roots
    vOrder = Abc_NtkDfsIterNodes( pNtk, vNodes );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        pObj.pCopy = Abc_AigAnd( (Abc_Aig_t *)pNtkAig.pManFunc, Abc_ObjChild0Copy(pObj), Abc_ObjChild1Copy(pObj) );
    }
    Vec_PtrFree( vOrder );
    // finalize the network
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObjPo = Abc_NtkCreatePo(pNtkAig);
        pDriver = Abc_ObjNotCond(Abc_ObjRegular(pObj).pCopy, Abc_ObjIsComplement(pObj));
        Abc_ObjAddFanin( pObjPo, pDriver );
        Abc_ObjAssignName( pObjPo, Abc_ObjName(pObjPo), nil );
    }
    Vec_PtrFree( vNodes );
    // perform cleanup if requested
    if nNodes = Abc_AigCleanup((Abc_Aig_t *)pNtkAig.pManFunc) { 
        fmt.Printf("Abc_NtkTopAnd(): AIG cleanup removed %d nodes (this is a bug).\n", nNodes); }
    // make sure everything is okay
    if !Abc_NtkCheck( pNtkAig ) {
        fmt.Printf("Abc_NtkStrash: The network check has failed.\n");
        Abc_NtkDelete( pNtkAig );
        return nil;
    }
    return pNtkAig;
}

/**Function*************************************************************

  Synopsis    [Writes the AIG into a file for parsing.]

  Description [Ordering: c0, pis, ands, pos. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkWriteAig(pNtk *Abc_Ntk_t, pFileName string) {
    var pFile *os.File;
    var vId2Num *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, iLit int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk) == 0");
    if pFileName == nil {
        pFile = os.Stdout;
    } else {
        pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666); }
    if pFile == nil {
        fmt.Printf("Cannot open output file.\n");
        return;
    }
    vId2Num = Vec_IntAlloc( 2*Abc_NtkObjNumMax(pNtk) );
    Vec_IntFill( vId2Num, 2*Abc_NtkObjNumMax(pNtk), -1 );

    iLit = 0;
    Vec_IntWriteEntry( vId2Num, 2*Abc_ObjId(Abc_AigConst1(pNtk))+1, iLit++ );
    Vec_IntWriteEntry( vId2Num, 2*Abc_ObjId(Abc_AigConst1(pNtk))+0, iLit++ );
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Vec_IntWriteEntry( vId2Num, 2*Abc_ObjId(pObj)+0, iLit++ );
        Vec_IntWriteEntry( vId2Num, 2*Abc_ObjId(pObj)+1, iLit++ );
    }
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Vec_IntWriteEntry( vId2Num, 2*Abc_ObjId(pObj)+0, iLit++ );
        Vec_IntWriteEntry( vId2Num, 2*Abc_ObjId(pObj)+1, iLit++ );
    }
    }
    fmt.Fprintf(pFile, "{\n");
    fmt.Fprintf(pFile, "    \"%s\", ", Abc_NtkName(pNtk));
    fmt.Fprintf(pFile, "//  pi=%d  po=%d  and=%d", Abc_NtkPiNum(pNtk), Abc_NtkPoNum(pNtk), Abc_NtkNodeNum(pNtk));
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, "    { ");
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "\"%s\",", Abc_ObjName(pObj));
    }
    fmt.Fprintf(pFile, "NULL },\n");
    fmt.Fprintf(pFile, "    { ");
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "\"%s\",", Abc_ObjName(pObj));
    }
    fmt.Fprintf(pFile, "NULL },\n");
    fmt.Fprintf(pFile, "    { ");
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        fmt.Fprintf(pFile, "%d,", Vec_IntEntry(vId2Num, 2*Abc_ObjFaninId0(pObj) + Abc_ObjFaninC0(pObj))); }
    }
    fmt.Fprintf(pFile, "0 },\n");
    fmt.Fprintf(pFile, "    { ");
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        fmt.Fprintf(pFile, "%d,", Vec_IntEntry(vId2Num, 2*Abc_ObjFaninId1(pObj) + Abc_ObjFaninC1(pObj))); }
    }
    fmt.Fprintf(pFile, "0 },\n");
    fmt.Fprintf(pFile, "    { ");
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pObj) = Abc_NtkPo(pNtk, i)), 1); i++  {
        fmt.Fprintf(pFile, "%d,", Vec_IntEntry(vId2Num, 2*Abc_ObjFaninId0(pObj) + Abc_ObjFaninC0(pObj)));
    }
    fmt.Fprintf(pFile, "0 },\n");
    fmt.Fprintf(pFile, "},\n");
    if pFile != os.Stdout {
        _ = (pFile).Close(); }
    Vec_IntFree( vId2Num );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPutOnTop(pNtk *Abc_Ntk_t, pNtk2 *Abc_Ntk_t) *Abc_Ntk_t {
    var vNodes *Vec_Ptr_t;
    var pNtkNew *Abc_Ntk_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    assert(Abc_NtkIsLogic(pNtk2), "Abc_NtkIsLogic(pNtk2)");
    assert(Abc_NtkPoNum(pNtk) == Abc_NtkPiNum(pNtk2), "Abc_NtkPoNum(pNtk) == Abc_NtkPiNum(pNtk2)");
    // clean the node copy fields
    Abc_NtkCleanCopy( pNtk );
    Abc_NtkCleanCopy( pNtk2 );
    // duplicate the name and the spec
    pNtkNew = Abc_NtkAlloc( pNtk.ntkType, pNtk.ntkFunc, 1 );
    pNtkNew.pName = Extra_UtilStrsav(pNtk.pName);
    pNtkNew.pSpec = Extra_UtilStrsav(pNtk.pSpec);
    // clone CIs/CIs/boxes
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pObj) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pObj, 1 );
    }
    // add internal nodes
    vNodes = Abc_NtkDfs( pNtk, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy );
        }
    }
    Vec_PtrFree( vNodes );
    // transfer to the POs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk2)) && (((pObj) = Abc_NtkPi(pNtk2, i)), 1); i++  {
        pObj.pCopy = Abc_ObjChild0Copy( Abc_NtkPo(pNtk, i) );
    }
    // add internal nodes
    vNodes = Abc_NtkDfs( pNtk2, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pObj, 0 );
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Abc_ObjAddFanin( pObj.pCopy, pFanin.pCopy );
        }
    }
    Vec_PtrFree( vNodes );
    // clone CIs/CIs/boxes
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk2)) && (((pObj) = Abc_NtkPo(pNtk2, i)), 1); i++  {
        Abc_NtkDupObj( pNtkNew, pObj, 1 );
        Abc_ObjAddFanin( pObj.pCopy, Abc_ObjChild0Copy(pObj) );
    }
    if !Abc_NtkCheck( pNtkNew ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkPutOnTop(): Network check has failed.\n"); }
    return pNtkNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

