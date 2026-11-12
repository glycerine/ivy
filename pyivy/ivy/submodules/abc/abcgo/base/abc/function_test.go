package abc

import (
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/map/mio"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestFunctionSOPBDDPhaseRoundtrips(t *testing.T) {
	for _, mode := range []int32{-1, 0, 1} {
		n, o, pis := fanFixture("1-01 0\n-10- 0\n--11 0\n")
		before := fanTruth(o, pis)
		if Abc_NtkSopToBdd(n) != 1 || n.Function != uint8(ABC_FUNC_BDD) {
			t.Fatal("SOP→BDD failed")
		}
		if Abc_NtkBddToSop(n, mode, 100) != 1 {
			t.Fatal("BDD→SOP failed")
		}
		if !reflect.DeepEqual(before, fanTruth(o, pis)) {
			t.Fatal(mode, o.Data)
		}
		if err := n.Check(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestFunctionSOPHopBDDRoundtripAndExport(t *testing.T) {
	n, o, pis := fanFixture("1-01 1\n-10- 1\n--11 1\n")
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, o)
	before := fanTruth(o, pis)
	if Abc_NtkSopToAig(n) != 1 {
		t.Fatal("SOP→Hop failed")
	}
	graph := Abc_NtkAigToGia(n, 0)
	if graph == nil {
		t.Fatal("local AIG export failed")
	}
	for assignment, want := range before {
		values := make([]bool, len(pis))
		for i := range values {
			values[i] = assignment&(1<<i) != 0
		}
		outputs, err := graph.Evaluate(values)
		if err != nil || len(outputs) != 1 || outputs[0] != want {
			t.Fatal(assignment, outputs, want, err)
		}
	}
	if Abc_NtkAigToBdd(n) != 1 || Abc_NtkBddToSop(n, -1, 100) != 1 {
		t.Fatal("Hop→BDD→SOP failed")
	}
	if !reflect.DeepEqual(before, fanTruth(o, pis)) {
		t.Fatal(o.Data)
	}
}
func TestFunctionBDDConversionLimitsAreAtomic(t *testing.T) {
	n, o, _ := fanFixture("1--- 1\n-1-- 1\n--1- 1\n---1 1\n")
	if Abc_NtkSopToBdd(n) != 1 {
		t.Fatal("conversion failed")
	}
	root := o.Data
	before := append([]Edge(nil), o.fanins...)
	if Abc_NtkBddToSop(n, 1, 1) != 0 {
		t.Fatal("cube limit ignored")
	}
	if n.Function != uint8(ABC_FUNC_BDD) || o.Data != root || !reflect.DeepEqual(before, o.fanins) {
		t.Fatal("failed conversion mutated network")
	}
	if Abc_NtkBddToSop(n, -1, 1) != 1 {
		t.Fatal("short negative cover should fit")
	}
}
func TestFunctionConstantConversionRemovesReciprocalFanins(t *testing.T) {
	n, o, pis := fanFixture("---- 1\n")
	if Abc_NtkSopToBdd(n) != 1 || Abc_NtkBddToSop(n, -1, 10) != 1 {
		t.Fatal("constant conversion failed")
	}
	if len(o.fanins) != 0 || o.Data != " 1\n" {
		t.Fatal(o.Data, o.fanins)
	}
	for _, pi := range pis {
		if len(pi.fanouts) != 0 {
			t.Fatal("stale fanout after constant reduction")
		}
	}
}
func TestFunctionDirectPhasePreservesTransparentBarrier(t *testing.T) {
	n, o, pis := fanFixture("11-- 0\n")
	before := fanTruth(o, pis)
	barrier := n.NewObject(ObjectNode)
	barrier.Barrier = true
	Abc_ObjAddFanin(barrier, o)
	Abc_NtkLogicMakeDirectSops(n)
	if !reflect.DeepEqual(before, fanTruth(o, pis)) {
		t.Fatal(o.Data)
	}
	if strings.Contains(o.Data.(string), " 0\n") {
		t.Fatal("negative phase retained")
	}
	if Abc_NtkSopToBdd(n) != 1 || Abc_NtkBddToSop(n, -1, 100) != 1 {
		t.Fatal("barrier conversion failed")
	}
	if barrier.Data != nil || len(barrier.fanins) != 1 {
		t.Fatal("barrier was converted")
	}
}
func TestFunctionAllPrimesAndDontCareBounds(t *testing.T) {
	dd := core.NewBDDManager(3)
	root := Abc_ConvertSopToBdd(dd, "01- 1\n1-1 1\n", nil)
	var scratch []byte
	text := Abc_ConvertBddToSop(nil, dd, root, root, 3, 1, &scratch, 1)
	rows := fanSOPRows(text, 3)
	if len(rows) != 3 {
		t.Fatal("missing consensus prime", text)
	}
	rebuilt := Abc_ConvertSopToBdd(dd, text, nil)
	if rebuilt != root {
		t.Fatal("prime cover function mismatch", text)
	}
	bounded := Abc_ConvertBddToSop(nil, dd, root, dd.One, 3, 0, &scratch, 1)
	boundFunc := Abc_ConvertSopToBdd(dd, bounded, nil)
	if !core.Cudd_bddLeq(dd, root, boundFunc) {
		t.Fatal("constant upper bound incorrectly produced false", bounded)
	}
}
func TestFunctionLargeFaninReorderingPreservesSemantics(t *testing.T) {
	n := NewNetwork("large")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	var pis []*Object
	for i := 0; i < 11; i++ {
		pis = append(pis, n.NewObject(ObjectPI))
	}
	o := n.NewObject(ObjectNode)
	o.Data = "1---------0 1\n-1-------1- 1\n"
	for _, pi := range pis {
		Abc_ObjAddFanin(o, pi)
	}
	before := fanTruth(o, pis)
	if Abc_NtkSopToBdd(n) != 1 || Abc_NtkBddToSop(n, -1, 100) != 1 {
		t.Fatal("large fanin conversion failed")
	}
	if !reflect.DeepEqual(before, fanTruth(o, pis)) {
		t.Fatal("reordered edge/function mismatch", o.Data)
	}
}
func TestFunctionMappedGateToSOP(t *testing.T) {
	library, err := mio.ReadLibrary("tiny", "GATE and2 1 Y=a*b;\nPIN * NONINV 1 999 1 0 1 0\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	n := NewNetwork("mapped")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_MAP)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	o := n.NewObject(ObjectNode)
	o.Data = mio.Mio_LibraryReadAnd2(library)
	if o.Data == nil {
		t.Fatal("and gate missing")
	}
	Abc_ObjAddFanin(o, a)
	Abc_ObjAddFanin(o, b)
	if Abc_NtkMapToSop(n) != 1 {
		t.Fatal("unmap failed")
	}
	truth := fanTruth(o, []*Object{a, b})
	if !reflect.DeepEqual(truth, []bool{false, false, false, true}) {
		t.Fatal(o.Data, truth)
	}
}
func TestFunctionOriginalABCOracle(t *testing.T) {
	data, err := os.ReadFile("test_vectors/function_off_oracle.blif")
	if err != nil {
		t.Fatal(err)
	}
	n := NewNetwork("oracle")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	node := n.NewObject(ObjectNode)
	for _, name := range []string{"a", "b", "c", "d"} {
		pi := n.NewObject(ObjectPI)
		pi.SetName(name)
		Abc_ObjAddFanin(node, pi)
	}
	node.Data = "1-01 0\n-10- 0\n--11 0\n"
	if Abc_NtkSopToBdd(n) != 1 || Abc_NtkBddToSop(n, -1, 100) != 1 {
		t.Fatal("oracle conversion failed")
	}
	var inputs []string
	var rows strings.Builder
	inside := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, ".names ") {
			fields := strings.Fields(line)
			inputs = fields[1 : len(fields)-1]
			inside = true
			continue
		}
		if strings.HasPrefix(line, ".") {
			inside = false
		}
		if inside && line != "" {
			rows.WriteString(line)
			rows.WriteByte('\n')
		}
	}
	var actual []string
	for _, edge := range node.fanins {
		actual = append(actual, edge.Object.Name())
	}
	if !reflect.DeepEqual(actual, inputs) || node.Data != rows.String() {
		t.Fatalf("Go %v %q; original ABC %v %q", actual, node.Data, inputs, rows.String())
	}
}
func TestFunctionIsoAfterManagerConversions(t *testing.T) {
	n, node, _ := fanFixture("1-01 1\n-10- 1\n--11 1\n")
	other := n.NewObject(ObjectNode)
	other.Data = node.Data
	for _, edge := range node.fanins {
		Abc_ObjAddFanin(other, edge)
	}
	for _, o := range []*Object{node, other} {
		po := n.NewObject(ObjectPO)
		Abc_ObjAddFanin(po, o)
	}
	for _, convert := range []func(*Network) int32{Abc_NtkSopToAig, Abc_NtkAigToBdd} {
		if convert(n) != 1 {
			t.Fatal("manager conversion failed")
		}
		if Abc_NtkFunctionalIso(n, 0, 1, 1) != 1 {
			t.Fatal("functional iso lost representation support", n.Function)
		}
	}
}
func TestFunctionLocalAIGCopyRetainsComplements(t *testing.T) {
	source, node, pis := fanFixture("1-01 1\n-10- 1\n--11 1\n")
	before := fanTruth(node, pis)
	if Abc_NtkSopToAig(source) != 1 {
		t.Fatal("conversion failed")
	}
	target := NewNetwork("target")
	var copied []*Object
	for _, pi := range pis {
		input := target.NewObject(ObjectPI)
		copied = append(copied, input)
		pi.SetCopyEdge(input.Signal())
	}
	root := Abc_ConvertAigToAig(target, node)
	for assignment, want := range before {
		values := map[*Object]bool{}
		for i, input := range copied {
			values[input] = assignment&(1<<i) != 0
		}
		got, err := target.Evaluate(root, values)
		if err != nil || got != want {
			t.Fatal(assignment, got, want, err)
		}
	}
}
func TestFunctionNetworkEvaluationAcrossRepresentations(t *testing.T) {
	n, node, pis := fanFixture("1-01 0\n-10- 0\n--11 0\n")
	want := fanTruth(node, pis)
	check := func() {
		t.Helper()
		for assignment, expected := range want {
			values := map[*Object]bool{}
			for i, pi := range pis {
				values[pi] = assignment&(1<<i) != 0
			}
			got, err := n.Evaluate(node.Signal(), values)
			if err != nil || got != expected {
				t.Fatal(n.Function, assignment, got, expected, err)
			}
		}
	}
	check()
	if Abc_NtkSopToAig(n) != 1 {
		t.Fatal("Hop conversion failed")
	}
	check()
	if Abc_NtkAigToBdd(n) != 1 {
		t.Fatal("BDD conversion failed")
	}
	check()
	if Abc_NtkBddToSop(n, -1, 100) != 1 {
		t.Fatal("SOP conversion failed")
	}
	check()
}
