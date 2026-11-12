package msat

import "fmt"

/**CFile****************************************************************

  FileName    [msatOrder.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [The manager of variable assignment.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatOrder.c,v 1.0 2005/05/30 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


/* 
The J-boundary (justification boundary) is defined as a set of unassigned 
variables belonging to the cone of interest, such that for each of them,
there exist an adjacent assigned variable in the cone of interest.
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Msat_OrderVar_t =  Msat_OrderVar_t_;
type Msat_OrderRing_t = Msat_OrderRing_t_;

// the variable data structure
type Msat_OrderVar_t_ struct {
    pNext     *Msat_OrderVar_t;
    pPrev     *Msat_OrderVar_t;
    Num                  int32;
}

// the ring of variables data structure (J-boundary)
type Msat_OrderRing_t_ struct {
    pRoot     *Msat_OrderVar_t; 
    nItems                 int32;
}

// the variable package data structure
type Msat_Order_t_ struct {
    pSat       *Msat_Solver_t;         // the SAT solver 
    pVars     *Msat_OrderVar_t;        // the storage for variables
    nVarsAlloc                 int32;   // the number of variables allocated
    rVars     Msat_OrderRing_t;        // the J-boundary as a ring of variables
}

//The solver can communicate to the variable order the following parts:
//- the array of current assignments (pSat->pAssigns)
//- the array of variable activities (pSat->pdActivity)
//- the array of variables currently in the cone (pSat->vConeVars)
//- the array of arrays of variables adjucent to each(pSat->vAdjacents)

// Original macro:
// #define Msat_OrderVarIsInBoundary( p, i )   ((p)->pVars[i].pNext)

// Original macro:
// #define Msat_OrderVarIsAssigned( p, i )     ((p)->pSat->pAssigns[i] != MSAT_VAR_UNASSIGNED)

// Original macro:
// #define Msat_OrderVarIsUsedInCone( p, i )   ((p)->pSat->vVarsUsed->pArray[i])


// iterator through the entries in J-boundary
// Original macro:
// #define Msat_OrderRingForEachEntry( pRing, pVar, pNext )         \
//     for ( pVar = pRing,                                         \
//           pNext = pVar? pVar->pNext : NULL;                     \
//           pVar;                                                 \
//           pVar = (pNext != pRing)? pNext : NULL,                \
//           pNext = pVar? pVar->pNext : NULL )


// static void Msat_OrderRingAddLast( Msat_OrderRing_t * pRing, Msat_OrderVar_t * pVar );

// static void Msat_OrderRingRemove( Msat_OrderRing_t * pRing, Msat_OrderVar_t * pVar );


var timeSelect int64;
var timeAssign int64;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the ordering structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderAlloc(pSat *Msat_Solver_t) *Msat_Order_t {
    var p *Msat_Order_t;
    p = ALLOC( Msat_Order_t, 1 );
    memset( p, 0, sizeof(Msat_Order_t) );
    p.pSat = pSat;
    Msat_OrderSetBounds( p, pSat.nVarsAlloc );
    return p;
}

/**Function*************************************************************

  Synopsis    [Sets the bound of the ordering structure.]

  Description [Should be called whenever the SAT solver is resized.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderSetBounds(p *Msat_Order_t, nVarsMax int32) {
    var i int32;
    // add variables if they are missing
    if p.nVarsAlloc < nVarsMax {
        p.pVars = REALLOC( Msat_OrderVar_t, p.pVars, nVarsMax );
        for  i = p.nVarsAlloc; i < nVarsMax; i++  {
            p.pVars[i].pNext = p.pVars[i].pPrev = nil;
            p.pVars[i].Num = i;
        }
        p.nVarsAlloc = nVarsMax;
    }
}

/**Function*************************************************************

  Synopsis    [Cleans the ordering structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderClean(p *Msat_Order_t, vCone *Msat_IntVec_t) {
    var pVar, pNext *Msat_OrderVar_t;
    // quickly undo the ring
    // inlining of macro Msat_OrderRingForEachEntry
    for  pVar = p.rVars.pRoot,                                                   pNext = pVar? pVar.pNext : nil;                               (pVar != nil);                                                           pVar = (pNext != p.rVars.pRoot)? pNext : nil,                          pNext = pVar? pVar.pNext : nil  {
        pVar.pNext = pVar.pPrev = nil;
    }
    p.rVars.pRoot  = nil;
    p.rVars.nItems = 0;
}

/**Function*************************************************************

  Synopsis    [Checks that the J-boundary is okay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderCheck(p *Msat_Order_t) int32 {
    var pVar, pNext *Msat_OrderVar_t;
    var vRound *Msat_IntVec_t;
    var pRound *int32;
    var nRound int32;
    var pVars *int32;
    var nVars int32;
    var i int32;
    var k int32;
    var Counter int32;

    // go through all the variables in the boundary
    // inlining of macro Msat_OrderRingForEachEntry
    for  pVar = p.rVars.pRoot,                                                   pNext = pVar? pVar.pNext : nil;                               (pVar != nil);                                                           pVar = (pNext != p.rVars.pRoot)? pNext : nil,                          pNext = pVar? pVar.pNext : nil  {
        // inlining of macro Msat_OrderVarIsAssigned
        assert(!((p).pSat.pAssigns[pVar.Num] != MSAT_VAR_UNASSIGNED), "!/*__starting_point_inline_macro:Msat_OrderVarIsAssigned__*/((p).pSat.pAssigns[pVar.Num] != MSAT_VAR_UNASSIGNED)");
        // go though all the variables in the neighborhood
        // and check that it is true that there is least one assigned
        vRound = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( p.pSat.vAdjacents, pVar.Num );
        nRound = Msat_IntVecReadSize( vRound );
        pRound = Msat_IntVecReadArray( vRound );
        for  i = 0; i < nRound; i++  {
            // inlining of macro Msat_OrderVarIsUsedInCone
            if !((p).pSat.vVarsUsed.pArray[pRound[i]]) {
                continue; }
            // inlining of macro Msat_OrderVarIsAssigned
            if  (p).pSat.pAssigns[pRound[i]] != MSAT_VAR_UNASSIGNED {
                break; }
        }
