package core
import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestPowerOriginalCOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("original-C oracle uses ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional original-C oracle requires cc")
	}
	source := os.Getenv("ABC_SOURCE_ROOT")
	if source == "" {
		_, file, _, _ := runtime.Caller(0)
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../src"))
	}
	binary := filepath.Join(t.TempDir(), "power-oracle")
	command := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-Wno-unused-result", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, "../test_vectors/power_oracle.c", filepath.Join(source, "aig/gia/giaUtil.c"), "-o", binary)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build original-C power oracle: %v\n%s", err, output)
	}
	random := rand.New(rand.NewSource(9107))
	var input bytes.Buffer
	var expected [][]uint32
	for trial := 0; trial < 80; trial++ {
		g := aig.NewGraph()
		edges := make([]aig.Edge, 6)
		for i := range edges {
			edges[i] = g.NewInput()
		}
		for i := 0; i < 18; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		for i := 0; i < 4; i++ {
			g.AddOutput(edges[len(edges)-1-i])
		}
		registers := trial % 3
		aig.Aig_ManSetRegNum(g, int32(registers))
		options := ActivityOptions{Words: 1 + trial%10, Frames: 20 + trial%29, Prefix: trial % 17, Transitions: trial%2 == 0, RandomFactor: trial%4 - 1}
		probabilities, err := SimulateActivity(context.Background(), g, options)
		if err != nil {
			t.Fatal(err)
		}
		nodes := g.Nodes()
		fmt.Fprintf(&input, "S %d %d %d %d %d %d %d %d %d\n", len(nodes), len(g.Inputs()), len(g.Outputs()), registers, options.Words, options.Frames, options.Prefix, boolInt(options.Transitions), options.RandomFactor)
		want := []uint32{uint32(len(nodes))}
		for id, n := range nodes {
			kind := 0
			switch n.Kind {
			case aig.Input:
				kind = 1
			case aig.And:
				kind = 2
			case aig.Output:
				kind = 3
			case aig.Buffer:
				kind = 4
			}
			a, b := uint32(0), uint32(0)
			if kind >= 2 {
				a = uint32(n.Fanins[0])
				if kind == 2 {
					b = uint32(n.Fanins[1])
				}
			}
			fmt.Fprintf(&input, "%d %d %d\n", kind, a, b)
			want = append(want, math.Float32bits(probabilities[id]))
		}
		expected = append(expected, want)
	}
	// Drive the original Dar_LibEval_rec directly with arbitrary DAG sharing,
	// complements, reused-node flags and non-dyadic float32 probabilities.
	for trial := 0; trial < 200; trial++ {
		g := aig.NewGraph()
		aig.Aig_ManIncrementTravId(g)
		count := 8 + trial%31
		class := preparedClass{}
		existing := make([]aig.Edge, count)
		initial := make([]float32, count)
		fmt.Fprintf(&input, "E %d %d\n", count, count-1)
		for i := 0; i < count; i++ {
			a, b := uint32(0), uint32(0)
			state := 0
			existing[i] = aig.InvalidEdge
			initial[i] = random.Float32()
			if i >= 4 {
				a = uint32(random.Intn(i)*2 + random.Intn(2))
				b = uint32(random.Intn(i)*2 + random.Intn(2))
				class.nodes = append(class.nodes, preparedNode{fanins: [2]uint32{a, b}})
				state = random.Intn(3)
			}
			if i < 4 || state != 0 {
				existing[i] = g.NewInput()
				if state == 2 {
					aig.Aig_ObjSetTravIdCurrent(g, existing[i])
				}
			}
			fmt.Fprintf(&input, "%d %d %d %d\n", a, b, state, math.Float32bits(initial[i]))
		}
		area, power := evaluateLibraryCandidate(g, &class, existing, count-1, initial, true)
		expected = append(expected, []uint32{uint32(area), math.Float32bits(power)})
	}
	// Source reference decrements fanouts in traversal order; a shared internal
	// node is visited on its last reference, not its first DFS occurrence.
	for trial := 0; trial < 100; trial++ {
		g := aig.NewGraph()
		a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
		shared := g.AND(a, b.Not())
		left := g.AND(shared, c)
		right := g.AND(shared, d.Not())
		root := g.AND(left, right)
		g.AddOutput(root)
		leaves := g.Inputs()
		if trial%2 == 0 {
			leaves = append(leaves, shared)
		}
		size := aig.Aig_NodeMffcLabelCut(g, root, leaves)
		probabilities := make([]float32, len(g.Nodes()))
		for i := range probabilities {
			probabilities[i] = random.Float32()
		}
		fmt.Fprintf(&input, "M %d %d\n", len(probabilities), root.NodeID())
		for i, n := range g.Nodes() {
			kind := 1
			if n.Kind == aig.And {
				kind = 2
			}
			if n.Kind == aig.Buffer || n.Kind == aig.Output {
				kind = 4
			}
			refs := n.Refs
			for _, leaf := range leaves {
				if leaf.NodeID() == i {
					refs++
				}
			}
			first, second := uint32(0), uint32(0)
			if kind != 1 {
				first = uint32(n.Fanins[0])
				if kind == 2 {
					second = uint32(n.Fanins[1])
				}
			}
			fmt.Fprintf(&input, "%d %d %d %d %d\n", kind, first, second, refs, math.Float32bits(probabilities[i]))
		}
		expected = append(expected, []uint32{uint32(size), math.Float32bits(mffcPower(g, root, probabilities))})
	}
	cmd := exec.Command(binary)
	cmd.Stdin = &input
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("original C power oracle: %v\n%s", err, output)
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for trial, want := range expected {
		if !scanner.Scan() {
			t.Fatalf("missing C result %d: %v", trial, scanner.Err())
		}
		reader := bytes.NewReader(scanner.Bytes())
		for column, value := range want {
			var got uint32
			if _, err := fmt.Fscan(reader, &got); err != nil || got != value {
				t.Fatalf("C mismatch trial %d column %d: got %08x want %08x (%v)", trial, column, got, value, err)
			}
		}
	}
	if scanner.Scan() {
		t.Fatal("unexpected additional C result")
	}
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
