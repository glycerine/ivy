package dau

/**CFile****************************************************************

  FileName    [dauDsd2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Disjoint-support decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauDsd2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

//#include DSD_MAX_VAR 12
//#include DSD_MAX_WRD ((DSD_MAX_VAR > 6) ? (1 << (DSD_MAX_VAR-6)) : 1)

   
type Dua_Obj_t struct {
    Type       int32;                // dec type (1=var; 2=and; 3=xor; 4=mux; 5=prime)
    nFans       int32;               // fanin count
    pFans [DSD_MAX_VAR]byte;  // fanins
}

   
type Dua_Dsd_t struct {
    nSupp       int32;               // original variables
    nVars       int32;               // remaining variables
    nWords       int32;              // largest non-dec prime
    nObjs       int32;               // object count
    iRoot       int32;               // the root of the tree
    pObjs [DSD_MAX_VAR]Dua_Obj_t;  // objects
    word      pTruth[DSD_MAX_WRD]; // original/current truth table
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Makes the fCof1-th cofactor of iVar the 0-th cofactor.]

  Description [Variable iVar becomes last varaible; others shift back.
  Only the 0-th cofactor is computed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Tt6HalfUnShuffleVars(t word, iVar int32, fCof1 int32) word {
    static word Masks[6] = {
        ABC_CONST(0x5555555555555555),
        ABC_CONST(0x3333333333333333),
        ABC_CONST(0x0F0F0F0F0F0F0F0F),
        ABC_CONST(0x00FF00FF00FF00FF),
        ABC_CONST(0x0000FFFF0000FFFF),
        ABC_CONST(0x00000000FFFFFFFF) 
    };
    var v int32;
    var s int32 = (1 << iVar);
    t = (t >> (fCof1 ? 0 : s)) & Masks[iVar];
    for  v = iVar, s = (1 << v); v < 5; v++, s <<= 1  {
        t = ((t >> s) | t) & Masks[v+1];
    }
    return t;
}

func Abc_TtHalfUnShuffleVars(pTruth *word, nVars int32, iVar int32, jVar int32, fCof1 int32) {
    var w int32;
    var nWords int32 = Abc_TtWordNum( nVars );
    if iVar == jVar {
        return; }
    assert(iVar < jVar, "iVar < jVar");
    if iVar < 5 {
        for  w = 0; w < nWords; w++  {
            pTruth[w] = Abc_Tt6HalfUnShuffleVars( pTruth[w], iVar, fCof1 );
        }
        iVar = 5;
    }
    if jVar < 6 {
        for  w = 0; w < nWords; w++  {
            pTruth[w] = (pTruth[w] << 32) | pTruth[w];
        }
        return;
    }
    if iVar == 5 {
        var pTruthU *uint32  = (unsigned *)pTruth;
        for  w = 0; w < nWords; w += 2  {
            pTruthU[w] = pTruthU[w+1];
        }
        iVar = 6;
    }
    {
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        var j int32;
        var jStep int32 = Abc_TtWordNum(jVar);
        for  ; pTruth < pLimit; pTruth += jStep  {
            for  i = 0; i < jStep; i += iStep  {
                for  j = 0; j < iStep; j++  {
                    pTruth[w++] = pTruth[iStep + i + j];
                }
            }
        }
        assert(w == (nWords >> 1), "w == (nWords >> 1)");
        return;
    }    
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dua_DsdInit(pRes *Dua_Dsd_t, pTruth *word, nVars int32) {
    var i int32;
    pRes.nSupp  = nVars;
    pRes.nVars  = nVars;
    pRes.nWords = Abc_TtWordNum( nVars );
    pRes.nObjs  = 1;
    pRes.iRoot  = Abc_Var2Lit( 0, 0 );
    pRes.pObjs[0].Type = 5;
    pRes.pObjs[0].nFans = nVars;
    for  i = 0; i < nVars; i++  {
        pRes.pObjs[0].pFans[i] = byte(Abc_Var2Lit( i, 0 ));
    }
    memcpy( pRes.pTruth, pTruth, sizeof(word) * pRes.nWords );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns 1 if the truth table was complemented
func Dua_DsdTryConst(pTruth *word, nVars int32) int32 {
    if !(pTruth[0] & 1) {
        return 0; }
    Abc_TtNot( pTruth, Abc_TtWordNum(nVars) );
    return 1;
}
func Dua_DsdTryVar(pTruth *word, nWords int32, iVar int32) int32 {
    var nWordsI int32  = Abc_TtWordNum(iVar);
    var c0 word;
    if iVar < 6 {
        c0 = Abc_Tt6Cofactor0( pTruth[0], iVar );
    } else {
        c0 = pTruth[0];
    }
    var c1 word;
    if iVar < 6 {
        c1 = Abc_Tt6Cofactor1( pTruth[0], iVar );
    } else {
        c1 = pTruth[nWords];
    }
    if c0 != c1 {
        if c1 < c0 && c1 < ~c1 { // flip
        
            Abc_TtFlip( pTruth, nWords, iVar );
            return 0;
        }
        if ~c1 < c0 && ~c1 < c1 { // flip and compl
        
            Abc_TtFlipNot( pTruth, nWords, iVar );
            return 1;
        }
    }
    if iVar < 6 {
        var pLimit *word  = pTruth + nWords;
        for  pTruth++; pTruth < pLimit; pTruth++  {
            c0 = Abc_Tt6Cofactor0( pTruth[0], iVar );
            c1 = Abc_Tt6Cofactor1( pTruth[0], iVar );
            if c0 == c1 {
                continue; }
            if c0 < c1 {
                return 0; }
            for  ; pTruth < pLimit; pTruth++  {
                pTruth[0] = Abc_Tt6Flip( pTruth[0], iVar );
            }
            return 0;
        }
    
    } else {
        for  ; pTruth < pLimit; pTruth += (nWordsI << 1)  {
        for  w = 0; w < nWordsI; w++  {
            c0 = pTruth[0];
            c1 = pTruth[nWordsI];
            if c0 == c1 {
                continue; }
            if c0 < c1 {
                return 0; }
            for  ; pTruth < pLimit; pTruth += (nWordsI << 1)  {
            for  ; w < nWordsI; w++  {
                // inlining of macro ABC_SWAP
                 var t word  = pTruth[0]; pTruth[0] = pTruth[nWordsI]; pTruth[nWordsI] = t; }
            };
            return 0;
        }
        }
    }
    assert(false, "0");
    return -1;
}
func Dua_DsdCheckCof0Const0(pTruth *word, nWords int32, iVar int32) int32 {
    if nWords == 1 {
        return (pTruth[0] & s_Truths6Neg[iVar]) == 0; }
    if iVar <= 5 {
        var w int32;
        for  w = 0; w < nWords; w++  {
            if pTruth[w] & s_Truths6Neg[iVar] {
                return 0; }
        }
        return 1;
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += (iStep << 1)  {
            for  i = 0; i < iStep; i++  {
                if pTruth[i] {
                    return 0; }
            }
        }
        return 1;
    }    
}
func Dua_DsdCheckCofsEqualNot(pTruth *word, nWords int32, iVar int32) int32 {
    if nWords == 1 {
        return (pTruth[0] & s_Truths6Neg[iVar]) == ((~pTruth[0] & s_Truths6[iVar]) >> (1 << iVar)); }
    if iVar <= 5 {
        var w int32;
        var shift int32 = (1 << iVar);
        for  w = 0; w < nWords; w++  {
            if (pTruth[w] & s_Truths6Neg[iVar]) != ((~pTruth[w] & s_Truths6[iVar]) >> shift) {
                return 0; }
        }
        return 1;
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += (iStep << 1)  {
            for  i = 0; i < iStep; i++  {
                if pTruth[i] != ~pTruth[i + iStep] {
                    return 0; }
            }
        }
        return 1;
    }    
}
func Dua_DsdOneVar(pRes *Dua_Dsd_t) int32 {
    var v int32;
    var fCompl int32;
    var fChange int32 = 1;
    fCompl = Dua_DsdTryConst( pRes.pTruth, pRes.nWords );
    for  (fChange != 0) && pRes.nVars > 2  {
        fChange = 0;
        for  v = 0; v < pRes.nVars; v++  {
            fCompl ^= Dua_DsdTryVar( pRes.pTruth, pRes.nWords, v );
            if Dua_DsdCheckCof0Const0( pRes.pTruth, pRes.nWords, v ) != 0 {
                fChange = 1;
                // record AND(v, F)
            
            } else if Dua_DsdCheckCofsEqualNot( pRes.pTruth, pRes.nWords, v ) != 0 {
                fChange = 1;
                // record XOR(v, F)
            }
        }
    }
    return fCompl;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dua_DsdTrySwap(pTruth *word, nWords int32, iVar int32) int32 {
    static word s_PMasks[5][3] = {
        { ABC_CONST(0x9999999999999999), ABC_CONST(0x2222222222222222), ABC_CONST(0x4444444444444444) },
        { ABC_CONST(0xC3C3C3C3C3C3C3C3), ABC_CONST(0x0C0C0C0C0C0C0C0C), ABC_CONST(0x3030303030303030) },
        { ABC_CONST(0xF00FF00FF00FF00F), ABC_CONST(0x00F000F000F000F0), ABC_CONST(0x0F000F000F000F00) },
        { ABC_CONST(0xFF0000FFFF0000FF), ABC_CONST(0x0000FF000000FF00), ABC_CONST(0x00FF000000FF0000) },
        { ABC_CONST(0xFFFF00000000FFFF), ABC_CONST(0x00000000FFFF0000), ABC_CONST(0x0000FFFF00000000) }
    };
    if iVar < 5 {
        var Shift int32  = (1 << iVar);
        var c01 word;
        var c10 word;
        var pLimit *word = pTruth + nWords;
        for  ; pTruth < pLimit; pTruth++  {
            c01 = (pTruth[0] & s_PMasks[iVar][1]);
            c10 = (pTruth[0] & s_PMasks[iVar][2]) >> Shift;
            if c01 == c10 {
                continue; }
            if c01 < c10 {
                return 0; }
            pTruth[0] = (pTruth[0] & s_PMasks[iVar][0]) | ((pTruth[0] & s_PMasks[iVar][1]) << Shift) | ((pTruth[0] & s_PMasks[iVar][2]) >> Shift);
            return 1;
        }
    
    } else if iVar == 5 {
        var pTruthU *uint32  = (unsigned *)pTruth;
        var pLimitU *uint32  = (unsigned *)(pTruth + nWords);
        for  ; pTruthU < pLimitU; pTruthU += 4  {
            c01 = pTruthU[1];
            c10 = pTruthU[2];
            if c01 == c10 {
                continue; }
            if c01 < c10 {
                return 0; }
            for  ; pTruthU < pLimitU; pTruthU += 4  {
                // inlining of macro ABC_SWAP
                 var t uint32  = pTruthU[1]; pTruthU[1] = pTruthU[2]; pTruthU[2] = t; };
            return 1;
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += 4*iStep  {
        for  i = 0; i < iStep; i++  {
            c01 = pTruth[i + iStep];
            c10 = pTruth[i + 2*iStep];
            if c01 == c10 {
                continue; }
            if c01 < c10 {
                return 0; }
            for  ; pTruth < pLimit; pTruth += 4*iStep  {
            for  ; i < iStep; i++  {
                // inlining of macro ABC_SWAP
                 var t word  = pTruth[1]; pTruth[1] = pTruth[2]; pTruth[2] = t; }
            };
            return 1;
        }
        }
    }
    return 2;
}
func Dua_DsdCheckDecomp(pTruth *word, nWords int32, iVar int32) int32 {
    static word s_PMasks[5][4] = {
        { ABC_CONST(0x1111111111111111), ABC_CONST(0x2222222222222222), ABC_CONST(0x4444444444444444), ABC_CONST(0x8888888888888888) },
        { ABC_CONST(0x0303030303030303), ABC_CONST(0x0C0C0C0C0C0C0C0C), ABC_CONST(0x3030303030303030), ABC_CONST(0xC0C0C0C0C0C0C0C0) },
        { ABC_CONST(0x000F000F000F000F), ABC_CONST(0x00F000F000F000F0), ABC_CONST(0x0F000F000F000F00), ABC_CONST(0xF000F000F000F000) },
        { ABC_CONST(0x000000FF000000FF), ABC_CONST(0x0000FF000000FF00), ABC_CONST(0x00FF000000FF0000), ABC_CONST(0xFF000000FF000000) },
        { ABC_CONST(0x000000000000FFFF), ABC_CONST(0x00000000FFFF0000), ABC_CONST(0x0000FFFF00000000), ABC_CONST(0xFFFF000000000000) }
    };
    var fC0eC1 int32 = 1;
    var fC0eC3 int32 = 1;
    if iVar < 5 {
        var Shift int32  = (1 << iVar);
        var c01 word;
        var c10 word;
        var pLimit *word = pTruth + nWords;
        for  ; pTruth < pLimit; pTruth++  {
            if (fC0eC1 != 0) && (pTruth[0] & s_PMasks[iVar][0]) != ((pTruth[0] & s_PMasks[iVar][1]) >>    Shift) {
                fC0eC1 = 0; }
            if (fC0eC3 != 0) && (pTruth[0] & s_PMasks[iVar][0]) != ((pTruth[0] & s_PMasks[iVar][3]) >> (3*Shift)) {
                fC0eC3 = 0; }
            if !(fC0eC1 != 0) && !(fC0eC3 != 0) {
                return 0; }
        }
    }
    if iVar == 5 {
        var pTruthU *uint32  = (unsigned *)pTruth;
        var pLimitU *uint32  = (unsigned *)(pTruth + nWords);
        for  ; pTruthU < pLimitU; pTruthU += 4  {
            if (fC0eC1 != 0) && pTruthU[0] != pTruthU[1] {
                fC0eC1 = 0; }
            if (fC0eC3 != 0) && pTruthU[0] != pTruthU[3] {
                fC0eC3 = 0; }
            if !(fC0eC1 != 0) && !(fC0eC3 != 0) {
                return 0; }
        }
    
    } else { // if ( iVar > 5 )
    
        var pLimit *word  = pTruth + nWords;
        var i int32;
        var iStep int32 = Abc_TtWordNum(iVar);
        for  ; pTruth < pLimit; pTruth += 4*iStep  {
        for  i = 0; i < iStep; i++  {
            if (fC0eC1 != 0) && pTruth[0] != pTruth[1] {
                fC0eC1 = 0; }
            if (fC0eC3 != 0) && pTruth[0] != pTruth[3] {
                fC0eC3 = 0; }
            if !(fC0eC1 != 0) && !(fC0eC3 != 0) {
                return 0; }
        }
        }
    }
    assert(fC0eC1 != fC0eC3, "fC0eC1 != fC0eC3");
    if fC0eC1 != 0 {
        return 1;
    } else {
        return 2;
    }
}

// returns 1 if decomposition detected
func Dua_DsdTwoVars(pRes *Dua_Dsd_t) int32 {
    var v int32;
    var RetValue int32;
    var fChange int32 = 1;
    for  (fChange != 0) && pRes.nVars > 2  {
        fChange = 0;
        for  v = 0; v < pRes.nVars - 1; v++  {
            RetValue = Dua_DsdTrySwap( pRes.pTruth, pRes.nWords, v );            
            if RetValue == 1 {
                fChange = 1; }
            if RetValue != 2 {
                continue; }
            // vars are symmetric, check decomp
            RetValue = Dua_DsdCheckDecomp( pRes.pTruth, pRes.nWords, v );
            if RetValue == 0 {
                continue; }
            if RetValue == 1 {
                fChange = 1;
                // record AND(a, b)
            
            } else {
                fChange = 1;
                // record XOR(a, b)
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Check DSD for bound-set [iVar; jVar).]

  Description [Return D-func if decomposable.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dua_DsdRangeVars(pTruth *word, nVars int32, iVar int32, jVar int32, fPerform int32) word {
    var Part int32;
    var nParts int32 = 1 << (nVars - jVar);
    var Mint int32;
    var nMints int32 = 1 << (jVar - iVar);
    var MaskOne, MaskAll word;
    assert(jVar - iVar > 2, "jVar - iVar > 2");
    assert(jVar - iVar < 7, "jVar - iVar < 7");
    if iVar < 6 {
        var Shift int32 = 6 - iVar;
        var MaskF int32 = (1 << Shift) - 1;
        var iMint int32;
        var MaskFF word  = (((word)1) << (1 << iVar)) - 1;
        var Cof0, Cof1, Value word; 
        for  Part = 0; Part < nParts; Part++  {
            MaskOne = 0;
            Cof0 = Cof1 = ~(word)0;
            for  Mint = 0; Mint < nMints; Mint, iMint = Mint+1, iMint+1  {
                Value = (pTruth[iMint>>Shift] >> ((iMint & MaskF)<<iVar)) & MaskFF;
                if !~Cof0 || Cof0 == Value {
                    Cof0 = Value;
                } else if !~Cof1 || Cof1 == Value {
                    Cof1 = Value;
                    MaskOne |= ((word)1) << Mint;
                
                } else {
                    return 0; }
            }
            if Part == 0 {
                MaskAll = MaskOne;
            } else if MaskAll != MaskOne {
                return 0; }
            if fPerform != 0 {
                assert(~Cof0 && ~Cof1, "~Cof0 && ~Cof1");
                Mint = 2 * Part;
                Value = (pTruth[Mint>>Shift] >> ((Mint & MaskF)<<nVarsF)) & MaskFF;
                pTruth[Mint>>Shift] ^= (Value ^ Cof0) << ((Mint & MaskF)<<nVarsF)
                Mint = 2 * Part + 1;
                Value = (pTruth[Mint>>Shift] >> ((Mint & MaskF)<<nVarsF)) & MaskFF;
                pTruth[Mint>>Shift] ^= (Value ^ Cof1) << ((Mint & MaskF)<<nVarsF)
            }
        }
        // stretch
        if nVars - (jVar - iVar) + 1 < 6 {
            pTruth[0] = Abc_Tt6Stretch( pTruth[0], nVars - (jVar - iVar) + 1 ); }
    
    } else {
        var nWordsF int32  = Abc_TtWordNum(iVar); 
        var iWord int32;
        var nBytes int32 = sizeof(word) * nWordsF;
        var pCof0, pCof1 *word;
        for  Part = 0; Part < nParts; Part++  {
            MaskOne = 0;
            pCof0 = pCof1 = nil;
            for  Mint = 0; Mint < nMints; Mint++, iWord += nWordsF  {
                if !pCof0 || !memcmp(pCof0, pTruth + iWord, (size_t)nBytes) {
                    pCof0 = pTruth + iWord;
                } else if !pCof1 || !memcmp(pCof1, pTruth + iWord, (size_t)nBytes) {
                    pCof1 = pTruth + iWord;
                    MaskOne |= ((word)1) << Mint;
                
                } else {
                    return 0; }
            }
            if Part == 0 {
                MaskAll = MaskOne;
            } else if MaskAll != MaskOne {
                return 0; }
            if fPerform != 0 {
                assert(pCof0 && (pCof1 != nil), "pCof0 && pCof1");
                memcpy( pTruth + (2 * Part + 0) * nWordsF, pCof0, (size_t)nBytes );
                memcpy( pTruth + (2 * Part + 1) * nWordsF, pCof1, (size_t)nBytes );
            }
        }
    }
    return MaskAll;
}

/**Function*************************************************************

  Synopsis    [Check DSD for bound-set [0; iVar).]

  Description [Return D-func if decomposable.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dua_DsdRangeVars0(pTruth *word, nVars int32, iVar int32, fPerform int32) int32 {
    var i int32;
    var nParts int32 = 1 << (nVars - iVar);
    assert(iVar > 2 && iVar < nVars, "iVar > 2 && iVar < nVars");
    if iVar == 3 {
        unsigned char * pTruthP = (unsigned char *)pTruth, Dfunc = pTruthP[0];
        for  i = 1; i < nParts; i++  {
            if pTruthP[i] != Dfunc && pTruthP[i] != ~Dfunc {
                return 0; }
        }
    
    } else if iVar == 4 {
        unsigned short * pTruthP = (unsigned short *)pTruth, Dfunc = pTruthP[0];
        for  i = 1; i < nParts; i++  {
            if pTruthP[i] != Dfunc && pTruthP[i] != ~Dfunc {
                return 0; }
        }
    
    } else if iVar == 5 {
        unsigned int * pTruthP = (unsigned int *)pTruth, Dfunc = pTruthP[0];
        for  i = 1; i < nParts; i++  {
            if pTruthP[i] != Dfunc && pTruthP[i] != ~Dfunc {
                return 0; }
        }
    
    } else {
        var nStep int32  = 1 << (6 - iVar);
        assert(iVar >= 6, "iVar >= 6");
        for  i = 1; i < nParts; i++  {
            if !Abc_TtEqual(pTruth, pTruth + i * nStep, nStep) && !Abc_TtEqualNot(pTruth, pTruth + i * nStep, nStep) {
                return 0; }
        }
    }
    return 1;
}
func Dua_DsdRangeVars0Derive(pTruth *word, nVars int32, iVar int32) {
    var i int32;
    var nParts int32 = 1 << (nVars - iVar);
    assert(iVar > 2 && iVar < nVars, "iVar > 2 && iVar < nVars");
    if iVar == 3 {
        unsigned char * pTruthP = (unsigned char *)pTruth, Dfunc = pTruthP[0];
        for  i = 0; i < nParts; i++  {
            if Abc_TtGetBit(pTruth, i) ^ (pTruthP[i] != Dfunc) {
                Abc_TtXorBit(pTruth, i); }
        }
    
    } else if iVar == 4 {
        unsigned short * pTruthP = (unsigned short *)pTruth, Dfunc = pTruthP[0];
        for  i = 0; i < nParts; i++  {
            if Abc_TtGetBit(pTruth, i) ^ (pTruthP[i] != Dfunc) {
                Abc_TtXorBit(pTruth, i); }
        }
    
    } else if iVar == 5 {
        unsigned int * pTruthP = (unsigned int *)pTruth, Dfunc = pTruthP[0];
        for  i = 0; i < nParts; i++  {
            if Abc_TtGetBit(pTruth, i) ^ (pTruthP[i] != Dfunc) {
                Abc_TtXorBit(pTruth, i); }
        }
    
    } else {
        var Dfunc word  = pTruth[0];
        assert(iVar == 6, "iVar == 6");
        for  i = 0; i < nParts; i++  {
            if Abc_TtGetBit(pTruth, i) ^ (pTruth[i] != Dfunc) {
                Abc_TtXorBit(pTruth, i); }
        }
   }
    // stretch
    if nVars - iVar + 1 < 6 {
        pTruth[0] = Abc_Tt6Stretch( pTruth[0], nVars - iVar + 1 < 6 ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dua_DsdTest(pTruth *word, nVar int32) {
    var Res Dua_Dsd_t;
    var pRes *Dua_Dsd_t = &Res;
    Dua_DsdInit( pRes, pTruth, nVars );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END


