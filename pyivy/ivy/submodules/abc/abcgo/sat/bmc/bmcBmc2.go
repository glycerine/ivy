package bmc

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

const (
	ABS_ZER int32 = absZero
	ABS_ONE int32 = absOne
	ABS_UND int32 = absUnknown
)

func Abs_ManSimInfoNot(value int32) int32          { return ternaryNot(value) }
func Abs_ManSimInfoAnd(value0, value1 int32) int32 { return ternaryAnd(value0, value1) }

// Abs_ManSimInfoGet reads an object's packed ternary value for one frame.
func Abs_ManSimInfoGet(frames *vec.Vec_Ptr_t, object *aig.Edge, frame int32) int32 {
	if frames == nil || object == nil || frame < 0 || int(frame) >= vec.Vec_PtrSize(frames) {
		return 0
	}
	words, ok := vec.Vec_PtrEntry(frames, int(frame)).([]uint32)
	if !ok || words == nil {
		return 0
	}
	id := uint32(object.NodeID())
	word := int(id >> 4)
	if word >= len(words) {
		return 0
	}
	return int32(packed2Get(words[word], id))
}

// Abs_ManSimInfoSet replaces an object's packed ternary value for one frame.
func Abs_ManSimInfoSet(frames *vec.Vec_Ptr_t, object *aig.Edge, frame, value int32) {
	if frames == nil || object == nil || frame < 0 || int(frame) >= vec.Vec_PtrSize(frames) {
		return
	}
	if value < ABS_ZER || value > ABS_UND {
		panic("bmc: invalid ternary state")
	}
	words, ok := vec.Vec_PtrEntry(frames, int(frame)).([]uint32)
	if !ok || words == nil {
		panic("bmc: simulation frame does not contain packed words")
	}
	id := uint32(object.NodeID())
	word := int(id >> 4)
	if word >= len(words) {
		panic("bmc: object outside simulation frame")
	}
	words[word] = packed2Set(words[word], id, uint32(value))
}

// Abs_ManTernarySimulate propagates the zero initial state and unknown primary
// inputs through the requested frames. It stops once a property is not proved
// false, matching the source routine's early-return contract.
func Abs_ManTernarySimulate(graph *aig.Graph, maxFrames, verbose int32) *vec.Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(0)
	if graph == nil || maxFrames <= 0 || aig.Aig_ManRegNum(graph) <= 0 {
		return result
	}
	inputs, outputs, nodes := graph.Inputs(), graph.Outputs(), graph.Nodes()
	regs := int(aig.Aig_ManRegNum(graph))
	if regs > len(inputs) || regs > len(outputs) {
		return result
	}
	primaryInputs := len(inputs) - regs
	primaryOutputs := len(outputs) - regs
	state := make([]int32, regs) // C's default initial state is all zero.
	for i := range state {
		state[i] = ABS_ZER
	}
	for frame := int32(0); frame < maxFrames; frame++ {
		values := make([]int32, len(nodes))
		values[0] = ABS_ONE
		for i, input := range inputs {
			if i < primaryInputs {
				values[input.NodeID()] = ABS_UND
			} else {
				values[input.NodeID()] = state[i-primaryInputs]
			}
		}
		for id, node := range nodes {
			switch node.Kind {
			case aig.Constant, aig.Input:
			case aig.And:
				values[id] = ternaryAnd(ternaryAIGValue(values, node.Fanins[0]), ternaryAIGValue(values, node.Fanins[1]))
			case aig.ExclusiveOr:
				left, right := ternaryAIGValue(values, node.Fanins[0]), ternaryAIGValue(values, node.Fanins[1])
				if left == ABS_UND || right == ABS_UND {
					values[id] = ABS_UND
				} else if left == right {
					values[id] = ABS_ZER
				} else {
					values[id] = ABS_ONE
				}
			case aig.Buffer, aig.Output:
				values[id] = ternaryAIGValue(values, node.Fanins[0])
			}
		}
		packed := make([]uint32, (len(nodes)+15)/16)
		for id, value := range values {
			if value != 0 {
				packed[id>>4] = packed2Set(packed[id>>4], uint32(id), uint32(value))
			}
		}
		vec.Vec_PtrPush(result, packed)
		proved := true
		for i := 0; i < primaryOutputs; i++ {
			if ternaryAIGValue(values, graph.Node(outputs[i]).Fanins[0]) != ABS_ZER {
				proved = false
				break
			}
		}
		if !proved {
			if verbose != 0 {
				fmt.Printf("Ternary simulation found a non-zero or unknown output in frame %d.\n", frame)
			}
			return result
		}
		for i := 0; i < regs; i++ {
			driver := graph.Node(outputs[primaryOutputs+i]).Fanins[0]
			state[i] = ternaryAIGValue(values, driver)
		}
	}
	if verbose != 0 {
		fmt.Printf("Ternary simulation proved all outputs false for %d frames.\n", vec.Vec_PtrSize(result))
	}
	return result
}

