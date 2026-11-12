package dau

import "fmt"

/**CFile****************************************************************

  FileName    [dauGia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware unmapping.]

  Synopsis    [Coverting DSD into GIA.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: dauGia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dauInt.h"
//#include "aig/gia/gia.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );


const DAU_DSD_MAX_VAR = 12

var m_Calls int32;
var m_NonDsd int32;
var m_Non1Step int32;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Derives GIA for the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdToGiaCompose_rec(pGia *Gia_Man_t, Func word, pFanins *int32, nVars int32) int32 {
    var t0, t1 int32;
    if Func == 0 {
        return 0; }
    if Func == ~(word)0 {
        return 1; }
    assert(nVars > 0, "nVars > 0");
    if nVars--; nVars == 0 {
        assert(Func == s_Truths6[0] || Func == s_Truths6Neg[0], "Func == s_Truths6[0] || Func == s_Truths6Neg[0]");
        return Abc_LitNotCond( pFanins[0], int32((Func == s_Truths6Neg[0])) );
    }
    if !Abc_Tt6HasVar(Func, nVars) {
        return Dau_DsdToGiaCompose_rec( pGia, Func, pFanins, nVars ); }
    t0 = Dau_DsdToGiaCompose_rec( pGia, Abc_Tt6Cofactor0(Func, nVars), pFanins, nVars );
    t1 = Dau_DsdToGiaCompose_rec( pGia, Abc_Tt6Cofactor1(Func, nVars), pFanins, nVars );
    if pGia.pMuxes {
        return Gia_ManHashMuxReal( pGia, pFanins[nVars], t1, t0 );
    } else {
        return Gia_ManHashMux( pGia, pFanins[nVars], t1, t0 ); }
}

/**Function*************************************************************

  Synopsis    [Derives GIA for the DSD formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdToGia2_rec(pGia *Gia_Man_t, pStr string, p *string, pMatches *int32, pLits *int32, vCover *Vec_Int_t) int32 {
    var fCompl int32;
    if **p == '!' {
        (*p)++, fCompl = 1; }
    if **p >= 'a' && **p < 'a' + DAU_DSD_MAX_VAR { // var
        return Abc_LitNotCond( pLits[**p - 'a'], fCompl ); }
    if **p == '(' { // and/or
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var Res int32 = 1;
        var Lit int32;
        assert(**p == '(' && *q == ')', "**p == '(' && *q == ')'");
        for  (*p)++; *p < q; (*p)++  {
            Lit = Dau_DsdToGia2_rec( pGia, pStr, p, pMatches, pLits, vCover );
            Res = Gia_ManHashAnd( pGia, Res, Lit );
        }
        assert(*p == q, "*p == q");
        return Abc_LitNotCond( Res, fCompl );
    }
    if **p == '[' { // xor
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var Res, Lit int32;
        assert(**p == '[' && *q == ']', "**p == '[' && *q == ']'");
        for  (*p)++; *p < q; (*p)++  {
            Lit = Dau_DsdToGia2_rec( pGia, pStr, p, pMatches, pLits, vCover );
            if pGia.pMuxes {
                Res = Gia_ManHashXorReal( pGia, Res, Lit );
            } else {
                Res = Gia_ManHashXor( pGia, Res, Lit ); }
        }
        assert(*p == q, "*p == q");
        return Abc_LitNotCond( Res, fCompl );
    }
    if **p == '<' { // mux
    
        var nVars int32;
        var Temp [3]int32;
        var pTemp *int32 = Temp;
        var Res int32;
        var Fanins [DAU_DSD_MAX_VAR]int32;
        var pLits2 *int32;
        var pOld string  = *p;
        var q string  = pStr + pMatches[ *p - pStr ];
        // read fanins
        if *(q+1) == '{' {
            var q2 string;
            *p = q+1;
            q2 = pStr + pMatches[ *p - pStr ];
            assert(**p == '{' && *q2 == '}', "**p == '{' && *q2 == '}'");
            for  nVars = 0, (*p)++; *p < q2; (*p)++, nVars++  {
                Fanins[nVars] = Dau_DsdToGia2_rec( pGia, pStr, p, pMatches, pLits, vCover );
            }
            assert(*p == q2, "*p == q2");
            pLits2 = Fanins;
        
        } else {
            pLits2 = pLits; }
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
        // derive MUX components
        for  (*p)++; *p < q; (*p)++  {
            *pTemp++ = Dau_DsdToGia2_rec( pGia, pStr, p, pMatches, pLits2, vCover );
        }
        assert(pTemp == Temp + 3, "pTemp == Temp + 3");
        assert(*p == q, "*p == q");
        if *(q+1) == '{' { // and/or
        
            var q string  = pStr + pMatches[ ++(*p) - pStr ];
            assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
            *p = q;
        }
        if pGia.pMuxes {
            Res = Gia_ManHashMuxReal( pGia, Temp[0], Temp[1], Temp[2] );
        } else {
            Res = Gia_ManHashMux( pGia, Temp[0], Temp[1], Temp[2] ); }
        return Abc_LitNotCond( Res, fCompl );
    }
    if (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9') {
        var vLeaves Vec_Int_t;  var q string;
        var pFunc [DAU_DSD_MAX_VAR > 6 ? (1 << (DAU_DSD_MAX_VAR-6)) : 1]word;
        var Fanins [DAU_DSD_MAX_VAR]int32;
        var Res int32; 
        var i int32;
        var nVars int32 = Abc_TtReadHex( pFunc, *p );
        *p += Abc_TtHexDigitNum( nVars );
        q = pStr + pMatches[ *p - pStr ];
        assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
        for  i = 0, (*p)++; *p < q; (*p)++, i++  {
            Fanins[i] = Dau_DsdToGia2_rec( pGia, pStr, p, pMatches, pLits, vCover );
        }
        assert(i == nVars, "i == nVars");
        assert(*p == q, "*p == q");
//        Res = Dau_DsdToGia2Compose_rec( pGia, Func, Fanins, nVars );
        vLeaves.nCap = nVars;
        vLeaves.nSize = nVars;
        vLeaves.pArray = Fanins;        
        Res = Kit_TruthToGia( pGia, (unsigned *)pFunc, nVars, vCover, &vLeaves, 1 );
        m_Non1Step++;
        return Abc_LitNotCond( Res, fCompl );
    }
    assert(false, "0");
    return 0;
}
func Dau_DsdToGia2(pGia *Gia_Man_t, p string, pLits *int32, vCover *Vec_Int_t) int32 {
    var Res int32;
    if *p == '0' && *(p+1) == 0 {
        Res = 0;
    } else if *p == '1' && *(p+1) == 0 {
        Res = 1;
    } else {
        Res = Dau_DsdToGia2_rec( pGia, p, &p, Dau_DsdComputeMatches(p), pLits, vCover ); }
    assert(*++p == 0, "*++p == 0");
    return Res;
}

/**Function*************************************************************

  Synopsis    [Derives GIA for the DSD formula.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dau_DsdAddToArray(pGia *Gia_Man_t, pFans *int32, nFans int32, iFan int32) {
    var i int32;
    pFans[nFans] = iFan;
    if nFans == 0 {
        return; }
    for  i = nFans; i > 0; i--  {
        if Gia_ObjLevelId(pGia, Abc_Lit2Var(pFans[i])) <= Gia_ObjLevelId(pGia, Abc_Lit2Var(pFans[i-1])) {
            return; }
        // inlining of macro ABC_SWAP
        { var t int32  = pFans[i]; pFans[i] = pFans[i-1]; pFans[i-1] = t; };
    }
}
func Dau_DsdBalance(pGia *Gia_Man_t, pFans *int32, nFans int32, fAnd int32) int32 {
    var pObj *Gia_Obj_t;
    var iFan0, iFan1, iFan int32;
    if nFans == 1 {
        return pFans[0]; }
    assert(nFans > 1, "nFans > 1");
    iFan0 = pFans[--nFans];
    iFan1 = pFans[--nFans];
    if Vec_IntSize(&pGia.vHTable) == 0 {
        if fAnd != 0 {
            iFan = Gia_ManAppendAnd2( pGia, iFan0, iFan1 );
        } else if pGia.pMuxes {
            var fCompl int32  = Abc_LitIsCompl(iFan0) ^ Abc_LitIsCompl(iFan1);
            iFan = Gia_ManAppendXorReal( pGia, Abc_LitRegular(iFan0), Abc_LitRegular(iFan1) );
            iFan = Abc_LitNotCond( iFan, fCompl );
        
        } else { 
            iFan = Gia_ManAppendXor2( pGia, iFan0, iFan1 ); }
    
    } else {
        if fAnd != 0 {
            iFan = Gia_ManHashAnd( pGia, iFan0, iFan1 );
        } else if pGia.pMuxes {
            iFan = Gia_ManHashXorReal( pGia, iFan0, iFan1 );
        } else { 
            iFan = Gia_ManHashXor( pGia, iFan0, iFan1 ); }
    }
    pObj = Gia_ManObj(pGia, Abc_Lit2Var(iFan));
    if Gia_ObjIsAnd(pObj) {
        if fAnd != 0 {
            Gia_ObjSetAndLevel( pGia, pObj );
        } else if pGia.pMuxes {
            Gia_ObjSetXorLevel( pGia, pObj );
        } else {
            if Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) {
                Gia_ObjSetAndLevel( pGia, Gia_ObjFanin0(pObj) ); }
            if Gia_ObjIsAnd(Gia_ObjFanin1(pObj)) {
                Gia_ObjSetAndLevel( pGia, Gia_ObjFanin1(pObj) ); }
            Gia_ObjSetAndLevel( pGia, pObj );
        }
    }
    Dau_DsdAddToArray( pGia, pFans, nFans++, iFan );
    return Dau_DsdBalance( pGia, pFans, nFans, fAnd );
}
func Dau_DsdToGia_rec(pGia *Gia_Man_t, pStr string, p *string, pMatches *int32, pLits *int32, vCover *Vec_Int_t) int32 {
    var fCompl int32;
    if **p == '!' {
        (*p)++, fCompl = 1; }
    if **p >= 'a' && **p < 'a' + DAU_DSD_MAX_VAR { // var
        return Abc_LitNotCond( pLits[**p - 'a'], fCompl ); }
    if **p == '(' { // and/or
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var pFans [DAU_DSD_MAX_VAR]int32;
        var nFans int32;
        var Fan int32;
        assert(**p == '(' && *q == ')', "**p == '(' && *q == ')'");
        for  (*p)++; *p < q; (*p)++  {
            Fan = Dau_DsdToGia_rec( pGia, pStr, p, pMatches, pLits, vCover );
            Dau_DsdAddToArray( pGia, pFans, nFans++, Fan );
        }
        Fan = Dau_DsdBalance( pGia, pFans, nFans, 1 );
        assert(*p == q, "*p == q");
        return Abc_LitNotCond( Fan, fCompl );
    }
    if **p == '[' { // xor
    
        var q string  = pStr + pMatches[ *p - pStr ];
        var pFans [DAU_DSD_MAX_VAR]int32;
        var nFans int32;
        var Fan int32;
        assert(**p == '[' && *q == ']', "**p == '[' && *q == ']'");
        for  (*p)++; *p < q; (*p)++  {
            Fan = Dau_DsdToGia_rec( pGia, pStr, p, pMatches, pLits, vCover );
            Dau_DsdAddToArray( pGia, pFans, nFans++, Fan );
        }
        Fan = Dau_DsdBalance( pGia, pFans, nFans, 0 );
        assert(*p == q, "*p == q");
        return Abc_LitNotCond( Fan, fCompl );
    }
    if **p == '<' { // mux
    
        var pObj *Gia_Obj_t;
        var nVars int32;
        var Temp [3]int32;
        var pTemp *int32 = Temp;
        var Res int32;
        var Fanins [DAU_DSD_MAX_VAR]int32;
        var pLits2 *int32;
        var pOld string  = *p;
        var q string  = pStr + pMatches[ *p - pStr ];
        // read fanins
        if *(q+1) == '{' {
            var q2 string;
            *p = q+1;
            q2 = pStr + pMatches[ *p - pStr ];
            assert(**p == '{' && *q2 == '}', "**p == '{' && *q2 == '}'");
            for  nVars = 0, (*p)++; *p < q2; (*p)++, nVars++  {
                Fanins[nVars] = Dau_DsdToGia_rec( pGia, pStr, p, pMatches, pLits, vCover );
            }
            assert(*p == q2, "*p == q2");
            pLits2 = Fanins;
        
        } else {
            pLits2 = pLits; }
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
        // derive MUX components
        for  (*p)++; *p < q; (*p)++  {
            *pTemp++ = Dau_DsdToGia_rec( pGia, pStr, p, pMatches, pLits2, vCover );
        }
        assert(pTemp == Temp + 3, "pTemp == Temp + 3");
        assert(*p == q, "*p == q");
        if *(q+1) == '{' { // and/or
        
            var q string  = pStr + pMatches[ ++(*p) - pStr ];
            assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
            *p = q;
        }
        if Vec_IntSize(&pGia.vHTable) == 0 {
            if pGia.pMuxes {
                Res = Gia_ManAppendMux( pGia, Temp[0], Temp[1], Temp[2] );
            } else {
                Res = Gia_ManAppendMux2( pGia, Temp[0], Temp[1], Temp[2] ); }
        
        } else {
            if pGia.pMuxes {
                Res = Gia_ManHashMuxReal( pGia, Temp[0], Temp[1], Temp[2] );
            } else {
                Res = Gia_ManHashMux( pGia, Temp[0], Temp[1], Temp[2] ); }
        }
        pObj = Gia_ManObj(pGia, Abc_Lit2Var(Res));
        if Gia_ObjIsAnd(pObj) {
            if pGia.pMuxes && Vec_IntSize(&pGia.vHTable) {
                Gia_ObjSetMuxLevel( pGia, pObj );
            } else {
                if Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) {
                    Gia_ObjSetAndLevel( pGia, Gia_ObjFanin0(pObj) ); }
                if Gia_ObjIsAnd(Gia_ObjFanin1(pObj)) {
                    Gia_ObjSetAndLevel( pGia, Gia_ObjFanin1(pObj) ); }
                Gia_ObjSetAndLevel( pGia, pObj );
            }
        }
        return Abc_LitNotCond( Res, fCompl );
    }
    if (**p >= 'A' && **p <= 'F') || (**p >= '0' && **p <= '9') {
        var vLeaves Vec_Int_t;  var q string;
        var pFunc [DAU_DSD_MAX_VAR > 6 ? (1 << (DAU_DSD_MAX_VAR-6)) : 1]word;
        var Fanins [DAU_DSD_MAX_VAR]int32;
        var Res int32;
        var nObjOld int32; 
        var i int32;
        var nVars int32 = Abc_TtReadHex( pFunc, *p );
        *p += Abc_TtHexDigitNum( nVars );
        q = pStr + pMatches[ *p - pStr ];
        assert(**p == '{' && *q == '}', "**p == '{' && *q == '}'");
        for  i = 0, (*p)++; *p < q; (*p)++, i++  {
            Fanins[i] = Dau_DsdToGia_rec( pGia, pStr, p, pMatches, pLits, vCover );
        }
        assert(i == nVars, "i == nVars");
        assert(*p == q, "*p == q");
        vLeaves.nCap = nVars;
        vLeaves.nSize = nVars;
        vLeaves.pArray = Fanins;      
        nObjOld = Gia_ManObjNum(pGia);
        Res = Kit_TruthToGia( pGia, (unsigned *)pFunc, nVars, vCover, &vLeaves, Vec_IntSize(&pGia.vHTable) != 0 );
//        assert( nVars <= 6 );
//        Res = Dau_DsdToGiaCompose_rec( pGia, pFunc[0], Fanins, nVars );
        for  i = nObjOld; i < Gia_ManObjNum(pGia); i++  {
            Gia_ObjSetGateLevel( pGia, Gia_ManObj(pGia, i) );
        }
        m_Non1Step++;
        return Abc_LitNotCond( Res, fCompl );
    }
    assert(false, "0");
    return 0;
}
func Dau_DsdToGia(pGia *Gia_Man_t, p string, pLits *int32, vCover *Vec_Int_t) int32 {
    var Res int32;
    if *p == '0' && *(p+1) == 0 {
        Res = 0;
    } else if *p == '1' && *(p+1) == 0 {
        Res = 1;
    } else {
        Res = Dau_DsdToGia_rec( pGia, p, &p, Dau_DsdComputeMatches(p), pLits, vCover ); }
    assert(*++p == 0, "*++p == 0");
    return Res;
}

/**Function*************************************************************

  Synopsis    [Convert TT to GIA via DSD.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsm_ManTruthToGia(p any, pTruth *word, vLeaves *Vec_Int_t, vCover *Vec_Int_t) int32 {
    var fUseMuxes int32;
    var fDelayBalance int32  = 1;
    var pGia *Gia_Man_t  = (Gia_Man_t *)p;
    var nSizeNonDec int32;
    var pDsd [1000]byte;
    var pTruthCopy [DAU_MAX_WORD]word;
    Abc_TtCopy( pTruthCopy, pTruth, Abc_TtWordNum(Vec_IntSize(vLeaves)), 0 );
    m_Calls++;
    assert(Vec_IntSize(vLeaves) <= DAU_DSD_MAX_VAR, "Vec_IntSize(vLeaves) <= DAU_DSD_MAX_VAR");
    if Vec_IntSize(vLeaves) == 0 {
        return int32((pTruth[0] & 1)); }
    if Vec_IntSize(vLeaves) == 1 {
        return Abc_LitNotCond( Vec_IntEntry(vLeaves, 0), int32((pTruth[0] & 1)) ); }
    // collect delay information
    if (fDelayBalance != 0) && (fUseMuxes != 0) {
        var i int32;
        var iLit int32;
        var pVarLevels [DAU_DSD_MAX_VAR]int32;
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iLit) = Vec_IntEntry(vLeaves, i)), 1); i++  {
            pVarLevels[i] = Gia_ObjLevelId( pGia, Abc_Lit2Var(iLit) );
        }
        nSizeNonDec = Dau_DsdDecomposeLevel( pTruthCopy, Vec_IntSize(vLeaves), fUseMuxes, 1, pDsd, pVarLevels );
    
    } else {
        nSizeNonDec = Dau_DsdDecompose( pTruthCopy, Vec_IntSize(vLeaves), fUseMuxes, 1, pDsd ); }
    if nSizeNonDec != 0 {
        m_NonDsd++; }
//    printf( "%s\n", pDsd );
    if (fDelayBalance != 0) && pGia.vLevels {
        return Dau_DsdToGia( pGia, pDsd, Vec_IntArray(vLeaves), vCover );
    } else {
        return Dau_DsdToGia2( pGia, pDsd, Vec_IntArray(vLeaves), vCover ); }
}

/**Function*************************************************************

  Synopsis    [Convert TT to GIA via DSD.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsm_ManReportStats() {
    fmt.Printf("Calls = %d. NonDSD = %d. Non1Step = %d.\n", m_Calls, m_NonDsd, m_Non1Step);
    m_Calls = m_NonDsd = m_Non1Step = 0;
}

/**Function*************************************************************

  Synopsis    [Performs structural hashing on the LUT functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsm_ManDeriveGia(pGia any, fUseMuxes int32) any {
    var p *Gia_Man_t  = (Gia_Man_t *)pGia;
    var pNew, pTemp *Gia_Man_t;
    var vCover, vLeaves *Vec_Int_t;
    var pObj *Gia_Obj_t; 
    var k, i, iLut, iVar int32;
    var pTruth *word;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");   
    // create new manager
    pNew = Gia_ManStart( 6*Gia_ManObjNum(p)/5 + 100 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    pNew.vLevels = Vec_IntStart( 6*Gia_ManObjNum(p)/5 + 100 );
    if fUseMuxes != 0 {
        // inlining of macro ABC_CALLOC
        pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned))); }
    // map primary inputs
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // iterate through nodes used in the mapping
    vLeaves = Vec_IntAlloc( 16 );
    vCover  = Vec_IntAlloc( 1 << 16 );
    Gia_ManHashStart( pNew );
    Gia_ObjComputeTruthTableStart( p, Gia_ManLutSizeMax(p) );
    // inlining of macro Gia_ManForEachAnd
    for  iLut = 0; (iLut < p.nObjs) && ((pObj) = Gia_ManObj(p, iLut)); iLut++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if Gia_ObjIsBuf(pObj) {
            pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
            continue;
        }
        if !Gia_ObjIsLut(p, iLut) {
            continue; }
        // collect leaves
        Vec_IntClear( vLeaves );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iLut) && ((iVar = Gia_ObjLutFanins(p,iLut)[k]),1); k++  {
            Vec_IntPush( vLeaves, iVar );
        }
        pTruth = Gia_ObjComputeTruthTableCut( p, Gia_ManObj(p, iLut), vLeaves );
        // collect incoming literals
        Vec_IntClear( vLeaves );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iLut) && ((iVar = Gia_ObjLutFanins(p,iLut)[k]),1); k++  {
            Vec_IntPush( vLeaves, Gia_ManObj(p, iVar).Value );
        }
        Gia_ManObj(p, iLut).Value = Dsm_ManTruthToGia( pNew, pTruth, vLeaves, vCover );
    }
    }
    Gia_ObjComputeTruthTableStop( p );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vCover );
/*
    Gia_ManForEachAnd( pNew, pObj, i )
    {
        int iLev  = Gia_ObjLevelId(pNew, i);
        int iLev0 = Gia_ObjLevelId(pNew, Gia_ObjFaninId0(pObj, i));
        int iLev1 = Gia_ObjLevelId(pNew, Gia_ObjFaninId1(pObj, i));
        assert( iLev == 1 + Abc_MaxInt(iLev0, iLev1) );
    }
*/
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

