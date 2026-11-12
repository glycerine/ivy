package msat

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


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// the variable package data structure
type Msat_Order_t_ struct {
    pSat       *Msat_Solver_t;         // the SAT solver 
    vIndex       *Msat_IntVec_t;       // the heap
    vHeap       *Msat_IntVec_t;        // the mapping of var num into its heap num
}

//The solver can communicate to the variable order the following parts:
//- the array of current assignments (pSat->pAssigns)
//- the array of variable activities (pSat->pdActivity)
//- the array of variables currently in the cone (pSat->vConeVars)
//- the array of arrays of variables adjucent to each(pSat->vAdjacents)

// Original macro:
// #define HLEFT(i)               ((i)<<1)

// Original macro:
// #define HRIGHT(i)             (((i)<<1)+1)

// Original macro:
// #define HPARENT(i)             ((i)>>1)

// Original macro:
// #define HCOMPARE(p, i, j)      ((p)->pSat->pdActivity[i] > (p)->pSat->pdActivity[j])

// Original macro:
// #define HHEAP(p, i)            ((p)->vHeap->pArray[i])

// Original macro:
// #define HSIZE(p)               ((p)->vHeap->nSize)

// Original macro:
// #define HOKAY(p, i)            ((i) >= 0 && (i) < (p)->vIndex->nSize)

// Original macro:
// #define HINHEAP(p, i)          (HOKAY(p, i) && (p)->vIndex->pArray[i] != 0)

// Original macro:
// #define HEMPTY(p)              (HSIZE(p) == 1)


// static int Msat_HeapCheck_rec( Msat_Order_t * p, int i );

// static int Msat_HeapGetTop( Msat_Order_t * p );

// static void Msat_HeapInsert( Msat_Order_t * p, int n );

// static void Msat_HeapIncrease( Msat_Order_t * p, int n );

// static void Msat_HeapPercolateUp( Msat_Order_t * p, int i );

// static void Msat_HeapPercolateDown( Msat_Order_t * p, int i );


var timeSelect abctime;

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
    // inlining of macro ABC_ALLOC
    p = ((Msat_Order_t *) malloc(sizeof(Msat_Order_t) * (size_t)(1)));
    memset( p, 0, sizeof(Msat_Order_t) );
    p.pSat   = pSat;
    p.vIndex = Msat_IntVecAlloc( 0 );
    p.vHeap  = Msat_IntVecAlloc( 0 );
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
    Msat_IntVecGrow( p.vIndex, nVarsMax );
    Msat_IntVecGrow( p.vHeap, nVarsMax + 1 );
    p.vIndex.nSize = nVarsMax;
    p.vHeap.nSize = 0;
}

/**Function*************************************************************

  Synopsis    [Cleans the ordering structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderClean(p *Msat_Order_t, vCone *Msat_IntVec_t) {
    var i int32;
    for  i = 0; i < p.vIndex.nSize; i++  {
        p.vIndex.pArray[i] = 0;
    }
    for  i = 0; i < vCone.nSize; i++  {
        assert(i+1 < p.vHeap.nCap, "i+1 < p.vHeap.nCap");
        p.vHeap.pArray[i+1] = vCone.pArray[i];

        assert(vCone.pArray[i] < p.vIndex.nSize, "vCone.pArray[i] < p.vIndex.nSize");
        p.vIndex.pArray[vCone.pArray[i]] = i+1;
    }
    p.vHeap.nSize = vCone.nSize + 1;
}

/**Function*************************************************************

  Synopsis    [Checks that the J-boundary is okay.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderCheck(p *Msat_Order_t) int32 {
    return Msat_HeapCheck_rec( p, 1 );
}

/**Function*************************************************************

  Synopsis    [Frees the ordering structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderFree(p *Msat_Order_t) {
    Msat_IntVecFree( p.vHeap );
    Msat_IntVecFree( p.vIndex );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}



/**Function*************************************************************

  Synopsis    [Selects the next variable to assign.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderVarSelect(p *Msat_Order_t) int32 {
    // Activity based decision:
//    while (!heap.empty()){
//        Var next = heap.getmin();
//        if (toLbool(assigns[next]) == l_Undef)
//            return next;
//    }
//    return var_Undef;

    var Var int32;
    var clk abctime  = Abc_Clock();

    // inlining of macro HEMPTY
    // inlining of macro HSIZE
    for  !(((p).vHeap.nSize) == 1)  {
        Var = Msat_HeapGetTop(p);
        if (p).pSat.pAssigns[Var] == MSAT_VAR_UNASSIGNED {
//assert( Msat_OrderCheck(p) );
timeSelect += Abc_Clock() - clk;
            return Var;
        }
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
}

/**Function*************************************************************

  Synopsis    [Updates the order after a variable is unassigned.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderVarUnassigned(p *Msat_Order_t, Var int32) {
//    if (!heap.inHeap(x))
//        heap.insert(x);

    var clk abctime  = Abc_Clock();
    // inlining of macro HINHEAP
    // inlining of macro HOKAY
    if !(((Var) >= 0 && (Var) < (p).vIndex.nSize) && (p).vIndex.pArray[Var] != 0) {
        Msat_HeapInsert( p, Var ); }
timeSelect += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    [Updates the order after a variable changed weight.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_OrderUpdate(p *Msat_Order_t, Var int32) {
//    if (heap.inHeap(x))
//        heap.increase(x);

    var clk abctime  = Abc_Clock();
    // inlining of macro HINHEAP
    // inlining of macro HOKAY
    if  ((Var) >= 0 && (Var) < (p).vIndex.nSize) && (p).vIndex.pArray[Var] != 0 {
        Msat_HeapIncrease( p, Var ); }
timeSelect += Abc_Clock() - clk;
}




/**Function*************************************************************

  Synopsis    [Checks the heap property recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_HeapCheck_rec(p *Msat_Order_t, i int32) int32 {
    // inlining of macro HSIZE
    return i >= ((p).vHeap.nSize) ||
        (
            // inlining of macro HPARENT
            // inlining of macro HCOMPARE
            // inlining of macro HHEAP
            ( ((i)>>1) == 0 || !((p).pSat.pdActivity[((p).vHeap.pArray[i])] > (p).pSat.pdActivity[((p).vHeap.pArray[((i)>>1)])]) ) &&

            // inlining of macro HLEFT
            Msat_HeapCheck_rec( p, ((i)<<1) ) && 
            
            // inlining of macro HRIGHT
            Msat_HeapCheck_rec( p, (((i)<<1)+1) )
        );
}

/**Function*************************************************************

  Synopsis    [Retrieves the minimum element.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_HeapGetTop(p *Msat_Order_t) int32 {
    var Result, NewTop int32;
    // inlining of macro HHEAP
    Result                    = ((p).vHeap.pArray[1]);
    NewTop                    = Msat_IntVecPop( p.vHeap );
    p.vHeap.pArray[1]       = NewTop;
    p.vIndex.pArray[NewTop] = 1;
    p.vIndex.pArray[Result] = 0;
    if p.vHeap.nSize > 1 {
        Msat_HeapPercolateDown( p, 1 ); }
    return Result;
}

/**Function*************************************************************

  Synopsis    [Inserts the new element.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_HeapInsert(p *Msat_Order_t, n int32) {
    // inlining of macro HOKAY
    assert(((n) >= 0 && (n) < (p).vIndex.nSize), "/*__starting_point_inline_macro:HOKAY__*/((n) >= 0 && (n) < (p).vIndex.nSize)");
    // inlining of macro HSIZE
    p.vIndex.pArray[n] = ((p).vHeap.nSize);
    Msat_IntVecPush( p.vHeap, n );
    Msat_HeapPercolateUp( p, p.vIndex.pArray[n] );
}

