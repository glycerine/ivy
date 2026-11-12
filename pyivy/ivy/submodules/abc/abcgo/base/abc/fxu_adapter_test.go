package abc

import (
	"reflect"
	"testing"
)

func TestFXUUnchangedRoundtripNormalizesComplementedEdges(t *testing.T) {
	n, node, pis := fanFixture("1-01 0\n-10- 0\n--11 0\n")
	before := fanTruth(node, pis)
	sops, fanins, err := n.FXUCovers()
	if err != nil {
		t.Fatal(err)
	}
	if err := n.ReplaceFXUCovers(sops, fanins); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, fanTruth(node, pis)) {
		t.Fatal("exported phase normalization changed function")
	}
	for _, edge := range node.fanins {
		if edge.Complement {
			t.Fatal("exported cover did not absorb phase")
		}
	}
}
func TestFXUForwardExtractedDefinitionsPreserveOutputObject(t *testing.T) {
	n, node, pis := fanFixture("1111 1\n")
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, node)
	before := fanTruth(node, pis)
	sops, fanins, err := n.FXUCovers()
	if err != nil {
		t.Fatal(err)
	}
	base := len(sops)
	sops = append(sops, "11 1\n", "11 1\n")
	fanins = append(fanins, []int32{int32(base + 1), int32(node.fanins[2].Object.ID)}, []int32{int32(node.fanins[0].Object.ID), int32(node.fanins[1].Object.ID)})
	sops[node.ID] = "11 1\n"
	fanins[node.ID] = []int32{int32(base), int32(node.fanins[3].Object.ID)}
	// Export absorbed the original complemented input into its column. Apply
	// the same inverse to the extracted divisor's corresponding local literal.
	sops[base] = "10 1\n"
	if err := n.ReplaceFXUCovers(sops, fanins); err != nil {
		t.Fatal(err)
	}
	if po.fanins[0].Object != node {
		t.Fatal("output identity changed")
	}
	if !reflect.DeepEqual(before, fanTruth(node, pis)) {
		t.Fatal("extracted graph function changed")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestFXUCycleAndMalformedChangesAreAtomic(t *testing.T) {
	n, node, _ := fanFixture("1111 1\n")
	sops, fanins, err := n.FXUCovers()
	if err != nil {
		t.Fatal(err)
	}
	oldData, oldFanins, count := node.Data, append([]Edge(nil), node.fanins...), len(n.objects)
	sops[node.ID] = "1 1\n"
	fanins[node.ID] = []int32{int32(node.ID)}
	if err := n.ReplaceFXUCovers(sops, fanins); err == nil {
		t.Fatal("cycle accepted")
	}
	if node.Data != oldData || !reflect.DeepEqual(oldFanins, node.fanins) || len(n.objects) != count {
		t.Fatal("failed edit changed network")
	}
	fanins[node.ID] = []int32{999}
	if err := n.ReplaceFXUCovers(sops, fanins); err == nil {
		t.Fatal("missing reference accepted")
	}
}
func TestFXUAppendedEmptyRowsRemapVirtualIDs(t *testing.T) {
	n := NewNetwork("holes")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_SOP)
	pi := n.NewObject(ObjectPI)
	node := n.NewObject(ObjectNode)
	node.Data = "1 1\n"
	Abc_ObjAddFanin(node, pi)
	sops, fanins, err := n.FXUCovers()
	if err != nil {
		t.Fatal(err)
	}
	base := len(sops)
	sops = append(sops, "", "1 1\n")
	fanins = append(fanins, nil, []int32{int32(pi.ID)})
	fanins[node.ID] = []int32{int32(base + 1)}
	if err := n.ReplaceFXUCovers(sops, fanins); err != nil {
		t.Fatal(err)
	}
	if node.fanins[0].Object.ID != base || node.fanins[0].Object.fanins[0].Object != pi {
		t.Fatal("virtual row IDs were not remapped")
	}
}
func TestFXUEmptyFalseCoverIsNotMissingDefinition(t *testing.T) {
	n, node, pis := fanFixture("")
	sops, fanins, err := n.FXUCovers()
	if err != nil {
		t.Fatal(err)
	}
	if sops[node.ID] == "" {
		t.Fatal("false function confused with missing row")
	}
	if err := n.ReplaceFXUCovers(sops, fanins); err != nil {
		t.Fatal(err)
	}
	for _, value := range fanTruth(node, pis) {
		if value {
			t.Fatal("false function changed")
		}
	}
}
