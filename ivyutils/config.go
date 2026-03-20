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
	c.Diagnose = NewBooleanParameter("diagnose", false)
	c.Coverage = NewBooleanParameter("coverage", true)
	c.CheckedAction = NewParameter("action", "")
	c.OptTrusted = NewBooleanParameter("trusted", false)
	c.OptMC = NewBooleanParameter("mc", false)
	c.OptTrace = NewBooleanParameter("trace", false)
	c.OptSeparate = NewParameter("separate", nil)
	c.OptUncheckedProps = NewParameter("unchecked_properties", nil)
	c.OptIvyStats = NewBooleanParameter("ivy_stats", false)
	c.PriorityActions = NewParameter("prioritize", nil)
	c.NoCheckGuarantees = NewBooleanParameter("no_check_guarantees", false)
	c.Profiling = NewBooleanParameter("profile", false)
	c.OptSummary = NewBooleanParameter("summary", false)
	// CheckUnprovable corresponds to Python's act.check_unprovable
	// (ivy_actions.py:25). When true, only unprovable assertions are checked.
	c.CheckUnprovable = NewBooleanParameter("unprovable", false)
	return c
}
