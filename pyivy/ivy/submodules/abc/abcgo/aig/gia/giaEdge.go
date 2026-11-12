package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaEdge.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Edge-related procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaEdge.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Gia_ObjEdgeCount(iObj int32, vEdge1 *Vec_Int_t, vEdge2 *Vec_Int_t) int32 {
    return (Vec_IntEntry(vEdge1, iObj) > 0) + (Vec_IntEntry(vEdge2, iObj) > 0);
}
func Gia_ObjEdgeAdd(iObj int32, iNext int32, vEdge1 *Vec_Int_t, vEdge2 *Vec_Int_t) int32 {
    var RetValue int32;
    if Vec_IntEntry(vEdge1, iObj) == 0 {
        Vec_IntWriteEntry(vEdge1, iObj, iNext);
    } else if Vec_IntEntry(vEdge2, iObj) == 0 {
        Vec_IntWriteEntry(vEdge2, iObj, iNext);
    } else { RetValue = 1; }
    return RetValue;
}
func Gia_ObjEdgeRemove(iObj int32, iNext int32, vEdge1 *Vec_Int_t, vEdge2 *Vec_Int_t) {
    assert(Vec_IntEntry(vEdge1, iObj) == iNext || Vec_IntEntry(vEdge2, iObj) == iNext, "Vec_IntEntry(vEdge1, iObj) == iNext || Vec_IntEntry(vEdge2, iObj) == iNext");
    if Vec_IntEntry(vEdge1, iObj) == iNext {
        Vec_IntWriteEntry( vEdge1, iObj, Vec_IntEntry(vEdge2, iObj) ); }
    Vec_IntWriteEntry( vEdge2, iObj, 0 );
}
func Gia_ObjEdgeClean(iObj int32, vEdge1 *Vec_Int_t, vEdge2 *Vec_Int_t) {
    Vec_IntWriteEntry( vEdge1, iObj, 0 );
    Vec_IntWriteEntry( vEdge2, iObj, 0 );
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transforms edge assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEdgeFromArray(p *Gia_Man_t, vArray *Vec_Int_t) {
    var i, iObj1, iObj2, Count int32;
    Vec_IntFreeP( &p.vEdge1 );
    Vec_IntFreeP( &p.vEdge2 );
    p.vEdge1 = Vec_IntStart( Gia_ManObjNum(p) );
    p.vEdge2 = Vec_IntStart( Gia_ManObjNum(p) );
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vArray)) && (((iObj1) = Vec_IntEntry(vArray, i)), 1) && (((iObj2) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
        assert(iObj1 < iObj2, "iObj1 < iObj2");
        Count += Gia_ObjEdgeAdd( iObj1, iObj2, p.vEdge1, p.vEdge2 );
        Count += Gia_ObjEdgeAdd( iObj2, iObj1, p.vEdge1, p.vEdge2 );
    }
    if Count != 0 { 
        fmt.Printf("Found %d violations during edge conversion.\n", Count); }
}
func Gia_ManEdgeToArray(p *Gia_Man_t) *Vec_Int_t {
    var iObj, iFanin int32;
    var vArray *Vec_Int_t  = Vec_IntAlloc( 1000 );
    assert(p.vEdge1 && p.vEdge2, "p.vEdge1 && p.vEdge2");
    assert(Vec_IntSize(p.vEdge1) == Gia_ManObjNum(p), "Vec_IntSize(p.vEdge1) == Gia_ManObjNum(p)");
    assert(Vec_IntSize(p.vEdge2) == Gia_ManObjNum(p), "Vec_IntSize(p.vEdge2) == Gia_ManObjNum(p)");
    for  iObj = 0; iObj < Gia_ManObjNum(p); iObj++  {
        iFanin = Vec_IntEntry( p.vEdge1, iObj );
        if (iFanin != 0) && iFanin < iObj {
            Vec_IntPushTwo( vArray, iFanin, iObj ); }
        iFanin = Vec_IntEntry( p.vEdge2, iObj );
        if (iFanin != 0) && iFanin < iObj {
            Vec_IntPushTwo( vArray, iFanin, iObj ); }
    }
    return vArray;
}

