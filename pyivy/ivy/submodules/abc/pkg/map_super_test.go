package abc

// Go-only validation for the src/map/super port.

import (
	"bytes"
	"errors"
	"testing"
)

func evalAndSuper(lib *AndSuperLibrary, id, assignment int, complement bool) bool {
	gate := lib.Gates[id]
	var value bool
	switch {
	case gate.Input == -1:
		value = false
	case gate.Input >= 0:
		value = assignment&(1<<uint(gate.Input)) != 0
	default:
		value = evalAndSuper(lib, gate.Left.Gate, assignment, gate.Left.Complement) && evalAndSuper(lib, gate.Right.Gate, assignment, gate.Right.Complement)
	}
	return value != complement
}

func checkAndSuperTruth(t testing.TB, lib *AndSuperLibrary) {
	t.Helper()
	if err := lib.Check(); err != nil {
		t.Fatalf("check: %v", err)
	}
	for id, gate := range lib.Gates {
		var truth uint32
		for assignment := 0; assignment < 1<<uint(lib.Inputs); assignment++ {
			if evalAndSuper(lib, id, assignment, false) {
				truth |= uint32(1) << uint(assignment)
			}
		}
		if truth != gate.Truth {
			t.Fatalf("gate %d truth = %08x, evaluation = %08x", id, gate.Truth, truth)
		}
	}
}

func TestAndSupergateEnumerationTruthAndText(t *testing.T) {
	for inputs := 0; inputs <= 5; inputs++ {
		lib, err := PrecomputeAndSupergates(inputs, 1, AndSuperOptions{})
		if err != nil {
			t.Fatalf("inputs %d: %v", inputs, err)
		}
		checkAndSuperTruth(t, lib)
		text, err := lib.MarshalText()
		if err != nil || !bytes.Contains(text, []byte("# AND2/INV supergates")) {
			t.Fatalf("marshal inputs %d: %q, %v", inputs, text, err)
		}
		again, _ := lib.MarshalText()
		if !bytes.Equal(text, again) {
			t.Fatal("AND library serialization is not deterministic")
		}
	}
	lib, _ := PrecomputeAndSupergates(2, 1, AndSuperOptions{})
	if len(lib.Gates) != 7 || lib.Tried != 15 {
		t.Fatalf("two-input level-one inventory = %d gates, %d tries", len(lib.Gates), lib.Tried)
	}
	second, err := PrecomputeAndSupergates(2, 2, AndSuperOptions{})
	if err != nil || second.Tried != 91 {
		t.Fatalf("two-input level-two source tries = %d, %v", second.Tried, err)
	}
	if super2WriteCompare("(A)", "<B>") != 1 || super2WriteCompare("z", "a") != -1 || super2WriteCompare("(a)", "a") != -1 {
		t.Fatal("branch comparison does not preserve the source ordering")
	}
}

func TestAndSupergateErrorsAndAtomicFailure(t *testing.T) {
	for _, tc := range []struct{ inputs, levels int }{{-1, 0}, {6, 0}, {2, -1}, {2, 6}} {
		if lib, err := PrecomputeAndSupergates(tc.inputs, tc.levels, AndSuperOptions{}); lib != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("dimensions %+v = %#v, %v", tc, lib, err)
		}
	}
	if lib, err := PrecomputeAndSupergates(3, 2, AndSuperOptions{MaxGates: 3}); lib != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("gate limit = %#v, %v", lib, err)
	}
	if lib, err := PrecomputeAndSupergates(3, 2, AndSuperOptions{MaxAttempts: 1}); lib != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("attempt limit = %#v, %v", lib, err)
	}
	if lib, err := PrecomputeAndSupergates(3, 2, AndSuperOptions{Stop: func() bool { return true }}); lib != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel = %#v, %v", lib, err)
	}
	var nilLib *AndSuperLibrary
	if _, err := nilLib.MarshalText(); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil marshal = %v", err)
	}
}

func superTestRoots() []SuperPrimitiveGate {
	return []SuperPrimitiveGate{
		{Name: "AND2", Fanins: 2, Truth: 0b1000, Area: 1, PinDelays: []float64{1, 1}},
		{Name: "OR2", Fanins: 2, Truth: 0b1110, Area: 1.2, PinDelays: []float64{1.1, 1.1}},
		{Name: "INV", Fanins: 1, Truth: 0b01, Area: 0.2, PinDelays: []float64{0.2}},
	}
}

func evalSuperGate(lib *SuperLibrary, id, assignment int) bool {
	gate := lib.Gates[id]
	if gate.Root < 0 {
		return assignment&(1<<uint(gate.Input)) != 0
	}
	rootAssignment := 0
	for pin, fanin := range gate.Fanins {
		if evalSuperGate(lib, fanin, assignment) {
			rootAssignment |= 1 << uint(pin)
		}
	}
	return lib.Roots[gate.Root].Truth&(uint64(1)<<uint(rootAssignment)) != 0
}

