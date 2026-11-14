package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/dsd"
)

func abcTransferBDD(source, destination *core.DdManager, function *core.DdNode) *core.DdNode {
	memo := map[*core.DdNode]*core.DdNode{}
	var transfer func(*core.DdNode) *core.DdNode
	transfer = func(edge *core.DdNode) *core.DdNode {
		if result := memo[edge]; result != nil {
			return result
		}
		regular := core.Cudd_Regular(edge)
		if core.Cudd_IsConstant(regular) {
			result := core.Cudd_NotCond(core.Cudd_ReadOne(destination), edge != regular)
			memo[edge] = result
			return result
		}
		low := core.Cudd_NotCond(regular.E, edge != regular)
		high := core.Cudd_NotCond(regular.T, edge != regular)
		level := source.Perm[regular.Index]
		index := destination.InvPerm[level]
		result := core.Cudd_bddIte(destination, core.Cudd_bddIthVar(destination, index), transfer(high), transfer(low))
		memo[edge] = result
		return result
	}
	return transfer(function)
}

// Abc_NtkDsdGlobal derives global BDDs for a strashed network and replaces
// each output cone by its disjoint-support decomposition.
func Abc_NtkDsdGlobal(n *abc.Network, verbose, printTree, short int32) *abc.Network {
	if n == nil || abc.Abc_NtkIsStrash(n) == 0 {
		return nil
	}
	dd := core.NewBDDManager(max(1, len(n.Inputs())))
	functions := make([]*core.DdNode, 0, len(n.Outputs()))
	for _, output := range n.Outputs() {
		fanins := output.Fanins()
		if len(fanins) != 1 {
			core.Cudd_Quit(dd)
			return nil
		}
		truth, err := abcNetworkBDD(n, dd, fanins[0])
		if err != nil {
			core.Cudd_Quit(dd)
			return nil
		}
		functions = append(functions, truth)
	}
	if verbose != 0 {
		fmt.Printf("Shared BDD size = %6d nodes.\n", core.Cudd_ReadKeys(dd))
	}
	result, err := abcDsdBuildNetwork(n, dd, functions, printTree != 0, short != 0, verbose != 0)
	core.Cudd_Quit(dd)
	if err == nil && result != nil && n.Exdc != nil {
		result.Exdc = abc.Abc_NtkDup(n.Exdc)
	}
	if err != nil || result == nil || abc.Abc_NtkCheck(result) == 0 {
		if result != nil {
			abc.Abc_NtkDelete(result)
		}
		return nil
	}
	return result
}

func abcNetworkBDD(n *abc.Network, dd *core.DdManager, root abc.Edge) (*core.DdNode, error) {
	inputs := n.Inputs()
	for i := range inputs {
		core.Cudd_bddIthVar(dd, i)
	}
	indices := make(map[*abc.Object]int, len(inputs))
	for i, input := range inputs {
		indices[input] = i
	}
	memo := make(map[*abc.Object]*core.DdNode)
	var convert func(*abc.Object) (*core.DdNode, error)
	convert = func(object *abc.Object) (*core.DdNode, error) {
		if f := memo[object]; f != nil {
			return f, nil
		}
		if object.Kind == abc.ObjectConst1 {
			return core.Cudd_ReadOne(dd), nil
		}
		if index, ok := indices[object]; ok {
			return core.Cudd_bddIthVar(dd, index), nil
		}
		if object.Data == "abcgo:AIG AND" && len(object.Fanins()) == 2 {
			left, err := convert(object.Fanins()[0].Object)
			if err != nil {
				return nil, err
			}
			left = core.Cudd_NotCond(left, object.Fanins()[0].Complement)
			right, err := convert(object.Fanins()[1].Object)
			if err != nil {
				return nil, err
			}
			right = core.Cudd_NotCond(right, object.Fanins()[1].Complement)
			f := core.Cudd_bddAnd(dd, left, right)
			memo[object] = f
			return f, nil
		}
		return nil, fmt.Errorf("unsupported strashed node %d", object.ID)
	}
	if f, err := convert(root.Object); err == nil {
		return core.Cudd_NotCond(f, root.Complement), nil
	}
	if len(inputs) > 20 {
		return nil, fmt.Errorf("unsupported strashed root %d with %d inputs", root.Object.ID, len(inputs))
	}
	values := make(map[*abc.Object]bool, len(inputs))
	// Network.Evaluate may report malformed graph errors. Probe each row before
	// recursion so these errors remain ordinary API results rather than panics.
	for mask := 0; mask < 1<<len(inputs); mask++ {
		for i, input := range inputs {
			values[input] = mask&(1<<i) != 0
		}
		if _, err := n.Evaluate(root, values); err != nil {
			return nil, err
		}
	}
	var failed error
	var safeBuild func(int, int) *core.DdNode
	safeBuild = func(index, assignment int) *core.DdNode {
		if failed != nil {
			return core.Cudd_Not(core.Cudd_ReadOne(dd))
		}
		if index == len(inputs) {
			for i, input := range inputs {
				values[input] = assignment&(1<<i) != 0
			}
			v, err := n.Evaluate(root, values)
			if err != nil {
				failed = err
				return core.Cudd_Not(core.Cudd_ReadOne(dd))
			}
			if v {
				return core.Cudd_ReadOne(dd)
			}
			return core.Cudd_Not(core.Cudd_ReadOne(dd))
		}
		lo := safeBuild(index+1, assignment)
		hi := safeBuild(index+1, assignment|(1<<index))
		return core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, index), hi, lo)
	}
	f := safeBuild(0, 0)
	return f, failed
}

