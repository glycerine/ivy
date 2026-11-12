package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Saig_RefMan_t holds a CEX-guided unrolling and the mapping from its symbolic
// frame inputs back to the original primary inputs and frames.
type Saig_RefMan_t struct {
	pAig      *aig.Graph
	pCex      *util.Abc_Cex_t
	nInputs   int32
	fVerbose  int32
	pFrames   *aig.Graph
	vMapPiF2A *vec.Vec_Int_t
}

// Saig_ManUnrollWithCex creates the CEX-frame unrolling of one selected output.
// The first nInputs primary inputs are fixed from the CEX; later inputs are
// symbolic and recorded as (original PI, frame) pairs.
func Saig_ManUnrollWithCex(graph *aig.Graph, cex *util.Abc_Cex_t, nInputs int32, mapOut **vec.Vec_Int_t) *aig.Graph {
	if graph == nil || cex == nil || cex.IFrame < 0 || nInputs < 0 || nInputs > cex.NPis {
		return nil
	}
	inputs, outputs := graph.Inputs(), graph.Outputs()
	regCount := int(aig.Aig_ManRegNum(graph))
	piCount, poCount := len(inputs)-regCount, len(outputs)-regCount
	if piCount != int(cex.NPis) || regCount != int(cex.NRegs) || cex.IPo < 0 || int(cex.IPo) >= poCount {
		return nil
	}
	dst := aig.NewGraph()
	dst.Name, dst.Spec = graph.Name, graph.Spec
	mapping := vec.Vec_IntAlloc(int(cex.IFrame+1) * max(0, piCount-int(nInputs)) * 2)
	state := make([]aig.Edge, regCount)
	for i := range state {
		if util.Abc_InfoHasBit(cex.PData, int32(i)) {
			state[i] = dst.Const1()
		} else {
			state[i] = dst.Const0()
		}
	}
	nodes := graph.Nodes()
	var property aig.Edge
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		mapped := make([]aig.Edge, len(nodes))
		known := make([]bool, len(nodes))
		mapped[0], known[0] = dst.Const1(), true
		for ci, input := range inputs {
			id := input.NodeID()
			if ci >= piCount {
				mapped[id], known[id] = state[ci-piCount], true
				continue
			}
			if int32(ci) < nInputs {
				bit := cex.NRegs + frame*cex.NPis + int32(ci)
				if util.Abc_InfoHasBit(cex.PData, bit) {
					mapped[id] = dst.Const1()
				} else {
					mapped[id] = dst.Const0()
				}
				known[id] = true
				continue
			}
			mapped[id], known[id] = dst.NewInput(), true
			vec.Vec_IntPush(mapping, int32(ci))
			vec.Vec_IntPush(mapping, frame)
		}
		var copyEdge func(aig.Edge) aig.Edge
		copyEdge = func(edge aig.Edge) aig.Edge {
			id := edge.NodeID()
			if id < 0 || id >= len(nodes) || !graph.ValidEdge(edge) {
				panic("saig: invalid edge in CEX unroll")
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
					panic("saig: source input missing in CEX unroll")
				default:
					panic("saig: unsupported node in CEX unroll")
				}
				known[id] = true
			}
			result := mapped[id]
			if edge.Complemented() {
				result = result.Not()
			}
			return result
		}
		for id, node := range nodes {
			if !graph.ValidEdge(aig.Edge(id) << 1) {
				continue
			}
			if node.Kind == aig.And || node.Kind == aig.ExclusiveOr || node.Kind == aig.Buffer {
				copyEdge(aig.Edge(id) << 1)
			}
		}
		if frame == cex.IFrame {
			property = copyEdge(aig.Aig_ObjChild0(graph, outputs[cex.IPo])).Not()
		}
		if frame < cex.IFrame {
			for reg := range state {
				li := aig.Aig_ManLi(graph, int32(reg))
				state[reg] = copyEdge(aig.Aig_ObjChild0(graph, li))
			}
		}
	}
	dst.AddOutput(property)
	if mapOut != nil {
		*mapOut = mapping
	}
	return dst
}

func Saig_RefManStart(graph *aig.Graph, cex *util.Abc_Cex_t, nInputs, verbose int32) *Saig_RefMan_t {
	manager := &Saig_RefMan_t{pAig: graph, pCex: cex, nInputs: nInputs, fVerbose: verbose}
	manager.pFrames = Saig_ManUnrollWithCex(graph, cex, nInputs, &manager.vMapPiF2A)
	return manager
}

func Saig_RefManStop(manager *Saig_RefMan_t) {
	if manager == nil {
		return
	}
	manager.pFrames = nil
	manager.vMapPiF2A = nil
	manager.pAig = nil
	manager.pCex = nil
}

