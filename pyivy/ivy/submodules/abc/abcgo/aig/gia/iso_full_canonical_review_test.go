package gia

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func (model isoSignatureReviewModel) evaluate(input []bool) []bool {
	values := make([]bool, len(model.nodes))
	ci := 0
	for id, node := range model.nodes {
		switch node.kind {
		case 'p', 'r':
			values[id] = input[ci]
			ci++
		case 'a':
			values[id] = (values[node.edges[0].id] != (node.edges[0].phase != 0)) && (values[node.edges[1].id] != (node.edges[1].phase != 0))
		case 'o':
			values[id] = values[node.edges[0].id] != (node.edges[0].phase != 0)
		}
	}
	return values
}

func TestIsoFullCanonicalizationPreservesGraphAndRegisterPairing(t *testing.T) {
	phases := []uint32{0, 0x3fff, 0x1555, 0x2aaa}
	for bit := 0; bit < 14; bit++ {
		phases = append(phases, 1<<bit)
	}
	for registers := 0; registers <= 2; registers++ {
		for _, phase := range phases {
			for _, single := range []bool{false, true} {
				t.Run(fmt.Sprintf("registers%d/phases%04x/single%t", registers, phase, single), func(t *testing.T) {
					source, model := isoSignatureReviewGraph(t, registers, phase, single)
					source.pName, source.PSpec = "canonical-review", "source-spec"
					keep := make([]bool, len(model.nodes))
					fields := make([][8]uint32, len(model.nodes))
					var pending []int
					for id, node := range model.nodes {
						obj := Gia_ManObj(source, int32(id))
						obj.fMark0, obj.fMark1 = uint32(id%2), uint32((id+1)%2)
						fields[id] = isoSignatureReviewFields(obj)
						if node.kind != 'a' {
							keep[id] = true
						}
						if node.kind == 'o' {
							pending = append(pending, node.edges[0].id)
						}
					}
					for len(pending) > 0 {
						id := pending[len(pending)-1]
						pending = pending[:len(pending)-1]
						if keep[id] {
							continue
						}
						keep[id] = true
						for _, edge := range model.nodes[id].edges {
							pending = append(pending, edge.id)
						}
					}
					expectedObjects := int32(0)
					for _, yes := range keep {
						if yes {
							expectedObjects++
						}
					}
					canonical := Gia_ManIsoCanonicize(source, 0)
					if canonical == nil || canonical == source || Gia_ManObjNum(canonical) != expectedObjects || Gia_ManPiNum(canonical) != 2 || Gia_ManRegNum(canonical) != int32(registers) || Gia_ManPoNum(canonical) != Gia_ManPoNum(source) {
						t.Fatal("canonicalization changed the interface or failed to remove only dangling gates")
					}
					if canonical.pName != source.pName || canonical.PSpec != source.PSpec {
						t.Fatal("canonicalization lost source identity")
					}
					// The source copy contract leaves an original-to-new literal in Value.
					// Verify that mapping as a graph isomorphism independently of label order.
					mapping := make([]int32, len(model.nodes))
					seen := make([]bool, expectedObjects)
					for id, node := range model.nodes {
						original := Gia_ManObj(source, int32(id))
						if isoSignatureReviewFields(original) != fields[id] {
							t.Fatalf("source object%d structural/mark fields changed", id)
						}
						if !keep[id] {
							continue
						}
						literal := original.Value
						if literal&1 != 0 || literal>>1 >= uint32(expectedObjects) || seen[literal>>1] {
							t.Fatalf("object%d invalid or duplicate copy literal%08x", id, literal)
						}
						mapping[id] = int32(literal)
						seen[literal>>1] = true
						copied := Gia_ManObj(canonical, int32(literal>>1))
						valid := false
						switch node.kind {
						case 'c':
							valid = Gia_ObjIsConst0(copied) != 0
						case 'p':
							valid = Gia_ObjIsPi(canonical, copied) != 0
						case 'r':
							valid = Gia_ObjIsRo(canonical, copied) != 0
						case 'a':
							valid = Gia_ObjIsAnd(copied) != 0
						case 'o':
							valid = Gia_ObjIsCo(copied) != 0
						}
						if !valid {
							t.Fatalf("object%d changed kind", id)
						}
					}
					for id, node := range model.nodes {
						if !keep[id] || len(node.edges) == 0 {
							continue
						}
						copyId := mapping[id] >> 1
						copied := Gia_ManObj(canonical, copyId)
						expected0 := mapping[node.edges[0].id] ^ int32(node.edges[0].phase)
						actual0 := Gia_ObjFaninLit0(copied, copyId)
						if node.kind == 'o' {
							if actual0 != expected0 {
								t.Fatalf("CO%d copied the wrong edge", id)
							}
						} else {
							expected1 := mapping[node.edges[1].id] ^ int32(node.edges[1].phase)
							actual1 := Gia_ObjFaninLit1(copied, copyId)
							if !(actual0 == expected0 && actual1 == expected1 || actual0 == expected1 && actual1 == expected0) {
								t.Fatalf("AND%d copied different fanins/phases", id)
							}
						}
					}
					for _, pair := range model.regs {
						ri := Gia_ManObj(canonical, mapping[pair[0]]>>1)
						ro := Gia_ManObj(canonical, mapping[pair[1]]>>1)
						if Gia_ObjIsRi(canonical, ri) == 0 || Gia_ObjCioId(ri)-Gia_ManPoNum(canonical) != Gia_ObjCioId(ro)-Gia_ManPiNum(canonical) {
							t.Fatal("canonical register pairing changed")
						}
					}
					cis := int(Gia_ManCiNum(source))
					for assignment := 0; assignment < 1<<cis; assignment++ {
						originalInput, newInput := make([]bool, cis), make([]bool, cis)
						for i := range originalInput {
							originalInput[i] = assignment&(1<<i) != 0
							copied := Gia_ManObj(canonical, mapping[i+1]>>1)
							newInput[Gia_ObjCioId(copied)] = originalInput[i]
						}
						expected := model.evaluate(originalInput)
						actual, err := giaEvaluateNodes(canonical, newInput)
						if err != nil {
							t.Fatal(err)
						}
						for id, yes := range keep {
							if yes && actual[mapping[id]>>1] != expected[id] {
								t.Fatalf("assignment%d object%d changed function", assignment, id)
							}
						}
					}
				})
			}
		}
	}
}

