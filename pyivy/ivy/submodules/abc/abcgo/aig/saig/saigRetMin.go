package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

// Saig_ManRetimeInitState finds a satisfying assignment to all outputs of an
// initial-state network. The returned vector contains the network's CI values.
func Saig_ManRetimeInitState(graph *aig.Graph) *vec.Vec_Int_t {
	if graph == nil {
		return nil
	}
	cnf, err := encodeAIG(graph)
	if err != nil {
		return nil
	}
	assumptions := make([]bsat.Literal, len(graph.Outputs()))
	for i, output := range graph.Outputs() {
		assumptions[i] = cnf.lit(aig.Aig_ObjChild0(graph, output)) ^ 1
	}
	cnf.solver.SetLimits(1_000_000, 0)
	if cnf.solver.Solve(assumptions...) != bsat.Satisfiable {
		return nil
	}
	model := cnf.solver.Model()
	values := vec.Vec_IntAlloc(len(graph.Inputs()))
	for _, input := range graph.Inputs() {
		value := int32(0)
		if model[cnf.vars[input.NodeID()]] {
			value = 1
		}
		vec.Vec_IntPush(values, value)
	}
	return values
}

// Saig_ManRetimeUnsatCore returns the first output represented in the
// assumption core. It returns -1 when the outputs are jointly satisfiable.
func Saig_ManRetimeUnsatCore(graph *aig.Graph, verbose int32) int32 {
	if graph == nil {
		return -1
	}
	cnf, err := encodeAIG(graph)
	if err != nil {
		return -1
	}
	outputs := graph.Outputs()
	assumptions := make([]bsat.Literal, len(outputs))
	for i, output := range outputs {
		assumptions[i] = cnf.lit(aig.Aig_ObjChild0(graph, output)) ^ 1
	}
	cnf.solver.SetLimits(1_000_000, 0)
	if cnf.solver.Solve(assumptions...) != bsat.Unsatisfiable {
		return -1
	}
	core := cnf.solver.UnsatCore()
	for i, assumption := range assumptions {
		for _, member := range core {
			if member == assumption^1 {
				return int32(i)
			}
		}
	}
	return -1
}

func Saig_ManMarkCone_rec(graph *aig.Graph, object aig.Edge) {
	if graph == nil || !graph.ValidEdge(object) || aig.Aig_ObjIsTravIdCurrent(graph, object) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(graph, object)
	node := graph.Node(object)
	if node.Kind == aig.And || node.Kind == aig.ExclusiveOr || node.Kind == aig.Buffer || node.Kind == aig.Output {
		for _, fanin := range node.Fanins {
			if fanin != aig.InvalidEdge {
				Saig_ManMarkCone_rec(graph, fanin)
			}
		}
	}
}

// Saig_ManRetimeCountCut counts distinct edges crossing from the marked cone
// rooted at cut into its complement.
func Saig_ManRetimeCountCut(graph *aig.Graph, cut *vec.Vec_Ptr_t) int32 {
	if graph == nil || cut == nil {
		return 0
	}
	aig.Aig_ManIncrementTravId(graph)
	for _, item := range vec.Vec_PtrArray(cut) {
		Saig_ManMarkCone_rec(graph, item.(aig.Edge))
	}
	seen := make(map[aig.Edge]bool)
	for id, node := range graph.Nodes() {
		object := aig.Edge(id) << 1
		if !graph.ValidEdge(object) || aig.Aig_ObjIsTravIdCurrent(graph, object) {
			continue
		}
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer && node.Kind != aig.Output {
			continue
		}
		for _, fanin := range node.Fanins {
			if fanin != aig.InvalidEdge && aig.Aig_ObjIsTravIdCurrent(graph, fanin) {
				seen[aig.Aig_Regular(fanin)] = true
			}
		}
	}
	return int32(len(seen))
}

// Saig_ManRetimeDup_rec remains as a source-name helper. Native duplication is
// performed by the retiming routines using an explicit per-call edge map.
func Saig_ManRetimeDup_rec(_ *aig.Graph, _ aig.Edge) {}

