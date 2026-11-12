package abc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestObjectNamesAndFakeNames(t *testing.T) {
	n := NewNetwork("names")
	named, unnamed := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	if got := Abc_ObjAssignName(named, "n2", ""); got != "n2" {
		t.Fatal(got)
	}
	if got := Abc_ObjName(unnamed); got != "n2_1" {
		t.Fatalf("unique temporary name %q", got)
	}
	if got := Abc_ObjAssignName(unnamed, "x", "_suffix"); got != "x_suffix" {
		t.Fatal(got)
	}
	if Abc_ObjNamePrefix(unnamed, "pre_") != "pre_x_suffix" || Abc_ObjNameDummy("p", 7, 3) != "p007" {
		t.Fatal("name formatting")
	}
	names := Abc_NodeGetFakeNames(27)
	if vec.Vec_PtrEntry(names, 0) != "a0" || vec.Vec_PtrEntry(names, 26) != "a1" {
		t.Fatal("fake name order")
	}
	Abc_NodeFreeNames(names)
}

func TestTransferNetlistTerminalNames(t *testing.T) {
	old := NewNetwork("old")
	old.Type = uint8(ABC_NTK_NETLIST)
	pi := old.NewObject(ObjectPI)
	net := old.NewObject(ObjectNet)
	po := old.NewObject(ObjectPO)
	Abc_ObjAddFanin(net, pi)
	Abc_ObjAddFanin(po, net)
	Abc_ObjAssignName(net, "signal", "")
	dst := NewNetwork("new")
	outPI, outPO := dst.NewObject(ObjectPI), dst.NewObject(ObjectPO)
	pi.Copy, po.Copy = outPI, outPO
	Abc_NtkTrasferNames(old, dst)
	if Abc_ObjName(outPI) != "signal" || Abc_ObjName(outPO) != "signal" {
		t.Fatal("net names were not transferred to both terminals")
	}
	if got := Abc_NtkCollectCioNames(dst, 1); len(got) != 1 || got[0] != "signal" {
		t.Fatal(got)
	}
}

