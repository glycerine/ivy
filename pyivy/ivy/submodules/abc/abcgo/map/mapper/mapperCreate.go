package mapper

import (
	"fmt"
	"math"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [mapperCreate.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperCreate.c,v 1.15 2005/02/28 05:34:26 alanmi Exp $]

***********************************************************************/

//#include "mapperInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void            Map_TableCreate( Map_Man_t * p );

// static void            Map_TableResize( Map_Man_t * p );


// hash key for the structural hash table
func Map_HashKey2(p0 *Map_Node_t, p1 *Map_Node_t, TableSize int32) uint32 { return uint32((((ABC_PTRUINT_T)(p0) + (ABC_PTRUINT_T)(p1) * 12582917) % TableSize)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads parameters from the mapping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_ManReadInputNum(p *Map_Man_t) int32 { return p.nInputs;    }
func Map_ManReadOutputNum(p *Map_Man_t) int32 { return p.nOutputs;   }
func Map_ManReadBufNum(p *Map_Man_t) int32 { return Map_NodeVecReadSize(p.vMapBufs);   }
func Map_ManReadInputs(p *Map_Man_t) **Map_Node_t { return p.pInputs;    }
func Map_ManReadOutputs(p *Map_Man_t) **Map_Node_t { return p.pOutputs;   }
func Map_ManReadBufs(p *Map_Man_t) **Map_Node_t { return Map_NodeVecReadArray(p.vMapBufs);  }
func Map_ManReadBufDriver(p *Map_Man_t, i int32) *Map_Node_t { return Map_ManReadBufs(p)[i].p1;           }
func Map_ManReadConst1(p *Map_Man_t) *Map_Node_t { return p.pConst1;    }
func Map_ManReadInputArrivals(p *Map_Man_t) *Map_Time_t { return p.pInputArrivals; }
func Map_ManReadOutputRequireds(p *Map_Man_t) *Map_Time_t { return p.pOutputRequireds; }
func Map_ManReadGenLib(p *Map_Man_t) *Mio_Library_t { return p.pSuperLib.pGenlib; }
func Map_ManReadVerbose(p *Map_Man_t) int32 { return p.fVerbose;   }
func Map_ManReadAreaFinal(p *Map_Man_t) float32 { return p.AreaFinal;  }
func Map_ManReadRequiredGlo(p *Map_Man_t) float32 { return p.fRequiredGlo; }
func Map_ManSetOutputNames(p *Map_Man_t, ppNames *string) { p.ppOutputNames = ppNames;}
func Map_ManSetAreaRecovery(p *Map_Man_t, fAreaRecovery int32) { p.fAreaRecovery = fAreaRecovery;}
func Map_ManSetDelayTarget(p *Map_Man_t, DelayTarget float32) { p.DelayTarget = DelayTarget;}
func Map_ManSetInputArrivals(p *Map_Man_t, pArrivals *Map_Time_t) { p.pInputArrivals = pArrivals;    }
func Map_ManSetOutputRequireds(p *Map_Man_t, pRequireds *Map_Time_t) { p.pOutputRequireds = pRequireds; }
func Map_ManSetObeyFanoutLimits(p *Map_Man_t, fObeyFanoutLimits int32) { p.fObeyFanoutLimits = fObeyFanoutLimits;     }
func Map_ManSetNumIterations(p *Map_Man_t, nIterations int32) { p.nIterations = nIterations;     }
func Map_ManReadFanoutViolations(p *Map_Man_t) int32 { return p.nFanoutViolations;   }  
func Map_ManSetFanoutViolations(p *Map_Man_t, nVio int32) { p.nFanoutViolations = nVio;   }  
func Map_ManSetChoiceNodeNum(p *Map_Man_t, nChoiceNodes int32) { p.nChoiceNodes = nChoiceNodes; }  
func Map_ManSetChoiceNum(p *Map_Man_t, nChoices int32) { p.nChoices = nChoices;     }   
func Map_ManSetVerbose(p *Map_Man_t, fVerbose int32) { p.fVerbose = fVerbose;     }   
func Map_ManSetSwitching(p *Map_Man_t, fSwitching int32) { p.fSwitching = fSwitching; }   
func Map_ManSetSkipFanout(p *Map_Man_t, fSkipFanout int32) { p.fSkipFanout = fSkipFanout; }   
func Map_ManSetUseProfile(p *Map_Man_t) { p.fUseProfile = 1;         }   

/**Function*************************************************************

  Synopsis    [Reads parameters from the mapping node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeReadMan(p *Map_Node_t) *Map_Man_t { return p.p;                  }
func Map_NodeReadData(p *Map_Node_t, fPhase int32) string { if fPhase != 0 {
    return p.pData1;
} else {
    return p.pData0;
}  }
func Map_NodeReadNum(p *Map_Node_t) int32 { return p.Num;                }
// inlining of macro Map_Regular
func Map_NodeReadLevel(p *Map_Node_t) int32 { return ((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) .Level; }
func Map_NodeReadCuts(p *Map_Node_t) *Map_Cut_t { return p.pCuts;              }
func Map_NodeReadCutBest(p *Map_Node_t, fPhase int32) *Map_Cut_t { return p.pCutBest[fPhase];   }
func Map_NodeReadOne(p *Map_Node_t) *Map_Node_t { return p.p1;                 }
func Map_NodeReadTwo(p *Map_Node_t) *Map_Node_t { return p.p2;                 }
func Map_NodeSetData(p *Map_Node_t, fPhase int32, pData string) { if fPhase != 0 { p.pData1 = pData; } else { p.pData0 = pData; } }
func Map_NodeSetNextE(p *Map_Node_t, pNextE *Map_Node_t) { p.pNextE = pNextE;       }
func Map_NodeSetRepr(p *Map_Node_t, pRepr *Map_Node_t) { p.pRepr = pRepr;         }
func Map_NodeSetSwitching(p *Map_Node_t, Switching float32) { p.Switching = Switching; }

/**Function*************************************************************

  Synopsis    [Checks the type of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// inlining of macro Map_Regular
func Map_NodeIsConst(p *Map_Node_t) int32 {  return (((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).Num == -1;    }
// inlining of macro Map_Regular
func Map_NodeIsVar(p *Map_Node_t) int32 {  return (((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).p1 == nil && (((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).Num >= 0; }
// inlining of macro Map_Regular
func Map_NodeIsBuf(p *Map_Node_t) int32 {  return (((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).p1 != nil && (((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).p2 == nil;  }
// inlining of macro Map_Regular
func Map_NodeIsAnd(p *Map_Node_t) int32 {  return (((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).p1 != nil && (((Map_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).p2 != nil;  }
// inlining of macro Map_IsComplement
func Map_NodeComparePhase(p1 *Map_Node_t, p2 *Map_Node_t) int32 { assert(!(((int32(((ABC_PTRUINT_T) (p1) & 01)) != 0))), "!/*__starting_point_inline_macro:Map_IsComplement__*/((int32(((ABC_PTRUINT_T) (p1) & 01))))"); assert(!(((int32(((ABC_PTRUINT_T) (p2) & 01)) != 0))), "!/*__starting_point_inline_macro:Map_IsComplement__*/((int32(((ABC_PTRUINT_T) (p2) & 01))))"); return p1.fInv ^ p2.fInv; }

/**Function*************************************************************

  Synopsis    [Reads parameters from the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_CutReadSuperBest(p *Map_Cut_t, fPhase int32) *Map_Super_t { return p.M[fPhase].pSuperBest;}
func Map_CutReadSuper0(p *Map_Cut_t) *Map_Super_t { return p.M[0].pSuperBest;}
func Map_CutReadSuper1(p *Map_Cut_t) *Map_Super_t { return p.M[1].pSuperBest;}
func Map_CutReadLeavesNum(p *Map_Cut_t) int32 { return p.nLeaves;  }
func Map_CutReadLeaves(p *Map_Cut_t) **Map_Node_t { return p.ppLeaves; }
func Map_CutReadPhaseBest(p *Map_Cut_t, fPhase int32) uint32 { return p.M[fPhase].uPhaseBest;}
func Map_CutReadPhase0(p *Map_Cut_t) uint32 { return p.M[0].uPhaseBest;}
func Map_CutReadPhase1(p *Map_Cut_t) uint32 { return p.M[1].uPhaseBest;}
func Map_CutReadNext(p *Map_Cut_t) *Map_Cut_t { return p.pNext;          }

/**Function*************************************************************

  Synopsis    [Reads parameters from the supergate.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_SuperReadFormula(p *Map_Super_t) string {  return p.pFormula; }
func Map_SuperReadRoot(p *Map_Super_t) *Mio_Gate_t {  return p.pRoot;    }
func Map_SuperReadNum(p *Map_Super_t) int32 {  return p.Num;      }
func Map_SuperReadFanins(p *Map_Super_t) **Map_Super_t {  return p.pFanins;  }
func Map_SuperReadFaninNum(p *Map_Super_t) int32 {  return p.nFanins;  }
func Map_SuperReadNext(p *Map_Super_t) *Map_Super_t {  return p.pNext;    }
func Map_SuperReadNumPhases(p *Map_Super_t) int32 {  return p.nPhases;  }
func Map_SuperReadPhases(p *Map_Super_t) *unsigned char {  return p.uPhases;  }
func Map_SuperReadFanoutLimit(p *Map_Super_t) int32 {  return p.nFanLimit;}

func Map_SuperLibReadGenLib(p *Map_SuperLib_t) *Mio_Library_t {  return p.pGenlib;  }
func Map_SuperLibReadAreaInv(p *Map_SuperLib_t) float32 {  return p.AreaInv;  }
func Map_SuperLibReadDelayInv(p *Map_SuperLib_t) Map_Time_t {  return p.tDelayInv;}
func Map_SuperLibReadVarsMax(p *Map_SuperLib_t) int32 {  return p.nVarsMax; }


/**Function*************************************************************

  Synopsis    [Create the mapping manager.]

  Description [The number of inputs and outputs is assumed to be
  known is advance. It is much simpler to have them fixed upfront.
  When it comes to representing the object graph in the form of
  AIG, the resulting manager is similar to the regular AIG manager, 
  except that it does not use reference counting (and therefore
  does not have garbage collections). It does have table resizing.
  The data structure is more flexible to represent additional 
  information needed for mapping.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_ManCreate(nInputs int32, nOutputs int32, fVerbose int32) *Map_Man_t {
    var p *Map_Man_t;
    var i int32;

    // derive the supergate library
    if Abc_FrameReadLibSuper() == nil {
        fmt.Printf("The supergate library is not specified. Use \"read_super\".\n");
        return nil;
    }

    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Map_Man_t *) malloc(sizeof(Map_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Map_Man_t) );
    p.pSuperLib = (Map_SuperLib_t *)Abc_FrameReadLibSuper();
    p.nVarsMax  = p.pSuperLib.nVarsMax;
    p.fVerbose  = fVerbose;
    p.fEpsilon  = float32(0.001);
    assert(p.nVarsMax > 0, "p.nVarsMax > 0");

    if p.nVarsMax == 5 {
        Extra_Truth4VarN( &p.uCanons, &p.uPhases, &p.pCounters, 8 ); }

    // start various data structures
    Map_TableCreate( p );
    Map_MappingSetupTruthTables( p.uTruths );
    Map_MappingSetupTruthTablesLarge( p.uTruthsLarge );
//    printf( "Node = %d bytes. Cut = %d bytes. Super = %d bytes.\n", sizeof(Map_Node_t), sizeof(Map_Cut_t), sizeof(Map_Super_t) ); 
    p.mmNodes    = Extra_MmFixedStart( sizeof(Map_Node_t) );
    p.mmCuts     = Extra_MmFixedStart( sizeof(Map_Cut_t) );

    // make sure the constant node will get index -1
    p.nNodes = -1;
    // create the constant node
    p.pConst1    = Map_NodeCreate( p, nil, nil );
    p.vMapObjs   = Map_NodeVecAlloc( 100 );
    p.vMapBufs   = Map_NodeVecAlloc( 100 );
    p.vVisited   = Map_NodeVecAlloc( 100 );

    // create the PI nodes
    p.nInputs = nInputs;
    // inlining of macro ABC_ALLOC
    p.pInputs = ((Map_Node_t * *) malloc(sizeof(Map_Node_t *) * (size_t)(nInputs)));
    for  i = 0; i < nInputs; i++  {
        p.pInputs[i] = Map_NodeCreate( p, nil, nil );
    }

    // create the place for the output nodes
    p.nOutputs = nOutputs;
    // inlining of macro ABC_ALLOC
    p.pOutputs = ((Map_Node_t * *) malloc(sizeof(Map_Node_t *) * (size_t)(nOutputs)));
    memset( p.pOutputs, 0, sizeof(Map_Node_t *) * nOutputs );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates the mapping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_ManFree(p *Map_Man_t) {
//    int i;
//    for ( i = 0; i < p->vMapObjs->nSize; i++ )
//        Map_NodeVecFree( p->vMapObjs->pArray[i]->vFanouts );
//    Map_NodeVecFree( p->pConst1->vFanouts );
    Map_NodeVecFree( p.vMapObjs );
    Map_NodeVecFree( p.vMapBufs );
    Map_NodeVecFree( p.vVisited );
    if p.uCanons {   p.uCanons = nil; }
    if p.uPhases {   p.uPhases = nil; }
    if p.pCounters { p.pCounters = nil; }
    Extra_MmFixedStop( p.mmNodes );
    Extra_MmFixedStop( p.mmCuts );
    p.pNodeDelays = nil;
    p.pInputArrivals = nil;
    p.pOutputRequireds = nil;
    p.pInputs = nil;
    p.pOutputs = nil;
    p.pBins = nil;
    p.ppOutputNames = nil;
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Creates node delays.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_ManCreateNodeDelays(p *Map_Man_t, LogFan int32) {
    var pNode *Map_Node_t;
    var k int32;
    assert(p.pNodeDelays == nil, "p.pNodeDelays == nil");
    // inlining of macro ABC_CALLOC
    p.pNodeDelays = ((float *) calloc((size_t)(p.vMapObjs.nSize), sizeof(float)));
    for  k = 0; k < p.vMapObjs.nSize; k++  {
        pNode = p.vMapObjs.pArray[k];
        if pNode.nRefs == 0 {
            continue; }
        p.pNodeDelays[k] = 0.014426 * LogFan * p.pSuperLib.tDelayInv.Worst * math.Log(float64(float64(pNode.nRefs))); // 1.4426 = 1/ln(2)
//        printf( "%d = %d (%.2f)  ", k, pNode->nRefs, p->pNodeDelays[k] );
    }
//    printf( "\n" );
}


/**Function*************************************************************

  Synopsis    [Deallocates the mapping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_ManPrintTimeStats(p *Map_Man_t) {
    fmt.Printf("N-canonical = %d. Matchings = %d.  Phases = %d.  ", p.nCanons, p.nMatches, p.nPhases);
    fmt.Printf("Choice nodes = %d. Choices = %d.\n", p.nChoiceNodes, p.nChoices);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("ToMap"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeToMap)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Cuts "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCuts)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Truth"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTruth)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Match"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMatch)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Area "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeArea)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Sweep"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeSweep)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("ToNet"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeToNet)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    if p.time1 { Abc_Print(1, "%s =", ("time1"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time1)))/(float64(int64(time.Second))));; }
    // inlining of macro ABC_PRT
    if p.time2 { Abc_Print(1, "%s =", ("time2"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time2)))/(float64(int64(time.Second))));; }
    // inlining of macro ABC_PRT
    if p.time3 { Abc_Print(1, "%s =", ("time3"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.time3)))/(float64(int64(time.Second))));; }
}

/**Function*************************************************************

  Synopsis    [Prints the mapping stats.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_ManPrintStatsToFile(pName string, Area float32, Delay float32, Time abctime) {
    var pTable *os.File;
    pTable = os.OpenFile("map_stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    fmt.Fprintf(pTable, "%s ", pName);
    fmt.Fprintf(pTable, "%4.2f ", Area);
    fmt.Fprintf(pTable, "%4.2f ", Delay);
    fmt.Fprintf(pTable, "%4.2f\n", float32((Time))/float32((int64(time.Second))));
    _ = (pTable).Close();
}

/**Function*************************************************************

  Synopsis    [Creates a new node.]

  Description [This procedure should be called to create the constant
  node and the PI nodes first.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeCreate(p *Map_Man_t, p1 *Map_Node_t, p2 *Map_Node_t) *Map_Node_t {
    var pNode *Map_Node_t;
    // create the node
    pNode = (Map_Node_t *)Extra_MmFixedEntryFetch( p.mmNodes );
    memset( pNode, 0, sizeof(Map_Node_t) );
    pNode.tRequired[0].Rise = pNode.tRequired[0].Fall = pNode.tRequired[0].Worst = MAP_FLOAT_LARGE;
    pNode.tRequired[1].Rise = pNode.tRequired[1].Fall = pNode.tRequired[1].Worst = MAP_FLOAT_LARGE;
    pNode.p1  = p1; 
    pNode.p2  = p2;
    pNode.p = p;
    // set the number of this node
    pNode.Num = p.nNodes++;
    // place to store the fanouts
//    pNode->vFanouts = Map_NodeVecAlloc( 5 );
    // store this node in the internal array
    if pNode.Num >= 0 {
        Map_NodeVecPush( p.vMapObjs, pNode );
    } else {
        pNode.fInv = 1; }
    // set the level of this node
    if p1 != nil {
#ifdef MAP_ALLOCATE_FANOUT
        // create the fanout info
        Map_NodeAddFaninFanout( Map_Regular(p1), pNode );
        if ( p2 )
        Map_NodeAddFaninFanout( Map_Regular(p2), pNode );
#endif

        if p2 != nil {
            // inlining of macro MAP_MAX
            // inlining of macro Map_Regular
            pNode.Level = 1 + (((((Map_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .Level) > (((Map_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .Level))? (((Map_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .Level) : (((Map_Node_t *)((ABC_PTRUINT_T)(pNode.p2) & ~01)) .Level));
            // inlining of macro Map_NodeIsSimComplement
            // inlining of macro Map_IsComplement
            // inlining of macro Map_Regular
            pNode.fInv  = (((int32(((ABC_PTRUINT_T) (p1) & 01))))? !(((Map_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .fInv) : (p1).fInv) & (((int32(((ABC_PTRUINT_T) (p2) & 01))))? !(((Map_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .fInv) : (p2).fInv);
        
        } else {
            // inlining of macro Map_Regular
            pNode.Level = ((Map_Node_t *)((ABC_PTRUINT_T)(pNode.p1) & ~01)) .Level;
            // inlining of macro Map_NodeIsSimComplement
            
            // inlining of macro Map_IsComplement
            if (int32(((ABC_PTRUINT_T) (p1) & 01)) != 0) {
                // inlining of macro Map_Regular
                pNode.fInv = !(((Map_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .fInv);
            } else {
                pNode.fInv = (p1).fInv;
            }
        }
    }
    // reference the inputs (will be used to compute the number of fanouts)
    // inlining of macro Map_NodeRef
    // inlining of macro Map_Regular
    if p1 != nil { ((((Map_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) ).nRefs++); }
    // inlining of macro Map_NodeRef
    // inlining of macro Map_Regular
    if p2 != nil { ((((Map_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) ).nRefs++); }

    pNode.nRefEst[0] = pNode.nRefEst[1] = -1;
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Create the unique table of AND gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_TableCreate(pMan *Map_Man_t) {
    assert(pMan.pBins == nil, "pMan.pBins == nil");
    pMan.nBins = Abc_PrimeCudd(5000);
    // inlining of macro ABC_ALLOC
    pMan.pBins = ((Map_Node_t * *) malloc(sizeof(Map_Node_t *) * (size_t)(pMan.nBins)));
    memset( pMan.pBins, 0, sizeof(Map_Node_t *) * pMan.nBins );
    pMan.nNodes = 0;
}

/**Function*************************************************************

  Synopsis    [Looks up the AND2 node in the unique table.]

  Description [This procedure implements one-level hashing. All the nodes
  are hashed by their children. If the node with the same children was already
  created, it is returned by the call to this procedure. If it does not exist,
  this procedure creates a new node with these children. ]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeAnd(pMan *Map_Man_t, p1 *Map_Node_t, p2 *Map_Node_t) *Map_Node_t {
    var pEnt *Map_Node_t;
    var Key uint32;

    if p1 == p2 {
        return p1; }
    // inlining of macro Map_Not
    if p1 == ((Map_Node_t *)((ABC_PTRUINT_T)(p2) ^ 01)) {
        // inlining of macro Map_Not
        return ((Map_Node_t *)((ABC_PTRUINT_T)(pMan.pConst1) ^ 01)) ; }
    if Map_NodeIsConst(p1) != 0 {
        if p1 == pMan.pConst1 {
            return p2; }
        // inlining of macro Map_Not
        return ((Map_Node_t *)((ABC_PTRUINT_T)(pMan.pConst1) ^ 01)) ;
    }
    if Map_NodeIsConst(p2) != 0 {
        if p2 == pMan.pConst1 {
            return p1; }
        // inlining of macro Map_Not
        return ((Map_Node_t *)((ABC_PTRUINT_T)(pMan.pConst1) ^ 01)) ;
    }

    // inlining of macro Map_Regular
    if  ((Map_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .Num > ((Map_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .Num {
        pEnt = p1;
p1 = p2;
p2 = pEnt; }

    Key = Map_HashKey2( p1, p2, pMan.nBins );
    for  pEnt = pMan.pBins[Key]; (pEnt != nil); pEnt = pEnt.pNext  {
        if pEnt.p1 == p1 && pEnt.p2 == p2 {
            return pEnt; }
    }
    // resize the table
    if pMan.nNodes >= 2 * pMan.nBins {
        Map_TableResize( pMan );
        Key = Map_HashKey2( p1, p2, pMan.nBins );
    }
    // create the new node
    pEnt = Map_NodeCreate( pMan, p1, p2 );
    // add the node to the corresponding linked list in the table
    pEnt.pNext = pMan.pBins[Key];
    pMan.pBins[Key] = pEnt;
    return pEnt;
}


/**Function*************************************************************

  Synopsis    [Resizes the table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_TableResize(pMan *Map_Man_t) {
    var pBinsNew **Map_Node_t;
    var pEnt, pEnt2 *Map_Node_t;
    var nBinsNew, Counter, i int32;
    var clk abctime;
    var Key uint32;

clk = Abc_Clock();
    // get the new table size
    nBinsNew = Abc_PrimeCudd(2 * pMan.nBins); 
    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Map_Node_t * *) malloc(sizeof(Map_Node_t *) * (size_t)(nBinsNew)));
    memset( pBinsNew, 0, sizeof(Map_Node_t *) * nBinsNew );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < pMan.nBins; i++  {
        for  pEnt = pMan.pBins[i], pEnt2 = pEnt? pEnt.pNext: nil; (pEnt != nil); 
              pEnt = pEnt2, pEnt2 = pEnt? pEnt.pNext: nil  {
            Key = Map_HashKey2( pEnt.p1, pEnt.p2, nBinsNew );
            pEnt.pNext = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            Counter++;
        }
    }
    assert(Counter == pMan.nNodes - pMan.nInputs, "Counter == pMan.nNodes - pMan.nInputs");
    if pMan.fVerbose {
//        printf( "Increasing the unique table size from %6d to %6d. ", pMan->nBins, nBinsNew );
//        ABC_PRT( "Time", Abc_Clock() - clk );
    }
    // replace the table and the parameters
    pMan.pBins = nil;
    pMan.pBins = pBinsNew;
    pMan.nBins = nBinsNew;
}



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeBuf(p *Map_Man_t, p1 *Map_Node_t) *Map_Node_t {
    var pNode *Map_Node_t  = Map_NodeCreate( p, p1, nil );
    Map_NodeVecPush( p.vMapBufs, pNode );
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Sets the node to be equivalent to the given one.]

  Description [This procedure is a work-around for the equivalence check.
  Does not verify the equivalence. Use at the user's risk.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Map_NodeSetChoice(pMan *Map_Man_t, pNodeOld *Map_Node_t, pNodeNew *Map_Node_t) {
    pNodeNew.pNextE = pNodeOld.pNextE;
    pNodeOld.pNextE = pNodeNew;
    pNodeNew.pRepr  = pNodeOld;
}



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