func ternaryAIGValue(values []int32, edge aig.Edge) int32 {
	value := values[edge.NodeID()]
	if edge.Complemented() && value != 0 {
		return ternaryNot(value)
	}
	return value
}

func Abs_ManFreeAray(frames *vec.Vec_Ptr_t) { vec.Vec_PtrFree(frames) }

// Saig_Bmc_t is bmcBmc2.c's BMC manager: the user's AIG is unrolled
// interval by interval into the frames manager pFrm, and each interval is
// re-derived as a small AIG whose CNF is stitched into one SAT solver.
type Saig_Bmc_t struct {
	// parameters
	nFramesMax  int32 // the max number of timeframes to consider
	nNodesMax   int32 // the max number of nodes to add
	nConfMaxOne int32 // the max number of conflicts at a target
	nConfMaxAll int32 // the max number of conflicts for all targets
	fVerbose    int32 // enables verbose output
	// AIG managers
	pAig     *aig.Graph // the user's AIG manager
	pFrm     *aig.Graph // Frames manager
	vVisited []int32    // nodes visited in Frames
	// node mapping
	nObjs    int32     // the largest number of an AIG object
	vAig2Frm [][]int32 // per frame, AIG object ID -> Frames literal (-1 if none)
	// SAT solver
	pSat        *bsat.Solver // SAT solver
	nSatVars    int32        // the number of used SAT variables
	vObj2Var    []int32      // mapping of frames objects in CNF variables
	nStitchVars int32
	// subproblems
	vTargets    []aig.Edge // targets to be solved in this interval
	iFramePrev  int32      // previous frame
	iFrameLast  int32      // last frame
	iOutputLast int32      // last output
	iFrameFail  int32      // failed frame
	iOutputFail int32      // failed output
	// pData of the Frames objects (their copies in the interval AIG)
	pData []aig.Edge
}

// saigBmcNull stands for the NULL object pointer.
const saigBmcNull = aig.InvalidEdge

// Saig_BmcObjFrame returns the Frames object of pObj in frame i; an
// object removed from Frames by Aig_ManCleanup reads as NULL and its
// mapping is reset to -1.
func Saig_BmcObjFrame(p *Saig_Bmc_t, pObj aig.Edge, i int32) aig.Edge {
	vFrame := p.vAig2Frm[i]
	iObjLit := vFrame[pObj.NodeID()]
	if iObjLit == -1 {
		return saigBmcNull
	}
	pRes := aig.Edge(iObjLit) &^ 1
	if !p.pFrm.ValidEdge(pRes) {
		vFrame[pObj.NodeID()] = -1
		return saigBmcNull
	}
	return aig.Aig_NotCond(pRes, iObjLit&1)
}

func Saig_BmcObjSetFrame(p *Saig_Bmc_t, pObj aig.Edge, i int32, pNode aig.Edge) {
	if int(i) == len(p.vAig2Frm) {
		vFrame := make([]int32, p.nObjs)
		for k := range vFrame {
			vFrame[k] = -1
		}
		p.vAig2Frm = append(p.vAig2Frm, vFrame)
	}
	assert(int(i) < len(p.vAig2Frm), "i < Vec_PtrSize(p->vAig2Frm)")
	iObjLit := int32(-1)
	if pNode != saigBmcNull {
		iObjLit = int32(pNode) // Abc_Var2Lit(Aig_ObjId(Aig_Regular(pNode)), Aig_IsComplement(pNode))
	}
	p.vAig2Frm[i][pObj.NodeID()] = iObjLit
}

func saigBmcNotCond(e aig.Edge, c int32) aig.Edge {
	if e == saigBmcNull {
		return e
	}
	return aig.Aig_NotCond(e, c)
}

