package abc

import (
	"fmt"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/opt/sfm"
)

// Abc_NtkExtractSfm copies a mapped logic network into SFM's compact indexing.
// The returned object slice maps every SFM ID to its original ABC object.
func Abc_NtkExtractSfm(n *Network) (*sfm.Sfm_Ntk_t, []*Object, error) {
	if n == nil || Abc_NtkIsMappedLogic(n) == 0 {
		return nil, nil, fmt.Errorf("sfm: expected mapped logic network")
	}
	if err := n.Check(); err != nil {
		return nil, nil, err
	}
	roots := append(n.ObjectsOfKind(ObjectNode), n.Outputs()...)
	topo, err := n.Topological(roots)
	if err != nil {
		return nil, nil, err
	}
	objects := append([]*Object(nil), n.Inputs()...)
	for _, obj := range topo {
		if obj.Kind == ObjectNode {
			objects = append(objects, obj)
		}
	}
	objects = append(objects, n.Outputs()...)
	ids := make(map[*Object]int, len(objects))
	for i, obj := range objects {
		if _, ok := ids[obj]; ok {
			return nil, nil, fmt.Errorf("sfm: duplicate object %d", obj.ID)
		}
		ids[obj] = i
	}
	fanins := make([][]int, len(objects))
	truths := make([][]uint64, len(objects))
	gates := make([]*sfm.Gate, len(objects))
	for i, obj := range objects {
		truths[i] = []uint64{0}
		if obj.IsCI() {
			continue
		}
		for _, edge := range obj.fanins {
			if edge.Complement {
				return nil, nil, fmt.Errorf("sfm: complemented edge at object %d", obj.ID)
			}
			id, ok := ids[edge.Object]
			if !ok {
				return nil, nil, fmt.Errorf("sfm: unsupported fanin %d at object %d", edge.Object.ID, obj.ID)
			}
			fanins[i] = append(fanins[i], id)
		}
		if obj.Kind != ObjectNode {
			continue
		}
		gate, ok := obj.Data.(*mio.Mio_Gate_t)
		if !ok || gate == nil || int(mio.Mio_GateReadPinNum(gate)) != len(obj.fanins) {
			return nil, nil, fmt.Errorf("sfm: object %d has invalid mapped gate", obj.ID)
		}
		if len(obj.fanins) > 6 {
			return nil, nil, fmt.Errorf("sfm: object %d has more than six gate inputs", obj.ID)
		}
		gates[i] = sfm.Sfm_GateFromMio(gate)
		truths[i] = append([]uint64(nil), gates[i].Truth...)
	}
	result, err := sfm.NewNetwork(fanins, len(n.Inputs()), len(n.Outputs()), truths, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	copy(result.Mapping, gates)
	return result, objects, nil
}

// Abc_NtkApplySfm applies an optimized copy back to its original network.
// Existing live nodes and terminals retain their identities; new mapped nodes
// are added and removed nodes are deleted after all replacement edges exist.
func Abc_NtkApplySfm(n *Network, p *sfm.Sfm_Ntk_t, original []*Object) error {
	if n == nil || p == nil || Abc_NtkIsMappedLogic(n) == 0 || p.Pis != len(n.Inputs()) || p.Pos != len(n.Outputs()) || len(original) < p.Pis+p.Pos || len(p.Mapping) != len(p.Fanins) {
		return fmt.Errorf("sfm: incompatible mapped network")
	}
	if err := n.Check(); err != nil {
		return err
	}
	if err := p.Check(); err != nil {
		return err
	}
	oldInternalEnd := len(original) - p.Pos
	newInternalEnd := len(p.Fanins) - p.Pos
	if newInternalEnd < oldInternalEnd {
		return fmt.Errorf("sfm: original IDs are missing")
	}
	for i, obj := range original {
		if !n.contains(obj) {
			return fmt.Errorf("sfm: original object %d is missing", i)
		}
		if i < p.Pis && n.Inputs()[i] != obj || i >= oldInternalEnd && n.Outputs()[i-oldInternalEnd] != obj || i >= p.Pis && i < oldInternalEnd && obj.Kind != ObjectNode {
			return fmt.Errorf("sfm: original object order changed")
		}
	}
	for id := p.Pis; id < newInternalEnd; id++ {
		if p.Deleted[id] {
			continue
		}
		gate := p.Mapping[id]
		if gate == nil || gate.Cell == nil || gate.Inputs != len(p.Fanins[id]) || int(mio.Mio_GateReadPinNum(gate.Cell)) != len(p.Fanins[id]) || len(gate.Truth) != 1 || gate.Truth[0] != mio.Mio_GateReadTruth(gate.Cell) {
			return fmt.Errorf("sfm: invalid mapped gate at %d", id)
		}
	}
	objects := make([]*Object, len(p.Fanins))
	copy(objects[:oldInternalEnd], original[:oldInternalEnd])
	copy(objects[newInternalEnd:], original[oldInternalEnd:])
	for id := oldInternalEnd; id < newInternalEnd; id++ {
		if !p.Deleted[id] {
			objects[id] = n.NewObject(ObjectNode)
		}
	}
	for id := p.Pis; id < len(objects); id++ {
		obj := objects[id]
		if obj == nil {
			continue
		}
		for len(obj.fanins) > 0 {
			if err := n.DeleteFanin(obj, len(obj.fanins)-1); err != nil {
				panic(err)
			}
		}
	}
	for id := p.Pis; id < len(objects); id++ {
		if p.Deleted[id] {
			continue
		}
		obj := objects[id]
		if id < newInternalEnd {
			obj.Data = p.Mapping[id].Cell
		}
		for _, f := range p.Fanins[id] {
			if err := n.AddFanin(obj, objects[f].Signal()); err != nil {
				panic(err)
			}
		}
	}
	for id := p.Pis; id < oldInternalEnd; id++ {
		if p.Deleted[id] {
			if err := n.DeleteObject(objects[id]); err != nil {
				panic(err)
			}
		}
	}
	n.timingMapping = nil
	Abc_NtkLevel(n)
	return n.Check()
}
