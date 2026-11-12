package scl

// Load/gain-driven buffering and sizing from ABC's sclBufSize.c.
import (
	"fmt"
	"io"
	"os"
	"sort"
)

type busFanout struct {
	Edge           FanoutPin
	Cin, Departure float32
}
type Bus_Man_t struct {
	Network               BufferingTimingNetwork
	Library               *SC_Lib
	Parameters            SC_BusPars
	Inverter, InputDriver *SC_Cell
	WireCaps              []float32
	Loads, Departures     map[int]float32
	FanoutInfo            map[FanoutPin]busFanout
	OutputLoads           map[int]float32
	Output                io.Writer
	Delay                 float32
}

func Bus_ManStart(net BufferingTimingNetwork, lib *SC_Lib, params *SC_BusPars) (*Bus_Man_t, error) {
	if params == nil || params.Degree < 2 || params.GainRatio <= 0 || params.Slew < 0 {
		return nil, fmt.Errorf("SCL: invalid buffer sizing parameters")
	}
	if err := ensureSclMapping(lib, net); err != nil {
		return nil, err
	}
	p := &Bus_Man_t{Network: net, Library: lib, Parameters: *params, Loads: make(map[int]float32), Departures: make(map[int]float32), FanoutInfo: make(map[FanoutPin]busFanout), OutputLoads: make(map[int]float32), Output: os.Stdout}
	cell, err := bufferUnaryCell(net, !params.AddBuffers)
	if err != nil {
		return nil, err
	}
	p.Inverter = cell
	if cell.Prev != nil {
		p.Inverter = cell.Prev
	}
	if !params.AddBuffers && !params.SizeOnly {
		phases, ok := net.(ReadablePhases)
		if !ok || !phases.SclHasPhases() {
			return nil, fmt.Errorf("SCL: inverter-based sizing requires phase metadata")
		}
	}
	if params.UseWireLoads {
		model := Abc_SclFindWireLoadModel(lib, Abc_SclGetTotalArea(net))
		if model != nil {
			maximum := 0
			for _, id := range net.SclObjectIDs() {
				maximum = max(maximum, len(bufferEdges(net, id)))
			}
			p.WireCaps, err = Abc_SclFindWireCaps(model, maximum)
			if err != nil {
				return nil, err
			}
		}
	}
	if err = Bus_ManReadInOutLoads(p); err != nil {
		return nil, err
	}
	return p, nil
}
func Bus_ManStop(p *Bus_Man_t) {
	if p != nil {
		p.Loads = nil
		p.Departures = nil
		p.FanoutInfo = nil
		p.WireCaps = nil
		p.OutputLoads = nil
	}
}
func Bus_ManReadInOutLoads(p *Bus_Man_t) error {
	timing, err := NewTimingManager(p.Library, p.Network)
	if err != nil {
		return err
	}
	defer Abc_SclManFree(timing)
	if err = Abc_SclManReadSlewAndLoad(timing, p.Network); err != nil {
		return err
	}
	for _, id := range p.Network.SclOutputs() {
		p.OutputLoads[id] = SC_PairAve(&timing.Loads[id])
	}
	p.InputDriver = timing.InputDriver
	if p.InputDriver != nil {
		if _, err := bufferPinEstimate(p.InputDriver, 0, 0, 0); err != nil {
			return err
		}
	}
	return nil
}

