package saig

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Aig_Gla1Man_t owns the incremental SAT state used by gate-level abstraction.
type Aig_Gla1Man_t struct {
	pAig       *aig.Graph
	nConfLimit int32
	nFramesMax int32
	fVerbose   int32
	nFrames    int32
	vIncluded  []int32
	pSat       *bsat.Solver
	lastFrames *aig.Graph
	lastCNF    *aigCNF
	lastFrame  int32
	startedAt  time.Time
	objToVec   []int32
	vecToVar   [][]int32
	varToInfo  [][2]int32
	assigned   []bool
	addedFrame [][]bool
	vPis       []int
	vPPis      []int
	vFlops     []int
	vNodes     []int
}

func Aig_Gla1AddConst(solver *bsat.Solver, variable, complement int32) int32 {
	lit := bsat.Positive(variable)
	if complement != 0 {
		lit ^= 1
	}
	if solver.AddClause(lit) {
		return 1
	}
	return 0
}

func Aig_Gla1AddBuffer(solver *bsat.Solver, v0, v1, complement int32) int32 {
	x, y := bsat.Positive(v0), bsat.Positive(v1)
	if complement != 0 {
		y ^= 1
	}
	if !solver.AddClause(x^1, y) || !solver.AddClause(x, y^1) {
		return 0
	}
	return 1
}

func Aig_Gla1AddNode(solver *bsat.Solver, result, v0, v1, c0, c1 int32) int32 {
	z, a, b := bsat.Positive(result), bsat.Positive(v0), bsat.Positive(v1)
	if c0 != 0 {
		a ^= 1
	}
	if c1 != 0 {
		b ^= 1
	}
	if !solver.AddClause(z^1, a) || !solver.AddClause(z^1, b) || !solver.AddClause(z, a^1, b^1) {
		return 0
	}
	return 1
}

func Aig_Gla1ManStart(src *aig.Graph, oldClasses *vec.Vec_Int_t, naiveCNF int32) *Aig_Gla1Man_t {
	if src == nil {
		return nil
	}
	included := make([]int32, len(src.Nodes()))
	if oldClasses != nil {
		limit := min(len(included), vec.Vec_IntSize(oldClasses))
		for i := 0; i < limit; i++ {
			included[i] = vec.Vec_IntEntry(oldClasses, i)
		}
	}
	if len(included) > 0 {
		included[0] = 1
	}
	solver := bsat.NewSolver()
	solver.NewVariable() // SAT variable zero is reserved as the unallocated marker.
	return &Aig_Gla1Man_t{pAig: src, nFrames: 32, vIncluded: included, pSat: solver, startedAt: time.Now()}
}

func Aig_Gla1ManStop(manager *Aig_Gla1Man_t) {
	if manager == nil {
		return
	}
	manager.pSat = nil
	manager.lastFrames = nil
	manager.lastCNF = nil
	manager.vIncluded = nil
}

func Aig_Gla1CollectAbstr(manager *Aig_Gla1Man_t) {
	if manager == nil || manager.pAig == nil {
		return
	}
	if len(manager.vIncluded) < len(manager.pAig.Nodes()) {
		manager.vIncluded = append(manager.vIncluded, make([]int32, len(manager.pAig.Nodes())-len(manager.vIncluded))...)
	}
	manager.vPis, manager.vPPis, manager.vFlops, manager.vNodes = nil, nil, nil, nil
	for id, assigned := range manager.assigned {
		if !assigned || !manager.pAig.ValidEdge(aig.Edge(id)<<1) {
			continue
		}
		object := aig.Edge(id) << 1
		switch {
		case Saig_ObjIsPi(manager.pAig, object):
			manager.vPis = append(manager.vPis, id)
		case manager.vIncluded[id] == 0:
			manager.vPPis = append(manager.vPPis, id)
		case aig.Aig_ObjIsNode(manager.pAig, object):
			manager.vNodes = append(manager.vNodes, id)
		case Saig_ObjIsLo(manager.pAig, object):
			manager.vFlops = append(manager.vFlops, id)
		}
	}
}

