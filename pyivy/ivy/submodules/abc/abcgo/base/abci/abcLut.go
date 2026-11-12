package abci

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcLut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Superchoicing for K-LUTs.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcLut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "opt/cut/cut.h"

//ABC_NAMESPACE_IMPL_START

const LARGE_LEVEL = 1000000

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const SCL_LUT_MAX =          6   // the maximum LUT size
const SCL_VARS_MAX =        15   // the maximum number of variables
const SCL_NODE_MAX =      1000   // the maximum number of nodes

   
type Abc_ManScl_t struct {
    // paramers
    nLutSize                int32;    // the LUT size
    nCutSizeMax                int32; // the max number of leaves of the cone
    nNodesMax                int32;   // the max number of divisors in the cone
    nWords                int32;      // the number of machine words in sim info
    // structural representation of the cone
    vLeaves         *Vec_Ptr_t;     // leaves of the cut
    vVolume         *Vec_Ptr_t;     // volume of the cut
    pBSet [SCL_VARS_MAX]int32; // bound set
    // functional representation of the cone
    uTruth          *uint32;      // truth table of the cone
    // representation of truth tables
    uVars         **uint32;       // elementary truth tables
    uSims         **uint32;       // truth tables of the nodes
    uCofs         **uint32;       // truth tables of the cofactors
}

var s_pLeaves *Vec_Ptr_t = nil;

// static Cut_Man_t * Abc_NtkStartCutManForScl( Abc_Ntk_t * pNtk, int nLutSize );

// static Abc_ManScl_t * Abc_ManSclStart( int nLutSize, int nCutSizeMax, int nNodesMax );

// static void Abc_ManSclStop( Abc_ManScl_t * p );

// static void Abc_NodeLutMap( Cut_Man_t * pManCuts, Abc_Obj_t * pObj );


// static Abc_Obj_t * Abc_NodeSuperChoiceLut( Abc_ManScl_t * pManScl, Abc_Obj_t * pObj );

// static int Abc_NodeDecomposeStep( Abc_ManScl_t * pManScl );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs superchoicing for K-LUTs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkSuperChoiceLut(pNtk *Abc_Ntk_t, nLutSize int32, nCutSizeMax int32, fVerbose int32) int32 {
    var pProgress *ProgressBar;
    var pManCut *Abc_ManCut_t;
    var pManScl *Abc_ManScl_t;
    var pManCuts *Cut_Man_t;
    var pObj, pFanin, pObjTop *Abc_Obj_t;
    var i, LevelMax, nNodes int32;
    var nNodesTried, nNodesDec, nNodesExist, nNodesUsed int32;

    assert(Abc_NtkIsSopLogic(pNtk), "Abc_NtkIsSopLogic(pNtk)");
    if nLutSize < 3 || nLutSize > SCL_LUT_MAX {
        fmt.Printf("LUT size (%d) does not belong to the interval: 3 <= LUT size <= %d\n", nLutSize, SCL_LUT_MAX);
        return 0;
    }
    if nCutSizeMax <= nLutSize || nCutSizeMax > SCL_VARS_MAX {
        fmt.Printf("Cut size (%d) does not belong to the interval: LUT size (%d) < Cut size <= %d\n", nCutSizeMax, nLutSize, SCL_VARS_MAX);
        return 0;
    }

    assert(nLutSize <= SCL_LUT_MAX, "nLutSize <= SCL_LUT_MAX");
    assert(nCutSizeMax <= SCL_VARS_MAX, "nCutSizeMax <= SCL_VARS_MAX");
    nNodesTried = nNodesDec = nNodesExist = nNodesUsed = 0;

    // set the delays of the CIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        pObj.Level = 0;
    }

//Abc_NtkLevel( pNtk );
 
    // start the managers
    pManScl = Abc_ManSclStart( nLutSize, nCutSizeMax, 1000 );
    pManCuts = Abc_NtkStartCutManForScl( pNtk, nLutSize );
    pManCut = Abc_NtkManCutStart( nCutSizeMax, 100000, 100000, 100000 );
    s_pLeaves = Abc_NtkManCutReadCutSmall( pManCut );
    pManScl.vVolume = Abc_NtkManCutReadVisited( pManCut );

    // process each internal node (assuming topological order of nodes!!!)
    nNodes = Abc_NtkObjNumMax(pNtk);
    pProgress = Extra_ProgressBarStart( os.Stdout, nNodes );
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
//        if ( i != nNodes-1 )
//            continue;
        Extra_ProgressBarUpdate( pProgress, i, nil );
        if i >= nNodes {
            break; }
        if Abc_ObjFaninNum(pObj) != 2 {
            continue; }
        nNodesTried++;

        // map this node using regular cuts
