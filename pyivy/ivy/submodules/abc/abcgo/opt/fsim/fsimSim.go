package fsim

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [fsimSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast sequential AIG simulator.]

  Synopsis    [Simulation procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: fsimSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "fsimInt.h"
//#include "aig/ssw/ssw.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimInfoRandom(p *Fsim_Man_t, pInfo *uint32) {
    var w int32;
    for  w = p.nWords-1; w >= 0; w--  {
        pInfo[w] = Aig_ManRandom( 0 );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimInfoZero(p *Fsim_Man_t, pInfo *uint32) {
    var w int32;
    for  w = p.nWords-1; w >= 0; w--  {
        pInfo[w] = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Returns index of the first pattern that failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimInfoIsZero(p *Fsim_Man_t, pInfo *uint32) int32 {
    var w int32;
    for  w = p.nWords-1; w >= 0; w--  {
        if pInfo[w] {
            return 32*(w-1) + Aig_WordFindFirstBit( pInfo[w] ); }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimInfoOne(p *Fsim_Man_t, pInfo *uint32) {
    var w int32;
    for  w = p.nWords-1; w >= 0; w--  {
        pInfo[w] = ~0;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimInfoCopy(p *Fsim_Man_t, pInfo *uint32, pInfo0 *uint32) {
    var w int32;
    for  w = p.nWords-1; w >= 0; w--  {
        pInfo[w] = pInfo0[w];
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimulateCi(p *Fsim_Man_t, iNode int32, iCi int32) {
    var pInfo *uint32   = Fsim_SimData( p, iNode % p.nFront );
    var pInfo0 *uint32  = Fsim_SimDataCi( p, iCi );
    var w int32;
    for  w = p.nWords-1; w >= 0; w--  {
        pInfo[w] = pInfo0[w];
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimulateCo(p *Fsim_Man_t, iCo int32, iFan0 int32) {
    var pInfo *uint32   = Fsim_SimDataCo( p, iCo );
    var pInfo0 *uint32  = Fsim_SimData( p, Fsim_Lit2Var(iFan0) % p.nFront );
    var w int32;
    if Fsim_LitIsCompl(iFan0) {
        for  w = p.nWords-1; w >= 0; w--  {
            pInfo[w] = ~pInfo0[w];
        }
    } else { //if ( !Fsim_LitIsCompl(iFan0) )
        for  w = p.nWords-1; w >= 0; w--  {
            pInfo[w] = pInfo0[w];
        } }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimulateNode(p *Fsim_Man_t, iNode int32, iFan0 int32, iFan1 int32) {
    var pInfo *uint32   = Fsim_SimData( p, iNode % p.nFront );
    var pInfo0 *uint32  = Fsim_SimData( p, Fsim_Lit2Var(iFan0) % p.nFront );
    var pInfo1 *uint32  = Fsim_SimData( p, Fsim_Lit2Var(iFan1) % p.nFront );
    var w int32;
    if Fsim_LitIsCompl(iFan0) && Fsim_LitIsCompl(iFan1) {
        for  w = p.nWords-1; w >= 0; w--  {
            pInfo[w] = ~(pInfo0[w] | pInfo1[w]);
        }
    } else if Fsim_LitIsCompl(iFan0) && !Fsim_LitIsCompl(iFan1) {
        for  w = p.nWords-1; w >= 0; w--  {
            pInfo[w] = ~pInfo0[w] & pInfo1[w];
        }
    } else if !Fsim_LitIsCompl(iFan0) && Fsim_LitIsCompl(iFan1) {
        for  w = p.nWords-1; w >= 0; w--  {
            pInfo[w] = pInfo0[w] & ~pInfo1[w];
        }
    } else { //if ( !Fsim_LitIsCompl(iFan0) && !Fsim_LitIsCompl(iFan1) )
        for  w = p.nWords-1; w >= 0; w--  {
            pInfo[w] = pInfo0[w] & pInfo1[w];
        } }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimInfoInit(p *Fsim_Man_t) {
    var iPioNum, i int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCis2Ids)) && (((iPioNum) = Vec_IntEntry(p.vCis2Ids, i)), 1); i++  {
        if iPioNum < p.nPis {
            Fsim_ManSimInfoRandom( p, Fsim_SimDataCi(p, i) );
        } else {
            Fsim_ManSimInfoZero( p, Fsim_SimDataCi(p, i) ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimInfoTransfer(p *Fsim_Man_t) {
    var iPioNum, i int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCis2Ids)) && (((iPioNum) = Vec_IntEntry(p.vCis2Ids, i)), 1); i++  {
        if iPioNum < p.nPis {
            Fsim_ManSimInfoRandom( p, Fsim_SimDataCi(p, i) );
        } else {
            Fsim_ManSimInfoCopy( p, Fsim_SimDataCi(p, i), Fsim_SimDataCo(p, p.nPos+iPioNum-p.nPis) ); }
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManRestoreNum(p *Fsim_Man_t) int32 {
    var ch, i, x int32;
    for  i = 0; (ch = *p.pDataCur++) & 0x80; i++  {
        x |= (ch & 0x7f) << (7 * i);
    }
    assert(p.pDataCur - p.pDataAig < p.nDataAig, "p.pDataCur - p.pDataAig < p.nDataAig");
    return x | (ch << (7 * i));
}

/**Function*************************************************************

  Synopsis    []

  Description []
  
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManRestoreObj(p *Fsim_Man_t, pObj *Fsim_Obj_t) int32 {
    var iValue int32  = Fsim_ManRestoreNum( p );
    if (iValue & 3) == 3 { // and
    
        pObj.iNode = (iValue >> 2) + p.iNodePrev;
        pObj.iFan0 = (pObj.iNode << 1) - Fsim_ManRestoreNum( p );
        pObj.iFan1 = pObj.iFan0 - Fsim_ManRestoreNum( p );
        p.iNodePrev = pObj.iNode;
    
    } else if (iValue & 3) == 1 { // ci
    
        pObj.iNode = (iValue >> 2) + p.iNodePrev;
        pObj.iFan0 = 0;
        pObj.iFan1 = 0;
        p.iNodePrev = pObj.iNode;
    
    } else { // if ( (iValue & 1) == 0 ) // co
    
        pObj.iNode = 0;
        pObj.iFan0 = ((p.iNodePrev << 1) | 1) - (iValue >> 1);
        pObj.iFan1 = 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimulateRound2(p *Fsim_Man_t) {
    var pObj *Fsim_Obj_t;
    var i, iCis, iCos int32;
    if Aig_ObjRefs(Aig_ManConst1(p.pAig)) {
        Fsim_ManSimInfoOne( p, Fsim_SimData(p, 1) ); }
    // inlining of macro Fsim_ManForEachObj
    for  i = 2, p.pDataCur = p.pDataAig, p.iNodePrev = 0, pObj = &p.Obj;        i < p.nObjs && (Fsim_ManRestoreObj( p, pObj ) != 0); i++  {
        if pObj.iFan0 == 0 {
            Fsim_ManSimulateCi( p, pObj.iNode, iCis++ );
        } else if pObj.iFan1 == 0 {
            Fsim_ManSimulateCo( p, iCos++, pObj.iFan0 );
        } else {
            Fsim_ManSimulateNode( p, pObj.iNode, pObj.iFan0, pObj.iFan1 ); }
    }
    assert(iCis == p.nCis, "iCis == p.nCis");
    assert(iCos == p.nCos, "iCos == p.nCos");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimulateRound(p *Fsim_Man_t) {
    var pCur, pEnd *int32;
    var iCis, iCos int32;
    if p.pDataAig2 == nil {
        Fsim_ManSimulateRound2( p );
        return;
    }
    if Aig_ObjRefs(Aig_ManConst1(p.pAig)) {
        Fsim_ManSimInfoOne( p, Fsim_SimData(p, 1) ); }
    pCur = p.pDataAig2 + 6;
    pEnd = p.pDataAig2 + 3 * p.nObjs;
    for  pCur < pEnd  {
        if pCur[1] == 0 {
            Fsim_ManSimulateCi( p, pCur[0], iCis++ );
        } else if pCur[2] == 0 {
            Fsim_ManSimulateCo( p, iCos++, pCur[1] );
        } else {
            Fsim_ManSimulateNode( p, pCur[0], pCur[1], pCur[2] ); }
        pCur += 3;
    }
    assert(iCis == p.nCis, "iCis == p.nCis");
    assert(iCos == p.nCos, "iCos == p.nCos");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimulateRoundTest(p *Fsim_Man_t) {
    var pObj *Fsim_Obj_t;
    var i int32;
    var clk int64  = time.Now().UnixNano();
    // inlining of macro Fsim_ManForEachObj
    for  i = 2, p.pDataCur = p.pDataAig, p.iNodePrev = 0, pObj = &p.Obj;        i < p.nObjs && (Fsim_ManRestoreObj( p, pObj ) != 0); i++  {
    }
//    ABC_PRT( "Unpacking time", p->pPars->nIters * (clock() - clk) );
}

/**Function*************************************************************

  Synopsis    [Returns index of the PO and pattern that failed it.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManCheckPos(p *Fsim_Man_t, piPo *int32, piPat *int32) int32 {
    var i, iPat int32;
    for  i = 0; i < p.nPos; i++  {
        iPat = Fsim_ManSimInfoIsZero( p, Fsim_SimDataCo(p, i) );
        if iPat >= 0 {
            *piPo = i;
            *piPat = iPat;
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManGenerateCounter(pAig *Aig_Man_t, iFrame int32, iOut int32, nWords int32, iPat int32, vCis2Ids *Vec_Int_t) *Abc_Cex_t {
    var p *Abc_Cex_t;
    var pData *uint32;
    var f, i, w, iPioId, Counter int32;
    p = Abc_CexAlloc( Aig_ManRegNum(pAig), Saig_ManPiNum(pAig), iFrame+1 );
    p.iFrame = iFrame;
    p.iPo = iOut;
    // fill in the binary data
    Aig_ManRandom( 1 );
    Counter = p.nRegs;
    // inlining of macro ABC_ALLOC
    pData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nWords)));
    for  f = 0; f <= iFrame; f++, Counter += p.nPis  {
    for  i = 0; i < Aig_ManPiNum(pAig); i++  {
        iPioId = Vec_IntEntry( vCis2Ids, i );
        if iPioId >= p.nPis {
            continue; }
        for  w = nWords-1; w >= 0; w--  {
            pData[w] = Aig_ManRandom( 0 );
        }
        if Aig_InfoHasBit( pData, iPat ) {
            Aig_InfoSetBit( p.pData, Counter + iPioId ); }
    }
    }
    pData = nil;
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fsim_ManSimulate(pAig *Aig_Man_t, pPars *Fsim_ParSim_t) int32 {
    var p *Fsim_Man_t;
    var Status Sec_MtrStatus_t;
    var i, iOut, iPat int32;
    var clk int64;
    var clkTotal int64 = time.Now().UnixNano();
    var clk2 int64;
    var clk2Total int64;
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    if pPars.fCheckMiter {
        Status = Sec_MiterStatus( pAig );
        if Status.nSat > 0 {
            fmt.Printf("Miter is trivially satisfiable (output %d).\n", Status.iOut);
            return 1;
        }
        if Status.nUndec == 0 {
            fmt.Printf("Miter is trivially unsatisfiable.\n");
            return 0;
        }
    }
    // create manager
    clk = time.Now().UnixNano();
    p = Fsim_ManCreate( pAig );
    p.nWords = pPars.nWords;
    if pPars.fVerbose {
        fmt.Printf("Obj = %8d (%8d). Cut = %6d. Front = %6d.  FrtMem = %7.2f MB. ", p.nObjs, p.nCis + p.nNodes, p.nCrossCutMax, p.nFront, 4.0*p.nWords*(p.nFront)/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }
    // create simulation frontier
    clk = time.Now().UnixNano();
    Fsim_ManFront( p, pPars.fCompressAig );
    if pPars.fVerbose {
        fmt.Printf("Max ID = %8d. Log max ID = %2d.  AigMem = %7.2f MB (%5.2f byte/obj).  ", p.iNumber, Aig_Base2Log(p.iNumber), 1.0*(p.pDataCur-p.pDataAig)/(1<<20), 1.0*(p.pDataCur-p.pDataAig)/p.nObjs); 
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
    }
    // perform simulation
    Aig_ManRandom( 1 );
    assert(p.pDataSim == nil, "p.pDataSim == nil");
    // inlining of macro ABC_ALLOC
    p.pDataSim = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nWords * p.nFront)));
    // inlining of macro ABC_ALLOC
    p.pDataSimCis = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nWords * p.nCis)));
    // inlining of macro ABC_ALLOC
    p.pDataSimCos = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nWords * p.nCos)));
    Fsim_ManSimInfoInit( p );
    for  i = 0; i < pPars.nIters; i++  { 
        Fsim_ManSimulateRound( p );
        if pPars.fVerbose {
            fmt.Printf("Frame %4d out of %4d and timeout %3d sec. ", i+1, pPars.nIters, pPars.TimeLimit);
            fmt.Printf("Time = %7.2f sec\r", (1.0*time.Now().UnixNano()-clkTotal)/int64(time.Second));
        }
        if pPars.fCheckMiter && (Fsim_ManCheckPos( p, &iOut, &iPat ) != 0) {
            assert(pAig.pSeqModel == nil, "pAig.pSeqModel == nil");
            pAig.pSeqModel = Fsim_ManGenerateCounter( pAig, i, iOut, p.nWords, iPat, p.vCis2Ids );
            if pPars.fVerbose {
            fmt.Printf("Miter is satisfiable after simulation (output %d).\n", iOut); }
            break;
        }
        if (time.Now().UnixNano() - clkTotal)/int64(time.Second) >= pPars.TimeLimit {
            break; }
        clk2 = time.Now().UnixNano();
        if i < pPars.nIters - 1 {
            Fsim_ManSimInfoTransfer( p ); }
        clk2Total += time.Now().UnixNano() - clk2;
    }
    if pAig.pSeqModel == nil {
        fmt.Printf("No bug detected after %d frames with time limit %d seconds.\n", i+1, pPars.TimeLimit); }
    if pPars.fVerbose {
        fmt.Printf("Maxcut = %8d.  AigMem = %7.2f MB.  SimMem = %7.2f MB.  ", p.nCrossCutMax, p.pDataAig2? 12.0*p.nObjs/(1<<20) : 1.0*(p.pDataCur-p.pDataAig)/(1<<20), 4.0*p.nWords*(p.nFront+p.nCis+p.nCos)/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Sim time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clkTotal)))/(float64(int64(time.Second))));;

//        ABC_PRT( "Additional time", clk2Total );
//        Fsim_ManSimulateRoundTest( p );
//        Fsim_ManSimulateRoundTest2( p );
    }
    Fsim_ManDelete( p );
    return pAig.pSeqModel != nil;

}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

