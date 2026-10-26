package abc

// Go-only exact-table, ownership and fuzz tests for src/misc/espresso/main.h.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoCommandAndDebugTables(t *testing.T) {
	commands := EspressoCommandDescriptors()
	if len(commands) != 43 || commands[0] != (EspressoCommandDescriptor{"ESPRESSO", EspressoCommandMinimize, 1, true, true}) || commands[len(commands)-1].Name != "test" {
		t.Fatal(len(commands), commands[0], commands[len(commands)-1])
	}
	for _, descriptor := range commands {
		got, ok := LookupEspressoCommand(descriptor.Name)
		if !ok || got != descriptor || descriptor.InputPLAs < 1 || descriptor.InputPLAs > 2 {
			t.Fatal(descriptor, got, ok)
		}
	}
	if first, _ := LookupEspressoCommand("single_output"); first.Kind != EspressoCommandSingleOutput {
		t.Fatal(first)
	}
	if second, _ := LookupEspressoCommand("so"); second.Kind != EspressoCommandSingleOutput {
		t.Fatal(second)
	}
	commands[0].Name = "changed"
	if again := EspressoCommandDescriptors(); again[0].Name != "ESPRESSO" {
		t.Fatal(again[0])
	}
	if _, ok := LookupEspressoCommand("espresso"); ok {
		t.Fatal("lookup must remain source-case-sensitive")
	}

	debug := EspressoDebugDescriptors()
	if len(debug) != 16 {
		t.Fatal(len(debug))
	}
	for _, descriptor := range debug {
		mask, ok := LookupEspressoDebug(descriptor.Name)
		if !ok || mask != descriptor.Mask || mask == 0 {
			t.Fatal(descriptor, mask, ok)
		}
	}
	if mask, _ := LookupEspressoDebug("expand1"); mask != EspressoDebugExpand|EspressoDebugExpandDetail {
		t.Fatal(mask)
	}
	debug[0].Name = "changed"
	if slices.Equal(debug, EspressoDebugDescriptors()) {
		t.Fatal("descriptor storage aliased")
	}
}

func TestApplyEspressoBehaviorOptionsOwned(t *testing.T) {
	options := DefaultEspressoRuntimeOptions()
	names := []string{"eat", "eatdots", "fast", "kiss", "ness", "nirr", "nunwrap", "onset", "pos", "random", "strong"}
	for _, name := range names {
		var err error
		options, err = ApplyEspressoBehaviorOption(options, name)
		if err != nil {
			t.Fatal(name, err)
		}
	}
	if options.EchoComments || options.EchoUnknownCommands || !options.SingleExpand || !options.KISS || options.RemoveEssential || options.ForceIrredundant || options.UnwrapOnset || !options.RecomputeOnset || !options.Positive || !options.UseRandomOrder || !options.UseSuperGasp {
		t.Fatal(options)
	}
	base := DefaultEspressoRuntimeOptions()
	if _, err := ApplyEspressoBehaviorOption(base, "unknown"); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if base != DefaultEspressoRuntimeOptions() {
		t.Fatal("input value mutated")
	}
}

func FuzzEspressoDriverTables(f *testing.F) {
	f.Add("ESPRESSO")
	f.Add("expand1")
	f.Add("random")
	f.Add("unknown")
	f.Fuzz(func(t *testing.T, name string) {
		command, commandOK := LookupEspressoCommand(name)
		if commandOK {
			found := false
			for _, candidate := range EspressoCommandDescriptors() {
				found = found || candidate == command && candidate.Name == name
			}
			if !found {
				t.Fatal(name, command)
			}
		}
		mask, debugOK := LookupEspressoDebug(name)
		if debugOK && mask == 0 {
			t.Fatal(name, mask)
		}
		base := DefaultEspressoRuntimeOptions()
		applied, err := ApplyEspressoBehaviorOption(base, name)
		if err != nil && !errors.Is(err, ErrUnsupported) {
			t.Fatal(name, err)
		}
		if err == nil && applied == base {
			t.Fatal(name, "recognized behavior option made no change")
		}
	})
}
