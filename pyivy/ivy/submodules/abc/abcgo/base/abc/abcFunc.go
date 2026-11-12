package abc

import (
	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab/primes"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/map/mio"
	"strings"
)

const ABC_MAX_CUBES = 100000

func funcSopVariables(text string) int {
	if index := strings.IndexByte(text, ' '); index >= 0 {
		return index
	}
	panic("SOP has no output separator")
}
func Abc_ConvertSopToBdd(dd *core.DdManager, text string, variables []*core.DdNode) *core.DdNode {
	count := funcSopVariables(text)
	rows := fanSOPRows(text, count)
	sum := dd.LogicZero
	variable := func(index int) *core.DdNode {
		if variables != nil {
			return variables[index]
		}
		return core.Cudd_bddIthVar(dd, index)
	}
	if len(rows) > 0 && (rows[0].output == 'x' || rows[0].output == 'n') {
		for i := 0; i < count; i++ {
			sum = core.Cudd_bddXor(dd, sum, variable(i))
		}
	} else {
		for _, row := range rows {
			cube := dd.One
			for i, c := range row.pattern {
				if c == '0' {
					cube = core.Cudd_bddAnd(dd, cube, core.Cudd_Not(variable(i)))
				} else if c == '1' {
					cube = core.Cudd_bddAnd(dd, cube, variable(i))
				}
			}
			sum = core.Cudd_bddOr(dd, sum, cube)
		}
	}
	if len(rows) > 0 && (rows[0].output == '0' || rows[0].output == 'n') {
		sum = core.Cudd_Not(sum)
	}
	return sum
}
func funcMaximumFanins(n *Network) int {
	maximum := 0
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if len(o.fanins) > maximum {
			maximum = len(o.fanins)
		}
	}
	return maximum
}
func funcBDDFromBoolean(dd *core.DdManager, fn func([]bool) bool, count int) *core.DdNode {
	values := make([]bool, count)
	var build func(int) *core.DdNode
	build = func(index int) *core.DdNode {
		if index == count {
			if fn(values) {
				return dd.One
			}
			return dd.LogicZero
		}
		values[index] = false
		low := build(index + 1)
		values[index] = true
		high := build(index + 1)
		return core.Cudd_bddIte(dd, core.Cudd_bddIthVar(dd, index), high, low)
	}
	return build(0)
}
func funcSopBDD(dd *core.DdManager, o *Object) *core.DdNode {
	if fn, ok := o.Data.(func([]bool) bool); ok {
		return funcBDDFromBoolean(dd, fn, len(o.fanins))
	}
	return Abc_ConvertSopToBdd(dd, o.Data.(string), nil)
}
func funcRemapBDD(source, destination *core.DdManager, root *core.DdNode, fanins []Edge) (*core.DdNode, []Edge) {
	permutation := make([]int, len(source.Vars))
	for i := range permutation {
		permutation[i] = -1
	}
	var ordered []Edge
	for _, index := range source.InvPerm {
		if index < len(fanins) {
			permutation[index] = len(ordered)
			ordered = append(ordered, fanins[index])
		}
	}
	return source.Transfer(root, destination, permutation), ordered
}
func Abc_NtkSopToBdd(n *Network) int32 {
	main := core.NewBDDManager(funcMaximumFanins(n))
	temporary := core.NewBDDManager(funcMaximumFanins(n))
	converted := make(map[*Object]*core.DdNode)
	reordered := make(map[*Object][]Edge)
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		manager := main
		if len(o.fanins) > 10 {
			manager = temporary
		}
		root := funcSopBDD(manager, o)
		if root == nil {
			return 0
		}
		if manager == temporary {
			core.Cudd_Ref(root)
			core.Cudd_ReduceHeap(temporary, core.CUDD_REORDER_SYMM_SIFT, 0)
			held := root
			root, reordered[o] = funcRemapBDD(temporary, main, root, o.fanins)
			core.Cudd_RecursiveDeref(temporary, held)
		}
		if root == nil {
			return 0
		}
		converted[o] = root
	}
	for o, root := range converted {
		o.Data = root
		core.Cudd_Ref(root)
		if fanins, ok := reordered[o]; ok {
			o.fanins = fanins
		}
	}
	n.FunctionManager = main
	n.Function = uint8(ABC_FUNC_BDD)
	clear(n.hashed)
	return 1
}

