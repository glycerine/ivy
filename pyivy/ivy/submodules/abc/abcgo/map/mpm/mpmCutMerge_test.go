package mpm

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func cutMergeSourceCases() string {
	var out bytes.Buffer
	for mode := 0; mode < 4; mode++ {
		p := &Mpm_Man_t{pPars: &Mpm_Par_t{}, nLutSize: 6}
		if mode == 3 {
			p.nLutSize = 2
		}
		if mode == 1 {
			p.pPars.FUseDsd = 1
		}
		if mode == 2 {
			p.pPars.FUseTruth = 1
		}
		unit := &Mpm_Uni_t{}
		vec.Vec_PtrPush(&p.vFreeUnits, unit)
		a := &Mpm_Cut_t{nLeaves: 2, pLeaves: []int32{4, 11}, iFunc: 37, fCompl: 1}
		b := &Mpm_Cut_t{nLeaves: 2, pLeaves: []int32{10, 6}}
		c := &Mpm_Cut_t{nLeaves: 1, pLeaves: []int32{8}}
		if mode == 0 {
			c = nil
		}
		cut := Mpm_ManMergeCuts(p, a, b, c)
		if cut == nil {
			fmt.Fprintln(&out, "null")
			continue
		}
		fmt.Fprintf(&out, "%d %d %d %d %d %d", cut.nLeaves, cut.iFunc, cut.fCompl, p.uPermMask[1], p.uComplMask[1], p.nCutsMerged)
		for _, lit := range cut.pLeaves[:cut.nLeaves] {
			fmt.Fprintf(&out, " %d", lit)
		}
		fmt.Fprintln(&out)
	}
	return out.String()
}
func TestCutMergeSourceCases(t *testing.T) {
	want := "3 33554431 0 0 0 1 4 11 6\n4 33554431 0 261703 2 1 4 11 6 8\n4 33554431 0 0 0 1 4 6 8 11\nnull\n"
	if got := cutMergeSourceCases(); got != want {
		t.Fatalf("got\n%s want\n%s", got, want)
	}
}
func TestCutMergeOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("oracle requires ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("oracle requires C compiler")
	}
	source := os.Getenv("ABC_SOURCE_ROOT")
	if source == "" {
		_, file, _, _ := runtime.Caller(0)
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	}
	bin := filepath.Join(t.TempDir(), "mpm-cut")
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, filepath.Join(source, "../abcgo/map/mpm/test_vectors/cut_merge_oracle.c"), "-o", bin)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile C oracle: %v\n%s", err, output)
	}
	output, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("C oracle: %v\n%s", err, output)
	}
	if got := cutMergeSourceCases(); got != string(output) {
		t.Fatalf("Go:\n%s C:\n%s", got, output)
	}
}
