package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

func dressPrototype(n *abc.Network, representative *abc.Object, name string, complement bool) *abc.Object {
	prototype := n.NewObject(abc.ObjectNode)
	copy := representative.Signal()
	if complement {
		copy = copy.Not()
	}
	prototype.SetCopyEdge(copy)
	net := n.NewObject(abc.ObjectNet)
	abc.Abc_ObjAddFanin(net, prototype)
	net.SetName(name)
	return prototype
}

func TestDressDeriveMappingKeepsFirstPrototypeAndPolarity(t *testing.T) {
	original := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_SOP, 1)
	fraig := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	representative := fraig.NewObject(abc.ObjectNode)
	first := dressPrototype(original, representative, "first", true)
	dressPrototype(original, representative, "second", false)

	mapping := Abc_NtkDressDeriveMapping(original)
	value, exists := mapping.Lookup(representative)
	if !exists {
		t.Fatal("representative has no mapping")
	}
	got, ok := value.(abc.Edge)
	if !ok {
		t.Fatalf("mapping value has type %T, want abc.Edge", value)
	}
	if got.Object != first || !got.Complement {
		t.Fatalf("first mapping = (%p, complemented=%v), want first prototype %p complemented", got.Object, got.Complement, first)
	}
}

func TestDressTransferNamesHonorsExistingNamesAndInversion(t *testing.T) {
	original := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_SOP, 1)
	fraig := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	directRep := fraig.NewObject(abc.ObjectNode)
	invertedRep := fraig.NewObject(abc.ObjectNode)
	dressPrototype(original, directRep, "direct_net", false)
	dressPrototype(original, invertedRep, "inverted_net", true)
	mapping := Abc_NtkDressDeriveMapping(original)

	logic := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	preNamed := logic.NewObject(abc.ObjectNode)
	preNamed.SetName("preserved")
	preNamed.SetCopyEdge(directRep.Signal())
	direct := logic.NewObject(abc.ObjectNode)
	direct.SetCopyEdge(directRep.Signal())
	inverted := logic.NewObject(abc.ObjectNode)
	inverted.SetCopyEdge(invertedRep.Signal())

	Abc_NtkDressTransferNames(logic, mapping, 0)
	if got := direct.Name(); got != "direct_net" {
		t.Fatalf("direct transferred name = %q, want direct_net", got)
	}
	if got := inverted.Name(); got != "inverted_net_inv" {
		t.Fatalf("complemented transferred name = %q, want inverted_net_inv", got)
	}
	if got := preNamed.Name(); got != "preserved" {
		t.Fatalf("existing name changed to %q", got)
	}
	if got := mapping.Count(); got != 0 {
		t.Fatalf("unconsumed prototype mappings = %d, want 0", got)
	}
}
