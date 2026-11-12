package abs

import "fmt"

/**CFile****************************************************************

  FileName    [absRpm.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Structural reparameterization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absRpm.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "abs.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_IMPL_START 

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Gia_ObjDom(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Vec_IntEntry(p.vDoms, Gia_ObjId(p, pObj));   }
func Gia_ObjSetDom(p *Gia_Man_t, pObj *Gia_Obj_t, d int32) { Vec_IntWriteEntry(p.vDoms, Gia_ObjId(p, pObj), d);  }

// static int Abs_ManSupport2( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp );

// static int Abs_GiaObjDeref_rec( Gia_Man_t * p, Gia_Obj_t * pNode );

// static int Abs_GiaObjRef_rec( Gia_Man_t * p, Gia_Obj_t * pNode );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes one-node dominators.]

  Description [For each node, computes the closest one-node dominator,
  which can be the node itself if the node has no other dominators.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAddDom(p *Gia_Man_t, pObj *Gia_Obj_t, iDom0 int32) {
    var iDom1, iDomNext int32;
    if Gia_ObjDom(p, pObj) == -1 {
        Gia_ObjSetDom( p, pObj, iDom0 );
        return;
    }
    iDom1 = Gia_ObjDom( p, pObj );
    for {
        if iDom0 > iDom1 {
            iDomNext = Gia_ObjDom( p, Gia_ManObj(p, iDom1) );
            if iDomNext == iDom1 {
                break; }
            iDom1 = iDomNext;
            continue;
        }
        if iDom1 > iDom0 {
            iDomNext = Gia_ObjDom( p, Gia_ManObj(p, iDom0) );
            if iDomNext == iDom0 {
                break; }
            iDom0 = iDomNext;
            continue;
        }
        assert(iDom0 == iDom1, "iDom0 == iDom1");
        Gia_ObjSetDom( p, pObj, iDom0 );
        return;
    }
    Gia_ObjSetDom( p, pObj, Gia_ObjId(p, pObj) );
}
func Gia_ManComputeDoms(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    if p.vDoms == nil {
        p.vDoms = Vec_IntAlloc( 0 ); }
    Vec_IntFill( p.vDoms, Gia_ManObjNum(p), -1 );
    // inlining of macro Gia_ManForEachObjReverse
    for  i = p.nObjs - 1; (i >= 0) && ((pObj) = Gia_ManObj(p, i)); i--  {
        if i == 0 || Gia_ObjIsCi(pObj) {
            continue; }
        if pObj.fMark1 || (p.pRefs && Gia_ObjIsAnd(pObj) && Gia_ObjRefNum(p, pObj) == 0) {
            continue; }
        if Gia_ObjIsCo(pObj) {
            Gia_ObjSetDom( p, pObj, i );
            Gia_ManAddDom( p, Gia_ObjFanin0(pObj), i );
            continue;
        }
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        Gia_ManAddDom( p, Gia_ObjFanin0(pObj), i );
        Gia_ManAddDom( p, Gia_ObjFanin1(pObj), i );
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCreateSupps(p *Gia_Man_t, fVerbose int32) *Vec_Wec_t {
    var clk abctime  = Abc_Clock();
    var pObj *Gia_Obj_t; var i, Id int32;
    var vSupps *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Vec_IntPush( Vec_WecEntry(vSupps, Id), i );
    }
    // inlining of macro Gia_ManForEachAnd
    for  Id = 0; (Id < p.nObjs) && ((pObj) = Gia_ManObj(p, Id)); Id++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Vec_IntTwoMerge2( Vec_WecEntry(vSupps, Gia_ObjFaninId0(pObj, Id)), 
                          Vec_WecEntry(vSupps, Gia_ObjFaninId1(pObj, Id)), 
                          Vec_WecEntry(vSupps, Id) ); }
    } 
//    Gia_ManForEachCo( p, pObj, i )
//        Vec_IntAppend( Vec_WecEntry(vSupps, Gia_ObjId(p, pObj)), Vec_WecEntry(vSupps, Gia_ObjFaninId0p(p, pObj)) );
    if fVerbose != 0 {
        Abc_PrintTime( 1, "Support computation", Abc_Clock() - clk ); }
    return vSupps;
}
func Gia_ManDomTest(p *Gia_Man_t) {
    var vDoms *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vSupp *Vec_Int_t  = Vec_IntAlloc( 100 );
    var vSupps *Vec_Wec_t  = Gia_ManCreateSupps( p, 1 );
    var vDomeds *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(p) );
    var pObj, pDom *Gia_Obj_t; var i, Id, nMffcSize int32;
    Gia_ManCreateRefs( p );
    Gia_ManComputeDoms( p );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if Gia_ObjDom(p, pObj) == -1 {
            continue; }
        for  pDom = Gia_ManObj(p, Gia_ObjDom(p, pObj)); Gia_ObjIsAnd(pDom); pDom = Gia_ManObj(p, Gia_ObjDom(p, pDom))  {
            Vec_IntPush( Vec_WecEntry(vDomeds, Gia_ObjId(p, pDom)), i );
        }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Vec_IntEqual(Vec_WecEntry(vSupps, i), Vec_WecEntry(vDomeds, i)) {
            Vec_IntPush( vDoms, i ); }
    }
    Vec_WecFree( vSupps );
    Vec_WecFree( vDomeds );

    // check MFFC sizes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDoms)) && (((Id) = Vec_IntEntry(vDoms, i)), 1); i++  {
        Gia_ObjRefInc( p, Gia_ManObj(p, Id) );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDoms)) && (((Id) = Vec_IntEntry(vDoms, i)), 1); i++  {
        nMffcSize = Gia_NodeMffcSizeSupp( p, Gia_ManObj(p, Id), vSupp );
        fmt.Printf("%d(%d:%d) ", Id, Vec_IntSize(vSupp), nMffcSize);
    }
    fmt.Printf("\n");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vDoms)) && (((Id) = Vec_IntEntry(vDoms, i)), 1); i++  {
        Gia_ObjRefDec( p, Gia_ManObj(p, Id) );
    }

//    Vec_IntPrint( vDoms );
    Vec_IntFree( vDoms );
    Vec_IntFree( vSupp );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManTestDoms2(p *Gia_Man_t) {
    var vNodes *Vec_Int_t;
    var pObj, pDom *Gia_Obj_t;
    var clk abctime  = Abc_Clock();
    var i int32;
    assert(p.vDoms == nil, "p.vDoms == nil");
    Gia_ManComputeDoms( p );
/*
    Gia_ManForEachPi( p, pObj, i )
        if ( Gia_ObjId(p, pObj) != Gia_ObjDom(p, pObj) )
            printf( "PI =%6d  Id =%8d. Dom =%8d.\n", i, Gia_ObjId(p, pObj), Gia_ObjDom(p, pObj) );
*/
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    // for each dominated PI, when if the PIs is in a leaf of the MFFC of the dominator
    Gia_ManCleanMark1( p );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark1 = 1;
    }
    vNodes = Vec_IntAlloc( 100 );
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if Gia_ObjId(p, pObj) == Gia_ObjDom(p, pObj) {
            continue; }

        pDom = Gia_ManObj(p, Gia_ObjDom(p, pObj));
        if Gia_ObjIsCo(pDom) {
            assert(Gia_ObjFanin0(pDom) == pObj, "Gia_ObjFanin0(pDom) == pObj");
            continue;
        }
        assert(Gia_ObjIsAnd(pDom), "Gia_ObjIsAnd(pDom)");
        Abs_GiaObjDeref_rec( p, pDom );
        Abs_ManSupport2( p, pDom, vNodes );
        Abs_GiaObjRef_rec( p, pDom );

        if Vec_IntFind(vNodes, Gia_ObjId(p, pObj)) == -1 {
            fmt.Printf("FAILURE.\n"); }
