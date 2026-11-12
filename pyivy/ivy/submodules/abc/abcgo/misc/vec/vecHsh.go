package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecHsh.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Hashing vector entries.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecHsh.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

   
type Hsh_Int1Man_t struct {
    vData   *Vec_Int_t;       // stored data
    vNext   *Vec_Int_t;       // next items
    vTable   *Vec_Int_t;      // hash table
}



   
type Hsh_IntObj_t struct {
    iData          int32;
    iNext          int32;
}

typedef union Hsh_IntObjWord_t_ Hsh_IntObjWord_t;
union Hsh_IntObjWord_t_
{
    Hsh_IntObj_t wObj;
    word         wWord;
};

   
type Hsh_IntMan_t struct {
    nSize          int32;       // data size
    vData   *Vec_Int_t;       // data storage
    vTable   *Vec_Int_t;      // hash table
    vObjs   *Vec_Wrd_t;       // hash objects
}



   
type Hsh_VecObj_t struct {
    nSize          int32;
    iNext          int32;
    pArray [0]int32;
}

   
type Hsh_VecMan_t struct {
    vTable   *Vec_Int_t;      // hash table
    vData   *Vec_Int_t;       // data storage
    vMap   *Vec_Int_t;        // mapping entries into data;
    vTemp    Vec_Int_t;       // temporary array
    vTemp1    Vec_Int_t;      // temporary array
    vTemp2    Vec_Int_t;      // temporary array
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Hsh_IntData(p *Hsh_IntMan_t, iData int32) *uint32 { return (unsigned *)Vec_IntEntryP( p.vData, p.nSize * iData );             }
func Hsh_IntObj(p *Hsh_IntMan_t, iObj int32) *Hsh_IntObj_t { if iObj == -1 {
    return nil;
} else {
    return (Hsh_IntObj_t *)Vec_WrdEntryP( p.vObjs, iObj );
} }
func Hsh_IntWord(iData int32, iNext int32) word { var Obj Hsh_IntObjWord_t  = Hsh_IntObjWord_t{ {iData, iNext} }; return Obj.wWord;                }

func Hsh_VecObj(p *Hsh_VecMan_t, i int32) *Hsh_VecObj_t { if i == -1 {
    return nil;
} else {
    return (Hsh_VecObj_t *)Vec_IntEntryP(p.vData, Vec_IntEntry(p.vMap, i));
}  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Hashing data entries composed of nSize integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_Int1ManStart(nEntries int32) *Hsh_Int1Man_t {
    var p *Hsh_Int1Man_t;
    p = new(Hsh_Int1Man_t);
    p.vData  = Vec_IntAlloc( nEntries );
    p.vNext  = Vec_IntAlloc( nEntries );
    p.vTable = Vec_IntStartFull( Abc_PrimeCudd(nEntries) );
    return p;
}
func Hsh_Int1ManStop(p *Hsh_Int1Man_t) {
    Vec_IntFree( p.vData );
    Vec_IntFree( p.vNext );
    Vec_IntFree( p.vTable );
    ABC_FREE( p );
}
func Hsh_Int1ManEntryNum(p *Hsh_Int1Man_t) int32 {
    return Vec_IntSize(p.vData);
}


/**Function*************************************************************

  Synopsis    [Hashing individual integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// https://en.wikipedia.org/wiki/Jenkins_hash_function
func Hsh_Int1ManHash(Data int32, nTableSize int32) int32 {
    var i, hash uint32;
    unsigned char * pDataC = (unsigned char *)&Data;
    for  i = 0; i < 4; i++  {
        hash += pDataC[i];
        hash += hash << 10;
        hash ^= hash >> 6;
    }
    hash += hash << 3;
    hash ^= hash >> 11;
    hash += hash << 15;
    return int32((hash % nTableSize));
}
func Hsh_Int1ManLookupInt(p *Hsh_Int1Man_t, Data int32) *int32 {
    var Item int32;
    var pPlace *int32 = Vec_IntEntryP( p.vTable, Hsh_Int1ManHash(Data, Vec_IntSize(p.vTable)) );
    for  Item = *pPlace; Item >= 0; Item = Vec_IntEntry(p.vNext, Item)  {
        if Vec_IntEntry(p.vData, Item) == int32(Data) {
            return pPlace; }
    }
    assert(Item == -1, "Item == -1");
    return pPlace;
}
func Hsh_Int1ManLookup(p *Hsh_Int1Man_t, Data int32) int32 {
    return *Hsh_Int1ManLookupInt(p, Data);
}
func Hsh_Int1ManAdd(p *Hsh_Int1Man_t, Data int32) int32 {
    var i int32;
    var Entry int32;
    var pPlace *int32;
    if Vec_IntSize(p.vData) > Vec_IntSize(p.vTable) {
        Vec_IntFill( p.vTable, Abc_PrimeCudd(2*Vec_IntSize(p.vTable)), -1 );
        Vec_IntForEachEntry( p.vData, Entry, i )
        {
            pPlace = Vec_IntEntryP( p.vTable, Hsh_Int1ManHash(Entry, Vec_IntSize(p.vTable)) );
            Vec_IntWriteEntry( p.vNext, i, *pPlace ); *pPlace = i;
        }
    }
    pPlace = Hsh_Int1ManLookupInt( p, Data );
    if *pPlace == -1 {
        *pPlace = Vec_IntSize(p.vData);
        Vec_IntPush( p.vData, Data );
        Vec_IntPush( p.vNext, -1 );
    }
    return *pPlace;
}

/**Function*************************************************************

  Synopsis    [Test procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_Int1ManHashArrayTest() {
    var p *Hsh_Int1Man_t  = Hsh_Int1ManStart( 10 );

    fmt.Printf("Entry = %d.   Hashed = %d.\n", 5, Hsh_Int1ManAdd(p, 5 ));
    fmt.Printf("Entry = %d.   Hashed = %d.\n", 7, Hsh_Int1ManAdd(p, 7 ));
    fmt.Printf("Entry = %d.   Hashed = %d.\n", 7, Hsh_Int1ManAdd(p, 7 ));
    fmt.Printf("Entry = %d.   Hashed = %d.\n", 7, Hsh_Int1ManAdd(p, 7 ));
    fmt.Printf("Entry = %d.   Hashed = %d.\n", 8, Hsh_Int1ManAdd(p, 8 ));
    fmt.Printf("Entry = %d.   Hashed = %d.\n", 5, Hsh_Int1ManAdd(p, 5 ));
    fmt.Printf("Entry = %d.   Hashed = %d.\n", 3, Hsh_Int1ManAdd(p, 3 ));
    fmt.Printf("Entry = %d.   Hashed = %d.\n", 4, Hsh_Int1ManAdd(p, 4 ));

    fmt.Printf("Size = %d.\n", Hsh_Int1ManEntryNum(p));

    Hsh_Int1ManStop( p );
}


/**Function*************************************************************

  Synopsis    [Hashing data entries composed of nSize integers.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_IntManStart(vData *Vec_Int_t, nSize int32, nEntries int32) *Hsh_IntMan_t {
    var p *Hsh_IntMan_t;
    p = new(Hsh_IntMan_t);
    p.nSize  = nSize;
    p.vData  = vData;
    p.vTable = Vec_IntStartFull( Abc_PrimeCudd(nEntries) );
    p.vObjs  = Vec_WrdAlloc( nEntries );
    return p;
}
func Hsh_IntManStop(p *Hsh_IntMan_t) {
    Vec_IntFree( p.vTable );
    Vec_WrdFree( p.vObjs );
    ABC_FREE( p );
}
func Hsh_IntManEntryNum(p *Hsh_IntMan_t) int32 {
    return Vec_WrdSize(p.vObjs);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// https://en.wikipedia.org/wiki/Jenkins_hash_function
func Hsh_IntManHash(pData *uint32, nSize int32, nTableSize int32) int32 {
    var i int32; var hash uint32;
    unsigned char * pDataC = (unsigned char *)pData;
    nSize <<= 2;
    for  i != nSize  {
        hash += pDataC[i++];
        hash += hash << 10;
        hash ^= hash >> 6;
    }
    hash += hash << 3;
    hash ^= hash >> 11;
    hash += hash << 15;
    return int32((hash % nTableSize));
}
var spStatic_Hsh_IntManHash2_s_Primes [7]int32 = [7]int32{ 4177, 5147, 5647, 6343, 7103, 7873, 8147 };
func Hsh_IntManHash2(pData *uint32, nSize int32, nTableSize int32) int32 {
    
    unsigned char * pDataC = (unsigned char *)pData;
    var c int32;
    var nChars int32 = nSize * 4;
    var Key uint32;
    for  c = 0; c < nChars; c++  {
        Key += pDataC[c] * spStatic_Hsh_IntManHash2_s_Primes[c % 7];
    }
    return int32((Key % nTableSize));
}
func Hsh_IntManLookup(p *Hsh_IntMan_t, pData *uint32) *int32 {
    var pObj *Hsh_IntObj_t;
    var pPlace *int32  = Vec_IntEntryP( p.vTable, Hsh_IntManHash(pData, p.nSize, Vec_IntSize(p.vTable)) );
    for  ; (pObj = Hsh_IntObj(p, *pPlace)); pPlace = &pObj.iNext  {
        if !memcmp( pData, Hsh_IntData(p, pObj.iData), sizeof(int) * (size_t)p.nSize ) {
            return pPlace; }
    }
    assert(*pPlace == -1, "*pPlace == -1");
    return pPlace;
}
func Hsh_IntManAdd(p *Hsh_IntMan_t, iData int32) int32 {
    var i int32;
    var pPlace *int32;
    if Vec_WrdSize(p.vObjs) > Vec_IntSize(p.vTable) {
        Vec_IntFill( p.vTable, Abc_PrimeCudd(2*Vec_IntSize(p.vTable)), -1 );
        for  i = 0; i < Vec_WrdSize(p.vObjs); i++  {
            pPlace = Vec_IntEntryP( p.vTable, Hsh_IntManHash(Hsh_IntData(p, Hsh_IntObj(p, i).iData), p.nSize, Vec_IntSize(p.vTable)) );
            Hsh_IntObj(p, i).iNext = *pPlace;  *pPlace = i;
        }
    }
    pPlace = Hsh_IntManLookup( p, Hsh_IntData(p, iData) );
    if *pPlace == -1 {
        *pPlace = Vec_WrdSize(p.vObjs);
        Vec_WrdPush( p.vObjs, Hsh_IntWord(iData, -1) );
        return Vec_WrdSize(p.vObjs) - 1;
    }
    return (word *)Hsh_IntObj(p, *pPlace) - Vec_WrdArray(p.vObjs);
}

/**Function*************************************************************

  Synopsis    [Hashes data by value.]

  Description [Array vData contains data entries, each of 'nSize' integers.
  The resulting array contains the indexes of unique data entries.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_IntManHashArray(vData *Vec_Int_t, nSize int32) *Vec_Int_t {
    var p *Hsh_IntMan_t;
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32;
    var nEntries int32 = Vec_IntSize(vData) / nSize;
    assert(Vec_IntSize(vData) % nSize == 0, "Vec_IntSize(vData) % nSize == 0");
    p = Hsh_IntManStart( vData, nSize, nEntries );
    for  i = 0; i < nEntries; i++  {
        Vec_IntPush( vRes, Hsh_IntManAdd(p, i) );
    }
    Hsh_IntManStop( p );
    return vRes;
}
func Hsh_WrdManHashArray(vDataW *Vec_Wrd_t, nSize int32) *Vec_Int_t {
    var p *Hsh_IntMan_t;
    var Data Vec_Int_t  = Vec_Int_t{ 2*Vec_WrdCap(vDataW), 2*Vec_WrdSize(vDataW), (int *)Vec_WrdArray(vDataW) };
    var vData *Vec_Int_t  = &Data;
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var i int32;
    var nEntries int32 = Vec_IntSize(vData) / (2*nSize);
    assert(Vec_IntSize(vData) % (2*nSize) == 0, "Vec_IntSize(vData) % (2*nSize) == 0");
    p = Hsh_IntManStart( vData, (2*nSize), nEntries );
    for  i = 0; i < nEntries; i++  {
        Vec_IntPush( vRes, Hsh_IntManAdd(p, i) );
    }
    Hsh_IntManStop( p );
    return vRes;
}
func Hsh_WrdManHashArrayStart(vDataW *Vec_Wrd_t, nSize int32) *Hsh_IntMan_t {
    var p *Hsh_IntMan_t;
    var i int32;
    var nEntries int32 = Vec_WrdSize(vDataW) / nSize;
    var vData *Vec_Int_t  = Vec_IntAlloc( 2*Vec_WrdSize(vDataW) );
    memcpy( Vec_IntArray(vData), Vec_WrdArray(vDataW), sizeof(word)*(size_t)Vec_WrdSize(vDataW) );
    vData.nSize = 2*Vec_WrdSize(vDataW);
/*
    for ( i = 0; i < 30; i++ )
    {
        extern void Extra_PrintHex( FILE * pFile, unsigned * pTruth, int nVars );
        Extra_PrintHex( stdout, (unsigned *) Vec_WrdEntryP(vDataW, i), 6 );  printf( "  " );
        Kit_DsdPrintFromTruth( (unsigned *) Vec_WrdEntryP(vDataW, i), 6 );   printf( "\n" );
    }
*/
    assert(Vec_IntSize(vData) % (2*nSize) == 0, "Vec_IntSize(vData) % (2*nSize) == 0");
    p = Hsh_IntManStart( vData, (2*nSize), nEntries );
    for  i = 0; i < nEntries; i++  {
        Hsh_IntManAdd( p, i );
    }
    assert(Vec_WrdSize(p.vObjs) == nEntries, "Vec_WrdSize(p.vObjs) == nEntries");
    return p;
}

/**Function*************************************************************

  Synopsis    [Test procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_IntManHashArrayTest() {
    var vData *Vec_Int_t  = Vec_IntAlloc( 10 );
    var vRes *Vec_Int_t;
    Vec_IntPush( vData, 12 );
    Vec_IntPush( vData, 17 );
    Vec_IntPush( vData, 13 );
    Vec_IntPush( vData, 12 );
    Vec_IntPush( vData, 15 );
    Vec_IntPush( vData, 3 );
    Vec_IntPush( vData, 16 );
    Vec_IntPush( vData, 16 );
    Vec_IntPush( vData, 12 );
    Vec_IntPush( vData, 17 );
    Vec_IntPush( vData, 12 );
    Vec_IntPush( vData, 12 );

    vRes = Hsh_IntManHashArray( vData, 2 );

    Vec_IntPrint( vData );
    Vec_IntPrint( vRes );

    Vec_IntFree( vData );
    Vec_IntFree( vRes );
}





/**Function*************************************************************

  Synopsis    [Hashing integer arrays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_VecManStart(nEntries int32) *Hsh_VecMan_t {
    var p *Hsh_VecMan_t;
    p = new(Hsh_VecMan_t);
    p.vTable = Vec_IntStartFull( Abc_PrimeCudd(nEntries) );
    p.vData  = Vec_IntAlloc( nEntries * 4 );
    p.vMap   = Vec_IntAlloc( nEntries );
    return p;
}
func Hsh_VecManStop(p *Hsh_VecMan_t) {
    Vec_IntFree( p.vTable );
    Vec_IntFree( p.vData );
    Vec_IntFree( p.vMap );
    ABC_FREE( p );
}
func Hsh_VecReadArray(p *Hsh_VecMan_t, i int32) *int32 {
    return (int*)Hsh_VecObj(p, i) + 2;
}
func Hsh_VecReadEntry(p *Hsh_VecMan_t, i int32) *Vec_Int_t {
    var pObj *Hsh_VecObj_t  = Hsh_VecObj( p, i );
    p.vTemp.nSize = p.vTemp.nCap = pObj.nSize;
    p.vTemp.pArray = (int*)pObj + 2;
    return &p.vTemp;
}
func Hsh_VecReadEntry1(p *Hsh_VecMan_t, i int32) *Vec_Int_t {
    var pObj *Hsh_VecObj_t  = Hsh_VecObj( p, i );
    p.vTemp1.nSize = p.vTemp1.nCap = pObj.nSize;
    p.vTemp1.pArray = (int*)pObj + 2;
    return &p.vTemp1;
}
func Hsh_VecReadEntry2(p *Hsh_VecMan_t, i int32) *Vec_Int_t {
    var pObj *Hsh_VecObj_t  = Hsh_VecObj( p, i );
    p.vTemp2.nSize = p.vTemp2.nCap = pObj.nSize;
    p.vTemp2.pArray = (int*)pObj + 2;
    return &p.vTemp2;
}
func Hsh_VecSize(p *Hsh_VecMan_t) int32 {
    return Vec_IntSize(p.vMap);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Hsh_VecManHash_s_Primes [7]uint32 = [7]uint32{4177, 5147, 5647, 6343, 7103, 7873, 8147};
func Hsh_VecManHash(vVec *Vec_Int_t, nTableSize int32) int32 {
    
    var Key uint32;
    var i, Entry int32;
    Vec_IntForEachEntry( vVec, Entry, i )
        Key += uint32(Entry) * spStatic_Hsh_VecManHash_s_Primes[i % 7];
    return int32((Key % nTableSize));
}
func Hsh_VecManAdd(p *Hsh_VecMan_t, vVec *Vec_Int_t) int32 {
    var pObj *Hsh_VecObj_t;
    var i int32;
    var Ent int32;
    var pPlace *int32;
    if Vec_IntSize(p.vMap) > Vec_IntSize(p.vTable) {
        Vec_IntFill( p.vTable, Abc_PrimeCudd(2*Vec_IntSize(p.vTable)), -1 );
        for  i = 0; i < Vec_IntSize(p.vMap); i++  {
            pPlace = Vec_IntEntryP( p.vTable, Hsh_VecManHash(Hsh_VecReadEntry(p, i), Vec_IntSize(p.vTable)) );
            Hsh_VecObj(p, i).iNext = *pPlace; *pPlace = i;
        }
    }
    pPlace = Vec_IntEntryP( p.vTable, Hsh_VecManHash(vVec, Vec_IntSize(p.vTable)) );
    for  ; (pObj = Hsh_VecObj(p, *pPlace)); pPlace = &pObj.iNext  {
        if pObj.nSize == Vec_IntSize(vVec) && !memcmp( pObj.pArray, Vec_IntArray(vVec), sizeof(int) * (size_t)pObj.nSize ) {
            return *pPlace; }
    }
    *pPlace = Vec_IntSize(p.vMap);
    assert(Vec_IntSize(p.vData) % 2 == 0, "Vec_IntSize(p.vData) % 2 == 0");
    Vec_IntPush( p.vMap, Vec_IntSize(p.vData) );
    Vec_IntPush( p.vData, Vec_IntSize(vVec) );
    Vec_IntPush( p.vData, -1 );
    Vec_IntForEachEntry( vVec, Ent, i )
        Vec_IntPush( p.vData, Ent );
    if Vec_IntSize(vVec) & 1 {
        Vec_IntPush( p.vData, -1 ); }
    return Vec_IntSize(p.vMap) - 1;
}

/**Function*************************************************************

  Synopsis    [Test procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hsh_VecManHashTest() {
    var p *Hsh_VecMan_t;
    var vTemp *Vec_Int_t;
    var vRes *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var i int32;
    
    p = Hsh_VecManStart( 5 );
    for  i = 0; i < 20; i++  {
        vTemp = Vec_IntStartNatural( i );
        Vec_IntPush( vRes, Hsh_VecManAdd( p, vTemp ) );
        Vec_IntFree( vTemp );
    }
    for  ; i > 0; i--  {
        vTemp = Vec_IntStartNatural( i );
        Vec_IntPush( vRes, Hsh_VecManAdd( p, vTemp ) );
        Vec_IntFree( vTemp );
    }
    Vec_IntPrint( vRes );
    Vec_IntFree( vRes );

    Hsh_VecManStop( p );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_END



