package abc

import (
	"fmt"
	"slices"
)

func edgePair(a, b Edge) [2]Edge {
	if a.Object.ID > b.Object.ID || (a.Object.ID == b.Object.ID && a.Complement && !b.Complement) {
		a, b = b, a
	}
	return [2]Edge{a, b}
}

// Object IDs are nonnegative. Complementing all ID bits encodes an inverted
// edge without reducing the range of supported IDs.
func andHashKey(pair [2]Edge) [2]int {
	a, b := pair[0].Object.ID, pair[1].Object.ID
	if pair[0].Complement {
		a = ^a
	}
	if pair[1].Complement {
		b = ^b
	}
	return [2]int{a, b}
}
func (n *Network) One() Edge { return n.constant.Signal() }
func (n *Network) lookupAnd(a, b Edge) (Edge, bool) {
	return n.lookupAndAllowed(a, b, nil)
}

func (n *Network) lookupAndAllowed(a, b Edge, excluded map[*Object]bool) (Edge, bool) {

	if !n.contains(a.Object) || !n.contains(b.Object) {
		panic("foreign AIG input")
	}
	one := n.One()
	zero := one.Not()
	switch {
	case a == zero || b == zero:
		return zero, true
	case a == one:
		return b, true
	case b == one:
		return a, true
	case a == b:
		return a, true
	case a == b.Not():
		return zero, true
	}
	if Abc_ObjFanoutNum(a.Object) == 0 || Abc_ObjFanoutNum(b.Object) == 0 {
		return Edge{}, false
	}
	pair := edgePair(a, b)
	key := andHashKey(pair)
	if o := n.cachedAnd(pair, key); o != nil && !excluded[o] {
		return o.Signal(), true
	}
	if found := findAndFanout(pair, excluded); found != nil {
		e := found.Signal()
		n.hashed[key] = e
		return e, true
	}
	return Edge{}, false
}

func (n *Network) cachedAnd(pair [2]Edge, key [2]int) *Object {
	for e := n.hashed[key]; e.Object != nil; e = n.hashed[key] {
		o := e.Object
		if n.contains(o) && aigNode(o) && edgePair(o.fanins[0], o.fanins[1]) == pair {
			return o
		}
		n.removeAndKey(key, o)
	}
	return nil
}

// Either operand's fanouts contain every AND with this key. Duplicate
// unhashed nodes can exist, so retain the ID order of the object-table scan.
func findAndFanout(pair [2]Edge, excluded map[*Object]bool) *Object {
	fanouts := pair[0].Object.fanouts
	if other := pair[1].Object.fanouts; len(other) < len(fanouts) {
		fanouts = other
	}
	var first *Object
	for _, o := range fanouts {
		if aigNode(o) && edgePair(o.fanins[0], o.fanins[1]) == pair && !excluded[o] && (first == nil || first.ID > o.ID) {
			first = o
		}
	}
	return first
}

// A native hash bucket returns the last inserted AND with this fanin pair.
// Only duplicate pairs need additional storage beyond the regular hash entry.
func (n *Network) registerAnd(o *Object) {
	pair := edgePair(o.fanins[0], o.fanins[1])
	key := andHashKey(pair)
	previous := n.cachedAnd(pair, key)
	if previous != nil && previous != o {
		if n.duplicateAnds == nil {
			n.duplicateAnds = make(map[[2]int][]*Object)
		}
		group := n.duplicateAnds[key]
		if len(group) == 0 {
			group = append(group, previous)
		}
		n.duplicateAnds[key] = append(group, o)
		n.extraAnds++
	}
	n.hashed[key] = o.Signal()
}

// Remove the object from its old key before editing its fanins or function.
func (n *Network) detachAnd(o *Object) {
	if len(n.hashed) == 0 || o.Kind != ObjectNode || len(o.fanins) != 2 {
		return
	}
	n.removeAndKey(andHashKey(edgePair(o.fanins[0], o.fanins[1])), o)
}

func (n *Network) removeAndKey(key [2]int, o *Object) {
	if group := n.duplicateAnds[key]; len(group) > 0 {
		if i := slices.Index(group, o); i >= 0 {
			group = slices.Delete(group, i, i+1)
			n.extraAnds--
			if len(group) == 1 {
				delete(n.duplicateAnds, key)
			} else {
				n.duplicateAnds[key] = group
			}
		}
		n.hashed[key] = group[len(group)-1].Signal()
	} else if n.hashed[key].Object == o {
		delete(n.hashed, key)
	}
}

// An edited raw node can precede an already cached fanout in object-ID order.
func (n *Network) changedAnd(o *Object) {
	if len(n.hashed) == 0 || o.ID == len(n.objects)-1 || !aigNode(o) {
		return
	}
	key := andHashKey(edgePair(o.fanins[0], o.fanins[1]))
	if head := n.hashed[key].Object; head != nil && head.ID > o.ID && len(n.duplicateAnds[key]) == 0 {
		delete(n.hashed, key)
	}
}

const aigANDMarker = "abcgo:AIG AND"

func (n *Network) checkAndInputs(a, b Edge) {
	if !n.contains(a.Object) || !n.contains(b.Object) || a.Object.Kind == ObjectPO || b.Object.Kind == ObjectPO {
		panic("invalid AIG input")
	}
}

// Both pins have been validated and the destination is an empty AND node.
// Attach the two fanout occurrences even when both pins use the same driver.
func setAndPins(o *Object, a, b Edge) {
	o.Data = aigANDMarker
	o.fanins = append(o.fanins, a, b)
	a.Object.addFanout(o)
	b.Object.addFanout(o)
	o.Level = 1 + max(a.Object.Level, b.Object.Level)
	o.Phase = (a.Object.Phase != a.Complement) && (b.Object.Phase != b.Complement)
}

