package sfm

import "fmt"

/**CFile****************************************************************

  FileName    [sfmArea.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Area optimization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmArea.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"
//#include "map/mio/mio.h"
//#include "misc/util/utilTruth.h"
//#include "misc/util/utilNam.h"
//#include "map/scl/sclLib.h"
//#include "map/scl/sclCon.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Precompute cell parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkPrecomputeCellPairs(pCells *Mio_Cell2_t, nCells int32) *Vec_Int_t {
    var vInfo *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var iBestArea, tCur, iThis word;
    var pPerm [7]*int32;
    var nPerms [7]int32;
    var Perm [7]int32;
    var Perm1 *int32;
    var Perm2 *int32;
    var iBestCell, iBestPerm, iBestDiff int32;
    var i, k, n, v, p, Count int32;
    var iGate1 int32 = -1;
    var iGate2 int32 = -1;

    for  i = 1; i <= 6; i++  {
        pPerm[i] = Extra_PermSchedule( i );
    }
    for  i = 1; i <= 6; i++  {
        nPerms[i] = Extra_Factorial( i );
    }

    for  i = 2; i < nCells; i++  {
        var nFanins int32  = pCells[i].nFanins;
        for  n = 0; n <= nFanins; n++  {
            // get the truth table
            if n == nFanins {
                iThis = ~pCells[i].uTruth;
            } else {
                iThis = Abc_Tt6Flip(pCells[i].uTruth, n);
            }
            // init the comparison
            iBestArea = ~((word)0);
            iBestCell = iBestPerm = iBestDiff = -1;
            // iterate through cells
            for  k = 2; k < nCells; k++  {
                if nFanins != int32(pCells[k].nFanins) {
                    continue; }
                if i != k && pCells[i].uTruth == pCells[k].uTruth {
                    iGate1 = i;
                    iGate2 = k;
                    Count++;
                    continue;
                }
                // set unit permutation
                for  v = 0; v < nFanins; v++  {
                    Perm[v] = v;
                }
                // go through all permutation of Cell[k]
                tCur = pCells[k].uTruth;
                for  p = 0; p < nPerms[nFanins]; p++  {
                    if iThis == tCur && iBestArea > pCells[k].AreaW {
                        iBestArea = pCells[k].AreaW;
                        iBestCell = k;
                        iBestPerm = 0;
                        for  v = 0; v < nFanins; v++  {
                            iBestPerm |= (v << (Perm[v] << 2));
                        }
                        if pCells[i].AreaW >= pCells[k].AreaW {
                            iBestDiff = int32((pCells[i].AreaW - pCells[k].AreaW));
                        } else {
                            iBestDiff = -int32((pCells[k].AreaW - pCells[i].AreaW));
                        }
                    }
                    if nPerms[nFanins] == 1 {
                        continue; }
                    // update
                    tCur  = Abc_Tt6SwapAdjacent( tCur, pPerm[nFanins][p] );
                    Perm1 = Perm + pPerm[nFanins][p];
                    Perm2 = Perm1 + 1;
                    // inlining of macro ABC_SWAP
                    { var t int32  = *Perm1; *Perm1 = *Perm2; *Perm2 = t; };
                }
                assert(tCur == pCells[k].uTruth, "tCur == pCells[k].uTruth");
            }
            Vec_IntPushThree( vInfo, iBestCell, iBestPerm, iBestDiff );
        }
    }

    for  i = 1; i <= 6; i++  {
        // inlining of macro ABC_FREE
        ((pPerm[i]) ? (free((char *) (pPerm[i])), (pPerm[i]) = 0) : 0);
    }
    if Count != 0 {
        fmt.Printf("In this library, %d cell pairs have equal functions (for example, %s and %s).\n", Count/2, pCells[iGate1].pName, pCells[iGate2].pName); }
    return vInfo;
}
func Abc_NtkPrecomputeFirsts(pCells *Mio_Cell2_t, nCells int32) *Vec_Int_t {
    var i, Index int32;
    var vFirst *Vec_Int_t  = Vec_IntStartFull( 2 ); 
    for  i = 2; i < nCells; i++  {
        Vec_IntPush( vFirst, Index );
        Index += 3 * (pCells[i].nFanins + 1);
    }
    assert(nCells == Vec_IntSize(vFirst), "nCells == Vec_IntSize(vFirst)");
    return vFirst;
}
func Abc_NtkPrecomputePrint(pCells *Mio_Cell2_t, nCells int32, vInfo *Vec_Int_t) int32 {
    var i, n, v, Index, nRecUsed int32;
    for  i = 2; i < nCells; i++  {
        var nFanins int32  = pCells[i].nFanins;
        fmt.Printf("%3d : %8s   Fanins = %d   ", i, pCells[i].pName, nFanins);
        Dau_DsdPrintFromTruth( &pCells[i].uTruth, nFanins );
        for  n = 0; n <= nFanins; n++, Index += 3  {
            var iCellA int32  = Vec_IntEntry( vInfo, Index+0 );
            var iPerm int32   = Vec_IntEntry( vInfo, Index+1 );
            var Diff int32    = Vec_IntEntry( vInfo, Index+2 );
            if iCellA == -1 {
                continue; }
            fmt.Printf("%d : {", n);
            for  v = 0; v < nFanins; v++  {
                fmt.Printf(" %d ", (iPerm >> (v << 2)) & 15);
            }
            fmt.Printf("}  Index = %d  ", Index);

            fmt.Printf("Gain = %6.2f  ", Scl_Int2Flt(Diff));
            Dau_DsdPrintFromTruth( &pCells[iCellA].uTruth, pCells[iCellA].nFanins );
            nRecUsed++;
        }
    }
    return nRecUsed;
}

func Abc_NtkPrecomputeCellPairsTest() {
    var nCells int32;
    var pCells *Mio_Cell2_t  = Mio_CollectRootsNewDefault2( 6, &nCells, 0 );
    var vInfo *Vec_Int_t  = Abc_NtkPrecomputeCellPairs( pCells, nCells );
    var nRecUsed int32  = Abc_NtkPrecomputePrint( pCells, nCells, vInfo );
    // iterate through the cells
    var vFirst *Vec_Int_t  = Abc_NtkPrecomputeFirsts( pCells, nCells );
    fmt.Printf("Used records = %d.  All records = %d.\n", nRecUsed, Vec_IntSize(vInfo)/3 - nRecUsed);
    assert(nCells == Vec_IntSize(vFirst), "nCells == Vec_IntSize(vFirst)");
    Vec_IntFree( vFirst );
    Vec_IntFree( vInfo );
    // inlining of macro ABC_FREE
    ((pCells) ? (free((char *) (pCells)), (pCells) = 0) : 0);
}

func Abc_NodeCheckFanoutHasFanin(pNode *Abc_Obj_t, pFanin *Abc_Obj_t) int32 {
    var pThis *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pThis) = Abc_ObjFanout(pNode, i)), 1); i++  {
        if Abc_NodeFindFanin(pThis, pFanin) >= 0 {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Evaluate changes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjHasDupFanins(pObj *Abc_Obj_t) int32 {
    var pArray *int32  = pObj.vFanins.pArray;
    var i int32;
    var k int32;
    var Limit int32 = Abc_ObjFaninNum(pObj);
    for  i = 0; i < Limit; i++  {
        for  k = i+1; k < Limit; k++  {
            if pArray[i] == pArray[k] {
                return 1; }
        }
    }
    return 0;
}
func Abc_ObjHasDupFanouts(pObj *Abc_Obj_t) int32 {
    var pArray *int32  = pObj.vFanouts.pArray;
    var i int32;
    var k int32;
    var Limit int32 = Abc_ObjFanoutNum(pObj);
    for  i = 0; i < Limit; i++  {
        for  k = i+1; k < Limit; k++  {
            if pArray[i] == pArray[k] {
                return 1; }
        }
    }
    return 0;
}
func Abc_ObjChangeEval(pObj *Abc_Obj_t, vInfo *Vec_Int_t, vFirst *Vec_Int_t, InvArea int32, pfUseInv *int32) int32 {
    var pNext *Abc_Obj_t;
    //Mio_Gate_t * pGate = (Mio_Gate_t *)pObj->pData;
    var iFanCell int32;
    var iNodeCell int32 = Mio_GateReadCell( (Mio_Gate_t *)pObj.pData );
    var pFanInfo *int32;
    var pNodeInfo *int32 = Vec_IntEntryP( vInfo, Vec_IntEntry(vFirst, iNodeCell) );
    var i int32;
    var fNeedInv int32;
    var Gain int32;
    var iFanin int32 = Abc_ObjFaninNum(pObj);
    var fUseInv int32 = Abc_NodeIsInv(pObj);
    assert(iFanin > 0, "iFanin > 0");
    *pfUseInv = 0;
    if pNodeInfo[3*iFanin] == -1 {
        return 0; }
    if fUseInv != 0 {
        Gain = InvArea;
    } else {
        Gain = pNodeInfo[3*iFanin+2]; }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pNext) = Abc_ObjFanout(pObj, i)), 1); i++  {
        if (fUseInv != 0) && Abc_NodeFindFanin(pNext, Abc_ObjFanin0(pObj)) >= 0 {
            return 0; }
        if Abc_ObjHasDupFanins(pNext) != 0 {
            return 0; }
        if !Abc_ObjIsNode(pNext) || Abc_NodeIsBuf(pNext) {
            fNeedInv = 1;
            continue;
        }
        if Abc_NodeIsInv(pNext) {
            if Abc_NodeCheckFanoutHasFanin(pNext, pObj) >= 0 {
                return 0; }
            Gain += InvArea;
            continue;
        }
        iFanCell = Mio_GateReadCell( (Mio_Gate_t *)pNext.pData );
        pFanInfo = Vec_IntEntryP( vInfo, Vec_IntEntry(vFirst, iFanCell) );
        iFanin   = Abc_NodeFindFanin( pNext, pObj );
        if pFanInfo[3*iFanin] == -1 {
            fNeedInv = 1;
            continue;
        }
        Gain += pFanInfo[3*iFanin+2];
    }
    if fNeedInv != 0 {
        Gain -= InvArea; }
    *pfUseInv = fNeedInv;
    return Gain;
}
func Abc_ObjChangeUpdate(pObj *Abc_Obj_t, iFanin int32, pCells *Mio_Cell2_t, pNodeInfo *int32, vTemp *Vec_Int_t) {
    var v int32;
    var Perm int32;
    var iNodeCell int32 = pNodeInfo[3*iFanin];
    //Mio_Gate_t * pGate = (Mio_Gate_t *)pObj->pData;
    //Abc_ObjPrint( stdout, pObj );
    //printf( "Replacing fanout %d with %s by %s with fanin %d.\n", Abc_ObjId(pObj), Mio_GateReadName(pGate), Mio_GateReadName((Mio_Gate_t *)pCells[iNodeCell].pMioGate), iFanin );
    pObj.pData = (Mio_Gate_t *)pCells[iNodeCell].pMioGate;
    Perm = pNodeInfo[3*iFanin+1];
    Vec_IntClear( vTemp );
    for  v = 0; v < Abc_ObjFaninNum(pObj); v++  {
        Vec_IntPush( vTemp, Abc_ObjFaninId(pObj, (Perm >> (v << 2)) & 15) );
    }
    Vec_IntClear( &pObj.vFanins );
    Vec_IntAppend( &pObj.vFanins, vTemp );
}
func Abc_ObjChangePerform(pObj *Abc_Obj_t, vInfo *Vec_Int_t, vFirst *Vec_Int_t, fUseInv int32, vTemp *Vec_Int_t, vFanout *Vec_Ptr_t, vFanout2 *Vec_Ptr_t, pCells *Mio_Cell2_t) {
    var pNext *Abc_Obj_t;
    var pNext2 *Abc_Obj_t;
    var pNodeInv *Abc_Obj_t = nil;
    var iFanCell int32;
    var iNodeCell int32 = Mio_GateReadCell( (Mio_Gate_t *)pObj.pData );
    var pFanInfo *int32;
    var pNodeInfo *int32 = Vec_IntEntryP( vInfo, Vec_IntEntry(vFirst, iNodeCell) );
    var i int32;
    var k int32;
    var iFanin int32 = Abc_ObjFaninNum(pObj);
    assert(iFanin > 0 && pNodeInfo[3*iFanin] != -1, "iFanin > 0 && pNodeInfo[3*iFanin] != -1");
    // update the node
    Abc_NodeCollectFanouts( pObj, vFanout );
    if Abc_NodeIsInv(pObj) {
        var pFanin *Abc_Obj_t  = Abc_ObjFanin0(pObj);
        // inlining of macro Vec_PtrForEachEntry
        for  k = 0; (k < Vec_PtrSize(vFanout)) && (((pNext) = (Abc_Obj_t *)Vec_PtrEntry(vFanout, k)), 1); k++  {
            Abc_ObjPatchFanin( pNext, pObj, pFanin );
        }
        assert(Abc_ObjFanoutNum(pObj) == 0, "Abc_ObjFanoutNum(pObj) == 0");
        Abc_NtkDeleteObj(pObj);
        pObj = pFanin;
//        assert( fUseInv == 0 );
    
    } else {
        Abc_ObjChangeUpdate( pObj, iFanin, pCells, pNodeInfo, vTemp ); }
    // add inverter if needed
    if fUseInv != 0 {
        pNodeInv = Abc_NtkCreateNodeInv(pObj.pNtk, pObj); }
    // update the fanouts
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFanout)) && (((pNext) = (Abc_Obj_t *)Vec_PtrEntry(vFanout, i)), 1); i++  {
        if !Abc_ObjIsNode(pNext) || Abc_NodeIsBuf(pNext) {
            Abc_ObjPatchFanin( pNext, pObj, pNodeInv );
            continue;
        }
        if Abc_NodeIsInv(pNext) {
            Abc_NodeCollectFanouts( pNext, vFanout2 );
            // inlining of macro Vec_PtrForEachEntry
            for  k = 0; (k < Vec_PtrSize(vFanout2)) && (((pNext2) = (Abc_Obj_t *)Vec_PtrEntry(vFanout2, k)), 1); k++  {
                Abc_ObjPatchFanin( pNext2, pNext, pObj );
            }
            assert(Abc_ObjFanoutNum(pNext) == 0, "Abc_ObjFanoutNum(pNext) == 0");
            Abc_NtkDeleteObj(pNext);
            continue;
        }
        iFanin   = Abc_NodeFindFanin( pNext, pObj );
        iFanCell = Mio_GateReadCell( (Mio_Gate_t *)pNext.pData );
        pFanInfo = Vec_IntEntryP( vInfo, Vec_IntEntry(vFirst, iFanCell) );
        if pFanInfo[3*iFanin] == -1 {
            Abc_ObjPatchFanin( pNext, pObj, pNodeInv );
            continue;
        }
        Abc_ObjChangeUpdate( pNext, iFanin, pCells, pFanInfo, vTemp );
    }
}
func Abc_NtkChangePerform(pNtk *Abc_Ntk_t, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var i, fNeedInv, nCells, Gain, GainAll, Count, CountInv int32;
    var pCells *Mio_Cell2_t  = Mio_CollectRootsNewDefault2( 6, &nCells, 0 );
    var vInfo *Vec_Int_t  = Abc_NtkPrecomputeCellPairs( pCells, nCells );
    var vFirst *Vec_Int_t  = Abc_NtkPrecomputeFirsts( pCells, nCells );

    var vFanout *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var vFanout2 *Vec_Ptr_t  = Vec_PtrAlloc( 100 );
    var vTemp *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pObj *Abc_Obj_t;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if Abc_ObjFaninNum(pObj) < 2 && !Abc_NodeIsInv(pObj) {
            continue; }
        if Abc_ObjHasDupFanouts(pObj) != 0 {
            continue; }
        Gain = Abc_ObjChangeEval( pObj, vInfo, vFirst, int32(pCells[3].AreaW), &fNeedInv );
        if Gain <= 0 {
            continue; }
        //printf( "Obj %d\n", Abc_ObjId(pObj) );
        Count++;
        CountInv += Abc_NodeIsInv(pObj);
        GainAll += Gain;
        Abc_ObjChangePerform( pObj, vInfo, vFirst, fNeedInv, vTemp, vFanout, vFanout2, pCells );
    }
    }
    Vec_PtrFree( vFanout2 );
    Vec_PtrFree( vFanout );
    Vec_IntFree( vTemp );

    Vec_IntFree( vFirst );
    Vec_IntFree( vInfo );
    // inlining of macro ABC_FREE
    ((pCells) ? (free((char *) (pCells)), (pCells) = 0) : 0);

    if fVerbose != 0 {
        fmt.Printf("Total gain in area = %6.2f after %d changes (including %d inverters). ", Scl_Int2Flt(GainAll), Count, CountInv); }
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