func abcDsdBuildNetwork(src *abc.Network, dd *core.DdManager, functions []*core.DdNode, printTree, short, verbose bool) (*abc.Network, error) {
	manager := dsd.Dsd_ManagerStart(dd, len(src.Inputs()), 0)
	if manager == nil {
		return nil, fmt.Errorf("DSD manager allocation failed")
	}
	defer dsd.Dsd_ManagerStop(manager)
	dsd.Dsd_Decompose(manager, functions, len(functions))
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	newDD := out.FunctionManager.(*core.DdManager)
	for i := range src.Inputs() {
		core.Cudd_bddIthVar(newDD, i)
	}
	copies := map[*dsd.Dsd_Node_t]*abc.Object{}
	const1 := abc.Abc_NtkCreateNodeConst1(out)
	copies[dsd.Dsd_Regular(dsd.Dsd_ManagerReadConst1(manager))] = const1
	inputs := src.Inputs()
	for i, old := range inputs {
		pi := out.NewObject(abc.ObjectPI)
		pi.SetName(old.Name())
		copies[dsd.Dsd_Regular(dsd.Dsd_ManagerReadInput(manager, i))] = pi
	}
	var construct func(*dsd.Dsd_Node_t) (*abc.Object, error)
	construct = func(edge *dsd.Dsd_Node_t) (*abc.Object, error) {
		regular := dsd.Dsd_Regular(edge)
		if got := copies[regular]; got != nil {
			return got, nil
		}
		typ := dsd.Dsd_NodeReadType(regular)
		decs := dsd.Dsd_NodeReadDecs(regular)
		if len(decs) < 2 {
			return nil, fmt.Errorf("invalid DSD node with %d decompositions", len(decs))
		}
		node := out.NewObject(abc.ObjectNode)
		for _, child := range decs {
			fanin, err := construct(child)
			if err != nil {
				return nil, err
			}
			if err = out.AddFanin(node, abc.Edge{Object: fanin, Complement: false}); err != nil {
				return nil, err
			}
		}
		var local *core.DdNode
		switch typ {
		case dsd.DSD_NODE_OR:
			local = core.Cudd_Not(core.Cudd_ReadOne(newDD))
			for i, child := range decs {
				v := core.Cudd_NotCond(core.Cudd_bddIthVar(newDD, i), dsd.Dsd_IsComplement(child))
				local = core.Cudd_bddOr(newDD, local, v)
			}
		case dsd.DSD_NODE_EXOR:
			local = core.Cudd_Not(core.Cudd_ReadOne(newDD))
			for i := range decs {
				local = core.Cudd_bddXor(newDD, local, core.Cudd_bddIthVar(newDD, i))
			}
		case dsd.DSD_NODE_PRIME:
			local = dsd.Dsd_TreeGetPrimeFunction(dd, regular)
			local = abcTransferBDD(dd, newDD, local)
		default:
			return nil, fmt.Errorf("unsupported DSD node kind %d", typ)
		}
		node.Data = local
		copies[regular] = node
		return node, nil
	}
	roots := make([]*dsd.Dsd_Node_t, len(functions))
	for i := range roots {
		roots[i] = dsd.Dsd_ManagerReadRoot(manager, i)
		if _, err := construct(roots[i]); err != nil {
			return nil, err
		}
	}
	for i, root := range roots {
		driver, err := construct(root)
		if err != nil {
			return nil, err
		}
		po := out.NewObject(abc.ObjectPO)
		po.SetName(src.Outputs()[i].Name())
		if err = out.AddFanin(po, abc.Edge{Object: driver, Complement: dsd.Dsd_IsComplement(root)}); err != nil {
			return nil, err
		}
	}
	if printTree {
		namesIn, namesOut := abcNetworkIONames(src)
		if verbose {
			dsd.Dsd_TreePrint(os.Stdout, manager, namesIn, namesOut, int(boolInt(short)), -1)
		} else {
			dsd.Dsd_TreePrint2(os.Stdout, manager, namesIn, namesOut, -1)
		}
	}
	return out, nil
}

