package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	"strings"
)

func Abc_NtkAlloc(kind Abc_NtkType_t, function Abc_NtkFunc_t, useMemory int32) *Network {
	n := NewNetwork("")
	n.Type, n.Function = uint8(kind), uint8(function)
	switch {
	case kind == ABC_NTK_STRASH:
		n.FunctionManager = Abc_AigAlloc(n)
	case function == ABC_FUNC_SOP || function == ABC_FUNC_BLIFMV:
		n.FunctionManager = mem.Mem_FlexStart()
	case function == ABC_FUNC_BDD:
		n.FunctionManager = core.NewBDDManager(20)
	case function == ABC_FUNC_AIG:
		n.FunctionManager = hop.Hop_ManStart()
	case function == ABC_FUNC_MAP || function == ABC_FUNC_BLACKBOX || function == ABC_FUNC_NONE: // libraries are supplied by the caller
	default:
		panic("abc: unsupported network function")
	}
	return n
}
func ntkCopyMetadata(dst, src *Network) {
	dst.Name, dst.Spec = src.Name, src.Spec
	dst.Constraints, dst.Barriers = src.Constraints, src.Barriers
	dst.AndGateDelay, dst.WireLoad = src.AndGateDelay, src.WireLoad
	dst.Onehots = make([][]int32, len(src.Onehots))
	for i, row := range src.Onehots {
		dst.Onehots[i] = append([]int32(nil), row...)
	}
	dst.ObjectPermutation = append([]int32(nil), src.ObjectPermutation...)
	dst.Phases = append([]int32(nil), src.Phases...)
	dst.LTLProperties = append([]string(nil), src.LTLProperties...)
	if src.SequenceModel != nil {
		dst.SequenceModel = util.Abc_CexDup(src.SequenceModel, src.SequenceModel.NRegs)
	}
	dst.phaseTracking = src.phaseTracking
}
func ntkNewLike(src *Network, kind Abc_NtkType_t, function Abc_NtkFunc_t) *Network {
	n := Abc_NtkAlloc(kind, function, 1)
	ntkCopyMetadata(n, src)
	if function == ABC_FUNC_MAP {
		n.FunctionManager = src.FunctionManager
	}
	return n
}

