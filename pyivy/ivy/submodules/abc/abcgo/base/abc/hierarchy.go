package abc

import "fmt"

// Design owns named modules and their ordered top-level models. Box.Data holds
// the instantiated *Network; box fanins/fanouts map by position to model PIs/POs.
type Design struct {
	Name    string
	Modules []*Network
	Tops    []*Network
	byName  map[string]*Network
}
type Abc_Des_t = Design

func NewDesign(name string) *Design { return &Design{Name: name, byName: make(map[string]*Network)} }
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
