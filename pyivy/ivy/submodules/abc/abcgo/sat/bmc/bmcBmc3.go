package bmc

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [bmcBmc3.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Simple BMC package.]

  Author      [Alan Mishchenko in collaboration with Niklas Een.]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcBmc3.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "proof/fra/fra.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "sat/satoko/satoko.h"
//#include "sat/glucose/AbcGlucose.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/vec/vecWec.h"
//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Gia_ManBmc_t struct {
    // input/output data
    pPars    *Saig_ParBmc_t;       // parameters
    pAig        *Aig_Man_t;        // user AIG
    vCexes        *Vec_Ptr_t;      // counter-examples
    // intermediate data
    vMapping        *Vec_Int_t;    // mapping
    vMapRefs        *Vec_Int_t;    // mapping references
//    Vec_Vec_t *       vSects;      // sections
    vId2Num        *Vec_Int_t;     // number of each node 
    vTerInfo        *Vec_Ptr_t;    // ternary information
    vId2Var        *Vec_Ptr_t;     // SAT vars for each object
    vVisited        *Vec_Wec_t;    // visited nodes
    pTime4Outs          *abctime;  // timeout per output
    // hash table
    vData        *Vec_Int_t;       // storage for cuts
    vHash     *Hsh_IntMan_t;       // hash table
    vId2Lit        *Vec_Int_t;     // mapping cuts into literals
    nHashHit               int32;    // hash table hits
    nHashMiss               int32;   // hash table misses
    nBufNum               int32;     // the number of simple nodes
    nDupNum               int32;     // the number of simple nodes
    nUniProps               int32;   // propagating learned clause values
    nLitUsed               int32;    // used literals
    nLitUseless               int32; // useless literals
    // SAT solver
    pSat       *sat_solver;        // SAT solver
    pSat2         *satoko_t;       // SAT solver
    pSat3  *bmcg_sat_solver;       // SAT solver
    nSatVars               int32;    // SAT variables
    nObjNums               int32;    // SAT objects
    nWordNum               int32;    // unsigned words for ternary simulation
    pSopSizes string;
pSops *string;    // CNF representation
}

// extern int Gia_ManToBridgeResult( FILE * pFile, int Result, Abc_Cex_t * pCex, int iPoProved );


