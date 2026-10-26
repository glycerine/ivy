package abc

// Go-only driver parser and formatting tests for src/misc/espresso/main.c.

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestParseEspressoDriverArgumentsModernAndLegacy(t *testing.T) {
	modern := []string{"-D", "exact", "-ofdr", "-e", "strong", "-vexpand1", "-s", "-x", "-S2", "-r1-3", "input.pla"}
	before := slices.Clone(modern)
	config, err := ParseEspressoDriverArguments(modern, EspressoDriverParseOptions{})
	if err != nil || config.Command.Kind != EspressoCommandExact || config.OutputType != EspressoPLAOutputOn|EspressoPLAOutputDontCare|EspressoPLAOutputOff || !config.Runtime.UseSuperGasp || !config.Runtime.VerboseDebug || config.Runtime.Debug != EspressoDebugExpand|EspressoDebugExpandDetail || !config.Runtime.Summary || config.Runtime.PrintSolution || config.Strategy != 2 || config.First != 1 || config.Last != 3 || !slices.Equal(config.Inputs, []string{"input.pla"}) {
		t.Fatal(config, err)
	}
	if !slices.Equal(modern, before) {
		t.Fatal("arguments mutated", modern, before)
	}

	legacy, err := ParseEspressoDriverArguments([]string{"-do", "so", "-out", "fr", "-fast", "-fdr", "-fr", "-f", "-"}, EspressoDriverParseOptions{})
	if err != nil || legacy.Command.Kind != EspressoCommandSingleOutput || legacy.OutputType != EspressoPLAOutputOn|EspressoPLAOutputOff || legacy.InputType != EspressoPLAOutputOn || !legacy.Runtime.SingleExpand || !slices.Equal(legacy.Inputs, []string{"-"}) {
		t.Fatal(legacy, err)
	}
}

func TestParseEspressoDriverArgumentsDefaultsErrorsAndUsage(t *testing.T) {
	config, err := ParseEspressoDriverArguments(nil, EspressoDriverParseOptions{})
	if err != nil || config.Command.Kind != EspressoCommandMinimize || config.OutputType != EspressoPLAOutputOn || config.InputType != EspressoPLAOutputOn|EspressoPLAOutputDontCare || config.First != -1 || config.Last != -1 || !config.Runtime.PrintSolution {
		t.Fatal(config, err)
	}
	checks := []struct {
		args    []string
		options EspressoDriverParseOptions
		want    error
	}{
		{[]string{"-Dunknown"}, EspressoDriverParseOptions{}, ErrUnsupported},
		{[]string{"-Sbad"}, EspressoDriverParseOptions{}, ErrDimension},
		{[]string{"-r1"}, EspressoDriverParseOptions{}, ErrDimension},
		{[]string{"a", "b"}, EspressoDriverParseOptions{}, ErrDimension},
		{[]string{"a"}, EspressoDriverParseOptions{MaxArguments: 1, MaxBytes: 1}, nil},
		{[]string{"ab"}, EspressoDriverParseOptions{MaxBytes: 1}, ErrLimit},
		{[]string{"a", "b"}, EspressoDriverParseOptions{MaxArguments: 1}, ErrLimit},
		{nil, EspressoDriverParseOptions{Stop: func() bool { return true }}, ErrCancelled},
	}
	for _, check := range checks {
		_, parseErr := ParseEspressoDriverArguments(check.args, check.options)
		if !errors.Is(parseErr, check.want) {
			t.Fatal(check.args, parseErr, check.want)
		}
	}
	usage := EspressoDriverUsage()
	for _, descriptor := range EspressoCommandDescriptors() {
		if !strings.Contains(usage, descriptor.Name) {
			t.Fatal("usage omits", descriptor.Name)
		}
	}
	for _, line := range strings.Split(usage, "\n") {
		if len(line) > 76 {
			t.Fatal("overlong usage line", len(line), line)
		}
	}
}

func FuzzParseEspressoDriverArguments(f *testing.F) {
	f.Add("-Dexact -ofdr -S2 -r0-1 input.pla")
	f.Add("-do so -out fr -fast -")
	f.Add("-e unknown")
	f.Fuzz(func(t *testing.T, text string) {
		if len(text) > 2048 {
			return
		}
		args := strings.Fields(text)
		before := slices.Clone(args)
		config, err := ParseEspressoDriverArguments(args, EspressoDriverParseOptions{MaxArguments: 128, MaxBytes: 2048})
		if !slices.Equal(args, before) {
			t.Fatal("input mutated", args, before)
		}
		if err != nil {
			return
		}
		known, ok := LookupEspressoCommand(config.Command.Name)
		if !ok || known != config.Command || len(config.Inputs) > config.Command.InputPLAs || config.First < -1 || config.Last < -1 {
			t.Fatal(config)
		}
		if _, ok := LookupEspressoPLAType(espressoOutputTypeLetters(config.InputType)); !ok {
			t.Fatal(config.InputType)
		}
	})
}
