package abc

import (
	"fmt"
	aig "github.com/glycerine/abcgo/aig/aig"
	"reflect"
	"testing"
)

func TestHierarchyDirectAndCachedAIGDerivation(t *testing.T) {
	root := testHierarchicalDesign(t, true)
	models := Abc_NtkCollectHie(root)
	if len(models) != 3 || models[0].Name != "leaf" || models[2] != root {
		t.Fatal(models)
	}
	root.Data = "keep caller data"
	direct, cached := Abc_NtkDeriveFlatGia(root), Abc_NtkDeriveFlatGia2(root, models)
	if direct == nil || cached == nil {
		t.Fatal("hierarchy derivation failed")
	}
	if root.Data != "keep caller data" {
		t.Fatal("derivation overwrote user data")
	}
	for assignment := 0; assignment < 4; assignment++ {
		x, y := assignment&1 != 0, assignment&2 != 0
		want := !(x && !y)
		for _, graph := range []*aig.Graph{direct, cached} {
			got, err := graph.Evaluate([]bool{x, y})
			if err != nil || len(got) != 1 || got[0] != want {
				t.Fatal(assignment, got, want, err)
			}
		}
	}
	if Abc_NtkCountInst_rec(root) != 3 || Abc_NtkCountNodes_rec(root) != 2 {
		t.Fatal("hierarchy counts incorrect")
	}
	if graph := Abc_NtkDeriveFlatGia2(root, []*Network{root, models[1], models[0]}); graph == nil {
		t.Fatal("actual dependency caching required")
	}
}
func TestHierarchyDerivationReusedModelsAndWireAliases(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	a, b := root.ObjectsOfKind(ObjectPI)[0].fanouts[0], root.ObjectsOfKind(ObjectPI)[1].fanouts[0]
	net := testInstance(root, leaf, "second", b, a)
	testOutputNet(root, "second", net.Signal())
	graph := Abc_NtkDeriveFlatGia2(root, Abc_NtkCollectHie(root))
	if graph == nil {
		t.Fatal("reused model failed")
	}
	for assignment := 0; assignment < 4; assignment++ {
		x, y := assignment&1 != 0, assignment&2 != 0
		got, err := graph.Evaluate([]bool{x, y})
		if err != nil || !reflect.DeepEqual(got, []bool{x && !y, y && !x}) {
			t.Fatal(got, err)
		}
	}
	wire := testModule("wire")
	input := testInputNet(wire, "x")
	po := wire.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, input)
	graph = Abc_NtkDeriveFlatGia(wire)
	if graph == nil {
		t.Fatal("wire model failed")
	}
	got, err := graph.Evaluate([]bool{true})
	if err != nil || !reflect.DeepEqual(got, []bool{true}) {
		t.Fatal(got, err)
	}
}
func TestHierarchyBoxBoundaryPhasesMatchFlattenedNetwork(t *testing.T) {
	for phases := 0; phases < 256; phases++ {
		root := testHierarchicalDesign(t, false)
		leaf := root.Design.Module("leaf")
		box := root.ObjectsOfKind(ObjectWhitebox)[0]
		pi := root.ObjectsOfKind(ObjectPI)[0]
		formalPI := leaf.ObjectsOfKind(ObjectPI)[0]
		box.fanins[0].Object.fanins[0].Complement = phases&1 != 0
		box.fanins[0].Complement = phases&2 != 0
		leaf.ObjectsOfKind(ObjectPO)[0].fanins[0].Complement = phases&4 != 0
		box.fanouts[0].fanins[0].Complement = phases&8 != 0
		box.fanouts[0].fanouts[0].fanins[0].Complement = phases&16 != 0
		root.ObjectsOfKind(ObjectPO)[0].fanins[0].Complement = phases&32 != 0
		formalPI.fanouts[0].fanins[0].Complement = phases&64 != 0
		pi.fanouts[0].fanins[0].Complement = phases&128 != 0
		graph := Abc_NtkDeriveFlatGia(root)
		cached := Abc_NtkDeriveFlatGia2(root, Abc_NtkCollectHie(root))
		flat := Abc_NtkFlattenLogicHierarchy2(root)
		if graph == nil || cached == nil || flat == nil {
			t.Fatal("phase conversion failed", phases)
		}
		for assignment := 0; assignment < 4; assignment++ {
			inputs := []bool{assignment&1 != 0, assignment&2 != 0}
			expected, err := graph.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			alternative, err := cached.Evaluate(inputs)
			if err != nil || !reflect.DeepEqual(alternative, expected) {
				t.Fatal("cached phase transfer mismatch", phases, assignment, alternative, expected, err)
			}
			values := map[*Object]bool{}
			for i, pi := range flat.ObjectsOfKind(ObjectPI) {
				values[pi] = inputs[i]
			}
			got, err := flat.Evaluate(flat.ObjectsOfKind(ObjectPO)[0].Signal(), values)
			if err != nil || got != expected[0] {
				t.Fatalf("phases=%08b assignment=%d got %v want %v: %v", phases, assignment, got, expected, err)
			}
		}
	}
}
func TestHierarchyReaderBoundaryAndMissingService(t *testing.T) {
	hierarchyReader.RLock()
	previous := hierarchyReader.read
	hierarchyReader.RUnlock()
	defer SetHierarchyReader(previous)
	SetHierarchyReader(nil)
	if _, err := ReadHierarchyAIG("missing", false); err == nil {
		t.Fatal("missing IO service accepted")
	}
	SetHierarchyReader(func(filename string, multiValued, check bool) (*Network, error) {
		if filename != "fixture" || multiValued || !check {
			return nil, fmt.Errorf("wrong reader contract")
		}
		return testHierarchicalDesign(t, false), nil
	})
	graph, err := ReadHierarchyAIG("fixture", false)
	if err != nil || graph == nil {
		t.Fatal(graph, err)
	}
	outputs, err := graph.Evaluate([]bool{true, false})
	if err != nil || !reflect.DeepEqual(outputs, []bool{true}) {
		t.Fatal(outputs, err)
	}
}
func TestHierarchyRecursiveModelsRejected(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	testInstance(leaf, root, "cycle", leaf.ObjectsOfKind(ObjectPI)[0].fanouts[0], leaf.ObjectsOfKind(ObjectPI)[1].fanouts[0])
	if Abc_NtkCheckRecursive(root) != 1 || Abc_NtkDeriveFlatGia(root) != nil || Abc_NtkCollectHie(root) != nil {
		t.Fatal("indirect recursion accepted")
	}
}
