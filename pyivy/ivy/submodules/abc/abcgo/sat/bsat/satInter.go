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

// This is a line-by-line port of satInter.c: the proof is re-derived with
// the manager's own two-watched-literal BCP over the stored clauses (watch
// lists are threaded through Sto_Cls_t.pNext0/pNext1 exactly as in C), and
// the truth-table interpolant is accumulated along the resolution trace.
// The order of propagation and resolution determines the interpolant, so
// it must follow the C code exactly.

// variable assignments
const LIT_UNDEF lit = -1

// interpolation manager
type Int_Man_t = Int_Man_t_
type Int_Man_t_ struct {
	// clauses of the problems
	pCnf     *Sto_Man_t // the set of CNF clauses for A and B
	pGloVars [16]int32  // global variables
	nGloVars int32      // the number of global variables
	// various parameters
	fVerbose    int32 // verbosiness flag
	fProofVerif int32 // verifies the proof
	fProofWrite int32 // writes the proof file
	// internal BCP
	nRootSize  int32        // the number of root level assignments
	nTrailSize int32        // the number of assignments made
	pTrail     []lit        // chronological order of assignments (size nVars)
	pAssigns   []lit        // assignments by variable (size nVars)
	pSeens     []byte       // temporary mark (size nVars)
	pReasons   []*Sto_Cls_t // reasons for each assignment (size nVars)
	pWatches   []*Sto_Cls_t // watched clauses for each literal (size 2*nVars)
	// interpolation data
	nVarsAB   int32    // the number of global variables
	pVarTypes []int32  // variable type (size nVars) [1=A, 0=B, <0=AB]
	pInters   []uint32 // storage for interpolants as truth tables (size nClauses)
	nWords    int32    // the number of words in the truth table
	// proof recording
	Counter    int32   // counter of resolved clauses
	pProofNums []int32 // the proof numbers for each clause (size nClauses)
	pFile      *os.File
	// internal verification
	pResLits      []lit // the literals of the resolvent
	nResLits      int32 // the number of literals of the resolvent
	nResLitsAlloc int32 // the number of literals of the resolvent
	// runtime stats
	timeBcp, timeTrace, timeTotal int64
	lastErr                       error
}

// procedure to get hold of the clauses' truth table
func Int_ManTruthRead(p *Int_Man_t, pCls *Sto_Cls_t) []uint32 {
	start := int(pCls.Id) * int(p.nWords)
	return p.pInters[start : start+int(p.nWords)]
}
func Int_ManTruthClear(p []uint32, nWords int32) {
	for i := nWords - 1; i >= 0; i-- {
		p[i] = 0
	}
}
func Int_ManTruthFill(p []uint32, nWords int32) {
	for i := nWords - 1; i >= 0; i-- {
		p[i] = ^uint32(0)
	}
}
func Int_ManTruthCopy(p, q []uint32, nWords int32) {
	for i := nWords - 1; i >= 0; i-- {
		p[i] = q[i]
	}
}
func Int_ManTruthAnd(p, q []uint32, nWords int32) {
	for i := nWords - 1; i >= 0; i-- {
		p[i] &= q[i]
	}
}
func Int_ManTruthOr(p, q []uint32, nWords int32) {
	for i := nWords - 1; i >= 0; i-- {
		p[i] |= q[i]
	}
}
func Int_ManTruthOrNot(p, q []uint32, nWords int32) {
	for i := nWords - 1; i >= 0; i-- {
		p[i] |= ^q[i]
	}
}

// reading/writing the proof for a clause
func Int_ManProofGet(p *Int_Man_t, pCls *Sto_Cls_t) int32    { return p.pProofNums[pCls.Id] }
func Int_ManProofSet(p *Int_Man_t, pCls *Sto_Cls_t, n int32) { p.pProofNums[pCls.Id] = n }

// Int_ManAlloc allocates the proof manager.
func Int_ManAlloc() *Int_Man_t {
	p := &Int_Man_t{}
	// verification
	p.nResLitsAlloc = 1 << 16
	p.pResLits = make([]lit, p.nResLitsAlloc)
	// parameters
	p.fProofWrite = 0
	p.fProofVerif = 1
	return p
}

