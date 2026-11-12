package bmc

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [bmcMulti.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Proving multi-output properties.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcMulti.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "proof/ssw/ssw.h"
//#include "misc/extra/extra.h"
//#include "aig/gia/giaAig.h"
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Divides outputs into solved and unsolved.]

  Description [Return array of unsolved outputs to extract into a new AIG.
  Updates the resulting CEXes (vCexesOut) and current output map (vOutMap).]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManProcessOutputs(vCexesIn *Vec_Ptr_t, vCexesOut *Vec_Ptr_t, vOutMap *Vec_Int_t) *Vec_Int_t {
    var pCex *Abc_Cex_t; 
    var vLeftOver *Vec_Int_t;
    var i, iOut int32;
    assert(Vec_PtrSize(vCexesIn) == Vec_IntSize(vOutMap), "Vec_PtrSize(vCexesIn) == Vec_IntSize(vOutMap)");
    vLeftOver = Vec_IntAlloc( Vec_PtrSize(vCexesIn) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOutMap)) && (((iOut) = Vec_IntEntry(vOutMap, i)), 1); i++  {
        assert(Vec_PtrEntry(vCexesOut, iOut) == nil, "Vec_PtrEntry(vCexesOut, iOut) == nil");
        pCex = (Abc_Cex_t *)Vec_PtrEntry( vCexesIn, i );
        if pCex != nil { // found a CEX for output iOut
        
            Vec_PtrWriteEntry( vCexesIn, i, nil );
            Vec_PtrWriteEntry( vCexesOut, iOut, pCex );
        
        } else { // still unsolved
        
            Vec_IntWriteEntry( vOutMap, Vec_IntSize(vLeftOver), iOut );
            Vec_IntPush( vLeftOver, i );
        }
    }
    Vec_IntShrink( vOutMap, Vec_IntSize(vLeftOver) );
    return vLeftOver;
}

