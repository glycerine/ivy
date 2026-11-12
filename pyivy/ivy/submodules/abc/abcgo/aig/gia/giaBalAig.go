package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaBalance.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [AIG balancing.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaBalance.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecHash.h"
//#include "misc/vec/vecQue.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// operation manager
   
type Dam_Man_t struct {
    pGia       *Gia_Man_t;      // user's AIG
    vNod2Set       *Vec_Int_t;  // node ID into fanin set
    vDiv2Nod       *Vec_Int_t;  // div ID into root node set
    vSetStore       *Vec_Int_t; // fanin set storage
    vNodStore       *Vec_Int_t; // root node set storage
    vCounts       *Vec_Flt_t;   // occur counts
    vNodLevR       *Vec_Int_t;  // node reverse level
    vDivLevR       *Vec_Int_t;  // divisor reverse level
    vVisit       *Vec_Int_t;    // visited MUXes
    vQue       *Vec_Que_t;      // pairs by their weight
    vHash   *Hash_IntMan_t;     // pair hash table
    clkStart          abctime;  // starting the clock
    nLevelMax              int32; // maximum level
    nDivs              int32;     // extracted divisor count
    nAnds              int32;     // total AND node count
    nGain              int32;     // total gain in AND nodes
    nGainX              int32;    // gain from XOR nodes
}

func Dam_ObjHand(p *Dam_Man_t, i int32) int32 { if i < Vec_IntSize(p.vNod2Set) {
    return Vec_IntEntry(p.vNod2Set, i);
} else {
    return 0;
}                      }
func Dam_ObjSet(p *Dam_Man_t, i int32) *int32 { var h int32  = Dam_ObjHand(p, i); if h == 0 { return nil; }
 return Vec_IntEntryP(p.vSetStore, h); }

func Dam_DivHand(p *Dam_Man_t, d int32) int32 { if d < Vec_IntSize(p.vDiv2Nod) {
    return Vec_IntEntry(p.vDiv2Nod, d);
} else {
    return 0;
}                      }
func Dam_DivSet(p *Dam_Man_t, d int32) *int32 { var h int32  = Dam_DivHand(p, d); if h == 0 { return nil; }
 return Vec_IntEntryP(p.vNodStore, h); }


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Simplify multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimplifyXor(vSuper *Vec_Int_t) {
    var i int32;
    var k int32;
    var Prev int32 = -1;
    var This int32;
    var fCompl int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((This) = Vec_IntEntry(vSuper, i)), 1); i++  {
        if This == 0 {
            continue; }
        if This == 1 {
            fCompl ^= 1; 
        } else if Prev != This {
            Vec_IntWriteEntry( vSuper, k++, This );
Prev = This;
        } else {
            Prev = -1;
k--; }
    }
    Vec_IntShrink( vSuper, k );
    if Vec_IntSize( vSuper ) == 0 {
        Vec_IntPush( vSuper, fCompl );
    } else if fCompl != 0 {
        Vec_IntWriteEntry( vSuper, 0, Abc_LitNot(Vec_IntEntry(vSuper, 0)) ); }
}
func Gia_ManSimplifyAnd(vSuper *Vec_Int_t) {
    var i int32;
    var k int32;
    var Prev int32 = -1;
    var This int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSuper)) && (((This) = Vec_IntEntry(vSuper, i)), 1); i++  {
        if This == 0 { Vec_IntFill(vSuper, 1, 0); return; }
        if This == 1 {
            continue; }
        if Prev == -1 || Abc_Lit2Var(Prev) != Abc_Lit2Var(This) {
            Vec_IntWriteEntry( vSuper, k++, This );
Prev = This;
        } else if Prev != This { Vec_IntFill(vSuper, 1, 0); return; }
    }
    Vec_IntShrink( vSuper, k );
    if Vec_IntSize( vSuper ) == 0 {
        Vec_IntPush( vSuper, 1 ); }
}

