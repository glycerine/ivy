package abci

import (
	"cmp"
	"fmt"
	"os"
	"slices"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_NtkBalance globally balances the AIG by the number of logic levels.
func Abc_NtkBalance(n *Abc_Ntk_t, duplicate, selective, updateLevel int32) *Abc_Ntk_t {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 {
		panic("Abc_NtkBalance requires a strashed network")
	}
	if selective != 0 {
		abc.Abc_NtkStartReverseLevels(n, 0)
		Abc_NtkMarkCriticalNodes(n)
	}
	balanced := abc.Abc_NtkStartFrom(n, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	balanced.ReserveAnds(int(abc.Abc_NtkNodeNum(n))) // a size hint: balancing keeps about as many ANDs
	balanced.ReserveAIG(0, 0, 0, int(abc.Abc_NtkNodeNum(n)))
	Abc_NtkBalancePerform(n, balanced, duplicate, selective, updateLevel)
	abc.Abc_NtkFinalize(n, balanced)
	abc.Abc_AigCleanup(balanceAigManager(balanced))
	if selective != 0 {
		abc.Abc_NtkStopReverseLevels(n)
		abc.Abc_NtkCleanMarkA(n)
	}
	if n.Exdc != nil {
		balanced.Exdc = abc.Abc_NtkDup(n.Exdc)
	}
	if abc.Abc_NtkCheck(balanced) == 0 {
		fmt.Fprintln(os.Stdout, "Abc_NtkBalance: The network check has failed.")
		abc.Abc_NtkDelete(balanced)
		return nil
	}
	return balanced
}

// Abc_NtkBalancePerform balances each CO cone in the destination network.
func Abc_NtkBalancePerform(old, balanced *Abc_Ntk_t, duplicate, selective, updateLevel int32) {
	// The C helper initializes all node levels from arrival times before the
	// balancing pass. Abc_NtkLevel uses the network's CILevel callback.
	abc.Abc_NtkLevel(old)
	for i := int32(0); i < abc.Abc_NtkCiNum(old); i++ {
		oldCI, newCI := abc.Abc_NtkCi(old, i), abc.Abc_NtkCi(balanced, i)
		if oldCI != nil && newCI != nil {
			oldCI.CopyEdge = newCI.Signal()
			oldCI.CopyEdge.Object.Level = oldCI.Level
		}
	}
	progress := extra.Extra_ProgressBarStart(os.Stdout, abc.Abc_NtkCoNum(old))
	cos := old.Outputs()
	if old.Barriers != 0 {
		// Abc_NtkForEachLiPo starts with the final barrier COs, then visits
		// ordinary COs in their original order.
		ordered := make([]*Abc_Obj_t, 0, len(cos))
		start := len(cos) - old.Barriers
		ordered = append(ordered, cos[start:]...)
		ordered = append(ordered, cos[:start]...)
		cos = ordered
	}
	for i, co := range cos {
		extra.Extra_ProgressBarUpdate(progress, int32(i), "")
		balanceNode(balanced, abc.Abc_ObjFanin0(co).Signal(), nil, 0, duplicate, selective, updateLevel)
		if old.Barriers != 0 && i < old.Barriers {
			firstFanout := abc.Abc_ObjFanout0(co)
			if firstFanout != nil && abc.Abc_ObjFanoutNum(firstFanout) > 0 {
				abc.Abc_ObjFanout0(firstFanout).Level = abc.Abc_ObjFanin0(co).Level
			}
		}
	}
	extra.Extra_ProgressBarStop(progress)
}

// Abc_NodeBalanceFindLeft finds the earliest same-level candidate for pairing.
func Abc_NodeBalanceFindLeft(super *Vec_Ptr_t) int32 {
	return int32(balanceFindLeft(vec.Vec_PtrArray(super), edgeFromAny))
}

func balanceFindLeft[T any](entries []T, asEdge func(T) abc.Edge) int {
	if len(entries) < 3 {
		return 0
	}
	i := len(entries) - 2
	level := asEdge(entries[i]).Object.Level
	for i > 0 && asEdge(entries[i-1]).Object.Level == level {
		i--
	}
	return i
}

func balanceAigManager(n *Abc_Ntk_t) *abc.Abc_Aig_t {
	if manager, ok := n.FunctionManager.(*abc.Abc_Aig_t); ok {
		return manager
	}
	return abc.Abc_AigAlloc(n)
}

// Abc_NodeBalancePermute moves a pair that hashes to an existing AND toward
// the end of the sorted supergate, preserving the original sharing heuristic.
func Abc_NodeBalancePermute(n *Abc_Ntk_t, super *Vec_Ptr_t, leftBound int32) {
	balancePermute(n, vec.Vec_PtrArray(super), int(leftBound), edgeFromAny)
}

func balancePermute[T any](n *Abc_Ntk_t, entries []T, left int, asEdge func(T) abc.Edge) {
	right := len(entries) - 2
	if left > right {
		panic("invalid balance candidate range")
	}
	if left == right {
		return
	}
	first := asEdge(entries[right+1])
	manager := balanceAigManager(n)
	for i := right; i >= left; i-- {
		if manager.LookupAndEdges(first, asEdge(entries[i])).Object != nil {
			entries[i], entries[right] = entries[right], entries[i]
			return
		}
	}
}

// Abc_NodeBalance_rec recursively rebalances the implication supergate rooted
// at oldEdge and returns its equivalent edge in the destination AIG.
func Abc_NodeBalance_rec(n *Abc_Ntk_t, old *Abc_Obj_t, storage *Vec_Vec_t, level, duplicate, selective, updateLevel int32) *Abc_Obj_t {
	return balanceNode(n, old.Signal(), storage, level, duplicate, selective, updateLevel).Object
}

func balanceNode(n *Abc_Ntk_t, oldEdge abc.Edge, storage *Vec_Vec_t, level, duplicate, selective, updateLevel int32) abc.Edge {
	old := oldEdge.Object
	if old == nil || oldEdge.Complement {
		panic("Abc_NodeBalance_rec expects a regular node")
	}
	if old.CopyEdge.Object != nil {
		return old.CopyEdge
	}
	if abc.Abc_ObjIsNode(old) == 0 {
		panic("balance root is not an AND node")
	}
	var local [8]abc.Edge
	edges := local[:0]
	var super *Vec_Ptr_t
	if storage == nil {
		collected, result := collectBalanceEdges(old.Signal(), balanceLeaves{edges: edges}, 1, duplicate, selective)
		edges = collected.edges
		for _, e := range edges {
			e.Object.MarkB = false
		}
		if result < 0 {
			edges = edges[:0]
		}
	} else {
		super = balanceCone(old.Signal(), storage, level, duplicate, selective)
		for _, item := range vec.Vec_PtrArray(super) {
			edges = append(edges, edgeFromAny(item))
		}
	}
	if len(edges) == 0 {
		old.CopyEdge = abc.Abc_ObjNot(abc.Abc_AigConst1(n))
		return old.CopyEdge
	}
	for i, edge := range edges {
		newEdge := balanceNode(n, edge.Object.Signal(), storage, level+1, duplicate, selective, updateLevel)
		if edge.Complement {
			newEdge = newEdge.Not()
		}
		edges[i] = newEdge
	}
	if len(edges) < 2 {
		fmt.Fprintln(os.Stdout, "BUG!")
	}
	slices.SortStableFunc(edges, balanceEdgeOrder)
	manager := balanceAigManager(n)
	for len(edges) > 1 {
		left := 0
		if updateLevel != 0 {
			left = balanceFindLeft(edges, identityBalanceEdge)
		}
		balancePermute(n, edges, left, identityBalanceEdge)
		end := len(edges) - 2
		combined := manager.AndEdges(edges[end+1], edges[end])
		combined.Object.Level = int(abc.Abc_ObjLevelNew(combined.Object))
		edges = insertBalanceEdge(edges[:end], combined)
	}
	old.CopyEdge = edges[0]
	if super != nil {
		vec.Vec_PtrClear(super)
	}
	return old.CopyEdge
}

// Abc_NodeBalanceCone collects the flattened AND supergate at a given depth.
func Abc_NodeBalanceCone(root *Abc_Obj_t, storage *Vec_Vec_t, level, duplicate, selective int32) *Vec_Ptr_t {
	return balanceCone(root.Signal(), storage, level, duplicate, selective)
}

func balanceCone(root abc.Edge, storage *Vec_Vec_t, level, duplicate, selective int32) *Vec_Ptr_t {
	if root.Complement {
		panic("Abc_NodeBalanceCone expects a regular node")
	}
	for vec.Vec_VecSize(storage) <= int(level) {
		vec.Vec_VecPush(storage, int(level), nil)
	}
	nodes := vec.Vec_VecEntry(storage, int(level))
	vec.Vec_PtrClear(nodes)
	ret := balanceConeRec(root, nodes, 1, duplicate, selective)
	for _, item := range vec.Vec_PtrArray(nodes) {
		edge := edgeFromAny(item)
		edge.Object.MarkB = false
	}
	if ret == -1 {
		vec.Vec_PtrClear(nodes)
	}
	return nodes
}

// Abc_NodeBalanceCone_rec flattens AND nodes and detects repeated polarities.
func Abc_NodeBalanceCone_rec(node *Abc_Obj_t, super *Vec_Ptr_t, first, duplicate, selective int32) int32 {
	return balanceConeRec(node.Signal(), super, first, duplicate, selective)
}

func balanceConeRec(edge abc.Edge, super *Vec_Ptr_t, first, duplicate, selective int32) int32 {
	if super == nil {
		panic("nil balance leaf vector")
	}
	_, result := collectBalanceEdges(edge, balanceLeaves{vector: super}, first, duplicate, selective)
	return result
}

// Public cone calls append to their supplied ABC vector. The whole-network
// balancer uses typed edges so a recursion frame can translate them in place.
type balanceLeaves struct {
	edges  []abc.Edge
	vector *Vec_Ptr_t
}

func repeatBalanceLeaf[T any](entries []T, edge abc.Edge, asEdge func(T) abc.Edge) int32 {
	opposite := false
	for _, entry := range entries {
		e := asEdge(entry)
		if e == edge {
			return 1
		}
		if e == edge.Not() {
			opposite = true
		}
	}
	if opposite {
		return -1
	}
	panic("marked balance node missing from supergate")
}
func collectBalanceEdges(edge abc.Edge, leaves balanceLeaves, first, duplicate, selective int32) (balanceLeaves, int32) {
	node := edge.Object
	if node.MarkB {
		if leaves.vector != nil {
			return leaves, repeatBalanceLeaf(vec.Vec_PtrArray(leaves.vector), edge, edgeFromAny)
		}
		return leaves, repeatBalanceLeaf(leaves.edges, edge, identityBalanceEdge)
	}
	count := len(leaves.edges)
	if leaves.vector != nil {
		count = vec.Vec_PtrSize(leaves.vector)
	}
	if first == 0 && (edge.Complement || abc.Abc_ObjIsNode(node) == 0 || (duplicate == 0 && selective == 0 && abc.Abc_ObjFanoutNum(node) > 1) || count > 10000) {
		if leaves.vector != nil {
			vec.Vec_PtrPush(leaves.vector, edge)
		} else {
			leaves.edges = append(leaves.edges, edge)
		}
		node.MarkB = true
		return leaves, 0
	}
	if edge.Complement || abc.Abc_ObjIsNode(node) == 0 {
		panic("invalid node in balance cone")
	}
	leaves, r1 := collectBalanceEdges(abc.Abc_ObjChild0(node), leaves, 0, duplicate, selective)
	leaves, r2 := collectBalanceEdges(abc.Abc_ObjChild1(node), leaves, 0, duplicate, selective)
	if r1 == -1 || r2 == -1 {
		return leaves, -1
	}
	if r1 != 0 || r2 != 0 {
		return leaves, 1
	}
	return leaves, 0
}

// Abc_NodeBalanceConeExor_rec collects leaves of an XOR-marked cone.
func Abc_NodeBalanceConeExor_rec(node *Abc_Obj_t, super *Vec_Ptr_t, first int32) int32 {
	return balanceConeExorRec(node.Signal(), super, first)
}

func balanceConeExorRec(edge abc.Edge, super *Vec_Ptr_t, first int32) int32 {
	for _, item := range vec.Vec_PtrArray(super) {
		if edgeFromAny(item) == edge {
			return 1
		}
	}
	node := edge.Object
	if first == 0 && (abc.Abc_NodeIsExorType(node) == 0 || abc.Abc_ObjIsNode(node) == 0) {
		vec.Vec_PtrPush(super, edge)
		return 0
	}
	if edge.Complement || abc.Abc_ObjIsNode(node) == 0 || abc.Abc_NodeIsExorType(node) == 0 {
		panic("invalid XOR node in balance cone")
	}
	fanins := node.Fanins()
	operands := fanins[0].Object.Fanins()
	if len(operands) != 2 {
		return -1
	}
	r1 := balanceConeExorRec(operands[0], super, 0)
	r2 := balanceConeExorRec(operands[1], super, 0)
	if r1 == -1 || r2 == -1 {
		return -1
	}
	if r1 != 0 || r2 != 0 {
		return 1
	}
	return 0
}

// Abc_NodeBalanceConeExor retains the C helper's diagnostic-only behavior.
func Abc_NodeBalanceConeExor(node *Abc_Obj_t) *Vec_Ptr_t {
	if abc.Abc_NodeIsExorType(node) == 0 {
		return nil
	}
	super := vec.Vec_PtrAlloc(10)
	balanceConeExorRec(node.Signal(), super, 1)
	fmt.Printf("%d ", vec.Vec_PtrSize(super))
	vec.Vec_PtrFree(super)
	return nil
}

// Abc_NodeFindCone_rec recursively attaches implication supergates as CopyData.
func Abc_NodeFindCone_rec(node *Abc_Obj_t) *Vec_Ptr_t {
	if node == nil || abc.Abc_ObjIsCi(node) != 0 {
		return nil
	}
	var nodes *Vec_Ptr_t
	if abc.Abc_NodeIsMuxType(node) != 0 {
		var thenEdge, elseEdge abc.Edge
		control := abc.Abc_NodeRecognizeMux(node, &thenEdge, &elseEdge)
		nodes = vec.Vec_PtrAlloc(3)
		vec.Vec_PtrPush(nodes, control.Signal())
		vec.Vec_PtrPushUnique(nodes, thenEdge.Object.Signal())
		vec.Vec_PtrPushUnique(nodes, elseEdge.Object.Signal())
	} else {
		nodes = vec.Vec_PtrAlloc(4)
		ret := balanceConeRec(node.Signal(), nodes, 1, 1, 0)
		for _, item := range vec.Vec_PtrArray(nodes) {
			edgeFromAny(item).Object.MarkB = false
		}
		if ret == -1 {
			vec.Vec_PtrClear(nodes)
		}
	}
	for _, item := range vec.Vec_PtrArray(nodes) {
		edge := edgeFromAny(item)
		fanin := edge.Object
		if fanin.CopyData != nil {
			continue
		}
		fanin.CopyData = Abc_NodeFindCone_rec(fanin)
	}
	return nodes
}

// Abc_NtkBalanceAttach attaches implication supergates to all CO drivers.
func Abc_NtkBalanceAttach(n *Abc_Ntk_t) {
	abc.Abc_NtkCleanCopy(n)
	for _, co := range n.Outputs() {
		driver := abc.Abc_ObjFanin0(co)
		if driver.CopyData == nil {
			driver.CopyData = Abc_NodeFindCone_rec(driver)
		}
	}
}

// Abc_NtkBalanceDetach frees the attached cone vectors.
func Abc_NtkBalanceDetach(n *Abc_Ntk_t) {
	for _, node := range abc.Abc_NtkNodes(n) {
		if super, ok := node.CopyData.(*Vec_Ptr_t); ok && super != nil {
			vec.Vec_PtrFree(super)
			node.CopyData = nil
		}
	}
}

// Abc_NtkBalanceLevel_rec computes the implication-supergate depth.
func Abc_NtkBalanceLevel_rec(node *Abc_Obj_t) int32 {
	if node.Level > 0 || abc.Abc_ObjIsCi(node) != 0 {
		return int32(node.Level)
	}
	super, ok := node.CopyData.(*Vec_Ptr_t)
	if !ok || super == nil {
		panic("missing attached balance cone")
	}
	max := int32(0)
	for _, item := range vec.Vec_PtrArray(super) {
		fanin := edgeFromAny(item).Object
		if level := Abc_NtkBalanceLevel_rec(fanin); level > max {
			max = level
		}
	}
	node.Level = int(max + 1)
	return max + 1
}

// Abc_NtkBalanceLevel computes implication-supergate levels from all COs.
func Abc_NtkBalanceLevel(n *Abc_Ntk_t) {
	for _, node := range n.Objects() {
		node.Level = 0
	}
	for _, co := range n.Outputs() {
		Abc_NtkBalanceLevel_rec(abc.Abc_ObjFanin0(co))
	}
}

// Abc_NtkMarkCriticalNodes marks nodes within one level of the required time.
func Abc_NtkMarkCriticalNodes(n *Abc_Ntk_t) {
	var count int32
	for _, node := range abc.Abc_NtkNodes(n) {
		if abc.Abc_ObjRequiredLevel(node)-int32(node.Level) <= 1 {
			node.MarkA = true
			count++
		}
	}
	percent := float64(0)
	if nodes := abc.Abc_NtkNodeNum(n); nodes != 0 {
		percent = 100 * float64(count) / float64(nodes)
	}
	fmt.Printf("The number of nodes on the critical paths = %6d  (%5.2f %%)\n", count, percent)
}

func edgeFromAny(value any) abc.Edge {
	switch edge := value.(type) {
	case abc.Edge:
		return edge
	case *Abc_Obj_t:
		return edge.Signal()
	default:
		panic(fmt.Sprintf("unexpected balance edge %T", value))
	}
}

func identityBalanceEdge(edge abc.Edge) abc.Edge { return edge }

func balanceEdgeOrder(a, b abc.Edge) int {
	if a.Object.Level != b.Object.Level {
		return cmp.Compare(b.Object.Level, a.Object.Level)
	}
	return cmp.Compare(b.Object.ID, a.Object.ID)
}

func insertBalanceEdge(edges []abc.Edge, edge abc.Edge) []abc.Edge {
	if slices.Contains(edges, edge) {
		return edges
	}
	i := len(edges)
	edges = append(edges, edge)
	for i > 0 && edge.Object.Level > edges[i-1].Object.Level {
		edges[i] = edges[i-1]
		i--
	}
	edges[i] = edge
	return edges
}
