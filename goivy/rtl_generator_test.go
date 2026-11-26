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

func TestGenerateRTLILHardwareExamples(t *testing.T) {
	for _, tc := range []struct {
		file, top, feature string
		outputs, registers int
	}{
		{"bfe_concat", "top", "module \\top.swap", 8, 8},
		{"pipe_cpu", "cpu", "cell $memwr_v2", 8, 4249},
		{"ooo_cpu_alu_dd_ref", "cpu", "cell $dff $dff$cpu.rob_a$", 0, 0},
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
