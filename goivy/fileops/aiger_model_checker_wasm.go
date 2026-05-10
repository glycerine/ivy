//go:build wasm

package fileops

import "fmt"

// AigerModelCheckResult is the raw result from running an external
// file-oriented AIGER model checker. Ivy-level trace decoding stays in goivy.
type AigerModelCheckResult struct {
	Proved  bool
	Output  string
	Witness []byte
	Error   error
}

// RunAigerModelChecker is unavailable in browser wasm builds because it relies
// on temporary files and native subprocesses.
func RunAigerModelChecker(aigerStr string, cmd func(aigfilename, outfilename string) []string, scrape func(string) bool) (*AigerModelCheckResult, error) {
	return nil, fmt.Errorf("filesystem is not available in browser")
}
