package abc

import (
	"fmt"
	"math"
	"sort"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/reo"
	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_NodeMinimumBase removes inputs absent from a local BDD's support and
// renumbers the remaining variables to match the compacted fanin list.
func Abc_NodeMinimumBase(node *Object) int32 {
	if node == nil || node.Kind != ObjectNode || Abc_NtkIsBddLogic(node.network) == 0 {
		panic("minimum base requires a BDD logic node")
	}
	dd := node.network.FunctionManager.(*core.DdManager)
	old := node.Data.(*core.DdNode)
	support := dd.Support(old)
	sort.Ints(support)
	if len(support) == len(node.fanins) {
		return 0
	}
	vector := append([]*core.DdNode(nil), dd.Vars...)
	used := make([]bool, len(node.fanins))
	for newIndex, oldIndex := range support {
		if oldIndex < 0 || oldIndex >= len(node.fanins) {
			panic("BDD support exceeds node fanins")
		}
		used[oldIndex] = true
		vector[oldIndex] = core.Cudd_bddIthVar(dd, newIndex)
	}
	mapped := core.Cudd_bddVectorCompose(dd, old, vector)
	if mapped == nil {
		panic("BDD support remapping failed")
	}
	core.Cudd_Ref(mapped)
	core.Cudd_RecursiveDeref(dd, old)
	node.Data = mapped
	for i := len(used) - 1; i >= 0; i-- {
		if !used[i] {
			if err := node.network.DeleteFanin(node, i); err != nil {
				panic(err)
			}
		}
	}
	return 1
}

func Abc_NtkMinimumBase(n *Network) int32 {
	var changed int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		changed += Abc_NodeMinimumBase(node)
	}
	return changed
}

// Abc_NodeRemoveDupFanins_int substitutes one repeated input, then compacts its support.
func Abc_NodeRemoveDupFanins_int(node *Object) int32 {
	if node == nil || node.Kind != ObjectNode || Abc_NtkIsBddLogic(node.network) == 0 {
		panic("duplicate removal requires a BDD logic node")
	}
	dd := node.network.FunctionManager.(*core.DdManager)
	for i, fanin := range node.fanins {
		for j := 0; j < i; j++ {
			prior := node.fanins[j]
			if fanin.Object != prior.Object {
				continue
			}
			vector := append([]*core.DdNode(nil), dd.Vars...)
			replacement := core.Cudd_bddIthVar(dd, j)
			if prior.Complement != fanin.Complement {
				replacement = core.Cudd_Not(replacement)
			}
			vector[i] = replacement
			old := node.Data.(*core.DdNode)
			mapped := core.Cudd_bddVectorCompose(dd, old, vector)
			if mapped == nil {
				panic("BDD duplicate-fanin substitution failed")
			}
			core.Cudd_Ref(mapped)
			core.Cudd_RecursiveDeref(dd, old)
			node.Data = mapped
			Abc_NodeMinimumBase(node)
			return 1
		}
	}
	return 0
}

func Abc_NodeRemoveDupFanins(node *Object) int32 {
	var count int32
	for Abc_NodeRemoveDupFanins_int(node) != 0 {
		count++
	}
	return count
}

func Abc_NtkRemoveDupFanins(n *Network) int32 {
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		count += Abc_NodeRemoveDupFanins(node)
	}
	return count
}

func Abc_NodeCheckDupFanin(fanin, fanout *Object, index *int32) int32 {
	var count int32
	for i, edge := range fanout.fanins {
		if edge.Object == fanin {
			count++
			if index != nil {
				*index = int32(i)
			}
		}
	}
	return count
}

func Abc_NodeCollapseSuppSize(fanin, fanout *Object, fanins *vec.Vec_Ptr_t) int32 {
	vec.Vec_PtrClear(fanins)
	for _, edge := range fanout.fanins {
		if edge.Object != fanin {
			vec.Vec_PtrPushUnique(fanins, edge.Object)
		}
	}
	for _, edge := range fanin.fanins {
		vec.Vec_PtrPushUnique(fanins, edge.Object)
	}
	return int32(vec.Vec_PtrSize(fanins))
}

