package sfm

import (
	"fmt"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [sfmDec.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based optimization using internal don't-cares.]

  Synopsis    [SAT-based decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: sfmDec.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sfmInt.h"
//#include "misc/st/st.h"
//#include "map/mio/mio.h"
//#include "base/abc/abc.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"
//#include "map/mio/exp.h"
//#include "map/scl/sclCon.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Sfm_Dec_t struct {
    // external
    pPars        *Sfm_Par_t;       // parameters
    pLib        *Sfm_Lib_t;        // library
    pTim        *Sfm_Tim_t;        // timing
    pMit        *Sfm_Mit_t;        // timing
    pNtk        *Abc_Ntk_t;        // network
    // library
    vGateSizes         Vec_Int_t;  // fanin counts
    vGateFuncs         Vec_Wrd_t;  // gate truth tables
    vGateCnfs         Vec_Wec_t;   // gate CNFs
    vGateHands         Vec_Ptr_t;  // gate handles
    GateConst0               int32;  // special gates
    GateConst1               int32;  // special gates
    GateBuffer               int32;  // special gates
    GateInvert               int32;  // special gates
    GateAnd [4]int32;  // special gates
    GateOr [4]int32;   // special gates
    // objects
    nDivs               int32;       // the number of divisors
    nMffc               int32;       // the number of divisors
    AreaMffc               int32;    // the area of gates in MFFC
    DelayMin               int32;    // temporary min delay
    iTarget               int32;     // target node
    iUseThis               int32;    // next cofactoring var to try
    DeltaCrit               int32;   // critical delta
    AreaInv               int32;     // inverter area
    DelayInv               int32;    // inverter delay
    pGateInv       *Mio_Gate_t;    // inverter
    uCareSet              word;    // computed careset
    vObjRoots         Vec_Int_t;   // roots of the window
    vObjGates         Vec_Int_t;   // functionality
    vObjFanins         Vec_Wec_t;  // fanin IDs
    vObjMap         Vec_Int_t;     // object map
    vObjDec         Vec_Int_t;     // decomposition
    vObjMffc         Vec_Int_t;    // MFFC nodes
    vObjInMffc         Vec_Int_t;  // inputs of MFFC nodes
    vObjSims         Vec_Wrd_t;    // simulation patterns
    vObjSims2         Vec_Wrd_t;   // simulation patterns
    vMatchGates         Vec_Ptr_t; // matched gates
    vMatchFans         Vec_Ptr_t;  // matched fanins
    // solver
    pSat       *sat_solver;        // reusable solver 
    vClauses         Vec_Wec_t;    // CNF clauses for the node
    Vec_Int_t         vImpls[2];   // onset/offset implications
    Vec_Wrd_t         vSets[2];    // onset/offset patterns
    nPats [2]int32;    // CEX count
    nPatWords [2]int32;// CEX words
    nDivWords               int32;   // div words
    nDivWordsAlloc               int32; // div words
    word              TtElems[SFM_SUPP_MAX][SFM_WORD_MAX];
    word *            pTtElems[SFM_SUPP_MAX];
    word *            pDivWords[SFM_SUPP_MAX];
    // temporary
    vNewNodes         Vec_Int_t;
    vGateTfi         Vec_Int_t;
    vGateTfo         Vec_Int_t;
    vGateCut         Vec_Int_t;
    vGateTemp         Vec_Int_t;
    vGateMffc         Vec_Int_t;
    vCands         Vec_Int_t;
    word              Copy[4];
    nSuppVars               int32;
    // statistics
    timeLib           abctime;
    timeWin           abctime;
    timeCnf           abctime;
    timeSat           abctime;
    timeSatSat           abctime;
    timeSatUnsat           abctime;
    timeEval           abctime;
    timeTime           abctime;
    timeOther           abctime;
    timeStart           abctime;
    timeTotal           abctime;
    nTotalNodesBeg               int32;
    nTotalEdgesBeg               int32;
    nTotalNodesEnd               int32;
    nTotalEdgesEnd               int32;
    nNodesTried               int32;
    nNodesChanged               int32;
    nNodesConst0               int32;
    nNodesConst1               int32;
    nNodesBuf               int32;
    nNodesInv               int32;
    nNodesAndOr               int32;
    nNodesResyn               int32;
    nSatCalls               int32;
    nSatCallsSat               int32;
    nSatCallsUnsat               int32;
    nSatCallsOver               int32;
    nTimeOuts               int32;
    nNoDecs               int32;
    nEfforts               int32;
    nMaxDivs               int32;
    nMaxWin               int32;
    nAllDivs              word;
    nAllWin              word;
    nLuckySizes [SFM_SUPP_MAX+1]int32;
    nLuckyGates [SFM_SUPP_MAX+1]int32;
}

const SFM_MASK_PI =     1  // supp(node) is contained in supp(TFI(pivot))
const SFM_MASK_INPUT =  2  // supp(node) does not overlap with supp(TFI(pivot))
const SFM_MASK_FANIN =  4  // the same as above (pointed to by node with SFM_MASK_PI | SFM_MASK_INPUT)
const SFM_MASK_MFFC =   8  // MFFC nodes, including the target node
const SFM_MASK_PIVOT = 16  // the target node

func Sfm_DecMan(p *Abc_Obj_t) *Sfm_Dec_t { return (Sfm_Dec_t *)p.pNtk.pData;                  }
func Sfm_DecObjSim(p *Sfm_Dec_t, pObj *Abc_Obj_t) word { return Vec_WrdEntry(&p.vObjSims, Abc_ObjId(pObj));  }
func Sfm_DecObjSim2(p *Sfm_Dec_t, pObj *Abc_Obj_t) word { return Vec_WrdEntry(&p.vObjSims2, Abc_ObjId(pObj)); }
func Sfm_DecDivPats(p *Sfm_Dec_t, d int32, c int32) *word { return Vec_WrdEntryP(&p.vSets[c], d*SFM_SIM_WORDS); }

func Sfm_ManReadObjDelay(p *Sfm_Dec_t, Id int32) int32 { if p.pMit {
    return Sfm_MitReadObjDelay(p.pMit, Id);
} else {
    return Sfm_TimReadObjDelay(p.pTim, Id);
} }
func Sfm_ManReadNtkDelay(p *Sfm_Dec_t) int32 { if p.pMit {
    return Sfm_MitReadNtkDelay(p.pMit);
} else {
    return Sfm_TimReadNtkDelay(p.pTim);
}     }
func Sfm_ManReadNtkMinSlack(p *Sfm_Dec_t) int32 { if p.pMit {
    return Sfm_MitReadNtkMinSlack(p.pMit);
} else {
    return 0;
}                                }



////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Setup parameter structure.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_DecStart(pPars *Sfm_Par_t, pLib *Mio_Library_t, pNtk *Abc_Ntk_t) *Sfm_Dec_t {
    extern void Sfm_LibPreprocess( Mio_Library_t * pLib, Vec_Int_t * vGateSizes, Vec_Wrd_t * vGateFuncs, Vec_Wec_t * vGateCnfs, Vec_Ptr_t * vGateHands );
    // inlining of macro ABC_CALLOC
    var p *Sfm_Dec_t  = &Sfm_Dec_t{}; var i int32;
    p.timeStart = Abc_Clock();
    p.pPars     = pPars;
    p.pNtk      = pNtk;
    p.pSat      = sat_solver_new();
    p.pGateInv  = Mio_LibraryReadInv( pLib );
    p.AreaInv   = Scl_Flt2Int(Mio_GateReadArea(p.pGateInv));
    p.DelayInv  = Scl_Flt2Int(Mio_GateReadDelayMax(p.pGateInv));
    if pPars.DeltaCrit {
        p.DeltaCrit = Scl_Flt2Int(float32(pPars.DeltaCrit));
    } else {
        p.DeltaCrit = 5 * Scl_Flt2Int(Mio_LibraryReadDelayInvMax(pLib)) / 2;
    }
p.timeLib = Abc_Clock();
    p.pLib = Sfm_LibPrepare( pPars.nVarMax, 1, !pPars.fArea, pPars.fVerbose, pPars.fLibVerbose );
p.timeLib = Abc_Clock() - p.timeLib;
    if !pPars.fArea {
        if Abc_FrameReadLibScl() {
            p.pMit = Sfm_MitStart( pLib, (SC_Lib *)Abc_FrameReadLibScl(), Scl_ConReadMan(), pNtk, p.DeltaCrit ); }
        if p.pMit == nil {
            p.pTim = Sfm_TimStart( pLib, Scl_ConReadMan(), pNtk, p.DeltaCrit ); }
    }
    if pPars.fVeryVerbose {
//    if ( pPars->fVerbose )
        Sfm_LibPrint( p.pLib ); }
    pNtk.pData = p;
    // enter library
    assert(Abc_NtkIsMappedLogic(pNtk), "Abc_NtkIsMappedLogic(pNtk)");
    Sfm_LibPreprocess( pLib, &p.vGateSizes, &p.vGateFuncs, &p.vGateCnfs, &p.vGateHands );
    p.GateConst0 = Mio_GateReadValue( Mio_LibraryReadConst0(pLib) );
    p.GateConst1 = Mio_GateReadValue( Mio_LibraryReadConst1(pLib) );
    p.GateBuffer = Mio_GateReadValue( Mio_LibraryReadBuf(pLib) );
    p.GateInvert = Mio_GateReadValue( Mio_LibraryReadInv(pLib) );
    // elementary truth tables
    for  i = 0; i < SFM_SUPP_MAX; i++  {
        p.pTtElems[i] = p.TtElems[i];
    }
    Abc_TtElemInit( p.pTtElems, SFM_SUPP_MAX );
    p.iUseThis = -1;
    return p;
}
func Sfm_DecStop(p *Sfm_Dec_t) {
    var pNtk *Abc_Ntk_t  = p.pNtk;
    var pObj *Abc_Obj_t; var i int32;
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else
        if int32(pObj.Level) != Abc_ObjLevelNew(pObj) {
            fmt.Printf("Level count mismatch at node %d.\n", i); }
    }
    Sfm_LibStop( p.pLib );
    if p.pTim { Sfm_TimStop( p.pTim ); }
    if p.pMit { Sfm_MitStop( p.pMit ); }
    // divisors
    for  i = 0; i < SFM_SUPP_MAX; i++  {
        p.pDivWords[i] = nil;
    }
    // library
    Vec_IntErase( &p.vGateSizes );
    Vec_WrdErase( &p.vGateFuncs );
    Vec_WecErase( &p.vGateCnfs );
    Vec_PtrErase( &p.vGateHands );
    // objects
    Vec_IntErase( &p.vObjRoots );
    Vec_IntErase( &p.vObjGates );
    Vec_WecErase( &p.vObjFanins );
    Vec_IntErase( &p.vObjMap );
    Vec_IntErase( &p.vObjDec );
    Vec_IntErase( &p.vObjMffc );
    Vec_IntErase( &p.vObjInMffc );
    Vec_WrdErase( &p.vObjSims );
    Vec_WrdErase( &p.vObjSims2 );
    Vec_PtrErase( &p.vMatchGates );
    Vec_PtrErase( &p.vMatchFans );
    // solver
    sat_solver_delete( p.pSat );
    Vec_WecErase( &p.vClauses );
    Vec_IntErase( &p.vImpls[0] );
    Vec_IntErase( &p.vImpls[1] );
    Vec_WrdErase( &p.vSets[0] );
    Vec_WrdErase( &p.vSets[1] );
    // temporary
    Vec_IntErase( &p.vNewNodes );
    Vec_IntErase( &p.vGateTfi );
    Vec_IntErase( &p.vGateTfo );
    Vec_IntErase( &p.vGateCut );
    Vec_IntErase( &p.vGateTemp );
    Vec_IntErase( &p.vGateMffc );
    Vec_IntErase( &p.vCands );
    p = nil;
    pNtk.pData = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_ObjSimulate(pObj *Abc_Obj_t) word {
    var p *Sfm_Dec_t  = Sfm_DecMan( pObj );
    var vExpr *Vec_Int_t  = Mio_GateReadExpr( (Mio_Gate_t *)pObj.pData );
    var pFanin *Abc_Obj_t; var i int32; var uFanins [6]word;
    assert(Abc_ObjFaninNum(pObj) <= 6, "Abc_ObjFaninNum(pObj) <= 6");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        uFanins[i] = Sfm_DecObjSim( p, pFanin );
    }
    return Exp_Truth6( Abc_ObjFaninNum(pObj), vExpr, uFanins );
}
func Sfm_ObjSimulate2(pObj *Abc_Obj_t) word {
    var p *Sfm_Dec_t  = Sfm_DecMan( pObj );
    var vExpr *Vec_Int_t  = Mio_GateReadExpr( (Mio_Gate_t *)pObj.pData );
    var pFanin *Abc_Obj_t; var i int32; var uFanins [6]word;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if pFanin.iTemp & SFM_MASK_PIVOT {
            uFanins[i] = Sfm_DecObjSim2( p, pFanin );
        } else {
            uFanins[i] = Sfm_DecObjSim( p, pFanin ); }
    }
    return Exp_Truth6( Abc_ObjFaninNum(pObj), vExpr, uFanins );
}
func Sfm_NtkSimulate(pNtk *Abc_Ntk_t) {
    var vNodes *Vec_Ptr_t;
    var pObj *Abc_Obj_t; var i int32; var uTemp word;
    var p *Sfm_Dec_t  = Sfm_DecMan( Abc_NtkPi(pNtk, 0) );
    Vec_WrdFill( &p.vObjSims,  2*Abc_NtkObjNumMax(pNtk), 0 );
    Vec_WrdFill( &p.vObjSims2, 2*Abc_NtkObjNumMax(pNtk), 0 );
    Gia_ManRandomW(1);
    assert(p.pPars.fUseSim, "p.pPars.fUseSim");
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        Vec_WrdWriteEntry( &p.vObjSims, Abc_ObjId(pObj), (uTemp = Gia_ManRandomW(0)) );
        //printf( "Inpt = %5d : ", Abc_ObjId(pObj) );
        //Extra_PrintBinary( stdout, (unsigned *)&uTemp, 64 );
        //printf( "\n" );
    }
    vNodes = Abc_NtkDfs( pNtk, 1 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vNodes)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vNodes, i)), 1); i++  {
        Vec_WrdWriteEntry( &p.vObjSims, Abc_ObjId(pObj), (uTemp = Sfm_ObjSimulate(pObj)) );
        //printf( "Obj  = %5d : ", Abc_ObjId(pObj) );
        //Extra_PrintBinary( stdout, (unsigned *)&uTemp, 64 );
        //printf( "\n" );
    }
    Vec_PtrFree( vNodes );
}
func Sfm_ObjSimulateNode(pObj *Abc_Obj_t) {
    var p *Sfm_Dec_t  = Sfm_DecMan( pObj );
    if !p.pPars.fUseSim {
        return; }
    Vec_WrdWriteEntry( &p.vObjSims, Abc_ObjId(pObj), Sfm_ObjSimulate(pObj) );
    if pObj.iTemp & SFM_MASK_PIVOT {
        Vec_WrdWriteEntry( &p.vObjSims2, Abc_ObjId(pObj), Sfm_ObjSimulate2(pObj) ); }
}
func Sfm_ObjFlipNode(pObj *Abc_Obj_t) {
    var p *Sfm_Dec_t  = Sfm_DecMan( pObj );
    if !p.pPars.fUseSim {
        return; }
    Vec_WrdWriteEntry( &p.vObjSims2, Abc_ObjId(pObj), ~Sfm_DecObjSim(p, pObj) );
}
func Sfm_ObjFindCareSet(pNtk *Abc_Ntk_t, vRoots *Vec_Int_t) word {
    var p *Sfm_Dec_t  = Sfm_DecMan( Abc_NtkPi(pNtk, 0) );
    var pObj *Abc_Obj_t; var i int32; var Res word;
    if !p.pPars.fUseSim {
        return 0; }
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vRoots,i))), 1); i++  {         if (pObj) == nil {} else {
        Res |= Sfm_DecObjSim(p, pObj) ^ Sfm_DecObjSim2(p, pObj); }
    }
    return Res;
}
func Sfm_ObjSetupSimInfo(pObj *Abc_Obj_t) {
    var p *Sfm_Dec_t  = Sfm_DecMan( pObj ); var i int32;
    p.nPats[0]     = p.nPats[1] = 0;
    p.nPatWords[0] = p.nPatWords[1] = 0;
    Vec_WrdFill( &p.vSets[0], p.nDivs*SFM_SIM_WORDS, 0 );
    Vec_WrdFill( &p.vSets[1], p.nDivs*SFM_SIM_WORDS, 0 );
    // alloc divwords
    p.nDivWords = Abc_Bit6WordNum( 4 * p.nDivs );
    if p.nDivWordsAlloc < p.nDivWords {
        p.nDivWordsAlloc = Abc_MaxInt( 16, p.nDivWords );
        for  i = 0; i < SFM_SUPP_MAX; i++  {
            // inlining of macro ABC_REALLOC
            p.pDivWords[i] = ((p.pDivWords[i]) ? ((word *) realloc((char *)(p.pDivWords[i]), sizeof(word) * (size_t)(p.nDivWordsAlloc))) :          ((word *) malloc(sizeof(word) * (size_t)(p.nDivWordsAlloc))));
        }
    }
    memset( p.pDivWords[0], 0, sizeof(word) * p.nDivWords );
    // collect simulation info
    if p.pPars.fUseSim && p.uCareSet != 0 {
        var uCareSet word  = p.uCareSet;
        var uValues word   = Sfm_DecObjSim(p, pObj);
        var c int32;
        var d int32;
        var i int32;
        var Indexes [2][64]int32;
        assert(p.iTarget == pObj.iTemp, "p.iTarget == pObj.iTemp");
        assert(p.pPars.fUseSim, "p.pPars.fUseSim");
        // find what patterns go to on-set/off-set
        for  i = 0; i < 64; i++  {
            if (uCareSet >> i) & 1 {
                c = !((uValues >> i) & 1);
                Indexes[c][p.nPats[c]++] = i;
            }
        }
        for  c = 0; c < 2; c++  {
            p.nPatWords[c] = 1 + (p.nPats[c] >> 6);
        }
        // write patterns
        for  d = 0; d < p.nDivs; d++  {
            var uSim word  = Vec_WrdEntry( &p.vObjSims, Vec_IntEntry(&p.vObjMap, d) );
            for  c = 0; c < 2; c++  {
                for  i = 0; i < p.nPats[c]; i++  {
                    if (uSim >> Indexes[c][i]) & 1 {
                        Abc_TtSetBit( Sfm_DecDivPats(p, d, c), i ); }
                }
            }
        }
        //printf( "Node %d : Onset = %d. Offset = %d.\n", pObj->Id, p->nPats[0], p->nPats[1] );
    }
}
func Sfm_ObjSetdownSimInfo(pObj *Abc_Obj_t) {
    var nPatKeep int32  = 32;
    var p *Sfm_Dec_t  = Sfm_DecMan( pObj );
    var c, d int32; var uSim word;
var uSims [2]word;
var uMask word;
    if !p.pPars.fUseSim {
        return; }
    for  d = 0; d < p.nDivs; d++  {
        uSim = Vec_WrdEntry( &p.vObjSims, Vec_IntEntry(&p.vObjMap, d) );
        for  c = 0; c < 2; c++  {
            uMask = Abc_Tt6Mask( Abc_MinInt(p.nPats[c], nPatKeep) );
            uSims[c] = (Sfm_DecDivPats(p, d, c)[0] & uMask) | (uSim & ~uMask);
            uSim >>= 32;
        }
        uSim = (uSims[0] & 0xFFFFFFFF) | (uSims[1] << 32);
        Vec_WrdWriteEntry( &p.vObjSims, Vec_IntEntry(&p.vObjMap, d), uSim );
    }
}
/*
void Sfm_ObjSetdownSimInfo( Abc_Obj_t * pObj )
{
    int nPatKeep = 32;
    Sfm_Dec_t * p = Sfm_DecMan( pObj );
    word uSim, uMaskKeep[2];
    int c, d, nKeeps[2]; 
    for ( c = 0; c < 2; c++ )
    {
        nKeeps[c] = Abc_MaxInt(p->nPats[c], nPatKeep);
        uMaskKeep[c] = Abc_Tt6Mask( nKeeps[c] );
    }
    for ( d = 0; d < p->nDivs; d++ )
    {
        uSim = Vec_WrdEntry( &p->vObjSims, Vec_IntEntry(&p->vObjMap, d) ) << (nKeeps[0] + nKeeps[1]);
        uSim |= (Vec_WrdEntry(&p->vSets[0], d) & uMaskKeep[0]) | ((Vec_WrdEntry(&p->vSets[1], d) & uMaskKeep[1]) << nKeeps[0]);
        Vec_WrdWriteEntry( &p->vObjSims, Vec_IntEntry(&p->vObjMap, d), uSim );
    }
}
*/

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_DecPrepareSolver(p *Sfm_Dec_t) int32 {
    var vRoots *Vec_Int_t  = &p.vObjRoots;
    var vFaninVars *Vec_Int_t  = &p.vGateTemp;
    var vLevel, vClause *Vec_Int_t;
    var i, k, Gate, iObj, RetValue int32;
    var nTfiSize int32  = p.iTarget + 1; // including node
    var nWinSize int32  = Vec_IntSize(&p.vObjGates);
    var nSatVars int32  = 2 * nWinSize - nTfiSize;
    assert(nWinSize == Vec_IntSize(&p.vObjGates), "nWinSize == Vec_IntSize(&p.vObjGates)");
    assert(p.iTarget < nWinSize, "p.iTarget < nWinSize");
    // create SAT solver
    sat_solver_restart( p.pSat );
    sat_solver_setnvars( p.pSat, nSatVars + Vec_IntSize(vRoots) );
    // add CNF clauses for the TFI
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vObjGates)) && (((Gate) = Vec_IntEntry(&p.vObjGates, i)), 1); i++  {
        if Gate == -1 {
            continue; }
        // generate CNF 
        vLevel = Vec_WecEntry( &p.vObjFanins, i );
        Vec_IntPush( vLevel, i );
        Sfm_TranslateCnf( &p.vClauses, (Vec_Str_t *)Vec_WecEntry(&p.vGateCnfs, Gate), vLevel, -1 );
        Vec_IntPop( vLevel );
        // add clauses
        // inlining of macro Vec_WecForEachLevel
        for  k = 0; (k < Vec_WecSize(&p.vClauses)) && (((vClause) = Vec_WecEntry(&p.vClauses, k)), 1); k++  {
            if Vec_IntSize(vClause) == 0 {
                break; }
            RetValue = sat_solver_addclause( p.pSat, Vec_IntArray(vClause), Vec_IntArray(vClause) + Vec_IntSize(vClause) );
            if RetValue == 0 {
                return 0; }
        }
    }
    // add CNF clauses for the TFO
    // inlining of macro Vec_IntForEachEntryStart
    for  i = nTfiSize; (i < Vec_IntSize(&p.vObjGates)) && (((Gate) = Vec_IntEntry(&p.vObjGates, i)), 1); i++  {
        assert(Gate != -1, "Gate != -1");
        vLevel = Vec_WecEntry( &p.vObjFanins, i );
        Vec_IntClear( vFaninVars );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iObj) = Vec_IntEntry(vLevel, k)), 1); k++  {
            Vec_IntPush( vFaninVars, iObj <= p.iTarget ? iObj : iObj + nWinSize - nTfiSize );
        }
        Vec_IntPush( vFaninVars, i + nWinSize - nTfiSize );
        // generate CNF 
        Sfm_TranslateCnf( &p.vClauses, (Vec_Str_t *)Vec_WecEntry(&p.vGateCnfs, Gate), vFaninVars, p.iTarget );
        // add clauses
        // inlining of macro Vec_WecForEachLevel
        for  k = 0; (k < Vec_WecSize(&p.vClauses)) && (((vClause) = Vec_WecEntry(&p.vClauses, k)), 1); k++  {
            if Vec_IntSize(vClause) == 0 {
                break; }
            RetValue = sat_solver_addclause( p.pSat, Vec_IntArray(vClause), Vec_IntArray(vClause) + Vec_IntSize(vClause) );
            if RetValue == 0 {
                return 0; }
        }
    }
    if nTfiSize < nWinSize {
        // create XOR clauses for the roots
        Vec_IntClear( vFaninVars );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vRoots)) && (((iObj) = Vec_IntEntry(vRoots, i)), 1); i++  {
            Vec_IntPush( vFaninVars, Abc_Var2Lit(nSatVars, 0) );
            sat_solver_add_xor( p.pSat, iObj, iObj + nWinSize - nTfiSize, nSatVars++, 0 );
        }
        // make OR clause for the last nRoots variables
        RetValue = sat_solver_addclause( p.pSat, Vec_IntArray(vFaninVars), Vec_IntLimit(vFaninVars) );
        if RetValue == 0 {
            return 0; }
        assert(nSatVars == sat_solver_nvars(p.pSat), "nSatVars == sat_solver_nvars(p.pSat)");
    
    } else { assert(Vec_IntSize(vRoots) == 1, "Vec_IntSize(vRoots) == 1"); }
    // finalize
    RetValue = sat_solver_simplify( p.pSat );
    return 1;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_DecFindCost(p *Sfm_Dec_t, c int32, iLit int32, pMask *word) int32 {
    var pPats *word  = Sfm_DecDivPats( p, Abc_Lit2Var(iLit), !c );
    return Abc_TtCountOnesVecMask( pPats, pMask, p.nPatWords[!c], Abc_LitIsCompl(iLit) );
}
func Sfm_DecPrint(p *Sfm_Dec_t, Masks [][SFM_SIM_WORDS]word) {
    var c, i, k, Entry int32;
    for  c = 0; c < 2; c++  {
        var vLevel *Vec_Int_t  = Vec_WecEntry( &p.vObjFanins, p.iTarget );
        fmt.Printf("%s-SET of object %d (divs = %d) with gate \"%s\" and fanins: ", c ? "OFF": "ON", p.iTarget, p.nDivs, Mio_GateReadName((Mio_Gate_t *)Vec_PtrEntry(&p.vGateHands, Vec_IntEntry(&p.vObjGates,p.iTarget))));
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vLevel)) && (((Entry) = Vec_IntEntry(vLevel, i)), 1); i++  {
            fmt.Printf("%d ", Entry);
        }
        fmt.Printf("\n");

        fmt.Printf("Implications: ");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(&p.vImpls[c])) && (((Entry) = Vec_IntEntry(&p.vImpls[c], i)), 1); i++  {
            fmt.Printf("%s%d(%d) ", Abc_LitIsCompl(Entry)? "!":"", Abc_Lit2Var(Entry), Sfm_DecFindCost(p, c, Entry, Masks[!c]));
        }
        fmt.Printf("\n");
        fmt.Printf("     ");
        for  i = 0; i < p.nDivs; i++  {
            fmt.Printf("%d", (i / 10) % 10);
        }
        fmt.Printf("\n");
        fmt.Printf("     ");
        for  i = 0; i < p.nDivs; i++  {
            fmt.Printf("%d", i % 10);
        }
        fmt.Printf("\n");
        for  k = 0; k < p.nPats[c]; k++  {
            fmt.Printf("%2d : ", k);
            for  i = 0; i < p.nDivs; i++  {
                fmt.Printf("%d", Abc_TtGetBit(Sfm_DecDivPats(p, i, c), k));
            }
            fmt.Printf("\n");
        }
        //printf( "\n" );
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_DecVarCost(p *Sfm_Dec_t, Masks [][SFM_SIM_WORDS]word, d int32, Counts [][2]int32) {
    var c int32; 
    for  c = 0; c < 2; c++  {
        var pPats *word  = Sfm_DecDivPats( p, d, c );
        var Num int32  = Abc_TtCountOnesVec( Masks[c], p.nPatWords[c] );
        Counts[c][1] = Abc_TtCountOnesVecMask( pPats, Masks[c], p.nPatWords[c], 0 );
        Counts[c][0] = Num - Counts[c][1];
        assert(Counts[c][0] >= 0 && Counts[c][1] >= 0, "Counts[c][0] >= 0 && Counts[c][1] >= 0");
    }
    //printf( "%5d %5d   %5d %5d \n", Counts[0][0], Counts[0][1], Counts[1][0], Counts[1][1] );
}
func Sfm_DecFindBestVar2(p *Sfm_Dec_t, Masks [][SFM_SIM_WORDS]word) int32 {
    var Counts [2][2]int32;
    var d int32;
    var VarBest int32 = -1;
    var CostBest int32 = ABC_INFINITY;
    var Cost int32;
    for  d = 0; d < p.nDivs; d++  {
        Sfm_DecVarCost( p, Masks, d, Counts );
        if (Counts[0][0] < Counts[0][1]) == (Counts[1][0] < Counts[1][1]) {
            continue; }
        Cost = Abc_MinInt(Counts[0][0], Counts[0][1]) + Abc_MinInt(Counts[1][0], Counts[1][1]);
        if CostBest > Cost {
            CostBest = Cost;
            VarBest = d;
        }
    }
    return VarBest;
}
func Sfm_DecFindBestVar(p *Sfm_Dec_t, Masks [][SFM_SIM_WORDS]word) int32 {
    var c int32;
    var i int32;
    var iLit int32;
    var Var int32 = -1;
    var Cost int32;
    var CostMin int32 = ABC_INFINITY;
    for  c = 0; c < 2; c++  {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(&p.vImpls[c])) && (((iLit) = Vec_IntEntry(&p.vImpls[c], i)), 1); i++  {
            if Vec_IntSize(&p.vImpls[c]) > 1 && Vec_IntFind(&p.vObjDec, Abc_Lit2Var(iLit)) >= 0 {
                continue; }
            Cost = Sfm_DecFindCost( p, c, iLit, Masks[!c] );
            if CostMin > Cost {
                CostMin = Cost;
                Var = Abc_Lit2Var(iLit);
            }
        }
    }
    return Var;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_DecMffcArea(pNtk *Abc_Ntk_t, vMffc *Vec_Int_t) int32 {
    var pObj *Abc_Obj_t; 
    var i, nAreaMffc int32;
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vMffc) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vMffc,i))), 1); i++  {         if (pObj) == nil {} else {
        nAreaMffc += Scl_Flt2Int(Mio_GateReadArea((Mio_Gate_t *)pObj.pData)); }
    }
    return nAreaMffc;
}
func Sfm_MffcDeref_rec(pObj *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    var Area int32 = Scl_Flt2Int(Mio_GateReadArea((Mio_Gate_t *)pObj.pData));
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        assert(pFanin.vFanouts.nSize > 0, "pFanin.vFanouts.nSize > 0");
        if --pFanin.vFanouts.nSize == 0 && !Abc_ObjIsCi(pFanin) {
            Area += Sfm_MffcDeref_rec( pFanin ); }
    }
    return Area;
}
func Sfm_MffcRef_rec(pObj *Abc_Obj_t, vMffc *Vec_Int_t) int32 {
    var pFanin *Abc_Obj_t;
    var i int32;
    var Area int32 = Scl_Flt2Int(Mio_GateReadArea((Mio_Gate_t *)pObj.pData));
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if pFanin.vFanouts.nSize++ == 0 && !Abc_ObjIsCi(pFanin) {
            Area += Sfm_MffcRef_rec( pFanin, vMffc ); }
    }
    if vMffc != nil {
        Vec_IntPush( vMffc, Abc_ObjId(pObj) ); }
    return Area;
}
func Sfm_DecMffcAreaReal(pPivot *Abc_Obj_t, vCut *Vec_Int_t, vMffc *Vec_Int_t) int32 {
    var pNtk *Abc_Ntk_t  = pPivot.pNtk;
    var pObj *Abc_Obj_t; 
    var i, Area1, Area2 int32;
    assert(Abc_ObjIsNode(pPivot), "Abc_ObjIsNode(pPivot)");
    if vMffc != nil {
        Vec_IntClear( vMffc ); }
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCut) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vCut,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.vFanouts.nSize++; }
    }
    Area1 = Sfm_MffcDeref_rec( pPivot );
    Area2 = Sfm_MffcRef_rec( pPivot, vMffc );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vCut) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vCut,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.vFanouts.nSize--; }
    }
    assert(Area1 == Area2, "Area1 == Area2");
    return Area1;
}
func Sfm_DecPrepareVec(vMap *Vec_Int_t, pNodes *int32, nNodes int32, vCut *Vec_Int_t) {
    var i int32;
    Vec_IntClear( vCut );
    for  i = 0; i < nNodes; i++  {
        Vec_IntPush( vCut, Vec_IntEntry(vMap, pNodes[i]) );
    }    
}
func Sfm_DecComputeFlipInvGain(p *Sfm_Dec_t, pPivot *Abc_Obj_t, pfNeedInv *int32) int32 {
    var pFanout *Abc_Obj_t; 
    var pGate, pGateNew *Mio_Gate_t;
    var i, Handle, fNeedInv, Gain int32;
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pPivot)) && (((pFanout) = Abc_ObjFanout(pPivot, i)), 1); i++  {
        if !Abc_ObjIsNode(pFanout) {
            fNeedInv = 1;
            continue;
        }
        pGate = (Mio_Gate_t*)pFanout.pData;
        if Abc_ObjFaninNum(pFanout) == 1 && Mio_GateIsInv(pGate) {
            Gain += p.AreaInv;
            continue;
        }
        Handle = Sfm_LibFindComplInputGate( &p.vGateFuncs, Mio_GateReadValue(pGate), Abc_ObjFaninNum(pFanout), Abc_NodeFindFanin(pFanout, pPivot), nil );
        if Handle == -1 {
            fNeedInv = 1;
            continue;
        }
        pGateNew = (Mio_Gate_t *)Vec_PtrEntry( &p.vGateHands, Handle );
        Gain += Scl_Flt2Int(Mio_GateReadArea(pGate)) - Scl_Flt2Int(Mio_GateReadArea(pGateNew));
    }
    if fNeedInv {
        Gain -= p.AreaInv; }
    if pfNeedInv != nil {
        *pfNeedInv = fNeedInv; }
    return Gain;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_DecCombineDec(p *Sfm_Dec_t, pTruth0 *word, pTruth1 *word, pSupp0 *int32, pSupp1 *int32, nSupp0 int32, nSupp1 int32, pTruth *word, pSupp *int32, Var int32) int32 {
    var vVec0 Vec_Int_t  = Vec_Int_t{ 2*SFM_SUPP_MAX, nSupp0, pSupp0 };
    var vVec1 Vec_Int_t  = Vec_Int_t{ 2*SFM_SUPP_MAX, nSupp1, pSupp1 };
    var vVec Vec_Int_t   = Vec_Int_t{ 2*SFM_SUPP_MAX, 0,      pSupp  };
    var nWords0 int32  = Abc_TtWordNum(nSupp0);
    var nSupp, iSuppVar int32;
    // check the case of equal cofactors
    if nSupp0 == nSupp1 && !memcmp(pSupp0, pSupp1, sizeof(int)*nSupp0) && !memcmp(pTruth0, pTruth1, sizeof(word)*nWords0) {
        memcpy( pSupp,  pSupp0,  sizeof(int)*nSupp0   );
        memcpy( pTruth, pTruth0, sizeof(word)*nWords0 );
        Abc_TtStretch6( pTruth, nSupp0, p.pPars.nVarMax );
        return nSupp0;
    }
    // merge support variables
    Vec_IntTwoMerge2Int( &vVec0, &vVec1, &vVec );
    Vec_IntPushOrder( &vVec, Var );
    nSupp = Vec_IntSize( &vVec );
    if nSupp > p.pPars.nVarMax {
        return -2; }
    // expand truth tables
    Abc_TtStretch6( pTruth0, nSupp0, nSupp );
    Abc_TtStretch6( pTruth1, nSupp1, nSupp );
    Abc_TtExpand( pTruth0, nSupp, pSupp0, nSupp0, pSupp, nSupp );
    Abc_TtExpand( pTruth1, nSupp, pSupp1, nSupp1, pSupp, nSupp );
    // perform operation
    iSuppVar = Vec_IntFind( &vVec, Var );
    Abc_TtMux( pTruth, p.pTtElems[iSuppVar], pTruth1, pTruth0, Abc_TtWordNum(nSupp) );
    Abc_TtStretch6( pTruth, nSupp, p.pPars.nVarMax );
    return nSupp;
}
func Sfm_DecPeformDec_rec(p *Sfm_Dec_t, pTruth *word, pSupp *int32, pAssump *int32, nAssump int32, Masks [][SFM_SIM_WORDS]word, fCofactor int32, nSuppAdd int32) int32 {
    var nBTLimit int32  = p.pPars.nBTLimit;
//    int fVerbose = p->pPars->fVeryVerbose;
    var c int32;
    var i int32;
    var d int32;
    var iLit int32;
    var status int32;
    var Var int32 = -1;
    var pDivWords *word  = p.pDivWords[nAssump];
    var clk abctime;
    assert(nAssump <= SFM_SUPP_MAX, "nAssump <= SFM_SUPP_MAX");
    if p.pPars.fVeryVerbose {
        fmt.Printf("\nObject %d\n", p.iTarget);
        fmt.Printf("Divs = %d.  Nodes = %d.  Mffc = %d.  Mffc area = %.2f.    ", p.nDivs, Vec_IntSize(&p.vObjGates), p.nMffc, Scl_Int2Flt(p.AreaMffc));
        fmt.Printf("Pat0 = %d.  Pat1 = %d.    ", p.nPats[0], p.nPats[1]);
        fmt.Printf("\n");
        if nAssump != 0 {
            fmt.Printf("Cofactor: ");
            for  i = 0; i < nAssump; i++  {
                fmt.Printf(" %s%d", Abc_LitIsCompl(pAssump[i])? "!":"", Abc_Lit2Var(pAssump[i]));
            }
            fmt.Printf("\n");
        }
    }
    // check constant
    for  c = 0; c < 2; c++  {
        if !Abc_TtIsConst0(Masks[c], p.nPatWords[c]) { // there are some patterns
            continue; }
        p.nSatCalls++;
        pAssump[nAssump] = Abc_Var2Lit( p.iTarget, c );
        clk = Abc_Clock();
        status = sat_solver_solve( p.pSat, pAssump, pAssump + nAssump + 1, nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            p.nTimeOuts++;
            return -2;
        }
        if status == l_False {
            p.nSatCallsUnsat++;
            p.timeSatUnsat += Abc_Clock() - clk;
            Abc_TtConst( pTruth, Abc_TtWordNum(p.pPars.nVarMax), c );
            if p.pPars.fVeryVerbose {
                fmt.Printf("Found constant %d.\n", c); }
            return 0;
        }
        assert(status == l_True, "status == l_True");
        p.nSatCallsSat++;
        p.timeSatSat += Abc_Clock() - clk;
        if p.nPats[c] == 64*SFM_SIM_WORDS {
            p.nSatCallsOver++;
            continue;//return -2;//continue;
        }
        for  i = 0; i < p.nDivs; i++  {
            if sat_solver_var_value(p.pSat, i) {
                Abc_TtSetBit( Sfm_DecDivPats(p, i, c), p.nPats[c] ); }
        }
        p.nPatWords[c] = 1 + (p.nPats[c] >> 6);
        Abc_TtSetBit( Masks[c], p.nPats[c]++ );
    }

    if p.iUseThis != -1 {
        Var = p.iUseThis;
        p.iUseThis = -1;
        goto cofactor;
    }

    // check implications
    Vec_IntClear( &p.vImpls[0] );
    Vec_IntClear( &p.vImpls[1] );
    for  d = 0; d < p.nDivs; d++  {
        var Impls [2]int32 = [2]int32{-1, -1};
        for  c = 0; c < 2; c++  {
            var pPats *word  = Sfm_DecDivPats( p, d, c );
            var fHas0s int32  = Abc_TtIntersect( pPats, Masks[c], p.nPatWords[c], 1 );
            var fHas1s int32  = Abc_TtIntersect( pPats, Masks[c], p.nPatWords[c], 0 );
            if (fHas0s != 0) && (fHas1s != 0) {
                continue; }
            pAssump[nAssump]   = Abc_Var2Lit( p.iTarget, c );
            pAssump[nAssump+1] = Abc_Var2Lit( d, fHas1s ); // if there are 1s, check if 0 is SAT
            clk = Abc_Clock();
            if Abc_TtGetBit( pDivWords, 4*d+2*c+fHas1s ) {
                p.nSatCallsUnsat--;
                status = l_False;
            
            } else {
                p.nSatCalls++;
                status = sat_solver_solve( p.pSat, pAssump, pAssump + nAssump + 2, nBTLimit, 0, 0, 0 );
            }
            if status == l_Undef {
                p.nTimeOuts++;
                return -2;
            }
            if status == l_False {
                p.nSatCallsUnsat++;
                p.timeSatUnsat += Abc_Clock() - clk;
                Impls[c] = Abc_LitNot(pAssump[nAssump+1]);
                Vec_IntPush( &p.vImpls[c], Abc_LitNot(pAssump[nAssump+1]) );
                Abc_TtSetBit( pDivWords, 4*d+2*c+fHas1s );
                continue;
            }
            assert(status == l_True, "status == l_True");
            p.nSatCallsSat++;
            p.timeSatSat += Abc_Clock() - clk;
            if p.nPats[c] == 64*SFM_SIM_WORDS {
                p.nSatCallsOver++;
                continue;//return -2;//continue;
            }
            // record this status
            for  i = 0; i < p.nDivs; i++  {
                if sat_solver_var_value(p.pSat, i) {
                    Abc_TtSetBit( Sfm_DecDivPats(p, i, c), p.nPats[c] ); }
            }
            p.nPatWords[c] = 1 + (p.nPats[c] >> 6);
            Abc_TtSetBit( Masks[c], p.nPats[c]++ );
        }
        if Impls[0] == -1 || Impls[1] == -1 {
            continue; }
        if Impls[0] == Impls[1] {
            Vec_IntPop( &p.vImpls[0] );
            Vec_IntPop( &p.vImpls[1] );
            continue;
        }
        assert(Abc_Lit2Var(Impls[0]) == Abc_Lit2Var(Impls[1]), "Abc_Lit2Var(Impls[0]) == Abc_Lit2Var(Impls[1])");
        // found buffer/inverter
        Abc_TtUnit( pTruth, Abc_TtWordNum(p.pPars.nVarMax), Abc_LitIsCompl(Impls[0]) );
        pSupp[0] = Abc_Lit2Var(Impls[0]);
        if p.pPars.fVeryVerbose {
            fmt.Printf("Found variable %s%d.\n", Abc_LitIsCompl(Impls[0]) ? "!":"", pSupp[0]); }
        return 1;        
    }
    if nSuppAdd > p.pPars.nVarMax - 2 {
        if p.pPars.fVeryVerbose {
            fmt.Printf("The number of assumption is more than MFFC size.\n"); }
        return -2;
    }
    // try using all implications at once
    if p.pPars.fUseAndOr {
    for  c = 0; c < 2; c++  {
        if Vec_IntSize(&p.vImpls[!c]) < 2 {
            continue; }
        p.nSatCalls++;
        pAssump[nAssump] = Abc_Var2Lit( p.iTarget, c );
        assert(Vec_IntSize(&p.vImpls[!c]) < SFM_WIN_MAX-10, "Vec_IntSize(&p.vImpls[!c]) < SFM_WIN_MAX-10");
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(&p.vImpls[!c])) && (((iLit) = Vec_IntEntry(&p.vImpls[!c], i)), 1); i++  {
            pAssump[nAssump+1+i] = iLit;
        }
        clk = Abc_Clock();
        status = sat_solver_solve( p.pSat, pAssump, pAssump + nAssump+1+i, nBTLimit, 0, 0, 0 );
        if status == l_Undef {
            p.nTimeOuts++;
            return -2;
        }
        if status == l_False {
            var pFinal *int32;
            var nFinal int32 = sat_solver_final( p.pSat, &pFinal );
            p.nSatCallsUnsat++;
            p.timeSatUnsat += Abc_Clock() - clk;
            if nFinal + nSuppAdd > 6 {
                continue; }
            // collect only relevant literals
            for  i = d = 0; i < nFinal; i++  {
                if Vec_IntFind(&p.vImpls[!c], Abc_LitNot(pFinal[i])) >= 0 {
                    pSupp[d++] = Abc_LitNot(pFinal[i]); }
            }
            nFinal = d;
            // create AND/OR gate
            assert(nFinal <= 6, "nFinal <= 6");
            if c != 0 {
                *pTruth = ~(word)0;
                for  i = 0; i < nFinal; i++  {
                    *pTruth &= Abc_LitIsCompl(pSupp[i]) ? ~s_Truths6[i] : s_Truths6[i];
                    pSupp[i] = Abc_Lit2Var(pSupp[i]);
                }
            
            } else {
                *pTruth = 0;
                for  i = 0; i < nFinal; i++  {
                    *pTruth |= Abc_LitIsCompl(pSupp[i]) ? s_Truths6[i] : ~s_Truths6[i];
                    pSupp[i] = Abc_Lit2Var(pSupp[i]);
                }
            }
            Abc_TtStretch6( pTruth, nFinal, p.pPars.nVarMax );
            p.nNodesAndOr++;
            if p.pPars.fVeryVerbose {
                fmt.Printf("Found %d-input AND/OR gate.\n", nFinal); }
            return nFinal;
        }
        assert(status == l_True, "status == l_True");
        p.nSatCallsSat++;
        p.timeSatSat += Abc_Clock() - clk;
        if p.nPats[c] == 64*SFM_SIM_WORDS {
            p.nSatCallsOver++;
            continue;//return -2;//continue;
        }
        for  i = 0; i < p.nDivs; i++  {
            if sat_solver_var_value(p.pSat, i) {
                Abc_TtSetBit( Sfm_DecDivPats(p, i, c), p.nPats[c] ); }
        }
        p.nPatWords[c] = 1 + (p.nPats[c] >> 6);
        Abc_TtSetBit( Masks[c], p.nPats[c]++ );
    } }

    // find the best cofactoring variable
