package abc

// Standalone semantic, sequential, limit, concurrency and fuzz tests for saigDual.c.

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

type saigTestRail struct{ zero, one bool }

func saigTestRailNot(value saigTestRail) saigTestRail {
	care := value.zero || value.one
	return saigTestRail{value.one && care, value.zero && care}
}

func saigTestRailFanin(values []saigTestRail, literal AigLit) saigTestRail {
	value := values[literal.ID()]
	if literal.Complemented() {
		return saigTestRailNot(value)
	}
	return value
}

func saigDualExpected(source *AigManager, raw []bool, options SAIGDualRailOptions) []bool {
	values := make([]saigTestRail, source.ObjectCount())
	values[0] = saigTestRail{false, true}
	primaryInputs := source.PrimaryInputCount()
	rawIndex := 0
	for index, id := range source.cis {
		var zeroRaw, oneRaw bool
		switch {
		case index < options.DualPrimaryInputs:
			zeroRaw, oneRaw = raw[rawIndex], raw[rawIndex+1]
			rawIndex += 2
		case index < primaryInputs:
			oneRaw = raw[rawIndex]
			zeroRaw = !oneRaw
			rawIndex++
		default:
			zeroRaw, oneRaw = raw[rawIndex], raw[rawIndex+1]
			rawIndex += 2
			dual := options.DualRegisters
			if options.DCRegisters != nil {
				dual = options.DCRegisters[index-primaryInputs]
			}
			if !dual {
				zeroRaw = !zeroRaw
			}
		}
		values[id] = saigTestRail{zeroRaw && !oneRaw, oneRaw && !zeroRaw}
	}
	for id, object := range source.objects {
		switch object.Type {
		case AigBuffer, AigCO:
			values[id] = saigTestRailFanin(values, object.Fanin0)
		case AigAND:
			left := saigTestRailFanin(values, object.Fanin0)
			right := saigTestRailFanin(values, object.Fanin1)
			values[id] = saigTestRail{left.zero || right.zero, left.one && right.one}
		}
	}
	start, end := 0, source.PrimaryOutputCount()
	if options.MiterRegisters {
		start, end = source.PrimaryOutputCount(), source.OutputCount()
	}
	violation := false
	for _, outputID := range source.cos[start:end] {
		value := values[outputID]
		switch {
		case options.CheckZero:
			violation = violation || value.zero && !value.one
		case options.CheckOne:
			violation = violation || !value.zero && value.one
		default:
			violation = violation || !value.zero && !value.one
		}
	}
	if options.ComplementOutput {
		violation = !violation
	}
	result := []bool{violation}
	for index, outputID := range source.cos[source.PrimaryOutputCount():] {
		value := values[outputID]
		dual := options.DualRegisters
		if options.DCRegisters != nil {
			dual = options.DCRegisters[index]
		}
		result = append(result, value.zero != !dual, value.one)
	}
	return result
}

func saigDualFixture(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	state, _ := graph.CreateCI()
	left, err := graph.And(a.Not(), state)
	if err != nil {
		t.Fatal(err)
	}
	property, _ := graph.Or(left, b)
	next, _ := graph.Xor(a, state)
	graph.CreateCO(property)
	graph.CreateCO(next)
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	graph.Name, graph.Spec = "dual", "source-only.aig"
	return graph
}

func TestSAIGDuplicateDualExhaustiveTernarySemantics(t *testing.T) {
	source := saigDualFixture(t)
	before := source.Clone()
	for dualInputs := 0; dualInputs <= source.PrimaryInputCount(); dualInputs++ {
		for dualRegisters := 0; dualRegisters < 2; dualRegisters++ {
			for miterRegisters := 0; miterRegisters < 2; miterRegisters++ {
				for complement := 0; complement < 2; complement++ {
					for mode := 0; mode < 4; mode++ {
						for dcMode := 0; dcMode < 3; dcMode++ {
							options := SAIGDualRailOptions{
								DualPrimaryInputs: dualInputs,
								DualRegisters:     dualRegisters != 0,
								MiterRegisters:    miterRegisters != 0,
								ComplementOutput:  complement != 0,
								CheckZero:         mode&1 != 0,
								CheckOne:          mode&2 != 0,
							}
							if dcMode != 0 {
								options.DCRegisters = []bool{dcMode == 2}
							}
							result, err := SAIGDuplicateDual(source, options)
							if err != nil {
								t.Fatal(options, err)
							}
							if result.RegisterCount() != 2 || result.PrimaryInputCount() != 2+dualInputs || result.PrimaryOutputCount() != 1 || result.Name != source.Name || result.Spec != "" {
								t.Fatalf("options=%+v partition=%d/%d/%d name/spec=%q/%q", options, result.PrimaryInputCount(), result.PrimaryOutputCount(), result.RegisterCount(), result.Name, result.Spec)
							}
							for assignment := 0; assignment < 1<<result.InputCount(); assignment++ {
								raw := make([]bool, result.InputCount())
								for index := range raw {
									raw[index] = assignment&(1<<index) != 0
								}
								got, evalErr := result.Evaluate(raw)
								want := saigDualExpected(source, raw, options)
								if evalErr != nil || !reflect.DeepEqual(got, want) {
									t.Fatalf("options=%+v assignment=%d got=%v want=%v err=%v", options, assignment, got, want, evalErr)
								}
							}
						}
					}
				}
			}
		}
	}
	if !reflect.DeepEqual(source.objects, before.objects) || !reflect.DeepEqual(source.cis, before.cis) || !reflect.DeepEqual(source.cos, before.cos) {
		t.Fatal("dual construction mutated source")
	}
}

