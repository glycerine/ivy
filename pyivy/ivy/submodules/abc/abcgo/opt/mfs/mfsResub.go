package mfs

import "fmt"

/**CFile****************************************************************

  FileName    [mfsResub.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Procedures to perform resubstitution.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfsResub.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mfsInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Updates the network after resubstitution.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsUpdateNetwork(p *Mfs_Man_t, pObj *Abc_Obj_t, vMfsFanins *Vec_Ptr_t, pFunc *Hop_Obj_t) {
    var pObjNew, pFanin *Abc_Obj_t;
    var k int32;
    // create the new node
    pObjNew = Abc_NtkCreateNode( pObj.pNtk );
    pObjNew.pData = pFunc;
    // inlining of macro Vec_PtrForEachEntry
    for  k = 0; (k < Vec_PtrSize(vMfsFanins)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(vMfsFanins, k)), 1); k++  {
        Abc_ObjAddFanin( pObjNew, pFanin );
    }
    // replace the old node by the new node
//printf( "Replacing node " ); Abc_ObjPrint( stdout, pObj );
//printf( "Inserting node " ); Abc_ObjPrint( stdout, pObjNew );
    // update the level of the node
    Abc_NtkUpdate( pObj, pObjNew, p.vLevels );
}

/**Function*************************************************************

  Synopsis    [Prints resub candidate stats.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsPrintResubStats(p *Mfs_Man_t) {
    var pFanin, pNode *Abc_Obj_t;
    var i, k, nAreaCrits, nAreaExpanse int32;
    var nFaninMax int32  = Abc_NtkGetFaninMax(p.pNtk);
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pNode) = Abc_NtkObj(p.pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, k)), 1); k++  {
            if !Abc_ObjIsCi(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 {
                nAreaCrits++;
                nAreaExpanse += int32((Abc_ObjFaninNum(pNode) < nFaninMax));
            }
        } }
    }
//    printf( "Total area-critical fanins = %d. Belonging to expandable nodes = %d.\n", 
//        nAreaCrits, nAreaExpanse );
}

/**Function*************************************************************

  Synopsis    [Tries resubstitution.]

  Description [Returns 1 if it is feasible, or 0 if c-ex is found.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsTryResubOnce(p *Mfs_Man_t, pCands *int32, nCands int32) int32 {
    var fVeryVerbose int32;
    var pData *uint32;
    var RetValue int32;
    var RetValue2 int32 = -1;
    var iVar int32;
    var i int32;//, clk = Abc_Clock();
/*
    if ( p->pPars->fGiaSat )
    {
        RetValue2 = Abc_NtkMfsTryResubOnceGia( p, pCands, nCands );
p->timeGia += Abc_Clock() - clk;
        return RetValue2;
    }
*/ 
    p.nSatCalls++;
    RetValue = sat_solver_solve( p.pSat, pCands, pCands + nCands, (ABC_INT64_T)p.pPars.nBTLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
//    assert( RetValue == l_False || RetValue == l_True );

    if RetValue != l_Undef && RetValue2 != -1 {
        assert((RetValue == l_False) == (RetValue2 == 1), "(RetValue == l_False) == (RetValue2 == 1)");
    }

    if RetValue == l_False {
        if fVeryVerbose != 0 {
        fmt.Printf("U "); }
        return 1;
    }
    if RetValue != l_True {
        if fVeryVerbose != 0 {
        fmt.Printf("T "); }
        p.nTimeOuts++;
        return -1;
    }
    if fVeryVerbose != 0 {
    fmt.Printf("S "); }
    p.nSatCexes++;
    // store the counter-example
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vProjVarsSat)) && (((iVar) = Vec_IntEntry(p.vProjVarsSat, i)), 1); i++  {
        pData = (unsigned *)Vec_PtrEntry( p.vDivCexes, i );
        if !sat_solver_var_value( p.pSat, iVar ) { // remove 0s!!!
        
            assert(Abc_InfoHasBit(pData, p.nCexes), "Abc_InfoHasBit(pData, p.nCexes)");
            Abc_InfoXorBit( pData, p.nCexes );
        }
    }
    p.nCexes++;
    return 0;

}

