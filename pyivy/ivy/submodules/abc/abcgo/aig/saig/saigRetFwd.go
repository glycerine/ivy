package saig

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Aig_StaticFanout_t is an immutable adjacency list for one graph snapshot.
type Aig_StaticFanout_t struct{ nodes [][]aig.Edge }

func Aig_ManStaticFanoutStart(graph *aig.Graph) *Aig_StaticFanout_t {
	if graph == nil {
		return nil
	}
	fanouts := &Aig_StaticFanout_t{nodes: make([][]aig.Edge, len(graph.Nodes()))}
	for id, node := range graph.Nodes() {
		if !graph.ValidEdge(aig.Edge(id) << 1) {
			continue
		}
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer && node.Kind != aig.Output {
			continue
		}
		for _, fanin := range node.Fanins {
			if fanin != aig.InvalidEdge {
				fanouts.nodes[fanin.NodeID()] = append(fanouts.nodes[fanin.NodeID()], aig.Edge(id)<<1)
			}
		}
	}
	return fanouts
}

func Aig_ObjFanoutStatic(fanouts *Aig_StaticFanout_t, object aig.Edge, index int32) aig.Edge {
	if fanouts == nil || index < 0 || object.NodeID() < 0 || object.NodeID() >= len(fanouts.nodes) || int(index) >= len(fanouts.nodes[object.NodeID()]) {
		return aig.InvalidEdge
	}
	return fanouts.nodes[object.NodeID()][index]
}

func Aig_ObjSetFanoutStatic(fanouts *Aig_StaticFanout_t, object, fanout aig.Edge) {
	if fanouts == nil {
		return
	}
	fanouts.nodes[object.NodeID()] = append(fanouts.nodes[object.NodeID()], fanout)
}

func Saig_ManMarkAutonomous_rec(graph *aig.Graph, object aig.Edge, fanouts *Aig_StaticFanout_t, reached []bool) {
	if graph == nil || fanouts == nil || !graph.ValidEdge(object) {
		return
	}
	id := object.NodeID()
	if reached[id] {
		return
	}
	reached[id] = true
	for i := int32(0); ; i++ {
		fanout := Aig_ObjFanoutStatic(fanouts, object, i)
		if fanout == aig.InvalidEdge {
			break
		}
		Saig_ManMarkAutonomous_rec(graph, fanout, fanouts, reached)
	}
}

// Saig_ManMarkAutonomous marks the forward-reachable nodes from constant one
// and primary inputs, with latch-input to latch-output edges added virtually.
func Saig_ManMarkAutonomous(graph *aig.Graph) {
	if graph == nil {
		return
	}
	fanouts := Aig_ManStaticFanoutStart(graph)
	for reg := int32(0); reg < Saig_ManRegNum(graph); reg++ {
		li, lo := Saig_ManLi(graph, reg), Saig_ManLo(graph, reg)
		Aig_ObjSetFanoutStatic(fanouts, li, lo)
	}
	aig.Aig_ManIncrementTravId(graph)
	reached := make([]bool, len(graph.Nodes()))
	Saig_ManMarkAutonomous_rec(graph, graph.Const1(), fanouts, reached)
	for pi := int32(0); pi < Saig_ManPiNum(graph); pi++ {
		Saig_ManMarkAutonomous_rec(graph, aig.Aig_ManCi(graph, pi), fanouts, reached)
	}
	for id, value := range reached {
		if value {
			aig.Aig_ObjSetTravIdCurrent(graph, aig.Edge(id)<<1)
		}
	}
}

func Saig_ManRetimeForwardOne(graph *aig.Graph, fixedRegisters, movedGates *int32) *aig.Graph {
	if graph == nil {
		return nil
	}
	Saig_ManMarkAutonomous(graph)
	aig.Aig_ManIncrementTravId(graph)
	if fixedRegisters != nil {
		*fixedRegisters = 0
	}
	for reg := int32(0); reg < Saig_ManRegNum(graph); reg++ {
		lo := Saig_ManLo(graph, reg)
		if aig.Aig_ObjIsTravIdPrevious(graph, lo) {
			aig.Aig_ObjSetTravIdCurrent(graph, lo)
		} else if fixedRegisters != nil {
			*fixedRegisters++
		}
	}
	if movedGates != nil {
		*movedGates = 0
	}
	for id, node := range graph.Nodes() {
		object := aig.Edge(id) << 1
		if !graph.ValidEdge(object) || !aig.Aig_ObjIsNode(graph, object) {
			continue
		}
		if aig.Aig_ObjIsTravIdCurrent(graph, node.Fanins[0]) && aig.Aig_ObjIsTravIdCurrent(graph, node.Fanins[1]) {
			aig.Aig_ObjSetTravIdCurrent(graph, object)
			if movedGates != nil {
				*movedGates++
			}
		}
	}
	for reg := int32(0); reg < Saig_ManRegNum(graph); reg++ {
		aig.Aig_ObjSetTravIdCurrent(graph, Saig_ManLo(graph, reg))
	}
	aig.Aig_ManIncrementTravId(graph)
	cut := vec.Vec_PtrAlloc(32)
	seen := make(map[int]bool)
	for id, node := range graph.Nodes() {
		object := aig.Edge(id) << 1
		if !graph.ValidEdge(object) || aig.Aig_ObjIsTravIdPrevious(graph, object) {
			continue
		}
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer && node.Kind != aig.Output {
			continue
		}
		for _, fanin := range node.Fanins {
			if fanin == aig.InvalidEdge || !aig.Aig_ObjIsTravIdPrevious(graph, fanin) {
				continue
			}
			if !seen[fanin.NodeID()] {
				seen[fanin.NodeID()] = true
				vec.Vec_PtrPush(cut, aig.Aig_Regular(fanin))
			}
		}
	}
	return Saig_ManRetimeDupForward(graph, cut)
}

