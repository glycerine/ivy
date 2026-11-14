package abc

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func ntkFixture() *Network {
	n := Abc_NtkCreateFromSops("sample", []string{"11 1\n", "1- 1\n-1 1\n"})
	n.Spec = "sample.blif"
	return n
}
func ntkCheckSame(t *testing.T, a, b *Network) {
	t.Helper()
	if err := b.Check(); err != nil {
		t.Fatal(err)
	}
	if len(a.cis) != len(b.cis) || len(a.cos) != len(b.cos) {
		t.Fatal("interface mismatch")
	}
	for assignment := 0; assignment < 1<<len(a.cis); assignment++ {
		va, vb := make(map[*Object]bool), make(map[*Object]bool)
		for i, o := range a.cis {
			va[o] = assignment&(1<<i) != 0
			vb[b.cis[i]] = va[o]
		}
		for i, o := range a.cos {
			x, err := a.Evaluate(o.Signal(), va)
			if err != nil {
				t.Fatal(err)
			}
			y, err := b.Evaluate(b.cos[i].Signal(), vb)
			if err != nil || x != y {
				t.Fatalf("output%d assignment%d", i, assignment)
			}
		}
	}
}
func TestNetworkDupMetadataAndIsolation(t *testing.T) {
	n := ntkFixture()
	n.vFins = vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(n.vFins, 4)
	vec.Vec_IntPush(n.vFins, 7)
	n.Constraints = 1
	n.Onehots = [][]int32{{1, 2}}
	n.ObjectPermutation = []int32{0, 1}
	n.Phases = []int32{1, 0}
	n.LTLProperties = []string{"G p"}
	n.SequenceModel = util.Abc_CexAlloc(0, 2, 1)
	n.AndGateDelay = 2.5
	n.WireLoad = "wire"
	n.CILevel = func(o *Object) int { return 4 + o.ID }
	copy := Abc_NtkDup(n)
	ntkCheckSame(t, n, copy)
	if copy.Name != n.Name || copy.Spec != n.Spec || copy.Constraints != 1 || copy.AndGateDelay != 2.5 || copy.WireLoad != "wire" {
		t.Fatal("metadata lost")
	}
	copy.Onehots[0][0] = 99
	copy.Phases[0] = 0
	copy.ObjectPermutation[0] = 8
	copy.LTLProperties[0] = "changed"
	if n.Onehots[0][0] != 1 || n.Phases[0] != 1 || n.ObjectPermutation[0] != 0 || n.LTLProperties[0] != "G p" {
		t.Fatal("mutable metadata shared")
	}
	if copy.SequenceModel == n.SequenceModel {
		t.Fatal("counterexample shared")
	}
	if copy.vFins == n.vFins || !reflect.DeepEqual(vec.Vec_IntArray(copy.vFins), vec.Vec_IntArray(n.vFins)) {
		t.Fatal("fault metadata was not copied independently")
	}
	vec.Vec_IntWriteEntry(copy.vFins, 0, 99)
	if vec.Vec_IntEntry(n.vFins, 0) != 4 {
		t.Fatal("fault metadata is shared")
	}
	for i, o := range copy.cis {
		if copy.CILevel(o) != n.CILevel(n.cis[i]) {
			t.Fatal("arrival levels lost")
		}
	}
	n.CILevel = nil
	if copy.CILevel(copy.cis[0]) == 0 {
		t.Fatal("callback not snapshotted")
	}
	if copy.FunctionManager == n.FunctionManager {
		t.Fatal("SOP manager shared")
	}
}
func TestNetworkLocalManagerDup(t *testing.T) {
	for _, function := range []Abc_NtkFunc_t{ABC_FUNC_AIG, ABC_FUNC_BDD} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, function, 1)
		a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
		node := n.NewObject(ObjectNode)
		Abc_ObjAddFanin(node, a)
		Abc_ObjAddFanin(node, b)
		if function == ABC_FUNC_AIG {
			m := n.FunctionManager.(*hop.Hop_Man_t)
			node.Data = hop.Hop_And(m, hop.Hop_IthVar(m, 0), hop.Hop_IthVar(m, 1))
		} else {
			m := n.FunctionManager.(*core.DdManager)
			node.Data = core.Cudd_bddAnd(m, core.Cudd_bddIthVar(m, 0), core.Cudd_bddIthVar(m, 1))
			core.Cudd_Ref(node.Data.(*core.DdNode))
		}
		Abc_ObjAddFanin(n.NewObject(ObjectPO), node)
		copy := Abc_NtkDup(n)
		if copy.FunctionManager == n.FunctionManager {
			t.Fatal("local function manager shared")
		}
		ntkCheckSame(t, n, copy)
	}
}
func TestNetworkConesTargetsAndComposition(t *testing.T) {
	n := ntkFixture()
	cone := Abc_NtkCreateCone(n, n.kindOrder[ObjectNode][0], "and", 0)
	if len(cone.cis) != 2 || len(cone.cos) != 1 {
		t.Fatal("cone shape")
	}
	target := Abc_NtkCreateTarget(n, n.cos, []int32{0, 1})
	for assignment := 0; assignment < 4; assignment++ {
		inputs := map[*Object]bool{target.cis[0]: assignment&1 != 0, target.cis[1]: assignment&2 != 0}
		got, err := target.Evaluate(target.cos[0].Signal(), inputs)
		if err != nil || got != (assignment == 1 || assignment == 2) {
			t.Fatal("target truth")
		}
	}
	top := Abc_NtkCreateFromSops("top", []string{"11 1\n"})
	bottom := Abc_NtkCreateFromSops("bottom", []string{"1 1\n", "1 0\n"})
	composed := Abc_NtkAttachBottom(top, bottom)
	for _, value := range []bool{false, true} {
		got, err := composed.Evaluate(composed.cos[0].Signal(), map[*Object]bool{composed.cis[0]: value})
		if err != nil || got {
			t.Fatal("composition differs")
		}
	}
}
func TestNetworkSequentialTerminalConversion(t *testing.T) {
	n := ntkFixture()
	Abc_NtkMakeSeq(n, 1)
	if Abc_NtkLatchNum(n) != 1 || Abc_NtkPiNum(n) != 1 || Abc_NtkPoNum(n) != 1 {
		t.Fatal("sequential interface")
	}
	copy := Abc_NtkDup(n)
	ntkCheckSame(t, n, copy)
	Abc_NtkMakeComb(copy, 0)
	if len(copy.boxes) != 0 || len(copy.cis) != 2 || len(copy.cos) != 2 {
		t.Fatal("combinational interface")
	}
	if err := copy.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestNetworkPermutationRoundTrip(t *testing.T) {
	n := ntkFixture()
	pis, pos := n.ObjectsOfKind(ObjectPI), n.ObjectsOfKind(ObjectPO)
	Abc_NtkPermute(n, 1, 1, 0, "")
	Abc_NtkUnpermute(n)
	for i, o := range pis {
		if n.kindOrder[ObjectPI][i] != o {
			t.Fatal("PI order not restored")
		}
	}
	for i, o := range pos {
		if n.kindOrder[ObjectPO][i] != o {
			t.Fatal("PO order not restored")
		}
	}
}

func TestNetworkDupPreservesCompactWitness(t *testing.T) {
	n := ntkFixture()
	Abc_NtkMakeSeq(n, 1)
	n.SequenceModel = util.Abc_CexAlloc(0, 1, 1)
	n.SequenceModel.PData[0] = 1
	copy := Abc_NtkDup(n)
	if copy.SequenceModel.NRegs != 0 || copy.SequenceModel.NPis != 1 || copy.SequenceModel.NBits != 1 || copy.SequenceModel.PData[0] != 1 {
		t.Fatal("compact induction witness was reinterpreted as a full trace")
	}
	copy.SequenceModel.PData[0] = 0
	if n.SequenceModel.PData[0] != 1 {
		t.Fatal("compact witness storage shared")
	}
}
func TestNetworkSelectAndEditOutputs(t *testing.T) {
	n := ntkFixture()
	selected := Abc_NtkMakeOnePo(n, 1, 1)
	for x := 0; x < 4; x++ {
		got, err := selected.Evaluate(selected.cos[0].Signal(), map[*Object]bool{selected.cis[0]: x&1 != 0, selected.cis[1]: x&2 != 0})
		if err != nil || got != (x != 0) {
			t.Fatal("selected output differs")
		}
	}
	aig := Abc_NtkCreateTarget(n, []*Object{n.cos[0]}, []int32{1})
	second := aig.NewObject(ObjectPO)
	Abc_ObjAddFanin(second, aig.One())
	Abc_NtkSwapOneOutput(aig, 1)
	got, err := aig.Evaluate(aig.cos[0].Signal(), map[*Object]bool{aig.cis[0]: false, aig.cis[1]: false})
	if err != nil || !got {
		t.Fatal("swapped constant lost")
	}
	Abc_NtkDropOneOutput(aig, 0, 0, 0)
	got, err = aig.Evaluate(aig.cos[0].Signal(), map[*Object]bool{aig.cis[0]: false, aig.cis[1]: false})
	if err != nil || got {
		t.Fatal("dropped output not zero")
	}
}

func TestDeletingClonePreservesSourceManager(t *testing.T) {
	for _, function := range []Abc_NtkFunc_t{ABC_FUNC_AIG, ABC_FUNC_BDD} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, function, 1)
		a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
		node := n.NewObject(ObjectNode)
		Abc_ObjAddFanin(node, a)
		Abc_ObjAddFanin(node, b)
		if function == ABC_FUNC_AIG {
			m := n.FunctionManager.(*hop.Hop_Man_t)
			node.Data = hop.Hop_And(m, hop.Hop_IthVar(m, 0), hop.Hop_IthVar(m, 1))
		} else {
			m := n.FunctionManager.(*core.DdManager)
			node.Data = core.Cudd_bddAnd(m, core.Cudd_bddIthVar(m, 0), core.Cudd_bddIthVar(m, 1))
			core.Cudd_Ref(node.Data.(*core.DdNode))
		}
		Abc_ObjAddFanin(n.NewObject(ObjectPO), node)
		clone := cloneNetwork(n)
		if n.Copy != clone || node.CopyEdge.Object == nil {
			t.Fatal("hierarchy copy binding lost")
		}
		Abc_NtkDelete(clone)
		for x := 0; x < 4; x++ {
			got, err := n.Evaluate(n.cos[0].Signal(), map[*Object]bool{a: x&1 != 0, b: x&2 != 0})
			if err != nil || got != (x == 3) {
				t.Fatal("deleting clone invalidated source manager")
			}
		}
		Abc_NtkDelete(n)
	}
}
func TestFanoutSplittingAtTwo(t *testing.T) {
	n := ntkFixture()
	driver := n.kindOrder[ObjectNode][0]
	for i := 0; i < 5; i++ {
		Abc_ObjAddFanin(n.NewObject(ObjectPO), driver)
	}
	copy := Abc_NtkNodeDup(n, 2, 0)
	ntkCheckSame(t, n, copy)
	for _, o := range copy.kindOrder[ObjectNode] {
		if len(o.fanouts) >= 2 {
			t.Fatal("fanout bound not realized")
		}
	}
}