func TestIsoFullRefinementCompletesLabelsIndependentlyOfScratch(t *testing.T) {
	phases := []uint32{0, 0x3fff, 0x1555, 0x2aaa}
	for bit := 0; bit < 14; bit++ {
		phases = append(phases, 1<<bit)
	}
	for registers := 0; registers <= 2; registers++ {
		for _, phase := range phases {
			for _, single := range []bool{false, true} {
				t.Run(fmt.Sprintf("registers%d/phases%04x/single%t", registers, phase, single), func(t *testing.T) {
					graph, model := isoSignatureReviewGraph(t, registers, phase, single)
					fields := make([][8]uint32, len(model.nodes))
					for id := range model.nodes {
						obj := Gia_ManObj(graph, int32(id))
						obj.fMark0, obj.fMark1 = uint32(id%2), uint32((id+1)%2)
						fields[id] = isoSignatureReviewFields(obj)
					}
					var first []uint32
					for pass := 0; pass < 2; pass++ {
						for id := range model.nodes {
							value := uint32(id) * 0x9e3779b9
							if pass != 0 {
								value = ^value
							}
							Gia_ManObj(graph, int32(id)).Value = value
						}
						groups := Gia_IsoDeriveEquivPos(graph, 0, 0)
						if int(Gia_ManObjNum(graph)) != len(model.nodes) {
							t.Fatal("refinement changed graph object count")
						}
						labels := make([]uint32, len(model.nodes))
						seen := make([]bool, len(model.nodes)+1)
						for id := range model.nodes {
							obj := Gia_ManObj(graph, int32(id))
							label := obj.Value
							if label == 0 || label > uint32(len(model.nodes)) || seen[label] {
								t.Fatalf("pass%d object%d label%d is incomplete/duplicated", pass, id, label)
							}
							seen[label] = true
							labels[id] = label
							if isoSignatureReviewFields(obj) != fields[id] {
								t.Fatalf("refinement changed source object%d fields", id)
							}
							if pass != 0 && label != first[id] {
								t.Fatalf("object%d label depends on preexisting scratch", id)
							}
						}
						if labels[0] != 1 {
							t.Fatal("constant did not retain the source's first unique label")
						}
						if Gia_ManPoNum(graph) == 1 {
							if groups != nil {
								t.Fatal("single-output refinement should not allocate output classes")
							}
						} else {
							if groups == nil || vec.Vec_PtrSize(groups) != int(Gia_ManPoNum(graph)) {
								t.Fatal("full refinement did not label every output uniquely")
							}
							for output := int32(0); output < Gia_ManPoNum(graph); output++ {
								group := vec.Vec_IntArray(vec.Vec_PtrEntry(groups, int(output)).(*Vec_Int_t))
								if len(group) != 1 || group[0] != output {
									t.Fatalf("output%d did not retain source singleton class", output)
								}
							}
						}
						first = labels
					}
				})
			}
		}
	}
}
