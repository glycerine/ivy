package bsat

import (
	"fmt"
	"os"
	"time"
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

const LIT_UNDEF lit = -1

type Int_Man_t = Int_Man_t_
type Int_Man_t_ struct {
	pCnf                               *Sto_Man_t
	pGloVars                           [16]int32
	nGloVars, nVarsAB, nWords          int32
	pVarTypes                          []int32
	pInters                            []uint32
	fVerbose, fProofVerif, fProofWrite int32
	Counter                            int32
	timeTotal                          int64
	lastErr                            error
}

func Int_ManAlloc() *Int_Man_t { return &Int_Man_t{fProofVerif: 1} }
func Int_ManFree(p *Int_Man_t) {
	if p != nil {
		*p = Int_Man_t{}
	}
}
func Int_ManSetGlobalVars(p *Int_Man_t, n int32) []int32 {
	if n < 0 || n > 16 {
		panic("invalid requested shared variable count")
	}
	p.nGloVars = n
	return p.pGloVars[:n]
}
func Int_ManGlobalVars(p *Int_Man_t) int32 {
	types, shared, err := interpolationVariables(p.pCnf, p.pGloVars[:p.nGloVars])
	if err != nil {
		p.lastErr = err
		return 0
	}
	p.pVarTypes = types
	p.nVarsAB = int32(len(shared))
	return p.nVarsAB
}
func Int_ManTruthRead(p *Int_Man_t, c *Sto_Cls_t) []uint32 {
	start := int(c.Id) * int(p.nWords)
	return p.pInters[start : start+int(p.nWords)]
}
func Int_ManTruthClear(p []uint32, n int32) { clear(p[:n]) }
func Int_ManTruthFill(p []uint32, n int32) {
	for i := range p[:n] {
		p[i] = ^uint32(0)
	}
}
func Int_ManTruthCopy(p, q []uint32, n int32) { copy(p[:n], q[:n]) }
func Int_ManTruthAnd(p, q []uint32, n int32) {
	for i := range p[:n] {
		p[i] &= q[i]
	}
}
func Int_ManTruthOr(p, q []uint32, n int32) {
	for i := range p[:n] {
		p[i] |= q[i]
	}
}
func Int_ManTruthOrNot(p, q []uint32, n int32) {
	for i := range p[:n] {
		p[i] |= ^q[i]
	}
}
func (p *Int_Man_t_) root(c *Sto_Cls_t) {
	out := Int_ManTruthRead(p, c)
	if c.fA == 0 {
		Int_ManTruthFill(out, p.nWords)
		return
	}
	for _, l := range c.pLits {
		kind := p.pVarTypes[lit_var(l)]
		if kind >= 0 {
			continue
		}
		variable := -kind - 1
		for bit := 0; bit < len(out)*32; bit++ {
			value := (bit >> uint(variable)) & 1
			if value == 1-int(lit_sign(l)) {
				out[bit/32] |= uint32(1) << uint(bit%32)
			}
		}
	}
}
func (p *Int_Man_t_) begin(final, conflict *Sto_Cls_t) {
	copy(Int_ManTruthRead(p, final), Int_ManTruthRead(p, conflict))
}
func (p *Int_Man_t_) resolve(final, reason *Sto_Cls_t, pivot lit) {
	out, in := Int_ManTruthRead(p, final), Int_ManTruthRead(p, reason)
	if p.pVarTypes[lit_var(pivot)] == 1 {
		Int_ManTruthOr(out, in, p.nWords)
	} else {
		Int_ManTruthAnd(out, in, p.nWords)
	}
}
func Int_ManInterpolate(p *Int_Man_t, cnf *Sto_Man_t, verbose int32, result *[]uint32) int32 {
	started := time.Now()
	*result = nil
	p.lastErr = nil
	p.pCnf = cnf
	p.fVerbose = verbose
	Int_ManGlobalVars(p)
	if p.lastErr != nil {
		return 0
	}
	if p.nVarsAB > 8 {
		p.lastErr = fmt.Errorf("truth interpolation supports at most eight shared variables")
		return 0
	}
	p.nWords = 1
	if p.nVarsAB > 5 {
		p.nWords = 1 << uint(p.nVarsAB-5)
	}
	p.pInters = make([]uint32, int(cnf.nClauses)*int(p.nWords))
	var file *os.File
	var err error
	if p.fProofWrite != 0 {
		file, err = os.Create("proof.cnf_")
		if err != nil {
			p.lastErr = err
			return 0
		}
	}
	// An interface holding a nil *os.File is not a nil writer.
	var trace *resolutionTrace
	if file == nil {
		trace, err = reconstructResolution(cnf, p, nil, 0)
	} else {
		trace, err = reconstructResolution(cnf, p, file, 0)
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}
	p.timeTotal += time.Since(started).Nanoseconds()
	p.lastErr = err
	if trace != nil {
		p.Counter = trace.counter
	}
	if err != nil {
		return 0
	}
	*result = Int_ManTruthRead(p, cnf.pEmpty)
	if verbose != 0 {
		fmt.Printf("Vars = %d. Roots = %d. Learned = %d. Resol steps = %d.\n", cnf.nVars, cnf.nRoots, cnf.nClauses-cnf.nRoots, p.Counter)
	}
	return p.nVarsAB
}
func Int_ManError(p *Int_Man_t) error { return p.lastErr }

// TruthInterpolant returns an owned truth table and its shared-variable order.
func TruthInterpolant(cnf *Sto_Man_t, globals []int32) ([]uint32, []int32, error) {
	p := Int_ManAlloc()
	if len(globals) > 16 {
		return nil, nil, fmt.Errorf("too many shared truth variables")
	}
	copy(Int_ManSetGlobalVars(p, int32(len(globals))), globals)
	var result []uint32
	Int_ManInterpolate(p, cnf, 0, &result)
	if p.lastErr != nil {
		return nil, nil, p.lastErr
	}
	_, order, err := interpolationVariables(cnf, globals)
	return append([]uint32(nil), result...), order, err
}
