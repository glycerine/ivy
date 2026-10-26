package abc

// Go-only tests for src/opt/fsim/{fsim.h,fsimInt.h,fsimFront.c,fsimMan.c,
// fsimSim.c}.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func fsimTestGraph(t testing.TB, invert bool) *AigManager {
	t.Helper()
	graph := NewAigManager()
	x, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	y, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	latch, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	if invert {
		y = y.Not()
	}
	next, err := graph.And(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(latch); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(next); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = graph.Check(); err != nil {
		t.Fatal(err)
	}
	return graph
}

func fsimSeedGraph(t testing.TB, seed uint64, invert bool) *AigManager {
	t.Helper()
	graph := NewAigManager()
	x, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	y, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	latch, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	bases := []AigLit{x, y, latch}
	current := x
	choice := seed ^ 0xd1b54a32d192ed03
	for depth := 0; depth < 6; depth++ {
		choice = choice*6364136223846793005 + 1442695040888963407
		other := bases[int(choice%uint64(len(bases)))]
		if other.ID() == current.ID() {
			other = bases[(int(choice%uint64(len(bases)))+1)%len(bases)]
		}
		if choice>>8&1 != 0 {
			other = other.Not()
		}
		current, err = graph.And(current, other)
		if err != nil {
			t.Fatal(err)
		}
	}
	if invert {
		current = current.Not()
	}
	if _, err = graph.CreateCO(current); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(y.Not()); err != nil {
		t.Fatal(err)
	}
	if _, err = graph.CreateCO(current.Not()); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = graph.Check(); err != nil {
		t.Fatal(err)
	}
	return graph
}

func fsimScalarExpected(t testing.TB, manager *FsimManager, parameters FsimSimulationParams, seed uint64) (frames, property, pattern int, found bool) {
	t.Helper()
	random := newFsimRandom(seed)
	states := make([]bool, 32)
	for frame := 0; frame < parameters.Iterations; frame++ {
		inputWords := make([]uint32, manager.nPis)
		for _, pio := range manager.cisToPIO {
			if pio < manager.nPis {
				inputWords[pio] = random.word()
			}
		}
		outputWords := make([]uint32, manager.nPos)
		nextStates := make([]bool, 32)
		for bit := 0; bit < 32; bit++ {
			primary := make([]bool, manager.nPis)
			for input := range primary {
				primary[input] = inputWords[input]&(uint32(1)<<uint(bit)) != 0
			}
			outputs, next, err := manager.graph.Step(primary, []bool{states[bit]})
			if err != nil {
				t.Fatal(err)
			}
			for output, value := range outputs {
				if value {
					outputWords[output] |= uint32(1) << uint(bit)
				}
			}
			nextStates[bit] = next[0]
		}
		for output, word := range outputWords {
			if word != 0 {
				return frame + 1, output, bits.TrailingZeros32(word), true
			}
		}
		states = nextStates
	}
	return parameters.Iterations, -1, -1, false
}