// Graph-changing constructions start with fresh per-graph metadata. The
// immutable mapping library is the only function-manager state borrowed here.
func ntkFreshLike(src *Network, kind Abc_NtkType_t, function Abc_NtkFunc_t) *Network {
	dst := Abc_NtkAlloc(kind, function, 1)
	if function == ABC_FUNC_MAP {
		dst.FunctionManager = src.FunctionManager
	}
	return dst
}
func ntkClearCopies(n *Network) {
	for _, o := range n.Objects() {
		o.Copy = nil
		o.CopyEdge = Edge{}
	}
	setCopy(n.constant, n.One())
}
func ntkCopiedEdge(e Edge) Edge {
	copy := getCopy(e.Object)
	if copy.Object == nil {
		panic("abc: missing network copy")
	}
	copy.Complement = copy.Complement != e.Complement
	return copy
}
func ntkCopyTiming(dst, src *Network) {
	if src.MappingTiming != nil {
		Abc_NtkSetMappingTiming(dst, src.MappingTiming)
	}
	if src.CILevel != nil {
		levels := make(map[*Object]int)
		for _, o := range src.cis {
			if copy := getCopy(o); copy.Object != nil {
				levels[copy.Object] = src.CILevel(o)
			}
		}
		dst.CILevel = func(o *Object) int { return levels[o] }
	}
	if src.timingMapping == nil {
		return
	}
	if m, ok := src.timingMapping.(*scl.NetworkMapping); ok {
		copy := &scl.NetworkMapping{Lib: m.Lib, Mio: m.Mio, MioGates: make([]*mio.Mio_Gate_t, len(dst.objects))}
		if m.GateIDs != nil {
			copy.GateIDs = make([]int32, len(dst.objects))
			for i := range copy.GateIDs {
				copy.GateIDs[i] = -1
			}
		}
		for _, o := range src.Objects() {
			e := getCopy(o)
			if e.Object == nil {
				continue
			}
			id := e.Object.ID
			if o.ID < len(m.MioGates) {
				copy.MioGates[id] = m.MioGates[o.ID]
			}
			if o.ID < len(m.GateIDs) {
				copy.GateIDs[id] = m.GateIDs[o.ID]
			}
		}
		dst.timingMapping = copy
		return
	}
	if cloner, ok := src.timingMapping.(NetworkFunctionCloner); ok {
		dst.timingMapping = cloner.CloneForNetwork(dst)
		return
	}
	panic("abc: timing metadata needs a network clone implementation")
}
func Abc_NtkStartFrom(src *Network, kind Abc_NtkType_t, function Abc_NtkFunc_t) *Network {
	if src == nil {
		return nil
	}
	dst := ntkNewLike(src, kind, function)
	ntkClearCopies(src)
	setCopy(src.constant, dst.One())
	names := int32(1)
	if kind == ABC_NTK_NETLIST {
		names = 0
	}
	for _, o := range src.ObjectsOfKind(ObjectPI) {
		Abc_NtkDupObj(dst, o, names)
	}
	for _, o := range src.ObjectsOfKind(ObjectPO) {
		Abc_NtkDupObj(dst, o, names)
	}
	for _, box := range src.boxes {
		Abc_NtkDupBox(dst, box, names)
	}
	for _, o := range src.cis {
		getCopy(o).Object.Level = o.Level
	}
	Abc_NtkOrderCisCos(dst)
	ntkCopyTiming(dst, src)
	return dst
}
func Abc_NtkStartFromWithLatches(src *Network, kind Abc_NtkType_t, function Abc_NtkFunc_t, count int32) *Network {
	if src == nil {
		return nil
	}
	if Abc_NtkLatchNum(src) != 0 || count < 0 || count > Abc_NtkPiNum(src) || count > Abc_NtkPoNum(src) {
		panic("abc: invalid terminal latch conversion")
	}
	dst := ntkNewLike(src, kind, function)
	ntkClearCopies(src)
	setCopy(src.constant, dst.One())
	pis, pos := src.ObjectsOfKind(ObjectPI), src.ObjectsOfKind(ObjectPO)
	names := int32(1)
	if kind == ABC_NTK_NETLIST {
		names = 0
	}
	for _, o := range pis[:len(pis)-int(count)] {
		Abc_NtkDupObj(dst, o, names)
	}
	for _, o := range pos[:len(pos)-int(count)] {
		Abc_NtkDupObj(dst, o, names)
	}
	for i := 0; i < int(count); i++ {
		latch, in, out := dst.NewObject(ObjectLatch), dst.NewObject(ObjectBI), dst.NewObject(ObjectBO)
		Abc_LatchSetInit0(latch)
		setCopy(pos[len(pos)-int(count)+i], in.Signal())
		setCopy(pis[len(pis)-int(count)+i], out.Signal())
		Abc_ObjAddFanin(latch, in)
		Abc_ObjAddFanin(out, latch)
	}
	Abc_NtkOrderCisCos(dst)
	ntkCopyTiming(dst, src)
	return dst
}
func Abc_NtkStartFromNoLatches(src *Network, kind Abc_NtkType_t, function Abc_NtkFunc_t) *Network {
	if src == nil {
		return nil
	}
	if kind == ABC_NTK_NETLIST {
		panic("abc: no-latch start requires logic representation")
	}
	dst := ntkFreshLike(src, kind, function)
	dst.Name, dst.Spec = src.Name, src.Spec
	dst.Constraints, dst.Barriers = src.Constraints, src.Barriers
	dst.ObjectPermutation = append([]int32(nil), src.ObjectPermutation...)
	dst.AndGateDelay = src.AndGateDelay
	ntkClearCopies(src)
	setCopy(src.constant, dst.One())
	for _, k := range []ObjectKind{ObjectPI, ObjectPO} {
		for _, o := range src.ObjectsOfKind(k) {
			Abc_NtkDupObj(dst, o, 1)
		}
	}
	for _, box := range src.boxes {
		if box.Kind != ObjectLatch {
			Abc_NtkDupBox(dst, box, 1)
		}
	}
	Abc_NtkOrderCisCos(dst)
	ntkCopyTiming(dst, src)
	return dst
}
func Abc_NtkFinalize(src, dst *Network) {
	for _, o := range src.cos {
		if len(o.fanins) != 1 {
			panic("abc: output has no driver")
		}
		e := o.fanins[0]
		if e.Object.Kind == ObjectNet {
			inner := e.Object.fanins[0]
			inner.Complement = inner.Complement != e.Complement
			e = inner
		}
		Abc_ObjAddFanin(getCopy(o).Object, ntkCopiedEdge(e))
	}
	ntkCopyTiming(dst, src)
}
func Abc_NtkStartRead(name string) *Network {
	n := Abc_NtkAlloc(ABC_NTK_NETLIST, ABC_FUNC_SOP, 1)
	n.Spec = name
	n.Name = name
	if dot := strings.LastIndexByte(name, '.'); dot >= 0 {
		n.Name = name[:dot]
	}
	if n.Name == "" || n.Name == "." {
		n.Name = "unknown"
	}
	return n
}
func Abc_NtkFinalizeRead(n *Network) {
	if Abc_NtkHasBlackbox(n) != 0 && len(n.boxes) == 0 {
		box := n.NewObject(ObjectBlackbox)
		for _, pi := range n.ObjectsOfKind(ObjectPI) {
			term := n.NewObject(ObjectBI)
			Abc_ObjAddFanin(term, pi.fanouts[0])
			Abc_ObjAddFanin(box, term)
		}
		for _, po := range n.ObjectsOfKind(ObjectPO) {
			term := n.NewObject(ObjectBO)
			Abc_ObjAddFanin(term, box)
			Abc_ObjAddFanin(po.fanins[0].Object, term)
		}
		return
	}
	if Abc_NtkIsNetlist(n) == 0 {
		panic("abc: finalize read requires netlist")
	}
	for _, test := range []struct {
		name string
		one  bool
	}{{"1'b0", false}, {"1'b1", true}} {
		net := Abc_NtkFindNet(n, test.name)
		if net == nil {
			continue
		}
		if len(net.fanouts) == 0 {
			Abc_NtkDeleteObj(net)
		} else if len(net.fanins) == 0 {
			var node *Object
			if test.one {
				node = Abc_NtkCreateNodeConst1(n)
			} else {
				node = Abc_NtkCreateNodeConst0(n)
			}
			Abc_ObjAddFanin(net, node)
		}
	}
	Abc_NtkFixNonDrivenNets(n)
	Abc_NtkOrderCisCos(n)
}