//    if ( !fCofactor || Vec_IntSize(&p->vImpls[0]) + Vec_IntSize(&p->vImpls[1]) > 2 )
    Var = Sfm_DecFindBestVar( p, Masks );
//    if ( Var == -1 )
//        Var = Sfm_DecFindBestVar2( p, Masks );

/*
    {
        int Lit0 = Vec_IntSize(&p->vImpls[0]) ? Vec_IntEntry(&p->vImpls[0], 0) : -1;
        int Lit1 = Vec_IntSize(&p->vImpls[1]) ? Vec_IntEntry(&p->vImpls[1], 0) : -1;
        if ( Lit0 == -1 && Lit1 >= 0 )
            Var = Abc_Lit2Var(Lit1);
        else if ( Lit1 == -1 && Lit0 >= 0 )
            Var = Abc_Lit2Var(Lit0);
        else if ( Lit0 >= 0 && Lit1 >= 0 )
        {
            if ( Lit0 < Lit1 )
                Var = Abc_Lit2Var(Lit0);
            else
                Var = Abc_Lit2Var(Lit1);
        }
    }
*/

    if Var == -1 && (fCofactor != 0) {
        //for ( Var = p->nDivs - 1; Var >= 0; Var-- )
        // inlining of macro Vec_IntForEachEntryReverse
        for  i = Vec_IntSize(&p.vObjInMffc) - 1; (i >= 0) && (((Var) = Vec_IntEntry(&p.vObjInMffc, i)), 1); i--  {
            if Vec_IntFind(&p.vObjDec, Var) == -1 {
                break; }
        }
//        if ( i == Vec_IntSize(&p->vObjInMffc) )
        if i == -1 {
            Var = -1; }
        fCofactor = 0;
    }

    if p.pPars.fVeryVerbose {
        Sfm_DecPrint( p, Masks );
        fmt.Printf("Best var %d\n", Var);
        fmt.Printf("\n");
    }