func TestFsimLiteralAndCodecDefinitions(t *testing.T) {
	native := ABCRandom{}
	random := newFsimRandom(0)
	for index := 0; index < 16; index++ {
		if got, want := random.word(), native.Uint32(false); got != want {
			t.Fatalf("native random word %d: got %08x want %08x", index, got, want)
		}
	}
	for variable := 0; variable < 128; variable++ {
		for _, complement := range []bool{false, true} {
			literal := fsimVar2Lit(variable, complement)
			if fsimLit2Var(literal) != variable || fsimLitIsCompl(literal) != complement || fsimLitNot(fsimLitNot(literal)) != literal || fsimLitNotIf(literal, true) != literal^1 || fsimLitNotIf(literal, false) != literal || fsimLitRegular(literal) != variable<<1 {
				t.Fatalf("literal helpers failed for %d/%v", variable, complement)
			}
		}
	}
	objects := []FsimObject{
		{Node: 1},
		{Node: 3},
		{Node: 7, Fanin0: 7, Fanin1: 2},
		{Fanin0: 13},
	}
	data, err := EncodeFsimObjects(objects, 100)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeFsimObjects(data, len(objects))
	if err != nil || !slices.Equal(decoded, objects) {
		t.Fatalf("codec round trip: %#v %v", decoded, err)
	}
	data[0] ^= 1
	again, err := EncodeFsimObjects(objects, 100)
	if err != nil || slices.Equal(data, again) {
		t.Fatal("encoded results must be independently owned")
	}
	if _, err = EncodeFsimObjects(objects, 1); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %v", err)
	}
	if _, err = EncodeFsimObjects([]FsimObject{{Node: -1}}, 10); !errors.Is(err, ErrInvariant) {
		t.Fatalf("bad object: %v", err)
	}
	if _, err = DecodeFsimObjects([]byte{0x80}, 1); !errors.Is(err, ErrFormat) {
		t.Fatalf("truncated varint: %v", err)
	}
	if _, err = DecodeFsimObjects([]byte{1, 0}, 1); !errors.Is(err, ErrInvariant) {
		t.Fatalf("trailing data: %v", err)
	}
	if _, err = DecodeFsimObjects(nil, 1<<30); !errors.Is(err, ErrFormat) {
		t.Fatalf("impossible decoded count: %v", err)
	}
}

