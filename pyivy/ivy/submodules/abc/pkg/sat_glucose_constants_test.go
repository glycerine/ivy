package abc

// Exact contract tests for the shared Glucose configuration headers.

import "testing"

func TestGlucoseConstants(t *testing.T) {
	if !GlucoseDynamicLevelCount || !GlucoseConstantClauseRemoval || !GlucoseVariableActivityUpdates || GlucoseRemoveClauseRatio != 2 || GlucoseBlockingRestartLower != 10000 || Glucose2CircuitExperiment != 1 {
		t.Fatal(GlucoseDynamicLevelCount, GlucoseConstantClauseRemoval, GlucoseVariableActivityUpdates, GlucoseRemoveClauseRatio, GlucoseBlockingRestartLower, Glucose2CircuitExperiment)
	}
	want := GlucoseConstantConfig{true, true, true, 2, 10000}
	first := DefaultGlucoseConstantConfig()
	second := DefaultGlucoseConstantConfig()
	if first != want || second != want {
		t.Fatal(first, second, want)
	}
	first.RemoveClauseRatio = 9
	if second != want {
		t.Fatal("config values are not independent", first, second)
	}
}
