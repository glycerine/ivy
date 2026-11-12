package gia

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

/**CFile****************************************************************

  FileName    [giaMuxes.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Multiplexer profiling algorithm.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMuxes.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilNam.h"
//#include "misc/util/utilTruth.h"
//#include "misc/vec/vecWec.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Counts XORs and MUXes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountMuxXor(p *Gia_Man_t, pnMuxes *int32, pnXors *int32) {
    var pObj, pFan0, pFan1 *Gia_Obj_t; var i int32;
    *pnMuxes = *pnXors = 0;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjIsMuxType(pObj) {
            continue; }
        if Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            (*pnXors)++;
        } else {
            (*pnMuxes)++; }
    }
    }
}
func Gia_ManPrintMuxStats(p *Gia_Man_t) {
    var nAnds, nMuxes, nXors, nTotal int32;
    if p.pMuxes {
        nAnds  = Gia_ManAndNotBufNum(p)-Gia_ManXorNum(p)-Gia_ManMuxNum(p);
        nXors  = Gia_ManXorNum(p);
        nMuxes = Gia_ManMuxNum(p);
        nTotal = nAnds + 3*nXors + 3*nMuxes;
    
    } else {
        Gia_ManCountMuxXor( p, &nMuxes, &nXors );
        nAnds  = Gia_ManAndNotBufNum(p) - 3*nMuxes - 3*nXors;
        nTotal = Gia_ManAndNotBufNum(p);
    }
    Abc_Print( 1, "stats:  " );
    Abc_Print( 1, "xor =%8d %6.2f %%   ", nXors,  300.0*nXors/nTotal );
    Abc_Print( 1, "mux =%8d %6.2f %%   ", nMuxes, 300.0*nMuxes/nTotal );
    Abc_Print( 1, "and =%8d %6.2f %%   ", nAnds,  100.0*nAnds/nTotal );
    Abc_Print( 1, "obj =%8d  ", Gia_ManAndNotBufNum(p) );
    fflush( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Derives GIA with MUXes.]

  Description [Create MUX if the sum of fanin references does not exceed limit.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupMuxes(p *Gia_Man_t, Limit int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pFan0, pFan1, pFanC, pSiblNew, pObjNew *Gia_Obj_t;
    var i int32;
    assert(p.pMuxes == nil, "p.pMuxes == nil");
    assert(Limit >= 0, "Limit >= 0"); // allows to create AIG with XORs without MUXes
    // inlining of macro ABC_FREE
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    Gia_ManCreateRefs( p ); 
    // start the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    if Gia_ManHasChoices(p) {
        // inlining of macro ABC_CALLOC
        pNew.pSibls = ((int *) calloc((size_t)(pNew.nObjsAlloc), sizeof(int))); }
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsBuf(pObj) {
            pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if !Gia_ObjIsMuxType(pObj) || Gia_ObjSibl(p, Gia_ObjFaninId0(pObj, i)) || Gia_ObjSibl(p, Gia_ObjFaninId1(pObj, i)) {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            pObj.Value = Gia_ManHashXorReal( pNew, Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1)) );
        } else if Gia_ObjRefNum(p, Gia_ObjFanin0(pObj)) + Gia_ObjRefNum(p, Gia_ObjFanin1(pObj)) > Limit {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else {
            pFanC = Gia_ObjRecognizeMux( pObj, &pFan1, &pFan0 );
            pObj.Value = Gia_ManHashMuxReal( pNew, Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFanC)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0)) );
        }
        if !Gia_ObjSibl(p, i) {
            continue; }
        pObjNew  = Gia_ManObj( pNew, Abc_Lit2Var(pObj.Value) );
        pSiblNew = Gia_ManObj( pNew, Abc_Lit2Var(Gia_ObjSiblObj(p, i).Value) );
        if Gia_ObjIsAnd(pObjNew) && Gia_ObjIsAnd(pSiblNew) && Gia_ObjId(pNew, pObjNew) > Gia_ObjId(pNew, pSiblNew) {
            pNew.pSibls[Gia_ObjId(pNew, pObjNew)] = Gia_ObjId(pNew, pSiblNew); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Derives GIA without MUXes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupNoMuxes(p *Gia_Man_t, fSkipBufs int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(p.pMuxes != nil || Gia_ManXorNum(p), "p.pMuxes != nil || Gia_ManXorNum(p)");
    // start the new manager
    pNew = Gia_ManStart( 5000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsBuf(pObj) {
            if fSkipBufs != 0 {
                pObj.Value = Gia_ObjFanin0Copy(pObj);
            } else {
                pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
            }
        } else if Gia_ObjIsMuxId(p, i) {
            pObj.Value = Gia_ManHashMux( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManHashXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { 
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Test these procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupMuxesTest(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pNew2 *Gia_Man_t;
    pNew = Gia_ManDupMuxes( p, 2 );
    pNew2 = Gia_ManDupNoMuxes( pNew, 0 );
    Gia_ManPrintStats( p, nil );
    Gia_ManPrintStats( pNew, nil );
    Gia_ManPrintStats( pNew2, nil );
    Gia_ManStop( pNew );
//    Gia_ManStop( pNew2 );
    return pNew2;
}


/**Function*************************************************************

  Synopsis    [Test these procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMuxRestructure(p *Gia_Man_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, nNodes int32;
    var vUsed *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    assert(!Gia_ManHasChoices(p), "!Gia_ManHasChoices(p)");
    assert(!Gia_ManHasMapping(p), "!Gia_ManHasMapping(p)");
    assert(p.pMuxes != nil, "p.pMuxes != nil");
    // inlining of macro ABC_FREE
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    Gia_ManCreateRefs( p ); 
    // start the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    Gia_ManConst0(p).Value = 0;
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pNew );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsBuf(pObj) {
            pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsMuxId(p, i) && 
                  Gia_ObjIsMuxId(p, Gia_ObjFaninId0(pObj, i)) && !Vec_BitEntry(vUsed, Gia_ObjFaninId0(pObj, i)) && 
                  Gia_ObjIsMuxId(p, Gia_ObjFaninId1(pObj, i)) && !Vec_BitEntry(vUsed, Gia_ObjFaninId1(pObj, i)) &&
                  Gia_ObjFaninId2(p, Gia_ObjFaninId0(pObj, i)) == Gia_ObjFaninId2(p, Gia_ObjFaninId1(pObj, i)) {
            var pFan1 *Gia_Obj_t  = Gia_ObjFanin1(pObj);
            var Value0 int32  = Gia_ManHashMux( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin2Copy(p, pFan1), Gia_ObjFanin0Copy(pObj) );
            var Value1 int32  = Gia_ManHashMux( pNew, Value0, Gia_ObjFanin1Copy(pFan1), Gia_ObjFanin0Copy(pFan1) );
            pObj.Value = Gia_ManHashMux( pNew, Gia_ObjFanin2Copy(p, pObj), Value1, Value0 );
            Vec_BitWriteEntry( vUsed, Gia_ObjFaninId0(pObj, i), 1 );
            Vec_BitWriteEntry( vUsed, Gia_ObjFaninId1(pObj, i), 1 );
            Vec_BitWriteEntry( vUsed, i, 1 );
            nNodes++;
        
        } else if Gia_ObjIsMuxId(p, i) {
            pObj.Value = Gia_ManHashMux( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManHashXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { 
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    Vec_BitFree( vUsed );
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Gia_ManDupMuxRestructure(p *Gia_Man_t) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var pNew *Gia_Man_t = Gia_ManDupMuxes( p, 2 );
    pNew = Gia_ManMuxRestructure( pTemp = pNew );  Gia_ManStop( pTemp );
    pNew = Gia_ManDupNoMuxes( pTemp = pNew, 0 );   Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns the size of MUX structure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MuxRef_rec(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t;
    if !Gia_ObjIsMuxId(p, iObj) {
        return 0; }
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjRefInc(p, pObj) {
        return 0; }
    return Gia_MuxRef_rec( p, Gia_ObjFaninId0p(p, pObj) ) + 
           Gia_MuxRef_rec( p, Gia_ObjFaninId1p(p, pObj) ) + 
           Gia_MuxRef_rec( p, Gia_ObjFaninId2p(p, pObj) ) + 1;
}
func Gia_MuxRef(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    assert(Gia_ObjIsMuxId(p, iObj), "Gia_ObjIsMuxId(p, iObj)");
    return Gia_MuxRef_rec( p, Gia_ObjFaninId0p(p, pObj) ) + 
           Gia_MuxRef_rec( p, Gia_ObjFaninId1p(p, pObj) ) + 
           Gia_MuxRef_rec( p, Gia_ObjFaninId2p(p, pObj) ) + 1;
}
func Gia_MuxDeref_rec(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t;
    if !Gia_ObjIsMuxId(p, iObj) {
        return 0; }
    pObj = Gia_ManObj( p, iObj );
    if Gia_ObjRefDec(p, pObj) {
        return 0; }
    return Gia_MuxDeref_rec( p, Gia_ObjFaninId0p(p, pObj) ) + 
           Gia_MuxDeref_rec( p, Gia_ObjFaninId1p(p, pObj) ) + 
           Gia_MuxDeref_rec( p, Gia_ObjFaninId2p(p, pObj) ) + 1;
}
func Gia_MuxDeref(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    assert(Gia_ObjIsMuxId(p, iObj), "Gia_ObjIsMuxId(p, iObj)");
    return Gia_MuxDeref_rec( p, Gia_ObjFaninId0p(p, pObj) ) + 
           Gia_MuxDeref_rec( p, Gia_ObjFaninId1p(p, pObj) ) + 
           Gia_MuxDeref_rec( p, Gia_ObjFaninId2p(p, pObj) ) + 1;
}
func Gia_MuxMffcSize(p *Gia_Man_t, iObj int32) int32 {
    var Count1, Count2 int32;
    if !Gia_ObjIsMuxId(p, iObj) {
        return 0; }
    Count1 = Gia_MuxDeref( p, iObj );
    Count2 = Gia_MuxRef( p, iObj );
    assert(Count1 == Count2, "Count1 == Count2");
    return Count1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MuxStructPrint_rec(p *Gia_Man_t, iObj int32, fFirst int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var iCtrl int32;
    if !(fFirst != 0) && (!Gia_ObjIsMuxId(p, iObj) || Gia_ObjRefNumId(p, iObj) > 0) {
//        printf( "%d", iObj );
        fmt.Printf("<%02d>", Gia_ObjLevelId(p, iObj));
        return;
    }
    iCtrl = Gia_ObjFaninId2p(p, pObj);
    fmt.Printf(" [(");
    if Gia_ObjIsMuxId(p, iCtrl) && Gia_ObjRefNumId(p, iCtrl) == 0 {
        Gia_MuxStructPrint_rec( p, iCtrl, 0 );
    } else {
        fmt.Printf("%d", iCtrl);
        fmt.Printf("<%d>", Gia_ObjLevelId(p, iCtrl));
    }
    fmt.Printf(")");
    if Gia_ObjFaninC2(p, pObj) {
        Gia_MuxStructPrint_rec( p, Gia_ObjFaninId0p(p, pObj), 0 );
        fmt.Printf("|");
        Gia_MuxStructPrint_rec( p, Gia_ObjFaninId1p(p, pObj), 0 );
        fmt.Printf("]");
    
    } else {
        Gia_MuxStructPrint_rec( p, Gia_ObjFaninId1p(p, pObj), 0 );
        fmt.Printf("|");
        Gia_MuxStructPrint_rec( p, Gia_ObjFaninId0p(p, pObj), 0 );
        fmt.Printf("]");
    }
}
func Gia_MuxStructPrint(p *Gia_Man_t, iObj int32) {
    var Count1, Count2 int32;
    assert(Gia_ObjIsMuxId(p, iObj), "Gia_ObjIsMuxId(p, iObj)");
    Count1 = Gia_MuxDeref( p, iObj );
    Gia_MuxStructPrint_rec( p, iObj, 1 );
    Count2 = Gia_MuxRef( p, iObj );
    assert(Count1 == Count2, "Count1 == Count2");
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_MuxStructDump_rec(p *Gia_Man_t, iObj int32, fFirst int32, vStr *Vec_Str_t, nDigitsId int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
    var iCtrl int32;
    if !(fFirst != 0) && (!Gia_ObjIsMuxId(p, iObj) || Gia_ObjRefNumId(p, iObj) > 0) {
        return; }
    iCtrl = Gia_ObjFaninId2p(p, pObj);
    Vec_StrPush( vStr, '[' );
    Vec_StrPush( vStr, '(' );
    if Gia_ObjIsMuxId(p, iCtrl) && Gia_ObjRefNumId(p, iCtrl) == 0 {
        Gia_MuxStructDump_rec( p, iCtrl, 0, vStr, nDigitsId );
    } else {
        Vec_StrPrintNumStar( vStr, iCtrl, nDigitsId ); }
    Vec_StrPush( vStr, ')' );
    if Gia_ObjFaninC2(p, pObj) {
        Gia_MuxStructDump_rec( p, Gia_ObjFaninId0p(p, pObj), 0, vStr, nDigitsId );
        Vec_StrPush( vStr, '|' );
        Gia_MuxStructDump_rec( p, Gia_ObjFaninId1p(p, pObj), 0, vStr, nDigitsId );
        Vec_StrPush( vStr, ']' );
    
    } else {
        Gia_MuxStructDump_rec( p, Gia_ObjFaninId1p(p, pObj), 0, vStr, nDigitsId );
        Vec_StrPush( vStr, '|' );
        Gia_MuxStructDump_rec( p, Gia_ObjFaninId0p(p, pObj), 0, vStr, nDigitsId );
        Vec_StrPush( vStr, ']' );
    }
}
func Gia_MuxStructDump(p *Gia_Man_t, iObj int32, vStr *Vec_Str_t, nDigitsNum int32, nDigitsId int32) int32 {
    var Count1, Count2 int32;
    assert(Gia_ObjIsMuxId(p, iObj), "Gia_ObjIsMuxId(p, iObj)");
    Count1 = Gia_MuxDeref( p, iObj );
    Vec_StrClear( vStr );
    Vec_StrPrintNumStar( vStr, Count1, nDigitsNum );
    Gia_MuxStructDump_rec( p, iObj, 1, vStr, nDigitsId );
    Vec_StrPush( vStr, '\0' );
    Count2 = Gia_MuxRef( p, iObj );
    assert(Count1 == Count2, "Count1 == Count2");
    return Count1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMuxCompare(pp1 *string, pp2 *string) int32 {
    var Diff int32  = int32(strings.Compare(*pp1, *pp2));
    if Diff < 0 {
        return 1; }
    if Diff > 0 { 
        return -1; }
    return 0; 
}
func Gia_ManMuxCountOne(p string) int32 {
    var Count int32;
    for  ; *p; p++  {
        Count += (*p == '[');
    }
    return Count;
}

    
type Mux_Man_t struct {
    pGia      *Gia_Man_t;      // manager
    pNames      *Abc_Nam_t;    // hashing name into ID
    vTops      *Vec_Wec_t;     // top nodes for each struct
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mux_ManAlloc(pGia *Gia_Man_t) *Mux_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Mux_Man_t{};
    p.pGia   = pGia;
    p.pNames = Abc_NamStart( 10000, 50 );
    p.vTops  = Vec_WecAlloc( 1000 );
    Vec_WecPushLevel( p.vTops );
    return p;
}
func Mux_ManFree(p *Mux_Man_t) {
    Abc_NamStop( p.pNames );
    Vec_WecFree( p.vTops );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMuxProfile(p *Mux_Man_t, fWidth int32) int32 {
    var i, Entry, Counter, Total int32;
    var vVec, vCounts *Vec_Int_t;
    vCounts = Vec_IntStart( 1000 );
    if fWidth != 0 {
        // inlining of macro Vec_WecForEachLevelStart
        for  i = 1; (i < Vec_WecSize(p.vTops)) && (((vVec) = Vec_WecEntry(p.vTops, i)), 1); i++  {
            Vec_IntAddToEntry( vCounts, Abc_MinInt(Vec_IntSize(vVec), 999), 1 );
        }
    
    } else {
        for  i = 1; i < Vec_WecSize(p.vTops); i++  {
            Vec_IntAddToEntry( vCounts, Abc_MinInt(strconv.Atoi(Abc_NamStr(p.pNames, i)), 999), 1 );
        }
    }
    Total = Vec_IntCountPositive(vCounts);
    if Total == 0 {
        return 0; }
    fmt.Printf("The distribution of MUX tree %s:\n", fWidth ? "widths" : "sizes");
    Counter = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCounts)) && (((Entry) = Vec_IntEntry(vCounts, i)), 1); i++  {
        if !(Entry != 0) { continue; }
        if Counter++; Counter == 12 {
            fmt.Printf("\n");
Counter = 0; }
        fmt.Printf("  %d=%d", i, Entry);
    }
    fmt.Printf("\nSummary: ");
    fmt.Printf("Max = %d  ", Vec_IntFindMax(vCounts));
    fmt.Printf("Ave = %.2f", 1.0*Vec_IntSum(vCounts)/Total);
    fmt.Printf("\n");
    Vec_IntFree( vCounts );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMuxProfiling(p *Gia_Man_t) {
    var pMan *Mux_Man_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vStr *Vec_Str_t;
    var vFans, vVec *Vec_Int_t;
    var i, Counter, fFound, iStructId, nDigitsId int32;
    var clk abctime  = Abc_Clock();

    pNew = Gia_ManDupMuxes( p, 2 );
    nDigitsId = Abc_Base10Log( Gia_ManObjNum(pNew) );

    pMan = Mux_ManAlloc( pNew );
     
    Gia_ManLevelNum( pNew );
    Gia_ManCreateRefs( pNew );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(pNew.vCos)) && ((pObj) = Gia_ManCo(pNew, i)); i++  {
        Gia_ObjRefFanin0Inc( pNew, pObj );
    }

    vStr = Vec_StrAlloc( 1000 );
    vFans = Gia_ManFirstFanouts( pNew );
    // inlining of macro Gia_ManForEachMux
    for  i = 0; (i < pNew.nObjs) && ((pObj) = Gia_ManObj(pNew, i)); i++  {      if !Gia_ObjIsMuxId(pNew, i) {} else {
        // skip MUXes in the middle of the tree (which have only one MUX fanout)
        if Gia_ObjRefNumId(pNew, i) == 1 && Gia_ObjIsMuxId(pNew, Vec_IntEntry(vFans, i)) {
            continue; }
        // this node is the root of the MUX structure - create hash key
        Counter = Gia_MuxStructDump( pNew, i, vStr, 3, nDigitsId );
        if Counter == 1 {
            continue; }
        iStructId = Abc_NamStrFindOrAdd( pMan.pNames, Vec_StrArray(vStr), &fFound );
        if !fFound {
            Vec_WecPushLevel( pMan.vTops ); }
        assert(Abc_NamObjNumMax(pMan.pNames) == Vec_WecSize(pMan.vTops), "Abc_NamObjNumMax(pMan.pNames) == Vec_WecSize(pMan.vTops)");
        Vec_IntPush( Vec_WecEntry(pMan.vTops, iStructId), i );
    }
    }
    Vec_StrFree( vStr );
    Vec_IntFree( vFans );

    fmt.Printf("MUX structure profile for AIG \"%s\":\n", p.pName);
    fmt.Printf("Total MUXes = %d.  Total trees = %d.  Unique trees = %d.  Memory = %.2f MB   ", Gia_ManMuxNum(pNew), Vec_WecSizeSize(pMan.vTops), Vec_WecSize(pMan.vTops)-1, 1.0*Abc_NamMemUsed(pMan.pNames)/(1<<20));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );

    if Gia_ManMuxProfile(pMan, 0) != 0 {
        Gia_ManMuxProfile( pMan, 1 );

        // short the first ones
        fmt.Printf("The first %d structures: \n", 10);
        // inlining of macro Vec_WecForEachLevelStartStop
        for  i = 1; (i < Abc_MinInt(Vec_WecSize(pMan.vTops), 10)) && (((vVec) = Vec_WecEntry(pMan.vTops, i)), 1); i++  {
            var pTemp string  = Abc_NamStr(pMan.pNames, i);
            fmt.Printf("%5d : ", i);
            fmt.Printf("Occur = %4d   ", Vec_IntSize(vVec));
            fmt.Printf("Size = %4d   ", strconv.Atoi(pTemp));
            fmt.Printf("%s\n", pTemp);
        }

        // print trees for the first one
        Counter = 0;
        // inlining of macro Vec_WecForEachLevelStart
        for  i = 1; (i < Vec_WecSize(pMan.vTops)) && (((vVec) = Vec_WecEntry(pMan.vTops, i)), 1); i++  {
            var pTemp string  = Abc_NamStr(pMan.pNames, i);
            if Vec_IntSize(vVec) > 5 && strconv.Atoi(pTemp) > 5 {
                var k, Entry int32;
                fmt.Printf("For example, structure %d has %d MUXes and bit-width %d:\n", i, strconv.Atoi(pTemp), Vec_IntSize(vVec));
                // inlining of macro Vec_IntForEachEntry
                for  k = 0; (k < Vec_IntSize(vVec)) && (((Entry) = Vec_IntEntry(vVec, k)), 1); k++  {
                    Gia_MuxStructPrint( pNew, Entry );
                }
                if Counter++; Counter == 5 {
                    break; }
            }
        }
    }

    Mux_ManFree( pMan );
    Gia_ManStop( pNew );
}


/**Function*************************************************************

  Synopsis    [Compute one-level TFI/TFO structural signatures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

// these are object/fanin/fanout attributes
// http://stackoverflow.com/questions/9907160/how-to-convert-enum-names-to-string-in-c
// Original macro:
// #define GIA_FOREACH_ITEM(ITEM) \
//     ITEM(C0)    \
//     ITEM(PO)    \
//     ITEM(PI)    \
//     ITEM(FF)    \
//     ITEM(XOR)   \
//     ITEM(MUX)   \
//     ITEM(AND)   \
//     ITEM(iC0)   \
//     ITEM(iC1)   \
//     ITEM(iPI)   \
//     ITEM(iFF)   \
//     ITEM(iXOR)  \
//     ITEM(iMUX)  \
//     ITEM(iAND)  \
//     ITEM(iANDn) \
//     ITEM(iANDp) \
//     ITEM(oPO)   \
//     ITEM(oFF)   \
//     ITEM(oXOR)  \
//     ITEM(oMUXc) \
//     ITEM(oMUXd) \
//     ITEM(oAND)  \
//     ITEM(oANDn) \
//     ITEM(oANDp) \
//     ITEM(GIA_END)


// Original macro:
// #define GENERATE_ENUM(ENUM) ENUM,

 type Gia_ObjType_t int32
// inlining of macro GIA_FOREACH_ITEM
// inlining of macro GENERATE_ENUM
const ( C0 Gia_ObjType_t = iota
        // inlining of macro GENERATE_ENUM
        PO
        // inlining of macro GENERATE_ENUM
        PI
        // inlining of macro GENERATE_ENUM
        FF
        // inlining of macro GENERATE_ENUM
        XOR
       // inlining of macro GENERATE_ENUM
       MUX
       // inlining of macro GENERATE_ENUM
       AND
       // inlining of macro GENERATE_ENUM
       iC0
       // inlining of macro GENERATE_ENUM
       iC1
       // inlining of macro GENERATE_ENUM
       iPI
       // inlining of macro GENERATE_ENUM
       iFF
       // inlining of macro GENERATE_ENUM
       iXOR
      // inlining of macro GENERATE_ENUM
      iMUX
      // inlining of macro GENERATE_ENUM
      iAND
      // inlining of macro GENERATE_ENUM
      iANDn
     // inlining of macro GENERATE_ENUM
     iANDp
     // inlining of macro GENERATE_ENUM
     oPO
       // inlining of macro GENERATE_ENUM
       oFF
       // inlining of macro GENERATE_ENUM
       oXOR
      // inlining of macro GENERATE_ENUM
      oMUXc
     // inlining of macro GENERATE_ENUM
     oMUXd
     // inlining of macro GENERATE_ENUM
     oAND
      // inlining of macro GENERATE_ENUM
      oANDn
     // inlining of macro GENERATE_ENUM
     oANDp
     // inlining of macro GENERATE_ENUM
     GIA_END
 
)

 

// Original macro:
// #define GENERATE_STRING(STRING) #STRING,

// inlining of macro GIA_FOREACH_ITEM
static char * GIA_TYPE_STRINGS[] = { GENERATE_STRING(C0)        GENERATE_STRING(PO)        GENERATE_STRING(PI)        GENERATE_STRING(FF)        GENERATE_STRING(XOR)       GENERATE_STRING(MUX)       GENERATE_STRING(AND)       GENERATE_STRING(iC0)       GENERATE_STRING(iC1)       GENERATE_STRING(iPI)       GENERATE_STRING(iFF)       GENERATE_STRING(iXOR)      GENERATE_STRING(iMUX)      GENERATE_STRING(iAND)      GENERATE_STRING(iANDn)     GENERATE_STRING(iANDp)     GENERATE_STRING(oPO)       GENERATE_STRING(oFF)       GENERATE_STRING(oXOR)      GENERATE_STRING(oMUXc)     GENERATE_STRING(oMUXd)     GENERATE_STRING(oAND)      GENERATE_STRING(oANDn)     GENERATE_STRING(oANDp)     GENERATE_STRING(GIA_END) };

func Gia_ManProfileStructuresTest(p *Gia_Man_t) {
    var i int32;
    for  i = 0; i < GIA_END; i++  {
        fmt.Printf("%d = %s\n", i, GIA_TYPE_STRINGS[i]);
    }
}



// find object code
func Gia_ManEncodeObj(p *Gia_Man_t, i int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, i );
    assert(!Gia_ObjIsRi(p, pObj), "!Gia_ObjIsRi(p, pObj)");
    if Gia_ObjIsConst0(pObj) {
        return C0; }
    if Gia_ObjIsPo(p, pObj) {
        return PO; }
    if Gia_ObjIsPi(p, pObj) {
        return PI; }
    if Gia_ObjIsCi(pObj) {
        return FF; }
    if Gia_ObjIsXor(pObj) {
        return XOR; }
    if Gia_ObjIsMux(p, pObj) {
        return MUX; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return AND;
}
// find fanin code
func Gia_ManEncodeFanin(p *Gia_Man_t, iLit int32) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, Abc_Lit2Var(iLit) );
    if Gia_ObjIsConst0(pObj) {
        return iC0; }
    if Gia_ObjIsPi(p, pObj) {
        return iPI; }
    if Gia_ObjIsCi(pObj) {
        return iFF; }
    if Gia_ObjIsXor(pObj) {
        return iXOR; }
    if Gia_ObjIsMux(p, pObj) {
        return iMUX; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return iAND;
//    if ( Abc_LitIsCompl(iLit) )
//        return iANDn;
//    else
//        return iANDp;
}
// find fanout code
func Gia_ManEncodeFanout(p *Gia_Man_t, pObj *Gia_Obj_t, i int32) Gia_ObjType_t {
//    int iLit;
    if Gia_ObjIsPo(p, pObj) {
        return oPO; }
    if Gia_ObjIsCo(pObj) {
        return oFF; }
    if Gia_ObjIsXor(pObj) {
        return oXOR; }
    if Gia_ObjIsMux(p, pObj) {
        if i == 2 {
            return oMUXc;
        } else {
            return oMUXd;
        } }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return oAND;
//    iLit = i ? Gia_ObjFaninLit1p(p, pObj) : Gia_ObjFaninLit0p(p, pObj);
//    if ( Abc_LitIsCompl(iLit) )
//        return oANDn;
//    else
//        return oANDp;
}

func Gia_ManProfileCollect(p *Gia_Man_t, i int32, vCode *Vec_Int_t, vCodeOffsets *Vec_Int_t, vArray *Vec_Int_t) {
    var k int32;
    Vec_IntClear( vArray );
    for  k = Vec_IntEntry(vCodeOffsets, i); k < Vec_IntEntry(vCodeOffsets, i+1); k++  {
        Vec_IntPush( vArray, Vec_IntEntry(vCode, k) );
    }
}

func Gia_ManProfilePrintOne(p *Gia_Man_t, i int32, vArray *Vec_Int_t) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, i );
    var k, nFanins, nFanouts int32;
    if Gia_ObjIsRi(p, pObj) {
        return; }
    if Gia_ObjIsRo(p, pObj) {
        nFanins = 1;
    } else {
        nFanins = Gia_ObjFaninNum(p, pObj);
    }
    nFanouts = Gia_ObjFanoutNum(p, pObj);

    fmt.Printf("%6d : ", i);
    for  k = 0; k < nFanins; k++  {
        fmt.Printf("  %5s", GIA_TYPE_STRINGS[Vec_IntEntry(vArray, k + 1)]);
    }
    for       ; k < 3; k++  {
        fmt.Printf("  %5s", "");
    }
    fmt.Printf("  ->");
    fmt.Printf(" %5s", GIA_TYPE_STRINGS[Vec_IntEntry(vArray, 0)]);
    fmt.Printf("  ->");
    if nFanouts > 0 {
        var Count int32 = 1;
        var Prev int32 = Vec_IntEntry(vArray, 1 + nFanins);
        for  k = 1; k < nFanouts; k++  {
            if Prev != Vec_IntEntry(vArray, k + 1 + nFanins) {
                fmt.Printf("  %d x %s", Count, GIA_TYPE_STRINGS[Prev]);
                Prev = Vec_IntEntry(vArray, k + 1 + nFanins);
                Count = 0;
            }
            Count++;
        }
        fmt.Printf("  %d x %s", Count, GIA_TYPE_STRINGS[Prev]);
    }
    fmt.Printf("\n");
}

func Gia_ManProfileHash(p *Gia_Man_t, vCode *Vec_Int_t, vCodeOffsets *Vec_Int_t) *Vec_Int_t {
    var pHash *Hsh_VecMan_t;
    var vRes, vArray *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    vRes = Vec_IntAlloc( Gia_ManObjNum(p) );
    pHash = Hsh_VecManStart( Gia_ManObjNum(p) );
    // add empty entry
    vArray = Vec_IntAlloc( 100 );
    Hsh_VecManAdd( pHash, vArray );
    // iterate through the entries
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Gia_ManProfileCollect( p, i, vCode, vCodeOffsets, vArray );
        Vec_IntPush( vRes, Hsh_VecManAdd( pHash, vArray ) );
    }
    Hsh_VecManStop( pHash );
    Vec_IntFree( vArray );
    assert(Vec_IntSize(vRes) == Gia_ManObjNum(p), "Vec_IntSize(vRes) == Gia_ManObjNum(p)");
    return vRes;
}


func Gia_ManProfileStructuresInt(p *Gia_Man_t, nLimit int32, fVerbose int32) {
    var vRes, vCount, vFirst *Vec_Int_t;
    var vCode, vCodeOffsets, vArray *Vec_Int_t;
    var pObj, pFanout *Gia_Obj_t;
    var i int32;
    var k int32;
    var nFanins int32;
    var nFanouts int32;
    var pPerm *int32;
    var nClasses int32;
    assert(p.pMuxes, "p.pMuxes");
    Gia_ManStaticFanoutStart( p );
    // create fanout codes
    vArray = Vec_IntAlloc( 100 );
    vCode = Vec_IntAlloc( 5 * Gia_ManObjNum(p) );
    vCodeOffsets = Vec_IntAlloc( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        Vec_IntPush( vCodeOffsets, Vec_IntSize(vCode) );
        if Gia_ObjIsRi(p, pObj) {
            continue; }
        nFanins = Gia_ObjFaninNum(p, pObj);
        nFanouts = Gia_ObjFanoutNum(p, pObj);
        Vec_IntPush( vCode, Gia_ManEncodeObj(p, i) );
        if nFanins == 3 {
            var iLit int32  = Gia_ObjFaninLit2p(p, pObj);
            Vec_IntPush( vCode, Gia_ManEncodeFanin(p, Abc_LitRegular(iLit)) );
            if Abc_LitIsCompl(iLit) {
                Vec_IntPush( vCode, Gia_ManEncodeFanin(p, Gia_ObjFaninLit0p(p, pObj)) );
                Vec_IntPush( vCode, Gia_ManEncodeFanin(p, Gia_ObjFaninLit1p(p, pObj)) );
            
            } else {
                Vec_IntPush( vCode, Gia_ManEncodeFanin(p, Gia_ObjFaninLit1p(p, pObj)) );
                Vec_IntPush( vCode, Gia_ManEncodeFanin(p, Gia_ObjFaninLit0p(p, pObj)) );
            }
        
        } else if nFanins == 2 {
            var Code0 int32  = Gia_ManEncodeFanin(p, Gia_ObjFaninLit0p(p, pObj));
            var Code1 int32  = Gia_ManEncodeFanin(p, Gia_ObjFaninLit1p(p, pObj));
            Vec_IntPush( vCode, Code0 < Code1 ? Code0 : Code1 );
            Vec_IntPush( vCode, Code0 < Code1 ? Code1 : Code0 );
        
        } else if nFanins == 1 {
            Vec_IntPush( vCode, Gia_ManEncodeFanin(p, Gia_ObjFaninLit0p(p, pObj)) );
        } else if Gia_ObjIsRo(p, pObj) {
            Vec_IntPush( vCode, Gia_ManEncodeFanin(p, Gia_ObjFaninLit0p(p, Gia_ObjRoToRi(p, pObj))) ); }

        // add fanouts
        Vec_IntClear( vArray );
        // inlining of macro Gia_ObjForEachFanoutStatic
        for  k = 0; (k < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, k)), 1); k++  {
            var Index int32  = Gia_ObjWhatFanin( p, pFanout, pObj );
            var Type Gia_ObjType_t  = Gia_ManEncodeFanout( p, pFanout, Index );
            Vec_IntPush( vArray, Type );
        }
        Vec_IntSort( vArray, 0 );
        Vec_IntAppend( vCode, vArray );
    }
    assert(Vec_IntSize(vCodeOffsets) == Gia_ManObjNum(p), "Vec_IntSize(vCodeOffsets) == Gia_ManObjNum(p)");
    Vec_IntPush( vCodeOffsets, Vec_IntSize(vCode) );
    // print the results
    if fVerbose != 0 {
        fmt.Printf("Showing TFI/node/TFO structures for all nodes:\n");
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            Gia_ManProfileCollect( p, i, vCode, vCodeOffsets, vArray );
            Gia_ManProfilePrintOne( p, i, vArray );
        }
    }

    // collect statistics
    vRes = Gia_ManProfileHash( p, vCode, vCodeOffsets );
    //Vec_IntPrint( vRes );

    // count how many times each class appears
    nClasses = Vec_IntFindMax(vRes) + 1;
    vCount = Vec_IntStart( nClasses );
    vFirst = Vec_IntStart( nClasses );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        var Entry int32  = Vec_IntEntry( vRes, i );
        if Gia_ObjIsRi(p, pObj) {
            continue; }
        if Vec_IntEntry(vCount, Entry) == 0 {
            Vec_IntWriteEntry( vFirst, Entry, i ); }
        Vec_IntAddToEntry( vCount, Entry, -1 );
    }
    // sort the counts
    pPerm = Abc_MergeSortCost( Vec_IntArray(vCount), Vec_IntSize(vCount) );
    fmt.Printf("Showing TFI/node/TFO structures that appear more than %d times.\n", nLimit);
    for  i = 0; i < nClasses-1; i++  {
        if nLimit > -Vec_IntEntry(vCount, pPerm[i]) {
            break; }
        fmt.Printf("%6d : ", i);
        fmt.Printf("%6d : ", pPerm[i]);
        fmt.Printf("Weight =%6d  ", -Vec_IntEntry(vCount, pPerm[i]));
        fmt.Printf("First obj =");
        // print the object
        Gia_ManProfileCollect( p, Vec_IntEntry(vFirst, pPerm[i]), vCode, vCodeOffsets, vArray );
        Gia_ManProfilePrintOne( p, Vec_IntEntry(vFirst, pPerm[i]), vArray );
    }

    // cleanup
    // inlining of macro ABC_FREE
    ((pPerm) ? (free((char *) (pPerm)), (pPerm) = 0) : 0);
    Vec_IntFree( vRes );
    Vec_IntFree( vCount );
    Vec_IntFree( vFirst );

    Vec_IntFree( vArray );
    Vec_IntFree( vCode );
    Vec_IntFree( vCodeOffsets );
    Gia_ManStaticFanoutStop( p );
}
func Gia_ManProfileStructures(p *Gia_Man_t, nLimit int32, fVerbose int32) {
    if p.pMuxes {
        Gia_ManProfileStructuresInt( p, nLimit, fVerbose );
    } else {
        var pNew *Gia_Man_t  = Gia_ManDupMuxes( p, 2 );
        Gia_ManProfileStructuresInt( pNew, nLimit, fVerbose );
        Gia_ManStop( pNew );
    }
}

/**Function*************************************************************

  Synopsis    [Circuit restructuring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMarkTfi_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if !Gia_ObjIsAnd(pObj) {
        return; }
    Gia_ManMarkTfi_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ManMarkTfi_rec( p, Gia_ObjFanin1(pObj) );
}
func Gia_ManFindSharedInputs(p *Gia_Man_t) *Vec_Int_t {
    var pObj, pObj2 *Gia_Obj_t; var i, k, Value int32;
    var vRes *Vec_Int_t  = Vec_IntStart( Gia_ManCiNum(p) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManIncrementTravId( p );
        Gia_ManMarkTfi_rec( p, Gia_ObjFanin0(pObj) );
        // inlining of macro Gia_ManForEachCi
        for  k = 0; (k < Vec_IntSize(p.vCis)) && ((pObj2) = Gia_ManCi(p, k)); k++  {
            if Gia_ObjIsTravIdCurrent(p, pObj2) {
                Vec_IntAddToEntry( vRes, k, 1 ); }
        }
    }
    k = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vRes)) && (((Value) = Vec_IntEntry(vRes, i)), 1); i++  {
        if Value == Gia_ManCoNum(p) {
            Vec_IntWriteEntry( vRes, k++, i ); }
    }
    Vec_IntShrink( vRes, k );
    //printf( "Found %d candidate inputs.\n", Vec_IntSize(vRes) );
    if Vec_IntSize(vRes) == 0 || Vec_IntSize(vRes) > 10 {
        Vec_IntFreeP(&vRes); }
    return vRes;
}
func Gia_ManFindCofs(p *Gia_Man_t, vRes *Vec_Int_t, ppNew **Gia_Man_t) *Vec_Wec_t {
    var pObj *Gia_Obj_t;
    var vCofs *Vec_Wec_t  = Vec_WecStart( 1 << Vec_IntSize(vRes) );
    var Value int32;
    var i int32;
    var m int32;
    var nMints int32 = 1 << Vec_IntSize(vRes);
    var pNew *Gia_Man_t  = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    Gia_ManHashAlloc( pNew );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    assert(Vec_IntSize(vRes) < Gia_ManCiNum(p), "Vec_IntSize(vRes) < Gia_ManCiNum(p)");
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    for  m = 0; m < nMints; m++  {
        var vLayer *Vec_Int_t  = Vec_WecEntry( vCofs, m );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vRes)) && (((Value) = Vec_IntEntry(vRes, i)), 1); i++  {
            Gia_ManCi(p, Value).Value = uint32(((m >> i) & 1));
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
            Vec_IntPush( vLayer, Gia_ObjFanin0Copy(pObj) );
        }
        assert(Vec_IntSize(vLayer) == Gia_ManCoNum(p), "Vec_IntSize(vLayer) == Gia_ManCoNum(p)");
        //printf( "%3d : ", m ); Vec_IntPrint( vLayer );
    }
    if ppNew != nil {
        *ppNew = pNew; }
    return vCofs;
}
func Gia_ManFindEquivClasses(vCofs *Vec_Wec_t) *Vec_Int_t {
    var vVec *Vec_Int_t; var i, k, Lev int32;
    var vMap *Vec_Int_t     = Vec_IntAlloc( Vec_WecSize(vCofs) );
    var vUnique *Vec_Int_t  = Vec_IntAlloc( Vec_WecSize(vCofs) );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vCofs)) && (((vVec) = Vec_WecEntry(vCofs, i)), 1); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vUnique)) && (((Lev) = Vec_IntEntry(vUnique, k)), 1); k++  {
            if Vec_IntEqual(vVec, Vec_WecEntry(vCofs, Lev)) {
                break; }
        }
        Vec_IntPush( vMap, k );
        if k == Vec_IntSize(vUnique) {
            Vec_IntPush( vUnique, i ); }
    }
    //printf( "Found %d equiv clasess.\n", Vec_IntSize(vUnique) );
    //Vec_IntPrint( vUnique );
    Vec_IntFree( vUnique );
    assert(Vec_IntSize(vMap) == Vec_WecSize(vCofs), "Vec_IntSize(vMap) == Vec_WecSize(vCofs)");
    //Vec_IntPrint( vMap );
    return vMap;
}
func Gia_ManFindMuxTree_rec(pNew *Gia_Man_t, pCtrl *int32, nCtrl int32, vData *Vec_Int_t, Shift int32) int32 {
    var iLit0, iLit1 int32;
    if nCtrl-- == 0 {
        return Vec_IntEntry( vData, Shift ); }
    iLit0 = Gia_ManFindMuxTree_rec( pNew, pCtrl, nCtrl, vData, Shift );
    iLit1 = Gia_ManFindMuxTree_rec( pNew, pCtrl, nCtrl, vData, Shift + (1<<nCtrl) );
    return Gia_ManHashMux( pNew, pCtrl[nCtrl], iLit1, iLit0 );
}
func Gia_ManFindDerive(pNew *Gia_Man_t, nOuts int32, vIns *Vec_Int_t, vCofs *Vec_Wec_t, vMap *Vec_Int_t) {
    extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );
    var vMemory *Vec_Int_t  = Vec_IntAlloc( 1 << 16 );
    var vLeaves *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vUsed *Vec_Int_t  = Vec_IntStart( Vec_WecSize(vCofs) );
    var vBits *Vec_Int_t  = Vec_IntAlloc( 10 );
    var vData *Vec_Int_t  = Vec_IntAlloc( Vec_WecSize(vCofs) );
    var nWords int32  = Abc_TtWordNum(Vec_IntSize(vIns));
    // inlining of macro ABC_ALLOC
    var pTruth *word  = ((word *) malloc(sizeof(word) * (size_t)(nWords)));
    var nValues int32  = Vec_IntFindMax(vMap)+1;
    var i int32;
    var k int32;
    var Value int32;
    var nBits int32 = Abc_Base2Log(nValues);
    assert(nBits < Vec_IntSize(vIns), "nBits < Vec_IntSize(vIns)");
    assert(Vec_IntSize(vMap) == Vec_WecSize(vCofs), "Vec_IntSize(vMap) == Vec_WecSize(vCofs)");
    assert(Vec_IntSize(vMap) == (1 << Vec_IntSize(vIns)), "Vec_IntSize(vMap) == (1 << Vec_IntSize(vIns))");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vIns)) && (((Value) = Vec_IntEntry(vIns, i)), 1); i++  {
        Vec_IntPush( vLeaves, Gia_ObjToLit(pNew, Gia_ManCi(pNew, Value)) );
    }
    for  i = 0; i < nBits; i++  {
        Abc_TtClear( pTruth, nWords );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vMap)) && (((Value) = Vec_IntEntry(vMap, k)), 1); k++  {
            if (Value >> i) & 1 {
                Abc_TtSetBit( pTruth, k ); }
        }
        if nBits < 6 {
            pTruth[0] = Abc_Tt6Stretch( pTruth[0], Vec_IntSize(vIns) ); }
        Vec_IntPush( vBits, Kit_TruthToGia(pNew, (unsigned*)pTruth, Vec_IntSize(vIns), vMemory, vLeaves, 1) );
        //printf( "Bit %d : ", i ); Dau_DsdPrintFromTruth( pTruth, Vec_IntSize(vIns) );
    }
    for  i = 0; i < nValues; i++  {
        var Cof int32  = Vec_IntFind(vMap, i);
        assert(Cof >= 0 && Cof < Vec_WecSize(vCofs), "Cof >= 0 && Cof < Vec_WecSize(vCofs)");
        Vec_IntWriteEntry( vUsed, Cof, 1 );
    }
    for  i = 0; i < nOuts; i++  {
        var vLevel *Vec_Int_t;
        Vec_IntClear( vData );
        // inlining of macro Vec_WecForEachLevel
        for  k = 0; (k < Vec_WecSize(vCofs)) && (((vLevel) = Vec_WecEntry(vCofs, k)), 1); k++  {
            if Vec_IntEntry(vUsed, k) {
                Vec_IntPush( vData, Vec_IntEntry(vLevel, i) ); }
        }
        for  Vec_IntSize(vData) < (1 << nBits)  {
            Vec_IntPush( vData, 0 );
        }
        assert(Vec_IntSize(vData) == (1 << nBits), "Vec_IntSize(vData) == (1 << nBits)");
        assert(Vec_IntSize(vBits) == nBits, "Vec_IntSize(vBits) == nBits");
        Value = Gia_ManFindMuxTree_rec( pNew, Vec_IntArray(vBits), Vec_IntSize(vBits), vData, 0 );
        Gia_ManAppendCo( pNew, Value );
    }
    // inlining of macro ABC_FREE
    ((pTruth) ? (free((char *) (pTruth)), (pTruth) = 0) : 0);
    Vec_IntFree( vUsed );
    Vec_IntFree( vBits );
    Vec_IntFree( vData );
    Vec_IntFree( vLeaves );
    Vec_IntFree( vMemory );
}
func Gia_ManCofStructure(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t  = nil;
    var vIns *Vec_Int_t  = Gia_ManFindSharedInputs( p );
    var vCfs *Vec_Wec_t;
    if vIns != nil {
        vCfs = Gia_ManFindCofs( p, vIns, &pNew );
    } else {
        vCfs = nil;
    }
    var vMap *Vec_Int_t;
    if vCfs != nil {
        vMap = Gia_ManFindEquivClasses( vCfs );
    } else {
        vMap = nil;
    }
    if (vMap != nil) && Abc_Base2Log(Vec_IntFindMax(vMap)+1) < Vec_IntSize(vIns) {
        var pTemp *Gia_Man_t; 
        Gia_ManFindDerive( pNew, Gia_ManCoNum(p), vIns, vCfs, vMap );
        pNew = Gia_ManCleanup( pTemp = pNew );
        Gia_ManStop( pTemp );
    
    } else {
        Gia_ManStopP( &pNew ); }
    Vec_WecFreeP( &vCfs );
    Vec_IntFreeP( &vMap );
    Vec_IntFreeP( &vIns );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

