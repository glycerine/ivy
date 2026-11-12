package sfm

import (
	"fmt"

	"github.com/glycerine/abcgo/map/mio"
)

// sfmOptimizeMappedArea runs the native mapped-network area pass.
func sfmOptimizeMappedArea(p *Sfm_Ntk_t, pars Sfm_Par_t, library *Sfm_Lib_t, handles []*mio.Mio_Gate_t) (int, error) {
	if p == nil || library == nil || pars.VarMax < 1 || pars.VarMax > SFM_SUPP_MAX || pars.MffcMax < 1 || pars.DecMax < 0 || pars.DecMax > SFM_DEC_MAX {
		return 0, fmt.Errorf("sfm: invalid mapped area parameters")
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
	try := func(id int) int {
		if p.Deleted[id] || p.Fixed[id] || len(p.Fanouts[id]) == 0 || (pars.NodeOne != 0 && id != pars.NodeOne) {
			return -1
		}
		if pars.MffcMin > 1 {
			_, cone := sfmMappedMffcArea(p, id, nil)
			if len(cone) < pars.MffcMin {
				return -1
			}
		}
		gates, fanins, order, cone, ok := sfmFindAreaNodeMatch(p, id, pars, library)
		if !ok && pars.MoreEffort {
			window := sfmExtractAreaWindow(p, id, pars)
			for i := len(window.mffcInputs) - 1; i >= 0; i-- {
				gates, fanins, order, cone, ok = sfmFindAreaNodeMatchForced(p, id, pars, library, window.mffcInputs[i])
				if ok {
					break
				}
			}
		}
		if ok && sfmInsertAreaMatch(p, id, gates, fanins, order, cone, handles) {
			changed++
			return len(p.Fanins) - p.Pos - 1
		}
		return -1
	}
	if !pars.AreaRev {
		for id := p.Pis; id < stop; id++ {
			if pars.NodesMax > 0 && id > pars.NodesMax {
				break
			}
			try(id)
		}
		return changed, nil
	}
	seen := make(map[int]bool)
	var queue []int
	add := func(id int) {
		if Sfm_ObjIsNode(p, id) && !p.Deleted[id] && !seen[id] {
			seen[id] = true
			queue = append(queue, id)
		}
	}
	for id := stop; id < len(p.Fanins); id++ {
		add(p.Fanins[id][0])
	}
	for i := 0; i < len(queue); i++ {
		id := queue[i]
		if p.Deleted[id] {
			continue
		}
		if replacement := try(id); replacement >= 0 {
			id = replacement
		}
		for _, fanin := range p.Fanins[id] {
			add(fanin)
		}
	}
	return changed, nil
}
