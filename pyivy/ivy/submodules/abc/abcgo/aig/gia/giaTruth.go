package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaTruth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Truth table computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaTruth.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecMem.h"
//#include "misc/vec/vecWec.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

var s_Truth6 [6]word = [6]word{
    ABC_CONST(0xAAAAAAAAAAAAAAAA),
    ABC_CONST(0xCCCCCCCCCCCCCCCC),
    ABC_CONST(0xF0F0F0F0F0F0F0F0),
    ABC_CONST(0xFF00FF00FF00FF00),
    ABC_CONST(0xFFFF0000FFFF0000),
    ABC_CONST(0xFFFFFFFF00000000),
};

func Gla_ObjTruthElem(p *Gia_Man_t, i int32) *word { return (word *)Vec_PtrEntry( p.vTtInputs, i );                                           }
func Gla_ObjTruthNodeId(p *Gia_Man_t, Id int32) *word { return Vec_WrdArray(p.vTtMemory) + p.nTtWords * Id;                                     }
func Gla_ObjTruthNode(p *Gia_Man_t, pObj *Gia_Obj_t) *word { return Vec_WrdArray(p.vTtMemory) + p.nTtWords * Gia_ObjNum(p, pObj);                    }
func Gla_ObjTruthFree1(p *Gia_Man_t) *word { return Vec_WrdArray(p.vTtMemory) + Vec_WrdSize(p.vTtMemory) - p.nTtWords * 1;          }
func Gla_ObjTruthFree2(p *Gia_Man_t) *word { return Vec_WrdArray(p.vTtMemory) + Vec_WrdSize(p.vTtMemory) - p.nTtWords * 2;          }
func Gla_ObjTruthConst0(p *Gia_Man_t, pDst *word) *word { var w int32; for  w = 0; w < p.nTtWords; w++  { pDst[w] = 0; }
 return pDst;                      }
