package scl

/**CFile****************************************************************

  FileName    [sclTime.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Timing/gate-sizing manager.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclTime.h,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "misc/vec/vec.h"
//#include "sclLib.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

             
type SC_Time struct {
    pLib        *SC_Lib;          // library
    vCis     *Vec_Int_t;          // comb inputs
    vCos     *Vec_Int_t;          // comb outputs 
    nObjs            int32;         // allocated size
    // get assignment
    vGates     *Vec_Int_t;        // mapping of objId into gateId
    // timing information
    pLoads       *SC_Pair;        // loads for each gate
    pLoads2       *SC_Pair;       // loads for each gate
    pDepts       *SC_Pair;        // departures for each gate
    pTimes       *SC_Pair;        // arrivals for each gate
    pSlews       *SC_Pair;        // slews for each gate
    pTimes2       *SC_Pair;       // arrivals for each gate
    pSlews2       *SC_Pair;       // slews for each gate
    pSlack         *float32;        // slacks for each gate
    pWLoadUsed   *SC_WireLoad;    // name of the used WireLoad model
    // optimization parameters
    SumArea          float32;       // total area
    MaxDelay          float32;      // max delay
    SumArea0          float32;      // total area at the begining 
    MaxDelay0          float32;     // max delay at the begining
    BestDelay          float32;     // best delay in the middle
}

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

func Scl_ObjCell(p *SC_Time, i int32) *SC_Cell { return SC_LibCell( p.pLib, Vec_IntEntry(p.vGates, i) );   }
func Scl_ObjSetCell(p *SC_Time, i int32, pCell *SC_Cell) { Vec_IntWriteEntry( p.vGates, i, pCell.Id );               }

func Scl_ObjLoad(p *SC_Time, i int32) *SC_Pair { return p.pLoads + i;  }
func Scl_ObjLoad2(p *SC_Time, i int32) *SC_Pair { return p.pLoads2 + i; }
func Scl_ObjDept(p *SC_Time, i int32) *SC_Pair { return p.pDepts + i;  }
func Scl_ObjTime(p *SC_Time, i int32) *SC_Pair { return p.pTimes + i;  }
func Scl_ObjSlew(p *SC_Time, i int32) *SC_Pair { return p.pSlews + i;  }
func Scl_ObjTime2(p *SC_Time, i int32) *SC_Pair { return p.pTimes2 + i; }
func Scl_ObjSlew2(p *SC_Time, i int32) *SC_Pair { return p.pSlews2 + i; }

func Scl_ObjTimeMax(p *SC_Time, i int32) float32 { return Abc_MaxFloat(Scl_ObjTime(p, i).rise, Scl_ObjTime(p, i).fall);  }
func Scl_ObjDepthMax(p *SC_Time, i int32) float32 { return Abc_MaxFloat(Scl_ObjDept(p, i).rise, Scl_ObjDept(p, i).fall);  }
func Scl_ObjGetSlack(p *SC_Time, i int32, D float32) float32 { return D - Abc_MaxFloat(Scl_ObjTime(p, i).rise + Scl_ObjDept(p, i).rise, Scl_ObjTime(p, i).fall + Scl_ObjDept(p, i).fall);  }
func Scl_ObjGetSlackR(p *SC_Time, i int32, D float32) float32 { return D - (Scl_ObjTime(p, i).rise + Scl_ObjDept(p, i).rise);  }
func Scl_ObjGetSlackF(p *SC_Time, i int32, D float32) float32 { return D - (Scl_ObjTime(p, i).fall + Scl_ObjDept(p, i).fall);  }
func Scl_ObjSlack(p *SC_Time, i int32) float32 { return p.pSlack[i];   }

func Scl_ObjDupFanin(p *SC_Time, i int32, iFanin int32) { *Scl_ObjTime(p, i) = *Scl_ObjTime(p, iFanin);                    }
func Scl_ObjGain(p *SC_Time, i int32) float32 { return 0.5*((Scl_ObjTime2(p, i).rise - Scl_ObjTime(p, i).rise) + (Scl_ObjTime2(p, i).fall - Scl_ObjTime(p, i).fall)); }
func Scl_ObjLegal(p *SC_Time, i int32, D float32) int32 { return Scl_ObjTime(p, i).rise <= Scl_ObjTime2(p, i).rise + Scl_ObjGetSlackR(p, i, D) && Scl_ObjTime(p, i).fall <= Scl_ObjTime2(p, i).fall + Scl_ObjGetSlackF(p, i, D); }

func Scl_ObjLoadFf(p *SC_Time, i int32, fRise int32) float64 { if fRise != 0 {
    return Scl_ObjLoad(p, i).rise;
} else {
    return Scl_ObjLoad(p, i).fall;
} }
func Scl_ObjTimePs(p *SC_Time, i int32, fRise int32) float64 { if fRise != 0 {
    return Scl_ObjTime(p, i).rise;
} else {
    return Scl_ObjTime(p, i).fall;
} }
func Scl_ObjSlewPs(p *SC_Time, i int32, fRise int32) float64 { if fRise != 0 {
    return Scl_ObjSlew(p, i).rise;
} else {
    return Scl_ObjSlew(p, i).fall;
} }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Constructor/destructor of STA manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_ManAlloc(pLib *SC_Lib, vCis *Vec_Int_t, vCos *Vec_Int_t, nObjs int32) *SC_Time {
    
    // inlining of macro ABC_CALLOC
    p := &SC_Time{};
    p.pLib      = pLib;
    p.vCis      = vCis;
    p.vCos      = vCos;
    p.nObjs     = nObjs;
    // inlining of macro ABC_CALLOC
    p.pLoads    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pLoads2   = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pDepts    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pTimes    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pSlews    = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pTimes2   = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_CALLOC
    p.pSlews2   = ((SC_Pair *) calloc((size_t)(p.nObjs), sizeof(SC_Pair)));
    // inlining of macro ABC_FALLOC
    p.pSlack    = ((float *) memset(malloc(sizeof(float) * (size_t)(p.nObjs)), 0xff, sizeof(float) * (size_t)(p.nObjs)));
    return p;
}
func Scl_ManFree(p *SC_Time) {
    Vec_IntFreeP( &p.vGates );
    // inlining of macro ABC_FREE
    ((p.pLoads) ? (free((char *) (p.pLoads)), (p.pLoads) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pLoads2) ? (free((char *) (p.pLoads2)), (p.pLoads2) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDepts) ? (free((char *) (p.pDepts)), (p.pDepts) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTimes) ? (free((char *) (p.pTimes)), (p.pTimes) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSlews) ? (free((char *) (p.pSlews)), (p.pSlews) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTimes2) ? (free((char *) (p.pTimes2)), (p.pTimes2) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSlews2) ? (free((char *) (p.pSlews2)), (p.pSlews2) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSlack) ? (free((char *) (p.pSlack)), (p.pSlack) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}
func Scl_ManCleanTime(p *SC_Time) {
    memset( p.pDepts, 0, sizeof(SC_Pair) * p.nObjs );
    memset( p.pTimes, 0, sizeof(SC_Pair) * p.nObjs );
    memset( p.pSlews, 0, sizeof(SC_Pair) * p.nObjs );
}


/**Function*************************************************************

  Synopsis    [Stores/retrieves timing information for the logic cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_ConeStore(p *SC_Time, vCone *Vec_Int_t) {
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCone)) && (((iObj) = Vec_IntEntry(vCone, i)), 1); i++  {
        *Scl_ObjTime2(p, iObj) = *Scl_ObjTime(p, iObj);
        *Scl_ObjSlew2(p, iObj) = *Scl_ObjSlew(p, iObj);
    }
}
func Scl_ConeRestore(p *SC_Time, vCone *Vec_Int_t) {
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCone)) && (((iObj) = Vec_IntEntry(vCone, i)), 1); i++  {
        *Scl_ObjTime(p, iObj) = *Scl_ObjTime2(p, iObj);
        *Scl_ObjSlew(p, iObj) = *Scl_ObjSlew2(p, iObj);
    }
}
func Scl_ConeClear(p *SC_Time, vCone *Vec_Int_t) {
    var Zero SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCone)) && (((iObj) = Vec_IntEntry(vCone, i)), 1); i++  {
        *Scl_ObjTime(p, iObj) = Zero;
        *Scl_ObjSlew(p, iObj) = Zero;
    }
}

/**Function*************************************************************

  Synopsis    [Timing computation for pin/gate/cone/network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_PinTimeArrival(p *SC_Time, pTime *SC_Timing, iObj int32, iFanin int32) {
    var pArrIn *SC_Pair    = Scl_ObjTime( p, iFanin );
    var pSlewIn *SC_Pair   = Scl_ObjSlew( p, iFanin );
    var pLoad *SC_Pair     = Scl_ObjLoad( p, iObj );
    var pArrOut *SC_Pair   = Scl_ObjTime( p, iObj );   // modified
    var pSlewOut *SC_Pair  = Scl_ObjSlew( p, iObj );   // modified
    Scl_LibPinArrival( pTime, pArrIn, pSlewIn, pLoad, pArrOut, pSlewOut );
}
func Scl_PinTimeDeparture(p *SC_Time, pTime *SC_Timing, iObj int32, iFanin int32) {
    var pDepIn *SC_Pair    = Scl_ObjDept( p, iFanin );   // modified
    var pSlewIn *SC_Pair   = Scl_ObjSlew( p, iFanin );
    var pLoad *SC_Pair     = Scl_ObjLoad( p, iObj );
    var pDepOut *SC_Pair   = Scl_ObjDept( p, iObj );
    Scl_LibPinDeparture( pTime, pDepIn, pSlewIn, pLoad, pDepOut );
}



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
