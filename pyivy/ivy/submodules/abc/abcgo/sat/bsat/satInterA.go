package bsat

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
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

// Line-by-line port of src/sat/bsat/satInterA.c (McMillan's interpolant as
// an AIG, computed by replaying the stored resolution proof with BCP).

// variable assignments
const intaLIT_UNDEF lit = -1 // 0xffffffff

// interpolation manager
type Inta_Man_t = Inta_Man_t_
type Inta_Man_t_ struct {
	// clauses of the problems
	pCnf    *Sto_Man_t // the set of CNF clauses for A and B
	vVarsAB []int32    // the array of global variables
	// various parameters
	fVerbose    int32 // verbosiness flag
	fProofVerif int32 // verifies the proof
	fProofWrite int32 // writes the proof file
	nVarsAlloc  int32 // the allocated size of var arrays
	nClosAlloc  int32 // the allocated size of clause arrays
	// internal BCP
	nRootSize  int32        // the number of root level assignments
	nTrailSize int32        // the number of assignments made
	pTrail     []lit        // chronological order of assignments (size nVars)
	pAssigns   []lit        // assignments by variable (size nVars)
	pSeens     []byte       // temporary mark (size nVars)
	pReasons   []*Sto_Cls_t // reasons for each assignment (size nVars)
	pWatches   []*Sto_Cls_t // watched clauses for each literal (size 2*nVars)
	// interpolation data
	pAig         *aig.Graph // the AIG manager for recording the interpolant
	pVarTypes    []int32    // variable type (size nVars) [1=A, 0=B, <0=AB]
	pInters      []aig.Edge // storage for interpolants as truth tables (size nClauses)
	nIntersAlloc int32      // the allocated size of truth table array
	// proof recording
	Counter    int32    // counter of resolved clauses
	pProofNums []int32  // the proof numbers for each clause (size nClauses)
	pFile      *os.File // the file for proof recording
	// internal verification
	vResLits []lit
	// runtime stats
	timeBcp   int64 // the runtime for BCP
	timeTrace int64 // the runtime of trace construction
	timeTotal int64 // the total runtime of interpolation
	// clock is the Abc_Clock domain of TimeToStop (nil: util.Abc_Clock).
	clock func() int64
}

// procedure to get hold of the clauses' truth table
func Inta_ManAigRead(pMan *Inta_Man_t, pCls *Sto_Cls_t) *aig.Edge { return &pMan.pInters[pCls.Id] }
func Inta_ManAigClear(pMan *Inta_Man_t, p *aig.Edge)              { *p = aig.Aig_ManConst0(pMan.pAig) }
func Inta_ManAigFill(pMan *Inta_Man_t, p *aig.Edge)               { *p = aig.Aig_ManConst1(pMan.pAig) }
func Inta_ManAigCopy(pMan *Inta_Man_t, p, q *aig.Edge)            { *p = *q }
func Inta_ManAigAnd(pMan *Inta_Man_t, p, q *aig.Edge)             { *p = aig.Aig_And(pMan.pAig, *p, *q) }
func Inta_ManAigOr(pMan *Inta_Man_t, p, q *aig.Edge)              { *p = aig.Aig_Or(pMan.pAig, *p, *q) }
func Inta_ManAigOrNot(pMan *Inta_Man_t, p, q *aig.Edge) {
	*p = aig.Aig_Or(pMan.pAig, *p, aig.Aig_Not(*q))
}
func Inta_ManAigOrVar(pMan *Inta_Man_t, p *aig.Edge, v int32) {
	*p = aig.Aig_Or(pMan.pAig, *p, aig.Aig_IthVar(pMan.pAig, v))
}
func Inta_ManAigOrNotVar(pMan *Inta_Man_t, p *aig.Edge, v int32) {
	*p = aig.Aig_Or(pMan.pAig, *p, aig.Aig_Not(aig.Aig_IthVar(pMan.pAig, v)))
}

// reading/writing the proof for a clause
func Inta_ManProofGet(p *Inta_Man_t, pCls *Sto_Cls_t) int32    { return p.pProofNums[pCls.Id] }
func Inta_ManProofSet(p *Inta_Man_t, pCls *Sto_Cls_t, n int32) { p.pProofNums[pCls.Id] = n }

func (p *Inta_Man_t) now() int64 {
	if p.clock != nil {
		return p.clock()
	}
	return util.Abc_Clock()
}