// Int_ManSetGlobalVars returns the array of global variables to be filled.
func Int_ManSetGlobalVars(p *Int_Man_t, nGloVars int32) []int32 {
	if nGloVars < 0 || nGloVars > 16 {
		panic("invalid requested shared variable count")
	}
	p.nGloVars = nGloVars
	return p.pGloVars[:nGloVars]
}

// Int_ManGlobalVars counts common variables in the clauses of A and B.
func Int_ManGlobalVars(p *Int_Man_t) int32 {
	// mark the variable encountered in the clauses of A
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		if pClause.fA == 0 {
			break
		}
		for v := int32(0); v < pClause.nLits; v++ {
			p.pVarTypes[lit_var(pClause.pLits[v])] = 1
		}
	}

	if p.nGloVars != 0 {
		for v := int32(0); v < p.nGloVars; v++ {
			p.pVarTypes[p.pGloVars[v]] = -v - 1
		}
		return p.nGloVars
	}

	// check variables that appear in clauses of B
	nVarsAB := int32(0)
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		if pClause.fA != 0 {
			continue
		}
		for v := int32(0); v < pClause.nLits; v++ {
			Var := lit_var(pClause.pLits[v])
			if p.pVarTypes[Var] == 1 { // var of A
				// change it into a global variable
				nVarsAB++
				p.pVarTypes[Var] = -1
			}
		}
	}

	// order global variables
	nVarsAB = 0
	for v := int32(0); v < p.pCnf.nVars; v++ {
		if p.pVarTypes[v] == -1 {
			p.pVarTypes[v] -= nVarsAB
			nVarsAB++
		}
	}
	return nVarsAB
}

// Int_ManResize resizes the proof manager for the current CNF.
func Int_ManResize(p *Int_Man_t) {
	nVars := int(p.pCnf.nVars)
	if len(p.pAssigns) < nVars {
		p.pTrail = make([]lit, nVars)
		p.pAssigns = make([]lit, nVars)
		p.pSeens = make([]byte, nVars)
		p.pVarTypes = make([]int32, nVars)
		p.pReasons = make([]*Sto_Cls_t, nVars)
		p.pWatches = make([]*Sto_Cls_t, 2*nVars)
	}
	// clean the free space
	for i := 0; i < nVars; i++ {
		p.pAssigns[i] = LIT_UNDEF
		p.pSeens[i] = 0
		p.pVarTypes[i] = 0
		p.pReasons[i] = nil
	}
	clear(p.pWatches[:2*nVars])
	// compute the number of common variables
	p.nVarsAB = Int_ManGlobalVars(p)
	// compute the number of words in the truth table
	p.nWords = 1
	if p.nVarsAB > 5 {
		p.nWords = 1 << (p.nVarsAB - 5)
	}
	// check if resizing of clauses is needed
	if len(p.pProofNums) < int(p.pCnf.nClauses) {
		p.pProofNums = make([]int32, p.pCnf.nClauses)
	}
	clear(p.pProofNums[:p.pCnf.nClauses])
	// check if resizing of truth tables is needed (the old content is kept,
	// as by ABC_REALLOC; it is not cleaned)
	if need := int(p.nWords) * int(p.pCnf.nClauses); len(p.pInters) < need {
		p.pInters = append(p.pInters, make([]uint32, need-len(p.pInters))...)
	}
}

// Int_ManFree deallocates the proof manager.
func Int_ManFree(p *Int_Man_t) {
	if p != nil {
		*p = Int_Man_t{}
	}
}

// Int_ManPrintClause prints the clause.
func Int_ManPrintClause(p *Int_Man_t, pClause *Sto_Cls_t) {
	fmt.Printf("Clause ID = %d. Proof = %d. {", pClause.Id, Int_ManProofGet(p, pClause))
	for i := int32(0); i < pClause.nLits; i++ {
		fmt.Printf(" %d", pClause.pLits[i])
	}
	fmt.Printf(" }\n")
}