func Abc_ConvertZddToSop_rec(dd *core.DdManager, cover *core.DdNode, text *[]byte, fanins int32, cube *[]byte, phase int32, count *int32) {
	if cover == dd.Zero {
		return
	}
	if cover == dd.One {
		*text = append(*text, (*cube)[:fanins]...)
		*text = append(*text, ' ', byte('0'+phase), '\n')
		*count++
		return
	}
	positive, negative, absent, ok := core.ZDDCoverCofactors(dd, cover, int(cover.Index&^1))
	if !ok {
		panic("invalid paired ZDD cover")
	}
	index := int(cover.Index / 2)
	if index >= int(fanins) {
		panic("ZDD support exceeds fanins")
	}
	(*cube)[index] = '0'
	Abc_ConvertZddToSop_rec(dd, negative, text, fanins, cube, phase, count)
	(*cube)[index] = '1'
	Abc_ConvertZddToSop_rec(dd, positive, text, fanins, cube, phase, count)
	(*cube)[index] = '-'
	Abc_ConvertZddToSop_rec(dd, absent, text, fanins, cube, phase, count)
}
func Abc_ConvertZddToSop(dd *core.DdManager, cover *core.DdNode, text *[]byte, fanins int32, cube *[]byte, phase int32) int32 {
	*cube = append((*cube)[:0], strings.Repeat("-", int(fanins))...)
	*text = (*text)[:0]
	var count int32
	Abc_ConvertZddToSop_rec(dd, cover, text, fanins, cube, phase, &count)
	return count
}
func Abc_CountZddCubes_rec(dd *core.DdManager, cover *core.DdNode, count *int32) {
	if cover == dd.Zero || *count > ABC_MAX_CUBES {
		return
	}
	if cover == dd.One {
		*count++
		return
	}
	Abc_CountZddCubes_rec(dd, cover.T, count)
	Abc_CountZddCubes_rec(dd, cover.E, count)
}
func Abc_CountZddCubes(dd *core.DdManager, cover *core.DdNode) int32 {
	var count int32
	Abc_CountZddCubes_rec(dd, cover, &count)
	return count
}

