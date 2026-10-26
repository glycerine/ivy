package abc

// Go-only ownership, lookup and exhaustive tests for src/misc/espresso/globals.c.

import (
	"errors"
	"math/bits"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestEspressoRuntimeOptionsDefaultsProjectionAndIsolation(t *testing.T) {
	first := DefaultEspressoRuntimeOptions()
	second := DefaultEspressoRuntimeOptions()
	if !first.EchoComments || !first.EchoUnknownCommands || !first.ForceIrredundant || !first.PrintSolution || !first.RemoveEssential || !first.UnwrapOnset {
		t.Fatal(first)
	}
	first.UseSuperGasp = true
	first.SkipMakeSparse = true
	first.RecomputeOnset = true
	projected, err := first.MinimizeOptions()
	if err != nil || projected.ReduceOrder != EspressoReduceDistanceOrder || !projected.UseSuperGasp || !projected.SkipMakeSparse || !projected.RecomputeOnset || second.UseRandomOrder {
		t.Fatal(projected, second)
	}
	first.UseRandomOrder = true
	randomProjected, err := first.MinimizeOptions()
	if err != nil || !randomProjected.UseRandomOrder || randomProjected.RandomSeed != 314973 {
		t.Fatal(randomProjected, err)
	}

	var wait sync.WaitGroup
	results := make(chan bool, 32)
	for index := 0; index < 32; index++ {
		wait.Add(1)
		go func(random bool) {
			defer wait.Done()
			options := DefaultEspressoRuntimeOptions()
			options.UseRandomOrder = random
			projected, err := options.MinimizeOptions()
			results <- err == nil && projected.UseRandomOrder
		}(index&1 != 0)
	}
	wait.Wait()
	close(results)
	counts := [2]int{}
	for unsupported := range results {
		counts[btoi(unsupported)]++
	}
	if counts != [2]int{16, 16} {
		t.Fatal(counts)
	}
}

func TestEspressoPLATypeDescriptorsAndRuntimeTotalsOwnership(t *testing.T) {
	descriptors := EspressoPLATypeDescriptors()
	if len(descriptors) != 20 {
		t.Fatal(len(descriptors))
	}
	for _, descriptor := range descriptors {
		withDash, ok := LookupEspressoPLAType(descriptor.Name)
		withoutDash, plainOK := LookupEspressoPLAType(descriptor.Name[1:])
		if !ok || !plainOK || withDash != descriptor.Type || withoutDash != descriptor.Type {
			t.Fatal(descriptor, withDash, withoutDash, ok, plainOK)
		}
	}
	descriptors[0].Name = "changed"
	if again := EspressoPLATypeDescriptors(); again[0].Name != "-f" {
		t.Fatal(again[0])
	}
	for _, bad := range []string{"", "-", "F", "unknown", "--f"} {
		if value, ok := LookupEspressoPLAType(bad); ok || value != 0 {
			t.Fatal(bad, value, ok)
		}
	}

	first := NewEspressoRuntimeTotals()
	second := NewEspressoRuntimeTotals()
	if len(first.Names) != 16 || first.Names[0] != "READ       " || first.Names[15] != "WRITE      " {
		t.Fatal(first.Names)
	}
	first.Names[0] = "changed"
	first.Calls[0] = 9
	if second.Names[0] != "READ       " || second.Calls[0] != 0 || slices.Equal(first.Names, second.Names) {
		t.Fatal(first, second)
	}
}

func TestEspressoBytePopulationExhaustive(t *testing.T) {
	for value := 0; value < 256; value++ {
		want := 0
		for bit := 0; bit < 8; bit++ {
			want += value >> uint(bit) & 1
		}
		if got := espressoBytePopulation(byte(value)); got != want || got != bits.OnesCount8(uint8(value)) {
			t.Fatal(value, got, want)
		}
	}
}

func TestEspressoRuntimeFormatting(t *testing.T) {
	totals := NewEspressoRuntimeTotals()
	totals.Milliseconds[0], totals.Calls[0] = 1000, 2
	totals.Milliseconds[15], totals.Calls[15] = 1, 1
	formatted, err := totals.FormatRuntime(1024)
	want := "# READ       \t2 call(s) for 1.00 sec (99.8%)\n# WRITE      \t1 call(s) for 0.00 sec (0.0%)\n"
	if err != nil || formatted != want {
		t.Fatal(formatted, want, err)
	}
	if _, err = totals.FormatRuntime(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	bad := NewEspressoRuntimeTotals()
	bad.Milliseconds[0] = -1
	if _, err = bad.FormatRuntime(1024); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	large := NewEspressoRuntimeTotals()
	large.Milliseconds[0], large.Calls[0] = int64(^uint64(0)>>1)-1, 1
	formatted, err = large.FormatRuntime(1024)
	if err != nil || !strings.Contains(formatted, "(99.9%)") {
		t.Fatal(formatted, err)
	}
}

func FuzzEspressoGlobalReplacements(f *testing.F) {
	f.Add(byte(0), "f")
	f.Add(byte(0xff), "-fdrc")
	f.Add(byte(0x55), "unknown")
	f.Fuzz(func(t *testing.T, value byte, name string) {
		got := espressoBytePopulation(value)
		want := 0
		for current := value; current != 0; current &= current - 1 {
			want++
		}
		if got != want {
			t.Fatal(value, got, want)
		}
		kind, ok := LookupEspressoPLAType(name)
		if ok {
			found := false
			for _, descriptor := range EspressoPLATypeDescriptors() {
				if name == descriptor.Name || name == descriptor.Name[1:] {
					found = descriptor.Type == kind
				}
			}
			if !found {
				t.Fatal(name, kind)
			}
		}
	})
}