/**Function*************************************************************

  Synopsis    [Prints mapping statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManConvertPackingToEdges(p *Gia_Man_t) {
    var i int32;
    var k int32;
    var Entry int32;
    var nEntries int32;
    var nEntries2 int32;
    var nNodes [4]int32;
    var Count int32;
    if p.vPacking == nil {
        return; }
    Vec_IntFreeP( &p.vEdge1 );
    Vec_IntFreeP( &p.vEdge2 );
    p.vEdge1 = Vec_IntStart( Gia_ManObjNum(p) );
    p.vEdge2 = Vec_IntStart( Gia_ManObjNum(p) );
    // iterate through structures
    nEntries = Vec_IntEntry( p.vPacking, 0 );
    nEntries2 = 0;
    // inlining of macro Vec_IntForEachEntryStart
    for  i = 1; (i < Vec_IntSize(p.vPacking)) && (((Entry) = Vec_IntEntry(p.vPacking, i)), 1); i++  {
        assert(Entry > 0 && Entry < 4, "Entry > 0 && Entry < 4");
        i++;
        for  k = 0; k < Entry; k, i = k+1, i+1  {
            nNodes[k] = Vec_IntEntry(p.vPacking, i);
        }
        i--;
        nEntries2++;
        // create edges
        if Entry == 2 {
            Count += Gia_ObjEdgeAdd( nNodes[0], nNodes[1], p.vEdge1, p.vEdge2 );
            Count += Gia_ObjEdgeAdd( nNodes[1], nNodes[0], p.vEdge1, p.vEdge2 );
        
        } else if Entry == 3 {
            Count += Gia_ObjEdgeAdd( nNodes[0], nNodes[2], p.vEdge1, p.vEdge2 );
            Count += Gia_ObjEdgeAdd( nNodes[2], nNodes[0], p.vEdge1, p.vEdge2 );
            Count += Gia_ObjEdgeAdd( nNodes[1], nNodes[2], p.vEdge1, p.vEdge2 );
            Count += Gia_ObjEdgeAdd( nNodes[2], nNodes[1], p.vEdge1, p.vEdge2 );
        }
    }
    assert(nEntries == nEntries2, "nEntries == nEntries2");
    if Count != 0 {
        fmt.Printf("Skipped %d illegal edges.\n", Count); }
}

/**Function*************************************************************

  Synopsis    [Evaluates given edge assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjHaveEdge(p *Gia_Man_t, iObj int32, iNext int32) int32 {
    return Vec_IntEntry(p.vEdge1, iObj) == iNext || Vec_IntEntry(p.vEdge2, iObj) == iNext;
}
func Gia_ObjCheckEdge(p *Gia_Man_t, iObj int32, iNext int32) int32 {
    return Gia_ObjHaveEdge( p, iObj, iNext );
}
func Gia_ObjEvalEdgeDelay(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t) int32 {
    var nEdgeDelay int32  = 2;
    var i, iFan, Delay, DelayMax int32;
    if Gia_ManHasMapping(p) && Gia_ObjIsLut(p, iObj) {
        assert(Gia_ObjLutSize(p, iObj) <= 4, "Gia_ObjLutSize(p, iObj) <= 4");
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p,iObj) && ((iFan = Gia_ObjLutFanins(p,iObj)[i]),1); i++  {
            Delay = Vec_IntEntry(vDelay, iFan) + (Gia_ObjHaveEdge(p, iObj, iFan) ? nEdgeDelay : 10);
            DelayMax = Abc_MaxInt( DelayMax, Delay );
        }
    
    } else if Gia_ObjIsLut2(p, iObj) {
        assert(Gia_ObjLutSize2(p, iObj) <= 4, "Gia_ObjLutSize2(p, iObj) <= 4");
        // inlining of macro Gia_LutForEachFanin2
        for  i = 0; i < Gia_ObjLutSize2(p,iObj) && ((iFan = Gia_ObjLutFanin2(p,iObj,i)),1); i++  {
            Delay = Vec_IntEntry(vDelay, iFan) + (Gia_ObjHaveEdge(p, iObj, iFan) ? nEdgeDelay : 10);
            DelayMax = Abc_MaxInt( DelayMax, Delay );
        }
    
    } else { assert(false, "0"); }
    return DelayMax;
}
func Gia_ManEvalEdgeDelay(p *Gia_Man_t) int32 {
    var k, iLut, DelayMax int32;
    assert(p.vEdge1 && p.vEdge2, "p.vEdge1 && p.vEdge2");
    Vec_IntFreeP( &p.vEdgeDelay );
    p.vEdgeDelay = Vec_IntStart( Gia_ManObjNum(p) );
    if Gia_ManHasMapping(p) {
        if p.pManTime != nil && Tim_ManBoxNum((Tim_Man_t*)p.pManTime) {
            var pObj *Gia_Obj_t; 
            var vNodes *Vec_Int_t  = Gia_ManOrderWithBoxes( p );
            Tim_ManIncrementTravId( (Tim_Man_t*)p.pManTime );
            // inlining of macro Gia_ManForEachObjVec
            for  k = 0; (k < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,k))); k++  {
                iLut = Gia_ObjId( p, pObj );
                if Gia_ObjIsAnd(pObj) {
                    if Gia_ObjIsLut(p, iLut) {
                        Vec_IntWriteEntry( p.vEdgeDelay, iLut, Gia_ObjEvalEdgeDelay(p, iLut, p.vEdgeDelay) ); }
                
                } else if Gia_ObjIsCi(pObj) {
                    var arrTime int32  = Tim_ManGetCiArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj) );
                    Vec_IntWriteEntry( p.vEdgeDelay,  iLut, arrTime );
                
                } else if Gia_ObjIsCo(pObj) {
                    var arrTime int32  = Vec_IntEntry( p.vEdgeDelay, Gia_ObjFaninId0(pObj, iLut) );
                    Tim_ManSetCoArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj), arrTime );
                
                } else if !Gia_ObjIsConst0(pObj) { 
                    assert(false, "0"); }
            }
            Vec_IntFree( vNodes );
        
        } else {
            // inlining of macro Gia_ManForEachLut
            for  iLut = 1; iLut < Gia_ManObjNum(p); iLut++  { if !Gia_ObjIsLut(p, iLut) {} else {
                Vec_IntWriteEntry( p.vEdgeDelay, iLut, Gia_ObjEvalEdgeDelay(p, iLut, p.vEdgeDelay) ); }
            }
        }
    
    } else if Gia_ManHasMapping2(p) {
        if p.pManTime != nil && Tim_ManBoxNum((Tim_Man_t*)p.pManTime) {
            var pObj *Gia_Obj_t; 
            var vNodes *Vec_Int_t  = Gia_ManOrderWithBoxes( p );
            Tim_ManIncrementTravId( (Tim_Man_t*)p.pManTime );
            // inlining of macro Gia_ManForEachObjVec
            for  k = 0; (k < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,k))); k++  {
                iLut = Gia_ObjId( p, pObj );
                if Gia_ObjIsAnd(pObj) {
                    if Gia_ObjIsLut2(p, iLut) {
                        Vec_IntWriteEntry( p.vEdgeDelay, iLut, Gia_ObjEvalEdgeDelay(p, iLut, p.vEdgeDelay) ); }
                
                } else if Gia_ObjIsCi(pObj) {
                    var arrTime int32  = Tim_ManGetCiArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj) );
                    Vec_IntWriteEntry( p.vEdgeDelay,  iLut, arrTime );
                
                } else if Gia_ObjIsCo(pObj) {
                    var arrTime int32  = Vec_IntEntry( p.vEdgeDelay, Gia_ObjFaninId0(pObj, iLut) );
                    Tim_ManSetCoArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj), arrTime );
                
                } else if !Gia_ObjIsConst0(pObj) { 
                    assert(false, "0"); }
            }
            Vec_IntFree( vNodes );
        
        } else {
            // inlining of macro Gia_ManForEachLut2
            for  iLut = 1; iLut < Gia_ManObjNum(p); iLut++  { if !Gia_ObjIsLut2(p, iLut) {} else {
                Vec_IntWriteEntry( p.vEdgeDelay, iLut, Gia_ObjEvalEdgeDelay(p, iLut, p.vEdgeDelay) ); }
            }
        }
    
    } else { assert(false, "0"); }
    // inlining of macro Gia_ManForEachCoDriverId
    for  k = 0; (k < Vec_IntSize(p.vCos)) && (((iLut) = Gia_ObjFaninId0p(p, Gia_ManCo(p, k))), 1); k++  {
        DelayMax = Abc_MaxInt( DelayMax, Vec_IntEntry(p.vEdgeDelay, iLut) );
    }
    return DelayMax;
}
func Gia_ManEvalEdgeCount(p *Gia_Man_t) int32 {
    return (Vec_IntCountPositive(p.vEdge1) + Vec_IntCountPositive(p.vEdge2))/2;
}


/**Function*************************************************************

  Synopsis    [Finds edge assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeEdgeDelay(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t, vEdge1 *Vec_Int_t, vEdge2 *Vec_Int_t, fUseTwo int32) int32 {
    var i, iFan, Delay, Status1, Status2 int32;
    var DelayMax, DelayMax2, nCountMax int32;
    var iFanMax1 int32 = -1;
    var iFanMax2 int32 = -1;
    Vec_IntWriteEntry(vEdge1, iObj, 0);
    Vec_IntWriteEntry(vEdge2, iObj, 0);
    if Gia_ManHasMapping(p) && Gia_ObjIsLut(p, iObj) {
        assert(Gia_ObjLutSize(p, iObj) <= 4, "Gia_ObjLutSize(p, iObj) <= 4");
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p,iObj) && ((iFan = Gia_ObjLutFanins(p,iObj)[i]),1); i++  {
            Delay = Vec_IntEntry( vDelay, iFan ) + 10;
            if DelayMax < Delay {
                DelayMax2 = DelayMax;
                DelayMax  = Delay;
                iFanMax1  = iFan;
                nCountMax = 1;
            
            } else if DelayMax == Delay {
                iFanMax2  = iFan;
                nCountMax++;
                if !(fUseTwo != 0) {
                    DelayMax2 = DelayMax; }
            
            } else {
                DelayMax2 = Abc_MaxInt( DelayMax2, Delay ); }
        }
    
    } else if Gia_ObjIsLut2(p, iObj) {
        assert(Gia_ObjLutSize2(p, iObj) <= 4, "Gia_ObjLutSize2(p, iObj) <= 4");
        // inlining of macro Gia_LutForEachFanin2
        for  i = 0; i < Gia_ObjLutSize2(p,iObj) && ((iFan = Gia_ObjLutFanin2(p,iObj,i)),1); i++  {
            Delay = Vec_IntEntry( vDelay, iFan ) + 10;
            if DelayMax < Delay {
                DelayMax2 = DelayMax;
                DelayMax  = Delay;
                iFanMax1  = iFan;
                nCountMax = 1;
            
            } else if DelayMax == Delay {
                iFanMax2  = iFan;
                nCountMax++;
                if !(fUseTwo != 0) {
                    DelayMax2 = DelayMax; }
            
            } else {
                DelayMax2 = Abc_MaxInt( DelayMax2, Delay ); }
        }
    
    } else { assert(false, "0"); }
    assert(nCountMax > 0, "nCountMax > 0");
    if DelayMax <= 10 { // skip first level
    } else if nCountMax == 1 {
        Status1 = Gia_ObjEdgeCount( iFanMax1, vEdge1, vEdge2 );
        if Status1 <= 1 {
            Gia_ObjEdgeAdd( iFanMax1, iObj, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iObj, iFanMax1, vEdge1, vEdge2 );
            DelayMax = Abc_MaxInt( DelayMax2, DelayMax - 8 );
            Vec_IntWriteEntry( vDelay, iObj, DelayMax );
            return DelayMax;
        }
    
    } else if (fUseTwo != 0) && nCountMax == 2 {
        Status1 = Gia_ObjEdgeCount( iFanMax1, vEdge1, vEdge2 );
        Status2 = Gia_ObjEdgeCount( iFanMax2, vEdge1, vEdge2 );
        if Status1 <= 1 && Status2 <= 1 {
            Gia_ObjEdgeAdd( iFanMax1, iObj, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iFanMax2, iObj, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iObj, iFanMax1, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iObj, iFanMax2, vEdge1, vEdge2 );
            DelayMax = Abc_MaxInt( DelayMax2, DelayMax - 8 );
            Vec_IntWriteEntry( vDelay, iObj, DelayMax );
            return DelayMax;
        }
    }
    Vec_IntWriteEntry( vDelay, iObj, DelayMax );
    return DelayMax;
}
func Gia_ManComputeEdgeDelay(p *Gia_Man_t, fUseTwo int32) int32 {
    var k, iLut, DelayMax int32;
    Vec_IntFreeP( &p.vEdgeDelay );
    Vec_IntFreeP( &p.vEdge1 );
    Vec_IntFreeP( &p.vEdge2 );
    p.vEdge1 = Vec_IntStart( Gia_ManObjNum(p) );
    p.vEdge2 = Vec_IntStart( Gia_ManObjNum(p) );
    p.vEdgeDelay = Vec_IntStart( Gia_ManObjNum(p) );
    if Gia_ManHasMapping(p) {
        if p.pManTime != nil && Tim_ManBoxNum((Tim_Man_t*)p.pManTime) {
            var pObj *Gia_Obj_t; 
            var vNodes *Vec_Int_t  = Gia_ManOrderWithBoxes( p );
            Tim_ManIncrementTravId( (Tim_Man_t*)p.pManTime );
            // inlining of macro Gia_ManForEachObjVec
            for  k = 0; (k < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,k))); k++  {
                iLut = Gia_ObjId( p, pObj );
                if Gia_ObjIsAnd(pObj) {
                    if Gia_ObjIsLut(p, iLut) {
                        Gia_ObjComputeEdgeDelay( p, iLut, p.vEdgeDelay, p.vEdge1, p.vEdge2, fUseTwo ); }
                
                } else if Gia_ObjIsCi(pObj) {
                    var arrTime int32  = Tim_ManGetCiArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj) );
                    Vec_IntWriteEntry( p.vEdgeDelay,  iLut, arrTime );
                
                } else if Gia_ObjIsCo(pObj) {
                    var arrTime int32  = Vec_IntEntry( p.vEdgeDelay, Gia_ObjFaninId0(pObj, iLut) );
                    Tim_ManSetCoArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj), arrTime );
                
                } else if !Gia_ObjIsConst0(pObj) { 
                    assert(false, "0"); }
            }
            Vec_IntFree( vNodes );
        
        } else {
            // inlining of macro Gia_ManForEachLut
            for  iLut = 1; iLut < Gia_ManObjNum(p); iLut++  { if !Gia_ObjIsLut(p, iLut) {} else {
                Gia_ObjComputeEdgeDelay( p, iLut, p.vEdgeDelay, p.vEdge1, p.vEdge2, fUseTwo ); }
            }
        }
    
    } else if Gia_ManHasMapping2(p) {
        if p.pManTime != nil && Tim_ManBoxNum((Tim_Man_t*)p.pManTime) {
            var pObj *Gia_Obj_t; 
            var vNodes *Vec_Int_t  = Gia_ManOrderWithBoxes( p );
            Tim_ManIncrementTravId( (Tim_Man_t*)p.pManTime );
            // inlining of macro Gia_ManForEachObjVec
            for  k = 0; (k < Vec_IntSize(vNodes)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vNodes,k))); k++  {
                iLut = Gia_ObjId( p, pObj );
                if Gia_ObjIsAnd(pObj) {
                    if Gia_ObjIsLut2(p, iLut) {
                        Gia_ObjComputeEdgeDelay( p, iLut, p.vEdgeDelay, p.vEdge1, p.vEdge2, fUseTwo ); }
                
                } else if Gia_ObjIsCi(pObj) {
                    var arrTime int32  = Tim_ManGetCiArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj) );
                    Vec_IntWriteEntry( p.vEdgeDelay,  iLut, arrTime );
                
                } else if Gia_ObjIsCo(pObj) {
                    var arrTime int32  = Vec_IntEntry( p.vEdgeDelay, Gia_ObjFaninId0(pObj, iLut) );
                    Tim_ManSetCoArrival( (Tim_Man_t*)p.pManTime, Gia_ObjCioId(pObj), arrTime );
                
                } else if !Gia_ObjIsConst0(pObj) { 
                    assert(false, "0"); }
            }
            Vec_IntFree( vNodes );
        
        } else {
            // inlining of macro Gia_ManForEachLut2
            for  iLut = 1; iLut < Gia_ManObjNum(p); iLut++  { if !Gia_ObjIsLut2(p, iLut) {} else {
                Gia_ObjComputeEdgeDelay( p, iLut, p.vEdgeDelay, p.vEdge1, p.vEdge2, fUseTwo ); }
            }
        }
    
    } else { assert(false, "0"); }
    // inlining of macro Gia_ManForEachCoDriverId
    for  k = 0; (k < Vec_IntSize(p.vCos)) && (((iLut) = Gia_ObjFaninId0p(p, Gia_ManCo(p, k))), 1); k++  {
        DelayMax = Abc_MaxInt( DelayMax, Vec_IntEntry(p.vEdgeDelay, iLut) );
    }
    //printf( "The number of edges = %d.  Delay = %d.\n", Gia_ManEvalEdgeCount(p), DelayMax );
    return DelayMax;
}

/**Function*************************************************************

  Synopsis    [Finds edge assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjComputeEdgeDelay2(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t, vEdge1 *Vec_Int_t, vEdge2 *Vec_Int_t, vFanMax1 *Vec_Int_t, vFanMax2 *Vec_Int_t, vCountMax *Vec_Int_t) int32 {
    var i, iFan, DelayFanin, Status1, Status2 int32;
    var DelayMax, nCountMax int32;
    var iFanMax1 int32 = -1;
    var iFanMax2 int32 = -1;
    Vec_IntWriteEntry(vEdge1, iObj, 0);
    Vec_IntWriteEntry(vEdge2, iObj, 0);
    // analyze this node
    DelayMax = Vec_IntEntry( vDelay, iObj );
    nCountMax = Vec_IntEntry( vCountMax, iObj );
    if DelayMax == 0 {
    } else if nCountMax == 1 {
        iFanMax1 = Vec_IntEntry( vFanMax1, iObj );
        Status1 = Gia_ObjEdgeCount( iFanMax1, vEdge1, vEdge2 );
        if Status1 <= 1 {
            Gia_ObjEdgeAdd( iFanMax1, iObj, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iObj, iFanMax1, vEdge1, vEdge2 );
            DelayMax--;
        }
    
    } else if nCountMax == 2 {
        iFanMax1 = Vec_IntEntry( vFanMax1, iObj );
        iFanMax2 = Vec_IntEntry( vFanMax2, iObj );
        Status1 = Gia_ObjEdgeCount( iFanMax1, vEdge1, vEdge2 );
        Status2 = Gia_ObjEdgeCount( iFanMax2, vEdge1, vEdge2 );
        if Status1 <= 1 && Status2 <= 1 {
            Gia_ObjEdgeAdd( iFanMax1, iObj, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iFanMax2, iObj, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iObj, iFanMax1, vEdge1, vEdge2 );
            Gia_ObjEdgeAdd( iObj, iFanMax2, vEdge1, vEdge2 );
            DelayMax--;
        }
    }
    Vec_IntWriteEntry( vDelay, iObj, DelayMax );
    // computed DelayMax at this point
    if Gia_ManHasMapping(p) && Gia_ObjIsLut(p, iObj) {
        // inlining of macro Gia_LutForEachFanin
        for  i = 0; i < Gia_ObjLutSize(p,iObj) && ((iFan = Gia_ObjLutFanins(p,iObj)[i]),1); i++  {
            DelayFanin = Vec_IntEntry( vDelay, iFan );
            if DelayFanin < DelayMax + 1 {
                Vec_IntWriteEntry( vDelay, iFan, DelayMax + 1 );
                Vec_IntWriteEntry( vFanMax1, iFan, iObj );
                Vec_IntWriteEntry( vCountMax, iFan, 1 );
            
            } else if DelayFanin == DelayMax + 1 {
                Vec_IntWriteEntry( vFanMax2, iFan, iObj );
                Vec_IntAddToEntry( vCountMax, iFan, 1 );
            }
        }
    
    } else if Gia_ObjIsLut2(p, iObj) {
        // inlining of macro Gia_LutForEachFanin2
        for  i = 0; i < Gia_ObjLutSize2(p,iObj) && ((iFan = Gia_ObjLutFanin2(p,iObj,i)),1); i++  {
            DelayFanin = Vec_IntEntry( vDelay, iFan );
            if DelayFanin < DelayMax + 1 {
                Vec_IntWriteEntry( vDelay, iFan, DelayMax + 1 );
                Vec_IntWriteEntry( vFanMax1, iFan, iObj );
                Vec_IntWriteEntry( vCountMax, iFan, 1 );
            
            } else if DelayFanin == DelayMax + 1 {
                Vec_IntWriteEntry( vFanMax2, iFan, iObj );
                Vec_IntAddToEntry( vCountMax, iFan, 1 );
            }
        }
    
    } else { assert(false, "0"); }
    return DelayMax;
}
func Gia_ManComputeEdgeDelay2(p *Gia_Man_t) int32 {
    var k, iLut, DelayMax int32;
    var vFanMax1 *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vFanMax2 *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vCountMax *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    assert(p.pManTime == nil, "p.pManTime == nil");
    Vec_IntFreeP( &p.vEdgeDelay );
    Vec_IntFreeP( &p.vEdge1 );
    Vec_IntFreeP( &p.vEdge2 );
    p.vEdgeDelay = Vec_IntStart( Gia_ManObjNum(p) );
    p.vEdge1 = Vec_IntStart( Gia_ManObjNum(p) );
    p.vEdge2 = Vec_IntStart( Gia_ManObjNum(p) );
//    Gia_ManForEachCoDriverId( p, iLut, k )
//        Vec_IntWriteEntry( p->vEdgeDelay, iLut, 1 );
    if Gia_ManHasMapping(p) {
        // inlining of macro Gia_ManForEachLutReverse
        for  iLut = Gia_ManObjNum(p) - 1; iLut > 0; iLut--  { if !Gia_ObjIsLut(p, iLut) {} else {
            Gia_ObjComputeEdgeDelay2( p, iLut, p.vEdgeDelay, p.vEdge1, p.vEdge2, vFanMax1, vFanMax2, vCountMax ); }
        }
    } else if Gia_ManHasMapping2(p) {
        // inlining of macro Gia_ManForEachLut2Reverse
        for  iLut = Gia_ManObjNum(p) - 1; iLut > 0; iLut--  { if !Gia_ObjIsLut2(p, iLut) {} else {
            Gia_ObjComputeEdgeDelay2( p, iLut, p.vEdgeDelay, p.vEdge1, p.vEdge2, vFanMax1, vFanMax2, vCountMax ); }
        }
    } else { assert(false, "0"); }
    // inlining of macro Gia_ManForEachCiId
    for  k = 0; (k < Vec_IntSize(p.vCis)) && ((iLut) = Gia_ObjId(p, Gia_ManCi(p, k))); k++  {
        DelayMax = Abc_MaxInt( DelayMax, Vec_IntEntry(p.vEdgeDelay, iLut) );
    }
    Vec_IntFree( vFanMax1 );
    Vec_IntFree( vFanMax2 );
    Vec_IntFree( vCountMax );
    //printf( "The number of edges = %d.  Delay = %d.\n", Gia_ManEvalEdgeCount(p), DelayMax );
    return DelayMax;
}

/**Function*************************************************************

  Synopsis    [Finds edge assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUpdateMapping(p *Gia_Man_t, vNodes *Vec_Int_t, vWin *Vec_Wec_t) {
    var i, iNode int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vNodes)) && (((iNode) = Vec_IntEntry(vNodes, i)), 1); i++  {
        // inlining of macro ABC_SWAP
         var t Vec_Int_t  = *Vec_WecEntry(p.vMapping2, iNode); *Vec_WecEntry(p.vMapping2, iNode) = *Vec_WecEntry(vWin, i); *Vec_WecEntry(vWin, i) = t; };
}
func Gia_ManEvalWindowInc(p *Gia_Man_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t, vWin *Vec_Wec_t, vTemp *Vec_Int_t, fUseTwo int32) int32 {
    var i, iLut, Delay, DelayMax int32;
    assert(Vec_IntSize(vNodes) == Vec_WecSize(vWin), "Vec_IntSize(vNodes) == Vec_WecSize(vWin)");
    Gia_ManUpdateMapping( p, vNodes, vWin );
    Gia_ManCollectTfo( p, vLeaves, vTemp );
    Vec_IntReverseOrder( vTemp );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vTemp)) && (((iLut) = Vec_IntEntry(vTemp, i)), 1); i++  {
        if !Gia_ObjIsLut(p, iLut) {
            continue; }
        Delay = Gia_ObjComputeEdgeDelay( p, iLut, p.vEdgeDelay, p.vEdge1, p.vEdge2, fUseTwo );
        DelayMax = Abc_MaxInt( DelayMax, Delay );
    }
    Gia_ManUpdateMapping( p, vNodes, vWin );
    return DelayMax;
}
func Gia_ManEvalWindow(p *Gia_Man_t, vLeaves *Vec_Int_t, vNodes *Vec_Int_t, vWin *Vec_Wec_t, vTemp *Vec_Int_t, fUseTwo int32) int32 {
    var DelayMax int32;
    assert(Vec_IntSize(vNodes) == Vec_WecSize(vWin), "Vec_IntSize(vNodes) == Vec_WecSize(vWin)");
    Gia_ManUpdateMapping( p, vNodes, vWin );
    DelayMax = Gia_ManComputeEdgeDelay( p, fUseTwo );
    Gia_ManUpdateMapping( p, vNodes, vWin );
    return DelayMax;
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Edg_ManToMapping(p *Gia_Man_t) {
    var iObj, iFanin, k int32;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    Vec_WecFreeP( &p.vMapping2 );
    Vec_WecFreeP( &p.vFanouts2 );
    p.vMapping2 = Vec_WecStart( Gia_ManObjNum(p) );
    p.vFanouts2 = Vec_WecStart( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachLut
    for  iObj = 1; iObj < Gia_ManObjNum(p); iObj++  { if !Gia_ObjIsLut(p, iObj) {} else {
        assert(Gia_ObjLutSize(p, iObj) <= 4, "Gia_ObjLutSize(p, iObj) <= 4");
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,iObj) && ((iFanin = Gia_ObjLutFanins(p,iObj)[k]),1); k++  {
            Vec_WecPush( p.vMapping2, iObj, iFanin );
            Vec_WecPush( p.vFanouts2, iFanin, iObj );
        }
    }
    }
}

/**Function*************************************************************

  Synopsis    [Computes delay for the given edge assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Edg_ObjEvalEdgeDelay(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t) int32 {
    var DelayEdge int32; // 2;
    var DelayNoEdge int32  = 1;
    var i, iFan, Delay, DelayMax int32;
    assert(Gia_ObjIsLut2(p, iObj), "Gia_ObjIsLut2(p, iObj)");
    // inlining of macro Gia_LutForEachFanin2
    for  i = 0; i < Gia_ObjLutSize2(p,iObj) && ((iFan = Gia_ObjLutFanin2(p,iObj,i)),1); i++  {
        Delay = Vec_IntEntry(vDelay, iFan) + (Gia_ObjHaveEdge(p, iObj, iFan) ? DelayEdge : DelayNoEdge);
        DelayMax = Abc_MaxInt( DelayMax, Delay );
    }
    //printf( "Obj %d - Level %d\n", iObj, DelayMax );
    return DelayMax;
}
func Edg_ManEvalEdgeDelay(p *Gia_Man_t) int32 {
    var iLut, Delay, DelayMax int32;
    assert(p.vEdge1 && p.vEdge2, "p.vEdge1 && p.vEdge2");
    if p.vEdgeDelay == nil {
        p.vEdgeDelay = Vec_IntStart( Gia_ManObjNum(p) );
    } else {
        Vec_IntFill( p.vEdgeDelay, Gia_ManObjNum(p), 0 ); }
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(p); iLut++  { if !Gia_ObjIsLut2(p, iLut) {} else {
        Delay = Edg_ObjEvalEdgeDelay(p, iLut, p.vEdgeDelay);
        Vec_IntWriteEntry( p.vEdgeDelay, iLut, Delay );
        DelayMax = Abc_MaxInt( DelayMax, Delay );
    }
    }
    return DelayMax;
}

func Edg_ObjEvalEdgeDelayR(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t) int32 {
    var DelayEdge int32; // 2;
    var DelayNoEdge int32  = 1;
    var i, iFan, Delay, DelayMax int32;
    assert(Gia_ObjIsLut2(p, iObj), "Gia_ObjIsLut2(p, iObj)");
    // inlining of macro Gia_LutForEachFanout2
    for  i = 0; i < Gia_ObjLutFanoutNum2(p,iObj) && ((iFan = Gia_ObjLutFanout2(p,iObj,i)),1); i++  {
        Delay = Vec_IntEntry(vDelay, iFan) + (Gia_ObjHaveEdge(p, iObj, iFan) ? DelayEdge : DelayNoEdge);
        DelayMax = Abc_MaxInt( DelayMax, Delay );
    }
    //printf( "Obj %d - LevelR %d\n", iObj, DelayMax );
    return DelayMax;
}
func Edg_ManEvalEdgeDelayR(p *Gia_Man_t) int32 {
//    int k, DelayNoEdge = 1;
    var iLut, Delay, DelayMax int32;
    assert(p.vEdge1 && p.vEdge2, "p.vEdge1 && p.vEdge2");
    if p.vEdgeDelayR == nil {
        p.vEdgeDelayR = Vec_IntStart( Gia_ManObjNum(p) );
    } else {
        Vec_IntFill( p.vEdgeDelayR, Gia_ManObjNum(p), 0 ); }
//    Gia_ManForEachCoDriverId( p, iLut, k )
//        Vec_IntWriteEntry( p->vEdgeDelayR, iLut, DelayNoEdge );
    // inlining of macro Gia_ManForEachLut2Reverse
    for  iLut = Gia_ManObjNum(p) - 1; iLut > 0; iLut--  { if !Gia_ObjIsLut2(p, iLut) {} else {
        Delay = Edg_ObjEvalEdgeDelayR(p, iLut, p.vEdgeDelayR);
        Vec_IntWriteEntry( p.vEdgeDelayR, iLut, Delay );
        DelayMax = Abc_MaxInt( DelayMax, Delay );
    }
    }
    return DelayMax;
}

func Edg_ManCollectCritEdges(p *Gia_Man_t, vEdges *Vec_Wec_t, DelayMax int32) {
    var vLevel *Vec_Int_t;
    var k, iLut, Delay1, Delay2 int32;
    assert(p.vEdge1 && p.vEdge2, "p.vEdge1 && p.vEdge2");
    Vec_WecClear( vEdges );
    Vec_WecInit( vEdges, DelayMax + 1 );
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(p); iLut++  { if !Gia_ObjIsLut2(p, iLut) {} else {
        Delay1 = Vec_IntEntry( p.vEdgeDelay, iLut );
        Delay2 = Vec_IntEntry( p.vEdgeDelayR, iLut );
        assert(Delay1 + Delay2 <= DelayMax, "Delay1 + Delay2 <= DelayMax");
        if Delay1 + Delay2 == DelayMax {
            Vec_WecPush( vEdges, Delay1, iLut ); }
    }
    }
    // every level should have critical nodes, except the first one
    //Vec_WecPrint( vEdges, 0 );
    // inlining of macro Vec_WecForEachLevelStart
    for  k = 1; (k < Vec_WecSize(vEdges)) && (((vLevel) = Vec_WecEntry(vEdges, k)), 1); k++  {
        assert(Vec_IntSize(vLevel) > 0, "Vec_IntSize(vLevel) > 0");
    }
}


/**Function*************************************************************

  Synopsis    [Update one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Edg_ObjImprove(p *Gia_Man_t, iObj int32, nEdgeLimit int32, DelayMax int32, fVerbose int32) int32 {
    var nFaninsC, nFanoutsC int32;    // critical
    var nFaninsEC, nFanoutsEC int32;   // edge-critical
    var nFaninsENC, nFanoutsENC int32;  // edge-non-critial
    var pFanins, pFanouts [4]int32;
    var nEdgeDiff, nEdges, Count int32;
    var i, iNext, Delay1, Delay2 int32;
    // count how many fanins have critical edge
    Delay1 = Vec_IntEntry( p.vEdgeDelayR, iObj );
    //if ( Delay1 > 1 )
    // inlining of macro Gia_LutForEachFanin2
    for  i = 0; i < Gia_ObjLutSize2(p,iObj) && ((iNext = Gia_ObjLutFanin2(p,iObj,i)),1); i++  {
        if !Gia_ObjIsAnd(Gia_ManObj(p, iNext)) {
            continue; }
        Delay2 = Vec_IntEntry( p.vEdgeDelay, iNext );
        if Gia_ObjHaveEdge(p, iObj, iNext) != 0 {
            nEdges++;
            assert(Delay1 + Delay2 <= DelayMax, "Delay1 + Delay2 <= DelayMax");
            if Delay1 + Delay2 == DelayMax {
                nFaninsEC++;
            } else {
                nFaninsENC++; }
        
        } else {
            assert(Delay1 + Delay2 + 1 <= DelayMax, "Delay1 + Delay2 + 1 <= DelayMax");
            if Delay1 + Delay2 + 1 == DelayMax {
                pFanins[nFaninsC++] = iNext; }
        }
    }
    // count how many fanouts have critical edge
    Delay1 = Vec_IntEntry( p.vEdgeDelay, iObj );
    //if ( Delay2 < DelayMax - 1 )
    // inlining of macro Gia_LutForEachFanout2
    for  i = 0; i < Gia_ObjLutFanoutNum2(p,iObj) && ((iNext = Gia_ObjLutFanout2(p,iObj,i)),1); i++  {
        //if ( !Gia_ObjIsAnd(Gia_ManObj(p, iNext)) )
        //    continue;
        assert(Gia_ObjIsAnd(Gia_ManObj(p, iNext)), "Gia_ObjIsAnd(Gia_ManObj(p, iNext))");
        Delay2 = Vec_IntEntry( p.vEdgeDelayR, iNext );
        if Gia_ObjHaveEdge(p, iObj, iNext) != 0 {
            nEdges++;
            assert(Delay1 + Delay2 <= DelayMax, "Delay1 + Delay2 <= DelayMax");
            if Delay1 + Delay2 == DelayMax {
                nFanoutsEC++;
            } else {
                nFanoutsENC++; }
        
        } else {
            assert(Delay1 + Delay2 + 1 <= DelayMax, "Delay1 + Delay2 + 1 <= DelayMax");
            if Delay1 + Delay2 + 1 == DelayMax {
                if nFanoutsC < nEdgeLimit {
                    pFanouts[nFanoutsC] = iNext; }
                nFanoutsC++;
            }
        }
    }
    if fVerbose != 0 {
        fmt.Printf("%8d : ", iObj);
        fmt.Printf("Edges = %d  ", nEdges);
        fmt.Printf("Fanins (all %d  EC %d  ENC %d  C %d)  ", Gia_ObjLutSize2(p, iObj), nFaninsEC, nFaninsENC, nFaninsC);
        fmt.Printf("Fanouts (all %d  EC %d  ENC %d  C %d)  ", Gia_ObjLutFanoutNum2(p, iObj), nFanoutsEC, nFanoutsENC, nFanoutsC);
    }
    // consider simple cases
    assert(nEdges <= nEdgeLimit, "nEdges <= nEdgeLimit");
    if nEdges == nEdgeLimit {
        if fVerbose != 0 {
            fmt.Printf("Full\n"); }
        return 0;
    }
    nEdgeDiff = nEdgeLimit - nEdges;
    // check if fanins or fanouts could be improved
    if nFaninsEC == 0 && nFaninsC && nFaninsC <= nEdgeDiff {
        for  i = 0; i < nFaninsC; i++  {
            if Gia_ObjEdgeCount(pFanins[i], p.vEdge1, p.vEdge2) == nEdgeLimit {
                break; }
        }
        if i == nFaninsC {
            for  i = 0; i < nFaninsC; i++  {
                Count += Gia_ObjEdgeAdd( iObj, pFanins[i], p.vEdge1, p.vEdge2 );
                Count += Gia_ObjEdgeAdd( pFanins[i], iObj, p.vEdge1, p.vEdge2 );
            }
            if Count != 0 {
                fmt.Printf("Wrong number of edges.\n"); }
            if fVerbose != 0 {
                fmt.Printf("Fixed %d critical fanins\n", nFaninsC); }
            return 1;
        }
    }
    if nFanoutsEC == 0 && (nFanoutsC != 0) && nFanoutsC <= nEdgeDiff {
        for  i = 0; i < nFanoutsC; i++  {
            if Gia_ObjEdgeCount(pFanouts[i], p.vEdge1, p.vEdge2) == nEdgeLimit {
                break; }
        }
        if i == nFanoutsC {
            for  i = 0; i < nFanoutsC; i++  {
                Count += Gia_ObjEdgeAdd( iObj, pFanouts[i], p.vEdge1, p.vEdge2 );
                Count += Gia_ObjEdgeAdd( pFanouts[i], iObj, p.vEdge1, p.vEdge2 );
            }
            if Count != 0 {
                fmt.Printf("Wrong number of edges.\n"); }
            if fVerbose != 0 {
                fmt.Printf("Fixed %d critical fanouts\n", nFanoutsC); }
            return 1;
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Cannot fix\n"); }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Finds edge assignment.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Edg_ManAssignEdgeNew(p *Gia_Man_t, nEdges int32, fVerbose int32) int32 {
    var DelayNoEdge int32  = 1;
    var fLevelVerbose int32;
    var vLevel *Vec_Int_t;
    var vEdges *Vec_Wec_t  = Vec_WecStart(0);
    var vEdge1 *Vec_Int_t = nil;
    var vEdge2 *Vec_Int_t = nil;
    var DelayD int32;
    var DelayR int32;
    var DelayPrev int32 = ABC_INFINITY;
    var k int32;
    var j int32;
    var i int32;
    var iLast int32 = -1;
    var iObj int32;
    if fVerbose != 0 {
        fmt.Printf("Running edge assignment with E = %d.\n", nEdges); }
    // create fanouts
    Edg_ManToMapping( p );
    // create empty assignment
    Vec_IntFreeP( &p.vEdge1 );
    Vec_IntFreeP( &p.vEdge2 );
    p.vEdge1 = Vec_IntStart( Gia_ManObjNum(p) );
    p.vEdge2 = Vec_IntStart( Gia_ManObjNum(p) );
    // perform optimization
    for  i = 0; i < 10000; i++  {
        // if there is no improvement after 10 iterations, quit
        if i > iLast + 50 {
            break; }
        // create delay information
        DelayD = Edg_ManEvalEdgeDelay( p );
        DelayR = Edg_ManEvalEdgeDelayR( p );
        assert(DelayD == DelayR + DelayNoEdge, "DelayD == DelayR + DelayNoEdge");
        if DelayPrev > DelayD {
            //printf( "Saving backup point at %d levels.\n", DelayD );
            Vec_IntFreeP( &vEdge1 );  vEdge1 = Vec_IntDup( p.vEdge1 );
            Vec_IntFreeP( &vEdge2 );  vEdge2 = Vec_IntDup( p.vEdge2 );
            DelayPrev = DelayD;
            iLast = i;
        }
        if fVerbose != 0 {
            fmt.Printf("\nIter %4d : Delay = %4d\n", i, DelayD); }
        // collect critical nodes (nodes with critical edges)
        Edg_ManCollectCritEdges( p, vEdges, DelayD );
        // sort levels according to the number of critical edges
        if fLevelVerbose != 0 {
            // inlining of macro Vec_WecForEachLevel
            for  k = 0; (k < Vec_WecSize(vEdges)) && (((vLevel) = Vec_WecEntry(vEdges, k)), 1); k++  {
                Vec_IntPush( vLevel, k );
            }
        }
        Vec_WecSort( vEdges, 0 );
        if fLevelVerbose != 0 {
            // inlining of macro Vec_WecForEachLevel
            for  k = 0; (k < Vec_WecSize(vEdges)) && (((vLevel) = Vec_WecEntry(vEdges, k)), 1); k++  {
                var Level int32  = Vec_IntPop( vLevel );
                fmt.Printf("%d: Level %2d : ", k, Level);
                Vec_IntPrint( vLevel );
            }
        }
        // inlining of macro Vec_WecForEachLevel
        for  k = 0; (k < Vec_WecSize(vEdges)) && (((vLevel) = Vec_WecEntry(vEdges, k)), 1); k++  {
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(vLevel)) && (((iObj) = Vec_IntEntry(vLevel, j)), 1); j++  {
                if Edg_ObjImprove(p, iObj, nEdges, DelayD, fVerbose) != 0 { // improved
                    break; }
            }
            if j < Vec_IntSize(vLevel) {
                break; }
        }
        if k == Vec_WecSize(vEdges) { // if we could not improve anything, quit
            break; }
    }
    Vec_WecFree( vEdges );
    // update to the saved version
    Vec_IntFreeP( &p.vEdge1 );  p.vEdge1 = vEdge1;
    Vec_IntFreeP( &p.vEdge2 );  p.vEdge2 = vEdge2;
    return DelayD;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

