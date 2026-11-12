package abci

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [abcMffc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Computing multi-output maximum fanout-free cones.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMffc.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"

//ABC_NAMESPACE_IMPL_START
 
////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Dereferences and collects the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MffcDeref_rec(pNode *Abc_Obj_t, vNodes *Vec_Ptr_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    if Abc_ObjIsCi(pNode) {
        return; }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        assert(pFanin.vFanouts.nSize > 0, "pFanin.vFanouts.nSize > 0");
        if --pFanin.vFanouts.nSize == 0 {
            Abc_MffcDeref_rec( pFanin, vNodes ); }
    }
    if vNodes {
        Vec_PtrPush( vNodes, pNode ); }
}

/**Function*************************************************************

  Synopsis    [References the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MffcRef_rec(pNode *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    if Abc_ObjIsCi(pNode) {
        return; }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if pFanin.vFanouts.nSize++ == 0 {
            Abc_MffcRef_rec( pFanin ); }
    }
}

/**Function*************************************************************

  Synopsis    [Collects nodes belonging to the MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MffcCollectNodes(pNodes **Abc_Obj_t, nNodes int32, vNodes *Vec_Ptr_t) {
    var i int32;
    Vec_PtrClear( vNodes );
    for  i = 0; i < nNodes; i++  {
        Abc_MffcDeref_rec( pNodes[i], vNodes );
    }
    for  i = 0; i < nNodes; i++  {
        Abc_MffcRef_rec( pNodes[i] );
    }
}

/**Function*************************************************************

  Synopsis    [Collects leaves of the MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_MffcCollectLeaves(vNodes *Vec_Ptr_t, vLeaves *Vec_Ptr_t) {
    var pNode, pFanin *Abc_Obj_t;
    var i, k int32;
    assert(Vec_PtrSize(vNodes) > 0, "Vec_PtrSize(vNodes) > 0");
    pNode = (Abc_Obj_t *)Vec_PtrEntry( vNodes, 0 );
    // label them
    Abc_NtkIncrementTravId( pNode.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Abc_NodeSetTravIdCurrent( pNode );
    }
    // collect non-labeled fanins
    Vec_PtrClear( vLeaves );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
    // inlining of macro Abc_ObjForEachFanin
    for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
        if Abc_NodeIsTravIdCurrent(pFanin) {
            continue; }
        Abc_NodeSetTravIdCurrent( pFanin );
        Vec_PtrPush( vLeaves, pFanin );
    }
    }
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes that are roots of MFFCs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcMarkRoots(pNtk *Abc_Ntk_t, fSkipPis int32) *Vec_Ptr_t {
    var vRoots, vNodes, vLeaves *Vec_Ptr_t;
    var pObj, pLeaf *Abc_Obj_t;
    var i, k int32;
    Abc_NtkCleanMarkA( pNtk );
    // mark the drivers of combinational outputs
    vRoots  = Vec_PtrAlloc( 1000 );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pObj = Abc_ObjFanin0( pObj );
//        if ( Abc_ObjIsCi(pObj) || Abc_ObjFaninNum(pObj) == 0 || pObj->fMarkA )
        if Abc_ObjIsCi(pObj) || pObj.fMarkA {
            continue; }
        pObj.fMarkA = 1;
        Vec_PtrPush( vRoots, pObj );
    }
    // explore starting from the drivers
    vNodes  = Vec_PtrAlloc( 100 );
    vLeaves = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        if Abc_ObjIsCi(pObj) {
            continue; }
        // collect internal nodes 
        Abc_MffcCollectNodes( &pObj, 1, vNodes );
        // collect leaves
        Abc_MffcCollectLeaves( vNodes, vLeaves );
        // add non-PI leaves
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pLeaf) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            if ((fSkipPis != 0) && Abc_ObjIsCi(pLeaf)) || pLeaf.fMarkA {
                continue; }
            pLeaf.fMarkA = 1;
            Vec_PtrPush( vRoots, pLeaf );
        }
    }
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vNodes );
    Abc_NtkCleanMarkA( pNtk );
    return vRoots;
}

/**Function*************************************************************

  Synopsis    [Collect fanout reachable root nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcCollectFanout_rec(pObj *Abc_Obj_t, vFanouts *Vec_Ptr_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    if Abc_ObjIsCo(pObj) {
        return; }
    if Abc_ObjFanoutNum(pObj) > 64 {
        return; }
    if Abc_NodeIsTravIdCurrent(pObj) {
        return; }
    Abc_NodeSetTravIdCurrent(pObj);
    if pObj.fMarkA {
        if pObj.vFanouts.nSize > 0 {
            Vec_PtrPush( vFanouts, pObj ); }
        return;
    }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
        Abc_NktMffcCollectFanout_rec( pFanout, vFanouts );
    }
}

/**Function*************************************************************

  Synopsis    [Collect fanout reachable root nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcCollectFanout(pNodes **Abc_Obj_t, nNodes int32, vLeaves *Vec_Ptr_t, vFanouts *Vec_Ptr_t) {
    var pFanin, pFanout *Abc_Obj_t;
    var i, k int32;
    // dereference nodes
    for  i = 0; i < nNodes; i++  {
        Abc_MffcDeref_rec( pNodes[i], nil );
    }
    // collect fanouts
    Vec_PtrClear( vFanouts );
    pFanin = (Abc_Obj_t *)Vec_PtrEntry( vLeaves, 0 );
    Abc_NtkIncrementTravId( pFanin.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pFanin)) && (((pFanout) = Abc_ObjFanout(pFanin, k)), 1); k++  {
            Abc_NktMffcCollectFanout_rec( pFanout, vFanouts );
        }
    }
    // reference nodes
    for  i = 0; i < nNodes; i++  {
        Abc_MffcRef_rec( pNodes[i] );
    }
}

/**Function*************************************************************

  Synopsis    [Grow one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcGrowOne(pNtk *Abc_Ntk_t, ppObjs **Abc_Obj_t, nObjs int32, vNodes *Vec_Ptr_t, vLeaves *Vec_Ptr_t, vFanouts *Vec_Ptr_t) *Abc_Obj_t {
    var pFanout *Abc_Obj_t;
    var pFanoutBest *Abc_Obj_t = nil;
    var CostBest float64  = 0.0;
    var i, k int32;
    Abc_MffcCollectNodes( ppObjs, nObjs, vNodes );
    Abc_MffcCollectLeaves( vNodes, vLeaves );
    // collect fanouts of all fanins
    Abc_NktMffcCollectFanout( ppObjs, nObjs, vLeaves, vFanouts );
    // try different fanouts
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanouts)) && (((pFanout) = (Abc_Obj_t *)Vec_PtrEntry(vFanouts, i)), 1); i++  {
        for  k = 0; k < nObjs; k++  {
            if pFanout == ppObjs[k] {
                break; }
        }
        if k < nObjs {
            continue; }
        ppObjs[nObjs] = pFanout;
        Abc_MffcCollectNodes( ppObjs, nObjs+1, vNodes );
        Abc_MffcCollectLeaves( vNodes, vLeaves );
        if pFanoutBest == nil || CostBest < 1.0 * Vec_PtrSize(vNodes)/Vec_PtrSize(vLeaves) {
            CostBest = 1.0 * Vec_PtrSize(vNodes)/Vec_PtrSize(vLeaves);
            pFanoutBest = pFanout;
        }
    }
    return pFanoutBest;
}

/**Function*************************************************************

  Synopsis    [Procedure to increase MFF size by pairing nodes.]

  Description [For each node in the array vRoots, find a matching node, 
  so that the ratio of nodes inside to the leaf nodes is maximized.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcGrowRoots(pNtk *Abc_Ntk_t, vRoots *Vec_Ptr_t) *Vec_Ptr_t {
    var vRoots1, vNodes, vLeaves, vFanouts *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var pRoot2 *Abc_Obj_t;
    var pNodes [2]*Abc_Obj_t;
    var i int32;
    Abc_NtkCleanMarkA( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        pObj.fMarkA = 1;
    }
    vRoots1  = Vec_PtrAlloc( 100 );
    vNodes   = Vec_PtrAlloc( 100 );
    vLeaves  = Vec_PtrAlloc( 100 );
    vFanouts = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        pNodes[0] = pObj;
        pRoot2 = Abc_NktMffcGrowOne( pNtk, pNodes, 1, vNodes, vLeaves, vFanouts );
        Vec_PtrPush( vRoots1, pRoot2 );
    }
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vFanouts );
    Abc_NtkCleanMarkA( pNtk );
    return vRoots1;
}

/**Function*************************************************************

  Synopsis    [Procedure to increase MFF size by pairing nodes.]

  Description [For each node in the array vRoots, find a matching node, 
  so that the ratio of nodes inside to the leaf nodes is maximized.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcGrowRootsAgain(pNtk *Abc_Ntk_t, vRoots *Vec_Ptr_t, vRoots1 *Vec_Ptr_t) *Vec_Ptr_t {
    var vRoots2, vNodes, vLeaves, vFanouts *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var pRoot2 *Abc_Obj_t;
    var ppObjs [3]*Abc_Obj_t;
    var i int32;
    Abc_NtkCleanMarkA( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        pObj.fMarkA = 1;
    }
    vRoots2  = Vec_PtrAlloc( 100 );
    vNodes   = Vec_PtrAlloc( 100 );
    vLeaves  = Vec_PtrAlloc( 100 );
    vFanouts = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        ppObjs[0] = pObj;
        ppObjs[1] = (Abc_Obj_t *)Vec_PtrEntry( vRoots1, i );
        if ppObjs[1] == nil {
            Vec_PtrPush( vRoots2, nil );
            continue;
        }
        pRoot2    = Abc_NktMffcGrowOne( pNtk, ppObjs, 2, vNodes, vLeaves, vFanouts );
        Vec_PtrPush( vRoots2, pRoot2 );
    }
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vFanouts );
    Abc_NtkCleanMarkA( pNtk );
    assert(Vec_PtrSize(vRoots) == Vec_PtrSize(vRoots2), "Vec_PtrSize(vRoots) == Vec_PtrSize(vRoots2)");
    return vRoots2;
}

/**Function*************************************************************

  Synopsis    [Testbench.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcPrint(pFileName string, pNodes **Abc_Obj_t, nNodes int32, vNodes *Vec_Ptr_t, vLeaves *Vec_Ptr_t) {
    var pFile *os.File;
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    // convert the network
    Abc_NtkToSop( pNodes[0].pNtk, -1, ABC_INFINITY );
    // write the file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, ".model %s_part\n", pNodes[0].pNtk.pName);
    fmt.Fprintf(pFile, ".inputs");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pObj));
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs");
    for  i = 0; i < nNodes; i++  {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pNodes[i]));
    }
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        fmt.Fprintf(pFile, ".names");
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            fmt.Fprintf(pFile, " %s", Abc_ObjName(pFanin));
        }
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pObj));
        fmt.Fprintf(pFile, "\n%s", (char *)pObj.pData);
    }
    fmt.Fprintf(pFile, ".end\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Testbench.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcPrintInt(pFileName string, pNtk *Abc_Ntk_t, vRoots *Vec_Int_t, vNodes *Vec_Int_t, vLeaves *Vec_Int_t) {
    var pFile *os.File;
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    // convert the network
    Abc_NtkToSop( pNtk, -1, ABC_INFINITY );
    // write the file
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, ".model %s_part\n", pNtk.pName);
    fmt.Fprintf(pFile, ".inputs");
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vLeaves) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vLeaves,i))), 1); i++  {         if (pObj) == nil {} else {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pObj)); }
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs");
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vRoots,i))), 1); i++  {         if (pObj) == nil {} else {
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pObj)); }
    }
    fmt.Fprintf(pFile, "\n");
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        fmt.Fprintf(pFile, ".names");
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            fmt.Fprintf(pFile, " %s", Abc_ObjName(pFanin));
        }
        fmt.Fprintf(pFile, " %s", Abc_ObjName(pObj));
        fmt.Fprintf(pFile, "\n%s", (char *)pObj.pData);
    }
    }
    fmt.Fprintf(pFile, ".end\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Testbench.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcTest(pNtk *Abc_Ntk_t) {
    var pFileName string // was [1000]byte in C

    var vRoots, vRoots1, vRoots2, vNodes, vLeaves *Vec_Ptr_t;
    var pNodes [3]*Abc_Obj_t;
    var pObj *Abc_Obj_t;
    var i, nNodes, nNodes2 int32;
    vRoots  = Abc_NktMffcMarkRoots( pNtk, 1 );
    vRoots1 = Abc_NktMffcGrowRoots( pNtk, vRoots );
    vRoots2 = Abc_NktMffcGrowRootsAgain( pNtk, vRoots, vRoots1 );
    vNodes  = Vec_PtrAlloc( 100 );
    vLeaves = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        fmt.Printf("%6d : ", i);

        Abc_MffcCollectNodes( &pObj, 1, vNodes );
        Abc_MffcCollectLeaves( vNodes, vLeaves );
        nNodes += Vec_PtrSize(vNodes);
        fmt.Printf("%6d  ", Abc_ObjId(pObj));
        fmt.Printf("Vol =%3d  ", Vec_PtrSize(vNodes));
        fmt.Printf("Cut =%3d  ", Vec_PtrSize(vLeaves));
        if Vec_PtrSize(vLeaves) < 2 {
            fmt.Printf("\n");
            continue;
        }

        pNodes[0] = pObj;
        pNodes[1] = (Abc_Obj_t *)Vec_PtrEntry( vRoots1, i );
        pNodes[2] = (Abc_Obj_t *)Vec_PtrEntry( vRoots2, i );
        if pNodes[1] == nil || pNodes[2] == nil {
            fmt.Printf("\n");
            continue;
        }
        Abc_MffcCollectNodes( pNodes, 3, vNodes );
        Abc_MffcCollectLeaves( vNodes, vLeaves );
        nNodes2 += Vec_PtrSize(vNodes);
        fmt.Printf("%6d  ", Abc_ObjId(pNodes[1]));
        fmt.Printf("%6d  ", Abc_ObjId(pNodes[2]));
        fmt.Printf("Vol =%3d  ", Vec_PtrSize(vNodes));
        fmt.Printf("Cut =%3d  ", Vec_PtrSize(vLeaves));
 
        fmt.Printf("%4.2f  ", 1.0 * Vec_PtrSize(vNodes)/Vec_PtrSize(vLeaves));
        fmt.Printf("\n");

        // generate file
        if Vec_PtrSize(vNodes) < 10 {
            continue; }
        pFileName = fmt.Sprintf("%s_mffc%04d_%02d.blif", Abc_NtkName(pNtk), Abc_ObjId(pObj), Vec_PtrSize(vNodes));
        Abc_NktMffcPrint( pFileName, pNodes, 3, vNodes, vLeaves );
    }
    fmt.Printf("Total nodes = %d.  Root nodes = %d.  Mffc nodes = %d.  Mffc nodes2 = %d.\n", Abc_NtkNodeNum(pNtk), Vec_PtrSize(vRoots), nNodes, nNodes2);
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vRoots );
    Vec_PtrFree( vRoots1 );
    Vec_PtrFree( vRoots2 );
}



/**Function*************************************************************

  Synopsis    [Create the network of supernodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcTestSuper(pNtk *Abc_Ntk_t) {
    var vRoots, vFanins, vFanouts, vNodes, vLeaves *Vec_Ptr_t;
    var pObj, pFanin *Abc_Obj_t; 
    var vCounts, vNumbers, vSizes, vMarks *Vec_Int_t;
    var vNode1, vNode2 *Vec_Int_t;
    var i, k, Entry, nSizes int32;
    var clk abctime  = Abc_Clock();
    vRoots   = Abc_NktMffcMarkRoots( pNtk, 1 );
    vFanins  = Vec_PtrStart( Abc_NtkObjNumMax(pNtk) );
    vFanouts = Vec_PtrStart( Abc_NtkObjNumMax(pNtk) );
    vCounts  = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    vNode1   = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    vNode2   = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    vSizes   = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    vMarks   = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );

    // create fanins/fanouts
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Vec_PtrWriteEntry( vFanins,  Abc_ObjId(pObj), Vec_IntAlloc(8) );
        Vec_PtrWriteEntry( vFanouts, Abc_ObjId(pObj), Vec_IntAlloc(8) );
    }
    // add fanins/fanouts
    vNodes   = Vec_PtrAlloc( 100 );
    vLeaves  = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Abc_MffcCollectNodes( &pObj, 1, vNodes );
        Abc_MffcCollectLeaves( vNodes, vLeaves );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            if !Abc_ObjIsNode(pFanin) {
                continue; }
            Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vFanins,  Abc_ObjId(pObj)),   Abc_ObjId(pFanin) );
            Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vFanouts, Abc_ObjId(pFanin)), Abc_ObjId(pObj) );
            // count how many times each object is a fanin
            Vec_IntAddToEntry( vCounts, Abc_ObjId(pFanin), 1 );
        }
        Vec_IntWriteEntry( vSizes, Abc_ObjId(pObj), Vec_PtrSize(vNodes) );
    }

    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Abc_MffcCollectNodes( &pObj, 1, vNodes );
        Abc_MffcCollectLeaves( vNodes, vLeaves );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            if !Abc_ObjIsNode(pFanin) {
                continue; }
            if Vec_IntEntry(vCounts, Abc_ObjId(pFanin)) != 2 {
                continue; }
            if Vec_IntEntry(vNode1, Abc_ObjId(pFanin)) == 0 {
                Vec_IntWriteEntry( vNode1, Abc_ObjId(pFanin), Abc_ObjId(pObj) );
            } else { //if ( Vec_IntEntry(vNode2, Abc_ObjId(pFanin)) == 0 )
                Vec_IntWriteEntry( vNode2, Abc_ObjId(pFanin), Abc_ObjId(pObj) ); }

            Vec_IntWriteEntry( vMarks, Abc_ObjId(pFanin), 1 );
            Vec_IntWriteEntry( vMarks, Abc_ObjId(pObj),   1 );
        }
    }

    // count sizes
    nSizes = 0;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        if Vec_IntEntry( vMarks, Abc_ObjId(pObj) ) {
            nSizes += Vec_IntEntry( vSizes, Abc_ObjId(pObj) ); }
    }
    fmt.Printf("Included = %6d.  Total = %6d. (%6.2f %%)\n", nSizes, Abc_NtkNodeNum(pNtk), 100.0 * nSizes / Abc_NtkNodeNum(pNtk));


    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        if Vec_IntEntry(vCounts, Abc_ObjId(pObj)) != 2 {
            continue; }
        fmt.Printf("%d ", Vec_IntEntry( vSizes, Abc_ObjId(pObj) ) + 
                               Vec_IntEntry( vSizes, Vec_IntEntry(vNode1, Abc_ObjId(pObj)) ) + 
                               Vec_IntEntry( vSizes, Vec_IntEntry(vNode2, Abc_ObjId(pObj)) ));
    }
    fmt.Printf("\n");

    // print how many times they appear
    vNumbers = Vec_IntStart( 32 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCounts)) && (((Entry) = Vec_IntEntry(vCounts, i)), 1); i++  {
/*
if ( Entry == 2 )
{
    pObj = Abc_NtkObj( pNtk, i );
    Abc_MffcCollectNodes( &pObj, 1, vNodes );
    Abc_MffcCollectLeaves( vNodes, vLeaves );
    printf( "%d(%d) ", Vec_PtrSize(vNodes), Vec_PtrSize(vLeaves) );
}
*/
        if Entry == 0 {
            continue; }
        if Entry <= 10 {
            Vec_IntAddToEntry( vNumbers, Entry, 1 );
        } else if Entry <= 100 {
            Vec_IntAddToEntry( vNumbers, 10 + Entry/10, 1 );
        } else if Entry < 1000 {
            Vec_IntAddToEntry( vNumbers, 20 + Entry/100, 1 );
        } else {
            Vec_IntAddToEntry( vNumbers, 30, 1 ); }
    }
    for  i = 1; i <= 10; i++  {
        if Vec_IntEntry(vNumbers,i) {
            fmt.Printf("       n =  %4d   %6d\n", i, Vec_IntEntry(vNumbers,i)); }
    }
    for  i = 11; i <= 20; i++  {
        if Vec_IntEntry(vNumbers,i) {
            fmt.Printf("%4d < n <= %4d   %6d\n", 10*(i-10), 10*(i-9), Vec_IntEntry(vNumbers,i)); }
    }
    for  i = 21; i < 30; i++  {
        if Vec_IntEntry(vNumbers,i) {
            fmt.Printf("%4d < n <= %4d   %6d\n", 100*(i-20), 100*(i-19), Vec_IntEntry(vNumbers,i)); }
    }
    if Vec_IntEntry(vNumbers,31) {
    fmt.Printf("       n >  1000   %6d\n", Vec_IntEntry(vNumbers,30)); }
    fmt.Printf("Total MFFCs = %d. ", Vec_PtrSize(vRoots));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Vec_IntFree( vNumbers );
    Vec_PtrFree( vNodes );
    Vec_PtrFree( vLeaves );

    // delete fanins/fanouts
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vFanins,  Abc_ObjId(pObj)) );
        Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vFanouts, Abc_ObjId(pObj)) );
    }

    Vec_IntFree( vCounts );
    Vec_PtrFree( vFanouts );
    Vec_PtrFree( vFanins );
    Vec_PtrFree( vRoots );
}


