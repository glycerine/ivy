package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"testing"
)

// Order an abstract graph independently of GIA's traversal helpers.
func checkIsoCanonicalReview(t *testing.T, registers int, phases uint32, single bool, profile int, withPermutation bool) {
	t.Helper()
	graph, model := isoSignatureReviewGraph(t, registers, phases, single)
	labels := make([]uint32, len(model.nodes))
	fields := make([][8]uint32, len(model.nodes))
	for id := range labels {
		labels[id] = uint32(id + 1)
	}
	switch profile {
	case 1:
		slices.Reverse(labels)
	case 2:
		state := uint32(17)
		for i := len(labels) - 1; i > 0; i-- {
			state = state*1664525 + 1013904223
			j := int(state % uint32(i+1))
			labels[i], labels[j] = labels[j], labels[i]
		}
	}
	Gia_ManIncrementTravId(graph)
	for id, label := range labels {
		obj := Gia_ManObj(graph, int32(id))
		obj.Value = label
		obj.fMark0, obj.fMark1 = uint32(id%2), uint32((id+1)%2)
		fields[id] = isoSignatureReviewFields(obj)
		Gia_ObjSetTravIdCurrent(graph, obj) // Old marks must be superseded.
	}
	expectedValues := slices.Clone(labels)
	var pis, ros, pos []int32
	for id, node := range model.nodes {
		switch node.kind {
		case 'p':
			pis = append(pis, int32(id))
		case 'r':
			ros = append(ros, int32(id))
		case 'o':
			ri := false
			for _, pair := range model.regs {
				ri = ri || pair[0] == id
			}
			if !ri {
				pos = append(pos, int32(id))
			}
		}
	}
	// Small insertion sort is separate from production pointer sorting.
	order := func(ids []int32, values []uint32) {
		for i := 1; i < len(ids); i++ {
			for j := i; j > 0 && int32(values[ids[j]]) < int32(values[ids[j-1]]); j-- {
				ids[j], ids[j-1] = ids[j-1], ids[j]
			}
		}
	}
	order(pis, labels)
	order(ros, labels)
	if len(pos) > 1 {
		for _, id := range pos {
			edge := model.nodes[id].edges[0]
			expectedValues[id] = 2*labels[edge.id] + uint32(edge.phase)
		}
		order(pos, expectedValues)
	}
	wantCis := slices.Clone(pis)
	wantCos := slices.Clone(pos)
	for _, ro := range ros {
		wantCis = append(wantCis, ro)
		for _, pair := range model.regs {
			if pair[1] == int(ro) {
				wantCos = append(wantCos, int32(pair[0]))
			}
		}
	}
	seen := make([]bool, len(model.nodes))
	seen[0] = true
	for _, id := range wantCis {
		seen[id] = true
	}
	var wantAnds []int32
	var visit func(int)
	visit = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		node := model.nodes[id]
		if node.kind != 'a' {
			t.Fatal("oracle reached non-AND terminal")
		}
		children := []int{node.edges[0].id, node.edges[1].id}
		if model.nodes[children[0]].kind == 'a' && model.nodes[children[1]].kind == 'a' && labels[children[0]] > labels[children[1]] {
			children[0], children[1] = children[1], children[0]
		}
		for _, child := range children {
			visit(child)
		}
		wantAnds = append(wantAnds, int32(id))
	}
	for _, id := range wantCos {
		visit(model.nodes[id].edges[0].id)
	}
	cis, ands, cos := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	for _, v := range []*Vec_Int_t{cis, ands, cos} {
		vec.Vec_IntPush(v, -99)
	}
	var permutation *Vec_Int_t
	var arg **Vec_Int_t
	if withPermutation {
		arg = &permutation
	}
	Gia_ManFindCaninicalOrder(graph, cis, ands, cos, arg)
	for _, check := range []struct {
		name string
		got  *Vec_Int_t
		want []int32
	}{
		{"CI", cis, wantCis}, {"AND", ands, wantAnds}, {"CO", cos, wantCos},
	} {
		if !slices.Equal(vec.Vec_IntArray(check.got), check.want) {
			t.Fatalf("%s got%v want%v", check.name, vec.Vec_IntArray(check.got), check.want)
		}
	}
	if withPermutation {
		var want []int32
		for _, id := range pis {
			want = append(want, id-1)
		}
		if !slices.Equal(vec.Vec_IntArray(permutation), want) {
			t.Fatalf("PI permutation got%v want%v", vec.Vec_IntArray(permutation), want)
		}
	}
	if int(Gia_ManObjNum(graph)) != len(model.nodes) {
		t.Fatal("object count changed")
	}
	for id, value := range expectedValues {
		obj := Gia_ManObj(graph, int32(id))
		if obj.Value != value || isoSignatureReviewFields(obj) != fields[id] {
			t.Fatalf("object%d value/fields changed unexpectedly", id)
		}
		if (Gia_ObjIsTravIdCurrent(graph, obj) != 0) != seen[id] {
			t.Fatalf("object%d current traversal mark differs", id)
		}
	}
}

func TestIsoCanonicalTraversalAgainstIndependentGraph(t *testing.T) {
	phases := []uint32{0, 0x3fff, 0x1555, 0x2aaa}
	for bit := 0; bit < 14; bit++ {
		phases = append(phases, 1<<bit)
	}
	for registers := 0; registers <= 2; registers++ {
		for _, phase := range phases {
			for _, single := range []bool{false, true} {
				for profile := 0; profile < 3; profile++ {
					for _, permutation := range []bool{false, true} {
						t.Run(fmt.Sprintf("registers%d/phases%04x/single%t/labels%d/permutation%t", registers, phase, single, profile, permutation), func(t *testing.T) {
							checkIsoCanonicalReview(t, registers, phase, single, profile, permutation)
						})
					}
				}
			}
		}
	}
}
