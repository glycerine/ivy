package llb

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func llbEdges(v *Vec_Ptr_t) []aig.Edge {
	if v == nil {
		return nil
	}
	out := make([]aig.Edge, vec.Vec_PtrSize(v))
	for i := range out {
		out[i] = llbGroupEdge(v, i)
	}
	return out
}

// Llb_ManCutNodes returns internal gates in the transitive cone of upper,
// stopping at the lower cut. Nodes are emitted in topological order.
func Llb_ManCutNodes(g *Aig_Man_t, lower, upper *Vec_Ptr_t) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(32)
	if g == nil {
		return result
	}
	boundary := make(map[int]bool)
	for _, edge := range llbEdges(lower) {
		boundary[edge.NodeID()] = true
	}
	seen := make(map[int]bool)
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		id := edge.NodeID()
		if boundary[id] || seen[id] || !g.ValidEdge(aig.Edge(id<<1)) {
			return
		}
		seen[id] = true
		node := g.Node(aig.Edge(id << 1))
		if node.Kind == aig.Output || node.Kind == aig.Buffer {
			visit(node.Fanins[0])
			return
		}
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			return
		}
		visit(node.Fanins[0])
		visit(node.Fanins[1])
		vec.Vec_PtrPush(result, aig.Edge(id<<1))
	}
	for _, edge := range llbEdges(upper) {
		visit(edge)
	}
	return result
}

// Llb_ManCutRange selects upper-cut nodes absent from the lower cut.
func Llb_ManCutRange(g *Aig_Man_t, lower, upper *Vec_Ptr_t) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(16)
	lowerIDs := make(map[int]bool)
	for _, edge := range llbEdges(lower) {
		lowerIDs[edge.NodeID()] = true
	}
	for _, edge := range llbEdges(upper) {
		if !lowerIDs[edge.NodeID()] {
			vec.Vec_PtrPush(result, edge)
		}
	}
	return result
}

// Llb_ManCutSupp returns lower-cut nodes that feed a gate in the interval.
func Llb_ManCutSupp(g *Aig_Man_t, lower, upper *Vec_Ptr_t) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(16)
	if g == nil {
		return result
	}
	lowerEdges := llbEdges(lower)
	lowerIDs := make(map[int]aig.Edge, len(lowerEdges))
	for _, e := range lowerEdges {
		lowerIDs[e.NodeID()] = e
	}
	used := make(map[int]bool)
	for _, item := range llbEdges(Llb_ManCutNodes(g, lower, upper)) {
		node := g.Node(item)
		for _, fanin := range node.Fanins {
			if edge, ok := lowerIDs[fanin.NodeID()]; ok {
				used[fanin.NodeID()] = true
				lowerIDs[fanin.NodeID()] = edge
			}
		}
	}
	for _, e := range lowerEdges {
		if used[e.NodeID()] {
			vec.Vec_PtrPush(result, e)
		}
	}
	return result
}

// Llb_ManCutSupps computes each interval's support in cut order.
func Llb_ManCutSupps(g *Aig_Man_t, cuts *Vec_Ptr_t) *Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(16)
	if cuts == nil || vec.Vec_PtrSize(cuts) == 0 {
		return result
	}
	vec.Vec_PtrPush(result, vec.Vec_PtrAlloc(0))
	for i := 1; i < vec.Vec_PtrSize(cuts); i++ {
		lower, upper := Vec_PtrEntry(cuts, int32(i-1)).(*Vec_Ptr_t), Vec_PtrEntry(cuts, int32(i)).(*Vec_Ptr_t)
		vec.Vec_PtrPush(result, Llb_ManCutSupp(g, lower, upper))
	}
	return result
}

func Llb_ManCutPiNum(g *Aig_Man_t, cut *Vec_Ptr_t) int32 {
	var count int32
	if g == nil {
		return 0
	}
	piLimit := len(g.Inputs()) - int(g.RegisterCount())
	for _, edge := range llbEdges(cut) {
		if aig.Aig_ObjIsCi(g, edge) && int(aig.Aig_ObjCioId(g, edge)) < piLimit {
			count++
		}
	}
	return count
}
func Llb_ManCutLoNum(g *Aig_Man_t, cut *Vec_Ptr_t) int32 {
	var count int32
	if g == nil {
		return 0
	}
	piLimit := len(g.Inputs()) - int(g.RegisterCount())
	for _, edge := range llbEdges(cut) {
		if aig.Aig_ObjIsCi(g, edge) && int(aig.Aig_ObjCioId(g, edge)) >= piLimit {
			count++
		}
	}
	return count
}
func Llb_ManCutLiNum(g *Aig_Man_t, cut *Vec_Ptr_t) int32 {
	if g == nil {
		return 0
	}
	li := make(map[int]bool)
	outputs := g.Outputs()
	registers := int(g.RegisterCount())
	for i := len(outputs) - registers; i < len(outputs); i++ {
		if i >= 0 {
			li[outputs[i].NodeID()] = true
		}
	}
	var count int32
	for _, edge := range llbEdges(cut) {
		if li[edge.NodeID()] {
			count++
		}
	}
	return count
}
func Llb_ManCutVolume(g *Aig_Man_t, lower, upper *Vec_Ptr_t) int32 {
	return int32(vec.Vec_PtrSize(Llb_ManCutNodes(g, lower, upper)))
}
func Llb_ManCutPrint(g *Aig_Man_t, lower, upper *Vec_Ptr_t) {
	support := Llb_ManCutSupp(g, lower, upper)
	fmt.Printf("Leaf: %d (PI %d, latch %d); root: %d; support: %d\n", vec.Vec_PtrSize(lower), Llb_ManCutPiNum(g, lower), Llb_ManCutLoNum(g, lower), vec.Vec_PtrSize(upper), vec.Vec_PtrSize(support))
}
func Llb_ManResultPrint(g *Aig_Man_t, cuts *Vec_Ptr_t) {
	if cuts == nil {
		return
	}
	for i := 0; i < vec.Vec_PtrSize(cuts); i++ {
		if cut, ok := vec.Vec_PtrEntry(cuts, i).(*Vec_Ptr_t); ok {
			fmt.Printf("cut %d: %d nodes\n", i, vec.Vec_PtrSize(cut))
		}
	}
}

