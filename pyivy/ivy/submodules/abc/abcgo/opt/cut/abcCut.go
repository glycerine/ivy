package cut

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [abcCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Interface to cut computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcCut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "cut.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Abc_NtkPrintCuts( void * p, Abc_Ntk_t * pNtk, int fSeq );

// static void Abc_NtkPrintCuts_( void * p, Abc_Ntk_t * pNtk, int fSeq );



var nTotal, nGood, nEqual int32;

// temporary
//Vec_Int_t * Abc_NtkGetNodeAttributes( Abc_Ntk_t * pNtk ) { return NULL; }
func Abc_NtkGetNodeAttributes(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vAttrs *Vec_Int_t  = Vec_IntStart( Abc_NtkObjNumMax(pNtk) + 1 );
    var i int32;
    var pObj *Abc_Obj_t;

//    Abc_NtkForEachCi( pNtk, pObj, i )
//        Vec_IntWriteEntry( vAttrs, pObj->Id, 1 );

    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
//        if ( Abc_ObjIsNode(pObj) && (rand() % 4 == 0) )
        if Abc_ObjIsNode(pObj) && Abc_ObjFanoutNum(pObj) > 1 && !Abc_NodeIsMuxControlType(pObj) && (rand() % 3 == 0) {
            Vec_IntWriteEntry( vAttrs, pObj.Id, 1 ); }
    }
    }
    return vAttrs; 
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCuts(pNtk *Abc_Ntk_t, pParams *Cut_Params_t) *Cut_Man_t {
    var pProgress *ProgressBar;
    var p *Cut_Man_t;
    var pObj, pNode *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var vChoices *Vec_Int_t;
    var i int32;
    var clk int64  = time.Now().UnixNano();

    extern void Abc_NtkBalanceAttach( Abc_Ntk_t * pNtk );
    extern void Abc_NtkBalanceDetach( Abc_Ntk_t * pNtk );

    nTotal = nGood = nEqual = 0;

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    // start the manager
    pParams.nIdsMax = Abc_NtkObjNumMax( pNtk );
    p = Cut_ManStart( pParams );
    // compute node attributes if local or global cuts are requested
    if pParams.fGlobal || pParams.fLocal {
        extern Vec_Int_t * Abc_NtkGetNodeAttributes( Abc_Ntk_t * pNtk );
        Cut_ManSetNodeAttrs( p, Abc_NtkGetNodeAttributes(pNtk) );
    }
    // prepare for cut dropping
    if pParams.fDrop {
        Cut_ManSetFanoutCounts( p, Abc_NtkFanoutCounts(pNtk) ); }
    // set cuts for PIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjFanoutNum(pObj) > 0 {
            Cut_NodeSetTriv( p, pObj.Id ); }
    }
    // compute cuts for internal nodes
    vNodes = Abc_AigDfs( pNtk, 0, 1 ); // collects POs
    vChoices = Vec_IntAlloc( 100 );
    pProgress = Extra_ProgressBarStart( os.Stdout, Vec_PtrSize(vNodes) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // when we reached a CO, it is time to deallocate the cuts
        if Abc_ObjIsCo(pObj) {
            if pParams.fDrop {
                Cut_NodeTryDroppingCuts( p, Abc_ObjFaninId0(pObj) ); }
            continue;
        }
        // skip constant node, it has no cuts
//        if ( Abc_NodeIsConst(pObj) )
//            continue;
        Extra_ProgressBarUpdate( pProgress, i, nil );
        // compute the cuts to the internal node
        Abc_NodeGetCuts( p, pObj, pParams.fDag, pParams.fTree );  
        // consider dropping the fanins cuts
        if pParams.fDrop {
            Cut_NodeTryDroppingCuts( p, Abc_ObjFaninId0(pObj) );
            Cut_NodeTryDroppingCuts( p, Abc_ObjFaninId1(pObj) );
        }
        // add cuts due to choices
        if Abc_AigNodeIsChoice(pObj) {
            Vec_IntClear( vChoices );
            for  pNode = pObj; (pNode != nil); pNode = pNode.pData  {
                Vec_IntPush( vChoices, pNode.Id );
            }
            Cut_NodeUnionCuts( p, vChoices );
        }
    }
    Extra_ProgressBarStop( pProgress );
    Vec_PtrFree( vNodes );
    Vec_IntFree( vChoices );