func TestOrderNamesPreservesLatchTerminalPairing(t *testing.T) {
	n := NewNetwork("sort")
	z, a := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	z.SetName("z")
	a.SetName("a")
	qz := Abc_NtkAddLatch(n, z, ABC_INIT_ZERO)
	qa := Abc_NtkAddLatch(n, a, ABC_INIT_ONE)
	qz.SetName("z_state")
	qa.SetName("a_state")
	pz, pa := n.NewObject(ObjectPO), n.NewObject(ObjectPO)
	pz.SetName("z_out")
	pa.SetName("a_out")
	Abc_ObjAddFanin(pz, qz)
	Abc_ObjAddFanin(pa, qa)
	Abc_NtkOrderCisCos(n)
	Abc_NtkOrderObjsByName(n, 1)
	if n.cis[0] != a || n.cis[1] != z || n.cis[2] != qa || n.cis[3] != qz || n.cos[0] != pa || n.cos[1] != pz {
		t.Fatal("incorrect alphabetical terminal order")
	}
	if n.cos[2] != qa.fanins[0].Object.fanins[0].Object || n.cos[3] != qz.fanins[0].Object.fanins[0].Object {
		t.Fatal("latch input and output order diverged")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestTransferTerminalOrderByName(t *testing.T) {
	old := NewNetwork("old")
	for _, name := range []string{"a", "b"} {
		old.NewObject(ObjectPI).SetName(name)
	}
	for _, name := range []string{"x", "y"} {
		old.NewObject(ObjectPO).SetName(name)
	}
	new := NewNetwork("new")
	for _, name := range []string{"b", "a"} {
		new.NewObject(ObjectPI).SetName(name)
	}
	for _, name := range []string{"y", "x"} {
		new.NewObject(ObjectPO).SetName(name)
	}
	if Abc_NodeCompareCiCo(old, new) != 0 {
		t.Fatal("unsorted interfaces matched")
	}
	Abc_NtkTransferOrder(old, new)
	if Abc_NodeCompareCiCo(old, new) != 1 {
		t.Fatal("interface order not transferred")
	}
	if Abc_ObjName(new.ObjectsOfKind(ObjectPI)[0]) != "a" || Abc_ObjName(new.ObjectsOfKind(ObjectPO)[0]) != "x" {
		t.Fatal("per-kind order disagrees with terminal order")
	}
}

func TestDummyLatchNamesAvoidExistingPrefixes(t *testing.T) {
	n := NewNetwork("prefix")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	pi := n.NewObject(ObjectPI)
	pi.SetName("lli0")
	input, latch, output := newBarrier(n, ABC_INIT_ZERO)
	Abc_ObjAddFanin(input, pi)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, output)
	Abc_NtkAddDummyBoxNames(n)
	if Abc_ObjName(input) != "llli0" || Abc_ObjName(output) != "lllo0" || Abc_ObjName(latch) != "l0" {
		t.Fatalf("input=%s output=%s latch=%s", Abc_ObjName(input), Abc_ObjName(output), Abc_ObjName(latch))
	}
	Abc_NtkShortNames(n)
	if Abc_ObjName(pi) != "pi0" || Abc_ObjName(po) != "po0" || Abc_ObjName(input) != "li0" || Abc_ObjName(output) != "lo0" {
		t.Fatal("short names do not match interface")
	}
}

func TestCleanNamesAndRedirectBufferPolarity(t *testing.T) {
	n := NewNetwork("redirect")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	pi := n.NewObject(ObjectPI)
	Abc_ObjAssignName(pi, "same", "")
	buffer := Abc_NtkCreateNodeBuf(n, pi)
	buffer.SetName("temporary")
	if err := n.PatchFanin(buffer, 0, pi.Signal().Not()); err != nil {
		t.Fatal(err)
	}
	po := n.NewObject(ObjectPO)
	Abc_ObjAssignName(po, "same", "")
	Abc_ObjAddFanin(po, buffer.Signal().Not())
	Abc_NtkRedirectCiCo(n)
	if len(po.fanins) != 1 || po.fanins[0] != pi.Signal() || n.contains(buffer) {
		t.Fatal("redirect lost composed polarity or left the buffer")
	}
	Abc_NtkCleanNames(n)
	if Abc_ObjName(pi) != "same" || Abc_ObjName(po) != "same" {
		t.Fatal("terminal names lost during cleanup")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestNameIDFileTransferAndComplement(t *testing.T) {
	old := NewNetwork("netlist")
	old.Type = uint8(ABC_NTK_NETLIST)
	old.Function = uint8(ABC_FUNC_SOP)
	old.Spec = filepath.Join(t.TempDir(), "sample.blif")
	pi := old.NewObject(ObjectPI)
	inputNet := old.NewObject(ObjectNet)
	inputNet.SetName("input")
	Abc_ObjAddFanin(inputNet, pi)
	node := old.NewObject(ObjectNode)
	node.Data = "1 1\n"
	Abc_ObjAddFanin(node, inputNet)
	outputNet := old.NewObject(ObjectNet)
	outputNet.SetName("output")
	Abc_ObjAddFanin(outputNet, node)
	po := old.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, outputNet)
	Abc_NtkOrderCisCos(old)
	Abc_NtkStartNameIds(old)
	dst := NewNetwork("logic")
	dst.Type = uint8(ABC_NTK_LOGIC)
	dst.Spec = old.Spec
	newPI, newPO := dst.NewObject(ObjectPI), dst.NewObject(ObjectPO)
	setCopy(pi, newPI.Signal())
	setCopy(po, newPO.Signal().Not())
	Abc_NtkTransferNameIds(old, dst)
	if dst.NameIDs[newPI.ID] != 2 || dst.NameIDs[newPO.ID] != 5 {
		t.Fatalf("bad transferred IDs %v", dst.NameIDs)
	}
	Abc_NtkUpdateNameIds(dst)
	data, err := os.ReadFile(abcNameIDFile(dst.Spec))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 || strings.Join(strings.Fields(lines[0]), " ") != "input 1" || strings.Join(strings.Fields(lines[1]), " ") != "output -2" {
		t.Fatalf("unexpected file %q", data)
	}
	if dst.NameIDs != nil {
		t.Fatal("name IDs not cleared")
	}
}
