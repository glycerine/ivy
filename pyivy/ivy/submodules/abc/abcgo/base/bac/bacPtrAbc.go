package bac

import (
	"fmt"
	"strings"
)

/**CFile****************************************************************

  FileName    [bacPtrAbc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Simple interface with external tools.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 29, 2014.]

  Revision    [$Id: bacPtrAbc.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bac.h"
//#include "base/abc/abc.h"
//#include "map/mio/mio.h"
//#include "base/main/mainInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Node type conversions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ptr_HopToType(pObj *Abc_Obj_t) string {
    static word uTruth, uTruths6[3] = {
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
    };
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    uTruth = Hop_ManComputeTruth6( (Hop_Man_t *)Abc_ObjNtk(pObj).pManFunc, (Hop_Obj_t *)pObj.pData, Abc_ObjFaninNum(pObj) );
/*
    if ( uTruth ==  0 )                           return "BAC_BOX_C0";
    if ( uTruth == ~(word)0 )                     return "BAC_BOX_C1";
    if ( uTruth ==  uTruths6[0] )                 return "BAC_BOX_BUF";
    if ( uTruth == ~uTruths6[0] )                 return "BAC_BOX_INV";
    if ( uTruth == (uTruths6[0] & uTruths6[1]) )  return "BAC_BOX_AND";
    if ( uTruth ==~(uTruths6[0] & uTruths6[1]) )  return "BAC_BOX_NAND";
    if ( uTruth == (uTruths6[0] | uTruths6[1]) )  return "BAC_BOX_OR";
    if ( uTruth ==~(uTruths6[0] | uTruths6[1]) )  return "BAC_BOX_NOR";
    if ( uTruth == (uTruths6[0] ^ uTruths6[1]) )  return "BAC_BOX_XOR";
    if ( uTruth ==~(uTruths6[0] ^ uTruths6[1]) )  return "BAC_BOX_XNOR";
*/
    if uTruth ==  0 {                           return "Const0T"; }
    if uTruth == ~(word)0 {                     return "Const1T"; }
    if uTruth ==  uTruths6[0] {                 return "BufT"; }
    if uTruth == ~uTruths6[0] {                 return "InvT"; }
    if uTruth == (uTruths6[0] & uTruths6[1]) {  return "AndT"; }
    if uTruth ==~(uTruths6[0] & uTruths6[1]) {  return "NandT"; }
    if uTruth == (uTruths6[0] | uTruths6[1]) {  return "OrT"; }
    if uTruth ==~(uTruths6[0] | uTruths6[1]) {  return "NorT"; }
    if uTruth == (uTruths6[0] ^ uTruths6[1]) {  return "XorT"; }
    if uTruth ==~(uTruths6[0] ^ uTruths6[1]) {  return "XnorT"; }
    assert(false, "0");
    return nil;
}


