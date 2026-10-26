package abc

// Derived from ABC src/base/abc/{abc.h,abcInt.h,abcObj.c,abcNtk.c,abcFanio.c}.

import (
	"errors"
	"fmt"
	"slices"
)

var ErrABCNetworkClosed = errors.New("ABC network is closed")

// ABCNetworkType preserves Abc_NtkType_t.
type ABCNetworkType uint8

const (
	ABCNetworkNone ABCNetworkType = iota
	ABCNetworkNetlist
	ABCNetworkLogic
	ABCNetworkStrash
	ABCNetworkOther
)

// ABCNetworkFunction preserves Abc_NtkFunc_t.
type ABCNetworkFunction uint8

const (
	ABCFunctionNone ABCNetworkFunction = iota
	ABCFunctionSOP
	ABCFunctionBDD
	ABCFunctionAIG
	ABCFunctionMap
	ABCFunctionBLIFMV
	ABCFunctionBlackbox
	ABCFunctionOther
)

// ABCObjectType preserves Abc_ObjType_t.
type ABCObjectType uint8

const (
	ABCObjectNone ABCObjectType = iota
	ABCObjectConst1
	ABCObjectPI
	ABCObjectPO
	ABCObjectBI
	ABCObjectBO
	ABCObjectNet
	ABCObjectNode
	ABCObjectLatch
	ABCObjectWhitebox
	ABCObjectBlackbox
	abcObjectTypeCount
)

// ABCLatchInit preserves Abc_InitType_t.
type ABCLatchInit uint8

const (
	ABCInitNone ABCLatchInit = iota
	ABCInitZero
	ABCInitOne
	ABCInitDC
	ABCInitOther
)

// ABCObjectEdge replaces pointer tagging with an explicit object ID and phase.
type ABCObjectEdge struct {
	ID           int
	Complemented bool
}

func (e ABCObjectEdge) Not() ABCObjectEdge {
	e.Complemented = !e.Complemented
	return e
}

// ABCNetworkLimits bounds retained state and permits deterministic cancellation.
type ABCNetworkLimits struct {
	MaxObjects   int
	MaxEdges     int
	MaxNameBytes int
	Stop         func() bool
}

func (l ABCNetworkLimits) normalized() ABCNetworkLimits {
	if l.MaxObjects <= 0 {
		l.MaxObjects = 1 << 20
	}
	if l.MaxEdges <= 0 {
		l.MaxEdges = 1 << 22
	}
	if l.MaxNameBytes <= 0 {
		l.MaxNameBytes = 1 << 24
	}
	return l
}

// ABCNetworkObject is an owned snapshot of the source object representation.
type ABCNetworkObject struct {
	ID       int
	Type     ABCObjectType
	Name     string
	Function string
	Init     ABCLatchInit
	Level    uint32
	Phase    bool
	MarkA    bool
	MarkB    bool
	MarkC    bool
	Fanins   []ABCObjectEdge
	Fanouts  []int
	Alive    bool
}

func (o ABCNetworkObject) clone() ABCNetworkObject {
	o.Fanins = slices.Clone(o.Fanins)
	o.Fanouts = slices.Clone(o.Fanouts)
	return o
}

// ABCNetwork is the concrete owned counterpart of Abc_Ntk_t. Object IDs are
// stable and deleted slots are retained as tombstones, matching vObjs.
type ABCNetwork struct {
	Type         ABCNetworkType
	Function     ABCNetworkFunction
	Name         string
	Spec         string
	Step         int
	Constraints  int
	AndGateDelay float64

	objects                   []ABCNetworkObject
	pis, pos, cis, cos, boxes []int
	counts                    [abcObjectTypeCount]int
	edges                     int
	nameBytes                 int
	nameToID                  map[string]int
	limits                    ABCNetworkLimits
	closed                    bool
}

func NewABCNetwork(typ ABCNetworkType, fn ABCNetworkFunction, limits ABCNetworkLimits) (*ABCNetwork, error) {
	if typ <= ABCNetworkNone || typ >= ABCNetworkOther {
		return nil, fmt.Errorf("abc network type: %w", ErrDimension)
	}
	if fn <= ABCFunctionNone || fn >= ABCFunctionOther {
		return nil, fmt.Errorf("abc network function: %w", ErrDimension)
	}
	if typ == ABCNetworkStrash && fn != ABCFunctionAIG {
		return nil, fmt.Errorf("strash requires AIG functionality: %w", ErrDimension)
	}
	n := &ABCNetwork{Type: typ, Function: fn, limits: limits.normalized(), nameToID: make(map[string]int)}
	if typ == ABCNetworkStrash {
		n.objects = append(n.objects, ABCNetworkObject{ID: 0, Type: ABCObjectConst1, Alive: true})
		n.counts[ABCObjectConst1] = 1
	} else {
		n.objects = append(n.objects, ABCNetworkObject{ID: 0, Type: ABCObjectNone})
	}
	return n, nil
}