// Int_ManPrintResolvent prints the resolvent.
func Int_ManPrintResolvent(pResLits []lit, nResLits int32) {
	fmt.Printf("Resolvent: {")
	for i := int32(0); i < nResLits; i++ {
		fmt.Printf(" %d", pResLits[i])
	}
	fmt.Printf(" }\n")
}

// Int_ManWatchClause adds one clause to the watcher list.
func Int_ManWatchClause(p *Int_Man_t, pClause *Sto_Cls_t, Lit lit) {
	if pClause.pLits[0] == Lit {
		pClause.pNext0 = p.pWatches[lit_neg(Lit)]
	} else {
		assert(pClause.pLits[1] == Lit, "pClause->pLits[1] == Lit")
		pClause.pNext1 = p.pWatches[lit_neg(Lit)]
	}
	p.pWatches[lit_neg(Lit)] = pClause
}

// Int_ManEnqueue records implication.
func Int_ManEnqueue(p *Int_Man_t, Lit lit, pReason *Sto_Cls_t) bool {
	Var := lit_var(Lit)
	if p.pAssigns[Var] != LIT_UNDEF {
		return p.pAssigns[Var] == Lit
	}
	p.pAssigns[Var] = Lit
	p.pReasons[Var] = pReason
	p.pTrail[p.nTrailSize] = Lit
	p.nTrailSize++
	return true
}

// Int_ManCancelUntil undoes the assignments down to the given level.
func Int_ManCancelUntil(p *Int_Man_t, Level int32) {
	for i := p.nTrailSize - 1; i >= Level; i-- {
		Lit := p.pTrail[i]
		Var := lit_var(Lit)
		p.pReasons[Var] = nil
		p.pAssigns[Var] = LIT_UNDEF
	}
	p.nTrailSize = Level
}

// Int_ManPropagateOne propagates one assignment.
func Int_ManPropagateOne(p *Int_Man_t, Lit lit) *Sto_Cls_t {
	LitF := lit_neg(Lit)
	// iterate through the literals
	ppPrev := &p.pWatches[Lit]
	for pCur := p.pWatches[Lit]; pCur != nil; pCur = *ppPrev {
		// make sure the false literal is in the second literal of the clause
		if pCur.pLits[0] == LitF {
			pCur.pLits[0] = pCur.pLits[1]
			pCur.pLits[1] = LitF
			pCur.pNext0, pCur.pNext1 = pCur.pNext1, pCur.pNext0
		}
		assert(pCur.pLits[1] == LitF, "pCur->pLits[1] == LitF")

		// if the first literal is true, the clause is satisfied
		if pCur.pLits[0] == p.pAssigns[lit_var(pCur.pLits[0])] {
			ppPrev = &pCur.pNext1
			continue
		}

		// look for a new literal to watch
		i := int32(2)
		for ; i < pCur.nLits; i++ {
			// skip the case when the literal is false
			if lit_neg(pCur.pLits[i]) == p.pAssigns[lit_var(pCur.pLits[i])] {
				continue
			}
			// the literal is either true or unassigned - watch it
			pCur.pLits[1] = pCur.pLits[i]
			pCur.pLits[i] = LitF
			// remove this clause from the watch list of Lit
			*ppPrev = pCur.pNext1
			// add this clause to the watch list of pCur->pLits[i] (now it is pCur->pLits[1])
			Int_ManWatchClause(p, pCur, pCur.pLits[1])
			break
		}
		if i < pCur.nLits { // found new watch
			continue
		}

		// clause is unit - enqueue new implication
		if Int_ManEnqueue(p, pCur.pLits[0], pCur) {
			ppPrev = &pCur.pNext1
			continue
		}

		// conflict detected - return the conflict clause
		return pCur
	}
	return nil
}

// Int_ManPropagate propagates the current assignments.
func Int_ManPropagate(p *Int_Man_t, Start int32) *Sto_Cls_t {
	clk := time.Now()
	for i := Start; i < p.nTrailSize; i++ {
		pClause := Int_ManPropagateOne(p, p.pTrail[i])
		if pClause != nil {
			p.timeBcp += time.Since(clk).Nanoseconds()
			return pClause
		}
	}
	p.timeBcp += time.Since(clk).Nanoseconds()
	return nil
}

