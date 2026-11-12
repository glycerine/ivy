package abc

import (
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/nm"
	"strconv"
	"strings"
)

func Abc_ObjAlloc(n *Network, kind ObjectKind) *Object {
	return &Object{network: n, ID: -1, Kind: kind}
}
func Abc_ObjRecycle(o *Object) {
	if o.network != nil && o.network.contains(o) {
		panic("cannot recycle a live object")
	}
	*o = Object{ID: -1}
}
func Abc_NtkCreateObj(n *Network, kind ObjectKind) *Object { return n.NewObject(kind) }
func Abc_NtkDeleteObj(o *Object) {
	n := o.network
	if !n.contains(o) || o == n.constant {
		panic("invalid object deletion")
	}
	for _, fanout := range o.Fanouts() {
		Abc_ObjDeleteFanin(fanout, o)
	}
	if err := n.DeleteObject(o); err != nil {
		panic(err)
	}
	Abc_ObjRecycle(o)
}
func Abc_NtkDeleteObjPo(o *Object) {
	if o.Kind != ObjectPO {
		panic("expected PO")
	}
	Abc_NtkDeleteObj(o)
}
func Abc_NtkDeleteObj_rec(o *Object, onlyNodes int32) { deleteDangling(o, onlyNodes != 0) }
func Abc_NtkDeleteAll_rec(o *Object) {
	n := o.network
	fanins := o.Fanins()
	if len(o.fanouts) != 0 {
		panic("object has fanouts")
	}
	Abc_NtkDeleteObj(o)
	for _, e := range fanins {
		source := e.Object
		if n.contains(source) && source != n.constant && source.Kind != ObjectNode && len(source.fanouts) == 0 {
			Abc_NtkDeleteAll_rec(source)
		}
	}
}

// Go functions, immutable strings and explicit graph handles can be shared;
// specialized mutable functionality should implement CloneForNetwork.
type NetworkFunctionCloner interface{ CloneForNetwork(*Network) any }

