package mio

/**CFile****************************************************************

  FileName    [exp.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Boolean expression.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: exp.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 



////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

const EXP_CONST0 = -1
const EXP_CONST1 = -2

func Exp_Const0() *Vec_Int_t {
    var vExp *Vec_Int_t;
    vExp = Vec_IntAlloc( 1 );
    Vec_IntPush( vExp, EXP_CONST0 );
    return vExp;
}
func Exp_Const1() *Vec_Int_t {
    var vExp *Vec_Int_t;
    vExp = Vec_IntAlloc( 1 );
    Vec_IntPush( vExp, EXP_CONST1 );
    return vExp;
}
func Exp_IsConst(p *Vec_Int_t) int32 {
    return Vec_IntEntry(p,0) == EXP_CONST0 || Vec_IntEntry(p,0) == EXP_CONST1;
}
func Exp_IsConst0(p *Vec_Int_t) int32 {
    return Vec_IntEntry(p,0) == EXP_CONST0;
}
func Exp_IsConst1(p *Vec_Int_t) int32 {
    return Vec_IntEntry(p,0) == EXP_CONST1;
}
func Exp_Var(iVar int32) *Vec_Int_t {
    var vExp *Vec_Int_t;
    vExp = Vec_IntAlloc( 1 );
    Vec_IntPush( vExp, 2 * iVar );
    return vExp;
}
func Exp_LitShift(nVars int32, Lit int32, Shift int32) int32 {
    if Lit < 2 * nVars {
        return Lit; }
    return Lit + 2 * Shift;
}
func Exp_IsLit(p *Vec_Int_t) int32 {
    return Vec_IntSize(p) == 1 && !Exp_IsConst(p);
}
func Exp_NodeNum(p *Vec_Int_t) int32 {
    return Vec_IntSize(p)/2;
}
func Exp_Not(p *Vec_Int_t) *Vec_Int_t {
    Vec_IntWriteEntry( p, 0, Vec_IntEntry(p,0) ^ 1 );
    return p;
}
func Exp_PrintLit(nVars int32, Lit int32) {
    if Lit == EXP_CONST0 {
        Abc_Print( 1, "Const0" );
    } else if Lit == EXP_CONST1 {
        Abc_Print( 1, "Const1" );
    } else if Lit < 2 * nVars {
        Abc_Print( 1, "%s%c", (Lit&1) ? "!" : " ", 'a' + Lit/2 );
    } else {
        Abc_Print( 1, "%s%d", (Lit&1) ? "!" : " ", Lit/2 ); }
}
func Exp_Print(nVars int32, p *Vec_Int_t) {
    var i int32;
    for  i = 0; i < Exp_NodeNum(p); i++  {
        Abc_Print( 1, "%2d = ", nVars + i );
        Exp_PrintLit( nVars, Vec_IntEntry(p, 2*i+0) );
        Abc_Print( 1, " & " );
        Exp_PrintLit( nVars, Vec_IntEntry(p, 2*i+1) );
        Abc_Print( 1, "\n" );
    }
    Abc_Print( 1, " F = " );
    Exp_PrintLit( nVars, Vec_IntEntryLast(p) );
    Abc_Print( 1, "\n" );
}
func Exp_Reverse(p *Vec_Int_t) *Vec_Int_t {
    Vec_IntReverseOrder( p );
    return p;
}
func Exp_PrintReverse(nVars int32, p *Vec_Int_t) {
    Exp_Reverse( p );
    Exp_Print( nVars, p );
    Exp_Reverse( p );
}
func Exp_And(pMan *int32, nVars int32, p0 *Vec_Int_t, p1 *Vec_Int_t, fCompl0 int32, fCompl1 int32) *Vec_Int_t {
    var i int32;
    var Len0 int32 = Vec_IntSize(p0);
    var Len1 int32 = Vec_IntSize(p1);
    var r *Vec_Int_t  = Vec_IntAlloc( Len0 + Len1 + 1 );
    assert((Len0 & 1) && (Len1 & 1), "(Len0 & 1) && (Len1 & 1)");
    Vec_IntPush( r, 2 * (nVars + Len0/2 + Len1/2) );
    Vec_IntPush( r, Exp_LitShift( nVars, Vec_IntEntry(p0, 0) ^ fCompl0, Len1/2 ) );
    Vec_IntPush( r, Vec_IntEntry(p1, 0) ^ fCompl1 );
    for  i = 1; i < Len0; i++  {
        Vec_IntPush( r, Exp_LitShift( nVars, Vec_IntEntry(p0, i), Len1/2 ) );
    }
    for  i = 1; i < Len1; i++  {
        Vec_IntPush( r, Vec_IntEntry(p1, i) );
    }
    assert(Vec_IntSize(r) == Len0 + Len1 + 1, "Vec_IntSize(r) == Len0 + Len1 + 1");
    return r;
}
func Exp_Or(pMan *int32, nVars int32, p0 *Vec_Int_t, p1 *Vec_Int_t) *Vec_Int_t {
    return Exp_Not( Exp_And(pMan, nVars, p0, p1, 1, 1) );
}
func Exp_Xor(pMan *int32, nVars int32, p0 *Vec_Int_t, p1 *Vec_Int_t) *Vec_Int_t {
    var i int32;
    var Len0 int32 = Vec_IntSize(p0);
    var Len1 int32 = Vec_IntSize(p1);
    var r *Vec_Int_t  = Vec_IntAlloc( Len0 + Len1 + 5 );
    assert((Len0 & 1) && (Len1 & 1), "(Len0 & 1) && (Len1 & 1)");
    Vec_IntPush( r, 2 * (nVars + Len0/2 + Len1/2 + 2)     );
    Vec_IntPush( r, 2 * (nVars + Len0/2 + Len1/2 + 1) + 1 );
    Vec_IntPush( r, 2 * (nVars + Len0/2 + Len1/2 + 0) + 1 );
    Vec_IntPush( r, Exp_LitShift( nVars, Vec_IntEntry(p0, 0) ^ 1, Len1/2 ) );
    Vec_IntPush( r, Vec_IntEntry(p1, 0) );
    Vec_IntPush( r, Exp_LitShift( nVars, Vec_IntEntry(p0, 0), Len1/2 ) );
    Vec_IntPush( r, Vec_IntEntry(p1, 0) ^ 1 );
    for  i = 1; i < Len0; i++  {
        Vec_IntPush( r, Exp_LitShift( nVars, Vec_IntEntry(p0, i), Len1/2 ) );
    }
    for  i = 1; i < Len1; i++  {
        Vec_IntPush( r, Vec_IntEntry(p1, i) );
    }
    assert(Vec_IntSize(r) == Len0 + Len1 + 5, "Vec_IntSize(r) == Len0 + Len1 + 5");
    return Exp_Not( r );
}
func Exp_Truth6Lit(nVars int32, Lit int32, puFanins *word, puNodes *word) word {
    if Lit == EXP_CONST0 {
        return 0; }
    if Lit == EXP_CONST1 {
        return ~(word)0; }
    if Lit < 2 * nVars {
        if Lit&1 {
            return ~puFanins[Lit/2];
        } else {
            return puFanins[Lit/2];
        } }
    if Lit&1 {
        return ~puNodes[Lit/2-nVars];
    } else {
        return puNodes[Lit/2-nVars];
    }
}
func Exp_Truth6(nVars int32, p *Vec_Int_t, puFanins *word) word {
    static word Truth6[6] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00),
        ABC_CONST(0xFFFF0000FFFF0000),
        ABC_CONST(0xFFFFFFFF00000000)
    };
    var puNodes *word;
    var Res word;
    var i int32;
    if puFanins == nil {
        puFanins = (word *)Truth6; }
    puNodes = ABC_CALLOC( word, Exp_NodeNum(p) );
    for  i = 0; i < Exp_NodeNum(p); i++  {
        puNodes[i] = Exp_Truth6Lit( nVars, Vec_IntEntry(p, 2*i+0), puFanins, puNodes ) & 
                     Exp_Truth6Lit( nVars, Vec_IntEntry(p, 2*i+1), puFanins, puNodes );
    }
    Res = Exp_Truth6Lit( nVars, Vec_IntEntryLast(p), puFanins, puNodes );
    ABC_FREE( puNodes );
    return Res;
}
func Exp_Truth8(nVars int32, p *Vec_Int_t, puFanins **word, puRes *word) {
    var Truth8 [8][4]word = [8][4]word{
        { ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA),ABC_CONST(0xAAAAAAAAAAAAAAAA) },
        { ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC),ABC_CONST(0xCCCCCCCCCCCCCCCC) },
        { ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0),ABC_CONST(0xF0F0F0F0F0F0F0F0) },
        { ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00),ABC_CONST(0xFF00FF00FF00FF00) },
        { ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000),ABC_CONST(0xFFFF0000FFFF0000) },
        { ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000),ABC_CONST(0xFFFFFFFF00000000) },
        { ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF) },
        { ABC_CONST(0x0000000000000000),ABC_CONST(0x0000000000000000),ABC_CONST(0xFFFFFFFFFFFFFFFF),ABC_CONST(0xFFFFFFFFFFFFFFFF) },
    };
    var puFaninsInt [8]*word;
    var pStore *word;
    var pThis *word = nil;
    var i int32;
    var k int32;
    var iRoot int32 = Vec_IntEntryLast(p);
    if puFanins == nil {
        puFanins = puFaninsInt;
        for  k = 0; k < 8; k++  {
            puFanins[k] = Truth8[k];
        }
    }
    if Exp_NodeNum(p) == 0 {
        assert(iRoot < 2 * nVars, "iRoot < 2 * nVars");
        if iRoot == EXP_CONST0 || iRoot == EXP_CONST1 {
            for  k = 0; k < 4; k++  {
                puRes[k] = iRoot == EXP_CONST0 ? 0 : ~(word)0;
            }
        } else {
            for  k = 0; k < 4; k++  {
                puRes[k] = Abc_LitIsCompl(iRoot) ? ~puFanins[Abc_Lit2Var(iRoot)][k] : puFanins[Abc_Lit2Var(iRoot)][k];
            } }
        return;
    }
    pStore = ABC_CALLOC( word, 4 * Exp_NodeNum(p) );
    for  i = 0; i < Exp_NodeNum(p); i++  {
        var iVar0 int32  = Abc_Lit2Var( Vec_IntEntry(p, 2*i+0) );
        var iVar1 int32  = Abc_Lit2Var( Vec_IntEntry(p, 2*i+1) );
        var fCompl0 int32  = Abc_LitIsCompl( Vec_IntEntry(p, 2*i+0) );
        var fCompl1 int32  = Abc_LitIsCompl( Vec_IntEntry(p, 2*i+1) );
        var pIn0 *word;
        if iVar0 < nVars {
            pIn0 = puFanins[iVar0];
        } else {
            pIn0 = pStore + 4 * (iVar0 - nVars);
        }
        var pIn1 *word;
        if iVar1 < nVars {
            pIn1 = puFanins[iVar1];
        } else {
            pIn1 = pStore + 4 * (iVar1 - nVars);
        }
        pThis = pStore + 4 * i;
        if (fCompl0 != 0) && (fCompl1 != 0) {
            for  k = 0; k < 4; k++  {
                pThis[k] = ~pIn0[k] & ~pIn1[k];
            }
        } else if (fCompl0 != 0) && !(fCompl1 != 0) {
            for  k = 0; k < 4; k++  {
                pThis[k] = ~pIn0[k] &  pIn1[k];
            }
        } else if !(fCompl0 != 0) && (fCompl1 != 0) {
            for  k = 0; k < 4; k++  {
                pThis[k] =  pIn0[k] & ~pIn1[k];
            }
        } else { //if ( !fCompl0 && !fCompl1 )
            for  k = 0; k < 4; k++  {
                pThis[k] =  pIn0[k] &  pIn1[k];
            } }
    }
    assert(Abc_Lit2Var(iRoot) - nVars == i - 1, "Abc_Lit2Var(iRoot) - nVars == i - 1");
    for  k = 0; k < 4; k++  {
        puRes[k] = Abc_LitIsCompl(iRoot) ? ~pThis[k] : pThis[k];
    }
    ABC_FREE( pStore );
}
func Exp_TruthLit(nVars int32, Lit int32, puFanins **word, puNodes **word, pRes *word, nWords int32) {
    var w int32;
    if Lit == EXP_CONST0 {
        for  w = 0; w < nWords; w++  {
            pRes[w] = 0;
        }
    } else if Lit == EXP_CONST1 {
        for  w = 0; w < nWords; w++  {
            pRes[w] = ~(word)0;
        }
    } else if Lit < 2 * nVars {
        for  w = 0; w < nWords; w++  {
            pRes[w] = (Lit&1) ? ~puFanins[Lit/2][w] : puFanins[Lit/2][w];
        }
    } else {
        for  w = 0; w < nWords; w++  {
            pRes[w] = (Lit&1) ? ~puNodes[Lit/2-nVars][w] : puNodes[Lit/2-nVars][w];
        } }
}
func Exp_Truth(nVars int32, p *Vec_Int_t, pRes *word) {
    static word Truth6[6] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00),
        ABC_CONST(0xFFFF0000FFFF0000),
        ABC_CONST(0xFFFFFFFF00000000)
    };
    var puFanins **word;
    var puNodes **word;
    var pTemp0 *word;
    var pTemp1 *word;
    var i int32;
    var w int32;
    var nWords int32;
    if nVars <= 6 {
        nWords = 1;
    } else {
        nWords = 1 << (nVars-6);
    }
    // create elementary variables
    puFanins = ABC_ALLOC( word *, nVars );
    for  i = 0; i < nVars; i++  {
        puFanins[i] = ABC_ALLOC( word, nWords );
    }
    // assign elementary truth tables
    for  i = 0; i < nVars; i++  {
        if i < 6 {
            for  w = 0; w < nWords; w++  {
                puFanins[i][w] = Truth6[i];
            }
        } else {
            for  w = 0; w < nWords; w++  {
                puFanins[i][w] = (w & (1 << (i-6))) ? ~(word)0 : 0;
            } }
    }
    // create intermediate nodes
    puNodes = ABC_ALLOC( word *, Exp_NodeNum(p) );
    for  i = 0; i < Exp_NodeNum(p); i++  {
        puNodes[i] = ABC_ALLOC( word, nWords );
    }
    // evaluate the expression
    pTemp0 = ABC_ALLOC( word, nWords );
    pTemp1 = ABC_ALLOC( word, nWords );
    for  i = 0; i < Exp_NodeNum(p); i++  {
        Exp_TruthLit( nVars, Vec_IntEntry(p, 2*i+0), puFanins, puNodes, pTemp0, nWords );
        Exp_TruthLit( nVars, Vec_IntEntry(p, 2*i+1), puFanins, puNodes, pTemp1, nWords );
        for  w = 0; w < nWords; w++  {
            puNodes[i][w] = pTemp0[w] & pTemp1[w];
        }
    }
    ABC_FREE( pTemp0 );
    ABC_FREE( pTemp1 );
    // copy the final result
    Exp_TruthLit( nVars, Vec_IntEntryLast(p), puFanins, puNodes, pRes, nWords );
    // cleanup
    for  i = 0; i < nVars; i++  {
        ABC_FREE( puFanins[i] );
    }
    ABC_FREE( puFanins );
    for  i = 0; i < Exp_NodeNum(p); i++  {
        ABC_FREE( puNodes[i] );
    }
    ABC_FREE( puNodes );
}

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_END




////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

