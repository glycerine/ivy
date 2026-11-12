package abs

import "fmt"

/**CFile****************************************************************

  FileName    [absDup.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Duplication procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absDup.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "abs.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAbsFlops_rec(pNew *Gia_Man_t, pObj *Gia_Obj_t) {
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupAbsFlops_rec( pNew, Gia_ObjFanin0(pObj) );
    Gia_ManDupAbsFlops_rec( pNew, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Extractes a flop-level abstraction given a flop map.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAbsFlops(p *Gia_Man_t, vFlopClasses *Vec_Int_t) *Gia_Man_t { 
    var pNew, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, nFlops int32;
    Gia_ManFillValue( p );
    // start the new manager
    pNew = Gia_ManStart( 5000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create PIs
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // create additional PIs
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if !Vec_IntEntry(vFlopClasses, i) {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    // create ROs
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Vec_IntEntry(vFlopClasses, i) {
            pObj.Value = Gia_ManAppendCi(pNew); }
    }
    // create POs
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDupAbsFlops_rec( pNew, Gia_ObjFanin0(pObj) );
        Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // create RIs
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        if Vec_IntEntry(vFlopClasses, i) {
            Gia_ManDupAbsFlops_rec( pNew, Gia_ObjFanin0(pObj) );
            Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
            nFlops++;
        }
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, nFlops );
    // clean up
    pNew = Gia_ManSeqCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Returns the array of neighbors.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaCollectAssigned(p *Gia_Man_t, vGateClasses *Vec_Int_t) *Vec_Int_t {
    var vAssigned *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, Entry int32;
    vAssigned = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vGateClasses)) && (((Entry) = Vec_IntEntry(vGateClasses, i)), 1); i++  {
        if Entry == 0 {
            continue; }
        assert(Entry > 0, "Entry > 0");
        pObj = Gia_ManObj( p, i );
        Vec_IntPush( vAssigned, Gia_ObjId(p, pObj) );
        if Gia_ObjIsAnd(pObj) {
            Vec_IntPush( vAssigned, Gia_ObjFaninId0p(p, pObj) );
            Vec_IntPush( vAssigned, Gia_ObjFaninId1p(p, pObj) );
        
        } else if Gia_ObjIsRo(p, pObj) {
            Vec_IntPush( vAssigned, Gia_ObjFaninId0p(p, Gia_ObjRoToRi(p, pObj)) );
        } else { assert(Gia_ObjIsConst0(pObj), "Gia_ObjIsConst0(pObj)"); }
    }
    Vec_IntUniqify( vAssigned );
    return vAssigned;
}

/**Function*************************************************************

  Synopsis    [Collects PIs and PPIs of the abstraction.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGlaCollect(p *Gia_Man_t, vGateClasses *Vec_Int_t, pvPis **Vec_Int_t, pvPPis **Vec_Int_t, pvFlops **Vec_Int_t, pvNodes **Vec_Int_t) { 
    var vAssigned *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    assert(Vec_IntSize(vGateClasses) == Gia_ManObjNum(p), "Vec_IntSize(vGateClasses) == Gia_ManObjNum(p)");
    // create included objects and their fanins
    vAssigned = Gia_GlaCollectAssigned( p, vGateClasses );
    // create additional arrays
    if pvPis {   *pvPis   = Vec_IntAlloc( 100 ); }
    if pvPPis {  *pvPPis  = Vec_IntAlloc( 100 ); }
    if pvFlops { *pvFlops = Vec_IntAlloc( 100 ); }
    if pvNodes { *pvNodes = Vec_IntAlloc( 1000 ); }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vAssigned)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vAssigned,i))); i++  {
        if Gia_ObjIsPi(p, pObj) { if pvPis {   Vec_IntPush( *pvPis, Gia_ObjId(p,pObj) ); }   
        } else if !Vec_IntEntry(vGateClasses, Gia_ObjId(p,pObj)) { if pvPPis {  Vec_IntPush( *pvPPis, Gia_ObjId(p,pObj) ); }  
        } else if Gia_ObjIsRo(p, pObj) { if pvFlops { Vec_IntPush( *pvFlops, Gia_ObjId(p,pObj) ); } 
        } else if Gia_ObjIsAnd(pObj) { if pvNodes { Vec_IntPush( *pvNodes, Gia_ObjId(p,pObj) ); } 
        } else { assert(Gia_ObjIsConst0(pObj), "Gia_ObjIsConst0(pObj)"); }
    }
    Vec_IntFree( vAssigned );
}

/**Function*************************************************************

  Synopsis    [Duplicates the AIG manager recursively.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAbsGates_rec(pNew *Gia_Man_t, pObj *Gia_Obj_t) {
    if ~pObj.Value {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDupAbsGates_rec( pNew, Gia_ObjFanin0(pObj) );
    Gia_ManDupAbsGates_rec( pNew, Gia_ObjFanin1(pObj) );
    pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
}

/**Function*************************************************************

  Synopsis    [Extractes a gate-level abstraction given a gate map.]

  Description [The array contains 1 for those objects (const, RO, AND)
  that are included in the abstraction; 0, otherwise.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupAbsGates(p *Gia_Man_t, vGateClasses *Vec_Int_t) *Gia_Man_t { 
    var vPis, vPPis, vFlops, vNodes *Vec_Int_t;
    var pNew, pTemp *Gia_Man_t;
    var pObj, pCopy *Gia_Obj_t;
    var i int32;//, nFlops = 0;
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    assert(Vec_IntSize(vGateClasses) == Gia_ManObjNum(p), "Vec_IntSize(vGateClasses) == Gia_ManObjNum(p)");

    // create additional arrays
    Gia_ManGlaCollect( p, vGateClasses, &vPis, &vPPis, &vFlops, &vNodes );

    // start the new manager
    pNew = Gia_ManStart( 5000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    // create constant
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // create PIs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPis,i))); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // create additional PIs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vPPis)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vPPis,i))); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // create ROs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vFlops)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vFlops,i))); i++  {
        pObj.Value = Gia_ManAppendCi(pNew);
    }
    // create internal nodes
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,i))); i++  {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    }
//        Gia_ManDupAbsGates_rec( pNew, pObj );
    // create PO
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // create RIs
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vFlops)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vFlops,i))); i++  {
        Gia_ObjRoToRi(p, pObj).Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(Gia_ObjRoToRi(p, pObj)) );
    }
    Gia_ManSetRegNum( pNew, Vec_IntSize(vFlops) );
    // clean up
    pNew = Gia_ManSeqCleanup( pTemp = pNew );
    // transfer copy values: (p -> pTemp -> pNew) => (p -> pNew)
    if Gia_ManObjNum(pTemp) != Gia_ManObjNum(pNew) {
//        printf( "Gia_ManDupAbsGates() Internal error: object mismatch.\n" );
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if !~pObj.Value {
                continue; }
            assert(!Abc_LitIsCompl(pObj.Value), "!Abc_LitIsCompl(pObj.Value)");
            pCopy = Gia_ObjCopy( pTemp, pObj );
            if !~pCopy.Value {
                Vec_IntWriteEntry( vGateClasses, i, 0 );
                pObj.Value = ~0;
                continue;
            }
            assert(!Abc_LitIsCompl(pCopy.Value), "!Abc_LitIsCompl(pCopy.Value)");
            pObj.Value = pCopy.Value;
        }
    }
    Gia_ManStop( pTemp );

    Vec_IntFree( vPis );
    Vec_IntFree( vPPis );
    Vec_IntFree( vFlops );
    Vec_IntFree( vNodes );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintFlopClasses(p *Gia_Man_t) {
    var Counter0, Counter1 int32;
    if p.vFlopClasses == nil {
        return; }
    if Vec_IntSize(p.vFlopClasses) != Gia_ManRegNum(p) {
        fmt.Printf("Gia_ManPrintFlopClasses(): The number of flop map entries differs from the number of flops.\n");
        return;
    }
    Counter0 = Vec_IntCountEntry( p.vFlopClasses, 0 );
    Counter1 = Vec_IntCountEntry( p.vFlopClasses, 1 );
    fmt.Printf("Flop-level abstraction:  Excluded FFs = %d  Included FFs = %d  (%.2f %%) ", Counter0, Counter1, 100.0*Counter1/(Counter0 + Counter1 + 1));
    if Counter0 + Counter1 < Gia_ManRegNum(p) {
        fmt.Printf("and there are other FF classes..."); }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintGateClasses(p *Gia_Man_t) {
    var vPis, vPPis, vFlops, vNodes *Vec_Int_t;
    var nTotal int32;
    if p.vGateClasses == nil {
        return; }
    if Vec_IntSize(p.vGateClasses) != Gia_ManObjNum(p) {
        fmt.Printf("Gia_ManPrintGateClasses(): The number of flop map entries differs from the number of flops.\n");
        return;
    }
    // create additional arrays
    Gia_ManGlaCollect( p, p.vGateClasses, &vPis, &vPPis, &vFlops, &vNodes );
    nTotal = 1 + Vec_IntSize(vFlops) + Vec_IntSize(vNodes);
    fmt.Printf("Gate-level abstraction:  PI = %d  PPI = %d  FF = %d (%.2f %%)  AND = %d (%.2f %%)  Obj = %d (%.2f %%)\n", Vec_IntSize(vPis), Vec_IntSize(vPPis), Vec_IntSize(vFlops), 100.0*Vec_IntSize(vFlops)/(Gia_ManRegNum(p)+1), Vec_IntSize(vNodes), 100.0*Vec_IntSize(vNodes)/(Gia_ManAndNum(p)+1), nTotal, 100.0*nTotal             /(Gia_ManRegNum(p)+Gia_ManAndNum(p)+1));
    Vec_IntFree( vPis );
    Vec_IntFree( vPPis );
    Vec_IntFree( vFlops );
    Vec_IntFree( vNodes );
}

/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintObjClasses(p *Gia_Man_t) {
    var vSeens *Vec_Int_t;  // objects seen so far
    var vAbs *Vec_Int_t  = p.vObjClasses;
    var i int32;
    var k int32;
    var Entry int32;
    var iStart int32;
    var iStop int32 = -1;
    var nFrames int32;
    var nObjBits, nObjMask, iObj, iFrame, nWords int32;
    var pInfo *uint32;
    var pCountAll, pCountUni *int32;
    if vAbs == nil {
        return; }
    nFrames = Vec_IntEntry( vAbs, 0 );
    assert(Vec_IntEntry(vAbs, nFrames+1) == Vec_IntSize(vAbs), "Vec_IntEntry(vAbs, nFrames+1) == Vec_IntSize(vAbs)");
    // inlining of macro ABC_ALLOC
    pCountAll = ((int *) malloc(sizeof(int) * (size_t)(nFrames + 1)));
    // inlining of macro ABC_ALLOC
    pCountUni = ((int *) malloc(sizeof(int) * (size_t)(nFrames + 1)));
    // start storage for seen objects
    nWords = Abc_BitWordNum( nFrames );
    vSeens = Vec_IntStart( Gia_ManObjNum(p) * nWords );
    // get the bitmasks
    nObjBits = Abc_Base2Log( Gia_ManObjNum(p) );
    nObjMask = (1 << nObjBits) - 1;
    assert(Gia_ManObjNum(p) <= nObjMask, "Gia_ManObjNum(p) <= nObjMask");
    // print info about frames
    fmt.Printf("Frame   Core   F0   F1   F2   F3 ...\n");
    for  i = 0; i < nFrames; i++  {
        iStart = Vec_IntEntry( vAbs, i+1 );
        iStop  = Vec_IntEntry( vAbs, i+2 );
        memset( pCountAll, 0, sizeof(int) * (nFrames + 1) );
        memset( pCountUni, 0, sizeof(int) * (nFrames + 1) );
        // inlining of macro Vec_IntForEachEntryStartStop
        for  k = iStart; (k < iStop) && (((Entry) = Vec_IntEntry(vAbs, k)), 1); k++  {
            iObj   = (Entry &  nObjMask);
            iFrame = (Entry >> nObjBits);
            pInfo  = (unsigned *)Vec_IntEntryP( vSeens, nWords * iObj );
            if Abc_InfoHasBit(pInfo, iFrame) == 0 {
                Abc_InfoSetBit( pInfo, iFrame );
                pCountUni[iFrame+1]++;
                pCountUni[0]++;
            }
            pCountAll[iFrame+1]++;
            pCountAll[0]++;
        }
        assert(pCountAll[0] == (iStop - iStart), "pCountAll[0] == (iStop - iStart)");
//        printf( "%5d%5d  ", pCountAll[0], pCountUni[0] ); 
        fmt.Printf("%3d :", i);
        fmt.Printf("%7d", pCountAll[0]); 
        if i >= 10 {
            for  k = 0; k < 4; k++  {
                fmt.Printf("%5d", pCountAll[k+1]);
            } 
            fmt.Printf("  ...");
            for  k = i-4; k <= i; k++  {
                fmt.Printf("%5d", pCountAll[k+1]);
            } 
        
        } else {
            for  k = 0; k <= i; k++  {
                if k <= i {
                    fmt.Printf("%5d", pCountAll[k+1]); }
            } 
        }
//        for ( k = 0; k < nFrames; k++ )
//            if ( k <= i )
//                printf( "%5d", pCountAll[k+1] ); 
        fmt.Printf("\n");
    }
    assert(iStop == Vec_IntSize(vAbs), "iStop == Vec_IntSize(vAbs)");
    Vec_IntFree( vSeens );
    pCountAll = nil;
    pCountUni = nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

