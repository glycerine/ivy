package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcMfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Optimization with don't-cares.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcMfs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "bool/kit/kit.h"
//#include "opt/sfm/sfm.h"
//#include "base/io/ioAbc.h"

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
func Abc_NtkAssignIDs(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    vNodes = Abc_NtkDfs( pNtk, 0 );
    Abc_NtkCleanCopy( pNtk );
    Abc_NtkForEachCi( pNtk, pObj, i )
        pObj.iTemp = i;
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
        pObj.iTemp = Abc_NtkCiNum(pNtk) + i;
    Abc_NtkForEachCo( pNtk, pObj, i )
        pObj.iTemp = Abc_NtkCiNum(pNtk) + Vec_PtrSize(vNodes) + i;
    return vNodes;
}
func Abc_NtkAssignIDs2(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_NtkCleanCopy( pNtk );
    Abc_NtkForEachCi( pNtk, pObj, i )
        pObj.iTemp = i;
    vNodes = Vec_PtrAlloc( Abc_NtkNodeNum(pNtk) );
    Abc_NtkForEachNode( pNtk, pObj, i )
    {
        pObj.iTemp = Abc_NtkCiNum(pNtk) + Vec_PtrSize(vNodes);
        Vec_PtrPush( vNodes, pObj );
    }
    assert(Vec_PtrSize(vNodes) == Abc_NtkNodeNum(pNtk), "Vec_PtrSize(vNodes) == Abc_NtkNodeNum(pNtk)");
    Abc_NtkForEachCo( pNtk, pObj, i )
        pObj.iTemp = Abc_NtkCiNum(pNtk) + Vec_PtrSize(vNodes) + i;
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Assign truth table sizes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAssignStarts(pNtk *Abc_Ntk_t, vNodes *Vec_Ptr_t, pnTotal *int32) *Vec_Int_t {
    var pObj *Abc_Obj_t; var i, Counter int32;
    var vStarts *Vec_Int_t  = Vec_IntStart( Abc_NtkObjNum(pNtk) );
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
    {
        Vec_IntWriteEntry( vStarts, pObj.iTemp, Counter );
        Counter += Abc_Truth6WordNum( Abc_ObjFaninNum(pObj) );
    }
    *pnTotal = Counter;
    return vStarts;
}

var spStatic_Abc_NtkFillTruthStore_Truth6 [6]word = [6]word{
            0xAAAAAAAAAAAAAAAA,
            0xCCCCCCCCCCCCCCCC,
            0xF0F0F0F0F0F0F0F0,
            0xFF00FF00FF00FF00,
            0xFFFF0000FFFF0000,
            0xFFFFFFFF00000000,
        };
func Abc_NtkFillTruthStore(TruthStore [][1<<10]word) {
    if TruthStore[0][0] == 0 {
        
        var nVarsMax int32  = 16;
        var nWordsMax int32  = (1 << 10);
        var i, k int32;
        assert(nVarsMax <= 16, "nVarsMax <= 16");
        for  i = 0; i < 6; i++  {
            for  k = 0; k < nWordsMax; k++  {
                TruthStore[i][k] = spStatic_Abc_NtkFillTruthStore_Truth6[i];
            }
        }
        for  i = 6; i < nVarsMax; i++  {
            for  k = 0; k < nWordsMax; k++  {
                TruthStore[i][k] = ((k >> (i-6)) & 1) ? ~(word)0 : 0;
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Extracts information about the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkExtractMfs(pNtk *Abc_Ntk_t, nFirstFixed int32) *Sfm_Ntk_t {
    var TruthStore [16][1<<10]word = [16][1<<10]word{{0}};
    var pTruths [16]*word = [16]*word{nil};
    var pCube [1<<10]word ;
    var vNodes *Vec_Ptr_t;
    var vFanins *Vec_Wec_t;
    var vFixed *Vec_Str_t;
    var vTruths *Vec_Wrd_t;
    var vArray *Vec_Int_t;
    var vStarts *Vec_Int_t;
    var vTruths2 *Vec_Wrd_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k, nObjs, nTotal int32;
    if nFirstFixed != 0 {
        vNodes = Abc_NtkAssignIDs2(pNtk);
    } else {
        vNodes = Abc_NtkAssignIDs(pNtk);
    }
    nObjs   = Abc_NtkCiNum(pNtk) + Vec_PtrSize(vNodes) + Abc_NtkCoNum(pNtk);
    vFanins = Vec_WecStart( nObjs );
    vFixed  = Vec_StrStart( nObjs );
    vTruths = Vec_WrdStart( nObjs );
    vStarts = Abc_NtkAssignStarts( pNtk, vNodes, &nTotal );
    vTruths2= Vec_WrdStart( nTotal );
    Abc_NtkFillTruthStore( TruthStore );
    for  i = 0; i < 16; i++  {
        pTruths[i] = TruthStore[i];
    }
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
    {
        if Abc_ObjFaninNum(pObj) <= 6 {
            var uTruth word  = Abc_SopToTruth((char *)pObj.pData, Abc_ObjFaninNum(pObj));
            Vec_WrdWriteEntry( vTruths, pObj.iTemp, uTruth );
            if uTruth == 0 || ~uTruth == 0 {
                continue; }
        
        } else {
            var nWords int32   = Abc_Truth6WordNum( Abc_ObjFaninNum(pObj) );
            var Offset int32   = Vec_IntEntry( vStarts, pObj.iTemp );
            var pRes *word  = Vec_WrdEntryP( vTruths2, Offset );
            Abc_SopToTruthBig( (char *)pObj.pData, Abc_ObjFaninNum(pObj), pTruths, pCube, pRes );
            // check const0
            for  k = 0; k < nWords; k++  {
                if pRes[k] {
                    break; }
            }
            if k == nWords {
                continue; }
            // check const1
            for  k = 0; k < nWords; k++  {
                if ~pRes[k] {
                    break; }
            }
            if k == nWords {
                continue; }
        }
        vArray = Vec_WecEntry( vFanins, pObj.iTemp );
        Vec_IntGrow( vArray, Abc_ObjFaninNum(pObj) );
        Abc_ObjForEachFanin( pObj, pFanin, k )
            Vec_IntPush( vArray, pFanin.iTemp );
        //Vec_IntPrint( vArray );
    }
    Abc_NtkForEachCo( pNtk, pObj, i )
    {
        vArray = Vec_WecEntry( vFanins, pObj.iTemp );
        Vec_IntGrow( vArray, Abc_ObjFaninNum(pObj) );
        Abc_ObjForEachFanin( pObj, pFanin, k )
            Vec_IntPush( vArray, pFanin.iTemp );
    }
    Vec_PtrFree( vNodes );
    for  i = Abc_NtkCiNum(pNtk); i < Abc_NtkCiNum(pNtk) + nFirstFixed; i++  {
        Vec_StrWriteEntry( vFixed, i, byte(1) );
    }
    // update fixed
    assert(nFirstFixed >= 0 && nFirstFixed < Abc_NtkNodeNum(pNtk), "nFirstFixed >= 0 && nFirstFixed < Abc_NtkNodeNum(pNtk)");
//    for ( i = Abc_NtkCiNum(pNtk); i + Abc_NtkCoNum(pNtk) < Abc_NtkObjNum(pNtk); i++ )
//        if ( rand() % 10 == 0 )
//            Vec_StrWriteEntry( vFixed, i, (char)1 );
    return Sfm_NtkConstruct( vFanins, Abc_NtkCiNum(pNtk), Abc_NtkCoNum(pNtk), vFixed, nil, vTruths, vStarts, vTruths2 );
}
func Abc_NtkExtractMfs2(pNtk *Abc_Ntk_t, iPivot int32) *Sfm_Ntk_t {
    var TruthStore [16][1<<10]word = [16][1<<10]word{{0}};
    var pTruths [16]*word = [16]*word{nil};
    var pCube [1<<10]word ;
    var vNodes *Vec_Ptr_t;
    var vFanins *Vec_Wec_t;
    var vFixed *Vec_Str_t;
    var vTruths *Vec_Wrd_t;
    var vArray *Vec_Int_t;
    var vStarts *Vec_Int_t;
    var vTruths2 *Vec_Wrd_t;
    var pObj, pFanin *Abc_Obj_t;
    var i, k, nObjs, nTotal int32;
    vNodes  = Abc_NtkAssignIDs2(pNtk);
    nObjs   = Abc_NtkCiNum(pNtk) + Vec_PtrSize(vNodes) + Abc_NtkCoNum(pNtk);
    vFanins = Vec_WecStart( nObjs );
    vFixed  = Vec_StrStart( nObjs );
    vTruths = Vec_WrdStart( nObjs );
    vStarts = Abc_NtkAssignStarts( pNtk, vNodes, &nTotal );
    vTruths2= Vec_WrdAlloc( nTotal );
    Abc_NtkFillTruthStore( TruthStore );
    for  i = 0; i < 16; i++  {
        pTruths[i] = TruthStore[i];
    }
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pObj, i )
    {
        if Abc_ObjFaninNum(pObj) <= 6 {
            var uTruth word  = Abc_SopToTruth((char *)pObj.pData, Abc_ObjFaninNum(pObj));
            Vec_WrdWriteEntry( vTruths, pObj.iTemp, uTruth );
            if uTruth == 0 || ~uTruth == 0 {
                continue; }
        
        } else {
            var nWords int32   = Abc_Truth6WordNum( Abc_ObjFaninNum(pObj) );
            var Offset int32   = Vec_IntEntry( vStarts, pObj.iTemp );
            var pRes *word  = Vec_WrdEntryP( vTruths2, Offset );
            Abc_SopToTruthBig( (char *)pObj.pData, Abc_ObjFaninNum(pObj), pTruths, pCube, pRes );
            // check const0
            for  k = 0; k < nWords; k++  {
                if pRes[k] {
                    break; }
            }
            if k == nWords {
                continue; }
            // check const1
            for  k = 0; k < nWords; k++  {
                if ~pRes[k] {
                    break; }
            }
            if k == nWords {
                continue; }
        }
        vArray = Vec_WecEntry( vFanins, pObj.iTemp );
        Vec_IntGrow( vArray, Abc_ObjFaninNum(pObj) );
        Abc_ObjForEachFanin( pObj, pFanin, k )
            Vec_IntPush( vArray, pFanin.iTemp );
    }
    Abc_NtkForEachCo( pNtk, pObj, i )
    {
        vArray = Vec_WecEntry( vFanins, pObj.iTemp );
        Vec_IntGrow( vArray, Abc_ObjFaninNum(pObj) );
        Abc_ObjForEachFanin( pObj, pFanin, k )
            Vec_IntPush( vArray, pFanin.iTemp );
    }
    Vec_PtrFree( vNodes );
    // set fixed attributes
    Abc_NtkForEachNode( pNtk, pObj, i )
        if ( i >= iPivot )
            Vec_StrWriteEntry( vFixed, pObj.iTemp, byte(1) );
    return Sfm_NtkConstruct( vFanins, Abc_NtkCiNum(pNtk), Abc_NtkCoNum(pNtk), vFixed, nil, vTruths, vStarts, vTruths2 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkInsertMfs(pNtk *Abc_Ntk_t, p *Sfm_Ntk_t) {
    var vCover, vMap, vArray *Vec_Int_t;
    var pNode *Abc_Obj_t;
    var pTruth *word;
    var i, k, Fanin int32;
    // map new IDs into old nodes
    vMap = Vec_IntStart( Abc_NtkObjNumMax(pNtk) );
    Abc_NtkForEachCi( pNtk, pNode, i )
        Vec_IntWriteEntry( vMap, pNode.iTemp, Abc_ObjId(pNode) );
    Abc_NtkForEachNode( pNtk, pNode, i )
        if ( pNode.iTemp > 0 )
            Vec_IntWriteEntry( vMap, pNode.iTemp, Abc_ObjId(pNode) );
    // remove old fanins
    Abc_NtkForEachNode( pNtk, pNode, i )
        if ( !Sfm_NodeReadFixed(p, pNode.iTemp) )
            Abc_ObjRemoveFanins( pNode );
    // create new fanins
    vCover = Vec_IntAlloc( 1 << 16 );
    Abc_NtkForEachNode( pNtk, pNode, i )
    {
        if pNode.iTemp == 0 || Sfm_NodeReadFixed(p, pNode.iTemp) {
            continue; }
        if !Sfm_NodeReadUsed(p, pNode.iTemp) {
            Abc_NtkDeleteObj( pNode );
            continue;
        }
        // update fanins
        vArray = Sfm_NodeReadFanins( p, pNode.iTemp );
        Vec_IntForEachEntry( vArray, Fanin, k )
            Abc_ObjAddFanin( pNode, Abc_NtkObj(pNtk, Vec_IntEntry(vMap, Fanin)) );
        pTruth = Sfm_NodeReadTruth( p, pNode.iTemp );
        pNode.pData = Abc_SopCreateFromTruthIsop( (Mem_Flex_t *)pNtk.pManFunc, Vec_IntSize(vArray), pTruth, vCover );
        assert(Abc_SopGetVarNum((char *)pNode.pData) == Vec_IntSize(vArray), "Abc_SopGetVarNum((char *)pNode.pData) == Vec_IntSize(vArray)");
    }
    Vec_IntFree( vCover );
    Vec_IntFree( vMap );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPerformMfs(pNtk *Abc_Ntk_t, pPars *Sfm_Par_t) int32 {
    var p *Sfm_Ntk_t;
    var nFaninMax, nNodes int32;
    assert(Abc_NtkIsLogic(pNtk), "Abc_NtkIsLogic(pNtk)");
    Abc_NtkSweep( pNtk, 0 );
    // count fanouts
    nFaninMax = Abc_NtkGetFaninMax( pNtk );
    if nFaninMax > 15 {
        Abc_Print( 1, "Currently \"mfs\" cannot process the network containing nodes with more than 15 fanins.\n" );
        return 1;
    }
    if !Abc_NtkHasSop(pNtk) {
        if !Abc_NtkToSop( pNtk, -1, ABC_INFINITY ) {
            fmt.Printf("Conversion to SOP has failed due to low resource limit.\n");
            return 0;
        } }
    // collect information
    p = Abc_NtkExtractMfs( pNtk, pPars.nFirstFixed );
    // perform optimization
    nNodes = Sfm_NtkPerform( p, pPars );
    if nNodes == 0 {
//        Abc_Print( 1, "The network is not changed by \"mfs\".\n" );
    
    } else {
        Abc_NtkInsertMfs( pNtk, p );
        if pPars.fVerbose {
            Abc_Print( 1, "The network has %d nodes changed by \"mfs\".\n", nNodes ); }
    }
    Sfm_NtkFree( p );
    return 1;
}



/**Function*************************************************************

  Synopsis    [Unrolls logic network while dropping some next-state functions.]

  Description [Returns the unrolled network.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkUnrollAndDrop(p *Abc_Ntk_t, nFrames int32, nFramesAdd int32, vFlops *Vec_Int_t, piPivot *int32) *Abc_Ntk_t {
    var pNtk *Abc_Ntk_t; 
    var pFanin, pNode *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i, k, f, Value int32;
    assert(nFramesAdd >= 0, "nFramesAdd >= 0");
    assert(Abc_NtkIsLogic(p), "Abc_NtkIsLogic(p)");
    assert(Vec_IntSize(vFlops) == Abc_NtkLatchNum(p), "Vec_IntSize(vFlops) == Abc_NtkLatchNum(p)");
    *piPivot = -1;
    // start the network
    pNtk = Abc_NtkAlloc( p.ntkType, p.ntkFunc, 1 );
    pNtk.pName = Extra_UtilStrsav(Abc_NtkName(p));
    // add CIs for the new network
    Abc_NtkForEachCi( p, pNode, i )
        pNode.pCopy = Abc_NtkCreatePi( pNtk );
    // iterate unrolling
    vNodes = Abc_NtkDfs( p, 0 );
    for  f = 0; f <= nFrames + nFramesAdd; f++  {
        if f > 0 {
            Abc_NtkForEachPi( p, pNode, i )
                pNode.pCopy = Abc_NtkCreatePi( pNtk );
        }
        Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pNode, i )
        {
            Abc_NtkDupObj( pNtk, pNode, 0 );
            Abc_ObjForEachFanin( pNode, pFanin, k )
                Abc_ObjAddFanin( pNode.pCopy, pFanin.pCopy );
        }
        Abc_NtkForEachCo( p, pNode, i )
            pNode.pCopy = Abc_ObjFanin0(pNode).pCopy;
        Abc_NtkForEachPo( p, pNode, i )
            Abc_ObjAddFanin( Abc_NtkCreatePo(pNtk), pNode.pCopy );
        // add buffers
        if f == 0 {
            *piPivot = Abc_NtkObjNum(pNtk);
//            Abc_NtkForEachLatchInput( p, pNode, i )
//                pNode->pCopy = Abc_NtkCreateNodeBuf( pNtk, pNode->pCopy );
        }
        // transfer to flop outputs
        Abc_NtkForEachLatch( p, pNode, i )
            Abc_ObjFanout0(pNode).pCopy = Abc_ObjFanin0(pNode).pCopy;
        // add final POs
        if f > nFramesAdd {
            Vec_IntForEachEntry( vFlops, Value, i )
            {
                if Value == 0 {
                    continue; }
                pNode = Abc_NtkCo( p, Abc_NtkPoNum(p) + i );
                Abc_ObjAddFanin( Abc_NtkCreatePo(pNtk), pNode.pCopy );
            }
        }
    }
    Vec_PtrFree( vNodes );
    Abc_NtkAddDummyPiNames( pNtk );
    Abc_NtkAddDummyPoNames( pNtk );
    // perform combinational cleanup
    Abc_NtkCleanup( pNtk, 0 );
    if !Abc_NtkCheck( pNtk ) {
        fmt.Fprintf(os.Stdout, "Abc_NtkCreateFromNode(): Network check has failed.\n"); }
    return pNtk;
}

/**Function*************************************************************

  Synopsis    [Updates the original network to include optimized nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkReinsertNodes(p *Abc_Ntk_t, pNtk *Abc_Ntk_t, iPivot int32) {
    var pNode, pNodeNew, pFaninNew *Abc_Obj_t;
    var vNodes *Vec_Ptr_t;
    var i, k int32;
    assert(Abc_NtkIsLogic(p), "Abc_NtkIsLogic(p)");
    assert(Abc_NtkCiNum(p) <= Abc_NtkCiNum(pNtk), "Abc_NtkCiNum(p) <= Abc_NtkCiNum(pNtk)");
    vNodes = Abc_NtkDfs( p, 0 );
    // clean old network
    Abc_NtkCleanCopy( p );
    Abc_NtkForEachNode( p, pNode, i )
    {
        Abc_ObjRemoveFanins( pNode );
        pNode.pData = Abc_SopRegister( (Mem_Flex_t *)p.pManFunc, (char *)" 0\n" );
    }
    // map CIs
    Abc_NtkForEachCi( p, pNode, i )
        Abc_NtkCi(pNtk, i).pCopy = pNode;
    // map internal nodes
    assert(Vec_PtrSize(vNodes) + Abc_NtkCiNum(p) + Abc_NtkPoNum(p) == iPivot, "Vec_PtrSize(vNodes) + Abc_NtkCiNum(p) + Abc_NtkPoNum(p) == iPivot");
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pNode, i )
    {
        pNodeNew = Abc_NtkObj( pNtk, Abc_NtkCiNum(p) + i + 1 );
        if pNodeNew == nil {
            continue; }
        pNodeNew.pCopy = pNode;
    }
    // connect internal nodes
    Vec_PtrForEachEntry( Abc_Obj_t *, vNodes, pNode, i )
    {
        pNodeNew = Abc_NtkObj( pNtk, Abc_NtkCiNum(p) + i + 1 );
        if pNodeNew == nil {
            continue; }
        assert(pNodeNew.pCopy == pNode, "pNodeNew.pCopy == pNode");
        Abc_ObjForEachFanin( pNodeNew, pFaninNew, k )
            Abc_ObjAddFanin( pNodeNew.pCopy, pFaninNew.pCopy );
        pNode.pData = Abc_SopRegister( (Mem_Flex_t *)p.pManFunc, (char *)pNodeNew.pData );
    }
    Vec_PtrFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Performs MFS for the unrolled network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsAfterICheck(p *Abc_Ntk_t, nFrames int32, nFramesAdd int32, vFlops *Vec_Int_t, pPars *Sfm_Par_t) int32 {
    var pp *Sfm_Ntk_t;
    var nFaninMax, nNodes int32;
    var pNtk *Abc_Ntk_t;
    var iPivot int32;
    assert(Abc_NtkIsLogic(p), "Abc_NtkIsLogic(p)");
    // count fanouts
    nFaninMax = Abc_NtkGetFaninMax( p );
    if nFaninMax > 15 {
        Abc_Print( 1, "Currently \"mfs\" cannot process the network containing nodes with more than 15 fanins.\n" );
        return 0;
    }
    if !Abc_NtkHasSop(p) {
        Abc_NtkToSop( p, -1, ABC_INFINITY ); }
    // derive unfolded network
    pNtk = Abc_NtkUnrollAndDrop( p, nFrames, nFramesAdd, vFlops, &iPivot );
    Io_WriteBlifLogic( pNtk, "unroll_dump.blif", 0 );
    // collect information
    pp = Abc_NtkExtractMfs2( pNtk, iPivot );
    // perform optimization
    nNodes = Sfm_NtkPerform( pp, pPars );
    if nNodes == 0 {
//        Abc_Print( 1, "The network is not changed by \"mfs\".\n" );
    
    } else {
        Abc_NtkInsertMfs( pNtk, pp );
        if pPars.fVerbose {
            Abc_Print( 1, "The network has %d nodes changed by \"mfs\".\n", nNodes ); }
        Abc_NtkReinsertNodes( p, pNtk, iPivot );
    }
    Abc_NtkDelete( pNtk );
    Sfm_NtkFree( pp );
    // perform final sweep
    Abc_NtkSweep( p, 0 );
    if !Abc_NtkHasSop(p) {
        Abc_NtkToSop( p, -1, ABC_INFINITY ); }
    return 1;

}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