//        else
//            printf( "Success.\n" );
    }
    Vec_IntFree( vNodes );
    Gia_ManCleanMark1( p );
}

/**Function*************************************************************

  Synopsis    [Collect PI doms.]

  Description [Assumes that some PIs and ANDs are marked with fMark1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectDoms(p *Gia_Man_t) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var Lev int32;
    var LevMax int32 = ABC_INFINITY;
    var i, iDom, iDomNext int32;
    vNodes = Vec_IntAlloc( 100 );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if !pObj.fMark1 {
            continue; }
        if p.pRefs && Gia_ObjRefNum(p, pObj) == 0 {
            continue; }
        iDom = Gia_ObjDom(p, pObj);
        if iDom == -1 {
            continue; }
        if iDom == i {
            continue; }
        for  Lev = 0; Lev < LevMax && Gia_ObjIsAnd( Gia_ManObj(p, iDom) ); Lev++  {
            Vec_IntPush( vNodes, iDom );
            iDomNext = Gia_ObjDom( p, Gia_ManObj(p, iDom) );
            if iDomNext == iDom {
                break; }
            iDom = iDomNext;
        }
    }
    Vec_IntUniqify( vNodes );
    return vNodes;
}
func Gia_ManComputePiDoms(p *Gia_Man_t) *Vec_Int_t {
    var vNodes *Vec_Int_t;
    Gia_ManComputeDoms( p );
    vNodes = Gia_ManCollectDoms( p );
//    Vec_IntPrint( vNodes );
    return vNodes;
}
func Gia_ManTestDoms(p *Gia_Man_t) {
    var vNodes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // mark PIs
//    Gia_ManCreateRefs( p );
    Gia_ManCleanMark1( p );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark1 = 1;
    }
    // compute dominators
    assert(p.vDoms == nil, "p.vDoms == nil");
    vNodes = Gia_ManComputePiDoms( p );
//    printf( "Nodes = %d. Doms = %d.\n", Gia_ManAndNum(p), Vec_IntSize(vNodes) );
    Vec_IntFree( vNodes );
    // unmark PIs
    Gia_ManCleanMark1( p );
}


/**Function*************************************************************

  Synopsis    [Counts flops without fanout.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountFanoutlessFlops(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    var Counter int32;
    Gia_ManCreateRefs( p );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Gia_ObjRefNum(p, pObj) == 0 {
            Counter++; }
    }
    fmt.Printf("Fanoutless flops = %d.\n", Counter);
    p.pRefs = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountPisNodes_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vPis *Vec_Int_t, vAnds *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if pObj.fMark1 {
        Vec_IntPush( vPis, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManCountPisNodes_rec( p, Gia_ObjFanin0(pObj), vPis, vAnds );
    Gia_ManCountPisNodes_rec( p, Gia_ObjFanin1(pObj), vPis, vAnds );
    Vec_IntPush( vAnds, Gia_ObjId(p, pObj) );
}
func Gia_ManCountPisNodes(p *Gia_Man_t, vPis *Vec_Int_t, vAnds *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // mark const0 and flop output
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ObjSetTravIdCurrent( p, pObj );
    }
    // count PIs and internal nodes reachable from COs
    Vec_IntClear( vPis );
    Vec_IntClear( vAnds );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManCountPisNodes_rec( p, Gia_ObjFanin0(pObj), vPis, vAnds );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_GiaObjDeref_rec(p *Gia_Man_t, pNode *Gia_Obj_t) int32 {
    var pFanin *Gia_Obj_t;
    var Counter int32;
    if pNode.fMark1 || Gia_ObjIsRo(p, pNode) {
        return 0; }
    assert(Gia_ObjIsAnd(pNode), "Gia_ObjIsAnd(pNode)");
    pFanin = Gia_ObjFanin0(pNode);
    assert(Gia_ObjRefNum(p, pFanin) > 0, "Gia_ObjRefNum(p, pFanin) > 0");
    if Gia_ObjRefDec(p, pFanin) == 0 {
        Counter += Abs_GiaObjDeref_rec( p, pFanin ); }
    pFanin = Gia_ObjFanin1(pNode);
    assert(Gia_ObjRefNum(p, pFanin) > 0, "Gia_ObjRefNum(p, pFanin) > 0");
    if Gia_ObjRefDec(p, pFanin) == 0 {
        Counter += Abs_GiaObjDeref_rec( p, pFanin ); }
    return Counter + 1;
}
func Abs_GiaObjRef_rec(p *Gia_Man_t, pNode *Gia_Obj_t) int32 {
    var pFanin *Gia_Obj_t;
    var Counter int32;
    if pNode.fMark1 || Gia_ObjIsRo(p, pNode) {
        return 0; }
    assert(Gia_ObjIsAnd(pNode), "Gia_ObjIsAnd(pNode)");
    pFanin = Gia_ObjFanin0(pNode);
    if Gia_ObjRefInc(p, pFanin) == 0 {
        Counter += Abs_GiaObjRef_rec( p, pFanin ); }
    pFanin = Gia_ObjFanin1(pNode);
    if Gia_ObjRefInc(p, pFanin) == 0 {
        Counter += Abs_GiaObjRef_rec( p, pFanin ); }
    return Counter + 1;
}

/**Function*************************************************************

  Synopsis    [Returns the number of nodes with zero refs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_GiaSortNodes(p *Gia_Man_t, vSupp *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var nSize int32  = Vec_IntSize(vSupp);
    var i, RetValue int32;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vSupp)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vSupp,i))); i++  {
        if i < nSize && Gia_ObjRefNum(p, pObj) == 0 && !Gia_ObjIsRo(p, pObj) { // add removable leaves
        
            assert(pObj.fMark1, "pObj.fMark1");
            Vec_IntPush( vSupp, Gia_ObjId(p, pObj) );
        }
    }
    RetValue = Vec_IntSize(vSupp) - nSize;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vSupp)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vSupp,i))); i++  {
        if i < nSize && !(Gia_ObjRefNum(p, pObj) == 0 && !Gia_ObjIsRo(p, pObj)) { // add non-removable leaves
            Vec_IntPush( vSupp, Gia_ObjId(p, pObj) ); }
    }
    assert(Vec_IntSize(vSupp) == 2 * nSize, "Vec_IntSize(vSupp) == 2 * nSize");
    memmove( Vec_IntArray(vSupp), Vec_IntArray(vSupp) + nSize, sizeof(int) * nSize );
    Vec_IntShrink( vSupp, nSize );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Computes support in terms of PIs and flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_ManSupport1_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if pObj.fMark1 || Gia_ObjIsRo(p, pObj) {
        Vec_IntPush( vSupp, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Abs_ManSupport1_rec( p, Gia_ObjFanin0(pObj), vSupp );
    Abs_ManSupport1_rec( p, Gia_ObjFanin1(pObj), vSupp );
}
func Abs_ManSupport1(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t) int32 {
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Vec_IntClear( vSupp );
    Gia_ManIncrementTravId( p );
    Abs_ManSupport1_rec( p, pObj, vSupp );
    return Vec_IntSize(vSupp);
}

/**Function*************************************************************

  Synopsis    [Computes support of the MFFC.]

  Description [Should be called when pObj's cone is dereferenced.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_ManSupport2_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if pObj.fMark1 || Gia_ObjIsRo(p, pObj) || Gia_ObjRefNum(p, pObj) > 0 {
        Vec_IntPush( vSupp, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Abs_ManSupport2_rec( p, Gia_ObjFanin0(pObj), vSupp );
    Abs_ManSupport2_rec( p, Gia_ObjFanin1(pObj), vSupp );
}
func Abs_ManSupport2(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t) int32 {
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Vec_IntClear( vSupp );
    Gia_ManIncrementTravId( p );
    Abs_ManSupport2_rec( p, Gia_ObjFanin0(pObj), vSupp );
    Abs_ManSupport2_rec( p, Gia_ObjFanin1(pObj), vSupp );
    Gia_ObjSetTravIdCurrent(p, pObj);
    return Vec_IntSize(vSupp);
}

/**Function*************************************************************

  Synopsis    [Computes support of the extended MFFC.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_ManSupport3(p *Gia_Man_t, pObj *Gia_Obj_t, vSupp *Vec_Int_t) int32 {
    var pTemp, pFan0, pFan1 *Gia_Obj_t;
    var i, nSize0 int32;
    // collect MFFC
    Abs_ManSupport2( p, pObj, vSupp );
    // move dominated to the front
    nSize0 = Abs_GiaSortNodes( p, vSupp );
    assert(nSize0 > 0, "nSize0 > 0");
    // consider remaining nodes
    for {
        var fChanges int32;
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vSupp)) && ((pTemp) = Gia_ManObj(p, Vec_IntEntry(vSupp,i))); i++  {
            if i < nSize0 {
                continue; }
            if !Gia_ObjIsAnd(pTemp) {
                continue; }
            assert(!pTemp.fMark1, "!pTemp.fMark1");
            assert(Gia_ObjRefNum(p, pTemp) > 0, "Gia_ObjRefNum(p, pTemp) > 0");
            pFan0 = Gia_ObjFanin0(pTemp);
            pFan1 = Gia_ObjFanin1(pTemp);
            if Gia_ObjIsTravIdCurrent(p, pFan0) && Gia_ObjIsTravIdCurrent(p, pFan1) {
                Vec_IntRemove( vSupp, Gia_ObjId(p, pTemp) );
                fChanges = 1;
                break;
            }
            if Gia_ObjIsTravIdCurrent(p, pFan0) {
                Vec_IntRemove( vSupp, Gia_ObjId(p, pTemp) );
                Vec_IntPush( vSupp, Gia_ObjId(p, pFan1) );
                assert(!Gia_ObjIsTravIdCurrent(p, pFan1), "!Gia_ObjIsTravIdCurrent(p, pFan1)");
                Gia_ObjSetTravIdCurrent(p, pFan1);
                fChanges = 1;
                break;
            }
            if Gia_ObjIsTravIdCurrent(p, pFan1) {
                Vec_IntRemove( vSupp, Gia_ObjId(p, pTemp) );
                Vec_IntPush( vSupp, Gia_ObjId(p, pFan0) );
                assert(!Gia_ObjIsTravIdCurrent(p, pFan0), "!Gia_ObjIsTravIdCurrent(p, pFan0)");
                Gia_ObjSetTravIdCurrent(p, pFan0);
                fChanges = 1;
                break;
            }
        }
        if !(fChanges != 0) {
            break; }
    }
    return Vec_IntSize(vSupp);
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_GiaCofPrint(pTruth *word, nSize int32, nSize0 int32, Res int32) int32 {
    var i, Bit int32;
    var nBits int32  = (1 << nSize);
    var nStep int32  = (1 << nSize0);
    var Mark [2]int32 = [2]int32{1,1};
    for  i = 0; i < nBits; i++  {
        if i % nStep == 0 {
            fmt.Printf(" ");
            assert((Res != 0) || (Mark[0] && Mark[1]), "Res || (Mark[0] && Mark[1])");
            Mark[0] = Mark[1] = 0;
        }
        Bit = Abc_InfoHasBit((unsigned *)pTruth, i);
        Mark[Bit] = 1;
        fmt.Printf("%d", Bit);
    }
    fmt.Printf("\n");
    assert((Res != 0) || (Mark[0] && Mark[1]), "Res || (Mark[0] && Mark[1])");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if truth table has no const cofactors.]

  Description [The cofactoring variables are the (nSize-nSize0)
  most significant vars.  Each cofactor depends on nSize0 vars.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_GiaCheckTruth(pTruth *word, nSize int32, nSize0 int32) int32 {
    unsigned char * pStr = (unsigned char *)pTruth;
    var nStr int32;
    if nSize >= 3 {
        nStr = (1 << (nSize - 3));
    } else {
        nStr = 1;
    }
    var i, k, nSteps int32;
    assert(nSize0 > 0 && nSize0 <= nSize, "nSize0 > 0 && nSize0 <= nSize");
    if nSize0 == 1 {
        for  i = 0; i < nStr; i++  {
            if ((uint32(pStr[i]) ^ (uint32(pStr[i]) >> 1)) & 0x55) != 0x55 {
                return 0; }
        }
        return 1;
    }
    if nSize0 == 2 {
        for  i = 0; i < nStr; i++  {
            if (uint32(pStr[i]) & 0xF) == 0x0 || ((uint32(pStr[i]) >> 4) & 0xF) == 0x0 || 
                 (uint32(pStr[i]) & 0xF) == 0xF || ((uint32(pStr[i]) >> 4) & 0xF) == 0xF {
                return 0; }
        }
        return 1;
    }
    assert(nSize0 >= 3, "nSize0 >= 3");
    nSteps = (1 << (nSize0 - 3));
    for  i = 0; i < nStr; i += nSteps  {
        for  k = 0; k < nSteps; k++  {
            if (uint32(pStr[i+k]) & 0xFF) != 0x00 {
                break; }
        }
        if k == nSteps {
            break; }
        for  k = 0; k < nSteps; k++  {
            if (uint32(pStr[i+k]) & 0xFF) != 0xFF {
                break; }
        }
        if k == nSteps {
            break; }
    }
    assert(i <= nStr, "i <= nStr");
    return int32(( i == nStr ));
}

/**Function*************************************************************

  Synopsis    [Returns 1 if truth table has const cofactors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_RpmPerformMark(p *Gia_Man_t, nCutMax int32, fVerbose int32, fVeryVerbose int32) {
    var vPis, vAnds, vDoms *Vec_Int_t;
    var vSupp, vSupp1, vSupp2 *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var pTruth *word;
    var Iter, i, nSize0, nNodes int32;
    var fHasConst int32;
    var fChanges int32 = 1;
    Gia_ManCreateRefs( p );
    Gia_ManCleanMark1( p );
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.fMark1 = 1;
    }
    vPis = Vec_IntAlloc( 100 );
    vAnds = Vec_IntAlloc( 100 );
    vSupp1 = Vec_IntAlloc( 100 );
    vSupp2 = Vec_IntAlloc( 100 );
    for  Iter = 0; (fChanges != 0); Iter++  {
        fChanges = 0;
        vDoms = Gia_ManComputePiDoms( p );
        // count the number of PIs and internal nodes
        if (fVerbose != 0) || (fVeryVerbose != 0) {
            Gia_ManCountPisNodes( p, vPis, vAnds );
            fmt.Printf("Iter %3d :  ", Iter);
            fmt.Printf("PI = %5d  (%6.2f %%)  ", Vec_IntSize(vPis), 100.0 * Vec_IntSize(vPis)  / Gia_ManPiNum(p));
            fmt.Printf("And = %6d  (%6.2f %%) ", Vec_IntSize(vAnds), 100.0 * Vec_IntSize(vAnds) / Gia_ManAndNum(p));
            fmt.Printf("Dom = %5d  (%6.2f %%)  ", Vec_IntSize(vDoms), 100.0 * Vec_IntSize(vDoms) / Gia_ManAndNum(p));
            fmt.Printf("\n");
        }
//        pObj = Gia_ObjFanin0( Gia_ManPo(p, 1) );
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vDoms)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vDoms,i))); i++  {
            assert(!pObj.fMark1, "!pObj.fMark1");
            assert(Gia_ObjRefNum( p, pObj ) > 0, "Gia_ObjRefNum( p, pObj ) > 0");
            // dereference root node
            nNodes = Abs_GiaObjDeref_rec( p, pObj );
/*
            // compute support of full cone
            if ( Abs_ManSupport1(p, pObj, vSupp1) > nCutMax )
//            if ( 1 )
            {
                // check support of MFFC
                if ( Abs_ManSupport2(p, pObj, vSupp2) > nCutMax )
//                if ( 1 )
                {
                    Abs_GiaObjRef_rec( p, pObj );
                    continue;
                }
                vSupp = vSupp2;
//                printf( "-" );
            }
            else
            {
                vSupp = vSupp1;
//                printf( "+" );
            }
*/
            if Abs_ManSupport2(p, pObj, vSupp2) > nCutMax {
                Abs_GiaObjRef_rec( p, pObj );
                continue;
            }
            vSupp = vSupp2;

            // order nodes by their ref counts
            nSize0 = Abs_GiaSortNodes( p, vSupp );
            assert(nSize0 > 0 && nSize0 <= nCutMax, "nSize0 > 0 && nSize0 <= nCutMax");
            // check if truth table has const cofs
            pTruth = Gia_ObjComputeTruthTableCut( p, pObj, vSupp );
            if pTruth == nil {
                Abs_GiaObjRef_rec( p, pObj );
                continue;
            }
            fHasConst = !Abs_GiaCheckTruth( pTruth, Vec_IntSize(vSupp), nSize0 );
            if fVeryVerbose != 0 {
                fmt.Printf("Nodes =%3d ", nNodes);
                fmt.Printf("Size =%3d ", Vec_IntSize(vSupp));
                fmt.Printf("Size0 =%3d  ", nSize0);
                fmt.Printf("%3s", fHasConst ? "yes" : "no");
                Abs_GiaCofPrint( pTruth, Vec_IntSize(vSupp), nSize0, fHasConst );
            }
            if fHasConst != 0 {
                Abs_GiaObjRef_rec( p, pObj );
                continue;
            }
            // pObj can be reparamed
            pObj.fMark1 = 1;
            fChanges = 1;
        }
        Vec_IntFree( vDoms );
    }
    // count the number of PIs and internal nodes
    if fVeryVerbose != 0 {
        Gia_ManCountPisNodes( p, vPis, vAnds );
        fmt.Printf("Iter %3d :  ", Iter);
        fmt.Printf("PI = %5d  (%6.2f %%)  ", Vec_IntSize(vPis), 100.0 * Vec_IntSize(vPis)  / Gia_ManPiNum(p));
        fmt.Printf("And = %6d  (%6.2f %%) ", Vec_IntSize(vAnds), 100.0 * Vec_IntSize(vAnds) / Gia_ManAndNum(p));
//        printf( "Dom = %5d  (%6.2f %%)  ", Vec_IntSize(vDoms), 100.0 * Vec_IntSize(vDoms) / Gia_ManAndNum(p)  );
        fmt.Printf("\n");
    }
    // cleanup
    Vec_IntFree( vPis );
    Vec_IntFree( vAnds );
    Vec_IntFree( vSupp1 );
    Vec_IntFree( vSupp2 );
