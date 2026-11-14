package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/nm"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// ObjectKind retains the ABC object-kind numbering while replacing packed
// bitfields and tagged pointers with ordinary Go values.
type ObjectKind uint8

const (
	ObjectNone ObjectKind = iota
	ObjectConst1
	ObjectPI
	ObjectPO
	ObjectBI
	ObjectBO
	ObjectNet
	ObjectNode
	ObjectLatch
	ObjectWhitebox
	ObjectBlackbox
)

// Edge denotes an object or its Boolean complement.
type Edge struct {
	Object     *Object
	Complement bool
}

func (e Edge) Not() Edge { e.Complement = !e.Complement; return e }

// Object is a network object. Fanin order and fanout multiplicity are retained.
type Object struct {
	MVEncoding []Edge
	// BoxInputPolarity records the optional input-complement mask used by ABC's
	// hierarchical gate strashing interface.
	BoxInputPolarity    []uint32
	CopyEdge            Edge
	CopyData            any // temporary non-object copy payload (e.g. a mapped gate)
	Barrier             bool
	Choice              *Object
	Next                *Object
	Phase               bool
	Persist             bool // ABC's fPersist flag
	reverseLevel        int
	travID              uint64
	crossCutVisits      int
	ID                  int
	Kind                ObjectKind
	Level               int
	Temp                int
	DTemp               float64
	TerminalIndex       int
	Data                any
	Copy                *Object
	MarkA, MarkB, MarkC bool
	network             *Network
	fanins              []Edge
	fanouts             []*Object
}

func (o *Object) Network() *Network  { return o.network }
func (o *Object) Signal() Edge       { return Edge{Object: o} }
func (o *Object) Fanins() []Edge     { return append([]Edge(nil), o.fanins...) }
func (o *Object) Fanouts() []*Object { return append([]*Object(nil), o.fanouts...) }
func (o *Object) Name() string       { return nm.Nm_ManFindNameById(o.network.names, int32(o.ID)) }
func (o *Object) SetName(name string) {
	nm.Nm_ManDeleteIdName(o.network.names, int32(o.ID))
	nm.Nm_ManStoreIdName(o.network.names, int32(o.ID), int32(o.Kind), name, "")
}
func (o *Object) IsCI() bool { return o.Kind == ObjectPI || o.Kind == ObjectBO }
func (o *Object) IsCO() bool { return o.Kind == ObjectPO || o.Kind == ObjectBI }

// Network owns objects with stable IDs; deletion leaves holes, as in ABC.
// Mutations must be serialized by the caller.
type Network struct {
	Faults                    []int32
	Constraints               int32
	Exdc, Excare              *Network
	Onehots                   [][]int32
	SequenceModel             *util.Abc_Cex_t
	seqModelVec               *vec.Vec_Ptr_t
	ObjectPermutation, Phases []int32
	NameIDs                   []int32
	Topo                      []int32
	AndGateDelay              float64
	WireLoad                  string
	LTLProperties             []string
	Model                     []int32
	Data                      any
	Attributes                []*vec.Vec_Att_t
	Step                      int32
	DTemp                     float64
	FunctionManager           any
	timingMapping             any
	phaseTracking             bool
	mvVars                    map[*Object]*MVVariable
	Design                    *Design
	Copy, AltView             *Network
	Type, Function            uint8
	Barriers                  int
	// CILevel supplies timing-derived initial levels when a timing manager is attached.
	CILevel             func(*Object) int
	kindOrder           map[ObjectKind][]*Object
	cis, cos, boxes     []*Object
	travID              uint64
	levelMax            int
	reverseLevelsActive bool
	refs                map[*Object]int
	Name, Spec          string
	objects             []*Object
	names               *nm.Manager
	hashed              map[[2]Edge]Edge
	constant            *Object
}

