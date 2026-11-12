package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecHash.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Hashing integer pairs/triples into an integer.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecHash.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

   
type Hash_IntObj_t struct {
    iData0          int32;
    iData1          int32;
    iData2          int32;
    iNext          int32;
}

   
type Hash_IntMan_t struct {
    vTable   *Vec_Int_t;      // hash table
    vObjs   *Vec_Int_t;       // hash objects
    nRefs          int32;       // reference counter for the manager
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Hash_IntObj(p *Hash_IntMan_t, i int32) *Hash_IntObj_t { if i != 0 {
    return (Hash_IntObj_t *)Vec_IntEntryP(p.vObjs, 4*i);
} else {
    return nil;
}  }
func Hash_IntObjData0(p *Hash_IntMan_t, i int32) int32 { return Hash_IntObj(p, i).iData0;                                 }
func Hash_IntObjData1(p *Hash_IntMan_t, i int32) int32 { return Hash_IntObj(p, i).iData1;                                 }
func Hash_IntObjData2(p *Hash_IntMan_t, i int32) int32 { return Hash_IntObj(p, i).iData2;                                 }

func Hash_Int2ObjInc(p *Hash_IntMan_t, i int32) int32 { return Hash_IntObj(p, i).iData2++;                         }
func Hash_Int2ObjDec(p *Hash_IntMan_t, i int32) int32 { return --Hash_IntObj(p, i).iData2;                         }
func Hash_Int2ObjSetData2(p *Hash_IntMan_t, i int32, d int32) { Hash_IntObj(p, i).iData2 = d;                              }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Hashes pairs of intergers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_IntManStart(nSize int32) *Hash_IntMan_t {
    var p *Hash_IntMan_t;  nSize += 100;
    p = new(Hash_IntMan_t);
    p.vTable = Vec_IntStart( Abc_PrimeCudd(nSize) );
    p.vObjs  = Vec_IntAlloc( 4*nSize );
    Vec_IntFill( p.vObjs, 4, 0 );
    p.nRefs  = 1;
    return p;
}
func Hash_IntManStop(p *Hash_IntMan_t) {
    Vec_IntFree( p.vObjs );
    Vec_IntFree( p.vTable );
    ABC_FREE( p );
}
func Hash_IntManRef(p *Hash_IntMan_t) *Hash_IntMan_t {
    p.nRefs++;
    return p;
}
func Hash_IntManDeref(p *Hash_IntMan_t) {
    if p == nil {
        return; }
    if --p.nRefs == 0 {
        Hash_IntManStop( p ); }
}
func Hash_IntManEntryNum(p *Hash_IntMan_t) int32 {
    return Vec_IntSize(p.vObjs)/4 - 1;
}
func Hash_IntManProfile(p *Hash_IntMan_t) {
    var pObj *Hash_IntObj_t;
    var i, Count, Entry int32;
    Vec_IntForEachEntry( p.vTable, Entry, i )
    {
        Count = 0;
        for  pObj = Hash_IntObj( p, Entry ); (pObj != nil); pObj = Hash_IntObj( p, pObj.iNext )  {
            Count++;
        }
        fmt.Printf("%d ", Count);
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_Int2ManHash(iData0 int32, iData1 int32, nTableSize int32) int32 {
    return (4177 * uint32(iData0) + 7873 * uint32(iData1)) % uint32(nTableSize);
}
func Hash_Int2ManLookup(p *Hash_IntMan_t, iData0 int32, iData1 int32) *int32 {
    var pObj *Hash_IntObj_t;
    var pPlace *int32  = Vec_IntEntryP( p.vTable, Hash_Int2ManHash(iData0, iData1, Vec_IntSize(p.vTable)) );
    for  ; (pObj = Hash_IntObj(p, *pPlace)); pPlace = &pObj.iNext  {
        if pObj.iData0 == iData0 && pObj.iData1 == iData1 {
            return pPlace; }
    }
    assert(*pPlace == 0, "*pPlace == 0");
    return pPlace;
}
func Hash_Int2ManInsert(p *Hash_IntMan_t, iData0 int32, iData1 int32, iData2 int32) int32 {
    var pObj *Hash_IntObj_t;
    var i int32;
    var nObjs int32;
    var pPlace *int32;
    nObjs = Vec_IntSize(p.vObjs)/4;
    if nObjs > Vec_IntSize(p.vTable) {
//        printf( "Resizing...\n" );
        Vec_IntFill( p.vTable, Abc_PrimeCudd(2*Vec_IntSize(p.vTable)), 0 );
        for  i = 1; i < nObjs; i++  {
            pObj = Hash_IntObj( p, i );
            pObj.iNext = 0;
            pPlace = Hash_Int2ManLookup( p, pObj.iData0, pObj.iData1 );
            assert(*pPlace == 0, "*pPlace == 0");
            *pPlace = i;
        }
    }
    pPlace = Hash_Int2ManLookup( p, iData0, iData1 );
    if *pPlace {
        return *pPlace; }
    *pPlace = nObjs;
    Vec_IntPush( p.vObjs, iData0 );
    Vec_IntPush( p.vObjs, iData1 );
    Vec_IntPush( p.vObjs, iData2 );
    Vec_IntPush( p.vObjs, 0 );
    return nObjs;
}

/**Function*************************************************************

  Synopsis    [Hashes triples of intergers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_Int3ManHash(iData0 int32, iData1 int32, iData2 int32, nTableSize int32) int32 {
    return (4177 * uint32(iData0) + 7873 * uint32(iData1) + 1699 * uint32(iData2)) % uint32(nTableSize);
}
func Hsh_Int3ManLookup(p *Hash_IntMan_t, iData0 int32, iData1 int32, iData2 int32) *int32 {
    var pObj *Hash_IntObj_t;
    var pPlace *int32  = Vec_IntEntryP( p.vTable, Hsh_Int3ManHash(iData0, iData1, iData2, Vec_IntSize(p.vTable)) );
    for  ; (pObj = Hash_IntObj(p, *pPlace)); pPlace = &pObj.iNext  {
        if pObj.iData0 == iData0 && pObj.iData1 == iData1 && pObj.iData2 == iData2 {
            return pPlace; }
    }
    assert(*pPlace == 0, "*pPlace == 0");
    return pPlace;
}
func Hsh_Int3ManInsert(p *Hash_IntMan_t, iData0 int32, iData1 int32, iData2 int32) int32 {
    var pObj *Hash_IntObj_t;
    var i int32;
    var nObjs int32;
    var pPlace *int32;
    nObjs = Vec_IntSize(p.vObjs)/4;
    if nObjs > Vec_IntSize(p.vTable) {
//        printf( "Resizing...\n" );
        Vec_IntFill( p.vTable, Abc_PrimeCudd(2*Vec_IntSize(p.vTable)), 0 );
        for  i = 1; i < nObjs; i++  {
            pObj = Hash_IntObj( p, i );
            pObj.iNext = 0;
            pPlace = Hsh_Int3ManLookup( p, pObj.iData0, pObj.iData1, pObj.iData2 );
            assert(*pPlace == 0, "*pPlace == 0");
            *pPlace = i;
        }
    }
    pPlace = Hsh_Int3ManLookup( p, iData0, iData1, iData2 );
    if *pPlace {
        return *pPlace; }
    *pPlace = nObjs;
    Vec_IntPush( p.vObjs, iData0 );
    Vec_IntPush( p.vObjs, iData1 );
    Vec_IntPush( p.vObjs, iData2 );
    Vec_IntPush( p.vObjs, 0 );
    return nObjs;
}

/**Function*************************************************************

  Synopsis    [Test procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hash_IntManHashArrayTest() {
    var p *Hash_IntMan_t;
    var RetValue int32;

    p = Hash_IntManStart( 10 );

    RetValue = Hash_Int2ManInsert( p, 10, 11, 12 );
    assert((RetValue != 0), "RetValue");

    RetValue = Hash_Int2ManInsert( p, 20, 21, 22 );
    assert((RetValue != 0), "RetValue");

    RetValue = Hash_Int2ManInsert( p, 10, 11, 12 );
    assert(!(RetValue != 0), "!RetValue");
 
    Hash_IntManStop( p );
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_END