// Int_ManProofWriteOne writes one root clause into a file.
func Int_ManProofWriteOne(p *Int_Man_t, pClause *Sto_Cls_t) {
	p.Counter++
	Int_ManProofSet(p, pClause, p.Counter)
	if p.fProofWrite != 0 {
		fmt.Fprintf(p.pFile, "%d", Int_ManProofGet(p, pClause))
		for v := int32(0); v < pClause.nLits; v++ {
			fmt.Fprintf(p.pFile, " %d", lit_print(pClause.pLits[v]))
		}
		fmt.Fprintf(p.pFile, " 0 0\n")
	}
}

// Int_ManProofTraceOne traces the proof for one clause.
func Int_ManProofTraceOne(p *Int_Man_t, pConflict, pFinal *Sto_Cls_t) int32 {
	fPrint := false
	clk := time.Now()

	// collect resolvent literals
	if p.fProofVerif != 0 {
		assert(pConflict.nLits <= p.nResLitsAlloc, "(int)pConflict->nLits <= p->nResLitsAlloc")
		copy(p.pResLits, pConflict.pLits[:pConflict.nLits])
		p.nResLits = pConflict.nLits
	}

	// mark all the variables in the conflict as seen
	for v := int32(0); v < pConflict.nLits; v++ {
		p.pSeens[lit_var(pConflict.pLits[v])] = 1
	}

	// start the anticedents
	if p.pCnf.nClausesA != 0 {
		Int_ManTruthCopy(Int_ManTruthRead(p, pFinal), Int_ManTruthRead(p, pConflict), p.nWords)
	}

	// follow the trail backwards
	PrevId := Int_ManProofGet(p, pConflict)
	for i := p.nTrailSize - 1; i >= 0; i-- {
		// skip literals that are not involved
		Var := lit_var(p.pTrail[i])
		if p.pSeens[Var] == 0 {
			continue
		}
		p.pSeens[Var] = 0

		// skip literals of the resulting clause
		pReason := p.pReasons[Var]
		if pReason == nil {
			continue
		}
		assert(p.pTrail[i] == pReason.pLits[0], "p->pTrail[i] == pReason->pLits[0]")

		// add the variables to seen
		for v := int32(1); v < pReason.nLits; v++ {
			p.pSeens[lit_var(pReason.pLits[v])] = 1
		}

		// record the reason clause
		assert(Int_ManProofGet(p, pReason) > 0, "Int_ManProofGet(p, pReason) > 0")
		p.Counter++
		if p.fProofWrite != 0 {
			fmt.Fprintf(p.pFile, "%d * %d %d 0\n", p.Counter, PrevId, Int_ManProofGet(p, pReason))
		}
		PrevId = p.Counter

		if p.pCnf.nClausesA != 0 {
			if p.pVarTypes[Var] == 1 { // var of A
				Int_ManTruthOr(Int_ManTruthRead(p, pFinal), Int_ManTruthRead(p, pReason), p.nWords)
			} else {
				Int_ManTruthAnd(Int_ManTruthRead(p, pFinal), Int_ManTruthRead(p, pReason), p.nWords)
			}
		}

		// resolve the temporary resolvent with the reason clause
		if p.fProofVerif != 0 {
			if fPrint {
				Int_ManPrintResolvent(p.pResLits, p.nResLits)
			}
			// check that the var is present in the resolvent
			v1 := int32(0)
			for ; v1 < p.nResLits; v1++ {
				if lit_var(p.pResLits[v1]) == Var {
					break
				}
			}
			if v1 == p.nResLits {
				fmt.Printf("Recording clause %d: Cannot find variable %d in the temporary resolvent.\n", pFinal.Id, Var)
			}
			if p.pResLits[v1] != lit_neg(pReason.pLits[0]) {
				fmt.Printf("Recording clause %d: The resolved variable %d is in the wrong polarity.\n", pFinal.Id, Var)
			}
			// remove this variable from the resolvent
			assert(lit_var(p.pResLits[v1]) == Var, "lit_var(p->pResLits[v1]) == Var")
			p.nResLits--
			for ; v1 < p.nResLits; v1++ {
				p.pResLits[v1] = p.pResLits[v1+1]
			}
			// add variables of the reason clause
			for v2 := int32(1); v2 < pReason.nLits; v2++ {
				for v1 = 0; v1 < p.nResLits; v1++ {
					if lit_var(p.pResLits[v1]) == lit_var(pReason.pLits[v2]) {
						break
					}
				}
				// if it is a new variable, add it to the resolvent
				if v1 == p.nResLits {
					if p.nResLits == p.nResLitsAlloc {
						fmt.Printf("Recording clause %d: Ran out of space for intermediate resolvent.\n", pFinal.Id)
					}
					p.pResLits[p.nResLits] = pReason.pLits[v2]
					p.nResLits++
					continue
				}
				// if the variable is the same, the literal should be the same too
				if p.pResLits[v1] == pReason.pLits[v2] {
					continue
				}
				// the literal is different
				fmt.Printf("Recording clause %d: Trying to resolve the clause with more than one opposite literal.\n", pFinal.Id)
			}
		}
	}

	// use the resulting clause to check the correctness of resolution
	if p.fProofVerif != 0 {
		if fPrint {
			Int_ManPrintResolvent(p.pResLits, p.nResLits)
		}
		v1 := int32(0)
		for ; v1 < p.nResLits; v1++ {
			v2 := int32(0)
			for ; v2 < pFinal.nLits; v2++ {
				if pFinal.pLits[v2] == p.pResLits[v1] {
					break
				}
			}
			if v2 < pFinal.nLits {
				continue
			}
			break
		}
		if v1 < p.nResLits {
			fmt.Printf("Recording clause %d: The final resolvent is wrong.\n", pFinal.Id)
			Int_ManPrintClause(p, pConflict)
			Int_ManPrintResolvent(p.pResLits, p.nResLits)
			Int_ManPrintClause(p, pFinal)
		}

		// if there are literals in the clause that are not in the resolvent
		// it means that the derived resolvent is stronger than the clause
		// we can replace the clause with the resolvent by removing these literals
		if p.nResLits != pFinal.nLits {
			for v1 = 0; v1 < pFinal.nLits; v1++ {
				v2 := int32(0)
				for ; v2 < p.nResLits; v2++ {
					if pFinal.pLits[v1] == p.pResLits[v2] {
						break
					}
				}
				if v2 < p.nResLits {
					continue
				}
				// remove literal v1 from the final clause
				pFinal.nLits--
				for v2 = v1; v2 < pFinal.nLits; v2++ {
					pFinal.pLits[v2] = pFinal.pLits[v2+1]
				}
				v1--
			}
			pFinal.pLits = pFinal.pLits[:pFinal.nLits]
			assert(p.nResLits == pFinal.nLits, "p->nResLits == (int)pFinal->nLits")
		}
	}
	p.timeTrace += time.Since(clk).Nanoseconds()

	Int_ManProofSet(p, pFinal, p.Counter)
	return p.Counter
}

