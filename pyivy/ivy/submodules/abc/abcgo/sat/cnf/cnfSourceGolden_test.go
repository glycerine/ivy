package cnf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

type sourceGolden struct {
	Source      string    `json:"source"`
	Inputs      int       `json:"inputs"`
	Nodes       [][2]int  `json:"nodes"`
	Outputs     []int     `json:"outputs"`
	Registers   int       `json:"registers"`
	FreeOutputs int       `json:"free_outputs"`
	Variables   int       `json:"variables"`
	InputVars   []int32   `json:"input_vars"`
	OutputVars  []int32   `json:"output_vars"`
	Clauses     [][]int32 `json:"clauses"`
}

// The expected values were produced by the original C Cnf_Derive, not by
// the Go implementation. The fixture describes the graph as a topological
// list of signed references: 1..Inputs are CIs and later IDs are AND nodes.
func TestCnfDeriveAgainstSourceGoldens(t *testing.T) {
	for _, name := range []string{"reordered_free", "sequential_two_free"} {
		t.Run(name, func(t *testing.T) {
			bytes, err := os.ReadFile(filepath.Join("test_vectors", name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var want sourceGolden
			if err := json.Unmarshal(bytes, &want); err != nil {
				t.Fatal(err)
			}
			g := aig.NewGraph()
			edges := []aig.Edge{g.Const1()}
			for i := 0; i < want.Inputs; i++ {
				edges = append(edges, g.NewInput())
			}
			literal := func(index int) aig.Edge {
				if index < 0 {
					return edges[-index].Not()
				}
				return edges[index]
			}
			for _, node := range want.Nodes {
				edges = append(edges, g.AND(literal(node[0]), literal(node[1])))
			}
			for _, output := range want.Outputs {
				g.AddOutput(literal(output))
			}
			if want.Registers != 0 {
				aig.Aig_ManSetRegNum(g, int32(want.Registers))
			}
			got := Cnf_Derive(g, int32(want.FreeOutputs))
			if got == nil {
				t.Fatal("Cnf_Derive returned nil")
			}
			if got.Variables != want.Variables || !reflect.DeepEqual(got.Clauses, want.Clauses) {
				t.Fatalf("variables=%d clauses=%v; want variables=%d clauses=%v", got.Variables, got.Clauses, want.Variables, want.Clauses)
			}
			vars := func(edges []aig.Edge) []int32 {
				out := make([]int32, len(edges))
				for i, e := range edges {
					out[i] = got.VarNums[e.NodeID()]
				}
				return out
			}
			if gotInputs := vars(g.Inputs()); !reflect.DeepEqual(gotInputs, want.InputVars) {
				t.Errorf("input variables %v, want %v", gotInputs, want.InputVars)
			}
			if gotOutputs := vars(g.Outputs()); !reflect.DeepEqual(gotOutputs, want.OutputVars) {
				t.Errorf("output variables %v, want %v", gotOutputs, want.OutputVars)
			}
		})
	}
}
