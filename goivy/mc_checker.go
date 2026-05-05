package goivy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

// ModelChecker is the interface for hardware model checkers.
type ModelChecker interface {
	// Cmd returns the command to run the model checker.
	// aigfilename: path to the AIG file
	// outfilename: path for the witness output
	Cmd(aigfilename, outfilename string) []string

	// Scrape examines the model checker output and returns true if
	// the property was proved (no counterexample).
	Scrape(alltext string) bool
}

// ABCModelChecker implements ModelChecker using the ABC tool.
type ABCModelChecker struct {
	// ABCPath overrides the default ABC binary path.
	// If empty, the default path is used.
	ABCPath string
}

// Cmd returns the ABC command line for model checking.
func (mc *ABCModelChecker) Cmd(aigfilename, outfilename string) []string {
	abcPath := mc.ABCPath
	if abcPath == "" {
		// Default: look for abc next to the executable
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
func (mc *ABCModelChecker) Scrape(alltext string) bool {
	return strings.Contains(alltext, "Property proved")
}

// CheckResult holds the result of a model checking run.
type MCCheckResult struct {
	Proved       bool                // true if property holds
	Trace        *WitnessTrace       // raw witness (non-nil if counterexample found)
	DecodedTrace *AigerMatchHandler2 // decoded Ivy trace (non-nil if counterexample decoded)
	Error        error               // non-nil if model checker failed
}

// RunABC runs ABC on the given AIGER string and returns the result.
func RunABC(aigerStr string, mc ModelChecker, mod *Module) (*MCCheckResult, error) {
	if mc == nil {
		mc = &ABCModelChecker{}
	}

	// Write AIGER to temp file
	tmpFile, err := os.CreateTemp("", "ivy_mc_*.aag")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	aagName := tmpFile.Name()
	defer os.Remove(aagName)

	if _, err := tmpFile.WriteString(aigerStr); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write AIGER: %w", err)
	}
	tmpFile.Close()

	// Convert AAG to AIG (if aigtoaig is available)
	aigName := strings.TrimSuffix(aagName, ".aag") + ".aig"
	defer os.Remove(aigName)

	aigtoaigPath := "aigtoaig" // assume on PATH
	ret := exec.Command(aigtoaigPath, aagName, aigName)
	if err := ret.Run(); err != nil {
		// If aigtoaig fails, try using the AAG file directly
		aigName = aagName
	}

	// Run model checker
	outName := strings.TrimSuffix(aagName, ".aag") + ".out"
	defer os.Remove(outName)

	cmd := mc.Cmd(aigName, outName)
	if len(cmd) == 0 {
		return nil, fmt.Errorf("empty model checker command")
	}

	p := exec.Command(cmd[0], cmd[1:]...)
	var stdout bytes.Buffer
	p.Stdout = &stdout

	if err := p.Run(); err != nil {
		return &MCCheckResult{Error: fmt.Errorf("model checker failed: %w", err)}, nil
	}

	alltext := stdout.String()
	if mod != nil && mod.Cfg.MCVerbose {
		fmt.Println(alltext)
	}

	if mc.Scrape(alltext) {
		return &MCCheckResult{Proved: true}, nil
	}

	// Parse witness
	trace, err := ParseWitnessFile(outName)
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

	// Write AIGER to temp file
	tmpFile, err := os.CreateTemp("", "ivy_mc_*.aag")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	aagName := tmpFile.Name()
	defer os.Remove(aagName)

	if _, err := tmpFile.WriteString(aigerStr); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write AIGER: %w", err)
	}
	tmpFile.Close()

	// Convert AAG to AIG
	aigName := strings.TrimSuffix(aagName, ".aag") + ".aig"
	defer os.Remove(aigName)

	aigtoaigPath := "aigtoaig"
	if mod.Cfg.MCVerbose {
		fmt.Printf("aigtoaig_path:%s\n", aigtoaigPath)
	}
	ret := exec.Command(aigtoaigPath, aagName, aigName)
	if err := ret.Run(); err != nil {
		aigName = aagName
	}

	// Run model checker
	outName := strings.TrimSuffix(aagName, ".aag") + ".out"
	defer os.Remove(outName)

	checker := &ABCModelChecker{}
	cmd := checker.Cmd(aigName, outName)
	p := exec.Command(cmd[0], cmd[1:]...)
	var stdout bytes.Buffer
	p.Stdout = &stdout

	if err := p.Run(); err != nil {
		xtracer.Trace("mc.CheckIsolate EXIT proved=false err=model checker failed")
		return nil, fmt.Errorf("failed to run model checker: %w", err)
	}

	alltext := stdout.String()
	if mod.Cfg.MCVerbose {
		fmt.Println("\nModel checker output:")
		fmt.Println(strings.Repeat("-", 80))
		fmt.Print(alltext)
		fmt.Println(strings.Repeat("-", 80))
	}

	if p.ProcessState != nil && !p.ProcessState.Success() {
		return nil, fmt.Errorf("model checker returned non-zero status")
	}

	if checker.Scrape(alltext) {
		xtracer.Trace("mc.CheckIsolate EXIT proved=true err=<nil>")
		return &MCCheckResult{Proved: true}, nil
	}

	// Counterexample found — decode the witness into an Ivy trace
	xtracer.Trace("mc.CheckIsolate EXIT proved=false err=<nil>")
	decodedTrace, err := AigerWitnessToIvyTrace2(result, outName, mod)
	if err != nil {
		return &MCCheckResult{Error: fmt.Errorf("trace decode failed: %w", err)}, nil
	}

	return &MCCheckResult{DecodedTrace: decodedTrace}, nil
}
