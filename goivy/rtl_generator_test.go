package goivy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/rtlileqv"
)

const rtlExampleDir = "../pyivy/ivy/doc/examples/hardware"

func TestRTLILEquivalenceRejectsDifferentClock(t *testing.T) {
	yosys := os.Getenv("IVY_YOSYS")
	if yosys == "" {
		yosys = "yosys"
	}
	if _, err := exec.LookPath(yosys); err != nil {
		t.Skip("Yosys unavailable")
	}
	const design = `module \top
  wire input 1 \clk
  wire input 2 \rst
  wire input 3 \other
  wire input 4 \d
  wire output 5 \q
  cell $dff $reg
    parameter \WIDTH 1
    parameter \CLK_POLARITY %d
    connect \CLK \%s
    connect \D \d
    connect \Q \q
  end
end
`
	const config = `[gold]
file = gold.il
top = top
clk = clk
rst = rst

[gate]
file = gate.il
top = top
clk = clk
rst = rst
`
	for _, tt := range []struct {
		name, clock, diagnostic string
		goldEdge, gateEdge      int
	}{
		{"other input", "other", `[gate] flip-flop is outside clock "clk"`, 1, 1},
		{"opposite edge", "clk", "designs use different clock edges", 1, 0},
		{"both falling edge", "clk", "", 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, data := range map[string]string{
				"gold.il":    fmt.Sprintf(design, tt.goldEdge, "clk"),
				"gate.il":    fmt.Sprintf(design, tt.gateEdge, tt.clock),
				"check.ileq": config,
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			conf := filepath.Join(dir, "check.ileq")
			t.Run("GoIvy", func(t *testing.T) {
				result, err := rtlileqv.Check(context.Background(), conf)
				if tt.diagnostic == "" {
					if err != nil || result.Registers != 1 || len(result.Failures) != 0 {
						t.Fatalf("expected equivalent designs, got %+v, %v", result, err)
					}
				} else if err == nil || !strings.Contains(err.Error(), tt.diagnostic) {
					t.Fatalf("expected %q, got %+v, %v", tt.diagnostic, result, err)
				}
			})
			t.Run("PyIvy", func(t *testing.T) {
				const script = `
import sys
from ivy import ivy_rtlil_eqv as eqv
gold, gate = eqv.parse_config(sys.argv[1])
_, _, gold_edge = eqv.run_yosys(sys.argv[2], gold, sys.argv[3])
_, _, gate_edge = eqv.run_yosys(sys.argv[2], gate, sys.argv[3])
eqv.check_clock_edges(gold_edge, gate_edge)
`
				cmd := pythonIvyCommandForTest(t, "-c", script, conf, yosys, dir)
				out, err := cmd.CombinedOutput()
				if tt.diagnostic == "" && err != nil {
					t.Fatalf("expected successful normalization, got %v\n%s", err, out)
				}
				if tt.diagnostic != "" && (err == nil || !strings.Contains(string(out), tt.diagnostic)) {
					t.Fatalf("expected %q, got %v\n%s", tt.diagnostic, err, out)
				}
			})
		})
	}
}

func TestGenerateRTLILHardwareExamples(t *testing.T) {
	for _, tc := range []struct {
		file, top, feature string
		outputs, registers int
	}{
		{"bfe_concat", "top", "module \\top.swap", 8, 8},
		{"pipe_cpu", "cpu", "cell $memwr_v2", 8, 4249},
		{"ooo_cpu_alu_dd_ref", "cpu", "cell $dff $dff$cpu.rob_a$", 0, 0},
		{"germans/cachemem-abs", "memory_system", "module \\memory_system.cachemem.dir_body", 0, 0},
	} {
		t.Run(tc.file, func(t *testing.T) {
			source := filepath.Join(rtlExampleDir, tc.file)
			netlist, err := GenerateRTLIL(source + ".ivy")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(netlist, tc.feature) {
				t.Fatalf("netlist should contain %q", tc.feature)
			}
			if tc.registers == 0 {
				return
			}

			t.Run("equivalence", func(t *testing.T) {
				yosys := os.Getenv("IVY_YOSYS")
				if yosys == "" {
					yosys = "yosys"
				}
				if _, err := exec.LookPath(yosys); err != nil {
					t.Skip("Yosys unavailable")
				}
				reference, err := filepath.Abs(source + ".il")
				if err != nil {
					t.Fatal(err)
				}
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "gate.il"), []byte(netlist), 0o644); err != nil {
					t.Fatal(err)
				}
				config := filepath.Join(dir, "check.ileq")
				const sections = "[gold]\nfile = %s\ntop = %s\nclk = posedge\nrst = rst\n" +
					"[gate]\nfile = gate.il\ntop = %s\nclk = posedge\nrst = rst\n"
				if err := os.WriteFile(config, []byte(fmt.Sprintf(sections, reference, tc.top, tc.top)), 0o644); err != nil {
					t.Fatal(err)
				}
				result, err := rtlileqv.Check(context.Background(), config)
				if err != nil {
					t.Fatal(err)
				}
				if result.Outputs != tc.outputs || result.Registers != tc.registers || len(result.Failures) != 0 {
					t.Fatalf("RTLIL comparison = %+v", result)
				}
			})
		})
	}
}

