package rwr

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [rwrMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Rewriting manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwrMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwr.h"
//#include "base/main/main.h"
//#include "bool/dec/dec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts rewriting manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManStart(fPrecompute int32) *Rwr_Man_t {
    var pManDec *Dec_Man_t;
    var p *Rwr_Man_t;
    var clk abctime  = Abc_Clock();
clk = Abc_Clock();
    // inlining of macro ABC_ALLOC
    p = ((Rwr_Man_t *) malloc(sizeof(Rwr_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Rwr_Man_t) );
    p.nFuncs = (1<<16);
    pManDec   = (Dec_Man_t *)Abc_FrameReadManDec();
    p.puCanons = pManDec.puCanons; 
    p.pPhases  = pManDec.pPhases; 
    p.pPerms   = pManDec.pPerms; 
    p.pMap     = pManDec.pMap; 
    // initialize practical NPN classes
    p.pPractical  = Rwr_ManGetPractical( p );
    // create the table
    // inlining of macro ABC_ALLOC
    p.pTable = ((Rwr_Node_t * *) malloc(sizeof(Rwr_Node_t *) * (size_t)(p.nFuncs)));
    memset( p.pTable, 0, sizeof(Rwr_Node_t *) * p.nFuncs );
    // create the elementary nodes
    p.pMmNode  = Extra_MmFixedStart( sizeof(Rwr_Node_t) );
    p.vForest  = Vec_PtrAlloc( 100 );
    Rwr_ManAddVar( p, 0x0000, fPrecompute ); // constant 0
    Rwr_ManAddVar( p, 0xAAAA, fPrecompute ); // var A
    Rwr_ManAddVar( p, 0xCCCC, fPrecompute ); // var B
    Rwr_ManAddVar( p, 0xF0F0, fPrecompute ); // var C
    Rwr_ManAddVar( p, 0xFF00, fPrecompute ); // var D
    p.nClasses = 5;
    // other stuff
    p.nTravIds   = 1;
    p.pPerms4    = Extra_Permutations( 4 );
    p.vLevNums   = Vec_IntAlloc( 50 );
    p.vFanins    = Vec_PtrAlloc( 50 );
    p.vFaninsCur = Vec_PtrAlloc( 50 );
    p.vNodesTemp = Vec_PtrAlloc( 50 );
    if fPrecompute != 0 {   // precompute subgraphs
        Rwr_ManPrecompute( p );
//        Rwr_ManPrint( p );
        Rwr_ManWriteToArray( p );
    
    } else {   // load saved subgraphs
        Rwr_ManLoadFromArray( p, 0 );
//        Rwr_ManPrint( p );
        Rwr_ManPreprocess( p );
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
func Rwr_ManStop(p *Rwr_Man_t) {
    if p.vClasses {
        var pNode *Rwr_Node_t;
        var i, k int32;
        // inlining of macro Vec_VecForEachEntry
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; i < Vec_VecSize(p.vClasses); i++  {                                                        for  k = 0; (k < Vec_PtrSize(Vec_VecEntry(p.vClasses, i))) && (((pNode) = (Rwr_Node_t *)Vec_PtrEntry(Vec_VecEntry(p.vClasses, i), k)), 1); k++  { 
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
    Extra_MmFixedStop( p.pMmNode );
    p.pMapInv = nil;
    p.pTable = nil;
    p.pPractical = nil;
    p.pPerms4 = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManPrintStats(p *Rwr_Man_t) {
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
    fmt.Printf("Gain              = %8d. (%6.2f %%).\n", p.nNodesBeg-p.nNodesEnd, 100.0*(p.nNodesBeg-p.nNodesEnd)/p.nNodesBeg);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Start       "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeStart)))/(float64(int64(time.Second))));;
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Cuts        "));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeCut)))/(float64(int64(time.Second))));;
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
*/
    fmt.Printf("\n");

}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManPrintStatsFile(p *Rwr_Man_t) {
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
func Rwr_ManReadDecs(p *Rwr_Man_t) any {
    return p.pGraph;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManReadLeaves(p *Rwr_Man_t) *Vec_Ptr_t {
    return p.vFanins;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManReadCompl(p *Rwr_Man_t) int32 {
    return p.fCompl;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManAddTimeCuts(p *Rwr_Man_t, Time abctime) {
    p.timeCut += Time;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManAddTimeUpdate(p *Rwr_Man_t, Time abctime) {
    p.timeUpdate += Time;
}

/**Function*************************************************************

  Synopsis    [Stops the resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManAddTimeTotal(p *Rwr_Man_t, Time abctime) {
    p.timeTotal += Time;
}


/**Function*************************************************************

  Synopsis    [Precomputes AIG subgraphs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_Precompute() {
    var p *Rwr_Man_t;
    p = Rwr_ManStart( 1 );
    Rwr_ManStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

