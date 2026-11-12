package acec

import (
	"fmt"
	"strconv"
)

/**CFile****************************************************************

  FileName    [acecPo.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [CEC for arithmetic circuits.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: acecPo.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acecInt.h"
//#include "misc/vec/vecWec.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Parses signature given on the command line.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acec_ParseSignatureMono(p string, pStop string, vLevel *Vec_Int_t) {
    var pTemp string  = p;
    var Const int32  = ABC_INFINITY;
    var fMinus int32  = pTemp[0] == '-';
    if pTemp[0] == '+' || pTemp[0] == '-' || pTemp[0] == '(' { 
        pTemp++; }
    for  pTemp < pStop  {
        if pTemp[0] == 'i' { // input
            Vec_IntPush( vLevel, -1-strconv.Atoi(++pTemp) );
        } else if pTemp[0] == 'o' { // output
            Vec_IntPush( vLevel, strconv.Atoi(++pTemp) );
        } else { // coefficient
        
            assert(Const == ABC_INFINITY, "Const == ABC_INFINITY");
            Const = 1 + strconv.Atoi(pTemp);
        }
        for  pTemp[0] >= '0' && pTemp[0] <= '9'  {
            pTemp++;
        }
        assert(pTemp ==  pStop || pTemp[0] == '*', "pTemp ==  pStop || pTemp[0] == '*'");
        pTemp++;
    }
    assert(Const != ABC_INFINITY, "Const != ABC_INFINITY");
    Vec_IntPush( vLevel, fMinus ? -Const : Const );
}
func Acec_ParseSignatureOne(p string, pStop string) *Vec_Wec_t {
    var vMonos *Vec_Wec_t  = Vec_WecAlloc( 10 );
    var pTemp string = p;
    var pNext string;
    assert(p[0] == '(' && pStop[0] == ')', "p[0] == '(' && pStop[0] == ')'");
    for  pTemp[0] != ')'  {
        for  pNext = pTemp+1; pNext < pStop; pNext++  {
            if pNext[0] == '+' || pNext[0] == '-' {
                break; }
        }
        assert(pNext[0] == '+' || pNext[0] == '-' || pNext[0] == ')', "pNext[0] == '+' || pNext[0] == '-' || pNext[0] == ')'");
        Acec_ParseSignatureMono( pTemp, pNext, Vec_WecPushLevel(vMonos) );
        pTemp = pNext;
    }
    return vMonos;
}
func Acec_ParseDistribute(vM1 *Vec_Wec_t, vM2 *Vec_Wec_t, vAdd *Vec_Wec_t) *Vec_Wec_t {
    var vMonos *Vec_Wec_t  = Vec_WecAlloc( 10 );
    var vLevel1, vLevel2, vLevel *Vec_Int_t;
    var i, k, n, Entry int32;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vM1)) && (((vLevel1) = Vec_WecEntry(vM1, i)), 1); i++  {
    // inlining of macro Vec_WecForEachLevel
    for  k = 0; (k < Vec_WecSize(vM2)) && (((vLevel2) = Vec_WecEntry(vM2, k)), 1); k++  {
        vLevel = Vec_WecPushLevel(vMonos);
        // inlining of macro Vec_IntForEachEntryStop
        for  n = 0; (n < Vec_IntSize(vLevel1)-1) && (((Entry) = Vec_IntEntry(vLevel1, n)), 1); n++  {
            Vec_IntPush(vLevel, Entry);
        }
        // inlining of macro Vec_IntForEachEntryStop
        for  n = 0; (n < Vec_IntSize(vLevel2)-1) && (((Entry) = Vec_IntEntry(vLevel2, n)), 1); n++  {
            Vec_IntPush(vLevel, Entry);
        }
        Vec_IntPush(vLevel, Vec_IntEntryLast(vLevel1)+Vec_IntEntryLast(vLevel2)-1);
    }
    }
    // inlining of macro Vec_WecForEachLevel
    for  k = 0; (k < Vec_WecSize(vAdd)) && (((vLevel1) = Vec_WecEntry(vAdd, k)), 1); k++  {
        vLevel = Vec_WecPushLevel(vMonos);
        // inlining of macro Vec_IntForEachEntry
        for  n = 0; (n < Vec_IntSize(vLevel1)) && (((Entry) = Vec_IntEntry(vLevel1, n)), 1); n++  {
            Vec_IntPush(vLevel, Entry);
        }
    }
    return vMonos;
}
func Acec_ParseSignature(p string) *Vec_Wec_t {
    var vAdd *Vec_Wec_t = nil;
    var vTemp1 *Vec_Wec_t;
    var vTemp2 *Vec_Wec_t;
    var vRes *Vec_Wec_t;
    if p[0] == '(' {
        var pStop string  = strstr(p, ")");
        if pStop == nil {
            return nil; }
        vTemp1 = Acec_ParseSignatureOne( p, pStop );
        if pStop[1] == 0 {
            return vTemp1; }
        if pStop[1] == '*' {
            var p2 string  = pStop + 2;
            var pStop2 string  = strstr(p2, ")");
            if p2[0] != '(' {
                return nil; }
            if pStop2 == nil {
                return nil; }
            vTemp2 = Acec_ParseSignatureOne( p2, pStop2 );
            if pStop2[1] == 0 {
                vRes = Acec_ParseDistribute( vTemp1, vTemp2, vAdd );
                Vec_WecFree( vTemp1 );
                Vec_WecFree( vTemp2 );
                return vRes;
            }
            if pStop2[1] == '+' {
                var p3 string  = pStop2 + 2;
                var pStop3 string  = strstr(p3, ")");
                if p3[0] != '(' {
                    return nil; }
                if pStop3 == nil {
                    return nil; }
                vAdd = Acec_ParseSignatureOne( p3, pStop3 );
                vRes = Acec_ParseDistribute( vTemp1, vTemp2, vAdd );
                Vec_WecFree( vTemp1 );
                Vec_WecFree( vTemp2 );
                Vec_WecFree( vAdd );
                return vRes;
            }
            assert(false, "0");
        }
        assert(false, "0");
    
    } else {
        var Len int32  = len(p);
        // inlining of macro ABC_ALLOC
        var pCopy string  = ((char *) malloc(sizeof(char) * (size_t)(Len+3)));
        pCopy[0] = '(';
        strcpy( pCopy+1, p );
        pCopy[Len+1] = ')';
        pCopy[Len+2] = '\0';
        vRes = Acec_ParseSignatureOne( pCopy, pCopy + Len + 1 );
        pCopy = nil;
        return vRes;
    }
    return nil;
}
func Acec_PrintSignature(vMonos *Vec_Wec_t) {
    var vLevel *Vec_Int_t; var i, k, Entry int32;
    fmt.Printf("Output signature with %d monomials:\n", Vec_WecSize(vMonos));
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vMonos)) && (((vLevel) = Vec_WecEntry(vMonos, i)), 1); i++  {
        fmt.Printf("  %s2^%d", Vec_IntEntryLast(vLevel) > 0 ? "+":"-", Abc_AbsInt(Vec_IntEntryLast(vLevel))-1);
        // inlining of macro Vec_IntForEachEntryStop
        for  k = 0; (k < Vec_IntSize(vLevel)-1) && (((Entry) = Vec_IntEntry(vLevel, k)), 1); k++  {
            fmt.Printf(" * ");
            if Entry < 0 {
                fmt.Printf("i%d", -Entry-1);
            } else { 
                fmt.Printf("o%d", Entry); }
        }
        fmt.Printf("\n");
    }
}
func Acec_ParseSignatureTest() {
    var pSign string  = "(4*o1+2*o2+1*o3)*(4*i4+2*i5+1*i6)+(4*o4+2*o5+1*o6)";
    var vMonos *Vec_Wec_t  = Acec_ParseSignature( pSign );
    Acec_PrintSignature( vMonos );
    Vec_WecFree( vMonos );
}

/**Function*************************************************************

  Synopsis    [Checks that items are unique and in order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntPushOrderAbs(p *Vec_Int_t, Entry int32) {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        assert(Entry != p.pArray[i], "Entry != p.pArray[i]");
    }
    if p.nSize == p.nCap {
        if p.nCap < 16 {
            Vec_IntGrow( p, 16 );
        } else {
            Vec_IntGrow( p, 2 * p.nCap ); }
    }
    p.nSize++;
    for  i = p.nSize-2; i >= 0; i--  {
        if Abc_AbsInt(p.pArray[i]) < Abc_AbsInt(Entry) {
            p.pArray[i+1] = p.pArray[i];
        } else {
            break; }
    }
    p.pArray[i+1] = Entry;
}
func Vec_IntAppendMinusAbs(vVec1 *Vec_Int_t, vVec2 *Vec_Int_t, fMinus int32) {
    var Entry, i int32;
    Vec_IntClear( vVec1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vVec2)) && (((Entry) = Vec_IntEntry(vVec2, i)), 1); i++  {
        Vec_IntPushOrderAbs( vVec1, fMinus ? -Entry : Entry );
    }
}
func Vec_IntCheckUniqueOrderAbs(p *Vec_Int_t) {
    var i int32;
    for  i = 1; i < p.nSize; i++  {
        assert(Abc_AbsInt(p.pArray[i-1]) > Abc_AbsInt(p.pArray[i]), "Abc_AbsInt(p.pArray[i-1]) > Abc_AbsInt(p.pArray[i])");
    }
}
func Vec_IntCheckUniqueOrder(p *Vec_Int_t) {
    var i int32;
    for  i = 1; i < p.nSize; i++  {
        assert(p.pArray[i-1] < p.pArray[i], "p.pArray[i-1] < p.pArray[i]");
    }
}

/**Function*************************************************************

  Synopsis    [Prints polynomial.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynPrintMono(vConst *Vec_Int_t, vMono *Vec_Int_t, Prev int32) {
    var k, Entry int32;
    fmt.Printf("%c ", Prev != Abc_AbsInt(Vec_IntEntry(vConst, 0)) ? '|' : ' ');
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vConst)) && (((Entry) = Vec_IntEntry(vConst, k)), 1); k++  {
        fmt.Printf("%s2^%d", Entry < 0 ? "-" : "+", Abc_AbsInt(Entry)-1);
    }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vMono)) && (((Entry) = Vec_IntEntry(vMono, k)), 1); k++  {
        fmt.Printf(" * i%d", Entry-1);
    }
    fmt.Printf("\n");
}
func Gia_PolynPrint(vPolyn *Vec_Wec_t) {
    var vConst, vMono *Vec_Int_t;  
    var i int32;
    var Prev int32 = -1;
    fmt.Printf("Polynomial with %d monomials:\n", Vec_WecSize(vPolyn)/2);
    for  i = 0; i < Vec_WecSize(vPolyn)/2; i++  {
        vConst = Vec_WecEntry( vPolyn, 2*i+0 );
        vMono  = Vec_WecEntry( vPolyn, 2*i+1 );
        Gia_PolynPrintMono( vConst, vMono, Prev );
        Prev = Abc_AbsInt( Vec_IntEntry(vConst, 0) );
    }
}
func Gia_PolynPrintStats(vPolyn *Vec_Wec_t) {
    var vConst, vCountsP, vCountsN *Vec_Int_t;
    var i, Entry, Max int32;
    fmt.Printf("Input signature with %d monomials:\n", Vec_WecSize(vPolyn)/2);
    for  i = 0; i < Vec_WecSize(vPolyn)/2; i++  {
        vConst = Vec_WecEntry( vPolyn, 2*i+0 );
        Max = Abc_MaxInt( Max, Abc_AbsInt(Abc_AbsInt(Vec_IntEntry(vConst, 0))) );
    }
    vCountsP = Vec_IntStart( Max + 1 );
    vCountsN = Vec_IntStart( Max + 1 );
    for  i = 0; i < Vec_WecSize(vPolyn)/2; i++  {
        vConst = Vec_WecEntry( vPolyn, 2*i+0 );
        Entry = Vec_IntEntry(vConst, 0);
        if Entry > 0 {
            Vec_IntAddToEntry( vCountsP, Entry, 1 );
        } else {
            Vec_IntAddToEntry( vCountsN, -Entry, 1 ); }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCountsN)) && (((Entry) = Vec_IntEntry(vCountsN, i)), 1); i++  {
        if Entry != 0 {
            fmt.Printf("  -2^%d appears %d times\n", Abc_AbsInt(i)-1, Entry); }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCountsP)) && (((Entry) = Vec_IntEntry(vCountsP, i)), 1); i++  {
        if Entry != 0 {
            fmt.Printf("  +2^%d appears %d times\n", Abc_AbsInt(i)-1, Entry); }
    }
    Vec_IntFree( vCountsP );
    Vec_IntFree( vCountsN );
}

/**Function*************************************************************

  Synopsis    [Collects polynomial.]

  Description [Collects non-trivial monomials in the increasing order 
  of the absolute value of the their first coefficients.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynGetResultCompare(p0 *int32, p1 *int32) int32 {
    if p0[2] < p1[2] { return -1; }
    if p0[2] > p1[2] { return  1; }
    return 0;
}
func Gia_PolynGetResult(pHashC *Hsh_VecMan_t, pHashM *Hsh_VecMan_t, vCoefs *Vec_Int_t) *Vec_Wec_t {
    var vClass, vLevel, vArray *Vec_Int_t;
    var vPolyn, vSorted *Vec_Wec_t;
    var i, k, iConst, iMono, iFirst int32;
    // find the largest 
    var nLargest, nNonConst int32;
    // inlining of macro Vec_IntForEachEntry
    for  iMono = 0; (iMono < Vec_IntSize(vCoefs)) && (((iConst) = Vec_IntEntry(vCoefs, iMono)), 1); iMono++  {
        //Vec_IntPrint( Hsh_VecReadEntry(pHashM, iMono) );
        if iConst == 0 { 
            continue; }
        vArray = Hsh_VecReadEntry( pHashC, iConst );
        nLargest = Abc_MaxInt( nLargest, Abc_AbsInt(Vec_IntEntry(vArray, 0)) );
        nNonConst++;
    }
    // sort by the size of the largest coefficient
    vSorted = Vec_WecStart( nLargest+1 );
    // inlining of macro Vec_IntForEachEntry
    for  iMono = 0; (iMono < Vec_IntSize(vCoefs)) && (((iConst) = Vec_IntEntry(vCoefs, iMono)), 1); iMono++  {
        if iConst == 0 { 
            continue; }
        vArray = Hsh_VecReadEntry( pHashC, iConst );
        vLevel = Vec_WecEntry( vSorted, Abc_AbsInt(Vec_IntEntry(vArray, 0)) );
        vArray = Hsh_VecReadEntry( pHashM, iMono );
        if Vec_IntSize(vArray) {
            iFirst = Vec_IntEntry(vArray, 0);
        } else {
            iFirst = -1;
        }
        Vec_IntPushThree( vLevel, iConst, iMono, iFirst );
    }
    // reload in the given order
    vPolyn = Vec_WecAlloc( 2*nNonConst );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vSorted)) && (((vClass) = Vec_WecEntry(vSorted, i)), 1); i++  {
        // sort monomials by the index of the first variable
        qsort( Vec_IntArray(vClass), (size_t)(Vec_IntSize(vClass)/3), 12, (int (*)(void *, void *))Gia_PolynGetResultCompare );
        // inlining of macro Vec_IntForEachEntryTriple
        for  k = 0; (k+2 < Vec_IntSize(vClass)) && (((iConst) = Vec_IntEntry(vClass, k)), 1) && (((iMono) = Vec_IntEntry(vClass, k+1)), 1) && (((iFirst) = Vec_IntEntry(vClass, k+2)), 1); k += 3  {
            vArray = Hsh_VecReadEntry( pHashC, iConst );
            Vec_IntCheckUniqueOrderAbs( vArray );
            vLevel = Vec_WecPushLevel( vPolyn );
            Vec_IntGrow( vLevel, Vec_IntSize(vArray) );
            Vec_IntAppend( vLevel, vArray );

            vArray = Hsh_VecReadEntry( pHashM, iMono );
            Vec_IntCheckUniqueOrder( vArray );
            vLevel = Vec_WecPushLevel( vPolyn );
            Vec_IntGrow( vLevel, Vec_IntSize(vArray) );
            Vec_IntAppend( vLevel, vArray );
        }
    }
    assert(Vec_WecSize(vPolyn) == 2*nNonConst, "Vec_WecSize(vPolyn) == 2*nNonConst");
    Vec_WecFree( vSorted );
    return vPolyn;
}


/**Function*************************************************************

  Synopsis    [Derives new constant.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynMergeConstOne(vConst *Vec_Int_t, New int32) {
    var i, Old int32;
    assert(New != 0, "New != 0");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vConst)) && (((Old) = Vec_IntEntry(vConst, i)), 1); i++  {
        assert(Old != 0, "Old != 0");
        if Old == New { // A == B
        
            Vec_IntDrop( vConst, i );
            Gia_PolynMergeConstOne( vConst, New > 0 ? New + 1 : New - 1 );
            return;
        }
        if Abc_AbsInt(Old) == Abc_AbsInt(New) { // A == -B
        
            Vec_IntDrop( vConst, i );
            return;
        }
        if Old + New == 1 || Old + New == -1 {  // sign(A) != sign(B)  &&  abs(abs(A)-abs(B)) == 1 
        
            var Value int32  = Abc_MinInt( Abc_AbsInt(Old), Abc_AbsInt(New) );
            Vec_IntDrop( vConst, i );
            Gia_PolynMergeConstOne( vConst, (Old + New == 1) ? Value : -Value );
            return;
        }
    }
    Vec_IntPushOrderAbs( vConst, New );
}
func Gia_PolynMergeConst(vTempC *Vec_Int_t, pHashC *Hsh_VecMan_t, iConstAdd int32) {
    var i, New int32;
    var vConstAdd *Vec_Int_t  = Hsh_VecReadEntry( pHashC, iConstAdd );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vConstAdd)) && (((New) = Vec_IntEntry(vConstAdd, i)), 1); i++  {
        Gia_PolynMergeConstOne( vTempC, New );
        vConstAdd = Hsh_VecReadEntry( pHashC, iConstAdd );
    }
    Vec_IntCheckUniqueOrderAbs( vConstAdd );
    //Vec_IntPrint( vConstAdd );
}
func Gia_PolynBuildAdd(pHashC *Hsh_VecMan_t, pHashM *Hsh_VecMan_t, vCoefs *Vec_Int_t, vLit2Mono *Vec_Wec_t, vTempC *Vec_Int_t, vTempM *Vec_Int_t) int32 {
    var i, iLit, iConst, iConstNew int32;
    var iMono int32  = Hsh_VecManAdd(pHashM, vTempM);
    if iMono == Vec_IntSize(vCoefs) { // new monomial
    
        // map monomial into a constant
        assert(Vec_IntSize(vTempC) > 0, "Vec_IntSize(vTempC) > 0");
        iConst = Hsh_VecManAdd( pHashC, vTempC );
        Vec_IntPush( vCoefs, iConst );
        // map literals into monomial
        assert(Vec_IntSize(vTempM) > 0, "Vec_IntSize(vTempM) > 0");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vTempM)) && (((iLit) = Vec_IntEntry(vTempM, i)), 1); i++  {
            Vec_WecPush( vLit2Mono, iLit, iMono );
        } 
        //printf( "New monomial: \n" );
        //Gia_PolynPrintMono( vTempC, vTempM );
        return 1;
    }
    // this monomial exists
    iConst = Vec_IntEntry( vCoefs, iMono );
    if iConst != 0 {
        Gia_PolynMergeConst( vTempC, pHashC, iConst ); }
    iConstNew = Hsh_VecManAdd( pHashC, vTempC );
    Vec_IntWriteEntry( vCoefs, iMono, iConstNew );
    //printf( "Old monomial: \n" );
    //Gia_PolynPrintMono( vTempC, vTempM );
    if (iConst != 0) && !(iConstNew != 0) {
        return -1; }
    if !(iConst != 0) && (iConstNew != 0) {
        return 1; }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Computing for literals.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynHandleOne(pHashC *Hsh_VecMan_t, pHashM *Hsh_VecMan_t, vCoefs *Vec_Int_t, vLit2Mono *Vec_Wec_t, vTempC *Vec_Int_t, vTempM *Vec_Int_t, iMono int32, iLitOld int32, iLitNew0 int32, iLitNew1 int32) int32 {
    var status int32;
    var iConst int32 = Vec_IntEntry( vCoefs, iMono );
    var vArrayC *Vec_Int_t  = Hsh_VecReadEntry( pHashC, iConst );
    var vArrayM *Vec_Int_t  = Hsh_VecReadEntry( pHashM, iMono );
    // create new monomial
    Vec_IntClear( vTempM );
    Vec_IntAppend( vTempM, vArrayM );
    status = Vec_IntRemove( vTempM, iLitOld );
    assert((status != 0), "status");
    // create new monomial
    if iLitNew0 == -1 && iLitNew1 == -1 {     // no new lit - the same const
        Vec_IntAppendMinusAbs( vTempC, vArrayC, 0 );
    } else if iLitNew0 > -1 && iLitNew1 == -1 { // one new lit - opposite const
    
        Vec_IntAppendMinusAbs( vTempC, vArrayC, 1 );
        Vec_IntPushUniqueOrder( vTempM, iLitNew0 );
    
    } else if iLitNew0 > -1 && iLitNew1 > -1 {  // both new lit - the same const
    
        Vec_IntAppendMinusAbs( vTempC, vArrayC, 0 );
        Vec_IntPushUniqueOrder( vTempM, iLitNew0 );
        Vec_IntPushUniqueOrder( vTempM, iLitNew1 );
    
    } else { assert(false, "0"); }
    return Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC, vTempM );
}

func Gia_PolynBuildNew2(pGia *Gia_Man_t, vRootLits *Vec_Int_t, nExtra int32, vLeaves *Vec_Int_t, vNodes *Vec_Int_t, fSigned int32, fVerbose int32, fVeryVerbose int32) *Vec_Wec_t {
    var clk abctime  = Abc_Clock();
    var vPolyn *Vec_Wec_t;
    var vLit2Mono *Vec_Wec_t  = Vec_WecStart( 2 * Gia_ManObjNum(pGia) ); // mapping AIG literals into monomials
    var pHashC *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );    // hash table for constants
    var pHashM *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );    // hash table for monomials
    var vCoefs *Vec_Int_t     = Vec_IntAlloc( 1000 );       // monomial coefficients
    var vTempC *Vec_Int_t     = Vec_IntAlloc( 10 );         // temporary array
    var vTempM *Vec_Int_t     = Vec_IntAlloc( 10 );         // temporary array
    var i, k, iObj, iLit, iMono, nMonos, nBuilds int32;

    // add 0-constant and 1-monomial
    Hsh_VecManAdd( pHashC, vTempC );
    Hsh_VecManAdd( pHashM, vTempM );
    Vec_IntPush( vCoefs, 0 );

    // create output signature
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRootLits)) && (((iLit) = Vec_IntEntry(vRootLits, i)), 1); i++  {
        var Value int32  = 1 + Abc_MinInt( i, Vec_IntSize(vRootLits)-nExtra );
        Vec_IntFill( vTempC, 1, (fSigned && i == Vec_IntSize(vRootLits)-1-nExtra) ? -Value : Value );
        Vec_IntFill( vTempM, 1, iLit );
        nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC, vTempM );
        nBuilds++;
    }

    // perform construction for internal nodes
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vNodes) - 1; (i >= 0) && (((iObj) = Vec_IntEntry(vNodes, i)), 1); i--  {
        var pObj *Gia_Obj_t  = Gia_ManObj( pGia, iObj );
        var iLits [2]int32 = [2]int32{ Abc_Var2Lit(iObj, 0),         Abc_Var2Lit(iObj, 1)         };
        var iFans [2]int32 = [2]int32{ Gia_ObjFaninLit0(pObj, iObj), Gia_ObjFaninLit1(pObj, iObj) };
        // add inverter
        var vArray *Vec_Int_t  = Vec_WecEntry( vLit2Mono, iLits[1] );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((iMono) = Vec_IntEntry(vArray, k)), 1); k++  {
            if Vec_IntEntry(vCoefs, iMono) > 0 {
                nMonos += Gia_PolynHandleOne( pHashC, pHashM, vCoefs, vLit2Mono, vTempC, vTempM, iMono, iLits[1], -1, -1 );
                nMonos += Gia_PolynHandleOne( pHashC, pHashM, vCoefs, vLit2Mono, vTempC, vTempM, iMono, iLits[1], iLits[0], -1 );
                Vec_IntWriteEntry( vCoefs, iMono, 0 );
                nMonos--;
                nBuilds++;
                nBuilds++;
            }
        }
        // add AND gate
        vArray = Vec_WecEntry( vLit2Mono, iLits[0] );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((iMono) = Vec_IntEntry(vArray, k)), 1); k++  {
            if Vec_IntEntry(vCoefs, iMono) > 0 {
                nMonos += Gia_PolynHandleOne( pHashC, pHashM, vCoefs, vLit2Mono, vTempC, vTempM, iMono, iLits[0], iFans[0], iFans[1] );
                Vec_IntWriteEntry( vCoefs, iMono, 0 );
                nMonos--;
                nBuilds++;
            }
        }
        //printf( "Obj %5d : nMonos = %6d  nUsed = %6d\n", iObj, nBuilds, nMonos );
    }

    // complement leave nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iObj) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        var iLits [2]int32 = [2]int32{ Abc_Var2Lit(iObj, 0),  Abc_Var2Lit(iObj, 1) };
        // add inverter
        var vArray *Vec_Int_t  = Vec_WecEntry( vLit2Mono, iLits[1] );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((iMono) = Vec_IntEntry(vArray, k)), 1); k++  {
            if Vec_IntEntry(vCoefs, iMono) > 0 {
                nMonos += Gia_PolynHandleOne( pHashC, pHashM, vCoefs, vLit2Mono, vTempC, vTempM, iMono, iLits[1], -1, -1 );
                nMonos += Gia_PolynHandleOne( pHashC, pHashM, vCoefs, vLit2Mono, vTempC, vTempM, iMono, iLits[1], iLits[0], -1 );
                Vec_IntWriteEntry( vCoefs, iMono, 0 );
                nMonos--;
                nBuilds++;
            }
        }
    }

    // get the results
    vPolyn = Gia_PolynGetResult( pHashC, pHashM, vCoefs );

    fmt.Printf("HashC = %d. HashM = %d.  Total = %d. Left = %d.  Used = %d.  ", Hsh_VecSize(pHashC), Hsh_VecSize(pHashM), nBuilds, nMonos, Vec_WecSize(vPolyn)/2);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    Vec_IntFree( vTempC );
    Vec_IntFree( vTempM );
    Vec_IntFree( vCoefs );
    Vec_WecFree( vLit2Mono );
    Hsh_VecManStop( pHashC );
    Hsh_VecManStop( pHashM );
    return vPolyn;
}


/**Function*************************************************************

  Synopsis    [Computing for objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynPrepare2(vTempC []*Vec_Int_t, vTempM []*Vec_Int_t, iObj int32, iCst int32) {
    Vec_IntFill( vTempC[0], 1, iCst );
    Vec_IntFill( vTempC[1], 1, -iCst );
    Vec_IntClear( vTempM[0] );
    Vec_IntFill( vTempM[1], 1, iObj );
}
func Gia_PolynPrepare4(vTempC []*Vec_Int_t, vTempM []*Vec_Int_t, vConst *Vec_Int_t, vMono *Vec_Int_t, iObj int32, iFan0 int32, iFan1 int32) {
    var i, k, Entry int32;
    for  i = 0; i < 4; i++  {
        Vec_IntAppendMinusAbs( vTempC[i], vConst, i & 1 );
    }
    for  i = 0; i < 4; i++  {
        Vec_IntClear( vTempM[i] );
    }
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vMono)) && (((Entry) = Vec_IntEntry(vMono, k)), 1); k++  {
        if Entry != iObj {
            for  i = 0; i < 4; i++  {
                Vec_IntPush( vTempM[i], Entry );
            } }
    }
    Vec_IntPushUniqueOrder( vTempM[1], iFan0 );
    Vec_IntPushUniqueOrder( vTempM[2], iFan1 );
    Vec_IntPushUniqueOrder( vTempM[3], iFan0 );
    Vec_IntPushUniqueOrder( vTempM[3], iFan1 );
}

func Gia_PolynBuildNew(pGia *Gia_Man_t, vSign *Vec_Wec_t, vRootLits *Vec_Int_t, nExtra int32, vLeaves *Vec_Int_t, vNodes *Vec_Int_t, fSigned int32, fVerbose int32, fVeryVerbose int32) *Vec_Wec_t {
    var clk abctime  = Abc_Clock();
    var vPolyn *Vec_Wec_t;
    var vLit2Mono *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(pGia) ); // mapping AIG literals into monomials
    var pHashC *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );    // hash table for constants
    var pHashM *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );    // hash table for monomials
    var vCoefs *Vec_Int_t     = Vec_IntAlloc( 1000 );       // monomial coefficients
    var vTempC, vTempM [4]*Vec_Int_t;                // temporary array
    var i, k, iObj, iLit, iMono, iConst, nMonos, nBuilds int32;
    for  i = 0; i < 4; i++  {
        vTempC[i] = Vec_IntAlloc( 10 );
    }
    for  i = 0; i < 4; i++  {
        vTempM[i] = Vec_IntAlloc( 10 );
    }

    // add 0-constant and 1-monomial
    Hsh_VecManAdd( pHashC, vTempC[0] );
    Hsh_VecManAdd( pHashM, vTempM[0] );
    Vec_IntPush( vCoefs, 0 );

    if nExtra != 0 {
        fmt.Printf("Assigning %d outputs from %d to %d rank %d.\n", nExtra, Vec_IntSize(vRootLits)-nExtra, Vec_IntSize(vRootLits)-1, Vec_IntSize(vRootLits)-nExtra); }

    // create output signature
    if vSign != nil {
        var vLevel *Vec_Int_t; var Entry, OutLit int32;
        // inlining of macro Vec_WecForEachLevel
        for  i = 0; (i < Vec_WecSize(vSign)) && (((vLevel) = Vec_WecEntry(vSign, i)), 1); i++  {
            OutLit = -1;
            Vec_IntClear( vTempM[0] );
            Vec_IntFill( vTempC[0], 1, Vec_IntEntryLast(vLevel) );
            // inlining of macro Vec_IntForEachEntryStop
            for  k = 0; (k < Vec_IntSize(vLevel)-1) && (((Entry) = Vec_IntEntry(vLevel, k)), 1); k++  {
                if Entry < 0 { // input
                    Vec_IntPushUniqueOrder( vTempM[0], Vec_IntEntry(vLeaves, -1-Entry) );
                } else { // output
                
                    assert(OutLit == -1, "OutLit == -1"); // only one output literal is expected
                    OutLit = Vec_IntEntry(vRootLits, Entry);
                }
            }
            if OutLit == -1 {
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[0] );   // mono without out
            } else if !Abc_LitIsCompl(OutLit) { // positive literal
            
                Vec_IntPushUniqueOrder( vTempM[0], Abc_Lit2Var(OutLit) );
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[0] );   // mono with pos out
            
            } else { // negative literal
            
                // first monomial
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[0] );   // mono without out
                // second monomial
                Vec_IntFill( vTempC[0], 1, -Vec_IntEntryLast(vLevel) );
                Vec_IntPushUniqueOrder( vTempM[0], Abc_Lit2Var(OutLit) );
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[0] );   // mono with neg out
            }
            nBuilds++;
        }
    
    } else {
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRootLits)) && (((iLit) = Vec_IntEntry(vRootLits, i)), 1); i++  {
        var Value int32  = 1 + Abc_MinInt( i, Vec_IntSize(vRootLits)-nExtra );
        Gia_PolynPrepare2( vTempC, vTempM, Abc_Lit2Var(iLit), Value );
        if (fSigned != 0) && i >= Vec_IntSize(vRootLits)-nExtra-1 {
            if fVeryVerbose != 0 { fmt.Printf("Out %d : Negative   Value = %d\n", i, Value-1); }
            if Abc_LitIsCompl(iLit) {
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[1], vTempM[0] );   // -C
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[1] );   //  C * Driver
                nBuilds++;
            
            } else {
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[1], vTempM[1] ); }   // -C * Driver
        
        } else {
            if fVeryVerbose != 0 { fmt.Printf("Out %d : Positive   Value = %d\n", i, Value-1); }
            if Abc_LitIsCompl(iLit) {
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[0] );   //  C
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[1], vTempM[1] );   // -C * Driver
                nBuilds++;
            
            } else {
                nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[1] ); }   //  C * Driver
        }
        nBuilds++;
    } }

    // perform construction for internal nodes
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vNodes) - 1; (i >= 0) && (((iObj) = Vec_IntEntry(vNodes, i)), 1); i--  {
        var pObj *Gia_Obj_t  = Gia_ManObj( pGia, iObj );
        var vArray *Vec_Int_t  = Vec_WecEntry( vLit2Mono, iObj );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((iMono) = Vec_IntEntry(vArray, k)), 1); k++  {
            if (iConst = Vec_IntEntry(vCoefs, iMono)) > 0 {
                var vArrayC *Vec_Int_t  = Hsh_VecReadEntry( pHashC, iConst );
                var vArrayM *Vec_Int_t  = Hsh_VecReadEntry( pHashM, iMono );
                Gia_PolynPrepare4( vTempC, vTempM, vArrayC, vArrayM, iObj, Gia_ObjFaninId0(pObj, iObj), Gia_ObjFaninId1(pObj, iObj) );
                if Gia_ObjIsXor(pObj) {
                
                } else if Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) {  //  C * (1 - x) * (1 - y)
                
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[0] );   //  C * 1
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[1], vTempM[1] );   // -C * x
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[3], vTempM[2] );   // -C * y 
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[2], vTempM[3] );   //  C * x * y
                    nBuilds += 3;
                
                } else if Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) { //  C * (1 - x) * y
                
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[2] );   //  C * y 
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[1], vTempM[3] );   // -C * x * y
                    nBuilds += 2;
                
                } else if !Gia_ObjFaninC0(pObj) && Gia_ObjFaninC1(pObj) { //  C * x * (1 - y)
                
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[1] );   //  C * x 
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[1], vTempM[3] );   // -C * x * y
                    nBuilds++;
                
                } else {   
                    nMonos += Gia_PolynBuildAdd( pHashC, pHashM, vCoefs, vLit2Mono, vTempC[0], vTempM[3] ); } //  C * x * y
                Vec_IntWriteEntry( vCoefs, iMono, 0 );
                nMonos--;
                nBuilds++;
            }
        }
        //printf( "Obj %5d : nMonos = %6d  nUsed = %6d\n", iObj, nBuilds, nMonos );
    }

    // get the results
    vPolyn = Gia_PolynGetResult( pHashC, pHashM, vCoefs );

    fmt.Printf("HashC = %d. HashM = %d.  Total = %d. Left = %d.  Used = %d.  ", Hsh_VecSize(pHashC), Hsh_VecSize(pHashM), nBuilds, nMonos, Vec_WecSize(vPolyn)/2);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    for  i = 0; i < 4; i++  {
        Vec_IntFree( vTempC[i] );
    }
    for  i = 0; i < 4; i++  {
        Vec_IntFree( vTempM[i] );
    }
    Vec_IntFree( vCoefs );
    Vec_WecFree( vLit2Mono );
    Hsh_VecManStop( pHashC );
    Hsh_VecManStop( pHashM );
    return vPolyn;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_PolynBuild2Test(pGia *Gia_Man_t, pSign string, nExtra int32, fSigned int32, fVerbose int32, fVeryVerbose int32) {
    var vPolyn *Vec_Wec_t;
    var vRootLits *Vec_Int_t  = Vec_IntAlloc( Gia_ManCoNum(pGia) );
    var vLeaves *Vec_Int_t    = Vec_IntAlloc( Gia_ManCiNum(pGia) );
    var vNodes *Vec_Int_t     = Vec_IntAlloc( Gia_ManAndNum(pGia) );
    var pObj *Gia_Obj_t;
    var vMonos *Vec_Wec_t  = nil;
    var i int32;

    if pSign != nil && (vMonos = Acec_ParseSignature(pSign)) == nil {
        fmt.Printf("Canont parse the output signatures.\n");
        return;
    }
    if (vMonos != nil) && (fVerbose != 0) {
        Acec_PrintSignature( vMonos ); }

    // print logic level
    if nExtra == -1 {
        var LevelMax int32 = -1;
        var iMax int32 = -1;
        Gia_ManLevelNum( pGia );
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((pObj) = Gia_ManCo(pGia, i)); i++  {
            if LevelMax < Gia_ObjLevel(pGia, pObj) {
                LevelMax = Gia_ObjLevel(pGia, pObj);
                iMax = i;
            }
        }
        nExtra = Gia_ManCoNum(pGia) - iMax - 1;
        fmt.Printf("Determined the number of extra outputs to be %d.\n", nExtra);
    }

    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < pGia.nObjs) && ((pObj) = Gia_ManObj(pGia, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            Vec_IntPush( vLeaves, i );
        } else if Gia_ObjIsAnd(pObj) {
            Vec_IntPush( vNodes, i );
        } else if Gia_ObjIsCo(pObj) {
            Vec_IntPush( vRootLits, Gia_ObjFaninLit0p(pGia, pObj) ); }
    }

    vPolyn = Gia_PolynBuildNew( pGia, vMonos, vRootLits, nExtra, vLeaves, vNodes, fSigned, fVerbose, fVeryVerbose );
    //printf( "Polynomial has %d monomials.\n", Vec_WecSize(vPolyn)/2 );
    if (fVerbose != 0) || (fVeryVerbose != 0) {
        Gia_PolynPrintStats( vPolyn ); }
    if fVeryVerbose != 0 {
        Gia_PolynPrint( vPolyn ); }
    Vec_WecFree( vPolyn );

    Vec_IntFree( vRootLits );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vNodes );
    Vec_WecFreeP( &vMonos );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