func Saig_BmcObjChild0(p *Saig_Bmc_t, pObj aig.Edge, i int32) aig.Edge {
	return saigBmcNotCond(Saig_BmcObjFrame(p, aig.Aig_ObjFanin0(p.pAig, pObj), i), aig.Aig_ObjFaninC0(p.pAig, pObj))
}
func Saig_BmcObjChild1(p *Saig_Bmc_t, pObj aig.Edge, i int32) aig.Edge {
	return saigBmcNotCond(Saig_BmcObjFrame(p, aig.Aig_ObjFanin1(p.pAig, pObj), i), aig.Aig_ObjFaninC1(p.pAig, pObj))
}

func Saig_BmcSatNum(p *Saig_Bmc_t, pObj aig.Edge) int32 {
	if id := pObj.NodeID(); id < len(p.vObj2Var) {
		return p.vObj2Var[id]
	}
	return 0
}
func Saig_BmcSetSatNum(p *Saig_Bmc_t, pObj aig.Edge, num int32) {
	id := pObj.NodeID()
	if id >= len(p.vObj2Var) {
		p.vObj2Var = append(p.vObj2Var, make([]int32, id+1-len(p.vObj2Var))...)
	}
	p.vObj2Var[id] = num
}

// Saig_BmcManStart creates the manager. Satoko is not ported; the bsat
// solver is used for either setting of fUseSatoko.
func Saig_BmcManStart(pAig *aig.Graph, nFramesMax, nNodesMax, nConfMaxOne, nConfMaxAll, fVerbose, fUseSatoko int32) *Saig_Bmc_t {
	p := &Saig_Bmc_t{
		nFramesMax:  nFramesMax,
		nNodesMax:   nNodesMax,
		nConfMaxOne: nConfMaxOne,
		nConfMaxAll: nConfMaxAll,
		fVerbose:    fVerbose,
		pAig:        pAig,
		nObjs:       aig.Aig_ManObjNumMax(pAig),
	}
	// create node and variable mappings
	p.vObj2Var = make([]int32, p.nObjs)
	// start the AIG manager and map primary inputs
	p.pFrm = aig.Aig_ManStart(p.nObjs)
	for i := int32(0); i < saig.Saig_ManRegNum(pAig); i++ {
		Saig_BmcObjSetFrame(p, saig.Saig_ManLo(pAig, i), 0, aig.Aig_ManConst0(p.pFrm))
	}
	// create SAT solver
	p.nSatVars = 1
	lit := bsat.Positive(p.nSatVars)
	p.pSat = bsat.NewSolver()
	p.pSat.SetLearnedPolicy(10000, 5000, 75)
	p.pSat.SetNumVars(2000)
	p.pSat.AddClause(lit)
	Saig_BmcSetSatNum(p, aig.Aig_ManConst1(p.pFrm), p.nSatVars)
	p.nSatVars++
	// other data structures
	p.vTargets = make([]aig.Edge, 0, 1000)
	p.vVisited = make([]int32, 0, 1000)
	p.iOutputFail = -1
	p.iFrameFail = -1
	return p
}

func Saig_BmcManStop(p *Saig_Bmc_t) {
	if p.pSat != nil {
		p.pSat.Delete()
		p.pSat = nil
	}
	p.pFrm = nil
	p.vAig2Frm = nil
	p.vObj2Var = nil
	p.vTargets = nil
	p.vVisited = nil
	p.pData = nil
}

