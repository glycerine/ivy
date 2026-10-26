package abc

// Tests for the structural SAIG port from src/aig/saig.

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

func saigStructuralFixture(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	graph.Name = "saig-structural"
	primary0, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	primary1, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	latch0, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	latch1, err := graph.CreateCI()
	if err != nil {
		t.Fatal(err)
	}
	property0, err := graph.Xor(primary0, latch0)
	if err != nil {
		t.Fatal(err)
	}
	property1, err := graph.And(primary1, latch1)
	if err != nil {
		t.Fatal(err)
	}
	next0, err := graph.Xor(primary0, latch1)
	if err != nil {
		t.Fatal(err)
	}
	next1, err := graph.And(latch0, primary1)
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range []AigLit{property0, property1, next0, next1} {
		if _, err = graph.CreateCO(output); err != nil {
			t.Fatal(err)
		}
	}
	if err = graph.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	if err = graph.Check(); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestSaigTerminalPartitionsAndConeProfile(t *testing.T) {
	graph := saigStructuralFixture(t)
	if got, want := len(graph.PrimaryInputs()), 2; got != want {
		t.Fatalf("primary inputs %d want %d", got, want)
	}
	if got, want := len(graph.LatchOutputs()), 2; got != want {
		t.Fatalf("latch outputs %d want %d", got, want)
	}
	if got, want := len(graph.PrimaryOutputs()), 2; got != want {
		t.Fatalf("primary outputs %d want %d", got, want)
	}
	if got, want := len(graph.LatchInputs()), 2; got != want {
		t.Fatalf("latch inputs %d want %d", got, want)
	}
	lo := graph.LatchOutputs()[1]
	li, err := graph.LatchInputForOutput(lo)
	if err != nil || li != graph.LatchInputs()[1] {
		t.Fatalf("LO->LI got %d, %v", li, err)
	}
	back, err := graph.LatchOutputForInput(li)
	if err != nil || back != lo {
		t.Fatalf("LI->LO got %d, %v", back, err)
	}
	if graph.SequentialObjectRole(graph.PrimaryInputs()[0]) != SaigObjectPrimaryInput || graph.SequentialObjectRole(graph.PrimaryOutputs()[0]) != SaigObjectPrimaryOutput || graph.SequentialObjectRole(lo) != SaigObjectLatchOutput || graph.SequentialObjectRole(li) != SaigObjectLatchInput {
		t.Fatal("terminal role mismatch")
	}
	profile, err := graph.SequentialConeProfile(1, SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := []SaigConeStage{
		{Frame: 0, Total: 1, Current: 1, New: 1},
		{Frame: 1, Total: 2, Current: 1, New: 1},
		{Frame: 2, Total: 2, Current: 1, Older: 1},
	}
	if !reflect.DeepEqual(profile, want) {
		t.Fatalf("profile %#v want %#v", profile, want)
	}
	profiles, err := graph.SequentialConeProfiles(SaigStructuralOptions{})
	if err != nil || len(profiles) != graph.PrimaryOutputCount() || !reflect.DeepEqual(profiles[1], profile) {
		t.Fatalf("all profiles %v, %v", profiles, err)
	}
	if _, err = graph.SequentialConeProfile(2, SaigStructuralOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("bad output error %v", err)
	}
	if status, bddErr := graph.VerifyUsingBDDs(SaigBDDParameters{}); status != 0 || !errors.Is(bddErr, ErrUnsupported) {
		t.Fatalf("BDD disposition %d, %v", status, bddErr)
	}
}

func TestSaigDuplicateOrOutputsExhaustive(t *testing.T) {
	source := saigStructuralFixture(t)
	target, err := source.DuplicateOrPrimaryOutputs(SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if target.PrimaryOutputCount() != 1 || target.RegisterCount() != 2 || target.InputCount() != source.InputCount() {
		t.Fatalf("partition %d/%d/%d", target.PrimaryOutputCount(), target.RegisterCount(), target.InputCount())
	}
	for bits := 0; bits < 16; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0}
		state := []bool{bits&4 != 0, bits&8 != 0}
		sourceOutputs, sourceNext, sourceErr := source.Step(inputs, state)
		targetOutputs, targetNext, targetErr := target.Step(inputs, state)
		if sourceErr != nil || targetErr != nil {
			t.Fatalf("step errors %v %v", sourceErr, targetErr)
		}
		if targetOutputs[0] != (sourceOutputs[0] || sourceOutputs[1]) || !reflect.DeepEqual(targetNext, sourceNext) {
			t.Fatalf("bits %04b got %v/%v source %v/%v", bits, targetOutputs, targetNext, sourceOutputs, sourceNext)
		}
	}
	if source.Check() != nil {
		t.Fatal("source mutated")
	}
}

func TestSaigEquivalenceMiterAndTrim(t *testing.T) {
	source := saigStructuralFixture(t)
	pair := [][2]AigID{{source.PrimaryInputs()[0], 0}, {source.PrimaryInputs()[1], source.PrimaryInputs()[1]}}
	miter, err := source.CreateSequentialEquivalenceMiter(pair, true, SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if miter.PrimaryOutputCount() != 2 || miter.RegisterCount() != 2 {
		t.Fatalf("miter partition %d/%d", miter.PrimaryOutputCount(), miter.RegisterCount())
	}
	for bits := 0; bits < 16; bits++ {
		inputs := []bool{bits&1 != 0, bits&2 != 0}
		state := []bool{bits&4 != 0, bits&8 != 0}
		outputs, _, stepErr := miter.Step(inputs, state)
		if stepErr != nil {
			t.Fatal(stepErr)
		}
		if outputs[0] != inputs[0] || outputs[1] {
			t.Fatalf("miter bits %04b outputs %v", bits, outputs)
		}
	}

	trimSource := NewAigManager()
	unused, _ := trimSource.CreateCI()
	_ = unused
	used, _ := trimSource.CreateCI()
	state, _ := trimSource.CreateCI()
	node, _ := trimSource.And(used, state)
	_, _ = trimSource.CreateCO(node)
	_, _ = trimSource.CreateCO(used)
	if err = trimSource.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	trimmed, err := trimSource.TrimUnusedPrimaryInputs(SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := trimmed.OriginalCIs, []int{1, 2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("origin CIs %v want %v", got, want)
	}
	if trimmed.Graph.PrimaryInputCount() != 1 || trimmed.Graph.RegisterCount() != 1 {
		t.Fatalf("trim partition %d/%d", trimmed.Graph.PrimaryInputCount(), trimmed.Graph.RegisterCount())
	}
	allUnused := NewAigManager()
	_, _ = allUnused.CreateCI()
	_, _ = allUnused.CreateCI()
	_, _ = allUnused.CreateCO(AigConst0)
	all, err := allUnused.TrimUnusedPrimaryInputs(SaigStructuralOptions{})
	if err != nil || all.Graph.PrimaryInputCount() != 2 {
		t.Fatalf("all-unused preservation %d, %v", all.Graph.PrimaryInputCount(), err)
	}
}

func TestSaigAbstractionAndConeDuplication(t *testing.T) {
	source := saigStructuralFixture(t)
	abstracted, err := source.DuplicateSequentialAbstraction([]int{1}, SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if abstracted.RegisterCount() != 1 || abstracted.PrimaryInputCount() != 3 || abstracted.PrimaryOutputCount() != 2 {
		t.Fatalf("abstraction partition %d/%d/%d", abstracted.PrimaryInputCount(), abstracted.PrimaryOutputCount(), abstracted.RegisterCount())
	}
	for bits := 0; bits < 16; bits++ {
		p0, p1 := bits&1 != 0, bits&2 != 0
		r0, r1 := bits&4 != 0, bits&8 != 0
		sourceOut, sourceNext, _ := source.Step([]bool{p0, p1}, []bool{r0, r1})
		gotOut, gotNext, stepErr := abstracted.Step([]bool{p0, p1, r0}, []bool{r1})
		if stepErr != nil || !reflect.DeepEqual(gotOut, sourceOut) || len(gotNext) != 1 || gotNext[0] != sourceNext[1] {
			t.Fatalf("abstraction bits %04b got %v/%v want %v/%v err %v", bits, gotOut, gotNext, sourceOut, sourceNext[1:], stepErr)
		}
	}
	allAbstracted, err := source.DuplicateSequentialAbstraction([]int{1, 0}, SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 16; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		state := []bool{bits&4 != 0, bits&8 != 0}
		sourceOut, sourceNext, _ := source.Step(input, state)
		gotOut, gotNext, stepErr := allAbstracted.Step(input, state)
		if stepErr != nil || !reflect.DeepEqual(gotOut, sourceOut) || !reflect.DeepEqual(gotNext, sourceNext) {
			t.Fatalf("unsorted abstraction bits %04b got %v/%v want %v/%v err %v", bits, gotOut, gotNext, sourceOut, sourceNext, stepErr)
		}
	}

	cone, err := source.DuplicateSequentialCones([]int{1}, SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cone.PrimaryInputs, []int{0, 1}) || !reflect.DeepEqual(cone.Registers, []int{0, 1}) || !reflect.DeepEqual(cone.Outputs, []int{1}) {
		t.Fatalf("cone maps %#v", cone)
	}
	for bits := 0; bits < 16; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		state := []bool{bits&4 != 0, bits&8 != 0}
		sourceOut, sourceNext, _ := source.Step(input, state)
		coneOut, coneNext, stepErr := cone.Graph.Step(input, state)
		if stepErr != nil || coneOut[0] != sourceOut[1] || !reflect.DeepEqual(coneNext, sourceNext) {
			t.Fatalf("cone bits %04b got %v/%v source %v/%v err %v", bits, coneOut, coneNext, sourceOut, sourceNext, stepErr)
		}
	}
}

func TestSaigCounterexampleOperations(t *testing.T) {
	graph := saigStructuralFixture(t)
	trace, err := NewCounterexample(2, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	trace.Property = 1
	if err = trace.SetInitial(0, true); err != nil {
		t.Fatal(err)
	}
	if err = trace.SetInput(0, 0, true); err != nil {
		t.Fatal(err)
	}
	if err = trace.SetInput(0, 1, true); err != nil {
		t.Fatal(err)
	}
	if err = trace.SetInput(1, 1, true); err != nil {
		t.Fatal(err)
	}
	if err = trace.SetInput(2, 1, true); err != nil {
		t.Fatal(err)
	}
	valid, err := graph.VerifySequentialCounterexample(trace, SaigStructuralOptions{})
	if err != nil || !valid {
		t.Fatalf("verify %v, %v", valid, err)
	}
	failed, err := graph.FindSequentialFailedOutput(trace, SaigStructuralOptions{})
	if err != nil || failed != 0 {
		t.Fatalf("first failed %d, %v", failed, err)
	}
	state, err := graph.SequentialFailingState(trace, false, SaigStructuralOptions{})
	if err != nil || !reflect.DeepEqual(state, []bool{true, true}) {
		t.Fatalf("failing state %v, %v", state, err)
	}
	next, err := graph.SequentialFailingState(trace, true, SaigStructuralOptions{})
	if err != nil || !reflect.DeepEqual(next, []bool{true, true}) {
		t.Fatalf("next state %v, %v", next, err)
	}
	extended, err := graph.ExtendSequentialCounterexample(trace, SaigStructuralOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if extended.Registers() != 0 || extended.Inputs() != 4 || extended.Frames() != 3 || extended.Property != trace.Property {
		t.Fatalf("extended dimensions %d/%d/%d/%d", extended.Registers(), extended.Inputs(), extended.Frames(), extended.Property)
	}
	want := [][]bool{{true, true, true, false}, {false, true, true, true}, {false, true, true, true}}
	for frame := range want {
		for index := range want[frame] {
			value, inputErr := extended.Input(frame, index)
			if inputErr != nil || value != want[frame][index] {
				t.Fatalf("extended[%d][%d]=%v,%v want %v", frame, index, value, inputErr, want[frame][index])
			}
		}
	}
	bad := trace.Clone()
	bad.Property = 1
	for frame := 0; frame < bad.Frames(); frame++ {
		for input := 0; input < bad.Inputs(); input++ {
			_ = bad.SetInput(frame, input, false)
		}
	}
	_ = bad.SetInitial(0, false)
	_ = bad.SetInitial(1, false)
	if got, verifyErr := graph.VerifySequentialCounterexample(bad, SaigStructuralOptions{}); verifyErr != nil || got {
		t.Fatalf("nonfailing verify %v, %v", got, verifyErr)
	}
	if got, stateErr := graph.SequentialFailingState(bad, false, SaigStructuralOptions{}); got != nil || !errors.Is(stateErr, ErrInvariant) {
		t.Fatalf("nonfailing state %v, %v", got, stateErr)
	}
}

func TestSaigInitialPhaseExhaustive(t *testing.T) {
	source := saigStructuralFixture(t)
	for initialBits := 0; initialBits < 4; initialBits++ {
		initial := []bool{initialBits&1 != 0, initialBits&2 != 0}
		target, err := source.DuplicateWithInitialPhase(initial, SaigStructuralOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for inputBits := 0; inputBits < 4; inputBits++ {
			inputs := []bool{inputBits&1 != 0, inputBits&2 != 0}
			sourceOutputs, sourceNext, _ := source.Step(inputs, initial)
			targetOutputs, encodedNext, stepErr := target.Step(inputs, []bool{false, false})
			if stepErr != nil || !reflect.DeepEqual(targetOutputs, sourceOutputs) {
				t.Fatalf("initial/input %02b/%02b outputs %v/%v err %v", initialBits, inputBits, targetOutputs, sourceOutputs, stepErr)
			}
			for index := range sourceNext {
				if encodedNext[index] != (sourceNext[index] != initial[index]) {
					t.Fatalf("initial/input %02b/%02b next %v source %v", initialBits, inputBits, encodedNext, sourceNext)
				}
			}
		}
	}
}

func TestSaigStructuralLimitsCancellationAndAtomicity(t *testing.T) {
	graph := saigStructuralFixture(t)
	before := graph.Clone()
	if got, err := graph.DuplicateOrPrimaryOutputs(SaigStructuralOptions{MaxObjects: 2}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("limit got %v, %v", got, err)
	}
	if got, err := graph.DuplicateOrPrimaryOutputs(SaigStructuralOptions{Stop: func() bool { return true }}); got != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel got %v, %v", got, err)
	}
	if graph.Check() != nil || !reflect.DeepEqual(graph.objects, before.objects) || !reflect.DeepEqual(graph.cis, before.cis) || !reflect.DeepEqual(graph.cos, before.cos) {
		t.Fatal("failed operation mutated source")
	}
	if got, err := graph.DuplicateSequentialAbstraction([]int{0, 0}, SaigStructuralOptions{}); got != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("duplicate register got %v, %v", got, err)
	}
	trace, _ := NewCounterexample(2, 2, 2)
	trace.Property = 0
	if got, err := graph.VerifySequentialCounterexample(trace, SaigStructuralOptions{MaxFrames: 1}); got || !errors.Is(err, ErrLimit) {
		t.Fatalf("frame limit got %v, %v", got, err)
	}
}

func FuzzSaigStructuralTraceModel(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x55aa, 0xdeadbeef} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		graph := saigStructuralFixture(t)
		random := rand.New(rand.NewSource(int64(seed)))
		frames := 1 + random.Intn(8)
		trace, err := NewCounterexample(2, 2, frames)
		if err != nil {
			t.Fatal(err)
		}
		trace.Property = random.Intn(2)
		state := []bool{random.Intn(2) != 0, random.Intn(2) != 0}
		for index, value := range state {
			if err = trace.SetInitial(index, value); err != nil {
				t.Fatal(err)
			}
		}
		var finalOutputs, finalNext []bool
		for frame := 0; frame < frames; frame++ {
			inputs := []bool{random.Intn(2) != 0, random.Intn(2) != 0}
			for index, value := range inputs {
				if err = trace.SetInput(frame, index, value); err != nil {
					t.Fatal(err)
				}
			}
			finalOutputs, finalNext, err = graph.Step(inputs, state)
			if err != nil {
				t.Fatal(err)
			}
			if frame != frames-1 {
				state = finalNext
			}
		}
		got, err := graph.VerifySequentialCounterexample(trace, SaigStructuralOptions{})
		if err != nil || got != finalOutputs[trace.Property] {
			t.Fatalf("verify %v,%v want %v", got, err, finalOutputs[trace.Property])
		}
		failed, err := graph.FindSequentialFailedOutput(trace, SaigStructuralOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantFailed := -1
		for index, value := range finalOutputs {
			if value {
				wantFailed = index
				break
			}
		}
		if failed != wantFailed {
			t.Fatalf("failed output %d want %d", failed, wantFailed)
		}
		if got {
			actualState, stateErr := graph.SequentialFailingState(trace, false, SaigStructuralOptions{})
			actualNext, nextErr := graph.SequentialFailingState(trace, true, SaigStructuralOptions{})
			if stateErr != nil || nextErr != nil || !reflect.DeepEqual(actualState, state) || !reflect.DeepEqual(actualNext, finalNext) {
				t.Fatalf("state %v/%v errors %v/%v want %v/%v", actualState, actualNext, stateErr, nextErr, state, finalNext)
			}
			extended, extendErr := graph.ExtendSequentialCounterexample(trace, SaigStructuralOptions{})
			if extendErr != nil || extended.Inputs() != graph.InputCount() || extended.Frames() != frames {
				t.Fatalf("extend %v, %v", extended, extendErr)
			}
		}
	})
}

func saigRandomStructuralGraph(t testing.TB, random *rand.Rand) *AigManager {
	t.Helper()
	graph := NewAigManager()
	primaryCount := 1 + random.Intn(3)
	registerCount := 1 + random.Intn(3)
	literals := []AigLit{AigConst0, AigConst1}
	for index := 0; index < primaryCount+registerCount; index++ {
		literal, err := graph.CreateCI()
		if err != nil {
			t.Fatal(err)
		}
		literals = append(literals, literal)
	}
	for index := 0; index < 1+random.Intn(10); index++ {
		left := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
		right := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
		var literal AigLit
		var err error
		switch random.Intn(3) {
		case 0:
			literal, err = graph.And(left, right)
		case 1:
			literal, err = graph.Or(left, right)
		default:
			literal, err = graph.Xor(left, right)
		}
		if err != nil {
			t.Fatal(err)
		}
		literals = append(literals, literal)
	}
	outputCount := 1 + random.Intn(3)
	for index := 0; index < outputCount+registerCount; index++ {
		driver := literals[random.Intn(len(literals))].NotIf(random.Intn(2) != 0)
		if _, err := graph.CreateCO(driver); err != nil {
			t.Fatal(err)
		}
	}
	if err := graph.SetRegisterCount(registerCount); err != nil {
		t.Fatal(err)
	}
	if err := graph.Check(); err != nil {
		t.Fatal(err)
	}
	return graph
}

func FuzzSaigStructuralTransforms(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 3, 0x12345678, 0xfeedface} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		random := rand.New(rand.NewSource(int64(seed)))
		source := saigRandomStructuralGraph(t, random)
		inputs := make([]bool, source.PrimaryInputCount())
		state := make([]bool, source.RegisterCount())
		initial := make([]bool, source.RegisterCount())
		for index := range inputs {
			inputs[index] = random.Intn(2) != 0
		}
		for index := range state {
			state[index] = random.Intn(2) != 0
			initial[index] = random.Intn(2) != 0
		}
		sourceOutputs, sourceNext, err := source.Step(inputs, state)
		if err != nil {
			t.Fatal(err)
		}
		ored, err := source.DuplicateOrPrimaryOutputs(SaigStructuralOptions{})
		if err != nil {
			t.Fatal(err)
		}
		oredOutputs, oredNext, err := ored.Step(inputs, state)
		if err != nil {
			t.Fatal(err)
		}
		wantOr := false
		for _, value := range sourceOutputs {
			wantOr = wantOr || value
		}
		if len(oredOutputs) != 1 || oredOutputs[0] != wantOr || !reflect.DeepEqual(oredNext, sourceNext) {
			t.Fatalf("OR duplicate %v/%v want %v/%v", oredOutputs, oredNext, wantOr, sourceNext)
		}

		phase, err := source.DuplicateWithInitialPhase(initial, SaigStructuralOptions{})
		if err != nil {
			t.Fatal(err)
		}
		phaseOutputs, phaseNext, err := phase.Step(inputs, make([]bool, len(initial)))
		if err != nil {
			t.Fatal(err)
		}
		initialOutputs, initialNext, err := source.Step(inputs, initial)
		if err != nil || !reflect.DeepEqual(phaseOutputs, initialOutputs) {
			t.Fatalf("phase outputs %v/%v err %v", phaseOutputs, initialOutputs, err)
		}
		for index := range phaseNext {
			if phaseNext[index] != (initialNext[index] != initial[index]) {
				t.Fatalf("phase next %v source %v initial %v", phaseNext, initialNext, initial)
			}
		}

		selected := random.Intn(source.PrimaryOutputCount())
		cone, err := source.DuplicateSequentialCones([]int{selected}, SaigStructuralOptions{})
		if err != nil {
			t.Fatal(err)
		}
		coneInputs := make([]bool, len(cone.PrimaryInputs))
		for index, original := range cone.PrimaryInputs {
			coneInputs[index] = inputs[original]
		}
		coneState := make([]bool, len(cone.Registers))
		for index, original := range cone.Registers {
			coneState[index] = state[original]
		}
		coneOutputs, coneNext, err := cone.Graph.Step(coneInputs, coneState)
		if err != nil || len(coneOutputs) != 1 || coneOutputs[0] != sourceOutputs[selected] {
			t.Fatalf("cone outputs %v err %v source %v selected %d", coneOutputs, err, sourceOutputs, selected)
		}
		for index, original := range cone.Registers {
			if coneNext[index] != sourceNext[original] {
				t.Fatalf("cone next %v source %v registers %v", coneNext, sourceNext, cone.Registers)
			}
		}
		if err := source.Check(); err != nil {
			t.Fatalf("source mutated: %v", err)
		}
	})
}
