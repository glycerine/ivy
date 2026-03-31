package mc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy/module"
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
		// Default: look for abc in a bin/ subdirectory of the executable
		exePath, err := os.Executable()
		if err == nil {
			abcPath = filepath.Join(filepath.Dir(exePath), "bin", "abc")
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
type CheckResult struct {
	Proved  bool          // true if property holds
	Trace   *WitnessTrace // non-nil if counterexample found
	Error   error         // non-nil if model checker failed
}

// RunABC runs ABC on the given AIGER string and returns the result.
func RunABC(aigerStr string, mc ModelChecker, mod *module.Module) (*CheckResult, error) {
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
		return &CheckResult{Error: fmt.Errorf("model checker failed: %w", err)}, nil
	}

	alltext := stdout.String()
	if mod != nil && mod.Cfg.MCVerbose {
		fmt.Println(alltext)
	}

	if mc.Scrape(alltext) {
		return &CheckResult{Proved: true}, nil
	}

	// Parse witness
	trace, err := ParseWitnessFile(outName)
	if err != nil {
		return &CheckResult{Error: fmt.Errorf("failed to parse witness: %w", err)}, nil
	}

	return &CheckResult{Trace: trace}, nil
}

// CheckIsolate is the main entry point for model checking an isolate.
// It calls ToAiger to convert the module to AIGER, writes it to a temp file,
// runs the ABC model checker, and parses the result.
//
// Python: ivy_mc.py:1684-1767
func CheckIsolate(mod *module.Module, method string) (*CheckResult, error) {
	if method == "" {
		method = "mc"
	}

	if mod.Cfg.MCVerbose {
		fmt.Println()
		fmt.Println(strings.Repeat("*", 80))
		fmt.Println()
	}

	// Convert to AIGER
	result, err := ToAiger(mod, method)
	if err != nil {
		return nil, fmt.Errorf("to_aiger failed: %w", err)
	}

	// Get AIGER string
	aigerStr := result.Aiger.String()

	// Run ABC model checker
	checkResult, err := RunABC(aigerStr, nil, mod)
	if err != nil {
		return nil, fmt.Errorf("model checker failed: %w", err)
	}

	if checkResult.Proved {
		return &CheckResult{Proved: true}, nil
	}

	if checkResult.Error != nil {
		return checkResult, nil
	}

	// If counterexample found, we would reconstruct the trace here
	// using aiger_witness_to_ivy_trace2. For now, return the raw trace.
	return checkResult, nil
}