func Bus_SclObjCin(p *Bus_Man_t, edge FanoutPin) float32 { return p.FanoutInfo[edge].Cin }
func Bus_SclObjSetCin(p *Bus_Man_t, edge FanoutPin, value float32) {
	info := p.FanoutInfo[edge]
	info.Edge = edge
	info.Cin = value
	p.FanoutInfo[edge] = info
}
func Bus_SclObjETime(p *Bus_Man_t, edge FanoutPin) float32 { return p.FanoutInfo[edge].Departure }
func Bus_SclObjSetETime(p *Bus_Man_t, edge FanoutPin, value float32) {
	info := p.FanoutInfo[edge]
	info.Edge = edge
	info.Departure = value
	p.FanoutInfo[edge] = info
}
func Bus_SclObjLoad(p *Bus_Man_t, id int) float32           { return p.Loads[id] }
func Bus_SclObjSetLoad(p *Bus_Man_t, id int, value float32) { p.Loads[id] = value }
func Bus_SclObjDept(p *Bus_Man_t, id int) float32           { return p.Departures[id] }
func Bus_SclObjUpdateDept(p *Bus_Man_t, id int, value float32) {
	p.Departures[id] = max(p.Departures[id], value)
}
func Abc_NtkComputeEdgeDept(p *Bus_Man_t, edge FanoutPin, slew float32) (float32, error) {
	if bufferOutputs(p.Network)[edge.Node] {
		return p.Departures[edge.Node], nil
	}
	if p.Network.SclIsBarrier(edge.Node) {
		return p.Departures[edge.Node], nil
	}
	cell := bufferCell(p.Network, edge.Node)
	if cell == nil || edge.Pin < 0 || edge.Pin >= cell.Inputs {
		return 0, fmt.Errorf("SCL: invalid timed fanout edge")
	}
	delay, err := bufferPinEstimate(cell, edge.Pin, slew, p.Loads[edge.Node])
	if err != nil {
		return 0, err
	}
	if !sclFinite(delay) || delay < 0 {
		return 0, fmt.Errorf("SCL: invalid edge delay")
	}
	return p.Departures[edge.Node] + delay, nil
}
func Abc_NtkComputeNodeDeparture(p *Bus_Man_t, id int, slew float32) (float32, error) {
	departure := float32(0)
	for _, edge := range bufferEdges(p.Network, id) {
		value, err := Abc_NtkComputeEdgeDept(p, edge, slew)
		if err != nil {
			return 0, err
		}
		departure = max(departure, value)
	}
	p.Departures[id] = departure
	return departure, nil
}
func Abc_NtkComputeFanoutInfo(p *Bus_Man_t, id int, slew float32) error {
	outputs := bufferOutputs(p.Network)
	for _, edge := range bufferEdges(p.Network, id) {
		info := busFanout{Edge: edge}
		if outputs[edge.Node] {
			info.Cin = p.OutputLoads[edge.Node]
			info.Departure = p.Departures[edge.Node]
		} else if p.Network.SclIsBarrier(edge.Node) {
			info.Cin = p.Loads[edge.Node]
			info.Departure = p.Departures[edge.Node]
		} else {
			cell := bufferCell(p.Network, edge.Node)
			if cell == nil || edge.Pin >= cell.Inputs {
				return fmt.Errorf("SCL: missing fanout cell")
			}
			info.Cin = SC_CellPinCap(cell, edge.Pin)
			value, err := Abc_NtkComputeEdgeDept(p, edge, slew)
			if err != nil {
				return err
			}
			info.Departure = value
		}
		if !sclFinite(info.Cin) || info.Cin < 0 {
			return fmt.Errorf("SCL: invalid fanout capacitance")
		}
		p.FanoutInfo[edge] = info
	}
	return nil
}
func Abc_NtkComputeFanoutLoad(p *Bus_Man_t, edges []FanoutPin) float32 {
	load := Abc_SclFindWireLoad(p.WireCaps, len(edges))
	for _, edge := range edges {
		load += p.FanoutInfo[edge].Cin
	}
	return load
}
func Abc_NtkComputeNodeLoad(p *Bus_Man_t, id int) float32 {
	load := Abc_NtkComputeFanoutLoad(p, bufferEdges(p.Network, id))
	p.Loads[id] = load
	return load
}
func Bus_SclCompareFanouts(p *Bus_Man_t, a, b FanoutPin) int {
	x, y := p.FanoutInfo[a], p.FanoutInfo[b]
	if x.Departure < y.Departure {
		return -1
	}
	if x.Departure > y.Departure {
		return 1
	}
	if x.Cin > y.Cin {
		return -1
	}
	if x.Cin < y.Cin {
		return 1
	}
	if a.Node < b.Node {
		return -1
	}
	if a.Node > b.Node {
		return 1
	}
	if a.Pin < b.Pin {
		return -1
	}
	if a.Pin > b.Pin {
		return 1
	}
	return 0
}
func Bus_SclInsertFanout(p *Bus_Man_t, edges []FanoutPin, edge FanoutPin) []FanoutPin {
	index := sort.Search(len(edges), func(i int) bool { return Bus_SclCompareFanouts(p, edges[i], edge) > 0 })
	edges = append(edges, FanoutPin{})
	copy(edges[index+1:], edges[index:])
	edges[index] = edge
	return edges
}
func Bus_SclCheckSortedFanout(p *Bus_Man_t, edges []FanoutPin) bool {
	for i := 1; i < len(edges); i++ {
		if Bus_SclCompareFanouts(p, edges[i-1], edges[i]) > 0 {
			return false
		}
	}
	return true
}
func Abc_NtkPrintFanoutProfile(p *Bus_Man_t, id int) {
	Abc_NtkPrintFanoutProfileVec(p, bufferEdges(p.Network, id))
}
func Abc_NtkPrintFanoutProfileVec(p *Bus_Man_t, edges []FanoutPin) {
	for _, edge := range edges {
		info := p.FanoutInfo[edge]
		fmt.Fprintf(p.Output, "%d:%d departure=%.2f load=%.2f\n", edge.Node, edge.Pin, info.Departure, info.Cin)
	}
}
func Abc_SclOneNodePrint(p *Bus_Man_t, id int) {
	cell := bufferCell(p.Network, id)
	name := "PI"
	if cell != nil {
		name = cell.Name
	}
	fmt.Fprintf(p.Output, "Node %d %s load=%.2f departure=%.2f\n", id, name, p.Loads[id], p.Departures[id])
}
func smallestBufferGate(cell *SC_Cell, minimum float32) *SC_Cell {
	root := cell.Repr
	if root == nil {
		root = cell
	}
	last := root
	for current := root; current != nil; current = current.Next {
		last = current
		if SC_CellPinCapAve(current) > minimum {
			return current
		}
		if current.Next == root {
			break
		}
	}
	return last
}

