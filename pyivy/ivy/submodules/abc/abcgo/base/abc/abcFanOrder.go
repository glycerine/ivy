package abc

import "strings"

/**CFile****************************************************************

  FileName    [abcFanOrder.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Fanin ordering procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcFanOrder.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

  Synopsis    [Reorder fanins of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkOrderFaninsById(pNtk *Abc_Ntk_t) {
    var vOrder *Vec_Int_t;
    var vStore *Vec_Str_t;
    var pNode *Abc_Obj_t;
    var pSop, pSopNew string;
    var pCube, pCubeNew string;
    var nVars int32;
    var i int32;
    var v int32;
    var pOrder *int32;
    assert(Abc_NtkHasSop(pNtk), "Abc_NtkHasSop(pNtk)");
    vOrder = Vec_IntAlloc( 100 );
    vStore = Vec_StrAlloc( 100 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        pSop = (char *)pNode.pData;
        nVars = Abc_SopGetVarNum(pSop);
        assert(nVars == Abc_ObjFaninNum(pNode), "nVars == Abc_ObjFaninNum(pNode)");
        Vec_IntClear( vOrder );
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vOrder, v );
        }
        pOrder = Vec_IntArray(vOrder);
        Vec_IntSelectSortCost( pOrder, nVars, &pNode.vFanins );
        // copy the cover
        Vec_StrGrow( vStore, Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1 );
        memcpy( Vec_StrArray(vStore), pSop, (size_t)(Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1) );
        pSopNew = pCubeNew = pSop;
        pSop = Vec_StrArray(vStore);
        // generate permuted one
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
            for  v = 0; v < nVars; v++  {
                pCubeNew[v] = '-';
            }
            for  v = 0; v < nVars; v++  {
                if pCube[pOrder[v]] == '0' {
                    pCubeNew[v] = '0';
                } else if pCube[pOrder[v]] == '1' {
                    pCubeNew[v] = '1'; }
            }
            pCubeNew += nVars + 3;
        }
        pNode.pData = pSopNew;
        Vec_IntSort( &pNode.vFanins, 0 );
//        Vec_IntPrint( vOrder );
    }
    }
    Vec_IntFree( vOrder );
    Vec_StrFree( vStore );
}

/**Function*************************************************************

  Synopsis    [Returns fanin permutation to reorders columns lexicographically.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSopTranspose(pSop string, nVars int32, vCubes *Vec_Ptr_t, vStore *Vec_Str_t) {
    var pCube string; 
    var nCubes, v, c int32;
    // collect original cubes
    Vec_PtrClear( vCubes );
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        Vec_PtrPush( vCubes, pCube );
    }
    // rebuild the cubes
    Vec_StrClear( vStore );
    for  v = 0; v < nVars; v++  {
        // inlining of macro Vec_PtrForEachEntry
        for  c = 0; (c < Vec_PtrSize(vCubes)) && (((pCube) = (char *)Vec_PtrEntry(vCubes, c)), 1); c++  {
            Vec_StrPush( vStore, pCube[v] );
        }
        Vec_StrPush( vStore, '\0' );
    }
    // get the cubes
    nCubes = Vec_PtrSize( vCubes );
    Vec_PtrClear( vCubes );
    for  v = 0; v < nVars; v++  {
        Vec_PtrPush( vCubes, Vec_StrEntryP(vStore, v*(nCubes+1)) );
    }
}
func Vec_StrSelectSortCost(pArray *string, nSize int32, vPerm *Vec_Int_t) {
    var i int32;
    var j int32;
    var best_i int32;
    var pPerm *int32;
    Vec_IntClear( vPerm );
    for  i = 0; i < nSize; i++  {
        Vec_IntPush( vPerm, i );
    }
    pPerm = Vec_IntArray( vPerm );
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if int32(strings.Compare(pArray[j], pArray[best_i])) < 0 {
                best_i = j; }
        }
        // inlining of macro ABC_SWAP
        { var t string  = pArray[i]; pArray[i] = pArray[best_i]; pArray[best_i] = t; };
        // inlining of macro ABC_SWAP
        { var t int32  = pPerm[i]; pPerm[i] = pPerm[best_i]; pPerm[best_i] = t; };
    }
}
func Abc_NtkOrderFaninsBySortingColumns(pNtk *Abc_Ntk_t) {
    var vOrder *Vec_Int_t;
    var vCounts *Vec_Int_t;
    var vFanins *Vec_Int_t;
    var vStore *Vec_Str_t;
    var vCubes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var pSop, pSopNew string;
    var pCube, pCubeNew string;
    var nVars int32;
    var i int32;
    var v int32;
    var pOrder *int32;
    assert(Abc_NtkHasSop(pNtk), "Abc_NtkHasSop(pNtk)");
    vOrder = Vec_IntAlloc( 100 );
    vStore = Vec_StrAlloc( 100 );
    vCubes = Vec_PtrAlloc( 100 );
    vCounts = Vec_IntAlloc( 100 );
    vFanins = Vec_IntAlloc( 100 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        pSop = (char *)pNode.pData;
        nVars = Abc_SopGetVarNum(pSop);
        assert(nVars == Abc_ObjFaninNum(pNode), "nVars == Abc_ObjFaninNum(pNode)");
        // create a transposed SOP
        Abc_NtkSopTranspose( pSop, nVars, vCubes, vStore );
        // create permutation
        Vec_StrSelectSortCost( (char **)Vec_PtrArray(vCubes), nVars, vOrder );
        pOrder = Vec_IntArray(vOrder);
        // copy the cover
        Vec_StrGrow( vStore, Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1 );
        memcpy( Vec_StrArray(vStore), pSop, (size_t)(Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1) );
        pSopNew = pCubeNew = pSop;
        pSop = Vec_StrArray(vStore);
        // generate permuted one
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
            for  v = 0; v < nVars; v++  {
                pCubeNew[v] = '-';
            }
            for  v = 0; v < nVars; v++  {
                if pCube[pOrder[v]] == '0' {
                    pCubeNew[v] = '0';
                } else if pCube[pOrder[v]] == '1' {
                    pCubeNew[v] = '1'; }
            }
            pCubeNew += nVars + 3;
        }
        pNode.pData = pSopNew;
        // generate the fanin order
        Vec_IntClear( vFanins );
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vFanins, Abc_ObjFaninId( pNode, pOrder[v] ) );
        }
        Vec_IntClear( &pNode.vFanins );
        Vec_IntAppend( &pNode.vFanins, vFanins );
    }
    }
    Vec_IntFree( vFanins );
    Vec_IntFree( vCounts );
    Vec_IntFree( vOrder );
    Vec_StrFree( vStore );
    Vec_PtrFree( vCubes );
}


/**Function*************************************************************

  Synopsis    [Reorders columns by literal and then lexicographically.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSelectSortCost2(pArray *string, nSize int32, vCounts *Vec_Int_t, vPerm *Vec_Int_t) {
    var i int32;
    var j int32;
    var best_i int32;
    var pPerm *int32;
    Vec_IntClear( vPerm );
    for  i = 0; i < nSize; i++  {
        Vec_IntPush( vPerm, i );
    }
    pPerm = Vec_IntArray( vPerm );
    for  i = 0; i < nSize-1; i++  {
        best_i = i;
        for  j = i+1; j < nSize; j++  {
            if Vec_IntEntry(vCounts, pPerm[j]) <  Vec_IntEntry(vCounts, pPerm[best_i]) || 
                (Vec_IntEntry(vCounts, pPerm[j]) == Vec_IntEntry(vCounts, pPerm[best_i]) && int32(strings.Compare(pArray[j], pArray[best_i])) < 0) {
                best_i = j; }
        }
        // inlining of macro ABC_SWAP
        { var t string  = pArray[i]; pArray[i] = pArray[best_i]; pArray[best_i] = t; };
        // inlining of macro ABC_SWAP
        { var t int32  = pPerm[i]; pPerm[i] = pPerm[best_i]; pPerm[best_i] = t; };
    }
}
func Abc_NtkOrderFaninsByLitCount(pNtk *Abc_Ntk_t) {
    var vOrder *Vec_Int_t;
    var vCounts *Vec_Int_t;
    var vFanins *Vec_Int_t;
    var vStore *Vec_Str_t;
    var vCubes *Vec_Ptr_t;
    var pNode *Abc_Obj_t;
    var pSop, pSopNew string;
    var pCube, pCubeNew string;
    var nVars int32;
    var i int32;
    var v int32;
    var pOrder *int32;
    assert(Abc_NtkHasSop(pNtk), "Abc_NtkHasSop(pNtk)");
    vOrder = Vec_IntAlloc( 100 );
    vStore = Vec_StrAlloc( 100 );
    vCubes = Vec_PtrAlloc( 100 );
    vCounts = Vec_IntAlloc( 100 );
    vFanins = Vec_IntAlloc( 100 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        pSop = (char *)pNode.pData;
        nVars = Abc_SopGetVarNum(pSop);
        assert(nVars == Abc_ObjFaninNum(pNode), "nVars == Abc_ObjFaninNum(pNode)");
        // count literals
        Vec_IntFill( vCounts, nVars, 0 );
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
            for  v = 0; v < nVars; v++  {
                if pCube[v] != '-' {
                    Vec_IntAddToEntry( vCounts, v, 1 ); }
            }
        }

        // create a transposed SOP
        Abc_NtkSopTranspose( pSop, nVars, vCubes, vStore );
        // create permutation
        Vec_StrSelectSortCost2( (char **)Vec_PtrArray(vCubes), nVars, vCounts, vOrder );
        pOrder = Vec_IntArray(vOrder);
/*
        // find good order
        Vec_IntClear( vOrder );
        for ( v = 0; v < nVars; v++ )
            Vec_IntPush( vOrder, v );
        pOrder = Vec_IntArray(vOrder);
        Vec_IntSelectSortCost( pOrder, nVars, vCounts );
*/
        // copy the cover
        Vec_StrGrow( vStore, Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1 );
        memcpy( Vec_StrArray(vStore), pSop, (size_t)(Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1) );
        pSopNew = pCubeNew = pSop;
        pSop = Vec_StrArray(vStore);
        // generate permuted one
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
            for  v = 0; v < nVars; v++  {
                pCubeNew[v] = '-';
            }
            for  v = 0; v < nVars; v++  {
                if pCube[pOrder[v]] == '0' {
                    pCubeNew[v] = '0';
                } else if pCube[pOrder[v]] == '1' {
                    pCubeNew[v] = '1'; }
            }
            pCubeNew += nVars + 3;
        }
        pNode.pData = pSopNew;
        // generate the fanin order
        Vec_IntClear( vFanins );
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vFanins, Abc_ObjFaninId( pNode, pOrder[v] ) );
        }
        Vec_IntClear( &pNode.vFanins );
        Vec_IntAppend( &pNode.vFanins, vFanins );
    }
    }
    Vec_IntFree( vFanins );
    Vec_IntFree( vCounts );
    Vec_IntFree( vOrder );
    Vec_StrFree( vStore );
    Vec_PtrFree( vCubes );
}
func Abc_NtkOrderFaninsByLitCountAndCubeCount(pNtk *Abc_Ntk_t) {
    // assuming that the fanins are sorted by the number of literals in each cube
    // this procedure sorts the literals appearing only once by the number of their cube
    var vOrder *Vec_Int_t;
    var vCounts *Vec_Int_t;
    var vFanins *Vec_Int_t;
    var vCubeNum *Vec_Int_t;
    var vStore *Vec_Str_t;
    var pNode *Abc_Obj_t;
    var pSop, pSopNew string;
    var pCube, pCubeNew string;
    var nVars int32;
    var i int32;
    var v int32;
    var iCube int32;
    var pOrder *int32;
    assert(Abc_NtkHasSop(pNtk), "Abc_NtkHasSop(pNtk)");
    vStore = Vec_StrAlloc( 100 );
    vOrder = Vec_IntAlloc( 100 );
    vCounts = Vec_IntAlloc( 100 );
    vFanins = Vec_IntAlloc( 100 );
    vCubeNum = Vec_IntAlloc( 100 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        pSop = (char *)pNode.pData;
        nVars = Abc_SopGetVarNum(pSop);
        assert(nVars == Abc_ObjFaninNum(pNode), "nVars == Abc_ObjFaninNum(pNode)");
        // count literals and remember the cube where each literal appears
        Vec_IntFill( vCounts, nVars, 0 );
        Vec_IntFill( vCubeNum, nVars, 0 );
        iCube = 0;
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
            for  v = 0; v < nVars; v++  {
                if pCube[v] != '-' {
                    Vec_IntAddToEntry( vCounts, v, 1 );
                    Vec_IntWriteEntry( vCubeNum, v, iCube );
                }
            }
            iCube++;
        }
        // create new order
        for  v = 0; v < nVars; v++  {
            if Vec_IntEntry(vCounts, v) == 1 {
                Vec_IntWriteEntry( vCounts, v, Vec_IntEntry(vCubeNum, v) );
            } else {
                Vec_IntWriteEntry( vCounts, v, ABC_INFINITY ); }
        }
        // find good order
        Vec_IntClear( vOrder );
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vOrder, v );
        }
        pOrder = Vec_IntArray(vOrder);
        Vec_IntSelectSortCost( pOrder, nVars, vCounts );
        // copy the cover
        Vec_StrGrow( vStore, Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1 );
        memcpy( Vec_StrArray(vStore), pSop, (size_t)(Abc_SopGetCubeNum(pSop) * (nVars + 3) + 1) );
        pSopNew = pCubeNew = pSop;
        pSop = Vec_StrArray(vStore);
        // generate permuted one
        // inlining of macro Abc_SopForEachCube
        for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
            for  v = 0; v < nVars; v++  {
                pCubeNew[v] = '-';
            }
            for  v = 0; v < nVars; v++  {
                if pCube[pOrder[v]] == '0' {
                    pCubeNew[v] = '0';
                } else if pCube[pOrder[v]] == '1' {
                    pCubeNew[v] = '1'; }
            }
            pCubeNew += nVars + 3;
        }
        pNode.pData = pSopNew;
        // generate the fanin order
        Vec_IntClear( vFanins );
        for  v = 0; v < nVars; v++  {
            Vec_IntPush( vFanins, Abc_ObjFaninId( pNode, pOrder[v] ) );
        }
        Vec_IntClear( &pNode.vFanins );
        Vec_IntAppend( &pNode.vFanins, vFanins );
    }
    }
    Vec_IntFree( vCubeNum );
    Vec_IntFree( vFanins );
    Vec_IntFree( vCounts );
    Vec_IntFree( vOrder );
    Vec_StrFree( vStore );
}

