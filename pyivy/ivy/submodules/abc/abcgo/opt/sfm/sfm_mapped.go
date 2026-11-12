package sfm

import (
	"fmt"

	"github.com/glycerine/abcgo/map/mio"
)

// Sfm_GateFromMio copies a library gate's Boolean function and timing data.
func Sfm_GateFromMio(gate *mio.Mio_Gate_t) *Gate {
	if gate == nil {
		return nil
	}
	return sfmGateFromMio(gate)
}

// Sfm_OptimizeMapped runs the native area or timing pass on a mapped network.
// The caller owns the network and supplies its gate library explicitly.
func Sfm_OptimizeMapped(p *Sfm_Ntk_t, pars Sfm_Par_t, library *mio.Mio_Library_t) (int, error) {
	if p == nil || library == nil || pars.VarMax < 1 || pars.VarMax > SFM_SUPP_MAX {
		return 0, fmt.Errorf("sfm: invalid mapped network or library")
	}
	_, _, _, handles := Sfm_LibPreprocess(library)
	cells := mio.Mio_CollectRootsNew2(library, int32(min(6, pars.VarMax)), nil, 0)
	if len(cells) < 4 {
		return 0, fmt.Errorf("sfm: library lacks basic gates")
	}
	delay := int32(0)
	if !pars.Area {
		delay = 1
	}
	prepared := sfmLibPrepareCells(int32(pars.VarMax), 1, delay, 0, 0, cells)
	defer Sfm_LibStop(prepared)
	if pars.Area {
		changed, err := sfmOptimizeMappedArea(p, pars, prepared, handles)
		if err != nil {
			return changed, err
		}
		// The source follows area decomposition with its polarity optimization.
		// Preserve gate pointer identity when connecting the network to its library.
		polarityCells := make([]*Gate, 0, len(handles)+4)
		byGate := make(map[*mio.Mio_Gate_t]*Gate, len(handles))
		add := func(g *mio.Mio_Gate_t) {
			if g == nil || byGate[g] != nil {
				return
			}
			cell := sfmGateFromMio(g)
			byGate[g] = cell
			polarityCells = append(polarityCells, cell)
		}
		for i := 0; i < 4; i++ {
			add(cells[i].Gate())
		}
		for _, g := range handles {
			if mio.Mio_GateReadPinNum(g) > 0 && mio.Mio_GateReadPinNum(g) <= 6 {
				add(g)
			}
		}
		for id := p.Pis; id < len(p.Fanins)-p.Pos; id++ {
			if p.Deleted[id] {
				continue
			}
			if p.Mapping[id] == nil || byGate[p.Mapping[id].Cell] == nil {
				return changed, fmt.Errorf("sfm: mapped gate %d is outside the library", id)
			}
			p.Mapping[id] = byGate[p.Mapping[id].Cell]
		}
		if Abc_NtkChangePerform(p, polarityCells, pars.Verbose) > 0 {
			changed++
		}
		return changed, nil
	}
	delta := pars.DeltaCrit * 1000
	if delta == 0 {
		delta = 5 * int(float32(mio.Mio_LibraryReadDelayInvMax(library))*1000) / 2
	}
	if delta <= 0 {
		return 0, fmt.Errorf("sfm: critical timing delta must be positive")
	}
	return sfmOptimizeMappedDelay(p, pars, prepared, handles, Sfm_TimStart(p, delta))
}