/**Function*************************************************************

  Synopsis    [Create Ptr from Abc_Ntk_t.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ptr_AbcObjName(pObj *Abc_Obj_t) string {
    if Abc_ObjIsNet(pObj) || Abc_ObjIsBox(pObj) {
        return Abc_ObjName(pObj); }
    if Abc_ObjIsCi(pObj) || Abc_ObjIsNode(pObj) {
        return Ptr_AbcObjName(Abc_ObjFanout0(pObj)); }
    if Abc_ObjIsCo(pObj) {
        return Ptr_AbcObjName(Abc_ObjFanin0(pObj)); }
    assert(false, "0");
    return nil;
}
func Ptr_CheckArray(vArray *Vec_Ptr_t) int32 {
    assert(Vec_PtrSize(vArray) == Vec_PtrCap(vArray), "Vec_PtrSize(vArray) == Vec_PtrCap(vArray)");
    return 1;
}
func Ptr_AbcDeriveNode(pObj *Abc_Obj_t) *Vec_Ptr_t {
    var pFanin *Abc_Obj_t; var i int32;
    var vNode *Vec_Ptr_t  = Vec_PtrAllocExact( 2 + 2 * (1 + Abc_ObjFaninNum(pObj)) );
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    if Abc_NtkHasAig(pObj.pNtk) {
        Vec_PtrPush( vNode, Ptr_HopToType(pObj) );
    } else if Abc_NtkHasSop(pObj.pNtk) {
        Vec_PtrPush( vNode, Ptr_SopToTypeName((char *)pObj.pData) );
    } else { assert(false, "0"); }
    Vec_PtrPush( vNode, Ptr_AbcObjName(pObj) );
    assert(Abc_ObjFaninNum(pObj) <= 2, "Abc_ObjFaninNum(pObj) <= 2");
    Abc_ObjForEachFanin( pObj, pFanin, i )
    {
        Vec_PtrPush( vNode, (void*)(i ? "r" : "l") );
        Vec_PtrPush( vNode, Ptr_AbcObjName(pFanin) );
    }
    Vec_PtrPush( vNode, (void*)("o") );
    Vec_PtrPush( vNode, Ptr_AbcObjName(pObj) );
    assert((Ptr_CheckArray(vNode) != 0), "Ptr_CheckArray(vNode)");
    return vNode;
}
func Ptr_AbcDeriveBox(pObj *Abc_Obj_t) *Vec_Ptr_t {
    var pNext *Abc_Obj_t; var i int32;
    var pNtk *Abc_Ntk_t  = Abc_ObjModel(pObj);
    var vBox *Vec_Ptr_t  = Vec_PtrAllocExact( 2 + 2 * Abc_ObjFaninNum(pObj) + 2 * Abc_ObjFanoutNum(pObj) );
    assert(Abc_ObjIsBox(pObj), "Abc_ObjIsBox(pObj)");
    Vec_PtrPush( vBox, Abc_NtkName(pNtk) );
    Vec_PtrPush( vBox, Ptr_AbcObjName(pObj) );
    Abc_ObjForEachFanin( pObj, pNext, i )
    {
        Vec_PtrPush( vBox, Ptr_AbcObjName(Abc_NtkPi(pNtk, i)) );
        Vec_PtrPush( vBox, Ptr_AbcObjName(pNext) );
    }
    Abc_ObjForEachFanout( pObj, pNext, i )
    {
        Vec_PtrPush( vBox, Ptr_AbcObjName(Abc_NtkPo(pNtk, i)) );
        Vec_PtrPush( vBox, Ptr_AbcObjName(pNext) );
    }
    assert((Ptr_CheckArray(vBox) != 0), "Ptr_CheckArray(vBox)");
    return vBox;
}
func Ptr_AbcDeriveBoxes(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var pObj *Abc_Obj_t; var i int32;
    var vBoxes *Vec_Ptr_t  = Vec_PtrAllocExact( Abc_NtkBoxNum(pNtk) + Abc_NtkNodeNum(pNtk) );
    Abc_NtkForEachBox( pNtk, pObj, i )
        Vec_PtrPush( vBoxes, Ptr_AbcDeriveBox(pObj) );
    Abc_NtkForEachNode( pNtk, pObj, i )
        Vec_PtrPush( vBoxes, Ptr_AbcDeriveNode(pObj) );
    assert((Ptr_CheckArray(vBoxes) != 0), "Ptr_CheckArray(vBoxes)");
    return vBoxes;
}

func Ptr_AbcDeriveInputs(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var pObj *Abc_Obj_t; var i int32;
    var vSigs *Vec_Ptr_t  = Vec_PtrAllocExact( Abc_NtkPiNum(pNtk) );
    Abc_NtkForEachPi( pNtk, pObj, i )
        Vec_PtrPush( vSigs, Ptr_AbcObjName(pObj) );
    assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)");
    return vSigs;
}
func Ptr_AbcDeriveOutputs(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var pObj *Abc_Obj_t; var i int32;
    var vSigs *Vec_Ptr_t  = Vec_PtrAllocExact( Abc_NtkPoNum(pNtk) );
    Abc_NtkForEachPo( pNtk, pObj, i )
        Vec_PtrPush( vSigs, Ptr_AbcObjName(pObj) );
    assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)");
    return vSigs;
}
func Ptr_AbcDeriveNtk(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vNtk *Vec_Ptr_t  = Vec_PtrAllocExact( 5 );
    Vec_PtrPush( vNtk, Abc_NtkName(pNtk) );
    Vec_PtrPush( vNtk, Ptr_AbcDeriveInputs(pNtk) );
    Vec_PtrPush( vNtk, Ptr_AbcDeriveOutputs(pNtk) );
    Vec_PtrPush( vNtk, Vec_PtrAllocExact(0) );
    Vec_PtrPush( vNtk, Ptr_AbcDeriveBoxes(pNtk) );
    assert((Ptr_CheckArray(vNtk) != 0), "Ptr_CheckArray(vNtk)");
    return vNtk;
}
func Ptr_AbcDeriveDes(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vDes *Vec_Ptr_t;
    var pTemp *Abc_Ntk_t; var i int32;
    vDes = Vec_PtrAllocExact( 1 + Vec_PtrSize(pNtk.pDesign.vModules) );
    Vec_PtrPush( vDes, pNtk.pDesign.pName );
    Vec_PtrForEachEntry( Abc_Ntk_t *, pNtk.pDesign.vModules, pTemp, i )
        Vec_PtrPush( vDes, Ptr_AbcDeriveNtk(pTemp) );
    assert((Ptr_CheckArray(vDes) != 0), "Ptr_CheckArray(vDes)");
    return vDes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ptr_ManExperiment(pNtk *Abc_Ntk_t) {
    var clk abctime  = Abc_Clock();
    var pFileName string  = Extra_FileNameGenericAppend(pNtk.pDesign.pName, "_out.blif");
    var vDes *Vec_Ptr_t  = Ptr_AbcDeriveDes( pNtk );
    fmt.Printf("Converting to Ptr:  Memory = %6.3f MB  ", 1.0*Bac_PtrMemory(vDes)/(1<<20));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Bac_PtrDumpBlif( pFileName, vDes );
    fmt.Printf("Finished writing output file \"%s\".  ", pFileName);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Bac_PtrFree( vDes );
}

/**Function*************************************************************

  Synopsis    [Create Bac_Man_t from tech-ind Ptr.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ptr_NameToType(pSop string) int32 {
    if !(int32(strings.Compare(pSop, "Const0T")) != 0) {      return BAC_BOX_CF; }
    if !(int32(strings.Compare(pSop, "Const1T")) != 0) {      return BAC_BOX_CT; }
    if !(int32(strings.Compare(pSop, "BufT")) != 0) {         return BAC_BOX_BUF; }
    if !(int32(strings.Compare(pSop, "InvT")) != 0) {         return BAC_BOX_INV; }
    if !(int32(strings.Compare(pSop, "AndT")) != 0) {         return BAC_BOX_AND; }
    if !(int32(strings.Compare(pSop, "NandT")) != 0) {        return BAC_BOX_NAND; }
    if !(int32(strings.Compare(pSop, "OrT")) != 0) {          return BAC_BOX_OR; }
    if !(int32(strings.Compare(pSop, "NorT")) != 0) {         return BAC_BOX_NOR; }
    if !(int32(strings.Compare(pSop, "XorT")) != 0) {         return BAC_BOX_XOR; }
    if !(int32(strings.Compare(pSop, "XnorT")) != 0) {        return BAC_BOX_XNOR; }
    return BAC_OBJ_BOX;
}
func Ptr_ManCountNtk(vNtk *Vec_Ptr_t) int32 {
    var vInputs *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1);
    var vOutputs *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2);
    var vNodes *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 3);
    var vBoxes *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 4);
    var vBox *Vec_Ptr_t; var i, Counter int32;
    assert(Vec_PtrSize(vNodes) == 0, "Vec_PtrSize(vNodes) == 0");
    Counter += Vec_PtrSize(vInputs);
    Counter += Vec_PtrSize(vOutputs);
    Vec_PtrForEachEntry( Vec_Ptr_t *, vBoxes, vBox, i )
        Counter += Vec_PtrSize(vBox)/2;
    return Counter;
}
func Bac_BoxCountOutputs(pNtk *Bac_Ntk_t, pBoxNtk string) int32 {
    var ModuleId int32  = Bac_ManNtkFindId( pNtk.pDesign, pBoxNtk );
    if ModuleId == 0 {
        return 1; }
    return Bac_NtkPoNumAlloc( Bac_ManNtk(pNtk.pDesign, ModuleId) );
}
func Bac_NtkDeriveFromPtr(pNtk *Bac_Ntk_t, vNtk *Vec_Ptr_t, vMap *Vec_Int_t, vBox2Id *Vec_Int_t) int32 {
    var pName string;
    var pModuleName string = (char *)Vec_PtrEntry(vNtk, 0);
    var vInputs *Vec_Ptr_t   = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1);
    var vOutputs *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2);
    var vBoxes *Vec_Ptr_t = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 4);
    var vBox *Vec_Ptr_t;
    var i, k, iObj, iTerm, NameId int32;
    // start network with the given name
    NameId = Abc_NamStrFindOrAdd( pNtk.pDesign.pStrs, pModuleName, nil );
    assert(Bac_NtkNameId(pNtk) == NameId, "Bac_NtkNameId(pNtk) == NameId");
    // map driven NameIds into their ObjIds for PIs
    Vec_PtrForEachEntry( char *, vInputs, pName, i )
    {
        NameId = Abc_NamStrFindOrAdd( pNtk.pDesign.pStrs, pName, nil );
        if Vec_IntGetEntryFull(vMap, NameId) != -1 { fmt.Printf("PI with name \"%s\" is not unique module \"%s\".\n", pName, pModuleName); return 0; }
        iObj = Bac_ObjAlloc( pNtk, BAC_OBJ_PI, -1 );
        Bac_ObjSetName( pNtk, iObj, Abc_Var2Lit2(NameId, BAC_NAME_BIN) );
        Vec_IntSetEntryFull( vMap, NameId, iObj );
        Bac_NtkAddInfo( pNtk, Abc_Var2Lit2(NameId, 1), -1, -1 );
    }
    // map driven NameIds into their ObjIds for BOs
    Vec_IntClear( vBox2Id );
    Vec_PtrForEachEntry( Vec_Ptr_t *, vBoxes, vBox, i )
    {
        var pBoxNtk string  = (char *)Vec_PtrEntry(vBox, 0);
        var pBoxName string  = (char *)Vec_PtrEntry(vBox, 1);
        var nOutputs int32  = Bac_BoxCountOutputs( pNtk, pBoxNtk );
        var nInputs int32  = Vec_PtrSize(vBox)/2 - nOutputs - 1;
        var NtkId int32  = Bac_ManNtkFindId( pNtk.pDesign, pBoxNtk );
        assert(Vec_PtrSize(vBox) % 2 == 0, "Vec_PtrSize(vBox) % 2 == 0");
        assert(nOutputs > 0 && 2*(nOutputs + 1) <= Vec_PtrSize(vBox), "nOutputs > 0 && 2*(nOutputs + 1) <= Vec_PtrSize(vBox)");
        iObj = Bac_BoxAlloc( pNtk, (Bac_ObjType_t)Ptr_NameToType(pBoxNtk), nInputs, nOutputs, NtkId );
        if NtkId > 0 {
            Bac_NtkSetHost( Bac_ManNtk(pNtk.pDesign, NtkId), Bac_NtkId(pNtk), iObj ); }
        Bac_ObjSetName( pNtk, iObj, Abc_Var2Lit2(Abc_NamStrFindOrAdd(pNtk.pDesign.pStrs, pBoxName, nil), BAC_NAME_BIN) );
        Bac_BoxForEachBo( pNtk, iObj, iTerm, k )
        {
            pName = (char *)Vec_PtrEntry( vBox, Vec_PtrSize(vBox) - 2*(nOutputs - k) + 1 );
            NameId = Abc_NamStrFindOrAdd( pNtk.pDesign.pStrs, pName, nil );
            if Vec_IntGetEntryFull(vMap, NameId) != -1 { fmt.Printf("Signal \"%s\" has multiple drivers in module \"%s\".\n", pName, pModuleName); return 0; }
            Bac_ObjSetName( pNtk, iTerm, Abc_Var2Lit2(NameId, BAC_NAME_BIN) );
            Vec_IntSetEntryFull( vMap, NameId, iTerm );
        }
        Vec_IntPush( vBox2Id, iObj );
    }
    assert(Vec_IntSize(vBox2Id) == Vec_PtrSize(vBoxes), "Vec_IntSize(vBox2Id) == Vec_PtrSize(vBoxes)");
    // connect BIs
    Vec_PtrForEachEntry( Vec_Ptr_t *, vBoxes, vBox, i )
    {
        iObj = Vec_IntEntry( vBox2Id, i );
        Bac_BoxForEachBi( pNtk, iObj, iTerm, k )
        {
            pName = (char *)Vec_PtrEntry( vBox, 2*(k + 1) + 1 );
            NameId = Abc_NamStrFindOrAdd( pNtk.pDesign.pStrs, pName, nil );
            if Vec_IntGetEntryFull(vMap, NameId) == -1 {
                fmt.Printf("Signal \"%s\" in not driven in module \"%s\".\n", pName, pModuleName); }
            Bac_ObjSetFanin( pNtk, iTerm, Vec_IntGetEntryFull(vMap, NameId) );
        }
    }
    // connect POs
    Vec_PtrForEachEntry( char *, vOutputs, pName, i )
    {
        NameId = Abc_NamStrFindOrAdd( pNtk.pDesign.pStrs, pName, nil );
        if Vec_IntGetEntryFull(vMap, NameId) == -1 {
            fmt.Printf("PO with name \"%s\" in not driven in module \"%s\".\n", pName, pModuleName); }
        iObj = Bac_ObjAlloc( pNtk, BAC_OBJ_PO, Vec_IntGetEntryFull(vMap, NameId) );
        Bac_NtkAddInfo( pNtk, Abc_Var2Lit2(NameId, 2), -1, -1 );
    }
    // update map
    Bac_NtkForEachCi( pNtk, iObj )
        Vec_IntSetEntryFull( vMap, Bac_ObjNameId(pNtk, iObj), -1 );
    // double check
    Vec_IntForEachEntry( vMap, iObj, i )
        assert(iObj == -1, "iObj == -1");
    assert(Bac_NtkObjNum(pNtk) == Vec_StrCap(&pNtk.vType), "Bac_NtkObjNum(pNtk) == Vec_StrCap(&pNtk.vType)");
    return 1;
}
func Bac_PtrTransformToCba(vDes *Vec_Ptr_t) *Bac_Man_t {
    var pName string  = (char *)Vec_PtrEntry(vDes, 0);
    var pNew *Bac_Man_t  = Bac_ManAlloc( pName, Vec_PtrSize(vDes) - 1 );
    var vMap *Vec_Int_t  = Vec_IntStartFull( 1000 );
    var vBox2Id *Vec_Int_t  = Vec_IntAlloc( 1000 );
    // create interfaces
    var pNtk *Bac_Ntk_t; var i int32;
    Bac_ManForEachNtk( pNew, pNtk, i )
    {
        var vNtk *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vDes, i);
        var vInputs *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 1);
        var vOutputs *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vNtk, 2);
        var NameId int32  = Abc_NamStrFindOrAdd( pNew.pStrs, (char *)Vec_PtrEntry(vNtk, 0), nil );
        Bac_NtkAlloc( pNtk, NameId, Vec_PtrSize(vInputs), Vec_PtrSize(vOutputs), Ptr_ManCountNtk(vNtk) );
        Bac_NtkStartNames( pNtk );
    }
    // parse the networks
    Bac_ManForEachNtk( pNew, pNtk, i )
    {
        var vNtk *Vec_Ptr_t  = (Vec_Ptr_t *)Vec_PtrEntry(vDes, i);
        if !(Bac_NtkDeriveFromPtr( pNtk, vNtk, vMap, vBox2Id ) != 0) {
            break; }
    }
    if i <= Bac_ManNtkNum(pNew) {
       Bac_ManFree(pNew);
pNew = nil; }
    Vec_IntFree( vBox2Id );
    Vec_IntFree( vMap );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Create Ptr from mapped Bac_Man_t.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bac_NtkTransformToPtrBox(p *Bac_Ntk_t, iBox int32) *Vec_Ptr_t {
    var i int32;
    var iTerm int32;
    var fUser int32 = Bac_ObjIsBoxUser( p, iBox );
    var pBoxNtk *Bac_Ntk_t  = Bac_BoxNtk( p, iBox );
    var pLib *Mio_Library_t  = (Mio_Library_t *)p.pDesign.pMioLib;
    var pGate *Mio_Gate_t;
    if pLib != nil {
        pGate = Mio_LibraryReadGateByName( pLib, Bac_BoxNtkName(p, iBox), nil );
    } else {
        pGate = nil;
    }
    var vBox *Vec_Ptr_t  = Vec_PtrAllocExact( 2*Bac_BoxSize(p, iBox) );
    Vec_PtrPush( vBox, Bac_BoxNtkName(p, iBox) );
    Vec_PtrPush( vBox, Bac_ObjNameStr(p, iBox) );
    Bac_BoxForEachBi( p, iBox, iTerm, i )
    {
        Vec_PtrPush( vBox, fUser ? Bac_ObjNameStr(pBoxNtk, Bac_NtkPi(pBoxNtk, i)) : Mio_GateReadPinName(pGate, i) );
        Vec_PtrPush( vBox, Bac_ObjNameStr(p, iTerm) );
    }
    Bac_BoxForEachBo( p, iBox, iTerm, i )
    {
        Vec_PtrPush( vBox, fUser ? Bac_ObjNameStr(pBoxNtk, Bac_NtkPo(pBoxNtk, i)) : Mio_GateReadOutName(pGate) );
        Vec_PtrPush( vBox, Bac_ObjNameStr(p, iTerm) );
    }
    assert((Ptr_CheckArray(vBox) != 0), "Ptr_CheckArray(vBox)");
    return vBox;
}
func Bac_NtkTransformToPtrBoxes(p *Bac_Ntk_t) *Vec_Ptr_t {
    var iBox int32;
    var vBoxes *Vec_Ptr_t  = Vec_PtrAllocExact( Bac_NtkBoxNum(p) );
    Bac_NtkForEachBox( p, iBox )
        Vec_PtrPush( vBoxes, Bac_NtkTransformToPtrBox(p, iBox) );
    assert((Ptr_CheckArray(vBoxes) != 0), "Ptr_CheckArray(vBoxes)");
    return vBoxes;
}

func Bac_NtkTransformToPtrInputs(p *Bac_Ntk_t) *Vec_Ptr_t {
    var i, iTerm int32;
    var vSigs *Vec_Ptr_t  = Vec_PtrAllocExact( Bac_NtkPiNum(p) );
    Bac_NtkForEachPi( p, iTerm, i )
        Vec_PtrPush( vSigs, Bac_ObjNameStr(p, iTerm) );
    assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)");
    return vSigs;
}
func Bac_NtkTransformToPtrOutputs(p *Bac_Ntk_t) *Vec_Ptr_t {
    var i, iTerm int32;
    var vSigs *Vec_Ptr_t  = Vec_PtrAllocExact( Bac_NtkPoNum(p) );
    Bac_NtkForEachPo( p, iTerm, i )
        Vec_PtrPush( vSigs, Bac_ObjNameStr(p, iTerm) );
    assert((Ptr_CheckArray(vSigs) != 0), "Ptr_CheckArray(vSigs)");
    return vSigs;
}
func Bac_NtkTransformToPtr(p *Bac_Ntk_t) *Vec_Ptr_t {
    var vNtk *Vec_Ptr_t  = Vec_PtrAllocExact(5);
    Vec_PtrPush( vNtk, Bac_NtkName(p) );
    Vec_PtrPush( vNtk, Bac_NtkTransformToPtrInputs(p) );
    Vec_PtrPush( vNtk, Bac_NtkTransformToPtrOutputs(p) );
    Vec_PtrPush( vNtk, Vec_PtrAllocExact(0) );
    Vec_PtrPush( vNtk, Bac_NtkTransformToPtrBoxes(p) );
    assert((Ptr_CheckArray(vNtk) != 0), "Ptr_CheckArray(vNtk)");
    return vNtk;
}
func Bac_PtrDeriveFromCba(p *Bac_Man_t) *Vec_Ptr_t {
    var vDes *Vec_Ptr_t;
    var pTemp *Bac_Ntk_t; var i int32;
    if p == nil {
        return nil; }
    if p.pMioLib == nil {
        fmt.Printf("Cannot transform CBA network into Ptr because it is not mapped.\n");
        return nil;
    }
    Bac_ManAssignInternWordNames( p );
    vDes = Vec_PtrAllocExact( 1 + Bac_ManNtkNum(p) );
    Vec_PtrPush( vDes, p.pName );
    Bac_ManForEachNtk( p, pTemp, i )
        Vec_PtrPush( vDes, Bac_NtkTransformToPtr(pTemp) );
    assert((Ptr_CheckArray(vDes) != 0), "Ptr_CheckArray(vDes)");
    return vDes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