func Aig_Gla1DeriveAbs(manager *Aig_Gla1Man_t) *aig.Graph {
	if manager == nil || manager.pAig == nil {
		return nil
	}
	Aig_Gla1CollectAbstr(manager)
	src := manager.pAig
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	objects := src.Nodes()
	mapped := make([]aig.Edge, len(objects))
	known := make([]bool, len(objects))
	mapped[0], known[0] = dst.Const1(), true
	for _, id := range manager.vPis {
		e := dst.NewInput()
		mapped[id], known[id] = e, true
	}
	for _, id := range manager.vPPis {
		e := dst.NewInput()
		mapped[id], known[id] = e, true
	}
	for _, id := range manager.vFlops {
		e := dst.NewInput()
		mapped[id], known[id] = e, true
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		if id < 0 || id >= len(objects) || !src.ValidEdge(edge) {
			panic("saig: invalid edge in gate abstraction")
		}
		if !known[id] {
			node := objects[id]
			switch node.Kind {
			case aig.And:
				mapped[id] = dst.AND(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.ExclusiveOr:
				mapped[id] = dst.NewXor(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.Buffer, aig.Output:
				mapped[id] = copyEdge(node.Fanins[0])
			default:
				panic("saig: abstraction references an unmapped CI")
			}
			known[id] = true
		}
		result := mapped[id]
		if edge.Complemented() {
			result = result.Not()
		}
		return result
	}
	outputs := src.Outputs()
	poCount := len(outputs) - int(aig.Aig_ManRegNum(src))
	for _, output := range outputs[:poCount] {
		dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, output)))
	}
	for _, id := range manager.vFlops {
		li := Saig_ManLi(src, int32(Saig_ObjRegId(src, aig.Edge(id)<<1)))
		dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, li)))
	}
	aig.Aig_ManSetRegNum(dst, int32(len(manager.vFlops)))
	aig.Aig_ManSetConstrNum(dst, aig.Aig_ManConstrNum(src))
	return dst
}

// Aig_Gla1FetchVecId assigns one time-frame vector to an object ID.
func Aig_Gla1FetchVecId(manager *Aig_Gla1Man_t, object aig.Edge) int32 {
	if manager == nil || manager.pAig == nil || !manager.pAig.ValidEdge(object) {
		return -1
	}
	if manager.objToVec == nil {
		manager.objToVec = make([]int32, len(manager.pAig.Nodes()))
		manager.assigned = make([]bool, len(manager.pAig.Nodes()))
		manager.addedFrame = make([][]bool, len(manager.pAig.Nodes()))
	}
	id := object.NodeID()
	if manager.objToVec[id] == 0 {
		manager.objToVec[id] = int32(len(manager.vecToVar)) + 1
		manager.vecToVar = append(manager.vecToVar, make([]int32, manager.nFrames))
	}
	manager.assigned[id] = true
	return manager.objToVec[id]
}

// Aig_Gla1FetchVar returns the SAT variable for object/frame, allocating it
// and recording the reverse map when first requested.
func Aig_Gla1FetchVar(manager *Aig_Gla1Man_t, object aig.Edge, frame int32) int32 {
	if manager == nil || frame < 0 {
		return -1
	}
	for frame >= manager.nFrames {
		for i := range manager.vecToVar {
			manager.vecToVar[i] = append(manager.vecToVar[i], make([]int32, manager.nFrames)...)
		}
		manager.nFrames *= 2
	}
	vecID := Aig_Gla1FetchVecId(manager, object)
	if vecID <= 0 {
		return -1
	}
	vars := manager.vecToVar[vecID-1]
	if vars[frame] == 0 {
		vars[frame] = manager.pSat.NewVariable()
		manager.varToInfo = append(manager.varToInfo, [2]int32{int32(object.NodeID()), frame})
	}
	return vars[frame]
}

