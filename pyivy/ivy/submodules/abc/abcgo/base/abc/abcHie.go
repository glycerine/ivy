package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/map/mio"
	"sort"
	"strings"
)

func hieNetName(o *Object) string {
	if name := o.Name(); name != "" {
		return name
	}
	return fmt.Sprintf("n%d", o.ID)
}
func hieInputNet(o *Object) (*Object, error) {
	if len(o.fanins) != 1 || o.fanins[0].Object.Kind != ObjectNet {
		return nil, fmt.Errorf("terminal %d has no input net", o.ID)
	}
	return o.fanins[0].Object, nil
}
func hieOutputNet(o *Object) (*Object, error) {
	if len(o.fanouts) != 1 || o.fanouts[0].Kind != ObjectNet {
		return nil, fmt.Errorf("terminal %d has no output net", o.ID)
	}
	return o.fanouts[0], nil
}
func hieOrderTerminals(n *Network) { Abc_NtkOrderCisCos(n) }
func hieCopyBinding(o *Object, dst *Network) *Object {
	if o.Copy != nil && o.Copy.network == dst {
		return o.Copy
	}
	return nil
}
func hieConnect(dst *Network, object *Object, edge Edge) error {
	if object == nil || edge.Object == nil {
		return fmt.Errorf("unmapped hierarchy connection")
	}
	return dst.AddFanin(object, edge)
}
func hieUnaryNode(dst *Network, input, formal *Object, invert bool) (*Object, error) {
	if invert && Abc_ObjMvVarNum(formal) != 2 {
		return nil, fmt.Errorf("complemented nonbinary hierarchy port")
	}
	node := Abc_NtkCreateNodeBuf(dst, input)
	switch Abc_NtkFunc_t(dst.Function) {
	case ABC_FUNC_SOP:
		node.Data = "1 1\n"
		if invert {
			node.Data = "0 1\n"
		}
	case ABC_FUNC_BLIFMV:
		size := int(Abc_ObjMvVarNum(formal))
		table := &MVTable{DomainSizes: []int{size, size}, Cubes: []MVCube{{Inputs: []MVTerm{{Any: true, EqualInput: -1}}, Output: MVTerm{EqualInput: 0}}}}
		if invert {
			table.Cubes = []MVCube{{Inputs: []MVTerm{{Values: []int{0}, EqualInput: -1}}, Output: MVTerm{Values: []int{1}, EqualInput: -1}}, {Inputs: []MVTerm{{Values: []int{1}, EqualInput: -1}}, Output: MVTerm{Values: []int{0}, EqualInput: -1}}}
		}
		node.Data = table
	case ABC_FUNC_AIG:
		if manager, ok := dst.FunctionManager.(*hop.Hop_Man_t); ok {
			node.Data = hop.Hop_NotCond(hop.Hop_IthVar(manager, 0), abcBool(invert))
		} else {
			return nil, fmt.Errorf("local AIG manager missing")
		}
	case ABC_FUNC_BDD:
		if manager, ok := dst.FunctionManager.(*core.DdManager); ok {
			node.Data = core.Cudd_NotCond(core.Cudd_bddIthVar(manager, 0), invert)
		} else {
			return nil, fmt.Errorf("BDD manager missing")
		}
	case ABC_FUNC_MAP:
		manager, ok := dst.FunctionManager.(*mio.Mio_Library_t)
		if !ok {
			return nil, fmt.Errorf("mapped hierarchy needs MIO library")
		}
		gate := mio.Mio_LibraryReadBuf(manager)
		if invert {
			gate = mio.Mio_LibraryReadInv(manager)
		}
		if gate == nil {
			return nil, fmt.Errorf("mapped hierarchy needs unary cell")
		}
		node.Data = gate
	}
	return node, nil
}
func hieBridgeNet(dst *Network, input, output, formal *Object) error {
	if input == output {
		return nil
	}
	if len(output.fanins) != 0 {
		return fmt.Errorf("multiple drivers for hierarchical net %s", hieNetName(output))
	}
	node, err := hieUnaryNode(dst, input, formal, false)
	if err != nil {
		return err
	}
	return hieConnect(dst, output, node.Signal())
}
func hiePhaseNet(dst *Network, reference, formal *Object) *Object {
	net := dst.NewObject(ObjectNet)
	base := fmt.Sprintf("%s_phase_%d", hieNetName(reference), net.ID)
	name := base
	for index := 1; Abc_NtkFindNet(dst, name) != nil; index++ {
		name = fmt.Sprintf("%s_%d", base, index)
	}
	net.SetName(name)
	copyMVVariable(formal, net)
	return net
}