// Saig_ManRetimeDupBackward moves registers from latch-input functions to the
// specified cut. Cut entries are regular AIG edges.
func Saig_ManRetimeDupBackward(src *aig.Graph, cut *vec.Vec_Ptr_t, initial *vec.Vec_Int_t) *aig.Graph {
	if src == nil || cut == nil {
		return nil
	}
	inputs, outputs, nodes := src.Inputs(), src.Outputs(), src.Nodes()
	regs, piCount := int(aig.Aig_ManRegNum(src)), int(Saig_ManPiNum(src))
	poCount := int(Saig_ManPoNum(src))
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	mapped := make([]aig.Edge, len(nodes))
	known := make([]bool, len(nodes))
	mapped[0], known[0] = dst.Const1(), true
	for i := 0; i < piCount; i++ {
		e := inputs[i]
		mapped[e.NodeID()], known[e.NodeID()] = dst.NewInput(), true
	}
	cutEdges := make([]aig.Edge, vec.Vec_PtrSize(cut))
	for i := range cutEdges {
		e, ok := vec.Vec_PtrEntry(cut, i).(aig.Edge)
		if !ok {
			return nil
		}
		e = aig.Aig_Regular(e)
		cutEdges[i] = e
		register := dst.NewInput()
		if initial != nil && i < vec.Vec_IntSize(initial) && vec.Vec_IntEntry(initial, i) != 0 {
			register = register.Not()
		}
		mapped[e.NodeID()], known[e.NodeID()] = register, true
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		if id < 0 || id >= len(nodes) || !src.ValidEdge(edge) {
			panic("saig: invalid edge in backward retiming")
		}
		if !known[id] {
			node := nodes[id]
			switch node.Kind {
			case aig.And:
				mapped[id] = dst.AND(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.ExclusiveOr:
				mapped[id] = dst.NewXor(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1]))
			case aig.Buffer, aig.Output:
				mapped[id] = copyEdge(node.Fanins[0])
			case aig.Input:
				panic("saig: cut omits a required input")
			default:
				panic("saig: unsupported node in backward retiming")
			}
			known[id] = true
		}
		result := mapped[id]
		if edge.Complemented() {
			result = result.Not()
		}
		return result
	}
	// Capture old latch next-state functions before cutting them out of the map.
	next := make([]aig.Edge, regs)
	for i := range next {
		next[i] = copyEdge(aig.Aig_ObjChild0(src, outputs[poCount+i]))
	}
	for i := range cutEdges {
		known[cutEdges[i].NodeID()] = false
	}
	mapped[0], known[0] = dst.Const1(), true
	for i := 0; i < piCount; i++ {
		e := inputs[i]
		mapped[e.NodeID()], known[e.NodeID()] = dst.Inputs()[i], true
	}
	for i := range next {
		mapped[inputs[piCount+i].NodeID()], known[inputs[piCount+i].NodeID()] = next[i], true
	}
	for _, edge := range outputs[:poCount] {
		dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, edge)))
	}
	for i, edge := range cutEdges {
		phase := false
		if initial != nil && i < vec.Vec_IntSize(initial) {
			phase = vec.Vec_IntEntry(initial, i) != 0
		}
		dst.AddOutput(edgeNotIf(copyEdge(edge), phase))
	}
	aig.Aig_ManSetRegNumForConstruction(dst, int32(len(cutEdges)))
	dst.CopyTruePiCount(src)
	dst.CopyTruePoCount(src)
	return dst
}

// Saig_ManRetimeDupInitState emits each old latch-input function as an output.
func Saig_ManRetimeDupInitState(src *aig.Graph, cut *vec.Vec_Ptr_t) *aig.Graph {
	if src == nil || cut == nil {
		return nil
	}
	inputs, outputs, nodes := src.Inputs(), src.Outputs(), src.Nodes()
	regs := int(aig.Aig_ManRegNum(src))
	poCount := len(outputs) - regs
	dst := aig.NewGraph()
	mapped := make([]aig.Edge, len(nodes))
	known := make([]bool, len(nodes))
	mapped[0], known[0] = dst.Const1(), true
	for _, item := range vec.Vec_PtrArray(cut) {
		edge := item.(aig.Edge)
		mapped[edge.NodeID()], known[edge.NodeID()] = dst.NewInput(), true
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		if !known[id] {
			n := nodes[id]
			switch n.Kind {
			case aig.And:
				mapped[id] = dst.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.ExclusiveOr:
				mapped[id] = dst.NewXor(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.Buffer, aig.Output:
				mapped[id] = copyEdge(n.Fanins[0])
			default:
				panic("saig: init-state cut omits input")
			}
			known[id] = true
		}
		result := mapped[id]
		if edge.Complemented() {
			result = result.Not()
		}
		return result
	}
	for i := 0; i < regs; i++ {
		dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, outputs[poCount+i])))
	}
	_ = inputs
	return dst
}