// SOP strings are Go-owned; the legacy allocator parameter is unnecessary.
func Abc_ConvertBddToSop(_ any, dd *core.DdManager, on, onDC *core.DdNode, fanins, allPrimes int32, cube *[]byte, mode int32) string {
	if !core.Cudd_bddLeq(dd, on, onDC) {
		panic("inconsistent BDD bounds")
	}
	if on == dd.LogicZero {
		return " 0\n"
	}
	if onDC == dd.One {
		return " 1\n"
	}
	if mode < -1 || mode > 1 || mode == -1 && allPrimes != 0 {
		panic("invalid SOP phase mode")
	}
	coverFor := func(lower, upper *core.DdNode) *core.DdNode {
		if allPrimes != 0 {
			return primes.Extra_zddPrimes(dd, upper)
		}
		var cover *core.DdNode
		if core.Cudd_zddIsop(dd, lower, upper, &cover) == nil {
			return nil
		}
		return cover
	}
	var cover *core.DdNode
	phase := mode
	if mode == -1 {
		negative := coverFor(core.Cudd_Not(onDC), core.Cudd_Not(on))
		positive := coverFor(on, onDC)
		if negative == nil || positive == nil {
			return ""
		}
		if Abc_CountZddCubes(dd, positive) <= Abc_CountZddCubes(dd, negative) {
			cover = positive
			phase = 1
		} else {
			cover = negative
			phase = 0
		}
	} else if mode == 0 {
		cover = coverFor(core.Cudd_Not(onDC), core.Cudd_Not(on))
	} else {
		cover = coverFor(on, onDC)
	}
	if cover == nil || Abc_CountZddCubes(dd, cover) > ABC_MAX_CUBES {
		return ""
	}
	var text []byte
	Abc_ConvertZddToSop(dd, cover, &text, fanins, cube, phase)
	return string(text)
}
func Abc_NtkBddToSop(n *Network, mode, cubeLimit int32) int32 {
	dd, ok := n.FunctionManager.(*core.DdManager)
	if !ok {
		return 0
	}
	converted := make(map[*Object]string)
	var cube []byte
	total := 0
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		root, ok := o.Data.(*core.DdNode)
		if !ok {
			return 0
		}
		text := Abc_ConvertBddToSop(nil, dd, root, root, int32(len(o.fanins)), 0, &cube, mode)
		if text == "" {
			return 0
		}
		total += strings.Count(text, "\n")
		if cubeLimit >= 0 && total > int(cubeLimit) {
			return 0
		}
		converted[o] = text
	}
	for o, text := range converted {
		if funcSopVariables(text) == 0 {
			Abc_ObjRemoveFanins(o)
		}
		o.Data = text
	}
	n.FunctionManager = nil
	n.Function = uint8(ABC_FUNC_SOP)
	Abc_NtkSortSops(n)
	return 1
}
func Abc_NodeBddToCnf(o *Object, allocator any, cube *[]byte, allPrimes int32, negative, positive *string) {
	dd, ok := o.network.FunctionManager.(*core.DdManager)
	if !ok {
		panic("BDD functionality manager missing")
	}
	root := o.Data.(*core.DdNode)
	*negative = Abc_ConvertBddToSop(allocator, dd, root, root, int32(len(o.fanins)), allPrimes, cube, 0)
	*positive = Abc_ConvertBddToSop(allocator, dd, root, root, int32(len(o.fanins)), allPrimes, cube, 1)
}
func Abc_NtkLogicMakeDirectSops(n *Network) {
	dd := core.NewBDDManager(funcMaximumFanins(n))
	converted := make(map[*Object]string)
	var cube []byte
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		rows := fanNodeRows(o)
		if len(rows) == 0 || rows[0].output != '0' && rows[0].output != 'n' {
			continue
		}
		root := Abc_ConvertSopToBdd(dd, o.Data.(string), nil)
		text := Abc_ConvertBddToSop(nil, dd, root, root, int32(len(o.fanins)), 0, &cube, 1)
		if text == "" {
			panic("SOP direct phase conversion exceeded cube limit")
		}
		converted[o] = text
	}
	for o, text := range converted {
		if funcSopVariables(text) == 0 {
			Abc_ObjRemoveFanins(o)
		}
		o.Data = text
	}
}

