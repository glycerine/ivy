package abc

// Line-by-line port of src/base/abci/abcSat.c (Abc_NtkMiterSat and the CNF
// construction for strashed and BDD-logic networks) on top of the port of
// ABC's sat_solver (sat/bsat, satSolver.c). It lives in base/abc because
// opt/fret calls it too; base/abci re-exports it.

import (
	"bufio"
	"fmt"
	"os"

	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// nMuxes is abcSat.c's static MUX counter.
var nMuxes int32

// satVarOf reads (int)(ABC_PTRINT_T)pObj->pCopy: the SAT variable stored in
// CopyData, with NULL (nil) reading as variable 0.
func satVarOf(o *Object) int32 {
	switch v := o.CopyData.(type) {
	case nil:
		return 0
	case int32:
		return v
	}
	panic("SAT variable missing from ABC object")
}

// satToLitCond is toLitCond(v, c).
func satToLitCond(v int32, c int32) bsat.Literal {
	if c != 0 {
		return bsat.Negative(v)
	}
	return bsat.Positive(v)
}

func satIsCompl(e Edge) int32 { return abcBool(e.Complement) }

// Abc_NtkMiterSat attempts to solve the miter using an internal SAT
// sat_solver. Returns -1 if timed out; 0 if SAT; 1 if UNSAT.
func Abc_NtkMiterSat(pNtk *Abc_Ntk_t, nConfLimit, nInsLimit int64, fVerbose int32, pNumConfs, pNumInspects *int64) int32 {
	var RetValue int32

	if pNumConfs != nil {
		*pNumConfs = 0
	}
	if pNumInspects != nil {
		*pNumInspects = 0
	}

	if Abc_NtkLatchNum(pNtk) != 0 {
		panic("Abc_NtkMiterSat: Abc_NtkLatchNum(pNtk) == 0")
	}

	// load clauses into the sat_solver
	created := Abc_NtkMiterSatCreate(pNtk, 0)
	if created == nil {
		return 1
	}
	pSat := created.(*bsat.Solver)

	// simplify the problem
	status := pSat.SimplifyC()
	if status == 0 {
		pSat.Delete()
		return 1
	}

	// solve the miter
	if fVerbose != 0 {
		pSat.SetVerbosityC(1)
	}
	status = pSat.SolveC(nil, nConfLimit, nInsLimit, 0, 0)
	switch bsat.Status(status) {
	case bsat.Unknown:
		RetValue = -1
	case bsat.Satisfiable:
		RetValue = 0
	case bsat.Unsatisfiable:
		RetValue = 1
	default:
		panic("Abc_NtkMiterSat: unexpected solver status")
	}

	// if the problem is SAT, get the counterexample
	if bsat.Status(status) == bsat.Satisfiable {
		vCiIds := Abc_NtkGetCiSatVarNums(pNtk)
		nSize := int32(vec.Vec_IntSize(vCiIds))
		pNtk.Model = bsat.Sat_SolverGetModel(pSat, vec.Vec_IntArray(vCiIds), nSize)[:nSize]
		vec.Vec_IntFree(vCiIds)
	}
	// free the sat_solver
	if fVerbose != 0 {
		bsat.Sat_SolverPrintStats(os.Stdout, pSat)
	}

	st := pSat.Statistics()
	if pNumConfs != nil {
		*pNumConfs = int64(int32(st.Conflicts))
	}
	if pNumInspects != nil {
		*pNumInspects = int64(int32(st.Inspections))
	}

	// sat_solver_store_write( pSat, "trace.cnf" ); sat_solver_store_free( pSat );
	if pStore := pSat.StoreReleaseC(); pStore != nil {
		satStoreDumpClauses(pStore, "trace.cnf")
		bsat.Sto_ManFree(pStore)
	}

	pSat.Delete()
	return RetValue
}

// satStoreDumpClauses is Sto_ManDumpClauses (satStore.c) with a buffered
// writer, as C's stdio buffers its fprintf calls.
func satStoreDumpClauses(p *bsat.Sto_Man_t, pFileName string) {
	pFile, err := os.Create(pFileName)
	if err != nil {
		fmt.Printf("Error: Cannot open output file (%s).\n", pFileName)
		return
	}
	w := bufio.NewWriter(pFile)
	if err = bsat.WriteClauseStore(w, p); err == nil {
		err = w.Flush()
	}
	if closeErr := pFile.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Println(err)
	}
}

