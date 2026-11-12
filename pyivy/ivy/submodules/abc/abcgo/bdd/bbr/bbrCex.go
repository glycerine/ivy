package bbr

import "fmt"

/**CFile****************************************************************

  FileName    [bbrCex.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD-based reachability analysis.]

  Synopsis    [Procedures to derive a satisfiable counter-example.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bbrCex.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bbr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern DdNode * Bbr_bddComputeRangeCube( DdManager * dd, int iStart, int iStop );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Derives the counter-example using the set of reached states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_ManVerifyUsingBddsCountExample(p *Aig_Man_t, dd *DdManager, pbParts **DdNode, vOnionRings *Vec_Ptr_t, bCubeFirst *DdNode, iOutput int32, fVerbose int32, fSilent int32) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Aig_Obj_t;
    var pTree *Bbr_ImageTree_t;
    var bCubeNs, bState, bImage *DdNode;
    var bTemp, bVar, bRing *DdNode;
    var i, v, RetValue, nPiOffset int32;
    var pValues string;
    var clk abctime  = Abc_Clock();
//printf( "\nDeriving counter-example.\n" );

    // allocate room for the counter-example
    pCex = Abc_CexAlloc( Saig_ManRegNum(p), Saig_ManPiNum(p), Vec_PtrSize(vOnionRings)+1 );
    pCex.iFrame = Vec_PtrSize(vOnionRings);
    pCex.iPo = iOutput;
    nPiOffset = Saig_ManRegNum(p) + Saig_ManPiNum(p) * Vec_PtrSize(vOnionRings);

    // create the cube of NS variables
    bCubeNs  = Bbr_bddComputeRangeCube( dd, Saig_ManCiNum(p), Saig_ManCiNum(p)+Saig_ManRegNum(p) );    Cudd_Ref( bCubeNs );
    pTree = Bbr_bddImageStart( dd, bCubeNs, Saig_ManRegNum(p), pbParts, Saig_ManCiNum(p), dd.vars, 100000000, fVerbose );
    Cudd_RecursiveDeref( dd, bCubeNs );
    if pTree == nil {
        if !(fSilent != 0) {
            fmt.Printf("BDDs blew up during qualitification scheduling.  "); }
        return nil;
    }

    // allocate room for the cube
    pValues = ABC_ALLOC( char, dd.size );

    // get the last cube
    RetValue = Cudd_bddPickOneCube( dd, bCubeFirst, pValues );
    assert(RetValue, "RetValue");

    // write PIs of counter-example
    Saig_ManForEachPi( p, pObj, i )
        if ( pValues[i] == 1 )
            Abc_InfoSetBit( pCex.pData, nPiOffset + i );
    nPiOffset -= Saig_ManPiNum(p);

    // write state in terms of NS variables
    bState = (dd).one; Cudd_Ref( bState );
    Saig_ManForEachLo( p, pObj, i )
    {
        // inlining of macro Cudd_NotCond
        bVar = ((DdNode *)((ptrint)(dd.vars[Saig_ManCiNum(p)+i]) ^ (pValues[Saig_ManPiNum(p)+i] != 1)));
        bState = Cudd_bddAnd( dd, bTemp = bState, bVar );  Cudd_Ref( bState );
        Cudd_RecursiveDeref( dd, bTemp ); 
    }

    // perform backward analysis
    Vec_PtrForEachEntryReverse( DdNode *, vOnionRings, bRing, v )
    { 
        // compute the next states
        bImage = Bbr_bddImageCompute( pTree, bState );           
        if bImage == nil {
            Cudd_RecursiveDeref( dd, bState );
            if !(fSilent != 0) {
                fmt.Printf("BDDs blew up during image computation.  "); }
            Bbr_bddImageTreeDelete( pTree );
            ABC_FREE( pValues );
            return nil;
        }
        Cudd_Ref( bImage );
        Cudd_RecursiveDeref( dd, bState );

        // intersect with the previous set
        bImage = Cudd_bddAnd( dd, bTemp = bImage, bRing );  Cudd_Ref( bImage );
        Cudd_RecursiveDeref( dd, bTemp );

        // find any assignment of the BDD
        RetValue = Cudd_bddPickOneCube( dd, bImage, pValues );
        assert(RetValue, "RetValue");
        Cudd_RecursiveDeref( dd, bImage );

        // write PIs of counter-example
        Saig_ManForEachPi( p, pObj, i )
            if ( pValues[i] == 1 )
                Abc_InfoSetBit( pCex.pData, nPiOffset + i );
        nPiOffset -= Saig_ManPiNum(p);

        // check that we get the init state
        if v == 0 {
            Saig_ManForEachLo( p, pObj, i )
                assert(pValues[Saig_ManPiNum(p)+i] == 0, "pValues[Saig_ManPiNum(p)+i] == 0");
            break;
        }

        // write state in terms of NS variables
        bState = (dd).one; Cudd_Ref( bState );
        Saig_ManForEachLo( p, pObj, i )
        {
            // inlining of macro Cudd_NotCond
            bVar = ((DdNode *)((ptrint)(dd.vars[Saig_ManCiNum(p)+i]) ^ (pValues[Saig_ManPiNum(p)+i] != 1)));
            bState = Cudd_bddAnd( dd, bTemp = bState, bVar );  Cudd_Ref( bState );
            Cudd_RecursiveDeref( dd, bTemp ); 
        }
    }
    // cleanup
    Bbr_bddImageTreeDelete( pTree );
    ABC_FREE( pValues );
    // verify the counter example
    if Vec_PtrSize(vOnionRings) < 1000 {
    RetValue = Saig_ManVerifyCex( p, pCex );
    if RetValue == 0 && !(fSilent != 0) {
        fmt.Printf("Aig_ManVerifyUsingBdds(): Counter-example verification has FAILED.\n"); }
    }
    if (fVerbose != 0) && !(fSilent != 0) {
    ABC_PRT( "Counter-example generation time", Abc_Clock() - clk );
    }
    return pCex;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