func ntkDuplicate(src *Network, dfs, skipBarriers bool) *Network {
	if src == nil {
		return nil
	}
	if dfs && (Abc_NtkIsStrash(src) != 0 || Abc_NtkIsNetlist(src) != 0) {
		panic("abc: DFS duplication needs logic")
	}
	dst := Abc_NtkStartFrom(src, Abc_NtkType_t(src.Type), Abc_NtkFunc_t(src.Function))
	var order []*Object
	if dfs {
		order = Abc_NtkDfs(src, 0)
	} else if Abc_NtkIsStrash(src) != 0 {
		var err error
		order, err = src.Topological(src.ObjectsOfKind(ObjectNode))
		if err != nil {
			panic(err)
		}
	} else {
		order = src.Objects()
	}
	for _, o := range order {
		if o == src.constant || getCopy(o).Object != nil {
			continue
		}
		if skipBarriers && isBarrierBuffer(o) {
			setCopy(o, ntkCopiedEdge(o.fanins[0]))
			continue
		}
		if Abc_NtkIsStrash(src) != 0 && o.Kind == ObjectNode {
			if len(o.fanins) != 2 {
				panic("abc: malformed AIG node")
			}
			setCopy(o, dst.AndEdges(ntkCopiedEdge(o.fanins[0]), ntkCopiedEdge(o.fanins[1])))
		} else {
			copy := Abc_NtkDupObj(dst, o, abcBool(Abc_NtkHasBlackbox(src) != 0 && o.Kind == ObjectNet))
			copy.Barrier = o.Barrier
			copy.TerminalIndex = o.TerminalIndex
		}
	}
	for _, o := range src.Objects() {
		copy := getCopy(o)
		if copy.Object == nil || o == src.constant || o.Kind >= ObjectLatch || o.Kind == ObjectBO || skipBarriers && isBarrierBuffer(o) || Abc_NtkIsStrash(src) != 0 && o.Kind == ObjectNode {
			continue
		}
		for _, e := range o.fanins {
			if getCopy(e.Object).Object != nil {
				Abc_ObjAddFanin(copy.Object, ntkCopiedEdge(e))
			}
		}
	}
	for _, o := range src.Objects() {
		if o.Choice != nil && getCopy(o).Object != nil {
			getCopy(o).Object.Choice = getCopy(o.Choice).Object
		}
	}
	if src.Exdc != nil {
		dst.Exdc = Abc_NtkDup(src.Exdc)
	}
	if src.Excare != nil {
		dst.Excare = Abc_NtkDup(src.Excare)
	}
	for _, o := range src.Objects() {
		if copy := getCopy(o); copy.Object != nil {
			setCopy(o, copy)
		}
	}
	ntkCopyTiming(dst, src)
	src.Copy = dst
	return dst
}
func Abc_NtkDup(src *Network) *Network    { return ntkDuplicate(src, false, false) }
func Abc_NtkDupDfs(src *Network) *Network { return ntkDuplicate(src, true, false) }
func Abc_NtkDupDfsNoBarBufs(src *Network) *Network {
	dst := ntkDuplicate(src, true, true)
	if dst != nil {
		dst.Barriers = 0
	}
	return dst
}
func Abc_NtkRestrashWithLatches(src *Network, count int32) *Network {
	if Abc_NtkIsStrash(src) == 0 {
		panic("abc: expected strashed network")
	}
	dst := Abc_NtkStartFromWithLatches(src, ABC_NTK_STRASH, ABC_FUNC_AIG, count)
	order, err := src.Topological(src.ObjectsOfKind(ObjectNode))
	if err != nil {
		panic(err)
	}
	for _, o := range order {
		if o.Kind == ObjectNode {
			setCopy(o, dst.AndEdges(ntkCopiedEdge(o.fanins[0]), ntkCopiedEdge(o.fanins[1])))
		}
	}
	Abc_NtkFinalize(src, dst)
	return dst
}
func Abc_NtkDupTransformMiter(src *Network) *Network {
	if Abc_NtkIsStrash(src) == 0 || Abc_NtkPoNum(src)%2 != 0 {
		panic("abc: paired strashed outputs required")
	}
	dst := ntkFreshLike(src, ABC_NTK_STRASH, ABC_FUNC_AIG)
	dst.Name, dst.Spec = src.Name, src.Spec
	dst.Constraints, dst.Barriers = src.Constraints, src.Barriers
	ntkClearCopies(src)
	setCopy(src.constant, dst.One())
	for _, pi := range src.ObjectsOfKind(ObjectPI) {
		Abc_NtkDupObj(dst, pi, 1)
	}
	pos := src.ObjectsOfKind(ObjectPO)
	for i := 0; i < len(pos); i += 2 {
		Abc_NtkDupObj(dst, pos[i], 1)
	}
	for _, box := range src.boxes {
		Abc_NtkDupBox(dst, box, 1)
	}
	order, err := src.Topological(src.ObjectsOfKind(ObjectNode))
	if err != nil {
		panic(err)
	}
	for _, o := range order {
		if o.Kind == ObjectNode {
			setCopy(o, dst.AndEdges(ntkCopiedEdge(o.fanins[0]), ntkCopiedEdge(o.fanins[1])))
		}
	}
	for i := 0; i < len(pos); i += 2 {
		a, b := ntkCopiedEdge(pos[i].fanins[0]), ntkCopiedEdge(pos[i+1].fanins[0])
		Abc_ObjAddFanin(pos[i].Copy, ntkXor(dst, a, b))
	}
	for _, o := range src.ObjectsOfKind(ObjectBI) {
		Abc_ObjAddFanin(o.Copy, ntkCopiedEdge(o.fanins[0]))
	}
	Abc_AigCleanup(dst.FunctionManager.(*Abc_Aig_t))
	return dst
}
func ntkXor(n *Network, a, b Edge) Edge {
	return n.AndEdges(n.AndEdges(a, b.Not()).Not(), n.AndEdges(a.Not(), b).Not()).Not()
}
func Abc_NtkDelete(n *Network) {
	if n == nil {
		return
	}
	for _, o := range n.Objects() {
		if o.MarkA || o.MarkB || o.MarkC {
			panic(fmt.Sprintf("abc: deleting marked object %d", o.ID))
		}
	}
	if n.Exdc != nil {
		Abc_NtkDelete(n.Exdc)
	}
	if n.Excare != nil {
		Abc_NtkDelete(n.Excare)
	}
	for i, attribute := range n.Attributes {
		if attribute != nil {
			Abc_NtkAttrFree(n, int32(i), 1)
		}
	}
	switch manager := n.FunctionManager.(type) {
	case *Abc_Aig_t:
		Abc_AigFree(manager)
	case *core.DdManager:
		for _, o := range n.kindOrder[ObjectNode] {
			if node, ok := o.Data.(*core.DdNode); ok {
				core.Cudd_RecursiveDeref(manager, node)
			}
		}
		core.Cudd_Quit(manager)
	case *hop.Hop_Man_t:
		hop.Hop_ManStop(manager)
	case *mem.Mem_Flex_t:
		mem.Mem_FlexStop(manager, 0)
	}
	for _, o := range n.Objects() {
		o.network = nil
		o.fanins = nil
		o.fanouts = nil
	}
	*n = Network{}
}
