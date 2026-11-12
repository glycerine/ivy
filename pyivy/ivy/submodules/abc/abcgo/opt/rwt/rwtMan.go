package rwt

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [rwtMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Rewriting manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwtMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwt.h"
//#include "bool/deco/deco.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

static unsigned short * s_puCanons = NULL; 
var s_pPhases string = nil; 
var s_pPerms string = nil; 
static unsigned char *  s_pMap = NULL;

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts residual rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManGlobalStart() { 
    if s_puCanons == nil {
        Extra_Truth4VarNPN( &s_puCanons, &s_pPhases, &s_pPerms, &s_pMap ); }
}

/**Function*************************************************************

  Synopsis    [Starts residual rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManGlobalStop() { 
    // inlining of macro ABC_FREE
    ((s_puCanons) ? (free((char *) (s_puCanons)), (s_puCanons) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s_pPhases) ? (free((char *) (s_pPhases)), (s_pPhases) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s_pPerms) ? (free((char *) (s_pPerms)), (s_pPerms) = 0) : 0);
    // inlining of macro ABC_FREE
    ((s_pMap) ? (free((char *) (s_pMap)), (s_pMap) = 0) : 0);
} 

/**Function*************************************************************

  Synopsis    [Starts rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManStart(fPrecompute int32) *Rwt_Man_t {
    var p *Rwt_Man_t;
    var clk abctime  = Abc_Clock();
clk = Abc_Clock();
    // inlining of macro ABC_ALLOC
    p = ((Rwt_Man_t *) malloc(sizeof(Rwt_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Rwt_Man_t) );
    p.nFuncs = (1<<16);
    // copy the global tables
    Rwt_ManGlobalStart();
    p.puCanons = s_puCanons; 
    p.pPhases  = s_pPhases; 
    p.pPerms   = s_pPerms; 
    p.pMap     = s_pMap; 
    // initialize practical NPN classes
    p.pPractical  = Rwt_ManGetPractical( p );
    // create the table
    // inlining of macro ABC_ALLOC
    p.pTable = ((Rwt_Node_t * *) malloc(sizeof(Rwt_Node_t *) * (size_t)(p.nFuncs)));
    memset( p.pTable, 0, sizeof(Rwt_Node_t *) * p.nFuncs );
    // create the elementary nodes
    p.pMmNode  = Mem_FixedStart( sizeof(Rwt_Node_t) );
    p.vForest  = Vec_PtrAlloc( 100 );
    Rwt_ManAddVar( p, 0x0000, fPrecompute ); // constant 0
    Rwt_ManAddVar( p, 0xAAAA, fPrecompute ); // var A
    Rwt_ManAddVar( p, 0xCCCC, fPrecompute ); // var B
    Rwt_ManAddVar( p, 0xF0F0, fPrecompute ); // var C
    Rwt_ManAddVar( p, 0xFF00, fPrecompute ); // var D
    p.nClasses = 5;
    // other stuff
    p.nTravIds   = 1;
    p.pPerms4    = Extra_Permutations( 4 );
    p.vLevNums   = Vec_IntAlloc( 50 );
    p.vFanins    = Vec_PtrAlloc( 50 );
    p.vFaninsCur = Vec_PtrAlloc( 50 );
    p.vNodesTemp = Vec_PtrAlloc( 50 );
    if fPrecompute != 0 {   // precompute subgraphs
//        Rwt_ManPrecompute( p );
//        Rwt_ManPrint( p );
//        Rwt_ManWriteToArray( p );
    
    } else {   // load saved subgraphs
        Rwt_ManLoadFromArray( p, 0 );
//        Rwt_ManPrint( p );
        Rwt_ManPreprocess( p );
    }
p.timeStart = Abc_Clock() - clk;
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManStop(p *Rwt_Man_t) {
    if p.vClasses {
        var pNode *Rwt_Node_t;
        var i, k int32;
        // inlining of macro Vec_VecForEachEntry
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; i < Vec_VecSize(p.vClasses); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(p.vClasses, i))) && (((pNode) = (Rwt_Node_t *)Vec_PtrEntry(Vec_VecEntry(p.vClasses, i), k)), 1); k++  { 
            Dec_GraphFree( (Dec_Graph_t *)pNode.pNext );
        }
        }
    }
    if p.vClasses {  Vec_VecFree( p.vClasses ); }
    Vec_PtrFree( p.vNodesTemp );
    Vec_PtrFree( p.vForest );
    Vec_IntFree( p.vLevNums );
    Vec_PtrFree( p.vFanins );
    Vec_PtrFree( p.vFaninsCur );
    Mem_FixedStop( p.pMmNode, 0 );
    // inlining of macro ABC_FREE
    ((p.pMapInv) ? (free((char *) (p.pMapInv)), (p.pMapInv) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTable) ? (free((char *) (p.pTable)), (p.pTable) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPractical) ? (free((char *) (p.pPractical)), (p.pPractical) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPerms4) ? (free((char *) (p.pPerms4)), (p.pPerms4) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManPrintStats(p *Rwt_Man_t) {
    var i, Counter int32;
    for  i = 0; i < 222; i++  {
        Counter += (p.nScores[i] > 0);
    }

    fmt.Printf("Rewriting statistics:\n");
    fmt.Printf("Total cuts tries  = %8d.\n", p.nCutsGood);
    fmt.Printf("Bad cuts found    = %8d.\n", p.nCutsBad);
    fmt.Printf("Total subgraphs   = %8d.\n", p.nSubgraphs);
    fmt.Printf("Used NPN classes  = %8d.\n", Counter);
    fmt.Printf("Nodes considered  = %8d.\n", p.nNodesConsidered);
    fmt.Printf("Nodes rewritten   = %8d.\n", p.nNodesRewritten);
    fmt.Printf("Calculated gain   = %8d.\n", p.nNodesGained);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Start       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeStart)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Cuts        "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCut)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Truth       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTruth)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Resynthesis "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeRes)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("    Mffc    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMffc)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("    Eval    "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeEval)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Update      "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeUpdate)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("TOTAL       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeTotal)))/(float64(int64(time.Second))));;

/*
    printf( "The scores are:\n" );
    for ( i = 0; i < 222; i++ )
        if ( p->nScores[i] > 0 )
        {
            extern void Ivy_TruthDsdComputePrint( unsigned uTruth );
            printf( "%3d = %8d  canon = %5d  ", i, p->nScores[i], p->pMapInv[i] );
            Ivy_TruthDsdComputePrint( (unsigned)p->pMapInv[i] | ((unsigned)p->pMapInv[i] << 16) );
        }
    printf( "\n" );
*/
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManPrintStatsFile(p *Rwt_Man_t) {
    var pTable *os.File;
    pTable = os.OpenFile("stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666);
    fmt.Fprintf(pTable, "%d ", p.nCutsGood);
    fmt.Fprintf(pTable, "%d ", p.nSubgraphs);
    fmt.Fprintf(pTable, "%d ", p.nNodesRewritten);
    fmt.Fprintf(pTable, "%d", p.nNodesGained);
    fmt.Fprintf(pTable, "\n");
    _ = (pTable).Close();
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManReadDecs(p *Rwt_Man_t) any {
    return p.pGraph;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManReadLeaves(p *Rwt_Man_t) *Vec_Ptr_t {
    return p.vFanins;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManReadCompl(p *Rwt_Man_t) int32 {
    return p.fCompl;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManAddTimeCuts(p *Rwt_Man_t, Time abctime) {
    p.timeCut += Time;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManAddTimeUpdate(p *Rwt_Man_t, Time abctime) {
    p.timeUpdate += Time;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_ManAddTimeTotal(p *Rwt_Man_t, Time abctime) {
    p.timeTotal += Time;
}


/**Function*************************************************************

  Synopsis    [Precomputes AIG subgraphs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwt_Precompute() {
    var p *Rwt_Man_t;
    p = Rwt_ManStart( 1 );
    Rwt_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

