// Static and incremental timing from sclSize.c.
package scl

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func (p *SC_Man) rebuildOrder() error {
	ids := p.Network.SclObjectIDs()
	known := make(map[int]bool, len(ids))
	maximum := -1
	for _, id := range ids {
		if id < 0 || known[id] {
			return fmt.Errorf("SCL: invalid or duplicate object ID %d", id)
		}
		known[id] = true
		maximum = max(maximum, id)
	}
	p.ensureCapacity(maximum + 1)
	inputs, outputs := make(map[int]bool), make(map[int]bool)
	for _, id := range p.Network.SclInputs() {
		if !known[id] || inputs[id] || len(p.Network.SclFanins(id)) != 0 {
			return fmt.Errorf("SCL: invalid input %d", id)
		}
		inputs[id] = true
	}
	for _, id := range p.Network.SclOutputs() {
		if !known[id] || outputs[id] || inputs[id] || len(p.Network.SclFanins(id)) != 1 {
			return fmt.Errorf("SCL: invalid output %d", id)
		}
		outputs[id] = true
	}

	incoming := make(map[int]int, len(ids))
	reverse := make(map[int][]int, len(ids))
	for _, id := range ids {
		for _, fanin := range p.Network.SclFanins(id) {
			if !known[fanin] {
				return fmt.Errorf("SCL: object %d has missing fanin %d", id, fanin)
			}
			incoming[id]++
			reverse[fanin] = append(reverse[fanin], id)
		}
	}
	for _, id := range ids {
		actual := append([]int(nil), p.Network.SclFanouts(id)...)
		expected := append([]int(nil), reverse[id]...)
		sort.Ints(actual)
		sort.Ints(expected)
		if len(actual) != len(expected) {
			return fmt.Errorf("SCL: object %d fanout multiplicity mismatch", id)
		}
		for i := range actual {
			if actual[i] != expected[i] {
				return fmt.Errorf("SCL: object %d fanout adjacency mismatch", id)
			}
		}
	}
	var queue []int
	for _, id := range ids {
		if incoming[id] == 0 {
			queue = append(queue, id)
		}
	}
	sort.Ints(queue)
	order := make([]int, 0, len(ids))
	for cursor := 0; cursor < len(queue); cursor++ {
		id := queue[cursor]
		order = append(order, id)
		level := 0
		for _, fanin := range p.Network.SclFanins(id) {
			level = max(level, p.Levels[fanin]+1)
		}
		p.Levels[id] = level
		for _, fanout := range reverse[id] {
			incoming[fanout]--
			if incoming[fanout] == 0 {
				queue = append(queue, fanout)
			}
		}
	}
	if len(order) != len(ids) {
		return fmt.Errorf("SCL: combinational cycle in timing network")
	}
	p.order = order
	return nil
}
func Abc_SclFindCriticalCo(p *SC_Man, rise *int32) int {
	best := -1
	maximum := float32(math.Inf(-1))
	*rise = 0
	for _, id := range p.Network.SclOutputs() {
		if p.Times[id].Rise > maximum {
			best, maximum, *rise = id, p.Times[id].Rise, 1
		}
		if p.Times[id].Fall > maximum {
			best, maximum, *rise = id, p.Times[id].Fall, 0
		}
	}
	return best
}
func Abc_SclFindMostCriticalFanin2(p *SC_Man, rise *int32, id int) int {
	best := -1
	maximum := float32(math.Inf(-1))
	*rise = 0
	for _, fanin := range p.Network.SclFanins(id) {
		if p.Times[fanin].Rise > maximum {
			best, maximum, *rise = fanin, p.Times[fanin].Rise, 1
		}
		if p.Times[fanin].Fall > maximum {
			best, maximum, *rise = fanin, p.Times[fanin].Fall, 0
		}
	}
	return best
}
func Abc_SclFindMostCriticalFanin(p *SC_Man, rise *int32, id int) int {
	best := -1
	slack := float32(math.Inf(1))
	*rise = 0
	for _, fanin := range p.Network.SclFanins(id) {
		candidate := Abc_SclObjGetSlack(p, fanin, p.InitialDelay)
		if candidate < slack {
			best, slack = fanin, candidate
		}
	}
	if best >= 0 && p.Times[best].Rise >= p.Times[best].Fall {
		*rise = 1
	}
	return best
}
func Abc_SclTimeFanin(p *SC_Man, timing *SC_Timing, id, fanin int) {
	Scl_LibPinArrival(timing, &p.Times[fanin], &p.Slews[fanin], &p.Loads[id], &p.Times[id], &p.Slews[id])
}
func Abc_SclDeptFanin(p *SC_Man, timing *SC_Timing, id, fanin int) {
	Scl_LibPinDeparture(timing, &p.Departures[fanin], &p.Slews[fanin], &p.Loads[id], &p.Departures[id])
}
func Abc_SclTimeNode(p *SC_Man, id int, reverse int32) error {
	if _, input := p.inputIndex[id]; input {
		if reverse != 0 {
			return nil
		}
		if p.InputDriver != nil {
			Scl_LibHandleInputDriver(p.InputDriver, &p.Loads[id], &p.Times[id], &p.Slews[id])
		}
		return nil
	}
	fans := p.Network.SclFanins(id)
	if _, output := p.outputIndex[id]; output {
		if len(fans) != 1 {
			return fmt.Errorf("SCL: output %d needs one driver", id)
		}
		if reverse == 0 {
			p.Times[id], p.Slews[id] = p.Times[fans[0]], p.Slews[fans[0]]
			p.OutputTimes[p.outputIndex[id]] = SC_PairMax(&p.Times[id])
		}
		return nil
	}
	if len(fans) == 0 {
		return nil
	}
	if p.Network.SclIsBarrier(id) {
		if len(fans) != 1 {
			return fmt.Errorf("SCL: barrier %d needs one driver", id)
		}
		if reverse != 0 {
			p.Departures[fans[0]].Rise = max(p.Departures[fans[0]].Rise, p.Departures[id].Rise)
			p.Departures[fans[0]].Fall = max(p.Departures[fans[0]].Fall, p.Departures[id].Fall)
		} else {
			p.Times[id], p.Slews[id] = p.Times[fans[0]], p.Slews[fans[0]]
		}
		return nil
	}
	cell := Abc_SclObjCell(p, id)
	if cell == nil || cell.Inputs != len(fans) {
		return fmt.Errorf("SCL: object %d has invalid cell arity", id)
	}
	load, departure := p.Loads[id], p.Departures[id]
	value := float32(0)
	if p.EstLoadAve > 0 && p.EstLoadMax > 0 {
		value = SC_PairAve(&load) / (p.EstLoadAve * p.EstLoadMax)
	}
	if value > 1 {
		cap := p.EstLoadAve * p.EstLoadMax
		p.Loads[id] = SC_Pair{cap, cap}
		if reverse != 0 {
			delta := p.EstLinear * float32(math.Log(float64(value)))
			p.Departures[id].Rise += delta
			p.Departures[id].Fall += delta
		}
		p.EstimatedNodes++
	}
	defer func() {
		p.Loads[id] = load
		if reverse != 0 {
			p.Departures[id] = departure
		}
	}()
	if reverse == 0 {
		p.Times[id], p.Slews[id] = SC_Pair{}, SC_Pair{}
	}
	for pin, fanin := range fans {
		timing := Scl_CellPinTime(cell, pin)
		if timing == nil {
			return fmt.Errorf("SCL: missing timing arc for %s input %d", cell.Name, pin)
		}
		if reverse != 0 {
			Abc_SclDeptFanin(p, timing, id, fanin)
		} else {
			Abc_SclTimeFanin(p, timing, id, fanin)
		}
	}
	if value > 1 && reverse == 0 {
		delta := p.EstLinear * float32(math.Log(float64(value)))
		p.Times[id].Rise += delta
		p.Times[id].Fall += delta
	}
	return nil
}
func Abc_SclTimeCone(p *SC_Man, cone []int) error {
	// Sort by current network levels rather than trusting the caller's ID order.
	ordered := append([]int(nil), cone...)
	sort.SliceStable(ordered, func(i, j int) bool { return p.Levels[ordered[i]] < p.Levels[ordered[j]] })
	Abc_SclConeClean(p, ordered)
	for _, id := range ordered {
		if err := Abc_SclTimeNode(p, id, 0); err != nil {
			return err
		}
	}
	return nil
}
func Abc_SclTimeNtkRecompute(p *SC_Man, area, delay *float32, reverse int32, user float32) error {
	if !sclFinite(user) || user < 0 {
		return fmt.Errorf("SCL: invalid required delay")
	}
	if err := p.rebuildOrder(); err != nil {
		return err
	}
	if err := ensureSclMapping(p.Lib, p.Network); err != nil {
		return err
	}
	if err := Abc_SclComputeLoad(p); err != nil {
		return err
	}
	Abc_SclManCleanTime(p)
	p.EstimatedNodes = 0
	for _, id := range p.order {
		if err := Abc_SclTimeNode(p, id, 0); err != nil {
			return err
		}
	}
	measured := Abc_SclReadMaxDelay(p)
	if reverse != 0 {
		measured = max(measured, user)
		p.EstimatedNodes = 0
		for i := len(p.order) - 1; i >= 0; i-- {
			if err := Abc_SclTimeNode(p, p.order[i], 1); err != nil {
				return err
			}
		}
	}
	p.SumArea, p.MaxDelay = Abc_SclGetTotalArea(p.Network), measured
	if area != nil {
		*area = p.SumArea
	}
	if delay != nil {
		*delay = measured
	}
	return nil
}
func Abc_SclDeptObj(p *SC_Man, id int) error {
	p.Departures[id] = SC_Pair{}
	for _, fanout := range p.Network.SclFanouts(id) {
		if _, output := p.outputIndex[fanout]; output {
			continue
		}
		if p.Network.SclIsBarrier(fanout) {
			p.Departures[id].Rise = max(p.Departures[id].Rise, p.Departures[fanout].Rise)
			p.Departures[id].Fall = max(p.Departures[id].Fall, p.Departures[fanout].Fall)
			continue
		}
		cell := Abc_SclObjCell(p, fanout)
		if cell == nil {
			continue
		}
		for pin, fanin := range p.Network.SclFanins(fanout) {
			if fanin != id {
				continue
			}
			timing := Scl_CellPinTime(cell, pin)
			if timing == nil {
				return fmt.Errorf("SCL: missing departure timing arc")
			}
			load, departure := p.Loads[fanout], p.Departures[fanout]
			if p.EstLoadMax > 0 && p.EstLoadAve > 0 {
				value := SC_PairAve(&load) / (p.EstLoadMax * p.EstLoadAve)
				if value > 1 {
					cap := p.EstLoadMax * p.EstLoadAve
					p.Loads[fanout] = SC_Pair{cap, cap}
					delta := p.EstLinear * float32(math.Log(float64(value)))
					p.Departures[fanout].Rise += delta
					p.Departures[fanout].Fall += delta
				}
			}
			Abc_SclDeptFanin(p, timing, fanout, id)
			p.Loads[fanout], p.Departures[fanout] = load, departure
		}
	}
	return nil
}
func Abc_SclTimeIncInsert(p *SC_Man, id int) { p.Changed = append(p.Changed, id) }
func Abc_SclTimeIncUpdate(p *SC_Man) (int, error) {
	if len(p.Changed) == 0 {
		return 0, nil
	}
	if err := p.rebuildOrder(); err != nil {
		return 0, err
	}
	scheduled := make(map[int]bool)
	known := make(map[int]bool, len(p.order))
	for _, id := range p.order {
		known[id] = true
	}
	for _, id := range p.Changed {
		if id < 0 || id >= len(p.Times) || !known[id] {
			return 0, fmt.Errorf("SCL: invalid changed object %d", id)
		}
		scheduled[id] = true
		for _, fanin := range p.Network.SclFanins(id) {
			scheduled[fanin] = true
		}
	}
	p.Changed = nil
	updates := 0
	for _, id := range p.order {
		if !scheduled[id] {
			continue
		}
		updates++
		// A changed arc can alter a fanin's departure even when this node's
		// departure remains unchanged (for example after its output load changes).
		for _, fanin := range p.Network.SclFanins(id) {
			scheduled[fanin] = true
		}
		oldTime, oldSlew := p.Times[id], p.Slews[id]
		if err := Abc_SclTimeNode(p, id, 0); err != nil {
			return updates, err
		}
		if !SC_PairEqualE(&oldTime, &p.Times[id], .1) || !SC_PairEqualE(&oldSlew, &p.Slews[id], .1) {
			for _, fanout := range p.Network.SclFanouts(id) {
				scheduled[fanout] = true
			}
		}
	}
	for i := len(p.order) - 1; i >= 0; i-- {
		id := p.order[i]
		if !scheduled[id] {
			continue
		}
		before := p.Departures[id]
		if err := Abc_SclDeptObj(p, id); err != nil {
			return updates, err
		}
		if !SC_PairEqualE(&before, &p.Departures[id], .1) {
			for _, fanin := range p.Network.SclFanins(id) {
				scheduled[fanin] = true
			}
		}
	}
	p.IncrementalUpdates = updates
	p.MaxDelay = Abc_SclReadMaxDelay(p)
	p.SumArea = Abc_SclGetTotalArea(p.Network)
	return updates, nil
}
func Abc_SclTimeIncCheckLevel(p *SC_Man) error {
	for _, id := range p.Network.SclObjectIDs() {
		expected := 0
		for _, fanin := range p.Network.SclFanins(id) {
			if fanin < 0 || fanin >= len(p.Levels) {
				return fmt.Errorf("SCL: invalid fanin level")
			}
			expected = max(expected, p.Levels[fanin]+1)
		}
		if id < 0 || id >= len(p.Levels) || p.Levels[id] != expected {
			return fmt.Errorf("SCL: object %d level is stale", id)
		}
	}
	return nil
}
func Abc_SclTimeIncUpdateLevel(p *SC_Man, id int) error     { return p.rebuildOrder() }
func Abc_SclTimeIncUpdateLevel_rec(p *SC_Man, id int) error { return p.rebuildOrder() }
func Abc_SclManReadSlewAndLoad(p *SC_Man, network TimingNetwork) error {
	frame := abcmain.Abc_FrameGetGlobalFrame()
	load := frame.MaxLoad
	if !sclFinite(load) || load < 0 {
		return fmt.Errorf("SCL: invalid default output load")
	}
	for _, id := range network.SclOutputs() {
		p.Loads[id] = SC_Pair{load, load}
	}
	name := frame.StringValue("pDrivingCell")
	if name != "" {
		cell := Abc_SclCellFind(p.Lib, name)
		if cell < 0 || p.Lib.Cells[cell].Inputs != 1 {
			return fmt.Errorf("SCL: default input driving cell %s unavailable or not unary", name)
		}
		p.InputDriver = p.Lib.Cells[cell]
	}
	return nil
}
func Abc_SclManStart(lib *SC_Lib, network TimingNetwork, wire, depart int32, user float32, ratio int32) (*SC_Man, error) {
	p, err := NewTimingManager(lib, network)
	if err != nil {
		return nil, err
	}
	if ratio < 0 {
		return nil, fmt.Errorf("SCL: negative buffer-tree ratio")
	}
	if ratio > 0 {
		p.EstLoadMax = .01 * float32(ratio)
		p.EstLinear = 100
	}
	if err = Abc_SclManReadSlewAndLoad(p, network); err != nil {
		return nil, err
	}
	if wire != 0 {
		p.WireLoad = Abc_SclFindWireLoadModel(lib, Abc_SclGetTotalArea(network))
	}
	if err = Abc_SclTimeNtkRecompute(p, &p.InitialArea, &p.InitialDelay, depart, user); err != nil {
		return nil, err
	}
	return p, nil
}
func Abc_SclTimeNtkPrint(p *SC_Man, all, path int32) error {
	phase := int32(0)
	end := Abc_SclFindCriticalCo(p, &phase)
	if end < 0 {
		return fmt.Errorf("SCL: network has no output")
	}
	p.ReportDelay = SC_PairMax(&p.Times[end])
	writer := p.Output
	if writer == nil {
		writer = os.Stdout
	}
	if _, err := fmt.Fprintf(writer, "Gates = %d Area = %.2f Delay = %.2f\n", len(sclMappedNodes(p.Network)), p.SumArea, p.ReportDelay); err != nil {
		return err
	}
	printNode := func(id int) error {
		cell := Abc_SclObjCell(p, id)
		name := "terminal"
		if cell != nil {
			name = cell.Name
		}
		_, err := fmt.Fprintf(writer, "%d %s arrival=(%.3f,%.3f) slew=(%.3f,%.3f) load=(%.3f,%.3f) slack=%.3f\n", id, name, p.Times[id].Rise, p.Times[id].Fall, p.Slews[id].Rise, p.Slews[id].Fall, p.Loads[id].Rise, p.Loads[id].Fall, Abc_SclObjGetSlack(p, id, p.InitialDelay))
		return err
	}
	if all != 0 {
		for _, id := range p.order {
			if p.Network.SclIsNode(id) {
				if err := printNode(id); err != nil {
					return err
				}
			}
		}
	}
	if path != 0 {
		var trace []int
		seen := make(map[int]bool)
		for id := end; id >= 0; id = Abc_SclFindMostCriticalFanin(p, &phase, id) {
			if seen[id] {
				return fmt.Errorf("SCL: critical path cycle")
			}
			seen[id] = true
			trace = append(trace, id)
		}
		for i := len(trace) - 1; i >= 0; i-- {
			if err := printNode(trace[i]); err != nil {
				return err
			}
		}
	}
	return nil
}
func Abc_SclTimePerformInt(lib *SC_Lib, network TimingNetwork, ratio, wire, all, path, dump int32) error {
	p, err := Abc_SclManStart(lib, network, wire, 1, 0, ratio)
	if err != nil {
		return err
	}
	defer Abc_SclManFree(p)
	if err = Abc_SclTimeNtkPrint(p, all, path); err != nil {
		return err
	}
	if dump != 0 {
		return Abc_SclDumpStats(p, "stats.txt", 0)
	}
	return nil
}
func Abc_SclTimePerform(lib *SC_Lib, network TimingNetwork, ratio, wire, all, path, dump int32) error {
	return Abc_SclTimePerformInt(lib, network, ratio, wire, all, path, dump)
}
func Abc_SclCheckCommonInputs(network TimingNetwork, id, fanin int) bool {
	seen := make(map[int]bool)
	for _, input := range network.SclFanins(fanin) {
		seen[input] = true
	}
	for _, input := range network.SclFanins(id) {
		if seen[input] {
			return true
		}
	}
	return false
}
func Abc_SclPrintFaninPairs(p *SC_Man, network TimingNetwork) {
	for _, id := range sclMappedNodes(network) {
		for _, fanin := range network.SclFanins(id) {
			if network.SclIsNode(fanin) && len(network.SclFanouts(fanin)) == 1 && Abc_SclCheckCommonInputs(network, id, fanin) {
				fmt.Fprintf(p.Output, "Node %d and fanin %d have common inputs\n", id, fanin)
			}
		}
	}
}

