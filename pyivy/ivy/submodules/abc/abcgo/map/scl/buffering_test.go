package scl

import (
	"bytes"
	"encoding/json"
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"os"
	"reflect"
	"sort"
	"testing"
)

type bufferingTestNode struct {
	fanins        []int
	node, barrier bool
}
type bufferingTestNetwork struct {
	name            string
	nodes           map[int]bufferingTestNode
	inputs, outputs []int
	mapping         NetworkMapping
	phases          map[FanoutPin]bool
	enabled         bool
}

func (n *bufferingTestNetwork) SclName() string { return n.name }
func (n *bufferingTestNetwork) SclObjectIDs() []int {
	ids := make([]int, 0, len(n.nodes))
	for id := range n.nodes {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
func (n *bufferingTestNetwork) SclInputs() []int       { return n.inputs }
func (n *bufferingTestNetwork) SclOutputs() []int      { return n.outputs }
func (n *bufferingTestNetwork) SclFanins(id int) []int { return n.nodes[id].fanins }
func (n *bufferingTestNetwork) SclFanouts(id int) []int {
	var result []int
	for _, other := range n.SclObjectIDs() {
		for _, fanin := range n.nodes[other].fanins {
			if fanin == id {
				result = append(result, other)
			}
		}
	}
	return result
}
func (n *bufferingTestNetwork) SclIsNode(id int) bool       { return n.nodes[id].node }
func (n *bufferingTestNetwork) SclIsBarrier(id int) bool    { return n.nodes[id].barrier }
func (n *bufferingTestNetwork) SclMapping() *NetworkMapping { return &n.mapping }
func (n *bufferingTestNetwork) SclHasPhases() bool          { return n.enabled }
func (n *bufferingTestNetwork) SclFaninPhase(id, pin int) bool {
	return n.enabled && n.phases[FanoutPin{id, pin}]
}
func (n *bufferingTestNetwork) SclToggleFaninPhase(id, pin int) error {
	if !n.enabled {
		return fmt.Errorf("disabled phases")
	}
	if pin < 0 || pin >= len(n.nodes[id].fanins) {
		return fmt.Errorf("invalid phase pin")
	}
	edge := FanoutPin{id, pin}
	n.phases[edge] = !n.phases[edge]
	return nil
}
func (n *bufferingTestNetwork) SclSetPhases(enabled bool) error {
	n.enabled = enabled
	if !enabled {
		clear(n.phases)
	}
	return nil
}
func (n *bufferingTestNetwork) SclPatchFanin(id, pin, newID int) error {
	node, ok := n.nodes[id]
	if !ok || pin < 0 || pin >= len(node.fanins) {
		return fmt.Errorf("invalid pin")
	}
	if _, ok = n.nodes[newID]; !ok {
		return fmt.Errorf("missing new fanin")
	}
	node.fanins[pin] = newID
	n.nodes[id] = node
	return nil
}
func (n *bufferingTestNetwork) SclDeleteNode(id int) error {
	if len(n.SclFanouts(id)) != 0 {
		return fmt.Errorf("node %d still used", id)
	}
	delete(n.nodes, id)
	return nil
}
func (n *bufferingTestNetwork) SclCreateNode(fanins []int, gate int32) (int, error) {
	id := 0
	for current := range n.nodes {
		id = max(id, current+1)
	}
	for _, fanin := range fanins {
		if _, ok := n.nodes[fanin]; !ok {
			return -1, fmt.Errorf("missing fanin")
		}
	}
	n.nodes[id] = bufferingTestNode{fanins: append([]int(nil), fanins...), node: true}
	for len(n.mapping.GateIDs) <= id {
		n.mapping.GateIDs = append(n.mapping.GateIDs, -1)
	}
	n.mapping.GateIDs[id] = gate
	return id, nil
}
func (n *bufferingTestNetwork) SclClone() (TimingNetwork, error) {
	copy := &bufferingTestNetwork{name: n.name, nodes: make(map[int]bufferingTestNode), inputs: append([]int(nil), n.inputs...), outputs: append([]int(nil), n.outputs...), mapping: n.mapping, phases: make(map[FanoutPin]bool), enabled: n.enabled}
	copy.mapping.GateIDs = append([]int32(nil), n.mapping.GateIDs...)
	copy.mapping.MioGates = append(copy.mapping.MioGates[:0:0], n.mapping.MioGates...)
	for id, node := range n.nodes {
		node.fanins = append([]int(nil), node.fanins...)
		copy.nodes[id] = node
	}
	for edge, value := range n.phases {
		copy.phases[edge] = value
	}
	return copy, nil
}
func bufferTestLibrary() *SC_Lib {
	lib := &SC_Lib{Name: "buffer-test", UnitTime: 12, UnitCapMultiplier: 1, UnitCapExponent: 15}
	for class, truth := range []uint64{0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0x8888888888888888} {
		for size := 0; size < 3; size++ {
			arity := 1
			if class == 2 {
				arity = 2
			}
			cell := &SC_Cell{Name: fmt.Sprintf("class%d_%d", class, size), ID: len(lib.Cells), Inputs: arity, Outputs: 1, Area: float32(size + 1)}
			timings := make([]*SC_Timings, arity)
			for pin := 0; pin < arity; pin++ {
				name := fmt.Sprint("A", pin)
				cap := float32((size + 1) * (pin + 1))
				cell.Pins = append(cell.Pins, &SC_Pin{Name: name, Direction: ScDirInput, RiseCap: cap, FallCap: cap})
				arc := &SC_Timing{RelatedPin: name, Sense: ScTSPos, CellRise: linearTimingSurface(1, .01, .05), CellFall: linearTimingSurface(1, .01, .05), RiseTrans: linearTimingSurface(.1, .01, .01), FallTrans: linearTimingSurface(.1, .01, .01)}
				if class == 1 {
					arc.Sense = ScTSNeg
				}
				timings[pin] = &SC_Timings{Name: name, Timings: []*SC_Timing{arc}}
			}
			function := "A0"
			if class == 1 {
				function = "!A0"
			}
			if class == 2 {
				function = "A0 & A1"
			}
			cell.Pins = append(cell.Pins, &SC_Pin{Name: "Y", Direction: ScDirOutput, Function: function, Truth: []uint64{truth}, Timings: timings})
			lib.Cells = append(lib.Cells, cell)
		}
	}
	Abc_SclHashCells(lib)
	Abc_SclLinkCells(lib)
	return lib
}
func bufferTestFixture(fanouts int) *bufferingTestNetwork {
	lib := bufferTestLibrary()
	net := &bufferingTestNetwork{name: "fanout", nodes: map[int]bufferingTestNode{0: {}, 3: {}, 5: {fanins: []int{0, 3}, node: true}}, inputs: []int{0, 3}, mapping: NetworkMapping{Lib: lib, GateIDs: make([]int32, 10+fanouts)}, phases: make(map[FanoutPin]bool)}
	for i := range net.mapping.GateIDs {
		net.mapping.GateIDs[i] = -1
	}
	net.mapping.GateIDs[5] = 6
	for i := 0; i < fanouts; i++ {
		id := 10 + i
		net.nodes[id] = bufferingTestNode{fanins: []int{5}}
		net.outputs = append(net.outputs, id)
	}
	return net
}
func evaluateBufferNetwork(t *testing.T, net TimingNetwork, pattern int) []bool {
	t.Helper()
	values := make(map[int]bool)
	for i, id := range net.SclInputs() {
		values[id] = pattern&(1<<uint(i)) != 0
	}
	order, err := bufferOrder(net)
	if err != nil {
		t.Fatal(err)
	}
	outputs := bufferOutputs(net)
	for _, id := range order {
		fans := net.SclFanins(id)
		if len(fans) == 0 {
			continue
		}
		index := 0
		for pin, fanin := range fans {
			value := values[fanin] != bufferPhase(net, id, pin)
			if value {
				index |= 1 << uint(pin)
			}
		}
		if outputs[id] || net.SclIsBarrier(id) {
			values[id] = index&1 != 0
		} else {
			cell := bufferCell(net, id)
			truth := SC_CellFunc(cell)
			values[id] = truth[index/64]&(1<<uint(index%64)) != 0
		}
	}
	result := make([]bool, len(net.SclOutputs()))
	for i, id := range net.SclOutputs() {
		result[i] = values[id]
	}
	return result
}
func requireBufferEquivalent(t *testing.T, before, after TimingNetwork) {
	t.Helper()
	for pattern := 0; pattern < 1<<uint(len(before.SclInputs())); pattern++ {
		a, b := evaluateBufferNetwork(t, before, pattern), evaluateBufferNetwork(t, after, pattern)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("buffering changed function input=%d before=%v after=%v", pattern, a, b)
		}
	}
}
func TestBufferDegreeAndInputOwnership(t *testing.T) {
	before := bufferTestFixture(13)
	snapshot := append([]int(nil), before.SclFanouts(5)...)
	after, err := Abc_SclPerformBuffering(before, 0, 3, false, false)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
	for _, id := range after.SclObjectIDs() {
		if len(bufferEdges(after, id)) > 3 {
			t.Fatalf("object %d exceeds degree: %d", id, len(bufferEdges(after, id)))
		}
	}
	if !reflect.DeepEqual(snapshot, before.SclFanouts(5)) || len(after.SclObjectIDs()) <= len(before.SclObjectIDs()) {
		t.Fatal("input mutated or no real buffers added")
	}
}
func TestInverterDegreeBufferingPreservesPhases(t *testing.T) {
	before := bufferTestFixture(11)
	before.enabled = true
	after, err := Abc_SclPerformBuffering(before, 0, 3, true, false)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
}
func TestBufferPhaseRoundTrip(t *testing.T) {
	before := bufferTestFixture(2)
	before.enabled = true
	before.phases[FanoutPin{5, 0}] = true
	physical, err := Abc_SclBufferPhase(before, false)
	if err != nil {
		t.Fatal(err)
	}
	if physical.(ReadablePhases).SclHasPhases() {
		t.Fatal("physical phase metadata remains enabled")
	}
	requireBufferEquivalent(t, before, physical)
	encoded, err := Abc_SclUnBufferPhase(physical, false)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, encoded)
	if len(encoded.SclObjectIDs()) >= len(physical.SclObjectIDs()) {
		t.Fatal("phase encoding retained physical inverter")
	}
}
func TestUnbufferRemovesChains(t *testing.T) {
	before := bufferTestFixture(2)
	buf, _ := before.SclCreateNode([]int{5}, 0)
	inv, _ := before.SclCreateNode([]int{buf}, 3)
	again, _ := before.SclCreateNode([]int{inv}, 3)
	before.SclPatchFanin(10, 0, again)
	before.SclPatchFanin(11, 0, inv)
	after, err := Abc_SclUnBufferPerform(before, false)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
	if len(after.SclObjectIDs()) >= len(before.SclObjectIDs()) {
		t.Fatal("unary chain not removed")
	}
}
func TestOldCriticalBufferingActuallyDuplicates(t *testing.T) {
	before := bufferTestFixture(12)
	after, err := Abc_SclBufPerform(before, 3, 5, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
	if len(after.SclObjectIDs()) <= len(before.SclObjectIDs()) {
		t.Fatal("no clone/tree created")
	}
}
func TestLoadGainBufferSizing(t *testing.T) {
	before := bufferTestFixture(15)
	params := SC_BusPars{GainRatio: 300, Slew: 10, Degree: 3, AddBuffers: true}
	after, err := Abc_SclBufferingPerform(before, before.mapping.Lib, &params)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
	if len(after.SclObjectIDs()) <= len(before.SclObjectIDs()) {
		t.Fatal("load-driven buffering inserted no buffers")
	}
	for _, id := range after.SclObjectIDs() {
		if after.SclIsNode(id) && len(bufferEdges(after, id)) > params.Degree {
			t.Fatalf("degree violation at %d", id)
		}
	}
}
func TestLoadGainInvertersAndDuplicatePins(t *testing.T) {
	before := bufferTestFixture(10)
	before.enabled = true
	before.nodes[5] = bufferingTestNode{fanins: []int{0, 0}, node: true}
	params := SC_BusPars{GainRatio: 300, Slew: 10, Degree: 3, AddBuffers: false, BufferPIs: true}
	after, err := Abc_SclBufferingPerform(before, before.mapping.Lib, &params)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
}

func (n *bufferingTestNetwork) SclCloneTopological() (TimingNetwork, error) {
	order, err := bufferOrder(n)
	if err != nil {
		return nil, err
	}
	result := &bufferingTestNetwork{name: n.name, nodes: make(map[int]bufferingTestNode), mapping: NetworkMapping{Lib: n.mapping.Lib, Mio: n.mapping.Mio}, phases: make(map[FanoutPin]bool), enabled: n.enabled}
	mapping := make(map[int]int)
	copyNode := func(old int) {
		id := len(result.nodes)
		mapping[old] = id
		original := n.nodes[old]
		fans := make([]int, len(original.fanins))
		for pin, fanin := range original.fanins {
			fans[pin] = mapping[fanin]
			result.phases[FanoutPin{id, pin}] = n.phases[FanoutPin{old, pin}]
		}
		original.fanins = fans
		result.nodes[id] = original
		gate := int32(-1)
		if old < len(n.mapping.GateIDs) {
			gate = n.mapping.GateIDs[old]
		}
		result.mapping.GateIDs = append(result.mapping.GateIDs, gate)
		if old < len(n.mapping.MioGates) {
			result.mapping.MioGates = append(result.mapping.MioGates, n.mapping.MioGates[old])
		} else {
			result.mapping.MioGates = append(result.mapping.MioGates, nil)
		}
	}
	for _, id := range n.inputs {
		copyNode(id)
		result.inputs = append(result.inputs, mapping[id])
	}
	for _, id := range order {
		if _, ok := mapping[id]; !ok {
			copyNode(id)
		}
	}
	for _, id := range n.outputs {
		result.outputs = append(result.outputs, mapping[id])
	}
	return result, nil
}
func TestRegisteredBufferAndUnbufferCommands(t *testing.T) {
	frame := abcmain.NewFrame()
	defer frame.Close()
	frame.Out = &bytes.Buffer{}
	frame.Err = &bytes.Buffer{}
	Scl_Init(frame)
	network := bufferTestFixture(12)
	frame.SetResource("pLibScl", network.mapping.Lib)
	frame.SetCurrentNetwork(network)
	command, ok := frame.LookupCommand("buffer")
	if !ok {
		t.Fatal("buffer command not registered")
	}
	if status := command.Handler(frame, []string{"buffer", "-N3", "-S10"}); status != 0 {
		t.Fatalf("buffer command status %d: %s", status, frame.Err)
	}
	buffered := frame.Network().(TimingNetwork)
	requireBufferEquivalent(t, network, buffered)
	if len(buffered.SclObjectIDs()) <= len(network.SclObjectIDs()) {
		t.Fatal("registered command inserted no buffers")
	}
	command, ok = frame.LookupCommand("unbuffer")
	if !ok {
		t.Fatal("unbuffer command not registered")
	}
	if status := command.Handler(frame, []string{"unbuffer"}); status != 0 {
		t.Fatalf("unbuffer status %d: %s", status, frame.Err)
	}
	restored := frame.Network().(TimingNetwork)
	requireBufferEquivalent(t, network, restored)
	if len(restored.SclObjectIDs()) != len(network.SclObjectIDs()) {
		t.Fatal("registered unbuffer retained inserted unary nodes")
	}
}
func TestDuplicatePinCapacitanceStaysPerEdge(t *testing.T) {
	net := bufferTestFixture(1)
	net.nodes[5] = bufferingTestNode{fanins: []int{0, 0}, node: true}
	params := SC_BusPars{GainRatio: 300, Slew: 10, Degree: 3, AddBuffers: true}
	manager, err := Bus_ManStart(net, net.mapping.Lib, &params)
	if err != nil {
		t.Fatal(err)
	}
	if err = Abc_NtkComputeFanoutInfo(manager, 0, 10); err != nil {
		t.Fatal(err)
	}
	if load := Abc_NtkComputeNodeLoad(manager, 0); load != 3 {
		t.Fatalf("distinct pin loads collapsed: %g", load)
	}
}

func TestBufferingVectorCases(t *testing.T) {
	var cases []struct {
		Name       string `json:"name"`
		Fanouts    int    `json:"fanouts"`
		Degree     int    `json:"degree"`
		RootDegree int    `json:"root_degree"`
		Inverters  bool   `json:"inverters"`
	}
	data, err := os.ReadFile("test_vectors/buffering_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, item := range cases {
		t.Run(item.Name, func(t *testing.T) {
			before := bufferTestFixture(item.Fanouts)
			before.enabled = item.Inverters
			after, err := Abc_SclPerformBuffering(before, item.RootDegree, item.Degree, item.Inverters, false)
			if err != nil {
				t.Fatal(err)
			}
			requireBufferEquivalent(t, before, after)
			for _, id := range after.SclObjectIDs() {
				if len(bufferEdges(after, id)) > item.Degree {
					t.Fatalf("degree exceeded at %d", id)
				}
			}
		})
	}
}
func TestBarrierTransparentInverterBuffering(t *testing.T) {
	before := bufferTestFixture(12)
	before.enabled = true
	before.nodes[8] = bufferingTestNode{fanins: []int{5}, node: true, barrier: true}
	for _, output := range before.outputs {
		before.SclPatchFanin(output, 0, 8)
	}
	after, err := Abc_SclPerformBuffering(before, 0, 3, true, false)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
	params := SC_BusPars{GainRatio: 300, Slew: 10, Degree: 3, AddBuffers: true}
	after, err = Abc_SclBufferingPerform(before, before.mapping.Lib, &params)
	if err != nil {
		t.Fatal(err)
	}
	requireBufferEquivalent(t, before, after)
}

type readonlyBufferNetwork struct{ TimingNetwork }
type shallowBufferNetwork struct{ *bufferingTestNetwork }

func (n shallowBufferNetwork) SclClone() (TimingNetwork, error) { return n.bufferingTestNetwork, nil }
func TestBufferErrorsPreserveInput(t *testing.T) {
	before := bufferTestFixture(12)
	originalIDs := append([]int(nil), before.SclObjectIDs()...)
	originalFans := append([]int(nil), before.SclFanouts(5)...)
	for _, net := range []TimingNetwork{readonlyBufferNetwork{before}, shallowBufferNetwork{before}} {
		if _, err := Abc_SclPerformBuffering(net, 0, 3, false, false); err == nil {
			t.Fatal("unsafe/missing clone capability accepted")
		}
	}
	if _, err := Abc_SclPerformBuffering(before, 0, 1, false, false); err == nil {
		t.Fatal("degree-one tree accepted")
	}
	if _, err := Abc_SclPerformBuffering(before, 0, 3, true, false); err == nil {
		t.Fatal("inverter buffering accepted missing phases")
	}
	params := SC_BusPars{GainRatio: 1, Slew: 10, Degree: 2, AddBuffers: true}
	// A low gain must terminate without altering the original network.
	_, _ = Abc_SclBufferingPerform(before, before.mapping.Lib, &params)
	if !reflect.DeepEqual(originalIDs, before.SclObjectIDs()) || !reflect.DeepEqual(originalFans, before.SclFanouts(5)) {
		t.Fatal("rejected transform mutated input")
	}
}
func TestBufferRejectsMissingTimingWithoutPanic(t *testing.T) {
	before := bufferTestFixture(12)
	before.mapping.Lib.Cells[6].Pins[2].Timings = nil
	if _, err := Abc_SclBufPerform(before, 3, 5, false, false, false); err == nil {
		t.Fatal("missing pin timing accepted")
	}
	params := SC_BusPars{GainRatio: 300, Slew: 10, Degree: 3, AddBuffers: true}
	if _, err := Abc_SclBufferingPerform(before, before.mapping.Lib, &params); err == nil {
		t.Fatal("load buffering accepted missing timing")
	}
}
func TestBufferFanoutComparatorIsStrict(t *testing.T) {
	p := &Bus_Man_t{FanoutInfo: map[FanoutPin]busFanout{{1, 0}: {Cin: 1, Departure: 2}, {2, 0}: {Cin: 1, Departure: 2}}}
	a, b := FanoutPin{1, 0}, FanoutPin{2, 0}
	if Bus_SclCompareFanouts(p, a, a) != 0 || Bus_SclCompareFanouts(p, a, b) != -Bus_SclCompareFanouts(p, b, a) {
		t.Fatal("fanout comparator is not antisymmetric")
	}
}

func TestForcedTwoFanoutGroupingUsesCompleteLoad(t *testing.T) {
	net := bufferTestFixture(2)
	params := SC_BusPars{GainRatio: 100, Slew: 10, Degree: 2, AddBuffers: true}
	manager, err := Bus_ManStart(net, net.mapping.Lib, &params)
	if err != nil {
		t.Fatal(err)
	}
	edges := bufferEdges(net, 5)
	for _, edge := range edges {
		manager.OutputLoads[edge.Node] = 4
		manager.FanoutInfo[edge] = busFanout{Edge: edge, Cin: 4}
	}
	// The first fanout alone exceeds the target, but at least two must move.
	id, remaining, err := Abc_SclAddOneInv(manager, 5, edges, 1)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Loads[id] != 8 || len(remaining) != 1 {
		t.Fatalf("forced group load=%g fanouts=%v", manager.Loads[id], remaining)
	}
}

func TestBufferPrimaryInputControlIsEffective(t *testing.T) {
	before := bufferTestFixture(13)
	for _, output := range before.outputs {
		gate, err := before.SclCreateNode([]int{0, 3}, 6)
		if err != nil {
			t.Fatal(err)
		}
		if err = before.SclPatchFanin(output, 0, gate); err != nil {
			t.Fatal(err)
		}
	}
	if err := before.SclDeleteNode(5); err != nil {
		t.Fatal(err)
	}
	disabled, err := Abc_SclBufPerform(before, 3, 5, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	enabled, err := Abc_SclBufPerform(before, 3, 5, true, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(bufferEdges(disabled, disabled.SclInputs()[0])) != 13 || len(bufferEdges(enabled, enabled.SclInputs()[0])) > 5 {
		t.Fatal("old buffer PI flag had no effect")
	}
	requireBufferEquivalent(t, before, enabled)
	params := SC_BusPars{GainRatio: 300, Slew: 10, Degree: 3, AddBuffers: true, BufferPIs: true}
	enabled, err = Abc_SclBufferingPerform(before, before.mapping.Lib, &params)
	if err != nil {
		t.Fatal(err)
	}
	if len(bufferEdges(enabled, enabled.SclInputs()[0])) > 3 {
		t.Fatal("load buffer PI flag had no effect")
	}
	requireBufferEquivalent(t, before, enabled)
}