// Aig_Gla1ObjAddToSolver constrains a selected object in one frame. PIs remain
// free; initial latch values are zero, and later latches equal the prior LI.
func Aig_Gla1ObjAddToSolver(manager *Aig_Gla1Man_t, object aig.Edge, frame int32) int32 {
	if manager == nil || manager.pAig == nil || !manager.pAig.ValidEdge(object) || frame < 0 {
		return 0
	}
	if frame >= manager.nFrames {
		Aig_Gla1FetchVar(manager, object, frame)
	}
	id := object.NodeID()
	if manager.vIncluded[id] == 0 {
		return 1
	}
	if len(manager.addedFrame[id]) <= int(frame) {
		manager.addedFrame[id] = append(manager.addedFrame[id], make([]bool, int(frame)+1-len(manager.addedFrame[id]))...)
	}
	if manager.addedFrame[id][frame] {
		return 1
	}
	manager.addedFrame[id][frame] = true
	var node aig.Node
	if object.NodeID() < len(manager.pAig.Nodes()) {
		node = manager.pAig.Node(object)
	}
	variable := Aig_Gla1FetchVar(manager, object, frame)
	switch node.Kind {
	case aig.Constant:
		return Aig_Gla1AddConst(manager.pSat, variable, 0)
	case aig.Input:
		if !Saig_ObjIsLo(manager.pAig, object) {
			return 1
		}
		li := Saig_ObjLoToLi(manager.pAig, object)
		liDriver := aig.Aig_ObjChild0(manager.pAig, li)
		if frame == 0 {
			return Aig_Gla1AddConst(manager.pSat, variable, 1)
		}
		return Aig_Gla1AddBuffer(manager.pSat, variable, Aig_Gla1FetchVar(manager, liDriver, frame-1), boolInt(liDriver.Complemented()))
	case aig.And:
		return Aig_Gla1AddNode(manager.pSat, variable,
			Aig_Gla1FetchVar(manager, node.Fanins[0], frame),
			Aig_Gla1FetchVar(manager, node.Fanins[1], frame),
			boolInt(node.Fanins[0].Complemented()), boolInt(node.Fanins[1].Complemented()))
	case aig.ExclusiveOr:
		// Encode XOR from its four truth-table clauses.
		z := bsat.Positive(variable)
		a := bsat.Positive(Aig_Gla1FetchVar(manager, node.Fanins[0], frame))
		b := bsat.Positive(Aig_Gla1FetchVar(manager, node.Fanins[1], frame))
		if node.Fanins[0].Complemented() {
			a ^= 1
		}
		if node.Fanins[1].Complemented() {
			b ^= 1
		}
		if !manager.pSat.AddClause(a^1, b^1, z^1) || !manager.pSat.AddClause(a, b, z^1) || !manager.pSat.AddClause(a, b^1, z) || !manager.pSat.AddClause(a^1, b, z) {
			return 0
		}
		return 1
	default:
		return 0
	}
}

func Aig_Gla1CollectAssigned(manager *Aig_Gla1Man_t, classes *vec.Vec_Int_t) {
	if manager == nil || classes == nil {
		return
	}
	for id := 0; id < vec.Vec_IntSize(classes); id++ {
		if vec.Vec_IntEntry(classes, id) == 0 || id >= len(manager.pAig.Nodes()) {
			continue
		}
		object := aig.Edge(id) << 1
		Aig_Gla1FetchVecId(manager, object)
		node := manager.pAig.Node(object)
		for _, fanin := range node.Fanins {
			if fanin != aig.InvalidEdge {
				Aig_Gla1FetchVecId(manager, fanin)
			}
		}
	}
}

