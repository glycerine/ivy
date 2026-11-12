package llb

import "fmt"

/**CFile****************************************************************

  FileName    [llb2Flow.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Flow computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Flow.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Llb_ObjSetPath(pObj *Aig_Obj_t, pNext *Aig_Obj_t) int32 { pObj.pData = (void *)pNext; return 1;  }
func Llb_ObjGetPath(pObj *Aig_Obj_t) *Aig_Obj_t { return (Aig_Obj_t *)pObj.pData;        }
func Llb_ObjGetFanoutPath(p *Aig_Man_t, pObj *Aig_Obj_t) *Aig_Obj_t { 
    var pFanout *Aig_Obj_t;
    var i int32;
    var iFanout int32 = -1;
    assert((Llb_ObjGetPath(pObj) != nil), "Llb_ObjGetPath(pObj)"); 
    Aig_ObjForEachFanout( p, pObj, pFanout, iFanout, i )
        if ( Llb_ObjGetPath(pFanout) == pObj )
            return pFanout;
    return nil;
}

// extern Vec_Ptr_t * Llb_ManCutSupp( Aig_Man_t * p, Vec_Ptr_t * vLower, Vec_Ptr_t * vUpper );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [For each cut, returns PIs that can be quantified.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutSupps(p *Aig_Man_t, vResult *Vec_Ptr_t) *Vec_Ptr_t {
    var vSupps, vOne, vLower, vUpper *Vec_Ptr_t;
    var i int32;
    vSupps = Vec_PtrAlloc( 100 );
    Vec_PtrPush( vSupps, Vec_PtrAlloc(0) );
    vLower = (Vec_Ptr_t *)Vec_PtrEntry( vResult, 0 );
    Vec_PtrForEachEntryStart( Vec_Ptr_t *, vResult, vUpper, i, 1 )
    {
        vOne  = Llb_ManCutSupp( p, vLower, vUpper );
        Vec_PtrPush( vSupps, vOne );
        vLower = vUpper;
    }
    assert(Vec_PtrSize(vSupps) == Vec_PtrSize(vResult), "Vec_PtrSize(vSupps) == Vec_PtrSize(vResult)");
    return vSupps;
}

/**Function*************************************************************

  Synopsis    [For each cut, returns PIs that can be quantified.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutMap(p *Aig_Man_t, vResult *Vec_Ptr_t, vSupps *Vec_Ptr_t) *Vec_Ptr_t {
    var fShowMatrix int32  = 1;
    var vMaps, vOne *Vec_Ptr_t;
    var vMap, vPrev, vNext *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var piFirst, piLast *int32;
    var i, k, CounterPlus, CounterMinus, Counter int32;

    vMaps = Vec_PtrAlloc( 100 );
    Vec_PtrForEachEntry( Vec_Ptr_t *, vResult, vOne, i )
    {
        vMap = Vec_IntStart( Aig_ManObjNumMax(p) );
        Vec_PtrForEachEntry( Aig_Obj_t *, vOne, pObj, k )
        {
            if !Saig_ObjIsPi(p, pObj) {
                Vec_IntWriteEntry( vMap, pObj.Id, 1 ); }
//            else
//printf( "*" );
//printf( "%d ", pObj->Id );
        }
        Vec_PtrPush( vMaps, vMap );
//printf( "\n" );
    }
    Vec_PtrPush( vMaps, Vec_IntStart( Aig_ManObjNumMax(p) ) );
    assert(Vec_PtrSize(vMaps) == Vec_PtrSize(vResult)+1, "Vec_PtrSize(vMaps) == Vec_PtrSize(vResult)+1");

    // collect the first and last PIs
    piFirst = ABC_ALLOC( int, Saig_ManPiNum(p) );
    piLast  = ABC_ALLOC( int, Saig_ManPiNum(p) );
    Saig_ManForEachPi( p, pObj, i )
        piFirst[i] = piLast[i] = -1;
    Vec_PtrForEachEntry( Vec_Ptr_t *, vSupps, vOne, i )
    {
        Vec_PtrForEachEntry( Aig_Obj_t *, vOne, pObj, k )
        {
            if !Saig_ObjIsPi(p, pObj) {
                continue; }
            if piFirst[Aig_ObjCioId(pObj)] == -1 {
                 piFirst[Aig_ObjCioId(pObj)] = i; }
            piLast[Aig_ObjCioId(pObj)] = i;
        }
    }
    // PIs feeding into the flops should be extended to the last frame
    Saig_ManForEachLi( p, pObj, i )
    {
        if !Saig_ObjIsPi(p, Aig_ObjFanin0(pObj)) {
            continue; }
        piLast[Aig_ObjCioId(Aig_ObjFanin0(pObj))] = Vec_PtrSize(vMaps)-1;
    }

    // set the PI map
    Saig_ManForEachPi( p, pObj, i )
    {
        if piFirst[i] == -1 {
            continue; }
        if piFirst[i] == piLast[i] {
            vMap = (Vec_Int_t *)Vec_PtrEntry( vMaps, piFirst[i] );
            Vec_IntWriteEntry( vMap, pObj.Id, 2 );
            continue;
        }

        // set support for all in between        
        for  k = piFirst[i]; k <= piLast[i]; k++  {
            vMap = (Vec_Int_t *)Vec_PtrEntry( vMaps, k );
            Vec_IntWriteEntry( vMap, pObj.Id, 1 );
        }
    }
    ABC_FREE( piFirst );
    ABC_FREE( piLast );


    // find all that will appear here
    Counter = Aig_ManRegNum(p);
    fmt.Printf("%d ", Counter);
    Vec_PtrForEachEntryStart( Vec_Int_t *, vMaps, vMap, i, 1 )
    {
        vPrev = (Vec_Int_t *)Vec_PtrEntry( vMaps, i-1 );
        if i == Vec_PtrSize(vMaps)-1 {
            vNext = nil;
        } else {
            vNext = (Vec_Int_t *)Vec_PtrEntry( vMaps, i+1 );
        }

        CounterPlus = CounterMinus = 0;
        Aig_ManForEachObj( p, pObj, k )
        {
            if Saig_ObjIsPi(p, pObj) {
                if Vec_IntEntry(vPrev, k) == 0 && Vec_IntEntry(vMap, k) == 1 {
                    CounterPlus++; }
                if Vec_IntEntry(vMap, k) == 1 && (vNext == nil || Vec_IntEntry(vNext, k) == 0) {
                    CounterMinus++; }
            
            } else {
                if Vec_IntEntry(vPrev, k) == 0 && Vec_IntEntry(vMap, k) == 1 {
                    CounterPlus++; }
                if Vec_IntEntry(vPrev, k) == 1 && Vec_IntEntry(vMap, k) == 0 {
                    CounterMinus++; }
            }
        }
        Counter = Counter + CounterPlus - CounterMinus;
        fmt.Printf("%d=%d ", i, Counter);
    }
    fmt.Printf("\n");

    if !(fShowMatrix != 0) {
        return vMaps; }
    Aig_ManForEachObj( p, pObj, i )
    {
        if !Aig_ObjIsCi(pObj) && !Aig_ObjIsNode(pObj) {
            continue; }
        Vec_PtrForEachEntry( Vec_Int_t *, vMaps, vMap, k )
            if ( Vec_IntEntry(vMap, i) )
                break;
        if k == Vec_PtrSize(vMaps) {
            continue; }
        fmt.Printf("Obj = %4d : ", i);
        if Saig_ObjIsPi(p,pObj) {
            fmt.Printf("pi  ");
        } else if Saig_ObjIsLo(p,pObj) {
            fmt.Printf("lo  ");
        } else if Aig_ObjIsNode(pObj) {
            fmt.Printf("and "); }

        Vec_PtrForEachEntry( Vec_Int_t *, vMaps, vMap, k )
            fmt.Printf("%d", Vec_IntEntry(vMap, i));
        fmt.Printf("\n");
    }
    return vMaps;
}

/**Function*************************************************************

  Synopsis    [Counts the number of PIs in the cut]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutPiNum(p *Aig_Man_t, vMinCut *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    Vec_PtrForEachEntry( Aig_Obj_t *, vMinCut, pObj, i )
        if ( Saig_ObjIsPi(p,pObj) )
            Counter++;
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of LOs in the cut]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutLoNum(p *Aig_Man_t, vMinCut *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    Vec_PtrForEachEntry( Aig_Obj_t *, vMinCut, pObj, i )
        if ( Saig_ObjIsLo(p,pObj) )
            Counter++;
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of LIs in the cut]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutLiNum(p *Aig_Man_t, vMinCut *Vec_Ptr_t) int32 {
    var pFanout *Aig_Obj_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var k int32;
    var iFanout int32 = -1;
    var Counter int32;
    Vec_PtrForEachEntry( Aig_Obj_t *, vMinCut, pObj, i )
    {
        if Aig_ObjIsCi(pObj) {
            continue; }
        Aig_ObjForEachFanout( p, pObj, pFanout, iFanout, k )
        {
            if Saig_ObjIsLi(p, pFanout) {
                Counter++;
                break;
            }
        }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutVolume_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    return 1 + Llb_ManCutVolume_rec(p, Aig_ObjFanin0(pObj)) + 
        Llb_ManCutVolume_rec(p, Aig_ObjFanin1(pObj));
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutVolume(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // mark the lower cut with the traversal ID
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
        Aig_ObjSetTravIdCurrent( p, pObj );
    // count the upper cut
    Vec_PtrForEachEntry( Aig_Obj_t *, vUpper, pObj, i )
        Counter += Llb_ManCutVolume_rec( p, pObj );
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutNodes_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vNodes *Vec_Ptr_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Llb_ManCutNodes_rec(p, Aig_ObjFanin0(pObj), vNodes);
    Llb_ManCutNodes_rec(p, Aig_ObjFanin1(pObj), vNodes);
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutNodes(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // mark the lower cut with the traversal ID
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
        Aig_ObjSetTravIdCurrent( p, pObj );
    // count the upper cut
    vNodes = Vec_PtrAlloc( 100 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vUpper, pObj, i )
        Llb_ManCutNodes_rec( p, pObj, vNodes );
    return vNodes;
}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutSupp(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) *Vec_Ptr_t {
    var vNodes, vSupp *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vNodes = Llb_ManCutNodes( p, vLower, vUpper );
    // mark support of the nodes
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vNodes, pObj, i )
    {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        Aig_ObjSetTravIdCurrent( p, Aig_ObjFanin0(pObj) );
        Aig_ObjSetTravIdCurrent( p, Aig_ObjFanin1(pObj) );
    }
    Vec_PtrFree( vNodes );
    // collect the support nodes
    vSupp = Vec_PtrAlloc( 100 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
        if ( Aig_ObjIsTravIdCurrent(p, pObj) )
            Vec_PtrPush( vSupp, pObj );
    return vSupp;

}

/**Function*************************************************************

  Synopsis    [Computes volume of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutRange(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) *Vec_Ptr_t {
    var vRange *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // mark the lower cut with the traversal ID
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
        Aig_ObjSetTravIdCurrent( p, pObj );
    // collect the upper ones that are not marked
    vRange = Vec_PtrAlloc( 100 );
    Vec_PtrForEachEntry( Aig_Obj_t *, vUpper, pObj, i )
        if ( !Aig_ObjIsTravIdCurrent(p, pObj) )
            Vec_PtrPush( vRange, pObj );
    return vRange;
}




/**Function*************************************************************

  Synopsis    [Prints the given cluster.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManCutPrint(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) {
    var vSupp, vRange *Vec_Ptr_t;
    var Pis, Ffs, And int32;

    Pis = Llb_ManCutPiNum(p, vLower);
    Ffs = Llb_ManCutLoNum(p, vLower);
    And = Vec_PtrSize(vLower) - Pis - Ffs;
    fmt.Printf("Leaf: %3d=%3d+%3d+%3d  ", Vec_PtrSize(vLower), Pis, Ffs, And);

    Pis = Llb_ManCutPiNum(p, vUpper);
    Ffs = Llb_ManCutLiNum(p, vUpper);
    And = Vec_PtrSize(vUpper) - Pis - Ffs;
    fmt.Printf("Root: %3d=%3d+%3d+%3d  ", Vec_PtrSize(vUpper), Pis, Ffs, And);

    vSupp = Llb_ManCutSupp( p, vLower, vUpper );
    Pis = Llb_ManCutPiNum(p, vSupp);
    Ffs = Llb_ManCutLoNum(p, vSupp);
    And = Vec_PtrSize(vSupp) - Pis - Ffs;
    fmt.Printf("Supp: %3d=%3d+%3d+%3d  ", Vec_PtrSize(vSupp), Pis, Ffs, And);

    vRange = Llb_ManCutRange( p, vLower, vUpper );
    Pis = Llb_ManCutPiNum(p, vRange);
    Ffs = Llb_ManCutLiNum(p, vRange);
    And = Vec_PtrSize(vRange) - Pis - Ffs;
    fmt.Printf("Range: %3d=%3d+%3d+%3d  ", Vec_PtrSize(vRange), Pis, Ffs, And);

    fmt.Printf("S =%3d. V =%3d.\n", Vec_PtrSize(vSupp)+Vec_PtrSize(vRange), Llb_ManCutVolume(p, vLower, vUpper));
    Vec_PtrFree( vSupp );
    Vec_PtrFree( vRange );
/*   
    {
        Aig_Obj_t * pObj;
        int i;
        printf( "Lower: " );
        Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
            printf( " %d", pObj->Id );
        printf( "     " );
        printf( "Upper: " );
        Vec_PtrForEachEntry( Aig_Obj_t *, vUpper, pObj, i )
            printf( " %d", pObj->Id );
        printf( "\n" );
    }
*/
}

