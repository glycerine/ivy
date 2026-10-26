package abc

// Go-only merged-boundary, timing, packing, lifecycle, and fuzz checks for src/aig/gia/giaIff.c.

import (
	"errors"
	"math"
	"math/rand"
	"slices"
	"testing"
)

type giaIffFixtureData struct {
	manager     *GiaManager
	mapping     *GiaLUTMapping
	a, b, c, d  GiaLit
	left, right GiaLit
	root        GiaLit
}

func giaIffFixture(t testing.TB) giaIffFixtureData {
	t.Helper()
	manager := NewGiaManager()
	a, _ := manager.AppendCI()
	b, _ := manager.AppendCI()
	c, _ := manager.AppendCI()
	d, _ := manager.AppendCI()
	left, _ := manager.AppendAnd(a, b)
	right, _ := manager.AppendAnd(c, d)
	root, _ := manager.AppendAnd(left, right)
	manager.AppendCO(root)
	mapping, err := NewGiaLUTMapping(manager.ObjectCount(), []GiaLUTRecord{
		{Root: left.ID(), Fanins: []GiaID{a.ID(), b.ID()}, MuxID: int(left.ID())},
		{Root: right.ID(), Fanins: []GiaID{c.ID(), d.ID()}, MuxID: int(right.ID())},
		{Root: root.ID(), Fanins: []GiaID{left.ID(), right.ID()}, MuxID: int(root.ID())},
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return giaIffFixtureData{manager, mapping, a, b, c, d, left, right, root}
}

func giaIffTiming(t testing.TB, manager *GiaManager, arrivals []float32) *TimManager {
	t.Helper()
	timing, err := NewTimManager(manager.InputCount(), manager.OutputCount(), 0)
	if err != nil {
		t.Fatal(err)
	}
	for index, arrival := range arrivals {
		if err = timing.InitPIArrival(index, arrival); err != nil {
			t.Fatal(err)
		}
	}
	return timing
}

func TestGiaIffBoundaryCandidatesAndTriplePacking(t *testing.T) {
	fixture := giaIffFixture(t)
	timing := giaIffTiming(t, fixture.manager, []float32{0, 0, 0, 0})
	library := GiaIffLUTLibrary{DelayByInputs: []float32{0, 1, 1, 0.5, 0.25}}
	session, err := NewGiaIffSession(fixture.manager, fixture.mapping, library, timing, GiaIffOptions{LUTSize: 4, Degree: 3})
	if err != nil {
		t.Fatal(err)
	}
	boundary, err := session.Boundary(fixture.root.ID(), int(fixture.left.ID()), -1)
	if err != nil || !slices.Equal(boundary, []GiaID{fixture.right.ID(), fixture.a.ID(), fixture.b.ID()}) {
		t.Fatal(boundary, err)
	}
	boundary, err = session.Boundary(fixture.root.ID(), int(fixture.left.ID()), int(fixture.right.ID()))
	if err != nil || !slices.Equal(boundary, []GiaID{fixture.a.ID(), fixture.b.ID(), fixture.c.ID(), fixture.d.ID()}) {
		t.Fatal(boundary, err)
	}
	result, err := session.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Single != 2 || result.Pair != 0 || result.Triple != 1 || result.MaximumDelay != 0.25 || result.Arrivals[fixture.root.ID()] != 0.25 || result.Match2[fixture.root.ID()] != int(fixture.left.ID()) || result.Match3[fixture.root.ID()] != int(fixture.right.ID()) {
		t.Fatal(result)
	}
	groups := result.Packing.Groups()
	if len(groups) != 1 || !slices.Equal(groups[0], []GiaID{fixture.left.ID(), fixture.right.ID(), fixture.root.ID()}) {
		t.Fatal(groups)
	}
	co, err := result.Timing.CO(0)
	if err != nil || co.Arrival != 0.25 {
		t.Fatal(co, err)
	}
	again, err := session.Run()
	if err != nil || !slices.Equal(again.Arrivals, result.Arrivals) {
		t.Fatal(again, err)
	}
	again.Arrivals[fixture.root.ID()] = 99
	again.Packing.packed[0] = 99
	third, _ := session.Run()
	if third.Arrivals[fixture.root.ID()] != 0.25 || third.Packing.packed[0] != 1 {
		t.Fatal("results alias session")
	}
	automaticDelays := make([]float32, 11)
	copy(automaticDelays, library.DelayByInputs)
	automatic, automaticErr := BuildGiaIffPacking(fixture.manager, fixture.mapping, GiaIffLUTLibrary{DelayByInputs: automaticDelays}, timing, GiaIffOptions{})
	if automaticErr != nil || automatic.Triple != 1 || automatic.MaximumDelay != 0.25 {
		t.Fatal(automatic, automaticErr)
	}
	session.Close()
	session.Close()
	if _, err = session.Run(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestGiaIffPairSelectionOrderAndInputOwnership(t *testing.T) {
	fixture := giaIffFixture(t)
	timing := giaIffTiming(t, fixture.manager, []float32{0, 0, 0, 0})
	library := GiaIffLUTLibrary{DelayByInputs: []float32{0, 1, 1, 0.5, 0.25}}
	beforeObjects := slices.Clone(fixture.manager.objects)
	beforePacked := fixture.mapping.Packed()
	result, err := BuildGiaIffPacking(fixture.manager, fixture.mapping, library, timing, GiaIffOptions{LUTSize: 4, Degree: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Single != 2 || result.Pair != 1 || result.Triple != 0 || result.MaximumDelay != 1.5 || result.Match2[fixture.root.ID()] != int(fixture.left.ID()) || result.Match3[fixture.root.ID()] != -1 {
		t.Fatal(result)
	}
	groups := result.Packing.Groups()
	if len(groups) != 2 || !slices.Equal(groups[0], []GiaID{fixture.right.ID()}) || !slices.Equal(groups[1], []GiaID{fixture.left.ID(), fixture.root.ID()}) {
		t.Fatal(groups)
	}
	if !slices.Equal(beforeObjects, fixture.manager.objects) || !slices.Equal(beforePacked, fixture.mapping.Packed()) {
		t.Fatal("source graph or mapping mutated")
	}
	ci, _ := timing.CI(0)
	co, _ := timing.CO(0)
	if ci.TravID != 0 || co.Arrival != 0 {
		t.Fatal("source timing mutated", ci, co)
	}
	degreeThreeLibrary := GiaIffLUTLibrary{DelayByInputs: []float32{0, 1, 1, 0.5, 2}}
	degreeThree, err := BuildGiaIffPacking(fixture.manager, fixture.mapping, degreeThreeLibrary, timing, GiaIffOptions{LUTSize: 4, Degree: 3})
	if err != nil || degreeThree.Pair != 1 || degreeThree.Triple != 0 || degreeThree.Match2[fixture.root.ID()] != int(fixture.left.ID()) || degreeThree.Match3[fixture.root.ID()] != -1 {
		t.Fatal(degreeThree, err)
	}
}

func independentGiaIffBoundary(mapping *GiaLUTMapping, objects int, root GiaID, skip2, skip3 int) []GiaID {
	record, _ := mapping.Record(root)
	seen := make([]bool, objects)
	result := make([]GiaID, 0)
	add := func(id GiaID) {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	for _, fanin := range record.Fanins {
		if int(fanin) != skip2 && int(fanin) != skip3 {
			add(fanin)
		}
	}
	if skip2 >= 0 {
		child, _ := mapping.Record(GiaID(skip2))
		for _, fanin := range child.Fanins {
			if int(fanin) != skip3 {
				add(fanin)
			}
		}
	}
	if skip3 >= 0 {
		child, _ := mapping.Record(GiaID(skip3))
		for _, fanin := range child.Fanins {
			if int(fanin) != skip2 {
				add(fanin)
			}
		}
	}
	return result
}

func independentGiaIffTimes(manager *GiaManager, mapping *GiaLUTMapping, delays []float32, inputs []float32, degree int) ([]float32, []int, []int) {
	times := make([]float32, manager.ObjectCount())
	for index := range times {
		times[index] = float32(-math.MaxFloat32)
	}
	times[0] = 0
	for index, id := range manager.cis {
		times[id] = inputs[index]
	}
	match2, match3 := make([]int, len(times)), make([]int, len(times))
	for index := range match2 {
		match2[index], match3[index] = -1, -1
	}
	value := func(boundary []GiaID) float32 {
		best := float32(-math.MaxFloat32)
		for _, id := range boundary {
			best = max(best, times[id])
		}
		return best + delays[len(boundary)]
	}
	for id, object := range manager.objects {
		record, mapped := mapping.Record(GiaID(id))
		if mapped {
			best := value(independentGiaIffBoundary(mapping, len(times), GiaID(id), -1, -1))
			for _, child := range record.Fanins {
				if child == 0 || manager.objects[child].IsCI() {
					continue
				}
				candidate := value(independentGiaIffBoundary(mapping, len(times), GiaID(id), int(child), -1))
				if best > candidate {
					best, match2[id] = candidate, int(child)
				}
			}
			if degree == 3 {
				for _, left := range record.Fanins {
					for _, right := range record.Fanins {
						if left == right || left == 0 || right == 0 || manager.objects[left].IsCI() || manager.objects[right].IsCI() {
							continue
						}
						candidate := value(independentGiaIffBoundary(mapping, len(times), GiaID(id), int(left), int(right)))
						if best > candidate {
							best, match2[id], match3[id] = candidate, int(left), int(right)
						}
					}
				}
			}
			times[id] = best
		} else if object.IsCO() {
			driver, _ := object.Fanin0(GiaID(id))
			times[id] = times[driver.ID()]
		}
	}
	return times, match2, match3
}

func TestGiaIffErrorsCancellationAndDeep(t *testing.T) {
	fixture := giaIffFixture(t)
	timing := giaIffTiming(t, fixture.manager, []float32{0, 0, 0, 0})
	library := GiaIffLUTLibrary{DelayByInputs: make([]float32, 11)}
	for index := 1; index < len(library.DelayByInputs); index++ {
		library.DelayByInputs[index] = 1
	}
	for _, test := range []struct {
		manager *GiaManager
		mapping *GiaLUTMapping
		timing  *TimManager
		library GiaIffLUTLibrary
		options GiaIffOptions
	}{{nil, fixture.mapping, timing, library, GiaIffOptions{}}, {fixture.manager, nil, timing, library, GiaIffOptions{}}, {fixture.manager, fixture.mapping, nil, library, GiaIffOptions{}}, {fixture.manager, fixture.mapping, timing, GiaIffLUTLibrary{}, GiaIffOptions{}}, {fixture.manager, fixture.mapping, timing, library, GiaIffOptions{Degree: 1}}, {fixture.manager, fixture.mapping, timing, library, GiaIffOptions{MaxObjects: 1}}, {fixture.manager, fixture.mapping, timing, library, GiaIffOptions{MaxMappingEntries: 1}}} {
		if session, err := NewGiaIffSession(test.manager, test.mapping, test.library, test.timing, test.options); err == nil || session != nil {
			t.Fatal(test.options, session, err)
		}
	}
	wrongTiming, _ := NewTimManager(1, 1, 0)
	if session, err := NewGiaIffSession(fixture.manager, fixture.mapping, library, wrongTiming, GiaIffOptions{}); err == nil || session != nil {
		t.Fatal(session, err)
	}
	for threshold := 1; threshold <= 40; threshold++ {
		calls := 0
		_, err := BuildGiaIffPacking(fixture.manager, fixture.mapping, library, timing, GiaIffOptions{LUTSize: 4, Degree: 3, Stop: func() bool { calls++; return calls >= threshold }})
		if err != nil && !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, err)
		}
	}
	if _, err := BuildGiaIffPacking(fixture.manager, fixture.mapping, library, timing, GiaIffOptions{LUTSize: 4, Degree: 2, MaxGroups: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := BuildGiaIffPacking(fixture.manager, fixture.mapping, library, timing, GiaIffOptions{LUTSize: 4, Degree: 2, MaxVisits: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}

	deep := NewGiaManager()
	a, _ := deep.AppendCI()
	b, _ := deep.AppendCI()
	chain, _ := deep.AppendAnd(a, b)
	records := []GiaLUTRecord{{Root: chain.ID(), Fanins: []GiaID{a.ID(), b.ID()}, MuxID: int(chain.ID())}}
	for index := 0; index < 10000; index++ {
		previous := chain
		chain, _ = deep.AppendAnd(chain, a)
		records = append(records, GiaLUTRecord{Root: chain.ID(), Fanins: []GiaID{previous.ID(), a.ID()}, MuxID: int(chain.ID())})
	}
	deep.AppendCO(chain)
	deepMapping, err := NewGiaLUTMapping(deep.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	deepTiming := giaIffTiming(t, deep, []float32{0, 0})
	result, err := BuildGiaIffPacking(deep, deepMapping, GiaIffLUTLibrary{DelayByInputs: []float32{0, 1, 1, 1}}, deepTiming, GiaIffOptions{LUTSize: 2, Degree: 2})
	if err != nil || result.MaximumDelay <= 0 || len(result.Packing.Groups()) == 0 {
		t.Fatal(result.MaximumDelay, err)
	}
}

func FuzzGiaIffTimingAndPacking(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(0x9e3779b97f4a7c15))
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		manager := NewGiaManager()
		inputs := 2 + random.Intn(4)
		signals := make([]GiaLit, inputs)
		arrivals := make([]float32, inputs)
		for index := range signals {
			signals[index], _ = manager.AppendCI()
			arrivals[index] = float32(random.Intn(5)) / 2
		}
		records := make([]GiaLUTRecord, 0)
		for index := 0; index < 1+random.Intn(10); index++ {
			left := signals[random.Intn(len(signals))]
			right := signals[random.Intn(len(signals))]
			if left.ID() == right.ID() {
				continue
			}
			gate, err := manager.AppendAnd(left.NotIf(random.Intn(2) != 0), right.NotIf(random.Intn(2) != 0))
			if err != nil {
				continue
			}
			records = append(records, GiaLUTRecord{Root: gate.ID(), Fanins: []GiaID{left.ID(), right.ID()}, MuxID: int(gate.ID())})
			signals = append(signals, gate)
		}
		if len(records) == 0 {
			gate, _ := manager.AppendAnd(signals[0], signals[1])
			records = append(records, GiaLUTRecord{Root: gate.ID(), Fanins: []GiaID{signals[0].ID(), signals[1].ID()}, MuxID: int(gate.ID())})
			signals = append(signals, gate)
		}
		manager.AppendCO(signals[len(signals)-1])
		mapping, err := NewGiaLUTMapping(manager.ObjectCount(), records, 4096)
		if err != nil {
			t.Fatal(seed, err)
		}
		timing := giaIffTiming(t, manager, arrivals)
		delays := make([]float32, 17)
		for size := 1; size < len(delays); size++ {
			delays[size] = 0.5 + float32(size%4)/4
		}
		degree := 2 + random.Intn(2)
		result, err := BuildGiaIffPacking(manager, mapping, GiaIffLUTLibrary{DelayByInputs: delays}, timing, GiaIffOptions{LUTSize: 4, Degree: degree, MaxObjects: 4096, MaxMappingEntries: 16384, MaxGroups: 4096, MaxVisits: 16384})
		if err != nil {
			t.Fatal(seed, err)
		}
		wantTimes, want2, want3 := independentGiaIffTimes(manager, mapping, delays, arrivals, degree)
		if !slices.Equal(result.Arrivals, wantTimes) || !slices.Equal(result.Match2, want2) || !slices.Equal(result.Match3, want3) {
			t.Fatal(seed, degree, result.Arrivals, wantTimes, result.Match2, want2, result.Match3, want3)
		}
		if _, err := GiaLUTPackingFromPacked(mapping, result.Packing.Packed(), 16384); err != nil {
			t.Fatal(seed, err)
		}
	})
}