func Gia_ManReportProgress(pFile *os.File, prop_no int32, depth int32) {
    extern int Gia_ManToBridgeProgress( FILE * pFile, int Size, unsigned char * pBuffer );
    var buf string // was [100]byte in C

    buf = fmt.Sprintf("property: safe<%d>\nbug-free-depth: %d\n", prop_no, depth);
    Gia_ManToBridgeProgress(pFile, strlen(buf), (unsigned char *)buf);
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

const SAIG_TER_NON = 0
const SAIG_TER_ZER = 1
const SAIG_TER_ONE = 2
const SAIG_TER_UND = 3

func Saig_ManBmcSimInfoNot(Value int32) int32 {
    if Value == SAIG_TER_ZER {
        return SAIG_TER_ONE; }
    if Value == SAIG_TER_ONE {
        return SAIG_TER_ZER; }
    return SAIG_TER_UND;
}

func Saig_ManBmcSimInfoAnd(Value0 int32, Value1 int32) int32 {
    if Value0 == SAIG_TER_ZER || Value1 == SAIG_TER_ZER {
        return SAIG_TER_ZER; }
    if Value0 == SAIG_TER_ONE && Value1 == SAIG_TER_ONE {
        return SAIG_TER_ONE; }
    return SAIG_TER_UND;
}

func Saig_ManBmcSimInfoGet(pInfo *uint32, pObj *Aig_Obj_t) int32 {
    return 3 & (pInfo[Aig_ObjId(pObj) >> 4] >> ((Aig_ObjId(pObj) & 15) << 1));
}

func Saig_ManBmcSimInfoSet(pInfo *uint32, pObj *Aig_Obj_t, Value int32) {
    assert(Value >= SAIG_TER_ZER && Value <= SAIG_TER_UND, "Value >= SAIG_TER_ZER && Value <= SAIG_TER_UND");
    Value ^= Saig_ManBmcSimInfoGet( pInfo, pObj );
    pInfo[Aig_ObjId(pObj) >> 4] ^= (Value << ((Aig_ObjId(pObj) & 15) << 1));
}

func Saig_ManBmcSimInfoClear(pInfo *uint32, pObj *Aig_Obj_t) int32 {
    var Value int32  = Saig_ManBmcSimInfoGet( pInfo, pObj );
    pInfo[Aig_ObjId(pObj) >> 4] ^= (Value << ((Aig_ObjId(pObj) & 15) << 1));
    return Value;
}

/**Function*************************************************************

  Synopsis    [Returns the number of LIs with binary ternary info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcTerSimCount01(p *Aig_Man_t, pInfo *uint32) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    if pInfo == nil {
        return Saig_ManRegNum(p); }
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i+Saig_ManPoNum(p))), 1); i++  {
        if !Aig_ObjIsConst1(Aig_ObjFanin0(pObj)) {
            Counter += (Saig_ManBmcSimInfoGet(pInfo, pObj) != SAIG_TER_UND); }
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation of one frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcTerSimOne(p *Aig_Man_t, pPrev *uint32) *uint32 {
    var pObj, pObjLi *Aig_Obj_t;
    var pInfo *uint32;
    var i, Val0, Val1 int32;
    // inlining of macro ABC_CALLOC
    pInfo = ((unsigned *) calloc((size_t)(Abc_BitWordNum(2 * Aig_ManObjNumMax(p))), sizeof(unsigned)));
    Saig_ManBmcSimInfoSet( pInfo, Aig_ManConst1(p), SAIG_TER_ONE );
    // inlining of macro Saig_ManForEachPi
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPiNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        Saig_ManBmcSimInfoSet( pInfo, pObj, SAIG_TER_UND );
    }
    if pPrev == nil {
        // inlining of macro Saig_ManForEachLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i+Saig_ManPiNum(p))), 1); i++  {
            Saig_ManBmcSimInfoSet( pInfo, pObj, SAIG_TER_ZER );
        }
    
    } else {
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p)) && (((pObjLi) = Saig_ManLi(p, i)), 1)            && (((pObj)=Saig_ManLo(p, i)), 1); i++  {
            Saig_ManBmcSimInfoSet( pInfo, pObj, Saig_ManBmcSimInfoGet(pPrev, pObjLi) );
        }
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        Val0 = Saig_ManBmcSimInfoGet( pInfo, Aig_ObjFanin0(pObj) );
        Val1 = Saig_ManBmcSimInfoGet( pInfo, Aig_ObjFanin1(pObj) );
        if Aig_ObjFaninC0(pObj) {
            Val0 = Saig_ManBmcSimInfoNot( Val0 ); }
        if Aig_ObjFaninC1(pObj) {
            Val1 = Saig_ManBmcSimInfoNot( Val1 ); }
        Saig_ManBmcSimInfoSet( pInfo, pObj, Saig_ManBmcSimInfoAnd(Val0, Val1) );
    }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Val0 = Saig_ManBmcSimInfoGet( pInfo, Aig_ObjFanin0(pObj) );
        if Aig_ObjFaninC0(pObj) {
            Val0 = Saig_ManBmcSimInfoNot( Val0 ); }
        Saig_ManBmcSimInfoSet( pInfo, pObj, Val0 );
    }
    return pInfo;    
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes and PIs in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcTerSim(p *Aig_Man_t) *Vec_Ptr_t {
    var vInfos *Vec_Ptr_t;
    var pInfo *uint32  = nil;
    var i int32;
    var TerPrev int32 = ABC_INFINITY;
    var TerCur int32;
    var CountIncrease int32;
    vInfos = Vec_PtrAlloc( 100 );
    for  i = 0; i < 1000 && CountIncrease < 5 && TerPrev > 0; i++  {
        TerCur = Saig_ManBmcTerSimCount01( p, pInfo );
        if TerCur >= TerPrev {
            CountIncrease++; }
        TerPrev = TerCur;
        pInfo = Saig_ManBmcTerSimOne( p, pInfo );
        Vec_PtrPush( vInfos, pInfo );
    }
    return vInfos;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcTerSimTest(p *Aig_Man_t) {
    var vInfos *Vec_Ptr_t;
    var pInfo *uint32;
    var i int32;
    vInfos = Saig_ManBmcTerSim( p );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vInfos)) && (((pInfo) = (unsigned *)Vec_PtrEntry(vInfos, i)), 1); i++  {
        Abc_Print( 1, "%d=%d ", i, Saig_ManBmcTerSimCount01(p, pInfo) );
    }
    Abc_Print( 1, "\n" );
    Vec_PtrFreeFree( vInfos );
}



/**Function*************************************************************

  Synopsis    [Count the number of non-ternary per frame.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcCountNonternary_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vInfos *Vec_Ptr_t, pInfo *uint32, f int32, pCounter *int32) int32 { 
    var Value int32  = Saig_ManBmcSimInfoClear( pInfo, pObj );
    if Value == SAIG_TER_NON {
        return 0; }
    assert(f >= 0, "f >= 0");
    pCounter[f] += (Value == SAIG_TER_UND);
    if Saig_ObjIsPi(p, pObj) || (f == 0 && Saig_ObjIsLo(p, pObj)) || Aig_ObjIsConst1(pObj) {
        return 0; }
    if Saig_ObjIsLi(p, pObj) {
        return Saig_ManBmcCountNonternary_rec( p, Aig_ObjFanin0(pObj), vInfos, pInfo, f, pCounter ); }
    if Saig_ObjIsLo(p, pObj) {
        return Saig_ManBmcCountNonternary_rec( p, Saig_ObjLoToLi(p, pObj), vInfos, (unsigned *)Vec_PtrEntry(vInfos, f-1), f-1, pCounter ); }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Saig_ManBmcCountNonternary_rec( p, Aig_ObjFanin0(pObj), vInfos, pInfo, f, pCounter );
    Saig_ManBmcCountNonternary_rec( p, Aig_ObjFanin1(pObj), vInfos, pInfo, f, pCounter );
    return 0;
}
func Saig_ManBmcCountNonternary(p *Aig_Man_t, vInfos *Vec_Ptr_t, iFrame int32) {
    // inlining of macro ABC_CALLOC
    int i, * pCounters = ((int *) calloc((size_t)(iFrame + 1), sizeof(int)));
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry(vInfos, iFrame);
    assert(Saig_ManBmcSimInfoGet( pInfo, Aig_ManCo(p, 0) ) == SAIG_TER_UND, "Saig_ManBmcSimInfoGet( pInfo, Aig_ManCo(p, 0) ) == SAIG_TER_UND");
    Saig_ManBmcCountNonternary_rec( p, Aig_ObjFanin0(Aig_ManCo(p, 0)), vInfos, pInfo, iFrame, pCounters );
    for  i = 0; i <= iFrame; i++  {
        Abc_Print( 1, "%d=%d ", i, pCounters[i] );
    }
    Abc_Print( 1, "\n" );
    pCounters = nil;
}


/**Function*************************************************************

  Synopsis    [Returns the number of POs with binary ternary info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcTerSimCount01Po(p *Aig_Man_t, pInfo *uint32) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Counter += (Saig_ManBmcSimInfoGet(pInfo, pObj) != SAIG_TER_UND);
    }
    return Counter;
}
func Saig_ManBmcTerSimPo(p *Aig_Man_t) *Vec_Ptr_t {
    var vInfos *Vec_Ptr_t;
    var pInfo *uint32  = nil;
    var i, nPoBin int32;
    vInfos = Vec_PtrAlloc( 100 );
    for  i = 0; ; i++  {
        if i % 100 == 0 {
            Abc_Print( 1, "Frame %5d\n", i ); }
        pInfo = Saig_ManBmcTerSimOne( p, pInfo );
        Vec_PtrPush( vInfos, pInfo );
        nPoBin = Saig_ManBmcTerSimCount01Po( p, pInfo );
        if nPoBin < Saig_ManPoNum(p) {
            break; }
    }
    Abc_Print( 1, "Detected terminary PO in frame %d.\n", i );
    Saig_ManBmcCountNonternary( p, vInfos, i );
    return vInfos;
}
func Saig_ManBmcTerSimTestPo(p *Aig_Man_t) {
    var vInfos *Vec_Ptr_t;
    vInfos = Saig_ManBmcTerSimPo( p );
    Vec_PtrFreeFree( vInfos );
}




/**Function*************************************************************

  Synopsis    [Collects internal nodes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcDfs_rec(p *Aig_Man_t, pObj *Aig_Obj_t, vNodes *Vec_Ptr_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p, pObj);
    if Aig_ObjIsNode(pObj) {
        Saig_ManBmcDfs_rec( p, Aig_ObjFanin0(pObj), vNodes );
        Saig_ManBmcDfs_rec( p, Aig_ObjFanin1(pObj), vNodes );
    }
    Vec_PtrPush( vNodes, pObj );
}

/**Function*************************************************************

  Synopsis    [Collects internal nodes and PIs in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcDfsNodes(p *Aig_Man_t, vRoots *Vec_Ptr_t) *Vec_Ptr_t {
    var vNodes *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    vNodes = Vec_PtrAlloc( 100 );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vRoots)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vRoots, i)), 1); i++  {
        Saig_ManBmcDfs_rec( p, Aig_ObjFanin0(pObj), vNodes );
    }
    return vNodes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcSections(p *Aig_Man_t) *Vec_Vec_t {
    var vSects, vRoots, vCone *Vec_Ptr_t;
    var pObj, pObjPo *Aig_Obj_t;
    var i int32;
    Aig_ManIncrementTravId( p );
    Aig_ObjSetTravIdCurrent( p, Aig_ManConst1(p) );
    // start the roots
    vRoots = Vec_PtrAlloc( 1000 );
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObjPo) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Aig_ObjSetTravIdCurrent( p, pObjPo );
        Vec_PtrPush( vRoots, pObjPo );
    }
    // compute the cones
    vSects = Vec_PtrAlloc( 20 );
    for  Vec_PtrSize(vRoots) > 0  {
        vCone = Saig_ManBmcDfsNodes( p, vRoots );
        Vec_PtrPush( vSects, vCone );
        // get the next set of roots
        Vec_PtrClear( vRoots );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vCone)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vCone, i)), 1); i++  {
            if !Saig_ObjIsLo(p, pObj) {
                continue; }
            pObjPo = Saig_ObjLoToLi( p, pObj );
            if Aig_ObjIsTravIdCurrent(p, pObjPo) {
                continue; }
            Aig_ObjSetTravIdCurrent( p, pObjPo );
            Vec_PtrPush( vRoots, pObjPo );
        }
    }
    Vec_PtrFree( vRoots );
    return (Vec_Vec_t *)vSects;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcSectionsTest(p *Aig_Man_t) {
    var vSects *Vec_Vec_t;
    var vOne *Vec_Ptr_t;
    var i int32;
    vSects = Saig_ManBmcSections( p );
    // inlining of macro Vec_VecForEachLevel
    for  i = 0; (i < Vec_VecSize(vSects)) && (((vOne) = Vec_VecEntry(vSects, i)), 1); i++  {
        Abc_Print( 1, "%d=%d ", i, Vec_PtrSize(vOne) );
    }
    Abc_Print( 1, "\n" );
    Vec_VecFree( vSects );
}



/**Function*************************************************************

  Synopsis    [Collects the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcSupergate_rec(pObj *Aig_Obj_t, vSuper *Vec_Ptr_t) {
    // if the new node is complemented or a PI, another gate begins
    if Aig_IsComplement(pObj) || Aig_ObjIsCi(pObj) { // || (Aig_ObjRefs(pObj) > 1) )
    
        Vec_PtrPushUnique( vSuper, Aig_Regular(pObj) );
        return;
    }
    // go through the branches
    Saig_ManBmcSupergate_rec( Aig_ObjChild0(pObj), vSuper );
    Saig_ManBmcSupergate_rec( Aig_ObjChild1(pObj), vSuper );
}

/**Function*************************************************************

  Synopsis    [Collect the topmost supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcSupergate(p *Aig_Man_t, iPo int32) *Vec_Ptr_t {
    var vSuper *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    vSuper = Vec_PtrAlloc( 10 );
    pObj = Aig_ManCo( p, iPo );
    pObj = Aig_ObjChild0( pObj );
    if !Aig_IsComplement(pObj) {
        Vec_PtrPush( vSuper, pObj );
        return vSuper;
    }
    pObj = Aig_Regular( pObj );
    if !Aig_ObjIsNode(pObj) {
        Vec_PtrPush( vSuper, pObj );
        return vSuper;
    }
    Saig_ManBmcSupergate_rec( Aig_ObjChild0(pObj), vSuper );
    Saig_ManBmcSupergate_rec( Aig_ObjChild1(pObj), vSuper );
    return vSuper;
}

/**Function*************************************************************

  Synopsis    [Returns the number of nodes with ref counter more than 1.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcCountRefed(p *Aig_Man_t, vSuper *Vec_Ptr_t) int32 {
    var pObj *Aig_Obj_t;
    var i, Counter int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vSuper)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(vSuper, i)), 1); i++  {
        assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
        Counter += (Aig_ObjRefs(pObj) > 1);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcSupergateTest(p *Aig_Man_t) {
    var vSuper *Vec_Ptr_t;
    var pObj *Aig_Obj_t;
    var i int32;
    Abc_Print( 1, "Supergates: " );
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        vSuper = Saig_ManBmcSupergate( p, i );
        Abc_Print( 1, "%d=%d(%d) ", i, Vec_PtrSize(vSuper), Saig_ManBmcCountRefed(p, vSuper) );
        Vec_PtrFree( vSuper );
    }
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcWriteBlif(p *Aig_Man_t, vMapping *Vec_Int_t, pFileName string) {
    var pFile *os.File;
    var pSopSizes string;
    var pSops *string;
    var pObj *Aig_Obj_t;
    var Vals [4]byte;
    var i int32;
    var k int32;
    var b int32;
    var iFan int32;
    var iTruth int32;
    var pData *int32;
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        Abc_Print( 1, "Cannot open file %s\n", pFileName );
        return;
    }
    fmt.Fprintf(pFile, ".model test\n");
    fmt.Fprintf(pFile, ".inputs");
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCis, i)), 1); i++  {
        fmt.Fprintf(pFile, " n%d", Aig_ObjId(pObj));
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".outputs");
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, " n%d", Aig_ObjId(pObj));
    }
    fmt.Fprintf(pFile, "\n");
    fmt.Fprintf(pFile, ".names");
    fmt.Fprintf(pFile, " n%d\n", Aig_ObjId(Aig_ManConst1(p)));
    fmt.Fprintf(pFile, " 1\n");

    Cnf_ReadMsops( &pSopSizes, &pSops );
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        if Vec_IntEntry(vMapping, i) == 0 {
            continue; }
        pData = Vec_IntEntryP( vMapping, Vec_IntEntry(vMapping, i) );
        fmt.Fprintf(pFile, ".names");
        for  iFan = 0; iFan < 4; iFan++  {
            if pData[iFan+1] >= 0 {
                fmt.Fprintf(pFile, " n%d", pData[iFan+1]);
            } else {
                break; }
        }
        fmt.Fprintf(pFile, " n%d\n", i);
        // write SOP
        iTruth = pData[0] & 0xffff;
        for  k = 0; k < pSopSizes[iTruth]; k++  {
            var Lit int32  = pSops[iTruth][k];
            for  b = 3; b >= 0; b--  {
                if Lit % 3 == 0 {
                    Vals[b] = '0';
                } else if Lit % 3 == 1 {
                    Vals[b] = '1';
                } else {
                    Vals[b] = '-'; }
                Lit = Lit / 3;
            }
            for  b = 0; b < iFan; b++  {
                fmt.Fprintf(pFile, "%c", Vals[b]);
            }
            fmt.Fprintf(pFile, " 1\n");
        }
    }
    }
    free( pSopSizes );
    free( pSops[1] );
    free( pSops );

    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        fmt.Fprintf(pFile, ".names");
        fmt.Fprintf(pFile, " n%d", Aig_ObjId(Aig_ObjFanin0(pObj)));
        fmt.Fprintf(pFile, " n%d\n", Aig_ObjId(pObj));
        fmt.Fprintf(pFile, "%d 1\n", !Aig_ObjFaninC0(pObj));
    }
    fmt.Fprintf(pFile, ".end\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcMappingTest(p *Aig_Man_t) {
    var vMapping *Vec_Int_t;
    vMapping = Cnf_DeriveMappingArray( p );
    Saig_ManBmcWriteBlif( p, vMapping, "test.blif" );
    Vec_IntFree( vMapping );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcComputeMappingRefs(p *Aig_Man_t, vMap *Vec_Int_t) *Vec_Int_t {
    var vRefs *Vec_Int_t;
    var pObj *Aig_Obj_t;
    var i int32;
    var iFan int32;
    var pData *int32;
    vRefs = Vec_IntStart( Aig_ManObjNumMax(p) );
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vCos, i)), 1); i++  {
        Vec_IntAddToEntry( vRefs, Aig_ObjFaninId0(pObj), 1 );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
        if Vec_IntEntry(vMap, i) == 0 {
            continue; }
        pData = Vec_IntEntryP( vMap, Vec_IntEntry(vMap, i) );
        for  iFan = 0; iFan < 4; iFan++  {
            if pData[iFan+1] >= 0 {
                Vec_IntAddToEntry( vRefs, pData[iFan+1], 1 ); }
        }
    }
    }
    return vRefs;
}

/**Function*************************************************************

  Synopsis    [Create manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_Bmc3ManStart(pAig *Aig_Man_t, nTimeOutOne int32, nConfLimit int32, fUseSatoko int32, fUseGlucose int32) *Gia_ManBmc_t {
    
    var pObj *Aig_Obj_t;
    var i int32;
//    assert( Aig_ManRegNum(pAig) > 0 );
    // inlining of macro ABC_CALLOC
    p := &Gia_ManBmc_t{};
    p.pAig = pAig;
    // create mapping
    p.vMapping = Cnf_DeriveMappingArray( pAig );
    p.vMapRefs = Saig_ManBmcComputeMappingRefs( pAig, p.vMapping );
    // create sections
//    p->vSects = Saig_ManBmcSections( pAig );
    // map object IDs into their numbers and section numbers
    p.nObjNums = 0;
    p.vId2Num  = Vec_IntStartFull( Aig_ManObjNumMax(pAig) );
    Vec_IntWriteEntry( p.vId2Num,  Aig_ObjId(Aig_ManConst1(pAig)), p.nObjNums++ );
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        Vec_IntWriteEntry( p.vId2Num,  Aig_ObjId(pObj), p.nObjNums++ );
    }
    // inlining of macro Aig_ManForEachNode
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else
        if Vec_IntEntry(p.vMapping, Aig_ObjId(pObj)) > 0 {
            Vec_IntWriteEntry( p.vId2Num,  Aig_ObjId(pObj), p.nObjNums++ ); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Vec_IntWriteEntry( p.vId2Num,  Aig_ObjId(pObj), p.nObjNums++ );
    }
    p.vId2Var  = Vec_PtrAlloc( 100 );
    p.vTerInfo = Vec_PtrAlloc( 100 );
    p.vVisited = Vec_WecAlloc( 100 );
    // create solver
    p.nSatVars = 1;
    if fUseSatoko != 0 {
        var opts satoko_opts_t;
        satoko_default_opts(&opts);
        opts.conf_limit = nConfLimit;
        p.pSat2 = satoko_create();  
        satoko_configure(p.pSat2, &opts);
        satoko_setnvars(p.pSat2, 1000);
    
    } else if fUseGlucose != 0 {
        //opts.conf_limit = nConfLimit;
        p.pSat3 = bmcg_sat_solver_start();  
        for  i = 0; i < 1000; i++  {
            bmcg_sat_solver_addvar( p.pSat3 );
        }
    
    } else {
        p.pSat  = sat_solver_new();
        sat_solver_setnvars(p.pSat, 1000);
    }
    Cnf_ReadMsops( &p.pSopSizes, &p.pSops );
    // terminary simulation 
    p.nWordNum = Abc_BitWordNum( 2 * Aig_ManObjNumMax(pAig) );
    // hash table
    p.vData = Vec_IntAlloc( 5 * 10000 );
    p.vHash = Hsh_IntManStart( p.vData, 5, 10000 );
    p.vId2Lit = Vec_IntAlloc( 10000 );
    // time spent on each outputs
    if nTimeOutOne {
        // inlining of macro ABC_ALLOC
        p.pTime4Outs = ((abctime *) malloc(sizeof(abctime) * (size_t)(Saig_ManPoNum(pAig))));
        for  i = 0; i < Saig_ManPoNum(pAig); i++  {
            p.pTime4Outs[i] = nTimeOutOne * int64(time.Second) / 1000 + 1;
        }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Delete manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_Bmc3ManStop(p *Gia_ManBmc_t) {
    if p.pPars.fVerbose {
        var nUsedVars int32;
        if p.pSat {
            nUsedVars = sat_solver_count_usedvars(p.pSat);
        } else {
            nUsedVars = 0;
        }
        Abc_Print( 1, "LStart(P) = %d  LDelta(Q) = %d  LRatio(R) = %d  ReduceDB = %d  Vars = %d  Used = %d (%.2f %%)\n", 
            p.pSat ? p.pSat.nLearntStart     : 0, 
            p.pSat ? p.pSat.nLearntDelta     : 0, 
            p.pSat ? p.pSat.nLearntRatio     : 0, 
            p.pSat ? p.pSat.nDBreduces       : 0, 
            p.pSat ? sat_solver_nvars(p.pSat) : p.pSat3 ? bmcg_sat_solver_varnum(p.pSat3) : satoko_varnum(p.pSat2), 
            nUsedVars, 
            100.0*nUsedVars/(p.pSat ? sat_solver_nvars(p.pSat) : p.pSat3 ? bmcg_sat_solver_varnum(p.pSat3) : satoko_varnum(p.pSat2)) );
        Abc_Print( 1, "Buffs = %d. Dups = %d.   Hash hits = %d.  Hash misses = %d.  UniProps = %d.\n", 
            p.nBufNum, p.nDupNum, p.nHashHit, p.nHashMiss, p.nUniProps );
    }
//    Aig_ManCleanMarkA( p->pAig );
    if p.vCexes {
        assert(p.pAig.vSeqModelVec == nil, "p.pAig.vSeqModelVec == nil");
        p.pAig.vSeqModelVec = p.vCexes;
        p.vCexes = nil;
    }
//    Vec_PtrFreeFree( p->vCexes );
    Vec_WecFree( p.vVisited );
    Vec_IntFree( p.vMapping );
    Vec_IntFree( p.vMapRefs );
//    Vec_VecFree( p->vSects );
    Vec_IntFree( p.vId2Num );
    Vec_VecFree( (Vec_Vec_t *)p.vId2Var );
    Vec_PtrFreeFree( p.vTerInfo );
    if p.pSat {  sat_solver_delete( p.pSat ); }
    if p.pSat2 { satoko_destroy( p.pSat2 ); }
    if p.pSat3 { bmcg_sat_solver_stop( p.pSat3 ); }
    p.pTime4Outs = nil;
    Vec_IntFree( p.vData );
    Hsh_IntManStop( p.vHash );
    Vec_IntFree( p.vId2Lit );
    p.pSopSizes = nil;
    p.pSops[1] = nil;
    p.pSops = nil;
    p = nil;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcMapping(p *Gia_ManBmc_t, pObj *Aig_Obj_t) *int32 {
    if Vec_IntEntry(p.vMapping, Aig_ObjId(pObj)) == 0 {
        return nil; }
    return Vec_IntEntryP( p.vMapping, Vec_IntEntry(p.vMapping, Aig_ObjId(pObj)) );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcLiteral(p *Gia_ManBmc_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var vFrame *Vec_Int_t;
    var ObjNum int32;
    assert(!Aig_ObjIsNode(pObj) || (Saig_ManBmcMapping(p, pObj) != nil), "!Aig_ObjIsNode(pObj) || Saig_ManBmcMapping(p, pObj)");
    ObjNum  = Vec_IntEntry( p.vId2Num, Aig_ObjId(pObj) );
    assert(ObjNum >= 0, "ObjNum >= 0");
    vFrame  = (Vec_Int_t *)Vec_PtrEntry( p.vId2Var, iFrame );
    assert(vFrame != nil, "vFrame != nil");
    return Vec_IntEntry( vFrame, ObjNum );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcSetLiteral(p *Gia_ManBmc_t, pObj *Aig_Obj_t, iFrame int32, iLit int32) int32 {
    var vFrame *Vec_Int_t;
    var ObjNum int32;
    assert(!Aig_ObjIsNode(pObj) || (Saig_ManBmcMapping(p, pObj) != nil), "!Aig_ObjIsNode(pObj) || Saig_ManBmcMapping(p, pObj)");
    ObjNum  = Vec_IntEntry( p.vId2Num, Aig_ObjId(pObj) );
    vFrame  = (Vec_Int_t *)Vec_PtrEntry( p.vId2Var, iFrame );
    Vec_IntWriteEntry( vFrame, ObjNum, iLit );
/*
    if ( Vec_IntEntry( p->vMapRefs, Aig_ObjId(pObj) ) > 1 )
        p->nLitUsed++;
    else
        p->nLitUseless++;
*/
    return iLit;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Saig_ManBmcCof0_s_Truth [4]int32 = [4]int32{ 0xAAAA, 0xCCCC, 0xF0F0, 0xFF00 };
func Saig_ManBmcCof0(t int32, v int32) int32 {
    
    return 0xffff & ((t & ~spStatic_Saig_ManBmcCof0_s_Truth[v]) | ((t & ~spStatic_Saig_ManBmcCof0_s_Truth[v]) << (1<<v)));
}
var spStatic_Saig_ManBmcCof1_s_Truth [4]int32 = [4]int32{ 0xAAAA, 0xCCCC, 0xF0F0, 0xFF00 };
func Saig_ManBmcCof1(t int32, v int32) int32 {
    
    return 0xffff & ((t & spStatic_Saig_ManBmcCof1_s_Truth[v]) | ((t & spStatic_Saig_ManBmcCof1_s_Truth[v]) >> (1<<v)));
}
func Saig_ManBmcCofEqual(t int32, v int32) int32 {
    assert(v >= 0 && v <= 3, "v >= 0 && v <= 3");
    if v == 0 {
        return ((t & 0xAAAA) >> 1) == (t & 0x5555); }
    if v == 1 {
        return ((t & 0xCCCC) >> 2) == (t & 0x3333); }
    if v == 2 {
        return ((t & 0xF0F0) >> 4) == (t & 0x0F0F); }
    if v == 3 {
        return ((t & 0xFF00) >> 8) == (t & 0x00FF); }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Derives CNF for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcReduceTruth(uTruth int32, Lits []int32) int32 {
    var v int32;
    for  v = 0; v < 4; v++  {
        if Lits[v] == 0 {
            uTruth = Saig_ManBmcCof0(uTruth, v);
            Lits[v] = -1;
        
        } else if Lits[v] == 1 {
            uTruth = Saig_ManBmcCof1(uTruth, v);
            Lits[v] = -1;
        }
    }
    for  v = 0; v < 4; v++  {
        if Lits[v] == -1 {
            assert((Saig_ManBmcCofEqual(uTruth, v) != 0), "Saig_ManBmcCofEqual(uTruth, v)");
        } else if Saig_ManBmcCofEqual(uTruth, v) != 0 {
            Lits[v] = -1; }
    }
    return uTruth;
}

/**Function*************************************************************

  Synopsis    [Derives CNF for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcAddClauses(p *Gia_ManBmc_t, uTruth int32, Lits []int32, iLitOut int32) {
    var i int32;
    var k int32;
    var b int32;
    var CutLit int32;
    var nClaLits int32;
    var ClaLits [5]int32;
    assert(uTruth > 0 && uTruth < 0xffff, "uTruth > 0 && uTruth < 0xffff");
    // write positive/negative polarity
    for  i = 0; i < 2; i++  {
        if i != 0 {
            uTruth = 0xffff & ~uTruth; }
//        Extra_PrintBinary( stdout, &uTruth, 16 ); Abc_Print( 1, "\n" );
        for  k = 0; k < p.pSopSizes[uTruth]; k++  {
            nClaLits = 0;
            ClaLits[nClaLits++] = i ? lit_neg(iLitOut) : iLitOut;
            CutLit = p.pSops[uTruth][k];
            for  b = 3; b >= 0; b--  {
                if CutLit % 3 == 0 { // value 0 --> write positive literal
                
                    assert(Lits[b] > 1, "Lits[b] > 1");
                    ClaLits[nClaLits++] = Lits[b];
                
                } else if CutLit % 3 == 1 { // value 1 --> write negative literal
                
                    assert(Lits[b] > 1, "Lits[b] > 1");
                    ClaLits[nClaLits++] = lit_neg(Lits[b]);
                }
                CutLit = CutLit / 3;
            }
            if p.pSat2 {
                if !satoko_add_clause( p.pSat2, ClaLits, nClaLits ) {
                    assert(false, "0"); }
            
            } else if p.pSat3 {
                if !bmcg_sat_solver_addclause( p.pSat3, ClaLits, nClaLits ) {
                    assert(false, "0"); }
            
            } else {
                if !sat_solver_addclause( p.pSat, ClaLits, ClaLits+nClaLits ) {
                    assert(false, "0"); }
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Derives CNF for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcCreateCnf_rec(p *Gia_ManBmc_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    extern unsigned Dar_CutSortVars( unsigned uTruth, int * pVars );
    var pMapping *int32;
    var i int32;
    var iLit int32;
    var Lits [5]int32;
    var uTruth int32;
    iLit = Saig_ManBmcLiteral( p, pObj, iFrame );
    if iLit != ~0 {
        return iLit; } 
    assert(iFrame >= 0, "iFrame >= 0");
    if Aig_ObjIsCi(pObj) {
        if Saig_ObjIsPi(p.pAig, pObj) {
            iLit = toLit( p.nSatVars++ );
        } else {
            iLit = Saig_ManBmcCreateCnf_rec( p, Saig_ObjLoToLi(p.pAig, pObj), iFrame-1 ); }
        return Saig_ManBmcSetLiteral( p, pObj, iFrame, iLit );
    }
    if Aig_ObjIsCo(pObj) {
        iLit = Saig_ManBmcCreateCnf_rec( p, Aig_ObjFanin0(pObj), iFrame );
        if Aig_ObjFaninC0(pObj) {
            iLit = lit_neg(iLit); }
        return Saig_ManBmcSetLiteral( p, pObj, iFrame, iLit );
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    pMapping = Saig_ManBmcMapping( p, pObj );
    for  i = 0; i < 4; i++  {
        if pMapping[i+1] == -1 {
            Lits[i] = -1;
        } else {
            Lits[i] = Saig_ManBmcCreateCnf_rec( p, Aig_ManObj(p.pAig, pMapping[i+1]), iFrame ); }
    }
    uTruth = 0xffff & uint32(pMapping[0]);
    // propagate constants
    uTruth = Saig_ManBmcReduceTruth( uTruth, Lits );
    if uTruth == 0 || uTruth == 0xffff {
        iLit = (uTruth == 0xffff);
        return Saig_ManBmcSetLiteral( p, pObj, iFrame, iLit );
    }
    // canonicize inputs
    uTruth = Dar_CutSortVars( uTruth, Lits );
    assert(uTruth != 0 && uTruth != 0xffff, "uTruth != 0 && uTruth != 0xffff");
    if uTruth == 0xAAAA || uTruth == 0x5555 {
        iLit = Abc_LitNotCond( Lits[0], uTruth == 0x5555 );
        p.nBufNum++;
    
    } else {
        var iEntry, iRes int32;
        var fCompl int32  = (uTruth & 1);
        Lits[4] = (uTruth & 1) ? 0xffff & ~uTruth : uTruth;
        iEntry = Vec_IntSize(p.vData) / 5;
        assert(iEntry * 5 == Vec_IntSize(p.vData), "iEntry * 5 == Vec_IntSize(p.vData)");
        for  i = 0; i < 5; i++  {
            Vec_IntPush( p.vData, Lits[i] );
        }
        iRes = Hsh_IntManAdd( p.vHash, iEntry );
        if iRes == iEntry {
            iLit = toLit( p.nSatVars++ );
            Saig_ManBmcAddClauses( p, Lits[4], Lits, iLit );
            assert(iEntry == Vec_IntSize(p.vId2Lit), "iEntry == Vec_IntSize(p.vId2Lit)");
            Vec_IntPush( p.vId2Lit, iLit );
            p.nHashMiss++;
        
        } else {
            iLit = Vec_IntEntry( p.vId2Lit, iRes );
            Vec_IntShrink( p.vData, Vec_IntSize(p.vData) - 5 );
            p.nHashHit++;
        }
        iLit = Abc_LitNotCond( iLit, fCompl );
    }
    return Saig_ManBmcSetLiteral( p, pObj, iFrame, iLit );
}


/**Function*************************************************************

  Synopsis    [Derives CNF for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcCreateCnf_iter(p *Gia_ManBmc_t, pObj *Aig_Obj_t, iFrame int32, vVisit *Vec_Int_t) {
    if Saig_ManBmcLiteral( p, pObj, iFrame ) != ~0 {
        return; } 
    if Aig_ObjIsTravIdCurrent(p.pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(p.pAig, pObj);
    if Aig_ObjIsCi(pObj) {
        if Saig_ObjIsLo(p.pAig, pObj) {
            Vec_IntPush( vVisit, Saig_ObjLoToLi(p.pAig, pObj).Id ); }
        return;
    }
    if Aig_ObjIsCo(pObj) {
        Saig_ManBmcCreateCnf_iter( p, Aig_ObjFanin0(pObj), iFrame, vVisit );
        return;
    
    } else {
        var pMapping *int32;
        var i int32;
        assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
        pMapping = Saig_ManBmcMapping( p, pObj );
        for  i = 0; i < 4; i++  {
            if pMapping[i+1] != -1 {
                Saig_ManBmcCreateCnf_iter( p, Aig_ManObj(p.pAig, pMapping[i+1]), iFrame, vVisit ); }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Recursively performs terminary simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcRunTerSim_rec(p *Gia_ManBmc_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var pInfo *uint32  = (unsigned *)Vec_PtrEntry( p.vTerInfo, iFrame );
    var Val0 int32;
    var Val1 int32;
    var Value int32 = Saig_ManBmcSimInfoGet( pInfo, pObj );
    if Value != SAIG_TER_NON {
/*
        // check the value of this literal in the SAT solver
        if ( Value == SAIG_TER_UND && Saig_ManBmcMapping(p, pObj) )
        {
            int Lit = Saig_ManBmcLiteral( p, pObj, iFrame );
            if ( Lit >= 0  )
            {
                assert( Lit >= 2 );
                if ( p->pSat->assigns[Abc_Lit2Var(Lit)] < 2 )
                {
                    p->nUniProps++;
                    if ( Abc_LitIsCompl(Lit) ^ (p->pSat->assigns[Abc_Lit2Var(Lit)] == 0) )
                        Value = SAIG_TER_ONE;
                    else
                        Value = SAIG_TER_ZER;

//                    Value = SAIG_TER_UND;  // disable!

                    // use the new value
                    Saig_ManBmcSimInfoSet( pInfo, pObj, Value );
                    // transfer to the unrolling
                    if ( Value != SAIG_TER_UND )
                        Saig_ManBmcSetLiteral( p, pObj, iFrame, (int)(Value == SAIG_TER_ONE) );
                }
            }
        }
*/
        return Value;
    }
    if Aig_ObjIsCo(pObj) {
        Value = Saig_ManBmcRunTerSim_rec( p, Aig_ObjFanin0(pObj), iFrame );
        if Aig_ObjFaninC0(pObj) {
            Value = Saig_ManBmcSimInfoNot( Value ); }
    
    } else if Saig_ObjIsLo(p.pAig, pObj) {
        assert(iFrame > 0, "iFrame > 0");
        Value = Saig_ManBmcRunTerSim_rec( p, Saig_ObjLoToLi(p.pAig, pObj), iFrame - 1 );
    
    } else if Aig_ObjIsNode(pObj) {
        Val0 = Saig_ManBmcRunTerSim_rec( p, Aig_ObjFanin0(pObj), iFrame  );
        Val1 = Saig_ManBmcRunTerSim_rec( p, Aig_ObjFanin1(pObj), iFrame  );
        if Aig_ObjFaninC0(pObj) {
            Val0 = Saig_ManBmcSimInfoNot( Val0 ); }
        if Aig_ObjFaninC1(pObj) {
            Val1 = Saig_ManBmcSimInfoNot( Val1 ); }
        Value = Saig_ManBmcSimInfoAnd( Val0, Val1 );
    
    } else { assert(false, "0"); }
    Saig_ManBmcSimInfoSet( pInfo, pObj, Value );
    // transfer to the unrolling
    if (Saig_ManBmcMapping(p, pObj) != nil) && Value != SAIG_TER_UND {
        Saig_ManBmcSetLiteral( p, pObj, iFrame, int32((Value == SAIG_TER_ONE)) ); }
    return Value;
}