/**Function*************************************************************

  Synopsis    [Collects the leaves and the roots of the window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffCollectLeafRoot(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, vLeaves *Vec_Ptr_t, vRoots *Vec_Ptr_t) {
    var pObj, pNext *Abc_Obj_t;
    var i, k int32;
    // mark
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.fMarkA = 1;
    }
    // collect leaves
    Vec_PtrClear( vLeaves );
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if pNext.fMarkA || Abc_NodeIsTravIdCurrent(pNext) {
                continue; }
            Abc_NodeSetTravIdCurrent(pNext);
            Vec_PtrPush( vLeaves, pNext );
        }
    }
    // collect roots
    Vec_PtrClear( vRoots );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
            if !pNext.fMarkA {
                Vec_PtrPush( vRoots, pObj );
                break;
            }
        }
    }
    // unmark
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.fMarkA = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Collects the leaves and the roots of the window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffCollectLeafRootInt(pNtk *Abc_Ntk_t, vNodes *Vec_Int_t, vLeaves *Vec_Int_t, vRoots *Vec_Int_t) {
    var pObj, pNext *Abc_Obj_t;
    var i, k int32;
    // mark
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.fMarkA = 1; }
    }
    // collect leaves
    Vec_IntClear( vLeaves );
    Abc_NtkIncrementTravId( pNtk );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pNext) = Abc_ObjFanin(pObj, k)), 1); k++  {
            if pNext.fMarkA || Abc_NodeIsTravIdCurrent(pNext) {
                continue; }
            Abc_NodeSetTravIdCurrent(pNext);
            Vec_IntPush( vLeaves, Abc_ObjId(pNext) );
        } }
    }
    // collect roots
    if vRoots {
        Vec_IntClear( vRoots );
        // inlining of macro Abc_NtkForEachObjVec
        for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
            // inlining of macro Abc_ObjForEachFanout
            for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, k)), 1); k++  {
                if !pNext.fMarkA {
                    Vec_IntPush( vRoots, Abc_ObjId(pObj) );
                    break;
                }
            }
        }
        }
    }
    // unmark
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vNodes) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vNodes,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.fMarkA = 0; }
    }
}


/**Function*************************************************************

  Synopsis    [Create the network of supernodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcTestIdeaOne(pNtk *Abc_Ntk_t, pObj *Abc_Obj_t) {
    var vNodes, vLeaves, vRoots, vVolume *Vec_Ptr_t;
    var vLeaves2, vRoots2, vVolume2 *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var pNodeBest *Abc_Obj_t = pObj;
    var Cost float64;
    var CostBest float64 = 0.0;
    var i, k int32;
    vNodes   = Vec_PtrAlloc( 100 );
    vLeaves  = Vec_PtrAlloc( 100 );
    vRoots   = Vec_PtrAlloc( 100 );
    vVolume  = Vec_PtrAlloc( 100 );
    vLeaves2 = Vec_PtrAlloc( 100 );
    vRoots2  = Vec_PtrAlloc( 100 );
    vVolume2 = Vec_PtrAlloc( 100 );
fmt.Printf("\n");
    for  i = 1; i <= 16; i++  {
        Vec_PtrPush( vNodes, pNodeBest );
        Abc_NktMffCollectLeafRoot( pNtk, vNodes, vLeaves, vRoots );
        Abc_MffcCollectNodes( (Abc_Obj_t **)Vec_PtrArray(vRoots), Vec_PtrSize(vRoots), vVolume );

        fmt.Printf("%2d : Node =%6d (%2d%3d)  Cost =%6.2f   ", i, Abc_ObjId(pNodeBest), Abc_ObjFaninNum(pNodeBest), Abc_ObjFanoutNum(pNodeBest), CostBest);
        fmt.Printf("Leaf =%2d  Root =%2d  Vol =%2d\n", Vec_PtrSize(vLeaves), Vec_PtrSize(vRoots), Vec_PtrSize(vVolume));

        // try including different nodes
        pNodeBest = nil;
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            if !Abc_ObjIsNode(pNode) {
                continue; }
            Vec_PtrPush( vNodes, pNode );
            Abc_NktMffCollectLeafRoot( pNtk, vNodes, vLeaves2, vRoots2 );
            Abc_MffcCollectNodes( (Abc_Obj_t **)Vec_PtrArray(vRoots2), Vec_PtrSize(vRoots2), vVolume2 );
            Cost = 1.0 * Vec_PtrSize(vVolume2) / (Vec_PtrSize(vLeaves2) + 3 * Vec_PtrSize(vRoots2));
            if pNodeBest == nil || CostBest < Cost {
                pNodeBest = pNode;
                CostBest  = Cost;
            }
            Vec_PtrPop( vNodes );
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vRoots)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, k)), 1); k++  {
            if Vec_PtrFind(vNodes, pNode) >= 0 {
                continue; }
            if !Abc_ObjIsNode(pNode) {
                continue; }
            Vec_PtrPush( vNodes, pNode );
            Abc_NktMffCollectLeafRoot( pNtk, vNodes, vLeaves2, vRoots2 );
            Abc_MffcCollectNodes( (Abc_Obj_t **)Vec_PtrArray(vRoots2), Vec_PtrSize(vRoots2), vVolume2 );
            Cost = 1.0 * Vec_PtrSize(vVolume2) / (Vec_PtrSize(vLeaves2) + 3 * Vec_PtrSize(vRoots2));
            if pNodeBest == nil || CostBest < Cost {
                pNodeBest = pNode;
                CostBest  = Cost;
            }
            Vec_PtrPop( vNodes );
        }
        if pNodeBest == nil {
            break; }
    }

    Vec_PtrFree( vNodes );
    Vec_PtrFree( vLeaves );
    Vec_PtrFree( vRoots );
    Vec_PtrFree( vVolume );
    Vec_PtrFree( vLeaves2 );
    Vec_PtrFree( vRoots2 );
    Vec_PtrFree( vVolume2 );
}

/**Function*************************************************************

  Synopsis    [Create the network of supernodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcTestIdea(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t; 
    var i int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
        if Abc_ObjIsNode(pObj) && Abc_ObjId(pObj) % 100 == 0 {
            Abc_NktMffcTestIdeaOne( pNtk, pObj ); }
    }
}





/**Function*************************************************************

  Synopsis    [Creates MFFCs and their fanins/fanouts/volumes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcDerive(pNtk *Abc_Ntk_t, pvFanins **Vec_Ptr_t, pvFanouts **Vec_Ptr_t, pvVolumes **Vec_Ptr_t) *Vec_Ptr_t {
    var vRoots, vFanins, vFanouts, vVolumes, vNodes, vLeaves *Vec_Ptr_t;
    var pObj, pFanin *Abc_Obj_t; 
    var i, k int32;
    var clk abctime  = Abc_Clock();
    // create roots
    vRoots   = Abc_NktMffcMarkRoots( pNtk, 0 );
    // create fanins/fanouts/volumes
    vFanins  = Vec_PtrStart( Abc_NtkObjNumMax(pNtk) );
    vFanouts = Vec_PtrStart( Abc_NtkObjNumMax(pNtk) );
    vVolumes = Vec_PtrStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Vec_PtrWriteEntry( vFanins,  Abc_ObjId(pObj), Vec_IntAlloc(8) );
        Vec_PtrWriteEntry( vFanouts, Abc_ObjId(pObj), Vec_IntAlloc(8) );
        Vec_PtrWriteEntry( vVolumes, Abc_ObjId(pObj), Vec_IntAlloc(8) );
    }
    // add fanins/fanouts
    vNodes   = Vec_PtrAlloc( 100 );
    vLeaves  = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        if Abc_ObjIsCi(pObj) {
            continue; }
        Abc_MffcCollectNodes( &pObj, 1, vNodes );
        Abc_MffcCollectLeaves( vNodes, vLeaves );
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vLeaves)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, k)), 1); k++  {
            Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vFanins,  Abc_ObjId(pObj)),   Abc_ObjId(pFanin) );
            Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vFanouts, Abc_ObjId(pFanin)), Abc_ObjId(pObj) );
        }
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vNodes)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, k)), 1); k++  {
            Vec_IntPush( (Vec_Int_t *)Vec_PtrEntry(vVolumes, Abc_ObjId(pObj)),   Abc_ObjId(pFanin) );
        }
    }

    Vec_PtrFree( vNodes );
    Vec_PtrFree( vLeaves );
    // sort
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Vec_IntSort( (Vec_Int_t *)Vec_PtrEntry(vFanins,  Abc_ObjId(pObj)), 0 );
        Vec_IntSort( (Vec_Int_t *)Vec_PtrEntry(vFanouts, Abc_ObjId(pObj)), 0 );
    }
    // return
    *pvFanins  = vFanins;
    *pvFanouts = vFanouts;
    *pvVolumes = vVolumes;
    return vRoots;
}

/**Function*************************************************************

  Synopsis    [Frees MFFCs and their fanins/fanouts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcFree(vRoots *Vec_Ptr_t, vFanins *Vec_Ptr_t, vFanouts *Vec_Ptr_t, vVolumes *Vec_Ptr_t) {
    var pObj *Abc_Obj_t; 
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vFanins,  Abc_ObjId(pObj)) );
        Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vFanouts, Abc_ObjId(pObj)) );
        Vec_IntFree( (Vec_Int_t *)Vec_PtrEntry(vVolumes, Abc_ObjId(pObj)) );
    }
    Vec_PtrFree( vVolumes );
    Vec_PtrFree( vFanouts );
    Vec_PtrFree( vFanins );
    Vec_PtrFree( vRoots );
}


/**Function*************************************************************

  Synopsis    [Returns the cost of two supports.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcCostTwo(vSupp1 *Vec_Int_t, vSupp2 *Vec_Int_t, Volume int32, Limit int32) float64 {
    var nCommon int32  = Vec_IntTwoCountCommon( vSupp1, vSupp2 );
//printf( "s1=%2d s2=%2d c=%2d v=%2d   ", Vec_IntSize(vSupp1), Vec_IntSize(vSupp2), nCommon, Volume );
    if Vec_IntSize(vSupp1) + Vec_IntSize(vSupp2) - nCommon > Limit {
        return float64(-ABC_INFINITY); }
    return 0.6 * nCommon - 1.2 * Vec_IntSize(vSupp2) + 0.8 * Volume;
}

/**Function*************************************************************

  Synopsis    [Returns support of the group.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcSupport(vThis *Vec_Ptr_t, vFanins *Vec_Ptr_t) *Vec_Int_t {
    var vIns, vIns2, vTemp *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i int32;
    vIns = Vec_IntAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vThis)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vThis, i)), 1); i++  {
        vIns2 = (Vec_Int_t *)Vec_PtrEntry( vFanins, Abc_ObjId(pObj) );
        vIns  = Vec_IntTwoMerge( vTemp = vIns, vIns2 );
        Vec_IntFree( vTemp );
    }
    return vIns;
}

/**Function*************************************************************

  Synopsis    [Returns the best merger for the cluster of given node (pPivot).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcFindBest(pNtk *Abc_Ntk_t, vMarks *Vec_Int_t, vIns *Vec_Int_t, vFanins *Vec_Ptr_t, vFanouts *Vec_Ptr_t, vVolumes *Vec_Ptr_t, Limit int32) *Abc_Obj_t {
    var vIns2, vOuts, vOuts2, vTemp *Vec_Int_t;
    var pPivot2 *Abc_Obj_t;
    var pObj *Abc_Obj_t;
    var pObjBest *Abc_Obj_t = nil;
    var Cost float64;
    var CostBest float64 = float64(-ABC_INFINITY);
    var i, Volume int32;
    // collect the fanouts of the fanins
    vOuts = Vec_IntAlloc( 100 );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vIns) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vIns,i))), 1); i++  {         if (pObj) == nil {} else {
        vOuts2 = (Vec_Int_t *)Vec_PtrEntry( vFanouts, Abc_ObjId(pObj) );
        if Vec_IntSize(vOuts2) > 16 {
            continue; }
        vOuts  = Vec_IntTwoMerge( vTemp = vOuts, vOuts2 );
        Vec_IntFree( vTemp );
    }
    }
    // check the pairs
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vOuts) && (((pPivot2) = Abc_NtkObj(pNtk, Vec_IntEntry(vOuts,i))), 1); i++  {         if (pPivot2) == nil {} else {       
        if Vec_IntEntry(vMarks, Abc_ObjId(pPivot2)) == 0 {
            continue; }
        vIns2  = (Vec_Int_t *)Vec_PtrEntry( vFanins, Abc_ObjId(pPivot2) );
        Volume = Vec_IntSize((Vec_Int_t *)Vec_PtrEntry(vVolumes, Abc_ObjId(pPivot2)));
        Cost   = Abc_NktMffcCostTwo( vIns, vIns2, Volume, Limit );
//printf( "%5d  %2d\n", Abc_ObjId(pPivot2), Cost  );
        if Cost == float64(-ABC_INFINITY) {
            continue; }
        if pObjBest == nil || CostBest < Cost {
            pObjBest = pPivot2;
            CostBest = Cost;
        }
    }
    }    
//printf( "Choosing %d\n", pObjBest->Id );
    Vec_IntFree( vOuts );
    return pObjBest;
}

/**Function*************************************************************

  Synopsis    [Processes one cluster.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcSaveOne(vThis *Vec_Ptr_t, vVolumes *Vec_Ptr_t) *Vec_Int_t {
    var vVolume, vResult *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, k, Entry int32;
    vResult = Vec_IntAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vThis)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vThis, i)), 1); i++  {
        vVolume = (Vec_Int_t *)Vec_PtrEntry( vVolumes, Abc_ObjId(pObj) );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVolume)) && (((Entry) = Vec_IntEntry(vVolume, k)), 1); k++  {
            Vec_IntPush( vResult, Entry );
        }
    }
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in decreasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeCompareVolumeDecrease(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
    var Diff int32  = Abc_ObjRegular(*pp1).iTemp - Abc_ObjRegular(*pp2).iTemp;
    if Diff > 0 {
        return -1; }
    if Diff < 0 { 
        return 1; }
    Diff = Abc_ObjRegular(*pp1).Id - Abc_ObjRegular(*pp2).Id;
    if Diff > 0 {
        return -1; }
    if Diff < 0 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Create the network of supernodes.]

  Description [Returns array of interger arrays of IDs of nodes
  included in a disjoint structural decomposition of the network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcServer(pNtk *Abc_Ntk_t, nInMax int32, nOutMax int32) *Vec_Ptr_t {
    var vResult, vThis *Vec_Ptr_t;
    var vPivots, vFanins, vFanouts, vVolumes *Vec_Ptr_t;
    var vLeaves, vMarks *Vec_Int_t;
    var pObj, pObj2 *Abc_Obj_t;
    var i, k int32;
    assert(nOutMax >= 1 && nOutMax <= 32, "nOutMax >= 1 && nOutMax <= 32");
    vResult = Vec_PtrAlloc( 100 );
    // create fanins/fanouts
    vPivots = Abc_NktMffcDerive( pNtk, &vFanins, &vFanouts, &vVolumes );
    // sort by their MFFC size
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPivots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vPivots, i)), 1); i++  {
        pObj.iTemp = Vec_IntSize((Vec_Int_t *)Vec_PtrEntry(vVolumes, Abc_ObjId(pObj)));
    }
    Vec_PtrSort( vPivots, (int (*)(void))Abc_NodeCompareVolumeDecrease ); 
    // create marks
    vMarks = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPivots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vPivots, i)), 1); i++  {
        if Abc_ObjIsNode(pObj) && Vec_IntSize((Vec_Int_t *)Vec_PtrEntry(vVolumes, Abc_ObjId(pObj))) > 1 {
            Vec_IntWriteEntry( vMarks, Abc_ObjId(pObj), 1 ); }
    }
    // consider nodes in the order of the marks
    vThis = Vec_PtrAlloc( 10 );
//    while ( 1 )
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPivots)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vPivots, i)), 1); i++  {
//        pObj = Abc_NtkObj( pNtk, 589 );
        if Vec_IntEntry(vMarks, Abc_ObjId(pObj)) == 0 {
            continue; }
        // start the set
        Vec_PtrClear( vThis );        
        Vec_PtrPush( vThis, pObj );
        Vec_IntWriteEntry( vMarks, Abc_ObjId(pObj), 0 );
        // quit if exceeded the limit
        vLeaves = (Vec_Int_t *)Vec_PtrEntry( vFanins, Abc_ObjId(pObj) );
        if Vec_IntSize(vLeaves) > nInMax {
            Vec_PtrPush( vResult, Abc_NktMffcSaveOne(vThis, vVolumes) );
            continue;
        }
        // try adding one node at a time
        for  k = 1; k < nOutMax; k++  {
            // quit if exceeded the limit
            vLeaves = Abc_NktMffcSupport( vThis, vFanins );
            assert(Vec_IntSize(vLeaves) <= nInMax, "Vec_IntSize(vLeaves) <= nInMax");
            pObj2 = Abc_NktMffcFindBest( pNtk, vMarks, vLeaves, vFanins, vFanouts, vVolumes, nInMax );
            Vec_IntFree( vLeaves );
            // quit if there is no extension
            if pObj2 == nil {
                break; }
            Vec_PtrPush( vThis, pObj2 );
            Vec_IntWriteEntry( vMarks, Abc_ObjId(pObj2), 0 );
        }
        Vec_PtrPush( vResult, Abc_NktMffcSaveOne(vThis, vVolumes) );
//        break;
    }
    Vec_PtrFree( vThis );
    Vec_IntFree( vMarks );
    // delele fanins/outputs
    Abc_NktMffcFree( vPivots, vFanins, vFanouts, vVolumes );
    return vResult;
}

/**Function*************************************************************

  Synopsis    [Testbench.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NktMffcServerTest(pNtk *Abc_Ntk_t) {
    var pFileName string // was [1000]byte in C

    var vGlobs *Vec_Ptr_t;
    var vGlob, vLeaves, vRoots *Vec_Int_t;
    var Cost float64;
    var CostAll float64 = 0.0;
    var i, k, Entry, nNodes int32;
    var clk abctime  = Abc_Clock();
    vGlobs  = Abc_NktMffcServer( pNtk, 18, 3 );
    vLeaves = Vec_IntAlloc( 100 );
    vRoots  = Vec_IntAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGlobs)) && (((vGlob) = (Vec_Int_t *)Vec_PtrEntry(vGlobs, i)), 1); i++  {
        nNodes += Vec_IntSize(vGlob);
        Abc_NktMffCollectLeafRootInt( pNtk, vGlob, vLeaves, vRoots );
        if Vec_IntSize(vGlob) <= Vec_IntSize(vRoots) {
            continue; }
        Cost = 1.0 * Vec_IntSize(vGlob)/(Vec_IntSize(vLeaves) + Vec_IntSize(vRoots));
        CostAll += Cost;
        if Cost < 0.5 {
            continue; }

        fmt.Printf("%6d : Root =%3d. Leaf =%3d. Node =%4d. ", i, Vec_IntSize(vRoots), Vec_IntSize(vLeaves), Vec_IntSize(vGlob));
        fmt.Printf("Cost =%6.2f     ", Cost);
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vRoots)) && (((Entry) = Vec_IntEntry(vRoots, k)), 1); k++  {
            fmt.Printf("%d ", Entry);
        }
        fmt.Printf("\n");

        pFileName = fmt.Sprintf("%sc%04di%02dn%02d.blif", Abc_NtkName(pNtk), i, Vec_IntSize(vLeaves), Vec_IntSize(vGlob));
        Abc_NktMffcPrintInt( pFileName, pNtk, vRoots, vGlob, vLeaves );
    }
    Vec_IntFree( vLeaves );
    Vec_IntFree( vRoots );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vGlobs)) && (((vGlob) = (Vec_Int_t *)Vec_PtrEntry(vGlobs, i)), 1); i++  {
        Vec_IntFree( vGlob );
    }
    Vec_PtrFree( vGlobs );
    fmt.Printf("Total = %6d.  Nodes = %6d.  ", Abc_NtkNodeNum(pNtk), nNodes);
    fmt.Printf("Cost = %6.2f  ", CostAll);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

//ABC_NAMESPACE_IMPL_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


