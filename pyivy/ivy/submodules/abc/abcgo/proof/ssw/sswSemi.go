package ssw

import "time"

/**CFile****************************************************************

  FileName    [sswSemi.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Semiformal for equivalence classes.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswSemi.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

    // BMC manager

type Ssw_Sem_t struct {
    // parameters
    nConfMaxStart              int32;  // the starting conflict limit
    nConfMax              int32;       // the intermediate conflict limit
    nFramesSweep              int32;   // the number of frames to sweep
    fVerbose              int32;       // prints output statistics
    // equivalences considered
    pMan       *Ssw_Man_t;           // SAT sweeping manager
    vTargets       *Vec_Ptr_t;       // the nodes that are watched
    // storage for patterns
    nPatternsAlloc              int32; // the max number of interesting states
    nPatterns              int32;      // the number of patterns
    vPatterns       *Vec_Ptr_t;      // storage for the interesting states
    vHistory       *Vec_Int_t;       // what state and how many steps
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SemManStart(pMan *Ssw_Man_t, nConfMax int32, fVerbose int32) *Ssw_Sem_t {
    var p *Ssw_Sem_t;
    var pObj *Aig_Obj_t;
    var i int32;
    // create interpolation manager
    // inlining of macro ABC_ALLOC
    p = ((Ssw_Sem_t *) malloc(sizeof(Ssw_Sem_t) * (size_t)(1)));
    memset( p, 0, sizeof(Ssw_Sem_t) );
    p.nConfMaxStart  = nConfMax;
    p.nConfMax       = nConfMax;
    p.nFramesSweep   = Abc_MaxInt( (1<<21)/Aig_ManNodeNum(pMan.pAig), pMan.nFrames );
    p.fVerbose       = fVerbose;
    // equivalences considered
    p.pMan           = pMan;
    p.vTargets       = Vec_PtrAlloc( Saig_ManPoNum(p.pMan.pAig) );
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(p.pMan.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pMan.pAig.vCos, i)), 1); i++  {
        Vec_PtrPush( p.vTargets, Aig_ObjFanin0(pObj) );
    }
    // storage for patterns
    p.nPatternsAlloc = 512;
    p.nPatterns      = 1;
    p.vPatterns      = Vec_PtrAllocSimInfo( Aig_ManRegNum(p.pMan.pAig), Abc_BitWordNum(p.nPatternsAlloc) );
    Vec_PtrCleanSimInfo( p.vPatterns, 0, Abc_BitWordNum(p.nPatternsAlloc) );
    p.vHistory       = Vec_IntAlloc( 100 );
    Vec_IntPush( p.vHistory, 0 );
    // update arrays of the manager
    assert(false, "0");
/*
    ABC_FREE( p->pMan->pNodeToFrames );
    Vec_IntFree( p->pMan->vSatVars );
    p->pMan->pNodeToFrames = ABC_CALLOC( Aig_Obj_t *, Aig_ManObjNumMax(p->pMan->pAig) * p->nFramesSweep );
    p->pMan->vSatVars      = Vec_IntStart( Aig_ManObjNumMax(p->pMan->pAig) * (p->nFramesSweep+1) );
*/
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SemManStop(p *Ssw_Sem_t) {
    Vec_PtrFree( p.vTargets );
    Vec_PtrFree( p.vPatterns );
    Vec_IntFree( p.vHistory );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SemCheckTargets(p *Ssw_Sem_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vTargets)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.vTargets, i)), 1); i++  {
        if !Ssw_ObjIsConst1Cand(p.pMan.pAig, pObj) {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManFilterBmcSavePattern(p *Ssw_Sem_t) {
    var pInfo *uint32;
    var pObj *Aig_Obj_t;
    var i int32;
    if p.nPatterns >= p.nPatternsAlloc {
        return; }
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pMan.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pMan.pAig.vCis, i+Saig_ManPiNum(p.pMan.pAig))), 1); i++  {
        pInfo = (unsigned *)Vec_PtrEntry( p.vPatterns, i );
        if Abc_InfoHasBit( p.pMan.pPatWords, Saig_ManPiNum(p.pMan.pAig) + i ) {
            Abc_InfoSetBit( pInfo, p.nPatterns ); }
    }
    p.nPatterns++;
}

