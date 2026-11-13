package bmc

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [bmcCexTools.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [CEX analysis and optimization toolbox.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexTools.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Bmc_CexBitMask(iBit int32) word { assert(iBit < 64, "iBit < 64"); return ~(((word)1) << iBit); }
func Bmc_CexCopySim(vSims *Vec_Wrd_t, iObjTo int32, iObjFrom int32) { Vec_WrdWriteEntry( vSims, iObjTo, iObjFrom );     }
func Bmc_CexAndSim(vSims *Vec_Wrd_t, iObjTo int32, i int32, j int32) { Vec_WrdWriteEntry( vSims, iObjTo, Vec_WrdEntry(vSims, i) & Vec_WrdEntry(vSims, j) ); }
func Bmc_CexOrSim(vSims *Vec_Wrd_t, iObjTo int32, i int32, j int32) { Vec_WrdWriteEntry( vSims, iObjTo, Vec_WrdEntry(vSims, i) | Vec_WrdEntry(vSims, j) ); }
func Bmc_CexSim(vSims *Vec_Wrd_t, iObj int32, i int32) int32 { return (Vec_WrdEntry(vSims, iObj) >> i) & 1;      }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexBitCount(p *Abc_Cex_t, nInputs int32) int32 {
    var k, Counter, Counter2 int32;
    if p == nil {
        fmt.Printf("The counter example is NULL.\n");
        return -1;
    }
    for  k = 0; k < p.nBits; k++  {
        Counter += Abc_InfoHasBit(p.pData, k);
        if (k - p.nRegs) % p.nPis < nInputs {
            Counter2 += Abc_InfoHasBit(p.pData, k); }
    }
    return Counter2;
}
/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexDumpStats(p *Gia_Man_t, pCex *Abc_Cex_t, pCexCare *Abc_Cex_t, pCexEss *Abc_Cex_t, pCexMin *Abc_Cex_t, clk abctime) {
    var nInputs int32     = Gia_ManPiNum(p);
    var nBitsTotal int32  = (pCex.iFrame + 1) * nInputs;
    var nBitsCare int32   = Bmc_CexBitCount(pCexCare, nInputs);
    var nBitsDC int32     = nBitsTotal - nBitsCare;
    var nBitsEss int32    = Bmc_CexBitCount(pCexEss, nInputs);
    var nBitsOpt int32    = nBitsCare - nBitsEss;
    var nBitsMin int32    = Bmc_CexBitCount(pCexMin, nInputs);

    var pTable *os.File   = os.OpenFile("cex/stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    fmt.Fprintf(pTable, "%s ", p.pName);
    fmt.Fprintf(pTable, "%d ", nInputs);
    fmt.Fprintf(pTable, "%d ", Gia_ManRegNum(p));
    fmt.Fprintf(pTable, "%d ", pCex.iFrame + 1);
    fmt.Fprintf(pTable, "%d ", nBitsTotal);
    fmt.Fprintf(pTable, "%.2f ", 100.0 * nBitsDC  / nBitsTotal);
    fmt.Fprintf(pTable, "%.2f ", 100.0 * nBitsOpt / nBitsTotal);
    fmt.Fprintf(pTable, "%.2f ", 100.0 * nBitsEss / nBitsTotal);
    fmt.Fprintf(pTable, "%.2f ", 100.0 * nBitsMin / nBitsTotal);
    fmt.Fprintf(pTable, "%.2f ", 1.0*clk/(int64(time.Second)));
    fmt.Fprintf(pTable, "\n");
    _ = (pTable).Close();
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexDumpAogStats(p *Gia_Man_t, clk abctime) {
    var pTable *os.File   = os.OpenFile("cex/aig_stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    fmt.Fprintf(pTable, "%s ", p.pName);
    fmt.Fprintf(pTable, "%d ", Gia_ManPiNum(p));
    fmt.Fprintf(pTable, "%d ", Gia_ManAndNum(p));
    fmt.Fprintf(pTable, "%d ", Gia_ManLevelNum(p));
    fmt.Fprintf(pTable, "%.2f ", 1.0*clk/(int64(time.Second)));
    fmt.Fprintf(pTable, "\n");
    _ = (pTable).Close();
}

/**Function*************************************************************

  Synopsis    [Performs initialized unrolling till the last frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexPerformUnrolling(p *Gia_Man_t, pCex *Abc_Cex_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var i, k int32;
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    pNew = Gia_ManStart( (pCex.iFrame + 1) * Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachRi
    for k = 0; k < Gia_ManRegNum(p); k++ {
    	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
    	if pObj == nil { break }
        pObj.Value = 0;
    }
    Gia_ManHashAlloc( pNew );
    for  i = 0; i <= pCex.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for k = 0; k < Gia_ManPiNum(p); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            pObj.Value = Gia_ManAppendCi( pNew );
        }
        // inlining of macro Gia_ManForEachRiRo
        for k = 0; k < Gia_ManRegNum(p); k++ {
        	pObjRi = Gia_ManCo(p, Gia_ManPoNum(p)+k)
        	if pObjRi == nil { break }
        	pObjRo = Gia_ManCi(p, Gia_ManPiNum(p)+k)
        	if pObjRo == nil { break }
            pObjRo.Value = pObjRi.Value;
        }
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        }
    }
    Gia_ManHashStop( pNew );
    pObj = Gia_ManPo(p, pCex.iPo);
    Gia_ManAppendCo( pNew, pObj.Value );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Bmc_CexPerformUnrollingTest(p *Gia_Man_t, pCex *Abc_Cex_t) {
    var pNew *Gia_Man_t;
    var clk abctime  = Abc_Clock();
    pNew = Bmc_CexPerformUnrolling( p, pCex );
    Gia_ManPrintStats( pNew, nil );
    Gia_AigerWrite( pNew, "unroll.aig", 0, 0, 0 );
//Bmc_CexDumpAogStats( pNew, Abc_Clock() - clk );
    Gia_ManStop( pNew );
    fmt.Printf("CE-induced network is written into file \"unroll.aig\".\n");
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}


/**Function*************************************************************

  Synopsis    [Computes CE-induced network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexBuildNetwork(p *Gia_Man_t, pCex *Abc_Cex_t) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var fCompl0, fCompl1 int32;
    var i int32;
    var k int32;
    var iBit int32 = pCex.nRegs;
    // start the manager
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( "unate" );
    // set const0
    Gia_ManConst0(p).fMark0 = 0;
    Gia_ManConst0(p).fMark1 = 1;
    Gia_ManConst0(p).Value  = ~0;
    // set init state
    // inlining of macro Gia_ManForEachRi
    for k = 0; k < Gia_ManRegNum(p); k++ {
    	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
    	if pObj == nil { break }
        pObj.fMark0 = 0;
        pObj.fMark1 = 1;
        pObj.Value  = ~0;
    }
    Gia_ManHashAlloc( pNew );
    for  i = 0; i <= pCex.iFrame; i++  {
        //  primary inputs
        // inlining of macro Gia_ManForEachPi
        for k = 0; k < Gia_ManPiNum(p); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Abc_InfoHasBit( pCex.pData, iBit++ );
            pObj.fMark1 = 0;
            pObj.Value  = Gia_ManAppendCi(pNew);
        }
        // transfer 
        // inlining of macro Gia_ManForEachRiRo
        for k = 0; k < Gia_ManRegNum(p); k++ {
        	pObjRi = Gia_ManCo(p, Gia_ManPoNum(p)+k)
        	if pObjRi == nil { break }
        	pObjRo = Gia_ManCi(p, Gia_ManPiNum(p)+k)
        	if pObjRo == nil { break }
            pObjRo.fMark0 = pObjRi.fMark0;
            pObjRo.fMark1 = pObjRi.fMark1;
            pObjRo.Value  = pObjRi.Value;
        }
        // internal nodes
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
            pObj.fMark0 = fCompl0 & fCompl1;
            if pObj.fMark0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 & Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 | Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
            } else if !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin1(pObj).fMark1;
            } else { assert(false, "0"); }
            pObj.Value = ~0;
            if pObj.fMark1 {
                continue; }
            if pObj.fMark0 {
                pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.Value = Gia_ManHashOr( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
            } else if !fCompl0 {
                pObj.Value = Gia_ObjFanin0(pObj).Value;
            } else if !(fCompl1 != 0) {
                pObj.Value = Gia_ObjFanin1(pObj).Value;
            } else { assert(false, "0"); }
            assert(pObj.Value > 0, "pObj.Value > 0");
        }
        }
        // combinational outputs
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
            pObj.Value  = Gia_ObjFanin0(pObj).Value;
        }
    }
    Gia_ManHashStop( pNew );
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // create primary output
    pObj = Gia_ManPo(p, pCex.iPo);
    assert(pObj.fMark0 == 1, "pObj.fMark0 == 1");
    assert(pObj.fMark1 == 0, "pObj.fMark1 == 0");
    assert(pObj.Value > 0, "pObj.Value > 0");
    Gia_ManAppendCo( pNew, pObj.Value );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}
func Bmc_CexBuildNetworkTest(p *Gia_Man_t, pCex *Abc_Cex_t) {
    var pNew *Gia_Man_t;
    var clk abctime  = Abc_Clock();
    pNew = Bmc_CexBuildNetwork( p, pCex );
    Gia_ManPrintStats( pNew, nil );
    Gia_AigerWrite( pNew, "unate.aig", 0, 0, 0 );
//Bmc_CexDumpAogStats( pNew, Abc_Clock() - clk );
    Gia_ManStop( pNew );
    fmt.Printf("CE-induced network is written into file \"unate.aig\".\n");
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}


/**Function*************************************************************

  Synopsis    [Prints one counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexPrint(pCex *Abc_Cex_t, nRealPis int32, fVerbose int32) {
    var i int32;
    var k int32;
    var Count int32;
    var iBit int32 = pCex.nRegs;
    Abc_CexPrintStatsInputs( pCex, nRealPis );
    if !(fVerbose != 0) {
        return; }

    for  i = 0; i <= pCex.iFrame; i++  {
        Count = 0;
        fmt.Printf("%3d : ", i);
        for  k = 0; k < nRealPis; k++  {
            Count += Abc_InfoHasBit(pCex.pData, iBit);
            fmt.Printf("%d", Abc_InfoHasBit(pCex.pData, iBit++));
        }
        fmt.Printf(" %3d ", Count);
        Count = 0;
        for       ; k < pCex.nPis; k++  {
            Count += Abc_InfoHasBit(pCex.pData, iBit);
            fmt.Printf("%d", Abc_InfoHasBit(pCex.pData, iBit++));
        }
        fmt.Printf(" %3d\n", Count);
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
}

/**Function*************************************************************

  Synopsis    [Verifies the care set of the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexVerify(p *Gia_Man_t, pCex *Abc_Cex_t, pCexCare *Abc_Cex_t) int32 {
    var pObj *Gia_Obj_t;
    var i, k int32;
//    assert( pCex->nRegs > 0 );
//    assert( pCexCare->nRegs == 0 );
    Gia_ObjTerSimSet0( Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachRi
    for k = 0; k < Gia_ManRegNum(p); k++ {
    	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
    	if pObj == nil { break }
        Gia_ObjTerSimSet0( pObj );
    }
    for  i = 0; i <= pCex.iFrame; i++  {
        // inlining of macro Gia_ManForEachPi
        for k = 0; k < Gia_ManPiNum(p); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            if !Abc_InfoHasBit( pCexCare.pData, pCexCare.nRegs + i * pCexCare.nPis + k ) {
                Gia_ObjTerSimSetX( pObj );
            } else if Abc_InfoHasBit( pCex.pData, pCex.nRegs + i * pCex.nPis + k ) {
                Gia_ObjTerSimSet1( pObj );
            } else {
                Gia_ObjTerSimSet0( pObj ); }
        }
        // inlining of macro Gia_ManForEachRo
        for k = 0; k < Gia_ManRegNum(p); k++ {
        	pObj = Gia_ManCi(p, Gia_ManPiNum(p)+k)
        	if pObj == nil { break }
            Gia_ObjTerSimRo( p, pObj );
        }
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            Gia_ObjTerSimAnd( pObj ); }
        }
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            Gia_ObjTerSimCo( pObj );
        }
    }
    pObj = Gia_ManPo( p, pCex.iPo );
    return Gia_ObjTerSimGet1(pObj);
}

/**Function*************************************************************

  Synopsis    [Computes care bits of the CEX.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexCareBits_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    var fCompl0, fCompl1 int32;
    if Gia_ObjIsConst0(pObj) {
        return; }
    if pObj.fMark1 {
        return; }
    pObj.fMark1 = 1;
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
    fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
    if pObj.fMark0 {
        assert(fCompl0 == 1 && fCompl1 == 1, "fCompl0 == 1 && fCompl1 == 1");
        Bmc_CexCareBits_rec( p, Gia_ObjFanin0(pObj) );
        Bmc_CexCareBits_rec( p, Gia_ObjFanin1(pObj) );
    
    } else {
        assert(fCompl0 == 0 || fCompl1 == 0, "fCompl0 == 0 || fCompl1 == 0");
        if fCompl0 == 0 {
            Bmc_CexCareBits_rec( p, Gia_ObjFanin0(pObj) ); }
        if fCompl1 == 0 {
            Bmc_CexCareBits_rec( p, Gia_ObjFanin1(pObj) ); }
    }
}
func Bmc_CexCareBits2_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    var fCompl0, fCompl1 int32;
    if Gia_ObjIsConst0(pObj) {
        return; }
    if pObj.fMark1 {
        return; }
    pObj.fMark1 = 1;
    if Gia_ObjIsCi(pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
    fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
    if pObj.fMark0 {
        assert(fCompl0 == 1 && fCompl1 == 1, "fCompl0 == 1 && fCompl1 == 1");
        Bmc_CexCareBits2_rec( p, Gia_ObjFanin0(pObj) );
        Bmc_CexCareBits2_rec( p, Gia_ObjFanin1(pObj) );
    
    } else {
        assert(fCompl0 == 0 || fCompl1 == 0, "fCompl0 == 0 || fCompl1 == 0");
        if fCompl0 == 0 {
            Bmc_CexCareBits2_rec( p, Gia_ObjFanin0(pObj) );
        /**/ 
        } else 
        /**/ 
        if fCompl1 == 0 {
            Bmc_CexCareBits2_rec( p, Gia_ObjFanin1(pObj) ); }
    }
}
func Bmc_CexCareBits(p *Gia_Man_t, pCexState *Abc_Cex_t, pCexImpl *Abc_Cex_t, pCexEss *Abc_Cex_t, fFindAll int32, fVerbose int32) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var pObj *Gia_Obj_t;
    var fCompl0, fCompl1 int32;
    var i, k, iBit int32;
    assert(pCexState.nRegs == 0, "pCexState.nRegs == 0");
    // start the counter-example
    pNew = Abc_CexAlloc( 0, Gia_ManCiNum(p), pCexState.iFrame + 1 );
    pNew.iFrame = pCexState.iFrame;
    pNew.iPo    = pCexState.iPo;
    // set initial state
    Gia_ManCleanMark01(p);
    // set const0
    Gia_ManConst0(p).fMark0 = 0;
    Gia_ManConst0(p).fMark1 = 1;
    for  i = pCexState.iFrame; i >= 0; i--  {
        // set correct values
        iBit = pCexState.nPis * i;
        // inlining of macro Gia_ManForEachCi
        for k = 0; k < Vec_IntSize(p.vCis); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Abc_InfoHasBit(pCexState.pData, iBit+k);
            pObj.fMark1 = 0;
            if pCexImpl != nil {
                pObj.fMark1 |= Abc_InfoHasBit(pCexImpl.pData, iBit+k); }
            if pCexEss != nil {
                pObj.fMark1 |= Abc_InfoHasBit(pCexEss.pData, iBit+k); }
        }
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
            pObj.fMark0 = fCompl0 & fCompl1;
            if pObj.fMark0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 & Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 | Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
            } else if !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin1(pObj).fMark1;
            } else { assert(false, "0"); }
        }
        }
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // move values from COs to CIs
        if i == pCexState.iFrame {
            assert(Gia_ManPo(p, pCexState.iPo).fMark0 == 1, "Gia_ManPo(p, pCexState.iPo).fMark0 == 1");
            if fFindAll != 0 {
                Bmc_CexCareBits_rec( p, Gia_ObjFanin0(Gia_ManPo(p, pCexState.iPo)) );
            } else {
                Bmc_CexCareBits2_rec( p, Gia_ObjFanin0(Gia_ManPo(p, pCexState.iPo)) ); }
        
        } else {
            // inlining of macro Gia_ManForEachRi
            for k = 0; k < Gia_ManRegNum(p); k++ {
            	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
            	if pObj == nil { break }
                if Abc_InfoHasBit(pNew.pData, pNew.nPis * (i+1) + Gia_ManPiNum(p) + k) {
                     if fFindAll != 0 {
                         Bmc_CexCareBits_rec( p, Gia_ObjFanin0(pObj) );
                     } else {
                         Bmc_CexCareBits2_rec( p, Gia_ObjFanin0(pObj) ); }
                }
            }
        }
        // save the results
        // inlining of macro Gia_ManForEachCi
        for k = 0; k < Vec_IntSize(p.vCis); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            if pObj.fMark1 {
                pObj.fMark1 = 0;
                if pCexImpl == nil || !Abc_InfoHasBit(pCexImpl.pData, pNew.nPis * i + k) {
                    Abc_InfoSetBit(pNew.pData, pNew.nPis * i + k); }
            }
        }
    }
    if pCexEss != nil {
        fmt.Printf("Minimized:    ");
    } else {
        fmt.Printf("Care bits:    "); }
    Bmc_CexPrint( pNew, Gia_ManPiNum(p), fVerbose );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Simulates one bit to check whether it is essential.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexEssentialBitOne(p *Gia_Man_t, pCexState *Abc_Cex_t, iBit int32, pCexPrev *Abc_Cex_t, pfEqual *int32) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var pObj *Gia_Obj_t;
    var i, k, fCompl0, fCompl1 int32;
    assert(pCexState.nRegs == 0, "pCexState.nRegs == 0");
    assert(iBit < pCexState.nBits, "iBit < pCexState.nBits");
    if pfEqual != nil {
        *pfEqual = 0; }
    // start the counter-example
    pNew = Abc_CexAllocFull( 0, Gia_ManCiNum(p), pCexState.iFrame + 1 );
    pNew.iFrame = pCexState.iFrame;
    pNew.iPo    = pCexState.iPo;
    // clean bit
    Abc_InfoXorBit( pNew.pData, iBit );
    // simulate the remaining frames
    Gia_ManConst0(p).fMark0 = 0;
    Gia_ManConst0(p).fMark1 = 1;
    for  i = iBit / pCexState.nPis; i <= pCexState.iFrame; i++  {
        // inlining of macro Gia_ManForEachCi
        for k = 0; k < Vec_IntSize(p.vCis); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Abc_InfoHasBit( pCexState.pData, i * pCexState.nPis + k );
            pObj.fMark1 = Abc_InfoHasBit( pNew.pData,      i * pCexState.nPis + k );
        }
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
            pObj.fMark0 = fCompl0 & fCompl1;
            if pObj.fMark0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 & Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 && !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1 | Gia_ObjFanin1(pObj).fMark1;
            } else if !fCompl0 {
                pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
            } else if !(fCompl1 != 0) {
                pObj.fMark1 = Gia_ObjFanin1(pObj).fMark1;
            } else { assert(false, "0"); }
        }
        }
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            pObj.fMark1 = Gia_ObjFanin0(pObj).fMark1;
        }
        if i < pCexState.iFrame {
            var fChanges int32;
            var fEqual int32     = (pCexPrev != nil);
            var iBitShift int32  = (i + 1) * pCexState.nPis + Gia_ManPiNum(p);
            // inlining of macro Gia_ManForEachRi
            for k = 0; k < Gia_ManRegNum(p); k++ {
            	pObj = Gia_ManCo(p, Gia_ManPoNum(p)+k)
            	if pObj == nil { break }
                if (fEqual != 0) && (pCexPrev != nil) && int32(pObj.fMark1) != Abc_InfoHasBit(pCexPrev.pData, iBitShift + k) {
                    fEqual = 0; }
                if !pObj.fMark1 {
                    fChanges = 1;
                    Abc_InfoXorBit( pNew.pData, iBitShift + k );
                }
            }
            if !(fChanges != 0) || (fEqual != 0) {
                if pfEqual != nil {
                    *pfEqual = fEqual; }
                Abc_CexFree( pNew );
                return nil;
            }
        }
