package reo

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [reoCore.c]

  PackageName [REO: A specialized DD reordering engine.]

  Synopsis    [Implementation of the core reordering procedure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 15, 2002.]

  Revision    [$Id: reoCore.c,v 1.0 2002/15/10 03:00:00 alanmi Exp $]

***********************************************************************/

//#include "reo.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int  reoRecursiveDeref( reo_unit * pUnit );

// static int  reoCheckZeroRefs( reo_plane * pPlane );

// static int  reoCheckLevels( reo_man * p );


var s_AplBefore float64;
var s_AplAfter float64;

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoReorderArray(p *reo_man, dd *DdManager, Funcs []*DdNode, FuncsRes []*DdNode, nFuncs int32, pOrder *int32) {
    var Counter, i int32;

    // set the initial parameters
    p.dd     = dd;
    p.pOrder = pOrder;
    p.nTops  = nFuncs;
    // get the initial number of nodes
    p.nNodesBeg = Cudd_SharingSize( Funcs, nFuncs );     
    // resize the internal data structures of the manager if necessary
    // inlining of macro ddMax
    reoResizeStructures( p, (((dd.sizeZ) > (dd.size)) ? (dd.sizeZ) : (dd.size)), p.nNodesBeg, nFuncs );
    // compute the support
    p.pSupp = Extra_VectorSupportArray( dd, Funcs, nFuncs, p.pSupp );
    // get the number of support variables
    p.nSupp = 0;
    for  i = 0; i < dd.size; i++  {
        p.nSupp += p.pSupp[i];
    }

    // if it is the constant function, no need to reorder
    if p.nSupp == 0 {
        for  i = 0; i < nFuncs; i++  {
            FuncsRes[i] = Funcs[i]; Cudd_Ref( FuncsRes[i] );
        }
        return;
    }

    // create the internal variable maps
    // go through variable levels in the manager
    Counter = 0;
    for  i = 0; i < dd.size; i++  {
        if p.pSupp[ dd.invperm[i] ] {
            p.pMapToPlanes[ dd.invperm[i] ] = Counter;
            p.pMapToDdVarsOrig[Counter]      = dd.invperm[i];
            if !p.fRemapUp {
                p.pMapToDdVarsFinal[Counter] = dd.invperm[i];
            } else {
                p.pMapToDdVarsFinal[Counter] = dd.invperm[Counter]; }
            p.pOrderInt[Counter]        = Counter;
            Counter++;
        }
    }

    // set the initial parameters
    p.nUnitsUsed = 0;
    p.nNodesCur  = 0;
    p.fThisIsAdd = 0;
    p.Signature++;
    // transfer the function from the CUDD package into REO"s internal data structure
    for  i = 0; i < nFuncs; i++  {
        p.pTops[i] = reoTransferNodesToUnits_rec( p, Funcs[i] );
    }
    assert(p.nNodesBeg == p.nNodesCur, "p.nNodesBeg == p.nNodesCur");

    if !p.fThisIsAdd && p.fMinWidth {
        fmt.Printf("An important message from the REO reordering engine:\n");
        fmt.Printf("The BDD given to the engine for reordering contains complemented edges.\n");
        fmt.Printf("Currently, such BDDs cannot be reordered for the minimum width.\n");
        fmt.Printf("Therefore, minimization for the number of BDD nodes is performed.\n");
        fflush( os.Stdout );
        p.fMinApl   = 0;
        p.fMinWidth = 0;
    }

    if p.fMinWidth {
        reoProfileWidthStart(p);
    } else if p.fMinApl {
        reoProfileAplStart(p);
    } else { 
        reoProfileNodesStart(p); }

    if p.fVerbose {
        fmt.Printf("INITIAL:\n");
        if p.fMinWidth {
            reoProfileWidthPrint(p);
        } else if p.fMinApl {
            reoProfileAplPrint(p);
        } else {
            reoProfileNodesPrint(p); }
    }
 
    ///////////////////////////////////////////////////////////////////
    // performs the reordering
    p.nSwaps   = 0;
    p.nNISwaps = 0;
    for  i = 0; i < p.nIters; i++  {
        reoReorderSift( p );
        // print statistics after each iteration
        if p.fVerbose {
            fmt.Printf("ITER #%d:\n", i+1);
            if p.fMinWidth {
                reoProfileWidthPrint(p);
            } else if p.fMinApl {
                reoProfileAplPrint(p);
            } else {
                reoProfileNodesPrint(p); }
        }
        // if the cost function did not change, stop iterating
        if p.fMinWidth {
            p.nWidthEnd = p.nWidthCur;
            assert(p.nWidthEnd <= p.nWidthBeg, "p.nWidthEnd <= p.nWidthBeg");
            if p.nWidthEnd == p.nWidthBeg {
                break; }
        
        } else if p.fMinApl {
            p.nAplEnd = p.nAplCur;
            assert(p.nAplEnd <= p.nAplBeg, "p.nAplEnd <= p.nAplBeg");
            if p.nAplEnd == p.nAplBeg {
                break; }
        
        } else {
            p.nNodesEnd = p.nNodesCur;
            assert(p.nNodesEnd <= p.nNodesBeg, "p.nNodesEnd <= p.nNodesBeg");
            if p.nNodesEnd == p.nNodesBeg {
                break; }
        }
    }
    assert((reoCheckLevels( p ) != 0), "reoCheckLevels( p )");
    ///////////////////////////////////////////////////////////////////

s_AplBefore = p.nAplBeg;
s_AplAfter  = p.nAplEnd;

    // set the initial parameters
    p.nRefNodes  = 0;
    p.nNodesCur  = 0;
    p.Signature++;
    // transfer the BDDs from REO's internal data structure to CUDD
    for  i = 0; i < nFuncs; i++  {
        FuncsRes[i] = reoTransferUnitsToNodes_rec( p, p.pTops[i] ); Cudd_Ref( FuncsRes[i] );
    }
    // undo the DDs referenced for storing in the cache
    for  i = 0; i < p.nRefNodes; i++  {
        Cudd_RecursiveDeref( dd, p.pRefNodes[i] );
    }
    // verify zero refs of the terminal nodes
    for  i = 0; i < nFuncs; i++  {
        assert((reoRecursiveDeref( p.pTops[i] ) != 0), "reoRecursiveDeref( p.pTops[i] )");
    }
    assert((reoCheckZeroRefs( &(p.pPlanes[p.nSupp]) ) != 0), "reoCheckZeroRefs( &(p.pPlanes[p.nSupp]) )");

    // prepare the variable map to return to the user
    if p.pOrder {
        // i is the current level in the planes data structure
        // p->pOrderInt[i] is the original level in the planes data structure
        // p->pMapToDdVarsOrig[i] is the variable, into which we remap when we construct the BDD from planes
        // p->pMapToDdVarsOrig[ p->pOrderInt[i] ] is the original BDD variable corresponding to this level
        // Therefore, p->pOrder[ p->pMapToDdVarsFinal[i] ] = p->pMapToDdVarsOrig[ p->pOrderInt[i] ]
        // creates the permutation, which remaps the resulting BDD variable into the original BDD variable
        for  i = 0; i < p.nSupp; i++  {
            p.pOrder[ p.pMapToDdVarsFinal[i] ] = p.pMapToDdVarsOrig[ p.pOrderInt[i] ];
        } 
    }

    if p.fVerify {
        var fVerification int32;
        var FuncRemapped *DdNode;
        var pOrder *int32;

        if p.pOrder == nil {
            pOrder = ABC_ALLOC( int, p.nSupp );
            for  i = 0; i < p.nSupp; i++  {
                pOrder[ p.pMapToDdVarsFinal[i] ] = p.pMapToDdVarsOrig[ p.pOrderInt[i] ];
            } 
        
        } else {
            pOrder = p.pOrder; }

        fVerification = 1;
        for  i = 0; i < nFuncs; i++  {
            // verify the result
            if p.fThisIsAdd {
                FuncRemapped = Cudd_addPermute( dd, FuncsRes[i], pOrder );
            } else {
                FuncRemapped = Cudd_bddPermute( dd, FuncsRes[i], pOrder ); }
            Cudd_Ref( FuncRemapped );

            if FuncRemapped != Funcs[i] {
                fVerification = 0;
                fmt.Printf("REO: Internal verification has failed!\n");
                fflush( os.Stdout );
            }
            Cudd_RecursiveDeref( dd, FuncRemapped );
        }
        if fVerification != 0 {
            fmt.Printf("REO: Internal verification is okay!\n"); }

        if p.pOrder == nil {
            ABC_FREE( pOrder ); }
    }

    // recycle the data structure
    for  i = 0; i <= p.nSupp; i++  {
        reoUnitsRecycleUnitList( p, p.pPlanes + i );
    }
}