//        pObj->Level = 0;
        Abc_NodeLutMap( pManCuts, pObj );
        // compute the cut
        pManScl.vLeaves = Abc_NodeFindCut( pManCut, pObj, 0 );
        if Vec_PtrSize(pManScl.vLeaves) <= nLutSize {
            continue; }
        // get the volume of the cut
        if Vec_PtrSize(pManScl.vVolume) > SCL_NODE_MAX {
            continue; }
        nNodesDec++;

        // decompose the cut
        pObjTop = Abc_NodeSuperChoiceLut( pManScl, pObj );
        if pObjTop == nil {
            continue; }
        nNodesExist++;

        // if there is no delay improvement, skip; otherwise, update level
        if pObjTop.Level >= pObj.Level {
            Abc_NtkDeleteObj_rec( pObjTop, 1 );
            continue;
        }
        pObj.Level = pObjTop.Level;
        nNodesUsed++;
    }
    }
    Extra_ProgressBarStop( pProgress );

    // delete the managers
    Abc_ManSclStop( pManScl );
    Abc_NtkManCutStop( pManCut );
    Cut_ManStop( pManCuts );

    // get the largest arrival time
    LevelMax = 0;
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(pNtk)) && (((pObj) = Abc_NtkCo(pNtk, i)), 1); i++  {
        pFanin = Abc_ObjFanin0( pObj );
        // skip inv/buf
        if Abc_ObjFaninNum(pFanin) == 1 {
            pFanin = Abc_ObjFanin0( pFanin ); }
        // get the new level
        LevelMax = Abc_MaxInt( LevelMax, int32(pFanin.Level) );
    }

    if fVerbose != 0 {
    fmt.Printf("Try = %d. Dec = %d. Exist = %d. Use = %d. SUPER = %d levels of %d-LUTs.\n", nNodesTried, nNodesDec, nNodesExist, nNodesUsed, LevelMax, nLutSize); }
