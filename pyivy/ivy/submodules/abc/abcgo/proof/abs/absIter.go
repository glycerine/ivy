package abs

import "fmt"

/**CFile****************************************************************

  FileName    [absIter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Iterative improvement of abstraction.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absIter.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abs.h"
//#include "sat/bmc/bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Gia_ObjIsInGla(p *Gia_Man_t, pObj *Gia_Obj_t) int32 { return Vec_IntEntry(p.vGateClasses, Gia_ObjId(p, pObj));  }
func Gia_ObjAddToGla(p *Gia_Man_t, pObj *Gia_Obj_t) { Vec_IntWriteEntry(p.vGateClasses, Gia_ObjId(p, pObj), 1); }
func Gia_ObjRemFromGla(p *Gia_Man_t, pObj *Gia_Obj_t) { Vec_IntWriteEntry(p.vGateClasses, Gia_ObjId(p, pObj), 0); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_IterTryImprove(p *Gia_Man_t, nTimeOut int32, iFrame0 int32) int32 {
    var pAbs *Gia_Man_t  = Gia_ManDupAbsGates( p, p.vGateClasses );
    var pAig *Aig_Man_t  = Gia_ManToAigSimple( pAbs );
    var nStart int32;
    var nFrames int32;
    if iFrame0 != 0 {
        nFrames = iFrame0 + 1;
    } else {
        nFrames = 10000000;
    }
    var nNodeDelta int32   =         2000;
    var nBTLimit int32;
    var nBTLimitAll int32;
    var fVerbose int32;
    var RetValue, iFrame int32; 
    RetValue = Saig_BmcPerform( pAig, nStart, nFrames, nNodeDelta, nTimeOut, nBTLimit, nBTLimitAll, fVerbose, 0, &iFrame, 1, 0 );
    assert(RetValue == 0 || RetValue == -1, "RetValue == 0 || RetValue == -1");
    Aig_ManStop( pAig );
    Gia_ManStop( pAbs );
    return iFrame;
}
func Gia_ManShrinkGla(p *Gia_Man_t, nFrameMax int32, nTimeOut int32, fUsePdr int32, fUseSat int32, fUseBdd int32, fVerbose int32) *Gia_Man_t {
    var pObj *Gia_Obj_t;
    var i, iFrame0, iFrame int32;
    var nTotal, nRemoved int32;
    var vGScopy *Vec_Int_t;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    assert(Gia_ManPoNum(p) == 1, "Gia_ManPoNum(p) == 1");
    assert(p.vGateClasses != nil, "p.vGateClasses != nil");
    vGScopy = Vec_IntDup( p.vGateClasses );
    if nFrameMax == 0 {
        iFrame0 = Gia_IterTryImprove( p, 0, 0 );
    } else {
        iFrame0 = nFrameMax - 1; }
    for {
        var fChanges int32;
        // inlining of macro Gia_ManForEachObj1
        for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            if pObj.fMark0 {
                continue; }
            if !(Gia_ObjIsInGla(p, pObj) != 0) {
                continue; }
            if pObj == Gia_ObjFanin0( Gia_ManPo(p, 0) ) {
                continue; }
            if Gia_ObjIsAnd(pObj) {
                if (Gia_ObjIsInGla(p, Gia_ObjFanin0(pObj)) != 0) && (Gia_ObjIsInGla(p, Gia_ObjFanin1(pObj)) != 0) {
                    continue; }
            }
            if Gia_ObjIsRo(p, pObj) {
                if Gia_ObjIsInGla(p, Gia_ObjFanin0(Gia_ObjRoToRi(p, pObj))) != 0 {
                    continue; }
            }        
            clk = Abc_Clock();
            fmt.Printf("%5d : ", nTotal);
            fmt.Printf("Obj =%7d   ", i);
            Gia_ObjRemFromGla( p, pObj );
            iFrame = Gia_IterTryImprove( p, nTimeOut, iFrame0 );
            if nFrameMax != 0 {
                assert(iFrame <= nFrameMax, "iFrame <= nFrameMax");
            } else {
                assert(iFrame <= iFrame0, "iFrame <= iFrame0"); }
            fmt.Printf("Frame =%6d   ", iFrame);
            if iFrame < iFrame0 {
                pObj.fMark0 = 1;
                Gia_ObjAddToGla( p, pObj );
                fmt.Printf("           ");
            
            } else {
                fChanges = 1;
                nRemoved++;
                fmt.Printf("Removing   ");
                Vec_IntWriteEntry( vGScopy, Gia_ObjId(p, pObj), 0 );
            }
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            nTotal++;
            // update the classes
            Vec_IntFreeP( &p.vGateClasses );
            p.vGateClasses = Vec_IntDup(vGScopy);
        }
        if !(fChanges != 0) {
            break; }
    }
    Gia_ManCleanMark0(p);
    Vec_IntFree( vGScopy );
    fmt.Printf("Tried = %d.  ", nTotal);
    fmt.Printf("Removed = %d. (%.2f %%)  ", nRemoved, 100.0 * nRemoved / Vec_IntCountPositive(p.vGateClasses));
    Abc_PrintTime( 1, "Time",  Abc_Clock() - clkTotal );
    return nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