func Aig_Gla1DeriveCex(manager *Aig_Gla1Man_t, frame int32) *util.Abc_Cex_t {
	if manager == nil || manager.lastFrames == nil || manager.lastCNF == nil || manager.lastCNF.solver == nil || frame != manager.lastFrame {
		return nil
	}
	Aig_Gla1CollectAbstr(manager)
	regs, pis := int32(len(manager.vFlops)), int32(len(manager.vPis)+len(manager.vPPis))
	cex := util.Abc_CexAlloc(regs, pis, frame+1)
	cex.IFrame = frame
	read := func(id int, f int32) bool {
		if id < 0 || id >= len(manager.objToVec) {
			return false
		}
		vecID := manager.objToVec[id]
		if vecID <= 0 || int(vecID-1) >= len(manager.vecToVar) || int(f) >= len(manager.vecToVar[vecID-1]) {
			return false
		}
		variable := manager.vecToVar[vecID-1][f]
		if variable == 0 {
			return false
		}
		value, ok := manager.lastCNF.solver.Value(variable)
		return ok && value
	}
	for f := int32(0); f <= frame; f++ {
		for i, id := range manager.vPis {
			if read(id, f) {
				util.Abc_InfoSetBit(cex.PData, regs+f*pis+int32(i))
			}
		}
		for i, id := range manager.vPPis {
			if read(id, f) {
				util.Abc_InfoSetBit(cex.PData, regs+f*pis+int32(len(manager.vPis)+i))
			}
		}
	}
	return cex
}

type cbaInputFrame struct{ input, frame int32 }

// saigManCbaFilterInputsWithFrames follows one sufficient reason through the
// time-expanded graph and retains each pseudo-input's frame.
func saigManCbaFilterInputsWithFrames(graph *aig.Graph, firstFlopPI int32, cex *util.Abc_Cex_t, verbose int32) []cbaInputFrame {
	if graph == nil || cex == nil || firstFlopPI < 0 {
		return nil
	}
	inputs, outputs := graph.Inputs(), graph.Outputs()
	regs := int(aig.Aig_ManRegNum(graph))
	piCount := len(inputs) - regs
	poCount := len(outputs) - regs
	if piCount != int(cex.NPis) || firstFlopPI > cex.NPis || cex.IPo < 0 || int(cex.IPo) >= poCount {
		return nil
	}
	valuesByFrame := make([][]bool, cex.IFrame+1)
	state := make([]bool, regs)
	for i := range state {
		state[i] = util.Abc_InfoHasBit(cex.PData, int32(i))
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		inputValues := make([]bool, len(inputs))
		for i := 0; i < piCount; i++ {
			inputValues[i] = util.Abc_InfoHasBit(cex.PData, cex.NRegs+frame*cex.NPis+int32(i))
		}
		copy(inputValues[piCount:], state)
		values, result, err := evalSequentialGraphFrame(graph, inputValues)
		if err != nil {
			return nil
		}
		valuesByFrame[frame] = values
		if frame < cex.IFrame {
			copy(state, result[poCount:])
		}
	}

	seenInput := make(map[cbaInputFrame]struct{})
	inputsInReasonOrder := make([]cbaInputFrame, 0)
	seenNode := make(map[[2]int32]struct{})
	var visit func(aig.Edge, int32)
	visit = func(edge aig.Edge, frame int32) {
		edge = aig.Aig_Regular(edge)
		if !graph.ValidEdge(edge) || frame < 0 || int(frame) >= len(valuesByFrame) {
			return
		}
		key := [2]int32{int32(edge.NodeID()), frame}
		if _, exists := seenNode[key]; exists {
			return
		}
		seenNode[key] = struct{}{}
		node := graph.Node(edge)
		switch node.Kind {
		case aig.Constant:
			return
		case aig.Input:
			if Saig_ObjIsLo(graph, edge) {
				if frame == 0 {
					return
				}
				reg := Saig_ObjRegId(graph, edge)
				if reg >= 0 {
					li := Saig_ManLi(graph, reg)
					visit(aig.Aig_ObjChild0(graph, li), frame-1)
				}
				return
			}
			ci := aig.Aig_ObjCioId(graph, edge)
			if ci >= firstFlopPI {
				input := cbaInputFrame{input: ci - firstFlopPI, frame: frame}
				if _, exists := seenInput[input]; !exists {
					seenInput[input] = struct{}{}
					inputsInReasonOrder = append(inputsInReasonOrder, input)
				}
			}
		case aig.Buffer, aig.Output:
			visit(node.Fanins[0], frame)
		case aig.ExclusiveOr:
			visit(node.Fanins[0], frame)
			visit(node.Fanins[1], frame)
		case aig.And:
			v0 := valuesByFrame[frame][node.Fanins[0].NodeID()] != node.Fanins[0].Complemented()
			v1 := valuesByFrame[frame][node.Fanins[1].NodeID()] != node.Fanins[1].Complemented()
			switch {
			case v0 && v1:
				visit(node.Fanins[0], frame)
				visit(node.Fanins[1], frame)
			case !v0 && v1:
				visit(node.Fanins[0], frame)
			case v0 && !v1:
				visit(node.Fanins[1], frame)
			default:
				if reasonPriority(graph, valuesByFrame, node.Fanins[0], frame) <= reasonPriority(graph, valuesByFrame, node.Fanins[1], frame) {
					visit(node.Fanins[0], frame)
				} else {
					visit(node.Fanins[1], frame)
				}
			}
		}
	}
	po := outputs[cex.IPo]
	visit(aig.Aig_ObjChild0(graph, po), cex.IFrame)
	if verbose != 0 {
		fmt.Printf("CBA reason includes %d frame pseudo-inputs.\n", len(inputsInReasonOrder))
	}
	return inputsInReasonOrder
}

