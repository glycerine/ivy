package scl

import (
	"fmt"
	"sort"
	"time"
)

// Abc_SclFindWindow collects the pivot's fanins, two levels of fanout,
// and the boundary nodes whose changed arrival can affect the remaining graph.
func Abc_SclFindWindow(net TimingNetwork, pivot int) (nodes, evals []int, err error) {
	if !net.SclIsNode(pivot) {
		return nil, nil, fmt.Errorf("downsizing pivot %d is not a node", pivot)
	}
	inputs := map[int]bool{}
	for _, id := range net.SclInputs() {
		inputs[id] = true
	}
	set := map[int]bool{pivot: true}
	for _, id := range net.SclFanins(pivot) {
		if inputs[id] || len(net.SclFanins(id)) > 0 {
			set[id] = true
		}
	}
	for _, id := range net.SclFanouts(pivot) {
		if !net.SclIsNode(id) {
			continue
		}
		set[id] = true
		for _, next := range net.SclFanouts(id) {
			if net.SclIsNode(next) {
				set[next] = true
			}
		}
	}
	for id := range set {
		nodes = append(nodes, id)
	}
	sort.Ints(nodes)
	for _, id := range nodes {
		for _, next := range net.SclFanouts(id) {
			if !set[next] {
				evals = append(evals, id)
				break
			}
		}
	}
	if len(evals) == 0 {
		return nil, nil, fmt.Errorf("downsizing window at %d has no observable boundary", pivot)
	}
	return nodes, evals, nil
}
func Abc_SclCheckOverlap(marked map[int]bool, nodes []int) bool {
	for _, id := range nodes {
		if marked[id] {
			return true
		}
	}
	return false
}

func Abc_SclCheckImprovement(p *SC_Man, id int, nodes, evals []int, notches, gap int) (bool, error) {
	start := time.Now()
	defer func() { p.TimeSize += time.Since(start) }()
	old := Abc_SclObjCell(p, id)
	if old == nil {
		return false, fmt.Errorf("node %d has no assigned cell", id)
	}
	Abc_SclConeStore(p, nodes)
	Abc_SclEvalStore(p, evals)
	Abc_SclLoadStore(p, id)
	restore := func() { Abc_SclObjSetCell(p, id, old); Abc_SclLoadRestore(p, id); Abc_SclConeRestore(p, nodes) }
	best := -1
	bestGain := -float32(gap)
	for candidate, index := old, 0; ; index++ {
		if candidate.Area < old.Area {
			if index > notches {
				break
			}
			Abc_SclObjSetCell(p, id, candidate)
			if e := Abc_SclUpdateLoad(p, id, old, candidate); e != nil {
				restore()
				return false, e
			}
			if e := Abc_SclTimeCone(p, nodes); e != nil {
				restore()
				return false, e
			}
			Abc_SclObjSetCell(p, id, old)
			Abc_SclLoadRestore(p, id)
			gain := Abc_SclEvalPerformLegal(p, evals, p.InitialDelay)
			if gain != -1 && gain > bestGain {
				bestGain = gain
				best = candidate.ID
			}
		}
		candidate = candidate.Prev
		if candidate == nil || candidate == old {
			break
		}
	}
	restore()
	if best < 0 {
		return false, nil
	}
	next := p.Lib.Cells[best]
	Abc_SclObjSetCell(p, id, next)
	if e := Abc_SclUpdateLoad(p, id, old, next); e != nil {
		restore()
		return false, e
	}
	p.SumArea += next.Area - old.Area
	Abc_SclTimeIncInsert(p, id)
	return true, nil
}
func Abc_NtkCollectNodesByArea(p *SC_Man, net TimingNetwork) []int {
	var ids []int
	for _, id := range net.SclObjectIDs() {
		if net.SclIsNode(id) && !net.SclIsBarrier(id) && len(net.SclFanins(id)) > 0 {
			ids = append(ids, id)
		}
	}
	sort.SliceStable(ids, func(i, j int) bool {
		a, b := Abc_SclObjCell(p, ids[i]), Abc_SclObjCell(p, ids[j])
		if a.Area == b.Area {
			return ids[i] < ids[j]
		}
		return a.Area > b.Area
	})
	return ids
}
func Abc_SclDnsizePrint(p *SC_Man, iteration, attempts, overlaps, changes int, verbose bool) {
	label := fmt.Sprintf("%5d", iteration)
	if iteration < 0 {
		label = "Total"
	}
	area, delay := float32(0), float32(0)
	if p.InitialArea != 0 {
		area = 100 * (p.SumArea - p.InitialArea) / p.InitialArea
	}
	if p.InitialDelay != 0 {
		delay = 100 * (p.MaxDelay - p.InitialDelay) / p.InitialDelay
	}
	end := "\r"
	if verbose {
		end = "\n"
	}
	fmt.Fprintf(p.Output, "%s : Try =%6d  Over =%6d  Fail =%6d  Win =%6d  A: %.2f (%+5.1f %%)  D: %.2f ps (%+5.1f %%)%s", label, attempts, overlaps, attempts-overlaps-changes, changes, p.SumArea, area, p.MaxDelay, delay, end)
}
func Abc_SclDnsizePerformInt(lib *SC_Lib, net TimingNetwork, pars *SC_SizePars) error {
	if pars == nil {
		return fmt.Errorf("downsizing parameters are nil")
	}
	p, e := Abc_SclManStart(lib, net, boolInt(pars.UseWireLoads), boolInt(pars.UseDepartures), float32(pars.DelayUser), int32(pars.BufferTreeEstimate))
	if e != nil {
		return e
	}
	defer Abc_SclManFree(p)
	return DownsizeManager(p, pars)
}
func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

