package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"slices"
)

// Abc_Aig_t owns structural hashing for an ABC object network. It deliberately
// keeps ABC Object identities rather than moving nodes into a second graph.
type Abc_Aig_t struct {
	Network *Network
	// FreeNodeCuts releases cached cuts before an AND is removed or rewired.
	FreeNodeCuts     func(id int32)
	bins             int
	tracking         bool
	added, updated   []*Object
	replacementMarks faninMarks
	pending          map[*Object]int
	excluded         map[*Object]bool
}

func Abc_HashKey2(a, b any, size int32) uint32 {
	if size <= 0 {
		panic("invalid hash table size")
	}
	x, y := objectEdge(a), objectEdge(b)
	key := uint32(x.Object.ID)*7937 ^ uint32(y.Object.ID)*2971
	if x.Complement {
		key ^= 911
	}
	if y.Complement {
		key ^= 353
	}
	return key % uint32(size)
}
func Abc_AigAlloc(n *Network) *Abc_Aig_t {
	return &Abc_Aig_t{Network: n, bins: 10007, pending: make(map[*Object]int), excluded: make(map[*Object]bool)}
}
func Abc_AigFree(m *Abc_Aig_t) {
	if len(m.pending) != 0 {
		panic("pending replacements")
	}
	m.added = nil
	m.updated = nil
	m.replacementMarks = faninMarks{}
	m.tracking = false
	clear(m.Network.hashed)
	clear(m.Network.duplicateAnds)
	m.Network.extraAnds = 0
}
func aigNode(o *Object) bool {
	return o != nil && o.Kind == ObjectNode && o.Data == aigANDMarker && len(o.fanins) == 2
}
func Abc_AigCleanup(m *Abc_Aig_t) int32 {
	before := Abc_NtkNodeNum(m.Network)
	for _, o := range m.Network.ObjectsOfKind(ObjectNode) {
		if m.Network.contains(o) && aigNode(o) && !m.excluded[o] && len(o.fanouts) == 0 {
			Abc_AigDeleteNode(m, o)
		}
	}
	return before - Abc_NtkNodeNum(m.Network)
}
func Abc_AigCheck(m *Abc_Aig_t) int32 {
	if m.Network.Check() != nil || !checkKindOrder(m.Network, ObjectNode) {
		return 0
	}
	return checkAigNodes(m)
}

func checkAigNodes(m *Abc_Aig_t) int32 {
	for o := range m.Network.logicNodes {
		if !aigNode(o) || m.excluded[o] {
			return 0
		}
		if o.Choice != nil {
			choices := map[*Object]bool{o: true}
			for choice := o.Choice; choice != nil; choice = choice.Choice {
				if !m.Network.contains(choice) || choices[choice] || len(choice.fanouts) > 0 {
					return 0
				}
				choices[choice] = true
			}
		}
		a, b := o.fanins[0], o.fanins[1]
		if a.Object == b.Object || a.Object == m.Network.constant || b.Object == m.Network.constant || Abc_ObjFanoutNum(a.Object) == 0 || Abc_ObjFanoutNum(b.Object) == 0 {
			return 0
		}
		if m.Network.hashed[andHashKey(edgePair(a, b))] == o.Signal() {
			continue
		}
		e, ok := m.Network.lookupAnd(a, b)
		if !ok || e.Object != o || e.Complement {
			return 0
		}
	}
	return 1
}
func Abc_AigLevel(n *Network) int32 {
	max := 0
	for _, o := range n.Objects() {
		if o.IsCO() && len(o.fanins) > 0 && o.fanins[0].Object.Level > max {
			max = o.fanins[0].Object.Level
		}
	}
	return int32(max)
}
func appendUnique(objects *[]*Object, o *Object) {
	for _, old := range *objects {
		if old == o {
			return
		}
	}
	*objects = append(*objects, o)
}
func aigSetNode(o *Object, a, b Edge) {
	if a.Object.ID > b.Object.ID {
		a, b = b, a
	}
	setAndPins(o, a, b)
	o.Copy = nil
}
func Abc_AigAndCreate(m *Abc_Aig_t, a, b any) Edge {
	x, y := objectEdge(a), objectEdge(b)
	n := m.Network
	n.checkAndInputs(x, y)
	if len(n.hashed)+n.extraAnds > 2*m.bins {
		Abc_AigResize(m)
	}
	o := n.newAndObject()
	aigSetNode(o, x, y)
	e := o.Signal()
	n.registerAnd(o)
	if m.tracking {
		m.added = append(m.added, o)
	}
	return e
}
func Abc_AigAndCreateFrom(m *Abc_Aig_t, a, b any, o *Object) Edge {
	x, y := objectEdge(a), objectEdge(b)
	m.Network.checkAndInputs(x, y)
	if !m.Network.contains(o) || o.Kind != ObjectNode || len(o.fanins) != 0 {
		panic("recreation requires live empty node")
	}
	delete(m.excluded, o)
	aigSetNode(o, x, y)
	m.Network.registerAnd(o)
	return o.Signal()
}
func Abc_AigAndLookup(m *Abc_Aig_t, a, b any) Edge {
	return m.LookupAndEdges(objectEdge(a), objectEdge(b))
}

