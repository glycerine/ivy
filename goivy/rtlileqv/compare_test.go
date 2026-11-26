package rtlileqv

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const goldGraph = `aag 6 3 1 1 2
2
4
6
8 10
12
10 4 6
12 8 6
`

const goldNames = `input 0 0 clock
input 1 0 reset
input 2 0 x
latch 0 0 q
latch 0 0 shared_q
output 0 0 y
`

const gateGraph = `aag 8 3 2 1 3
2
4
6
8 12
10 0
16
12 6 2
14 8 2
16 14 1
`

const gateNames = `input 0 0 x
input 1 0 clk
input 2 0 rst
latch 0 0 register_q
latch 0 0 shared_q
output 0 0 y
`

func compareText(t *testing.T, gateSource, mapSource string) (Result, error) {
	t.Helper()
	gold, err := readAiger(strings.NewReader(goldGraph), strings.NewReader(goldNames))
	if err != nil {
		t.Fatal(err)
	}
	gate, err := readAiger(strings.NewReader(gateSource), strings.NewReader(mapSource))
	if err != nil {
		t.Fatal(err)
	}
	return compare(gold, gate,
		designConfig{section: "gold", clock: "clock", reset: "reset"},
		designConfig{section: "gate", clock: "clk", reset: "rst"})
}

func TestCompareEquivalent(t *testing.T) {
	result, err := compareText(t, gateGraph, gateNames)
	if err != nil {
		t.Fatal(err)
	}
	if result.Outputs != 1 || result.Registers != 1 || len(result.Failures) != 0 {
		t.Fatalf("comparison = %+v, want one matched output and one matched register", result)
	}
}

func TestCompareReportsChangedNextState(t *testing.T) {
	graph := strings.Replace(gateGraph, "12 6 2", "12 7 2", 1)
	result, err := compareText(t, graph, gateNames)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Failures) != 1 || result.Failures[0].Cone != "reg shared_q[0]" {
		t.Fatalf("failures = %+v, want the aliased register's next-state cone", result.Failures)
	}
	inputs := result.Failures[0].Inputs
	if _, reset := inputs["rst[0]"]; !reset || !inputs["x[0]"] || len(inputs) != 2 {
		t.Fatalf("counterexample inputs = %v, want x=true and the reset bit", inputs)
	}
}

func TestCompareRejectsUnmatchedBits(t *testing.T) {
	for _, tc := range []struct{ name, graph, names, want string }{
		{"input", gateGraph, strings.Replace(gateNames, "input 0 0 x", "input 0 0 other", 1), "input x[0] is only in gold"},
		{"output", gateGraph, strings.Replace(gateNames, "output 0 0 y", "output 0 0 other", 1), "output y[0] is only in gold"},
		{"register", gateGraph, strings.Replace(gateNames, "latch 0 0 shared_q", "latch 0 0 other", 1), "register q[0] is only in gold"},
		{"ambiguous register", gateGraph, gateNames + "latch 1 0 shared_q\n", "aliases several gate registers"},
		{"unnamed live register", strings.Replace(gateGraph, "14 8 2", "14 10 2", 1), gateNames, "live register bit at position 1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := compareText(t, tc.graph, tc.names)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestCheckWithYosys(t *testing.T) {
	executable := os.Getenv("IVY_YOSYS")
	if executable == "" {
		executable = "yosys"
	}
	if _, err := exec.LookPath(executable); err != nil {
		t.Skip("Yosys unavailable")
	}
	const circuit = `module \tiny
  wire input 1 \clk
  wire input 2 \rst
  wire input 3 \x
  wire output 4 \y
  connect \y \x
end
`
	const config = `[gold]
file = design one.il
top = tiny
clk = clk
rst = rst

[gate]
file = design two.il
top = tiny
clk = clk
rst = rst
`
	dir := t.TempDir()
	for filename, content := range map[string]string{"design one.il": circuit, "design two.il": circuit, "compare.ileq": config} {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	result, err := Check(context.Background(), filepath.Join(dir, "compare.ileq"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Outputs != 1 || result.Registers != 0 || len(result.Failures) != 0 {
		t.Fatalf("comparison = %+v, want one equivalent output", result)
	}
	inverted := strings.Replace(circuit, `  connect \y \x`, `  cell $_NOT_ $invert
    connect \A \x
    connect \Y \y
  end`, 1)
	if err := os.WriteFile(filepath.Join(dir, "design two.il"), []byte(inverted), 0600); err != nil {
		t.Fatal(err)
	}
	result, err = Check(context.Background(), filepath.Join(dir, "compare.ileq"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Failures) != 1 || result.Failures[0].Cone != "output y[0]" {
		t.Fatalf("comparison = %+v, want the inverted output's counterexample", result)
	}
}