/*
        if ( i == 20 )
        {
            printf( "Frame %d : ", i );
            Gia_ManForEachRi( p, pObj, k )
                printf( "%d", pObj->fMark1 );
            printf( "\n" );
        }
*/
    }
    return pNew;
}
func Bmc_CexEssentialBitTest(p *Gia_Man_t, pCexState *Abc_Cex_t) {
    var pNew *Abc_Cex_t;
    var b int32;
    for  b = 0; b < pCexState.nBits; b++  {
        if b % 100 {
            continue; }
        pNew = Bmc_CexEssentialBitOne( p, pCexState, b, nil, nil );
        Bmc_CexPrint( pNew, Gia_ManPiNum(p), 0 );

        if Gia_ManPo(p, pCexState.iPo).fMark1 {
            fmt.Printf("Not essential\n");
        } else {
            fmt.Printf("Essential\n"); }

        Abc_CexFree( pNew );
    }
}

/**Function*************************************************************

  Synopsis    [Computes essential bits of the CEX.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexEssentialBits(p *Gia_Man_t, pCexState *Abc_Cex_t, pCexCare *Abc_Cex_t, fVerbose int32) *Abc_Cex_t {
    var pNew *Abc_Cex_t;
    var pTemp *Abc_Cex_t;
    var pPrev *Abc_Cex_t = nil;
    var b, fEqual, fPrevStatus int32;
//    abctime clk = Abc_Clock();
    assert(pCexState.nBits == pCexCare.nBits, "pCexState.nBits == pCexCare.nBits");
    // start the counter-example
    pNew = Abc_CexAlloc( 0, Gia_ManCiNum(p), pCexState.iFrame + 1 );
    pNew.iFrame = pCexState.iFrame;
    pNew.iPo    = pCexState.iPo;
    // iterate through care-bits
    for  b = 0; b < pCexState.nBits; b++  {
        // skip don't-care bits
        if !Abc_InfoHasBit(pCexCare.pData, b) {
            continue; }

        // skip state bits
        if b % pCexCare.nPis >= Gia_ManPiNum(p) {
            Abc_InfoSetBit( pNew.pData, b );
            continue;
        }

        // check if this is an essential bit
        pTemp = Bmc_CexEssentialBitOne( p, pCexState, b, pPrev, &fEqual );
//        pTemp = Bmc_CexEssentialBitOne( p, pCexState, b, NULL, &fEqual );
        if pTemp == nil {
            if (fEqual != 0) && (fPrevStatus != 0) {
                Abc_InfoSetBit( pNew.pData, b ); }
            continue;
        }
//        Bmc_CexPrint( pTemp, Gia_ManPiNum(p), fVerbose );
        Abc_CexFree( pPrev );
        pPrev = pTemp;

        // record essential bit
        fPrevStatus = !Gia_ManPo(p, pCexState.iPo).fMark1;
        if !Gia_ManPo(p, pCexState.iPo).fMark1 {
            Abc_InfoSetBit( pNew.pData, b ); }
    }
    Abc_CexFreeP( &pPrev );
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    fmt.Printf("Essentials:   ");
    Bmc_CexPrint( pNew, Gia_ManPiNum(p), fVerbose );
    return pNew;
}
 

/**Function*************************************************************

  Synopsis    [Computes essential bits of the CEX.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bmc_CexTest(p *Gia_Man_t, pCex *Abc_Cex_t, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var pCexImpl *Abc_Cex_t    = nil;
    var pCexStates *Abc_Cex_t  = Bmc_CexInnerStates( p, pCex, &pCexImpl, fVerbose );
    var pCexCare *Abc_Cex_t    = Bmc_CexCareBits( p, pCexStates, pCexImpl, nil, 1, fVerbose );
    var pCexEss, pCexMin *Abc_Cex_t;

    if !(Bmc_CexVerify( p, pCex, pCexCare ) != 0) {
        fmt.Printf("Counter-example care-set verification has failed.\n"); }

    pCexEss = Bmc_CexEssentialBits( p, pCexStates, pCexCare, fVerbose );
    pCexMin = Bmc_CexCareBits( p, pCexStates, pCexImpl, pCexEss, 0, fVerbose );

    if !(Bmc_CexVerify( p, pCex, pCexMin ) != 0) {
        fmt.Printf("Counter-example min-set verification has failed.\n"); }

//    Bmc_CexDumpStats( p, pCex, pCexCare, pCexEss, pCexMin, Abc_Clock() - clk );

    Abc_CexFreeP( &pCexStates );
    Abc_CexFreeP( &pCexImpl );
    Abc_CexFreeP( &pCexCare );
    Abc_CexFreeP( &pCexEss );
    Abc_CexFreeP( &pCexMin );

    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
//    Bmc_CexBuildNetworkTest( p, pCex );
//    Bmc_CexPerformUnrollingTest( p, pCex );
}



/**Function*************************************************************

  Synopsis    [Count the number of care bits.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountCareBits(p *Gia_Man_t, vPats *Vec_Wec_t) {
    var pObj *Gia_Obj_t;
    var i, k, fCompl0, fCompl1 int32;
    var Counter word;
    var vPat *Vec_Int_t;
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vPats)) && (((vPat) = Vec_WecEntry(vPats, i)), 1); i++  {
        var Count int32;
        assert(Vec_IntSize(vPat) == Gia_ManCiNum(p), "Vec_IntSize(vPat) == Gia_ManCiNum(p)");

        // propagate forward
        Gia_ManConst0(p).fMark0 = 0;
        Gia_ManConst0(p).fMark1 = 0;
        // inlining of macro Gia_ManForEachCi
        for k = 0; k < Vec_IntSize(p.vCis); k++ {
        	pObj = Gia_ManCi(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Vec_IntEntry(vPat, k);
            pObj.fMark1 = 0;
        }
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
            pObj.fMark0 = fCompl0 & fCompl1;
            pObj.fMark1 = 0;
        }
        }
        // inlining of macro Gia_ManForEachCo
        for k = 0; k < Vec_IntSize(p.vCos); k++ {
        	pObj = Gia_ManCo(p, k)
        	if pObj == nil { break }
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
            Gia_ObjFanin0(pObj).fMark1 = 1;
        }
        // propagate backward
        // inlining of macro Gia_ManForEachAndReverse
        for k = p.nObjs - 1; k > 0; k-- {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            if !pObj.fMark1 {
                continue; }
            if pObj.fMark0 == 0 {
                fCompl0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
                fCompl1 = Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj);
                assert(fCompl0 == 0 || fCompl1 == 0, "fCompl0 == 0 || fCompl1 == 0");
                if fCompl1 == 0 {
                    Gia_ObjFanin1(pObj).fMark1 = 1;
                } else if fCompl0 == 0 {
                    Gia_ObjFanin0(pObj).fMark1 = 1; }
                
            
            } else {
                Gia_ObjFanin0(pObj).fMark1 = 1;
                Gia_ObjFanin1(pObj).fMark1 = 1;
            }
        }
        }
        // inlining of macro Gia_ManForEachAnd
        for k = 0; k < p.nObjs; k++ {
        	pObj = Gia_ManObj(p, k)
        	if pObj == nil { break }
        	if !Gia_ObjIsAnd(pObj) {} else {
            Count += pObj.fMark1; }
        }
        Counter += Count;

        //printf( "%3d = %8d\n", i, Count );
    }
    Counter /= Vec_WecSize(vPats);
    fmt.Printf("Stats over %d patterns: Average care-nodes = %d (%6.2f %%)\n", Vec_WecSize(vPats), int32(Counter), 100.0*int32(Counter)/Gia_ManAndNum(p));
}

func Mnist_ReadImages1_() *unsigned char {
    var Size int32  = 60000 * 28 * 28 + 16;
    unsigned char * pData = (unsigned char *)malloc( Size );
    var pFile *os.File  = os.Open("train-images.idx3-ubyte");
    var RetValue int32  = fread( pData, 1, Size, pFile );
    assert(RetValue == Size, "RetValue == Size");
    _ = (pFile).Close();
    return pData;
}
func Mnist_ReadImages_(nPats int32) *Vec_Wec_t {
    var vPats *Vec_Wec_t  = Vec_WecStart( nPats );
    unsigned char * pL1 = Mnist_ReadImages1_();
    var i, k, x int32;
    for  i = 0; i < nPats; i++  {
        for  x = 0; x < 784; x++  {
            for  k = 0; k < 8; k++  {
                Vec_WecPush( vPats, i, (pL1[16 + i * 784 + x] >> k) & 1 );
            }
        }
    }
    free( pL1 );
    return vPats;
}
func Gia_ManCountCareBitsTest(p *Gia_Man_t) {
    var vPats *Vec_Wec_t  = Mnist_ReadImages_( 100 );
    Gia_ManCountCareBits( p, vPats );
    Vec_WecFree( vPats );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