func Saig_ManGetRegistersToExclude(graph *aig.Graph) *vec.Vec_Ptr_t {
	bad := vec.Vec_PtrAlloc(4)
	if graph == nil {
		return bad
	}
	positive, negative := make(map[int]bool), make(map[int]bool)
	for reg := int32(0); reg < Saig_ManRegNum(graph); reg++ {
		driver := aig.Aig_ObjChild0(graph, Saig_ManLi(graph, reg))
		if driver.Complemented() {
			negative[driver.NodeID()] = true
		} else {
			positive[driver.NodeID()] = true
		}
	}
	for reg := int32(0); reg < Saig_ManRegNum(graph); reg++ {
		driver := aig.Aig_ObjChild0(graph, Saig_ManLi(graph, reg))
		if positive[driver.NodeID()] && negative[driver.NodeID()] {
			vec.Vec_PtrPush(bad, Saig_ManLi(graph, reg))
		}
	}
	return bad
}

// Saig_ManHideBadRegs moves selected latch interfaces into the PI/PO region.
func Saig_ManHideBadRegs(graph *aig.Graph, bad *vec.Vec_Ptr_t) int32 {
	if graph == nil || bad == nil || vec.Vec_PtrSize(bad) == 0 {
		return 0
	}
	inputs, outputs := graph.Inputs(), graph.Outputs()
	regs := int(aig.Aig_ManRegNum(graph))
	piCount, poCount := len(inputs)-regs, len(outputs)-regs
	if int(graph.TruePiCount()) != piCount || int(graph.TruePoCount()) != poCount {
		panic("saig: stored terminal counts do not match interface")
	}
	selected := make(map[int]bool)
	var selectedOrder []int
	for _, item := range vec.Vec_PtrArray(bad) {
		edge, ok := item.(aig.Edge)
		if !ok {
			panic("saig: bad register must be a latch input")
		}
		reg := -1
		for i, li := range outputs[poCount:] {
			if edge == li {
				reg = i
				break
			}
		}
		if reg < 0 || selected[reg] {
			panic("saig: invalid or duplicate bad latch input")
		}
		selected[reg] = true
		selectedOrder = append(selectedOrder, reg)
	}
	newInputs := append([]aig.Edge(nil), inputs[:piCount]...)
	newOutputs := append([]aig.Edge(nil), outputs[:poCount]...)
	for _, reg := range selectedOrder {
		newInputs = append(newInputs, inputs[piCount+reg])
		newOutputs = append(newOutputs, outputs[poCount+reg])
	}
	for reg := 0; reg < regs; reg++ {
		if !selected[reg] {
			newInputs = append(newInputs, inputs[piCount+reg])
			newOutputs = append(newOutputs, outputs[poCount+reg])
		}
	}
	count := int32(len(selected))
	if count == 0 {
		return 0
	}
	if err := aig.Aig_ManReorderCisCos(graph, newInputs, newOutputs, int32(regs)-count); err != nil {
		panic(err)
	}
	graph.SetTruePiCount(graph.TruePiCount() + count)
	graph.SetTruePoCount(graph.TruePoCount() + count)
	return count
}

func Saig_ManExposeBadRegs(graph *aig.Graph, count int32) {
	if graph == nil {
		panic("saig: nil graph")
	}
	// ABC exposes registers by changing the three stored counts only.
	truePis, truePos := graph.TruePiCount(), graph.TruePoCount()
	aig.Aig_ManSetRegNumForConstruction(graph, graph.RegisterCount()+count)
	graph.SetTruePiCount(truePis - count)
	graph.SetTruePoCount(truePos - count)
}

// Saig_ManRetimeMinAreaBackward derives the source minimum cut and retries
// after excluding registers whose initial constraints have no solution.
func Saig_ManRetimeMinAreaBackward(graph *aig.Graph, verbose int32) *aig.Graph {
	if graph != nil && aig.Aig_ManConstrNum(graph) != 0 &&
		graph.Assertions > Saig_ManPoNum(graph)-aig.Aig_ManConstrNum(graph) {
		// The intermediate hidden-output interface cannot express overlapping
		// assertion and constraint ranges. Return the restored interface.
		return saigRetimeMinAreaBackward(graph, verbose, true)
	}
	return saigRetimeMinAreaBackward(graph, verbose, false)
}

