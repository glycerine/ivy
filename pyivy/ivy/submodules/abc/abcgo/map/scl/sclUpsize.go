package scl

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// MutableTimingNetwork identifies fanins by pin index, retaining duplicate-edge
// multiplicity when tentative bypass transformations are undone.
type MutableTimingNetwork interface {
	TimingNetwork
	SclPatchFanin(node, pin, newID int) error
	SclDeleteNode(id int) error
}
type PhasedTimingNetwork interface {
	SclHasPhases() bool
	SclToggleFaninPhase(node, pin int) error
}
type BypassEdit struct {
	Fanout, Pin, Driver, Buffer int
	Inverted                    bool
}
type UpsizeState struct {
	Edits    []BypassEdit
	Critical map[int]bool
}
type upsizeCandidate struct {
	id, cell, fanout, driver, pin int
	gain                          float32
	invert                        bool
}

func Abc_SclFindTFO_rec(net TimingNetwork, id int, nodes, outputs *[]int, seen map[int]bool) {
	if seen[id] {
		return
	}
	seen[id] = true
	for _, out := range net.SclOutputs() {
		if out == id {
			*outputs = append(*outputs, id)
			return
		}
	}
	for _, fan := range net.SclFanouts(id) {
		Abc_SclFindTFO_rec(net, fan, nodes, outputs, seen)
	}
	if len(net.SclFanins(id)) > 0 {
		*nodes = append(*nodes, id)
	}
}
func Abc_SclFindTFO(net TimingNetwork, path []int) []int {
	var nodes, outputs []int
	seen := map[int]bool{}
	for _, id := range path {
		// A path node driven only by a CI must still start its own TFO; otherwise
		// the no-departure driver never observes its improved output arrival.
		Abc_SclFindTFO_rec(net, id, &nodes, &outputs, seen)
		for _, fanin := range net.SclFanins(id) {
			if net.SclIsNode(fanin) {
				Abc_SclFindTFO_rec(net, fanin, &nodes, &outputs, seen)
			}
		}
	}
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return append(nodes, outputs...)
}
func Abc_SclFindCriticalCoWindow(p *SC_Man, window int) []int {
	threshold := Abc_SclReadMaxDelay(p) * (100 - float32(window)) / 100
	var ids []int
	for _, id := range p.Network.SclOutputs() {
		if Abc_SclObjTimeMax(p, id) >= threshold {
			ids = append(ids, id)
		}
	}
	return ids
}
func Abc_SclFindCriticalNodeWindow_rec(p *SC_Man, id int, path *[]int, slack float32, departure bool, seen map[int]bool) {
	if seen[id] || !p.Network.SclIsNode(id) {
		return
	}
	seen[id] = true
	arrival := Abc_SclGetMaxDelayNodeFanins(p, id)
	if departure {
		arrival = Abc_SclObjGetSlack(p, id, p.MaxDelay)
	}
	arrival = max(arrival, 0)
	for _, fanin := range p.Network.SclFanins(id) {
		if !p.Network.SclIsNode(fanin) || len(p.Network.SclFanins(fanin)) == 0 {
			continue
		}
		remaining := slack - (arrival - Abc_SclObjTimeMax(p, fanin))
		if departure {
			remaining = slack - (Abc_SclObjGetSlack(p, fanin, p.MaxDelay) - arrival)
		}
		if remaining >= 0 {
			Abc_SclFindCriticalNodeWindow_rec(p, fanin, path, remaining, departure, seen)
		}
	}
	if len(p.Network.SclFanins(id)) > 0 && !p.Network.SclIsBarrier(id) {
		*path = append(*path, id)
	}
}
func Abc_SclFindCriticalNodeWindow(p *SC_Man, outputs []int, window int, departure bool) ([]int, map[int]bool) {
	delay := Abc_SclReadMaxDelay(p)
	slack := delay * float32(window) / 100
	seen, critical := map[int]bool{}, map[int]bool{}
	var path []int
	for _, id := range outputs {
		remaining := slack - (delay - Abc_SclObjTimeMax(p, id))
		if remaining >= 0 {
			fanin := p.Network.SclFanins(id)
			if len(fanin) > 0 {
				Abc_SclFindCriticalNodeWindow_rec(p, fanin[0], &path, remaining, departure, seen)
			}
		}
		critical[id] = true
	}
	for id := range seen {
		critical[id] = true
	}
	return path, critical
}
func Abc_SclUnmarkCriticalNodeWindow(critical map[int]bool, path []int) {
	for _, id := range path {
		delete(critical, id)
	}
}
func Abc_SclCountNearCriticalNodes(p *SC_Man) int {
	path, _ := Abc_SclFindCriticalNodeWindow(p, Abc_SclFindCriticalCoWindow(p, 5), 5, false)
	return len(path)
}
func Abc_SclFindNodesToUpdate(net TimingNetwork, pivot int, critical map[int]bool, extra int) (nodes, evals []int, err error) {
	if !net.SclIsNode(pivot) || !critical[pivot] {
		return nil, nil, fmt.Errorf("upsizing pivot %d is not critical", pivot)
	}
	inputs := map[int]bool{}
	for _, id := range net.SclInputs() {
		inputs[id] = true
	}
	set := map[int]bool{pivot: true}
	if extra >= 0 {
		set[extra] = true
	}
	for _, id := range net.SclFanins(pivot) {
		if inputs[id] || len(net.SclFanins(id)) > 0 {
			set[id] = true
		}
	}
	for _, id := range net.SclFanouts(pivot) {
		if !net.SclIsNode(id) || !critical[id] {
			continue
		}
		set[id] = true
		for _, next := range net.SclFanouts(id) {
			if net.SclIsNode(next) && critical[next] {
				set[next] = true
			}
		}
	}
	for id := range set {
		nodes = append(nodes, id)
	}
	sort.Ints(nodes)
	for _, id := range nodes {
		for _, fan := range net.SclFanouts(id) {
			if critical[fan] && !set[fan] {
				evals = append(evals, id)
				break
			}
		}
	}
	if len(evals) == 0 {
		return nil, nil, fmt.Errorf("upsizing window has no critical boundary at %d", pivot)
	}
	return nodes, evals, nil
}
func Abc_SclFindBestCell(p *SC_Man, id int, recalculate, evals []int, notches, gap int, gainBest *float32) (int, error) {
	old := Abc_SclObjCell(p, id)
	Abc_SclConeStore(p, recalculate)
	Abc_SclEvalStore(p, evals)
	Abc_SclLoadStore(p, id)
	restore := func() { Abc_SclObjSetCell(p, id, old); Abc_SclLoadRestore(p, id); Abc_SclConeRestore(p, recalculate) }
	best := -1
	gain := -float32(gap)
	noChange := 0
	for c, k := old, 0; ; k++ {
		if c != old {
			if k > notches {
				break
			}
			Abc_SclObjSetCell(p, id, c)
			if e := Abc_SclUpdateLoad(p, id, old, c); e != nil {
				restore()
				return -1, e
			}
			if e := Abc_SclTimeCone(p, recalculate); e != nil {
				restore()
				return -1, e
			}
			Abc_SclObjSetCell(p, id, old)
			Abc_SclLoadRestore(p, id)
			g := Abc_SclEvalPerform(p, evals)
			if g > gain {
				gain = g
				best = c.ID
				noChange = 1
			} else if noChange > 0 {
				noChange++
			}
			if noChange == 4 {
				break
			}
		}
		c = c.Next
		if c == nil || c == old {
			break
		}
	}
	restore()
	*gainBest = gain
	return best, nil
}
func Abc_SclObjCheckMarkedFanFans(net TimingNetwork, id int, marked map[int]bool) bool {
	if marked[id] {
		return true
	}
	for _, n := range net.SclFanins(id) {
		if marked[n] {
			return true
		}
	}
	for _, n := range net.SclFanouts(id) {
		if marked[n] {
			return true
		}
	}
	return false
}
func Abc_SclObjMarkFanFans(id int, marked map[int]bool) { marked[id] = true }
func Abc_SclFindUpsizes(p *SC_Man, path []int, state *UpsizeState, ratio, notches, iteration, gap int, conservative bool) (int, error) {
	if state == nil {
		return 0, fmt.Errorf("upsizing requires path state")
	}
	var candidates []upsizeCandidate
	for _, id := range path {
		if last := p.NodeIteration[id]; last >= 0 && int(last)+5 > iteration {
			continue
		}
		nodes, evals, e := Abc_SclFindNodesToUpdate(p.Network, id, state.Critical, -1)
		if e != nil {
			return 0, e
		}
		var gain float32
		cell, e := Abc_SclFindBestCell(p, id, nodes, evals, notches, gap, &gain)
		if e != nil {
			return 0, e
		}
		if cell >= 0 {
			if gain <= 0 {
				return 0, fmt.Errorf("upsizing candidate %d has nonpositive gain %g", id, gain)
			}
			candidates = append(candidates, upsizeCandidate{id: id, cell: cell, gain: gain})
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].gain > candidates[j].gain })
	limit := min(len(candidates), max(int(.01*float64(ratio)*float64(len(path))), 1))
	count := 0
	marked := map[int]bool{}
	for _, c := range candidates {
		if Abc_SclObjCheckMarkedFanFans(p.Network, c.id, marked) {
			continue
		}
		Abc_SclObjMarkFanFans(c.id, marked)
		old, next := Abc_SclObjCell(p, c.id), p.Lib.Cells[c.cell]
		if e := Abc_SclUpdateLoad(p, c.id, old, next); e != nil {
			return count, e
		}
		Abc_SclObjSetCell(p, c.id, next)
		p.SumArea += next.Area - old.Area
		p.Updates = append(p.Updates, int32(c.id), int32(c.cell))
		Abc_SclTimeIncInsert(p, c.id)
		p.NodeIteration[c.id] = int32(iteration)
		count++
		if count == limit {
			break
		}
	}
	return count, nil
}
func Abc_SclApplyUpdateToBest(best, gates []int32, updates *[]int32) error {
	if len(*updates)%2 != 0 {
		return fmt.Errorf("invalid cell update pairs")
	}
	for i := 0; i < len(*updates); i += 2 {
		id := int((*updates)[i])
		if id < 0 || id >= len(best) {
			return fmt.Errorf("cell update ID out of range")
		}
		best[id] = (*updates)[i+1]
	}
	*updates = nil
	for i, v := range best {
		if i >= len(gates) || gates[i] != v {
			return fmt.Errorf("best mapping inconsistent at %d", i)
		}
	}
	return nil
}

