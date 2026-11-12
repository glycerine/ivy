package scl

import (
	"fmt"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [sclUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Standard-cell library representation.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko, Niklas Een]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - August 24, 2012.]

  Revision    [$Id: sclUtil.c,v 1.0 2012/08/24 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sclSize.h"
//#include "map/mio/mio.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Converts pNode->pData gates into array of SC_Lit gate IDs and back.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclMioGates2SclGates(pLib *SC_Lib, p *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var i, gateId, bufferId int32;
    // find buffer
    if Mio_LibraryReadBuf((Mio_Library_t *)p.pManFunc) == nil {
        fmt.Printf("Cannot find buffer in the current library. Quitting.\n");
        return;
    }
    bufferId = Abc_SclCellFind( pLib, Mio_GateReadName(Mio_LibraryReadBuf((Mio_Library_t *)p.pManFunc)) );
    assert(bufferId >= 0, "bufferId >= 0");
    // remap cells
    assert(p.vGates == nil, "p.vGates == nil");
    p.vGates = Vec_IntStartFull( Abc_NtkObjNumMax(p) );
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        gateId = Abc_SclCellFind( pLib, Mio_GateReadName((Mio_Gate_t *)pObj.pData) );
        assert(gateId >= 0, "gateId >= 0");
        Vec_IntWriteEntry( p.vGates, i, gateId );
    }
    }
    p.pSCLib = pLib;
}
func Abc_SclSclGates2MioGates(pLib *SC_Lib, p *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var pCell *SC_Cell;
    var i, Counter, CounterAll int32;
    assert(p.vGates != nil, "p.vGates != nil");
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        pCell = Abc_SclObjCell(pObj);
        assert(pCell.n_inputs == Abc_ObjFaninNum(pObj), "pCell.n_inputs == Abc_ObjFaninNum(pObj)");
        pObj.pData = Mio_LibraryReadGateByName( (Mio_Library_t *)p.pManFunc, pCell.pName, nil );
        Counter += (pObj.pData == nil);
        assert(pObj.fMarkA == 0 && pObj.fMarkB == 0, "pObj.fMarkA == 0 && pObj.fMarkB == 0");
        CounterAll++;
    }
    }
    if Counter != 0 {
        fmt.Printf("Could not find %d (out of %d) gates in the current library.\n", Counter, CounterAll); }
    Vec_IntFreeP( &p.vGates );
    p.pSCLib = nil;
}