func Abc_NtkDupObj(n *Network, old *Object, copyName int32) *Object {
	o := n.NewObject(old.Kind)
	copyMVVariable(old, o)
	o.Level = old.Level
	if cloner, ok := old.Data.(NetworkFunctionCloner); ok {
		o.Data = cloner.CloneForNetwork(n)
	} else {
		switch data := old.Data.(type) {
		case *core.DdNode:
			source, ok := old.network.FunctionManager.(*core.DdManager)
			if !ok {
				panic("BDD source manager missing")
			}
			destination, ok := n.FunctionManager.(*core.DdManager)
			if !ok {
				destination = core.NewBDDManager(len(source.Vars))
				n.FunctionManager = destination
			}
			if source == destination {
				o.Data = data
			} else {
				o.Data = source.Transfer(data, destination, nil)
			}
			core.Cudd_Ref(o.Data.(*core.DdNode))
		case *hop.Hop_Obj_t:
			source, ok := old.network.FunctionManager.(*hop.Hop_Man_t)
			if !ok {
				panic("Hop source manager missing")
			}
			destination, ok := n.FunctionManager.(*hop.Hop_Man_t)
			if !ok {
				destination = hop.Hop_ManStart()
				n.FunctionManager = destination
			}
			o.Data = hop.Hop_Transfer(source, destination, data, int32(len(old.fanins)))
		default:
			o.Data = old.Data
		}
	}
	if copyName != 0 && (old.IsCI() || old.IsCO() || Abc_ObjIsBox(old) != 0 || old.Kind == ObjectNet) {
		if name, ok := old.network.names.Lookup(int32(old.ID)); ok {
			o.SetName(name)
		}
	}
	old.Copy = o
	return o
}
func Abc_NtkDupBox(n *Network, box *Object, copyName int32) *Object {
	if Abc_ObjIsBox(box) == 0 {
		panic("expected box")
	}
	dup := Abc_NtkDupObj(n, box, copyName)
	for _, fanin := range box.fanins {
		term := Abc_NtkDupObj(n, fanin.Object, copyName)
		Abc_ObjAddFanin(dup, term)
	}
	for _, fanout := range box.fanouts {
		term := Abc_NtkDupObj(n, fanout, copyName)
		Abc_ObjAddFanin(term, dup)
	}
	return dup
}
func Abc_NtkCloneObj(old *Object) *Object {
	o := old.network.NewObject(old.Kind)
	// Match ABC's cloning convention: copy regular fanins, but not functionality.
	for _, fanin := range old.fanins {
		Abc_ObjAddFanin(o, fanin.Object)
	}
	return o
}
func findNamedKind(n *Network, name string, kinds ...ObjectKind) *Object {
	for _, kind := range kinds {
		id := nm.Nm_ManFindIdByName(n.names, name, int32(kind))
		if id >= 0 {
			return n.Object(int(id))
		}
	}
	return nil
}
func Abc_NtkFindNode(n *Network, name string) *Object {
	if term := findNamedKind(n, name, ObjectPO, ObjectBI); term != nil && len(term.fanins) > 0 {
		return term.fanins[0].Object
	}
	if o := findNamedKind(n, name, ObjectNode); o != nil {
		return o
	}
	if !strings.HasPrefix(name, "n") {
		return nil
	}
	id, err := strconv.Atoi(name[1:])
	if err != nil {
		return nil
	}
	o := n.Object(id)
	if o != nil && o.Kind == ObjectNode {
		return o
	}
	return nil
}
func Abc_NtkFindNet(n *Network, name string) *Object { return findNamedKind(n, name, ObjectNet) }
func Abc_NtkFindCi(n *Network, name string) *Object {
	return findNamedKind(n, name, ObjectPI, ObjectBO)
}
func Abc_NtkFindCo(n *Network, name string) *Object {
	return findNamedKind(n, name, ObjectPO, ObjectBI)
}
func Abc_NtkFindOrCreateNet(n *Network, name string) *Object {
	if o := Abc_NtkFindNet(n, name); o != nil {
		return o
	}
	o := n.NewObject(ObjectNet)
	o.SetName(name)
	return o
}
func createLogicNode(n *Network, fanins []*Object, fn func([]bool) bool) *Object {
	o := n.NewObject(ObjectNode)
	for _, fanin := range fanins {
		Abc_ObjAddFanin(o, fanin)
	}
	o.Data = fn
	return o
}
func Abc_NtkCreateNodeConst0(n *Network) *Object {
	return createLogicNode(n, nil, func([]bool) bool { return false })
}
func Abc_NtkCreateNodeConst1(n *Network) *Object {
	return createLogicNode(n, nil, func([]bool) bool { return true })
}
func Abc_NtkCreateNodeInv(n *Network, fanin *Object) *Object {
	return createLogicNode(n, []*Object{fanin}, func(v []bool) bool { return !v[0] })
}
func Abc_NtkCreateNodeBuf(n *Network, fanin *Object) *Object {
	return createLogicNode(n, []*Object{fanin}, func(v []bool) bool { return v[0] })
}