func Abc_ObjFaninNumberNew(fanins *vec.Vec_Ptr_t, object *Object) int32 {
	for i := 0; i < vec.Vec_PtrSize(fanins); i++ {
		if vec.Vec_PtrEntry(fanins, i) == object {
			return int32(i)
		}
	}
	return -1
}

func Abc_NodeCollapsePermMap(node, skip *Object, fanins *vec.Vec_Ptr_t, permutation []int32) int32 {
	if len(permutation) < max(len(node.fanins), vec.Vec_PtrSize(fanins)) {
		panic("permutation buffer too short")
	}
	for i := 0; i < vec.Vec_PtrSize(fanins); i++ {
		permutation[i] = int32(i)
	}
	for i, edge := range node.fanins {
		if edge.Object != skip {
			permutation[i] = Abc_ObjFaninNumberNew(fanins, edge.Object)
			if permutation[i] < 0 {
				return 0
			}
		}
	}
	return 1
}

// Abc_NodeCollapseFunc composes two local BDDs over their ordered union of inputs.
func Abc_NodeCollapseFunc(fanin, fanout *Object, fanins *vec.Vec_Ptr_t, faninMap, fanoutMap []int32) *core.DdNode {
	var position int32
	if Abc_NodeCheckDupFanin(fanin, fanout, &position) != 1 {
		return nil
	}
	count := int(Abc_NodeCollapseSuppSize(fanin, fanout, fanins))
	if Abc_NodeCollapsePermMap(fanin, nil, fanins, faninMap) == 0 || Abc_NodeCollapsePermMap(fanout, fanin, fanins, fanoutMap) == 0 {
		panic("collapse fanin mapping failed")
	}
	dd := fanin.network.FunctionManager.(*core.DdManager)
	if count > 0 {
		core.Cudd_bddIthVar(dd, count-1)
	}
	vector := append([]*core.DdNode(nil), dd.Vars...)
	for i, edge := range fanin.fanins {
		v := dd.Vars[faninMap[i]]
		if edge.Complement {
			v = core.Cudd_Not(v)
		}
		vector[i] = v
	}
	composed := core.Cudd_bddVectorCompose(dd, fanin.Data.(*core.DdNode), vector)
	if composed == nil {
		return nil
	}
	vector = append([]*core.DdNode(nil), dd.Vars...)
	for i, edge := range fanout.fanins {
		var v *core.DdNode
		if i == int(position) {
			v = composed
		} else {
			v = dd.Vars[fanoutMap[i]]
		}
		if edge.Complement {
			v = core.Cudd_Not(v)
		}
		vector[i] = v
	}
	return core.Cudd_bddVectorCompose(dd, fanout.Data.(*core.DdNode), vector)
}

func Abc_NodeCollapse(fanin, fanout *Object, fanins *vec.Vec_Ptr_t, faninMap, fanoutMap []int32) int32 {
	if fanin.network != fanout.network || Abc_NtkIsBddLogic(fanin.network) == 0 {
		panic("collapse requires nodes in one BDD logic network")
	}
	function := Abc_NodeCollapseFunc(fanin, fanout, fanins, faninMap, fanoutMap)
	if function == nil {
		return 0
	}
	core.Cudd_Ref(function)
	n := fanin.network
	replacement := n.NewObject(ObjectNode)
	for i := 0; i < vec.Vec_PtrSize(fanins); i++ {
		Abc_ObjAddFanin(replacement, vec.Vec_PtrEntry(fanins, i))
	}
	replacement.Data = function
	Abc_NodeMinimumBase(replacement)
	Abc_ObjTransferFanout(fanout, replacement)
	Abc_NtkDeleteObj_rec(fanout, 1)
	return 1
}

func Abc_NodeCountAppearances(fanin, fanout *Object) int32 {
	manager := fanin.network.FunctionManager.(*hop.Hop_Man_t)
	for i, edge := range fanout.fanins {
		if edge.Object == fanin {
			return hop.Hop_ObjFanoutCount(fanout.Data.(*hop.Hop_Obj_t), hop.Hop_IthVar(manager, int32(i)))
		}
	}
	panic("node is not a fanin")
}

