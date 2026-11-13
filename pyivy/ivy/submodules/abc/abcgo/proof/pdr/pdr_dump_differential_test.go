package pdr

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestInvariantDumpAgainstLocalABC(t *testing.T) {
	abc, err := filepath.Abs(filepath.Join("..", "..", "..", "abc"))
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(abc); err != nil || info.IsDir() || info.Mode()&0111 == 0 {
		t.Skip("local ABC executable is unavailable")
	}
	dir := t.TempDir()
	blif := filepath.Join(dir, "inv.blif")
	const source = ".model inv\n.inputs i\n.outputs o\n.latch n q 0\n.names n\n.names q o\n1 1\n.end\n"
	if err := os.WriteFile(blif, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	cfile := filepath.Join(dir, "c.pla")
	cmd := "read_blif " + blif + "; strash; pdr -d -I " + cfile
	output, err := exec.Command(abc, "-c", cmd).CombinedOutput()
	if err != nil || !strings.Contains(string(output), "Property proved") {
		t.Fatalf("ABC: %v\n%s", err, output)
	}
	g := aig.NewGraph()
	_ = g.NewInput() // primary input
	q := g.NewInput()
	g.Name = "inv"
	g.AddOutput(q)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fSilent, pars.fDumpInv = 1, 1
	pars.pInvFileName = filepath.Join(dir, "go.pla")
	if got := Pdr_ManSolve(g, &pars); got != 1 {
		t.Fatalf("Go result %d", got)
	}
	// Ignore date/name comments and optional input-name annotations, while
	// comparing the dimensions and actual invariant cover.
	cover := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var result []string
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, ".i ") || strings.HasPrefix(line, ".o ") || strings.HasPrefix(line, ".p ") || line == ".e" || strings.HasSuffix(line, " 1") {
				result = append(result, line)
			}
		}
		return strings.Join(result, "\n")
	}
	if c, goCover := cover(cfile), cover(pars.pInvFileName); c != goCover {
		t.Fatalf("C cover:\n%s\nGo cover:\n%s", c, goCover)
	}
}