// Fanin vectors are Go slices in the port; ownership stays with the caller.
func Abc_NtkCreateNodeAnd(n *Network, fanins []*Object) *Object {
	return createLogicNode(n, fanins, func(v []bool) bool {
		for _, b := range v {
			if !b {
				return false
			}
		}
		return true
	})
}
func Abc_NtkCreateNodeOr(n *Network, fanins []*Object) *Object {
	return createLogicNode(n, fanins, func(v []bool) bool {
		for _, b := range v {
			if b {
				return true
			}
		}
		return false
	})
}
func Abc_NtkCreateNodeExor(n *Network, fanins []*Object) *Object {
	return createLogicNode(n, fanins, func(v []bool) bool {
		r := false
		for _, b := range v {
			r = r != b
		}
		return r
	})
}
func Abc_NtkCreateNodeMux(n *Network, c, t, e *Object) *Object {
	return createLogicNode(n, []*Object{c, t, e}, func(v []bool) bool {
		if v[0] {
			return v[1]
		}
		return v[2]
	})
}
func Abc_NodeIsConst(o *Object) int32 { return abcBool(o.Kind == ObjectNode && len(o.fanins) == 0) }
func booleanFunction(o *Object) (func([]bool) bool, bool) {
	if fn, ok := o.Data.(func([]bool) bool); ok {
		return fn, true
	}
	if o.Data == aigANDMarker {
		return func(v []bool) bool { return v[0] && v[1] }, true
	}
	switch value := o.Data.(type) {
	case string:
		return func(values []bool) bool { return evaluateLocalSOP(value, values) }, true
	case *mio.Mio_Gate_t:
		if value == nil {
			return nil, false
		}
		return func(values []bool) bool { return evaluateLocalSOP(mio.Mio_GateReadSop(value), values) }, true
	case *core.DdNode:
		manager, ok := o.network.FunctionManager.(*core.DdManager)
		if !ok {
			return nil, false
		}
		return func(values []bool) bool {
			result, err := manager.Evaluate(value, values)
			if err != nil {
				panic(err)
			}
			return result
		}, true
	case *hop.Hop_Obj_t:
		return func(values []bool) bool { return evaluateLocalHop(value, values) }, true
	}
	return nil, false
}
func evaluateLocalSOP(text string, values []bool) bool {
	if zero := strings.IndexByte(text, 0); zero >= 0 {
		text = text[:zero]
	}
	phase := byte('1')
	matched := false
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		if len(line) != len(values)+2 || line[len(values)] != ' ' {
			panic("invalid SOP arity")
		}
		phase = line[len(values)+1]
		if phase == 'x' || phase == 'n' {
			parity := false
			for _, value := range values {
				parity = parity != value
			}
			return parity != (phase == 'n')
		}
		match := true
		for i, c := range line[:len(values)] {
			if c != '-' && (c == '1') != values[i] {
				match = false
				break
			}
		}
		matched = matched || match
	}
	return matched == (phase == '1')
}
func evaluateLocalHop(root *hop.Hop_Obj_t, values []bool) bool {
	memo := map[*hop.Hop_Obj_t]bool{}
	var visit func(*hop.Hop_Obj_t) bool
	visit = func(edge *hop.Hop_Obj_t) bool {
		node := hop.Hop_Regular(edge)
		result, ok := memo[node]
		if !ok {
			switch {
			case hop.Hop_ObjIsConst1(node) != 0:
				result = true
			case hop.Hop_ObjIsPi(node) != 0:
				result = values[node.PioNum]
			default:
				a, b := visit(hop.Hop_ObjChild0(node)), visit(hop.Hop_ObjChild1(node))
				if hop.Hop_ObjIsExor(node) != 0 {
					result = a != b
				} else {
					result = a && b
				}
			}
			memo[node] = result
		}
		return result != (hop.Hop_IsComplement(edge) != 0)
	}
	return visit(root)
}
func localFunction(o *Object) func([]bool) bool {
	if function, ok := booleanFunction(o); ok {
		return function
	}
	panic("object functionality is not Boolean")
}
func Abc_NodeIsConst0(o *Object) int32 {
	return abcBool(Abc_NodeIsConst(o) != 0 && !localFunction(o)(nil))
}
func Abc_NodeIsConst1(o *Object) int32 {
	return abcBool(Abc_NodeIsConst(o) != 0 && localFunction(o)(nil))
}
func Abc_NodeIsBuf(o *Object) int32 {
	if o.Kind != ObjectNode || len(o.fanins) != 1 {
		return 0
	}
	fn := localFunction(o)
	return abcBool(!fn([]bool{false}) && fn([]bool{true}))
}
func Abc_NodeIsInv(o *Object) int32 {
	if o.Kind != ObjectNode || len(o.fanins) != 1 {
		return 0
	}
	fn := localFunction(o)
	return abcBool(fn([]bool{false}) && !fn([]bool{true}))
}
func Abc_NodeComplement(o *Object) {
	fn := localFunction(o)
	o.Data = func(v []bool) bool { return !fn(v) }
	clear(o.network.hashed)
}
func Abc_NodeComplementInput(o, fanin *Object) {
	index := Abc_ObjFanoutFaninNum(o, fanin)
	if index < 0 {
		panic("missing fanin")
	}
	fn := localFunction(o)
	o.Data = func(v []bool) bool {
		copyValues := append([]bool(nil), v...)
		copyValues[index] = !copyValues[index]
		return fn(copyValues)
	}
	clear(o.network.hashed)
}
