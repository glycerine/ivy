package sfm

import (
	"fmt"

	"github.com/glycerine/abcgo/map/mio"
)

// sfmOptimizeMappedDelay repeatedly processes timing-prioritized mapped nodes.
// Failed nodes are marked until the pass ends, matching ABC's traversal.
func sfmOptimizeMappedDelay(p *Sfm_Ntk_t, pars Sfm_Par_t, lib *Sfm_Lib_t, handles []*mio.Mio_Gate_t, timing *Sfm_Tim_t) (int, error) {
	if p == nil || lib == nil || timing == nil || timing.Network != p || pars.VarMax < 1 || pars.VarMax > SFM_SUPP_MAX || pars.MffcMax < 1 || pars.DecMax < 0 || pars.DecMax > SFM_DEC_MAX || pars.TimeWin < 0 || pars.TimeWin > 100 {
		return 0, fmt.Errorf("sfm: invalid mapped delay parameters")
	}
	if err := p.refresh(); err != nil {
		return 0, err
	}
	for id := p.Pis; id < len(p.Fanins)-p.Pos; id++ {
		if !p.Deleted[id] && (p.Mapping[id] == nil || len(p.Mapping[id].Pins) != len(p.Fanins[id])) {
			return 0, fmt.Errorf("sfm: node %d lacks timing data", id)
		}
	}
	Sfm_TimTrace(timing)
	failed := make(map[int]bool)
	changed := 0
	for pars.NodesMax == 0 || changed < pars.NodesMax {
		var candidates []int
		if pars.NodeOne != 0 {
			candidates = []int{pars.NodeOne}
		} else {
			candidates = Sfm_TimPriorityNodes(timing, pars.TimeWin)
		}
		improved := false
		for _, id := range candidates {
			if id < p.Pis || id >= len(p.Fanins)-p.Pos || p.Deleted[id] || p.Fixed[id] || failed[id] {
				continue
			}
			gates, fanins, order, cone, ok := sfmFindDelayNodeMatch(p, id, pars, lib, timing)
			if !ok && pars.MoreEffort {
				window := sfmExtractWindow(p, id, pars, timing)
				for i := len(window.mffcInputs) - 1; i >= 0; i-- {
					gates, fanins, order, cone, ok = sfmFindDelayNodeMatchForced(p, id, pars, lib, timing, window.mffcInputs[i])
					if ok {
						break
					}
				}
			}
			if !ok || !sfmInsertAreaMatch(p, id, gates, fanins, order, cone, handles) {
				failed[id] = true
				continue
			}
			changed++
			improved = true
			Sfm_TimTrace(timing)
			break
		}
		if pars.NodeOne != 0 || !improved {
			break
		}
	}
	return changed, nil
}
