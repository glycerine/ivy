package cnf

import "github.com/glycerine/abcgo/aig/aig"

func Cnf_ManPostprocess_old(*Cnf_Man_t) {}

// Cnf_ManTransferCuts restores the selected DAR cuts on referenced
// nodes. Explicitly supplied CNF cuts have no DAR representation.
func Cnf_ManTransferCuts(manager *Cnf_Man_t) {
	if manager == nil || manager.pManAig == nil {
		return
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	for id := range manager.darBest {
		edge := aig.Edge(id) << 1
		if !manager.pManAig.ValidEdge(edge) || aig.Aig_ObjRefs(manager.pManAig, edge) == 0 {
			manager.selectedCuts[id] = nil
		} else if manager.darBest[id].fBest {
			manager.selectedCuts[id] = Cnf_CutCreate(manager, &edge)
		}
	}
}
func Cnf_ManFreeCuts(manager *Cnf_Man_t) {
	if manager == nil {
		return
	}
	for _, cut := range manager.selectedCuts {
		Cnf_CutFree(cut)
	}
	manager.selectedCuts = nil
	manager.darBest = nil
	if manager.pCuts != nil {
		aig.Aig_ManCutStop(manager.pCuts)
		manager.pCuts = nil
	}
}

// Cnf_ManPostprocess folds at most one singly referenced leaf per mapped
// gate. As in cnfPost.c, supplied cuts use the current mapped reference counts.
func Cnf_ManPostprocess(manager *Cnf_Man_t) {
	if manager == nil || manager.pManAig == nil {
		return
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
		Cnf_ManScanMapping(manager, 0, 0)
	}
	g := manager.pManAig
	for id, current := range manager.selectedCuts {
		edge := aig.Edge(id) << 1
		if current == nil || !g.ValidEdge(edge) || !aig.Aig_ObjIsNode(g, edge) || aig.Aig_ObjRefs(g, edge) == 0 {
			continue
		}
		// Stable insertion by increasing fanin cost retains leaf order on ties.
		var order [16]int32
		count := 0
		for _, leaf := range current.pFanins {
			e := aig.Edge(leaf) << 1
			if leaf < 0 || int(leaf) >= len(manager.selectedCuts) || !g.ValidEdge(e) || !aig.Aig_ObjIsNode(g, e) || aig.Aig_ObjRefs(g, e) != 1 {
				continue
			}
			fan := manager.selectedCuts[leaf]
			if fan == nil {
				continue
			}
			i := count
			for i > 0 && manager.selectedCuts[order[i-1]].Cost > fan.Cost {
				order[i] = order[i-1]
				i--
			}
			order[i] = leaf
			count++
		}
		for _, leaf := range order[:count] {
			fan := manager.selectedCuts[leaf]
			merged := Cnf_CutCompose(manager, current, fan, leaf)
			if merged == nil {
				continue
			}
			if merged.Cost == 127 || int(merged.Cost) > int(current.Cost)+int(fan.Cost) {
				Cnf_CutFree(merged)
				continue
			}
			Cnf_CutUpdateRefs(manager, current, fan, merged)
			manager.selectedCuts[id] = merged
			manager.selectedCuts[leaf] = nil
			Cnf_CutFree(current)
			Cnf_CutFree(fan)
			break
		}
	}
}