// Saig_BmcIntervalConstruct_rec performs the actual construction of the output.
func Saig_BmcIntervalConstruct_rec(p *Saig_Bmc_t, pObj aig.Edge, i int32, vVisited *[]int32) aig.Edge {
	pRes := Saig_BmcObjFrame(p, pObj, i)
	if pRes != saigBmcNull {
		return pRes
	}
	if saig.Saig_ObjIsPi(p.pAig, pObj) {
		pRes = aig.Aig_ObjCreateCi(p.pFrm)
	} else if saig.Saig_ObjIsLo(p.pAig, pObj) {
		pRes = Saig_BmcIntervalConstruct_rec(p, saig.Saig_ObjLoToLi(p.pAig, pObj), i-1, vVisited)
	} else if p.pAig.Kind(pObj) == aig.Output {
		Saig_BmcIntervalConstruct_rec(p, aig.Aig_ObjFanin0(p.pAig, pObj), i, vVisited)
		pRes = Saig_BmcObjChild0(p, pObj, i)
	} else {
		Saig_BmcIntervalConstruct_rec(p, aig.Aig_ObjFanin0(p.pAig, pObj), i, vVisited)
		if Saig_BmcObjChild0(p, pObj, i) == aig.Aig_ManConst0(p.pFrm) {
			pRes = aig.Aig_ManConst0(p.pFrm)
		} else {
			Saig_BmcIntervalConstruct_rec(p, aig.Aig_ObjFanin1(p.pAig, pObj), i, vVisited)
			pRes = aig.Aig_And(p.pFrm, Saig_BmcObjChild0(p, pObj, i), Saig_BmcObjChild1(p, pObj, i))
		}
	}
	assert(pRes != saigBmcNull, "pRes != NULL")
	Saig_BmcObjSetFrame(p, pObj, i, pRes)
	*vVisited = append(*vVisited, int32(pObj.NodeID()), i)
	return pRes
}

// Saig_BmcInterval adds new AIG nodes to the frames.
func Saig_BmcInterval(p *Saig_Bmc_t) {
	nNodes := aig.Aig_ManObjNum(p.pFrm)
	p.vTargets = p.vTargets[:0]
	p.iFramePrev = p.iFrameLast
	for ; p.iFrameLast < p.nFramesMax; p.iFrameLast, p.iOutputLast = p.iFrameLast+1, 0 {
		if p.iOutputLast == 0 {
			Saig_BmcObjSetFrame(p, aig.Aig_ManConst1(p.pAig), p.iFrameLast, aig.Aig_ManConst1(p.pFrm))
		}
		for ; p.iOutputLast < saig.Saig_ManPoNum(p.pAig); p.iOutputLast++ {
			if aig.Aig_ManObjNum(p.pFrm) >= nNodes+p.nNodesMax {
				return
			}
			p.vVisited = p.vVisited[:0]
			pTarget := Saig_BmcIntervalConstruct_rec(p, aig.Aig_ManCo(p.pAig, p.iOutputLast), p.iFrameLast, &p.vVisited)
			p.vTargets = append(p.vTargets, pTarget)
			aig.Aig_ObjCreateCo(p.pFrm, pTarget)
			aig.Aig_ManCleanup(p.pFrm) // it is not efficient to cleanup the whole manager!!!
			// check if the node is gone
			for k := 0; k+1 < len(p.vVisited); k += 2 {
				Saig_BmcObjFrame(p, aig.Aig_ManObj(p.pAig, p.vVisited[k]), p.vVisited[k+1])
			}
		}
	}
}

func Saig_BmcIntervalToAig_rec(p *Saig_Bmc_t, pNew *aig.Graph, pObj aig.Edge) aig.Edge {
	id := pObj.NodeID()
	if p.pData[id] != saigBmcNull {
		return p.pData[id]
	}
	p.vVisited = append(p.vVisited, int32(id))
	if Saig_BmcSatNum(p, pObj) != 0 || aig.Aig_ObjIsCi(p.pFrm, pObj) {
		if !aig.Aig_ObjIsCi(p.pFrm, pObj) {
			p.nStitchVars++
		}
		p.pData[id] = aig.Aig_ObjCreateCi(pNew)
		return p.pData[id]
	}
	Saig_BmcIntervalToAig_rec(p, pNew, aig.Aig_ObjFanin0(p.pFrm, pObj))
	Saig_BmcIntervalToAig_rec(p, pNew, aig.Aig_ObjFanin1(p.pFrm, pObj))
	assert(p.pData[id] == saigBmcNull, "pObj->pData == NULL")
	child0 := aig.Aig_NotCond(p.pData[aig.Aig_ObjFanin0(p.pFrm, pObj).NodeID()], aig.Aig_ObjFaninC0(p.pFrm, pObj))
	child1 := aig.Aig_NotCond(p.pData[aig.Aig_ObjFanin1(p.pFrm, pObj).NodeID()], aig.Aig_ObjFaninC1(p.pFrm, pObj))
	p.pData[id] = aig.Aig_And(pNew, child0, child1)
	return p.pData[id]
}

