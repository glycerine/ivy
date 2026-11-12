package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecVec.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable vector of resizable vectors.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecVec.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

         
type Vec_Vec_t struct {
    nCap              int32;
    nSize              int32;
    pArray           *any;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// iterators through levels 
// Original macro:
// #define Vec_VecForEachLevel( vGlob, vVec, i )                                                 \
//     for ( i = 0; (i < Vec_VecSize(vGlob)) && (((vVec) = Vec_VecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelStart( vGlob, vVec, i, LevelStart )                                \
//     for ( i = LevelStart; (i < Vec_VecSize(vGlob)) && (((vVec) = Vec_VecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelStop( vGlob, vVec, i, LevelStop )                                  \
//     for ( i = 0; (i < LevelStop) && (((vVec) = Vec_VecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelStartStop( vGlob, vVec, i, LevelStart, LevelStop )                 \
//     for ( i = LevelStart; (i < LevelStop) && (((vVec) = Vec_VecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelReverse( vGlob, vVec, i )                                          \
//     for ( i = Vec_VecSize(vGlob)-1; (i >= 0) && (((vVec) = Vec_VecEntry(vGlob, i)), 1); i-- )

// Original macro:
// #define Vec_VecForEachLevelReverseStartStop( vGlob, vVec, i, LevelStart, LevelStop )          \
//     for ( i = LevelStart-1; (i >= LevelStop) && (((vVec) = Vec_VecEntry(vGlob, i)), 1); i-- )

// Original macro:
// #define Vec_VecForEachLevelTwo( vGlob1, vGlob2, vVec1, vVec2, i )                                                 \
//     for ( i = 0; (i < Vec_VecSize(vGlob1)) && (((vVec1) = Vec_VecEntry(vGlob1, i)), 1) && (((vVec2) = Vec_VecEntry(vGlob2, i)), 1); i++ )


// iterators through levels 
// Original macro:
// #define Vec_VecForEachLevelInt( vGlob, vVec, i )                                              \
//     for ( i = 0; (i < Vec_VecSize(vGlob)) && (((vVec) = Vec_VecEntryInt(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelIntStart( vGlob, vVec, i, LevelStart )                             \
//     for ( i = LevelStart; (i < Vec_VecSize(vGlob)) && (((vVec) = Vec_VecEntryInt(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelIntStop( vGlob, vVec, i, LevelStop )                               \
//     for ( i = 0; (i < LevelStop) && (((vVec) = Vec_VecEntryInt(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelIntStartStop( vGlob, vVec, i, LevelStart, LevelStop )              \
//     for ( i = LevelStart; (i < LevelStop) && (((vVec) = Vec_VecEntryInt(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_VecForEachLevelIntReverse( vGlob, vVec, i )                                       \
//     for ( i = Vec_VecSize(vGlob)-1; (i >= 0) && (((vVec) = Vec_VecEntryInt(vGlob, i)), 1); i-- )

// Original macro:
// #define Vec_VecForEachLevelIntReverseStartStop( vGlob, vVec, i, LevelStart, LevelStop )       \
//     for ( i = LevelStart-1; (i >= LevelStop) && (((vVec) = Vec_VecEntryInt(vGlob, i)), 1); i-- )

// Original macro:
// #define Vec_VecForEachLevelIntTwo( vGlob1, vGlob2, vVec1, vVec2, i )                          \
//     for ( i = 0; (i < Vec_VecSize(vGlob1)) && (((vVec1) = Vec_VecEntryInt(vGlob1, i)), 1) && (((vVec2) = Vec_VecEntryInt(vGlob2, i)), 1); i++ )


// iteratores through entries
// Original macro:
// #define Vec_VecForEachEntry( Type, vGlob, pEntry, i, k )                                      \
//     for ( i = 0; i < Vec_VecSize(vGlob); i++ )                                                \
//         Vec_PtrForEachEntry( Type, Vec_VecEntry(vGlob, i), pEntry, k ) 

// Original macro:
// #define Vec_VecForEachEntryLevel( Type, vGlob, pEntry, i, Level )                             \
//         Vec_PtrForEachEntry( Type, Vec_VecEntry(vGlob, Level), pEntry, i ) 

// Original macro:
// #define Vec_VecForEachEntryStart( Type, vGlob, pEntry, i, k, LevelStart )                     \
//     for ( i = LevelStart; i < Vec_VecSize(vGlob); i++ )                                       \
//         Vec_PtrForEachEntry( Type, Vec_VecEntry(vGlob, i), pEntry, k ) 

// Original macro:
// #define Vec_VecForEachEntryStartStop( Type, vGlob, pEntry, i, k, LevelStart, LevelStop )      \
//     for ( i = LevelStart; i < LevelStop; i++ )                                                \
//         Vec_PtrForEachEntry( Type, Vec_VecEntry(vGlob, i), pEntry, k ) 

// Original macro:
// #define Vec_VecForEachEntryReverse( Type, vGlob, pEntry, i, k )                               \
//     for ( i = 0; i < Vec_VecSize(vGlob); i++ )                                                \
//         Vec_PtrForEachEntryReverse( Type, Vec_VecEntry(vGlob, i), pEntry, k ) 

// Original macro:
// #define Vec_VecForEachEntryReverseReverse( Type, vGlob, pEntry, i, k )                        \
//     for ( i = Vec_VecSize(vGlob) - 1; i >= 0; i-- )                                           \
//         Vec_PtrForEachEntryReverse( Type, Vec_VecEntry(vGlob, i), pEntry, k ) 

// Original macro:
// #define Vec_VecForEachEntryReverseStart( Type, vGlob, pEntry, i, k, LevelStart )              \
//     for ( i = LevelStart-1; i >= 0; i-- )                                                     \
//         Vec_PtrForEachEntry( Type, Vec_VecEntry(vGlob, i), pEntry, k ) 


// iteratores through entries
// Original macro:
// #define Vec_VecForEachEntryInt( vGlob, Entry, i, k )                                          \
//     for ( i = 0; i < Vec_VecSize(vGlob); i++ )                                                \
//         Vec_IntForEachEntry( Vec_VecEntryInt(vGlob, i), Entry, k ) 

// Original macro:
// #define Vec_VecForEachEntryIntLevel( vGlob, Entry, i, Level )                                 \
//         Vec_IntForEachEntry( Vec_VecEntryInt(vGlob, Level), Entry, i ) 

// Original macro:
// #define Vec_VecForEachEntryIntStart( vGlob, Entry, i, k, LevelStart )                         \
//     for ( i = LevelStart; i < Vec_VecSize(vGlob); i++ )                                       \
//         Vec_IntForEachEntry( Vec_VecEntryInt(vGlob, i), Entry, k ) 

// Original macro:
// #define Vec_VecForEachEntryIntStartStop( vGlob, Entry, i, k, LevelStart, LevelStop )          \
//     for ( i = LevelStart; i < LevelStop; i++ )                                                \
//         Vec_IntForEachEntry( Vec_VecEntryInt(vGlob, i), Entry, k ) 

// Original macro:
// #define Vec_VecForEachEntryIntReverse( vGlob, Entry, i, k )                                   \
//     for ( i = 0; i < Vec_VecSize(vGlob); i++ )                                                \
//         Vec_IntForEachEntryReverse( Vec_VecEntryInt(vGlob, i), Entry, k ) 

// Original macro:
// #define Vec_VecForEachEntryIntReverseReverse( vGlob, Entry, i, k )                            \
//     for ( i = Vec_VecSize(vGlob) - 1; i >= 0; i-- )                                           \
//         Vec_IntForEachEntryReverse( Vec_VecEntryInt(vGlob, i), Entry, k ) 

// Original macro:
// #define Vec_VecForEachEntryIntReverseStart( vGlob, Entry, i, k, LevelStart )                  \
//     for ( i = LevelStart-1; i >= 0; i-- )                                                     \
//         Vec_IntForEachEntry( Vec_VecEntryInt(vGlob, i), Entry, k ) 


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecAlloc(nCap int32) *Vec_Vec_t {
    var p *Vec_Vec_t;
    p = new(Vec_Vec_t);
    if nCap > 0 && nCap < 8 {
        nCap = 8; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( void *, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecStart(nSize int32) *Vec_Vec_t {
    var p *Vec_Vec_t;
    var i int32;
    p = Vec_VecAlloc( nSize );
    for  i = 0; i < nSize; i++  {
        p.pArray[i] = Vec_PtrAlloc( 0 );
    }
    p.nSize = nSize;
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecExpand(p *Vec_Vec_t, Level int32) {
    var i int32;
    if p.nSize >= Level + 1 {
        return; }
    Vec_PtrGrow( (Vec_Ptr_t *)p, Level + 1 );
    for  i = p.nSize; i <= Level; i++  {
        p.pArray[i] = Vec_PtrAlloc( 0 );
    }
    p.nSize = Level + 1;
}
func Vec_VecExpandInt(p *Vec_Vec_t, Level int32) {
    var i int32;
    if p.nSize >= Level + 1 {
        return; }
    Vec_IntGrow( (Vec_Int_t *)p, Level + 1 );
    for  i = p.nSize; i <= Level; i++  {
        p.pArray[i] = Vec_PtrAlloc( 0 );
    }
    p.nSize = Level + 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecSize(p *Vec_Vec_t) int32 {
    return p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecCap(p *Vec_Vec_t) int32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecLevelSize(p *Vec_Vec_t, i int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return Vec_PtrSize( (Vec_Ptr_t *)p.pArray[i] );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecEntry(p *Vec_Vec_t, i int32) *Vec_Ptr_t {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return (Vec_Ptr_t *)p.pArray[i];
}
func Vec_VecEntryInt(p *Vec_Vec_t, i int32) *Vec_Int_t {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return (Vec_Int_t *)p.pArray[i];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecMemory(p *Vec_Vec_t) float64 {
    var i int32;
    var Mem float64;
    if p == nil {  return 0.0; }
    Mem = Vec_PtrMemory( (Vec_Ptr_t *)p );
    for  i = 0; i < p.nSize; i++  {
        if Vec_VecEntry(p, i) != nil {
            Mem += Vec_PtrMemory( Vec_VecEntry(p, i) ); }
    }
    return Mem;
}
func Vec_VecMemoryInt(p *Vec_Vec_t) float64 {
    var i int32;
    var Mem float64;
    if p == nil {  return 0.0; }
    Mem = Vec_PtrMemory( (Vec_Ptr_t *)p );
    for  i = 0; i < p.nSize; i++  {
        if Vec_VecEntry(p, i) != nil {
            Mem += Vec_IntMemory( Vec_VecEntryInt(p, i) ); }
    }
    return Mem;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecEntryEntry(p *Vec_Vec_t, i int32, k int32) any {
    return Vec_PtrEntry( Vec_VecEntry(p, i), k );
}
func Vec_VecEntryEntryInt(p *Vec_Vec_t, i int32, k int32) int32 {
    return Vec_IntEntry( Vec_VecEntryInt(p, i), k );
}

/**Function*************************************************************

  Synopsis    [Frees the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecFree(p *Vec_Vec_t) {
    var vVec *Vec_Ptr_t;
    var i int32;
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(p)) && (((vVec) = Vec_VecEntry(p, i)), 1); i++  {
        if vVec { Vec_PtrFree( vVec ); }
    }
    Vec_PtrFree( (Vec_Ptr_t *)p );
}
func Vec_VecErase(p *Vec_Vec_t) {
    var vVec *Vec_Ptr_t;
    var i int32;
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(p)) && (((vVec) = Vec_VecEntry(p, i)), 1); i++  {
        if vVec { Vec_PtrFree( vVec ); }
    }
    Vec_PtrErase( (Vec_Ptr_t *)p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecFreeP(p **Vec_Vec_t) {
    if *p == nil {
        return; }
    Vec_VecFree( *p );
    *p = nil;
}

/**Function*************************************************************

  Synopsis    [Frees the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecDup(p *Vec_Vec_t) *Vec_Vec_t {
    var vNew, vVec *Vec_Ptr_t;
    var i int32;
    vNew = Vec_PtrAlloc( Vec_VecSize(p) );
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(p)) && (((vVec) = Vec_VecEntry(p, i)), 1); i++  {
        Vec_PtrPush( vNew, Vec_PtrDup(vVec) );
    }
    return (Vec_Vec_t *)vNew;
}
func Vec_VecDupInt(p *Vec_Vec_t) *Vec_Vec_t {
    var vNew *Vec_Ptr_t;
    var vVec *Vec_Int_t;
    var i int32;
    vNew = Vec_PtrAlloc( Vec_VecSize(p) );
    // inlining of macro Vec_VecForEachLevelInt
    for  i = 0; (i < Vec_VecSize(p)) && (((vVec) = Vec_VecEntryInt(p, i)), 1); i++  {
        Vec_PtrPush( vNew, Vec_IntDup(vVec) );
    }
    return (Vec_Vec_t *)vNew;
}

/**Function*************************************************************

  Synopsis    [Frees the vector of vectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecSizeSize(p *Vec_Vec_t) int32 {
    var vVec *Vec_Ptr_t;
    var i, Counter int32;
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(p)) && (((vVec) = Vec_VecEntry(p, i)), 1); i++  {
        Counter += vVec.nSize;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecClear(p *Vec_Vec_t) {
    var vVec *Vec_Ptr_t;
    var i int32;
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(p)) && (((vVec) = Vec_VecEntry(p, i)), 1); i++  {
        Vec_PtrClear( vVec );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecPush(p *Vec_Vec_t, Level int32, Entry any) {
    if p.nSize < Level + 1 {
        var i int32;
        Vec_PtrGrow( (Vec_Ptr_t *)p, Level + 1 );
        for  i = p.nSize; i < Level + 1; i++  {
            p.pArray[i] = Vec_PtrAlloc( 0 );
        }
        p.nSize = Level + 1;
    }
    Vec_PtrPush( Vec_VecEntry(p, Level), Entry );
}
func Vec_VecPushInt(p *Vec_Vec_t, Level int32, Entry int32) {
    if p.nSize < Level + 1 {
        var i int32;
        Vec_PtrGrow( (Vec_Ptr_t *)p, Level + 1 );
        for  i = p.nSize; i < Level + 1; i++  {
            p.pArray[i] = Vec_IntAlloc( 0 );
        }
        p.nSize = Level + 1;
    }
    Vec_IntPush( Vec_VecEntryInt(p, Level), Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecPushUnique(p *Vec_Vec_t, Level int32, Entry any) {
    if p.nSize < Level + 1 {
        Vec_VecPush( p, Level, Entry );
    } else {
        Vec_PtrPushUnique( Vec_VecEntry(p, Level), Entry ); }
}
func Vec_VecPushUniqueInt(p *Vec_Vec_t, Level int32, Entry int32) {
    if p.nSize < Level + 1 {
        Vec_VecPushInt( p, Level, Entry );
    } else {
        Vec_IntPushUnique( Vec_VecEntryInt(p, Level), Entry ); }
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two arrays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecSortCompare1(pp1 **Vec_Ptr_t, pp2 **Vec_Ptr_t) int32 {
    if Vec_PtrSize(*pp1) < Vec_PtrSize(*pp2) {
        return -1; }
    if Vec_PtrSize(*pp1) > Vec_PtrSize(*pp2) { 
        return 1; }
    return 0; 
}
func Vec_VecSortCompare2(pp1 **Vec_Ptr_t, pp2 **Vec_Ptr_t) int32 {
    if Vec_PtrSize(*pp1) > Vec_PtrSize(*pp2) {
        return -1; }
    if Vec_PtrSize(*pp1) < Vec_PtrSize(*pp2) { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecSort(p *Vec_Vec_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(void *), 
                        (int (*)(void *, void *)) Vec_VecSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(void *), 
                        (int (*)(void *, void *)) Vec_VecSortCompare1 ); }
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecSortCompare3(pp1 **Vec_Int_t, pp2 **Vec_Int_t) int32 {
    if Vec_IntEntry(*pp1,0) < Vec_IntEntry(*pp2,0) {
        return -1; }
    if Vec_IntEntry(*pp1,0) > Vec_IntEntry(*pp2,0) { 
        return 1; }
    return 0; 
}
func Vec_VecSortCompare4(pp1 **Vec_Int_t, pp2 **Vec_Int_t) int32 {
    if Vec_IntEntry(*pp1,0) > Vec_IntEntry(*pp2,0) {
        return -1; }
    if Vec_IntEntry(*pp1,0) < Vec_IntEntry(*pp2,0) { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecSortByFirstInt(p *Vec_Vec_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(void *), 
                        (int (*)(void *, void *)) Vec_VecSortCompare4 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(void *), 
                        (int (*)(void *, void *)) Vec_VecSortCompare3 ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_VecPrintInt(p *Vec_Vec_t, fSkipSingles int32) {
    var i, k, Entry int32;
    // inlining of macro Vec_VecForEachEntryInt
    for ( i = 0; i < Vec_VecSize(p); i++ )                                                        Vec_IntForEachEntry( Vec_VecEntryInt(p, i), Entry, k ) 
    {
        if (fSkipSingles != 0) && Vec_VecLevelSize(p, i) == 1 {
            break; }
        if k == 0 {
            fmt.Printf(" %4d : {", i); }
        fmt.Printf(" %d", Entry);
        if k == Vec_VecLevelSize(p, i) - 1 {
            fmt.Printf(" }\n"); }
    }
}

//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