// Saig_ManCbaFilterInputs returns the distinct pseudo-primary inputs on a
// sufficient reason for the selected property output to fail.
func Saig_ManCbaFilterInputs(graph *aig.Graph, firstFlopPI int32, cex *util.Abc_Cex_t, verbose int32) *vec.Vec_Int_t {
	perFrame := saigManCbaFilterInputsWithFrames(graph, firstFlopPI, cex, 0)
	indices := make(map[int32]struct{}, len(perFrame))
	for _, input := range perFrame {
		indices[input.input] = struct{}{}
	}
	result := vec.Vec_IntAlloc(len(indices))
	ordered := make([]int32, 0, len(indices))
	for _, input := range perFrame {
		if _, exists := indices[input.input]; exists {
			delete(indices, input.input)
			ordered = append(ordered, input.input)
		}
	}
	for _, input := range ordered {
		vec.Vec_IntPush(result, input)
	}
	if verbose != 0 {
		fmt.Printf("CBA reason includes %d pseudo-primary inputs.\n", vec.Vec_IntSize(result))
	}
	return result
}

func reasonPriority(graph *aig.Graph, valuesByFrame [][]bool, edge aig.Edge, frame int32) int {
	edge = aig.Aig_Regular(edge)
	if !graph.ValidEdge(edge) || frame < 0 || int(frame) >= len(valuesByFrame) {
		return int(^uint(0) >> 1)
	}
	node := graph.Node(edge)
	if node.Kind == aig.Input {
		return int(aig.Aig_ObjCioId(graph, edge)) + int(frame)*graph.InputCount()
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		return edge.NodeID()
	}
	a, b := node.Fanins[0], node.Fanins[1]
	v0 := valuesByFrame[frame][a.NodeID()] != a.Complemented()
	v1 := valuesByFrame[frame][b.NodeID()] != b.Complemented()
	p0, p1 := reasonPriority(graph, valuesByFrame, a, frame), reasonPriority(graph, valuesByFrame, b, frame)
	if node.Kind == aig.ExclusiveOr || (v0 && v1) {
		if p0 > p1 {
			return p0
		}
		return p1
	}
	if !v0 && v1 {
		return p0
	}
	if v0 && !v1 {
		return p1
	}
	if p0 < p1 {
		return p0
	}
	return p1
}

