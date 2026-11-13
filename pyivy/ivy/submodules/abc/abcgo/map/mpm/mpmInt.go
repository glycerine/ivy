package mpm

/**CFile****************************************************************

  FileName    [mpmInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Configurable technology mapper.]

  Synopsis    [Interal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 1, 2013.]

  Revision    [$Id: mpmInt.h,v 1.00 2013/06/01 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////
 
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/tim/tim.h"
//#include "misc/mem/mem2.h"
//#include "misc/vec/vec.h"
//#include "misc/vec/vecMem.h"
//#include "misc/vec/vecHsh.h"
//#include "misc/vec/vecWec.h"
//#include "misc/util/utilTruth.h"
//#include "mpmMig.h"
//#include "mpm.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////
 
const MPM_CUT_MAX =      32

const MPM_UNIT_TIME =     1
const MPM_UNIT_AREA =    20
const MPM_UNIT_EDGE =    50
const MPM_UNIT_REFS =   100

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

     // 8 bytes + NLeaves * 4 bytes
type Mpm_Cut_t struct {
    hNext              int32;                    // next cut
    iFunc uint32; /* C bitfield: 25 bits; review packing and truncation */           // function
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */
    fUseless uint32; /* C bitfield: 1 bits; review packing and truncation */           // internal flag
    nLeaves uint32; /* C bitfield: 5 bits; review packing and truncation */           // leaves
    pLeaves [1]int32;               // leaves
}
     // 48 bytes
type Mpm_Uni_t struct { 
    mTime              int32;                    // arrival time
    mArea              int32;                    // area (flow)
    mEdge              int32;                    // edge (flow)
    mAveRefs              int32;                 // area references
    uSign             word;                    // cut signature
    Cost              int32;                     // user cost
    pCut        Mpm_Cut_t;                     // new cut
    Data [MPM_VAR_MAX-1]int32;      // padding
}

   
type Mpm_Dsd_t struct {
    nVars              int32;                    // support size
    nAnds              int32;                    // the number of AND gates
    nClauses              int32;                 // the number of CNF clauses
    uTruth             word;                   // truth table
    pStr            string;                     // description 
}

   
type Mpm_Man_t struct {
    pMig       *Mig_Man_t;                     // AIG manager
    pPars       *Mpm_Par_t;                    // mapping parameters
    // mapping parameters
    nLutSize              int32;                 // LUT size
    nNumCuts              int32;                 // cut count
    nTruWords              int32;                // words in the truth table
    pLibLut    *Mpm_LibLut_t;                  // LUT library
    // mapping attributes  
    fMainRun              int32;                 // after preprocessing is finished
    GloRequired              int32;              // global arrival time
    GloArea             word;                  // total area
    GloEdge             word;                  // total edge
    // cut management
    pManCuts      *Mmr_Step_t;                 // cut memory
    // temporary cut storage
    nCutStore              int32;                // number of cuts in storage
    pCutStore [MPM_CUT_MAX+1]*Mpm_Uni_t; // storage for cuts
    pCutUnits [MPM_CUT_MAX+1]Mpm_Uni_t; // cut info units
    vFreeUnits        Vec_Ptr_t;               // free cut info units
    vTemp       *Vec_Ptr_t;                    // storage for cuts
    // cut comparison
    int (* pCutCmp) (Mpm_Uni_t *, Mpm_Uni_t *);// procedure to compare cuts
    // fanin cuts/signatures
    nCuts [3]int32;                 // fanin cut counts
    pCuts [3][MPM_CUT_MAX+1]*Mpm_Cut_t;  // fanin cuts
    word             pSigns[3][MPM_CUT_MAX+1]; // fanin cut signatures
    // truth tables
    vTtMem       *Vec_Mem_t;                   // truth table memory and hash table
    funcCst0              int32;                 // constant 0
    funcVar0              int32;                 // variable 0
    word             Truth0[(1 << ((MPM_VAR_MAX)-6))];
    word             Truth1[(1 << ((MPM_VAR_MAX)-6))];
    word             TruthC[(1 << ((MPM_VAR_MAX)-6))];
    word             Truth[(1 << ((MPM_VAR_MAX)-6))];
    // DSD
    pDsd6       *Mpm_Dsd_t;                    // NPN class information
    pHash    *Hsh_IntMan_t;                    // maps DSD functions into NPN classes
    vConfgRes       *Vec_Int_t;                // configurations
    vPerm6       *Vec_Wrd_t;                   // permutations of DSD classes
    Perm6 [720][6]byte;            // permutations
    vMap2Perm       *Vec_Int_t;                // maps number into its permutation
    uPermMask [3]uint32;
    uComplMask [3]uint32;
    vNpnConfigs       *Vec_Wec_t;
    // mapping attributes
    vCutBests        Vec_Int_t;                // cut best
    vCutLists        Vec_Int_t;                // cut list
    vMigRefs        Vec_Int_t;                 // original references
    vMapRefs        Vec_Int_t;                 // exact mapping references
    vEstRefs        Vec_Int_t;                 // estimated mapping references
    vRequireds        Vec_Int_t;               // required time
    vTimes        Vec_Int_t;                   // arrival time
    vAreas        Vec_Int_t;                   // area
    vEdges        Vec_Int_t;                   // edge
    nCountDsd [600]int32;
    nNonDsd              int32;
    nNoMatch              int32;
    // statistics
    nCutsMerged              int32;
    nCutsMergedAll              int32;
    nSmallSupp              int32;
    timeDerive          abctime;
    timeMerge          abctime;
    timeEval          abctime;
    timeCompare          abctime;
    timeStore          abctime;
    timeOther          abctime;
    timeTotal          abctime;
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Mpm_ObjCutBest(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vCutBests, Mig_ObjId(pObj));            }
func Mpm_ObjSetCutBest(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vCutBests, Mig_ObjId(pObj), i);           }