cofactor:
    // cofactor the problem
    if Var >= 0 {
        var uTruth [2][SFM_WORD_MAX]word;
        var MasksNext [2][SFM_SIM_WORDS]word;
        var w int32;
        var Supp [2][2*SFM_SUPP_MAX]int32;
        var nSupp [2]int32 ;
        Vec_IntPush( &p.vObjDec, Var );
        for  i = 0; i < 2; i++  {
            for  c = 0; c < 2; c++  {
                Abc_TtAndSharp( MasksNext[c], Masks[c], Sfm_DecDivPats(p, Var, c), p.nPatWords[c], !i );
                for  w = p.nPatWords[c]; w < SFM_SIM_WORDS; w++  {
                    MasksNext[c][w] = 0;
                }
            }
            pAssump[nAssump] = Abc_Var2Lit( Var, !i );
            memcpy( p.pDivWords[nAssump+1], p.pDivWords[nAssump], sizeof(word) * p.nDivWords );
            nSupp[i] = Sfm_DecPeformDec_rec( p, uTruth[i], Supp[i], pAssump, nAssump+1, MasksNext, fCofactor, (i ? nSupp[0] : 0) + nSuppAdd + 1 );
            if nSupp[i] == -2 {
                return -2; }
        }
        // combine solutions
        return Sfm_DecCombineDec( p, uTruth[0], uTruth[1], Supp[0], Supp[1], nSupp[0], nSupp[1], pTruth, pSupp, Var );
    }
    return -2;
}
func Sfm_DecPeformDec2(p *Sfm_Dec_t, pObj *Abc_Obj_t) int32 {
    var uTruth [SFM_DEC_MAX][SFM_WORD_MAX]word;
    var Masks [2][SFM_SIM_WORDS]word;
    var pSupp [SFM_DEC_MAX][2*SFM_SUPP_MAX]int32;
    var nSupp [SFM_DEC_MAX]int32;
    var pAssump [SFM_WIN_MAX]int32;
    var fVeryVerbose int32  = p.pPars.fPrintDecs || p.pPars.fVeryVerbose;
    var nDecs int32  = Abc_MaxInt(p.pPars.nDecMax, 1);
    //int fNeedInv, AreaGainInv = Sfm_DecComputeFlipInvGain(p, pObj, &fNeedInv);
    var i int32;
    var RetValue int32;
    var Prev int32;
    var iBest int32 = -1;
    var AreaThis int32;
    var AreaNew int32;//, AreaNewInv;
    var GainThis int32;
    var GainBest int32 = -1;
    var iLibObj int32;
    var iLibObjBest int32 = -1; 
    assert(p.pPars.fArea == 1, "p.pPars.fArea == 1");
//printf( "AreaGainInv = %8.2f  ", Scl_Int2Flt(AreaGainInv) );
    //Sfm_DecPrint( p, NULL );
    if fVeryVerbose != 0 {
        fmt.Printf("\nNode %4d : MFFC %2d\n", p.iTarget, p.nMffc); }
    assert(p.pPars.nDecMax <= SFM_DEC_MAX, "p.pPars.nDecMax <= SFM_DEC_MAX");
    Sfm_ObjSetupSimInfo( pObj );
    Vec_IntClear( &p.vObjDec );
    for  i = 0; i < nDecs; i++  {
        // reduce the variable array
        if Vec_IntSize(&p.vObjDec) > Prev {
            Vec_IntShrink( &p.vObjDec, Prev ); }
        Prev = Vec_IntSize(&p.vObjDec) + 1;
        // perform decomposition 
        Abc_TtMask( Masks[0], SFM_SIM_WORDS, p.nPats[0] );
        Abc_TtMask( Masks[1], SFM_SIM_WORDS, p.nPats[1] );        
        nSupp[i] = Sfm_DecPeformDec_rec( p, uTruth[i], pSupp[i], pAssump, 0, Masks, 1, 0 );
        if nSupp[i] == -2 {
            if fVeryVerbose != 0 {
                fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1]); }
            continue;
        }
        if fVeryVerbose != 0 {
            fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  Supp = %d  ", i, p.nPats[0], p.nPats[1], nSupp[i]); }
        if fVeryVerbose != 0 {
            Dau_DsdPrintFromTruth( uTruth[i], nSupp[i] ); }
        if nSupp[i] < 2 {
            p.nSuppVars = nSupp[i];
            Abc_TtCopy( p.Copy, uTruth[i], SFM_WORD_MAX, 0 ); 
            RetValue = Sfm_LibImplementSimple( p.pLib, uTruth[i], pSupp[i], nSupp[i], &p.vObjGates, &p.vObjFanins );
            assert(nSupp[i] <= p.pPars.nVarMax, "nSupp[i] <= p.pPars.nVarMax");
            p.nLuckySizes[nSupp[i]]++;
            assert(RetValue <= 2, "RetValue <= 2");
            p.nLuckyGates[RetValue]++;
            //printf( "\n" );
            return RetValue;
        }

        p.nSuppVars = nSupp[i];
        Abc_TtCopy( p.Copy, uTruth[i], SFM_WORD_MAX, 0 ); 
        AreaNew = Sfm_LibFindAreaMatch( p.pLib, uTruth[i], nSupp[i], &iLibObj );
/*
        uTruth[i][0] = ~uTruth[i][0];
        AreaNewInv = Sfm_LibFindAreaMatch( p->pLib, uTruth[i], nSupp[i], NULL );
        uTruth[i][0] = ~uTruth[i][0];

        if ( AreaNew > 0 && AreaNewInv > 0 && AreaNew - AreaNewInv + AreaGainInv > 0 )
            printf( "AreaNew = %8.2f   AreaNewInv = %8.2f   Gain = %8.2f   Total = %8.2f\n", 
                Scl_Int2Flt(AreaNew), Scl_Int2Flt(AreaNewInv), Scl_Int2Flt(AreaNew - AreaNewInv), Scl_Int2Flt(AreaNew - AreaNewInv + AreaGainInv) );
        else
            printf( "\n" );
*/
        if AreaNew == -1 {
            continue; }


        // compute area savings
        Sfm_DecPrepareVec( &p.vObjMap, pSupp[i], nSupp[i], &p.vGateCut );
        AreaThis = Sfm_DecMffcAreaReal(pObj, &p.vGateCut, nil);
        assert(p.AreaMffc <= AreaThis, "p.AreaMffc <= AreaThis");
        if p.pPars.fZeroCost ? (AreaNew > AreaThis) : (AreaNew >= AreaThis) {
            continue; }
        // find the best gain
        GainThis = AreaThis - AreaNew;
        assert(GainThis >= 0, "GainThis >= 0");
        if GainBest < GainThis {
            GainBest    = GainThis;
            iLibObjBest = iLibObj;
            iBest       = i;
        }
    }
    Sfm_ObjSetdownSimInfo( pObj );
    if iBest == -1 {
        if fVeryVerbose != 0 {
            fmt.Printf("Best  : NO DEC.\n"); }
        p.nNoDecs++;
        //printf( "\n" );
        return -2;
    }
    if fVeryVerbose != 0 {
        fmt.Printf("Best %d: %d  ", iBest, nSupp[iBest]); }
    if fVeryVerbose != 0 {
      Dau_DsdPrintFromTruth( uTruth[iBest], nSupp[iBest] ); }
    // implement
    assert(iLibObjBest >= 0, "iLibObjBest >= 0");
    RetValue = Sfm_LibImplementGatesArea( p.pLib, pSupp[iBest], nSupp[iBest], iLibObjBest, &p.vObjGates, &p.vObjFanins );
    assert(nSupp[iBest] <= p.pPars.nVarMax, "nSupp[iBest] <= p.pPars.nVarMax");
    p.nLuckySizes[nSupp[iBest]]++;
    assert(RetValue <= 2, "RetValue <= 2");
    p.nLuckyGates[RetValue]++;
    return 1;
}
func Sfm_DecPeformDec3(p *Sfm_Dec_t, pObj *Abc_Obj_t) int32 {
    var uTruth [SFM_DEC_MAX][SFM_WORD_MAX]word;
    var Masks [2][SFM_SIM_WORDS]word;
    var pSupp [SFM_DEC_MAX][2*SFM_SUPP_MAX]int32;
    var nSupp [SFM_DEC_MAX]int32;
    var pAssump [SFM_WIN_MAX]int32;
    var fVeryVerbose int32  = p.pPars.fPrintDecs || p.pPars.fVeryVerbose;
    var nDecs int32  = Abc_MaxInt(p.pPars.nDecMax, 1);
    var i int32;
    var k int32;
    var DelayOrig int32;
    var DelayMin int32;
    var GainMax int32;
    var AreaMffc int32;
    var nMatches int32;
    var iBest int32 = -1;
    var RetValue int32;
    var Prev int32; 
    var pGate1Best *Mio_Gate_t = nil;
    var pGate2Best *Mio_Gate_t = nil; 
    var pFans1Best string = nil;
    var pFans2Best string = nil;
    assert(p.pPars.fArea == 0, "p.pPars.fArea == 0");
    p.DelayMin = 0;
    //Sfm_DecPrint( p, NULL );
    if fVeryVerbose != 0 {
        fmt.Printf("\nNode %4d : MFFC %2d\n", p.iTarget, p.nMffc); }
    // set limit on search for decompositions in delay-model
    assert(p.pPars.nDecMax <= SFM_DEC_MAX, "p.pPars.nDecMax <= SFM_DEC_MAX");
    Sfm_ObjSetupSimInfo( pObj );
    Vec_IntClear( &p.vObjDec );
    for  i = 0; i < nDecs; i++  {
        GainMax = 0;
        DelayMin = DelayOrig = Sfm_ManReadObjDelay( p, Abc_ObjId(pObj) );
        // reduce the variable array
        if Vec_IntSize(&p.vObjDec) > Prev {
            Vec_IntShrink( &p.vObjDec, Prev ); }
        Prev = Vec_IntSize(&p.vObjDec) + 1;
        // perform decomposition 
        Abc_TtMask( Masks[0], SFM_SIM_WORDS, p.nPats[0] );
        Abc_TtMask( Masks[1], SFM_SIM_WORDS, p.nPats[1] );        
        nSupp[i] = Sfm_DecPeformDec_rec( p, uTruth[i], pSupp[i], pAssump, 0, Masks, 1, 0 );
        if nSupp[i] == -2 {
            if fVeryVerbose != 0 {
                fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1]); }
            continue;
        }
        if fVeryVerbose != 0 {
            fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  Supp = %d  ", i, p.nPats[0], p.nPats[1], nSupp[i]); }
        if fVeryVerbose != 0 {
            Dau_DsdPrintFromTruth( uTruth[i], nSupp[i] ); }
        if p.pTim && nSupp[i] == 1 && uTruth[i][0] == ABC_CONST(0x5555555555555555) && DelayMin <= p.DelayInv + Sfm_ManReadObjDelay(p, Vec_IntEntry(&p.vObjMap, pSupp[i][0])) {
            if fVeryVerbose != 0 {
                fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1]); }
            continue;
        }
        if p.pMit && nSupp[i] == 1 && uTruth[i][0] == ABC_CONST(0x5555555555555555) {
            if fVeryVerbose != 0 {
                fmt.Printf("Dec  %d: Pat0 = %2d  Pat1 = %2d  NO DEC.\n", i, p.nPats[0], p.nPats[1]); }
            continue;
        }
        if nSupp[i] < 2 {
            p.nSuppVars = nSupp[i];
            Abc_TtCopy( p.Copy, uTruth[i], SFM_WORD_MAX, 0 ); 
            RetValue = Sfm_LibImplementSimple( p.pLib, uTruth[i], pSupp[i], nSupp[i], &p.vObjGates, &p.vObjFanins );
            assert(nSupp[i] <= p.pPars.nVarMax, "nSupp[i] <= p.pPars.nVarMax");
            p.nLuckySizes[nSupp[i]]++;
            assert(RetValue <= 2, "RetValue <= 2");
            p.nLuckyGates[RetValue]++;
            return RetValue;
        }

        // get MFFC
        Sfm_DecPrepareVec( &p.vObjMap, pSupp[i], nSupp[i], &p.vGateCut ); // returns cut in p->vGateCut
        AreaMffc = Sfm_DecMffcAreaReal(pObj, &p.vGateCut, &p.vGateMffc ); // returns MFFC in p->vGateMffc

        // try the delay
        p.nSuppVars = nSupp[i];
        Abc_TtCopy( p.Copy, uTruth[i], SFM_WORD_MAX, 0 ); 
        nMatches = Sfm_LibFindDelayMatches( p.pLib, uTruth[i], pSupp[i], nSupp[i], &p.vMatchGates, &p.vMatchFans );
        for  k = 0; k < nMatches; k++  {
            var clk abctime  = Abc_Clock();
            var pGate1 *Mio_Gate_t  = (Mio_Gate_t *)Vec_PtrEntry( &p.vMatchGates, 2*k+0 );
            var pGate2 *Mio_Gate_t  = (Mio_Gate_t *)Vec_PtrEntry( &p.vMatchGates, 2*k+1 );
            var AreaNew int32  = Scl_Flt2Int( Mio_GateReadArea(pGate1) + (pGate2 ? Mio_GateReadArea(pGate2) : 0.0) );
            var pFans1 string  = (char *)Vec_PtrEntry( &p.vMatchFans, 2*k+0 );
            var pFans2 string  = (char *)Vec_PtrEntry( &p.vMatchFans, 2*k+1 );
            var vFanins Vec_Int_t  = Vec_Int_t{ nSupp[i], nSupp[i], pSupp[i] };
            // skip identical gate
            //if ( pGate2 == NULL && pGate1 == (Mio_Gate_t *)pObj->pData )
            //    continue;
            if p.pMit {
                var Gain int32  = Sfm_MitEvalRemapping( p.pMit, &p.vGateMffc, pObj, &vFanins, &p.vObjMap, pGate1, pFans1, pGate2, pFans2 );
                if p.pPars.DelAreaRatio && AreaNew > AreaMffc && (Gain / (AreaNew - AreaMffc)) < p.pPars.DelAreaRatio {
                    continue; }
                if GainMax < Gain {
                    GainMax    = Gain;
                    pGate1Best = pGate1;
                    pGate2Best = pGate2;
                    pFans1Best = pFans1;
                    pFans2Best = pFans2;
                    iBest      = i;
                }
            
            } else {
                var Delay int32  = Sfm_TimEvalRemapping( p.pTim, &vFanins, &p.vObjMap, pGate1, pFans1, pGate2, pFans2 );
                if p.pPars.DelAreaRatio && AreaNew > AreaMffc && (Delay / (AreaNew - AreaMffc)) < p.pPars.DelAreaRatio {
                    continue; }
                if DelayMin > Delay {
                    DelayMin   = Delay;
                    pGate1Best = pGate1;
                    pGate2Best = pGate2;
                    pFans1Best = pFans1;
                    pFans2Best = pFans2;
                    iBest      = i;
                }
            }
            p.timeEval += Abc_Clock() - clk;
        }
    }
