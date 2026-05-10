//go:build !wasm

package fileops

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AigerModelCheckResult is the raw result from running an external
// file-oriented AIGER model checker. Ivy-level trace decoding stays in goivy.
type AigerModelCheckResult struct {
	Proved  bool
	Output  string
	Witness []byte
	Error   error
}

// RunAigerModelChecker writes an AIGER program to temporary files, optionally
// converts it through aigtoaig, runs the model-checker command, and reads back
// the witness file when the scrape function does not report a proof.
func RunAigerModelChecker(aigerStr string, cmd func(aigfilename, outfilename string) []string, scrape func(string) bool) (*AigerModelCheckResult, error) {
	if cmd == nil {
		return nil, fmt.Errorf("missing model checker command")
	}
	if scrape == nil {
		return nil, fmt.Errorf("missing model checker output scraper")
	}

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
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close AIGER temp file: %w", err)
	}

	aigName := strings.TrimSuffix(aagName, ".aag") + ".aig"
	defer os.Remove(aigName)
	if err := exec.Command("aigtoaig", aagName, aigName).Run(); err != nil {
		aigName = aagName
	}

	outName := strings.TrimSuffix(aagName, ".aag") + ".out"
	defer os.Remove(outName)

	args := cmd(aigName, outName)
	if len(args) == 0 {
		return nil, fmt.Errorf("empty model checker command")
	}

	p := exec.Command(args[0], args[1:]...)
	var stdout bytes.Buffer
	p.Stdout = &stdout
	if err := p.Run(); err != nil {
		return &AigerModelCheckResult{Output: stdout.String(), Error: err}, nil
	}
	if p.ProcessState != nil && !p.ProcessState.Success() {
		return nil, fmt.Errorf("model checker returned non-zero status")
	}

	alltext := stdout.String()
	if scrape(alltext) {
		return &AigerModelCheckResult{Proved: true, Output: alltext}, nil
	}

	witness, err := os.ReadFile(outName)
	if err != nil {
		return &AigerModelCheckResult{Output: alltext, Error: fmt.Errorf("failed to read witness: %w", err)}, nil
	}
	return &AigerModelCheckResult{Output: alltext, Witness: witness}, nil
}
