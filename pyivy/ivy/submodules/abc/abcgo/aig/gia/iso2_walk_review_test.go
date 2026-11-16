package gia

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func (model isoSignatureReviewModel) primaryOutputs() []int32 {
	var result []int32
	for id, node := range model.nodes {
		if node.kind != 'o' {
			continue
		}
		ri := false
		for _, pair := range model.regs {
			ri = ri || pair[0] == id
		}
		if !ri {
			result = append(result, int32(id))
		}
	}
	return result
}

// Explicit entry/exit frames avoid using the production recursive walks.
func (model isoSignatureReviewModel) order(values []uint64, outputs []int32) (roots, order, mapping []int32, seen []bool) {
	pos := model.primaryOutputs()
	for _, output := range outputs {
		roots = append(roots, pos[output])
	}
	mapping = make([]int32, len(model.nodes))
	for i := range mapping {
		mapping[i] = -77
	}
	seen = make([]bool, len(model.nodes))
	type frame struct {
		id   int
		exit bool
	}
	for root := 0; root < len(roots); root++ {
		stack := []frame{{id: int(roots[root])}}
		for len(stack) > 0 {
			next := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if next.exit {
				mapping[next.id] = int32(len(order))
				order = append(order, int32(next.id))
				continue
			}
			if seen[next.id] {
				continue
			}
			seen[next.id] = true
			stack = append(stack, frame{next.id, true})
			node := model.nodes[next.id]
			if node.kind == 'r' {
				for _, pair := range model.regs {
					if pair[1] == next.id {
						roots = append(roots, int32(pair[0]))
					}
				}
			}
			var children []int
			for _, edge := range node.edges {
				children = append(children, edge.id)
			}
			if len(children) == 2 && values[children[0]] > values[children[1]] {
				children[0], children[1] = children[1], children[0]
			}
			for i := len(children) - 1; i >= 0; i-- {
				stack = append(stack, frame{id: children[i]})
			}
		}
	}
	return
}

func TestIso2OrderAgainstIndependentSequentialGraph(t *testing.T) {
	selections := [][]int32{nil, {0}, {1}, {1, 0}, {0, 0}, {1, 0, 1}}
	for registers := 0; registers <= 2; registers++ {
		for _, phases := range []uint32{0, 0x3fff, 0x1555, 0x2aaa} {
			for profile := 0; profile < 3; profile++ {
				for selection, outputs := range selections {
					t.Run(fmt.Sprintf("registers%d/phases%04x/values%d/selection%d", registers, phases, profile, selection), func(t *testing.T) {
						graph, model := isoSignatureReviewGraph(t, registers, phases)
						values := make([]uint64, len(model.nodes))
						fields := make([][8]uint32, len(model.nodes))
						for id := range values {
							if profile == 1 {
								values[id] = uint64(len(values) - id)
							} else if profile == 2 {
								values[id] = uint64(uint32(id) * 0x89abcdef)
							}
							obj := Gia_ManObj(graph, int32(id))
							obj.Value = uint32(values[id])
							obj.fMark0, obj.fMark1 = uint32(id%2), uint32((id+1)%2)
							fields[id] = isoSignatureReviewFields(obj)
						}
						wantRoots, wantOrder, wantMap, seen := model.order(values, outputs)
						roots, order, mapping := vec.Vec_IntStart(1), vec.Vec_IntStart(1), vec.Vec_IntStart(len(values))
						for id := range values {
							vec.Vec_IntWriteEntry(mapping, id, -77)
						}
						Gia_ManIncrementTravId(graph)
						for id := range values {
							Gia_ObjSetTravIdCurrentId(graph, int32(id))
						}
						epoch := graph.nTravIds
						// A trailing invalid PO proves the count, rather than slice length, controls traversal.
						input := append(slices.Clone(outputs), 999)
						Gia_Iso2ManCollectOrder(graph, input, int32(len(outputs)), roots, order, mapping)
						for _, check := range []struct {
							name string
							got  *Vec_Int_t
							want []int32
						}{
							{"roots", roots, wantRoots}, {"order", order, wantOrder}, {"mapping", mapping, wantMap},
						} {
							if !slices.Equal(vec.Vec_IntArray(check.got), check.want) {
								t.Fatalf("%s got%v want%v", check.name, vec.Vec_IntArray(check.got), check.want)
							}
						}
						if graph.nTravIds != epoch+1 {
							t.Fatal("order did not start exactly one traversal epoch")
						}
						for id := range values {
							if (Gia_ObjIsTravIdCurrentId(graph, int32(id)) != 0) != seen[id] {
								t.Fatalf("object%d traversal mark differs", id)
							}
						}
						checkIsoSignatureReview(t, graph, values, fields, "sequential order")
						if registers == 0 {
							comb := Gia_Iso2ManCollectOrder2(graph, input, int32(len(outputs)))
							if !slices.Equal(vec.Vec_IntArray(comb), wantOrder) {
								t.Fatalf("combinational order%v want%v", vec.Vec_IntArray(comb), wantOrder)
							}
							for id := range values {
								if (Gia_ObjIsTravIdCurrentId(graph, int32(id)) != 0) != seen[id] {
									t.Fatalf("combinational object%d mark differs", id)
								}
							}
							checkIsoSignatureReview(t, graph, values, fields, "combinational order")
						}
					})
				}
			}
		}
	}
}