//    if ( fVerbose )
//    printf( "The network is superchoiced for %d levels of %d-LUTs.\n", LevelMax, nLutSize );

    // clean the data field
    // inlining of macro Abc_NtkForEachObj
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pObj) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pObj) == nil {} else {
        pObj.pNext = nil; }
    }

    // check
    if !Abc_NtkCheck( pNtk ) {
        fmt.Printf("Abc_NtkSuperChoiceLut: The network check has failed.\n");
        return 0;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Performs LUT mapping of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeLutMap(pManCuts *Cut_Man_t, pObj *Abc_Obj_t) {
    var pCut *Cut_Cut_t;
    var pFanin *Abc_Obj_t;
    var i, DelayMax int32;
    pCut = (Cut_Cut_t *)Abc_NodeGetCutsRecursive( pManCuts, pObj, 0, 0 );
    assert(pCut != nil, "pCut != nil");
    assert(pObj.Level == 0, "pObj.Level == 0");
    // go through the cuts
    pObj.Level = LARGE_LEVEL;
    for  pCut = pCut.pNext; (pCut != nil); pCut = pCut.pNext  {
        DelayMax = 0;
        for  i = 0; i < int32(pCut.nLeaves); i++  {
            pFanin = Abc_NtkObj( pObj.pNtk, pCut.pLeaves[i] );
//            assert( Abc_ObjIsCi(pFanin) || pFanin->Level > 0 ); // should hold if node ordering is topological
            if DelayMax < int32(pFanin.Level) {
                DelayMax = pFanin.Level; }
        }
        if int32(pObj.Level) > DelayMax {
            pObj.Level = DelayMax; }
    }
    assert(pObj.Level < LARGE_LEVEL, "pObj.Level < LARGE_LEVEL");
    pObj.Level++;
//    printf( "%d(%d) ", pObj->Id, pObj->Level );
}

/**Function*************************************************************

  Synopsis    [Starts the cut manager for rewriting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_NtkStartCutManForScl_Params Cut_Params_t;
var spStatic_Abc_NtkStartCutManForScl_pParams *Cut_Params_t = &spStatic_Abc_NtkStartCutManForScl_Params;
func Abc_NtkStartCutManForScl(pNtk *Abc_Ntk_t, nLutSize int32) *Cut_Man_t {
    
    var pManCut *Cut_Man_t;
    var pObj *Abc_Obj_t;
    var i int32;
    // start the cut manager
    memset( spStatic_Abc_NtkStartCutManForScl_pParams, 0, sizeof(Cut_Params_t) );
    spStatic_Abc_NtkStartCutManForScl_pParams.nVarsMax  = nLutSize; // the max cut size ("k" of the k-feasible cuts)
    spStatic_Abc_NtkStartCutManForScl_pParams.nKeepMax  = 500;   // the max number of cuts kept at a node
    spStatic_Abc_NtkStartCutManForScl_pParams.fTruth    = 0;     // compute truth tables
    spStatic_Abc_NtkStartCutManForScl_pParams.fFilter   = 1;     // filter dominated cuts
    spStatic_Abc_NtkStartCutManForScl_pParams.fSeq      = 0;     // compute sequential cuts
    spStatic_Abc_NtkStartCutManForScl_pParams.fDrop     = 0;     // drop cuts on the fly
    spStatic_Abc_NtkStartCutManForScl_pParams.fVerbose  = 0;     // the verbosiness flag
    spStatic_Abc_NtkStartCutManForScl_pParams.nIdsMax   = Abc_NtkObjNumMax( pNtk );
    pManCut = Cut_ManStart( spStatic_Abc_NtkStartCutManForScl_pParams );
    if spStatic_Abc_NtkStartCutManForScl_pParams.fDrop {
        Cut_ManSetFanoutCounts( pManCut, Abc_NtkFanoutCounts(pNtk) ); }
    // set cuts for PIs
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(pNtk)) && (((pObj) = Abc_NtkCi(pNtk, i)), 1); i++  {
        if Abc_ObjFanoutNum(pObj) > 0 {
            Cut_NodeSetTriv( pManCut, pObj.Id ); }
    }
    return pManCut;
}

/**Function*************************************************************

  Synopsis    [Starts the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManSclStart(nLutSize int32, nCutSizeMax int32, nNodesMax int32) *Abc_ManScl_t {
    var p *Abc_ManScl_t;
    var i, k int32;
    assert(sizeof(unsigned) == 4, "sizeof(unsigned) == 4");
    // inlining of macro ABC_ALLOC
    p = ((Abc_ManScl_t *) malloc(sizeof(Abc_ManScl_t) * (size_t)(1)));
    *p = Abc_ManScl_t{};
    p.nLutSize    = nLutSize;
    p.nCutSizeMax = nCutSizeMax;
    p.nNodesMax   = nNodesMax;
    p.nWords      = Extra_TruthWordNum(nCutSizeMax);
    // allocate simulation info
    p.uVars = (unsigned **)Extra_ArrayAlloc( nCutSizeMax, p.nWords, 4 );
    p.uSims = (unsigned **)Extra_ArrayAlloc( nNodesMax, p.nWords, 4 );
    p.uCofs = (unsigned **)Extra_ArrayAlloc( 2 << nLutSize, p.nWords, 4 );
    memset( p.uVars[0], 0, nCutSizeMax * p.nWords * 4 );
    // assign elementary truth tables
    for  k = 0; k < p.nCutSizeMax; k++  {
        for  i = 0; i < p.nWords * 32; i++  {
            if i & (1 << k) {
                p.uVars[k][i>>5] |= (1 << (i&31)); }
        }
    }
    // other data structures
//    p->vBound = Vec_IntAlloc( nCutSizeMax );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ManSclStop(p *Abc_ManScl_t) {
//    Vec_IntFree( p->vBound );
    p.uVars = nil;
    p.uSims = nil;
    p.uCofs = nil;
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSuperChoiceTruth(pManScl *Abc_ManScl_t) *uint32 {
    var pObj *Abc_Obj_t;
    var puData0 *uint32;
    var puData1 *uint32;
    var puData *uint32 = nil;
    var pSop string;
    var i, k int32;
    // set elementary truth tables
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManScl.vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(pManScl.vLeaves, i)), 1); i++  {
        pObj.pNext = (Abc_Obj_t *)pManScl.uVars[i];
    }
    // compute truth tables for internal nodes
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pManScl.vVolume)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(pManScl.vVolume, i)), 1); i++  {
        // set storage for the node's simulation info
        pObj.pNext = (Abc_Obj_t *)pManScl.uSims[i];
        // get pointer to the simulation info
        puData  = (unsigned *)pObj.pNext;
        puData0 = (unsigned *)Abc_ObjFanin0(pObj).pNext;
        puData1 = (unsigned *)Abc_ObjFanin1(pObj).pNext;
        // simulate
        pSop = (char *)pObj.pData;
        if pSop[0] == '0' && pSop[1] == '0' {
            for  k = 0; k < pManScl.nWords; k++  {
                puData[k] = ~puData0[k] & ~puData1[k];
            }
        } else if pSop[0] == '0' {
            for  k = 0; k < pManScl.nWords; k++  {
                puData[k] = ~puData0[k] & puData1[k];
            }
        } else if pSop[1] == '0' {
            for  k = 0; k < pManScl.nWords; k++  {
                puData[k] = puData0[k] & ~puData1[k];
            }
        } else { 
            for  k = 0; k < pManScl.nWords; k++  {
                puData[k] = puData0[k] & puData1[k];
            } }
    }
    return puData;
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSuperChoiceCollect2_rec(pObj *Abc_Obj_t, vVolume *Vec_Ptr_t) {
    if pObj.fMarkC {
        return; }
    pObj.fMarkC = 1;
    assert(Abc_ObjFaninNum(pObj) == 2, "Abc_ObjFaninNum(pObj) == 2");
    Abc_NodeSuperChoiceCollect2_rec( Abc_ObjFanin0(pObj), vVolume );
    Abc_NodeSuperChoiceCollect2_rec( Abc_ObjFanin1(pObj), vVolume );
    Vec_PtrPush( vVolume, pObj );
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSuperChoiceCollect2(pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t, vVolume *Vec_Ptr_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.fMarkC = 1;
    }
    Vec_PtrClear( vVolume );
    Abc_NodeSuperChoiceCollect2_rec( pRoot, vVolume );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.fMarkC = 0;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vVolume)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vVolume, i)), 1); i++  {
        pObj.fMarkC = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSuperChoiceCollect_rec(pObj *Abc_Obj_t, vLeaves *Vec_Ptr_t, vVolume *Vec_Ptr_t) {
    if pObj.fMarkB {
        Vec_PtrPush( vLeaves, pObj );
        pObj.fMarkB = 0;
    }
    if pObj.fMarkC {
        return; }
    pObj.fMarkC = 1;
    assert(Abc_ObjFaninNum(pObj) == 2, "Abc_ObjFaninNum(pObj) == 2");
    Abc_NodeSuperChoiceCollect_rec( Abc_ObjFanin0(pObj), vLeaves, vVolume );
    Abc_NodeSuperChoiceCollect_rec( Abc_ObjFanin1(pObj), vLeaves, vVolume );
    Vec_PtrPush( vVolume, pObj );
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description [Orders the leaves topologically.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSuperChoiceCollect(pRoot *Abc_Obj_t, vLeaves *Vec_Ptr_t, vVolume *Vec_Ptr_t) {
    var pObj *Abc_Obj_t;
    var i, nLeaves int32;
    nLeaves = Vec_PtrSize(vLeaves);
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.fMarkB = pObj.fMarkC = 1;
    }
    Vec_PtrClear( vVolume );
    Vec_PtrClear( vLeaves );
    Abc_NodeSuperChoiceCollect_rec( pRoot, vLeaves, vVolume );
    assert(Vec_PtrSize(vLeaves) == nLeaves, "Vec_PtrSize(vLeaves) == nLeaves");
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vLeaves)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vLeaves, i)), 1); i++  {
        pObj.fMarkC = 0;
    }
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vVolume)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vVolume, i)), 1); i++  {
        pObj.fMarkC = 0;
    }
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeLeavesRemove(vLeaves *Vec_Ptr_t, uPhase uint32, nVars int32) {
    var i int32;
    for  i = nVars - 1; i >= 0; i--  {
        if uPhase & (1 << i) {
            Vec_PtrRemove( vLeaves, Vec_PtrEntry(vLeaves, i) ); }
    }
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeGetLevel(pObj *Abc_Obj_t) int32 {
    var pFanin *Abc_Obj_t;
    var i, Level int32;
    Level = 0;
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pObj)) && (((pFanin) = Abc_ObjFanin(pObj, i)), 1); i++  {
        Level = Abc_MaxInt( Level, int32(pFanin.Level) );
    }
    return Level + 1;
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeSuperChoiceLut(p *Abc_ManScl_t, pObj *Abc_Obj_t) *Abc_Obj_t {
    var pFanin, pObjNew *Abc_Obj_t;
    var i, nVars, uSupport, nSuppVars int32;
    // collect the cone using DFS (excluding leaves)
    Abc_NodeSuperChoiceCollect2( pObj, p.vLeaves, p.vVolume );
    assert(Vec_PtrEntryLast(p.vVolume) == pObj, "Vec_PtrEntryLast(p.vVolume) == pObj");  
    // compute the truth table
    p.uTruth = Abc_NodeSuperChoiceTruth( p );
    // get the support of this truth table
    nVars = Vec_PtrSize(p.vLeaves);
    uSupport = Extra_TruthSupport(p.uTruth, nVars);
    nSuppVars = Extra_WordCountOnes(uSupport);
    assert(nSuppVars <= nVars, "nSuppVars <= nVars");
    if nSuppVars == 0 {
        pObj.Level = 0;
        return nil;
    }
    if nSuppVars == 1 {
        // find the variable
        for  i = 0; i < nVars; i++  {
            if uSupport & (1 << i) {
                break; }
        }
        assert(i < nVars, "i < nVars");
        pFanin = (Abc_Obj_t *)Vec_PtrEntry( p.vLeaves, i );
        pObj.Level = pFanin.Level;
        return nil;
    }
    // support-minimize the truth table
    if nSuppVars != nVars {
        Extra_TruthShrink( p.uCofs[0], p.uTruth, nSuppVars, nVars, uSupport );
        Extra_TruthCopy( p.uTruth, p.uCofs[0], nVars );
        Abc_NodeLeavesRemove( p.vLeaves, ((1 << nVars) - 1) & ~uSupport, nVars );
    }
//    return NULL;
    // decompose the truth table recursively
    for  Vec_PtrSize(p.vLeaves) > p.nLutSize  {
        if !(Abc_NodeDecomposeStep( p ) != 0) {
            // inlining of macro Vec_PtrForEachEntry
            for  i = 0; (i < Vec_PtrSize(p.vLeaves)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, i)), 1); i++  {
                if Abc_ObjIsNode(pFanin) && Abc_ObjFanoutNum(pFanin) == 0 {
                    Abc_NtkDeleteObj_rec( pFanin, 1 ); }
            }
            return nil;
        }
    }
    // create the topmost node
    pObjNew = Abc_NtkCreateNode( pObj.pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vLeaves)) && (((pFanin) = (Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, i)), 1); i++  {
        Abc_ObjAddFanin( pObjNew, pFanin );
    }
    // create the function
    pObjNew.pData = Abc_SopCreateFromTruth( (Mem_Flex_t *)pObj.pNtk.pManFunc, Vec_PtrSize(p.vLeaves), p.uTruth ); // need ISOP
    pObjNew.Level = Abc_NodeGetLevel( pObjNew );
    return pObjNew;
}

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in increasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeCompareLevelsInc(pp1 *int32, pp2 *int32) int32 {
    var pNode1, pNode2 *Abc_Obj_t;
    pNode1 = (Abc_Obj_t *)Vec_PtrEntry(s_pLeaves, *pp1);
    pNode2 = (Abc_Obj_t *)Vec_PtrEntry(s_pLeaves, *pp2);
    if pNode1.Level < pNode2.Level {
        return -1; }
    if pNode1.Level > pNode2.Level { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Selects the earliest arriving nodes from the array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeDecomposeSort(pLeaves **Abc_Obj_t, nVars int32, pBSet *int32, nLutSize int32) {
    var pTemp [SCL_VARS_MAX]*Abc_Obj_t;
    var i, k, kBest, LevelMin int32;
    assert(nLutSize < nVars, "nLutSize < nVars");
    assert(nVars <= SCL_VARS_MAX, "nVars <= SCL_VARS_MAX");
    // copy nodes into the internal storage
//    printf( "(" );
    for  i = 0; i < nVars; i++  {
        pTemp[i] = pLeaves[i];
//        printf( " %d", pLeaves[i]->Level );
    }
//    printf( " )\n" );
    // choose one node at a time
    for  i = 0; i < nLutSize; i++  {
        kBest = -1;
        LevelMin = LARGE_LEVEL;
        for  k = 0; k < nVars; k++  {
            if pTemp[k] && LevelMin > int32(pTemp[k].Level) {
                LevelMin = pTemp[k].Level;
                kBest = k;
            }
        }
        pBSet[i] = kBest;
        pTemp[kBest] = nil;
    }
}

/**Function*************************************************************

  Synopsis    [Performs superchoicing for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Abc_NodeDecomposeStep_pCofClasses [1<<SCL_LUT_MAX][1<<SCL_LUT_MAX]byte;
var spStatic_Abc_NodeDecomposeStep_nCofClasses [1<<SCL_LUT_MAX]byte;
func Abc_NodeDecomposeStep(p *Abc_ManScl_t) int32 {
    
    
    var pNtk *Abc_Ntk_t;
    var pObjNew *Abc_Obj_t;
    var pFanin *Abc_Obj_t;
    var pNodesNew [SCL_LUT_MAX]*Abc_Obj_t;
    var pTruthCof *uint32;
    var pTruthClass *uint32;
    var pTruth *uint32;
    var uPhase uint32;
    var i, k, c, v, w, nVars, nVarsNew, nClasses, nCofs int32;
    // set the network
    pNtk = ((Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, 0)).pNtk;
    // find the earliest nodes
    nVars = Vec_PtrSize(p.vLeaves);
    assert(nVars > p.nLutSize, "nVars > p.nLutSize");
/*
    for ( v = 0; v < nVars; v++ )
        p->pBSet[v] = v;
    qsort( (void *)p->pBSet, (size_t)nVars, sizeof(int), 
            (int (*)(const void *, const void *)) Abc_NodeCompareLevelsInc );
*/
    Abc_NodeDecomposeSort( (Abc_Obj_t **)Vec_PtrArray(p.vLeaves), Vec_PtrSize(p.vLeaves), p.pBSet, p.nLutSize );
    assert(((Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, p.pBSet[0])).Level <=
            ((Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, p.pBSet[1])).Level, "((Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, p.pBSet[0])).Level <=\n        ((Abc_Obj_t *)Vec_PtrEntry(p.vLeaves, p.pBSet[1])).Level");
    // cofactor w.r.t. the selected variables
    Extra_TruthCopy( p.uCofs[1], p.uTruth, nVars );
    c = 2;
    for  v = 0; v < p.nLutSize; v++  {
        for  k = 0; k < (1<<v); k++  {
            Extra_TruthCopy( p.uCofs[c], p.uCofs[c/2], nVars );
            Extra_TruthCopy( p.uCofs[c+1], p.uCofs[c/2], nVars );
            Extra_TruthCofactor0( p.uCofs[c], nVars, p.pBSet[v] );
            Extra_TruthCofactor1( p.uCofs[c+1], nVars, p.pBSet[v] );
            c += 2;
        }
    }
    assert(c == (2 << p.nLutSize), "c == (2 << p.nLutSize)");
    // count unique cofactors
    nClasses = 0;
    nCofs = (1 << p.nLutSize);
    for  i = 0; i < nCofs; i++  {
        pTruthCof = p.uCofs[ nCofs + i ];
        for  k = 0; k < nClasses; k++  {
            pTruthClass = p.uCofs[ nCofs + spStatic_Abc_NodeDecomposeStep_pCofClasses[k][0] ];
            if Extra_TruthIsEqual( pTruthCof, pTruthClass, nVars ) {
                spStatic_Abc_NodeDecomposeStep_pCofClasses[k][int32(spStatic_Abc_NodeDecomposeStep_nCofClasses[k]++) ] = i;
                break;
            }
        }
        if k != nClasses {
            continue; }
        // not found
        spStatic_Abc_NodeDecomposeStep_pCofClasses[nClasses][0] = i;
        spStatic_Abc_NodeDecomposeStep_nCofClasses[nClasses] = 1;
        nClasses++;
        if nClasses > nCofs/2 {
            return 0; }
    }
    // the number of cofactors is acceptable
    nVarsNew = Abc_Base2Log( nClasses );
    assert(nVarsNew < p.nLutSize, "nVarsNew < p.nLutSize");
    // create the remainder truth table
    // for each class of cofactors, multiply cofactor truth table by its code
    Extra_TruthClear( p.uTruth, nVars );
    for  k = 0; k < nClasses; k++  {
        pTruthClass = p.uCofs[ nCofs + spStatic_Abc_NodeDecomposeStep_pCofClasses[k][0] ];
        for  v = 0; v < nVarsNew; v++  {
            if k & (1 << v) {
                Extra_TruthAnd( pTruthClass, pTruthClass, p.uVars[p.pBSet[v]], nVars );
            } else {
                Extra_TruthSharp( pTruthClass, pTruthClass, p.uVars[p.pBSet[v]], nVars ); }
        }
        Extra_TruthOr( p.uTruth, p.uTruth, pTruthClass, nVars );
    }
    // create nodes
    pTruth = p.uCofs[0];
    for  v = 0; v < nVarsNew; v++  {
        Extra_TruthClear( pTruth, p.nLutSize );
        for  k = 0; k < nClasses; k++  {
            if k & (1 << v) {
                for  i = 0; i < spStatic_Abc_NodeDecomposeStep_nCofClasses[k]; i++  {
                    pTruthCof = p.uCofs[1];
                    Extra_TruthFill( pTruthCof, p.nLutSize );
                    for  w = 0; w < p.nLutSize; w++  {
                        if spStatic_Abc_NodeDecomposeStep_pCofClasses[k][i] & (1 << (p.nLutSize-1-w)) {
                            Extra_TruthAnd( pTruthCof, pTruthCof, p.uVars[w], p.nLutSize );
                        } else {
                            Extra_TruthSharp( pTruthCof, pTruthCof, p.uVars[w], p.nLutSize ); }
                    }
                    Extra_TruthOr( pTruth, pTruth, pTruthCof, p.nLutSize );
                } }
        }
        // implement the node
        pObjNew = Abc_NtkCreateNode( pNtk );
        for  i = 0; i < p.nLutSize; i++  {
            pFanin = (Abc_Obj_t *)Vec_PtrEntry( p.vLeaves, p.pBSet[i] );
            Abc_ObjAddFanin( pObjNew, pFanin );
        }
        // create the function
        pObjNew.pData = Abc_SopCreateFromTruth( (Mem_Flex_t *)pNtk.pManFunc, p.nLutSize, pTruth ); // need ISOP
        pObjNew.Level = Abc_NodeGetLevel( pObjNew );
        pNodesNew[v] = pObjNew;
    }
    // put the new nodes back into the list
    for  v = 0; v < nVarsNew; v++  {
        Vec_PtrWriteEntry( p.vLeaves, p.pBSet[v], pNodesNew[v] );
    }
    // compute the variables that should be removed
    uPhase = 0;
    for  v = nVarsNew; v < p.nLutSize; v++  {
        uPhase |= (1 << p.pBSet[v]);
    }
    // remove entries from the array
    Abc_NodeLeavesRemove( p.vLeaves, uPhase, nVars );
    // update truth table
    Extra_TruthShrink( p.uCofs[0], p.uTruth, nVars - p.nLutSize + nVarsNew, nVars, ((1 << nVars) - 1) & ~uPhase );
    Extra_TruthCopy( p.uTruth, p.uCofs[0], nVars );
    assert(!Extra_TruthVarInSupport( p.uTruth, nVars, nVars - p.nLutSize + nVarsNew ), "!Extra_TruthVarInSupport( p.uTruth, nVars, nVars - p.nLutSize + nVarsNew )");
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END



