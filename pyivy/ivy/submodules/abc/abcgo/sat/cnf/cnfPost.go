package cnf

import "github.com/glycerine/abcgo/aig/aig"

func Cnf_ManPostprocess_old(*Cnf_Man_t) {}

// Cnf_ManTransferCuts is retained as a compatibility name. Cut transfer is now
// an owned-slice copy performed while Cnf_DeriveMapping selects each cut.
func Cnf_ManTransferCuts(manager *Cnf_Man_t) {
	if manager == nil {
		return
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
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
	if manager.pCuts != nil {
		aig.Aig_ManCutStop(manager.pCuts)
		manager.pCuts = nil
	}
}

// Cnf_ManPostprocess tries one-level cut substitutions and keeps a result only
// when its estimated clause cost decreases.
func Cnf_ManPostprocess(manager *Cnf_Man_t) {
	if manager == nil || manager.pManAig == nil {
		return
	}
	if len(manager.selectedCuts) == 0 {
		Cnf_DeriveMapping(manager)
	}
	for id, current := range manager.selectedCuts {
		if current == nil {
			continue
		}
		for _, leaf := range append([]int32(nil), current.pFanins...) {
			if int(leaf) >= len(manager.selectedCuts) {
				continue
			}
			candidate := manager.selectedCuts[leaf]
			if candidate == nil {
				continue
			}
			merged := Cnf_CutCompose(manager, current, candidate, leaf)
			if merged == nil {
				continue
			}
			if merged.Cost < current.Cost {
				Cnf_CutFree(current)
				manager.selectedCuts[id] = merged
				current = merged
			} else {
				Cnf_CutFree(merged)
			}
		}
	}
}
