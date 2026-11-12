package bbr

import (
	"fmt"
	"math"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [bbrReach.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD-based reachability analysis.]

  Synopsis    [Procedures to perform reachability analysis.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bbrReach.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bbr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern Abc_Cex_t * Aig_ManVerifyUsingBddsCountExample( Aig_Man_t * p, DdManager * dd, 
//     DdNode ** pbParts, Vec_Ptr_t * vOnionRings, DdNode * bCubeFirst, 
//     int iOutput, int fVerbose, int fSilent );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default resynthesis parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbr_ManSetDefaultParams(p *Saig_ParBbr_t) {
    memset( p, 0, sizeof(Saig_ParBbr_t) );
    p.TimeLimit     =     0;
    p.nBddMax       = 50000;
    p.nIterMax      =  1000;
    p.fPartition    =     1;
    p.fReorder      =     1;
    p.fReorderImage =     1;
    p.fVerbose      =     0;
    p.fSilent       =     0;
    p.iFrame        =    -1;
}

/**Function********************************************************************

  Synopsis    [Performs the reordering-sensitive step of Extra_bddMove().]

  Description []

  SideEffects []

  SeeAlso     []

******************************************************************************/
func Bbr_bddComputeRangeCube(dd *DdManager, iStart int32, iStop int32) *DdNode {
    var bTemp, bProd *DdNode;
    var i int32;
    assert(iStart <= iStop, "iStart <= iStop");
    assert(iStart >= 0 && iStart <= dd.size, "iStart >= 0 && iStart <= dd.size");
    assert(iStop >= 0  && iStop  <= dd.size, "iStop >= 0  && iStop  <= dd.size");
    bProd = (dd).one;         Cudd_Ref( bProd );
    for  i = iStart; i < iStop; i++  {
        bProd = Cudd_bddAnd( dd, bTemp = bProd, dd.vars[i] );      Cudd_Ref( bProd );
        Cudd_RecursiveDeref( dd, bTemp ); 
    }
    Cudd_Deref( bProd );
    return bProd;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bbr_StopManager(dd *DdManager) {
    var RetValue int32;
    // check for remaining references in the package
    RetValue = Cudd_CheckZeroRef( dd );
    if RetValue > 0 {
        fmt.Printf("\nThe number of referenced nodes = %d\n\n", RetValue); }
//  Cudd_PrintInfo( dd, stdout );
    Cudd_Quit( dd );
}

/**Function*************************************************************

  Synopsis    [Computes the initial state and sets up the variable map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManInitStateVarMap(dd *DdManager, p *Aig_Man_t, fVerbose int32) *DdNode {
    var pbVarsX, pbVarsY **DdNode;
    var bTemp, bProd *DdNode;
    var pLatch *Aig_Obj_t;
    var i int32;

    // set the variable mapping for Cudd_bddVarMap()
    pbVarsX = ABC_ALLOC( DdNode *, dd.size );
    pbVarsY = ABC_ALLOC( DdNode *, dd.size );
    bProd = (dd).one;         Cudd_Ref( bProd );
    Saig_ManForEachLo( p, pLatch, i )
    {
        pbVarsX[i] = dd.vars[ Saig_ManPiNum(p) + i ];
        pbVarsY[i] = dd.vars[ Saig_ManCiNum(p) + i ];
        // get the initial value of the latch
        // inlining of macro Cudd_Not
        bProd = Cudd_bddAnd( dd, bTemp = bProd, ((DdNode *)((ptrint)(pbVarsX[i]) ^ 01)) );      Cudd_Ref( bProd );
        Cudd_RecursiveDeref( dd, bTemp ); 
    }
    Cudd_SetVarMap( dd, pbVarsX, pbVarsY, Saig_ManRegNum(p) );
    ABC_FREE( pbVarsX );
    ABC_FREE( pbVarsY );

    Cudd_Deref( bProd );
    return bProd;
}

/**Function*************************************************************

  Synopsis    [Collects the array of output BDDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCreateOutputs(dd *DdManager, p *Aig_Man_t) **DdNode {
    var pbOutputs **DdNode;
    var pNode *Aig_Obj_t;
    var i int32;
    // compute the transition relation
    pbOutputs = ABC_ALLOC( DdNode *, Saig_ManPoNum(p) );
    Saig_ManForEachPo( p, pNode, i )
    {
        pbOutputs[i] = Aig_ObjGlobalBdd(pNode);  Cudd_Ref( pbOutputs[i] );
    }
    return pbOutputs;
}

/**Function*************************************************************

  Synopsis    [Collects the array of partition BDDs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManCreatePartitions(dd *DdManager, p *Aig_Man_t, fReorder int32, fVerbose int32) **DdNode {
    var pbParts **DdNode;
    var bVar *DdNode;
    var pNode *Aig_Obj_t;
    var i int32;

    // extand the BDD manager to represent NS variables
    assert(dd.size == Saig_ManCiNum(p), "dd.size == Saig_ManCiNum(p)");
    Cudd_bddIthVar( dd, Saig_ManCiNum(p) + Saig_ManRegNum(p) - 1 );

    // enable reordering
    if fReorder != 0 {
        Cudd_AutodynEnable( dd, CUDD_REORDER_SYMM_SIFT );
    } else {
        Cudd_AutodynDisable( dd ); }

    // compute the transition relation
    pbParts = ABC_ALLOC( DdNode *, Saig_ManRegNum(p) );
    Saig_ManForEachLi( p, pNode, i )
    {
        bVar  = Cudd_bddIthVar( dd, Saig_ManCiNum(p) + i );
        pbParts[i] = Cudd_bddXnor( dd, bVar, Aig_ObjGlobalBdd(pNode) );  Cudd_Ref( pbParts[i] );
    }
    // free global BDDs
    Aig_ManFreeGlobalBdds( p, dd );

    // reorder and disable reordering
    if fReorder != 0 {
        if fVerbose != 0 {
            fmt.Fprintf(os.Stdout, "BDD nodes in the partitions before reordering %d.\n", Cudd_SharingSize(pbParts,Saig_ManRegNum(p))); }
        Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 100 );
        Cudd_AutodynDisable( dd );
        if fVerbose != 0 {
            fmt.Fprintf(os.Stdout, "BDD nodes in the partitions after reordering %d.\n", Cudd_SharingSize(pbParts,Saig_ManRegNum(p))); }
    }
    return pbParts;
}

/**Function*************************************************************

  Synopsis    [Computes the set of unreachable states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManComputeReachable(dd *DdManager, p *Aig_Man_t, pbParts **DdNode, bInitial *DdNode, pbOutputs **DdNode, pPars *Saig_ParBbr_t, fCheckOutputs int32) int32 {
    var fInternalReorder int32;
    var pTree *Bbr_ImageTree_t  = nil; // Suppress "might be used uninitialized"
    var pTree2 *Bbr_ImageTree2_t  = nil; // Supprses "might be used uninitialized"
    var bReached, bCubeCs *DdNode;
    var bCurrent *DdNode;
    var bNext *DdNode  = nil; // Suppress "might be used uninitialized"
    var bTemp *DdNode;
    var method Cudd_ReorderingType;
    var i, nIters, nBddSize, status int32;
    var nThreshold int32  = 10000;
    var clk abctime  = Abc_Clock();
    var vOnionRings *Vec_Ptr_t;
    var fixedPoint int32;

    status = Cudd_ReorderingStatus( dd, &method );
    if status != 0 {
        Cudd_AutodynDisable( dd ); }

    // start the image computation
    bCubeCs  = Bbr_bddComputeRangeCube( dd, Saig_ManPiNum(p), Saig_ManCiNum(p) );    Cudd_Ref( bCubeCs );
    if pPars.fPartition {
        pTree = Bbr_bddImageStart( dd, bCubeCs, Saig_ManRegNum(p), pbParts, Saig_ManRegNum(p), dd.vars+Saig_ManCiNum(p), pPars.nBddMax, pPars.fVerbose );
    } else {
        pTree2 = Bbr_bddImageStart2( dd, bCubeCs, Saig_ManRegNum(p), pbParts, Saig_ManRegNum(p), dd.vars+Saig_ManCiNum(p), pPars.fVerbose ); }
    Cudd_RecursiveDeref( dd, bCubeCs );
    if pTree == nil {
        if !pPars.fSilent {
            fmt.Printf("BDDs blew up during qualitification scheduling.  "); }
        return -1;
    }

    if status != 0 {
        Cudd_AutodynEnable( dd, method ); }

    // start the onion rings
    vOnionRings = Vec_PtrAlloc( 1000 );

    // perform reachability analysis
    bCurrent = bInitial;   Cudd_Ref( bCurrent );
    bReached = bInitial;   Cudd_Ref( bReached );
    Vec_PtrPush( vOnionRings, bCurrent );  Cudd_Ref( bCurrent );
    for  nIters = 0; nIters < pPars.nIterMax; nIters++  { 
        // check the runtime limit
        if pPars.TimeLimit && pPars.TimeLimit <= (Abc_Clock()-clk)/int64(time.Second) {
            fmt.Printf("Reached timeout after image computation (%d seconds).\n", pPars.TimeLimit);
            Vec_PtrFree( vOnionRings );
            // undo the image tree
            if pPars.fPartition {
                Bbr_bddImageTreeDelete( pTree );
            } else {
                Bbr_bddImageTreeDelete2( pTree2 ); }
            pPars.iFrame = nIters - 1;
            return -1;
        }

        // compute the next states
        if pPars.fPartition {
            bNext = Bbr_bddImageCompute( pTree, bCurrent );           
        } else {
            bNext = Bbr_bddImageCompute2( pTree2, bCurrent ); }  
        if bNext == nil {
            if !pPars.fSilent {
                fmt.Printf("BDDs blew up during image computation.  "); }
            if pPars.fPartition {
                Bbr_bddImageTreeDelete( pTree );
            } else {
                Bbr_bddImageTreeDelete2( pTree2 ); }
            Vec_PtrFree( vOnionRings );
            pPars.iFrame = nIters - 1;
            return -1;
        }
        Cudd_Ref( bNext );
        Cudd_RecursiveDeref( dd, bCurrent );

        // remap these states into the current state vars
        bNext = Cudd_bddVarMap( dd, bTemp = bNext );                    Cudd_Ref( bNext );
        Cudd_RecursiveDeref( dd, bTemp );
        // check if there are any new states
        if Cudd_bddLeq( dd, bNext, bReached ) {
            fixedPoint = 1;
            break;
        }
        // check the BDD size
        nBddSize = Cudd_DagSize(bNext);
        if nBddSize > pPars.nBddMax {
            break; }
        // check the result
        for  i = 0; i < Saig_ManPoNum(p); i++  {
            // inlining of macro Cudd_Not
            if fCheckOutputs && !Cudd_bddLeq( dd, bNext, ((DdNode *)((ptrint)(pbOutputs[i]) ^ 01)) ) {
                var bIntersect *DdNode;
                bIntersect = Cudd_bddIntersect( dd, bNext, pbOutputs[i] );  Cudd_Ref( bIntersect );
                assert(p.pSeqModel == nil, "p.pSeqModel == nil");
                p.pSeqModel = Aig_ManVerifyUsingBddsCountExample( p, dd, pbParts, 
                    vOnionRings, bIntersect, i, pPars.fVerbose, pPars.fSilent ); 
                Cudd_RecursiveDeref( dd, bIntersect );
                if !pPars.fSilent {
                    Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d. ", i, p.pName, Vec_PtrSize(vOnionRings) ); }
                Cudd_RecursiveDeref( dd, bReached );
                bReached = nil;
                pPars.iFrame = nIters;
                break;
            }
        }
        if i < Saig_ManPoNum(p) {
            break; }
        // get the new states
        // inlining of macro Cudd_Not
        bCurrent = Cudd_bddAnd( dd, bNext, ((DdNode *)((ptrint)(bReached) ^ 01)) );        Cudd_Ref( bCurrent );
        Vec_PtrPush( vOnionRings, bCurrent );  Cudd_Ref( bCurrent );
        // minimize the new states with the reached states
//        bCurrent = Cudd_bddConstrain( dd, bTemp = bCurrent, Cudd_Not(bReached) ); Cudd_Ref( bCurrent );
//        Cudd_RecursiveDeref( dd, bTemp );
        // add to the reached states
        bReached = Cudd_bddOr( dd, bTemp = bReached, bNext );           Cudd_Ref( bReached );
        Cudd_RecursiveDeref( dd, bTemp );
        Cudd_RecursiveDeref( dd, bNext );
        if pPars.fVerbose {
            fmt.Fprintf(os.Stdout, "Frame = %3d. BDD = %5d. ", nIters, nBddSize); }
        if (fInternalReorder != 0) && pPars.fReorder && nBddSize > nThreshold {
            if pPars.fVerbose {
                fmt.Fprintf(os.Stdout, "Reordering... Before = %5d. ", Cudd_DagSize(bReached)); }
            Cudd_ReduceHeap( dd, CUDD_REORDER_SYMM_SIFT, 100 );
            Cudd_AutodynDisable( dd );
            if pPars.fVerbose {
                fmt.Fprintf(os.Stdout, "After = %5d.\r", Cudd_DagSize(bReached)); }
            nThreshold *= 2;
        }
        if pPars.fVerbose {
//            fprintf( stdout, "\r" );
            fmt.Fprintf(os.Stdout, "\n"); }

        if pPars.fVerbose {
            var nMints float64  = Cudd_CountMinterm(dd, bReached, Saig_ManRegNum(p) );
//            Extra_bddPrint( dd, bReached );printf( "\n" );
            fmt.Fprintf(os.Stdout, "Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/math.Pow(float64(2.0), float64(Saig_ManRegNum(p))));
            fflush( os.Stdout ); 
        }

    }
    Cudd_RecursiveDeref( dd, bNext );
    // free the onion rings
    Vec_PtrForEachEntry( DdNode *, vOnionRings, bTemp, i )
        Cudd_RecursiveDeref( dd, bTemp );
    Vec_PtrFree( vOnionRings );
    // undo the image tree
    if pPars.fPartition {
        Bbr_bddImageTreeDelete( pTree );
    } else {
        Bbr_bddImageTreeDelete2( pTree2 ); }
    if bReached == nil {
        return 0; } // proved reachable
    // report the stats
    if pPars.fVerbose {
        var nMints float64  = Cudd_CountMinterm(dd, bReached, Saig_ManRegNum(p) );
        if nIters > pPars.nIterMax || nBddSize > pPars.nBddMax {
            fmt.Fprintf(os.Stdout, "Reachability analysis is stopped after %d frames.\n", nIters);
        } else {
            fmt.Fprintf(os.Stdout, "Reachability analysis completed after %d frames.\n", nIters); }
        fmt.Fprintf(os.Stdout, "Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/math.Pow(float64(2.0), float64(Saig_ManRegNum(p))));
        fflush( os.Stdout );
    }
//ABC_PRB( dd, bReached );
    Cudd_RecursiveDeref( dd, bReached );
    // SPG
    //
    if fixedPoint != 0 {
      if !pPars.fSilent {
        fmt.Printf("The miter is proved unreachable after %d iterations.  ", nIters);
      }
      pPars.iFrame = nIters - 1;
      return 1;
    }
    assert(nIters >= pPars.nIterMax || nBddSize >= pPars.nBddMax, "nIters >= pPars.nIterMax || nBddSize >= pPars.nBddMax");
    if !pPars.fSilent {
      fmt.Printf("Verified only for states reachable in %d frames.  ", nIters); }
    pPars.iFrame = nIters - 1;
    return -1; // undecided
}

/**Function*************************************************************

  Synopsis    [Performs reachability to see if any PO can be asserted.]

  Description []
               
  SideEffects []
 
  SeeAlso     []

***********************************************************************/
func Aig_ManVerifyUsingBdds_int(p *Aig_Man_t, pPars *Saig_ParBbr_t) int32 {
    var fCheckOutputs int32  = !pPars.fSkipOutCheck;
    var dd *DdManager;
    var pbParts, pbOutputs **DdNode;
    var bInitial, bTemp *DdNode;
    var RetValue, i int32;
    var clk abctime  = Abc_Clock();
    var vOnionRings *Vec_Ptr_t;

    assert(Saig_ManRegNum(p) > 0, "Saig_ManRegNum(p) > 0");

    // compute the global BDDs of the latches
    dd = Aig_ManComputeGlobalBdds( p, pPars.nBddMax, 1, pPars.fReorder, pPars.fVerbose );    
    if dd == nil {
        if !pPars.fSilent {
            fmt.Printf("The number of intermediate BDD nodes exceeded the limit (%d).\n", pPars.nBddMax); }
        return -1;
    }
    if pPars.fVerbose {
        fmt.Printf("Shared BDD size is %6d nodes.\n", Cudd_ReadKeys(dd) - Cudd_ReadDead(dd)); }

    // check the runtime limit
    if pPars.TimeLimit && pPars.TimeLimit <= (Abc_Clock()-clk)/int64(time.Second) {
        fmt.Printf("Reached timeout after constructing global BDDs (%d seconds).\n", pPars.TimeLimit);
        Cudd_Quit( dd );
        return -1;
    }

    // start the onion rings
    vOnionRings = Vec_PtrAlloc( 1000 );

    // save outputs
    pbOutputs = Aig_ManCreateOutputs( dd, p );

    // create partitions
    pbParts = Aig_ManCreatePartitions( dd, p, pPars.fReorder, pPars.fVerbose );

    // create the initial state and the variable map
    bInitial  = Aig_ManInitStateVarMap( dd, p, pPars.fVerbose );  Cudd_Ref( bInitial );

    // set reordering
    if pPars.fReorderImage {
        Cudd_AutodynEnable( dd, CUDD_REORDER_SYMM_SIFT ); }

    // check the result
    RetValue = -1;
    for  i = 0; i < Saig_ManPoNum(p); i++  {
        // inlining of macro Cudd_Not
        if fCheckOutputs && !Cudd_bddLeq( dd, bInitial, ((DdNode *)((ptrint)(pbOutputs[i]) ^ 01)) ) {
            var bIntersect *DdNode;
            bIntersect = Cudd_bddIntersect( dd, bInitial, pbOutputs[i] );  Cudd_Ref( bIntersect );
            assert(p.pSeqModel == nil, "p.pSeqModel == nil");
            p.pSeqModel = Aig_ManVerifyUsingBddsCountExample( p, dd, pbParts, 
                vOnionRings, bIntersect, i, pPars.fVerbose, pPars.fSilent ); 
            Cudd_RecursiveDeref( dd, bIntersect );
            if !pPars.fSilent {
                Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d. ", i, p.pName, -1 ); }
            RetValue = 0;
            break;
        }
    }
    // free the onion rings
    Vec_PtrForEachEntry( DdNode *, vOnionRings, bTemp, i )
        Cudd_RecursiveDeref( dd, bTemp );
    Vec_PtrFree( vOnionRings );
    // explore reachable states
    if RetValue == -1 {
        RetValue = Aig_ManComputeReachable( dd, p, pbParts, bInitial, pbOutputs, pPars, fCheckOutputs ); } 

    // cleanup
    Cudd_RecursiveDeref( dd, bInitial );
    for  i = 0; i < Saig_ManRegNum(p); i++  {
        Cudd_RecursiveDeref( dd, pbParts[i] );
    }
    ABC_FREE( pbParts );
    for  i = 0; i < Saig_ManPoNum(p); i++  {
        Cudd_RecursiveDeref( dd, pbOutputs[i] );
    }
    ABC_FREE( pbOutputs );
//    if ( RetValue == -1 )
        Cudd_Quit( dd );
//    else
//        Bbr_StopManager( dd );

    // report the runtime
    if !pPars.fSilent {
    ABC_PRT( "Time", Abc_Clock() - clk );
    fflush( os.Stdout );
    }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Performs reachability to see if any PO can be asserted.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManVerifyUsingBdds(pInit *Aig_Man_t, pPars *Saig_ParBbr_t) int32 {
    var pCexOld, pCexNew *Abc_Cex_t;
    var p *Aig_Man_t;
    var pObj *Aig_Obj_t;
    var vInputMap *Vec_Int_t;
    var i, k, Entry, iBitOld, iBitNew, RetValue int32;
//    pPars->fVerbose = 1;
    // check if there are PIs without fanout
    Saig_ManForEachPi( pInit, pObj, i )
        if ( Aig_ObjRefs(pObj) == 0 )
            break;
    if i == Saig_ManPiNum(pInit) {
        return Aig_ManVerifyUsingBdds_int( pInit, pPars ); }
    // create new AIG
    p = Aig_ManDupTrim( pInit );
    assert(Aig_ManCiNum(p) < Aig_ManCiNum(pInit), "Aig_ManCiNum(p) < Aig_ManCiNum(pInit)");
    assert(Aig_ManRegNum(p) == Aig_ManRegNum(pInit), "Aig_ManRegNum(p) == Aig_ManRegNum(pInit)");
    RetValue = Aig_ManVerifyUsingBdds_int( p, pPars );
    if RetValue != 0 {
        Aig_ManStop( p );
        return RetValue;
    }
    // the problem is satisfiable - remap the pattern
    pCexOld = p.pSeqModel;
    assert(pCexOld != nil, "pCexOld != nil");
    // create input map
    vInputMap = Vec_IntAlloc( Saig_ManPiNum(pInit) );
    Saig_ManForEachPi( pInit, pObj, i )
        if ( pObj.pData != nil )
            Vec_IntPush( vInputMap, Aig_ObjCioId((Aig_Obj_t *)pObj.pData) );
        else
            Vec_IntPush( vInputMap, -1 );
    // create new pattern
    pCexNew = Abc_CexAlloc( Saig_ManRegNum(pInit), Saig_ManPiNum(pInit), pCexOld.iFrame+1 );
    pCexNew.iFrame = pCexOld.iFrame;
    pCexNew.iPo    = pCexOld.iPo;
    // copy the bit-data
    for  iBitOld = 0; iBitOld < pCexOld.nRegs; iBitOld++  {
        if Abc_InfoHasBit( pCexOld.pData, iBitOld ) {
            Abc_InfoSetBit( pCexNew.pData, iBitOld ); }
    }
    // copy the primary input data
    iBitNew = iBitOld;
    for  i = 0; i <= pCexNew.iFrame; i++  {
        Vec_IntForEachEntry( vInputMap, Entry, k )
        {
            if Entry == -1 {
                continue; }
            if Abc_InfoHasBit( pCexOld.pData, iBitOld + Entry ) {
                Abc_InfoSetBit( pCexNew.pData, iBitNew + k ); }
        }
        iBitOld += Saig_ManPiNum(p);
        iBitNew += Saig_ManPiNum(pInit);
    }
    assert(iBitOld < iBitNew, "iBitOld < iBitNew");
    assert(iBitOld == pCexOld.nBits, "iBitOld == pCexOld.nBits");
    assert(iBitNew == pCexNew.nBits, "iBitNew == pCexNew.nBits");
    Vec_IntFree( vInputMap );
    pInit.pSeqModel = pCexNew;
    Aig_ManStop( p );
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