/**Function*************************************************************

  Synopsis    [Collect multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSuperCollectXor_rec(p *Gia_Man_t, pObj *Gia_Obj_t, fStrict int32) {
    assert(!Gia_IsComplement(pObj), "!Gia_IsComplement(pObj)");
    if !Gia_ObjIsXor(pObj) ||     
        ((fStrict != 0) && Gia_ObjRefNum(p, pObj) > 1) || 
        Gia_ObjRefNum(p, pObj) > 2 || 
        (Gia_ObjRefNum(p, pObj) == 2 && (Gia_ObjRefNum(p, Gia_ObjFanin0(pObj)) == 1 || Gia_ObjRefNum(p, Gia_ObjFanin1(pObj)) == 1)) || 
        Vec_IntSize(p.vSuper) > 50 {
        Vec_IntPush( p.vSuper, Gia_ObjToLit(p, pObj) );
        return;
    }
    assert(!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj), "!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj)");
    Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin0(pObj), fStrict );
    Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin1(pObj), fStrict );
}
func Gia_ManSuperCollectAnd_rec(p *Gia_Man_t, pObj *Gia_Obj_t, fStrict int32) {
    if Gia_IsComplement(pObj) || 
        !Gia_ObjIsAndReal(p, pObj) || 
        ((fStrict != 0) && Gia_ObjRefNum(p, pObj) > 1) || 
        Gia_ObjRefNum(p, pObj) > 2 || 
        (Gia_ObjRefNum(p, pObj) == 2 && (Gia_ObjRefNum(p, Gia_ObjFanin0(pObj)) == 1 || Gia_ObjRefNum(p, Gia_ObjFanin1(pObj)) == 1)) || 
        Vec_IntSize(p.vSuper) > 50 {
        Vec_IntPush( p.vSuper, Gia_ObjToLit(p, pObj) );
        return;
    }
    Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild0(pObj), fStrict );
    Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild1(pObj), fStrict );
}
func Gia_ManSuperCollect(p *Gia_Man_t, pObj *Gia_Obj_t, fStrict int32) {
//    int nSize;
    if p.vSuper == nil {
        p.vSuper = Vec_IntAlloc( 1000 );
    } else {
        Vec_IntClear( p.vSuper ); }
    if Gia_ObjIsXor(pObj) {
        assert(!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj), "!Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj)");
        Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin0(pObj), fStrict );
        Gia_ManSuperCollectXor_rec( p, Gia_ObjFanin1(pObj), fStrict );
//        nSize = Vec_IntSize(vSuper);
        Vec_IntSort( p.vSuper, 0 );
        Gia_ManSimplifyXor( p.vSuper );
//        if ( nSize != Vec_IntSize(vSuper) )
//            printf( "X %d->%d  ", nSize, Vec_IntSize(vSuper) );
    
    } else if Gia_ObjIsAndReal(p, pObj) {
        Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild0(pObj), fStrict );
        Gia_ManSuperCollectAnd_rec( p, Gia_ObjChild1(pObj), fStrict );
//        nSize = Vec_IntSize(vSuper);
        Vec_IntSort( p.vSuper, 0 );
        Gia_ManSimplifyAnd( p.vSuper );
//        if ( nSize != Vec_IntSize(vSuper) )
//            printf( "A %d->%d  ", nSize, Vec_IntSize(vSuper) );
    
    } else { assert(false, "0"); }
//    if ( nSize > 10 )
//        printf( "%d ", nSize );
    assert(Vec_IntSize(p.vSuper) > 0, "Vec_IntSize(p.vSuper) > 0");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFindSharedNode(pNew *Gia_Man_t, vSuper *Vec_Int_t, iLit0 int32) int32 {
    var i int32;
    var iLit1 int32 = Vec_IntEntryLast(vSuper);
    // iterate through the nodes whose level is equal to that of the last one
    var iLit1Level int32  = Gia_ObjLevelId(pNew, Abc_Lit2Var(iLit1));
    for  i = Vec_IntSize(vSuper)-1; i >= 0; i--  {
        var iLit2 int32  = Vec_IntEntry(vSuper, i);
        if iLit1Level != Gia_ObjLevelId(pNew, Abc_Lit2Var(iLit2)) {
            break; }
        if Abc_Lit2Var(iLit0) != Abc_Lit2Var(iLit2) && !Gia_ManHashLookupInt(pNew, iLit0, iLit2) { // new node
            continue; }
        // swap iLit2 and iLit1
        if iLit2 != iLit1 {
            Vec_IntWriteEntry( vSuper, i, iLit1 );
            Vec_IntWriteEntry( vSuper, Vec_IntSize(vSuper)-1, iLit2 );
        }
        break;
    }
    return Vec_IntPop(vSuper);    
}
func Gia_ManPrepareLastTwo(pNew *Gia_Man_t, vSuper *Vec_Int_t) {
    var i int32;
    var k int32;
    var Stop int32;
    var Lit1 int32;
    var Lit2 int32;
    var Level1 int32;
    var Level2 int32;
    var pArray *int32;
    var nSize int32  = Vec_IntSize(vSuper);
    if nSize == 2 {
        return; }
    assert(nSize > 2, "nSize > 2");
    Level1 = Gia_ObjLevelId( pNew, Abc_Lit2Var(Vec_IntEntry(vSuper, nSize-2)) );
    // find the first one with Level1
    for  Stop = nSize-3; Stop >= 0; Stop--  {
        Level2 = Gia_ObjLevelId( pNew, Abc_Lit2Var(Vec_IntEntry(vSuper, Stop)) );
        if Level1 != Level2 {
            break; }
    }
    if Stop == nSize-3 {
        return; }
    // avoid worst-case quadratic behavior by looking at the last 8 nodes
    Stop = Abc_MaxInt( Stop, nSize - 9 );
    for  i = nSize - 1; i > Stop; i--  {
        for  k = i - 1; k > Stop; k--  {
            Lit1 = Vec_IntEntry(vSuper, i);
            Lit2 = Vec_IntEntry(vSuper, k);
            if Abc_Lit2Var(Lit1) != Abc_Lit2Var(Lit2) && !Gia_ManHashLookupInt(pNew, Lit1, Lit2) { // new node
                continue; }
            // move Lit1 to be last and Lit2 to be the one before
            pArray = Vec_IntArray( vSuper );
            if i != nSize-1 {
                // inlining of macro ABC_SWAP
                 var t int32  = pArray[i]; pArray[i] = pArray[nSize-1]; pArray[nSize-1] = t; };
            if k != nSize-2 {
                // inlining of macro ABC_SWAP
                 var t int32  = pArray[k]; pArray[k] = pArray[nSize-2]; pArray[nSize-2] = t; };
        }
    }
}
func Gia_ManCreateGate(pNew *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t) {
    var iLit0 int32  = Vec_IntPop(vSuper);
    var iLit1 int32  = Vec_IntPop(vSuper);
//    int iLit1 = Gia_ManFindSharedNode(pNew, vSuper, iLit0);
    var iLit, i int32;
    if !Gia_ObjIsXor(pObj) {
        iLit = Gia_ManHashAnd( pNew, iLit0, iLit1 );
    } else if pNew.pMuxes {
        iLit = Gia_ManHashXorReal( pNew, iLit0, iLit1 );
    } else { 
        iLit = Gia_ManHashXor( pNew, iLit0, iLit1 ); }
    Vec_IntPush( vSuper, iLit );
    Gia_ObjSetGateLevel( pNew, Gia_ManObj(pNew, Abc_Lit2Var(iLit)) );
    // shift to the corrent location
    for  i = Vec_IntSize(vSuper)-1; i > 0; i--  {
        var iLit1 int32  = Vec_IntEntry(vSuper, i);
        var iLit2 int32  = Vec_IntEntry(vSuper, i-1);
        if Gia_ObjLevelId(pNew, Abc_Lit2Var(iLit1)) <= Gia_ObjLevelId(pNew, Abc_Lit2Var(iLit2)) {
            break; }
        Vec_IntWriteEntry( vSuper, i,   iLit2 );
        Vec_IntWriteEntry( vSuper, i-1, iLit1 );
    }
}
func Gia_ManBalanceGate(pNew *Gia_Man_t, pObj *Gia_Obj_t, vSuper *Vec_Int_t, pLits *int32, nLits int32) int32 {
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    Vec_IntClear( vSuper );
    if nLits == 1 {
        Vec_IntPush( vSuper, pLits[0] );
    } else if nLits == 2 {
        Vec_IntPush( vSuper, pLits[0] );
        Vec_IntPush( vSuper, pLits[1] );
        Gia_ManCreateGate( pNew, pObj, vSuper );
    
    } else if nLits > 2 {
        // collect levels
        var i int32;
        var pArray *int32;
        var pPerm *int32; //int iLit;
        for  i = 0; i < nLits; i++  {
            Vec_IntPush( vSuper, Gia_ObjLevelId(pNew, Abc_Lit2Var(pLits[i])) );
        }
        // sort by level
        Vec_IntGrow( vSuper, 4 * nLits );        
        pArray = Vec_IntArray( vSuper );
        pPerm = pArray + nLits;
        Abc_QuickSortCostData( pArray, nLits, 1, (word *)(pArray + 2 * nLits), pPerm );
        // collect in the increasing order of level
        for  i = 0; i < nLits; i++  {
            Vec_IntWriteEntry( vSuper, i, pLits[pPerm[i]] );
        }
        Vec_IntShrink( vSuper, nLits );
/*            
        Vec_IntForEachEntry( vSuper, iLit, i )
            printf( "%d ", Gia_ObjLevel(pNew, Gia_ManObj( pNew, Abc_Lit2Var(iLit) )) );
        printf( "\n" );
*/
        // perform incremental extraction
        for  Vec_IntSize(vSuper) > 1  {
            if !Gia_ObjIsXor(pObj) {
                Gia_ManPrepareLastTwo( pNew, vSuper ); }
            Gia_ManCreateGate( pNew, pObj, vSuper );
        }
    }
    // consider trivial case
    assert(Vec_IntSize(vSuper) == 1, "Vec_IntSize(vSuper) == 1");
    return Vec_IntEntry(vSuper, 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManBalance_rec(pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t, fStrict int32) {
    var i, iLit, iBeg, iEnd int32;
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(!Gia_ObjIsBuf(pObj), "!Gia_ObjIsBuf(pObj)");
    // handle MUX
    if Gia_ObjIsMux(p, pObj) {
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin0(pObj), fStrict );
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin1(pObj), fStrict );
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin2(p, pObj), fStrict );
        pObj.Value = Gia_ManHashMuxReal( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        Gia_ObjSetGateLevel( pNew, Gia_ManObj(pNew, Abc_Lit2Var(pObj.Value)) );
        return;
    }
    // find supergate
    Gia_ManSuperCollect( p, pObj, fStrict );
    // save entries
    if p.vStore == nil {
        p.vStore = Vec_IntAlloc( 1000 ); }
    iBeg = Vec_IntSize( p.vStore );
    Vec_IntAppend( p.vStore, p.vSuper );
    iEnd = Vec_IntSize( p.vStore );
    // call recursively
    // inlining of macro Vec_IntForEachEntryStartStop
    for  i = iBeg; (i < iEnd) && (((iLit) = Vec_IntEntry(p.vStore, i)), 1); i++  {
        var pTemp *Gia_Obj_t  = Gia_ManObj( p, Abc_Lit2Var(iLit) );
        Gia_ManBalance_rec( pNew, p, pTemp, fStrict );
        Vec_IntWriteEntry( p.vStore, i, Abc_LitNotCond(pTemp.Value, Abc_LitIsCompl(iLit)) );
    }
    assert(Vec_IntSize(p.vStore) == iEnd, "Vec_IntSize(p.vStore) == iEnd");
    // consider general case
    pObj.Value = Gia_ManBalanceGate( pNew, pObj, p.vSuper, Vec_IntEntryP(p.vStore, iBeg), iEnd-iBeg );
    Vec_IntShrink( p.vStore, iBeg );
}
func Gia_ManBalanceInt(p *Gia_Man_t, fStrict int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManFillValue( p );
    Gia_ManCreateRefs( p ); 
    // start the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    pNew.vLevels = Vec_IntStart( pNew.nObjsAlloc );
    // create constant and inputs
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // set arrival times for the input of the new AIG
    if p.vCiArrs {
        var Id int32;
        var And2Delay int32;
        if p.And2Delay {
            And2Delay = p.And2Delay;
        } else {
            And2Delay = 1;
        }
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(pNew.vCis)) && ((Id) = Gia_ObjId(pNew, Gia_ManCi(pNew, i))); i++  {
            Vec_IntWriteEntry( pNew.vLevels, Id, Vec_IntEntry(p.vCiArrs, i)/And2Delay );
        }
    
    } else if p.vInArrs {
        var Id int32;
        var And2Delay int32;
        if p.And2Delay {
            And2Delay = p.And2Delay;
        } else {
            And2Delay = 1;
        }
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(pNew.vCis)) && ((Id) = Gia_ObjId(pNew, Gia_ManCi(pNew, i))); i++  {
            Vec_IntWriteEntry( pNew.vLevels, Id, int32((Vec_FltEntry(p.vInArrs, i)/And2Delay)) );
        }
    }
    // create internal nodes
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachBuf
    for  i = Gia_ManBufNum(p) ? 0 : p.nObjs; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsBuf(pObj) {} else {
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin0(pObj), fStrict );
        pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
        Gia_ObjSetGateLevel( pNew, Gia_ManObj(pNew, Abc_Lit2Var(pObj.Value)) );
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManBalance_rec( pNew, p, Gia_ObjFanin0(pObj), fStrict );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    assert(!(fStrict != 0) || Gia_ManObjNum(pNew) <= Gia_ManObjNum(p), "!fStrict || Gia_ManObjNum(pNew) <= Gia_ManObjNum(p)");
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManBalance(p *Gia_Man_t, fSimpleAnd int32, fStrict int32, fVerbose int32) *Gia_Man_t {
    var pNew, pNew1, pNew2 *Gia_Man_t;
    if fVerbose != 0 {      Gia_ManPrintStats( p, nil ); }
    if fSimpleAnd != 0 {
        pNew = Gia_ManDup( p );
    } else {
        pNew = Gia_ManDupMuxes( p, 2 );
    }
    Gia_ManTransferTiming( pNew, p );
    if fVerbose != 0 {      Gia_ManPrintStats( pNew, nil ); }
    pNew1 = Gia_ManBalanceInt( pNew, fStrict );
    Gia_ManTransferTiming( pNew1, pNew );
    if fVerbose != 0 {      Gia_ManPrintStats( pNew1, nil ); }
    Gia_ManStop( pNew );
    pNew2 = Gia_ManDupNoMuxes( pNew1, 0 );
    Gia_ManTransferTiming( pNew2, pNew1 );
    if fVerbose != 0 {      Gia_ManPrintStats( pNew2, nil ); }
    Gia_ManStop( pNew1 );
    return pNew2;
}





