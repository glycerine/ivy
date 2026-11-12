package abc

// Native counterparts of abcHieGia.c. Graph edges use the AIG package's
// const-one literal 0 convention, rather than GIA's const-zero convention.
import (
	"fmt"
	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
)

// Boundary buffers survive sweeping even when their logical value becomes a
// constant or another boundary. Signals records that value separately from
// Buffers, which keeps the mapper-to-hierarchy port correspondence stable.
type HierarchyBufferInfo struct {
	Ports            []*Object
	Buffers, Signals []aig.Edge
	Groups           []HierarchyBufferGroup
}
type HierarchyBufferGroup struct {
	InputStart, InputCount, OutputStart, OutputCount int
	Instance                                         *Object
}

func hierarchyBuffer(g *aig.Graph, driver aig.Edge) aig.Edge {
	return aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{driver, 0}})
}
func (b *HierarchyBufferInfo) append(g *aig.Graph, port *Object, driver aig.Edge) aig.Edge {
	edge := hierarchyBuffer(g, driver)
	b.Ports = append(b.Ports, port)
	b.Buffers = append(b.Buffers, edge)
	b.Signals = append(b.Signals, edge)
	return edge
}
func Abc_NodeStrashToGia_rec(g *aig.Graph, o *hop.Hop_Obj_t) { Abc_ConvertAigToGia_rec1(g, o) }
func Abc_NodeStrashToGia(g *aig.Graph, o *Object) aig.Edge {
	args := make([]aig.Edge, len(o.fanins))
	for i, e := range o.fanins {
		if e.Object.Temp < 0 {
			panic("unassigned hierarchy fanin")
		}
		args[i] = aig.Edge(e.Object.Temp)
		if e.Complement {
			args[i] = args[i].Not()
		}
	}
	return localFunctionAIG(g, o, args)
}

// This variant visits all instantiated ports, including unused outputs, and
// therefore accepts reused models. Each expansion owns a separate signal map.
func Abc_NtkFlattenHierarchyGia2_rec(g *aig.Graph, n *Network, counter *int32, groups *[]int32) {
	info, _ := g.Data.(*HierarchyBufferInfo)
	if info == nil {
		info = &HierarchyBufferInfo{}
		g.Data = info
	}
	c := &hierarchyAIG{cached: map[*Network]*aig.Graph{}, active: map[*Network]bool{}, orders: map[*Network][]*Object{}}
	var stack []HierarchyBufferGroup
	c.enterBox = func(g *aig.Graph, box *Object, args []aig.Edge) []aig.Edge {
		model := boxModel(box)
		group := HierarchyBufferGroup{InputStart: len(info.Buffers), InputCount: len(args), Instance: box}
		stack = append(stack, group)
		if groups != nil {
			for i, pi := range model.ObjectsOfKind(ObjectPI) {
				args[i] = info.append(g, pi, args[i])
			}
		}
		if counter != nil {
			(*counter)++
		}
		return args
	}
	c.leaveBox = func(g *aig.Graph, box *Object, outputs []aig.Edge) []aig.Edge {
		group := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		group.OutputStart = len(info.Buffers)
		group.OutputCount = len(outputs)
		if groups != nil {
			for i, po := range boxModel(box).ObjectsOfKind(ObjectPO) {
				outputs[i] = info.append(g, po, outputs[i])
			}
			info.Groups = append(info.Groups, group)
			*groups = append(*groups, int32(group.InputStart), int32(group.InputCount), int32(group.OutputStart), int32(group.OutputCount))
		}
		return outputs
	}
	inputs := make([]aig.Edge, 0)
	for _, pi := range n.ObjectsOfKind(ObjectPI) {
		if pi.Temp < 0 {
			panic("unassigned root input")
		}
		inputs = append(inputs, aig.Edge(pi.Temp))
	}
	outputs, err := c.expand(g, n, inputs)
	if err != nil {
		panic(err)
	}
	for i, po := range n.ObjectsOfKind(ObjectPO) {
		po.Temp = int(outputs[i])
	}
}
func Abc_NtkFlattenHierarchyGia2(n *Network) *aig.Graph {
	if !hieValidateModels(n) {
		panic("invalid hierarchy")
	}
	g := aig.NewGraph()
	g.Name, g.Spec = n.Name, n.Spec
	for _, pi := range n.ObjectsOfKind(ObjectPI) {
		pi.Temp = int(g.NewInput())
	}
	var counter int32 = 1
	var groups []int32
	Abc_NtkFlattenHierarchyGia2_rec(g, n, &counter, &groups)
	for _, po := range n.ObjectsOfKind(ObjectPO) {
		g.AddOutput(aig.Edge(po.Temp))
	}
	return hierarchySweep(g, false)
}

