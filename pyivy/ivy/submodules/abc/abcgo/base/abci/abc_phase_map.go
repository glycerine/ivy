package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/opt/sfm"
)

func abcPhaseMap(n *Abc_Ntk_t, verbose int32) error {
	library, ok := n.FunctionManager.(*mio.Mio_Library_t)
	if !ok || library == nil {
		return fmt.Errorf("phase_map: missing cell library")
	}
	cells := mio.Mio_CollectRootsNew2(library, 6, nil, 0)
	if len(cells) < 4 {
		return fmt.Errorf("phase_map: library lacks constants, buffer or inverter")
	}
	gates := make([]*sfm.Gate, len(cells))
	byCell := make(map[*mio.Mio_Gate_t]*sfm.Gate, len(cells))
	for i := range cells {
		gates[i] = sfm.Sfm_GateFromMio(cells[i].Gate())
		byCell[cells[i].Gate()] = gates[i]
	}
	network, objects, err := abc.Abc_NtkExtractSfm(n)
	if err != nil {
		return err
	}
	defer sfm.Sfm_NtkFree(network)
	// SFM identifies cells by pointer, so the extracted nodes and the
	// polarity table must share gate handles. Preserve non-root cells too.
	for id, gate := range network.Mapping {
		if gate == nil {
			continue
		}
		canonical := byCell[gate.Cell]
		if canonical == nil {
			canonical = gate
			byCell[gate.Cell] = gate
			gates = append(gates, gate)
		}
		network.Mapping[id] = canonical
	}
	if sfm.Abc_NtkChangePerform(network, gates, verbose != 0) == 0 {
		return nil
	}
	return abc.Abc_NtkApplySfm(n, network, objects)
}
