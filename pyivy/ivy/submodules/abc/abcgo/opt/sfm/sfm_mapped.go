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
		return sfmOptimizeMappedArea(p, pars, prepared, handles)
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
