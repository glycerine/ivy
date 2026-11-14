package sim

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"math"
	"math/rand"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestOriginalCWordSquareAndStructuralOracles(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("source section collection uses ELF")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	binary := filepath.Join(t.TempDir(), "oracle")
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", filepath.Join(root, "src"), filepath.Join(root, "test_vectors/sim/oracle.c"), "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile source oracle: %v\n%s", err, out)
	}
	r := rand.New(rand.NewSource(843))
	var input bytes.Buffer
	var expected []string
	for trial := 0; trial < 200; trial++ {
		n := 1 + r.Intn(8)
		a, b := make([]uint32, n), make([]uint32, n)
		ca, cb := r.Intn(2), r.Intn(2)
		fmt.Fprintf(&input, "W %d %d %d", n, ca, cb)
		for i := range a {
			a[i] = r.Uint32()
			fmt.Fprintf(&input, " %x", a[i])
		}
		for i := range b {
			b[i] = r.Uint32()
			fmt.Fprintf(&input, " %x", b[i])
		}
		fmt.Fprintln(&input)
		g := aig.NewGraph()
		x, y := g.NewInput(), g.NewInput()
		if ca != 0 {
			x = x.Not()
		}
		if cb != 0 {
			y = y.Not()
		}
		z := g.AND(x, y)
		info := Sim_UtilInfoAlloc(len(g.Nodes()), n, true)
		copy(info[x.NodeID()], a)
		copy(info[y.NodeID()], b)
		Sim_UtilSimulateNodeOne(g, z, info, n, 0)
		var line strings.Builder
		for _, v := range info[z.NodeID()] {
			fmt.Fprintf(&line, "%x ", v)
		}
		boolInt := func(v bool) int {
			if v {
				return 1
			}
			return 0
		}
		fmt.Fprintf(&line, "%d %d %d %d %.9g", Sim_UtilCountOnes(a, n), boolInt(Sim_UtilInfoIsEqual(a, b, n)), boolInt(Sim_UtilInfoIsImp(a, b, n)), boolInt(Sim_UtilInfoIsClause(a, b, n)), Sim_ComputeSwitching(a, n))
		expected = append(expected, line.String())
		ni := 1 + r.Intn(70)
		words := SIM_NUM_WORDS(ni)
		pat, values := make([]uint32, words), make([]uint32, words)
		fmt.Fprintf(&input, "D %d", ni)
		for w := range pat {
			pat[w] = r.Uint32()
			fmt.Fprintf(&input, " %x", pat[w])
		}
		for w := range values {
			values[w] = r.Uint32()
			fmt.Fprintf(&input, " %x", values[w])
		}
		fmt.Fprintln(&input)
		p := &Sym_Man_t{Supports: make([][]int, 1)}
		for i := 0; i < ni; i++ {
			p.Supports[0] = append(p.Supports[0], i)
		}
		m := NewMatrix(ni)
		Sim_SymmsDeriveInfo(p, pat, values, m, 0)
		line.Reset()
		for i := 0; i < ni; i++ {
			for j := i + 1; j < ni; j++ {
				fmt.Fprint(&line, boolInt(m.Has(i, j)))
			}
		}
		expected = append(expected, line.String())
	}
	for trial := 0; trial < 100; trial++ {
		g := aig.NewGraph()
		var pool []aig.Edge
		for i := 0; i < 6; i++ {
			pool = append(pool, g.NewInput())
		}
		for i := 0; i < 25; i++ {
			a, b := pool[r.Intn(len(pool))], pool[r.Intn(len(pool))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			if r.Intn(2) == 0 {
				b = b.Not()
			}
			e := g.AND(a, b)
			if e.NodeID() != 0 {
				pool = append(pool, e)
			}
		}
		var gates []aig.Edge
		for _, e := range g.LiveEdges() {
			if g.Node(e).Kind == aig.And {
				gates = append(gates, e)
			}
		}
		fmt.Fprintf(&input, "S 6 %d", len(gates))
		for _, e := range gates {
			n := g.Node(e)
			fmt.Fprintf(&input, " %d %d", n.Fanins[0], n.Fanins[1])
			g.AddOutput(e)
		}
		fmt.Fprintln(&input)
		structural := Sim_ComputeStrSupp(g)
		support := Sim_UtilInfoAlloc(len(gates), 1, true)
		matrices := make([]Matrix, len(gates))
		for i := range gates {
			support[i][0] = 63
			matrices[i] = NewMatrix(6)
		}
		Sim_SymmsStructCompute(g, matrices, support)
		var line strings.Builder
		for o, e := range gates {
			fmt.Fprintf(&line, "%x:", structural[e.NodeID()][0])
			for a := 0; a < 6; a++ {
				for b := a + 1; b < 6; b++ {
					v := 0
					if matrices[o].Has(a, b) {
						v = 1
					}
					fmt.Fprint(&line, v)
				}
			}
			fmt.Fprint(&line, " ")
		}
		expected = append(expected, strings.TrimSpace(line.String()))
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("source oracle: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("got %d cases want%d", len(lines), len(expected))
	}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == expected[i] {
			continue
		}
		if i < 400 && i%2 == 0 {
			a, b := strings.Fields(line), strings.Fields(expected[i])
			if len(a) == len(b) {
				last := len(a) - 1
				var x, y float64
				fmt.Sscan(a[last], &x)
				fmt.Sscan(b[last], &y)
				if strings.Join(a[:last], " ") == strings.Join(b[:last], " ") && math.Abs(x-y) < 1e-7 {
					continue
				}
			}
		}
		t.Fatalf("case%d\nC  %s\nGo %s", i, line, expected[i])
	}
}

