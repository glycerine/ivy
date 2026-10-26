package abc

// Go-only semantic dispatch tests for src/misc/espresso/main.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoDriverConfigFor(t testing.TB, name string) EspressoDriverConfig {
	t.Helper()
	descriptor, ok := LookupEspressoCommand(name)
	if !ok {
		t.Fatal(name)
	}
	return EspressoDriverConfig{Command: descriptor, Runtime: DefaultEspressoRuntimeOptions(), First: -1, Last: -1}
}

func TestExecuteEspressoCommandBooleanOperationsAndOwnership(t *testing.T) {
	first := espressoPLAFromOutputTruths(t, 2, []uint64{0b1010})
	second := espressoPLAFromOutputTruths(t, 2, []uint64{0b1100})
	firstBefore := espressoFamilyMasks(t, first.On)
	checks := []struct {
		name string
		want uint64
	}{
		{"intersect", 0b1000},
		{"union", 0b1110},
		{"sharp", 0b0010},
		{"dsharp", 0b0010},
		{"xor", 0b0110},
	}
	for _, check := range checks {
		result, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, check.name), first, second, EspressoDriverExecutionOptions{})
		if err != nil || result.PLA == nil || espressoFamilyTruth(t, result.PLA.Layout, result.PLA.On) != check.want {
			t.Fatal(check.name, result.PLA, err)
		}
	}
	if !slices.Equal(firstBefore, espressoFamilyMasks(t, first.On)) {
		t.Fatal("dispatch mutated input")
	}
}

func TestExecuteEspressoCommandUnaryDiagnosticsAndErrors(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 2, []uint64{0b1010})
	for _, name := range []string{"echo", "ESPRESSO", "exact", "qm", "simplify", "contain", "disjoint", "lexsort", "minterms", "d1merge", "d1merge_in", "primes", "separate", "expand", "irred", "reduce", "gasp", "super_gasp", "make_sparse"} {
		result, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, name), pla, nil, EspressoDriverExecutionOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 128, MaxIterations: 64})
		if err != nil || result.PLA == nil || espressoFamilyTruth(t, result.PLA.Layout, result.PLA.On) != 0b1010 {
			t.Fatal(name, result.PLA, err)
		}
	}
	stats, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, "stats"), pla, nil, EspressoDriverExecutionOptions{})
	if err != nil || stats.Text == "" {
		t.Fatal(stats, err)
	}
	taut, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, "taut"), pla, nil, EspressoDriverExecutionOptions{})
	if err != nil || !taut.Decided || taut.Decision {
		t.Fatal(taut, err)
	}
	consistent, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, "check"), pla, nil, EspressoDriverExecutionOptions{})
	if err != nil || !consistent.Decided || !consistent.Decision {
		t.Fatal(consistent, err)
	}
	unsupported, _ := LookupEspressoCommand("many")
	if _, err = ExecuteEspressoCommand(EspressoDriverConfig{Command: unsupported, Runtime: DefaultEspressoRuntimeOptions()}, pla, nil, EspressoDriverExecutionOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err = ExecuteEspressoCommand(espressoDriverConfigFor(t, "intersect"), pla, nil, EspressoDriverExecutionOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = ExecuteEspressoCommand(espressoDriverConfigFor(t, "echo"), pla, nil, EspressoDriverExecutionOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func TestExecuteEspressoCommandOutputPhasePairAndVerification(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 2, []uint64{0b0110, 0b1000})
	limits := EspressoDriverExecutionOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 20, MaxRows: 1 << 14, MaxDepth: 128, MaxIterations: 64}
	for _, name := range []string{"so", "so_both", "opo", "opoall"} {
		result, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, name), pla, nil, limits)
		if err != nil || result.PLA == nil || len(result.PositivePhase) != 2 {
			t.Fatal(name, result, err)
		}
		for output, want := range []uint64{0b0110, 0b1000} {
			if !result.PositivePhase[output] {
				want = 0xf &^ want
			}
			if got := espressoStoredOutputTruth(t, result.PLA, output); got != want {
				t.Fatal(name, output, got, want, result.PositivePhase)
			}
		}
	}
	for _, name := range []string{"pair", "pairall"} {
		result, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, name), pla, nil, limits)
		if err != nil || result.PLA == nil || len(result.Pairing.Pairs) != 1 {
			t.Fatal(name, result, err)
		}
		for output, want := range []uint64{0b0110, 0b1000} {
			if got := espressoPairedTruth(t, pla, EspressoPairResult{PLA: result.PLA, Pairing: result.Pairing}, output); got != want {
				t.Fatal(name, output, got, want)
			}
		}
	}

	verified, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, "verify"), pla, pla, limits)
	if err != nil || !verified.Decided || !verified.Decision {
		t.Fatal(verified, err)
	}
	labels, _ := pla.DefaultLabels()
	labeled, _ := NewEspressoPLA(pla.Layout, pla.On, pla.DontCare, pla.Off, labels)
	plaVerified, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, "PLAverify"), labeled, labeled, limits)
	if err != nil || !plaVerified.Decision {
		t.Fatal(plaVerified, err)
	}
}

func FuzzExecuteEspressoBooleanCommands(f *testing.F) {
	f.Add(byte(0xa), byte(0xc), byte(0))
	f.Add(byte(0xf), byte(0x3), byte(4))
	f.Fuzz(func(t *testing.T, firstByte, secondByte, commandByte byte) {
		firstTruth, secondTruth := uint64(firstByte&0xf), uint64(secondByte&0xf)
		first := espressoPLAFromOutputTruths(t, 2, []uint64{firstTruth})
		second := espressoPLAFromOutputTruths(t, 2, []uint64{secondTruth})
		names := []string{"intersect", "union", "sharp", "dsharp", "xor"}
		name := names[int(commandByte)%len(names)]
		result, err := ExecuteEspressoCommand(espressoDriverConfigFor(t, name), first, second, EspressoDriverExecutionOptions{MaxCubes: 1 << 12, MaxOperations: 1 << 18})
		if err != nil {
			t.Fatal(name, firstTruth, secondTruth, err)
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
		if got := espressoFamilyTruth(t, result.PLA.Layout, result.PLA.On); got != want {
			t.Fatal(name, firstTruth, secondTruth, got, want)
		}
	})
}
