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

type Intb_Man_t = Intb_Man_t_
type Intb_Man_t_ struct{ aigInterpolation }

func Intb_ManAlloc() *Intb_Man_t { return &Intb_Man_t{} }
func Intb_ManFree(p *Intb_Man_t) {
	if p != nil {
		*p = Intb_Man_t{}
	}
}
func Intb_ManError(p *Intb_Man_t) error { return p.lastErr }
func Intb_ManInterpolate(p *Intb_Man_t, cnf *Sto_Man_t, globals []int32, verbose int32) *aig.Graph {
	return p.interpolate(cnf, globals, 0, true, verbose)
}
func Intb_ManDeriveClauses(p *Intb_Man_t, cnf *Sto_Man_t, partA int32) *aig.Graph {
	return p.deriveClauses(cnf, partA)
}