/*
*Function*************************************************************

	Synopsis    [Allocate proof manager.]

**********************************************************************
*/
func Inta_ManAlloc() *Inta_Man_t {
	var p *Inta_Man_t
	// allocate the manager
	p = &Inta_Man_t{}
	// verification
	p.vResLits = make([]lit, 0, 1000)
	// parameters
	p.fProofWrite = 0
	p.fProofVerif = 1
	return p
}

/*
*Function*************************************************************

	Synopsis    [Count common variables in the clauses of A and B.]

**********************************************************************
*/
func Inta_ManGlobalVars(p *Inta_Man_t) int32 {
	var LargeNum int32 = -100000000
	var Var, nVarsAB, v int32
	// mark the variable encountered in the clauses of A
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		if pClause.fA == 0 {
			break
		}
		for v = 0; v < pClause.nLits; v++ {
			p.pVarTypes[lit_var(pClause.pLits[v])] = 1
		}
	}
	// check variables that appear in clauses of B
	nVarsAB = 0
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		if pClause.fA != 0 {
			continue
		}
		for v = 0; v < pClause.nLits; v++ {
			Var = lit_var(pClause.pLits[v])
			if p.pVarTypes[Var] == 1 { // var of A
				// change it into a global variable
				nVarsAB++
				p.pVarTypes[Var] = LargeNum
			}
		}
	}
	assert(nVarsAB <= int32(len(p.vVarsAB)), "nVarsAB <= Vec_IntSize(p->vVarsAB)")
	// order global variables
	nVarsAB = 0
	for _, Var = range p.vVarsAB {
		nVarsAB++
		p.pVarTypes[Var] = -nVarsAB
	}
	// check that there is no extra global variables
	for v = 0; v < p.pCnf.nVars; v++ {
		assert(p.pVarTypes[v] != LargeNum, "p->pVarTypes[v] != LargeNum")
	}
	return nVarsAB
}

/*
*Function*************************************************************

	Synopsis    [Resize proof manager.]

**********************************************************************
*/
func Inta_ManResize(p *Inta_Man_t) {
	p.Counter = 0
	// check if resizing is needed
	if p.nVarsAlloc < p.pCnf.nVars {
		// find the new size
		if p.nVarsAlloc == 0 {
			p.nVarsAlloc = 1
		}
		for p.nVarsAlloc < p.pCnf.nVars {
			p.nVarsAlloc *= 2
		}
		// resize the arrays
		p.pTrail = make([]lit, p.nVarsAlloc)
		p.pAssigns = make([]lit, p.nVarsAlloc)
		p.pSeens = make([]byte, p.nVarsAlloc)
		p.pVarTypes = make([]int32, p.nVarsAlloc)
		p.pReasons = make([]*Sto_Cls_t, p.nVarsAlloc)
		p.pWatches = make([]*Sto_Cls_t, p.nVarsAlloc*2)
	}
	// clean the free space
	for i := int32(0); i < p.pCnf.nVars; i++ {
		p.pAssigns[i] = intaLIT_UNDEF
		p.pSeens[i] = 0
		p.pVarTypes[i] = 0
		p.pReasons[i] = nil
	}
	clear(p.pWatches[:p.pCnf.nVars*2])
	// compute the number of common variables
	Inta_ManGlobalVars(p)
	// check if resizing of clauses is needed
	if p.nClosAlloc < p.pCnf.nClauses {
		// find the new size
		if p.nClosAlloc == 0 {
			p.nClosAlloc = 1
		}
		for p.nClosAlloc < p.pCnf.nClauses {
			p.nClosAlloc *= 2
		}
		// resize the arrays
		p.pProofNums = make([]int32, p.nClosAlloc)
	}
	clear(p.pProofNums[:p.pCnf.nClauses])
	// check if resizing of truth tables is needed
	if p.nIntersAlloc < p.pCnf.nClauses {
		p.nIntersAlloc = p.pCnf.nClauses
		p.pInters = make([]aig.Edge, p.nIntersAlloc)
	}
	clear(p.pInters[:p.pCnf.nClauses])
}

/*
*Function*************************************************************

	Synopsis    [Deallocate proof manager.]

**********************************************************************
*/
func Inta_ManFree(p *Inta_Man_t) {
	/*
	   printf( "Runtime stats:\n" );
	   ABC_PRT( "BCP     ", p->timeBcp   );
	   ABC_PRT( "Trace   ", p->timeTrace );
	   ABC_PRT( "TOTAL   ", p->timeTotal );
	*/
	if p != nil {
		*p = Inta_Man_t{}
	}
}

