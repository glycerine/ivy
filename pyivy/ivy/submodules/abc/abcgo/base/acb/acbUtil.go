package acb

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [acbUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: acbUtil.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acb.h"
//#include "base/abc/abc.h"
//#include "base/io/ioAbc.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Collecting TFI and TFO.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ObjCollectTfi_rec(p *Acb_Ntk_t, iObj, fTerm int32) {
    if Acb_ObjSetTravIdCur(p, iObj) != 0 { return }
    if fTerm == 0 && Acb_ObjIsCi(p, iObj) != 0 { return }
    fanins := Acb_ObjFanins(p, iObj)
    for i := int32(0); i < fanins[0]; i++ {
        Acb_ObjCollectTfi_rec(p, fanins[i+1], fTerm)
    }
    Vec_IntPush(&p.vArray0, iObj)
}
func Acb_ObjCollectTfi(p *Acb_Ntk_t, iObj, fTerm int32) *Vec_Int_t {
    Vec_IntClear(&p.vArray0)
    Acb_NtkIncTravId(p)
    if iObj > 0 {
        for _, node := range Vec_IntArray(&p.vSuppOld) { Acb_ObjCollectTfi_rec(p, node, fTerm) }
        Acb_ObjCollectTfi_rec(p, iObj, fTerm)
    } else {
        for i := int32(0); i < Acb_NtkCoNum(p); i++ { Acb_ObjCollectTfi_rec(p, Acb_NtkCo(p, i), fTerm) }
    }
    return &p.vArray0
}
func Acb_ObjCollectTfiVec(p *Acb_Ntk_t, objects *Vec_Int_t) *Vec_Int_t {
    Vec_IntClear(&p.vArray0)
    Acb_NtkIncTravId(p)
    for _, obj := range Vec_IntArray(objects) { Acb_ObjCollectTfi_rec(p, obj, 0) }
    return &p.vArray0
}
func Acb_ObjCollectTfo_rec(p *Acb_Ntk_t, iObj, fTerm int32) {
    if Acb_ObjSetTravIdCur(p, iObj) != 0 { return }
    if fTerm == 0 && Acb_ObjIsCo(p, iObj) != 0 { return }
    for _, fanout := range Vec_IntArray(Vec_WecEntry(&p.vFanouts, int(iObj))) {
        if Acb_ObjType(p, fanout) != 0 { Acb_ObjCollectTfo_rec(p, fanout, fTerm) }
    }
    Vec_IntPush(&p.vArray1, iObj)
}
func Acb_ObjCollectTfo(p *Acb_Ntk_t, iObj, fTerm int32) *Vec_Int_t {
    Vec_IntClear(&p.vArray1)
    Acb_NtkIncTravId(p)
    if iObj > 0 { Acb_ObjCollectTfo_rec(p, iObj, fTerm) } else {
        for i := int32(0); i < Acb_NtkCiNum(p); i++ { Acb_ObjCollectTfo_rec(p, Acb_NtkCi(p, i), fTerm) }
    }
    return &p.vArray1
}
func Acb_ObjCollectTfoVec(p *Acb_Ntk_t, objects *Vec_Int_t) *Vec_Int_t {
    if Acb_NtkHasObjFanout(p) == 0 { Acb_NtkCreateFanout(p) }
    Vec_IntClear(&p.vArray1)
    Acb_NtkIncTravId(p)
    for _, obj := range Vec_IntArray(objects) { Acb_ObjCollectTfo_rec(p, obj, 0) }
    return &p.vArray1
}

