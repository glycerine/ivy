package abc

import (
	"fmt"
	"reflect"

	"github.com/glycerine/abcgo/aig/hop"
)

// Design owns named modules and their ordered top-level models. Box.Data holds
// the instantiated *Network; box fanins/fanouts map by position to model PIs/POs.
type Design struct {
	Name            string
	Modules         []*Network
	Tops            []*Network
	FunctionManager *hop.Hop_Man_t
	byName          map[string]*Network
}
type Abc_Des_t = Design

func NewDesign(name string) *Design { return &Design{Name: name, byName: make(map[string]*Network)} }

func Abc_DesCreate(name string) *Design {
	d := NewDesign(name)
	d.FunctionManager = hop.Hop_ManStart()
	return d
}

func Abc_DesAddModel(d *Design, n *Network) int32 {
	if d.AddModule(n) != nil {
		return 0
	}
	return 1
}

func Abc_DesFindModelByName(d *Design, name string) *Network { return d.Module(name) }

// sameManager compares manager identity without panicking on non-comparable data.
func sameManager(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
	return av.Type() == bv.Type() && av.Comparable() && av.Interface() == bv.Interface()
}

func Abc_DesCleanManPointer(d *Design, manager any) {
	if d == nil {
		return
	}
	if sameManager(d.FunctionManager, manager) {
		d.FunctionManager = nil
	}
	for _, module := range d.Modules {
		if module != nil && sameManager(module.FunctionManager, manager) {
			module.FunctionManager = nil
		}
	}
}

func Abc_DesFree(d *Design, saved *Network) {
	if d == nil {
		return
	}
	var released []any
	if d.FunctionManager != nil && (saved == nil || !sameManager(d.FunctionManager, saved.FunctionManager)) {
		released = append(released, d.FunctionManager)
		hop.Hop_ManStop(d.FunctionManager)
	}
	for _, module := range d.Modules {
		if module == nil || module == saved {
			continue
		}
		manager := module.FunctionManager
		shared := saved != nil && sameManager(manager, saved.FunctionManager)
		for _, previous := range released {
			shared = shared || sameManager(manager, previous)
		}
		if shared {
			module.FunctionManager = nil
		} else if manager != nil {
			released = append(released, manager)
		}
		module.Design = nil
		Abc_NtkDelete(module)
	}
	if saved != nil && saved.Design == d {
		saved.Design = nil
	}
	d.Modules, d.Tops, d.byName = nil, nil, nil
	d.FunctionManager = nil
}

func Abc_DesDupBlackboxes(d *Design, saved *Network) *Design {
	if d == nil || saved == nil || len(d.Tops) == 0 || len(d.Modules) < 2 {
		panic("black-box duplication requires a saved top and a multi-module design")
	}
	result := Abc_DesCreate(d.Name)
	// ABC retains the caller's saved network without transferring its ownership.
	result.Tops = append(result.Tops, saved)
	result.Modules = append(result.Modules, saved)
	for _, module := range d.Modules {
		if Abc_NtkHasBlackbox(module) != 0 {
			result.Modules = append(result.Modules, Abc_NtkDup(module))
		}
	}
	return result
}

func Abc_DesPrint(d *Design) {
	fmt.Printf("Models of design %s:\n", d.Name)
	for i, module := range d.Modules {
		fmt.Printf("%2d : %20s   nd = %6d   lat = %6d   whitebox = %3d   blackbox = %3d\n", i+1, module.Name, Abc_NtkNodeNum(module), Abc_NtkLatchNum(module), Abc_NtkWhiteboxNum(module), Abc_NtkBlackboxNum(module))
		if Abc_NtkBlackboxNum(module) == 0 {
			continue
		}
		for _, box := range module.Boxes() {
			model, ok := box.Data.(*Network)
			if !ok || model == nil {
				continue
			}
			switch box.Kind {
			case ObjectWhitebox:
				fmt.Printf("     %20s (whitebox)\n", model.Name)
			case ObjectBlackbox:
				fmt.Printf("     %20s (blackbox)\n", model.Name)
			}
		}
	}
}