type hierarchyFlatten struct {
	dst     *Network
	counter *int32
	suffix  bool
	active  map[*Network]bool
}

func (context *hierarchyFlatten) flatten(src *Network, prefix string) error {
	if context.active[src] {
		return fmt.Errorf("recursive hierarchy at %s", src.Name)
	}
	context.active[src] = true
	defer delete(context.active, src)
	dst := context.dst
	if src.Function == uint8(ABC_FUNC_BLACKBOX) {
		boxes := src.Boxes()
		if len(boxes) != 1 {
			return fmt.Errorf("blackbox model %s has no unique body", src.Name)
		}
		body := boxes[0]
		if len(body.fanins) != len(src.ObjectsOfKind(ObjectPI)) || len(body.fanouts) != len(src.ObjectsOfKind(ObjectPO)) {
			return fmt.Errorf("blackbox model port mismatch")
		}
		copy := Abc_NtkDupBox(dst, body, 1)
		copy.Data = src
		for i, pi := range src.ObjectsOfKind(ObjectPI) {
			net, err := hieOutputNet(pi)
			if err != nil {
				return err
			}
			if err := hieConnect(dst, copy.fanins[i].Object, Edge{Object: hieCopyBinding(net, dst)}); err != nil {
				return err
			}
		}
		for i, po := range src.ObjectsOfKind(ObjectPO) {
			net, err := hieInputNet(po)
			if err != nil {
				return err
			}
			target := hieCopyBinding(net, dst)
			if po.Copy != nil && po.Copy.network == dst && po.Copy.Kind == ObjectNet {
				target = po.Copy
			}
			if err := hieConnect(dst, target, copy.fanouts[i].Signal()); err != nil {
				return err
			}
		}
		return nil
	}
	*context.counter++
	suffix := ""
	if context.suffix {
		if *context.counter != 0 {
			suffix = fmt.Sprintf("_%s_%d", src.Name, *context.counter)
		}
	} else {
		if *context.counter != 0 {
			prefix += fmt.Sprintf("(%d)", *context.counter)
		}
		prefix += "|"
	}
	nameFor := func(net *Object) string {
		if context.suffix {
			return hieNetName(net) + suffix
		}
		return prefix + hieNetName(net)
	}
	setCopy(src.constant, dst.One())
	for _, net := range src.ObjectsOfKind(ObjectNet) {
		if hieCopyBinding(net, dst) == nil {
			setCopy(net, Abc_NtkFindOrCreateNet(dst, nameFor(net)).Signal())
		}
		if variable := Abc_ObjMvVar(net); variable != nil {
			existing := Abc_ObjMvVar(net.Copy)
			if existing != nil && existing.Values != variable.Values {
				return fmt.Errorf("hierarchical MV domain mismatch on %s", hieNetName(net))
			}
			if existing == nil {
				copyMVVariable(net, net.Copy)
			}
		}
	}
	excluded := make(map[*Object]bool)
	for _, pi := range src.ObjectsOfKind(ObjectPI) {
		excluded[pi] = true
	}
	for _, po := range src.ObjectsOfKind(ObjectPO) {
		excluded[po] = true
		net, err := hieInputNet(po)
		if err != nil {
			return err
		}
		if strings.HasPrefix(hieNetName(net), "abc_property") {
			target := hieCopyBinding(net, dst)
			if po.Copy != nil && po.Copy.network == dst && po.Copy.Kind == ObjectNet {
				target = po.Copy
			}
			already := false
			for _, fanout := range target.fanouts {
				already = already || fanout.Kind == ObjectPO
			}
			if !already {
				property := dst.NewObject(ObjectPO)
				property.SetName(nameFor(net))
				if err := hieConnect(dst, property, target.Signal()); err != nil {
					return err
				}
			}
			target.SetName(nameFor(net))
		}
	}
	for _, box := range src.boxes {
		if box.Kind == ObjectLatch {
			continue
		}
		excluded[box] = true
		for _, edge := range box.fanins {
			excluded[edge.Object] = true
		}
		for _, bo := range box.fanouts {
			excluded[bo] = true
		}
	}
	for _, object := range src.Objects() {
		if excluded[object] || hieCopyBinding(object, dst) != nil {
			continue
		}
		copy := Abc_NtkDupObj(dst, object, 0)
		copy.Barrier = object.Barrier
		setCopy(object, copy.Signal())
	}
	for _, object := range src.Objects() {
		if excluded[object] {
			continue
		}
		copy := hieCopyBinding(object, dst)
		for _, edge := range object.fanins {
			if excluded[edge.Object] {
				continue
			}
			if err := hieConnect(dst, copy, Edge{Object: hieCopyBinding(edge.Object, dst), Complement: edge.Complement}); err != nil {
				return err
			}
		}
	}
	for _, box := range src.boxes {
		if box.Kind == ObjectLatch {
			continue
		}
		model, ok := box.Data.(*Network)
		if !ok || model == nil {
			return fmt.Errorf("box %d has no model", box.ID)
		}
		pis, pos := model.ObjectsOfKind(ObjectPI), model.ObjectsOfKind(ObjectPO)
		if len(box.fanins) != len(pis) || len(box.fanouts) != len(pos) {
			return fmt.Errorf("box %s model port mismatch", box.Name())
		}
		clearCopies(model)
		for i, edge := range box.fanins {
			actual, err := hieInputNet(edge.Object)
			if err != nil {
				return err
			}
			formal, err := hieOutputNet(pis[i])
			if err != nil {
				return err
			}
			target := hieCopyBinding(actual, dst)
			phase := edge.Complement != edge.Object.fanins[0].Complement
			phase = phase != formal.fanins[0].Complement
			if phase {
				inverse, err := hieUnaryNode(dst, target, formal, true)
				if err != nil {
					return err
				}
				target = hiePhaseNet(dst, target, formal)
				if err := hieConnect(dst, target, inverse.Signal()); err != nil {
					return err
				}
			}
			setCopy(formal, target.Signal())
		}
		for i, bo := range box.fanouts {
			actual, err := hieOutputNet(bo)
			if err != nil {
				return err
			}
			formal, err := hieInputNet(pos[i])
			if err != nil {
				return err
			}
			target := hieCopyBinding(actual, dst)
			binding := target
			phase := pos[i].fanins[0].Complement != bo.fanins[0].Complement
			phase = phase != actual.fanins[0].Complement
			if phase {
				binding = hiePhaseNet(dst, target, formal)
				inverse, err := hieUnaryNode(dst, binding, formal, true)
				if err != nil {
					return err
				}
				if err := hieConnect(dst, target, inverse.Signal()); err != nil {
					return err
				}
			}
			if existing := hieCopyBinding(formal, dst); existing != nil && existing != binding {
				if err := hieBridgeNet(dst, existing, binding, formal); err != nil {
					return err
				}
			} else {
				setCopy(formal, binding.Signal())
			}
			setCopy(pos[i], target.Signal())
		}
		childPrefix := prefix + model.Name
		if context.suffix {
			childPrefix = ""
		}
		if err := context.flatten(model, childPrefix); err != nil {
			return err
		}
	}
	return nil
}
func Abc_NtkFlattenLogicHierarchy2_rec(dst, src *Network, counter *int32) {
	context := hierarchyFlatten{dst: dst, counter: counter, suffix: true, active: make(map[*Network]bool)}
	if err := context.flatten(src, ""); err != nil {
		panic(err)
	}
}
func Abc_NtkFlattenLogicHierarchy_rec(dst, src *Network, counter *int32, prefix *[]byte) {
	context := hierarchyFlatten{dst: dst, counter: counter, active: make(map[*Network]bool)}
	if err := context.flatten(src, string(*prefix)); err != nil {
		panic(err)
	}
}
func hieBlackboxDesign(src, dst *Network) error {
	if src.Design == nil {
		return nil
	}
	design := NewDesign(src.Design.Name)
	if err := design.AddModule(dst); err != nil {
		return err
	}
	design.Tops = []*Network{dst}
	models := map[*Network]*Network{}
	for _, model := range src.Design.Modules {
		if model.Function == uint8(ABC_FUNC_BLACKBOX) && model != src {
			copy := cloneNetwork(model)
			if err := design.AddModule(copy); err != nil {
				return err
			}
			models[model] = copy
		}
	}
	for _, box := range dst.ObjectsOfKind(ObjectBlackbox) {
		model, ok := box.Data.(*Network)
		if !ok || model == nil {
			return fmt.Errorf("blackbox has no model")
		}
		copy := models[model]
		if copy == nil {
			copy = cloneNetwork(model)
			if err := design.AddModule(copy); err != nil {
				return err
			}
			models[model] = copy
		}
		box.Data = copy
	}
	return nil
}
func hieValidateModels(root *Network) bool {
	seen := make(map[*Network]bool)
	var visit func(*Network) bool
	visit = func(model *Network) bool {
		if model == nil || model.Type != uint8(ABC_NTK_NETLIST) {
			return false
		}
		if seen[model] {
			return true
		}
		seen[model] = true
		if model.Check() != nil {
			return false
		}
		if model.Function == uint8(ABC_FUNC_BLACKBOX) {
			return true
		}
		for _, box := range model.boxes {
			if box.Kind == ObjectLatch {
				continue
			}
			child, ok := box.Data.(*Network)
			if !ok || !visit(child) {
				return false
			}
		}
		return true
	}
	return visit(root)
}
func flattenLogicHierarchy(src *Network, suffix bool) *Network {
	if src.Type != uint8(ABC_NTK_NETLIST) || Abc_NtkIsAcyclicHierarchy(src) == 0 || !hieValidateModels(src) {
		return nil
	}
	dst := copyNetworkMetadata(src)
	clearCopies(src)
	setCopy(src.constant, dst.One())
	for _, pi := range src.ObjectsOfKind(ObjectPI) {
		net, err := hieOutputNet(pi)
		if err != nil {
			return nil
		}
		copy := Abc_NtkDupObj(dst, pi, 1)
		setCopy(pi, copy.Signal())
		target := Abc_NtkFindOrCreateNet(dst, hieNetName(net))
		setCopy(net, target.Signal())
		if err := hieConnect(dst, target, Edge{Object: copy, Complement: net.fanins[0].Complement}); err != nil {
			return nil
		}
	}
	for _, po := range src.ObjectsOfKind(ObjectPO) {
		net, err := hieInputNet(po)
		if err != nil {
			return nil
		}
		copy := Abc_NtkDupObj(dst, po, 1)
		setCopy(po, copy.Signal())
		target := Abc_NtkFindOrCreateNet(dst, hieNetName(net))
		setCopy(net, target.Signal())
		if err := hieConnect(dst, copy, Edge{Object: target, Complement: po.fanins[0].Complement}); err != nil {
			return nil
		}
	}
	counter := int32(-1)
	context := hierarchyFlatten{dst: dst, counter: &counter, suffix: suffix, active: make(map[*Network]bool)}
	if err := context.flatten(src, src.Name); err != nil {
		return nil
	}
	if err := hieBlackboxDesign(src, dst); err != nil {
		return nil
	}
	hieOrderTerminals(dst)
	if Abc_NtkCheck(dst) == 0 {
		return nil
	}
	fmt.Printf("Hierarchy reader flattened %d instances of logic boxes and left %d black boxes.\n", counter, len(dst.ObjectsOfKind(ObjectBlackbox)))
	return dst
}
func Abc_NtkFlattenLogicHierarchy2(src *Network) *Network { return flattenLogicHierarchy(src, true) }
func Abc_NtkFlattenLogicHierarchy(src *Network) *Network  { return flattenLogicHierarchy(src, false) }
func Abc_NtkCompareNames(a, b *Network) int32             { return int32(strings.Compare(a.Name, b.Name)) }
func Abc_NtkPrintBoxInfo(n *Network) {
	if n.Design == nil {
		return
	}
	modules := n.Design.Modules
	sort.SliceStable(modules, func(i, j int) bool { return modules[i].Name < modules[j].Name })
	for i, module := range modules {
		if module == n {
			modules[0], modules[i] = modules[i], modules[0]
			break
		}
	}
	printModel := func(model *Network) {
		fmt.Printf("MODULE  %-30s : PI=%6d PO=%6d BB=%6d ND=%6d Lev=%5d\n", model.Name, len(model.ObjectsOfKind(ObjectPI)), len(model.ObjectsOfKind(ObjectPO)), len(model.boxes), len(model.ObjectsOfKind(ObjectNode)), Abc_NtkLevel(model))
	}
	for _, module := range modules {
		if len(module.boxes) == 0 {
			continue
		}
		printModel(module)
		counts := make(map[*Network]int)
		for _, box := range module.boxes {
			if model, ok := box.Data.(*Network); ok {
				counts[model]++
			}
		}
		for _, model := range modules {
			if counts[model] > 0 {
				fmt.Printf("%15d : %s\n", counts[model], model.Name)
			}
		}
	}
	for _, module := range modules {
		if len(module.boxes) == 0 {
			printModel(module)
		}
	}
}

