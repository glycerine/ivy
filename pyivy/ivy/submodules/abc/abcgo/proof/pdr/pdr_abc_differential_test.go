package pdr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// When the local ABC executable is available, compare independent ports of
// the same sequential machine. The exhaustive state oracle tests do not
// require the executable and remain authoritative in environments without it.
func TestPdrAgainstLocalABC(t *testing.T) {
	abc, err := filepath.Abs(filepath.Join("..", "..", "..", "abc"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(abc)
	if err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("local ABC executable is unavailable")
	}
	seed := uint32(0x9ac71423)
	next := func() uint8 { seed ^= seed << 13; seed ^= seed >> 17; seed ^= seed << 5; return uint8(seed) }
	for trial := 0; trial < 32; trial++ {
		tables := [3]uint8{next(), next(), next()}
		var blif strings.Builder
		blif.WriteString(".model pdr_compare\n.inputs i\n.outputs o\n.latch n0 q0 0\n.latch n1 q1 0\n")
		for i, output := range []string{"n0", "n1", "o"} {
			fmt.Fprintf(&blif, ".names i q0 q1 %s\n", output)
			for row := 0; row < 8; row++ {
				if tables[i]&(1<<row) != 0 {
					fmt.Fprintf(&blif, "%d%d%d 1\n", row&1, row>>1&1, row>>2&1)
				}
			}
		}
		blif.WriteString(".end\n")
		path := filepath.Join(t.TempDir(), "machine.blif")
		if err := os.WriteFile(path, []byte(blif.String()), 0600); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(abc, "-c", "read_blif "+path+"; strash; pdr").CombinedOutput()
		if err != nil {
			t.Fatalf("trial %d: ABC failed: %v\n%s", trial, err, output)
		}
		text := string(output)
		failed := strings.Contains(text, "was asserted in frame")
		if !failed && !strings.Contains(text, "Property proved") {
			t.Fatalf("trial %d: unexpected ABC result: %s", trial, text)
		}
		g := aig.NewGraph()
		vars := [3]aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		truth := func(table uint8) aig.Edge {
			result := g.Const0()
			for row := 0; row < 8; row++ {
				if table&(1<<row) == 0 {
					continue
				}
				term := g.Const1()
				for i, v := range vars {
					if row&(1<<i) == 0 {
						v = v.Not()
					}
					term = g.AND(term, v)
				}
				result = g.AND(result.Not(), term.Not()).Not()
			}
			return result
		}
		g.AddOutput(truth(tables[2]))
		g.AddOutput(truth(tables[0]))
		g.AddOutput(truth(tables[1]))
		aig.Aig_ManSetRegNum(g, 2)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.fSilent = 1
		pars.fNewXSim = int32(trial & 1)
		pars.fMonoCnf = int32((trial >> 1) & 1)
		if pars.fMonoCnf != 0 {
			pars.fNewXSim = 0
		}
		result := Pdr_ManSolve(g, &pars)
		if (result == 0) != failed || result < 0 {
			t.Fatalf("trial %d tables %v: ABC %s Go %d", trial, tables, text, result)
		}
		if failed && (g.SequentialCounterexample() == nil || saig.Saig_ManVerifyCex(g, g.SequentialCounterexample()) != 1) {
			t.Fatalf("trial %d: invalid Go witness", trial)
		}
	}
}
