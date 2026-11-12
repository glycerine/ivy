package fpga

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [fpgaCreate.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaCreate.c,v 1.8 2004/09/30 21:18:09 satrajit Exp $]

***********************************************************************/

//#include "fpgaInt.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void            Fpga_TableCreate( Fpga_Man_t * p );

// static void            Fpga_TableResize( Fpga_Man_t * p );

// static Fpga_Node_t *   Fpga_TableLookup( Fpga_Man_t * p, Fpga_Node_t * p1, Fpga_Node_t * p2 );


// hash key for the structural hash table
func Fpga_HashKey2(p0 *Fpga_Node_t, p1 *Fpga_Node_t, TableSize int32) uint32 { return uint32((((ABC_PTRUINT_T)(p0) + (ABC_PTRUINT_T)(p1) * 12582917) % TableSize)); }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Reads parameters of the mapping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_ManReadInputNum(p *Fpga_Man_t) int32 { return p.nInputs;    }
func Fpga_ManReadOutputNum(p *Fpga_Man_t) int32 { return p.nOutputs;   }
func Fpga_ManReadInputs(p *Fpga_Man_t) **Fpga_Node_t { return p.pInputs;    }
func Fpga_ManReadOutputs(p *Fpga_Man_t) **Fpga_Node_t { return p.pOutputs;   }
func Fpga_ManReadConst1(p *Fpga_Man_t) *Fpga_Node_t { return p.pConst1;    }
func Fpga_ManReadInputArrivals(p *Fpga_Man_t) *float32 { return p.pInputArrivals;}
func Fpga_ManReadVerbose(p *Fpga_Man_t) int32 { return p.fVerbose;   }
func Fpga_ManReadVarMax(p *Fpga_Man_t) int32 { return p.pLutLib.LutMax;     }
func Fpga_ManReadLutAreas(p *Fpga_Man_t) *float32 { return p.pLutLib.pLutAreas;  }
func Fpga_ManReadMapping(p *Fpga_Man_t) *Fpga_NodeVec_t { return p.vMapping;   }
func Fpga_ManSetOutputNames(p *Fpga_Man_t, ppNames *string) { p.ppOutputNames = ppNames; }
func Fpga_ManSetInputArrivals(p *Fpga_Man_t, pArrivals *float32) { p.pInputArrivals = pArrivals; }
func Fpga_ManSetAreaRecovery(p *Fpga_Man_t, fAreaRecovery int32) { p.fAreaRecovery = fAreaRecovery;}
func Fpga_ManSetDelayLimit(p *Fpga_Man_t, DelayLimit float32) { p.DelayLimit   = DelayLimit;    }
func Fpga_ManSetAreaLimit(p *Fpga_Man_t, AreaLimit float32) { p.AreaLimit    = AreaLimit;     }
func Fpga_ManSetChoiceNodeNum(p *Fpga_Man_t, nChoiceNodes int32) { p.nChoiceNodes = nChoiceNodes;  }  
func Fpga_ManSetChoiceNum(p *Fpga_Man_t, nChoices int32) { p.nChoices = nChoices;          }   
func Fpga_ManSetVerbose(p *Fpga_Man_t, fVerbose int32) { p.fVerbose = fVerbose;          }   
func Fpga_ManSetSwitching(p *Fpga_Man_t, fSwitching int32) { p.fSwitching = fSwitching;      }   
func Fpga_ManSetLatchPaths(p *Fpga_Man_t, fLatchPaths int32) { p.fLatchPaths = fLatchPaths;    }   
func Fpga_ManSetLatchNum(p *Fpga_Man_t, nLatches int32) { p.nLatches = nLatches;          }   
func Fpga_ManSetDelayTarget(p *Fpga_Man_t, DelayTarget float32) { p.DelayTarget = DelayTarget;    }   
func Fpga_ManSetName(p *Fpga_Man_t, pFileName string) { p.pFileName = pFileName;        }   

