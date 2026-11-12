package fsim

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOriginalWordAndTernaryStorageOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("original C fixture needs cc")
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	binary := filepath.Join(t.TempDir(), "oracle")
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", filepath.Join(root, "src"), "-I", filepath.Join(root, "test_vectors/fsim/include"), filepath.Join(root, "test_vectors/fsim/oracle.c"), "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build source oracle: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(981))
	var input bytes.Buffer
	var expected []string
	for trial := 0; trial < 200; trial++ {
		n := 1 + rng.Intn(8)
		a, b := rng.Intn(2), rng.Intn(2)
		left, right := make([]uint32, n), make([]uint32, n)
		fmt.Fprintf(&input, "W %d %d %d", n, a, b)
		for i := range left {
			left[i] = rng.Uint32()
			fmt.Fprintf(&input, " %x", left[i])
		}
		for i := range right {
			right[i] = rng.Uint32()
			fmt.Fprintf(&input, " %x", right[i])
		}
		fmt.Fprintln(&input)
		g := aig.NewGraph()
		x, y := g.NewInput(), g.NewInput()
		if a != 0 {
			x = x.Not()
		}
		if b != 0 {
			y = y.Not()
		}
		g.AddOutput(g.AND(x, y))
		p, err := Fsim_ManCreate(g)
		if err != nil {
			t.Fatal(err)
		}
		p.words = n
		out, err := p.Round([][]uint32{left, right})
		if err != nil {
			t.Fatal(err)
		}
		var line strings.Builder
		for _, w := range out[0] {
			fmt.Fprintf(&line, "%x ", w)
		}
		expected = append(expected, strings.TrimSpace(line.String()))
		bins := 1009
		fmt.Fprintf(&input, "H %d %d 0", n, bins)
		for _, w := range left {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
		expected = append(expected, fmt.Sprint(Fsim_ManTerStateHash(left, n, bins)))
		index, value := rng.Intn(128), 1+rng.Intn(3)
		fmt.Fprintf(&input, "P 0 %d %d\n", index, value)
		storage := make([]uint32, 8)
		Fsim_ManTerSimInfoSet(storage, index, value)
		expected = append(expected, fmt.Sprintf("%x %d", storage[index>>4], Fsim_ManTerSimInfoGet(storage, index)))
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run source oracle: %v\n%s", err, out)
	}
	scan := bufio.NewScanner(bytes.NewReader(out))
	line := 0
	for scan.Scan() {
		if line >= len(expected) || strings.TrimSpace(scan.Text()) != expected[line] {
			t.Fatalf("source mismatch line%d C=%s Go=%s", line, scan.Text(), expected[line])
		}
		line++
	}
	if scan.Err() != nil || line != len(expected) {
		t.Fatal("truncated source oracle", line, scan.Err())
	}
}