func (n *ABCNetwork) checkOpen() error {
	if n == nil || n.closed {
		return ErrABCNetworkClosed
	}
	if n.limits.Stop != nil && n.limits.Stop() {
		return ErrCancelled
	}
	return nil
}

func (n *ABCNetwork) live(id int) (*ABCNetworkObject, error) {
	if err := n.checkOpen(); err != nil {
		return nil, err
	}
	if id < 0 || id >= len(n.objects) || !n.objects[id].Alive {
		return nil, fmt.Errorf("ABC object %d: %w", id, ErrDimension)
	}
	return &n.objects[id], nil
}

func validABCObjectType(t ABCObjectType) bool { return t >= ABCObjectConst1 && t < abcObjectTypeCount }
func abcIsCI(t ABCObjectType) bool            { return t == ABCObjectPI || t == ABCObjectBO }
func abcIsCO(t ABCObjectType) bool            { return t == ABCObjectPO || t == ABCObjectBI }
func abcIsBox(t ABCObjectType) bool {
	return t == ABCObjectLatch || t == ABCObjectWhitebox || t == ABCObjectBlackbox
}

func (n *ABCNetwork) CreateObject(t ABCObjectType) (int, error) {
	if err := n.checkOpen(); err != nil {
		return -1, err
	}
	if !validABCObjectType(t) || t == ABCObjectConst1 {
		return -1, fmt.Errorf("ABC object type: %w", ErrDimension)
	}
	if len(n.objects) >= n.limits.MaxObjects {
		return -1, ErrLimit
	}
	id := len(n.objects)
	o := ABCNetworkObject{ID: id, Type: t, Alive: true}
	if t == ABCObjectLatch {
		o.Init = ABCInitNone
	}
	n.objects = append(n.objects, o)
	n.counts[t]++
	switch t {
	case ABCObjectPI:
		n.pis = append(n.pis, id)
		n.cis = append(n.cis, id)
	case ABCObjectPO:
		n.pos = append(n.pos, id)
		n.cos = append(n.cos, id)
	case ABCObjectBI:
		n.cos = append(n.cos, id)
	case ABCObjectBO:
		n.cis = append(n.cis, id)
	case ABCObjectLatch, ABCObjectWhitebox, ABCObjectBlackbox:
		n.boxes = append(n.boxes, id)
	}
	return id, nil
}

func (n *ABCNetwork) CreatePI() (int, error)   { return n.CreateObject(ABCObjectPI) }
func (n *ABCNetwork) CreatePO() (int, error)   { return n.CreateObject(ABCObjectPO) }
func (n *ABCNetwork) CreateBI() (int, error)   { return n.CreateObject(ABCObjectBI) }
func (n *ABCNetwork) CreateBO() (int, error)   { return n.CreateObject(ABCObjectBO) }
func (n *ABCNetwork) CreateNet() (int, error)  { return n.CreateObject(ABCObjectNet) }
func (n *ABCNetwork) CreateNode() (int, error) { return n.CreateObject(ABCObjectNode) }

// CreateLatchBox creates latch, BI and BO in source object order.
func (n *ABCNetwork) CreateLatchBox(init ABCLatchInit) (bi, latch, bo int, err error) {
	if init > ABCInitDC {
		return -1, -1, -1, fmt.Errorf("latch initialization: %w", ErrDimension)
	}
	if err = n.checkOpen(); err != nil {
		return -1, -1, -1, err
	}
	if len(n.objects)+3 > n.limits.MaxObjects || n.edges+2 > n.limits.MaxEdges {
		return -1, -1, -1, ErrLimit
	}
	base := n.Clone()
	if latch, err = n.CreateObject(ABCObjectLatch); err == nil {
		bi, err = n.CreateBI()
	}
	if err == nil {
		bo, err = n.CreateBO()
	}
	if err == nil {
		n.objects[latch].Init = init
		err = n.AddFanin(latch, ABCObjectEdge{ID: bi})
	}
	if err == nil {
		err = n.AddFanin(bo, ABCObjectEdge{ID: latch})
	}
	if err != nil {
		n.restore(base)
		return -1, -1, -1, err
	}
	return bi, latch, bo, nil
}