func Mpm_CutWordNum(nLeaves int32) int32 { return ((sizeof(Mpm_Cut_t) + (nLeaves << 2)) >> 3);             }
func Mpm_CutFetch(p *Mpm_Man_t, h int32) *Mpm_Cut_t { var pCut *Mpm_Cut_t  = (Mpm_Cut_t *)Mmr_StepEntry( p.pManCuts, h );  assert(Mpm_CutWordNum(pCut.nLeaves) == (h & p.pManCuts.uMask), "Mpm_CutWordNum(pCut.nLeaves) == (h & p.pManCuts.uMask)"); return pCut; }
func Mpm_ObjCutBestP(p *Mpm_Man_t, pObj *Mig_Obj_t) *Mpm_Cut_t { return Mpm_CutFetch( p, Mpm_ObjCutBest(p, pObj) );              }

func Mpm_ObjCutList(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vCutLists, Mig_ObjId(pObj));            }
func Mpm_ObjCutListP(p *Mpm_Man_t, pObj *Mig_Obj_t) *int32 { return Vec_IntEntryP(&p.vCutLists, Mig_ObjId(pObj));           }
func Mpm_ObjSetCutList(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vCutLists, Mig_ObjId(pObj), i);           }

func Mpm_CutLeafNum(pCut *Mpm_Cut_t) int32 { return pCut.nLeaves;                                           }
func Mpm_CutTruth(p *Mpm_Man_t, iFunc int32) *word { return Vec_MemReadEntry(p.vTtMem, iFunc);                      }

func Mig_ObjMigRefNum(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vMigRefs, Mig_ObjId(pObj));             }
func Mig_ObjMigRefDec(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntAddToEntry(&p.vMigRefs, Mig_ObjId(pObj), -1);    }

func Mpm_ManCleanMapRefs(p *Mpm_Man_t) { Vec_IntFill( &p.vMapRefs, Mig_ManObjNum(p.pMig), 0 );         }
func Mpm_ObjMapRef(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vMapRefs, Mig_ObjId(pObj));             }
func Mpm_ObjSetMapRef(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vMapRefs, Mig_ObjId(pObj), i);            }
 
func Mpm_ObjEstRef(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vEstRefs, Mig_ObjId(pObj));             }
func Mpm_ObjSetEstRef(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vEstRefs, Mig_ObjId(pObj), i);            }

func Mpm_ManCleanRequired(p *Mpm_Man_t) { Vec_IntFill(&p.vRequireds,Mig_ManObjNum(p.pMig),ABC_INFINITY);}
func Mpm_ObjRequired(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vRequireds, Mig_ObjId(pObj));           }
func Mpm_ObjSetRequired(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vRequireds, Mig_ObjId(pObj), i);          }

func Mpm_ObjTime(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vTimes, Mig_ObjId(pObj));               }
func Mpm_ObjSetTime(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vTimes, Mig_ObjId(pObj), i);              }

func Mpm_ObjArea(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vAreas, Mig_ObjId(pObj));               }
func Mpm_ObjSetArea(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vAreas, Mig_ObjId(pObj), i);              }

func Mpm_ObjEdge(p *Mpm_Man_t, pObj *Mig_Obj_t) int32 { return Vec_IntEntry(&p.vEdges, Mig_ObjId(pObj));               }
func Mpm_ObjSetEdge(p *Mpm_Man_t, pObj *Mig_Obj_t, i int32) { Vec_IntWriteEntry(&p.vEdges, Mig_ObjId(pObj), i);              }

