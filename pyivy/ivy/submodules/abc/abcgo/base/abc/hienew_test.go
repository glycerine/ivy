package abc

import (
	"github.com/glycerine/abcgo/aig/ioa"
	"os"
	"reflect"
	"strings"
	"testing"
)

const testCBlif = `.model top
.inputs 3
.subckt leaf 2 1 2 5
.xor 8 6
.outputs 2 10 11
.end
.model leaf
.inputs 2
.and 2 4
.outputs 1 6
.end
`

func TestAUParseForwardModelAndPolarity(t *testing.T) {
	root, err := ParseCBlif(strings.NewReader(testCBlif), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer Au_ManDelete(root.Man)
	if root.ID != 1 || Au_ManNtkRoot(root.Man) != root || Au_NtkBoxNum(root) != 1 {
		t.Fatal("model reordering")
	}
	graph := Au_NtkDeriveFlatGia(root)
	for bits := 0; bits < 8; bits++ {
		x, y, z := bits&1 != 0, bits&2 != 0, bits&4 != 0
		want := (x && !y) != z
		got, err := graph.Evaluate([]bool{x, y, z})
		if err != nil || !reflect.DeepEqual(got, []bool{want, !want}) {
			t.Fatal(bits, got, want, err)
		}
	}
	leaf := Au_ManFindNtkP(root.Man, "leaf")
	and := leaf.Objects[4]
	if Au_ObjSuppSize(and) != 2 {
		t.Fatal("support traversal")
	}
	Au_ManCountThings(root.Man)
	if root.Nodes != 2 || root.Boxes != 1 || root.Ports != 8 {
		t.Fatal(root.Nodes, root.Boxes, root.Ports)
	}
}
func TestAUParserRangesAndDiagnostics(t *testing.T) {
	text := ".model top\n.inputs 3\n.outputs 5 2:3 3*2\n.end\n"
	n, err := ParseCBlif(strings.NewReader(text), "ranges")
	if err != nil {
		t.Fatal(err)
	}
	defer Au_ManDelete(n.Man)
	got, err := Au_NtkDeriveFlatGia(n).Evaluate([]bool{false, true, false})
	if err != nil || !reflect.DeepEqual(got, []bool{false, true, false, true, true}) {
		t.Fatal(got, err)
	}
	for _, bad := range []string{".and 2 4\n", ".model t\n.inputs -1\n", ".model t\n.inputs 1\n.outputs 2 2\n", ".model t\n.inputs 1\n.outputs 1 6\n", ".model t\n.subckt missing 0 0\n.end\n"} {
		if _, err := ParseCBlif(strings.NewReader(bad), "bad"); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
}
func TestAUTernarySimulationMatchesExhaustiveRefinements(t *testing.T) {
	n, err := ParseCBlif(strings.NewReader(testCBlif), "sim")
	if err != nil {
		t.Fatal(err)
	}
	defer Au_ManDelete(n.Man)
	graph := Au_NtkDeriveFlatGia(n)
	for a := AU_VAL0; a <= AU_VALX; a++ {
		for b := AU_VAL0; b <= AU_VALX; b++ {
			for c := AU_VAL0; c <= AU_VALX; c++ {
				values := []int{a, b, c}
				got := auSimulate(n, values, map[*Au_Ntk_t]bool{})
				possible := make([]int, 2)
				for bits := 0; bits < 8; bits++ {
					input := make([]bool, 3)
					valid := true
					for i, v := range values {
						input[i] = bits&(1<<i) != 0
						if v != AU_VALX && (v == AU_VAL1) != input[i] {
							valid = false
						}
					}
					if !valid {
						continue
					}
					result, _ := graph.Evaluate(input)
					for i, v := range result {
						if v {
							possible[i] |= AU_VAL1
						} else {
							possible[i] |= AU_VAL0
						}
					}
				}
				if !reflect.DeepEqual(got, possible) {
					t.Fatal(values, got, possible)
				}
			}
		}
	}
	// The final PO must use its own fanin, not the stale last object from C's loop.
	root := Au_NtkAlloc(Au_ManAlloc("po"), "po")
	zero := Au_NtkCreateConst0(root)
	Au_NtkCreatePo(root, zero*2)
	Au_NtkCreatePo(root, zero*2+1)
	if got := auSimulate(root, nil, map[*Au_Ntk_t]bool{}); !reflect.DeepEqual(got, []int{AU_VAL0, AU_VAL1}) {
		t.Fatal(got)
	}
}
func TestAUDerivePreservesHierarchyAndSourceData(t *testing.T) {
	root := testHierarchicalDesign(t, true)
	for _, o := range root.Design.Module("leaf").ObjectsOfKind(ObjectNode) {
		o.Data = "10 1\n"
	}
	for _, o := range root.Design.Module("middle").ObjectsOfKind(ObjectNode) {
		o.Data = "0 1\n"
	}
	root.Data = "caller"
	expected := Abc_NtkDeriveFlatGia(root)
	actual := Au_ManDeriveTest(root)
	if root.Data != "caller" {
		t.Fatal("clobbered source data")
	}
	for bits := 0; bits < 4; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		want, _ := expected.Evaluate(input)
		got, err := actual.Evaluate(input)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(got, want, err)
		}
	}
}
func TestAUMutualRecursionDetected(t *testing.T) {
	m := Au_ManAlloc("cycle")
	a, b := Au_NtkAlloc(m, "a"), Au_NtkAlloc(m, "b")
	Au_NtkCreateBox(a, nil, 0, b.ID)
	Au_NtkCreateBox(b, nil, 0, a.ID)
	if Au_NtkCheckRecursive(a) == 0 {
		t.Fatal("indirect hierarchy recursion missed")
	}
}

func TestAUOriginalCOracle(t *testing.T) {
	file, err := os.Open("test_vectors/hierarchy_au.cblif")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	root, err := ParseCBlif(file, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	defer Au_ManDelete(root.Man)
	data, err := os.ReadFile("test_vectors/hierarchy_au_oracle.aig")
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := ioa.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	graph := Au_NtkDeriveFlatGia(root)
	for bits := 0; bits < 8; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		values := make([]bool, oracle.MaxVar+1)
		read := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
		for i, lit := range oracle.Inputs {
			values[lit>>1] = input[i]
		}
		for _, gate := range oracle.ANDs {
			values[gate.Literal>>1] = read(gate.Left) && read(gate.Right)
		}
		want := make([]bool, len(oracle.Outputs))
		for i, lit := range oracle.Outputs {
			want[i] = read(lit)
		}
		got, err := graph.Evaluate(input)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(bits, got, want, err)
		}
	}
}
