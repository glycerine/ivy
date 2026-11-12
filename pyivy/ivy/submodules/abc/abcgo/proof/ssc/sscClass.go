package ssc

import "fmt"

/**CFile****************************************************************

  FileName    [sscClass.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sweeping under constraints.]

  Synopsis    [Equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 29, 2008.]

  Revision    [$Id: sscClass.c,v 1.00 2008/07/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sscInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes hash key of the simuation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Ssc_GiaSimHashKey_s_Primes [16]int32 = [16]int32{ 
        1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 
        4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147 };
func Ssc_GiaSimHashKey(p *Gia_Man_t, iObj int32, nTableSize int32) int32 {
    
    var pSim *word  = Gia_ObjSim( p, iObj );
    var uHash uint32;
    var i int32;
    var nWords int32 = Gia_ObjSimWords(p);
    if pSim[0] & 1 {
        for  i = 0; i < nWords; i++  {
            uHash ^= ~pSim[i] * spStatic_Ssc_GiaSimHashKey_s_Primes[i & 0xf];
        }
    } else {
        for  i = 0; i < nWords; i++  {
            uHash ^= pSim[i] * spStatic_Ssc_GiaSimHashKey_s_Primes[i & 0xf];
        } }
    return int32((uHash % nTableSize));
}

/**Function*************************************************************

  Synopsis    [Returns 1 if sim patterns are equal up to the complement.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaSimIsConst0(p *Gia_Man_t, iObj0 int32) int32 {
    var w int32;
    var nWords int32 = Gia_ObjSimWords(p);
    var pSim *word  = Gia_ObjSim( p, iObj0 );
    if pSim[0] & 1 {
        for  w = 0; w < nWords; w++  {
            if ~pSim[w] {
                return 0; }
        }
    
    } else {
        for  w = 0; w < nWords; w++  {
            if pSim[w] {
                return 0; }
        }
    }
    return 1;
}
func Ssc_GiaSimAreEqual(p *Gia_Man_t, iObj0 int32, iObj1 int32) int32 {
    var w int32;
    var nWords int32 = Gia_ObjSimWords(p);
    var pSim0 *word  = Gia_ObjSim( p, iObj0 );
    var pSim1 *word  = Gia_ObjSim( p, iObj1 );
    if (pSim0[0] & 1) != (pSim1[0] & 1) {
        for  w = 0; w < nWords; w++  {
            if pSim0[w] != ~pSim1[w] {
                return 0; }
        }
    
    } else {
        for  w = 0; w < nWords; w++  {
            if pSim0[w] != pSim1[w] {
                return 0; }
        }
    }
    return 1;
}
func Ssc_GiaSimAreEqualBit(p *Gia_Man_t, iObj0 int32, iObj1 int32) int32 {
    var pObj0 *Gia_Obj_t  = Gia_ManObj( p, iObj0 );
    var pObj1 *Gia_Obj_t  = Gia_ManObj( p, iObj1 );
    return (pObj0.fPhase ^ pObj0.fMark0) == (pObj1.fPhase ^ pObj1.fMark0);
}

/**Function*************************************************************

  Synopsis    [Refines one equivalence class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaSimClassCreate(p *Gia_Man_t, vClass *Vec_Int_t) {
    var Repr int32 = GIA_VOID;
    var EntPrev int32 = -1;
    var Ent int32;
    var i int32;
    assert(Vec_IntSize(vClass) > 0, "Vec_IntSize(vClass) > 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vClass)) && (((Ent) = Vec_IntEntry(vClass, i)), 1); i++  {
        if i == 0 {
            Repr = Ent;
            Gia_ObjSetRepr( p, Ent, GIA_VOID );
            EntPrev = Ent;
        
        } else {
            assert(Repr < Ent, "Repr < Ent");
            Gia_ObjSetRepr( p, Ent, Repr );
            Gia_ObjSetNext( p, EntPrev, Ent );
            EntPrev = Ent;
        }
    }
    Gia_ObjSetNext( p, EntPrev, 0 );
}

/**Function*************************************************************

  Synopsis    [Refines one equivalence class using individual bit-pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaSimClassRefineOneBit(p *Gia_Man_t, i int32) int32 {
    var pObj *Gia_Obj_t;
    var Ent int32;
    assert(Gia_ObjIsHead( p, i ), "Gia_ObjIsHead( p, i )");
    Vec_IntClear( p.vClassOld );
    Vec_IntClear( p.vClassNew );
    Vec_IntPush( p.vClassOld, i );
    pObj = Gia_ManObj(p, i);
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), Ent = Gia_ObjNext(p, i); Ent > 0; Ent = Gia_ObjNext(p, Ent)  {
        if Ssc_GiaSimAreEqualBit( p, i, Ent ) != 0 {
            Vec_IntPush( p.vClassOld, Ent );
        } else {
            Vec_IntPush( p.vClassNew, Ent ); }
    }
    if Vec_IntSize( p.vClassNew ) == 0 {
        return 0; }
    Ssc_GiaSimClassCreate( p, p.vClassOld );
    Ssc_GiaSimClassCreate( p, p.vClassNew );
    return 1;
}


/**Function*************************************************************

  Synopsis    [Refines one class using simulation patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaSimClassRefineOne(p *Gia_Man_t, i int32) int32 {
    var pObj *Gia_Obj_t;
    var Ent int32;
    assert(Gia_ObjIsHead( p, i ), "Gia_ObjIsHead( p, i )");
    Vec_IntClear( p.vClassOld );
    Vec_IntClear( p.vClassNew );
    Vec_IntPush( p.vClassOld, i );
    pObj = Gia_ManObj(p, i);
    // inlining of macro Gia_ClassForEachObj1
    for  assert(Gia_ObjIsHead(p, i), "Gia_ObjIsHead(p, i)"), Ent = Gia_ObjNext(p, i); Ent > 0; Ent = Gia_ObjNext(p, Ent)  {
        if Ssc_GiaSimAreEqual( p, i, Ent ) != 0 {
            Vec_IntPush( p.vClassOld, Ent );
        } else {
            Vec_IntPush( p.vClassNew, Ent ); }
    }
    if Vec_IntSize( p.vClassNew ) == 0 {
        return 0; }
    Ssc_GiaSimClassCreate( p, p.vClassOld );
    Ssc_GiaSimClassCreate( p, p.vClassNew );
    if Vec_IntSize(p.vClassNew) > 1 {
        return 1 + Ssc_GiaSimClassRefineOne( p, Vec_IntEntry(p.vClassNew,0) ); }
    return 1;
}
func Ssc_GiaSimProcessRefined(p *Gia_Man_t, vRefined *Vec_Int_t) {
    var pTable *int32;
    var nTableSize int32;
    var i int32;
    var k int32;
    var Key int32;
    if Vec_IntSize(vRefined) == 0 {
        return; }
    nTableSize = Abc_PrimeCudd( 100 + Vec_IntSize(vRefined) / 3 );
    // inlining of macro ABC_CALLOC
    pTable = ((int *) calloc((size_t)(nTableSize), sizeof(int)));
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        assert(!(Ssc_GiaSimIsConst0( p, i ) != 0), "!Ssc_GiaSimIsConst0( p, i )");
        Key = Ssc_GiaSimHashKey( p, i, nTableSize );
        if pTable[Key] == 0 {
            assert(Gia_ObjRepr(p, i) == 0, "Gia_ObjRepr(p, i) == 0");
            assert(Gia_ObjNext(p, i) == 0, "Gia_ObjNext(p, i) == 0");
            Gia_ObjSetRepr( p, i, GIA_VOID );
        
        } else {
            Gia_ObjSetNext( p, pTable[Key], i );
            Gia_ObjSetRepr( p, i, Gia_ObjRepr(p, pTable[Key]) );
            if Gia_ObjRepr(p, i) == GIA_VOID {
                Gia_ObjSetRepr( p, i, pTable[Key] ); }
            assert(Gia_ObjRepr(p, i) > 0, "Gia_ObjRepr(p, i) > 0");
        }
        pTable[Key] = i;
    }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vRefined)) && (((i) = Vec_IntEntry(vRefined, k)), 1); k++  {
        if Gia_ObjIsHead( p, i ) {
            Ssc_GiaSimClassRefineOne( p, i ); }
    }
    // inlining of macro ABC_FREE
    ((pTable) ? (free((char *) (pTable)), (pTable) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Refines equivalence classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaClassesInit(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pReprs == nil, "p.pReprs == nil");
    // inlining of macro ABC_CALLOC
    p.pReprs = ((Gia_Rpr_t *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(Gia_Rpr_t)));
    // inlining of macro ABC_CALLOC
    p.pNexts = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int)));
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Gia_ObjSetRepr( p, i, Gia_ObjIsCand(pObj) ? 0 : GIA_VOID );
    }
    if p.vClassOld == nil {
        p.vClassOld = Vec_IntAlloc( 100 ); }
    if p.vClassNew == nil {
        p.vClassNew = Vec_IntAlloc( 100 ); }
}
func Ssc_GiaClassesRefine(p *Gia_Man_t) int32 {
    var vRefinedC *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    assert(p.pReprs != nil, "p.pReprs != nil");
    vRefinedC = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachCand
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsCand(pObj) {} else
        if Gia_ObjIsTail(p, i) {
            Counter += Ssc_GiaSimClassRefineOne( p, Gia_ObjRepr(p, i) );
        } else if Gia_ObjIsConst(p, i) && !(Ssc_GiaSimIsConst0(p, i) != 0) {
            Vec_IntPush( vRefinedC, i ); }
    }
    Ssc_GiaSimProcessRefined( p, vRefinedC );
    Counter += Vec_IntSize( vRefinedC );
    Vec_IntFree( vRefinedC );
    return Counter;
}


/**Function*************************************************************

  Synopsis    [Check if the pairs have been disproved.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssc_GiaClassesCheckPairs(p *Gia_Man_t, vDisPairs *Vec_Int_t) {
    var i int32;
    var iRepr int32;
    var iObj int32;
    var Result int32 = 1;
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vDisPairs)) && (((iRepr) = Vec_IntEntry(vDisPairs, i)), 1) && (((iObj) = Vec_IntEntry(vDisPairs, i+1)), 1); i += 2  {
        if iRepr == Gia_ObjRepr(p, iObj) {
            fmt.Printf("Pair (%d, %d) are still equivalent.\n", iRepr, iObj);
Result = 0; }
    }
//    if ( Result )
//        printf( "Classes are refined correctly.\n" );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

