package fpga

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestMappingAgainstOriginalC(t *testing.T) {
	compiler, e := exec.LookPath("cc")
	if e != nil {
		t.Skip("C compiler unavailable")
	}
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(t.TempDir(), "fpga-reference")
	args := []string{"-O1", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I" + filepath.Join(root, "src"), "test_vectors/reference.c"}
	for _, name := range []string{"Core", "Create", "Cut", "CutUtils", "Vec", "Utils", "Time", "Match", "Switch", "Lib"} {
		args = append(args, filepath.Join(root, "src/map/fpga/fpga"+name+".c"))
	}
	args = append(args, filepath.Join(root, "src/misc/extra/extraUtilMemory.c"), "-Wl,--gc-sections", "-lm", "-o", binary)
	if log, e := exec.Command(compiler, args...).CombinedOutput(); e != nil {
		t.Fatalf("compile original C: %v\n%s", e, log)
	}
	for k := 2; k <= 6; k++ {
		for kind := 0; kind < 2; kind++ {
			p := manager(t, 6, k)
			root := p.Inputs[0]
			if kind == 0 {
				for _, n := range p.Inputs[1:] {
					root = Fpga_NodeAnd(p, root, n)
				}
			} else {
				a := Fpga_NodeExor(p, p.Inputs[0], p.Inputs[1])
				b := Fpga_NodeAnd(p, p.Inputs[2], p.Inputs[3])
				root = Fpga_NodeMux(p, p.Inputs[4], a, b)
				root = Fpga_NodeOr(p, root, p.Inputs[5])
			}
			p.Outputs = []*Fpga_Node_t{root, p.Inputs[0]}
			if e := p.Map(); e != nil {
				t.Fatal(e)
			}
			assertEquivalent(t, p)
			log, e := exec.Command(binary, strconv.Itoa(k), strconv.Itoa(kind)).CombinedOutput()
			if e != nil {
				t.Fatalf("C mapper: %v\n%s", e, log)
			}
			index := strings.Index(string(log), "RESULT ")
			if index < 0 {
				t.Fatalf("missing reference result %s", log)
			}
			var area, delay float32
			var luts int
			if _, e := fmt.Sscanf(string(log)[index:], "RESULT %g %g %d", &area, &delay, &luts); e != nil {
				t.Fatal(e)
			}
			if p.Stats.Area != area || p.Stats.Delay != delay || p.Stats.LUTs != luts {
				t.Fatalf("k%d kind%d: native area/delay/LUT %g/%g/%d original %g/%g/%d", k, kind, p.Stats.Area, p.Stats.Delay, p.Stats.LUTs, area, delay, luts)
			}
		}
	}
}
