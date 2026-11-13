package wln

import "fmt"

/**CFile****************************************************************

  FileName    [wlnObj.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Object construction procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnObj.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creating objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Wln_ObjName_Buffer string // was [100]byte in C

func Wln_ObjName(p *Wln_Ntk_t, iObj int32) string {
    
    if Wln_NtkHasNameId(p) && Wln_ObjNameId(p, iObj) {
        return Abc_NamStr( p.pManName, Wln_ObjNameId(p, iObj) ); }
    spStatic_Wln_ObjName_Buffer = fmt.Sprintf("n%d", iObj);
    return spStatic_Wln_ObjName_Buffer;
}                          
func Wln_ObjConstString(p *Wln_Ntk_t, iObj int32) string {
    assert(Wln_ObjIsConst(p, iObj), "Wln_ObjIsConst(p, iObj)");
    return Abc_NamStr( p.pManName, Wln_ObjFanin0(p, iObj) );
}                          
func Wln_ObjUpdateType(p *Wln_Ntk_t, iObj int32, Type int32) {
    assert(Wln_ObjIsNone(p, iObj), "Wln_ObjIsNone(p, iObj)");
    p.nObjs[Wln_ObjType(p, iObj)]--;
    Vec_IntWriteEntry( &p.vTypes, iObj, Type );
    p.nObjs[Wln_ObjType(p, iObj)]++;
}
func Wln_ObjSetConst(p *Wln_Ntk_t, iObj int32, NameId int32) {
    assert(Wln_ObjIsConst(p, iObj), "Wln_ObjIsConst(p, iObj)");
    Wln_ObjSetFanin( p, iObj, 0, NameId );
}
func Wln_ObjSetSlice(p *Wln_Ntk_t, iObj int32, SliceId int32) {
    assert(Wln_ObjIsSlice(p, iObj), "Wln_ObjIsSlice(p, iObj)");
    Wln_ObjSetFanin( p, iObj, 1, SliceId );
}
func Wln_ObjAddFanin(p *Wln_Ntk_t, iObj, i int32) {
    p.vFanins[iObj].appendFanin(i)
}
func Wln_ObjAddFanins(p *Wln_Ntk_t, iObj int32, vFanins *Vec_Int_t) int32 {
    var i, iFanin int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFanins)) && (((iFanin) = Vec_IntEntry(vFanins, i)), 1); i++  {
        Wln_ObjAddFanin( p, iObj, iFanin );
    }
    return iObj;
}
func Wln_ObjAlloc(p *Wln_Ntk_t, Type int32, Signed int32, End int32, Beg int32) int32 {
    var iObj int32  = Vec_IntSize(&p.vTypes);
    if iObj == Vec_IntCap(&p.vTypes) {
        // inlining of macro ABC_REALLOC
        
        if p.vFanins {
            p.vFanins = ((Wln_Vec_t *) realloc((char *)(p.vFanins), sizeof(Wln_Vec_t) * (size_t)(2 * iObj)));
        } else {
            p.vFanins = ((Wln_Vec_t *) malloc(sizeof(Wln_Vec_t) * (size_t)(2 * iObj)));
        }
        memset( p.vFanins + iObj, 0, sizeof(Wln_Vec_t) * iObj );
        Vec_IntGrow( &p.vTypes, 2 * iObj );
    }
    assert(iObj == Vec_StrSize(&p.vSigns), "iObj == Vec_StrSize(&p.vSigns)");
    assert(iObj == Vec_IntSize(&p.vRanges), "iObj == Vec_IntSize(&p.vRanges)");
    Vec_IntPush( &p.vTypes, Type );
    Vec_StrPush( &p.vSigns, byte(Signed) );
    Vec_IntPush( &p.vRanges, Hash_Int2ManInsert(p.pRanges, End, Beg, 0) );
    if Wln_ObjIsCi(p, iObj) { Wln_ObjSetFanin( p, iObj, 1, Vec_IntSize(&p.vCis) );
Vec_IntPush( &p.vCis, iObj ); }
    if Wln_ObjIsCo(p, iObj) { Wln_ObjSetFanin( p, iObj, 1, Vec_IntSize(&p.vCos) );
Vec_IntPush( &p.vCos, iObj ); }
    if Wln_ObjIsFf(p, iObj) { Vec_IntPush( &p.vFfs, iObj ); }
    p.nObjs[Type]++;
    return iObj;
}
func Wln_ObjClone(pNew *Wln_Ntk_t, p *Wln_Ntk_t, iObj int32) int32 {
    return Wln_ObjAlloc( pNew, Wln_ObjType(p, iObj), Wln_ObjIsSigned(p, iObj), Wln_ObjRangeEnd(p, iObj), Wln_ObjRangeBeg(p, iObj) );
}
func Wln_ObjCreateCo(p *Wln_Ntk_t, iFanin int32) int32 {
    var iCo int32  = Wln_ObjClone( p, p, iFanin );
    Wln_ObjUpdateType( p, iCo, ABC_OPER_CO );
    Wln_ObjAddFanin( p, iCo, iFanin );
    return iCo;
}
func Wln_ObjPrint(p *Wln_Ntk_t, iObj int32) {
    var k int32;
    var iFanin int32;
    var Type int32 = Wln_ObjType(p, iObj);
    fmt.Printf("Obj %6d : Type = %6s  Fanins = %d : ", iObj, Abc_OperName(Type), Wln_ObjFaninNum(p, iObj));
    // inlining of macro Wln_ObjForEachFanin
    for  k = 0; (k < Wln_ObjFaninNum(p, iObj)) && (((iFanin) = Wln_ObjFanin(p, iObj, k)), 1); k++  { if !(iFanin != 0) {} else {
        fmt.Printf("%5d ", iFanin); }
    }
    fmt.Printf("\n");
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

