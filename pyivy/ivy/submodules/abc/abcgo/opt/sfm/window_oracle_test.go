package sfm

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOriginalWindowLevelsAndMffcOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("source oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	binary := filepath.Join(t.TempDir(), "oracle")
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", filepath.Join(root, "src"), filepath.Join(root, "test_vectors/sfm/window_oracle.c"), "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile original C: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(132))
	var input bytes.Buffer
	var expected []string
	for trial := 0; trial < 300; trial++ {
		fs := make([][]int, 6)
		ts := make([]uint64, 6)
		for i := 0; i < 22; i++ {
			row := make([]int, 1+rng.Intn(4))
			for k := range row {
				row[k] = rng.Intn(len(fs))
			}
			fs = append(fs, row)
			ts = append(ts, rng.Uint64())
		}
		fs = append(fs, []int{26}, []int{27})
		ts = append(ts, 0, 0)
		p := testNetwork(t, fs, 6, 2, ts)
		for id := 6; id < 28; id++ {
			p.Fixed[id] = rng.Intn(5) == 0
			p.Empty[id] = rng.Intn(6) == 0
		}
		if err := p.refresh(); err != nil {
			t.Fatal(err)
		}
		p.Params.TfoLevMax = rng.Intn(4)
		p.Params.FanoutMax = rng.Intn(7)
		p.Params.WinSizeMax = 1 + rng.Intn(35)
		p.Params.GrowthLevel = rng.Intn(3)
		Sfm_NtkPrepare(p)
		target := 6 + rng.Intn(22)
		fmt.Fprintf(&input, "6 22 2 %d %d %d %d %d", target, p.Params.TfoLevMax, p.Params.FanoutMax, p.Params.WinSizeMax, p.Params.GrowthLevel)
		for id, row := range fs {
			fixed, empty := 0, 0
			if p.Fixed[id] {
				fixed = 1
			}
			if p.Empty[id] {
				empty = 1
			}
			fmt.Fprintf(&input, " %d %d %d", len(row), fixed, empty)
			for _, f := range row {
				fmt.Fprintf(&input, " %d", f)
			}
		}
		fmt.Fprintln(&input)
		mffc := Sfm_ObjMffcSize(p, target)
		ok := Sfm_NtkCreateWindow(p, target, false)
		var line strings.Builder
		value := 0
		if ok {
			value = 1
		}
		fmt.Fprintf(&line, "%d %d ", value, mffc)
		for _, row := range [][]int{p.Levels, p.Reverse, p.nodes, p.divs, p.roots, p.tfo, p.order} {
			for _, v := range row {
				fmt.Fprintf(&line, "%d,", v)
			}
			fmt.Fprint(&line, "|")
		}
		expected = append(expected, line.String())
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("source oracle: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("cases%d want%d", len(lines), len(expected))
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Fatalf("case%d\nC  %s\nGo %s", i, line, expected[i])
		}
	}
}
