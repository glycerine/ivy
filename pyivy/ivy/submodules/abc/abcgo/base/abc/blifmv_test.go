package abc

import (
	"reflect"
	"strings"
	"testing"
)

func TestMVTableParserRoundtrip(t *testing.T) {
	table, err := ParseMVTable("d=0\n(0, 2) - 1\n- =0 =1\n", []int{3, 3, 3})
	if err != nil {
		t.Fatal(err)
	}
	again, err := ParseMVTable(table.String(), table.DomainSizes)
	if err != nil || !reflect.DeepEqual(table, again) {
		t.Fatal(err, table, again)
	}
	for a := 0; a < 3; a++ {
		for b := 0; b < 3; b++ {
			before, _ := table.Evaluate([]int{a, b})
			after, _ := again.Evaluate([]int{a, b})
			if !reflect.DeepEqual(before, after) {
				t.Fatal(a, b, before, after)
			}
		}
	}
	fixed, err := ParseMVTable("d0\n!0 (1,2) 2\n", []int{3, 3, 3})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		in  []int
		out []int
	}{{[]int{1, 1}, []int{2}}, {[]int{0, 1}, []int{0}}, {[]int{2, 0}, []int{0}}} {
		got, err := fixed.Evaluate(test.in)
		if err != nil || !reflect.DeepEqual(got, test.out) {
			t.Fatal(test, got, err)
		}
	}
	for _, text := range []string{"(0,3) - 1\n", "- =2 1\n", "=0 - 1\n", "- - 4\n", "(0,2 - 1\n", "d0\nd1\n"} {
		if _, err := ParseMVTable(text, []int{3, 3, 3}); err == nil {
			t.Fatal("accepted malformed table", text)
		}
	}
}
func makeMVPermutation(t *testing.T) *Network {
	t.Helper()
	n := testModule("mv")
	n.Function = uint8(ABC_FUNC_BLIFMV)
	input := testInputNet(n, "input")
	Abc_NtkSetMvVarValues(input, 3)
	node := n.NewObject(ObjectNode)
	node.Data = "0 1\n1 2\n2 0\n"
	Abc_ObjAddFanin(node, input)
	output := netForDriver(n, node.Signal())
	output.SetName("output")
	Abc_NtkSetMvVarValues(output, 3)
	po := n.NewObject(ObjectPO)
	po.SetName("output")
	Abc_ObjAddFanin(po, output)
	return n
}
func TestMVStrashEncodingTruthTable(t *testing.T) {
	old := makeMVPermutation(t)
	encoded := Abc_NtkStrashBlifMv(old)
	if encoded == nil {
		t.Fatal("MV strash failed")
	}
	if len(encoded.ObjectsOfKind(ObjectPI)) != 2 || len(encoded.ObjectsOfKind(ObjectPO)) != 2 {
		t.Fatal("binary interface")
	}
	for value, want := range []int{1, 2, 0} {
		inputs := map[*Object]bool{}
		for bit, pi := range encoded.ObjectsOfKind(ObjectPI) {
			inputs[pi] = value&(1<<bit) != 0
		}
		got := 0
		for bit, po := range encoded.ObjectsOfKind(ObjectPO) {
			v, err := encoded.Evaluate(po.Signal(), inputs)
			if err != nil {
				t.Fatal(err)
			}
			if v {
				got |= 1 << bit
			}
		}
		if got != want {
			t.Fatal(value, got, want)
		}
	}
	for _, net := range old.ObjectsOfKind(ObjectNet) {
		if net.MVEncoding != nil {
			t.Fatal("temporary encoding retained")
		}
	}
	if Abc_AigCheck(Abc_AigAlloc(encoded)) != 1 {
		t.Fatal("invalid encoded AIG")
	}
}
func TestMVVariableLifetimeAndPackedCost(t *testing.T) {
	n := testModule("attributes")
	net := n.NewObject(ObjectNet)
	if Abc_ObjMvVarNum(net) != 2 {
		t.Fatal("default binary")
	}
	Abc_NtkStartMvVars(n)
	Abc_NtkSetMvVarValues(net, 4)
	Abc_NtkSetMvVarValues(net, 5)
	if Abc_ObjMvVarNum(net) != 4 {
		t.Fatal("reassignment")
	}
	Abc_NtkFreeMvVars(n)
	if Abc_ObjMvVarNum(net) != 2 {
		t.Fatal("attribute release")
	}
	text := "123 rest"
	if Abc_StringGetNumber(&text) != 123 || text != " rest" {
		t.Fatal(text)
	}
	packed := Abc_NodeConvertSopToMvSop(2, []uint32{1}, []uint32{2 | 2<<2})
	if packed != "0-0\n111\n" || Abc_NodeEvalMvCost(2, []uint32{1}, []uint32{2 | 2<<2}) != 2 {
		t.Fatal(packed)
	}
	if Abc_NodeConvertSopToMvSop(3, nil, []uint32{0}) != "---1\n" {
		t.Fatal("constant packed cover")
	}
}
func TestMVSkeletonInsertionRoundtrip(t *testing.T) {
	base := makeMVPermutation(t)
	skeleton := Abc_NtkSkeletonBlifMv(base)
	if skeleton == nil {
		t.Fatal("skeleton")
	}
	box := skeleton.ObjectsOfKind(ObjectWhitebox)[0]
	if len(box.fanins) != 2 || len(box.fanouts) != 2 || Abc_ObjMvVarNum(skeleton.ObjectsOfKind(ObjectPI)[0].fanouts[0]) != 3 {
		t.Fatal("MV skeleton interface")
	}
	logic := testModule("implementation")
	low, high := testInputNet(logic, "low"), testInputNet(logic, "high")
	lowNode := createLogicNode(logic, []*Object{low, high}, func(v []bool) bool { return !v[0] && !v[1] })
	highNode := createLogicNode(logic, []*Object{low, high}, func(v []bool) bool { return v[0] && !v[1] })
	testOutputNet(logic, "out0", lowNode.Signal())
	testOutputNet(logic, "out1", highNode.Signal())
	if Abc_NtkConvertToBlifMv(logic) != 1 {
		t.Fatal("Boolean-to-MV conversion")
	}
	restored := Abc_NtkInsertBlifMv(base, logic)
	if restored == nil {
		t.Fatal("insertion")
	}
	if len(restored.ObjectsOfKind(ObjectWhitebox)) != 0 {
		t.Fatal("internal box not flattened")
	}
	for value, want := range []int{1, 2, 0} {
		pi := restored.ObjectsOfKind(ObjectPI)[0]
		po := restored.ObjectsOfKind(ObjectPO)[0]
		got, err := restored.EvaluateMV(po, map[*Object]int{pi: value})
		if err != nil || !reflect.DeepEqual(got, []int{want}) {
			t.Fatal(value, got, err, want)
		}
	}
}
func TestSopToMVDefaultPreservesBothPolarities(t *testing.T) {
	for _, cover := range []string{"1- 1\n-1 1\n", "00 0\n"} {
		table, err := boolCoverToMV(cover, 2)
		if err != nil {
			t.Fatal(err)
		}
		for pattern := 0; pattern < 4; pattern++ {
			inputs := []int{pattern & 1, (pattern >> 1) & 1}
			got, err := table.Evaluate(inputs)
			want := 0
			if pattern != 0 {
				want = 1
			}
			if err != nil || !reflect.DeepEqual(got, []int{want}) {
				t.Fatal(cover, pattern, got, err)
			}
		}
	}
	table, err := ParseMVTable("- 0\n- 1\n", []int{2, 2})
	if err != nil {
		t.Fatal(err)
	}
	values, _ := table.Evaluate([]int{0})
	if !reflect.DeepEqual(values, []int{0, 1}) {
		t.Fatal("nondeterminism lost", values)
	}
	if !strings.Contains(table.String(), "- 1\n") {
		t.Fatal(table)
	}
}