/**Function*************************************************************

  Synopsis    [Reads the parameters of the LUT library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_LibReadLutMax(pLib *Fpga_LutLib_t) int32 { return pLib.LutMax; }

/**Function*************************************************************

  Synopsis    [Reads parameters of the mapping node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeReadData0(p *Fpga_Node_t) string { return p.pData0;    }
func Fpga_NodeReadData1(p *Fpga_Node_t) *Fpga_Node_t { return p.pLevel;    }
func Fpga_NodeReadRefs(p *Fpga_Node_t) int32 { return p.nRefs;     }
func Fpga_NodeReadNum(p *Fpga_Node_t) int32 { return p.Num;       }
// inlining of macro Fpga_Regular
func Fpga_NodeReadLevel(p *Fpga_Node_t) int32 { return ((Fpga_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) .Level;  }
func Fpga_NodeReadCuts(p *Fpga_Node_t) *Fpga_Cut_t { return p.pCuts;     }
func Fpga_NodeReadCutBest(p *Fpga_Node_t) *Fpga_Cut_t { return p.pCutBest;  }
func Fpga_NodeReadOne(p *Fpga_Node_t) *Fpga_Node_t { return p.p1;        }
func Fpga_NodeReadTwo(p *Fpga_Node_t) *Fpga_Node_t { return p.p2;        }
func Fpga_NodeSetData0(p *Fpga_Node_t, pData string) { p.pData0 = pData;  }
func Fpga_NodeSetData1(p *Fpga_Node_t, pNode *Fpga_Node_t) { p.pLevel = pNode;  }
func Fpga_NodeSetNextE(p *Fpga_Node_t, pNextE *Fpga_Node_t) { p.pNextE = pNextE; }
func Fpga_NodeSetRepr(p *Fpga_Node_t, pRepr *Fpga_Node_t) { p.pRepr = pRepr;   }
func Fpga_NodeSetSwitching(p *Fpga_Node_t, Switching float32) { p.Switching = Switching;   }

/**Function*************************************************************

  Synopsis    [Checks the type of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// inlining of macro Fpga_Regular
func Fpga_NodeIsConst(p *Fpga_Node_t) int32 {  return (((Fpga_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).Num == -1;    }
// inlining of macro Fpga_Regular
func Fpga_NodeIsVar(p *Fpga_Node_t) int32 {  return (((Fpga_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).p1 == nil && (((Fpga_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).Num >= 0; }
// inlining of macro Fpga_Regular
func Fpga_NodeIsAnd(p *Fpga_Node_t) int32 {  return (((Fpga_Node_t *)((ABC_PTRUINT_T)(p) & ~01)) ).p1 != nil;  }
// inlining of macro Fpga_IsComplement
func Fpga_NodeComparePhase(p1 *Fpga_Node_t, p2 *Fpga_Node_t) int32 { assert(!(((int32(((ABC_PTRUINT_T) (p1) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (p1) & 01))))"); assert(!(((int32(((ABC_PTRUINT_T) (p2) & 01)) != 0))), "!/*__starting_point_inline_macro:Fpga_IsComplement__*/((int32(((ABC_PTRUINT_T) (p2) & 01))))"); return p1.fInv ^ p2.fInv; }