func abcNetworkIONames(n *abc.Network) ([]string, []string) {
	in, out := n.Inputs(), n.Outputs()
	a, b := make([]string, len(in)), make([]string, len(out))
	for i, o := range in {
		a[i] = o.Name()
	}
	for i, o := range out {
		b[i] = o.Name()
	}
	return a, b
}

// Abc_NtkDsdConstructNode creates one internal decomposition node. Child
// marks contain destination object IDs, matching the C mark-to-pointer map.
func Abc_NtkDsdConstructNode(manager *dsd.Dsd_Manager_t, dsdNode *dsd.Dsd_Node_t, dst *abc.Network, counters []int) (*abc.Object, error) {
	if manager == nil || dsdNode == nil || dst == nil {
		return nil, fmt.Errorf("nil DSD construction argument")
	}
	typ, decs := dsd.Dsd_NodeReadType(dsdNode), dsd.Dsd_NodeReadDecs(dsdNode)
	if len(decs) < 2 {
		return nil, fmt.Errorf("DSD internal node has %d decompositions", len(decs))
	}
	ddDst, ok := dst.FunctionManager.(*core.DdManager)
	if !ok {
		return nil, fmt.Errorf("destination is not a BDD network")
	}
	node := dst.NewObject(abc.ObjectNode)
	for _, child := range decs {
		regular := dsd.Dsd_Regular(child)
		fanin := dst.Object(dsd.Dsd_NodeReadMark(regular))
		if fanin == nil {
			return nil, fmt.Errorf("DSD child has no destination mark")
		}
		if err := dst.AddFanin(node, abc.Edge{Object: fanin}); err != nil {
			return nil, err
		}
	}
	var local *core.DdNode
	switch typ {
	case dsd.DSD_NODE_CONST1:
		local = core.Cudd_ReadOne(ddDst)
	case dsd.DSD_NODE_OR:
		local = core.Cudd_Not(core.Cudd_ReadOne(ddDst))
		for i, child := range decs {
			variable := core.Cudd_NotCond(core.Cudd_bddIthVar(ddDst, i), dsd.Dsd_IsComplement(child))
			local = core.Cudd_bddOr(ddDst, local, variable)
		}
	case dsd.DSD_NODE_EXOR:
		local = core.Cudd_Not(core.Cudd_ReadOne(ddDst))
		for i := range decs {
			local = core.Cudd_bddXor(ddDst, local, core.Cudd_bddIthVar(ddDst, i))
		}
	case dsd.DSD_NODE_PRIME:
		if len(counters) > 0 {
			i := min(len(decs), 10)
			if i < len(counters) {
				counters[i]++
			}
		}
		local = dsd.Dsd_TreeGetPrimeFunction(dsd.Dsd_ManagerReadDd(manager), dsd.Dsd_Regular(dsdNode))
		local = abcTransferBDD(dsd.Dsd_ManagerReadDd(manager), ddDst, local)
	default:
		return nil, fmt.Errorf("unsupported DSD node type %d", typ)
	}
	node.Data = local
	dsd.Dsd_NodeSetMark(dsd.Dsd_Regular(dsdNode), node.ID)
	return node, nil
}

