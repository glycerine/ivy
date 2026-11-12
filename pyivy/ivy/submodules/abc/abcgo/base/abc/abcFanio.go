package abc

import "fmt"

/**CFile****************************************************************

  FileName    [abcFanio.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Various procedures to connect fanins/fanouts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcFanio.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

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
func Vec_IntPushMem(pMemMan *Mem_Step_t, p *Vec_Int_t, Entry int32) {
    if p.nSize == p.nCap {
        var pArray *int32;
        var i int32;

        if p.nSize == 0 {
            p.nCap = 1; }
        if pMemMan != nil {
            pArray = (int *)Mem_StepEntryFetch( pMemMan, p.nCap * 8 );
        } else {
            // inlining of macro ABC_ALLOC
            pArray = ((int *) malloc(sizeof(int) * (size_t)(p.nCap * 2))); }
        if p.pArray {
            for  i = 0; i < p.nSize; i++  {
                pArray[i] = p.pArray[i];
            }
            if pMemMan != nil {
                Mem_StepEntryRecycle( pMemMan, (char *)p.pArray, p.nCap * 4 );
            } else {
                // inlining of macro ABC_FREE
                ((p.pArray) ? (free((char *) (p.pArray)), (p.pArray) = 0) : 0); }
        }
        p.nCap *= 2;
        p.pArray = pArray;
    }
    p.pArray[p.nSize++] = Entry;
}

/**Function*************************************************************

  Synopsis    [Creates fanout/fanin relationship between the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjAddFanin(pObj *Abc_Obj_t, pFanin *Abc_Obj_t) {
    var pFaninR *Abc_Obj_t  = Abc_ObjRegular(pFanin);
    assert(!Abc_ObjIsComplement(pObj), "!Abc_ObjIsComplement(pObj)");
    assert(pObj.pNtk == pFaninR.pNtk, "pObj.pNtk == pFaninR.pNtk");
    assert(pObj.Id >= 0 && pFaninR.Id >= 0, "pObj.Id >= 0 && pFaninR.Id >= 0");
    assert(!Abc_ObjIsPi(pObj) && !Abc_ObjIsPo(pFaninR), "!Abc_ObjIsPi(pObj) && !Abc_ObjIsPo(pFaninR)");    // fanin of PI or fanout of PO
    assert(!Abc_ObjIsCo(pObj) || !Abc_ObjFaninNum(pObj), "!Abc_ObjIsCo(pObj) || !Abc_ObjFaninNum(pObj)");  // CO with two fanins
    assert(!Abc_ObjIsNet(pObj) || !Abc_ObjFaninNum(pObj), "!Abc_ObjIsNet(pObj) || !Abc_ObjFaninNum(pObj)"); // net with two fanins
    Vec_IntPushMem( pObj.pNtk.pMmStep, &pObj.vFanins,     pFaninR.Id );
    Vec_IntPushMem( pObj.pNtk.pMmStep, &pFaninR.vFanouts, pObj.Id    );
    if Abc_ObjIsComplement(pFanin) {
        Abc_ObjSetFaninC( pObj, Abc_ObjFaninNum(pObj)-1 ); }
}


/**Function*************************************************************

  Synopsis    [Destroys fanout/fanin relationship between the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjDeleteFanin(pObj *Abc_Obj_t, pFanin *Abc_Obj_t) {
    assert(!Abc_ObjIsComplement(pObj), "!Abc_ObjIsComplement(pObj)");
    assert(!Abc_ObjIsComplement(pFanin), "!Abc_ObjIsComplement(pFanin)");
    assert(pObj.pNtk == pFanin.pNtk, "pObj.pNtk == pFanin.pNtk");
    assert(pObj.Id >= 0 && pFanin.Id >= 0, "pObj.Id >= 0 && pFanin.Id >= 0");
    if !Vec_IntRemove( &pObj.vFanins, pFanin.Id ) {
        fmt.Printf("The obj %d is not found among the fanins of obj %d ...\n", pFanin.Id, pObj.Id);
        return;
    }
    if !Vec_IntRemove( &pFanin.vFanouts, pObj.Id ) {
        fmt.Printf("The obj %d is not found among the fanouts of obj %d ...\n", pObj.Id, pFanin.Id);
        return;
    }
}


/**Function*************************************************************

  Synopsis    [Destroys fanout/fanin relationship between the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjRemoveFanins(pObj *Abc_Obj_t) {
    var vFaninsOld *Vec_Int_t;
    var pFanin *Abc_Obj_t;
    var k int32;
    // remove old fanins
    vFaninsOld = &pObj.vFanins;
    for  k = vFaninsOld.nSize - 1; k >= 0; k--  {
        pFanin = Abc_NtkObj( pObj.pNtk, vFaninsOld.pArray[k] );
        Abc_ObjDeleteFanin( pObj, pFanin );
    }
    pObj.fCompl0 = 0;
    pObj.fCompl1 = 0;
    assert(vFaninsOld.nSize == 0, "vFaninsOld.nSize == 0");
}

/**Function*************************************************************

  Synopsis    [Replaces a fanin of the node.]

  Description [The node is pObj. An old fanin of this node (pFaninOld) has to be
  replaced by a new fanin (pFaninNew). Assumes that the node and the old fanin 
  are not complemented. The new fanin can be complemented. In this case, the
  polarity of the new fanin will change, compared to the polarity of the old fanin.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjPatchFanin(pObj *Abc_Obj_t, pFaninOld *Abc_Obj_t, pFaninNew *Abc_Obj_t) {
    var pFaninNewR *Abc_Obj_t  = Abc_ObjRegular(pFaninNew);
    var iFanin int32;//, nLats;//, fCompl;
    assert(!Abc_ObjIsComplement(pObj), "!Abc_ObjIsComplement(pObj)");
    assert(!Abc_ObjIsComplement(pFaninOld), "!Abc_ObjIsComplement(pFaninOld)");
    assert(pFaninOld != pFaninNewR, "pFaninOld != pFaninNewR");
//    assert( pObj != pFaninOld );
//    assert( pObj != pFaninNewR );
    assert(pObj.pNtk == pFaninOld.pNtk, "pObj.pNtk == pFaninOld.pNtk");
    assert(pObj.pNtk == pFaninNewR.pNtk, "pObj.pNtk == pFaninNewR.pNtk");
    if (iFanin = Vec_IntFind( &pObj.vFanins, pFaninOld.Id )) == -1 {
        fmt.Printf("Node %s is not among", Abc_ObjName(pFaninOld));
        fmt.Printf(" the fanins of node %s...\n", Abc_ObjName(pObj));
        return;
    }

    // remember the attributes of the old fanin
//    fCompl = Abc_ObjFaninC(pObj, iFanin);
    // replace the old fanin entry by the new fanin entry (removes attributes)
    Vec_IntWriteEntry( &pObj.vFanins, iFanin, pFaninNewR.Id );
    // set the attributes of the new fanin
//    if ( fCompl ^ Abc_ObjIsComplement(pFaninNew) )
//        Abc_ObjSetFaninC( pObj, iFanin );
    if Abc_ObjIsComplement(pFaninNew) {
        Abc_ObjXorFaninC( pObj, iFanin ); }

//    if ( Abc_NtkIsSeq(pObj->pNtk) && (nLats = Seq_ObjFaninL(pObj, iFanin)) )
//        Seq_ObjSetFaninL( pObj, iFanin, nLats );
    // update the fanout of the fanin
    if !Vec_IntRemove( &pFaninOld.vFanouts, pObj.Id ) {
        fmt.Printf("Node %s is not among", Abc_ObjName(pObj));
        fmt.Printf(" the fanouts of its old fanin %s...\n", Abc_ObjName(pFaninOld));
//        return;
    }
    Vec_IntPushMem( pObj.pNtk.pMmStep, &pFaninNewR.vFanouts, pObj.Id );
}

/**Function*************************************************************

  Synopsis    [Replaces pObj by iObjNew in the fanin arrays of the fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjPatchFanoutFanin(pObj *Abc_Obj_t, iObjNew int32) {
    var pFanout *Abc_Obj_t;
    var i, k, Entry int32;
    // update fanouts of the node to point to this one
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(&pFanout.vFanins)) && (((Entry) = Vec_IntEntry(&pFanout.vFanins, k)), 1); k++  {
            if Entry == int32(Abc_ObjId(pObj)) {
                Vec_IntWriteEntry( &pFanout.vFanins, k, iObjNew );
                break;
            }
        }
        assert(k < Vec_IntSize(&pFanout.vFanins), "k < Vec_IntSize(&pFanout.vFanins)");
    }
}

/**Function*************************************************************

  Synopsis    [Inserts one-input node of the type specified between the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjInsertBetween(pNodeIn *Abc_Obj_t, pNodeOut *Abc_Obj_t, Type Abc_ObjType_t) *Abc_Obj_t {
    var pNodeNew *Abc_Obj_t;
    var iFanoutIndex, iFaninIndex int32;
    // find pNodeOut among the fanouts of pNodeIn
    if (iFanoutIndex = Vec_IntFind( &pNodeIn.vFanouts, pNodeOut.Id )) == -1 {
        fmt.Printf("Node %s is not among", Abc_ObjName(pNodeOut));
        fmt.Printf(" the fanouts of node %s...\n", Abc_ObjName(pNodeIn));
        return nil;
    }
    // find pNodeIn among the fanins of pNodeOut
    if (iFaninIndex = Vec_IntFind( &pNodeOut.vFanins, pNodeIn.Id )) == -1 {
        fmt.Printf("Node %s is not among", Abc_ObjName(pNodeIn));
        fmt.Printf(" the fanins of node %s...\n", Abc_ObjName(pNodeOut));
        return nil;
    }
    // create the new node
    pNodeNew = Abc_NtkCreateObj( pNodeIn.pNtk, Type );
    // add pNodeIn as fanin and pNodeOut as fanout
    Vec_IntPushMem( pNodeNew.pNtk.pMmStep, &pNodeNew.vFanins,  pNodeIn.Id  );
    Vec_IntPushMem( pNodeNew.pNtk.pMmStep, &pNodeNew.vFanouts, pNodeOut.Id );
    // update the fanout of pNodeIn
    Vec_IntWriteEntry( &pNodeIn.vFanouts, iFanoutIndex, pNodeNew.Id );
    // update the fanin of pNodeOut
    Vec_IntWriteEntry( &pNodeOut.vFanins, iFaninIndex, pNodeNew.Id );
    return pNodeNew;
}

/**Function*************************************************************

  Synopsis    [Transfers fanout from the old node to the new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjTransferFanout(pNodeFrom *Abc_Obj_t, pNodeTo *Abc_Obj_t) {
    var vFanouts *Vec_Ptr_t;
    var nFanoutsOld, i int32;
    assert(!Abc_ObjIsComplement(pNodeFrom), "!Abc_ObjIsComplement(pNodeFrom)");
    assert(!Abc_ObjIsComplement(pNodeTo), "!Abc_ObjIsComplement(pNodeTo)");
    assert(!Abc_ObjIsPo(pNodeFrom) && !Abc_ObjIsPo(pNodeTo), "!Abc_ObjIsPo(pNodeFrom) && !Abc_ObjIsPo(pNodeTo)");
    assert(pNodeFrom.pNtk == pNodeTo.pNtk, "pNodeFrom.pNtk == pNodeTo.pNtk");
    assert(pNodeFrom != pNodeTo, "pNodeFrom != pNodeTo");
    assert(!Abc_ObjIsNode(pNodeFrom) || Abc_ObjFanoutNum(pNodeFrom) > 0, "!Abc_ObjIsNode(pNodeFrom) || Abc_ObjFanoutNum(pNodeFrom) > 0");
    // get the fanouts of the old node
    nFanoutsOld = Abc_ObjFanoutNum(pNodeTo);
    vFanouts = Vec_PtrAlloc( nFanoutsOld );
    Abc_NodeCollectFanouts( pNodeFrom, vFanouts );
    // patch the fanin of each of them
    for  i = 0; i < vFanouts.nSize; i++  {
        Abc_ObjPatchFanin( (Abc_Obj_t *)vFanouts.pArray[i], pNodeFrom, pNodeTo );
    }
    assert(Abc_ObjFanoutNum(pNodeFrom) == 0, "Abc_ObjFanoutNum(pNodeFrom) == 0");
    assert(Abc_ObjFanoutNum(pNodeTo) == nFanoutsOld + vFanouts.nSize, "Abc_ObjFanoutNum(pNodeTo) == nFanoutsOld + vFanouts.nSize");
    Vec_PtrFree( vFanouts );
}

/**Function*************************************************************

  Synopsis    [Replaces the node by a new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjReplace(pNodeOld *Abc_Obj_t, pNodeNew *Abc_Obj_t) {
    assert(!Abc_ObjIsComplement(pNodeOld), "!Abc_ObjIsComplement(pNodeOld)");
    assert(!Abc_ObjIsComplement(pNodeNew), "!Abc_ObjIsComplement(pNodeNew)");
    assert(pNodeOld.pNtk == pNodeNew.pNtk, "pNodeOld.pNtk == pNodeNew.pNtk");
    assert(pNodeOld != pNodeNew, "pNodeOld != pNodeNew");
    assert(Abc_ObjFanoutNum(pNodeOld) > 0, "Abc_ObjFanoutNum(pNodeOld) > 0");
    // transfer the fanouts to the old node
    Abc_ObjTransferFanout( pNodeOld, pNodeNew );
    // remove the old node
    Abc_NtkDeleteObj_rec( pNodeOld, 1 );
}

/**Function*************************************************************

  Synopsis    [Replaces a node by a constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjReplaceByConstant(pNode *Abc_Obj_t, fConst1 int32) {
    var pNodeNew *Abc_Obj_t;
    assert(Abc_NtkIsLogic(pNode.pNtk), "Abc_NtkIsLogic(pNode.pNtk)");
    assert(!Abc_ObjIsCo(pNode), "!Abc_ObjIsCo(pNode)");
    if fConst1 != 0 {
        pNodeNew = Abc_NtkCreateNodeConst1(pNode.pNtk);
    } else {
        pNodeNew = Abc_NtkCreateNodeConst0(pNode.pNtk);
    }
    // transfer the fanouts to the old node
    Abc_ObjTransferFanout( pNode, pNodeNew );
    // remove the old node
    if Abc_ObjIsNode(pNode) {
        Abc_NtkDeleteObj_rec( pNode, 1 ); }
}

/**Function*************************************************************

  Synopsis    [Returns the index of the fanin in the fanin list of the fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjFanoutFaninNum(pFanout *Abc_Obj_t, pFanin *Abc_Obj_t) int32 {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pFanout)) && (((pObj) = Abc_ObjFanin(pFanout, i)), 1); i++  {
        if pObj == pFanin {
            return i; }
    }
    return -1;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