func TestGenerateRTLILRejectsUnsupportedArraysAndFunctions(t *testing.T) {
	for _, tc := range []struct{ file, diagnostic string }{
		{"memtest2", "cannot translate initialization of array"},
		{"arrcopy", "not a point write"},
		{"to_rtl_bad_fun", "non-wire function"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			netlist, err := GenerateRTLIL(filepath.Join(rtlExampleDir, tc.file+".ivy"))
			if err == nil || !strings.Contains(err.Error(), tc.diagnostic) || netlist != "" {
				t.Fatalf("expected %q and no netlist, got %v", tc.diagnostic, err)
			}
		})
	}
}

func TestRTLILRejectsClockUpdatesThatCannotBeRepresented(t *testing.T) {
	const declarations = `#lang ivy1.8
relation value
export wire observed : bool
definition observed = value
after init { value := false }
`
	for _, tt := range []struct {
		name, clocks, goError, pyError string
	}{
		{"body", "export action posedge = { value := ~value }", "clock posedge must have an empty body", "clock posedge must have an empty body"},
		{"two clocks", `export action clock_a
export action clock_b
after clock_a { value := true }
after clock_b { value := false }
`, "updated by both", "written on more than one clock"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "clock.ivy")
			if err := os.WriteFile(file, []byte(declarations+tt.clocks+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Run("GoIvy", func(t *testing.T) {
				if _, err := GenerateRTLIL(file); err == nil || !strings.Contains(err.Error(), tt.goError) {
					t.Fatalf("expected a diagnostic containing %q, got %v", tt.goError, err)
				}
			})
			t.Run("PyIvy", func(t *testing.T) {
				cmd := pythonIvyCommandForTest(t, "-O", "-c", "from ivy import ivy_to_rtl; ivy_to_rtl.main()", filepath.Base(file))
				cmd.Dir = filepath.Dir(file)
				if out, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(out), tt.pyError) {
					t.Fatalf("expected a diagnostic containing %q, got %v\n%s", tt.pyError, err, out)
				}
			})
		})
	}
}

func TestGenerateRTLILPreservesWritePriority(t *testing.T) {
	netlist, err := GenerateRTLIL(filepath.Join(rtlExampleDir, "multiwrite.ivy"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(netlist, "cell $memwr_v2 ") != 2 {
		t.Fatal("expected two independent write ports")
	}
	yosys := os.Getenv("IVY_YOSYS")
	if yosys == "" {
		yosys = "yosys"
	}
	if _, err := exec.LookPath(yosys); err != nil {
		t.Skip("Yosys unavailable")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "multiwrite.il"), []byte(netlist), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"multiwrite_equiv.ys", "multiwrite_golden.sv"} {
		data, err := os.ReadFile(filepath.Join(rtlExampleDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(yosys, "-Q", "-s", "multiwrite_equiv.ys")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Yosys rejected memory write priority: %v\n%s", err, out)
	}
}

func TestGenerateRTLILPreservesWritePriorityAcrossBranches(t *testing.T) {
	yosys := os.Getenv("IVY_YOSYS")
	if yosys == "" {
		yosys = "yosys"
	}
	if _, err := exec.LookPath(yosys); err != nil {
		t.Skip("Yosys unavailable")
	}
	const source = `#lang ivy1.8
type addr
interpret addr -> bv[2]
type word
interpret word -> bv[2]
export action posedge
isolate top = {
    var mem(A:addr) : word
    wire choose : bool
    wire a : addr
    wire b : addr
    wire r : addr
    wire rd : word
    definition rd = mem(r)
    after init { mem(A) := 0 }
    after posedge {
        if choose {
            mem(a) := 1;
            mem(b) := 2
        } else {
            mem(b) := 2;
            mem(a) := 1
        }
    }
}
`
	const golden = `module top(input \posedge , rst, choose, input [1:0] a, b, r, output [1:0] rd);
    reg [1:0] mem [0:3];
    assign rd = mem[r];
    always @(posedge \posedge ) begin
        if (choose) begin
            mem[a] <= 2'd1;
            mem[b] <= 2'd2;
        end else begin
            mem[b] <= 2'd2;
            mem[a] <= 2'd1;
        end
    end
endmodule
`
	script, err := os.ReadFile(filepath.Join(rtlExampleDir, "multiwrite_equiv.ys"))
	if err != nil {
		t.Fatal(err)
	}
	for _, engine := range []string{"GoIvy", "PyIvy"} {
		t.Run(engine, func(t *testing.T) {
			dir := t.TempDir()
			for name, contents := range map[string]string{
				"multiwrite.ivy": source, "multiwrite_golden.sv": golden, "multiwrite_equiv.ys": string(script),
			} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if engine == "GoIvy" {
				netlist, err := GenerateRTLIL(filepath.Join(dir, "multiwrite.ivy"))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "multiwrite.il"), []byte(netlist), 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				translate := pythonIvyCommandForTest(t, "-O", "-c", "from ivy import ivy_to_rtl; ivy_to_rtl.main()", "multiwrite.ivy")
				translate.Dir = dir
				if out, err := translate.CombinedOutput(); err != nil {
					t.Fatalf("translate PyIvy RTL: %v\n%s", err, out)
				}
			}
			cmd := exec.Command(yosys, "-Q", "-s", "multiwrite_equiv.ys")
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("branch-specific write priority was not preserved: %v\n%s", err, out)
			}
		})
	}
}