// LookupAndEdges resolves a phased pair in this manager's active hash table.
func (m *Abc_Aig_t) LookupAndEdges(x, y Edge) Edge {
	e, _ := m.Network.lookupAndAllowed(x, y, m.excluded)
	return e
}
func Abc_AigAnd(m *Abc_Aig_t, a, b any) Edge {
	if e := Abc_AigAndLookup(m, a, b); e.Object != nil {
		return e
	}
	return Abc_AigAndCreate(m, a, b)
}
func Abc_AigOr(m *Abc_Aig_t, a, b any) Edge {
	return Abc_AigAnd(m, objectEdge(a).Not(), objectEdge(b).Not()).Not()
}
func Abc_AigXor(m *Abc_Aig_t, a, b any) Edge {
	x, y := objectEdge(a), objectEdge(b)
	return Abc_AigOr(m, Abc_AigAnd(m, x, y.Not()), Abc_AigAnd(m, x.Not(), y))
}
func Abc_AigMux(m *Abc_Aig_t, c, t, e any) Edge {
	control := objectEdge(c)
	return Abc_AigOr(m, Abc_AigAnd(m, control, t), Abc_AigAnd(m, control.Not(), e))
}
func Abc_AigConst1(n *Network) Edge { return n.One() }
func Abc_AigXorLookup(m *Abc_Aig_t, a, b any, typ *int32) Edge {
	if typ != nil {
		*typ = 0
	}
	x, y := objectEdge(a), objectEdge(b)
	l, r := Abc_AigAndLookup(m, x.Not(), y.Not()), Abc_AigAndLookup(m, x, y)
	if l.Object != nil && r.Object != nil {
		out := Abc_AigAndLookup(m, l.Not(), r.Not())
		if out.Object != nil && typ != nil {
			*typ = 1
		}
		return out
	}
	l, r = Abc_AigAndLookup(m, x, y.Not()), Abc_AigAndLookup(m, x.Not(), y)
	if l.Object != nil && r.Object != nil {
		out := Abc_AigAndLookup(m, l.Not(), r.Not())
		if out.Object != nil {
			return out.Not()
		}
	}
	return Edge{}
}
func Abc_AigMuxLookup(m *Abc_Aig_t, c, t, e any, typ *int32) Edge {
	if typ != nil {
		*typ = 0
	}
	control, yes, no := objectEdge(c), objectEdge(t), objectEdge(e)
	l, r := Abc_AigAndLookup(m, control, yes.Not()), Abc_AigAndLookup(m, control.Not(), no.Not())
	if l.Object != nil && r.Object != nil {
		out := Abc_AigAndLookup(m, l.Not(), r.Not())
		if out.Object != nil && typ != nil {
			*typ = 1
		}
		return out
	}
	l, r = Abc_AigAndLookup(m, control, yes), Abc_AigAndLookup(m, control.Not(), no)
	if l.Object != nil && r.Object != nil {
		out := Abc_AigAndLookup(m, l.Not(), r.Not())
		if out.Object != nil {
			return out.Not()
		}
	}
	return Edge{}
}
func Abc_AigAndDelete(m *Abc_Aig_t, o *Object) {
	m.Network.detachAnd(o)
	m.excluded[o] = true
	if m.FreeNodeCuts != nil {
		m.FreeNodeCuts(int32(o.ID))
	}
}

