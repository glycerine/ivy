package scl

/**CFile****************************************************************

  FileName    [sclLib.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Simplified library representation for STA.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclLib.h,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>
//#include <math.h>
//#include "misc/vec/vec.h"

//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

const ABC_SCL_CUR_VERSION = 8

 type SC_Dir int32
const  
(
    sc_dir_NULL SC_Dir = iota

    sc_dir_Input

    sc_dir_Output

    sc_dir_InOut

    sc_dir_Internal


)

 

 type SC_TSense int32
const      // -- timing sense, positive-, negative- or non-unate
(
    sc_ts_NULL SC_TSense = iota

    sc_ts_Pos

    sc_ts_Neg

    sc_ts_Non


)

 

           
type SC_Pair struct {
    rise      float32;
    fall      float32;
}
          
type SC_PairI struct {
    rise        int32;
    fall        int32;
}

      
type SC_SizePars struct {
    nIters        int32;
    nIterNoChange        int32;
    Window        int32;           // used for upsizing
    Ratio        int32;            // used for upsizing
    Notches        int32;
    DelayUser        int32;
    DelayGap        int32;
    TimeOut        int32;
    BuffTreeEst        int32;      // ratio for buffer tree estimation
    BypassFreq        int32;       // frequency to try bypassing
    fUseDept        int32;
    fDumpStats        int32;
    fUseWireLoads        int32;
    fVerbose        int32;
    fVeryVerbose        int32;
}

       
type SC_BusPars struct {
    GainRatio        int32;       // target gain
    Slew        int32;            // target slew
    nDegree        int32;         // max branching factor
    fSizeOnly        int32;       // perform only sizing
    fAddBufs        int32;        // add buffers
    fBufPis        int32;         // use CI buffering
    fUseWireLoads        int32;   // wire loads
    fVerbose        int32;        // verbose
    fVeryVerbose        int32;    // verbose
}

////////////////////////////////////////////////////////////////////////
///                    STRUCTURE DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

type SC_WireLoad =    SC_WireLoad_;
type SC_WireLoadSel = SC_WireLoadSel_;
type SC_TableTempl =  SC_TableTempl_;
type SC_Surface =     SC_Surface_;
type SC_Timing =      SC_Timing_;
type SC_Timings =     SC_Timings_;
type SC_Pin =         SC_Pin_;
type SC_Cell =        SC_Cell_;
type SC_Lib =         SC_Lib_;

type SC_WireLoad_ struct {
    pName          string;
    cap          float32;            // }- multiply estimation in 'fanout_len[].snd' with this value
    slope          float32;          // used to extrapolate wireload for large fanout count
    vFanout      Vec_Int_t;        // Vec<Pair<uint,float> > -- pairs '(#fanouts, est-wire-len)'
    vLen      Vec_Flt_t;
}

type SC_WireLoadSel_ struct {
    pName          string;
    vAreaFrom      Vec_Flt_t;      // Vec<Trip<float,float,Str> > -- triplets '(from-area, upto-area, wire-load-model)'; range is [from, upto[
    vAreaTo      Vec_Flt_t;
    vWireLoadModel      Vec_Ptr_t;
}

type SC_TableTempl_ struct {
    pName          string;
    vVars      Vec_Ptr_t;          // Vec<Str>         -- name of variable (numbered from 0, not 1 as in the Liberty file) 
    vIndex      Vec_Ptr_t;         // Vec<Vec<float> > -- this is the point of measurement in table for the given variable 
}

type SC_Surface_ struct {
    pName          string;
    vIndex0      Vec_Flt_t;        // Vec<float>       -- correspondes to "index_1" in the liberty file (for timing: slew)
    vIndex1      Vec_Flt_t;        // Vec<float>       -- correspondes to "index_2" in the liberty file (for timing: load)
    vData      Vec_Ptr_t;          // Vec<Vec<float> > -- 'data[i0][i1]' gives value at '(index0[i0], index1[i1])' 
    vIndex0I      Vec_Int_t;       // Vec<float>       -- correspondes to "index_1" in the liberty file (for timing: slew)
    vIndex1I      Vec_Int_t;       // Vec<float>       -- correspondes to "index_2" in the liberty file (for timing: load)
    vDataI      Vec_Ptr_t;         // Vec<Vec<float> > -- 'data[i0][i1]' gives value at '(index0[i0], index1[i1])' 
    approx [3][6]float32;
}

type SC_Timing_ struct {
    related_pin          string;    // -- related pin
    tsense      SC_TSense;         // -- timing sense (positive_unate, negative_unate, non_unate)
    when_text          string;      // -- logic condition on inputs triggering this delay model for the output (currently not used)
    pCellRise     SC_Surface;      // -- Used to compute pin-to-pin delay
    pCellFall     SC_Surface;
    pRiseTrans     SC_Surface;     // -- Used to compute output slew
    pFallTrans     SC_Surface;
}

type SC_Timings_ struct {
    pName          string;          // -- the 'related_pin' field
    vTimings      Vec_Ptr_t;       // structures of type SC_Timing
}

type SC_Pin_ struct {
    pName          string;
    dir         SC_Dir;
    cap          float32;            // -- this value is used if 'rise_cap' and 'fall_cap' is missing (copied by 'postProcess()'). (not used)
    rise_cap          float32;       // }- used for input pins ('cap' too).
    fall_cap          float32;       // }
    rise_capI            int32;      // }- used for input pins ('cap' too).
    fall_capI            int32;      // }
    max_out_cap          float32;    // } (not used)
    max_out_slew          float32;   // }- used only for output pins (max values must not be exceeded or else mapping is illegal) (not used)
    func_text          string;      // }
    vFunc      Vec_Wrd_t;          // }
    vRTimings      Vec_Ptr_t;      // -- for output pins
//    SC_Timing      Timing;         // -- for output pins  
}

type SC_Cell_ struct {
    pName          string;
    Id            int32;
    fSkip            int32;          // skip this cell during genlib computation
    seq            int32;            // -- set to TRUE by parser if a sequential element
    unsupp            int32;         // -- set to TRUE by parser if cell contains information we cannot handle
    area          float32;
    leakage          float32;
    areaI            int32;
    leakageI            int32;
    drive_strength            int32; // -- some library files provide this field (currently unused, but may be a good hint for sizing) (not used)
    vPins      Vec_Ptr_t;          // NamedSet<SC_Pin> 
    n_inputs            int32;       // -- 'pins[0 .. n_inputs-1]' are input pins
    n_outputs            int32;      // -- 'pins[n_inputs .. n_inputs+n_outputs-1]' are output pins
    pNext       *SC_Cell;          // same-functionality cells linked into a ring by area
    pPrev       *SC_Cell;          // same-functionality cells linked into a ring by area
    pRepr       *SC_Cell;          // representative of the class
    pAve       *SC_Cell;           // average size cell of this class
    Order            int32;          // order of the gate in the list
    nGates            int32;         // the number of gates in the list      
}

type SC_Lib_ struct {
    pName          string;
    pFileName          string;
    default_wire_load          string;
    default_wire_load_sel          string;
    default_max_out_slew          float32;   // -- 'default_max_transition'; this is copied to each output pin where 'max_transition' is not defined  (not used)
    unit_time            int32;      // -- Valid 9..12. Unit is '10^(-val)' seconds (e.g. 9=1ns, 10=100ps, 11=10ps, 12=1ps)
    unit_cap_fst          float32;   // -- First part is a multiplier, second either 12 or 15 for 'pf' or 'ff'.
    unit_cap_snd            int32;
    vWireLoads      Vec_Ptr_t;     // NamedSet<SC_WireLoad>
    vWireLoadSels      Vec_Ptr_t;  // NamedSet<SC_WireLoadSel>
    vTempls      Vec_Ptr_t;        // NamedSet<SC_TableTempl>  
    vCells      Vec_Ptr_t;         // NamedSet<SC_Cell>
    vCellClasses      Vec_Ptr_t;   // NamedSet<SC_Cell>
    pBins           *int32;          // hashing gateName -> gateId
    nBins            int32;
}

////////////////////////////////////////////////////////////////////////
///                       GLOBAL VARIABLES                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                       MACRO DEFINITIONS                          ///
////////////////////////////////////////////////////////////////////////

func SC_PairClean(d *SC_Pair) { d.rise = d.fall = 0;                 }
func SC_PairMax(d *SC_Pair) float32 { return Abc_MaxFloat(d.rise, d.fall); }
func SC_PairMin(d *SC_Pair) float32 { return Abc_MinFloat(d.rise, d.fall); }
func SC_PairAve(d *SC_Pair) float32 { return 0.5 * d.rise + 0.5 * d.fall;  }
func SC_PairDup(d *SC_Pair, s *SC_Pair) { *d = *s;                               }
func SC_PairMove(d *SC_Pair, s *SC_Pair) { *d = *s; s.rise = s.fall = 0;        }
func SC_PairAdd(d *SC_Pair, s *SC_Pair) { d.rise += s.rise; d.fall += s.fall;}
func SC_PairEqual(d *SC_Pair, s *SC_Pair) int32 { return d.rise == s.rise && d.fall == s.fall;                }
func SC_PairEqualE(d *SC_Pair, s *SC_Pair, E float32) int32 { return d.rise - s.rise < E && s.rise - d.rise < E &&  d.fall - s.fall < E && s.fall - d.fall < E;    }

func SC_LibCellNum(p *SC_Lib) int32 { return Vec_PtrSize(&p.vCells);                                  }
func SC_LibCell(p *SC_Lib, i int32) *SC_Cell { return (SC_Cell *)Vec_PtrEntry(&p.vCells, i);                   }
func SC_CellPin(p *SC_Cell, i int32) *SC_Pin { return (SC_Pin *)Vec_PtrEntry(&p.vPins, i);                     }
func SC_CellFunc(p *SC_Cell) *Vec_Wrd_t { return &SC_CellPin(p, p.n_inputs).vFunc;                       }
func SC_CellPinCap(p *SC_Cell, i int32) float32 { return 0.5 * SC_CellPin(p, i).rise_cap + 0.5 * SC_CellPin(p, i).fall_cap; }
func SC_CellPinCapAve(p *SC_Cell) float32 { var i int32; var c float32; for i = 0; i < p.n_inputs; i++ { c += SC_CellPinCap(p, i); }
 return c / Abc_MaxInt(1, p.n_inputs); }
func SC_CellPinOutFunc(p *SC_Cell, i int32) string { return SC_CellPin(p, p.n_inputs + i).func_text;               }
func SC_CellPinName(p *SC_Cell, i int32) string { return SC_CellPin(p, i).pName;                                 }

// Original macro:
// #define SC_LibForEachCell( p, pCell, i )         Vec_PtrForEachEntry( SC_Cell *, &p->vCells, pCell, i )

// Original macro:
// #define SC_LibForEachCellClass( p, pCell, i )    Vec_PtrForEachEntry( SC_Cell *, &p->vCellClasses, pCell, i )

// Original macro:
// #define SC_LibForEachWireLoad( p, pWL, i )       Vec_PtrForEachEntry( SC_WireLoad *, &p->vWireLoads, pWL, i )

// Original macro:
// #define SC_LibForEachWireLoadSel( p, pWLS, i )   Vec_PtrForEachEntry( SC_WireLoadSel *, &p->vWireLoadSels, pWLS, i )

// Original macro:
// #define SC_LibForEachTempl( p, pTempl, i )       Vec_PtrForEachEntry( SC_TableTempl *, &p->vTempls, pTempl, i )

// Original macro:
// #define SC_CellForEachPin( p, pPin, i )          Vec_PtrForEachEntry( SC_Pin *, &p->vPins, pPin, i )

// Original macro:
// #define SC_CellForEachPinIn( p, pPin, i )        Vec_PtrForEachEntryStop( SC_Pin *, &p->vPins, pPin, i, p->n_inputs )

// Original macro:
// #define SC_CellForEachPinOut( p, pPin, i )       Vec_PtrForEachEntryStart( SC_Pin *, &p->vPins, pPin, i, p->n_inputs )

// Original macro:
// #define SC_RingForEachCell( pRing, pCell, i )    for ( i = 0, pCell = pRing; i == 0 || pCell != pRing; pCell = pCell->pNext, i++ )

// Original macro:
// #define SC_RingForEachCellRev( pRing, pCell, i ) for ( i = 0, pCell = pRing; i == 0 || pCell != pRing; pCell = pCell->pPrev, i++ )

// Original macro:
// #define SC_PinForEachRTiming( p, pRTime, i )     Vec_PtrForEachEntry( SC_Timings *, &p->vRTimings, pRTime, i )



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Constructors of the library data-structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclWireLoadAlloc() *SC_WireLoad {
    
    // inlining of macro ABC_CALLOC
    p := &SC_WireLoad{};
    return p;
}
func Abc_SclWireLoadSelAlloc() *SC_WireLoadSel {
    
    // inlining of macro ABC_CALLOC
    p := &SC_WireLoadSel{};
    return p;
}
func Abc_SclTableTemplAlloc() *SC_TableTempl {
    
    // inlining of macro ABC_CALLOC
    p := &SC_TableTempl{};
    return p;
}
func Abc_SclSurfaceAlloc() *SC_Surface {
    
    // inlining of macro ABC_CALLOC
    p := &SC_Surface{};
    return p;
}
func Abc_SclTimingAlloc() *SC_Timing {
    
    // inlining of macro ABC_CALLOC
    p := &SC_Timing{};
    return p;
}
func Abc_SclTimingsAlloc() *SC_Timings {
    
    // inlining of macro ABC_CALLOC
    p := &SC_Timings{};
    return p;
}
func Abc_SclPinAlloc() *SC_Pin {
    
    // inlining of macro ABC_CALLOC
    p := &SC_Pin{};
    p.max_out_slew = -1;
    return p;
}
func Abc_SclCellAlloc() *SC_Cell {
    
    // inlining of macro ABC_CALLOC
    p := &SC_Cell{};
    return p;
}
func Abc_SclLibAlloc() *SC_Lib {
    
    // inlining of macro ABC_CALLOC
    p := &SC_Lib{};
    p.default_max_out_slew = -1;
    p.unit_time      = 9;
    p.unit_cap_fst   = 1;
    p.unit_cap_snd   = 12;
    return p;
}


/**Function*************************************************************

  Synopsis    [Destructors of the library data-structures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclWireLoadFree(p *SC_WireLoad) {
    Vec_IntErase( &p.vFanout );
    Vec_FltErase( &p.vLen );
    p.pName = nil;
    p = nil;
}
func Abc_SclWireLoadSelFree(p *SC_WireLoadSel) {
    Vec_FltErase( &p.vAreaFrom );
    Vec_FltErase( &p.vAreaTo );
    Vec_PtrFreeData( &p.vWireLoadModel );
    Vec_PtrErase( &p.vWireLoadModel );
    p.pName = nil;
    p = nil;
}
func Abc_SclTableTemplFree(p *SC_TableTempl) {
    Vec_PtrFreeData( &p.vVars );
    Vec_PtrErase( &p.vVars );
    Vec_VecErase( (Vec_Vec_t *)&p.vIndex );
    p.pName = nil;
    p = nil;
}
func Abc_SclSurfaceFree(p *SC_Surface) {
    Vec_FltErase( &p.vIndex0 );
    Vec_FltErase( &p.vIndex1 );
    Vec_IntErase( &p.vIndex0I );
    Vec_IntErase( &p.vIndex1I );
    Vec_VecErase( (Vec_Vec_t *)&p.vData );
    Vec_VecErase( (Vec_Vec_t *)&p.vDataI );
    p.pName = nil;
//    ABC_FREE( p );
}
func Abc_SclTimingFree(p *SC_Timing) {
    Abc_SclSurfaceFree( &p.pCellRise );
    Abc_SclSurfaceFree( &p.pCellFall );
    Abc_SclSurfaceFree( &p.pRiseTrans );
    Abc_SclSurfaceFree( &p.pFallTrans );
    p.related_pin = nil;
    p.when_text = nil;
    p = nil;
}
func Abc_SclTimingsFree(p *SC_Timings) {
    var pTemp *SC_Timing;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vTimings)) && (((pTemp) = (SC_Timing *)Vec_PtrEntry(&p.vTimings, i)), 1); i++  {
        Abc_SclTimingFree( pTemp );
    }
    Vec_PtrErase( &p.vTimings );
    p.pName = nil;
    p = nil;
}
func Abc_SclPinFree(p *SC_Pin) {
    var pTemp *SC_Timings;
    var i int32;
    // inlining of macro SC_PinForEachRTiming
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vRTimings)) && (((pTemp) = (SC_Timings *)Vec_PtrEntry(&p.vRTimings, i)), 1); i++  {
        Abc_SclTimingsFree( pTemp );
    }
    Vec_PtrErase( &p.vRTimings );
    Vec_WrdErase( &p.vFunc );
    p.func_text = nil;
    p.pName = nil;
    p = nil;
}
func Abc_SclCellFree(p *SC_Cell) {
    var pTemp *SC_Pin;
    var i int32;
    // inlining of macro SC_CellForEachPin
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vPins)) && (((pTemp) = (SC_Pin *)Vec_PtrEntry(&p.vPins, i)), 1); i++  {
        Abc_SclPinFree( pTemp );
    }
    Vec_PtrErase( &p.vPins );
    p.pName = nil;
    p = nil;
}
func Abc_SclLibFree(p *SC_Lib) {
    var pWL *SC_WireLoad;
    var pWLS *SC_WireLoadSel;
    var pTempl *SC_TableTempl;
    var pCell *SC_Cell;
    var i int32;
    // inlining of macro SC_LibForEachWireLoad
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vWireLoads)) && (((pWL) = (SC_WireLoad *)Vec_PtrEntry(&p.vWireLoads, i)), 1); i++  {
        Abc_SclWireLoadFree( pWL );
    }
    Vec_PtrErase( &p.vWireLoads );
    // inlining of macro SC_LibForEachWireLoadSel
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vWireLoadSels)) && (((pWLS) = (SC_WireLoadSel *)Vec_PtrEntry(&p.vWireLoadSels, i)), 1); i++  {
        Abc_SclWireLoadSelFree( pWLS );
    }
    Vec_PtrErase( &p.vWireLoadSels );
    // inlining of macro SC_LibForEachTempl
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vTempls)) && (((pTempl) = (SC_TableTempl *)Vec_PtrEntry(&p.vTempls, i)), 1); i++  {
        Abc_SclTableTemplFree( pTempl );
    }
    Vec_PtrErase( &p.vTempls );
    // inlining of macro SC_LibForEachCell
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&p.vCells)) && (((pCell) = (SC_Cell *)Vec_PtrEntry(&p.vCells, i)), 1); i++  {
        Abc_SclCellFree( pCell );
    }
    Vec_PtrErase( &p.vCells );
    Vec_PtrErase( &p.vCellClasses );
    p.pName = nil;
    p.pFileName = nil;
    p.default_wire_load = nil;
    p.default_wire_load_sel = nil;
    p.pBins = nil;
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Lookup table delay computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibLookup(p *SC_Surface, slew float32, load float32) float32 {
    var pIndex0, pIndex1, pDataS, pDataS1 *float32;
    var sfrac, lfrac, p0, p1 float32;
    var s, l int32;

    // handle constant table
    if Vec_FltSize(&p.vIndex0) == 1 && Vec_FltSize(&p.vIndex1) == 1 {
        var vTemp *Vec_Flt_t  = (Vec_Flt_t *)Vec_PtrEntry(&p.vData, 0);
        assert(Vec_PtrSize(&p.vData) == 1, "Vec_PtrSize(&p.vData) == 1");
        assert(Vec_FltSize(vTemp) == 1, "Vec_FltSize(vTemp) == 1");
        return Vec_FltEntry(vTemp, 0);
    }

    // Find closest sample points in surface:
    pIndex0 = Vec_FltArray(&p.vIndex0);
    for  s = 1; s < Vec_FltSize(&p.vIndex0)-1; s++  {
        if pIndex0[s] > slew {
            break; }
    }
    s--;

    pIndex1 = Vec_FltArray(&p.vIndex1);
    for  l = 1; l < Vec_FltSize(&p.vIndex1)-1; l++  {
        if pIndex1[l] > load {
            break; }
    }
    l--;

    // Interpolate (or extrapolate) function value from sample points:
    sfrac = (slew - pIndex0[s]) / (pIndex0[s+1] - pIndex0[s]);
    lfrac = (load - pIndex1[l]) / (pIndex1[l+1] - pIndex1[l]);

    pDataS  = Vec_FltArray( (Vec_Flt_t *)Vec_PtrEntry(&p.vData, s) );
    pDataS1 = Vec_FltArray( (Vec_Flt_t *)Vec_PtrEntry(&p.vData, s+1) );

    p0 = pDataS [l] + lfrac * (pDataS [l+1] - pDataS [l]);
    p1 = pDataS1[l] + lfrac * (pDataS1[l+1] - pDataS1[l]);

    return p0 + sfrac * (p1 - p0);      // <<== multiply result with K factor here 
}
func Scl_LibPinArrival(pTime *SC_Timing, pArrIn *SC_Pair, pSlewIn *SC_Pair, pLoad *SC_Pair, pArrOut *SC_Pair, pSlewOut *SC_Pair) {
    if pTime.tsense == sc_ts_Pos || pTime.tsense == sc_ts_Non {
        pArrOut.rise  = Abc_MaxFloat( pArrOut.rise,  pArrIn.rise + Scl_LibLookup(&pTime.pCellRise,  pSlewIn.rise, pLoad.rise) );
        pArrOut.fall  = Abc_MaxFloat( pArrOut.fall,  pArrIn.fall + Scl_LibLookup(&pTime.pCellFall,  pSlewIn.fall, pLoad.fall) );
        pSlewOut.rise = Abc_MaxFloat( pSlewOut.rise,                Scl_LibLookup(&pTime.pRiseTrans, pSlewIn.rise, pLoad.rise) );
        pSlewOut.fall = Abc_MaxFloat( pSlewOut.fall,                Scl_LibLookup(&pTime.pFallTrans, pSlewIn.fall, pLoad.fall) );
    }
    if pTime.tsense == sc_ts_Neg || pTime.tsense == sc_ts_Non {
        pArrOut.rise  = Abc_MaxFloat( pArrOut.rise,  pArrIn.fall + Scl_LibLookup(&pTime.pCellRise,  pSlewIn.fall, pLoad.rise) );
        pArrOut.fall  = Abc_MaxFloat( pArrOut.fall,  pArrIn.rise + Scl_LibLookup(&pTime.pCellFall,  pSlewIn.rise, pLoad.fall) );
        pSlewOut.rise = Abc_MaxFloat( pSlewOut.rise,                Scl_LibLookup(&pTime.pRiseTrans, pSlewIn.fall, pLoad.rise) );
        pSlewOut.fall = Abc_MaxFloat( pSlewOut.fall,                Scl_LibLookup(&pTime.pFallTrans, pSlewIn.rise, pLoad.fall) );
    }
}
func Scl_LibPinDeparture(pTime *SC_Timing, pDepIn *SC_Pair, pSlewIn *SC_Pair, pLoad *SC_Pair, pDepOut *SC_Pair) {
    if pTime.tsense == sc_ts_Pos || pTime.tsense == sc_ts_Non {
        pDepIn.rise  = Abc_MaxFloat( pDepIn.rise,  pDepOut.rise + Scl_LibLookup(&pTime.pCellRise,  pSlewIn.rise, pLoad.rise) );
        pDepIn.fall  = Abc_MaxFloat( pDepIn.fall,  pDepOut.fall + Scl_LibLookup(&pTime.pCellFall,  pSlewIn.fall, pLoad.fall) );
    }
    if pTime.tsense == sc_ts_Neg || pTime.tsense == sc_ts_Non {
        pDepIn.fall  = Abc_MaxFloat( pDepIn.fall,  pDepOut.rise + Scl_LibLookup(&pTime.pCellRise,  pSlewIn.fall, pLoad.rise) );
        pDepIn.rise  = Abc_MaxFloat( pDepIn.rise,  pDepOut.fall + Scl_LibLookup(&pTime.pCellFall,  pSlewIn.rise, pLoad.fall) );
    }
}

/**Function*************************************************************

  Synopsis    [Lookup table delay computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibLookupI(p *SC_Surface, slew int32, load int32) int32 {
    var pIndex0, pIndex1, pDataS, pDataS1 *int32;
    var p0, p1, s, l int32;
    var lFrac0, lFrac1, sFrac iword;

    // handle constant table
    if Vec_IntSize(&p.vIndex0I) == 1 && Vec_IntSize(&p.vIndex1I) == 1 {
        var vTemp *Vec_Int_t  = (Vec_Int_t *)Vec_PtrEntry(&p.vDataI, 0);
        assert(Vec_PtrSize(&p.vDataI) == 1, "Vec_PtrSize(&p.vDataI) == 1");
        assert(Vec_IntSize(vTemp) == 1, "Vec_IntSize(vTemp) == 1");
        return Vec_IntEntry(vTemp, 0);
    }

    // Find closest sample points in surface:
    pIndex0 = Vec_IntArray(&p.vIndex0I);
    for  s = 1; s < Vec_IntSize(&p.vIndex0I)-1; s++  {
        if pIndex0[s] > slew {
            break; }
    }
    s--;

    pIndex1 = Vec_IntArray(&p.vIndex1I);
    for  l = 1; l < Vec_IntSize(&p.vIndex1I)-1; l++  {
        if pIndex1[l] > load {
            break; }
    }
    l--;

    pDataS  = Vec_IntArray( (Vec_Int_t *)Vec_PtrEntry(&p.vDataI, s) );
    pDataS1 = Vec_IntArray( (Vec_Int_t *)Vec_PtrEntry(&p.vDataI, s+1) );

    // Interpolate (or extrapolate) function value from sample points:
//    lfrac = (load - pIndex1[l]) / (pIndex1[l+1] - pIndex1[l]);
//    sfrac = (slew - pIndex0[s]) / (pIndex0[s+1] - pIndex0[s]);

    lFrac0 = (iword)(pDataS [l+1] - pDataS [l]) * (iword)(load - pIndex1[l]) / (iword)(pIndex1[l+1] - pIndex1[l]);
    lFrac1 = (iword)(pDataS1[l+1] - pDataS1[l]) * (iword)(load - pIndex1[l]) / (iword)(pIndex1[l+1] - pIndex1[l]);

//    p0 = pDataS [l] + lfrac * (pDataS [l+1] - pDataS [l]);
//    p1 = pDataS1[l] + lfrac * (pDataS1[l+1] - pDataS1[l]);

    p0 = pDataS [l] + int32(lFrac0);
    p1 = pDataS1[l] + int32(lFrac1);

    sFrac = (iword)(p1 - p0) * (iword)(slew - pIndex0[s]) / (iword)(pIndex0[s+1] - pIndex0[s]);

//    return p0 + sfrac * (p1 - p0); 
    return p0 + int32(sFrac);   
}
func Scl_LibPinArrivalI(pTime *SC_Timing, pArrIn *SC_PairI, pSlewIn *SC_PairI, pLoad *SC_PairI, pArrOut *SC_PairI, pSlewOut *SC_PairI, pArray *int32) {
    if pTime.tsense == sc_ts_Pos || pTime.tsense == sc_ts_Non {
        pArrOut.rise  = Abc_MaxInt( pArrOut.rise,  pArrIn.rise + (pArray[0] = Scl_LibLookupI(&pTime.pCellRise,  pSlewIn.rise, pLoad.rise)) );
        pArrOut.fall  = Abc_MaxInt( pArrOut.fall,  pArrIn.fall + (pArray[1] = Scl_LibLookupI(&pTime.pCellFall,  pSlewIn.fall, pLoad.fall)) );
        pSlewOut.rise = Abc_MaxInt( pSlewOut.rise,                             Scl_LibLookupI(&pTime.pRiseTrans, pSlewIn.rise, pLoad.rise) );
        pSlewOut.fall = Abc_MaxInt( pSlewOut.fall,                             Scl_LibLookupI(&pTime.pFallTrans, pSlewIn.fall, pLoad.fall) );
    }
    if pTime.tsense == sc_ts_Neg || pTime.tsense == sc_ts_Non {
        pArrOut.rise  = Abc_MaxInt( pArrOut.rise,  pArrIn.fall + (pArray[2] = Scl_LibLookupI(&pTime.pCellRise,  pSlewIn.fall, pLoad.rise)) );
        pArrOut.fall  = Abc_MaxInt( pArrOut.fall,  pArrIn.rise + (pArray[3] = Scl_LibLookupI(&pTime.pCellFall,  pSlewIn.rise, pLoad.fall)) );
        pSlewOut.rise = Abc_MaxInt( pSlewOut.rise,                             Scl_LibLookupI(&pTime.pRiseTrans, pSlewIn.fall, pLoad.rise) );
        pSlewOut.fall = Abc_MaxInt( pSlewOut.fall,                             Scl_LibLookupI(&pTime.pFallTrans, pSlewIn.rise, pLoad.fall) );
    }
}
func Scl_LibPinRequiredI(pTime *SC_Timing, pReqIn *SC_PairI, pReqOut *SC_PairI, pArray *int32) {
    if pTime.tsense == sc_ts_Pos || pTime.tsense == sc_ts_Non {
        pReqIn.rise  = Abc_MinInt( pReqIn.rise,  pReqOut.rise - pArray[0] );
        pReqIn.fall  = Abc_MinInt( pReqIn.fall,  pReqOut.fall - pArray[1] );
    }
    if pTime.tsense == sc_ts_Neg || pTime.tsense == sc_ts_Non {
        pReqIn.fall  = Abc_MinInt( pReqIn.fall,  pReqOut.rise - pArray[2] );
        pReqIn.rise  = Abc_MinInt( pReqIn.rise,  pReqOut.fall - pArray[3] );
    }
}

/**Function*************************************************************

  Synopsis    [Compute one timing edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_CellPinTime(pCell *SC_Cell, iPin int32) *SC_Timing {
    var pPin *SC_Pin;
    var pRTime *SC_Timings;
    assert(iPin >= 0 && iPin < pCell.n_inputs, "iPin >= 0 && iPin < pCell.n_inputs");
    pPin = SC_CellPin( pCell, pCell.n_inputs );
    assert(Vec_PtrSize(&pPin.vRTimings) == pCell.n_inputs, "Vec_PtrSize(&pPin.vRTimings) == pCell.n_inputs");
    pRTime = (SC_Timings *)Vec_PtrEntry( &pPin.vRTimings, iPin );
    if Vec_PtrSize(&pRTime.vTimings) == 0 {
        return nil; }
    assert(Vec_PtrSize(&pRTime.vTimings) == 1, "Vec_PtrSize(&pRTime.vTimings) == 1");
    return (SC_Timing *)Vec_PtrEntry( &pRTime.vTimings, 0 );
}
func Scl_LibPinArrivalEstimate(pCell *SC_Cell, iPin int32, Slew float32, Load float32) float32 {
    var LoadIn SC_Pair  = SC_Pair{ Load, Load };
    var ArrIn SC_Pair   = SC_Pair{ 0.0, 0.0 };
    var ArrOut SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var SlewIn SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var SlewOut SC_Pair  = SC_Pair{ 0.0, 0.0 };
//    Vec_Flt_t * vIndex0 = pTime->pCellRise->vIndex0; // slew
//    SlewIn.fall = SlewIn.rise = Vec_FltEntry( vIndex0, Vec_FltSize(vIndex0)/2 );
    SlewIn.fall = SlewIn.rise = Slew; 
    Scl_LibPinArrival( Scl_CellPinTime(pCell, iPin), &ArrIn, &SlewIn, &LoadIn, &ArrOut, &SlewOut );
    return  0.5 * ArrOut.fall +  0.5 * ArrOut.rise;
}
func Scl_LibHandleInputDriver(pCell *SC_Cell, pLoadIn *SC_Pair, pArrOut *SC_Pair, pSlewOut *SC_Pair) {
    var LoadIn SC_Pair    = SC_Pair{ 0.0, 0.0 }; // zero input load
    var ArrIn SC_Pair     = SC_Pair{ 0.0, 0.0 }; // zero input time
    var SlewIn SC_Pair    = SC_Pair{ 0.0, 0.0 }; // zero input slew
    var ArrOut0 SC_Pair   = SC_Pair{ 0.0, 0.0 }; // output time under zero load
    var ArrOut1 SC_Pair   = SC_Pair{ 0.0, 0.0 }; // output time under given load
    var SlewOut SC_Pair   = SC_Pair{ 0.0, 0.0 }; // output slew under zero load 
    pSlewOut.fall = pSlewOut.rise = 0;
    assert(pCell.n_inputs == 1, "pCell.n_inputs == 1");
    Scl_LibPinArrival( Scl_CellPinTime(pCell, 0), &ArrIn, &SlewIn, &LoadIn, &ArrOut0, &SlewOut );
    Scl_LibPinArrival( Scl_CellPinTime(pCell, 0), &ArrIn, &SlewIn, pLoadIn, &ArrOut1, pSlewOut );
    pArrOut.fall = ArrOut1.fall - ArrOut0.fall;
    pArrOut.rise = ArrOut1.rise - ArrOut0.rise;
}

/**Function*************************************************************

  Synopsis    [Compute one timing edge.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Scl_LibPinArrivalEstimateI(pCell *SC_Cell, iPin int32, Slew int32, Load int32) int32 {
    var Arrray [4]int32;
    var LoadIn SC_PairI  = SC_PairI{ Load, Load };
    var ArrIn SC_PairI   = SC_PairI{ 0, 0 };
    var ArrOut SC_PairI  = SC_PairI{ 0, 0 };
    var SlewIn SC_PairI  = SC_PairI{ 0, 0 };
    var SlewOut SC_PairI  = SC_PairI{ 0, 0 };
//    Vec_Flt_t * vIndex0 = pTime->pCellRise->vIndex0; // slew
//    SlewIn.fall = SlewIn.rise = Vec_FltEntry( vIndex0, Vec_FltSize(vIndex0)/2 );
    SlewIn.fall = SlewIn.rise = Slew; 
    Scl_LibPinArrivalI( Scl_CellPinTime(pCell, iPin), &ArrIn, &SlewIn, &LoadIn, &ArrOut, &SlewOut, Arrray );
    return (ArrOut.fall + ArrOut.rise) >> 1;
}
func Scl_LibHandleInputDriver2(pCell *SC_Cell, pLoadIn *SC_PairI, pArrOut *SC_PairI, pSlewOut *SC_PairI) {
    var Arrray [4]int32;
    var LoadIn SC_PairI    = SC_PairI{ 0, 0 }; // zero input load
    var ArrIn SC_PairI     = SC_PairI{ 0, 0 }; // zero input time
    var SlewIn SC_PairI    = SC_PairI{ 0, 0 }; // zero input slew
    var ArrOut0 SC_PairI   = SC_PairI{ 0, 0 }; // output time under zero load
    var ArrOut1 SC_PairI   = SC_PairI{ 0, 0 }; // output time under given load
    var SlewOut SC_PairI   = SC_PairI{ 0, 0 }; // output slew under zero load 
    pSlewOut.fall = pSlewOut.rise = 0;
    assert(pCell.n_inputs == 1, "pCell.n_inputs == 1");
    Scl_LibPinArrivalI( Scl_CellPinTime(pCell, 0), &ArrIn, &SlewIn, &LoadIn, &ArrOut0, &SlewOut, Arrray );
    Scl_LibPinArrivalI( Scl_CellPinTime(pCell, 0), &ArrIn, &SlewIn, pLoadIn, &ArrOut1, pSlewOut, Arrray );
    pArrOut.fall = ArrOut1.fall - ArrOut0.fall;
    pArrOut.rise = ArrOut1.rise - ArrOut0.rise;
}

/*=== sclLiberty.c ===============================================================*/
// extern SC_Lib *      Abc_SclReadLiberty( char * pFileName, int fVerbose, int fVeryVerbose );

