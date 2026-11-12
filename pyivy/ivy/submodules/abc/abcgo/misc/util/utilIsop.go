package util

import "fmt"

/**CFile****************************************************************

  FileName    [utilIsop.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [ISOP computation.]

  Synopsis    [ISOP computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - October 4, 2014.]

  Revision    [$Id: utilIsop.c,v 1.00 2014/10/04 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const ABC_ISOP_MAX_VAR =  16
#define ABC_ISOP_MAX_WORD (ABC_ISOP_MAX_VAR > 6 ? (1 << (ABC_ISOP_MAX_VAR-6)) : 1)
const ABC_ISOP_MAX_CUBE =  0xFFFF

typedef word FUNC_ISOP( word *, word *, word *, word, int * );

var Abc_Isop7Cover FUNC_ISOP; 
var Abc_Isop8Cover FUNC_ISOP;
var Abc_Isop9Cover FUNC_ISOP; 
var Abc_Isop10Cover FUNC_ISOP;
var Abc_Isop11Cover FUNC_ISOP;
var Abc_Isop12Cover FUNC_ISOP;
var Abc_Isop13Cover FUNC_ISOP;
var Abc_Isop14Cover FUNC_ISOP;
var Abc_Isop15Cover FUNC_ISOP;
var Abc_Isop16Cover FUNC_ISOP;

static FUNC_ISOP * s_pFuncIsopCover[17] =
{
    NULL, // 0
    NULL, // 1
    NULL, // 2
    NULL, // 3
    NULL, // 4
    NULL, // 5
    NULL, // 6
    Abc_Isop7Cover,  // 7
    Abc_Isop8Cover,  // 8
    Abc_Isop9Cover,  // 9
    Abc_Isop10Cover, // 10
    Abc_Isop11Cover, // 11
    Abc_Isop12Cover, // 12
    Abc_Isop13Cover, // 13
    Abc_Isop14Cover, // 14
    Abc_Isop15Cover, // 15
    Abc_Isop16Cover  // 16
};

// extern word Abc_IsopCheck( word * pOn, word * pOnDc, word * pRes, int nVars, word CostLim, int * pCover );

// extern word Abc_EsopCheck( word * pOn,                            int nVars, word CostLim, int * pCover );


func Abc_Cube2Cost(nCubes int32) word { return (word)nCubes << 32;       }
func Abc_CostCubes(Cost word) int32 { return int32((Cost >> 32));        }
func Abc_CostLits(Cost word) int32 { return int32((Cost & 0xFFFFFFFF)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [These procedures assume that function has exact support.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_IsopAddLits(pCover *int32, Cost0 word, Cost1 word, Var int32) {
    if pCover != nil {
        var c, nCubes0, nCubes1 int32;
        nCubes0 = Abc_CostCubes( Cost0 );
        nCubes1 = Abc_CostCubes( Cost1 );
        for  c = 0; c < nCubes0; c++  {
            pCover[c] |= (1 << Abc_Var2Lit(Var,0));
        }
        for  c = 0; c < nCubes1; c++  {
            pCover[nCubes0+c] |= (1 << Abc_Var2Lit(Var,1));
        }
    }
}
func Abc_Isop6Cover(uOn word, uOnDc word, pRes *word, nVars int32, CostLim word, pCover *int32) word {
    var uOn0, uOn1, uOnDc0, uOnDc1, uRes0, uRes1, uRes2 word;
    var Cost0, Cost1, Cost2 word;  var Var int32; 
    assert(nVars <= 6, "nVars <= 6");
    assert((uOn & ~uOnDc) == 0, "(uOn & ~uOnDc) == 0");
    if uOn == 0 {
        pRes[0] = 0;
        return 0;
    }
    if uOnDc == ~(word)0 {
        pRes[0] = ~(word)0;
        if pCover != nil { pCover[0] = 0; }
        return Abc_Cube2Cost(1);
    }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if Abc_Tt6HasVar( uOn, Var ) || Abc_Tt6HasVar( uOnDc, Var ) {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    uOn0   = Abc_Tt6Cofactor0( uOn,   Var );
    uOn1   = Abc_Tt6Cofactor1( uOn  , Var );
    uOnDc0 = Abc_Tt6Cofactor0( uOnDc, Var );
    uOnDc1 = Abc_Tt6Cofactor1( uOnDc, Var );
    // solve for cofactors
    Cost0 = Abc_Isop6Cover( uOn0 & ~uOnDc1, uOnDc0, &uRes0, Var, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    Cost1 = Abc_Isop6Cover( uOn1 & ~uOnDc0, uOnDc1, &uRes1, Var, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    Cost2 = Abc_Isop6Cover( (uOn0 & ~uRes0) | (uOn1 & ~uRes1), uOnDc0 & uOnDc1, &uRes2, Var, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    *pRes = uRes2 | (uRes0 & s_Truths6Neg[Var]) | (uRes1 & s_Truths6[Var]);
    assert((uOn & ~*pRes) == 0 && (*pRes & ~uOnDc) == 0, "(uOn & ~*pRes) == 0 && (*pRes & ~uOnDc) == 0");
    Abc_IsopAddLits( pCover, Cost0, Cost1, Var );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop7Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn0, uOn1, uOn2, uOnDc2, uRes0, uRes1, uRes2 word;
    var Cost0, Cost1, Cost2 word;  var nVars int32  = 6; 
    assert((pOn[0] & ~pOnDc[0]) == 0, "(pOn[0] & ~pOnDc[0]) == 0");
    assert((pOn[1] & ~pOnDc[1]) == 0, "(pOn[1] & ~pOnDc[1]) == 0");
    // cofactor
    uOn0 = pOn[0] & ~pOnDc[1];
    uOn1 = pOn[1] & ~pOnDc[0];
    // solve for cofactors
    Cost0 = Abc_IsopCheck( &uOn0, pOnDc,   &uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    Cost1 = Abc_IsopCheck( &uOn1, pOnDc+1, &uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    uOn2 = (pOn[0] & ~uRes0) | (pOn[1] & ~uRes1);
    uOnDc2 = pOnDc[0] & pOnDc[1];
    Cost2 = Abc_IsopCheck( &uOn2, &uOnDc2,  &uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    pRes[0] = uRes2 | uRes0;
    pRes[1] = uRes2 | uRes1;
    assert((pOn[0] & ~pRes[0]) == 0 && (pRes[0] & ~pOnDc[0]) == 0, "(pOn[0] & ~pRes[0]) == 0 && (pRes[0] & ~pOnDc[0]) == 0");
    assert((pOn[1] & ~pRes[1]) == 0 && (pRes[1] & ~pOnDc[1]) == 0, "(pOn[1] & ~pRes[1]) == 0 && (pRes[1] & ~pOnDc[1]) == 0");
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop8Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [2]word;
    var Cost0, Cost1, Cost2 word;  var nVars int32  = 7;
    // negative cofactor
    uOn2[0] = pOn[0] & ~pOnDc[2];
    uOn2[1] = pOn[1] & ~pOnDc[3];
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,   uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    uOn2[0] = pOn[2] & ~pOnDc[0];
    uOn2[1] = pOn[3] & ~pOnDc[1];
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+2, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    uOn2[0] = (pOn[0] & ~uRes0[0]) | (pOn[2] & ~uRes1[0]); uOnDc2[0] = pOnDc[0] & pOnDc[2];
    uOn2[1] = (pOn[1] & ~uRes0[1]) | (pOn[3] & ~uRes1[1]); uOnDc2[1] = pOnDc[1] & pOnDc[3];
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,  uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    pRes[0] = uRes2[0] | uRes0[0];
    pRes[1] = uRes2[1] | uRes0[1];
    pRes[2] = uRes2[0] | uRes1[0];
    pRes[3] = uRes2[1] | uRes1[1];
    assert((pOn[0] & ~pRes[0]) == 0 && (pOn[1] & ~pRes[1]) == 0 && (pOn[2] & ~pRes[2]) == 0 && (pOn[3] & ~pRes[3]) == 0, "(pOn[0] & ~pRes[0]) == 0 && (pOn[1] & ~pRes[1]) == 0 && (pOn[2] & ~pRes[2]) == 0 && (pOn[3] & ~pRes[3]) == 0");
    assert((pRes[0] & ~pOnDc[0])==0 && (pRes[1] & ~pOnDc[1])==0 && (pRes[2] & ~pOnDc[2])==0 && (pRes[3] & ~pOnDc[3])==0, "(pRes[0] & ~pOnDc[0])==0 && (pRes[1] & ~pOnDc[1])==0 && (pRes[2] & ~pOnDc[2])==0 && (pRes[3] & ~pOnDc[3])==0");
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop9Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [4]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 8;
var nWords int32 = 4;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop10Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [8]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 9;
var nWords int32 = 8;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop11Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [16]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 10;
var nWords int32 = 16;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop12Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [32]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 11;
var nWords int32 = 32;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop13Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [64]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 12;
var nWords int32 = 64;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop14Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [128]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 13;
var nWords int32 = 128;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop15Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [256]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 14;
var nWords int32 = 256;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_Isop16Cover(pOn *word, pOnDc *word, pRes *word, CostLim word, pCover *int32) word {
    var uOn2, uOnDc2, uRes0, uRes1, uRes2 [512]word;
    var Cost0, Cost1, Cost2 word; var c int32;
var nVars int32 = 15;
var nWords int32 = 512;
    // negative cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c] & ~pOnDc[c+nWords];
    }
    Cost0 = Abc_IsopCheck( uOn2, pOnDc,        uRes0, nVars, CostLim, pCover );
    if Cost0 >= CostLim { return CostLim; }
    // positive cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = pOn[c+nWords] & ~pOnDc[c];
    }
    Cost1 = Abc_IsopCheck( uOn2, pOnDc+nWords, uRes1, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    if Cost0 + Cost1 >= CostLim { return CostLim; }
    // middle cofactor
    for  c = 0; c < nWords; c++  {
        uOn2[c] = (pOn[c] & ~uRes0[c]) | (pOn[c+nWords] & ~uRes1[c]), uOnDc2[c] = pOnDc[c] & pOnDc[c+nWords];
    }
    Cost2 = Abc_IsopCheck( uOn2, uOnDc2,       uRes2, nVars, CostLim, pCover ? pCover + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1) : nil );
    if Cost0 + Cost1 + Cost2 >= CostLim { return CostLim; }
    // derive the final truth table
    for  c = 0; c < nWords; c++  {
        pRes[c] = uRes2[c] | uRes0[c], pRes[c+nWords] = uRes2[c] | uRes1[c];
    }
    // verify
    for  c = 0; c < (nWords<<1); c++  {
        assert((pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0, "(pOn[c]  & ~pRes[c] ) == 0 && (pRes[c] & ~pOnDc[c]) == 0");
    }
    Abc_IsopAddLits( pCover, Cost0, Cost1, nVars );
    return Cost0 + Cost1 + Cost2 + Abc_CostCubes(Cost0) + Abc_CostCubes(Cost1);
}
func Abc_IsopCheck(pOn *word, pOnDc *word, pRes *word, nVars int32, CostLim word, pCover *int32) word {
    var nVarsNew int32; var Cost word;
    if nVars <= 6 {
        return Abc_Isop6Cover( *pOn, *pOnDc, pRes, nVars, CostLim, pCover ); }
    for  nVarsNew = nVars; nVarsNew > 6; nVarsNew--  {
        if Abc_TtHasVar( pOn, nVars, nVarsNew-1 ) || Abc_TtHasVar( pOnDc, nVars, nVarsNew-1 ) {
            break; }
    }
    if nVarsNew == 6 {
        Cost = Abc_Isop6Cover( *pOn, *pOnDc, pRes, nVarsNew, CostLim, pCover );
    } else {
        Cost = s_pFuncIsopCover[nVarsNew]( pOn, pOnDc, pRes, CostLim, pCover ); }
    Abc_TtStretch6( pRes, nVarsNew, nVars );
    return Cost;
}

/**Function*************************************************************

  Synopsis    [Create truth table for the given cover.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_IsopTtElems_TtElems [ABC_ISOP_MAX_VAR+1][ABC_ISOP_MAX_WORD]word;
var spStatic_Abc_IsopTtElems_pTtElems [ABC_ISOP_MAX_VAR+1]*word = [ABC_ISOP_MAX_VAR+1]*word{nil};
func Abc_IsopTtElems() **word {
    
    if spStatic_Abc_IsopTtElems_pTtElems[0] == nil {
        var v int32;
        for  v = 0; v <= ABC_ISOP_MAX_VAR; v++  {
            spStatic_Abc_IsopTtElems_pTtElems[v] = spStatic_Abc_IsopTtElems_TtElems[v];
        }
        Abc_TtElemInit( spStatic_Abc_IsopTtElems_pTtElems, ABC_ISOP_MAX_VAR );
    }
    return spStatic_Abc_IsopTtElems_pTtElems;
}
func Abc_IsopBuildTruth(vCover *Vec_Int_t, nVars int32, pRes *word, fXor int32, fCompl int32) {
    var pTtElems **word  = Abc_IsopTtElems();
    var pCube [ABC_ISOP_MAX_WORD]word;
    var nWords int32  = Abc_TtWordNum( nVars );
    var c, v, Cube int32;
    assert(nVars <= ABC_ISOP_MAX_VAR, "nVars <= ABC_ISOP_MAX_VAR");
    Abc_TtClear( pRes, nWords );
    // inlining of macro Vec_IntForEachEntry
    for  c = 0; (c < Vec_IntSize(vCover)) && (((Cube) = Vec_IntEntry(vCover, c)), 1); c++  {
        Abc_TtFill( pCube, nWords );
        for  v = 0; v < nVars; v++  {
            if ((Cube >> (v << 1)) & 3) == 1 {
                Abc_TtSharp( pCube, pCube, pTtElems[v], nWords );
            } else if ((Cube >> (v << 1)) & 3) == 2 {
                Abc_TtAnd( pCube, pCube, pTtElems[v], nWords, 0 ); }
        }
        if fXor != 0 {
            Abc_TtXor( pRes, pRes, pCube, nWords, 0 );
        } else {
            Abc_TtOr( pRes, pRes, pCube, nWords ); }
    }
    if fCompl != 0 {
        Abc_TtNot( pRes, nWords ); }
}
func Abc_IsopVerify(pFunc *word, nVars int32, pRes *word, vCover *Vec_Int_t, fXor int32, fCompl int32) {
    Abc_IsopBuildTruth( vCover, nVars, pRes, fXor, fCompl );
    if !Abc_TtEqual( pFunc, pRes, Abc_TtWordNum(nVars) ) {
        fmt.Printf("Verification failed.\n"); }
//    else
//        printf( "Verification succeeded.\n" );
}

/**Function*************************************************************

  Synopsis    [This procedure assumes that function has exact support.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_Isop(pFunc *word, nVars int32, nCubeLim int32, vCover *Vec_Int_t, fTryBoth int32) int32 {
    var pRes [ABC_ISOP_MAX_WORD]word;
    var Cost0 word;
    var Cost1 word;
    var Cost word;
    var CostInit word = Abc_Cube2Cost(nCubeLim);
    assert(nVars <= ABC_ISOP_MAX_VAR, "nVars <= ABC_ISOP_MAX_VAR");
    Vec_IntGrow( vCover, 1 << (nVars-1) );
    if fTryBoth != 0 {
        Cost0 = Abc_IsopCheck( pFunc, pFunc, pRes, nVars, CostInit, nil );
        Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
        Cost1 = Abc_IsopCheck( pFunc, pFunc, pRes, nVars, Cost0, nil );
        Cost = Abc_MinWord( Cost0, Cost1 );
        if Cost == CostInit {
            Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
            return -1;
        }
        if Cost == Cost0 {
            Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
            Abc_IsopCheck( pFunc, pFunc, pRes, nVars, CostInit, Vec_IntArray(vCover) );
        
        } else { // if ( Cost == Cost1 )
        
            Abc_IsopCheck( pFunc, pFunc, pRes, nVars, CostInit, Vec_IntArray(vCover) );
            Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
        }
    
    } else {
        Cost = Cost0 = Abc_IsopCheck( pFunc, pFunc, pRes, nVars, CostInit, Vec_IntArray(vCover) );
        if Cost == CostInit {
            return -1; }
    }
    vCover.nSize = Abc_CostCubes(Cost);
    assert(vCover.nSize <= vCover.nCap, "vCover.nSize <= vCover.nCap");
//    Abc_IsopVerify( pFunc, nVars, pRes, vCover, 0, Cost != Cost0 );
    return Cost != Cost0;
}

/**Function*************************************************************

  Synopsis    [Compute CNF assuming it does not exceed the limit.]

  Description [Please note that pCover should have at least 32 extra entries!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_IsopCnf(pFunc *word, nVars int32, nCubeLim int32, pCover *int32) int32 {
    var pRes [ABC_ISOP_MAX_WORD]word;
    var Cost0 word;
    var Cost1 word;
    var CostInit word = Abc_Cube2Cost(nCubeLim);
    var c, nCubes0, nCubes1 int32;
    assert(nVars <= ABC_ISOP_MAX_VAR, "nVars <= ABC_ISOP_MAX_VAR");
    assert(Abc_TtHasVar( pFunc, nVars, nVars - 1 ), "Abc_TtHasVar( pFunc, nVars, nVars - 1 )");
    if nVars > 6 {
        Cost0 = s_pFuncIsopCover[nVars]( pFunc, pFunc, pRes, CostInit, pCover );
    } else {
        Cost0 = Abc_Isop6Cover( *pFunc, *pFunc, pRes, nVars, CostInit, pCover ); }
    if Cost0 >= CostInit {
        return CostInit; }
    Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
    if nVars > 6 {
        Cost1 = s_pFuncIsopCover[nVars]( pFunc, pFunc, pRes, CostInit, pCover ? pCover + Abc_CostCubes(Cost0) : nil );
    } else {
        Cost1 = Abc_Isop6Cover( *pFunc, *pFunc, pRes, nVars, CostInit, pCover ? pCover + Abc_CostCubes(Cost0) : nil ); }
    Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
    if Cost0 + Cost1 >= CostInit {
        return CostInit; }
    nCubes0 = Abc_CostCubes(Cost0);
    nCubes1 = Abc_CostCubes(Cost1);
    if pCover != nil {
        for  c = 0; c < nCubes0; c++  {
            pCover[c] |= (1 << Abc_Var2Lit(nVars, 0));
        }
        for  c = 0; c < nCubes1; c++  {
            pCover[c+nCubes0] |= (1 << Abc_Var2Lit(nVars, 1));
        }
    }
    return nCubes0 + nCubes1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_IsopCountLits(vCover *Vec_Int_t, nVars int32) int32 {
    var i, k, Entry, Literal, nLits int32;
    if Vec_IntSize(vCover) == 0 || (Vec_IntSize(vCover) == 1 && Vec_IntEntry(vCover, 0) == 0) {
        return 0; }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCover)) && (((Entry) = Vec_IntEntry(vCover, i)), 1); i++  { 
        for  k = 0; k < nVars; k++  {
            Literal = 3 & (Entry >> (k << 1));
            if Literal == 1 { // neg literal
                nLits++;
            } else if Literal == 2 { // pos literal
                nLits++;
            } else if Literal != 0 { 
                assert(false, "0"); }
        }
    }
    return nLits;
}
func Abc_IsopPrintCover(vCover *Vec_Int_t, nVars int32, fCompl int32) {
    var i, k, Entry, Literal int32;
    if Vec_IntSize(vCover) == 0 || (Vec_IntSize(vCover) == 1 && Vec_IntEntry(vCover, 0) == 0) {
        fmt.Printf("Constant %d\n", Vec_IntSize(vCover));
        return;
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCover)) && (((Entry) = Vec_IntEntry(vCover, i)), 1); i++  { 
        for  k = 0; k < nVars; k++  {
            Literal = 3 & (Entry >> (k << 1));
            if Literal == 1 { // neg literal
                fmt.Printf("0");
            } else if Literal == 2 { // pos literal
                fmt.Printf("1");
            } else if Literal == 0 { 
                fmt.Printf("-");
            } else { assert(false, "0"); }
        }
        fmt.Printf(" %d\n", !fCompl);
    }
}
func Abc_IsopPrint(t *word, nVars int32, vCover *Vec_Int_t, fTryBoth int32) {
    var fCompl int32  = Abc_Isop( t, nVars, ABC_ISOP_MAX_CUBE, vCover, fTryBoth );
    Abc_IsopPrintCover( vCover, nVars, fCompl );
}


/**Function*************************************************************

  Synopsis    [These procedures assume that function has exact support.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_EsopAddLits(pCover *int32, r0 word, r1 word, r2 word, Max word, Var int32) int32 {
    var i, c0, c1, c2 int32;
    if Max == r0 {
        c2 = Abc_CostCubes(r2);
        if pCover != nil {
            c0 = Abc_CostCubes(r0);
            c1 = Abc_CostCubes(r1);
            for  i = 0; i < c1; i++  {
                pCover[i] = pCover[c0+i];
            }
            for  i = 0; i < c2; i++  {
                pCover[c1+i] = pCover[c0+c1+i] | (1 << Abc_Var2Lit(Var,0));
            }
        }
        return c2;
    
    } else if Max == r1 {
        c2 = Abc_CostCubes(r2);
        if pCover != nil {
            c0 = Abc_CostCubes(r0);
            c1 = Abc_CostCubes(r1);
            for  i = 0; i < c2; i++  {
                pCover[c0+i] = pCover[c0+c1+i] | (1 << Abc_Var2Lit(Var,1));
            }
        }
        return c2;
    
    } else {
        c0 = Abc_CostCubes(r0);
        c1 = Abc_CostCubes(r1);
        if pCover != nil {
            c2 = Abc_CostCubes(r2);
            for  i = 0; i < c0; i++  {
                pCover[i] |= (1 << Abc_Var2Lit(Var,0));
            }
            for  i = 0; i < c1; i++  {
                pCover[c0+i] |= (1 << Abc_Var2Lit(Var,1));
            }
        }
        return c0 + c1;
    }
}
func Abc_Esop6Cover(t word, nVars int32, CostLim word, pCover *int32) word {
    var c0, c1 word;
    var r0, r1, r2, Max word;
    var Var int32;
    assert(nVars <= 6, "nVars <= 6");
    if t == 0 {
        return 0; }
    if t == ~(word)0 {
        if pCover != nil { *pCover = 0; }
        return Abc_Cube2Cost(1);
    }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if Abc_Tt6HasVar( t, Var ) {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    c0 = Abc_Tt6Cofactor0( t, Var );
    c1 = Abc_Tt6Cofactor1( t, Var );
    // call recursively
    r0 = Abc_Esop6Cover( c0,      Var, CostLim, pCover ? pCover : nil );
    if r0 >= CostLim { return CostLim; }
    r1 = Abc_Esop6Cover( c1,      Var, CostLim, pCover ? pCover + Abc_CostCubes(r0) : nil );
    if r1 >= CostLim { return CostLim; }
    r2 = Abc_Esop6Cover( c0 ^ c1, Var, CostLim, pCover ? pCover + Abc_CostCubes(r0) + Abc_CostCubes(r1) : nil );
    if r2 >= CostLim { return CostLim; }
    Max = Abc_MaxWord( r0, Abc_MaxWord(r1, r2) );
    if r0 + r1 + r2 - Max >= CostLim { return CostLim; }
    return r0 + r1 + r2 - Max + Abc_EsopAddLits( pCover, r0, r1, r2, Max, Var );
}
func Abc_EsopCover(pOn *word, nVars int32, CostLim word, pCover *int32) word {
    var r0, r1, r2, Max word;
    var c int32;
    var nWords int32 = (1 << (nVars - 7));
    assert(nVars > 6, "nVars > 6");
    assert(Abc_TtHasVar( pOn, nVars, nVars - 1 ), "Abc_TtHasVar( pOn, nVars, nVars - 1 )");
    r0 = Abc_EsopCheck( pOn,        nVars-1, CostLim, pCover );
    if r0 >= CostLim { return CostLim; }
    r1 = Abc_EsopCheck( pOn+nWords, nVars-1, CostLim, pCover ? pCover + Abc_CostCubes(r0) : nil );
    if r1 >= CostLim { return CostLim; }
    for  c = 0; c < nWords; c++  {
        pOn[c] ^= pOn[nWords+c];
    }
    r2 = Abc_EsopCheck( pOn, nVars-1, CostLim, pCover ? pCover + Abc_CostCubes(r0) + Abc_CostCubes(r1) : nil );
    for  c = 0; c < nWords; c++  {
        pOn[c] ^= pOn[nWords+c];
    }
    if r2 >= CostLim { return CostLim; }
    Max = Abc_MaxWord( r0, Abc_MaxWord(r1, r2) );
    if r0 + r1 + r2 - Max >= CostLim { return CostLim; }
    return r0 + r1 + r2 - Max + Abc_EsopAddLits( pCover, r0, r1, r2, Max, nVars-1 );
}
func Abc_EsopCheck(pOn *word, nVars int32, CostLim word, pCover *int32) word {
    var nVarsNew int32; var Cost word;
    if nVars <= 6 {
        return Abc_Esop6Cover( *pOn, nVars, CostLim, pCover ); }
    for  nVarsNew = nVars; nVarsNew > 6; nVarsNew--  {
        if Abc_TtHasVar( pOn, nVars, nVarsNew-1 ) {
            break; }
    }
    if nVarsNew == 6 {
        Cost = Abc_Esop6Cover( *pOn, nVarsNew, CostLim, pCover );
    } else {
        Cost = Abc_EsopCover( pOn, nVarsNew, CostLim, pCover ); }
    return Cost;
}


/**Function*************************************************************

  Synopsis    [Perform ISOP computation by subtracting cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_TtIntersect2(pIn1 *word, pIn2 *word, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if pIn1[w] & pIn2[w] {
            return 1; }
    }
    return 0;
}
func Abc_TtCheckWithCubePos2Neg(t *word, c *word, nWords int32, iVar int32) int32 {
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            if t[i] & (c[i] >> Shift) {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[Step+i] & c[i] {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtCheckWithCubeNeg2Pos(t *word, c *word, nWords int32, iVar int32) int32 {
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            if t[i] & (c[i] << Shift) {
                return 0; }
        }
        return 1;
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                if t[i] & c[Step+i] {
                    return 0; }
            }
        }
        return 1;
    }
}
func Abc_TtExpandCubePos2Neg(t *word, nWords int32, iVar int32) {
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            t[i] |= (t[i] >> Shift);
        }
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                t[i] = t[Step+i];
            }
        }
    }
}
func Abc_TtExpandCubeNeg2Pos(t *word, nWords int32, iVar int32) {
    if iVar < 6 {
        var i int32;
        var Shift int32 = (1 << iVar);
        for  i = 0; i < nWords; i++  {
            t[i] |= (t[i] << Shift);
        }
    
    } else {
        var i int32;
        var Step int32 = (1 << (iVar - 6));
        var tLimit *word  = t + nWords;
        for  ; t < tLimit; t += 2*Step  {
            for  i = 0; i < Step; i++  {
                t[Step+i] = t[i];
            }
        }
    }
}
func Abc_IsopNew(pOn *word, pOnDc *word, pRes *word, nVars int32, CostLim word, pCover *int32) word {
    var pCube [ABC_ISOP_MAX_WORD]word;
    var pOnset [ABC_ISOP_MAX_WORD]word;
    var pOffset [ABC_ISOP_MAX_WORD]word;
    var pBlocks [16]int32;
    var nBlocks int32;
    var vTwo int32;
    var uTwo int32;
    var nWords int32  = Abc_TtWordNum(nVars);
    var c, v, u, iMint, Cube, nLits int32;
    assert(nVars <= ABC_ISOP_MAX_VAR, "nVars <= ABC_ISOP_MAX_VAR");
    Abc_TtClear( pRes, nWords );
    Abc_TtCopy( pOnset, pOn, nWords, 0 ); 
    Abc_TtCopy( pOffset, pOnDc, nWords, 1 ); 
    if nVars < 6 {
        pOnset[0] >>= (64 - (1 << nVars)); }
    for  c = 0; !Abc_TtIsConst0(pOnset, nWords); c++  {
        // pick one minterm
        iMint = Abc_TtFindFirstBit(pOnset, nVars);
        for  Cube = v = 0; v < nVars; v++  {
            Cube |= (1 << Abc_Var2Lit(v, (iMint >> v) & 1));
        }
        // check expansion for the minterm
        nBlocks = 0;
        for  v = 0; v < nVars; v++  {
            if pBlocks[v] = Abc_TtGetBit(pOffset, iMint ^ (1 << v)) {
                nBlocks++; }
        }
        // check second step
        if nBlocks == nVars { // cannot expand
        
            Abc_TtSetBit( pRes, iMint );
            Abc_TtXorBit( pOnset, iMint );
            pCover[c] = Cube;
            nLits += nVars;
            continue;
        }
        // check dual expansion
        vTwo = uTwo = -1;
        if nBlocks < nVars - 1 {
            for  v = 0; v < nVars && vTwo == -1; v++  {
            if !pBlocks[v] {
            for  u = v + 1; u < nVars; u++  {
            if !pBlocks[u] {
                if Abc_TtGetBit( pOffset, iMint ^ (1 << v) ^ (1 << u) ) {
                    continue; }
                // can expand both directions
                vTwo = v;
                uTwo = u;
                break;
            }
            } }
            }
        }
        if vTwo == -1 { // can expand only one
        
            for  v = 0; v < nVars; v++  {
                if !pBlocks[v] {
                    break; }
            }
            Abc_TtSetBit( pRes, iMint );
            Abc_TtSetBit( pRes, iMint ^ (1 << v) );
            Abc_TtXorBit( pOnset, iMint );
            if Abc_TtGetBit(pOnset, iMint ^ (1 << v)) {
                Abc_TtXorBit( pOnset, iMint ^ (1 << v) ); }
            pCover[c] = Cube & ~(3 << Abc_Var2Lit(v, 0));
            nLits += nVars - 1;
            continue;
        }
        if nBlocks == nVars - 2 && vTwo >= 0 { // can expand only these two
        
            Abc_TtSetBit( pRes, iMint );
            Abc_TtSetBit( pRes, iMint ^ (1 << vTwo) );
            Abc_TtSetBit( pRes, iMint ^ (1 << uTwo) );
            Abc_TtSetBit( pRes, iMint ^ (1 << vTwo) ^ (1 << uTwo) );
            Abc_TtXorBit( pOnset, iMint );
            if Abc_TtGetBit(pOnset, iMint ^ (1 << vTwo)) {
                Abc_TtXorBit( pOnset, iMint ^ (1 << vTwo) ); }
            if Abc_TtGetBit(pOnset, iMint ^ (1 << uTwo)) {
                Abc_TtXorBit( pOnset, iMint ^ (1 << uTwo) ); }
            if Abc_TtGetBit(pOnset, iMint ^ (1 << vTwo) ^ (1 << uTwo)) {
                Abc_TtXorBit( pOnset, iMint ^ (1 << vTwo) ^ (1 << uTwo) ); }
            pCover[c] = Cube & ~(3 << Abc_Var2Lit(vTwo, 0)) & ~(3 << Abc_Var2Lit(uTwo, 0));
            nLits += nVars - 2;
            continue;
        }
        // can expand others as well
        Abc_TtClear( pCube, nWords );
        Abc_TtSetBit( pCube, iMint );
        Abc_TtSetBit( pCube, iMint ^ (1 << vTwo) );
        Abc_TtSetBit( pCube, iMint ^ (1 << uTwo) );
        Abc_TtSetBit( pCube, iMint ^ (1 << vTwo) ^ (1 << uTwo) );
        Cube &= ~(3 << Abc_Var2Lit(vTwo, 0)) & ~(3 << Abc_Var2Lit(uTwo, 0));
        assert(!(Abc_TtIntersect2(pCube, pOffset, nWords) != 0), "!Abc_TtIntersect2(pCube, pOffset, nWords)");        
        // expand against offset
        for  v = 0; v < nVars; v++  {
        if v != vTwo && v != uTwo {
            var Shift int32  = Abc_Var2Lit( v, 0 );
            if (Cube >> Shift) == 2 && (Abc_TtCheckWithCubePos2Neg(pOffset, pCube, nWords, v) != 0) { // pos literal
            
                Abc_TtExpandCubePos2Neg( pCube, nWords, v );
                Cube &= ~(3 << Shift);
            
            } else if (Cube >> Shift) == 1 && (Abc_TtCheckWithCubeNeg2Pos(pOffset, pCube, nWords, v) != 0) { // neg literal
            
                Abc_TtExpandCubeNeg2Pos( pCube, nWords, v );
                Cube &= ~(3 << Shift);
            
            } else { 
                nLits++; }
        }
        }
        // add cube to solution
        Abc_TtOr( pRes, pRes, pCube, nWords );
        Abc_TtSharp( pOnset, pOnset, pCube, nWords );
        pCover[c] = Cube;
    }
    pRes[0] = Abc_Tt6Stretch( pRes[0], nVars );
    return Abc_Cube2Cost(c) | nLits;
}
func Abc_IsopTestNew() {
    var nVars int32  = 4;
    var vCover *Vec_Int_t  = Vec_IntAlloc( 1000 );
    var r word;
    var t word = (s_Truths6[0] & s_Truths6[1]) ^ (s_Truths6[2] & s_Truths6[3]);
    var copy word = t;
//    word r, t = ~s_Truths6[0] | (s_Truths6[1] & s_Truths6[2] & s_Truths6[3]), copy = t;
//    word r, t = 0xABCDABCDABCDABCD, copy = t;
//    word r, t = 0x6996000000006996, copy = t;
//    word Cost = Abc_IsopNew( &t, &t, &r, nVars, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), Vec_IntArray(vCover) );
    var Cost word  = Abc_EsopCheck( &t, nVars, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), Vec_IntArray(vCover) );
    vCover.nSize = Abc_CostCubes(Cost);
    assert(vCover.nSize <= vCover.nCap, "vCover.nSize <= vCover.nCap");
    fmt.Printf("Cubes = %d.  Lits = %d.\n", Abc_CostCubes(Cost), Abc_CostLits(Cost));
    Abc_IsopPrintCover( vCover, nVars, 0 );
    Abc_IsopVerify( &copy, nVars, &r, vCover, 1, 0 );
    Vec_IntFree( vCover );
}

/**Function*************************************************************

  Synopsis    [Compute CNF assuming it does not exceed the limit.]

  Description [Please note that pCover should have at least 32 extra entries!]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_IsopTest_TotalCost [6]word ;
var spStatic_Abc_IsopTest_TotalTime [6]abctime ;
var spStatic_Abc_IsopTest_Counter int32;
func Abc_IsopTest(pFunc *word, nVars int32, vCover *Vec_Int_t) int32 {
    var fVerbose int32;    
    
    
    
    var pRes [ABC_ISOP_MAX_WORD]word;
    var Cost word;
    var clk abctime;
    spStatic_Abc_IsopTest_Counter++;
    if spStatic_Abc_IsopTest_Counter == 9999 {
        Abc_PrintTime( 1, "0", spStatic_Abc_IsopTest_TotalTime[0] );
        Abc_PrintTime( 1, "1", spStatic_Abc_IsopTest_TotalTime[1] );
        Abc_PrintTime( 1, "2", spStatic_Abc_IsopTest_TotalTime[2] );
        Abc_PrintTime( 1, "3", spStatic_Abc_IsopTest_TotalTime[3] );
        Abc_PrintTime( 1, "4", spStatic_Abc_IsopTest_TotalTime[4] );
        Abc_PrintTime( 1, "5", spStatic_Abc_IsopTest_TotalTime[5] );
    }
    assert(nVars <= ABC_ISOP_MAX_VAR, "nVars <= ABC_ISOP_MAX_VAR");
//    if ( fVerbose )
//    Dau_DsdPrintFromTruth( pFunc, nVars ), printf( "         " );

    clk = Abc_Clock();
    Cost = Abc_IsopCheck( pFunc, pFunc, pRes, nVars, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), Vec_IntArray(vCover) );
    vCover.nSize = Abc_CostCubes(Cost);
    assert(vCover.nSize <= vCover.nCap, "vCover.nSize <= vCover.nCap");
    if fVerbose != 0 {
    fmt.Printf("%5d %7d  ", Abc_CostCubes(Cost), Abc_CostLits(Cost)); }
//    Abc_IsopVerify( pFunc, nVars, pRes, vCover, 0, 0 );
    spStatic_Abc_IsopTest_TotalCost[0] += Abc_CostCubes(Cost);
    spStatic_Abc_IsopTest_TotalTime[0] += Abc_Clock() - clk;


    clk = Abc_Clock();
    Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
    Cost = Abc_IsopCheck( pFunc, pFunc, pRes, nVars, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), Vec_IntArray(vCover) );
    Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
    vCover.nSize = Abc_CostCubes(Cost);
    assert(vCover.nSize <= vCover.nCap, "vCover.nSize <= vCover.nCap");
    if fVerbose != 0 {
    fmt.Printf("%5d %7d  ", Abc_CostCubes(Cost), Abc_CostLits(Cost)); }
//    Abc_IsopVerify( pFunc, nVars, pRes, vCover, 0, 1 );
    spStatic_Abc_IsopTest_TotalCost[1] += Abc_CostCubes(Cost);
    spStatic_Abc_IsopTest_TotalTime[1] += Abc_Clock() - clk;

/*
    clk = Abc_Clock();
    Cost = Abc_EsopCheck( pFunc, nVars, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), Vec_IntArray(vCover) );
    vCover->nSize = Abc_CostCubes(Cost);
    assert( vCover->nSize <= vCover->nCap );
    if ( fVerbose )
    printf( "%5d %7d   ", Abc_CostCubes(Cost), Abc_CostLits(Cost) );
//    Abc_IsopVerify( pFunc, nVars, pRes, vCover, 1, 0 );
    TotalCost[2] += Abc_CostCubes(Cost);
    TotalTime[2] += Abc_Clock() - clk;


    clk = Abc_Clock();
    Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
    Cost = Abc_EsopCheck( pFunc, nVars, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), Vec_IntArray(vCover) );
    Abc_TtNot( pFunc, Abc_TtWordNum(nVars) );
    vCover->nSize = Abc_CostCubes(Cost);
    assert( vCover->nSize <= vCover->nCap );
    if ( fVerbose )
    printf( "%5d %7d   ", Abc_CostCubes(Cost), Abc_CostLits(Cost) );
//    Abc_IsopVerify( pFunc, nVars, pRes, vCover, 1, 1 );
    TotalCost[3] += Abc_CostCubes(Cost);
    TotalTime[3] += Abc_Clock() - clk;
*/