func Gia_ManPrintBarBufDrivers(g *aig.Graph) {
	nodes := g.Nodes()
	seen := map[int]bool{}
	constants, shared, liveConstants, liveShared := 0, 0, 0, 0
	for _, edge := range g.LiveEdges() {
		node := nodes[edge.NodeID()]
		if node.Kind != aig.Buffer {
			continue
		}
		driver := node.Fanins[0].NodeID()
		if driver == 0 {
			constants++
			if node.Refs > 0 {
				liveConstants++
			}
		} else if seen[driver] {
			shared++
			if node.Refs > 0 {
				liveShared++
			}
		} else {
			seen[driver] = true
		}
	}
	fmt.Printf("Barbuf drivers: %d constant (%d used), %d shared (%d used).\n", constants, liveConstants, shared, liveShared)
}
func Gia_ManPatchBufDriver(g *aig.Graph, id int, driver aig.Edge) {
	edge := aig.Edge(id) << 1
	if !g.ValidEdge(edge) || !g.ValidEdge(driver) || driver.NodeID() >= id || g.Node(edge).Kind != aig.Buffer {
		panic("invalid buffer patch")
	}
	seen := map[int]bool{}
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		if e.NodeID() == id {
			panic("buffer patch would create a cycle")
		}
		if seen[e.NodeID()] {
			return
		}
		seen[e.NodeID()] = true
		node := g.Node(e)
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			visit(node.Fanins[1])
			fallthrough
		case aig.Buffer, aig.Output:
			visit(node.Fanins[0])
		}
	}
	visit(driver)
	aig.Aig_ObjConnect(g, edge, driver, aig.InvalidEdge)
}

// Cleanup roots include every physical buffer, as in Gia_ManCleanup. Ordinary
// AIG cleanup roots only outputs and would discard unused boundary identities.
func hierarchySweep(old *aig.Graph, simplify bool) *aig.Graph {
	g := aig.NewGraph()
	g.Name, g.Spec = old.Name, old.Spec
	g.Data = old.Data
	nodes := old.Nodes()
	copies := map[int]aig.Edge{0: 0}
	physical := map[int]aig.Edge{}
	for _, input := range old.Inputs() {
		copies[input.NodeID()] = g.NewInput()
	}
	type representative struct{ buffer, driver aig.Edge }
	first := map[int]representative{}
	active := map[int]bool{}
	var visit func(aig.Edge) aig.Edge
	visit = func(edge aig.Edge) aig.Edge {
		id := edge.NodeID()
		value, ok := copies[id]
		if !ok {
			if active[id] {
				panic("cyclic hierarchy AIG")
			}
			active[id] = true
			node := nodes[id]
			switch node.Kind {
			case aig.And:
				value = g.AND(visit(node.Fanins[0]), visit(node.Fanins[1]))
			case aig.ExclusiveOr:
				value = g.NewXor(visit(node.Fanins[0]), visit(node.Fanins[1]))
			case aig.Output:
				value = visit(node.Fanins[0])
			case aig.Buffer:
				driver := visit(node.Fanins[0])
				value = hierarchyBuffer(g, driver)
				physical[id] = value
				if simplify {
					if driver.NodeID() == 0 {
						Gia_ManPatchBufDriver(g, value.NodeID(), 1)
						value = driver
					} else if previous, exists := first[driver.NodeID()]; exists {
						Gia_ManPatchBufDriver(g, value.NodeID(), 1)
						value = previous.buffer
						if previous.driver.Complemented() != driver.Complemented() {
							value = value.Not()
						}
					} else {
						first[driver.NodeID()] = representative{value, driver}
					}
				}
			default:
				panic("unsupported hierarchy AIG node")
			}
			copies[id] = value
			delete(active, id)
		}
		if edge.Complemented() {
			return value.Not()
		}
		return value
	}
	for _, edge := range old.LiveEdges() {
		if nodes[edge.NodeID()].Kind == aig.Buffer {
			visit(edge)
		}
	}
	for _, output := range old.Outputs() {
		g.AddOutput(visit(output))
	}
	aig.Aig_ManSetRegNum(g, old.RegisterCount())
	if info, ok := old.Data.(*HierarchyBufferInfo); ok {
		clone := &HierarchyBufferInfo{Ports: append([]*Object(nil), info.Ports...), Groups: append([]HierarchyBufferGroup(nil), info.Groups...)}
		for i, buffer := range info.Buffers {
			clone.Buffers = append(clone.Buffers, physical[buffer.NodeID()])
			clone.Signals = append(clone.Signals, visit(info.Signals[i]))
		}
		g.Data = clone
	}
	if err := g.Check(); err != nil {
		panic(err)
	}
	return g
}
func Gia_ManSweepHierarchy(g *aig.Graph) *aig.Graph { return hierarchySweep(g, true) }