func TestFlopFilePermutationUsesInverse(t *testing.T) {
	n := Abc_NtkCreateFromSops("perm", []string{"1111 1\n", "1--- 1\n", "-1-- 1\n", "--1- 1\n"})
	Abc_NtkMakeSeq(n, 3)
	old := append([]*Object(nil), n.boxes...)
	name := filepath.Join(t.TempDir(), "flops.txt")
	if err := os.WriteFile(name, []byte("2\n0\n1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	Abc_NtkPermute(n, 0, 0, 0, name)
	offset := int(Abc_NtkPiNum(n) + Abc_NtkPoNum(n))
	want := []int32{1, 2, 0}
	for i, index := range want {
		if n.ObjectPermutation[offset+i] != index {
			t.Fatalf("saved index%d: got%d want%d", i, n.ObjectPermutation[offset+i], index)
		}
	}
	Abc_NtkUnpermute(n)
	for i, index := range want {
		if n.boxes[i] != old[index] {
			t.Fatal("source inverse order not applied")
		}
	}
}

func TestStartReadPreservesPathName(t *testing.T) {
	n := Abc_NtkStartRead("designs/alu.v1.blif")
	if n.Name != "designs/alu.v1" || n.Spec != "designs/alu.v1.blif" {
		t.Fatalf("name%q spec%q", n.Name, n.Spec)
	}
}
func TestChangedGraphsDoNotRetainStaleWitness(t *testing.T) {
	n := ntkFixture()
	n.SequenceModel = util.Abc_CexAlloc(0, 2, 1)
	n.Onehots = [][]int32{{0, 1}}
	n.Phases = []int32{1, 0}
	n.WireLoad = "old load"
	n.Constraints = 1
	for _, g := range []*Network{Abc_NtkDouble(n), Abc_NtkBottom(n, 1), Abc_NtkCreateCone(n, n.kindOrder[ObjectNode][0], "cone", 1)} {
		if g.SequenceModel != nil || len(g.Onehots) != 0 || len(g.Phases) != 0 || g.WireLoad != "" || g.Constraints != 0 {
			t.Fatal("changed graph retained source-dependent metadata")
		}
	}
	Abc_NtkMakeSeq(n, 1)
	g := Abc_NtkStartFromNoLatches(n, ABC_NTK_LOGIC, ABC_FUNC_SOP)
	if g.SequenceModel != nil || len(g.Onehots) != 0 || len(g.Phases) != 0 || g.WireLoad != "" {
		t.Fatal("removed latches retained source-dependent metadata")
	}
	if g.Constraints != n.Constraints || g.Name != n.Name || g.Spec != n.Spec {
		t.Fatal("preserved no-latch interface metadata lost")
	}
}
func TestMiterTransformMetadataPolicy(t *testing.T) {
	n := ntkFixture()
	g := Abc_NtkCreateTarget(n, []*Object{n.cos[0]}, []int32{1})
	Abc_ObjAddFanin(g.NewObject(ObjectPO), g.cis[0])
	g.Name, g.Spec = "miter", "miter.aig"
	g.Constraints = 1
	g.Barriers = 2
	g.SequenceModel = util.Abc_CexAlloc(0, 2, 1)
	g.Onehots = [][]int32{{0, 1}}
	g.Phases = []int32{1}
	g.WireLoad = "old load"
	out := Abc_NtkDupTransformMiter(g)
	if out.SequenceModel != nil || len(out.Onehots) != 0 || len(out.Phases) != 0 || out.WireLoad != "" {
		t.Fatal("miter transformation retained stale metadata")
	}
	if out.Constraints != 1 || out.Barriers != 2 || out.Name != g.Name || out.Spec != g.Spec {
		t.Fatal("miter interface metadata lost")
	}
}

type ntkTestGate struct {
	a, b, c int32
	kind    int
}
type ntkTestLiteralNetwork struct {
	gates   []ntkTestGate
	outputs []int32
}

func (g ntkTestLiteralNetwork) ObjectCount() int        { return len(g.gates) }
func (g ntkTestLiteralNetwork) InputLiterals() []int32  { return []int32{2, 4} }
func (g ntkTestLiteralNetwork) OutputLiterals() []int32 { return g.outputs }
func (g ntkTestLiteralNetwork) Gate(id int32) (int32, int32, int32, int) {
	v := g.gates[id]
	return v.a, v.b, v.c, v.kind
}
func TestLiteralNetworkConstruction(t *testing.T) {
	source := ntkTestLiteralNetwork{gates: []ntkTestGate{{kind: 0}, {kind: 1}, {kind: 1}, {a: 2, b: 4, kind: 2}, {a: 2, b: 3, kind: 2}, {a: 7, kind: 5}, {a: 10, kind: 6}}, outputs: []int32{12}}
	single := Abc_NtkCreateFromGias("single", []NtkLiteralNetwork{source}, nil)
	source.outputs = []int32{12, 8}
	multi := Abc_NtkCreateFromGias("multi", nil, source)
	for x := 0; x < 4; x++ {
		for _, n := range []*Network{single, multi} {
			inputs := map[*Object]bool{n.cis[0]: x&1 != 0, n.cis[1]: x&2 != 0}
			got, err := n.Evaluate(n.cos[0].Signal(), inputs)
			if err != nil || got != (x != 3) {
				t.Fatal("literal NAND conversion")
			}
			if n == multi {
				got, err = n.Evaluate(n.cos[1].Signal(), inputs)
				if err != nil || got {
					t.Fatal("complemented simplified AND lost")
				}
			}
		}
	}
}
func TestConeArrayAndAppend(t *testing.T) {
	source := ntkFixture()
	array := Abc_NtkCreateConeArray(source, source.cos, 1)
	ntkCheckSame(t, source, array)
	strash := Abc_NtkCreateTarget(source, []*Object{source.cos[0]}, []int32{1})
	part := Abc_NtkCreateConeArray(strash, strash.cos, 1)
	beforeIn, beforeOut := len(part.cis), len(part.cos)
	Abc_NtkAppendToCone(part, strash, strash.cos)
	if len(part.cis) != beforeIn || len(part.cos) != beforeOut {
		t.Fatal("append duplicated interface")
	}
	ntkCheckSame(t, strash, part)
}