/**Function*************************************************************

  Synopsis    [Count the number of nodes driving the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkIsPiBuffers(p *Acb_Ntk_t, iObj int32) int32 {
    if Acb_ObjIsCi(p, iObj) != 0 { return 1 }
    if Acb_ObjFaninNum(p, iObj) != 1 { return 0 }
    return Acb_NtkIsPiBuffers(p, Acb_ObjFanin(p, iObj, 0))
}
func Acb_NtkCountPiBuffers(p *Acb_Ntk_t, objects *Vec_Int_t) int32 {
    var count int32
    for _, obj := range Vec_IntArray(objects) { count += Acb_NtkIsPiBuffers(p, obj) }
    return count
}
func Acb_NtkCountPoDrivers(p *Acb_Ntk_t, objects *Vec_Int_t) int32 {
    var count int32
    Acb_NtkIncTravId(p)
    for i := int32(0); i < Acb_NtkCoNum(p); i++ {
        obj := Acb_NtkCo(p, i)
        fanin := Acb_ObjFanin0(p, obj)
        Acb_ObjSetTravIdCur(p, obj)
        Acb_ObjSetTravIdCur(p, fanin)
        if Acb_ObjFaninNum(p, fanin) == 1 { Acb_ObjSetTravIdCur(p, Acb_ObjFanin0(p, fanin)) }
    }
    for _, obj := range Vec_IntArray(objects) { count += Acb_ObjIsTravIdCur(p, obj) }
    return count
}

/**Function*************************************************************

  Synopsis    [Compute MFFC size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkNodeDeref_rec(vRefs *Vec_Int_t, p *Acb_Ntk_t, iObj int32) int32 {
    var i int32;
    var Fanin int32;
    var pFanins *int32;
    var Counter int32 = 1;
    if Acb_ObjIsCi(p, iObj) {
        return 0; }
    // inlining of macro Acb_ObjForEachFaninFast
    for  i = 0, pFanins = Acb_ObjFanins(p, iObj); i < pFanins[0] && ((Fanin = pFanins[i+1]), 1); i++  {
        assert(Vec_IntEntry(vRefs, Fanin) > 0, "Vec_IntEntry(vRefs, Fanin) > 0");
        Vec_IntAddToEntry( vRefs, Fanin, -1 );
        if Vec_IntEntry(vRefs, Fanin) == 0 {
            Counter += Acb_NtkNodeDeref_rec( vRefs, p, Fanin ); }
    }
    return Counter;
}
func Acb_NtkNodeRef_rec(vRefs *Vec_Int_t, p *Acb_Ntk_t, iObj int32) int32 {
    var i int32;
    var Fanin int32;
    var pFanins *int32;
    var Counter int32 = 1;
    if Acb_ObjIsCi(p, iObj) {
        return 0; }
    // inlining of macro Acb_ObjForEachFaninFast
    for  i = 0, pFanins = Acb_ObjFanins(p, iObj); i < pFanins[0] && ((Fanin = pFanins[i+1]), 1); i++  {
        if Vec_IntEntry(vRefs, Fanin) == 0 {
            Counter += Acb_NtkNodeRef_rec( vRefs, p, Fanin ); }
        Vec_IntAddToEntry( vRefs, Fanin, 1 );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Computing and updating direct and reverse logic level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkCollectDeref_rec(vRefs *Vec_Int_t, p *Acb_Ntk_t, iObj int32, vRes *Vec_Int_t) {
    var i int32;
    var Fanin int32;
    var pFanins *int32;
    if Acb_ObjIsCi(p, iObj) {
        return; }
    Vec_IntPush( vRes, iObj );
    // inlining of macro Acb_ObjForEachFaninFast
    for  i = 0, pFanins = Acb_ObjFanins(p, iObj); i < pFanins[0] && ((Fanin = pFanins[i+1]), 1); i++  {
        assert(Vec_IntEntry(vRefs, Fanin) > 0, "Vec_IntEntry(vRefs, Fanin) > 0");
        Vec_IntAddToEntry( vRefs, Fanin, -1 );
        if Vec_IntEntry(vRefs, Fanin) == 0 {
            Acb_NtkCollectDeref_rec( vRefs, p, Fanin, vRes ); }
    }
}
func Acb_NtkCollectMffc(p *Acb_Ntk_t, vObjsRefed *Vec_Int_t, vObjsDerefed *Vec_Int_t) *Vec_Int_t {
    var vRes *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vRefs *Vec_Int_t  = Vec_IntStart( Acb_NtkObjNumMax(p) );
    var i int32;
    var iObj int32;
    var Fanin int32;
    var pFanins *int32;
    // inlining of macro Acb_NtkForEachObj
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Acb_ObjType(p, iObj) {} else {   
        // inlining of macro Acb_ObjForEachFaninFast
        for  i = 0, pFanins = Acb_ObjFanins(p, iObj); i < pFanins[0] && ((Fanin = pFanins[i+1]), 1); i++  {
            Vec_IntAddToEntry( vRefs, Fanin, 1 );
        } }
    }
    // inlining of macro Acb_NtkForEachCo
    for  i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1); i++  { 
        Vec_IntAddToEntry( vRefs, iObj, 1 );
    }
    if vObjsRefed != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vObjsRefed)) && (((iObj) = Vec_IntEntry(vObjsRefed, i)), 1); i++  {
            Vec_IntAddToEntry( vRefs, iObj, 1 );
        } }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjsDerefed)) && (((iObj) = Vec_IntEntry(vObjsDerefed, i)), 1); i++  {
        if Acb_ObjIsCo(p, iObj) {
            iObj = Acb_ObjFanin0(p, iObj); }
        if Vec_IntEntry(vRefs, iObj) != 0 {
            Acb_NtkCollectDeref_rec( vRefs, p, iObj, vRes ); }
    }
    Vec_IntFree( vRefs );
    Vec_IntUniqify( vRes );
    return vRes;
}

func Acb_NamesToIds(pNtk *Acb_Ntk_t, vNamesInv *Vec_Int_t, vNames *Vec_Ptr_t) *Vec_Int_t {
    var vRes *Vec_Int_t  = Vec_IntAlloc( Vec_PtrSize(vNames) );
    var pName string; var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
        var NameId int32  = Acb_NtkStrId(pNtk, pName);
        var iObjId int32;
        if NameId < 1 {
            fmt.Printf("Cannot find name \"%s\" in the network \"%s\".\n", pName, pNtk.pDesign.pName);
        } else {
            iObjId = Vec_IntEntry( vNamesInv, NameId ); }
        Vec_IntPush( vRes, iObjId );
    }
    return vRes;
}
func Acb_NtkCollectMfsGates(pFileName string, vNamesRefed *Vec_Ptr_t, vNamesDerefed *Vec_Ptr_t, nGates []int32) int32 {
    var pNtkF *Acb_Ntk_t         = Acb_VerilogSimpleRead( pFileName, nil );
    var vNamesInv *Vec_Int_t     = Vec_IntInvert( &pNtkF.vObjName, 0 ) ;
    var vObjsRefed *Vec_Int_t    = Acb_NamesToIds( pNtkF, vNamesInv, vNamesRefed );
    var vObjsDerefed *Vec_Int_t  = Acb_NamesToIds( pNtkF, vNamesInv, vNamesDerefed );
    var vNodes *Vec_Int_t        = Acb_NtkCollectMffc( pNtkF, vObjsRefed, vObjsDerefed );
    var i int32;
    var iObj int32;
    var RetValue int32 = Vec_IntSize(vNodes);
    Vec_IntFree( vNamesInv );
    Vec_IntFree( vObjsRefed );
    Vec_IntFree( vObjsDerefed );
    for  i = 0; i < 5; i++  {
        nGates[i] = 0;
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((iObj) = Vec_IntEntry(vNodes, i)), 1); i++  {
        var nFan int32  = Acb_ObjFaninNum(pNtkF, iObj);
        var Type int32  = Acb_ObjType( pNtkF, iObj );
        if Type == ABC_OPER_CONST_F { 
            nGates[0]++;
        } else if Type == ABC_OPER_CONST_T { 
            nGates[1]++;
        } else if Type == ABC_OPER_BIT_BUF || Type == ABC_OPER_CO { 
            nGates[2]++;
        } else if Type == ABC_OPER_BIT_INV { 
            nGates[3]++;
        } else {
            assert(nFan >= 2, "nFan >= 2");
            nGates[4] += Acb_ObjFaninNum(pNtkF, iObj)-1;
        }
    }
    Vec_IntFree( vNodes );
    Acb_ManFree( pNtkF.pDesign );
    return RetValue;
}
func Acb_NtkReturnMfsGates(pFileName string, vNodes *Vec_Ptr_t) *Vec_Ptr_t {
    var vMffc *Vec_Ptr_t      = Vec_PtrAlloc( 100 );
    var pNtkF *Acb_Ntk_t      = Acb_VerilogSimpleRead( pFileName, nil );
    var vNamesInv *Vec_Int_t  = Vec_IntInvert( &pNtkF.vObjName, 0 ) ;
    var vNodeObjs *Vec_Int_t  = Acb_NamesToIds( pNtkF, vNamesInv, vNodes );
    var vNodeMffc *Vec_Int_t  = Acb_NtkCollectMffc( pNtkF, nil, vNodeObjs );
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodeMffc)) && (((iObj) = Vec_IntEntry(vNodeMffc, i)), 1); i++  {
        Vec_PtrPush( vMffc, Abc_UtilStrsav( Acb_ObjNameStr(pNtkF, iObj) ) );
    }
//Vec_IntPrint( vNodeMffc );
//Vec_PtrPrintNames( vMffc );
    Vec_IntFree( vNodeMffc );
    Vec_IntFree( vNodeObjs );
    Vec_IntFree( vNamesInv );
    Acb_ManFree( pNtkF.pDesign );
    return vMffc;
}

/**Function*************************************************************

  Synopsis    [Computing and updating direct and reverse logic level.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ObjComputeLevelD(p *Acb_Ntk_t, iObj int32) int32 {
    var pFanins *int32;
    var iFanin int32;
    var k int32;
    var Level int32;
    // inlining of macro Acb_ObjForEachFaninFast
    for  k = 0, pFanins = Acb_ObjFanins(p, iObj); k < pFanins[0] && ((iFanin = pFanins[k+1]), 1); k++  {
        Level = Abc_MaxInt( Level, Acb_ObjLevelD(p, iFanin) );
    }
    return Acb_ObjSetLevelD( p, iObj, Level + !Acb_ObjIsCio(p, iObj) );
}
func Acb_NtkComputeLevelD(p *Acb_Ntk_t, vTfo *Vec_Int_t) int32 {
    // it is assumed that vTfo contains CO nodes and level of new nodes was already updated
    var i, iObj, Level int32;
    if !Acb_NtkHasObjLevelD( p ) {
        Acb_NtkCleanObjLevelD( p ); }
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vTfo) - 1; (i >= 0) && (((iObj) = Vec_IntEntry(vTfo, i)), 1); i--  {
        Acb_ObjComputeLevelD( p, iObj );
    }
    // inlining of macro Acb_NtkForEachCo
    for  i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1); i++  { 
        Level = Abc_MaxInt( Level, Acb_ObjLevelD(p, iObj) );
    }
    p.LevelMax = Level;
    return Level;
}

func Acb_ObjComputeLevelR(p *Acb_Ntk_t, iObj int32) int32 {
    var iFanout, k, Level int32;
    // inlining of macro Acb_ObjForEachFanout
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(Vec_WecEntry(&p.vFanouts, iObj))) && (((iFanout) = Vec_IntEntry(Vec_WecEntry(&p.vFanouts, iObj), k)), 1); k++  { if !Acb_ObjType(p, iFanout) {} else { 
        Level = Abc_MaxInt( Level, Acb_ObjLevelR(p, iFanout) ); }
    }
    return Acb_ObjSetLevelR( p, iObj, Level + !Acb_ObjIsCio(p, iObj) );
}
func Acb_NtkComputeLevelR(p *Acb_Ntk_t, vTfi *Vec_Int_t) int32 {
    // it is assumed that vTfi contains CI nodes
    var i, iObj, Level int32;
    if !Acb_NtkHasObjLevelR( p ) {
        Acb_NtkCleanObjLevelR( p ); }
    // inlining of macro Vec_IntForEachEntryReverse
    for  i = Vec_IntSize(vTfi) - 1; (i >= 0) && (((iObj) = Vec_IntEntry(vTfi, i)), 1); i--  {
        Acb_ObjComputeLevelR( p, iObj );
    }
    // inlining of macro Acb_NtkForEachCi
    for  i = 0; (i < Acb_NtkCiNum(p))  && (((iObj) = Acb_NtkCi(p, i)), 1); i++  { 
        Level = Abc_MaxInt( Level, Acb_ObjLevelR(p, iObj) );
    }
//    assert( p->LevelMax == Level );
    p.LevelMax = Level;
    return Level;
}

func Acb_NtkUpdateLevelD(p *Acb_Ntk_t, Pivot int32) {
    var vTfo *Vec_Int_t  = Acb_ObjCollectTfo( p, Pivot, 1 );
    Acb_NtkComputeLevelD( p, vTfo );
}

/**Function*************************************************************

  Synopsis    [Computing and updating direct and reverse path count.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ObjSlack(p *Acb_Ntk_t, iObj int32) int32 {
    var LevelSum int32  = Acb_ObjLevelD(p, iObj) + Acb_ObjLevelR(p, iObj);
    assert(!Acb_ObjIsCio(p, iObj) + p.LevelMax >= LevelSum, "!Acb_ObjIsCio(p, iObj) + p.LevelMax >= LevelSum");
    return !Acb_ObjIsCio(p, iObj) + p.LevelMax - LevelSum;
}

func Acb_ObjComputePathD(p *Acb_Ntk_t, iObj int32) int32 {
    var pFanins *int32;
    var iFanin int32;
    var k int32;
    var Path int32;
    assert(!Acb_ObjIsCi(p, iObj), "!Acb_ObjIsCi(p, iObj)");
    // inlining of macro Acb_ObjForEachFaninFast
    for  k = 0, pFanins = Acb_ObjFanins(p, iObj); k < pFanins[0] && ((iFanin = pFanins[k+1]), 1); k++  {
        if !(Acb_ObjSlack(p, iFanin) != 0) {
            Path += Acb_ObjPathD(p, iFanin); }
    }
    return Acb_ObjSetPathD( p, iObj, Path );
}
func Acb_NtkComputePathsD(p *Acb_Ntk_t, vTfo *Vec_Int_t, fReverse int32) int32 {
    var i, iObj, Path int32;
    //Vec_IntPrint( vTfo );
    if !Acb_NtkHasObjPathD( p ) {
        Acb_NtkCleanObjPathD( p ); }
    // it is assumed that vTfo contains CI nodes
    //assert( Acb_ObjSlack(p, Vec_IntEntry(vTfo, 0)) );
    if fReverse != 0 {
        // inlining of macro Vec_IntForEachEntryReverse
        for  i = Vec_IntSize(vTfo) - 1; (i >= 0) && (((iObj) = Vec_IntEntry(vTfo, i)), 1); i--  {
            if Acb_ObjIsCi(p, iObj) {
                Acb_ObjSetPathD( p, iObj, Acb_ObjSlack(p, iObj) == 0 );
            } else if Acb_ObjSlack(p, iObj) != 0 {
                Acb_ObjSetPathD( p, iObj, 0 );
            } else {
                Acb_ObjComputePathD( p, iObj ); }
        }
    
    } else {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vTfo)) && (((iObj) = Vec_IntEntry(vTfo, i)), 1); i++  {
            if Acb_ObjIsCi(p, iObj) {
                Acb_ObjSetPathD( p, iObj, Acb_ObjSlack(p, iObj) == 0 );
            } else if Acb_ObjSlack(p, iObj) != 0 {
                Acb_ObjSetPathD( p, iObj, 0 );
            } else {
                Acb_ObjComputePathD( p, iObj ); }
        }
    }
    // inlining of macro Acb_NtkForEachCo
    for  i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1); i++  { 
        Path += Acb_ObjPathD(p, iObj);
    }
    p.nPaths = Path;
    return Path;
}

func Acb_ObjComputePathR(p *Acb_Ntk_t, iObj int32) int32 {
    var iFanout, k, Path int32;
    assert(!Acb_ObjIsCo(p, iObj), "!Acb_ObjIsCo(p, iObj)");
    // inlining of macro Acb_ObjForEachFanout
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(Vec_WecEntry(&p.vFanouts, iObj))) && (((iFanout) = Vec_IntEntry(Vec_WecEntry(&p.vFanouts, iObj), k)), 1); k++  { if !Acb_ObjType(p, iFanout) {} else 
        if !(Acb_ObjSlack(p, iFanout) != 0) {
            Path += Acb_ObjPathR(p, iFanout); }
    }
    return Acb_ObjSetPathR( p, iObj, Path );
}
func Acb_NtkComputePathsR(p *Acb_Ntk_t, vTfi *Vec_Int_t, fReverse int32) int32 {
    var i, iObj, Path int32;
    if !Acb_NtkHasObjPathR( p ) {
        Acb_NtkCleanObjPathR( p ); }
    // it is assumed that vTfi contains CO nodes
    //assert( Acb_ObjSlack(p, Vec_IntEntry(vTfi, 0)) );
    if fReverse != 0 {
        // inlining of macro Vec_IntForEachEntryReverse
        for  i = Vec_IntSize(vTfi) - 1; (i >= 0) && (((iObj) = Vec_IntEntry(vTfi, i)), 1); i--  {
            if Acb_ObjIsCo(p, iObj) {
                Acb_ObjSetPathR( p, iObj, Acb_ObjSlack(p, iObj) == 0 );
            } else if Acb_ObjSlack(p, iObj) != 0 {
                Acb_ObjSetPathR( p, iObj, 0 );
            } else {
                Acb_ObjComputePathR( p, iObj ); }
        }
    
    } else {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vTfi)) && (((iObj) = Vec_IntEntry(vTfi, i)), 1); i++  {
            if Acb_ObjIsCo(p, iObj) {
                Acb_ObjSetPathR( p, iObj, Acb_ObjSlack(p, iObj) == 0 );
            } else if Acb_ObjSlack(p, iObj) != 0 {
                Acb_ObjSetPathR( p, iObj, 0 );
            } else {
                Acb_ObjComputePathR( p, iObj ); }
        }
    }
    // inlining of macro Acb_NtkForEachCi
    for  i = 0; (i < Acb_NtkCiNum(p))  && (((iObj) = Acb_NtkCi(p, i)), 1); i++  { 
        Path += Acb_ObjPathR(p, iObj);
    }
//    assert( p->nPaths == Path );
    p.nPaths = Path;
    return Path;
}

func Acb_NtkPrintPaths(p *Acb_Ntk_t) {
    var iObj int32;
    // inlining of macro Acb_NtkForEachObj
    for  iObj = 1; iObj < Vec_StrSize(&p.vObjType); iObj++  { if !Acb_ObjType(p, iObj) {} else {
        fmt.Printf("Obj = %5d :   ", iObj);
        fmt.Printf("LevelD = %5d  ", Acb_ObjLevelD(p, iObj));
        fmt.Printf("LevelR = %5d    ", Acb_ObjLevelR(p, iObj));
        fmt.Printf("PathD = %5d  ", Acb_ObjPathD(p, iObj));
        fmt.Printf("PathR = %5d    ", Acb_ObjPathR(p, iObj));
        fmt.Printf("Paths = %5d  ", Acb_ObjPathD(p, iObj) * Acb_ObjPathR(p, iObj));
        fmt.Printf("\n");
    }
    }
}

func Acb_NtkComputePaths(p *Acb_Ntk_t) int32 {
    var LevelD, LevelR int32;
    var vTfi *Vec_Int_t  = Acb_ObjCollectTfi( p, -1, 1 );
    var vTfo *Vec_Int_t  = Acb_ObjCollectTfo( p, -1, 1 );
    Acb_NtkComputeLevelD( p, vTfo );
    LevelD = p.LevelMax;
    Acb_NtkComputeLevelR( p, vTfi );
    LevelR = p.LevelMax;
    assert(LevelD == LevelR, "LevelD == LevelR");
    Acb_NtkComputePathsD( p, vTfo, 1 );
    Acb_NtkComputePathsR( p, vTfi, 1 );
    return p.nPaths;
}
func Abc_NtkComputePaths(p *Abc_Ntk_t) {
    extern Acb_Ntk_t * Acb_NtkFromAbc( Abc_Ntk_t * p );
    var pNtk *Acb_Ntk_t  = Acb_NtkFromAbc( p );
    Acb_NtkCreateFanout( pNtk );
    Acb_NtkCleanObjCounts( pNtk );
    fmt.Printf("Computed %d paths.\n", Acb_NtkComputePaths(pNtk));
    Acb_NtkPrintPaths( pNtk );
    Acb_ManFree( pNtk.pDesign );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_ObjUpdatePriority(p *Acb_Ntk_t, iObj int32) {
    var nPaths int32;
    if Acb_ObjIsCio(p, iObj) || Acb_ObjLevelD(p, iObj) == 1 {
        return; }
    if p.vQue == nil {
        Acb_NtkCleanObjCounts( p );
        p.vQue = Vec_QueAlloc( 1000 );
        Vec_QueSetPriority( p.vQue, Vec_FltArrayP(&p.vCounts) );
    }
    nPaths = Acb_ObjPathD(p, iObj) * Acb_ObjPathR(p, iObj);
    Acb_ObjSetCounts( p, iObj, float32(nPaths) );
    if Vec_QueIsMember( p.vQue, iObj ) {
//printf( "Updating object %d with count %d\n", iObj, nPaths );
        Vec_QueUpdate( p.vQue, iObj );
    
    } else if nPaths != 0 {
//printf( "Adding object %d with count %d\n", iObj, nPaths );
        Vec_QuePush( p.vQue, iObj );
    }
}
func Acb_NtkUpdateTiming(p *Acb_Ntk_t, iObj int32) {
    var i int32;
    var Entry int32;
    var LevelMax int32 = p.LevelMax;
    var LevelD, LevelR, nPaths1, nPaths2 int32;
    // assuming that direct level of the new nodes (including iObj) is up to date
    var vTfi *Vec_Int_t  = Acb_ObjCollectTfi( p, iObj, 1 );
    var vTfo *Vec_Int_t  = Acb_ObjCollectTfo( p, iObj, 1 );
    if iObj > 0 {
        assert(Vec_IntEntryLast(vTfi) == iObj, "Vec_IntEntryLast(vTfi) == iObj");
        assert(Vec_IntEntryLast(vTfo) == iObj, "Vec_IntEntryLast(vTfo) == iObj");
        Vec_IntPop( vTfo );
    }
    Acb_NtkComputeLevelD( p, vTfo );
    LevelD = p.LevelMax;
    Acb_NtkComputeLevelR( p, vTfi );
    LevelR = p.LevelMax;
    assert(LevelD == LevelR, "LevelD == LevelR");
    if iObj > 0 && LevelMax > p.LevelMax { // reduced level
    
        iObj = -1;
        vTfi = Acb_ObjCollectTfi( p, -1, 1 );
        vTfo = Acb_ObjCollectTfo( p, -1, 1 );   
        Vec_QueClear( p.vQue );
        // add backup here
    }
    if iObj > 0 {
    Acb_NtkComputePathsD( p, vTfi, 0 ); }
    Acb_NtkComputePathsD( p, vTfo, 1 );
    nPaths1 = p.nPaths;
    if iObj > 0 {
    Acb_NtkComputePathsR( p, vTfo, 0 ); }
    Acb_NtkComputePathsR( p, vTfi, 1 );
    nPaths2 = p.nPaths;
    assert(nPaths1 == nPaths2, "nPaths1 == nPaths2");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vTfi)) && (((Entry) = Vec_IntEntry(vTfi, i)), 1); i++  {
        Acb_ObjUpdatePriority( p, Entry );
    }
    if iObj > 0 {
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vTfo)) && (((Entry) = Vec_IntEntry(vTfo, i)), 1); i++  {
        Acb_ObjUpdatePriority( p, Entry );
    } }

//    printf( "Updating timing for object %d.\n", iObj );
//    Acb_NtkPrintPaths( p );
//    while ( (Entry = (int)Vec_QueTopPriority(p->vQue)) > 0 )
//        printf( "Obj = %5d : Prio = %d.\n", Vec_QuePop(p->vQue), Entry );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkPrintNode(p *Acb_Ntk_t, iObj int32) {
    var k int32;
    var iFanin int32;
    var pFanins *int32; 
    fmt.Printf("Node %5d : ", iObj);
    // inlining of macro Acb_ObjForEachFaninFast
    for  k = 0, pFanins = Acb_ObjFanins(p, iObj); k < pFanins[0] && ((iFanin = pFanins[k+1]), 1); k++  {
        fmt.Printf("%d ", iFanin);
    }
    fmt.Printf("LevelD = %d. LevelR = %d.\n", Acb_ObjLevelD(p, iObj), Acb_ObjLevelR(p, iObj));
}
func Acb_NtkCreateNode(p *Acb_Ntk_t, uTruth word, vSupp *Vec_Int_t) int32 {
    var Pivot int32  = Acb_ObjAlloc( p, ABC_OPER_LUT, Vec_IntSize(vSupp), 0 );
    Acb_ObjSetTruth( p, Pivot, uTruth );
    Acb_ObjAddFanins( p, Pivot, vSupp );
    Acb_ObjAddFaninFanout( p, Pivot );
    Acb_ObjComputeLevelD( p, Pivot );
    return Pivot;
}
func Acb_NtkResetNode(p *Acb_Ntk_t, Pivot int32, uTruth word, vSupp *Vec_Int_t) {
    // remember old fanins
    var k int32;
    var iFanin int32;
    var pFanins *int32; 
    var vFanins *Vec_Int_t  = Vec_IntAlloc( 6 );
    assert(!Acb_ObjIsCio(p, Pivot), "!Acb_ObjIsCio(p, Pivot)");
    // inlining of macro Acb_ObjForEachFaninFast
    for  k = 0, pFanins = Acb_ObjFanins(p, Pivot); k < pFanins[0] && ((iFanin = pFanins[k+1]), 1); k++  {
        Vec_IntPush( vFanins, iFanin );
    }
    // update function
    Vec_WrdSetEntry( &p.vObjTruth, Pivot, uTruth );
    Vec_StrErase(&p.vCnfs[Pivot]);
    // remove old fanins
    Acb_ObjRemoveFaninFanout( p, Pivot );
    Acb_ObjRemoveFanins( p, Pivot );
    // add new fanins
    if vSupp != nil {
        assert(Acb_ObjFanoutNum(p, Pivot) > 0, "Acb_ObjFanoutNum(p, Pivot) > 0");
        Acb_ObjAddFanins( p, Pivot, vSupp );
        Acb_ObjAddFaninFanout( p, Pivot );
    
    } else if Acb_ObjFanoutNum(p, Pivot) == 0 {
        Acb_ObjCleanType( p, Pivot ); }
    // delete dangling fanins
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vFanins)) && (((iFanin) = Vec_IntEntry(vFanins, k)), 1); k++  {
        if !Acb_ObjIsCio(p, iFanin) && Acb_ObjFanoutNum(p, iFanin) == 0 {
            Acb_NtkResetNode( p, iFanin, 0, nil ); }
    }
    Vec_IntFree( vFanins );
}
func Acb_NtkSaveSupport(p *Acb_Ntk_t, iObj int32) {
    var k int32;
    var iFanin int32;
    var pFanins *int32; 
    Vec_IntClear( &p.vSuppOld );
    // inlining of macro Acb_ObjForEachFaninFast
    for  k = 0, pFanins = Acb_ObjFanins(p, iObj); k < pFanins[0] && ((iFanin = pFanins[k+1]), 1); k++  {
        Vec_IntPush( &p.vSuppOld, iFanin );
    }
}
func Acb_NtkUpdateNode(p *Acb_Ntk_t, Pivot int32, uTruth word, vSupp *Vec_Int_t) {
    //int Level = Acb_ObjLevelD(p, Pivot);
    Acb_NtkSaveSupport( p, Pivot );
    //Acb_NtkPrintNode( p, Pivot );
    Acb_NtkResetNode( p, Pivot, uTruth, vSupp );
    Acb_ObjComputeLevelD( p, Pivot );
    //assert( Level > Acb_ObjLevelD(p, Pivot) );
    //Acb_NtkPrintNode( p, Pivot );
    if p.vQue == nil {
        Acb_NtkUpdateLevelD( p, Pivot );
    } else {
//        Acb_NtkUpdateTiming( p, Pivot );
        Acb_NtkUpdateTiming( p, -1 ); }
    Vec_IntClear( &p.vSuppOld );
}

/**Function*************************************************************

  Synopsis    [Derive AIG for one network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkFindNodes2_rec(p *Acb_Ntk_t, iObj int32, vNodes *Vec_Int_t) {
    var pFanin *int32;
    var iFanin int32;
    var i int32;
    if Acb_ObjSetTravIdCur(p, iObj) {
        return; }
    if Acb_ObjIsCi(p, iObj) {
        return; }
    // inlining of macro Acb_ObjForEachFaninFast
    for  i = 0, pFanin = Acb_ObjFanins(p, iObj); i < pFanin[0] && ((iFanin = pFanin[i+1]), 1); i++  {
        Acb_NtkFindNodes2_rec( p, iFanin, vNodes );
    }
    assert(!Acb_ObjIsCo(p, iObj), "!Acb_ObjIsCo(p, iObj)");
    Vec_IntPush( vNodes, iObj );
}
func Acb_NtkFindNodes2(p *Acb_Ntk_t) *Vec_Int_t {
    var i, iObj int32;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( 1000 );
    Acb_NtkIncTravId( p );
    // inlining of macro Acb_NtkForEachCo
    for  i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1); i++  { 
        Acb_NtkFindNodes2_rec( p, Acb_ObjFanin(p, iObj, 0), vNodes );
    }
    return vNodes;
}
func Acb_ObjToGia2(pNew *Gia_Man_t, fUseBuf int32, p *Acb_Ntk_t, iObj int32, vTemp *Vec_Int_t, fUseXors int32) int32 {
    //char * pName = Abc_NamStr( p->pDesign->pStrs, Acb_ObjName(p, iObj) );
    var pFanin *int32;
    var iFanin int32;
    var k int32;
    var Type int32;
    var Res int32;
    assert(!Acb_ObjIsCio(p, iObj), "!Acb_ObjIsCio(p, iObj)");
    Vec_IntClear( vTemp );
    // inlining of macro Acb_ObjForEachFaninFast
    for  k = 0, pFanin = Acb_ObjFanins(p, iObj); k < pFanin[0] && ((iFanin = pFanin[k+1]), 1); k++  {
        assert(Acb_ObjCopy(p, iFanin) >= 0, "Acb_ObjCopy(p, iFanin) >= 0");
        Vec_IntPush( vTemp, Acb_ObjCopy(p, iFanin) );
    }
    Type = Acb_ObjType( p, iObj );
    if Type == ABC_OPER_CONST_F { 
        return 0; }
    if Type == ABC_OPER_CONST_T { 
        return 1; }
    if Type == ABC_OPER_BIT_BUF || Type == ABC_OPER_BIT_INV {
        if fUseBuf != 0 {
            Res = Gia_ManAppendBuf(pNew, Vec_IntEntry(vTemp, 0));
        } else {
            Res = Vec_IntEntry(vTemp, 0);
        }
        return Abc_LitNotCond( Res, Type == ABC_OPER_BIT_INV );
    }
    if Type == ABC_OPER_BIT_AND || Type == ABC_OPER_BIT_NAND {
        Res = 1;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vTemp)) && (((iFanin) = Vec_IntEntry(vTemp, k)), 1); k++  {
            Res = Gia_ManAppendAnd2( pNew, Res, iFanin );
        }
        return Abc_LitNotCond( Res, Type == ABC_OPER_BIT_NAND );
    }
    if Type == ABC_OPER_BIT_OR || Type == ABC_OPER_BIT_NOR {
        Res = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vTemp)) && (((iFanin) = Vec_IntEntry(vTemp, k)), 1); k++  {
            Res = Gia_ManAppendOr2( pNew, Res, iFanin );
        }
        return Abc_LitNotCond( Res, Type == ABC_OPER_BIT_NOR );
    }
    if Type == ABC_OPER_BIT_XOR || Type == ABC_OPER_BIT_NXOR {
        Res = 0;
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vTemp)) && (((iFanin) = Vec_IntEntry(vTemp, k)), 1); k++  {
            if fUseXors != 0 {
                Res = Gia_ManAppendXorReal2(pNew, Res, iFanin);
            } else {
                Res = Gia_ManAppendXor2(pNew, Res, iFanin);
            }
        }
        return Abc_LitNotCond( Res, Type == ABC_OPER_BIT_NXOR );
    }
    assert(false, "0");
    return -1;
}
func Acb_NtkToGia2(p *Acb_Ntk_t, fUseBuf int32, fUseXors int32, vTargets *Vec_Int_t, nTargets int32) *Gia_Man_t {
    var pNew, pOne *Gia_Man_t;
    var vFanins, vNodes *Vec_Int_t;
    var i, iObj int32;
    pNew = Gia_ManStart( 2 * Acb_NtkObjNum(p) + 1000 );
    pNew.pName = Abc_UtilStrsav(Acb_NtkName(p));
    Acb_NtkCleanObjCopies( p );
    // inlining of macro Acb_NtkForEachCi
    for  i = 0; (i < Acb_NtkCiNum(p))  && (((iObj) = Acb_NtkCi(p, i)), 1); i++  { 
        Acb_ObjSetCopy( p, iObj, Gia_ManAppendCi(pNew) );
    }
    if vTargets != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vTargets)) && (((iObj) = Vec_IntEntry(vTargets, i)), 1); i++  {
            Acb_ObjSetCopy( p, iObj, Gia_ManAppendCi(pNew) );
        }
    } else {
        for  i = 0; i < nTargets; i++  {
            Gia_ManAppendCi(pNew);
        } }
    vFanins = Vec_IntAlloc( 4 );
    vNodes  = Acb_NtkFindNodes2( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((iObj) = Vec_IntEntry(vNodes, i)), 1); i++  {
        if Acb_ObjCopy(p, iObj) == -1 { // skip targets assigned above
            Acb_ObjSetCopy( p, iObj, Acb_ObjToGia2(pNew, fUseBuf, p, iObj, vFanins, fUseXors) ); }
    }
    Vec_IntFree( vNodes );
    Vec_IntFree( vFanins );
    // inlining of macro Acb_NtkForEachCo
    for  i = 0; (i < Acb_NtkCoNum(p))  && (((iObj) = Acb_NtkCo(p, i)), 1); i++  { 
        Gia_ManAppendCo( pNew, Acb_ObjCopy(p, Acb_ObjFanin(p, iObj, 0)) );
    }
    pNew = Gia_ManCleanup( pOne = pNew );
    Gia_ManUpdateCopy( &p.vObjCopy, pOne );
    Gia_ManStop( pOne );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkCollectCopies(p *Acb_Ntk_t, pGia *Gia_Man_t, pvNodesR **Vec_Ptr_t, pvPolar **Vec_Bit_t) *Vec_Int_t {
    var i int32;
    var iObj int32;
    var iLit int32;
    var nTargets int32 = Vec_IntSize(&p.vTargets);
    var vObjs *Vec_Int_t    = Acb_NtkFindNodes2( p );
    var vNodes *Vec_Int_t   = Vec_IntAlloc( Acb_NtkObjNum(p) );
    var vNodesR *Vec_Ptr_t  = Vec_PtrStart( Gia_ManObjNum(pGia) );
    var vDriver *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var vPolar *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(pGia) );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        if i < Gia_ManCiNum(pGia) - nTargets {
            Vec_PtrWriteEntry( vNodesR, iObj, Abc_UtilStrsav(Acb_ObjNameStr(p, Acb_NtkCi(p, i))) );
        } else {
            Vec_PtrWriteEntry( vNodesR, iObj, Abc_UtilStrsav(Acb_ObjNameStr(p, Vec_IntEntry(&p.vTargets, i-(Gia_ManCiNum(pGia) - nTargets)))) ); }
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        Vec_BitWriteEntry( vDriver, Gia_ObjFaninId0(Gia_ManObj(pGia, iObj), iObj), 1 );
        Vec_PtrWriteEntry( vNodesR, iObj, Abc_UtilStrsav(Acb_ObjNameStr(p, Acb_NtkCo(p, i))) );
        Vec_IntPush( vNodes, iObj );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        if (iLit = Acb_ObjCopy(p, iObj)) >= 0 && Gia_ObjIsAnd(Gia_ManObj(pGia, Abc_Lit2Var(iLit))) {
            if !Vec_BitEntry(vDriver, Abc_Lit2Var(iLit)) && Vec_PtrEntry(vNodesR, Abc_Lit2Var(iLit)) == nil {
                Vec_PtrWriteEntry( vNodesR, Abc_Lit2Var(iLit), Abc_UtilStrsav(Acb_ObjNameStr(p, iObj)) );
                Vec_IntPush( vNodes, Abc_Lit2Var(iLit) );
                Vec_BitWriteEntry( vPolar, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLit) );
            }
        }
    }
    Vec_BitFree( vDriver );
    Vec_IntFree( vObjs );
    Vec_IntSort( vNodes, 0 );
    *pvNodesR = vNodesR;
    *pvPolar  = vPolar;
    return vNodes;
}
func Acb_NtkCollectUser(p *Acb_Ntk_t, vUser *Vec_Ptr_t) *Vec_Int_t {
    var pTemp string; var i int32;
    var vRes *Vec_Int_t  = Vec_IntAlloc( Vec_PtrSize(vUser) );
    var vMap *Vec_Int_t  = Vec_IntStart( Abc_NamObjNumMax(Acb_NtkNam(p)) );
    // inlining of macro Acb_NtkForEachNode
    for  i = 1; i < Vec_StrSize(&p.vObjType); i++  { if !Acb_ObjType(p, i) || Acb_ObjIsCio(p, i) {} else   
        if Acb_ObjName(p, i) > 0 {
            Vec_IntWriteEntry( vMap, Acb_ObjName(p, i), Acb_ObjCopy(p, i) ); }
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vUser)) && (((pTemp) = (char *)Vec_PtrEntry(vUser, i)), 1); i++  {
        if Acb_NtkStrId(p, pTemp) < Vec_IntSize(vMap) {
            var iLit int32  = Vec_IntEntry( vMap, Acb_NtkStrId(p, pTemp) );
            assert(iLit > 0, "iLit > 0");
            //printf( "Obj = %3d  Name = %3d  Copy = %3d\n", i, Acb_NtkStrId(p, pTemp), iLit );
            Vec_IntPush( vRes, Abc_Lit2Var(iLit) );
        }
    }
    assert(Vec_IntSize(vRes) == Vec_PtrSize(vUser), "Vec_IntSize(vRes) == Vec_PtrSize(vUser)");
    Vec_IntFree( vMap );
    Vec_IntUniqify( vRes );
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkExtract(pFileName0 string, pFileName1 string, fUseXors int32, fVerbose int32, ppGiaF **Gia_Man_t, ppGiaG **Gia_Man_t, fUseBuf int32, pvNodes **Vec_Int_t, pvNodesR **Vec_Ptr_t, pvPolar **Vec_Bit_t) int32 {
    var pNtkF *Acb_Ntk_t  = Acb_VerilogSimpleRead( pFileName0, nil );
    var pNtkG *Acb_Ntk_t  = Acb_VerilogSimpleRead( pFileName1, nil );
    var RetValue int32  = -1;
    if pNtkF && pNtkG {
        var nTargets int32  = Vec_IntSize(&pNtkF.vTargets);
        var pGiaF *Gia_Man_t  = Acb_NtkToGia2( pNtkF, fUseBuf, fUseXors, &pNtkF.vTargets, 0 );
        var pGiaG *Gia_Man_t  = Acb_NtkToGia2( pNtkG, 0,       0,        nil, nTargets );
        pGiaF.pSpec = Abc_UtilStrsav( pNtkF.pDesign.pSpec );
        pGiaG.pSpec = Abc_UtilStrsav( pNtkG.pDesign.pSpec );
        assert(Acb_NtkCiNum(pNtkF) == Acb_NtkCiNum(pNtkG), "Acb_NtkCiNum(pNtkF) == Acb_NtkCiNum(pNtkG)");
        assert(Acb_NtkCoNum(pNtkF) == Acb_NtkCoNum(pNtkG), "Acb_NtkCoNum(pNtkF) == Acb_NtkCoNum(pNtkG)");
        *ppGiaF  = pGiaF;
        *ppGiaG  = pGiaG;
        *pvNodes = Acb_NtkCollectCopies( pNtkF, pGiaF, pvNodesR, pvPolar );
        RetValue = nTargets;
    }
    if pNtkF { Acb_ManFree( pNtkF.pDesign ); }
    if pNtkG { Acb_ManFree( pNtkG.pDesign ); }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCollectCopies(p *Abc_Ntk_t, pGia *Gia_Man_t, pvNodesR **Vec_Ptr_t, pvPolar **Vec_Bit_t) *Vec_Int_t {
    var i, iObj, iLit int32;
    var pObj *Abc_Obj_t;
    var vObjs *Vec_Ptr_t    = Abc_NtkDfs( p, 0 );
    var vNodes *Vec_Int_t   = Vec_IntAlloc( Abc_NtkObjNumMax(p) );
    var vNodesR *Vec_Ptr_t  = Vec_PtrStart( Gia_ManObjNum(pGia) );
    var vDriver *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(pGia) );
    var vPolar *Vec_Bit_t   = Vec_BitStart( Gia_ManObjNum(pGia) );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        Vec_PtrWriteEntry( vNodesR, iObj, Abc_UtilStrsav(Abc_ObjName(Abc_NtkCi(p, i))) );
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        Vec_BitWriteEntry( vDriver, Gia_ObjFaninId0(Gia_ManObj(pGia, iObj), iObj), 1 );
        Vec_PtrWriteEntry( vNodesR, iObj, Abc_UtilStrsav(Abc_ObjName(Abc_NtkCo(p, i))) );
        Vec_IntPush( vNodes, iObj );
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjs)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vObjs, i)), 1); i++  {
        if (iLit = pObj.iTemp) >= 0 && Gia_ObjIsAnd(Gia_ManObj(pGia, Abc_Lit2Var(iLit))) {
            if !Vec_BitEntry(vDriver, Abc_Lit2Var(iLit)) && Vec_PtrEntry(vNodesR, Abc_Lit2Var(iLit)) == nil {
                Vec_PtrWriteEntry( vNodesR, Abc_Lit2Var(iLit), Abc_UtilStrsav(Abc_ObjName(pObj)) );
                Vec_IntPush( vNodes, Abc_Lit2Var(iLit) );
                Vec_BitWriteEntry( vPolar, Abc_Lit2Var(iLit), Abc_LitIsCompl(iLit) );
            }
        }
    }
    Vec_BitFree( vDriver );
    Vec_PtrFree( vObjs );
    Vec_IntSort( vNodes, 0 );
    *pvNodesR = vNodesR;
    *pvPolar  = vPolar;
    return vNodes;
}
func Abc_ObjToGia2(pNew *Gia_Man_t, p *Abc_Ntk_t, pObj *Abc_Obj_t, vTemp *Vec_Int_t, fUseXors int32) int32 {
    var pFanin *Abc_Obj_t; var k int32;
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    Vec_IntClear( vTemp );
    // inlining of macro Abc_ObjForEachFanin
    for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
        assert(pFanin.iTemp >= 0, "pFanin.iTemp >= 0");
        Vec_IntPush( vTemp, pFanin.iTemp );
    }
    if Abc_ObjFaninNum(pObj) == 0 {
        if Abc_SopIsConst0( (char*)pObj.pData ) {
            return 0;
        } else {
            return 1;
        } }
    if Abc_ObjFaninNum(pObj) == 1 {
        if Abc_SopIsBuf( (char*)pObj.pData ) {
            return Vec_IntEntry(vTemp, 0);
        } else {
            return Abc_LitNot(Vec_IntEntry(vTemp, 0));
        } }
    if Abc_ObjFaninNum(pObj) == 2 { // nand2
        return Abc_LitNot( Gia_ManAppendAnd2( pNew, Vec_IntEntry(vTemp, 0), Vec_IntEntry(vTemp, 1) ) ); }
    assert(false, "0");
    return -1;
}
func Abc_NtkToGia2(p *Abc_Ntk_t, fUseXors int32) *Gia_Man_t {
    var pNew, pOne *Gia_Man_t;
    var vFanins *Vec_Int_t;
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t; var i int32;
    pNew = Gia_ManStart( 2 * Abc_NtkObjNumMax(p) + 1000 );
    pNew.pName = Abc_UtilStrsav(Abc_NtkName(p));
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {            if (pObj) == nil {} else {
        pObj.iTemp = -1; }
    }
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p)) && (((pObj) = Abc_NtkCi(p, i)), 1); i++  {
        pObj.iTemp = Gia_ManAppendCi(pNew);
    }
    vFanins = Vec_IntAlloc( 4 );
    vNodes  = Abc_NtkDfs( p, 0 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        pObj.iTemp = Abc_ObjToGia2(pNew, p, pObj, vFanins, fUseXors);
    }
    Vec_PtrFree( vNodes );
    Vec_IntFree( vFanins );
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p)) && (((pObj) = Abc_NtkCo(p, i)), 1); i++  {
        Gia_ManAppendCo( pNew, Abc_ObjFanin0(pObj).iTemp );
    }
    pNew = Gia_ManCleanup( pOne = pNew );
    //Gia_ManUpdateCopy( &p->vObjCopy, pOne );
    Gia_ManStop( pOne );
    return pNew;
}
func Abc_NtkExtract(pFileName0 string, pFileName1 string, fUseXors int32, fVerbose int32, ppGiaF **Gia_Man_t, ppGiaG **Gia_Man_t, pvNodes **Vec_Int_t, pvNodesR **Vec_Ptr_t, pvPolar **Vec_Bit_t) int32 {
    var pNtkF *Abc_Ntk_t  = Io_Read( pFileName0, Io_ReadFileType(pFileName0), 1, 0 );
    var pNtkG *Abc_Ntk_t  = Io_Read( pFileName1, Io_ReadFileType(pFileName1), 1, 0 );
    var RetValue int32  = -1;
    if pNtkF && pNtkG {
        var pGiaF *Gia_Man_t  = Abc_NtkToGia2( pNtkF, fUseXors );
        var pGiaG *Gia_Man_t  = Abc_NtkToGia2( pNtkG, 0 );
        assert(Abc_NtkCiNum(pNtkF) == Abc_NtkCiNum(pNtkG), "Abc_NtkCiNum(pNtkF) == Abc_NtkCiNum(pNtkG)");
        assert(Abc_NtkCoNum(pNtkF) == Abc_NtkCoNum(pNtkG), "Abc_NtkCoNum(pNtkF) == Abc_NtkCoNum(pNtkG)");
        pGiaF.pSpec = Abc_UtilStrsav( pNtkF.pSpec );
        pGiaG.pSpec = Abc_UtilStrsav( pNtkG.pSpec );
        *ppGiaF  = pGiaF;
        *ppGiaG  = pGiaG;
        *pvNodes = Abc_NtkCollectCopies( pNtkF, pGiaF, pvNodesR, pvPolar );
        RetValue = 0;
    }
    if pNtkF { Abc_NtkDelete( pNtkF ); }
    if pNtkG { Abc_NtkDelete( pNtkG ); }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkPlaces(pFileName string, vNames *Vec_Ptr_t) *Vec_Int_t {
    var vPlaces *Vec_Int_t; var First int32 = 1;
var Pos int32 = -1;
var fComment int32;
    var pTemp string;
    var pBuffer string = Extra_FileReadContents( pFileName );
    var pLimit string  = pBuffer + len(pBuffer);
    if pBuffer == nil {
        return nil; }
    vPlaces = Vec_IntAlloc( Vec_PtrSize(vNames) );
    for  pTemp = pBuffer; *pTemp; pTemp++  {
        if *pTemp == '\n' {
            fComment = 0; }
        if *pTemp == '/' && *(pTemp + 1) == '/' {
            fComment = 1; }
        if fComment != 0 {
            continue; }

        if *pTemp == '\n' {
            Pos = pTemp - pBuffer + 1;
        } else if *pTemp == '(' {
            if First != 0 {
                First = 0;
            } else {
                var pToken string  = strtok( pTemp+1, "  \n\r\t," );
                var pName string; var i int32;
                // inlining of macro Vec_PtrForEachEntry
                for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
                    if !(int32(strings.Compare(pName, pToken)) != 0) {
                        Vec_IntPushTwo( vPlaces, Pos, i ); }
                }
                pTemp = pToken + len(pToken);
                for  *pTemp == 0  {
                    pTemp++;
                }
                assert(pTemp < pLimit, "pTemp < pLimit");
            }
        }
    }
    assert(Vec_IntSize(vPlaces) == 2*Vec_PtrSize(vNames), "Vec_IntSize(vPlaces) == 2*Vec_PtrSize(vNames)");
    pBuffer = nil;
    return vPlaces;
}
func Acb_NtkInsert(pFileNameIn string, pFileNameOut string, vNames *Vec_Ptr_t, fNumber int32, fSkipMffc int32) {
    var i, k, Prev, Pos, Pos2, iObj int32;
    var vPlaces *Vec_Int_t;
    var pName, pBuffer string;
    var pFile *os.File  = os.OpenFile(pFileNameOut, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileNameOut);
        return;
    }
    pBuffer = Extra_FileReadContents( pFileNameIn );
    if pBuffer == nil {
        _ = (pFile).Close();
        fmt.Printf("Cannot open input file \"%s\".\n", pFileNameIn);
        return;
    }
    if fSkipMffc != 0 {
        var vMffcNames *Vec_Ptr_t  = Acb_NtkReturnMfsGates( pFileNameIn, vNames );
        vPlaces = Acb_NtkPlaces( pFileNameIn, vMffcNames );
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vPlaces)) && (((Pos) = Vec_IntEntry(vPlaces, i)), 1) && (((iObj) = Vec_IntEntry(vPlaces, i+1)), 1); i += 2  {
            for  k = Prev; k < Pos; k++  {
                fputc( pBuffer[k], pFile );
            }
            fmt.Fprintf(pFile, "// MFFC %d = %s //", iObj, (char *)Vec_PtrEntry(vMffcNames, iObj));
            Prev = Pos;
        }
        Vec_IntFree( vPlaces );
        Vec_PtrFreeFree( vMffcNames );
    
    } else {
        vPlaces = Acb_NtkPlaces( pFileNameIn, vNames );
        // inlining of macro Vec_IntForEachEntryDouble
        for  i = 0; (i+1 < Vec_IntSize(vPlaces)) && (((Pos) = Vec_IntEntry(vPlaces, i)), 1) && (((iObj) = Vec_IntEntry(vPlaces, i+1)), 1); i += 2  {
            for  k = Prev; k < Pos; k++  {
                fputc( pBuffer[k], pFile );
            }
            fmt.Fprintf(pFile, "// [t_%d = %s] //", iObj, (char *)Vec_PtrEntry(vNames, iObj));
            Prev = Pos;
        }
        Vec_IntFree( vPlaces );
    }
    pName = strstr(pBuffer, "endmodule");
    Pos2 = pName - pBuffer;
    for  k = Prev; k < Pos2; k++  {
        fputc( pBuffer[k], pFile );
    }
    fmt.Fprintf(pFile, "\n\n");
    fmt.Fprintf(pFile, "  wire ");
    if fNumber != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
            fmt.Fprintf(pFile, " t_%d%s", i, i==Vec_PtrSize(vNames)-1 ? ";" : ",");
        }
    
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
            fmt.Fprintf(pFile, " t%d_%s%s", i, pName, i==Vec_PtrSize(vNames)-1 ? ";" : ",");
        }
    }
    fmt.Fprintf(pFile, "\n\n");
    if fNumber != 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
            fmt.Fprintf(pFile, "  buf( %s, t_%d );\n", pName, i);
        }
    
    } else {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vNames)) && (((pName) = (char *)Vec_PtrEntry(vNames, i)), 1); i++  {
            fmt.Fprintf(pFile, "  buf( %s, t%d_%s );\n", pName, i, pName);
        }
    }
    fmt.Fprintf(pFile, "\n");
    for  k = Pos2; pBuffer[k]; k++  {
        fputc( pBuffer[k], pFile );
    }
    pBuffer = nil;
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_Ntk4CollectAdd(pNtk *Acb_Ntk_t, iObj int32, vRes *Vec_Int_t, vDists *Vec_Int_t, Dist int32) {
    if Acb_ObjSetTravIdCur(pNtk, iObj) {
        return; }
    Vec_IntWriteEntry( vDists, iObj, Dist );
    Vec_IntPush( vRes, iObj );
}
func Acb_Ntk4CollectRing(pNtk *Acb_Ntk_t, vStart *Vec_Int_t, vRes *Vec_Int_t, vDists *Vec_Int_t) {
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vStart)) && (((iObj) = Vec_IntEntry(vStart, i)), 1); i++  {
        var k int32;
        var iFanin int32;
        var pFanins *int32;
        var Weight int32 = Vec_IntEntry(vDists, iObj); 
        // inlining of macro Acb_ObjForEachFaninFast
        for  k = 0, pFanins = Acb_ObjFanins(pNtk, iObj); k < pFanins[0] && ((iFanin = pFanins[k+1]), 1); k++  {
            Acb_Ntk4CollectAdd( pNtk, iFanin, vRes, vDists, Weight + 1*(Acb_ObjFaninNum(pNtk, iObj) > 1) );
        }
        // inlining of macro Acb_ObjForEachFanout
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(Vec_WecEntry(&pNtk.vFanouts, iObj))) && (((iFanin) = Vec_IntEntry(Vec_WecEntry(&pNtk.vFanouts, iObj), k)), 1); k++  { if !Acb_ObjType(pNtk, iFanin) {} else { 
            Acb_Ntk4CollectAdd( pNtk, iFanin, vRes, vDists, Weight + 2*(Acb_ObjFaninNum(pNtk, iObj) > 1) ); }
        }
    }
}
func Acb_Ntk4DumpWeightsInt(pNtk *Acb_Ntk_t, vObjs *Vec_Int_t, pFileName string) {
    var i, iObj int32;//, Count = 0;//, Weight;
    var vDists, vStart, vNexts *Vec_Int_t;
    var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Canont open input file \"%s\".\n", pFileName);
        return;
    }
    vStart = Vec_IntAlloc( 100 );
    vNexts = Vec_IntAlloc( 100 );
    vDists = Vec_IntStart( Acb_NtkObjNumMax(pNtk) );
    Acb_NtkIncTravId( pNtk );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        Acb_ObjSetTravIdCur(pNtk, iObj);
        Vec_IntWriteEntry( vDists, iObj, 1 );
        Vec_IntPush( vStart, iObj );
    }    
    for {
        Acb_Ntk4CollectRing( pNtk, vStart, vNexts, vDists );
        if Vec_IntSize(vNexts) == 0 {
            break; }
        Vec_IntClear( vStart );
        // inlining of macro ABC_SWAP
        { var t Vec_Int_t  = *vStart; *vStart = *vNexts; *vNexts = t; };
    }
    Vec_IntFree( vStart );
    Vec_IntFree( vNexts );
    // create weights
//    Vec_IntForEachEntry( vDists, Weight, i )
//        if ( Weight && Acb_ObjNameStr(pNtk, i)[0] != '1' )
//            fprintf( pFile, "%s %d\n", Acb_ObjNameStr(pNtk, i), 10000+Weight );
/*
    // mark reachable
    Vec_IntClear( &pNtk->vArray0 );
    Acb_NtkIncTravId( pNtk );
    Acb_NtkForEachCo( pNtk, iObj, i )
        if ( !Vec_IntEntry(vStatus, i) )
            Acb_ObjCollectTfi_rec( pNtk, iObj, 0 );
*/
    // inlining of macro Acb_NtkForEachObj
    for  iObj = 1; iObj < Vec_StrSize(&pNtk.vObjType); iObj++  { if !Acb_ObjType(pNtk, iObj) {} else {
        var pName string  = Acb_ObjNameStr(pNtk, iObj);
        var Weight int32  = Vec_IntEntry(vDists, iObj);
        if Weight == 0 {
            Weight = 10000; }
/*
        if ( !Acb_ObjSetTravIdCur(pNtk, iObj) )
        {
            Count++;
            continue;
        }
*/
        fmt.Fprintf(pFile, "%s %d\n", pName, 100000+Weight);
    }
    }
    //printf( "Skipped %d nodes.\n", Count );
    Vec_IntFree( vDists );
    _ = (pFile).Close();
}
func Acb_Ntk4DumpWeights(pFileNameIn string, vObjNames *Vec_Ptr_t, pFileName string) {
    var pName string; var i, iObj int32;
    var vObjs *Vec_Int_t  = Vec_IntAlloc( Vec_PtrSize(vObjNames) );
    var pNtkF *Acb_Ntk_t  = Acb_VerilogSimpleRead( pFileNameIn, nil );
    Acb_NtkCreateFanout( pNtkF );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vObjNames)) && (((pName) = (char *)Vec_PtrEntry(vObjNames, i)), 1); i++  {
        // inlining of macro Acb_NtkForEachObj
        for  iObj = 1; iObj < Vec_StrSize(&pNtkF.vObjType); iObj++  { if !Acb_ObjType(pNtkF, iObj) {} else   
            if !(int32(strings.Compare(Acb_ObjNameStr(pNtkF, iObj), pName)) != 0) {
                Vec_IntPush( vObjs, iObj ); }
        }
    }
    Acb_Ntk4DumpWeightsInt( pNtkF, vObjs, pFileName );
    Acb_ManFree( pNtkF.pDesign );
    Vec_IntFree( vObjs );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Acb_NtkRunSim(pFileName []string, nTimeout int32, nWords int32, nBeam int32, LevL int32, LevU int32, fOrder int32, fFancy int32, fUseBuf int32, fRandom int32, fUseWeights int32, fInputs int32, fSkipMffc int32, fVerbose int32, fVeryVerbose int32) {
    extern int Gia_Sim4Try( char * pFileName0, char * pFileName1, char * pFileName2, int nTimeout, int nWords, int nBeam, int LevL, int LevU, int fOrder, int fFancy, int fUseBuf, int fSkipMffc, int fVerbose );
    extern void Acb_NtkRunEco( char * pFileNames[4], int nTimeout, int fCheck, int fRandom, int fInputs, int fVerbose, int fVeryVerbose );
    var pFileNames [4]string = [4]string{ pFileName[2], pFileName[1], fUseWeights ? (char *)"weights.txt" : nil, pFileName[2] };
    if Gia_Sim4Try( pFileName[0], pFileName[1], pFileName[2], nTimeout, nWords, nBeam, LevL, LevU, fOrder, fFancy, fUseBuf, fSkipMffc, fVerbose ) {
        Acb_NtkRunEco( pFileNames, nTimeout, 1, fRandom, fInputs, fVerbose, fVeryVerbose ); }
} 


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