/**Function*************************************************************

  Synopsis    [Derives CNF for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcCreateCnf(p *Gia_ManBmc_t, pObj *Aig_Obj_t, iFrame int32) int32 {
    var vVisit, vVisit2 *Vec_Int_t;
    var pTemp *Aig_Obj_t;
    var Lit, f, i int32;
    // perform ternary simulation
    var Value int32  = Saig_ManBmcRunTerSim_rec( p, pObj, iFrame );
    if Value != SAIG_TER_UND {
        return int32((Value == SAIG_TER_ONE)); }
    // construct CNF if value is ternary
//    Lit = Saig_ManBmcCreateCnf_rec( p, pObj, iFrame );
    Vec_WecClear( p.vVisited );
    vVisit = Vec_WecPushLevel( p.vVisited );
    Vec_IntPush( vVisit, Aig_ObjId(pObj) );
    for  f = iFrame; f >= 0; f--  {
        Aig_ManIncrementTravId( p.pAig );
        vVisit2 = Vec_WecPushLevel( p.vVisited );
        vVisit = Vec_WecEntry( p.vVisited, Vec_WecSize(p.vVisited)-2 );
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vVisit) && (((pTemp) = Aig_ManObj(p.pAig, Vec_IntEntry(vVisit,i))), 1); i++  {
            Saig_ManBmcCreateCnf_iter( p, pTemp, f, vVisit2 );
        }
        if Vec_IntSize(vVisit2) == 0 {
            break; }
    }
    // inlining of macro Vec_WecForEachLevelReverse
    for  f = Vec_WecSize(p.vVisited)-1; (f >= 0) && (((vVisit) = Vec_WecEntry(p.vVisited, f)), 1); f--  {
        // inlining of macro Aig_ManForEachObjVec
        for  i = 0; i < Vec_IntSize(vVisit) && (((pTemp) = Aig_ManObj(p.pAig, Vec_IntEntry(vVisit,i))), 1); i++  {
            Saig_ManBmcCreateCnf_rec( p, pTemp, iFrame-f );
        }
    }
    Lit = Saig_ManBmcLiteral( p, pObj, iFrame );
    // extend the SAT solver
    if p.pSat2 {
        satoko_setnvars( p.pSat2, p.nSatVars );
    } else if p.pSat3 {
        for  i = bmcg_sat_solver_varnum(p.pSat3); i < p.nSatVars; i++  {
            bmcg_sat_solver_addvar( p.pSat3 );
        }
    
    } else {
        sat_solver_setnvars( p.pSat, p.nSatVars ); }
    return Lit;
}



/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in decreasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Aig_NodeCompareRefsIncrease(pp1 **Aig_Obj_t, pp2 **Aig_Obj_t) int32 {
    var Diff int32  = Aig_ObjRefs(*pp1) - Aig_ObjRefs(*pp2);
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    Diff = Aig_ObjId(*pp1) - Aig_ObjId(*pp2);
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ParBmcSetDefaultParams(p *Saig_ParBmc_t) {
    memset( p, 0, sizeof(Saig_ParBmc_t) );
    p.nStart         =     0;    // maximum number of timeframes 
    p.nFramesMax     =     0;    // maximum number of timeframes 
    p.nConfLimit     =     0;    // maximum number of conflicts at a node
    p.nConfLimitJump =     0;    // maximum number of conflicts after jumping
    p.nFramesJump    =     0;    // the number of tiemframes to jump
    p.nTimeOut       =     0;    // approximate timeout in seconds
    p.nTimeOutGap    =     0;    // time since the last CEX found
    p.nPisAbstract   =     0;    // the number of PIs to abstract
    p.fSolveAll      =     0;    // stops on the first SAT instance
    p.fDropSatOuts   =     0;    // replace sat outputs by constant 0
    p.nLearnedStart  = 10000;    // starting learned clause limit
    p.nLearnedDelta  =  2000;    // delta of learned clause limit
    p.nLearnedPerce  =    80;    // ratio of learned clause limit
    p.fVerbose       =     0;    // verbose 
    p.fNotVerbose    =     0;    // skip line-by-line print-out 
    p.iFrame         =    -1;    // explored up to this frame
    p.nFailOuts      =     0;    // the number of failed outputs
    p.nDropOuts      =     0;    // the number of timed out outputs
    p.timeLastSolved =     0;    // time when the last one was solved
}

/**Function*************************************************************

  Synopsis    [Returns time to stop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcTimeToStop(pPars *Saig_ParBmc_t, nTimeToStopNG abctime) abctime {
    var nTimeToStopGap abctime;
    if pPars.nTimeOutGap {
        nTimeToStopGap = pPars.nTimeOutGap * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStopGap = 0;
    }
    var nTimeToStop abctime;
    if nTimeToStopNG && nTimeToStopGap {
        if nTimeToStopNG < nTimeToStopGap {
            nTimeToStop = nTimeToStopNG;
        } else {
            nTimeToStop = nTimeToStopGap;
        }
    } else if nTimeToStopNG {
        nTimeToStop = nTimeToStopNG;
    } else if nTimeToStopGap {
        nTimeToStop = nTimeToStopGap; }
    return nTimeToStop;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManGenerateCex(p *Gia_ManBmc_t, f int32, i int32) *Abc_Cex_t {
    var pObjPi *Aig_Obj_t;
    var pCex *Abc_Cex_t  = Abc_CexMakeTriv( Aig_ManRegNum(p.pAig), Saig_ManPiNum(p.pAig), Saig_ManPoNum(p.pAig), f*Saig_ManPoNum(p.pAig)+i );
    var j int32;
    var k int32;
    var iBit int32 = Saig_ManRegNum(p.pAig);
    for  j = 0; j <= f; j++, iBit += Saig_ManPiNum(p.pAig)  {
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  k = 0; (k < Saig_ManPiNum(p.pAig)) && (((pObjPi) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, k)), 1); k++  {
            var iLit int32  = Saig_ManBmcLiteral( p, pObjPi, j );
            if p.pSat2 {
                if iLit != ~0 && satoko_read_cex_varvalue(p.pSat2, lit_var(iLit)) {
                    Abc_InfoSetBit( pCex.pData, iBit + k ); }
            
            } else if p.pSat3 {
                if iLit != ~0 && bmcg_sat_solver_read_cex_varvalue(p.pSat3, lit_var(iLit)) {
                    Abc_InfoSetBit( pCex.pData, iBit + k ); }
            
            } else {
                if iLit != ~0 && sat_solver_var_value(p.pSat, lit_var(iLit)) {
                    Abc_InfoSetBit( pCex.pData, iBit + k ); }
            }
        }
    }
    return pCex;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManCallSolver(p *Gia_ManBmc_t, Lit int32) int32 {
    if Lit == 0 {
        return l_False; }
    if Lit == 1 {
        return l_True; }
    if p.pSat2 {
        return satoko_solve_assumptions_limit( p.pSat2, &Lit, 1, p.pPars.nConfLimit );
    } else if p.pSat3 {
        bmcg_sat_solver_set_conflict_budget( p.pSat3, p.pPars.nConfLimit );
        return bmcg_sat_solver_solve( p.pSat3, &Lit, 1 );
    
    } else {
        return sat_solver_solve( p.pSat, &Lit, &Lit + 1, (ABC_INT64_T)p.pPars.nConfLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 ); }
}

/**Function*************************************************************

  Synopsis    [Bounded model checking engine.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Saig_ManBmcScalable(pAig *Aig_Man_t, pPars *Saig_ParBmc_t) int32 {
    var p *Gia_ManBmc_t;
    var pObj *Aig_Obj_t;
    var pCexNew, pCexNew0 *Abc_Cex_t;
    var pLogFile *os.File  = nil;
    var pInfo *uint32;
    var RetValue int32 = -1;
    var fFirst int32 = 1;
    var nJumpFrame int32;
    var fUnfinished int32;
    var nOutDigits int32  = Abc_Base10Log( Saig_ManPoNum(pAig) );
    var i, f, k, Lit, status int32;
    var clk abctime;
    var clk2 abctime;
    var clkSatRun abctime;
    var clkOther abctime;
    var clkTotal abctime = Abc_Clock();
    var nTimeUnsat, nTimeSat, nTimeUndec, clkOne abctime;
    var nTimeToStopNG, nTimeToStop abctime;
    if pPars.pLogFileName {
        pLogFile = os.OpenFile(pPars.pLogFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666); }
    if pPars.nTimeOutOne && pPars.nTimeOut == 0 {
        pPars.nTimeOut = pPars.nTimeOutOne * Saig_ManPoNum(pAig) / 1000 + 1; }
    if pPars.nTimeOutOne && !pPars.fSolveAll {
        pPars.nTimeOutOne = 0; }
    if pPars.nTimeOut {
        nTimeToStopNG = pPars.nTimeOut * int64(time.Second) + Abc_Clock();
    } else {
        nTimeToStopNG = 0;
    }
    nTimeToStop   = Saig_ManBmcTimeToStop( pPars, nTimeToStopNG );
    // create BMC manager
    p = Saig_Bmc3ManStart( pAig, pPars.nTimeOutOne, pPars.nConfLimit, pPars.fUseSatoko, pPars.fUseGlucose );
    p.pPars = pPars;
    if p.pSat {
        p.pSat.nLearntStart = p.pPars.nLearnedStart;
        p.pSat.nLearntDelta = p.pPars.nLearnedDelta;
        p.pSat.nLearntRatio = p.pPars.nLearnedPerce;
        p.pSat.nLearntMax   = p.pSat.nLearntStart;
        p.pSat.fNoRestarts  = p.pPars.fNoRestarts;
        p.pSat.RunId        = p.pPars.RunId;
        p.pSat.pFuncStop    = p.pPars.pFuncStop;
    
    } else if p.pSat3 {
//        satoko_set_runid(p->pSat3, p->pPars->RunId);
//        satoko_set_stop_func(p->pSat3, p->pPars->pFuncStop);
    
    } else {
        satoko_set_runid(p.pSat2, p.pPars.RunId);
        satoko_set_stop_func(p.pSat2, p.pPars.pFuncStop);
    }
    if pPars.fSolveAll && p.vCexes == nil {
        p.vCexes = Vec_PtrStart( Saig_ManPoNum(pAig) ); }
    if pPars.fVerbose {
        Abc_Print( 1, "Running \"bmc3\". PI/PO/Reg = %d/%d/%d. And =%7d. Lev =%6d. ObjNums =%6d.\n",// Sect =%3d.\n", 
            Saig_ManPiNum(pAig), Saig_ManPoNum(pAig), Saig_ManRegNum(pAig),
            Aig_ManNodeNum(pAig), Aig_ManLevelNum(pAig), p.nObjNums );//, Vec_VecSize(p->vSects) );
        Abc_Print( 1, "Params: FramesMax = %d. Start = %d. ConfLimit = %d. TimeOut = %d. SolveAll = %d.\n", 
            pPars.nFramesMax, pPars.nStart, pPars.nConfLimit, pPars.nTimeOut, pPars.fSolveAll );
    } 
    if pPars.nFramesMax {
        pPars.nFramesMax = pPars.nFramesMax;
    } else {
        pPars.nFramesMax = ABC_INFINITY;
    }
    // set runtime limit
    if nTimeToStop {
        if p.pSat2 {
            satoko_set_runtime_limit( p.pSat2, nTimeToStop );
        } else if p.pSat3 {
            bmcg_sat_solver_set_runtime_limit( p.pSat3, nTimeToStop );
        } else {
            sat_solver_set_runtime_limit( p.pSat, nTimeToStop ); }
    }
    // perform frames
    Aig_ManRandom( 1 );
    pPars.timeLastSolved = Abc_Clock();
    for  f = 0; f < pPars.nFramesMax; f++  {
        // stop BMC after exploring all reachable states
        if !pPars.nFramesJump && Aig_ManRegNum(pAig) < 30 && f == (1 << Aig_ManRegNum(pAig)) {
            Abc_Print( 1, "Stopping BMC because all 2^%d reachable states are visited.\n", Aig_ManRegNum(pAig) );
            if p.pPars.fUseBridge {
                // inlining of macro Saig_ManForEachPo
                // inlining of macro Vec_PtrForEachEntryStop
                for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
                    if !(p.vCexes && Vec_PtrEntry(p.vCexes, i)) && !(p.pTime4Outs && p.pTime4Outs[i] == 0) { // not SAT and not timed out
                        Gia_ManToBridgeResult( os.Stdout, 1, nil, i ); }
                } }
            if pPars.nFailOuts {
                RetValue = 0;
            } else {
                RetValue = 1;
            }
            goto finish;
        }
        // stop BMC if all targets are solved
        if pPars.fSolveAll && pPars.nFailOuts + pPars.nDropOuts >= Saig_ManPoNum(pAig) {
            Abc_Print( 1, "Stopping BMC because all targets are disproved or timed out.\n" );
            if pPars.nFailOuts {
                RetValue = 0;
            } else {
                RetValue = 1;
            }
            goto finish;
        }
        // consider the next timeframe
        if (RetValue == -1 || pPars.fSolveAll) && pPars.nStart == 0 && !(nJumpFrame != 0) {
            pPars.iFrame = f-1; }
        // map nodes of this section
        Vec_PtrPush( p.vId2Var, Vec_IntStartFull(p.nObjNums) );
        // inlining of macro ABC_CALLOC
        Vec_PtrPush( p.vTerInfo, (pInfo = ((unsigned *) calloc((size_t)(p.nWordNum), sizeof(unsigned)))) );
/*
        // cannot remove mapping of frame values for any timeframes
        // because with constant propagation they may be needed arbitrarily far
        if ( f > 2*Vec_VecSize(p->vSects) )
        {
            int iFrameOld = f - 2*Vec_VecSize( p->vSects );
            void * pMemory = Vec_IntReleaseArray( Vec_PtrEntry(p->vId2Var, iFrameOld) );
            ABC_FREE( pMemory );
        } 
*/
        // prepare some nodes
        Saig_ManBmcSetLiteral( p, Aig_ManConst1(pAig), f, 1 );
        Saig_ManBmcSimInfoSet( pInfo, Aig_ManConst1(pAig), SAIG_TER_ONE );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
            Saig_ManBmcSimInfoSet( pInfo, pObj, SAIG_TER_UND );
        }
        if f == 0 {
            // inlining of macro Saig_ManForEachLo
            for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
                Saig_ManBmcSetLiteral( p, pObj, 0, 0 );
                Saig_ManBmcSimInfoSet( pInfo, pObj, SAIG_TER_ZER );
            }
        }
        if (pPars.nStart && f < pPars.nStart) || ((nJumpFrame != 0) && f < nJumpFrame) {
            continue; }
        // create CNF upfront
        if pPars.fSolveAll {
            // inlining of macro Saig_ManForEachPo
            // inlining of macro Vec_PtrForEachEntryStop
            for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
                if i >= Saig_ManPoNum(pAig) {
                    break; }
                // check for timeout
                if pPars.nTimeOutGap && pPars.timeLastSolved && Abc_Clock() > pPars.timeLastSolved + pPars.nTimeOutGap * int64(time.Second) {
                    Abc_Print( 1, "Reached gap timeout (%d seconds).\n",  pPars.nTimeOutGap );
                    goto finish;
                }
                if nTimeToStop && Abc_Clock() > nTimeToStop {
                    if !pPars.fSilent {
                        Abc_Print( 1, "Reached timeout (%d seconds).\n",  pPars.nTimeOut ); }
                    goto finish;
                }
                // skip solved outputs
                if p.vCexes && Vec_PtrEntry(p.vCexes, i) {
                    continue; }
                // skip output whose time has run out
                if p.pTime4Outs && p.pTime4Outs[i] == 0 {
                    continue; }
                // add constraints for this output
clk2 = Abc_Clock();
                Saig_ManBmcCreateCnf( p, pObj, f );
clkOther += Abc_Clock() - clk2;
            }
        }
        // solve SAT
        clk = Abc_Clock(); 
        // inlining of macro Saig_ManForEachPo
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
            if i >= Saig_ManPoNum(pAig) {
                break; }
            // check for timeout
            if pPars.nTimeOutGap && pPars.timeLastSolved && Abc_Clock() > pPars.timeLastSolved + pPars.nTimeOutGap * int64(time.Second) {
                Abc_Print( 1, "Reached gap timeout (%d seconds).\n",  pPars.nTimeOutGap );
                goto finish;
            }
            if nTimeToStop && Abc_Clock() > nTimeToStop {
                if !pPars.fSilent {
                    Abc_Print( 1, "Reached timeout (%d seconds).\n",  pPars.nTimeOut ); }
                goto finish;
            }
            if p.pPars.pFuncStop && p.pPars.pFuncStop(p.pPars.RunId) {
                if !pPars.fSilent {
                    Abc_Print( 1, "Bmc3 got callbacks.\n" ); }
                goto finish;
            }
            // skip solved outputs
            if p.vCexes && Vec_PtrEntry(p.vCexes, i) {
                continue; }
            // skip output whose time has run out
            if p.pTime4Outs && p.pTime4Outs[i] == 0 {
                continue; }
            // add constraints for this output