func (n *Network) CreateAndEdges(a, b Edge) Edge {
	n.checkAndInputs(a, b)
	o := n.newAndObject()
	setAndPins(o, a, b)
	return o.Signal()
}
func (n *Network) AndEdges(a, b Edge) Edge {
	if e, ok := n.lookupAnd(a, b); ok {
		return e
	}
	e := n.CreateAndEdges(a, b)
	n.hashed[andHashKey(edgePair(a, b))] = e
	return e
}

// Evaluate evaluates an AIG cone, or arbitrary nodes whose Data implements
// func([]bool) bool. Inputs include PIs and box-output combinational inputs.
func (n *Network) Evaluate(root Edge, inputs map[*Object]bool) (bool, error) {
	order, err := n.Topological([]*Object{root.Object})
	if err != nil {
		return false, err
	}
	values := make(map[*Object]bool, len(order))
	for _, o := range order {
		if o.Kind == ObjectConst1 {
			values[o] = true
			continue
		}
		if o.IsCI() {
			value, ok := inputs[o]
			if !ok {
				return false, fmt.Errorf("missing input %d", o.ID)
			}
			values[o] = value
			continue
		}
		args := make([]bool, len(o.fanins))
		for i, e := range o.fanins {
			args[i] = values[e.Object] != e.Complement
		}
		if isBarrierBuffer(o) {
			values[o] = args[0]
			continue
		}
		if o.Kind == ObjectNode {
			if fn, ok := booleanFunction(o); ok {
				values[o] = fn(args)
				continue
			}
		}
		if (o.IsCO() || o.Kind == ObjectNet) && len(args) == 1 {
			values[o] = args[0]
			continue
		}
		return false, fmt.Errorf("object %d has no Boolean function", o.ID)
	}
	return values[root.Object] != root.Complement, nil
}

// DecNetwork adapts explicit edges to bool/dec's dependency-free interfaces.
type DecNetwork struct{ *Network }

func (d DecNetwork) ConstantOne() any     { return d.One().DecValue() }
func (d DecNetwork) Complement(a any) any { return a.(Edge).Not().DecValue() }
func (d DecNetwork) And(a, b any) any {
	if manager, ok := d.FunctionManager.(*Abc_Aig_t); ok {
		return Abc_AigAnd(manager, a, b).DecValue()
	}
	return d.AndEdges(a.(Edge), b.(Edge)).DecValue()
}
func (d DecNetwork) CreateAnd(a, b any) any { return d.CreateAndEdges(a.(Edge), b.(Edge)).DecValue() }
func (d DecNetwork) Variable(i int) any     { return d.kindOrder[ObjectPI][i].Signal().DecValue() }
func (d DecNetwork) lookupEdge(a, b Edge) Edge {
	if manager, ok := d.FunctionManager.(*Abc_Aig_t); ok {
		return manager.LookupAndEdges(a, b)
	}
	e, _ := d.lookupAnd(a, b)
	return e
}
func (d DecNetwork) LookupAnd(a, b any) any {
	if e := d.lookupEdge(a.(Edge), b.(Edge)); e.Object != nil {
		return e.DecValue()
	}
	return nil
}

// LookupAndLevel applies decomposition edge phases to the value operands.
// A trivial result supplies its intrinsic level. An ordinary AND returns -1
// so the caller can compute its level from the two candidate subgraphs.
func (d DecNetwork) LookupAndLevel(a, b any, invertA, invertB bool) (any, int) {
	x, y := a.(Edge), b.(Edge)
	x.Complement = x.Complement != invertA
	y.Complement = y.Complement != invertB
	e := d.lookupEdge(x, y)
	if e.Object == nil {
		return nil, -1
	}
	level := -1
	switch e.Object {
	case d.constant:
		level = 0
	case x.Object, y.Object:
		level = e.Object.Level
	}
	return e.DecValue(), level
}
func (d DecNetwork) SameRegular(a, b any) bool { return a.(Edge).Object == b.(Edge).Object }
func (d DecNetwork) Level(a any) int           { return a.(Edge).Object.Level }
func (d DecNetwork) InDeletionCone(a any) bool { return Abc_NodeIsTravIdCurrent(a.(Edge).Object) != 0 }
func (d DecNetwork) Replace(old, replacement any, updateLevels bool) {
	from, to := old.(Edge), replacement.(Edge)
	if manager, ok := d.FunctionManager.(*Abc_Aig_t); ok {
		update := int32(0)
		if updateLevels {
			update = 1
		}
		Abc_AigReplace(manager, from.Object, to, update)
		return
	}
	if from.Object == to.Object {
		panic("replacement refers to replaced object")
	}
	if !d.contains(from.Object) || !d.contains(to.Object) || from.Object == d.constant {
		panic("invalid replacement object")
	}
	cone, err := d.Topological([]*Object{to.Object})
	if err != nil {
		panic(err)
	}
	for _, o := range cone {
		if o == from.Object {
			panic("replacement depends on replaced object")
		}
	}
	fanouts := from.Object.Fanouts()
	for _, o := range fanouts {
		for i, e := range o.fanins {
			if e.Object == from.Object {
				r := to
				r.Complement = r.Complement != e.Complement != from.Complement
				if err := d.PatchFanin(o, i, r); err != nil {
					panic(err)
				}
			}
		}
	}
	if err := d.DeleteObject(from.Object); err != nil {
		panic(err)
	}
	if updateLevels {
		order, err := d.Topological(d.Objects())
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
		}
	}
}