/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dam_ManAlloc(pGia *Gia_Man_t) *Dam_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Dam_Man_t{};
    p.clkStart = Abc_Clock();
    p.vVisit = Vec_IntAlloc( 1000 );
    p.pGia = pGia;
    return p;
}
func Dam_ManFree(p *Dam_Man_t) {   
    Vec_IntFreeP( &p.vVisit );
    Vec_IntFreeP( &p.vDivLevR );
    Vec_IntFreeP( &p.vNodLevR );
    Vec_IntFreeP( &p.vNod2Set );
    Vec_IntFreeP( &p.vDiv2Nod );
    Vec_IntFreeP( &p.vSetStore );
    Vec_IntFreeP( &p.vNodStore );
    Vec_FltFreeP( &p.vCounts );
    Vec_QueFreeP( &p.vQue );
    Hash_IntManStop( p.vHash );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Collect initial multi-input gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dam_ManCollectSets_rec(p *Dam_Man_t, Id int32) {
    var pObj *Gia_Obj_t;
    var i, iBeg, iEnd, iLit int32;
    if (Dam_ObjHand(p, Id) != 0) || Id == 0 {
        return; }
    pObj = Gia_ManObj(p.pGia, Id);
    if Gia_ObjIsCi(pObj) {
        return; }
    if Gia_ObjIsBuf(pObj) {
        Dam_ManCollectSets_rec( p, Gia_ObjFaninId0(pObj, Id) );
        return;
    }
    if Gia_ObjIsMux(p.pGia, pObj) {
        if pObj.fMark0 {
            return; }
        pObj.fMark0 = 1;
        Vec_IntPush( p.vVisit, Id );
        Dam_ManCollectSets_rec( p, Gia_ObjFaninId0(pObj, Id) );
        Dam_ManCollectSets_rec( p, Gia_ObjFaninId1(pObj, Id) );
        Dam_ManCollectSets_rec( p, Gia_ObjFaninId2(p.pGia, Id) );
        p.nAnds += 3;
        return;
    }
    Gia_ManSuperCollect( p.pGia, pObj, 0 );
    Vec_IntWriteEntry( p.vNod2Set, Id, Vec_IntSize(p.vSetStore) );
    Vec_IntPush( p.vSetStore, Vec_IntSize(p.pGia.vSuper) );
    p.nAnds += (1 + 2 * Gia_ObjIsXor(pObj)) * (Vec_IntSize(p.pGia.vSuper) - 1);
    // save entries
    iBeg = Vec_IntSize( p.vSetStore );
    Vec_IntAppend( p.vSetStore, p.pGia.vSuper );
    iEnd = Vec_IntSize( p.vSetStore );
    // call recursively
    // inlining of macro Vec_IntForEachEntryStartStop
    for  i = iBeg; (i < iEnd) && (((iLit) = Vec_IntEntry(p.vSetStore, i)), 1); i++  {
        Dam_ManCollectSets_rec( p, Abc_Lit2Var(iLit) );
    }
}
func Dam_ManCollectSets(p *Dam_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    Gia_ManCreateRefs( p.pGia );
    p.vNod2Set  = Vec_IntStart( Gia_ManObjNum(p.pGia) );
    p.vSetStore = Vec_IntAlloc( Gia_ManObjNum(p.pGia) );
    Vec_IntPush( p.vSetStore, -1 );
    Vec_IntClear( p.vVisit );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pGia.vCos)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
        Dam_ManCollectSets_rec( p, Gia_ObjFaninId0p(p.pGia, pObj) );
    }
    // inlining of macro ABC_FREE
    ((p.pGia.pRefs) ? (free((char *) (p.pGia.pRefs)), (p.pGia.pRefs) = 0) : 0);
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(p.vVisit)) && ((pObj) = Gia_ManObj(p.pGia, Vec_IntEntry(p.vVisit,i))); i++  {
        pObj.fMark0 = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Create divisors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dam_ManDivSlack(p *Dam_Man_t, iLit0 int32, iLit1 int32, LevR int32) int32 {
    var Lev0 int32  = Gia_ObjLevel(p.pGia, Gia_ManObj(p.pGia, Abc_Lit2Var(iLit0)));
    var Lev1 int32  = Gia_ObjLevel(p.pGia, Gia_ManObj(p.pGia, Abc_Lit2Var(iLit1)));
    var Slack int32  = p.nLevelMax - LevR - Abc_MaxInt(Lev0, Lev1) - 1 - int32((iLit0 > iLit1));
    return Abc_MinInt( Slack, 100 );
}
func Dam_ManCreateMultiRefs(p *Dam_Man_t, pvRefsAnd **Vec_Int_t, pvRefsXor **Vec_Int_t) {
    var vRefsAnd, vRefsXor *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var k int32;
    var pSet *int32;
    vRefsAnd = Vec_IntStart( 2 * Gia_ManObjNum(p.pGia) );
    vRefsXor = Vec_IntStart( Gia_ManObjNum(p.pGia) );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !(Dam_ObjHand(p, i) != 0) {
            continue; }
        pSet = Dam_ObjSet(p, i);
        if Gia_ObjIsXor(pObj) {
            for  k = 1; k <= pSet[0]; k++  {
                assert(!Abc_LitIsCompl(pSet[k]), "!Abc_LitIsCompl(pSet[k])");
                Vec_IntAddToEntry( vRefsXor, Abc_Lit2Var(pSet[k]), 1 );
            }
        } else if Gia_ObjIsAndReal(p.pGia, pObj) {
            for  k = 1; k <= pSet[0]; k++  {
                Vec_IntAddToEntry( vRefsAnd, pSet[k], 1 );
            }
        } else { assert(false, "0"); }
    }
    }
    *pvRefsAnd = vRefsAnd;
    *pvRefsXor = vRefsXor;
}
func Dam_ManCreatePairs(p *Dam_Man_t, fVerbose int32) {
    var pObj *Gia_Obj_t;
    var vHash *Hash_IntMan_t;
    var vRefsAnd, vRefsXor, vSuper, vDivs, vRemap, vLevRMax *Vec_Int_t;
    var i int32;
    var j int32;
    var k int32;
    var Num int32;
    var FanK int32;
    var FanJ int32;
    var nRefs int32;
    var iNode int32;
    var iDiv int32;
    var pSet *int32;
    var nPairsAll, nPairsTried, nPairsUsed, nPairsXor int32;
    var nDivsAll, nDivsUsed, nDivsXor int32;
    Dam_ManCollectSets( p );
    vSuper = p.pGia.vSuper;
    vDivs  = Vec_IntAlloc( Gia_ManObjNum(p.pGia) );
    vHash  = Hash_IntManStart( Gia_ManObjNum(p.pGia)/2 );
    vLevRMax = Vec_IntStart( 1000 );
    Dam_ManCreateMultiRefs( p, &vRefsAnd, &vRefsXor );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pGia.nObjs) && ((pObj) = Gia_ManObj(p.pGia, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !(Dam_ObjHand(p, i) != 0) {
            continue; }
        pSet = Dam_ObjSet(p, i);
        nPairsAll += pSet[0] * (pSet[0] - 1) / 2;
        Vec_IntClear(vSuper);
        if Gia_ObjIsXor(pObj) {
            for  k = 1; k <= pSet[0]; k++  {
                if Vec_IntEntry(vRefsXor, Abc_Lit2Var(pSet[k])) > 1 {
                    Vec_IntPush( vSuper, pSet[k] ); }
            }
        
        } else if Gia_ObjIsAndReal(p.pGia, pObj) {
            for  k = 1; k <= pSet[0]; k++  {
                if Vec_IntEntry(vRefsAnd, pSet[k]) > 1 {
                    Vec_IntPush( vSuper, pSet[k] ); }
            }
        
        } else { assert(false, "0"); }
        if Vec_IntSize(vSuper) < 2 {
            continue; }
        // enumerate pairs
        nPairsTried += Vec_IntSize(vSuper) * (Vec_IntSize(vSuper) - 1) / 2;
        Vec_IntPush( vDivs, -i ); // remember node
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vSuper)) && (((FanK) = Vec_IntEntry(vSuper, k)), 1); k++  {
        // inlining of macro Vec_IntForEachEntryStart
        for  j = k+1; (j < Vec_IntSize(vSuper)) && (((FanJ) = Vec_IntEntry(vSuper, j)), 1); j++  {
            if (FanK > FanJ) ^ Gia_ObjIsXor(pObj) {
                Num = Hash_Int2ManInsert( vHash, FanJ, FanK, 0 );
            } else {
                Num = Hash_Int2ManInsert( vHash, FanK, FanJ, 0 ); }
            if Hash_Int2ObjInc( vHash, Num ) == 1 {
                nDivsUsed++;
                nDivsXor += Gia_ObjIsXor(pObj);
            }
            Vec_IntPush( vDivs, Num ); // remember devisor
            // update reverse level
            if Num >= Vec_IntSize(vLevRMax) {
                Vec_IntFillExtra( vLevRMax, 3 * Vec_IntSize(vLevRMax) / 2, 0 ); }
            Vec_IntUpdateEntry( vLevRMax, Num, Vec_IntEntry(p.vNodLevR, i) );
        }
        }
    }
    }
    Vec_IntFree( vRefsAnd );
    Vec_IntFree( vRefsXor );