func evalSequentialGraphFrame(graph *aig.Graph, inputs []bool) ([]bool, []bool, error) {
	if len(inputs) != len(graph.Inputs()) {
		return nil, nil, fmt.Errorf("saig: got %d frame inputs, want %d", len(inputs), len(graph.Inputs()))
	}
	nodes := graph.Nodes()
	values := make([]bool, len(nodes))
	for i, input := range graph.Inputs() {
		values[input.NodeID()] = inputs[i]
	}
	for id, node := range nodes {
		if !graph.ValidEdge(aig.Edge(id) << 1) {
			continue
		}
		switch node.Kind {
		case aig.Constant:
			values[id] = true
		case aig.Input:
		case aig.And:
			values[id] = (values[node.Fanins[0].NodeID()] != node.Fanins[0].Complemented()) && (values[node.Fanins[1].NodeID()] != node.Fanins[1].Complemented())
		case aig.ExclusiveOr:
			values[id] = (values[node.Fanins[0].NodeID()] != node.Fanins[0].Complemented()) != (values[node.Fanins[1].NodeID()] != node.Fanins[1].Complemented())
		case aig.Buffer, aig.Output:
			values[id] = values[node.Fanins[0].NodeID()] != node.Fanins[0].Complemented()
		}
	}
	outputs := make([]bool, len(graph.Outputs()))
	for i, output := range graph.Outputs() {
		outputs[i] = values[aig.Aig_ObjChild0(graph, output).NodeID()] != aig.Aig_ObjChild0(graph, output).Complemented()
	}
	return values, outputs, nil
}

func Aig_Gla1PrintAbstr(manager *Aig_Gla1Man_t, frame, iteration, variables, conflicts int32) {
	if manager == nil || manager.fVerbose == 0 {
		return
	}
	fmt.Printf("== %3d == %4d vars=%d conflicts=%d included=%d\n", frame, iteration, variables, conflicts, countIncluded(manager.vIncluded))
}

func Aig_Gla1ExtendIncluded(manager *Aig_Gla1Man_t) {
	if manager == nil || manager.pAig == nil {
		return
	}
	var include func(aig.Edge)
	seen := make(map[int]bool)
	include = func(edge aig.Edge) {
		id := edge.NodeID()
		if seen[id] || !manager.pAig.ValidEdge(edge) {
			return
		}
		seen[id] = true
		manager.vIncluded[id] = 1
		for _, fanin := range manager.pAig.Node(edge).Fanins {
			if fanin != aig.InvalidEdge {
				include(fanin)
			}
		}
	}
	for id, included := range manager.vIncluded {
		if included != 0 {
			include(aig.Edge(id) << 1)
		}
	}
}