func Abc_AigResize(m *Abc_Aig_t) {
	m.bins = int(util.Abc_PrimeCudd(uint32(3 * m.bins)))
	for key, group := range m.Network.duplicateAnds {
		slices.Reverse(group)
		m.Network.hashed[key] = group[len(group)-1].Signal()
	}
}

func Abc_AigRehash(m *Abc_Aig_t) {
	n := m.Network
	old := n.duplicateAnds
	var tracked map[*Object]bool
	if len(old) > 0 {
		tracked = make(map[*Object]bool)
		for _, group := range old {
			for _, o := range group {
				tracked[o] = true
			}
		}
	}
	n.duplicateAnds = nil
	n.extraAnds = 0
	clear(n.hashed)
	for o := range n.logicNodes {
		if aigNode(o) && !m.excluded[o] {
			pair := edgePair(o.fanins[0], o.fanins[1])
			o.fanins[0], o.fanins[1] = pair[0], pair[1]
			if !tracked[o] {
				n.registerAnd(o)
			}
		}
	}
	// The C rehash prepends each entry while walking the old bucket head first.
	for _, group := range old {
		for i := len(group) - 1; i >= 0; i-- {
			o := group[i]
			if n.contains(o) && aigNode(o) && !m.excluded[o] {
				n.registerAnd(o)
			}
		}
	}
}

func Abc_AigMiter_rec(m *Abc_Aig_t, objects []Edge, count int32) Edge {
	if count < 0 || int(count) > len(objects) {
		panic("invalid miter size")
	}
	if count == 0 {
		return m.Network.One().Not()
	}
	if count == 1 {
		return objects[0]
	}
	middle := count / 2
	return Abc_AigOr(m, Abc_AigMiter_rec(m, objects, middle), Abc_AigMiter_rec(m, objects[middle:], count-middle))
}
func Abc_AigMiter(m *Abc_Aig_t, pairs *[]Edge, implication int32) Edge {
	if len(*pairs)%2 != 0 {
		panic("miter requires pairs")
	}
	outputs := make([]Edge, 0, len(*pairs)/2)
	for i := 0; i < len(*pairs); i += 2 {
		a, b := (*pairs)[i], (*pairs)[i+1]
		if implication != 0 {
			outputs = append(outputs, Abc_AigAnd(m, a, b.Not()))
		} else {
			outputs = append(outputs, Abc_AigXor(m, a, b))
		}
	}
	*pairs = outputs
	return Abc_AigMiter_rec(m, outputs, int32(len(outputs)))
}
func Abc_AigMiter2(m *Abc_Aig_t, pairs []Edge) Edge {
	if len(pairs)%2 != 0 {
		panic("miter requires pairs")
	}
	out := m.Network.One().Not()
	for i := 0; i < len(pairs); i += 2 {
		out = Abc_AigOr(m, out, Abc_AigXor(m, pairs[i], pairs[i+1]))
	}
	return out
}

type aigReplacement struct {
	old *Object
	new Edge
}