func Gia_ManFlattenLogicPrepare(n *Network) int32 {
	fillHierarchyTemps(n)
	for i, o := range n.ObjectsOfKind(ObjectPI) {
		o.TerminalIndex = i
	}
	for i, o := range n.ObjectsOfKind(ObjectPO) {
		o.TerminalIndex = i
	}
	for _, box := range n.Boxes() {
		if box.Kind == ObjectLatch {
			panic("sequential hierarchy is not supported")
		}
		for i, e := range box.fanins {
			e.Object.TerminalIndex = i
		}
		for i, o := range box.fanouts {
			o.TerminalIndex = i
		}
	}
	return int32(len(n.ObjectsOfKind(ObjectPI)) + len(n.ObjectsOfKind(ObjectPO)))
}
func Abc_NtkFlattenHierarchyGia_rec(g *aig.Graph, supers *[]*Object, o *Object, buffers *[]*Object) aig.Edge {
	if o.Temp >= 0 {
		return aig.Edge(o.Temp)
	}
	// -2 distinguishes a recursion cycle from an unvisited object.
	if o.Temp == -2 {
		panic("cyclic hierarchy net")
	}
	o.Temp = -2
	read := func(e Edge) aig.Edge {
		value := Abc_NtkFlattenHierarchyGia_rec(g, supers, e.Object, buffers)
		if e.Complement {
			value = value.Not()
		}
		return value
	}
	var value aig.Edge
	switch o.Kind {
	case ObjectConst1:
		value = 0
	case ObjectNet, ObjectPO, ObjectBI:
		if len(o.fanins) != 1 {
			panic("hierarchy wire must have one driver")
		}
		value = read(o.fanins[0])
	case ObjectPI:
		if len(*supers) == 0 {
			panic("unassigned root input")
		}
		box := (*supers)[len(*supers)-1]
		*supers = (*supers)[:len(*supers)-1]
		value = read(box.fanins[o.TerminalIndex])
		*supers = append(*supers, box)
		if buffers != nil {
			value = g.Data.(*HierarchyBufferInfo).append(g, o, value)
			*buffers = append(*buffers, o)
		}
	case ObjectBO:
		if len(o.fanins) != 1 {
			panic("box output must have one box")
		}
		box := o.fanins[0].Object
		*supers = append(*supers, box)
		po := boxModel(box).ObjectsOfKind(ObjectPO)[o.TerminalIndex]
		value = Abc_NtkFlattenHierarchyGia_rec(g, supers, po, buffers)
		*supers = (*supers)[:len(*supers)-1]
		if buffers != nil {
			value = g.Data.(*HierarchyBufferInfo).append(g, po, value)
			*buffers = append(*buffers, po)
		}
		if o.fanins[0].Complement {
			value = value.Not()
		}
	case ObjectNode:
		args := make([]aig.Edge, len(o.fanins))
		for i, e := range o.fanins {
			args[i] = read(e)
		}
		if o.Barrier {
			if len(args) != 1 {
				panic("barrier arity")
			}
			value = args[0]
		} else {
			value = localFunctionAIG(g, o, args)
		}
	default:
		panic("unsupported hierarchical object")
	}
	o.Temp = int(value)
	return value
}
func Abc_NtkFlattenHierarchyGia(n *Network, buffers *[]*Object, verbose int32) *aig.Graph {
	if !hieValidateModels(n) {
		panic("invalid hierarchy")
	}
	models := Abc_NtkCollectHie(n)
	uses := map[*Network]int{}
	for _, model := range models {
		Gia_ManFlattenLogicPrepare(model)
		for _, box := range model.Boxes() {
			uses[boxModel(box)]++
		}
	}
	for _, count := range uses {
		if count > 1 {
			panic("demand-driven hierarchy flattening requires unique models")
		}
	}
	g := aig.NewGraph()
	g.Name, g.Spec = n.Name, n.Spec
	g.Data = &HierarchyBufferInfo{}
	for _, pi := range n.ObjectsOfKind(ObjectPI) {
		pi.Temp = int(g.NewInput())
	}
	var supers, ports []*Object
	for _, po := range n.ObjectsOfKind(ObjectPO) {
		g.AddOutput(Abc_NtkFlattenHierarchyGia_rec(g, &supers, po, &ports))
	}
	if buffers != nil {
		*buffers = ports
	}
	if verbose != 0 {
		fmt.Printf("Hierarchy reader added %d boundary buffers.\n", len(ports))
	}
	return Gia_ManSweepHierarchy(g)
}

