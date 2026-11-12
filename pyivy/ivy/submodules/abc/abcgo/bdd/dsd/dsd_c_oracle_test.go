package dsd

import (
	"bytes"
	"fmt"
	bdd "github.com/glycerine/abcgo/bdd/cudd/core"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func serializeTree(dd *DdManager, p *Dsd_Node_t, n int) string {
	r := Dsd_Regular(p)
	var table uint64
	for a := 0; a < 1<<uint(n); a++ {
		if evalBDD(dd, r.G, uint64(a)) {
			table |= 1 << uint(a)
		}
	}
	comp := 0
	if Dsd_IsComplement(p) {
		comp = 1
	}
	var out strings.Builder
	local := uint64(0)
	if r.Type == DSD_NODE_PRIME {
		f := Dsd_TreeGetPrimeFunction(dd, r)
		for a := 0; a < 1<<uint(n); a++ {
			if evalBDD(dd, f, uint64(a)) {
				local |= 1 << uint(a)
			}
		}
	}
	fmt.Fprintf(&out, "%d:%d:%016x:%016x[", comp, r.Type, table, local)
	if r.Type != DSD_NODE_BUF {
		for i, c := range r.pDecs {
			if i > 0 {
				out.WriteByte(',')
			}
			out.WriteString(serializeTree(dd, c, n))
		}
	}
	out.WriteByte(']')
	return out.String()
}
func TestOriginalCDecompositionOracle(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for original-source differential")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-O1", "-ffunction-sections", "-fdata-sections", "-DABC_USE_STDINT_H", "-I" + filepath.Join(root, "src"), filepath.Join(root, "abcgo/bdd/dsd/test_vectors/oracle.c")}
	for _, pattern := range []string{"src/bdd/dsd/*.c", "src/bdd/cudd/*.c", "src/bdd/mtr/*.c", "src/bdd/epd/*.c", "src/misc/st/*.c"} {
		files, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatal(err)
		}
		if len(files) == 0 {
			t.Skip("original C sources unavailable")
		}
		for _, f := range files {
			if filepath.Base(f) != "testcudd.c" && filepath.Base(f) != "dsdCheck.c" {
				args = append(args, f)
			}
		}
	}
	for _, f := range []string{"src/bdd/extrab/extraBddMisc.c", "src/misc/extra/extraUtilUtil.c", "src/misc/util/utilSort.c"} {
		args = append(args, filepath.Join(root, f))
	}
	// Original Dsd_CheckCacheClear mistakenly clears entry 0 repeatedly.
	// Correct only this undefined-memory initialization in the test oracle.
	cacheSource, err := os.ReadFile(filepath.Join(root, "src/bdd/dsd/dsdCheck.c"))
	if err != nil {
		t.Fatal(err)
	}
	cacheSource = bytes.ReplaceAll(cacheSource, []byte("pCache->pTable[0].bX[0] = NULL;"), []byte("pCache->pTable[i].bX[0] = NULL;"))
	cachePath := filepath.Join(t.TempDir(), "dsdCheck.c")
	if err := os.WriteFile(cachePath, cacheSource, 0600); err != nil {
		t.Fatal(err)
	}
	args = append(args, "-I"+filepath.Join(root, "src/bdd/dsd"), cachePath)
	binary := filepath.Join(t.TempDir(), "dsd-oracle")
	args = append(args, "-Wl,--gc-sections", "-lm", "-o", binary)
	if output, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile original C: %v\n%s", err, output)
	}
	var input bytes.Buffer
	expected := []string{}
	rng := rand.New(rand.NewSource(849))
	for reverse := 0; reverse < 2; reverse++ {
		for n := 0; n <= 6; n++ {
			dd := bdd.NewBDDManager(n)
			if reverse != 0 {
				order := make([]int, n)
				for i := range order {
					order[i] = n - 1 - i
				}
				bdd.Cudd_ShuffleHeap(dd, order)
			}
			m := Dsd_ManagerStart(dd, n, 0)
			count := 700
			if n <= 3 {
				count = 1 << uint(1<<uint(n))
			}
			for i := 0; i < count; i++ {
				mask := uint64(i)
				if n > 3 {
					mask = rng.Uint64()
					if i%2 == 0 {
						mask = structuredTable(rng, n)
					}
				}
				if n < 6 {
					mask &= (uint64(1) << uint(1<<uint(n))) - 1
				}
				fmt.Fprintf(&input, "%d %d %x\n", n, reverse, mask)
				Dsd_ManagerStop(m)
				m = Dsd_ManagerStart(dd, n, 0)
				f := tableBDD(dd, mask, n)
				p := Dsd_DecomposeOne(m, f)
				expected = append(expected, serializeTree(dd, p, n))
			}
			Dsd_ManagerStop(m)
		}
	}
	command := exec.Command(binary)
	command.Stdin = &input
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("original C: %v\n%s", err, output)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("C returned %d lines, expected %d\n%s", len(lines), len(expected), output)
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Fatalf("case %d:\nC  %s\nGo %s", i, line, expected[i])
		}
	}
}