// Saig_BmcIntervalToAig creates the AIG of the newly added nodes.
func Saig_BmcIntervalToAig(p *Saig_Bmc_t) *aig.Graph {
	n := p.pFrm.NodeSlots()
	if len(p.pData) < n {
		p.pData = append(p.pData, make([]aig.Edge, n-len(p.pData))...)
	}
	for k := range p.pData {
		assert(p.pData[k] == saigBmcNull || p.pData[k] == 0, "pObj->pData == NULL")
		p.pData[k] = saigBmcNull
	}
	pNew := aig.Aig_ManStart(p.nNodesMax)
	p.pData[aig.Aig_ManConst1(p.pFrm).NodeID()] = aig.Aig_ManConst1(pNew)
	p.vVisited = p.vVisited[:0]
	p.vVisited = append(p.vVisited, int32(aig.Aig_ManConst1(p.pFrm).NodeID()))
	for _, pObj := range p.vTargets {
		pObjNew := Saig_BmcIntervalToAig_rec(p, pNew, aig.Aig_Regular(pObj))
		assert(aig.Aig_IsComplement(pObjNew) == 0, "!Aig_IsComplement(pObjNew)")
		aig.Aig_ObjCreateCo(pNew, pObjNew)
	}
	return pNew
}

// Saig_BmcLoadCnf derives CNF for the newly added nodes.
func Saig_BmcLoadCnf(p *Saig_Bmc_t, pCnf *cnf.Cnf_Dat_t) {
	for _, iObj := range p.vVisited {
		// get the new variable of this node
		pObjNew := p.pData[iObj]
		p.pData[iObj] = saigBmcNull
		VarNumNew := pCnf.VarNums[pObjNew.NodeID()]
		if VarNumNew == -1 {
			continue
		}
		// get the old variable of this node
		pObj := aig.Aig_ManObj(p.pFrm, iObj)
		VarNumOld := Saig_BmcSatNum(p, pObj)
		if VarNumOld == 0 {
			Saig_BmcSetSatNum(p, pObj, VarNumNew)
			continue
		}
		// add clauses connecting existing variables
		if !p.pSat.AddClause(bsat.Positive(VarNumOld), bsat.Negative(VarNumNew)) {
			assert(false, "sat_solver_addclause")
		}
		if !p.pSat.AddClause(bsat.Negative(VarNumOld), bsat.Positive(VarNumNew)) {
			assert(false, "sat_solver_addclause")
		}
	}
	// add CNF to the SAT solver
	clauses := cnf.Cnf_DataClauses(pCnf)
	i := 0
	for ; i < len(clauses); i++ {
		if !p.pSat.AddClause(clauses[i]...) {
			break
		}
	}
	if i < len(clauses) {
		fmt.Printf("SAT solver became UNSAT after adding clauses.\n")
	}
}

func Saig_BmcDeriveFailed(p *Saig_Bmc_t, iTargetFail int) {
	p.iOutputFail = p.iOutputLast
	p.iFrameFail = p.iFrameLast
	for k := len(p.vTargets); k > iTargetFail; k-- {
		if p.iOutputFail == 0 {
			p.iOutputFail = saig.Saig_ManPoNum(p.pAig)
			p.iFrameFail--
		}
		p.iOutputFail--
	}
}

func Saig_BmcGenerateCounterExample(p *Saig_Bmc_t) *util.Abc_Cex_t {
	nPis := saig.Saig_ManPiNum(p.pAig)
	// start the counter-example
	pCex := util.Abc_CexAlloc(aig.Aig_ManRegNum(p.pAig), nPis, p.iFrameFail+1)
	pCex.IFrame = p.iFrameFail
	pCex.IPo = p.iOutputFail
	// copy the bit data
	for f := int32(0); f <= p.iFrameFail; f++ {
		for i := int32(0); i < nPis; i++ {
			pObjFrm := Saig_BmcObjFrame(p, aig.Aig_ManCi(p.pAig, i), f)
			if pObjFrm == saigBmcNull {
				continue
			}
			iVarNum := Saig_BmcSatNum(p, pObjFrm)
			if iVarNum == 0 {
				continue
			}
			if value, _ := p.pSat.Value(iVarNum); value {
				util.Abc_InfoSetBit(pCex.PData, pCex.NRegs+nPis*f+i)
			}
		}
	}
	// verify the counter example
	if saig.Saig_ManVerifyCex(p.pAig, pCex) == 0 {
		fmt.Printf("Saig_BmcGenerateCounterExample(): Counter-example is invalid.\n")
		pCex = nil
	}
	return pCex
}