func (n *ABCNetwork) SetObjectName(id int, name string) error {
	o, err := n.live(id)
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("empty ABC object name: %w", ErrDimension)
	}
	if old, ok := n.nameToID[name]; ok && old != id {
		return fmt.Errorf("duplicate ABC object name: %w", ErrDimension)
	}
	newBytes := n.nameBytes - len(o.Name) + len(name)
	if newBytes > n.limits.MaxNameBytes {
		return ErrLimit
	}
	if o.Name != "" {
		delete(n.nameToID, o.Name)
	}
	o.Name = name
	n.nameToID[name] = id
	n.nameBytes = newBytes
	return nil
}

func (n *ABCNetwork) FindObject(name string) (int, bool) {
	if n == nil || n.closed {
		return -1, false
	}
	id, ok := n.nameToID[name]
	return id, ok
}

func (n *ABCNetwork) Object(id int) (ABCNetworkObject, error) {
	o, err := n.live(id)
	if err != nil {
		return ABCNetworkObject{}, err
	}
	return o.clone(), nil
}

func removeIntOnce(v []int, x int) ([]int, bool) {
	for i, y := range v {
		if y == x {
			return append(v[:i], v[i+1:]...), true
		}
	}
	return v, false
}

func (n *ABCNetwork) AddFanin(id int, edge ABCObjectEdge) error {
	o, err := n.live(id)
	if err != nil {
		return err
	}
	f, err := n.live(edge.ID)
	if err != nil {
		return err
	}
	if id == edge.ID || o.Type == ABCObjectConst1 || o.Type == ABCObjectPI || f.Type == ABCObjectPO {
		return fmt.Errorf("invalid ABC edge: %w", ErrDimension)
	}
	if (abcIsCO(o.Type) || o.Type == ABCObjectNet) && len(o.Fanins) != 0 {
		return fmt.Errorf("single-driver ABC object: %w", ErrDimension)
	}
	if n.edges >= n.limits.MaxEdges {
		return ErrLimit
	}
	o.Fanins = append(o.Fanins, edge)
	f.Fanouts = append(f.Fanouts, id)
	n.edges++
	return nil
}