/**Function*************************************************************

  Synopsis    [Resizes the internal manager data structures.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoResizeStructures(p *reo_man, nDdVarsMax int32, nNodesMax int32, nFuncs int32) {
    // resize data structures depending on the number of variables in the DD manager
    if p.nSuppAlloc == 0 {
        p.pSupp             = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pOrderInt         = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pMapToPlanes      = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pMapToDdVarsOrig  = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pMapToDdVarsFinal = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pPlanes           = ABC_CALLOC( reo_plane, nDdVarsMax + 1 );
        p.pVarCosts         = ABC_ALLOC( double,     nDdVarsMax + 1 );
        p.pLevelOrder       = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.nSuppAlloc        = nDdVarsMax + 1;
    
    } else if p.nSuppAlloc < nDdVarsMax {
        ABC_FREE( p.pSupp );
        ABC_FREE( p.pOrderInt );
        ABC_FREE( p.pMapToPlanes );
        ABC_FREE( p.pMapToDdVarsOrig );
        ABC_FREE( p.pMapToDdVarsFinal );
        ABC_FREE( p.pPlanes );
        ABC_FREE( p.pVarCosts );
        ABC_FREE( p.pLevelOrder );

        p.pSupp             = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pOrderInt         = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pMapToPlanes      = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pMapToDdVarsOrig  = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pMapToDdVarsFinal = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.pPlanes           = ABC_CALLOC( reo_plane, nDdVarsMax + 1 );
        p.pVarCosts         = ABC_ALLOC( double,     nDdVarsMax + 1 );
        p.pLevelOrder       = ABC_ALLOC( int,        nDdVarsMax + 1 );
        p.nSuppAlloc        = nDdVarsMax + 1;
    }

    // resize the data structures depending on the number of nodes
    if p.nRefNodesAlloc == 0 {
        p.nNodesMaxAlloc  = nNodesMax;
        p.nTableSize      = 3*nNodesMax + 1;
        p.nRefNodesAlloc  = 3*nNodesMax + 1;
        p.nMemChunksAlloc = (10*nNodesMax + 1)/REO_CHUNK_SIZE + 1;

        p.HTable          = ABC_CALLOC( reo_hash,  p.nTableSize );
        p.pRefNodes       = ABC_ALLOC( DdNode *,   p.nRefNodesAlloc );
        p.pWidthCofs      = ABC_ALLOC( reo_unit *, p.nRefNodesAlloc );
        p.pMemChunks      = ABC_ALLOC( reo_unit *, p.nMemChunksAlloc );
    
    } else if p.nNodesMaxAlloc < nNodesMax {
        var pTemp **reo_unit;
        var nMemChunksAllocPrev int32  = p.nMemChunksAlloc;

        p.nNodesMaxAlloc  = nNodesMax;
        p.nTableSize      = 3*nNodesMax + 1;
        p.nRefNodesAlloc  = 3*nNodesMax + 1;
        p.nMemChunksAlloc = (10*nNodesMax + 1)/REO_CHUNK_SIZE + 1;

        ABC_FREE( p.HTable );
        ABC_FREE( p.pRefNodes );
        ABC_FREE( p.pWidthCofs );
        p.HTable          = ABC_CALLOC( reo_hash,    p.nTableSize );
        p.pRefNodes       = ABC_ALLOC(  DdNode *,    p.nRefNodesAlloc );
        p.pWidthCofs      = ABC_ALLOC(  reo_unit *,  p.nRefNodesAlloc );
        // p->pMemChunks should be reallocated because it contains pointers currently in use
        pTemp              = ABC_ALLOC(  reo_unit *,  p.nMemChunksAlloc );
        memmove( pTemp, p.pMemChunks, sizeof(reo_unit *) * nMemChunksAllocPrev );
        ABC_FREE( p.pMemChunks );
        p.pMemChunks      = pTemp;
    }

    // resize the data structures depending on the number of functions
    if p.nTopsAlloc == 0 {
        p.pTops      = ABC_ALLOC( reo_unit *, nFuncs );
        p.nTopsAlloc = nFuncs;
    
    } else if p.nTopsAlloc < nFuncs {
        ABC_FREE( p.pTops );
        p.pTops      = ABC_ALLOC( reo_unit *, nFuncs );
        p.nTopsAlloc = nFuncs;
    }
}


/**Function*************************************************************

  Synopsis    [Dereferences units the data structure after reordering.]

  Description [This function is only useful for debugging.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoRecursiveDeref(pUnit *reo_unit) int32 {
    var pUnitR *reo_unit;
    // inlining of macro Unit_Regular
    pUnitR = ((reo_unit *)((ABC_PTRUINT_T)(pUnit) & ~01));
    pUnitR.n--;
    // inlining of macro Unit_IsConstant
    if  int32(((pUnitR).lev == REO_CONST_LEVEL)) != 0 {
        return 1; }
    if pUnitR.n == 0 {
        reoRecursiveDeref( pUnitR.pE );
        reoRecursiveDeref( pUnitR.pT );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Checks the zero references for the given plane.]

  Description [This function is only useful for debugging.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoCheckZeroRefs(pPlane *reo_plane) int32 {
    var pUnit *reo_unit;
    for  pUnit = pPlane.pHead; (pUnit != nil); pUnit = pUnit.Next  {
        if pUnit.n != 0 {
            assert(false, "0");
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Checks the zero references for the given plane.]

  Description [This function is only useful for debugging.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func reoCheckLevels(p *reo_man) int32 {
    var pUnit *reo_unit;
    var i int32;

    for  i = 0; i < p.nSupp; i++  {
        // there are some nodes left on each level
        assert(p.pPlanes[i].statsNodes, "p.pPlanes[i].statsNodes");
        for  pUnit = p.pPlanes[i].pHead; (pUnit != nil); pUnit = pUnit.Next  {
            // the level is properly set
            assert(pUnit.lev == i, "pUnit.lev == i");
        }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                         END OF FILE                              ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

