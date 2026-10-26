package abc

// Standalone tests for the active boundary in src/base/io/ioReadEdif.c.

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

func TestABCEDIFAlwaysUnsupportedBeforeInputAndOptions(t *testing.T) {
	inputs := [][]byte{
		nil,
		{},
		[]byte("not EDIF"),
		[]byte("(edif sample (library gates) (design top))"),
		make([]byte, 1<<16),
	}
	for _, input := range inputs {
		for _, check := range []bool{false, true} {
			var stops int
			limits := IOLimits{
				MaxInputBytes: -1, MaxOutputBytes: -1, MaxObjects: -1,
				MaxEdges: -1, MaxCubes: -1, MaxOperations: -1,
				MaxBytes: -1, MaxDepth: -1, MaxFanins: -1,
				Stop: func() bool { stops++; return true },
			}
			network, err := ReadABCEDIF(input, check, limits)
			if network != nil || !errors.Is(err, ErrUnsupported) || stops != 0 {
				t.Fatalf("len=%d check=%v: network=%v err=%v stops=%d", len(input), check, network, err, stops)
			}
		}
	}
}

func TestABCEDIFConcurrentCalls(t *testing.T) {
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 64; worker++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			network, err := ReadABCEDIF([]byte{byte(index)}, index&1 != 0, IOLimits{Stop: func() bool { return true }})
			if network != nil || !errors.Is(err, ErrUnsupported) {
				failures.Add(1)
			}
		}(worker)
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures: %d", failures.Load())
	}
}

func FuzzABCEDIFUnsupported(f *testing.F) {
	f.Add([]byte{}, false, 0)
	f.Add([]byte("(edif x)"), true, -1)
	f.Add([]byte{0, 1, 2, 255}, false, 1)
	f.Fuzz(func(t *testing.T, data []byte, check bool, limit int) {
		calls := 0
		network, err := ReadABCEDIF(data, check, IOLimits{MaxInputBytes: limit, MaxObjects: limit, Stop: func() bool { calls++; return true }})
		if network != nil || !errors.Is(err, ErrUnsupported) || calls != 0 {
			t.Fatalf("network=%v err=%v stop calls=%d", network, err, calls)
		}
	})
}