// Inta_ManError is kept for callers of the earlier Go API; the C manager
// reports no errors (a NULL interpolant means timeout).
func Inta_ManError(p *Inta_Man_t) error { return nil }

/*
*Function*************************************************************

	Synopsis    [Prints the clause.]

**********************************************************************
*/
func Inta_ManPrintClause(p *Inta_Man_t, pClause *Sto_Cls_t) {
	var i int32
	fmt.Printf("Clause ID = %d. Proof = %d. {", pClause.Id, Inta_ManProofGet(p, pClause))
	for i = 0; i < pClause.nLits; i++ {
		fmt.Printf(" %d", pClause.pLits[i])
	}
	fmt.Printf(" }\n")
}

/*
*Function*************************************************************

	Synopsis    [Prints the resolvent.]

**********************************************************************
*/
func Inta_ManPrintResolvent(vResLits []lit) {
	fmt.Printf("Resolvent: {")
	for _, Entry := range vResLits {
		fmt.Printf(" %d", Entry)
	}
	fmt.Printf(" }\n")
}

/*
*Function*************************************************************

	Synopsis    [Prints the interpolant for one clause.]

**********************************************************************
*/
func Inta_ManPrintInterOne(p *Inta_Man_t, pClause *Sto_Cls_t) {
	fmt.Printf("Clause %2d :  ", pClause.Id)
	//    Extra_PrintBinary___( stdout, Inta_ManAigRead(p, pClause), (1 << p->nVarsAB) );
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Adds one clause to the watcher list.]

**********************************************************************
*/
func Inta_ManWatchClause(p *Inta_Man_t, pClause *Sto_Cls_t, Lit lit) {
	assert(lit_check(Lit, p.pCnf.nVars) != 0, "lit_check(Lit, p->pCnf->nVars)")
	if pClause.pLits[0] == Lit {
		pClause.pNext0 = p.pWatches[lit_neg(Lit)]
	} else {
		assert(pClause.pLits[1] == Lit, "pClause->pLits[1] == Lit")
		pClause.pNext1 = p.pWatches[lit_neg(Lit)]
	}
	p.pWatches[lit_neg(Lit)] = pClause
}

/*
*Function*************************************************************

	Synopsis    [Records implication.]

**********************************************************************
*/
func Inta_ManEnqueue(p *Inta_Man_t, Lit lit, pReason *Sto_Cls_t) int32 {
	Var := lit_var(Lit)
	if p.pAssigns[Var] != intaLIT_UNDEF {
		if p.pAssigns[Var] == Lit {
			return 1
		}
		return 0
	}
	p.pAssigns[Var] = Lit
	p.pReasons[Var] = pReason
	p.pTrail[p.nTrailSize] = Lit
	p.nTrailSize++
	//printf( "assigning var %d value %d\n", Var, !lit_sign(Lit) );
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Records implication.]

**********************************************************************
*/
func Inta_ManCancelUntil(p *Inta_Man_t, Level int32) {
	var Lit lit
	var i, Var int32
	for i = p.nTrailSize - 1; i >= Level; i-- {
		Lit = p.pTrail[i]
		Var = lit_var(Lit)
		p.pReasons[Var] = nil
		p.pAssigns[Var] = intaLIT_UNDEF
		//printf( "cancelling var %d\n", Var );
	}
	p.nTrailSize = Level
}

/*
*Function*************************************************************

	Synopsis    [Propagate one assignment.]

**********************************************************************
*/
func Inta_ManPropagateOne(p *Inta_Man_t, Lit lit) *Sto_Cls_t {
	var pCur, pTemp *Sto_Cls_t
	LitF := lit_neg(Lit)
	var i int32
	// iterate through the literals
	ppPrev := &p.pWatches[Lit]
	for pCur = p.pWatches[Lit]; pCur != nil; pCur = *ppPrev {
		// make sure the false literal is in the second literal of the clause
		if pCur.pLits[0] == LitF {
			pCur.pLits[0] = pCur.pLits[1]
			pCur.pLits[1] = LitF
			pTemp = pCur.pNext0
			pCur.pNext0 = pCur.pNext1
			pCur.pNext1 = pTemp
		}
		assert(pCur.pLits[1] == LitF, "pCur->pLits[1] == LitF")

		// if the first literal is true, the clause is satisfied
		if pCur.pLits[0] == p.pAssigns[lit_var(pCur.pLits[0])] {
			ppPrev = &pCur.pNext1
			continue
		}

		// look for a new literal to watch
		for i = 2; i < pCur.nLits; i++ {
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
			Inta_ManWatchClause(p, pCur, pCur.pLits[1])
			break
		}
		if i < pCur.nLits { // found new watch
			continue
		}

		// clause is unit - enqueue new implication
		if Inta_ManEnqueue(p, pCur.pLits[0], pCur) != 0 {
			ppPrev = &pCur.pNext1
			continue
		}

		// conflict detected - return the conflict clause
		return pCur
	}
	return nil
}

