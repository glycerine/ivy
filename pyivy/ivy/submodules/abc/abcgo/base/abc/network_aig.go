package abc

import "fmt"

func edgePair(a, b Edge) [2]Edge {
	if a.Object.ID > b.Object.ID || (a.Object.ID == b.Object.ID && a.Complement && !b.Complement) {
		a, b = b, a
	}
	return [2]Edge{a, b}
}
func (n *Network) One() Edge { return n.constant.Signal() }
func (n *Network) lookupAnd(a, b Edge) (Edge, bool) {
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
	pair := edgePair(a, b)
	if e, ok := n.hashed[pair]; ok {
		o := e.Object
		if n.contains(o) && o.Data == aigANDMarker && len(o.fanins) == 2 && edgePair(o.fanins[0], o.fanins[1]) == pair {
			return e, true
		}
		delete(n.hashed, pair)
	}
	for _, o := range n.objects {
		if o != nil && o.Kind == ObjectNode && o.Data == aigANDMarker && len(o.fanins) == 2 && edgePair(o.fanins[0], o.fanins[1]) == pair {
			e := o.Signal()
			n.hashed[pair] = e
			return e, true
		}
	}
	return Edge{}, false
}

const aigANDMarker = "abcgo:AIG AND"

func (n *Network) CreateAndEdges(a, b Edge) Edge {
	if !n.contains(a.Object) || !n.contains(b.Object) {
		panic("foreign AIG input")
	}
	o := n.NewObject(ObjectNode)
	o.Data = aigANDMarker
	if err := n.AddFanin(o, a); err != nil {
		panic(err)
	}
	if err := n.AddFanin(o, b); err != nil {
		panic(err)
	}
	return o.Signal()
}
func (n *Network) AndEdges(a, b Edge) Edge {
	if e, ok := n.lookupAnd(a, b); ok {
		return e
	}
	e := n.CreateAndEdges(a, b)
	n.hashed[edgePair(a, b)] = e
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

func (d DecNetwork) ConstantOne() any     { return d.One() }
func (d DecNetwork) Complement(a any) any { return a.(Edge).Not() }
func (d DecNetwork) And(a, b any) any {
	x, y := a.(Edge), b.(Edge)
	if found, ok := d.lookupAnd(x, y); ok && Abc_NodeIsTravIdCurrent(found.Object) != 0 {
		// A replacement graph must not retain a structural-hash hit from the
		// root's current MFFC: that node is about to be removed. ABC counts
		// such hits as newly added nodes and rebuilds them during graph mapping.
		fresh := d.CreateAndEdges(x, y)
		d.hashed[edgePair(x, y)] = fresh
		return fresh
	}
	return d.AndEdges(x, y)
}
func (d DecNetwork) CreateAnd(a, b any) any { return d.CreateAndEdges(a.(Edge), b.(Edge)) }
func (d DecNetwork) Variable(i int) any     { pis := d.ObjectsOfKind(ObjectPI); return pis[i].Signal() }
func (d DecNetwork) LookupAnd(a, b any) any {
	if e, ok := d.lookupAnd(a.(Edge), b.(Edge)); ok {
		return e
	}
	return nil
}
func (d DecNetwork) SameRegular(a, b any) bool { return a.(Edge).Object == b.(Edge).Object }
func (d DecNetwork) Level(a any) int           { return a.(Edge).Object.Level }
func (d DecNetwork) InDeletionCone(a any) bool { return a.(Edge).Object.MarkA }
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