PRT( "Total", time.Now().UnixNano() - clk );
//Abc_NtkPrintCuts( p, pNtk, 0 );
//    Cut_ManPrintStatsToFile( p, pNtk->pSpec, clock() - clk );

    // temporary printout of stats
    if nTotal {
    fmt.Printf("Total cuts = %d. Good cuts = %d.  Ratio = %5.2f\n", nTotal, nGood, (float64(nGood))/nTotal); }
    return p;
}

/**Function*************************************************************

  Synopsis    [Cut computation using the oracle.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCutsOracle(pNtk *Abc_Ntk_t, p *Cut_Oracle_t) {
    var pObj *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i int32;
    var clk int64  = time.Now().UnixNano();
    var fDrop int32  = Cut_OracleReadDrop(p);

    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");

    // prepare cut droppping
    if fDrop != 0 {
        Cut_OracleSetFanoutCounts( p, Abc_NtkFanoutCounts(pNtk) ); }

    // set cuts for PIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjFanoutNum(pObj) > 0 {
            Cut_OracleNodeSetTriv( p, pObj.Id ); }
    }

    // compute cuts for internal nodes
    vNodes = Abc_AigDfs( pNtk, 0, 1 ); // collects POs
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // when we reached a CO, it is time to deallocate the cuts
        if Abc_ObjIsCo(pObj) {
            if fDrop != 0 {
                Cut_OracleTryDroppingCuts( p, Abc_ObjFaninId0(pObj) ); }
            continue;
        }
        // skip constant node, it has no cuts
//        if ( Abc_NodeIsConst(pObj) )
//            continue;
        // compute the cuts to the internal node
        Cut_OracleComputeCuts( p, pObj.Id, Abc_ObjFaninId0(pObj), Abc_ObjFaninId1(pObj),  
                Abc_ObjFaninC0(pObj), Abc_ObjFaninC1(pObj) );
        // consider dropping the fanins cuts
        if fDrop != 0 {
            Cut_OracleTryDroppingCuts( p, Abc_ObjFaninId0(pObj) );
            Cut_OracleTryDroppingCuts( p, Abc_ObjFaninId1(pObj) );
        }
    }
    Vec_PtrFree( vNodes );
//PRT( "Total", clock() - clk );
//Abc_NtkPrintCuts_( p, pNtk, 0 );
}


/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSeqCuts(pNtk *Abc_Ntk_t, pParams *Cut_Params_t) *Cut_Man_t {
/*
    Cut_Man_t *  p;
    Abc_Obj_t * pObj, * pNode;
    int i, nIters, fStatus;
    Vec_Int_t * vChoices;
    clock_t clk = clock();

    assert( Abc_NtkIsSeq(pNtk) );
    assert( pParams->fSeq );
//    assert( Abc_NtkIsDfsOrdered(pNtk) );

    // start the manager
    pParams->nIdsMax = Abc_NtkObjNumMax( pNtk );
    pParams->nCutSet = Abc_NtkCutSetNodeNum( pNtk );
    p = Cut_ManStart( pParams );

    // set cuts for the constant node and the PIs
    pObj = Abc_AigConst1(pNtk);
    if ( Abc_ObjFanoutNum(pObj) > 0 )
        Cut_NodeSetTriv( p, pObj->Id );
    Abc_NtkForEachPi( pNtk, pObj, i )
    {
//printf( "Setting trivial cut %d.\n", pObj->Id );
        Cut_NodeSetTriv( p, pObj->Id );
    }
    // label the cutset nodes and set their number in the array
    // assign the elementary cuts to the cutset nodes
    Abc_SeqForEachCutsetNode( pNtk, pObj, i )
    {
        assert( pObj->fMarkC == 0 );
        pObj->fMarkC = 1;
        pObj->pCopy = (Abc_Obj_t *)i;
        Cut_NodeSetTriv( p, pObj->Id );
//printf( "Setting trivial cut %d.\n", pObj->Id );
    }

    // process the nodes
    vChoices = Vec_IntAlloc( 100 );
    for ( nIters = 0; nIters < 10; nIters++ )
    {
//printf( "ITERATION %d:\n", nIters );
        // compute the cuts for the internal nodes
        Abc_AigForEachAnd( pNtk, pObj, i )
        {
            Abc_NodeGetCutsSeq( p, pObj, nIters==0 );
            // add cuts due to choices
            if ( Abc_AigNodeIsChoice(pObj) )
            {
                Vec_IntClear( vChoices );
                for ( pNode = pObj; pNode; pNode = pNode->pData )
                    Vec_IntPush( vChoices, pNode->Id );
                Cut_NodeUnionCutsSeq( p, vChoices, (pObj->fMarkC ? (int)pObj->pCopy : -1), nIters==0 );
            }
        }
        // merge the new cuts with the old cuts
        Abc_NtkForEachPi( pNtk, pObj, i )
            Cut_NodeNewMergeWithOld( p, pObj->Id );
        Abc_AigForEachAnd( pNtk, pObj, i )
            Cut_NodeNewMergeWithOld( p, pObj->Id );
        // for the cutset, transfer temp cuts to new cuts
        fStatus = 0;
        Abc_SeqForEachCutsetNode( pNtk, pObj, i )
            fStatus |= Cut_NodeTempTransferToNew( p, pObj->Id, i );
        if ( fStatus == 0 )
            break;
    }
    Vec_IntFree( vChoices );

    // if the status is not finished, transfer new to old for the cutset
    Abc_SeqForEachCutsetNode( pNtk, pObj, i )
        Cut_NodeNewMergeWithOld( p, pObj->Id );

    // transfer the old cuts to the new positions
    Abc_NtkForEachObj( pNtk, pObj, i )
        Cut_NodeOldTransferToNew( p, pObj->Id );

    // unlabel the cutset nodes
    Abc_SeqForEachCutsetNode( pNtk, pObj, i )
        pObj->fMarkC = 0;
if ( pParams->fVerbose )
{
PRT( "Total", clock() - clk );
printf( "Converged after %d iterations.\n", nIters );
}
//Abc_NtkPrintCuts( p, pNtk, 1 );
    return p;
*/
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeGetCutsRecursive(p any, pObj *Abc_Obj_t, fDag int32, fTree int32) any {
    var pList any;
    if pList = Abc_NodeReadCuts( p, pObj ) {
        return pList; }
    Abc_NodeGetCutsRecursive( p, Abc_ObjFanin0(pObj), fDag, fTree );
    Abc_NodeGetCutsRecursive( p, Abc_ObjFanin1(pObj), fDag, fTree );
    return Abc_NodeGetCuts( p, pObj, fDag, fTree );
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeGetCuts(p any, pObj *Abc_Obj_t, fDag int32, fTree int32) any {
    var pFanin *Abc_Obj_t;
    var fDagNode, fTriv, TreeCode int32;
//    assert( Abc_NtkIsStrash(pObj->pNtk) );
    assert(Abc_ObjFaninNum(pObj) == 2, "Abc_ObjFaninNum(pObj) == 2");


    // check if the node is a DAG node
    fDagNode = (Abc_ObjFanoutNum(pObj) > 1 && !Abc_NodeIsMuxControlType(pObj));
    // increment the counter of DAG nodes
    if fDagNode { Cut_ManIncrementDagNodes( p ); }
    // add the trivial cut if the node is a DAG node, or if we compute all cuts
    fTriv = fDagNode || !fDag;
    // check if fanins are DAG nodes
    if fTree != 0 {
        pFanin = Abc_ObjFanin0(pObj);
        TreeCode |=  (Abc_ObjFanoutNum(pFanin) > 1 && !Abc_NodeIsMuxControlType(pFanin));
        pFanin = Abc_ObjFanin1(pObj);
        TreeCode |= ((Abc_ObjFanoutNum(pFanin) > 1 && !Abc_NodeIsMuxControlType(pFanin)) << 1);
    }


    // changes due to the global/local cut computation
    {
        var pParams *Cut_Params_t  = Cut_ManReadParams(p);
        if pParams.fLocal {
            var vNodeAttrs *Vec_Int_t  = Cut_ManReadNodeAttrs(p);
            fDagNode = Vec_IntEntry( vNodeAttrs, pObj.Id );
            if fDagNode { Cut_ManIncrementDagNodes( p ); }
//            fTriv = fDagNode || !pParams->fGlobal;
            fTriv = !Vec_IntEntry( vNodeAttrs, pObj.Id );
            TreeCode = 0;
            pFanin = Abc_ObjFanin0(pObj);
            TreeCode |=  Vec_IntEntry( vNodeAttrs, pFanin.Id );
            pFanin = Abc_ObjFanin1(pObj);
            TreeCode |= (Vec_IntEntry( vNodeAttrs, pFanin.Id ) << 1);
        }
    }
    return Cut_NodeComputeCuts( p, pObj.Id, Abc_ObjFaninId0(pObj), Abc_ObjFaninId1(pObj),  
        Abc_ObjFaninC0(pObj), Abc_ObjFaninC1(pObj), fTriv, TreeCode );  
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeGetCutsSeq(p any, pObj *Abc_Obj_t, fTriv int32) {
    var CutSetNum int32;
    assert(Abc_NtkIsSeq(pObj.pNtk), "Abc_NtkIsSeq(pObj.pNtk)");
    assert(Abc_ObjFaninNum(pObj) == 2, "Abc_ObjFaninNum(pObj) == 2");
    if pObj.fMarkC {
        fTriv = 0;
    } else {
        fTriv = fTriv;
    }
    if pObj.fMarkC {
        CutSetNum = int32(pObj.pCopy);
    } else {
        CutSetNum = -1;
    }
    Cut_NodeComputeCutsSeq( p, pObj.Id, Abc_ObjFaninId0(pObj), Abc_ObjFaninId1(pObj),  
        Abc_ObjFaninC0(pObj), Abc_ObjFaninC1(pObj), Seq_ObjFaninL0(pObj), Seq_ObjFaninL1(pObj), fTriv, CutSetNum );  
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeReadCuts(p any, pObj *Abc_Obj_t) any {
    return Cut_NodeReadCutsNew( p, pObj.Id );  
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeFreeCuts(p any, pObj *Abc_Obj_t) {
    Cut_NodeFreeCuts( p, pObj.Id );
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintCuts(p any, pNtk *Abc_Ntk_t, fSeq int32) {
    var pMan *Cut_Man_t  = p;
    var pList *Cut_Cut_t;
    var pObj *Abc_Obj_t;
    var i int32;
    fmt.Printf("Cuts of the network:\n");
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        pList = Abc_NodeReadCuts( p, pObj );
        fmt.Printf("Node %s:\n", Abc_ObjName(pObj));
        Cut_CutPrintList( pList, fSeq );
    }
    }
}

/**Function*************************************************************

  Synopsis    [Computes the cuts for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrintCuts_(p any, pNtk *Abc_Ntk_t, fSeq int32) {
    var pMan *Cut_Man_t  = p;
    var pList *Cut_Cut_t;
    var pObj *Abc_Obj_t;
    pObj = Abc_NtkObj( pNtk, 2 * Abc_NtkObjNum(pNtk) / 3 );
    pList = Abc_NodeReadCuts( p, pObj );
    fmt.Printf("Node %s:\n", Abc_ObjName(pObj));
    Cut_CutPrintList( pList, fSeq );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