// restoreHidden keeps the intermediate hidden-output interface private and
// restores those registers before returning to the top-level retiming pass.
func saigRetimeMinAreaBackward(graph *aig.Graph, verbose int32, restoreHidden bool) *aig.Graph {
	if graph == nil || aig.Aig_ManRegNum(graph) == 0 {
		return nil
	}
	bad := Saig_ManGetRegistersToExclude(graph)
	originalPOs := int(Saig_ManPoNum(graph))
	constraints := int(aig.Aig_ManConstrNum(graph))
	// Assertions denote a prefix, so overlapping asserted constraints cannot
	// be moved around hidden POs without changing that representation.
	canHide := constraints == 0 || int(graph.Assertions) <= originalPOs-constraints
	if !restoreHidden && !canHide && vec.Vec_PtrSize(bad) != 0 {
		return nil
	}
	if verbose != 0 && vec.Vec_PtrSize(bad) > 0 {
		fmt.Printf("Excluding %d registers that cannot be backward retimed.\n", vec.Vec_PtrSize(bad))
	}
	for {
		Saig_ManHideBadRegs(graph, bad)
		edges := saigRetimeMinCut(graph, false)
		if len(edges) >= int(graph.RegisterCount()) {
			return nil
		}
		cut := vec.Vec_PtrAlloc(len(edges))
		for _, edge := range edges {
			vec.Vec_PtrPush(cut, edge)
		}
		initialGraph := Saig_ManRetimeDupInitState(graph, cut)
		initial := Saig_ManRetimeInitState(initialGraph)
		if initial != nil {
			result := Saig_ManRetimeDupBackward(graph, cut, initial)
			if result != nil && restoreHidden {
				hidden := Saig_ManPoNum(result) - int32(originalPOs)
				if hidden < 0 {
					panic("saig: invalid hidden output count")
				}
				aig.Aig_ManSetRegNum(result, result.RegisterCount()+hidden)
				return result
			}
			if result != nil && constraints != 0 {
				outputs := result.Outputs()
				pos := int(Saig_ManPoNum(result))
				hidden := pos - originalPOs
				if hidden < 0 {
					panic("saig: invalid hidden output count")
				}
				if hidden > 0 {
					firstConstraint := originalPOs - constraints
					reordered := append([]aig.Edge(nil), outputs[:firstConstraint]...)
					reordered = append(reordered, outputs[originalPOs:pos]...)
					reordered = append(reordered, outputs[firstConstraint:originalPOs]...)
					reordered = append(reordered, outputs[pos:]...)
					if err := aig.Aig_ManReorderCisCos(result, result.Inputs(), reordered, result.RegisterCount()); err != nil {
						panic(err)
					}
				}
			}
			return result
		}
		badReg := Saig_ManRetimeUnsatCore(initialGraph, verbose)
		if badReg < 0 || badReg >= graph.RegisterCount() {
			return nil
		}
		if !restoreHidden && !canHide {
			return nil
		}
		if verbose != 0 {
			fmt.Printf("Excluding register %d.\n", badReg)
		}
		bad = vec.Vec_PtrAlloc(1)
		vec.Vec_PtrPush(bad, Saig_ManLi(graph, badReg))
	}
}

// Saig_ManRetimeMinArea follows ABC's minimum-cut forward and backward
// retiming passes, preserving zero initial state when requested.
func Saig_ManRetimeMinArea(graph *aig.Graph, maxIterations, forwardOnly, backwardOnly, initialState, verbose int32) *aig.Graph {
	if graph == nil {
		return nil
	}
	current := aig.Aig_ManDupSimple(graph)
	if backwardOnly == 0 {
		for i := int32(0); i < maxIterations && current.RegisterCount() > 0; i++ {
			edges := saigRetimeMinCut(current, true)
			if len(edges) >= int(current.RegisterCount()) {
				break
			}
			cut := vec.Vec_PtrAlloc(len(edges))
			for _, edge := range edges {
				vec.Vec_PtrPush(cut, edge)
			}
			current = Saig_ManRetimeDupForward(current, cut)
			if current == nil {
				panic("saig: forward retiming failed")
			}
			if verbose != 0 {
				aig.Aig_ManReportImprovement(graph, current)
			}
		}
	}
	changed := false
	if forwardOnly == 0 {
		for i := int32(0); i < maxIterations && current.RegisterCount() > 0; i++ {
			var next *aig.Graph
			if initialState == 0 {
				edges := saigRetimeMinCut(current, false)
				if len(edges) >= int(current.RegisterCount()) {
					break
				}
				cut := vec.Vec_PtrAlloc(len(edges))
				for _, edge := range edges {
					vec.Vec_PtrPush(cut, edge)
				}
				next = Saig_ManRetimeDupBackward(current, cut, nil)
			} else {
				next = saigRetimeMinAreaBackward(aig.Aig_ManDupSimple(current), verbose, true)
			}
			if next == nil {
				break
			}
			current = next
			changed = true
			if verbose != 0 {
				aig.Aig_ManReportImprovement(graph, current)
			}
		}
	}
	if forwardOnly == 0 && initialState == 0 && changed {
		fmt.Println("Assuming const-0 init-state after backward retiming. Result will not verify.")
	}
	return current
}

func edgeNotIf(edge aig.Edge, yes bool) aig.Edge {
	if yes {
		return edge.Not()
	}
	return edge
}
