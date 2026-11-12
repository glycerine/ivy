package vec

import "fmt"

/**CFile****************************************************************

  FileName    [vecStr.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Resizable arrays of characters.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecStr.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

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

         
type Vec_Str_t struct {
    nCap              int32;
    nSize              int32;
    pArray            string;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Vec_StrForEachEntry( vVec, Entry, i )                                               \
//     for ( i = 0; (i < Vec_StrSize(vVec)) && (((Entry) = Vec_StrEntry(vVec, i)), 1); i++ )   


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrAlloc(nCap int32) *Vec_Str_t {
    var p *Vec_Str_t;
    p = new(Vec_Str_t);
    if nCap > 0 && nCap < 16 {
        nCap = 16; }
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( char, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}
func Vec_StrAllocExact(nCap int32) *Vec_Str_t {
    var p *Vec_Str_t;
    assert(nCap >= 0, "nCap >= 0");
    p = new(Vec_Str_t);
    p.nSize  = 0;
    p.nCap   = nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( char, p.nCap );
    } else {
        p.pArray = nil;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocates a vector with the given size and cleans it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrStart(nSize int32) *Vec_Str_t {
    var p *Vec_Str_t;
    p = Vec_StrAlloc( nSize );
    p.nSize = nSize;
    memset( p.pArray, 0, sizeof(char) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrAllocArray(pArray string, nSize int32) *Vec_Str_t {
    var p *Vec_Str_t;
    p = new(Vec_Str_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = pArray;
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates the vector from an integer array of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrAllocArrayCopy(pArray string, nSize int32) *Vec_Str_t {
    var p *Vec_Str_t;
    p = new(Vec_Str_t);
    p.nSize  = nSize;
    p.nCap   = nSize;
    p.pArray = ABC_ALLOC( char, nSize );
    memcpy( p.pArray, pArray, sizeof(char) * (size_t)nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Duplicates the integer array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrDup(pVec *Vec_Str_t) *Vec_Str_t {
    var p *Vec_Str_t;
    p = new(Vec_Str_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    if p.nCap {
        p.pArray = ABC_ALLOC( char, p.nCap );
    } else {
        p.pArray = nil;
    }
    memcpy( p.pArray, pVec.pArray, sizeof(char) * (size_t)pVec.nSize );
    return p;
}

/**Function*************************************************************

  Synopsis    [Transfers the array into another vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrDupArray(pVec *Vec_Str_t) *Vec_Str_t {
    var p *Vec_Str_t;
    p = new(Vec_Str_t);
    p.nSize  = pVec.nSize;
    p.nCap   = pVec.nCap;
    p.pArray = pVec.pArray;
    pVec.nSize  = 0;
    pVec.nCap   = 0;
    pVec.pArray = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrZero(p *Vec_Str_t) {
    p.pArray = nil;
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_StrErase(p *Vec_Str_t) {
    ABC_FREE( p.pArray );
    p.nSize = 0;
    p.nCap = 0;
}
func Vec_StrFree(p *Vec_Str_t) {
    ABC_FREE( p.pArray );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrFreeP(p **Vec_Str_t) {
    if *p == nil {
        return; }
    ABC_FREE( (*p).pArray );
    ABC_FREE( (*p) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrReleaseArray(p *Vec_Str_t) string {
    var pArray string  = p.pArray;
    p.nCap = 0;
    p.nSize = 0;
    p.pArray = nil;
    return pArray;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrArray(p *Vec_Str_t) string {
    return p.pArray;
}
func Vec_StrLimit(p *Vec_Str_t) string {
    return p.pArray + p.nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSize(p *Vec_Str_t) int32 {
    return p.nSize;
}
func Vec_StrSetSize(p *Vec_Str_t, nSize int32) {
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrCap(p *Vec_Str_t) int32 {
    return p.nCap;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrMemory(p *Vec_Str_t) float64 {
    if !(p != nil) {
        return 0.0;
    } else {
        return 1.0 * sizeof(char) * (size_t)p.nCap + sizeof(Vec_Str_t);
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrEntry(p *Vec_Str_t, i int32) byte {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray[i];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrEntryP(p *Vec_Str_t, i int32) string {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    return p.pArray + i;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrWriteEntry(p *Vec_Str_t, i int32, Entry byte) {
    assert(i >= 0 && i < p.nSize, "i >= 0 && i < p.nSize");
    p.pArray[i] = Entry;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrEntryLast(p *Vec_Str_t) byte {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[p.nSize-1];
}

/**Function*************************************************************

  Synopsis    [Resizes the vector to the given capacity.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrGrow(p *Vec_Str_t, nCapMin int32) {
    if p.nCap >= nCapMin {
        return; }
    p.pArray = ABC_REALLOC( char, p.pArray, nCapMin ); 
    p.nCap   = nCapMin;
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrFill(p *Vec_Str_t, nSize int32, Fill byte) {
    var i int32;
    Vec_StrGrow( p, nSize );
    p.nSize = nSize;
    for  i = 0; i < p.nSize; i++  {
        p.pArray[i] = Fill;
    }
}

/**Function*************************************************************

  Synopsis    [Fills the vector with given number of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrFillExtra(p *Vec_Str_t, nSize int32, Fill byte) {
    var i int32;
    if nSize <= p.nSize {
        return; }
    if nSize > 2 * p.nCap {
        Vec_StrGrow( p, nSize );
    } else if nSize > p.nCap {
        Vec_StrGrow( p, 2 * p.nCap ); }
    for  i = p.nSize; i < nSize; i++  {
        p.pArray[i] = Fill;
    }
    p.nSize = nSize;
}

/**Function*************************************************************

  Synopsis    [Returns the entry even if the place not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrGetEntry(p *Vec_Str_t, i int32) byte {
    Vec_StrFillExtra( p, i + 1, 0 );
    return Vec_StrEntry( p, i );
}

/**Function*************************************************************

  Synopsis    [Inserts the entry even if the place does not exist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSetEntry(p *Vec_Str_t, i int32, Entry byte) {
    Vec_StrFillExtra( p, i + 1, 0 );
    Vec_StrWriteEntry( p, i, Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrShrink(p *Vec_Str_t, nSizeNew int32) {
    assert(p.nSize >= nSizeNew, "p.nSize >= nSizeNew");
    p.nSize = nSizeNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrClear(p *Vec_Str_t) {
    p.nSize = 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrPush(p *Vec_Str_t, Entry byte) {
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_StrGrow( p, 16 );
        } else {
            Vec_StrGrow( p, 2 * p.nCap ); }
    }
    p.pArray[p.nSize++] = Entry;
}
func Vec_StrPushBuffer(p *Vec_Str_t, pBuffer string, nSize int32) {
    if p.nSize + nSize > p.nCap {
        Vec_StrGrow( p, 2 * (p.nSize + nSize) ); }
    memcpy( p.pArray + p.nSize, pBuffer, (size_t)nSize );
    p.nSize += nSize;
}

/**Function*************************************************************

  Synopsis    [Returns the last entry and removes it from the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrPop(p *Vec_Str_t) byte {
    assert(p.nSize > 0, "p.nSize > 0");
    return p.pArray[--p.nSize];
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrIntPrint(p *Vec_Str_t) {
    var i int32;
    fmt.Printf("Vector has %d entries: {", Vec_StrSize(p));
    for  i = 0; i < Vec_StrSize(p); i++  {
        fmt.Printf(" %d", int32(Vec_StrEntry(p, i)));
    }
    fmt.Printf(" }\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrPrintNum(p *Vec_Str_t, Num int32) {
    var i int32;
    var Digits [16]byte;
    if Num == 0 {
        Vec_StrPush( p, '0' );
        return;
    }
    if Num < 0 {
        Vec_StrPush( p, '-' );
        Num = -Num;
    }
    for  i = 0; (Num != 0); Num /= 10,  i++  {
        Digits[i] = Num % 10;
    }
    for  i--; i >= 0; i--  {
        Vec_StrPush( p, byte(('0' + Digits[i])) );
    }
}
func Vec_StrPrintNumStar(p *Vec_Str_t, Num int32, nDigits int32) {
    var i int32;
    var Digits [16]byte ;
    if Num == 0 {
        for  i = 0; i < nDigits; i++  {
            Vec_StrPush( p, '0' );
        }
        return;
    }
    if Num < 0 {
        Vec_StrPush( p, '-' );
        Num = -Num;
        nDigits--;
    }
    for  i = 0; (Num != 0); Num /= 10,  i++  {
        Digits[i] = Num % 10;
    }
    for  i = Abc_MaxInt(i, nDigits)-1; i >= 0; i--  {
        Vec_StrPush( p, byte(('0' + Digits[i])) );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrPrintStr(p *Vec_Str_t, pStr string) {
    var i int32;
    var Length int32 = int32(len(pStr));
    for  i = 0; i < Length; i++  {
        Vec_StrPush( p, pStr[i] );
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
#ifdef WIN32
#define vsnprintf _vsnprintf
#endif

static inline char * Vec_StrPrintF( Vec_Str_t * p, char * format, ... )
{
    var nAdded int32;
    var nSize int32 = 1000; 
    va_list args;  va_start( args, format );
    Vec_StrGrow( p, Vec_StrSize(p) + nSize );
    nAdded = vsnprintf( Vec_StrLimit(p), nSize, format, args );
    if nAdded > nSize {
        Vec_StrGrow( p, Vec_StrSize(p) + nAdded + nSize );
        nSize = vsnprintf( Vec_StrLimit(p), nAdded, format, args );
        assert(nSize == nAdded, "nSize == nAdded");
    }
    p.nSize += nAdded;
    va_end( args );
    return Vec_StrLimit(p) - nAdded;
}

/**Function*************************************************************

  Synopsis    [Appends the string to the char vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrAppend(p *Vec_Str_t, pString string) {
    Vec_StrPrintStr( p, pString );
}
func Vec_StrCopy(p *Vec_Str_t, pString string) {
    Vec_StrClear( p );
    Vec_StrAppend( p, pString );
    Vec_StrPush( p, '\0' );
}

/**Function*************************************************************

  Synopsis    [Reverses the order of entries.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrReverseOrder(p *Vec_Str_t) {
    var i, Temp int32;
    for  i = 0; i < p.nSize/2; i++  {
        Temp = p.pArray[i];
        p.pArray[i] = p.pArray[p.nSize-1-i];
        p.pArray[p.nSize-1-i] = Temp;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSum(p *Vec_Str_t) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += int32(p.pArray[i]);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrCountEntry(p *Vec_Str_t, Entry byte) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] == Entry);
    }
    return Counter;
}
func Vec_StrCountLarger(p *Vec_Str_t, Entry byte) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] > Entry);
    }
    return Counter;
}
func Vec_StrCountSmaller(p *Vec_Str_t, Entry byte) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (p.pArray[i] < Entry);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrCountEntryLit(p *Vec_Str_t, Entry byte) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (Abc_Lit2Var(int32(p.pArray[i])) == Entry);
    }
    return Counter;
}
func Vec_StrCountLargerLit(p *Vec_Str_t, Entry byte) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (Abc_Lit2Var(int32(p.pArray[i])) > Entry);
    }
    return Counter;
}
func Vec_StrCountSmallerLit(p *Vec_Str_t, Entry byte) int32 {
    var i, Counter int32;
    for  i = 0; i < p.nSize; i++  {
        Counter += (Abc_Lit2Var(int32(p.pArray[i])) < Entry);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Compares two strings.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrEqual(p1 *Vec_Str_t, p2 *Vec_Str_t) int32 {
    var i int32;
    if p1.nSize != p2.nSize {
        return 0; }
    for  i = 0; i < p1.nSize; i++  {
        if p1.pArray[i] != p2.pArray[i] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSortCompare1(pp1 string, pp2 string) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 < *pp2 {
        return -1; }
    if *pp1 > *pp2 { //
        return 1; }
    return 0; //
}

/**Function*************************************************************

  Synopsis    [Comparison procedure for two clauses.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSortCompare2(pp1 string, pp2 string) int32 {
    // for some reason commenting out lines (as shown) led to crashing of the release version
    if *pp1 > *pp2 {
        return -1; }
    if *pp1 < *pp2 { //
        return 1; }
    return 0; //
}

/**Function*************************************************************

  Synopsis    [Sorting the entries by their integer value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrSort(p *Vec_Str_t, fReverse int32) {
    if fReverse != 0 { 
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(char), 
                        (int (*)(void *, void *)) Vec_StrSortCompare2 );
    } else {
        qsort( (void *)p.pArray, (size_t)p.nSize, sizeof(char), 
                        (int (*)(void *, void *)) Vec_StrSortCompare1 ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrCompareVec(p1 *Vec_Str_t, p2 *Vec_Str_t) int32 {
    if p1 == nil || p2 == nil {
        return (p1 != nil) - (p2 != nil); }
    if Vec_StrSize(p1) != Vec_StrSize(p2) {
        return Vec_StrSize(p1) - Vec_StrSize(p2); }
    return memcmp( Vec_StrArray(p1), Vec_StrArray(p2), (size_t)Vec_StrSize(p1) );
}


/**Function*************************************************************

  Synopsis    [Binary I/O for numbers (int/float/etc) and strings (char *).]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_StrPutI_ne(vOut *Vec_Str_t, Val int32) {
    var i int32;
//    for ( i = 0; i < 4; i++ )
    for  i = 3; i >= 0; i--  {
        Vec_StrPush( vOut, byte((Val >> (8*i))) );
    }
}
func Vec_StrGetI_ne(vOut *Vec_Str_t, pPos *int32) int32 {
    var i int32;
    var Val int32;
//    for ( i = 0; i < 4; i++ )
    for  i = 3; i >= 0; i--  {
        Val |= (int32((unsigned char))Vec_StrEntry(vOut, (*pPos)++) << (8*i));
    }
    return Val;
}

func Vec_StrPutI(vOut *Vec_Str_t, Val int32) {
    for  ; Val >= 0x80; Val >>= 7  {
        Vec_StrPush( vOut, (unsigned char)(Val | 0x80) );
    }
    Vec_StrPush( vOut, (unsigned char)Val );
}
func Vec_StrGetI(vOut *Vec_Str_t, pPos *int32) int32 {
    unsigned char ch;
    var i, Val int32;
    for  (ch = Vec_StrEntry(vOut, (*pPos)++)) & 0x80  {
        Val |= ((ch & 0x7f) << (7 * i++));
    }
    return Val | (ch << (7 * i));
}

func Vec_StrPutW(vOut *Vec_Str_t, Val word) {
    var i int32;
    for  i = 0; i < 8; i++  {
        Vec_StrPush( vOut, byte((Val >> (8*i))) );
    }
}
func Vec_StrGetW(vOut *Vec_Str_t, pPos *int32) word {
    var i int32;
    var Val word;
    for  i = 0; i < 8; i++  {
        Val |= ((word)(unsigned char)Vec_StrEntry(vOut, (*pPos)++) << (8*i));
    }
    return Val;
}

func Vec_StrPutF(vOut *Vec_Str_t, Val float32) {
    union { var num float32; unsigned char data[4]; }
 tmp;
    tmp.num = Val;
    Vec_StrPush( vOut, tmp.data[0] );
    Vec_StrPush( vOut, tmp.data[1] );
    Vec_StrPush( vOut, tmp.data[2] );
    Vec_StrPush( vOut, tmp.data[3] );
}
func Vec_StrGetF(vOut *Vec_Str_t, pPos *int32) float32 {
    union { var num float32; unsigned char data[4]; }
 tmp;
    tmp.data[0] = Vec_StrEntry( vOut, (*pPos)++ );
    tmp.data[1] = Vec_StrEntry( vOut, (*pPos)++ );
    tmp.data[2] = Vec_StrEntry( vOut, (*pPos)++ );
    tmp.data[3] = Vec_StrEntry( vOut, (*pPos)++ );
    return tmp.num;
}

func Vec_StrPutD(vOut *Vec_Str_t, Val float64) {
    union { var num float64; unsigned char data[8]; }
 tmp;
    var i int32;
    var Lim int32 = sizeof(double);
    tmp.num = Val;
    for  i = 0; i < Lim; i++  {
        Vec_StrPush( vOut, tmp.data[i] );
    }
}
func Vec_StrGetD(vOut *Vec_Str_t, pPos *int32) float64 {
    union { var num float64; unsigned char data[8]; }
 tmp;
    var i int32;
    var Lim int32 = sizeof(double);
    for  i = 0; i < Lim; i++  {
        tmp.data[i] = Vec_StrEntry( vOut, (*pPos)++ );
    }
    return tmp.num;
}

func Vec_StrPutS(vOut *Vec_Str_t, pStr string) {
    for  *pStr  {
        Vec_StrPush( vOut, *pStr++ );
    }
    Vec_StrPush( vOut, byte(0) );
}
func Vec_StrGetS(vOut *Vec_Str_t, pPos *int32) string {
    var pStr string  = Vec_StrEntryP( vOut, *pPos );
    for  Vec_StrEntry(vOut, (*pPos)++)  {; }
    return Abc_UtilStrsav(pStr);
}

func Vec_StrPutC(vOut *Vec_Str_t, c byte) {
    Vec_StrPush( vOut, c );
}
func Vec_StrGetC(vOut *Vec_Str_t, pPos *int32) byte {
    return Vec_StrEntry(vOut, (*pPos)++);
}



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

