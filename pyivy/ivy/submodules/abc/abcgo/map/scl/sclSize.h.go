// Native timing/sizing manager from sclSize.h. Topology remains owned by its
// network; this interface lets ABC graphs supply it without a reverse import.
package scl

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/glycerine/abcgo/map/mio"
)

type TimingNetwork interface {
	SclName() string
	SclObjectIDs() []int
	SclInputs() []int
	SclOutputs() []int
	SclFanins(int) []int
	SclFanouts(int) []int
	SclIsNode(int) bool
	SclIsBarrier(int) bool
	SclMapping() *NetworkMapping
}
type NetworkMapping struct {
	Lib      *SC_Lib
	Mio      *mio.Mio_Library_t
	GateIDs  []int32
	MioGates []*mio.Mio_Gate_t
}
type SC_Man struct {
	statsName                                                            string
	statsNodes                                                           int
	statsArea, statsDelay                                                float32
	statsStart                                                           time.Time
	Lib                                                                  *SC_Lib
	Network                                                              TimingNetwork
	Mapping                                                              *NetworkMapping
	Loads, Departures, Times, Slews                                      []SC_Pair
	WireLoad                                                             *SC_WireLoad
	WireCaps                                                             []float32
	InputDrive                                                           []float32
	OutputTimes                                                          []float32
	InputDriver                                                          *SC_Cell
	BestGates, Updates, Updates2                                         []int32
	Loads2, Loads3                                                       []SC_Pair
	Times2, Slews2, Times3                                               []SC_Pair
	NodeGain                                                             []float32
	NodeGate, NodeIteration, BestFanins                                  []int32
	Changed                                                              []int
	Levels                                                               []int
	SumArea, MaxDelay, InitialArea, InitialDelay, BestDelay, ReportDelay float32
	EstLoadMax, EstLoadAve, EstLinear                                    float32
	EstimatedNodes, IncrementalUpdates                                   int
	TimeTotal, TimeCone, TimeSize, TimeTiming, TimeOther                 time.Duration
	Output                                                               io.Writer
	order                                                                []int
	inputIndex, outputIndex                                              map[int]int
}