// Int_ManProofRecordOne records the proof for one clause.
func Int_ManProofRecordOne(p *Int_Man_t, pClause *Sto_Cls_t) int32 {
	// empty clause never ends up there
	if pClause.nLits == 0 {
		fmt.Printf("Error: Empty clause is attempted.\n")
	}

	// add assumptions to the trail
	assert(pClause.fRoot == 0, "!pClause->fRoot")
	assert(p.nTrailSize == p.nRootSize, "p->nTrailSize == p->nRootSize")

	// if any of the clause literals are already assumed
	// it means that the clause is redundant and can be skipped
	for i := int32(0); i < pClause.nLits; i++ {
		if p.pAssigns[lit_var(pClause.pLits[i])] == pClause.pLits[i] {
			return 1
		}
	}

	for i := int32(0); i < pClause.nLits; i++ {
		if !Int_ManEnqueue(p, lit_neg(pClause.pLits[i]), nil) {
			return 0 // impossible
		}
	}

	// propagate the assumptions
	pConflict := Int_ManPropagate(p, p.nRootSize)
	if pConflict == nil {
		return 0 // cannot prove
	}

	// skip the clause if it is weaker or the same as the conflict clause
	if pClause.nLits >= pConflict.nLits {
		// check if every literal of conflict clause can be found in the given clause
		i := int32(0)
		for ; i < pConflict.nLits; i++ {
			j := int32(0)
			for ; j < pClause.nLits; j++ {
				if pConflict.pLits[i] == pClause.pLits[j] {
					break
				}
			}
			if j == pClause.nLits { // literal pConflict->pLits[i] is not found
				break
			}
		}
		if i == pConflict.nLits { // all lits are found
			// undo to the root level
			Int_ManCancelUntil(p, p.nRootSize)
			return 1
		}
	}

	// construct the proof
	Int_ManProofTraceOne(p, pConflict, pClause)

	// undo to the root level
	Int_ManCancelUntil(p, p.nRootSize)

	// add large clauses to the watched lists
	if pClause.nLits > 1 {
		Int_ManWatchClause(p, pClause, pClause.pLits[0])
		Int_ManWatchClause(p, pClause, pClause.pLits[1])
		return 1
	}
	assert(pClause.nLits == 1, "pClause->nLits == 1")

	// if the clause proved is unit, add it and propagate
	if !Int_ManEnqueue(p, pClause.pLits[0], pClause) {
		return 0 // impossible
	}

	// propagate the assumption
	pConflict = Int_ManPropagate(p, p.nRootSize)
	if pConflict != nil {
		// construct the proof
		Int_ManProofTraceOne(p, pConflict, p.pCnf.pEmpty)
		if p.fVerbose != 0 {
			fmt.Printf("Found last conflict after adding unit clause number %d!\n", pClause.Id)
		}
		return 0
	}

	// update the root level
	p.nRootSize = p.nTrailSize
	return 1
}

