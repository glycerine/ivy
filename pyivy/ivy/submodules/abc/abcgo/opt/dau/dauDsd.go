package dau

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [dauDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Disjoint-support decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauDsd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

/* 
    This code performs truth-table-based decomposition for 6-variable functions.
    Representation of operations:
    ! = not; 
    (ab) = a and b;  
    [ab] = a xor b;  
    <abc> = ITE( a, b, c )
    FUNCTION{abc} = FUNCTION( a, b, c )
*/


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Elementary truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dau_DsdTtElems_TtElems [DAU_MAX_VAR+1][DAU_MAX_WORD]word;
var spStatic_Dau_DsdTtElems_pTtElems [DAU_MAX_VAR+1]*word = [DAU_MAX_VAR+1]*word{nil};
func Dau_DsdTtElems() **word {
    
    if spStatic_Dau_DsdTtElems_pTtElems[0] == nil {
        var v int32;
        for  v = 0; v <= DAU_MAX_VAR; v++  {
            spStatic_Dau_DsdTtElems_pTtElems[v] = spStatic_Dau_DsdTtElems_TtElems[v];
        }
        Abc_TtElemInit( spStatic_Dau_DsdTtElems_pTtElems, DAU_MAX_VAR );
    }
    return spStatic_Dau_DsdTtElems_pTtElems;
}

/**Function*************************************************************

  Synopsis    [DSD formula manipulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dau_DsdComputeMatches_pMatches [DAU_MAX_STR]int32;
func Dau_DsdComputeMatches(p string) *int32 {
    
    var pNested [DAU_MAX_VAR]int32;
    var v, nNested int32;
    for  v = 0; p[v]; v++  {
        spStatic_Dau_DsdComputeMatches_pMatches[v] = 0;
        if p[v] == '(' || p[v] == '[' || p[v] == '<' || p[v] == '{' {
            pNested[nNested++] = v;
        } else if p[v] == ')' || p[v] == ']' || p[v] == '>' || p[v] == '}' {
            spStatic_Dau_DsdComputeMatches_pMatches[pNested[--nNested]] = v; }
        assert(nNested < DAU_MAX_VAR, "nNested < DAU_MAX_VAR");
    }
    assert(nNested == 0, "nNested == 0");
    return spStatic_Dau_DsdComputeMatches_pMatches;
}

/**Function*************************************************************

  Synopsis    [Generate random permutation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdFindVarNum(pDsd string) int32 {
    var vMax int32;
    pDsd--;
    for  *++pDsd  {
        if *pDsd >= 'a' && *pDsd <= 'z' {
            vMax = Abc_MaxInt( vMax, *pDsd - 'a' ); }
    }
    return vMax + 1;
}
func Dau_DsdGenRandPerm(pPerm *int32, nVars int32) {
    var v, vNew int32;
    for  v = 0; v < nVars; v++  {
        pPerm[v] = v;
    }
    for  v = 0; v < nVars; v++  {
        vNew = rand() % nVars;
        // inlining of macro ABC_SWAP
        { var t int32  = pPerm[v]; pPerm[v] = pPerm[vNew]; pPerm[vNew] = t; };
    }
}
func Dau_DsdPermute(pDsd string) {
    var pPerm [16]int32;
    var nVars int32  = Dau_DsdFindVarNum( pDsd );
    Dau_DsdGenRandPerm( pPerm, nVars );
    pDsd--;
    for  *++pDsd  {
        if *pDsd >= 'a' && *pDsd < 'a' + nVars {
            *pDsd = 'a' + pPerm[*pDsd - 'a']; }
    }
}

/**Function*************************************************************

  Synopsis    [Normalize the ordering of components.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdNormalizeCopy(pDest string, pSour string, pMarks *int32, i int32) string {
    var s int32;
    for  s = pMarks[i]; s < pMarks[i+1]; s++  {
        *pDest++ = pSour[s];
    }
    return pDest;
}
func Dau_DsdNormalizeCompare(pStr string, pMarks *int32, i int32, j int32) int32 {
    var pStr1 string  = pStr + pMarks[i];
    var pStr2 string  = pStr + pMarks[j];
    var pLimit1 string  = pStr + pMarks[i+1];
    var pLimit2 string  = pStr + pMarks[j+1];
    for  ; pStr1 < pLimit1 && pStr2 < pLimit2; pStr1, pStr2 = pStr1+1, pStr2+1  {
        if !(*pStr1 >= 'a' &&  *pStr1 <= 'z') {
            pStr2--;
            continue;
        }
        if !(*pStr2 >= 'a' &&  *pStr2 <= 'z') {
            pStr1--;
            continue;
        }
        if *pStr1 < *pStr2 {
            return -1; }
        if *pStr1 > *pStr2 {
            return 1; }
    }
    assert(pStr1 < pLimit1 || pStr2 < pLimit2, "pStr1 < pLimit1 || pStr2 < pLimit2");
    if pStr1 == pLimit1 {
        return -1; }
    if pStr2 == pLimit2 {
        return 1; }
    assert(false, "0");
    return 0;
}
var spStatic_Dau_DsdNormalizePerm_pPerm [DAU_MAX_VAR]int32;
func Dau_DsdNormalizePerm(pStr string, pMarks *int32, nMarks int32) *int32 {
    
    var i, k int32;
    for  i = 0; i < nMarks; i++  {
        spStatic_Dau_DsdNormalizePerm_pPerm[i] = i;
    }
    for  i = 0; i < nMarks; i++  {
        var iBest int32  = i;
        for  k = i + 1; k < nMarks; k++  {
            if Dau_DsdNormalizeCompare( pStr, pMarks, spStatic_Dau_DsdNormalizePerm_pPerm[iBest], spStatic_Dau_DsdNormalizePerm_pPerm[k] ) == 1 {
                iBest = k; }
        }
        // inlining of macro ABC_SWAP
        { var t int32  = spStatic_Dau_DsdNormalizePerm_pPerm[i]; spStatic_Dau_DsdNormalizePerm_pPerm[i] = spStatic_Dau_DsdNormalizePerm_pPerm[iBest]; spStatic_Dau_DsdNormalizePerm_pPerm[iBest] = t; };
    }
    return spStatic_Dau_DsdNormalizePerm_pPerm;
}
var spStatic_Dau_DsdNormalize_rec_pBuffer [DAU_MAX_STR]byte;
func Dau_DsdNormalize_rec(pStr string, p *string, pMatches *int32) {
    
    if **p == '!' {
        (*p)++; }
    for  (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9')  {
        (*p)++;
    }
    if **p == '<' {
        var q string  = pStr + pMatches[*p - pStr];
        if *(q+1) == '{' {
            *p = q+1; }
    }
    if **p >= 'a' && **p <= 'z' { // var
        return; }
    if **p == '(' || **p == '[' { // and/or/xor
    
        var pStore string;
        var pOld string = *p + 1;
        var q string  = pStr + pMatches[ *p - pStr ];
        var i int32;
        var pPerm *int32;
        var nMarks int32;
        var pMarks [DAU_MAX_VAR+1]int32;
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            pMarks[nMarks++] = *p - pStr;
            Dau_DsdNormalize_rec( pStr, p, pMatches );
        }
        pMarks[nMarks] = *p - pStr;
        assert(*p == q, "*p == q");
        // add to buffer in good order
        pPerm = Dau_DsdNormalizePerm( pStr, pMarks, nMarks );
        // copy to the buffer
        pStore = spStatic_Dau_DsdNormalize_rec_pBuffer;
        for  i = 0; i < nMarks; i++  {
            pStore = Dau_DsdNormalizeCopy( pStore, pStr, pMarks, pPerm[i] );
        }
        assert(pStore - spStatic_Dau_DsdNormalize_rec_pBuffer == *p - pOld, "pStore - spStatic_Dau_DsdNormalize_rec_pBuffer == *p - pOld");
        memcpy( pOld, spStatic_Dau_DsdNormalize_rec_pBuffer, (size_t)(pStore - spStatic_Dau_DsdNormalize_rec_pBuffer) );
        return;
    }
    if **p == '<' || **p == '{' { // mux
    
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        if (**p == '<') && (*(q+1) == '{') {
            *p = q+1;
            Dau_DsdNormalize_rec( pStr, p, pMatches );
            return;
        }
        for  (*p)++; *p < q; (*p)++  {
            Dau_DsdNormalize_rec( pStr, p, pMatches );
        }
        assert(*p == q, "*p == q");
        return;
    }
    assert(false, "0");
}
func Dau_DsdNormalize(pDsd string) {
    if pDsd[1] != 0 {
        Dau_DsdNormalize_rec( pDsd, &pDsd, Dau_DsdComputeMatches(pDsd) ); }
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdCountAnds_rec(pStr string, p *string, pMatches *int32) int32 {
    if **p == '!' {
        (*p)++; }
    for  (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9')  {
        (*p)++;
    }
    if **p == '<' {
        var q string  = pStr + pMatches[*p - pStr];
        if *(q+1) == '{' {
            *p = q+1; }
    }
    if **p >= 'a' && **p <= 'z' { // var
        return 0; }
    if **p == '(' || **p == '[' { // and/or/xor
    
        var Counter int32;
        var AddOn int32;
        if **p == '(' {
            AddOn = 1;
        } else {
            AddOn = 3;
        }
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            Counter += AddOn + Dau_DsdCountAnds_rec( pStr, p, pMatches );
        }
        assert(*p == q, "*p == q");
        return Counter - AddOn;
    }
    if **p == '<' || **p == '{' { // mux
    
        var Counter int32  = 3;
        var q string  = pStr + pMatches[ *p - pStr ];
        assert(*q == **p + 1 + (**p != '('), "*q == **p + 1 + (**p != '(')");
        for  (*p)++; *p < q; (*p)++  {
            Counter += Dau_DsdCountAnds_rec( pStr, p, pMatches );
        }
        assert(*p == q, "*p == q");
        return Counter;
    }
    assert(false, "0");
    return 0;
}
func Dau_DsdCountAnds(pDsd string) int32 {
    if pDsd[1] == 0 {
        return 0; }
    return Dau_DsdCountAnds_rec( pDsd, &pDsd, Dau_DsdComputeMatches(pDsd) );
}

/**Function*************************************************************

  Synopsis    [Computes truth table for the DSD formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_Dsd6TruthCompose_rec(Func word, pFanins *word, nVars int32) word {
    var t0, t1 word;
    if Func == 0 {
        return 0; }
    if Func == ~(word)0 {
        return ~(word)0; }
    assert(nVars > 0, "nVars > 0");
    if nVars--; nVars == 0 {
        assert(Func == s_Truths6[0] || Func == s_Truths6Neg[0], "Func == s_Truths6[0] || Func == s_Truths6Neg[0]");
        if Func == s_Truths6[0] {
            return pFanins[0];
        } else {
            return ~pFanins[0];
        }
    }
    if !Abc_Tt6HasVar(Func, nVars) {
        return Dau_Dsd6TruthCompose_rec( Func, pFanins, nVars ); }
    t0 = Dau_Dsd6TruthCompose_rec( Abc_Tt6Cofactor0(Func, nVars), pFanins, nVars );
    t1 = Dau_Dsd6TruthCompose_rec( Abc_Tt6Cofactor1(Func, nVars), pFanins, nVars );
    return (~pFanins[nVars] & t0) | (pFanins[nVars] & t1);
}
func Dau_Dsd6ToTruth_rec(pStr string, p *string, pMatches *int32, pTruths *word) word {
    var fCompl int32;
    if **p == '!' {
        (*p)++, fCompl = 1; }
    if **p >= 'a' && **p <= 'f' { // var
    
        assert(**p - 'a' >= 0 && **p - 'a' < 6, "**p - 'a' >= 0 && **p - 'a' < 6");
        if fCompl != 0 {
            return ~pTruths[**p - 'a'];
        } else {
            return pTruths[**p - 'a'];
        }
    }
    if **p == '(' { // and/or
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var Res word  = ~(word)0;
        assert(**p == '(' && *q == ')', "**p == '(' && *q == ')'");
        for  (*p)++; *p < q; (*p)++  {
            Res &= Dau_Dsd6ToTruth_rec( pStr, p, pMatches, pTruths );
        }
        assert(*p == q, "*p == q");
        if fCompl != 0 {
            return ~Res;
        } else {
            return Res;
        }
    }
    if **p == '[' { // xor
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var Res word;
        assert(**p == '[' && *q == ']', "**p == '[' && *q == ']'");
        for  (*p)++; *p < q; (*p)++  {
            Res ^= Dau_Dsd6ToTruth_rec( pStr, p, pMatches, pTruths );
        }
        assert(*p == q, "*p == q");
        if fCompl != 0 {
            return ~Res;
        } else {
            return Res;
        }
    }
    if **p == '<' { // mux
    
        var nVars int32;
        var Temp [3]word;
        var pTemp *word = Temp;
        var Res word;
        var Fanins [6]word;
        var pTruths2 *word;
        var pOld string  = *p;
        var q string  = pStr + pMatches[ *p - pStr ];
        // read fanins
        if *(q+1) == '{' {
            var q2 string;
            *p = q+1;
            q2 = pStr + pMatches[ *p - pStr ];
            assert(**p == '{' && *q2 == '}', "**p == '{' && *q2 == '}'");
            for  nVars = 0, (*p)++; *p < q2; (*p)++, nVars++  {
                Fanins[nVars] = Dau_Dsd6ToTruth_rec( pStr, p, pMatches, pTruths );
            }
            assert(*p == q2, "*p == q2");
            pTruths2 = Fanins;
        
        } else {
            pTruths2 = pTruths; }
        // read MUX
        *p = pOld;
        q = pStr + pMatches[ *p - pStr ];
        assert(**p == '<' && *q == '>', "**p == '<' && *q == '>'");
        // verify internal variables
        if nVars != 0 {
            for  ; pOld < q; pOld++  {
                if *pOld >= 'a' && *pOld <= 'z' {
                    assert(*pOld - 'a' < nVars, "*pOld - 'a' < nVars"); }
            } }
        // derive MAX components
        for  (*p)++; *p < q; (*p)++  {
            *pTemp++ = Dau_Dsd6ToTruth_rec( pStr, p, pMatches, pTruths2 );
        }
        assert(pTemp == Temp + 3, "pTemp == Temp + 3");
        assert(*p == q, "*p == q");
        if *(q+1) == '{' { // and/or
        
            var q string  = pStr + pMatches[ ++(*p) - pStr ];
            assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
            *p = q;
        }
        Res = (Temp[0] & Temp[1]) | (~Temp[0] & Temp[2]);
        if fCompl != 0 {
            return ~Res;
        } else {
            return Res;
        }
    }
    if (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9') {
        var Func word;
        var Fanins [6]word;
        var Res word;
        var q string;
        var i int32;
        var nVars int32 = Abc_TtReadHex( &Func, *p );
        *p += Abc_TtHexDigitNum( nVars );
        q = pStr + pMatches[ *p - pStr ];
        assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
        for  i = 0, (*p)++; *p < q; (*p)++, i++  {
            Fanins[i] = Dau_Dsd6ToTruth_rec( pStr, p, pMatches, pTruths );
        }
        assert(i == nVars, "i == nVars");
        assert(*p == q, "*p == q");
        Res = Dau_Dsd6TruthCompose_rec( Func, Fanins, nVars );
        if fCompl != 0 {
            return ~Res;
        } else {
            return Res;
        }
    }
    assert(false, "0");
    return 0;
}
func Dau_Dsd6ToTruth(p string) word {
    var Res word;
    if *p == '0' && *(p+1) == 0 {
        Res = 0;
    } else if *p == '1' && *(p+1) == 0 {
        Res = ~(word)0;
    } else {
        Res = Dau_Dsd6ToTruth_rec( p, &p, Dau_DsdComputeMatches(p), s_Truths6 ); }
    assert(*++p == 0, "*++p == 0");
    return Res;
}

/**Function*************************************************************

  Synopsis    [Computes truth table for the DSD formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdTruth6Compose_rec(Func word, pFanins [][DAU_MAX_WORD]word, pRes *word, nVars int32, nWordsR int32) {
    if Func == 0 {
        Abc_TtConst0( pRes, nWordsR );
        return;
    }
    if Func == ~(word)0 {
        Abc_TtConst1( pRes, nWordsR );
        return;
    }
    assert(nVars > 0, "nVars > 0");
    if nVars--; nVars == 0 {
        assert(Func == s_Truths6[0] || Func == s_Truths6Neg[0], "Func == s_Truths6[0] || Func == s_Truths6Neg[0]");
        Abc_TtCopy( pRes, pFanins[0], nWordsR, Func == s_Truths6Neg[0] );
        return;
    }
    if !Abc_Tt6HasVar(Func, nVars) {
        Dau_DsdTruth6Compose_rec( Func, pFanins, pRes, nVars, nWordsR );
        return;
    }
    {
        var pTtTemp [2][DAU_MAX_WORD]word;
        Dau_DsdTruth6Compose_rec( Abc_Tt6Cofactor0(Func, nVars), pFanins, pTtTemp[0], nVars, nWordsR );
        Dau_DsdTruth6Compose_rec( Abc_Tt6Cofactor1(Func, nVars), pFanins, pTtTemp[1], nVars, nWordsR );
        Abc_TtMux( pRes, pFanins[nVars], pTtTemp[1], pTtTemp[0], nWordsR );
        return;
    }
}
func Dau_DsdTruthCompose_rec(pFunc *word, pFanins [][DAU_MAX_WORD]word, pRes *word, nVars int32, nWordsR int32) {
    var nWordsF int32;
    if nVars <= 6 {
        Dau_DsdTruth6Compose_rec( pFunc[0], pFanins, pRes, nVars, nWordsR );
        return;
    }
    nWordsF = Abc_TtWordNum( nVars );
    assert(nWordsF > 1, "nWordsF > 1");
    if Abc_TtIsConst0(pFunc, nWordsF) {
        Abc_TtConst0( pRes, nWordsR );
        return;
    }
    if Abc_TtIsConst1(pFunc, nWordsF) {
        Abc_TtConst1( pRes, nWordsR );
        return;
    }
    if !Abc_TtHasVar( pFunc, nVars, nVars-1 ) {
        Dau_DsdTruthCompose_rec( pFunc, pFanins, pRes, nVars-1, nWordsR );
        return;
    }
    {
        var pTtTemp [2][DAU_MAX_WORD]word;
        nVars--;
        Dau_DsdTruthCompose_rec( pFunc,             pFanins, pTtTemp[0], nVars, nWordsR );
        Dau_DsdTruthCompose_rec( pFunc + nWordsF/2, pFanins, pTtTemp[1], nVars, nWordsR );
        Abc_TtMux( pRes, pFanins[nVars], pTtTemp[1], pTtTemp[0], nWordsR );
        return;
    }
}
func Dau_DsdToTruth_rec(pStr string, p *string, pMatches *int32, pTtElems **word, pRes *word, nVars int32) {
    var nWords int32  = Abc_TtWordNum( nVars );
    var fCompl int32;
    if **p == '!' {
        (*p)++, fCompl = 1; }
    if **p >= 'a' && **p <= 'z' { // var
    
        assert(**p - 'a' >= 0 && **p - 'a' < nVars, "**p - 'a' >= 0 && **p - 'a' < nVars");
        Abc_TtCopy( pRes, pTtElems[**p - 'a'], nWords, fCompl );
        return;
    }
    if **p == '(' { // and/or
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var pTtTemp [DAU_MAX_WORD]word;
        assert(**p == '(' && *q == ')', "**p == '(' && *q == ')'");
        Abc_TtConst1( pRes, nWords );
        for  (*p)++; *p < q; (*p)++  {
            Dau_DsdToTruth_rec( pStr, p, pMatches, pTtElems, pTtTemp, nVars );
            Abc_TtAnd( pRes, pRes, pTtTemp, nWords, 0 );
        }
        assert(*p == q, "*p == q");
        if fCompl != 0 { Abc_TtNot( pRes, nWords ); }
        return;
    }
    if **p == '[' { // xor
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var pTtTemp [DAU_MAX_WORD]word;
        assert(**p == '[' && *q == ']', "**p == '[' && *q == ']'");
        Abc_TtConst0( pRes, nWords );
        for  (*p)++; *p < q; (*p)++  {
            Dau_DsdToTruth_rec( pStr, p, pMatches, pTtElems, pTtTemp, nVars );
            Abc_TtXor( pRes, pRes, pTtTemp, nWords, 0 );
        }
        assert(*p == q, "*p == q");
        if fCompl != 0 { Abc_TtNot( pRes, nWords ); }
        return;
    }
    if **p == '<' { // mux
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var pTtTemp [3][DAU_MAX_WORD]word;
        var i int32;
        assert(**p == '<' && *q == '>', "**p == '<' && *q == '>'");
        for  i = 0, (*p)++; *p < q; (*p)++, i++  {
            Dau_DsdToTruth_rec( pStr, p, pMatches, pTtElems, pTtTemp[i], nVars );
        }
        assert(i == 3, "i == 3");
        Abc_TtMux( pRes, pTtTemp[0], pTtTemp[1], pTtTemp[2], nWords );
        assert(*p == q, "*p == q");
        if fCompl != 0 { Abc_TtNot( pRes, nWords ); }
        return;
    }
    if (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9') {
        var pFanins [DAU_MAX_VAR][DAU_MAX_WORD]word;
        var pFunc [DAU_MAX_WORD]word;
        var q string;
        var i int32;
        var nVarsF int32 = Abc_TtReadHex( pFunc, *p );
        *p += Abc_TtHexDigitNum( nVarsF );
        q = pStr + pMatches[ *p - pStr ];
        assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
        for  i = 0, (*p)++; *p < q; (*p)++, i++  {
            Dau_DsdToTruth_rec( pStr, p, pMatches, pTtElems, pFanins[i], nVars );
        }
        assert(i == nVarsF, "i == nVarsF");
        assert(*p == q, "*p == q");
        Dau_DsdTruthCompose_rec( pFunc, pFanins, pRes, nVarsF, nWords );
        if fCompl != 0 { Abc_TtNot( pRes, nWords ); }
        return;
    }
    assert(false, "0");
}
func Dau_DsdToTruth(p string, nVars int32) *word {
    var nWords int32  = Abc_TtWordNum( nVars );
    var pTtElems **word  = Dau_DsdTtElems();
    var pRes *word  = pTtElems[DAU_MAX_VAR];
    assert(nVars <= DAU_MAX_VAR, "nVars <= DAU_MAX_VAR");
    if Dau_DsdIsConst0(p) {
        Abc_TtConst0( pRes, nWords );
    } else if Dau_DsdIsConst1(p) {
        Abc_TtConst1( pRes, nWords );
    } else {
        Dau_DsdToTruth_rec( p, &p, Dau_DsdComputeMatches(p), pTtElems, pRes, nVars ); }
    assert(*++p == 0, "*++p == 0");
    return pRes;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdTest2() {
//    char * p = Abc_UtilStrsav( "!(ab!(de[cf]))" );
//    char * p = Abc_UtilStrsav( "!(a![d<ecf>]b)" );
//    word t = Dau_Dsd6ToTruth( p );
}



/**Function*************************************************************

  Synopsis    [Performs DSD.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dau_DsdPerformReplace_pTemp [DAU_MAX_STR]byte;
func Dau_DsdPerformReplace(pBuffer string, PosStart int32, Pos int32, Symb int32, pNext string) int32 {
    
    var pCur string  = spStatic_Dau_DsdPerformReplace_pTemp;
    var i, k, RetValue int32;
    for  i = PosStart; i < Pos; i++  {
        if pBuffer[i] != Symb {
            *pCur++ = pBuffer[i];
        } else {
            for  k = 0; pNext[k]; k++  {
                *pCur++ = pNext[k];
            } }
    }
    RetValue = PosStart + (pCur - spStatic_Dau_DsdPerformReplace_pTemp);
    for  i = PosStart; i < RetValue; i++  {
        pBuffer[i] = spStatic_Dau_DsdPerformReplace_pTemp[i-PosStart];
    }
    return RetValue;
}
func Dau_DsdPerform_rec(t word, pBuffer string, Pos int32, pVars *int32, nVars int32) int32 {
    var pNest string // was [10]byte in C

    var Cof0 [6]word;
    var Cof1 [6]word;
    var Cof [4]word;
    var pVarsNew [6]int32;
    var nVarsNew int32;
    var PosStart int32;
    var v, u, vBest, CountBest int32;
    assert(Pos < DAU_MAX_STR, "Pos < DAU_MAX_STR");
    // perform support minimization
    nVarsNew = 0;
    for  v = 0; v < nVars; v++  {
        if Abc_Tt6HasVar( t, pVars[v] ) {
            pVarsNew[ nVarsNew++ ] = pVars[v]; }
    }
    assert(nVarsNew > 0, "nVarsNew > 0");
    // special case when function is a var
    if nVarsNew == 1 {
        if t == s_Truths6[ pVarsNew[0] ] {
            pBuffer[Pos++] = 'a' +  pVarsNew[0];
            return Pos;
        }
        if t == ~s_Truths6[ pVarsNew[0] ] {
            pBuffer[Pos++] = '!';
            pBuffer[Pos++] = 'a' +  pVarsNew[0];
            return Pos;
        }
        assert(false, "0");
        return Pos;
    }
    // decompose on the output side
    for  v = 0; v < nVarsNew; v++  {
        Cof0[v] = Abc_Tt6Cofactor0( t, pVarsNew[v] );
        Cof1[v] = Abc_Tt6Cofactor1( t, pVarsNew[v] );
        assert(Cof0[v] != Cof1[v], "Cof0[v] != Cof1[v]");
        if Cof0[v] == 0 { // ax
        
            pBuffer[Pos++] = '(';
            pBuffer[Pos++] = 'a' + pVarsNew[v];
            Pos = Dau_DsdPerform_rec( Cof1[v], pBuffer, Pos, pVarsNew, nVarsNew );
            pBuffer[Pos++] = ')';
            return Pos;
        }
        if Cof0[v] == ~(word)0 { // !(ax)
        
            pBuffer[Pos++] = '!';
            pBuffer[Pos++] = '(';
            pBuffer[Pos++] = 'a' + pVarsNew[v];
            Pos = Dau_DsdPerform_rec( ~Cof1[v], pBuffer, Pos, pVarsNew, nVarsNew );
            pBuffer[Pos++] = ')';
            return Pos;
        }
        if Cof1[v] == 0 { // !ax
        
            pBuffer[Pos++] = '(';
            pBuffer[Pos++] = '!';
            pBuffer[Pos++] = 'a' + pVarsNew[v];
            Pos = Dau_DsdPerform_rec( Cof0[v], pBuffer, Pos, pVarsNew, nVarsNew );
            pBuffer[Pos++] = ')';
            return Pos;
        }
        if Cof1[v] == ~(word)0 { // !(!ax)
        
            pBuffer[Pos++] = '!';
            pBuffer[Pos++] = '(';
            pBuffer[Pos++] = '!';
            pBuffer[Pos++] = 'a' + pVarsNew[v];
            Pos = Dau_DsdPerform_rec( ~Cof0[v], pBuffer, Pos, pVarsNew, nVarsNew );
            pBuffer[Pos++] = ')';
            return Pos;
        }
        if Cof0[v] == ~Cof1[v] { // a^x
        
            pBuffer[Pos++] = '[';
            pBuffer[Pos++] = 'a' + pVarsNew[v];
            Pos = Dau_DsdPerform_rec( Cof0[v], pBuffer, Pos, pVarsNew, nVarsNew );
            pBuffer[Pos++] = ']';
            return Pos;
        }
    }
    // decompose on the input side
    for  v = 0; v < nVarsNew; v++  {
    for  u = v+1; u < nVarsNew; u++  {
        Cof[0] = Abc_Tt6Cofactor0( Cof0[v], pVarsNew[u] );
        Cof[1] = Abc_Tt6Cofactor1( Cof0[v], pVarsNew[u] );
        Cof[2] = Abc_Tt6Cofactor0( Cof1[v], pVarsNew[u] );
        Cof[3] = Abc_Tt6Cofactor1( Cof1[v], pVarsNew[u] );
        if Cof[0] == Cof[1] && Cof[0] == Cof[2] { // vu
        
            PosStart = Pos;
            pNest = fmt.Sprintf("(%c%c)", 'a' + pVarsNew[v], 'a' + pVarsNew[u]);
            Pos = Dau_DsdPerform_rec( (s_Truths6[pVarsNew[u]] & Cof[3]) | (~s_Truths6[pVarsNew[u]] & Cof[0]), pBuffer, Pos, pVarsNew, nVarsNew );
            Pos = Dau_DsdPerformReplace( pBuffer, PosStart, Pos, 'a' + pVarsNew[u], pNest );
            return Pos;
        }
        if Cof[0] == Cof[1] && Cof[0] == Cof[3] { // v!u
        
            PosStart = Pos;
            pNest = fmt.Sprintf("(%c!%c)", 'a' + pVarsNew[v], 'a' + pVarsNew[u]);
            Pos = Dau_DsdPerform_rec( (s_Truths6[pVarsNew[u]] & Cof[2]) | (~s_Truths6[pVarsNew[u]] & Cof[0]), pBuffer, Pos, pVarsNew, nVarsNew );
            Pos = Dau_DsdPerformReplace( pBuffer, PosStart, Pos, 'a' + pVarsNew[u], pNest );
            return Pos;
        }
        if Cof[0] == Cof[2] && Cof[0] == Cof[3] { // !vu
        
            PosStart = Pos;
            pNest = fmt.Sprintf("(!%c%c)", 'a' + pVarsNew[v], 'a' + pVarsNew[u]);
            Pos = Dau_DsdPerform_rec( (s_Truths6[pVarsNew[u]] & Cof[1]) | (~s_Truths6[pVarsNew[u]] & Cof[0]), pBuffer, Pos, pVarsNew, nVarsNew );
            Pos = Dau_DsdPerformReplace( pBuffer, PosStart, Pos, 'a' + pVarsNew[u], pNest );
            return Pos;
        }
        if Cof[1] == Cof[2] && Cof[1] == Cof[3] { // !v!u
        
            PosStart = Pos;
            pNest = fmt.Sprintf("(!%c!%c)", 'a' + pVarsNew[v], 'a' + pVarsNew[u]);
            Pos = Dau_DsdPerform_rec( (s_Truths6[pVarsNew[u]] & Cof[0]) | (~s_Truths6[pVarsNew[u]] & Cof[1]), pBuffer, Pos, pVarsNew, nVarsNew );
            Pos = Dau_DsdPerformReplace( pBuffer, PosStart, Pos, 'a' + pVarsNew[u], pNest );
            return Pos;
        }
        if Cof[0] == Cof[3] && Cof[1] == Cof[2] { // v+u
        
            PosStart = Pos;
            pNest = fmt.Sprintf("[%c%c]", 'a' + pVarsNew[v], 'a' + pVarsNew[u]);
            Pos = Dau_DsdPerform_rec( (s_Truths6[pVarsNew[u]] & Cof[1]) | (~s_Truths6[pVarsNew[u]] & Cof[0]), pBuffer, Pos, pVarsNew, nVarsNew );
            Pos = Dau_DsdPerformReplace( pBuffer, PosStart, Pos, 'a' + pVarsNew[u], pNest );
            return Pos;
        }
    }
    } 
    // find best variable for MUX decomposition
    vBest = -1;
    CountBest = 10;
    for  v = 0; v < nVarsNew; v++  {
        var CountCur int32;
        for  u = 0; u < nVarsNew; u++  {
            if u != v && Abc_Tt6HasVar(Cof0[v], pVarsNew[u]) && Abc_Tt6HasVar(Cof1[v], pVarsNew[u]) {
                CountCur++; }
        }
        if CountBest > CountCur {
            CountBest = CountCur;
            vBest = v;
        }
        if CountCur == 0 {
            break; }
    }
    // perform MUX decomposition
    pBuffer[Pos++] = '<';
    pBuffer[Pos++] = 'a' + pVarsNew[vBest];
    Pos = Dau_DsdPerform_rec( Cof1[vBest], pBuffer, Pos, pVarsNew, nVarsNew );
    Pos = Dau_DsdPerform_rec( Cof0[vBest], pBuffer, Pos, pVarsNew, nVarsNew );
    pBuffer[Pos++] = '>';
    return Pos;
}
var spStatic_Dau_DsdPerform_pBuffer [DAU_MAX_STR]byte;
func Dau_DsdPerform(t word) string {
    
    var pVarsNew [6]int32 = [6]int32{0, 1, 2, 3, 4, 5};
    var Pos int32;
    if t == 0 {
        spStatic_Dau_DsdPerform_pBuffer[Pos++] = '0';
    } else if t == ~(word)0 {
        spStatic_Dau_DsdPerform_pBuffer[Pos++] = '1';
    } else {
        Pos = Dau_DsdPerform_rec( t, spStatic_Dau_DsdPerform_pBuffer, Pos, pVarsNew, 6 ); }
    spStatic_Dau_DsdPerform_pBuffer[Pos++] = 0;
//    printf( "%d ", strlen(pBuffer) );
//    printf( "%s ->", pBuffer );
    Dau_DsdRemoveBraces( spStatic_Dau_DsdPerform_pBuffer, Dau_DsdComputeMatches(spStatic_Dau_DsdPerform_pBuffer) );
//    printf( " %s\n", pBuffer );
    return spStatic_Dau_DsdPerform_pBuffer;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdTest3() {
//    word t = s_Truths6[0] & s_Truths6[1] & s_Truths6[2];
//    word t = ~s_Truths6[0] | (s_Truths6[1] ^ ~s_Truths6[2]);
//    word t = (s_Truths6[1] & s_Truths6[2]) | (s_Truths6[0] & s_Truths6[3]);
//    word t = (~s_Truths6[1] & ~s_Truths6[2]) | (s_Truths6[0] ^ s_Truths6[3]);
//    word t = ((~s_Truths6[1] & ~s_Truths6[2]) | (s_Truths6[0] ^ s_Truths6[3])) ^ s_Truths6[5];
//    word t = ((s_Truths6[1] & ~s_Truths6[2]) ^ (s_Truths6[0] & s_Truths6[3])) & s_Truths6[5];
//    word t = (~(~s_Truths6[0] & ~s_Truths6[4]) & s_Truths6[2]) | (~s_Truths6[1] & ~s_Truths6[0] & ~s_Truths6[4]);
//    word t = 0x0000000000005F3F;
//    word t = 0xF3F5030503050305;
//    word t = (s_Truths6[0] & s_Truths6[1] & (s_Truths6[2] ^ s_Truths6[4])) | (~s_Truths6[0] & ~s_Truths6[1] & ~(s_Truths6[2] ^ s_Truths6[4]));
//    word t = 0x05050500f5f5f5f3;
    var t word  = ABC_CONST(0x9ef7a8d9c7193a0f);
    var p string  = Dau_DsdPerform( t );
    var t2 word  = Dau_Dsd6ToTruth( p );
    if t != t2 {
        fmt.Printf("Verification failed.\n"); }
}




/**Function*************************************************************

  Synopsis    [Find the best cofactoring variable.]

  Description [Return -2 if non-DSD; -1 if full DSD; otherwise,
  returns cofactoring variables i (i >= 0).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdCheck1Step(p any, pTruth *word, nVarsInit int32, pVarLevels *int32) int32 {
    var pCofTemp [DAU_MAX_WORD]word;
    var pVarPrios [DAU_MAX_VAR]int32;
    var nWords int32  = Abc_TtWordNum(nVarsInit);
    var nSizeNonDec, nSizeNonDec0, nSizeNonDec1 int32;
    var i int32;
    var vBest int32 = -2;
    var nSumCofsBest int32 = ABC_INFINITY;
    var nSumCofs int32;
    nSizeNonDec = Dau_DsdDecompose( pTruth, nVarsInit, 0, 0, nil );
    if nSizeNonDec == 0 {
        return -1; }
    assert(nSizeNonDec > 0, "nSizeNonDec > 0");
    // find variable priority
    for  i = 0; i < nVarsInit; i++  {
        pVarPrios[i] = i;
    }
    if pVarLevels != nil {
        extern int Dau_DsdLevelVar( void * pMan, int iVar );
        var pVarLevels [DAU_MAX_VAR]int32;
        for  i = 0; i < nVarsInit; i++  {
            pVarLevels[i] = -Dau_DsdLevelVar( p, i );
        }
//        for ( i = 0; i < nVarsInit; i++ )
//            printf( "%d ", -pVarLevels[i] );
//        printf( "\n" );
        Vec_IntSelectSortCost2( pVarPrios, nVarsInit, pVarLevels );
//        for ( i = 0; i < nVarsInit; i++ )
//            printf( "%d ", pVarPrios[i] );
//        printf( "\n\n" );
    }
    for  i = 0; i < nVarsInit; i++  {
        assert(pVarPrios[i] >= 0 && pVarPrios[i] < nVarsInit, "pVarPrios[i] >= 0 && pVarPrios[i] < nVarsInit");
        // try first cofactor
        Abc_TtCofactor0p( pCofTemp, pTruth, nWords, pVarPrios[i] );
        nSumCofs = Abc_TtSupportSize( pCofTemp, nVarsInit );
        nSizeNonDec0 = Dau_DsdDecompose( pCofTemp, nVarsInit, 0, 0, nil );
        // try second cofactor
        Abc_TtCofactor1p( pCofTemp, pTruth, nWords, pVarPrios[i] );
        nSumCofs += Abc_TtSupportSize( pCofTemp, nVarsInit );
        nSizeNonDec1 = Dau_DsdDecompose( pCofTemp, nVarsInit, 0, 0, nil );
        // compare cofactors
        if nSizeNonDec0 || (nSizeNonDec1 != 0) {
            continue; }
        if nSumCofsBest > nSumCofs {
            vBest = pVarPrios[i];
            nSumCofsBest = nSumCofs;
        }
    }
    return vBest;
}

/**Function*************************************************************

  Synopsis    [Data-structure to store DSD information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
   
type Dau_Dsd_t struct {
    nVarsInit      int32;            // the initial number of variables
    nVarsUsed      int32;            // the current number of variables
    nPos      int32;                 // writing position
    nSizeNonDec      int32;          // size of the largest non-decomposable block
    nConsts      int32;              // the number of constant decompositions
    uConstMask      int32;           // constant decomposition mask
    fSplitPrime      int32;          // represent prime function as 1-step DSD
    fWriteTruth      int32;          // writing truth table as a hex string
    pVarLevels     *int32;           // variable levels
    pVarDefs [32][8]byte;      // variable definitions
    Cache [32][32]byte;        // variable cache
    pOutput [DAU_MAX_STR]byte; // output stream
}

var s_Times [3]abctime ;

/**Function*************************************************************

  Synopsis    [Manipulation of DSD data-structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdInitialize(p *Dau_Dsd_t, nVarsInit int32) {
    var i, v, u int32;
    assert(nVarsInit >= 0 && nVarsInit <= 16, "nVarsInit >= 0 && nVarsInit <= 16");
    p.nVarsInit   = nVarsInit;
    p.nVarsUsed   = nVarsInit;
    p.nPos        = 0;
    p.nSizeNonDec = 0;
    p.nConsts     = 0;
    p.uConstMask  = 0;
    for  i = 0; i < nVarsInit; i++  {
        p.pVarDefs[i][0] = 'a' + i, p.pVarDefs[i][1] = 0;
    }
    for  v = 0; v < nVarsInit; v++  {
    for  u = 0; u < nVarsInit; u++  {
        p.Cache[v][u] = 0;
    }
    }

}
func Dau_DsdWriteString(p *Dau_Dsd_t, pStr string) {
    for  *pStr  {
        p.pOutput[ p.nPos++ ] = *pStr++;
    }
}
func Dau_DsdWriteVar(p *Dau_Dsd_t, iVar int32, fInv int32) {
    var pStr string;
    if fInv != 0 {
        p.pOutput[ p.nPos++ ] = '!'; }
    for  pStr = p.pVarDefs[iVar]; *pStr; pStr++  {
        if *pStr >= 'a' + p.nVarsInit && *pStr < 'a' + p.nVarsUsed {
            Dau_DsdWriteVar( p, *pStr - 'a', 0 );
        } else {
            p.pOutput[ p.nPos++ ] = *pStr; }
    }
}
func Dau_DsdLevelVar(pMan any, iVar int32) int32 {
    var p *Dau_Dsd_t  = (Dau_Dsd_t *)pMan;
    var pStr string;
    var LevelMax, Level int32;
    for  pStr = p.pVarDefs[iVar]; *pStr; pStr++  {
        if *pStr >= 'a' + p.nVarsInit && *pStr < 'a' + p.nVarsUsed {
            Level = 1 + Dau_DsdLevelVar( p, *pStr - 'a' );
        } else {
            Level = p.pVarLevels[*pStr - 'a']; }
        LevelMax = Abc_MaxInt( LevelMax, Level );
    }
    return LevelMax;
}
func Dau_DsdTranslate(p *Dau_Dsd_t, pVars *int32, nVars int32, pStr string) {
    for  ; *pStr; pStr++  {
        if *pStr >= 'a' && *pStr < 'a' + nVars {
            Dau_DsdWriteVar( p, pVars[*pStr - 'a'], 0 );
        } else {
            p.pOutput[ p.nPos++ ] = *pStr; }
    }
}
func Dau_DsdWritePrime(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    var v int32;
    var RetValue int32 = 2;
    assert(nVars > 2, "nVars > 2");
    if p.fSplitPrime {
        var pCofTemp [DAU_MAX_WORD]word;
        var nWords int32  = Abc_TtWordNum(nVars);
        var vBest int32  = Dau_DsdCheck1Step( p, pTruth, nVars, p.pVarLevels );
        assert(vBest != -1, "vBest != -1");
        if vBest == -2 { // non-dec
            p.nPos += Abc_TtWriteHexRev( p.pOutput + p.nPos, pTruth, nVars );
        } else {
            var pRes [DAU_MAX_STR]byte;
            var nNonDecSize int32;
            // compose the result
            Dau_DsdWriteString( p, "<" );
            Dau_DsdWriteVar( p, vBest, 0 );
            // split decomposition
            Abc_TtCofactor1p( pCofTemp, pTruth, nWords, vBest );
            nNonDecSize = Dau_DsdDecompose( pCofTemp, nVars, 0, p.fWriteTruth, pRes );
            assert(nNonDecSize == 0, "nNonDecSize == 0");
            Dau_DsdWriteString( p, pRes );
            // split decomposition
            Abc_TtCofactor0p( pCofTemp, pTruth, nWords, vBest );
            nNonDecSize = Dau_DsdDecompose( pCofTemp, nVars, 0, p.fWriteTruth, pRes );
            assert(nNonDecSize == 0, "nNonDecSize == 0");
            Dau_DsdWriteString( p, pRes );
            Dau_DsdWriteString( p, ">" );
            RetValue = 1;
        }
    
    } else if p.fWriteTruth {
        p.nPos += Abc_TtWriteHexRev( p.pOutput + p.nPos, pTruth, nVars ); }
    Dau_DsdWriteString( p, "{" );
    for  v = 0; v < nVars; v++  {
        Dau_DsdWriteVar( p, pVars[v], 0 );
    }
    Dau_DsdWriteString( p, "}" );
    p.nSizeNonDec = nVars;
    return RetValue;
}
func Dau_DsdFinalize(p *Dau_Dsd_t) {
    var i int32;
    for  i = 0; i < p.nConsts; i++  {
        p.pOutput[ p.nPos++ ] = ((p.uConstMask >> (p.nConsts-1-i)) & 1) ? ']' : ')';
    }
    p.pOutput[ p.nPos++ ] = 0;
}
func Dau_DsdAddVarDef(p *Dau_Dsd_t, pStr string) int32 {
    var u int32;
    assert(len(pStr) < 8, "strlen(pStr) < 8");
    assert(p.nVarsUsed < 32, "p.nVarsUsed < 32");
    for  u = 0; u < p.nVarsUsed; u++  {
        p.Cache[p.nVarsUsed][u] = 0;
    }
    for  u = 0; u < p.nVarsUsed; u++  {
        p.Cache[u][p.nVarsUsed] = 0;
    }
    sprintf( p.pVarDefs[p.nVarsUsed++], "%s", pStr );
    return p.nVarsUsed - 1;
}
func Dau_DsdFindVarDef(pVars *int32, nVars int32, VarDef int32) int32 {
    var v int32;
    for  v = 0; v < nVars; v++  {
        if pVars[v] == VarDef {
            break; }
    }
    assert(v < nVars, "v < nVars");
    return v;
}
func Dau_DsdInsertVarCache(p *Dau_Dsd_t, v int32, u int32, Status int32) {
    assert(v != u, "v != u");
    assert(Status > 0 && Status < 4, "Status > 0 && Status < 4");
    assert(p.Cache[v][u] == 0, "p.Cache[v][u] == 0");
    p.Cache[v][u] = Status;
}
func Dau_DsdLookupVarCache(p *Dau_Dsd_t, v int32, u int32) int32 {
    return p.Cache[v][u];
}

/**Function*************************************************************

  Synopsis    [Procedures specialized for 6-variable functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_Dsd6DecomposeSingleVarOne(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32) int32 {
    // consider negative cofactors
    if pTruth[0] & 1 {        
        if Abc_Tt6Cof0IsConst1( pTruth[0], v ) { // !(ax)
        
            Dau_DsdWriteString( p, "!(" );
            pTruth[0] = ~Abc_Tt6Cofactor1( pTruth[0], v );
            goto finish;            
        }
    
    } else {
        if Abc_Tt6Cof0IsConst0( pTruth[0], v ) { // ax
        
            Dau_DsdWriteString( p, "(" );
            pTruth[0] = Abc_Tt6Cofactor1( pTruth[0], v );
            goto finish;            
        }
    }
    // consider positive cofactors
    if pTruth[0] >> 63 {        
        if Abc_Tt6Cof1IsConst1( pTruth[0], v ) { // !(!ax)
        
            Dau_DsdWriteString( p, "!(!" );
            pTruth[0] = ~Abc_Tt6Cofactor0( pTruth[0], v );
            goto finish;            
        }
    
    } else {
        if Abc_Tt6Cof1IsConst0( pTruth[0], v ) { // !ax
        
            Dau_DsdWriteString( p, "(!" );
            pTruth[0] = Abc_Tt6Cofactor0( pTruth[0], v );
            goto finish;            
        }
    }
    // consider equal cofactors
    if Abc_Tt6CofsOpposite( pTruth[0], v ) { // [ax]
    
        Dau_DsdWriteString( p, "[" );
        pTruth[0] = Abc_Tt6Cofactor0( pTruth[0], v );
        p.uConstMask |= (1 << p.nConsts);
        goto finish;            
    }
    return 0;

finish:
    p.nConsts++;
    Dau_DsdWriteVar( p, pVars[v], 0 );
    pVars[v] = pVars[nVars-1];
    Abc_TtSwapVars( pTruth, nVars, v, nVars-1 );
    return 1;
}
func Dau_Dsd6DecomposeSingleVar(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    var clk abctime  = Abc_Clock();
    assert(nVars > 1, "nVars > 1");
    for {
        var v int32;
        for  v = nVars - 1; v >= 0 && nVars > 1; v--  {
            if Dau_Dsd6DecomposeSingleVarOne( p, pTruth, pVars, nVars, v ) != 0 {
                nVars--;
                break;
            }
        }
        if v == -1 || nVars == 1 {
            break; }
    }
    if nVars == 1 {
        Dau_DsdWriteVar( p, pVars[--nVars], int32((pTruth[0] & 1)) ); }
    s_Times[0] += Abc_Clock() - clk;
    return nVars;
}
func Dau_Dsd6FindSupportOne(p *Dau_Dsd_t, tCof0 word, tCof1 word, pVars *int32, nVars int32, v int32, u int32) int32 {
    var Status int32;
    if p {
        Status = Dau_DsdLookupVarCache( p, pVars[v], pVars[u] );
    } else {
        Status = 0;
    }
    if Status == 0 {
        Status = (Abc_Tt6HasVar(tCof1, u) << 1) | Abc_Tt6HasVar(tCof0, u);
        if p {
            Dau_DsdInsertVarCache( p, pVars[v], pVars[u], Status ); }
    }
    return Status;
}
func Dau_Dsd6FindSupports(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32) uint32 {
    var u int32;
    var uSupports uint32;
    var tCof0 word  = Abc_Tt6Cofactor0( pTruth[0], v );
    var tCof1 word  = Abc_Tt6Cofactor1( pTruth[0], v );
//Kit_DsdPrintFromTruth( (unsigned *)&tCof0, 6 );printf( "\n" );
//Kit_DsdPrintFromTruth( (unsigned *)&tCof1, 6 );printf( "\n" );
    for  u = 0; u < nVars; u++  {
        if u != v {
            uSupports |= (Dau_Dsd6FindSupportOne( p, tCof0, tCof1, pVars, nVars, v, u ) << (2 * u)); }
    }
    return uSupports;
}
func Dau_DsdPrintSupports(uSupp uint32, nVars int32) {
    var v, Value int32;
    fmt.Printf("Cofactor supports: ");
    for  v = nVars - 1; v >= 0; v--  {
        Value = ((uSupp >> (2*v)) & 3);
        if Value == 1 {
            fmt.Printf("01");
        } else if Value == 2 {
            fmt.Printf("10");
        } else if Value == 3 {
            fmt.Printf("11");
        } else { 
            fmt.Printf("00"); }
        if v != 0 {
            fmt.Printf("-"); }
    }
    fmt.Printf("\n");
}
// checks decomposability with respect to the pair (v, u)
func Dau_Dsd6DecomposeDoubleVarsOne(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32, u int32) int32 {
    var pBuffer string // was [10]byte in C
 
    var tCof0 word  = Abc_Tt6Cofactor0( pTruth[0], v );
    var tCof1 word  = Abc_Tt6Cofactor1( pTruth[0], v );
    var Status int32  = Dau_Dsd6FindSupportOne( p, tCof0, tCof1, pVars, nVars, v, u );
    assert(v > u, "v > u");
//printf( "Checking %s and %s.\n", p->pVarDefs[pVars[v]], p->pVarDefs[pVars[u]] );

//    Kit_DsdPrintFromTruth( (unsigned *)pTruth, 6 );printf( "\n" );
    if Status == 3 {
        // both F(v=0) and F(v=1) depend on u
        if Abc_Tt6Cof0EqualCof1(tCof0, tCof1, u) && Abc_Tt6Cof0EqualCof1(tCof1, tCof0, u) { // v+u
        
            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor1(tCof0, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            pBuffer = fmt.Sprintf("[%c%c]", 'a' + pVars[v], 'a' + pVars[u]);
            goto finish;
        }
    
    } else if Status == 2 {
        // F(v=0) does not depend on u; F(v=1) depends on u
        if Abc_Tt6Cof0EqualCof0(tCof0, tCof1, u) { // vu
        
            pBuffer = fmt.Sprintf("(%c%c)", 'a' + pVars[v], 'a' + pVars[u]);
            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor1(tCof1, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            goto finish;
        }
        if Abc_Tt6Cof0EqualCof1(tCof0, tCof1, u) { // v!u
        
            pBuffer = fmt.Sprintf("(%c!%c)", 'a' + pVars[v], 'a' + pVars[u]);
            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor0(tCof1, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            goto finish;
        }
    
    } else if Status == 1 {
        // F(v=0) depends on u; F(v=1) does not depend on u
        if Abc_Tt6Cof0EqualCof1(tCof0, tCof1, u) { // !vu
        
            pBuffer = fmt.Sprintf("(!%c%c)", 'a' + pVars[v], 'a' + pVars[u]);
            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor1(tCof0, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            goto finish;
        }
        if Abc_Tt6Cof1EqualCof1(tCof0, tCof1, u) { // !v!u
        
            pBuffer = fmt.Sprintf("(!%c!%c)", 'a' + pVars[v], 'a' + pVars[u]);
            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor1(tCof1, u));
            goto finish;
        }
    }
    return nVars;
finish: 
    // finalize decomposition
    assert(pBuffer[0], "pBuffer[0]");
    pVars[u] = Dau_DsdAddVarDef( p, pBuffer );
    pVars[v] = pVars[nVars-1];
    Abc_TtSwapVars( pTruth, nVars, v, nVars-1 );
    if Dau_Dsd6DecomposeSingleVarOne( p, pTruth, pVars, --nVars, u ) {
        nVars = Dau_Dsd6DecomposeSingleVar( p, pTruth, pVars, --nVars ); }
    return nVars;
}
func Dau_Dsd6DecomposeDoubleVars(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    var clk abctime  = Abc_Clock();
    for {
        var v, u, nVarsOld int32;
        for  v = nVars - 1; v > 0; v--  {
            for  u = v - 1; u >= 0; u--  {
                if Dau_DsdLookupVarCache( p, pVars[v], pVars[u] ) != 0 {
                    continue; }
                nVarsOld = nVars;
                nVars = Dau_Dsd6DecomposeDoubleVarsOne( p, pTruth, pVars, nVars, v, u );
                if nVars == 0 {
                    s_Times[1] += Abc_Clock() - clk;
                    return 0;
                }
                if nVarsOld > nVars {
                    break; }
            }
            if u >= 0 { // found
                break; }
        }
        if v == 0 { // not found
            break; }
    }
    s_Times[1] += Abc_Clock() - clk;
    return nVars;
}

// look for MUX-decomposable variable on top or at the bottom
func Dau_Dsd6DecomposeTripleVarsOuter(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32) int32 {
    extern int Dau_DsdDecomposeInt( Dau_Dsd_t * p, word * pTruth, int nVarsInit );
    var P1 Dau_Dsd_t;
    var p1 *Dau_Dsd_t = &P1;
    var tCof0, tCof1 word;
    p1.fSplitPrime = 0;
    p1.fWriteTruth = p.fWriteTruth;
    // move this variable to the top
    // inlining of macro ABC_SWAP
    { var t int32  = pVars[v]; pVars[v] = pVars[nVars-1]; pVars[nVars-1] = t; };
    Abc_TtSwapVars( pTruth, nVars, v, nVars-1 );
    // cofactor w.r.t the last variable
    tCof0 = Abc_Tt6Cofactor0( pTruth[0], nVars - 1 );
    tCof1 = Abc_Tt6Cofactor1( pTruth[0], nVars - 1 );
    // compose the result
    Dau_DsdWriteString( p, "<" );
    Dau_DsdWriteVar( p, pVars[nVars - 1], 0 );
    // split decomposition
    Dau_DsdDecomposeInt( p1, &tCof1, nVars - 1 );
    Dau_DsdTranslate( p, pVars, nVars - 1, p1.pOutput );
    p.nSizeNonDec = p1.nSizeNonDec;
    if p1.nSizeNonDec {
        *pTruth = tCof1; }
    // split decomposition
    Dau_DsdDecomposeInt( p1, &tCof0, nVars - 1 );
    Dau_DsdTranslate( p, pVars, nVars - 1, p1.pOutput );
    Dau_DsdWriteString( p, ">" );
    p.nSizeNonDec = Abc_MaxInt( p.nSizeNonDec, p1.nSizeNonDec );
    if p1.nSizeNonDec {
        *pTruth = tCof0; }
    return 0;
}
func Dau_Dsd6DecomposeTripleVarsInner(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32, uSupports uint32) int32 {
    var iVar0 int32  = Abc_TtSuppFindFirst(  uSupports & (~uSupports >> 1) & 0x55555555 ) >> 1;
    var iVar1 int32  = Abc_TtSuppFindFirst( ~uSupports & ( uSupports >> 1) & 0x55555555 ) >> 1;
    var tCof0 word  = Abc_Tt6Cofactor0( pTruth[0], v );
    var tCof1 word  = Abc_Tt6Cofactor1( pTruth[0], v );
    var C00 word  = Abc_Tt6Cofactor0( tCof0, iVar0 );
    var C01 word  = Abc_Tt6Cofactor1( tCof0, iVar0 );
    var C10 word  = Abc_Tt6Cofactor0( tCof1, iVar1 );
    var C11 word  = Abc_Tt6Cofactor1( tCof1, iVar1 );
    var fEqual0 int32  = (C00 == C10) && (C01 == C11);
    var fEqual1 int32  = (C00 == C11) && (C01 == C10);
    if (fEqual0 != 0) || (fEqual1 != 0) {
        var pBuffer string // was [10]byte in C

        var VarId int32  = pVars[iVar0];
        pTruth[0] = (s_Truths6[v] & C11) | (~s_Truths6[v] & C10);
        pBuffer = fmt.Sprintf("<%c%c%s%c>", 'a' + pVars[v], 'a' + pVars[iVar1], fEqual1 ? "!":"", 'a' + pVars[iVar0]);
        pVars[v] = Dau_DsdAddVarDef( p, pBuffer );
        // remove iVar1
        // inlining of macro ABC_SWAP
        { var t int32  = pVars[iVar1]; pVars[iVar1] = pVars[nVars-1]; pVars[nVars-1] = t; };
        Abc_TtSwapVars( pTruth, nVars, iVar1, nVars-1 ); nVars--;
        // remove iVar0
        iVar0 = Dau_DsdFindVarDef( pVars, nVars, VarId );
        // inlining of macro ABC_SWAP
        { var t int32  = pVars[iVar0]; pVars[iVar0] = pVars[nVars-1]; pVars[nVars-1] = t; };
        Abc_TtSwapVars( pTruth, nVars, iVar0, nVars-1 ); nVars--;
        // find the new var
        v = Dau_DsdFindVarDef( pVars, nVars, p.nVarsUsed-1 );
        // remove single variables if possible
        if Dau_Dsd6DecomposeSingleVarOne( p, pTruth, pVars, nVars, v ) != 0 {
            nVars = Dau_Dsd6DecomposeSingleVar( p, pTruth, pVars, --nVars ); }
        return nVars;
    }
    return nVars;
}
func Dau_Dsd6DecomposeTripleVars(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    var clk abctime  = Abc_Clock();
    for {
        var v int32;
//        Kit_DsdPrintFromTruth( (unsigned *)pTruth, 6 ); printf( "\n" );
        for  v = nVars - 1; v >= 0; v--  {
            var uSupports uint32  = Dau_Dsd6FindSupports( p, pTruth, pVars, nVars, v );
//            Dau_DsdPrintSupports( uSupports, nVars );
            if (uSupports & (uSupports >> 1) & 0x55555555) == 0 { // non-overlapping supports
                return Dau_Dsd6DecomposeTripleVarsOuter( p, pTruth, pVars, nVars, v ); }
            if Abc_TtSuppOnlyOne( uSupports & (~uSupports >> 1) & 0x55555555) &&
                 Abc_TtSuppOnlyOne(~uSupports & ( uSupports >> 1) & 0x55555555) { // one unique variable in each cofactor
            
                var nVarsNew int32  = Dau_Dsd6DecomposeTripleVarsInner( p, pTruth, pVars, nVars, v, uSupports );
                if nVarsNew == nVars {
                    continue; }
                if nVarsNew == 0 {
                    s_Times[2] += Abc_Clock() - clk;
                    return 0;
                }
                nVars = Dau_Dsd6DecomposeDoubleVars( p, pTruth, pVars, nVarsNew );
                if nVars == 0 {
                    s_Times[2] += Abc_Clock() - clk;
                    return 0;
                }
                break;
            }
        }
        if v == -1 {
            s_Times[2] += Abc_Clock() - clk;
            return nVars;
        }
    }
    assert(false, "0");
    return -1;
}
func Dau_Dsd6DecomposeInternal(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    // decompose single variales on the output side
    nVars = Dau_Dsd6DecomposeSingleVar( p, pTruth, pVars, nVars );
    if nVars == 0 {
        return 0; }
    // decompose double variables on the input side
    nVars = Dau_Dsd6DecomposeDoubleVars( p, pTruth, pVars, nVars );
    if nVars == 0 {
        return 0; }
    // decompose MUX on the output/input side
    nVars = Dau_Dsd6DecomposeTripleVars( p, pTruth, pVars, nVars );
    if nVars == 0 {
        return 0; }
    // write non-decomposable function
    return Dau_DsdWritePrime( p, pTruth, pVars, nVars );
}

/**Function*************************************************************

  Synopsis    [Procedures specialized for 6-variable functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdDecomposeSingleVarOne(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32) int32 {
    var nWords int32  = Abc_TtWordNum(nVars);
    // consider negative cofactors
    if pTruth[0] & 1 {        
        if Abc_TtCof0IsConst1( pTruth, nWords, v ) { // !(ax)
        
            Dau_DsdWriteString( p, "!(" );
            Abc_TtCofactor1( pTruth, nWords, v );
            Abc_TtNot( pTruth, nWords );
            goto finish;            
        }
    
    } else {
        if Abc_TtCof0IsConst0( pTruth, nWords, v ) { // ax
        
            Dau_DsdWriteString( p, "(" );
            Abc_TtCofactor1( pTruth, nWords, v );
            goto finish;            
        }
    }
    // consider positive cofactors
    if pTruth[nWords-1] >> 63 {        
        if Abc_TtCof1IsConst1( pTruth, nWords, v ) { // !(!ax)
        
            Dau_DsdWriteString( p, "!(!" );
            Abc_TtCofactor0( pTruth, nWords, v );
            Abc_TtNot( pTruth, nWords );
            goto finish;            
        }
    
    } else {
        if Abc_TtCof1IsConst0( pTruth, nWords, v ) { // !ax
        
            Dau_DsdWriteString( p, "(!" );
            Abc_TtCofactor0( pTruth, nWords, v );
            goto finish;            
        }
    }
    // consider equal cofactors
    if Abc_TtCofsOpposite( pTruth, nWords, v ) { // [ax]
    
        Dau_DsdWriteString( p, "[" );
        Abc_TtCofactor0( pTruth, nWords, v );
        p.uConstMask |= (1 << p.nConsts);
        goto finish;            
    }
    return 0;

finish:
    p.nConsts++;
    Dau_DsdWriteVar( p, pVars[v], 0 );
    pVars[v] = pVars[nVars-1];
    Abc_TtSwapVars( pTruth, nVars, v, nVars-1 );
    return 1;
}
func Dau_DsdDecomposeSingleVar(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    var clk abctime  = Abc_Clock();
    assert(nVars > 1, "nVars > 1");
    for {
        var v int32;
        for  v = nVars - 1; v >= 0 && nVars > 1; v--  {
            if Dau_DsdDecomposeSingleVarOne( p, pTruth, pVars, nVars, v ) != 0 {
                nVars--;
                break;
            }
        }
        if v == -1 || nVars == 1 {
            break; }
    }
    if nVars == 1 {
        Dau_DsdWriteVar( p, pVars[--nVars], int32((pTruth[0] & 1)) ); }
    s_Times[0] += Abc_Clock() - clk;
    return nVars;
}

func Dau_DsdFindSupportOne(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32, u int32) int32 {
    var nWords int32  = Abc_TtWordNum(nVars);
    var Status int32;
    if p {
        Status = Dau_DsdLookupVarCache( p, pVars[v], pVars[u] );
    } else {
        Status = 0;
    }
    if Status == 0 {
//        Status = (Abc_Tt6HasVar(tCof1, u) << 1) | Abc_Tt6HasVar(tCof0, u);
        if v < u {
            Status = (!Abc_TtCheckEqualCofs(pTruth, nWords, v, u, 1, 3) << 1) | !Abc_TtCheckEqualCofs(pTruth, nWords, v, u, 0, 2);
        } else { // if ( v > u )
            Status = (!Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 2, 3) << 1) | !Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 0, 1); }
        assert(Status != 0, "Status != 0");
        if p {
            Dau_DsdInsertVarCache( p, pVars[v], pVars[u], Status ); }
    }
    return Status;
}
func Dau_DsdFindSupports(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32) uint32 {
    var u int32;
    var uSupports uint32;
//Kit_DsdPrintFromTruth( (unsigned *)&tCof0, 6 );printf( "\n" );
//Kit_DsdPrintFromTruth( (unsigned *)&tCof1, 6 );printf( "\n" );
    for  u = 0; u < nVars; u++  {
        if u != v {
            uSupports |= (Dau_DsdFindSupportOne( p, pTruth, pVars, nVars, v, u ) << (2 * u)); }
    }
    return uSupports;
}

// checks decomposability with respect to the pair (v, u)
func Dau_DsdDecomposeDoubleVarsOne(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32, u int32) int32 {
    var pBuffer string // was [10]byte in C
 
    var nWords int32  = Abc_TtWordNum(nVars);
    var Status int32  = Dau_DsdFindSupportOne( p, pTruth, pVars, nVars, v, u );
    assert(v > u, "v > u");
//printf( "Checking %s and %s.\n", p->pVarDefs[pVars[v]], p->pVarDefs[pVars[u]] );
    if Status == 3 {
        // both F(v=0) and F(v=1) depend on u
//        if ( Abc_Tt6Cof0EqualCof1(tCof0, tCof1, u) && Abc_Tt6Cof0EqualCof1(tCof1, tCof0, u) ) // v+u
        if Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 0, 3) && Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 1, 2) { // 00=11 01=10 v+u
        
            var pTtTemp [2][DAU_MAX_WORD]word;
            pBuffer = fmt.Sprintf("[%c%c]", 'a' + pVars[v], 'a' + pVars[u]);
//            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor1(tCof0, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            Abc_TtCofactor0p( pTtTemp[0], pTruth, nWords, v );
            Abc_TtCofactor0( pTtTemp[0], nWords, u );
            Abc_TtCofactor0p( pTtTemp[1], pTruth, nWords, v );
            Abc_TtCofactor1( pTtTemp[1], nWords, u );
            Abc_TtMux( pTruth, Dau_DsdTtElems()[u], pTtTemp[1], pTtTemp[0], nWords );
            goto finish;
        }
    
    } else if Status == 2 {
        // F(v=0) does not depend on u; F(v=1) depends on u
//        if ( Abc_Tt6Cof0EqualCof0(tCof0, tCof1, u) ) // vu
        if Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 0, 2) { // 00=10 vu
        
            var pTtTemp [2][DAU_MAX_WORD]word;
            pBuffer = fmt.Sprintf("(%c%c)", 'a' + pVars[v], 'a' + pVars[u]);
//            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor1(tCof1, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            Abc_TtCofactor0p( pTtTemp[0], pTruth, nWords, v );
            Abc_TtCofactor0( pTtTemp[0], nWords, u );
            Abc_TtCofactor1p( pTtTemp[1], pTruth, nWords, v );
            Abc_TtCofactor1( pTtTemp[1], nWords, u );
            Abc_TtMux( pTruth, Dau_DsdTtElems()[u], pTtTemp[1], pTtTemp[0], nWords );
            goto finish;
        }
//        if ( Abc_Tt6Cof0EqualCof1(tCof0, tCof1, u) ) // v!u
        if Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 0, 3) { // 00=11 v!u
        
            var pTtTemp [2][DAU_MAX_WORD]word;
            pBuffer = fmt.Sprintf("(%c!%c)", 'a' + pVars[v], 'a' + pVars[u]);
//            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor0(tCof1, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            Abc_TtCofactor0p( pTtTemp[0], pTruth, nWords, v );
            Abc_TtCofactor0( pTtTemp[0], nWords, u );
            Abc_TtCofactor1p( pTtTemp[1], pTruth, nWords, v );
            Abc_TtCofactor0( pTtTemp[1], nWords, u );
            Abc_TtMux( pTruth, Dau_DsdTtElems()[u], pTtTemp[1], pTtTemp[0], nWords );
            goto finish;
        }
    
    } else if Status == 1 {
        // F(v=0) depends on u; F(v=1) does not depend on u
//        if ( Abc_Tt6Cof0EqualCof1(tCof0, tCof1, u) ) // !vu
        if Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 0, 3) { // 00=11 !vu
        
            var pTtTemp [2][DAU_MAX_WORD]word;
            pBuffer = fmt.Sprintf("(!%c%c)", 'a' + pVars[v], 'a' + pVars[u]);
//            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor1(tCof0, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u));
            Abc_TtCofactor0p( pTtTemp[0], pTruth, nWords, v );
            Abc_TtCofactor0( pTtTemp[0], nWords, u );
            Abc_TtCofactor0p( pTtTemp[1], pTruth, nWords, v );
            Abc_TtCofactor1( pTtTemp[1], nWords, u );
            Abc_TtMux( pTruth, Dau_DsdTtElems()[u], pTtTemp[1], pTtTemp[0], nWords );
            goto finish;
        }
//        if ( Abc_Tt6Cof1EqualCof1(tCof0, tCof1, u) ) // !v!u
        if Abc_TtCheckEqualCofs(pTruth, nWords, u, v, 1, 3) { // 01=11 !v!u
        
            var pTtTemp [2][DAU_MAX_WORD]word;
            pBuffer = fmt.Sprintf("(!%c!%c)", 'a' + pVars[v], 'a' + pVars[u]);
//            pTruth[0] = (s_Truths6[u] & Abc_Tt6Cofactor0(tCof0, u)) | (~s_Truths6[u] & Abc_Tt6Cofactor1(tCof1, u));
            Abc_TtCofactor1p( pTtTemp[0], pTruth, nWords, v );
            Abc_TtCofactor1( pTtTemp[0], nWords, u );
            Abc_TtCofactor0p( pTtTemp[1], pTruth, nWords, v );
            Abc_TtCofactor0( pTtTemp[1], nWords, u );
            Abc_TtMux( pTruth, Dau_DsdTtElems()[u], pTtTemp[1], pTtTemp[0], nWords );
            goto finish;
        }
    }
    return nVars;
finish: 
    // finalize decomposition
    assert(pBuffer[0], "pBuffer[0]");
    pVars[u] = Dau_DsdAddVarDef( p, pBuffer );
    pVars[v] = pVars[nVars-1];
    Abc_TtSwapVars( pTruth, nVars, v, nVars-1 );
    if Dau_DsdDecomposeSingleVarOne( p, pTruth, pVars, --nVars, u ) {
        nVars = Dau_DsdDecomposeSingleVar( p, pTruth, pVars, --nVars ); }
    return nVars;
}
func Dau_DsdDecomposeDoubleVars(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    var clk abctime  = Abc_Clock();
    for {
        var v, u, nVarsOld int32;
        for  v = nVars - 1; v > 0; v--  {
            for  u = v - 1; u >= 0; u--  {
                if Dau_DsdLookupVarCache( p, pVars[v], pVars[u] ) != 0 {
                    continue; }
                nVarsOld = nVars;
                nVars = Dau_DsdDecomposeDoubleVarsOne( p, pTruth, pVars, nVars, v, u );
                if nVars == 0 {
                    s_Times[1] += Abc_Clock() - clk;
                    return 0;
                }
                if nVarsOld > nVars {
                    break; }
            }
            if u >= 0 { // found
                break; }
        }
        if v == 0 { // not found
            break; }
    }
    s_Times[1] += Abc_Clock() - clk;
    return nVars;
}

// look for MUX-decomposable variable on top or at the bottom
func Dau_DsdDecomposeTripleVarsOuter(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32) int32 {
    extern int Dau_DsdDecomposeInt( Dau_Dsd_t * p, word * pTruth, int nVarsInit );
    var P1 Dau_Dsd_t;
    var p1 *Dau_Dsd_t = &P1;
    var pTtCof [2][DAU_MAX_WORD]word;
    var nWords int32  = Abc_TtWordNum(nVars);
    p1.fSplitPrime = 0;
    p1.fWriteTruth = p.fWriteTruth;
    // move this variable to the top
    // inlining of macro ABC_SWAP
    { var t int32  = pVars[v]; pVars[v] = pVars[nVars-1]; pVars[nVars-1] = t; };
    Abc_TtSwapVars( pTruth, nVars, v, nVars-1 );
    // cofactor w.r.t the last variable
//    tCof0 = Abc_Tt6Cofactor0( pTruth[0], nVars - 1 );
//    tCof1 = Abc_Tt6Cofactor1( pTruth[0], nVars - 1 );
    Abc_TtCofactor0p( pTtCof[0], pTruth, nWords, nVars - 1 );
    Abc_TtCofactor1p( pTtCof[1], pTruth, nWords, nVars - 1 );
    // compose the result
    Dau_DsdWriteString( p, "<" );
    Dau_DsdWriteVar( p, pVars[nVars - 1], 0 );
    // split decomposition
    Dau_DsdDecomposeInt( p1, pTtCof[1], nVars - 1 );
    Dau_DsdTranslate( p, pVars, nVars - 1, p1.pOutput );
    p.nSizeNonDec = p1.nSizeNonDec;
    if p1.nSizeNonDec {
        Abc_TtCopy( pTruth, pTtCof[1], Abc_TtWordNum(p1.nSizeNonDec), 0 ); }
    // split decomposition
    Dau_DsdDecomposeInt( p1, pTtCof[0], nVars - 1 );
    Dau_DsdTranslate( p, pVars, nVars - 1, p1.pOutput );
    Dau_DsdWriteString( p, ">" );
    p.nSizeNonDec = Abc_MaxInt( p.nSizeNonDec, p1.nSizeNonDec );
    if p1.nSizeNonDec {
        Abc_TtCopy( pTruth, pTtCof[0], Abc_TtWordNum(p1.nSizeNonDec), 0 ); }
    return 0;
}
func Dau_DsdDecomposeTripleVarsInner(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32, v int32, uSupports uint32) int32 {
    var nWords int32  = Abc_TtWordNum(nVars);
    var iVar0 int32  = Abc_TtSuppFindFirst(  uSupports & (~uSupports >> 1) & 0x55555555 ) >> 1;
    var iVar1 int32  = Abc_TtSuppFindFirst( ~uSupports & ( uSupports >> 1) & 0x55555555 ) >> 1;
    var fEqual0, fEqual1 int32;
//    word tCof0 = Abc_Tt6Cofactor0( pTruth[0], v );
//    word tCof1 = Abc_Tt6Cofactor1( pTruth[0], v );
//    word C00 = Abc_Tt6Cofactor0( tCof0, iVar0 );
//    word C01 = Abc_Tt6Cofactor1( tCof0, iVar0 );
//    word C10 = Abc_Tt6Cofactor0( tCof1, iVar1 );
//    word C11 = Abc_Tt6Cofactor1( tCof1, iVar1 );
//    int fEqual0 = (C00 == C10) && (C01 == C11);
//    int fEqual1 = (C00 == C11) && (C01 == C10);
    var pTtCof [2][DAU_MAX_WORD]word;
    var pTtFour [2][2][DAU_MAX_WORD]word;
    Abc_TtCofactor0p( pTtCof[0], pTruth, nWords, v );
    Abc_TtCofactor1p( pTtCof[1], pTruth, nWords, v );
    Abc_TtCofactor0p( pTtFour[0][0], pTtCof[0], nWords, iVar0 );
    Abc_TtCofactor1p( pTtFour[0][1], pTtCof[0], nWords, iVar0 );
    Abc_TtCofactor0p( pTtFour[1][0], pTtCof[1], nWords, iVar1 );
    Abc_TtCofactor1p( pTtFour[1][1], pTtCof[1], nWords, iVar1 );
    fEqual0 = Abc_TtEqual(pTtFour[0][0], pTtFour[1][0], nWords) && Abc_TtEqual(pTtFour[0][1], pTtFour[1][1], nWords);
    fEqual1 = Abc_TtEqual(pTtFour[0][0], pTtFour[1][1], nWords) && Abc_TtEqual(pTtFour[0][1], pTtFour[1][0], nWords);
    if (fEqual0 != 0) || (fEqual1 != 0) {
        var pBuffer string // was [10]byte in C

        var VarId int32  = pVars[iVar0];
//        pTruth[0] = (s_Truths6[v] & C11) | (~s_Truths6[v] & C10);
        Abc_TtMux( pTruth, Dau_DsdTtElems()[v], pTtFour[1][1], pTtFour[1][0], nWords );
        pBuffer = fmt.Sprintf("<%c%c%s%c>", 'a' + pVars[v], 'a' + pVars[iVar1], fEqual1 ? "!":"", 'a' + pVars[iVar0]);
        pVars[v] = Dau_DsdAddVarDef( p, pBuffer );
        // remove iVar1
        // inlining of macro ABC_SWAP
        { var t int32  = pVars[iVar1]; pVars[iVar1] = pVars[nVars-1]; pVars[nVars-1] = t; };
        Abc_TtSwapVars( pTruth, nVars, iVar1, nVars-1 ); nVars--;
        // remove iVar0
        iVar0 = Dau_DsdFindVarDef( pVars, nVars, VarId );
        // inlining of macro ABC_SWAP
        { var t int32  = pVars[iVar0]; pVars[iVar0] = pVars[nVars-1]; pVars[nVars-1] = t; };
        Abc_TtSwapVars( pTruth, nVars, iVar0, nVars-1 ); nVars--;
        // find the new var
        v = Dau_DsdFindVarDef( pVars, nVars, p.nVarsUsed-1 );
        // remove single variables if possible
        if Dau_DsdDecomposeSingleVarOne( p, pTruth, pVars, nVars, v ) != 0 {
            nVars = Dau_DsdDecomposeSingleVar( p, pTruth, pVars, --nVars ); }
        return nVars;
    }
    return nVars;
}
func Dau_DsdDecomposeTripleVars(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    var clk abctime  = Abc_Clock();
    for {
        var v int32;
//        Kit_DsdPrintFromTruth( (unsigned *)pTruth, 6 ); printf( "\n" );
        for  v = nVars - 1; v >= 0; v--  {
            var uSupports uint32  = Dau_DsdFindSupports( p, pTruth, pVars, nVars, v );
//            Dau_DsdPrintSupports( uSupports, nVars );
            if (uSupports & (uSupports >> 1) & 0x55555555) == 0 { // non-overlapping supports
                return Dau_DsdDecomposeTripleVarsOuter( p, pTruth, pVars, nVars, v ); }
            if Abc_TtSuppOnlyOne( uSupports & (~uSupports >> 1) & 0x55555555) &&
                 Abc_TtSuppOnlyOne(~uSupports & ( uSupports >> 1) & 0x55555555) { // one unique variable in each cofactor
            
                var nVarsNew int32  = Dau_DsdDecomposeTripleVarsInner( p, pTruth, pVars, nVars, v, uSupports );
                if nVarsNew == nVars {
                    continue; }
                if nVarsNew == 0 {
                    s_Times[2] += Abc_Clock() - clk;
                    return 0;
                }
                nVars = Dau_DsdDecomposeDoubleVars( p, pTruth, pVars, nVarsNew );
                if nVars == 0 {
                    s_Times[2] += Abc_Clock() - clk;
                    return 0;
                }
                break;
            }
        }
        if v == -1 {
            s_Times[2] += Abc_Clock() - clk;
            return nVars;
        }
    }
    assert(false, "0");
    return -1;
}
func Dau_DsdDecomposeInternal(p *Dau_Dsd_t, pTruth *word, pVars *int32, nVars int32) int32 {
    // decompose single variales on the output side
    nVars = Dau_DsdDecomposeSingleVar( p, pTruth, pVars, nVars );
    if nVars == 0 {
        return 0; }
    // decompose double variables on the input side
    nVars = Dau_DsdDecomposeDoubleVars( p, pTruth, pVars, nVars );
    if nVars == 0 {
        return 0; }
    // decompose MUX on the output/input side
    nVars = Dau_DsdDecomposeTripleVars( p, pTruth, pVars, nVars );
    if nVars == 0 {
        return 0; }
    // write non-decomposable function
    return Dau_DsdWritePrime( p, pTruth, pVars, nVars );
}

/**Function*************************************************************

  Synopsis    [Fast DSD for truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdMinBase(pTruth *word, nVars int32, pVarsNew *int32) int32 {
    var v int32;
    for  v = 0; v < nVars; v++  {
        pVarsNew[v] = v;
    }
    for  v = nVars - 1; v >= 0; v--  {
        if Abc_TtHasVar( pTruth, nVars, v ) {
            continue; }
        Abc_TtSwapVars( pTruth, nVars, v, nVars-1 );
        pVarsNew[v] = pVarsNew[--nVars];
    }
    return nVars;
}
func Dau_DsdDecomposeInt(p *Dau_Dsd_t, pTruth *word, nVarsInit int32) int32 {
    var Status int32;
    var nVars int32;
    var pVars [16]int32;
    Dau_DsdInitialize( p, nVarsInit );
    nVars = Dau_DsdMinBase( pTruth, nVarsInit, pVars );
    assert(nVars > 0 && nVars <= nVarsInit, "nVars > 0 && nVars <= nVarsInit");
    if nVars == 1 {
        Dau_DsdWriteVar( p, pVars[0], int32((pTruth[0] & 1)) );
    } else if nVars <= 6 {
        Status = Dau_Dsd6DecomposeInternal( p, pTruth, pVars, nVars );
    } else {
        Status = Dau_DsdDecomposeInternal( p, pTruth, pVars, nVars ); }
    Dau_DsdFinalize( p );
    return Status;
}
func Dau_DsdDecompose(pTruth *word, nVarsInit int32, fSplitPrime int32, fWriteTruth int32, pRes string) int32 {
    var P Dau_Dsd_t;
    var p *Dau_Dsd_t = &P;
    p.fSplitPrime = fSplitPrime;
    p.fWriteTruth = fWriteTruth;
    p.pVarLevels  = nil;
    p.nSizeNonDec = 0;
    if (pTruth[0] & 1) == 0 && Abc_TtIsConst0(pTruth, Abc_TtWordNum(nVarsInit)) { if pRes { pRes[0] = '0', pRes[1] = 0; } 
    } else if (pTruth[0] & 1) && Abc_TtIsConst1(pTruth, Abc_TtWordNum(nVarsInit)) { if pRes { pRes[0] = '1', pRes[1] = 0; } 
    } else {
        var Status int32  = Dau_DsdDecomposeInt( p, pTruth, nVarsInit );
        Dau_DsdRemoveBraces( p.pOutput, Dau_DsdComputeMatches(p.pOutput) );
        if pRes {
            strcpy( pRes, p.pOutput ); }
        assert((fSplitPrime != 0) || Status != 1, "fSplitPrime || Status != 1");
        if (fSplitPrime != 0) && Status == 2 {
            return -1; }
    }
//    assert( p->nSizeNonDec == 0 );
    return p.nSizeNonDec;
}
func Dau_DsdDecomposeLevel(pTruth *word, nVarsInit int32, fSplitPrime int32, fWriteTruth int32, pRes string, pVarLevels *int32) int32 {
    var P Dau_Dsd_t;
    var p *Dau_Dsd_t = &P;
    p.fSplitPrime = fSplitPrime;
    p.fWriteTruth = fWriteTruth;
    p.pVarLevels  = pVarLevels;
    p.nSizeNonDec = 0;
    if (pTruth[0] & 1) == 0 && Abc_TtIsConst0(pTruth, Abc_TtWordNum(nVarsInit)) { if pRes { pRes[0] = '0', pRes[1] = 0; } 
    } else if (pTruth[0] & 1) && Abc_TtIsConst1(pTruth, Abc_TtWordNum(nVarsInit)) { if pRes { pRes[0] = '1', pRes[1] = 0; } 
    } else {
        var Status int32  = Dau_DsdDecomposeInt( p, pTruth, nVarsInit );
        Dau_DsdRemoveBraces( p.pOutput, Dau_DsdComputeMatches(p.pOutput) );
        if pRes {
            strcpy( pRes, p.pOutput ); }
        assert((fSplitPrime != 0) || Status != 1, "fSplitPrime || Status != 1");
        if (fSplitPrime != 0) && Status == 2 {
            return -1; }
    }
//    assert( p->nSizeNonDec == 0 );
    return p.nSizeNonDec;
}
func Dau_DsdPrintFromTruthFile(pFile *os.File, pTruth *word, nVarsInit int32) {
    var pRes [DAU_MAX_STR]byte;
    var pTemp [DAU_MAX_WORD]word;
    Abc_TtCopy( pTemp, pTruth, Abc_TtWordNum(nVarsInit), 0 );
    Dau_DsdDecompose( pTemp, nVarsInit, 0, 1, pRes );
    fmt.Fprintf(pFile, "%s\n", pRes);
}
func Dau_DsdPrintFromTruth(pTruth *word, nVarsInit int32) {
    var pRes [DAU_MAX_STR]byte;
    var pTemp [DAU_MAX_WORD]word;
    Abc_TtCopy( pTemp, pTruth, Abc_TtWordNum(nVarsInit), 0 );
    Dau_DsdDecompose( pTemp, nVarsInit, 0, 1, pRes );
    fmt.Fprintf(os.Stdout, "%s\n", pRes);
}
func Dau_DsdPrintFromTruth2(pTruth *word, nVarsInit int32) {
    var pRes [DAU_MAX_STR]byte;
    var pTemp [DAU_MAX_WORD]word;
    Abc_TtCopy( pTemp, pTruth, Abc_TtWordNum(nVarsInit), 0 );
    Dau_DsdDecompose( pTemp, nVarsInit, 0, 1, pRes );
    fmt.Fprintf(os.Stdout, "%s", pRes);
}

func Dau_DsdTest44() {
    var pRes [DAU_MAX_STR]byte;
//    char * pStr = "(!(!a<bcd>)!(!fe))";
//    char * pStr = "([acb]<!edf>)";
//    char * pStr = "!(f!(b!c!(d[ea])))";
//    char * pStr = "[!(a[be])!(c!df)]";
//    char * pStr = "<(e<bca>)fd>";
//    char * pStr = "[d8001{abef}c]";
    var pStr string  = "[dc<a(cbd)(!c!b!d)>{abef}]";
//    char * pStr3;
    var t word  = Dau_Dsd6ToTruth( pStr );
//    return;
    var nNonDec int32  = Dau_DsdDecompose( &t, 6, 1, 1, pRes );
//    Dau_DsdNormalize( pStr2 );
//    Dau_DsdExtract( pStr, 2, 0 );
    t = 0; 
    nNonDec = 0;
}



func Dau_DsdTest888() {
    var pDsd [DAU_MAX_STR]byte;
    var nVars int32  = 9;
//    char * pStr = "[(abc)(def)(ghi)]";
//    char * pStr = "[a!b!(c!d[e(fg)hi])]";
//    char * pStr = "[(abc)(def)]";
//    char * pStr = "[(abc)(def)]";
//    char * pStr = "[abcdefg]";
//    char * pStr = "[<abc>(de[ghi])]";
    var pStr string  = "(<abc>(<def><ghi>))";
    var pTruth *word  = Dau_DsdToTruth( pStr, 9 );
    var i, Status int32;
//    Kit_DsdPrintFromTruth( (unsigned *)pTruth, 9 ); printf( "\n" );
/*
    for ( i = 0; i < 6; i++ )
    {
        unsigned uSupp = Dau_Dsd6FindSupports( NULL, pTruth, NULL, 6, i );
        Dau_DsdPrintSupports( uSupp, 6 );
    }
*/
/*
    printf( "\n" );
    for ( i = 0; i < nVars; i++ )
    {
        unsigned uSupp = Dau_DsdFindSupports( NULL, pTruth, NULL, nVars, i );
        Dau_DsdPrintSupports( uSupp, nVars );
    }
*/
    Status = Dau_DsdDecompose( pTruth, nVars, 0, 0, pDsd );
    i = 0;
}

