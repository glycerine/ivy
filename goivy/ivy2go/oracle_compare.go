package ivy2go

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// oracle_compare.go is the behavioural oracle harness described in
// ARCHITECTURE_TODO.md §3.8 Tier 4. Given two already-built binaries
// — one emitted by ivy2cpp, one emitted by ivy2go, both for the same
// Ivy fixture — it runs both with the same stdin transcript and
// reports stdout divergence.
//
// The harness is intentionally minimal: it does NOT drive the
// build pipelines for each side. Tier 4 fixture-orchestration code
// (build C++ → BuildOutput in ivy2cpp; build Go → BuildOutput here;
// pair binaries; iterate over .in transcripts) layers on top of
// CompareGoVsCpp and lives in a separate suite under cmd/ when it
// lands.

// CompareGoVsCpp runs the binaries at goBin and cppBin, feeds each
// `stdin` on standard input, and diffs their stdouts.
//
// Returns nil on a clean match; an OracleDiff error otherwise. The
// error includes the first diverging line so callers can surface a
// minimal repro.
func CompareGoVsCpp(goBin, cppBin, stdin string) error {
	goOut, err := runForOracle(goBin, stdin)
	if err != nil {
		return fmt.Errorf("ivy2go oracle: go binary %q: %w", goBin, err)
	}
	cppOut, err := runForOracle(cppBin, stdin)
	if err != nil {
		return fmt.Errorf("ivy2go oracle: cpp binary %q: %w", cppBin, err)
	}
	if diff := diffNormalised(goOut, cppOut); diff != nil {
		return diff
	}
	return nil
}

// runForOracle executes `bin` with stdin and returns its stdout (or
// an error if the binary exits non-zero).
func runForOracle(bin, stdin string) (string, error) {
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w (stderr: %s)", err, strings.TrimSpace(errBuf.String()))
	}
	return out.String(), nil
}

// OracleDiff describes the first stdout-line divergence between two
// binaries.
type OracleDiff struct {
	Line int
	Go   string
	Cpp  string
}

func (d *OracleDiff) Error() string {
	return fmt.Sprintf("ivy2go oracle: divergence at line %d\n  go : %q\n  cpp: %q",
		d.Line, d.Go, d.Cpp)
}

// diffNormalised compares two stdouts modulo trailing whitespace,
// returning an *OracleDiff on the first differing line or nil on a
// clean match.
func diffNormalised(goOut, cppOut string) error {
	goLines := splitNormalisedLines(goOut)
	cppLines := splitNormalisedLines(cppOut)
	n := len(goLines)
	if len(cppLines) < n {
		n = len(cppLines)
	}
	for i := 0; i < n; i++ {
		if goLines[i] != cppLines[i] {
			return &OracleDiff{Line: i + 1, Go: goLines[i], Cpp: cppLines[i]}
		}
	}
	if len(goLines) != len(cppLines) {
		return &OracleDiff{
			Line: n + 1,
			Go:   sideOrEmpty(goLines, n),
			Cpp:  sideOrEmpty(cppLines, n),
		}
	}
	return nil
}

func splitNormalisedLines(s string) []string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	// Trim trailing whitespace per line; drop a trailing empty.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t")
	}
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func sideOrEmpty(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return "<EOF>"
}