// DownsizeManager preserves the original area-priority rounds and postpones
// overlapping windows until the next incremental timing update.
func DownsizeManager(p *SC_Man, pars *SC_SizePars) error {
	start := time.Now()
	var deadline time.Time
	if pars.Timeout > 0 {
		deadline = start.Add(time.Duration(pars.Timeout) * time.Second)
	}
	expired := func() bool { return !deadline.IsZero() && time.Now().After(deadline) }
	p.BestGates = append([]int32(nil), p.Mapping.GateIDs...)
	if pars.Verbose {
		fmt.Fprintf(p.Output, "Parameters: Iters =%5d. UseDept =%d. UseWL =%d. Target =%5d ps. DelayGap =%3d ps. Timeout =%4d sec\n", pars.Iterations, boolInt(pars.UseDepartures), boolInt(pars.UseWireLoads), pars.DelayUser, pars.DelayGap, pars.Timeout)
	}
	for iteration := 0; iteration < pars.Iterations; iteration++ {
		pending := Abc_NtkCollectNodesByArea(p, p.Network)
		allAttempts, allOverlap, allChanges := 0, 0, 0
		for round := 0; len(pending) > 0; round++ {
			marked := map[int]bool{}
			var later []int
			attempts, overlaps, changes := 0, 0, 0
			for _, id := range pending {
				coneStart := time.Now()
				nodes, evals, e := Abc_SclFindWindow(p.Network, id)
				p.TimeCone += time.Since(coneStart)
				if e != nil {
					return e
				}
				attempts++
				if Abc_SclCheckOverlap(marked, nodes) {
					overlaps++
					later = append(later, id)
					continue
				}
				changed, e := Abc_SclCheckImprovement(p, id, nodes, evals, pars.Notches, pars.DelayGap)
				if e != nil {
					return e
				}
				if changed {
					changes++
					for _, n := range nodes {
						marked[n] = true
					}
				}
			}
			// Candidate priorities are unchanged across rounds except successfully resized
			// cells, which are not deferred. The retained order is therefore the same heap
			// priority order as the original queue reinsertion.
			pending = later
			timingStart := time.Now()
			if len(p.Changed) > 0 {
				if _, e := Abc_SclTimeIncUpdate(p); e != nil {
					return e
				}
			} else {
				if e := Abc_SclTimeNtkRecompute(p, &p.SumArea, &p.MaxDelay, boolInt(pars.UseDepartures), float32(pars.DelayUser)); e != nil {
					return e
				}
			}
			p.TimeTiming += time.Since(timingStart)
			p.MaxDelay = Abc_SclReadMaxDelay(p)
			if pars.UseDepartures && pars.DelayUser > 0 {
				p.MaxDelay = max(p.MaxDelay, float32(pars.DelayUser))
			}
			Abc_SclDnsizePrint(p, round, attempts, overlaps, changes, pars.VeryVerbose)
			allAttempts += attempts
			allOverlap += overlaps
			allChanges += changes
			if expired() {
				break
			}
		}
		if pars.Verbose {
			Abc_SclDnsizePrint(p, -1, allAttempts, allOverlap, allChanges, true)
		}
		if expired() || allAttempts == 0 {
			break
		}
	}
	p.TimeTotal = time.Since(start)
	p.TimeOther = p.TimeTotal - p.TimeCone - p.TimeSize - p.TimeTiming
	if pars.DumpStats {
		if e := Abc_SclDumpStats(p, "stats2.txt", p.TimeTotal); e != nil {
			return e
		}
	}
	if expired() {
		fmt.Fprintf(p.Output, "Gate sizing timed out at %d seconds.\n", pars.Timeout)
	}
	return Abc_SclSclGates2MioGates(p.Lib, p.Network)
}
func Abc_SclDnsizePerform(lib *SC_Lib, net TimingNetwork, pars *SC_SizePars) error {
	return Abc_SclDnsizePerformInt(lib, net, pars)
}
