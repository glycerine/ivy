package scl

// Mapped-network topology buffering from ABC's sclBuffer.c.
import (
	"container/heap"
	"fmt"
	"github.com/glycerine/abcgo/map/mio"
	"io"
	"math"
	"os"
	"sort"
)

const BUF_SCALE = 1000

// BufferingTimingNetwork creates a mapped node and updates fanout adjacency.
// Creation preserves pin order; IDs may be sparse and need not be contiguous.
type BufferingTimingNetwork interface {
	MutableTimingNetwork
	SclCreateNode(fanins []int, gateID int32) (int, error)
}
type CloneableTimingNetwork interface{ SclClone() (TimingNetwork, error) }
type ReadablePhases interface {
	PhasedTimingNetwork
	SclFaninPhase(node, pin int) bool
}
type ConfigurablePhases interface {
	ReadablePhases
	SclSetPhases(enabled bool) error
}
type FanoutPin struct{ Node, Pin int }

func bufferCell(net TimingNetwork, id int) *SC_Cell {
	m := net.SclMapping()
	if m == nil || m.Lib == nil || id < 0 || id >= len(m.GateIDs) {
		return nil
	}
	gate := m.GateIDs[id]
	if gate < 0 || int(gate) >= len(m.Lib.Cells) {
		return nil
	}
	return m.Lib.Cells[gate]
}
func bufferPhase(net TimingNetwork, node, pin int) bool {
	if phases, ok := net.(ReadablePhases); ok && phases.SclHasPhases() {
		return phases.SclFaninPhase(node, pin)
	}
	return false
}
func bufferEdges(net TimingNetwork, driver int) []FanoutPin {
	var edges []FanoutPin
	seen := make(map[int]bool)
	for _, node := range net.SclFanouts(driver) {
		if seen[node] {
			continue
		}
		seen[node] = true
		for pin, id := range net.SclFanins(node) {
			if id == driver {
				edges = append(edges, FanoutPin{node, pin})
			}
		}
	}
	return edges
}
func bufferInputs(net TimingNetwork) map[int]bool {
	set := make(map[int]bool)
	for _, id := range net.SclInputs() {
		set[id] = true
	}
	return set
}
func bufferOutputs(net TimingNetwork) map[int]bool {
	set := make(map[int]bool)
	for _, id := range net.SclOutputs() {
		set[id] = true
	}
	return set
}
func bufferOrder(net TimingNetwork) ([]int, error) {
	if net == nil {
		return nil, fmt.Errorf("SCL: nil buffering network")
	}
	ids := make(map[int]bool)
	for _, id := range net.SclObjectIDs() {
		if id < 0 || ids[id] {
			return nil, fmt.Errorf("SCL: invalid/duplicate object ID %d", id)
		}
		ids[id] = true
	}
	for _, id := range net.SclInputs() {
		if !ids[id] || len(net.SclFanins(id)) != 0 {
			return nil, fmt.Errorf("SCL: invalid combinational input %d", id)
		}
	}
	for _, id := range net.SclOutputs() {
		if !ids[id] || len(net.SclFanins(id)) != 1 {
			return nil, fmt.Errorf("SCL: invalid combinational output %d", id)
		}
	}
	state := make(map[int]uint8)
	order := make([]int, 0, len(ids))
	var visit func(int) error
	visit = func(id int) error {
		if !ids[id] {
			return fmt.Errorf("SCL: missing fanin %d", id)
		}
		if state[id] == 1 {
			return fmt.Errorf("SCL: cycle at object %d", id)
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		for _, fanin := range net.SclFanins(id) {
			if err := visit(fanin); err != nil {
				return err
			}
		}
		state[id] = 2
		order = append(order, id)
		return nil
	}
	for _, id := range net.SclObjectIDs() {
		if err := visit(id); err != nil {
			return nil, err
		}
	}
	return order, nil
}
func cloneBufferNetwork(net TimingNetwork) (BufferingTimingNetwork, error) {
	if _, err := bufferOrder(net); err != nil {
		return nil, err
	}
	clone, ok := net.(CloneableTimingNetwork)
	if !ok {
		return nil, fmt.Errorf("SCL: buffering requires network cloning")
	}
	copied, err := clone.SclClone()
	if err != nil {
		return nil, err
	}
	if copied != nil && copied.SclMapping() == net.SclMapping() {
		return nil, fmt.Errorf("SCL: cloning returned shared mapping storage")
	}
	result, ok := copied.(BufferingTimingNetwork)
	if !ok {
		return nil, fmt.Errorf("SCL: network clone cannot create mapped nodes")
	}
	if phases, ok := result.(PhasedTimingNetwork); ok && phases.SclHasPhases() {
		if _, readable := result.(ReadablePhases); !readable {
			return nil, fmt.Errorf("SCL: network clone cannot read its fanin phases")
		}
	}
	if result.SclMapping() == nil || result.SclMapping().Lib == nil {
		return nil, fmt.Errorf("SCL: buffering requires a mapped cell library")
	}
	if err := ensureSclMapping(result.SclMapping().Lib, result); err != nil {
		return nil, err
	}
	return result, nil
}
func finishBufferNetwork(net BufferingTimingNetwork) (TimingNetwork, error) {
	order, err := bufferOrder(net)
	if err != nil {
		return nil, err
	}
	reachable := make(map[int]bool)
	var visit func(int)
	visit = func(id int) {
		if reachable[id] {
			return
		}
		reachable[id] = true
		for _, fanin := range net.SclFanins(id) {
			visit(fanin)
		}
	}
	for _, output := range net.SclOutputs() {
		visit(output)
	}
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		if net.SclIsNode(id) && !reachable[id] && !net.SclIsBarrier(id) {
			if err := net.SclDeleteNode(id); err != nil {
				return nil, err
			}
		}
	}
	if m := net.SclMapping(); m.Mio != nil {
		if err := Abc_SclSclGates2MioGates(m.Lib, net); err != nil {
			return nil, err
		}
	}
	if topo, ok := net.(interface{ SclCloneTopological() (TimingNetwork, error) }); ok {
		return topo.SclCloneTopological()
	}
	return net, nil
}
func bufferUnaryCell(net TimingNetwork, invert bool) (*SC_Cell, error) {
	m := net.SclMapping()
	if m == nil || m.Lib == nil {
		return nil, fmt.Errorf("SCL: missing buffer library")
	}
	flag := int32(1)
	if invert {
		flag = 0
	}
	cell := Abc_SclFindInvertor(m.Lib, flag)
	if cell == nil {
		return nil, fmt.Errorf("SCL: library lacks requested unary cell")
	}
	return cell, nil
}
func assignBufferCell(net TimingNetwork, id int, cell *SC_Cell) error {
	m := net.SclMapping()
	if id < 0 || m == nil || m.Lib == nil || cell == nil || cell.ID < 0 || cell.ID >= len(m.Lib.Cells) || m.Lib.Cells[cell.ID] != cell {
		return fmt.Errorf("SCL: invalid mapped cell assignment")
	}
	var gate *mio.Mio_Gate_t
	if m.Mio != nil {
		gate = mio.Mio_LibraryReadGateByName(m.Mio, cell.Name, "")
		if gate == nil {
			return fmt.Errorf("SCL: MIO gate %s unavailable", cell.Name)
		}
	}
	for len(m.GateIDs) <= id {
		m.GateIDs = append(m.GateIDs, -1)
	}
	m.GateIDs[id] = int32(cell.ID)
	if gate != nil {
		for len(m.MioGates) <= id {
			m.MioGates = append(m.MioGates, nil)
		}
		m.MioGates[id] = gate
	}
	return nil
}
func createBufferNode(net BufferingTimingNetwork, fanins []int, cell *SC_Cell) (int, error) {
	if cell == nil || cell.Inputs != len(fanins) {
		return -1, fmt.Errorf("SCL: invalid new mapped cell")
	}
	id, err := net.SclCreateNode(append([]int(nil), fanins...), int32(cell.ID))
	if err != nil {
		return -1, err
	}
	if id < 0 || !net.SclIsNode(id) {
		return -1, fmt.Errorf("SCL: network created invalid node ID")
	}
	if err = assignBufferCell(net, id, cell); err != nil {
		return -1, err
	}
	return id, nil
}
func cloneBufferNode(net BufferingTimingNetwork, id int) (int, error) {
	cell := bufferCell(net, id)
	clone, err := createBufferNode(net, net.SclFanins(id), cell)
	if err != nil {
		return -1, err
	}
	if phases, ok := net.(ReadablePhases); ok && phases.SclHasPhases() {
		for pin := range net.SclFanins(id) {
			if phases.SclFaninPhase(id, pin) {
				if err := phases.SclToggleFaninPhase(clone, pin); err != nil {
					return -1, err
				}
			}
		}
	}
	return clone, nil
}
func Abc_SclReportDupFanins(net TimingNetwork) []FanoutPin {
	var duplicates []FanoutPin
	for _, node := range net.SclObjectIDs() {
		seen := make(map[int]bool)
		for pin, id := range net.SclFanins(node) {
			if seen[id] {
				duplicates = append(duplicates, FanoutPin{node, pin})
			}
			seen[id] = true
		}
	}
	return duplicates
}
func Abc_SclObjIsBufInv(net TimingNetwork, id int) bool {
	cell := bufferCell(net, id)
	if !net.SclIsNode(id) || net.SclIsBarrier(id) || len(net.SclFanins(id)) != 1 || cell == nil {
		return false
	}
	truth := SC_CellFunc(cell)
	return len(truth) == 1 && (truth[0] == 0xaaaaaaaaaaaaaaaa || truth[0] == 0x5555555555555555)
}
func Abc_SclIsInv(net TimingNetwork, id int) bool {
	cell := bufferCell(net, id)
	return cell != nil && cell.Inputs == 1 && len(SC_CellFunc(cell)) == 1 && SC_CellFunc(cell)[0] == 0x5555555555555555
}
func Abc_SclGetRealFaninLit(net TimingNetwork, id int) (int, error) {
	active := make(map[int]bool)
	inverted := false
	for Abc_SclObjIsBufInv(net, id) {
		if active[id] {
			return -1, fmt.Errorf("SCL: unary cycle at %d", id)
		}
		active[id] = true
		inverted = inverted != Abc_SclIsInv(net, id)
		inverted = inverted != bufferPhase(net, id, 0)
		id = net.SclFanins(id)[0]
	}
	lit := id << 1
	if inverted {
		lit |= 1
	}
	return lit, nil
}
func Abc_SclCheckNtk(net TimingNetwork, verbose bool) bool {
	order, err := bufferOrder(net)
	if err != nil {
		return false
	}
	for _, id := range order {
		if net.SclIsNode(id) && !net.SclIsBarrier(id) && len(net.SclFanouts(id)) == 0 {
			return false
		}
	}
	if verbose {
		fmt.Println("SCL: acyclic network without dangling nodes")
	}
	return true
}
func Abc_SclUnBufferPerform(network TimingNetwork, verbose bool) (TimingNetwork, error) {
	net, err := cloneBufferNetwork(network)
	if err != nil {
		return nil, err
	}
	order, _ := bufferOrder(net)
	cache := make(map[int]int)
	for _, id := range order {
		if Abc_SclIsInv(net, id) {
			fanin := net.SclFanins(id)[0]
			if !Abc_SclObjIsBufInv(net, fanin) && !bufferPhase(net, id, 0) {
				cache[fanin] = id
			}
		}
	}
	for _, id := range order {
		for pin, fanin := range append([]int(nil), net.SclFanins(id)...) {
			if !Abc_SclObjIsBufInv(net, fanin) {
				continue
			}
			lit, e := Abc_SclGetRealFaninLit(net, fanin)
			if e != nil {
				return nil, e
			}
			driver := lit >> 1
			if lit&1 != 0 {
				inverse, ok := cache[driver]
				if !ok {
					cell, e := bufferUnaryCell(net, true)
					if e != nil {
						return nil, e
					}
					inverse, e = createBufferNode(net, []int{driver}, cell)
					if e != nil {
						return nil, e
					}
					cache[driver] = inverse
				}
				driver = inverse
			}
			if driver != fanin {
				if err := net.SclPatchFanin(id, pin, driver); err != nil {
					return nil, err
				}
			}
		}
	}
	return finishBufferNetwork(net)
}
func Abc_SclCountMaxPhases(net TimingNetwork) int {
	order, err := bufferOrder(net)
	if err != nil {
		return -1
	}
	levels := make(map[int]int)
	maximum := 0
	for _, id := range order {
		for pin, fanin := range net.SclFanins(id) {
			level := levels[fanin]
			if bufferPhase(net, id, pin) {
				level++
			}
			levels[id] = max(levels[id], level)
		}
		maximum = max(maximum, levels[id])
	}
	return maximum
}
func Abc_SclBufferPhase(network TimingNetwork, verbose bool) (TimingNetwork, error) {
	net, err := cloneBufferNetwork(network)
	if err != nil {
		return nil, err
	}
	phases, ok := net.(ConfigurablePhases)
	if !ok || !phases.SclHasPhases() {
		return nil, fmt.Errorf("SCL: explicit phase materialization requires phase metadata")
	}
	cell, err := bufferUnaryCell(net, true)
	if err != nil {
		return nil, err
	}
	inputs := bufferInputs(net)
	cache := make(map[int]int)
	order, _ := bufferOrder(net)
	for _, id := range order {
		for pin, fanin := range append([]int(nil), net.SclFanins(id)...) {
			if !phases.SclFaninPhase(id, pin) {
				continue
			}
			inverse, ok := cache[fanin]
			if !ok || inputs[fanin] {
				inverse, err = createBufferNode(net, []int{fanin}, cell)
				if err != nil {
					return nil, err
				}
				cache[fanin] = inverse
			}
			if err = net.SclPatchFanin(id, pin, inverse); err != nil {
				return nil, err
			}
			if err = phases.SclToggleFaninPhase(id, pin); err != nil {
				return nil, err
			}
		}
	}
	if err = phases.SclSetPhases(false); err != nil {
		return nil, err
	}
	return finishBufferNetwork(net)
}
func Abc_SclUnBufferPhase(network TimingNetwork, verbose bool) (TimingNetwork, error) {
	net, err := cloneBufferNetwork(network)
	if err != nil {
		return nil, err
	}
	phases, ok := net.(ConfigurablePhases)
	if !ok {
		return nil, fmt.Errorf("SCL: network cannot store fanin phases")
	}
	if err = phases.SclSetPhases(true); err != nil {
		return nil, err
	}
	order, _ := bufferOrder(net)
	for _, id := range order {
		if Abc_SclObjIsBufInv(net, id) {
			continue
		}
		for pin, fanin := range append([]int(nil), net.SclFanins(id)...) {
			lit, e := Abc_SclGetRealFaninLit(net, fanin)
			if e != nil {
				return nil, e
			}
			if lit>>1 == fanin && lit&1 == 0 {
				continue
			}
			if err = net.SclPatchFanin(id, pin, lit>>1); err != nil {
				return nil, err
			}
			if lit&1 != 0 {
				if err = phases.SclToggleFaninPhase(id, pin); err != nil {
					return nil, err
				}
			}
		}
	}
	return finishBufferNetwork(net)
}
func Abc_NodeInvUpdateObjFanoutPolarity(net TimingNetwork, driver int, edge FanoutPin) error {
	fans := net.SclFanins(edge.Node)
	if edge.Pin < 0 || edge.Pin >= len(fans) || fans[edge.Pin] != driver {
		return fmt.Errorf("SCL: invalid inverter fanout edge")
	}
	phases, ok := net.(ReadablePhases)
	if !ok || !phases.SclHasPhases() {
		return fmt.Errorf("SCL: inverter insertion requires fanin phases")
	}
	if net.SclIsBarrier(edge.Node) || Abc_SclObjIsBufInv(net, edge.Node) {
		return Abc_NodeInvUpdateFanPolarity(net, edge.Node)
	}
	return phases.SclToggleFaninPhase(edge.Node, edge.Pin)
}
func Abc_NodeInvUpdateFanPolarity(net TimingNetwork, id int) error {
	for _, edge := range bufferEdges(net, id) {
		if err := Abc_NodeInvUpdateObjFanoutPolarity(net, id, edge); err != nil {
			return err
		}
	}
	return nil
}
func bufferReverseLevels(net TimingNetwork) (map[int]int, error) {
	order, err := bufferOrder(net)
	if err != nil {
		return nil, err
	}
	levels := make(map[int]int)
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		for _, fanout := range net.SclFanouts(id) {
			levels[id] = max(levels[id], levels[fanout]+1)
		}
	}
	return levels, nil
}
func Abc_SclComputeReverseLevel(net TimingNetwork, id int) int {
	levels, err := bufferReverseLevels(net)
	if err != nil {
		return -1
	}
	level := 0
	for _, fanout := range net.SclFanouts(id) {
		level = max(level, levels[fanout])
	}
	return level + 1
}
func Abc_NodeCompareLevels(a, b int, levels map[int]int) int {
	if levels[a] < levels[b] {
		return -1
	}
	if levels[a] > levels[b] {
		return 1
	}
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
func Abc_SclPerformBufferingOne(net BufferingTimingNetwork, id, degree int, useInverters, verbose bool) (int, error) {
	if degree < 2 {
		return -1, fmt.Errorf("SCL: buffer degree must be at least two")
	}
	edges := bufferEdges(net, id)
	if len(edges) <= degree {
		return -1, fmt.Errorf("SCL: object %d is within fanout bound", id)
	}
	levels, err := bufferReverseLevels(net)
	if err != nil {
		return -1, err
	}
	sort.SliceStable(edges, func(i, j int) bool { return Abc_NodeCompareLevels(edges[i].Node, edges[j].Node, levels) < 0 })
	originalDegree := degree
	if len(edges) < 2*degree {
		previous, current, last := edges[len(edges)-1-degree], edges[degree-1], edges[len(edges)-1]
		if levels[current.Node] == levels[last.Node] && levels[previous.Node] < levels[current.Node] {
			first := 0
			for first < len(edges) && levels[edges[first].Node] != levels[last.Node] {
				first++
			}
			if first > 1 {
				degree = first
			}
		}
		if degree == originalDegree && degree > len(edges)-degree {
			degree = (len(edges) + 1) / 2
		}
	}
	cell, err := bufferUnaryCell(net, useInverters)
	if err != nil {
		return -1, err
	}
	buffer, err := createBufferNode(net, []int{id}, cell)
	if err != nil {
		return -1, err
	}
	for _, edge := range edges[:degree] {
		if err = net.SclPatchFanin(edge.Node, edge.Pin, buffer); err != nil {
			return -1, err
		}
		if useInverters {
			if err = Abc_NodeInvUpdateObjFanoutPolarity(net, buffer, edge); err != nil {
				return -1, err
			}
		}
	}
	return buffer, nil
}
func Abc_SclPerformBuffering_rec(net BufferingTimingNetwork, id, rootDegree, degree int, useInverters, verbose bool, visited map[int]bool) error {
	if visited[id] {
		return nil
	}
	visited[id] = true
	for _, fanout := range append([]int(nil), net.SclFanouts(id)...) {
		if err := Abc_SclPerformBuffering_rec(net, fanout, rootDegree, degree, useInverters, verbose, visited); err != nil {
			return err
		}
	}
	old := len(bufferEdges(net, id))
	for len(bufferEdges(net, id)) > degree {
		if _, err := Abc_SclPerformBufferingOne(net, id, degree, useInverters, verbose); err != nil {
			return err
		}
	}
	if rootDegree > 0 && old > rootDegree {
		cell, err := bufferUnaryCell(net, useInverters)
		if err != nil {
			return err
		}
		edges := bufferEdges(net, id)
		buffer, err := createBufferNode(net, []int{id}, cell)
		if err != nil {
			return err
		}
		for _, edge := range edges {
			if err = net.SclPatchFanin(edge.Node, edge.Pin, buffer); err != nil {
				return err
			}
			if useInverters {
				if err = Abc_NodeInvUpdateObjFanoutPolarity(net, buffer, edge); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func Abc_SclPerformBuffering(network TimingNetwork, rootDegree, degree int, useInverters, verbose bool) (TimingNetwork, error) {
	if degree < 2 || rootDegree < 0 {
		return nil, fmt.Errorf("SCL: invalid buffering degree")
	}
	net, err := cloneBufferNetwork(network)
	if err != nil {
		return nil, err
	}
	if useInverters {
		phases, ok := net.(ReadablePhases)
		if !ok || !phases.SclHasPhases() {
			return nil, fmt.Errorf("SCL: inverter buffering requires phases")
		}
	}
	visited := make(map[int]bool)
	for _, input := range net.SclInputs() {
		if err = Abc_SclPerformBuffering_rec(net, input, rootDegree, degree, useInverters, verbose, visited); err != nil {
			return nil, err
		}
	}
	return finishBufferNetwork(net)
}

// Buf_Man_t implements the older critical/noncritical separation heuristic.
// Timing arrays are maps keyed by stable IDs; no spare object capacity is kept.
type Buf_Man_t struct {
	Network                                                               BufferingTimingNetwork
	FanMin, FanMax                                                        int
	BufferPIs                                                             bool
	Arrivals, Departures                                                  map[int]int
	Edges                                                                 map[int][]int
	DelayMax                                                              int
	DelayInv                                                              float32
	NonCritical                                                           []FanoutPin
	Separate, Duplicate, BranchBuffers, BranchInverters, CriticalBranches int
	Output                                                                io.Writer
}

func Abc_BufNodeArr(p *Buf_Man_t, id int) int              { return p.Arrivals[id] }
func Abc_BufNodeDep(p *Buf_Man_t, id int) int              { return p.Departures[id] }
func Abc_BufSetNodeArr(p *Buf_Man_t, id, value int)        { p.Arrivals[id] = value }
func Abc_BufSetNodeDep(p *Buf_Man_t, id, value int)        { p.Departures[id] = value }
func Abc_BufEdgeDelay(p *Buf_Man_t, id, pin int) int       { return p.Edges[id][pin] }
func Abc_BufSetEdgeDelay(p *Buf_Man_t, id, pin, value int) { p.Edges[id][pin] = value }
func Abc_BufNodeSlack(p *Buf_Man_t, id int) int {
	return p.DelayMax - p.Arrivals[id] - p.Departures[id]
}
func Abc_BufEdgeSlack(p *Buf_Man_t, id int, edge FanoutPin) int {
	return p.DelayMax - p.Arrivals[id] - p.Departures[edge.Node] - p.Edges[edge.Node][edge.Pin]
}
func bufferPinEstimate(cell *SC_Cell, pin int, slew, load float32) (float32, error) {
	if cell == nil || cell.Outputs != 1 || pin < 0 || pin >= cell.Inputs || len(cell.Pins) <= cell.Inputs || cell.Pins[cell.Inputs] == nil {
		return 0, fmt.Errorf("SCL: invalid timing cell")
	}
	groups := cell.Pins[cell.Inputs].Timings
	if pin >= len(groups) || groups[pin] == nil || len(groups[pin].Timings) == 0 || groups[pin].Timings[0] == nil {
		return 0, fmt.Errorf("SCL: missing timing arc in %s", cell.Name)
	}
	arc := groups[pin].Timings[0]
	for _, surface := range []*SC_Surface{&arc.CellRise, &arc.CellFall, &arc.RiseTrans, &arc.FallTrans} {
		a, b := len(surface.Index0), len(surface.Index1)
		if a == 0 || b == 0 || (a == 1) != (b == 1) || len(surface.Data) != a {
			return 0, fmt.Errorf("SCL: malformed timing surface in %s", cell.Name)
		}
		for _, axis := range [][]float32{surface.Index0, surface.Index1} {
			for i, v := range axis {
				if !sclFinite(v) || i > 0 && v <= axis[i-1] {
					return 0, fmt.Errorf("SCL: invalid timing axis in %s", cell.Name)
				}
			}
		}
		for _, row := range surface.Data {
			if len(row) != b {
				return 0, fmt.Errorf("SCL: malformed timing row in %s", cell.Name)
			}
			for _, v := range row {
				if !sclFinite(v) {
					return 0, fmt.Errorf("SCL: nonfinite timing value in %s", cell.Name)
				}
			}
		}
	}
	value := Scl_LibPinArrivalEstimate(cell, pin, slew, load)
	if !sclFinite(value) || value < 0 {
		return 0, fmt.Errorf("SCL: invalid estimated delay in %s", cell.Name)
	}
	return value, nil
}
func Abc_BufCreateEdges(p *Buf_Man_t, id int) error {
	fans := p.Network.SclFanins(id)
	delays := make([]int, len(fans))
	outputs := bufferOutputs(p.Network)
	if !outputs[id] && !p.Network.SclIsBarrier(id) {
		cell := bufferCell(p.Network, id)
		if len(fans) > 0 && cell == nil {
			return fmt.Errorf("SCL: missing mapped cell %d", id)
		}
		for pin := range fans {
			var delay float32
			m := p.Network.SclMapping()
			if id < len(m.MioGates) && m.MioGates[id] != nil {
				delay = mio.Mio_GateReadPinDelay(m.MioGates[id], int32(pin))
			} else {
				value, err := bufferPinEstimate(cell, pin, 0, 0)
				if err != nil {
					return err
				}
				delay = value
			}

			if !sclFinite(delay) || delay < 0 {
				return fmt.Errorf("SCL: invalid gate delay")
			}
			delays[pin] = int(float32(BUF_SCALE) * delay / p.DelayInv)
		}
	}
	p.Edges[id] = delays
	return nil
}
func Abc_BufComputeArr(p *Buf_Man_t, id int) int {
	arrival := 0
	for pin, fanin := range p.Network.SclFanins(id) {
		arrival = max(arrival, p.Arrivals[fanin]+p.Edges[id][pin])
	}
	p.Arrivals[id] = arrival
	return arrival
}
func Abc_BufComputeDep(p *Buf_Man_t, id int) int {
	departure := 0
	for _, edge := range bufferEdges(p.Network, id) {
		departure = max(departure, p.Departures[edge.Node]+p.Edges[edge.Node][edge.Pin])
	}
	p.Departures[id] = departure
	return departure
}
func Abc_BufUpdateGlobal(p *Buf_Man_t) {
	p.DelayMax = 0
	for _, output := range p.Network.SclOutputs() {
		p.DelayMax = max(p.DelayMax, p.Arrivals[output])
	}
}
func (p *Buf_Man_t) recompute() error {
	order, err := bufferOrder(p.Network)
	if err != nil {
		return err
	}
	for _, id := range order {
		if err = Abc_BufCreateEdges(p, id); err != nil {
			return err
		}
		Abc_BufComputeArr(p, id)
	}
	for i := len(order) - 1; i >= 0; i-- {
		Abc_BufComputeDep(p, order[i])
	}
	Abc_BufUpdateGlobal(p)
	return nil
}
func Buf_ManStart(net BufferingTimingNetwork, fanMin, fanMax int, bufferPIs bool) (*Buf_Man_t, error) {
	if fanMin < 2 || fanMax < 2 {
		return nil, fmt.Errorf("SCL: fanout limits must be at least two")
	}
	inv, err := bufferUnaryCell(net, true)
	if err != nil {
		return nil, err
	}
	var delay float32
	if m := net.SclMapping(); m.Mio != nil && mio.Mio_LibraryReadInv(m.Mio) != nil {
		delay = mio.Mio_GateReadPinDelay(mio.Mio_LibraryReadInv(m.Mio), 0)
	} else {
		delay, err = bufferPinEstimate(inv, 0, 0, 0)
		if err != nil {
			return nil, err
		}
	}
	if delay <= 0 || !sclFinite(delay) {
		return nil, fmt.Errorf("SCL: invalid inverter reference delay")
	}
	p := &Buf_Man_t{Network: net, FanMin: fanMin, FanMax: fanMax, BufferPIs: bufferPIs, Arrivals: make(map[int]int), Departures: make(map[int]int), Edges: make(map[int][]int), DelayInv: delay, Output: os.Stdout}
	if err = p.recompute(); err != nil {
		return nil, err
	}
	return p, nil
}
func Buf_ManStop(p *Buf_Man_t) {
	if p != nil {
		p.Arrivals = nil
		p.Departures = nil
		p.Edges = nil
		p.NonCritical = nil
	}
}
func Abc_BufAddToQue(p *Buf_Man_t, id int) bool {
	return len(bufferEdges(p.Network, id)) >= p.FanMin && (p.BufferPIs || !bufferInputs(p.Network)[id])
}
func Abc_BufCollectTfoCone_rec(net TimingNetwork, id int, result *[]int, seen map[int]bool) {
	if seen[id] {
		return
	}
	seen[id] = true
	for _, fanout := range net.SclFanouts(id) {
		Abc_BufCollectTfoCone_rec(net, fanout, result, seen)
	}
	if net.SclIsNode(id) {
		*result = append(*result, id)
	}
}
func Abc_BufCollectTfoCone(p *Buf_Man_t, id int) []int {
	var result []int
	Abc_BufCollectTfoCone_rec(p.Network, id, &result, make(map[int]bool))
	return result
}
func Abc_BufCollectTfiCone_rec(net TimingNetwork, id int, result *[]int, seen map[int]bool) {
	if seen[id] {
		return
	}
	seen[id] = true
	for _, fanin := range net.SclFanins(id) {
		Abc_BufCollectTfiCone_rec(net, fanin, result, seen)
	}
	if net.SclIsNode(id) {
		*result = append(*result, id)
	}
}
func Abc_BufCollectTfiCone(p *Buf_Man_t, id int) []int {
	var result []int
	Abc_BufCollectTfiCone_rec(p.Network, id, &result, make(map[int]bool))
	return result
}
func Abc_BufUpdateArr(p *Buf_Man_t, id int) error { return p.recompute() }
func Abc_BufUpdateDep(p *Buf_Man_t, id int) error { return p.recompute() }
func Abc_BufSortByDelay(p *Buf_Man_t, id int) []FanoutPin {
	edges := bufferEdges(p.Network, id)
	sort.SliceStable(edges, func(i, j int) bool {
		a, b := Abc_BufEdgeSlack(p, id, edges[i]), Abc_BufEdgeSlack(p, id, edges[j])
		if a != b {
			return a < b
		}
		if edges[i].Node != edges[j].Node {
			return edges[i].Node < edges[j].Node
		}
		return edges[i].Pin < edges[j].Pin
	})
	return edges
}
func Abc_BufComputeAverage(p *Buf_Man_t, id int, order []FanoutPin) int {
	if len(order) == 0 {
		return 0
	}
	sum := 0
	for _, edge := range order {
		sum += Abc_BufEdgeSlack(p, id, edge)
	}
	return sum / len(order)
}
func Abc_BufPrintOne(p *Buf_Man_t, id int) {
	fmt.Fprintf(p.Output, "Node %d arrival=%d departure=%d fanouts=%v\n", id, p.Arrivals[id], p.Departures[id], Abc_BufSortByDelay(p, id))
}
func Abc_BufFindNonBuffDriver(p *Buf_Man_t, id int) int {
	for Abc_SclObjIsBufInv(p.Network, id) && !Abc_SclIsInv(p.Network, id) {
		id = p.Network.SclFanins(id)[0]
	}
	return id
}
func Abc_BufCountNonCritical(p *Buf_Man_t, id int) int {
	p.NonCritical = nil
	for _, edge := range bufferEdges(p.Network, id) {
		if Abc_BufEdgeSlack(p, id, edge) > 7*BUF_SCALE/2 {
			p.NonCritical = append(p.NonCritical, edge)
		}
	}
	return len(p.NonCritical)
}
func Abc_BufPerformOne(p *Buf_Man_t, id int, skipDup, verbose bool) (bool, error) {
	net := p.Network
	edges := bufferEdges(net, id)
	noncritical := Abc_BufCountNonCritical(p, id)
	critical := len(edges) - noncritical
	if critical > 0 && noncritical > 1 {
		cell, err := bufferUnaryCell(net, false)
		if err != nil {
			return false, err
		}
		buffer, err := createBufferNode(net, []int{id}, cell)
		if err != nil {
			return false, err
		}
		for _, edge := range p.NonCritical {
			if err = net.SclPatchFanin(edge.Node, edge.Pin, buffer); err != nil {
				return false, err
			}
		}
		p.Separate++
	} else if !skipDup && critical > 0 && net.SclIsNode(id) && !net.SclIsBarrier(id) && len(edges) > p.FanMin {
		clone, err := cloneBufferNode(net, id)
		if err != nil {
			return false, err
		}
		for _, edge := range edges[:len(edges)/2] {
			if err = net.SclPatchFanin(edge.Node, edge.Pin, clone); err != nil {
				return false, err
			}
		}
		p.Duplicate++
	} else if critical > 0 && len(edges) > 8 || len(edges) > p.FanMax {
		degree := min(10, int(math.Sqrt(float64(len(edges)))))
		if degree < 2 {
			return false, nil
		}
		invert := Abc_SclObjIsBufInv(net, id) && !Abc_SclIsInv(net, id)
		cell, err := bufferUnaryCell(net, invert)
		if err != nil {
			return false, err
		}
		if invert {
			if err = assignBufferCell(net, id, cell); err != nil {
				return false, err
			}
			p.BranchInverters++
		} else {
			p.BranchBuffers++
		}
		buffers := make([]int, degree)
		for i := range buffers {
			buffers[i], err = createBufferNode(net, []int{id}, cell)
			if err != nil {
				return false, err
			}
		}
		large := len(edges)/degree + 1
		largeCount := len(edges) % degree
		boundary := large * largeCount
		for i, edge := range edges {
			group := 0
			if i < boundary {
				group = i / large
			} else {
				group = largeCount + (i-boundary)/(large-1)
			}
			if err = net.SclPatchFanin(edge.Node, edge.Pin, buffers[group]); err != nil {
				return false, err
			}
		}
		if critical > 0 {
			p.CriticalBranches++
		}
	} else {
		return false, nil
	}
	if err := p.recompute(); err != nil {
		return false, err
	}
	if verbose {
		Abc_BufPrintOne(p, id)
	}
	return true, nil
}
func Abc_BufReplaceBufsByInvs(net BufferingTimingNetwork) error {
	order, err := bufferOrder(net)
	if err != nil {
		return err
	}
	cell, err := bufferUnaryCell(net, true)
	if err != nil {
		return err
	}
	for _, id := range order {
		if !Abc_SclObjIsBufInv(net, id) || Abc_SclIsInv(net, id) {
			continue
		}
		driver := net.SclFanins(id)[0]
		inverse, err := createBufferNode(net, []int{driver}, cell)
		if err != nil {
			return err
		}
		if err = net.SclPatchFanin(id, 0, inverse); err != nil {
			return err
		}
		if err = assignBufferCell(net, id, cell); err != nil {
			return err
		}
	}
	return nil
}
func Abc_SclBufPerform(network TimingNetwork, fanMin, fanMax int, bufferPIs, skipDup, verbose bool) (TimingNetwork, error) {
	net, err := cloneBufferNetwork(network)
	if err != nil {
		return nil, err
	}
	p, err := Buf_ManStart(net, fanMin, fanMax, bufferPIs)
	if err != nil {
		return nil, err
	}
	defer Buf_ManStop(p)
	growthLimit := 2*len(net.SclObjectIDs()) + 100
	queue := &bufferQueue{}
	versions := make(map[int]int)
	enqueue := func(id int) {
		versions[id]++
		if Abc_BufAddToQue(p, id) {
			heap.Push(queue, bufferQueueItem{id: id, fanouts: len(bufferEdges(net, id)), version: versions[id]})
		}
	}
	for _, id := range net.SclObjectIDs() {
		enqueue(id)
	}
	for queue.Len() > 0 {
		item := heap.Pop(queue).(bufferQueueItem)
		if item.version != versions[item.id] || !Abc_BufAddToQue(p, item.id) {
			continue
		}
		pivot := item.id
		oldIDs := make(map[int]bool)
		for _, id := range net.SclObjectIDs() {
			oldIDs[id] = true
		}
		changed, err := Abc_BufPerformOne(p, pivot, skipDup, verbose)
		if err != nil {
			return nil, err
		}
		if !changed {
			continue
		}
		if len(net.SclObjectIDs()) > growthLimit {
			return nil, fmt.Errorf("SCL: critical buffering exceeds original growth capacity")
		}
		enqueue(pivot)
		for _, fanin := range net.SclFanins(pivot) {
			enqueue(fanin)
		}
		for _, id := range net.SclObjectIDs() {
			if !oldIDs[id] {
				enqueue(id)
			}
		}
	}
	return finishBufferNetwork(net)
}

type bufferQueueItem struct{ id, fanouts, version int }
type bufferQueue []bufferQueueItem

func (q bufferQueue) Len() int { return len(q) }
func (q bufferQueue) Less(i, j int) bool {
	if q[i].fanouts != q[j].fanouts {
		return q[i].fanouts > q[j].fanouts
	}
	return q[i].id < q[j].id
}
func (q bufferQueue) Swap(i, j int)   { q[i], q[j] = q[j], q[i] }
func (q *bufferQueue) Push(value any) { *q = append(*q, value.(bufferQueueItem)) }
func (q *bufferQueue) Pop() any {
	last := len(*q) - 1
	value := (*q)[last]
	*q = (*q)[:last]
	return value
}