// Abc_NtkDsdConstruct maps every DSD root into a destination BDD network.
func Abc_NtkDsdConstruct(manager *dsd.Dsd_Manager_t, src, dst *abc.Network) error {
	if manager == nil || src == nil || dst == nil {
		return fmt.Errorf("nil DSD network construction argument")
	}
	dsd.Dsd_NodeSetMark(dsd.Dsd_ManagerReadConst1(manager), dst.One().Object.ID)
	if len(src.Inputs()) != len(dst.Inputs()) {
		return fmt.Errorf("source and destination CI counts differ")
	}
	for i := range src.Inputs() {
		dsd.Dsd_NodeSetMark(dsd.Dsd_ManagerReadInput(manager, i), dst.Inputs()[i].ID)
	}
	count := 0
	nodes := dsd.Dsd_TreeCollectNodesDfs(manager, &count)
	for _, node := range nodes {
		if _, err := Abc_NtkDsdConstructNode(manager, node, dst, nil); err != nil {
			return err
		}
	}
	for i, co := range src.Outputs() {
		if len(co.Fanins()) != 1 {
			return fmt.Errorf("output %d does not have one driver", i)
		}
		root := dsd.Dsd_ManagerReadRoot(manager, i)
		driver := dst.Object(dsd.Dsd_NodeReadMark(dsd.Dsd_Regular(root)))
		if driver == nil {
			return fmt.Errorf("DSD root %d has no destination node", i)
		}
		po := dst.NewObject(abc.ObjectPO)
		po.SetName(co.Name())
		if err := dst.AddFanin(po, abc.Edge{Object: driver, Complement: dsd.Dsd_IsComplement(root)}); err != nil {
			return err
		}
	}
	return nil
}

func Extra_bddComputeSum(dd *core.DdManager, cubes []*core.DdNode) *core.DdNode {
	result := core.Cudd_Not(core.Cudd_ReadOne(dd))
	for _, cube := range cubes {
		result = core.Cudd_bddOr(dd, result, cube)
	}
	return result
}

// Abc_NtkDsdInternal is retained for callers that already own global BDDs.
func Abc_NtkDsdInternal(n *abc.Network, verbose, printTree, short int32) *abc.Network {
	if n == nil {
		return nil
	}
	dd, ok := n.FunctionManager.(*core.DdManager)
	if !ok {
		return nil
	}
	functions := make([]*core.DdNode, 0, len(n.Outputs()))
	for _, co := range n.Outputs() {
		f, ok := co.Fanins()[0].Object.Data.(*core.DdNode)
		if !ok {
			return nil
		}
		functions = append(functions, core.Cudd_NotCond(f, co.Fanins()[0].Complement))
	}
	out, err := abcDsdBuildNetwork(n, dd, functions, printTree != 0, short != 0, verbose != 0)
	if err != nil {
		return nil
	}
	return out
}

// Abc_NtkDsdLocal decomposes BDD nodes whose support exceeds two variables.
func Abc_NtkDsdLocal(n *abc.Network, verbose, recursive int32) int32 {
	if n == nil || abc.Abc_NtkIsBddLogic(n) == 0 {
		return 0
	}
	dd, ok := n.FunctionManager.(*core.DdManager)
	if !ok {
		return 0
	}
	manager := dsd.Dsd_ManagerStart(dd, len(dd.Vars), 0)
	if manager == nil {
		return 0
	}
	defer dsd.Dsd_ManagerStop(manager)
	nodes := Abc_NtkCollectNodesForDsd(n)
	counters := make([]int, 11)
	for i := 0; i < len(nodes); i++ {
		Abc_NodeDecompDsdAndMux(nodes[i], &nodes, manager, recursive != 0, counters)
	}
	if verbose != 0 {
		fmt.Println("Number of non-decomposable functions:")
		for i := 3; i < 10; i++ {
			fmt.Printf("Inputs = %d.  Functions = %6d.\n", i, counters[i])
		}
		fmt.Printf("Inputs > 9.  Functions = %6d.\n", counters[10])
	}
	if abc.Abc_NtkCheck(n) != 0 {
		return 1
	}
	return 0
}