// Abc_NtkGetCiSatVarNums returns the array of CI SAT variables.
func Abc_NtkGetCiSatVarNums(pNtk *Abc_Ntk_t) *vec.Vec_Int_t {
	vCiIds := vec.Vec_IntAlloc(int(Abc_NtkCiNum(pNtk)))
	for _, pObj := range pNtk.Inputs() {
		vec.Vec_IntPush(vCiIds, satVarOf(pObj))
	}
	return vCiIds
}

// Abc_NtkClauseTriv adds the trivial clause.
func Abc_NtkClauseTriv(pSat *bsat.Solver, pNode Edge, vVars []bsat.Literal) int32 {
	vVars = append(vVars[:0], satToLitCond(satVarOf(pNode.Object), satIsCompl(pNode)))
	return pSat.AddClauseC(vVars)
}

// Abc_NtkClauseTop adds the clause asserting one of the outputs.
func Abc_NtkClauseTop(pSat *bsat.Solver, vNodes []Edge, vVars []bsat.Literal) int32 {
	vVars = vVars[:0]
	for _, pNode := range vNodes {
		vVars = append(vVars, satToLitCond(satVarOf(pNode.Object), satIsCompl(pNode)))
	}
	return pSat.AddClauseC(vVars)
}

// Abc_NtkClauseAnd adds the clauses of the multi-input AND gate.
func Abc_NtkClauseAnd(pSat *bsat.Solver, pNode *Object, vSuper []Edge, vVars []bsat.Literal) int32 {
	if Abc_ObjIsNode(pNode) == 0 {
		panic("Abc_NtkClauseAnd: Abc_ObjIsNode( pNode )")
	}
	Var := satVarOf(pNode)
	for _, e := range vSuper {
		fComp1 := satIsCompl(e)
		Var1 := satVarOf(e.Object)
		// suppose the AND-gate is A * B = C
		// add !A => !C   or   A + !C
		vVars = append(vVars[:0], satToLitCond(Var1, fComp1), satToLitCond(Var, 1))
		if pSat.AddClauseC(vVars) == 0 {
			return 0
		}
	}
	// add A & B => C   or   !A + !B + C
	vVars = vVars[:0]
	for _, e := range vSuper {
		fComp1 := satIsCompl(e)
		Var1 := satVarOf(e.Object)
		vVars = append(vVars, satToLitCond(Var1, 1^fComp1))
	}
	vVars = append(vVars, satToLitCond(Var, 0))
	return pSat.AddClauseC(vVars)
}

// Abc_NtkClauseMux adds the clauses of f = ITE(c, t, e).
func Abc_NtkClauseMux(pSat *bsat.Solver, pNode, pNodeC *Object, pNodeT, pNodeE Edge, vVars []bsat.Literal) int32 {
	if Abc_NodeIsMuxType(pNode) == 0 {
		panic("Abc_NtkClauseMux: Abc_NodeIsMuxType( pNode )")
	}
	// get the variable numbers
	VarF := satVarOf(pNode)
	VarI := satVarOf(pNodeC)
	VarT := satVarOf(pNodeT.Object)
	VarE := satVarOf(pNodeE.Object)
	// get the complementation flags
	fCompT := satIsCompl(pNodeT)
	fCompE := satIsCompl(pNodeE)

	// f = ITE(i, t, e)
	// i' + t' + f
	// i' + t  + f'
	// i  + e' + f
	// i  + e  + f'
	vVars = append(vVars[:0], satToLitCond(VarI, 1), satToLitCond(VarT, 1^fCompT), satToLitCond(VarF, 0))
	if pSat.AddClauseC(vVars) == 0 {
		return 0
	}
	vVars = append(vVars[:0], satToLitCond(VarI, 1), satToLitCond(VarT, 0^fCompT), satToLitCond(VarF, 1))
	if pSat.AddClauseC(vVars) == 0 {
		return 0
	}
	vVars = append(vVars[:0], satToLitCond(VarI, 0), satToLitCond(VarE, 1^fCompE), satToLitCond(VarF, 0))
	if pSat.AddClauseC(vVars) == 0 {
		return 0
	}
	vVars = append(vVars[:0], satToLitCond(VarI, 0), satToLitCond(VarE, 0^fCompE), satToLitCond(VarF, 1))
	if pSat.AddClauseC(vVars) == 0 {
		return 0
	}

	if VarT == VarE {
		return 1
	}

	// two additional clauses
	// t' & e' -> f'       t  + e   + f'
	// t  & e  -> f        t' + e'  + f
	vVars = append(vVars[:0], satToLitCond(VarT, 0^fCompT), satToLitCond(VarE, 0^fCompE), satToLitCond(VarF, 1))
	if pSat.AddClauseC(vVars) == 0 {
		return 0
	}
	vVars = append(vVars[:0], satToLitCond(VarT, 1^fCompT), satToLitCond(VarE, 1^fCompE), satToLitCond(VarF, 0))
	return pSat.AddClauseC(vVars)
}