func checkSuperTruth(t testing.TB, lib *SuperLibrary) {
	t.Helper()
	if err := lib.Check(); err != nil {
		t.Fatalf("check: %v", err)
	}
	for id, gate := range lib.Gates {
		var truth uint64
		for assignment := 0; assignment < 1<<uint(lib.Variables); assignment++ {
			if evalSuperGate(lib, id, assignment) {
				truth |= uint64(1) << uint(assignment)
			}
		}
		if truth != gate.Truth {
			t.Fatalf("gate %d truth = %016x, evaluation = %016x", id, gate.Truth, truth)
		}
	}
}

func TestSupergateCompositionParetoAndOwnership(t *testing.T) {
	roots := superTestRoots()
	lib, err := PrecomputeSupergates(roots, SuperOptions{Name: "cells.genlib", Variables: 3, Levels: 2, MaxAttempts: 10000})
	if err != nil {
		t.Fatal(err)
	}
	checkSuperTruth(t, lib)
	if len(lib.LibraryIDs) == 0 || lib.Levels != 2 || lib.Tried == 0 {
		t.Fatalf("empty computation: %+v", lib)
	}
	// Constructor ownership: changing caller roots cannot change the library.
	roots[0].Name = "changed"
	roots[0].PinDelays[0] = 99
	if lib.Roots[0].Name != "AND2" || lib.Roots[0].PinDelays[0] != 1 {
		t.Fatal("primitive roots were borrowed")
	}
	for _, id := range lib.LibraryIDs {
		gate := lib.Gates[id]
		if !superConsecutiveSupport(gate.PinDelays) {
			t.Fatalf("non-prefix support gate retained: %d %+v", id, gate.PinDelays)
		}
		if gate.Truth == 0 || gate.Truth == superMask64(1<<uint(lib.Variables)) {
			t.Fatalf("constant library gate %d", id)
		}
	}
}

func superMask64(bits int) uint64 {
	if bits == 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(bits) - 1
}

func TestSupergateParetoFrontier(t *testing.T) {
	roots := []SuperPrimitiveGate{
		{Name: "SMALL", Fanins: 2, Truth: 8, Area: 1, PinDelays: []float64{3, 3}},
		{Name: "FAST", Fanins: 2, Truth: 8, Area: 3, PinDelays: []float64{1, 1}},
		{Name: "WORSE", Fanins: 2, Truth: 8, Area: 4, PinDelays: []float64{4, 4}},
	}
	lib, err := PrecomputeSupergates(roots, SuperOptions{Variables: 2, Levels: 1})
	if err != nil {
		t.Fatal(err)
	}
	andCount := 0
	for _, id := range lib.LibraryIDs {
		if lib.Gates[id].Truth == 8 {
			andCount++
			if lib.Roots[lib.Gates[id].Root].Name == "WORSE" {
				t.Fatal("dominated gate retained")
			}
		}
	}
	if andCount != 2 || lib.Lookups == 0 {
		t.Fatalf("AND Pareto count %d, lookups %d", andCount, lib.Lookups)
	}
}

func TestSupergateClassLimitRetainsExistingParetoClass(t *testing.T) {
	roots := []SuperPrimitiveGate{
		{Name: "SMALL", Fanins: 2, Truth: 8, Area: 1, PinDelays: []float64{3, 3}},
		{Name: "FAST", Fanins: 2, Truth: 8, Area: 3, PinDelays: []float64{1, 1}},
	}
	lib, err := PrecomputeSupergates(roots, SuperOptions{Variables: 2, Levels: 1, MaxGates: 3})
	if err != nil {
		t.Fatal(err)
	}
	andCount := 0
	for _, id := range lib.LibraryIDs {
		if lib.Gates[id].Truth == 8 {
			andCount++
		}
	}
	if andCount != 2 {
		t.Fatalf("class limit dropped Pareto alternatives in existing class: %d", andCount)
	}
}

func TestSupergateSerialization(t *testing.T) {
	lib, err := PrecomputeSupergates(superTestRoots()[:2], SuperOptions{Name: "demo", Variables: 2, Levels: 1})
	if err != nil {
		t.Fatal(err)
	}
	tree, err := lib.MarshalTree()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(tree, []byte("demo\n2\n")) || !bytes.Contains(tree, []byte("* AND2")) || !bytes.Contains(tree, []byte("* OR2")) {
		t.Fatalf("tree:\n%s", tree)
	}
	old, err := lib.MarshalOld()
	if err != nil || !bytes.Contains(old, []byte("AND2(a,b)")) || !bytes.Contains(old, []byte("OR2(a,b)")) {
		t.Fatalf("old:\n%s\n%v", old, err)
	}
	tree2, _ := lib.MarshalTree()
	if !bytes.Equal(tree, tree2) {
		t.Fatal("tree output changed across calls")
	}
}