//printf( "Gain max = %d.\n", GainMax );
    Sfm_ObjSetdownSimInfo( pObj );
    if iBest == -1 {
        if fVeryVerbose != 0 {
            fmt.Printf("Best  : NO DEC.\n"); }
        p.nNoDecs++;
        return -2;
    }
    if fVeryVerbose != 0 {
        fmt.Printf("Best %d: %d  ", iBest, nSupp[iBest]); }
//    if ( fVeryVerbose )
//        Dau_DsdPrintFromTruth( uTruth[iBest], nSupp[iBest] );
    RetValue = Sfm_LibImplementGatesDelay( p.pLib, pSupp[iBest], pGate1Best, pGate2Best, pFans1Best, pFans2Best, &p.vObjGates, &p.vObjFanins );
    assert(nSupp[iBest] <= p.pPars.nVarMax, "nSupp[iBest] <= p.pPars.nVarMax");
    p.nLuckySizes[nSupp[iBest]]++;
    assert(RetValue <= 2, "RetValue <= 2");
    p.nLuckyGates[RetValue]++;
    p.DelayMin = DelayMin;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Incremental level update.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkUpdateIncLevel_rec(pObj *Abc_Obj_t) {
    var pFanout *Abc_Obj_t;
    var i int32;
    var LevelNew int32 = Abc_ObjLevelNew(pObj);
    if LevelNew == Abc_ObjLevel(pObj) && Abc_ObjIsNode(pObj) && Abc_ObjFaninNum(pObj) > 0 {
        return; }
    pObj.Level = LevelNew;
    if !Abc_ObjIsCo(pObj) {
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
            Abc_NtkUpdateIncLevel_rec( pFanout );
        } }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkDfsCheck_rec(pObj *Abc_Obj_t, pPivot *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t; var i int32;
    if pObj == pPivot {
        return 0; }
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return 1; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCi(pObj) {
        return 1; }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        if !(Abc_NtkDfsCheck_rec(pFanin, pPivot) != 0) {
            return 0; }
    }
    return 1;
}
func Abc_NtkDfsReverseOne_rec(pObj *Abc_Obj_t, vTfo *Vec_Int_t, nLevelMax int32, nFanoutMax int32) {
    var pFanout *Abc_Obj_t; var i int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCo(pObj) || Abc_ObjLevel(pObj) > nLevelMax {
        return; }
    assert(Abc_ObjIsNode( pObj ), "Abc_ObjIsNode( pObj )");
    if Abc_ObjFanoutNum(pObj) <= nFanoutMax {
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
            if Abc_ObjIsCo(pFanout) || Abc_ObjLevel(pFanout) > nLevelMax {
                break; }
        }
        if i == Abc_ObjFanoutNum(pObj) {
            // inlining of macro Abc_ObjForEachFanout
            for  i = 0; (i < Abc_ObjFanoutNum(pObj)) && (((pFanout) = Abc_ObjFanout(pObj, i)), 1); i++  {
                Abc_NtkDfsReverseOne_rec( pFanout, vTfo, nLevelMax, nFanoutMax );
            } }
    }
    Vec_IntPush( vTfo, Abc_ObjId(pObj) );
    pObj.iTemp = 0;
}
func Abc_NtkDfsOne_rec(pObj *Abc_Obj_t, vTfi *Vec_Int_t, nLevelMin int32, CiLabel int32) int32 {
    var pFanin *Abc_Obj_t; var i int32;
    if Abc_NodeIsTravIdCurrent( pObj ) {
        return pObj.iTemp; }
    Abc_NodeSetTravIdCurrent( pObj );
    if Abc_ObjIsCi(pObj) || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0) {
        Vec_IntPush( vTfi, Abc_ObjId(pObj) );
        return (pObj.iTemp = CiLabel);
    }
    assert(Abc_ObjIsNode(pObj), "Abc_ObjIsNode(pObj)");
    if Abc_ObjFaninNum(pObj) {
        pObj.iTemp = 0;
    } else {
        pObj.iTemp = CiLabel;
    }
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        pObj.iTemp |= Abc_NtkDfsOne_rec( pFanin, vTfi, nLevelMin, CiLabel );
    }
    Vec_IntPush( vTfi, Abc_ObjId(pObj) );
    Sfm_ObjSimulateNode( pObj );
    return pObj.iTemp;
}
func Sfm_DecAddNode(pObj *Abc_Obj_t, vMap *Vec_Int_t, vGates *Vec_Int_t, fSkip int32, fVeryVerbose int32) {
    if fVeryVerbose != 0 {
        fmt.Printf("%d:%d(%d) ", Vec_IntSize(vMap), Abc_ObjId(pObj), pObj.iTemp); }
    if fVeryVerbose != 0 {
        Abc_ObjPrint( os.Stdout, pObj ); }
    Vec_IntPush( vMap, Abc_ObjId(pObj) );
    Vec_IntPush( vGates, fSkip ? -1 : Mio_GateReadValue((Mio_Gate_t *)pObj.pData) );
}
func Sfm_DecNodeIsMffc(p *Abc_Obj_t, nLevelMin int32) int32 {
    return Abc_ObjIsNode(p) && Abc_ObjFanoutNum(p) == 1 && Abc_NodeIsTravIdCurrent(p) && (Abc_ObjLevel(p) >= nLevelMin || Abc_ObjFaninNum(p) == 0);
}
func Sfm_DecNodeIsMffcInput(p *Abc_Obj_t, nLevelMin int32, pTim *Sfm_Tim_t, pPivot *Abc_Obj_t) int32 {
    return Abc_NodeIsTravIdCurrent(p) && Sfm_TimNodeIsNonCritical(pTim, pPivot, p);
}
func Sfm_DecNodeIsMffcInput2(p *Abc_Obj_t, nLevelMin int32, pMit *Sfm_Mit_t, pPivot *Abc_Obj_t) int32 {
    return Abc_NodeIsTravIdCurrent(p) && Sfm_MitNodeIsNonCritical(pMit, pPivot, p);
}
func Sfm_DecMarkMffc(pPivot *Abc_Obj_t, nLevelMin int32, nMffcMax int32, fVeryVerbose int32, vMffc *Vec_Int_t, vInMffc *Vec_Int_t, pTim *Sfm_Tim_t, pMit *Sfm_Mit_t) {
    var pFanin, pFanin2, pFanin3, pObj *Abc_Obj_t; var i, k, n int32;
    assert(nMffcMax > 0, "nMffcMax > 0");
    Vec_IntFill( vMffc, 1, Abc_ObjId(pPivot) );
    if pMit != nil {
        pPivot.iTemp |= SFM_MASK_MFFC;
        pPivot.iTemp |= SFM_MASK_PIVOT;
        // collect MFFC inputs (these are low-delay nodes close to the pivot)
        Vec_IntClear(vInMffc);
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            if Sfm_DecNodeIsMffcInput2(pFanin, nLevelMin, pMit, pPivot) != 0 {
                Vec_IntPushUnique( vInMffc, Abc_ObjId(pFanin) ); }
        }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k)), 1); k++  {
                if Sfm_DecNodeIsMffcInput2(pFanin2, nLevelMin, pMit, pPivot) != 0 {
                    Vec_IntPushUnique( vInMffc, Abc_ObjId(pFanin2) ); }
            }
        }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k)), 1); k++  {
                // inlining of macro Abc_ObjForEachFanin
                for  n = 0; (n < Abc_ObjFaninNum(pFanin2)) && (((pFanin3) = Abc_ObjFanin(pFanin2, n)), 1); n++  {
                    if Sfm_DecNodeIsMffcInput2(pFanin3, nLevelMin, pMit, pPivot) != 0 {
                        Vec_IntPushUnique( vInMffc, Abc_ObjId(pFanin3) ); }
                }
            }
        }
    
    } else if pTim != nil {
        pPivot.iTemp |= SFM_MASK_MFFC;
        pPivot.iTemp |= SFM_MASK_PIVOT;
        // collect MFFC inputs (these are low-delay nodes close to the pivot)
        Vec_IntClear(vInMffc);
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            if Sfm_DecNodeIsMffcInput(pFanin, nLevelMin, pTim, pPivot) != 0 {
                Vec_IntPushUnique( vInMffc, Abc_ObjId(pFanin) ); }
        }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k)), 1); k++  {
                if Sfm_DecNodeIsMffcInput(pFanin2, nLevelMin, pTim, pPivot) != 0 {
                    Vec_IntPushUnique( vInMffc, Abc_ObjId(pFanin2) ); }
            }
        }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k)), 1); k++  {
                // inlining of macro Abc_ObjForEachFanin
                for  n = 0; (n < Abc_ObjFaninNum(pFanin2)) && (((pFanin3) = Abc_ObjFanin(pFanin2, n)), 1); n++  {
                    if Sfm_DecNodeIsMffcInput(pFanin3, nLevelMin, pTim, pPivot) != 0 {
                        Vec_IntPushUnique( vInMffc, Abc_ObjId(pFanin3) ); }
                }
            }
        }

