package abci

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestDetectObjectClassesByRecursiveFanoutSet(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_AIG, 1)
	n.Name = "tfo-classes"
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	shared := n.CreateAndEdges(a.Signal(), b.Signal())
	po0, po1 := n.NewObject(abc.ObjectPO), n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po0, shared.Object)
	abc.Abc_ObjAddFanin(po1, shared.Object)
	po2 := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po2, c)

	objects := vec.Vec_IntAlloc(3)
	for _, object := range []*abc.Object{a, b, c} {
		Vec_IntPush(objects, int32(object.ID))
	}
	var coSets *Vec_Wec_t
	classes := Abc_NtkDetectObjClasses(n, objects, &coSets)
	defer vec.Vec_WecFree(classes)
	defer vec.Vec_WecFree(coSets)
	defer Vec_IntFree(objects)
	if got := vec.Vec_WecSize(classes); got != 2 {
		t.Fatalf("class count %d, want {a,b} and {c}", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(classes, 0)); !reflect.DeepEqual(got, []int32{int32(a.ID), int32(b.ID)}) {
		t.Fatalf("first class = %v", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(coSets, 0)); !reflect.DeepEqual(got, []int32{int32(po0.ID), int32(po1.ID)}) {
		t.Fatalf("shared TFO CO set = %v", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(coSets, 1)); !reflect.DeepEqual(got, []int32{int32(po2.ID)}) {
		t.Fatalf("separate TFO CO set = %v", got)
	}
}

func TestFaultTypeNamesAndSavedNetworkMetadata(t *testing.T) {
	for _, test := range []struct {
		name string
		kind int32
	}{
		{"SA0", -99}, {"SA1", -98}, {"NEG", -97}, {"RDOB_AND", -96},
		{"RDOB_NAND", -95}, {"RDOB_OR", -94}, {"RDOB_NOR", -93},
		{"RDOB_XOR", -92}, {"RDOB_NXOR", -91}, {"RDOB_NOT", -90}, {"RDOB_BUFF", -89},
	} {
		if got := Io_ReadFinType(test.name); got != test.kind || Io_WriteFinType(got) != test.name {
			t.Fatalf("fault type %q mapped to %d and back to %q", test.name, got, Io_WriteFinType(got))
		}
	}
	if Io_ReadFinType("not-a-fault") != int32(ABC_FIN_NONE) || Io_WriteFinType(-88) != "Unknown" {
		t.Fatal("unknown fault type handling")
	}
	n := abc.NewNetwork("fault-metadata")
	pairs := vec.Vec_IntAlloc(2)
	Vec_IntPush(pairs, 12)
	Vec_IntPush(pairs, int32(ABC_FIN_SA0))
	abc.Abc_NtkSetFins(n, pairs)
	if got := abc.Abc_NtkReadFins(n); got != pairs || vec.Vec_IntEntry(got, 0) != 12 {
		t.Fatal("network fault metadata accessor lost the source vector")
	}
}

func TestFrameExtendBuildsSecondFrameAndLatchInput(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	n.Constraints = 1
	x := n.NewObject(abc.ObjectPI)
	state := n.NewObject(abc.ObjectPI)
	for i := 0; i < 8; i++ {
		po := n.NewObject(abc.ObjectPO)
		driver := x
		if i == 6 {
			driver = state
		}
		abc.Abc_ObjAddFanin(po, driver)
	}
	Abc_NtkFrameExtend(n)
	if got := abc.Abc_NtkPiNum(n); got != 3 {
		t.Fatalf("PI count after frame extension = %d, want original two plus copied ordinary input", got)
	}
	if got := abc.Abc_NtkPoNum(n); got != 12 {
		t.Fatalf("PO count after frame extension = %d, want four ordinary output copies", got)
	}
	if got := abc.Abc_ObjName(abc.Abc_NtkPi(n, 2)); got != abc.Abc_ObjName(x)+"_frame1" {
		t.Fatalf("copied PI name = %q", got)
	}
	if got := abc.Abc_ObjFaninNum(abc.Abc_NtkPo(n, 8)); got != 1 {
		t.Fatalf("copied PO fanin count = %d", got)
	}
	if err := n.Check(); err != nil {
		t.Fatalf("network invalid after frame extension: %v", err)
	}
}

func TestDetectFinClassesUsesExactFaultedOutputFunctions(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	n.Name = "fault-equivalence"
	x, y := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	n1 := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{x, y})
	n2 := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{x, y})
	out := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{n1, n2})
	po := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po, out)
	pairs := vec.Vec_IntAlloc(14)
	vec.Vec_IntPush(pairs, -1)
	vec.Vec_IntPush(pairs, -1)
	for _, node := range []*abc.Object{n1, n2} {
		for _, kind := range []int32{int32(ABC_FIN_SA0), int32(ABC_FIN_SA1), int32(ABC_FIN_NEG)} {
			vec.Vec_IntPush(pairs, int32(node.ID))
			vec.Vec_IntPush(pairs, kind)
		}
	}
	abc.Abc_NtkSetFins(n, pairs)
	classes := Abc_NtkDetectFinClasses(n, 0)
	if classes == nil {
		t.Fatal("fault detector rejected a supported SOP network")
	}
	defer vec.Vec_WecFree(classes)
	want := [][]int32{{1, 3, 4, 6}, {2, 5}}
	if got := vec.Vec_WecSize(classes); got != len(want) {
		t.Fatalf("class count %d, want %d", got, len(want))
	}
	for i, expected := range want {
		got := vec.Vec_IntArray(vec.Vec_WecEntry(classes, i))
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("class %d = %v, want %v", i, got, expected)
		}
	}
}