// Abc_NtkCollectSupergate_rec collects the nodes to be combined into one
// multi-input AND gate.
func Abc_NtkCollectSupergate_rec(pNode Edge, vSuper *[]Edge, fFirst, fStopAtMux int32) int32 {
	// check if the node is visited
	if pNode.Object.MarkB {
		// check if the node occurs in the same polarity
		for _, e := range *vSuper {
			if e == pNode {
				return 1
			}
		}
		// check if the node is present in the opposite polarity
		for _, e := range *vSuper {
			if e == pNode.Not() {
				return -1
			}
		}
		panic("Abc_NtkCollectSupergate_rec: marked node is not in the supergate")
	}
	// if the new node is complemented or a PI, another gate begins
	if fFirst == 0 {
		if pNode.Complement || Abc_ObjIsNode(pNode.Object) == 0 || Abc_ObjFanoutNum(pNode.Object) > 1 || (fStopAtMux != 0 && Abc_NodeIsMuxType(pNode.Object) != 0) {
			*vSuper = append(*vSuper, pNode)
			pNode.Object.MarkB = true
			return 0
		}
	}
	if pNode.Complement || Abc_ObjIsNode(pNode.Object) == 0 {
		panic("Abc_NtkCollectSupergate_rec: expected a regular internal node")
	}
	// go through the branches
	RetValue1 := Abc_NtkCollectSupergate_rec(Abc_ObjChild0(pNode.Object), vSuper, 0, fStopAtMux)
	RetValue2 := Abc_NtkCollectSupergate_rec(Abc_ObjChild1(pNode.Object), vSuper, 0, fStopAtMux)
	if RetValue1 == -1 || RetValue2 == -1 {
		return -1
	}
	// return 1 if at least one branch has a duplicate
	return abcBool(RetValue1 != 0 || RetValue2 != 0)
}

// Abc_NtkCollectSupergate returns in *vNodes the implication supergate of
// pNode; empty if it contains a node and its complement.
func Abc_NtkCollectSupergate(pNode *Object, fStopAtMux int32, vNodes *[]Edge) {
	// collect the nodes in the implication supergate
	*vNodes = (*vNodes)[:0]
	RetValue := Abc_NtkCollectSupergate_rec(pNode.Signal(), vNodes, 1, fStopAtMux)
	if len(*vNodes) <= 1 {
		panic("Abc_NtkCollectSupergate: vNodes->nSize > 1")
	}
	// unmark the visited nodes
	for _, e := range *vNodes {
		e.Object.MarkB = false
	}
	// if we found the node and its complement in the same implication supergate,
	// return empty set of nodes (meaning that we should use constant-0 node)
	if RetValue == -1 {
		*vNodes = (*vNodes)[:0]
	}
}

// Abc_NtkNodeFactor computes the factor of the node.
func Abc_NtkNodeFactor(pObj *Object, nLevelMax int32) int32 {
	if int32(pObj.Level) > nLevelMax {
		panic("Abc_NtkNodeFactor: (int)pObj->Level <= nLevelMax")
	}
	return int32(100000000.0 * (1 + 0.01*float64(pObj.Level)))
}