/*
        printf( "Node %d: (%.2f)  ", pPivot->Id, Scl_Int2Flt(Sfm_ManReadObjDelay(p, Abc_ObjId(pPivot)))  );
        Abc_ObjForEachFanin( pPivot, pFanin, i )
            printf( "%d: %.2f  ", Abc_ObjLevel(pFanin), Scl_Int2Flt(Sfm_ManReadObjDelay(p, Abc_ObjId(pFanin))) );
        printf( "\n" );

        printf( "Node %d: ", pPivot->Id );
        Abc_NtkForEachObjVec( vInMffc, pPivot->pNtk, pObj, i )
            printf( "%d: %.2f  ", Abc_ObjLevel(pObj), Scl_Int2Flt(Sfm_ManReadObjDelay(p, Abc_ObjId(pObj))) );
        printf( "\n" );
*/
    
    } else {

        // collect MFFC
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            if (Sfm_DecNodeIsMffc(pFanin, nLevelMin) != 0) && Vec_IntSize(vMffc) < nMffcMax {
                Vec_IntPushUnique( vMffc, Abc_ObjId(pFanin) ); }
        }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            if (Sfm_DecNodeIsMffc(pFanin, nLevelMin) != 0) && Vec_IntSize(vMffc) < nMffcMax {
                // inlining of macro Abc_ObjForEachFanin
                for  k = 0; (k < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k)), 1); k++  {
                    if (Sfm_DecNodeIsMffc(pFanin2, nLevelMin) != 0) && Vec_IntSize(vMffc) < nMffcMax {
                        Vec_IntPushUnique( vMffc, Abc_ObjId(pFanin2) ); }
                } }
        }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pPivot)) && (((pFanin) = Abc_ObjFanin(pPivot, i)), 1); i++  {
            if (Sfm_DecNodeIsMffc(pFanin, nLevelMin) != 0) && Vec_IntSize(vMffc) < nMffcMax {
                // inlining of macro Abc_ObjForEachFanin
                for  k = 0; (k < Abc_ObjFaninNum(pFanin)) && (((pFanin2) = Abc_ObjFanin(pFanin, k)), 1); k++  {
                    if (Sfm_DecNodeIsMffc(pFanin2, nLevelMin) != 0) && Vec_IntSize(vMffc) < nMffcMax {
                        // inlining of macro Abc_ObjForEachFanin
                        for  n = 0; (n < Abc_ObjFaninNum(pFanin2)) && (((pFanin3) = Abc_ObjFanin(pFanin2, n)), 1); n++  {
                            if (Sfm_DecNodeIsMffc(pFanin3, nLevelMin) != 0) && Vec_IntSize(vMffc) < nMffcMax {
                                Vec_IntPushUnique( vMffc, Abc_ObjId(pFanin3) ); }
                        } }
                } }
        }
        // mark MFFC
        assert(Vec_IntSize(vMffc) <= nMffcMax, "Vec_IntSize(vMffc) <= nMffcMax");
        // inlining of macro Abc_NtkForEachObjVec
        for  i = 0; i < Vec_IntSize(vMffc) && (((pObj) = Abc_NtkObj(pPivot.pNtk, Vec_IntEntry(vMffc,i))), 1); i++  {         if (pObj) == nil {} else {
            pObj.iTemp |= SFM_MASK_MFFC; }
        }
        pPivot.iTemp |= SFM_MASK_PIVOT;
        // collect MFFC inputs
        Vec_IntClear(vInMffc);
        // inlining of macro Abc_NtkForEachObjVec
        for  i = 0; i < Vec_IntSize(vMffc) && (((pObj) = Abc_NtkObj(pPivot.pNtk, Vec_IntEntry(vMffc,i))), 1); i++  {         if (pObj) == nil {} else {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                if Abc_NodeIsTravIdCurrent(pFanin) && pFanin.iTemp == SFM_MASK_PI {
                    Vec_IntPushUnique( vInMffc, Abc_ObjId(pFanin) ); }
            } }
        }

