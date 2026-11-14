package llb

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func llbIsLi(graph *aig.Graph, edge aig.Edge) bool {
	start := int(aig.Aig_ManCoNum(graph) - aig.Aig_ManRegNum(graph))
	for i := start; i < int(aig.Aig_ManCoNum(graph)); i++ {
		if aig.Aig_ManCo(graph, int32(i)) == edge {
			return true
		}
	}
	return false
}

func llbIsPo(graph *aig.Graph, edge aig.Edge) bool {
	for i := int32(0); i < aig.Aig_ManCoNum(graph)-aig.Aig_ManRegNum(graph); i++ {
		if aig.Aig_ManCo(graph, i) == edge {
			return true
		}
	}
	return false
}

func llbFanouts(graph *aig.Graph, edge aig.Edge) []aig.Edge {
	var result []aig.Edge
	for i, fan := int32(0), aig.Aig_ObjFanout0Int(graph, int32(edge.NodeID())); i < aig.Aig_ObjRefs(graph, edge); i, fan = i+1, aig.Aig_ObjFanoutNext(graph, fan) {
		result = append(result, aig.Aig_ManObj(graph, fan>>1))
	}
	return result
}

func llbTracePathRec(graph *aig.Graph, object, pivot aig.Edge, failed, successful map[int]bool) bool {
	id := object.NodeID()
	if failed[id] {
		return false
	}
	if successful[id] {
		return true
	}
	if llbIsLi(graph, object) || llbIsPo(graph, object) {
		return false
	}
	if object.NodeID() == pivot.NodeID() {
		return true
	}
	if !aig.Aig_ObjIsCand(graph, object) {
		panic("llb: non-candidate object in pivot fanout traversal")
	}
	for _, fanout := range llbFanouts(graph, object) {
		if !llbTracePathRec(graph, fanout, pivot, failed, successful) {
			failed[id] = true
			return false
		}
	}
	successful[id] = true
	return true
}

func Llb_ManTracePaths_rec(graph *Aig_Man_t, object, pivot aig.Edge) int32 {
	if llbTracePathRec(graph, object, pivot, map[int]bool{}, map[int]bool{}) {
		return 1
	}
	return 0
}

func Llb_ManTracePaths(graph *Aig_Man_t, pivot aig.Edge) int32 {
	var count int32
	inputs := graph.Inputs()
	start := len(inputs) - int(aig.Aig_ManRegNum(graph))
	failed, successful := map[int]bool{}, map[int]bool{}
	for i := start; i < len(inputs); i++ {
		if llbTracePathRec(graph, inputs[i], pivot, failed, successful) {
			count++
		}
	}
	return count
}

func Llb_ManTestCuts(graph *Aig_Man_t) {
	aig.Aig_ManFanoutStart(graph)
	defer aig.Aig_ManFanoutStop(graph)
	for _, edge := range graph.LiveEdges() {
		if !aig.Aig_ObjIsNode(graph, edge) || aig.Aig_ObjRefs(graph, edge) <= 1 {
			continue
		}
		fmt.Printf("Obj =%5d.  Lev =%3d.  Fanout =%5d.  Count = %3d.\n", edge.NodeID(), graph.Node(edge).Level, aig.Aig_ObjRefs(graph, edge), Llb_ManTracePaths(graph, edge))
	}
}

func llbLabelCone(graph *aig.Graph, edge aig.Edge, marked map[int]bool) {
	id := edge.NodeID()
	if marked[id] {
		return
	}
	marked[id] = true
	node := graph.Node(aig.Aig_Regular(edge))
	switch node.Kind {
	case aig.Constant, aig.Input:
		return
	case aig.Output, aig.Buffer:
		llbLabelCone(graph, node.Fanins[0], marked)
	case aig.And, aig.ExclusiveOr:
		llbLabelCone(graph, node.Fanins[0], marked)
		llbLabelCone(graph, node.Fanins[1], marked)
	default:
		panic("llb: unsupported node in latch-input cone")
	}
}

func Llb_ManLabelLiCones_rec(graph *Aig_Man_t, object aig.Edge) {
	if aig.Aig_ObjIsMarkB(graph, object) {
		return
	}
	aig.Aig_ObjSetMarkB(graph, object)
	node := graph.Node(aig.Aig_Regular(object))
	if node.Kind == aig.Constant || node.Kind == aig.Input {
		return
	}
	if node.Kind == aig.Output || node.Kind == aig.Buffer {
		Llb_ManLabelLiCones_rec(graph, node.Fanins[0])
		return
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("llb: unsupported node in latch-input cone")
	}
	Llb_ManLabelLiCones_rec(graph, node.Fanins[0])
	Llb_ManLabelLiCones_rec(graph, node.Fanins[1])
}

func Llb_ManLabelLiCones(graph *Aig_Man_t) {
	aig.Aig_ManCleanMarkB(graph)
	aig.Aig_ObjSetMarkB(graph, aig.Aig_ManConst1(graph))
	for _, input := range graph.Inputs() {
		aig.Aig_ObjSetMarkB(graph, input)
	}
	outputs := graph.Outputs()
	start := len(outputs) - int(aig.Aig_ManRegNum(graph))
	for i := start; i < len(outputs); i++ {
		node := graph.Node(outputs[i])
		Llb_ManLabelLiCones_rec(graph, node.Fanins[0])
	}
}

func Llb_ManMarkInternalPivots(graph *Aig_Man_t) {
	muxes := aig.Aig_ManMuxesCollect(graph)
	aig.Aig_ManMuxesDeref(graph, muxes)
	defer aig.Aig_ManMuxesRef(graph, muxes)
	aig.Aig_ManCleanMarkB(graph)
	Llb_ManLabelLiCones(graph)
	aig.Aig_ManFanoutStart(graph)
	defer aig.Aig_ManFanoutStop(graph)
	for _, edge := range graph.LiveEdges() {
		if !aig.Aig_ObjIsNode(graph, edge) || !aig.Aig_ObjIsMarkB(graph, edge) || aig.Aig_ObjRefs(graph, edge) <= 1 {
			continue
		}
		if Llb_ManTracePaths(graph, edge) > 0 {
			aig.Aig_ObjSetMarkA(graph, edge)
		}
	}
	aig.Aig_ManCleanMarkB(graph)
}

func Llb_ManMarkPivotNodes(graph *Aig_Man_t, useInternal int32) *Vec_Int_t {
	if graph == nil {
		return vec.Vec_IntAlloc(0)
	}
	aig.Aig_ManCleanMarkA(graph)
	for _, input := range graph.Inputs() {
		aig.Aig_ObjSetMarkA(graph, input)
	}
	outputs := graph.Outputs()
	start := len(outputs) - int(aig.Aig_ManRegNum(graph))
	for i := start; i < len(outputs); i++ {
		aig.Aig_ObjSetMarkA(graph, outputs[i])
	}
	if useInternal != 0 {
		Llb_ManMarkInternalPivots(graph)
	}
	v := vec.Vec_IntAlloc(len(graph.LiveEdges()))
	for _, input := range graph.Inputs() {
		vec.Vec_IntPush(v, int32(input.NodeID()))
	}
	for _, edge := range graph.LiveEdges() {
		if aig.Aig_ObjIsNode(graph, edge) && aig.Aig_ObjIsMarkA(graph, edge) {
			vec.Vec_IntPush(v, int32(edge.NodeID()))
		}
	}
	for i := start; i < len(outputs); i++ {
		vec.Vec_IntPush(v, int32(outputs[i].NodeID()))
	}
	return v
}