// Bounded first-visit DFS: a register bridge preserves depth; AND edges use one
// unit. The first path claims a shared node even if a later path has more depth.
func (model isoSignatureReviewModel) cone(values []uint64, root int, depth int32) (sum uint64, seen []bool) {
	seen = make([]bool, len(model.nodes))
	type frame struct {
		id    int
		depth int32
	}
	stack := []frame{{root, depth}}
	for len(stack) > 0 {
		next := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if next.depth == 0 || seen[next.id] {
			continue
		}
		seen[next.id] = true
		sum = (sum + values[next.id]) & 0xffffffff
		node := model.nodes[next.id]
		if node.kind == 'a' {
			for i := len(node.edges) - 1; i >= 0; i-- {
				stack = append(stack, frame{node.edges[i].id, next.depth - 1})
			}
		} else if node.kind == 'r' {
			for _, pair := range model.regs {
				if pair[1] == next.id {
					stack = append(stack, frame{model.nodes[pair[0]].edges[0].id, next.depth})
				}
			}
		}
	}
	return
}

func TestIso2BoundedConesAndOrderedUpdatesAgainstScalarModel(t *testing.T) {
	for registers := 0; registers <= 2; registers++ {
		for _, phases := range []uint32{0, 0x3fff, 0x1555, 0x2aaa} {
			for _, depth := range []int32{0, 1, 2, 3, 8} {
				for _, reverse := range []bool{false, true} {
					t.Run(fmt.Sprintf("registers%d/phases%04x/depth%d/reverse%t", registers, phases, depth, reverse), func(t *testing.T) {
						graph, model := isoSignatureReviewGraph(t, registers, phases)
						values := make([]uint64, len(model.nodes))
						fields := make([][8]uint32, len(model.nodes))
						p := Gia_Iso2ManStart(graph)
						defer Gia_Iso2ManStop(p)
						for id := range values {
							values[id] = (0xfffffff0 + uint64(id)*101) & 0xffffffff
							obj := Gia_ManObj(graph, int32(id))
							obj.Value = uint32(values[id])
							obj.fMark0, obj.fMark1 = uint32(id%2), uint32((id+1)%2)
							fields[id] = isoSignatureReviewFields(obj)
						}
						for id, node := range model.nodes {
							if node.kind == 'o' {
								continue
							}
							sum, seen := model.cone(values, id, depth)
							if got := Gia_Iso2ManCone(graph, int32(id), depth); uint64(got) != sum {
								t.Fatalf("root%d sum%08x want%08x", id, got, sum)
							}
							for at := range seen {
								if (Gia_ObjIsTravIdCurrentId(graph, int32(at)) != 0) != seen[at] {
									t.Fatalf("root%d object%d cone mark differs", id, at)
								}
							}
						}
						checkIsoSignatureReview(t, graph, values, fields, "bounded cones")
						tied := vec.Vec_IntArray(p.vTied)
						if reverse {
							slices.Reverse(tied)
						}
						order := slices.Clone(tied)
						var seen []bool
						for _, id := range order {
							sum, marks := model.cone(values, int(id), depth)
							values[id] = (values[id] + sum) & 0xffffffff
							seen = marks
						}
						epoch := graph.nTravIds
						Gia_Iso2ManUpdate(p, depth)
						checkIsoSignatureReview(t, graph, values, fields, "ordered cone update")
						if graph.nTravIds != epoch+int32(len(order)) || !slices.Equal(vec.Vec_IntArray(p.vTied), order) {
							t.Fatal("update changed candidate order or traversal count")
						}
						for id := range seen {
							if (Gia_ObjIsTravIdCurrentId(graph, int32(id)) != 0) != seen[id] {
								t.Fatalf("updated object%d cone mark differs", id)
							}
						}
					})
				}
			}
		}
	}
}