// Int_ManProcessRoots propagates the root clauses.
func Int_ManProcessRoots(p *Int_Man_t) int32 {
	// make sure the root clauses are preceeding the learnt clauses
	Counter := int32(0)
	for pClause := p.pCnf.pHead; pClause != nil; pClause = pClause.pNext {
		assert((pClause.fA != 0) == (Counter < p.pCnf.nClausesA), "(int)pClause->fA == (Counter < (int)p->pCnf->nClausesA)")
		assert((pClause.fRoot != 0) == (Counter < p.pCnf.nRoots), "(int)pClause->fRoot == (Counter < (int)p->pCnf->nRoots)")
		Counter++
	}
	assert(p.pCnf.nClauses == Counter, "p->pCnf->nClauses == Counter")

	// make sure the last clause if empty
	assert(p.pCnf.pTail.nLits == 0, "p->pCnf->pTail->nLits == 0")

	// go through the root unit clauses
	p.nTrailSize = 0
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		// create watcher lists for the root clauses
		if pClause.nLits > 1 {
			Int_ManWatchClause(p, pClause, pClause.pLits[0])
			Int_ManWatchClause(p, pClause, pClause.pLits[1])
		}
		// empty clause and large clauses
		if pClause.nLits != 1 {
			continue
		}
		// unit clause
		if !Int_ManEnqueue(p, pClause.pLits[0], pClause) {
			// detected root level conflict
			Int_ManProofTraceOne(p, pClause, p.pCnf.pEmpty)
			if p.fVerbose != 0 {
				fmt.Printf("Found root level conflict!\n")
			}
			return 0
		}
	}

	// propagate the root unit clauses
	if pClause := Int_ManPropagate(p, 0); pClause != nil {
		// detected root level conflict
		Int_ManProofTraceOne(p, pClause, p.pCnf.pEmpty)
		if p.fVerbose != 0 {
			fmt.Printf("Found root level conflict!\n")
		}
		return 0
	}

	// set the root level
	p.nRootSize = p.nTrailSize
	return 1
}