// Abc_NtkCollectNodesForDsd returns nodes with more than two fanins, in network order.
func Abc_NtkCollectNodesForDsd(n *abc.Network) []*abc.Object {
	var nodes []*abc.Object
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if Abc_NodeIsForDsd(node) != 0 {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func Abc_NodeIsForDsd(node *abc.Object) int32 {
	if node != nil && node.Kind == abc.ObjectNode && len(node.Fanins()) > 2 {
		return 1
	}
	return 0
}

// Abc_NodeFindMuxVar uses the C routine's minimum sum of cofactor supports;
// strict comparison preserves the first variable on ties.
func Abc_NodeFindMuxVar(dd *core.DdManager, function *core.DdNode, nvars int) int {
	best, bestCost := -1, int(^uint(0)>>1)
	for i := 0; i < nvars; i++ {
		variable := core.Cudd_bddIthVar(dd, i)
		low := core.Cudd_Cofactor(dd, function, core.Cudd_Not(variable))
		high := core.Cudd_Cofactor(dd, function, variable)
		cost := core.Cudd_SupportSize(dd, low) + core.Cudd_SupportSize(dd, high)
		if cost < bestCost {
			best, bestCost = i, cost
		}
	}
	return best
}

// Abc_NodeDecompDsdAndMux applies one source-style DSD or recursive mux step.
func Abc_NodeDecompDsdAndMux(node *abc.Object, pending *[]*abc.Object, manager *dsd.Dsd_Manager_t, recursive bool, counters []int) {
	if node == nil || manager == nil {
		return
	}
	dd := dsd.Dsd_ManagerReadDd(manager)
	f, ok := node.Data.(*core.DdNode)
	if !ok {
		return
	}
	root := dsd.Dsd_DecomposeOne(manager, f)
	regular := dsd.Dsd_Regular(root)
	fanins := node.Fanins()
	if recursive && dsd.Dsd_NodeReadDecsNum(regular) == len(fanins) {
		i := Abc_NodeFindMuxVar(dd, f, len(fanins))
		if i < 0 || i >= len(fanins) {
			return
		}
		control := fanins[i]
		low := abc.Abc_NtkCloneObj(node)
		high := abc.Abc_NtkCloneObj(node)
		for _, child := range []*abc.Object{low, high} {
			for j := len(child.Fanins()) - 1; j >= 0; j-- {
				_ = node.Network().DeleteFanin(child, j)
			}
			for _, edge := range fanins {
				_ = node.Network().AddFanin(child, edge)
			}
		}
		low.Data = core.Cudd_Cofactor(dd, f, core.Cudd_Not(core.Cudd_bddIthVar(dd, i)))
		core.Cudd_Ref(low.Data.(*core.DdNode))
		abc.Abc_NodeMinimumBase(low)
		high.Data = core.Cudd_Cofactor(dd, f, core.Cudd_bddIthVar(dd, i))
		core.Cudd_Ref(high.Data.(*core.DdNode))
		abc.Abc_NodeMinimumBase(high)
		for j := len(fanins) - 1; j >= 0; j-- {
			_ = node.Network().DeleteFanin(node, j)
		}
		_ = node.Network().AddFanin(node, control)
		_ = node.Network().AddFanin(node, abc.Edge{Object: high})
		_ = node.Network().AddFanin(node, abc.Edge{Object: low})
		node.Data = core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1), core.Cudd_bddIthVar(dd, 2))
		if pending != nil {
			if Abc_NodeIsForDsd(low) != 0 {
				*pending = append(*pending, low)
			}
			if Abc_NodeIsForDsd(high) != 0 {
				*pending = append(*pending, high)
			}
		}
		return
	}
	if dsd.Dsd_NodeReadDecsNum(regular) <= 1 {
		return
	}
	startID := len(node.Network().Objects())
	_, err := abcDsdAppendLocal(node.Network(), dd, manager, node, root, false)
	if err != nil {
		return
	}
	for _, created := range node.Network().ObjectsOfKind(abc.ObjectNode) {
		if created != node && created.ID >= startID && Abc_NodeIsForDsd(created) != 0 && recursive && pending != nil {
			*pending = append(*pending, created)
		}
	}
	seen := map[*dsd.Dsd_Node_t]bool{}
	var countPrimes func(*dsd.Dsd_Node_t)
	countPrimes = func(p *dsd.Dsd_Node_t) {
		p = dsd.Dsd_Regular(p)
		if p == nil || seen[p] {
			return
		}
		seen[p] = true
		if dsd.Dsd_NodeReadType(p) == dsd.DSD_NODE_PRIME {
			i := min(dsd.Dsd_NodeReadDecsNum(p), 10)
			if i < len(counters) {
				counters[i]++
			}
		}
		for _, child := range dsd.Dsd_NodeReadDecs(p) {
			countPrimes(child)
		}
	}
	countPrimes(root)
}