// Aig_Gla1ManPerform checks successive frames using an incremental abstraction.
// A spurious abstract CEX is filtered to its pseudo-input support; those source
// objects are included and constrained, then the same frame is solved again.
// It returns nil for a concrete counterexample and otherwise the current gate
// bitmap, including when SAT reaches a resource limit.
func Aig_Gla1ManPerform(src *aig.Graph, oldClasses *vec.Vec_Int_t, start, maxFrames, conflictLimit, timeLimit, naiveCNF, verbose int32, piFrame *int32) *vec.Vec_Int_t {
	if src == nil || len(src.Outputs())-int(aig.Aig_ManRegNum(src)) != 1 {
		if piFrame != nil {
			*piFrame = 0
		}
		return nil
	}
	if maxFrames == 0 {
		maxFrames = 10000
	}
	manager := Aig_Gla1ManStart(src, oldClasses, naiveCNF)
	manager.nConfLimit, manager.nFramesMax, manager.fVerbose = conflictLimit, maxFrames, verbose
	deadline := time.Time{}
	if timeLimit > 0 {
		deadline = time.Now().Add(time.Duration(timeLimit) * time.Second)
		manager.pSat.SetDeadline(deadline)
	}
	Aig_Gla1FetchVecId(manager, src.Const1())
	outputs := src.Outputs()
	po := outputs[0]
	property := aig.Aig_ObjChild0(src, po)
	var result *vec.Vec_Int_t
	frameDone := int32(0)
	finishedWithCEX := false
	resourceLimited := false
	lastStatus := bsat.Unknown
	for frame := int32(0); frame < maxFrames; frame++ {
		frameDone = frame
		// Constraints for already assigned objects are added in every frame.
		// Fetching fanin variables during clause construction may assign new
		// objects, so scan the full source ID range each time.
		for id := 0; id < len(manager.assigned); id++ {
			if manager.assigned[id] && manager.vIncluded[id] != 0 {
				if Aig_Gla1ObjAddToSolver(manager, aig.Edge(id)<<1, frame) == 0 {
					if verbose != 0 {
						fmt.Printf("Error! SAT solver became UNSAT while adding object %d at frame %d.\n", id, frame)
					}
				}
			}
		}
		if frame < start {
			continue
		}
		propertyVar := Aig_Gla1FetchVar(manager, property, frame)
		assumption := bsat.Positive(propertyVar)
		if property.Complemented() {
			assumption ^= 1
		}
		Aig_Gla1CollectAbstr(manager)
		for iteration := int32(0); ; iteration++ {
			manager.pSat.SetLimits(int64(conflictLimit), 0)
			before := manager.pSat.Statistics().Conflicts
			status := manager.pSat.Solve(assumption)
			lastStatus = status
			after := manager.pSat.Statistics().Conflicts
			if status != bsat.Satisfiable {
				if verbose != 0 {
					if status == bsat.Unsatisfiable {
						fmt.Printf("== %3d == SAT solver returned UNSAT after %5d conflicts.\n", frame, after-before)
					} else {
						fmt.Printf("== %3d == SAT solver returned UNDECIDED after %5d conflicts.\n", frame, after-before)
					}
				}
				break
			}
			manager.lastFrames, manager.lastCNF, manager.lastFrame = src, &aigCNF{solver: manager.pSat}, frame
			abstract := Aig_Gla1DeriveAbs(manager)
			cex := Aig_Gla1DeriveCex(manager, frame)
			refine := Saig_ManCbaFilterInputs(abstract, int32(len(manager.vPis)), cex, 0)
			refined := false
			for i := 0; i < vec.Vec_IntSize(refine); i++ {
				ppiIndex := int(vec.Vec_IntEntry(refine, i))
				if ppiIndex < 0 || ppiIndex >= len(manager.vPPis) {
					continue
				}
				id := manager.vPPis[ppiIndex]
				if id < 0 || id >= len(manager.vIncluded) || manager.vIncluded[id] != 0 {
					continue
				}
				manager.vIncluded[id] = 1
				Aig_Gla1FetchVecId(manager, aig.Edge(id)<<1)
				refined = true
				for prior := int32(0); prior <= frame; prior++ {
					Aig_Gla1ObjAddToSolver(manager, aig.Edge(id)<<1, prior)
				}
			}
			if !refined {
				result = nil // the abstract counterexample is concrete
				finishedWithCEX = true
				break
			}
			Aig_Gla1CollectAbstr(manager)
			Aig_Gla1PrintAbstr(manager, frame, iteration, manager.pSat.NumVars(), int32(after-before))
		}
		if finishedWithCEX {
			break
		}
		if lastStatus == bsat.Unknown {
			result = includedBitmapSlice(manager.vIncluded)
			resourceLimited = true
			break
		}
	}
	if piFrame != nil {
		*piFrame = frameDone
		if frameDone == maxFrames-1 && lastStatus == bsat.Unsatisfiable {
			*piFrame = maxFrames
		}
	}
	if result == nil && !finishedWithCEX && !resourceLimited {
		result = includedBitmapSlice(manager.vIncluded)
	}
	Aig_Gla1ManStop(manager)
	return result
}

func includedBitmap(src *aig.Graph) *vec.Vec_Int_t {
	nodes := src.Nodes()
	result := vec.Vec_IntStart(len(nodes))
	for id := range nodes {
		if src.ValidEdge(aig.Edge(id) << 1) {
			vec.Vec_IntWriteEntry(result, id, 1)
		}
	}
	return result
}

func includedBitmapSlice(included []int32) *vec.Vec_Int_t {
	result := vec.Vec_IntStart(len(included))
	for id, value := range included {
		vec.Vec_IntWriteEntry(result, id, value)
	}
	return result
}

func countIncluded(included []int32) int {
	count := 0
	for _, value := range included {
		if value != 0 {
			count++
		}
	}
	return count
}

func boolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