func Abc_ConvertSopToAigInternal(manager *hop.Hop_Man_t, text string) *hop.Hop_Obj_t {
	count := funcSopVariables(text)
	rows := fanSOPRows(text, count)
	sum := hop.Hop_ManConst0(manager)
	if len(rows) > 0 && (rows[0].output == 'x' || rows[0].output == 'n') {
		for i := 0; i < count; i++ {
			sum = hop.Hop_Exor(manager, sum, hop.Hop_IthVar(manager, int32(i)))
		}
	} else {
		for _, row := range rows {
			cube := hop.Hop_ManConst1(manager)
			for i, c := range row.pattern {
				if c == '0' {
					cube = hop.Hop_And(manager, cube, hop.Hop_Not(hop.Hop_IthVar(manager, int32(i))))
				} else if c == '1' {
					cube = hop.Hop_And(manager, cube, hop.Hop_IthVar(manager, int32(i)))
				}
			}
			sum = hop.Hop_Or(manager, sum, cube)
		}
	}
	if len(rows) > 0 && (rows[0].output == '0' || rows[0].output == 'n') {
		sum = hop.Hop_Not(sum)
	}
	return sum
}
func Abc_ConvertSopToAig(manager *hop.Hop_Man_t, text string) *hop.Hop_Obj_t {
	variables := funcSopVariables(text)
	rows := fanSOPRows(text, variables)
	if variables > 2 && len(rows) > 1 && rows[0].output != 'x' && rows[0].output != 'n' {
		return dec.Dec_GraphFactorSop(manager, text).(*hop.Hop_Obj_t)
	}
	return Abc_ConvertSopToAigInternal(manager, text)
}
func Abc_NtkSopToAig(n *Network) int32 {
	manager := hop.Hop_ManStart()
	maximum := funcMaximumFanins(n)
	if maximum > 0 {
		hop.Hop_IthVar(manager, int32(maximum-1))
	}
	converted := make(map[*Object]*hop.Hop_Obj_t)
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		var root *hop.Hop_Obj_t
		if text, ok := o.Data.(string); ok {
			root = Abc_ConvertSopToAig(manager, text)
		} else if fn, ok := o.Data.(func([]bool) bool); ok {
			values := make([]bool, len(o.fanins))
			var build func(int) *hop.Hop_Obj_t
			build = func(index int) *hop.Hop_Obj_t {
				if index == len(values) {
					if fn(values) {
						return hop.Hop_ManConst1(manager)
					}
					return hop.Hop_ManConst0(manager)
				}
				values[index] = false
				low := build(index + 1)
				values[index] = true
				high := build(index + 1)
				return hop.Hop_Mux(manager, hop.Hop_IthVar(manager, int32(index)), high, low)
			}
			root = build(0)
		} else {
			return 0
		}
		if root == nil {
			return 0
		}
		converted[o] = root
	}
	for o, root := range converted {
		o.Data = root
	}
	n.FunctionManager = manager
	n.Function = uint8(ABC_FUNC_AIG)
	return 1
}
func Abc_ConvertAigToBdd_rec1(dd *core.DdManager, o *hop.Hop_Obj_t) {
	o = hop.Hop_Regular(o)
	if hop.Hop_ObjIsNode(o) == 0 || hop.Hop_ObjIsMarkA(o) != 0 {
		return
	}
	Abc_ConvertAigToBdd_rec1(dd, hop.Hop_ObjFanin0(o))
	Abc_ConvertAigToBdd_rec1(dd, hop.Hop_ObjFanin1(o))
	a := hop.Hop_ObjData(hop.Hop_ObjFanin0(o)).(*core.DdNode)
	b := hop.Hop_ObjData(hop.Hop_ObjFanin1(o)).(*core.DdNode)
	a = core.Cudd_NotCond(a, hop.Hop_ObjFaninC0(o) != 0)
	b = core.Cudd_NotCond(b, hop.Hop_ObjFaninC1(o) != 0)
	var root *core.DdNode
	if hop.Hop_ObjIsExor(o) != 0 {
		root = core.Cudd_bddXor(dd, a, b)
	} else {
		root = core.Cudd_bddAnd(dd, a, b)
	}
	core.Cudd_Ref(root)
	hop.Hop_ObjSetData(o, root)
	hop.Hop_ObjSetMarkA(o)
}
func Abc_ConvertAigToBdd_rec2(dd *core.DdManager, o *hop.Hop_Obj_t) {
	o = hop.Hop_Regular(o)
	if hop.Hop_ObjIsNode(o) == 0 || hop.Hop_ObjIsMarkA(o) == 0 {
		return
	}
	Abc_ConvertAigToBdd_rec2(dd, hop.Hop_ObjFanin0(o))
	Abc_ConvertAigToBdd_rec2(dd, hop.Hop_ObjFanin1(o))
	core.Cudd_RecursiveDeref(dd, hop.Hop_ObjData(o).(*core.DdNode))
	hop.Hop_ObjSetData(o, nil)
	hop.Hop_ObjClearMarkA(o)
}
func Abc_ConvertAigToBdd(dd *core.DdManager, root *hop.Hop_Obj_t) *core.DdNode {
	regular := hop.Hop_Regular(root)
	if hop.Hop_ObjIsConst1(regular) != 0 {
		return core.Cudd_NotCond(dd.One, hop.Hop_IsComplement(root) != 0)
	}
	Abc_ConvertAigToBdd_rec1(dd, regular)
	result := core.Cudd_NotCond(hop.Hop_ObjData(regular).(*core.DdNode), hop.Hop_IsComplement(root) != 0)
	core.Cudd_Ref(result)
	Abc_ConvertAigToBdd_rec2(dd, regular)
	core.Cudd_Deref(result)
	return result
}
func Abc_NtkAigToBdd(n *Network) int32 {
	manager, ok := n.FunctionManager.(*hop.Hop_Man_t)
	if !ok {
		return 0
	}
	maximum := funcMaximumFanins(n)
	temporary, main := core.NewBDDManager(maximum), core.NewBDDManager(maximum)
	hop.Hop_ObjSetData(hop.Hop_ManConst1(manager), temporary.One)
	for i := 0; i < maximum; i++ {
		hop.Hop_ObjSetData(hop.Hop_IthVar(manager, int32(i)), core.Cudd_bddIthVar(temporary, i))
	}
	converted := make(map[*Object]*core.DdNode)
	reordered := make(map[*Object][]Edge)
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		root, ok := o.Data.(*hop.Hop_Obj_t)
		if !ok {
			return 0
		}
		function := Abc_ConvertAigToBdd(temporary, root)
		if function == nil {
			return 0
		}
		core.Cudd_Ref(function)
		core.Cudd_ReduceHeap(temporary, core.CUDD_REORDER_SYMM_SIFT, 0)
		held := function
		function, reordered[o] = funcRemapBDD(temporary, main, function, o.fanins)
		core.Cudd_RecursiveDeref(temporary, held)
		if function == nil {
			return 0
		}
		converted[o] = function
	}
	for o, root := range converted {
		o.Data = root
		core.Cudd_Ref(root)
		o.fanins = reordered[o]
	}
	n.FunctionManager = main
	n.Function = uint8(ABC_FUNC_BDD)
	return 1
}

