package mapper

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [mapperCore.c]

  PackageName [MVSIS 1.3: Multi-valued logic synthesis system.]

  Synopsis    [Generic technology mapping engine.]

  Author      [MVSIS Group]

  Affiliation [UC Berkeley]

  Date        [Ver. 2.0. Started - June 1, 2004.]

  Revision    [$Id: mapperCore.c,v 1.7 2004/10/01 23:41:04 satrajit Exp $]

***********************************************************************/

//#include "mapperInt.h"
//#include "resm.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Performs technology mapping for the given object graph.]

	Description [The object graph is stored in the mapping manager.
	First, the AND nodes that fanout into POs are collected in the DFS order.
	Two preprocessing steps are performed: the k-feasible cuts are computed
	for each node and the truth tables are computed for each cut. Next, the
	delay-optimal matches are assigned for each node, followed by several
	iterations of area recoveryd: using area flow (global optimization)
	and using exact area at a node (local optimization).]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Map_Mapping(p *Map_Man_t) int32 {
	var fShowSwitching int32
	var fUseAreaFlow int32 = 1
	var fUseExactArea int32 = boolToInt(p.fSwitching == 0)
	var fUseExactAreaWithPhase int32 = boolToInt(p.fSwitching == 0)
	var clk abctime

	//////////////////////////////////////////////////////////////////////
	// perform pre-mapping computations
	if p.fVerbose != 0 {
		Map_MappingReportChoices(p)
	}
	Map_MappingSetChoiceLevels(p) // should always be called before mapping!
	//    return 1;

	// compute the cuts of nodes in the DFS order
	clk = Abc_Clock()
	Map_MappingCuts(p)
	p.timeCuts = Abc_Clock() - clk
	// derive the truth tables
	clk = Abc_Clock()
	Map_MappingTruths(p)
	p.timeTruth = Abc_Clock() - clk
	//////////////////////////////////////////////////////////////////////
	//ABC_PRT( "Truths", Abc_Clock() - clk );

	//////////////////////////////////////////////////////////////////////
	// compute the minimum-delay mapping
	clk = Abc_Clock()
	p.fMappingMode = 0
	if Map_MappingMatches(p) == 0 {
		return 0
	}
	p.timeMatch = Abc_Clock() - clk
	// compute the references and collect the nodes used in the mapping
	Map_MappingSetRefs(p)
	p.AreaBase = Map_MappingGetArea(p)
	if p.fVerbose != 0 {
		fmt.Printf("Delay    : %s = %8.2f  Flow = %11.1f  Area = %11.1f  %4.1f %%   ", mapMetricLabel(fShowSwitching), mapMetricValue(p, fShowSwitching), Map_MappingGetAreaFlow(p), p.AreaBase, 0.0)
		// inlining of macro ABC_PRT
		Abc_Print(1, "%s =", ("Time"))
		Abc_Print(1, "%9.2f sec\n", 1.0*(float64((p.timeMatch)))/(float64(int64(time.Second))))
	}
	//////////////////////////////////////////////////////////////////////

	if p.fAreaRecovery == 0 {
		if p.fVerbose != 0 {
			Map_MappingPrintOutputArrivals(p)
		}
		return 1
	}

	//////////////////////////////////////////////////////////////////////
	// perform area recovery using area flow
	clk = Abc_Clock()
	if fUseAreaFlow != 0 {
		// compute the required times
		Map_TimeComputeRequiredGlobal(p)
		// recover area flow
		p.fMappingMode = 1
		Map_MappingMatches(p)
		// compute the references and collect the nodes used in the mapping
		Map_MappingSetRefs(p)
		p.AreaFinal = Map_MappingGetArea(p)
		if p.fVerbose != 0 {
			fmt.Printf("AreaFlow : %s = %8.2f  Flow = %11.1f  Area = %11.1f  %4.1f %%   ", mapMetricLabel(fShowSwitching), mapMetricValue(p, fShowSwitching), Map_MappingGetAreaFlow(p), p.AreaFinal, 100.0*(p.AreaBase-p.AreaFinal)/p.AreaBase)
			// inlining of macro ABC_PRT
			Abc_Print(1, "%s =", ("Time"))
			Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))))
		}
	}
	p.timeArea += Abc_Clock() - clk
	//////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////
	// perform area recovery using exact area
	clk = Abc_Clock()
	if fUseExactArea != 0 {
		// compute the required times
		Map_TimeComputeRequiredGlobal(p)
		// recover area
		p.fMappingMode = 2
		Map_MappingMatches(p)
		// compute the references and collect the nodes used in the mapping
		Map_MappingSetRefs(p)
		p.AreaFinal = Map_MappingGetArea(p)
		if p.fVerbose != 0 {
			fmt.Printf("Area     : %s = %8.2f  Flow = %11.1f  Area = %11.1f  %4.1f %%   ", mapMetricLabel(fShowSwitching), mapMetricValue(p, fShowSwitching), 0.0, p.AreaFinal, 100.0*(p.AreaBase-p.AreaFinal)/p.AreaBase)
			// inlining of macro ABC_PRT
			Abc_Print(1, "%s =", ("Time"))
			Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))))
		}
	}
	p.timeArea += Abc_Clock() - clk
	//////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////
	// perform area recovery using exact area
	clk = Abc_Clock()
	if fUseExactAreaWithPhase != 0 {
		// compute the required times
		Map_TimeComputeRequiredGlobal(p)
		// recover area
		p.fMappingMode = 3
		Map_MappingMatches(p)
		// compute the references and collect the nodes used in the mapping
		Map_MappingSetRefs(p)
		p.AreaFinal = Map_MappingGetArea(p)
		if p.fVerbose != 0 {
			fmt.Printf("Area     : %s = %8.2f  Flow = %11.1f  Area = %11.1f  %4.1f %%   ", mapMetricLabel(fShowSwitching), mapMetricValue(p, fShowSwitching), 0.0, p.AreaFinal, 100.0*(p.AreaBase-p.AreaFinal)/p.AreaBase)
			// inlining of macro ABC_PRT
			Abc_Print(1, "%s =", ("Time"))
			Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))))
		}
	}
	p.timeArea += Abc_Clock() - clk
	//////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////
	// perform area recovery using exact area
	clk = Abc_Clock()
	if p.fSwitching != 0 {
		// compute the required times
		Map_TimeComputeRequiredGlobal(p)
		// recover switching activity
		p.fMappingMode = 4
		Map_MappingMatches(p)
		// compute the references and collect the nodes used in the mapping
		Map_MappingSetRefs(p)
		p.AreaFinal = Map_MappingGetArea(p)
		if p.fVerbose != 0 {
			fmt.Printf("Switching: %s = %8.2f  Flow = %11.1f  Area = %11.1f  %4.1f %%   ", mapMetricLabel(fShowSwitching), mapMetricValue(p, fShowSwitching), 0.0, p.AreaFinal, 100.0*(p.AreaBase-p.AreaFinal)/p.AreaBase)
			// inlining of macro ABC_PRT
			Abc_Print(1, "%s =", ("Time"))
			Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))))
		}

		// compute the required times
		Map_TimeComputeRequiredGlobal(p)
		// recover switching activity
		p.fMappingMode = 4
		Map_MappingMatches(p)
		// compute the references and collect the nodes used in the mapping
		Map_MappingSetRefs(p)
		p.AreaFinal = Map_MappingGetArea(p)
		if p.fVerbose != 0 {
			fmt.Printf("Switching: %s = %8.2f  Flow = %11.1f  Area = %11.1f  %4.1f %%   ", mapMetricLabel(fShowSwitching), mapMetricValue(p, fShowSwitching), 0.0, p.AreaFinal, 100.0*(p.AreaBase-p.AreaFinal)/p.AreaBase)
			// inlining of macro ABC_PRT
			Abc_Print(1, "%s =", ("Time"))
			Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))))
		}
	}
	p.timeArea += Abc_Clock() - clk
	//////////////////////////////////////////////////////////////////////

	// print the arrival times of the latest outputs
	if p.fVerbose != 0 {
		Map_MappingPrintOutputArrivals(p)
	}
	return 1
}

//ABC_NAMESPACE_IMPL_END

func mapMetricLabel(switching int32) string {
	if switching != 0 {
		return "Switch"
	}
	return "Delay"
}

func mapMetricValue(p *Map_Man_t, switching int32) float32 {
	if switching != 0 {
		return Map_MappingGetSwitching(p)
	}
	return p.fRequiredGlo
}