func Abc_DesDup(d *Design) *Design {
	if d == nil {
		return nil
	}
	copy := Abc_DesCreate(d.Name)
	models := make(map[*Network]*Network, len(d.Modules))
	for _, module := range d.Modules {
		duplicate := Abc_NtkDup(module)
		if err := copy.AddModule(duplicate); err != nil {
			panic(err)
		}
		models[module] = duplicate
	}
	for _, top := range d.Tops {
		copy.Tops = append(copy.Tops, models[top])
	}
	for _, module := range d.Modules {
		duplicate := models[module]
		duplicate.AltView = models[module.AltView]
		for _, box := range module.Boxes() {
			if box.Kind != ObjectWhitebox && box.Kind != ObjectBlackbox {
				continue
			}
			if model, ok := box.Data.(*Network); ok && model != nil {
				if replacement := models[model]; replacement != nil {
					box.Copy.Data = replacement
				}
			}
		}
	}
	return copy
}

func Abc_DesDeriveRoot(d *Design) *Network {
	if d == nil || len(d.Modules) == 0 {
		return nil
	}
	if len(d.Modules) > 1 {
		fmt.Println("The design includes more than one module and is currently not used.")
		return nil
	}
	root := d.Modules[0]
	root.FunctionManager = d.FunctionManager
	d.FunctionManager = nil
	d.Modules = nil
	d.Tops = nil
	delete(d.byName, root.Name)
	root.Design = nil
	return root
}

func Abc_DesFindTopLevelModels(d *Design) int32 {
	referenced := make(map[*Network]bool)
	for _, module := range d.Modules {
		for _, box := range module.Boxes() {
			if box.Kind == ObjectLatch {
				continue
			}
			if model, ok := box.Data.(*Network); ok && model != nil {
				referenced[model] = true
			}
		}
	}
	d.Tops = d.Tops[:0]
	for _, module := range d.Modules {
		if !referenced[module] {
			d.Tops = append(d.Tops, module)
		}
	}
	return int32(len(d.Tops))
}
func (d *Design) AddModule(n *Network) error {
	if d.byName == nil {
		d.byName = make(map[string]*Network)
	}
	if n == nil {
		return fmt.Errorf("nil module")
	}
	for _, registered := range d.Modules {
		if registered == n {
			return fmt.Errorf("module already registered")
		}
	}
	if _, exists := d.byName[n.Name]; exists {
		return fmt.Errorf("duplicate module %q", n.Name)
	}
	if n.Design != nil && n.Design != d {
		return fmt.Errorf("module already belongs to another design")
	}
	d.byName[n.Name] = n
	d.Modules = append(d.Modules, n)
	n.Design = d
	return nil
}
func (d *Design) Module(name string) *Network { return d.byName[name] }
func (n *Network) Inputs() []*Object          { return append([]*Object(nil), n.cis...) }
func (n *Network) Outputs() []*Object         { return append([]*Object(nil), n.cos...) }
func (n *Network) Boxes() []*Object           { return append([]*Object(nil), n.boxes...) }
func (n *Network) TerminalIDs(kind ObjectKind) []int {
	objects := n.ObjectsOfKind(kind)
	ids := make([]int, len(objects))
	for i, o := range objects {
		ids[i] = o.ID
	}
	return ids
}