//    Hash_IntManProfile( vHash );
    // remove entries that appear only once
    p.vHash     = Hash_IntManStart( 3 * nDivsUsed /2 );
    p.vCounts   = Vec_FltAlloc( 2 * nDivsUsed );           Vec_FltPush( p.vCounts, ABC_INFINITY );
    p.vQue      = Vec_QueAlloc( Vec_FltCap(p.vCounts) );
    Vec_QueSetPriority( p.vQue, Vec_FltArrayP(p.vCounts) );
    // mapping div to node
    p.vDiv2Nod  = Vec_IntAlloc( 2 * nDivsUsed );           Vec_IntPush( p.vDiv2Nod, ABC_INFINITY );
    p.vNodStore = Vec_IntAlloc( Gia_ManObjNum(p.pGia) );  Vec_IntPush( p.vNodStore, -1 );
    nDivsAll     = Hash_IntManEntryNum(vHash);
    vRemap       = Vec_IntStartFull( nDivsAll+1 );
    for  i = 1; i <= nDivsAll; i++  {
        nRefs = Hash_IntObjData2(vHash, i);
        if nRefs < 2 {
            continue; }
        nPairsUsed += nRefs;
        if Hash_IntObjData0(vHash, i) > Hash_IntObjData1(vHash, i) {
            nPairsXor += nRefs; } 
        Num = Hash_Int2ManInsert( p.vHash, Hash_IntObjData0(vHash, i), Hash_IntObjData1(vHash, i), 0 );
        assert(Num == Hash_IntManEntryNum(p.vHash), "Num == Hash_IntManEntryNum(p.vHash)");
        assert(Num == Vec_FltSize(p.vCounts), "Num == Vec_FltSize(p.vCounts)");
        Vec_FltPush( p.vCounts, nRefs + 0.005*Dam_ManDivSlack(p, Hash_IntObjData0(vHash, i), Hash_IntObjData1(vHash, i), Vec_IntEntry(vLevRMax, i)) );
        Vec_QuePush( p.vQue, Num );
        // remember divisors
        assert(Num == Vec_IntSize(p.vDiv2Nod), "Num == Vec_IntSize(p.vDiv2Nod)");
        Vec_IntPush( p.vDiv2Nod, Vec_IntSize(p.vNodStore) );
        Vec_IntPush( p.vNodStore, 0 );
        Vec_IntFillExtra( p.vNodStore, Vec_IntSize(p.vNodStore) + nRefs, -1 );
        // remember entry
        Vec_IntWriteEntry( vRemap, i, Num );
    }
    assert(Vec_FltSize(p.vCounts) == Hash_IntManEntryNum(p.vHash)+1, "Vec_FltSize(p.vCounts) == Hash_IntManEntryNum(p.vHash)+1");
    assert(Vec_IntSize(p.vDiv2Nod) == nDivsUsed+1, "Vec_IntSize(p.vDiv2Nod) == nDivsUsed+1");
    Hash_IntManStop( vHash );
    Vec_IntFree( vLevRMax );
    // fill in the divisors
    iNode = -1;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDivs)) && (((iDiv) = Vec_IntEntry(vDivs, i)), 1); i++  {
        if iDiv < 0 {
            iNode = -iDiv;
            continue;
        }
        Num = Vec_IntEntry( vRemap, iDiv );
        if Num == -1 {
            continue; }
        pSet = Dam_DivSet( p, Num );
        pSet[++pSet[0]] = iNode;
    }
    Vec_IntFree( vRemap );
    Vec_IntFree( vDivs );
    // create storage for reverse level of divisor during update
    p.vDivLevR = Vec_IntStart( 2 * nDivsUsed );
    // make sure divisors are added correctly
