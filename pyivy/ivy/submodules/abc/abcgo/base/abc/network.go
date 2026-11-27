package abc

import (
	"fmt"
	"slices"
	"sync/atomic"

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

// DecValue returns this Edge as the value used by decomposition and rewriting.
// Repeated queries share immutable boxed edges in both phases.
func (e Edge) DecValue() any {
	if e.Object == nil {
		return e
	}
	values := e.Object.decValues.Load()
	if values == nil {
		signal := e.Object.Signal()
		values = &[2]any{signal, signal.Not()}
		if !e.Object.decValues.CompareAndSwap(nil, values) {
			values = e.Object.decValues.Load()
		}
	}
	if e.Complement {
		return values[1]
	}
	return values[0]
}

// objectAux stores rarely used object fields out of line.
type objectAux struct {
	mvEncoding []Edge
	// boxInputPolarity records the optional input-complement mask used by
	// ABC's hierarchical gate strashing interface.
	boxInputPolarity []uint32
	terminalIndex    int
	dTemp            float64
	crossCutVisits   int
}

// extra returns o's auxiliary fields, allocating them on first use.
func (o *Object) extra() *objectAux {
	if o.aux == nil {
		o.aux = new(objectAux)
	}
	return o.aux
}

// MVEncoding returns a multi-valued net's per-value signals.
func (o *Object) MVEncoding() []Edge {
	if o.aux == nil {
		return nil
	}
	return o.aux.mvEncoding
}
func (o *Object) SetMVEncoding(e []Edge) {
	if e != nil || o.aux != nil {
		o.extra().mvEncoding = e
	}
}

// BoxInputPolarity returns the box's optional input-complement mask.
func (o *Object) BoxInputPolarity() []uint32 {
	if o.aux == nil {
		return nil
	}
	return o.aux.boxInputPolarity
}
func (o *Object) SetBoxInputPolarity(p []uint32) {
	if p != nil || o.aux != nil {
		o.extra().boxInputPolarity = p
	}
}

// TerminalIndex is the position of a terminal within its box or model.
func (o *Object) TerminalIndex() int {
	if o.aux == nil {
		return 0
	}
	return o.aux.terminalIndex
}
func (o *Object) SetTerminalIndex(i int) {
	if i != 0 || o.aux != nil {
		o.extra().terminalIndex = i
	}
}

// DTemp is ABC's temporary floating-point object field.
func (o *Object) DTemp() float64 {
	if o.aux == nil {
		return 0
	}
	return o.aux.dTemp
}
func (o *Object) SetDTemp(d float64) {
	if d != 0 || o.aux != nil {
		o.extra().dTemp = d
	}
}

// Object is a network object. Fanin order and fanout multiplicity are retained.
type Object struct {
	// aux holds fields few objects use, keeping AIG cells small.
	aux                 *objectAux
	CopyEdge            Edge
	CopyData            any // temporary non-object copy payload (e.g. a mapped gate)
	fanoutAdjustment    int
	Choice              *Object
	Next                *Object
	Phase               bool
	Persist             bool // ABC's fPersist flag
	Barrier             bool
	MarkA, MarkB, MarkC bool
	Kind                ObjectKind
	reverseLevel        int
	travID              uint64
	ID                  int
	nodeSlot            int
	Level               int
	Temp                int
	Data                any
	Copy                *Object
	decValues           atomic.Pointer[[2]any]
	network             *Network
	fanins              []Edge
	fanouts             []*Object
}

func (o *Object) Network() *Network  { return o.network }
func (o *Object) Signal() Edge       { return Edge{Object: o} }
func (o *Object) Fanins() []Edge     { return append([]Edge(nil), o.fanins...) }
func (o *Object) Fanouts() []*Object { return append([]*Object(nil), o.fanouts...) }

// FaninNum, FaninAt, FanoutCount and FanoutAt read the live fanin and
// fanout arrays without copying them, as o.FaninNum(), o.FaninAt(int32(i)),
// o.FanoutCount() and o.FanoutAt(int32(i)) do with a copy.
func (o *Object) FaninNum() int            { return len(o.fanins) }
func (o *Object) FaninAt(i int32) Edge     { return o.fanins[i] }
func (o *Object) FanoutCount() int         { return len(o.fanouts) }
func (o *Object) FanoutAt(i int32) *Object { return o.fanouts[i] }
func (o *Object) Name() string             { return nm.Nm_ManFindNameById(o.network.names, int32(o.ID)) }
func (o *Object) SetName(name string) {
	nm.Nm_ManDeleteIdName(o.network.names, int32(o.ID))
	nm.Nm_ManStoreIdName(o.network.names, int32(o.ID), int32(o.Kind), name, "")
}
func (o *Object) IsCI() bool { return o.Kind == ObjectPI || o.Kind == ObjectBO }
func (o *Object) IsCO() bool { return o.Kind == ObjectPO || o.Kind == ObjectBI }

// Network owns objects with stable IDs; deletion leaves holes, as in ABC.
// Mutations must be serialized by the caller.
type Network struct {
	Faults []int32
	// vFins stores fault assignments as alternating object ID / fault type
	// entries, matching Abc_Ntk_t's source vFins field.
	vFins                     *vec.Vec_Int_t
	Constraints               int32
	Exdc, Excare              *Network
	Onehots                   [][]int32
	SequenceModel             *util.Abc_Cex_t
	seqModelVec               *vec.Vec_Ptr_t
	ObjectPermutation, Phases []int32
	NameIDs                   []int32
	Topo                      []int32
	AndGateDelay              float64
	MappingTiming             *MappingTiming
	// TimingManager stores source-compatible per-network timing state owned by
	// clients such as base/abci. It is kept separate from MappingTiming, which
	// carries only the mapper's CI/CO scalar annotations.
	TimingManager any
	// lutTimes is ABC's temporary per-object arrival/required/slack storage.
	lutTimes        []float32
	WireLoad        string
	LTLProperties   []string
	Model           []int32
	Data            any
	Attributes      []*vec.Vec_Att_t
	Step            int32
	DTemp           float64
	FunctionManager any
	timingMapping   any
	phaseTracking   bool
	mvVars          map[*Object]*MVVariable
	Design          *Design
	Copy, AltView   *Network
	Type, Function  uint8
	Barriers        int
	barBufs2        int32
	// CILevel supplies timing-derived initial levels when a timing manager is attached.
	CILevel             func(*Object) int
	kindOrder           [ObjectBlackbox + 1][]*Object
	nodeHoles           int
	cis, cos, boxes     []*Object
	travID              uint64
	levelMax            int
	reverseLevelsActive bool
	Name, Spec          string
	objects             []*Object
	names               *nm.Manager
	hashed              andTable
	duplicateAnds       map[[2]int][]*Object
	extraAnds           int
	constant            *Object
	andCells            []andCell
	andSlab             int
}

func NewNetwork(name string) *Network {
	n := &Network{Step: -1, DTemp: -1, travID: 1, Type: 3, Function: 3, Name: name, names: nm.Nm_ManCreate(16)}
	n.constant = n.NewObject(ObjectConst1)
	n.constant.Phase = true
	return n
}

// NewAIGNetworkWithCapacity creates an empty strashed network with room for
// the supplied terminal counts and estimated ANDs. Hashing may fold most gates,
// so limit the speculative reservation until the network actually grows.
func NewAIGNetworkWithCapacity(name string, inputs, latches, outputs, ands int) *Network {
	n := NewNetwork(name)
	n.ReserveAIG(inputs, latches, outputs, ands)
	return n
}

// ReserveAnds sizes the structural hash table and the object arrays for up
// to ands gates, for callers that know an upper bound (restrashing an AIG
// can only fold gates away).
func (n *Network) ReserveAnds(ands int) {
	if ands <= 0 {
		return
	}
	if n.hashed.len() == 0 {
		n.hashed.reserve(ands)
	}
	n.objects = slices.Grow(n.objects, ands)
	n.kindOrder[ObjectNode] = slices.Grow(n.kindOrder[ObjectNode], ands)
}

// ReserveAIG reserves room for the specified additional terminals and gates.
// Bound speculative gate storage because structural hashing can fold gates away.
func (n *Network) ReserveAIG(inputs, latches, outputs, ands int) {
	if inputs < 0 || latches < 0 || outputs < 0 || ands < 0 {
		panic("negative AIG capacity")
	}
	gates := min(ands, 4096)
	if gates > cap(n.kindOrder[ObjectNode])-len(n.kindOrder[ObjectNode]) && n.hashed.len() == 0 {
		n.hashed.reserve(gates)
	}
	n.objects = slices.Grow(n.objects, inputs+3*latches+outputs+gates)
	n.kindOrder[ObjectNode] = slices.Grow(n.kindOrder[ObjectNode], gates)
	n.kindOrder[ObjectPI] = slices.Grow(n.kindOrder[ObjectPI], inputs)
	n.kindOrder[ObjectPO] = slices.Grow(n.kindOrder[ObjectPO], outputs)
	n.cis = slices.Grow(n.cis, inputs+latches)
	n.cos = slices.Grow(n.cos, outputs+latches)
}

func (n *Network) NewObject(kind ObjectKind) *Object {
	return n.addObject(new(Object), kind)
}

// An AIG gate has two fanins and usually only one or two fanouts. Keep the
// initial backing arrays in the same allocation as its stable Object handle.
// andCell keeps an AND's first fanins and fanouts beside its header.
type andCell struct {
	Object
	inputs [2]Edge
	uses   [2]*Object
}

// newAndObject takes cells from slabs, like ABC's fixed-size object pool.
// Slabs double up to 512 cells, so small networks stay small.
func (n *Network) newAndObject() *Object {
	if len(n.andCells) == 0 {
		n.andSlab = min(512, max(8, 2*n.andSlab))
		n.andCells = make([]andCell, n.andSlab)
	}
	cell := &n.andCells[0]
	n.andCells = n.andCells[1:]
	cell.fanins, cell.fanouts = cell.inputs[:0], cell.uses[:0]
	return n.addObject(&cell.Object, ObjectNode)
}

func (n *Network) addObject(o *Object, kind ObjectKind) *Object {
	if kind <= ObjectNone || kind > ObjectBlackbox {
		panic("invalid ABC object kind")
	}
	o.ID, o.Kind, o.network = len(n.objects), kind, n
	n.objects = append(n.objects, o)
	order := n.kindOrder[kind]
	if kind == ObjectNode {
		o.nodeSlot = len(order)
	}
	n.kindOrder[kind] = append(order, o)
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
	if kind > ObjectBlackbox {
		return nil
	}
	if kind == ObjectNode && n.nodeHoles != 0 {
		nodes := make([]*Object, 0, len(n.kindOrder[kind])-n.nodeHoles)
		for o := range n.logicNodes {
			nodes = append(nodes, o)
		}
		return nodes
	}
	return append([]*Object(nil), n.kindOrder[kind]...)
}

// logicNodes visits live nodes in creation order. Deletions leave empty slots
// until half the directory is empty, at which point the live prefix is packed.
func (n *Network) logicNodes(yield func(*Object) bool) {
	for _, o := range n.kindOrder[ObjectNode] {
		if o != nil && !yield(o) {
			return
		}
	}
}

func (n *Network) removeLogicNode(o *Object) {
	order := n.kindOrder[ObjectNode]
	order[o.nodeSlot] = nil
	n.nodeHoles++
	if n.nodeHoles < len(order)-n.nodeHoles {
		return
	}
	next := 0
	for _, live := range order {
		if live != nil {
			order[next] = live
			live.nodeSlot = next
			next++
		}
	}
	clear(order[next:])
	n.kindOrder[ObjectNode] = order[:next]
	n.nodeHoles = 0
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
	if old > ObjectBlackbox {
		return fmt.Errorf("object kind index is inconsistent")
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
	n.detachAnd(o)
	o.Kind = kind
	n.changedAnd(o)
	newOrder := append(n.kindOrder[kind], o)
	for i := len(newOrder) - 1; i > 0 && newOrder[i].ID < newOrder[i-1].ID; i-- {
		newOrder[i], newOrder[i-1] = newOrder[i-1], newOrder[i]
	}
	n.kindOrder[kind] = newOrder
	return nil
}
func (n *Network) contains(o *Object) bool { return o != nil && o.network == n && n.Object(o.ID) == o }

// Once a two-slot array is replaced, an AND's allocation still contains it.
func (o *Object) addFanout(use *Object) {
	old := o.fanouts
	o.fanouts = append(old, use)
	if len(old) == 2 && cap(old) == 2 {
		clear(old)
	}
}

func clearInitialPair[T any](items []T) {
	if cap(items) == 2 {
		clear(items[:2])
	}
}

func (o *Object) replaceFanins(inputs []Edge) {
	clearInitialPair(o.fanins)
	o.fanins = inputs
}

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
	n.detachAnd(o)
	old := o.fanins
	o.fanins = append(old, e)
	if len(old) == 2 && cap(old) == 2 {
		clear(old)
	}
	n.changedAnd(o)
	e.Object.addFanout(o)
	if !o.IsCI() && e.Object.Level+1 > o.Level {
		o.Level = e.Object.Level + 1
	}
	return nil
}
func (n *Network) DeleteFanin(o *Object, index int) error {
	if !n.contains(o) || index < 0 || index >= len(o.fanins) {
		return fmt.Errorf("invalid fanin index")
	}
	n.detachAnd(o)
	source := o.fanins[index].Object
	o.fanins = slices.Delete(o.fanins, index, index+1)
	source.fanouts = removeObjectOrder(source.fanouts, o)
	n.changedAnd(o)
	return nil
}
func (n *Network) PatchFanin(o *Object, index int, e Edge) error {
	if !n.contains(o) || !n.contains(e.Object) || e.Object.Kind == ObjectPO || index < 0 || index >= len(o.fanins) {
		return fmt.Errorf("invalid fanin replacement")
	}
	changed := o.fanins[index] != e
	if changed {
		n.detachAnd(o)
	}
	old := o.fanins[index].Object
	old.fanouts = removeObjectOrder(old.fanouts, o)
	o.fanins[index] = e
	e.Object.addFanout(o)
	if changed {
		n.changedAnd(o)
	}
	return nil
}
func (n *Network) DeleteObject(o *Object) error {
	if !n.contains(o) || o == n.constant || o.Kind <= ObjectNone || o.Kind > ObjectBlackbox {
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
	if o.Kind == ObjectNode {
		n.removeLogicNode(o)
	} else {
		n.kindOrder[o.Kind] = removeObjectOrder(n.kindOrder[o.Kind], o)
	}
	switch {
	case o.IsCI():
		n.cis = removeObjectOrder(n.cis, o)
	case o.IsCO():
		n.cos = removeObjectOrder(n.cos, o)
	case Abc_ObjIsBox(o) != 0:
		n.boxes = removeObjectOrder(n.boxes, o)
	}
	o.network = nil
	return nil
}

// A separate generation for each traversal avoids mutating objects' shared
// traversal IDs. A serialized AIG editor reuses these marks across replacements.
type faninMarks struct {
	dense  []uint8
	sparse map[int]*[256]uint8
	active uint8
}

func (s *faninMarks) begin(size int) {
	if size > len(s.dense) {
		s.dense = slices.Grow(s.dense, size-len(s.dense))[:size]
	}
	s.active += 2
	if s.active == 0 {
		clear(s.dense)
		s.active = 2
	}
}

func (s *faninMarks) cell(id int) *uint8 {
	if s.sparse == nil {
		return &s.dense[id]
	}
	page := s.sparse[id/256]
	if page == nil {
		page = new([256]uint8)
		s.sparse[id/256] = page
	}
	return &page[id%256]
}

func (n *Network) visitFanin(o *Object, marks *faninMarks, order *[]*Object) error {
	if !n.contains(o) {
		return fmt.Errorf("foreign object in traversal")
	}
	color := marks.cell(o.ID)
	if *color == marks.active {
		return fmt.Errorf("cycle at object %d", o.ID)
	}
	if *color == marks.active+1 {
		return nil
	}
	*color = marks.active
	if !o.IsCI() {
		for _, e := range o.fanins {
			if err := n.visitFanin(e.Object, marks, order); err != nil {
				return err
			}
		}
	}
	*color = marks.active + 1
	if order != nil {
		*order = append(*order, o)
	}
	return nil
}

// Topological returns the transitive fanin cone of roots, detecting cycles.
func (n *Network) Topological(roots []*Object) ([]*Object, error) {
	if len(roots) == 0 {
		return nil, nil
	}
	if n == nil {
		return nil, fmt.Errorf("foreign object in traversal")
	}
	marks := faninMarks{active: 1}
	// A small cone can occupy a handful of widely separated IDs.
	if len(n.objects) > 65536 && len(roots) < len(n.objects)/256 {
		marks.sparse = make(map[int]*[256]uint8)
	} else {
		marks.dense = make([]uint8, len(n.objects))
	}
	capacity := min(len(roots), len(n.objects))
	if len(roots) == 1 && marks.sparse == nil && n.contains(roots[0]) {
		// A deep root implies a long output even when it is the only root.
		// Double the depth estimate to accommodate branching fanins.
		depth := max(0, min(roots[0].Level, (len(n.objects)-1)/2))
		capacity = max(capacity, depth*2+1)
	}
	order := make([]*Object, 0, capacity)
	for _, o := range roots {
		if err := n.visitFanin(o, &marks, &order); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// Check verifies reciprocal edges, object ownership and terminal arities.
func (n *Network) Check() error {
	// Group each fanin by source so its expected destinations can be
	// compared with the fanout list in linear time, including duplicate pins.
	offsets := make([]int, len(n.objects)+1)
	fanins := 0
	for id, o := range n.objects {
		offsets[id+1] = offsets[id]
		if o == nil {
			continue
		}
		if o.network != n || o.ID != id {
			return fmt.Errorf("invalid ownership")
		}
		if (o.Kind == ObjectPI || o.Kind == ObjectConst1) && len(o.fanins) != 0 {
			return fmt.Errorf("input has fanins")
		}
		if (o.IsCO() || o.Kind == ObjectNet) && len(o.fanins) > 1 {
			return fmt.Errorf("output has multiple fanins")
		}
		offsets[id+1] += len(o.fanouts)
		fanins += len(o.fanins)
	}
	if fanins != offsets[len(n.objects)] {
		return fmt.Errorf("fanin/fanout mismatch")
	}
	destinations := make([]int, fanins)
	cursor := append([]int(nil), offsets[:len(n.objects)]...)
	for id, o := range n.objects {
		if o == nil {
			continue
		}
		for _, e := range o.fanins {
			if !n.contains(e.Object) {
				return fmt.Errorf("foreign fanin")
			}
			source := e.Object.ID
			if cursor[source] == offsets[source+1] {
				return fmt.Errorf("fanin/fanout mismatch")
			}
			destinations[cursor[source]] = id
			cursor[source]++
		}
	}
	// Equal total sizes and bounded insertion filled every source range.
	counts := cursor
	clear(counts)
	for id, o := range n.objects {
		if o == nil {
			continue
		}
		for _, dest := range destinations[offsets[id]:offsets[id+1]] {
			counts[dest]++
		}
		for _, dest := range o.fanouts {
			if !n.contains(dest) || counts[dest.ID] == 0 {
				return fmt.Errorf("fanin/fanout mismatch")
			}
			counts[dest.ID]--
		}
	}
	// The connection scratch is now free for a topological cycle check.
	return n.checkAcyclic(counts, offsets[:0])
}

func (n *Network) checkAcyclic(remaining, ready []int) error {
	live := 0
	for id, o := range n.objects {
		if o == nil {
			continue
		}
		live++
		if !o.IsCI() {
			remaining[id] = len(o.fanins)
		}
		if remaining[id] == 0 {
			ready = append(ready, id)
		}
	}
	for i := 0; i < len(ready); i++ {
		for _, dest := range n.objects[ready[i]].fanouts {
			if dest.IsCI() {
				continue
			}
			remaining[dest.ID]--
			if remaining[dest.ID] == 0 {
				ready = append(ready, dest.ID)
			}
		}
	}
	if len(ready) != live {
		// Recover the precise offending DFS back edge for the diagnostic.
		_, err := n.Topological(n.Objects())
		return err
	}
	return nil
}
