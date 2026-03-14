package mc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// WitnessTrace holds a parsed counterexample trace from a model checker.
type WitnessTrace struct {
	Steps []WitnessStep
}

// WitnessStep is a single step in a counterexample trace.
type WitnessStep struct {
	Pre    string // pre-state latch values
	Input  string // input values
	Output string // output values
	Post   string // post-state latch values
}

// ParseWitnessFile reads an AIGER witness file (as produced by ABC).
// The format is:
//
//	Line 1: "1" (indicates counterexample found)
//	Remaining lines: "pre inp out post" (space-separated digit strings)
func ParseWitnessFile(filename string) (*WitnessTrace, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open witness file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return nil, fmt.Errorf("witness file is empty")
	}
	first := strings.TrimSpace(scanner.Text())
	if first != "1" {
		return nil, fmt.Errorf("expected '1' on first line of witness, got %q", first)
	}

	var trace WitnessTrace
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n\r")
		cols := strings.Split(line, " ")
		if len(cols) != 4 {
			return nil, fmt.Errorf("bad witness line: expected 4 columns, got %d", len(cols))
		}
		trace.Steps = append(trace.Steps, WitnessStep{
			Pre:    cols[0],
			Input:  cols[1],
			Output: cols[2],
			Post:   cols[3],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading witness file: %w", err)
	}

	return &trace, nil
}

// SimulateTrace replays a witness trace on an Aiger circuit, calling
// the handler for each step.
func SimulateTrace(a *Aiger, trace *WitnessTrace, handler func(step int, a *Aiger)) error {
	a.Reset()
	for i, step := range trace.Steps {
		if len(step.Input) != len(a.Inputs) {
			return fmt.Errorf("step %d: input length %d != %d", i, len(step.Input), len(a.Inputs))
		}
		a.Step(step.Input)
		handler(i, a)
		a.Advance()
	}
	return nil
}

// MatchHandlerBase provides default implementations for the MatchHandler interface.
type MatchHandlerBase struct{}

// Eval evaluates a condition in the trace context.
func (h *MatchHandlerBase) Eval(cond string) bool {
	if cond == "false" {
		return false
	}
	if cond == "true" {
		return true
	}
	// Default: assume true
	return true
}

// Handle processes an action in the trace.
func (h *MatchHandlerBase) Handle(action, env string) {
	// Default: no-op
}

// DoReturn handles a return action.
func (h *MatchHandlerBase) DoReturn(action, env string) {
	// Default: no-op
}