/**Function*************************************************************

  Synopsis    [Performs resubstitution for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsSolveSatResub(p *Mfs_Man_t, pNode *Abc_Obj_t, iFanin int32, fOnlyRemove int32, fSkipUpdate int32) int32 {
    var fVeryVerbose int32;//p->pPars->fVeryVerbose && Vec_PtrSize(p->vDivs) < 200;// || pNode->Id == 556;
    var pData *uint32;
    var pCands [MFS_FANIN_MAX]int32;
    var RetValue, iVar, i, nCands, nWords, w int32;
    var clk abctime;
    var pFanin *Abc_Obj_t;
    var pFunc *Hop_Obj_t;
    assert(iFanin >= 0, "iFanin >= 0");
    p.nTryRemoves++;

    // clean simulation info
    Vec_PtrFillSimInfo( p.vDivCexes, 0, p.nDivWords ); 
    p.nCexes = 0;
    if p.pPars.fVeryVerbose {
//        printf( "\n" );
        fmt.Printf("%5d : Lev =%3d. Leaf =%3d. Node =%3d. Divs =%3d.  Fanin = %4d (%d/%d), MFFC = %d\n", pNode.Id, pNode.Level, Vec_PtrSize(p.vSupp), Vec_PtrSize(p.vNodes), Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode), Abc_ObjFaninId(pNode, iFanin), iFanin, Abc_ObjFaninNum(pNode), Abc_ObjFanoutNum(Abc_ObjFanin(pNode, iFanin)) == 1 ? Abc_NodeMffcLabel(Abc_ObjFanin(pNode, iFanin), nil) : 0);
    }

    // try fanins without the critical fanin
    nCands = 0;
    Vec_PtrClear( p.vMfsFanins );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if i == iFanin {
            continue; }
        Vec_PtrPush( p.vMfsFanins, pFanin );
        iVar = Vec_PtrSize(p.vDivs) - Abc_ObjFaninNum(pNode) + i;
        pCands[nCands++] = toLitCond( Vec_IntEntry( p.vProjVarsSat, iVar ), 1 );
    }
    RetValue = Abc_NtkMfsTryResubOnce( p, pCands, nCands );
    if RetValue == -1 {
        return 0; }
    if RetValue == 1 {
        if p.pPars.fVeryVerbose {
            fmt.Printf("Node %d: Fanin %d can be removed.\n", pNode.Id, iFanin); }
        p.nNodesResub++;
        p.nNodesGainedLevel++;
        if fSkipUpdate != 0 {
            return 1; }
clk = Abc_Clock();
        // derive the function
        pFunc = Abc_NtkMfsInterplate( p, pCands, nCands );
        if pFunc == nil {
            return 0; }
        // update the network
        Abc_NtkMfsUpdateNetwork( p, pNode, p.vMfsFanins, pFunc );
p.timeInt += Abc_Clock() - clk;
        p.nRemoves++;
        return 1;
    }

    if (fOnlyRemove != 0) || p.pPars.fRrOnly {
        return 0; }

    p.nTryResubs++;
    if fVeryVerbose != 0 {
        for  i = 0; i < 9; i++  {
            fmt.Printf(" ");
        }
        for  i = 0; i < Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode); i++  {
            fmt.Printf("%d", i % 10);
        }
        for  i = 0; i < Abc_ObjFaninNum(pNode); i++  {
            if i == iFanin {
                fmt.Printf("*");
            } else {
                fmt.Printf("%c", 'a' + i); }
        }
        fmt.Printf("\n");
    }
    iVar = -1;
    for {
        if fVeryVerbose != 0 {
            fmt.Printf("%3d: %3d ", p.nCexes, iVar);
            for  i = 0; i < Vec_PtrSize(p.vDivs); i++  {
                pData = (unsigned *)Vec_PtrEntry( p.vDivCexes, i );
                fmt.Printf("%d", Abc_InfoHasBit(pData, p.nCexes-1));
            }
            fmt.Printf("\n");
        }

        // find the next divisor to try
        nWords = Abc_BitWordNum(p.nCexes);
        assert(nWords <= p.nDivWords, "nWords <= p.nDivWords");
        for  iVar = 0; iVar < Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode); iVar++  {
            if p.pPars.fPower {
                var pDiv *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, iVar);
                // only accept the divisor if it is "cool"
                if Abc_MfsObjProb(p, pDiv) >= 0.15 {
                    continue; }
            }
            pData  = (unsigned *)Vec_PtrEntry( p.vDivCexes, iVar );
            for  w = 0; w < nWords; w++  {
                if pData[w] != ~0 {
                    break; }
            }
            if w == nWords {
                break; }
        }
        if iVar == Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode) {
            return 0; }

        pCands[nCands] = toLitCond( Vec_IntEntry(p.vProjVarsSat, iVar), 1 );
        RetValue = Abc_NtkMfsTryResubOnce( p, pCands, nCands+1 );
        if RetValue == -1 {
            return 0; }
        if RetValue == 1 {
            if p.pPars.fVeryVerbose {
                fmt.Printf("Node %d: Fanin %d can be replaced by divisor %d.\n", pNode.Id, iFanin, iVar); }
            p.nNodesResub++;
            p.nNodesGainedLevel++;
            if fSkipUpdate != 0 {
                return 1; }
clk = Abc_Clock();
            // derive the function
            pFunc = Abc_NtkMfsInterplate( p, pCands, nCands+1 );
            if pFunc == nil {
                return 0; }
            // update the network
            Vec_PtrPush( p.vMfsFanins, Vec_PtrEntry(p.vDivs, iVar) );
            Abc_NtkMfsUpdateNetwork( p, pNode, p.vMfsFanins, pFunc );
p.timeInt += Abc_Clock() - clk;
            p.nResubs++;
            return 1;
        }
        if p.nCexes >= p.pPars.nWinMax {
            break; }
    }
    if p.pPars.fVeryVerbose {
        fmt.Printf("Node %d: Cannot find replacement for fanin %d.\n", pNode.Id, iFanin); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs resubstitution for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsSolveSatResub2(p *Mfs_Man_t, pNode *Abc_Obj_t, iFanin int32, iFanin2 int32) int32 {
    var fVeryVerbose int32  = p.pPars.fVeryVerbose && Vec_PtrSize(p.vDivs) < 80;
    var pData, pData2 *uint32;
    var pCands [MFS_FANIN_MAX]int32;
    var RetValue, iVar, iVar2, i, w, nCands, nWords, fBreak int32;
    var clk abctime;
    var pFanin *Abc_Obj_t;
    var pFunc *Hop_Obj_t;
    assert(iFanin >= 0, "iFanin >= 0");
    assert(iFanin2 >= 0 || iFanin2 == -1, "iFanin2 >= 0 || iFanin2 == -1");

    // clean simulation info
    Vec_PtrFillSimInfo( p.vDivCexes, 0, p.nDivWords ); 
    p.nCexes = 0;
    if fVeryVerbose != 0 {
        fmt.Printf("\n");
        fmt.Printf("Node %5d : Level = %2d. Divs = %3d.  Fanins = %d/%d (out of %d). MFFC = %d\n", pNode.Id, pNode.Level, Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode), iFanin, iFanin2, Abc_ObjFaninNum(pNode), Abc_ObjFanoutNum(Abc_ObjFanin(pNode, iFanin)) == 1 ? Abc_NodeMffcLabel(Abc_ObjFanin(pNode, iFanin), nil) : 0);
    }

    // try fanins without the critical fanin
    nCands = 0;
    Vec_PtrClear( p.vMfsFanins );
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if i == iFanin || i == iFanin2 {
            continue; }
        Vec_PtrPush( p.vMfsFanins, pFanin );
        iVar = Vec_PtrSize(p.vDivs) - Abc_ObjFaninNum(pNode) + i;
        pCands[nCands++] = toLitCond( Vec_IntEntry( p.vProjVarsSat, iVar ), 1 );
    }
    RetValue = Abc_NtkMfsTryResubOnce( p, pCands, nCands );
    if RetValue == -1 {
        return 0; }
    if RetValue == 1 {
        if fVeryVerbose != 0 {
        fmt.Printf("Node %d: Fanins %d/%d can be removed.\n", pNode.Id, iFanin, iFanin2); }
        p.nNodesResub++;
        p.nNodesGainedLevel++;
clk = Abc_Clock();
        // derive the function
        pFunc = Abc_NtkMfsInterplate( p, pCands, nCands );
        if pFunc == nil {
            return 0; }
        // update the network
        Abc_NtkMfsUpdateNetwork( p, pNode, p.vMfsFanins, pFunc );
p.timeInt += Abc_Clock() - clk;
        return 1;
    }

    if fVeryVerbose != 0 {
        for  i = 0; i < 11; i++  {
            fmt.Printf(" ");
        }
        for  i = 0; i < Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode); i++  {
            fmt.Printf("%d", i % 10);
        }
        for  i = 0; i < Abc_ObjFaninNum(pNode); i++  {
            if i == iFanin || i == iFanin2 {
                fmt.Printf("*");
            } else {
                fmt.Printf("%c", 'a' + i); }
        }
        fmt.Printf("\n");
    }
    iVar = iVar2 = -1;
    for {
              
      
        if fVeryVerbose != 0 {
            fmt.Printf("%3d: %2d %2d ", p.nCexes, iVar, iVar2);
            for  i = 0; i < Vec_PtrSize(p.vDivs); i++  {
                pData = (unsigned *)Vec_PtrEntry( p.vDivCexes, i );
                fmt.Printf("%d", Abc_InfoHasBit(pData, p.nCexes-1));
            }
            fmt.Printf("\n");
        }

        // find the next divisor to try
        nWords = Abc_BitWordNum(p.nCexes);
        assert(nWords <= p.nDivWords, "nWords <= p.nDivWords");
        fBreak = 0;
        for  iVar = 1; iVar < Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode); iVar++  {
            pData  = (unsigned *)Vec_PtrEntry( p.vDivCexes, iVar );
               
            if p.pPars.fPower {
                var pDiv *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, iVar);
                // only accept the divisor if it is "cool"
                if Abc_MfsObjProb(p, pDiv) >= 0.12 {
                    continue; }
            }
      
            for  iVar2 = 0; iVar2 < iVar; iVar2++  {
                pData2 = (unsigned *)Vec_PtrEntry( p.vDivCexes, iVar2 );
              
                if p.pPars.fPower {
                    var pDiv *Abc_Obj_t  = (Abc_Obj_t *)Vec_PtrEntry(p.vDivs, iVar2);
                    // only accept the divisor if it is "cool"
                    if Abc_MfsObjProb(p, pDiv) >= 0.12 {
                        continue; }
                }
      
                for  w = 0; w < nWords; w++  {
                    if (pData[w] | pData2[w]) != ~0 {
                        break; }
                }
                if w == nWords {
                    fBreak = 1;
                    break;
                }
            }
            if fBreak != 0 {
                break; }
        }
        if iVar == Vec_PtrSize(p.vDivs)-Abc_ObjFaninNum(pNode) {
            return 0; }

        pCands[nCands]   = toLitCond( Vec_IntEntry(p.vProjVarsSat, iVar2), 1 );
        pCands[nCands+1] = toLitCond( Vec_IntEntry(p.vProjVarsSat, iVar), 1 );
        RetValue = Abc_NtkMfsTryResubOnce( p, pCands, nCands+2 );
        if RetValue == -1 {
            return 0; }
        if RetValue == 1 {
            if fVeryVerbose != 0 {
            fmt.Printf("Node %d: Fanins %d/%d can be replaced by divisors %d/%d.\n", pNode.Id, iFanin, iFanin2, iVar, iVar2); }
            p.nNodesResub++;
            p.nNodesGainedLevel++;
clk = Abc_Clock();
            // derive the function
            pFunc = Abc_NtkMfsInterplate( p, pCands, nCands+2 );
            if pFunc == nil {
                return 0; }
            // update the network
            Vec_PtrPush( p.vMfsFanins, Vec_PtrEntry(p.vDivs, iVar2) );
            Vec_PtrPush( p.vMfsFanins, Vec_PtrEntry(p.vDivs, iVar) );
            assert(Vec_PtrSize(p.vMfsFanins) == nCands + 2, "Vec_PtrSize(p.vMfsFanins) == nCands + 2");
            Abc_NtkMfsUpdateNetwork( p, pNode, p.vMfsFanins, pFunc );
p.timeInt += Abc_Clock() - clk;
            return 1;
        }
        if p.nCexes >= p.pPars.nWinMax {
            break; }
    }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Evaluates the possibility of replacing given edge by another edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsEdgeSwapEval(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 1 );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Evaluates the possibility of replacing given edge by another edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsEdgePower(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    // try replacing area critical fanins
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if Abc_MfsObjProb(p, pFanin) >= 0.35 {
            if Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 0 ) != 0 {
                return 1; }
        } else if Abc_MfsObjProb(p, pFanin) >= 0.25 { // sjang
        
            if Abc_NtkMfsSolveSatResub( p, pNode, i, 1, 0 ) != 0 {
                return 1; }
        }
        }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs resubstitution for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsResubNode(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    // try replacing area critical fanins
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if !Abc_ObjIsCi(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 {
            if Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 0 ) != 0 {
                return 1; }
        }
    }
    // try removing redundant edges
    if !p.pPars.fArea {
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
            if Abc_ObjIsCi(pFanin) || Abc_ObjFanoutNum(pFanin) != 1 {
                if Abc_NtkMfsSolveSatResub( p, pNode, i, 1, 0 ) != 0 {
                    return 1; }
            }
        }
    }
    if Abc_ObjFaninNum(pNode) == p.nFaninMax {
        return 0; }
/*
    // try replacing area critical fanins while adding two new fanins
    Abc_ObjForEachFanin( pNode, pFanin, i )
        if ( !Abc_ObjIsCi(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 )
        {
            if ( Abc_NtkMfsSolveSatResub2( p, pNode, i, -1 ) )
                return 1;
        }
*/
    return 0;
}

/**Function*************************************************************

  Synopsis    [Performs resubstitution for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsResubNode2(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
    var pFanin, pFanin2 *Abc_Obj_t;
    var i, k int32;
/*
    Abc_ObjForEachFanin( pNode, pFanin, i )
        if ( !Abc_ObjIsCi(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 )
        {
            if ( Abc_NtkMfsSolveSatResub( p, pNode, i, 0, 0 ) )
                return 1;
        }
*/
    if Abc_ObjFaninNum(pNode) < 2 {
        return 0; }
    // try replacing one area critical fanin and one other fanin while adding two new fanins
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pFanin) = Abc_ObjFanin(pNode, i)), 1); i++  {
        if !Abc_ObjIsCi(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 {
            // consider second fanin to remove at the same time
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pNode)) && (((pFanin2) = Abc_ObjFanin(pNode, k)), 1); k++  {
                if i != k && (Abc_NtkMfsSolveSatResub2( p, pNode, i, k ) != 0) {
                    return 1; }
            }
        }
    }
    return 0;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

