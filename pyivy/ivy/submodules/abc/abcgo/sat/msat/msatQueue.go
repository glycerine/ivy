package msat

import "fmt"

/**CFile****************************************************************

  FileName    [msatQueue.c]

  PackageName [A C version of SAT solver MINISAT, originally developed 
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of 
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [The manager of the assignment propagation queue.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatQueue.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

//#include "msatInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Msat_Queue_t_ struct {
    nVars         int32;
    pVars        *int32;
    iFirst         int32;
    iLast         int32;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the variable propagation queue.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_QueueAlloc(nVars int32) *Msat_Queue_t {
    var p *Msat_Queue_t;
    // inlining of macro ABC_ALLOC
    p = ((Msat_Queue_t *) malloc(sizeof(Msat_Queue_t) * (size_t)(1)));
    memset( p, 0, sizeof(Msat_Queue_t) );
    p.nVars  = nVars;
    // inlining of macro ABC_ALLOC
    p.pVars  = ((int *) malloc(sizeof(int) * (size_t)(nVars)));
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocate the variable propagation queue.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_QueueFree(p *Msat_Queue_t) {
    // inlining of macro ABC_FREE
    ((p.pVars) ? (free((char *) (p.pVars)), (p.pVars) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Reads the queue size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_QueueReadSize(p *Msat_Queue_t) int32 {
    return p.iLast - p.iFirst;
}

/**Function*************************************************************

  Synopsis    [Insert an entry into the queue.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_QueueInsert(p *Msat_Queue_t, Lit int32) {
    if p.iLast == p.nVars {
        var i int32;
        assert(false, "0");
        for  i = 0; i < p.iLast; i++  {
            fmt.Printf("entry = %2d  lit = %2d  var = %2d \n", i, p.pVars[i], p.pVars[i]/2);
        }
    }
    assert(p.iLast < p.nVars, "p.iLast < p.nVars");
    p.pVars[p.iLast++] = Lit;
}

/**Function*************************************************************

  Synopsis    [Extracts an entry from the queue.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_QueueExtract(p *Msat_Queue_t) int32 {
    if p.iFirst == p.iLast {
        return -1; }
    return p.pVars[p.iFirst++];
}

/**Function*************************************************************

  Synopsis    [Resets the queue.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_QueueClear(p *Msat_Queue_t) {
    p.iFirst = 0;
    p.iLast  = 0;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