func Abc_NodeCountAppearancesAll(node *Object) int32 {
	var count int32
	for _, fanout := range node.Fanouts() {
		count += Abc_NodeCountAppearances(node, fanout)
	}
	return count
}

// composeHop maps local input variables simultaneously, including their phases.
func composeHop(manager *hop.Hop_Man_t, root *hop.Hop_Obj_t, inputs []*hop.Hop_Obj_t) *hop.Hop_Obj_t {
	variables := make(map[*hop.Hop_Obj_t]*hop.Hop_Obj_t, len(inputs))
	for i, value := range inputs {
		variables[hop.Hop_IthVar(manager, int32(i))] = value
	}
	memo := make(map[*hop.Hop_Obj_t]*hop.Hop_Obj_t)
	var visit func(*hop.Hop_Obj_t) *hop.Hop_Obj_t
	visit = func(edge *hop.Hop_Obj_t) *hop.Hop_Obj_t {
		complemented := hop.Hop_IsComplement(edge) != 0
		object := hop.Hop_Regular(edge)
		value, ok := memo[object]
		if !ok {
			switch {
			case hop.Hop_ObjIsConst1(object) != 0:
				value = hop.Hop_ManConst1(manager)
			case hop.Hop_ObjIsPi(object) != 0:
				value = variables[object]
				if value == nil {
					panic("Hop input exceeds node fanins")
				}
			case hop.Hop_ObjIsNode(object) != 0:
				a, b := visit(hop.Hop_ObjFanin0(object)), visit(hop.Hop_ObjFanin1(object))
				if hop.Hop_ObjFaninC0(object) != 0 {
					a = hop.Hop_Not(a)
				}
				if hop.Hop_ObjFaninC1(object) != 0 {
					b = hop.Hop_Not(b)
				}
				if hop.Hop_ObjIsExor(object) != 0 {
					value = hop.Hop_Exor(manager, a, b)
				} else {
					value = hop.Hop_And(manager, a, b)
				}
			default:
				panic("unsupported Hop object")
			}
			memo[object] = value
		}
		if complemented {
			return hop.Hop_Not(value)
		}
		return value
	}
	return visit(root)
}

func Abc_NodeCollapseFunc1(fanin, fanout *Object, fanins *vec.Vec_Ptr_t, faninMap, fanoutMap []int32) *hop.Hop_Obj_t {
	var position int32
	if Abc_NodeCheckDupFanin(fanin, fanout, &position) != 1 {
		return nil
	}
	Abc_NodeCollapseSuppSize(fanin, fanout, fanins)
	if Abc_NodeCollapsePermMap(fanin, nil, fanins, faninMap) == 0 || Abc_NodeCollapsePermMap(fanout, fanin, fanins, fanoutMap) == 0 {
		panic("collapse fanin mapping failed")
	}
	manager := fanin.network.FunctionManager.(*hop.Hop_Man_t)
	values := make([]*hop.Hop_Obj_t, len(fanin.fanins))
	for i, edge := range fanin.fanins {
		values[i] = hop.Hop_IthVar(manager, faninMap[i])
		if edge.Complement {
			values[i] = hop.Hop_Not(values[i])
		}
	}
	internal := composeHop(manager, fanin.Data.(*hop.Hop_Obj_t), values)
	values = make([]*hop.Hop_Obj_t, len(fanout.fanins))
	for i, edge := range fanout.fanins {
		if i == int(position) {
			values[i] = internal
		} else {
			values[i] = hop.Hop_IthVar(manager, fanoutMap[i])
		}
		if edge.Complement {
			values[i] = hop.Hop_Not(values[i])
		}
	}
	return composeHop(manager, fanout.Data.(*hop.Hop_Obj_t), values)
}