func Gla_ObjTruthDup(p *Gia_Man_t, pDst *word, pSrc *word, c int32) *word { var w int32; for  w = 0; w < p.nTtWords; w++  { pDst[w] = c ? ~pSrc[w] : pSrc[w]; }
 return pDst; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Compute truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_LutComputeTruth6Simple_rec(p *Gia_Man_t, iObj int32) word {
    var Truth0, Truth1, Truth word;
    var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
    if Gia_ObjIsConst0(pObj) {
        return 0; }
    if Gia_ObjIsCi(pObj) {
        return s_Truths6[Gia_ObjCioId(pObj)]; }
    Truth0 = Gia_LutComputeTruth6Simple_rec( p, Gia_ObjFaninId0(pObj, iObj) );
    Truth1 = Gia_LutComputeTruth6Simple_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    if Gia_ObjFaninC0(pObj) {
        Truth0 = ~Truth0;
    } else {
        Truth0 = Truth0;
    }
    if Gia_ObjFaninC1(pObj) {
        Truth1 = ~Truth1;
    } else {
        Truth1 = Truth1;
    }
    if Gia_ObjIsXor(pObj) {
        Truth = Truth0 ^ Truth1;
    } else {
        Truth = Truth0 & Truth1;
    }
    return Truth;
}
func Gia_LutComputeTruth6Simple(p *Gia_Man_t, iPo int32) word {
    var pObj *Gia_Obj_t  = Gia_ManPo( p, iPo );
    var Truth word  = Gia_LutComputeTruth6Simple_rec( p, Gia_ObjFaninId0p(p, pObj) );
    if Gia_ObjFaninC0(pObj) {
        return ~Truth;
    } else {
        return Truth;
    }
}

func Gia_LutComputeTruth6Map_rec(p *Gia_Man_t, iObj int32, vMap *Vec_Int_t) word {
    var Truth0, Truth1, Truth word;
    var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
    if Gia_ObjIsConst0(pObj) {
        return 0; }
    if Gia_ObjIsCi(pObj) {
        return s_Truths6[Vec_IntEntry(vMap, Gia_ObjCioId(pObj))]; }
    Truth0 = Gia_LutComputeTruth6Map_rec( p, Gia_ObjFaninId0(pObj, iObj), vMap );
    Truth1 = Gia_LutComputeTruth6Map_rec( p, Gia_ObjFaninId1(pObj, iObj), vMap );
    if Gia_ObjFaninC0(pObj) {
        Truth0 = ~Truth0;
    } else {
        Truth0 = Truth0;
    }
    if Gia_ObjFaninC1(pObj) {
        Truth1 = ~Truth1;
    } else {
        Truth1 = Truth1;
    }
    if Gia_ObjIsXor(pObj) {
        Truth = Truth0 ^ Truth1;
    } else {
        Truth = Truth0 & Truth1;
    }
    return Truth;
}
func Gia_LutComputeTruth6Map(p *Gia_Man_t, iPo int32, vMap *Vec_Int_t) word {
    var pObj *Gia_Obj_t  = Gia_ManPo( p, iPo );
    var Truth word  = Gia_LutComputeTruth6Map_rec( p, Gia_ObjFaninId0p(p, pObj), vMap );
    if Gia_ObjFaninC0(pObj) {
        return ~Truth;
    } else {
        return Truth;
    }
}

/**Function*************************************************************

  Synopsis    [Generate MUX tree of the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var s_Truths5 [5]uint32 = [5]uint32{
    0xAAAAAAAA,
    0xCCCCCCCC,
    0xF0F0F0F0,
    0xFF00FF00,
    0xFFFF0000,
};
func Abc_Tt5HasVar(t uint32, iVar int32) int32 {
    return ((t << (1<<iVar)) & s_Truths5[iVar]) != (t & s_Truths5[iVar]);
}
func Abc_Tt5Cofactor0(t uint32, iVar int32) uint32 {
    assert(iVar >= 0 && iVar < 6, "iVar >= 0 && iVar < 6");
    return (t & ~s_Truths5[iVar]) | ((t & ~s_Truths5[iVar]) << (1<<iVar));
}
func Abc_Tt5Cofactor1(t uint32, iVar int32) uint32 {
    assert(iVar >= 0 && iVar < 6, "iVar >= 0 && iVar < 6");
    return (t & s_Truths5[iVar]) | ((t & s_Truths5[iVar]) >> (1<<iVar));
}
func Gia_Truth5ToGia(p *Gia_Man_t, pVarLits *int32, nVars int32, Truth uint32, fHash int32) int32 {
    var Var, Lit0, Lit1 int32; 
    if Truth == 0 {
        return 0; }
    if ~Truth == 0 {
        return 1; }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if Abc_Tt5HasVar( Truth, Var ) != 0 {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    Lit0 = Gia_Truth5ToGia( p, pVarLits, Var, Abc_Tt5Cofactor0(Truth, Var), fHash );
    Lit1 = Gia_Truth5ToGia( p, pVarLits, Var, Abc_Tt5Cofactor1(Truth, Var), fHash );
    if fHash != 0 {
        return Gia_ManHashMux( p, pVarLits[Var], Lit1, Lit0 );
    } else {
        return Gia_ManAppendMux( p, pVarLits[Var], Lit1, Lit0 ); }
}

/**Function*************************************************************

  Synopsis    [Generate MUX tree of the truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_Truth6ToGia(p *Gia_Man_t, pVarLits *int32, nVars int32, Truth word, fHash int32) int32 {
    var Var, Lit0, Lit1 int32; 
    if Truth == 0 {
        return 0; }
    if ~Truth == 0 {
        return 1; }
    assert(nVars > 0, "nVars > 0");
    // find the topmost var
    for  Var = nVars-1; Var >= 0; Var--  {
        if Abc_Tt6HasVar( Truth, Var ) {
             break; }
    }
    assert(Var >= 0, "Var >= 0");
    // cofactor
    Lit0 = Gia_Truth6ToGia( p, pVarLits, Var, Abc_Tt6Cofactor0(Truth, Var), fHash );
    Lit1 = Gia_Truth6ToGia( p, pVarLits, Var, Abc_Tt6Cofactor1(Truth, Var), fHash );
    if fHash != 0 {
        return Gia_ManHashMux( p, pVarLits[Var], Lit1, Lit0 );
    } else {
        return Gia_ManAppendMux( p, pVarLits[Var], Lit1, Lit0 ); }
}
func Gia_Truth6ToGiaTest(p *Gia_Man_t) {
    var Size int32  = 5;
    var Truth, TruthNew word;
    var vMap *Vec_Int_t  = Vec_IntStartFull( Gia_ManCiNum(p) );
    var vSupp *Vec_Int_t  = Vec_IntStart( 100 );
    var nCos int32 = Gia_ManCoNum(p);
    var Count int32;
    var i, k, Id, ObjId, iLitNew int32;
    Gia_ManHashAlloc( p );
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++  {
        Gia_ManCollectCis( p, &Id, 1, vSupp ); // ObjIds
        if Vec_IntSize(vSupp) <= Size && i < nCos {
            var pVarLits [6]int32;
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSupp)) && (((ObjId) = Vec_IntEntry(vSupp, k)), 1); k++  {
                var CioId int32  = Gia_ObjCioId(Gia_ManObj(p, ObjId));
                Vec_IntWriteEntry( vMap, CioId, k );
                pVarLits[k] = Abc_Var2Lit( ObjId, 0 );
            }
            Truth = Gia_LutComputeTruth6Map( p, i, vMap );
            if Size == 5 {
                iLitNew = Gia_Truth5ToGia( p, pVarLits, Vec_IntSize(vSupp), uint32(Truth), 1 );
            } else {
                iLitNew = Gia_Truth6ToGia( p, pVarLits, Vec_IntSize(vSupp), Truth, 1 ); }
            Gia_ManAppendCo( p, iLitNew );
            TruthNew = Gia_LutComputeTruth6Map( p, Gia_ManCoNum(p)-1, vMap );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vSupp)) && (((ObjId) = Vec_IntEntry(vSupp, k)), 1); k++  {
                var CioId int32  = Gia_ObjCioId(Gia_ManObj(p, ObjId));
                Vec_IntWriteEntry( vMap, CioId, -1 );
            }
            if Truth != TruthNew {
                fmt.Printf("Error for output %d.\n", i); }
            Count++;
            //Dau_DsdPrintFromTruth( &Truth, Vec_IntSize(vSupp) );
        }
    }
    Gia_ManHashStop( p );
    fmt.Printf("Finished processing %d outputs.\n", Count);
    Vec_IntFree( vSupp );
    Vec_IntFree( vMap );
}

/**Function*************************************************************

  Synopsis    [Compute truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_LutComputeTruth6_rec(p *Gia_Man_t, iNode int32, vTruths *Vec_Wrd_t) word {
    var pObj *Gia_Obj_t;
    var Truth0, Truth1 word;
    if Gia_ObjIsTravIdCurrentId(p, iNode) {
        return Vec_WrdEntry(vTruths, iNode); }
    Gia_ObjSetTravIdCurrentId(p, iNode);
    pObj = Gia_ManObj( p, iNode );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Truth0 = Gia_LutComputeTruth6_rec( p, Gia_ObjFaninId0p(p, pObj), vTruths );
    Truth1 = Gia_LutComputeTruth6_rec( p, Gia_ObjFaninId1p(p, pObj), vTruths );
    if Gia_ObjFaninC0(pObj) {
        Truth0 = ~Truth0; }
    if Gia_ObjFaninC1(pObj) {
        Truth1 = ~Truth1; }
    Vec_WrdWriteEntry( vTruths, iNode, Truth0 & Truth1 );
    return Truth0 & Truth1;
}
func Gia_LutComputeTruth6(p *Gia_Man_t, iObj int32, vTruths *Vec_Wrd_t) word {
    var k, iFan int32;
    assert(Gia_ObjIsLut(p, iObj), "Gia_ObjIsLut(p, iObj)");
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_LutForEachFanin
    for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFan = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
        Vec_WrdWriteEntry( vTruths, iFan, s_Truths6[k] );
        Gia_ObjSetTravIdCurrentId( p, iFan );
    }
    return Gia_LutComputeTruth6_rec( p, iObj, vTruths );
}

/**Function*************************************************************

  Synopsis    [Computes truth table of a 6-LUT.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeTruthTable6Lut_rec(p *Gia_Man_t, iObj int32, vTemp *Vec_Wrd_t) {
    var uTruth0, uTruth1 word;
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ObjComputeTruthTable6Lut_rec( p, Gia_ObjFaninId0p(p, pObj), vTemp );
    Gia_ObjComputeTruthTable6Lut_rec( p, Gia_ObjFaninId1p(p, pObj), vTemp );
    uTruth0 = Vec_WrdEntry( vTemp, Gia_ObjFaninId0p(p, pObj) );
    if Gia_ObjFaninC0(pObj) {
        uTruth0 = ~uTruth0;
    } else {
        uTruth0 = uTruth0;
    }
    uTruth1 = Vec_WrdEntry( vTemp, Gia_ObjFaninId1p(p, pObj) );
    if Gia_ObjFaninC1(pObj) {
        uTruth1 = ~uTruth1;
    } else {
        uTruth1 = uTruth1;
    }
    Vec_WrdWriteEntry( vTemp, iObj, uTruth0 & uTruth1 );
}
func Gia_ObjComputeTruthTable6Lut(p *Gia_Man_t, iObj int32, vTemp *Vec_Wrd_t) word {
    var i, Fanin int32;
    assert(Vec_WrdSize(vTemp) == Gia_ManObjNum(p), "Vec_WrdSize(vTemp) == Gia_ManObjNum(p)");
    assert(Gia_ObjIsLut(p, iObj), "Gia_ObjIsLut(p, iObj)");
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_LutForEachFanin
    for  i = 0; i < Gia_ObjLutSize(p,iObj) && ((Fanin = Gia_ObjLutFanins(p,iObj)[i]),1); i++  {
        Gia_ObjSetTravIdCurrentId( p, Fanin );
        Vec_WrdWriteEntry( vTemp, Fanin, s_Truth6[i] );
    }
    assert(i <= 6, "i <= 6");
    Gia_ObjComputeTruthTable6Lut_rec( p, iObj, vTemp );
    return Vec_WrdEntry( vTemp, iObj );
}

/**Function*************************************************************

  Synopsis    [Computes truth table up to 6 inputs in terms of CIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeTruth6(p *Gia_Man_t, iObj int32, vSupp *Vec_Int_t, vTemp *Vec_Wrd_t) word {
    var i, Fanin int32;
    assert(Vec_WrdSize(vTemp) == Gia_ManObjNum(p), "Vec_WrdSize(vTemp) == Gia_ManObjNum(p)");
    assert(Vec_IntSize(vSupp) <= 6, "Vec_IntSize(vSupp) <= 6");
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSupp)) && (((Fanin) = Vec_IntEntry(vSupp, i)), 1); i++  {
        Gia_ObjSetTravIdCurrentId( p, Fanin );
        Vec_WrdWriteEntry( vTemp, Fanin, s_Truth6[i] );
    }
    Gia_ObjComputeTruthTable6Lut_rec( p, iObj, vTemp );
    return Vec_WrdEntry( vTemp, iObj );
}
func Gia_ObjComputeTruth6CisSupport_rec(p *Gia_Man_t, iObj int32, vSupp *Vec_Int_t) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    if Gia_ObjIsCi(pObj) {
        Vec_IntPushOrder( vSupp, iObj );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ObjComputeTruth6CisSupport_rec( p, Gia_ObjFaninId0p(p, pObj), vSupp );
    Gia_ObjComputeTruth6CisSupport_rec( p, Gia_ObjFaninId1p(p, pObj), vSupp );
}
func Gia_ObjComputeTruth6Cis(p *Gia_Man_t, iLit int32, vSupp *Vec_Int_t, vTemp *Vec_Wrd_t) word {
    var iObj int32  = Abc_Lit2Var(iLit); 
    Vec_IntClear( vSupp );
    if !(iObj != 0) { if Abc_LitIsCompl(iLit) {
        return ~(word)0;
    } else {
        return (word)0;
    } }
    Gia_ManIncrementTravId( p );
    Gia_ObjComputeTruth6CisSupport_rec( p, iObj, vSupp );
    if Vec_IntSize(vSupp) > 6 {
        return 0; }
    Gia_ObjComputeTruth6( p, iObj, vSupp, vTemp );
    if Abc_LitIsCompl(iLit) {
        return ~Vec_WrdEntry(vTemp, iObj);
    } else {
        return Vec_WrdEntry(vTemp, iObj);
    }
}

/**Function*************************************************************

  Synopsis    [Computes truth table up to 6 inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeTruthTable6_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vTruths *Vec_Wrd_t) {
    var uTruth0, uTruth1 word;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    assert(!pObj.fMark0, "!pObj.fMark0");
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ObjComputeTruthTable6_rec( p, Gia_ObjFanin0(pObj), vTruths );
    Gia_ObjComputeTruthTable6_rec( p, Gia_ObjFanin1(pObj), vTruths );
    uTruth0 = Vec_WrdEntry( vTruths, Gia_ObjFanin0(pObj).Value );
    if Gia_ObjFaninC0(pObj) {
        uTruth0 = ~uTruth0;
    } else {
        uTruth0 = uTruth0;
    }
    uTruth1 = Vec_WrdEntry( vTruths, Gia_ObjFanin1(pObj).Value );
    if Gia_ObjFaninC1(pObj) {
        uTruth1 = ~uTruth1;
    } else {
        uTruth1 = uTruth1;
    }
    pObj.Value = Vec_WrdSize(vTruths);
    Vec_WrdPush( vTruths, uTruth0 & uTruth1 );
}
func Gia_ObjComputeTruthTable6(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t, vTruths *Vec_Wrd_t) word {
    var pLeaf *Gia_Obj_t;
    var i int32;
    assert(Vec_IntSize(vSupp) <= 6, "Vec_IntSize(vSupp) <= 6");
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(!pObj.fMark0, "!pObj.fMark0");
    Vec_WrdClear( vTruths );
    Gia_ManIncrementTravId( p );
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vSupp)) && ((pLeaf) = Gia_ManObj(p, Vec_IntEntry(vSupp,i))); i++  {
        assert(pLeaf.fMark0 || Gia_ObjIsRo(p, pLeaf), "pLeaf.fMark0 || Gia_ObjIsRo(p, pLeaf)");
        pLeaf.Value = Vec_WrdSize(vTruths);
        Vec_WrdPush( vTruths, s_Truth6[i] );
        Gia_ObjSetTravIdCurrent(p, pLeaf);
    }
    Gia_ObjComputeTruthTable6_rec( p, pObj, vTruths );
    return Vec_WrdEntryLast( vTruths );
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes reachable from the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjCollectInternal_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if !Gia_ObjIsAnd(pObj) {
        return; }
    if pObj.fMark0 {
        return; }
    pObj.fMark0 = 1;
    Gia_ObjCollectInternal_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ObjCollectInternal_rec( p, Gia_ObjFanin1(pObj) );
    Gia_ObjSetNum( p, pObj, Vec_IntSize(p.vTtNodes) );
    Vec_IntPush( p.vTtNodes, Gia_ObjId(p, pObj) );
}
func Gia_ObjCollectInternal(p *Gia_Man_t, pObj *Gia_Obj_t) {
    Vec_IntClear( p.vTtNodes );
    Gia_ObjCollectInternal_rec( p, pObj );
}

/**Function*************************************************************

  Synopsis    [Computing the truth table for GIA object.]

  Description [The truth table should be used by the calling application
  (or saved into the user's storage) before this procedure is called again.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeTruthTable(p *Gia_Man_t, pObj *Gia_Obj_t) *word {
    var pTemp, pRoot *Gia_Obj_t;
    var pTruth, pTruthL, pTruth0, pTruth1 *word;
    var i int32;
    if p.vTtMemory == nil {
        p.nTtVars   = Gia_ManPiNum( p );
        p.nTtWords  = Abc_Truth6WordNum( p.nTtVars );
        p.vTtNums   = Vec_IntStart( Gia_ManObjNum(p) + 1000 );
        p.vTtNodes  = Vec_IntAlloc( 256 );
        p.vTtInputs = Vec_PtrAllocTruthTables( Abc_MaxInt(6, p.nTtVars) );
        p.vTtMemory = Vec_WrdStart( p.nTtWords * 256 );
    
    } else {
        // make sure the number of primary inputs did not change 
        // since the truth table computation storage was prepared
        assert(p.nTtVars == Gia_ManPiNum(p), "p.nTtVars == Gia_ManPiNum(p)");
    }
    // extend ID numbers
    if Vec_IntSize(p.vTtNums) < Gia_ManObjNum(p) {
        Vec_IntFillExtra( p.vTtNums, Gia_ManObjNum(p), 0 ); }
    // collect internal nodes
    if Gia_ObjIsCo(pObj) {
        pRoot = Gia_ObjFanin0(pObj);
    } else {
        pRoot = pObj;
    }
    Gia_ObjCollectInternal( p, pRoot );
    // extend TT storage
    if Vec_WrdSize(p.vTtMemory) < p.nTtWords * (Vec_IntSize(p.vTtNodes) + 2) {
        Vec_WrdFillExtra( p.vTtMemory, p.nTtWords * (Vec_IntSize(p.vTtNodes) + 2), 0 ); }
    // compute the truth table for internal nodes
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vTtNodes)) && ((pTemp) = Gia_ManObj(p, Vec_IntEntry(p.vTtNodes,i))); i++  {
        pTemp.fMark0 = 0; // unmark nodes marked by Gia_ObjCollectInternal()
        pTruth  = Gla_ObjTruthNode(p, pTemp);
        pTruthL = pTruth + p.nTtWords;
        if Gia_ObjIsAnd(Gia_ObjFanin0(pTemp)) {
            pTruth0 = Gla_ObjTruthNode(p, Gia_ObjFanin0(pTemp));
        } else {
            pTruth0 = Gla_ObjTruthElem(p, Gia_ObjCioId(Gia_ObjFanin0(pTemp)) );
        }
        if Gia_ObjIsAnd(Gia_ObjFanin1(pTemp)) {
            pTruth1 = Gla_ObjTruthNode(p, Gia_ObjFanin1(pTemp));
        } else {
            pTruth1 = Gla_ObjTruthElem(p, Gia_ObjCioId(Gia_ObjFanin1(pTemp)) );
        }
        if Gia_ObjFaninC0(pTemp) {
            if Gia_ObjFaninC1(pTemp) {
                for  pTruth < pTruthL  {
                    *pTruth++ = ~*pTruth0++ & ~*pTruth1++;
                }
            } else {
                for  pTruth < pTruthL  {
                    *pTruth++ = ~*pTruth0++ &  *pTruth1++;
                } }
        
        } else {
            if Gia_ObjFaninC1(pTemp) {
                for  pTruth < pTruthL  {
                    *pTruth++ =  *pTruth0++ & ~*pTruth1++;
                }
            } else {
                for  pTruth < pTruthL  {
                    *pTruth++ =  *pTruth0++ &  *pTruth1++;
                } }
        }
    }
    // compute the final table
    if Gia_ObjIsConst0(pRoot) {
        pTruth = Gla_ObjTruthConst0( p, Gla_ObjTruthFree1(p) );
    } else if Gia_ObjIsPi(p, pRoot) {
        pTruth = Gla_ObjTruthElem( p, Gia_ObjCioId(pRoot) );
    } else if Gia_ObjIsAnd(pRoot) {
        pTruth = Gla_ObjTruthNode( p, pRoot );
    } else {
        pTruth = nil; }
    return Gla_ObjTruthDup( p, Gla_ObjTruthFree2(p), pTruth, Gia_ObjIsCo(pObj) && Gia_ObjFaninC0(pObj) );
}

/**Function*************************************************************

  Synopsis    [Testing truth table computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeTruthTableTest(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var pTruth *uint32;
    var clk abctime  = Abc_Clock();
    var i int32;
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pTruth = (unsigned *)Gia_ObjComputeTruthTable( p, pObj );
//        Extra_PrintHex( stdout, pTruth, Gia_ManPiNum(p) ); printf( "\n" );
    }
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeTruthTableStart(p *Gia_Man_t, nVarsMax int32) {
    assert(p.vTtMemory == nil, "p.vTtMemory == nil");
    p.nTtVars   = nVarsMax;
    p.nTtWords  = Abc_Truth6WordNum( p.nTtVars );
    p.vTtNodes  = Vec_IntAlloc( 256 );
    p.vTtInputs = Vec_PtrAllocTruthTables( Abc_MaxInt(6, p.nTtVars) );
    p.vTtMemory = Vec_WrdStart( p.nTtWords * 64 );
    p.vTtNums   = Vec_IntAlloc( Gia_ManObjNum(p) + 1000 );
    Vec_IntFill( p.vTtNums, Vec_IntCap(p.vTtNums), -ABC_INFINITY );
}
func Gia_ObjComputeTruthTableStop(p *Gia_Man_t) {
    p.nTtVars   = 0;
    p.nTtWords  = 0;
    Vec_IntFreeP( &p.vTtNums );
    Vec_IntFreeP( &p.vTtNodes );
    Vec_PtrFreeP( &p.vTtInputs );
    Vec_WrdFreeP( &p.vTtMemory );
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes reachable from the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjCollectInternalCut_rec(p *Gia_Man_t, iObj int32) {
    if Gia_ObjHasNumId(p, iObj) {
        return; }
    assert(Gia_ObjIsAnd(Gia_ManObj(p, iObj)), "Gia_ObjIsAnd(Gia_ManObj(p, iObj))");
    Gia_ObjCollectInternalCut_rec( p, Gia_ObjFaninId0(Gia_ManObj(p, iObj), iObj) );
    Gia_ObjCollectInternalCut_rec( p, Gia_ObjFaninId1(Gia_ManObj(p, iObj), iObj) );
    Gia_ObjSetNumId( p, iObj, Vec_IntSize(p.vTtNodes) );
    Vec_IntPush( p.vTtNodes, iObj );
}
func Gia_ObjCollectInternalCut(p *Gia_Man_t, iRoot int32, vLeaves *Vec_Int_t) {
    var i, iObj int32;
    assert(!Gia_ObjHasNumId(p, iRoot), "!Gia_ObjHasNumId(p, iRoot)");
    assert(Gia_ObjIsAnd(Gia_ManObj(p, iRoot)), "Gia_ObjIsAnd(Gia_ManObj(p, iRoot))");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iObj) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        if Gia_ObjHasNumId(p, iObj) { // if cuts have repeated variables, skip 
            continue; }
        assert(!Gia_ObjHasNumId(p, iObj), "!Gia_ObjHasNumId(p, iObj)");
        Gia_ObjSetNumId( p, iObj, -i );
    }
    assert(!Gia_ObjHasNumId(p, iRoot), "!Gia_ObjHasNumId(p, iRoot)"); // the root cannot be one of the leaves
    Vec_IntClear( p.vTtNodes );
    Vec_IntPush( p.vTtNodes, -1 );
    Gia_ObjCollectInternalCut_rec( p, iRoot );
}

/**Function*************************************************************

  Synopsis    [Computes the truth table of pRoot in terms of leaves.]

  Description [The root cannot be one of the leaves.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeTruthTableCut(p *Gia_Man_t, pRoot *Gia_Obj_t, vLeaves *Vec_Int_t) *word {
    var pTemp *Gia_Obj_t;
    var pTruth, pTruthL, pTruth0, pTruth1 *word;
    var i, iObj, Id0, Id1 int32;
    assert(p.vTtMemory != nil, "p.vTtMemory != nil");
    assert(Vec_IntSize(vLeaves) <= p.nTtVars, "Vec_IntSize(vLeaves) <= p.nTtVars");
    // extend ID numbers
    if Vec_IntSize(p.vTtNums) < Gia_ManObjNum(p) {
        Vec_IntFillExtra( p.vTtNums, Gia_ManObjNum(p), -ABC_INFINITY ); }
    // collect internal nodes
    Gia_ObjCollectInternalCut( p, Gia_ObjId(p, pRoot), vLeaves );
    // extend TT storage
    if Vec_WrdSize(p.vTtMemory) < p.nTtWords * (Vec_IntSize(p.vTtNodes) + 2) {
        Vec_WrdFillExtra( p.vTtMemory, p.nTtWords * (Vec_IntSize(p.vTtNodes) + 2), 0 ); }
    // compute the truth table for internal nodes
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(p.vTtNodes)) && (((iObj) = Vec_IntEntry(p.vTtNodes, i)), 1); i++  {
        assert(i == Gia_ObjNumId(p, iObj), "i == Gia_ObjNumId(p, iObj)");
        pTemp   = Gia_ManObj( p, iObj );
        pTruth  = Gla_ObjTruthNodeId( p, i );  
        pTruthL = pTruth + p.nTtWords;
        Id0 = Gia_ObjNumId( p, Gia_ObjFaninId0(pTemp, iObj) );
        Id1 = Gia_ObjNumId( p, Gia_ObjFaninId1(pTemp, iObj) );
        if Id0 > 0 {
            pTruth0 = Gla_ObjTruthNodeId(p, Id0);
        } else {
            pTruth0 = Gla_ObjTruthElem(p, -Id0);
        }
        if Id1 > 0 {
            pTruth1 = Gla_ObjTruthNodeId(p, Id1);
        } else {
            pTruth1 = Gla_ObjTruthElem(p, -Id1);
        }
        if Gia_ObjFaninC0(pTemp) {
            if Gia_ObjFaninC1(pTemp) {
                for  pTruth < pTruthL  {
                    *pTruth++ = ~*pTruth0++ & ~*pTruth1++;
                }
            } else {
                for  pTruth < pTruthL  {
                    *pTruth++ = ~*pTruth0++ &  *pTruth1++;
                } }
        
        } else {
            if Gia_ObjFaninC1(pTemp) {
                for  pTruth < pTruthL  {
                    *pTruth++ =  *pTruth0++ & ~*pTruth1++;
                }
            } else {
                for  pTruth < pTruthL  {
                    *pTruth++ =  *pTruth0++ &  *pTruth1++;
                } }
        }
    }
    pTruth = Gla_ObjTruthNode( p, pRoot );
    // unmark leaves marked by Gia_ObjCollectInternal()
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iObj) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        Gia_ObjResetNumId( p, iObj );
    }
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(p.vTtNodes)) && (((iObj) = Vec_IntEntry(p.vTtNodes, i)), 1); i++  {
        Gia_ObjResetNumId( p, iObj );
    }
    return pTruth;
}

/**Function*************************************************************

  Synopsis    [Reduces GIA to contain isomorphic POs.]

  Description [The root cannot be one of the leaves.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManIsoNpnReduce(p *Gia_Man_t, pvPosEquivs **Vec_Ptr_t, fVerbose int32) *Gia_Man_t {
    var pCanonPerm [16]byte;
    var i, iObj, uCanonPhase, nVars, lastId, truthId int32;
    var IndexCon int32 = -1;
    var IndexVar int32 = -1;
    var vPosEquivs *Vec_Wec_t  = Vec_WecAlloc( 100 );
    var pTruth *word;
    var pObj *Gia_Obj_t;
    var vTtMem [17]*Vec_Mem_t;   // truth table memory and hash table
    var pNew *Gia_Man_t  = nil;
    var vLeaves *Vec_Int_t  = Vec_IntAlloc( 16 );
    var vFirsts *Vec_Int_t;
    var vTt2Class [17]*Vec_Int_t;
    for  i = 0; i < 17; i++  {
        vTtMem[i] = Vec_MemAlloc( Abc_TtWordNum(i), 10 );
        Vec_MemHashAlloc( vTtMem[i], 1000 );
        vTt2Class[i] = Vec_IntStartFull( Gia_ManCoNum(p)+1 );
    }
    Gia_ObjComputeTruthTableStart( p, 16 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        iObj = Gia_ObjId(p, pObj);
        Gia_ManCollectCis( p, &iObj, 1, vLeaves );
        if Vec_IntSize(vLeaves) > 16 {
            Vec_IntPush( Vec_WecPushLevel(vPosEquivs), i );
            continue;
        }
        pObj = Gia_ObjFanin0(pObj);
        if Gia_ObjIsConst0(pObj) {
            if IndexCon == -1 {
                IndexCon = Vec_WecSize(vPosEquivs);
                Vec_WecPushLevel(vPosEquivs);
            }
            Vec_WecPush( vPosEquivs, IndexCon, i );
            continue;
        }
        if Gia_ObjIsCi(pObj) {
            if IndexVar == -1 {
                IndexVar = Vec_WecSize(vPosEquivs);
                Vec_WecPushLevel(vPosEquivs);
            }
            Vec_WecPush( vPosEquivs, IndexVar, i );
            continue;
        }
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        pTruth = Gia_ObjComputeTruthTableCut( p, pObj, vLeaves );
        Abc_TtMinimumBase( pTruth, nil, Vec_IntSize(vLeaves), &nVars );
        if nVars == 0 {
            if IndexCon == -1 {
                IndexCon = Vec_WecSize(vPosEquivs);
                Vec_WecPushLevel(vPosEquivs);
            }
            Vec_WecPush( vPosEquivs, IndexCon, i );
            continue;
        }
        if nVars == 1 {
            if IndexVar == -1 {
                IndexVar = Vec_WecSize(vPosEquivs);
                Vec_WecPushLevel(vPosEquivs);
            }
            Vec_WecPush( vPosEquivs, IndexVar, i );
            continue;
        }
        uCanonPhase = Abc_TtCanonicize( pTruth, nVars, pCanonPerm );
        lastId = Vec_MemEntryNum( vTtMem[nVars] );
        truthId = Vec_MemHashInsert( vTtMem[nVars], pTruth );
        if lastId != Vec_MemEntryNum( vTtMem[nVars] ) { // new one
        
            assert(Vec_IntEntry(vTt2Class[nVars], truthId) == -1, "Vec_IntEntry(vTt2Class[nVars], truthId) == -1");
            Vec_IntWriteEntry( vTt2Class[nVars], truthId, Vec_WecSize(vPosEquivs) );
            Vec_WecPushLevel(vPosEquivs);
        }
        assert(Vec_IntEntry(vTt2Class[nVars], truthId) >= 0, "Vec_IntEntry(vTt2Class[nVars], truthId) >= 0");
        Vec_WecPush( vPosEquivs, Vec_IntEntry(vTt2Class[nVars], truthId), i );
    }
    Gia_ObjComputeTruthTableStop( p );
    Vec_IntFree( vLeaves );
    for  i = 0; i < 17; i++  {
        Vec_MemHashFree( vTtMem[i] );
        Vec_MemFree( vTtMem[i] );
        Vec_IntFree( vTt2Class[i] );
    }

    // find the first outputs and derive GIA
    vFirsts = Vec_WecCollectFirsts( vPosEquivs );
    pNew = Gia_ManDupCones( p, Vec_IntArray(vFirsts), Vec_IntSize(vFirsts), 0 );
    Vec_IntFree( vFirsts );
    // report and return    
    if fVerbose != 0 { 
        fmt.Printf("Nontrivial classes:\n");
        Vec_WecPrint( vPosEquivs, 1 );
    }
    if pvPosEquivs {
        *pvPosEquivs = Vec_WecConvertToVecPtr( vPosEquivs ); }
    Vec_WecFree( vPosEquivs );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