/**Function*************************************************************

  Synopsis    [Performs fraiging for the internal nodes.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_ManFilterBmc(pBmc *Ssw_Sem_t, iPat int32, fCheckTargets int32) int32 {
    var p *Ssw_Man_t  = pBmc.pMan;
    var pObj, pObjNew, pObjLi, pObjLo *Aig_Obj_t;
    var pInfo *uint32;
    var i, f, RetValue, fFirst int32;
    var clk abctime;
clk = Abc_Clock();

    // start initialized timeframes
    p.pFrames = Aig_ManStart( Aig_ManObjNumMax(p.pAig) * 3 );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i+Saig_ManPiNum(p.pAig))), 1); i++  {
        pInfo = (unsigned *)Vec_PtrEntry( pBmc.vPatterns, i );
        pObjNew = Aig_NotCond( Aig_ManConst1(p.pFrames), !Abc_InfoHasBit(pInfo, iPat) );
        Ssw_ObjSetFrame( p, pObj, 0, pObjNew );
    }

    // sweep internal nodes
    RetValue = pBmc.nFramesSweep;
    for  f = 0; f < pBmc.nFramesSweep; f++  {
        // map constants and PIs
        Ssw_ObjSetFrame( p, Aig_ManConst1(p.pAig), f, Aig_ManConst1(p.pFrames) );
        // inlining of macro Saig_ManForEachPi
        // inlining of macro Vec_PtrForEachEntryStop
        for  i = 0; (i < Saig_ManPiNum(p.pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vCis, i)), 1); i++  {
            Ssw_ObjSetFrame( p, pObj, f, Aig_ObjCreateCi(p.pFrames) );
        }
        // sweep internal nodes
        // inlining of macro Aig_ManForEachNode
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(p.pAig.vObjs)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(p.pAig.vObjs, i)), 1); i++  { if (pObj) == nil || !Aig_ObjIsNode(pObj) {} else {
            pObjNew = Aig_And( p.pFrames, Ssw_ObjChild0Fra(p, pObj, f), Ssw_ObjChild1Fra(p, pObj, f) );
            Ssw_ObjSetFrame( p, pObj, f, pObjNew );
            if Ssw_ManSweepNode( p, pObj, f, 1, nil ) {
                Ssw_ManFilterBmcSavePattern( pBmc );
                if fFirst == 0 {
                    fFirst = 1;
                    pBmc.nConfMax *= 10;
                }
            }
            if f > 0 && p.pMSat.pSat.stats.conflicts >= pBmc.nConfMax {
                RetValue = -1;
                break;
            }
        }
        }
        // quit if this is the last timeframe
        if p.pMSat.pSat.stats.conflicts >= pBmc.nConfMax {
            RetValue += f + 1;
            break;
        }
        if (fCheckTargets != 0) && (Ssw_SemCheckTargets( pBmc ) != 0) {
            break; }
        // transfer latch input to the latch outputs 
        // build logic cones for register outputs
        // inlining of macro Saig_ManForEachLiLo
        for  i = 0; (i < Saig_ManRegNum(p.pAig)) && (((pObjLi) = Saig_ManLi(p.pAig, i)), 1)            && (((pObjLo)=Saig_ManLo(p.pAig, i)), 1); i++  {
            pObjNew = Ssw_ObjChild0Fra(p, pObjLi,f);
            Ssw_ObjSetFrame( p, pObjLo, f+1, pObjNew );
            Ssw_CnfNodeAddToSolver( p.pMSat, Aig_Regular(pObjNew) );
        }
//Abc_Print( 1, "Frame %2d : Conflicts = %6d. \n", f, p->pSat->stats.conflicts );
    }
    if fFirst != 0 {
        pBmc.nConfMax /= 10; }

    // cleanup
    Ssw_ClassesCheck( p.ppClasses );
p.timeBmc += Abc_Clock() - clk;
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if one of the targets has failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_FilterUsingSemi(pMan *Ssw_Man_t, fCheckTargets int32, nConfMax int32, fVerbose int32) int32 {
    var p *Ssw_Sem_t;
    var RetValue, Frames, Iter int32;
    var clk abctime  = Abc_Clock();
    p = Ssw_SemManStart( pMan, nConfMax, fVerbose );
    if (fCheckTargets != 0) && (Ssw_SemCheckTargets( p ) != 0) {
        assert(false, "0");
        Ssw_SemManStop( p );
        return 1;
    }
    if fVerbose != 0 {
        Abc_Print( 1, "AIG : C = %6d. Cl = %6d. Nodes = %6d.  ConfMax = %6d. FramesMax = %6d.\n",
            Ssw_ClassesCand1Num(p.pMan.ppClasses), Ssw_ClassesClassNum(p.pMan.ppClasses),
            Aig_ManNodeNum(p.pMan.pAig), p.nConfMax, p.nFramesSweep );
    }
    RetValue = 0;
    for  Iter = 0; Iter < p.nPatterns; Iter++  {
clk = Abc_Clock();
        pMan.pMSat = Ssw_SatStart( 0 );
        Frames = Ssw_ManFilterBmc( p, Iter, fCheckTargets );
        if fVerbose != 0 {
            Abc_Print( 1, "%3d : C = %6d. Cl = %6d. NR = %6d. F = %3d. C = %5d. P = %3d. %s ",
                Iter, Ssw_ClassesCand1Num(p.pMan.ppClasses), Ssw_ClassesClassNum(p.pMan.ppClasses),
                Aig_ManNodeNum(p.pMan.pFrames), Frames, int32(p.pMan.pMSat.pSat.stats.conflicts), p.nPatterns,
                p.pMan.nSatFailsReal? "f" : " " );
            // inlining of macro ABC_PRT
            Abc_Print(1, "%s =", ("T"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
        }
        Ssw_ManCleanup( p.pMan );
        if (fCheckTargets != 0) && (Ssw_SemCheckTargets( p ) != 0) {
            Abc_Print( 1, "Target is hit!!!\n" );
            RetValue = 1;
        }
        if p.nPatterns >= p.nPatternsAlloc {
            break; }
    }
    Ssw_SemManStop( p );

    pMan.nStrangers = 0;
    pMan.nSatCalls = 0;
    pMan.nSatProof = 0;
    pMan.nSatFailsReal = 0;
    pMan.nSatCallsUnsat = 0;
    pMan.nSatCallsSat = 0;
    pMan.timeSimSat = 0;
    pMan.timeSat = 0;
    pMan.timeSatSat = 0;
    pMan.timeSatUnsat = 0;
    pMan.timeSatUndec = 0;
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
