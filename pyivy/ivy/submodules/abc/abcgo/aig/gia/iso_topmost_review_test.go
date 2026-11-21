package gia

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func isoReviewCapture(t *testing.T, run func()) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "iso-report")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = old; file.Close() }()
	run()
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestIsoTopmostReportPreservesSourceTraversalState(t *testing.T) {
	for registers := 0; registers <= 2; registers++ {
		for _, phase := range []uint32{0, 0x3fff, 0x1555} {
			for profile := 0; profile < 4; profile++ {
				t.Run(fmt.Sprintf("registers%d/phases%04x/classes%d", registers, phase, profile), func(t *testing.T) {
					graph, model := isoSignatureReviewGraph(t, registers, phase)
					p := Gia_IsoManStart(graph)
					defer Gia_IsoManStop(p)
					var groups [][]int
					for id, node := range model.nodes {
						p.pLevels[id] = node.level
						if profile == 0 || profile == 2 && node.kind != 'a' || profile == 3 && node.kind != 'r' {
							continue
						}
						for len(groups) <= int(node.level) {
							groups = append(groups, nil)
						}
						groups[node.level] = append(groups[node.level], id)
					}
					var members []int
					for _, group := range groups {
						if len(group) == 0 {
							continue
						}
						vec.Vec_IntPush(p.vClasses, int32(len(members)))
						vec.Vec_IntPush(p.vClasses, int32(len(group)))
						members = append(members, group...)
					}
					marked := make([]bool, len(model.nodes))
					for i, id := range members {
						p.pStoreW[i] = uint64(id)<<32 | uint64(99+i)
						node := model.nodes[id]
						if node.kind == 'a' {
							for _, edge := range node.edges {
								marked[edge.id] = true
							}
						} else if node.kind == 'r' {
							for _, pair := range model.regs {
								if pair[1] == id {
									marked[model.nodes[pair[0]].edges[0].id] = true
								}
							}
						}
					}
					var want strings.Builder
					count, begin := 0, 0
					for _, group := range groups {
						for _, id := range group {
							if !marked[id] {
								count++
								fmt.Fprintf(&want, "%5d : Obj %6d : Level = %4d.  iBegin = %4d.  Size = %4d.\n", count, id, model.nodes[id].level, begin, len(group))
								break
							}
						}
						begin += len(group)
					}
					values := make([]uint32, len(model.nodes))
					fields := make([][8]uint32, len(model.nodes))
					Gia_ManIncrementTravId(graph)
					for id := range model.nodes {
						obj := Gia_ManObj(graph, int32(id))
						obj.Value = uint32(id*31 + 17)
						values[id] = obj.Value
						obj.fMark0, obj.fMark1 = uint8(id%2), uint8((id+1)%2)
						fields[id] = isoSignatureReviewFields(obj)
						Gia_ObjSetTravIdCurrent(graph, obj)
					}
					epoch := graph.nTravIds
					classes, store := slices.Clone(vec.Vec_IntArray(p.vClasses)), slices.Clone(p.pStoreW)
					got := isoReviewCapture(t, func() { Gia_IsoReportTopmost(p) })
					if got != want.String() {
						t.Errorf("report got%q want%q", got, want.String())
					}
					if graph.nTravIds != epoch+1 {
						t.Errorf("traversal epoch got%d want%d", graph.nTravIds, epoch+1)
					}
					for id := range model.nodes {
						obj := Gia_ManObj(graph, int32(id))
						if (Gia_ObjIsTravIdCurrent(graph, obj) != 0) != marked[id] {
							t.Errorf("object%d current mark differs from source fanin marking", id)
						}
						if obj.Value != values[id] || isoSignatureReviewFields(obj) != fields[id] {
							t.Errorf("object%d unrelated fields changed", id)
						}
					}
					if !slices.Equal(vec.Vec_IntArray(p.vClasses), classes) || !slices.Equal(p.pStoreW, store) {
						t.Error("report changed class storage")
					}
				})
			}
		}
	}
}

func TestIsoTopmostEmptyReportReusesTraversalStorage(t *testing.T) {
	graph, _ := isoSignatureReviewGraph(t, 1, 0)
	p := Gia_IsoManStart(graph)
	defer Gia_IsoManStop(p)
	Gia_ManIncrementTravId(graph) // Warm traversal storage independently of report.
	if got := testing.AllocsPerRun(100, func() { Gia_IsoReportTopmost(p) }); got != 0 {
		t.Fatalf("empty report allocated%g times", got)
	}
}
