package ivyutils

/*
type Config struct {
	Diagnose          bool   `json:"diagnose"`
	Coverage          bool   `json:"coverage"`
	CheckedAction     string `json:"action"`
	OptTrusted        bool   `json:"trusted"`
	OptMC             bool   `json:"mc"`
	OptTrace          bool   `json:"trace"`
	OptSeparate       any    `json:"separate"`
	OptUncheckedProps any    `json:"unchecked_properties"`
	OptIvyStats       bool   `json:"ivy_stats"`
	PriorityActions   any    `json:"prioritize"`
	NoCheckGuarantees bool   `json:"no_check_guarantees"`
	Profiling         bool   `json:"profile"`
	OptSummary        bool   `json:"summary"`

	// CheckUnprovable corresponds to Python's act.check_unprovable
	// (ivy_actions.py:25). When true, only unprovable assertions are checked.
	CheckOnlyUnprovable bool `json:"unprovable"`
}

func NewConfig() *Config {
	return &Config{
		Coverage: true,
	}
}
*/

type Config struct {
	// ParamRegistry is the per-config parameter registry. If nil, the
	// global Registry is used.
	ParamRegistry *ParameterRegistry

	Diagnose          *Parameter
	Coverage          *Parameter
	CheckedAction     *Parameter
	OptTrusted        *Parameter
	OptMC             *Parameter
	OptTrace          *Parameter
	OptSeparate       *Parameter
	OptUncheckedProps *Parameter
	OptIvyStats       *Parameter
	PriorityActions   *Parameter
	NoCheckGuarantees *Parameter
	Profiling         *Parameter
	OptSummary        *Parameter

	// CheckUnprovable corresponds to Python's act.check_unprovable
	// (ivy_actions.py:25). When true, only unprovable assertions are checked.
	CheckUnprovable *Parameter

	// moved from top of check/check.go

	// Failures tracks the number of failed checks during verification.
	Failures int

	// CheckedActionFound tracks whether a checked action was found.
	CheckedActionFound bool

	// CheckLineno is the current line number being checked, or empty for all.
	CheckLineno string
}

func NewConfig() (c *Config) {
	c = &Config{}
	c.ParamRegistry = NewParameterRegistry()
	c.Diagnose = NewBooleanParameterOn(c.ParamRegistry, "diagnose", false)
	c.Coverage = NewBooleanParameterOn(c.ParamRegistry, "coverage", true)
	c.CheckedAction = NewParameterOn(c.ParamRegistry, "action", "")
	c.OptTrusted = NewBooleanParameterOn(c.ParamRegistry, "trusted", false)
	c.OptMC = NewBooleanParameterOn(c.ParamRegistry, "mc", false)
	c.OptTrace = NewBooleanParameterOn(c.ParamRegistry, "trace", false)
	c.OptSeparate = NewParameterOn(c.ParamRegistry, "separate", nil)
	c.OptUncheckedProps = NewParameterOn(c.ParamRegistry, "unchecked_properties", nil)
	c.OptIvyStats = NewBooleanParameterOn(c.ParamRegistry, "ivy_stats", false)
	c.PriorityActions = NewParameterOn(c.ParamRegistry, "prioritize", nil)
	c.NoCheckGuarantees = NewBooleanParameterOn(c.ParamRegistry, "no_check_guarantees", false)
	c.Profiling = NewBooleanParameterOn(c.ParamRegistry, "profile", false)
	c.OptSummary = NewBooleanParameterOn(c.ParamRegistry, "summary", false)
	// CheckUnprovable corresponds to Python's act.check_unprovable
	// (ivy_actions.py:25). When true, only unprovable assertions are checked.
	c.CheckUnprovable = NewBooleanParameterOn(c.ParamRegistry, "unprovable", false)
	return c
}
