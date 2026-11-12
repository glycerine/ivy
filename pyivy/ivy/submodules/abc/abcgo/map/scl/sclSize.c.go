package scl

import (
	"fmt"
	"math"
	"os"
)

/**CFile****************************************************************

  FileName    [sclSize.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Core timing analysis used in gate-sizing.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclSize.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclSize.h"
//#include "map/mio/mio.h"
//#include "misc/vec/vecWec.h"
//#include "base/main/main.h"

#ifdef WIN32
#include <windows.h>
#endif

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Finding most critical objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindCriticalCo(p *SC_Man, pfRise *int32) *Abc_Obj_t {
    var pObj *Abc_Obj_t;
    var pPivot *Abc_Obj_t = nil;
    var fMaxArr float32;
    var i int32;
    assert(Abc_NtkPoNum(p.pNtk) > 0, "Abc_NtkPoNum(p.pNtk) > 0");
    Abc_NtkForEachCo( p.pNtk, pObj, i )
    {
        var pArr *SC_Pair  = Abc_SclObjTime( p, pObj );
        if fMaxArr < pArr.rise {  fMaxArr = pArr.rise;
*pfRise = 1;
pPivot = pObj; }
        if fMaxArr < pArr.fall {  fMaxArr = pArr.fall;
*pfRise = 0;
pPivot = pObj; }
    }
    if fMaxArr == 0 {
        pPivot = Abc_NtkPo(p.pNtk, 0); }
    assert(pPivot != nil, "pPivot != nil");
    return pPivot;
}
// assumes that slacks are not available (uses arrival times)
func Abc_SclFindMostCriticalFanin2(p *SC_Man, pfRise *int32, pNode *Abc_Obj_t) *Abc_Obj_t {
    var pFanin *Abc_Obj_t;
    var pPivot *Abc_Obj_t = nil;
    var fMaxArr float32;
    var i int32;
    Abc_ObjForEachFanin( pNode, pFanin, i )
    {
        var pArr *SC_Pair  = Abc_SclObjTime( p, pFanin );
        if fMaxArr < pArr.rise {  fMaxArr = pArr.rise;
*pfRise = 1;
pPivot = pFanin; }
        if fMaxArr < pArr.fall {  fMaxArr = pArr.fall;
*pfRise = 0;
pPivot = pFanin; }
    }
    return pPivot;
}
// assumes that slack are available
func Abc_SclFindMostCriticalFanin(p *SC_Man, pfRise *int32, pNode *Abc_Obj_t) *Abc_Obj_t {
    var pFanin *Abc_Obj_t;
    var pPivot *Abc_Obj_t = nil;
    var fMinSlack float32  = ABC_INFINITY;
    var pArr *SC_Pair;
    var i int32;
    *pfRise = 0;
    // find min-slack node
    Abc_ObjForEachFanin( pNode, pFanin, i )
        if fMinSlack > Abc_SclObjGetSlack( p, pFanin, p.MaxDelay0 ) {
            fMinSlack = Abc_SclObjGetSlack( p, pFanin, p.MaxDelay0 );
            pPivot = pFanin;
        }
    if pPivot == nil {
        return nil; }
    // find its leading phase
    pArr = Abc_SclObjTime( p, pPivot );
    *pfRise = (pArr.rise >= pArr.fall);
    return pPivot;
}

/**Function*************************************************************

  Synopsis    [Printing timing information for the node/network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclTimeNodePrint(p *SC_Man, pObj *Abc_Obj_t, fRise int32, Length int32, maxDelay float32) {
    var pCell *SC_Cell;
    if Abc_ObjIsNode(pObj) {
        pCell = Abc_SclObjCell(pObj);
    } else {
        pCell = nil;
    }
    fmt.Printf("%8d : ", Abc_ObjId(pObj));
    fmt.Printf("%d ", Abc_ObjFaninNum(pObj));
    fmt.Printf("%4d ", Abc_ObjFanoutNum(pObj));
    fmt.Printf("%-*s ", Length, pCell ? pCell.pName : "pi");
    fmt.Printf("A =%7.2f  ", pCell ? pCell.area : 0.0);
    fmt.Printf("D%s =", fRise ? "r" : "f");
    fmt.Printf("%6.1f", Abc_SclObjTimeMax(p, pObj));
    fmt.Printf("%7.1f ps  ", -Abc_AbsFloat(Abc_SclObjTimeOne(p, pObj, 0) - Abc_SclObjTimeOne(p, pObj, 1)));
    fmt.Printf("S =%6.1f ps  ", Abc_SclObjSlewMax(p, pObj));
    fmt.Printf("Cin =%5.1f ff  ", pCell ? SC_CellPinCapAve(pCell) : 0.0);
    fmt.Printf("Cout =%6.1f ff  ", Abc_SclObjLoadMax(p, pObj));
    fmt.Printf("Cmax =%6.1f ff  ", pCell ? SC_CellPin(pCell, pCell.n_inputs).max_out_cap : 0.0);
    fmt.Printf("G =%5d  ", pCell ? int32((100.0 * Abc_SclObjLoadAve(p, pObj) / SC_CellPinCapAve(pCell))) : 0);
//    printf( "SL =%6.1f ps",     Abc_SclObjSlackMax(p, pObj, p->MaxDelay0) );
    fmt.Printf("\n");
}
func Abc_SclTimeNtkPrint(p *SC_Man, fShowAll int32, fPrintPath int32) {
    var fReversePath int32  = 1;
    var i, nLength, fRise int32;
    var pObj *Abc_Obj_t;
    var pPivot *Abc_Obj_t = Abc_SclFindCriticalCo( p, &fRise ); 
    var maxDelay float32  = Abc_SclObjTimeOne( p, pPivot, fRise );
    p.ReportDelay = maxDelay;

#ifdef WIN32
    printf( "WireLoad = \"%s\"  ", p->pWLoadUsed ? p->pWLoadUsed->pName : "none" );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 14 ); // yellow
    printf( "Gates =%7d ",         Abc_NtkNodeNum(p->pNtk) );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 7 );  // normal
    printf( "(%5.1f %%)   ",       100.0 * Abc_SclGetBufInvCount(p->pNtk) / Abc_NtkNodeNum(p->pNtk) );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 10 ); // green
    printf( "Cap =%5.1f ff ",      p->EstLoadAve );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 7 );  // normal
    printf( "(%5.1f %%)   ",       Abc_SclGetAverageSize(p->pNtk) );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 11 ); // blue
    printf( "Area =%12.2f ",       Abc_SclGetTotalArea(p->pNtk) );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 7 );  // normal
    printf( "(%5.1f %%)   ",       100.0 * Abc_SclCountMinSize(p->pLib, p->pNtk, 0) / Abc_NtkNodeNum(p->pNtk) );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 13 ); // magenta
    printf( "Delay =%9.2f ps  ",   maxDelay );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 7 );  // normal
    printf( "(%5.1f %%)   ",       100.0 * Abc_SclCountNearCriticalNodes(p) / Abc_NtkNodeNum(p->pNtk) );
    printf( "            \n" );
#else
    Abc_Print( 1, "WireLoad = \"%s\"  ",   p->pWLoadUsed ? p->pWLoadUsed->pName : "none" );
    Abc_Print( 1, "%sGates =%7d%s ",       "\033[1;33m", Abc_NtkNodeNum(p->pNtk),      "\033[0m" ); // yellow
    Abc_Print( 1, "(%5.1f %%)   ",         100.0 * Abc_SclGetBufInvCount(p->pNtk) / Abc_NtkNodeNum(p->pNtk) );
    Abc_Print( 1, "%sCap =%5.1f ff%s ",    "\033[1;32m", p->EstLoadAve,                "\033[0m" ); // green
    Abc_Print( 1, "(%5.1f %%)   ",         Abc_SclGetAverageSize(p->pNtk) );
    Abc_Print( 1, "%sArea =%12.2f%s ",     "\033[1;36m", Abc_SclGetTotalArea(p->pNtk), "\033[0m" ); // blue
    Abc_Print( 1, "(%5.1f %%)   ",         100.0 * Abc_SclCountMinSize(p->pLib, p->pNtk, 0) / Abc_NtkNodeNum(p->pNtk) );
    Abc_Print( 1, "%sDelay =%9.2f ps%s  ", "\033[1;35m", maxDelay,                     "\033[0m" ); // magenta
    Abc_Print( 1, "(%5.1f %%)   ",         100.0 * Abc_SclCountNearCriticalNodes(p) / Abc_NtkNodeNum(p->pNtk) );
    Abc_Print( 1, "            \n" );
#endif

    if fShowAll != 0 {
//        printf( "Timing information for all nodes: \n" );
        // find the longest cell name
        Abc_NtkForEachNodeReverse( p.pNtk, pObj, i )
            if ( Abc_ObjFaninNum(pObj) > 0 )
                nLength = Abc_MaxInt( nLength, len(Abc_SclObjCell(pObj).pName) );
        // print timing
        Abc_NtkForEachNodeReverse( p.pNtk, pObj, i )
            if ( Abc_ObjFaninNum(pObj) > 0 )
                Abc_SclTimeNodePrint( p, pObj, -1, nLength, maxDelay );
    }
    if fPrintPath != 0 {
        var pTemp *Abc_Obj_t;
        var pPrev *Abc_Obj_t = nil;
        var iStart int32 = -1;
        var iEnd int32 = -1;
        var vPath *Vec_Ptr_t;
//        printf( "Critical path: \n" );        
        // find the longest cell name
        pObj = Abc_ObjFanin0(pPivot);
        i = 0;
        for  (pObj != nil) && Abc_ObjIsNode(pObj)  {
            i++;
            nLength = Abc_MaxInt( nLength, len(Abc_SclObjCell(pObj).pName) );
            pObj = Abc_SclFindMostCriticalFanin( p, &fRise, pObj );
        }

        // print timing
        if !(fReversePath != 0) {
            // print timing
            pObj = Abc_ObjFanin0(pPivot);
            for  pObj != nil  {//&& Abc_ObjIsNode(pObj) )
            
                fmt.Printf("Path%3d --", i--);
                Abc_SclTimeNodePrint( p, pObj, fRise, nLength, maxDelay );
                pPrev = pObj;
                pObj = Abc_SclFindMostCriticalFanin( p, &fRise, pObj );
            }
        
        } else {
            // collect path nodes
            vPath = Vec_PtrAlloc( 100 );
            Vec_PtrPush( vPath, pPivot );
            pObj = Abc_ObjFanin0(pPivot);
            for  pObj != nil  {//&& Abc_ObjIsNode(pObj) )
            
                Vec_PtrPush( vPath, pObj );
                pPrev = pObj;
                pObj = Abc_SclFindMostCriticalFanin( p, &fRise, pObj );
            }
            Vec_PtrForEachEntryReverse( Abc_Obj_t *, vPath, pObj, i )
            {
                fmt.Printf("Path%3d --", Vec_PtrSize(vPath)-1-i);
                Abc_SclTimeNodePrint( p, pObj, fRise, nLength, maxDelay );
                if i == 1 {
                    break; }
            }
            Vec_PtrFree( vPath );
        }
        // print start-point and end-point
        Abc_NtkForEachPi( p.pNtk, pTemp, iStart )
            if ( pTemp == pPrev )
                break;
        Abc_NtkForEachPo( p.pNtk, pTemp, iEnd )
            if ( pTemp == pPivot )
                break;
        fmt.Printf("Start-point = pi%0*d.  End-point = po%0*d.\n", Abc_Base10Log( Abc_NtkPiNum(p.pNtk) ), iStart, Abc_Base10Log( Abc_NtkPoNum(p.pNtk) ), iEnd);
    }
}

/**Function*************************************************************

  Synopsis    [Timing computation for pin/gate/cone/network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclTimeFanin(p *SC_Man, pTime *SC_Timing, pObj *Abc_Obj_t, pFanin *Abc_Obj_t) {
    var pArrIn *SC_Pair    = Abc_SclObjTime( p, pFanin );
    var pSlewIn *SC_Pair   = Abc_SclObjSlew( p, pFanin );
    var pLoad *SC_Pair     = Abc_SclObjLoad( p, pObj );
    var pArrOut *SC_Pair   = Abc_SclObjTime( p, pObj );   // modified
    var pSlewOut *SC_Pair  = Abc_SclObjSlew( p, pObj );   // modified
    Scl_LibPinArrival( pTime, pArrIn, pSlewIn, pLoad, pArrOut, pSlewOut );
}
func Abc_SclDeptFanin(p *SC_Man, pTime *SC_Timing, pObj *Abc_Obj_t, pFanin *Abc_Obj_t) {
    var pDepIn *SC_Pair    = Abc_SclObjDept( p, pFanin );   // modified
    var pSlewIn *SC_Pair   = Abc_SclObjSlew( p, pFanin );
    var pLoad *SC_Pair     = Abc_SclObjLoad( p, pObj );
    var pDepOut *SC_Pair   = Abc_SclObjDept( p, pObj );
    Scl_LibPinDeparture( pTime, pDepIn, pSlewIn, pLoad, pDepOut );
}
func Abc_SclDeptObj(p *SC_Man, pObj *Abc_Obj_t) {
    var pTime *SC_Timing;
    var pFanout *Abc_Obj_t;
    var i int32;
    SC_PairClean( Abc_SclObjDept(p, pObj) );
    Abc_ObjForEachFanout( pObj, pFanout, i )
    {
        if Abc_ObjIsCo(pFanout) || Abc_ObjIsLatch(pFanout) {
            continue; }
        pTime = Scl_CellPinTime( Abc_SclObjCell(pFanout), Abc_NodeFindFanin(pFanout, pObj) );
        Abc_SclDeptFanin( p, pTime, pFanout, pObj );
    }
}
func Abc_SclObjLoadValue(p *SC_Man, pObj *Abc_Obj_t) float32 {
//    float Value = Abc_MaxFloat(pLoad->fall, pLoad->rise) / (p->EstLoadAve * p->EstLoadMax);
    return (0.5 * Abc_SclObjLoad(p, pObj).fall + 0.5 * Abc_SclObjLoad(p, pObj).rise) / (p.EstLoadAve * p.EstLoadMax);
}
func Abc_SclTimeCi(p *SC_Man, pObj *Abc_Obj_t) {
    if p.pPiDrive != nil {
        var pLoad *SC_Pair  = Abc_SclObjLoad( p, pObj );
        var pTime *SC_Pair  = Abc_SclObjTime( p, pObj );
        var pSlew *SC_Pair  = Abc_SclObjSlew( p, pObj );
        Scl_LibHandleInputDriver( p.pPiDrive, pLoad, pTime, pSlew );
    }
}
func Abc_SclTimeNode(p *SC_Man, pObj *Abc_Obj_t, fDept int32) {
    var pTime *SC_Timing;
    var pCell *SC_Cell;
    var k int32;
    var pLoad *SC_Pair  = Abc_SclObjLoad( p, pObj );
    var LoadRise float32  = pLoad.rise;
    var LoadFall float32  = pLoad.fall;
    var DeptRise float32;
    var DeptFall float32;
    var Value float32;
    if p.EstLoadMax {
        Value = Abc_SclObjLoadValue( p, pObj );
    } else {
        Value = 0;
    }
    var pFanin *Abc_Obj_t;
    if Abc_ObjIsCi(pObj) {
        assert(!(fDept != 0), "!fDept");
        Abc_SclTimeCi( p, pObj );
        return;
    }
    if Abc_ObjIsCo(pObj) {
        if !(fDept != 0) {
            Abc_SclObjDupFanin( p, pObj );
            Vec_FltWriteEntry( p.vTimesOut, pObj.iData, Abc_SclObjTimeMax(p, pObj) );
            Vec_QueUpdate( p.vQue, pObj.iData );
        }
        return;
    }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
//    if ( !(Abc_ObjFaninNum(pObj) == 1 && Abc_ObjIsPi(Abc_ObjFanin0(pObj))) && p->EstLoadMax && Value > 1 )
    if p.EstLoadMax && Value > 1 {
        pLoad.rise = p.EstLoadAve * p.EstLoadMax;
        pLoad.fall = p.EstLoadAve * p.EstLoadMax;
        if fDept != 0 {
            var pDepOut *SC_Pair   = Abc_SclObjDept( p, pObj );
            var EstDelta float32  = p.EstLinear * math.Log(float64(Value));
            DeptRise = pDepOut.rise;
            DeptFall = pDepOut.fall;
            pDepOut.rise += EstDelta;
            pDepOut.fall += EstDelta;
        }
        p.nEstNodes++;
    }
    // get the library cell
    pCell = Abc_SclObjCell( pObj );
    // compute for each fanin
    Abc_ObjForEachFanin( pObj, pFanin, k )
    {
        pTime = Scl_CellPinTime( pCell, k );
        if fDept != 0 {
            Abc_SclDeptFanin( p, pTime, pObj, pFanin );
        } else {
            Abc_SclTimeFanin( p, pTime, pObj, pFanin ); }
    }
    if p.EstLoadMax && Value > 1 {
        pLoad.rise = LoadRise;
        pLoad.fall = LoadFall;
        if fDept != 0 {
            var pDepOut *SC_Pair   = Abc_SclObjDept( p, pObj );
            pDepOut.rise = DeptRise;
            pDepOut.fall = DeptFall;
        
        } else {
            var pArrOut *SC_Pair   = Abc_SclObjTime( p, pObj );
            var EstDelta float32  = p.EstLinear * math.Log(float64(Value));
            pArrOut.rise += EstDelta;
            pArrOut.fall += EstDelta;
        }
    }
}
func Abc_SclTimeCone(p *SC_Man, vCone *Vec_Int_t) {
    var fVerbose int32;
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_SclConeClean( p, vCone );
    Abc_NtkForEachObjVec( vCone, p.pNtk, pObj, i )
    {
        if (fVerbose != 0) && Abc_ObjIsNode(pObj) {
        fmt.Printf("  Updating node %d with gate %s\n", Abc_ObjId(pObj), Abc_SclObjCell(pObj).pName); }
        if (fVerbose != 0) && Abc_ObjIsNode(pObj) {
        fmt.Printf("    before (%6.1f ps  %6.1f ps)   ", Abc_SclObjTimeOne(p, pObj, 1), Abc_SclObjTimeOne(p, pObj, 0)); }
        Abc_SclTimeNode( p, pObj, 0 );
        if (fVerbose != 0) && Abc_ObjIsNode(pObj) {
        fmt.Printf("after (%6.1f ps  %6.1f ps)\n", Abc_SclObjTimeOne(p, pObj, 1), Abc_SclObjTimeOne(p, pObj, 0)); }
    }
}
func Abc_SclTimeNtkRecompute(p *SC_Man, pArea *float32, pDelay *float32, fReverse int32, DUser float32) {
    var pObj *Abc_Obj_t;
    var D float32;
    var i int32;
    Abc_SclComputeLoad( p );
    Abc_SclManCleanTime( p );
    p.nEstNodes = 0;
    Abc_NtkForEachCi( p.pNtk, pObj, i )
        Abc_SclTimeNode( p, pObj, 0 );
    Abc_NtkForEachNode1( p.pNtk, pObj, i )
        Abc_SclTimeNode( p, pObj, 0 );
    Abc_NtkForEachCo( p.pNtk, pObj, i )
        Abc_SclTimeNode( p, pObj, 0 );
    D = Abc_SclReadMaxDelay( p );
    if (fReverse != 0) && DUser > 0 && D < DUser {
        D = DUser; }
    if pArea != nil {
        *pArea = Abc_SclGetTotalArea(p.pNtk); }
    if pDelay != nil {
        *pDelay = D; }
    if fReverse != 0 {
        p.nEstNodes = 0;
        Abc_NtkForEachNodeReverse1( p.pNtk, pObj, i )
            Abc_SclTimeNode( p, pObj, 1 );
    }
}

/**Function*************************************************************

  Synopsis    [Incremental timing update.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclTimeIncUpdateClean(p *SC_Man) {
    var vLevel *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, k int32;
    Vec_WecForEachLevel( p.vLevels, vLevel, i )
    {
        Abc_NtkForEachObjVec( vLevel, p.pNtk, pObj, k )
        {
            assert(pObj.fMarkC == 1, "pObj.fMarkC == 1");
            pObj.fMarkC = 0;
        }
        Vec_IntClear( vLevel );
    }
}
func Abc_SclTimeIncAddNode(p *SC_Man, pObj *Abc_Obj_t) {
    assert(!Abc_ObjIsLatch(pObj), "!Abc_ObjIsLatch(pObj)");
    assert(pObj.fMarkC == 0, "pObj.fMarkC == 0");
    pObj.fMarkC = 1;
    Vec_IntPush( Vec_WecEntry(p.vLevels, Abc_ObjLevel(pObj)), Abc_ObjId(pObj) );
    p.nIncUpdates++;
}
func Abc_SclTimeIncAddFanins(p *SC_Man, pObj *Abc_Obj_t) {
    var pFanin *Abc_Obj_t;
    var i int32;
    Abc_ObjForEachFanin( pObj, pFanin, i )
//        if ( !pFanin->fMarkC && Abc_ObjIsNode(pFanin) )
        if ( !pFanin.fMarkC && !Abc_ObjIsLatch(pFanin) )
            Abc_SclTimeIncAddNode( p, pFanin );
}
func Abc_SclTimeIncAddFanouts(p *SC_Man, pObj *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    Abc_ObjForEachFanout( pObj, pFanout, i )
        if ( !pFanout.fMarkC && !Abc_ObjIsLatch(pFanout) )
            Abc_SclTimeIncAddNode( p, pFanout );
}
func Abc_SclTimeIncUpdateArrival(p *SC_Man) {
    var vLevel *Vec_Int_t;
    var ArrOut, SlewOut SC_Pair;
    var pArrOut, pSlewOut *SC_Pair;
    var pObj *Abc_Obj_t;
    var E float32  = float32(0.1);
    var i, k int32;
    Vec_WecForEachLevel( p.vLevels, vLevel, i )
    {
        Abc_NtkForEachObjVec( vLevel, p.pNtk, pObj, k )
        {
            if Abc_ObjIsCo(pObj) {
                Abc_SclObjDupFanin( p, pObj );
                Vec_FltWriteEntry( p.vTimesOut, pObj.iData, Abc_SclObjTimeMax(p, pObj) );
                Vec_QueUpdate( p.vQue, pObj.iData );
                continue;
            }
            pArrOut  = Abc_SclObjTime( p, pObj );
            pSlewOut = Abc_SclObjSlew( p, pObj );
            SC_PairMove( &ArrOut,  pArrOut  );
            SC_PairMove( &SlewOut, pSlewOut );
            Abc_SclTimeNode( p, pObj, 0 );
//            if ( !SC_PairEqual(&ArrOut, pArrOut) || !SC_PairEqual(&SlewOut, pSlewOut) )
            if !SC_PairEqualE(&ArrOut, pArrOut, E) || !SC_PairEqualE(&SlewOut, pSlewOut, E) {
                Abc_SclTimeIncAddFanouts( p, pObj ); }
        }
    }
    p.MaxDelay = Abc_SclReadMaxDelay( p );
}
func Abc_SclTimeIncUpdateDeparture(p *SC_Man) {
    var vLevel *Vec_Int_t;
    var DepOut SC_Pair;
    var pDepOut *SC_Pair;
    var pObj *Abc_Obj_t;
    var E float32  = float32(0.1);
    var i, k int32;
    Vec_WecForEachLevelReverse( p.vLevels, vLevel, i )
    {
        Abc_NtkForEachObjVec( vLevel, p.pNtk, pObj, k )
        {
            pDepOut = Abc_SclObjDept( p, pObj );
            SC_PairMove( &DepOut, pDepOut );
            Abc_SclDeptObj( p, pObj );
//            if ( !SC_PairEqual(&DepOut, pDepOut) )
            if !SC_PairEqualE(&DepOut, pDepOut, E) {
                Abc_SclTimeIncAddFanins( p, pObj ); }
        }
    } 
    p.MaxDelay = Abc_SclReadMaxDelay( p );
}
func Abc_SclTimeIncCheckLevel(pNtk *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_NtkForEachObj( pNtk, pObj, i )
        if ( int32(pObj.Level) != Abc_ObjLevelNew(pObj) )
            fmt.Printf("Level of node %d is out of date!\n", i);
}
func Abc_SclTimeIncUpdate(p *SC_Man) int32 {
    var pObj *Abc_Obj_t;
    var i, RetValue int32;
    if Vec_IntSize(p.vChanged) == 0 {
        return 0; }
//    Abc_SclTimeIncCheckLevel( p->pNtk );
    Abc_NtkForEachObjVec( p.vChanged, p.pNtk, pObj, i )
    {
        Abc_SclTimeIncAddFanins( p, pObj );
        if pObj.fMarkC {
            continue; }
        Abc_SclTimeIncAddNode( p, pObj );
    }
    Vec_IntClear( p.vChanged );
    Abc_SclTimeIncUpdateArrival( p );
    Abc_SclTimeIncUpdateDeparture( p );
    Abc_SclTimeIncUpdateClean( p );
    RetValue = p.nIncUpdates;
    p.nIncUpdates = 0;
    return RetValue;
}
func Abc_SclTimeIncInsert(p *SC_Man, pObj *Abc_Obj_t) {
    Vec_IntPush( p.vChanged, Abc_ObjId(pObj) );
}
func Abc_SclTimeIncUpdateLevel_rec(pObj *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    var LevelNew int32 = Abc_ObjLevelNew(pObj);
    if LevelNew == int32(pObj.Level) && Abc_ObjIsNode(pObj) && Abc_ObjFaninNum(pObj) > 0 {
        return; }
    pObj.Level = LevelNew;
    Abc_ObjForEachFanout( pObj, pFanout, i )
        Abc_SclTimeIncUpdateLevel_rec( pFanout );
}
func Abc_SclTimeIncUpdateLevel(pObj *Abc_Obj_t) {
    Abc_SclTimeIncUpdateLevel_rec( pObj );
}



/**Function*************************************************************

  Synopsis    [Read input slew and output load.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclManReadSlewAndLoad(p *SC_Man, pNtk *Abc_Ntk_t) {
    if Abc_FrameReadMaxLoad() {
        var pObj *Abc_Obj_t;  var i int32;
        var MaxLoad float32  = Abc_FrameReadMaxLoad();
//        printf( "Default output load is specified (%.2f ff).\n", MaxLoad );
        Abc_NtkForEachPo( pNtk, pObj, i )
        {
            var pLoad *SC_Pair  = Abc_SclObjLoad( p, pObj );
            pLoad.rise = pLoad.fall = MaxLoad;
        }
    }
    if Abc_FrameReadDrivingCell() {
        var iCell int32  = Abc_SclCellFind( p.pLib, Abc_FrameReadDrivingCell() );
        if iCell == -1 {
            fmt.Printf("Cannot find the default PI driving cell (%s) in the library.\n", Abc_FrameReadDrivingCell());
        } else {
//            printf( "Default PI driving cell is specified (%s).\n", Abc_FrameReadDrivingCell() );
            p.pPiDrive = SC_LibCell( p.pLib, iCell );
            assert(p.pPiDrive != nil, "p.pPiDrive != nil");
            assert(p.pPiDrive.n_inputs == 1, "p.pPiDrive.n_inputs == 1");
        }
    }
}
 
/**Function*************************************************************

  Synopsis    [Prepare timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclManStart(pLib *SC_Lib, pNtk *Abc_Ntk_t, fUseWireLoads int32, fDept int32, DUser float32, nTreeCRatio int32) *SC_Man {
    var p *SC_Man  = Abc_SclManAlloc( pLib, pNtk );
    if nTreeCRatio != 0 {
        p.EstLoadMax = 0.01 * nTreeCRatio;  // max ratio of Cout/Cave when the estimation is used
        p.EstLinear  = 100;                  // linear coefficient
    }
    Abc_SclMioGates2SclGates( pLib, pNtk );
    Abc_SclManReadSlewAndLoad( p, pNtk );
    if fUseWireLoads != 0 {
        if pNtk.pWLoadUsed == nil {            
            p.pWLoadUsed = Abc_SclFindWireLoadModel( pLib, Abc_SclGetTotalArea(p.pNtk) );
            if p.pWLoadUsed {
            pNtk.pWLoadUsed = Abc_UtilStrsav( p.pWLoadUsed.pName ); }
        
        } else {
            p.pWLoadUsed = Abc_SclFetchWireLoadModel( pLib, pNtk.pWLoadUsed ); }
    }
    Abc_SclTimeNtkRecompute( p, &p.SumArea0, &p.MaxDelay0, fDept, DUser );
    p.SumArea  = p.SumArea0;
    p.MaxDelay = p.MaxDelay0;
    return p;
}

/**Function*************************************************************

  Synopsis    [Printing out timing information for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclTimePerformInt(pLib *SC_Lib, pNtk *Abc_Ntk_t, nTreeCRatio int32, fUseWireLoads int32, fShowAll int32, fPrintPath int32, fDumpStats int32) {
    var p *SC_Man;
    p = Abc_SclManStart( pLib, pNtk, fUseWireLoads, 1, 0, nTreeCRatio );
    Abc_SclTimeNtkPrint( p, fShowAll, fPrintPath );
    if fDumpStats != 0 {
        Abc_SclDumpStats( p, "stats.txt", 0 ); }
    Abc_SclManFree( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclTimePerform(pLib *SC_Lib, pNtk *Abc_Ntk_t, nTreeCRatio int32, fUseWireLoads int32, fShowAll int32, fPrintPath int32, fDumpStats int32) {
    var pNtkNew *Abc_Ntk_t  = pNtk;
    if pNtk.nBarBufs2 > 0 {
        pNtkNew = Abc_NtkDupDfsNoBarBufs( pNtk ); }
    Abc_SclTimePerformInt( pLib, pNtkNew, nTreeCRatio, fUseWireLoads, fShowAll, fPrintPath, fDumpStats );
    if pNtk.nBarBufs2 > 0 {
        Abc_NtkDelete( pNtkNew ); }
}



/**Function*************************************************************

  Synopsis    [Printing out fanin information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclCheckCommonInputs(pObj *Abc_Obj_t, pFanin *Abc_Obj_t) int32 {
    var pTemp *Abc_Obj_t;
    var i int32;
    Abc_ObjForEachFanin( pObj, pTemp, i )
        if Abc_NodeFindFanin( pFanin, pTemp ) >= 0 {
            fmt.Printf("Node %d and its fanin %d have common fanin %d.\n", Abc_ObjId(pObj), Abc_ObjId(pFanin), Abc_ObjId(pTemp));

            fmt.Printf("%-16s : ", Mio_GateReadName((Mio_Gate_t *)pObj.pData)); 
            Abc_ObjPrint( os.Stdout, pObj );

            fmt.Printf("%-16s : ", Mio_GateReadName((Mio_Gate_t *)pFanin.pData)); 
            Abc_ObjPrint( os.Stdout, pFanin );

            if pTemp.pData {
            fmt.Printf("%-16s : ", Mio_GateReadName((Mio_Gate_t *)pTemp.pData)); } 
            Abc_ObjPrint( os.Stdout, pTemp );
            return 1;
        }
    return 0;
}
func Abc_SclPrintFaninPairs(p *SC_Man, pNtk *Abc_Ntk_t) {
    var pObj, pFanin *Abc_Obj_t;
    var i, k int32;
    Abc_NtkForEachNode( pNtk, pObj, i )
        Abc_ObjForEachFanin( pObj, pFanin, k )
            if ( Abc_ObjIsNode(pFanin) && Abc_ObjFanoutNum(pFanin) == 1 )
                Abc_SclCheckCommonInputs( pObj, pFanin );
}

/**Function*************************************************************

  Synopsis    [Printing out buffer information.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ObjIsBuffer(pObj *Abc_Obj_t) int32 { return Abc_ObjIsNode(pObj) && Abc_ObjFaninNum(pObj) == 1; }
func Abc_SclHasBufferFanout(pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i int32;
    Abc_ObjForEachFanout( pObj, pFanout, i )
        if ( Abc_ObjIsBuffer(pFanout) )
            return 1;
    return 0;
}
func Abc_SclCountBufferFanoutsInt(pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i, Counter int32;
    Abc_ObjForEachFanout( pObj, pFanout, i )
        if ( Abc_ObjIsBuffer(pFanout) )
            Counter += Abc_SclCountBufferFanoutsInt( pFanout );
    return Counter + Abc_ObjIsBuffer(pObj);
}
func Abc_SclCountBufferFanouts(pObj *Abc_Obj_t) int32 {
    return Abc_SclCountBufferFanoutsInt(pObj) - Abc_ObjIsBuffer(pObj);
}
func Abc_SclCountNonBufferFanoutsInt(pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i, Counter int32;
    if !(Abc_ObjIsBuffer(pObj) != 0) {
        return 1; }
    Abc_ObjForEachFanout( pObj, pFanout, i )
        Counter += Abc_SclCountNonBufferFanoutsInt( pFanout );
    return Counter;
}
func Abc_SclCountNonBufferFanouts(pObj *Abc_Obj_t) int32 {
    var pFanout *Abc_Obj_t;
    var i, Counter int32;
    Abc_ObjForEachFanout( pObj, pFanout, i )
        Counter += Abc_SclCountNonBufferFanoutsInt( pFanout );
    return Counter;
}
func Abc_SclCountNonBufferDelayInt(p *SC_Man, pObj *Abc_Obj_t) float32 {
    var pFanout *Abc_Obj_t;
    var Delay float32;
    var i int32; 
    if !(Abc_ObjIsBuffer(pObj) != 0) {
        return Abc_SclObjTimeMax(p, pObj); }
    Abc_ObjForEachFanout( pObj, pFanout, i )
        Delay += Abc_SclCountNonBufferDelayInt( p, pFanout );
    return Delay;
}
func Abc_SclCountNonBufferDelay(p *SC_Man, pObj *Abc_Obj_t) float32 {
    var pFanout *Abc_Obj_t;
    var Delay float32;
    var i int32; 
    Abc_ObjForEachFanout( pObj, pFanout, i )
        Delay += Abc_SclCountNonBufferDelayInt( p, pFanout );
    return Delay;
}
func Abc_SclCountNonBufferLoadInt(p *SC_Man, pObj *Abc_Obj_t) float32 {
    var pFanout *Abc_Obj_t;
    var Load float32;
    var i int32; 
    if !(Abc_ObjIsBuffer(pObj) != 0) {
        return 0; }
    Abc_ObjForEachFanout( pObj, pFanout, i )
        Load += Abc_SclCountNonBufferLoadInt( p, pFanout );
    Load += 0.5 * Abc_SclObjLoad(p, pObj).rise + 0.5 * Abc_SclObjLoad(p, pObj).fall;
    Load -= 0.5 * SC_CellPin(Abc_SclObjCell(pObj), 0).rise_cap + 0.5 * SC_CellPin(Abc_SclObjCell(pObj), 0).fall_cap;
    return Load;
}
func Abc_SclCountNonBufferLoad(p *SC_Man, pObj *Abc_Obj_t) float32 {
    var pFanout *Abc_Obj_t;
    var Load float32;
    var i int32; 
    Abc_ObjForEachFanout( pObj, pFanout, i )
        Load += Abc_SclCountNonBufferLoadInt( p, pFanout );
    Load += 0.5 * Abc_SclObjLoad(p, pObj).rise + 0.5 * Abc_SclObjLoad(p, pObj).fall;
    return Load;
}
func Abc_SclPrintBuffersOne(p *SC_Man, pObj *Abc_Obj_t, nOffset int32) {
    var i int32;
    for  i = 0; i < nOffset; i++  {
        fmt.Printf("    ");
    }
    fmt.Printf("%6d: %-16s (%2d:%3d:%3d)  ", Abc_ObjId(pObj), Abc_ObjIsPi(pObj) ? "pi" : Mio_GateReadName((Mio_Gate_t *)pObj.pData), Abc_ObjFanoutNum(pObj), Abc_SclCountBufferFanouts(pObj), Abc_SclCountNonBufferFanouts(pObj));
    for  ; i < 4; i++  {
        fmt.Printf("    ");
    }
    fmt.Printf("a =%5.2f  ", Abc_ObjIsPi(pObj) ? 0 : Abc_SclObjCell(pObj).area);
    fmt.Printf("d = (");
    fmt.Printf("%6.0f ps; ", Abc_SclObjTimeOne(p, pObj, 1));
    fmt.Printf("%6.0f ps)  ", Abc_SclObjTimeOne(p, pObj, 0));
    fmt.Printf("l =%5.0f ff  ", Abc_SclObjLoadMax(p, pObj));
    fmt.Printf("s =%5.0f ps   ", Abc_SclObjSlewMax(p, pObj));
    fmt.Printf("sl =%5.0f ps   ", Abc_SclObjSlackMax(p, pObj, p.MaxDelay0));
    if nOffset == 0 {
    fmt.Printf("L =%5.0f ff   ", Abc_SclCountNonBufferLoad(p, pObj));
    fmt.Printf("Lx =%5.0f ff  ", 100.0*Abc_SclCountNonBufferLoad(p, pObj)/p.EstLoadAve);
    fmt.Printf("Dx =%5.0f ps  ", Abc_SclCountNonBufferDelay(p, pObj)/Abc_SclCountNonBufferFanouts(pObj) - Abc_SclObjTimeOne(p, pObj, 1));
    fmt.Printf("Cx =%5.0f ps", (Abc_SclCountNonBufferDelay(p, pObj)/Abc_SclCountNonBufferFanouts(pObj) - Abc_SclObjTimeOne(p, pObj, 1))/math.Log(float64(Abc_SclCountNonBufferLoad(p, pObj)/p.EstLoadAve)));
    }
    fmt.Printf("\n");
}
func Abc_SclPrintBuffersInt(p *SC_Man, pObj *Abc_Obj_t, nOffset int32) {
    var pFanout *Abc_Obj_t;
    var i int32;
    Abc_SclPrintBuffersOne( p, pObj, nOffset );
    assert((Abc_ObjIsBuffer(pObj) != 0), "Abc_ObjIsBuffer(pObj)");
    Abc_ObjForEachFanout( pObj, pFanout, i )
        if ( Abc_ObjIsBuffer(pFanout) )
            Abc_SclPrintBuffersInt( p, pFanout, nOffset + 1 );
}
func Abc_SclPrintBufferTrees(p *SC_Man, pNtk *Abc_Ntk_t) {
    var pObj, pFanout *Abc_Obj_t;
    var i, k int32;
    Abc_NtkForEachObj( pNtk, pObj, i )
    {
        if !(Abc_ObjIsBuffer(pObj) != 0) && Abc_SclCountBufferFanouts(pObj) > 3 {           
            Abc_SclPrintBuffersOne( p, pObj, 0 );
            Abc_ObjForEachFanout( pObj, pFanout, k )
                if ( Abc_ObjIsBuffer(pFanout) )
                    Abc_SclPrintBuffersInt( p, pFanout, 1 );
            fmt.Printf("\n");
        }
    }
}
func Abc_SclPrintBuffers(pLib *SC_Lib, pNtk *Abc_Ntk_t, fVerbose int32) {
    var fUseWireLoads int32;
    var p *SC_Man;
    assert(Abc_NtkIsMappedLogic(pNtk), "Abc_NtkIsMappedLogic(pNtk)");
    p = Abc_SclManStart( pLib, pNtk, fUseWireLoads, 1, 0, 10000 ); 
    Abc_SclPrintBufferTrees( p, pNtk ); 
//    Abc_SclPrintFaninPairs( p, pNtk );
    Abc_SclManFree( p );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