// Abc_NtkMiterSatCreateInt sets up the SAT sat_solver for a strashed network.
func Abc_NtkMiterSatCreateInt(pSat *bsat.Solver, pNtk *Abc_Ntk_t) int32 {
	fUseMuxes := int32(1)

	if Abc_NtkIsStrash(pNtk) == 0 {
		panic("Abc_NtkMiterSatCreateInt: Abc_NtkIsStrash(pNtk)")
	}

	// clean the CI node pointers
	for _, pNode := range pNtk.Inputs() {
		pNode.CopyData = nil
	}

	// start the data structures
	vNodes := make([]*Object, 0, 1000) // the nodes corresponding to vars in the sat_solver
	vSuper := make([]Edge, 0, 100)     // the nodes belonging to the given implication supergate
	vVars := make([]bsat.Literal, 0, 100)

	addNode := func(pFanin *Object) {
		if !pFanin.MarkA {
			pFanin.MarkA = true
			pFanin.CopyData = int32(len(vNodes))
			vNodes = append(vNodes, pFanin)
		}
	}

	// add the clause for the constant node
	pConst := Abc_AigConst1(pNtk).Object
	pConst.MarkA = true
	pConst.CopyData = int32(len(vNodes))
	vNodes = append(vNodes, pConst)
	Abc_NtkClauseTriv(pSat, pConst.Signal(), vVars)

	// collect the nodes that need clauses and top-level assignments
	vSuper = vSuper[:0]
	for _, pNode := range pNtk.Outputs() {
		// create the node's variable
		addNode(Abc_ObjFanin0(pNode))
		// add the trivial clause
		vSuper = append(vSuper, Abc_ObjChild0(pNode))
	}
	if Abc_NtkClauseTop(pSat, vSuper, vVars) == 0 {
		return 0
	}

	// add the clauses
	for i := 0; i < len(vNodes); i++ {
		pNode := vNodes[i]
		if Abc_AigNodeIsAnd(pNode) == 0 {
			continue
		}
		if fUseMuxes != 0 && Abc_NodeIsMuxType(pNode) != 0 {
			nMuxes++

			var pNodeT, pNodeE Edge
			pNodeC := Abc_NodeRecognizeMux(pNode, &pNodeT, &pNodeE)
			vSuper = append(vSuper[:0], pNodeC.Signal(), pNodeT, pNodeE)
			// add the fanin nodes to explore
			for _, e := range vSuper {
				addNode(e.Object)
			}
			// add the clauses
			if Abc_NtkClauseMux(pSat, pNode, pNodeC, pNodeT, pNodeE, vVars) == 0 {
				return 0
			}
		} else {
			// get the supergate
			Abc_NtkCollectSupergate(pNode, fUseMuxes, &vSuper)
			// add the fanin nodes to explore
			for _, e := range vSuper {
				addNode(e.Object)
			}
			// add the clauses
			if len(vSuper) == 0 {
				if Abc_NtkClauseTriv(pSat, pNode.Signal().Not(), vVars) == 0 {
					return 0
				}
			} else {
				if Abc_NtkClauseAnd(pSat, pNode, vSuper, vVars) == 0 {
					return 0
				}
			}
		}
	}
	return 1
}

// Abc_NtkMiterSatCreate sets up the SAT sat_solver; it returns a
// *bsat.Solver, or nil when the CNF is trivially UNSAT.
func Abc_NtkMiterSatCreate(pNtk *Abc_Ntk_t, fAllPrimes int32) any {
	if Abc_NtkIsStrash(pNtk) == 0 && Abc_NtkIsBddLogic(pNtk) == 0 {
		panic("Abc_NtkMiterSatCreate: Abc_NtkIsStrash(pNtk) || Abc_NtkIsBddLogic(pNtk)")
	}
	if Abc_NtkIsBddLogic(pNtk) != 0 {
		if pSat := Abc_NtkMiterSatCreateLogic(pNtk, fAllPrimes); pSat != nil {
			return pSat
		}
		return nil
	}

	nMuxes = 0
	pSat := bsat.NewSolver()
	RetValue := Abc_NtkMiterSatCreateInt(pSat, pNtk)
	pSat.StoreMarkRootsC()

	for _, pNode := range pNtk.Objects() {
		pNode.MarkA = false
	}
	if RetValue == 0 {
		pSat.Delete()
		return nil
	}
	return pSat
}