func Mpm_VarsClear(V2P, P2V []int32, nVars int32) {
 for i:=int32(0); i<nVars; i++ { V2P[i], P2V[i] = i, i }
}
func Mpm_VarsSwap(V2P, P2V []int32, iVar, jVar int32) {
 V2P[P2V[iVar]] = jVar
 V2P[P2V[jVar]] = iVar
 P2V[iVar] ^= P2V[jVar]
 P2V[jVar] ^= P2V[iVar]
 P2V[iVar] ^= P2V[jVar]
}

// iterators over object cuts
// Original macro:
// #define Mpm_ObjForEachCut( p, pObj, hCut, pCut )                         \
//     for ( hCut = Mpm_ObjCutList(p, pObj); hCut && (pCut = Mpm_CutFetch(p, hCut)); hCut = pCut->hNext )

// Original macro:
// #define Mpm_ObjForEachCutSafe( p, pObj, hCut, pCut, hNext )              \
//     for ( hCut = Mpm_ObjCutList(p, pObj); hCut && (pCut = Mpm_CutFetch(p, hCut)) && ((hNext = pCut->hNext), 1); hCut = hNext )


// iterators over cut leaves
// Original macro:
// #define Mpm_CutForEachLeafId( pCut, iLeafId, i )                         \
//     for ( i = 0; i < (int)pCut->nLeaves && ((iLeafId = Abc_Lit2Var(pCut->pLeaves[i])), 1); i++ )

// Original macro:
// #define Mpm_CutForEachLeafLit( pCut, iLeafLit, i )                         \
//     for ( i = 0; i < (int)pCut->nLeaves && ((iLeafLit = pCut->pLeaves[i]), 1); i++ )

// Original macro:
// #define Mpm_CutForEachLeaf( p, pCut, pLeaf, i )                          \
//     for ( i = 0; i < (int)pCut->nLeaves && (pLeaf = Mig_ManObj(p, Abc_Lit2Var(pCut->pLeaves[i]))); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== mpmAbc.c ===========================================================*/
// extern Mig_Man_t *           Mig_ManCreate( void * pGia );

// extern void *                Mpm_ManFromIfLogic( Mpm_Man_t * pMan );

/*=== mpmMan.c ===========================================================*/
// extern Mpm_Man_t *           Mpm_ManStart( Mig_Man_t * pMig, Mpm_Par_t * pPars );

// extern void                  Mpm_ManStop( Mpm_Man_t * p );

// extern void                  Mpm_ManPrintStatsInit( Mpm_Man_t * p );

// extern void                  Mpm_ManPrintStats( Mpm_Man_t * p );

/*=== mpmDsd.c ===========================================================*/
// extern void                  Mpm_ManPrintDsdStats( Mpm_Man_t * p );

// extern void                  Mpm_ManPrintPerm( unsigned s );

// extern void                  Mpm_ManPrecomputePerms( Mpm_Man_t * p );

// extern word                  Mpm_CutTruthFromDsd( Mpm_Man_t * pMan, Mpm_Cut_t * pCut, int iDsdLit );

// extern int                   Mpm_CutCheckDsd6( Mpm_Man_t * p, word t );

// extern int                   Mpm_CutComputeDsd6( Mpm_Man_t * p, Mpm_Cut_t * pCut, Mpm_Cut_t * pCut0, Mpm_Cut_t * pCut1, Mpm_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, int Type );

/*=== mpmGates.c ===========================================================*/
// extern Vec_Wec_t *           Mpm_ManFindDsdMatches( Mpm_Man_t * p, void * pScl );

/*=== mpmLib.c ===========================================================*/
// extern Mpm_LibLut_t *        Mpm_LibLutSetSimple( int nLutSize );

// extern void                  Mpm_LibLutFree( Mpm_LibLut_t * pLib );

/*=== mpmMap.c ===========================================================*/
// extern void                  Mpm_CutPrint( Mpm_Cut_t * pCut );

// extern void                  Mpm_ManPrepare( Mpm_Man_t * p );

// extern void                  Mpm_ManPerform( Mpm_Man_t * p );

/*=== mpmTruth.c ===========================================================*/
// extern int                   Mpm_CutComputeTruth( Mpm_Man_t * p, Mpm_Cut_t * pCut, Mpm_Cut_t * pCut0, Mpm_Cut_t * pCut1, Mpm_Cut_t * pCutC, int fCompl0, int fCompl1, int fComplC, int Type );


// extern void                  Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );


//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