func Abc_NtkMapToSop(n *Network) int32 {
	converted := make(map[*Object]string)
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			continue
		}
		gate, ok := o.Data.(*mio.Mio_Gate_t)
		if !ok {
			return 0
		}
		text := mio.Mio_GateReadSop(gate)
		fanSOPRows(text, len(o.fanins))
		converted[o] = text
	}
	for o, text := range converted {
		o.Data = text
	}
	n.FunctionManager = nil
	n.Function = uint8(ABC_FUNC_SOP)
	return 1
}
func Abc_NtkSopToBlifMv(n *Network) int32 { return Abc_NtkConvertToBlifMv(n) }
func Abc_NtkToSop(n *Network, mode, cubeLimit int32) int32 {
	switch Abc_NtkFunc_t(n.Function) {
	case ABC_FUNC_BLACKBOX:
		return 1
	case ABC_FUNC_SOP:
		if mode == -1 {
			return 1
		}
		if Abc_NtkSopToBdd(n) == 0 {
			return 0
		}
		return Abc_NtkBddToSop(n, mode, cubeLimit)
	case ABC_FUNC_MAP:
		return Abc_NtkMapToSop(n)
	case ABC_FUNC_BDD:
		return Abc_NtkBddToSop(n, mode, cubeLimit)
	case ABC_FUNC_AIG:
		if Abc_NtkAigToBdd(n) == 0 {
			return 0
		}
		return Abc_NtkBddToSop(n, mode, cubeLimit)
	}
	return 0
}
func Abc_NtkToBdd(n *Network) int32 {
	switch Abc_NtkFunc_t(n.Function) {
	case ABC_FUNC_BLACKBOX, ABC_FUNC_BDD:
		return 1
	case ABC_FUNC_MAP:
		if Abc_NtkMapToSop(n) == 0 {
			return 0
		}
		return Abc_NtkSopToBdd(n)
	case ABC_FUNC_SOP:
		if Abc_NtkSopToAig(n) == 0 {
			return 0
		}
		return Abc_NtkAigToBdd(n)
	case ABC_FUNC_AIG:
		return Abc_NtkAigToBdd(n)
	}
	return 0
}
func Abc_NtkToAig(n *Network) int32 {
	switch Abc_NtkFunc_t(n.Function) {
	case ABC_FUNC_BLACKBOX, ABC_FUNC_AIG:
		return 1
	case ABC_FUNC_MAP:
		if Abc_NtkMapToSop(n) == 0 {
			return 0
		}
		return Abc_NtkSopToAig(n)
	case ABC_FUNC_BDD:
		if Abc_NtkBddToSop(n, -1, int32(^uint32(0)>>1)) == 0 {
			return 0
		}
		return Abc_NtkSopToAig(n)
	case ABC_FUNC_SOP:
		return Abc_NtkSopToAig(n)
	}
	return 0
}