/**Function*************************************************************

  Synopsis    [Split large nodes by dividing their SOPs in half.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSplitLarge(pNode *Abc_Obj_t) {
    var pNode1, pNode2, pFanin *Abc_Obj_t;
    var CutPoint int32;
    var nVars int32 = Abc_ObjFaninNum(pNode);
    var i int32;
    var nCubes int32 = Abc_SopGetCubeNum((char *)pNode.pData);
    pNode1 = Abc_NtkDupObj( pNode.pNtk, pNode, 0 );
    pNode2 = Abc_NtkDupObj( pNode.pNtk, pNode, 0 );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_ObjAddFanin( pNode1, pFanin );
    }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_ObjAddFanin( pNode2, pFanin );
    }    
    // update the node
    Abc_ObjRemoveFanins( pNode );
    Abc_ObjAddFanin( pNode, pNode1 );
    Abc_ObjAddFanin( pNode, pNode2 );
    pNode.pData = Abc_SopCreateOr( (Mem_Flex_t *)pNode.pNtk.pManFunc, 2, nil );
    // update covers of the nodes
    assert(nCubes > 1, "nCubes > 1");
    CutPoint = (nCubes / 2) * (nVars + 3);
    ((char *)pNode1.pData)[CutPoint] = 0;
    pNode2.pData = (char *)pNode2.pData + CutPoint;
}
func Abc_NtkSplitLarge(pNtk *Abc_Ntk_t, nFaninsMax int32, nCubesMax int32) {
    var pNode *Abc_Obj_t;
    var nObjOld int32  = Abc_NtkObjNumMax(pNtk);
    var i, nCubes int32;
    assert(Abc_NtkHasSop(pNtk), "Abc_NtkHasSop(pNtk)");
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        if i == nObjOld {
            break; }
        nCubes = Abc_SopGetCubeNum((char *)pNode.pData);
        if (Abc_ObjFaninNum(pNode) > nFaninsMax && nCubes > 1) || nCubes > nCubesMax {
            Abc_NodeSplitLarge( pNode ); }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Sorts the cubes in a topological order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeCompareCubes1(pp1 *string, pp2 *string) int32 {
    return int32(strings.Compare(*pp1, *pp2));
}
func Abc_NodeCompareCubes2(pp1 *string, pp2 *string) int32 {
    var pStr1 string  = *pp1;
    var pStr2 string  = *pp2;
    var i, nNum1, nNum2 int32;
    for  i = 0; pStr1[i]; i++  {
        nNum1 += (pStr1[i] != '-');
        nNum2 += (pStr2[i] != '-');
    }
    if nNum1 > nNum2 {
        return -1; }
    if nNum1 < nNum2 {
        return 1; }
    return int32(strings.Compare(*pp1, *pp2));
}
func Abc_NodeSortCubes(pNode *Abc_Obj_t, vCubes *Vec_Ptr_t, vStore *Vec_Str_t, fWeight int32) {
    var pCube, pPivot string;
    var pSop string  = (char *)pNode.pData;
    var i int32;
    var nVars int32 = Abc_ObjFaninNum(pNode);
    Vec_PtrClear( vCubes );
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        assert(pCube[nVars] == ' ', "pCube[nVars] == ' '");
        pCube[nVars] = 0;
        Vec_PtrPush( vCubes, pCube );
    }
    if fWeight != 0 {
        Vec_PtrSort( vCubes, (int (*)())Abc_NodeCompareCubes2 );
    } else {
        Vec_PtrSort( vCubes, (int (*)())Abc_NodeCompareCubes1 ); }
    Vec_StrGrow( vStore, Vec_PtrSize(vCubes) * (nVars + 3) );
    pPivot = Vec_StrArray( vStore );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vCubes)) && (((pCube) = (char *)Vec_PtrEntry(vCubes, i)), 1); i++  {
        assert(pCube[nVars] == 0, "pCube[nVars] == 0");
        pCube[nVars] = ' ';
        memcpy( pPivot, pCube, (size_t)(nVars + 3) );
        pPivot += nVars + 3;
    }
    memcpy( pSop, Vec_StrArray(vStore), (size_t)(Vec_PtrSize(vCubes) * (nVars + 3)) );
}
func Abc_NtkSortCubes(pNtk *Abc_Ntk_t, fWeight int32) {
    var vCubes *Vec_Ptr_t;
    var vStore *Vec_Str_t;
    var pNode *Abc_Obj_t;
    var i int32;
    assert(Abc_NtkHasSop(pNtk), "Abc_NtkHasSop(pNtk)");
    vCubes = Vec_PtrAlloc( 1000 );
    vStore = Vec_StrAlloc( 1000 );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Abc_NodeSortCubes( pNode, vCubes, vStore, fWeight ); }
    }
    Vec_StrFree( vStore );
    Vec_PtrFree( vCubes );
}


/**Function*************************************************************

  Synopsis    [Sorts fanins of each node to make SOPs more readable.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSortSops(pNtk *Abc_Ntk_t) {
    Abc_NtkSortCubes( pNtk, 1 );
    Abc_NtkOrderFaninsByLitCount( pNtk );
    Abc_NtkSortCubes( pNtk, 0 );
    Abc_NtkOrderFaninsByLitCountAndCubeCount( pNtk );
    Abc_NtkSortCubes( pNtk, 0 );
}

/**Function*************************************************************

  Synopsis    [Makes cover legitimate for "fast_extract".]

  Description [Iteratively removes distance-1 and contained cubes.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_CubeContain(pCube1 string, pCube2 string, nVars int32) int32 {
    var v int32;
    var fCont12 int32 = 1;
    var fCont21 int32 = 1;
    for  v = 0; v < nVars; v++  {
        if pCube1[v] == pCube2[v] {
            continue; }
        if pCube1[v] == '-' {
            fCont21 = 0;
        } else if pCube2[v] == '-' {
            fCont12 = 0;
        } else {
            return 0; }
        if !(fCont12 != 0) && !(fCont21 != 0) {
            return 0; }
    }
    assert((fCont21 != 0) || (fCont12 != 0), "fCont21 || fCont12");
    return (fCont21 << 1) | fCont12;
}
func Abc_NodeMakeSCCFree(pNode *Abc_Obj_t) int32 {
    var pSop string  = (char *)pNode.pData;
    var pCube, pCube2, pSopNew string;
    var nVars int32  = Abc_ObjFaninNum(pNode);
    var Status, nCount int32;
    // inlining of macro Abc_SopForEachCubePair
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {                                                         for  pCube2 = (pCube + (nVars) + 3); *pCube2; pCube2 += (nVars) + 3  {
        if pCube[0] == 'z' || pCube2[0] == 'z' {
            continue; }
        Status = Abc_CubeContain( pCube, pCube2, nVars );
        nCount += int32((Status > 0));
        if Status & 1 {
            pCube2[0] = 'z';
        } else if Status & 2 {
            pCube[0] = 'z'; }
    }
    }
    if nCount == 0 {
        return 0; }
    // create new cover
    pSopNew = (char *)pNode.pData;
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
        if pCube[0] == 'z' {
            continue; }
        memcpy( pSopNew, pCube, (size_t)(nVars + 3) );
        pSopNew += nVars + 3;
    }
    *pSopNew = 0;
    return 1;
}
func Abc_NodeMakeDist1Free(pNode *Abc_Obj_t) {
    var pSop string  = (char *)pNode.pData;
    var pCube, pCube2 string;
    var i int32;
    var nVars int32 = Abc_ObjFaninNum(pNode);
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
    // inlining of macro Abc_SopForEachCube
    for  pCube2 = (pCube + nVars + 3); *pCube2; pCube2 += (nVars) + 3  {
        var Counter int32;
        var iDiff int32 = -1;
        for  i = 0; i < nVars; i++  {
            if pCube[i] != pCube2[i] {
                Counter++, iDiff = i; }
        }
        if Counter == 1 && ((pCube[iDiff] == '0' && pCube2[iDiff] == '1') || (pCube[iDiff] == '1' && pCube2[iDiff] == '0')) {
            pCube[iDiff] = pCube2[iDiff] = '-'; }
    }
    }
}
func Abc_NodeCheckDist1Free(pNode *Abc_Obj_t) {
    var pSop string  = (char *)pNode.pData;
    var pCube, pCube2 string;
    var i int32;
    var nVars int32 = Abc_ObjFaninNum(pNode);
    // inlining of macro Abc_SopForEachCube
    for  pCube = (pSop); *pCube; pCube += (nVars) + 3  {
    // inlining of macro Abc_SopForEachCube
    for  pCube2 = (pSop); *pCube2; pCube2 += (nVars) + 3  {
        var Counter int32;
        if pCube == pCube2 {
            continue; }
        for  i = 0; i < nVars; i++  {
            if pCube[i] != pCube2[i] {
                Counter++; }
        }
        assert(Counter > 1, "Counter > 1");
    }
    }
}
func Abc_NodeMakeLegit(pNode *Abc_Obj_t) int32 {
    var i int32;
    var fChanges int32 = 1;
    for  i = 0; (fChanges != 0); i++  {
        Abc_NodeMakeDist1Free( pNode );
        fChanges = Abc_NodeMakeSCCFree( pNode );
    }
//    Abc_NodeCheckDist1Free( pNode );
    return i > 1;
}
func Abc_NtkMakeLegit(pNtk *Abc_Ntk_t) int32 {
    var pNode *Abc_Obj_t;
    var i, Counter int32;
    assert(Abc_NtkHasSop(pNtk), "Abc_NtkHasSop(pNtk)");
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Counter += Abc_NodeMakeLegit( pNode ); }
    }
    if Counter != 0 {
        Abc_Print( 0, "%d nodes were made dist1-cube-free and/or single-cube-containment-free.\n", Counter ); }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

