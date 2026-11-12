package abc

import (
	"testing"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/opt/sfm"
)

func TestAbcExtractSfmMappedNetwork(t *testing.T) {
	n, _, _, pi, driver, consumer := abcSCLFixture(t)
	mapped, objects, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 4 || objects[0] != pi || objects[1] != driver || objects[2] != consumer || objects[3] != n.Outputs()[0] {
		t.Fatalf("object order: %v", objects)
	}
	for _, input := range []bool{false, true} {
		out, err := mapped.Evaluate([]bool{input})
		if err != nil || len(out) != 1 || out[0] != input {
			t.Fatalf("input %v: %v %v", input, out, err)
		}
	}
	consumer.fanins[0].Complement = true
	if _, _, err := Abc_NtkExtractSfm(n); err == nil {
		t.Fatal("complemented mapped edge accepted")
	}
}

func TestAbcApplySfmMappedNetwork(t *testing.T) {
	n, _, _, pi, driver, consumer := abcSCLFixture(t)
	p, original, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	// Remove the first buffer and reconnect the second directly to the input.
	p.Deleted[1] = true
	p.Fanins[1] = nil
	p.Fanins[2] = []int{0}
	if err := Abc_NtkApplySfm(n, p, original); err != nil {
		t.Fatal(err)
	}
	if n.Object(driver.ID) != nil || n.Object(consumer.ID) != consumer || len(consumer.fanins) != 1 || consumer.fanins[0].Object != pi {
		t.Fatal("mapped network replacement did not preserve live identities")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestAbcApplySfmRejectsInvalidGateWithoutMutation(t *testing.T) {
	n, _, _, _, _, consumer := abcSCLFixture(t)
	p, original, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	old := consumer.fanins[0]
	p.Mapping[2] = nil
	if err := Abc_NtkApplySfm(n, p, original); err == nil {
		t.Fatal("invalid gate accepted")
	}
	if consumer.fanins[0] != old {
		t.Fatal("failed validation changed network")
	}
}

func TestAbcPerformMfs3MappedNetwork(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 2 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 2 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 2 0
.end
`
	library, err := mio.ReadLibrary("fixture", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	n := NewNetwork("mapped and")
	n.Type, n.Function, n.FunctionManager = uint8(ABC_NTK_LOGIC), uint8(ABC_FUNC_MAP), library
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	x, y := n.NewObject(ObjectNode), n.NewObject(ObjectNode)
	x.Data, y.Data = mio.Mio_LibraryReadGateByName(library, "and2", ""), mio.Mio_LibraryReadGateByName(library, "and2", "")
	Abc_ObjAddFanin(x, a)
	Abc_ObjAddFanin(x, b)
	Abc_ObjAddFanin(y, x)
	Abc_ObjAddFanin(y, a)
	out := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(out, y)
	pars := sfm.Sfm_Par_t{Area: true, TfoLevMax: 2, TfiLevMax: 3, FanoutMax: 10, MffcMax: 3, VarMax: 6, UseAndOr: true, NodeOne: y.ID}
	pars.NodesMax = y.ID - 1
	changed, err := Abc_NtkPerformMfs3(n, &pars)
	if err != nil || changed != 0 {
		t.Fatalf("node limit: changed %d: %v", changed, err)
	}
	pars.NodesMax = y.ID
	changed, err = Abc_NtkPerformMfs3(n, &pars)
	if err != nil || changed != 1 {
		t.Fatalf("changed %d: %v", changed, err)
	}
	p, _, err := Abc_NtkExtractSfm(n)
	if err != nil {
		t.Fatal(err)
	}
	for a := 0; a < 4; a++ {
		got, err := p.Evaluate([]bool{a&1 != 0, a&2 != 0})
		if err != nil || len(got) != 1 || got[0] != (a == 3) {
			t.Fatalf("input %d: %v %v", a, got, err)
		}
	}
}