func Abc_NtkConvertBlackboxes(src *Network) *Network {
	if src.Type != uint8(ABC_NTK_NETLIST) || len(src.ObjectsOfKind(ObjectWhitebox)) > 0 {
		return nil
	}
	dst := copyNetworkMetadata(src)
	clearCopies(src)
	setCopy(src.constant, dst.One())
	excluded := make(map[*Object]bool)
	for _, o := range src.ObjectsOfKind(ObjectPO) {
		excluded[o] = true
	}
	for _, box := range src.ObjectsOfKind(ObjectBlackbox) {
		excluded[box] = true
		for _, edge := range box.fanins {
			excluded[edge.Object] = true
		}
		for _, bo := range box.fanouts {
			excluded[bo] = true
			pi := dst.NewObject(ObjectPI)
			net, err := hieOutputNet(bo)
			if err != nil {
				return nil
			}
			pi.SetName(hieNetName(net))
			setCopy(bo, pi.Signal())
		}
	}
	for _, o := range src.Objects() {
		if excluded[o] || o == src.constant {
			continue
		}
		copy := Abc_NtkDupObj(dst, o, 1)
		copy.Barrier = o.Barrier
		setCopy(o, copy.Signal())
	}
	for _, o := range src.Objects() {
		if excluded[o] {
			continue
		}
		for _, edge := range o.fanins {
			if err := hieConnect(dst, o.Copy, Edge{Object: hieCopyBinding(edge.Object, dst), Complement: edge.Complement}); err != nil {
				return nil
			}
		}
	}
	seen := make(map[*Object]bool)
	for _, co := range src.cos {
		if len(co.fanouts) > 0 && co.fanouts[0].Kind == ObjectLatch {
			continue
		}
		net, err := hieInputNet(co)
		if err != nil {
			return nil
		}
		if seen[net] {
			continue
		}
		seen[net] = true
		po := dst.NewObject(ObjectPO)
		name := co.Name()
		if name == "" {
			name = hieNetName(net)
		}
		po.SetName(name)
		if err := hieConnect(dst, po, net.Copy.Signal()); err != nil {
			return nil
		}
	}
	hieOrderTerminals(dst)
	if Abc_NtkCheck(dst) == 0 {
		return nil
	}
	return dst
}
func hieHasPO(net *Object) bool {
	for _, fanout := range net.fanouts {
		if fanout.Kind == ObjectPO {
			return true
		}
	}
	return false
}
func Abc_NtkInsertNewLogic(hierarchy, logic *Network) *Network {
	if hierarchy.Type != uint8(ABC_NTK_NETLIST) || logic.Type != uint8(ABC_NTK_NETLIST) || len(hierarchy.ObjectsOfKind(ObjectWhitebox)) > 0 || len(logic.ObjectsOfKind(ObjectWhitebox)) > 0 || len(logic.ObjectsOfKind(ObjectBlackbox)) > 0 {
		return nil
	}
	dst := copyNetworkMetadata(logic)
	dst.Name, dst.Spec = hierarchy.Name, hierarchy.Spec
	clearCopies(logic)
	setCopy(logic.constant, dst.One())
	bindPI := func(oldNet *Object, target *Object) bool {
		net := Abc_NtkFindNet(logic, hieNetName(oldNet))
		if net == nil || len(net.fanins) != 1 || net.fanins[0].Object.Kind != ObjectPI || hieCopyBinding(net, dst) != nil {
			return false
		}
		copy := Abc_NtkFindOrCreateNet(dst, hieNetName(oldNet))
		setCopy(net, copy.Signal())
		pi := net.fanins[0].Object
		if target == nil {
			target = Abc_NtkDupObj(dst, pi, 1)
		}
		setCopy(pi, target.Signal())
		return true
	}
	for _, pi := range hierarchy.ObjectsOfKind(ObjectPI) {
		net, err := hieOutputNet(pi)
		if err != nil || !bindPI(net, nil) {
			return nil
		}
	}
	for _, box := range hierarchy.ObjectsOfKind(ObjectBlackbox) {
		copy := Abc_NtkDupBox(dst, box, 1)
		for i, bo := range box.fanouts {
			net, err := hieOutputNet(bo)
			if err != nil || !bindPI(net, copy.fanouts[i]) {
				return nil
			}
		}
	}
	outputNet := func(oldNet *Object) *Object {
		net := Abc_NtkFindNet(logic, hieNetName(oldNet))
		if net == nil || !hieHasPO(net) {
			return nil
		}
		if hieCopyBinding(net, dst) == nil {
			setCopy(net, Abc_NtkFindOrCreateNet(dst, hieNetName(oldNet)).Signal())
		}
		return net.Copy
	}
	for _, po := range hierarchy.ObjectsOfKind(ObjectPO) {
		net, err := hieInputNet(po)
		if err != nil {
			return nil
		}
		target := outputNet(net)
		if target == nil {
			return nil
		}
		copy := Abc_NtkDupObj(dst, po, 1)
		if err := hieConnect(dst, copy, target.Signal()); err != nil {
			return nil
		}
	}
	for _, box := range hierarchy.ObjectsOfKind(ObjectBlackbox) {
		for i, edge := range box.fanins {
			net, err := hieInputNet(edge.Object)
			if err != nil {
				return nil
			}
			target := outputNet(net)
			if target == nil {
				return nil
			}
			if err := hieConnect(dst, box.Copy.fanins[i].Object, target.Signal()); err != nil {
				return nil
			}
		}
	}
	for _, o := range logic.Objects() {
		if o.Kind == ObjectPO || hieCopyBinding(o, dst) != nil {
			continue
		}
		copy := Abc_NtkDupObj(dst, o, 1)
		copy.Barrier = o.Barrier
		setCopy(o, copy.Signal())
	}
	for _, o := range logic.Objects() {
		if o.Kind == ObjectPO {
			continue
		}
		copy := hieCopyBinding(o, dst)
		for _, edge := range o.fanins {
			if err := hieConnect(dst, copy, Edge{Object: hieCopyBinding(edge.Object, dst), Complement: edge.Complement}); err != nil {
				return nil
			}
		}
	}
	hieOrderTerminals(dst)
	if Abc_NtkCheck(dst) == 0 {
		return nil
	}
	if design := hierarchy.Design; design != nil {
		for i, module := range design.Modules {
			if module == hierarchy {
				design.Modules[i] = dst
			}
		}
		for i, top := range design.Tops {
			if top == hierarchy {
				design.Tops[i] = dst
			}
		}
		if design.byName == nil {
			design.byName = make(map[string]*Network)
		}
		design.byName[dst.Name] = dst
		dst.Design = design
		hierarchy.Design = nil
	}
	return dst
}