/*
*Function*************************************************************

	Synopsis    [Propagate the current assignments.]

**********************************************************************
*/
func Inta_ManPropagate(p *Inta_Man_t, Start int32) *Sto_Cls_t {
	var pClause *Sto_Cls_t
	var i int32
	clk := util.Abc_Clock()
	for i = Start; i < p.nTrailSize; i++ {
		pClause = Inta_ManPropagateOne(p, p.pTrail[i])
		if pClause != nil {
			p.timeBcp += util.Abc_Clock() - clk
			return pClause
		}
	}
	p.timeBcp += util.Abc_Clock() - clk
	return nil
}

/*
*Function*************************************************************

	Synopsis    [Writes one root clause into a file.]

**********************************************************************
*/
func Inta_ManProofWriteOne(p *Inta_Man_t, pClause *Sto_Cls_t) {
	p.Counter++
	Inta_ManProofSet(p, pClause, p.Counter)
	if p.fProofWrite != 0 {
		var v int32
		fmt.Fprintf(p.pFile, "%d", Inta_ManProofGet(p, pClause))
		for v = 0; v < pClause.nLits; v++ {
			fmt.Fprintf(p.pFile, " %d", lit_print(pClause.pLits[v]))
		}
		fmt.Fprintf(p.pFile, " 0 0\n")
	}
}

// intaVecIntRemove is Vec_IntRemove: deletes the first occurrence of Entry.
func intaVecIntRemove(v []lit, Entry lit) []lit {
	var i int
	for i = 0; i < len(v); i++ {
		if v[i] == Entry {
			break
		}
	}
	if i == len(v) {
		return v
	}
	copy(v[i:], v[i+1:])
	return v[:len(v)-1]
}

