package scl

import (
	"fmt"
	"math"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [sclLibUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Various library utilities.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclLibUtil.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclLib.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reading library from file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_SclHashString_s_Primes [10]int32 = [10]int32{ 1291, 1699, 2357, 4177, 5147, 5647, 6343, 7103, 7873, 8147 };
func Abc_SclHashString(pName string, TableSize int32) uint32 {
    
    var i, Key uint32;
    for  i = 0; pName[i] != '\0'; i++  {
        Key += spStatic_Abc_SclHashString_s_Primes[i%10]*pName[i]*pName[i];
    }
    return Key % TableSize;
}
func Abc_SclHashLookup(p *SC_Lib, pName string) *int32 {
    var i int32;
    for  i = Abc_SclHashString(pName, p.nBins); i < p.nBins; i = (i + 1) % p.nBins  {
        if p.pBins[i] == -1 || !(int32(strings.Compare(pName, SC_LibCell(p, p.pBins[i]).pName)) != 0) {
            return p.pBins + i; }
    }
    assert(false, "0");
    return nil;
}
func Abc_SclHashCells(p *SC_Lib) {
    var pCell *SC_Cell;
    var i int32;
    var pPlace *int32;
    assert(p.nBins == 0, "p.nBins == 0");
    p.nBins = Abc_PrimeCudd( 5 * SC_LibCellNum(p) );
    p.pBins = ABC_FALLOC( int, p.nBins );
    SC_LibForEachCell( p, pCell, i )
    {
        pPlace = Abc_SclHashLookup( p, pCell.pName );
        assert(*pPlace == -1, "*pPlace == -1");
        *pPlace = i;
    }
}
func Abc_SclCellFind(p *SC_Lib, pName string) int32 {
    var pPlace *int32  = Abc_SclHashLookup( p, pName );
    if pPlace != nil {
        return *pPlace;
    } else {
        return -1;
    }
}
func Abc_SclClassCellNum(pClass *SC_Cell) int32 {
    var pCell *SC_Cell;
    var i, Count int32;
    SC_RingForEachCell( pClass, pCell, i )
        if ( !pCell.fSkip )
            Count++;
    return Count;
}
func Abc_SclLibClassNum(pLib *SC_Lib) int32 {
    var pRepr *SC_Cell;
    var i, Count int32;
    SC_LibForEachCellClass( pLib, pRepr, i )
        Count++;
    return Count;
}

/**Function*************************************************************

  Synopsis    [Change cell names and pin names.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclIsChar(c byte) int32 {
    return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_';
}
func Abc_SclIsName(c byte) int32 {
    return Abc_SclIsChar(c) || (c >= '0' && c <= '9');
}
func Abc_SclFindLimit(pName string) string {
    assert((Abc_SclIsChar(*pName) != 0), "Abc_SclIsChar(*pName)");
    for  Abc_SclIsName(*pName) != 0  {
        pName++;
    }
    return pName;
}
func Abc_SclAreEqual(pBase string, pName string, pLimit string) int32 {
    return !strncmp( pBase, pName, pLimit - pName );
}
func Abc_SclShortFormula(pCell *SC_Cell, pForm string, pBuffer string) {
    var pPin *SC_Pin; var i int32;
    var pTemp, pLimit string;
    for  pTemp = pForm; *pTemp;  {
        if !(Abc_SclIsChar(*pTemp) != 0) {
            *pBuffer++ = *pTemp++;
            continue;
        }
        pLimit = Abc_SclFindLimit( pTemp );
        SC_CellForEachPinIn( pCell, pPin, i )
            if Abc_SclAreEqual( pPin.pName, pTemp, pLimit ) != 0 {
                *pBuffer++ = 'a' + i;
                break;
            }
        assert(i < pCell.n_inputs, "i < pCell.n_inputs");
        pTemp = pLimit;
    }
    *pBuffer++ = 0;
}
func Abc_SclTimingUpdate(pCell *SC_Cell, p *SC_Timing, Buffer string) {
    var pPin *SC_Pin; var i int32;
    SC_CellForEachPinIn( pCell, pPin, i )
        if p.related_pin && !(int32(strings.Compare(p.related_pin, pPin.pName)) != 0) {
            ABC_FREE( p.related_pin );
            Buffer = fmt.Sprintf("%c", 'a'+i);
            p.related_pin = Abc_UtilStrsav( Buffer );
        }
}
func Abc_SclTimingsUpdate(pCell *SC_Cell, p *SC_Timings, Buffer string) {
    var pTemp *SC_Timing; var i int32;
    Vec_PtrForEachEntry( SC_Timing *, &p.vTimings, pTemp, i )
        Abc_SclTimingUpdate( pCell, pTemp, Buffer );
}
func Abc_SclPinUpdate(pCell *SC_Cell, p *SC_Pin, Buffer string) {
    // update pin names in the timing
    var pTemp *SC_Timings; var i int32;
    SC_PinForEachRTiming( p, pTemp, i )
    {
        var pPin *SC_Pin; var k int32;
        Abc_SclTimingsUpdate( pCell, pTemp, Buffer );
        SC_CellForEachPinIn( pCell, pPin, k )
            if pTemp.pName && !(int32(strings.Compare(pTemp.pName, pPin.pName)) != 0) {
                ABC_FREE( pTemp.pName );
                Buffer = fmt.Sprintf("%c", 'a'+k);
                pTemp.pName = Abc_UtilStrsav( Buffer );
            }
    }
    // update formula
    Abc_SclShortFormula( pCell, p.func_text, Buffer );
    ABC_FREE( p.func_text );
    p.func_text = Abc_UtilStrsav( Buffer );
}
func Abc_SclShortNames(p *SC_Lib) {
    var Buffer string // was [10000]byte in C

    var pClass, pCell *SC_Cell; var pPin *SC_Pin;
    var i int32;
    var k int32;
    var n int32;
    var nClasses int32 = Abc_SclLibClassNum(p);
    var nDigits int32  = Abc_Base10Log( nClasses );
    // itereate through classes
    SC_LibForEachCellClass( p, pClass, i )
    {
        var nDigits2 int32  = Abc_Base10Log( Abc_SclClassCellNum(pClass) );
        SC_RingForEachCell( pClass, pCell, k )
        {
            ABC_FREE( pCell.pName );
            Buffer = fmt.Sprintf("g%0*d_%0*d", nDigits, i, nDigits2, k);
            pCell.pName = Abc_UtilStrsav( Buffer );
            // formula
            SC_CellForEachPinOut( pCell, pPin, n )
                Abc_SclPinUpdate( pCell, pPin, Buffer );
            // pin names
            SC_CellForEachPinIn( pCell, pPin, n )
            {
                ABC_FREE( pPin.pName );
                Buffer = fmt.Sprintf("%c", 'a'+n);
                pPin.pName = Abc_UtilStrsav( Buffer );
            }
            SC_CellForEachPinOut( pCell, pPin, n )
            {
                ABC_FREE( pPin.pName );
                Buffer = fmt.Sprintf("%c", 'z'-n+pCell.n_inputs);
                pPin.pName = Abc_UtilStrsav( Buffer );
            }
        }
    }
    p.nBins = 0;
    ABC_FREE( p.pBins );
    Abc_SclHashCells( p );
    // update library name
    fmt.Printf("Renaming library \"%s\" into \"%s%d\".\n", p.pName, "lib", SC_LibCellNum(p));
    ABC_FREE( p.pName );
    Buffer = fmt.Sprintf("lib%d", SC_LibCellNum(p));
    p.pName = Abc_UtilStrsav( Buffer );
}
/**Function*************************************************************

  Synopsis    [Links equal gates into rings while sorting them by area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclCompareCells(pp1 **SC_Cell, pp2 **SC_Cell) int32 {
    if (*pp1).n_inputs < (*pp2).n_inputs {
        return -1; }
    if (*pp1).n_inputs > (*pp2).n_inputs {
        return 1; }
//    if ( (*pp1)->area < (*pp2)->area )
//        return -1;
//    if ( (*pp1)->area > (*pp2)->area )
//        return 1;
    if SC_CellPinCapAve(*pp1) < SC_CellPinCapAve(*pp2) {
        return -1; }
    if SC_CellPinCapAve(*pp1) > SC_CellPinCapAve(*pp2) {
        return 1; }
    return int32(strings.Compare((*pp1).pName, (*pp2).pName));
}
func Abc_SclLinkCells(p *SC_Lib) {
    var vList *Vec_Ptr_t;
    var pCell *SC_Cell;
    var pRepr *SC_Cell = nil;
    var i, k int32;
    assert(Vec_PtrSize(&p.vCellClasses) == 0, "Vec_PtrSize(&p.vCellClasses) == 0");
    SC_LibForEachCell( p, pCell, i )
    {
        // find gate with the same function
        SC_LibForEachCellClass( p, pRepr, k )
            if ( pCell.n_inputs  == pRepr.n_inputs && 
                 pCell.n_outputs == pRepr.n_outputs && 
                 Vec_WrdEqual(SC_CellFunc(pCell), SC_CellFunc(pRepr)) )
                break;
        if k == Vec_PtrSize(&p.vCellClasses) {
            Vec_PtrPush( &p.vCellClasses, pCell );
            pCell.pNext = pCell.pPrev = pCell;
            continue;
        }
        // add it to the list before the cell
        pRepr.pPrev.pNext = pCell; pCell.pNext = pRepr;
        pCell.pPrev = pRepr.pPrev; pRepr.pPrev = pCell;
    }
    // sort cells by size then by name
    qsort( (void *)Vec_PtrArray(&p.vCellClasses), Vec_PtrSize(&p.vCellClasses), sizeof(void *), (int(*)(void *,void *))Abc_SclCompareCells );
    // sort cell lists
    vList = Vec_PtrAlloc( 100 );
    SC_LibForEachCellClass( p, pRepr, k )
    {
        Vec_PtrClear( vList );
        SC_RingForEachCell( pRepr, pCell, i )
            Vec_PtrPush( vList, pCell );
        qsort( (void *)Vec_PtrArray(vList), (size_t)Vec_PtrSize(vList), sizeof(void *), (int(*)(void *,void *))Abc_SclCompareCells );
        // create new representative
        pRepr = (SC_Cell *)Vec_PtrEntry( vList, 0 );
        pRepr.pNext = pRepr.pPrev = pRepr;
        pRepr.pRepr = pRepr;
        pRepr.pAve  = (SC_Cell *)Vec_PtrEntry( vList, Vec_PtrSize(vList)/2 );
        pRepr.Order = 0;
        pRepr.nGates = Vec_PtrSize(vList);
        // relink cells
        Vec_PtrForEachEntryStart( SC_Cell *, vList, pCell, i, 1 )
        {
            pRepr.pPrev.pNext = pCell; pCell.pNext = pRepr;
            pCell.pPrev = pRepr.pPrev; pRepr.pPrev = pCell;
            pCell.pRepr = pRepr;
            pCell.pAve  = (SC_Cell *)Vec_PtrEntry( vList, Vec_PtrSize(vList)/2 );
            pCell.Order = i;
            pCell.nGates = Vec_PtrSize(vList);
        }
        // update list
        Vec_PtrWriteEntry( &p.vCellClasses, k, pRepr );
    }
    Vec_PtrFree( vList );
}

/**Function*************************************************************

  Synopsis    [Returns the largest inverter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindInvertor(p *SC_Lib, fFindBuff int32) *SC_Cell {
    var pCell *SC_Cell  = nil;
    var Truth word;
    if fFindBuff != 0 {
        Truth = ABC_CONST(0xAAAAAAAAAAAAAAAA);
    } else {
        Truth = ABC_CONST(0x5555555555555555);
    }
    var k int32;
    SC_LibForEachCellClass( p, pCell, k )
        if ( pCell.n_inputs == 1 && Vec_WrdEntry(&SC_CellPin(pCell, 1).vFunc, 0) == Truth )
            break;
    // take representative
    if pCell != nil {
        return pCell.pRepr;
    } else {
        return nil;
    }
}
func Abc_SclFindSmallestGate(p *SC_Cell, CinMin float32) *SC_Cell {
    var pRes *SC_Cell  = nil;
    var i int32;
    SC_RingForEachCell( p.pRepr, pRes, i )
        if ( SC_CellPinCapAve(pRes) > CinMin )
            return pRes;
    // take the largest gate
    return p.pRepr.pPrev;
}

/**Function*************************************************************

  Synopsis    [Returns the wireload model for the given area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFetchWireLoadModel(p *SC_Lib, pWLoadUsed string) *SC_WireLoad {
    var pWL *SC_WireLoad  = nil;
    var i int32;
    // Get the actual table and reformat it for 'wire_cap' output:
    assert(pWLoadUsed != nil, "pWLoadUsed != nil");
    SC_LibForEachWireLoad( p, pWL, i )
        if ( !int32(strings.Compare(pWL.pName, pWLoadUsed)) )
            break;
    if i == Vec_PtrSize(&p.vWireLoads) {
        Abc_Print( -1, "Cannot find wire load model \"%s\".\n", pWLoadUsed );
        os.Exit(int(1));
    }
//    printf( "Using wireload model \"%s\".\n", pWL->pName );
    return pWL;
}
func Abc_SclFindWireLoadModel(p *SC_Lib, Area float32) *SC_WireLoad {
    var pWLoadUsed string  = nil;
    var i int32;
    if p.default_wire_load_sel && len(p.default_wire_load_sel) {
        var pWLS *SC_WireLoadSel  = nil;
        SC_LibForEachWireLoadSel( p, pWLS, i )
            if ( !int32(strings.Compare(pWLS.pName, p.default_wire_load_sel)) )
                break;
        if i == Vec_PtrSize(&p.vWireLoadSels) {
            Abc_Print( -1, "Cannot find wire load selection model \"%s\".\n", p.default_wire_load_sel );
            os.Exit(int(1));
        }
        for  i = 0; i < Vec_FltSize(&pWLS.vAreaFrom); i++ {
            if Area >= Vec_FltEntry(&pWLS.vAreaFrom, i) && Area <  Vec_FltEntry(&pWLS.vAreaTo, i) {
                pWLoadUsed = (char *)Vec_PtrEntry(&pWLS.vWireLoadModel, i);
                break;
            }
        }
        if i == Vec_FltSize(&pWLS.vAreaFrom) {
            pWLoadUsed = (char *)Vec_PtrEntryLast(&pWLS.vWireLoadModel); }
    
    } else if p.default_wire_load && len(p.default_wire_load) {
        pWLoadUsed = p.default_wire_load;
    } else {
//        Abc_Print( 0, "No wire model given.\n" );
        return nil;
    }
    return Abc_SclFetchWireLoadModel( p, pWLoadUsed );
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the library has delay info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclHasDelayInfo(pScl any) int32 {
    var p *SC_Lib  = (SC_Lib *)pScl;
    var pCell *SC_Cell;
    var pTime *SC_Timing;
    pCell = Abc_SclFindInvertor(p, 0);
    if pCell == nil {
        return 0; }
    pTime = Scl_CellPinTime( pCell, 0 );
    if pTime == nil {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns "average" slew.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclComputeAverageSlew(p *SC_Lib) float32 {
    var pCell *SC_Cell;
    var pTime *SC_Timing;
    var vIndex *Vec_Flt_t;
    pCell = Abc_SclFindInvertor(p, 0);
    if pCell == nil {
        return 0; }
    pTime = Scl_CellPinTime( pCell, 0 );
    if pTime == nil {
        return 0; }
    vIndex = &pTime.pCellRise.vIndex0; // slew
    return Vec_FltEntry( vIndex, Vec_FltSize(vIndex)/3 );
}

/**Function*************************************************************

  Synopsis    [Compute delay parameters of pin/cell/class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclComputeParametersPin(p *SC_Lib, pCell *SC_Cell, iPin int32, Slew float32, pLD *float32, pPD *float32) int32 {
    var Load0, Load1, Load2 SC_Pair;
    var ArrIn SC_Pair   = SC_Pair{ 0.0, 0.0 };
    var SlewIn SC_Pair  = SC_Pair{ Slew, Slew };
    var ArrOut0 SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var ArrOut1 SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var ArrOut2 SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var SlewOut SC_Pair  = SC_Pair{ 0.0, 0.0 };
    var pTime *SC_Timing  = Scl_CellPinTime( pCell, iPin );
    var vIndex *Vec_Flt_t;
    if pTime != nil {
        vIndex = &pTime.pCellRise.vIndex1;
    } else {
        vIndex = nil;
    } // capacitance
    if vIndex == nil {
        return 0; }
    // handle constant table
    if Vec_FltSize(vIndex) == 1 {
        *pLD = 0;
        *pPD = Vec_FltEntry( (Vec_Flt_t *)Vec_PtrEntry(&pTime.pCellRise.vData, 0), 0 );
        return 1;
    }
    // get load points
    Load0.rise = Load0.fall = 0.0;
    Load1.rise = Load1.fall = Vec_FltEntry( vIndex, 0 );
    Load2.rise = Load2.fall = Vec_FltEntry( vIndex, Vec_FltSize(vIndex) - 2 );
    // compute delay
    Scl_LibPinArrival( pTime, &ArrIn, &SlewIn, &Load0, &ArrOut0, &SlewOut );
    Scl_LibPinArrival( pTime, &ArrIn, &SlewIn, &Load1, &ArrOut1, &SlewOut );
    Scl_LibPinArrival( pTime, &ArrIn, &SlewIn, &Load2, &ArrOut2, &SlewOut );
    ArrOut0.rise = 0.5 * ArrOut0.rise + 0.5 * ArrOut0.fall;
    ArrOut1.rise = 0.5 * ArrOut1.rise + 0.5 * ArrOut1.fall;
    ArrOut2.rise = 0.5 * ArrOut2.rise + 0.5 * ArrOut2.fall;
    // get tangent
    *pLD = (ArrOut2.rise - ArrOut1.rise) / ((Load2.rise - Load1.rise) / SC_CellPinCap(pCell, iPin));
    // get constant
    *pPD = ArrOut0.rise;
    return 1;
}
func Abc_SclComputeParametersCell(p *SC_Lib, pCell *SC_Cell, Slew float32, pLD *float32, pPD *float32) int32 {
    var pPin *SC_Pin;
    var LD, PD, ld, pd float32;
    var i int32;
    LD = PD = ld = pd = 0;
    SC_CellForEachPinIn( pCell, pPin, i )
    {
        if !(Abc_SclComputeParametersPin( p, pCell, i, Slew, &ld, &pd ) != 0) {
            return 0; }
        LD += ld; PD += pd;
    }
    *pLD = LD / Abc_MaxInt(1, pCell.n_inputs);
    *pPD = PD / Abc_MaxInt(1, pCell.n_inputs);
    return 1;
}
func Abc_SclComputeParametersClass(p *SC_Lib, pRepr *SC_Cell, Slew float32, pLD *float32, pPD *float32) {
    var pCell *SC_Cell;
    var LD, PD, ld, pd float32;
    var i, Count int32;
    LD = PD = ld = pd = 0;
    SC_RingForEachCell( pRepr, pCell, i )
    {
        Abc_SclComputeParametersCell( p, pCell, Slew, &ld, &pd );
        LD += ld; PD += pd;
        Count++;
    }
    *pLD = LD / Abc_MaxInt(1, Count);
    *pPD = PD / Abc_MaxInt(1, Count);
}
func Abc_SclComputeParametersClassPin(p *SC_Lib, pRepr *SC_Cell, iPin int32, Slew float32, pLD *float32, pPD *float32) {
    var pCell *SC_Cell;
    var LD, PD, ld, pd float32;
    var i, Count int32;
    LD = PD = ld = pd = 0;
    SC_RingForEachCell( pRepr, pCell, i )
    {
        Abc_SclComputeParametersPin( p, pCell, iPin, Slew, &ld, &pd );
        LD += ld; PD += pd;
        Count++;
    }
    *pLD = LD / Abc_MaxInt(1, Count);
    *pPD = PD / Abc_MaxInt(1, Count);
}
func Abc_SclComputeDelayCellPin(p *SC_Lib, pCell *SC_Cell, iPin int32, Slew float32, Gain float32) float32 {
    var LD, PD float32;
    Abc_SclComputeParametersPin( p, pCell, iPin, Slew, &LD, &PD );
    return 0.01 * LD * Gain + PD;
}
func Abc_SclComputeDelayClassPin(p *SC_Lib, pRepr *SC_Cell, iPin int32, Slew float32, Gain float32) float32 {
    var pCell *SC_Cell;
    var Delay float32;
    var i, Count int32;
    SC_RingForEachCell( pRepr, pCell, i )
    {
        if pCell.fSkip { 
            continue; }
//        if ( pRepr == pCell ) // skip the first gate
//            continue;
        Delay += Abc_SclComputeDelayCellPin( p, pCell, iPin, Slew, Gain );
        Count++;
    }
    return Delay / Abc_MaxInt(1, Count);
}
func Abc_SclComputeAreaClass(pRepr *SC_Cell) float32 {
    var pCell *SC_Cell;
    var Area float32;
    var i, Count int32;
    SC_RingForEachCell( pRepr, pCell, i )
    {
        if pCell.fSkip { 
            continue; }
        Area += pCell.area;
        Count++;
    }
    return Area / Abc_MaxInt(1, Count);
}

/**Function*************************************************************

  Synopsis    [Print cells]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclMarkSkippedCells(p *SC_Lib) {
    var FileName string // was [1000]byte in C

    var Buffer [1000]byte;
    var pName string;
    var pCell *SC_Cell;
    var pFile *os.File;
    var CellId, nSkipped int32;
    FileName = fmt.Sprintf("%s.skip", p.pName);
    pFile = fopen( FileName, "rb" );
    if pFile == nil {
        return; }
    for  fgets( Buffer, 999, pFile ) != nil  {
        pName = strtok( Buffer, "\r\n\t " );
        if pName == nil {
            continue; }
        CellId = Abc_SclCellFind( p, pName );
        if CellId == -1 {
            fmt.Printf("Cannot find cell \"%s\" in the library \"%s\".\n", pName, p.pName);
            continue;
        }
        pCell = SC_LibCell( p, CellId );
        pCell.fSkip = 1;
        nSkipped++;
    }
    _ = (pFile).Close();
    fmt.Printf("Marked %d cells for skipping in the library \"%s\".\n", nSkipped, p.pName);
}
func Abc_SclPrintCells(p *SC_Lib, SlewInit float32, Gain float32, fInvOnly int32, fShort int32) {
    var pCell, pRepr *SC_Cell;
    var pPin *SC_Pin;
    var i, j, k, nLength int32;
    var Slew float32;
    if SlewInit == 0 {
        Slew = Abc_SclComputeAverageSlew(p);
    } else {
        Slew = SlewInit;
    }
    var LD, PD float32;
    assert(Vec_PtrSize(&p.vCellClasses) > 0, "Vec_PtrSize(&p.vCellClasses) > 0");
    fmt.Printf("Library \"%s\" ", p.pName);
    fmt.Printf("has %d cells in %d classes.  ", Vec_PtrSize(&p.vCells), Vec_PtrSize(&p.vCellClasses));
    if !(fShort != 0) {
        fmt.Printf("Delay estimate is based on slew %.2f ps and gain %.2f.", Slew, Gain); }
    fmt.Printf("\n");
    Abc_SclMarkSkippedCells( p );
    // find the longest name
    SC_LibForEachCellClass( p, pRepr, k )
        SC_RingForEachCell( pRepr, pCell, i )
            nLength = Abc_MaxInt( nLength, len(pRepr.pName) );
    // print cells
    SC_LibForEachCellClass( p, pRepr, k )
    {
        if (fInvOnly != 0) && pRepr.n_inputs != 1 {
            continue; }
        SC_CellForEachPinOut( pRepr, pPin, i )
        {
            if i == pRepr.n_inputs {
                fmt.Printf("Class%4d : ", k);
                fmt.Printf("Cells =%3d   ", Abc_SclClassCellNum(pRepr));
                fmt.Printf("Ins =%2d  ", pRepr.n_inputs);
                fmt.Printf("Outs =%2d  ", pRepr.n_outputs);
            
            } else {
                fmt.Printf("                                            "); }
            if pPin.func_text {
                fmt.Printf("%-30s", pPin.func_text); }
            fmt.Printf("    ");
            Kit_DsdPrintFromTruth( (unsigned *)Vec_WrdArray(&pPin.vFunc), pRepr.n_inputs );
            fmt.Printf("\n");
            if fShort != 0 {
                continue; }
            SC_RingForEachCell( pRepr, pCell, j )
            {
                fmt.Printf("  %3d ", j+1);
                fmt.Printf("%s", pCell.fSkip ? "s" : " ");
                fmt.Printf(" : ");
                fmt.Printf("%-*s  ", nLength, pCell.pName);
                fmt.Printf("%2d   ", pCell.drive_strength);
                fmt.Printf("A =%8.2f  ", pCell.area);
                fmt.Printf("L =%8.2f  ", pCell.leakage);
                if pCell.n_outputs == 1 {
                    if Abc_SclComputeParametersCell( p, pCell, Slew, &LD, &PD ) != 0 {
                        fmt.Printf("D =%6.1f ps  ", 0.01 * Gain * LD + PD);
                        fmt.Printf("LD =%6.1f ps  ", LD);
                        fmt.Printf("PD =%6.1f ps    ", PD);
                        fmt.Printf("C =%5.1f ff  ", SC_CellPinCapAve(pCell));
                        fmt.Printf("Cm =%5.0f ff    ", SC_CellPin(pCell, pCell.n_inputs).max_out_cap);
                        fmt.Printf("Sm =%5.1f ps ", SC_CellPin(pCell, pCell.n_inputs).max_out_slew);
                    }
                }
                fmt.Printf("\n");
            }
            break;
        }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclConvertLeakageIntoArea(p *SC_Lib, A float32, B float32) {
    var pCell *SC_Cell; var i int32;
    SC_LibForEachCell( p, pCell, i )
        pCell.area = A * pCell.area + B * pCell.leakage;
}


/**Function*************************************************************

  Synopsis    [Print cells]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclLibNormalizeSurface(p *SC_Surface, Time float32, Load float32) {
    var vArray *Vec_Flt_t;
    var i, k int32; var Entry float32;
    Vec_FltForEachEntry( &p.vIndex0, Entry, i ) // slew
        Vec_FltWriteEntry( &p.vIndex0, i, Time * Entry );
    Vec_FltForEachEntry( &p.vIndex1, Entry, i ) // load
        Vec_FltWriteEntry( &p.vIndex1, i, Load * Entry );
    Vec_PtrForEachEntry( Vec_Flt_t *, &p.vData, vArray, k )
        Vec_FltForEachEntry( vArray, Entry, i ) // delay/slew
            Vec_FltWriteEntry( vArray, i, Time * Entry );
}
func Abc_SclLibNormalize(p *SC_Lib) {
    var pWL *SC_WireLoad;
    var pCell *SC_Cell;
    var pPin *SC_Pin;
    var pTimings *SC_Timings;
    var pTiming *SC_Timing;
    var i, k, m, n int32;
    var Time float32  = 1.0 * math.Pow(float64(10.0), float64(12 - p.unit_time));
    var Load float32  = p.unit_cap_fst * math.Pow(float64(10.0), float64(15 - p.unit_cap_snd));
    if Time == 1 && Load == 1 {
        return; }
    p.unit_time = 12;
    p.unit_cap_fst = 1;
    p.unit_cap_snd = 15;
    p.default_max_out_slew *= Time;
    SC_LibForEachWireLoad( p, pWL, i )
        pWL.cap *= Load;
    SC_LibForEachCell( p, pCell, i )
    SC_CellForEachPin( pCell, pPin, k )
    {
        pPin.cap *= Load;
        pPin.rise_cap *= Load;
        pPin.fall_cap *= Load;
        pPin.max_out_cap *= Load;
        pPin.max_out_slew *= Time;
        SC_PinForEachRTiming( pPin, pTimings, m )
        Vec_PtrForEachEntry( SC_Timing *, &pTimings.vTimings, pTiming, n )
        {
            Abc_SclLibNormalizeSurface( &pTiming.pCellRise, Time, Load );
            Abc_SclLibNormalizeSurface( &pTiming.pCellFall, Time, Load );
            Abc_SclLibNormalizeSurface( &pTiming.pRiseTrans, Time, Load );
            Abc_SclLibNormalizeSurface( &pTiming.pFallTrans, Time, Load );
        }
    }
}

/**Function*************************************************************

  Synopsis    [Derives simple GENLIB library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclProduceGenlibStrSimple(p *SC_Lib) *Vec_Str_t {
    var Buffer string // was [200]byte in C

    var vStr *Vec_Str_t;
    var pCell *SC_Cell;
    var pPin, pPinOut *SC_Pin;
    var i int32;
    var j int32;
    var k int32;
    var Count int32 = 2;
    // mark skipped cells
//    Abc_SclMarkSkippedCells( p );
    vStr = Vec_StrAlloc( 1000 );
    Vec_StrPrintStr( vStr, "GATE _const0_            0.00 z=CONST0;\n" );
    Vec_StrPrintStr( vStr, "GATE _const1_            0.00 z=CONST1;\n" );
    SC_LibForEachCell( p, pCell, i )
    {
        if pCell.n_inputs == 0 {
            continue; }
        assert(len(pCell.pName) < 200, "strlen(pCell.pName) < 200");
        SC_CellForEachPinOut( pCell, pPinOut, j )
        {
            Vec_StrPrintStr( vStr, "GATE " );
            Buffer = fmt.Sprintf("%-16s", pCell.pName);
            Vec_StrPrintStr( vStr, Buffer );
            Vec_StrPrintStr( vStr, " " );
            Buffer = fmt.Sprintf("%7.2f", pCell.area);
            Vec_StrPrintStr( vStr, Buffer );
            Vec_StrPrintStr( vStr, " " );
            Vec_StrPrintStr( vStr, pPinOut.pName );
            Vec_StrPrintStr( vStr, "=" );
            Vec_StrPrintStr( vStr, pPinOut.func_text ? pPinOut.func_text : "?" );
            Vec_StrPrintStr( vStr, ";\n" );
            SC_CellForEachPinIn( pCell, pPin, k )
            {
                Vec_StrPrintStr( vStr, "         PIN " );
                Buffer = fmt.Sprintf("%-4s", pPin.pName);
                Vec_StrPrintStr( vStr, Buffer );
                Buffer = fmt.Sprintf(" UNKNOWN  1  999  1.00  0.00  1.00  0.00\n");
                Vec_StrPrintStr( vStr, Buffer );
            }
            Count++;
        }
    }
    Vec_StrPrintStr( vStr, "\n.end\n" );
    Vec_StrPush( vStr, '\0' );
//    printf( "GENLIB library with %d gates is produced:\n", Count );
//    printf( "%s", Vec_StrArray(vStr) );
    return vStr;
}
func Abc_SclDeriveGenlibSimple(pScl any) *Mio_Library_t {
    var p *SC_Lib  = (SC_Lib *)pScl;
    var vStr *Vec_Str_t  = Abc_SclProduceGenlibStrSimple( p );
    var pLib *Mio_Library_t  = Mio_LibraryRead( p.pFileName, Vec_StrArray(vStr), nil, 0 );  
    Vec_StrFree( vStr );
    if pLib {
        fmt.Printf("Derived GENLIB library \"%s\" with %d gates.\n", p.pName, SC_LibCellNum(p));
    } else {
        fmt.Printf("Reading library has filed.\n"); }
    return pLib;
}


/**Function*************************************************************

  Synopsis    [Derive GENLIB library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclProduceGenlibStr(p *SC_Lib, Slew float32, Gain float32, nGatesMin int32, pnCellCount *int32) *Vec_Str_t {
    var Buffer string // was [200]byte in C

    var vStr *Vec_Str_t;
    var pRepr *SC_Cell;
    var pPin *SC_Pin;
    var i int32;
    var k int32;
    var Count int32 = 2;
    var nClassMax int32;
    // find the largest number of cells in a class
    SC_LibForEachCellClass( p, pRepr, i )
        if ( pRepr.n_outputs == 1 )
            nClassMax = Abc_MaxInt( nClassMax, Abc_SclClassCellNum(pRepr) );
    // update the number
    if (nGatesMin != 0) && nGatesMin >= nClassMax {
        nGatesMin = 0; }
    // mark skipped cells
    Abc_SclMarkSkippedCells( p );
    vStr = Vec_StrAlloc( 1000 );
    Vec_StrPrintStr( vStr, "GATE _const0_            0.00 z=CONST0;\n" );
    Vec_StrPrintStr( vStr, "GATE _const1_            0.00 z=CONST1;\n" );
    SC_LibForEachCellClass( p, pRepr, i )
    {
        if pRepr.n_inputs == 0 {
            continue; }
        if pRepr.n_outputs > 1 {
            continue; }
        if (nGatesMin != 0) && pRepr.n_inputs > 2 && Abc_SclClassCellNum(pRepr) < nGatesMin {
            continue; }
        assert(len(pRepr.pName) < 200, "strlen(pRepr.pName) < 200");
        Vec_StrPrintStr( vStr, "GATE " );
        Buffer = fmt.Sprintf("%-16s", pRepr.pName);
        Vec_StrPrintStr( vStr, Buffer );
        Vec_StrPrintStr( vStr, " " );
//        sprintf( Buffer, "%7.2f", Abc_SclComputeAreaClass(pRepr) );
        Buffer = fmt.Sprintf("%7.2f", pRepr.area);
        Vec_StrPrintStr( vStr, Buffer );
        Vec_StrPrintStr( vStr, " " );
        Vec_StrPrintStr( vStr, SC_CellPinName(pRepr, pRepr.n_inputs) );
        Vec_StrPrintStr( vStr, "=" );
        Vec_StrPrintStr( vStr, SC_CellPinOutFunc(pRepr, 0) ? SC_CellPinOutFunc(pRepr, 0) : "?" );
        Vec_StrPrintStr( vStr, ";\n" );
        SC_CellForEachPinIn( pRepr, pPin, k )
        {
            var Delay float32  = Abc_SclComputeDelayClassPin( p, pRepr, k, Slew, Gain );
            assert(Delay > 0, "Delay > 0");
            Vec_StrPrintStr( vStr, "         PIN " );
            Buffer = fmt.Sprintf("%-4s", pPin.pName);
            Vec_StrPrintStr( vStr, Buffer );
            Buffer = fmt.Sprintf(" UNKNOWN  1  999  %7.2f  0.00  %7.2f  0.00\n", Delay, Delay);
            Vec_StrPrintStr( vStr, Buffer );
        }
        Count++;
    }
    Vec_StrPrintStr( vStr, "\n.end\n" );
    Vec_StrPush( vStr, '\0' );
//    printf( "GENLIB library with %d gates is produced:\n", Count );
//    printf( "%s", Vec_StrArray(vStr) );
    if pnCellCount != nil {
        *pnCellCount = Count; }
    return vStr;
}
func Abc_SclProduceGenlibStrProfile(p *SC_Lib, pLib *Mio_Library_t, Slew float32, Gain float32, nGatesMin int32, pnCellCount *int32) *Vec_Str_t {
    var Buffer string // was [200]byte in C

    var vStr *Vec_Str_t;
    var pRepr *SC_Cell;
    var pPin *SC_Pin;
    var i int32;
    var k int32;
    var Count int32 = 2;
    var nClassMax int32;
    // find the largest number of cells in a class
    SC_LibForEachCellClass( p, pRepr, i )
        if ( pRepr.n_outputs == 1 )
            nClassMax = Abc_MaxInt( nClassMax, Abc_SclClassCellNum(pRepr) );
    // update the number
    if (nGatesMin != 0) && nGatesMin >= nClassMax {
        nGatesMin = 0; }
    // mark skipped cells
    Abc_SclMarkSkippedCells( p );
    vStr = Vec_StrAlloc( 1000 );
    Vec_StrPrintStr( vStr, "GATE _const0_            0.00 z=CONST0;\n" );
    Vec_StrPrintStr( vStr, "GATE _const1_            0.00 z=CONST1;\n" );
    SC_LibForEachCell( p, pRepr, i )
    {
        if pRepr.n_inputs == 0 {
            continue; }
        if pRepr.n_outputs > 1 {
            continue; }
        if (nGatesMin != 0) && pRepr.n_inputs > 2 && Abc_SclClassCellNum(pRepr) < nGatesMin {
            continue; }
        // check if the gate is in the profile
        if pRepr.n_inputs > 1 {
            var pGate *Mio_Gate_t  = Mio_LibraryReadGateByName( pLib, pRepr.pName, nil );
            if pGate == nil || Mio_GateReadProfile(pGate) == 0 {
                continue; }
        }
        // process gate
        assert(len(pRepr.pName) < 200, "strlen(pRepr.pName) < 200");
        Vec_StrPrintStr( vStr, "GATE " );
        Buffer = fmt.Sprintf("%-16s", pRepr.pName);
        Vec_StrPrintStr( vStr, Buffer );
        Vec_StrPrintStr( vStr, " " );
//        sprintf( Buffer, "%7.2f", Abc_SclComputeAreaClass(pRepr) );
        Buffer = fmt.Sprintf("%7.2f", pRepr.area);
        Vec_StrPrintStr( vStr, Buffer );
        Vec_StrPrintStr( vStr, " " );
        Vec_StrPrintStr( vStr, SC_CellPinName(pRepr, pRepr.n_inputs) );
        Vec_StrPrintStr( vStr, "=" );
        Vec_StrPrintStr( vStr, SC_CellPinOutFunc(pRepr, 0) ? SC_CellPinOutFunc(pRepr, 0) : "?" );
        Vec_StrPrintStr( vStr, ";\n" );
        SC_CellForEachPinIn( pRepr, pPin, k )
        {
            var Delay float32  = Abc_SclComputeDelayClassPin( p, pRepr, k, Slew, Gain );
            assert(Delay > 0, "Delay > 0");
            Vec_StrPrintStr( vStr, "         PIN " );
            Buffer = fmt.Sprintf("%-4s", pPin.pName);
            Vec_StrPrintStr( vStr, Buffer );
            Buffer = fmt.Sprintf(" UNKNOWN  1  999  %7.2f  0.00  %7.2f  0.00\n", Delay, Delay);
            Vec_StrPrintStr( vStr, Buffer );
        }
        Count++;
    }
    Vec_StrPrintStr( vStr, "\n.end\n" );
    Vec_StrPush( vStr, '\0' );
//    printf( "GENLIB library with %d gates is produced:\n", Count );
//    printf( "%s", Vec_StrArray(vStr) );
    if pnCellCount != nil {
        *pnCellCount = Count; }
    return vStr;
}
func Abc_SclDumpGenlib(pFileName string, p *SC_Lib, SlewInit float32, Gain float32, nGatesMin int32) {
    var nCellCount int32;
    var FileName string // was [1000]byte in C

    var Slew float32;
    if SlewInit == 0 {
        Slew = Abc_SclComputeAverageSlew(p);
    } else {
        Slew = SlewInit;
    }
    var vStr *Vec_Str_t;
    var pFile *os.File;
    if pFileName == nil {
        FileName = fmt.Sprintf("%s_s%03d_g%03d_m%d.genlib", p.pName, int32(Slew), int32(Gain), nGatesMin);
    } else {
        FileName = fmt.Sprintf("%s", pFileName); }
    pFile = fopen( FileName, "wb" );
    if pFile == nil {
        printf( "Cannot open file \"%s\" for writing.\n", FileName );
        return;
    }
    vStr = Abc_SclProduceGenlibStr( p, Slew, Gain, nGatesMin, &nCellCount );
    fmt.Fprintf(pFile, "%s", Vec_StrArray(vStr));
    Vec_StrFree( vStr );
    _ = (pFile).Close();
    printf( "Written GENLIB library with %d gates into file \"%s\".\n", nCellCount, FileName );
}
func Abc_SclDeriveGenlib(pScl any, pMio any, SlewInit float32, Gain float32, nGatesMin int32, fVerbose int32) *Mio_Library_t {
    var nCellCount int32;
    var p *SC_Lib  = (SC_Lib *)pScl;
    var Slew float32;
    if SlewInit == 0 {
        Slew = Abc_SclComputeAverageSlew(p);
    } else {
        Slew = SlewInit;
    }
    var vStr *Vec_Str_t;
    var pLib *Mio_Library_t;
    if pMio == nil {
        vStr = Abc_SclProduceGenlibStr( p, Slew, Gain, nGatesMin, &nCellCount );
    } else {
        vStr = Abc_SclProduceGenlibStrProfile( p, (Mio_Library_t *)pMio, Slew, Gain, nGatesMin, &nCellCount ); }
    pLib = Mio_LibraryRead( p.pFileName, Vec_StrArray(vStr), nil, 0 );  
    Vec_StrFree( vStr );
    if !pLib {
        fmt.Printf("Reading library has filed.\n");
    } else if fVerbose != 0 {
        fmt.Printf("Derived GENLIB library \"%s\" with %d gates using slew %.2f ps and gain %.2f.\n", p.pName, nCellCount, Slew, Gain); }
    return pLib;
}

/**Function*************************************************************

  Synopsis    [Install library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclInstallGenlib(pScl any, SlewInit float32, Gain float32, nGatesMin int32) {
    var p *SC_Lib  = (SC_Lib *)pScl;
    var vStr, vStr2 *Vec_Str_t;
    var Slew float32;
    if SlewInit == 0 {
        Slew = Abc_SclComputeAverageSlew(p);
    } else {
        Slew = SlewInit;
    }
    var RetValue int32;
    var nGateCount int32 = SC_LibCellNum(p);
    if Gain == 0 {
        vStr = Abc_SclProduceGenlibStrSimple(p);
    } else {
        vStr = Abc_SclProduceGenlibStr( p, Slew, Gain, nGatesMin, &nGateCount ); }
    vStr2 = Vec_StrDup( vStr );
    RetValue = Mio_UpdateGenlib2( vStr, vStr2, p.pName, 0 );
    Vec_StrFree( vStr );
    Vec_StrFree( vStr2 );
    if !(RetValue != 0) {
        fmt.Printf("Reading library has filed.\n");
    } else if Gain != 0 {
        fmt.Printf("Derived GENLIB library \"%s\" with %d gates using slew %.2f ps and gain %.2f.\n", p.pName, nGateCount, Slew, Gain); }
//    else
//        printf( "Derived unit-delay GENLIB library \"%s\" with %d gates.\n", p->pName, nGateCount );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