// Saig_RefManFindReason returns CIs of the CEX-frame unrolling that form a
// sufficient structural reason for the selected failing output.
func Saig_RefManFindReason(manager *Saig_RefMan_t) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(0)
	if manager == nil || manager.pAig == nil || manager.pCex == nil {
		return result
	}
	frames := saigManCbaFilterInputsWithFrames(manager.pAig, manager.nInputs, manager.pCex, 0)
	for _, reason := range frames {
		original := manager.nInputs + reason.input
		for pi := 0; pi+1 < vec.Vec_IntSize(manager.vMapPiF2A); pi += 2 {
			if vec.Vec_IntEntry(manager.vMapPiF2A, pi) == original && vec.Vec_IntEntry(manager.vMapPiF2A, pi+1) == reason.frame {
				vec.Vec_IntPush(result, int32(pi/2))
				break
			}
		}
	}
	return result
}

func Saig_RefManReason2Inputs(manager *Saig_RefMan_t, reasons *vec.Vec_Int_t) *vec.Vec_Int_t {
	if manager == nil || manager.vMapPiF2A == nil || reasons == nil {
		return vec.Vec_IntAlloc(0)
	}
	result := vec.Vec_IntAlloc(vec.Vec_IntSize(reasons))
	seen := make(map[int32]bool)
	for i := 0; i < vec.Vec_IntSize(reasons); i++ {
		index := int(vec.Vec_IntEntry(reasons, i)) * 2
		if index < 0 || index+1 >= vec.Vec_IntSize(manager.vMapPiF2A) {
			continue
		}
		input := vec.Vec_IntEntry(manager.vMapPiF2A, index)
		if !seen[input] {
			seen[input] = true
			vec.Vec_IntPush(result, input)
		}
	}
	return result
}

func Saig_RefManReason2Cex(manager *Saig_RefMan_t, reasons *vec.Vec_Int_t) *util.Abc_Cex_t {
	if manager == nil || manager.pCex == nil {
		return nil
	}
	care := util.Abc_CexDup(manager.pCex, manager.pCex.NRegs)
	for i := range care.PData {
		care.PData[i] = 0
	}
	for i := 0; i < vec.Vec_IntSize(reasons); i++ {
		index := int(vec.Vec_IntEntry(reasons, i)) * 2
		if index < 0 || index+1 >= vec.Vec_IntSize(manager.vMapPiF2A) {
			continue
		}
		input := vec.Vec_IntEntry(manager.vMapPiF2A, index)
		frame := vec.Vec_IntEntry(manager.vMapPiF2A, index+1)
		util.Abc_InfoSetBit(care.PData, care.NRegs+frame*care.NPis+input)
	}
	return care
}

// Saig_RefManSetPhases evaluates the selected property in the frame graph,
// optionally replacing uncared frame-input values with value1.
func Saig_RefManSetPhases(manager *Saig_RefMan_t, care *util.Abc_Cex_t, value1 int32) int32 {
	if manager == nil || manager.pFrames == nil || manager.pCex == nil {
		return 0
	}
	inputs := make([]bool, manager.pFrames.InputCount())
	for i := range inputs {
		mapIndex := int32(2 * i)
		original := vec.Vec_IntEntry(manager.vMapPiF2A, int(mapIndex))
		frame := vec.Vec_IntEntry(manager.vMapPiF2A, int(mapIndex+1))
		bit := manager.pCex.NRegs + frame*manager.pCex.NPis + original
		inputs[i] = util.Abc_InfoHasBit(manager.pCex.PData, bit)
		if care != nil && !util.Abc_InfoHasBit(care.PData, bit) {
			inputs[i] = value1 != 0
		}
	}
	_, outputs, err := evalSequentialGraphFrame(manager.pFrames, inputs)
	if err != nil || len(outputs) == 0 {
		return 0
	}
	return boolInt(outputs[0])
}

// Saig_ManFindCexCareBits keeps only the CEX-frame input bits that occur on a
// sufficient reason for the selected property failure. fNewOrder is retained
// for source compatibility; the source's SAT minimization call is commented out.
func Saig_ManFindCexCareBits(graph *aig.Graph, cex *util.Abc_Cex_t, nInputs, fNewOrder, verbose int32) *util.Abc_Cex_t {
	_ = fNewOrder
	manager := Saig_RefManStart(graph, cex, nInputs, verbose)
	if manager.pFrames == nil {
		return nil
	}
	reasons := Saig_RefManFindReason(manager)
	care := Saig_RefManReason2Cex(manager, reasons)
	if verbose != 0 {
		fmt.Printf("CEX care set retains %d frame inputs.\n", vec.Vec_IntSize(reasons))
	}
	Saig_RefManStop(manager)
	return care
}

// Saig_ManExtendCounterExampleTest3 returns original PI IDs that the CEX reason
// needs, excluding the first iFirstFlopPi concrete inputs.
func Saig_ManExtendCounterExampleTest3(graph *aig.Graph, firstFlopPI int32, cex *util.Abc_Cex_t, verbose int32) *vec.Vec_Int_t {
	manager := Saig_RefManStart(graph, cex, firstFlopPI, verbose)
	if manager.pFrames == nil {
		return nil
	}
	reasons := Saig_RefManFindReason(manager)
	inputs := Saig_RefManReason2Inputs(manager, reasons)
	if verbose != 0 {
		fmt.Printf("CEX reason contains %d frame inputs and %d original PIs.\n", vec.Vec_IntSize(reasons), vec.Vec_IntSize(inputs))
	}
	Saig_RefManStop(manager)
	return inputs
}