/*
    // try new computation
    clk = Abc_Clock();
    Cost = Abc_IsopNew( pFunc, pFunc, pRes, nVars, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), Vec_IntArray(vCover) );
    vCover->nSize = Abc_CostCubes(Cost);
    assert( vCover->nSize <= vCover->nCap );
    if ( fVerbose )
    printf( "%5d %7d   ", Abc_CostCubes(Cost), Abc_CostLits(Cost) );
    Abc_IsopVerify( pFunc, nVars, pRes, vCover, 0, 0 );
    TotalCost[4] += Abc_CostCubes(Cost);
    TotalTime[4] += Abc_Clock() - clk;
*/
/*
    // try old computation
    clk = Abc_Clock();
    Cost = Kit_TruthIsop( (unsigned *)pFunc, nVars, vCover, 1 );
    vCover->nSize = Vec_IntSize(vCover);
    assert( vCover->nSize <= vCover->nCap );
    if ( fVerbose )
    printf( "%5d %7d   ", Vec_IntSize(vCover), Abc_IsopCountLits(vCover, nVars) );
    TotalCost[4] += Vec_IntSize(vCover);
    TotalTime[4] += Abc_Clock() - clk;
*/

    clk = Abc_Clock();
    Cost = Abc_Isop( pFunc, nVars, ABC_ISOP_MAX_CUBE, vCover, 1 );
    if fVerbose != 0 {
    fmt.Printf("%5d %7d   ", Vec_IntSize(vCover), Abc_IsopCountLits(vCover, nVars)); }
    spStatic_Abc_IsopTest_TotalCost[5] += Vec_IntSize(vCover);
    spStatic_Abc_IsopTest_TotalTime[5] += Abc_Clock() - clk;

    if fVerbose != 0 {
    fmt.Printf("  | %8d %8d %8d %8d %8d %8d", int32(spStatic_Abc_IsopTest_TotalCost[0]), int32(spStatic_Abc_IsopTest_TotalCost[1]), int32(spStatic_Abc_IsopTest_TotalCost[2]), int32(spStatic_Abc_IsopTest_TotalCost[3]), int32(spStatic_Abc_IsopTest_TotalCost[4]), int32(spStatic_Abc_IsopTest_TotalCost[5])); }
    if fVerbose != 0 {
    fmt.Printf("\n"); }
//Abc_IsopPrintCover( vCover, nVars, 0 );
    return 1; 
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