// elementary truth tables
var intUTruths = [8][8]uint32{
	{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
	{0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC},
	{0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0},
	{0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00},
	{0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000},
	{0x00000000, 0xFFFFFFFF, 0x00000000, 0xFFFFFFFF, 0x00000000, 0xFFFFFFFF, 0x00000000, 0xFFFFFFFF},
	{0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF, 0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF},
	{0x00000000, 0x00000000, 0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
}

// Int_ManPrepareInter sets the interpolants of the root clauses.
func Int_ManPrepareInter(p *Int_Man_t) {
	assert(p.nVarsAB <= 8, "p->nVarsAB <= 8")
	// set interpolants for root clauses
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		if pClause.fA == 0 { // clause of B
			Int_ManTruthFill(Int_ManTruthRead(p, pClause), p.nWords)
			continue
		}
		// clause of A
		Int_ManTruthClear(Int_ManTruthRead(p, pClause), p.nWords)
		for v := int32(0); v < pClause.nLits; v++ {
			Var := lit_var(pClause.pLits[v])
			if p.pVarTypes[Var] < 0 { // global var
				VarAB := -p.pVarTypes[Var] - 1
				assert(VarAB >= 0 && VarAB < p.nVarsAB, "VarAB >= 0 && VarAB < p->nVarsAB")
				if lit_sign(pClause.pLits[v]) != 0 { // negative var
					Int_ManTruthOrNot(Int_ManTruthRead(p, pClause), intUTruths[VarAB][:], p.nWords)
				} else {
					Int_ManTruthOr(Int_ManTruthRead(p, pClause), intUTruths[VarAB][:], p.nWords)
				}
			}
		}
	}
}

// Int_ManInterpolate computes interpolant for the given CNF. Returns the
// number of common variable found and interpolant. Returns 0, if something
// did not work.
func Int_ManInterpolate(p *Int_Man_t, pCnf *Sto_Man_t, fVerbose int32, ppResult *[]uint32) int32 {
	clkTotal := time.Now()
	p.lastErr = nil
	// check that the CNF makes sense
	if pCnf == nil || pCnf.nVars <= 0 || pCnf.nClauses <= 0 || pCnf.pTail == nil {
		p.lastErr = fmt.Errorf("invalid interpolation CNF")
		*ppResult = nil
		return 0
	}
	p.pCnf = pCnf
	p.fVerbose = fVerbose
	*ppResult = nil

	// adjust the manager
	Int_ManResize(p)
	if p.nVarsAB > 8 {
		p.lastErr = fmt.Errorf("truth interpolation supports at most eight shared variables")
		return 0
	}

	// prepare the interpolant computation
	Int_ManPrepareInter(p)

	// construct proof for each clause
	// start the proof
	if p.fProofWrite != 0 {
		f, err := os.Create("proof.cnf_")
		if err != nil {
			p.lastErr = err
			return 0
		}
		p.pFile = f
		p.Counter = 0
	}

	// write the root clauses
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		Int_ManProofWriteOne(p, pClause)
	}

	// propagate root level assignments
	if Int_ManProcessRoots(p) != 0 {
		// if there is no conflict, consider learned clauses
		for pClause := p.pCnf.pHead; pClause != nil; pClause = pClause.pNext {
			if pClause.fRoot != 0 {
				continue
			}
			if Int_ManProofRecordOne(p, pClause) == 0 {
				break
			}
		}
	}

	// stop the proof
	if p.fProofWrite != 0 {
		p.pFile.Close()
		p.pFile = nil
	}

	if fVerbose != 0 {
		fmt.Printf("Vars = %d. Roots = %d. Learned = %d. Resol steps = %d.  Ave = %.2f.\n",
			p.pCnf.nVars, p.pCnf.nRoots, p.pCnf.nClauses-p.pCnf.nRoots, p.Counter,
			1.0*float64(p.Counter-p.pCnf.nRoots)/float64(p.pCnf.nClauses-p.pCnf.nRoots))
		p.timeTotal += time.Since(clkTotal).Nanoseconds()
	}
	*ppResult = Int_ManTruthRead(p, p.pCnf.pTail)
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
	order := make([]int32, p.nVarsAB)
	for v := int32(0); v < cnf.nVars; v++ {
		if t := p.pVarTypes[v]; t < 0 {
			order[-t-1] = v
		}
	}
	return append([]uint32(nil), result...), order, nil
}
