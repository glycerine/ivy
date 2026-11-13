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

