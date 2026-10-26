package abc

// Go-only tests for src/map/amap/amapMan.c and amapCore.c.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func buildAMapCoreTestAIG(t testing.TB) *AigManager {
	t.Helper()
	manager := NewAigManager()
	a, err := manager.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := manager.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	node, err := manager.And(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.CreateCO(node); err != nil {
		t.Fatal(err)
	}
	return manager
}

func TestAMapDefaultsReplaceEveryField(t *testing.T) {
	want := AMapParams{FlowIterations: 1, AreaIterations: 4, MaximumCuts: 500, UseXORs: true, Epsilon: 0.001}
	if got := DefaultAMapParams(); got != want {
		t.Fatalf("defaults = %+v want %+v", got, want)
	}
	params := AMapParams{
		FlowIterations: 99, AreaIterations: 98, MaximumCuts: 97,
		UseMUXes: true, FreeInverters: true, AreaDelayRatio: 4,
		Verbose: true,
	}
	if err := SetDefaultAMapParams(&params); err != nil || params != want {
		t.Fatalf("set defaults = %+v, %v", params, err)
	}
	if err := SetDefaultAMapParams(nil); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil params = %v", err)
	}
}

func TestAMapManagerStartStopAndNativeCapacities(t *testing.T) {
	manager, err := NewAMapManager(37, AMapManagerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := manager.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.EpsilonInternal != 0.01 || snapshot.ObjectCapacity != 100 || snapshot.ObjectChunkSize != 8 ||
		snapshot.InputCapacity != 100 || snapshot.OutputCapacity != 100 ||
		snapshot.TemporaryCapacity != 100 || snapshot.CutCapacities != [3]int{100, 100, 100} ||
		snapshot.MatchSlots != 0 || snapshot.CutSlots != 0 {
		t.Fatalf("start snapshot = %+v", snapshot)
	}
	if err = manager.Close(); err != nil || !manager.Closed() {
		t.Fatalf("close = %v, closed %v", err, manager.Closed())
	}
	if err = manager.Close(); err != nil {
		t.Fatalf("second close = %v", err)
	}
	if _, err = manager.Snapshot(); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed snapshot = %v", err)
	}
	if err = (*AMapManager)(nil).Close(); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil close = %v", err)
	}
	large, err := NewAMapManager(800, AMapManagerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	largeSnapshot, err := large.Snapshot()
	if err != nil || largeSnapshot.ObjectCapacity != 100 || largeSnapshot.ObjectChunkSize != 100 {
		t.Fatalf("large manager = %+v, %v", largeSnapshot, err)
	}
	if err = large.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRunAMapCoreOrchestrationOwnershipAndFeatures(t *testing.T) {
	aig := buildAMapCoreTestAIG(t)
	params := DefaultAMapParams()
	params.UseMUXes = true
	params.FreeInverters = false
	library := &AMapLibrary{Name: "test", InverterArea: 1.25, HasXOR: false, HasMUX: true, NodeCount: 7}
	order := make([]int, 0, 3)
	returnedFanins := []int{2, 3}
	var retained *AMapManager
	pipeline := AMapPipeline{
		Create: func(manager *AMapManager, graph *AigManager) error {
			retained = manager
			order = append(order, 1)
			if graph != aig {
				t.Fatal("graph identity changed")
			}
			snapshot, err := manager.Snapshot()
			if err != nil {
				return err
			}
			if snapshot.InverterArea != 1.25 || snapshot.UseXOR || !snapshot.UseMUX || snapshot.MatchSlots != 14 || snapshot.CutSlots != 14 {
				t.Fatalf("configured snapshot = %+v", snapshot)
			}
			return nil
		},
		Map: func(manager *AMapManager) error {
			order = append(order, 2)
			return manager.check()
		},
		Produce: func(manager *AMapManager) ([]AMapOutput, error) {
			order = append(order, 3)
			return []AMapOutput{{Name: "and2", Type: 0, Fanins: returnedFanins}}, nil
		},
	}
	outputs, err := RunAMapCore(aig, params, library, pipeline, AMapManagerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(order, []int{1, 2, 3}) || len(outputs) != 1 || outputs[0].Name != "and2" || !slices.Equal(outputs[0].Fanins, []int{2, 3}) {
		t.Fatalf("order/output = %v %+v", order, outputs)
	}
	returnedFanins[0] = 99
	if outputs[0].Fanins[0] != 2 {
		t.Fatal("result retained producer storage")
	}
	if retained == nil || !retained.Closed() {
		t.Fatal("manager was not closed after success")
	}

	params.FreeInverters = true
	params.UseXORs = true
	library.HasXOR = true
	pipeline.Create = func(manager *AMapManager, _ *AigManager) error {
		snapshot, snapshotErr := manager.Snapshot()
		if snapshotErr != nil {
			return snapshotErr
		}
		if snapshot.InverterArea != 0 || !snapshot.UseXOR || !snapshot.UseMUX {
			t.Fatalf("free-inverter snapshot = %+v", snapshot)
		}
		return nil
	}
	if _, err = RunAMapCore(aig, params, library, pipeline, AMapManagerOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestRunAMapCoreErrorsLimitsCancellationAndCleanup(t *testing.T) {
	aig := buildAMapCoreTestAIG(t)
	params := DefaultAMapParams()
	library := &AMapLibrary{InverterArea: 1, NodeCount: 3}
	ok := AMapPipeline{
		Create: func(*AMapManager, *AigManager) error { return nil },
		Map:    func(*AMapManager) error { return nil },
		Produce: func(*AMapManager) ([]AMapOutput, error) {
			return []AMapOutput{{Type: 1}}, nil
		},
	}
	badParams := []AMapParams{
		{},
		{FlowIterations: -1, AreaIterations: 1, MaximumCuts: 1, Epsilon: 1},
		{FlowIterations: 1, AreaIterations: -1, MaximumCuts: 1, Epsilon: 1},
		{FlowIterations: 1, AreaIterations: 1, MaximumCuts: 0, Epsilon: 1},
		{FlowIterations: 1, AreaIterations: 1, MaximumCuts: 1, Epsilon: -1},
		{FlowIterations: 1, AreaIterations: 1, MaximumCuts: 1, Epsilon: float32(math.Inf(1))},
		{FlowIterations: 1, AreaIterations: 1, MaximumCuts: 1, Epsilon: 1, AreaDelayRatio: -1},
	}
	for _, bad := range badParams {
		if result, err := RunAMapCore(aig, bad, library, ok, AMapManagerOptions{}); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("bad params %+v = %+v, %v", bad, result, err)
		}
	}
	for _, badLibrary := range []*AMapLibrary{
		nil,
		{NodeCount: -1},
		{NodeCount: 1, InverterArea: -1},
		{NodeCount: 1, InverterArea: math.Inf(1)},
	} {
		if result, err := RunAMapCore(aig, params, badLibrary, ok, AMapManagerOptions{}); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("bad library %+v = %+v, %v", badLibrary, result, err)
		}
	}
	if result, err := RunAMapCore(nil, params, library, ok, AMapManagerOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil graph = %+v, %v", result, err)
	}
	for _, invalid := range []AMapPipeline{{}, {Create: ok.Create}, {Create: ok.Create, Map: ok.Map}} {
		if result, err := RunAMapCore(aig, params, library, invalid, AMapManagerOptions{}); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("invalid pipeline = %+v, %v", result, err)
		}
	}
	if result, err := RunAMapCore(aig, params, library, ok, AMapManagerOptions{MaxTemporaryEntries: 5}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("temporary limit = %+v, %v", result, err)
	}
	if _, err := NewAMapManager(2, AMapManagerOptions{MaxNodes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("node limit = %v", err)
	}
	if _, err := NewAMapManager(-1, AMapManagerOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative node capacity = %v", err)
	}

	phaseError := errors.New("phase failure")
	for phase := 0; phase < 3; phase++ {
		var retained *AMapManager
		failing := AMapPipeline{
			Create: func(manager *AMapManager, _ *AigManager) error {
				retained = manager
				if phase == 0 {
					return phaseError
				}
				return nil
			},
			Map: func(*AMapManager) error {
				if phase == 1 {
					return phaseError
				}
				return nil
			},
			Produce: func(*AMapManager) ([]AMapOutput, error) {
				if phase == 2 {
					return nil, phaseError
				}
				return nil, nil
			},
		}
		if result, err := RunAMapCore(aig, params, library, failing, AMapManagerOptions{}); result != nil || !errors.Is(err, phaseError) || retained == nil || !retained.Closed() {
			t.Fatalf("phase %d = %+v, %v, closed %v", phase, result, err, retained != nil && retained.Closed())
		}
	}

	for cancelAt := 1; cancelAt <= 5; cancelAt++ {
		calls := 0
		options := AMapManagerOptions{Stop: func() bool { calls++; return calls == cancelAt }}
		if result, err := RunAMapCore(aig, params, library, ok, options); result != nil || !errors.Is(err, ErrCancelled) {
			t.Fatalf("cancel %d = %+v, %v (calls %d)", cancelAt, result, err, calls)
		}
	}
	invalidOutput := ok
	invalidOutput.Produce = func(*AMapManager) ([]AMapOutput, error) { return []AMapOutput{{Type: 2}}, nil }
	if result, err := RunAMapCore(aig, params, library, invalidOutput, AMapManagerOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatalf("invalid output = %+v, %v", result, err)
	}
}

func TestAMapCoreIndependentModelCampaign(t *testing.T) {
	aig := buildAMapCoreTestAIG(t)
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		return state
	}
	for execution := 0; execution < 5000; execution++ {
		bits := next()
		params := DefaultAMapParams()
		params.UseMUXes = bits&1 != 0
		params.UseXORs = bits&2 != 0
		params.FreeInverters = bits&4 != 0
		area := float64(1+bits%1000) / 32
		library := &AMapLibrary{
			InverterArea: area,
			HasMUX:       bits&8 != 0,
			HasXOR:       bits&16 != 0,
			NodeCount:    int((bits >> 5) % 31),
		}
		visited := 0
		pipeline := AMapPipeline{
			Create: func(manager *AMapManager, _ *AigManager) error {
				visited = visited*4 + 1
				snapshot, err := manager.Snapshot()
				if err != nil {
					return err
				}
				wantArea := area
				if params.FreeInverters {
					wantArea = 0
				}
				if snapshot.InverterArea != wantArea || snapshot.ObjectCapacity != 100 || snapshot.ObjectChunkSize != 8 ||
					snapshot.UseMUX != (params.UseMUXes && library.HasMUX) ||
					snapshot.UseXOR != (params.UseXORs && library.HasXOR) || snapshot.MatchSlots != 2*library.NodeCount || snapshot.CutSlots != 2*library.NodeCount {
					return ErrInvariant
				}
				return nil
			},
			Map: func(*AMapManager) error { visited = visited*4 + 2; return nil },
			Produce: func(*AMapManager) ([]AMapOutput, error) {
				visited = visited*4 + 3
				return []AMapOutput{{Type: int16(int(bits%3) - 1), Fanins: []int{execution}}}, nil
			},
		}
		outputs, err := RunAMapCore(aig, params, library, pipeline, AMapManagerOptions{})
		if err != nil || visited != 27 || len(outputs) != 1 || outputs[0].Fanins[0] != execution {
			t.Fatalf("execution %d: visited %d output %+v err %v", execution, visited, outputs, err)
		}
	}
}

func FuzzAMapCore(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4})
	f.Add([]byte{255, 0, 19, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		aig := buildAMapCoreTestAIG(t)
		flags := data[0]
		params := DefaultAMapParams()
		params.UseMUXes = flags&1 != 0
		params.UseXORs = flags&2 != 0
		params.FreeInverters = flags&4 != 0
		nodes := 0
		if len(data) > 1 {
			nodes = int(data[1] % 32)
		}
		library := &AMapLibrary{
			InverterArea: float64(1+flags) / 17,
			HasMUX:       flags&8 != 0,
			HasXOR:       flags&16 != 0,
			NodeCount:    nodes,
		}
		calls := 0
		fanins := make([]int, len(data))
		for index, value := range data {
			fanins[index] = int(value)
		}
		pipeline := AMapPipeline{
			Create: func(manager *AMapManager, _ *AigManager) error {
				calls++
				snapshot, err := manager.Snapshot()
				if err != nil {
					return err
				}
				if snapshot.MatchSlots != 2*nodes || snapshot.UseMUX != (params.UseMUXes && library.HasMUX) || snapshot.UseXOR != (params.UseXORs && library.HasXOR) {
					return ErrInvariant
				}
				return nil
			},
			Map: func(*AMapManager) error { calls++; return nil },
			Produce: func(*AMapManager) ([]AMapOutput, error) {
				calls++
				return []AMapOutput{{Type: int16(int(flags%3) - 1), Fanins: fanins}}, nil
			},
		}
		outputs, err := RunAMapCore(aig, params, library, pipeline, AMapManagerOptions{})
		if err != nil || calls != 3 || len(outputs) != 1 || !slices.Equal(outputs[0].Fanins, fanins) {
			t.Fatalf("calls %d outputs %+v err %v", calls, outputs, err)
		}
		if len(fanins) > 0 {
			fanins[0] ^= 255
			if outputs[0].Fanins[0] == fanins[0] {
				t.Fatal("output aliases producer fanins")
			}
		}
	})
}
