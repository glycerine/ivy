package abc

// Go-only format, round-trip, concurrency and fuzz tests for abcLog.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func abcLogTrace(t testing.TB, registers, inputs, frames int) *Counterexample {
	t.Helper()
	trace, err := NewCounterexample(registers, inputs, frames)
	if err != nil {
		t.Fatal(err)
	}
	trace.Property = 3
	for bit := 0; bit < trace.BitCount(); bit++ {
		trace.set(bit, bit%3 == 1)
	}
	return trace
}

func TestABCLogExactFormatAndRoundTrips(t *testing.T) {
	trace := abcLogTrace(t, 2, 2, 3)
	data, err := MarshalABCLog(ABCLogRecord{Status: ABCLogSAT, BugFreeDepth: 4, Engine: "pdr", Counterexample: trace}, ABCLogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), "snl_SAT 4 pdr 3 2\n01\n001001\n"; got != want {
		t.Fatalf("log=%q want=%q", got, want)
	}
	parsed, err := ParseABCLog(data, ABCLogOptions{})
	if err != nil || parsed.Status != ABCLogSAT || parsed.BugFreeDepth != 4 || parsed.Engine != "pdr" || parsed.Counterexample == nil || parsed.Counterexample.Property != 3 || parsed.Counterexample.Registers() != 2 || parsed.Counterexample.Inputs() != 2 || parsed.Counterexample.Frames() != 3 || !slices.Equal(parsed.Counterexample.PackedWords(), trace.PackedWords()) {
		t.Fatalf("parsed=%+v trace=%v err=%v", parsed, parsed.Counterexample, err)
	}
	for _, status := range []ABCLogStatus{ABCLogUNSAT, ABCLogUnknown} {
		encoded, marshalErr := MarshalABCLog(ABCLogRecord{Status: status, BugFreeDepth: 7}, ABCLogOptions{})
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		decoded, parseErr := ParseABCLog(encoded, ABCLogOptions{})
		if parseErr != nil || decoded.Status != status || decoded.BugFreeDepth != 7 || decoded.Engine != "unknown" || decoded.Counterexample != nil {
			t.Fatalf("status=%d decoded=%+v err=%v text=%q", status, decoded, parseErr, encoded)
		}
	}
}