func NewTimingManager(lib *SC_Lib, network TimingNetwork) (*SC_Man, error) {
	if lib == nil || network == nil || network.SclMapping() == nil {
		return nil, fmt.Errorf("SCL: timing requires a library and mapped network")
	}
	if lib.UnitTime != 12 || lib.UnitCapExponent != 15 || lib.UnitCapMultiplier != 1 {
		return nil, fmt.Errorf("SCL: timing library must use normalized ps/fF units")
	}
	maxID := -1
	for _, id := range network.SclObjectIDs() {
		if id < 0 {
			return nil, fmt.Errorf("SCL: negative object ID")
		}
		maxID = max(maxID, id)
	}
	if err := ensureSclMapping(lib, network); err != nil {
		return nil, err
	}
	p := &SC_Man{Lib: lib, Network: network, Mapping: network.SclMapping(), Output: os.Stdout, inputIndex: make(map[int]int), outputIndex: make(map[int]int)}
	p.ensureCapacity(maxID + 1)
	for i, id := range network.SclInputs() {
		p.inputIndex[id] = i
	}
	for i, id := range network.SclOutputs() {
		p.outputIndex[id] = i
	}
	p.OutputTimes = make([]float32, len(p.outputIndex))
	p.InputDrive = make([]float32, len(p.inputIndex))
	if err := p.rebuildOrder(); err != nil {
		return nil, err
	}
	return p, nil
}
func (p *SC_Man) ensureCapacity(n int) {
	growPair := func(a []SC_Pair) []SC_Pair {
		if len(a) < n {
			return append(a, make([]SC_Pair, n-len(a))...)
		}
		return a
	}
	p.Loads = growPair(p.Loads)
	p.Departures = growPair(p.Departures)
	p.Times = growPair(p.Times)
	p.Slews = growPair(p.Slews)
	if len(p.Levels) < n {
		p.Levels = append(p.Levels, make([]int, n-len(p.Levels))...)
	}
	if len(p.NodeGain) < n {
		p.NodeGain = append(p.NodeGain, make([]float32, n-len(p.NodeGain))...)
	}
	growInts := func(a []int32, fill int32) []int32 {
		for len(a) < n {
			a = append(a, fill)
		}
		return a
	}
	p.NodeGate = growInts(p.NodeGate, -1)
	p.NodeIteration = growInts(p.NodeIteration, -1)
	p.BestFanins = growInts(p.BestFanins, 0)
}
func Abc_SclManAlloc(lib *SC_Lib, network TimingNetwork) (*SC_Man, error) {
	return NewTimingManager(lib, network)
}
func Abc_SclManFree(p *SC_Man) {
	if p != nil {
		*p = SC_Man{}
	}
}
func Abc_SclObjLib(p *SC_Man) *SC_Lib { return p.Lib }
func Abc_SclObjCellId(p *SC_Man, id int) int32 {
	if id < 0 || id >= len(p.Mapping.GateIDs) {
		return -1
	}
	return p.Mapping.GateIDs[id]
}
func Abc_SclObjCell(p *SC_Man, id int) *SC_Cell {
	cell := Abc_SclObjCellId(p, id)
	if cell < 0 || int(cell) >= len(p.Lib.Cells) {
		return nil
	}
	return p.Lib.Cells[cell]
}
func Abc_SclObjSetCell(p *SC_Man, id int, cell *SC_Cell) {
	if cell == nil || cell.ID < 0 || cell.ID >= len(p.Lib.Cells) || p.Lib.Cells[cell.ID] != cell {
		panic("SCL: foreign cell assignment")
	}
	p.Mapping.GateIDs[id] = int32(cell.ID)
}
func Abc_SclObjLoad(p *SC_Man, id int) *SC_Pair   { return &p.Loads[id] }
func Abc_SclObjDept(p *SC_Man, id int) *SC_Pair   { return &p.Departures[id] }
func Abc_SclObjTime(p *SC_Man, id int) *SC_Pair   { return &p.Times[id] }
func Abc_SclObjSlew(p *SC_Man, id int) *SC_Pair   { return &p.Slews[id] }
func Abc_SclObjLoadMax(p *SC_Man, id int) float64 { return float64(SC_PairMax(&p.Loads[id])) }
func Abc_SclObjLoadAve(p *SC_Man, id int) float32 { return SC_PairAve(&p.Loads[id]) }
func Abc_SclObjTimeOne(p *SC_Man, id int, rise int32) float64 {
	if rise != 0 {
		return float64(p.Times[id].Rise)
	}
	return float64(p.Times[id].Fall)
}
func Abc_SclObjTimeMax(p *SC_Man, id int) float32 { return SC_PairMax(&p.Times[id]) }
func Abc_SclObjSlewMax(p *SC_Man, id int) float64 { return float64(SC_PairMax(&p.Slews[id])) }
func Abc_SclObjGetSlackR(p *SC_Man, id int, d float32) float32 {
	return d - p.Times[id].Rise - p.Departures[id].Rise
}
func Abc_SclObjGetSlackF(p *SC_Man, id int, d float32) float32 {
	return d - p.Times[id].Fall - p.Departures[id].Fall
}
func Abc_SclObjGetSlack(p *SC_Man, id int, d float32) float32 {
	return min(Abc_SclObjGetSlackR(p, id, d), Abc_SclObjGetSlackF(p, id, d))
}
func Abc_SclObjSlackMax(p *SC_Man, id int, d float32) float64 {
	return float64(Abc_SclObjGetSlack(p, id, d))
}
func Abc_SclObjDupFanin(p *SC_Man, id int) {
	fans := p.Network.SclFanins(id)
	if len(fans) != 1 {
		panic("SCL: output requires one driver")
	}
	p.Times[id] = p.Times[fans[0]]
}
func Abc_SclObjInDrive(p *SC_Man, id int) float32       { return p.InputDrive[p.inputIndex[id]] }
func Abc_SclObjSetInDrive(p *SC_Man, id int, c float32) { p.InputDrive[p.inputIndex[id]] = c }
func Abc_SclManCleanTime(p *SC_Man)                     { clear(p.Times); clear(p.Slews); clear(p.Departures) }
func Abc_SclLoadStore(p *SC_Man, id int) {
	p.Loads2 = p.Loads2[:0]
	for _, fanin := range p.Network.SclFanins(id) {
		p.Loads2 = append(p.Loads2, p.Loads[fanin])
	}
}
func Abc_SclLoadRestore(p *SC_Man, id int) {
	fans := p.Network.SclFanins(id)
	if len(fans) != len(p.Loads2) {
		panic("SCL: load backup shape changed")
	}
	for i, fanin := range fans {
		p.Loads[fanin] = p.Loads2[i]
	}
}
func Abc_SclLoadStore3(p *SC_Man, id int) {
	p.Loads3 = append(p.Loads3[:0], p.Loads[id])
	for _, fanin := range p.Network.SclFanins(id) {
		p.Loads3 = append(p.Loads3, p.Loads[fanin])
	}
}
func Abc_SclLoadRestore3(p *SC_Man, id int) {
	fans := p.Network.SclFanins(id)
	if len(fans)+1 != len(p.Loads3) {
		panic("SCL: load backup shape changed")
	}
	p.Loads[id] = p.Loads3[0]
	for i, fanin := range fans {
		p.Loads[fanin] = p.Loads3[i+1]
	}
}
func Abc_SclConeStore(p *SC_Man, cone []int) {
	p.Times2 = p.Times2[:0]
	p.Slews2 = p.Slews2[:0]
	for _, id := range cone {
		p.Times2 = append(p.Times2, p.Times[id])
		p.Slews2 = append(p.Slews2, p.Slews[id])
	}
}
func Abc_SclConeRestore(p *SC_Man, cone []int) {
	if len(cone) != len(p.Times2) || len(cone) != len(p.Slews2) {
		panic("SCL: cone backup shape changed")
	}
	for i, id := range cone {
		p.Times[id], p.Slews[id] = p.Times2[i], p.Slews2[i]
	}
}
func Abc_SclEvalStore(p *SC_Man, cone []int) {
	p.Times3 = p.Times3[:0]
	for _, id := range cone {
		p.Times3 = append(p.Times3, p.Times[id])
	}
}
func Abc_SclEvalPerform(p *SC_Man, cone []int) float32 {
	if len(cone) == 0 {
		return 0
	}
	if len(cone) != len(p.Times3) {
		panic("SCL: evaluation backup shape changed")
	}
	var evaluation float32
	for i, id := range cone {
		diff := p.Times3[i].Rise + p.Times3[i].Fall - p.Times[id].Rise - p.Times[id].Fall
		if diff < 0 {
			diff *= 1.5
		}
		evaluation += diff / 2
	}
	return evaluation / float32(len(cone))
}
func Abc_SclEvalPerformLegal(p *SC_Man, cone []int, d float32) float32 {
	if len(cone) == 0 {
		return 0
	}
	if len(cone) != len(p.Times3) {
		panic("SCL: evaluation backup shape changed")
	}
	var evaluation float32
	for i, id := range cone {
		rise, fall := p.Times3[i].Rise-p.Times[id].Rise, p.Times3[i].Fall-p.Times[id].Fall
		if rise+Abc_SclObjGetSlackR(p, id, d) < 0 || fall+Abc_SclObjGetSlackF(p, id, d) < 0 {
			return -1
		}
		evaluation += (rise + fall) / 2
	}
	return evaluation / float32(len(cone))
}
func Abc_SclConeClean(p *SC_Man, cone []int) {
	for _, id := range cone {
		p.Times[id] = SC_Pair{}
		p.Slews[id] = SC_Pair{}
	}
}
func Abc_SclGetMaxDelay(p *SC_Man) float32 {
	var delay float32
	for _, id := range p.Network.SclOutputs() {
		delay = max(delay, SC_PairMax(&p.Times[id]))
	}
	return delay
}
func Abc_SclReadMaxDelay(p *SC_Man) float32 { return Abc_SclGetMaxDelay(p) }
func Abc_SclGetMaxDelayNodeFanins(p *SC_Man, id int) float32 {
	var delay float32
	for _, fanin := range p.Network.SclFanins(id) {
		delay = max(delay, SC_PairMax(&p.Times[fanin]))
	}
	return delay
}
func Abc_SclObjResiable(p *SC_Man, id int, up int32) *SC_Cell {
	old := Abc_SclObjCell(p, id)
	if old == nil {
		return nil
	}
	if up != 0 {
		if old.Next != nil && old.Next.Order > old.Order {
			return old.Next
		}
	} else if old.Prev != nil && old.Prev.Order < old.Order {
		return old.Prev
	}
	return nil
}
