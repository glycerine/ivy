package gia

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [giaEra.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Explicit reachability analysis.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaEra.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/mem/mem.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// explicit state representation
   
type Gia_ObjEra_t struct {
    Num            int32;          // ID of this state
    Cond            int32;         // input condition
    iPrev            int32;        // previous state
    iNext            int32;        // next state in the hash table
    pData [0]uint32;     // state bits
}

// explicit state reachability
   
type Gia_ManEra_t struct {
    pAig     *Gia_Man_t;         // user's AIG manager
    nWordsSim            int32;       // 2^(PInum)
    nWordsDat            int32;   // Abc_BitWordNum
    pDataSim      *uint32;     // simulation data
    pMemory   *Mem_Fixed_t;      // memory manager
    vStates     *Vec_Ptr_t;      // reached states
    pStateNew  *Gia_ObjEra_t;    // temporary state
    iCurState            int32;    // the current state
    vBugTrace     *Vec_Int_t;    // the sequence of transitions
    vStgDump     *Vec_Int_t;     // STG written into a file
    // hash table for states
    nBins            int32;         
    pBins      *uint32;
}

func Gia_ManEraData(p *Gia_ManEra_t, i int32) *uint32 { return p.pDataSim + i * p.nWordsSim;  }
func Gia_ManEraState(p *Gia_ManEra_t, i int32) *Gia_ObjEra_t { return (Gia_ObjEra_t *)Vec_PtrEntry(p.vStates, i);  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates reachability manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEraCreate(pAig *Gia_Man_t) *Gia_ManEra_t {
    var vTruths *Vec_Ptr_t;
    
    var pTruth, pSimInfo *uint32;
    var i int32;
    // inlining of macro ABC_CALLOC
    p := &Gia_ManEra_t{};
    p.pAig      = pAig;
    p.nWordsSim = Abc_TruthWordNum( Gia_ManPiNum(pAig) );
    p.nWordsDat = Abc_BitWordNum( Gia_ManRegNum(pAig) );
    // inlining of macro ABC_ALLOC
    p.pDataSim  = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nWordsSim*Gia_ManObjNum(pAig))));
    p.pMemory   = Mem_FixedStart( sizeof(Gia_ObjEra_t) + sizeof(unsigned) * p.nWordsDat );
    p.vStates   = Vec_PtrAlloc( 100000 );
    p.nBins     = Abc_PrimeCudd( 100000 );
    // inlining of macro ABC_CALLOC
    p.pBins     = ((unsigned *) calloc((size_t)(p.nBins), sizeof(unsigned)));
    Vec_PtrPush( p.vStates, nil );
    // assign primary input values
    vTruths = Vec_PtrAllocTruthTables( Gia_ManPiNum(pAig) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vTruths)) && (((pTruth) = (unsigned *)Vec_PtrEntry(vTruths, i)), 1); i++  {
        pSimInfo = Gia_ManEraData( p, Gia_ObjId(pAig, Gia_ManPi(pAig, i)) );
        memcpy( pSimInfo, pTruth, sizeof(unsigned) * p.nWordsSim );
    }
    Vec_PtrFree( vTruths );
    // assign constant zero node
    pSimInfo = Gia_ManEraData( p, 0 );
    memset( pSimInfo, 0, sizeof(unsigned) * p.nWordsSim );
    p.vStgDump = Vec_IntAlloc( 1000 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes reachability manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEraFree(p *Gia_ManEra_t) {
    Mem_FixedStop( p.pMemory, 0 );
    Vec_IntFree( p.vStgDump );
    Vec_PtrFree( p.vStates );
    if p.vBugTrace { Vec_IntFree( p.vBugTrace ); }
    // inlining of macro ABC_FREE
    ((p.pDataSim) ? (free((char *) (p.pDataSim)), (p.pDataSim) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Creates new state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEraCreateState(p *Gia_ManEra_t) *Gia_ObjEra_t {
    var pNew *Gia_ObjEra_t;
    pNew = (Gia_ObjEra_t *)Mem_FixedEntryFetch( p.pMemory );
    pNew.Num = Vec_PtrSize( p.vStates );
    pNew.iPrev = 0;
    Vec_PtrPush( p.vStates, pNew );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Computes hash value of the node using its simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_ManEraStateHash_s_FPrimes [128]int32 = [128]int32{ 
        1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 
        1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997, 
        2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 
        2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089, 
        3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671, 
        3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243, 
        4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871, 
        4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471, 
        5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073, 
        6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689, 
        6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309, 
        7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933, 
        8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
    };
func Gia_ManEraStateHash(pState *uint32, nWordsSim int32, nTableSize int32) int32 {
    
    var uHash uint32;
    var i int32;
    uHash = 0;
    for  i = 0; i < nWordsSim; i++  {
        uHash ^= pState[i] * spStatic_Gia_ManEraStateHash_s_FPrimes[i & 0x7F];
    }
    return uHash % nTableSize;
}

/**Function*************************************************************

  Synopsis    [Returns the place of this state in the table or NULL if it exists.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEraHashFind(p *Gia_ManEra_t, pState *Gia_ObjEra_t, pStateNum *int32) *uint32 {
    var pThis *Gia_ObjEra_t;
    var pPlace *uint32  = p.pBins + Gia_ManEraStateHash( pState.pData, p.nWordsDat, p.nBins );
    for  pThis = (*pPlace)? Gia_ManEraState(p, *pPlace) : nil; (pThis != nil); 
          pPlace = (unsigned *)&pThis.iNext, pThis = (*pPlace)? Gia_ManEraState(p, *pPlace) : nil  {
              if !memcmp( pState.pData, pThis.pData, sizeof(unsigned) * p.nWordsDat ) {
                  if pStateNum != nil {
                      *pStateNum = pThis.Num; }
                  return nil;
              }
    }
    if pStateNum != nil {
      *pStateNum = -1; }
    return pPlace;
}

/**Function*************************************************************

  Synopsis    [Resizes the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManEraHashResize(p *Gia_ManEra_t) {
    var pThis *Gia_ObjEra_t;
    var pBinsOld, piPlace *uint32;
    var nBinsOld, iNext, Counter, i int32;
    assert(p.pBins != nil, "p.pBins != nil");
    // replace the table
    pBinsOld = p.pBins;
    nBinsOld = p.nBins;
    p.nBins = Abc_PrimeCudd( 3 * p.nBins ); 
    // inlining of macro ABC_CALLOC
    p.pBins = ((unsigned *) calloc((size_t)(p.nBins), sizeof(unsigned)));
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < nBinsOld; i++  {
    for  pThis = (pBinsOld[i]? Gia_ManEraState(p, pBinsOld[i]) : nil),
          iNext = (pThis? pThis.iNext : 0);  
          (pThis != nil);  pThis = (iNext? Gia_ManEraState(p, iNext) : nil),   
          iNext = (pThis? pThis.iNext : 0)   {
        assert(pThis.Num, "pThis.Num");
        pThis.iNext = 0;
        piPlace = Gia_ManEraHashFind( p, pThis, nil );
        assert(*piPlace == 0, "*piPlace == 0"); // should not be there
        *piPlace = pThis.Num;
        Counter++;
    }
    }
    assert(Counter == Vec_PtrSize( p.vStates ) - 1, "Counter == Vec_PtrSize( p.vStates ) - 1");
    // inlining of macro ABC_FREE
    ((pBinsOld) ? (free((char *) (pBinsOld)), (pBinsOld) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Initialize register output to the given state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInsertState(p *Gia_ManEra_t, pState *Gia_ObjEra_t) {
    var pObj *Gia_Obj_t;
    var pSimInfo *uint32;
    var i int32;
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCi(p.pAig, Gia_ManPiNum(p.pAig)+i)); i++  {
        pSimInfo = Gia_ManEraData( p, Gia_ObjId(p.pAig, pObj) );
        if Abc_InfoHasBit(pState.pData, i) {
            memset( pSimInfo, 0xff, sizeof(unsigned) * p.nWordsSim );
        } else {
            memset( pSimInfo, 0, sizeof(unsigned) * p.nWordsSim ); }
    }
}

/**Function*************************************************************

  Synopsis    [Returns -1 if outputs are not asserted.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManOutputAsserted(p *Gia_ManEra_t, pObj *Gia_Obj_t) int32 {
    var pInfo *uint32   = Gia_ManEraData( p, Gia_ObjId(p.pAig, pObj) );
    var w int32;
    for  w = 0; w < p.nWordsSim; w++  {
        if pInfo[w] {
            return 32*w + Gia_WordFindFirstBit( pInfo[w] ); }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimulateCo(p *Gia_ManEra_t, pObj *Gia_Obj_t) {
    var Id int32  = Gia_ObjId(p.pAig, pObj);
    var pInfo *uint32   = Gia_ManEraData( p, Id );
    var pInfo0 *uint32  = Gia_ManEraData( p, Gia_ObjFaninId0(pObj, Id) );
    var w int32;
    if Gia_ObjFaninC0(pObj) {
        for  w = p.nWordsSim-1; w >= 0; w--  {
            pInfo[w] = ~pInfo0[w];
        }
    } else { 
        for  w = p.nWordsSim-1; w >= 0; w--  {
            pInfo[w] = pInfo0[w];
        } }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSimulateNode(p *Gia_ManEra_t, pObj *Gia_Obj_t) {
    var Id int32  = Gia_ObjId(p.pAig, pObj);
    var pInfo *uint32   = Gia_ManEraData( p, Id );
    var pInfo0 *uint32  = Gia_ManEraData( p, Gia_ObjFaninId0(pObj, Id) );
    var pInfo1 *uint32  = Gia_ManEraData( p, Gia_ObjFaninId1(pObj, Id) );
    var w int32;
    if Gia_ObjFaninC0(pObj) {
        if Gia_ObjFaninC1(pObj) {
            for  w = p.nWordsSim-1; w >= 0; w--  {
                pInfo[w] = ~(pInfo0[w] | pInfo1[w]);
            }
        } else { 
            for  w = p.nWordsSim-1; w >= 0; w--  {
                pInfo[w] = ~pInfo0[w] & pInfo1[w];
            } }
    
    } else {
        if Gia_ObjFaninC1(pObj) {
            for  w = p.nWordsSim-1; w >= 0; w--  {
                pInfo[w] = pInfo0[w] & ~pInfo1[w];
            }
        } else { 
            for  w = p.nWordsSim-1; w >= 0; w--  {
                pInfo[w] = pInfo0[w] & pInfo1[w];
            } }
    }
}

/**Function*************************************************************

  Synopsis    [Performs one iteration of reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformOneIter(p *Gia_ManEra_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            Gia_ManSimulateNode( p, pObj );
        } else if Gia_ObjIsCo(pObj) {
            Gia_ManSimulateCo( p, pObj ); }
    }
}

/**Function*************************************************************

  Synopsis    [Performs one iteration of reachability analysis.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectBugTrace(p *Gia_ManEra_t, pState *Gia_ObjEra_t, iCond int32) *Vec_Int_t {
    var vTrace *Vec_Int_t;
    vTrace = Vec_IntAlloc( 10 );
    Vec_IntPush( vTrace, iCond );
    for  ; pState; pState = pState.iPrev ? Gia_ManEraState(p, pState.iPrev) : nil  {
        Vec_IntPush( vTrace, pState.Cond );
    }
    Vec_IntReverseOrder( vTrace );
    return vTrace;
}

/**Function*************************************************************

  Synopsis    [Counts the depth of state transitions leading ot this state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountDepth(p *Gia_ManEra_t) int32 {
    var pState *Gia_ObjEra_t;
    var Counter int32;
    pState = (Gia_ObjEra_t *)Vec_PtrEntryLast( p.vStates );
    if pState.iPrev == 0 && Vec_PtrSize(p.vStates) > 3 {
        pState = (Gia_ObjEra_t *)Vec_PtrEntry( p.vStates, Vec_PtrSize(p.vStates) - 2 ); }
    for  ; pState; pState = pState.iPrev ? Gia_ManEraState(p, pState.iPrev) : nil  {
        Counter++;
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Analized reached states.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAnalyzeResult(p *Gia_ManEra_t, pState *Gia_ObjEra_t, fMiter int32, fStgDump int32) int32 {
    var pObj *Gia_Obj_t;
    var pSimInfo *uint32;
    var piPlace *uint32;
    var uOutput uint32;
    var i, k, iCond, nMints, iNextState int32;
    // check if the miter is asserted
    if fMiter != 0 {
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, i)); i++  {
            iCond = Gia_ManOutputAsserted( p, pObj );
            if iCond >= 0 {
                p.vBugTrace = Gia_ManCollectBugTrace( p, pState, iCond );
                return 1;
            }
        }
    }
    // collect reached states 
    nMints = (1 << Gia_ManPiNum(p.pAig));
    for  k = 0; k < nMints; k++  {
        if p.pStateNew == nil {
            p.pStateNew = Gia_ManEraCreateState( p ); }
        p.pStateNew.pData[p.nWordsDat-1] = 0;
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, Gia_ManPoNum(p.pAig)+i)); i++  {
            pSimInfo = Gia_ManEraData( p, Gia_ObjId(p.pAig, pObj) );
            if Abc_InfoHasBit(p.pStateNew.pData, i) != Abc_InfoHasBit(pSimInfo, k) {
                Abc_InfoXorBit( p.pStateNew.pData, i ); }
        }
        if fStgDump != 0 {
            uOutput = 0;
            // inlining of macro Gia_ManForEachPo
            for  i = 0; (i < Gia_ManPoNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, i)); i++  {
                pSimInfo = Gia_ManEraData( p, Gia_ObjId(p.pAig, pObj) );
                if Abc_InfoHasBit(pSimInfo, k) && i < 32 {
                    Abc_InfoXorBit( &uOutput, i ); }
            }
        }
        piPlace = Gia_ManEraHashFind( p, p.pStateNew, &iNextState );
        if fStgDump != 0 { Vec_IntPush( p.vStgDump, k ); }
        if fStgDump != 0 { Vec_IntPush( p.vStgDump, pState.Num ); }
        if piPlace == nil {
            if fStgDump != 0 { Vec_IntPush( p.vStgDump, iNextState ); }
            if fStgDump != 0 { Vec_IntPush( p.vStgDump, uOutput ); }
            continue;
        }
        if fStgDump != 0 { Vec_IntPush( p.vStgDump, p.pStateNew.Num ); }
        if fStgDump != 0 { Vec_IntPush( p.vStgDump, uOutput ); }
//printf( "Inserting %d ", Vec_PtrSize(p->vStates) );
//Extra_PrintBinary( stdout, p->pStateNew->pData, Gia_ManRegNum(p->pAig) );  printf( "\n" );
        assert(*piPlace == 0, "*piPlace == 0");
        *piPlace = p.pStateNew.Num;
        p.pStateNew.Cond = k;
        p.pStateNew.iPrev = pState.Num;
        p.pStateNew.iNext = 0;
        p.pStateNew = nil;
        // expand hash table if needed
        if Vec_PtrSize(p.vStates) > 2 * p.nBins {
            Gia_ManEraHashResize( p ); }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Resizes the hash table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCollectReachable(pAig *Gia_Man_t, nStatesMax int32, fMiter int32, fDumpFile int32, fVerbose int32) int32 { 
    var p *Gia_ManEra_t;
    var pState *Gia_ObjEra_t;
    var Hash int32;
    var clk abctime  = Abc_Clock();
    var RetValue int32  = 1;
    assert(Gia_ManPiNum(pAig) <= 12, "Gia_ManPiNum(pAig) <= 12");
    assert(Gia_ManRegNum(pAig) > 0, "Gia_ManRegNum(pAig) > 0");
    p = Gia_ManEraCreate( pAig );
    // create init state
    pState = Gia_ManEraCreateState( p );
    pState.Cond  = 0; 
    pState.iPrev = 0;
    pState.iNext = 0; 
    memset( pState.pData, 0, sizeof(unsigned) * p.nWordsDat );
    Hash = Gia_ManEraStateHash(pState.pData, p.nWordsDat, p.nBins);
    p.pBins[ Hash ] = pState.Num;
    // process reachable states
    for  p.iCurState < Vec_PtrSize( p.vStates ) - 1  {
        if Vec_PtrSize(p.vStates) >= nStatesMax {
            fmt.Printf("Reached the limit on states traversed (%d).  ", nStatesMax);
            RetValue = -1;
            break;
        }
        pState = Gia_ManEraState( p, ++p.iCurState );
        if p.iCurState > 1 && pState.iPrev == 0 {
            continue; }
//printf( "Extracting %d  ", p->iCurState );
//Extra_PrintBinary( stdout, p->pStateNew->pData, Gia_ManRegNum(p->pAig) );  printf( "\n" );
        Gia_ManInsertState( p, pState );
        Gia_ManPerformOneIter( p );
        if (Gia_ManAnalyzeResult( p, pState, fMiter, fDumpFile ) != 0) && (fMiter != 0) {
            RetValue = 0;
            fmt.Printf("Miter failed in state %d after %d transitions.  ", p.iCurState, Vec_IntSize(p.vBugTrace)-1);
            break;
        }
        if (fVerbose != 0) && p.iCurState % 5000 == 0 {
            fmt.Printf("States =%10d. Reached =%10d. R = %5.3f. Depth =%6d. Mem =%9.2f MB.  ", p.iCurState, Vec_PtrSize(p.vStates), 1.0*p.iCurState/Vec_PtrSize(p.vStates), Gia_ManCountDepth(p), (1.0/(1<<20))*(1.0*Vec_PtrSize(p.vStates)*(sizeof(Gia_ObjEra_t) + sizeof(unsigned) * p.nWordsDat) + 
                               1.0*p.nBins*sizeof(unsigned) + 1.0*p.vStates.nCap * sizeof(void*)));
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
    }
    fmt.Printf("Reachability analysis traversed %d states with depth %d.  ", p.iCurState-1, Gia_ManCountDepth(p));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    if fDumpFile != 0 {
        var pFileName string  = "test.stg";
        var pFile *os.File  = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
        if pFile == nil {
            fmt.Printf("Cannot open file \"%s\" for writing.\n", pFileName);
        } else {
            Gia_ManStgPrint( pFile, p.vStgDump, Gia_ManPiNum(pAig), Gia_ManPoNum(pAig), p.iCurState-1 );
            _ = (pFile).Close();
            fmt.Printf("Extracted STG was written into file \"%s\".\n", pFileName);
        }
    }
    Gia_ManEraFree( p );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