func cellIsInverter(p *SC_Man, id int) bool {
	c := Abc_SclObjCell(p, id)
	return c != nil && c.Inputs == 1 && c.Outputs == 1 && len(SC_CellFunc(c)) == 1 && SC_CellFunc(c)[0] == 0x5555555555555555
}
func hasPhases(net TimingNetwork) bool {
	p, ok := net.(PhasedTimingNetwork)
	return ok && p.SclHasPhases()
}
func Abc_SclFindBypasses(p *SC_Man, path []int, state *UpsizeState, ratio, notches, iteration, gap int, verbose bool) (int, error) {
	if state == nil {
		return 0, fmt.Errorf("bypass requires path state")
	}
	net, ok := p.Network.(MutableTimingNetwork)
	if !ok {
		return 0, fmt.Errorf("buffer bypass requires mutable standard-cell topology")
	}
	var candidates []upsizeCandidate
	for _, buffer := range path {
		fans := net.SclFanins(buffer)
		if len(fans) != 1 || !net.SclIsNode(fans[0]) {
			continue
		}
		driver := fans[0]
		extra := -1
		invert := hasPhases(net) && cellIsInverter(p, buffer)
		if !hasPhases(net) && cellIsInverter(p, buffer) {
			if !cellIsInverter(p, driver) {
				continue
			}
			parents := net.SclFanins(driver)
			if len(parents) != 1 || !net.SclIsNode(parents[0]) {
				continue
			}
			driver = parents[0]
			extra = buffer
		}
		nodes, evals, e := Abc_SclFindNodesToUpdate(net, driver, state.Critical, extra)
		if e != nil {
			return 0, e
		}
		best := upsizeCandidate{id: buffer, cell: -1, gain: 0, driver: driver, invert: invert}
		for _, fanout := range append([]int(nil), net.SclFanouts(buffer)...) {
			if !net.SclIsNode(fanout) || !state.Critical[fanout] {
				continue
			}
			pin := -1
			duplicate := false
			for i, fanin := range net.SclFanins(fanout) {
				if fanin == driver || extra >= 0 && fanin == extra {
					duplicate = true
				}
				if fanin == buffer {
					pin = i
				}
			}
			if duplicate || pin < 0 {
				continue
			}
			Abc_SclLoadStore3(p, buffer)
			if e = Abc_SclUpdateLoadSplit(p, buffer, fanout); e != nil {
				return 0, e
			}
			if e = net.SclPatchFanin(fanout, pin, driver); e != nil {
				Abc_SclLoadRestore3(p, buffer)
				return 0, e
			}
			var gain float32
			cell, trialErr := Abc_SclFindBestCell(p, driver, nodes, evals, notches, gap, &gain)
			Abc_SclLoadRestore3(p, buffer)
			undoErr := net.SclPatchFanin(fanout, pin, buffer)
			if trialErr != nil {
				return 0, trialErr
			}
			if undoErr != nil {
				return 0, undoErr
			}
			if cell >= 0 && gain > best.gain {
				best.cell = cell
				best.gain = gain
				best.fanout = fanout
				best.pin = pin
			}
		}
		if best.cell >= 0 {
			candidates = append(candidates, best)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].gain > candidates[j].gain })
	marked := map[int]bool{}
	count := 0
	for _, c := range candidates {
		if marked[c.id] || marked[c.fanout] || marked[c.driver] {
			continue
		}
		marked[c.id], marked[c.fanout], marked[c.driver] = true, true, true
		if e := Abc_SclUpdateLoadSplit(p, c.id, c.fanout); e != nil {
			return count, e
		}
		Abc_SclAddWireLoad(p, c.id, 1)
		Abc_SclAddWireLoad(p, c.driver, 1)
		if e := net.SclPatchFanin(c.fanout, c.pin, c.driver); e != nil {
			return count, e
		}
		Abc_SclAddWireLoad(p, c.id, 0)
		Abc_SclAddWireLoad(p, c.driver, 0)
		if c.invert {
			if e := net.(PhasedTimingNetwork).SclToggleFaninPhase(c.fanout, c.pin); e != nil {
				_ = net.SclPatchFanin(c.fanout, c.pin, c.id)
				return count, e
			}
		}
		state.Edits = append(state.Edits, BypassEdit{Fanout: c.fanout, Pin: c.pin, Driver: c.driver, Buffer: c.id, Inverted: c.invert})
		if e := Abc_SclTimeIncUpdateLevel(p, c.fanout); e != nil {
			return count, e
		}
		old, next := Abc_SclObjCell(p, c.driver), p.Lib.Cells[c.cell]
		if e := Abc_SclUpdateLoad(p, c.driver, old, next); e != nil {
			return count, e
		}
		Abc_SclObjSetCell(p, c.driver, next)
		p.SumArea += next.Area - old.Area
		p.Updates = append(p.Updates, int32(c.driver), int32(c.cell))
		for _, id := range []int{c.id, c.fanout, c.driver} {
			Abc_SclTimeIncInsert(p, id)
			p.NodeIteration[id] = -1
		}
		if verbose {
			fmt.Fprintf(p.Output, "Node %d redirect fanout %d to driver %d; gain %.1f ps, gate %s -> %s\n", c.id, c.fanout, c.driver, c.gain, old.Name, next.Name)
		}
		count++
	}
	return count, nil
}
func Abc_SclUndoRecentChanges(p *SC_Man, edits []BypassEdit) error {
	if len(edits) == 0 {
		return nil
	}
	net, ok := p.Network.(MutableTimingNetwork)
	if !ok {
		return fmt.Errorf("undo bypass requires mutable topology")
	}
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		if e := net.SclPatchFanin(edit.Fanout, edit.Pin, edit.Buffer); e != nil {
			return e
		}
		if edit.Inverted {
			if e := net.(PhasedTimingNetwork).SclToggleFaninPhase(edit.Fanout, edit.Pin); e != nil {
				return e
			}
		}
	}
	return p.rebuildOrder()
}
func Abc_SclUpsizeRemoveDangling(p *SC_Man, net TimingNetwork) error {
	mutable, ok := net.(MutableTimingNetwork)
	if !ok {
		return fmt.Errorf("dangling-node cleanup requires mutable topology")
	}
	for {
		changed := false
		for _, id := range append([]int(nil), net.SclObjectIDs()...) {
			if net.SclIsNode(id) && !net.SclIsBarrier(id) && len(net.SclFanouts(id)) == 0 {
				cell := Abc_SclObjCell(p, id)
				if cell != nil {
					p.SumArea -= cell.Area
				}
				if e := mutable.SclDeleteNode(id); e != nil {
					return e
				}
				p.Mapping.GateIDs[id] = -1
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return p.rebuildOrder()
}
func Abc_SclUpsizePrintDiffs(p *SC_Man, lib *SC_Lib, net TimingNetwork) error {
	times, slews, loads := append([]SC_Pair(nil), p.Times...), append([]SC_Pair(nil), p.Slews...), append([]SC_Pair(nil), p.Loads...)
	if e := Abc_SclTimeNtkRecompute(p, nil, nil, 0, 0); e != nil {
		return e
	}
	for _, id := range net.SclObjectIDs() {
		if !net.SclIsNode(id) {
			continue
		}
		for _, entry := range []struct {
			name string
			a, b SC_Pair
		}{{"time", times[id], p.Times[id]}, {"slew", slews[id], p.Slews[id]}, {"load", loads[id], p.Loads[id]}} {
			if math.Abs(float64(entry.a.Rise-entry.b.Rise)) > .001 || math.Abs(float64(entry.a.Fall-entry.b.Fall)) > .001 {
				fmt.Fprintf(p.Output, "%d: %s differs old=%v recomputed=%v\n", id, entry.name, entry.a, entry.b)
			}
		}
	}
	return nil
}
func Abc_SclUpsizePrint(p *SC_Man, iteration, window, outputs, nodes, upsizes, cone int, verbose bool) {
	end := "\r"
	if verbose {
		end = "\n"
	}
	fmt.Fprintf(p.Output, "%4d Win:%3d. PO:%6d. Path:%7d. Gate:%5d. TFO:%7d. A: %.2f D: %.2f ps B: %.2f ps%s", iteration, window, outputs, nodes, upsizes, cone, p.SumArea, p.MaxDelay, p.BestDelay, end)
}

func UpsizeManager(p *SC_Man, pars *SC_SizePars) error {
	if p == nil || pars == nil {
		return fmt.Errorf("upsizing requires manager and parameters")
	}
	if pars.Window <= 0 || pars.Window > 100 {
		return fmt.Errorf("upsizing window must be in 1..100")
	}
	if pars.BypassFrequency > 0 {
		if _, ok := p.Network.(MutableTimingNetwork); !ok {
			return fmt.Errorf("requested bypass optimization requires mutable topology")
		}
	}
	start := time.Now()
	var deadline time.Time
	if pars.Timeout > 0 {
		deadline = start.Add(time.Duration(pars.Timeout) * time.Second)
	}
	expired := func() bool { return !deadline.IsZero() && time.Now().After(deadline) }
	p.BestGates = append([]int32(nil), p.Mapping.GateIDs...)
	p.BestDelay = p.InitialDelay
	state := &UpsizeState{}
	noChange := 0
	window := pars.Window
	if window == 1 && len(sclMappedNodes(p.Network)) > 40000 {
		window++
	}
	if pars.Verbose {
		fmt.Fprintf(p.Output, "Parameters: Iters=%d Window=%d Ratio=%d UseDept=%t UseWL=%t Target=%d DelayGap=%d Timeout=%d\n", pars.Iterations, window, pars.Ratio, pars.UseDepartures, pars.UseWireLoads, pars.DelayUser, pars.DelayGap, pars.Timeout)
	}
	if p.BestDelay <= float32(pars.DelayUser) {
		fmt.Fprintf(p.Output, "Current delay (%.2f ps) does not exceed target delay (%.2f ps). Upsizing is not performed.\n", p.BestDelay, float32(pars.DelayUser))
	} else {
		for iteration := 0; iteration < pars.Iterations; iteration++ {
			count := 0
			var outputs, path []int
			win := window
			if iteration%7 == 6 {
				win++
			}
			bypass := pars.BypassFrequency > 0 && iteration > 0 && iteration%pars.BypassFrequency == 0
			for ; win <= 100; win *= 2 {
				coneStart := time.Now()
				outputs = Abc_SclFindCriticalCoWindow(p, win)
				path, state.Critical = Abc_SclFindCriticalNodeWindow(p, outputs, win, pars.UseDepartures)
				p.TimeCone += time.Since(coneStart)
				sizeStart := time.Now()
				var e error
				if bypass {
					count, e = Abc_SclFindBypasses(p, path, state, pars.Ratio, pars.Notches, iteration, pars.DelayGap, pars.VeryVerbose)
				} else {
					count, e = Abc_SclFindUpsizes(p, path, state, pars.Ratio, pars.Notches, iteration, pars.DelayGap, pars.BypassFrequency > 0)
				}
				p.TimeSize += time.Since(sizeStart)
				if e != nil {
					return e
				}
				Abc_SclUnmarkCriticalNodeWindow(state.Critical, path)
				Abc_SclUnmarkCriticalNodeWindow(state.Critical, outputs)
				if count > 0 {
					break
				}
			}
			if count == 0 {
				break
			}
			timingStart := time.Now()
			coneSize := 0
			if pars.UseDepartures {
				if len(p.Changed) > 0 && !bypass {
					var e error
					coneSize, e = Abc_SclTimeIncUpdate(p)
					if e != nil {
						return e
					}
				} else {
					if e := Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); e != nil {
						return e
					}
				}
			} else {
				cone := Abc_SclFindTFO(p.Network, path)
				if e := Abc_SclTimeCone(p, cone); e != nil {
					return e
				}
				coneSize = len(cone)
			}
			p.TimeTiming += time.Since(timingStart)
			p.MaxDelay = Abc_SclReadMaxDelay(p)
			if p.BestDelay > p.MaxDelay {
				p.BestDelay = p.MaxDelay
				if e := Abc_SclApplyUpdateToBest(p.BestGates, p.Mapping.GateIDs, &p.Updates); e != nil {
					return e
				}
				state.Edits = nil
				noChange = 0
			} else {
				noChange++
			}
			Abc_SclUpsizePrint(p, iteration, win, len(outputs), len(path), count, coneSize, pars.VeryVerbose || pars.Verbose && noChange == 0)
			if expired() || noChange > pars.IterationsNoChange || p.BestDelay <= float32(pars.DelayUser) {
				break
			}
		}
	}
	p.Mapping.GateIDs = append(p.Mapping.GateIDs[:0], p.BestGates...)
	if pars.BypassFrequency > 0 {
		if e := Abc_SclUndoRecentChanges(p, state.Edits); e != nil {
			return e
		}
		if e := Abc_SclUpsizeRemoveDangling(p, p.Network); e != nil {
			return e
		}
	}
	if e := Abc_SclTimeNtkRecompute(p, &p.SumArea, &p.MaxDelay, 0, 0); e != nil {
		return e
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
func Abc_SclUpsizePerformInt(lib *SC_Lib, net TimingNetwork, pars *SC_SizePars) error {
	if pars == nil {
		return fmt.Errorf("upsizing parameters are nil")
	}
	p, e := Abc_SclManStart(lib, net, boolInt(pars.UseWireLoads), boolInt(pars.UseDepartures), 0, int32(pars.BufferTreeEstimate))
	if e != nil {
		return e
	}
	defer Abc_SclManFree(p)
	return UpsizeManager(p, pars)
}
func Abc_SclUpsizePerform(lib *SC_Lib, net TimingNetwork, pars *SC_SizePars) error {
	return Abc_SclUpsizePerformInt(lib, net, pars)
}