//    Gia_ManCleanMark1( p ); // this will erase markings
    p.pRefs = nil;
}

/**Function*************************************************************

  Synopsis    [Assumed that fMark1 marks the internal PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupRpm(p *Gia_Man_t) *Gia_Man_t {
    var vPis, vAnds *Vec_Int_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // derive PIs and internal nodes
    vPis = Vec_IntAlloc( 100 );
    vAnds = Vec_IntAlloc( 100 );
    Gia_ManCountPisNodes( p, vPis, vAnds );

    // duplicate AIG
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // create PIs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPis,i))); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // create flops
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // create internal nodes
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vAnds)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vAnds,i))); i++  {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
    // create COs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );

    // cleanup
    Vec_IntFree( vPis );
    Vec_IntFree( vAnds );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Performs structural reparametrization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_RpmPerform(p *Gia_Man_t, nCutMax int32, fVerbose int32, fVeryVerbose int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
//    Gia_ManTestDoms( p );
//    return NULL;
    // perform structural analysis
    Gia_ObjComputeTruthTableStart( p, nCutMax );
    Abs_RpmPerformMark( p, nCutMax, fVerbose, fVeryVerbose );
    Gia_ObjComputeTruthTableStop( p );
    // derive new AIG
    pNew = Gia_ManDupRpm( p );
    Gia_ManCleanMark1( p );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