func NewNetwork(name string) *Network {
	n := &Network{Step: -1, DTemp: -1, travID: 1, Type: 3, Function: 3, kindOrder: make(map[ObjectKind][]*Object), Name: name, names: nm.Nm_ManCreate(16), hashed: make(map[[2]Edge]Edge)}
	n.constant = n.NewObject(ObjectConst1)
	n.constant.Phase = true
	return n
}
func (n *Network) NewObject(kind ObjectKind) *Object {
	if kind <= ObjectNone || kind > ObjectBlackbox {
		panic("invalid ABC object kind")
	}
	o := &Object{ID: len(n.objects), Kind: kind, network: n}
	n.objects = append(n.objects, o)
	if n.kindOrder == nil {
		n.kindOrder = make(map[ObjectKind][]*Object)
	}
	n.kindOrder[kind] = append(n.kindOrder[kind], o)
	if o.IsCI() {
		n.cis = append(n.cis, o)
	}
	if o.IsCO() {
		n.cos = append(n.cos, o)
	}
	if kind >= ObjectLatch && kind <= ObjectBlackbox {
		n.boxes = append(n.boxes, o)
	}
	return o
}
func (n *Network) Object(id int) *Object {
	if id < 0 || id >= len(n.objects) {
		return nil
	}
	return n.objects[id]
}
func (n *Network) Objects() []*Object {
	out := make([]*Object, 0, len(n.objects))
	for _, o := range n.objects {
		if o != nil {
			out = append(out, o)
		}
	}
	return out
}
func (n *Network) ObjectsOfKind(kind ObjectKind) []*Object {
	return append([]*Object(nil), n.kindOrder[kind]...)
}

