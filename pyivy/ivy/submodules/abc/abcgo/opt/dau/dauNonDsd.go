package dau

import "fmt"

/**CFile****************************************************************

  FileName    [dauNonDsd.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauNonDsd.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Checks decomposability with given variable set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DecCheckSetTop5(p *word, nVars int32, nVarsF int32, nVarsB int32, nVarsS int32, uMaskS int32, pSched *int32, pDec *word, pComp *word) int32 {
    var Cof [2][64]word;
    var Value word;
    var MaskFF word  = (((word)1) << (1 << nVarsF)) - 1;
    var ShiftF int32 = 6 - nVarsF;
    var MaskF int32 = (1 << ShiftF) - 1;
    var pVarsS, pVarsB [16]int32;
    var nMints int32   = (1 << nVarsB);
    var nMintsB int32  = (1 <<(nVarsB-nVarsS));
    var nMintsS int32  = (1 << nVarsS);
    var s, b, v, m, Mint, MintB, MintS int32;
    assert(nVars == nVarsB + nVarsF, "nVars == nVarsB + nVarsF");
    assert(nVars <= 16, "nVars <= 16");
    assert(nVarsS <= 6, "nVarsS <= 6");
    assert(nVarsF >= 1 && nVarsF <= 5, "nVarsF >= 1 && nVarsF <= 5");
    // collect bound/shared variables
    for  s = b = v = 0; v < nVarsB; v++  {
        if (uMaskS >> v) & 1 {
            pVarsB[v] = -1, pVarsS[v] = s++;
        } else {
            pVarsS[v] = -1, pVarsB[v] = b++; }
    }
    assert(s == nVarsS, "s == nVarsS");
    assert(b == nVarsB-nVarsS, "b == nVarsB-nVarsS");
    // clean minterm storage
    for  s = 0; s < nMintsS; s++  {
        Cof[0][s] = Cof[1][s] = ~(word)0;
    }
    // iterate through bound set minters
    for  MintS = MintB = Mint = m = 0; m < nMints; m++  {
        // find minterm value
        Value = (p[Mint>>ShiftF] >> ((Mint&MaskF)<<nVarsF)) & MaskFF;
        // check if this cof already appeared 
        if !~Cof[0][MintS] || Cof[0][MintS] == Value {
            Cof[0][MintS] = Value;
        } else if !~Cof[1][MintS] || Cof[1][MintS] == Value {
            Cof[1][MintS] = Value;
            if pDec != nil {
                var iMintB int32  = MintS * nMintsB + MintB;
                pDec[iMintB>>6] |= (((word)1)<<(iMintB & 63));     
            }
        
        } else {
            return 0; }            
        // find next minterm
        v = pSched[m];
        Mint ^= (1 << v);
        if (uMaskS >> v) & 1 { // shared variable
            MintS ^= (1 << pVarsS[v]);
        } else {
            MintB ^= (1 << pVarsB[v]); }
    }
    // create composition function
    if pComp != nil {
        for  s = 0; s < nMintsS; s++  {
            pComp[s>>ShiftF] |= (Cof[0][s] << ((s&MaskF) << nVarsF));
            if ~Cof[1][s] { 
                pComp[(s+nMintsS)>>ShiftF] |= (Cof[1][s] << (((s+nMintsS)&MaskF) << nVarsF));
            } else {
                pComp[(s+nMintsS)>>ShiftF] |= (Cof[0][s] << (((s+nMintsS)&MaskF) << nVarsF)); }
        }
        if nVarsF + nVarsS + 1 < 6 {
            pComp[0] = Abc_Tt6Stretch( pComp[0], nVarsF + nVarsS + 1 ); }
    }
    if (pDec != nil) && nVarsB < 6 {
        pDec[0] = Abc_Tt6Stretch( pDec[0], nVarsB ); }
    return 1;
}
func Dau_DecCheckSetTop6(p *word, nVars int32, nVarsF int32, nVarsB int32, nVarsS int32, uMaskS int32, pSched *int32, pDec *word, pComp *word) int32 {
    var Cof [2][64]*word;
    var nWordsF int32  = Abc_TtWordNum(nVarsF); 
    var pVarsS, pVarsB [16]int32;
    var nMints int32   = (1 << nVarsB);
    var nMintsB int32  = (1 <<(nVarsB-nVarsS));
    var nMintsS int32  = (1 << nVarsS);
    var s, b, v, m, Mint, MintB, MintS int32;
    assert(nVars == nVarsB + nVarsF, "nVars == nVarsB + nVarsF");
    assert(nVars <= 16, "nVars <= 16");
    assert(nVarsS <= 6, "nVarsS <= 6");
    assert(nVarsF >= 6, "nVarsF >= 6");
    // collect bound/shared variables
    for  s = b = v = 0; v < nVarsB; v++  {
        if (uMaskS >> v) & 1 {
            pVarsB[v] = -1, pVarsS[v] = s++;
        } else {
            pVarsS[v] = -1, pVarsB[v] = b++; }
    }
    assert(s == nVarsS, "s == nVarsS");
    assert(b == nVarsB-nVarsS, "b == nVarsB-nVarsS");
    // clean minterm storage
    for  s = 0; s < nMintsS; s++  {
        Cof[0][s] = Cof[1][s] = nil;
    }
    // iterate through bound set minters
    for  MintS = MintB = Mint = m = 0; m < nMints; m++  {
        // check if this cof already appeared 
        if !Cof[0][MintS] || !memcmp(Cof[0][MintS], p + Mint * nWordsF, sizeof(word) * nWordsF) {
            Cof[0][MintS] = p + Mint * nWordsF;
        } else if !Cof[1][MintS] || !memcmp(Cof[1][MintS], p + Mint * nWordsF, sizeof(word) * nWordsF) {
            Cof[1][MintS] = p + Mint * nWordsF;
            if pDec != nil {
                var iMintB int32  = MintS * nMintsB + MintB;
                pDec[iMintB>>6] |= (((word)1)<<(iMintB & 63));     
            }
        
        } else {
            return 0; }            
        // find next minterm
        v = pSched[m];
        Mint ^= (1 << v);
        if (uMaskS >> v) & 1 { // shared variable
            MintS ^= (1 << pVarsS[v]);
        } else {
            MintB ^= (1 << pVarsB[v]); }
    }
    // create composition function
    if pComp != nil {
        for  s = 0; s < nMintsS; s++  {
            memcpy( pComp + s * nWordsF, Cof[0][s], sizeof(word) * nWordsF );
            if Cof[1][s] { 
                memcpy( pComp + (s+nMintsS) * nWordsF, Cof[1][s], sizeof(word) * nWordsF );
            } else {
                memcpy( pComp + (s+nMintsS) * nWordsF, Cof[0][s], sizeof(word) * nWordsF ); }
        }
    }
    if (pDec != nil) && nVarsB < 6 {
        pDec[0] = Abc_Tt6Stretch( pDec[0], nVarsB ); }
    return 1;
}
func Dau_DecCheckSetTop(p *word, nVars int32, nVarsF int32, nVarsB int32, nVarsS int32, uMaskS int32, pSched *int32, pDec *word, pComp *word) int32 {
    if nVarsF < 6 {
        return Dau_DecCheckSetTop5( p, nVars, nVarsF, nVarsB, nVarsS, uMaskS, pSched, pDec, pComp );
    } else {
        return Dau_DecCheckSetTop6( p, nVars, nVarsF, nVarsB, nVarsS, uMaskS, pSched, pDec, pComp ); }
}

/**Function*************************************************************

  Synopsis    [Checks decomposability with given BS variables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DecGetMinterm(p *word, g int32, nVarsS int32, uMaskAll int32) {
    var m, c, v int32;
    for  m = c = v = 0; v < nVarsS; v++  {
        if !((uMaskAll >> v) & 1) { // not shared bound set variable
        
            if (g >> v) & 1 {
                m |= (1 << c); }
            c++;
        }
    }
    assert(c >= 2, "c >= 2");
    p[m>>6] |= (((word)1)<<(m & 63));                               
}
func Dau_DecCheckSet5(p *word, nVars int32, nVarsF int32, uMaskAll int32, uMaskValue int32, pCof0 *word, pCof1 *word, pDec *word) int32 {
    var fFound0, fFound1 int32;
    var g int32;
    var gMax int32 = (1 << (nVars - nVarsF));
    var Shift int32 = 6 - nVarsF;
    var Mask int32 = (1 << Shift) - 1;
    var Mask2 word  = (((word)1) << (1 << nVarsF)) - 1;
    var Cof0, Cof1, Value word;
    assert(nVarsF >= 1 && nVarsF <= 5, "nVarsF >= 1 && nVarsF <= 5");
    if pDec != nil { *pDec = 0; }
    for  g = 0; g < gMax; g++  {
        if (g & uMaskAll) == uMaskValue { // this minterm g matches shared variable minterm uMaskValue
        
            Value = (p[g>>Shift] >> ((g&Mask)<<nVarsF)) & Mask2;
            if !fFound0 {
                Cof0 = Value;
fFound0 = 1;
            } else if Cof0 == Value {
                continue;
            } else if !(fFound1 != 0) {
                Cof1 = Value;
fFound1 = 1;
                if pDec != nil { Dau_DecGetMinterm( pDec, g, nVars-nVarsF, uMaskAll ); }
            
            } else if Cof1 == Value {
                if pDec != nil { Dau_DecGetMinterm( pDec, g, nVars-nVarsF, uMaskAll ); }
                continue;
            
            } else {
                return 0; }            
        }
    }
    if pCof0 != nil {
        assert(fFound0, "fFound0");
        if fFound1 != 0 {
            Cof1 = Cof1;
        } else {
            Cof1 = Cof0;
        }
        *pCof0 = Abc_Tt6Stretch( Cof0, nVarsF );
        *pCof1 = Abc_Tt6Stretch( Cof1, nVarsF );
    }
    return 1;
}
func Dau_DecCheckSet6(p *word, nVars int32, nVarsF int32, uMaskAll int32, uMaskValue int32, pCof0 *word, pCof1 *word, pDec *word) int32 {
    var fFound0, fFound1 int32;
    var g int32;
    var gMax int32 = (1 << (nVars - nVarsF));
    var nWords int32  = Abc_TtWordNum(nVarsF);
    var Cof0 *word = nil;
    var Cof1 *word = nil;
    assert(nVarsF >= 6 && nVarsF <= nVars - 2, "nVarsF >= 6 && nVarsF <= nVars - 2");
    if pDec != nil { *pDec = 0; }
    for  g = 0; g < gMax; g++  {
        if (g & uMaskAll) == uMaskValue {
            if !fFound0 {
                Cof0 = p + g * nWords;
fFound0 = 1;
            } else if !memcmp(Cof0, p + g * nWords, sizeof(word) * nWords) {
                continue;
            } else if !(fFound1 != 0) {
                Cof1 = p + g * nWords;
fFound1 = 1;
                if pDec != nil { Dau_DecGetMinterm( pDec, g, nVars-nVarsF, uMaskAll ); }
            
            } else if !memcmp(Cof1, p + g * nWords, sizeof(word) * nWords) {
                if pDec != nil { Dau_DecGetMinterm( pDec, g, nVars-nVarsF, uMaskAll ); }
                continue;
            
            } else {
                return 0; }            
        }
    }
    if pCof0 != nil {
        assert(fFound0, "fFound0");
        if fFound1 != 0 {
            Cof1 = Cof1;
        } else {
            Cof1 = Cof0;
        }
        memcpy( pCof0, Cof0, sizeof(word) * nWords );
        memcpy( pCof1, Cof1, sizeof(word) * nWords );
    }
    return 1;
}
func Dau_DecCheckSetAny(p *word, nVars int32, nVarsF int32, uMaskAll int32, uMaskValue int32, pCof0 *word, pCof1 *word, pDec *word) int32 {
    assert(nVarsF >= 1 && nVarsF <= nVars - 2, "nVarsF >= 1 && nVarsF <= nVars - 2");
    if nVarsF < 6 {
        return Dau_DecCheckSet5( p, nVars, nVarsF, uMaskAll, uMaskValue, pCof0, pCof1, pDec );
    } else {
        return Dau_DecCheckSet6( p, nVars, nVarsF, uMaskAll, uMaskValue, pCof0, pCof1, pDec ); }
}
func Dau_DecCheckSetTopOld(p *word, nVars int32, nVarsF int32, nVarsB int32, nVarsS int32, maskS int32, pCof0 **word, pCof1 **word, pDec **word) int32 {
    var i int32;
    var pVarsS [16]int32;
    var v int32;
    var m int32;
    var mMax int32 = (1 << nVarsS);
    var uMaskValue int32;
    assert(nVars >= 3 && nVars <= 16, "nVars >= 3 && nVars <= 16");
    assert(nVars == nVarsF + nVarsB, "nVars == nVarsF + nVarsB");
    assert(nVarsF >= 1 && nVarsF <= nVars - 2, "nVarsF >= 1 && nVarsF <= nVars - 2");
    assert(nVarsB >= 2 && nVarsB <= nVars - 1, "nVarsB >= 2 && nVarsB <= nVars - 1");
    assert(nVarsS >= 0 && nVarsS <= nVarsB - 2, "nVarsS >= 0 && nVarsS <= nVarsB - 2");
    if nVarsS == 0 {
        return Dau_DecCheckSetAny( p, nVars, nVarsF, 0, 0, pCof0? pCof0[0] : 0, pCof1? pCof1[0] : 0, pDec? pDec[0] : 0 ); }
    // collect shared variables
    assert(maskS > 0 && maskS < (1 << nVarsB), "maskS > 0 && maskS < (1 << nVarsB)");
    for  i, v = 0, 0; v < nVarsB; v++  {
        if (maskS >> v) & 1 {
            pVarsS[i++] = v; }
    }
    assert(i == nVarsS, "i == nVarsS");
    // go through shared set minterms
    for  m = 0; m < mMax; m++  {
        // generate share set mask
        uMaskValue = 0;
        for  v = 0; v < nVarsS; v++  {
            if (m >> v) & 1 {
                uMaskValue |= (1 << pVarsS[v]); }
        }
        assert((maskS & uMaskValue) == uMaskValue, "(maskS & uMaskValue) == uMaskValue");
        // check decomposition
        if !Dau_DecCheckSetAny( p, nVars, nVarsF, maskS, uMaskValue, pCof0? pCof0[m] : 0, pCof1? pCof1[m] : 0, pDec? pDec[m] : 0 ) {
            return 0; }
    }
   return 1;
}
 

/**Function*************************************************************

  Synopsis    [Variable sets.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DecCreateSet(pVarsB *int32, sizeB int32, maskS int32) uint32 {
    var uSet uint32; var v int32;
    for  v = 0; v < sizeB; v++  {
        uSet |= (1 <<  (pVarsB[v] << 1));
        if (maskS >> v) & 1 {
        uSet |= (1 << ((pVarsB[v] << 1)+1)); }
    }
    return uSet;
}
func Dau_DecSetHas01(Mask uint32) int32 {
    return (Mask & ((~Mask) >> 1) & 0x55555555);
}
static inline int Dau_DecSetIsContained( Vec_Int_t * vSets, unsigned New )
// Old=abcD contains New=abcDE
// Old=abcD contains New=abCD
{
    var Old uint32;
    var i, Entry int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSets)) && (((Entry) = Vec_IntEntry(vSets, i)), 1); i++  {
        Old = uint32(Entry);
        if (Old & ~New) == 0 && !(Dau_DecSetHas01(~Old & New) != 0) {
            return 1; }
    }
    return 0;
}
func Dau_DecSortSet(set uint32, nVars int32, pnUnique *int32, pnShared *int32, pnFree *int32) {
    var v int32;
    var nUnique, nShared, nFree int32;
    for  v = 0; v < nVars; v++  {
        var Value int32  = ((set >> (v << 1)) & 3);
        if Value == 1 {
            nUnique++;
        } else if Value == 3 {
            nShared++;
        } else if Value == 0 {
            nFree++;
        } else { assert(false, "0"); }
    }
    *pnUnique = nUnique;
    *pnShared = nShared;
    *pnFree   = nFree;
}
func Dau_DecPrintSet(set uint32, nVars int32, fNewLine int32) {
    var v, Counter int32;
    var nUnique, nShared, nFree int32;
    Dau_DecSortSet( set, nVars, &nUnique, &nShared, &nFree );
    fmt.Printf("S =%2d  D =%2d  C =%2d   ", nShared, nUnique+nShared, nShared+nFree+1);

    fmt.Printf("x=");
    for  v = 0; v < nVars; v++  {
        var Value int32  = ((set >> (v << 1)) & 3);
        if Value == 1 {
            fmt.Printf("%c", 'a' + v);
Counter++;
        } else if Value == 3 {
            fmt.Printf("%c", 'A' + v);
Counter++;
        } else { assert(Value == 0, "Value == 0"); }
    }
    fmt.Printf(" y=x");
    for  v = 0; v < nVars; v++  {
        var Value int32  = ((set >> (v << 1)) & 3);
        if Value == 0 {
            fmt.Printf("%c", 'a' + v);
Counter++;
        } else if Value == 3 {
            fmt.Printf("%c", 'A' + v);
Counter++; }
    }
    for  ; Counter < 15; Counter++  {
        fmt.Printf(" ");
    }
    if fNewLine != 0 {
        fmt.Printf("\n"); }
}
func Dau_DecReadSet(pStr string) uint32 {
    var uSet uint32; var v int32;
    for  v = 0; pStr[v]; v++  {
        if pStr[v] >= 'a' && pStr[v] <= 'z' {
            uSet |= (1 << ((pStr[v] - 'a') << 1));
        } else if pStr[v] >= 'A' && pStr[v] <= 'Z' {
            uSet |= (1 << ((pStr[v] - 'a') << 1)) | (1 << (((pStr[v] - 'a') << 1)+1));
        } else { break; }
    }
    return uSet;
}
func Dau_DecPrintSets(vSets *Vec_Int_t, nVars int32) {
    var i, Entry int32;
    fmt.Printf("The %d-variable set family contains %d sets:\n", nVars, Vec_IntSize(vSets));
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSets)) && (((Entry) = Vec_IntEntry(vSets, i)), 1); i++  {
        Dau_DecPrintSet( uint32(Entry), nVars, 1 );
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Find decomposable bound-sets of the given function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DecMoveFreeToLSB(p *word, nVars int32, V2P *int32, P2V *int32, maskB int32, sizeB int32) {
    var v, c int32;
    for  v = 0; v < nVars; v++  {
        if !((maskB >> v) & 1) {
            Abc_TtMoveVar( p, nVars, V2P, P2V, v, c++ ); }
    }
    assert(c == nVars - sizeB, "c == nVars - sizeB");
}
func Dau_DecFindSets_int(pInit *word, nVars int32, pSched []*int32) *Vec_Int_t {
    var vSets *Vec_Int_t  = Vec_IntAlloc( 32 );
    var V2P, P2V, pVarsB [16]int32;
    var Limit int32  = (1 << nVars);
    var c, v, sizeB, sizeS, maskB, maskS int32;
    var setMixed uint32;
    var p [1<<10]word; 
    memcpy( p, pInit, sizeof(word) * Abc_TtWordNum(nVars) );
    for  v = 0; v < nVars; v++  {
        assert(Abc_TtHasVar( p, nVars, v ), "Abc_TtHasVar( p, nVars, v )");
    }
    // initialize permutation
    for  v = 0; v < nVars; v++  {
        V2P[v] = P2V[v] = v;
    }
    // iterate through bound sets of each size in increasing order
    for  sizeB = 2; sizeB < nVars; sizeB++  { // bound set size
    for  maskB = 0; maskB < Limit; maskB++  { // bound set
    if Abc_TtBitCount16(maskB) == sizeB {
        // permute variables to have bound set on top
        Dau_DecMoveFreeToLSB( p, nVars, V2P, P2V, maskB, sizeB );
        // collect bound set vars on levels nVars-sizeB to nVars-1
        for  c = 0; c < sizeB; c++  {
            pVarsB[c] = P2V[nVars-sizeB+c];
        }
        // check disjoint
//        if ( Dau_DecCheckSetTopOld(p, nVars, nVars-sizeB, sizeB, 0, 0, NULL, NULL, NULL) )
        if Dau_DecCheckSetTop(p, nVars, nVars-sizeB, sizeB, 0, 0, pSched[sizeB], nil, nil) != 0 {
            Vec_IntPush( vSets, Dau_DecCreateSet(pVarsB, sizeB, 0) );
            continue;
        }
        if sizeB == 2 {
            continue; }
        // iterate through shared sets of each size in the increasing order
        for  sizeS = 1; sizeS <= sizeB - 2; sizeS++  {   // shared set size
        if sizeS <= 3 {
//        sizeS = 1; 
        for  maskS = 0; maskS < (1 << sizeB); maskS++  { // shared set
        if Abc_TtBitCount16(maskS) == sizeS {
            setMixed = Dau_DecCreateSet( pVarsB, sizeB, maskS );
//            printf( "Considering %10d ", setMixed );
//            Dau_DecPrintSet( setMixed, nVars );
            // check if it exists
            if Dau_DecSetIsContained(vSets, setMixed) {
                continue; }
            // check if it can be added
//            if ( Dau_DecCheckSetTopOld(p, nVars, nVars-sizeB, sizeB, sizeS, maskS, NULL, NULL, NULL) )
            if Dau_DecCheckSetTop(p, nVars, nVars-sizeB, sizeB, sizeS, maskS, pSched[sizeB], nil, nil) != 0 {
                Vec_IntPush( vSets, setMixed ); }
        }
        } }
        }
    }
    }
    }
    return vSets;
}
func Dau_DecFindSets(pInit *word, nVars int32) *Vec_Int_t {
    var vSets *Vec_Int_t;
    var v int32;
    var pSched [16]*int32 = [16]*int32{nil};
    for  v = 2; v < nVars; v++  {
        pSched[v] = Extra_GreyCodeSchedule( v );
    }
    vSets = Dau_DecFindSets_int( pInit, nVars, pSched );
    for  v = 2; v < nVars; v++  {
        // inlining of macro ABC_FREE
        ((pSched[v]) ? (free((char *) (pSched[v])), (pSched[v]) = 0) : 0);
    }
    return vSets;
}
func Dau_DecFindSetsTest2() {
    var vSets *Vec_Int_t;
    var a0 word  = (~s_Truths6[1] & s_Truths6[2]) | (s_Truths6[1] & s_Truths6[3]);
    var a1 word  = (~s_Truths6[1] & s_Truths6[4]) | (s_Truths6[1] & s_Truths6[5]);
    var t word   = (~s_Truths6[0] & a0)           | (s_Truths6[0] & a1); 
//    word t = ABC_CONST(0x7EFFFFFFFFFFFF7E); // and(gam1,gam2)
//    word t = ABC_CONST(0xB0F0BBFFB0F0BAFE); // some funct
//    word t = ABC_CONST(0x2B0228022B022802); // 5-var non-dec0x0F7700000F770000
//    word t = ABC_CONST(0x0F7700000F770000); // (!<(ab)cd>e)
//    word t = ABC_CONST(0x7F00000000000000); // (!(abc)def)
    var nVars int32  = 5;

    vSets   = Dau_DecFindSets( &t, nVars );
    Dau_DecPrintSets( vSets, nVars );
    Vec_IntFree( vSets );
}


/**Function*************************************************************

  Synopsis    [Replaces variables in the string.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DecVarReplace(pStr string, pPerm *int32, nVars int32) {
    var v int32;
    for  v = 0; pStr[v]; v++  {
        if pStr[v] >= 'a' && pStr[v] <= 'z' {
            assert(pStr[v] - 'a' < nVars, "pStr[v] - 'a' < nVars");
            pStr[v] = 'a' + pPerm[pStr[v] - 'a'];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Decomposes with the given bound-set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DecDecomposeSet(pInit *word, nVars int32, uSet uint32, pComp *word, pDec *word, pPermC *int32, pPermD *int32, pnVarsC *int32, pnVarsD *int32, pnVarsS *int32) int32 {
    var p [1<<13]word;
    var Cof [64]word;
    var Cof0 [64]word;
    var Cof1 [64]word;
    var Decs [64]word;
    var pCof0 [64]*word;
    var pCof1 [64]*word;
    var pDecs [64]*word;
    var MintC word;
    var MintD word;
    var V2P, P2V, pVarsU, pVarsS, pVarsF [16]int32;
    var nVarsU, nVarsS, nVarsF int32;
    var nWords int32  = Abc_TtWordNum(nVars);
    var v, d, c, Status, nDecs int32;
    assert(nVars <= 16, "nVars <= 16");
    for  v = 0; v < nVars; v++  {
        V2P[v] = P2V[v] = v;
    }
    memcpy( p, pInit, sizeof(word) * nWords );
    // sort variables
    for  v = 0; v < nVars; v++  {
        var Value int32  = (uSet >> (v<<1)) & 3;
        if Value == 0 {
            pVarsF[nVarsF++] = v;
        } else if Value == 1 {
            pVarsU[nVarsU++] = v;
        } else if Value == 3 {
            pVarsS[nVarsS++] = v;
        } else { assert(false, "0"); }
    }
    assert(nVarsS >= 0 && nVarsS <= 6, "nVarsS >= 0 && nVarsS <= 6");
    assert(nVarsF + nVarsS + 1 <= 6, "nVarsF + nVarsS + 1 <= 6");
    assert(nVarsU + nVarsS <= 6, "nVarsU + nVarsS <= 6");
    // init space for decomposition functions
    nDecs = (1 << nVarsS);
    for  d = 0; d < nDecs; d++  {
        pCof0[d] = Cof0 + d;
        pCof1[d] = Cof1 + d;
        pDecs[d] = Decs + d;
    }
    // permute variables
    c = 0;
    for  v = 0; v < nVarsF; v++  {
       Abc_TtMoveVar( p, nVars, V2P, P2V, pVarsF[v], c++ );
    }
    for  v = 0; v < nVarsS; v++  {
       Abc_TtMoveVar( p, nVars, V2P, P2V, pVarsS[v], c++ );
    }
    for  v = 0; v < nVarsU; v++  {
       Abc_TtMoveVar( p, nVars, V2P, P2V, pVarsU[v], c++ );
    }
    assert(c == nVars, "c == nVars");
    // check decomposition
    Status = Dau_DecCheckSetTopOld( p, nVars, nVarsF, nVarsS+nVarsU, nVarsS, Abc_InfoMask(nVarsS), pCof0, pCof1, pDecs );
    if !Status {
        return 0; }
    // compute cofactors
    assert(nVarsF + nVarsS < 6, "nVarsF + nVarsS < 6");
    for  d = 0; d < nDecs; d++  {
        Cof[d] = (pCof1[d][0] & s_Truths6[nVarsF + nVarsS]) | (pCof0[d][0] & ~s_Truths6[nVarsF + nVarsS]);
        pDecs[d][0] = Abc_Tt6Stretch( pDecs[d][0], nVarsU );
    }
    // compute the resulting functions
    pComp[0] = 0;
    pDec[0] = 0;
    for  d = 0; d < nDecs; d++  {
        // compute minterms for composition/decomposition function
        MintC = MintD = ~((word)0);
        for  v = 0; v < nVarsS; v++  {
            MintC &= ((d >> v) & 1) ? s_Truths6[nVarsF+v] : ~s_Truths6[nVarsF+v];
            MintD &= ((d >> v) & 1) ? s_Truths6[nVarsU+v] : ~s_Truths6[nVarsU+v];
        }
        // derive functions
        pComp[0] |= MintC & Cof[d];
        pDec[0] |= MintD & pDecs[d][0];
    }
    // derive variable permutations
    if pPermC != nil {
        for  v = 0; v < nVarsF; v++  {
            pPermC[v] = pVarsF[v];
        }
        for  v = 0; v < nVarsS; v++  {
            pPermC[nVarsF+v] = pVarsS[v];
        }
        pPermC[nVarsF + nVarsS] = nVars;
    }
    if pPermD != nil {
        for  v = 0; v < nVarsU; v++  {
            pPermD[v] = pVarsU[v];
        }
        for  v = 0; v < nVarsS; v++  {
            pPermD[nVarsU+v] = pVarsS[v];
        }
    }
    if pnVarsC != nil {
        *pnVarsC = nVarsF + nVarsS + 1; }
    if pnVarsD != nil {
        *pnVarsD = nVarsU + nVarsS; }
    if pnVarsS != nil {
        *pnVarsS = nVarsS; }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Testing procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DecVerify(pInit *word, nVars int32, pDsdC string, pDsdD string) int32 {
    var pC, pD, pRes [1<<13]word; // max = 16
    var nWordsC int32  = Abc_TtWordNum(nVars+1);
    var nWordsD int32  = Abc_TtWordNum(nVars);
    assert(nVars < 16, "nVars < 16");
    memcpy( pC, Dau_DsdToTruth(pDsdC, nVars+1), sizeof(word) * nWordsC );
    memcpy( pD, Dau_DsdToTruth(pDsdD, nVars), sizeof(word) * nWordsD );
    if nVars >= 6 {
        assert(nWordsD >= 1, "nWordsD >= 1");
        assert(nWordsC > 1, "nWordsC > 1");
        Abc_TtMux( pRes, pD, pC + nWordsD, pC, nWordsD );
    
    } else {
        var pC0 word  = Abc_Tt6Stretch( pC[0], nVars );
        var pC1 word  = Abc_Tt6Stretch( (pC[0] >> (1 << nVars)), nVars );
        Abc_TtMux( pRes, pD, &pC1, &pC0, nWordsD );
    }
    if !Abc_TtEqual(pInit, pRes, nWordsD) {
        fmt.Printf("      Verification failed"); }
//    else
//        printf( "      Verification successful" );
    fmt.Printf("\n");
    return 1;
}
func Dau_DecPerform6(p *word, nVars int32, uSet uint32) int32 {
    var tComp, tDec, tDec0, tComp0, tComp1, FuncC, FuncD word;
    var pDsdC, pDsdD [1000]byte;
    var pPermC, pPermD [16]int32;
    var nVarsC, nVarsD, nVarsS, nVarsU, nVarsF, nPairs int32;
    var i, m, v, status, ResC, ResD, Counter int32;
    status = Dau_DecDecomposeSet( p, nVars, uSet, &tComp, &tDec0, pPermC, pPermD, &nVarsC, &nVarsD, &nVarsS );
    if !status {
        fmt.Printf("  Decomposition does not exist\n");
        return 0;
    }
    nVarsU = nVarsD - nVarsS;
    nVarsF = nVarsC - nVarsS - 1;
    tComp0 = Abc_Tt6Cofactor0( tComp, nVarsF + nVarsS );
    tComp1 = Abc_Tt6Cofactor1( tComp, nVarsF + nVarsS );
    nPairs = 1 << (1 << nVarsS); 
    for  i = 0; i < nPairs; i++  {
        if i & 1 {
            continue; }
        // create miterms with this polarity
        FuncC = FuncD = 0;
        for  m = 0; m < (1 << nVarsS); m++  {
            var MintC, MintD word;
            if !((i >> m) & 1) {
                continue; }
            MintC = MintD = ~(word)0;
            for  v = 0; v < nVarsS; v++  {
                MintC &= ((m >> v) & 1) ? s_Truths6[nVarsF+v] : ~s_Truths6[nVarsF+v];
                MintD &= ((m >> v) & 1) ? s_Truths6[nVarsU+v] : ~s_Truths6[nVarsU+v];
            }
            FuncC |= MintC;
            FuncD |= MintD;
        }
        // uncomplement given variables
        tComp = (~s_Truths6[nVarsF + nVarsS] & ((tComp0 & ~FuncC) | (tComp1 & FuncC))) | (s_Truths6[nVarsF + nVarsS] & ((tComp1 & ~FuncC) | (tComp0 & FuncC)));
        tDec = tDec0 ^ FuncD;
        // decompose
        ResC = Dau_DsdDecompose( &tComp, nVarsC, 0, 1, pDsdC );
        ResD = Dau_DsdDecompose( &tDec, nVarsD, 0, 1, pDsdD );
        // replace variables
        Dau_DecVarReplace( pDsdD, pPermD, nVarsD );
        Dau_DecVarReplace( pDsdC, pPermC, nVarsC );
        // report
//        printf( "  " );
        fmt.Printf("%3d : ", Counter++);
        fmt.Printf("%24s  ", pDsdD);
        fmt.Printf("%24s  ", pDsdC);
        Dau_DecVerify( p, nVars, pDsdC, pDsdD );
    }
    return 1;
}

func Dau_DecPerform(pInit *word, nVars int32, uSet uint32) int32 {
    var p, pDec, pComp [1<<10]word; // at most 2^10 words
    var pDsdC, pDsdD [5000]byte;  // at most 2^12 hex digits
    var nVarsU, nVarsS, nVarsF, nVarsC, nVarsD int32;
    var V2P [16]int32;
    var P2V [16]int32;
    var pPermC [16]int32;
    var pPermD [16]int32;
    var pSched *int32;
    var v, i, status, ResC, ResD int32;
    var nWords int32  = Abc_TtWordNum(nVars);
    assert(nVars <= 16, "nVars <= 16");
    // backup the function
    memcpy( p, pInit, sizeof(word) * nWords );
    // get variable numbers
    Dau_DecSortSet( uSet, nVars, &nVarsU, &nVarsS, &nVarsF );
    // permute function and order variables
    for  v = 0; v < nVars; v++  {
        V2P[v] = P2V[v] = v;
    }
    for  i = v = 0; v < nVars; v++  {
        if ((uSet >> (v<<1)) & 3) == 0 { // free first
           Abc_TtMoveVar( p, nVars, V2P, P2V, v, i++ );
pPermC[nVarsC++] = v; }
    }
    for  v = 0; v < nVars; v++  {
        if ((uSet >> (v<<1)) & 3) == 3 { // share second
           Abc_TtMoveVar( p, nVars, V2P, P2V, v, i++ );
pPermC[nVarsC++] = v; }
    }
    pPermC[nVarsC++] = nVars;
    for  v = 0; v < nVars; v++  {
        if ((uSet >> (v<<1)) & 3) == 1 { // unique last
           Abc_TtMoveVar( p, nVars, V2P, P2V, v, i++ );
pPermD[nVarsD++] = v; }
    }
    for  v = 0; v < nVarsS; v++  {
        pPermD[nVarsD++] = pPermC[nVarsF+v];
    }
    assert(nVarsD == nVarsU + nVarsS, "nVarsD == nVarsU + nVarsS");
    assert(nVarsC == nVarsF + nVarsS + 1, "nVarsC == nVarsF + nVarsS + 1");
    assert(i == nVars, "i == nVars");
    // decompose
    pSched = Extra_GreyCodeSchedule( nVarsU + nVarsS );
    memset( pDec, 0, sizeof(word) * Abc_TtWordNum(nVarsD) );
    memset( pComp, 0, sizeof(word) * Abc_TtWordNum(nVarsC) );
    status = Dau_DecCheckSetTop( p, nVars, nVarsF, nVarsU + nVarsS, nVarsS, nVarsS ? Abc_InfoMask(nVarsS) : 0, pSched, pDec, pComp );
    // inlining of macro ABC_FREE
    ((pSched) ? (free((char *) (pSched)), (pSched) = 0) : 0);
    if !status {
        fmt.Printf("  Decomposition does not exist\n");
        return 0;
    }
//    Dau_DsdPrintFromTruth( stdout, pC, nVars+1 ); //printf( "\n" );
//    Dau_DsdPrintFromTruth( stdout, pD, nVars ); //printf( "\n" );
//    Kit_DsdPrintFromTruth( (unsigned *)pComp, 6 ); printf( "\n" );
//    Kit_DsdPrintFromTruth( (unsigned *)pDec, 6 );  printf( "\n" );
    // decompose
    ResC = Dau_DsdDecompose( pComp, nVarsC, 0, 1, pDsdC );
    ResD = Dau_DsdDecompose( pDec, nVarsD, 0, 1, pDsdD );
    // replace variables
    Dau_DecVarReplace( pDsdD, pPermD, nVarsD );
    Dau_DecVarReplace( pDsdC, pPermC, nVarsC );
    // report
    fmt.Printf("     ");
    fmt.Printf("%3d : ", 0);
    fmt.Printf("%24s  ", pDsdD);
    fmt.Printf("%24s  ", pDsdC);
    Dau_DecVerify( pInit, nVars, pDsdC, pDsdD );
    return 1;
}
func Dau_DecTrySets(pInit *word, nVars int32, fVerbose int32) {
    var vSets *Vec_Int_t;
    var i, Entry int32;
    assert(nVars <= 16, "nVars <= 16");
    vSets = Dau_DecFindSets( pInit, nVars );
    if !(fVerbose != 0) {
        Vec_IntFree( vSets );
        return;
    }
    Dau_DsdPrintFromTruth( pInit, nVars ); 
    fmt.Printf("This %d-variable function has %d decomposable variable sets:\n", nVars, Vec_IntSize(vSets));
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSets)) && (((Entry) = Vec_IntEntry(vSets, i)), 1); i++  {
        var uSet uint32  = uint32(Entry);
        fmt.Printf("Set %4d : ", i);
        if nVars > 6 {
            Dau_DecPrintSet( uSet, nVars, 0 );
            Dau_DecPerform( pInit, nVars, uSet );
        
        } else {
            Dau_DecPrintSet( uSet, nVars, 1 );
            Dau_DecPerform6( pInit, nVars, uSet );
        }
    }
    Vec_IntFree( vSets );
//    printf( "\n" );
}

func Dau_DecFindSetsTest3() {
    var a0 word  = (~s_Truths6[1] & s_Truths6[2]) | (s_Truths6[1] & s_Truths6[3]);
    var a1 word  = (~s_Truths6[1] & s_Truths6[4]) | (s_Truths6[1] & s_Truths6[5]);
    var t word   = (~s_Truths6[0] & a0)           | (s_Truths6[0] & a1); 
//    word t = ABC_CONST(0x0F7700000F770000); // (!<(ab)cd>e)
    var nVars int32  = 6;
    var pStr string  = "Bcd";
//    char * pStr = "Abcd";
//    char * pStr = "ab";
    var uSet uint32  = Dau_DecReadSet( pStr );
    Dau_DecPerform6( &t, nVars, uSet );
}

func Dau_DecFindSetsTest() {
    var nVars int32  = 6;
//    word a0 = (~s_Truths6[1] & s_Truths6[2]) | (s_Truths6[1] & s_Truths6[3]);
//    word a1 = (~s_Truths6[1] & s_Truths6[4]) | (s_Truths6[1] & s_Truths6[5]);
//    word t  = (~s_Truths6[0] & a0)           | (s_Truths6[0] & a1); 
//    word t = ABC_CONST(0x7EFFFFFFFFFFFF7E); // and(gam1,gam2)
//    word t = ABC_CONST(0xB0F0BBFFB0F0BAFE); // some funct
//    word t = ABC_CONST(0x00000000901FFFFF); // some funct
    var t word  = ABC_CONST(0x000030F00D0D3FFF); // some funct
//    word t = ABC_CONST(0x00000000690006FF); // some funct
//    word t = ABC_CONST(0x7000F80007FF0FFF); // some funct
//    word t = ABC_CONST(0x4133CB334133CB33); // some funct 5 var
//    word t = ABC_CONST(0x2B0228022B022802); // 5-var non-dec0x0F7700000F770000
//    word t = ABC_CONST(0x0F7700000F770000); // (!<(ab)cd>e)
//    word t = ABC_CONST(0x7F00000000000000); // (!(abc)def)
    Dau_DecTrySets( &t, nVars, 1 );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