// insertHierarchyCone copies a mapped cone into one module. The bindings cut
// traversal at module boundaries; the cache is local to that module so shared
// flat nodes never accidentally connect two different hierarchy instances.
func insertHierarchyCone(dst, flat *Network, bindings map[*Object]Edge) func(Edge) Edge {
	active := map[*Object]bool{}
	var copyEdge func(Edge) Edge
	copyEdge = func(edge Edge) Edge {
		o := edge.Object
		if o == nil {
			return netForDriver(dst, dst.One().Not()).Signal()
		}
		value, ok := bindings[o]
		if !ok {
			if o.network != flat {
				panic("foreign mapped cone")
			}
			if active[o] {
				panic("cyclic mapped cone")
			}
			active[o] = true
			switch {
			case o.Kind == ObjectConst1:
				value = netForDriver(dst, dst.One()).Signal()
			case o.Kind == ObjectPI || o.Kind == ObjectBO:
				panic("mapped cone crosses an unbound module input")
			case o.Kind == ObjectNet || o.Kind == ObjectPO || o.Kind == ObjectBI || o.Barrier:
				if len(o.fanins) != 1 {
					panic("mapped wire arity")
				}
				value = copyEdge(o.fanins[0])
			case o.Kind == ObjectNode:
				args := make([]Edge, len(o.fanins))
				for i, e := range o.fanins {
					args[i] = copyEdge(e)
				}
				node := Abc_NtkDupObj(dst, o, 0)
				for _, arg := range args {
					Abc_ObjAddFanin(node, arg)
				}
				value = netForDriver(dst, node.Signal()).Signal()
			default:
				panic("unsupported mapped cone object")
			}
			delete(active, o)
			bindings[o] = value
		}
		if edge.Complement {
			value = value.Not()
		}
		return value
	}
	return copyEdge
}
func Gia_ManInsertOne_rec(dst, flat *Network, o *Object) Edge {
	bindings := map[*Object]Edge{}
	for _, obj := range flat.Objects() {
		if e := getCopy(obj); e.Object != nil && e.Object.network == dst {
			bindings[obj] = e
		}
	}
	result := insertHierarchyCone(dst, flat, bindings)(Edge{Object: o})
	for obj, edge := range bindings {
		setCopy(obj, edge)
	}
	return result
}