// ChangeObjectKind updates the per-kind index while preserving the object's
// identity and its position in the combined box order.
func (n *Network) ChangeObjectKind(o *Object, kind ObjectKind) error {
	if n == nil || o == nil || o.network != n {
		return fmt.Errorf("object does not belong to network")
	}
	if kind <= ObjectNone || kind > ObjectBlackbox {
		return fmt.Errorf("invalid ABC object kind %d", kind)
	}
	old := o.Kind
	if old == kind {
		return nil
	}
	if old < ObjectLatch || kind < ObjectLatch {
		return fmt.Errorf("object kind change must stay within box kinds")
	}
	oldOrder := n.kindOrder[old]
	found := false
	for _, candidate := range oldOrder {
		if candidate == o {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("object kind index is inconsistent")
	}
	n.kindOrder[old] = removeObjectOrder(oldOrder, o)
	o.Kind = kind
	newOrder := append(n.kindOrder[kind], o)
	for i := len(newOrder) - 1; i > 0 && newOrder[i].ID < newOrder[i-1].ID; i-- {
		newOrder[i], newOrder[i-1] = newOrder[i-1], newOrder[i]
	}
	n.kindOrder[kind] = newOrder
	return nil
}
func (n *Network) contains(o *Object) bool { return o != nil && o.network == n && n.Object(o.ID) == o }
func (n *Network) AddFanin(o *Object, e Edge) error {
	if !n.contains(o) || !n.contains(e.Object) {
		return fmt.Errorf("fanin objects must belong to network")
	}
	if o.Kind == ObjectPI || o.Kind == ObjectConst1 || e.Object.Kind == ObjectPO {
		return fmt.Errorf("illegal terminal connection")
	}
	if (o.IsCO() || o.Kind == ObjectNet) && len(o.fanins) != 0 {
		return fmt.Errorf("object %d permits one fanin", o.ID)
	}
	clear(n.hashed)
	o.fanins = append(o.fanins, e)
	e.Object.fanouts = append(e.Object.fanouts, o)
	if !o.IsCI() && e.Object.Level+1 > o.Level {
		o.Level = e.Object.Level + 1
	}
	return nil
}
func (n *Network) DeleteFanin(o *Object, index int) error {
	if !n.contains(o) || index < 0 || index >= len(o.fanins) {
		return fmt.Errorf("invalid fanin index")
	}
	source := o.fanins[index].Object
	o.fanins = append(o.fanins[:index], o.fanins[index+1:]...)
	for i, d := range source.fanouts {
		if d == o {
			source.fanouts = append(source.fanouts[:i], source.fanouts[i+1:]...)
			break
		}
	}
	clear(n.hashed)
	return nil
}
func (n *Network) PatchFanin(o *Object, index int, e Edge) error {
	if !n.contains(o) || !n.contains(e.Object) || e.Object.Kind == ObjectPO || index < 0 || index >= len(o.fanins) {
		return fmt.Errorf("invalid fanin replacement")
	}
	old := o.fanins[index].Object
	for i, d := range old.fanouts {
		if d == o {
			old.fanouts = append(old.fanouts[:i], old.fanouts[i+1:]...)
			break
		}
	}
	o.fanins[index] = e
	e.Object.fanouts = append(e.Object.fanouts, o)
	clear(n.hashed)
	return nil
}
func (n *Network) DeleteObject(o *Object) error {
	if !n.contains(o) || o == n.constant {
		return fmt.Errorf("invalid object deletion")
	}
	if len(o.fanouts) > 0 {
		return fmt.Errorf("object still has fanouts")
	}
	if node, ok := o.Data.(*core.DdNode); ok && o.Kind == ObjectNode {
		if manager, ok := n.FunctionManager.(*core.DdManager); ok {
			core.Cudd_RecursiveDeref(manager, node)
		}
	}
	for len(o.fanins) > 0 {
		if err := n.DeleteFanin(o, len(o.fanins)-1); err != nil {
			return err
		}
	}
	nm.Nm_ManDeleteIdName(n.names, int32(o.ID))
	delete(n.mvVars, o)
	n.objects[o.ID] = nil
	n.kindOrder[o.Kind] = removeObjectOrder(n.kindOrder[o.Kind], o)
	n.cis = removeObjectOrder(n.cis, o)
	n.cos = removeObjectOrder(n.cos, o)
	n.boxes = removeObjectOrder(n.boxes, o)
	o.network = nil
	return nil
}

// Topological returns the transitive fanin cone of roots, detecting cycles.
func (n *Network) Topological(roots []*Object) ([]*Object, error) {
	state := make(map[*Object]uint8)
	var order []*Object
	var visit func(*Object) error
	visit = func(o *Object) error {
		if !n.contains(o) {
			return fmt.Errorf("foreign object in traversal")
		}
		if state[o] == 1 {
			return fmt.Errorf("cycle at object %d", o.ID)
		}
		if state[o] == 2 {
			return nil
		}
		state[o] = 1
		for _, e := range o.fanins {
			if o.IsCI() {
				break
			}
			if err := visit(e.Object); err != nil {
				return err
			}
		}
		state[o] = 2
		order = append(order, o)
		return nil
	}
	for _, o := range roots {
		if err := visit(o); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// Check verifies reciprocal edges, object ownership and terminal arities.
func (n *Network) Check() error {
	counts := make(map[[2]*Object]int)
	for _, o := range n.Objects() {
		if !n.contains(o) {
			return fmt.Errorf("invalid ownership")
		}
		if (o.Kind == ObjectPI || o.Kind == ObjectConst1) && len(o.fanins) != 0 {
			return fmt.Errorf("input has fanins")
		}
		if (o.IsCO() || o.Kind == ObjectNet) && len(o.fanins) > 1 {
			return fmt.Errorf("output has multiple fanins")
		}
		for _, e := range o.fanins {
			if !n.contains(e.Object) {
				return fmt.Errorf("foreign fanin")
			}
			counts[[2]*Object{e.Object, o}]++
		}
	}
	for _, o := range n.Objects() {
		for _, d := range o.fanouts {
			key := [2]*Object{o, d}
			counts[key]--
		}
	}
	for _, count := range counts {
		if count != 0 {
			return fmt.Errorf("fanin/fanout mismatch")
		}
	}
	_, err := n.Topological(n.Objects())
	return err
}