//        assert( i != nRound );
//        if ( i == nRound )
//            return 0;
        if i == nRound {
            Counter++; }
    }
    if Counter > 0 {
        fmt.Printf("%d(%d) ", Counter, p.rVars.nItems); }

    // we may also check other unassigned variables in the cone
    // to make sure that if they are not in J-boundary, 
    // then they do not have an assigned neighbor
    nVars = Msat_IntVecReadSize( p.pSat.vConeVars );
    pVars = Msat_IntVecReadArray( p.pSat.vConeVars );
    for  i = 0; i < nVars; i++  {
        // inlining of macro Msat_OrderVarIsUsedInCone
        assert(((p).pSat.vVarsUsed.pArray[pVars[i]]), "/*__starting_point_inline_macro:Msat_OrderVarIsUsedInCone__*/((p).pSat.vVarsUsed.pArray[pVars[i]])");
        // skip assigned vars, vars in the boundary, and vars not used in the cone
        // inlining of macro Msat_OrderVarIsAssigned
        if  ((p).pSat.pAssigns[pVars[i]] != MSAT_VAR_UNASSIGNED) || 
             // inlining of macro Msat_OrderVarIsInBoundary
             ((p).pVars[pVars[i]].pNext) {
            continue; }
        // make sure, it does not have assigned neighbors
        vRound = (Msat_IntVec_t *)Msat_ClauseVecReadEntry( p.pSat.vAdjacents, pVars[i] );
        nRound = Msat_IntVecReadSize( vRound );
        pRound = Msat_IntVecReadArray( vRound );
        for  k = 0; k < nRound; k++  {
            // inlining of macro Msat_OrderVarIsUsedInCone
            if !((p).pSat.vVarsUsed.pArray[pRound[k]]) {
                continue; }
            // inlining of macro Msat_OrderVarIsAssigned
            if  (p).pSat.pAssigns[pRound[k]] != MSAT_VAR_UNASSIGNED {
                break; }
        }
//        assert( k == nRound );
//        if ( k != nRound )
//            return 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Frees the ordering structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderFree(p *Msat_Order_t) {
    free( p.pVars );
    free( p );
}