clk2 = Abc_Clock();
            Lit = Saig_ManBmcCreateCnf( p, pObj, f );
clkOther += Abc_Clock() - clk2;
            // solve this output
            fUnfinished = 0;
            if p.pSat { sat_solver_compress( p.pSat ); }
            if p.pTime4Outs {
                assert(p.pTime4Outs[i] > 0, "p.pTime4Outs[i] > 0");
                clkOne = Abc_Clock();
                if p.pSat2 {
                    satoko_set_runtime_limit( p.pSat2, p.pTime4Outs[i] + Abc_Clock() );
                } else if p.pSat3 {
                    bmcg_sat_solver_set_runtime_limit( p.pSat3, p.pTime4Outs[i] + Abc_Clock() );
                } else {
                    sat_solver_set_runtime_limit( p.pSat, p.pTime4Outs[i] + Abc_Clock() ); }
            }
clk2 = Abc_Clock();
            status = Saig_ManCallSolver( p, Lit );
clkSatRun = Abc_Clock() - clk2;
            if pLogFile != nil {
                fmt.Fprintf(pLogFile, "Frame %5d  Output %5d  Time(ms) %8d %8d\n", f, i, Lit < 2 ? 0 : int32((clkSatRun * 1000 / int64(time.Second))), Lit < 2 ? 0 : Abc_MaxInt(0, Abc_MinInt(pPars.nTimeOutOne, pPars.nTimeOutOne - int32(((p.pTime4Outs[i] - clkSatRun) * 1000 / int64(time.Second)))))); }
            if p.pTime4Outs {
                var timeSince abctime  = Abc_Clock() - clkOne;
                assert(p.pTime4Outs[i] > 0, "p.pTime4Outs[i] > 0");
                p.pTime4Outs[i] = (p.pTime4Outs[i] > timeSince) ? p.pTime4Outs[i] - timeSince : 0;
                if p.pTime4Outs[i] == 0 && status != l_True {
                    pPars.nDropOuts++; }
            }
            if status == l_False {
nTimeUnsat += clkSatRun;
                if Lit != 0 {
                    // add final unit clause
                    Lit = lit_neg( Lit );
                    if p.pSat2 {
                        status = satoko_add_clause( p.pSat2, &Lit, 1 );
                    } else if p.pSat3 {
                        status = bmcg_sat_solver_addclause( p.pSat3, &Lit, 1 );
                    } else {
                        status = sat_solver_addclause( p.pSat, &Lit, &Lit + 1 ); }
                    assert((status != 0), "status");
                    // add learned units
                    if p.pSat {
                        for  k = 0; k < veci_size(&p.pSat.unit_lits); k++  {
                            Lit = veci_begin(&p.pSat.unit_lits)[k];
                            status = sat_solver_addclause( p.pSat, &Lit, &Lit + 1 );
                            assert((status != 0), "status");
                        }
                        veci_resize(&p.pSat.unit_lits, 0);
                        // propagate units
                        sat_solver_compress( p.pSat );
                    }
                }
                if p.pPars.fUseBridge {
                    Gia_ManReportProgress( os.Stdout, i, f ); }
            
            } else if status == l_True {
nTimeSat += clkSatRun;
                RetValue = 0;
                fFirst = 0;
                if !pPars.fSolveAll {
                    if pPars.fVerbose {
                        Abc_Print( 1, "%4d %s : ", f,  fUnfinished ? "-" : "+" );
                        Abc_Print( 1, "Var =%8.0f. ",  float64(p.nSatVars) );
                        Abc_Print( 1, "Cla =%9.0f. ",  float64((p.pSat ? p.pSat.stats.clauses   : p.pSat3 ? bmcg_sat_solver_clausenum(p.pSat3)   : satoko_clausenum(p.pSat2))) );
                        Abc_Print( 1, "Conf =%7.0f. ", float64((p.pSat ? p.pSat.stats.conflicts : p.pSat3 ? bmcg_sat_solver_conflictnum(p.pSat3) : satoko_conflictnum(p.pSat2))) );
//                        Abc_Print( 1, "Imp =%10.0f. ", (double)p->pSat->stats.propagations );
//                        Abc_Print( 1, "Uni =%7.0f. ",(double)(p->pSat ? sat_solver_count_assigned(p->pSat) : 0) );
//                        ABC_PRT( "Time", Abc_Clock() - clk );
                        Abc_Print( 1, "Learn =%7.0f. ", float64((p.pSat ? p.pSat.stats.learnts : p.pSat3 ? bmcg_sat_solver_learntnum(p.pSat3) : satoko_learntnum(p.pSat2))) );
                        Abc_Print( 1, "%4.0f MB",      4.25*(f+1)*p.nObjNums /(1<<20) );
                        Abc_Print( 1, "%4.0f MB",      1.0*(p.pSat ? sat_solver_memory(p.pSat) : 0)/(1<<20) );
                        Abc_Print( 1, "%9.2f sec  ",   float32((Abc_Clock() - clkTotal))/float32((int64(time.Second))) );
//                        Abc_Print( 1, "\n" );
//                        ABC_PRMn( "Id2Var", (f+1)*p->nObjNums*4 );
//                        ABC_PRMn( "SAT", 42 * p->pSat->size + 16 * (int)p->pSat->stats.clauses + 4 * (int)p->pSat->stats.clauses_literals );
//                        Abc_Print( 1, "S =%6d. ", p->nBufNum );
//                        Abc_Print( 1, "D =%6d. ", p->nDupNum );
                        Abc_Print( 1, "\n" );
                        fflush( os.Stdout );
                    }
                    pAig.pSeqModel = nil;
                    pAig.pSeqModel = Saig_ManGenerateCex( p, f, i );
                    goto finish;
                }
                pPars.nFailOuts++;
                if !pPars.fNotVerbose {
                    Abc_Print( 1, "Output %*d was asserted in frame %2d (solved %*d out of %*d outputs).\n",  
                        nOutDigits, i, f, nOutDigits, pPars.nFailOuts, nOutDigits, Saig_ManPoNum(pAig) ); }
                if p.vCexes == nil {
                    p.vCexes = Vec_PtrStart( Saig_ManPoNum(pAig) ); }
                if p.pPars.fUseBridge || pPars.fStoreCex {
                    pCexNew = Saig_ManGenerateCex( p, f, i );
                } else {
                    pCexNew = (Abc_Cex_t *)(ABC_PTRINT_T)1;
                }
                pCexNew0 = nil;
                if p.pPars.fUseBridge {
                    Gia_ManToBridgeResult( os.Stdout, 0, pCexNew, pCexNew.iPo );
                    //Abc_CexFree( pCexNew );
                    pCexNew0 = pCexNew; 
                    pCexNew = (Abc_Cex_t *)(ABC_PTRINT_T)1;
                }
                Vec_PtrWriteEntry( p.vCexes, i, Abc_CexDup(pCexNew, Saig_ManRegNum(pAig)) ); 
                if pPars.pFuncOnFail && pPars.pFuncOnFail(i, pPars.fStoreCex ? (Abc_Cex_t *)Vec_PtrEntry(p.vCexes, i) : nil) {
                    Abc_CexFreeP( &pCexNew0 );
                    Abc_Print( 1, "Quitting due to callback on fail.\n" );
                    goto finish;
                }
                // reset the timeout
                pPars.timeLastSolved = Abc_Clock();
                nTimeToStop = Saig_ManBmcTimeToStop( pPars, nTimeToStopNG );
                if nTimeToStop {
                    if p.pSat2 {
                        satoko_set_runtime_limit( p.pSat2, nTimeToStop );
                    } else if p.pSat3 {
                        bmcg_sat_solver_set_runtime_limit( p.pSat3, nTimeToStop );
                    } else {
                        sat_solver_set_runtime_limit( p.pSat, nTimeToStop ); }
                }

                // check if other outputs failed under the same counter-example
                // inlining of macro Saig_ManForEachPo
                // inlining of macro Vec_PtrForEachEntryStop
                for  k = 0; (k < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, k)), 1); k++  {
                    var pCexDup *Abc_Cex_t;
                    if k >= Saig_ManPoNum(pAig) {
                        break; }
                    // skip solved outputs
                    if p.vCexes && Vec_PtrEntry(p.vCexes, k) {
                        continue; }
                    // check if this output is solved
                    Lit = Saig_ManBmcCreateCnf( p, pObj, f );
                    if p.pSat2 {
                        if satoko_read_cex_varvalue(p.pSat2, lit_var(Lit)) == Abc_LitIsCompl(Lit) {
                            continue; }
                    
                    } else if p.pSat3 {
                        if bmcg_sat_solver_read_cex_varvalue(p.pSat3, lit_var(Lit)) == Abc_LitIsCompl(Lit) {
                            continue; }
                    
                    } else {
                        if sat_solver_var_value(p.pSat, lit_var(Lit)) == Abc_LitIsCompl(Lit) {
                            continue; }
                    }
                    // write entry
                    pPars.nFailOuts++;
                    if !pPars.fNotVerbose {
                        Abc_Print( 1, "Output %*d was asserted in frame %2d (solved %*d out of %*d outputs).\n",  
                            nOutDigits, k, f, nOutDigits, pPars.nFailOuts, nOutDigits, Saig_ManPoNum(pAig) ); }
                    // report to the bridge
                    if p.pPars.fUseBridge {
                        // set the output number
                        pCexNew0.iPo = k;
                        Gia_ManToBridgeResult( os.Stdout, 0, pCexNew0, pCexNew0.iPo );
                    }
                    // remember solved output
                    //Vec_PtrWriteEntry( p->vCexes, k, Abc_CexDup(pCexNew, Saig_ManRegNum(pAig)) );
                    pCexDup = Abc_CexDup(pCexNew, Saig_ManRegNum(pAig));
                    pCexDup.iPo = k;
                    Vec_PtrWriteEntry( p.vCexes, k, pCexDup );
                }
                Abc_CexFreeP( &pCexNew0 );
                Abc_CexFree( pCexNew );
            
            } else {
nTimeUndec += clkSatRun;
                assert(status == l_Undef, "status == l_Undef");
                if pPars.nFramesJump {
                    pPars.nConfLimit = pPars.nConfLimitJump;
                    nJumpFrame = f + pPars.nFramesJump;
                    fUnfinished = 1;
                    break;
                }
                if p.pTime4Outs == nil {
                    goto finish; }
            }
        }
        if pPars.fVerbose {
            if fFirst == 1 && f > 0 && (p.pSat ? p.pSat.stats.conflicts : p.pSat3 ? bmcg_sat_solver_conflictnum(p.pSat3) : satoko_conflictnum(p.pSat2)) > 1 {
                fFirst = 0;
//                Abc_Print( 1, "Outputs of frames up to %d are trivially UNSAT.\n", f );
            }
            Abc_Print( 1, "%4d %s : ", f, fUnfinished ? "-" : "+" );
            Abc_Print( 1, "Var =%8.0f. ", float64(p.nSatVars) );
//            Abc_Print( 1, "Used =%8.0f. ", (double)sat_solver_count_usedvars(p->pSat) );
            Abc_Print( 1, "Cla =%9.0f. ", float64((p.pSat ? p.pSat.stats.clauses   : p.pSat3 ? bmcg_sat_solver_clausenum(p.pSat3)   : satoko_clausenum(p.pSat2)))   );
            Abc_Print( 1, "Conf =%7.0f. ",float64((p.pSat ? p.pSat.stats.conflicts : p.pSat3 ? bmcg_sat_solver_conflictnum(p.pSat3) : satoko_conflictnum(p.pSat2))) );
//            Abc_Print( 1, "Imp =%10.0f. ", (double)p->pSat->stats.propagations );
//            Abc_Print( 1, "Uni =%7.0f. ", (double)(p->pSat ? sat_solver_count_assigned(p->pSat) : 0) );
            Abc_Print( 1, "Learn =%7.0f. ", float64((p.pSat ? p.pSat.stats.learnts : p.pSat3 ? bmcg_sat_solver_learntnum(p.pSat3) : satoko_learntnum(p.pSat2))) );
            if pPars.fSolveAll {
                Abc_Print( 1, "CEX =%5d. ", pPars.nFailOuts ); }
            if pPars.nTimeOutOne {
                Abc_Print( 1, "T/O =%4d. ", pPars.nDropOuts ); }
//            ABC_PRT( "Time", Abc_Clock() - clk );
//            Abc_Print( 1, "%4.0f MB",     4.0*Vec_IntSize(p->vVisited) /(1<<20) );
            Abc_Print( 1, "%4.0f MB",     4.0*(f+1)*p.nObjNums /(1<<20) );
            Abc_Print( 1, "%4.0f MB",     1.0*(p.pSat ? sat_solver_memory(p.pSat) : 0)/(1<<20) );
//            Abc_Print( 1, " %6d %6d ",   p->nLitUsed, p->nLitUseless );
            Abc_Print( 1, "%9.2f sec ",   1.0*(Abc_Clock() - clkTotal)/int64(time.Second) );
//            Abc_Print( 1, "\n" );
//            ABC_PRMn( "Id2Var", (f+1)*p->nObjNums*4 );
//            ABC_PRMn( "SAT", 42 * p->pSat->size + 16 * (int)p->pSat->stats.clauses + 4 * (int)p->pSat->stats.clauses_literals );
//            Abc_Print( 1, "Simples = %6d. ", p->nBufNum );
//            Abc_Print( 1, "Dups = %6d. ", p->nDupNum );
            Abc_Print( 1, "\n" );
            fflush( os.Stdout );
        }
    }
    // consider the next timeframe
    if (nJumpFrame != 0) && pPars.nStart == 0 {
        pPars.iFrame = nJumpFrame - pPars.nFramesJump;
    } else if RetValue == -1 && pPars.nStart == 0 {
        pPars.iFrame = f-1; }
//ABC_PRT( "CNF generation runtime", clkOther );
finish:
    if pPars.fVerbose {
        Abc_Print( 1, "Runtime:  " );
        Abc_Print( 1, "CNF = %.1f sec (%.1f %%)  ",   1.0*clkOther/int64(time.Second),   100.0*clkOther/(Abc_Clock() - clkTotal)   );
        Abc_Print( 1, "UNSAT = %.1f sec (%.1f %%)  ", 1.0*nTimeUnsat/int64(time.Second), 100.0*nTimeUnsat/(Abc_Clock() - clkTotal) );
        Abc_Print( 1, "SAT = %.1f sec (%.1f %%)  ",   1.0*nTimeSat/int64(time.Second),   100.0*nTimeSat/(Abc_Clock() - clkTotal)   );
        Abc_Print( 1, "UNDEC = %.1f sec (%.1f %%)",   1.0*nTimeUndec/int64(time.Second), 100.0*nTimeUndec/(Abc_Clock() - clkTotal) );
        Abc_Print( 1, "\n" );
    }
    Saig_Bmc3ManStop( p );
    fflush( os.Stdout );
    if pLogFile != nil {
        _ = (pLogFile).Close(); }
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

