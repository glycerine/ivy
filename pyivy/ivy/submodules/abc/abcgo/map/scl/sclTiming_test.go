package scl

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type timingFixture struct {
	Name     string        `json:"name"`
	IDs      []int         `json:"ids"`
	Inputs   []int         `json:"inputs"`
	Outputs  []int         `json:"outputs"`
	Fanins   map[int][]int `json:"fanins"`
	Nodes    []int         `json:"nodes"`
	RiseCap  float32       `json:"rise_cap"`
	FallCap  float32       `json:"fall_cap"`
	RiseLoad float32       `json:"output_rise_load"`
	FallLoad float32       `json:"output_fall_load"`
	Rise     float32       `json:"expected_rise"`
	Fall     float32       `json:"expected_fall"`
}
type timingTestNetwork struct {
	fixture  timingFixture
	mapping  NetworkMapping
	barriers map[int]bool
}

func (n *timingTestNetwork) SclName() string        { return n.fixture.Name }
func (n *timingTestNetwork) SclObjectIDs() []int    { return n.fixture.IDs }
func (n *timingTestNetwork) SclInputs() []int       { return n.fixture.Inputs }
func (n *timingTestNetwork) SclOutputs() []int      { return n.fixture.Outputs }
func (n *timingTestNetwork) SclFanins(id int) []int { return n.fixture.Fanins[id] }
func (n *timingTestNetwork) SclFanouts(id int) []int {
	var fanouts []int
	for _, other := range n.fixture.IDs {
		for _, fanin := range n.fixture.Fanins[other] {
			if fanin == id {
				fanouts = append(fanouts, other)
			}
		}
	}
	return fanouts
}
func (n *timingTestNetwork) SclIsNode(id int) bool {
	for _, node := range n.fixture.Nodes {
		if node == id {
			return true
		}
	}
	return false
}
func (n *timingTestNetwork) SclIsBarrier(id int) bool    { return n.barriers[id] }
func (n *timingTestNetwork) SclMapping() *NetworkMapping { return &n.mapping }
func linearTimingSurface(base, slew, load float32) SC_Surface {
	return SC_Surface{Index0: []float32{0, 10}, Index1: []float32{0, 10}, Data: [][]float32{{base, base + 10*load}, {base + 10*slew, base + 10*(slew+load)}}}
}
func timingTestLibrary(riseCap, fallCap float32) *SC_Lib {
	lib := &SC_Lib{Name: "timing", UnitTime: 12, UnitCapMultiplier: 1, UnitCapExponent: 15}
	for i := 0; i < 2; i++ {
		timing := &SC_Timing{RelatedPin: "A", Sense: ScTSPos, CellRise: linearTimingSurface(1, 0, .5), CellFall: linearTimingSurface(2, 0, .25), RiseTrans: linearTimingSurface(.2, 0, .1), FallTrans: linearTimingSurface(.3, 0, .2)}
		input := &SC_Pin{Name: "A", Direction: ScDirInput, RiseCap: riseCap * float32(i+1), FallCap: fallCap * float32(i+1)}
		output := &SC_Pin{Name: "Y", Direction: ScDirOutput, Function: "A", Truth: []uint64{0xaaaaaaaaaaaaaaaa}, Timings: []*SC_Timings{{Name: "A", Timings: []*SC_Timing{timing}}}}
		lib.Cells = append(lib.Cells, &SC_Cell{Name: []string{"small", "large"}[i], ID: i, Inputs: 1, Outputs: 1, Area: float32(i + 1), Pins: []*SC_Pin{input, output}})
	}
	Abc_SclHashCells(lib)
	Abc_SclLinkCells(lib)
	return lib
}
func loadTimingFixture(t *testing.T) (*SC_Lib, *timingTestNetwork) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("test_vectors", "timing_chain.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture timingFixture
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	lib := timingTestLibrary(fixture.RiseCap, fixture.FallCap)
	mapping := NetworkMapping{Lib: lib, GateIDs: make([]int32, 32)}
	for i := range mapping.GateIDs {
		mapping.GateIDs[i] = -1
	}
	for _, id := range fixture.Nodes {
		mapping.GateIDs[id] = 0
	}
	return lib, &timingTestNetwork{fixture: fixture, mapping: mapping}
}
func timingManagerFixture(t *testing.T) (*SC_Man, *timingTestNetwork) {
	t.Helper()
	lib, network := loadTimingFixture(t)
	p, err := NewTimingManager(lib, network)
	if err != nil {
		t.Fatal(err)
	}
	p.Output = io.Discard
	out := network.fixture.Outputs[0]
	p.Loads[out] = SC_Pair{network.fixture.RiseLoad, network.fixture.FallLoad}
	if err = Abc_SclTimeNtkRecompute(p, &p.InitialArea, &p.InitialDelay, 1, 0); err != nil {
		t.Fatal(err)
	}
	return p, network
}
func closeTiming(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-5 }
func TestSparseTopologicalTimingAndDepartures(t *testing.T) {
	p, network := timingManagerFixture(t)
	out := network.fixture.Outputs[0]
	if !closeTiming(p.Times[out].Rise, network.fixture.Rise) || !closeTiming(p.Times[out].Fall, network.fixture.Fall) {
		t.Fatal(p.Times[out])
	}
	if !closeTiming(p.Departures[1].Rise, network.fixture.Rise) || !closeTiming(p.Departures[1].Fall, network.fixture.Fall) || p.SumArea != 2 {
		t.Fatal("reverse timing", p.Departures[1], p.SumArea)
	}
	phase := int32(-1)
	if Abc_SclFindCriticalCo(p, &phase) != 31 || phase != 0 || Abc_SclFindMostCriticalFanin(p, &phase, 17) != 4 {
		t.Fatal("critical path")
	}
	var report bytes.Buffer
	p.Output = &report
	if err := Abc_SclTimeNtkPrint(p, 1, 1); err != nil || report.Len() == 0 {
		t.Fatal(err)
	}
	p.Output = errorTimingWriter{}
	if Abc_SclTimeNtkPrint(p, 0, 0) == nil {
		t.Fatal("ignored timing output error")
	}
}