/**Function*************************************************************

  Synopsis    [Reads parameters from the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_CutReadLeavesNum(p *Fpga_Cut_t) int32 {  return p.nLeaves;  }
func Fpga_CutReadLeaves(p *Fpga_Cut_t) **Fpga_Node_t {  return p.ppLeaves; }


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
func Fpga_ManCreate(nInputs int32, nOutputs int32, fVerbose int32) *Fpga_Man_t {
    var p *Fpga_Man_t;
    var i int32;

    // start the manager
    // inlining of macro ABC_ALLOC
    p = ((Fpga_Man_t *) malloc(sizeof(Fpga_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Fpga_Man_t) );
    p.pLutLib   = (Fpga_LutLib_t *)Abc_FrameReadLibLut();
    p.nVarsMax  = p.pLutLib.LutMax;
    p.fVerbose  = fVerbose;
    p.fAreaRecovery = 1;
    p.fEpsilon  = float32(0.001);

    Fpga_TableCreate( p );
//if ( p->fVerbose )
//    printf( "Node = %d (%d) bytes. Cut = %d bytes.\n", sizeof(Fpga_Node_t), FPGA_NUM_BYTES(sizeof(Fpga_Node_t)), sizeof(Fpga_Cut_t) ); 
    // inlining of macro FPGA_NUM_BYTES
    p.mmNodes  = Extra_MmFixedStart( (((sizeof(Fpga_Node_t))/16 + (((sizeof(Fpga_Node_t))%16) > 0))*16) );
    p.mmCuts   = Extra_MmFixedStart( sizeof(Fpga_Cut_t) );

    assert(p.nVarsMax > 0, "p.nVarsMax > 0");
//    Fpga_MappingSetupTruthTables( p->uTruths );

    // make sure the constant node will get index -1
    p.nNodes = -1;
    // create the constant node
    p.pConst1 = Fpga_NodeCreate( p, nil, nil );
    p.vNodesAll = Fpga_NodeVecAlloc( 1000 );
    p.vMapping = Fpga_NodeVecAlloc( 1000 );

    // create the PI nodes
    p.nInputs = nInputs;
    // inlining of macro ABC_ALLOC
    p.pInputs = ((Fpga_Node_t * *) malloc(sizeof(Fpga_Node_t *) * (size_t)(nInputs)));
    for  i = 0; i < nInputs; i++  {
        p.pInputs[i] = Fpga_NodeCreate( p, nil, nil );
    }

    // create the place for the output nodes
    p.nOutputs = nOutputs;
    // inlining of macro ABC_ALLOC
    p.pOutputs = ((Fpga_Node_t * *) malloc(sizeof(Fpga_Node_t *) * (size_t)(nOutputs)));
    memset( p.pOutputs, 0, sizeof(Fpga_Node_t *) * nOutputs );
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates the mapping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_ManFree(p *Fpga_Man_t) {
//    Fpga_ManStats( p );
//    int i;
//    for ( i = 0; i < p->vNodesAll->nSize; i++ )
//        Fpga_NodeVecFree( p->vNodesAll->pArray[i]->vFanouts );
//    Fpga_NodeVecFree( p->pConst1->vFanouts );
    if p.vMapping {
        Fpga_NodeVecFree( p.vMapping ); }
    if p.vAnds {    
        Fpga_NodeVecFree( p.vAnds ); }
    if p.vNodesAll {    
        Fpga_NodeVecFree( p.vNodesAll ); }
    Extra_MmFixedStop( p.mmNodes );
    Extra_MmFixedStop( p.mmCuts );
    // inlining of macro ABC_FREE
    ((p.ppOutputNames) ? (free((char *) (p.ppOutputNames)), (p.ppOutputNames) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pInputArrivals) ? (free((char *) (p.pInputArrivals)), (p.pInputArrivals) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pInputs) ? (free((char *) (p.pInputs)), (p.pInputs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pOutputs) ? (free((char *) (p.pOutputs)), (p.pOutputs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pBins) ? (free((char *) (p.pBins)), (p.pBins) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Prints runtime statistics of the mapping manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_ManPrintTimeStats(p *Fpga_Man_t) {
//    extern char * pNetName;
//    extern int TotalLuts;
//    FILE * pTable;

    
/*
    pTable = fopen( "stats.txt", "a+" );
    fprintf( pTable, "%s ", pNetName );
    fprintf( pTable, "%.0f ", p->fRequiredGlo );
//    fprintf( pTable, "%.0f ", p->fAreaGlo );//+ (float)nOutputInvs );
    fprintf( pTable, "%.0f ", (float)TotalLuts );
    fprintf( pTable, "%4.2f\n", (float)(p->timeTotal-p->timeToMap)/(float)(CLOCKS_PER_SEC) );
    fclose( pTable );
*/

//    printf( "N-canonical = %d. Matchings = %d.  ", p->nCanons, p->nMatches );
//    printf( "Choice nodes = %d. Choices = %d.\n", p->nChoiceNodes, p->nChoices );
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("ToMap"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeToMap)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Cuts "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCuts)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Match"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMatch)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Area "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRecover)))/(float64(int64(time.Second))));;
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
}