//        printf( "Node %d: ", pPivot->Id );
//        Abc_ObjForEachFanin( pPivot, pFanin, i )
//            printf( "%d ", Abc_ObjFanoutNum(pFanin) );
//        printf( "\n" );

//        Abc_NtkForEachObjVec( vInMffc, pPivot->pNtk, pObj, i )
//            printf( "%d ", Abc_ObjFanoutNum(pObj) );
//        printf( "\n" );

    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sfm_DecExtract(pNtk *Abc_Ntk_t, pPars *Sfm_Par_t, pPivot *Abc_Obj_t, vRoots *Vec_Int_t, vGates *Vec_Int_t, vFanins *Vec_Wec_t, vMap *Vec_Int_t, vTfi *Vec_Int_t, vTfo *Vec_Int_t, vMffc *Vec_Int_t, vInMffc *Vec_Int_t, pTim *Sfm_Tim_t, pMit *Sfm_Mit_t) int32 {
    var fVeryVerbose int32;//pPars->fVeryVerbose;
    var vLevel *Vec_Int_t;
    var pObj, pFanin *Abc_Obj_t;
    var nLevelMax int32  = pPivot.Level + pPars.nTfoLevMax;
    var nLevelMin int32  = pPivot.Level - pPars.nTfiLevMax;
    var i int32;
    var k int32;
    var nTfiSize int32;
    var nDivs int32 = -1;
    assert(Abc_ObjIsNode(pPivot), "Abc_ObjIsNode(pPivot)");
if fVeryVerbose != 0 {
fmt.Printf("\n\nTarget %d\n", Abc_ObjId(pPivot)); }
    // collect TFO nodes
    Vec_IntClear( vTfo );
    Abc_NtkIncrementTravId( pNtk );
    Abc_NtkDfsReverseOne_rec( pPivot, vTfo, nLevelMax, pPars.nFanoutMax );
    // count internal fanouts
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vTfo) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfo,i))), 1); i++  {         if (pObj) == nil {} else {
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
            pFanin.iTemp++;
        } }
    }
    // compute roots
    Vec_IntClear( vRoots );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vTfo) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfo,i))), 1); i++  {         if (pObj) == nil {} else
        if pObj.iTemp != Abc_ObjFanoutNum(pObj) {
            Vec_IntPush( vRoots, Abc_ObjId(pObj) ); }
    }
    assert(Vec_IntSize(vRoots) > 0, "Vec_IntSize(vRoots) > 0");
    // collect TFI and mark nodes
    Vec_IntClear( vTfi );
    Abc_NtkIncrementTravId( pNtk );
    Abc_NtkDfsOne_rec( pPivot, vTfi, nLevelMin, SFM_MASK_PI );
    nTfiSize = Vec_IntSize(vTfi);
    Sfm_ObjFlipNode( pPivot );
    // additinally mark MFFC
    Sfm_DecMarkMffc( pPivot, nLevelMin, pPars.nMffcMax, fVeryVerbose, vMffc, vInMffc, pTim, pMit );
    assert(Vec_IntSize(vMffc) <= pPars.nMffcMax, "Vec_IntSize(vMffc) <= pPars.nMffcMax");
if fVeryVerbose != 0 {
fmt.Printf("Mffc size = %d. Mffc area = %.2f. InMffc size = %d.\n", Vec_IntSize(vMffc), Scl_Int2Flt(Sfm_DecMffcArea(pNtk, vMffc)), Vec_IntSize(vInMffc)); }
    // collect TFI(TFO)
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vRoots,i))), 1); i++  {         if (pObj) == nil {} else {
        Abc_NtkDfsOne_rec( pObj, vTfi, nLevelMin, SFM_MASK_INPUT ); }
    }
    // mark input-only nodes pointed to by mixed nodes
    // inlining of macro Abc_NtkForEachObjVecStart
    for  i = nTfiSize; i < Vec_IntSize(vTfi) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfi,i))), 1); i++  {         if (pObj) == nil {} else
        if pObj.iTemp != SFM_MASK_INPUT {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                if pFanin.iTemp == SFM_MASK_INPUT {
                    pFanin.iTemp = SFM_MASK_FANIN; }
            } }
    }
    // collect nodes supported only on TFI fanins and not MFFC
if fVeryVerbose != 0 {
fmt.Printf("\nDivs:\n"); }
    Vec_IntClear( vMap );
    Vec_IntClear( vGates );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vTfi) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfi,i))), 1); i++  {         if (pObj) == nil {} else
        if pObj.iTemp == SFM_MASK_PI {
            Sfm_DecAddNode( pObj, vMap, vGates, Abc_ObjIsCi(pObj) || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0), fVeryVerbose ); }
    }
    nDivs = Vec_IntSize(vMap);
    // add other nodes that are not in TFO and not in MFFC
if fVeryVerbose != 0 {
fmt.Printf("\nSides:\n"); }
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vTfi) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfi,i))), 1); i++  {         if (pObj) == nil {} else
        if pObj.iTemp == (SFM_MASK_PI | SFM_MASK_INPUT) || pObj.iTemp == SFM_MASK_FANIN {
            Sfm_DecAddNode( pObj, vMap, vGates, pObj.iTemp == SFM_MASK_FANIN, fVeryVerbose ); }
    }
    // reorder nodes acording to delay
    if pMit != nil {
        var nDivsNew int32;
        var nOldSize int32 = Vec_IntSize(vMap);
        Vec_IntClear( vTfo );
        Vec_IntAppend( vTfo, vMap );
        nDivsNew = Sfm_MitSortArrayByArrival( pMit, vTfo, Abc_ObjId(pPivot) );
        // collect again
        Vec_IntClear( vMap );
        Vec_IntClear( vGates );
        // inlining of macro Abc_NtkForEachObjVec
        for  i = 0; i < Vec_IntSize(vTfo) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfo,i))), 1); i++  {         if (pObj) == nil {} else {
            Sfm_DecAddNode( pObj, vMap, vGates, Abc_ObjIsCi(pObj) || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0) || pObj.iTemp == SFM_MASK_FANIN, 0 ); }
        }
        assert(nOldSize == Vec_IntSize(vMap), "nOldSize == Vec_IntSize(vMap)");
        // update divisor count
        nDivs = nDivsNew;
    
    } else if pTim != nil {
        var nDivsNew int32;
        var nOldSize int32 = Vec_IntSize(vMap);
        Vec_IntClear( vTfo );
        Vec_IntAppend( vTfo, vMap );
        nDivsNew = Sfm_TimSortArrayByArrival( pTim, vTfo, Abc_ObjId(pPivot) );
        // collect again
        Vec_IntClear( vMap );
        Vec_IntClear( vGates );
        // inlining of macro Abc_NtkForEachObjVec
        for  i = 0; i < Vec_IntSize(vTfo) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfo,i))), 1); i++  {         if (pObj) == nil {} else {
            Sfm_DecAddNode( pObj, vMap, vGates, Abc_ObjIsCi(pObj) || (Abc_ObjLevel(pObj) < nLevelMin && Abc_ObjFaninNum(pObj) > 0) || pObj.iTemp == SFM_MASK_FANIN, 0 ); }
        }
        assert(nOldSize == Vec_IntSize(vMap), "nOldSize == Vec_IntSize(vMap)");
        // update divisor count
        nDivs = nDivsNew;
    }
    // add the TFO nodes
if fVeryVerbose != 0 {
fmt.Printf("\nTFO:\n"); }
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vTfi) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vTfi,i))), 1); i++  {         if (pObj) == nil {} else
        if pObj.iTemp >= SFM_MASK_MFFC {
            Sfm_DecAddNode( pObj, vMap, vGates, 0, fVeryVerbose ); }
    }
    assert(Vec_IntSize(vMap) == Vec_IntSize(vGates), "Vec_IntSize(vMap) == Vec_IntSize(vGates)");
