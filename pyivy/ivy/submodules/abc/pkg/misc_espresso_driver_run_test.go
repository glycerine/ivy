package abc

// Go-only end-to-end byte driver tests for src/misc/espresso/main.c.

import (
	"errors"
	"testing"
)

func espressoDriverInputBytes(t testing.TB, pla *EspressoPLA) []byte {
	t.Helper()
	raw, err := pla.MarshalPLA(EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff, EspressoPLAOutputOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestRunEspressoDriverEndToEndUnionAndSuppression(t *testing.T) {
	first := espressoPLAFromOutputTruths(t, 2, []uint64{0b1010})
	second := espressoPLAFromOutputTruths(t, 2, []uint64{0b1100})
	result, err := RunEspressoDriver([]string{"-Dunion", "-ofdr", "first.pla", "second.pla"}, [][]byte{espressoDriverInputBytes(t, first), espressoDriverInputBytes(t, second)}, EspressoDriverRunOptions{})
	if err != nil || len(result.Output) == 0 || result.Config.Command.Kind != EspressoCommandUnion || result.Execution.PLA.Filename != "first.pla" {
		t.Fatal(result.Config, string(result.Output), result.Execution.PLA, err)
	}
	parsed, err := ParseEspressoPLA(result.Output, EspressoPLAInputOptions{})
	if err != nil || espressoFamilyTruth(t, parsed.PLA.Layout, parsed.PLA.On) != 0b1110 {
		t.Fatal(string(result.Output), err)
	}

	suppressed, err := RunEspressoDriver([]string{"-x"}, [][]byte{espressoDriverInputBytes(t, first)}, EspressoDriverRunOptions{})
	if err != nil || len(suppressed.Output) != 0 || suppressed.Execution.PLA == nil {
		t.Fatal(suppressed, err)
	}
	stats, err := RunEspressoDriver([]string{"-Dstats"}, [][]byte{espressoDriverInputBytes(t, first)}, EspressoDriverRunOptions{})
	if err != nil || len(stats.Output) != 0 || stats.Execution.Text == "" {
		t.Fatal(stats, err)
	}
}

func TestRunEspressoDriverRandomReductionOrder(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 3, []uint64{0b01111110})
	result, err := RunEspressoDriver([]string{"-erandom", "-ofdr"}, [][]byte{espressoDriverInputBytes(t, pla)}, EspressoDriverRunOptions{Execution: EspressoDriverExecutionOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 128, MaxIterations: 64}})
	if err != nil || !result.Config.Runtime.UseRandomOrder || result.Execution.PLA == nil {
		t.Fatal(result.Config.Runtime, result.Execution.PLA, err)
	}
	parsed, err := ParseEspressoPLA(result.Output, EspressoPLAInputOptions{})
	if err != nil || espressoFamilyTruth(t, parsed.PLA.Layout, parsed.PLA.On) != 0b01111110 {
		t.Fatal(string(result.Output), err)
	}
}

func TestRunEspressoDriverErrorsLimitsAndCancellation(t *testing.T) {
	pla := espressoDriverInputBytes(t, espressoPLAFromOutputTruths(t, 2, []uint64{0b1010}))
	checks := []struct {
		args    []string
		inputs  [][]byte
		options EspressoDriverRunOptions
		want    error
	}{
		{nil, nil, EspressoDriverRunOptions{}, ErrDimension},
		{[]string{"-Dunion"}, [][]byte{pla}, EspressoDriverRunOptions{}, ErrDimension},
		{nil, [][]byte{[]byte("bad")}, EspressoDriverRunOptions{}, ErrDimension},
		{nil, [][]byte{pla}, EspressoDriverRunOptions{Input: EspressoPLAInputOptions{MaxBytes: 1}}, ErrLimit},
		{nil, [][]byte{pla}, EspressoDriverRunOptions{Execution: EspressoDriverExecutionOptions{Stop: func() bool { return true }}}, ErrCancelled},
		{nil, [][]byte{pla}, EspressoDriverRunOptions{Output: EspressoPLAOutputOptions{MaxBytes: 1}}, ErrLimit},
	}
	for _, check := range checks {
		_, err := RunEspressoDriver(check.args, check.inputs, check.options)
		if !errors.Is(err, check.want) {
			t.Fatal(check.args, err, check.want)
		}
	}
}