/*
*Function*************************************************************

	Synopsis    [Traces the proof for one clause.]

**********************************************************************
*/
func Inta_ManProofTraceOne(p *Inta_Man_t, pConflict, pFinal *Sto_Cls_t) int32 {
	var pReason *Sto_Cls_t
	var i, v, Var, PrevId int32
	fPrint := 0
	clk := util.Abc_Clock()

	// collect resolvent literals
	if p.fProofVerif != 0 {
		p.vResLits = p.vResLits[:0]
		for i = 0; i < pConflict.nLits; i++ {
			p.vResLits = append(p.vResLits, pConflict.pLits[i])
		}
	}

	// mark all the variables in the conflict as seen
	for v = 0; v < pConflict.nLits; v++ {
		p.pSeens[lit_var(pConflict.pLits[v])] = 1
	}

	// start the anticedents
	//    pFinal->pAntis = Vec_PtrAlloc( 32 );
	//    Vec_PtrPush( pFinal->pAntis, pConflict );

	if p.pCnf.nClausesA != 0 {
		Inta_ManAigCopy(p, Inta_ManAigRead(p, pFinal), Inta_ManAigRead(p, pConflict))
	}

	// follow the trail backwards
	PrevId = Inta_ManProofGet(p, pConflict)
	for i = p.nTrailSize - 1; i >= 0; i-- {
		// skip literals that are not involved
		Var = lit_var(p.pTrail[i])
		if p.pSeens[Var] == 0 {
			continue
		}
		p.pSeens[Var] = 0

		// skip literals of the resulting clause
		pReason = p.pReasons[Var]
		if pReason == nil {
			continue
		}
		assert(p.pTrail[i] == pReason.pLits[0], "p->pTrail[i] == pReason->pLits[0]")

		// add the variables to seen
		for v = 1; v < pReason.nLits; v++ {
			p.pSeens[lit_var(pReason.pLits[v])] = 1
		}

		// record the reason clause
		assert(Inta_ManProofGet(p, pReason) > 0, "Inta_ManProofGet(p, pReason) > 0")
		p.Counter++
		if p.fProofWrite != 0 {
			fmt.Fprintf(p.pFile, "%d * %d %d 0\n", p.Counter, PrevId, Inta_ManProofGet(p, pReason))
		}
		PrevId = p.Counter

		if p.pCnf.nClausesA != 0 {
			if p.pVarTypes[Var] == 1 { // var of A
				Inta_ManAigOr(p, Inta_ManAigRead(p, pFinal), Inta_ManAigRead(p, pReason))
			} else {
				Inta_ManAigAnd(p, Inta_ManAigRead(p, pFinal), Inta_ManAigRead(p, pReason))
			}
		}

		// resolve the temporary resolvent with the reason clause
		if p.fProofVerif != 0 {
			var v1, v2 int32
			Entry := lit(-1)
			if fPrint != 0 {
				Inta_ManPrintResolvent(p.vResLits)
			}
			// check that the var is present in the resolvent
			for v1 = 0; v1 < int32(len(p.vResLits)); v1++ {
				Entry = p.vResLits[v1]
				if lit_var(Entry) == Var {
					break
				}
			}
			if v1 == int32(len(p.vResLits)) {
				fmt.Printf("Recording clause %d: Cannot find variable %d in the temporary resolvent.\n", pFinal.Id, Var)
			}
			if Entry != lit_neg(pReason.pLits[0]) {
				fmt.Printf("Recording clause %d: The resolved variable %d is in the wrong polarity.\n", pFinal.Id, Var)
			}
			// remove variable v1 from the resolvent
			assert(lit_var(Entry) == Var, "lit_var(Entry) == Var")
			p.vResLits = intaVecIntRemove(p.vResLits, Entry)
			// add variables of the reason clause
			for v2 = 1; v2 < pReason.nLits; v2++ {
				for v1 = 0; v1 < int32(len(p.vResLits)); v1++ {
					Entry = p.vResLits[v1]
					if lit_var(Entry) == lit_var(pReason.pLits[v2]) {
						break
					}
				}
				// if it is a new variable, add it to the resolvent
				if v1 == int32(len(p.vResLits)) {
					p.vResLits = append(p.vResLits, pReason.pLits[v2])
					continue
				}
				// if the variable is the same, the literal should be the same too
				if Entry == pReason.pLits[v2] {
					continue
				}
				// the literal is different
				fmt.Printf("Recording clause %d: Trying to resolve the clause with more than one opposite literal.\n", pFinal.Id)
			}
		}

		//        Vec_PtrPush( pFinal->pAntis, pReason );
	}

	// unmark all seen variables
	//    for ( i = p->nTrailSize - 1; i >= 0; i-- )
	//        p->pSeens[lit_var(p->pTrail[i])] = 0;
	// check that the literals are unmarked
	//    for ( i = p->nTrailSize - 1; i >= 0; i-- )
	//        assert( p->pSeens[lit_var(p->pTrail[i])] == 0 );

	// use the resulting clause to check the correctness of resolution
	if p.fProofVerif != 0 {
		var v1, v2 int32
		if fPrint != 0 {
			Inta_ManPrintResolvent(p.vResLits)
		}
		for v1 = 0; v1 < int32(len(p.vResLits)); v1++ {
			Entry := p.vResLits[v1]
			for v2 = 0; v2 < pFinal.nLits; v2++ {
				if pFinal.pLits[v2] == Entry {
					break
				}
			}
			if v2 < pFinal.nLits {
				continue
			}
			break
		}
		if v1 < int32(len(p.vResLits)) {
			fmt.Printf("Recording clause %d: The final resolvent is wrong.\n", pFinal.Id)
			Inta_ManPrintClause(p, pConflict)
			Inta_ManPrintResolvent(p.vResLits)
			Inta_ManPrintClause(p, pFinal)
		}

		// if there are literals in the clause that are not in the resolvent
		// it means that the derived resolvent is stronger than the clause
		// we can replace the clause with the resolvent by removing these literals
		if int32(len(p.vResLits)) != pFinal.nLits {
			for v1 = 0; v1 < pFinal.nLits; v1++ {
				for v2 = 0; v2 < int32(len(p.vResLits)); v2++ {
					if pFinal.pLits[v1] == p.vResLits[v2] {
						break
					}
				}
				if v2 < int32(len(p.vResLits)) {
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
			assert(int32(len(p.vResLits)) == pFinal.nLits, "Vec_IntSize(p->vResLits) == (int)pFinal->nLits")
		}
	}
	p.timeTrace += util.Abc_Clock() - clk

	// return the proof pointer
	if p.pCnf.nClausesA != 0 {
		//        Inta_ManPrintInterOne( p, pFinal );
	}
	Inta_ManProofSet(p, pFinal, p.Counter)
	// make sure the same proof ID is not asssigned to two consecutive clauses
	assert(p.pProofNums[pFinal.Id-1] != p.Counter, "p->pProofNums[pFinal->Id-1] != p->Counter")
	//    if ( p->pProofNums[pFinal->Id] == p->pProofNums[pFinal->Id-1] )
	//        p->pProofNums[pFinal->Id] = p->pProofNums[pConflict->Id];
	return p.Counter
}

/*
*Function*************************************************************

	Synopsis    [Records the proof for one clause.]

**********************************************************************
*/
func Inta_ManProofRecordOne(p *Inta_Man_t, pClause *Sto_Cls_t) int32 {
	var pConflict *Sto_Cls_t
	var i int32

	// empty clause never ends up there
	assert(pClause.nLits > 0, "pClause->nLits > 0")
	if pClause.nLits == 0 {
		fmt.Printf("Error: Empty clause is attempted.\n")
	}

	assert(pClause.fRoot == 0, "!pClause->fRoot")
	assert(p.nTrailSize == p.nRootSize, "p->nTrailSize == p->nRootSize")

	// if any of the clause literals are already assumed
	// it means that the clause is redundant and can be skipped
	for i = 0; i < pClause.nLits; i++ {
		if p.pAssigns[lit_var(pClause.pLits[i])] == pClause.pLits[i] {
			return 1
		}
	}

	// add assumptions to the trail
	for i = 0; i < pClause.nLits; i++ {
		if Inta_ManEnqueue(p, lit_neg(pClause.pLits[i]), nil) == 0 {
			assert(false, "0") // impossible
			return 0
		}
	}

	// propagate the assumptions
	pConflict = Inta_ManPropagate(p, p.nRootSize)
	if pConflict == nil {
		assert(false, "0") // cannot prove
		return 0
	}

	// skip the clause if it is weaker or the same as the conflict clause
	if pClause.nLits >= pConflict.nLits {
		// check if every literal of conflict clause can be found in the given clause
		var j int32
		for i = 0; i < pConflict.nLits; i++ {
			for j = 0; j < pClause.nLits; j++ {
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
			Inta_ManCancelUntil(p, p.nRootSize)
			return 1
		}
	}

	// construct the proof
	Inta_ManProofTraceOne(p, pConflict, pClause)

	// undo to the root level
	Inta_ManCancelUntil(p, p.nRootSize)

	// add large clauses to the watched lists
	if pClause.nLits > 1 {
		Inta_ManWatchClause(p, pClause, pClause.pLits[0])
		Inta_ManWatchClause(p, pClause, pClause.pLits[1])
		return 1
	}
	assert(pClause.nLits == 1, "pClause->nLits == 1")

	// if the clause proved is unit, add it and propagate
	if Inta_ManEnqueue(p, pClause.pLits[0], pClause) == 0 {
		assert(false, "0") // impossible
		return 0
	}

	// propagate the assumption
	pConflict = Inta_ManPropagate(p, p.nRootSize)
	if pConflict != nil {
		// construct the proof
		Inta_ManProofTraceOne(p, pConflict, p.pCnf.pEmpty)
		//        if ( p->fVerbose )
		//            printf( "Found last conflict after adding unit clause number %d!\n", pClause->Id );
		return 0
	}

	// update the root level
	p.nRootSize = p.nTrailSize
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Propagate the root clauses.]

**********************************************************************
*/
func Inta_ManProcessRoots(p *Inta_Man_t) int32 {
	var pClause *Sto_Cls_t
	var Counter int32

	// make sure the root clauses are preceeding the learnt clauses
	Counter = 0
	for pClause = p.pCnf.pHead; pClause != nil; pClause = pClause.pNext {
		assert((pClause.fA != 0) == (Counter < p.pCnf.nClausesA), "(int)pClause->fA == (Counter < (int)p->pCnf->nClausesA)")
		assert((pClause.fRoot != 0) == (Counter < p.pCnf.nRoots), "(int)pClause->fRoot == (Counter < (int)p->pCnf->nRoots)")
		Counter++
	}
	assert(p.pCnf.nClauses == Counter, "p->pCnf->nClauses == Counter")

	// make sure the last clause if empty
	assert(p.pCnf.pTail.nLits == 0, "p->pCnf->pTail->nLits == 0")

	// go through the root unit clauses
	p.nTrailSize = 0
	for pClause = p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		// create watcher lists for the root clauses
		if pClause.nLits > 1 {
			Inta_ManWatchClause(p, pClause, pClause.pLits[0])
			Inta_ManWatchClause(p, pClause, pClause.pLits[1])
		}
		// empty clause and large clauses
		if pClause.nLits != 1 {
			continue
		}
		// unit clause
		assert(lit_check(pClause.pLits[0], p.pCnf.nVars) != 0, "lit_check(pClause->pLits[0], p->pCnf->nVars)")
		if Inta_ManEnqueue(p, pClause.pLits[0], pClause) == 0 {
			// detected root level conflict
			//            printf( "Error in Inta_ManProcessRoots(): Detected a root-level conflict too early!\n" );
			//            assert( 0 );
			// detected root level conflict
			Inta_ManProofTraceOne(p, pClause, p.pCnf.pEmpty)
			if p.fVerbose != 0 {
				fmt.Printf("Found root level conflict!\n")
			}
			return 0
		}
	}

	// propagate the root unit clauses
	pClause = Inta_ManPropagate(p, 0)
	if pClause != nil {
		// detected root level conflict
		Inta_ManProofTraceOne(p, pClause, p.pCnf.pEmpty)
		if p.fVerbose != 0 {
			fmt.Printf("Found root level conflict!\n")
		}
		return 0
	}

	// set the root level
	p.nRootSize = p.nTrailSize
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Records the proof.]

**********************************************************************
*/
func Inta_ManPrepareInter(p *Inta_Man_t) {
	var Var, VarAB, v int32

	// set interpolants for root clauses
	for pClause := p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		if pClause.fA == 0 { // clause of B
			Inta_ManAigFill(p, Inta_ManAigRead(p, pClause))
			//            Inta_ManPrintInterOne( p, pClause );
			continue
		}
		// clause of A
		Inta_ManAigClear(p, Inta_ManAigRead(p, pClause))
		for v = 0; v < pClause.nLits; v++ {
			Var = lit_var(pClause.pLits[v])
			if p.pVarTypes[Var] < 0 { // global var
				VarAB = -p.pVarTypes[Var] - 1
				assert(VarAB >= 0 && VarAB < int32(len(p.vVarsAB)), "VarAB >= 0 && VarAB < Vec_IntSize(p->vVarsAB)")
				if lit_sign(pClause.pLits[v]) != 0 { // negative var
					Inta_ManAigOrNotVar(p, Inta_ManAigRead(p, pClause), VarAB)
				} else {
					Inta_ManAigOrVar(p, Inta_ManAigRead(p, pClause), VarAB)
				}
			}
		}
		//        Inta_ManPrintInterOne( p, pClause );
	}
}

/*
*Function*************************************************************

	Synopsis    [Computes interpolant for the given CNF.]

	Description [Takes the interpolation manager, the CNF deriving by the SAT
	solver, which includes ClausesA, ClausesB, and learned clauses. Additional
	arguments are the vector of variables common to AB and the verbosiness flag.
	Returns the AIG manager with a single output, containing the interpolant.]

**********************************************************************
*/
func Inta_ManInterpolate(p *Inta_Man_t, pCnf *Sto_Man_t, TimeToStop int64, vVarsAB []int32, fVerbose int32) *aig.Graph {
	var pRes *aig.Graph
	var pObj aig.Edge
	var pClause *Sto_Cls_t
	RetValue := int32(1)
	clkTotal := util.Abc_Clock()

	if TimeToStop != 0 && p.now() > TimeToStop {
		return nil
	}

	// check that the CNF makes sense
	assert(pCnf.nVars > 0 && pCnf.nClauses > 0, "pCnf->nVars > 0 && pCnf->nClauses > 0")
	p.pCnf = pCnf
	p.fVerbose = fVerbose
	p.vVarsAB = vVarsAB
	pRes = aig.Aig_ManStart(10000)
	p.pAig = pRes
	aig.Aig_IthVar(p.pAig, int32(len(p.vVarsAB))-1)

	// adjust the manager
	Inta_ManResize(p)

	// prepare the interpolant computation
	Inta_ManPrepareInter(p)

	// construct proof for each clause
	// start the proof
	if p.fProofWrite != 0 {
		p.pFile, _ = os.Create("proof.cnf_")
		p.Counter = 0
	}

	// write the root clauses
	for pClause = p.pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		Inta_ManProofWriteOne(p, pClause)
		if TimeToStop != 0 && p.now() > TimeToStop {
			aig.Aig_ManStop(pRes)
			p.pAig = nil
			return nil
		}
	}

	// propagate root level assignments
	if Inta_ManProcessRoots(p) != 0 {
		// if there is no conflict, consider learned clauses
		for pClause = p.pCnf.pHead; pClause != nil; pClause = pClause.pNext {
			if pClause.fRoot != 0 {
				continue
			}
			if Inta_ManProofRecordOne(p, pClause) == 0 {
				RetValue = 0
				break
			}
			if TimeToStop != 0 && p.now() > TimeToStop {
				aig.Aig_ManStop(pRes)
				p.pAig = nil
				return nil
			}
		}
	}
	_ = RetValue

	// stop the proof
	if p.fProofWrite != 0 {
		p.pFile.Close()
		//        Sat_ProofChecker( "proof.cnf_" );
		p.pFile = nil
	}

	if fVerbose != 0 {
		//        ABC_PRT( "Interpo", Abc_Clock() - clkTotal );
		fmt.Printf("Vars = %d. Roots = %d. Learned = %d. Resol steps = %d.  Ave = %.2f.  Mem = %.2f MB  ",
			p.pCnf.nVars, p.pCnf.nRoots, p.pCnf.nClauses-p.pCnf.nRoots, p.Counter,
			1.0*float64(p.Counter-p.pCnf.nRoots)/float64(p.pCnf.nClauses-p.pCnf.nRoots),
			1.0*float64(Sto_ManMemoryReport(p.pCnf))/float64(1<<20))
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clkTotal)
		p.timeTotal += util.Abc_Clock() - clkTotal
	}

	pObj = *Inta_ManAigRead(p, p.pCnf.pTail)
	aig.Aig_ObjCreateCo(pRes, pObj)
	aig.Aig_ManCleanup(pRes)

	p.pAig = nil
	return pRes
}