// Saig_BmcSolveTargets solves targets with the given resource limit; it
// returns l_True (1), l_False (-1) or l_Undef (0).
func Saig_BmcSolveTargets(p *Saig_Bmc_t, nStart int32, pnOutsSolved *int32) int32 {
	assert(len(p.vTargets) > 0, "Vec_PtrSize(p->vTargets) > 0")
	if p.pSat.NeedsPropagation() {
		RetValue := p.pSat.Simplify()
		assert(RetValue, "RetValue != 0")
	}
	for i, pObj := range p.vTargets {
		solved := *pnOutsSolved
		*pnOutsSolved++
		if solved/saig.Saig_ManPoNum(p.pAig) < nStart {
			continue
		}
		if p.nConfMaxAll != 0 && p.pSat.ConflictCount() > int64(p.nConfMaxAll) {
			return int32(bsat.Unknown)
		}
		VarNum := Saig_BmcSatNum(p, aig.Aig_Regular(pObj))
		Lit := bsat.Positive(VarNum)
		if aig.Aig_IsComplement(pObj) != 0 {
			Lit = bsat.Negative(VarNum)
		}
		RetValue := p.pSat.SolveWithSourceLimits([]bsat.Literal{Lit}, int64(p.nConfMaxOne), 0, 0, 0)
		if RetValue == bsat.Unsatisfiable { // unsat
			// add final unit clause
			Lit ^= 1
			status := p.pSat.AddClause(Lit)
			assert(status, "status")
			// add learned units
			for _, unit := range p.pSat.TakeUnitLiterals() {
				status = p.pSat.AddClause(unit)
				assert(status, "status")
			}
			// propagate units
			p.pSat.Compress()
			continue
		}
		if RetValue == bsat.Unknown { // undecided
			return int32(bsat.Unknown)
		}
		// generate counter-example
		Saig_BmcDeriveFailed(p, i)
		if cex := Saig_BmcGenerateCounterExample(p); cex != nil {
			if err := p.pAig.SetSequentialCounterexample(cex); err != nil {
				p.pAig.ClearSequentialCounterexample()
			}
		} else {
			p.pAig.ClearSequentialCounterexample()
		}
		return int32(bsat.Satisfiable)
	}
	return int32(bsat.Unsatisfiable)
}

func Saig_BmcAddTargetsAsPos(p *Saig_Bmc_t) {
	for _, pObj := range p.vTargets {
		aig.Aig_ObjCreateCo(p.pFrm, pObj)
	}
	aig.Aig_ManPrintStats(p.pFrm)
	aig.Aig_ManCleanup(p.pFrm)
	aig.Aig_ManPrintStats(p.pFrm)
}