func TestRunManyEspressoDriverAndDiagnosticTest(t *testing.T) {
	first := espressoPLAFromOutputTruths(t, 2, []uint64{0b1010})
	second := espressoPLAFromOutputTruths(t, 2, []uint64{0b0110})
	result, err := RunEspressoDriver([]string{"-Dmany", "-ofdr"}, [][]byte{espressoDriverInputBytes(t, first), espressoDriverInputBytes(t, second)}, EspressoDriverRunOptions{Execution: EspressoDriverExecutionOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 128, MaxIterations: 64}})
	if err != nil || len(result.Executions) != 2 || result.Execution.PLA == nil || len(result.Output) == 0 {
		t.Fatal(result, err)
	}
	for index, want := range []uint64{0b1010, 0b0110} {
		if got := espressoFamilyTruth(t, result.Executions[index].PLA.Layout, result.Executions[index].PLA.On); got != want {
			t.Fatal(index, got, want)
		}
	}
	if _, err = RunEspressoDriver([]string{"-Dmany"}, [][]byte{espressoDriverInputBytes(t, first), espressoDriverInputBytes(t, second)}, EspressoDriverRunOptions{MaxOutputBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}

	testResult, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, "test"), first, nil, EspressoDriverExecutionOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 128})
	if err != nil || testResult.PLA == nil || espressoFamilyTruth(t, testResult.PLA.Layout, testResult.PLA.On) != 0b1010 {
		t.Fatal(testResult.PLA, err)
	}
}

func FuzzRunEspressoDriver(f *testing.F) {
	f.Add(byte(0xa), byte(0xc), byte(0))
	f.Add(byte(0xf), byte(0x0), byte(4))
	f.Fuzz(func(t *testing.T, firstByte, secondByte, commandByte byte) {
		firstTruth, secondTruth := uint64(firstByte&0xf), uint64(secondByte&0xf)
		first := espressoPLAFromOutputTruths(t, 2, []uint64{firstTruth})
		second := espressoPLAFromOutputTruths(t, 2, []uint64{secondTruth})
		names := []string{"intersect", "union", "sharp", "dsharp", "xor"}
		name := names[int(commandByte)%len(names)]
		result, err := RunEspressoDriver([]string{"-D" + name, "-ofdr"}, [][]byte{espressoDriverInputBytes(t, first), espressoDriverInputBytes(t, second)}, EspressoDriverRunOptions{Execution: EspressoDriverExecutionOptions{MaxCubes: 1 << 12, MaxOperations: 1 << 18}, Output: EspressoPLAOutputOptions{MaxCubes: 1 << 12, MaxParts: 1 << 14, MaxBytes: 1 << 16}})
		if err != nil {
			t.Fatal(name, firstTruth, secondTruth, err)
		}
		parsed, err := ParseEspressoPLA(result.Output, EspressoPLAInputOptions{MaxBytes: 1 << 16, MaxLines: 1 << 12, MaxVariables: 16, MaxParts: 128, MaxCubes: 1 << 12, MaxTokens: 1 << 14, MaxOperations: 1 << 18})
		if err != nil {
			t.Fatal(name, string(result.Output), err)
		}
		want := firstTruth & secondTruth
		switch name {
		case "union":
			want = firstTruth | secondTruth
		case "sharp", "dsharp":
			want = firstTruth &^ secondTruth
		case "xor":
			want = firstTruth ^ secondTruth
		}
		if got := espressoFamilyTruth(t, parsed.PLA.Layout, parsed.PLA.On); got != want {
			t.Fatal(name, firstTruth, secondTruth, got, want)
		}
	})
}
