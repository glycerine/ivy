package abc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestABCTraceNativeSmokeFixtures(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/rebuilt_smoke.json")
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Cases []struct {
			Input          string  `json:"input_hex"`
			Trace          *string `json:"witness"`
			ExpectedResult string  `json:"expected"`
		}
	}
	// Explicit fields avoid conflating a missing witness with an empty one.
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 3 {
		t.Fatal("smoke fixture size")
	}
	for _, record := range report.Cases {
		if record.ExpectedResult == "proved" {
			if record.Trace != nil {
				t.Fatal("safe fixture has witness")
			}
			continue
		}
		raw, err := hex.DecodeString(record.Input)
		if err != nil {
			t.Fatal(err)
		}
		m, err := ReadAiger(raw, AigerLimits{})
		if err != nil {
			t.Fatal(err)
		}
		if record.Trace == nil {
			t.Fatal("missing unsafe witness")
		}
		c, err := m.ReadABCCounterexample([]byte(*record.Trace), 0, TraceLimits{})
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := m.MarshalABCCounterexample(c)
		if err != nil || string(encoded) != *record.Trace {
			t.Fatal("native witness round trip", string(encoded), err)
		}
	}
}
func TestABCTraceRejectsCorruption(t *testing.T) {
	m, err := ReadAiger([]byte("aig 2 1 1 1 0\n2\n4\n"), AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	good := []byte("1\n0 1 0 1\n1 0 1 0\n")
	for _, data := range [][]byte{
		nil, []byte("1\n"), []byte("0\n"), []byte("2\n0 1 0 1\n1 0 1 0\n"),
		[]byte("1\n0 1 0 1\n0 0 1 0\n"), // state discontinuity
		[]byte("1\n0 1 1 1\n1 0 1 0\n"), // wrong PO
		[]byte("1\n0 1 0 0\n1 0 1 0\n"), // wrong next state
		[]byte("1\n0 0 0 0\n"),          // no failing property
		[]byte("1\n1 0 1 0\n"),          // invalid initial state
		[]byte("1\n0 x 0 1\n"), []byte("1\n0  1 0 1\n"),
		append(slices.Clone(good), '\n'), append(slices.Clone(good), 'x'), good[:len(good)-1],
	} {
		if _, err := m.ReadABCCounterexample(data, 0, TraceLimits{}); err == nil {
			t.Fatalf("accepted corrupt witness %q", data)
		}
	}
	for _, limits := range []TraceLimits{{MaxBytes: 1}, {MaxFrames: 1}, {MaxFrames: -1}} {
		if _, err := m.ReadABCCounterexample(good, 0, limits); err == nil {
			t.Fatal("limit ignored", limits)
		}
	}
	if _, err := m.ReadABCCounterexample(good, 1, TraceLimits{}); err == nil {
		t.Fatal("invalid property")
	}
}
func TestAigerCounterexampleInitializationAndConstraints(t *testing.T) {
	m, err := ReadAiger([]byte("aig 2 1 1 0 0 1 1\n2 1\n4\n2\n"), AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	c := cexTestNew(t, 1, 1, 1)
	c.SetInitial(0, true)
	c.SetInput(0, 0, true)
	if err := m.ValidateCounterexample(c); err != nil {
		t.Fatal(err)
	}
	encoded, err := m.MarshalABCCounterexample(c)
	if err != nil || string(encoded) != "1\n1 1 11 1\n" {
		t.Fatal(string(encoded), err)
	}
	again, err := m.ReadABCCounterexample(encoded, 0, TraceLimits{})
	if err != nil || !slices.Equal(cexBits(t, c), cexBits(t, again)) {
		t.Fatal(err)
	}
	c.SetInitial(0, false)
	if err := m.ValidateCounterexample(c); err == nil {
		t.Fatal("wrong initialization accepted")
	}
	c.SetInitial(0, true)
	c.SetInput(0, 0, false)
	if err := m.ValidateCounterexample(c); err == nil {
		t.Fatal("false constraint accepted")
	}
	c.SetInput(0, 0, true)
	c.Property = 1
	if err := m.ValidateCounterexample(c); err == nil {
		t.Fatal("constraint treated as property")
	}
	unknown, err := ReadAiger([]byte("aig 1 0 1 1 0\n2 2\n2\n"), AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	u := cexTestNew(t, 1, 0, 1)
	u.SetInitial(0, true)
	if err := unknown.ValidateCounterexample(u); err != nil {
		t.Fatal("unknown init", err)
	}
}
func TestABCTraceEmptyVectors(t *testing.T) {
	for _, data := range []string{"aig 0 0 0 1 0\n1\n", "aig 1 1 0 1 0\n2\n"} {
		m, err := ReadAiger([]byte(data), AigerLimits{})
		if err != nil {
			t.Fatal(err)
		}
		c := cexTestNew(t, 0, m.Graph().PrimaryInputCount(), 1)
		if c.Inputs() > 0 {
			c.SetInput(0, 0, true)
		}
		encoded, err := m.MarshalABCCounterexample(c)
		if err != nil {
			t.Fatal(err)
		}
		d, err := m.ReadABCCounterexample(encoded, 0, TraceLimits{})
		if err != nil || d.Validate() != nil {
			t.Fatal(string(encoded), err)
		}
	}
}
func FuzzABCTraceRead(f *testing.F) {
	f.Add([]byte("1\n0 1 0 1\n1 0 1 0\n"))
	f.Add([]byte("1\n1 1 1 1\n"))
	f.Add([]byte("0\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		m, err := ReadAiger([]byte("aig 2 1 1 1 0\n2\n4\n"), AigerLimits{})
		if err != nil {
			t.Fatal(err)
		}
		c, err := m.ReadABCCounterexample(data, 0, TraceLimits{MaxBytes: 65536, MaxFrames: 32})
		if err != nil {
			return
		}
		if err := m.ValidateCounterexample(c); err != nil {
			t.Fatal(err)
		}
		encoded, err := m.MarshalABCCounterexample(c)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(encoded, data) {
			t.Fatal("noncanonical successful witness")
		}
	})
}