func sclIsBuffer(network TimingNetwork, id int) bool {
	return network.SclIsNode(id) && len(network.SclFanins(id)) == 1
}
func Abc_SclHasBufferFanout(network TimingNetwork, id int) bool {
	for _, fanout := range network.SclFanouts(id) {
		if sclIsBuffer(network, fanout) {
			return true
		}
	}
	return false
}
func sclBufferWalk(network TimingNetwork, id int, visit func(int, bool) float32) float32 {
	memo := make(map[int]float32)
	active := make(map[int]bool)
	var walk func(int) float32
	walk = func(node int) float32 {
		if value, ok := memo[node]; ok {
			return value
		}
		if active[node] {
			panic("SCL: cycle in buffer tree")
		}
		active[node] = true
		buffer := sclIsBuffer(network, node)
		value := visit(node, buffer)
		if buffer {
			for _, fanout := range network.SclFanouts(node) {
				value += walk(fanout)
			}
		}
		delete(active, node)
		memo[node] = value
		return value
	}
	return walk(id)
}
func Abc_SclCountBufferFanoutsInt(network TimingNetwork, id int) int {
	memo, active := make(map[int]int), make(map[int]bool)
	var walk func(int) int
	walk = func(node int) int {
		if value, ok := memo[node]; ok {
			return value
		}
		if active[node] {
			panic("SCL: cycle in buffer tree")
		}
		active[node] = true
		value := 0
		if sclIsBuffer(network, node) {
			value++
		}
		for _, fanout := range network.SclFanouts(node) {
			if sclIsBuffer(network, fanout) {
				value += walk(fanout)
			}
		}
		delete(active, node)
		memo[node] = value
		return value
	}
	return walk(id)
}
func Abc_SclCountBufferFanouts(network TimingNetwork, id int) int {
	count := Abc_SclCountBufferFanoutsInt(network, id)
	if sclIsBuffer(network, id) {
		count--
	}
	return count
}
func Abc_SclCountNonBufferFanoutsInt(network TimingNetwork, id int) int {
	return int(sclBufferWalk(network, id, func(_ int, buffer bool) float32 {
		if !buffer {
			return 1
		}
		return 0
	}))
}
func Abc_SclCountNonBufferFanouts(network TimingNetwork, id int) int {
	count := 0
	for _, fanout := range network.SclFanouts(id) {
		count += Abc_SclCountNonBufferFanoutsInt(network, fanout)
	}
	return count
}
func Abc_SclCountNonBufferDelayInt(p *SC_Man, id int) float32 {
	return sclBufferWalk(p.Network, id, func(node int, buffer bool) float32 {
		if buffer {
			return 0
		}
		return SC_PairMax(&p.Times[node])
	})
}
func Abc_SclCountNonBufferDelay(p *SC_Man, id int) float32 {
	var total float32
	for _, fanout := range p.Network.SclFanouts(id) {
		total += Abc_SclCountNonBufferDelayInt(p, fanout)
	}
	return total
}
func Abc_SclCountNonBufferLoadInt(p *SC_Man, id int) float32 {
	return sclBufferWalk(p.Network, id, func(node int, buffer bool) float32 {
		if !buffer {
			return 0
		}
		cell := Abc_SclObjCell(p, node)
		cap := float32(0)
		if cell != nil && cell.Inputs > 0 {
			cap = SC_CellPinCap(cell, 0)
		}
		return SC_PairAve(&p.Loads[node]) - cap
	})
}
func Abc_SclCountNonBufferLoad(p *SC_Man, id int) float32 {
	total := SC_PairAve(&p.Loads[id])
	for _, fanout := range p.Network.SclFanouts(id) {
		total += Abc_SclCountNonBufferLoadInt(p, fanout)
	}
	return total
}
func Abc_SclPrintBuffersOne(p *SC_Man, id, offset int) error {
	cell := Abc_SclObjCell(p, id)
	name := "pi"
	area := float32(0)
	if cell != nil {
		name, area = cell.Name, cell.Area
	}
	writer := p.Output
	if writer == nil {
		writer = os.Stdout
	}
	_, err := fmt.Fprintf(writer, "%*s%d: %s fanout=%d buffers=%d endpoints=%d area=%.2f delay=(%.2f,%.2f) load=%.2f slew=%.2f slack=%.2f\n", max(offset, 0)*4, "", id, name, len(p.Network.SclFanouts(id)), Abc_SclCountBufferFanouts(p.Network, id), Abc_SclCountNonBufferFanouts(p.Network, id), area, p.Times[id].Rise, p.Times[id].Fall, SC_PairMax(&p.Loads[id]), SC_PairMax(&p.Slews[id]), Abc_SclObjGetSlack(p, id, p.InitialDelay))
	return err
}
func Abc_SclPrintBuffersInt(p *SC_Man, id, offset int) error {
	if err := Abc_SclPrintBuffersOne(p, id, offset); err != nil {
		return err
	}
	for _, fanout := range p.Network.SclFanouts(id) {
		if sclIsBuffer(p.Network, fanout) {
			if err := Abc_SclPrintBuffersInt(p, fanout, offset+1); err != nil {
				return err
			}
		}
	}
	return nil
}
func Abc_SclPrintBufferTrees(p *SC_Man, network TimingNetwork) error {
	for _, id := range network.SclObjectIDs() {
		if !sclIsBuffer(network, id) && Abc_SclCountBufferFanouts(network, id) > 3 {
			if err := Abc_SclPrintBuffersOne(p, id, 0); err != nil {
				return err
			}
			for _, fanout := range network.SclFanouts(id) {
				if sclIsBuffer(network, fanout) {
					if err := Abc_SclPrintBuffersInt(p, fanout, 1); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
func Abc_SclPrintBuffers(lib *SC_Lib, network TimingNetwork, verbose int32) error {
	p, err := Abc_SclManStart(lib, network, 0, 1, 0, 10000)
	if err != nil {
		return err
	}
	defer Abc_SclManFree(p)
	return Abc_SclPrintBufferTrees(p, network)
}
func Abc_SclDumpStats(p *SC_Man, filename string, elapsed time.Duration) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	name := p.Network.SclName()
	nodes := len(sclMappedNodes(p.Network))
	if p.statsName != name || p.statsStart.IsZero() {
		_, err = fmt.Fprintf(file, "\n%s %d %d %d %.0f %.0f ", filepath.Base(name), len(p.Network.SclInputs()), len(p.Network.SclOutputs()), nodes, p.SumArea, p.ReportDelay)
		if err == nil {
			p.statsName, p.statsNodes, p.statsArea, p.statsDelay, p.statsStart = name, nodes, p.SumArea, p.ReportDelay, time.Now()
		}
	} else {
		ratio := func(value, baseline float32) float32 {
			if baseline == 0 {
				return 0
			}
			return 100 * value / baseline
		}
		if elapsed <= 0 {
			elapsed = time.Since(p.statsStart)
		}
		_, err = fmt.Fprintf(file, " %.1f %.1f %.1f %.2f", ratio(float32(nodes), float32(p.statsNodes)), ratio(p.SumArea, p.statsArea), ratio(p.ReportDelay, p.statsDelay), elapsed.Seconds())
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}