func Dau_DsdTest555() {
    var nVars int32  = 10;
    var nWords int32  = Abc_TtWordNum(nVars);
    var pFileName string  = "_npn/npn/dsd10.txt";
    var pFile *os.File  = os.Open(pFileName);
    var Tru [2][DAU_MAX_WORD]word;
    var pTruth *word;
    var pBuffer [DAU_MAX_STR]byte;
    var pRes [DAU_MAX_STR]byte;
    var nSizeNonDec int32;
    var i, Counter int32;
    var clk abctime = Abc_Clock();
    var clkDec abctime;
    var clk2 abctime;
//    return;

    for  fgets( pBuffer, DAU_MAX_STR, pFile ) != nil  {
        var pStr2 string  = pBuffer + len(pBuffer)-1;
        if *pStr2 == '\n' {
            *pStr2-- = 0; }
        if *pStr2 == '\r' {
            *pStr2-- = 0; }
        if pBuffer[0] == 'V' || pBuffer[0] == 0 {
            continue; }
        Counter++; 

        for  i = 0; i < 1; i++  {
//            Dau_DsdPermute( pBuffer );
            pTruth = Dau_DsdToTruth( pBuffer[0] == '*' ? pBuffer + 1 : pBuffer, nVars );
            Abc_TtCopy( Tru[0], pTruth, nWords, 0 );
            Abc_TtCopy( Tru[1], pTruth, nWords, 0 );
            clk2 = Abc_Clock();
            nSizeNonDec = Dau_DsdDecompose( Tru[1], nVars, 0, 1, pRes );
            clkDec += Abc_Clock() - clk2;
            Dau_DsdNormalize( pRes );
//            pStr2 = Dau_DsdPerform( t ); nSizeNonDec = 0;
            assert(nSizeNonDec == 0, "nSizeNonDec == 0");
            pTruth = Dau_DsdToTruth( pRes, nVars );
            if !Abc_TtEqual( pTruth, Tru[0], nWords ) {
    //        Kit_DsdPrintFromTruth( (unsigned *)&t, 6 );
    //        printf( "  " );
    //        Kit_DsdPrintFromTruth( (unsigned *)&t2, 6 );
                fmt.Printf("%s -> %s \n", pBuffer, pRes);
                fmt.Printf("Verification failed.\n");
            }
        }
    }
    fmt.Printf("Finished trying %d decompositions.  ", Counter);
    Abc_PrintTime( 1, "Time", clkDec );
    Abc_PrintTime( 1, "Total", Abc_Clock() - clk );

    Abc_PrintTime( 1, "Time1", s_Times[0] );
    Abc_PrintTime( 1, "Time2", s_Times[1] );
    Abc_PrintTime( 1, "Time3", s_Times[2] );

    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