// Saig_BmcPerform performs BMC with the given parameters (bmcBmc2.c).
func Saig_BmcPerform(pAig *aig.Graph, nStart, nFramesMax, nNodesMax, nTimeOut, nConfMaxOne, nConfMaxAll, fVerbose, fVerbOverwrite int32, piFrames *int32, fSilent, fUseSatoko int32) int32 {
	if pAig == nil {
		return -1
	}
	var nOutsSolved int32
	RetValue := int32(-1)
	var nTimeToStop int64
	if nTimeOut != 0 {
		nTimeToStop = int64(nTimeOut)*int64(time.Second) + util.Abc_Clock()
	}
	clk := util.Abc_Clock()
	clkTotal := util.Abc_Clock()
	Status := int32(-1)
	if fVerbose != 0 {
		fmt.Printf("Running \"bmc2\". AIG:  PI/PO/Reg = %d/%d/%d.  Node = %6d. Lev = %5d.\n",
			saig.Saig_ManPiNum(pAig), saig.Saig_ManPoNum(pAig), saig.Saig_ManRegNum(pAig),
			aig.Aig_ManNodeNum(pAig), aig.Aig_ManLevelNum(pAig))
		fmt.Printf("Params: FramesMax = %d. NodesDelta = %d. ConfMaxOne = %d. ConfMaxAll = %d.\n",
			nFramesMax, nNodesMax, nConfMaxOne, nConfMaxAll)
	}
	if nFramesMax == 0 {
		nFramesMax = 1000000000 // ABC_INFINITY
	}
	p := Saig_BmcManStart(pAig, nFramesMax, nNodesMax, nConfMaxOne, nConfMaxAll, fVerbose, fUseSatoko)
	// set runtime limit
	if nTimeOut != 0 {
		p.pSat.SetRuntimeDeadline(util.Abc_Clock, nTimeToStop)
	}
	for Iter := 0; ; Iter++ {
		// add new logic interval to frames
		Saig_BmcInterval(p)
		if len(p.vTargets) == 0 {
			break
		}
		// convert logic slice into new AIG
		pNew := Saig_BmcIntervalToAig(p)
		// derive CNF for the new AIG
		pCnf := cnf.Cnf_Derive(pNew, aig.Aig_ManCoNum(pNew))
		cnf.Cnf_DataLift(pCnf, p.nSatVars)
		p.nSatVars += cnf.Cnf_DataVarNum(pCnf)
		// add this CNF to the solver
		Saig_BmcLoadCnf(p, pCnf)
		cnf.Cnf_DataFree(pCnf)
		// solve the targets
		RetValue = Saig_BmcSolveTargets(p, nStart, &nOutsSolved)
		if fVerbose != 0 {
			fmt.Printf("%4d : F =%5d. O =%4d.  And =%8d. Var =%8d. Conf =%7d. ",
				Iter, p.iFrameLast, p.iOutputLast, aig.Aig_ManNodeNum(p.pFrm), p.nSatVars,
				int32(p.pSat.ConflictCount()))
			fmt.Printf("%4.0f MB", 4.0*float64(p.iFrameLast+1)*float64(p.nObjs)/float64(1<<20))
			fmt.Printf("%9.2f sec", float64(util.Abc_Clock()-clkTotal)/float64(time.Second))
			fmt.Printf("\n")
		}
		if RetValue != int32(bsat.Unsatisfiable) {
			break
		}
		// check the timeout
		if nTimeOut != 0 && util.Abc_Clock() > nTimeToStop {
			if fSilent == 0 {
				fmt.Printf("Reached timeout (%d seconds).\n", nTimeOut)
			}
			if piFrames != nil {
				*piFrames = p.iFrameLast - 1
			}
			Saig_BmcManStop(p)
			return Status
		}
	}
	if RetValue == int32(bsat.Satisfiable) {
		assert(p.iFrameFail*saig.Saig_ManPoNum(p.pAig)+p.iOutputFail+1 == nOutsSolved, "iFrameFail * PoNum + iOutputFail + 1 == nOutsSolved")
		if fSilent == 0 {
			fmt.Printf("Output %d of miter \"%s\" was asserted in frame %d. ", p.iOutputFail, pAig.Name, p.iFrameFail)
		}
		Status = 0
		if piFrames != nil {
			*piFrames = p.iFrameFail - 1
		}
	} else { // if ( RetValue == l_False || RetValue == l_Undef )
		if fSilent == 0 {
			fmt.Printf("No output failed in %d frames.  ", max(p.iFramePrev-1, 0))
		}
		if piFrames != nil {
			if p.iOutputLast > 0 {
				*piFrames = p.iFramePrev - 2
			} else {
				*piFrames = p.iFramePrev - 1
			}
		}
	}
	if fSilent == 0 {
		elapsed := float64(util.Abc_Clock()-clk) / float64(time.Second)
		if fVerbOverwrite != 0 {
			fmt.Printf("%s =%9.2f sec\r", "Time", elapsed)
		} else {
			fmt.Printf("%s =%9.2f sec\n", "Time", elapsed)
		}
		if RetValue != int32(bsat.Satisfiable) {
			if p.iFrameLast >= p.nFramesMax {
				fmt.Printf("Reached limit on the number of timeframes (%d).\n", p.nFramesMax)
			} else if p.nConfMaxAll != 0 && int32(p.pSat.ConflictCount()) > p.nConfMaxAll {
				fmt.Printf("Reached global conflict limit (%d).\n", p.nConfMaxAll)
			} else if nTimeOut != 0 && util.Abc_Clock() > nTimeToStop {
				fmt.Printf("Reached timeout (%d seconds).\n", nTimeOut)
			} else {
				fmt.Printf("Reached local conflict limit (%d).\n", p.nConfMaxOne)
			}
		}
	}
	Saig_BmcManStop(p)
	return Status
}
