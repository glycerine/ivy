package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func edgeValue(object *Aig_Obj_t) aig.Edge {
	if object == nil {
		return aig.InvalidEdge
	}
	return *object
}
func pushEdgeUnique(list *Vec_Ptr_t, edge aig.Edge) {
	for _, value := range vec.Vec_PtrArray(list) {
		if value == edge {
			return
		}
	}
	vec.Vec_PtrPush(list, edge)
}

func Cnf_CollectLeaves_rec(graph *aig.Graph, root, object *Aig_Obj_t, leaves *Vec_Ptr_t, stopComplement int32) {
	if root == nil || object == nil {
		return
	}
	rootEdge, edge := *root, *object
	if graph == nil {
		pushEdgeUnique(leaves, edge)
		return
	}
	node := graph.Node(aig.Aig_Regular(edge))
	if edge != rootEdge && (node.MarkA || stopComplement != 0 && edge.Complemented()) {
		if stopComplement != 0 {
			pushEdgeUnique(leaves, edge)
		} else {
			pushEdgeUnique(leaves, aig.Aig_Regular(edge))
		}
		return
	}
	if node.Kind != aig.And {
		pushEdgeUnique(leaves, aig.Aig_Regular(edge))
		return
	}
	if stopComplement != 0 {
		left, right := aig.Aig_ObjChild0(graph, edge), aig.Aig_ObjChild1(graph, edge)
		Cnf_CollectLeaves_rec(graph, root, &left, leaves, stopComplement)
		Cnf_CollectLeaves_rec(graph, root, &right, leaves, stopComplement)
	} else {
		left, right := node.Fanins[0], node.Fanins[1]
		Cnf_CollectLeaves_rec(graph, root, &left, leaves, stopComplement)
		Cnf_CollectLeaves_rec(graph, root, &right, leaves, stopComplement)
	}
}

func Cnf_CollectLeaves(graph *aig.Graph, root *Aig_Obj_t, leaves *Vec_Ptr_t, stopComplement int32) {
	vec.Vec_PtrClear(leaves)
	Cnf_CollectLeaves_rec(graph, root, root, leaves, stopComplement)
}
func CollectAIGLeaves(graph *aig.Graph, root aig.Edge, stopComplement bool) []aig.Edge {
	if graph == nil {
		return nil
	}
	var stop int32
	if stopComplement {
		stop = 1
	}
	collected := vec.Vec_PtrAlloc(0)
	Cnf_CollectLeaves(graph, &root, collected, stop)
	leaves := make([]aig.Edge, 0, vec.Vec_PtrSize(collected))
	for _, leaf := range vec.Vec_PtrArray(collected) {
		leaves = append(leaves, leaf.(aig.Edge))
	}
	return leaves
}

func Cnf_CollectVolume_rec(graph *Aig_Man_t, object *Aig_Obj_t, nodes *Vec_Ptr_t) {
	if graph == nil || object == nil {
		return
	}
	edge := *object
	if aig.Aig_ObjIsTravIdCurrent(graph, edge) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(graph, edge)
	node := graph.Node(edge)
	if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
		left, right := node.Fanins[0], node.Fanins[1]
		Cnf_CollectVolume_rec(graph, &left, nodes)
		Cnf_CollectVolume_rec(graph, &right, nodes)
	}
	vec.Vec_PtrPush(nodes, edge)
}
func Cnf_CollectVolume(graph *Aig_Man_t, root *Aig_Obj_t, leaves, nodes *Vec_Ptr_t) {
	if graph == nil || root == nil {
		return
	}
	aig.Aig_ManIncrementTravId(graph)
	for _, value := range vec.Vec_PtrArray(leaves) {
		if edge, ok := value.(aig.Edge); ok {
			aig.Aig_ObjSetTravIdCurrent(graph, edge)
		}
	}
	vec.Vec_PtrClear(nodes)
	Cnf_CollectVolume_rec(graph, root, nodes)
}

// Cnf_CutDeriveTruth evaluates a cone over up to six listed leaves.
func Cnf_CutDeriveTruth(graph *Aig_Man_t, leaves, nodes *Vec_Ptr_t) uint64 {
	if graph == nil || vec.Vec_PtrSize(leaves) > 6 || vec.Vec_PtrSize(nodes) == 0 {
		return 0
	}
	leafIDs := make([]aig.Edge, vec.Vec_PtrSize(leaves))
	for i, value := range vec.Vec_PtrArray(leaves) {
		leafIDs[i], _ = value.(aig.Edge)
	}
	root, _ := vec.Vec_PtrEntryLast(nodes).(aig.Edge)
	var truth uint64
	// ABC truth words repeat the cut's table across all 64 bits.
	for assignment := 0; assignment < 64; assignment++ {
		values := make(map[int]bool, vec.Vec_PtrSize(nodes)+len(leafIDs))
		for i, edge := range leafIDs {
			values[edge.NodeID()] = (assignment&(1<<i) != 0) != edge.Complemented()
		}
		var eval func(aig.Edge) bool
		eval = func(edge aig.Edge) bool {
			id := edge.NodeID()
			var value bool
			if id == 0 {
				value = true
			} else if got, ok := values[id]; ok {
				value = got
			} else {
				node := graph.Node(edge)
				switch node.Kind {
				case aig.Input:
					value = false
				case aig.And:
					value = eval(node.Fanins[0]) && eval(node.Fanins[1])
				case aig.ExclusiveOr:
					value = eval(node.Fanins[0]) != eval(node.Fanins[1])
				case aig.Buffer, aig.Output:
					value = eval(node.Fanins[0])
				}
			}
			return value != edge.Complemented()
		}
		if eval(root) {
			truth |= uint64(1) << assignment
		}
	}
	return truth
}