func TestFinRefinementWorklistMatchesSourceOrderAndCallCount(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	x, y := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	n1 := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{x, y})
	n2 := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{x, y})
	root := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{n1, n2})
	po := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po, root)
	pairs := vec.Vec_IntAlloc(14)
	vec.Vec_IntPush(pairs, -1)
	vec.Vec_IntPush(pairs, -1)
	for _, node := range []*abc.Object{n1, n2} {
		for _, kind := range []int32{int32(ABC_FIN_SA0), int32(ABC_FIN_SA1), int32(ABC_FIN_NEG)} {
			vec.Vec_IntPush(pairs, int32(node.ID))
			vec.Vec_IntPush(pairs, kind)
		}
	}
	objects := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(objects, int32(n1.ID))
	vec.Vec_IntPush(objects, int32(n2.ID))
	var coSets *Vec_Wec_t
	objectClasses := Abc_NtkDetectObjClasses(n, objects, &coSets)
	defer vec.Vec_WecFree(objectClasses)
	defer vec.Vec_WecFree(coSets)
	ciSet, nodeSet := vec.Vec_IntAlloc(4), vec.Vec_IntAlloc(4)
	Abc_NtkFinMiterCollect(n, vec.Vec_WecEntry(coSets, 0), ciSet, nodeSet)
	faultList := vec.Vec_IntAlloc(6)
	for i := int32(1); i <= 6; i++ {
		vec.Vec_IntPush(faultList, i)
	}
	types := Abc_NtkFinComputeTypes(n)
	classes, calls := finRefineFaultClass(n, types, vec.Vec_WecEntry(coSets, 0), ciSet, nodeSet, pairs, faultList)
	if calls != 4 {
		t.Fatalf("SAT calls = %d, want source worklist count 4", calls)
	}
	want := [][]int32{{1, 3, 4, 6}, {2, 5}}
	if len(classes) != len(want) {
		t.Fatalf("refinement classes = %v, want %v", classes, want)
	}
	for i := range want {
		got := make([]int32, len(classes[i]))
		for j, item := range classes[i] {
			got[j] = item.fault
		}
		if !reflect.DeepEqual(got, want[i]) {
			t.Fatalf("refinement class %d = %v, want %v", i, got, want[i])
		}
	}
}

func TestDetectFinClassesUsesSatForNonstructuralEquivalentMiter(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	left0 := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{a, b})
	left1 := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{a, c})
	left := abc.Abc_NtkCreateNodeOr(n, []*abc.Object{left0, left1})
	right0 := abc.Abc_NtkCreateNodeOr(n, []*abc.Object{b, c})
	right := abc.Abc_NtkCreateNodeAnd(n, []*abc.Object{a, right0})
	root := n.NewObject(abc.ObjectNode)
	abc.Abc_ObjAddFanin(root, left)
	abc.Abc_ObjAddFanin(root, right)
	root.Data = "01 1\n10 1\n"
	po := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po, root)
	pairs := vec.Vec_IntAlloc(6)
	vec.Vec_IntPush(pairs, -1)
	vec.Vec_IntPush(pairs, -1)
	for _, node := range []*abc.Object{left, right} {
		vec.Vec_IntPush(pairs, int32(node.ID))
		vec.Vec_IntPush(pairs, int32(ABC_FIN_SA0))
	}
	abc.Abc_NtkSetFins(n, pairs)
	classes := Abc_NtkDetectFinClasses(n, 0)
	if classes == nil {
		t.Fatal("fault detector rejected a supported SOP network")
	}
	defer vec.Vec_WecFree(classes)
	if got := vec.Vec_WecSize(classes); got != 1 {
		t.Fatalf("class count %d, want the two SAT-proved equivalent faults", got)
	}
	if got := vec.Vec_IntArray(vec.Vec_WecEntry(classes, 0)); !reflect.DeepEqual(got, []int32{1, 2}) {
		t.Fatalf("equivalent fault class = %v, want [1 2]", got)
	}
}