func saigBlockFixture(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	state0, _ := graph.CreateCI()
	state1, _ := graph.CreateCI()
	property0, _ := graph.And(input, state0)
	property1, _ := graph.Or(input.Not(), state1)
	next0, _ := graph.Xor(input, state1)
	next1, _ := graph.And(state0.Not(), state1)
	graph.CreateCO(property0)
	graph.CreateCO(property1)
	graph.CreateCO(next0)
	graph.CreateCO(next1)
	if err := graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	graph.Name, graph.Spec = "block", "block.aig"
	return graph
}

func saigBlockExpected(t testing.TB, source *AigManager, cycles int, raw []bool) []bool {
	t.Helper()
	sourceValues, err := source.Evaluate(raw[:source.InputCount()])
	if err != nil {
		t.Fatal(err)
	}
	delays := raw[source.InputCount():]
	condition := true
	for _, value := range delays {
		condition = condition && value
	}
	drivers := append([]bool(nil), sourceValues...)
	drivers = append(drivers, true)
	if cycles > 1 {
		drivers = append(drivers, delays[:cycles-1]...)
	}
	for index := 0; index < source.PrimaryOutputCount()+cycles; index++ {
		drivers[index] = drivers[index] && condition
	}
	return drivers
}

func TestSAIGBlockPropertyCyclesNativeBoundaryAndOwnership(t *testing.T) {
	source := saigBlockFixture(t)
	before := source.Clone()
	for cycles := 1; cycles <= 4; cycles++ {
		result, err := SAIGBlockPropertyCycles(source, cycles, SAIGBlockPropertyOptions{})
		if err != nil {
			t.Fatal(cycles, err)
		}
		if result == source || result.Name != source.Name || result.Spec != source.Spec || result.PrimaryInputCount() != source.PrimaryInputCount() || result.PrimaryOutputCount() != source.PrimaryOutputCount() || result.RegisterCount() != source.RegisterCount()+cycles {
			t.Fatalf("cycles=%d partition=%d/%d/%d", cycles, result.PrimaryInputCount(), result.PrimaryOutputCount(), result.RegisterCount())
		}
		for assignment := 0; assignment < 1<<result.InputCount(); assignment++ {
			raw := make([]bool, result.InputCount())
			for index := range raw {
				raw[index] = assignment&(1<<index) != 0
			}
			got, evalErr := result.Evaluate(raw)
			want := saigBlockExpected(t, source, cycles, raw)
			if evalErr != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("cycles=%d assignment=%d got=%v want=%v err=%v", cycles, assignment, got, want, evalErr)
			}
		}
	}
	if !reflect.DeepEqual(source.objects, before.objects) || !reflect.DeepEqual(source.cis, before.cis) || !reflect.DeepEqual(source.cos, before.cos) {
		t.Fatal("blocking transformed source")
	}
}