func abcDsdAppendLocal(n *abc.Network, dd *core.DdManager, manager *dsd.Dsd_Manager_t, node *abc.Object, root *dsd.Dsd_Node_t, recursive bool) (*abc.Object, error) {
	oldFanins := node.Fanins()
	copies := make(map[*dsd.Dsd_Node_t]*abc.Object)
	for i, edge := range oldFanins {
		copies[dsd.Dsd_Regular(dsd.Dsd_ManagerReadInput(manager, i))] = edge.Object
	}
	var makeNode func(*dsd.Dsd_Node_t) (*abc.Object, error)
	makeNode = func(edge *dsd.Dsd_Node_t) (*abc.Object, error) {
		regular := dsd.Dsd_Regular(edge)
		if got := copies[regular]; got != nil {
			return got, nil
		}
		typ := dsd.Dsd_NodeReadType(regular)
		children := dsd.Dsd_NodeReadDecs(regular)
		out := n.NewObject(abc.ObjectNode)
		for _, child := range children {
			fanin, err := makeNode(child)
			if err != nil {
				return nil, err
			}
			if err = n.AddFanin(out, abc.Edge{Object: fanin, Complement: false}); err != nil {
				return nil, err
			}
		}
		var f *core.DdNode
		switch typ {
		case dsd.DSD_NODE_CONST1:
			f = core.Cudd_ReadOne(dd)
		case dsd.DSD_NODE_BUF:
			f = core.Cudd_bddIthVar(dd, 0)
		case dsd.DSD_NODE_OR:
			f = core.Cudd_Not(core.Cudd_ReadOne(dd))
			for i, child := range children {
				f = core.Cudd_bddOr(dd, f, core.Cudd_NotCond(core.Cudd_bddIthVar(dd, i), dsd.Dsd_IsComplement(child)))
			}
		case dsd.DSD_NODE_EXOR:
			f = core.Cudd_Not(core.Cudd_ReadOne(dd))
			for i := range children {
				f = core.Cudd_bddXor(dd, f, core.Cudd_bddIthVar(dd, i))
			}
		case dsd.DSD_NODE_PRIME:
			f = dsd.Dsd_TreeGetPrimeFunction(dd, regular)
		default:
			return nil, fmt.Errorf("unsupported local DSD node type %d", typ)
		}
		out.Data = f
		copies[regular] = out
		_ = recursive
		return out, nil
	}
	driver, err := makeNode(dsd.Dsd_Regular(root))
	if err != nil {
		return nil, err
	}
	for i := len(oldFanins) - 1; i >= 0; i-- {
		if err := n.DeleteFanin(node, i); err != nil {
			return nil, err
		}
	}
	if err := n.AddFanin(node, abc.Edge{Object: driver}); err != nil {
		return nil, err
	}
	base := core.Cudd_bddIthVar(dd, 0)
	node.Data = core.Cudd_NotCond(base, dsd.Dsd_IsComplement(root))
	return driver, nil
}

// Abc_NtkSparsify exposes the source API. Percentages outside [1,100] are
// retain one minterm for zero-sized on/off sets, matching ABC behavior.
func Abc_NtkSparsify(n *abc.Network, percentage, verbose int32) *abc.Network {
	if n == nil || abc.Abc_NtkIsComb(n) == 0 || abc.Abc_NtkIsBddLogic(n) == 0 {
		return nil
	}
	return Abc_NtkSparsifyInternal(n, percentage, verbose)
}