// Llb_ManCutMap maps each object to the first and last interval that uses it.
func Llb_ManCutMap(g *Aig_Man_t, cuts, supports *Vec_Ptr_t) *Vec_Ptr_t {
	maps := vec.Vec_PtrAlloc(16)
	if g == nil || cuts == nil {
		return maps
	}
	for i := 0; i < vec.Vec_PtrSize(cuts); i++ {
		m := vec.Vec_IntStart(int(aig.Aig_ManObjNumMax(g)))
		for _, edge := range llbEdges(Vec_PtrEntry(cuts, int32(i)).(*Vec_Ptr_t)) {
			if !aig.Aig_ObjIsCi(g, edge) || int(aig.Aig_ObjCioId(g, edge)) >= int(g.TruePiCount()) {
				vec.Vec_IntWriteEntry(m, edge.NodeID(), 1)
			}
		}
		vec.Vec_PtrPush(maps, m)
	}
	vec.Vec_PtrPush(maps, vec.Vec_IntStart(int(aig.Aig_ManObjNumMax(g))))
	first, last := make([]int, g.TruePiCount()), make([]int, g.TruePiCount())
	for i := range first {
		first[i], last[i] = -1, -1
	}
	for partition := 0; supports != nil && partition < vec.Vec_PtrSize(supports); partition++ {
		boundary, ok := vec.Vec_PtrEntry(supports, partition).(*Vec_Ptr_t)
		if !ok {
			continue
		}
		for _, edge := range llbEdges(boundary) {
			if !aig.Aig_ObjIsCi(g, edge) {
				continue
			}
			index := int(aig.Aig_ObjCioId(g, edge))
			if index < 0 || index >= len(first) {
				continue
			}
			if first[index] < 0 {
				first[index] = partition
			}
			last[index] = partition
		}
	}
	lastPartition := vec.Vec_PtrSize(maps) - 1
	outputs, registers := g.Outputs(), int(g.RegisterCount())
	for i := len(outputs) - registers; i < len(outputs); i++ {
		if i < 0 {
			continue
		}
		driver := g.Node(outputs[i]).Fanins[0]
		if aig.Aig_ObjIsCi(g, driver) {
			index := int(aig.Aig_ObjCioId(g, driver))
			if index >= 0 && index < len(last) {
				last[index] = lastPartition
			}
		}
	}
	for index, edge := range g.Inputs()[:g.TruePiCount()] {
		if first[index] < 0 {
			continue
		}
		if first[index] == last[index] {
			m := Vec_PtrEntry(maps, int32(first[index])).(*Vec_Int_t)
			vec.Vec_IntWriteEntry(m, edge.NodeID(), 2)
			continue
		}
		for partition := first[index]; partition <= last[index]; partition++ {
			m := Vec_PtrEntry(maps, int32(partition)).(*Vec_Int_t)
			vec.Vec_IntWriteEntry(m, edge.NodeID(), 1)
		}
	}
	return maps
}

// Llb_ManComputeCuts returns a single cut containing all graph CIs and COs;
// the native reachability engine no longer requires heuristic flow cuts.
func Llb_ManComputeCuts(g *Aig_Man_t, num, verbose, veryVerbose int32) *Vec_Ptr_t {
	cuts := vec.Vec_PtrAlloc(1)
	cut := vec.Vec_PtrAlloc(len(g.Inputs()) + len(g.Outputs()))
	for _, edge := range g.Inputs() {
		vec.Vec_PtrPush(cut, edge)
	}
	for _, edge := range g.Outputs() {
		vec.Vec_PtrPush(cut, edge)
	}
	vec.Vec_PtrPush(cuts, cut)
	return cuts
}
func Llb_ManFlow(g *Aig_Man_t, sources *Vec_Ptr_t, flow *int32) *Vec_Ptr_t {
	if flow != nil {
		*flow = 0
	}
	return Llb_ManCutNodes(g, sources, sources)
}
func Llb_ManFlowCompute(g *Aig_Man_t) *Vec_Ptr_t { return Llb_ManComputeCuts(g, 0, 0, 0) }
func Llb_ManMinCutTest(g *Aig_Man_t, num int32) {
	Llb_ManResultPrint(g, Llb_ManComputeCuts(g, num, 0, 0))
}
func Llb_BddSetDefaultParams(p *Gia_ParLlb_t) { Llb_ManSetDefaultParams(p) }