// Inta_ManInterpolateWithClock is Inta_ManInterpolate with TimeToStop
// measured by clock instead of Abc_Clock.
func Inta_ManInterpolateWithClock(p *Inta_Man_t, pCnf *Sto_Man_t, TimeToStop int64, vVarsAB []int32, fVerbose int32, clock func() int64) *aig.Graph {
	p.clock = clock
	defer func() { p.clock = nil }()
	return Inta_ManInterpolate(p, pCnf, TimeToStop, vVarsAB, fVerbose)
}

/*
*Function*************************************************************

	Synopsis    []

**********************************************************************
*/
func Inta_ManDeriveClauses(pMan *Inta_Man_t, pCnf *Sto_Man_t, fClausesA int32) *aig.Graph {
	var p *aig.Graph
	var pMiter, pSum, pLit aig.Edge
	var Var, VarAB, v int32
	p = aig.Aig_ManStart(10000)
	pMiter = aig.Aig_ManConst1(p)
	for pClause := pCnf.pHead; pClause != nil && pClause.fRoot != 0; pClause = pClause.pNext {
		if fClausesA^int32(pClause.fA) != 0 { // clause of B
			continue
		}
		// clause of A
		pSum = aig.Aig_ManConst0(p)
		for v = 0; v < pClause.nLits; v++ {
			Var = lit_var(pClause.pLits[v])
			if pMan.pVarTypes[Var] < 0 { // global var
				VarAB = -pMan.pVarTypes[Var] - 1
				assert(VarAB >= 0 && VarAB < int32(len(pMan.vVarsAB)), "VarAB >= 0 && VarAB < Vec_IntSize(pMan->vVarsAB)")
				pLit = aig.Aig_NotCond(aig.Aig_IthVar(p, VarAB), lit_sign(pClause.pLits[v]))
			} else {
				pLit = aig.Aig_NotCond(aig.Aig_IthVar(p, int32(len(pMan.vVarsAB))+1+Var), lit_sign(pClause.pLits[v]))
			}
			pSum = aig.Aig_Or(p, pSum, pLit)
		}
		pMiter = aig.Aig_And(p, pMiter, pSum)
	}
	aig.Aig_ObjCreateCo(p, pMiter)
	return p
}