func TestABCLogZeroInputsOwnershipAndConcurrency(t *testing.T) {
	trace := abcLogTrace(t, 3, 0, 2)
	data, err := MarshalABCLog(ABCLogRecord{Status: ABCLogSAT, BugFreeDepth: 1, Engine: "bmc", Counterexample: trace}, ABCLogOptions{})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseABCLog(data, ABCLogOptions{})
	if err != nil || parsed.Counterexample == nil || parsed.Counterexample.Registers() != 3 || parsed.Counterexample.Inputs() != 0 || parsed.Counterexample.Frames() != 2 {
		t.Fatalf("zero-input parsed=%+v err=%v text=%q", parsed, err, data)
	}
	parsed.Counterexample.SetInitial(0, !parsed.Counterexample.bit(0))
	again, err := ParseABCLog(data, ABCLogOptions{})
	if err != nil || slices.Equal(parsed.Counterexample.PackedWords(), again.Counterexample.PackedWords()) {
		t.Fatalf("parse result aliases input/state err=%v", err)
	}
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 32; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			decoded, runErr := ParseABCLog(data, ABCLogOptions{})
			if runErr != nil || decoded.Counterexample == nil || decoded.Counterexample.Registers() != 3 {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func TestABCLogMalformedLimitsAndCancellation(t *testing.T) {
	for _, test := range []struct {
		data []byte
		want error
	}{
		{nil, ErrDimension},
		{[]byte("bad 0 x\nNULL\nNULL\n"), ErrDimension},
		{[]byte("snl_SAT x pdr\nNULL\nNULL\n"), ErrDimension},
		{[]byte("snl_SAT 0 pdr 0 -1\nNULL\nNULL\n"), ErrDimension},
		{[]byte("snl_SAT 0 pdr\nNULL\n0\n"), ErrDimension},
		{[]byte("snl_SAT 0 pdr\n-\n0\n"), ErrUnsupported},
		{[]byte("snl_SAT 1 pdr 0 1\n0\n010\n"), ErrDimension},
		{[]byte("snl_UNSAT 0 pdr extra\nNULL\nNULL\n"), ErrDimension},
		{[]byte("snl_SAT 0 pdr\nNULL\nNULL"), nil},
	} {
		_, err := ParseABCLog(test.data, ABCLogOptions{})
		if test.want == nil {
			if err != nil {
				t.Fatalf("data=%q err=%v", test.data, err)
			}
		} else if !errors.Is(err, test.want) {
			t.Fatalf("data=%q err=%v want=%v", test.data, err, test.want)
		}
	}
	valid := []byte("snl_UNSAT 1 pdr\nNULL\nNULL\n")
	for _, options := range []ABCLogOptions{{MaxBytes: 1}, {MaxFrames: 1}, {MaxEngineBytes: 1}, {Stop: func() bool { return true }}} {
		_, err := ParseABCLog(valid, options)
		if !(errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatalf("options=%+v err=%v", options, err)
		}
	}
	trace := abcLogTrace(t, 1, 1, 1)
	for _, record := range []ABCLogRecord{{Status: 9}, {Status: ABCLogSAT, BugFreeDepth: -1}, {Status: ABCLogSAT, Engine: "two words"}, {Status: ABCLogSAT, Counterexample: trace}} {
		if record.Counterexample == trace {
			record.BugFreeDepth = 0
		}
		_, err := MarshalABCLog(record, ABCLogOptions{MaxBits: 1})
		if err == nil {
			t.Fatalf("record=%+v accepted", record)
		}
	}
}

func FuzzABCLogRoundTrip(f *testing.F) {
	f.Add(uint8(2), uint8(3), uint8(2), uint8(0), []byte{1, 0, 1, 1})
	f.Add(uint8(0), uint8(0), uint8(1), uint8(1), []byte{})
	f.Fuzz(func(t *testing.T, registerByte, inputByte, frameByte, statusByte uint8, data []byte) {
		registers := int(registerByte % 5)
		inputs := int(inputByte % 5)
		frames := 1 + int(frameByte%5)
		trace, err := NewCounterexample(registers, inputs, frames)
		if err != nil {
			t.Fatal(err)
		}
		trace.Property = int(statusByte % 4)
		for bit := 0; bit < trace.BitCount(); bit++ {
			if len(data) != 0 {
				trace.set(bit, data[bit%len(data)]&(1<<uint(bit&7)) != 0)
			}
		}
		status := []ABCLogStatus{ABCLogSAT, ABCLogUNSAT, ABCLogUnknown}[statusByte%3]
		record := ABCLogRecord{Status: status, BugFreeDepth: frames - 1, Engine: "engine", Counterexample: trace}
		encoded, err := MarshalABCLog(record, ABCLogOptions{MaxBytes: 4096, MaxFrames: 16, MaxBits: 256})
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := ParseABCLog(encoded, ABCLogOptions{MaxBytes: 4096, MaxFrames: 16, MaxBits: 256})
		if err != nil || parsed.Counterexample == nil || parsed.Status != status || parsed.BugFreeDepth != record.BugFreeDepth || !slices.Equal(parsed.Counterexample.PackedWords(), trace.PackedWords()) || parsed.Counterexample.Registers() != registers || parsed.Counterexample.Inputs() != inputs || parsed.Counterexample.Frames() != frames {
			t.Fatalf("parsed=%+v trace=%v err=%v encoded=%q", parsed, trace, err, encoded)
		}
	})
}

func FuzzABCLogParser(f *testing.F) {
	f.Add([]byte("snl_SAT 2 pdr 0\n01\n101\n"))
	f.Add([]byte("snl_UNK 0 unknown\nNULL\nNULL\n"))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		options := ABCLogOptions{MaxBytes: 4096, MaxFrames: 32, MaxBits: 512, MaxEngineBytes: 128}
		record, err := ParseABCLog(data, options)
		if err != nil {
			return
		}
		encoded, err := MarshalABCLog(record, options)
		if err != nil {
			t.Fatalf("accepted record did not marshal: %+v err=%v", record, err)
		}
		again, err := ParseABCLog(encoded, options)
		if err != nil || again.Status != record.Status || again.BugFreeDepth != record.BugFreeDepth || again.Engine != record.Engine || (again.Counterexample == nil) != (record.Counterexample == nil) {
			t.Fatalf("normalized parse failed: record=%+v again=%+v err=%v encoded=%q", record, again, err, encoded)
		}
		if record.Counterexample != nil && (!slices.Equal(again.Counterexample.PackedWords(), record.Counterexample.PackedWords()) || again.Counterexample.Registers() != record.Counterexample.Registers() || again.Counterexample.Inputs() != record.Counterexample.Inputs() || again.Counterexample.Frames() != record.Counterexample.Frames()) {
			t.Fatalf("normalized counterexample mismatch: record=%v again=%v", record.Counterexample, again.Counterexample)
		}
	})
}
