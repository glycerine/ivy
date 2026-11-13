package bsat

import (
	"github.com/glycerine/abcgo/aig/aig"
)

/**CFile****************************************************************

  FileName    [satInter.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT sat_solver.]

  Synopsis    [Interpolation package.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: satInter.c,v 1.4 2005/09/16 22:55:03 casem Exp $]

***********************************************************************/

type Inta_Man_t = Inta_Man_t_
type Inta_Man_t_ struct{ aigInterpolation }

func Inta_ManAlloc() *Inta_Man_t { return &Inta_Man_t{} }
func Inta_ManFree(p *Inta_Man_t) {
	if p != nil {
		*p = Inta_Man_t{}
	}
}
func Inta_ManError(p *Inta_Man_t) error { return p.lastErr }
func Inta_ManInterpolate(p *Inta_Man_t, cnf *Sto_Man_t, deadline abctime, globals []int32, verbose int32) *aig.Graph {
	return p.interpolate(cnf, globals, int64(deadline), false, verbose)
}

// Inta_ManInterpolateWithClock uses an absolute deadline in clock's domain.
func Inta_ManInterpolateWithClock(p *Inta_Man_t, cnf *Sto_Man_t, deadline int64, globals []int32, verbose int32, clock func() int64) *aig.Graph {
	return p.interpolateWithClock(cnf, globals, deadline, false, verbose, clock)
}
func Inta_ManDeriveClauses(p *Inta_Man_t, cnf *Sto_Man_t, partA int32) *aig.Graph {
	return p.deriveClauses(cnf, partA)
}
