package util

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [utilNam.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Manager for character strings.]

  Synopsis    [Manager for character strings.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: utilNam.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/


//#include <stdio.h>
//#include <string.h>
//#include <stdlib.h>
//#include <assert.h>

//#include "abc_global.h"
//#include "misc/vec/vec.h"
//#include "utilNam.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// this package maps non-empty character strings into natural numbers and back

// name manager
type Abc_Nam_t_ struct {
    // info storage for names
    nStore              int32;       // the size of allocated storage
    iHandle              int32;      // the current free handle
    pStore            string;       // storage for name objects
    // internal number mappings
    vInt2Handle        Vec_Int_t;  // mapping integers into handles
    vInt2Next        Vec_Int_t;    // mapping integers into nexts
    // hash table for names
    pBins             *int32;        // the hash table bins 
    nBins              int32;        // the number of bins 
    // manager recycling
    nRefs              int32;        // reference counter for the manager
    // internal buffer
    vBuffer        Vec_Str_t;      
}

func Abc_NamHandleToStr(p *Abc_Nam_t, h int32) string { return (char *)(p.pStore + h);                         }
func Abc_NamIntToHandle(p *Abc_Nam_t, i int32) int32 { return Vec_IntEntry(&p.vInt2Handle, i);                }
func Abc_NamIntToStr(p *Abc_Nam_t, i int32) string { return Abc_NamHandleToStr(p, Abc_NamIntToHandle(p,i));  }
func Abc_NamIntToNext(p *Abc_Nam_t, i int32) int32 { return Vec_IntEntry(&p.vInt2Next, i);                  }
func Abc_NamIntToNextP(p *Abc_Nam_t, i int32) *int32 { return Vec_IntEntryP(&p.vInt2Next, i);                 }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamStart(nObjs int32, nAveSize int32) *Abc_Nam_t {
    var p *Abc_Nam_t;
    if nObjs == 0 {
        nObjs = 16; }
    // inlining of macro ABC_CALLOC
    p = &Abc_Nam_t{};
    p.nStore      = ((nObjs * (nAveSize + 1) + 16) / 4) * 4;
    // inlining of macro ABC_ALLOC
    p.pStore      = ((char *) malloc(sizeof(char) * (size_t)(p.nStore)));
    p.nBins       = Abc_PrimeCudd( nObjs );
    // inlining of macro ABC_CALLOC
    p.pBins       = ((int *) calloc((size_t)(p.nBins), sizeof(int)));
    // 0th object is unused
    Vec_IntGrow( &p.vInt2Handle, nObjs );  Vec_IntPush( &p.vInt2Handle, -1 );
    Vec_IntGrow( &p.vInt2Next,   nObjs );  Vec_IntPush( &p.vInt2Next,   -1 );
    p.iHandle     = 4;
    memset( p.pStore, 0, 4 );
//Abc_Print( 1, "Starting nam with %d bins.\n", p->nBins );
    // start reference counting
    p.nRefs       = 1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamStop(p *Abc_Nam_t) {
//Abc_Print( 1, "Starting nam with %d bins.\n", p->nBins );
    Vec_StrErase( &p.vBuffer );
    Vec_IntErase( &p.vInt2Handle );
    Vec_IntErase( &p.vInt2Next );
    // inlining of macro ABC_FREE
    ((p.pStore) ? (free((char *) (p.pStore)), (p.pStore) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prints manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamPrint(p *Abc_Nam_t, pFileName string) {
    var pFile *os.File  = pFileName ? os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666) : os.Stdout;
    var h, i int32;
    if pFile == nil { fmt.Printf("Count node open file %s\n", pFileName); return; }
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(&p.vInt2Handle)) && (((h) = Vec_IntEntry(&p.vInt2Handle, i)), 1); i++  {
        fmt.Fprintf(pFile, "%8d = %s\n", i, Abc_NamHandleToStr(p, h));
    }
    if pFile != os.Stdout {
        _ = (pFile).Close(); }
}

/**Function*************************************************************

  Synopsis    [References the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamRef(p *Abc_Nam_t) *Abc_Nam_t {
    p.nRefs++;
    return p;
}

/**Function*************************************************************

  Synopsis    [Dereferences the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamDeref(p *Abc_Nam_t) {
    if p == nil {
        return; }
    if --p.nRefs == 0 {
        Abc_NamStop( p ); }
}

/**Function*************************************************************

  Synopsis    [Returns the number of used entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamObjNumMax(p *Abc_Nam_t) int32 {
    return Vec_IntSize(&p.vInt2Handle);
}

/**Function*************************************************************

  Synopsis    [Reports memory usage of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamMemUsed(p *Abc_Nam_t) int32 {
    if p == nil {
        return 0; }
    return sizeof(Abc_Nam_t) + p.iHandle + sizeof(int) * p.nBins + 
        sizeof(int) * (p.vInt2Handle.nSize + p.vInt2Next.nSize);
}

/**Function*************************************************************

  Synopsis    [Reports memory usage of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamMemAlloc(p *Abc_Nam_t) int32 {
    if p == nil {
        return 0; }
    return sizeof(Abc_Nam_t) + p.nStore + sizeof(int) * p.nBins + 
        sizeof(int) * (p.vInt2Handle.nCap + p.vInt2Next.nCap);
}

/**Function*************************************************************

  Synopsis    [Computes hash value of the C-string.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_NamStrHash_s_FPrimes [128]int32 = [128]int32{ 
        1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 
        1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997, 
        2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 
        2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 
        3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671, 
        3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 
        4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871, 
        4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471, 
        5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 
        6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689, 
        6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 
        7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 
        8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
    };
func Abc_NamStrHash(pStr string, pLim string, nTableSize int32) int32 {
    
    var i, uHash uint32;
    if pLim {
        for  uHash, i = 0, 0; pStr+i < pLim; i++  {
            if i & 1 { 
                uHash *= pStr[i] * spStatic_Abc_NamStrHash_s_FPrimes[i & 0x7F];
            } else {
                uHash ^= pStr[i] * spStatic_Abc_NamStrHash_s_FPrimes[i & 0x7F]; }
        }
    
    } else {
        for  uHash, i = 0, 0; pStr[i]; i++  {
            if i & 1 { 
                uHash *= pStr[i] * spStatic_Abc_NamStrHash_s_FPrimes[i & 0x7F];
            } else {
                uHash ^= pStr[i] * spStatic_Abc_NamStrHash_s_FPrimes[i & 0x7F]; }
        }
    }
    return uHash % nTableSize;
}
// https://en.wikipedia.org/wiki/Jenkins_hash_function
func Abc_NamStrHash2(pStr string, pLim string, nTableSize int32) int32 {
    var nSize int32;
    if pLim {
        nSize = pLim - pStr;
    } else {
        nSize = -1;
    }
    var i int32; var hash uint32;
    for  i != nSize && pStr[i]  {
        hash += pStr[i++];
        hash += hash << 10;
        hash ^= hash >> 6;
    }
    hash += hash << 3;
    hash ^= hash >> 11;
    hash += hash << 15;
    return int32((hash % nTableSize));
}

/**Function*************************************************************

  Synopsis    [Returns place where this string is, or should be.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamStrcmp(pStr string, pLim string, pThis string) int32 {
    if pLim {
        for  pStr < pLim  {
            if *pStr++ != *pThis++ {
                return 1; }
        }
        return *pThis != '\0';
    
    } else {
        for  *pStr  {
            if *pStr++ != *pThis++ {
                return 1; }
        }
        return *pThis != '\0';
    }
}
func Abc_NamStrHashFind(p *Abc_Nam_t, pStr string, pLim string) *int32 {
    var pThis string;
    var pPlace *int32  = (int *)(p.pBins + Abc_NamStrHash( pStr, pLim, p.nBins ));
    assert(*pStr, "*pStr");
    for  pThis = (*pPlace)? Abc_NamIntToStr(p, *pPlace) : nil; 
          pThis;    pPlace = Abc_NamIntToNextP(p, *pPlace), 
          pThis = (*pPlace)? Abc_NamIntToStr(p, *pPlace) : nil  {
              if !Abc_NamStrcmp( (char *)pStr, (char *)pLim, pThis ) {
                  break; }
    }
    return pPlace;
}

/**Function*************************************************************

  Synopsis    [Resizes the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamStrHashResize(p *Abc_Nam_t) {
    var vInt2HandleOld Vec_Int_t;  var pThis string; 
    var piPlace *int32;
    var pBinsOld *int32;
    var iHandleOld int32;
    var i int32;//, clk = Abc_Clock();
    assert(p.pBins != nil, "p.pBins != nil");
//    Abc_Print( 1, "Resizing names manager hash table from %6d to %6d. ", p->nBins, Abc_PrimeCudd( 3 * p->nBins ) );
    // replace the table
    pBinsOld = p.pBins;
    p.nBins = Abc_PrimeCudd( 3 * p.nBins ); 
    // inlining of macro ABC_CALLOC
    p.pBins = ((int *) calloc((size_t)(p.nBins), sizeof(int)));
    // replace the handles array
    vInt2HandleOld = p.vInt2Handle;
    Vec_IntZero( &p.vInt2Handle );
    Vec_IntGrow( &p.vInt2Handle, 2 * Vec_IntSize(&vInt2HandleOld) ); 
    Vec_IntPush( &p.vInt2Handle, -1 );
    Vec_IntClear( &p.vInt2Next );    Vec_IntPush( &p.vInt2Next, -1 );
    // rehash the entries from the old table
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(&vInt2HandleOld)) && (((iHandleOld) = Vec_IntEntry(&vInt2HandleOld, i)), 1); i++  {
        pThis   = Abc_NamHandleToStr( p, iHandleOld );
        piPlace = Abc_NamStrHashFind( p, pThis, nil );
        assert(*piPlace == 0, "*piPlace == 0");
        *piPlace = Vec_IntSize( &p.vInt2Handle );
        assert(Vec_IntSize( &p.vInt2Handle ) == i, "Vec_IntSize( &p.vInt2Handle ) == i");
        Vec_IntPush( &p.vInt2Handle, iHandleOld );
        Vec_IntPush( &p.vInt2Next,   0 );
    }
    Vec_IntErase( &vInt2HandleOld );
    // inlining of macro ABC_FREE
    ((pBinsOld) ? (free((char *) (pBinsOld)), (pBinsOld) = 0) : 0);
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    [Returns the index of the string in the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamStrFind(p *Abc_Nam_t, pStr string) int32 {
    return *Abc_NamStrHashFind( p, pStr, nil );
}
func Abc_NamStrFindLim(p *Abc_Nam_t, pStr string, pLim string) int32 {
    return *Abc_NamStrHashFind( p, pStr, pLim );
}

/**Function*************************************************************

  Synopsis    [Finds or adds the given name to storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamStrFindOrAdd(p *Abc_Nam_t, pStr string, pfFound *int32) int32 {
    var i, iHandleNew int32;
    var piPlace *int32;
    if !(pStr[0] != '\\' || pStr[len(pStr)-1] == ' ') {
        for  i = len(pStr) - 1; i >= 0; i--  {
            if *pStr == ' ' {
                break; }
        }
        assert(i < int32(len(pStr)), "i < int32(strlen(pStr))");
    }
    piPlace = Abc_NamStrHashFind( p, pStr, nil );
    if *piPlace {
        if pfFound != nil {
            *pfFound = 1; }
        return *piPlace;
    }
    if pfFound != nil {
        *pfFound = 0; }
    iHandleNew = p.iHandle + len(pStr) + 1;
    for  p.nStore < iHandleNew  {
        p.nStore *= 3;
        p.nStore /= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pStore {
            p.pStore = ((char *) realloc((char *)(p.pStore), sizeof(char) * (size_t)(p.nStore)));
        } else {
            p.pStore = ((char *) malloc(sizeof(char) * (size_t)(p.nStore)));
        }
    }
    assert(p.nStore >= iHandleNew, "p.nStore >= iHandleNew");
    // create new handle
    *piPlace = Vec_IntSize( &p.vInt2Handle );
    strcpy( Abc_NamHandleToStr( p, p.iHandle ), pStr );
    Vec_IntPush( &p.vInt2Handle, p.iHandle );
    Vec_IntPush( &p.vInt2Next, 0 );
    p.iHandle = iHandleNew;
    // extend the hash table
    if Vec_IntSize(&p.vInt2Handle) > 2 * p.nBins {
        Abc_NamStrHashResize( p ); }
    return Vec_IntSize(&p.vInt2Handle) - 1;
}
func Abc_NamStrFindOrAddLim(p *Abc_Nam_t, pStr string, pLim string, pfFound *int32) int32 {
    var iHandleNew int32;
    var piPlace *int32;
    var pStore string;
    assert(pStr < pLim, "pStr < pLim");
    piPlace = Abc_NamStrHashFind( p, pStr, pLim );
    if *piPlace {
        if pfFound != nil {
            *pfFound = 1; }
        return *piPlace;
    }
    if pfFound != nil {
        *pfFound = 0; }
    iHandleNew = p.iHandle + (pLim - pStr) + 1;
    for  p.nStore < iHandleNew  {
        p.nStore *= 3;
        p.nStore /= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pStore {
            p.pStore = ((char *) realloc((char *)(p.pStore), sizeof(char) * (size_t)(p.nStore)));
        } else {
            p.pStore = ((char *) malloc(sizeof(char) * (size_t)(p.nStore)));
        }
    }
    assert(p.nStore >= iHandleNew, "p.nStore >= iHandleNew");
    // create new handle
    *piPlace = Vec_IntSize( &p.vInt2Handle );
    pStore = Abc_NamHandleToStr( p, p.iHandle );
    strncpy( pStore, pStr, pLim - pStr );
    pStore[pLim - pStr] = 0;
    Vec_IntPush( &p.vInt2Handle, p.iHandle );
    Vec_IntPush( &p.vInt2Next, 0 );
    p.iHandle = iHandleNew;
    // extend the hash table
    if Vec_IntSize(&p.vInt2Handle) > 2 * p.nBins {
        Abc_NamStrHashResize( p ); }
    return Vec_IntSize(&p.vInt2Handle) - 1;
}
int Abc_NamStrFindOrAddF( Abc_Nam_t * p, char * format, ...  )
{
    var nAdded int32;
    var nSize int32 = 1000; 
    va_list args;  va_start( args, format );
    Vec_StrGrow( &p.vBuffer, Vec_StrSize(&p.vBuffer) + nSize );
    nAdded = vsnprintf( Vec_StrLimit(&p.vBuffer), nSize, format, args );
    if nAdded > nSize {
        Vec_StrGrow( &p.vBuffer, Vec_StrSize(&p.vBuffer) + nAdded + nSize );
        nSize = vsnprintf( Vec_StrLimit(&p.vBuffer), nAdded, format, args );
        assert(nSize == nAdded, "nSize == nAdded");
    }
    va_end( args );
    return Abc_NamStrFindOrAddLim( p, Vec_StrLimit(&p.vBuffer), Vec_StrLimit(&p.vBuffer) + nAdded, nil );
}

/**Function*************************************************************

  Synopsis    [Returns name from name ID.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamStr(p *Abc_Nam_t, NameId int32) string {
    if NameId > 0 {
        return Abc_NamIntToStr(p, NameId);
    } else {
        return nil;
    }
}

/**Function*************************************************************

  Synopsis    [Returns internal buffer.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamBuffer(p *Abc_Nam_t) *Vec_Str_t {
    Vec_StrClear(&p.vBuffer);
    return &p.vBuffer;
}

/**Function*************************************************************

  Synopsis    [For each ID of the first manager, gives ID of the second one.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamComputeIdMap(p1 *Abc_Nam_t, p2 *Abc_Nam_t) *Vec_Int_t {
    var vMap *Vec_Int_t;
    var pThis string;
    var piPlace *int32;
    var iHandle1 int32;
    var i int32;
    if p1 == p2 {
        return Vec_IntStartNatural( Abc_NamObjNumMax(p1) ); }
    vMap = Vec_IntStart( Abc_NamObjNumMax(p1) );
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(&p1.vInt2Handle)) && (((iHandle1) = Vec_IntEntry(&p1.vInt2Handle, i)), 1); i++  {
        pThis = Abc_NamHandleToStr( p1, iHandle1 );
        piPlace = Abc_NamStrHashFind( p2, pThis, nil );
        Vec_IntWriteEntry( vMap, i, *piPlace );
//        Abc_Print( 1, "%d->%d  ", i, *piPlace );
    }
    return vMap;
}

/**Function*************************************************************

  Synopsis    [Returns the number of common names in the array.]

  Description [The array contains name IDs in the first manager.
  The procedure returns the number of entries that correspond to names
  in the first manager that appear in the second manager.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamReportCommon(vNameIds1 *Vec_Int_t, p1 *Abc_Nam_t, p2 *Abc_Nam_t) int32 {
    var i, Entry, Counter int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNameIds1)) && (((Entry) = Vec_IntEntry(vNameIds1, i)), 1); i++  {
        assert(Entry > 0 && Entry < Abc_NamObjNumMax(p1), "Entry > 0 && Entry < Abc_NamObjNumMax(p1)");
        Counter += (Abc_NamStrFind(p2, Abc_NamStr(p1, Entry)) > 0);
//        if ( Abc_NamStrFind(p2, Abc_NamStr(p1, Entry)) == 0 )
//            Abc_Print( 1, "unique name <%s>\n", Abc_NamStr(p1, Entry) );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns the name that appears in p1 does not appear in p2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NamReportUnique(vNameIds1 *Vec_Int_t, p1 *Abc_Nam_t, p2 *Abc_Nam_t) string {
    var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNameIds1)) && (((Entry) = Vec_IntEntry(vNameIds1, i)), 1); i++  {
        assert(Entry > 0 && Entry < Abc_NamObjNumMax(p1), "Entry > 0 && Entry < Abc_NamObjNumMax(p1)");
        if Abc_NamStrFind(p2, Abc_NamStr(p1, Entry)) == 0 {
            return Abc_NamStr(p1, Entry); }
    }
    return nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