func Abc_AigReplace(m *Abc_Aig_t, old *Object, replacement any, update int32) {
	n := m.Network
	target := objectEdge(replacement)
	if !n.contains(old) || !n.contains(target.Object) || old == n.constant || old == target.Object {
		panic("invalid AIG replacement")
	}
	marks := &m.replacementMarks
	marks.begin(len(n.objects))
	if err := n.visitFanin(target.Object, marks, nil); err != nil {
		panic(err)
	}
	if marks.dense[old.ID] == marks.active+1 {
		panic("cyclic AIG replacement")
	}
	queue := []aigReplacement{{old, target}}
	m.pending[target.Object]++
	for len(queue) > 0 {
		item := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		m.pending[item.new.Object]--
		if m.pending[item.new.Object] == 0 {
			delete(m.pending, item.new.Object)
		}
		if !n.contains(item.old) {
			continue
		}
		visited := map[*Object]bool{}
		for _, fanout := range item.old.Fanouts() {
			if visited[fanout] {
				continue
			}
			visited[fanout] = true
			if fanout.IsCO() {
				for i, e := range fanout.fanins {
					if e.Object == item.old {
						r := item.new
						r.Complement = r.Complement != e.Complement
						if err := n.PatchFanin(fanout, i, r); err != nil {
							panic(err)
						}
					}
				}
				continue
			}
			if !aigNode(fanout) {
				panic("AIG replacement encountered non-AND fanout")
			}
			a, b := fanout.fanins[0], fanout.fanins[1]
			if a.Object == item.old {
				a = Edge{item.new.Object, a.Complement != item.new.Complement}
			}
			if b.Object == item.old {
				b = Edge{item.new.Object, b.Complement != item.new.Complement}
			}
			if found := Abc_AigAndLookup(m, a, b); found.Object != nil && found.Object != fanout {
				queue = append(queue, aigReplacement{fanout, found})
				m.pending[found.Object]++
				continue
			}
			Abc_AigAndDelete(m, fanout)
			Abc_ObjRemoveFanins(fanout)
			Abc_AigAndCreateFrom(m, a, b, fanout)
		}
		if n.contains(item.old) && len(item.old.fanouts) == 0 {
			Abc_AigDeleteNode(m, item.old)
		}
	}
	if update != 0 {
		Abc_AigUpdateLevel_int(m)
		Abc_AigUpdateLevelR_int(m)
	}
}
func Abc_AigReplace_int(m *Abc_Aig_t, old *Object, replacement any, update int32) {
	Abc_AigReplace(m, old, replacement, update)
}
func Abc_AigDeleteNode(m *Abc_Aig_t, o *Object) {
	n := m.Network
	if !n.contains(o) {
		return
	}
	if !aigNode(o) || len(o.fanouts) != 0 {
		panic("AIG deletion requires dangling AND")
	}
	if m.pending[o] > 0 {
		return
	}
	fanins := [2]*Object{o.fanins[0].Object, o.fanins[1].Object}
	if m.tracking {
		for _, input := range fanins {
			appendUnique(&m.updated, input)
		}
	}
	Abc_AigAndDelete(m, o)
	Abc_NtkDeleteObj(o)
	delete(m.excluded, o)
	for _, input := range fanins {
		if n.contains(input) && aigNode(input) && len(input.fanouts) == 0 {
			Abc_AigDeleteNode(m, input)
		}
	}
}
func Abc_AigUpdateLevel_int(m *Abc_Aig_t) {
	order, err := m.Network.Topological(m.Network.Objects())
	if err != nil {
		panic(err)
	}
	for _, o := range order {
		o.Level = 0
		if o.IsCI() {
			continue
		}
		for _, e := range o.fanins {
			if e.Object.Level+1 > o.Level {
				o.Level = e.Object.Level + 1
			}
		}
		o.MarkA = false
	}
}
func Abc_AigUpdateLevelR_int(m *Abc_Aig_t) {
	order, err := m.Network.Topological(m.Network.Objects())
	if err != nil {
		panic(err)
	}
	for i := len(order) - 1; i >= 0; i-- {
		o := order[i]
		o.reverseLevel = 0
		for _, fanout := range o.fanouts {
			if fanout.IsCI() {
				continue
			}
			if fanout.reverseLevel+1 > o.reverseLevel {
				o.reverseLevel = fanout.reverseLevel + 1
			}
		}
		o.MarkB = false
	}
}
func removeLevelEntry(levels *[][]*Object, o *Object) {
	for i := range *levels {
		for j, entry := range (*levels)[i] {
			if entry == o {
				(*levels)[i][j] = nil
				return
			}
		}
	}
	panic("node missing from level structure")
}
func Abc_AigRemoveFromLevelStructure(levels *[][]*Object, o *Object) {
	removeLevelEntry(levels, o)
	o.MarkA = false
}
func Abc_AigRemoveFromLevelStructureR(levels *[][]*Object, o *Object) {
	removeLevelEntry(levels, o)
	o.MarkB = false
}
func Abc_AigNodeHasComplFanoutEdge(o *Object) int32     { return aigComplFanout(o, false) }
func Abc_AigNodeHasComplFanoutEdgeTrav(o *Object) int32 { return aigComplFanout(o, true) }
func aigComplFanout(o *Object, current bool) int32 {
	for _, fanout := range o.fanouts {
		if current && Abc_NodeIsTravIdCurrent(fanout) == 0 {
			continue
		}
		for _, e := range fanout.fanins {
			if e.Object == o && e.Complement {
				return 1
			}
		}
	}
	return 0
}
func Abc_AigPrintNode(value any) {
	e := objectEdge(value)
	o := e.Object
	if o.IsCI() {
		fmt.Printf("CI %s complement=%v\n", o.Name(), e.Complement)
		return
	}
	if o.Kind == ObjectConst1 {
		fmt.Printf("Constant 1 complement=%v\n", e.Complement)
		return
	}
	fmt.Printf("%s complement=%v = %v\n", o.Name(), e.Complement, o.Fanins())
}
func Abc_AigNodeIsAcyclic(o, root *Object) int32 {
	seen := map[*Object]bool{}
	var visit func(*Object) bool
	visit = func(n *Object) bool {
		if n == root {
			return false
		}
		if seen[n] || n.IsCI() {
			return true
		}
		seen[n] = true
		for _, e := range n.fanins {
			if !visit(e.Object) {
				return false
			}
		}
		return true
	}
	for _, e := range o.fanins {
		if !visit(e.Object) {
			return 0
		}
	}
	return 1
}
func Abc_AigCheckFaninOrder(m *Abc_Aig_t) {
	for _, o := range m.Network.ObjectsOfKind(ObjectNode) {
		if aigNode(o) && o.fanins[0].Object.ID > o.fanins[1].Object.ID {
			panic("unordered AIG fanins")
		}
	}
}
func Abc_AigSetNodePhases(n *Network) {
	order, err := n.Topological(n.Objects())
	if err != nil {
		panic(err)
	}
	for _, o := range order {
		switch {
		case o.Kind == ObjectConst1:
			o.Phase = true
		case o.IsCI():
			o.Phase = false
			if o.Kind == ObjectBO && len(o.fanins) > 0 {
				o.Phase = o.fanins[0].Object.Data == ABC_INIT_ONE
			}
		case aigNode(o):
			o.Phase = (o.fanins[0].Object.Phase != o.fanins[0].Complement) && (o.fanins[1].Object.Phase != o.fanins[1].Complement)
		case o.IsCO() && len(o.fanins) == 1:
			o.Phase = o.fanins[0].Object.Phase != o.fanins[0].Complement
		}
	}
}
func Abc_AigUpdateStart(m *Abc_Aig_t, updated **[]*Object) *[]*Object {
	if m.tracking {
		panic("updates already active")
	}
	m.tracking = true
	m.added = []*Object{}
	m.updated = []*Object{}
	if updated != nil {
		*updated = &m.updated
	}
	return &m.added
}
func Abc_AigUpdateStop(m *Abc_Aig_t) {
	if !m.tracking {
		panic("updates not active")
	}
	m.tracking = false
	m.added = nil
	m.updated = nil
}
func Abc_AigUpdateReset(m *Abc_Aig_t) {
	if !m.tracking {
		panic("updates not active")
	}
	clear(m.added)
	clear(m.updated)
	m.added = m.added[:0]
	m.updated = m.updated[:0]
}
func Abc_AigCountNext(m *Abc_Aig_t) int32 {
	bins := map[uint32]int{}
	count := 0
	for _, o := range m.Network.ObjectsOfKind(ObjectNode) {
		if aigNode(o) && !m.excluded[o] {
			bins[Abc_HashKey2(o.fanins[0], o.fanins[1], int32(m.bins))]++
			count++
		}
	}
	collisions := int32(count - len(bins))
	fmt.Printf("Counter = %d. Nodes = %d. Ave = %.2f\n", collisions, count, float64(count)/float64(m.bins))
	return collisions
}
func Abc_NtkHelloWorld(_ *Network) { fmt.Println("Hello, World!") }
