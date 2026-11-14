package gia

import "fmt"

// Objects exports the original Gia_Obj_t encoding, including explicit CO nodes.
func (g *Graph) Objects() []*Gia_Obj_t {
	objects := make([]*Gia_Obj_t, len(g.nodes), len(g.nodes)+len(g.outputs))
	for i, n := range g.nodes {
		p := &Gia_Obj_t{iDiff0: objectNoFanin, iDiff1: objectNoFanin}
		switch n.kind {
		case graphInput:
			p.fTerm = 1
		case graphAnd:
			a, b := n.a, n.b
			if a > b {
				a, b = b, a
			}
			p.iDiff0 = uint32(i) - a.Object()
			p.fCompl0 = uint32(a & 1)
			p.iDiff1 = uint32(i) - b.Object()
			p.fCompl1 = uint32(b & 1)
		}
		objects[i] = p
	}
	for i, l := range g.inputs {
		objects[l.Object()].iDiff1 = uint32(i)
	}
	for i, l := range g.outputs {
		objects = append(objects, &Gia_Obj_t{fTerm: 1, iDiff0: uint32(len(objects)) - l.Object(), iDiff1: uint32(i), fCompl0: uint32(l & 1)})
	}
	store := &objectStore{objects: &objects}
	for id, p := range objects {
		p.id = int32(id)
		p.store = store
	}
	return objects
}

// GraphFromObjects imports the original object storage without pointer tagging
// or unsafe arithmetic. Specialized real XOR/MUX nodes require their companion
// manager arrays and are rejected rather than interpreted as plain AND nodes.
func GraphFromObjects(objects []*Gia_Obj_t) (*Graph, error) {
	if len(objects) == 0 || objects[0] == nil || objects[0].fTerm != 0 || Gia_ObjIsConst0(objects[0]) == 0 {
		return nil, fmt.Errorf("gia: missing constant object")
	}
	g := NewGraph()
	copies := make([]Literal, len(objects))
	edge := func(id int, diff, complement uint32) (Literal, error) {
		if diff == 0 || diff > uint32(id) {
			return 0, fmt.Errorf("gia: invalid backward fanin at object %d", id)
		}
		target := uint32(id) - diff
		if objects[target] == nil || Gia_ObjIsCo(objects[target]) != 0 {
			return 0, fmt.Errorf("gia: invalid fanin target at object %d", id)
		}
		return copies[target] ^ Literal(complement&1), nil
	}
	for id, p := range objects[1:] {
		id++
		if p == nil {
			return nil, fmt.Errorf("gia: nil object %d", id)
		}
		if Gia_ObjIsCi(p) != 0 {
			copies[id] = g.NewInput()
			continue
		}
		a, err := edge(id, p.iDiff0, p.fCompl0)
		if err != nil {
			return nil, err
		}
		if Gia_ObjIsCo(p) != 0 {
			g.AddOutput(a)
			copies[id] = a
			continue
		}
		if p.store != nil && p.store.muxes != nil && p.id >= 0 && int(p.id) < len(*p.store.muxes) && (*p.store.muxes)[p.id] != 0 {
			return nil, fmt.Errorf("gia: specialized MUX object %d requires normalization", id)
		}
		if Gia_ObjIsXor(p) != 0 {
			return nil, fmt.Errorf("gia: specialized XOR object %d requires normalization", id)
		}
		b, err := edge(id, p.iDiff1, p.fCompl1)
		if err != nil {
			return nil, err
		}
		copies[id] = g.And(a, b)
	}
	return g, nil
}
