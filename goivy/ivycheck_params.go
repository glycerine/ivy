package goivy

import (
	"fmt"
	"strings"
)

// ApplyIvyCheckParams maps goivy_check-style key=value parameters onto cfg.
// cmd/goivy_check and the browser wasm entrypoint share this function so the
// two front doors do not drift.
func ApplyIvyCheckParams(cfg *Config, params map[string]string) error {
	for key, val := range params {
		switch key {
		case "diagnose":
			cfg.Diagnose = parseIvyCheckBool(val)
		case "coverage":
			cfg.Coverage = parseIvyCheckBool(val)
		case "action":
			cfg.CheckedAction = val
		case "trusted":
			cfg.OptTrusted = parseIvyCheckBool(val)
		case "mc":
			cfg.OptMC = parseIvyCheckBool(val)
		case "trace":
			cfg.OptTrace = parseIvyCheckBool(val)
		case "separate":
			cfg.OptSeparate = parseIvyCheckBool(val)
			cfg.OptSeparateSet = true
		case "isolate":
			cfg.Isolate = val
		case "summary":
			cfg.OptSummary = parseIvyCheckBool(val)
		case "unprovable":
			cfg.OnlyCheckUnprovable = parseIvyCheckBool(val)
		case "unchecked_properties":
			cfg.OptUncheckedProps = val
		case "ivy_stats":
			cfg.OptIvyStats = parseIvyCheckBool(val)
		case "prioritize":
			cfg.PriorityActions = val
		case "no_check_guarantees":
			cfg.NoCheckGuarantees = parseIvyCheckBool(val)
		case "profile":
			cfg.Profiling = parseIvyCheckBool(val)
		case "macro_finder":
			cfg.SolverOpts.MacroFinder = parseIvyCheckBool(val)
		case "complete":
			cfg.CompleteLogic = val
		case "checked_assert":
			cfg.CheckLineno = val
		case "parser":
			panic("parser is no longer a choice; we only have the one now.")
		default:
			return fmt.Errorf("unknown parameter: %s", key)
		}
	}
	return nil
}

func parseIvyCheckBool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
