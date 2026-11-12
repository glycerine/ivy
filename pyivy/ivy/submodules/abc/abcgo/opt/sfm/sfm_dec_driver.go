package sfm

import (
	"fmt"

	"github.com/glycerine/abcgo/map/mio"
)

// sfmOptimizeMappedArea runs the native forward area pass on the original
// internal objects. Unsupported driver options are reported explicitly until
// their mapped-network implementations are available.
func sfmOptimizeMappedArea(p *Sfm_Ntk_t, pars Sfm_Par_t, library *Sfm_Lib_t, handles []*mio.Mio_Gate_t) (int, error) {
	if p == nil || library == nil || pars.VarMax < 1 || pars.VarMax > SFM_SUPP_MAX || pars.MffcMax < 1 {
		return 0, fmt.Errorf("sfm: invalid mapped area parameters")
	}
	if pars.AreaRev || pars.MoreEffort || pars.DecMax > 1 {
		return 0, fmt.Errorf("sfm: requested mapped area option is not yet implemented")
	}
	if err := p.refresh(); err != nil {
		return 0, err
	}
	for id := p.Pis; id < len(p.Fanins)-p.Pos; id++ {
		if !p.Deleted[id] && p.Mapping[id] == nil {
			return 0, fmt.Errorf("sfm: node %d is not mapped", id)
		}
	}
	stop, changed := len(p.Fanins)-p.Pos, 0
	for id := p.Pis; id < stop; id++ {
		if pars.NodesMax > 0 && id > pars.NodesMax {
			break
		}
		if p.Deleted[id] || p.Fixed[id] || len(p.Fanouts[id]) == 0 || (pars.NodeOne != 0 && id != pars.NodeOne) {
			continue
		}
		if pars.MffcMin > 1 {
			_, cone := sfmMappedMffcArea(p, id, nil)
			if len(cone) < pars.MffcMin {
				continue
			}
		}
		gates, fanins, order, cone, ok := sfmFindAreaNodeMatch(p, id, pars, library)
		if ok && sfmInsertAreaMatch(p, id, gates, fanins, order, cone, handles) {
			changed++
		}
	}
	return changed, nil
}
