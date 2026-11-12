package abc

import "fmt"

func hierarchyModules(root *Network) []*Network {
	var modules []*Network
	seen := map[*Network]bool{}
	var walk func(*Network)
	walk = func(n *Network) {
		if seen[n] {
			return
		}
		seen[n] = true
		modules = append(modules, n)
		for _, box := range n.Boxes() {
			if model, ok := box.Data.(*Network); ok && model != nil {
				walk(model)
			}
		}
	}
	walk(root)
	return modules
}
func Abc_NtkCheckSingleInstance(root *Network) int32 {
	if root.Design == nil {
		return 1
	}
	modules := root.Design.Modules
	uses := map[*Network]int{}
	known := map[*Network]bool{}
	for _, n := range modules {
		known[n] = true
	}
	if !known[root] {
		return 0
	}
	for _, n := range modules {
		for _, box := range n.Boxes() {
			model, ok := box.Data.(*Network)
			if !ok || model == nil {
				return 0
			}
			if !known[model] || len(model.ObjectsOfKind(ObjectLatch)) > 0 {
				return 0
			}
			uses[model]++
			if uses[model] > 1 || len(box.fanins) != len(model.ObjectsOfKind(ObjectPI)) || len(box.fanouts) != len(model.ObjectsOfKind(ObjectPO)) {
				return 0
			}
		}
	}
	state := map[*Network]uint8{}
	var visit func(*Network) bool
	visit = func(n *Network) bool {
		if state[n] == 1 {
			return false
		}
		if state[n] == 2 {
			return true
		}
		state[n] = 1
		for _, box := range n.Boxes() {
			if !visit(boxModel(box)) {
				return false
			}
		}
		state[n] = 2
		return true
	}
	for _, n := range modules {
		if !visit(n) {
			return 0
		}
	}
	return 1
}
func Abc_NtkCollectPiPos_rec(net *Object, li, lo *[]*Object) int32 {
	if net.Kind != ObjectNet || len(net.fanins) != 1 {
		panic("hierarchy traversal requires driven net")
	}
	if Abc_NodeIsTravIdCurrent(net) != 0 {
		return 0
	}
	Abc_NodeSetTravIdCurrent(net)
	driver := net.fanins[0].Object
	count := int32(0)
	if driver.Kind == ObjectNode {
		for _, e := range driver.fanins {
			count += Abc_NtkCollectPiPos_rec(e.Object, li, lo)
		}
		return count
	}
	if driver.Kind == ObjectConst1 {
		return 0
	}
	if driver.Kind == ObjectBO {
		if len(driver.fanins) != 1 {
			panic("box output has no box")
		}
		driver = driver.fanins[0].Object
	}
	if Abc_ObjIsBox(driver) == 0 {
		panic("unmarked primary input net")
	}
	for _, out := range driver.fanouts {
		if len(out.fanouts) != 1 {
			panic("box output must drive one net")
		}
		Abc_NodeSetTravIdCurrent(out.fanouts[0])
	}
	for _, e := range driver.fanins {
		if len(e.Object.fanins) != 1 {
			panic("box input requires net")
		}
		count += Abc_NtkCollectPiPos_rec(e.Object.fanins[0].Object, li, lo)
	}
	return count + Abc_NtkCollectPiPos_int(driver, boxModel(driver), li, lo)
}
func Abc_NtkCollectPiPos_int(box *Object, n *Network, li, lo *[]*Object) int32 {
	Abc_NtkIncrementTravId(n)
	for _, pi := range n.ObjectsOfKind(ObjectPI) {
		for _, net := range pi.fanouts {
			if net.Kind == ObjectNet {
				Abc_NodeSetTravIdCurrent(net)
			}
		}
	}
	if box != nil {
		for _, e := range box.fanins {
			*li = append(*li, e.Object)
		}
		*lo = append(*lo, n.ObjectsOfKind(ObjectPI)...)
	}
	count := int32(0)
	for _, po := range n.ObjectsOfKind(ObjectPO) {
		if len(po.fanins) != 1 {
			panic("PO requires net")
		}
		count += Abc_NtkCollectPiPos_rec(po.fanins[0].Object, li, lo)
	}
	if box != nil {
		*li = append(*li, n.ObjectsOfKind(ObjectPO)...)
		*lo = append(*lo, box.fanouts...)
		count++
	}
	return count
}
func Abc_NtkCollectPiPos(n *Network, li, lo *[]*Object) int32 {
	if li == nil || lo == nil {
		panic("missing boundary maps")
	}
	*li = nil
	*lo = nil
	return Abc_NtkCollectPiPos_int(nil, n, li, lo)
}
func toBarSignal(n *Network, net *Object) Edge {
	if e := getCopy(net); e.Object != nil {
		return e
	}
	if net.Kind != ObjectNet || len(net.fanins) != 1 {
		panic("flattening requires driven net")
	}
	driver := net.fanins[0]
	o := driver.Object
	if o.Kind == ObjectConst1 {
		e := n.One()
		e.Complement = driver.Complement
		setCopy(net, e)
		return e
	}
	if o.Kind != ObjectNode {
		panic("hierarchy boundary has no copy mapping")
	}
	dup := Abc_NtkDupObj(n, o, 0)
	result := dup.Signal()
	result.Complement = driver.Complement
	setCopy(net, result)
	for _, e := range o.fanins {
		fanin := toBarSignal(n, e.Object)
		fanin.Complement = fanin.Complement != e.Complement
		Abc_ObjAddFanin(dup, fanin)
	}
	return result
}
func Abc_NtkToBarBufs_rec(n *Network, net *Object) Edge { return toBarSignal(n, net) }
func Abc_NtkToBarBufs(root *Network) *Network {
	if root.Design == nil {
		panic("flattening requires a design")
	}
	if Abc_NtkCheckSingleInstance(root) == 0 {
		return nil
	}
	cleanHierarchyCopies(root)
	n := copyNetworkMetadata(root)
	n.Type = uint8(ABC_NTK_LOGIC)
	setCopy(root.constant, n.One())
	for _, pi := range root.ObjectsOfKind(ObjectPI) {
		dup := Abc_NtkDupObj(n, pi, 1)
		for _, net := range pi.fanouts {
			if net.Kind == ObjectNet {
				setCopy(net, dup.Signal())
			}
		}
	}
	for _, po := range root.ObjectsOfKind(ObjectPO) {
		Abc_NtkDupObj(n, po, 1)
	}
	var li, lo []*Object
	count := Abc_NtkCollectPiPos(root, &li, &lo)
	if len(li) != len(lo) {
		panic("unbalanced boundary mapping")
	}
	for i, from := range li {
		to := lo[i]
		input, _, output := newBarrier(n, ABC_INIT_ZERO)
		var internal *Network
		if len(from.fanins) == 1 && from.fanins[0].Object.network != root {
			internal = from.fanins[0].Object.network
		} else if len(to.fanouts) == 1 && to.fanouts[0].network != root {
			internal = to.fanouts[0].network
		} else {
			panic("boundary does not cross root module")
		}
		input.SetName(fmt.Sprintf("%s_%s_in", from.fanins[0].Object.Name(), internal.Name))
		output.SetName(fmt.Sprintf("%s_%s_out", to.fanouts[0].Name(), internal.Name))
		setCopy(from, input.Signal())
		if len(to.fanouts) != 1 || to.fanouts[0].Kind != ObjectNet {
			panic("boundary requires net")
		}
		setCopy(to.fanouts[0], output.Signal())
	}
	for _, module := range hierarchyModules(root) {
		for _, co := range module.Outputs() {
			copy := getCopy(co)
			if copy.Object == nil {
				continue
			}
			if len(co.fanins) != 1 {
				panic("CO requires one net")
			}
			e := toBarSignal(n, co.fanins[0].Object)
			e.Complement = e.Complement != co.fanins[0].Complement
			Abc_ObjAddFanin(copy.Object, e)
		}
	}
	n.Barriers = len(n.ObjectsOfKind(ObjectLatch))
	fmt.Printf("Hierarchy reader flattened %d instances of logic boxes and introduced %d barbufs.\n", count, n.Barriers)
	return n
}
func fromBarSignal(n *Network, o *Object) Edge {
	if e := getCopy(o); e.Object != nil {
		if e.Object.network != n {
			panic("logic crosses unmarked hierarchy boundary")
		}
		return e
	}
	if o.Kind == ObjectConst1 {
		return n.One()
	}
	if o.Kind == ObjectNet {
		if len(o.fanins) != 1 {
			panic("undriven optimized net")
		}
		e := fromBarSignal(n, o.fanins[0].Object)
		e.Complement = e.Complement != o.fanins[0].Complement
		if n.Type == uint8(ABC_NTK_NETLIST) {
			e = netForDriver(n, e).Signal()
		}
		setCopy(o, e)
		return e
	}
	if o.Kind != ObjectNode {
		panic("unmapped optimized boundary")
	}
	dup := Abc_NtkDupObj(n, o, 0)
	setCopy(o, dup.Signal())
	for _, fanin := range o.fanins {
		e := fromBarSignal(n, fanin.Object)
		e.Complement = e.Complement != fanin.Complement
		if n.Type == uint8(ABC_NTK_NETLIST) {
			e = netForDriver(n, e).Signal()
		}
		Abc_ObjAddFanin(dup, e)
	}
	result := dup.Signal()
	if n.Type == uint8(ABC_NTK_NETLIST) {
		result = netForDriver(n, result).Signal()
	}
	setCopy(o, result)
	return result
}
func Abc_NtkFromBarBufs_rec(n *Network, o *Object) Edge { return fromBarSignal(n, o) }
func Abc_NtkFromBarBufs(base, flat *Network) *Network {
	if base.Design == nil || len(flat.ObjectsOfKind(ObjectLatch)) != flat.Barriers || len(base.ObjectsOfKind(ObjectPI)) != len(flat.ObjectsOfKind(ObjectPI)) || len(base.ObjectsOfKind(ObjectPO)) != len(flat.ObjectsOfKind(ObjectPO)) {
		panic("incompatible hierarchy/barrier interfaces")
	}
	if Abc_NtkCheckSingleInstance(base) == 0 {
		panic("invalid hierarchy base")
	}
	if len(flat.Boxes()) != len(flat.ObjectsOfKind(ObjectLatch)) {
		panic("optimized graph contains non-barrier boxes")
	}
	for _, latch := range flat.ObjectsOfKind(ObjectLatch) {
		if len(latch.fanins) != 1 || len(latch.fanouts) != 1 || latch.fanins[0].Object.Kind != ObjectBI || latch.fanouts[0].Kind != ObjectBO {
			panic("malformed optimized barrier")
		}
	}
	for _, co := range flat.Outputs() {
		if len(co.fanins) != 1 {
			panic("optimized output undriven")
		}
	}
	cleanHierarchyCopies(base)
	modules := base.Design.Modules
	used := map[*Network]bool{}
	for _, module := range hierarchyModules(base) {
		used[module] = true
	}
	for _, module := range modules {
		clearCopies(module)
	}
	for _, module := range modules {
		if used[module] {
			cloneInterfaces(module)
		} else {
			cloneNetwork(module)
		}
	}
	design := NewDesign(base.Design.Name)
	for _, module := range modules {
		if err := design.AddModule(module.Copy); err != nil {
			panic(err)
		}
		if module.AltView != nil {
			module.Copy.AltView = module.AltView.Copy
		}
		for _, box := range module.Boxes() {
			box.Copy.Data = boxModel(box).Copy
		}
	}
	root := base.Copy
	for _, top := range base.Design.Tops {
		if top.Copy != nil {
			design.Tops = append(design.Tops, top.Copy)
		}
	}
	if len(design.Tops) == 0 {
		design.Tops = []*Network{root}
	}
	clearCopies(flat)
	for i, pi := range flat.ObjectsOfKind(ObjectPI) {
		setCopy(pi, root.ObjectsOfKind(ObjectPI)[i].Signal())
	}
	for i, po := range flat.ObjectsOfKind(ObjectPO) {
		setCopy(po, root.ObjectsOfKind(ObjectPO)[i].Signal())
	}
	var li, lo []*Object
	Abc_NtkCollectPiPos(base, &li, &lo)
	latches := flat.ObjectsOfKind(ObjectLatch)
	if len(li) != len(latches) || len(lo) != len(latches) {
		panic("barrier count changed")
	}
	for i, latch := range latches {
		if len(latch.fanins) != 1 || len(latch.fanouts) != 1 {
			panic("malformed barrier latch")
		}
		setCopy(latch.fanins[0].Object, getCopy(li[i]))
		setCopy(latch.fanouts[0], getCopy(lo[i]))
	}
	for _, co := range flat.Outputs() {
		copy := getCopy(co)
		if copy.Object == nil || len(co.fanins) != 1 {
			panic("unmapped optimized CO")
		}
		n := copy.Object.network
		e := fromBarSignal(n, co.fanins[0].Object)
		e.Complement = e.Complement != co.fanins[0].Complement
		if n.Type == uint8(ABC_NTK_NETLIST) {
			e = netForDriver(n, e).Signal()
		}
		Abc_ObjAddFanin(copy.Object, e)
	}

	for _, module := range modules {
		for _, ci := range module.Inputs() {
			copy := getCopy(ci)
			if copy.Object == nil {
				continue
			}
			net := netForDriver(copy.Object.network, copy)
			for _, oldNet := range ci.fanouts {
				if oldNet.Kind == ObjectNet {
					net.SetName(oldNet.Name())
					break
				}
			}
		}
		for _, co := range module.Outputs() {
			if len(co.fanins) == 1 && co.fanins[0].Object.Kind == ObjectNet && co.Copy != nil && len(co.Copy.fanins) == 1 {
				co.Copy.fanins[0].Object.SetName(co.fanins[0].Object.Name())
			}
		}
	}
	for _, ci := range flat.Inputs() {
		copy := getCopy(ci)
		if copy.Object != nil && len(ci.fanouts) > 0 && ci.fanouts[0].Kind == ObjectNet {
			netForDriver(copy.Object.network, copy).SetName(ci.fanouts[0].Name())
		}
	}

	return root
}
func Abc_NtkToBarBufsCollect_rec(o *Object, nodes *[]*Object) {
	if Abc_NodeIsTravIdCurrent(o) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(o)
	if o.Kind != ObjectNode && o.Kind != ObjectConst1 {
		panic("barrier traversal reached unordered boundary")
	}
	for _, e := range o.fanins {
		Abc_NtkToBarBufsCollect_rec(e.Object, nodes)
	}
	*nodes = append(*nodes, o)
}
func Abc_NtkToBarBufsCollect(n *Network) []*Object {
	latches := n.ObjectsOfKind(ObjectLatch)
	if n.Barriers == 0 || n.Barriers != len(latches) {
		panic("invalid barrier count")
	}
	Abc_NtkIncrementTravId(n)
	nodes := []*Object{n.constant}
	Abc_NodeSetTravIdCurrent(n.constant)
	for _, pi := range n.ObjectsOfKind(ObjectPI) {
		nodes = append(nodes, pi)
		Abc_NodeSetTravIdCurrent(pi)
	}
	for _, latch := range latches {
		if len(latch.fanins) != 1 || len(latch.fanouts) != 1 {
			panic("invalid barrier terminals")
		}
		in, out := latch.fanins[0].Object, latch.fanouts[0]
		if len(in.fanins) != 1 {
			panic("barrier input undriven")
		}
		Abc_NtkToBarBufsCollect_rec(in.fanins[0].Object, &nodes)
		nodes = append(nodes, in, latch, out)
		Abc_NodeSetTravIdCurrent(in)
		Abc_NodeSetTravIdCurrent(latch)
		Abc_NodeSetTravIdCurrent(out)
	}
	for _, po := range n.ObjectsOfKind(ObjectPO) {
		if len(po.fanins) != 1 {
			panic("PO undriven")
		}
		Abc_NtkToBarBufsCollect_rec(po.fanins[0].Object, &nodes)
		nodes = append(nodes, po)
		Abc_NodeSetTravIdCurrent(po)
	}
	if len(nodes) != len(n.Objects()) {
		panic("barrier-friendly traversal missed objects")
	}
	return nodes
}
func isBarrierBuffer(o *Object) bool {
	return o.Kind == ObjectNode && len(o.fanins) == 1 && (o.Barrier || o.Data == nil)
}
func Abc_NtkCountBarBufs(n *Network) int32 {
	count := int32(0)
	for _, o := range n.ObjectsOfKind(ObjectNode) {
		if isBarrierBuffer(o) {
			count++
		}
	}
	return count
}
func Abc_NtkBarBufsToBuffers(old *Network) *Network {
	if old.Design != nil {
		panic("convert flattened hierarchy first")
	}
	nodes := Abc_NtkToBarBufsCollect(old)
	n := copyNetworkMetadata(old)
	n.Barriers = 0
	clearCopies(old)
	setCopy(old.constant, n.One())
	for _, o := range nodes {
		switch o.Kind {
		case ObjectConst1:
			continue
		case ObjectPI:
			Abc_NtkDupObj(n, o, 1)
		case ObjectPO:
			dup := Abc_NtkDupObj(n, o, 1)
			e := getCopy(o.fanins[0].Object)
			e.Complement = e.Complement != o.fanins[0].Complement
			Abc_ObjAddFanin(dup, e)
		case ObjectBI, ObjectBO:
			e := getCopy(o.fanins[0].Object)
			e.Complement = e.Complement != o.fanins[0].Complement
			setCopy(o, e)
		case ObjectLatch:
			in := getCopy(o.fanins[0].Object)
			dup := n.NewObject(ObjectNode)
			dup.Barrier = true
			Abc_ObjAddFanin(dup, in)
			setCopy(o, dup.Signal())
		case ObjectNode:
			dup := Abc_NtkDupObj(n, o, 1)
			for _, fanin := range o.fanins {
				e := getCopy(fanin.Object)
				e.Complement = e.Complement != fanin.Complement
				Abc_ObjAddFanin(dup, e)
			}
		default:
			panic("unexpected barrier conversion object")
		}
	}
	return n
}
func Abc_NtkBarBufsFromBuffers(base, buffers *Network) *Network {
	if base.Barriers != len(base.ObjectsOfKind(ObjectLatch)) || len(base.ObjectsOfKind(ObjectPI)) != len(buffers.ObjectsOfKind(ObjectPI)) || len(base.ObjectsOfKind(ObjectPO)) != len(buffers.ObjectsOfKind(ObjectPO)) {
		panic("incompatible buffer interfaces")
	}
	if int(Abc_NtkCountBarBufs(buffers)) != base.Barriers {
		panic("barrier buffer count changed")
	}
	if err := buffers.Check(); err != nil {
		panic(err)
	}
	for _, po := range buffers.ObjectsOfKind(ObjectPO) {
		if len(po.fanins) != 1 {
			panic("optimized PO undriven")
		}
	}
	for _, latch := range base.ObjectsOfKind(ObjectLatch) {
		if len(latch.fanins) != 1 || len(latch.fanouts) != 1 || latch.fanins[0].Object.Kind != ObjectBI || latch.fanouts[0].Kind != ObjectBO {
			panic("invalid base barrier terminals")
		}
	}
	clearCopies(base)
	n := cloneInterfaces(base)
	n.Design = nil
	clearCopies(buffers)
	setCopy(buffers.constant, n.One())
	for i, pi := range buffers.ObjectsOfKind(ObjectPI) {
		setCopy(pi, n.ObjectsOfKind(ObjectPI)[i].Signal())
	}
	order, err := buffers.Topological(buffers.Objects())
	if err != nil {
		panic(err)
	}
	latches := n.ObjectsOfKind(ObjectLatch)
	index := 0
	for _, o := range order {
		if o.Kind != ObjectNode {
			continue
		}
		if isBarrierBuffer(o) {
			if index >= len(latches) {
				panic("too many barrier buffers")
			}
			latch := latches[index]
			index++
			in, out := latch.fanins[0].Object, latch.fanouts[0]
			e := getCopy(o.fanins[0].Object)
			e.Complement = e.Complement != o.fanins[0].Complement
			Abc_ObjAddFanin(in, e)
			setCopy(o, out.Signal())
		} else {
			dup := Abc_NtkDupObj(n, o, 1)
			for _, fanin := range o.fanins {
				e := getCopy(fanin.Object)
				e.Complement = e.Complement != fanin.Complement
				Abc_ObjAddFanin(dup, e)
			}
		}
	}
	if index != len(latches) {
		panic("barrier buffer count changed")
	}
	for i, po := range buffers.ObjectsOfKind(ObjectPO) {
		e := getCopy(po.fanins[0].Object)
		e.Complement = e.Complement != po.fanins[0].Complement
		Abc_ObjAddFanin(n.ObjectsOfKind(ObjectPO)[i], e)
	}
	return n
}
func Abc_NtkBarBufsOnOffTest(n *Network) *Network {
	return Abc_NtkBarBufsFromBuffers(n, Abc_NtkBarBufsToBuffers(n))
}
