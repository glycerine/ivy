package goivy

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseIvyCheckParams extracts leading key=value parameters exactly like
// Python ivy_init.read_params: only leading args containing "=" are parsed, and
// an argument with more than one "=" is a usage error.
func ParseIvyCheckParams(args []string) (map[string]string, []string, error) {
	params := make(map[string]string)
	remaining := args
	for len(remaining) > 0 && strings.Contains(remaining[0], "=") {
		parts := strings.Split(remaining[0], "=")
		if len(parts) > 2 {
			return nil, nil, fmt.Errorf("bad parameter: %s", remaining[0])
		}
		params[parts[0]] = parts[1]
		remaining = remaining[1:]
	}
	return params, remaining, nil
}

// ApplyIvyCheckParams maps goivy_check-style key=value parameters onto cfg.
// cmd/goivy_check and the browser wasm entrypoint share this function so the
// two front doors do not drift.
func ApplyIvyCheckParams(cfg *Config, params map[string]string) error {
	for key, val := range params {
		switch key {
		case "diagnose":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.Diagnose = parsed
		case "coverage":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.Coverage = parsed
		case "action":
			cfg.CheckedAction = val
		case "trusted":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptTrusted = parsed
		case "mc":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptMC = parsed
		case "trace":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptTrace = parsed
		case "separate":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptSeparate = parsed
			cfg.OptSeparateSet = true
		case "isolate":
			cfg.Isolate = val
		case "summary":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptSummary = parsed
		case "unprovable":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OnlyCheckUnprovable = parsed
		case "seed":
			parsed, err := parseIvyCheckInt(key, val)
			if err != nil {
				return err
			}
			cfg.SolverOpts.Seed = parsed
			cfg.SolverOpts.SeedSet = true
		case "incremental":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.SolverOpts.Incremental = parsed
		case "show_vcs":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.SolverOpts.ShowVCs = parsed
		case "detailed":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.TraceDetailed = parsed
		case "use_numerals":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IuCfg.UseNumerals = parsed
		case "new_ui":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IuCfg.UseNewUI = parsed
		case "catch":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IuCfg.Catch = parsed
		case "ui":
			cfg.IuCfg.DefaultUI = val
		case "debug":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IuCfg.EnableDebug = parsed
		case "mode":
			if !isIvyCheckMode(val) {
				return fmt.Errorf("bad parameter value: %s=%s", key, val)
			}
			cfg.IuCfg.DefaultMode = val
		case "fullqi":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.FullQI = parsed
		case "show_compiled":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.ShowCompiled = parsed
		case "coi":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.ConeOfInfluence = parsed
		case "filter_symbols":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.FilterSymbols = parsed
		case "create_imports":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.CreateImports = parsed
		case "enforce_axioms":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.EnforceAxioms = parsed
		case "interference":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.DoCheckInterference = parsed
		case "pedantic":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.Pedantic = parsed
		case "prefer_impls":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.PreferImpls = parsed
		case "keep_destructors":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.KeepDestructors = parsed
		case "isolate_mode":
			cfg.IsolateCfg.IsolateMode = val
		case "compile_with_invariants":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.CompileWithInvariants = parsed
		case "assume_invariants":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.IsolateCfg.AssumeInvariants = parsed
		case "ext":
			cfg.ExtAction = val
			cfg.IsolateCfg.ExtAction = val
		case "unchecked_properties":
			cfg.OptUncheckedProps = val
		case "ivy_stats":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptIvyStats = parsed
		case "prioritize":
			cfg.PriorityActions = val
			cfg.PriorityActionsSet = true
		case "no_check_guarantees":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.NoCheckGuarantees = parsed
		case "profile":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.Profiling = parsed
		case "macro_finder":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.SolverOpts.MacroFinder = parsed
		case "complete":
			if err := validateIvyCheckCompleteLogics(val); err != nil {
				return err
			}
			cfg.CompleteLogic = val
		case "mutax":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptMutax = parsed
		case "l2s_debug":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.L2SDebug = parsed
		case "ranking_debug":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.RankingDebug = parsed
		case "abs_init":
			parsed, err := parseIvyCheckBool(key, val)
			if err != nil {
				return err
			}
			cfg.OptionAbsInit = parsed
		case "assert":
			parsed, err := parseIvyCheckAssertLocation(val)
			if err != nil {
				return err
			}
			cfg.CheckLineno = parsed
		case "parser":
			return fmt.Errorf("unknown parameter: %s", key)
		default:
			return fmt.Errorf("unknown parameter: %s", key)
		}
	}
	return nil
}

func parseIvyCheckInt(key, s string) (int, error) {
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("bad parameter value: %s=%s", key, s)
	}
	return parsed, nil
}

func parseIvyCheckAssertLocation(s string) (string, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("bad parameter value: assert=%s", s)
	}
	line, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("bad parameter value: assert=%s", s)
	}
	return fmt.Sprintf("%s.ivy:%d", parts[0], line), nil
}

func parseIvyCheckBool(key, s string) (bool, error) {
	switch s {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("bad parameter value: %s=%s", key, s)
	}
}

func isIvyCheckMode(s string) bool {
	switch s {
	case "abstract", "concrete", "bounded", "induction", "pdr":
		return true
	default:
		return false
	}
}

func validateIvyCheckCompleteLogics(s string) error {
	parts := strings.Split(s, ",")
	for _, part := range parts {
		if !KnownLogics[part] {
			return fmt.Errorf("bad parameter value: complete=%s", s)
		}
	}
	return nil
}
