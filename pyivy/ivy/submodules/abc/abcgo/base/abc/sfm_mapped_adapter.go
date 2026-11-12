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