func rebuildHierarchyModule(old, flat *Network, portSignals map[*Object]Edge) *Network {
	// Capture boundary bindings before cloning rewrites the old Copy fields.
	dst := cloneInterfaces(old)
	dst.Function = uint8(ABC_FUNC_MAP)
	dst.FunctionManager = flat.FunctionManager
	bindings := map[*Object]Edge{}
	bind := func(global, local Edge) {
		if global.Object == nil || global.Object == flat.constant {
			return
		}
		if global.Complement {
			local = local.Not()
		}
		if _, exists := bindings[global.Object]; !exists {
			bindings[global.Object] = local
		}
	}
	for _, pi := range old.ObjectsOfKind(ObjectPI) {
		copied := getCopy(pi).Object
		net := netForDriver(dst, copied.Signal())
		if len(pi.fanouts) > 0 {
			net.SetName(pi.fanouts[0].Name())
		}
		bind(portSignals[pi], net.Signal())
	}
	for _, box := range old.Boxes() {
		model := boxModel(box)
		for i, bo := range box.fanouts {
			copied := getCopy(bo).Object
			net := netForDriver(dst, copied.Signal())
			if len(bo.fanouts) > 0 {
				net.SetName(bo.fanouts[0].Name())
			}
			bind(portSignals[model.ObjectsOfKind(ObjectPO)[i]], net.Signal())
		}
	}
	copyEdge := insertHierarchyCone(dst, flat, bindings)
	for _, po := range old.ObjectsOfKind(ObjectPO) {
		Abc_ObjAddFanin(getCopy(po).Object, copyEdge(portSignals[po]))
	}
	for _, box := range old.Boxes() {
		model := boxModel(box)
		for i, e := range box.fanins {
			Abc_ObjAddFanin(getCopy(e.Object).Object, copyEdge(portSignals[model.ObjectsOfKind(ObjectPI)[i]]))
		}
	}
	hieOrderTerminals(dst)
	if err := dst.Check(); err != nil {
		panic(err)
	}
	return dst
}
func commitHierarchyModule(old, fresh *Network) {
	design := old.Design
	*old = *fresh
	old.Design = design
	for _, object := range old.Objects() {
		object.network = old
	}
}
func Gia_ManInsertOne(old, flat *Network) {
	if flat.Function != uint8(ABC_FUNC_MAP) {
		panic("hierarchy insertion needs a mapped network")
	}
	signals := map[*Object]Edge{}
	for _, o := range old.ObjectsOfKind(ObjectPI) {
		signals[o] = getCopy(o)
	}
	for _, o := range old.ObjectsOfKind(ObjectPO) {
		signals[o] = getCopy(o)
	}
	for _, box := range old.Boxes() {
		for _, o := range boxModel(box).Objects() {
			if o.Kind == ObjectPI || o.Kind == ObjectPO {
				signals[o] = getCopy(o)
			}
		}
	}
	commitHierarchyModule(old, rebuildHierarchyModule(old, flat, signals))
}
func Abc_NtkInsertHierarchyGia(old, flat *Network, verbose int32) {
	if flat.Function != uint8(ABC_FUNC_MAP) {
		panic("hierarchy insertion needs a mapped network")
	}
	graph := Abc_NtkFlattenHierarchyGia(old, nil, 0)
	info := graph.Data.(*HierarchyBufferInfo)
	pis, pos := old.ObjectsOfKind(ObjectPI), old.ObjectsOfKind(ObjectPO)
	flatPI, flatPO := flat.ObjectsOfKind(ObjectPI), flat.ObjectsOfKind(ObjectPO)
	if len(pis) != len(flatPI) || len(pos) != len(flatPO) {
		panic("mapped hierarchy interface mismatch")
	}
	var barriers []*Object
	for _, o := range flat.ObjectsOfKind(ObjectNode) {
		if o.Barrier {
			barriers = append(barriers, o)
		}
	}
	if len(barriers) != len(info.Buffers) {
		panic("mapped hierarchy boundary count mismatch")
	}
	bufferIndex := map[int]int{}
	for i, buffer := range info.Buffers {
		bufferIndex[buffer.NodeID()] = i
	}
	signals := map[*Object]Edge{}
	for i, pi := range pis {
		signals[pi] = flatPI[i].Signal()
	}
	for i, po := range pos {
		signals[po] = flatPO[i].Signal()
	}
	for i, port := range info.Ports {
		signal := info.Signals[i]
		value := flat.One()
		if signal.NodeID() != 0 {
			index, ok := bufferIndex[signal.NodeID()]
			if !ok {
				panic("boundary signal lost during sweep")
			}
			value = barriers[index].Signal()
		}
		if signal.Complemented() {
			value = value.Not()
		}
		signals[port] = value
	}
	models := Abc_NtkCollectHie(old)
	fresh := make([]*Network, len(models))
	// Build every module before replacing any of them: model pointers continue to
	// refer to their original identities throughout this transaction.
	for i, model := range models {
		fresh[i] = rebuildHierarchyModule(model, flat, signals)
	}
	for i, model := range models {
		commitHierarchyModule(model, fresh[i])
	}
	if verbose != 0 {
		fmt.Printf("Reinserted mapped logic into %d hierarchy modules.\n", len(models))
	}
}