/**Function*************************************************************

  Synopsis    [Prints the given cluster.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManResultPrint(p *Aig_Man_t, vResult *Vec_Ptr_t) {
    var vLower *Vec_Ptr_t;
    var vUpper *Vec_Ptr_t = nil;
    var i int32;
    Vec_PtrForEachEntryReverse( Vec_Ptr_t *, vResult, vLower, i )
    {
        if i < Vec_PtrSize(vResult) - 1 {
            Llb_ManCutPrint( p, vLower, vUpper ); }
        vUpper = vLower;
    }
}

/**Function*************************************************************

  Synopsis    [Tries to find an augmenting path originating in this node.]

  Description [This procedure works for directed graphs only!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowBwdPath2_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var pFanout *Aig_Obj_t;
    assert(Aig_ObjIsNode(pObj) || Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj), "Aig_ObjIsNode(pObj) || Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj)");
    // skip visited nodes
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    // process node without flow
    if !(Llb_ObjGetPath(pObj) != nil) {
        // start the path if we reached a terminal node
        if pObj.fMarkA {
            return Llb_ObjSetPath( pObj, (Aig_Obj_t *)1 ); }
        // explore the fanins
//        Abc_ObjForEachFanin( pObj, pFanin, i )
//            if ( Abc_NtkMaxFlowBwdPath2_rec(pFanin) )
//                return Abc_ObjSetPath( pObj, pFanin );
        if Aig_ObjIsNode(pObj) {
            if Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin0(pObj) ) != 0 {
                return Llb_ObjSetPath( pObj, Aig_ObjFanin0(pObj) ); }
            if Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin1(pObj) ) != 0 {
                return Llb_ObjSetPath( pObj, Aig_ObjFanin1(pObj) ); }
        }
        return 0;
    }
    // pObj has flow - find the fanout with flow
    pFanout = Llb_ObjGetFanoutPath( p, pObj );
    if pFanout == nil {
        return 0; }
    // go through the fanins of the fanout with flow
//    Abc_ObjForEachFanin( pFanout, pFanin, i )
//        if ( Abc_NtkMaxFlowBwdPath2_rec( pFanin ) )
//            return Abc_ObjSetPath( pFanout, pFanin );
    assert(Aig_ObjIsNode(pFanout), "Aig_ObjIsNode(pFanout)");
    if Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin0(pFanout) ) != 0 {
        return Llb_ObjSetPath( pFanout, Aig_ObjFanin0(pFanout) ); }
    if Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin1(pFanout) ) != 0 {
        return Llb_ObjSetPath( pFanout, Aig_ObjFanin1(pFanout) ); }
    // try the fanout
    if Llb_ManFlowBwdPath2_rec( p, pFanout ) != 0 {
        return Llb_ObjSetPath( pFanout, nil ); }
    return 0;
}


/**Function*************************************************************

  Synopsis    [Cleans markB.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowLabelTfi_rec(p *Aig_Man_t, pObj *Aig_Obj_t) {
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj) {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Llb_ManFlowLabelTfi_rec( p, Aig_ObjFanin0(pObj) );
    Llb_ManFlowLabelTfi_rec( p, Aig_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowUpdateCut(p *Aig_Man_t, vMinCut *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // label the TFI of the cut nodes
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vMinCut, pObj, i )
        Llb_ManFlowLabelTfi_rec( p, pObj );
    // collect labeled fanins of non-labeled nodes
    Vec_PtrClear( vMinCut );
    Aig_ManIncrementTravId(p);
    Aig_ManForEachObj( p, pObj, i )
    {
        if !Aig_ObjIsCo(pObj) && !Aig_ObjIsNode(pObj) {
            continue; }
        if Aig_ObjIsTravIdCurrent(p, pObj) || Aig_ObjIsTravIdPrevious(p, pObj) {
            continue; }
        if Aig_ObjIsTravIdPrevious(p, Aig_ObjFanin0(pObj)) {
            Aig_ObjSetTravIdCurrent(p, Aig_ObjFanin0(pObj));
            Vec_PtrPush( vMinCut, Aig_ObjFanin0(pObj) );
        }
        if Aig_ObjIsNode(pObj) && Aig_ObjIsTravIdPrevious(p, Aig_ObjFanin1(pObj)) {
            Aig_ObjSetTravIdCurrent(p, Aig_ObjFanin1(pObj));
            Vec_PtrPush( vMinCut, Aig_ObjFanin1(pObj) );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Find minimum-volume minumum cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowMinCut(p *Aig_Man_t) *Vec_Ptr_t {
    var vMinCut *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // collect the cut nodes
    vMinCut = Vec_PtrAlloc( Aig_ManRegNum(p) );
    Aig_ManForEachObj( p, pObj, i )
    {
        // node without flow is not a cut node
        if !(Llb_ObjGetPath(pObj) != nil) {
            continue; }
        // unvisited node is below the cut
        if !Aig_ObjIsTravIdCurrent(p, pObj) {
            continue; }
        // add terminal with flow or node whose path is not visited
        if pObj.fMarkA || !Aig_ObjIsTravIdCurrent( p, Llb_ObjGetPath(pObj) ) {
            Vec_PtrPush( vMinCut, pObj ); }
    }
    return vMinCut;
}

/**Function*************************************************************

  Synopsis    [Verifies the min-cut is indeed a cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowVerifyCut_rec(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    // skip visited nodes
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return 1; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    // visit the node
    if Aig_ObjIsConst1(pObj) {
        return 1; }
    if Aig_ObjIsCi(pObj) {
        return 0; }
    // explore the fanins
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    if !(Llb_ManFlowVerifyCut_rec(p, Aig_ObjFanin0(pObj)) != 0) {
        return 0; }
    if !(Llb_ManFlowVerifyCut_rec(p, Aig_ObjFanin1(pObj)) != 0) {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Verifies the min-cut is indeed a cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowVerifyCut(p *Aig_Man_t, vMinCut *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // mark the cut with the current traversal ID
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vMinCut, pObj, i )
        Aig_ObjSetTravIdCurrent( p, pObj );
    // search from the latches for a path to the COs/CIs
    Saig_ManForEachLi( p, pObj, i )
    {
        if !(Llb_ManFlowVerifyCut_rec( p, Aig_ObjFanin0(pObj) ) != 0) {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Implementation of max-flow/min-cut computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlow(p *Aig_Man_t, vSources *Vec_Ptr_t, pnFlow *int32) *Vec_Ptr_t {
    var vMinCut *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var Flow, FlowCur, RetValue, i int32;
    // find the max-flow
    Flow = 0;
    Aig_ManCleanData( p );
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vSources, pObj, i )
    {
        assert(!pObj.fMarkA && pObj.fMarkB, "!pObj.fMarkA && pObj.fMarkB");
        if !Aig_ObjFanin0(pObj).fMarkB {
            FlowCur  = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin0(pObj) );
            Flow    += FlowCur;
            if FlowCur {
                Aig_ManIncrementTravId(p); }
        }
        if Aig_ObjIsNode(pObj) && !Aig_ObjFanin1(pObj).fMarkB {
            FlowCur  = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin1(pObj) );
            Flow    += FlowCur;
            if FlowCur {
                Aig_ManIncrementTravId(p); }
        }
    }
    if pnFlow != nil {
        *pnFlow = Flow; }

    // mark the nodes reachable from the latches
    Aig_ManIncrementTravId(p);
    Vec_PtrForEachEntry( Aig_Obj_t *, vSources, pObj, i )
    {
        assert(!pObj.fMarkA && pObj.fMarkB, "!pObj.fMarkA && pObj.fMarkB");
        if !Aig_ObjFanin0(pObj).fMarkB {
            RetValue = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin0(pObj) );
            assert(RetValue == 0, "RetValue == 0");
        }
        if Aig_ObjIsNode(pObj) && !Aig_ObjFanin1(pObj).fMarkB {
            RetValue = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin1(pObj) );
            assert(RetValue == 0, "RetValue == 0");
        }
    }

    // find the min-cut with the smallest volume
    vMinCut = Llb_ManFlowMinCut( p );
    assert(Vec_PtrSize(vMinCut) == Flow, "Vec_PtrSize(vMinCut) == Flow");
    // verify the cut
    if !(Llb_ManFlowVerifyCut(p, vMinCut) != 0) {
        fmt.Printf("Llb_ManFlow() error! The computed min-cut is not a cut!\n"); }
//    Llb_ManFlowPrintCut( p, vMinCut );
    return vMinCut;
}

/**Function*************************************************************

  Synopsis    [Implementation of max-flow/min-cut computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowCompute(p *Aig_Man_t) *Vec_Ptr_t {
    var vMinCut *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var Flow, FlowCur, RetValue, i int32;
    // find the max-flow
    Flow = 0;
    Aig_ManCleanData( p );
    Aig_ManIncrementTravId(p);
    Aig_ManForEachObj( p, pObj, i )
    {
        if !pObj.fMarkB {
            continue; }
        assert(!pObj.fMarkA, "!pObj.fMarkA");
        if !Aig_ObjFanin0(pObj).fMarkB {
//printf( "%d ", Aig_ObjFanin0(pObj)->Id );
            FlowCur  = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin0(pObj) );
            Flow    += FlowCur;
            if FlowCur {
                Aig_ManIncrementTravId(p); }
        }
        if Aig_ObjIsNode(pObj) && !Aig_ObjFanin1(pObj).fMarkB {
//printf( "%d ", Aig_ObjFanin1(pObj)->Id );
            FlowCur  = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin1(pObj) );
            Flow    += FlowCur;
            if FlowCur {
                Aig_ManIncrementTravId(p); }
        }
    }
//printf( "\n" );

    // mark the nodes reachable from the latches
    Aig_ManIncrementTravId(p);
    Aig_ManForEachObj( p, pObj, i )
    {
        if !pObj.fMarkB {
            continue; }
        assert(!pObj.fMarkA, "!pObj.fMarkA");
        if !Aig_ObjFanin0(pObj).fMarkB {
            RetValue = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin0(pObj) );
            assert(RetValue == 0, "RetValue == 0");
        }
        if Aig_ObjIsNode(pObj) && !Aig_ObjFanin1(pObj).fMarkB {
            RetValue = Llb_ManFlowBwdPath2_rec( p, Aig_ObjFanin1(pObj) );
            assert(RetValue == 0, "RetValue == 0");
        }
    }
    // find the min-cut with the smallest volume
    vMinCut = Llb_ManFlowMinCut( p );
    assert(Vec_PtrSize(vMinCut) == Flow, "Vec_PtrSize(vMinCut) == Flow");
//printf( "%d ", Vec_PtrSize(vMinCut) );
    Llb_ManFlowUpdateCut( p, vMinCut );
//printf( "%d   ", Vec_PtrSize(vMinCut) );
    // verify the cut
    if !(Llb_ManFlowVerifyCut(p, vMinCut) != 0) {
        fmt.Printf("Llb_ManFlow() error! The computed min-cut is not a cut!\n"); }
//    Llb_ManFlowPrintCut( p, vMinCut );
    return vMinCut;
}




/**Function*************************************************************

  Synopsis    [Cleans markB.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowCleanMarkB_rec(pObj *Aig_Obj_t) {
    if pObj.fMarkB == 0 {
        return; }
    pObj.fMarkB = 0;
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Llb_ManFlowCleanMarkB_rec( Aig_ObjFanin0(pObj) );
    Llb_ManFlowCleanMarkB_rec( Aig_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Cleans markB.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowSetMarkA_rec(pObj *Aig_Obj_t) {
    if pObj.fMarkA {
        return; }
    pObj.fMarkA = 1;
    if Aig_ObjIsCi(pObj) || Aig_ObjIsConst1(pObj) {
        return; }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Llb_ManFlowSetMarkA_rec( Aig_ObjFanin0(pObj) );
    Llb_ManFlowSetMarkA_rec( Aig_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Prepares flow computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowPrepareCut(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    // reset marks
    Aig_ManForEachObj( p, pObj, i )
    {
        pObj.fMarkA = 0;
        pObj.fMarkB = 1;
    }
    // clean PIs and const
    Aig_ManConst1(p).fMarkB = 0;
    Aig_ManForEachCi( p, pObj, i )
        pObj.fMarkB = 0;
    // clean upper cut
//printf( "Upper: ");
    Vec_PtrForEachEntry( Aig_Obj_t *, vUpper, pObj, i )
    {
        Llb_ManFlowCleanMarkB_rec( pObj );
//printf( "%d ", pObj->Id );
    }
//printf( "\n" );
    // set lower cut
//printf( "Lower: ");
    Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
    {
//printf( "%d ", pObj->Id );
        assert(pObj.fMarkB == 0, "pObj.fMarkB == 0");
        Llb_ManFlowSetMarkA_rec( pObj );
    }
//printf( "\n" );
}

/**Function*************************************************************

  Synopsis    [Prepares flow computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowUnmarkCone(p *Aig_Man_t, vCone *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_PtrForEachEntry( Aig_Obj_t *, vCone, pObj, i )
    {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        assert(pObj.fMarkB == 1, "pObj.fMarkB == 1");
        pObj.fMarkB = 0;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowCollectAndMarkCone_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vCone *Vec_Ptr_t) {
    var pFanout *Aig_Obj_t;
    var i int32;
    var iFanout int32 = -1;
    if Saig_ObjIsLi(p, pObj) {
        return; }
    if pObj.fMarkB {
        return; }
    if pObj.fMarkA == 0 {
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        pObj.fMarkB = 1;
        if Aig_ObjIsNode(pObj) {
            Vec_PtrPush( vCone, pObj ); }
    }
    Aig_ObjForEachFanout( p, pObj, pFanout, iFanout, i )
        Llb_ManFlowCollectAndMarkCone_rec( p, pFanout, vCone );
}

/**Function*************************************************************

  Synopsis    [Collects the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowCollectAndMarkCone(p *Aig_Man_t, vStarts *Vec_Ptr_t, vCone *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_PtrClear( vCone );
    Vec_PtrForEachEntry( Aig_Obj_t *, vStarts, pObj, i )
    {
        assert(pObj.fMarkA && !pObj.fMarkB, "pObj.fMarkA && !pObj.fMarkB");
        Llb_ManFlowCollectAndMarkCone_rec( p, pObj, vCone );
    }
}




/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManComputeCutLo(p *Aig_Man_t) *Vec_Ptr_t {
    var vMinCut *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vMinCut = Vec_PtrAlloc( 100 );
    Aig_ManForEachCi( p, pObj, i )
        Vec_PtrPush( vMinCut, pObj );
    return vMinCut;
}

/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManComputeCutLi(p *Aig_Man_t) *Vec_Ptr_t {
    var vMinCut *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    assert(Saig_ManPoNum(p) == 0, "Saig_ManPoNum(p) == 0");
    vMinCut = Vec_PtrAlloc( 100 );
    Aig_ManIncrementTravId(p);
    Saig_ManForEachLi( p, pObj, i )
    {
        pObj = Aig_ObjFanin0(pObj);
        if Aig_ObjIsConst1(pObj) {
            continue; }
        if Aig_ObjIsTravIdCurrent(p, pObj) {
            continue; }
        Aig_ObjSetTravIdCurrent(p, pObj);
        Vec_PtrPush( vMinCut, pObj );
    }
    return vMinCut;
}



/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowGetObjSet(p *Aig_Man_t, vLower *Vec_Ptr_t, iStart int32, nSize int32, vSet *Vec_Ptr_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_PtrClear( vSet );
    for  i = 0; i < nSize; i++  {
        pObj = (Aig_Obj_t *)Vec_PtrEntry( vLower, (iStart + i) % Vec_PtrSize(vLower) );
        Vec_PtrPush( vSet, pObj );
    }
}

/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManFlowFindBestCut(p *Aig_Man_t, vLower *Vec_Ptr_t, vUpper *Vec_Ptr_t, Num int32) *Vec_Ptr_t {
    var nVolMin int32  = Aig_ManNodeNum(p) / Num / 2;
    var vMinCut *Vec_Ptr_t;
    var vCone, vSet *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i, s, Vol, VolLower, VolUpper, VolCmp int32;
    var iBest int32 = -1;
    var iMinCut int32 = ABC_INFINITY;
    var iVolBest int32;

    Vol = Llb_ManCutVolume( p, vLower, vUpper );
    assert(Vol > nVolMin, "Vol > nVolMin");
    VolCmp = Abc_MinInt( nVolMin, Vol - nVolMin );
    vCone = Vec_PtrAlloc( 100 );
    vSet  = Vec_PtrAlloc( 100 );
    Llb_ManFlowPrepareCut( p, vLower, vUpper );
    for  s = 1; s < Aig_ManRegNum(p); s += 5  {
        Vec_PtrForEachEntry( Aig_Obj_t *, vLower, pObj, i )
        {
            Llb_ManFlowGetObjSet( p, vLower, i, s, vSet );
            Llb_ManFlowCollectAndMarkCone( p, vSet, vCone );
            if Vec_PtrSize(vCone) == 0 {
                continue; }
            vMinCut  = Llb_ManFlowCompute( p );
            Llb_ManFlowUnmarkCone( p, vCone );

            VolLower = Llb_ManCutVolume( p, vLower, vMinCut );
            VolUpper = Llb_ManCutVolume( p, vMinCut, vUpper );
            Vol = Abc_MinInt( VolLower, VolUpper );
            if Vol >= VolCmp &&  (iMinCut == -1 || 
                                    iMinCut >  Vec_PtrSize(vMinCut) || 
                                   (iMinCut == Vec_PtrSize(vMinCut) && iVolBest < Vol)) {
                iBest = i;
                iMinCut = Vec_PtrSize(vMinCut);
                iVolBest = Vol;
            }
            Vec_PtrFree( vMinCut );
        }
        if iBest >= 0 {
            break; }
    }
    if iBest == -1 {
        // cleanup
        Vec_PtrFree( vCone );
        Vec_PtrFree( vSet );
        return nil;
    }
    // get the best cut
    assert(iBest >= 0, "iBest >= 0");
    Llb_ManFlowGetObjSet( p, vLower, iBest, s, vSet );
    Llb_ManFlowCollectAndMarkCone( p, vSet, vCone );
    vMinCut = Llb_ManFlowCompute( p );
    Llb_ManFlowUnmarkCone( p, vCone );
    // cleanup
    Vec_PtrFree( vCone );
    Vec_PtrFree( vSet );
    return vMinCut;
}

/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManComputeCuts(p *Aig_Man_t, Num int32, fVerbose int32, fVeryVerbose int32) *Vec_Ptr_t {
    var nVolMax int32  = Aig_ManNodeNum(p) / Num;
    var vResult *Vec_Ptr_t;
    var vMinCut *Vec_Ptr_t = nil;
    var vLower *Vec_Ptr_t;
    var vUpper *Vec_Ptr_t;
    var i, k, nVol int32;
    var clk abctime  = Abc_Clock();
    vResult = Vec_PtrAlloc( 100 );
    Vec_PtrPush( vResult, Llb_ManComputeCutLo(p) );
    Vec_PtrPush( vResult, Llb_ManComputeCutLi(p) );
    for {
        // find a place to insert new cut
        vLower = (Vec_Ptr_t *)Vec_PtrEntry( vResult, 0 );
        Vec_PtrForEachEntryStart( Vec_Ptr_t *, vResult, vUpper, i, 1 )
        {
            nVol = Llb_ManCutVolume( p, vLower, vUpper );
            if nVol <= nVolMax {
                vLower = vUpper;
                continue;
            }

            if fVeryVerbose != 0 {
            Llb_ManCutPrint( p, vLower,  vUpper ); }
            vMinCut = Llb_ManFlowFindBestCut( p, vLower, vUpper, Num );
            if vMinCut == nil {
                if fVeryVerbose != 0 {
                fmt.Printf("Could not break the cut.\n"); }
                if fVeryVerbose != 0 {
                fmt.Printf("\n"); }
                vLower = vUpper;
                continue;
            }

            if fVeryVerbose != 0 {
            Llb_ManCutPrint( p, vMinCut, vUpper ); }
            if fVeryVerbose != 0 {
            Llb_ManCutPrint( p, vLower,  vMinCut ); }
            if fVeryVerbose != 0 {
            fmt.Printf("\n"); }

            break;
        }
        if i == Vec_PtrSize(vResult) {
            break; }
        // insert vMinCut before vUpper
        Vec_PtrPush( vResult, nil );
        for  k = Vec_PtrSize(vResult) - 1; k > i; k--  {
            Vec_PtrWriteEntry( vResult, k, Vec_PtrEntry(vResult, k-1) );
        }
        Vec_PtrWriteEntry( vResult, i, vMinCut );
    }
    if fVerbose != 0 {
        fmt.Printf("Finished computing %d partitions.  ", Vec_PtrSize(vResult) - 1);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        Llb_ManResultPrint( p, vResult );
    }
    return vResult;
}
 
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_BddSetDefaultParams(p *Gia_ParLlb_t) {
    memset( p, 0, sizeof(Gia_ParLlb_t) );
    p.nBddMax       =  1000000;
    p.nIterMax      = 10000000;
    p.nClusterMax   =       20;
    p.nHintDepth    =        0;
    p.HintFirst     =        0;
    p.fUseFlow      =        0;  // use flow 
    p.nVolumeMax    =      100;  // max volume
    p.nVolumeMin    =       30;  // min volume
    p.fReorder      =        1;
    p.fIndConstr    =        0;
    p.fUsePivots    =        0;
    p.fCluster      =        0;
    p.fSchedule     =        0;
    p.fVerbose      =        0;
    p.fVeryVerbose  =        0;
    p.fSilent       =        0;
    p.TimeLimit     =        0;
//    p->TimeLimit     =        0;
    p.TimeLimitGlo  =        0;
    p.TimeTarget    =        0;
    p.iFrame        =       -1;
}

/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManMinCutTest(pAig *Aig_Man_t, Num int32) {
    extern void Llb_BddConstructTest( Aig_Man_t * p, Vec_Ptr_t * vResult );
    extern void Llb_BddExperiment( Aig_Man_t * pInit, Aig_Man_t * pAig, Gia_ParLlb_t * pPars, Vec_Ptr_t * vResult, Vec_Ptr_t * vMaps );
 

//    int fVerbose = 1;
    var Pars Gia_ParLlb_t;
    var pPars *Gia_ParLlb_t = &Pars;
    var vResult *Vec_Ptr_t;//, * vSupps, * vMaps;
    var p *Aig_Man_t;

    Llb_BddSetDefaultParams( pPars );

    p = Aig_ManDupFlopsOnly( pAig );
//Aig_ManShow( p, 0, NULL );
    Aig_ManPrintStats( pAig );
    Aig_ManPrintStats( p );
    Aig_ManFanoutStart( p );

    vResult = Llb_ManComputeCuts( p, Num, 1, 0 );
//    vSupps  = Llb_ManCutSupps( p, vResult );
//    vMaps   = Llb_ManCutMap( p, vResult, vSupps );

//    Llb_BddExperiment( pAig, p, pPars, vResult, vMaps );
    Llb_CoreExperiment( pAig, p, pPars, vResult, 0 );

//    Vec_VecFree( (Vec_Vec_t *)vMaps );
//    Vec_VecFree( (Vec_Vec_t *)vSupps );
    Vec_VecFree( (Vec_Vec_t *)vResult );

    Aig_ManFanoutStop( p );
    Aig_ManCleanMarkAB( p );
    Aig_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