func TestMVEqualityUsesInputDomainNotOutputDomain(t *testing.T) {
	n := testModule("equality")
	a, b := testInputNet(n, "a"), testInputNet(n, "b")
	Abc_NtkSetMvVarValues(a, 3)
	Abc_NtkSetMvVarValues(b, 3)
	node := n.NewObject(ObjectNode)
	node.Data = "d0\n- =0 1\n"
	Abc_ObjAddFanin(node, a)
	Abc_ObjAddFanin(node, b)
	testOutputNet(n, "equal", node.Signal())
	encoded := Abc_NtkStrashBlifMv(n)
	if encoded == nil {
		t.Fatal("strash equality")
	}
	for av := 0; av < 3; av++ {
		for bv := 0; bv < 3; bv++ {
			values := map[*Object]bool{}
			for i, pi := range encoded.ObjectsOfKind(ObjectPI) {
				value := av
				if i >= 2 {
					value = bv
				}
				values[pi] = value&(1<<(i%2)) != 0
			}
			got, err := encoded.Evaluate(encoded.ObjectsOfKind(ObjectPO)[0].Signal(), values)
			if err != nil || got != (av == bv) {
				t.Fatal(av, bv, got, err)
			}
		}
	}
}
func TestMVDefaultEqualityAndFailureCleanup(t *testing.T) {
	n := makeMVPermutation(t)
	node := n.ObjectsOfKind(ObjectNode)[0]
	node.Data = "d=0\n0 2\n"
	encoded := Abc_NtkStrashBlifMv(n)
	if encoded == nil {
		t.Fatal("default equality")
	}
	for value, want := range []int{2, 1, 2} {
		values := map[*Object]bool{}
		for bit, pi := range encoded.ObjectsOfKind(ObjectPI) {
			values[pi] = value&(1<<bit) != 0
		}
		got := 0
		for bit, po := range encoded.ObjectsOfKind(ObjectPO) {
			v, err := encoded.Evaluate(po.Signal(), values)
			if err != nil {
				t.Fatal(err)
			}
			if v {
				got |= 1 << bit
			}
		}
		if got != want {
			t.Fatal(value, got, want)
		}
	}
	node.Data = "invalid"
	if got := Abc_NtkStrashBlifMv(n); got != nil {
		t.Fatal("accepted malformed cover")
	}
	for _, net := range n.ObjectsOfKind(ObjectNet) {
		if net.MVEncoding != nil {
			t.Fatal("failed conversion retained encoding graph")
		}
	}
}
func TestMVNondeterministicConstantSequentialInterface(t *testing.T) {
	n := testModule("free-var")
	testInputNet(n, "normal")
	constant := n.NewObject(ObjectNode)
	constant.Data = "-\n"
	net := netForDriver(n, constant.Signal())
	net.SetName("choice")
	in, latch, out := newBarrier(n, ABC_INIT_ZERO)
	Abc_ObjAddFanin(in, net)
	state := netForDriver(n, out.Signal())
	state.SetName("state")
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, state)
	_ = latch
	encoded := Abc_NtkStrashBlifMv(n)
	if encoded == nil {
		t.Fatal("sequential conversion")
	}
	cis := encoded.Inputs()
	if len(cis) != 3 || !strings.HasPrefix(cis[0].Name(), "free_var_") || cis[1].Kind != ObjectPI || cis[2].Kind != ObjectBO {
		t.Fatal("free variable/register ordering", cis)
	}
	if len(encoded.ObjectsOfKind(ObjectLatch)) != 1 {
		t.Fatal("register bit missing")
	}
	values := map[*Object]bool{}
	for _, ci := range cis {
		values[ci] = strings.HasPrefix(ci.Name(), "free_var_")
	}
	got, err := encoded.Evaluate(encoded.ObjectsOfKind(ObjectBI)[0].Signal(), values)
	if err != nil || !got {
		t.Fatal("free variable not connected to next state", got, err)
	}
}