func Abc_NtkSparsifyInternal(n *abc.Network, percentage, verbose int32) *abc.Network {
	if n == nil {
		return nil
	}
	sourceDD, ok := n.FunctionManager.(*core.DdManager)
	if !ok {
		return nil
	}
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	out.Name, out.Spec = n.Name, n.Spec
	destDD := out.FunctionManager.(*core.DdManager)
	oldToNew := map[*abc.Object]*abc.Object{}
	oldToNew[n.One().Object] = out.One().Object
	for _, old := range n.Inputs() {
		pi := out.NewObject(abc.ObjectPI)
		pi.SetName(old.Name())
		oldToNew[old] = pi
	}
	var roots []*abc.Object
	for _, co := range n.Outputs() {
		for _, edge := range co.Fanins() {
			roots = append(roots, edge.Object)
		}
	}
	order, err := n.Topological(roots)
	if err != nil {
		return nil
	}
	for _, old := range order {
		if old.IsCI() || old.Kind == abc.ObjectConst1 {
			continue
		}
		if old.Kind != abc.ObjectNode {
			continue
		}
		copy := out.NewObject(abc.ObjectNode)
		for _, edge := range old.Fanins() {
			fanin := oldToNew[edge.Object]
			if fanin == nil {
				return nil
			}
			if err := out.AddFanin(copy, abc.Edge{Object: fanin, Complement: edge.Complement}); err != nil {
				return nil
			}
		}
		function, ok := old.Data.(*core.DdNode)
		if !ok {
			return nil
		}
		copy.Data = abcTransferBDD(sourceDD, destDD, function)
		oldToNew[old] = copy
	}
	for _, co := range n.Outputs() {
		if len(co.Fanins()) != 1 {
			return nil
		}
		edge := co.Fanins()[0]
		name := co.Name()
		if edge.Object.IsCI() {
			driver := oldToNew[edge.Object]
			for c := 0; c < 2; c++ {
				po := out.NewObject(abc.ObjectPO)
				suffix := "_on"
				if c == 1 {
					suffix = "_off"
				}
				po.SetName(name + suffix)
				if err := out.AddFanin(po, abc.Edge{Object: driver, Complement: edge.Complement != (c == 1)}); err != nil {
					return nil
				}
			}
			continue
		}
		if edge.Object.Kind == abc.ObjectConst1 || len(edge.Object.Fanins()) == 0 {
			value, evalErr := n.Evaluate(edge, map[*abc.Object]bool{})
			if evalErr != nil {
				return nil
			}
			for c := 0; c < 2; c++ {
				complement := !(value != (c == 1))
				po := out.NewObject(abc.ObjectPO)
				suffix := "_on"
				if c == 1 {
					suffix = "_off"
				}
				po.SetName(name + suffix)
				if err := out.AddFanin(po, abc.Edge{Object: out.One().Object, Complement: complement}); err != nil {
					return nil
				}
			}
			continue
		}
		base := oldToNew[edge.Object]
		if base == nil {
			return nil
		}
		original, ok := base.Data.(*core.DdNode)
		if !ok {
			return nil
		}
		original = core.Cudd_NotCond(original, edge.Complement)
		for c := 0; c < 2; c++ {
			core.Cudd_Srandom(0)
			copy := out.NewObject(abc.ObjectNode)
			for _, fanin := range base.Fanins() {
				if err := out.AddFanin(copy, fanin); err != nil {
					return nil
				}
			}
			function := core.Cudd_NotCond(original, c == 1)
			copy.Data = Abc_NtkSparsifyInternalOne(destDD, function, len(copy.Fanins()), percentage)
			if copy.Data == nil {
				return nil
			}
			po := out.NewObject(abc.ObjectPO)
			suffix := "_on"
			if c == 1 {
				suffix = "_off"
			}
			po.SetName(name + suffix)
			if err := out.AddFanin(po, abc.Edge{Object: copy}); err != nil {
				return nil
			}
		}
	}
	_ = verbose
	if abc.Abc_NtkCheck(out) == 0 {
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

func Abc_NtkSparsifyInternalOne(dd *core.DdManager, function *core.DdNode, fanins int, percentage int32) *core.DdNode {
	space := int(core.Cudd_CountMinterm(dd, function, fanins))
	count := max(1, int(float64(percentage)*0.01*float64(space)))
	cubes := abcPickSourceMinterms(dd, function, fanins, count)
	if cubes == nil {
		return nil
	}
	result := core.Cudd_Not(core.Cudd_ReadOne(dd))
	for _, cube := range cubes {
		result = core.Cudd_bddOr(dd, result, cube)
	}
	return result
}

// abcPickSourceMinterms follows CUDD's recursive cofactor allocation and
// source ordering, then randomizes don't-cares with CUDD's seeded generator.
func abcPickSourceMinterms(dd *core.DdManager, function *core.DdNode, nvars, count int) []*core.DdNode {
	if count < 0 || nvars < 0 || float64(count) > core.Cudd_CountMinterm(dd, function, nvars) {
		return nil
	}
	strings := make([][]byte, count)
	for i := range strings {
		strings[i] = make([]byte, len(dd.Vars))
		for j := range strings[i] {
			strings[i][j] = '2'
		}
	}
	var pick func(*core.DdNode, [][]byte) bool
	pick = func(f *core.DdNode, rows [][]byte) bool {
		if len(rows) == 0 || f == core.Cudd_Not(core.Cudd_ReadOne(dd)) || f == core.Cudd_ReadOne(dd) {
			return true
		}
		node := core.Cudd_Regular(f)
		t, e := node.T, node.E
		if core.Cudd_IsComplement(f) {
			t, e = core.Cudd_Not(t), core.Cudd_Not(e)
		}
		oneCount := int(float64(len(rows))*core.Cudd_CountMinterm(dd, t, nvars)/(core.Cudd_CountMinterm(dd, t, nvars)+core.Cudd_CountMinterm(dd, e, nvars)) + 0.5)
		for i := 0; i < oneCount; i++ {
			rows[i][node.Index] = '1'
		}
		for i := oneCount; i < len(rows); i++ {
			rows[i][node.Index] = '0'
		}
		return pick(t, rows[:oneCount]) && pick(e, rows[oneCount:])
	}
	if !pick(function, strings) {
		return nil
	}
	selected := make([]*core.DdNode, count)
	saved := ""
	savePoint := -1
	saveFlag := false
	for i, row := range strings {
		isSame := false
		if !saveFlag {
			for j := i + 1; j < count; j++ {
				if string(row) == string(strings[j]) {
					savePoint = i
					saved = string(row)
					saveFlag = true
					break
				}
			}
		} else if string(row) == saved {
			isSame = true
		} else {
			saveFlag = false
			for j := i + 1; j < count; j++ {
				if string(row) == string(strings[j]) {
					savePoint = i
					saved = string(row)
					saveFlag = true
					break
				}
			}
		}
		for j := 0; j < nvars; j++ {
			index := int(core.Cudd_bddIthVar(dd, j).Index)
			if row[index] == '2' {
				if core.Cudd_Random()&0x20 != 0 {
					row[index] = '1'
				} else {
					row[index] = '0'
				}
			}
		}
		for isSame {
			isSame = false
			for j := savePoint; j < i; j++ {
				if string(row) == string(strings[j]) {
					isSame = true
					break
				}
			}
			if isSame {
				copy(row, []byte(saved))
				for j := 0; j < nvars; j++ {
					index := int(core.Cudd_bddIthVar(dd, j).Index)
					if row[index] == '2' {
						if core.Cudd_Random()&0x20 != 0 {
							row[index] = '1'
						} else {
							row[index] = '0'
						}
					}
				}
			}
		}
		cube := core.Cudd_ReadOne(dd)
		for j := 0; j < nvars; j++ {
			v := core.Cudd_bddIthVar(dd, j)
			if row[v.Index] == '0' {
				v = core.Cudd_Not(v)
			}
			cube = core.Cudd_bddAnd(dd, cube, v)
		}
		if !core.Cudd_bddLeq(dd, cube, function) {
			return nil
		}
		selected[i] = cube
	}
	return selected
}
