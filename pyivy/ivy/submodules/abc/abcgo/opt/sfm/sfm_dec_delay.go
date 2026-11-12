package sfm

import "github.com/glycerine/abcgo/map/mio"

// sfmFindDelayNodeMatch searches decompositions and evaluates the library's
// delay-profile matches against the current timing state.
func sfmFindDelayNodeMatch(p *Sfm_Ntk_t, pivot int, pars Sfm_Par_t, lib *Sfm_Lib_t, timing *Sfm_Tim_t) (gates []int32, fanins [][]int32, order, cone []int, ok bool) {
	return sfmFindDelayNodeMatchForced(p, pivot, pars, lib, timing, -1)
}

func sfmFindDelayNodeMatchForced(p *Sfm_Ntk_t, pivot int, pars Sfm_Par_t, lib *Sfm_Lib_t, timing *Sfm_Tim_t, forced int) (gates []int32, fanins [][]int32, order, cone []int, ok bool) {
	session, window, ok := sfmPrepareMappedSession(p, pivot, pars, timing)
	if !ok {
		return nil, nil, nil, nil, false
	}
	order = window.order
	forceIndex := -1
	if forced >= 0 {
		for i, id := range order[:window.divisors] {
			if id == forced {
				forceIndex = i
				break
			}
		}
		if forceIndex < 0 {
			return nil, nil, nil, nil, false
		}
	}
	previous := 0
	var bestSupport []int32
	var best Sfm_LibMatch
	found := false
	for i := 0; i < max(1, pars.DecMax); i++ {
		if len(session.history) > previous {
			session.history = session.history[:previous]
		}
		previous = len(session.history) + 1
		truth, support, success := session.attempt(forceIndex)
		forceIndex = -1
		if !success {
			continue
		}
		cut := make([]int, len(support))
		for j, id := range support {
			cut[j] = order[id]
		}
		oldArea, candidateCone := sfmMappedMffcArea(p, pivot, cut)
		if len(support) == 1 && truth[0] == 0x5555555555555555 && len(lib.pCells) > 3 {
			inv := lib.pCells[3].Gate()
			if inv != nil && Sfm_TimReadObjDelay(timing, pivot) <= int(float32(mio.Mio_GateReadDelayMax(inv))*1000)+Sfm_TimReadObjDelay(timing, cut[0]) {
				continue
			}
		}
		if len(support) < 2 {
			gates, fanins = Sfm_LibImplementSimple(lib, truth, support)
			return gates, fanins, order, candidateCone, true
		}
		localSupport := make([]int, len(support))
		for j, id := range support {
			localSupport[j] = int(id)
		}
		bestDelay := Sfm_TimReadObjDelay(timing, pivot)
		for _, match := range Sfm_LibFindDelayMatches(lib, truth) {
			first := sfmGateFromMio(match.Bottom)
			var second *Gate
			if match.Top != nil {
				second = sfmGateFromMio(match.Top)
			}
			pins1 := make([]int, len(match.BottomPins))
			for j, pin := range match.BottomPins {
				pins1[j] = int(pin)
			}
			pins2 := make([]int, len(match.TopPins))
			for j, pin := range match.TopPins {
				pins2[j] = int(pin)
			}
			delay := Sfm_TimEvalRemapping(timing, localSupport, order, first, pins1, second, pins2)
			newArea := first.Area
			if second != nil {
				newArea += second.Area
			}
			if pars.DelAreaRatio != 0 && newArea > oldArea && delay/(newArea-oldArea) < pars.DelAreaRatio {
				continue
			}
			if delay < bestDelay {
				bestDelay, best, found, cone = delay, match, true, candidateCone
				bestSupport = append(bestSupport[:0], support...)
			}
		}
	}
	if !found {
		return nil, nil, nil, nil, false
	}
	gates, fanins = Sfm_LibImplementGatesDelay(bestSupport, best.Bottom, best.Top, best.BottomPins, best.TopPins)
	return gates, fanins, order, cone, true
}
