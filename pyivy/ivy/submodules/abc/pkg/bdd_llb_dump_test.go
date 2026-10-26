package abc

// Standalone semantics, ownership, concurrency and fuzz tests for llb2Dump.c.

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func evalLLBReachedBLIF(data []byte, inputs map[string]bool) (bool, error) {
	return evalCascadeBLIF(bytes.ReplaceAll(data, []byte("Reached"), []byte("F")), inputs)
}

func llbBDDFromTruth(t testing.TB, variables int, values []bool) (*ExtraBDDManager, ExtraBDDLit) {
	t.Helper()
	manager, err := NewExtraBDDManager(variables, 1<<14)
	if err != nil {
		t.Fatal(err)
	}
	root, err := manager.analysisRun(nil, 1, ExtraBDDOptions{}, func(_ int, work *extraBDDWork) (ExtraBDDLit, error) {
		return manager.analysisBuild(values, work)
	})
	if err != nil {
		t.Fatal(err)
	}
	return manager, root
}

func TestLLBDummyName(t *testing.T) {
	tests := []struct {
		prefix string
		number int
		digits int
		want   string
	}{
		{"ff", 0, 1, "ff0"},
		{"ff", 3, 4, "ff0003"},
		{"", 12, 1, "12"},
		{"x", 12, 0, "x12"},
	}
	for _, test := range tests {
		got, err := LLBDummyName(test.prefix, test.number, test.digits)
		if err != nil || got != test.want {
			t.Fatalf("(%q,%d,%d)=%q,%v want %q", test.prefix, test.number, test.digits, got, err, test.want)
		}
	}
	for _, test := range []struct {
		prefix string
		number int
		digits int
	}{{"bad\x00name", 0, 1}, {"ff", -1, 1}, {"ff", 0, -1}, {"ff", 0, 19}} {
		if _, err := LLBDummyName(test.prefix, test.number, test.digits); !errors.Is(err, ErrDimension) {
			t.Fatalf("accepted %+v: %v", test, err)
		}
	}
}

func TestLLBMarshalReachedBLIFExactSemanticsAndOwnership(t *testing.T) {
	manager, root := llbBDDFromTruth(t, 2, []bool{false, true, true, false})
	data, err := LLBMarshalReachedBLIF(manager, root, "reached", LLBDumpOptions{})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, fragment := range []string{".model reached\n", ".inputs ff0 ff1\n", ".outputs Reached\n", "01 1\n", "10 1\n", ".end\n"} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("BLIF misses %q:\n%s", fragment, text)
		}
	}
	for assignment := 0; assignment < 4; assignment++ {
		got, evalErr := evalLLBReachedBLIF(data, map[string]bool{"ff0": assignment&1 != 0, "ff1": assignment&2 != 0})
		want, bddErr := manager.Evaluate(root, []bool{assignment&1 != 0, assignment&2 != 0})
		if evalErr != nil || bddErr != nil || got != want {
			t.Fatalf("assignment=%d got=%v want=%v evalErr=%v bddErr=%v", assignment, got, want, evalErr, bddErr)
		}
	}
	data[0] = 'X'
	again, err := LLBMarshalReachedBLIF(manager, root, "reached", LLBDumpOptions{})
	if err != nil || again[0] != '.' {
		t.Fatalf("output aliases retained storage: %q %v", again, err)
	}
}

func TestLLBMarshalReachedBLIFLimitsCancellationAndConcurrency(t *testing.T) {
	manager, root := llbBDDFromTruth(t, 3, []bool{false, true, true, false, true, false, false, true})
	baseline, err := LLBMarshalReachedBLIF(manager, root, "parity", LLBDumpOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, options := range []LLBDumpOptions{
		{MaxBytes: len(baseline) - 1}, {MaxNodes: 1}, {MaxOperations: 1},
	} {
		if data, dumpErr := LLBMarshalReachedBLIF(manager, root, "parity", options); data != nil || !errors.Is(dumpErr, ErrLimit) {
			t.Fatalf("options=%+v data=%q err=%v", options, data, dumpErr)
		}
	}
	if data, dumpErr := LLBMarshalReachedBLIF(manager, root, "parity", LLBDumpOptions{Stop: func() bool { return true }}); data != nil || !errors.Is(dumpErr, ErrCancelled) {
		t.Fatalf("cancel data=%q err=%v", data, dumpErr)
	}
	if data, dumpErr := LLBMarshalReachedBLIF(nil, root, "parity", LLBDumpOptions{}); data != nil || !errors.Is(dumpErr, ErrDimension) {
		t.Fatalf("nil data=%q err=%v", data, dumpErr)
	}
	if data, dumpErr := LLBMarshalReachedBLIF(manager, ExtraBDDLit(9999), "parity", LLBDumpOptions{}); data != nil || !errors.Is(dumpErr, ErrInvalidLiteral) {
		t.Fatalf("literal data=%q err=%v", data, dumpErr)
	}
	if data, dumpErr := LLBMarshalReachedBLIF(manager, root, "bad model", LLBDumpOptions{}); data != nil || !errors.Is(dumpErr, ErrDimension) {
		t.Fatalf("model data=%q err=%v", data, dumpErr)
	}

	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, runErr := LLBMarshalReachedBLIF(manager, root, "parity", LLBDumpOptions{})
			if runErr != nil || !slices.Equal(got, baseline) {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzLLBMarshalReachedBLIF(f *testing.F) {
	f.Add(uint8(2), uint64(6))
	f.Add(uint8(0), uint64(1))
	f.Fuzz(func(t *testing.T, variableByte uint8, truth uint64) {
		variables := int(variableByte % 6)
		values := make([]bool, 1<<uint(variables))
		for assignment := range values {
			values[assignment] = truth&(uint64(1)<<uint(assignment%64)) != 0
		}
		manager, root := llbBDDFromTruth(t, variables, values)
		data, err := LLBMarshalReachedBLIF(manager, root, "reached", LLBDumpOptions{MaxBytes: 1 << 20, MaxNodes: 1 << 16, MaxOperations: 1 << 20})
		if err != nil {
			t.Fatal(err)
		}
		for assignment, want := range values {
			inputs := make(map[string]bool, variables)
			for variable := 0; variable < variables; variable++ {
				name, _ := LLBDummyName("ff", variable, HopBase10Log(uint(variables)))
				inputs[name] = assignment&(1<<uint(variable)) != 0
			}
			got, evalErr := evalLLBReachedBLIF(data, inputs)
			if evalErr != nil || got != want {
				t.Fatalf("variables=%d assignment=%d got=%v want=%v err=%v\n%s", variables, assignment, got, want, evalErr, data)
			}
		}
	})
}