/*=== sclLibScl.c ===============================================================*/
// extern SC_Lib *      Abc_SclReadFromGenlib( void * pLib );

// extern SC_Lib *      Abc_SclReadFromStr( Vec_Str_t * vOut );

// extern SC_Lib *      Abc_SclReadFromFile( char * pFileName );

// extern void          Abc_SclWriteScl( char * pFileName, SC_Lib * p );

// extern void          Abc_SclWriteLiberty( char * pFileName, SC_Lib * p );

/*=== sclLibUtil.c ===============================================================*/
// extern void          Abc_SclHashCells( SC_Lib * p );

// extern int           Abc_SclCellFind( SC_Lib * p, char * pName );

// extern int           Abc_SclClassCellNum( SC_Cell * pClass );

// extern void          Abc_SclShortNames( SC_Lib * p );

// extern int           Abc_SclLibClassNum( SC_Lib * pLib );

// extern void          Abc_SclLinkCells( SC_Lib * p );

// extern void          Abc_SclPrintCells( SC_Lib * p, float Slew, float Gain, int fInvOnly, int fShort );

// extern void          Abc_SclConvertLeakageIntoArea( SC_Lib * p, float A, float B );

// extern void          Abc_SclLibNormalize( SC_Lib * p );

// extern SC_Cell *     Abc_SclFindInvertor( SC_Lib * p, int fFindBuff );

// extern SC_Cell *     Abc_SclFindSmallestGate( SC_Cell * p, float CinMin );

// extern SC_WireLoad * Abc_SclFindWireLoadModel( SC_Lib * p, float Area );

// extern SC_WireLoad * Abc_SclFetchWireLoadModel( SC_Lib * p, char * pName );

// extern int           Abc_SclHasDelayInfo( void * pScl );

// extern float         Abc_SclComputeAverageSlew( SC_Lib * p );

// extern void          Abc_SclDumpGenlib( char * pFileName, SC_Lib * p, float Slew, float Gain, int nGatesMin );

// extern void          Abc_SclInstallGenlib( void * pScl, float Slew, float Gain, int nGatesMin );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
