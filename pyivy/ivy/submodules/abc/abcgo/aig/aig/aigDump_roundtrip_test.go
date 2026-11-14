package aig_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
	_ "github.com/glycerine/abcgo/base/ver" // Register the production Verilog reader.
)

func TestBLIFPortNamesDoNotCollideWithGeneratedNets(t *testing.T) {
	cases := []struct {
		name    string
		inputs  []string
		outputs []string
	}{
		{"constant", []string{"n0", "b"}, nil},
		{"gate", []string{"n3", "b"}, nil},
		{"output", []string{"n4", "b"}, nil},
		{"default_input", nil, []string{"n1"}},
		{"output_gate", nil, []string{"n3"}},
		{"reserved_suffix", []string{"n0", "n0_1"}, []string{"n3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(a, b.Not()))
			var text bytes.Buffer
			if err := aig.WriteBLIF(&text, g, tc.inputs, tc.outputs); err != nil {
				t.Fatal(err)
			}
			n, err := abcio.ParseBLIF("roundtrip.blif", text.String(), false, true)
			if err != nil {
				t.Fatalf("cannot read exported BLIF: %v\n%s", err, text.String())
			}
			pis := n.ObjectsOfKind(abc.ObjectPI)
			pos := n.ObjectsOfKind(abc.ObjectPO)
			if len(pis) != 2 || len(pos) != 1 {
				t.Fatalf("interface dimensions: %d PI, %d PO", len(pis), len(pos))
			}
			for mask := 0; mask < 4; mask++ {
				values := map[*abc.Object]bool{pis[0]: mask&1 != 0, pis[1]: mask&2 != 0}
				got, err := n.Evaluate(pos[0].Signal(), values)
				if err != nil || got != (mask == 1) {
					t.Fatalf("mask %d: got %t (%v), want %t", mask, got, err, mask == 1)
				}
			}
		})
	}
}

func TestSequentialBLIFSuppliedNames(t *testing.T) {
	g := aig.NewGraph()
	a, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, state.Not()))
	g.AddOutput(g.NewXor(a, state))
	aig.Aig_ManSetRegNum(g, 1)
	var text bytes.Buffer
	if err := aig.WriteBLIF(&text, g, []string{"n0", "n3"}, []string{"n1", "n4"}); err != nil {
		t.Fatal(err)
	}
	n, err := abcio.ParseBLIF("state.blif", text.String(), false, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text.String(), ".latch n4 n3 0\n") {
		t.Fatalf("supplied latch terminal names were lost:\n%s", text.String())
	}
	checkDumpNetwork(t, g, n, false)
}

func dumpTruthGraph(table uint8, regs int32) *aig.Graph {
	g := aig.NewGraph()
	g.Name = "dump"
	inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
	root := g.Const0()
	for mask := 0; mask < 8; mask++ {
		if table&(1<<mask) == 0 {
			continue
		}
		term := g.Const1()
		for bit, input := range inputs {
			if mask&(1<<bit) == 0 {
				input = input.Not()
			}
			term = g.AND(term, input)
		}
		root = g.OR(root, term)
	}
	// Expose every gate to satisfy the source DFS completeness precondition.
	// Complemented, constant and direct-input outputs exercise pin polarity.
	g.AddOutput(root)
	g.AddOutput(root.Not())
	g.AddOutput(g.Const1())
	g.AddOutput(g.Const0())
	for id, node := range g.Nodes() {
		if node.Kind == aig.Input || node.Kind == aig.And {
			g.AddOutput(aig.Edge(id) << 1)
		}
	}
	aig.Aig_ManSetRegNum(g, regs)
	return g
}

func nativeDumpFiles(t *testing.T, graphs []*aig.Graph) string {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("native fixture uses GNU linker section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	output := t.TempDir()
	binary := filepath.Join(output, "dump-oracle")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src,
		filepath.Join(dir, "test_vectors/dump_compare.c"), filepath.Join(src, "aig/aig/aigUtil.c"),
		filepath.Join(src, "aig/aig/aigDfs.c"), "-Wl,--gc-sections", "-o", binary}
	if text, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile native dump fixture: %v: %s", err, text)
	}
	var input strings.Builder
	for _, g := range graphs {
		nodes := g.Nodes()
		nlogic := 0
		for _, n := range nodes {
			if n.Kind == aig.And {
				nlogic++
			}
		}
		fmt.Fprintf(&input, "%d %d %d %d\n", g.InputCount(), nlogic, g.OutputCount(), g.RegisterCount())
		for _, n := range nodes {
			if n.Kind == aig.And {
				fmt.Fprintf(&input, "%d %d\n", n.Fanins[0], n.Fanins[1])
			}
		}
		for _, out := range g.Outputs() {
			fmt.Fprintf(&input, "%d ", g.Node(out).Fanins[0])
		}
		input.WriteByte('\n')
	}
	cmd := exec.CommandContext(ctx, binary, output)
	cmd.Stdin = strings.NewReader(input.String())
	text, err := cmd.CombinedOutput()
	if err != nil || string(text) != fmt.Sprintf("dumped %d\n", len(graphs)) {
		t.Fatalf("native dump fixture: %v: %s", err, text)
	}
	return output
}