/**Function*************************************************************

  Synopsis    [Creates a new node.]

  Description [This procedure should be called to create the constant
  node and the PI nodes first.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeCreate(p *Fpga_Man_t, p1 *Fpga_Node_t, p2 *Fpga_Node_t) *Fpga_Node_t {
    var pNode *Fpga_Node_t;
    // create the node
    pNode = (Fpga_Node_t *)Extra_MmFixedEntryFetch( p.mmNodes );
    memset( pNode, 0, sizeof(Fpga_Node_t) );
    // set very large required time
    pNode.tRequired = FPGA_FLOAT_LARGE;
    pNode.aEstFanouts = -1;
    pNode.p1  = p1; 
    pNode.p2  = p2;
    // set the number of this node
    pNode.Num = p.nNodes++;
    // place to store the fanouts
//    pNode->vFanouts = Fpga_NodeVecAlloc( 5 );
    // store this node in the internal array
    if pNode.Num >= 0 {
        Fpga_NodeVecPush( p.vNodesAll, pNode );
    } else {
        pNode.fInv = 1; }
    // set the level of this node
    if p1 != nil {
#ifdef FPGA_ALLOCATE_FANOUT
        // create the fanout info
        Fpga_NodeAddFaninFanout( Fpga_Regular(p1), pNode );
        Fpga_NodeAddFaninFanout( Fpga_Regular(p2), pNode );
#endif
        // compute the level
        // inlining of macro FPGA_MAX
        // inlining of macro Fpga_Regular
        pNode.Level = 1 + (((((Fpga_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .Level) > (((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .Level))? (((Fpga_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .Level) : (((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .Level));
        // inlining of macro Fpga_NodeIsSimComplement
        // inlining of macro Fpga_IsComplement
        // inlining of macro Fpga_Regular
        pNode.fInv  = (((int32(((ABC_PTRUINT_T) (p1) & 01))))? !(((Fpga_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .fInv) : (p1).fInv) & (((int32(((ABC_PTRUINT_T) (p2) & 01))))? !(((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .fInv) : (p2).fInv);
    }
    // reference the inputs 
    // inlining of macro Fpga_NodeRef
    // inlining of macro Fpga_Regular
    if p1 != nil { ((((Fpga_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) ).nRefs++); }
    // inlining of macro Fpga_NodeRef
    // inlining of macro Fpga_Regular
    if p2 != nil { ((((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) ).nRefs++); }
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Create the unique table of AND gates.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_TableCreate(pMan *Fpga_Man_t) {
    assert(pMan.pBins == nil, "pMan.pBins == nil");
    pMan.nBins = Abc_PrimeCudd(50000);
    // inlining of macro ABC_ALLOC
    pMan.pBins = ((Fpga_Node_t * *) malloc(sizeof(Fpga_Node_t *) * (size_t)(pMan.nBins)));
    memset( pMan.pBins, 0, sizeof(Fpga_Node_t *) * pMan.nBins );
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
func Fpga_TableLookup(pMan *Fpga_Man_t, p1 *Fpga_Node_t, p2 *Fpga_Node_t) *Fpga_Node_t {
    var pEnt *Fpga_Node_t;
    var Key uint32;

    if p1 == p2 {
        return p1; }
    // inlining of macro Fpga_Not
    if p1 == ((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) ^ 01)) {
        // inlining of macro Fpga_Not
        return ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pConst1) ^ 01)) ; }
    if Fpga_NodeIsConst(p1) != 0 {
        if p1 == pMan.pConst1 {
            return p2; }
        // inlining of macro Fpga_Not
        return ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pConst1) ^ 01)) ;
    }
    if Fpga_NodeIsConst(p2) != 0 {
        if p2 == pMan.pConst1 {
            return p1; }
        // inlining of macro Fpga_Not
        return ((Fpga_Node_t *)((ABC_PTRUINT_T)(pMan.pConst1) ^ 01)) ;
    }

    // inlining of macro Fpga_Regular
    if  ((Fpga_Node_t *)((ABC_PTRUINT_T)(p1) & ~01)) .Num > ((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) & ~01)) .Num {
        pEnt = p1;
p1 = p2;
p2 = pEnt; }

    Key = Fpga_HashKey2( p1, p2, pMan.nBins );
    for  pEnt = pMan.pBins[Key]; (pEnt != nil); pEnt = pEnt.pNext  {
        if pEnt.p1 == p1 && pEnt.p2 == p2 {
            return pEnt; }
    }
    // resize the table
    if pMan.nNodes >= 2 * pMan.nBins {
        Fpga_TableResize( pMan );
        Key = Fpga_HashKey2( p1, p2, pMan.nBins );
    }
    // create the new node
    pEnt = Fpga_NodeCreate( pMan, p1, p2 );
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
func Fpga_TableResize(pMan *Fpga_Man_t) {
    var pBinsNew **Fpga_Node_t;
    var pEnt, pEnt2 *Fpga_Node_t;
    var nBinsNew, Counter, i int32;
    var clk int64;
    var Key uint32;

clk = time.Now().UnixNano();
    // get the new table size
    nBinsNew = Abc_PrimeCudd(2 * pMan.nBins); 
    // allocate a new array
    // inlining of macro ABC_ALLOC
    pBinsNew = ((Fpga_Node_t * *) malloc(sizeof(Fpga_Node_t *) * (size_t)(nBinsNew)));
    memset( pBinsNew, 0, sizeof(Fpga_Node_t *) * nBinsNew );
    // rehash the entries from the old table
    Counter = 0;
    for  i = 0; i < pMan.nBins; i++  {
        for  pEnt = pMan.pBins[i], pEnt2 = pEnt? pEnt.pNext: nil; (pEnt != nil); 
              pEnt = pEnt2, pEnt2 = pEnt? pEnt.pNext: nil  {
            Key = Fpga_HashKey2( pEnt.p1, pEnt.p2, nBinsNew );
            pEnt.pNext = pBinsNew[Key];
            pBinsNew[Key] = pEnt;
            Counter++;
        }
    }
    assert(Counter == pMan.nNodes - pMan.nInputs, "Counter == pMan.nNodes - pMan.nInputs");
    if pMan.fVerbose {
//        printf( "Increasing the unique table size from %6d to %6d. ", pMan->nBins, nBinsNew );
//        ABC_PRT( "Time", clock() - clk );
    }
    // replace the table and the parameters
    // inlining of macro ABC_FREE
    ((pMan.pBins) ? (free((char *) (pMan.pBins)), (pMan.pBins) = 0) : 0);
    pMan.pBins = pBinsNew;
    pMan.nBins = nBinsNew;
}



/**Function*************************************************************

  Synopsis    [Elementary AND operation on the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeAnd(p *Fpga_Man_t, p1 *Fpga_Node_t, p2 *Fpga_Node_t) *Fpga_Node_t {
    var pNode *Fpga_Node_t;
    pNode = Fpga_TableLookup( p, p1, p2 );     
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Elementary OR operation on the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeOr(p *Fpga_Man_t, p1 *Fpga_Node_t, p2 *Fpga_Node_t) *Fpga_Node_t {
    var pNode *Fpga_Node_t;
    // inlining of macro Fpga_Not
    pNode = ((Fpga_Node_t *)((ABC_PTRUINT_T)(Fpga_TableLookup( p, ((Fpga_Node_t *)((ABC_PTRUINT_T)(p1) ^ 01)) , ((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) ^ 01))  )) ^ 01)) ;  
    return pNode;
}

/**Function*************************************************************

  Synopsis    [Elementary EXOR operation on the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeExor(p *Fpga_Man_t, p1 *Fpga_Node_t, p2 *Fpga_Node_t) *Fpga_Node_t {
    // inlining of macro Fpga_Not
    return Fpga_NodeMux( p, p1, ((Fpga_Node_t *)((ABC_PTRUINT_T)(p2) ^ 01)) , p2 );
}

/**Function*************************************************************

  Synopsis    [Elementary MUX operation on the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeMux(p *Fpga_Man_t, pC *Fpga_Node_t, pT *Fpga_Node_t, pE *Fpga_Node_t) *Fpga_Node_t {
    var pAnd1, pAnd2, pRes *Fpga_Node_t;
    pAnd1 = Fpga_TableLookup( p, pC,           pT ); 
    // inlining of macro Fpga_Not
    pAnd2 = Fpga_TableLookup( p, ((Fpga_Node_t *)((ABC_PTRUINT_T)(pC) ^ 01)) , pE ); 
    pRes  = Fpga_NodeOr( p, pAnd1, pAnd2 );           
    return pRes;
}


/**Function*************************************************************

  Synopsis    [Sets the node to be equivalent to the given one.]

  Description [This procedure is a work-around for the equivalence check.
  Does not verify the equivalence. Use at the user's risk.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_NodeSetChoice(pMan *Fpga_Man_t, pNodeOld *Fpga_Node_t, pNodeNew *Fpga_Node_t) {
    pNodeNew.pNextE = pNodeOld.pNextE;
    pNodeOld.pNextE = pNodeNew;
    pNodeNew.pRepr  = pNodeOld;
}


    
/**Function*************************************************************

  Synopsis    [Prints some interesting stats.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_ManStats(p *Fpga_Man_t) {
    var pTable *os.File;
    pTable = os.OpenFile("stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    fmt.Fprintf(pTable, "%s ", p.pFileName);
    fmt.Fprintf(pTable, "%4d ", p.nInputs - p.nLatches);
    fmt.Fprintf(pTable, "%4d ", p.nOutputs - p.nLatches);
    fmt.Fprintf(pTable, "%4d ", p.nLatches);
    fmt.Fprintf(pTable, "%7d ", p.vAnds.nSize);
    fmt.Fprintf(pTable, "%7d ", Fpga_CutCountAll(p));
    fmt.Fprintf(pTable, "%2d\n", int32(p.fRequiredGlo));
    _ = (pTable).Close();
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