/**Function*************************************************************

  Synopsis    [Selects the next variable to assign.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderVarSelect(p *Msat_Order_t) int32 {
    var pVar, pNext, pVarBest *Msat_OrderVar_t;
    var pdActs *float64  = p.pSat.pdActivity;
    var dfActBest float64;
//    clock_t clk = clock();

    pVarBest = nil;
    dfActBest = -1.0;
    // inlining of macro Msat_OrderRingForEachEntry
    for  pVar = p.rVars.pRoot,                                                   pNext = pVar? pVar.pNext : nil;                               (pVar != nil);                                                           pVar = (pNext != p.rVars.pRoot)? pNext : nil,                          pNext = pVar? pVar.pNext : nil  {
        if dfActBest < pdActs[pVar.Num] {
            dfActBest = pdActs[pVar.Num];
            pVarBest  = pVar;
        }
    }
//timeSelect += clock() - clk;
//timeAssign += clock() - clk;

//if ( pVarBest && pVarBest->Num % 1000 == 0 )
//printf( "%d ", p->rVars.nItems );

//    Msat_OrderCheck( p );
    if pVarBest != nil {
        // inlining of macro Msat_OrderVarIsUsedInCone
        assert(((p).pSat.vVarsUsed.pArray[pVarBest.Num]), "/*__starting_point_inline_macro:Msat_OrderVarIsUsedInCone__*/((p).pSat.vVarsUsed.pArray[pVarBest.Num])");
        return pVarBest.Num;
    }
    return MSAT_ORDER_UNKNOWN;
}

/**Function*************************************************************

  Synopsis    [Updates J-boundary when the variable is assigned.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderVarAssigned(p *Msat_Order_t, Var int32) {
    var vRound *Msat_IntVec_t;
    var i int32;//, clk = clock();

    // make sure the variable is in the boundary
    assert(Var < p.nVarsAlloc, "Var < p.nVarsAlloc");
    // if it is not in the boundary (initial decision, random decision), do not remove
    // inlining of macro Msat_OrderVarIsInBoundary
    if  (p).pVars[Var].pNext {
        Msat_OrderRingRemove( &p.rVars, &p.pVars[Var] ); }
    // add to the boundary those neighbors that are (1) unassigned, (2) not in boundary
    // because for them we know that there is a variable (Var) which is assigned
    vRound = (Msat_IntVec_t *)p.pSat.vAdjacents.pArray[Var];
    for  i = 0; i < vRound.nSize; i++  {
        // inlining of macro Msat_OrderVarIsUsedInCone
        if !((p).pSat.vVarsUsed.pArray[vRound.pArray[i]]) {
            continue; }
        // inlining of macro Msat_OrderVarIsAssigned
        if  (p).pSat.pAssigns[vRound.pArray[i]] != MSAT_VAR_UNASSIGNED {
            continue; }
        // inlining of macro Msat_OrderVarIsInBoundary
        if  (p).pVars[vRound.pArray[i]].pNext {
            continue; }
        Msat_OrderRingAddLast( &p.rVars, &p.pVars[vRound.pArray[i]] );
    }
//timeSelect += clock() - clk;
//    Msat_OrderCheck( p );
}

/**Function*************************************************************

  Synopsis    [Updates the order after a variable is unassigned.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderVarUnassigned(p *Msat_Order_t, Var int32) {
    var vRound, vRound2 *Msat_IntVec_t;
    var i, k int32;//, clk = clock();

    // make sure the variable is not in the boundary
    assert(Var < p.nVarsAlloc, "Var < p.nVarsAlloc");
    // inlining of macro Msat_OrderVarIsInBoundary
    assert(!((p).pVars[Var].pNext), "!/*__starting_point_inline_macro:Msat_OrderVarIsInBoundary__*/((p).pVars[Var].pNext)");
    // go through its neigbors - if one of them is assigned add this var
    // add to the boundary those neighbors that are not there already
    // this will also get rid of variable outside of the current cone
    // because they are unassigned in Msat_SolverPrepare()
    vRound = (Msat_IntVec_t *)p.pSat.vAdjacents.pArray[Var];
    for  i = 0; i < vRound.nSize; i++  {
        // inlining of macro Msat_OrderVarIsAssigned
        if  (p).pSat.pAssigns[vRound.pArray[i]] != MSAT_VAR_UNASSIGNED { 
            break; }
    }
    if i != vRound.nSize {
        Msat_OrderRingAddLast( &p.rVars, &p.pVars[Var] ); }

    // unassigning a variable may lead to its adjacents dropping from the boundary
    for  i = 0; i < vRound.nSize; i++  {
        // inlining of macro Msat_OrderVarIsInBoundary
        if  (p).pVars[vRound.pArray[i]].pNext { // the neighbor is in the J-boundary (and unassigned)
            // inlining of macro Msat_OrderVarIsAssigned
            assert(!((p).pSat.pAssigns[vRound.pArray[i]] != MSAT_VAR_UNASSIGNED), "!/*__starting_point_inline_macro:Msat_OrderVarIsAssigned__*/((p).pSat.pAssigns[vRound.pArray[i]] != MSAT_VAR_UNASSIGNED)");
            vRound2 = (Msat_IntVec_t *)p.pSat.vAdjacents.pArray[vRound.pArray[i]];
            // go through its neighbors and determine if there is at least one assigned
            for  k = 0; k < vRound2.nSize; k++  {
                // inlining of macro Msat_OrderVarIsAssigned
                if  (p).pSat.pAssigns[vRound2.pArray[k]] != MSAT_VAR_UNASSIGNED { 
                    break; }
            }
            if k == vRound2.nSize { // there is no assigned vars, delete this one
                Msat_OrderRingRemove( &p.rVars, &p.pVars[vRound.pArray[i]] ); }
        }
    }