// SetTerminalOrder changes interface order without changing stable object IDs.
func (n *Network) SetTerminalOrder(kind ObjectKind, ids []int) error {
	if kind != ObjectPI && kind != ObjectPO && kind != ObjectBI && kind != ObjectBO {
		return fmt.Errorf("not a terminal kind")
	}
	current := n.ObjectsOfKind(kind)
	if len(current) != len(ids) {
		return fmt.Errorf("terminal order length mismatch")
	}
	ordered := make([]*Object, len(ids))
	seen := map[int]bool{}
	for i, id := range ids {
		o := n.Object(id)
		if o == nil || o.Kind != kind || seen[id] {
			return fmt.Errorf("invalid terminal permutation")
		}
		seen[id] = true
		ordered[i] = o
	}
	n.kindOrder[kind] = ordered
	combined := &n.cos
	if kind == ObjectPI || kind == ObjectBO {
		combined = &n.cis
	}
	next := 0
	for i, o := range *combined {
		if o.Kind == kind {
			(*combined)[i] = ordered[next]
			next++
		}
	}
	return nil
}
func removeObjectOrder(objects []*Object, target *Object) []*Object {
	for i, o := range objects {
		if o == target {
			return append(objects[:i], objects[i+1:]...)
		}
	}
	return objects
}
func copyNetworkMetadata(old *Network) *Network {
	n := NewNetwork(old.Name)
	n.Spec = old.Spec
	n.Type = old.Type
	n.Function = old.Function
	n.FunctionManager = old.FunctionManager
	n.Barriers = old.Barriers
	return n
}
func clearCopies(n *Network) {
	for _, o := range n.Objects() {
		o.Copy = nil
		o.CopyEdge = Edge{}
	}
}
func cleanHierarchyCopies(n *Network) {
	seen := map[*Network]bool{}
	var walk func(*Network)
	walk = func(module *Network) {
		if seen[module] {
			return
		}
		seen[module] = true
		clearCopies(module)
		for _, box := range module.Boxes() {
			if model, ok := box.Data.(*Network); ok {
				walk(model)
			}
		}
	}
	walk(n)
}
func setCopy(o *Object, e Edge) { o.Copy = e.Object; o.CopyEdge = e }
func getCopy(o *Object) Edge {
	if o.CopyEdge.Object != nil {
		return o.CopyEdge
	}
	if o.Copy != nil {
		return o.Copy.Signal()
	}
	return Edge{}
}
func newBarrier(n *Network, init Abc_InitType_t) (*Object, *Object, *Object) {
	input, latch, output := n.NewObject(ObjectBI), n.NewObject(ObjectLatch), n.NewObject(ObjectBO)
	latch.Data = init
	Abc_ObjAddFanin(latch, input)
	Abc_ObjAddFanin(output, latch)
	return input, latch, output
}
func boxModel(box *Object) *Network {
	model, ok := box.Data.(*Network)
	if !ok || model == nil {
		panic("box has no logic model")
	}
	return model
}

// cloneInterfaces clones ports and boxes, leaving all logic to reconstruction.
func cloneInterfaces(old *Network) *Network {
	n := copyNetworkMetadata(old)
	setCopy(old.constant, n.One())
	old.Copy = n
	for _, o := range old.ObjectsOfKind(ObjectPI) {
		setCopy(o, Abc_NtkDupObj(n, o, 1).Signal())
	}
	for _, o := range old.ObjectsOfKind(ObjectPO) {
		setCopy(o, Abc_NtkDupObj(n, o, 1).Signal())
	}
	for _, box := range old.Boxes() {
		dup := Abc_NtkDupObj(n, box, 1)
		setCopy(box, dup.Signal())
		for _, e := range box.fanins {
			term := Abc_NtkDupObj(n, e.Object, 1)
			setCopy(e.Object, term.Signal())
			Abc_ObjAddFanin(dup, term)
		}
		for _, term := range box.fanouts {
			out := Abc_NtkDupObj(n, term, 1)
			setCopy(term, out.Signal())
			Abc_ObjAddFanin(out, dup)
		}
	}
	return n
}
func netForDriver(n *Network, e Edge) *Object {
	if e.Object.Kind == ObjectNet && !e.Complement {
		return e.Object
	}
	for _, o := range e.Object.fanouts {
		if o.Kind == ObjectNet && len(o.fanins) == 1 && o.fanins[0] == e {
			return o
		}
	}
	net := n.NewObject(ObjectNet)
	Abc_ObjAddFanin(net, e)
	return net
}

// cloneNetwork preserves a module which the optimized top-level cone did not
// visit. All interfaces, logic, names and connection polarities remain owned by
// the new design rather than referring back to the old graph.

// Hierarchy cloning uses the public duplicate routine so function managers are
// independently owned and may be released when either network is deleted.
func cloneNetwork(old *Network) *Network { return Abc_NtkDup(old) }
