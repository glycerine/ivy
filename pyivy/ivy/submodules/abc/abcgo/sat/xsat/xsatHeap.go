package xsat

/**CFile****************************************************************

  FileName    [xsatHeap.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [xSAT - A SAT solver written in C.
               Read the license file for more info.]

  Synopsis    [Heap implementation.]

  Author      [Bruno Schmitt <boschmitt@inf.ufrgs.br>]

  Affiliation [UC Berkeley / UFRGS]

  Date        [Ver. 1.0. Started - November 10, 2016.]

  Revision    []

***********************************************************************/



//#include "misc/util/abc_global.h"
//#include "misc/vec/vecInt.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
   
type xSAT_Heap_t struct {
    vActivity  *Vec_Int_t;
    vIndices  *Vec_Int_t;
    vHeap  *Vec_Int_t;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapSize(h *xSAT_Heap_t) int32 {
    return Vec_IntSize( h.vHeap );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapInHeap(h *xSAT_Heap_t, Var int32) int32 {
    return ( Var < Vec_IntSize( h.vIndices ) ) && ( Vec_IntEntry( h.vIndices, Var ) >= 0 );
}

func Left(i int32) int32 { return 2 * i + 1; }
func Right(i int32) int32 { return ( i + 1 ) * 2; }
func Parent(i int32) int32 { return ( i - 1 ) >> 1; }
func Compare(p *xSAT_Heap_t, x int32, y int32) int32 {
    return uint32(  Vec_IntEntry( p.vActivity, x )) > uint32(  Vec_IntEntry( p.vActivity, y ));
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapPercolateUp(h *xSAT_Heap_t, i int32) {
    var x int32  = Vec_IntEntry( h.vHeap, i );
    var p int32  = Parent( i );

    for  i != 0 && (Compare( h, x, Vec_IntEntry( h.vHeap, p ) ) != 0)  {
        Vec_IntWriteEntry( h.vHeap, i, Vec_IntEntry( h.vHeap, p ) );
        Vec_IntWriteEntry( h.vIndices, Vec_IntEntry( h.vHeap, p ), i );
        i = p;
        p = Parent(p);
    }
    Vec_IntWriteEntry( h.vHeap, i, x );
    Vec_IntWriteEntry( h.vIndices, x, i );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapPercolateDown(h *xSAT_Heap_t, i int32) {
    var x int32  = Vec_IntEntry( h.vHeap, i );

    for  Left( i ) < Vec_IntSize( h.vHeap )  {
        var child int32;
        if Right( i ) < Vec_IntSize( h.vHeap ) &&
                    (Compare( h, Vec_IntEntry( h.vHeap, Right( i ) ), Vec_IntEntry( h.vHeap, Left( i ) ) ) != 0) {
            child = Right( i );
        } else {
            child = Left( i );
        }

        if !(Compare( h, Vec_IntEntry( h.vHeap, child ), x ) != 0) {
            break; }

        Vec_IntWriteEntry( h.vHeap, i, Vec_IntEntry( h.vHeap, child ) );
        Vec_IntWriteEntry( h.vIndices, Vec_IntEntry( h.vHeap, i ), i );
        i = child;
    }
    Vec_IntWriteEntry( h.vHeap, i, x );
    Vec_IntWriteEntry( h.vIndices, x, i );
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapAlloc(vActivity *Vec_Int_t) *xSAT_Heap_t {
    // inlining of macro ABC_ALLOC
    var p *xSAT_Heap_t  = ((xSAT_Heap_t *) malloc(sizeof(xSAT_Heap_t) * (size_t)(1)));
    p.vActivity = vActivity;
    p.vIndices = Vec_IntAlloc( 0 );
    p.vHeap = Vec_IntAlloc( 0 );

    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapFree(p *xSAT_Heap_t) {
    Vec_IntFree( p.vIndices );
    Vec_IntFree( p.vHeap );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapIncrease(h *xSAT_Heap_t, e int32) {
    assert((xSAT_HeapInHeap( h, e ) != 0), "xSAT_HeapInHeap( h, e )");
    xSAT_HeapPercolateDown( h, Vec_IntEntry( h.vIndices, e ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapDecrease(p *xSAT_Heap_t, e int32) {
    assert((xSAT_HeapInHeap( p, e ) != 0), "xSAT_HeapInHeap( p, e )");
    xSAT_HeapPercolateUp( p , Vec_IntEntry( p.vIndices, e ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapInsert(p *xSAT_Heap_t, n int32) {
    Vec_IntFillExtra( p.vIndices, n + 1, -1);
    assert(!(xSAT_HeapInHeap( p, n ) != 0), "!xSAT_HeapInHeap( p, n )");

    Vec_IntWriteEntry( p.vIndices, n, Vec_IntSize( p.vHeap ) );
    Vec_IntPush( p.vHeap, n );
    xSAT_HeapPercolateUp( p, Vec_IntEntry( p.vIndices, n ) );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapUpdate(p *xSAT_Heap_t, i int32) {
    if !(xSAT_HeapInHeap( p, i ) != 0) {
        xSAT_HeapInsert( p, i );
    } else {
        xSAT_HeapPercolateUp( p, Vec_IntEntry( p.vIndices, i ) );
        xSAT_HeapPercolateDown( p, Vec_IntEntry( p.vIndices, i ) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapBuild(p *xSAT_Heap_t, Vars *Vec_Int_t) {
    var i, Var int32;

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vHeap)) && (((Var) = Vec_IntEntry(p.vHeap, i)), 1); i++  {
        Vec_IntWriteEntry( p.vIndices, Var, -1 );
    }
    Vec_IntClear( p.vHeap );

    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(Vars)) && (((Var) = Vec_IntEntry(Vars, i)), 1); i++  {
        Vec_IntWriteEntry( p.vIndices, Var, i );
        Vec_IntPush( p.vHeap, Var );
    }

    for  ( i = Vec_IntSize( p.vHeap ) / 2 - 1 ); i >= 0; i--  {
        xSAT_HeapPercolateDown( p, i );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapClear(p *xSAT_Heap_t) {
    Vec_IntFill( p.vIndices, Vec_IntSize( p.vIndices ), -1 );
    Vec_IntClear( p.vHeap );
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func xSAT_HeapRemoveMin(p *xSAT_Heap_t) int32 {
    var x int32  = Vec_IntEntry( p.vHeap, 0 );
    Vec_IntWriteEntry( p.vHeap, 0, Vec_IntEntryLast( p.vHeap ) );
    Vec_IntWriteEntry( p.vIndices, Vec_IntEntry( p.vHeap, 0), 0 );
    Vec_IntWriteEntry( p.vIndices, x, -1 );
    Vec_IntPop( p.vHeap );
    if Vec_IntSize( p.vHeap ) > 1 {
        xSAT_HeapPercolateDown( p, 0 ); }
    return x;
}

//ABC_NAMESPACE_HEADER_END