//timeSelect += clock() - clk;
//    Msat_OrderCheck( p );
}

/**Function*************************************************************

  Synopsis    [Updates the order after a variable changed weight.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderUpdate(p *Msat_Order_t, Var int32) {
}


/**Function*************************************************************

  Synopsis    [Adds node to the end of the ring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderRingAddLast(pRing *Msat_OrderRing_t, pVar *Msat_OrderVar_t) {
//printf( "adding %d\n", pVar->Num );
    // check that the node is not in a ring
    assert(pVar.pPrev == nil, "pVar.pPrev == nil");
    assert(pVar.pNext == nil, "pVar.pNext == nil");
    // if the ring is empty, make the node point to itself
    pRing.nItems++;
    if pRing.pRoot == nil {
        pRing.pRoot  = pVar;
        pVar.pPrev  = pVar;
        pVar.pNext  = pVar;
        return;
    }
    // if the ring is not empty, add it as the last entry
    pVar.pPrev = pRing.pRoot.pPrev;
    pVar.pNext = pRing.pRoot;
    pVar.pPrev.pNext = pVar;
    pVar.pNext.pPrev = pVar;

    // move the root so that it points to the new entry
//    pRing->pRoot = pRing->pRoot->pPrev;
}

/**Function*************************************************************

  Synopsis    [Removes the node from the ring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderRingRemove(pRing *Msat_OrderRing_t, pVar *Msat_OrderVar_t) {
//printf( "removing %d\n", pVar->Num );
    // check that the var is in a ring
    assert(pVar.pPrev, "pVar.pPrev");
    assert(pVar.pNext, "pVar.pNext");
    pRing.nItems--;
    if pRing.nItems == 0 {
        assert(pRing.pRoot == pVar, "pRing.pRoot == pVar");
        pVar.pPrev = nil;
        pVar.pNext = nil;
        pRing.pRoot = nil;
        return;
    }
    // move the root if needed
    if pRing.pRoot == pVar {
        pRing.pRoot = pVar.pNext; }
    // move the root to the next entry after pVar
    // this way all the additions to the list will be traversed first
//    pRing->pRoot = pVar->pPrev;
    // delete the node
    pVar.pPrev.pNext = pVar.pNext;
    pVar.pNext.pPrev = pVar.pPrev;
    pVar.pPrev = nil;
    pVar.pNext = nil;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