// Saig_ManRetimeDupForward duplicates the cones above and below a retiming cut.
// Cut objects become latch outputs in the new graph; their next-state functions
// are emitted as the final latch-input outputs.
func Saig_ManRetimeDupForward(graph *aig.Graph, cut *vec.Vec_Ptr_t) *aig.Graph {
	if graph == nil || cut == nil {
		return nil
	}
	dst := aig.NewGraph()
	dst.Name, dst.Spec = graph.Name, graph.Spec
	nodes := graph.Nodes()
	mapped := make([]aig.Edge, len(nodes))
	known := make([]bool, len(nodes))
	mapped[0], known[0] = dst.Const1(), true
	for i := 0; i < int(Saig_ManPiNum(graph)); i++ {
		input := aig.Aig_ManCi(graph, int32(i))
		mapped[input.NodeID()], known[input.NodeID()] = dst.NewInput(), true
	}
	for i := 0; i < vec.Vec_PtrSize(cut); i++ {
		object := vec.Vec_PtrEntry(cut, i).(aig.Edge)
		id := object.NodeID()
		register := dst.NewInput()
		if aig.Aig_ObjPhase(graph, object) {
			register = register.Not()
		}
		mapped[id], known[id] = register, true
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		if id < 0 || id >= len(nodes) || !graph.ValidEdge(edge) {
			panic("saig: invalid edge while duplicating retiming cut")
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
				panic("saig: retiming cut omits a required latch output")
			default:
				panic("saig: unsupported node while duplicating retiming cut")
			}
			known[id] = true
		}
		result := mapped[id]
		if edge.Complemented() {
			result = result.Not()
		}
		return result
	}
	outputs := graph.Outputs()
	regCount := int(aig.Aig_ManRegNum(graph))
	poCount := int(Saig_ManPoNum(graph))
	for _, output := range outputs[:poCount] {
		dst.AddOutput(copyEdge(aig.Aig_ObjChild0(graph, output)))
	}
	nextState := make([]aig.Edge, regCount)
	for reg := 0; reg < regCount; reg++ {
		li := Saig_ManLi(graph, int32(reg))
		nextState[reg] = copyEdge(aig.Aig_ObjChild0(graph, li))
	}
	for reg := 0; reg < regCount; reg++ {
		lo := Saig_ManLo(graph, int32(reg))
		mapped[lo.NodeID()], known[lo.NodeID()] = nextState[reg], true
	}
	for i := 0; i < vec.Vec_PtrSize(cut); i++ {
		object := vec.Vec_PtrEntry(cut, i).(aig.Edge)
		if aig.Aig_ObjIsNode(graph, object) {
			known[object.NodeID()] = false
		}
	}
	for i := 0; i < vec.Vec_PtrSize(cut); i++ {
		object := vec.Vec_PtrEntry(cut, i).(aig.Edge)
		stateValue := copyEdge(object)
		if aig.Aig_ObjPhase(graph, object) {
			stateValue = stateValue.Not()
		}
		dst.AddOutput(stateValue)
	}
	aig.Aig_ManSetRegNumForConstruction(dst, int32(vec.Vec_PtrSize(cut)))
	dst.CopyTruePiCount(graph)
	dst.CopyTruePoCount(graph)
	aig.Aig_ManCleanup(dst)
	return dst
}

func Saig_ManRetimeForward(graph *aig.Graph, maxIterations, verbose int32) *aig.Graph {
	if graph == nil {
		return nil
	}
	result := graph
	moves := int32(1)
	for iteration := int32(0); iteration < maxIterations && moves > 0; iteration++ {
		started := time.Now()
		previous := result
		fixed := int32(0)
		result = Saig_ManRetimeForwardOne(previous, &fixed, &moves)
		if result == nil {
			return nil
		}
		if verbose != 0 {
			fmt.Printf("%2d : And = %6d. Reg = %5d. Unret = %5d. Move = %6d. elapsed %s\n", iteration+1, aig.Aig_ManNodeNum(previous), aig.Aig_ManRegNum(previous), fixed, moves, time.Since(started).Round(time.Millisecond))
		}
	}
	return aig.Aig_ManReduceLaches(result, verbose)
}