func Abc_NodeCollapse1(fanin, fanout *Object, fanins *vec.Vec_Ptr_t, faninMap, fanoutMap []int32) int32 {
	if fanin.network != fanout.network || Abc_NtkIsAigLogic(fanin.network) == 0 {
		panic("collapse requires nodes in one Hop logic network")
	}
	function := Abc_NodeCollapseFunc1(fanin, fanout, fanins, faninMap, fanoutMap)
	if function == nil {
		return 0
	}
	replacement := fanin.network.NewObject(ObjectNode)
	for i := 0; i < vec.Vec_PtrSize(fanins); i++ {
		Abc_ObjAddFanin(replacement, vec.Vec_PtrEntry(fanins, i))
	}
	replacement.Data = function
	Abc_ObjTransferFanout(fanout, replacement)
	Abc_NtkDeleteObj_rec(fanout, 1)
	return 1
}

func Abc_NodeIsExor(node *Object) int32 {
	count := len(node.fanins)
	if count < 3 || count > 6 {
		return 0
	}
	manager := node.network.FunctionManager.(*hop.Hop_Man_t)
	truth := hop.Hop_ManComputeTruth6(manager, node.Data.(*hop.Hop_Obj_t), int32(count))
	switch truth {
	case 0x6666666666666666, 0x9999999999999999, 0x9696969696969696, 0x6969696969696969, 0x6996699669966996, 0x9669966996699669, 0x9669699696696996, 0x6996966969969669, 0x6996966996696996, 0x9669699669969669:
		return 1
	}
	return 0
}

func Abc_NtkEliminate1One(n *Network, limit, maxSize, reverse, verbose int32) int32 {
	if maxSize <= 0 || Abc_NtkIsLogic(n) == 0 {
		panic("elimination requires a logic network and positive size limit")
	}
	if Abc_NtkToAig(n) == 0 {
		fmt.Println("Converting to AIG has failed.")
		return 0
	}
	nodes := Abc_NtkDfs(n, 0)
	if reverse != 0 {
		nodes = Abc_NtkDfsReverse(n)
	}
	fanins := vec.Vec_PtrAlloc(int(maxSize))
	for _, node := range nodes {
		if !n.contains(node) || node.Kind != ObjectNode || len(node.fanins) > int(maxSize) || Abc_NodeIsExor(node) != 0 {
			continue
		}
		fanouts := node.Fanouts()
		skip := false
		for _, fanout := range fanouts {
			if fanout.IsCO() || fanout.Kind != ObjectNode {
				skip = true
				break
			}
		}
		if skip || Abc_NodeCountAppearancesAll(node) > limit+2 {
			continue
		}
		for _, fanout := range fanouts {
			if Abc_NodeCollapseSuppSize(node, fanout, fanins) > maxSize {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		for _, fanout := range fanouts {
			if verbose != 0 {
				fmt.Printf("Collapsing fanin %5d (supp =%2d) into fanout %5d (supp =%2d) ", node.ID, len(node.fanins), fanout.ID, len(fanout.fanins))
			}
			size := max(int(maxSize), max(len(node.fanins), len(fanout.fanins)))
			if Abc_NodeCollapse1(node, fanout, fanins, make([]int32, size), make([]int32, size)) == 0 {
				panic("fanin collapse failed")
			}
			if verbose != 0 {
				newNode := n.objects[len(n.objects)-1]
				if newNode != nil {
					fmt.Printf("resulting in node %5d (supp =%2d).\n", newNode.ID, len(newNode.fanins))
				}
			}
		}
	}
	return 1
}

func Abc_NtkEliminate1(n *Network, limit, maxSize, iterations, reverse, verbose int32) int32 {
	for i := int32(0); i < iterations; i++ {
		count := Abc_NtkNodeNum(n)
		if Abc_NtkEliminate1One(n, limit, maxSize, reverse, verbose) == 0 {
			return 0
		}
		if count == Abc_NtkNodeNum(n) {
			break
		}
	}
	return 1
}

// Abc_NodeBddReorder minimizes a local BDD and applies the resulting input order.
func Abc_NodeBddReorder(reorder *reo.Manager, node *Object) {
	dd := node.network.FunctionManager.(*core.DdManager)
	roots, order, err := reorder.Reorder(dd, []*core.DdNode{node.Data.(*core.DdNode)})
	if err != nil {
		panic(err)
	}
	oldInputs := append([]Edge(nil), node.fanins...)
	if len(order) < len(oldInputs) {
		panic("BDD reorder returned short input order")
	}
	seen := make([]bool, len(oldInputs))
	for i := range oldInputs {
		index := order[i]
		if index < 0 || index >= len(oldInputs) || seen[index] {
			panic("BDD reorder returned invalid input order")
		}
		seen[index] = true
		node.fanins[i] = oldInputs[index]
	}
	old := node.Data.(*core.DdNode)
	core.Cudd_Ref(roots[0])
	node.Data = roots[0]
	core.Cudd_RecursiveDeref(dd, old)
}

func Abc_NtkBddReorder(n *Network, verbose int32) {
	Abc_NtkRemoveDupFanins(n)
	Abc_NtkMinimumBase(n)
	maxFanins := 0
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		maxFanins = max(maxFanins, len(node.fanins))
	}
	reorder := reo.Extra_ReorderInit(maxFanins, 100)
	defer reo.Extra_ReorderQuit(reorder)
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if len(node.fanins) < 3 {
			continue
		}
		if verbose != 0 {
			fmt.Printf("%10s: ", node.Name())
		}
		Abc_NodeBddReorder(reorder, node)
	}
}

