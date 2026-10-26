package abc

// Default parameter records from src/opt/fsim/fsimCore.c.

// FsimSimulationParams controls fast sequential miter simulation.
type FsimSimulationParams struct {
	Words       int
	Iterations  int
	TimeLimit   int
	CheckMiter  bool
	Verbose     bool
	CompressAIG bool
}

// DefaultFsimSimulationParams returns the native simulation defaults.
func DefaultFsimSimulationParams() FsimSimulationParams {
	return FsimSimulationParams{
		Words:      8,
		Iterations: 32,
		TimeLimit:  60,
		Verbose:    true,
	}
}

// FsimSwitchParams controls fast switching-activity estimation.
type FsimSwitchParams struct {
	Words              int
	Iterations         int
	PrefixFrames       int
	RandomPITransition int
	ProbabilityOne     bool
	ProbabilitySwitch  bool
	Verbose            bool
}

// DefaultFsimSwitchParams returns the native switching defaults.
func DefaultFsimSwitchParams() FsimSwitchParams {
	return FsimSwitchParams{
		Words:             1,
		Iterations:        48,
		PrefixFrames:      16,
		ProbabilityOne:    true,
		ProbabilitySwitch: true,
		Verbose:           true,
	}
}
