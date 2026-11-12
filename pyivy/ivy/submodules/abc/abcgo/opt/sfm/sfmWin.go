package sfm

import "fmt"

/**CFile****************************************************************

  FileName    [sfmWin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [Structural window computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmWin.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the MFFC size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ObjRef_rec(p *Sfm_Ntk_t, iObj int32) int32 {
    var i, iFanin, Value, Count int32;
    if Sfm_ObjIsPi(p, iObj) {
        return 0; }
    assert(Sfm_ObjIsNode(p, iObj), "Sfm_ObjIsNode(p, iObj)");
    Value = Sfm_ObjRefIncrement(p, iObj);
    if Value > 1 {
        return 0; }
    assert(Value == 1, "Value == 1");
    Count = 1;
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iObj)  && ((iFanin = Sfm_ObjFanin(p, iObj, i)), 1);  i++  {
        Count += Sfm_ObjRef_rec( p, iFanin );
    }
    return Count;
}
func Sfm_ObjRef(p *Sfm_Ntk_t, iObj int32) int32 {
    var i int32;
    var iFanin int32;
    var Count int32 = 1;
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iObj)  && ((iFanin = Sfm_ObjFanin(p, iObj, i)), 1);  i++  {
        Count += Sfm_ObjRef_rec( p, iFanin );
    }
    return Count;
}
func Sfm_ObjDeref_rec(p *Sfm_Ntk_t, iObj int32) int32 {
    var i, iFanin, Value, Count int32;
    if Sfm_ObjIsPi(p, iObj) {
        return 0; }
    assert(Sfm_ObjIsNode(p, iObj), "Sfm_ObjIsNode(p, iObj)");
    Value = Sfm_ObjRefDecrement(p, iObj);
    if Value > 0 {
        return 0; }
    assert(Value == 0, "Value == 0");
    Count = 1;
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iObj)  && ((iFanin = Sfm_ObjFanin(p, iObj, i)), 1);  i++  {
        Count += Sfm_ObjDeref_rec( p, iFanin );
    }
    return Count;
}
func Sfm_ObjDeref(p *Sfm_Ntk_t, iObj int32) int32 {
    var i int32;
    var iFanin int32;
    var Count int32 = 1;
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iObj)  && ((iFanin = Sfm_ObjFanin(p, iObj, i)), 1);  i++  {
        Count += Sfm_ObjDeref_rec( p, iFanin );
    }
    return Count;
}
func Sfm_ObjMffcSize(p *Sfm_Ntk_t, iObj int32) int32 {
    var Count1, Count2 int32;
    if Sfm_ObjIsPi(p, iObj) {
        return 0; }
    if Sfm_ObjFanoutNum(p, iObj) != 1 {
        return 0; }
    assert(Sfm_ObjIsNode( p, iObj ), "Sfm_ObjIsNode( p, iObj )");
    Count1 = Sfm_ObjDeref( p, iObj );
    Count2 = Sfm_ObjRef( p, iObj );
    assert(Count1 == Count2, "Count1 == Count2");
    return Count1;
}

/**Function*************************************************************

  Synopsis    [Working with traversal IDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkIncrementTravId(p *Sfm_Ntk_t) { p.nTravIds++;                                            }       
func Sfm_ObjSetTravIdCurrent(p *Sfm_Ntk_t, Id int32) { Vec_IntWriteEntry( &p.vTravIds, Id, p.nTravIds );       }
func Sfm_ObjIsTravIdCurrent(p *Sfm_Ntk_t, Id int32) int32 { return (Vec_IntEntry(&p.vTravIds, Id) == p.nTravIds);   }   
func Sfm_ObjIsTravIdPrevious(p *Sfm_Ntk_t, Id int32) int32 { return (Vec_IntEntry(&p.vTravIds, Id) == p.nTravIds-1); }   

func Sfm_NtkIncrementTravId2(p *Sfm_Ntk_t) { p.nTravIds2++;                                           }       
func Sfm_ObjSetTravIdCurrent2(p *Sfm_Ntk_t, Id int32) { Vec_IntWriteEntry( &p.vTravIds2, Id, p.nTravIds2 );     }
func Sfm_ObjIsTravIdCurrent2(p *Sfm_Ntk_t, Id int32) int32 { return (Vec_IntEntry(&p.vTravIds2, Id) == p.nTravIds2); }   


/**Function*************************************************************

  Synopsis    [Collects used internal nodes in a topological order.]

  Description [Additionally considers objects in groups as a single object
  and collects them in a topological order together as single entity.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkDfs_rec(p *Sfm_Ntk_t, iNode int32, vNodes *Vec_Int_t, vGroups *Vec_Wec_t, vGroupMap *Vec_Int_t, vBoxesLeft *Vec_Int_t) {
    var i, iFanin int32;
    if Sfm_ObjIsPi(p, iNode) {
        return; }
    if Sfm_ObjIsTravIdCurrent(p, iNode) != 0 {
        return; }
    if Vec_IntEntry(vGroupMap, iNode) >= 0 {
        var k int32;
        var iGroup int32 = Abc_Lit2Var( Vec_IntEntry(vGroupMap, iNode) );
        var vGroup *Vec_Int_t  = Vec_WecEntry( vGroups, iGroup );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vGroup)) && (((iNode) = Vec_IntEntry(vGroup, i)), 1); i++  {
            assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)");
        }
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vGroup)) && (((iNode) = Vec_IntEntry(vGroup, i)), 1); i++  {
            Sfm_ObjSetTravIdCurrent( p, iNode );
        }
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vGroup)) && (((iNode) = Vec_IntEntry(vGroup, i)), 1); i++  {
            // inlining of macro Sfm_ObjForEachFanin
            for  k = 0; k < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, k)), 1);  k++  {
                Sfm_NtkDfs_rec( p, iFanin, vNodes, vGroups, vGroupMap, vBoxesLeft );
            }
        }
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vGroup)) && (((iNode) = Vec_IntEntry(vGroup, i)), 1); i++  {
            Vec_IntPush( vNodes, iNode );
        }
        Vec_IntPush( vBoxesLeft, iGroup );
    
    } else {
        Sfm_ObjSetTravIdCurrent(p, iNode);
        // inlining of macro Sfm_ObjForEachFanin
        for  i = 0; i < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, i)), 1);  i++  {
            Sfm_NtkDfs_rec( p, iFanin, vNodes, vGroups, vGroupMap, vBoxesLeft );
        }
        Vec_IntPush( vNodes, iNode );
    }
}
func Sfm_NtkDfs(p *Sfm_Ntk_t, vGroups *Vec_Wec_t, vGroupMap *Vec_Int_t, vBoxesLeft *Vec_Int_t, fAllBoxes int32) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var i int32;
    Vec_IntClear( vBoxesLeft );
    vNodes = Vec_IntAlloc( p.nObjs );
    Sfm_NtkIncrementTravId( p );
    if fAllBoxes != 0 {
        var vGroup *Vec_Int_t;
        // inlining of macro Vec_WecForEachLevel
        for  i = 0; (i < Vec_WecSize(vGroups)) && (((vGroup) = Vec_WecEntry(vGroups, i)), 1); i++  {
            Sfm_NtkDfs_rec( p, Vec_IntEntry(vGroup, 0), vNodes, vGroups, vGroupMap, vBoxesLeft );
        }
    }
    // inlining of macro Sfm_NtkForEachPo
    for  i = p.nObjs - p.nPos; i < p.nObjs; i++  {
        Sfm_NtkDfs_rec( p, Sfm_ObjFanin(p, i, 0), vNodes, vGroups, vGroupMap, vBoxesLeft );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Check if this fanout overlaps with TFI cone of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkCheckOverlap_rec(p *Sfm_Ntk_t, iThis int32, iNode int32) int32 {
    var i, iFanin int32;
    if (Sfm_ObjIsTravIdCurrent2(p, iThis) != 0) || iThis == iNode {
        return 0; }
//    if ( Sfm_ObjIsTravIdCurrent(p, iThis) )
    if Sfm_ObjIsTravIdPrevious(p, iThis) != 0 {
        return 1; }
    Sfm_ObjSetTravIdCurrent2(p, iThis);
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iThis)  && ((iFanin = Sfm_ObjFanin(p, iThis, i)), 1);  i++  {
        if Sfm_NtkCheckOverlap_rec(p, iFanin, iNode) != 0 {
            return 1; }
    }
    return 0;
}
func Sfm_NtkCheckOverlap(p *Sfm_Ntk_t, iFan int32, iNode int32) int32 {
    Sfm_NtkIncrementTravId2( p );
    return Sfm_NtkCheckOverlap_rec( p, iFan, iNode );
}

/**Function*************************************************************

  Synopsis    [Recursively collects roots of the window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkCheckRoot(p *Sfm_Ntk_t, iNode int32, nLevelMax int32) int32 {
    var i, iFanout int32;
    // the node is the root if one of the following is true:
    // (1) the node has more than fanouts than the limit or has no fanouts (should not happen in general)
    if Sfm_ObjFanoutNum(p, iNode) == 0 || Sfm_ObjFanoutNum(p, iNode) > p.pPars.nFanoutMax {
        return 1; }
    // (2) the node has CO fanouts
    // (3) the node has fanouts above the cutoff level
    // inlining of macro Sfm_ObjForEachFanout
    for  i = 0; i < Sfm_ObjFanoutNum(p, iNode) && ((iFanout = Sfm_ObjFanout(p, iNode, i)), 1); i++  {
        if Sfm_ObjIsPo(p, iFanout) || Sfm_ObjLevel(p, iFanout) > nLevelMax {//|| !Sfm_NtkCheckOverlap(p, iFanout, iNode) )
            return 1; }
    }
    return 0;
}
func Sfm_NtkComputeRoots_rec(p *Sfm_Ntk_t, iNode int32, nLevelMax int32, vRoots *Vec_Int_t, vTfo *Vec_Int_t) {
    var i, iFanout int32;
    assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)");
    if Sfm_ObjIsTravIdCurrent(p, iNode) != 0 {
        return; }
    Sfm_ObjSetTravIdCurrent(p, iNode);
    if iNode != p.iPivotNode {
        Vec_IntPush( vTfo, iNode ); }
    // check if the node should be the root
    if Sfm_NtkCheckRoot( p, iNode, nLevelMax ) != 0 {
        Vec_IntPush( vRoots, iNode );
    } else { // if not, explore its fanouts
        // inlining of macro Sfm_ObjForEachFanout
        for  i = 0; i < Sfm_ObjFanoutNum(p, iNode) && ((iFanout = Sfm_ObjFanout(p, iNode, i)), 1); i++  {
            Sfm_NtkComputeRoots_rec( p, iFanout, nLevelMax, vRoots, vTfo );
        } }
}

/**Function*************************************************************

  Synopsis    [Collects divisors of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkAddDivisors(p *Sfm_Ntk_t, iNode int32, nLevelMax int32) {
    var i, iFanout int32;
    // inlining of macro Sfm_ObjForEachFanout
    for  i = 0; i < Sfm_ObjFanoutNum(p, iNode) && ((iFanout = Sfm_ObjFanout(p, iNode, i)), 1); i++  {
        // skip some of the fanouts if the number is large
        if p.pPars.nFanoutMax && i > p.pPars.nFanoutMax {
            return; }
        // skip TFI nodes, PO nodes, or nodes with high logic level
        if (Sfm_ObjIsTravIdCurrent(p, iFanout) != 0) || Sfm_ObjIsPo(p, iFanout) || Sfm_ObjLevel(p, iFanout) > nLevelMax {
            continue; }
        // handle single-input nodes
        if Sfm_ObjFaninNum(p, iFanout) == 1 {
            Vec_IntPush( p.vDivs, iFanout );
        // visit node for the first time
        } else if !(Sfm_ObjIsTravIdCurrent2(p, iFanout) != 0) {
            assert(Sfm_ObjFaninNum(p, iFanout) > 1, "Sfm_ObjFaninNum(p, iFanout) > 1");
            Sfm_ObjSetTravIdCurrent2( p, iFanout );
            Sfm_ObjResetFaninCount( p, iFanout );
        
        // visit node again
        } else if Sfm_ObjUpdateFaninCount(p, iFanout) == 0 {
            Vec_IntPush( p.vDivs, iFanout ); }
    }
}

/**Function*************************************************************

  Synopsis    [Fixed object is useful when it has a non-fixed fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ObjIsUseful(p *Sfm_Ntk_t, iNode int32) int32 {
    var i, iFanout int32;
    if !Sfm_ObjIsFixed(p, iNode) {
        return 1; }
    // inlining of macro Sfm_ObjForEachFanout
    for  i = 0; i < Sfm_ObjFanoutNum(p, iNode) && ((iFanout = Sfm_ObjFanout(p, iNode, i)), 1); i++  {
        if !Sfm_ObjIsFixed(p, iFanout) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Computes structural window.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_NtkCollectTfi_rec(p *Sfm_Ntk_t, iNode int32, vNodes *Vec_Int_t) int32 {
    var i, iFanin int32;
    if Sfm_ObjIsTravIdCurrent( p, iNode ) != 0 {
        return 0; }
    Sfm_ObjSetTravIdCurrent( p, iNode );
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iNode)  && ((iFanin = Sfm_ObjFanin(p, iNode, i)), 1);  i++  {
        if Sfm_NtkCollectTfi_rec( p, iFanin, vNodes ) != 0 {
            return 1; }
    }
    Vec_IntPush( vNodes, iNode );
    return p.pPars.nWinSizeMax && (Vec_IntSize(vNodes) > p.pPars.nWinSizeMax);
}
func Sfm_NtkCreateWindow(p *Sfm_Ntk_t, iNode int32, fVerbose int32) int32 {
    var i, k, iTemp int32;
    var clkDiv abctime;
    var clkWin abctime = Abc_Clock();

    assert(Sfm_ObjIsNode( p, iNode ), "Sfm_ObjIsNode( p, iNode )");
    p.iPivotNode = iNode;
    Vec_IntClear( p.vNodes );  // internal
    Vec_IntClear( p.vDivs );   // divisors
    Vec_IntClear( p.vRoots );  // roots
    Vec_IntClear( p.vTfo );    // roots
    Vec_IntClear( p.vOrder );  // variable order

    // collect transitive fanin
    Sfm_NtkIncrementTravId( p );
    if Sfm_NtkCollectTfi_rec( p, iNode, p.vNodes ) != 0 {
        p.nMaxDivs++;
        p.timeWin += Abc_Clock() - clkWin;
        return 0;
    }

    // create divisors
    clkDiv = Abc_Clock();
    Vec_IntClear( p.vDivs );
    Vec_IntAppend( p.vDivs, p.vNodes );
    Vec_IntPop( p.vDivs );
    // add non-topological divisors
    if Vec_IntSize(p.vDivs) < p.pPars.nWinSizeMax + 0 {
        Sfm_NtkIncrementTravId2( p );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vDivs)) && (((iTemp) = Vec_IntEntry(p.vDivs, i)), 1); i++  {
            if Vec_IntSize(p.vDivs) < p.pPars.nWinSizeMax + 0 {
//                Sfm_NtkAddDivisors( p, iTemp, Sfm_ObjLevel(p, iNode) - 1 );
                Sfm_NtkAddDivisors( p, iTemp, p.nLevelMax - Sfm_ObjLevelR(p, iNode) ); }
        } 
    }
    if Vec_IntSize(p.vDivs) > p.pPars.nWinSizeMax {
/*
        k = 0;
        Vec_IntForEachEntryStart( p->vDivs, iTemp, i, Vec_IntSize(p->vDivs) - p->pPars->nWinSizeMax )
            Vec_IntWriteEntry( p->vDivs, k++, iTemp );
        assert( k == p->pPars->nWinSizeMax );
*/
        Vec_IntShrink( p.vDivs, p.pPars.nWinSizeMax );
    }
    assert(Vec_IntSize(p.vDivs) <= p.pPars.nWinSizeMax, "Vec_IntSize(p.vDivs) <= p.pPars.nWinSizeMax");
    p.nMaxDivs += int32((Vec_IntSize(p.vDivs) == p.pPars.nWinSizeMax));
    // remove node/fanins from divisors
    // mark fanins
    Sfm_NtkIncrementTravId2( p );
    Sfm_ObjSetTravIdCurrent2( p, iNode );
    // inlining of macro Sfm_ObjForEachFanin
    for  i = 0; i < Sfm_ObjFaninNum(p, iNode)  && ((iTemp = Sfm_ObjFanin(p, iNode, i)), 1);  i++  {
        Sfm_ObjSetTravIdCurrent2( p, iTemp );
    }
    // compact divisors
    k = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vDivs)) && (((iTemp) = Vec_IntEntry(p.vDivs, i)), 1); i++  {
        if !(Sfm_ObjIsTravIdCurrent2(p, iTemp) != 0) && (Sfm_ObjIsUseful(p, iTemp) != 0) {
            Vec_IntWriteEntry( p.vDivs, k++, iTemp ); }
    }
    Vec_IntShrink( p.vDivs, k );
    assert(Vec_IntSize(p.vDivs) <= p.pPars.nWinSizeMax, "Vec_IntSize(p.vDivs) <= p.pPars.nWinSizeMax");
    clkDiv = Abc_Clock() - clkDiv;
    p.timeDiv += clkDiv;
    p.nTotalDivs += Vec_IntSize(p.vDivs);
 
    // collect TFO and window roots
    if p.pPars.nTfoLevMax > 0 && !(Sfm_NtkCheckRoot(p, iNode, Sfm_ObjLevel(p, iNode) + p.pPars.nTfoLevMax) != 0) {
        // explore transitive fanout
        Sfm_NtkIncrementTravId( p );
        Sfm_NtkComputeRoots_rec( p, iNode, Sfm_ObjLevel(p, iNode) + p.pPars.nTfoLevMax, p.vRoots, p.vTfo );
        assert(Vec_IntSize(p.vRoots) > 0, "Vec_IntSize(p.vRoots) > 0");
        assert(Vec_IntSize(p.vTfo) > 0, "Vec_IntSize(p.vTfo) > 0");
        // compute new leaves and nodes
        Sfm_NtkIncrementTravId( p );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vRoots)) && (((iTemp) = Vec_IntEntry(p.vRoots, i)), 1); i++  {
            if Sfm_NtkCollectTfi_rec( p, iTemp, p.vOrder ) != 0 {
                Vec_IntClear( p.vRoots );
                Vec_IntClear( p.vTfo );
                Vec_IntClear( p.vOrder );
                break;
            }
        }
        if Vec_IntSize(p.vRoots) > 0 {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vTfo)) && (((iTemp) = Vec_IntEntry(p.vTfo, i)), 1); i++  {
            if Sfm_NtkCollectTfi_rec( p, iTemp, p.vOrder ) != 0 {
                Vec_IntClear( p.vRoots );
                Vec_IntClear( p.vTfo );
                Vec_IntClear( p.vOrder );
                break;
            }
        } }
        if Vec_IntSize(p.vRoots) > 0 {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vDivs)) && (((iTemp) = Vec_IntEntry(p.vDivs, i)), 1); i++  {
            if Sfm_NtkCollectTfi_rec( p, iTemp, p.vOrder ) != 0 {
                Vec_IntClear( p.vRoots );
                Vec_IntClear( p.vTfo );
                Vec_IntClear( p.vOrder );
                break;
            }
        } }
    }

    if Vec_IntSize(p.vOrder) == 0 {
        var Temp int32  = p.pPars.nWinSizeMax;
        p.pPars.nWinSizeMax = 0;
        Sfm_NtkIncrementTravId( p );
        Sfm_NtkCollectTfi_rec( p, iNode, p.vOrder );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vDivs)) && (((iTemp) = Vec_IntEntry(p.vDivs, i)), 1); i++  {
            Sfm_NtkCollectTfi_rec( p, iTemp, p.vOrder );
        }
        p.pPars.nWinSizeMax = Temp;
    }

    // statistics
    p.timeWin += Abc_Clock() - clkWin - clkDiv;
    if !(fVerbose != 0) {
        return 1; }

    // print stats about the window
    fmt.Printf("%6d : ", iNode);
    fmt.Printf("Leaves = %5d. ", 0);
    fmt.Printf("Nodes = %5d. ", Vec_IntSize(p.vNodes));
    fmt.Printf("Roots = %5d. ", Vec_IntSize(p.vRoots));
    fmt.Printf("Divs = %5d. ", Vec_IntSize(p.vDivs));
    fmt.Printf("\n");
    return 1;
}
func Sfm_NtkWindowTest(p *Sfm_Ntk_t, iNode int32) {
    var i int32;
    // inlining of macro Sfm_NtkForEachNode
    for  i = p.nPis; i + p.nPos < p.nObjs; i++  {
        Sfm_NtkCreateWindow( p, i, 1 );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