/**Function*************************************************************

  Synopsis    [Inserts the new element.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_HeapIncrease(p *Msat_Order_t, n int32) {
    Msat_HeapPercolateUp( p, p.vIndex.pArray[n] );
}

/**Function*************************************************************

  Synopsis    [Moves the entry up.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_HeapPercolateUp(p *Msat_Order_t, i int32) {
    // inlining of macro HHEAP
    var x int32  = ((p).vHeap.pArray[i]);
    // inlining of macro HPARENT
    // inlining of macro HCOMPARE
    // inlining of macro HHEAP
    for  ((i)>>1) != 0 && ((p).pSat.pdActivity[x] > (p).pSat.pdActivity[((p).vHeap.pArray[((i)>>1)])])  {
        // inlining of macro HHEAP
        // inlining of macro HPARENT
        p.vHeap.pArray[i]            = ((p).vHeap.pArray[((i)>>1)]);
        // inlining of macro HHEAP
        p.vIndex.pArray[((p).vHeap.pArray[i])] = i;
        // inlining of macro HPARENT
        i                              = ((i)>>1);
    }
    p.vHeap.pArray[i]  = x;
    p.vIndex.pArray[x] = i;
}

/**Function*************************************************************

  Synopsis    [Moves the entry down.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Msat_HeapPercolateDown(p *Msat_Order_t, i int32) {
    // inlining of macro HHEAP
    var x int32  = ((p).vHeap.pArray[i]);
    var Child int32;
    // inlining of macro HLEFT
    // inlining of macro HSIZE
    for  ((i)<<1) < ((p).vHeap.nSize)  {
        // inlining of macro HRIGHT
        // inlining of macro HSIZE
        // inlining of macro HCOMPARE
        // inlining of macro HHEAP
        // inlining of macro HLEFT
        if  (((i)<<1)+1) < ((p).vHeap.nSize) && ((p).pSat.pdActivity[((p).vHeap.pArray[(((i)<<1)+1)])] > (p).pSat.pdActivity[((p).vHeap.pArray[((i)<<1)])]) {
            // inlining of macro HRIGHT
            Child = (((i)<<1)+1);
        } else {
            // inlining of macro HLEFT
            Child = ((i)<<1); }
        // inlining of macro HCOMPARE
        // inlining of macro HHEAP
        if !((p).pSat.pdActivity[((p).vHeap.pArray[Child])] > (p).pSat.pdActivity[x]) {
            break; }
        // inlining of macro HHEAP
        p.vHeap.pArray[i]            = ((p).vHeap.pArray[Child]);
        // inlining of macro HHEAP
        p.vIndex.pArray[((p).vHeap.pArray[i])] = i;
        i                              = Child;
    }
    p.vHeap.pArray[i]  = x;
    p.vIndex.pArray[x] = i;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

