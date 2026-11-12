package abc

import (
	aig "github.com/glycerine/abcgo/aig/aig"
	"reflect"
	"testing"
)

func TestBufferedHierarchyFlattenAndSweep(t *testing.T) {
	for _, nested := range []bool{false, true} {
		root := testHierarchicalDesign(t, nested)
		expected := Abc_NtkDeriveFlatGia(root)
		var ports []*Object
		graphs := []*aig.Graph{Abc_NtkFlattenHierarchyGia2(root), Abc_NtkFlattenHierarchyGia(root, &ports, 0)}
		for _, g := range graphs {
			info := g.Data.(*HierarchyBufferInfo)
			if len(info.Buffers) == 0 || len(info.Ports) != len(info.Buffers) {
				t.Fatal("lost boundaries")
			}
			swept := Gia_ManSweepHierarchy(g)
			if len(swept.Data.(*HierarchyBufferInfo).Buffers) != len(info.Buffers) {
				t.Fatal("sweep discarded physical boundary")
			}
			for bits := 0; bits < 4; bits++ {
				input := []bool{bits&1 != 0, bits&2 != 0}
				want, _ := expected.Evaluate(input)
				for _, graph := range []*aig.Graph{g, swept} {
					got, err := graph.Evaluate(input)
					if err != nil || !reflect.DeepEqual(got, want) {
						t.Fatal(bits, got, want, err)
					}
				}
			}
		}
		if len(ports) != len(graphs[1].Data.(*HierarchyBufferInfo).Buffers) {
			t.Fatal("port correspondence changed")
		}
	}
}
func TestHierarchySweepAliasesAndUnusedBuffers(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	info := &HierarchyBufferInfo{}
	g.Data = info
	first := info.append(g, nil, x)
	second := info.append(g, nil, x.Not())
	info.append(g, nil, 0)
	info.append(g, nil, 1)
	g.AddOutput(g.AND(first, second.Not()))
	swept := Gia_ManSweepHierarchy(g)
	result := swept.Data.(*HierarchyBufferInfo)
	if len(result.Buffers) != 4 || result.Signals[1] != result.Signals[0].Not() || result.Signals[2] != 0 || result.Signals[3] != 1 {
		t.Fatal(result)
	}
	for _, v := range []bool{false, true} {
		got, err := swept.Evaluate([]bool{v})
		if err != nil || got[0] != v {
			t.Fatal(got, err)
		}
	}
}
func TestBufferedHierarchyReusedModels(t *testing.T) {
	root := testHierarchicalDesign(t, false)
	leaf := root.Design.Module("leaf")
	a, b := root.ObjectsOfKind(ObjectPI)[0].fanouts[0], root.ObjectsOfKind(ObjectPI)[1].fanouts[0]
	testOutputNet(root, "again", testInstance(root, leaf, "second", b, a).Signal())
	g := Abc_NtkFlattenHierarchyGia2(root)
	for bits := 0; bits < 4; bits++ {
		x, y := bits&1 != 0, bits&2 != 0
		got, err := g.Evaluate([]bool{x, y})
		if err != nil || !reflect.DeepEqual(got, []bool{x && !y, y && !x}) {
			t.Fatal(got, err)
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("demand-driven variant accepted a reused model")
		}
	}()
	Abc_NtkFlattenHierarchyGia(root, nil, 0)
}
func testMappedHierarchyGraph(g *aig.Graph) *Network {
	n := NewNetwork("mapped")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_MAP)
	copies := map[int]Edge{0: n.One()}
	var read func(aig.Edge) Edge
	read = func(e aig.Edge) Edge {
		value := copies[e.NodeID()]
		if e.Complemented() {
			value = value.Not()
		}
		return value
	}
	for _, edge := range g.LiveEdges() {
		node := g.Node(edge)
		switch node.Kind {
		case aig.Input:
			copies[edge.NodeID()] = n.NewObject(ObjectPI).Signal()
		case aig.And, aig.ExclusiveOr:
			o := n.NewObject(ObjectNode)
			kind := node.Kind
			o.Data = func(v []bool) bool {
				if kind == aig.And {
					return v[0] && v[1]
				}
				return v[0] != v[1]
			}
			Abc_ObjAddFanin(o, read(node.Fanins[0]))
			Abc_ObjAddFanin(o, read(node.Fanins[1]))
			copies[edge.NodeID()] = o.Signal()
		case aig.Buffer:
			o := Abc_NtkCreateNodeBuf(n, nil)
			o.Barrier = true
			Abc_ObjAddFanin(o, read(node.Fanins[0]))
			copies[edge.NodeID()] = o.Signal()
		case aig.Output:
			o := n.NewObject(ObjectPO)
			Abc_ObjAddFanin(o, read(node.Fanins[0]))
			copies[edge.NodeID()] = o.Signal()
		}
	}
	return n
}
func TestHierarchyMappedReinsertion(t *testing.T) {
	root := testHierarchicalDesign(t, true)
	design := root.Design
	before := Abc_NtkDeriveFlatGia(root)
	flat := testMappedHierarchyGraph(Abc_NtkFlattenHierarchyGia(root, nil, 0))
	fanins := make(map[*Object]int)
	for _, o := range flat.Objects() {
		fanins[o] = len(o.fanins)
	}
	Abc_NtkInsertHierarchyGia(root, flat, 0)
	if root.Design != design || design.Module(root.Name) != root {
		t.Fatal("design identity changed")
	}
	after := Abc_NtkDeriveFlatGia(root)
	for bits := 0; bits < 4; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0}
		want, _ := before.Evaluate(input)
		got, err := after.Evaluate(input)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal(got, want, err)
		}
	}
	for o, count := range fanins {
		if len(o.fanins) != count {
			t.Fatal("reinsertion mutated mapped source fanins")
		}
	}
}

