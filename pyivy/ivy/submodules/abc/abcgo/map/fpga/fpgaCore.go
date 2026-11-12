package fpga

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [fpgaCore.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Technology mapping for variable-size-LUT FPGAs.]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - August 18, 2004.]

  Revision    [$Id: fpgaCore.c,v 1.7 2004/10/01 23:41:04 satrajit Exp $]

***********************************************************************/

//#include "fpgaInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static int  Fpga_MappingPostProcess( Fpga_Man_t * p );


var s_MappingTime int64;
var s_MappingMem int32;


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs technology mapping for the given object graph.]

  Description [The object graph is stored in the mapping manager.
  First, all the AND-nodes, which fanout into the POs, are collected
  in the DFS fashion. Next, three steps are performed: the k-feasible
  cuts are computed for each node, the truth tables are computed for
  each cut, and the delay-optimal matches are assigned for each node.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_Mapping(p *Fpga_Man_t) int32 {
    var clk int64;
    var clkTotal int64 = time.Now().UnixNano();
 
    // collect the nodes reachable from POs in the DFS order (including the choices)
    p.vAnds = Fpga_MappingDfs( p, 1 );
    Fpga_ManReportChoices( p ); // recomputes levels
    Fpga_MappingSetChoiceLevels( p );

    // compute the cuts of nodes in the DFS order
    clk = time.Now().UnixNano();
    Fpga_MappingCuts( p );
    p.timeCuts = time.Now().UnixNano() - clk;

    // match the truth tables to the supergates
    clk = time.Now().UnixNano();
    if !Fpga_MappingMatches( p, 1 ) {
        return 0; }
    p.timeMatch = time.Now().UnixNano() - clk;

    // perform area recovery
    clk = time.Now().UnixNano();
    if !(Fpga_MappingPostProcess( p ) != 0) {
        return 0; }
    p.timeRecover = time.Now().UnixNano() - clk;
//ABC_PRT( "Total mapping time", clock() - clkTotal );

    s_MappingTime = time.Now().UnixNano() - clkTotal;
    s_MappingMem = Fpga_CutCountAll(p) * (sizeof(Fpga_Cut_t) - sizeof(int) * (FPGA_MAX_LEAVES - p.nVarsMax));

    // print the AI-graph used for mapping
    //Fpga_ManShow( p, "test" );
//    if ( p->fVerbose )
//        Fpga_MappingPrintOutputArrivals( p );
    if p.fVerbose {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clkTotal)))/(float64(int64(time.Second))));;
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Postprocesses the mapped network for area recovery.]

  Description [This procedure assumes that the mapping is assigned.
  It iterates the loop, in which the required times are computed and
  the mapping is updated. It is conceptually similar to the paper: 
  V. Manohararajah, S. D. Brown, Z. G. Vranesic, Heuristics for area 
  minimization in LUT-based FPGA technology mapping. Proc. IWLS '04.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Fpga_MappingPostProcess(p *Fpga_Man_t) int32 {
    var fShowSwitching int32;
    var fRecoverAreaFlow int32   = 1;
    var fRecoverArea int32       = 1;
    var aAreaTotalCur, aAreaTotalCur2 float32;
    var Iter int32;
    var clk int64;

//if ( p->fVerbose )
//    printf( "Best clock period = %5.2f\n", Fpga_TimeComputeArrivalMax(p) );

    // compute area, set references, and collect nodes used in the mapping
    Iter = 1;
    aAreaTotalCur = Fpga_MappingSetRefsAndArea( p );
if p.fVerbose {
fmt.Printf("Iteration %dD :  Area = %8.1f  ", Iter++, aAreaTotalCur);
if fShowSwitching != 0 {
fmt.Printf("Switch = %8.1f  ", Fpga_MappingGetSwitching(p,p.vMapping));
} else {
fmt.Printf("Delay = %5.2f  ", Fpga_TimeComputeArrivalMax(p)); }

// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMatch)))/(float64(int64(time.Second))));;
}

    if !p.fAreaRecovery {
        return 1; }

    if fRecoverAreaFlow != 0 {
clk = time.Now().UnixNano();
        // compute the required times and the fanouts
        Fpga_TimeComputeRequiredGlobal( p, 1 );
        // remap topologically
        Fpga_MappingMatches( p, 0 );
        // get the resulting area
//        aAreaTotalCur = Fpga_MappingSetRefsAndArea( p );
        aAreaTotalCur = Fpga_MappingAreaTrav( p );
        // note that here we do not update the reference counter
        // for some reason, this works better on benchmarks
if p.fVerbose {
fmt.Printf("Iteration %dF :  Area = %8.1f  ", Iter++, aAreaTotalCur);
if fShowSwitching != 0 {
fmt.Printf("Switch = %8.1f  ", Fpga_MappingGetSwitching(p,p.vMapping));
} else {
fmt.Printf("Delay = %5.2f  ", Fpga_TimeComputeArrivalMax(p)); }
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
}
    }

    // update reference counters
    aAreaTotalCur2 = Fpga_MappingSetRefsAndArea( p );
    assert(aAreaTotalCur == aAreaTotalCur2, "aAreaTotalCur == aAreaTotalCur2");

    if fRecoverArea != 0 {
clk = time.Now().UnixNano();
        // compute the required times and the fanouts
        Fpga_TimeComputeRequiredGlobal( p, 0 );
        // remap topologically
        if p.fSwitching {
            Fpga_MappingMatchesSwitch( p );
        } else {
            Fpga_MappingMatchesArea( p ); }
        // get the resulting area
        aAreaTotalCur = Fpga_MappingSetRefsAndArea( p );
if p.fVerbose {
fmt.Printf("Iteration %d%s :  Area = %8.1f  ", Iter++, (p.fSwitching?"S":"A"), aAreaTotalCur);
if fShowSwitching != 0 {
fmt.Printf("Switch = %8.1f  ", Fpga_MappingGetSwitching(p,p.vMapping));
} else {
fmt.Printf("Delay = %5.2f  ", Fpga_TimeComputeArrivalMax(p)); }
// inlining of macro ABC_PRT
Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
}
    }

    p.fAreaGlo = aAreaTotalCur;
    return 1;
}


//ABC_NAMESPACE_IMPL_END

