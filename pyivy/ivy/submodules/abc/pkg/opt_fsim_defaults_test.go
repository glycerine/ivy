package abc

// Tests for src/opt/fsim/fsimCore.c.

import (
	"encoding/json"
	"os"
	"testing"
)

type fsimDefaultsOracle struct {
	Simulation struct {
		Words, Iterations, TimeLimit, CheckMiter, Verbose, CompressAIG int
	} `json:"simulation"`
	Switching struct {
		Words, Iterations, PrefixFrames, RandomPITransition int
		ProbabilityOne, ProbabilitySwitch, Verbose          int
	} `json:"switching"`
}

func TestFsimDefaultsNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/fsim_defaults.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle fsimDefaultsOracle
	if err = json.Unmarshal(data, &oracle); err != nil {
		t.Fatal(err)
	}
	simulation := DefaultFsimSimulationParams()
	if simulation.Words != oracle.Simulation.Words ||
		simulation.Iterations != oracle.Simulation.Iterations ||
		simulation.TimeLimit != oracle.Simulation.TimeLimit ||
		boolInt(simulation.CheckMiter) != oracle.Simulation.CheckMiter ||
		boolInt(simulation.Verbose) != oracle.Simulation.Verbose ||
		boolInt(simulation.CompressAIG) != oracle.Simulation.CompressAIG {
		t.Fatal(simulation, oracle.Simulation)
	}
	switching := DefaultFsimSwitchParams()
	if switching.Words != oracle.Switching.Words ||
		switching.Iterations != oracle.Switching.Iterations ||
		switching.PrefixFrames != oracle.Switching.PrefixFrames ||
		switching.RandomPITransition != oracle.Switching.RandomPITransition ||
		boolInt(switching.ProbabilityOne) != oracle.Switching.ProbabilityOne ||
		boolInt(switching.ProbabilitySwitch) != oracle.Switching.ProbabilitySwitch ||
		boolInt(switching.Verbose) != oracle.Switching.Verbose {
		t.Fatal(switching, oracle.Switching)
	}
	simulation.Words = -1
	switching.Iterations = -1
	if DefaultFsimSimulationParams().Words != oracle.Simulation.Words || DefaultFsimSwitchParams().Iterations != oracle.Switching.Iterations {
		t.Fatal("default records retained caller mutation")
	}
}