// Exercise the original complete simulation/FRAIG pipeline as well as the
// separately compiled source-stage fixtures above. The shipped ABC binary is
// optional; compiling the C stage oracle does not depend on it.
func TestOriginalABCSupportAndSymmetryCounts(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	binary := filepath.Join(root, "abc")
	if _, err := exec.LookPath(binary); err != nil {
		t.Skip("original ABC binary unavailable")
	}
	dir := t.TempDir()
	var command strings.Builder
	var wantsupp, wantsuppSource, wantsym []int
	for i, f := range []uint64{0x80, 0x96, 0xe8, 0xfe, 0x16, 0x69, 0x17, 0xe9, 0x81, 0x7e, 0x42, 0x53} {
		g := truthGraph(3, f)
		path := filepath.Join(dir, fmt.Sprintf("circuit%d.aig", i))
		if err := ioa.Ioa_WriteAiger(g, path, 0, 0); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&command, "read_aiger %s; print_supp; print_symm; ", path)
		supp := Sim_ComputeFunSupp(g, false)
		wantsupp = append(wantsupp, Sim_UtilCountOnes(supp[0], 1))
		sourceSupp := Sim_ComputeFunSuppSource(g, false)
		wantsuppSource = append(wantsuppSource, Sim_UtilCountOnes(sourceSupp[0], 1))
		wantsym = append(wantsym, Sim_ComputeTwoVarSymms(g, false))
	}
	out, err := exec.Command(binary, "-c", command.String()).CombinedOutput()
	if err != nil {
		t.Fatalf("original ABC: %v\n%s", err, out)
	}
	suppRE := regexp.MustCompile(`Total func supps\s*=\s*(\d+)`)
	symRE := regexp.MustCompile(`The total number of symmetries is (\d+)`)
	supp, sym := suppRE.FindAllSubmatch(out, -1), symRE.FindAllSubmatch(out, -1)
	if len(supp) != len(wantsupp) || len(sym) != len(wantsym) {
		t.Fatalf("missing ABC results: %s", out)
	}
	for i := range supp {
		a, _ := strconv.Atoi(string(supp[i][1]))
		b, _ := strconv.Atoi(string(sym[i][1]))
		if a != wantsupp[i] || a != wantsuppSource[i] || b != wantsym[i] {
			t.Fatalf("case%d C supp%d sym%d; Go supp%d/source supp%d sym%d", i, a, b, wantsupp[i], wantsuppSource[i], wantsym[i])
		}
	}
}
