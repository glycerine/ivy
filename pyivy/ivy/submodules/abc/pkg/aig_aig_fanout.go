package abc

// Ordered fanout storage from aig/aig/aigFanout.c.

import "slices"

// AigFanout identifies one fanout edge. Pin is zero for unary objects and the
// selected fanin of an AND. Complementation belongs to the fanout object's
// fanin literal and does not change the edge identity.
type AigFanout struct {
	Object AigID
	Pin    uint8
}

// AigFanoutIndex is an owned fanout snapshot. Object and edge order match
// Aig_ManFanoutStart. Add and Remove update only this index, which lets future
// graph-editing passes stage native fanout maintenance without attaching
// mutable auxiliary storage to AigManager.
type AigFanoutIndex struct {
	fanouts [][]AigFanout
	edges   int
}

// BuildAigFanoutIndex implements Aig_ManFanoutStart. Native AIG fanout storage
// excludes buffers; callers needing buffer-aware read-only traversal can use
// ReverseDFS instead.
func BuildAigFanoutIndex(g *AigManager, o AigTraversalOptions) (*AigFanoutIndex, error) {
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return nil, err
	}
	visit := aigTraversalVisit(checked)
	result := &AigFanoutIndex{fanouts: make([][]AigFanout, len(g.objects))}
	for fanoutID, object := range g.objects {
		if err = visit(); err != nil {
			return nil, err
		}
		appendEdge := func(literal AigLit, pin uint8) error {
			if err = visit(); err != nil {
				return err
			}
			source := literal.ID()
			result.fanouts[source] = append(result.fanouts[source], AigFanout{Object: AigID(fanoutID), Pin: pin})
			result.edges++
			return nil
		}
		switch object.Type {
		case AigNone, AigConstant1, AigCI:
		case AigCO:
			if err = appendEdge(object.Fanin0, 0); err != nil {
				return nil, err
			}
		case AigAND:
			if err = appendEdge(object.Fanin0, 0); err != nil {
				return nil, err
			}
			if err = appendEdge(object.Fanin1, 1); err != nil {
				return nil, err
			}
		case AigBuffer:
			return nil, ErrInvariant
		default:
			return nil, ErrDimension
		}
	}
	return result, nil
}

func (f *AigFanoutIndex) ObjectCount() int {
	if f == nil {
		return 0
	}
	return len(f.fanouts)
}

func (f *AigFanoutIndex) EdgeCount() int {
	if f == nil {
		return 0
	}
	return f.edges
}

// Fanouts returns an independent copy in native iteration order.
func (f *AigFanoutIndex) Fanouts(object AigID) ([]AigFanout, error) {
	if f == nil || uint64(object) >= uint64(len(f.fanouts)) {
		return nil, ErrInvalidLiteral
	}
	return slices.Clone(f.fanouts[object]), nil
}

func (f *AigFanoutIndex) Clone() *AigFanoutIndex {
	if f == nil {
		return nil
	}
	copy := &AigFanoutIndex{fanouts: make([][]AigFanout, len(f.fanouts)), edges: f.edges}
	for i := range f.fanouts {
		copy.fanouts[i] = slices.Clone(f.fanouts[i])
	}
	return copy
}

func (f *AigFanoutIndex) validEdge(source AigID, edge AigFanout) bool {
	return f != nil && uint64(source) < uint64(len(f.fanouts)) && edge.Object > 0 && uint64(edge.Object) < uint64(len(f.fanouts)) && edge.Pin < 2
}

func aigObjectFanoutEdge(g *AigManager, source, fanout AigID) (AigFanout, error) {
	if g == nil || uint64(source) >= uint64(len(g.objects)) || uint64(fanout) >= uint64(len(g.objects)) || fanout == 0 || g.objects[source].Type == AigNone {
		return AigFanout{}, ErrInvalidLiteral
	}
	object := g.objects[fanout]
	switch object.Type {
	case AigCO, AigBuffer:
		if object.Fanin0.ID() == source {
			return AigFanout{Object: fanout, Pin: 0}, nil
		}
	case AigAND:
		if object.Fanin0.ID() == source {
			return AigFanout{Object: fanout, Pin: 0}, nil
		}
		if object.Fanin1.ID() == source {
			return AigFanout{Object: fanout, Pin: 1}, nil
		}
	}
	return AigFanout{}, ErrInvariant
}

// Add implements the ordered-list effect of Aig_ObjAddFanout. The graph edit
// that establishes the corresponding structural edge is the caller's separate
// transaction; duplicate edge identities are rejected instead of corrupting
// the native intrusive list.
func (f *AigFanoutIndex) Add(source AigID, edge AigFanout) error {
	if !f.validEdge(source, edge) {
		return ErrInvalidLiteral
	}
	for _, current := range f.fanouts[source] {
		if current == edge {
			return ErrInvariant
		}
	}
	if f.edges == int(^uint(0)>>1) {
		return ErrLimit
	}
	f.fanouts[source] = append(f.fanouts[source], edge)
	f.edges++
	return nil
}

// AddObject implements Aig_ObjAddFanout's graph-derived fanin selection.
func (f *AigFanoutIndex) AddObject(g *AigManager, source, fanout AigID) error {
	edge, err := aigObjectFanoutEdge(g, source, fanout)
	if err != nil {
		return err
	}
	return f.Add(source, edge)
}

// Remove implements Aig_ObjRemoveFanout and preserves the iteration order of
// all remaining edges. Missing edges are typed invariant failures.
func (f *AigFanoutIndex) Remove(source AigID, edge AigFanout) error {
	if !f.validEdge(source, edge) {
		return ErrInvalidLiteral
	}
	list := f.fanouts[source]
	for i, current := range list {
		if current != edge {
			continue
		}
		copy(list[i:], list[i+1:])
		f.fanouts[source] = list[:len(list)-1]
		f.edges--
		return nil
	}
	return ErrInvariant
}

// RemoveObject implements Aig_ObjRemoveFanout's graph-derived fanin selection.
func (f *AigFanoutIndex) RemoveObject(g *AigManager, source, fanout AigID) error {
	edge, err := aigObjectFanoutEdge(g, source, fanout)
	if err != nil {
		return err
	}
	return f.Remove(source, edge)
}
