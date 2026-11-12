package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaFrames.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Timeframe unrolling.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaFrames.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Gia_ManFra_t struct {
    pPars   *Gia_ParFra_t;    // parameters
    pAig      *Gia_Man_t;     // AIG to unroll
    vIns      *Vec_Ptr_t;     // inputs of each timeframe    
    vAnds      *Vec_Ptr_t;    // nodes of each timeframe    
    vOuts      *Vec_Ptr_t;    // outputs of each timeframe    
}


   
type Gia_ManUnr_t struct {
    pPars   *Gia_ParFra_t;    // parameters
    pAig      *Gia_Man_t;     // AIG to unroll (points to pOrder)
    // internal data 
    pOrder      *Gia_Man_t;   // AIG reordered (points to pAig)
    vLimit      *Vec_Int_t;   // limits of each timeframe
    // data for each ordered node
    vRank      *Vec_Int_t;    // rank of each node
    vDegree      *Vec_Int_t;  // degree of each node
    vDegDiff      *Vec_Int_t; // degree of each node
    vFirst      *Vec_Int_t;   // first entry in the store
    vStore      *Vec_Int_t;   // store for saved data
    // the resulting AIG
    pNew      *Gia_Man_t;     // the resulting AIG
    LastLit             int32;  // the place to store the last literal
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Duplicates AIG for unrolling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollDup_rec(pNew *Gia_Man_t, pObj *Gia_Obj_t, Id int32) {
    if ~pObj.Value {
        return; }
    if Gia_ObjIsCi(pObj) {
        pObj.Value = Gia_ManAppendCi(pNew);
    } else if Gia_ObjIsCo(pObj) {
        Gia_ManUnrollDup_rec( pNew, Gia_ObjFanin0(pObj), Gia_ObjFaninId0(pObj, Id) );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    
    } else if Gia_ObjIsAnd(pObj) {
        Gia_ManUnrollDup_rec( pNew, Gia_ObjFanin0(pObj), Gia_ObjFaninId0(pObj, Id) );
        Gia_ManUnrollDup_rec( pNew, Gia_ObjFanin1(pObj), Gia_ObjFaninId1(pObj, Id) );
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    
    } else { assert(false, "0"); }
    Gia_ManObj(pNew, Abc_Lit2Var(pObj.Value)).Value = Id;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG for unrolling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollDup(p *Gia_Man_t, vLimit *Vec_Int_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(Vec_IntSize(vLimit) == 0, "Vec_IntSize(vLimit) == 0");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );

    // save constant class
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    Vec_IntPush( vLimit, Gia_ManObjNum(pNew) );
 
    // create first class
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManUnrollDup_rec( pNew, pObj, Gia_ObjId(p, pObj) );
    }
    Vec_IntPush( vLimit, Gia_ManObjNum(pNew) );

    // create next classes
    for  i = 1; i < Gia_ManObjNum(pNew); i++  {
        if i == Vec_IntEntryLast(vLimit) {
            Vec_IntPush( vLimit, Gia_ManObjNum(pNew) ); }
        pObj = Gia_ManObj( p, Gia_ManObj(pNew, i).Value );
        if Gia_ObjIsRo(p, pObj) {
            pObj = Gia_ObjRoToRi(p, pObj);
            assert(!~pObj.Value, "!~pObj.Value");
            Gia_ManUnrollDup_rec( pNew, pObj, Gia_ObjId(p, pObj) );
        }
    }
    Gia_ManSetRegNum( pNew, 0 );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Duplicates AIG for unrolling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollAbs(p *Gia_Man_t, nFrames int32) *Vec_Ptr_t {
    var fVerbose int32;
    var vFrames *Vec_Ptr_t;
    var vLimit, vOne *Vec_Int_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var nObjBits, nObjMask int32;
    var f, fMax, k, Entry, Prev, iStart, iStop, Size int32;
    // get the bitmasks
    nObjBits = Abc_Base2Log( Gia_ManObjNum(p) );
    nObjMask = (1 << nObjBits) - 1;
    assert(Gia_ManObjNum(p) <= nObjMask, "Gia_ManObjNum(p) <= nObjMask");
    // derive the tents
    vLimit = Vec_IntAlloc( 1000 );
    pNew = Gia_ManUnrollDup( p, vLimit );
    // debug printout
    if fVerbose != 0 {
        Prev = 1;
        fmt.Printf("Tents: ");
        // inlining of macro Vec_IntForEachEntryStart
        for  k = 1; (k < Vec_IntSize(vLimit)) && (((Entry) = Vec_IntEntry(vLimit, k)), 1); k++  {
            fmt.Printf("%d=%d ", k, Entry-Prev);
Prev = Entry;
        }
        fmt.Printf("  Unused=%d", Gia_ManObjNum(p) - Gia_ManObjNum(pNew));
        fmt.Printf("\n");
    }
    // create abstraction
    vFrames = Vec_PtrAlloc( Vec_IntSize(vLimit) );
    for  fMax = 0; fMax < nFrames; fMax++  {
        if fMax+1 < Vec_IntSize(vLimit) {
            Size = Vec_IntEntry(vLimit, fMax+1);
        } else {
            Size = Gia_ManObjNum(pNew);
        }
        vOne = Vec_IntAlloc( Size );
        for  f = 0; f <= fMax; f++  {
            if f   < Vec_IntSize(vLimit) {
                iStart = Vec_IntEntry(vLimit, f  );
            } else {
                iStart = 0;
            }
            if f+1 < Vec_IntSize(vLimit) {
                iStop = Vec_IntEntry(vLimit, f+1);
            } else {
                iStop = 0;
            }
            for  k = iStop - 1; k >= iStart; k--  {
                assert(Gia_ManObj(pNew, k).Value > 0, "Gia_ManObj(pNew, k).Value > 0");
                pObj  = Gia_ManObj(p, Gia_ManObj(pNew, k).Value);
                if Gia_ObjIsCo(pObj) || Gia_ObjIsPi(p, pObj) {
                    continue; }
                assert(Gia_ObjIsRo(p, pObj) || Gia_ObjIsAnd(pObj), "Gia_ObjIsRo(p, pObj) || Gia_ObjIsAnd(pObj)");
                Entry = ((fMax-f) << nObjBits) | Gia_ObjId(p, pObj);
                Vec_IntPush( vOne, Entry );
//                printf( "%d ", Gia_ManObj(pNew, k)->Value );
            }
//            printf( "\n" );
        }
        // add in reverse topological order
        Vec_IntSort( vOne, 1 );
        Vec_PtrPush( vFrames, vOne );
        assert(Vec_IntSize(vOne) <= Size - 1, "Vec_IntSize(vOne) <= Size - 1");
    }
    Vec_IntFree( vLimit );
    Gia_ManStop( pNew );
    return vFrames;
}

/**Function*************************************************************

  Synopsis    [Creates manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrStart(pAig *Gia_Man_t, pPars *Gia_ParFra_t) *Gia_ManUnr_t { 
    
    var pObj *Gia_Obj_t;
    var i, k, iRank, iFanin, Degree, Shift int32;
    var clk abctime  = Abc_Clock();

    // inlining of macro ABC_CALLOC
    p := &Gia_ManUnr_t{};
    p.pAig   = pAig;
    p.pPars  = pPars;
    // create order
    p.vLimit = Vec_IntAlloc( 0 );
    p.pOrder = Gia_ManUnrollDup( pAig, p.vLimit );
/*
    Vec_IntForEachEntryStart( p->vLimit, Shift, i, 1 )
        printf( "%d=%d ", i, Shift-Vec_IntEntry(p->vLimit, i-1) );
    printf( "\n" );
*/
    // assign rank
    p.vRank  = Vec_IntAlloc( Gia_ManObjNum(p.pOrder) );
    for  iRank = i = 0; i < Gia_ManObjNum(p.pOrder); Vec_IntPush(p.vRank, iRank), i++  {
        if Vec_IntEntry(p.vLimit, iRank) == i {
            iRank++; }
    }
    assert(iRank == Vec_IntSize(p.vLimit)-1, "iRank == Vec_IntSize(p.vLimit)-1");
    assert(Vec_IntSize(p.vRank) == Gia_ManObjNum(p.pOrder), "Vec_IntSize(p.vRank) == Gia_ManObjNum(p.pOrder)");
    // assign degree
    p.vDegree = Vec_IntStart( Gia_ManObjNum(p.pOrder) );
    p.vDegDiff= Vec_IntStart( 2* Gia_ManObjNum(p.pOrder) );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pOrder.nObjs) && ((pObj) = Gia_ManObj(p.pOrder, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        for  k = 0; k < 2; k++  {
            if k {
                iFanin = Gia_ObjFaninId1(pObj, i);
            } else {
                iFanin = Gia_ObjFaninId0(pObj, i);
            }
            Degree = Vec_IntEntry(p.vRank, i) - Vec_IntEntry(p.vRank, iFanin);
            Vec_IntWriteEntry( p.vDegDiff, 2*i + k, Degree );
            if Vec_IntEntry(p.vDegree, iFanin) < Degree {
                Vec_IntWriteEntry( p.vDegree, iFanin, Degree ); }
        }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  k = 0; (k < Vec_IntSize(p.pOrder.vCos)) && ((pObj) = Gia_ManCo(p.pOrder, k)); k++  {
        i      = Gia_ObjId( p.pOrder, pObj );
        iFanin = Gia_ObjFaninId0(pObj, i);
        Degree = Vec_IntEntry(p.vRank, i) - Vec_IntEntry(p.vRank, iFanin);
        Vec_IntWriteEntry( p.vDegDiff, 2*i, Degree );
        if Vec_IntEntry(p.vDegree, iFanin) < Degree {
            Vec_IntWriteEntry( p.vDegree, iFanin, Degree ); }
    }
    // assign first
    p.vFirst = Vec_IntAlloc( Gia_ManObjNum(p.pOrder) );
    p.vStore = Vec_IntStartFull( 2* Gia_ManObjNum(p.pOrder) + Vec_IntSum(p.vDegree) );
    for  Shift = i = 0; i < Gia_ManObjNum(p.pOrder); i++  {
        Vec_IntPush( p.vFirst, Shift );
        Vec_IntWriteEntry( p.vStore, Shift, 1 + Vec_IntEntry(p.vDegree, i) ); 
        Shift += 2 + Vec_IntEntry(p.vDegree, i);
    }
    assert(Shift == Vec_IntSize(p.vStore), "Shift == Vec_IntSize(p.vStore)");
    // cleanup
    Vec_IntFreeP( &p.vRank );
    Vec_IntFreeP( &p.vDegree );

    // print verbose output
    if pPars.fVerbose {
        fmt.Printf("Convergence = %d.  Dangling objects = %d.  Average degree = %.3f   ", Vec_IntSize(p.vLimit) - 1, Gia_ManObjNum(pAig) - Gia_ManObjNum(p.pOrder), 1.0*Vec_IntSize(p.vStore)/Gia_ManObjNum(p.pOrder) - 1.0);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollStop(pMan any) {
    var p *Gia_ManUnr_t  = (Gia_ManUnr_t *)pMan;
    Gia_ManStopP( &p.pOrder );
    Vec_IntFreeP( &p.vLimit );
    Vec_IntFreeP( &p.vRank );
    Vec_IntFreeP( &p.vDegree );
    Vec_IntFreeP( &p.vDegDiff );
    Vec_IntFreeP( &p.vFirst );
    Vec_IntFreeP( &p.vStore );
    // inlining of macro ABC_FREE
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Reading/writing entry from storage.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjUnrWrite(p *Gia_ManUnr_t, Id int32, Entry int32) {
    var i int32;
    var pArray *int32 = Vec_IntEntryP( p.vStore, Vec_IntEntry(p.vFirst, Id) );
    for  i = pArray[0]; i > 1; i--  {
        pArray[i] = pArray[i-1];
    }
    pArray[1] = Entry;
}
func Gia_ObjUnrRead(p *Gia_ManUnr_t, Id int32, Degree int32) int32 {
    var pArray *int32  = Vec_IntEntryP( p.vStore, Vec_IntEntry(p.vFirst, Id) );
    if Id == 0 {
        return 0; }
    assert(Degree >= 0 && Degree < pArray[0], "Degree >= 0 && Degree < pArray[0]");
    if Degree != 0 {
        Degree--; }
    return pArray[Degree + 1];
}
func Gia_ObjUnrReadCopy0(p *Gia_ManUnr_t, pObj *Gia_Obj_t, Id int32) int32 {
    var Lit int32  = Gia_ObjUnrRead(p, Gia_ObjFaninId0(pObj, Id), Vec_IntEntry(p.vDegDiff, 2*Id));
    return Abc_LitNotCond( Lit, Gia_ObjFaninC0(pObj) );
}
func Gia_ObjUnrReadCopy1(p *Gia_ManUnr_t, pObj *Gia_Obj_t, Id int32) int32 {
    var Lit int32  = Gia_ObjUnrRead(p, Gia_ObjFaninId1(pObj, Id), Vec_IntEntry(p.vDegDiff, 2*Id+1));
    return Abc_LitNotCond( Lit, Gia_ObjFaninC1(pObj) );
}
func Gia_ObjUnrReadCi(p *Gia_ManUnr_t, Id int32, f int32, pNew *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t  = Gia_ManObj( p.pOrder, Id );
    var pObjReal *Gia_Obj_t  = Gia_ManObj( p.pAig, pObj.Value );
    assert(Gia_ObjIsCi(pObjReal), "Gia_ObjIsCi(pObjReal)");
    if Gia_ObjIsPi(p.pAig, pObjReal) {
        if !p.pPars.fSaveLastLit { 
            pObj = Gia_ManPi( pNew, Gia_ManPiNum(p.pAig) * f + Gia_ObjCioId(pObjReal) );
        } else {
            pObj = Gia_ManPi( pNew, Gia_ManRegNum(p.pAig) + Gia_ManPiNum(p.pAig) * f + Gia_ObjCioId(pObjReal) ); }
        return Abc_Var2Lit( Gia_ObjId(pNew, pObj), 0 );
    }
    if f == 0 { // initialize!
    
        if p.pPars.fInit {
            return 0; }
        assert(Gia_ObjCioId(pObjReal) >= Gia_ManPiNum(p.pAig), "Gia_ObjCioId(pObjReal) >= Gia_ManPiNum(p.pAig)");
        if !p.pPars.fSaveLastLit { 
            pObj = Gia_ManPi( pNew, Gia_ManPiNum(p.pAig) * p.pPars.nFrames + Gia_ObjCioId(pObjReal)-Gia_ManPiNum(p.pAig) );
        } else {
            pObj = Gia_ManPi( pNew, Gia_ObjCioId(pObjReal)-Gia_ManPiNum(p.pAig) ); }
        return Abc_Var2Lit( Gia_ObjId(pNew, pObj), 0 );
    }
    pObj = Gia_ManObj( p.pOrder, Abc_Lit2Var(Gia_ObjRoToRi(p.pAig, pObjReal).Value) );
    assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
    return Gia_ObjUnrRead( p, Gia_ObjId(p.pOrder, pObj), 0 );
}

/**Function*************************************************************

  Synopsis    [Computes init/non-init unrolling without flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollStart(pAig *Gia_Man_t, pPars *Gia_ParFra_t) any {
    var p *Gia_ManUnr_t;
    var f, i int32;
    // start 
    p = Gia_ManUnrStart( pAig, pPars );
    // start timeframes
    assert(p.pNew == nil, "p.pNew == nil");
    p.pNew = Gia_ManStart( 10000 );
    p.pNew.pName = Abc_UtilStrsav( p.pAig.pName );
    p.pNew.pSpec = Abc_UtilStrsav( p.pAig.pSpec );
    Gia_ManHashAlloc( p.pNew );
    // create combinational inputs
    if !p.pPars.fSaveLastLit { // only in the case when unrolling depth is known
        for  f = 0; f < p.pPars.nFrames; f++  {
            for  i = 0; i < Gia_ManPiNum(p.pAig); i++  {
                Gia_ManAppendCi(p.pNew);
            }
        } }
    // create flop outputs
    if !p.pPars.fInit { // only in the case when initialization is not performed
        for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
            Gia_ManAppendCi(p.pNew);
        } }
    return p;
}

/**Function*************************************************************

  Synopsis    [Computes init/non-init unrolling without flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollAdd(pMan any, fMax int32) any {
    var p *Gia_ManUnr_t  = (Gia_ManUnr_t *)pMan;
    var pObj *Gia_Obj_t;
    var f, i, Lit, Beg, End int32;
    // create PIs on demand
    if p.pPars.fSaveLastLit { 
        for  i = 0; i < Gia_ManPiNum(p.pAig); i++  {
            Gia_ManAppendCi(p.pNew);
        } }
    // unroll another timeframe
    for  f = 0; f < fMax; f++  {
        if Vec_IntSize(p.vLimit) <= fMax-f {
            continue; }
        Beg = Vec_IntEntry( p.vLimit, fMax-f-1 );
        End = Vec_IntEntry( p.vLimit, fMax-f );
        for  i = Beg; i < End; i++  {
            pObj = Gia_ManObj( p.pOrder, i );
            if Gia_ObjIsAnd(pObj) {
                Lit = Gia_ManHashAnd( p.pNew, Gia_ObjUnrReadCopy0(p, pObj, i), Gia_ObjUnrReadCopy1(p, pObj, i) );
            } else if Gia_ObjIsCo(pObj) {
                Lit = Gia_ObjUnrReadCopy0(p, pObj, i);
                if f == fMax-1 {
                    if p.pPars.fSaveLastLit {
                        p.LastLit = Lit;
                    } else {
                        Gia_ManAppendCo( p.pNew, Lit ); }
                }
            
            } else if Gia_ObjIsCi(pObj) {
                Lit = Gia_ObjUnrReadCi( p, i, f, p.pNew );
            } else { assert(false, "0"); }
            assert(Lit >= 0, "Lit >= 0");
            Gia_ObjUnrWrite( p, i, Lit ); // should be exactly one call for each obj!
        }
    }
    return p.pNew;
}

/**Function*************************************************************

  Synopsis    [Read the last literal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnrollLastLit(pMan any) int32 {
    var p *Gia_ManUnr_t  = (Gia_ManUnr_t *)pMan;
    return p.LastLit;
}

/**Function*************************************************************

  Synopsis    [Computes init/non-init unrolling without flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManUnroll(pAig *Gia_Man_t, pPars *Gia_ParFra_t) *Gia_Man_t {
    var p *Gia_ManUnr_t;
    var pNew, pTemp *Gia_Man_t;
    var fMax int32;
    p = (Gia_ManUnr_t *)Gia_ManUnrollStart( pAig, pPars );
    for  fMax = 1; fMax <= p.pPars.nFrames; fMax++  {
        Gia_ManUnrollAdd( p, fMax );
    }
    assert(Gia_ManPoNum(p.pNew) == p.pPars.nFrames * Gia_ManPoNum(p.pAig), "Gia_ManPoNum(p.pNew) == p.pPars.nFrames * Gia_ManPoNum(p.pAig)");
    Gia_ManHashStop( p.pNew );
    Gia_ManSetRegNum( p.pNew, 0 );
//    Gia_ManPrintStats( pNew, 0 );
    // cleanup
    p.pNew = Gia_ManCleanup( pTemp = p.pNew );
    Gia_ManStop( pTemp );
//    Gia_ManPrintStats( pNew, 0 );
    pNew = p.pNew;  p.pNew = nil;
    Gia_ManUnrollStop( p );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Computes init/non-init unrolling without flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
Gia_Man_t * Gia_ManUnroll( Gia_ManUnr_t * p )
{
    Gia_Man_t * pNew, * pTemp;
    Gia_Obj_t * pObj;
    int fMax, f, i, Lit, Beg, End;
    // start timeframes
    pNew = Gia_ManStart( 10000 );
    pNew->pName = Abc_UtilStrsav( p->pAig->pName );
    pNew->pSpec = Abc_UtilStrsav( p->pAig->pSpec );
    Gia_ManHashAlloc( pNew );
    // create combinational inputs
    for ( f = 0; f < p->pPars->nFrames; f++ )
        for ( i = 0; i < Gia_ManPiNum(p->pAig); i++ )
            Gia_ManAppendCi(pNew);
    if ( !p->pPars->fInit )
        for ( i = 0; i < Gia_ManRegNum(p->pAig); i++ )
            Gia_ManAppendCi(pNew);
    // add nodes to each time-frame
//    Gia_ObjUnrWrite( p, 0, 0 );
    for ( fMax = 1; fMax <= p->pPars->nFrames; fMax++ )
        for ( f = 0; f < fMax; f++ )
        {
            if ( Vec_IntSize(p->vLimit) <= fMax-f )
                continue;
            Beg = Vec_IntEntry( p->vLimit, fMax-f-1 );
            End = Vec_IntEntry( p->vLimit, fMax-f );
            for ( i = Beg; i < End; i++ )
            {
                pObj = Gia_ManObj( p->pOrder, i );
                if ( Gia_ObjIsAnd(pObj) )
                    Lit = Gia_ManHashAnd( pNew, Gia_ObjUnrReadCopy0(p, pObj, i), Gia_ObjUnrReadCopy1(p, pObj, i) );
                else if ( Gia_ObjIsCo(pObj) )
                {
                    Lit = Gia_ObjUnrReadCopy0(p, pObj, i);
                    if ( f == fMax-1 )
                        Gia_ManAppendCo( pNew, Lit );
                }
                else if ( Gia_ObjIsCi(pObj) )
                    Lit = Gia_ObjUnrReadCi( p, i, f, pNew );
                else assert( 0 );
                assert( Lit >= 0 );
                Gia_ObjUnrWrite( p, i, Lit ); // should be exactly one call for each obj!
            }
        }
    assert( Gia_ManPoNum(pNew) == p->pPars->nFrames * Gia_ManPoNum(p->pAig) );
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, 0 );
//    Gia_ManPrintStats( pNew, 0 );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
//    Gia_ManPrintStats( pNew, 0 );
    return pNew;
} 
*/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFrames2(pAig *Gia_Man_t, pPars *Gia_ParFra_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var clk abctime  = Abc_Clock();
    pNew = Gia_ManUnroll( pAig, pPars );
    if pPars.fVerbose {
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
    return pNew;
}



/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraSetDefaultParams(p *Gia_ParFra_t) {
    memset( p, 0, sizeof(Gia_ParFra_t) );
    p.nFrames      =  32;    // the number of frames to unroll
    p.fInit        =   0;    // initialize the timeframes
    p.fVerbose     =   0;    // enables verbose output
}

/**Function*************************************************************

  Synopsis    [Creates manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraStart(pAig *Gia_Man_t, pPars *Gia_ParFra_t) *Gia_ManFra_t { 
    var p *Gia_ManFra_t;
    // inlining of macro ABC_ALLOC
    p = ((Gia_ManFra_t *) malloc(sizeof(Gia_ManFra_t) * (size_t)(1)));
    memset( p, 0, sizeof(Gia_ManFra_t) );
    p.pAig  = pAig;
    p.pPars = pPars;
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraStop(p *Gia_ManFra_t) {
    Vec_VecFree( (Vec_Vec_t *)p.vIns );
    Vec_VecFree( (Vec_Vec_t *)p.vAnds );
    Vec_VecFree( (Vec_Vec_t *)p.vOuts );
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Computes supports of all timeframes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFraSupports(p *Gia_ManFra_t) {
    var vIns *Vec_Int_t = nil;
    var vAnds *Vec_Int_t;
    var vOuts *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var f, i int32;
    p.vIns  = Vec_PtrStart( p.pPars.nFrames );
    p.vAnds = Vec_PtrStart( p.pPars.nFrames );
    p.vOuts = Vec_PtrStart( p.pPars.nFrames );
    Gia_ManIncrementTravId( p.pAig );
    for  f = p.pPars.nFrames - 1; f >= 0; f--  {
        vOuts = Gia_ManCollectPoIds( p.pAig );
        if vIns {
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(p.pAig, Vec_IntEntry(vIns,i))); i++  {
            if Gia_ObjIsRo(p.pAig, pObj) {
                Vec_IntPush( vOuts, Gia_ObjId( p.pAig, Gia_ObjRoToRi(p.pAig, pObj) ) ); }
        } }
        vIns = Vec_IntAlloc( 100 );
        Gia_ManCollectCis( p.pAig, Vec_IntArray(vOuts), Vec_IntSize(vOuts), vIns );
        vAnds = Vec_IntAlloc( 100 );
        Gia_ManCollectAnds( p.pAig, Vec_IntArray(vOuts), Vec_IntSize(vOuts), vAnds, nil );
        Vec_PtrWriteEntry( p.vIns,  f, vIns );
        Vec_PtrWriteEntry( p.vAnds, f, vAnds );
        Vec_PtrWriteEntry( p.vOuts, f, vOuts );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFramesInit(pAig *Gia_Man_t, pPars *Gia_ParFra_t) *Gia_Man_t {
    var fUseAllPis int32  = 1;
    var pFrames, pTemp *Gia_Man_t;
    var p *Gia_ManFra_t;
    var pObj *Gia_Obj_t;
    var vIns, vAnds, vOuts *Vec_Int_t;
    var i, f int32;
    p = Gia_ManFraStart( pAig, pPars );
    Gia_ManFraSupports( p );
    pFrames = Gia_ManStart( Vec_VecSizeSize((Vec_Vec_t*)p.vIns)+
        Vec_VecSizeSize((Vec_Vec_t*)p.vAnds)+Vec_VecSizeSize((Vec_Vec_t*)p.vOuts) );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    pFrames.pSpec = Abc_UtilStrsav( pAig.pSpec );
    Gia_ManHashAlloc( pFrames );
    Gia_ManConst0(pAig).Value = 0;
    for  f = 0; f < pPars.nFrames; f++  {
        vIns  = (Vec_Int_t *)Vec_PtrEntry( p.vIns,  f );
        vAnds = (Vec_Int_t *)Vec_PtrEntry( p.vAnds, f );
        vOuts = (Vec_Int_t *)Vec_PtrEntry( p.vOuts, f );
        if pPars.fVerbose {
            fmt.Printf("Frame %3d : CI = %6d. AND = %6d. CO = %6d.\n", f, Vec_IntSize(vIns), Vec_IntSize(vAnds), Vec_IntSize(vOuts)); }
        if fUseAllPis != 0 {
            // inlining of macro Gia_ManForEachPi
            for  i = 0; (i < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
                pObj.Value = Gia_ManAppendCi( pFrames );
            }
            if f == 0 {
                // inlining of macro Gia_ManForEachObjVec
                for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vIns,i))); i++  {
                    assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
                    if !Gia_ObjIsPi(pAig, pObj) {
                        pObj.Value = 0; }
                }
            
            } else {
                // inlining of macro Gia_ManForEachObjVec
                for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vIns,i))); i++  {
                    assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
                    if !Gia_ObjIsPi(pAig, pObj) {
                        pObj.Value = Gia_ObjRoToRi(pAig, pObj).Value; }
                }
            }
        
        } else {
            if f == 0 {
                // inlining of macro Gia_ManForEachObjVec
                for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vIns,i))); i++  {
                    assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
                    if Gia_ObjIsPi(pAig, pObj) {
                        pObj.Value = Gia_ManAppendCi( pFrames );
                    } else {
                        pObj.Value = 0; }
                }
            
            } else {
                // inlining of macro Gia_ManForEachObjVec
                for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vIns,i))); i++  {
                    assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
                    if Gia_ObjIsPi(pAig, pObj) {
                        pObj.Value = Gia_ManAppendCi( pFrames );
                    } else {
                        pObj.Value = Gia_ObjRoToRi(pAig, pObj).Value; }
                }
            }
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vAnds)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vAnds,i))); i++  {
            assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
            pObj.Value = Gia_ManHashAnd( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        }
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vOuts)) && ((pObj) = Gia_ManObj(pAig, Vec_IntEntry(vOuts,i))); i++  {
            assert(Gia_ObjIsCo(pObj), "Gia_ObjIsCo(pObj)");
            if Gia_ObjIsPo(pAig, pObj) {
                pObj.Value = Gia_ManAppendCo( pFrames, Gia_ObjFanin0Copy(pObj) );
            } else {
                pObj.Value = Gia_ObjFanin0Copy(pObj); }
        }
    }
    Gia_ManFraStop( p );
    Gia_ManHashStop( pFrames );
    if Gia_ManCombMarkUsed(pFrames) < Gia_ManAndNum(pFrames) {
        pFrames = Gia_ManDupMarked( pTemp = pFrames );
        if pPars.fVerbose {
            fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(pTemp), Gia_ManAndNum(pFrames)); }
        Gia_ManStop( pTemp );
    
    } else if pPars.fVerbose {
            fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(pFrames), Gia_ManAndNum(pFrames)); }
    return pFrames;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFrames(pAig *Gia_Man_t, pPars *Gia_ParFra_t) *Gia_Man_t {
    var pFrames, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vPoLits *Vec_Int_t  = nil;
    var i, f int32;
    assert(Gia_ManRegNum(pAig) > 0, "Gia_ManRegNum(pAig) > 0");
    assert(pPars.nFrames > 0, "pPars.nFrames > 0");
    if pPars.fInit {
        return Gia_ManFramesInit( pAig, pPars ); }
    if pPars.fOrPos {
        vPoLits = Vec_IntStart( Gia_ManPoNum(pAig) ); }
    pFrames = Gia_ManStart( pPars.nFrames * Gia_ManObjNum(pAig) );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    pFrames.pSpec = Abc_UtilStrsav( pAig.pSpec );
    if !pPars.fDisableSt {
        Gia_ManHashAlloc( pFrames ); }
    Gia_ManConst0(pAig).Value = 0;
    // create primary inputs
    for  f = 0; f < pPars.nFrames; f++  {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pFrames );
        }
    }
    // add internal nodes for each timeframe
    for  f = 0; f < pPars.nFrames; f++  {
        if f == 0 {
            // inlining of macro Gia_ManForEachRo
            for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
                pObj.Value = Gia_ManAppendCi( pFrames );
            }
        
        } else {
            // inlining of macro Gia_ManForEachRo
            for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
                pObj.Value = Gia_ObjRoToRi( pAig, pObj ).Value;
            }
        }
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
            pObj.Value = Gia_Obj2Lit( pFrames, Gia_ManPi(pFrames, f * Gia_ManPiNum(pAig) + i) );
        }
        if !pPars.fDisableSt {
            // inlining of macro Gia_ManForEachAnd
            for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
                pObj.Value = Gia_ManHashAnd( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
            }
        } else {
            // inlining of macro Gia_ManForEachAnd
            for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
                pObj.Value = Gia_ManAppendAnd2( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
            } }
        if vPoLits != nil {
            if !pPars.fDisableSt {
                // inlining of macro Gia_ManForEachPo
                for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
                    Vec_IntWriteEntry( vPoLits, i, Gia_ManHashOr(pFrames, Vec_IntEntry(vPoLits, i), Gia_ObjFanin0Copy(pObj)) );
                }
            } else {
                // inlining of macro Gia_ManForEachPo
                for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
                    Vec_IntWriteEntry( vPoLits, i, Abc_LitNot(Gia_ManAppendAnd2(pFrames, Abc_LitNot(Vec_IntEntry(vPoLits, i)), Abc_LitNot(Gia_ObjFanin0Copy(pObj)))) );
                } }
        
        } else {
            // inlining of macro Gia_ManForEachPo
            for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
                pObj.Value = Gia_ManAppendCo( pFrames, Gia_ObjFanin0Copy(pObj) );
            }
        }
        if f == pPars.nFrames - 1 {
            if vPoLits != nil {
                // inlining of macro Gia_ManForEachPo
                for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
                    pObj.Value = Gia_ManAppendCo( pFrames, Vec_IntEntry(vPoLits, i) );
                } }
            // inlining of macro Gia_ManForEachRi
            for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)); i++  {
                pObj.Value = Gia_ManAppendCo( pFrames, Gia_ObjFanin0Copy(pObj) );
            }
        
        } else {
            // inlining of macro Gia_ManForEachRi
            for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)); i++  {
                pObj.Value = Gia_ObjFanin0Copy(pObj);
            }
        }
    }
    Vec_IntFreeP( &vPoLits );
    if !pPars.fDisableSt {
        Gia_ManHashStop( pFrames ); }
    Gia_ManSetRegNum( pFrames, Gia_ManRegNum(pAig) );
    if Gia_ManCombMarkUsed(pFrames) < Gia_ManAndNum(pFrames) {
        pFrames = Gia_ManDupMarked( pTemp = pFrames );
        if pPars.fVerbose {
            fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(pTemp), Gia_ManAndNum(pFrames)); }
        Gia_ManStop( pTemp );
    
    } else if pPars.fVerbose {
            fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(pFrames), Gia_ManAndNum(pFrames)); }
    return pFrames;
}


/**Function*************************************************************

  Synopsis    [Perform init unrolling as long as PO(s) are constant 0.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManFramesInitSpecial(pAig *Gia_Man_t, nFrames int32, fVerbose int32) *Gia_Man_t {
    var pFrames, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    assert(Gia_ManRegNum(pAig) > 0, "Gia_ManRegNum(pAig) > 0");
    if nFrames > 0 {
        fmt.Printf("Computing specialized unrolling with %d frames...\n", nFrames); }
    pFrames = Gia_ManStart( Gia_ManObjNum(pAig) );
    pFrames.pName = Abc_UtilStrsav( pAig.pName );
    pFrames.pSpec = Abc_UtilStrsav( pAig.pSpec );
    Gia_ManHashAlloc( pFrames );
    Gia_ManConst0(pAig).Value = 0;
    for  f = 0; nFrames == 0 || f < nFrames; f++  {
        if (fVerbose != 0) && (f % 100 == 0) {
            fmt.Printf("%6d : ", f);
            Gia_ManPrintStats( pFrames, nil );
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCi(pAig, Gia_ManPiNum(pAig)+i)); i++  {
            if f {
                pObj.Value = Gia_ObjRoToRi( pAig, pObj ).Value;
            } else {
                pObj.Value = 0;
            }
        }
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(pAig)) && ((pObj) = Gia_ManCi(pAig, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pFrames );
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < pAig.nObjs) && ((pObj) = Gia_ManObj(pAig, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(pAig)) && ((pObj) = Gia_ManCo(pAig, i)); i++  {
            if Gia_ObjFanin0Copy(pObj) != 0 {
                break; }
        }
        if i < Gia_ManPoNum(pAig) {
            break; }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)); i++  {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Computed prefix of %d frames.\n", f); }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(pAig)) && ((pObj) = Gia_ManCo(pAig, Gia_ManPoNum(pAig)+i)); i++  {
        Gia_ManAppendCo( pFrames, pObj.Value );
    }
    Gia_ManHashStop( pFrames );
    pFrames = Gia_ManCleanup( pTemp = pFrames );
    if fVerbose != 0 {
        fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(pTemp), Gia_ManAndNum(pFrames)); }
    Gia_ManStop( pTemp );
    return pFrames;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