func Abc_NtkEliminate(n *Network, maxSize, reverse, verbose int32) int32 {
	if maxSize <= 0 || Abc_NtkIsLogic(n) == 0 {
		panic("elimination requires a logic network and positive size limit")
	}
	if Abc_NtkToBdd(n) == 0 {
		fmt.Println("Converting to BDD has failed.")
		return 0
	}
	Abc_NtkRemoveDupFanins(n)
	Abc_NtkMinimumBase(n)
	reachable := make(map[*Object]bool)
	for _, node := range Abc_NtkDfs(n, 0) {
		reachable[node] = true
	}
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if !reachable[node] {
			Abc_NtkDeleteObj(node)
		}
	}
	nodes := Abc_NtkDfs(n, 0)
	if reverse != 0 {
		nodes = Abc_NtkDfsReverse(n)
	}
	fanins := vec.Vec_PtrAlloc(int(maxSize))
	for _, node := range nodes {
		if !n.contains(node) || node.Kind != ObjectNode || len(node.fanins) > int(maxSize) {
			continue
		}
		fanouts := node.Fanouts()
		skip := false
		for _, fanout := range fanouts {
			if fanout.IsCO() || fanout.Kind != ObjectNode || Abc_NodeCollapseSuppSize(node, fanout, fanins) > maxSize {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		for _, fanout := range fanouts {
			if verbose != 0 {
				fmt.Printf("Collapsing fanin %5d (supp =%2d) into fanout %5d (supp =%2d) ", node.ID, len(node.fanins), fanout.ID, len(fanout.fanins))
			}
			size := max(int(maxSize)+1, max(len(node.fanins), len(fanout.fanins)))
			if Abc_NodeCollapse(node, fanout, fanins, make([]int32, size), make([]int32, size)) == 0 {
				panic("BDD fanin collapse failed")
			}
			if verbose != 0 {
				last := n.objects[len(n.objects)-1]
				if last != nil {
					fmt.Printf("resulting in node %5d (supp =%2d).\n", last.ID, len(last.fanins))
				}
			}
		}
	}
	Abc_NtkBddReorder(n, 0)
	return 1
}

// Abc_NodeSupport records BDD support without borrowing the node's mark bits.
func Abc_NodeSupport(root *core.DdNode, support *vec.Vec_Str_t, variables int32) int32 {
	if root == nil || support == nil || variables < 0 {
		panic("invalid BDD support arguments")
	}
	vec.Vec_StrFill(support, int(variables), 0)
	seen := make(map[*core.DdNode]bool)
	var count int32
	var visit func(*core.DdNode)
	visit = func(edge *core.DdNode) {
		if edge == nil || core.Cudd_IsConstant(edge) {
			return
		}
		node := core.Cudd_Regular(edge)
		if seen[node] {
			return
		}
		seen[node] = true
		if node.Index >= uint32(variables) {
			panic("BDD support exceeds declared variables")
		}
		index := int(node.Index)
		if vec.Vec_StrEntry(support, index) == 0 {
			vec.Vec_StrWriteEntry(support, index, 1)
			count++
		}
		visit(node.T)
		visit(node.E)
	}
	visit(root)
	return count
}

func Abc_ObjCompareByNumber(first, second *any) int {
	a, b := (*first).(*Object), (*second).(*Object)
	if a.Temp < b.Temp {
		return -1
	}
	if a.Temp > b.Temp {
		return 1
	}
	return 0
}

func Abc_ObjSortInReverseOrder(n *Network, nodes *vec.Vec_Ptr_t) {
	order := Abc_NtkDfsReverse(n)
	for i, node := range order {
		node.Temp = i
	}
	vec.Vec_PtrSort(nodes, Abc_ObjCompareByNumber)
	for _, node := range order {
		node.Temp = 0
	}
}

func Abc_NtkEliminateSpecial(n *Network, maxSize, verbose int32) int32 {
	if maxSize <= 0 || Abc_NtkIsLogic(n) == 0 {
		panic("elimination requires a logic network and positive size limit")
	}
	if Abc_NtkToBdd(n) == 0 {
		fmt.Println("Converting to BDD has failed.")
		return 0
	}
	Abc_NtkRemoveDupFanins(n)
	Abc_NtkMinimumBase(n)
	reachable := make(map[*Object]bool)
	for _, node := range Abc_NtkDfs(n, 0) {
		reachable[node] = true
	}
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if !reachable[node] {
			Abc_NtkDeleteObj(node)
		}
	}
	if Abc_NtkToSop(n, -1, math.MaxInt32) == 0 {
		fmt.Println("Converting to SOP has failed.")
		return 0
	}
	nodes := vec.Vec_PtrAlloc(0)
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if len(node.fanouts) != 1 {
			continue
		}
		fanout := node.fanouts[0]
		if fanout.Kind != ObjectNode {
			continue
		}
		inSop, ok := node.Data.(string)
		if !ok {
			panic("expected SOP node")
		}
		outSop, ok := fanout.Data.(string)
		if !ok {
			panic("expected SOP fanout")
		}
		if Abc_SopGetCubeNum(inSop) != 1 || Abc_SopGetCubeNum(outSop) != 1 {
			continue
		}
		index := Abc_ObjFanoutFaninNum(fanout, node)
		if index < 0 {
			panic("broken fanout reference")
		}
		if Abc_SopIsComplement(inSop) == Abc_SopGetIthCareLit(outSop, index) {
			continue
		}
		vec.Vec_PtrPush(nodes, node)
	}
	if vec.Vec_PtrSize(nodes) == 0 {
		return 1
	}
	Abc_ObjSortInReverseOrder(n, nodes)
	if Abc_NtkToBdd(n) == 0 {
		fmt.Println("Converting to BDD has failed.")
		return 0
	}
	fanins := vec.Vec_PtrAlloc(int(maxSize))
	for i := 0; i < vec.Vec_PtrSize(nodes); i++ {
		node := vec.Vec_PtrEntry(nodes, i).(*Object)
		if !n.contains(node) || node.Kind != ObjectNode {
			panic("selected elimination node was removed")
		}
		for _, fanout := range node.Fanouts() {
			if fanout.IsCO() {
				panic("selected elimination node drives an output")
			}
			if verbose != 0 {
				fmt.Printf("Collapsing fanin %5d (supp =%2d) into fanout %5d (supp =%2d) ", node.ID, len(node.fanins), fanout.ID, len(fanout.fanins))
			}
			size := max(int(maxSize)+1, len(node.fanins)+len(fanout.fanins))
			if Abc_NodeCollapse(node, fanout, fanins, make([]int32, size), make([]int32, size)) == 0 {
				panic("BDD fanin collapse failed")
			}
			if verbose != 0 {
				last := n.objects[len(n.objects)-1]
				if last != nil {
					fmt.Printf("resulting in node %5d (supp =%2d).\n", last.ID, len(last.fanins))
				}
			}
		}
	}
	Abc_NtkBddReorder(n, 0)
	return 1
}
