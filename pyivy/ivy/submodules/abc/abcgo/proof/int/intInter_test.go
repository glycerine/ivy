package int

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestVerifyInterpolantContainmentDiagnostics(t *testing.T) {
	graph := func(invert bool) *aig.Graph {
		g := aig.NewGraph()
		x := g.NewInput()
		if invert {
			x = x.Not()
		}
		g.AddOutput(x)
		return g
	}
	for _, tc := range []struct {
		name, want                   string
		lowerNeg, upperNeg, interNeg bool
	}{
		{"correct", "I is correct.", false, true, false},
		{"lower failure", "Property A => I fails.", true, true, false},
		{"upper failure", "Property I => !B fails.", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			old := os.Stdout
			os.Stdout = w
			defer func() { os.Stdout = old; r.Close() }()
			verifyInterpolant(graph(tc.lowerNeg), graph(tc.upperNeg), graph(tc.interNeg), "I")
			w.Close()
			out, err := io.ReadAll(r)
			if err != nil || !strings.Contains(string(out), tc.want) {
				t.Fatalf("output %q, error %v; want %q", out, err, tc.want)
			}
		})
	}
}

func TestInterpolantExpansionRetainsInputPositions(t *testing.T) {
	inter := aig.NewGraph()
	x := inter.NewInput()
	inter.AddOutput(x.Not())
	other := aig.NewGraph()
	other.NewInput()
	other.NewInput()
	other.AddOutput(other.Const0())
	expanded := Inter_ManDupExpand(inter, other)
	if expanded.InputCount() != 2 {
		t.Fatalf("inputs=%d", expanded.InputCount())
	}
	for mask := 0; mask < 4; mask++ {
		out, err := expanded.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || out[0] != (mask&1 == 0) {
			t.Fatalf("mask=%d output=%v error=%v", mask, out, err)
		}
	}
}

func TestInterpolantDumpIncludesEveryOutput(t *testing.T) {
	a := aig.NewGraph()
	x := a.NewInput()
	a.AddOutput(x)
	b := aig.NewGraph()
	y := b.NewInput()
	b.AddOutput(y.Not())
	path := filepath.Join(t.TempDir(), "interpolants.aig")
	p := &Inter_Man_t{pFileName: path, vInters: vec.Vec_PtrAlloc(2)}
	vec.Vec_PtrPush(p.vInters, a)
	vec.Vec_PtrPush(p.vInters, b)
	Inter_ManInterDump(p, 1)
	graph, err := ioa.Ioa_ReadAiger(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Inputs) != 1 || len(graph.Outputs) != 2 {
		t.Fatalf("dump dimensions inputs=%d outputs=%d", len(graph.Inputs), len(graph.Outputs))
	}
	if graph.Outputs[0] != graph.Inputs[0] || graph.Outputs[1] != graph.Inputs[0]^1 {
		t.Fatalf("dumped outputs %v, input %v", graph.Outputs, graph.Inputs)
	}
}