/**Function*************************************************************

  Synopsis    [Transfer gate sizes from AIG without barbufs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclTransferGates(pOld *Abc_Ntk_t, pNew *Abc_Ntk_t) {
    var pObj *Abc_Obj_t; var i int32;
    assert(pOld.nBarBufs2 > 0, "pOld.nBarBufs2 > 0");
    assert(pNew.nBarBufs2 == 0, "pNew.nBarBufs2 == 0");
    // inlining of macro Abc_NtkForEachNodeNotBarBuf
    for  i = 0; (i < Vec_PtrSize((pOld).vObjs)) && (((pObj) = Abc_NtkObj(pOld, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        if pObj.pCopy == nil {
            continue; }
        assert(Abc_ObjNtk(pObj.pCopy) == pNew, "Abc_ObjNtk(pObj.pCopy) == pNew");
        pObj.pData = pObj.pCopy.pData;
    }
    }
}

/**Function*************************************************************

  Synopsis    [Reports percentage of gates of each size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
const ABC_SCL_MAX_SIZE = 64
func Abc_SclManPrintGateSizes(pLib *SC_Lib, p *Abc_Ntk_t, vGates *Vec_Int_t) {
    var pObj *Abc_Obj_t;
    var pCell *SC_Cell;
    var i int32;
    var nGates int32;
    var Counters [ABC_SCL_MAX_SIZE]int32 ;
    var TotArea float64;
    var Areas [ABC_SCL_MAX_SIZE]float64 ;
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        pCell = SC_LibCell( pLib, Vec_IntEntry(vGates, Abc_ObjId(pObj)) );
        assert(pCell.Order < ABC_SCL_MAX_SIZE, "pCell.Order < ABC_SCL_MAX_SIZE");
        Counters[pCell.Order]++;
        Areas[pCell.Order] += pCell.area;
        TotArea += pCell.area;
        nGates++;
    }
    }
    fmt.Printf("Total gates = %d.  Total area = %.1f\n", nGates, TotArea);
    for  i = 0; i < ABC_SCL_MAX_SIZE; i++  {
        if Counters[i] == 0 {
            continue; }
        fmt.Printf("Cell size = %d.  ", i);
        fmt.Printf("Count = %6d  ", Counters[i]);
        fmt.Printf("(%5.1f %%)   ", 100.0 * Counters[i] / nGates);
        fmt.Printf("Area = %12.1f  ", Areas[i]);
        fmt.Printf("(%5.1f %%)  ", 100.0 * Areas[i] / TotArea);
        fmt.Printf("\n");
    }
}
func Abc_SclPrintGateSizes(pLib *SC_Lib, p *Abc_Ntk_t) {
    Abc_SclMioGates2SclGates( pLib, p );
    Abc_SclManPrintGateSizes( pLib, p, p.vGates );
    Abc_SclSclGates2MioGates( pLib, p );
    Vec_IntFreeP( &p.vGates );
    p.pSCLib = nil;
}

/**Function*************************************************************

  Synopsis    [Downsizes each gate to its minimium size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclFindMaxAreaCell(pRepr *SC_Cell) *SC_Cell {
    var pCell *SC_Cell;
    var pBest *SC_Cell = pRepr;
    var AreaBest float32  = pRepr.area;
    var i int32;
    // inlining of macro SC_RingForEachCell
    for  i, pCell = 0, pRepr; i == 0 || pCell != pRepr; pCell, i = pCell.pNext, i+1  {
        if AreaBest < pCell.area {
            AreaBest = pCell.area;
            pBest = pCell;
        }
    }
    return pBest;
}
func Abc_SclFindMinAreas(pLib *SC_Lib, fUseMax int32) *Vec_Int_t {
    var vMinCells *Vec_Int_t;
    var pCell *SC_Cell;
    var pRepr *SC_Cell = nil;
    var pBest *SC_Cell = nil;
    var i, k int32;
    // map each gate in the library into its min/max-size prototype
    vMinCells = Vec_IntStartFull( Vec_PtrSize(&pLib.vCells) );
    // inlining of macro SC_LibForEachCellClass
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(&pLib.vCellClasses)) && (((pRepr) = (SC_Cell *)Vec_PtrEntry(&pLib.vCellClasses, i)), 1); i++  {
        if fUseMax != 0 {
            pBest = Abc_SclFindMaxAreaCell(pRepr);
        } else {
            pBest = pRepr;
        }
        // inlining of macro SC_RingForEachCell
        for  k, pCell = 0, pRepr; k == 0 || pCell != pRepr; pCell, k = pCell.pNext, k+1  {
            Vec_IntWriteEntry( vMinCells, pCell.Id, pBest.Id );
        }
    }
    return vMinCells;
}
func Abc_SclMinsizePerform(pLib *SC_Lib, p *Abc_Ntk_t, fUseMax int32, fVerbose int32) {
    var vMinCells *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, gateId int32;
    vMinCells = Abc_SclFindMinAreas( pLib, fUseMax );
    Abc_SclMioGates2SclGates( pLib, p );
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        gateId = Vec_IntEntry( p.vGates, i );
        assert(gateId >= 0 && gateId < Vec_PtrSize(&pLib.vCells), "gateId >= 0 && gateId < Vec_PtrSize(&pLib.vCells)");
        gateId = Vec_IntEntry( vMinCells, gateId );
        assert(gateId >= 0 && gateId < Vec_PtrSize(&pLib.vCells), "gateId >= 0 && gateId < Vec_PtrSize(&pLib.vCells)");
        Vec_IntWriteEntry( p.vGates, i, gateId );
    }
    }
    Abc_SclSclGates2MioGates( pLib, p );
    Vec_IntFree( vMinCells );
}
func Abc_SclCountMinSize(pLib *SC_Lib, p *Abc_Ntk_t, fUseMax int32) int32 {
    var vMinCells *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var i, gateId, Counter int32;
    vMinCells = Abc_SclFindMinAreas( pLib, fUseMax );
    // inlining of macro Abc_NtkForEachNodeNotBarBuf1
    for  i = 0; (i < Vec_PtrSize((p).vObjs)) && (((pObj) = Abc_NtkObj(p, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) || !Abc_ObjFaninNum(pObj) || Abc_ObjIsBarBuf(pObj) {} else {
        gateId = Vec_IntEntry( p.vGates, i );
        Counter += ( gateId == Vec_IntEntry(vMinCells, gateId) );
    }
    }
    Vec_IntFree( vMinCells );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Reads timing constraints.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclReadTimingConstr(pAbc *Abc_Frame_t, pFileName string, fVerbose int32) {
    var Buffer [1000]byte;
    var pToken string;
    var pFile *os.File  = os.Open(pFileName);
    for  fgets( Buffer, 1000, pFile )  {
        pToken = strtok( Buffer, " \t\r\n" );
        if pToken == nil {
            continue; }
        if !(int32(strings.Compare(pToken, "set_driving_cell")) != 0) {
//        if ( !strcmp(pToken, "default_input_cell") )
        
            Abc_FrameSetDrivingCell( Abc_UtilStrsav(strtok(nil, " \t\r\n")) );
            if fVerbose != 0 { 
                fmt.Printf("Setting driving cell to be \"%s\".\n", Abc_FrameReadDrivingCell()); }
        
        } else if !(int32(strings.Compare(pToken, "set_load")) != 0) {
//        else if ( !strcmp(pToken, "default_output_load") )
        
            Abc_FrameSetMaxLoad( atof(strtok(nil, " \t\r\n")) );
            if fVerbose != 0 { 
                fmt.Printf("Setting output load to be %f.\n", Abc_FrameReadMaxLoad()); }
        
        } else { fmt.Printf("Unrecognized token \"%s\".\n", pToken); }
    }
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_SclExtractBarBufs(pNtk *Abc_Ntk_t) *Vec_Int_t {
    var vBufs *Vec_Int_t;
    var pBuffer *Mio_Gate_t;
    var pObj *Abc_Obj_t; var i int32;
    pBuffer = Mio_LibraryReadBuf( (Mio_Library_t *)pNtk.pManFunc );
    if pBuffer == nil {
        fmt.Printf("Cannot find buffer in the current library. Quitting.\n");
        return nil;
    }
    vBufs = Vec_IntAlloc( 100 );
    // inlining of macro Abc_NtkForEachBarBuf
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsBarBuf(pObj) {} else {
        assert(pObj.pData == nil, "pObj.pData == nil");
        pObj.pData = pBuffer;
        Vec_IntPush( vBufs, i );
    }
    }
    return vBufs;
}
func Abc_SclInsertBarBufs(pNtk *Abc_Ntk_t, vBufs *Vec_Int_t) {
    var pObj *Abc_Obj_t; var i int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vBufs) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vBufs,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.pData = nil; }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