// Abc_NodeAddClauses adds the clauses of an internal node of a BDD logic
// network from the SOPs of both phases; variables are object IDs.
func Abc_NodeAddClauses(pSat *bsat.Solver, pSop0, pSop1 string, pNode *Object, vVars []bsat.Literal) int32 {
	nFanins := pNode.FaninNum()

	root := pNode.Data.(*core.DdNode)
	if core.Cudd_Regular(root) == core.Cudd_ReadOne(pNode.network.FunctionManager.(*core.DdManager)) {
		vVars = append(vVars[:0], satToLitCond(int32(pNode.ID), abcBool(core.Cudd_IsComplement(root))))
		if pSat.AddClauseC(vVars) == 0 {
			fmt.Printf("The CNF is trivially UNSAT.\n")
			return 0
		}
		return 1
	}

	addPhase := func(pSop string, fCompl int32) int32 {
		for c := 0; ; c++ {
			// get the cube
			off := c * (nFanins + 3)
			if off >= len(pSop) || pSop[off] == 0 {
				break
			}
			pCube := pSop[off:]
			// add the clause
			vVars = vVars[:0]
			for i := 0; i < nFanins; i++ {
				pFanin := pNode.fanins[i].Object
				if pCube[i] == '0' {
					vVars = append(vVars, satToLitCond(int32(pFanin.ID), 0))
				} else if pCube[i] == '1' {
					vVars = append(vVars, satToLitCond(int32(pFanin.ID), 1))
				}
			}
			vVars = append(vVars, satToLitCond(int32(pNode.ID), fCompl))
			if pSat.AddClauseC(vVars) == 0 {
				fmt.Printf("The CNF is trivially UNSAT.\n")
				return 0
			}
		}
		return 1
	}
	// add clauses for the negative phase
	if addPhase(pSop0, 1) == 0 {
		return 0
	}
	// add clauses for the positive phase
	return addPhase(pSop1, 0)
}

// Abc_NodeAddClausesTop adds the clauses for the PO node.
func Abc_NodeAddClausesTop(pSat *bsat.Solver, pNode *Object, vVars []bsat.Literal) int32 {
	pFanin := Abc_ObjFanin0(pNode)
	vFanin, vNode := int32(pFanin.ID), int32(pNode.ID)
	var first, second [2]bsat.Literal
	if Abc_ObjFaninC0(pNode) != 0 {
		first = [2]bsat.Literal{satToLitCond(vFanin, 0), satToLitCond(vNode, 0)}
		second = [2]bsat.Literal{satToLitCond(vFanin, 1), satToLitCond(vNode, 1)}
	} else {
		first = [2]bsat.Literal{satToLitCond(vFanin, 1), satToLitCond(vNode, 0)}
		second = [2]bsat.Literal{satToLitCond(vFanin, 0), satToLitCond(vNode, 1)}
	}
	vVars = append(vVars[:0], first[:]...)
	if pSat.AddClauseC(vVars) == 0 {
		fmt.Printf("The CNF is trivially UNSAT.\n")
		return 0
	}
	vVars = append(vVars[:0], second[:]...)
	if pSat.AddClauseC(vVars) == 0 {
		fmt.Printf("The CNF is trivially UNSAT.\n")
		return 0
	}
	vVars = append(vVars[:0], satToLitCond(vNode, 0))
	if pSat.AddClauseC(vVars) == 0 {
		fmt.Printf("The CNF is trivially UNSAT.\n")
		return 0
	}
	return 1
}

// Abc_NtkMiterSatCreateLogic sets up the SAT sat_solver for a BDD logic
// network; nil means the CNF is trivially UNSAT.
func Abc_NtkMiterSatCreateLogic(pNtk *Abc_Ntk_t, fAllPrimes int32) *bsat.Solver {
	if Abc_NtkIsBddLogic(pNtk) == 0 {
		panic("Abc_NtkMiterSatCreateLogic: Abc_NtkIsBddLogic(pNtk)")
	}

	// transfer the IDs to the copy field
	for _, pNode := range pNtk.ObjectsOfKind(ObjectPI) {
		pNode.CopyData = int32(pNode.ID)
	}

	// start the data structures
	pSat := bsat.NewSolver()
	pSat.StoreAllocC()
	var vCube []byte
	vVars := make([]bsat.Literal, 0, 100)

	// add clauses for each internal nodes
	for _, pNode := range pNtk.ObjectsOfKind(ObjectNode) {
		// derive SOPs for both phases of the node
		var pSop0, pSop1 string
		Abc_NodeBddToCnf(pNode, nil, &vCube, fAllPrimes, &pSop0, &pSop1)
		// add the clauses to the sat_solver
		if Abc_NodeAddClauses(pSat, pSop0, pSop1, pNode, vVars) == 0 {
			pSat.Delete()
			return nil
		}
	}
	// add clauses for each PO
	for _, pNode := range pNtk.ObjectsOfKind(ObjectPO) {
		if Abc_NodeAddClausesTop(pSat, pNode, vVars) == 0 {
			pSat.Delete()
			return nil
		}
	}
	pSat.StoreMarkRootsC()
	return pSat
}
