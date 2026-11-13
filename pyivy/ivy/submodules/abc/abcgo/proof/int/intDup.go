package int

// Graph duplication helpers ported from ABC's proof/int/intDup.c.
// Edge maps replace the original temporary pData pointers.
import "github.com/glycerine/abcgo/aig/aig"

func Inter_ManStartInitState(nRegs int32) *aig.Graph {
	if nRegs <= 0 {
		panic("interpolation requires at least one register")
	}
	g := aig.NewGraph()
	allZero := g.Const1()
	for i := int32(0); i < nRegs; i++ {
		allZero = g.AND(allZero, g.NewInput().Not())
	}
	g.AddOutput(allZero)
	return g
}

// duplicateTransitionInputs maps the source's combinational inputs into a new
// graph and optionally inserts a control input before the register outputs.
func duplicateTransitionInputs(src *aig.Graph, control bool) (*aig.Graph, map[int]aig.Edge, aig.Edge) {
	if src == nil || src.RegisterCount() <= 0 {
		panic("interpolation requires a sequential graph")
	}
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	copies := map[int]aig.Edge{0: dst.Const1()}
	ctrl := aig.InvalidEdge
	inputs := src.Inputs()
	firstReg := len(inputs) - int(src.RegisterCount())
	for i, input := range inputs {
		if control && i == firstReg {
			ctrl = dst.NewInput()
		}
		copies[input.NodeID()] = dst.NewInput()
	}
	return dst, copies, ctrl
}

// duplicateTransitionEdge copies the source cone using an operation-local map.
func duplicateTransitionEdge(src, dst *aig.Graph, copies map[int]aig.Edge, edge aig.Edge) aig.Edge {
	id := edge.NodeID()
	result, ok := copies[id]
	if !ok {
		node := src.Node(edge)
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			// ABC's interpolation cone-copy loops call Aig_And for every
			// native node, including AIG_OBJ_EXOR.
			result = dst.AND(duplicateTransitionEdge(src, dst, copies, node.Fanins[0]), duplicateTransitionEdge(src, dst, copies, node.Fanins[1]))
		case aig.Buffer, aig.Output:
			result = duplicateTransitionEdge(src, dst, copies, node.Fanins[0])
		default:
			panic("interpolation encountered an unmapped graph input")
		}
		copies[id] = result
	}
	if edge.Complemented() {
		return result.Not()
	}
	return result
}

func appendConstraintOutputs(src, dst *aig.Graph, copies map[int]aig.Edge) {
	outputs := src.Outputs()
	trueOutputs := len(outputs) - int(src.RegisterCount())
	constraints := int(aig.Aig_ManConstrNum(src))
	for _, output := range outputs[trueOutputs-constraints : trueOutputs] {
		driver := src.Node(output).Fanins[0]
		dst.AddOutput(duplicateTransitionEdge(src, dst, copies, driver).Not())
	}
}

// Inter_ManStartDuplicated retains inverted constraints and register inputs.
func Inter_ManStartDuplicated(src *aig.Graph) *aig.Graph {
	dst, copies, _ := duplicateTransitionInputs(src, false)
	appendConstraintOutputs(src, dst, copies)
	outputs := src.Outputs()
	for _, output := range outputs[len(outputs)-int(src.RegisterCount()):] {
		dst.AddOutput(duplicateTransitionEdge(src, dst, copies, src.Node(output).Fanins[0]))
	}
	aig.Aig_ManSetRegNum(dst, src.RegisterCount())
	dst.CopyTruePiCount(src)
	dst.SetTruePoCount(aig.Aig_ManConstrNum(src))
	return dst
}

// Inter_ManStartOneOutput inserts the control input used by ABC's transition
// loop. With addFirstPO set, it exposes the first ordinary output instead.
func Inter_ManStartOneOutput(src *aig.Graph, addFirstPO int32) *aig.Graph {
	dst, copies, control := duplicateTransitionInputs(src, true)
	appendConstraintOutputs(src, dst, copies)
	outputs := src.Outputs()
	if addFirstPO != 0 {
		if len(outputs) <= int(src.RegisterCount()) {
			panic("interpolation graph has no ordinary output")
		}
		dst.AddOutput(duplicateTransitionEdge(src, dst, copies, src.Node(outputs[0]).Fanins[0]))
		dst.SetTruePiCount(int32(src.InputCount()) + 1)
		dst.SetTruePoCount(addFirstPO + aig.Aig_ManConstrNum(src))
		return dst
	}
	inputs := src.Inputs()
	firstReg := len(inputs) - int(src.RegisterCount())
	firstLI := len(outputs) - int(src.RegisterCount())
	for i := 0; i < int(src.RegisterCount()); i++ {
		lo := copies[inputs[firstReg+i].NodeID()]
		li := duplicateTransitionEdge(src, dst, copies, src.Node(outputs[firstLI+i]).Fanins[0])
		dst.AddOutput(dst.Mux(control, lo, li))
	}
	aig.Aig_ManSetRegNum(dst, src.RegisterCount())
	dst.SetTruePiCount(src.TruePiCount() + 1)
	dst.SetTruePoCount(aig.Aig_ManConstrNum(src))
	return dst
}