func TestDeletedMVObjectsReleaseAttributes(t *testing.T) {
	n := testModule("deleted-attribute")
	net := n.NewObject(ObjectNet)
	Abc_NtkSetMvVarValues(net, 3)
	Abc_NtkDeleteObj(net)
	if len(n.mvVars) != 0 {
		t.Fatal("deleted object retained in MV attribute manager")
	}
}
func TestSymbolicBinaryMVVariableSetter(t *testing.T) {
	n := NewNetwork("symbolic")
	o := n.NewObject(ObjectNet)
	variable := &MVVariable{Values: 2, Names: []string{"low", "high"}}
	if err := SetMVVariable(o, variable); err != nil {
		t.Fatal(err)
	}
	variable.Names[0] = "changed"
	if Abc_ObjMvVarNum(o) != 2 || Abc_ObjMvVar(o).Names[0] != "low" {
		t.Fatal("binary labels aliased or lost")
	}
	for _, bad := range []*MVVariable{{Values: 1}, {Values: 2, Names: []string{"one"}}, {Values: 2, Names: []string{"x", "x"}}, {Values: 2, Names: []string{"", "x"}}} {
		if err := SetMVVariable(o, bad); err == nil {
			t.Fatal("invalid domain accepted", bad)
		}
	}
	if Abc_ObjMvVar(o).Names[0] != "low" {
		t.Fatal("invalid update changed domain")
	}
	if err := SetMVVariable(o, nil); err != nil || Abc_ObjMvVar(o) != nil {
		t.Fatal("clear failed", err)
	}
	if err := n.DeleteObject(o); err != nil {
		t.Fatal(err)
	}
	if err := SetMVVariable(o, variable); err == nil {
		t.Fatal("annotated deleted object")
	}
}