func (n *ABCNetwork) DeleteFanin(id, fanin int) error {
	o, err := n.live(id)
	if err != nil {
		return err
	}
	f, err := n.live(fanin)
	if err != nil {
		return err
	}
	index := -1
	for i, e := range o.Fanins {
		if e.ID == fanin {
			index = i
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("fanin not found: %w", ErrDimension)
	}
	var ok bool
	f.Fanouts, ok = removeIntOnce(f.Fanouts, id)
	if !ok {
		return fmt.Errorf("fanout backlink missing: %w", ErrInvariant)
	}
	o.Fanins = append(o.Fanins[:index], o.Fanins[index+1:]...)
	n.edges--
	return nil
}

func (n *ABCNetwork) RemoveFanins(id int) error {
	o, err := n.live(id)
	if err != nil {
		return err
	}
	backup := n.Clone()
	for len(o.Fanins) > 0 {
		if err := n.DeleteFanin(id, o.Fanins[len(o.Fanins)-1].ID); err != nil {
			n.restore(backup)
			return err
		}
		o = &n.objects[id]
	}
	return nil
}

func (n *ABCNetwork) PatchFanin(id, oldFanin int, replacement ABCObjectEdge) error {
	o, err := n.live(id)
	if err != nil {
		return err
	}
	old, err := n.live(oldFanin)
	if err != nil {
		return err
	}
	newObj, err := n.live(replacement.ID)
	if err != nil {
		return err
	}
	if oldFanin == replacement.ID || id == replacement.ID || newObj.Type == ABCObjectPO {
		return fmt.Errorf("invalid replacement: %w", ErrDimension)
	}
	idx := -1
	for i, e := range o.Fanins {
		if e.ID == oldFanin {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("old fanin not found: %w", ErrDimension)
	}
	oldBack := slices.Index(old.Fanouts, id)
	if oldBack < 0 {
		return fmt.Errorf("old backlink missing: %w", ErrInvariant)
	}
	o.Fanins[idx] = ABCObjectEdge{ID: replacement.ID, Complemented: o.Fanins[idx].Complemented != replacement.Complemented}
	old.Fanouts = append(old.Fanouts[:oldBack], old.Fanouts[oldBack+1:]...)
	newObj.Fanouts = append(newObj.Fanouts, id)
	return nil
}

func (n *ABCNetwork) InsertBetween(fanin, fanout int, typ ABCObjectType) (int, error) {
	if typ != ABCObjectNode && typ != ABCObjectNet && typ != ABCObjectBI && typ != ABCObjectBO {
		return -1, fmt.Errorf("insert type: %w", ErrDimension)
	}
	if _, err := n.live(fanin); err != nil {
		return -1, err
	}
	out, err := n.live(fanout)
	if err != nil {
		return -1, err
	}
	idx := -1
	for i, e := range out.Fanins {
		if e.ID == fanin {
			idx = i
			break
		}
	}
	if idx < 0 {
		return -1, fmt.Errorf("edge not found: %w", ErrDimension)
	}
	if len(n.objects) >= n.limits.MaxObjects || n.edges+1 > n.limits.MaxEdges {
		return -1, ErrLimit
	}
	backup := n.Clone()
	id, err := n.CreateObject(typ)
	if err == nil {
		err = n.PatchFanin(fanout, fanin, ABCObjectEdge{ID: id})
	}
	if err == nil {
		err = n.AddFanin(id, ABCObjectEdge{ID: fanin})
	}
	if err != nil {
		n.restore(backup)
		return -1, err
	}
	return id, nil
}

func (n *ABCNetwork) TransferFanouts(from, to int) error {
	f, err := n.live(from)
	if err != nil {
		return err
	}
	if _, err = n.live(to); err != nil {
		return err
	}
	if from == to {
		return fmt.Errorf("same transfer object: %w", ErrDimension)
	}
	outs := slices.Clone(f.Fanouts)
	backup := n.Clone()
	for _, out := range outs {
		if err = n.PatchFanin(out, from, ABCObjectEdge{ID: to}); err != nil {
			n.restore(backup)
			return err
		}
	}
	return nil
}

func removeID(v []int, id int) []int {
	for i, x := range v {
		if x == id {
			return append(v[:i], v[i+1:]...)
		}
	}
	return v
}

func (n *ABCNetwork) DeleteObject(id int) error {
	o, err := n.live(id)
	if err != nil {
		return err
	}
	if o.Type == ABCObjectConst1 {
		return fmt.Errorf("delete distinguished constant: %w", ErrDimension)
	}
	backup := n.Clone()
	for len(o.Fanouts) > 0 {
		out := o.Fanouts[len(o.Fanouts)-1]
		if err = n.DeleteFanin(out, id); err != nil {
			n.restore(backup)
			return err
		}
		o = &n.objects[id]
	}
	if err = n.RemoveFanins(id); err != nil {
		n.restore(backup)
		return err
	}
	o = &n.objects[id]
	if o.Name != "" {
		delete(n.nameToID, o.Name)
		n.nameBytes -= len(o.Name)
	}
	n.pis = removeID(n.pis, id)
	n.pos = removeID(n.pos, id)
	n.cis = removeID(n.cis, id)
	n.cos = removeID(n.cos, id)
	n.boxes = removeID(n.boxes, id)
	n.counts[o.Type]--
	*o = ABCNetworkObject{ID: id, Type: ABCObjectNone}
	return nil
}

func (n *ABCNetwork) ReplaceObject(oldID, newID int) error {
	backup := n.Clone()
	if err := n.TransferFanouts(oldID, newID); err != nil {
		return err
	}
	if err := n.DeleteObject(oldID); err != nil {
		n.restore(backup)
		return err
	}
	return nil
}

func (n *ABCNetwork) ReplaceObjectByConstant(id int, one bool) (int, error) {
	o, err := n.live(id)
	if err != nil {
		return -1, err
	}
	if abcIsCO(o.Type) {
		return -1, fmt.Errorf("replace combinational output: %w", ErrDimension)
	}
	backup := n.Clone()
	var constant int
	if one {
		constant, err = n.CreateConst1()
	} else {
		constant, err = n.CreateConst0()
	}
	if err == nil {
		err = n.TransferFanouts(id, constant)
	}
	if err == nil && o.Type == ABCObjectNode {
		err = n.DeleteObject(id)
	}
	if err != nil {
		n.restore(backup)
		return -1, err
	}
	return constant, nil
}

func (n *ABCNetwork) FanoutFaninIndex(fanout, fanin int) (int, error) {
	o, err := n.live(fanout)
	if err != nil {
		return -1, err
	}
	if _, err = n.live(fanin); err != nil {
		return -1, err
	}
	for i, e := range o.Fanins {
		if e.ID == fanin {
			return i, nil
		}
	}
	return -1, fmt.Errorf("fanin not found: %w", ErrDimension)
}

func (n *ABCNetwork) SetNodeFunction(id int, function string) error {
	o, err := n.live(id)
	if err != nil {
		return err
	}
	if o.Type != ABCObjectNode {
		return fmt.Errorf("function on non-node: %w", ErrDimension)
	}
	o.Function = function
	return nil
}

func (n *ABCNetwork) CreateLogicNode(function string, fanins ...ABCObjectEdge) (int, error) {
	if err := n.checkOpen(); err != nil {
		return -1, err
	}
	if len(n.objects)+1 > n.limits.MaxObjects || n.edges+len(fanins) > n.limits.MaxEdges {
		return -1, ErrLimit
	}
	backup := n.Clone()
	id, err := n.CreateNode()
	if err == nil {
		n.objects[id].Function = function
	}
	for _, e := range fanins {
		if err == nil {
			err = n.AddFanin(id, e)
		}
	}
	if err != nil {
		n.restore(backup)
		return -1, err
	}
	return id, nil
}

func (n *ABCNetwork) CreateConst0() (int, error)                { return n.CreateLogicNode(" 0\n") }
func (n *ABCNetwork) CreateConst1() (int, error)                { return n.CreateLogicNode(" 1\n") }
func (n *ABCNetwork) CreateBuffer(e ABCObjectEdge) (int, error) { return n.CreateLogicNode("1 1\n", e) }
func (n *ABCNetwork) CreateInverter(e ABCObjectEdge) (int, error) {
	return n.CreateLogicNode("0 1\n", e)
}
func (n *ABCNetwork) CreateAND(es ...ABCObjectEdge) (int, error) {
	return n.CreateLogicNode("and", es...)
}
func (n *ABCNetwork) CreateOR(es ...ABCObjectEdge) (int, error) {
	return n.CreateLogicNode("or", es...)
}
func (n *ABCNetwork) CreateXOR(es ...ABCObjectEdge) (int, error) {
	return n.CreateLogicNode("xor", es...)
}
func (n *ABCNetwork) CreateMUX(c, t, e ABCObjectEdge) (int, error) {
	return n.CreateLogicNode("mux", c, t, e)
}

func (n *ABCNetwork) ObjectCount() int {
	if n == nil || n.closed {
		return 0
	}
	total := 0
	for _, c := range n.counts {
		total += c
	}
	return total
}
func (n *ABCNetwork) ObjectCountByType(t ABCObjectType) int {
	if n == nil || int(t) >= len(n.counts) {
		return 0
	}
	return n.counts[t]
}
func (n *ABCNetwork) PIs() []int {
	if n == nil {
		return nil
	}
	return slices.Clone(n.pis)
}
func (n *ABCNetwork) POs() []int {
	if n == nil {
		return nil
	}
	return slices.Clone(n.pos)
}
func (n *ABCNetwork) CIs() []int {
	if n == nil {
		return nil
	}
	return slices.Clone(n.cis)
}
func (n *ABCNetwork) COs() []int {
	if n == nil {
		return nil
	}
	return slices.Clone(n.cos)
}
func (n *ABCNetwork) Boxes() []int {
	if n == nil {
		return nil
	}
	return slices.Clone(n.boxes)
}

func (n *ABCNetwork) Clone() *ABCNetwork {
	if n == nil {
		return nil
	}
	c := *n
	c.objects = make([]ABCNetworkObject, len(n.objects))
	for i, o := range n.objects {
		c.objects[i] = o.clone()
	}
	c.pis = slices.Clone(n.pis)
	c.pos = slices.Clone(n.pos)
	c.cis = slices.Clone(n.cis)
	c.cos = slices.Clone(n.cos)
	c.boxes = slices.Clone(n.boxes)
	c.nameToID = make(map[string]int, len(n.nameToID))
	for k, v := range n.nameToID {
		c.nameToID[k] = v
	}
	return &c
}

func (n *ABCNetwork) restore(c *ABCNetwork) {
	if c != nil {
		*n = *c
	}
}

func (n *ABCNetwork) Validate() error {
	if err := n.checkOpen(); err != nil {
		return err
	}
	counts := [abcObjectTypeCount]int{}
	edges := 0
	names := 0
	back := make(map[[2]int]int)
	forward := make(map[[2]int]int)
	for id, o := range n.objects {
		if !o.Alive {
			continue
		}
		if o.ID != id || !validABCObjectType(o.Type) {
			return fmt.Errorf("object identity/type: %w", ErrInvariant)
		}
		counts[o.Type]++
		if o.Name != "" {
			if n.nameToID[o.Name] != id {
				return fmt.Errorf("name index: %w", ErrInvariant)
			}
			names += len(o.Name)
		}
		if o.Type == ABCObjectPI && len(o.Fanins) != 0 {
			return fmt.Errorf("PI has fanin: %w", ErrInvariant)
		}
		if o.Type == ABCObjectConst1 && (id != 0 || n.Type != ABCNetworkStrash || len(o.Fanins) != 0) {
			return fmt.Errorf("distinguished constant: %w", ErrInvariant)
		}
		if (abcIsCO(o.Type) || o.Type == ABCObjectNet) && len(o.Fanins) > 1 {
			return fmt.Errorf("multiple drivers: %w", ErrInvariant)
		}
		for _, e := range o.Fanins {
			if e.ID < 0 || e.ID >= len(n.objects) || !n.objects[e.ID].Alive {
				return fmt.Errorf("dangling fanin: %w", ErrInvariant)
			}
			if n.objects[e.ID].Type == ABCObjectPO {
				return fmt.Errorf("PO fanout: %w", ErrInvariant)
			}
			forward[[2]int{e.ID, id}]++
			edges++
		}
		for _, out := range o.Fanouts {
			if out < 0 || out >= len(n.objects) || !n.objects[out].Alive {
				return fmt.Errorf("dangling fanout: %w", ErrInvariant)
			}
			back[[2]int{id, out}]++
		}
	}
	if counts != n.counts || edges != n.edges || names != n.nameBytes {
		return fmt.Errorf("network counters: %w", ErrInvariant)
	}
	if (n.Type == ABCNetworkStrash) != (counts[ABCObjectConst1] == 1) {
		return fmt.Errorf("constant count: %w", ErrInvariant)
	}
	if len(forward) != len(back) {
		return fmt.Errorf("adjacency keys: %w", ErrInvariant)
	}
	for k, v := range forward {
		if back[k] != v {
			return fmt.Errorf("adjacency multiplicity: %w", ErrInvariant)
		}
	}
	if !equalIntSet(n.pis, n, func(t ABCObjectType) bool { return t == ABCObjectPI }) ||
		!equalIntSet(n.pos, n, func(t ABCObjectType) bool { return t == ABCObjectPO }) ||
		!equalIntSet(n.cis, n, abcIsCI) || !equalIntSet(n.cos, n, abcIsCO) ||
		!equalIntSet(n.boxes, n, abcIsBox) {
		return fmt.Errorf("terminal partition: %w", ErrInvariant)
	}
	named := 0
	for _, o := range n.objects {
		if o.Alive && o.Name != "" {
			named++
		}
	}
	if named != len(n.nameToID) {
		return fmt.Errorf("name cardinality: %w", ErrInvariant)
	}
	return nil
}

func equalIntSet(ids []int, n *ABCNetwork, accept func(ABCObjectType) bool) bool {
	want := 0
	for _, object := range n.objects {
		if object.Alive && accept(object.Type) {
			want++
		}
	}
	if len(ids) != want {
		return false
	}
	seen := make([]bool, len(n.objects))
	for _, id := range ids {
		if id < 0 || id >= len(n.objects) || seen[id] || !n.objects[id].Alive || !accept(n.objects[id].Type) {
			return false
		}
		seen[id] = true
	}
	return true
}

func (n *ABCNetwork) Close() {
	if n == nil || n.closed {
		return
	}
	n.closed = true
	n.objects = nil
	n.pis = nil
	n.pos = nil
	n.cis = nil
	n.cos = nil
	n.boxes = nil
	n.nameToID = nil
	n.edges = 0
	n.nameBytes = 0
}