// LocalAIGMapping retains original node IDs and local-cut boundaries on the
// importable AIG manager used by the Go replacement for the GIA export helper.
type LocalAIGMapping struct {
	OriginalIDs map[int]int
	Cuts        map[int][]int
	Simple      bool
}

func Abc_ConvertAigToGia_rec1(graph *aig.Graph, o *hop.Hop_Obj_t) {
	o = hop.Hop_Regular(o)
	if hop.Hop_ObjIsNode(o) == 0 || hop.Hop_ObjIsMarkA(o) != 0 {
		return
	}
	Abc_ConvertAigToGia_rec1(graph, hop.Hop_ObjFanin0(o))
	Abc_ConvertAigToGia_rec1(graph, hop.Hop_ObjFanin1(o))
	a, b := aig.Edge(hop.Hop_ObjIData(hop.Hop_ObjFanin0(o))), aig.Edge(hop.Hop_ObjIData(hop.Hop_ObjFanin1(o)))
	if hop.Hop_ObjFaninC0(o) != 0 {
		a = a.Not()
	}
	if hop.Hop_ObjFaninC1(o) != 0 {
		b = b.Not()
	}
	value := graph.AND(a, b)
	if hop.Hop_ObjIsExor(o) != 0 {
		value = graph.XOR(a, b)
	}
	hop.Hop_ObjSetIData(o, int32(value))
	hop.Hop_ObjSetMarkA(o)
}
func Abc_ConvertAigToGia_rec2(o *hop.Hop_Obj_t) {
	o = hop.Hop_Regular(o)
	if hop.Hop_ObjIsNode(o) == 0 || hop.Hop_ObjIsMarkA(o) == 0 {
		return
	}
	Abc_ConvertAigToGia_rec2(hop.Hop_ObjFanin0(o))
	Abc_ConvertAigToGia_rec2(hop.Hop_ObjFanin1(o))
	hop.Hop_ObjClearMarkA(o)
}
func Abc_ConvertAigToGia(graph *aig.Graph, root *hop.Hop_Obj_t) aig.Edge {
	regular := hop.Hop_Regular(root)
	if hop.Hop_ObjIsConst1(regular) != 0 {
		if hop.Hop_IsComplement(root) != 0 {
			return 1
		}
		return 0
	}
	Abc_ConvertAigToGia_rec1(graph, regular)
	Abc_ConvertAigToGia_rec2(regular)
	result := aig.Edge(hop.Hop_ObjIData(regular))
	if hop.Hop_IsComplement(root) != 0 {
		result = result.Not()
	}
	return result
}
func Abc_NtkAigToGia(n *Network, simple int32) *aig.Graph {
	manager, ok := n.FunctionManager.(*hop.Hop_Man_t)
	if !ok {
		return nil
	}
	graph := aig.NewGraph()
	graph.Name = n.Name
	graph.Spec = n.Spec
	mapping := &LocalAIGMapping{OriginalIDs: map[int]int{}, Cuts: map[int][]int{}, Simple: simple != 0}
	graph.Data = mapping
	n.constant.Temp = 0
	hop.Hop_ObjSetIData(hop.Hop_ManConst1(manager), 0)
	for _, o := range n.cis {
		o.Temp = int(graph.NewInput())
	}
	for _, o := range Abc_NtkDfs(n, 0) {
		var value aig.Edge
		if o.Barrier {
			edge := o.fanins[0]
			value = aig.Edge(edge.Object.Temp)
			if edge.Complement {
				value = value.Not()
			}
			value = aig.Aig_ObjCreate(graph, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{value, 0}})
		} else {
			for i, edge := range o.fanins {
				value := aig.Edge(edge.Object.Temp)
				if edge.Complement {
					value = value.Not()
				}
				hop.Hop_ObjSetIData(hop.Hop_IthVar(manager, int32(i)), int32(value))
			}
			root, ok := o.Data.(*hop.Hop_Obj_t)
			if !ok {
				return nil
			}
			value = Abc_ConvertAigToGia(graph, root)
			if simple == 0 && graph.Node(value).Kind == aig.And {
				if _, exists := mapping.Cuts[value.NodeID()]; !exists {
					cut := make([]int, len(o.fanins))
					for i, edge := range o.fanins {
						cut[i] = aig.Edge(edge.Object.Temp).NodeID()
					}
					mapping.Cuts[value.NodeID()] = cut
				}
			}
		}
		o.Temp = int(value)
		mapping.OriginalIDs[value.NodeID()] = o.ID
	}
	for _, o := range n.cos {
		edge := o.fanins[0]
		value := aig.Edge(edge.Object.Temp)
		if edge.Complement {
			value = value.Not()
		}
		graph.AddOutput(value)
	}
	aig.Aig_ManSetRegNum(graph, int32(len(n.ObjectsOfKind(ObjectLatch))))
	return graph
}
func Abc_ConvertAigToAig_rec(n *Network, o *hop.Hop_Obj_t) {
	o = hop.Hop_Regular(o)
	if hop.Hop_ObjIsNode(o) == 0 || hop.Hop_ObjIsMarkA(o) != 0 {
		return
	}
	Abc_ConvertAigToAig_rec(n, hop.Hop_ObjFanin0(o))
	Abc_ConvertAigToAig_rec(n, hop.Hop_ObjFanin1(o))
	a, b := hop.Hop_ObjData(hop.Hop_ObjFanin0(o)).(Edge), hop.Hop_ObjData(hop.Hop_ObjFanin1(o)).(Edge)
	if hop.Hop_ObjFaninC0(o) != 0 {
		a = a.Not()
	}
	if hop.Hop_ObjFaninC1(o) != 0 {
		b = b.Not()
	}
	value := n.AndEdges(a, b)
	if hop.Hop_ObjIsExor(o) != 0 {
		value = n.AndEdges(n.AndEdges(a, b.Not()).Not(), n.AndEdges(a.Not(), b).Not()).Not()
	}
	hop.Hop_ObjSetData(o, value)
	hop.Hop_ObjSetMarkA(o)
}
func Abc_ConvertAigToAig(n *Network, old *Object) Edge {
	manager, ok := old.network.FunctionManager.(*hop.Hop_Man_t)
	if !ok {
		panic("local Hop manager missing")
	}
	root := old.Data.(*hop.Hop_Obj_t)
	hop.Hop_ObjSetData(hop.Hop_ManConst1(manager), n.One())
	for i, edge := range old.fanins {
		value := getCopy(edge.Object)
		value.Complement = value.Complement != edge.Complement
		hop.Hop_ObjSetData(hop.Hop_IthVar(manager, int32(i)), value)
	}
	if hop.Hop_ObjIsConst1(hop.Hop_Regular(root)) != 0 {
		value := n.One()
		if hop.Hop_IsComplement(root) != 0 {
			value = value.Not()
		}
		return value
	}
	Abc_ConvertAigToAig_rec(n, hop.Hop_Regular(root))
	hop.Hop_ConeUnmark_rec(hop.Hop_Regular(root))
	value := hop.Hop_ObjData(hop.Hop_Regular(root)).(Edge)
	value.Complement = value.Complement != (hop.Hop_IsComplement(root) != 0)
	return value
}
