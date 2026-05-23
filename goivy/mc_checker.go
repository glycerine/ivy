package goivy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy/fileops"
	"github.com/glycerine/ivy/goivy/xtracer"
)

// CheckResult holds the result of a model checking run.
type MCCheckResult struct {
	Proved       bool                // true if property holds
	Trace        *WitnessTrace       // raw witness (non-nil if counterexample found)
	DecodedTrace *AigerMatchHandler2 // decoded Ivy trace (non-nil if counterexample decoded)
	Error        error               // non-nil if model checker failed
}

type MCCounterexampleFailure struct {
	Trace fmt.Stringer
	Cause error
}

func (e *MCCounterexampleFailure) Error() string {
	if e == nil {
		return "<nil model-check counterexample>"
	}
	if e.Trace != nil {
		return e.Trace.String()
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "model checking failed"
}

func (e *MCCounterexampleFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// ModelChecker describes an external hardware model checker.
type ModelChecker interface {
	// Cmd returns the command to run the model checker.
	// aigfilename is the path to the AIG/AAG file; outfilename receives a witness.
	Cmd(aigfilename, outfilename string) []string

	// Scrape examines model checker stdout and returns true when the property
	// was proved and no counterexample witness is expected.
	Scrape(alltext string) bool
}

// ABCModelChecker implements ModelChecker using the ABC tool.
type ABCModelChecker struct {
	// ABCPath overrides the default ABC binary path.
	ABCPath string
}

// Cmd returns the ABC command line for model checking.
func (c *ABCModelChecker) Cmd(aigfilename, outfilename string) []string {
	abcPath := c.ABCPath
	if abcPath == "" {
		exePath, err := os.Executable()
		if err == nil {
			abcPath = filepath.Join(filepath.Dir(exePath), "abc")
		} else {
			abcPath = "abc"
		}
	}
	return []string{abcPath, "-c",
		fmt.Sprintf("read_aiger %s; pdr; write_aiger_cex  %s", aigfilename, outfilename)}
}

// Scrape checks whether ABC's output indicates the property was proved.
func (c *ABCModelChecker) Scrape(alltext string) bool {
	return strings.Contains(alltext, "Property proved")
}

// RunABC runs ABC on the given AIGER string and returns the result.
func RunABC(aigerStr string, checker ModelChecker, mod *Module) (*MCCheckResult, error) {
	if checker == nil {
		checker = &ABCModelChecker{}
	}
	result, err := fileops.RunAigerModelChecker(aigerStr, checker.Cmd, checker.Scrape)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("model checker returned nil result")
	}

	if mod != nil && mod.Cfg.MCVerbose {
		fmt.Println(result.Output)
	}

	if result.Error != nil {
		return &MCCheckResult{Error: fmt.Errorf("model checker failed: %w", result.Error)}, nil
	}
	if result.Proved {
		return &MCCheckResult{Proved: true}, nil
	}

	trace, err := ParseWitnessBytes(result.Witness)
	if err != nil {
		return &MCCheckResult{Error: fmt.Errorf("failed to parse witness: %w", err)}, nil
	}

	return &MCCheckResult{Trace: trace}, nil
}

// CheckIsolate is the main entry point for model checking an isolate.
// It converts the module to AIGER, runs the ABC model checker, and on
// counterexample decodes the witness into an Ivy trace.
//
// Python: ivy_mc.py:1716-1802 check_isolate
func MCCheckIsolate(mod *Module, method string) (*MCCheckResult, error) {
	if method == "" {
		method = "mc"
	}
	xtracer.Trace("mc.CheckIsolate ENTER method=%s", method)

	if mod.Cfg.MCVerbose {
		fmt.Println()
		fmt.Println(strings.Repeat("*", 80))
		fmt.Println()
	}

	// Open logfile (Go uses goivy_mc.log to avoid conflicting with ivy_mc.log)
	logfile, err := os.OpenFile("goivy_mc.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		defer logfile.Close()
	}

	// Python: check_isolate lines 1733-1734
	// ext_acts = [mod.actions[x] for x in sorted(mod.public_actions)]
	// ext_act = ia.EnvAction(*ext_acts)
	// Note: to_aiger shadows this parameter, but the creation still increments the counter.
	pubNames := sortedPublicActions(mod)
	wastedExtActs := make([]Expr, 0, len(pubNames))
	for _, name := range pubNames {
		if act, ok := mod.Actions.Get2(name); ok {
			wastedExtActs = append(wastedExtActs, act)
		}
	}
	_ = NewEnvActionOn(mod.Cfg.ActCfg, wastedExtActs...)

	// Convert to AIGER
	result, err := ToAiger(mod, method)
	if err != nil {
		return nil, fmt.Errorf("to_aiger failed: %w", err)
	}

	// Get AIGER string
	aigerStr := result.Aiger.String()
	xtracer.Trace("mc.CheckIsolate postToAiger aigerLen=%d", len(aigerStr))

	if mod.Cfg.MCVerbose {
		fmt.Printf("aigtoaig_path:%s\n", "aigtoaig")
	}

	checker := &ABCModelChecker{}
	mcResult, err := fileops.RunAigerModelChecker(aigerStr, checker.Cmd, checker.Scrape)
	if err != nil {
		xtracer.Trace("mc.CheckIsolate EXIT proved=false err=%v", err)
		return nil, err
	}
	if mcResult == nil {
		err := fmt.Errorf("model checker returned nil result")
		xtracer.Trace("mc.CheckIsolate EXIT proved=false err=%v", err)
		return nil, err
	}
	if mcResult.Error != nil {
		xtracer.Trace("mc.CheckIsolate EXIT proved=false err=model checker failed")
		return nil, fmt.Errorf("failed to run model checker: %w", mcResult.Error)
	}

	if mod.Cfg.MCVerbose {
		fmt.Println("\nModel checker output:")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Print(mcResult.Output)
		fmt.Println(strings.Repeat("-", 80))
	}

	if mcResult.Proved {
		xtracer.Trace("mc.CheckIsolate EXIT proved=true err=<nil>")
		return &MCCheckResult{Proved: true}, nil
	}

	// Counterexample found — decode the witness into an Ivy trace
	xtracer.Trace("mc.CheckIsolate EXIT proved=false err=<nil>")
	decodedTrace, err := AigerWitnessToIvyTrace2Bytes(result, mcResult.Witness, mod)
	if err != nil {
		return &MCCheckResult{Error: fmt.Errorf("trace decode failed: %w", err)}, nil
	}

	return &MCCheckResult{DecodedTrace: decodedTrace}, nil
}