//    for ( i = 1; i <= nDivsUsed; i++ )
//        assert( Dam_DivSet(p, i)[0] == Vec_FltEntry(p->vCounts, i)+1 );
    if !(fVerbose != 0) {
        return; }
    // print stats
    fmt.Printf("Pairs:");
    fmt.Printf("  Total =%9d (%6.2f %%)", nPairsAll, 100.0 * nPairsAll   / Abc_MaxInt(nPairsAll, 1));
    fmt.Printf("  Tried =%9d (%6.2f %%)", nPairsTried, 100.0 * nPairsTried / Abc_MaxInt(nPairsAll, 1));
    fmt.Printf("  Used =%9d (%6.2f %%)", nPairsUsed, 100.0 * nPairsUsed  / Abc_MaxInt(nPairsAll, 1));
    fmt.Printf("  Xor =%9d (%6.2f %%)", nPairsXor, 100.0 * nPairsXor   / Abc_MaxInt(nPairsAll, 1));
    fmt.Printf("\n");
    fmt.Printf("Div:  ");
    fmt.Printf("  Total =%9d (%6.2f %%)", nDivsAll, 100.0 * nDivsAll    / Abc_MaxInt(nDivsAll, 1));
    fmt.Printf("  Tried =%9d (%6.2f %%)", nDivsAll, 100.0 * nDivsAll    / Abc_MaxInt(nDivsAll, 1));
    fmt.Printf("  Used =%9d (%6.2f %%)", nDivsUsed, 100.0 * nDivsUsed   / Abc_MaxInt(nDivsAll, 1));
    fmt.Printf("  Xor =%9d (%6.2f %%)", nDivsXor, 100.0 * nDivsXor    / Abc_MaxInt(nDivsAll, 1));
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Derives new AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dam_ManMultiAig_rec(pMan *Dam_Man_t, pNew *Gia_Man_t, p *Gia_Man_t, pObj *Gia_Obj_t) {
    var i int32;
    var pSet *int32;
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    pSet = Dam_ObjSet(pMan, Gia_ObjId(p, pObj));
    if pSet == nil {
        Dam_ManMultiAig_rec( pMan, pNew, p, Gia_ObjFanin0(pObj) );
        Dam_ManMultiAig_rec( pMan, pNew, p, Gia_ObjFanin1(pObj) );
        if Gia_ObjIsMux(p, pObj) {
            Dam_ManMultiAig_rec( pMan, pNew, p, Gia_ObjFanin2(p, pObj) );
            pObj.Value = Gia_ManHashMuxReal( pNew, Gia_ObjFanin2Copy(p, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        
        } else if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManHashXorReal( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { 
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        Gia_ObjSetGateLevel( pNew, Gia_ManObj(pNew, Abc_Lit2Var(pObj.Value)) );
        return;
    }
    assert(Gia_ObjIsXor(pObj) || Gia_ObjIsAndReal(p, pObj), "Gia_ObjIsXor(pObj) || Gia_ObjIsAndReal(p, pObj)");
    // call recursively
    for  i = 1; i <= pSet[0]; i++  {
        var pTemp *Gia_Obj_t  = Gia_ManObj( p, Abc_Lit2Var(pSet[i]) );
        Dam_ManMultiAig_rec( pMan, pNew, p, pTemp );
        pSet[i] = Abc_LitNotCond( pTemp.Value, Abc_LitIsCompl(pSet[i]) );
    }
    // create balanced gate
    pObj.Value = Gia_ManBalanceGate( pNew, pObj, p.vSuper, pSet + 1, pSet[0] );
}
func Dam_ManMultiAig(pMan *Dam_Man_t) *Gia_Man_t {
    var p *Gia_Man_t  = pMan.pGia;
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // start the new manager
    pNew = Gia_ManStart( 2*Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    pNew.vLevels = Vec_IntStart( pNew.nObjsAlloc );
    // create constant and inputs
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
        Vec_IntWriteEntry( pNew.vLevels, Abc_Lit2Var(pObj.Value), Gia_ObjLevel(p, pObj) );
    }
    // create internal nodes
    Gia_ManHashStart( pNew );
    // inlining of macro Gia_ManForEachBuf
    for  i = Gia_ManBufNum(p) ? 0 : p.nObjs; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsBuf(pObj) {} else {
        Dam_ManMultiAig_rec( pMan, pNew, p, Gia_ObjFanin0(pObj) );
        pObj.Value = Gia_ManAppendBuf( pNew, Gia_ObjFanin0Copy(pObj) );
        Gia_ObjSetGateLevel( pNew, Gia_ManObj(pNew, Abc_Lit2Var(pObj.Value)) );
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Dam_ManMultiAig_rec( pMan, pNew, p, Gia_ObjFanin0(pObj) );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
//    assert( Gia_ManObjNum(pNew) <= Gia_ManObjNum(p) );
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // perform cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Updates the data-structure after extracting one divisor.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dam_PrintDiv(p *Dam_Man_t, iDiv int32) {
    if iDiv == 0 {
        fmt.Printf("Final statistics after extracting %6d divisors:          ", p.nDivs);
    } else {
        var Buffer string // was [100]byte in C

        var iData0 int32  = Hash_IntObjData0(p.vHash, iDiv);
        var iData1 int32  = Hash_IntObjData1(p.vHash, iDiv);
        fmt.Printf("Div%5d : ", p.nDivs+1);
        fmt.Printf("D%-8d = ", iDiv);
        Buffer = fmt.Sprintf("%c%d", Abc_LitIsCompl(iData0)? '!':' ', Abc_Lit2Var(iData0));
        printf( "%8s ",   Buffer );
        fmt.Printf("%c  ", (iData0 < iData1) ? '*' : '+');
        Buffer = fmt.Sprintf("%c%d", Abc_LitIsCompl(iData1)? '!':' ', Abc_Lit2Var(iData1));
        printf( "%8s   ", Buffer );
        fmt.Printf("Weight %9.2f  ", Vec_FltEntry(p.vCounts, iDiv));
    }
    fmt.Printf("Divs =%8d  ", Hash_IntManEntryNum(p.vHash));
    fmt.Printf("Ands =%8d  ", p.nAnds - p.nGain);
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
}
func Dam_PrintQue(p *Dam_Man_t) {
    var i int32;
    fmt.Printf("Divisor queue: \n");
    for  i = 1; i <= Hash_IntManEntryNum(p.vHash); i++  {
        var iLit0 int32  = Hash_IntObjData0(p.vHash, i);
        var iLit1 int32  = Hash_IntObjData1(p.vHash, i);
        fmt.Printf("Div %7d : ", i);
        fmt.Printf("Weight %9.2f  ", Vec_FltEntry(p.vCounts, i));
        fmt.Printf("F = %c%c ", Abc_LitIsCompl(iLit0) ? '!': ' ', 'a' + Abc_Lit2Var(iLit0)-1);
        fmt.Printf("%c ", (Hash_IntObjData0(p.vHash, i) < Hash_IntObjData1(p.vHash, i)) ? '*':'+');
        fmt.Printf("%c%c   ", Abc_LitIsCompl(iLit1) ? '!': ' ', 'a' + Abc_Lit2Var(iLit1)-1);
        fmt.Printf("\n");
    }
}
func Dam_ManUpdateNode(p *Dam_Man_t, iObj int32, iLit0 int32, iLit1 int32, iLitNew int32, vDivs *Vec_Int_t) int32 {
    var pSet *int32  = Dam_ObjSet( p, iObj );
    var i, k, c, Num, iLit, iLit2, fPres int32;
    // check if literal can be found
    for  i = 1; i <= pSet[0]; i++  {
        if pSet[i] == iLit0 {
            break; }
    }
    if i > pSet[0] {
        return 0; }
    // check if literal can be found
    for  i = 1; i <= pSet[0]; i++  {
        if pSet[i] == iLit1 {
            break; }
    }
    if i > pSet[0] {
        return 0; }
    // compact literals
    Vec_IntPush( vDivs, -iObj );
    for  k = i = 1; i <= pSet[0]; i++  {
        if iLit0 == pSet[i] || iLit1 == pSet[i] {
            continue; }
        pSet[k++] = iLit = pSet[i];
        // reduce weights of the divisors
        fPres = 0;
        for  c = 0; c < 2; c++  {
            if c {
                iLit2 = iLit1;
            } else {
                iLit2 = iLit0;
            }
            if (iLit > iLit2) ^ (iLit0 > iLit1) {
                Num = *Hash_Int2ManLookup( p.vHash, iLit2, iLit );
            } else {
                Num = *Hash_Int2ManLookup( p.vHash, iLit, iLit2 ); }
            if Num > 0 {
                Vec_FltAddToEntry( p.vCounts, Num, -1 );
                if Vec_QueIsMember(p.vQue, Num) {
                    Vec_QueUpdate( p.vQue, Num );
                    fPres |= (1 << c);
                }
            }
        }
        if fPres != 3 {
            continue; }
        if (iLit > iLitNew) ^ (iLit0 > iLit1) {
            Num = Hash_Int2ManInsert( p.vHash, iLitNew, iLit, 0 );
        } else {
            Num = Hash_Int2ManInsert( p.vHash, iLit, iLitNew, 0 ); }
        Hash_Int2ObjInc( p.vHash, Num );
        Vec_IntPush( vDivs, Num );
        // update reverse level
        if Num >= Vec_IntSize(p.vDivLevR) {
            Vec_IntFillExtra( p.vDivLevR, 3 * Vec_IntSize(p.vDivLevR) / 2, 0 ); }
        Vec_IntUpdateEntry( p.vDivLevR, Num, Vec_IntEntry(p.vNodLevR, iObj) );
    }
    pSet[k] = iLitNew;
    pSet[0] = k;
    return 1;
}
func Dam_ManUpdate(p *Dam_Man_t, iDiv int32) {
    var vDivs *Vec_Int_t  = p.pGia.vSuper;
    var iLit0 int32  = Hash_IntObjData0(p.vHash, iDiv);
    var iLit1 int32  = Hash_IntObjData1(p.vHash, iDiv);
    var i int32;
    var iLitNew int32;
    var pSet *int32;
    var pNods *int32 = Dam_DivSet( p, iDiv );
    var nPresent, nPairsStart, nPairsStop, pPairsNew, nRefs int32;
    var fThisIsXor int32 = (iLit0 > iLit1);
    var iDivTemp int32;
    var iNode int32;
//    Dam_PrintQue( p );
    if fThisIsXor != 0 {
        iLitNew = Gia_ManAppendXorReal( p.pGia, iLit0, iLit1 );
    } else {
        iLitNew = Gia_ManAppendAnd( p.pGia, iLit0, iLit1 ); }
    Gia_ObjSetGateLevel( p.pGia, Gia_ManObj(p.pGia, Abc_Lit2Var(iLitNew)) );
//    printf( "%d ", Gia_ObjLevel(p->pGia, Gia_ManObj(p->pGia, Abc_Lit2Var(iLitNew))) );
    // replace entries
    assert(pNods[0] >= 2, "pNods[0] >= 2");
    nPairsStart = Hash_IntManEntryNum(p.vHash) + 1;
    Vec_IntClear( vDivs );
    for  i = 1; i <= pNods[0]; i++  {
        nPresent += Dam_ManUpdateNode( p, pNods[i], iLit0, iLit1, iLitNew, vDivs );
    }
    nPairsStop = Hash_IntManEntryNum(p.vHash) + 1;
    // extend arrayvs
    pPairsNew = 0;
    Vec_FltFillExtra( p.vCounts, nPairsStop, 0 );
    Vec_IntFillExtra( p.vDiv2Nod, nPairsStop, -1 );
    for  i = nPairsStart; i < nPairsStop; i++  {
        nRefs = Hash_IntObjData2(p.vHash, i);
        if nRefs < 2 {
            continue; }
        Vec_FltWriteEntry( p.vCounts, i, nRefs + 0.001*Dam_ManDivSlack(p, Hash_IntObjData0(p.vHash, i), Hash_IntObjData1(p.vHash, i), Vec_IntEntry(p.vDivLevR, i)) );
        Vec_QuePush( p.vQue, i );
        // remember divisors
        Vec_IntWriteEntry( p.vDiv2Nod, i, Vec_IntSize(p.vNodStore) );
        Vec_IntPush( p.vNodStore, 0 );
        Vec_IntFillExtra( p.vNodStore, Vec_IntSize(p.vNodStore) + nRefs, -1 );
        pPairsNew++;
    }
//    printf( "Added %d new pairs\n", pPairsNew );
    // fill in the divisors
    iNode = -1;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDivs)) && (((iDivTemp) = Vec_IntEntry(vDivs, i)), 1); i++  {
        if iDivTemp < 0 {
            iNode = -iDivTemp;
            continue;
        }
        if Vec_IntEntry(p.vDiv2Nod, iDivTemp) == -1 {
            continue; }
        pSet = Dam_DivSet( p, iDivTemp );
        pSet[++pSet[0]] = iNode;
    }
    // make sure divisors are added correctly
    for  i = nPairsStart; i < nPairsStop; i++  {
        if Vec_IntEntry(p.vDiv2Nod, i) > 0 {
            assert(Dam_DivSet(p, i)[0] == Hash_IntObjData2(p.vHash, i), "Dam_DivSet(p, i)[0] == Hash_IntObjData2(p.vHash, i)"); }
    }
    // update costs
    Vec_FltWriteEntry( p.vCounts, iDiv, 0 );
    p.nGain += (1 + 2 * fThisIsXor) * (nPresent - 1);
    p.nGainX += 3 * fThisIsXor * (nPresent - 1);
    p.nDivs++;
}

/**Function*************************************************************

  Synopsis    [Perform extraction for multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dam_ManAreaBalanceInt(pGia *Gia_Man_t, vCiLevels *Vec_Int_t, nNewNodesMax int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var p *Dam_Man_t;
    var i, iDiv int32;
    p = Dam_ManAlloc( pGia );
    p.nLevelMax = Gia_ManSetLevels( p.pGia, vCiLevels );
    p.vNodLevR = Gia_ManReverseLevel( p.pGia );
    Vec_IntFillExtra( p.pGia.vLevels, 3*Gia_ManObjNum(p.pGia)/2, 0 );
    Dam_ManCreatePairs( p, fVerbose );
    for  i = 0; i < nNewNodesMax && Vec_QueTopPriority(p.vQue) >= 2; i++  {
        iDiv = Vec_QuePop(p.vQue);
        if fVeryVerbose != 0 {
            Dam_PrintDiv( p, iDiv ); }
        Dam_ManUpdate( p, iDiv );
    }
    if fVeryVerbose != 0 {
        Dam_PrintDiv( p, 0 ); }
    pNew = Dam_ManMultiAig( p );
    if fVerbose != 0 {
        var nDivsAll int32  = Hash_IntManEntryNum(p.vHash);
        var nDivsUsed int32  = p.nDivs;
        fmt.Printf("Div:  ");
        fmt.Printf("  Total =%9d (%6.2f %%) ", nDivsAll, 100.0 * nDivsAll    / Abc_MaxInt(nDivsAll, 1));
        fmt.Printf("  Used =%9d (%6.2f %%)", nDivsUsed, 100.0 * nDivsUsed   / Abc_MaxInt(nDivsAll, 1));
        fmt.Printf("  Gain =%6d (%6.2f %%)", p.nGain, 100.0 * p.nGain / Abc_MaxInt(p.nAnds, 1));
        fmt.Printf("  GainX = %d  ", p.nGainX);
        Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    }
    Dam_ManFree( p );
    return pNew;
}
func Gia_ManAreaBalance(p *Gia_Man_t, fSimpleAnd int32, nNewNodesMax int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pNew0, pNew, pNew1, pNew2 *Gia_Man_t;
    var vCiLevels *Vec_Int_t;
    // set arrival times for the input of the new AIG
    if p.vCiArrs {
        var i int32;
        var Id int32;
        var And2Delay int32;
        if p.And2Delay {
            And2Delay = p.And2Delay;
        } else {
            And2Delay = 1;
        }
        Vec_IntFreeP( &p.vLevels );
        p.vLevels = Vec_IntStart( Gia_ManObjNum(p) );
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
            Vec_IntWriteEntry( p.vLevels, Id, Vec_IntEntry(p.vCiArrs, i)/And2Delay );
        }
    
    } else if p.vInArrs {
        var i int32;
        var Id int32;
        var And2Delay int32;
        if p.And2Delay {
            And2Delay = p.And2Delay;
        } else {
            And2Delay = 1;
        }
        // inlining of macro Gia_ManForEachCiId
        for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
            Vec_IntWriteEntry( p.vLevels, Id, int32((Vec_FltEntry(p.vInArrs, i)/And2Delay)) );
        }
    }
    // determine CI levels
    if p.pManTime && p.vLevels == nil {
        Gia_ManLevelWithBoxes( p ); }
    vCiLevels = Gia_ManGetCiLevels( p );
    // get the starting manager
    if Gia_ManHasMapping(p) {
        pNew0 = (Gia_Man_t *)Dsm_ManDeriveGia(p, 0);
    } else {
        pNew0 = Gia_ManDup(p);
    }
    Gia_ManTransferTiming( pNew0, p );
    if fVerbose != 0 {     Gia_ManPrintStats( pNew0, nil ); }
    // derive internal manager
    if fSimpleAnd != 0 {
        pNew = Gia_ManDup( pNew0 );
    } else {
        pNew = Gia_ManDupMuxes( pNew0, 2 );
    }
    Gia_ManTransferTiming( pNew, pNew0 );
    if fVerbose != 0 {     Gia_ManPrintStats( pNew, nil ); }
    if pNew0 != p { Gia_ManStop( pNew0 ); }
    // perform the operation
    pNew1 = Dam_ManAreaBalanceInt( pNew, vCiLevels, nNewNodesMax, fVerbose, fVeryVerbose );
    Gia_ManTransferTiming( pNew1, pNew );
    if fVerbose != 0 {     Gia_ManPrintStats( pNew1, nil ); }
    Gia_ManStop( pNew );
    Vec_IntFreeP( &vCiLevels );
    // derive the final result
    pNew2 = Gia_ManDupNoMuxes( pNew1, 0 );
    Gia_ManTransferTiming( pNew2, pNew1 );
    if fVerbose != 0 {     Gia_ManPrintStats( pNew2, nil ); }
    Gia_ManStop( pNew1 );
    // normalize if needed
    if !Gia_ManIsNormalized(pNew2) {
        pNew2 = Gia_ManDupNormalize( pNew1 = pNew2, 0 );
        Gia_ManTransferTiming( pNew2, pNew1 );
        Gia_ManStop( pNew1 );
    }
    //Gia_ManTransferTiming( pNew2, p );
    return pNew2;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