func checkDumpNetwork(t *testing.T, g *aig.Graph, n *abc.Network, clock bool) {
	t.Helper()
	if n == nil {
		t.Fatal("reader rejected exported circuit")
	}
	inputs, outputs := n.Inputs(), n.Outputs()
	if clock {
		if len(inputs) == 0 || abc.Abc_ObjFanout0(inputs[0]).Name() != "clock" {
			t.Fatal("missing leading clock input")
		}
		inputs = inputs[1:]
	}
	if len(inputs) != g.InputCount() || len(outputs) != g.OutputCount() || abc.Abc_NtkLatchNum(n) != g.RegisterCount() {
		t.Fatalf("exported dimensions: %d CI, %d CO, %d regs; want %d, %d, %d", len(inputs), len(outputs), abc.Abc_NtkLatchNum(n), g.InputCount(), g.OutputCount(), g.RegisterCount())
	}
	for _, latch := range n.ObjectsOfKind(abc.ObjectLatch) {
		if abc.Abc_LatchIsInit0(latch) != 1 {
			t.Fatal("exported register lost zero initialization")
		}
	}
	for mask := 0; mask < 1<<g.InputCount(); mask++ {
		assignment := make([]bool, len(inputs))
		values := map[*abc.Object]bool{}
		for _, ci := range n.Inputs() {
			values[ci] = false
		}
		for i, ci := range inputs {
			assignment[i] = mask&(1<<i) != 0
			values[ci] = assignment[i]
		}
		want, err := g.Evaluate(assignment)
		if err != nil {
			t.Fatal(err)
		}
		for i, out := range outputs {
			got, err := n.Evaluate(out.Signal(), values)
			if err != nil || got != want[i] {
				t.Fatalf("assignment %d output %d: %t (%v), want %t", mask, i, got, err, want[i])
			}
		}
	}
}

func readDumpNetwork(t *testing.T, path, format string, regs int32) *abc.Network {
	t.Helper()
	if format == "blif" {
		n, err := abcio.ReadBLIF(path, false, true)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	compact := strings.Join(strings.Fields(string(text)), "")
	if strings.Count(compact, "always@(posedgeclock)") != int(regs) {
		t.Fatal("registers do not all use the rising clock edge")
	}
	return abcio.Io_ReadVerilog(path, 1)
}

func TestDumpWritersAgainstNativeCircuits(t *testing.T) {
	var graphs []*aig.Graph
	for table := 0; table < 256; table++ {
		for regs := int32(0); regs <= 3; regs++ {
			graphs = append(graphs, dumpTruthGraph(uint8(table), regs))
		}
	}
	native := nativeDumpFiles(t, graphs)
	dir := t.TempDir()
	// The readers write progress bars directly to stdout. Keep them out of the
	// test report, retaining their diagnostics when a round trip fails.
	log, err := os.Create(filepath.Join(dir, "reader.log"))
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = log
	defer func() {
		os.Stdout = stdout
		log.Close()
		if t.Failed() {
			text, _ := os.ReadFile(log.Name())
			if len(text) > 2000 {
				text = text[len(text)-2000:]
			}
			t.Logf("last reader diagnostics: %s", text)
		}
	}()
	for i, g := range graphs {
		for _, source := range []string{"C", "Go"} {
			for _, format := range []string{"blif", "v"} {
				t.Run(fmt.Sprintf("%d/%s/%s", i, source, format), func(t *testing.T) {
					path := filepath.Join(native, fmt.Sprintf("%d.%s", i, format))
					if source == "Go" {
						path = filepath.Join(dir, "dump."+format)
						var err error
						if format == "blif" {
							err = aig.Aig_ManDumpBlif(g, path, nil, nil)
						} else {
							err = aig.Aig_ManDumpVerilog(g, path)
						}
						if err != nil {
							t.Fatal(err)
						}
					}
					n := readDumpNetwork(t, path, format, g.RegisterCount())
					checkDumpNetwork(t, g, n, format == "v" && g.RegisterCount() > 0)
				})
			}
		}
	}
}

func TestDumpWritersPreserveTypedXorAndBuffers(t *testing.T) {
	for regs := int32(0); regs <= 2; regs++ {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		root := g.NewXor(g.NewXor(a, b), g.NewXor(a, c))
		buffer := aig.Aig_ObjCreate(g, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{root.Not(), aig.InvalidEdge}})
		g.AddOutput(buffer.Not())
		g.AddOutput(buffer)
		g.AddOutput(g.NewXor(c.Not(), b))
		aig.Aig_ManSetRegNum(g, regs)
		for _, format := range []string{"blif", "v"} {
			t.Run(fmt.Sprintf("%d/%s", regs, format), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "dump."+format)
				var err error
				if format == "blif" {
					err = aig.Aig_ManDumpBlif(g, path, nil, nil)
				} else {
					err = aig.Aig_ManDumpVerilog(g, path)
				}
				if err != nil {
					t.Fatal(err)
				}
				n := readDumpNetwork(t, path, format, regs)
				checkDumpNetwork(t, g, n, format == "v" && regs > 0)
			})
		}
	}
}