// AddOneInv consumes a low-departure prefix of fanouts and replaces it with
// one appropriately sized unary stage. At least two edges are grouped.
func Abc_SclAddOneInv(p *Bus_Man_t, driver int, fanouts []FanoutPin, gain float32) (int, []FanoutPin, error) {
	if len(fanouts) < 2 || gain <= 0 || !sclFinite(gain) {
		return -1, nil, fmt.Errorf("SCL: invalid fanout grouping")
	}
	limit := min(p.Parameters.Degree, len(fanouts))
	target := SC_CellPinCap(p.Inverter, 0) * gain
	load := float32(0)
	stop := 0
	for stop < limit {
		load += p.FanoutInfo[fanouts[stop]].Cin + Abc_SclFindWireLoad(p.WireCaps, stop+1) - Abc_SclFindWireLoad(p.WireCaps, stop)
		stop++
		if load > target {
			break
		}
	}
	stop = min(max(stop, 2), len(fanouts))
	load = Abc_NtkComputeFanoutLoad(p, fanouts[:stop])
	cell := smallestBufferGate(p.Inverter, load/gain)
	buffer, err := createBufferNode(p.Network, []int{driver}, cell)
	if err != nil {
		return -1, nil, err
	}
	for _, edge := range fanouts[:stop] {
		if err = p.Network.SclPatchFanin(edge.Node, edge.Pin, buffer); err != nil {
			return -1, nil, err
		}
		if !p.Parameters.AddBuffers {
			if err = Abc_NodeInvUpdateObjFanoutPolarity(p.Network, buffer, edge); err != nil {
				return -1, nil, err
			}
		}
	}
	p.Loads[buffer] = load
	if _, err = Abc_NtkComputeNodeDeparture(p, buffer, float32(p.Parameters.Slew)); err != nil {
		return -1, nil, err
	}
	edge := FanoutPin{buffer, 0}
	departure, err := Abc_NtkComputeEdgeDept(p, edge, float32(p.Parameters.Slew))
	if err != nil {
		return -1, nil, err
	}
	p.FanoutInfo[edge] = busFanout{Edge: edge, Cin: SC_CellPinCap(cell, 0), Departure: departure}
	remaining := append([]FanoutPin(nil), fanouts[stop:]...)
	remaining = Bus_SclInsertFanout(p, remaining, edge)
	return buffer, remaining, nil
}
func Abc_SclBufSize(p *Bus_Man_t, gain float32) error {
	if gain <= 0 || !sclFinite(gain) {
		return fmt.Errorf("SCL: gain must be finite and positive")
	}
	gateGain := gain
	if p.Parameters.AddBuffers {
		gateGain *= gain
	}
	if !sclFinite(gateGain) {
		return fmt.Errorf("SCL: gain overflow")
	}
	order, err := bufferOrder(p.Network)
	if err != nil {
		return err
	}
	originalCount := len(order)
	inputs := bufferInputs(p.Network)
	for index := len(order) - 1; index >= 0; index-- {
		id := order[index]
		if !p.Network.SclIsNode(id) && !inputs[id] || p.Network.SclIsNode(id) && len(p.Network.SclFanins(id)) == 0 {
			continue
		}
		if len(p.Network.SclObjectIDs()) > 2*originalCount+1 {
			return fmt.Errorf("SCL: buffering exceeds original growth limit; gain is too low")
		}
		if err = Abc_NtkComputeFanoutInfo(p, id, float32(p.Parameters.Slew)); err != nil {
			return err
		}
		load := Abc_NtkComputeNodeLoad(p, id)
		cell := bufferCell(p.Network, id)
		cin := load
		if inputs[id] || p.Network.SclIsBarrier(id) {
			cell = p.InputDriver
			if cell != nil {
				cin = SC_CellPinCapAve(cell)
			}
		} else {
			if cell == nil {
				return fmt.Errorf("SCL: missing cell at object %d", id)
			}
			average := cell.Average
			if average == nil {
				average = cell
			}
			cin = SC_CellPinCapAve(average)
		}
		allowBuffer := !p.Parameters.SizeOnly && (!inputs[id] || p.Parameters.BufferPIs)
		fanouts := bufferEdges(p.Network, id)
		if allowBuffer && (len(fanouts) > p.Parameters.Degree || load > gateGain*cin) && len(fanouts) > 1 {
			sort.SliceStable(fanouts, func(i, j int) bool { return Bus_SclCompareFanouts(p, fanouts[i], fanouts[j]) < 0 })
			for len(fanouts) > p.Parameters.Degree || len(fanouts) > 1 && load > gateGain*cin {
				if len(p.Network.SclObjectIDs()) > 2*originalCount+1 {
					return fmt.Errorf("SCL: buffering exceeds original growth limit; gain is too low")
				}
				_, fanouts, err = Abc_SclAddOneInv(p, id, fanouts, gateGain)
				if err != nil {
					return err
				}
				load = Abc_NtkComputeFanoutLoad(p, fanouts)
			}
			p.Loads[id] = load
		}
		if inputs[id] {
			continue
		}
		if _, err = Abc_NtkComputeNodeDeparture(p, id, float32(p.Parameters.Slew)); err != nil {
			return err
		}
		if p.Network.SclIsBarrier(id) {
			continue
		}
		selected := smallestBufferGate(cell, load/gateGain)
		if err = assignBufferCell(p.Network, id, selected); err != nil {
			return err
		}
		if p.Parameters.VeryVerbose {
			Abc_SclOneNodePrint(p, id)
		}
	}
	for _, id := range p.Network.SclInputs() {
		departure, err := Abc_NtkComputeNodeDeparture(p, id, float32(p.Parameters.Slew))
		if err != nil {
			return err
		}
		if p.InputDriver != nil {
			load := SC_Pair{p.Loads[id], p.Loads[id]}
			arrival, slew := SC_Pair{}, SC_Pair{}
			Scl_LibHandleInputDriver(p.InputDriver, &load, &arrival, &slew)
			departure += SC_PairAve(&arrival)
		}
		p.Delay = max(p.Delay, departure)
	}
	if p.Parameters.Verbose {
		fmt.Fprintf(p.Output, "Buffering degree=%d gain=%d delay=%.2f\n", p.Parameters.Degree, p.Parameters.GainRatio, p.Delay)
	}
	return nil
}
func Abc_SclBufferingPerform(network TimingNetwork, lib *SC_Lib, params *SC_BusPars) (TimingNetwork, error) {
	net, err := cloneBufferNetwork(network)
	if err != nil {
		return nil, err
	}
	manager, err := Bus_ManStart(net, lib, params)
	if err != nil {
		return nil, err
	}
	defer Bus_ManStop(manager)
	if err = Abc_SclBufSize(manager, float32(params.GainRatio)*.01); err != nil {
		return nil, err
	}
	return finishBufferNetwork(net)
}