func TestFsimManagerFrontierCompressedAndOwnership(t *testing.T) {
	graph := fsimTestGraph(t, true)
	plain, err := NewFsimManager(graph, FsimBuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := NewFsimManager(graph, FsimBuildOptions{Compress: true})
	if err != nil {
		t.Fatal(err)
	}
	plainStats, compressedStats := plain.Statistics(), compressed.Statistics()
	if plainStats.PrimaryInputs != 2 || plainStats.PrimaryOutputs != 1 || plainStats.Inputs != 3 || plainStats.Outputs != 2 || plainStats.Nodes != 1 || plainStats.Objects != 8 || plainStats.MaximumCrossCut < 1 || plainStats.Frontier < 2 {
		t.Fatalf("unexpected statistics: %+v", plainStats)
	}
	if compressedStats.CompressedBytes == 0 || compressedStats.Objects != plainStats.Objects || !slices.Equal(plain.Objects(), compressed.Objects()) {
		t.Fatalf("compressed mismatch: %+v", compressedStats)
	}
	if err = plain.Verify(); err != nil {
		t.Fatal(err)
	}
	if err = compressed.Verify(); err != nil {
		t.Fatal(err)
	}
	objects := compressed.Objects()
	data := compressed.CompressedData()
	objects[0].Node++
	data[0] ^= 0xff
	if err = compressed.Verify(); err != nil {
		t.Fatalf("snapshot mutated manager: %v", err)
	}
	compressed.Close()
	compressed.Close()
	if _, err = compressed.Simulate(DefaultFsimSimulationParams(), FsimRunOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("closed manager: %v", err)
	}
}

func TestFsimPackedSequentialSimulationAndCounterexample(t *testing.T) {
	graph := fsimTestGraph(t, false)
	parameters := FsimSimulationParams{Words: 2, Iterations: 4, CheckMiter: true}
	plain, err := graph.FsimSimulate(parameters, FsimBuildOptions{}, FsimRunOptions{Seed: 17})
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := graph.FsimSimulate(parameters, FsimBuildOptions{Compress: true}, FsimRunOptions{Seed: 17})
	if err != nil {
		t.Fatal(err)
	}
	if plain.Frames != 2 || plain.Property != 0 || plain.Counterexample == nil || plain.Pattern < 0 || plain.Pattern >= 64 {
		t.Fatalf("unexpected result: %+v", plain)
	}
	if compressed.Frames != plain.Frames || compressed.Property != plain.Property || compressed.Pattern != plain.Pattern || compressed.Counterexample == nil || !slices.Equal(compressed.Counterexample.PackedWords(), plain.Counterexample.PackedWords()) {
		t.Fatalf("compressed simulation differs: %+v vs %+v", compressed, plain)
	}
	if err = graph.ReplayCounterexample(plain.Counterexample); err != nil {
		t.Fatalf("counterexample does not replay: %v", err)
	}
	initial, _ := plain.Counterexample.Initial(0)
	if initial {
		t.Fatal("FSIM initial registers must be zero")
	}
	parameters.CheckMiter = false
	withoutCheck, err := graph.FsimSimulate(parameters, FsimBuildOptions{Compress: true}, FsimRunOptions{Seed: 17})
	if err != nil || withoutCheck.Counterexample != nil || withoutCheck.Frames != parameters.Iterations || withoutCheck.Property != -1 {
		t.Fatalf("unchecked campaign: %+v %v", withoutCheck, err)
	}
}

func TestFsimIndependentModelCampaign(t *testing.T) {
	graph := fsimTestGraph(t, true)
	parameters := FsimSimulationParams{Words: 1, Iterations: 3, CheckMiter: true}
	for execution := 0; execution < 5000; execution++ {
		result, err := graph.FsimSimulate(parameters, FsimBuildOptions{Compress: execution&1 != 0}, FsimRunOptions{Seed: uint64(execution + 1)})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		if result.Counterexample == nil {
			t.Fatalf("execution %d did not find the two-frame latch witness", execution)
		}
		if err = graph.ReplayCounterexample(result.Counterexample); err != nil {
			t.Fatalf("execution %d replay: %v", execution, err)
		}
		for frame := 0; frame < result.Counterexample.Frames(); frame++ {
			x, _ := result.Counterexample.Input(frame, 0)
			y, _ := result.Counterexample.Input(frame, 1)
			if frame+1 == result.Counterexample.Frames() {
				previousX, _ := result.Counterexample.Input(frame-1, 0)
				previousY, _ := result.Counterexample.Input(frame-1, 1)
				if !previousX || previousY {
					t.Fatalf("execution %d model mismatch: current %v/%v previous %v/%v", execution, x, y, previousX, previousY)
				}
			}
		}
	}
}

func TestFsimErrorsLimitsCancellationAndAtomicity(t *testing.T) {
	graph := fsimTestGraph(t, false)
	if manager, err := NewFsimManager(nil, FsimBuildOptions{}); manager != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil graph: %v %v", manager, err)
	}
	if manager, err := NewFsimManager(graph, FsimBuildOptions{MaxObjects: 2}); manager != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("object limit: %v %v", manager, err)
	}
	if manager, err := NewFsimManager(graph, FsimBuildOptions{MaxFrontier: 1}); manager != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("frontier limit: %v %v", manager, err)
	}
	if manager, err := NewFsimManager(graph, FsimBuildOptions{Compress: true, MaxEncodedBytes: 1}); manager != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("stream limit: %v %v", manager, err)
	}
	compressedParameters := FsimSimulationParams{Words: 1, Iterations: 1, CompressAIG: true}
	if result, compressErr := graph.FsimSimulate(compressedParameters, FsimBuildOptions{MaxEncodedBytes: 1}, FsimRunOptions{}); result != (FsimResult{}) || !errors.Is(compressErr, ErrLimit) {
		t.Fatalf("parameter compression request: %+v %v", result, compressErr)
	}
	if manager, err := NewFsimManager(graph, FsimBuildOptions{Stop: func() bool { return true }}); manager != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("build cancellation: %v %v", manager, err)
	}
	manager, err := NewFsimManager(graph, FsimBuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	before := manager.Objects()
	polls := 0
	if result, simulateErr := manager.Simulate(FsimSimulationParams{Words: 64, Iterations: 1}, FsimRunOptions{Stop: func() bool {
		polls++
		return polls == 20
	}}); result != (FsimResult{}) || !errors.Is(simulateErr, ErrCancelled) {
		t.Fatalf("inner-loop cancellation: %+v %v after %d polls", result, simulateErr, polls)
	}
	for _, test := range []struct {
		parameters FsimSimulationParams
		options    FsimRunOptions
		want       error
	}{
		{FsimSimulationParams{}, FsimRunOptions{}, ErrDimension},
		{FsimSimulationParams{Words: -1}, FsimRunOptions{}, ErrDimension},
		{FsimSimulationParams{Words: 2, Iterations: 1}, FsimRunOptions{MaxWords: 1}, ErrLimit},
		{FsimSimulationParams{Words: 1, Iterations: 2}, FsimRunOptions{MaxFrames: 1}, ErrLimit},
		{FsimSimulationParams{Words: 1, Iterations: 1}, FsimRunOptions{MaxBytes: 1}, ErrLimit},
		{FsimSimulationParams{Words: 1, Iterations: 1}, FsimRunOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, simulateErr := manager.Simulate(test.parameters, test.options); result != (FsimResult{}) || !errors.Is(simulateErr, test.want) {
			t.Fatalf("simulate error: %+v %v want %v", result, simulateErr, test.want)
		}
		if !slices.Equal(before, manager.Objects()) {
			t.Fatal("failed simulation mutated manager")
		}
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	_, _ = buffered.CreateCO(buffer)
	if manager, err = NewFsimManager(buffered, FsimBuildOptions{}); manager != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("buffer handling: %v %v", manager, err)
	}
}

func FuzzFsimObjectCodec(f *testing.F) {
	f.Add([]byte{1}, uint8(1))
	f.Add([]byte{5, 9, 3, 2, 5}, uint8(3))
	f.Fuzz(func(t *testing.T, data []byte, rawCount uint8) {
		if len(data) > 256 {
			return
		}
		count := int(rawCount % 32)
		objects, err := DecodeFsimObjects(data, count)
		if err != nil {
			return
		}
		encoded, err := EncodeFsimObjects(objects, 1024)
		if err != nil {
			t.Fatalf("decoded objects did not re-encode: %#v: %v", objects, err)
		}
		if !slices.Equal(encoded, data) {
			t.Fatalf("noncanonical successful decode: %x != %x", encoded, data)
		}
	})
}

func FuzzFsimSequentialModel(f *testing.F) {
	f.Add(uint64(1), true)
	f.Add(uint64(17), false)
	f.Fuzz(func(t *testing.T, seed uint64, invert bool) {
		graph := fsimSeedGraph(t, seed, invert)
		parameters := FsimSimulationParams{Words: 1, Iterations: 3, CheckMiter: true}
		plainManager, err := NewFsimManager(graph, FsimBuildOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantFrames, wantProperty, wantPattern, wantFound := fsimScalarExpected(t, plainManager, parameters, seed)
		plain, err := plainManager.Simulate(parameters, FsimRunOptions{Seed: seed, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		compressed, err := graph.FsimSimulate(parameters, FsimBuildOptions{Compress: true}, FsimRunOptions{Seed: seed, MaxBytes: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		if plain.Frames != compressed.Frames || plain.Property != compressed.Property || plain.Pattern != compressed.Pattern {
			t.Fatalf("representation mismatch: %+v %+v", plain, compressed)
		}
		if plain.Frames != wantFrames || plain.Property != wantProperty || plain.Pattern != wantPattern || (plain.Counterexample != nil) != wantFound {
			t.Fatalf("scalar mismatch: got frames/property/pattern/found=%d/%d/%d/%v want %d/%d/%d/%v", plain.Frames, plain.Property, plain.Pattern, plain.Counterexample != nil, wantFrames, wantProperty, wantPattern, wantFound)
		}
		if plain.Counterexample != nil {
			if err = graph.ReplayCounterexample(plain.Counterexample); err != nil {
				t.Fatal(err)
			}
		}
	})
}
