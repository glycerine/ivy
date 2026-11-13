package wln

import "fmt"

/**CFile****************************************************************

  FileName    [wlnWlc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Network transformation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnWlc.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"
//#include "base/wlc/wlc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern int Ndr_TypeWlc2Ndr( int Type );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_TrasformNameId(pNew *Wln_Ntk_t, p *Wlc_Ntk_t, pObj *Wlc_Obj_t) int32 {
    return Abc_NamStrFindOrAdd( pNew.pManName, Wlc_ObjName(p, Wlc_ObjId(p, pObj)), nil );
}
func Wln_NtkFromWlc(p *Wlc_Ntk_t) *Wln_Ntk_t {
    var pObj *Wlc_Obj_t; 
    var Buffer string // was [1000]byte in C

    var i, j, n, Type, iFanin, iOutId, iBit int32;
    var vFanins *Vec_Int_t  = Vec_IntAlloc( 10 );
    var vInits *Vec_Int_t  = Vec_IntAlloc( Wlc_NtkFfNum(p) );
    var pNew *Wln_Ntk_t  = Wln_NtkAlloc( p.pName, Wlc_NtkObjNum(p)+Wlc_NtkCoNum(p)+Wlc_NtkFfNum(p) );
    pNew.pManName = Abc_NamStart( Abc_NamObjNumMax(p.pManName), 10 );
    if p.pSpec { pNew.pSpec = Abc_UtilStrsav( p.pSpec ); }
    pNew.fSmtLib = p.fSmtLib;
    Wlc_NtkCleanCopy( p );
    Wln_NtkCleanNameId( pNew );
    // add primary inputs
    // inlining of macro Wlc_NtkForEachPi
    for  i = 0; (i < Wlc_NtkPiNum(p)) && (((pObj) = Wlc_NtkPi(p, i)), 1); i++  {
        iOutId = Wln_ObjAlloc( pNew, ABC_OPER_CI, pObj.Signed, pObj.End, pObj.Beg );
        Wln_ObjSetNameId( pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj) );
        Wlc_ObjSetCopy( p, Wlc_ObjId(p, pObj), iOutId );
    }
    // create initial state of the flops
    // inlining of macro Wlc_NtkForEachCi
    for  i = 0; (i < Wlc_NtkCiNum(p)) && (((pObj) = Wlc_NtkCi(p, i)), 1); i++  {
        assert(i == Wlc_ObjCiId(pObj), "i == Wlc_ObjCiId(pObj)");
        if pObj.Type == WLC_OBJ_PI {
            continue; }
        for  j = 0; j < Wlc_ObjRange(pObj); j++  {
            if p.pInits[iBit+j] == 'x' {
                break; }
        }
        // print flop init state
        if true {
            fmt.Printf("Flop %3d init state: %d\'b", i-Wlc_NtkPiNum(p), Wlc_ObjRange(pObj));
            if j == Wlc_ObjRange(pObj) {
                var Count int32;
                for  n = 0; n < Wlc_ObjRange(pObj); n++  {
                    Count += p.pInits[iBit+n] == '0';
                }
                if Count == Wlc_ObjRange(pObj) {
                    fmt.Printf("0");
                } else {
                    for  n = 0; n < Wlc_ObjRange(pObj); n++  {
                        fmt.Printf("%c", p.pInits[iBit+n]);
                    } }
            
            } else {
                var Count int32;
                for  n = 0; n < Wlc_ObjRange(pObj); n++  {
                    Count += p.pInits[iBit+n] == 'x';
                }
                fmt.Printf("x");
                if Count != Wlc_ObjRange(pObj) {
                    fmt.Printf(" (range %d)", Wlc_ObjRange(pObj)); }
            }
            fmt.Printf("\n");
        }
        if j == Wlc_ObjRange(pObj) {
            Type = ABC_OPER_CONST;
        } else {
            Type = ABC_OPER_CI;
        } 
        iOutId = Wln_ObjAlloc( pNew, Type, pObj.Signed, pObj.End, pObj.Beg );
        if j == Wlc_ObjRange(pObj) { // constant
        
            var pString string  = Wln_ConstFromStr(p.pInits + iBit, Wlc_ObjRange(pObj));
            Wln_ObjSetConst( pNew, iOutId, Abc_NamStrFindOrAdd(pNew.pManName, pString, nil) );
            ABC_FREE( pString );
        }
        Buffer = fmt.Sprintf("ff_init_%d", Vec_IntSize(vInits));
        Wln_ObjSetNameId( pNew, iOutId, Abc_NamStrFindOrAdd(pNew.pManName, Buffer, nil) );
        Vec_IntPush( vInits, iOutId );
        iBit += Wlc_ObjRange(pObj);
    }
    assert(p.pInits == nil || iBit == int32(len(p.pInits)), "p.pInits == nil || iBit == int32(strlen(p.pInits))");
    // add flop outputs
    // inlining of macro Wlc_NtkForEachCi
    for  i = 0; (i < Wlc_NtkCiNum(p)) && (((pObj) = Wlc_NtkCi(p, i)), 1); i++  {
        assert(i == Wlc_ObjCiId(pObj), "i == Wlc_ObjCiId(pObj)");
        if pObj.Type == WLC_OBJ_PI {
            continue; }
        iOutId = Wln_ObjAlloc( pNew, ABC_OPER_DFFRSE, pObj.Signed, pObj.End, pObj.Beg );
        Wln_ObjSetNameId( pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj) );
        Wlc_ObjSetCopy( p, Wlc_ObjId(p, pObj), iOutId );
    }
    // add internal nodes 
    // inlining of macro Wlc_NtkForEachObj
    for  i = 1; (i < Wlc_NtkObjNumMax(p)) && (((pObj) = Wlc_NtkObj(p, i)), 1); i++  {
        if Wlc_ObjIsCi(pObj) || pObj.Type == 0 {
            continue; }
        iOutId = Wln_ObjAlloc( pNew, Ndr_TypeWlc2Ndr(pObj.Type), pObj.Signed, pObj.End, pObj.Beg );
        Vec_IntClear( vFanins );
        // inlining of macro Wlc_ObjForEachFanin
        for  n = 0; (n < Wlc_ObjFaninNum(pObj)) && (((iFanin) = Wlc_ObjFaninId(pObj, n)), 1); n++  {
            Vec_IntPush( vFanins, Wlc_ObjCopy(p, iFanin) );
        }
        Wln_ObjAddFanins( pNew, iOutId, vFanins );
        if pObj.Type == WLC_OBJ_BIT_SELECT {
            Wln_ObjSetSlice( pNew, iOutId, Hash_Int2ManInsert(pNew.pRanges, pObj.End, pObj.Beg, 0) );
        } else if pObj.Type == WLC_OBJ_CONST {
            var pString string  = Wln_ConstFromBits(Wlc_ObjConstValue(pObj), Wlc_ObjRange(pObj));
            Wln_ObjSetConst( pNew, iOutId, Abc_NamStrFindOrAdd(pNew.pManName, pString, nil) );
            ABC_FREE( pString );
        }
//        else if ( Type == ABC_OPER_BIT_MUX && Vec_IntSize(vFanins) == 3 )
//            ABC_SWAP( int, Wln_ObjFanins(p, iObj)[1], Wln_ObjFanins(p, iObj)[2] );
        Wln_ObjSetNameId( pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj) );
        Wlc_ObjSetCopy( p, i, iOutId );
    }
    // inlining of macro Wlc_NtkForEachPo
    for  i = 0; (i < Wlc_NtkPoNum(p)) && (((pObj) = Wlc_NtkPo(p, i)), 1); i++  {
        iOutId = Wln_ObjAlloc( pNew, ABC_OPER_CO, pObj.Signed, pObj.End, pObj.Beg );
        Wln_ObjAddFanin( pNew, iOutId, Wlc_ObjCopy(p, Wlc_ObjId(p, pObj)) );
        //Wln_ObjSetNameId( pNew, iOutId, Wln_TrasformNameId(pNew, p, pObj) );
    }
    assert(Vec_IntSize(vInits) == Wlc_NtkCoNum(p) - Wlc_NtkPoNum(p), "Vec_IntSize(vInits) == Wlc_NtkCoNum(p) - Wlc_NtkPoNum(p)");
    // inlining of macro Wlc_NtkForEachCo
    for  i = 0; (i < Wlc_NtkCoNum(p)) && (((pObj) = Wlc_NtkCo(p, i)), 1); i++  {
        if i < Wlc_NtkPoNum(p) {
            continue; }
        //char * pInNames[8] = {"d", "clk", "reset", "set", "enable", "async", "sre", "init"};
        Vec_IntClear( vFanins );
        Vec_IntPush( vFanins, Wlc_ObjCopy(p, Wlc_ObjFaninId0(pObj)) );
        for  n = 0; n < 6; n++  {
            Vec_IntPush( vFanins, 0 );
        }
        Vec_IntPush( vFanins, Vec_IntEntry(vInits, i-Wlc_NtkPoNum(p)) );
        Wln_ObjAddFanins( pNew, Vec_IntEntry(&pNew.vFfs, i-Wlc_NtkPoNum(p)), vFanins );
    }
    Vec_IntFree( vFanins );
    Vec_IntFree( vInits );
    return pNew;
}
func Wln_NtkFromWlcTest(p *Wlc_Ntk_t) {
    var pNew *Wln_Ntk_t  = Wln_NtkFromWlc( p );
    Wln_WriteVer( pNew, "test_wlc2wln.v" );
    Wln_NtkFree( pNew );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

