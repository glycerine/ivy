package abc

// Typed configuration constants from sat/glucose/Constants.h,
// sat/glucose2/Constants.h, and sat/glucose2/CGlucose.h.

const (
	GlucoseDynamicLevelCount       = true
	GlucoseConstantClauseRemoval   = true
	GlucoseVariableActivityUpdates = true
	GlucoseRemoveClauseRatio       = 2
	GlucoseBlockingRestartLower    = 10000
	Glucose2CircuitExperiment      = 1
)

type GlucoseConstantConfig struct {
	DynamicLevelCount, ConstantClauseRemoval, VariableActivityUpdates bool
	RemoveClauseRatio, BlockingRestartLower                           int
}

func DefaultGlucoseConstantConfig() GlucoseConstantConfig {
	return GlucoseConstantConfig{
		DynamicLevelCount:       GlucoseDynamicLevelCount,
		ConstantClauseRemoval:   GlucoseConstantClauseRemoval,
		VariableActivityUpdates: GlucoseVariableActivityUpdates,
		RemoveClauseRatio:       GlucoseRemoveClauseRatio,
		BlockingRestartLower:    GlucoseBlockingRestartLower,
	}
}