func Cnf_ObjGetLit(mapping *Vec_Int_t, object *Aig_Obj_t, complemented int32) int32 {
	if object == nil {
		return -1
	}
	if mapping == nil {
		return int32((*object).NodeID())<<1 | complemented&1
	}
	variable := vec.Vec_IntEntry(mapping, int((*object).NodeID()))
	if variable < 0 {
		return -1
	}
	return variable<<1 | complemented&1
}

// largeAndLeaves identifies the supergate case used by ABC for cuts larger
// than six leaves. A complemented internal edge is not an AND supergate.
func largeAndLeaves(graph *Aig_Man_t, root aig.Edge, leaves *Vec_Ptr_t) []aig.Edge {
	boundary := make(map[int]bool)
	for _, value := range vec.Vec_PtrArray(leaves) {
		edge, ok := value.(aig.Edge)
		if !ok {
			panic("cnf: invalid cut leaf")
		}
		boundary[edge.NodeID()] = true
	}
	var result []aig.Edge
	seen := make(map[aig.Edge]bool)
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		if boundary[edge.NodeID()] {
			if !seen[edge] {
				result = append(result, edge)
				seen[edge] = true
			}
			return
		}
		node := graph.Node(edge)
		if edge.Complemented() || node.Kind != aig.And {
			panic("cnf: large cut is not an AND supergate")
		}
		visit(node.Fanins[0])
		visit(node.Fanins[1])
	}
	visit(root)
	return result
}

// Cnf_ComputeClauses emits truth-table clauses for a root and its cut leaves.
func Cnf_ComputeClauses(graph *Aig_Man_t, root *Aig_Obj_t, leaves, nodes *Vec_Ptr_t, mapping, _ *Vec_Int_t, clauses *Vec_Int_t) {
	if graph == nil || root == nil {
		return
	}
	vec.Vec_IntClear(clauses)
	if vec.Vec_PtrSize(leaves) > 6 {
		inputs := largeAndLeaves(graph, *root, leaves)
		out := Cnf_ObjGetLit(mapping, root, 0)
		vec.Vec_IntPush(clauses, 0)
		vec.Vec_IntPush(clauses, out)
		for _, edge := range inputs {
			var sign int32
			if edge.Complemented() {
				sign = 1
			}
			lit := Cnf_ObjGetLit(mapping, &edge, sign)
			vec.Vec_IntPush(clauses, lit^1)
		}
		for _, edge := range inputs {
			var sign int32
			if edge.Complemented() {
				sign = 1
			}
			lit := Cnf_ObjGetLit(mapping, &edge, sign)
			vec.Vec_IntPush(clauses, 0)
			vec.Vec_IntPush(clauses, out^1)
			vec.Vec_IntPush(clauses, lit)
		}
		return
	}
	truth := Cnf_CutDeriveTruth(graph, leaves, nodes)
	out := Cnf_ObjGetLit(mapping, root, 0)
	n := vec.Vec_PtrSize(leaves)
	for assignment := 0; assignment < 1<<n; assignment++ {
		value := truth&(uint64(1)<<assignment) != 0
		vec.Vec_IntPush(clauses, 0)
		if value {
			vec.Vec_IntPush(clauses, out)
		} else {
			vec.Vec_IntPush(clauses, out^1)
		}
		for i, item := range vec.Vec_PtrArray(leaves) {
			edge, ok := item.(aig.Edge)
			if !ok {
				continue
			}
			sign := int32(assignment >> i & 1)
			if edge.Complemented() {
				sign ^= 1
			}
			lit := Cnf_ObjGetLit(mapping, &edge, sign)
			vec.Vec_IntPush(clauses, lit)
		}
	}
}

func Cnf_DeriveFastMark(graph *Aig_Man_t) {
	if graph == nil {
		return
	}
	for id, node := range graph.Nodes() {
		if id == 0 {
			continue
		}
		if node.Kind == aig.And || node.Kind == aig.ExclusiveOr || node.Kind == aig.Buffer {
			aig.Aig_ObjSetMarkA(graph, aig.Edge(id)<<1)
		}
	}
}
func Cnf_CutCountClauses(graph *Aig_Man_t, leaves, nodes *Vec_Ptr_t, _ *Vec_Int_t) int32 {
	if graph == nil {
		return 0
	}
	n := vec.Vec_PtrSize(leaves)
	if n > 6 {
		if vec.Vec_PtrSize(nodes) == 0 {
			panic("cnf: large cut has no root")
		}
		root, ok := vec.Vec_PtrEntryLast(nodes).(aig.Edge)
		if !ok {
			panic("cnf: invalid cut root")
		}
		return int32(len(largeAndLeaves(graph, root, leaves)) + 1)
	}
	return int32(1 << n)
}
func Cnf_CountCnfSize(graph *Aig_Man_t) int32 {
	if graph == nil {
		return 0
	}
	var count int32
	for _, node := range graph.Nodes() {
		if node.Kind == aig.And {
			count += 3
		} else if node.Kind == aig.ExclusiveOr {
			count += 4
		}
	}
	return count + int32(graph.OutputCount())
}
func Cnf_DeriveFastClauses(graph *Aig_Man_t, outputs int32) *Cnf_Dat_t {
	return deriveLegacyAIG(graph, outputs)
}
func Cnf_DeriveFast(graph *Aig_Man_t, outputs int32) *Cnf_Dat_t {
	return Cnf_DeriveFastClauses(graph, outputs)
}
