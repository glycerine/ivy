package if_

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestClusteringNativeOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("original C oracle needs cc")
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	src := filepath.Join(root, "src")
	binary := filepath.Join(t.TempDir(), "oracle")
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, filepath.Join(root, "test_vectors/map_if/clu_oracle.c"), filepath.Join(src, "misc/mem/mem.c"), "-o", binary)
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build clustering oracle: %v\n%s", err, data)
	}
	rng := rand.New(rand.NewSource(219))
	var input bytes.Buffer
	var expected []string
	for trial := 0; trial < 150; trial++ {
		n := int32(5 + trial%5)
		leaf := int32(3 + trial%3)
		rootSize := n - leaf + 1
		if rootSize > 6 || rootSize < 3 || n <= leaf {
			continue
		}
		table := make([]uint64, If_CluWordNum(n))
		if trial%2 == 0 {
			bound, free := rng.Uint64(), rng.Uint64()
			for a := 0; a < 1<<uint(n); a++ {
				b := bound >> uint(a&((1<<uint(leaf))-1)) & 1
				index := a>>uint(leaf) | int(b)<<uint(n-leaf)
				if free>>uint(index)&1 != 0 {
					ifTruthSet(table, a)
				}
			}
			if n < 6 {
				table[0] = If_CluAdjust(table[0], n)
			}
		} else {
			for i := range table {
				table[i] = rng.Uint64()
			}
			if n < 6 {
				table[0] = If_CluAdjust(table[0], n)
			}
		}
		if If_CluSupport(table, n) != (1<<uint(n))-1 {
			continue
		}
		start, stop := int32(0), int32(0)
		if trial%7 == 0 {
			start = 1
		} else if trial%7 == 1 {
			stop = 1
		}
		fmt.Fprintf(&input, "D %d %d %d %d %d", n, leaf, rootSize, start, stop)
		for _, w := range table {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
		var r If_Grp_t
		var f0, f1 uint64
		g := If_CluCheck(nil, table, n, start, stop, leaf, rootSize, &r, &f0, &f1, nil, 0)
		var want strings.Builder
		fmt.Fprintf(&want, "%d %d", g.nVars, g.nMyu)
		for i := int32(0); i < leaf; i++ {
			fmt.Fprintf(&want, " %d", g.pVars[i])
		}
		fmt.Fprintf(&want, " %d", r.nVars)
		for i := int32(0); i < r.nVars; i++ {
			fmt.Fprintf(&want, " %d", r.pVars[i])
		}
		fmt.Fprintf(&want, " %x %x", f0, f1)
		expected = append(expected, want.String())
		fmt.Fprintf(&input, "S %d 0 0 0 0", n)
		for _, w := range table {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
		copyTable := append([]uint64(nil), table...)
		perm := make([]int32, n)
		phase := If_CluSemiCanonicize(copyTable, n, perm)
		want.Reset()
		fmt.Fprint(&want, phase)
		for _, v := range perm {
			fmt.Fprintf(&want, " %d", v)
		}
		for _, w := range copyTable {
			fmt.Fprintf(&want, " %x", w)
		}
		expected = append(expected, want.String())
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run clustering source: %v\n%s", err, output)
	}
	scan := bufio.NewScanner(bytes.NewReader(output))
	scan.Buffer(make([]byte, 4096), 1<<20)
	line := 0
	for scan.Scan() {
		if line >= len(expected) {
			t.Fatal("extra source output")
		}
		if scan.Text() != expected[line] {
			actual, want := strings.Fields(scan.Text()), strings.Fields(expected[line])
			column := -1
			for i := range actual {
				if i >= len(want) || actual[i] != want[i] {
					column = i
					break
				}
			}
			t.Fatalf("source mismatch case%d col%d\nC %s\nGo %s", line, column, scan.Text(), expected[line])
		}
		line++
	}
	if scan.Err() != nil || line != len(expected) {
		t.Fatal("truncated clustering oracle", line, len(expected), scan.Err())
	}
}