func TestSupergateLimitsCancellationAndValidation(t *testing.T) {
	valid := superTestRoots()
	bad := [][]SuperPrimitiveGate{
		nil,
		{{Name: "", Fanins: 1, Truth: 1, Area: 1, PinDelays: []float64{1}}},
		{{Name: "X", Fanins: 0, Truth: 0, Area: 1}},
		{{Name: "X", Fanins: 2, Truth: 16, Area: 1, PinDelays: []float64{1, 1}}},
		{{Name: "X", Fanins: 1, Truth: 1, Area: -1, PinDelays: []float64{1}}},
	}
	for i, roots := range bad {
		if lib, err := PrecomputeSupergates(roots, SuperOptions{Variables: 2, Levels: 1}); lib != nil || !errors.Is(err, ErrDimension) {
			t.Fatalf("bad %d = %#v, %v", i, lib, err)
		}
	}
	if lib, err := PrecomputeSupergates(valid, SuperOptions{Variables: 3, Levels: 1, MaxGates: 2}); lib != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("too-small gate limit = %#v, %v", lib, err)
	}
	if lib, err := PrecomputeSupergates(valid, SuperOptions{Variables: 2, Levels: 2, MaxAttempts: 1}); lib != nil || !errors.Is(err, ErrLimit) {
		t.Fatalf("attempt limit = %#v, %v", lib, err)
	}
	if lib, err := PrecomputeSupergates(valid, SuperOptions{Variables: 2, Levels: 1, Stop: func() bool { return true }}); lib != nil || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel = %#v, %v", lib, err)
	}
	limited, err := PrecomputeSupergates(valid, SuperOptions{Variables: 2, Levels: 2, DelayMax: 0.1, AreaMax: 0.1})
	if err != nil || len(limited.LibraryIDs) != 0 {
		t.Fatalf("filtered = %d, %v", len(limited.LibraryIDs), err)
	}
	var nilLib *SuperLibrary
	if _, err = nilLib.MarshalTree(); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil tree = %v", err)
	}
}

func TestSuperCommandParsing(t *testing.T) {
	catalog := SuperCommandCatalog()
	if len(catalog) != 2 || catalog[0].Name != "super" || catalog[1].Name != "super2" {
		t.Fatalf("catalog = %+v", catalog)
	}
	catalog[0].Name = "changed"
	if SuperCommandCatalog()[0].Name != "super" {
		t.Fatal("command catalog storage was borrowed")
	}
	and, err := ParseSuper2Command([]string{"-I3", "-L", "2", "-vv"})
	if err != nil || and.Inputs != 3 || and.Levels != 2 || and.Verbose {
		t.Fatalf("super2 = %+v, %v", and, err)
	}
	command, err := ParseSuperCommand([]string{"-I", "4", "-L", "3", "-N", "50", "-T", "99", "-D", "2.5", "-A", "3.5", "-E", "skip", "-s", "-o", "-v", "lib.genlib"})
	if err != nil || command.Variables != 4 || command.Levels != 3 || command.MaxGates != 50 || command.TimeLimitSeconds != 99 || command.DelayMax != 2.5 || command.AreaMax != 3.5 || command.ExcludeName != "skip" || command.SkipInputInverters || !command.WriteOld || !command.Verbose || command.InputName != "lib.genlib" {
		t.Fatalf("super = %+v, %v", command, err)
	}
	for _, args := range [][]string{{}, {"-I"}, {"-I", "7", "x"}, {"a", "b"}, {"-D", "0", "x"}, {"-z", "x"}} {
		if _, err = ParseSuperCommand(args); !errors.Is(err, ErrDimension) {
			t.Fatalf("args %q = %v", args, err)
		}
	}
}

func TestSupergateModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state = state*6364136223846793005 + 1442695040888963407
		inputs := int(state%4) + 1
		levels := int(state>>8) & 1
		lib, err := PrecomputeAndSupergates(inputs, levels, AndSuperOptions{})
		if err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
		checkAndSuperTruth(t, lib)
	}
}

func FuzzSupergateTruth(f *testing.F) {
	for _, seed := range []struct{ variables, levels, mode uint8 }{{2, 1, 0}, {3, 1, 1}, {4, 0, 2}} {
		f.Add(seed.variables, seed.levels, seed.mode)
	}
	f.Fuzz(func(t *testing.T, variables, levels, mode uint8) {
		vars := int(variables%4) + 2
		lev := int(levels % 2)
		roots := superTestRoots()
		if mode&1 != 0 {
			roots = roots[:2]
		}
		lib, err := PrecomputeSupergates(roots, SuperOptions{Variables: vars, Levels: lev, MaxAttempts: 20000})
		if err != nil {
			t.Fatal(err)
		}
		checkSuperTruth(t, lib)
	})
}
