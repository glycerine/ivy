package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecWec.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable vector of resizable vectors.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecWec.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

         
type Vec_Wec_t struct {
    nCap              int32;
    nSize              int32;
    pArray       *Vec_Int_t;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// iterators through levels 
// Original macro:
// #define Vec_WecForEachLevel( vGlob, vVec, i )                                              \
//     for ( i = 0; (i < Vec_WecSize(vGlob)) && (((vVec) = Vec_WecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_WecForEachLevelVec( vLevels, vGlob, vVec, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(vLevels)) && (((vVec) = Vec_WecEntry(vGlob, Vec_IntEntry(vLevels, i))), 1); i++ )

// Original macro:
// #define Vec_WecForEachLevelStart( vGlob, vVec, i, LevelStart )                             \
//     for ( i = LevelStart; (i < Vec_WecSize(vGlob)) && (((vVec) = Vec_WecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_WecForEachLevelStop( vGlob, vVec, i, LevelStop )                               \
//     for ( i = 0; (i < LevelStop) && (((vVec) = Vec_WecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_WecForEachLevelStartStop( vGlob, vVec, i, LevelStart, LevelStop )              \
//     for ( i = LevelStart; (i < LevelStop) && (((vVec) = Vec_WecEntry(vGlob, i)), 1); i++ )

// Original macro:
// #define Vec_WecForEachLevelReverse( vGlob, vVec, i )                                       \
//     for ( i = Vec_WecSize(vGlob)-1; (i >= 0) && (((vVec) = Vec_WecEntry(vGlob, i)), 1); i-- )

// Original macro:
// #define Vec_WecForEachLevelReverseStartStop( vGlob, vVec, i, LevelStart, LevelStop )       \
//     for ( i = LevelStart-1; (i >= LevelStop) && (((vVec) = Vec_WecEntry(vGlob, i)), 1); i-- )

// Original macro:
// #define Vec_WecForEachLevelTwo( vGlob1, vGlob2, vVec1, vVec2, i )                          \
//     for ( i = 0; (i < Vec_WecSize(vGlob1)) && (((vVec1) = Vec_WecEntry(vGlob1, i)), 1) && (((vVec2) = Vec_WecEntry(vGlob2, i)), 1); i++ )

// Original macro:
// #define Vec_WecForEachLevelDouble( vGlob, vVec1, vVec2, i )                                \
//     for ( i = 0; (i < Vec_WecSize(vGlob)) && (((vVec1) = Vec_WecEntry(vGlob, i)), 1) && (((vVec2) = Vec_WecEntry(vGlob, i+1)), 1); i += 2 )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecAlloc(nCap int32) *Vec_Wec_t {
    var p *Vec_Wec_t;
    p = new(Vec_Wec_t);
    if nCap > 0 && nCap < 8 {
        nCap = 8; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_CALLOC( Vec_Int_t, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_WecAllocExact(nCap int32) *Vec_Wec_t {
    var p *Vec_Wec_t;
    assert(nCap >= 0, "nCap >= 0");
    p = new(Vec_Wec_t);
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_CALLOC( Vec_Int_t, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_WecStart(nSize int32) *Vec_Wec_t {
    var p *Vec_Wec_t;
    p = Vec_WecAlloc( nSize );
    p.nSize = nSize;
    return p;
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecGrow(p *Vec_Wec_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( Vec_Int_t, p.pArray, nCapMin ); 
    memset( p.pArray + p.nCap, 0, sizeof(Vec_Int_t) * (size_t)(nCapMin - p.nCap) );
    p.nCap   = nCapMin;
}
func Vec_WecInit(p *Vec_Wec_t, nSize int32) {
    Vec_WecGrow( p, nSize );
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecEntry(p *Vec_Wec_t, i int32) *Vec_Int_t {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray + i;
}
func Vec_WecEntryLast(p *Vec_Wec_t) *Vec_Int_t {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray + p.nSize - 1;
}
func Vec_WecEntryEntry(p *Vec_Wec_t, i int32, k int32) int32 {
    return Vec_IntEntry( Vec_WecEntry(p, i), k );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecArray(p *Vec_Wec_t) *Vec_Int_t {
    return p.pArray;
}
func Vec_WecLevelId(p *Vec_Wec_t, vLevel *Vec_Int_t) int32 {
    assert(p.pArray <= vLevel && vLevel < p.pArray + p.nSize, "p.pArray <= vLevel && vLevel < p.pArray + p.nSize");
    return vLevel - p.pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecCap(p *Vec_Wec_t) int32 {
    return p.nCap;
}
func Vec_WecSize(p *Vec_Wec_t) int32 {
    return p.nSize;
}
func Vec_WecLevelSize(p *Vec_Wec_t, i int32) int32 {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return Vec_IntSize( p.pArray + i );
}
func Vec_WecSizeSize(p *Vec_Wec_t) int32 {
    var vVec *Vec_Int_t;
    var i, Counter int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        Counter += Vec_IntSize(vVec);
    }
    return Counter;
}
func Vec_WecSizeUsed(p *Vec_Wec_t) int32 {
    var vVec *Vec_Int_t;
    var i, Counter int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        Counter += int32((Vec_IntSize(vVec) > 0));
    }
    return Counter;
}
func Vec_WecSizeUsedLimits(p *Vec_Wec_t, iStart int32, iStop int32) int32 {
    var vVec *Vec_Int_t;
    var i, Counter int32;
    // inlining of macro Vec_WecForEachLevelStartStop
    for  i = iStart; (i < iStop) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        Counter += int32((Vec_IntSize(vVec) > 0));
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecShrink(p *Vec_Wec_t, nSizeNew int32) {
    var vVec *Vec_Int_t; var i int32;
    // inlining of macro Vec_WecForEachLevelStart
    for  i = nSizeNew; (i < Vec_WecSize(p)) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        Vec_IntShrink( vVec, 0 );
    }
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecClear(p *Vec_Wec_t) {
    var vVec *Vec_Int_t;
    var i int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        Vec_IntClear( vVec );
    }
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecPush(p *Vec_Wec_t, Level int32, Entry int32) {
    if p.nSize < Level + 1 {
        Vec_WecGrow( p, Abc_MaxInt(2*p.nSize, Level + 1) );
        p.nSize = Level + 1;
    }
    Vec_IntPush( Vec_WecEntry(p, Level), Entry );
}
func Vec_WecPushLevel(p *Vec_Wec_t) *Vec_Int_t {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_WecGrow( p, 16 );
        } else {
            Vec_WecGrow( p, 2 * p.nCap ); }
    }
    ++p.nSize;
    return Vec_WecEntryLast( p );
}
func Vec_WecInsertLevel(p *Vec_Wec_t, i int32) *Vec_Int_t {
    var pTemp *Vec_Int_t;
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_WecGrow( p, 16 );
        } else {
            Vec_WecGrow( p, 2 * p.nCap ); }
    }
    ++p.nSize;
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    for  pTemp = p.pArray + p.nSize - 2; pTemp >= p.pArray + i; pTemp--  {
        pTemp[1] = pTemp[0];
    }
    Vec_IntZero( p.pArray + i );
    return p.pArray + i;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecMemory(p *Vec_Wec_t) float64 {
    var i int32;
    var Mem float64;
    if p == nil {  return 0.0; }
    Mem = sizeof(Vec_Int_t) * Vec_WecCap(p);
    for  i = 0; i < p.nSize; i++  {
        Mem += sizeof(int) * (size_t)Vec_IntCap( Vec_WecEntry(p, i) );
    }
    return Mem;
}

/**Function*************************************************************

  Synopsis    [Frees the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecZero(p *Vec_Wec_t) {
    p.pArray = nil;
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_WecErase(p *Vec_Wec_t) {
    var i int32;
    for  i = 0; i < p.nCap; i++  {
        ABC_FREE( p.pArray[i].pArray );
    }
    ABC_FREE( p.pArray );
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_WecFree(p *Vec_Wec_t) {
    Vec_WecErase( p );
    ABC_FREE( p );
}
func Vec_WecFreeP(p **Vec_Wec_t) {
    if *p == nil {
        return; }
    Vec_WecFree( *p );
    *p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecPushUnique(p *Vec_Wec_t, Level int32, Entry int32) {
    if p.nSize < Level + 1 {
        Vec_WecPush( p, Level, Entry );
    } else {
        Vec_IntPushUnique( Vec_WecEntry(p, Level), Entry ); }
}

/**Function*************************************************************

  Synopsis    [Frees the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecDup(p *Vec_Wec_t) *Vec_Wec_t {
    var vNew *Vec_Wec_t;
    var vVec *Vec_Int_t;
    var i, k, Entry int32;
    vNew = Vec_WecAlloc( Vec_WecSize(p) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        Vec_IntForEachEntry( vVec, Entry, k )
            Vec_WecPush( vNew, i, Entry );
    }
    return vNew;
}

/**Function*************************************************************

  Synopsis    [Sorting by array size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecSortCompare1(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    if Vec_IntSize(p1) < Vec_IntSize(p2) {
        return -1; }
    if Vec_IntSize(p1) > Vec_IntSize(p2) { 
        return 1; }
    return 0; 
}
func Vec_WecSortCompare2(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    if Vec_IntSize(p1) > Vec_IntSize(p2) {
        return -1; }
    if Vec_IntSize(p1) < Vec_IntSize(p2) { 
        return 1; }
    return 0; 
}
func Vec_WecSort(p *Vec_Wec_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Vec_Int_t), 
                        (int (*)(void *, void *)) Vec_WecSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Vec_Int_t), 
                        (int (*)(void *, void *)) Vec_WecSortCompare1 ); }
}


/**Function*************************************************************

  Synopsis    [Sorting by the first entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecSortCompare3(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    if Vec_IntEntry(p1,0) < Vec_IntEntry(p2,0) {
        return -1; }
    if Vec_IntEntry(p1,0) > Vec_IntEntry(p2,0) { 
        return 1; }
    return 0; 
}
func Vec_WecSortCompare4(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    if Vec_IntEntry(p1,0) > Vec_IntEntry(p2,0) {
        return -1; }
    if Vec_IntEntry(p1,0) < Vec_IntEntry(p2,0) { 
        return 1; }
    return 0; 
}
func Vec_WecSortByFirstInt(p *Vec_Wec_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Vec_Int_t), 
                        (int (*)(void *, void *)) Vec_WecSortCompare4 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Vec_Int_t), 
                        (int (*)(void *, void *)) Vec_WecSortCompare3 ); }
}

/**Function*************************************************************

  Synopsis    [Sorting by the last entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecSortCompare5(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    if Vec_IntEntryLast(p1) < Vec_IntEntryLast(p2) {
        return -1; }
    if Vec_IntEntryLast(p1) > Vec_IntEntryLast(p2) { 
        return 1; }
    return 0; 
}
func Vec_WecSortCompare6(p1 *Vec_Int_t, p2 *Vec_Int_t) int32 {
    if Vec_IntEntryLast(p1) > Vec_IntEntryLast(p2) {
        return -1; }
    if Vec_IntEntryLast(p1) < Vec_IntEntryLast(p2) { 
        return 1; }
    return 0; 
}
func Vec_WecSortByLastInt(p *Vec_Wec_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Vec_Int_t), 
                        (int (*)(void *, void *)) Vec_WecSortCompare6 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(Vec_Int_t), 
                        (int (*)(void *, void *)) Vec_WecSortCompare5 ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecPrint(p *Vec_Wec_t, fSkipSingles int32) {
    var vVec *Vec_Int_t;
    var i, k, Entry int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        if (fSkipSingles != 0) && Vec_IntSize(vVec) == 1 {
            continue; }
        fmt.Printf(" %4d : {", i);
        Vec_IntForEachEntry( vVec, Entry, k )
            fmt.Printf(" %d", Entry);
        fmt.Printf(" }\n");
    }
}
func Vec_WecPrintLits(p *Vec_Wec_t) {
    var vVec *Vec_Int_t;
    var i, k, iLit int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vVec) = Vec_WecEntry(p, i)), 1); i++  {
        fmt.Printf(" %4d : %2d  {", i, Vec_IntSize(vVec));
        Vec_IntForEachEntry( vVec, iLit, k )
            fmt.Printf(" %c%d", Abc_LitIsCompl(iLit) ? '-' : '+', Abc_Lit2Var(iLit));
        fmt.Printf(" }\n");
    }
}

/**Function*************************************************************

  Synopsis    [Derives the set of equivalence classes.]

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecCreateClasses(vMap *Vec_Int_t) *Vec_Wec_t {
    var vClasses *Vec_Wec_t;
    var i, Entry int32;
    vClasses = Vec_WecStart( Vec_IntFindMax(vMap) + 1 );
    Vec_IntForEachEntry( vMap, Entry, i )
        Vec_WecPush( vClasses, Entry, i );
    return vClasses;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecCountNonTrivial(p *Vec_Wec_t, pnUsed *int32) int32 {
    var vClass *Vec_Int_t;
    var i, nClasses int32;
    *pnUsed = 0;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vClass) = Vec_WecEntry(p, i)), 1); i++  {
        if Vec_IntSize(vClass) < 2 {
            continue; }
        nClasses++;
        (*pnUsed) += Vec_IntSize(vClass);
    }
    return nClasses;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecCollectFirsts(p *Vec_Wec_t) *Vec_Int_t {
    var vFirsts, vLevel *Vec_Int_t;
    var i int32;
    vFirsts = Vec_IntAlloc( Vec_WecSize(p) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vLevel) = Vec_WecEntry(p, i)), 1); i++  {
        if Vec_IntSize(vLevel) > 0 {
            Vec_IntPush( vFirsts, Vec_IntEntry(vLevel, 0) ); }
    }
    return vFirsts;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecConvertToVecPtr(p *Vec_Wec_t) *Vec_Ptr_t {
    var vCopy *Vec_Ptr_t;
    var vLevel *Vec_Int_t;
    var i int32;
    vCopy = Vec_PtrAlloc( Vec_WecSize(p) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(p)) && (((vLevel) = Vec_WecEntry(p, i)), 1); i++  {
        Vec_PtrPush( vCopy, Vec_IntDup(vLevel) );
    }
    return vCopy;
}


/**Function*************************************************************

  Synopsis    [Temporary vector marking.]

  Description [The vector should be static when the marking is used.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecIntHasMark(vVec *Vec_Int_t) int32 { return (vVec.nCap >> 30) & 1; }
func Vec_WecIntSetMark(vVec *Vec_Int_t) { vVec.nCap |= (1<<30);         }
func Vec_WecIntXorMark(vVec *Vec_Int_t) { vVec.nCap ^= (1<<30);         }
func Vec_WecMarkLevels(vCubes *Vec_Wec_t, vLevels *Vec_Int_t) {
    var vCube *Vec_Int_t;
    var i int32;
    // inlining of macro Vec_WecForEachLevelVec
    for  i = 0; (i < Vec_IntSize(vLevels)) && (((vCube) = Vec_WecEntry(vCubes, Vec_IntEntry(vLevels, i))), 1); i++  {
        assert(!(Vec_WecIntHasMark( vCube ) != 0), "!Vec_WecIntHasMark( vCube )");
        Vec_WecIntXorMark( vCube );
    }
}
func Vec_WecUnmarkLevels(vCubes *Vec_Wec_t, vLevels *Vec_Int_t) {
    var vCube *Vec_Int_t;
    var i int32;
    // inlining of macro Vec_WecForEachLevelVec
    for  i = 0; (i < Vec_IntSize(vLevels)) && (((vCube) = Vec_WecEntry(vCubes, Vec_IntEntry(vLevels, i))), 1); i++  {
        assert((Vec_WecIntHasMark( vCube ) != 0), "Vec_WecIntHasMark( vCube )");
        Vec_WecIntXorMark( vCube );
    }
}

/**Function*************************************************************

  Synopsis    [Removes 0-size vectors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_WecRemoveEmpty(vCubes *Vec_Wec_t) {
    var vCube *Vec_Int_t;
    var i, k int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vCubes)) && (((vCube) = Vec_WecEntry(vCubes, i)), 1); i++  {
        if Vec_IntSize(vCube) > 0 {
            vCubes.pArray[k++] = *vCube;
        } else {
            ABC_FREE( vCube.pArray ); }
    }
    for  i = k; i < Vec_WecSize(vCubes); i++  {
        Vec_IntZero( Vec_WecEntry(vCubes, i) );
    }
    Vec_WecShrink( vCubes, k );
//    Vec_WecSortByFirstInt( vCubes, 0 );
}


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