/**Function*************************************************************

  Synopsis    [Counts the number of constant 0 POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountConst0PosGia(p *Gia_Man_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Counter int32;
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Counter += (Gia_ObjFaninLit0p(p, pObj) == 0);
    }
    return Counter;
}
func Gia_ManCountConst0Pos(p *Aig_Man_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Counter += (Aig_ObjChild0(pObj) == Aig_ManConst0(p));
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMultiReport(p *Aig_Man_t, pStr string, nTotalPo int32, nTotalSize int32, clkStart abctime) {
    fmt.Printf("%3s : ", pStr);
    fmt.Printf("PI =%6d  ", Saig_ManPiNum(p));
    fmt.Printf("PO =%6d  ", Saig_ManPoNum(p));
    fmt.Printf("FF =%7d  ", Saig_ManRegNum(p));
    fmt.Printf("ND =%7d  ", Aig_ManNodeNum(p));
    fmt.Printf("Solved =%7d (%5.1f %%)  ", nTotalPo-Saig_ManPoNum(p), 100.0*(nTotalPo-Saig_ManPoNum(p))/Abc_MaxInt(1, nTotalPo));
    fmt.Printf("Size   =%7d (%5.1f %%)  ", Aig_ManObjNum(p), 100.0*Aig_ManObjNum(p)/Abc_MaxInt(1, nTotalSize));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clkStart );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMultiProveSyn(p *Aig_Man_t, fVerbose int32, fVeryVerbose int32) *Aig_Man_t {
    var pAig *Aig_Man_t;
    var pGia, pTemp *Gia_Man_t;
    pGia = Gia_ManFromAig( p );
    pGia = Gia_ManAigSyn2( pTemp = pGia, 1, 0, 0, 0, 0, 0, 0 );
    Gia_ManStop( pTemp );
    pAig = Gia_ManToAig( pGia, 0 );
    Gia_ManStop( pGia );
    return pAig;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMultiProveAig(p *Aig_Man_t, pPars *Bmc_MulPar_t) *Vec_Ptr_t {
    var ParsSim Ssw_RarPars_t;
    var pParsSim *Ssw_RarPars_t = &ParsSim;
    var ParsBmc Saig_ParBmc_t;
    var pParsBmc *Saig_ParBmc_t = &ParsBmc;
    var vOutMap, vLeftOver *Vec_Int_t;
    var vCexes *Vec_Ptr_t;
    var pTemp *Aig_Man_t;
    var clkStart abctime  = Abc_Clock();
    var nTimeToStop abctime;
    if pPars.TimeOutGlo {
        nTimeToStop = Abc_Clock() + pPars.TimeOutGlo * int64(time.Second);
    } else {
        nTimeToStop = 0;
    }
    var nTotalPo int32      = Saig_ManPoNum(p);
    var nTotalSize int32    = Aig_ManObjNum(p);
    var TimeOutLoc int32    = pPars.TimeOutLoc;
    var i int32;
    var RetValue int32 = -1;
    if pPars.fVerbose {
        fmt.Printf("MultiProve parameters: Global timeout = %d sec.  Local timeout = %d sec.  Time increase = %d %%.\n", pPars.TimeOutGlo, pPars.TimeOutLoc, pPars.TimeOutInc); }
    if pPars.fVerbose {
        fmt.Printf("Gap timout = %d sec. Per-output timeout = %d msec. Use synthesis = %d. Dump final = %d. Verbose = %d.\n", pPars.TimeOutGap, pPars.TimePerOut, pPars.fUseSyn, pPars.fDumpFinal, pPars.fVerbose); }
    // create output map
    vOutMap = Vec_IntStartNatural( Saig_ManPoNum(p) ); // maps current outputs into their original IDs
    vCexes  = Vec_PtrStart( Saig_ManPoNum(p) );        // maps solved outputs into their CEXes (or markers)
    for  i = 0; i < 1000; i++  {
        var nSolved int32  = Vec_PtrCountZero(vCexes);
        // perform SIM3
        Ssw_RarSetDefaultParams( pParsSim );
        pParsSim.fSolveAll = 1;
        pParsSim.fNotVerbose = 1;
        pParsSim.fSilent = !pPars.fVeryVerbose;
        pParsSim.TimeOut = TimeOutLoc;
        pParsSim.nRandSeed = (i * 17) % 500;
        pParsSim.nWords = 5;
        RetValue *= Ssw_RarSimulate( p, pParsSim );
        // sort outputs
        if p.vSeqModelVec {
            vLeftOver = Gia_ManProcessOutputs( p.vSeqModelVec, vCexes, vOutMap );
            if Vec_IntSize(vLeftOver) == 0 {
                break; }
            // remove solved        
            p = Saig_ManDupCones( pTemp = p, Vec_IntArray(vLeftOver), Vec_IntSize(vLeftOver) );
            Vec_IntFree( vLeftOver );
            Aig_ManStop( pTemp );
        }
        if pPars.fVerbose {
            Gia_ManMultiReport( p, "SIM", nTotalPo, nTotalSize, clkStart ); }

        // check timeout
        if nTimeToStop && Abc_Clock() + TimeOutLoc * int64(time.Second) > nTimeToStop {
            fmt.Printf("Global timeout (%d sec) is reached.\n", pPars.TimeOutGlo);
            break;
        }

        // perform BMC
        Saig_ParBmcSetDefaultParams( pParsBmc );
        pParsBmc.fSolveAll = 1;
        pParsBmc.fNotVerbose = 1;
        pParsBmc.fSilent = !pPars.fVeryVerbose;
        pParsBmc.nTimeOut = TimeOutLoc;
        pParsBmc.nTimeOutOne = pPars.TimePerOut;
        RetValue *= Saig_ManBmcScalable( p, pParsBmc );
        if pPars.fVeryVerbose {
            Abc_Print( 1, "Some outputs are SAT (%d out of %d) after %d frames.\n", 
                Saig_ManPoNum(p) - Vec_PtrCountZero(p.vSeqModelVec), Saig_ManPoNum(p), pParsBmc.iFrame ); }
        // sort outputs
        if p.vSeqModelVec {
            vLeftOver = Gia_ManProcessOutputs( p.vSeqModelVec, vCexes, vOutMap );
            if Vec_IntSize(vLeftOver) == 0 {
                break; }
            // remove solved        
            p = Saig_ManDupCones( pTemp = p, Vec_IntArray(vLeftOver), Vec_IntSize(vLeftOver) );
            Vec_IntFree( vLeftOver );
            Aig_ManStop( pTemp );
        }
        if pPars.fVerbose {
            Gia_ManMultiReport( p, "BMC", nTotalPo, nTotalSize, clkStart ); }

        // check timeout
        if nTimeToStop && Abc_Clock() + TimeOutLoc * int64(time.Second) > nTimeToStop {
            fmt.Printf("Global timeout (%d sec) is reached.\n", pPars.TimeOutGlo);
            break;
        }

        // check gap timeout
        if pPars.TimeOutGap && pPars.TimeOutGap <= TimeOutLoc && nSolved == Vec_PtrCountZero(vCexes) {
            fmt.Printf("Gap timeout (%d sec) is reached.\n", pPars.TimeOutGap);
            break;
        }

        // synthesize
        if pPars.fUseSyn {
            p = Gia_ManMultiProveSyn( pTemp = p, pPars.fVerbose, pPars.fVeryVerbose );
            Aig_ManStop( pTemp );
            if pPars.fVerbose {
                Gia_ManMultiReport( p, "SYN", nTotalPo, nTotalSize, clkStart ); }
        }

        // increase timeout
        TimeOutLoc += TimeOutLoc * pPars.TimeOutInc / 100;
    }
    Vec_IntFree( vOutMap );
    if pPars.fVerbose {
        fmt.Printf("The number of POs proved UNSAT by synthesis = %d.\n", Gia_ManCountConst0Pos(p)); }
    if pPars.fDumpFinal {
        var pFileName string  = Extra_FileNameGenericAppend( p.pName, "_out.aig" );
        Ioa_WriteAiger( p, pFileName, 0, 0 );
        fmt.Printf("Final AIG was dumped into file \"%s\".\n", pFileName);
    }
    Aig_ManStop( p );
    return vCexes;
}
func Gia_ManMultiProve(p *Gia_Man_t, pPars *Bmc_MulPar_t) int32 {
    var pAig *Aig_Man_t;
    if p.vSeqModelVec {
        Vec_PtrFreeFree( p.vSeqModelVec );
p.vSeqModelVec = nil; }
    pAig = Gia_ManToAig( p, 0 );
    p.vSeqModelVec = Gia_ManMultiProveAig( pAig, pPars ); // deletes pAig
    assert(Vec_PtrSize(p.vSeqModelVec) == Gia_ManPoNum(p), "Vec_PtrSize(p.vSeqModelVec) == Gia_ManPoNum(p)");
    if Vec_PtrCountZero(p.vSeqModelVec) == Vec_PtrSize(p.vSeqModelVec) {
        return -1;
    } else {
        return 0;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