func TestSAIGDualLimitsDepthAndConcurrency(t *testing.T) {
	source := saigDualFixture(t)
	if result, err := SAIGDuplicateDual(nil, SAIGDualRailOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil result=%v err=%v", result, err)
	}
	for _, options := range []SAIGDualRailOptions{
		{DualPrimaryInputs: -1}, {DualPrimaryInputs: 3}, {DCRegisters: []bool{true, false}}, {MaxObjects: -1}, {MaxWork: -1},
	} {
		if result, err := SAIGDuplicateDual(source, options); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("options=%+v result=%v err=%v", options, result, err)
		}
	}
	for _, options := range []SAIGDualRailOptions{{MaxObjects: 1}, {MaxWork: 1}, {Stop: func() bool { return true }}} {
		if result, err := SAIGDuplicateDual(source, options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("options=%+v result=%v err=%v", options, result, err)
		}
	}
	empty := NewAigManager()
	if result, err := SAIGDuplicateDual(empty, SAIGDualRailOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("empty result=%v err=%v", result, err)
	}
	if result, err := SAIGBlockPropertyCycles(nil, 1, SAIGBlockPropertyOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("nil block result=%v err=%v", result, err)
	}
	for _, row := range []struct {
		cycles  int
		options SAIGBlockPropertyOptions
	}{
		{0, SAIGBlockPropertyOptions{}}, {1, SAIGBlockPropertyOptions{MaxObjects: -1}}, {1, SAIGBlockPropertyOptions{MaxWork: -1}},
	} {
		if result, err := SAIGBlockPropertyCycles(source, row.cycles, row.options); result != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("block row=%+v result=%v err=%v", row, result, err)
		}
	}
	for _, options := range []SAIGBlockPropertyOptions{{MaxObjects: 1}, {MaxWork: 1}, {Stop: func() bool { return true }}} {
		if result, err := SAIGBlockPropertyCycles(source, 2, options); result != nil || !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("block options=%+v result=%v err=%v", options, result, err)
		}
	}

	deep := NewAigManager()
	root, _ := deep.CreateCI()
	for index := 0; index < 10000; index++ {
		input, _ := deep.CreateCI()
		root, _ = deep.And(root.NotIf(index&1 != 0), input.NotIf(index&2 != 0))
	}
	deep.CreateCO(root)
	if deep.ObjectCount() < 20000 {
		t.Fatalf("deep fixture collapsed to %d", deep.ObjectCount())
	}
	if result, err := SAIGDuplicateDual(deep, SAIGDualRailOptions{MaxObjects: 100000, MaxWork: 1000000}); err != nil || result.ObjectCount() < 10000 {
		t.Fatalf("deep result=%v err=%v", result, err)
	}

	baseline, err := SAIGDuplicateDual(source, SAIGDualRailOptions{DualPrimaryInputs: 2, DualRegisters: true})
	if err != nil {
		t.Fatal(err)
	}
	baselineOutputs, _ := baseline.Evaluate(make([]bool, baseline.InputCount()))
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			dual, dualErr := SAIGDuplicateDual(source, SAIGDualRailOptions{DualPrimaryInputs: 2, DualRegisters: true})
			blocked, blockErr := SAIGBlockPropertyCycles(source, 2, SAIGBlockPropertyOptions{})
			if dualErr != nil || blockErr != nil || blocked.RegisterCount() != 3 {
				failures.Add(1)
				return
			}
			outputs, evalErr := dual.Evaluate(make([]bool, dual.InputCount()))
			if evalErr != nil || !reflect.DeepEqual(outputs, baselineOutputs) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzSAIGDualTransforms(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{0xff, 0, 17})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		at := func(index int) byte { return data[index%len(data)] }
		source := NewAigManager()
		a, _ := source.CreateCI()
		b, _ := source.CreateCI()
		state, _ := source.CreateCI()
		pool := []AigLit{a, b, state}
		gates := 1 + int(at(0)%4)
		for gate := 0; gate < gates; gate++ {
			left := pool[int(at(2*gate+1))%len(pool)].NotIf(at(gate+2)&1 != 0)
			right := pool[int(at(2*gate+2))%len(pool)].NotIf(at(gate+3)&2 != 0)
			root, err := source.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, root)
		}
		source.CreateCO(pool[len(pool)-1].NotIf(at(1)&4 != 0))
		source.CreateCO(pool[int(at(2))%len(pool)].NotIf(at(3)&8 != 0))
		if err := source.SetRegisterCount(1); err != nil {
			t.Fatal(err)
		}
		options := SAIGDualRailOptions{
			DualPrimaryInputs: int(at(4) % 3),
			DualRegisters:     at(5)&1 != 0,
			MiterRegisters:    at(6)&2 != 0,
			ComplementOutput:  at(7)&4 != 0,
			CheckZero:         at(8)&8 != 0,
			CheckOne:          at(9)&16 != 0,
		}
		if at(10)&1 != 0 {
			options.DCRegisters = []bool{at(11)&1 != 0}
		}
		dual, err := SAIGDuplicateDual(source, options)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<dual.InputCount(); assignment++ {
			raw := make([]bool, dual.InputCount())
			for index := range raw {
				raw[index] = assignment&(1<<index) != 0
			}
			got, evalErr := dual.Evaluate(raw)
			want := saigDualExpected(source, raw, options)
			if evalErr != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("dual assignment=%d options=%+v got=%v want=%v err=%v", assignment, options, got, want, evalErr)
			}
		}
		cycles := 1 + int(at(12)%3)
		blocked, err := SAIGBlockPropertyCycles(source, cycles, SAIGBlockPropertyOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<blocked.InputCount(); assignment++ {
			raw := make([]bool, blocked.InputCount())
			for index := range raw {
				raw[index] = assignment&(1<<index) != 0
			}
			got, evalErr := blocked.Evaluate(raw)
			want := saigBlockExpected(t, source, cycles, raw)
			if evalErr != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("block assignment=%d cycles=%d got=%v want=%v err=%v", assignment, cycles, got, want, evalErr)
			}
		}
	})
}
