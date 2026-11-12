package glucose

/**CFile****************************************************************

  FileName    [AbcGlucose.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT solver Glucose 3.0 by Gilles Audemard and Laurent Simon.]

  Synopsis    [Interface to Glucose.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 6, 2017.]

  Revision    [$Id: AbcGlucose.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

import "github.com/glycerine/abcgo/aig/aig"

const (
	GLUCOSE_UNSAT int32 = -1
	GLUCOSE_SAT   int32 = 1
	GLUCOSE_UNDEC int32 = 0
)

type Glucose_Pars struct{ Pre, Verb, Cust, NConfls int32 }

func Glucose_CreatePars(pre, verbosity, custom, conflicts int32) Glucose_Pars {
	return Glucose_Pars{pre, verbosity, custom, conflicts}
}

type Bmcg_sat_solver = Solver

// The Go bridge uses the committed graph manager. GIA clients can convert their
// graph at the boundary; edge phases follow AIG's constant-one convention.
type Gia_Man_t = aig.Graph