if fVeryVerbose != 0 {
fmt.Printf("\n"); }
    // create node IDs
    Vec_WecClear( vFanins );
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vMap) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vMap,i))), 1); i++  {         if (pObj) == nil {} else {
        pObj.iTemp = i;
        vLevel = Vec_WecPushLevel( vFanins );
        if Vec_IntEntry(vGates, i) >= 0 {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, k)), 1); k++  {
                Vec_IntPush( vLevel, pFanin.iTemp );
            } }
    }
    }
    // compute care set
    Sfm_DecMan(pPivot).uCareSet = Sfm_ObjFindCareSet(pPivot.pNtk, vRoots);

    //printf( "care = %5d : ", Abc_ObjId(pPivot) );
    //Extra_PrintBinary( stdout, (unsigned *)&Sfm_DecMan(pPivot)->uCareSet, 64 );
    //printf( "\n" );

    // remap roots
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vRoots) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vRoots,i))), 1); i++  {         if (pObj) == nil {} else {
        Vec_IntWriteEntry( vRoots, i, pObj.iTemp ); }
    }
    // remap inputs to MFFC
    // inlining of macro Abc_NtkForEachObjVec
    for  i = 0; i < Vec_IntSize(vInMffc) && (((pObj) = Abc_NtkObj(pNtk, Vec_IntEntry(vInMffc,i))), 1); i++  {         if (pObj) == nil {} else {
        Vec_IntWriteEntry( vInMffc, i, pObj.iTemp ); }
    }

/*
    // check
    Abc_NtkForEachObjVec( vMap, pNtk, pObj, i )
    {
        if ( i == nDivs )
            break;
        Abc_NtkIncrementTravId( pNtk );
        assert( Abc_NtkDfsCheck_rec(pObj, pPivot) );
    }
*/  
    return nDivs;
}
func Sfm_DecInsert(pNtk *Abc_Ntk_t, pPivot *Abc_Obj_t, Limit int32, vGates *Vec_Int_t, vFanins *Vec_Wec_t, vMap *Vec_Int_t, vGateHandles *Vec_Ptr_t, GateBuf int32, GateInv int32, vFuncs *Vec_Wrd_t, vNewNodes *Vec_Int_t, pMit *Sfm_Mit_t) *Abc_Obj_t {
    var pObjNew *Abc_Obj_t  = nil; 
    var vLevel *Vec_Int_t;
    var i, k, iObj, Gate int32;
    if vNewNodes != nil {
        Vec_IntClear( vNewNodes ); }
    // assuming that new gates are appended at the end
    assert(Limit < Vec_IntSize(vGates), "Limit < Vec_IntSize(vGates)");
    assert(Limit == Vec_IntSize(vMap), "Limit == Vec_IntSize(vMap)");
    if Limit + 1 == Vec_IntSize(vGates) {
        Gate = Vec_IntEntryLast(vGates);
        if Gate == GateBuf {
            iObj = Vec_WecEntryEntry( vFanins, Limit, 0 );
            pObjNew = Abc_NtkObj( pNtk, Vec_IntEntry(vMap, iObj) );
            // transfer load
            if pMit != nil {
                Sfm_MitTransferLoad( pMit, pObjNew, pPivot ); }
            // replace logic cone
            Abc_ObjReplace( pPivot, pObjNew );
            // update level
            pObjNew.Level = 0;
            Abc_NtkUpdateIncLevel_rec( pObjNew );
            if vNewNodes != nil {
                Vec_IntPush( vNewNodes, Abc_ObjId(pObjNew) ); }
            return pObjNew;
        
        } else if vNewNodes == nil && Gate == GateInv {
            // check if fanouts can be updated
            var pFanout *Abc_Obj_t;
            // inlining of macro Abc_ObjForEachFanout
            for  i = 0; (i < Abc_ObjFanoutNum(pPivot)) && (((pFanout) = Abc_ObjFanout(pPivot, i)), 1); i++  {
                if !Abc_ObjIsNode(pFanout) || Sfm_LibFindComplInputGate(vFuncs, Mio_GateReadValue((Mio_Gate_t*)pFanout.pData), Abc_ObjFaninNum(pFanout), Abc_NodeFindFanin(pFanout, pPivot), nil) == -1 {
                    break; }
            }
            // update fanouts
            if i == Abc_ObjFanoutNum(pPivot) {
                // inlining of macro Abc_ObjForEachFanout
                for  i = 0; (i < Abc_ObjFanoutNum(pPivot)) && (((pFanout) = Abc_ObjFanout(pPivot, i)), 1); i++  {
                    var iFanin int32 = Abc_NodeFindFanin(pFanout, pPivot);
                    var iFaninNew int32 = -1;
                    var iGate int32  = Mio_GateReadValue((Mio_Gate_t*)pFanout.pData);
                    var iGateNew int32  = Sfm_LibFindComplInputGate( vFuncs, iGate, Abc_ObjFaninNum(pFanout), iFanin, &iFaninNew );
                    assert(iGateNew >= 0 && iGateNew != iGate && iFaninNew >= 0, "iGateNew >= 0 && iGateNew != iGate && iFaninNew >= 0");
                    pFanout.pData = Vec_PtrEntry( vGateHandles, iGateNew );
                    //assert( iFanin == iFaninNew );
                    // swap fanins
                    if iFanin != iFaninNew {
                        var pArray *int32  = Vec_IntArray( &pFanout.vFanins );
                        // inlining of macro ABC_SWAP
                        { var t int32  = pArray[iFanin]; pArray[iFanin] = pArray[iFaninNew]; pArray[iFaninNew] = t; };
                    }
                }
                iObj = Vec_WecEntryEntry( vFanins, Limit, 0 );
                pObjNew = Abc_NtkObj( pNtk, Vec_IntEntry(vMap, iObj) );
                Abc_ObjReplace( pPivot, pObjNew );
                // update level
                pObjNew.Level = 0;
                Abc_NtkUpdateIncLevel_rec( pObjNew );
                return pObjNew;
            }
        }
    }
    // introduce new gates
    // inlining of macro Vec_IntForEachEntryStart
    for  i = Limit; (i < Vec_IntSize(vGates)) && (((Gate) = Vec_IntEntry(vGates, i)), 1); i++  {
        vLevel = Vec_WecEntry( vFanins, i );
        pObjNew = Abc_NtkCreateNode( pNtk );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vLevel)) && (((iObj) = Vec_IntEntry(vLevel, k)), 1); k++  {
            Abc_ObjAddFanin( pObjNew, Abc_NtkObj(pNtk, Vec_IntEntry(vMap, iObj)) );
        }
        pObjNew.pData = Vec_PtrEntry( vGateHandles, Gate );
        assert(Abc_ObjFaninNum(pObjNew) == Mio_GateReadPinNum((Mio_Gate_t *)pObjNew.pData), "Abc_ObjFaninNum(pObjNew) == Mio_GateReadPinNum((Mio_Gate_t *)pObjNew.pData)");
        Vec_IntPush( vMap, Abc_ObjId(pObjNew) );
        if vNewNodes != nil {
            Vec_IntPush( vNewNodes, Abc_ObjId(pObjNew) ); }
    }
    // transfer load
    if pMit != nil {
        Sfm_MitTimingGrow( pMit );
        Sfm_MitTransferLoad( pMit, pObjNew, pPivot );
    }
    // replace logic cone
    Abc_ObjReplace( pPivot, pObjNew );
    // update level
    // inlining of macro Abc_NtkForEachObjVecStart
    for  i = Limit; i < Vec_IntSize(vMap) && (((pObjNew) = Abc_NtkObj(pNtk, Vec_IntEntry(vMap,i))), 1); i++  {         if (pObjNew) == nil {} else {
        Abc_NtkUpdateIncLevel_rec( pObjNew ); }
    }
    return pObjNew;
}
func Sfm_DecPrintStats(p *Sfm_Dec_t) {
    var i int32;
    fmt.Printf("Node = %d. Try = %d. Change = %d.   Const0 = %d. Const1 = %d. Buf = %d. Inv = %d. Gate = %d. AndOr = %d. Effort = %d.  NoDec = %d.\n", p.nTotalNodesBeg, p.nNodesTried, p.nNodesChanged, p.nNodesConst0, p.nNodesConst1, p.nNodesBuf, p.nNodesInv, p.nNodesResyn, p.nNodesAndOr, p.nEfforts, p.nNoDecs);
    fmt.Printf("MaxDiv = %d. MaxWin = %d.   AveDiv = %d. AveWin = %d.   Calls = %d. (Sat = %d. Unsat = %d.)  Over = %d.  T/O = %d.\n", p.nMaxDivs, p.nMaxWin, int32((p.nAllDivs/Abc_MaxInt(1, p.nNodesTried))), int32((p.nAllWin/Abc_MaxInt(1, p.nNodesTried))), p.nSatCalls, p.nSatCallsSat, p.nSatCallsUnsat, p.nSatCallsOver, p.nTimeOuts);

    p.timeTotal = Abc_Clock() - p.timeStart;
    p.timeOther = p.timeTotal - p.timeLib - p.timeWin - p.timeCnf - p.timeSat - p.timeTime;

    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Lib   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeLib)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeLib)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Win   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeWin)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeWin)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Cnf   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCnf)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeCnf)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Sat   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSat-p.timeEval)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSat-p.timeEval)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Sat  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatSat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatSat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", (" Unsat"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeSatUnsat)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeSatUnsat)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Eval  "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeEval)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeEval)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Timing"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTime)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTime)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("Other "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeOther)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeOther)))/(float64((p.timeTotal))) : 0.0);;
    // inlining of macro ABC_PRTP
    Abc_Print(1, "%s =", ("ALL   "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))), (float64((p.timeTotal)))? 100.0*(float64((p.timeTotal)))/(float64((p.timeTotal))) : 0.0);;

    fmt.Printf("Cone sizes:  ");
    for  i = 0; i <= SFM_SUPP_MAX; i++  {
        if p.nLuckySizes[i] {
            fmt.Printf("%d=%d  ", i, p.nLuckySizes[i]); }
    }
    fmt.Printf("  ");

    fmt.Printf("Gate sizes:  ");
    for  i = 0; i <= SFM_SUPP_MAX; i++  {
        if p.nLuckyGates[i] {
            fmt.Printf("%d=%d  ", i, p.nLuckyGates[i]); }
    }
    fmt.Printf("\n");

    fmt.Printf("Reduction:   ");
    fmt.Printf("Nodes  %6d out of %6d (%6.2f %%)   ", p.nTotalNodesBeg-p.nTotalNodesEnd, p.nTotalNodesBeg, 100.0*(p.nTotalNodesBeg-p.nTotalNodesEnd)/Abc_MaxInt(1, p.nTotalNodesBeg));
    fmt.Printf("Edges  %6d out of %6d (%6.2f %%)   ", p.nTotalEdgesBeg-p.nTotalEdgesEnd, p.nTotalEdgesBeg, 100.0*(p.nTotalEdgesBeg-p.nTotalEdgesEnd)/Abc_MaxInt(1, p.nTotalEdgesBeg));
    fmt.Printf("\n");
}
func Abc_NtkCountStats(p *Sfm_Dec_t, Limit int32) {
    var Gate int32;
    var nGates int32 = Vec_IntSize(&p.vObjGates);
    if nGates == Limit {
        return; }
    Gate = Vec_IntEntryLast(&p.vObjGates);
    if nGates > Limit + 1 {
        p.nNodesResyn++;
    } else if Gate == p.GateConst0 {
        p.nNodesConst0++;
    } else if Gate == p.GateConst1 {
        p.nNodesConst1++;
    } else if Gate == p.GateBuffer {
        p.nNodesBuf++;
    } else if Gate == p.GateInvert {
        p.nNodesInv++;
    } else { 
        p.nNodesResyn++; }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkAreaOptOne(p *Sfm_Dec_t, i int32) *Abc_Obj_t {
    var clk abctime;
    var pNtk *Abc_Ntk_t  = p.pNtk;
    var pPars *Sfm_Par_t  = p.pPars;
    var pObj *Abc_Obj_t  = Abc_NtkObj( p.pNtk, i ); 
    var Limit, RetValue int32;
    if pPars.nMffcMin > 1 && Abc_NodeMffcLabel(pObj, nil) < pPars.nMffcMin {
        return nil; }
    if pPars.iNodeOne && i != pPars.iNodeOne {
        return nil; }
    if pPars.iNodeOne {
        pPars.fVeryVerbose = int32((i == pPars.iNodeOne)); }
    p.nNodesTried++;
clk = Abc_Clock();
    p.nDivs = Sfm_DecExtract( pNtk, pPars, pObj, &p.vObjRoots, &p.vObjGates, &p.vObjFanins, &p.vObjMap, &p.vGateTfi, &p.vGateTfo, &p.vObjMffc, &p.vObjInMffc, nil, nil );
p.timeWin += Abc_Clock() - clk;
    if pPars.nWinSizeMax && pPars.nWinSizeMax < Vec_IntSize(&p.vObjGates) {
        return nil; }
    p.nMffc = Vec_IntSize(&p.vObjMffc);
    p.AreaMffc = Sfm_DecMffcArea(pNtk, &p.vObjMffc);
    p.nMaxDivs = Abc_MaxInt( p.nMaxDivs, p.nDivs );
    p.nAllDivs += p.nDivs;
    p.iTarget = pObj.iTemp;
    Limit = Vec_IntSize( &p.vObjGates );
    p.nMaxWin = Abc_MaxInt( p.nMaxWin, Limit );
    p.nAllWin += Limit;
clk = Abc_Clock();
    RetValue = Sfm_DecPrepareSolver( p );
p.timeCnf += Abc_Clock() - clk;
    if !(RetValue != 0) {
        return nil; }
clk = Abc_Clock();
    RetValue = Sfm_DecPeformDec2( p, pObj );
    if pPars.fMoreEffort && RetValue < 0 {
        var Var, i int32;
        // inlining of macro Vec_IntForEachEntryReverse
        for  i = Vec_IntSize(&p.vObjInMffc) - 1; (i >= 0) && (((Var) = Vec_IntEntry(&p.vObjInMffc, i)), 1); i--  {
            p.iUseThis = Var; 
            RetValue = Sfm_DecPeformDec2( p, pObj );
            p.iUseThis = -1;
            if RetValue < 0 {
                //printf( "Node %d: Not found among %d.\n", Abc_ObjId(pObj), Vec_IntSize(&p->vObjInMffc) );
            
            } else {
                p.nEfforts++;
                if p.pPars.fVerbose {
                    //printf( "Node %5d: (%2d out of %2d)  Gate=%s ", Abc_ObjId(pObj), i, Vec_IntSize(&p->vObjInMffc), Mio_GateReadName((Mio_Gate_t*)pObj->pData) );
                    //Dau_DsdPrintFromTruth( p->Copy, p->nSuppVars );
                }
                break;
            }
        }
    }
    if p.pPars.fVeryVerbose {
        fmt.Printf("\n\n"); }
p.timeSat += Abc_Clock() - clk;
    if RetValue < 0 {
        return nil; }
    p.nNodesChanged++;
    Abc_NtkCountStats( p, Limit );
    return Sfm_DecInsert( pNtk, pObj, Limit, &p.vObjGates, &p.vObjFanins, &p.vObjMap, &p.vGateHands, p.GateBuffer, p.GateInvert, &p.vGateFuncs, nil, p.pMit );
}
func Abc_NtkAreaOpt(p *Sfm_Dec_t) {
    var pObj *Abc_Obj_t; 
    var i int32;
    var nStop int32 = Abc_NtkObjNumMax(p.pNtk);
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {           if (pObj) == nil || !Abc_ObjIsNode(pObj) {} else {
        if i >= nStop || (p.pPars.nNodesMax && i > p.pPars.nNodesMax) {
            break; }
        Abc_NtkAreaOptOne( p, i );
    }
    }
}
func Abc_NtkAreaOpt2(p *Sfm_Dec_t) {
    var pObj, pObjNew, pFanin *Abc_Obj_t; 
    var i int32;
    var k int32;
    var nStop int32 = Abc_NtkObjNumMax(p.pNtk);
    var vFront *Vec_Ptr_t  = Vec_PtrAlloc( 1000 );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        assert(pObj.fMarkB == 0, "pObj.fMarkB == 0"); }
    }
    // start the queue of nodes to be tried
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pObj) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        if Abc_ObjIsNode(Abc_ObjFanin0(pObj)) && !Abc_ObjFanin0(pObj).fMarkB {
            Abc_ObjFanin0(pObj).fMarkB = 1;
            Vec_PtrPush( vFront, Abc_ObjFanin0(pObj) );
        }
    }
    // process nodes in this order
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vFront)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vFront, i)), 1); i++  {
        if Abc_ObjIsNone(pObj) {
            continue; }
        pObjNew = Abc_NtkAreaOptOne( p, Abc_ObjId(pObj) );
        if pObjNew != nil {
            if !Abc_ObjIsNode(pObjNew) || Abc_ObjFaninNum(pObjNew) == 0 || pObjNew.fMarkB {
                continue; }
            if int32(Abc_ObjId(pObjNew)) < nStop {
                pObjNew.fMarkB = 1;
                Vec_PtrPush( vFront, pObjNew );
                continue;
            }
        
        } else {
            pObjNew = pObj; }
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObjNew)) && (((pFanin) = Abc_ObjFanin(pObjNew, k)), 1); k++  {
            if Abc_ObjIsNode(pFanin) && Abc_ObjFaninNum(pObjNew) > 0 && !pFanin.fMarkB {
                pFanin.fMarkB = 1;
                Vec_PtrPush( vFront, pFanin );
            }
        }
    }
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((p.pNtk).vObjs)) && (((pObj) = Abc_NtkObj(p.pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        pObj.fMarkB = 0; }
    }
    Vec_PtrFree( vFront );
}