type errorTimingWriter struct{}

func (errorTimingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestIncrementalSizingMatchesFullRecompute(t *testing.T) {
	p, network := timingManagerFixture(t)
	old, new := p.Lib.Cells[0], p.Lib.Cells[1]
	if err := Abc_SclUpdateLoad(p, 17, old, new); err != nil {
		t.Fatal(err)
	}
	Abc_SclObjSetCell(p, 17, new)
	Abc_SclTimeIncInsert(p, 17)
	updates, err := Abc_SclTimeIncUpdate(p)
	if err != nil || updates == 0 {
		t.Fatal(updates, err)
	}
	times, departs, loads := append([]SC_Pair(nil), p.Times...), append([]SC_Pair(nil), p.Departures...), append([]SC_Pair(nil), p.Loads...)
	if err = Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); err != nil {
		t.Fatal(err)
	}
	for _, id := range network.fixture.IDs {
		if times[id] != p.Times[id] || departs[id] != p.Departures[id] || loads[id] != p.Loads[id] {
			t.Fatal("incremental/full mismatch", id, times[id], p.Times[id], departs[id], p.Departures[id], loads[id], p.Loads[id])
		}
	}
	if p.MaxDelay != 7 || p.SumArea != 3 {
		t.Fatal("resize result", p.MaxDelay, p.SumArea)
	}
	if count, err := Abc_SclTimeIncUpdate(p); err != nil || count != 0 {
		t.Fatal("empty update", count, err)
	}
}
func TestWireLoadInterpolationExtrapolationAndZeroPoints(t *testing.T) {
	model := &SC_WireLoad{Cap: 2, Slope: 3, Fanout: []int32{1, 3}, Length: []float32{2, 6}}
	caps, err := Abc_SclFindWireCaps(model, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(caps, []float32{0, 4, 8, 12, 18, 24}) {
		t.Fatal(caps)
	}
	model.Length[0] = 0
	caps, err = Abc_SclFindWireCaps(model, 5)
	if err != nil || caps[2] != 6 {
		t.Fatal("zero interpolation anchor", caps, err)
	}
	model.Fanout[1] = 1
	if _, err = Abc_SclFindWireCaps(model, 5); err == nil {
		t.Fatal("accepted duplicate anchor")
	}
}
func TestTimingInvalidTopologyAndMissingArcs(t *testing.T) {
	lib, network := loadTimingFixture(t)
	network.fixture.Fanins[4] = []int{17}
	if _, err := NewTimingManager(lib, network); err == nil {
		t.Fatal("accepted cycle")
	}
	lib, network = loadTimingFixture(t)
	network.fixture.Inputs = []int{99}
	if _, err := NewTimingManager(lib, network); err == nil {
		t.Fatal("accepted absent input")
	}
	p, _ := timingManagerFixture(t)
	p.Lib.Cells[0].Pins[1].Timings = nil
	if err := Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); err == nil {
		t.Fatal("accepted missing timing arc")
	}
}
func TestBarrierLoadAndConeBackups(t *testing.T) {
	lib, network := loadTimingFixture(t)
	network.fixture.IDs = append(network.fixture.IDs, 8)
	network.fixture.Nodes = append(network.fixture.Nodes, 8)
	network.fixture.Fanins[8] = []int{4}
	network.fixture.Fanins[17] = []int{8}
	network.barriers = map[int]bool{8: true}
	p, err := NewTimingManager(lib, network)
	if err != nil {
		t.Fatal(err)
	}
	p.Loads[31] = SC_Pair{4, 6}
	if err = Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); err != nil {
		t.Fatal(err)
	}
	if p.Times[8] != p.Times[4] || p.Loads[4] != (SC_Pair{2, 3}) || p.MaxDelay != 6.25 {
		t.Fatal("barrier timing/load", p.Times[8], p.Loads[4], p.MaxDelay)
	}
	cone := []int{4, 8, 17}
	Abc_SclConeStore(p, cone)
	before := p.Times[17]
	Abc_SclConeClean(p, cone)
	Abc_SclConeRestore(p, cone)
	if p.Times[17] != before {
		t.Fatal("backup restore")
	}
	Abc_SclLoadStore(p, 17)
	p.Loads[8] = SC_Pair{99, 99}
	Abc_SclLoadRestore(p, 17)
	if p.Loads[8] != (SC_Pair{2, 3}) {
		t.Fatal("load restore")
	}
}

func TestRandomIncrementalTimingDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(9912))
	for trial := 0; trial < 40; trial++ {
		count := 5 + rng.Intn(12)
		lib := timingTestLibrary(2, 3)
		fixture := timingFixture{Name: "random", Inputs: []int{0}, Fanins: make(map[int][]int)}
		mapping := NetworkMapping{Lib: lib, GateIDs: make([]int32, count+2)}
		for i := range mapping.GateIDs {
			mapping.GateIDs[i] = -1
		}
		for i := 0; i < count+2; i++ {
			fixture.IDs = append(fixture.IDs, i)
			if i > 0 {
				fixture.Fanins[i] = []int{i - 1}
			}
			if i > 0 && i <= count {
				fixture.Nodes = append(fixture.Nodes, i)
				mapping.GateIDs[i] = int32(rng.Intn(2))
			}
		}
		fixture.Outputs = []int{count + 1}
		rng.Shuffle(len(fixture.IDs), func(i, j int) { fixture.IDs[i], fixture.IDs[j] = fixture.IDs[j], fixture.IDs[i] })
		net := &timingTestNetwork{fixture: fixture, mapping: mapping}
		p, err := NewTimingManager(lib, net)
		if err != nil {
			t.Fatal(err)
		}
		p.Loads[count+1] = SC_Pair{4, 6}
		if err = Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); err != nil {
			t.Fatal(err)
		}
		pivot := 1 + rng.Intn(count)
		old := lib.Cells[mapping.GateIDs[pivot]]
		replacement := lib.Cells[1-mapping.GateIDs[pivot]]
		if err = Abc_SclUpdateLoad(p, pivot, old, replacement); err != nil {
			t.Fatal(err)
		}
		Abc_SclObjSetCell(p, pivot, replacement)
		Abc_SclTimeIncInsert(p, pivot)
		if _, err = Abc_SclTimeIncUpdate(p); err != nil {
			t.Fatal(err)
		}
		times, departs := append([]SC_Pair(nil), p.Times...), append([]SC_Pair(nil), p.Departures...)
		if err = Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); err != nil {
			t.Fatal(err)
		}
		for _, id := range fixture.IDs {
			if !closeTiming(times[id].Rise, p.Times[id].Rise) || !closeTiming(times[id].Fall, p.Times[id].Fall) || !closeTiming(departs[id].Rise, p.Departures[id].Rise) || !closeTiming(departs[id].Fall, p.Departures[id].Fall) {
				t.Fatal("incremental differential", trial, pivot, id, times[id], p.Times[id], departs[id], p.Departures[id])
			}
		}
	}
}
func TestNegativeAndNonUnateArcs(t *testing.T) {
	for _, sense := range []SC_TSense{ScTSNeg, ScTSNon} {
		p, net := timingManagerFixture(t)
		Scl_CellPinTime(p.Lib.Cells[0], 0).Sense = sense
		if err := Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); err != nil {
			t.Fatal(err)
		}
		if sense == ScTSNeg {
			if !closeTiming(p.Times[31].Rise, 5.75) || !closeTiming(p.Times[31].Fall, 5.5) {
				t.Fatal("inverting timing", p.Times[31])
			}
		} else if !closeTiming(p.Times[31].Rise, 5.75) || !closeTiming(p.Times[31].Fall, 6.25) {
			t.Fatal("non-unate timing", p.Times[31])
		}
		if !sclFinite(p.Departures[net.fixture.Inputs[0]].Rise) {
			t.Fatal("invalid reverse phase")
		}
	}
}
func TestBufferEstimateRestoresLoads(t *testing.T) {
	p, _ := timingManagerFixture(t)
	before := append([]SC_Pair(nil), p.Loads...)
	p.EstLoadMax = .5
	p.EstLinear = 100
	if err := Abc_SclTimeNtkRecompute(p, nil, nil, 1, 0); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, p.Loads) || p.EstimatedNodes == 0 || p.MaxDelay <= 6.25 {
		t.Fatal("buffer estimate corrupted physical loads", p.Loads, p.EstimatedNodes, p.MaxDelay)
	}
}

func TestTimingSparseHoleAndLevelValidation(t *testing.T) {
	p, _ := timingManagerFixture(t)
	Abc_SclTimeIncInsert(p, 3) // Within allocation, absent from sparse topology.
	if _, err := Abc_SclTimeIncUpdate(p); err == nil {
		t.Fatal("accepted missing changed object")
	}
	p.Changed = nil
	if err := Abc_SclTimeIncCheckLevel(p); err != nil {
		t.Fatal(err)
	}
	p.Levels[17] = 99
	if Abc_SclTimeIncCheckLevel(p) == nil {
		t.Fatal("did not detect stale level")
	}
	if err := Abc_SclTimeIncUpdateLevel(p, 17); err != nil {
		t.Fatal(err)
	}
	if err := Abc_SclTimeIncCheckLevel(p); err != nil {
		t.Fatal(err)
	}
}