func TestBufferedHierarchyBoundaryPhases(t *testing.T) {
	for phases := 0; phases < 256; phases++ {
		root := testHierarchicalDesign(t, false)
		leaf := root.Design.Module("leaf")
		box := root.ObjectsOfKind(ObjectWhitebox)[0]
		box.fanins[0].Object.fanins[0].Complement = phases&1 != 0
		box.fanins[0].Complement = phases&2 != 0
		leaf.ObjectsOfKind(ObjectPO)[0].fanins[0].Complement = phases&4 != 0
		box.fanouts[0].fanins[0].Complement = phases&8 != 0
		box.fanouts[0].fanouts[0].fanins[0].Complement = phases&16 != 0
		root.ObjectsOfKind(ObjectPO)[0].fanins[0].Complement = phases&32 != 0
		leaf.ObjectsOfKind(ObjectPI)[0].fanouts[0].fanins[0].Complement = phases&64 != 0
		root.ObjectsOfKind(ObjectPI)[0].fanouts[0].fanins[0].Complement = phases&128 != 0
		want := Abc_NtkDeriveFlatGia(root)
		graphs := []*aig.Graph{Abc_NtkFlattenHierarchyGia(root, nil, 0), Abc_NtkFlattenHierarchyGia2(root)}
		flat := testMappedHierarchyGraph(graphs[0])
		Abc_NtkInsertHierarchyGia(root, flat, 0)
		graphs = append(graphs, Abc_NtkDeriveFlatGia(root))
		for bits := 0; bits < 4; bits++ {
			input := []bool{bits&1 != 0, bits&2 != 0}
			expected, _ := want.Evaluate(input)
			for k, g := range graphs {
				got, err := g.Evaluate(input)
				if err != nil || !reflect.DeepEqual(got, expected) {
					t.Fatalf("phase %d variant %d: %v want %v err %v", phases, k, got, expected, err)
				}
			}
		}
	}
}