func Abc_NtkDelayOpt(p *Sfm_Dec_t) {
    var pNtk *Abc_Ntk_t  = p.pNtk;
    var pPars *Sfm_Par_t  = p.pPars; var n int32;
    Abc_NtkCleanMarkABC( pNtk );
    for  n = 0; pPars.nNodesMax == 0 || n < pPars.nNodesMax; n++  {
        var pObj, pObjNew *Abc_Obj_t; var clk abctime;
        var i, Limit, RetValue int32;
        // collect nodes
        if pPars.iNodeOne {
            Vec_IntFill( &p.vCands, 1, pPars.iNodeOne );
        } else if p.pTim && !Sfm_TimPriorityNodes(p.pTim, &p.vCands, p.pPars.nTimeWin) {
            break;
        } else if p.pMit && !Sfm_MitPriorityNodes(p.pMit, &p.vCands, p.pPars.nTimeWin) {
            break; }
        // try improving delay for the nodes according to the priority
        // inlining of macro Abc_NtkForEachObjVec
        for  i = 0; i < Vec_IntSize(&p.vCands) && (((pObj) = Abc_NtkObj(p.pNtk, Vec_IntEntry(&p.vCands,i))), 1); i++  {         if (pObj) == nil {} else {
            var OldId int32  = Abc_ObjId(pObj);
            var DelayOld int32  = Sfm_ManReadObjDelay(p, OldId);
            assert(pObj.fMarkA == 0, "pObj.fMarkA == 0");
            p.nNodesTried++;
clk = Abc_Clock();
            p.nDivs = Sfm_DecExtract( pNtk, pPars, pObj, &p.vObjRoots, &p.vObjGates, &p.vObjFanins, &p.vObjMap, &p.vGateTfi, &p.vGateTfo, &p.vObjMffc, &p.vObjInMffc, p.pTim, p.pMit );
p.timeWin += Abc_Clock() - clk;
            if p.nDivs < 2 || (pPars.nWinSizeMax && pPars.nWinSizeMax < Vec_IntSize(&p.vObjGates)) { 
                pObj.fMarkA = 1;
                continue;
            }
            p.nMffc = Vec_IntSize(&p.vObjMffc);
            p.AreaMffc = Sfm_DecMffcArea(pNtk, &p.vObjMffc);
            p.nMaxDivs = Abc_MaxInt( p.nMaxDivs, p.nDivs );
            p.nAllDivs += p.nDivs;
            p.iTarget = pObj.iTemp;
            Limit = Vec_IntSize( &p.vObjGates );
            p.nMaxWin = Abc_MaxInt( p.nMaxWin, Limit );
            p.nAllWin += Limit;
clk = Abc_Clock();
            RetValue = Sfm_DecPrepareSolver( p );
p.timeCnf += Abc_Clock() - clk;
            if !(RetValue != 0) {
                pObj.fMarkA = 1;
                continue;
            }
clk = Abc_Clock();
            RetValue = Sfm_DecPeformDec3( p, pObj );
            if pPars.fMoreEffort && RetValue < 0 {
                var Var, i int32;
                // inlining of macro Vec_IntForEachEntryReverse
                for  i = Vec_IntSize(&p.vObjInMffc) - 1; (i >= 0) && (((Var) = Vec_IntEntry(&p.vObjInMffc, i)), 1); i--  {
                    p.iUseThis = Var; 
                    RetValue = Sfm_DecPeformDec3( p, pObj );
                    p.iUseThis = -1;
                    if RetValue < 0 {
                        //printf( "Node %d: Not found among %d.\n", Abc_ObjId(pObj), Vec_IntSize(&p->vObjInMffc) );
                    
                    } else {
                        p.nEfforts++;
                        if p.pPars.fVerbose {
                            //printf( "Node %5d: (%2d out of %2d)  Gate=%s ", Abc_ObjId(pObj), i, Vec_IntSize(&p->vObjInMffc), Mio_GateReadName((Mio_Gate_t*)pObj->pData) );
                            //Dau_DsdPrintFromTruth( p->Copy, p->nSuppVars );
                        }
                        break;
                    }
                }
            }
            if p.pPars.fVeryVerbose {
                fmt.Printf("\n\n"); }
p.timeSat += Abc_Clock() - clk;
            if RetValue < 0 {
                pObj.fMarkA = 1;
                continue;
            }
            assert(Vec_IntSize(&p.vObjGates) - Limit > 0, "Vec_IntSize(&p.vObjGates) - Limit > 0");
            assert(Vec_IntSize(&p.vObjGates) - Limit <= 2, "Vec_IntSize(&p.vObjGates) - Limit <= 2");
            p.nNodesChanged++;
            Abc_NtkCountStats( p, Limit );
            // reduce load due to removed MFFC
            if p.pMit { Sfm_MitUpdateLoad( p.pMit, &p.vGateMffc, 0 ); } // assuming &p->vGateMffc contains MFFC
            Sfm_DecInsert( pNtk, pObj, Limit, &p.vObjGates, &p.vObjFanins, &p.vObjMap, &p.vGateHands, p.GateBuffer, p.GateInvert, &p.vGateFuncs, &p.vNewNodes, p.pMit );
            // increase load due to added new nodes
            if p.pMit { Sfm_MitUpdateLoad( p.pMit, &p.vNewNodes, 1 ); } // assuming &p->vNewNodes contains new nodes
clk = Abc_Clock();
            if p.pMit {
                Sfm_MitUpdateTiming( p.pMit, &p.vNewNodes );
            } else {
                Sfm_TimUpdateTiming( p.pTim, &p.vNewNodes ); }
p.timeTime += Abc_Clock() - clk;
            pObjNew = Abc_NtkObj( pNtk, Abc_NtkObjNumMax(pNtk)-1 );
            assert(p.pMit || p.DelayMin == 0 || p.DelayMin == Sfm_ManReadObjDelay(p, Abc_ObjId(pObjNew)), "p.pMit || p.DelayMin == 0 || p.DelayMin == Sfm_ManReadObjDelay(p, Abc_ObjId(pObjNew))");
            // report
            if pPars.fDelayVerbose {
                fmt.Printf("Node %5d  %5d :  I =%3d.  Cand = %5d (%6.2f %%)   Old =%8.2f.  New =%8.2f.  Final =%8.2f.  WNS =%8.2f.\n", OldId, Abc_NtkObjNumMax(p.pNtk), i, Vec_IntSize(&p.vCands), 100.0 * Vec_IntSize(&p.vCands) / Abc_NtkNodeNum(p.pNtk), Scl_Int2Flt(DelayOld), Scl_Int2Flt(Sfm_ManReadObjDelay(p, Abc_ObjId(pObjNew))), Scl_Int2Flt(Sfm_ManReadNtkDelay(p)), Scl_Int2Flt(Sfm_ManReadNtkMinSlack(p))); }
            break;
        }
        }
        if pPars.iNodeOne {
            break; }
    }
    Abc_NtkCleanMarkABC( pNtk );
}
func Abc_NtkPerformMfs3(pNtk *Abc_Ntk_t, pPars *Sfm_Par_t) {
    var p *Sfm_Dec_t  = Sfm_DecStart( pPars, (Mio_Library_t *)pNtk.pManFunc, pNtk );
    if pPars.fVerbose {
        fmt.Printf("Remapping parameters: ");
        if pPars.nTfoLevMax {
            fmt.Printf("TFO = %d. ", pPars.nTfoLevMax); }
        if pPars.nTfiLevMax {
            fmt.Printf("TFI = %d. ", pPars.nTfiLevMax); }
        if pPars.nFanoutMax {
            fmt.Printf("FanMax = %d. ", pPars.nFanoutMax); }
        if pPars.nWinSizeMax {
            fmt.Printf("WinMax = %d. ", pPars.nWinSizeMax); }
        if pPars.nBTLimit {
            fmt.Printf("Confl = %d. ", pPars.nBTLimit); }
        if pPars.nMffcMin && pPars.fArea {
            fmt.Printf("MffcMin = %d. ", pPars.nMffcMin); }
        if pPars.nMffcMax && pPars.fArea {
            fmt.Printf("MffcMax = %d. ", pPars.nMffcMax); }
        if pPars.nDecMax {
            fmt.Printf("DecMax = %d. ", pPars.nDecMax); }
        if pPars.iNodeOne {
            fmt.Printf("Pivot = %d. ", pPars.iNodeOne); }
        if !pPars.fArea {
            fmt.Printf("Win = %d. ", pPars.nTimeWin); }
        if !pPars.fArea {
            fmt.Printf("Delta = %.2f ps. ", Scl_Int2Flt(p.DeltaCrit)); }
        if pPars.fArea {
            fmt.Printf("0-cost = %s. ", pPars.fZeroCost ? "yes" : "no"); }
        fmt.Printf("Effort = %s. ", pPars.fMoreEffort ? "yes" : "no");
        fmt.Printf("Sim = %s. ", pPars.fUseSim ? "yes" : "no");
        fmt.Printf("\n");
    }
    // preparation steps
    Abc_NtkLevel( pNtk );
    if p.pPars.fUseSim {
        Sfm_NtkSimulate( pNtk ); }
    // record statistics
    if pPars.fVerbose { p.nTotalNodesBeg = Abc_NtkNodeNum(pNtk); }
    if pPars.fVerbose { p.nTotalEdgesBeg = Abc_NtkGetTotalFanins(pNtk); }
    // perform optimization
    if pPars.fArea {
        if pPars.fAreaRev {
            Abc_NtkAreaOpt2( p );
        } else {
            Abc_NtkAreaOpt( p ); }
    
    } else {
        Abc_NtkDelayOpt( p ); }
    // record statistics
    if pPars.fVerbose { p.nTotalNodesEnd = Abc_NtkNodeNum(pNtk); }
    if pPars.fVerbose { p.nTotalEdgesEnd = Abc_NtkGetTotalFanins(pNtk); }
    // print stats and quit
    if pPars.fVerbose {
    Sfm_DecPrintStats( p ); }
    if pPars.fLibVerbose {
        Sfm_LibPrint( p.pLib ); }
    Sfm_DecStop( p );
    if pPars.fArea {
        extern void Abc_NtkChangePerform( Abc_Ntk_t * pNtk, int fVerbose );
        Abc_NtkChangePerform( pNtk, pPars.fVerbose );
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

